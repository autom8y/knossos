// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/materialize/source"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// gitCommandTimeout is the maximum time allowed for git subprocesses.
const gitCommandTimeout = 50 * time.Millisecond

// ContextOutput represents the output of the context hook.
type ContextOutput struct {
	SessionID       string            `json:"session_id,omitempty"`
	Status          string            `json:"status,omitempty"`
	Initiative      string            `json:"initiative,omitempty"`
	Rite            string            `json:"rite,omitempty"`
	CurrentPhase    string            `json:"current_phase,omitempty"`
	ExecutionMode   string            `json:"execution_mode,omitempty"`
	HasSession      bool              `json:"has_session"`
	CompactState    string            `json:"compact_state,omitempty"`   // Rehydrated from COMPACT_STATE.md if present
	ThroughlineIDs  map[string]string `json:"throughline_ids,omitempty"` // Active throughline agent IDs
	GitBranch       string            `json:"git_branch,omitempty"`
	BaseBranch      string            `json:"base_branch,omitempty"`
	AvailableRites  []string          `json:"available_rites,omitempty"`
	AvailableAgents []string          `json:"available_agents,omitempty"`
	KnowStatus      string            `json:"know_status,omitempty"` // .know/ freshness summary line
}

// Text implements output.Textable for markdown output.
func (c ContextOutput) Text() string {
	if !c.HasSession {
		return "No active session"
	}

	var b strings.Builder
	b.WriteString("## Session Context\n")
	b.WriteString("| Field | Value |\n")
	b.WriteString("|-------|-------|\n")
	b.WriteString(fmt.Sprintf("| Session | %s |\n", c.SessionID))
	b.WriteString(fmt.Sprintf("| Status | %s |\n", c.Status))
	b.WriteString(fmt.Sprintf("| Initiative | %s |\n", c.Initiative))
	b.WriteString(fmt.Sprintf("| Rite | %s |\n", c.Rite))
	b.WriteString(fmt.Sprintf("| Mode | %s |\n", c.ExecutionMode))
	if len(c.AvailableRites) > 0 {
		b.WriteString(fmt.Sprintf("| Available Rites | %s |\n", strings.Join(c.AvailableRites, ", ")))
	}
	if len(c.AvailableAgents) > 0 {
		b.WriteString(fmt.Sprintf("| Available Agents | %s |\n", strings.Join(c.AvailableAgents, ", ")))
	}
	// Throughline agent IDs are shown unconditionally when present.
	// Omit the section entirely when no IDs are tracked.
	if len(c.ThroughlineIDs) > 0 {
		b.WriteString("\nThroughline Agents:\n")
		// Use sorted keys for deterministic output.
		keys := make([]string, 0, len(c.ThroughlineIDs))
		for k := range c.ThroughlineIDs {
			keys = append(keys, k)
		}
		sortStrings(keys)
		for _, k := range keys {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, c.ThroughlineIDs[k]))
		}
	}
	if c.KnowStatus != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", c.KnowStatus))
	}
	if c.CompactState != "" {
		b.WriteString("\n## Recovered State (from PreCompact checkpoint)\n")
		b.WriteString(c.CompactState)
	}
	return b.String()
}

// sortStrings sorts a slice of strings in place.
// Defined here to avoid importing sort just for this helper.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// newContextCmd creates the context hook subcommand.
func newContextCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Inject session context on SessionStart",
		Long: `Reads session context and outputs it for Claude Code injection.

This hook is triggered on SessionStart events. It reads:
- SESSION_CONTEXT.md if a session exists
- ACTIVE_RITE file for rite context

Output is formatted as a markdown table suitable for Claude context.

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runContext(ctx)
			})
		},
	}

	return cmd
}

func runContext(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runContextCore(ctx, printer)
}

// runContextCore contains the actual logic with injected printer for testing.
func runContextCore(ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a SessionStart event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSessionStart {
		printer.VerboseLog("debug", "skipping context hook for non-SessionStart event",
			map[string]any{"event": string(hookEnv.Event)})
		return outputNoSession(printer)
	}

	// Resolve session context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil {
		printer.VerboseLog("warn", "failed to read current session", map[string]any{"error": err.Error()})
		return outputNoSession(printer)
	}

	if resolver.ProjectRoot() == "" || sessionID == "" {
		return outputNoSession(printer)
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.VerboseLog("warn", "failed to load session context",
			map[string]any{"session_id": sessionID, "error": err.Error()})
		return outputNoSession(printer)
	}

	// Read active rite with backward compatibility
	activeRite := resolver.ReadActiveRite()
	if activeRite == "" {
		activeRite = sessCtx.ActiveRite
	}

	// Determine execution mode
	mode := determineExecutionMode(sessCtx, activeRite)

	// Gather git context (best-effort, errors produce empty strings)
	projectDir := resolver.ProjectRoot()
	gitBranch := getGitBranch(projectDir)
	baseBranch := getBaseBranch(projectDir)

	// Gather rite context using the 4-tier SourceResolver (project > user > knossos >
	// embedded). The old listAvailableRites() only read .knossos/rites/ which is
	// empty in most projects — rites come from KNOSSOS_HOME/rites/ or embedded.
	srcResolver := source.NewSourceResolver(projectDir)
	if embRites := common.EmbeddedRites(); embRites != nil {
		srcResolver.WithEmbeddedFS(embRites)
	}
	resolvedRites, _ := srcResolver.ListAvailableRites()
	availableRites := make([]string, len(resolvedRites))
	for i, r := range resolvedRites {
		availableRites[i] = r.Name
	}

	availableAgents := listAvailableAgents(resolver.AgentsDir())

	// Build output
	result := ContextOutput{
		SessionID:       sessCtx.SessionID,
		Status:          string(sessCtx.Status),
		Initiative:      sessCtx.Initiative,
		Rite:            activeRite,
		CurrentPhase:    sessCtx.CurrentPhase,
		ExecutionMode:   mode,
		HasSession:      true,
		GitBranch:       gitBranch,
		BaseBranch:      baseBranch,
		AvailableRites:  availableRites,
		AvailableAgents: availableAgents,
	}

	// Rehydrate from COMPACT_STATE.md if present (written by PreCompact hook)
	sessionDir := resolver.SessionDir(sessionID)
	compactState := consumeCompactCheckpoint(sessionDir, printer)
	if compactState != "" {
		result.CompactState = compactState
	}

	// Include active throughline agent IDs so they survive compaction re-injection.
	// readThroughlineIDs returns nil when no file exists — omitempty handles it.
	if ids := readThroughlineIDs(sessionDir); len(ids) > 0 {
		result.ThroughlineIDs = ids
	}

	// Check .know/ status (best-effort, <100ms — just readdir + parse frontmatter)
	if knowLine := knowStatus(projectDir, hookEnv.CWD); knowLine != "" {
		result.KnowStatus = knowLine
	}

	// Emit session_start event to clew log (best-effort, non-blocking)
	emitSessionStartEvent(sessionDir, sessCtx.SessionID, sessCtx.Initiative, sessCtx.Complexity, activeRite, printer)

	return printer.Print(result)
}

// outputNoSession outputs the no-session response.
func outputNoSession(printer *output.Printer) error {
	result := ContextOutput{HasSession: false}
	return printer.Print(result)
}

// consumeCompactCheckpoint reads COMPACT_STATE.md from the session directory
// and renames it to COMPACT_STATE.consumed.md to prevent re-injection.
// Returns the checkpoint content or empty string if no checkpoint exists.
func consumeCompactCheckpoint(sessionDir string, printer *output.Printer) string {
	checkpointPath := filepath.Join(sessionDir, CompactCheckpointFile)
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		return "" // No checkpoint — normal path
	}

	// Rename to consumed to prevent re-injection on next SessionStart
	consumedPath := filepath.Join(sessionDir, CompactCheckpointConsumed)
	if renameErr := os.Rename(checkpointPath, consumedPath); renameErr != nil {
		printer.VerboseLog("warn", "failed to rename compact checkpoint",
			map[string]any{"error": renameErr.Error()})
		// Still return the data — consumption is best-effort
	}

	return string(data)
}

// getGitBranch returns the current git branch name.
// Returns empty string if not in a git repo or on error.
func getGitBranch(projectDir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectDir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getBaseBranch returns the default branch of the origin remote.
// Falls back to "main" if it cannot be determined.
func getBaseBranch(projectDir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = projectDir
	out, err := cmd.Output()
	if err != nil {
		return "main"
	}
	ref := strings.TrimSpace(string(out))
	// Strip refs/remotes/origin/ prefix
	return strings.TrimPrefix(ref, "refs/remotes/origin/")
}

// listAvailableRites returns the names of directories under ritesDir that contain manifest.yaml.
// Returns nil on error or if the directory does not exist.
func listAvailableRites(ritesDir string) []string {
	entries, err := os.ReadDir(ritesDir)
	if err != nil {
		return nil
	}
	var rites []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifestPath := filepath.Join(ritesDir, e.Name(), "manifest.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			rites = append(rites, e.Name())
		}
	}
	return rites
}

// listAvailableAgents returns the names of .md files in agentsDir, with the extension stripped.
// Returns nil on error or if the directory does not exist.
func listAvailableAgents(agentsDir string) []string {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil
	}
	var agents []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if before, ok := strings.CutSuffix(name, ".md"); ok {
			agents = append(agents, before)
		}
	}
	return agents
}

// emitSessionStartEvent emits a session_start event to the clew log on SessionStart.
// All emissions are best-effort -- failures do not affect the context hook result.
func emitSessionStartEvent(sessionDir, sessionID, initiative, complexity, rite string, printer *output.Printer) {
	if sessionDir == "" || sessionID == "" {
		return
	}

	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = writer.Close() }()

	event := clewcontract.NewSessionStartEvent(sessionID, initiative, complexity, rite)
	writer.Write(event)

	if flushErr := writer.Flush(); flushErr != nil {
		printer.VerboseLog("warn", "failed to emit session_start event",
			map[string]any{"error": flushErr.Error()})
	}
}

// knowStatus checks .know/ domain freshness and returns a one-line summary string.
// Returns "" if .know/ doesn't exist or is empty. This runs in <100ms:
// it only reads directory entries and parses frontmatter (no full file reads beyond header).
// When cwd is set and differs from projectDir, hierarchical discovery walks from cwd
// up to projectDir, merging service-level and root-level .know/ domains.
func knowStatus(projectDir, cwd string) string {
	if projectDir == "" {
		return ""
	}
	startDir := cwd
	if startDir == "" {
		startDir = projectDir
	}
	domains, err := know.ReadMeta(startDir, projectDir)
	if err != nil || len(domains) == 0 {
		return ""
	}

	// Build per-domain status fragments
	count := len(domains)
	hasStale := false
	parts := make([]string, 0, count)
	for _, d := range domains {
		status := "fresh"
		if !d.Fresh {
			status = "STALE"
			hasStale = true
		}
		parts = append(parts, fmt.Sprintf("%s: %s, expires %s", d.Domain, status, d.Expires))
	}

	domainWord := "domain"
	if count != 1 {
		domainWord = "domains"
	}

	summary := fmt.Sprintf("Codebase knowledge: %d %s (%s).", count, domainWord, strings.Join(parts, ", "))
	if hasStale {
		summary += " Run /know to refresh."
	}
	return summary
}

// determineExecutionMode determines the execution mode based on session and rite.
func determineExecutionMode(sessCtx *session.Context, activeRite string) string {
	// No session = native mode
	if sessCtx == nil {
		return "native"
	}

	// Session with rite = orchestrated mode
	if activeRite != "" && activeRite != "none" {
		return "orchestrated"
	}

	// Session without rite = cross-cutting mode
	return "cross-cutting"
}

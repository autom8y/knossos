// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/materialize/source"
	"github.com/autom8y/knossos/internal/naxos"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
	"github.com/autom8y/knossos/internal/suggest"
)

// gitCommandTimeout is the maximum time allowed for git subprocesses.
const gitCommandTimeout = 50 * time.Millisecond

// StrandOutput represents a child session strand in hook output.
// Mirrors session.Strand but with independent JSON/YAML tags for output coupling.
type StrandOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	FrameRef  string `json:"frame_ref,omitempty"`
	LandedAt  string `json:"landed_at,omitempty"`
}

// ContextOutput represents the output of the context hook.
type ContextOutput struct {
	SessionID       string               `json:"session_id,omitempty"`
	CCSessionID     string               `json:"cc_session_id,omitempty"` // CC's own session ID (for claim)
	FrayedFrom      string               `json:"frayed_from,omitempty"`
	FrameRef        string               `json:"frame_ref,omitempty"`
	ParkSource      string               `json:"park_source,omitempty"`
	Complexity      string               `json:"complexity,omitempty"` // S2 field widening
	Strands         []StrandOutput       `json:"strands,omitempty"`    // S2 field widening
	ClaimedBy       string               `json:"claimed_by,omitempty"` // S2 field widening
	Status          string               `json:"status,omitempty"`
	Initiative      string               `json:"initiative,omitempty"`
	Rite            string               `json:"rite,omitempty"`
	CurrentPhase    string               `json:"current_phase,omitempty"`
	ExecutionMode   string               `json:"execution_mode,omitempty"`
	HasSession      bool                 `json:"has_session"`
	CompactState    string               `json:"compact_state,omitempty"`   // Rehydrated from COMPACT_STATE.md if present
	ThroughlineIDs  map[string]string    `json:"throughline_ids,omitempty"` // Active throughline agent IDs
	GitBranch       string               `json:"git_branch,omitempty"`
	BaseBranch      string               `json:"base_branch,omitempty"`
	AvailableRites  []string             `json:"available_rites,omitempty"`
	AvailableAgents []string             `json:"available_agents,omitempty"`
	KnowStatus      string               `json:"know_status,omitempty"`   // .know/ freshness summary line
	NaxosSummary    string               `json:"naxos_summary,omitempty"` // Naxos triage summary line
	Suggestions     []suggest.Suggestion `json:"suggestions,omitempty"`   // H5: proactive suggestions
	Procession      *session.Procession  `json:"procession,omitempty"`   // Active procession state
}

// Text implements output.Textable for YAML frontmatter output.
// The output is delimited by --- markers and followed by optional post-frontmatter
// markdown sections (Throughline Agents, Recovered State).
func (c ContextOutput) Text() string {
	var b strings.Builder

	// Opening delimiter and comment
	b.WriteString("---\n")
	b.WriteString("# Session Context (injected by ari hook context)\n")

	if !c.HasSession {
		// No-session path: emit minimal frontmatter
		b.WriteString("has_session: false\n")
		if c.CCSessionID != "" {
			fmt.Fprintf(&b, "cc_session_id: %q\n", c.CCSessionID)
		}
		b.WriteString("---\n")
		return b.String()
	}

	// Required fields (always present when has_session=true)
	fmt.Fprintf(&b, "session_id: %s\n", c.SessionID)
	if c.CCSessionID != "" {
		fmt.Fprintf(&b, "cc_session_id: %q\n", c.CCSessionID)
	}
	fmt.Fprintf(&b, "status: %s\n", c.Status)
	fmt.Fprintf(&b, "initiative: %q\n", c.Initiative)
	fmt.Fprintf(&b, "active_rite: %s\n", c.Rite)
	fmt.Fprintf(&b, "execution_mode: %s\n", c.ExecutionMode)

	// Optional scalar fields (omitempty)
	if c.CurrentPhase != "" {
		fmt.Fprintf(&b, "current_phase: %s\n", c.CurrentPhase)
	}
	if c.GitBranch != "" {
		fmt.Fprintf(&b, "git_branch: %s\n", c.GitBranch)
	}
	if c.BaseBranch != "" {
		fmt.Fprintf(&b, "base_branch: %s\n", c.BaseBranch)
	}
	if c.FrayedFrom != "" {
		fmt.Fprintf(&b, "frayed_from: %s\n", c.FrayedFrom)
	}
	if c.FrameRef != "" {
		fmt.Fprintf(&b, "frame_ref: %s\n", c.FrameRef)
	}
	if c.ParkSource != "" {
		fmt.Fprintf(&b, "park_source: %s\n", c.ParkSource)
	}

	// S2 fields: complexity, claimed_by, strands
	if c.Complexity != "" {
		fmt.Fprintf(&b, "complexity: %s\n", c.Complexity)
	}
	if c.ClaimedBy != "" {
		fmt.Fprintf(&b, "claimed_by: %s\n", c.ClaimedBy)
	}
	if len(c.Strands) > 0 {
		b.WriteString("strands:\n")
		for _, s := range c.Strands {
			fmt.Fprintf(&b, "  - session_id: %s\n", s.SessionID)
			fmt.Fprintf(&b, "    status: %s\n", s.Status)
			if s.FrameRef != "" {
				fmt.Fprintf(&b, "    frame_ref: %s\n", s.FrameRef)
			}
			if s.LandedAt != "" {
				fmt.Fprintf(&b, "    landed_at: %q\n", s.LandedAt)
			}
		}
	}

	// Procession state (cross-rite workflow)
	if c.Procession != nil {
		b.WriteString("procession:\n")
		fmt.Fprintf(&b, "  id: %s\n", c.Procession.ID)
		fmt.Fprintf(&b, "  type: %s\n", c.Procession.Type)
		fmt.Fprintf(&b, "  current_station: %s\n", c.Procession.CurrentStation)
		if len(c.Procession.CompletedStations) > 0 {
			b.WriteString("  completed_stations:\n")
			for _, cs := range c.Procession.CompletedStations {
				fmt.Fprintf(&b, "    - station: %s\n", cs.Station)
				fmt.Fprintf(&b, "      rite: %s\n", cs.Rite)
				fmt.Fprintf(&b, "      completed_at: %q\n", cs.CompletedAt)
			}
		}
		if c.Procession.NextStation != "" {
			fmt.Fprintf(&b, "  next_station: %s\n", c.Procession.NextStation)
		}
		if c.Procession.NextRite != "" {
			fmt.Fprintf(&b, "  next_rite: %s\n", c.Procession.NextRite)
		}
		if c.Procession.ArtifactDir != "" {
			fmt.Fprintf(&b, "  artifact_dir: %s\n", c.Procession.ArtifactDir)
		}
	}

	// Array fields rendered as YAML lists (omitempty)
	if len(c.AvailableRites) > 0 {
		b.WriteString("available_rites:\n")
		for _, r := range c.AvailableRites {
			fmt.Fprintf(&b, "  - %s\n", r)
		}
	}
	if len(c.AvailableAgents) > 0 {
		b.WriteString("available_agents:\n")
		for _, a := range c.AvailableAgents {
			fmt.Fprintf(&b, "  - %s\n", a)
		}
	}

	// know_status moves into frontmatter (single-line, no escaping needed)
	if c.KnowStatus != "" {
		fmt.Fprintf(&b, "know_status: %q\n", c.KnowStatus)
	}

	// naxos_summary: one-line triage result from NAXOS_TRIAGE.md (single-line, no escaping needed)
	if c.NaxosSummary != "" {
		fmt.Fprintf(&b, "naxos_summary: %q\n", c.NaxosSummary)
	}

	// Closing delimiter
	b.WriteString("---\n")

	// Post-frontmatter sections: Throughline Agents and CompactState live
	// outside the YAML block because they contain dynamic/multi-line content.
	if len(c.ThroughlineIDs) > 0 {
		b.WriteString("\nThroughline Agents:\n")
		// Use sorted keys for deterministic output.
		keys := make([]string, 0, len(c.ThroughlineIDs))
		for k := range c.ThroughlineIDs {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "  %s: %s\n", k, c.ThroughlineIDs[k])
		}
	}
	if c.CompactState != "" {
		b.WriteString("\n## Recovered State (from PreCompact checkpoint)\n")
		b.WriteString(c.CompactState)
	}

	// H5: Suggestions section (post-frontmatter, like Throughline and CompactState)
	if len(c.Suggestions) > 0 {
		b.WriteString("\n## Suggestions\n")
		for _, s := range c.Suggestions {
			fmt.Fprintf(&b, "- %s\n", s.Text)
		}
	}

	return b.String()
}

// convertStrands converts session.Strand slice to StrandOutput slice for hook output.
// Returns nil when input is nil or empty (omitempty will suppress the field).
func convertStrands(strands []session.Strand) []StrandOutput {
	if len(strands) == 0 {
		return nil
	}
	out := make([]StrandOutput, len(strands))
	for i, s := range strands {
		out[i] = StrandOutput{
			SessionID: s.SessionID,
			Status:    s.Status,
			FrameRef:  s.FrameRef,
			LandedAt:  s.LandedAt,
		}
	}
	return out
}

// newContextCmd creates the context hook subcommand.
func newContextCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Inject session context on SessionStart",
		Long: `Reads session context and outputs it for harness context injection.

This hook is triggered on SessionStart events. It reads:
- SESSION_CONTEXT.md if a session exists
- ACTIVE_RITE file for rite context

Output is formatted as YAML frontmatter suitable for Claude context.

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runContext(cmd, ctx)
			})
		},
	}

	return cmd
}

func runContext(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runContextCore(cmd, ctx, printer)
}

// runContextCore contains the actual logic with injected printer for testing.
func runContextCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv(cmd)

	// Authentication Check: Verify signature of raw payload
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Verify this is a SessionStart event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSessionStart {
		printer.VerboseLog("debug", "skipping context hook for non-SessionStart event",
			map[string]any{"event": string(hookEnv.Event)})
		return outputNoSession(printer, hookEnv.SessionID)
	}

	// Resolve session context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil {
		printer.VerboseLog("warn", "failed to read current session", map[string]any{"error": err.Error()})
		return outputNoSession(printer, hookEnv.SessionID)
	}

	if resolver.ProjectRoot() == "" || sessionID == "" {
		return outputNoSession(printer, hookEnv.SessionID)
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.VerboseLog("warn", "failed to load session context",
			map[string]any{"session_id": sessionID, "error": err.Error()})
		return outputNoSession(printer, hookEnv.SessionID)
	}

	// Read active rite with backward compatibility
	activeRite := resolver.ReadActiveRite()
	if activeRite == "" {
		activeRite = sessCtx.ActiveRite
	}

	// Determine execution mode
	mode := session.DeriveExecutionMode(sessCtx.Status, activeRite)

	// Gather git context (best-effort, errors produce empty strings)
	projectDir := resolver.ProjectRoot()
	gitBranch := getGitBranch(projectDir)
	baseBranch := getBaseBranch(projectDir)

	// Gather rite context using the 4-tier SourceResolver (project > user > knossos >
	// embedded). The old listAvailableRites() only read .knossos/rites/ which is
	// empty in most projects — rites come from KNOSSOS_HOME/rites/ or embedded.
	srcResolver := ctx.sourceResolver
	if srcResolver == nil {
		srcResolver = source.NewSourceResolver(projectDir)
	}
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
		CCSessionID:     hookEnv.SessionID,
		FrayedFrom:      sessCtx.FrayedFrom,
		FrameRef:        sessCtx.FrameRef,
		ParkSource:      sessCtx.ParkSource,
		Complexity:      sessCtx.Complexity,
		Strands:         convertStrands(sessCtx.Strands),
		ClaimedBy:       sessCtx.ClaimedBy,
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

	// Include procession state if active
	if sessCtx.Procession != nil {
		result.Procession = sessCtx.Procession
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

	// Naxos: inject triage summary line (fast path, <5ms — frontmatter only)
	sessionsDir := resolver.SessionsDir()
	if summary := naxos.ReadTriageSummary(sessionsDir); summary != "" {
		result.NaxosSummary = summary
	}

	// H5: Generate proactive suggestions (fail-open, advisory only)
	suggestInput := &suggest.SessionInput{
		SessionID:   sessCtx.SessionID,
		Initiative:  sessCtx.Initiative,
		Phase:       sessCtx.CurrentPhase,
		Rite:        activeRite,
		Complexity:  sessCtx.Complexity,
		ParkSource:  sessCtx.ParkSource,
		StrandCount: len(sessCtx.Strands),
	}
	if suggestions := suggest.SessionStartSuggestions(suggestInput); len(suggestions) > 0 {
		result.Suggestions = suggestions
	}

	// Naxos suggestions are lower priority than session-start suggestions; append after.
	// Only add if there is room under the per-event cap (2).
	if naxosInput := buildNaxosInput(sessionsDir); naxosInput != nil {
		const maxSuggestions = 2
		if len(result.Suggestions) < maxSuggestions {
			naxosSuggestions := suggest.OrphanHygieneSuggestions(naxosInput)
			for _, s := range naxosSuggestions {
				if len(result.Suggestions) >= maxSuggestions {
					break
				}
				result.Suggestions = append(result.Suggestions, s)
			}
		}
	}

	// Emit session_start event to clew log (best-effort, non-blocking)
	emitSessionStartEvent(sessionDir, sessCtx.SessionID, sessCtx.Initiative, sessCtx.Complexity, activeRite, printer)

	return printer.Print(result)
}

// outputNoSession outputs the no-session response.
// ccSessionID is forwarded so models can still use claim even without an active session.
func outputNoSession(printer *output.Printer, ccSessionID string) error {
	result := ContextOutput{HasSession: false, CCSessionID: ccSessionID}
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

// buildNaxosInput reads the NAXOS_TRIAGE.md artifact from sessionsDir and converts it
// into a suggest.NaxosInput. Returns nil on any error (fail-open: missing artifact is normal).
func buildNaxosInput(sessionsDir string) *suggest.NaxosInput {
	result, err := naxos.ReadTriageArtifact(sessionsDir)
	if err != nil {
		// Artifact absent or malformed — normal case on fresh projects.
		return nil
	}
	if result.TotalTriaged == 0 {
		return nil
	}

	// Convert map[naxos.Severity]int → map[string]int for the suggest layer.
	bySeverity := make(map[string]int, len(result.BySeverity))
	for sev, count := range result.BySeverity {
		bySeverity[string(sev)] = count
	}

	input := &suggest.NaxosInput{
		TotalTriaged: result.TotalTriaged,
		BySeverity:   bySeverity,
	}

	// Surface the top entry (first after priority sort) as a summary hint.
	if len(result.Entries) > 0 {
		top := result.Entries[0]
		input.TopEntry = &suggest.TriageEntrySummary{
			SessionID:   top.SessionID,
			Severity:    string(top.Severity),
			Reason:      string(top.Reason),
			Action:      string(top.SuggestedAction),
			InactiveFor: naxos.FormatDuration(top.InactiveFor),
		}
	}

	return input
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
			switch {
			case d.DependencyStale:
				status = "STALE (dep)"
			case d.LandChanged:
				status = "STALE (land)"
			default:
				status = "STALE"
			}
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

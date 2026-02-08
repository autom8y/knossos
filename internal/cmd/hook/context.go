// Package hook implements the ari hook commands.
package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// ContextOutput represents the output of the context hook.
type ContextOutput struct {
	SessionID     string `json:"session_id,omitempty"`
	Status        string `json:"status,omitempty"`
	Initiative    string `json:"initiative,omitempty"`
	Rite          string `json:"rite,omitempty"`
	CurrentPhase  string `json:"current_phase,omitempty"`
	ExecutionMode string `json:"execution_mode,omitempty"`
	HasSession    bool   `json:"has_session"`
	CompactState  string `json:"compact_state,omitempty"` // Rehydrated from COMPACT_STATE.md if present
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
	if c.CompactState != "" {
		b.WriteString("\n## Recovered State (from PreCompact checkpoint)\n")
		b.WriteString(c.CompactState)
	}
	return b.String()
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
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputNoSession(printer)
	}

	// Resolve session context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil {
		printer.VerboseLog("warn", "failed to read current session", map[string]interface{}{"error": err.Error()})
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
			map[string]interface{}{"session_id": sessionID, "error": err.Error()})
		return outputNoSession(printer)
	}

	// Read active rite with backward compatibility
	activeRite := resolver.ReadActiveRite()
	if activeRite == "" {
		activeRite = sessCtx.ActiveRite
	}

	// Determine execution mode
	mode := determineExecutionMode(sessCtx, activeRite)

	// Build output
	result := ContextOutput{
		SessionID:     sessCtx.SessionID,
		Status:        string(sessCtx.Status),
		Initiative:    sessCtx.Initiative,
		Rite:          activeRite,
		CurrentPhase:  sessCtx.CurrentPhase,
		ExecutionMode: mode,
		HasSession:    true,
	}

	// Rehydrate from COMPACT_STATE.md if present (written by PreCompact hook)
	sessionDir := resolver.SessionDir(sessionID)
	compactState := consumeCompactCheckpoint(sessionDir, printer)
	if compactState != "" {
		result.CompactState = compactState
	}

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
			map[string]interface{}{"error": renameErr.Error()})
		// Still return the data — consumption is best-effort
	}

	return string(data)
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

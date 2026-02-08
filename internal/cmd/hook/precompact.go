// Package hook implements the ari hook commands.
package hook

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// precompactResult is the output of the precompact hook.
// PreCompact is a side-effect hook (rotation) — it cannot block.
// CC has no hookSpecificOutput for PreCompact, so we emit plain JSON.
type precompactResult struct {
	Reason string `json:"reason,omitempty"`
}


// newPrecompactCmd creates the precompact hook subcommand.
func newPrecompactCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "precompact",
		Short: "Rotate SESSION_CONTEXT.md on context compaction",
		Long: `Rotates SESSION_CONTEXT.md when Claude Code compacts context window.

This hook is triggered on PreCompact events. It:
- Finds the active session directory
- Checks if SESSION_CONTEXT.md exceeds rotation threshold (200 lines)
- Archives old content to SESSION_CONTEXT.archived.md
- Keeps the most recent 80 lines of body content
- Always returns "allow" (rotation is a side effect, never blocks)

Output (stdout JSON):
  {}
  {"reason": "rotated SESSION_CONTEXT (archived 120 lines, kept 80)"}

Performance: <100ms for rotation operation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runPrecompact(ctx)
			})
		},
	}

	return cmd
}

func runPrecompact(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runPrecompactCore(ctx, printer)
}

// runPrecompactCore contains the actual logic with injected printer for testing.
func runPrecompactCore(ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PreCompact event (or empty for direct invocation/testing)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPreCompact {
		return outputAllowPrecompact(printer, "")
	}

	// Resolve session from hook context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil {
		printer.VerboseLog("warn", "failed to resolve session", map[string]interface{}{"error": err.Error()})
		return outputAllowPrecompact(printer, "")
	}

	// If no session or no project, nothing to rotate
	if sessionID == "" || resolver.ProjectRoot() == "" {
		return outputAllowPrecompact(printer, "")
	}

	// Find session directory
	sessionsDir := filepath.Join(resolver.ProjectRoot(), ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Check if session directory exists
	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if !fileExists(sessionContextPath) {
		// No SESSION_CONTEXT.md to rotate
		return outputAllowPrecompact(printer, "")
	}

	// Attempt rotation
	result, err := session.RotateSessionContext(sessionDir, session.DefaultMaxLines, session.DefaultKeepLines)
	if err != nil {
		printer.VerboseLog("error", "failed to rotate SESSION_CONTEXT", map[string]interface{}{
			"error":      err.Error(),
			"sessionDir": sessionDir,
		})
		// Don't fail the hook - allow compaction to proceed
		return outputAllowPrecompact(printer, "")
	}

	// Output result
	if result.Rotated {
		reason := fmt.Sprintf("rotated SESSION_CONTEXT (archived %d lines, kept %d)", result.ArchivedLines, result.KeptLines)
		return outputAllowPrecompact(printer, reason)
	}

	return outputAllowPrecompact(printer, "")
}

// outputAllowPrecompact outputs a precompact result as plain JSON.
// PreCompact is a side-effect hook — it cannot block. CC has no
// hookSpecificOutput schema for PreCompact, so we emit simple JSON.
func outputAllowPrecompact(printer *output.Printer, reason string) error {
	return printer.Print(precompactResult{Reason: reason})
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

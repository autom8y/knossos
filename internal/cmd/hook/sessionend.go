// Package hook implements the ari hook commands.
package hook

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// SessionEndOutput represents the output of the sessionend hook.
type SessionEndOutput struct {
	SessionID string `json:"session_id,omitempty"`
	Status    string `json:"status,omitempty"`
	WasParked bool   `json:"was_parked"`
	WasEnded  bool   `json:"was_ended"`
	Message   string `json:"message,omitempty"`
}

// Text implements output.Textable for text output.
func (s SessionEndOutput) Text() string {
	if !s.WasEnded {
		return s.Message
	}
	var b strings.Builder
	b.WriteString("Session ended: ")
	b.WriteString(s.SessionID)
	if s.WasParked {
		b.WriteString(" (auto-parked)")
	}
	return b.String()
}

// newSessionEndCmd creates the sessionend hook subcommand.
func newSessionEndCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessionend",
		Short: "Handle SessionEnd event for session cleanup",
		Long: `Handles CC SessionEnd events when the conversation window closes.

This hook is triggered on SessionEnd events. It:
- Emits a session.ended event to the clew
- Auto-parks the session if still ACTIVE
- Cleans up budget temp counter files

Unlike the Stop event (which fires per-turn), SessionEnd fires once
when the CC conversation terminates.

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runSessionEnd(ctx)
			})
		},
	}

	return cmd
}

func runSessionEnd(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runSessionEndCore(ctx, printer)
}

// runSessionEndCore contains the actual logic with injected printer for testing.
func runSessionEndCore(ctx *cmdContext, printer *output.Printer) error {
	hookEnv := ctx.getHookEnv()

	// Verify this is a SessionEnd event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSessionEnd {
		printer.VerboseLog("debug", "skipping sessionend hook for non-SessionEnd event",
			map[string]any{"event": string(hookEnv.Event)})
		return outputNoEnd(printer, "not a SessionEnd event")
	}

	// Resolve session context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil {
		printer.VerboseLog("warn", "failed to read current session", map[string]any{"error": err.Error()})
		return outputNoEnd(printer, "no active session")
	}

	if resolver.ProjectRoot() == "" || sessionID == "" {
		return outputNoEnd(printer, "no active session")
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.VerboseLog("warn", "failed to load session context",
			map[string]any{"session_id": sessionID, "error": err.Error()})
		return outputNoEnd(printer, "could not load session")
	}

	wasParked := false

	// Auto-park if still ACTIVE (belt-and-suspenders with Stop/autopark)
	if sessCtx.Status == session.StatusActive {
		fsm := session.NewFSM()
		if fsm.CanTransition(sessCtx.Status, session.StatusParked) {
			now := time.Now().UTC()
			sessCtx.Status = session.StatusParked
			sessCtx.ParkedAt = &now
			sessCtx.ParkedReason = "auto-parked on SessionEnd"
			sessCtx.ParkSource = "auto"

			if err := sessCtx.Save(ctxPath); err != nil {
				printer.VerboseLog("error", "failed to save session context",
					map[string]any{"session_id": sessionID, "error": err.Error()})
				return outputNoEnd(printer, "failed to save session")
			}
			wasParked = true
		}
	}

	// Emit session.ended event using synchronous EventWriter
	sessionDir := resolver.SessionDir(sessionID)
	event := clewcontract.NewSessionEndEvent(sessionID, string(sessCtx.Status), 0)
	writer, err := clewcontract.NewEventWriter(sessionDir)
	if err == nil {
		_ = writer.Write(event)
		_ = writer.Close()
	}

	// Clean up budget temp counter files (best-effort)
	cleanupBudgetFiles(hookEnv.SessionID)

	printer.VerboseLog("info", "session ended", map[string]any{
		"session_id": sessionID,
		"was_parked": wasParked,
	})

	result := SessionEndOutput{
		SessionID: sessionID,
		Status:    string(sessCtx.Status),
		WasParked: wasParked,
		WasEnded:  true,
		Message:   "Session ended",
	}

	return printer.Print(result)
}

// outputNoEnd outputs a no-op response when session end didn't occur.
func outputNoEnd(printer *output.Printer, reason string) error {
	result := SessionEndOutput{
		WasEnded: false,
		Message:  reason,
	}
	return printer.Print(result)
}

// cleanupBudgetFiles removes budget marker files for the session.
func cleanupBudgetFiles(sessionKey string) {
	if sessionKey == "" {
		return
	}

	// Sanitize key same way as budget.go resolveStateFile
	key := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, sessionKey)

	base := filepath.Join(os.TempDir(), "ari-msg-count-"+key)
	// Remove counter file and marker files
	for _, suffix := range []string{"", ".warned", ".park-warned"} {
		_ = os.Remove(base + suffix)
	}
}

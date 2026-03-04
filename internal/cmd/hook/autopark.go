// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// AutoparkOutput represents the output of the autopark hook.
type AutoparkOutput struct {
	SessionID      string `json:"session_id,omitempty"`
	Status         string `json:"status,omitempty"`
	PreviousStatus string `json:"previous_status,omitempty"`
	AutoParkedAt   string `json:"auto_parked_at,omitempty"`
	WasParked      bool   `json:"was_parked"`
	Message        string `json:"message,omitempty"`
}

// Text implements output.Textable for text output.
func (a AutoparkOutput) Text() string {
	if !a.WasParked {
		return a.Message
	}
	return "Session auto-parked: " + a.SessionID
}

// newAutoparkCmd creates the autopark hook subcommand.
func newAutoparkCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autopark",
		Short: "Auto-park session on Stop event",
		Long: `Automatically transitions active sessions to PARKED on Claude Code Stop.

This hook is triggered on Stop events. It:
- Checks for an active session
- Transitions from ACTIVE to PARKED if applicable
- Records auto_parked_at timestamp
- Gracefully no-ops if no active session

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runAutopark(ctx)
			})
		},
	}

	return cmd
}

func runAutopark(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runAutoparkCore(ctx, printer)
}

// runAutoparkCore contains the actual logic with injected printer for testing.
func runAutoparkCore(ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a Stop event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventStop {
		printer.VerboseLog("debug", "skipping autopark hook for non-Stop event",
			map[string]any{"event": string(hookEnv.Event)})
		return outputNoPark(printer, "not a Stop event")
	}

	// Resolve session context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil {
		printer.VerboseLog("warn", "failed to read current session", map[string]any{"error": err.Error()})
		return outputNoPark(printer, "no active session")
	}

	if resolver.ProjectRoot() == "" || sessionID == "" {
		return outputNoPark(printer, "no active session")
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.VerboseLog("warn", "failed to load session context",
			map[string]any{"session_id": sessionID, "error": err.Error()})
		return outputNoPark(printer, "could not load session")
	}

	// Only park if currently ACTIVE
	if sessCtx.Status != session.StatusActive {
		return outputNoPark(printer, "session not active (status: "+string(sessCtx.Status)+")")
	}

	// Validate transition using FSM
	fsm := session.NewFSM()
	if !fsm.CanTransition(sessCtx.Status, session.StatusParked) {
		return outputNoPark(printer, "invalid transition from "+string(sessCtx.Status))
	}

	// Record previous status
	previousStatus := sessCtx.Status

	// Update session state
	now := time.Now().UTC()
	sessCtx.Status = session.StatusParked
	sessCtx.ParkedAt = &now
	sessCtx.ParkedReason = "auto-parked on Stop"

	// Save session context
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.VerboseLog("error", "failed to save session context",
			map[string]any{"session_id": sessionID, "error": err.Error()})
		return outputNoPark(printer, "failed to save session")
	}

	// Log git status for audit purposes
	gitStatus := getGitStatusQuick()
	printer.VerboseLog("info", "session auto-parked", map[string]any{
		"session_id": sessionID,
		"git_status": gitStatus,
	})

	// Build output
	result := AutoparkOutput{
		SessionID:      sessionID,
		Status:         string(session.StatusParked),
		PreviousStatus: string(previousStatus),
		AutoParkedAt:   now.Format(time.RFC3339),
		WasParked:      true,
		Message:        "Session auto-parked",
	}

	return printer.Print(result)
}

// outputNoPark outputs a no-op response when parking didn't occur.
func outputNoPark(printer *output.Printer, reason string) error {
	result := AutoparkOutput{
		WasParked: false,
		Message:   reason,
	}
	return printer.Print(result)
}

// getGitStatusQuick returns a quick git status for logging.
func getGitStatusQuick() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "status", "--short")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	if strings.TrimSpace(string(out)) == "" {
		return "clean"
	}
	return "uncommitted"
}

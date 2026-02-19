package session

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

func newResumeCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a parked session",
		Long: `Resumes a parked session (PARKED -> ACTIVE).

The session must be in PARKED state. Use 'ari session status' to check.
Use -s to specify a session ID if not resuming the current session.

Examples:
  ari session resume
  ari session resume -s session-20260105-143000-abc12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(ctx)
		},
	}

	return cmd
}

func runResume(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()
	fsm := session.NewFSM()

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		printer.PrintError(err)
		return err
	}

	// Acquire exclusive lock
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout, "ari-session-resume")
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer sessionLock.Release()
	emitLockEvent(resolver, sessionID, "ari-session-resume")

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		printer.PrintError(err)
		return err
	}

	// Validate transition
	if err := fsm.ValidateTransition(sessCtx.Status, session.StatusActive); err != nil {
		printer.PrintError(err)
		return err
	}

	// Update context
	now := time.Now().UTC()
	previousStatus := sessCtx.Status
	sessCtx.Status = session.StatusActive
	sessCtx.ResumedAt = &now
	sessCtx.ParkedAt = nil
	sessCtx.ParkedReason = ""

	// Save context
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Emit lifecycle event
	sessionDir := resolver.SessionDir(sessionID)
	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer writer.Close()
	writer.Write(clewcontract.NewSessionResumedEvent(sessionID))
	if err := writer.Flush(); err != nil {
		printer.VerboseLog("warn", "failed to write event", map[string]interface{}{"error": err.Error()})
	}

	// Output result
	result := output.TransitionOutput{
		SessionID:      sessionID,
		Status:         string(session.StatusActive),
		PreviousStatus: string(previousStatus),
		ResumedAt:      now.Format(time.RFC3339),
	}

	return printer.PrintSuccess(result)
}

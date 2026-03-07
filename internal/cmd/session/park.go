package session

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

type parkOptions struct {
	reason string
}

func newParkCmd(ctx *cmdContext) *cobra.Command {
	var opts parkOptions

	cmd := &cobra.Command{
		Use:   "park",
		Short: "Suspend the current session",
		Long: `Suspend the current session (ACTIVE -> PARKED).

Parked sessions preserve their state and can be resumed later
with 'ari session resume'. A reason can be provided for the audit log.
Rotates SESSION_CONTEXT.md to compact state before parking.

Examples:
  ari session park
  ari session park -r "switching to higher priority work"
  ari session park --reason "end of day"

Context:
  Lifecycle command -- invoke via Moirai, not specialists directly.
  Autopark hook calls this on CC session stop events.
  Use 'ari session resume' to reactivate. Use 'ari session wrap' to archive.
  Always provide --reason for audit trail clarity.
  Emits session.parked and session.end events to the backplane.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPark(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.reason, "reason", "r", "Manual park", "Reason for parking")

	return cmd
}

func runPark(ctx *cmdContext, opts parkOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()
	fsm := session.NewFSM()

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		return common.PrintAndReturn(printer, err)
	}

	// Acquire exclusive lock
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout, "ari-session-park")
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}
	defer func() { _ = sessionLock.Release() }()
	emitLockEvent(resolver, sessionID, "ari-session-park")

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		return common.PrintAndReturn(printer, err)
	}

	// Validate transition
	if err := fsm.ValidateTransition(sessCtx.Status, session.StatusParked); err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Update context
	now := time.Now().UTC()
	previousStatus := sessCtx.Status
	sessCtx.Status = session.StatusParked
	sessCtx.ParkedAt = &now
	sessCtx.ParkedReason = opts.reason
	sessCtx.ParkSource = "manual"

	// Rotate SESSION_CONTEXT before parking to compact state for later resumption
	sessionDir := resolver.SessionDir(sessionID)
	rotResult, rotErr := session.RotateSessionContext(sessionDir, session.DefaultMaxLines, session.DefaultKeepLines)
	if rotErr != nil {
		printer.VerboseLog("warn", "failed to rotate SESSION_CONTEXT on park", map[string]any{"error": rotErr.Error()})
	} else if rotResult.Rotated {
		printer.VerboseLog("info", "rotated SESSION_CONTEXT on park", map[string]any{
			"archived_lines": rotResult.ArchivedLines,
			"kept_lines":     rotResult.KeptLines,
		})
	}

	// Save context
	if err := sessCtx.Save(ctxPath); err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Emit Clew Contract events
	parkWriter := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = parkWriter.Close() }()
	{
		// Lifecycle event
		parkWriter.Write(clewcontract.NewSessionParkedEvent(sessionID, opts.reason))

		// Session end event
		durationMs := time.Since(sessCtx.CreatedAt).Milliseconds()
		parkWriter.Write(clewcontract.NewSessionEndEvent(sessionID, "parked", durationMs))

		if flushErr := parkWriter.Flush(); flushErr != nil {
			printer.VerboseLog("warn", "failed to write events", map[string]any{"error": flushErr.Error()})
		}
	}

	// Output result
	result := output.TransitionOutput{
		SessionID:      sessionID,
		Status:         string(session.StatusParked),
		PreviousStatus: string(previousStatus),
		ParkedAt:       now.Format(time.RFC3339),
		ParkedReason:   opts.reason,
	}

	return printer.PrintSuccess(result)
}


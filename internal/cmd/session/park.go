package session

import (
	"os/exec"
	"strings"
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
		Long: `Suspends the current session (ACTIVE -> PARKED).

Parked sessions preserve their state and can be resumed later
with 'ari session resume'. A reason can be provided for the audit log.

Examples:
  ari session park
  ari session park -r "switching to higher priority work"
  ari session park --reason "end of day"`,
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
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		printer.PrintError(err)
		return err
	}

	// Acquire exclusive lock
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout)
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer sessionLock.Release()

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
	if err := fsm.ValidateTransition(sessCtx.Status, session.StatusParked); err != nil {
		printer.PrintError(err)
		return err
	}

	// Get git status
	gitStatus := getGitStatus()

	// Update context
	now := time.Now().UTC()
	previousStatus := sessCtx.Status
	sessCtx.Status = session.StatusParked
	sessCtx.ParkedAt = &now
	sessCtx.ParkedReason = opts.reason

	// Save context
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Emit event
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitParked(sessionID, opts.reason, gitStatus); err != nil {
		printer.VerboseLog("warn", "failed to emit park event", map[string]interface{}{"error": err.Error()})
	}

	// Emit Clew Contract session_end event
	sessionDir := resolver.SessionDir(sessionID)
	tcWriter, err := clewcontract.NewEventWriter(sessionDir)
	if err == nil {
		durationMs := time.Since(sessCtx.CreatedAt).Milliseconds()
		sessionEndEvent := clewcontract.NewSessionEndEvent(sessionID, "parked", durationMs)
		if err := tcWriter.Write(sessionEndEvent); err != nil {
			printer.VerboseLog("warn", "failed to emit session_end event", map[string]interface{}{"error": err.Error()})
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

func getGitStatus() string {
	cmd := exec.Command("git", "status", "--short")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	if strings.TrimSpace(string(out)) == "" {
		return "clean"
	}
	return "uncommitted changes"
}

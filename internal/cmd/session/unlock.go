package session

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
)

type unlockOptions struct {
	force bool
}

func newUnlockCmd(ctx *cmdContext) *cobra.Command {
	var opts unlockOptions

	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Manually release a session lock",
		Long: `Manually releases a session lock. Use --force to remove locks held by other processes.

Useful for recovering from stale locks left by crashed processes.

Examples:
  ari session unlock
  ari session unlock --force    # Release lock held by another process`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlock(ctx, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force unlock even if not owner")

	return cmd
}

func runUnlock(ctx *cmdContext, opts unlockOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

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

	// Check session exists
	ctxPath := resolver.SessionContextFile(sessionID)
	if _, err := os.Stat(ctxPath); os.IsNotExist(err) {
		err := errors.ErrSessionNotFound(sessionID)
		printer.PrintError(err)
		return err
	}

	// Check if lock is held
	holderPID, err := lockMgr.GetHolder(sessionID)
	wasStale := false

	if err != nil {
		// No lock holder - nothing to unlock
		result := output.LockOutput{
			SessionID: sessionID,
			Unlocked:  true,
			WasStale:  false,
		}
		return printer.PrintSuccess(result)
	}

	// Check if we're the owner
	currentPID := os.Getpid()
	if holderPID != currentPID && !opts.force {
		err := errors.NewWithDetails(errors.CodeGeneralError,
			"not lock owner. Use --force to override.",
			map[string]interface{}{
				"holder_pid":  holderPID,
				"current_pid": currentPID,
			})
		printer.PrintError(err)
		return err
	}

	// Check if holder process is dead (stale)
	if holderPID > 0 {
		process, err := os.FindProcess(holderPID)
		if err != nil || process.Signal(nil) != nil {
			wasStale = true
		}
	}

	// Force release
	if err := lockMgr.ForceRelease(sessionID); err != nil {
		printer.PrintError(err)
		return err
	}

	// Emit event
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitLockReleased(sessionID); err != nil {
		printer.VerboseLog("warn", "failed to emit unlock event", map[string]interface{}{"error": err.Error()})
	}

	result := output.LockOutput{
		SessionID: sessionID,
		Unlocked:  true,
		WasStale:  wasStale,
	}

	return printer.PrintSuccess(result)
}

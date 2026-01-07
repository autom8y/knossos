package session

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
)

type lockOptions struct {
	timeout int
}

func newLockCmd(ctx *cmdContext) *cobra.Command {
	var opts lockOptions

	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Manually acquire a session lock",
		Long:  `Manually acquires an exclusive lock on a session. Primarily for debugging.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLock(ctx, opts)
		},
	}

	cmd.Flags().IntVarP(&opts.timeout, "timeout", "T", 10, "Lock acquisition timeout in seconds")

	return cmd
}

func runLock(ctx *cmdContext, opts lockOptions) error {
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

	// Acquire exclusive lock
	timeout := time.Duration(opts.timeout) * time.Second
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, timeout)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// NOTE: Lock is held until process exits or unlock is called
	// For manual lock command, we keep it until user does something

	now := time.Now().UTC()
	pid := os.Getpid()

	// Emit event
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitLockAcquired(sessionID, pid); err != nil {
		printer.VerboseLog("warn", "failed to emit lock event", map[string]interface{}{"error": err.Error()})
	}

	result := output.LockOutput{
		SessionID:  sessionID,
		Locked:     true,
		LockPath:   sessionLock.Path(),
		HolderPID:  pid,
		AcquiredAt: now.Format(time.RFC3339),
	}

	// For manual lock, we hold the lock until exit
	// Print result and wait for process to be terminated
	if err := printer.PrintSuccess(result); err != nil {
		sessionLock.Release()
		return err
	}

	// Hold lock indefinitely (user must Ctrl+C or use unlock --force)
	printer.VerboseLog("info", "Lock held. Press Ctrl+C to release.", nil)
	select {} // Block forever - lock released when process dies
}

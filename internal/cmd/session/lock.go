package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/lock"
)

const (
	moiraiLockFilename      = ".moirai-lock"
	staleAfterSeconds       = 300
	validAgent              = "moirai"
)

type lockOptions struct {
	agent string
}

func newLockCmd(ctx *cmdContext) *cobra.Command {
	var opts lockOptions

	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Acquire a Moirai lock for session state mutation",
		Long: `Acquires an exclusive advisory lock for Moirai to mutate session context files.

This lock is checked by the writeguard hook to allow Moirai to write to
protected *_CONTEXT.md files. The lock is session-scoped and stale after
300 seconds.

Examples:
  ari session lock --agent moirai`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLock(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.agent, "agent", "a", "", "Agent name (must be 'moirai')")
	cmd.MarkFlagRequired("agent")

	return cmd
}

func newUnlockCmd(ctx *cmdContext) *cobra.Command {
	var opts lockOptions

	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Release a Moirai lock for session state mutation",
		Long: `Releases the Moirai lock, allowing other operations to proceed.

The agent name must match the lock holder. Always unlock after completing
mutations, even on error.

Examples:
  ari session unlock --agent moirai`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlock(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.agent, "agent", "a", "", "Agent name (must match lock holder)")
	cmd.MarkFlagRequired("agent")

	return cmd
}

func runLock(ctx *cmdContext, opts lockOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Validate agent name
	if opts.agent != validAgent {
		err := errors.New(errors.CodeValidationFailed,
			"invalid agent name: only 'moirai' is currently supported")
		printer.PrintError(err)
		return err
	}

	// Get current session ID
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

	// Construct lock file path
	sessionDir := resolver.SessionDir(sessionID)
	lockPath := filepath.Join(sessionDir, moiraiLockFilename)

	// Check if lock already exists
	if existingLock, err := lock.ReadMoiraiLock(lockPath); err == nil {
		// Lock exists, check if stale
		if !lock.IsMoiraiLockStale(existingLock) {
			err := errors.New(errors.CodeLifecycleViolation,
				"lock already held by "+existingLock.Agent)
			printer.PrintError(err)
			return err
		}
		// Stale lock, will overwrite with warning
		printer.VerboseLog("warn", "overwriting stale lock", map[string]interface{}{
			"stale_age_seconds": time.Since(existingLock.AcquiredAt).Seconds(),
		})
	}

	// Create lock
	moiraiLock := lock.MoiraiLock{
		Agent:             opts.agent,
		AcquiredAt:        time.Now().UTC(),
		SessionID:         sessionID,
		StaleAfterSeconds: staleAfterSeconds,
	}

	// Write lock file
	data, err := json.MarshalIndent(moiraiLock, "", "  ")
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to marshal lock", err))
		return err
	}

	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to write lock file", err))
		return err
	}

	// Output success
	result := map[string]interface{}{
		"success":    true,
		"action":     "lock",
		"agent":      opts.agent,
		"session_id": sessionID,
	}

	return printer.PrintSuccess(result)
}

func runUnlock(ctx *cmdContext, opts lockOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Validate agent name
	if opts.agent != validAgent {
		err := errors.New(errors.CodeValidationFailed,
			"invalid agent name: only 'moirai' is currently supported")
		printer.PrintError(err)
		return err
	}

	// Get current session ID
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

	// Construct lock file path
	sessionDir := resolver.SessionDir(sessionID)
	lockPath := filepath.Join(sessionDir, moiraiLockFilename)

	// Read existing lock
	existingLock, err := lock.ReadMoiraiLock(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := errors.New(errors.CodeValidationFailed, "no lock exists to release")
			printer.PrintError(err)
			return err
		}
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to read lock file", err))
		return err
	}

	// Validate agent matches
	if existingLock.Agent != opts.agent {
		err := errors.New(errors.CodeValidationFailed,
			"lock held by different agent: "+existingLock.Agent)
		printer.PrintError(err)
		return err
	}

	// Remove lock file
	if err := os.Remove(lockPath); err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to remove lock file", err))
		return err
	}

	// Output success
	result := map[string]interface{}{
		"success":    true,
		"action":     "unlock",
		"agent":      opts.agent,
		"session_id": sessionID,
	}

	return printer.PrintSuccess(result)
}


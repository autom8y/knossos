package session

import (
	"github.com/autom8y/knossos/internal/cmd/common"
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
		Short: "Acquire Moirai advisory lock",
		Long: `Acquire an exclusive advisory lock for Moirai to mutate session context files.

This lock is checked by the writeguard hook to allow Moirai to write to
protected *_CONTEXT.md files. The lock is session-scoped and stale after
300 seconds. Stale locks are automatically overwritten with a warning.

Examples:
  ari session lock --agent moirai

Context:
  Only Moirai should call this. The writeguard hook checks this lock.
  Always pair with 'ari session unlock' after mutations, even on error.
  Lock auto-expires after 300 seconds to prevent deadlocks.
  Other agents should use 'ari session field-set' or lifecycle commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLock(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.agent, "agent", "a", "", "Agent name (must be 'moirai')")
	_ = cmd.MarkFlagRequired("agent")

	return cmd
}

func newUnlockCmd(ctx *cmdContext) *cobra.Command {
	var opts lockOptions

	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Release Moirai advisory lock",
		Long: `Release the Moirai lock, allowing other operations to proceed.

The agent name must match the lock holder. Always unlock after completing
mutations, even on error. Fails if no lock exists or agent does not match.

Examples:
  ari session unlock --agent moirai

Context:
  Only Moirai should call this. Always call after lock, even on error.
  Fails cleanly if lock was already released or expired.
  Use 'ari session recover' to clean up stale locks in bulk.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlock(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.agent, "agent", "a", "", "Agent name (must match lock holder)")
	_ = cmd.MarkFlagRequired("agent")

	return cmd
}

func runLock(ctx *cmdContext, opts lockOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Validate agent name
	if opts.agent != validAgent {
		err := errors.New(errors.CodeValidationFailed,
			"invalid agent name: only 'moirai' is currently supported")
		return common.PrintAndReturn(printer, err)
	}

	// Get current session ID
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		return common.PrintAndReturn(printer, err)
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
			return common.PrintAndReturn(printer, err)
		}
		// Stale lock, will overwrite with warning
		printer.VerboseLog("warn", "overwriting stale lock", map[string]any{
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
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to marshal lock", err))
	}

	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to write lock file", err))
	}

	// Output success
	result := map[string]any{
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
		return common.PrintAndReturn(printer, err)
	}

	// Get current session ID
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		return common.PrintAndReturn(printer, err)
	}

	// Construct lock file path
	sessionDir := resolver.SessionDir(sessionID)
	lockPath := filepath.Join(sessionDir, moiraiLockFilename)

	// Read existing lock
	existingLock, err := lock.ReadMoiraiLock(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := errors.New(errors.CodeValidationFailed, "no lock exists to release")
			return common.PrintAndReturn(printer, err)
		}
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to read lock file", err))
	}

	// Validate agent matches
	if existingLock.Agent != opts.agent {
		err := errors.New(errors.CodeValidationFailed,
			"lock held by different agent: "+existingLock.Agent)
		return common.PrintAndReturn(printer, err)
	}

	// Remove lock file
	if err := os.Remove(lockPath); err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to remove lock file", err))
	}

	// Output success
	result := map[string]any{
		"success":    true,
		"action":     "unlock",
		"agent":      opts.agent,
		"session_id": sessionID,
	}

	return printer.PrintSuccess(result)
}


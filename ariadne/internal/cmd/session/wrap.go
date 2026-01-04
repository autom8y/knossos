package session

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/lock"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
	"github.com/autom8y/ariadne/internal/session"
)

type wrapOptions struct {
	skipChecks bool
	noArchive  bool
}

func newWrapCmd(ctx *cmdContext) *cobra.Command {
	var opts wrapOptions

	cmd := &cobra.Command{
		Use:   "wrap",
		Short: "Complete and archive a session",
		Long:  `Completes a session, transitioning to ARCHIVED state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWrap(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.skipChecks, "skip-checks", false, "Skip quality gate checks")
	cmd.Flags().BoolVar(&opts.noArchive, "no-archive", false, "Don't move to archive directory")

	return cmd
}

func runWrap(ctx *cmdContext, opts wrapOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()
	lockMgr := ctx.getLockManager()
	fsm := session.NewFSM()

	sessionID, err := ctx.getSessionID()
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
	sessionDir := resolver.SessionDir(sessionID)
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
	if err := fsm.ValidateTransition(sessCtx.Status, session.StatusArchived); err != nil {
		printer.PrintError(err)
		return err
	}

	// Update context
	now := time.Now().UTC()
	previousStatus := sessCtx.Status
	sessCtx.Status = session.StatusArchived
	sessCtx.ArchivedAt = &now

	// Save context
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Clear current session
	if err := ctx.clearCurrentSessionID(); err != nil {
		printer.VerboseLog("warn", "failed to clear current session", map[string]interface{}{"error": err.Error()})
	}

	// Emit event
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitArchived(sessionID, string(previousStatus)); err != nil {
		printer.VerboseLog("warn", "failed to emit archive event", map[string]interface{}{"error": err.Error()})
	}

	// Move to archive if requested
	var archivePath string
	archived := false
	if !opts.noArchive {
		archiveDir := resolver.ArchiveDir()
		if err := paths.EnsureDir(archiveDir); err != nil {
			printer.VerboseLog("warn", "failed to create archive directory", map[string]interface{}{"error": err.Error()})
		} else {
			archivePath = archiveDir + "/" + sessionID
			// Only move if target doesn't exist
			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				if err := os.Rename(sessionDir, archivePath); err != nil {
					printer.VerboseLog("warn", "failed to move to archive", map[string]interface{}{"error": err.Error()})
				} else {
					archived = true
				}
			}
		}
	}

	// Output result
	result := output.TransitionOutput{
		SessionID:      sessionID,
		Status:         string(session.StatusArchived),
		PreviousStatus: string(previousStatus),
		ArchivedAt:     now.Format(time.RFC3339),
		Archived:       archived,
		ArchivePath:    archivePath,
	}

	return printer.PrintSuccess(result)
}

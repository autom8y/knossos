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

type createOptions struct {
	complexity string
	team       string
}

func newCreateCmd(ctx *cmdContext) *cobra.Command {
	var opts createOptions

	cmd := &cobra.Command{
		Use:   "create <initiative>",
		Short: "Create a new session",
		Long:  `Create a new session, transitioning from NONE to ACTIVE state.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(ctx, args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.complexity, "complexity", "c", "MODULE",
		"Complexity level: PATCH, MODULE, SYSTEM, INITIATIVE, MIGRATION")
	cmd.Flags().StringVarP(&opts.team, "team", "t", "",
		"Team pack to activate (default: from ACTIVE_TEAM)")

	return cmd
}

func runCreate(ctx *cmdContext, initiative string, opts createOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()
	lockMgr := ctx.getLockManager()

	// Get team from flag or ACTIVE_TEAM file
	team := opts.team
	if team == "" {
		team = ctx.getActiveTeam()
	}

	// Validate complexity
	if !isValidComplexity(opts.complexity) {
		err := errors.New(errors.CodeUsageError, "invalid complexity: must be PATCH, MODULE, SYSTEM, INITIATIVE, or MIGRATION")
		printer.PrintError(err)
		return err
	}

	// Ensure sessions directory exists
	if err := paths.EnsureDir(resolver.SessionsDir()); err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to create sessions directory", err)
		printer.PrintError(err)
		return err
	}

	// Acquire exclusive lock for creation (using a special "create" lock)
	createLock, err := lockMgr.Acquire("__create__", lock.Exclusive, lock.DefaultTimeout)
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer createLock.Release()

	// Check for existing session
	currentID, err := ctx.getCurrentSessionID()
	if err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to read current session", err)
		printer.PrintError(err)
		return err
	}

	if currentID != "" {
		// Check if session directory exists
		if _, err := os.Stat(resolver.SessionDir(currentID)); err == nil {
			// Load session to check status
			ctxPath := resolver.SessionContextFile(currentID)
			existingCtx, err := session.LoadContext(ctxPath)
			if err == nil {
				err := errors.ErrSessionExists(currentID, string(existingCtx.Status))
				printer.PrintError(err)
				return err
			}
		}
	}

	// Create new session context
	newCtx := session.NewContext(initiative, opts.complexity, team)
	sessionDir := resolver.SessionDir(newCtx.SessionID)

	// Create session directory
	if err := paths.EnsureDir(sessionDir); err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to create session directory", err)
		printer.PrintError(err)
		return err
	}

	// Save context
	ctxPath := resolver.SessionContextFile(newCtx.SessionID)
	if err := newCtx.Save(ctxPath); err != nil {
		// Cleanup on failure
		os.RemoveAll(sessionDir)
		printer.PrintError(err)
		return err
	}

	// Ensure audit directory exists
	if err := paths.EnsureDir(resolver.AuditDir()); err != nil {
		// Non-fatal, but log
		printer.VerboseLog("warn", "failed to create audit directory", nil)
	}

	// Emit creation event
	emitter := ctx.getEventEmitter(newCtx.SessionID)
	if err := emitter.EmitCreated(newCtx.SessionID, initiative, opts.complexity, team); err != nil {
		// Non-fatal
		printer.VerboseLog("warn", "failed to emit creation event", map[string]interface{}{"error": err.Error()})
	}

	// Set as current session
	if err := ctx.setCurrentSessionID(newCtx.SessionID); err != nil {
		printer.VerboseLog("warn", "failed to set current session", map[string]interface{}{"error": err.Error()})
	}

	// Output result
	result := output.CreateOutput{
		SessionID:     newCtx.SessionID,
		SessionDir:    sessionDir,
		Status:        string(newCtx.Status),
		Initiative:    initiative,
		Complexity:    opts.complexity,
		Team:          team,
		CreatedAt:     newCtx.CreatedAt.Format(time.RFC3339),
		SchemaVersion: newCtx.SchemaVersion,
	}

	return printer.Print(result)
}

func isValidComplexity(c string) bool {
	switch c {
	case "PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION":
		return true
	default:
		return false
	}
}

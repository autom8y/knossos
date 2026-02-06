// Package session implements the ari session commands.
package session

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// cmdContext holds shared state for session commands.
type cmdContext struct {
	common.SessionContext
}

// NewSessionCmd creates the session command group.
func NewSessionCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     outputFlag,
				Verbose:    verboseFlag,
				ProjectDir: projectDir,
			},
			SessionID: sessionID,
		},
	}

	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage workflow sessions",
		Long: `Create, list, park, resume, wrap, and manage Claude Code workflow sessions.

Session lifecycle: NONE -> ACTIVE -> {PARKED, ARCHIVED}
  PARKED sessions can be resumed back to ACTIVE.
  ARCHIVED is terminal.

Examples:
  ari session create "user-auth feature" -c MODULE
  ari session status
  ari session park -r "switching context"
  ari session resume
  ari session wrap`,
	}

	// Add subcommands
	cmd.AddCommand(newCreateCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newParkCmd(ctx))
	cmd.AddCommand(newResumeCmd(ctx))
	cmd.AddCommand(newWrapCmd(ctx))
	cmd.AddCommand(newTransitionCmd(ctx))
	cmd.AddCommand(newMigrateCmd(ctx))
	cmd.AddCommand(newAuditCmd(ctx))
	cmd.AddCommand(newRecoverCmd(ctx))

	// Session commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// getEventEmitter creates an event emitter for a session.
func (c *cmdContext) getEventEmitter(sessionID string) *session.EventEmitter {
	return c.GetEventEmitter(sessionID)
}

// getActiveRite reads the active rite from ACTIVE_RITE file.
func (c *cmdContext) getActiveRite() string {
	ritePath := c.GetResolver().ActiveRiteFile()
	if data, err := os.ReadFile(ritePath); err == nil {
		return string(data)
	}
	return "none"
}

// Package session implements the ari session commands.
package session

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
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
		Long: `Create, list, park, resume, wrap, fray, recover, and manage Claude Code workflow sessions.

Session lifecycle: NONE -> ACTIVE -> {PARKED, ARCHIVED}
  PARKED sessions can be resumed back to ACTIVE.
  ARCHIVED is terminal.

Use 'ari session recover' to clean up stale locks and rebuild the session cache.

Examples:
  ari session create "user-auth feature" -c MODULE
  ari session status
  ari session park -r "switching context"
  ari session resume
  ari session wrap
  ari session recover --dry-run`,
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
	cmd.AddCommand(newFrayCmd(ctx))
	cmd.AddCommand(newLockCmd(ctx))
	cmd.AddCommand(newUnlockCmd(ctx))
	cmd.AddCommand(newGcCmd(ctx))

	// Session commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}


// getActiveRite reads the active rite from ACTIVE_RITE file.
// Returns "none" as a fallback if the file doesn't exist or is empty.
func (c *cmdContext) getActiveRite() string {
	rite := c.GetResolver().ReadActiveRite()
	if rite == "" {
		return "none"
	}
	return rite
}

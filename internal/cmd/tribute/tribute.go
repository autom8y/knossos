// Package tribute implements the ari tribute commands for session summary generation.
package tribute

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for tribute commands.
type cmdContext struct {
	common.SessionContext
}

// NewTributeCmd creates the tribute command group.
func NewTributeCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
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
		Use:   "tribute",
		Short: "Session summary operations (Minos Tribute)",
		Long: `Generate and manage TRIBUTE.md session summaries.

In the Knossos mythology, King Minos demanded tribute from Athens; in our system,
TRIBUTE.md is the "payment" for navigating the labyrinth--a comprehensive record
of what was accomplished, what decisions were made, and what artifacts were produced.

TRIBUTE.md serves as both human-readable documentation of a completed session
and machine-parseable metadata for analytics and future context loading.`,
	}

	// Add subcommands
	cmd.AddCommand(newGenerateCmd(ctx))

	// Tribute commands do NOT require project context (can specify session-dir)
	common.SetNeedsProject(cmd, false, true)
	common.SetGroupCommand(cmd)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

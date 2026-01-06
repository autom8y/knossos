// Package tribute implements the ari tribute commands for session summary generation.
package tribute

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/output"
)

// cmdContext holds shared state for tribute commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
	sessionID  *string
}

// NewTributeCmd creates the tribute command group.
func NewTributeCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
		sessionID:  sessionID,
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

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	format := output.FormatText
	if c.output != nil {
		format = output.ParseFormat(*c.output)
	}
	verbose := false
	if c.verbose != nil {
		verbose = *c.verbose
	}
	return output.NewPrinter(format, os.Stdout, os.Stderr, verbose)
}

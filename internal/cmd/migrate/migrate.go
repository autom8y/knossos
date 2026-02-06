// Package migrate implements the ari migrate commands.
package migrate

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for migrate commands.
type cmdContext struct {
	common.BaseContext
}

// NewMigrateCmd creates the migrate command group.
func NewMigrateCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run platform migrations",
		Long:  "Run platform migrations to update project configuration.",
	}

	// Add subcommands
	cmd.AddCommand(newRosterToKnossosCmd(ctx))

	// Migrate commands do not require project context by default
	// (they work at user-level ~/.claude)
	common.SetNeedsProject(cmd, false, true)

	return cmd
}

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

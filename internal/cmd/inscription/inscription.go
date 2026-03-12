// Package inscription implements the ari inscription commands.
// These commands manage the context file inscription system for the Knossos platform.
package inscription

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/inscription"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for inscription commands.
type cmdContext struct {
	common.BaseContext
}

// NewInscriptionCmd creates the inscription command group.
func NewInscriptionCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "inscription",
		Short: "Manage context file inscription system",
		Long: `Manage the context file inscription system for the Knossos platform.

The inscription system synchronizes context file content with templates and
project state, managing ownership of different regions:

  - knossos: Managed by Knossos templates, always synced
  - satellite: Owned by satellite project, never overwritten
  - regenerate: Generated from project state (ACTIVE_RITE, agents/)

Examples:
  ari inscription sync              # Sync context file with templates
  ari inscription sync --dry-run    # Preview changes without writing
  ari inscription validate          # Check manifest and context file
  ari inscription backups           # List available backups
  ari inscription rollback          # Restore from backup`,
	}

	// Add subcommands
	cmd.AddCommand(newSyncCmd(ctx))
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newRollbackCmd(ctx))
	cmd.AddCommand(newBackupsCmd(ctx))
	cmd.AddCommand(newDiffCmd(ctx))

	// Inscription commands require project context
	common.SetNeedsProject(cmd, true, true)
	common.SetGroupCommand(cmd)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// getPipeline creates an inscription pipeline from the context.
func (c *cmdContext) getPipeline() *inscription.Pipeline {
	resolver := c.GetResolver()
	return inscription.NewPipeline(resolver.ProjectRoot())
}

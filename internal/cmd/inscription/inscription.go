// Package inscription implements the ari inscription commands.
// These commands manage the CLAUDE.md inscription system for the Knossos platform.
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
		Short: "Manage CLAUDE.md inscription system",
		Long: `Manage the CLAUDE.md inscription system for the Knossos platform.

The inscription system synchronizes CLAUDE.md content with templates and
project state, managing ownership of different regions:

  - knossos: Managed by Knossos templates, always synced
  - satellite: Owned by satellite project, never overwritten
  - regenerate: Generated from project state (ACTIVE_RITE, agents/)

Examples:
  ari inscription sync              # Sync CLAUDE.md with templates
  ari inscription sync --dry-run    # Preview changes without writing
  ari inscription validate          # Check manifest and CLAUDE.md
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

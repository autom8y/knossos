// Package registry implements the ari registry commands for managing the
// org-level knowledge domain catalog.
package registry

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for registry commands.
type cmdContext struct {
	common.BaseContext
}

// NewRegistryCmd creates the registry command group.
func NewRegistryCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage the org knowledge domain registry",
		Long: `Sync, list, and inspect the cross-repo knowledge domain catalog.

The registry catalogs .know/ domains from GitHub repositories in your org,
enabling cross-repo knowledge discovery for Clew queries.

Examples:
  ari registry sync                    # Sync all repos for active org
  ari registry sync --org autom8y      # Sync for a specific org
  ari registry list                    # List all cataloged domains
  ari registry list --repo knossos     # Filter by repo
  ari registry list --stale            # Show only stale domains
  ari registry status                  # Show sync summary`,
	}

	cmd.AddCommand(newSyncCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))

	// Registry commands do NOT require a project context.
	common.SetNeedsProject(cmd, false, true)
	common.SetGroupCommand(cmd)

	return cmd
}

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

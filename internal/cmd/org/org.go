// Package org implements the ari org commands for managing organizations.
package org

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for org commands.
type cmdContext struct {
	common.BaseContext
}

// NewOrgCmd creates the org command group.
func NewOrgCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage organizations",
		Long: `Create, configure, and manage organization-level resources.

Organizations provide shared rites, agents, and mena (commands + skills)
across multiple projects for a team.

Examples:
  ari org init autom8y          # Bootstrap org directory
  ari org set autom8y           # Set active org
  ari org list                  # List available orgs
  ari org current               # Show active org`,
	}

	cmd.AddCommand(newInitCmd(ctx))
	cmd.AddCommand(newSetCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newCurrentCmd(ctx))

	// Org commands do NOT require a project context
	common.SetNeedsProject(cmd, false, true)

	return cmd
}

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

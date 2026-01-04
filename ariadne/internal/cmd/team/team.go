// Package team implements the ari team commands.
package team

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
	"github.com/autom8y/ariadne/internal/team"
)

// cmdContext holds shared state for team commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
}

// NewTeamCmd creates the team command group.
func NewTeamCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
	}

	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage agent teams",
		Long:  `List, switch, validate, and manage agent team packs.`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newSwitchCmd(ctx))
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newContextCmd(ctx))
	cmd.AddCommand(newMigrateContextCmd(ctx))

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

// getResolver creates a path resolver from the context.
func (c *cmdContext) getResolver() *paths.Resolver {
	projectDir := ""
	if c.projectDir != nil {
		projectDir = *c.projectDir
	}
	return paths.NewResolver(projectDir)
}

// getDiscovery creates a team discovery instance.
func (c *cmdContext) getDiscovery() *team.Discovery {
	resolver := c.getResolver()
	return team.NewDiscovery(resolver)
}

// getSwitcher creates a team switcher.
func (c *cmdContext) getSwitcher() *team.Switcher {
	resolver := c.getResolver()
	return team.NewSwitcher(resolver)
}

// getValidator creates a team validator.
func (c *cmdContext) getValidator() *team.Validator {
	resolver := c.getResolver()
	return team.NewValidator(resolver)
}

// getActiveTeam reads the active team from ACTIVE_TEAM file.
func (c *cmdContext) getActiveTeam() string {
	resolver := c.getResolver()
	data, err := os.ReadFile(resolver.ActiveTeamFile())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

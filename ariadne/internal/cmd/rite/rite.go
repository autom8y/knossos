// Package rite implements the ari rite commands.
package rite

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
	"github.com/autom8y/ariadne/internal/rite"
)

// cmdContext holds shared state for rite commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
}

// NewRiteCmd creates the rite command group.
func NewRiteCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
	}

	cmd := &cobra.Command{
		Use:   "rite",
		Short: "Manage rite invocations and composition",
		Long: `List, invoke, release, and manage rite partial composition.

Rites are composable practice bundles. The invoke operation is additive
(borrow components without switching), while swap is replacement
(same as team switch).`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newInfoCmd(ctx))
	cmd.AddCommand(newCurrentCmd(ctx))
	cmd.AddCommand(newInvokeCmd(ctx))
	cmd.AddCommand(newReleaseCmd(ctx))
	cmd.AddCommand(newSwapCmd(ctx))

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

// getDiscovery creates a rite discovery instance.
func (c *cmdContext) getDiscovery() *rite.Discovery {
	resolver := c.getResolver()
	return rite.NewDiscovery(resolver)
}

// getInvoker creates a rite invoker.
func (c *cmdContext) getInvoker() *rite.Invoker {
	resolver := c.getResolver()
	return rite.NewInvoker(resolver)
}

// getActiveRite reads the active rite from ACTIVE_RITE file with backward compatibility.
func (c *cmdContext) getActiveRite() string {
	resolver := c.getResolver()

	// Try new ACTIVE_RITE first
	ritePath := resolver.ActiveRiteFile()
	if data, err := os.ReadFile(ritePath); err == nil {
		return strings.TrimSpace(string(data))
	} else if os.IsNotExist(err) {
		// Fall back to legacy ACTIVE_TEAM file
		if data, err := os.ReadFile(resolver.ActiveTeamFile()); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	return ""
}

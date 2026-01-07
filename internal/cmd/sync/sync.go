// Package sync implements the ari sync commands.
package sync

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// cmdContext holds shared state for sync commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
}

// NewSyncCmd creates the sync command group.
func NewSyncCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
	}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize configuration with remotes",
		Long: `Synchronize .claude/ configuration with remote sources.

Sync tracks changes to configuration files and enables pulling updates
from remotes and pushing local changes back.

Supported remotes:
  - Local paths: /path/to/source, ./relative/path
  - HTTP(S) URLs: https://example.com/config
  - GitHub shorthand: org/repo
  - Git refs: HEAD:.claude/path, origin/main:.claude/path`,
	}

	// Add subcommands
	cmd.AddCommand(newMaterializeCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newPullCmd(ctx))
	cmd.AddCommand(newPushCmd(ctx))
	cmd.AddCommand(newDiffCmd(ctx))
	cmd.AddCommand(newResolveCmd(ctx))
	cmd.AddCommand(newHistoryCmd(ctx))
	cmd.AddCommand(newResetCmd(ctx))

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

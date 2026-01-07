// Package naxos implements the ari naxos commands.
// Naxos is the cleanup mechanism for abandoned sessions in the Knossos platform.
// Named after the island where Theseus abandoned Ariadne in Greek mythology.
package naxos

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// cmdContext holds shared state for naxos commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
}

// NewNaxosCmd creates the naxos command group.
func NewNaxosCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
	}

	cmd := &cobra.Command{
		Use:   "naxos",
		Short: "Cleanup tooling for abandoned sessions",
		Long: `Naxos provides cleanup tooling for abandoned sessions.

Named after the island where Theseus abandoned Ariadne in Greek mythology,
Naxos identifies sessions that may need cleanup attention:

  - Inactive sessions: No activity for extended period
  - Stale sails: Gray sails that haven't been upgraded
  - Incomplete wraps: Sessions marked for wrap but never completed

IMPORTANT: Naxos is report-only. It suggests actions but does not
automatically clean up sessions. The user determines what action to take.

Examples:
  ari naxos scan                       # Scan for orphaned sessions
  ari naxos scan --inactive-threshold 48h  # Custom inactivity threshold
  ari naxos scan --json                # JSON output for scripting`,
	}

	// Add subcommands
	cmd.AddCommand(newScanCmd(ctx))

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

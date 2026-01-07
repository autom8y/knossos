// Package naxos implements the ari naxos commands.
// Naxos is the cleanup mechanism for abandoned sessions in the Knossos platform.
// Named after the island where Theseus abandoned Ariadne in Greek mythology.
package naxos

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for naxos commands.
type cmdContext struct {
	common.BaseContext
}

// NewNaxosCmd creates the naxos command group.
func NewNaxosCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
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

	// Naxos commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

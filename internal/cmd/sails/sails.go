// Package sails implements the ari sails commands for White Sails quality gates.
package sails

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for sails commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
	sessionID  *string
}

// NewSailsCmd creates the sails command group.
func NewSailsCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
		sessionID:  sessionID,
	}

	cmd := &cobra.Command{
		Use:   "sails",
		Short: "White Sails quality gate operations",
		Long: `Manage White Sails quality gates for session confidence signaling.

White Sails provides typed contracts declaring computed confidence levels:
  WHITE - High confidence, ship without QA
  GRAY  - Unknown confidence, needs QA review
  BLACK - Known failure, do not ship

Quality gate checks return exit code 0 for WHITE, non-zero for GRAY/BLACK.`,
	}

	// Add subcommands
	cmd.AddCommand(newCheckCmd(ctx))

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

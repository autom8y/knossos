// Package sails implements the ari sails commands for White Sails quality gates.
package sails

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for sails commands.
type cmdContext struct {
	common.SessionContext
}

// NewSailsCmd creates the sails command group.
func NewSailsCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     outputFlag,
				Verbose:    verboseFlag,
				ProjectDir: projectDir,
			},
			SessionID: sessionID,
		},
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

	// Sails commands do NOT require project context (can check arbitrary paths)
	common.SetNeedsProject(cmd, false, true)
	common.SetGroupCommand(cmd)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

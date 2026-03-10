// Package procession implements the ari procession commands for cross-rite workflow management.
package procession

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for procession commands.
type cmdContext struct {
	common.SessionContext
}

// NewProcessionCmd creates the procession command group.
func NewProcessionCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
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
		Use:   "procession",
		Short: "Manage cross-rite coordinated workflows",
		Long: `Manage processions — template-defined, station-based workflows that coordinate
work across multiple rites. Each procession lives within a session and tracks
progress through an ordered sequence of rite-scoped stations.

Commands:
  list     - List available procession templates
  create   - Start a new procession from a template
  status   - Show current procession state
  proceed  - Advance to the next station
  recede   - Move back to an earlier station
  abandon  - Terminate the procession (session continues)

Examples:
  ari procession create --template=security-remediation
  ari procession status
  ari procession proceed --artifacts=.sos/wip/sr/HANDOFF-audit-to-assess.md
  ari procession recede --to=remediate
  ari procession abandon`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newCreateCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newProceedCmd(ctx))
	cmd.AddCommand(newRecedeCmd(ctx))
	cmd.AddCommand(newAbandonCmd(ctx))

	// Procession commands require project context
	common.SetNeedsProject(cmd, true, true)
	common.SetGroupCommand(cmd)

	return cmd
}

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

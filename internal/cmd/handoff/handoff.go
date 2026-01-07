// Package handoff implements the ari handoff commands for agent handoff management.
package handoff

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for handoff commands.
type cmdContext struct {
	common.SessionContext
}

// NewHandoffCmd creates the handoff command group.
func NewHandoffCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
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
		Use:   "handoff",
		Short: "Manage agent handoffs between workflow phases",
		Long: `Manage handoffs between agents during workflow execution.

Handoffs transfer work from one agent to another within a session,
ensuring proper artifact validation and context preservation.

Commands:
  prepare  - Validate readiness and emit task_end event
  execute  - Trigger transition and emit task_start event
  status   - Query current handoff state
  history  - Query handoff events from events.jsonl

Examples:
  ari handoff prepare --from=architect --to=principal-engineer
  ari handoff execute --artifact=TDD-user-auth --to=principal-engineer
  ari handoff status
  ari handoff history --limit=10`,
	}

	// Add subcommands
	cmd.AddCommand(newPrepareCmd(ctx))
	cmd.AddCommand(newExecuteCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newHistoryCmd(ctx))

	// Handoff commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

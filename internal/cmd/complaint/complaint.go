// Package complaint implements the ari complaint commands.
package complaint

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for complaint commands.
type cmdContext struct {
	common.BaseContext
}

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// NewComplaintCmd creates the complaint command group.
func NewComplaintCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "complaint",
		Short: "Manage complaints",
		Long: `View and manage Cassandra complaint artifacts from .sos/wip/complaints/.

Complaints are structured YAML files filed by agents when they encounter
framework friction. Use 'ari complaint list' to view filed complaints.

Examples:
  ari complaint list
  ari complaint list --severity=high
  ari complaint list --status=filed
  ari complaint list -o json`,
	}

	cmd.AddCommand(newListCmd(ctx))

	// Complaint commands do not require project context (complaints dir
	// is resolved relative to project, but gracefully empty if missing).
	common.SetNeedsProject(cmd, false, true)
	common.SetGroupCommand(cmd)

	return cmd
}

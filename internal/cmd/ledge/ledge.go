// Package ledge implements the ari ledge commands for work product management.
package ledge

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// cmdContext holds shared state for ledge commands.
type cmdContext struct {
	common.BaseContext
}

// NewLedgeCmd creates the ledge command group.
func NewLedgeCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "ledge",
		Short: "Manage work product artifacts",
		Long:  `Promote, list, and manage work product artifacts in the ledge.`,
	}

	cmd.AddCommand(newPromoteCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))

	common.SetNeedsProject(cmd, true, true)

	return cmd
}

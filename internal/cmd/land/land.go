// Package land implements the ari land commands for cross-session knowledge synthesis.
package land

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// cmdContext holds shared state for land commands.
type cmdContext struct {
	common.BaseContext
}

// NewLandCmd creates the land command group.
func NewLandCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "land",
		Short: "Manage cross-session knowledge synthesis",
		Long:  `Synthesize and manage persistent knowledge from session archives.`,
	}

	cmd.AddCommand(newSynthesizeCmd(ctx))

	common.SetNeedsProject(cmd, true, true)

	return cmd
}

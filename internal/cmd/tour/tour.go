package tour

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

type cmdContext struct {
	common.BaseContext
}

// NewTourCmd creates the ari tour command.
func NewTourCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "tour",
		Short: "Walk project directory structure",
		Long: `Display the knossos directory tree with file counts and contents.

Shows each managed directory (.claude/, .knossos/, .know/, .ledge/, .sos/)
with subdirectory listings and file counts from the live filesystem.

This is a read-only command -- it does not modify any state.

Examples:
  ari tour              # Human-readable directory tour
  ari tour -o json      # Machine-readable JSON output`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.GetPrinter(output.FormatText)
			resolver := ctx.GetResolver()

			tour := collectTour(resolver)
			return printer.Print(tour)
		},
	}

	common.SetNeedsProject(cmd, true, true)
	return cmd
}

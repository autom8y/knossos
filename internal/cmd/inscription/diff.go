package inscription

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func newDiffCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [region]",
		Short: "Show differences between current and generated content",
		Long: `Show differences between the current context file and what would be generated.

If a region name is provided, shows the diff for just that region.
Otherwise, shows diffs for all non-satellite regions.

This is useful for understanding what a sync operation would change
before running 'ari inscription sync'.

Examples:
  ari inscription diff                  # Diff all regions
  ari inscription diff quick-start      # Diff specific region
  ari inscription diff execution-mode   # Diff execution-mode region`,
		Args: common.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			regionName := ""
			if len(args) > 0 {
				regionName = args[0]
			}
			return runDiff(ctx, regionName)
		},
	}

	return cmd
}

func runDiff(ctx *cmdContext, regionName string) error {
	printer := ctx.getPrinter()
	pipeline := ctx.getPipeline()

	diff, err := pipeline.GetDiff(regionName)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	if diff == "" {
		printer.PrintLine("No differences found")
		return nil
	}

	out := DiffOutput{
		Region: regionName,
		Diff:   diff,
	}

	return printer.Print(out)
}

// DiffOutput represents diff result for output.
type DiffOutput struct {
	Region string `json:"region,omitempty"`
	Diff   string `json:"diff"`
}

// Text implements output.Textable for DiffOutput.
func (d DiffOutput) Text() string {
	return d.Diff
}

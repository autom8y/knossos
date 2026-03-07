package rite

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	ritelib "github.com/autom8y/knossos/internal/rite"
)

type releaseOptions struct {
	all    bool
	dryRun bool
}

func newReleaseCmd(ctx *cmdContext) *cobra.Command {
	var opts releaseOptions

	cmd := &cobra.Command{
		Use:   "release [name|invocation-id]",
		Short: "Release borrowed components",
		Long: `Releases borrowed components from a previous invocation.

Examples:
  ari rite release documentation           # Release specific rite
  ari rite release --all                   # Release everything borrowed
  ari rite release inv-20260106-abc123     # Release by invocation ID`,
		Args: common.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}
			return runRelease(ctx, target, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.all, "all", false, "Release all borrowed components")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview cleanup without applying")

	return cmd
}

func runRelease(ctx *cmdContext, target string, opts releaseOptions) error {
	printer := ctx.getPrinter()
	invoker := ctx.getInvoker()

	releaseOpts := ritelib.ReleaseOptions{
		Target: target,
		All:    opts.all,
		DryRun: opts.dryRun,
	}

	result, err := invoker.Release(releaseOpts)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	out := output.RiteReleaseOutput{
		ReleasedRites:      result.ReleasedRites,
		ReleasedSkills:     result.ReleasedSkills,
		ReleasedAgents:     result.ReleasedAgents,
		InvocationCount:    result.InvocationCount,
		TokensFreed:        result.TokensFreed,
		InscriptionUpdated: result.InscriptionUpdated,
		DryRun:             result.DryRun,
	}

	if opts.dryRun {
		return printer.Print(out)
	}

	return printer.PrintSuccess(out)
}

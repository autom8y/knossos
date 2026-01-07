package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/sync"
)

type diffOptions struct {
	showFull bool
}

func newDiffCmd(ctx *cmdContext) *cobra.Command {
	var opts diffOptions

	cmd := &cobra.Command{
		Use:   "diff [path]",
		Short: "Show local vs remote differences",
		Long: `Shows differences between local and remote configuration.

Without arguments, shows summary of all changed files.
With a path argument, shows detailed diff for that file.`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runDiff(ctx, path, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.showFull, "full", false, "Show full content (not just summary)")

	return cmd
}

func runDiff(ctx *cmdContext, path string, opts diffOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	differ := sync.NewDiffer(resolver)
	diffOpts := sync.FileDiffOptions{
		Path:     path,
		ShowFull: opts.showFull,
	}

	result, err := differ.Diff(diffOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	out := output.SyncDiffOutput{
		Path:         result.Path,
		HasChanges:   result.HasChanges,
		UnifiedDiff:  result.UnifiedDiff,
		Additions:    result.Additions,
		Deletions:    result.Deletions,
		TotalFiles:   result.TotalFiles,
		ChangedFiles: result.ChangedFiles,
	}

	if opts.showFull {
		out.LocalContent = result.LocalContent
		out.RemoteContent = result.RemoteContent
	}

	return printer.Print(out)
}

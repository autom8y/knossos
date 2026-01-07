package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/sync"
)

type resolveOptions struct {
	strategy string
	dryRun   bool
}

func newResolveCmd(ctx *cmdContext) *cobra.Command {
	var opts resolveOptions

	cmd := &cobra.Command{
		Use:   "resolve [path]",
		Short: "Resolve sync conflicts",
		Long: `Resolves sync conflicts using the specified strategy.

Without a path, resolves all conflicts.
With a path, resolves only that specific conflict.

Strategies:
  ours    Keep local changes, discard remote
  theirs  Accept remote changes, discard local
  merge   Attempt three-way merge (JSON files only)`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runResolve(ctx, path, opts)
		},
	}

	cmd.Flags().StringVar(&opts.strategy, "strategy", "ours", "Resolution strategy: ours, theirs, merge")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview resolution without applying")

	return cmd
}

func runResolve(ctx *cmdContext, path string, opts resolveOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()

	// Parse strategy
	var strategy sync.ResolveStrategy
	switch opts.strategy {
	case "ours":
		strategy = sync.ResolveOurs
	case "theirs":
		strategy = sync.ResolveTheirs
	case "merge":
		strategy = sync.ResolveMerge
	default:
		err := errors.New(errors.CodeUsageError, "Invalid strategy. Use: ours, theirs, or merge")
		printer.PrintError(err)
		return err
	}

	// Create history manager
	historyManager := sync.NewHistoryManager(resolver)

	// Perform resolution
	syncResolver := sync.NewResolver(resolver)
	resolveOpts := sync.ResolveOptions{
		Strategy: strategy,
		Path:     path,
		DryRun:   opts.dryRun,
	}

	result, err := syncResolver.Resolve(resolveOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	out := output.SyncResolveOutput{
		Path:           result.Path,
		Strategy:       result.Strategy,
		Resolved:       result.Resolved,
		RemainingCount: result.RemainingCount,
		Remaining:      result.Remaining,
	}

	// Record in history
	if result.Resolved && !opts.dryRun {
		historyManager.RecordResolve(path, opts.strategy, true)
	}

	return printer.Print(out)
}

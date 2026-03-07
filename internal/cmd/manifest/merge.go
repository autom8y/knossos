package manifest

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/manifest"
	"github.com/autom8y/knossos/internal/output"
)

type mergeOptions struct {
	outputPath string
	strategy   string
	dryRun     bool
}

func newMergeCmd(ctx *cmdContext) *cobra.Command {
	var opts mergeOptions

	cmd := &cobra.Command{
		Use:   "merge <base> <ours> <theirs>",
		Short: "Three-way merge of manifests",
		Long: `Performs three-way merge of manifest files with conflict detection.

Arguments:
  base    Common ancestor manifest
  ours    Our version (local changes)
  theirs  Their version (remote/incoming changes)

Strategies:
  smart   Field-level merge (default)
  ours    Prefer our changes on conflict
  theirs  Prefer their changes on conflict
  union   Merge arrays with union (no duplicates)`,
		Args:         common.ExactArgs(3),
		SilenceUsage: true, // Don't print usage on errors
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMerge(ctx, args[0], args[1], args[2], opts)
		},
	}

	cmd.Flags().StringVar(&opts.outputPath, "write-to", "", "Output path for merged manifest (default: stdout)")
	cmd.Flags().StringVar(&opts.strategy, "strategy", "smart", "Merge strategy: smart, ours, theirs, union")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview merge without writing")

	return cmd
}

func runMerge(ctx *cmdContext, basePath, oursPath, theirsPath string, opts mergeOptions) error {
	printer := ctx.getPrinter()

	// Load all three manifests
	base, err := manifest.Load(basePath)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	ours, err := manifest.Load(oursPath)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	theirs, err := manifest.Load(theirsPath)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Parse strategy
	var strategy manifest.MergeStrategy
	switch opts.strategy {
	case "ours":
		strategy = manifest.StrategyOurs
	case "theirs":
		strategy = manifest.StrategyTheirs
	case "union":
		strategy = manifest.StrategyUnion
	case "smart":
		strategy = manifest.StrategySmart
	default:
		err := errors.New(errors.CodeUsageError, "Invalid strategy: "+opts.strategy)
		return common.PrintAndReturn(printer, err)
	}

	// Perform merge
	mergeOpts := manifest.ManifestMergeOptions{
		Strategy: strategy,
		DryRun:   opts.dryRun,
	}
	result, err := manifest.Merge(base, ours, theirs, mergeOpts)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Convert to output format
	conflicts := make([]output.ManifestMergeConflict, len(result.Conflicts))
	for i, c := range result.Conflicts {
		conflicts[i] = output.ManifestMergeConflict{
			Path:        c.Path,
			BaseValue:   c.BaseValue,
			OursValue:   c.OursValue,
			TheirsValue: c.TheirsValue,
		}
	}

	var mergeChanges *output.ManifestMergeChanges
	if result.Changes != nil {
		mergeChanges = &output.ManifestMergeChanges{
			FromOurs:   result.Changes.FromOurs,
			FromTheirs: result.Changes.FromTheirs,
		}
	}

	out := output.ManifestMergeOutput{
		Base:          result.Base,
		Ours:          result.Ours,
		Theirs:        result.Theirs,
		Strategy:      result.Strategy,
		HasConflicts:  result.HasConflicts,
		Conflicts:     conflicts,
		Merged:        result.Merged,
		MergedMarkers: result.MergedMarkers,
		Changes:       mergeChanges,
		OutputPath:    opts.outputPath,
	}

	// Write output file if not dry-run and output path specified
	if !opts.dryRun && opts.outputPath != "" && opts.outputPath != "-" {
		// Determine format from output path
		format := manifest.FormatJSON
		if manifest.Format(opts.outputPath) == manifest.FormatYAML {
			format = manifest.FormatYAML
		}

		mergedManifest := result.ToManifest(opts.outputPath, format)
		if err := mergedManifest.Save(opts.outputPath); err != nil {
			return common.PrintAndReturn(printer, err)
		}
	}

	if err := printer.Print(out); err != nil {
		return err
	}

	// Return error exit code if conflicts exist
	if result.HasConflicts {
		conflictPaths := make([]string, len(result.Conflicts))
		for i, c := range result.Conflicts {
			conflictPaths[i] = c.Path
		}
		return errors.ErrMergeConflict(conflictPaths, opts.outputPath)
	}

	return nil
}

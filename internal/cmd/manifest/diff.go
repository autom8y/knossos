package manifest

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/manifest"
	"github.com/autom8y/knossos/internal/output"
)

type diffOptions struct {
	format      string
	ignoreOrder bool
}

func newDiffCmd(ctx *cmdContext) *cobra.Command {
	var opts diffOptions

	cmd := &cobra.Command{
		Use:   "diff <path1> <path2>",
		Short: "Compare two manifests",
		Long: `Compares two manifest files and shows differences.

Paths can be:
- Local file paths: .claude/manifest.json
- Git refs: HEAD:.claude/manifest.json, origin/main:.claude/manifest.json`,
		Args:         common.ExactArgs(2),
		SilenceUsage: true, // Don't print usage on errors
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(ctx, args[0], args[1], opts)
		},
	}

	cmd.Flags().StringVar(&opts.format, "format", "unified", "Diff format: unified, json, side-by-side")
	cmd.Flags().BoolVar(&opts.ignoreOrder, "ignore-order", false, "Ignore array ordering differences")

	return cmd
}

func runDiff(ctx *cmdContext, path1, path2 string, opts diffOptions) error {
	printer := ctx.getPrinter()

	// Load both manifests
	base, err := manifest.Load(path1)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	compare, err := manifest.Load(path2)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Compute diff
	diffOpts := manifest.ManifestDiffOptions{
		IgnoreOrder: opts.ignoreOrder,
	}
	result, err := manifest.Diff(base, compare, diffOpts)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Convert to output format
	changes := make([]output.ManifestDiffChange, len(result.Changes))
	for i, c := range result.Changes {
		changes[i] = output.ManifestDiffChange{
			Path:     c.Path,
			Type:     string(c.Type),
			OldValue: c.OldValue,
			NewValue: c.NewValue,
		}
	}

	out := output.ManifestDiffOutput{
		Base:          result.Base,
		Compare:       result.Compare,
		HasChanges:    result.HasChanges,
		Changes:       changes,
		Additions:     result.Additions,
		Modifications: result.Modifications,
		Deletions:     result.Deletions,
	}

	// Generate unified diff for text output
	if opts.format == "unified" {
		out.UnifiedDiff = result.FormatUnified()
	}

	if err := printer.Print(out); err != nil {
		return err
	}

	// Return exit code 1 if changes detected (useful for scripting)
	if result.HasChanges {
		return errors.New(errors.CodeGeneralError, "Changes detected")
	}

	return nil
}

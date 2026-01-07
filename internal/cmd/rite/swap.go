package rite

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	ritelib "github.com/autom8y/knossos/internal/rite"
)

type swapOptions struct {
	removeAll  bool
	keepAll    bool
	promoteAll bool
	dryRun     bool
	noSync     bool
}

func newSwapCmd(ctx *cmdContext) *cobra.Command {
	var opts swapOptions

	cmd := &cobra.Command{
		Use:   "swap <name>",
		Short: "Full context switch to another rite",
		Long: `Performs a full context switch (replacement, not additive).

This is equivalent to 'ari team switch' and maintains backward compatibility.
Any active invocations will be released before the swap.

Examples:
  ari rite swap security-rite
  ari rite swap 10x-dev-rite --remove-all`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSwap(ctx, args[0], opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.removeAll, "remove-all", "r", false, "Remove all orphaned agents from disk")
	cmd.Flags().BoolVarP(&opts.keepAll, "keep-all", "k", false, "Keep all orphaned agents in .claude/agents/")
	cmd.Flags().BoolVarP(&opts.promoteAll, "promote-all", "P", false, "Promote orphans to project-level agents")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&opts.noSync, "no-sync", false, "Skip CLAUDE.md inscription sync")

	return cmd
}

func runSwap(ctx *cmdContext, riteName string, opts swapOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	invoker := ctx.getInvoker()

	// Validate mutually exclusive flags
	flagCount := 0
	if opts.removeAll {
		flagCount++
	}
	if opts.keepAll {
		flagCount++
	}
	if opts.promoteAll {
		flagCount++
	}
	if flagCount > 1 {
		err := errors.New(errors.CodeUsageError,
			"Only one of --remove-all, --keep-all, or --promote-all may be specified")
		printer.PrintError(err)
		return err
	}

	// Release any active invocations before swap
	invocationsReleased := 0
	state, err := invoker.GetCurrentState()
	if err == nil && state.HasInvocations() {
		if !opts.dryRun {
			releaseResult, err := invoker.Release(ritelib.ReleaseOptions{All: true})
			if err == nil {
				invocationsReleased = releaseResult.InvocationCount
			}
		} else {
			invocationsReleased = state.InvocationCount()
		}
	}

	// Use rite switcher for the actual swap
	switcher := ritelib.NewSwitcher(resolver)

	switchOpts := ritelib.RiteSwitchOptions{
		TargetRite: riteName,
		RemoveAll:  opts.removeAll,
		KeepAll:    opts.keepAll,
		PromoteAll: opts.promoteAll,
		DryRun:     opts.dryRun,
		NoSync:     opts.noSync,
	}

	result, err := switcher.Switch(switchOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	out := output.RiteSwapOutput{
		Rite:                result.Rite,
		PreviousRite:        result.PreviousRite,
		SwitchedAt:          result.SwitchedAt.Format(time.RFC3339),
		AgentsInstalled:     result.AgentsInstalled,
		ClaudeMDUpdated:     result.ClaudeMDUpdated,
		ManifestPath:        result.ManifestPath,
		InscriptionSynced:   result.InscriptionSynced,
		InscriptionVersion:  result.InscriptionVersion,
		InvocationsReleased: invocationsReleased,
	}

	if result.OrphansHandled != nil {
		out.OrphansHandled = &output.OrphanHandleResult{
			Strategy: result.OrphansHandled.Strategy,
			Agents:   result.OrphansHandled.Agents,
		}
	}

	if len(result.SyncConflicts) > 0 {
		out.SyncConflicts = make([]output.InscriptionConflictOut, len(result.SyncConflicts))
		for i, c := range result.SyncConflicts {
			out.SyncConflicts[i] = output.InscriptionConflictOut{
				Region:    c.Region,
				Type:      c.Type,
				Message:   c.Message,
				Preserved: c.Preserved,
			}
		}
	}

	if opts.dryRun {
		return printer.Print(out)
	}

	return printer.PrintSuccess(out)
}

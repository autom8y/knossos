package team

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/team"
)

type switchOptions struct {
	removeAll  bool
	keepAll    bool
	promoteAll bool
	update     bool
	dryRun     bool
}

func newSwitchCmd(ctx *cmdContext) *cobra.Command {
	var opts switchOptions

	cmd := &cobra.Command{
		Use:   "switch <team-name>",
		Short: "Switch to a different team",
		Long:  `Switches the active team pack with atomic operations and orphan handling.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSwitch(ctx, args[0], opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.removeAll, "remove-all", "r", false, "Remove all orphaned agents from disk")
	cmd.Flags().BoolVarP(&opts.keepAll, "keep-all", "k", false, "Keep all orphaned agents in .claude/agents/")
	cmd.Flags().BoolVarP(&opts.promoteAll, "promote-all", "P", false, "Promote orphans to project-level agents")
	cmd.Flags().BoolVarP(&opts.update, "update", "u", false, "Re-pull agents even if already on target team")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without applying")

	return cmd
}

func runSwitch(ctx *cmdContext, teamName string, opts switchOptions) error {
	printer := ctx.getPrinter()
	switcher := ctx.getSwitcher()
	discovery := ctx.getDiscovery()

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

	switchOpts := team.SwitchOptions{
		TargetTeam: teamName,
		RemoveAll:  opts.removeAll,
		KeepAll:    opts.keepAll,
		PromoteAll: opts.promoteAll,
		Update:     opts.update,
		DryRun:     opts.dryRun,
	}

	// Handle dry-run specially
	if opts.dryRun {
		return runDryRun(ctx, teamName, switchOpts, printer, discovery)
	}

	// Execute switch
	result, err := switcher.Switch(switchOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	out := output.TeamSwitchOutput{
		Team:            result.Team,
		PreviousTeam:    result.PreviousTeam,
		SwitchedAt:      result.SwitchedAt.Format(time.RFC3339),
		AgentsInstalled: result.AgentsInstalled,
		ClaudeMDUpdated: result.ClaudeMDUpdated,
		ManifestPath:    result.ManifestPath,
	}

	if result.OrphansHandled != nil {
		out.OrphansHandled = &output.OrphanHandleResult{
			Strategy: result.OrphansHandled.Strategy,
			Agents:   result.OrphansHandled.Agents,
		}
	}

	return printer.PrintSuccess(out)
}

func runDryRun(ctx *cmdContext, teamName string, opts team.SwitchOptions, printer *output.Printer, discovery *team.Discovery) error {
	resolver := ctx.getResolver()

	// Get target team
	targetTeam, err := discovery.Get(teamName)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Load manifest to check orphans
	manifest, err := team.LoadManifest(resolver.AgentManifestFile())
	if err != nil {
		wrappedErr := errors.Wrap(errors.CodeGeneralError, "failed to load manifest", err)
		printer.PrintError(wrappedErr)
		return wrappedErr
	}

	// Detect orphans
	orphans := manifest.DetectOrphans(teamName)

	// Build agent list
	agents := make([]string, len(targetTeam.Agents))
	for i, a := range targetTeam.Agents {
		agents[i] = a + ".md"
	}

	out := output.TeamSwitchDryRunOutput{
		DryRun:                 true,
		WouldSwitchTo:          teamName,
		CurrentTeam:            manifest.ActiveTeam,
		WouldInstall:           agents,
		OrphansDetected:        orphans,
		OrphanStrategyRequired: len(orphans) > 0 && !opts.HasOrphanStrategy(),
	}

	if out.OrphanStrategyRequired {
		out.SuggestedFlags = []string{"--remove-all", "--keep-all", "--promote-all"}
	}

	return printer.Print(out)
}

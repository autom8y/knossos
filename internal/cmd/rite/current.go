package rite

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	ritelib "github.com/autom8y/knossos/internal/rite"
)

type currentOptions struct {
	borrowed bool
	native   bool
}

func newCurrentCmd(ctx *cmdContext) *cobra.Command {
	var opts currentOptions

	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show active rite and borrowed components",
		Long: `Displays the active rite, native components, and any borrowed components from invocations.

Includes token budget usage showing native, borrowed, and total tokens.

Examples:
  ari rite current
  ari rite current --borrowed
  ari rite current --native
  ari rite current -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCurrent(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.borrowed, "borrowed", false, "Show only borrowed components")
	cmd.Flags().BoolVar(&opts.native, "native", false, "Show only native components")

	return cmd
}

func runCurrent(ctx *cmdContext, opts currentOptions) error {
	printer := ctx.getPrinter()
	discovery := ctx.getDiscovery()
	invoker := ctx.getInvoker()

	// Get active rite
	activeRite := discovery.ActiveRiteName()

	out := output.RiteCurrentOutput{
		ActiveRite: activeRite,
	}

	// Get native components if active rite exists
	if activeRite != "" && !opts.borrowed {
		manifest, err := discovery.GetManifest(activeRite)
		if err == nil {
			out.NativeAgents = manifest.AgentNames()
			out.NativeSkills = manifest.SkillRefs()
		}
	}

	// Get invocation state
	state, err := invoker.GetCurrentState()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Build invocation output
	if !opts.native && len(state.Invocations) > 0 {
		out.Invocations = make([]output.InvocationOutput, len(state.Invocations))
		for i, inv := range state.Invocations {
			out.Invocations[i] = output.InvocationOutput{
				ID:        inv.ID,
				RiteName:  inv.RiteName,
				Component: inv.Component,
				Skills:    inv.Skills,
				InvokedAt: inv.InvokedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			// Convert agents to names
			for _, agent := range inv.Agents {
				out.Invocations[i].Agents = append(out.Invocations[i].Agents, agent.Name)
			}
		}

		// Collect all borrowed components
		out.BorrowedSkills = state.GetBorrowedSkills()
		for _, agent := range state.GetBorrowedAgents() {
			out.BorrowedAgents = append(out.BorrowedAgents, agent.Name)
		}
	}

	// Budget info
	out.Budget = output.CurrentBudgetOutput{
		NativeTokens:   state.Budget.NativeTokens,
		BorrowedTokens: state.Budget.BorrowedTokens,
		TotalTokens:    state.Budget.TotalTokens,
		BudgetLimit:    state.Budget.BudgetLimit,
		UsagePercent:   state.BudgetUsagePercent(),
	}

	// If we have an active rite and no native tokens set, estimate from manifest
	if activeRite != "" && state.Budget.NativeTokens == 0 {
		manifest, err := discovery.GetManifest(activeRite)
		if err == nil {
			budget := ritelib.NewBudgetCalculator()
			out.Budget.NativeTokens = budget.CalculateRiteCost(manifest)
			out.Budget.TotalTokens = out.Budget.NativeTokens + out.Budget.BorrowedTokens
			out.Budget.UsagePercent = float64(out.Budget.TotalTokens) / float64(out.Budget.BudgetLimit) * 100
		}
	}

	return printer.Print(out)
}

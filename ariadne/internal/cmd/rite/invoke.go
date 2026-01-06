package rite

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/output"
	ritelib "github.com/autom8y/ariadne/internal/rite"
)

type invokeOptions struct {
	dryRun        bool
	noInscription bool
}

func newInvokeCmd(ctx *cmdContext) *cobra.Command {
	var opts invokeOptions

	cmd := &cobra.Command{
		Use:   "invoke <name> [component]",
		Short: "Borrow components from another rite",
		Long: `Additively borrows components from another rite without switching context.

Component can be "skills", "agents", or omitted for all components.

Examples:
  ari rite invoke documentation                  # Borrow entire rite
  ari rite invoke documentation skills           # Borrow skills only
  ari rite invoke security-rite agents           # Borrow agents only
  ari rite invoke code-review agents --dry-run   # Preview changes`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			component := ""
			if len(args) > 1 {
				component = args[1]
			}
			return runInvoke(ctx, args[0], component, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview injection without applying")
	cmd.Flags().BoolVar(&opts.noInscription, "no-inscription", false, "Skip CLAUDE.md updates")

	return cmd
}

func runInvoke(ctx *cmdContext, riteName, component string, opts invokeOptions) error {
	printer := ctx.getPrinter()
	invoker := ctx.getInvoker()

	invokeOpts := ritelib.InvokeOptions{
		TargetRite:    riteName,
		Component:     component,
		DryRun:        opts.dryRun,
		NoInscription: opts.noInscription,
	}

	result, err := invoker.Invoke(invokeOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Convert agent names for output
	agentNames := make([]string, len(result.BorrowedAgents))
	for i, agent := range result.BorrowedAgents {
		agentNames[i] = agent.Name
	}

	out := output.RiteInvokeOutput{
		InvokedRite:        result.InvokedRite,
		Component:          result.Component,
		InvocationID:       result.InvocationID,
		BorrowedSkills:     result.BorrowedSkills,
		BorrowedAgents:     agentNames,
		InscriptionUpdated: result.InscriptionUpdated,
		EstimatedTokens:    result.EstimatedTokens,
		DryRun:             result.DryRun,
	}

	if opts.dryRun {
		return printer.Print(out)
	}

	return printer.PrintSuccess(out)
}

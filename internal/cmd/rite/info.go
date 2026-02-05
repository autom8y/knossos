package rite

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	ritelib "github.com/autom8y/knossos/internal/rite"
)

type infoOptions struct {
	budget     bool
	components bool
}

func newInfoCmd(ctx *cmdContext) *cobra.Command {
	var opts infoOptions

	cmd := &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed rite information",
		Long: `Displays detailed information about a rite including agents, skills, workflow, and budget.

Examples:
  ari rite info ecosystem
  ari rite info 10x-dev --budget
  ari rite info forge --components
  ari rite info ecosystem -o json`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(ctx, args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.budget, "budget", false, "Show detailed budget breakdown")
	cmd.Flags().BoolVar(&opts.components, "components", false, "Show component list only")

	return cmd
}

func runInfo(ctx *cmdContext, riteName string, opts infoOptions) error {
	printer := ctx.getPrinter()
	discovery := ctx.getDiscovery()

	// Get rite summary
	riteSummary, err := discovery.Get(riteName)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Get full manifest
	manifest, err := discovery.GetManifest(riteName)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	out := output.RiteInfoOutput{
		Name:          manifest.Name,
		DisplayName:   manifest.DisplayName,
		Description:   manifest.Description,
		Form:          string(manifest.Form),
		Path:          riteSummary.Path,
		Source:        riteSummary.Source,
		Active:        riteSummary.Active,
		SchemaVersion: manifest.SchemaVersion,
	}

	// Add agents
	out.Agents = make([]output.RiteAgentInfo, len(manifest.Agents))
	for i, agent := range manifest.Agents {
		out.Agents[i] = output.RiteAgentInfo{
			Name:     agent.Name,
			File:     agent.File,
			Role:     agent.Role,
			Produces: agent.Produces,
		}
	}

	// Add skills
	out.Skills = make([]output.RiteSkillInfo, len(manifest.Skills))
	for i, skill := range manifest.Skills {
		out.Skills[i] = output.RiteSkillInfo{
			Ref:      skill.Ref,
			Path:     skill.Path,
			External: skill.External,
		}
	}

	// Add workflow if present
	if manifest.Workflow != nil {
		out.Workflow = &output.RiteWorkflowInfo{
			Type:       manifest.Workflow.Type,
			EntryPoint: manifest.Workflow.EntryPoint,
		}
		for _, phase := range manifest.Workflow.Phases {
			out.Workflow.Phases = append(out.Workflow.Phases, phase.Name)
		}
	}

	// Add budget info
	budget := ritelib.NewBudgetCalculator()
	summaryCost := budget.CalculateSummaryCost(manifest)
	out.Budget = &output.RiteBudgetInfo{
		EstimatedTokens: summaryCost.TotalCost,
		AgentsCost:      summaryCost.AgentsCost,
		SkillsCost:      summaryCost.SkillsCost,
		WorkflowCost:    summaryCost.WorkflowCost,
	}

	return printer.Print(out)
}

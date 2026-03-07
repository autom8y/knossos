package rite

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/rite"
)

type listOptions struct {
	form    string
	project bool
	user    bool
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available rites",
		Long: `Lists all available rites from project and user directories.

Shows name, form, agent count, skill count, and active status for each rite.

Examples:
  ari rite list
  ari rite list --form full
  ari rite list --project
  ari rite list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.form, "form", "", "Filter by form (simple, practitioner, procedural, full)")
	cmd.Flags().BoolVar(&opts.project, "project", false, "Show project rites only")
	cmd.Flags().BoolVar(&opts.user, "user", false, "Show user rites only")

	return cmd
}

func runList(ctx *cmdContext, opts listOptions) error {
	printer := ctx.getPrinter()
	discovery := ctx.getDiscovery()

	var rites []rite.Rite
	var err error

	// Get rites based on filters
	switch {
	case opts.form != "":
		rites, err = discovery.ListByForm(rite.RiteForm(opts.form))
	case opts.project:
		rites, err = discovery.ListBySource("project")
	case opts.user:
		rites, err = discovery.ListBySource("user")
	default:
		rites, err = discovery.List()
	}

	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Build output structure
	summaries := make([]output.RiteSummary, len(rites))
	for i, r := range rites {
		summaries[i] = output.RiteSummary{
			Name:        r.Name,
			DisplayName: r.DisplayName,
			Description: r.Description,
			Form:        string(r.Form),
			Agents:      r.Agents,
			AgentCount:  r.AgentCount,
			Skills:      r.Skills,
			SkillCount:  r.SkillCount,
			Path:        r.Path,
			Source:      r.Source,
			Active:      r.Active,
		}
	}

	result := output.RiteListOutput{
		Rites:      summaries,
		Total:      len(rites),
		ActiveRite: discovery.ActiveRiteName(),
	}

	return printer.Print(result)
}

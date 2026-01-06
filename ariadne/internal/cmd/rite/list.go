package rite

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/rite"
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
		Long:  `Lists all available rites from project and user directories.`,
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
	if opts.form != "" {
		rites, err = discovery.ListByForm(rite.RiteForm(opts.form))
	} else if opts.project {
		rites, err = discovery.ListBySource("project")
	} else if opts.user {
		rites, err = discovery.ListBySource("user")
	} else {
		rites, err = discovery.List()
	}

	if err != nil {
		printer.PrintError(err)
		return err
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

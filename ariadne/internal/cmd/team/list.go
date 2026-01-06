package team

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/output"
)

type listOptions struct {
	format string
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available rites",
		Long:  `Lists all available rites (practice bundles) from project and user directories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "table", "Output format: table, name-only, json, yaml")

	return cmd
}

func runList(ctx *cmdContext, opts listOptions) error {
	printer := ctx.getPrinter()
	discovery := ctx.getDiscovery()

	teams, err := discovery.List()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Handle name-only format specially
	if opts.format == "name-only" {
		for _, t := range teams {
			printer.PrintLine(t.Name)
		}
		return nil
	}

	// Build output structure
	summaries := make([]output.TeamSummary, len(teams))
	for i, t := range teams {
		summaries[i] = output.TeamSummary{
			Name:        t.Name,
			Description: t.Description,
			Agents:      t.Agents,
			AgentCount:  t.AgentCount,
			Path:        t.Path,
			Active:      t.Active,
		}
	}

	result := output.TeamListOutput{
		Teams:      summaries,
		Total:      len(teams),
		ActiveTeam: discovery.ActiveRiteName(),
	}

	return printer.Print(result)
}

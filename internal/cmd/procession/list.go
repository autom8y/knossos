// Package procession implements the ari procession commands for cross-rite workflow management.
package procession

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	procmena "github.com/autom8y/knossos/internal/materialize/procession"
	"github.com/autom8y/knossos/internal/output"
)

// newListCmd creates the procession list subcommand.
func newListCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available procession templates",
		Long: `List all available procession templates resolved through the 5-tier
resolution chain (project > user > org > platform > embedded).

Higher-priority tiers shadow lower-priority ones by template name.

Examples:
  ari procession list
  ari procession list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx)
		},
	}
	return cmd
}

// templateSummary is a brief summary of a procession template for listing.
type templateSummary struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Source       string   `json:"source"`
	StationCount int      `json:"station_count"`
	Stations     []string `json:"stations"`
	EntryRite    string   `json:"entry_rite"`
}

// listOutput represents the output of procession list.
type listOutput struct {
	Templates []templateSummary `json:"templates"`
	Total     int               `json:"total"`
}

// Text implements output.Textable for listOutput.
func (o listOutput) Text() string {
	if len(o.Templates) == 0 {
		return "No procession templates found"
	}

	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATIONS\tENTRY RITE\tSOURCE")
	for _, t := range o.Templates {
		stationNames := strings.Join(t.Stations, " → ")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", t.Name, stationNames, t.EntryRite, t.Source)
	}
	_ = tw.Flush()
	fmt.Fprintf(&b, "\nTotal: %d template(s)\n", o.Total)
	return b.String()
}

var _ output.Textable = listOutput{}

// runList executes the procession list command.
func runList(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	projectRoot := resolver.ProjectRoot()
	embeddedFS := common.EmbeddedProcessions()

	resolved, err := procmena.ResolveProcessions(projectRoot, embeddedFS)
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to resolve procession templates", err))
	}

	// Sort by name for stable output
	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].Name < resolved[j].Name
	})

	templates := make([]templateSummary, len(resolved))
	for i, rp := range resolved {
		stations := make([]string, len(rp.Template.Stations))
		for j, s := range rp.Template.Stations {
			stations[j] = s.Name
		}

		entryRite := ""
		if len(rp.Template.Stations) > 0 {
			entryRite = rp.Template.Stations[0].Rite
		}

		templates[i] = templateSummary{
			Name:         rp.Name,
			Description:  rp.Template.Description,
			Source:       rp.Source,
			StationCount: len(rp.Template.Stations),
			Stations:     stations,
			EntryRite:    entryRite,
		}
	}

	return printer.Print(listOutput{
		Templates: templates,
		Total:     len(templates),
	})
}

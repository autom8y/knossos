package artifact

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
)

// listOutput is the structured output for ari artifact list.
type listOutput struct {
	Dimension string         `json:"dimension"`
	Counts    map[string]int `json:"counts"`
	Total     int            `json:"total"`
}

// Text implements output.Textable for human-readable table output.
func (l listOutput) Text() string {
	if len(l.Counts) == 0 {
		return "No artifacts found.\n"
	}

	var b strings.Builder

	// Header
	header := strings.ToUpper(l.Dimension)
	b.WriteString(fmt.Sprintf("%-30s %s\n", header, "COUNT"))
	b.WriteString(strings.Repeat("-", 40) + "\n")

	// Rows
	for key, count := range l.Counts {
		b.WriteString(fmt.Sprintf("%-30s %d\n", key, count))
	}

	b.WriteString(fmt.Sprintf("\nTotal: %d artifacts\n", l.Total))

	return b.String()
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	var by string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List artifact counts by dimension",
		Long: `List artifact counts grouped by a specific dimension.

Dimensions:
  phase       - Group by workflow phase
  type        - Group by artifact type
  specialist  - Group by producing agent
  session     - Group by session`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			querier := ctx.getQuerier()

			var counts map[string]int
			var err error
			var total int

			switch by {
			case "phase":
				phaseCounts, err := querier.ListPhases()
				if err != nil {
					printer.PrintError(err)
					return err
				}
				counts = make(map[string]int)
				for phase, count := range phaseCounts {
					counts[string(phase)] = count
					total += count
				}

			case "type":
				typeCounts, err := querier.ListTypes()
				if err != nil {
					printer.PrintError(err)
					return err
				}
				counts = make(map[string]int)
				for t, count := range typeCounts {
					counts[string(t)] = count
					total += count
				}

			case "specialist":
				counts, err = querier.ListSpecialists()
				if err != nil {
					printer.PrintError(err)
					return err
				}
				for _, count := range counts {
					total += count
				}

			case "session":
				counts, err = querier.ListSessions()
				if err != nil {
					printer.PrintError(err)
					return err
				}
				for _, count := range counts {
					total += count
				}

			default:
				e := errors.NewWithDetails(errors.CodeUsageError,
					"invalid dimension",
					map[string]any{"by": by, "valid": "phase, type, specialist, session"})
				printer.PrintError(e)
				return e
			}

			return printer.Print(listOutput{
				Dimension: by,
				Counts:    counts,
				Total:     total,
			})
		},
	}

	cmd.Flags().StringVar(&by, "by", "phase", "Dimension to group by (phase, type, specialist, session)")

	return cmd
}

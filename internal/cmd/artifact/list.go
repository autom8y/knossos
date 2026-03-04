package artifact

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
)

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

			// Format output
			format := output.ParseFormat(*ctx.Output)
			if format == output.FormatText {
				printCountsTable(by, counts, total)
			} else {
				out := map[string]any{
					"dimension": by,
					"counts":    counts,
					"total":     total,
				}
				return printer.Print(out)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&by, "by", "phase", "Dimension to group by (phase, type, specialist, session)")

	return cmd
}

func printCountsTable(dimension string, counts map[string]int, total int) {
	if len(counts) == 0 {
		fmt.Println("No artifacts found.")
		return
	}

	// Header
	header := strings.ToUpper(dimension)
	fmt.Printf("%-30s %s\n", header, "COUNT")
	fmt.Println(strings.Repeat("-", 40))

	// Rows
	for key, count := range counts {
		fmt.Printf("%-30s %d\n", key, count)
	}

	fmt.Printf("\nTotal: %d artifacts\n", total)
}

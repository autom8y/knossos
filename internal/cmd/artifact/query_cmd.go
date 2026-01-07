package artifact

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/artifact"
	"github.com/autom8y/knossos/internal/output"
)

func newQueryCmd(ctx *cmdContext) *cobra.Command {
	var (
		phaseFilter      string
		typeFilter       string
		specialistFilter string
		sessionFilter    string
		limit            int
		formatFlag       string
	)

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query the artifact registry",
		Long: `Query the project artifact registry with optional filters.

Multiple filters are ANDed together. Results can be output in JSON, YAML, or table format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			querier := ctx.getQuerier()

			// Build filter
			filter := artifact.QueryFilter{
				Phase:      artifact.Phase(phaseFilter),
				Type:       artifact.ArtifactType(typeFilter),
				Specialist: specialistFilter,
				SessionID:  sessionFilter,
			}

			// Execute query
			result, err := querier.Query(filter)
			if err != nil {
				printer.PrintError(err)
				return err
			}

			// Apply limit
			entries := result.Entries
			if limit > 0 && len(entries) > limit {
				entries = entries[:limit]
			}

			// Format output based on formatFlag
			format := output.ParseFormat(formatFlag)
			if format == output.FormatText {
				// Table format for text
				printTable(entries, result.Count)
			} else {
				// JSON or YAML
				output := map[string]interface{}{
					"entries": entries,
					"count":   len(entries),
					"total":   result.Count,
					"filter":  filter,
				}
				printer.Print(output)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&phaseFilter, "phase", "", "Filter by phase (requirements, design, implementation, validation)")
	cmd.Flags().StringVar(&typeFilter, "type", "", "Filter by type (prd, tdd, adr, test-plan, code, runbook)")
	cmd.Flags().StringVar(&specialistFilter, "specialist", "", "Filter by specialist agent")
	cmd.Flags().StringVar(&sessionFilter, "session", "", "Filter by session ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum entries to return")
	cmd.Flags().StringVar(&formatFlag, "format", "text", "Output format: json, yaml, table (default: text)")

	return cmd
}

func printTable(entries []artifact.Entry, total int) {
	if len(entries) == 0 {
		fmt.Println("No artifacts found.")
		return
	}

	// Header
	fmt.Printf("%-25s %-12s %-16s %-20s %s\n", "ARTIFACT ID", "TYPE", "PHASE", "SPECIALIST", "SESSION")
	fmt.Println(strings.Repeat("-", 120))

	// Rows
	for _, entry := range entries {
		sessionShort := entry.SessionID
		if len(sessionShort) > 30 {
			sessionShort = sessionShort[:27] + "..."
		}
		fmt.Printf("%-25s %-12s %-16s %-20s %s\n",
			entry.ArtifactID,
			entry.ArtifactType,
			entry.Phase,
			entry.Specialist,
			sessionShort,
		)
	}

	fmt.Printf("\nTotal: %d artifact(s)", total)
	if len(entries) < total {
		fmt.Printf(" (showing %d)", len(entries))
	}
	fmt.Println()
}

package artifact

import (
	"fmt"
	"github.com/autom8y/knossos/internal/cmd/common"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/artifact"
)

// queryOutput is the structured output for ari artifact query.
type queryOutput struct {
	Entries []artifact.Entry     `json:"entries"`
	Count   int                  `json:"count"`
	Total   int                  `json:"total"`
	Filter  artifact.QueryFilter `json:"filter"`
}

// Text implements output.Textable for human-readable table output.
func (q queryOutput) Text() string {
	if len(q.Entries) == 0 {
		return "No artifacts found.\n"
	}

	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "%-25s %-12s %-16s %-20s %s\n", "ARTIFACT ID", "TYPE", "PHASE", "SPECIALIST", "SESSION")
	b.WriteString(strings.Repeat("-", 120) + "\n")

	// Rows
	for _, entry := range q.Entries {
		sessionShort := entry.SessionID
		if len(sessionShort) > 30 {
			sessionShort = sessionShort[:27] + "..."
		}
		fmt.Fprintf(&b, "%-25s %-12s %-16s %-20s %s\n",
			entry.ArtifactID,
			entry.ArtifactType,
			entry.Phase,
			entry.Specialist,
			sessionShort)
	}

	fmt.Fprintf(&b, "\nTotal: %d artifact(s)", q.Total)
	if len(q.Entries) < q.Total {
		fmt.Fprintf(&b, " (showing %d)", len(q.Entries))
	}
	b.WriteString("\n")

	return b.String()
}

func newQueryCmd(ctx *cmdContext) *cobra.Command {
	var (
		phaseFilter      string
		typeFilter       string
		specialistFilter string
		sessionFilter    string
		limit            int
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
				return common.PrintAndReturn(printer, err)
			}

			// Apply limit
			entries := result.Entries
			if limit > 0 && len(entries) > limit {
				entries = entries[:limit]
			}

			return printer.Print(queryOutput{
				Entries: entries,
				Count:   len(entries),
				Total:   result.Count,
				Filter:  filter,
			})
		},
	}

	cmd.Flags().StringVar(&phaseFilter, "phase", "", "Filter by phase (requirements, design, implementation, validation)")
	cmd.Flags().StringVar(&typeFilter, "type", "", "Filter by type (prd, tdd, adr, test-plan, code, runbook)")
	cmd.Flags().StringVar(&specialistFilter, "specialist", "", "Filter by specialist agent")
	cmd.Flags().StringVar(&sessionFilter, "session", "", "Filter by session ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum entries to return")

	return cmd
}

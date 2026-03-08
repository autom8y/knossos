// Package ask implements the ari ask command for natural language CLI queries.
package ask

import (
	"fmt"
	"strings"
)

// AskOutput represents the full ask command output.
// Implements output.Textable for human-readable output.
// Serializes to JSON/YAML via struct tags for machine-readable output.
type AskOutput struct {
	Query   string           `json:"query"`
	Results []AskResultEntry `json:"results"`
	Total   int              `json:"total"`
	Context string           `json:"context,omitempty"` // active rite info if available
}

// AskResultEntry represents a single search result.
type AskResultEntry struct {
	Rank    int    `json:"rank"`
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Summary string `json:"summary"`
	Action  string `json:"action"`
	Score   int    `json:"score,omitempty"`
}

// Text implements output.Textable.
func (a AskOutput) Text() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Query: %q\n", a.Query))

	if len(a.Results) == 0 {
		b.WriteString("\nNo matches found.\n")
		if a.Context != "" {
			b.WriteString(fmt.Sprintf("\n%s\n", a.Context))
		}
		b.WriteString("\nSuggestions:\n")
		b.WriteString("  - Try broader terms (e.g. \"release\" instead of \"publish package\")\n")
		b.WriteString("  - Use ari explain to browse concepts\n")
		b.WriteString("  - Use ari --help to see all commands\n")
		return b.String()
	}

	b.WriteString("\n")
	for _, r := range a.Results {
		b.WriteString(fmt.Sprintf("  %d. %s [%s]\n", r.Rank, r.Name, r.Domain))
		b.WriteString(fmt.Sprintf("     %s\n", r.Summary))
		if r.Action != "" {
			b.WriteString(fmt.Sprintf("     Try: %s\n", r.Action))
		}
		b.WriteString("\n")
	}

	if a.Context != "" {
		b.WriteString(fmt.Sprintf("%s\n", a.Context))
	}

	plural := "results"
	if a.Total == 1 {
		plural = "result"
	}
	b.WriteString(fmt.Sprintf("%d %s found\n", a.Total, plural))

	return b.String()
}

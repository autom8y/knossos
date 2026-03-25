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
	Query          string             `json:"query"`
	Results        []AskResultEntry   `json:"results"`
	Total          int                `json:"total"`
	Context        string             `json:"context,omitempty"`         // active rite info if available
	SessionContext *AskSessionContext `json:"session_context,omitempty"` // session state used for scoring
}

// AskSessionContext reports the session state used for scoring in AskOutput.
// All fields use omitempty to ensure backward compatibility.
type AskSessionContext struct {
	SessionID       string              `json:"session_id,omitempty"`
	Phase           string              `json:"phase,omitempty"`
	Rite            string              `json:"rite,omitempty"`
	Complexity      string              `json:"complexity,omitempty"`
	Initiative      string              `json:"initiative,omitempty"`
	ActivitySummary *AskActivitySummary `json:"activity_summary,omitempty"`
}

// AskActivitySummary reports aggregated event counts from the session's events.jsonl.
// Omitted from output when no events were read.
type AskActivitySummary struct {
	FileChanges  int    `json:"file_changes,omitempty"`
	AgentTasks   int    `json:"agent_tasks,omitempty"`
	LastEventAge string `json:"last_event_age,omitempty"` // e.g., "2m", "1h"
}

// AskResultEntry represents a single search result.
type AskResultEntry struct {
	Rank      int    `json:"rank"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	Summary   string `json:"summary"`
	Action    string `json:"action"`
	Score     int    `json:"score,omitempty"`
	Source    string `json:"source,omitempty"`    // repo origin for knowledge results
	Freshness string `json:"freshness,omitempty"` // freshness annotation for knowledge results
}

// Text implements output.Textable.
func (a AskOutput) Text() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Query: %q\n", a.Query)

	if len(a.Results) == 0 {
		b.WriteString("\nNo matches found.\n")
		if a.Context != "" {
			fmt.Fprintf(&b, "\n%s\n", a.Context)
		}
		b.WriteString("\nSuggestions:\n")
		b.WriteString("  - Try broader terms (e.g. \"release\" instead of \"publish package\")\n")
		b.WriteString("  - Use ari explain to browse concepts\n")
		b.WriteString("  - Use ari --help to see all commands\n")
		return b.String()
	}

	b.WriteString("\n")
	for _, r := range a.Results {
		domainLabel := r.Domain
		if r.Source != "" {
			domainLabel = r.Domain + " from " + r.Source
		}
		fmt.Fprintf(&b, "  %d. %s [%s]\n", r.Rank, r.Name, domainLabel)
		fmt.Fprintf(&b, "     %s\n", r.Summary)
		if r.Freshness != "" {
			fmt.Fprintf(&b, "     Freshness: %s\n", r.Freshness)
		}
		if r.Action != "" {
			fmt.Fprintf(&b, "     Try: %s\n", r.Action)
		}
		b.WriteString("\n")
	}

	if a.Context != "" {
		fmt.Fprintf(&b, "%s\n", a.Context)
	}

	plural := "results"
	if a.Total == 1 {
		plural = "result"
	}
	fmt.Fprintf(&b, "%d %s found\n", a.Total, plural)

	return b.String()
}

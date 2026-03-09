package naxos

import (
	"fmt"
	"strings"
)

// TriageOutput is the output structure for the triage command.
// It implements output.Textable and output.Tabular interfaces.
type TriageOutput struct {
	Entries      []TriageEntry    `json:"entries"`
	TotalScanned int              `json:"total_scanned"`
	TotalTriaged int              `json:"total_triaged"`
	BySeverity   map[Severity]int `json:"by_severity"`
	SummaryLine  string           `json:"summary_line"`
	TriagedAt    string           `json:"triaged_at"`
}

// FromTriageResult creates a TriageOutput from a TriageResult.
func FromTriageResult(result *TriageResult) TriageOutput {
	return TriageOutput{
		Entries:      result.Entries,
		TotalScanned: result.TotalScanned,
		TotalTriaged: result.TotalTriaged,
		BySeverity:   result.BySeverity,
		SummaryLine:  result.SummaryLine,
		TriagedAt:    result.TriagedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// Headers implements output.Tabular.
func (o TriageOutput) Headers() []string {
	return []string{"#", "SESSION ID", "SEVERITY", "REASON", "INACTIVE", "ACTIONABLE", "SUGGESTED ACTION"}
}

// Rows implements output.Tabular.
func (o TriageOutput) Rows() [][]string {
	rows := make([][]string, len(o.Entries))
	for i, e := range o.Entries {
		// Truncate session ID if too long for display.
		sessionID := e.SessionID
		if len(sessionID) > 35 {
			sessionID = sessionID[:32] + "..."
		}

		actionable := "no"
		if e.Actionable {
			actionable = "yes"
		}

		rows[i] = []string{
			fmt.Sprintf("%d", i+1),
			sessionID,
			severitySymbol(e.Severity) + " " + e.Severity.String(),
			e.Reason.String(),
			FormatDuration(e.InactiveFor),
			actionable,
			e.SuggestedAction.String(),
		}
	}
	return rows
}

// Text implements output.Textable.
func (o TriageOutput) Text() string {
	var b strings.Builder

	// Header
	b.WriteString("Naxos Triage Report\n")
	b.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Summary
	fmt.Fprintf(&b, "Scanned:  %d sessions\n", o.TotalScanned)
	fmt.Fprintf(&b, "Triaged:  %d orphaned\n", o.TotalTriaged)

	if o.TotalTriaged > 0 {
		b.WriteString("\nBy Severity:\n")
		for _, sev := range []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow} {
			if count := o.BySeverity[sev]; count > 0 {
				fmt.Fprintf(&b, "  %s %s: %d\n", severitySymbol(sev), sev, count)
			}
		}
	}

	b.WriteString("\n")

	if o.TotalTriaged == 0 {
		b.WriteString("No orphaned sessions found. All sessions are healthy.\n")
		return b.String()
	}

	// Detailed entry list
	b.WriteString("Orphaned Sessions:\n")
	b.WriteString(strings.Repeat("-", 50) + "\n")

	for _, e := range o.Entries {
		fmt.Fprintf(&b, "\n%s [%s] %s\n", severitySymbol(e.Severity), e.Severity, e.SessionID)
		fmt.Fprintf(&b, "  Status:   %s\n", e.Status)
		if e.Initiative != "" {
			fmt.Fprintf(&b, "  Goal:     %s\n", truncate(e.Initiative, 50))
		}
		fmt.Fprintf(&b, "  Reason:   %s\n", e.Reason.Description())
		fmt.Fprintf(&b, "  Inactive: %s\n", FormatDuration(e.InactiveFor))
		if e.Actionable {
			fmt.Fprintf(&b, "  Action:   %s\n", e.SuggestedAction.Description())
		}
	}

	// Footer with action hints
	b.WriteString("\n" + strings.Repeat("-", 50) + "\n")
	b.WriteString("Run: ari session wrap --session <id>  | ari session resume <id>\n")

	return b.String()
}

// severitySymbol returns a visual indicator for the severity level.
func severitySymbol(s Severity) string {
	switch s {
	case SeverityCritical:
		return "[!!]"
	case SeverityHigh:
		return "[! ]"
	case SeverityMedium:
		return "[ ~]"
	case SeverityLow:
		return "[  ]"
	default:
		return "[? ]"
	}
}

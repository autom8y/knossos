package naxos

import (
	"fmt"
	"strings"
)

// ScanOutput is the output structure for the scan command.
// It implements output.Tabular and output.Textable interfaces.
type ScanOutput struct {
	OrphanedSessions []OrphanedSession `json:"orphaned_sessions"`
	TotalScanned     int               `json:"total_scanned"`
	TotalOrphaned    int               `json:"total_orphaned"`
	ScannedAt        string            `json:"scanned_at"`
	ByReason         ByReasonSummary   `json:"by_reason"`
	Config           ConfigSummary     `json:"config"`
}

// ByReasonSummary breaks down counts by reason.
type ByReasonSummary struct {
	Inactive       int `json:"inactive"`
	StaleSails     int `json:"stale_sails"`
	IncompleteWrap int `json:"incomplete_wrap"`
}

// ConfigSummary shows the scan configuration.
type ConfigSummary struct {
	InactiveThreshold   string `json:"inactive_threshold"`
	StaleSailsThreshold string `json:"stale_sails_threshold"`
	IncludeArchived     bool   `json:"include_archived"`
}

// FromScanResult creates a ScanOutput from a ScanResult.
func FromScanResult(result *ScanResult) ScanOutput {
	return ScanOutput{
		OrphanedSessions: result.OrphanedSessions,
		TotalScanned:     result.TotalScanned,
		TotalOrphaned:    result.TotalOrphaned,
		ScannedAt:        result.ScannedAt.Format("2006-01-02T15:04:05Z"),
		ByReason: ByReasonSummary{
			Inactive:       result.ByReason[ReasonInactive],
			StaleSails:     result.ByReason[ReasonStaleSails],
			IncompleteWrap: result.ByReason[ReasonIncompleteWrap],
		},
		Config: ConfigSummary{
			InactiveThreshold:   FormatDuration(result.Config.InactiveThreshold),
			StaleSailsThreshold: FormatDuration(result.Config.StaleSailsThreshold),
			IncludeArchived:     result.Config.IncludeArchived,
		},
	}
}

// Headers implements output.Tabular.
func (o ScanOutput) Headers() []string {
	return []string{"SESSION ID", "STATUS", "REASON", "INACTIVE", "SUGGESTED ACTION"}
}

// Rows implements output.Tabular.
func (o ScanOutput) Rows() [][]string {
	rows := make([][]string, len(o.OrphanedSessions))
	for i, s := range o.OrphanedSessions {
		// Truncate session ID if too long for display
		sessionID := s.SessionID
		if len(sessionID) > 35 {
			sessionID = sessionID[:32] + "..."
		}

		rows[i] = []string{
			sessionID,
			s.Status,
			reasonSymbol(s.Reason) + " " + s.Reason.String(),
			FormatDuration(s.InactiveFor),
			s.SuggestedAction.String(),
		}
	}
	return rows
}

// Text implements output.Textable.
func (o ScanOutput) Text() string {
	var b strings.Builder

	// Header
	b.WriteString("Naxos Session Scan Report\n")
	b.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Summary
	b.WriteString(fmt.Sprintf("Scanned: %d sessions\n", o.TotalScanned))
	b.WriteString(fmt.Sprintf("Orphaned: %d sessions\n", o.TotalOrphaned))
	b.WriteString("\n")

	if o.TotalOrphaned == 0 {
		b.WriteString("No orphaned sessions found. All sessions are healthy.\n")
		return b.String()
	}

	// Breakdown by reason
	b.WriteString("By Reason:\n")
	if o.ByReason.Inactive > 0 {
		b.WriteString(fmt.Sprintf("  %s Inactive (>%s): %d\n",
			reasonSymbol(ReasonInactive), o.Config.InactiveThreshold, o.ByReason.Inactive))
	}
	if o.ByReason.StaleSails > 0 {
		b.WriteString(fmt.Sprintf("  %s Stale Sails (>%s): %d\n",
			reasonSymbol(ReasonStaleSails), o.Config.StaleSailsThreshold, o.ByReason.StaleSails))
	}
	if o.ByReason.IncompleteWrap > 0 {
		b.WriteString(fmt.Sprintf("  %s Incomplete Wrap: %d\n",
			reasonSymbol(ReasonIncompleteWrap), o.ByReason.IncompleteWrap))
	}
	b.WriteString("\n")

	// Detailed list
	b.WriteString("Orphaned Sessions:\n")
	b.WriteString(strings.Repeat("-", 50) + "\n")

	for _, s := range o.OrphanedSessions {
		b.WriteString(fmt.Sprintf("\n%s %s\n", reasonSymbol(s.Reason), s.SessionID))
		b.WriteString(fmt.Sprintf("  Status: %s\n", s.Status))
		b.WriteString(fmt.Sprintf("  Initiative: %s\n", truncate(s.Initiative, 40)))
		b.WriteString(fmt.Sprintf("  Reason: %s\n", s.Reason.Description()))
		b.WriteString(fmt.Sprintf("  Inactive: %s\n", FormatDuration(s.InactiveFor)))
		if s.AdditionalInfo != "" {
			b.WriteString(fmt.Sprintf("  Info: %s\n", s.AdditionalInfo))
		}
		b.WriteString(fmt.Sprintf("  Suggested: %s\n", s.SuggestedAction.Description()))
	}

	// Footer with actions hint
	b.WriteString("\n" + strings.Repeat("-", 50) + "\n")
	b.WriteString("Actions:\n")
	b.WriteString("  To wrap:   ari session wrap --session <id>\n")
	b.WriteString("  To resume: ari session resume <id>\n")
	b.WriteString("  To delete: rm -rf .sos/sessions/<id>\n")

	return b.String()
}

// reasonSymbol returns a visual indicator for the reason.
func reasonSymbol(r OrphanReason) string {
	switch r {
	case ReasonInactive:
		return "[!]"
	case ReasonStaleSails:
		return "[~]"
	case ReasonIncompleteWrap:
		return "[x]"
	default:
		return "[?]"
	}
}

// truncate shortens a string to the given length with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

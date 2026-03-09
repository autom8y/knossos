package naxos

import (
	"fmt"
	"sort"
	"time"
)

// Severity classifies how urgently an orphaned session needs attention.
type Severity string

const (
	// SeverityCritical - immediate attention required.
	SeverityCritical Severity = "CRITICAL"
	// SeverityHigh - attention needed soon.
	SeverityHigh Severity = "HIGH"
	// SeverityMedium - attention needed within reasonable time.
	SeverityMedium Severity = "MEDIUM"
	// SeverityLow - low priority cleanup candidate.
	SeverityLow Severity = "LOW"
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	return string(s)
}

// TriageEntry is an OrphanedSession augmented with triage classification.
type TriageEntry struct {
	OrphanedSession
	// Severity is the urgency level for this session.
	Severity Severity `json:"severity"`
	// Priority is the numeric sort order (lower = more urgent).
	Priority int `json:"priority"`
	// Actionable indicates whether automated action is safe to recommend.
	Actionable bool `json:"actionable"`
	// TriagedAt is when this entry was triaged.
	TriagedAt time.Time `json:"triaged_at"`
}

// TriageResult holds the complete output of a triage pass over a ScanResult.
type TriageResult struct {
	// Entries is the list of triaged orphaned sessions, sorted by priority.
	Entries []TriageEntry `json:"entries"`
	// TotalScanned is how many sessions were examined in the underlying scan.
	TotalScanned int `json:"total_scanned"`
	// TotalTriaged is how many sessions are in Entries.
	TotalTriaged int `json:"total_triaged"`
	// BySeverity breaks down the count of entries by severity level.
	BySeverity map[Severity]int `json:"by_severity"`
	// SummaryLine is a human-readable one-line summary.
	SummaryLine string `json:"summary_line"`
	// TriagedAt is when the triage was performed.
	TriagedAt time.Time `json:"triaged_at"`
}

// Triage classifies each OrphanedSession in the scan result and returns a
// TriageResult sorted by priority (most urgent first).
func Triage(result *ScanResult) *TriageResult {
	now := time.Now().UTC()

	bySeverity := make(map[Severity]int)
	entries := make([]TriageEntry, 0, len(result.OrphanedSessions))

	for _, session := range result.OrphanedSessions {
		sev := computeSeverity(session)
		pri := computePriority(sev, session)
		entries = append(entries, TriageEntry{
			OrphanedSession: session,
			Severity:        sev,
			Priority:        pri,
			Actionable:      isActionable(sev),
			TriagedAt:       now,
		})
		bySeverity[sev]++
	}

	// Sort by priority ascending (lower = more urgent), stable sort for determinism.
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Priority != entries[j].Priority {
			return entries[i].Priority < entries[j].Priority
		}
		// Secondary: longer inactive first
		return entries[i].InactiveFor > entries[j].InactiveFor
	})

	tr := &TriageResult{
		Entries:      entries,
		TotalScanned: result.TotalScanned,
		TotalTriaged: len(entries),
		BySeverity:   bySeverity,
		TriagedAt:    now,
	}
	tr.SummaryLine = FormatSummaryLine(tr)

	return tr
}

// computeSeverity determines the severity of an orphaned session based on
// the reason it was flagged and how long it has been inactive.
func computeSeverity(s OrphanedSession) Severity {
	switch s.Reason {
	case ReasonIncompleteWrap:
		// Incomplete wraps always need immediate attention.
		return SeverityCritical

	case ReasonInactive:
		// Classify by inactivity duration.
		days := s.InactiveFor.Hours() / 24
		switch {
		case days > 30:
			return SeverityCritical
		case days > 7:
			return SeverityHigh
		default:
			return SeverityMedium
		}

	case ReasonStaleSails:
		// Classify stale sails by how long they have been parked.
		days := s.InactiveFor.Hours() / 24
		switch {
		case days > 14:
			return SeverityHigh
		default:
			return SeverityMedium
		}

	default:
		return SeverityLow
	}
}

// computePriority returns a numeric priority for sorting; lower is more urgent.
// CRITICAL=10, HIGH=20, MEDIUM=30, LOW=40, with INCOMPLETE_WRAP within CRITICAL
// getting priority 9 (lower = more urgent = sorts first).
func computePriority(sev Severity, s OrphanedSession) int {
	base := severityBase(sev)
	// INCOMPLETE_WRAP within CRITICAL sorts first among criticals.
	if sev == SeverityCritical && s.Reason == ReasonIncompleteWrap {
		return base*10 - 1
	}
	return base * 10
}

// severityBase maps severity to a base ordering value.
func severityBase(sev Severity) int {
	switch sev {
	case SeverityCritical:
		return 1
	case SeverityHigh:
		return 2
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 4
	default:
		return 5
	}
}

// isActionable returns true when the severity is high enough that automated
// action suggestions are appropriate.
func isActionable(sev Severity) bool {
	return sev == SeverityCritical || sev == SeverityHigh
}

// FormatSummaryLine produces a compact human-readable summary of the triage
// result suitable for embedding in artifact frontmatter or CLI output.
func FormatSummaryLine(result *TriageResult) string {
	if result.TotalTriaged == 0 {
		return fmt.Sprintf("%d sessions scanned. No orphaned sessions found.", result.TotalScanned)
	}

	session := "session"
	if result.TotalTriaged != 1 {
		session = "sessions"
	}

	// Build severity breakdown — only include levels with non-zero counts.
	parts := []string{}
	for _, sev := range []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow} {
		if count := result.BySeverity[sev]; count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", count, sev))
		}
	}

	breakdown := ""
	if len(parts) > 0 {
		breakdown = " ("
		for i, p := range parts {
			if i > 0 {
				breakdown += ", "
			}
			breakdown += p
		}
		breakdown += ")"
	}

	return fmt.Sprintf("%d orphaned %s%s. Run /naxos to triage.", result.TotalTriaged, session, breakdown)
}

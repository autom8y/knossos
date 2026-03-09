package naxos

import (
	"testing"
	"time"
)

// makeOrphaned builds an OrphanedSession for triage tests.
func makeOrphaned(id string, reason OrphanReason, inactiveFor time.Duration) OrphanedSession {
	now := time.Now().UTC()
	return OrphanedSession{
		SessionID:       id,
		SessionDir:      "/tmp/sessions/" + id,
		Status:          "ACTIVE",
		Initiative:      "Test " + id,
		Reason:          reason,
		SuggestedAction: ActionWrap,
		InactiveFor:     inactiveFor,
		CreatedAt:       now.Add(-inactiveFor),
		LastActivity:    now.Add(-inactiveFor),
	}
}

func TestComputeSeverity_Inactive(t *testing.T) {
	cases := []struct {
		name        string
		inactiveFor time.Duration
		want        Severity
	}{
		{"over 30 days → CRITICAL", 31 * 24 * time.Hour, SeverityCritical},
		{"exactly 30 days boundary → HIGH", 30 * 24 * time.Hour, SeverityHigh},
		{"over 7 days → HIGH", 8 * 24 * time.Hour, SeverityHigh},
		{"exactly 7 days boundary → MEDIUM", 7 * 24 * time.Hour, SeverityMedium},
		{"under 7 days → MEDIUM", 25 * time.Hour, SeverityMedium},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := makeOrphaned("session-abc123", ReasonInactive, tc.inactiveFor)
			got := computeSeverity(s)
			if got != tc.want {
				t.Errorf("computeSeverity() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestComputeSeverity_StaleSails(t *testing.T) {
	cases := []struct {
		name        string
		inactiveFor time.Duration
		want        Severity
	}{
		{"over 14 days → HIGH", 15 * 24 * time.Hour, SeverityHigh},
		{"exactly 14 days boundary → MEDIUM", 14 * 24 * time.Hour, SeverityMedium},
		{"7 days → MEDIUM", 8 * 24 * time.Hour, SeverityMedium},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := makeOrphaned("session-abc123", ReasonStaleSails, tc.inactiveFor)
			got := computeSeverity(s)
			if got != tc.want {
				t.Errorf("computeSeverity() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestComputeSeverity_IncompleteWrap(t *testing.T) {
	s := makeOrphaned("session-abc123", ReasonIncompleteWrap, time.Hour)
	got := computeSeverity(s)
	if got != SeverityCritical {
		t.Errorf("computeSeverity(INCOMPLETE_WRAP) = %q, want CRITICAL", got)
	}
}

func TestTriage_Empty(t *testing.T) {
	result := NewScanResult(DefaultConfig())
	result.TotalScanned = 5

	tr := Triage(result)

	if tr.TotalTriaged != 0 {
		t.Errorf("TotalTriaged = %d, want 0", tr.TotalTriaged)
	}
	if tr.TotalScanned != 5 {
		t.Errorf("TotalScanned = %d, want 5", tr.TotalScanned)
	}
	if len(tr.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(tr.Entries))
	}
	if tr.BySeverity == nil {
		t.Error("BySeverity is nil, want empty map")
	}
}

func TestTriage_SortedByPriority(t *testing.T) {
	result := NewScanResult(DefaultConfig())
	// Add in reverse urgency order: LOW, MEDIUM, HIGH, CRITICAL
	result.Add(makeOrphaned("session-low123456789012345678901234", ReasonStaleSails, 8*24*time.Hour))      // MEDIUM
	result.Add(makeOrphaned("session-crit12345678901234567890123", ReasonIncompleteWrap, time.Hour))      // CRITICAL (incomplete)
	result.Add(makeOrphaned("session-high12345678901234567890123", ReasonInactive, 10*24*time.Hour))     // HIGH
	result.Add(makeOrphaned("session-crit212345678901234567890123", ReasonInactive, 35*24*time.Hour))    // CRITICAL (inactive)

	tr := Triage(result)

	if tr.TotalTriaged != 4 {
		t.Fatalf("TotalTriaged = %d, want 4", tr.TotalTriaged)
	}

	// First two should be CRITICAL
	if tr.Entries[0].Severity != SeverityCritical {
		t.Errorf("Entries[0].Severity = %q, want CRITICAL", tr.Entries[0].Severity)
	}
	if tr.Entries[1].Severity != SeverityCritical {
		t.Errorf("Entries[1].Severity = %q, want CRITICAL", tr.Entries[1].Severity)
	}
	// INCOMPLETE_WRAP should sort before INACTIVE CRITICAL due to priority tie-breaking
	if tr.Entries[0].Reason != ReasonIncompleteWrap {
		t.Errorf("Entries[0].Reason = %q, want INCOMPLETE_WRAP", tr.Entries[0].Reason)
	}
	// Third should be HIGH
	if tr.Entries[2].Severity != SeverityHigh {
		t.Errorf("Entries[2].Severity = %q, want HIGH", tr.Entries[2].Severity)
	}
	// Last should be MEDIUM
	if tr.Entries[3].Severity != SeverityMedium {
		t.Errorf("Entries[3].Severity = %q, want MEDIUM", tr.Entries[3].Severity)
	}
}

func TestTriage_BySeverityCounts(t *testing.T) {
	result := NewScanResult(DefaultConfig())
	result.Add(makeOrphaned("session-c1x123456789012345678901234", ReasonIncompleteWrap, time.Hour))
	result.Add(makeOrphaned("session-c2x123456789012345678901234", ReasonInactive, 35*24*time.Hour))
	result.Add(makeOrphaned("session-h1x123456789012345678901234", ReasonInactive, 10*24*time.Hour))
	result.Add(makeOrphaned("session-m1x123456789012345678901234", ReasonInactive, 2*24*time.Hour))
	result.Add(makeOrphaned("session-m2x123456789012345678901234", ReasonStaleSails, 8*24*time.Hour))

	tr := Triage(result)

	if tr.BySeverity[SeverityCritical] != 2 {
		t.Errorf("BySeverity[CRITICAL] = %d, want 2", tr.BySeverity[SeverityCritical])
	}
	if tr.BySeverity[SeverityHigh] != 1 {
		t.Errorf("BySeverity[HIGH] = %d, want 1", tr.BySeverity[SeverityHigh])
	}
	if tr.BySeverity[SeverityMedium] != 2 {
		t.Errorf("BySeverity[MEDIUM] = %d, want 2", tr.BySeverity[SeverityMedium])
	}
}

func TestTriage_Actionable(t *testing.T) {
	result := NewScanResult(DefaultConfig())
	result.Add(makeOrphaned("session-crit12345678901234567890123", ReasonIncompleteWrap, time.Hour))
	result.Add(makeOrphaned("session-high12345678901234567890123", ReasonInactive, 10*24*time.Hour))
	result.Add(makeOrphaned("session-med1234567890123456789012345", ReasonInactive, 2*24*time.Hour))

	tr := Triage(result)

	// CRITICAL and HIGH should be actionable
	for _, entry := range tr.Entries {
		switch entry.Severity {
		case SeverityCritical, SeverityHigh:
			if !entry.Actionable {
				t.Errorf("entry %s severity=%s: Actionable=false, want true", entry.SessionID, entry.Severity)
			}
		case SeverityMedium, SeverityLow:
			if entry.Actionable {
				t.Errorf("entry %s severity=%s: Actionable=true, want false", entry.SessionID, entry.Severity)
			}
		}
	}
}

func TestTriage_TriagedAt(t *testing.T) {
	before := time.Now().UTC()
	result := NewScanResult(DefaultConfig())
	tr := Triage(result)
	after := time.Now().UTC()

	if tr.TriagedAt.Before(before) || tr.TriagedAt.After(after) {
		t.Errorf("TriagedAt %v not in expected range [%v, %v]", tr.TriagedAt, before, after)
	}
}

func TestFormatSummaryLine_Empty(t *testing.T) {
	tr := &TriageResult{
		TotalScanned: 10,
		TotalTriaged: 0,
		BySeverity:   make(map[Severity]int),
	}
	got := FormatSummaryLine(tr)
	want := "10 sessions scanned. No orphaned sessions found."
	if got != want {
		t.Errorf("FormatSummaryLine() = %q, want %q", got, want)
	}
}

func TestFormatSummaryLine_WithEntries(t *testing.T) {
	tr := &TriageResult{
		TotalScanned: 10,
		TotalTriaged: 3,
		BySeverity: map[Severity]int{
			SeverityCritical: 1,
			SeverityMedium:   2,
		},
	}
	got := FormatSummaryLine(tr)
	want := "3 orphaned sessions (1 CRITICAL, 2 MEDIUM). Run /naxos to triage."
	if got != want {
		t.Errorf("FormatSummaryLine() = %q, want %q", got, want)
	}
}

func TestFormatSummaryLine_Singular(t *testing.T) {
	tr := &TriageResult{
		TotalScanned: 5,
		TotalTriaged: 1,
		BySeverity: map[Severity]int{
			SeverityCritical: 1,
		},
	}
	got := FormatSummaryLine(tr)
	want := "1 orphaned session (1 CRITICAL). Run /naxos to triage."
	if got != want {
		t.Errorf("FormatSummaryLine() = %q, want %q", got, want)
	}
}

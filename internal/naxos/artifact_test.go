package naxos

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// buildTriageResult creates a TriageResult with known entries for testing.
// Session IDs are kept to <=35 characters so they are not truncated in the
// artifact markdown table, enabling a clean round-trip assertion.
func buildTriageResult(t *testing.T) *TriageResult {
	t.Helper()
	now := time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)

	result := NewScanResult(DefaultConfig())
	result.TotalScanned = 10
	result.Add(OrphanedSession{
		SessionID:       "session-20260101-abc123wrap",
		Status:          "ACTIVE",
		Initiative:      "Test Initiative Alpha",
		Reason:          ReasonIncompleteWrap,
		SuggestedAction: ActionWrap,
		InactiveFor:     2 * time.Hour,
		CreatedAt:       now.Add(-48 * time.Hour),
		LastActivity:    now.Add(-2 * time.Hour),
	})
	result.Add(OrphanedSession{
		SessionID:       "session-20260108-zzz789inactive",
		Status:          "ACTIVE",
		Initiative:      "Test Initiative Beta",
		Reason:          ReasonInactive,
		SuggestedAction: ActionResume,
		InactiveFor:     10 * 24 * time.Hour,
		CreatedAt:       now.Add(-10 * 24 * time.Hour),
		LastActivity:    now.Add(-10 * 24 * time.Hour),
	})

	tr := Triage(result)
	// Override TriagedAt to a fixed time for deterministic tests.
	tr.TriagedAt = now
	tr.SummaryLine = FormatSummaryLine(tr)
	return tr
}

func TestWriteReadTriageArtifact_RoundTrip(t *testing.T) {
	sessionsDir := t.TempDir()

	original := buildTriageResult(t)
	if err := WriteTriageArtifact(sessionsDir, original); err != nil {
		t.Fatalf("WriteTriageArtifact() error: %v", err)
	}

	// Verify file exists.
	artifactPath := filepath.Join(sessionsDir, TriageArtifactFile)
	if _, err := os.Stat(artifactPath); err != nil {
		t.Fatalf("artifact file not created: %v", err)
	}

	got, err := ReadTriageArtifact(sessionsDir)
	if err != nil {
		t.Fatalf("ReadTriageArtifact() error: %v", err)
	}

	// Verify summary fields round-trip.
	if got.TotalScanned != original.TotalScanned {
		t.Errorf("TotalScanned: got %d, want %d", got.TotalScanned, original.TotalScanned)
	}
	if got.TotalTriaged != original.TotalTriaged {
		t.Errorf("TotalTriaged: got %d, want %d", got.TotalTriaged, original.TotalTriaged)
	}
	if got.SummaryLine != original.SummaryLine {
		t.Errorf("SummaryLine: got %q, want %q", got.SummaryLine, original.SummaryLine)
	}

	// Verify triaged_at round-trips.
	if !got.TriagedAt.Equal(original.TriagedAt) {
		t.Errorf("TriagedAt: got %v, want %v", got.TriagedAt, original.TriagedAt)
	}

	// Verify by_severity.
	for sev, wantCount := range original.BySeverity {
		if got.BySeverity[sev] != wantCount {
			t.Errorf("BySeverity[%s]: got %d, want %d", sev, got.BySeverity[sev], wantCount)
		}
	}

	// Verify entry count.
	if len(got.Entries) != len(original.Entries) {
		t.Errorf("Entries len: got %d, want %d", len(got.Entries), len(original.Entries))
	}

	// Verify first entry session ID and severity round-trip.
	if len(got.Entries) > 0 {
		gotEntry := got.Entries[0]
		origEntry := original.Entries[0]
		if gotEntry.SessionID != origEntry.SessionID {
			t.Errorf("Entries[0].SessionID: got %q, want %q", gotEntry.SessionID, origEntry.SessionID)
		}
		if gotEntry.Severity != origEntry.Severity {
			t.Errorf("Entries[0].Severity: got %q, want %q", gotEntry.Severity, origEntry.Severity)
		}
		if gotEntry.Reason != origEntry.Reason {
			t.Errorf("Entries[0].Reason: got %q, want %q", gotEntry.Reason, origEntry.Reason)
		}
		if gotEntry.SuggestedAction != origEntry.SuggestedAction {
			t.Errorf("Entries[0].SuggestedAction: got %q, want %q", gotEntry.SuggestedAction, origEntry.SuggestedAction)
		}
	}
}

func TestReadTriageArtifact_NotFound(t *testing.T) {
	sessionsDir := t.TempDir()
	_, err := ReadTriageArtifact(sessionsDir)
	if err == nil {
		t.Error("ReadTriageArtifact() expected error for missing file, got nil")
	}
}

func TestReadTriageSummary_FastPath(t *testing.T) {
	sessionsDir := t.TempDir()

	original := buildTriageResult(t)
	if err := WriteTriageArtifact(sessionsDir, original); err != nil {
		t.Fatalf("WriteTriageArtifact() error: %v", err)
	}

	got := ReadTriageSummary(sessionsDir)
	if got != original.SummaryLine {
		t.Errorf("ReadTriageSummary() = %q, want %q", got, original.SummaryLine)
	}
}

func TestReadTriageSummary_MissingFile(t *testing.T) {
	sessionsDir := t.TempDir()
	got := ReadTriageSummary(sessionsDir)
	if got != "" {
		t.Errorf("ReadTriageSummary() = %q, want empty string for missing file", got)
	}
}

func TestWriteTriageArtifact_EmptyResult(t *testing.T) {
	sessionsDir := t.TempDir()

	result := NewScanResult(DefaultConfig())
	result.TotalScanned = 5
	tr := Triage(result)
	tr.TriagedAt = time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	tr.SummaryLine = FormatSummaryLine(tr)

	if err := WriteTriageArtifact(sessionsDir, tr); err != nil {
		t.Fatalf("WriteTriageArtifact() error: %v", err)
	}

	got, err := ReadTriageArtifact(sessionsDir)
	if err != nil {
		t.Fatalf("ReadTriageArtifact() error: %v", err)
	}

	if got.TotalScanned != 5 {
		t.Errorf("TotalScanned: got %d, want 5", got.TotalScanned)
	}
	if got.TotalTriaged != 0 {
		t.Errorf("TotalTriaged: got %d, want 0", got.TotalTriaged)
	}
	if len(got.Entries) != 0 {
		t.Errorf("Entries len: got %d, want 0", len(got.Entries))
	}
}

func TestWriteTriageArtifact_Idempotent(t *testing.T) {
	sessionsDir := t.TempDir()

	original := buildTriageResult(t)

	// Write twice.
	if err := WriteTriageArtifact(sessionsDir, original); err != nil {
		t.Fatalf("first WriteTriageArtifact() error: %v", err)
	}
	if err := WriteTriageArtifact(sessionsDir, original); err != nil {
		t.Fatalf("second WriteTriageArtifact() error: %v", err)
	}

	// Second write should produce identical file.
	got, err := ReadTriageArtifact(sessionsDir)
	if err != nil {
		t.Fatalf("ReadTriageArtifact() error: %v", err)
	}
	if got.TotalTriaged != original.TotalTriaged {
		t.Errorf("TotalTriaged after second write: got %d, want %d", got.TotalTriaged, original.TotalTriaged)
	}
}

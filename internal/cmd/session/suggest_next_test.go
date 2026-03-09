package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/naxos"
)

// =============================================================================
// suggest-next tests
// =============================================================================

// TestSuggestNext_NoArtifact verifies that a missing NAXOS_TRIAGE.md returns
// HasTriage=false with a "new session" action.
func TestSuggestNext_NoArtifact(t *testing.T) {
	projectDir := setupProjectDir(t)

	ctx := newTestContext(projectDir)
	out := captureSuggestNext(t, ctx)

	if out.HasTriage {
		t.Errorf("HasTriage = true, want false when artifact is absent")
	}
	if out.SuggestedAction != "new session" {
		t.Errorf("SuggestedAction = %q, want %q", out.SuggestedAction, "new session")
	}
	if out.Rationale == "" {
		t.Error("Rationale should not be empty")
	}
}

// TestSuggestNext_NoOrphans verifies that a triage with zero orphans returns
// SuggestedAction="new session" with a healthy-state rationale.
func TestSuggestNext_NoOrphans(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")

	writeTriageArtifact(t, sessionsDir, []*naxos.TriageEntry{})

	ctx := newTestContext(projectDir)
	out := captureSuggestNext(t, ctx)

	if !out.HasTriage {
		t.Errorf("HasTriage = false, want true when artifact exists")
	}
	if out.SuggestedAction != "new session" {
		t.Errorf("SuggestedAction = %q, want %q", out.SuggestedAction, "new session")
	}
	if out.TotalOrphans != 0 {
		t.Errorf("TotalOrphans = %d, want 0", out.TotalOrphans)
	}
}

// TestSuggestNext_CriticalOrphans verifies that critical orphans produce a
// "wrap {id}" action when there is no current session to match against.
func TestSuggestNext_CriticalOrphans(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")

	criticalID := "session-20260301-120000-crit0001"
	writeTriageArtifact(t, sessionsDir, []*naxos.TriageEntry{
		{
			OrphanedSession: naxos.OrphanedSession{
				SessionID:       criticalID,
				Initiative:      "refactor authentication module",
				Reason:          naxos.ReasonIncompleteWrap,
				SuggestedAction: naxos.ActionWrap,
			},
			Severity:   naxos.SeverityCritical,
			Priority:   9,
			Actionable: true,
		},
	})

	// No active session — no initiative matching possible.
	ctx := newTestContext(projectDir)
	out := captureSuggestNext(t, ctx)

	if !out.HasTriage {
		t.Errorf("HasTriage = false, want true")
	}
	if out.CriticalOrphans != 1 {
		t.Errorf("CriticalOrphans = %d, want 1", out.CriticalOrphans)
	}
	if !strings.HasPrefix(out.SuggestedAction, "wrap ") {
		t.Errorf("SuggestedAction = %q, want prefix %q", out.SuggestedAction, "wrap ")
	}
	if !strings.Contains(out.SuggestedAction, criticalID) {
		t.Errorf("SuggestedAction %q does not contain session ID %q", out.SuggestedAction, criticalID)
	}
}

// TestSuggestNext_MatchingInitiative verifies that a critical orphan whose
// initiative shares keywords with the active session produces "resume {id}".
// The orphan's session directory is created with SESSION_CONTEXT.md so
// enrichWithInitiatives can read the initiative from disk.
func TestSuggestNext_MatchingInitiative(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")

	criticalID := "session-20260301-130000-crit0002"

	// Create the orphan session directory with its own SESSION_CONTEXT.md so
	// enrichWithInitiatives can load the initiative from disk.
	writeOrphanSession(t, projectDir, criticalID, "authentication service refactor")

	writeTriageArtifact(t, sessionsDir, []*naxos.TriageEntry{
		{
			OrphanedSession: naxos.OrphanedSession{
				SessionID:       criticalID,
				Reason:          naxos.ReasonIncompleteWrap,
				SuggestedAction: naxos.ActionResume,
			},
			Severity:   naxos.SeverityCritical,
			Priority:   9,
			Actionable: true,
		},
	})

	// Create an active session with an initiative that overlaps the orphan.
	activeSessionID := "session-20260309-090000-active0main"
	writeActiveSession(t, projectDir, activeSessionID, "refactor authentication pipeline")

	ctx := newTestContext(projectDir, activeSessionID)
	out := captureSuggestNext(t, ctx)

	if !out.HasTriage {
		t.Errorf("HasTriage = false, want true")
	}
	if !strings.HasPrefix(out.SuggestedAction, "resume ") {
		t.Errorf("SuggestedAction = %q, want prefix %q", out.SuggestedAction, "resume ")
	}
	if !strings.Contains(out.SuggestedAction, criticalID) {
		t.Errorf("SuggestedAction %q does not contain %q", out.SuggestedAction, criticalID)
	}
	found := false
	for _, id := range out.RelatedOrphans {
		if id == criticalID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("RelatedOrphans %v does not include %q", out.RelatedOrphans, criticalID)
	}
}

// TestSuggestNext_Text verifies that text output contains the key fields.
func TestSuggestNext_Text(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")

	writeTriageArtifact(t, sessionsDir, []*naxos.TriageEntry{})

	ctx := newTestContext(projectDir)
	out := captureSuggestNext(t, ctx)

	text := out.Text()
	if !strings.Contains(text, "Suggested:") {
		t.Errorf("Text() output missing 'Suggested:' line:\n%s", text)
	}
	if !strings.Contains(text, "Rationale:") {
		t.Errorf("Text() output missing 'Rationale:' line:\n%s", text)
	}
}

// =============================================================================
// Helpers
// =============================================================================

// captureSuggestNext runs runSuggestNext with a JSON printer and unmarshals the
// output into a SuggestNextOutput. It fails the test immediately on any error.
func captureSuggestNext(t *testing.T, ctx *cmdContext) SuggestNextOutput {
	t.Helper()

	var buf strings.Builder
	outFmt := "json"
	verbose := false
	captureCtx := &cmdContext{
		SessionContext: ctx.SessionContext,
	}
	captureCtx.Output = &outFmt
	captureCtx.Verbose = &verbose

	// Redirect printer output to buf by temporarily replacing the context output.
	// We do this by creating a fresh printer — the printer in runSuggestNext
	// reads ctx.Output, so we can override it by mutating the flag pointer.
	captureCtx.SessionContext.BaseContext.Output = &outFmt

	// Use a pipe-backed approach: create a temp file to capture stdout.
	tmpFile := filepath.Join(t.TempDir(), "out.json")
	origStdout := os.Stdout
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("captureSuggestNext: create temp file: %v", err)
	}
	os.Stdout = f

	runErr := runSuggestNext(captureCtx)

	f.Close()
	os.Stdout = origStdout

	if runErr != nil {
		t.Fatalf("runSuggestNext returned error: %v", runErr)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("captureSuggestNext: read output: %v", err)
	}

	_ = buf // not used in pipe approach, keep for clarity

	var out SuggestNextOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("captureSuggestNext: unmarshal JSON %q: %v", string(data), err)
	}

	return out
}

// writeTriageArtifact creates a NAXOS_TRIAGE.md in sessionsDir populated with
// the provided entries.
func writeTriageArtifact(t *testing.T, sessionsDir string, entries []*naxos.TriageEntry) {
	t.Helper()

	now := time.Now().UTC()
	bySeverity := make(map[naxos.Severity]int)
	flatEntries := make([]naxos.TriageEntry, 0, len(entries))
	for _, e := range entries {
		bySeverity[e.Severity]++
		flatEntries = append(flatEntries, *e)
	}

	result := &naxos.TriageResult{
		Entries:      flatEntries,
		TotalScanned: len(entries) + 2,
		TotalTriaged: len(entries),
		BySeverity:   bySeverity,
		SummaryLine:  naxos.FormatSummaryLine(&naxos.TriageResult{TotalTriaged: len(entries), TotalScanned: len(entries) + 2, BySeverity: bySeverity}),
		TriagedAt:    now,
	}

	if err := naxos.WriteTriageArtifact(sessionsDir, result); err != nil {
		t.Fatalf("writeTriageArtifact: %v", err)
	}
}

// writeOrphanSession creates a SESSION_CONTEXT.md for a PARKED session with
// the given initiative. It does NOT update .current-session — it represents an
// orphaned session that the triage artifact references.
func writeOrphanSession(t *testing.T, projectDir, sessionID, initiative string) {
	t.Helper()

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("writeOrphanSession: mkdir %s: %v", sessionDir, err)
	}

	parkedAt := time.Now().UTC().Add(-48 * time.Hour)
	content := "---\n" +
		"schema_version: \"2.1\"\n" +
		"session_id: " + sessionID + "\n" +
		"status: PARKED\n" +
		"initiative: " + initiative + "\n" +
		"complexity: MODULE\n" +
		"created_at: " + parkedAt.Add(-time.Hour).Format(time.RFC3339) + "\n" +
		"parked_at: " + parkedAt.Format(time.RFC3339) + "\n" +
		"---\n\n# Session Context\n"

	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("writeOrphanSession: write context: %v", err)
	}
}

// writeActiveSession creates a SESSION_CONTEXT.md for an ACTIVE session with
// the given initiative, and writes .current-session to point at it.
func writeActiveSession(t *testing.T, projectDir, sessionID, initiative string) {
	t.Helper()

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("writeActiveSession: mkdir %s: %v", sessionDir, err)
	}

	content := "---\n" +
		"schema_version: \"2.1\"\n" +
		"session_id: " + sessionID + "\n" +
		"status: ACTIVE\n" +
		"initiative: " + initiative + "\n" +
		"complexity: MODULE\n" +
		"created_at: " + time.Now().UTC().Add(-30*time.Minute).Format(time.RFC3339) + "\n" +
		"---\n\n# Session Context\n"

	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("writeActiveSession: write context: %v", err)
	}

	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	if err := os.WriteFile(currentSessionFile, []byte(sessionID), 0644); err != nil {
		t.Fatalf("writeActiveSession: write .current-session: %v", err)
	}
}

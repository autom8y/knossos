package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

// setupTimelineTestSession creates a minimal session environment for timeline command tests.
// Returns (ctx, sessionsDir, sessionID, sessionDir).
func setupTimelineTestSession(t *testing.T) (*cmdContext, string, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectDir := tmpDir
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	sessionID := "session-20260226-120000-tltestxx"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	for _, dir := range []string{sessionDir, locksDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Write SESSION_CONTEXT.md with an empty ## Timeline section.
	contextContent := "---\n" +
		"schema_version: \"2.1\"\n" +
		"session_id: " + sessionID + "\n" +
		"status: ACTIVE\n" +
		"initiative: timeline test\n" +
		"complexity: MODULE\n" +
		"active_rite: ecosystem\n" +
		"current_phase: requirements\n" +
		"created_at: 2026-02-26T12:00:00Z\n" +
		"---\n\n# Session: timeline test\n\n## Timeline\n\n## Artifacts\n- PRD: pending\n"

	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(ctxPath, []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Write .current-session marker.
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	return ctx, sessionsDir, sessionID, sessionDir
}

// writeTimelineEntries appends pre-formatted timeline entries to SESSION_CONTEXT.md.
func writeTimelineEntries(t *testing.T, ctxPath string, entries []string) {
	t.Helper()
	for _, entry := range entries {
		if err := sess.AppendEntry(ctxPath, entry); err != nil {
			t.Fatalf("failed to append timeline entry %q: %v", entry, err)
		}
	}
}

// writeTypedEventsJSONL writes TypedEvents to events.jsonl in a session directory.
func writeTypedEventsJSONL(t *testing.T, sessionDir string, events []clewcontract.TypedEvent) {
	t.Helper()
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	f, err := os.OpenFile(eventsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("failed to open events.jsonl: %v", err)
	}
	defer f.Close()

	for _, e := range events {
		line, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("failed to marshal event: %v", err)
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			t.Fatalf("failed to write event: %v", err)
		}
	}
}

// TestRunTimeline_EmptyTimeline verifies empty timeline produces empty output without error.
func TestRunTimeline_EmptyTimeline(t *testing.T) {
	ctx, _, _, _ := setupTimelineTestSession(t)

	err := runTimeline(ctx, timelineOptions{})
	if err != nil {
		t.Fatalf("runTimeline with empty timeline returned error: %v", err)
	}
}

// TestRunTimeline_NoActiveSession verifies error returned when no session exists.
func TestRunTimeline_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("failed to create locks dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runTimeline(ctx, timelineOptions{})
	if err == nil {
		t.Fatal("expected error when no active session")
	}
}

// TestFilterTimeline_SinceFilter verifies --since excludes entries older than the cutoff.
func TestFilterTimeline_SinceFilter(t *testing.T) {
	now := time.Now()

	// Create an entry at "now minus 2 hours" (should be excluded by --since 1h).
	oldTime := now.Add(-2 * time.Hour)
	// Create an entry at "now minus 30 minutes" (should be included by --since 1h).
	recentTime := now.Add(-30 * time.Minute)

	entries := []sess.TimelineEntry{
		{
			Time:     time.Date(0, 1, 1, oldTime.Hour(), oldTime.Minute(), 0, 0, time.UTC),
			Category: "SESSION ",
			Summary:  "old entry",
			Raw:      "- 00:00 | SESSION  | old entry",
		},
		{
			Time:     time.Date(0, 1, 1, recentTime.Hour(), recentTime.Minute(), 0, 0, time.UTC),
			Category: "DECISION",
			Summary:  "recent entry",
			Raw:      "- 00:00 | DECISION | recent entry",
		},
	}

	filtered := filterTimeline(entries, timelineOptions{since: time.Hour})

	// Only the recent entry should pass the --since 1h filter.
	if len(filtered) != 1 {
		t.Errorf("filtered len = %d, want 1 (only recent entry)", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Summary != "recent entry" {
		t.Errorf("filtered[0].Summary = %q, want %q", filtered[0].Summary, "recent entry")
	}
}

// TestFilterTimeline_TypeFilter verifies --type filters case-insensitively by category.
func TestFilterTimeline_TypeFilter(t *testing.T) {
	entries := []sess.TimelineEntry{
		{Time: time.Time{}, Category: "SESSION ", Summary: "session event"},
		{Time: time.Time{}, Category: "DECISION", Summary: "a decision"},
		{Time: time.Time{}, Category: "AGENT   ", Summary: "agent delegation"},
		{Time: time.Time{}, Category: "COMMIT  ", Summary: "a commit"},
	}

	// Filter by "decision" (lowercase — should match "DECISION").
	filtered := filterTimeline(entries, timelineOptions{eventType: "decision"})
	if len(filtered) != 1 {
		t.Errorf("filtered len = %d, want 1", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Summary != "a decision" {
		t.Errorf("filtered[0].Summary = %q, want %q", filtered[0].Summary, "a decision")
	}

	// Filter by "AGENT" (uppercase — should match "AGENT   ").
	filtered = filterTimeline(entries, timelineOptions{eventType: "AGENT"})
	if len(filtered) != 1 {
		t.Errorf("filtered len = %d, want 1", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Summary != "agent delegation" {
		t.Errorf("filtered[0].Summary = %q, want %q", filtered[0].Summary, "agent delegation")
	}
}

// TestFilterTimeline_LastFilter verifies --last returns only N most recent entries after other filters.
func TestFilterTimeline_LastFilter(t *testing.T) {
	entries := []sess.TimelineEntry{
		{Time: time.Time{}, Category: "SESSION ", Summary: "entry 1"},
		{Time: time.Time{}, Category: "DECISION", Summary: "entry 2"},
		{Time: time.Time{}, Category: "AGENT   ", Summary: "entry 3"},
		{Time: time.Time{}, Category: "COMMIT  ", Summary: "entry 4"},
		{Time: time.Time{}, Category: "DECISION", Summary: "entry 5"},
	}

	// Request last 2 entries.
	filtered := filterTimeline(entries, timelineOptions{last: 2})
	if len(filtered) != 2 {
		t.Errorf("filtered len = %d, want 2", len(filtered))
	}
	// Should be the last 2 entries: entry 4 and entry 5.
	if len(filtered) >= 1 && filtered[0].Summary != "entry 4" {
		t.Errorf("filtered[0].Summary = %q, want %q", filtered[0].Summary, "entry 4")
	}
	if len(filtered) >= 2 && filtered[1].Summary != "entry 5" {
		t.Errorf("filtered[1].Summary = %q, want %q", filtered[1].Summary, "entry 5")
	}
}

// TestFilterTimeline_LastAppliedAfterType verifies --last is applied after --type filter.
func TestFilterTimeline_LastAppliedAfterType(t *testing.T) {
	entries := []sess.TimelineEntry{
		{Time: time.Time{}, Category: "SESSION ", Summary: "session 1"},
		{Time: time.Time{}, Category: "DECISION", Summary: "decision 1"},
		{Time: time.Time{}, Category: "DECISION", Summary: "decision 2"},
		{Time: time.Time{}, Category: "DECISION", Summary: "decision 3"},
	}

	// Filter type=DECISION, last=2 — should get the last 2 decisions.
	filtered := filterTimeline(entries, timelineOptions{eventType: "DECISION", last: 2})
	if len(filtered) != 2 {
		t.Errorf("filtered len = %d, want 2", len(filtered))
	}
	if len(filtered) >= 1 && filtered[0].Summary != "decision 2" {
		t.Errorf("filtered[0].Summary = %q, want %q", filtered[0].Summary, "decision 2")
	}
	if len(filtered) >= 2 && filtered[1].Summary != "decision 3" {
		t.Errorf("filtered[1].Summary = %q, want %q", filtered[1].Summary, "decision 3")
	}
}

// TestRunTimeline_FromEvents reads TypedEvents from events.jsonl.
func TestRunTimeline_FromEvents(t *testing.T) {
	ctx, _, _, sessionDir := setupTimelineTestSession(t)

	// Write two v3 TypedEvents to events.jsonl.
	events := []clewcontract.TypedEvent{
		clewcontract.NewTypedDecisionRecordedEvent("chose X over Y", "perf reason", nil),
		clewcontract.NewTypedAgentDelegatedEvent(clewcontract.SourceAgent, "architect", "", "design task", ""),
	}
	writeTypedEventsJSONL(t, sessionDir, events)

	err := runTimeline(ctx, timelineOptions{fromEvents: true})
	if err != nil {
		t.Fatalf("runTimeline --from-events returned error: %v", err)
	}

	// Verify by reading from events path directly.
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	entries, total, err := readFromEvents(eventsPath, timelineOptions{})
	if err != nil {
		t.Fatalf("readFromEvents failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(entries) != 2 {
		t.Errorf("entries len = %d, want 2", len(entries))
	}
}

// TestRunTimeline_FromEvents_EmptyFile verifies empty events.jsonl produces empty output.
func TestRunTimeline_FromEvents_EmptyFile(t *testing.T) {
	_, _, _, sessionDir := setupTimelineTestSession(t)

	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	entries, total, err := readFromEvents(eventsPath, timelineOptions{})
	if err != nil {
		t.Fatalf("readFromEvents on empty/missing file returned error: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(entries) != 0 {
		t.Errorf("entries len = %d, want 0", len(entries))
	}
}

// TestRunTimeline_FromEvents_NonCuratedEventsFiltered verifies non-curated events are excluded.
func TestRunTimeline_FromEvents_NonCuratedEventsFiltered(t *testing.T) {
	_, _, _, sessionDir := setupTimelineTestSession(t)

	// Write a curated v3 event (decision.recorded) and a non-curated v3 event (tool.use).
	// tool.use is a non-curated type — write it directly as JSONL.
	curatedEvent := clewcontract.NewTypedDecisionRecordedEvent("important decision", "", nil)

	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	f, err := os.Create(eventsPath)
	if err != nil {
		t.Fatalf("failed to create events.jsonl: %v", err)
	}

	// Write the curated event.
	curatedLine, _ := json.Marshal(curatedEvent)
	f.Write(append(curatedLine, '\n'))

	// Write a non-curated v3 event directly as JSON (tool.use is not in IsCuratedType).
	nonCuratedLine := `{"ts":"2026-02-26T12:05:00.000Z","type":"tool.use","source":"hook","data":{"tool":"Bash","input":{}}}`
	f.Write([]byte(nonCuratedLine + "\n"))
	f.Close()

	entries, total, err := readFromEvents(eventsPath, timelineOptions{})
	if err != nil {
		t.Fatalf("readFromEvents failed: %v", err)
	}
	// Only the curated event should appear.
	if total != 1 {
		t.Errorf("total = %d, want 1 (only curated event)", total)
	}
	if len(entries) != 1 {
		t.Errorf("entries len = %d, want 1", len(entries))
	}
	if len(entries) > 0 && entries[0].Category != "DECISION" {
		t.Errorf("entries[0].Category = %q, want DECISION", entries[0].Category)
	}
}

// TestRunTimeline_JSONOutput verifies JSON output matches expected structure.
func TestRunTimeline_JSONOutput(t *testing.T) {
	_, _, sessionID, sessionDir := setupTimelineTestSession(t)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

	// Add two timeline entries.
	writeTimelineEntries(t, ctxPath, []string{
		"- 14:03 | SESSION  | created: test (MODULE)",
		"- 14:05 | DECISION | chose X over Y",
	})

	// Read back and verify structure via readFromTimeline.
	entries, total, err := readFromTimeline(ctxPath, timelineOptions{})
	if err != nil {
		t.Fatalf("readFromTimeline failed: %v", err)
	}

	result := output.TimelineOutput{
		SessionID: sessionID,
		Entries:   entries,
		Total:     total,
		Filtered:  len(entries),
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Verify JSON structure.
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if parsed["session_id"] != sessionID {
		t.Errorf("session_id = %v, want %q", parsed["session_id"], sessionID)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if result.Filtered != 2 {
		t.Errorf("filtered = %d, want 2", result.Filtered)
	}

	// Verify entries array in JSON.
	entriesRaw, ok := parsed["entries"].([]any)
	if !ok {
		t.Fatalf("entries field is not an array")
	}
	if len(entriesRaw) != 2 {
		t.Errorf("entries array len = %d, want 2", len(entriesRaw))
	}

	// Verify first entry has time, category, summary fields.
	if len(entriesRaw) > 0 {
		entry0, ok := entriesRaw[0].(map[string]any)
		if !ok {
			t.Fatal("entries[0] is not a map")
		}
		if entry0["time"] == "" {
			t.Error("entries[0].time is empty")
		}
		if entry0["category"] == "" {
			t.Error("entries[0].category is empty")
		}
		if entry0["summary"] == "" {
			t.Error("entries[0].summary is empty")
		}
	}
}

// TestTimelineOutput_TextFormat verifies Text() output format.
func TestTimelineOutput_TextFormat(t *testing.T) {
	result := output.TimelineOutput{
		SessionID: "session-20260226-120000-tltestxx",
		Entries: []output.TimelineEntryOutput{
			{Time: "14:03", Category: "SESSION", Summary: "created: test (MODULE)"},
			{Time: "14:05", Category: "DECISION", Summary: "chose X over Y"},
		},
		Total:    2,
		Filtered: 2,
	}

	text := result.Text()

	if !strings.Contains(text, "Timeline for session-20260226-120000-tltestxx:") {
		t.Errorf("text missing session header: %q", text)
	}
	if !strings.Contains(text, "14:03") {
		t.Errorf("text missing first entry time: %q", text)
	}
	if !strings.Contains(text, "SESSION") {
		t.Errorf("text missing SESSION category: %q", text)
	}
	if !strings.Contains(text, "chose X over Y") {
		t.Errorf("text missing decision summary: %q", text)
	}
}

// TestTimelineOutput_EmptyText verifies Text() on empty entries shows "(no entries)".
func TestTimelineOutput_EmptyText(t *testing.T) {
	result := output.TimelineOutput{
		SessionID: "session-20260226-120000-tltestxx",
		Entries:   []output.TimelineEntryOutput{},
		Total:     0,
		Filtered:  0,
	}

	text := result.Text()
	if !strings.Contains(text, "(no entries)") {
		t.Errorf("empty timeline text = %q, want (no entries)", text)
	}
}

// TestReadFromTimeline_FilterByType verifies readFromTimeline applies --type filter.
func TestReadFromTimeline_FilterByType(t *testing.T) {
	_, _, _, sessionDir := setupTimelineTestSession(t)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

	writeTimelineEntries(t, ctxPath, []string{
		"- 14:03 | SESSION  | created: test (MODULE)",
		"- 14:05 | DECISION | chose X over Y",
		"- 14:07 | AGENT    | delegated architect",
	})

	entries, total, err := readFromTimeline(ctxPath, timelineOptions{eventType: "DECISION"})
	if err != nil {
		t.Fatalf("readFromTimeline failed: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3 (total before filter)", total)
	}
	if len(entries) != 1 {
		t.Errorf("filtered len = %d, want 1", len(entries))
	}
	if len(entries) > 0 && entries[0].Category != "DECISION" {
		t.Errorf("entries[0].Category = %q, want DECISION", entries[0].Category)
	}
}

// TestReadFromTimeline_FilterByLast verifies readFromTimeline applies --last filter.
func TestReadFromTimeline_FilterByLast(t *testing.T) {
	_, _, _, sessionDir := setupTimelineTestSession(t)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

	writeTimelineEntries(t, ctxPath, []string{
		"- 14:03 | SESSION  | created: test (MODULE)",
		"- 14:05 | DECISION | decision 1",
		"- 14:07 | DECISION | decision 2",
		"- 14:09 | COMMIT   | abc123f: feat: theme",
	})

	entries, total, err := readFromTimeline(ctxPath, timelineOptions{last: 2})
	if err != nil {
		t.Fatalf("readFromTimeline failed: %v", err)
	}
	if total != 4 {
		t.Errorf("total = %d, want 4", total)
	}
	if len(entries) != 2 {
		t.Errorf("filtered len = %d, want 2", len(entries))
	}
	// Last 2 entries: decision 2 and the commit.
	if len(entries) >= 1 && entries[0].Category != "DECISION" {
		t.Errorf("entries[0].Category = %q, want DECISION", entries[0].Category)
	}
	if len(entries) >= 2 && entries[1].Category != "COMMIT" {
		t.Errorf("entries[1].Category = %q, want COMMIT", entries[1].Category)
	}
}

// TestReadFromTimeline_CategoryTrimmed verifies category field in output has no trailing spaces.
func TestReadFromTimeline_CategoryTrimmed(t *testing.T) {
	_, _, _, sessionDir := setupTimelineTestSession(t)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

	// "SESSION " (padded to 8 chars) — output should be trimmed.
	writeTimelineEntries(t, ctxPath, []string{
		"- 14:03 | SESSION  | created: test (MODULE)",
	})

	entries, _, err := readFromTimeline(ctxPath, timelineOptions{})
	if err != nil {
		t.Fatalf("readFromTimeline failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(entries))
	}
	// Category should be trimmed of trailing spaces.
	if entries[0].Category != "SESSION" {
		t.Errorf("entries[0].Category = %q, want %q (no trailing spaces)", entries[0].Category, "SESSION")
	}
}

// TestReadTypedEventsFromPath_SkipsNonV3Lines verifies v1/v2 lines are skipped.
func TestReadTypedEventsFromPath_SkipsNonV3Lines(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, "events.jsonl")

	// Write a mix of v1 legacy, v2 flat, and v3 typed events.
	lines := []string{
		// v1 legacy: has "event" field but no "data" field.
		`{"timestamp":"2026-02-26T12:00:00Z","event":"session.created","from":"NONE","to":"ACTIVE"}`,
		// v2 flat: has "type" field but no "data" field.
		`{"ts":"2026-02-26T12:01:00.000Z","type":"agent.delegated","summary":"something"}`,
		// v3 typed: has "data" field.
		`{"ts":"2026-02-26T12:02:00.000Z","type":"decision.recorded","source":"agent","data":{"decision":"chose X","rationale":""}}`,
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(eventsPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write events.jsonl: %v", err)
	}

	events, err := readTypedEventsFromPath(eventsPath)
	if err != nil {
		t.Fatalf("readTypedEventsFromPath failed: %v", err)
	}

	// Only the v3 event should be returned.
	if len(events) != 1 {
		t.Errorf("events len = %d, want 1 (only v3 event)", len(events))
	}
	if len(events) > 0 && events[0].Type != clewcontract.EventTypeDecisionRecorded {
		t.Errorf("events[0].Type = %q, want %q", events[0].Type, clewcontract.EventTypeDecisionRecorded)
	}
}

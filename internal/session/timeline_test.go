package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/hook/clewcontract"
)

// --- FormatTimelineEntry tests ---

// TestFormatTimelineEntry_SessionCreated verifies session.created produces correct entry.
func TestFormatTimelineEntry_SessionCreated(t *testing.T) {
	event := clewcontract.NewTypedSessionCreatedEvent("session-001", "Add dark mode", "MODULE", "")
	// Patch timestamp to a known value for deterministic output.
	event.Ts = "2026-03-01T14:03:22.500Z"

	entry := FormatTimelineEntry(event)

	if !strings.HasPrefix(entry, "- 14:03 | SESSION  |") {
		t.Errorf("entry prefix = %q, want prefix \"- 14:03 | SESSION  |\"", entry)
	}
	if !strings.Contains(entry, "created: Add dark mode (MODULE)") {
		t.Errorf("entry = %q, want summary containing \"created: Add dark mode (MODULE)\"", entry)
	}
}

// TestFormatTimelineEntry_CategoryPadding verifies PHASE is padded to 8 chars (TF-02).
func TestFormatTimelineEntry_CategoryPadding(t *testing.T) {
	event := clewcontract.NewTypedPhaseTransitionedEvent("session-001", "requirements", "design")
	event.Ts = "2026-03-01T14:20:00.000Z"

	entry := FormatTimelineEntry(event)

	// "PHASE   " is 8 chars (3 trailing spaces)
	if !strings.Contains(entry, "| PHASE    |") {
		t.Errorf("entry = %q, want PHASE padded to 8 chars (PHASE   )", entry)
	}
}

// TestFormatTimelineEntry_TimeExtraction verifies UTC HH:MM extraction (TF-03).
func TestFormatTimelineEntry_TimeExtraction(t *testing.T) {
	event := clewcontract.NewTypedSessionCreatedEvent("session-001", "test", "PATCH", "")
	event.Ts = "2026-03-01T14:03:22.500Z"

	entry := FormatTimelineEntry(event)

	if !strings.HasPrefix(entry, "- 14:03 |") {
		t.Errorf("entry = %q, time component should be 14:03", entry)
	}
}

// TestFormatTimelineEntry_SummaryTruncation verifies 80-char summary truncation (TF-01).
func TestFormatTimelineEntry_SummaryTruncation(t *testing.T) {
	longDecision := strings.Repeat("x", 100)
	event := clewcontract.NewTypedDecisionRecordedEvent(longDecision, "", nil)
	event.Ts = "2026-03-01T14:15:00.000Z"

	entry := FormatTimelineEntry(event)

	// Extract summary portion (after "- HH:MM | DECISION |  ")
	parts := strings.SplitN(entry, " | ", 3)
	if len(parts) != 3 {
		t.Fatalf("entry %q does not have 3 pipe-delimited parts", entry)
	}
	summary := parts[2]
	if len(summary) > 80 {
		t.Errorf("summary length = %d, want <= 80", len(summary))
	}
	if !strings.HasSuffix(summary, "...") {
		t.Errorf("truncated summary should end with '...', got %q", summary)
	}
}

// TestFormatTimelineEntry_CommitSHAShortening verifies SHA is truncated to 7 chars (TF-04).
func TestFormatTimelineEntry_CommitSHAShortening(t *testing.T) {
	event := clewcontract.NewTypedCommitCreatedEvent("abc123f890abcdef1234567890", "feat: theme provider")
	event.Ts = "2026-03-01T14:12:00.000Z"

	entry := FormatTimelineEntry(event)

	if !strings.Contains(entry, "abc123f: feat: theme provider") {
		t.Errorf("entry = %q, want SHA shortened to 7 chars and message included", entry)
	}
}

// TestFormatTimelineEntry_SessionWrappedAbsentSailsColor verifies graceful degradation (TF-05).
func TestFormatTimelineEntry_SessionWrappedAbsentSailsColor(t *testing.T) {
	event := clewcontract.NewTypedSessionWrappedEvent("session-001", "", 0)
	event.Ts = "2026-03-01T15:00:00.000Z"

	entry := FormatTimelineEntry(event)

	if !strings.Contains(entry, "wrapped") {
		t.Errorf("entry = %q, want summary to contain 'wrapped'", entry)
	}
	if strings.Contains(entry, "(") {
		t.Errorf("entry = %q, absent sails_color should not add parentheses", entry)
	}
}

// TestFormatTimelineEntry_SessionWrappedWithSailsColor verifies color is included.
func TestFormatTimelineEntry_SessionWrappedWithSailsColor(t *testing.T) {
	event := clewcontract.NewTypedSessionWrappedEvent("session-001", "WHITE", 60000)
	event.Ts = "2026-03-01T15:00:00.000Z"

	entry := FormatTimelineEntry(event)

	if !strings.Contains(entry, "wrapped (WHITE)") {
		t.Errorf("entry = %q, want summary 'wrapped (WHITE)'", entry)
	}
}

// TestFormatTimelineEntry_MalformedTimestamp verifies fallback to 00:00 (edge case).
func TestFormatTimelineEntry_MalformedTimestamp(t *testing.T) {
	event := clewcontract.NewTypedSessionCreatedEvent("session-001", "test", "PATCH", "")
	event.Ts = "not-a-timestamp"

	entry := FormatTimelineEntry(event)

	if !strings.HasPrefix(entry, "- 00:00 |") {
		t.Errorf("entry = %q, malformed timestamp should fall back to 00:00", entry)
	}
}

// TestFormatTimelineEntry_AgentDelegated verifies agent delegation summary.
func TestFormatTimelineEntry_AgentDelegated(t *testing.T) {
	event := clewcontract.NewTypedAgentDelegatedEvent(
		clewcontract.SourceAgent, "architect", "", "component design", "")
	event.Ts = "2026-03-01T14:05:00.000Z"

	entry := FormatTimelineEntry(event)

	if !strings.Contains(entry, "delegated architect: component design") {
		t.Errorf("entry = %q, want agent delegation summary", entry)
	}
}

// TestFormatTimelineEntry_DecisionWithRationale verifies decision + rationale format.
func TestFormatTimelineEntry_DecisionWithRationale(t *testing.T) {
	event := clewcontract.NewTypedDecisionRecordedEvent("CSS vars over styled-components", "runtime perf", nil)
	event.Ts = "2026-03-01T14:15:00.000Z"

	entry := FormatTimelineEntry(event)

	if !strings.Contains(entry, "CSS vars over styled-components (runtime perf)") {
		t.Errorf("entry = %q, want decision with rationale", entry)
	}
}

// TestFormatTimelineEntry_CommandInvoked verifies command.invoked summary.
func TestFormatTimelineEntry_CommandInvoked(t *testing.T) {
	event := clewcontract.NewTypedCommandInvokedEvent("/consult", "skill")
	event.Ts = "2026-03-01T14:30:00.000Z"

	entry := FormatTimelineEntry(event)

	if !strings.Contains(entry, "/consult") {
		t.Errorf("entry = %q, want command name in summary", entry)
	}
}

// --- IsCuratedType tests ---

// TestIsCuratedType_CuratedTypes verifies all 11 curated types return true (TC-01).
func TestIsCuratedType_CuratedTypes(t *testing.T) {
	curated := []clewcontract.EventType{
		clewcontract.EventTypeSessionCreated,
		clewcontract.EventTypeSessionParked,
		clewcontract.EventTypeSessionResumed,
		clewcontract.EventTypeSessionWrapped,
		clewcontract.EventTypeSessionFrayed,
		clewcontract.EventTypePhaseTransitioned,
		clewcontract.EventTypeAgentDelegated,
		clewcontract.EventTypeAgentCompleted,
		clewcontract.EventTypeCommitCreated,
		clewcontract.EventTypeDecisionRecorded,
		clewcontract.EventTypeCommandInvoked,
	}
	for _, et := range curated {
		if !IsCuratedType(et) {
			t.Errorf("IsCuratedType(%q) = false, want true", et)
		}
	}
}

// TestIsCuratedType_BackplaneTypes verifies backplane-only types return false (TC-02).
func TestIsCuratedType_BackplaneTypes(t *testing.T) {
	notCurated := []clewcontract.EventType{
		clewcontract.EventTypeToolInvoked,
		clewcontract.EventTypeFileModified,
		clewcontract.EventTypeArtifactCreatedV3,
		clewcontract.EventTypeErrorOccurred,
		clewcontract.EventTypeSessionStart,
		clewcontract.EventTypeSessionEnd,
		clewcontract.EventTypeLockAcquired,
		clewcontract.EventTypeLockReleased,
	}
	for _, et := range notCurated {
		if IsCuratedType(et) {
			t.Errorf("IsCuratedType(%q) = true, want false", et)
		}
	}
}

// --- AppendEntry tests ---

// TestAppendEntry_CreatesTimelineSection verifies section creation when missing.
func TestAppendEntry_CreatesTimelineSection(t *testing.T) {
	dir := t.TempDir()
	ctxPath := filepath.Join(dir, "SESSION_CONTEXT.md")

	// Write a context file without a ## Timeline section.
	content := "---\nsession_id: test-001\nstatus: ACTIVE\n---\n\n# Session: test\n\n## Artifacts\n- PRD: pending\n"
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write context file: %v", err)
	}

	entry := "- 14:03 | SESSION  | created: test (MODULE)"
	if err := AppendEntry(ctxPath, entry); err != nil {
		t.Fatalf("AppendEntry failed: %v", err)
	}

	updated, err := os.ReadFile(ctxPath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	body := string(updated)
	if !strings.Contains(body, "## Timeline") {
		t.Error("## Timeline section should have been created")
	}
	if !strings.Contains(body, entry) {
		t.Errorf("entry %q should be in the file", entry)
	}
}

// TestAppendEntry_AppendsToExistingSection verifies appending to existing ## Timeline section.
func TestAppendEntry_AppendsToExistingSection(t *testing.T) {
	dir := t.TempDir()
	ctxPath := filepath.Join(dir, "SESSION_CONTEXT.md")

	content := "---\nsession_id: test-001\nstatus: ACTIVE\n---\n\n# Session: test\n\n## Timeline\n- 14:03 | SESSION  | created: test (MODULE)\n\n## Artifacts\n- PRD: pending\n"
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write context file: %v", err)
	}

	entry2 := "- 14:05 | AGENT    | delegated architect: design"
	if err := AppendEntry(ctxPath, entry2); err != nil {
		t.Fatalf("AppendEntry failed: %v", err)
	}

	updated, err := os.ReadFile(ctxPath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	body := string(updated)
	if !strings.Contains(body, "created: test (MODULE)") {
		t.Error("original timeline entry should be preserved")
	}
	if !strings.Contains(body, entry2) {
		t.Errorf("new entry %q should be in the file", entry2)
	}
	// Both entries should be before ## Artifacts
	timelineIdx := strings.Index(body, "## Timeline")
	artifactsIdx := strings.Index(body, "## Artifacts")
	entry2Idx := strings.Index(body, entry2)
	if entry2Idx > artifactsIdx {
		t.Error("new timeline entry should be inserted before ## Artifacts section")
	}
	if entry2Idx < timelineIdx {
		t.Error("new timeline entry should be after ## Timeline heading")
	}
}

// TestAppendEntry_FileNotFound verifies error when file is missing.
func TestAppendEntry_FileNotFound(t *testing.T) {
	err := AppendEntry("/nonexistent/path/SESSION_CONTEXT.md", "- 14:03 | SESSION  | created")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// --- ReadTimeline tests ---

// TestReadTimeline_ReturnsEntries verifies entries are read correctly.
func TestReadTimeline_ReturnsEntries(t *testing.T) {
	dir := t.TempDir()
	ctxPath := filepath.Join(dir, "SESSION_CONTEXT.md")

	content := "---\nsession_id: test-001\nstatus: ACTIVE\n---\n\n## Timeline\n" +
		"- 14:03 | SESSION  | created: test (MODULE)\n" +
		"- 14:05 | AGENT    | delegated architect: design\n" +
		"- 14:12 | COMMIT   | abc123f: feat: theme provider\n" +
		"\n## Artifacts\n"
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write context file: %v", err)
	}

	entries, err := ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("len(entries) = %d, want 3", len(entries))
	}
	if entries[0].Category != "SESSION " {
		t.Errorf("entries[0].Category = %q, want %q", entries[0].Category, "SESSION ")
	}
	if entries[1].Category != "AGENT   " {
		t.Errorf("entries[1].Category = %q, want %q", entries[1].Category, "AGENT   ")
	}
}

// TestReadTimeline_EmptySection verifies empty slice returned for empty section.
func TestReadTimeline_EmptySection(t *testing.T) {
	dir := t.TempDir()
	ctxPath := filepath.Join(dir, "SESSION_CONTEXT.md")

	content := "---\nsession_id: test-001\nstatus: ACTIVE\n---\n\n## Timeline\n\n## Artifacts\n"
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write context file: %v", err)
	}

	entries, err := ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0 for empty timeline", len(entries))
	}
}

// TestReadTimeline_MissingSection verifies nil slice returned when no section.
func TestReadTimeline_MissingSection(t *testing.T) {
	dir := t.TempDir()
	ctxPath := filepath.Join(dir, "SESSION_CONTEXT.md")

	content := "---\nsession_id: test-001\nstatus: ACTIVE\n---\n\n## Artifacts\n- PRD: pending\n"
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write context file: %v", err)
	}

	entries, err := ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline should not error for missing section, got: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0 for missing section", len(entries))
	}
}

// --- truncateSummary tests ---

// TestTruncateSummary_ShortString verifies no truncation for strings within limit.
func TestTruncateSummary_ShortString(t *testing.T) {
	s := "short summary"
	got := truncateSummary(s, 80)
	if got != s {
		t.Errorf("truncateSummary(%q, 80) = %q, want %q", s, got, s)
	}
}

// TestTruncateSummary_ExactLimit verifies no truncation at exactly limit.
func TestTruncateSummary_ExactLimit(t *testing.T) {
	s := strings.Repeat("x", 80)
	got := truncateSummary(s, 80)
	if got != s {
		t.Errorf("expected no truncation at exact limit, got len=%d", len(got))
	}
}

// TestTruncateSummary_ExceedsLimit verifies truncation and "..." suffix.
func TestTruncateSummary_ExceedsLimit(t *testing.T) {
	s := strings.Repeat("x", 100)
	got := truncateSummary(s, 80)
	if len(got) != 80 {
		t.Errorf("truncated length = %d, want 80", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("truncated string should end with '...'")
	}
}

// --- EventTypeToCategory tests ---

// TestEventTypeToCategory_AllCategories verifies all categories are padded to 8 chars.
func TestEventTypeToCategory_AllCategories(t *testing.T) {
	cases := []struct {
		eventType clewcontract.EventType
		want      string // 8-char padded form
	}{
		{clewcontract.EventTypeSessionCreated, "SESSION "},
		{clewcontract.EventTypeSessionParked, "SESSION "},
		{clewcontract.EventTypeSessionResumed, "SESSION "},
		{clewcontract.EventTypeSessionWrapped, "SESSION "},
		{clewcontract.EventTypeSessionFrayed, "SESSION "},
		{clewcontract.EventTypePhaseTransitioned, "PHASE   "},
		{clewcontract.EventTypeAgentDelegated, "AGENT   "},
		{clewcontract.EventTypeAgentCompleted, "AGENT   "},
		{clewcontract.EventTypeCommitCreated, "COMMIT  "},
		{clewcontract.EventTypeDecisionRecorded, "DECISION"},
		{clewcontract.EventTypeCommandInvoked, "COMMAND "},
		{"unknown.type", "NOTE    "},
	}

	for _, tc := range cases {
		got := EventTypeToCategory(tc.eventType)
		if got != tc.want {
			t.Errorf("EventTypeToCategory(%q) = %q (len=%d), want %q (len=%d)",
				tc.eventType, got, len(got), tc.want, len(tc.want))
		}
		if len(got) != 8 {
			t.Errorf("EventTypeToCategory(%q) length = %d, want exactly 8", tc.eventType, len(got))
		}
	}
}

// --- ExtractSummary tests ---

// TestExtractSummary_PhaseTransition verifies phase transition summary format.
func TestExtractSummary_PhaseTransition(t *testing.T) {
	event := clewcontract.NewTypedPhaseTransitionedEvent("session-001", "requirements", "design")
	summary := ExtractSummary(event)
	if summary != "requirements -> design" {
		t.Errorf("ExtractSummary = %q, want %q", summary, "requirements -> design")
	}
}

// TestExtractSummary_CommitShortSHA verifies SHA is always 7 chars.
func TestExtractSummary_CommitShortSHA(t *testing.T) {
	event := clewcontract.NewTypedCommitCreatedEvent("abc123f890abcdef", "feat: thing")
	summary := ExtractSummary(event)
	if !strings.HasPrefix(summary, "abc123f: ") {
		t.Errorf("ExtractSummary = %q, want prefix 'abc123f: '", summary)
	}
}

// TestExtractSummary_DecisionRationaleShortened verifies rationale >30 chars is truncated.
func TestExtractSummary_DecisionRationaleShortened(t *testing.T) {
	longRationale := strings.Repeat("r", 35)
	event := clewcontract.NewTypedDecisionRecordedEvent("my decision", longRationale, nil)
	summary := ExtractSummary(event)

	// Should contain "my decision (" and the rationale truncated to 30 chars with "..."
	if !strings.Contains(summary, "my decision") {
		t.Errorf("ExtractSummary = %q, should contain decision text", summary)
	}
	if !strings.Contains(summary, "(") {
		t.Errorf("ExtractSummary = %q, should contain parenthesized rationale", summary)
	}
	// Inner truncation: rationale truncated to 27 + "..." = 30 chars
	if strings.Contains(summary, longRationale) {
		t.Errorf("ExtractSummary = %q, long rationale should be truncated", summary)
	}
}

// TestExtractSummary_EmptyDataFallback verifies degradation with empty Data field.
func TestExtractSummary_EmptyDataFallback(t *testing.T) {
	event := clewcontract.TypedEvent{
		Ts:     "2026-03-01T14:00:00.000Z",
		Type:   clewcontract.EventTypeSessionCreated,
		Source: clewcontract.SourceCLI,
		Data:   json.RawMessage("{}"),
	}
	summary := ExtractSummary(event)
	if summary != "created: (unknown)" {
		t.Errorf("ExtractSummary with empty data = %q, want %q", summary, "created: (unknown)")
	}
}

// --- Time parsing edge cases ---

// TestFormatTimelineEntry_RFC3339Timestamp verifies RFC3339 timestamp is also handled.
func TestFormatTimelineEntry_RFC3339Timestamp(t *testing.T) {
	event := clewcontract.NewTypedSessionCreatedEvent("session-001", "test", "PATCH", "")
	// Use RFC3339 format (with timezone offset) -- should be converted to UTC HH:MM.
	event.Ts = "2026-03-01T14:03:22+00:00"

	entry := FormatTimelineEntry(event)

	if !strings.HasPrefix(entry, "- 14:03 |") {
		t.Errorf("entry = %q, RFC3339 timestamp should produce 14:03", entry)
	}
}

// TestParseTimelineEntry_ValidEntry verifies a well-formed line is parsed.
func TestParseTimelineEntry_ValidEntry(t *testing.T) {
	line := "- 14:03 | SESSION  | created: Add dark mode (MODULE)"
	entry, ok := parseTimelineEntry(line)
	if !ok {
		t.Fatal("parseTimelineEntry should succeed for valid entry")
	}
	if entry.Category != "SESSION " {
		t.Errorf("Category = %q, want %q", entry.Category, "SESSION ")
	}
	if entry.Summary != "created: Add dark mode (MODULE)" {
		t.Errorf("Summary = %q, want %q", entry.Summary, "created: Add dark mode (MODULE)")
	}
	// Time should be 14:03
	if entry.Time.Hour() != 14 || entry.Time.Minute() != 3 {
		t.Errorf("Time = %v, want 14:03", entry.Time)
	}
}

// TestParseTimelineEntry_InvalidLine verifies non-entry lines are rejected.
func TestParseTimelineEntry_InvalidLine(t *testing.T) {
	invalidLines := []string{
		"",
		"## Timeline",
		"- PRD: pending",
		"Some body text",
		"- 14:03 missing separator",
	}
	for _, line := range invalidLines {
		_, ok := parseTimelineEntry(line)
		if ok {
			t.Errorf("parseTimelineEntry(%q) should return false for invalid line", line)
		}
	}
}

// TestInsertIntoTimelineSection_CreatesSection verifies section creation behavior.
func TestInsertIntoTimelineSection_CreatesSection(t *testing.T) {
	body := "---\nsession_id: test\nstatus: ACTIVE\n---\n\n# Session\n\n## Artifacts\n"
	entry := "- 14:03 | SESSION  | created: test (MODULE)"

	result, err := insertIntoTimelineSection(body, entry)
	if err != nil {
		t.Fatalf("insertIntoTimelineSection failed: %v", err)
	}

	if !strings.Contains(result, "## Timeline") {
		t.Error("## Timeline section should be created")
	}
	if !strings.Contains(result, entry) {
		t.Errorf("entry %q should be in the result", entry)
	}
}

// TestInsertIntoTimelineSection_InsertsBeforeNextSection verifies ordering constraint.
func TestInsertIntoTimelineSection_InsertsBeforeNextSection(t *testing.T) {
	body := "---\nsession_id: test\n---\n\n## Timeline\n- 14:03 | SESSION  | first entry\n\n## Artifacts\n"
	entry := "- 14:05 | AGENT    | delegated architect"

	result, err := insertIntoTimelineSection(body, entry)
	if err != nil {
		t.Fatalf("insertIntoTimelineSection failed: %v", err)
	}

	timelineIdx := strings.Index(result, "## Timeline")
	artifactsIdx := strings.Index(result, "## Artifacts")
	entryIdx := strings.Index(result, entry)

	if entryIdx < timelineIdx {
		t.Error("new entry should be after ## Timeline")
	}
	if entryIdx > artifactsIdx {
		t.Error("new entry should be before ## Artifacts")
	}
}

// TestAppendEntry_MultipleAppends verifies sequential appends preserve order.
func TestAppendEntry_MultipleAppends(t *testing.T) {
	dir := t.TempDir()
	ctxPath := filepath.Join(dir, "SESSION_CONTEXT.md")

	content := "---\nsession_id: test-001\nstatus: ACTIVE\n---\n\n## Timeline\n"
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write context file: %v", err)
	}

	entries := []string{
		"- 14:03 | SESSION  | created: test (MODULE)",
		"- 14:05 | AGENT    | delegated architect: design",
		"- 14:12 | COMMIT   | abc123f: feat: component",
	}

	for _, e := range entries {
		if err := AppendEntry(ctxPath, e); err != nil {
			t.Fatalf("AppendEntry(%q) failed: %v", e, err)
		}
	}

	result, err := ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("len(ReadTimeline) = %d, want 3", len(result))
	}

	// Verify order preserved.
	for i, entry := range result {
		if entry.Raw != entries[i] {
			t.Errorf("entries[%d].Raw = %q, want %q", i, entry.Raw, entries[i])
		}
	}
	_ = time.Now() // suppress unused import warning
}

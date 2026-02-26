package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTestEventsFile creates a temporary events.jsonl file with the given content.
func writeTestEventsFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test events file: %v", err)
	}
	return path
}

// --- v1 legacy format ---

func TestReadEvents_LegacyFormat(t *testing.T) {
	content := `{"timestamp":"2026-01-15T10:30:00Z","event":"STATE_TRANSITION","from":"NONE","to":"ACTIVE","metadata":{"reason":"init"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Event != "STATE_TRANSITION" {
		t.Errorf("Event = %q, want STATE_TRANSITION", e.Event)
	}
	if e.Timestamp != "2026-01-15T10:30:00Z" {
		t.Errorf("Timestamp = %q, want 2026-01-15T10:30:00Z", e.Timestamp)
	}
}

// --- v2 flat format ---

func TestReadEvents_FlatFormat(t *testing.T) {
	content := `{"ts":"2026-02-20T14:03:00.000Z","type":"session.created","summary":"Session created: s-002","meta":{"session_id":"s-002","initiative":"refactor","complexity":"PATCH","rite":"10x-dev"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Event != "session.created" {
		t.Errorf("Event = %q, want session.created", e.Event)
	}
	if e.Timestamp != "2026-02-20T14:03:00.000Z" {
		t.Errorf("Timestamp = %q, want 2026-02-20T14:03:00.000Z", e.Timestamp)
	}
}

// --- v3 typed format ---

func TestReadEvents_TypedFormat(t *testing.T) {
	content := `{"ts":"2026-03-01T14:30:00.000Z","type":"session.created","source":"cli","data":{"session_id":"s-003","initiative":"dark mode","complexity":"MODULE","rite":"ecosystem"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Event != "session.created" {
		t.Errorf("Event = %q, want session.created", e.Event)
	}
	if e.Timestamp != "2026-03-01T14:30:00.000Z" {
		t.Errorf("Timestamp = %q, want 2026-03-01T14:30:00.000Z", e.Timestamp)
	}

	// Source should be in Metadata
	if e.Metadata == nil {
		t.Fatal("Metadata should not be nil for typed events")
	}
	if e.Metadata["source"] != "cli" {
		t.Errorf("Metadata.source = %v, want cli", e.Metadata["source"])
	}

	// Data should be in Metadata
	dataMap, ok := e.Metadata["data"].(map[string]any)
	if !ok {
		t.Fatal("Metadata.data should be a map")
	}
	if dataMap["initiative"] != "dark mode" {
		t.Errorf("Metadata.data.initiative = %v, want dark mode", dataMap["initiative"])
	}
}

// --- Mixed format (the key integration test) ---

func TestReadEvents_MixedFormats(t *testing.T) {
	// This is the realistic scenario during migration: all three formats interleaved.
	// Based on Appendix A.4 of SESSION-1 spec.
	content := `{"timestamp":"2026-01-15T10:30:00Z","event":"SESSION_CREATED","metadata":{"session_id":"old-session"}}
{"ts":"2026-02-20T14:03:00.000Z","type":"session.created","summary":"Session created: s-002","meta":{"session_id":"s-002","initiative":"refactor","complexity":"PATCH","rite":"10x-dev"}}
{"ts":"2026-03-01T14:30:00.000Z","type":"session.created","source":"cli","data":{"session_id":"s-003","initiative":"dark mode","complexity":"MODULE","rite":"ecosystem"}}
{"ts":"2026-03-01T14:32:00.000Z","type":"agent.delegated","source":"hook","data":{"agent_name":"architect","agent_type":"specialist","task_id":"task-001"}}
{"ts":"2026-03-01T14:45:00.000Z","type":"commit.created","source":"hook","data":{"sha":"abc123f","message":"feat: add theme provider"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}

	// Line 1: v1 legacy
	if events[0].Event != "SESSION_CREATED" {
		t.Errorf("events[0].Event = %q, want SESSION_CREATED", events[0].Event)
	}

	// Line 2: v2 flat
	if events[1].Event != "session.created" {
		t.Errorf("events[1].Event = %q, want session.created", events[1].Event)
	}
	if events[1].Timestamp != "2026-02-20T14:03:00.000Z" {
		t.Errorf("events[1].Timestamp = %q, want 2026-02-20T14:03:00.000Z", events[1].Timestamp)
	}

	// Line 3: v3 typed (session.created)
	if events[2].Event != "session.created" {
		t.Errorf("events[2].Event = %q, want session.created", events[2].Event)
	}
	if events[2].Metadata["source"] != "cli" {
		t.Errorf("events[2].Metadata.source = %v, want cli", events[2].Metadata["source"])
	}

	// Line 4: v3 typed (agent.delegated)
	if events[3].Event != "agent.delegated" {
		t.Errorf("events[3].Event = %q, want agent.delegated", events[3].Event)
	}
	if events[3].Metadata["source"] != "hook" {
		t.Errorf("events[3].Metadata.source = %v, want hook", events[3].Metadata["source"])
	}
	dataMap, ok := events[3].Metadata["data"].(map[string]any)
	if !ok {
		t.Fatal("events[3].Metadata.data should be a map")
	}
	if dataMap["agent_name"] != "architect" {
		t.Errorf("events[3].data.agent_name = %v, want architect", dataMap["agent_name"])
	}

	// Line 5: v3 typed (commit.created)
	if events[4].Event != "commit.created" {
		t.Errorf("events[4].Event = %q, want commit.created", events[4].Event)
	}
	commitData, ok := events[4].Metadata["data"].(map[string]any)
	if !ok {
		t.Fatal("events[4].Metadata.data should be a map")
	}
	if commitData["sha"] != "abc123f" {
		t.Errorf("events[4].data.sha = %v, want abc123f", commitData["sha"])
	}
}

// --- v3 detection precedence: "data" wins over "meta" ---

func TestReadEvents_V3Precedence_DataAndMeta(t *testing.T) {
	// If a line has both "data" AND "meta" fields, it should be treated as v3.
	// Per SESSION-1 spec: "data" field presence = v3, regardless of "meta".
	content := `{"ts":"2026-03-01T14:30:00.000Z","type":"session.created","source":"cli","data":{"session_id":"s-003"},"meta":{"legacy_field":"value"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// Should be treated as v3: Metadata.source should be set
	if events[0].Metadata["source"] != "cli" {
		t.Errorf("event with both data+meta should be treated as v3; source = %v, want cli", events[0].Metadata["source"])
	}
}

// --- Malformed lines skipped ---

func TestReadEvents_MalformedLinesSkipped(t *testing.T) {
	content := `{"ts":"2026-03-01T14:30:00.000Z","type":"session.created","source":"cli","data":{"session_id":"s-001"}}
this is not json at all
{"ts":"2026-03-01T14:31:00.000Z","type":"session.parked","source":"cli","data":{"session_id":"s-001","reason":"lunch"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	// Only 2 valid events; malformed line is skipped.
	if len(events) != 2 {
		t.Fatalf("expected 2 events (malformed line skipped), got %d", len(events))
	}
}

// --- Missing file returns empty slice (not error) ---

func TestReadEvents_MissingFile(t *testing.T) {
	events, err := ReadEvents("/nonexistent/path/events.jsonl")
	if err != nil {
		t.Fatalf("ReadEvents should return empty slice for missing file, got error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events for missing file, got %d", len(events))
	}
}

// --- Empty file returns empty slice ---

func TestReadEvents_EmptyFile(t *testing.T) {
	path := writeTestEventsFile(t, "")

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events for empty file, got %d", len(events))
	}
}

// --- v3 typed event: data field contents accessible ---

func TestReadEvents_TypedEvent_DataAccessible(t *testing.T) {
	content := `{"ts":"2026-03-01T14:30:00.000Z","type":"agent.completed","source":"hook","data":{"agent_name":"integration-engineer","outcome":"success","duration_ms":15000,"artifacts":["/path/to/file.go"]}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Event != "agent.completed" {
		t.Errorf("Event = %q, want agent.completed", e.Event)
	}

	dataMap, ok := e.Metadata["data"].(map[string]any)
	if !ok {
		t.Fatal("Metadata.data should be a map")
	}
	if dataMap["agent_name"] != "integration-engineer" {
		t.Errorf("data.agent_name = %v, want integration-engineer", dataMap["agent_name"])
	}
	if dataMap["outcome"] != "success" {
		t.Errorf("data.outcome = %v, want success", dataMap["outcome"])
	}
	// duration_ms is a number in JSON; float64 in Go
	if dataMap["duration_ms"] != float64(15000) {
		t.Errorf("data.duration_ms = %v, want 15000", dataMap["duration_ms"])
	}
}

// --- FilterEvents still works after ReadEvents changes ---

func TestFilterEvents_WithTypedEvents(t *testing.T) {
	content := `{"ts":"2026-03-01T14:30:00.000Z","type":"session.created","source":"cli","data":{"session_id":"s-001","initiative":"test","complexity":"PATCH"}}
{"ts":"2026-03-01T14:32:00.000Z","type":"agent.delegated","source":"hook","data":{"agent_name":"architect"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	// Filter to session.created only; time.Time{} = no time filter.
	filtered := FilterEvents(events, "session.created", time.Time{})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered event, got %d", len(filtered))
	}
	if filtered[0].Event != "session.created" {
		t.Errorf("filtered event = %q, want session.created", filtered[0].Event)
	}
}

func TestFilterEvents_TypedTimestampFormat(t *testing.T) {
	// FilterEvents must correctly parse the TypedEvent timestamp format (same as v2).
	// The filter should correctly include events after the cutoff time.
	content := `{"ts":"2026-03-01T14:30:00.000Z","type":"session.created","source":"cli","data":{"session_id":"s-001","initiative":"test","complexity":"PATCH"}}
{"ts":"2026-03-01T14:31:00.000Z","type":"session.parked","source":"cli","data":{"session_id":"s-001","reason":"break"}}
`
	path := writeTestEventsFile(t, content)

	events, err := ReadEvents(path)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	// Filter to events after 14:30:30 -- only the parked event should be included.
	cutoff, err := time.Parse("2006-01-02T15:04:05.000Z", "2026-03-01T14:30:30.000Z")
	if err != nil {
		t.Fatalf("failed to parse cutoff time: %v", err)
	}

	filtered := FilterEvents(events, "", cutoff)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 event after cutoff, got %d", len(filtered))
	}
	if filtered[0].Event != "session.parked" {
		t.Errorf("filtered event = %q, want session.parked", filtered[0].Event)
	}
}

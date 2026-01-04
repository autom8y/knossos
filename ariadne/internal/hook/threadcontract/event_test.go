package threadcontract

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventType_Constants(t *testing.T) {
	// Verify event type constants match expected schema values
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventTypeToolCall, "tool_call"},
		{EventTypeFileChange, "file_change"},
		{EventTypeCommand, "command"},
		{EventTypeDecision, "decision"},
		{EventTypeContextSwitch, "context_switch"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("EventType %v = %q, want %q", tt.eventType, tt.eventType, tt.expected)
		}
	}
}

func TestEvent_JSONMarshaling(t *testing.T) {
	event := Event{
		Timestamp: "2024-01-04T10:23:45.123Z",
		Type:      EventTypeToolCall,
		Tool:      "Edit",
		Path:      "/abs/path/file.go",
		Summary:   "Tool: Edit on /abs/path/file.go",
		Meta: map[string]interface{}{
			"lines_changed": 42,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Verify JSON structure
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check required fields
	if raw["ts"] != "2024-01-04T10:23:45.123Z" {
		t.Errorf("ts = %v, want 2024-01-04T10:23:45.123Z", raw["ts"])
	}
	if raw["type"] != "tool_call" {
		t.Errorf("type = %v, want tool_call", raw["type"])
	}
	if raw["tool"] != "Edit" {
		t.Errorf("tool = %v, want Edit", raw["tool"])
	}
	if raw["path"] != "/abs/path/file.go" {
		t.Errorf("path = %v, want /abs/path/file.go", raw["path"])
	}
	if raw["summary"] != "Tool: Edit on /abs/path/file.go" {
		t.Errorf("summary = %v, want 'Tool: Edit on /abs/path/file.go'", raw["summary"])
	}

	// Check meta
	meta, ok := raw["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if meta["lines_changed"] != float64(42) { // JSON numbers are float64
		t.Errorf("meta.lines_changed = %v, want 42", meta["lines_changed"])
	}
}

func TestEvent_JSONUnmarshaling(t *testing.T) {
	jsonStr := `{
		"ts": "2024-01-04T10:23:45.123Z",
		"type": "file_change",
		"path": "/some/file.txt",
		"summary": "Changed /some/file.txt",
		"meta": {"lines_changed": 10}
	}`

	var event Event
	if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if event.Timestamp != "2024-01-04T10:23:45.123Z" {
		t.Errorf("Timestamp = %v, want 2024-01-04T10:23:45.123Z", event.Timestamp)
	}
	if event.Type != EventTypeFileChange {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeFileChange)
	}
	if event.Path != "/some/file.txt" {
		t.Errorf("Path = %v, want /some/file.txt", event.Path)
	}
	if event.Summary != "Changed /some/file.txt" {
		t.Errorf("Summary = %v, want 'Changed /some/file.txt'", event.Summary)
	}
	if event.Meta["lines_changed"] != float64(10) {
		t.Errorf("Meta.lines_changed = %v, want 10", event.Meta["lines_changed"])
	}
}

func TestEvent_OmitEmpty(t *testing.T) {
	// Event with minimal fields
	event := Event{
		Timestamp: "2024-01-04T10:23:45.123Z",
		Type:      EventTypeDecision,
		Summary:   "Made a decision",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// These should be omitted
	if _, exists := raw["tool"]; exists {
		t.Error("tool should be omitted when empty")
	}
	if _, exists := raw["path"]; exists {
		t.Error("path should be omitted when empty")
	}
	if _, exists := raw["meta"]; exists {
		t.Error("meta should be omitted when nil")
	}
}

func TestNewToolCallEvent(t *testing.T) {
	meta := map[string]interface{}{"command": "ls -la"}
	event := NewToolCallEvent("Bash", "/some/path", meta)

	if event.Type != EventTypeToolCall {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeToolCall)
	}
	if event.Tool != "Bash" {
		t.Errorf("Tool = %v, want Bash", event.Tool)
	}
	if event.Path != "/some/path" {
		t.Errorf("Path = %v, want /some/path", event.Path)
	}
	if event.Summary != "Tool: Bash on /some/path" {
		t.Errorf("Summary = %v, want 'Tool: Bash on /some/path'", event.Summary)
	}
	if event.Meta["command"] != "ls -la" {
		t.Errorf("Meta.command = %v, want 'ls -la'", event.Meta["command"])
	}
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestNewToolCallEvent_NoPath(t *testing.T) {
	event := NewToolCallEvent("Bash", "", nil)

	if event.Summary != "Tool: Bash" {
		t.Errorf("Summary = %v, want 'Tool: Bash'", event.Summary)
	}
	if event.Path != "" {
		t.Errorf("Path = %v, want empty", event.Path)
	}
}

func TestNewFileChangeEvent(t *testing.T) {
	event := NewFileChangeEvent("/path/to/file.go", 42)

	if event.Type != EventTypeFileChange {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeFileChange)
	}
	if event.Path != "/path/to/file.go" {
		t.Errorf("Path = %v, want /path/to/file.go", event.Path)
	}
	if event.Summary != "Changed /path/to/file.go" {
		t.Errorf("Summary = %v, want 'Changed /path/to/file.go'", event.Summary)
	}
	if event.Meta["lines_changed"] != 42 {
		t.Errorf("Meta.lines_changed = %v, want 42", event.Meta["lines_changed"])
	}
}

func TestNewCommandEvent(t *testing.T) {
	event := NewCommandEvent("go test ./...", 0, 1500)

	if event.Type != EventTypeCommand {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeCommand)
	}
	if event.Summary != "go test ./..." {
		t.Errorf("Summary = %v, want 'go test ./...'", event.Summary)
	}
	if event.Meta["exit_code"] != 0 {
		t.Errorf("Meta.exit_code = %v, want 0", event.Meta["exit_code"])
	}
	if event.Meta["duration_ms"] != int64(1500) {
		t.Errorf("Meta.duration_ms = %v, want 1500", event.Meta["duration_ms"])
	}
}

func TestNewCommandEvent_LongCommand(t *testing.T) {
	longCommand := "a_very_long_command_that_exceeds_eighty_characters_and_should_be_truncated_for_readability_purposes_in_the_summary"
	event := NewCommandEvent(longCommand, 0, 100)

	if len(event.Summary) > 80 {
		t.Errorf("Summary length = %d, want <= 80", len(event.Summary))
	}
	if event.Summary[len(event.Summary)-3:] != "..." {
		t.Error("Summary should end with '...'")
	}
}

func TestNewDecisionEvent(t *testing.T) {
	meta := map[string]interface{}{"reason": "performance"}
	event := NewDecisionEvent("Selected approach A over B", meta)

	if event.Type != EventTypeDecision {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeDecision)
	}
	if event.Summary != "Selected approach A over B" {
		t.Errorf("Summary = %v, want 'Selected approach A over B'", event.Summary)
	}
	if event.Meta["reason"] != "performance" {
		t.Errorf("Meta.reason = %v, want 'performance'", event.Meta["reason"])
	}
}

func TestNewContextSwitchEvent(t *testing.T) {
	meta := map[string]interface{}{"from_file": "old.go"}
	event := NewContextSwitchEvent("Switched to new file", "/new/file.go", meta)

	if event.Type != EventTypeContextSwitch {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeContextSwitch)
	}
	if event.Path != "/new/file.go" {
		t.Errorf("Path = %v, want /new/file.go", event.Path)
	}
	if event.Summary != "Switched to new file" {
		t.Errorf("Summary = %v, want 'Switched to new file'", event.Summary)
	}
	if event.Meta["from_file"] != "old.go" {
		t.Errorf("Meta.from_file = %v, want 'old.go'", event.Meta["from_file"])
	}
}

func TestTimestamp_Format(t *testing.T) {
	ts := timestamp()

	// Should be parseable
	_, err := time.Parse("2006-01-02T15:04:05.000Z", ts)
	if err != nil {
		t.Errorf("Failed to parse timestamp %q: %v", ts, err)
	}

	// Should end with Z (UTC)
	if ts[len(ts)-1] != 'Z' {
		t.Errorf("Timestamp should end with Z, got %q", ts)
	}
}

func TestStamp_JSONMarshaling(t *testing.T) {
	stamp := Stamp{
		Ts:        time.Date(2024, 1, 4, 10, 23, 45, 123000000, time.UTC),
		Decision:  "Use PostgreSQL over MongoDB",
		Rationale: "Better ACID compliance for financial data",
		Rejected:  []string{"MongoDB", "CockroachDB"},
		Context: map[string]any{
			"team": "backend",
			"priority": "high",
		},
	}

	data, err := json.Marshal(stamp)
	if err != nil {
		t.Fatalf("Failed to marshal stamp: %v", err)
	}

	// Verify JSON structure
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if raw["decision"] != "Use PostgreSQL over MongoDB" {
		t.Errorf("decision = %v, want 'Use PostgreSQL over MongoDB'", raw["decision"])
	}
	if raw["rationale"] != "Better ACID compliance for financial data" {
		t.Errorf("rationale = %v, want 'Better ACID compliance for financial data'", raw["rationale"])
	}

	// Check rejected array
	rejected, ok := raw["rejected"].([]interface{})
	if !ok {
		t.Fatal("rejected is not a slice")
	}
	if len(rejected) != 2 {
		t.Errorf("rejected length = %d, want 2", len(rejected))
	}

	// Check context
	ctx, ok := raw["context"].(map[string]interface{})
	if !ok {
		t.Fatal("context is not a map")
	}
	if ctx["team"] != "backend" {
		t.Errorf("context.team = %v, want 'backend'", ctx["team"])
	}
}

func TestStamp_OmitEmpty(t *testing.T) {
	// Stamp with minimal fields
	stamp := Stamp{
		Ts:        time.Date(2024, 1, 4, 10, 23, 45, 123000000, time.UTC),
		Decision:  "Keep current approach",
		Rationale: "Works well enough",
	}

	data, err := json.Marshal(stamp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// These should be omitted
	if _, exists := raw["rejected"]; exists {
		t.Error("rejected should be omitted when empty")
	}
	if _, exists := raw["context"]; exists {
		t.Error("context should be omitted when nil")
	}
}

func TestNewStamp(t *testing.T) {
	rejected := []string{"Option A", "Option B"}
	context := map[string]any{"source": "review"}

	stamp := NewStamp("Chose Option C", "Most maintainable", rejected, context)

	if stamp.Decision != "Chose Option C" {
		t.Errorf("Decision = %v, want 'Chose Option C'", stamp.Decision)
	}
	if stamp.Rationale != "Most maintainable" {
		t.Errorf("Rationale = %v, want 'Most maintainable'", stamp.Rationale)
	}
	if len(stamp.Rejected) != 2 {
		t.Errorf("Rejected length = %d, want 2", len(stamp.Rejected))
	}
	if stamp.Ts.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestStamp_ToEvent(t *testing.T) {
	stamp := Stamp{
		Ts:        time.Date(2024, 1, 4, 10, 23, 45, 123000000, time.UTC),
		Decision:  "Use gRPC over REST",
		Rationale: "Better performance for internal services",
		Rejected:  []string{"REST", "GraphQL"},
		Context: map[string]any{
			"service": "order-service",
		},
	}

	event := stamp.ToEvent()

	// Check event type
	if event.Type != EventTypeDecision {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeDecision)
	}

	// Check timestamp format
	if event.Timestamp != "2024-01-04T10:23:45.123Z" {
		t.Errorf("Timestamp = %v, want '2024-01-04T10:23:45.123Z'", event.Timestamp)
	}

	// Check summary (should be the decision)
	if event.Summary != "Use gRPC over REST" {
		t.Errorf("Summary = %v, want 'Use gRPC over REST'", event.Summary)
	}

	// Check meta has rationale
	if event.Meta["rationale"] != "Better performance for internal services" {
		t.Errorf("Meta.rationale = %v, want 'Better performance for internal services'", event.Meta["rationale"])
	}

	// Check meta has rejected
	rejected, ok := event.Meta["rejected"].([]string)
	if !ok {
		t.Fatal("Meta.rejected is not a []string")
	}
	if len(rejected) != 2 {
		t.Errorf("Meta.rejected length = %d, want 2", len(rejected))
	}

	// Check context is merged
	if event.Meta["service"] != "order-service" {
		t.Errorf("Meta.service = %v, want 'order-service'", event.Meta["service"])
	}
}

func TestStamp_ToEvent_MinimalFields(t *testing.T) {
	stamp := Stamp{
		Ts:        time.Date(2024, 1, 4, 10, 23, 45, 0, time.UTC),
		Decision:  "Proceed with current plan",
		Rationale: "No blockers identified",
	}

	event := stamp.ToEvent()

	if event.Type != EventTypeDecision {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeDecision)
	}

	if event.Summary != "Proceed with current plan" {
		t.Errorf("Summary = %v, want 'Proceed with current plan'", event.Summary)
	}

	if event.Meta["rationale"] != "No blockers identified" {
		t.Errorf("Meta.rationale = %v, want 'No blockers identified'", event.Meta["rationale"])
	}

	// rejected should not be in meta when empty
	if _, exists := event.Meta["rejected"]; exists {
		t.Error("Meta.rejected should not exist when stamp.Rejected is empty")
	}
}

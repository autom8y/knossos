package clewcontract

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
		{EventTypeSailsGenerated, "sails_generated"},
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

func TestEventTypeSailsGenerated_Constant(t *testing.T) {
	if string(EventTypeSailsGenerated) != "sails_generated" {
		t.Errorf("EventTypeSailsGenerated = %q, want %q", EventTypeSailsGenerated, "sails_generated")
	}
}

func TestNewSailsGeneratedEvent(t *testing.T) {
	data := SailsGeneratedData{
		Color:        "WHITE",
		ComputedBase: "WHITE",
		Reasons:      []string{"all required proofs present and passing"},
		FilePath:     ".claude/sessions/session-20260105-143000-abc12345/WHITE_SAILS.yaml",
	}

	event := NewSailsGeneratedEvent("session-20260105-143000-abc12345", data)

	// Check event type
	if event.Type != EventTypeSailsGenerated {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeSailsGenerated)
	}

	// Check path
	if event.Path != data.FilePath {
		t.Errorf("Path = %v, want %v", event.Path, data.FilePath)
	}

	// Check summary (same color as base)
	expectedSummary := "Generated WHITE_SAILS: WHITE"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %v, want %v", event.Summary, expectedSummary)
	}

	// Check timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Check meta fields
	if event.Meta["session_id"] != "session-20260105-143000-abc12345" {
		t.Errorf("Meta.session_id = %v, want 'session-20260105-143000-abc12345'", event.Meta["session_id"])
	}
	if event.Meta["color"] != "WHITE" {
		t.Errorf("Meta.color = %v, want 'WHITE'", event.Meta["color"])
	}
	if event.Meta["computed_base"] != "WHITE" {
		t.Errorf("Meta.computed_base = %v, want 'WHITE'", event.Meta["computed_base"])
	}
	if event.Meta["file_path"] != data.FilePath {
		t.Errorf("Meta.file_path = %v, want %v", event.Meta["file_path"], data.FilePath)
	}

	// Check reasons
	reasons, ok := event.Meta["reasons"].([]string)
	if !ok {
		t.Fatal("Meta.reasons is not a []string")
	}
	if len(reasons) != 1 {
		t.Errorf("Meta.reasons length = %d, want 1", len(reasons))
	}
	if reasons[0] != "all required proofs present and passing" {
		t.Errorf("Meta.reasons[0] = %v, want 'all required proofs present and passing'", reasons[0])
	}
}

func TestNewSailsGeneratedEvent_WithModifier(t *testing.T) {
	// Test when color differs from computed base (modifier applied)
	data := SailsGeneratedData{
		Color:        "GRAY",
		ComputedBase: "WHITE",
		Reasons: []string{
			"all required proofs present and passing",
			"modifier DOWNGRADE_TO_GRAY applied: Changed retry logic in payment flow",
		},
		FilePath: ".claude/sessions/session-20260105-143000-def67890/WHITE_SAILS.yaml",
	}

	event := NewSailsGeneratedEvent("session-20260105-143000-def67890", data)

	// Check summary includes both colors when they differ
	expectedSummary := "Generated WHITE_SAILS: GRAY (base: WHITE)"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %v, want %v", event.Summary, expectedSummary)
	}

	if event.Meta["color"] != "GRAY" {
		t.Errorf("Meta.color = %v, want 'GRAY'", event.Meta["color"])
	}
	if event.Meta["computed_base"] != "WHITE" {
		t.Errorf("Meta.computed_base = %v, want 'WHITE'", event.Meta["computed_base"])
	}

	reasons, ok := event.Meta["reasons"].([]string)
	if !ok {
		t.Fatal("Meta.reasons is not a []string")
	}
	if len(reasons) != 2 {
		t.Errorf("Meta.reasons length = %d, want 2", len(reasons))
	}
}

func TestNewSailsGeneratedEvent_Gray(t *testing.T) {
	// Test GRAY result (open questions)
	data := SailsGeneratedData{
		Color:        "GRAY",
		ComputedBase: "GRAY",
		Reasons:      []string{"open questions present: gray ceiling applied"},
		FilePath:     ".claude/sessions/session-20260105-160000-ghi11111/WHITE_SAILS.yaml",
	}

	event := NewSailsGeneratedEvent("session-20260105-160000-ghi11111", data)

	// Summary should not include "(base: X)" when colors match
	expectedSummary := "Generated WHITE_SAILS: GRAY"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %v, want %v", event.Summary, expectedSummary)
	}

	if event.Meta["color"] != "GRAY" {
		t.Errorf("Meta.color = %v, want 'GRAY'", event.Meta["color"])
	}
	if event.Meta["computed_base"] != "GRAY" {
		t.Errorf("Meta.computed_base = %v, want 'GRAY'", event.Meta["computed_base"])
	}
}

func TestNewSailsGeneratedEvent_Black(t *testing.T) {
	// Test BLACK result (test failure)
	data := SailsGeneratedData{
		Color:        "BLACK",
		ComputedBase: "BLACK",
		Reasons:      []string{"proof 'tests' has status FAIL"},
		FilePath:     ".claude/sessions/session-20260105-170000-jkl22222/WHITE_SAILS.yaml",
	}

	event := NewSailsGeneratedEvent("session-20260105-170000-jkl22222", data)

	expectedSummary := "Generated WHITE_SAILS: BLACK"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %v, want %v", event.Summary, expectedSummary)
	}

	if event.Meta["color"] != "BLACK" {
		t.Errorf("Meta.color = %v, want 'BLACK'", event.Meta["color"])
	}
}

func TestNewSailsGeneratedEvent_JSONMarshaling(t *testing.T) {
	data := SailsGeneratedData{
		Color:        "WHITE",
		ComputedBase: "WHITE",
		Reasons:      []string{"all required proofs present and passing"},
		FilePath:     ".claude/sessions/session-20260105-143000-abc12345/WHITE_SAILS.yaml",
	}

	event := NewSailsGeneratedEvent("session-20260105-143000-abc12345", data)

	// Marshal to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Unmarshal to verify structure
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check type in JSON
	if raw["type"] != "sails_generated" {
		t.Errorf("type = %v, want 'sails_generated'", raw["type"])
	}

	// Check meta is present
	meta, ok := raw["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if meta["color"] != "WHITE" {
		t.Errorf("meta.color = %v, want 'WHITE'", meta["color"])
	}
	if meta["computed_base"] != "WHITE" {
		t.Errorf("meta.computed_base = %v, want 'WHITE'", meta["computed_base"])
	}

	// Check reasons array in JSON
	reasons, ok := meta["reasons"].([]interface{})
	if !ok {
		t.Fatal("meta.reasons is not a slice")
	}
	if len(reasons) != 1 {
		t.Errorf("meta.reasons length = %d, want 1", len(reasons))
	}
}

func TestSailsGeneratedData_EmptyReasons(t *testing.T) {
	// Test with empty reasons array
	data := SailsGeneratedData{
		Color:        "WHITE",
		ComputedBase: "WHITE",
		Reasons:      []string{},
		FilePath:     ".claude/sessions/session-20260105-143000-xyz99999/WHITE_SAILS.yaml",
	}

	event := NewSailsGeneratedEvent("session-20260105-143000-xyz99999", data)

	reasons, ok := event.Meta["reasons"].([]string)
	if !ok {
		t.Fatal("Meta.reasons is not a []string")
	}
	if len(reasons) != 0 {
		t.Errorf("Meta.reasons length = %d, want 0", len(reasons))
	}
}

func TestNewSailsGeneratedEvent_WithEvidencePaths(t *testing.T) {
	data := SailsGeneratedData{
		Color:        "WHITE",
		ComputedBase: "WHITE",
		Reasons:      []string{"all required proofs present and passing"},
		FilePath:     ".claude/sessions/session-20260105-143000-abc12345/WHITE_SAILS.yaml",
		EvidencePaths: &EvidencePaths{
			Tests: ".claude/sessions/session-20260105-143000-abc12345/test-output.log",
			Build: ".claude/sessions/session-20260105-143000-abc12345/build-output.log",
			Lint:  ".claude/sessions/session-20260105-143000-abc12345/lint-output.log",
		},
	}

	event := NewSailsGeneratedEvent("session-20260105-143000-abc12345", data)

	// Check type
	if event.Type != EventTypeSailsGenerated {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeSailsGenerated)
	}

	// Check evidence_paths in meta
	evidencePaths, ok := event.Meta["evidence_paths"].(map[string]string)
	if !ok {
		t.Fatal("Meta.evidence_paths is not a map[string]string")
	}

	if evidencePaths["tests"] != ".claude/sessions/session-20260105-143000-abc12345/test-output.log" {
		t.Errorf("evidence_paths.tests = %v, want test-output.log path", evidencePaths["tests"])
	}
	if evidencePaths["build"] != ".claude/sessions/session-20260105-143000-abc12345/build-output.log" {
		t.Errorf("evidence_paths.build = %v, want build-output.log path", evidencePaths["build"])
	}
	if evidencePaths["lint"] != ".claude/sessions/session-20260105-143000-abc12345/lint-output.log" {
		t.Errorf("evidence_paths.lint = %v, want lint-output.log path", evidencePaths["lint"])
	}
}

func TestNewSailsGeneratedEvent_WithAllEvidencePaths(t *testing.T) {
	data := SailsGeneratedData{
		Color:        "WHITE",
		ComputedBase: "WHITE",
		Reasons:      []string{"all proofs passing"},
		FilePath:     ".claude/sessions/session-test/WHITE_SAILS.yaml",
		EvidencePaths: &EvidencePaths{
			Tests:       ".claude/sessions/session-test/test-output.log",
			Build:       ".claude/sessions/session-test/build-output.log",
			Lint:        ".claude/sessions/session-test/lint-output.log",
			Adversarial: ".claude/sessions/session-test/adversarial-output.log",
			Integration: ".claude/sessions/session-test/integration-output.log",
		},
	}

	event := NewSailsGeneratedEvent("session-test", data)

	evidencePaths, ok := event.Meta["evidence_paths"].(map[string]string)
	if !ok {
		t.Fatal("Meta.evidence_paths is not a map[string]string")
	}

	// Check all five evidence paths
	if evidencePaths["tests"] == "" {
		t.Error("evidence_paths.tests should not be empty")
	}
	if evidencePaths["build"] == "" {
		t.Error("evidence_paths.build should not be empty")
	}
	if evidencePaths["lint"] == "" {
		t.Error("evidence_paths.lint should not be empty")
	}
	if evidencePaths["adversarial"] == "" {
		t.Error("evidence_paths.adversarial should not be empty")
	}
	if evidencePaths["integration"] == "" {
		t.Error("evidence_paths.integration should not be empty")
	}
}

func TestNewSailsGeneratedEvent_NoEvidencePaths(t *testing.T) {
	// Test without evidence paths (nil)
	data := SailsGeneratedData{
		Color:         "GRAY",
		ComputedBase:  "GRAY",
		Reasons:       []string{"missing proofs"},
		FilePath:      ".claude/sessions/session-test/WHITE_SAILS.yaml",
		EvidencePaths: nil,
	}

	event := NewSailsGeneratedEvent("session-test", data)

	// evidence_paths should not be in meta
	if _, exists := event.Meta["evidence_paths"]; exists {
		t.Error("Meta should not contain evidence_paths when EvidencePaths is nil")
	}
}

func TestNewSailsGeneratedEvent_EmptyEvidencePaths(t *testing.T) {
	// Test with empty evidence paths struct (no paths set)
	data := SailsGeneratedData{
		Color:         "GRAY",
		ComputedBase:  "GRAY",
		Reasons:       []string{"proofs not found"},
		FilePath:      ".claude/sessions/session-test/WHITE_SAILS.yaml",
		EvidencePaths: &EvidencePaths{},
	}

	event := NewSailsGeneratedEvent("session-test", data)

	// evidence_paths should not be in meta when all paths are empty
	if _, exists := event.Meta["evidence_paths"]; exists {
		t.Error("Meta should not contain evidence_paths when all paths are empty strings")
	}
}

func TestNewSailsGeneratedEvent_PartialEvidencePaths(t *testing.T) {
	// Test with only some evidence paths set
	data := SailsGeneratedData{
		Color:        "WHITE",
		ComputedBase: "WHITE",
		Reasons:      []string{"required proofs passing"},
		FilePath:     ".claude/sessions/session-test/WHITE_SAILS.yaml",
		EvidencePaths: &EvidencePaths{
			Tests: ".claude/sessions/session-test/test-output.log",
			Build: ".claude/sessions/session-test/build-output.log",
			// Lint, Adversarial, Integration are empty
		},
	}

	event := NewSailsGeneratedEvent("session-test", data)

	evidencePaths, ok := event.Meta["evidence_paths"].(map[string]string)
	if !ok {
		t.Fatal("Meta.evidence_paths is not a map[string]string")
	}

	// Should have tests and build
	if evidencePaths["tests"] == "" {
		t.Error("evidence_paths.tests should not be empty")
	}
	if evidencePaths["build"] == "" {
		t.Error("evidence_paths.build should not be empty")
	}

	// Should NOT have lint, adversarial, integration (empty strings omitted)
	if _, exists := evidencePaths["lint"]; exists {
		t.Error("evidence_paths should not contain lint when it's empty")
	}
	if _, exists := evidencePaths["adversarial"]; exists {
		t.Error("evidence_paths should not contain adversarial when it's empty")
	}
	if _, exists := evidencePaths["integration"]; exists {
		t.Error("evidence_paths should not contain integration when it's empty")
	}
}

func TestArtifactType_WhiteSails(t *testing.T) {
	// Test that ArtifactTypeWhiteSails is defined correctly
	if ArtifactTypeWhiteSails != "white_sails" {
		t.Errorf("ArtifactTypeWhiteSails = %v, want 'white_sails'", ArtifactTypeWhiteSails)
	}
}

func TestArtifactType_Count(t *testing.T) {
	// Test that we have 6 artifact types total
	artifactTypes := []ArtifactType{
		ArtifactTypePRD,
		ArtifactTypeTDD,
		ArtifactTypeADR,
		ArtifactTypeTestPlan,
		ArtifactTypeCode,
		ArtifactTypeWhiteSails,
	}

	if len(artifactTypes) != 6 {
		t.Errorf("Expected 6 artifact types, got %d", len(artifactTypes))
	}

	// Verify each type has correct value
	expected := map[ArtifactType]string{
		ArtifactTypePRD:        "prd",
		ArtifactTypeTDD:        "tdd",
		ArtifactTypeADR:        "adr",
		ArtifactTypeTestPlan:   "test_plan",
		ArtifactTypeCode:       "code",
		ArtifactTypeWhiteSails: "white_sails",
	}

	for artType, val := range expected {
		if string(artType) != val {
			t.Errorf("%v = %v, want %v", artType, string(artType), val)
		}
	}
}

func TestEventTypeTaskStart_Constant(t *testing.T) {
	if string(EventTypeTaskStart) != "task_start" {
		t.Errorf("EventTypeTaskStart = %q, want %q", EventTypeTaskStart, "task_start")
	}
}

func TestEventTypeTaskEnd_Constant(t *testing.T) {
	if string(EventTypeTaskEnd) != "task_end" {
		t.Errorf("EventTypeTaskEnd = %q, want %q", EventTypeTaskEnd, "task_end")
	}
}

func TestNewTaskStartEvent(t *testing.T) {
	event := NewTaskStartEvent("task-001", "architect", "design", "session-20260106-123456-abc")

	// Check event type
	if event.Type != EventTypeTaskStart {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeTaskStart)
	}

	// Check summary format
	expectedSummary := "Task started: task-001 by architect in design phase"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %v, want %v", event.Summary, expectedSummary)
	}

	// Check timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Check meta fields
	if event.Meta["task_id"] != "task-001" {
		t.Errorf("Meta.task_id = %v, want 'task-001'", event.Meta["task_id"])
	}
	if event.Meta["agent"] != "architect" {
		t.Errorf("Meta.agent = %v, want 'architect'", event.Meta["agent"])
	}
	if event.Meta["phase"] != "design" {
		t.Errorf("Meta.phase = %v, want 'design'", event.Meta["phase"])
	}
	if event.Meta["session_id"] != "session-20260106-123456-abc" {
		t.Errorf("Meta.session_id = %v, want 'session-20260106-123456-abc'", event.Meta["session_id"])
	}
}

func TestNewTaskEndEvent(t *testing.T) {
	artifacts := []string{
		"/Users/tomtenuta/Code/knossos/docs/requirements/PRD-test.md",
		"/Users/tomtenuta/Code/knossos/docs/design/TDD-test.md",
	}
	event := NewTaskEndEvent("task-001", "architect", "success", "session-20260106-123456-abc", 15000, artifacts)

	// Check event type
	if event.Type != EventTypeTaskEnd {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeTaskEnd)
	}

	// Check summary format
	expectedSummary := "Task ended: task-001 by architect - success (15000ms)"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %v, want %v", event.Summary, expectedSummary)
	}

	// Check timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Check meta fields
	if event.Meta["task_id"] != "task-001" {
		t.Errorf("Meta.task_id = %v, want 'task-001'", event.Meta["task_id"])
	}
	if event.Meta["agent"] != "architect" {
		t.Errorf("Meta.agent = %v, want 'architect'", event.Meta["agent"])
	}
	if event.Meta["outcome"] != "success" {
		t.Errorf("Meta.outcome = %v, want 'success'", event.Meta["outcome"])
	}
	if event.Meta["session_id"] != "session-20260106-123456-abc" {
		t.Errorf("Meta.session_id = %v, want 'session-20260106-123456-abc'", event.Meta["session_id"])
	}
	if event.Meta["duration_ms"] != int64(15000) {
		t.Errorf("Meta.duration_ms = %v, want 15000", event.Meta["duration_ms"])
	}

	// Check artifacts
	artifactsFromMeta, ok := event.Meta["artifacts"].([]string)
	if !ok {
		t.Fatal("Meta.artifacts is not a []string")
	}
	if len(artifactsFromMeta) != 2 {
		t.Errorf("Meta.artifacts length = %d, want 2", len(artifactsFromMeta))
	}
	if artifactsFromMeta[0] != artifacts[0] {
		t.Errorf("Meta.artifacts[0] = %v, want %v", artifactsFromMeta[0], artifacts[0])
	}
}

func TestNewTaskEndEvent_EmptyArtifacts(t *testing.T) {
	event := NewTaskEndEvent("task-002", "qa-adversary", "blocked", "session-xyz", 5000, []string{})

	// Check summary includes blocked status
	if event.Summary != "Task ended: task-002 by qa-adversary - blocked (5000ms)" {
		t.Errorf("Summary = %v, want blocked status", event.Summary)
	}

	// Check empty artifacts array
	artifacts, ok := event.Meta["artifacts"].([]string)
	if !ok {
		t.Fatal("Meta.artifacts is not a []string")
	}
	if len(artifacts) != 0 {
		t.Errorf("Meta.artifacts length = %d, want 0", len(artifacts))
	}
}

func TestNewTaskEndEvent_NilArtifacts(t *testing.T) {
	event := NewTaskEndEvent("task-003", "integration-engineer", "failed", "session-abc", 3000, nil)

	// Check that nil artifacts doesn't cause panic
	artifacts, ok := event.Meta["artifacts"].([]string)
	if !ok {
		// nil slices should still be typed correctly
		if event.Meta["artifacts"] != nil {
			t.Errorf("Meta.artifacts = %v (%T), want nil or []string", event.Meta["artifacts"], event.Meta["artifacts"])
		}
		return
	}
	if artifacts != nil && len(artifacts) != 0 {
		t.Errorf("Meta.artifacts length = %d, want 0 or nil", len(artifacts))
	}
}

func TestTaskEvents_JSONMarshaling(t *testing.T) {
	// Test task_start event
	startEvent := NewTaskStartEvent("task-001", "architect", "design", "session-test")
	startData, err := json.Marshal(startEvent)
	if err != nil {
		t.Fatalf("Failed to marshal task_start event: %v", err)
	}

	var startRaw map[string]interface{}
	if err := json.Unmarshal(startData, &startRaw); err != nil {
		t.Fatalf("Failed to unmarshal task_start JSON: %v", err)
	}

	// Check type in JSON
	if startRaw["type"] != "task_start" {
		t.Errorf("type = %v, want 'task_start'", startRaw["type"])
	}

	// Check meta is present
	startMeta, ok := startRaw["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if startMeta["task_id"] != "task-001" {
		t.Errorf("meta.task_id = %v, want 'task-001'", startMeta["task_id"])
	}

	// Test task_end event
	endEvent := NewTaskEndEvent("task-001", "architect", "success", "session-test", 10000, []string{"artifact.md"})
	endData, err := json.Marshal(endEvent)
	if err != nil {
		t.Fatalf("Failed to marshal task_end event: %v", err)
	}

	var endRaw map[string]interface{}
	if err := json.Unmarshal(endData, &endRaw); err != nil {
		t.Fatalf("Failed to unmarshal task_end JSON: %v", err)
	}

	// Check type in JSON
	if endRaw["type"] != "task_end" {
		t.Errorf("type = %v, want 'task_end'", endRaw["type"])
	}

	// Check meta is present
	endMeta, ok := endRaw["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if endMeta["outcome"] != "success" {
		t.Errorf("meta.outcome = %v, want 'success'", endMeta["outcome"])
	}
	if endMeta["duration_ms"] != float64(10000) { // JSON numbers are float64
		t.Errorf("meta.duration_ms = %v, want 10000", endMeta["duration_ms"])
	}
}

func TestTaskEvents_Pairing(t *testing.T) {
	// Verify that task_start and task_end use consistent field names
	taskID := "task-paired"
	agent := "documentation-engineer"
	sessionID := "session-paired"

	startEvent := NewTaskStartEvent(taskID, agent, "documentation", sessionID)
	endEvent := NewTaskEndEvent(taskID, agent, "success", sessionID, 8000, []string{})

	// Both should have matching task_id
	if startEvent.Meta["task_id"] != endEvent.Meta["task_id"] {
		t.Errorf("task_id mismatch: start=%v, end=%v", startEvent.Meta["task_id"], endEvent.Meta["task_id"])
	}

	// Both should have matching agent
	if startEvent.Meta["agent"] != endEvent.Meta["agent"] {
		t.Errorf("agent mismatch: start=%v, end=%v", startEvent.Meta["agent"], endEvent.Meta["agent"])
	}

	// Both should have matching session_id
	if startEvent.Meta["session_id"] != endEvent.Meta["session_id"] {
		t.Errorf("session_id mismatch: start=%v, end=%v", startEvent.Meta["session_id"], endEvent.Meta["session_id"])
	}
}

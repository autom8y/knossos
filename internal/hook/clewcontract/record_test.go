package clewcontract

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/hook"
)

func TestRecordToolEvent_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	env := &hook.Env{
		Event:    hook.EventPostToolUse,
		ToolName: "Bash",
	}

	toolInput, err := hook.ParseToolInput(`{
		"command": "go test ./...",
		"description": "Run tests"
	}`)
	if err != nil {
		t.Fatalf("Failed to parse tool input: %v", err)
	}

	if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
		t.Fatalf("RecordToolEvent failed: %v", err)
	}

	// Read back the event
	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Type != EventTypeToolCall {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeToolCall)
	}
	if event.Tool != "Bash" {
		t.Errorf("Tool = %v, want Bash", event.Tool)
	}
	if event.Meta["command"] != "go test ./..." {
		t.Errorf("Meta.command = %v, want 'go test ./...'", event.Meta["command"])
	}
	if event.Meta["description"] != "Run tests" {
		t.Errorf("Meta.description = %v, want 'Run tests'", event.Meta["description"])
	}
}

func TestRecordToolEvent_Edit(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	env := &hook.Env{
		Event:    hook.EventPostToolUse,
		ToolName: "Edit",
	}

	toolInput, err := hook.ParseToolInput(`{
		"file_path": "/path/to/file.go",
		"old_string": "foo",
		"new_string": "bar"
	}`)
	if err != nil {
		t.Fatalf("Failed to parse tool input: %v", err)
	}

	if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
		t.Fatalf("RecordToolEvent failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Tool != "Edit" {
		t.Errorf("Tool = %v, want Edit", event.Tool)
	}
	if event.Path != "/path/to/file.go" {
		t.Errorf("Path = %v, want /path/to/file.go", event.Path)
	}
	if event.Meta["has_old_string"] != true {
		t.Errorf("Meta.has_old_string = %v, want true", event.Meta["has_old_string"])
	}
	if event.Meta["has_new_string"] != true {
		t.Errorf("Meta.has_new_string = %v, want true", event.Meta["has_new_string"])
	}
}

func TestRecordToolEvent_Write(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	env := &hook.Env{
		Event:    hook.EventPostToolUse,
		ToolName: "Write",
	}

	toolInput, err := hook.ParseToolInput(`{
		"file_path": "/path/to/new.txt",
		"content": "Hello, World!"
	}`)
	if err != nil {
		t.Fatalf("Failed to parse tool input: %v", err)
	}

	if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
		t.Fatalf("RecordToolEvent failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Tool != "Write" {
		t.Errorf("Tool = %v, want Write", event.Tool)
	}
	if event.Path != "/path/to/new.txt" {
		t.Errorf("Path = %v, want /path/to/new.txt", event.Path)
	}
	if event.Meta["content_length"] != float64(13) { // JSON numbers are float64
		t.Errorf("Meta.content_length = %v, want 13", event.Meta["content_length"])
	}
}

func TestRecordToolEvent_Read(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	env := &hook.Env{
		Event:    hook.EventPostToolUse,
		ToolName: "Read",
	}

	toolInput, err := hook.ParseToolInput(`{
		"file_path": "/path/to/file.txt",
		"limit": 100,
		"offset": 50
	}`)
	if err != nil {
		t.Fatalf("Failed to parse tool input: %v", err)
	}

	if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
		t.Fatalf("RecordToolEvent failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Tool != "Read" {
		t.Errorf("Tool = %v, want Read", event.Tool)
	}
	if event.Meta["limit"] != float64(100) {
		t.Errorf("Meta.limit = %v, want 100", event.Meta["limit"])
	}
	if event.Meta["offset"] != float64(50) {
		t.Errorf("Meta.offset = %v, want 50", event.Meta["offset"])
	}
}

func TestRecordToolEvent_Glob(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	env := &hook.Env{
		Event:    hook.EventPostToolUse,
		ToolName: "Glob",
	}

	toolInput, err := hook.ParseToolInput(`{
		"pattern": "**/*.go",
		"path": "/project"
	}`)
	if err != nil {
		t.Fatalf("Failed to parse tool input: %v", err)
	}

	if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
		t.Fatalf("RecordToolEvent failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Tool != "Glob" {
		t.Errorf("Tool = %v, want Glob", event.Tool)
	}
	if event.Meta["pattern"] != "**/*.go" {
		t.Errorf("Meta.pattern = %v, want '**/*.go'", event.Meta["pattern"])
	}
}

func TestRecordToolEvent_Grep(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	env := &hook.Env{
		Event:    hook.EventPostToolUse,
		ToolName: "Grep",
	}

	toolInput, err := hook.ParseToolInput(`{
		"pattern": "func.*Test",
		"path": "/project"
	}`)
	if err != nil {
		t.Fatalf("Failed to parse tool input: %v", err)
	}

	if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
		t.Fatalf("RecordToolEvent failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Tool != "Grep" {
		t.Errorf("Tool = %v, want Grep", event.Tool)
	}
	if event.Meta["pattern"] != "func.*Test" {
		t.Errorf("Meta.pattern = %v, want 'func.*Test'", event.Meta["pattern"])
	}
}

func TestRecordToolEvent_Task(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	env := &hook.Env{
		Event:    hook.EventPostToolUse,
		ToolName: "Task",
	}

	toolInput, err := hook.ParseToolInput(`{}`)
	if err != nil {
		t.Fatalf("Failed to parse tool input: %v", err)
	}

	if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
		t.Fatalf("RecordToolEvent failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Tool != "Task" {
		t.Errorf("Tool = %v, want Task", event.Tool)
	}
	if event.Meta["delegation"] != true {
		t.Errorf("Meta.delegation = %v, want true", event.Meta["delegation"])
	}
}

func TestGetEventsPath(t *testing.T) {
	sessionDir := "/some/session/dir"
	expected := "/some/session/dir/events.jsonl"

	result := GetEventsPath(sessionDir)
	if result != expected {
		t.Errorf("GetEventsPath() = %v, want %v", result, expected)
	}
}

func TestRecordStamp(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	rejected := []string{"MongoDB", "CockroachDB"}

	if err := RecordStamp(sessionDir, "Use PostgreSQL", "Better ACID compliance", rejected); err != nil {
		t.Fatalf("RecordStamp failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Type != EventTypeDecision {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeDecision)
	}
	if event.Summary != "Use PostgreSQL" {
		t.Errorf("Summary = %v, want 'Use PostgreSQL'", event.Summary)
	}
	if event.Meta["rationale"] != "Better ACID compliance" {
		t.Errorf("Meta.rationale = %v, want 'Better ACID compliance'", event.Meta["rationale"])
	}

	// Check rejected is present
	rejectedMeta, ok := event.Meta["rejected"].([]any)
	if !ok {
		t.Fatal("Meta.rejected is not a slice")
	}
	if len(rejectedMeta) != 2 {
		t.Errorf("Meta.rejected length = %d, want 2", len(rejectedMeta))
	}
}

func TestRecordStamp_NoRejected(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	if err := RecordStamp(sessionDir, "Proceed with plan", "No blockers", nil); err != nil {
		t.Fatalf("RecordStamp failed: %v", err)
	}

	event := readLastEvent(t, filepath.Join(sessionDir, EventsFileName))

	if event.Type != EventTypeDecision {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeDecision)
	}
	if event.Summary != "Proceed with plan" {
		t.Errorf("Summary = %v, want 'Proceed with plan'", event.Summary)
	}

	// rejected should not be in meta
	if _, exists := event.Meta["rejected"]; exists {
		t.Error("Meta.rejected should not exist when no alternatives rejected")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// Integration test: simulate PostToolUse sequence
func TestIntegration_PostToolUseSequence(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")

	// Simulate a typical sequence: Read -> Edit -> Write -> Bash
	sequence := []struct {
		tool      string
		inputJSON string
	}{
		{"Read", `{"file_path": "/src/main.go"}`},
		{"Edit", `{"file_path": "/src/main.go", "old_string": "old", "new_string": "new"}`},
		{"Write", `{"file_path": "/src/new_file.go", "content": "package main"}`},
		{"Bash", `{"command": "go test ./...", "description": "Run tests"}`},
	}

	for _, s := range sequence {
		env := &hook.Env{
			Event:    hook.EventPostToolUse,
			ToolName: s.tool,
		}
		toolInput, err := hook.ParseToolInput(s.inputJSON)
		if err != nil {
			t.Fatalf("Failed to parse input for %s: %v", s.tool, err)
		}

		if err := RecordToolEvent(sessionDir, env, toolInput); err != nil {
			t.Fatalf("RecordToolEvent failed for %s: %v", s.tool, err)
		}
	}

	// Read all events
	events := readAllEvents(t, filepath.Join(sessionDir, EventsFileName))

	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}

	// Verify sequence order
	expectedTools := []string{"Read", "Edit", "Write", "Bash"}
	for i, tool := range expectedTools {
		if events[i].Tool != tool {
			t.Errorf("Event %d tool = %v, want %v", i, events[i].Tool, tool)
		}
	}

	// Verify all events have timestamps
	for i, e := range events {
		if e.Timestamp == "" {
			t.Errorf("Event %d has empty timestamp", i)
		}
	}
}

// Helper functions

func readLastEvent(t *testing.T, path string) Event {
	t.Helper()
	events := readAllEvents(t, path)
	if len(events) == 0 {
		t.Fatal("No events in file")
	}
	return events[len(events)-1]
}

func readAllEvents(t *testing.T, path string) []Event {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("Failed to parse event: %v", err)
		}
		events = append(events, e)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}

	return events
}

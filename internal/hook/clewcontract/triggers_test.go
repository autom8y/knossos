package clewcontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTriggerType_Constants(t *testing.T) {
	tests := []struct {
		triggerType TriggerType
		expected    string
	}{
		{TriggerFileCount, "file_count_threshold"},
		{TriggerContextSwitch, "context_switch"},
		{TriggerFailureRepeat, "failure_repeat"},
		{TriggerSacredPath, "sacred_path"},
	}

	for _, tt := range tests {
		if string(tt.triggerType) != tt.expected {
			t.Errorf("TriggerType %v = %q, want %q", tt.triggerType, tt.triggerType, tt.expected)
		}
	}
}

func TestDefaultTriggerConfig(t *testing.T) {
	config := DefaultTriggerConfig()

	if config.FileCountThreshold != 5 {
		t.Errorf("FileCountThreshold = %d, want 5", config.FileCountThreshold)
	}

	if len(config.SacredPaths) == 0 {
		t.Error("SacredPaths should not be empty")
	}

	// Verify default sacred paths
	expectedPaths := []string{".claude/", "*_CONTEXT.md", "CLAUDE.md", "docs/decisions/", "docs/requirements/"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range config.SacredPaths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SacredPaths missing expected path: %s", expected)
		}
	}
}

func TestTriggerResult_JSON(t *testing.T) {
	result := TriggerResult{
		Triggered: true,
		Type:      TriggerFileCount,
		Reason:    "8 files modified (threshold: 5)",
		Suggest:   "Consider /stamp: what approach are you taking?",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal TriggerResult: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if raw["triggered"] != true {
		t.Errorf("triggered = %v, want true", raw["triggered"])
	}
	if raw["type"] != "file_count_threshold" {
		t.Errorf("type = %v, want file_count_threshold", raw["type"])
	}
	if raw["reason"] != "8 files modified (threshold: 5)" {
		t.Errorf("reason = %v, want '8 files modified (threshold: 5)'", raw["reason"])
	}
	if raw["suggest"] != "Consider /stamp: what approach are you taking?" {
		t.Errorf("suggest = %v, want 'Consider /stamp: what approach are you taking?'", raw["suggest"])
	}
}

func TestTriggerResult_JSON_OmitEmpty(t *testing.T) {
	result := TriggerResult{
		Triggered: false,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal TriggerResult: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if _, exists := raw["type"]; exists {
		t.Error("type should be omitted when empty")
	}
	if _, exists := raw["reason"]; exists {
		t.Error("reason should be omitted when empty")
	}
	if _, exists := raw["suggest"]; exists {
		t.Error("suggest should be omitted when empty")
	}
}

func TestCheckTriggers_SacredPath_ClaudeDir(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	event := Event{
		Type: EventTypeToolCall,
		Tool: "Write",
		Path: "/project/.claude/agents/my-agent.md",
	}

	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, event, config)

	if !result.Triggered {
		t.Error("Expected trigger for .claude/ path")
	}
	if result.Type != TriggerSacredPath {
		t.Errorf("Type = %v, want %v", result.Type, TriggerSacredPath)
	}
}

func TestCheckTriggers_SacredPath_ContextMd(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	event := Event{
		Type: EventTypeToolCall,
		Tool: "Edit",
		Path: "/project/.claude/sessions/abc/SESSION_CONTEXT.md",
	}

	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, event, config)

	if !result.Triggered {
		t.Error("Expected trigger for *_CONTEXT.md path")
	}
	if result.Type != TriggerSacredPath {
		t.Errorf("Type = %v, want %v", result.Type, TriggerSacredPath)
	}
}

func TestCheckTriggers_SacredPath_ClaudeMd(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	event := Event{
		Type: EventTypeToolCall,
		Tool: "Write",
		Path: "/project/CLAUDE.md",
	}

	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, event, config)

	if !result.Triggered {
		t.Error("Expected trigger for CLAUDE.md")
	}
	if result.Type != TriggerSacredPath {
		t.Errorf("Type = %v, want %v", result.Type, TriggerSacredPath)
	}
}

func TestCheckTriggers_SacredPath_DecisionsDir(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	event := Event{
		Type: EventTypeToolCall,
		Tool: "Write",
		Path: "/project/docs/decisions/ADR-001.md",
	}

	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, event, config)

	if !result.Triggered {
		t.Error("Expected trigger for docs/decisions/ path")
	}
	if result.Type != TriggerSacredPath {
		t.Errorf("Type = %v, want %v", result.Type, TriggerSacredPath)
	}
}

func TestCheckTriggers_SacredPath_ReadDoesNotTrigger(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	// Read operations should NOT trigger sacred path
	event := Event{
		Type: EventTypeToolCall,
		Tool: "Read",
		Path: "/project/.claude/agents/my-agent.md",
	}

	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, event, config)

	if result.Triggered {
		t.Error("Read operations should not trigger sacred path")
	}
}

func TestCheckTriggers_FileCount_AtThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create events with 4 unique files (threshold is 5)
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 4; i++ {
		event := Event{
			Type: EventTypeToolCall,
			Tool: "Edit",
			Path: filepath.Join("/project", "file"+string(rune('0'+i))+".go"),
		}
		if err := writer.Write(event); err != nil {
			t.Fatal(err)
		}
	}
	writer.Close()

	// Add 5th file - should trigger
	currentEvent := Event{
		Type: EventTypeToolCall,
		Tool: "Edit",
		Path: "/project/file5.go",
	}

	// Write the 5th event
	writer, _ = NewEventWriter(sessionDir)
	writer.Write(currentEvent)
	writer.Close()

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, currentEvent, config)

	if !result.Triggered {
		t.Error("Expected trigger at file count threshold")
	}
	if result.Type != TriggerFileCount {
		t.Errorf("Type = %v, want %v", result.Type, TriggerFileCount)
	}
}

func TestCheckTriggers_FileCount_BelowThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create events with 3 unique files
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 3; i++ {
		event := Event{
			Type: EventTypeToolCall,
			Tool: "Edit",
			Path: filepath.Join("/project", "file"+string(rune('0'+i))+".go"),
		}
		if err := writer.Write(event); err != nil {
			t.Fatal(err)
		}
	}
	writer.Close()

	currentEvent := Event{
		Type: EventTypeToolCall,
		Tool: "Read",
		Path: "/project/file4.go",
	}

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, currentEvent, config)

	if result.Triggered && result.Type == TriggerFileCount {
		t.Error("Should not trigger below threshold")
	}
}

func TestCheckTriggers_FailureRepeat(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create first failed test event
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}

	firstFailure := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "go test ./...",
			"exit_code": 1,
		},
	}
	if err := writer.Write(firstFailure); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	// Current event is same failure
	currentEvent := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "go test ./...",
			"exit_code": 1,
		},
	}

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, currentEvent, config)

	if !result.Triggered {
		t.Error("Expected trigger for repeated failure")
	}
	if result.Type != TriggerFailureRepeat {
		t.Errorf("Type = %v, want %v", result.Type, TriggerFailureRepeat)
	}
}

func TestCheckTriggers_FailureRepeat_DifferentCommands(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create first failed event
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}

	firstFailure := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "go build ./...",
			"exit_code": 1,
		},
	}
	if err := writer.Write(firstFailure); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	// Current event is different command failure
	currentEvent := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "npm install",
			"exit_code": 1,
		},
	}

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, currentEvent, config)

	// Should not trigger - different command types
	if result.Triggered && result.Type == TriggerFailureRepeat {
		t.Error("Should not trigger for different command failures")
	}
}

func TestCheckTriggers_ContextSwitch(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	event := Event{
		Type:    EventTypeContextSwitch,
		Summary: "Switching to bug fix branch",
	}

	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, event, config)

	if !result.Triggered {
		t.Error("Expected trigger for context switch event")
	}
	if result.Type != TriggerContextSwitch {
		t.Errorf("Type = %v, want %v", result.Type, TriggerContextSwitch)
	}
}

func TestCheckTriggers_NoTrigger(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	event := Event{
		Type: EventTypeToolCall,
		Tool: "Read",
		Path: "/project/src/main.go",
	}

	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, event, config)

	if result.Triggered {
		t.Errorf("Unexpected trigger: %v - %s", result.Type, result.Reason)
	}
}

func TestCountUniqueFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}

	events := []Event{
		{Type: EventTypeToolCall, Tool: "Edit", Path: "/a.go"},
		{Type: EventTypeToolCall, Tool: "Edit", Path: "/b.go"},
		{Type: EventTypeToolCall, Tool: "Edit", Path: "/a.go"}, // Duplicate
		{Type: EventTypeToolCall, Tool: "Write", Path: "/c.go"},
		{Type: EventTypeToolCall, Tool: "Read", Path: "/d.go"}, // Read doesn't count
	}

	for _, e := range events {
		if err := writer.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	writer.Close()

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	count := CountUniqueFiles(eventsPath)

	if count != 3 { // a.go, b.go, c.go
		t.Errorf("CountUniqueFiles = %d, want 3", count)
	}
}

func TestCountUniqueFiles_NoFile(t *testing.T) {
	count := CountUniqueFiles("/nonexistent/path/events.jsonl")
	if count != 0 {
		t.Errorf("CountUniqueFiles for missing file = %d, want 0", count)
	}
}

func TestDetectRepeatedFailures(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}

	// Write a failed test
	failure := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "go test ./pkg/...",
			"exit_code": 1,
		},
	}
	if err := writer.Write(failure); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	// Check if similar failure is detected
	currentEvent := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "go test ./internal/...",
			"exit_code": 1,
		},
	}

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	if !DetectRepeatedFailures(eventsPath, currentEvent) {
		t.Error("Expected to detect repeated test failure")
	}
}

func TestDetectRepeatedFailures_SuccessDoesNotCount(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}

	// Write a successful test
	success := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "go test ./pkg/...",
			"exit_code": 0,
		},
	}
	if err := writer.Write(success); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	// Check if failure is detected as repeated (should not be)
	currentEvent := Event{
		Type: EventTypeToolCall,
		Tool: "Bash",
		Meta: map[string]interface{}{
			"command":   "go test ./internal/...",
			"exit_code": 1,
		},
	}

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	if DetectRepeatedFailures(eventsPath, currentEvent) {
		t.Error("Success should not count as repeated failure")
	}
}

func TestMatchSacredPattern(t *testing.T) {
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		// Directory patterns
		{"/project/.claude/agents/foo.md", ".claude/", true},
		{"/project/.claude/hooks/bar.sh", ".claude/", true},
		{"/project/src/main.go", ".claude/", false},

		// Wildcard patterns
		{"/project/SESSION_CONTEXT.md", "*_CONTEXT.md", true},
		{"/project/SPRINT_CONTEXT.md", "*_CONTEXT.md", true},
		{"/project/context.md", "*_CONTEXT.md", false},

		// Exact filename patterns
		{"/project/CLAUDE.md", "CLAUDE.md", true},
		{"/project/sub/CLAUDE.md", "CLAUDE.md", true},
		{"/project/README.md", "CLAUDE.md", false},

		// Path patterns
		{"/project/docs/decisions/ADR-001.md", "docs/decisions/", true},
		{"/project/docs/requirements/PRD-001.md", "docs/requirements/", true},
		{"/project/docs/other/file.md", "docs/decisions/", false},
	}

	for _, tt := range tests {
		result := matchSacredPattern(tt.path, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchSacredPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.expected)
		}
	}
}

func TestNormalizeCommandForComparison(t *testing.T) {
	tests := []struct {
		command  string
		expected string
	}{
		{"go test ./...", "go:test"},
		{"go test -v ./pkg/...", "go:test"},
		{"npm test", "npm:test"},
		{"go build ./...", "go:build"},
		{"npm run build", "npm:build"},
		{"ls -la", "ls"},
		{"", ""},
	}

	for _, tt := range tests {
		result := normalizeCommandForComparison(tt.command)
		if result != tt.expected {
			t.Errorf("normalizeCommandForComparison(%q) = %q, want %q", tt.command, result, tt.expected)
		}
	}
}

func TestCheckTriggers_Priority_SacredPathFirst(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create 10 events to trigger file count
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		event := Event{
			Type: EventTypeToolCall,
			Tool: "Edit",
			Path: filepath.Join("/project/src", "file"+string(rune('0'+i))+".go"),
		}
		if err := writer.Write(event); err != nil {
			t.Fatal(err)
		}
	}
	writer.Close()

	// Current event writes to sacred path AND would trigger file count
	currentEvent := Event{
		Type: EventTypeToolCall,
		Tool: "Write",
		Path: "/project/.claude/config.json",
	}

	eventsPath := filepath.Join(sessionDir, EventsFileName)
	config := DefaultTriggerConfig()
	result := CheckTriggers(eventsPath, currentEvent, config)

	// Sacred path should be checked first
	if !result.Triggered {
		t.Error("Expected trigger")
	}
	if result.Type != TriggerSacredPath {
		t.Errorf("Type = %v, want %v (sacred path has priority)", result.Type, TriggerSacredPath)
	}
}

func TestReadEventsFromFile_MalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, EventsFileName)

	// Write some valid and invalid JSON lines
	content := `{"ts":"2024-01-04T10:00:00.000Z","type":"tool.call","tool":"Read"}
not valid json
{"ts":"2024-01-04T10:01:00.000Z","type":"tool.call","tool":"Edit","path":"/a.go"}
`
	if err := os.WriteFile(eventsPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	events, err := readEventsFromFile(eventsPath)
	if err != nil {
		t.Fatalf("readEventsFromFile failed: %v", err)
	}

	// Should have 2 valid events (malformed line skipped)
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestClewOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   ClewOutput
		expected string
	}{
		{
			name: "recorded successfully",
			output: ClewOutput{
				Recorded:   true,
				EventsFile: "/path/to/events.jsonl",
			},
			expected: "Event recorded to /path/to/events.jsonl",
		},
		{
			name: "not recorded with reason",
			output: ClewOutput{
				Recorded: false,
				Reason:   "no active session",
			},
			expected: "Not recorded: no active session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.Text()
			if result != tt.expected {
				t.Errorf("Text() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRunClew_HooksDisabled(t *testing.T) {
	// Clear environment to disable hooks
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Edit",
		UseAriHooks: false,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:  &outputFlag,
				Verbose: &verboseFlag,
			},
		},
	}

	err := runClewWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Recorded {
		t.Error("Expected Recorded=false when hooks disabled")
	}
	if result.Reason != "hooks disabled" {
		t.Errorf("Reason = %q, want %q", result.Reason, "hooks disabled")
	}
}

func TestRunClew_WrongEventType(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
		ToolName:    "Edit",
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:  &outputFlag,
				Verbose: &verboseFlag,
			},
		},
	}

	err := runClewWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Recorded {
		t.Error("Expected Recorded=false for wrong event type")
	}
	if result.Reason != "not a PostToolUse event" {
		t.Errorf("Reason = %q, want %q", result.Reason, "not a PostToolUse event")
	}
}

func TestRunClew_NoActiveSession(t *testing.T) {
	// Create temp project dir without session
	tmpDir := t.TempDir()

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Edit",
		ProjectDir:  tmpDir,
		ToolInput:   `{"file_path": "/some/file.go"}`,
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runClewWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Recorded {
		t.Error("Expected Recorded=false when no active session")
	}
	if result.Reason != "no active session" {
		t.Errorf("Reason = %q, want %q", result.Reason, "no active session")
	}
}

func TestRunClew_WithActiveSession(t *testing.T) {
	// Create temp project structure with session
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	sessionsDir := filepath.Join(claudeDir, "sessions")
	sessionID := "test-session-001"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Create session directory structure
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write current session file (must be in sessions/.current-session per paths.go)
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	if err := os.WriteFile(currentSessionFile, []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write current session file: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Edit",
		ProjectDir:  tmpDir,
		ToolInput:   `{"file_path": "/some/file.go", "old_string": "foo", "new_string": "bar"}`,
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runClewWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Recorded {
		t.Errorf("Expected Recorded=true, got false with reason: %s", result.Reason)
	}

	// Verify events.jsonl was created
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	if _, err := os.Stat(eventsPath); os.IsNotExist(err) {
		t.Error("events.jsonl was not created")
	}
}

func TestRunClew_OrchestratorStamping(t *testing.T) {
	// Create temp project structure with session
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	sessionsDir := filepath.Join(claudeDir, "sessions")
	sessionID := "test-session-orchestrator"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Create session directory structure
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write current session file
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	if err := os.WriteFile(currentSessionFile, []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write current session file: %v", err)
	}

	// Simulate orchestrator CONSULTATION_RESPONSE with throughline
	orchestratorResult := `request_id: req-001

directive:
  action: invoke_specialist

specialist:
  name: principal-engineer
  prompt: "Complete prompt here"

state_update:
  current_phase: implementation
  next_phases: [validation]
  routing_rationale: "Design approved"

throughline:
  decision: "Route to principal-engineer for implementation"
  rationale: "TDD-user-auth approved with complete API contracts."
`

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Task",
		ProjectDir:  tmpDir,
		ToolInput:   `{"agent": "orchestrator", "prompt": "Analyze requirements"}`,
		ToolResult:  orchestratorResult,
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runClewWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Recorded {
		t.Errorf("Expected Recorded=true, got false with reason: %s", result.Reason)
	}

	// Verify events.jsonl was created and contains both tool_call AND decision events
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	eventsContent := string(eventsData)

	// Should have a tool_call event for Task
	if !strings.Contains(eventsContent, `"type":"tool_call"`) {
		t.Error("events.jsonl missing tool_call event")
	}
	if !strings.Contains(eventsContent, `"tool":"Task"`) {
		t.Error("events.jsonl missing Task tool")
	}

	// Should have a decision event from the stamp
	if !strings.Contains(eventsContent, `"type":"decision"`) {
		t.Error("events.jsonl missing decision event from stamp")
	}
	if !strings.Contains(eventsContent, "Route to principal-engineer for implementation") {
		t.Error("events.jsonl missing decision summary from throughline")
	}
	if !strings.Contains(eventsContent, "TDD-user-auth approved with complete API contracts") {
		t.Error("events.jsonl missing rationale from throughline")
	}
}

func TestRunClew_NonOrchestratorTask(t *testing.T) {
	// Create temp project structure with session
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	sessionsDir := filepath.Join(claudeDir, "sessions")
	sessionID := "test-session-regular-task"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Create session directory structure
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write current session file
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	if err := os.WriteFile(currentSessionFile, []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write current session file: %v", err)
	}

	// Regular Task result without throughline (non-orchestrator)
	regularTaskResult := `Task completed successfully.

Files modified:
- /path/to/file.go
`

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Task",
		ProjectDir:  tmpDir,
		ToolInput:   `{"agent": "principal-engineer", "prompt": "Implement feature"}`,
		ToolResult:  regularTaskResult,
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runClewWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Recorded {
		t.Errorf("Expected Recorded=true, got false with reason: %s", result.Reason)
	}

	// Verify events.jsonl has tool_call but NOT decision event (no throughline)
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	eventsContent := string(eventsData)

	// Should have a tool_call event for Task
	if !strings.Contains(eventsContent, `"type":"tool_call"`) {
		t.Error("events.jsonl missing tool_call event")
	}

	// Should NOT have a decision event (no throughline in non-orchestrator Task)
	if strings.Contains(eventsContent, `"type":"decision"`) {
		t.Error("events.jsonl should NOT have decision event for non-orchestrator Task")
	}
}

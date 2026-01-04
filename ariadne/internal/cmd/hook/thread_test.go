package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/test/hooks/testutil"
)

func TestThreadOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   ThreadOutput
		expected string
	}{
		{
			name: "recorded successfully",
			output: ThreadOutput{
				Recorded:   true,
				EventsFile: "/path/to/events.jsonl",
			},
			expected: "Event recorded to /path/to/events.jsonl",
		},
		{
			name: "not recorded with reason",
			output: ThreadOutput{
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

func TestRunThread_HooksDisabled(t *testing.T) {
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
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runThreadWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ThreadOutput
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

func TestRunThread_WrongEventType(t *testing.T) {
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
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runThreadWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ThreadOutput
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

func TestRunThread_NoActiveSession(t *testing.T) {
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
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &projectDir,
	}

	err := runThreadWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ThreadOutput
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

func TestRunThread_WithActiveSession(t *testing.T) {
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
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &projectDir,
	}

	err := runThreadWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ThreadOutput
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

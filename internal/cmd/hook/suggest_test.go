package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

func TestSuggestOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   SuggestOutput
		contains string
	}{
		{
			name:     "with message",
			output:   SuggestOutput{Message: "no active session"},
			contains: "no active session",
		},
		{
			name:     "no transition detected",
			output:   SuggestOutput{Detected: false},
			contains: "no phase transition detected",
		},
		{
			name: "transition detected",
			output: SuggestOutput{
				Detected:   true,
				Transition: "design -> implementation",
			},
			contains: "design -> implementation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.output.Text()
			if !stringContainsSubstr(got, tt.contains) {
				t.Errorf("Text() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestRunSuggestCore_NoSession(t *testing.T) {
	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSuggestCore(nil, ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SuggestOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Detected {
		t.Error("expected detected=false for no session")
	}
	if result.Message != "no active session" {
		t.Errorf("expected message 'no active session', got: %s", result.Message)
	}
}

func TestRunSuggestCore_WrongEvent(t *testing.T) {
	pipeHookStdin(t, "SessionStart", "", "")

	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSuggestCore(nil, ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SuggestOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Detected {
		t.Error("expected detected=false for wrong event")
	}
	if result.Message != "not a post_tool event" {
		t.Errorf("expected message 'not a post_tool event', got: %s", result.Message)
	}
}

func TestRunSuggestCore_NoPhaseChange(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-suggest-noop"

	// Create session directory structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write SESSION_CONTEXT.md with current_phase: design
	contextContent := `---
schema_version: "2.3"
session_id: session-suggest-noop
status: ACTIVE
created_at: 2026-03-08T10:00:00Z
initiative: "Test phase transition"
complexity: MODULE
active_rite: 10x-dev
current_phase: design
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644)

	// Write cache with same phase (no transition expected)
	os.WriteFile(filepath.Join(sessionDir, SuggestPhaseCacheFile), []byte("design"), 0644)

	pipeHookStdin(t, string(hook.EventPostTool), tmpDir, "")

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSuggestCore(nil, ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SuggestOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Detected {
		t.Error("expected detected=false when phase has not changed")
	}
}

func TestRunSuggestCore_FirstInvocation(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-suggest-first"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	contextContent := `---
schema_version: "2.3"
session_id: session-suggest-first
status: ACTIVE
created_at: 2026-03-08T10:00:00Z
initiative: "Test first invocation"
complexity: TASK
active_rite: 10x-dev
current_phase: requirements
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644)

	// No cache file (first invocation)

	pipeHookStdin(t, string(hook.EventPostTool), tmpDir, "")

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSuggestCore(nil, ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SuggestOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Detected {
		t.Error("expected detected=false on first invocation (no previous phase)")
	}

	// Verify cache was written
	cacheData, err := os.ReadFile(filepath.Join(sessionDir, SuggestPhaseCacheFile))
	if err != nil {
		t.Fatalf("cache file not created: %v", err)
	}
	if string(cacheData) != "requirements" {
		t.Errorf("cache = %q, want %q", string(cacheData), "requirements")
	}
}

func TestRunSuggestCore_PhaseTransitionDetected(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-suggest-transition"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	contextContent := `---
schema_version: "2.3"
session_id: session-suggest-transition
status: ACTIVE
created_at: 2026-03-08T10:00:00Z
initiative: "Test phase transition"
complexity: MODULE
active_rite: 10x-dev
current_phase: implementation
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644)

	// Cache has previous phase (design), current is implementation
	os.WriteFile(filepath.Join(sessionDir, SuggestPhaseCacheFile), []byte("design"), 0644)

	pipeHookStdin(t, string(hook.EventPostTool), tmpDir, "")

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSuggestCore(nil, ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SuggestOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if !result.Detected {
		t.Error("expected detected=true when phase changed")
	}
	if result.Transition != "design -> implementation" {
		t.Errorf("transition = %q, want %q", result.Transition, "design -> implementation")
	}
	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(result.Suggestions))
	}
	if result.Suggestions[0].Kind != "phase_transition" {
		t.Errorf("kind = %q, want %q", result.Suggestions[0].Kind, "phase_transition")
	}
	if !stringContainsSubstr(result.Suggestions[0].Text, "Design phase complete") {
		t.Errorf("suggestion text = %q, expected to contain 'Design phase complete'", result.Suggestions[0].Text)
	}

	// Verify cache was updated to current phase
	cacheData, err := os.ReadFile(filepath.Join(sessionDir, SuggestPhaseCacheFile))
	if err != nil {
		t.Fatalf("cache file not found: %v", err)
	}
	if string(cacheData) != "implementation" {
		t.Errorf("cache = %q, want %q", string(cacheData), "implementation")
	}
}

func TestReadPhaseCache(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("file does not exist", func(t *testing.T) {
		got := readPhaseCache(filepath.Join(tmpDir, "nonexistent"))
		if got != "" {
			t.Errorf("readPhaseCache = %q, want empty", got)
		}
	})

	t.Run("file exists with content", func(t *testing.T) {
		path := filepath.Join(tmpDir, "cache")
		os.WriteFile(path, []byte("design\n"), 0644)
		got := readPhaseCache(path)
		if got != "design" {
			t.Errorf("readPhaseCache = %q, want %q", got, "design")
		}
	})

	t.Run("file exists but empty", func(t *testing.T) {
		path := filepath.Join(tmpDir, "empty-cache")
		os.WriteFile(path, []byte(""), 0644)
		got := readPhaseCache(path)
		if got != "" {
			t.Errorf("readPhaseCache = %q, want empty", got)
		}
	})
}

func TestWritePhaseCache(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cache")

	writePhaseCache(path, "implementation")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cache: %v", err)
	}
	if string(data) != "implementation" {
		t.Errorf("cache = %q, want %q", string(data), "implementation")
	}
}

// stringContainsSubstr checks if s contains substr.
func stringContainsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

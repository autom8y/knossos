package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestAutoparkOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   AutoparkOutput
		contains string
	}{
		{
			name: "was parked",
			output: AutoparkOutput{
				SessionID: "session-test",
				WasParked: true,
			},
			contains: "Session auto-parked: session-test",
		},
		{
			name: "not parked",
			output: AutoparkOutput{
				WasParked: false,
				Message:   "no active session",
			},
			contains: "no active session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.Text()
			if result != tt.contains {
				t.Errorf("Text() = %q, want %q", result, tt.contains)
			}
		})
	}
}

func TestRunAutopark_EarlyExit_HooksDisabled(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.WasParked {
		t.Error("Expected WasParked=false when no context")
	}
}

func TestRunAutopark_WrongEvent(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart", // Not a Stop event
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.WasParked {
		t.Error("Expected WasParked=false for non-Stop event")
	}
	if result.Message != "not a stop event" {
		t.Errorf("Message = %q, want %q", result.Message, "not a stop event")
	}
}

func TestRunAutopark_NoSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal .sos structure (no session)
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "Stop",
		ProjectDir:  tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: nil,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.WasParked {
		t.Error("Expected WasParked=false when no session")
	}
}

func TestRunAutopark_ActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write SESSION_CONTEXT.md with ACTIVE status (unquoted for scan-based discovery)
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: ACTIVE
created_at: "2026-01-04T22:26:13Z"
initiative: "Test"
complexity: "MODULE"
active_rite: "test"
current_phase: "implementation"
---
`
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(contextFile, []byte(sessionContext), 0644); err != nil {
		t.Fatalf("Failed to write session context: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "Stop",
		ProjectDir:  tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: nil,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.WasParked {
		t.Error("Expected WasParked=true")
	}
	if result.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", result.SessionID, sessionID)
	}
	if result.Status != "PARKED" {
		t.Errorf("Status = %q, want %q", result.Status, "PARKED")
	}
	if result.PreviousStatus != "ACTIVE" {
		t.Errorf("PreviousStatus = %q, want %q", result.PreviousStatus, "ACTIVE")
	}
	if result.AutoParkedAt == "" {
		t.Error("AutoParkedAt should not be empty")
	}

	// Verify the file was actually updated
	updatedContent, err := os.ReadFile(contextFile)
	if err != nil {
		t.Fatalf("Failed to read updated context: %v", err)
	}
	if !bytes.Contains(updatedContent, []byte("status: PARKED")) {
		t.Error("Context file should contain 'status: PARKED'")
	}
	if !bytes.Contains(updatedContent, []byte("park_source: auto")) {
		t.Errorf("Context file should contain 'park_source: auto', got:\n%s", string(updatedContent))
	}
}

func TestRunAutopark_AlreadyParked(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write SESSION_CONTEXT.md with PARKED status
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: PARKED
created_at: "2026-01-04T22:26:13Z"
initiative: "Test"
complexity: "MODULE"
active_rite: "test"
current_phase: "implementation"
parked_at: "2026-01-04T23:00:00Z"
parked_reason: "manual"
---
`
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	os.WriteFile(contextFile, []byte(sessionContext), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "Stop",
		ProjectDir:  tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

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

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.WasParked {
		t.Error("Expected WasParked=false when already parked")
	}
	if !bytes.Contains([]byte(result.Message), []byte("not active")) {
		t.Errorf("Message should indicate session not active, got: %q", result.Message)
	}
}

// BenchmarkAutoparkHook_EarlyExit benchmarks the early exit path.
func BenchmarkAutoparkHook_EarlyExit(b *testing.B) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runAutoparkCore(nil, ctx, printer)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Early exit took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}


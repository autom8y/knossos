package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestSessionEndOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   SessionEndOutput
		contains string
	}{
		{
			name: "ended with auto-park",
			output: SessionEndOutput{
				SessionID: "session-test",
				WasEnded:  true,
				WasParked: true,
			},
			contains: "Session ended: session-test (auto-parked)",
		},
		{
			name: "ended without park",
			output: SessionEndOutput{
				SessionID: "session-test",
				WasEnded:  true,
				WasParked: false,
			},
			contains: "Session ended: session-test",
		},
		{
			name: "not ended",
			output: SessionEndOutput{
				WasEnded: false,
				Message:  "no active session",
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

func TestRunSessionEnd_WrongEvent(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event: "Stop", // Not a SessionEnd event
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

	err := runSessionEndCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runSessionEnd() error = %v", err)
	}

	var result SessionEndOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.WasEnded {
		t.Error("Expected WasEnded=false for non-SessionEnd event")
	}
	if result.Message != "not a session_end event" {
		t.Errorf("Message = %q, want %q", result.Message, "not a session_end event")
	}
}

func TestRunSessionEnd_ActiveSession_SetsParkSourceAuto(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write SESSION_CONTEXT.md with ACTIVE status
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
		Event:      "SessionEnd",
		ProjectDir: tmpDir,
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

	err := runSessionEndCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runSessionEnd() error = %v", err)
	}

	var result SessionEndOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.WasEnded {
		t.Error("Expected WasEnded=true")
	}
	if !result.WasParked {
		t.Error("Expected WasParked=true for active session")
	}
	if result.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", result.SessionID, sessionID)
	}
	if result.Status != "PARKED" {
		t.Errorf("Status = %q, want %q", result.Status, "PARKED")
	}

	// Verify the file was updated with park_source: auto
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

func TestRunSessionEnd_AlreadyParked_NoParkSourceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write SESSION_CONTEXT.md with PARKED status and manual park_source
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
parked_reason: "Manual park"
park_source: manual
---
`
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	os.WriteFile(contextFile, []byte(sessionContext), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionEnd",
		ProjectDir: tmpDir,
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

	err := runSessionEndCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runSessionEnd() error = %v", err)
	}

	var result SessionEndOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if !result.WasEnded {
		t.Error("Expected WasEnded=true (session ended even if already parked)")
	}
	if result.WasParked {
		t.Error("Expected WasParked=false when already parked")
	}

	// Verify park_source was NOT overwritten (still "manual")
	updatedContent, err := os.ReadFile(contextFile)
	if err != nil {
		t.Fatalf("Failed to read context: %v", err)
	}
	if !bytes.Contains(updatedContent, []byte("park_source: manual")) {
		t.Errorf("Context file should still contain 'park_source: manual', got:\n%s", string(updatedContent))
	}
}

package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

func TestPrecompact_NoSession(t *testing.T) {
	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result hook.PreCompactOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.HookEventName != "PreCompact" {
		t.Errorf("expected hookEventName=PreCompact, got: %s", result.HookSpecificOutput.HookEventName)
	}
	if result.HookSpecificOutput.Reason != "" {
		t.Errorf("expected empty reason for no session, got: %s", result.HookSpecificOutput.Reason)
	}
}

func TestPrecompact_NonPreCompactEvent(t *testing.T) {
	// Set environment for a different event type
	os.Setenv("CLAUDE_HOOK_EVENT", "PostToolUse")
	defer os.Unsetenv("CLAUDE_HOOK_EVENT")

	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result hook.PreCompactOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.Reason != "" {
		t.Errorf("expected empty reason for wrong event, got: %s", result.HookSpecificOutput.Reason)
	}
}

func TestPrecompact_NoSessionContextFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-test-123"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write .current-session so resolveSession finds the session ID
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	os.WriteFile(currentSessionFile, []byte(sessionID), 0644)

	// Set environment for PreCompact event
	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result hook.PreCompactOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.Reason != "" {
		t.Errorf("expected empty reason for no file, got: %s", result.HookSpecificOutput.Reason)
	}
}

func TestPrecompact_RotatesLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-test-456"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write .current-session so resolveSession finds the session ID
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	os.WriteFile(currentSessionFile, []byte(sessionID), 0644)

	// Create large SESSION_CONTEXT.md (frontmatter + 250 lines)
	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	var b strings.Builder
	b.WriteString(`---
session_id: session-test-456
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: test
complexity: simple
active_rite: forge
current_phase: requirements
---
`)
	for i := 1; i <= 250; i++ {
		b.WriteString("Test body line\n")
	}
	if err := os.WriteFile(sessionContextPath, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment for PreCompact event
	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result hook.PreCompactOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if !strings.Contains(result.HookSpecificOutput.Reason, "rotated SESSION_CONTEXT") {
		t.Errorf("expected rotation reason, got: %s", result.HookSpecificOutput.Reason)
	}
	if !strings.Contains(result.HookSpecificOutput.Reason, "archived") {
		t.Errorf("expected archived count in reason, got: %s", result.HookSpecificOutput.Reason)
	}

	// Verify archive was created
	archivePath := filepath.Join(sessionDir, "SESSION_CONTEXT.archived.md")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("archive file was not created")
	}

	// Verify rotated file is smaller
	rotatedContent, err := os.ReadFile(sessionContextPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Count(string(rotatedContent), "\n")
	if lines >= 250 {
		t.Errorf("expected file to be rotated (smaller), got %d lines", lines)
	}
}

func TestPrecompact_NoRotationForSmallFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-test-789"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write .current-session so resolveSession finds the session ID
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	os.WriteFile(currentSessionFile, []byte(sessionID), 0644)

	// Create small SESSION_CONTEXT.md
	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	content := `---
session_id: session-test-789
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: test
complexity: simple
active_rite: forge
current_phase: requirements
---
Small body
Just a few lines
`
	if err := os.WriteFile(sessionContextPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment for PreCompact event
	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result hook.PreCompactOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.Reason != "" {
		t.Errorf("expected empty reason for no rotation, got: %s", result.HookSpecificOutput.Reason)
	}

	// Verify no archive was created
	archivePath := filepath.Join(sessionDir, "SESSION_CONTEXT.archived.md")
	if _, err := os.Stat(archivePath); err == nil {
		t.Error("archive file should not have been created")
	}

	// Verify original file unchanged
	afterContent, _ := os.ReadFile(sessionContextPath)
	if string(afterContent) != content {
		t.Error("file was modified when it shouldn't have been")
	}
}

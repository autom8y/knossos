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

	var result precompactResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	// PreCompact is a side-effect hook — no decision fields expected
	if result.Reason != "" {
		t.Errorf("expected empty reason for no session, got: %s", result.Reason)
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

	var result precompactResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Reason != "" {
		t.Errorf("expected empty reason for wrong event, got: %s", result.Reason)
	}
}

func TestPrecompact_NoSessionContextFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-test-123"
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Set environment for PreCompact event
	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

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

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result precompactResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Reason != "" {
		t.Errorf("expected empty reason for no file, got: %s", result.Reason)
	}
}

func TestPrecompact_RotatesLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-test-456"
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

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

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result precompactResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if !strings.Contains(result.Reason, "rotated SESSION_CONTEXT") {
		t.Errorf("expected rotation reason, got: %s", result.Reason)
	}
	if !strings.Contains(result.Reason, "archived") {
		t.Errorf("expected archived count in reason, got: %s", result.Reason)
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
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

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

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result precompactResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Reason != "" {
		t.Errorf("expected empty reason for no rotation, got: %s", result.Reason)
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

func TestPrecompact_WritesCompactCheckpoint(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-checkpoint-001"
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Create large SESSION_CONTEXT.md to trigger rotation
	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	var b strings.Builder
	b.WriteString(`---
session_id: session-checkpoint-001
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: checkpoint test
complexity: MODULE
active_rite: ecosystem
current_phase: implementation
---
`)
	for i := 1; i <= 250; i++ {
		b.WriteString("Body content line\n")
	}
	os.WriteFile(sessionContextPath, []byte(b.String()), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

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

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify COMPACT_STATE.md was written
	checkpointPath := filepath.Join(sessionDir, CompactCheckpointFile)
	checkpointData, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("COMPACT_STATE.md was not created: %v", err)
	}

	content := string(checkpointData)
	// Verify key fields are in the checkpoint
	if !strings.Contains(content, "session-checkpoint-001") {
		t.Error("checkpoint missing session_id")
	}
	if !strings.Contains(content, "checkpoint test") {
		t.Error("checkpoint missing initiative")
	}
	if !strings.Contains(content, "ecosystem") {
		t.Error("checkpoint missing active_rite")
	}
	if !strings.Contains(content, "implementation") {
		t.Error("checkpoint missing current_phase")
	}
	if !strings.Contains(content, "MODULE") {
		t.Error("checkpoint missing complexity")
	}
}

func TestPrecompact_CheckpointSmallFile(t *testing.T) {
	// Even without rotation, checkpoint should still be written
	tmpDir := t.TempDir()
	sessionID := "session-checkpoint-small"
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Small SESSION_CONTEXT.md (won't trigger rotation)
	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	sessionContent := `---
session_id: session-checkpoint-small
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: small test
complexity: simple
active_rite: forge
current_phase: requirements
---
Small body content.
`
	os.WriteFile(sessionContextPath, []byte(sessionContent), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

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

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify checkpoint was written even without rotation
	checkpointPath := filepath.Join(sessionDir, CompactCheckpointFile)
	checkpointData, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("COMPACT_STATE.md was not created for small file: %v", err)
	}

	if !strings.Contains(string(checkpointData), "session-checkpoint-small") {
		t.Error("checkpoint missing session_id")
	}
}

func TestPrecompact_CheckpointIncludesThroughlineIDs(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-checkpoint-throughline"
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	sessionContent := `---
session_id: session-checkpoint-throughline
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: throughline test
complexity: simple
active_rite: ecosystem
current_phase: requirements
---
Small body.
`
	os.WriteFile(sessionContextPath, []byte(sessionContent), 0644)

	// Pre-seed .throughline-ids.json
	idData := `{"pythia":"agent-pythia-ck1","moirai":"agent-moirai-ck2"}`
	os.WriteFile(filepath.Join(sessionDir, ThroughlineIDsFile), []byte(idData), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

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

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify COMPACT_STATE.md includes throughline IDs section
	checkpointPath := filepath.Join(sessionDir, CompactCheckpointFile)
	checkpointData, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("COMPACT_STATE.md was not created: %v", err)
	}

	content := string(checkpointData)
	if !strings.Contains(content, "Throughline Agents") {
		t.Error("checkpoint missing Throughline Agents section")
	}
	if !strings.Contains(content, "agent-pythia-ck1") {
		t.Error("checkpoint missing pythia agent ID")
	}
	if !strings.Contains(content, "agent-moirai-ck2") {
		t.Error("checkpoint missing moirai agent ID")
	}
}

func TestPrecompact_CheckpointNoThroughlineIDsSection(t *testing.T) {
	// When no .throughline-ids.json exists, the Throughline Agents section should NOT appear
	tmpDir := t.TempDir()
	sessionID := "session-checkpoint-no-throughline"
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	sessionContent := `---
session_id: session-checkpoint-no-throughline
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: no throughline test
complexity: simple
active_rite: forge
current_phase: requirements
---
Small body.
`
	os.WriteFile(sessionContextPath, []byte(sessionContent), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventPreCompact))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

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

	if err := runPrecompactCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkpointPath := filepath.Join(sessionDir, CompactCheckpointFile)
	checkpointData, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("COMPACT_STATE.md was not created: %v", err)
	}

	if strings.Contains(string(checkpointData), "Throughline Agents") {
		t.Error("checkpoint should not contain Throughline Agents section when no IDs file exists")
	}
}

package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestContextOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   ContextOutput
		contains []string
	}{
		{
			name: "with session",
			output: ContextOutput{
				SessionID:     "session-20260104-222613-05a12c6b",
				Status:        "ACTIVE",
				Initiative:    "Test Initiative",
				Rite:          "10x-dev",
				CurrentPhase:  "design",
				ExecutionMode: "orchestrated",
				HasSession:    true,
			},
			contains: []string{
				"## Session Context",
				"| Session | session-20260104-222613-05a12c6b |",
				"| Status | ACTIVE |",
				"| Initiative | Test Initiative |",
				"| Rite | 10x-dev |",
				"| Mode | orchestrated |",
			},
		},
		{
			name:     "no session",
			output:   ContextOutput{HasSession: false},
			contains: []string{"No active session"},
		},
		{
			name: "with git and ecosystem fields",
			output: ContextOutput{
				SessionID:       "session-001",
				Status:          "ACTIVE",
				ExecutionMode:   "orchestrated",
				HasSession:      true,
				GitBranch:       "feat/backtick-migration",
				BaseBranch:      "main",
				AvailableRites:  []string{"10x-dev", "hygiene", "ecosystem"},
				AvailableAgents: []string{"orchestrator", "context-architect"},
			},
			contains: []string{
				"| Available Rites | 10x-dev, hygiene, ecosystem |",
				"| Available Agents | orchestrator, context-architect |",
			},
		},
		{
			name: "omits empty git fields",
			output: ContextOutput{
				SessionID:     "session-002",
				Status:        "ACTIVE",
				ExecutionMode: "cross-cutting",
				HasSession:    true,
			},
			contains: []string{
				"## Session Context",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.Text()
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("Text() missing expected content: %q\nGot: %s", want, result)
				}
			}
		})
	}
}

func TestContextOutput_Text_OmitsEmptyFields(t *testing.T) {
	// Verify that empty git/ecosystem fields do not appear in output
	out := ContextOutput{
		SessionID:     "session-003",
		Status:        "ACTIVE",
		ExecutionMode: "cross-cutting",
		HasSession:    true,
		GitBranch:     "",
		BaseBranch:    "",
	}
	text := out.Text()
	if strings.Contains(text, "Git Branch") {
		t.Error("Text() should not contain Git Branch when empty")
	}
	if strings.Contains(text, "Base Branch") {
		t.Error("Text() should not contain Base Branch when empty")
	}
	if strings.Contains(text, "Available Rites") {
		t.Error("Text() should not contain Available Rites when nil")
	}
	if strings.Contains(text, "Available Agents") {
		t.Error("Text() should not contain Available Agents when nil")
	}

	// Verify that populated git fields are also excluded from Text()
	// (git fields are only in JSON, never in Text() output)
	out2 := ContextOutput{
		SessionID:     "session-004",
		Status:        "ACTIVE",
		ExecutionMode: "orchestrated",
		HasSession:    true,
		GitBranch:     "feature/something",
		BaseBranch:    "main",
	}
	text2 := out2.Text()
	if strings.Contains(text2, "Git Branch") {
		t.Error("Text() should never contain Git Branch (removed from text output)")
	}
	if strings.Contains(text2, "Base Branch") {
		t.Error("Text() should never contain Base Branch (removed from text output)")
	}
}

func TestDetermineExecutionMode(t *testing.T) {
	tests := []struct {
		name       string
		hasSession bool
		activeRite string
		want       string
	}{
		{
			name:       "nil session is native",
			hasSession: false,
			activeRite: "",
			want:       "native",
		},
		{
			name:       "session with rite is orchestrated",
			hasSession: true,
			activeRite: "10x-dev",
			want:       "orchestrated",
		},
		{
			name:       "session with 'none' rite is cross-cutting",
			hasSession: true,
			activeRite: "none",
			want:       "cross-cutting",
		},
		{
			name:       "session with empty rite is cross-cutting",
			hasSession: true,
			activeRite: "",
			want:       "cross-cutting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sessCtx *session.Context
			if tt.hasSession {
				sessCtx = &session.Context{} // Non-nil
			}
			result := determineExecutionMode(sessCtx, tt.activeRite)
			if result != tt.want {
				t.Errorf("determineExecutionMode() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestRunContext_EarlyExit_HooksDisabled(t *testing.T) {
	// Test with no session context
	testutil.SetupEnv(t, &testutil.HookEnv{})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	// Create a minimal context
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionIDVal := ""
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDVal,
		},
	}

	// Override getPrinter to use our buffer
	err := runContextCore(ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	// Parse output
	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HasSession {
		t.Error("Expected HasSession=false when hooks disabled")
	}
}

func TestRunContext_WithActiveSession(t *testing.T) {
	// Create temp directory with session structure
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	// Create .claude/sessions structure
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write SESSION_CONTEXT.md
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: ACTIVE
created_at: "2026-01-04T22:26:13Z"
initiative: "Hooks Migration"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "implementation"
---

# Session: Hooks Migration
`
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(contextFile, []byte(sessionContext), 0644); err != nil {
		t.Fatalf("Failed to write session context: %v", err)
	}

	// Write ACTIVE_RITE
	activeRiteFile := filepath.Join(tmpDir, ".claude", "ACTIVE_RITE")
	if err := os.WriteFile(activeRiteFile, []byte("10x-dev"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	// Setup environment
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
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

	err := runContextCore(ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.HasSession {
		t.Error("Expected HasSession=true")
	}
	if result.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", result.SessionID, sessionID)
	}
	if result.Status != "ACTIVE" {
		t.Errorf("Status = %q, want %q", result.Status, "ACTIVE")
	}
	if result.Initiative != "Hooks Migration" {
		t.Errorf("Initiative = %q, want %q", result.Initiative, "Hooks Migration")
	}
	if result.ExecutionMode != "orchestrated" {
		t.Errorf("ExecutionMode = %q, want %q", result.ExecutionMode, "orchestrated")
	}
}

func TestRunContext_NoSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal .claude structure (no session)
	claudeDir := filepath.Join(tmpDir, ".claude", "sessions")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create claude dir: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
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

	err := runContextCore(ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HasSession {
		t.Error("Expected HasSession=false when no session exists")
	}
}

func TestRunContext_RehydratesCompactState(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-rehydrate"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write SESSION_CONTEXT.md
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-rehydrate"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "Rehydration Test"
complexity: "MODULE"
active_rite: "ecosystem"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".claude", "ACTIVE_RITE"), []byte("ecosystem"), 0644)

	// Write COMPACT_STATE.md (simulating PreCompact wrote it)
	checkpointContent := "# Compact State Checkpoint\n\n| Field | Value |\n|-------|-------|\n| session_id | session-20260208-100000-rehydrate |\n| initiative | Rehydration Test |\n"
	os.WriteFile(filepath.Join(sessionDir, CompactCheckpointFile), []byte(checkpointContent), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
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

	err := runContextCore(ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	// Verify compact state was rehydrated
	if result.CompactState == "" {
		t.Error("Expected CompactState to be populated from COMPACT_STATE.md")
	}
	if !strings.Contains(result.CompactState, "session-20260208-100000-rehydrate") {
		t.Error("CompactState should contain session_id")
	}
	if !strings.Contains(result.CompactState, "Rehydration Test") {
		t.Error("CompactState should contain initiative")
	}

	// Verify checkpoint was consumed (renamed)
	checkpointPath := filepath.Join(sessionDir, CompactCheckpointFile)
	if _, err := os.Stat(checkpointPath); err == nil {
		t.Error("COMPACT_STATE.md should have been renamed after consumption")
	}
	consumedPath := filepath.Join(sessionDir, CompactCheckpointConsumed)
	if _, err := os.Stat(consumedPath); os.IsNotExist(err) {
		t.Error("COMPACT_STATE.consumed.md should exist after consumption")
	}
}

func TestRunContext_NoCompactState(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-nochkpnt"

	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-nochkpnt"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "No Checkpoint"
complexity: "simple"
active_rite: "forge"
current_phase: "requirements"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".claude", "ACTIVE_RITE"), []byte("forge"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
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

	err := runContextCore(ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	// No checkpoint means CompactState should be empty
	if result.CompactState != "" {
		t.Errorf("Expected empty CompactState, got: %s", result.CompactState)
	}

	// Normal fields should still work
	if !result.HasSession {
		t.Error("Expected HasSession=true")
	}
	if result.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", result.SessionID, sessionID)
	}
}

func TestGetGitBranch_InGitRepo(t *testing.T) {
	// This test runs inside the knossos repo, so git branch should be populated
	branch := getGitBranch(".")
	if branch == "" {
		t.Skip("Not running inside a git repository")
	}
	// Branch name should be a non-empty string without newlines
	if strings.Contains(branch, "\n") {
		t.Errorf("getGitBranch() returned string with newline: %q", branch)
	}
}

func TestGetGitBranch_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	branch := getGitBranch(tmpDir)
	if branch != "" {
		t.Errorf("getGitBranch() in non-git dir returned %q, want empty", branch)
	}
}

func TestGetBaseBranch_FallbackToMain(t *testing.T) {
	// In a directory without a git remote, should fall back to "main"
	tmpDir := t.TempDir()
	base := getBaseBranch(tmpDir)
	if base != "main" {
		t.Errorf("getBaseBranch() fallback = %q, want %q", base, "main")
	}
}

func TestListAvailableRites(t *testing.T) {
	tmpDir := t.TempDir()
	ritesDir := filepath.Join(tmpDir, "rites")
	os.MkdirAll(ritesDir, 0755)

	// Create rite with manifest
	riteWithManifest := filepath.Join(ritesDir, "alpha")
	os.MkdirAll(riteWithManifest, 0755)
	os.WriteFile(filepath.Join(riteWithManifest, "manifest.yaml"), []byte("name: alpha"), 0644)

	// Create rite without manifest (should be excluded)
	riteWithoutManifest := filepath.Join(ritesDir, "beta")
	os.MkdirAll(riteWithoutManifest, 0755)

	// Create a file (not a directory, should be excluded)
	os.WriteFile(filepath.Join(ritesDir, "readme.md"), []byte("# Rites"), 0644)

	rites := listAvailableRites(ritesDir)
	if len(rites) != 1 {
		t.Fatalf("listAvailableRites() returned %d rites, want 1: %v", len(rites), rites)
	}
	if rites[0] != "alpha" {
		t.Errorf("listAvailableRites()[0] = %q, want %q", rites[0], "alpha")
	}
}

func TestListAvailableRites_NonexistentDir(t *testing.T) {
	rites := listAvailableRites("/nonexistent/path/rites")
	if rites != nil {
		t.Errorf("listAvailableRites() on missing dir = %v, want nil", rites)
	}
}

func TestListAvailableAgents(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	// Create agent files
	os.WriteFile(filepath.Join(agentsDir, "orchestrator.md"), []byte("# Orchestrator"), 0644)
	os.WriteFile(filepath.Join(agentsDir, "architect.md"), []byte("# Architect"), 0644)

	// Create non-md file (should be excluded)
	os.WriteFile(filepath.Join(agentsDir, "config.yaml"), []byte("key: value"), 0644)

	// Create subdirectory (should be excluded)
	os.MkdirAll(filepath.Join(agentsDir, "subdir"), 0755)

	agents := listAvailableAgents(agentsDir)
	if len(agents) != 2 {
		t.Fatalf("listAvailableAgents() returned %d agents, want 2: %v", len(agents), agents)
	}
	// os.ReadDir returns sorted entries
	if agents[0] != "architect" {
		t.Errorf("listAvailableAgents()[0] = %q, want %q", agents[0], "architect")
	}
	if agents[1] != "orchestrator" {
		t.Errorf("listAvailableAgents()[1] = %q, want %q", agents[1], "orchestrator")
	}
}

func TestListAvailableAgents_NonexistentDir(t *testing.T) {
	agents := listAvailableAgents("/nonexistent/path/agents")
	if agents != nil {
		t.Errorf("listAvailableAgents() on missing dir = %v, want nil", agents)
	}
}

func TestRunContext_WithActiveSession_IncludesRitesAndAgents(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-abcdef01"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-abcdef01"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "Rites Test"
complexity: "MODULE"
active_rite: "alpha"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".claude", "ACTIVE_RITE"), []byte("alpha"), 0644)

	// Create rites directory with manifests
	ritesDir := filepath.Join(tmpDir, "rites")
	os.MkdirAll(filepath.Join(ritesDir, "alpha"), 0755)
	os.WriteFile(filepath.Join(ritesDir, "alpha", "manifest.yaml"), []byte("name: alpha"), 0644)
	os.MkdirAll(filepath.Join(ritesDir, "beta"), 0755)
	os.WriteFile(filepath.Join(ritesDir, "beta", "manifest.yaml"), []byte("name: beta"), 0644)

	// Create agents directory
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(filepath.Join(agentsDir, "orchestrator.md"), []byte("# Orch"), 0644)
	os.WriteFile(filepath.Join(agentsDir, "builder.md"), []byte("# Build"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
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

	err := runContextCore(ctx, printer)
	if err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	// Verify rites
	if len(result.AvailableRites) != 2 {
		t.Errorf("AvailableRites length = %d, want 2: %v", len(result.AvailableRites), result.AvailableRites)
	}

	// Verify agents
	if len(result.AvailableAgents) != 2 {
		t.Errorf("AvailableAgents length = %d, want 2: %v", len(result.AvailableAgents), result.AvailableAgents)
	}

	// GitBranch will be empty since tmpDir is not a git repo — that is expected
	// BaseBranch should fall back to "main"
	if result.BaseBranch != "main" {
		t.Errorf("BaseBranch = %q, want %q (fallback)", result.BaseBranch, "main")
	}
}

// BenchmarkContextHook_EarlyExit benchmarks the early exit path (<5ms target).
func BenchmarkContextHook_EarlyExit(b *testing.B) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionIDVal := ""

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDVal,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runContextCore(ctx, printer)
	}

	// Report ns/op
	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Early exit took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkContextHook_FullExecution benchmarks full execution (<100ms target).
func BenchmarkContextHook_FullExecution(b *testing.B) {
	tmpDir := b.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	// Setup session structure
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: ACTIVE
created_at: "2026-01-04T22:26:13Z"
initiative: "Benchmark"
complexity: "MODULE"
active_rite: "test"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".claude", "ACTIVE_RITE"), []byte("test"), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", "SessionStart")
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
			SessionID: nil,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runContextCore(ctx, printer)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(100*time.Millisecond) {
		b.Errorf("Full execution took %.2f ms, target is <100ms", nsPerOp/1e6)
	}
}


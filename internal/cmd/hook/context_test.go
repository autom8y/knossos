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
			name:       "session with team is orchestrated",
			hasSession: true,
			activeRite: "10x-dev",
			want:       "orchestrated",
		},
		{
			name:       "session with 'none' team is cross-cutting",
			hasSession: true,
			activeRite: "none",
			want:       "cross-cutting",
		},
		{
			name:       "session with empty team is cross-cutting",
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
	err := runContextWithPrinter(ctx, printer)
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

	// Write .current-session
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	if err := os.WriteFile(currentSessionFile, []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write current session: %v", err)
	}

	// Write SESSION_CONTEXT.md
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: "ACTIVE"
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

	err := runContextWithPrinter(ctx, printer)
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

	err := runContextWithPrinter(ctx, printer)
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
		runContextWithPrinter(ctx, printer)
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
	os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: "ACTIVE"
created_at: "2026-01-04T22:26:13Z"
initiative: "Benchmark"
complexity: "MODULE"
active_rite: "test"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".claude", "ACTIVE_RITE"), []byte("test"), 0644)

	os.Setenv("USE_ARI_HOOKS", "1")
	os.Setenv("CLAUDE_HOOK_EVENT", "SessionStart")
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("USE_ARI_HOOKS")
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
		runContextWithPrinter(ctx, printer)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(100*time.Millisecond) {
		b.Errorf("Full execution took %.2f ms, target is <100ms", nsPerOp/1e6)
	}
}

// runContextWithPrinter is a helper that uses an injected printer for testing.
func runContextWithPrinter(ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Get resolver for path lookups
	resolver := ctx.GetResolver()
	if resolver.ProjectRoot() == "" {
		if hookEnv.ProjectDir != "" {
			resolver = newResolverFromPath(hookEnv.ProjectDir)
		} else {
			return outputNoSession(printer)
		}
	}

	// Get current session ID
	sessionIDStr, err := ctx.GetCurrentSessionID()
	if err != nil {
		return outputNoSession(printer)
	}

	if sessionIDStr == "" {
		return outputNoSession(printer)
	}

	sessionIDStr = strings.TrimSpace(sessionIDStr)

	// Load session context using real session package
	ctxPath := resolver.SessionContextFile(sessionIDStr)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		return outputNoSession(printer)
	}

	// Read active rite
	activeRite := resolver.ReadActiveRite()
	if activeRite == "" {
		activeRite = sessCtx.ActiveRite
	}

	mode := determineExecutionMode(sessCtx, activeRite)

	result := ContextOutput{
		SessionID:     sessCtx.SessionID,
		Status:        string(sessCtx.Status),
		Initiative:    sessCtx.Initiative,
		Rite:          activeRite,
		CurrentPhase:  sessCtx.CurrentPhase,
		ExecutionMode: mode,
		HasSession:    true,
	}

	return printer.Print(result)
}

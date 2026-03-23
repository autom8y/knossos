package hook

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// =============================================================================
// C2 Contract: Timeout Enforcement Tests
// Hooks must complete within their configured timeout to avoid blocking Claude.
// =============================================================================

func TestWithTimeout_Success(t *testing.T) {
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
		timeout: 100 * time.Millisecond,
	}

	var called bool
	err := ctx.withTimeout(func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("withTimeout() returned error: %v", err)
	}
	if !called {
		t.Error("Function was not called")
	}
}

func TestWithTimeout_ReturnsError(t *testing.T) {
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
		timeout: 100 * time.Millisecond,
	}

	expectedErr := errors.New("test error")
	err := ctx.withTimeout(func() error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("withTimeout() error = %v, want %v", err, expectedErr)
	}
}

func TestWithTimeout_Timeout(t *testing.T) {
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
		timeout: 10 * time.Millisecond,
	}

	start := time.Now()
	err := ctx.withTimeout(func() error {
		time.Sleep(100 * time.Millisecond) // Sleep longer than timeout
		return nil
	})
	elapsed := time.Since(start)

	// Should timeout, not complete
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("timed out")) {
		t.Errorf("Expected timeout message, got: %v", err)
	}

	// Should complete in approximately the timeout duration, not the sleep duration
	if elapsed > 50*time.Millisecond {
		t.Errorf("Timeout took %v, expected ~10ms", elapsed)
	}
}

func TestDefaultTimeout(t *testing.T) {
	if DefaultTimeout != 100*time.Millisecond {
		t.Errorf("DefaultTimeout = %v, want 100ms", DefaultTimeout)
	}
}

func TestMaxTimeout(t *testing.T) {
	if MaxTimeout != 500*time.Millisecond {
		t.Errorf("MaxTimeout = %v, want 500ms", MaxTimeout)
	}
}

func TestEarlyExitThreshold(t *testing.T) {
	if EarlyExitThreshold != 5*time.Millisecond {
		t.Errorf("EarlyExitThreshold = %v, want 5ms", EarlyExitThreshold)
	}
}

// =============================================================================
// cmdContext Helper Function Tests
// =============================================================================

func TestGetPrinter_DefaultJSON(t *testing.T) {
	ctx := &cmdContext{}

	printer := ctx.getPrinter()
	if printer == nil {
		t.Fatal("getPrinter() returned nil")
	}
}

func TestGetPrinter_WithFormat(t *testing.T) {
	formats := []string{"json", "text", "yaml"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			f := format
			ctx := &cmdContext{
				SessionContext: common.SessionContext{
					BaseContext: common.BaseContext{
						Output: &f,
					},
				},
			}

			printer := ctx.getPrinter()
			if printer == nil {
				t.Fatal("getPrinter() returned nil")
			}
		})
	}
}

func TestGetResolver_Empty(t *testing.T) {
	ctx := &cmdContext{}

	resolver := ctx.GetResolver()
	if resolver == nil {
		t.Fatal("getResolver() returned nil")
	}
}

func TestGetResolver_WithProjectDir(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				ProjectDir: &tmpDir,
			},
		},
	}

	resolver := ctx.GetResolver()
	if resolver == nil {
		t.Fatal("getResolver() returned nil")
	}
	if resolver.ProjectRoot() != tmpDir {
		t.Errorf("ProjectRoot() = %q, want %q", resolver.ProjectRoot(), tmpDir)
	}
}

func TestGetCurrentSessionID_FromContext(t *testing.T) {
	sessionID := "test-session-123"
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			SessionID: &sessionID,
		},
	}

	// GetSessionID checks the flag first, then falls back to file
	result, err := ctx.GetSessionID()
	if err != nil {
		t.Fatalf("GetSessionID() error = %v", err)
	}
	if result != sessionID {
		t.Errorf("GetSessionID() = %q, want %q", result, sessionID)
	}
}

func TestGetSessionID_FromScan(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260209-120000-abcdef01"

	// Create sessions directory with an ACTIVE session
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(contextFile, []byte("---\nstatus: ACTIVE\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				ProjectDir: &tmpDir,
			},
		},
	}

	result, err := ctx.GetSessionID()
	if err != nil {
		t.Fatalf("GetSessionID() error = %v", err)
	}
	if result != sessionID {
		t.Errorf("GetSessionID() = %q, want %q", result, sessionID)
	}
}

func TestGetSessionID_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sessions directory but no active sessions
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				ProjectDir: &tmpDir,
			},
		},
	}

	result, err := ctx.GetSessionID()
	if err != nil {
		t.Fatalf("GetSessionID() error = %v", err)
	}
	if result != "" {
		t.Errorf("GetSessionID() = %q, want empty string", result)
	}
}

func TestGetHookEnv(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Bash",
		ToolInput:  `{"command":"ls"}`,
		ProjectDir: "/test/project",
	})

	ctx := &cmdContext{}
	hookEnv := ctx.getHookEnv()

	if hookEnv == nil {
		t.Fatal("getHookEnv() returned nil")
	}
	if string(hookEnv.Event) != "pre_tool" {
		t.Errorf("Event = %q, want %q", hookEnv.Event, "pre_tool")
	}
	if hookEnv.ToolName != "Bash" {
		t.Errorf("ToolName = %q, want %q", hookEnv.ToolName, "Bash")
	}
}

// =============================================================================
// Integration Tests: Full Hook Command Execution
// =============================================================================

func TestNewHookCmd_SubcommandRegistration(t *testing.T) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""

	cmd := NewHookCmd(&outputFlag, &verboseFlag, &projectDir, &sessionID)

	// Verify all subcommands are registered
	subcommands := []string{"context", "autopark", "writeguard", "validate", "clew", "sessionend"}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			sub, _, err := cmd.Find([]string{name})
			if err != nil {
				t.Errorf("Subcommand %q not found: %v", name, err)
			}
			if sub == nil {
				t.Errorf("Subcommand %q is nil", name)
			}
		})
	}
}

func TestNewHookCmd_TimeoutFlag(t *testing.T) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""

	cmd := NewHookCmd(&outputFlag, &verboseFlag, &projectDir, &sessionID)

	// Verify timeout flag is registered
	flag := cmd.PersistentFlags().Lookup("timeout")
	if flag == nil {
		t.Error("--timeout flag not found")
		return
	}
	if flag.DefValue != "100" {
		t.Errorf("Default timeout = %q, want %q", flag.DefValue, "100")
	}
}

// =============================================================================
// Integration Tests: End-to-End Hook Behavior
// =============================================================================

func TestIntegration_ContextHook_NoSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal .sos structure (no session)
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

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
		timeout: DefaultTimeout,
	}

	err := runContextCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.HasSession {
		t.Error("Expected HasSession=false when no session exists")
	}
}

func TestIntegration_ValidateHook_Chain(t *testing.T) {
	// Test a sequence of commands like Claude would execute
	commands := []struct {
		command  string
		expected string
	}{
		{`{"command": "ls -la"}`, "allow"},
		{`{"command": "git status"}`, "allow"},
		{`{"command": "rm -rf .git"}`, "deny"},
		{`{"command": "git push --force origin main"}`, "deny"},
		{`{"command": "cat README.md"}`, "allow"},
	}

	for _, tc := range commands {
		t.Run(tc.command, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:     "PreToolUse",
				ToolName:  "Bash",
				ToolInput: tc.command,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			projectDir := ""
			ctx := &cmdContext{
				SessionContext: common.SessionContext{
					BaseContext: common.BaseContext{
						Output:     &outputFlag,
						Verbose:    &verboseFlag,
						ProjectDir: &projectDir,
					},
				},
				timeout: DefaultTimeout,
			}

			err := runValidateCore(nil, ctx, printer)
			if err != nil {
				t.Fatalf("runValidate() error = %v", err)
			}

			var result hook.PreToolUseOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v", err)
			}

			if result.HookSpecificOutput.PermissionDecision != tc.expected {
				t.Errorf("Decision = %q, want %q", result.HookSpecificOutput.PermissionDecision, tc.expected)
			}
		})
	}
}

func TestIntegration_WriteguardHook_Chain(t *testing.T) {
	// Test file write protection
	files := []struct {
		path     string
		expected string
	}{
		{"src/main.go", "allow"},
		{".sos/sessions/test/SESSION_CONTEXT.md", "deny"},
		{".sos/sessions/session-abc/sprints/s1/SPRINT_CONTEXT.md", "deny"},
		{"docs/README.md", "allow"},
		{"config.yaml", "allow"},
	}

	for _, tc := range files {
		t.Run(tc.path, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:     "PreToolUse",
				ToolName:  "Write",
				ToolInput: `{"file_path": "` + tc.path + `", "content": "test"}`,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			projectDir := ""
			ctx := &cmdContext{
				SessionContext: common.SessionContext{
					BaseContext: common.BaseContext{
						Output:     &outputFlag,
						Verbose:    &verboseFlag,
						ProjectDir: &projectDir,
					},
				},
				timeout: DefaultTimeout,
			}

			err := runWriteguardCore(nil, ctx, printer)
			if err != nil {
				t.Fatalf("runWriteguard() error = %v", err)
			}

			var result hook.PreToolUseOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v", err)
			}

			if result.HookSpecificOutput.PermissionDecision != tc.expected {
				t.Errorf("Decision = %q, want %q for path %q", result.HookSpecificOutput.PermissionDecision, tc.expected, tc.path)
			}
		})
	}
}

// =============================================================================
// Performance Tests: C2 Contract Compliance
// Hooks must complete within target times to avoid blocking Claude.
// =============================================================================

func BenchmarkHook_EarlyExitPath(b *testing.B) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
		timeout: DefaultTimeout,
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = runContextCore(nil, ctx, printer)
	}

	// Verify early exit is fast
	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(EarlyExitThreshold) {
		b.Errorf("Early exit took %.2f ms, target is <%v", nsPerOp/1e6, EarlyExitThreshold)
	}
}

func BenchmarkHook_TimeoutOverhead(b *testing.B) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
		timeout: DefaultTimeout,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.withTimeout(func() error {
			return nil
		})
	}

	// Verify timeout wrapper has minimal overhead
	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	// Timeout overhead should be < 1ms
	if nsPerOp > float64(1*time.Millisecond) {
		b.Errorf("Timeout overhead took %.2f ms, target is <1ms", nsPerOp/1e6)
	}
}

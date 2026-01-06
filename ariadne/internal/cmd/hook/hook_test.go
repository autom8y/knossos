package hook

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/test/hooks/testutil"
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
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &projectDir,
		sessionID:  &sessionID,
		timeout:    100 * time.Millisecond,
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
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &projectDir,
		sessionID:  &sessionID,
		timeout:    100 * time.Millisecond,
	}

	expectedErr := errors.New("test error")
	err := ctx.withTimeout(func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("withTimeout() error = %v, want %v", err, expectedErr)
	}
}

func TestWithTimeout_Timeout(t *testing.T) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""

	ctx := &cmdContext{
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &projectDir,
		sessionID:  &sessionID,
		timeout:    10 * time.Millisecond, // Very short timeout
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
	ctx := &cmdContext{
		output:  nil,
		verbose: nil,
	}

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
				output:  &f,
				verbose: nil,
			}

			printer := ctx.getPrinter()
			if printer == nil {
				t.Fatal("getPrinter() returned nil")
			}
		})
	}
}

func TestGetResolver_Empty(t *testing.T) {
	ctx := &cmdContext{
		projectDir: nil,
	}

	resolver := ctx.getResolver()
	if resolver == nil {
		t.Fatal("getResolver() returned nil")
	}
}

func TestGetResolver_WithProjectDir(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := &cmdContext{
		projectDir: &tmpDir,
	}

	resolver := ctx.getResolver()
	if resolver == nil {
		t.Fatal("getResolver() returned nil")
	}
	if resolver.ProjectRoot() != tmpDir {
		t.Errorf("ProjectRoot() = %q, want %q", resolver.ProjectRoot(), tmpDir)
	}
}

func TestShouldEarlyExit_HooksDisabled(t *testing.T) {
	// Clear USE_ARI_HOOKS
	testutil.SetupEnv(t, &testutil.HookEnv{
		UseAriHooks: false,
	})

	ctx := &cmdContext{}

	if !ctx.shouldEarlyExit() {
		t.Error("shouldEarlyExit() should return true when hooks disabled")
	}
}

func TestShouldEarlyExit_HooksEnabled(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		UseAriHooks: true,
	})

	ctx := &cmdContext{}

	if ctx.shouldEarlyExit() {
		t.Error("shouldEarlyExit() should return false when hooks enabled")
	}
}

func TestGetCurrentSessionID_FromContext(t *testing.T) {
	sessionID := "test-session-123"
	ctx := &cmdContext{
		sessionID: &sessionID,
	}

	result, err := ctx.getCurrentSessionID()
	if err != nil {
		t.Fatalf("getCurrentSessionID() error = %v", err)
	}
	if result != sessionID {
		t.Errorf("getCurrentSessionID() = %q, want %q", result, sessionID)
	}
}

func TestGetCurrentSessionID_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-from-file"

	// Create sessions directory and .current-session file
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	if err := os.WriteFile(currentSessionFile, []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	ctx := &cmdContext{
		projectDir: &tmpDir,
		sessionID:  nil,
	}

	result, err := ctx.getCurrentSessionID()
	if err != nil {
		t.Fatalf("getCurrentSessionID() error = %v", err)
	}
	if result != sessionID {
		t.Errorf("getCurrentSessionID() = %q, want %q", result, sessionID)
	}
}

func TestGetCurrentSessionID_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &cmdContext{
		projectDir: &tmpDir,
		sessionID:  nil,
	}

	result, err := ctx.getCurrentSessionID()
	if err != nil {
		t.Fatalf("getCurrentSessionID() error = %v", err)
	}
	if result != "" {
		t.Errorf("getCurrentSessionID() = %q, want empty string", result)
	}
}

func TestGetHookEnv(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command":"ls"}`,
		ProjectDir:  "/test/project",
		UseAriHooks: true,
	})

	ctx := &cmdContext{}
	hookEnv := ctx.getHookEnv()

	if hookEnv == nil {
		t.Fatal("getHookEnv() returned nil")
	}
	if string(hookEnv.Event) != "PreToolUse" {
		t.Errorf("Event = %q, want %q", hookEnv.Event, "PreToolUse")
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
	subcommands := []string{"context", "autopark", "writeguard", "route", "validate", "clew"}

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

	// Create minimal .claude structure (no session)
	claudeDir := filepath.Join(tmpDir, ".claude", "sessions")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create claude dir: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
		ProjectDir:  tmpDir,
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &tmpDir,
		sessionID:  nil,
		timeout:    DefaultTimeout,
	}

	err := runContextWithPrinter(ctx, printer)
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
		{`{"command": "rm -rf .git"}`, "block"},
		{`{"command": "git push --force origin main"}`, "block"},
		{`{"command": "cat README.md"}`, "allow"},
	}

	for _, tc := range commands {
		t.Run(tc.command, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:       "PreToolUse",
				ToolName:    "Bash",
				ToolInput:   tc.command,
				UseAriHooks: true,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			projectDir := ""
			ctx := &cmdContext{
				output:     &outputFlag,
				verbose:    &verboseFlag,
				projectDir: &projectDir,
				timeout:    DefaultTimeout,
			}

			err := runValidateWithPrinter(ctx, printer, "")
			if err != nil {
				t.Fatalf("runValidate() error = %v", err)
			}

			var result ValidateDecision
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v", err)
			}

			if result.Decision != tc.expected {
				t.Errorf("Decision = %q, want %q", result.Decision, tc.expected)
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
		{".claude/sessions/test/SESSION_CONTEXT.md", "block"},
		{".claude/sprints/s1/SPRINT_CONTEXT.md", "block"},
		{"docs/README.md", "allow"},
		{"config.yaml", "allow"},
	}

	for _, tc := range files {
		t.Run(tc.path, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:       "PreToolUse",
				ToolName:    "Write",
				ToolInput:   `{"file_path": "` + tc.path + `", "content": "test"}`,
				UseAriHooks: true,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			projectDir := ""
			ctx := &cmdContext{
				output:     &outputFlag,
				verbose:    &verboseFlag,
				projectDir: &projectDir,
				timeout:    DefaultTimeout,
			}

			err := runWriteguardWithPrinter(ctx, printer, "")
			if err != nil {
				t.Fatalf("runWriteguard() error = %v", err)
			}

			var result WriteGuardDecision
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v", err)
			}

			if result.Decision != tc.expected {
				t.Errorf("Decision = %q, want %q for path %q", result.Decision, tc.expected, tc.path)
			}
		})
	}
}

func TestIntegration_RouteHook_AllCategories(t *testing.T) {
	// Test all command categories
	commands := []struct {
		message  string
		command  string
		category CommandCategory
	}{
		{"/start Add feature", "/start", CategorySession},
		{"/park", "/park", CategorySession},
		{"/resume session-123", "/resume", CategorySession},
		{"/wrap", "/wrap", CategorySession},
		{"/consult Which team?", "/consult", CategoryOrchestrator},
		{"/task Implement auth", "/task", CategoryInitiative},
		{"/sprint Q1 work", "/sprint", CategoryInitiative},
		{"/commit Fix bug", "/commit", CategoryGit},
		{"/pr", "/pr", CategoryGit},
		{"/stamp Chose PostgreSQL", "/stamp", CategoryClew},
	}

	for _, tc := range commands {
		t.Run(tc.command, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:       "UserPromptSubmit",
				UserMessage: tc.message,
				UseAriHooks: true,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			ctx := &cmdContext{
				output:  &outputFlag,
				verbose: &verboseFlag,
				timeout: DefaultTimeout,
			}

			err := runRouteWithPrinter(ctx, printer)
			if err != nil {
				t.Fatalf("runRoute() error = %v", err)
			}

			var result RouteOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v", err)
			}

			if !result.Routed {
				t.Error("Expected Routed=true")
			}
			if result.Command != tc.command {
				t.Errorf("Command = %q, want %q", result.Command, tc.command)
			}
			if result.Category != tc.category {
				t.Errorf("Category = %q, want %q", result.Category, tc.category)
			}
		})
	}
}

// =============================================================================
// Performance Tests: C2 Contract Compliance
// Hooks must complete within target times to avoid blocking Claude.
// =============================================================================

func BenchmarkHook_EarlyExitPath(b *testing.B) {
	// Clear hooks to test early exit path
	os.Unsetenv("USE_ARI_HOOKS")

	outputFlag := "json"
	verboseFlag := false
	projectDir := ""

	ctx := &cmdContext{
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &projectDir,
		timeout:    DefaultTimeout,
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runContextWithPrinter(ctx, printer)
	}

	// Verify early exit is fast
	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(EarlyExitThreshold) {
		b.Errorf("Early exit took %.2f ms, target is <%v", nsPerOp/1e6, EarlyExitThreshold)
	}
}

func BenchmarkHook_TimeoutOverhead(b *testing.B) {
	os.Setenv("USE_ARI_HOOKS", "1")
	defer os.Unsetenv("USE_ARI_HOOKS")

	outputFlag := "json"
	verboseFlag := false
	projectDir := ""

	ctx := &cmdContext{
		output:     &outputFlag,
		verbose:    &verboseFlag,
		projectDir: &projectDir,
		timeout:    DefaultTimeout,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.withTimeout(func() error {
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

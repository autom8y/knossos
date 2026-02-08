package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestIsProtectedFile(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		protected bool
	}{
		{
			name:      "SESSION_CONTEXT.md is protected",
			filePath:  ".claude/sessions/session-123/SESSION_CONTEXT.md",
			protected: true,
		},
		{
			name:      "SPRINT_CONTEXT.md is protected",
			filePath:  ".claude/sprints/sprint-1/SPRINT_CONTEXT.md",
			protected: true,
		},
		{
			name:      "absolute path SESSION_CONTEXT.md",
			filePath:  "/home/user/project/.claude/sessions/foo/SESSION_CONTEXT.md",
			protected: true,
		},
		{
			name:      "regular file is not protected",
			filePath:  "src/main.go",
			protected: false,
		},
		{
			name:      "CLAUDE.md is not protected",
			filePath:  ".claude/CLAUDE.md",
			protected: false,
		},
		{
			name:      "random md file not protected",
			filePath:  "docs/SESSION_CONTEXT_OLD.md",
			protected: false,
		},
		{
			name:      "file ending with SESSION_CONTEXT.md",
			filePath:  "test_SESSION_CONTEXT.md",
			protected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isProtectedFile(tt.filePath)
			if result != tt.protected {
				t.Errorf("isProtectedFile(%q) = %v, want %v", tt.filePath, result, tt.protected)
			}
		})
	}
}

func TestParseFilePath(t *testing.T) {
	tests := []struct {
		name          string
		toolInput     string
		want          string
		expectWarning bool
	}{
		{
			name:          "valid Write input",
			toolInput:     `{"file_path": "/tmp/test.txt", "content": "hello"}`,
			want:          "/tmp/test.txt",
			expectWarning: false,
		},
		{
			name:          "valid Edit input",
			toolInput:     `{"file_path": "/tmp/test.txt", "old_string": "hello", "new_string": "world"}`,
			want:          "/tmp/test.txt",
			expectWarning: false,
		},
		{
			name:          "empty input",
			toolInput:     "",
			want:          "",
			expectWarning: false,
		},
		{
			name:          "invalid JSON",
			toolInput:     "not json",
			want:          "",
			expectWarning: true,
		},
		{
			name:          "malformed JSON - unterminated string",
			toolInput:     `{"file_path": "/tmp/test.txt`,
			want:          "",
			expectWarning: true,
		},
		{
			name:          "malformed JSON - invalid escape",
			toolInput:     `{"file_path": "\x"}`,
			want:          "",
			expectWarning: true,
		},
		{
			name:          "no file_path field",
			toolInput:     `{"other": "value"}`,
			want:          "",
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, true) // verbose=true to capture warnings

			result := parseFilePath(printer, tt.toolInput)
			if result != tt.want {
				t.Errorf("parseFilePath(%q) = %q, want %q", tt.toolInput, result, tt.want)
			}

			// Check if warning was logged when expected
			if tt.expectWarning {
				stderrStr := stderr.String()
				if !bytes.Contains([]byte(stderrStr), []byte("failed to parse tool input JSON")) {
					t.Errorf("Expected warning log for malformed JSON, but got no warning. stderr: %s", stderrStr)
				}
			}
		})
	}
}

func TestRunWriteguard_EarlyExit_HooksDisabled(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestRunWriteguard_BypassEnvVar(t *testing.T) {
	// This test has been updated to verify that writes to protected files
	// are blocked when no Moirai lock is held. The old MOIRAI_BYPASS env var
	// mechanism has been replaced with a lock file check.
	//
	// For a test of the lock bypass mechanism, see lock_test.go which tests
	// the full lock acquisition and release flow.

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Write",
		ToolInput:   `{"file_path": ".claude/sessions/test/SESSION_CONTEXT.md"}`,
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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	// Without a valid Moirai lock, write should be denied
	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q (no lock should deny)", result.HookSpecificOutput.PermissionDecision, "deny")
	}
}

func TestRunWriteguard_NonWriteTool(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "ls"}`,
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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q (non-write tool should allow)", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestRunWriteguard_BlockSessionContext(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Write",
		ToolInput:   `{"file_path": ".claude/sessions/test/SESSION_CONTEXT.md", "content": "bad"}`,
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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "deny")
	}
	if result.HookSpecificOutput.PermissionDecisionReason == "" {
		t.Error("Reason should not be empty for blocked write")
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("Moirai")) {
		t.Errorf("Reason should mention Moirai, got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestRunWriteguard_BlockSprintContext(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Edit",
		ToolInput:   `{"file_path": ".claude/sprints/current/SPRINT_CONTEXT.md", "old_string": "x", "new_string": "y"}`,
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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "deny")
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("SPRINT_CONTEXT")) {
		t.Errorf("Reason should mention SPRINT_CONTEXT, got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestRunWriteguard_AllowRegularFile(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Write",
		ToolInput:   `{"file_path": "src/main.go", "content": "package main"}`,
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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q (regular file should be allowed)", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

// BenchmarkWriteguardHook_Passthrough benchmarks the passthrough path (<5ms target).
func BenchmarkWriteguardHook_Passthrough(b *testing.B) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Write")
	os.Setenv("CLAUDE_TOOL_INPUT", `{"file_path": "src/main.go", "content": "x"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_TOOL_NAME")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
	}()

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
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runWriteguardCore(ctx, printer)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Passthrough took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkWriteguardHook_EarlyExit benchmarks early exit when disabled.
func BenchmarkWriteguardHook_EarlyExit(b *testing.B) {
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
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runWriteguardCore(ctx, printer)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Early exit took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

func TestWriteguard_StdinIntegration_AllowRegularFile(t *testing.T) {
	// Test that the full production path works with stdin JSON
	// This test verifies the fix for the env var vs stdin bug

	// Save and restore original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create pipe with CC-format JSON
	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"src/main.go","content":"package main"},"session_id":"test-session"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestWriteguard_StdinIntegration_BlockProtectedFile(t *testing.T) {
	// Test that the full production path blocks SESSION_CONTEXT.md via stdin

	// Save and restore original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create pipe with CC-format JSON targeting SESSION_CONTEXT.md
	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":".claude/sessions/test-session/SESSION_CONTEXT.md","content":"bad"},"session_id":"test-session"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

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
	}

	err := runWriteguardCore(ctx, printer)
	if err != nil {
		t.Fatalf("runWriteguard() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q (stdin should block protected file)", result.HookSpecificOutput.PermissionDecision, "deny")
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("SESSION_CONTEXT")) {
		t.Errorf("Reason should mention SESSION_CONTEXT, got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}


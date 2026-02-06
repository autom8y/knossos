package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name          string
		toolInput     string
		want          string
		expectWarning bool
	}{
		{
			name:          "valid Bash input",
			toolInput:     `{"command": "ls -la", "description": "List files"}`,
			want:          "ls -la",
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
			name:          "malformed JSON - unterminated",
			toolInput:     `{"command": "ls -la"`,
			want:          "",
			expectWarning: true,
		},
		{
			name:          "malformed JSON - trailing comma",
			toolInput:     `{"command": "ls", }`,
			want:          "",
			expectWarning: true,
		},
		{
			name:          "no command field",
			toolInput:     `{"description": "just a description"}`,
			want:          "",
			expectWarning: false,
		},
		{
			name:          "complex command",
			toolInput:     `{"command": "git push --force origin main", "description": "Force push"}`,
			want:          "git push --force origin main",
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, true) // verbose=true to capture warnings

			result := parseCommand(printer, tt.toolInput)
			if result != tt.want {
				t.Errorf("parseCommand(%q) = %q, want %q", tt.toolInput, result, tt.want)
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

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		blocked bool
		reason  string
	}{
		// Safe commands
		{
			name:    "simple ls",
			command: "ls -la",
			blocked: false,
		},
		{
			name:    "git status",
			command: "git status",
			blocked: false,
		},
		{
			name:    "git push to feature branch",
			command: "git push origin feature/my-feature",
			blocked: false,
		},
		{
			name:    "rm regular file",
			command: "rm -f /tmp/test.txt",
			blocked: false,
		},
		{
			name:    "rm -rf non-protected path",
			command: "rm -rf /tmp/build",
			blocked: false,
		},

		// Blocked: rm -rf on protected paths
		{
			name:    "rm -rf .git",
			command: "rm -rf .git",
			blocked: true,
			reason:  "Cannot rm -rf protected path: .git",
		},
		{
			name:    "rm -rf .claude",
			command: "rm -rf .claude/",
			blocked: true,
			reason:  "Cannot rm -rf protected path: .claude",
		},
		{
			name:    "rm -rf .github",
			command: "rm -rf .github",
			blocked: true,
			reason:  "Cannot rm -rf protected path: .github",
		},
		{
			name:    "rm -rf node_modules",
			command: "rm -rf node_modules",
			blocked: true,
			reason:  "Cannot rm -rf protected path: node_modules",
		},
		{
			name:    "rm -fr variant",
			command: "rm -fr .git",
			blocked: true,
			reason:  "Cannot rm -rf protected path: .git",
		},

		// Blocked: force push to main/master
		{
			name:    "git push --force origin main",
			command: "git push --force origin main",
			blocked: true,
			reason:  "Force push to main/master is blocked. Use --force-with-lease or push to a feature branch.",
		},
		{
			name:    "git push -f origin master",
			command: "git push -f origin master",
			blocked: true,
			reason:  "Force push to main/master is blocked. Use --force-with-lease or push to a feature branch.",
		},
		{
			name:    "git push --force to feature branch allowed",
			command: "git push --force origin feature/test",
			blocked: false,
		},

		// Blocked: --no-verify
		{
			name:    "git commit --no-verify",
			command: "git commit -m 'test' --no-verify",
			blocked: true,
			reason:  "Skipping hooks with --no-verify is blocked. Pre-commit hooks exist for a reason.",
		},
		{
			name:    "git push --no-verify",
			command: "git push --no-verify",
			blocked: true,
			reason:  "Skipping hooks with --no-verify is blocked. Pre-commit hooks exist for a reason.",
		},

		// Blocked: reset --hard
		{
			name:    "git reset --hard",
			command: "git reset --hard HEAD~1",
			blocked: true,
			reason:  "git reset --hard is blocked. Use git stash or git checkout for safer alternatives.",
		},
		{
			name:    "git reset --hard HEAD",
			command: "git reset --hard HEAD",
			blocked: true,
			reason:  "git reset --hard is blocked. Use git stash or git checkout for safer alternatives.",
		},
		{
			name:    "git reset --soft allowed",
			command: "git reset --soft HEAD~1",
			blocked: false,
		},

		// Blocked: git clean -fd
		{
			name:    "git clean -fd",
			command: "git clean -fd",
			blocked: true,
			reason:  "git clean -fd is blocked on protected branches. Use git stash or manual cleanup.",
		},
		{
			name:    "git clean -df variant",
			command: "git clean -df",
			blocked: true,
			reason:  "git clean -fd is blocked on protected branches. Use git stash or manual cleanup.",
		},
		{
			name:    "git clean -n allowed (dry run)",
			command: "git clean -n",
			blocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, reason := validateCommand(tt.command)
			if blocked != tt.blocked {
				t.Errorf("validateCommand(%q) blocked = %v, want %v", tt.command, blocked, tt.blocked)
			}
			if tt.blocked && reason != tt.reason {
				t.Errorf("validateCommand(%q) reason = %q, want %q", tt.command, reason, tt.reason)
			}
		})
	}
}

func TestRunValidate_EarlyExit_HooksDisabled(t *testing.T) {
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", result.Decision, "allow")
	}
}

func TestRunValidate_BypassEnvVar(t *testing.T) {
	env := testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "rm -rf .git"}`,
	})
	env.SetVar(ValidateBypassEnvVar, "1")

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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.Decision != "allow" {
		t.Errorf("Decision = %q, want %q (bypass should allow)", result.Decision, "allow")
	}
}

func TestRunValidate_NonBashTool(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Write",
		ToolInput:   `{"file_path": "/tmp/test.txt", "content": "hello"}`,
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.Decision != "allow" {
		t.Errorf("Decision = %q, want %q (non-Bash tool should allow)", result.Decision, "allow")
	}
}

func TestRunValidate_AllowSafeCommand(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "ls -la", "description": "List files"}`,
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", result.Decision, "allow")
	}
}

func TestRunValidate_BlockRmRfGit(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "rm -rf .git"}`,
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Decision != "block" {
		t.Errorf("Decision = %q, want %q", result.Decision, "block")
	}
	if result.Reason == "" {
		t.Error("Reason should not be empty for blocked command")
	}
	if !bytes.Contains([]byte(result.Reason), []byte(".git")) {
		t.Errorf("Reason should mention .git, got: %q", result.Reason)
	}
}

func TestRunValidate_BlockForcePush(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "git push --force origin main"}`,
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Decision != "block" {
		t.Errorf("Decision = %q, want %q", result.Decision, "block")
	}
	if !bytes.Contains([]byte(result.Reason), []byte("Force push")) {
		t.Errorf("Reason should mention Force push, got: %q", result.Reason)
	}
}

func TestRunValidate_BlockNoVerify(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "git commit -m 'fix' --no-verify"}`,
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Decision != "block" {
		t.Errorf("Decision = %q, want %q", result.Decision, "block")
	}
	if !bytes.Contains([]byte(result.Reason), []byte("--no-verify")) {
		t.Errorf("Reason should mention --no-verify, got: %q", result.Reason)
	}
}

func TestRunValidate_BlockResetHard(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "git reset --hard HEAD~1"}`,
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Decision != "block" {
		t.Errorf("Decision = %q, want %q", result.Decision, "block")
	}
	if !bytes.Contains([]byte(result.Reason), []byte("reset --hard")) {
		t.Errorf("Reason should mention reset --hard, got: %q", result.Reason)
	}
}

func TestRunValidate_BlockCleanFd(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   `{"command": "git clean -fd"}`,
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

	err := runValidateCore(ctx, printer, "")
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Decision != "block" {
		t.Errorf("Decision = %q, want %q", result.Decision, "block")
	}
	if !bytes.Contains([]byte(result.Reason), []byte("clean -fd")) {
		t.Errorf("Reason should mention clean -fd, got: %q", result.Reason)
	}
}

func TestRunValidate_StdinInput(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PreToolUse",
		ToolName:    "Bash",
		ToolInput:   "", // Empty env, will use stdin
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

	// Simulate stdin input with dangerous command
	stdinInput := `{"command": "rm -rf .git"}`
	err := runValidateCore(ctx, printer, stdinInput)
	if err != nil {
		t.Fatalf("runValidate() error = %v", err)
	}

	var result ValidateDecision
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.Decision != "block" {
		t.Errorf("Decision = %q, want %q (stdin should work)", result.Decision, "block")
	}
}

// BenchmarkValidateHook_Passthrough benchmarks the passthrough path (<5ms target).
func BenchmarkValidateHook_Passthrough(b *testing.B) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Bash")
	os.Setenv("CLAUDE_TOOL_INPUT", `{"command": "ls -la", "description": "List files"}`)
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
		runValidateCore(ctx, printer, "")
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Passthrough took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkValidateHook_EarlyExit benchmarks early exit when disabled.
func BenchmarkValidateHook_EarlyExit(b *testing.B) {
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
		runValidateCore(ctx, printer, "")
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Early exit took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkValidateHook_Validation benchmarks the full validation path.
func BenchmarkValidateHook_Validation(b *testing.B) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Bash")
	os.Setenv("CLAUDE_TOOL_INPUT", `{"command": "git push --force origin main"}`)
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
		runValidateCore(ctx, printer, "")
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Validation took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

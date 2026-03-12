package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"
)

func TestExtractCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{
			name:    "double-quoted message",
			command: `git commit -m "feat(auth): add login"`,
			want:    "feat(auth): add login",
		},
		{
			name:    "single-quoted message",
			command: `git commit -m 'fix: resolve race condition'`,
			want:    "fix: resolve race condition",
		},
		{
			name:    "no -m flag",
			command: "git commit",
			want:    "",
		},
		{
			name:    "amend without message",
			command: "git commit --amend",
			want:    "",
		},
		{
			name:    "message with flags before -m",
			command: `git commit --all -m "chore: update deps"`,
			want:    "chore: update deps",
		},
		{
			name:    "not a git command",
			command: "ls -la",
			want:    "",
		},
		{
			name:    "empty double-quoted message",
			command: `git commit -m ""`,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCommitMessage(tt.command)
			if got != tt.want {
				t.Errorf("extractCommitMessage(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestGitConventionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		command string
		allow   bool
	}{
		// Fast-path: non-git commands
		{
			name:    "simple ls",
			command: "ls -la",
			allow:   true,
		},
		{
			name:    "npm install",
			command: "npm install express",
			allow:   true,
		},

		// Fast-path: non-commit git commands
		{
			name:    "git push",
			command: "git push origin main",
			allow:   true,
		},
		{
			name:    "git status",
			command: "git status",
			allow:   true,
		},
		{
			name:    "git log",
			command: "git log --oneline -5",
			allow:   true,
		},

		// Valid conventional commits
		{
			name:    "feat with scope",
			command: `git commit -m "feat(auth): add JWT refresh token support"`,
			allow:   true,
		},
		{
			name:    "fix with scope",
			command: `git commit -m "fix(api): handle timeout on slow connections"`,
			allow:   true,
		},
		{
			name:    "chore without scope",
			command: `git commit -m "chore: update dependencies"`,
			allow:   true,
		},
		{
			name:    "docs with scope",
			command: `git commit -m "docs(readme): add installation instructions"`,
			allow:   true,
		},
		{
			name:    "refactor with scope",
			command: `git commit -m "refactor(core): extract validation logic"`,
			allow:   true,
		},
		{
			name:    "test with scope",
			command: `git commit -m "test(auth): add login flow tests"`,
			allow:   true,
		},
		{
			name:    "perf with scope",
			command: `git commit -m "perf(query): add index for user lookup"`,
			allow:   true,
		},
		{
			name:    "ci without scope",
			command: `git commit -m "ci: add GitHub Actions workflow"`,
			allow:   true,
		},
		{
			name:    "build with scope",
			command: `git commit -m "build(webpack): optimize bundle size"`,
			allow:   true,
		},
		{
			name:    "style without scope",
			command: `git commit -m "style: fix formatting"`,
			allow:   true,
		},
		{
			name:    "single-quoted valid",
			command: `git commit -m 'feat: add new feature'`,
			allow:   true,
		},

		// Releaser-specific formats (should pass)
		{
			name:    "chore deps bump",
			command: `git commit -m "chore(deps): bump @autom8y/sdk to 2.1.0"`,
			allow:   true,
		},
		{
			name:    "chore release publish",
			command: `git commit -m "chore(release): publish core-sdk v1.3.0"`,
			allow:   true,
		},

		// Invalid: no conventional type
		{
			name:    "plain message",
			command: `git commit -m "updated stuff"`,
			allow:   false,
		},
		{
			name:    "capital letter start",
			command: `git commit -m "Add new feature"`,
			allow:   false,
		},
		{
			name:    "missing colon",
			command: `git commit -m "feat add something"`,
			allow:   false,
		},
		{
			name:    "missing space after colon",
			command: `git commit -m "feat:add something"`,
			allow:   false,
		},
		{
			name:    "invalid type",
			command: `git commit -m "feature(auth): add login"`,
			allow:   false,
		},
		{
			name:    "WIP commit",
			command: `git commit -m "WIP"`,
			allow:   false,
		},
		{
			name:    "just a hash reference",
			command: `git commit -m "fix #123"`,
			allow:   false,
		},

		// Edge cases: allow
		{
			name:    "interactive commit no -m",
			command: "git commit",
			allow:   true,
		},
		{
			name:    "amend commit",
			command: "git commit --amend",
			allow:   true,
		},
		{
			name:    "amend with message",
			command: `git commit --amend -m "bad message format"`,
			allow:   true,
		},
		{
			name:    "heredoc message",
			command: `git commit -m "$(cat <<'EOF'` + "\nfeat: something\nEOF\n)\"",
			allow:   true,
		},
		{
			name:    "commit with --all flag",
			command: `git commit --all -m "feat: add feature with all staged"`,
			allow:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:     "PreToolUse",
				ToolName:  "Bash",
				ToolInput: `{"command": ` + jsonEscape(tt.command) + `}`,
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

			err := runGitConventionsCore(nil, ctx, printer)
			if err != nil {
				t.Fatalf("runGitConventionsCore() error = %v", err)
			}

			var result hook.PreToolUseOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
			}

			got := result.HookSpecificOutput.PermissionDecision
			if tt.allow && got != "allow" {
				t.Errorf("command %q: PermissionDecision = %q, want %q", tt.command, got, "allow")
			}
			if !tt.allow && got != "deny" {
				t.Errorf("command %q: PermissionDecision = %q, want %q", tt.command, got, "deny")
			}

			// Verify deny messages reference the skill
			if !tt.allow {
				reason := result.HookSpecificOutput.PermissionDecisionReason
				if !bytes.Contains([]byte(reason), []byte("commit:behavior")) {
					t.Errorf("Deny reason should reference commit:behavior skill, got: %q", reason)
				}
			}
		})
	}
}

func TestGitConventions_NonBashTool(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path": "/tmp/test.txt", "content": "hello"}`,
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

	err := runGitConventionsCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runGitConventionsCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q (non-Bash tool should allow)", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestGitConventions_StdinIntegration_AllowValid(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m \"feat(auth): add login flow\""},"session_id":"test-session"}`
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

	err := runGitConventionsCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runGitConventionsCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestGitConventions_StdinIntegration_DenyInvalid(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m \"updated some stuff\""},"session_id":"test-session"}`
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

	err := runGitConventionsCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runGitConventionsCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "deny")
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("commit:behavior")) {
		t.Errorf("Reason should reference commit:behavior, got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// BenchmarkGitConventions_FastPath benchmarks the non-git-commit fast path.
func BenchmarkGitConventions_FastPath(b *testing.B) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Bash")
	os.Setenv("CLAUDE_TOOL_INPUT", `{"command": "ls -la"}`)
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
		runGitConventionsCore(nil, ctx, printer)
	}
}

// jsonEscape wraps a string in JSON-safe quotes for embedding in tool_input.
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

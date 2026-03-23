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

func TestAttributionGuard_PatternDetection(t *testing.T) {
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

		// Clean commits — should allow
		{
			name:    "clean commit with scope",
			command: `git commit -m "feat(auth): add JWT refresh token support"`,
			allow:   true,
		},
		{
			name:    "clean commit without scope",
			command: `git commit -m "fix: resolve race condition"`,
			allow:   true,
		},
		{
			name:    "interactive commit no -m",
			command: "git commit",
			allow:   true,
		},
		{
			name:    "clean heredoc commit",
			command: "git commit -m \"$(cat <<'EOF'\nfeat(auth): add login flow\n\n- Implement token refresh\n- Add middleware\nEOF\n)\"",
			allow:   true,
		},

		// Co-Authored-By variants — should deny
		{
			name:    "Co-Authored-By standard format",
			command: `git commit -m "feat: add feature` + "\n\n" + `Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"`,
			allow:   false,
		},
		{
			name:    "co-authored-by lowercase",
			command: `git commit -m "fix: resolve bug` + "\n\n" + `co-authored-by: claude"`,
			allow:   false,
		},
		{
			name:    "Co-authored-by title case",
			command: `git commit -m "chore: update deps` + "\n\n" + `Co-authored-by: AI Assistant"`,
			allow:   false,
		},
		{
			name:    "CO-AUTHORED-BY all caps",
			command: `git commit -m "docs: update readme` + "\n\n" + `CO-AUTHORED-BY: Claude"`,
			allow:   false,
		},
		{
			name:    "Co-Authored-By with trailer flag",
			command: `git commit -m "feat: add feature" --trailer "Co-Authored-By: Claude <noreply@anthropic.com>"`,
			allow:   false,
		},

		// Generated with variants — should deny
		{
			name:    "Generated with Claude Code markdown",
			command: `git commit -m "feat: add feature` + "\n\n" + `Generated with [Claude Code](https://claude.com/claude-code)"`,
			allow:   false,
		},
		{
			name:    "generated with AI plain",
			command: `git commit -m "fix: bug` + "\n\n" + `generated with AI assistance"`,
			allow:   false,
		},
		{
			name:    "Generated with Anthropic",
			command: `git commit -m "chore: update` + "\n\n" + `Generated with Anthropic Claude"`,
			allow:   false,
		},

		// anthropic.com email — should deny
		{
			name:    "noreply@anthropic.com in message",
			command: `git commit -m "feat: add feature` + "\n\n" + `Author: Claude <noreply@anthropic.com>"`,
			allow:   false,
		},

		// Heredoc antipattern — THE primary case to catch
		{
			name: "heredoc with Co-Authored-By — the primary antipattern",
			command: `git commit -m "$(cat <<'EOF'
fix(ws2): wire RealTerraformRunner

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"`,
			allow: false,
		},
		{
			name: "heredoc with Generated with footer",
			command: `git commit -m "$(cat <<'EOF'
feat(api): add endpoint

Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"`,
			allow: false,
		},
		{
			name: "heredoc clean — no markers",
			command: `git commit -m "$(cat <<'EOF'
feat(api): add endpoint

- Implement REST handler
- Add validation
EOF
)"`,
			allow: true,
		},

		// Amend with markers — still blocked (attribution guard does not skip amend)
		{
			name:    "amend with Co-Authored-By",
			command: `git commit --amend -m "feat: add feature` + "\n\n" + `Co-Authored-By: Claude"`,
			allow:   false,
		},

		// Edge: git add && git commit chain
		{
			name:    "chained command with Co-Authored-By",
			command: `git add . && git commit -m "feat: feature` + "\n\n" + `Co-Authored-By: Claude"`,
			allow:   false,
		},
		{
			name:    "chained command clean",
			command: `git add . && git commit -m "feat: feature"`,
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

			err := runAttributionGuardCore(nil, ctx, printer)
			if err != nil {
				t.Fatalf("runAttributionGuardCore() error = %v", err)
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

			// Verify deny messages reference conventions
			if !tt.allow {
				reason := result.HookSpecificOutput.PermissionDecisionReason
				if !bytes.Contains([]byte(reason), []byte("conventions")) {
					t.Errorf("Deny reason should reference conventions skill, got: %q", reason)
				}
			}
		})
	}
}

func TestAttributionGuard_NonBashTool(t *testing.T) {
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

	err := runAttributionGuardCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAttributionGuardCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q (non-Bash tool should allow)", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestAttributionGuard_StdinIntegration_DenyCoAuthored(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m \"feat: add feature\n\nCo-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>\""},"session_id":"test-session"}`
	go func() {
		_, _ = w.Write([]byte(payload))
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

	err := runAttributionGuardCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAttributionGuardCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "deny")
	}
}

func TestAttributionGuard_StdinIntegration_AllowClean(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m \"feat(auth): add login flow\""},"session_id":"test-session"}`
	go func() {
		_, _ = w.Write([]byte(payload))
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

	err := runAttributionGuardCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAttributionGuardCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

// BenchmarkAttributionGuard_FastPath benchmarks the non-git-commit fast path.
func BenchmarkAttributionGuard_FastPath(b *testing.B) {
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
		_ = runAttributionGuardCore(nil, ctx, printer)
	}
}

// BenchmarkAttributionGuard_CommitScan benchmarks scanning a git commit command.
func BenchmarkAttributionGuard_CommitScan(b *testing.B) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Bash")
	os.Setenv("CLAUDE_TOOL_INPUT", `{"command": "git commit -m \"feat(auth): add JWT refresh token support\n\n- Implement token rotation on expiry\n- Add configurable expiration window\""}`)
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
		_ = runAttributionGuardCore(nil, ctx, printer)
	}
}

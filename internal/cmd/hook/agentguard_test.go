package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"
)

// makeAgentGuardCtx builds a minimal cmdContext for agent-guard tests.
func makeAgentGuardCtx() *cmdContext {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	return &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
	}
}

// runAgentGuardTest executes runAgentGuardCore with a given env and returns the parsed output.
func runAgentGuardTest(t *testing.T, env *testutil.HookEnv, agentName string, allowPaths []string) hook.PreToolUseOutput {
	t.Helper()
	testutil.SetupEnv(t, env)
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runAgentGuardCore(makeAgentGuardCtx(), printer, agentName, allowPaths); err != nil {
		t.Fatalf("runAgentGuardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}
	return result
}

// AG-1: Write to a path outside all allowed prefixes -- expect DENY.
func TestAgentGuard_DenyOutsideAllowedPaths(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"src/main.go","content":"package main"}`,
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny (path outside allowed prefix)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-2: Write to .sos/wip/ relative path -- expect ALLOW.
func TestAgentGuard_AllowWipPath(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":".sos/wip/gap.md","content":"# Gap Analysis"}`,
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (.sos/wip/ prefix matches)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-3: Write to docs/ecosystem/ -- expect ALLOW (second prefix matches).
func TestAgentGuard_AllowDocsEcosystem(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"docs/ecosystem/GAP-x.md","content":"# Doc"}`,
	}, "ecosystem-analyst", []string{".sos/wip/", "docs/ecosystem/"})

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (docs/ecosystem/ prefix matches)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-4: Write to internal/ path with two allow prefixes -- expect DENY.
func TestAgentGuard_DenyInternalPath(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"internal/agent/f.go","content":"package agent"}`,
	}, "ecosystem-analyst", []string{".sos/wip/", "docs/ecosystem/"})

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny (internal/ not in allowed paths)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-5: Non-PreToolUse event (PostToolUse) -- expect ALLOW passthrough.
func TestAgentGuard_AllowNonWriteEvent(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PostToolUse",
		ToolName:  "Read",
		ToolInput: `{"file_path":"anything"}`,
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (PostToolUse passes through)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-6: Well-formed JSON but file_path field is absent -- expect DENY (fail closed).
func TestAgentGuard_DenyMissingFilePath(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"content":"some content"}`,
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny (well-formed JSON with no file_path must fail closed)", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("missing")) {
		t.Errorf("Reason should mention 'missing', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// AG-7: Corrupt/unparseable JSON stdin -- expect ALLOW (graceful degradation).
func TestAgentGuard_AllowBadJSON(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `not valid json {{{`,
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (corrupt JSON must degrade gracefully)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-8: No --allow-path flags configured -- expect DENY (unconditional).
func TestAgentGuard_DenyNoAllowPaths(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"src/main.go","content":"package main"}`,
	}, "ecosystem-analyst", nil)

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny (no allow paths = unconditional deny)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-9: Absolute path containing /.sos/wip/ -- expect ALLOW (contains condition).
func TestAgentGuard_AllowAbsoluteWipPath(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"/Users/tom/project/.sos/wip/file.md","content":"# Notes"}`,
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (absolute path /.sos/wip/ matches)", result.HookSpecificOutput.PermissionDecision)
	}
}

// AG-10: Deny reason must include the agent name from --agent flag.
func TestAgentGuard_DenyReasonIncludesAgentName(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"src/main.go","content":"package main"}`,
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("ecosystem-analyst")) {
		t.Errorf("Reason should contain agent name 'ecosystem-analyst', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// AG-11: Default agent name "this agent" used when no --agent flag specified.
func TestAgentGuard_DefaultAgentName(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"src/main.go","content":"package main"}`,
	}, "this agent", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("this agent")) {
		t.Errorf("Reason should contain 'this agent' (default), got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// AG-12: Non-PreToolUse event name (SessionStart) -- expect ALLOW passthrough.
func TestAgentGuard_NonPreToolUseEvent(t *testing.T) {
	result := runAgentGuardTest(t, &testutil.HookEnv{
		Event: "SessionStart",
	}, "ecosystem-analyst", []string{".sos/wip/"})

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (SessionStart passes through)", result.HookSpecificOutput.PermissionDecision)
	}
}

// --- isAllowedPath unit tests ---

func TestIsAllowedPath(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		allowPaths []string
		want       bool
	}{
		{
			name:       "relative .sos/wip/ prefix match",
			filePath:   ".sos/wip/gap.md",
			allowPaths: []string{".sos/wip/"},
			want:       true,
		},
		{
			name:       "absolute path /.sos/wip/ contains match",
			filePath:   "/Users/tom/project/.sos/wip/SPIKE-x.md",
			allowPaths: []string{".sos/wip/"},
			want:       true,
		},
		{
			name:       "docs/ecosystem/ prefix match",
			filePath:   "docs/ecosystem/GAP-x.md",
			allowPaths: []string{".sos/wip/", "docs/ecosystem/"},
			want:       true,
		},
		{
			name:       "internal/ does not match .sos/wip/ or docs/ecosystem/",
			filePath:   "internal/agent/f.go",
			allowPaths: []string{".sos/wip/", "docs/ecosystem/"},
			want:       false,
		},
		{
			name:       "empty allow paths",
			filePath:   ".sos/wip/gap.md",
			allowPaths: nil,
			want:       false,
		},
		{
			name:       "trailing slash prevents sibling match",
			filePath:   ".wip-private/secret.md",
			allowPaths: []string{".sos/wip/"},
			want:       false,
		},
		{
			name:       "docs/ecosystem-old does not match docs/ecosystem/",
			filePath:   "docs/ecosystem-old/notes.md",
			allowPaths: []string{"docs/ecosystem/"},
			want:       false,
		},
		{
			name:       "relative .sos/wip/ matches prefix",
			filePath:   ".sos/wip/report.md",
			allowPaths: []string{".sos/wip/"},
			want:       true,
		},
		{
			name:       "absolute .sos/wip/ matches via contains",
			filePath:   "/Users/tom/project/.sos/wip/report.md",
			allowPaths: []string{".sos/wip/"},
			want:       true,
		},
		{
			name:       "deep nested wip/ matches via contains",
			filePath:   "deep/path/wip/report.md",
			allowPaths: []string{"wip/"},
			want:       true,
		},
		{
			name:       ".ledge/ does not match .sos/wip/",
			filePath:   ".ledge/specs/PRD-auth.md",
			allowPaths: []string{".sos/wip/"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedPath(tt.filePath, tt.allowPaths)
			if got != tt.want {
				t.Errorf("isAllowedPath(%q, %v) = %v, want %v", tt.filePath, tt.allowPaths, got, tt.want)
			}
		})
	}
}

// --- parseFilePathStrict unit tests ---

func TestParseFilePathStrict(t *testing.T) {
	tests := []struct {
		name           string
		toolInput      string
		wantPath       string
		wantParseError bool
	}{
		{
			name:           "valid Write input with file_path",
			toolInput:      `{"file_path": "src/main.go", "content": "x"}`,
			wantPath:       "src/main.go",
			wantParseError: false,
		},
		{
			name:           "valid JSON but file_path absent",
			toolInput:      `{"content": "some content"}`,
			wantPath:       "",
			wantParseError: false,
		},
		{
			name:           "corrupt JSON returns parse error",
			toolInput:      `not json {{`,
			wantPath:       "",
			wantParseError: true,
		},
		{
			name:           "empty toolInput returns parse error",
			toolInput:      "",
			wantPath:       "",
			wantParseError: true,
		},
		{
			name:           "unterminated JSON returns parse error",
			toolInput:      `{"file_path": "foo`,
			wantPath:       "",
			wantParseError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, true)
			gotPath, gotParseError := parseFilePathStrict(printer, tt.toolInput)
			if gotPath != tt.wantPath {
				t.Errorf("parseFilePathStrict(%q) path = %q, want %q", tt.toolInput, gotPath, tt.wantPath)
			}
			if gotParseError != tt.wantParseError {
				t.Errorf("parseFilePathStrict(%q) parseError = %v, want %v", tt.toolInput, gotParseError, tt.wantParseError)
			}
		})
	}
}

// --- Stdin integration tests ---

// TestAgentGuard_StdinIntegration_DenyOutsidePaths verifies the full production
// path denies writes to paths outside allowed prefixes when CC sends JSON via stdin.
func TestAgentGuard_StdinIntegration_DenyOutsidePaths(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"internal/agent/frontmatter.go"},"session_id":"test-session"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	err := runAgentGuardCore(makeAgentGuardCtx(), printer, "ecosystem-analyst", []string{".sos/wip/", "docs/ecosystem/"})
	if err != nil {
		t.Fatalf("runAgentGuardCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny (stdin: outside allowed paths)", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("ecosystem-analyst")) {
		t.Errorf("Reason should contain agent name, got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// TestAgentGuard_StdinIntegration_AllowWipPath verifies the full production
// path allows writes to .sos/wip/ when CC sends JSON via stdin.
func TestAgentGuard_StdinIntegration_AllowWipPath(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":".sos/wip/gap.md","content":"# Gap"},"session_id":"test-session"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	err := runAgentGuardCore(makeAgentGuardCtx(), printer, "ecosystem-analyst", []string{".sos/wip/", "docs/ecosystem/"})
	if err != nil {
		t.Fatalf("runAgentGuardCore() error = %v", err)
	}

	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (stdin: .sos/wip/ prefix matches)", result.HookSpecificOutput.PermissionDecision)
	}
}

// BenchmarkAgentGuard_Passthrough benchmarks the allow path (<5ms target).
func BenchmarkAgentGuard_Passthrough(b *testing.B) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Write")
	os.Setenv("CLAUDE_TOOL_INPUT", `{"file_path":".sos/wip/gap.md","content":"x"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_TOOL_NAME")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
	}()

	ctx := makeAgentGuardCtx()
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	allowPaths := []string{".sos/wip/", "docs/ecosystem/"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runAgentGuardCore(ctx, printer, "ecosystem-analyst", allowPaths)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Passthrough took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

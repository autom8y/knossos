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

func TestExtractSessionIDFromPath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "standard session context path",
			filePath: ".claude/sessions/session-20260209-120000-abcdef01/SESSION_CONTEXT.md",
			want:     "session-20260209-120000-abcdef01",
		},
		{
			name:     "absolute path",
			filePath: "/home/user/project/.claude/sessions/session-20260209-120000-abcdef01/SESSION_CONTEXT.md",
			want:     "session-20260209-120000-abcdef01",
		},
		{
			name:     "sprint context in session dir",
			filePath: ".claude/sessions/session-20260209-120000-abcdef01/SPRINT_CONTEXT.md",
			want:     "session-20260209-120000-abcdef01",
		},
		{
			name:     "no session ID in path",
			filePath: "src/main.go",
			want:     "",
		},
		{
			name:     "session prefix too short",
			filePath: ".claude/sessions/session-short/SESSION_CONTEXT.md",
			want:     "",
		},
		{
			name:     "empty path",
			filePath: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSessionIDFromPath(tt.filePath)
			if got != tt.want {
				t.Errorf("extractSessionIDFromPath(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestWriteguard_ParkedSession_MoiraiLockAllow(t *testing.T) {
	// Simulate a PARKED session where resolveSession() returns empty
	// but the file path contains the session ID and a valid Moirai lock exists.
	sessionID := "session-20260209-120000-abcdef01"
	tmpDir := t.TempDir()

	// Create session directory with valid Moirai lock
	sessionDir := tmpDir + "/.claude/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lockData := `{"agent":"moirai","acquired_at":"` + time.Now().Format(time.RFC3339) + `","session_id":"` + sessionID + `","stale_after_seconds":300}`
	if err := os.WriteFile(sessionDir+"/.moirai-lock", []byte(lockData), 0o644); err != nil {
		t.Fatal(err)
	}

	filePath := ".claude/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Write",
		ToolInput:  `{"file_path":"` + filePath + `","content":"status: ACTIVE"}`,
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
		t.Errorf("PermissionDecision = %q, want %q (PARKED session with valid Moirai lock should allow)",
			result.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestWriteguard_ParkedSession_NoLock(t *testing.T) {
	// PARKED session without a Moirai lock should still be blocked.
	sessionID := "session-20260209-120000-abcdef01"
	tmpDir := t.TempDir()

	// Create session directory WITHOUT a lock file
	sessionDir := tmpDir + "/.claude/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	filePath := ".claude/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Write",
		ToolInput:  `{"file_path":"` + filePath + `","content":"status: ACTIVE"}`,
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
		t.Errorf("PermissionDecision = %q, want %q (PARKED session without lock should deny)",
			result.HookSpecificOutput.PermissionDecision, "deny")
	}
}

func TestWriteguard_ParkedSession_StaleLock(t *testing.T) {
	// PARKED session with a stale Moirai lock should be blocked.
	sessionID := "session-20260209-120000-abcdef01"
	tmpDir := t.TempDir()

	// Create session directory with stale Moirai lock (acquired 10 minutes ago, stale after 5 seconds)
	sessionDir := tmpDir + "/.claude/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	staleTime := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
	lockData := `{"agent":"moirai","acquired_at":"` + staleTime + `","session_id":"` + sessionID + `","stale_after_seconds":5}`
	if err := os.WriteFile(sessionDir+"/.moirai-lock", []byte(lockData), 0o644); err != nil {
		t.Fatal(err)
	}

	filePath := ".claude/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Write",
		ToolInput:  `{"file_path":"` + filePath + `","content":"status: ACTIVE"}`,
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
		t.Errorf("PermissionDecision = %q, want %q (PARKED session with stale lock should deny)",
			result.HookSpecificOutput.PermissionDecision, "deny")
	}
}

// --- .wip/ validation tests (W1-W12) ---

// makeWipCtx builds a minimal cmdContext for writeguard tests.
func makeWipCtx() *cmdContext {
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

// runWipTest executes runWriteguardCore with a given env and returns the parsed output.
func runWipTest(t *testing.T, env *testutil.HookEnv) hook.PreToolUseOutput {
	t.Helper()
	testutil.SetupEnv(t, env)
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(makeWipCtx(), printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}
	return result
}

// makeToolInput builds a JSON tool_input string with proper escaping.
// content is JSON-encoded so embedded newlines/quotes are safe in the JSON value.
func makeToolInput(filePath, content string) string {
	contentJSON, _ := json.Marshal(content)
	return `{"file_path":"` + filePath + `","content":` + string(contentJSON) + `}`
}

// W1: .wip/ write with valid frontmatter (spike) — allow, no additionalContext.
func TestWipWrite_ValidFrontmatter_Spike(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".wip/SPIKE-test.md", "---\ntype: spike\n---\n\n# Spike content"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W2: .wip/ write with valid frontmatter (design) — allow, no additionalContext.
func TestWipWrite_ValidFrontmatter_Design(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".wip/DESIGN-foo.md", "---\ntype: design\n---\n\n# Design doc"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W3: .wip/ write with valid frontmatter (scratch) — allow, no additionalContext.
func TestWipWrite_ValidFrontmatter_Scratch(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".wip/scratch-notes.md", "---\ntype: scratch\n---\n\nsome notes"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W4: .wip/ write with missing frontmatter — allow with advisory additionalContext.
func TestWipWrite_MissingFrontmatter(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".wip/BAD.md", "# No frontmatter here"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (must not block)", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.AdditionalContext), []byte("require YAML frontmatter")) {
		t.Errorf("AdditionalContext should mention 'require YAML frontmatter', got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W5: .wip/ write with frontmatter but missing type field — allow with advisory additionalContext.
func TestWipWrite_MissingTypeField(t *testing.T) {
	// JSON-encode the content so embedded newlines are valid JSON escape sequences.
	content := "---\nstatus: draft\n---\n\n# Content"
	contentJSON, _ := json.Marshal(content)
	toolInput := `{"file_path":".wip/BAD.md","content":` + string(contentJSON) + `}`
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: toolInput,
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.AdditionalContext), []byte("must include a type field")) {
		t.Errorf("AdditionalContext should mention 'must include a type field', got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W6: .wip/ write with invalid type value — allow with advisory additionalContext listing valid types.
func TestWipWrite_InvalidTypeValue(t *testing.T) {
	content := "---\ntype: memo\n---\n\n# Content"
	contentJSON, _ := json.Marshal(content)
	toolInput := `{"file_path":".wip/BAD.md","content":` + string(contentJSON) + `}`
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: toolInput,
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	ctx := result.HookSpecificOutput.AdditionalContext
	if !bytes.Contains([]byte(ctx), []byte("not valid")) {
		t.Errorf("AdditionalContext should mention 'not valid', got: %q", ctx)
	}
	if !bytes.Contains([]byte(ctx), []byte("spike")) {
		t.Errorf("AdditionalContext should list valid types including 'spike', got: %q", ctx)
	}
}

// W7: .wip/ Edit does NOT trigger validation — allow with no additionalContext.
func TestWipEdit_BypassesValidation(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Edit",
		ToolInput: `{"file_path":".wip/existing.md","old_string":"old","new_string":"new"}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (edit bypasses wip validation)", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for Edit tool bypass, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W8: Non-.wip/ Write is unchanged (regression) — allow, no additionalContext.
func TestWipWrite_NonWipPath_Unchanged(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"src/main.go","content":"package main"}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (non-wip path unchanged)", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for non-wip path, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W9: Protected file unchanged (regression) — deny.
func TestWipWrite_ProtectedFile_Unchanged(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":".claude/sessions/test/SESSION_CONTEXT.md","content":"bad"}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny (protected file unchanged)", result.HookSpecificOutput.PermissionDecision)
	}
}

// W10: .wip/ with absolute path and valid frontmatter — allow, no additionalContext.
func TestWipWrite_AbsolutePath_ValidFrontmatter(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput("/Users/tom/project/.wip/SPIKE-x.md", "---\ntype: spike\n---\n\n# Spike"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (absolute .wip/ path)", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W11: .wip/ write with empty content — allow with advisory additionalContext.
func TestWipWrite_EmptyContent(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":".wip/empty.md","content":""}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (empty content must not block)", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext == "" {
		t.Error("AdditionalContext should advise frontmatter for empty content")
	}
}

// TestIsWipPath covers the isWipPath helper directly.
func TestIsWipPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{".wip/SPIKE-foo.md", true},
		{".wip/scratch.md", true},
		{"/home/user/project/.wip/DESIGN-x.md", true},
		{"/Users/tom/.wip/bar.md", true},
		{"src/main.go", false},
		{".claude/sessions/test/SESSION_CONTEXT.md", false},
		{"docs/wip-notes.md", false}, // "wip" in name but not .wip/ dir
		{"", false},
	}
	for _, tt := range tests {
		got := isWipPath(tt.path)
		if got != tt.want {
			t.Errorf("isWipPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// TestValidateWipFrontmatter covers the validateWipFrontmatter helper directly.
func TestValidateWipFrontmatter(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantValid bool
		wantType  string
		wantReason string
	}{
		{
			name:      "valid spike",
			content:   "---\ntype: spike\n---\n\nbody",
			wantValid: true,
			wantType:  "spike",
		},
		{
			name:      "valid qa",
			content:   "---\ntype: qa\n---\n",
			wantValid: true,
			wantType:  "qa",
		},
		{
			name:       "missing frontmatter",
			content:    "# just markdown",
			wantValid:  false,
			wantReason: "require YAML frontmatter",
		},
		{
			name:       "empty content",
			content:    "",
			wantValid:  false,
			wantReason: "require YAML frontmatter",
		},
		{
			name:       "frontmatter without type",
			content:    "---\nstatus: draft\n---\n",
			wantValid:  false,
			wantReason: "must include a type field",
		},
		{
			name:       "invalid type",
			content:    "---\ntype: memo\n---\n",
			wantValid:  false,
			wantReason: "not valid",
		},
		{
			name:      "extra fields allowed",
			content:   "---\ntype: design\nsession: sess-123\nstatus: draft\n---\n",
			wantValid: true,
			wantType:  "design",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, typeVal, reason := validateWipFrontmatter(tt.content)
			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v (reason: %q)", valid, tt.wantValid, reason)
			}
			if tt.wantValid && typeVal != tt.wantType {
				t.Errorf("typeVal = %q, want %q", typeVal, tt.wantType)
			}
			if !tt.wantValid && tt.wantReason != "" {
				if !bytes.Contains([]byte(reason), []byte(tt.wantReason)) {
					t.Errorf("reason = %q, want it to contain %q", reason, tt.wantReason)
				}
			}
		})
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

// TestWriteguard_ArchivedSession_DeniesWithClearMessage verifies that writes to
// an archived session's context file are denied with a helpful "session is archived"
// message instead of the generic "Use Moirai" message.
func TestWriteguard_ArchivedSession_DeniesWithClearMessage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the archive directory with an archived session
	sessionID := "session-20250105-120030-arctest1"
	archiveSessionDir := tmpDir + "/.claude/.archive/sessions/" + sessionID
	if err := os.MkdirAll(archiveSessionDir, 0755); err != nil {
		t.Fatalf("Failed to create archive session dir: %v", err)
	}

	// Create an archived SESSION_CONTEXT.md in the archive
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ARCHIVED
---
`
	if err := os.WriteFile(archiveSessionDir+"/SESSION_CONTEXT.md", []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write archived context: %v", err)
	}

	// Simulate write attempt to the LIVE session path (which no longer exists after proper archiving)
	liveContextPath := ".claude/sessions/" + sessionID + "/SESSION_CONTEXT.md"

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path": "` + liveContextPath + `"}`,
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

	// Should deny
	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny", result.HookSpecificOutput.PermissionDecision)
	}

	// Should say "archived" not the generic Moirai delegation message
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("archived")) {
		t.Errorf("Reason should mention 'archived', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}


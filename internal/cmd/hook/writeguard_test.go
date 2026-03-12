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
			filePath:  ".sos/sessions/session-123/SESSION_CONTEXT.md",
			protected: true,
		},
		{
			name:      "SPRINT_CONTEXT.md is protected",
			filePath:  ".sos/sessions/session-123/sprints/sprint-1/SPRINT_CONTEXT.md",
			protected: true,
		},
		{
			name:      "absolute path SESSION_CONTEXT.md",
			filePath:  "/home/user/project/.sos/sessions/foo/SESSION_CONTEXT.md",
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
			toolInput:     `{"file_path": "tmp/test.txt", "content": "hello"}`,
			want:          "tmp/test.txt",
			expectWarning: false,
		},
		{
			name:          "valid Edit input",
			toolInput:     `{"file_path": "tmp/test.txt", "old_string": "hello", "new_string": "world"}`,
			want:          "tmp/test.txt",
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
		{
			name:          "path traversal - parent directory",
			toolInput:     `{"file_path": "../../etc/passwd"}`,
			want:          "",
			expectWarning: true,
		},
		{
			name:          "path traversal - absolute path",
			toolInput:     `{"file_path": "/etc/passwd"}`,
			want:          "",
			expectWarning: true,
		},
		{
			name:          "path traversal - indirect",
			toolInput:     `{"file_path": "foo/../../etc/passwd"}`,
			want:          "",
			expectWarning: true,
		},
		{
			name:          "normalized valid path",
			toolInput:     `{"file_path": "foo/./bar.txt"}`,
			want:          "foo/bar.txt",
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
				if !bytes.Contains([]byte(stderrStr), []byte("failed to parse tool input JSON")) &&
					!bytes.Contains([]byte(stderrStr), []byte("blocked potential path traversal attempt")) {
					t.Errorf("Expected warning log, but got no recognized warning. stderr: %s", stderrStr)
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
		ToolInput:   `{"file_path": ".sos/sessions/test/SESSION_CONTEXT.md"}`,
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
		ToolInput:   `{"file_path": ".sos/sessions/test/SESSION_CONTEXT.md", "content": "bad"}`,
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
		ToolInput:   `{"file_path": ".sos/sessions/session-abc/sprints/current/SPRINT_CONTEXT.md", "old_string": "x", "new_string": "y"}`,
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
			filePath: ".sos/sessions/session-20260209-120000-abcdef01/SESSION_CONTEXT.md",
			want:     "session-20260209-120000-abcdef01",
		},
		{
			name:     "absolute path",
			filePath: "/home/user/project/.sos/sessions/session-20260209-120000-abcdef01/SESSION_CONTEXT.md",
			want:     "session-20260209-120000-abcdef01",
		},
		{
			name:     "sprint context in session dir",
			filePath: ".sos/sessions/session-20260209-120000-abcdef01/SPRINT_CONTEXT.md",
			want:     "session-20260209-120000-abcdef01",
		},
		{
			name:     "no session ID in path",
			filePath: "src/main.go",
			want:     "",
		},
		{
			name:     "session prefix too short",
			filePath: ".sos/sessions/session-short/SESSION_CONTEXT.md",
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
	sessionDir := tmpDir + "/.sos/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lockData := `{"agent":"moirai","acquired_at":"` + time.Now().Format(time.RFC3339) + `","session_id":"` + sessionID + `","stale_after_seconds":300}`
	if err := os.WriteFile(sessionDir+"/.moirai-lock", []byte(lockData), 0o644); err != nil {
		t.Fatal(err)
	}

	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
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
	sessionDir := tmpDir + "/.sos/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
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
	sessionDir := tmpDir + "/.sos/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	staleTime := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
	lockData := `{"agent":"moirai","acquired_at":"` + staleTime + `","session_id":"` + sessionID + `","stale_after_seconds":5}`
	if err := os.WriteFile(sessionDir+"/.moirai-lock", []byte(lockData), 0o644); err != nil {
		t.Fatal(err)
	}

	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
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

// --- .sos/wip/ validation tests (W1-W12) ---

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

// W1: .sos/wip/ write with valid frontmatter (spike) — allow, no additionalContext.
func TestWipWrite_ValidFrontmatter_Spike(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".sos/wip/SPIKE-test.md", "---\ntype: spike\n---\n\n# Spike content"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W2: .sos/wip/ write with valid frontmatter (design) — allow, no additionalContext.
func TestWipWrite_ValidFrontmatter_Design(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".sos/wip/DESIGN-foo.md", "---\ntype: design\n---\n\n# Design doc"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W3: .sos/wip/ write with valid frontmatter (scratch) — allow, no additionalContext.
func TestWipWrite_ValidFrontmatter_Scratch(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".sos/wip/scratch-notes.md", "---\ntype: scratch\n---\n\nsome notes"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W4: .sos/wip/ write with missing frontmatter — allow with advisory additionalContext.
func TestWipWrite_MissingFrontmatter(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".sos/wip/BAD.md", "# No frontmatter here"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (must not block)", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.AdditionalContext), []byte("require YAML frontmatter")) {
		t.Errorf("AdditionalContext should mention 'require YAML frontmatter', got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W5: .sos/wip/ write with frontmatter but missing type field — allow with advisory additionalContext.
func TestWipWrite_MissingTypeField(t *testing.T) {
	// JSON-encode the content so embedded newlines are valid JSON escape sequences.
	content := "---\nstatus: draft\n---\n\n# Content"
	contentJSON, _ := json.Marshal(content)
	toolInput := `{"file_path":".sos/wip/BAD.md","content":` + string(contentJSON) + `}`
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

// W6: .sos/wip/ write with invalid type value — allow with advisory additionalContext listing valid types.
func TestWipWrite_InvalidTypeValue(t *testing.T) {
	content := "---\ntype: memo\n---\n\n# Content"
	contentJSON, _ := json.Marshal(content)
	toolInput := `{"file_path":".sos/wip/BAD.md","content":` + string(contentJSON) + `}`
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

// W7: .sos/wip/ Edit does NOT trigger validation — allow with no additionalContext.
func TestWipEdit_BypassesValidation(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Edit",
		ToolInput: `{"file_path":".sos/wip/existing.md","old_string":"old","new_string":"new"}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (edit bypasses wip validation)", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for Edit tool bypass, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W8: Non-.sos/wip/ Write is unchanged (regression) — allow, no additionalContext.
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
		ToolInput: `{"file_path":".sos/sessions/test/SESSION_CONTEXT.md","content":"bad"}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny (protected file unchanged)", result.HookSpecificOutput.PermissionDecision)
	}
}

// W10: .sos/wip/ with absolute path and valid frontmatter — allow, no additionalContext.
func TestWipWrite_AbsolutePath_ValidFrontmatter(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput("/Users/tom/project/.sos/wip/SPIKE-x.md", "---\ntype: spike\n---\n\n# Spike"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want allow (absolute .sos/wip/ path)", result.HookSpecificOutput.PermissionDecision)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should be empty for valid frontmatter, got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// W11: .sos/wip/ write with empty content — allow with advisory additionalContext.
func TestWipWrite_EmptyContent(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":".sos/wip/empty.md","content":""}`,
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
		{".sos/wip/SPIKE-foo.md", true},
		{".sos/wip/scratch.md", true},
		{"/home/user/project/.sos/wip/DESIGN-x.md", true},
		{"/Users/tom/project/.sos/wip/bar.md", true},
		{"src/main.go", false},
		{".sos/sessions/test/SESSION_CONTEXT.md", false},
		{"docs/wip-notes.md", false},
		{".wip/SPIKE-foo.md", false},          // legacy .wip/ no longer matches
		{"/Users/tom/.wip/bar.md", false},      // legacy .wip/ no longer matches
		{".claude/wip/report.md", false},        // legacy .claude/wip/ no longer matches
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
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":".sos/sessions/test-session/SESSION_CONTEXT.md","content":"bad"},"session_id":"test-session"}`
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
	archiveSessionDir := tmpDir + "/.sos/archive/" + sessionID
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
	liveContextPath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"

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

// =============================================================================
// Section Classification Unit Tests (SC-01 through SC-18)
// Tests classifyEditSection() directly against the decision table in SESSION-4 Section 9.1.
// =============================================================================

func TestClassifyEditSection(t *testing.T) {
	tests := []struct {
		id          string
		description string
		toolInput   string
		want        SectionClass
	}{
		{
			id:          "SC-01",
			description: "Single timeline entry",
			toolInput:   `{"old_string":"- 14:30 | SESSION  | created: Add dark mode"}`,
			want:        SectionTimeline,
		},
		{
			id:          "SC-02",
			description: "Multiple timeline entries",
			toolInput:   `{"old_string":"- 14:30 | SESSION  | created\n- 14:35 | AGENT    | delegated"}`,
			want:        SectionTimeline,
		},
		{
			id:          "SC-03",
			description: "Timeline heading only",
			toolInput:   `{"old_string":"## Timeline"}`,
			want:        SectionTimeline,
		},
		{
			id:          "SC-04",
			description: "Timeline heading + entry",
			toolInput:   `{"old_string":"## Timeline\n- 14:30 | SESSION  | created"}`,
			want:        SectionTimeline,
		},
		{
			id:          "SC-05",
			description: "Frontmatter delimiter only",
			toolInput:   `{"old_string":"---"}`,
			want:        SectionFrontmatter,
		},
		{
			id:          "SC-06",
			description: "Frontmatter key: status",
			toolInput:   `{"old_string":"status: ACTIVE"}`,
			want:        SectionFrontmatter,
		},
		{
			id:          "SC-07",
			description: "Multiple frontmatter keys",
			toolInput:   `{"old_string":"status: ACTIVE\ncurrent_phase: requirements"}`,
			want:        SectionFrontmatter,
		},
		{
			id:          "SC-08",
			description: "Other section heading: Artifacts",
			toolInput:   `{"old_string":"## Artifacts"}`,
			want:        SectionOther,
		},
		{
			id:          "SC-09",
			description: "Other section heading: Blockers",
			toolInput:   `{"old_string":"## Blockers"}`,
			want:        SectionOther,
		},
		{
			id:          "SC-10",
			description: "Other section heading: Next Steps",
			toolInput:   `{"old_string":"## Next Steps"}`,
			want:        SectionOther,
		},
		{
			id:          "SC-11",
			description: "Custom section heading",
			toolInput:   `{"old_string":"## Design Decisions"}`,
			want:        SectionOther,
		},
		{
			id:          "SC-12",
			description: "Mixed: timeline entry + frontmatter key",
			toolInput:   `{"old_string":"- 14:30 | SESSION  | created\nstatus: ACTIVE"}`,
			want:        SectionMixed,
		},
		{
			id:          "SC-13",
			description: "Mixed: timeline entry + other section heading",
			toolInput:   `{"old_string":"- 14:30 | SESSION  | created\n## Artifacts"}`,
			want:        SectionMixed,
		},
		{
			id:          "SC-14",
			description: "Mixed: frontmatter delimiter + other section heading",
			toolInput:   `{"old_string":"---\n## Blockers"}`,
			want:        SectionMixed,
		},
		{
			id:          "SC-15",
			description: "Unknown: only context lines",
			toolInput:   `{"old_string":"Some random text\nAnother line"}`,
			want:        SectionUnknown,
		},
		{
			id:          "SC-16",
			description: "Unknown: empty old_string",
			toolInput:   `{"old_string":""}`,
			want:        SectionUnknown,
		},
		{
			id:          "SC-17",
			description: "Timeline entry + blank lines (blanks are neutral)",
			toolInput:   `{"old_string":"\n- 14:30 | SESSION  | created\n\n"}`,
			want:        SectionTimeline,
		},
		{
			id:          "SC-18",
			description: "Timeline entry + context lines (context lines are neutral)",
			toolInput:   `{"old_string":"- 14:30 | SESSION  | created\nSome note"}`,
			want:        SectionTimeline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id+"_"+tt.description, func(t *testing.T) {
			got := classifyEditSection(tt.toolInput)
			if got != tt.want {
				t.Errorf("classifyEditSection(%q) = %v, want %v", tt.toolInput, got, tt.want)
			}
		})
	}
}

// TestClassifyEditSection_MissingOldString covers the absent old_string field.
func TestClassifyEditSection_MissingOldString(t *testing.T) {
	// No old_string field at all — should fail closed.
	got := classifyEditSection(`{"file_path":"foo.md","new_string":"bar"}`)
	if got != SectionUnknown {
		t.Errorf("classifyEditSection() = %v, want SectionUnknown (missing old_string)", got)
	}
}

// TestClassifyEditSection_InvalidJSON covers malformed JSON input — should fail closed.
func TestClassifyEditSection_InvalidJSON(t *testing.T) {
	got := classifyEditSection("not json")
	if got != SectionUnknown {
		t.Errorf("classifyEditSection(%q) = %v, want SectionUnknown (invalid JSON)", "not json", got)
	}
}

// TestClassifyEditSection_FrontmatterKeys validates each of the 17 frontmatter keys.
func TestClassifyEditSection_FrontmatterKeys(t *testing.T) {
	keys := []string{
		"schema_version", "session_id", "status", "created_at",
		"initiative", "complexity", "active_rite", "rite",
		"current_phase", "timeline_version", "parked_at", "parked_reason",
		"archived_at", "resumed_at", "frayed_from", "fray_point", "strands",
	}
	for _, key := range keys {
		toolInput := `{"old_string":"` + key + `: somevalue"}`
		got := classifyEditSection(toolInput)
		if got != SectionFrontmatter {
			t.Errorf("key %q: classifyEditSection() = %v, want SectionFrontmatter", key, got)
		}
	}
}

// =============================================================================
// Integration Tests (WG-E1 through WG-R2)
// Tests runWriteguardCore() end-to-end for SESSION_CONTEXT.md section routing.
// =============================================================================

// makeSessionCtxWithLock creates a temp project dir with an optional Moirai lock for a session.
func makeSessionCtxWithLock(t *testing.T, sessionID string, withLock bool) (string, *cmdContext) {
	t.Helper()
	tmpDir := t.TempDir()
	sessionDir := tmpDir + "/.sos/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if withLock {
		lockData := `{"agent":"moirai","acquired_at":"` + time.Now().Format(time.RFC3339) + `","session_id":"` + sessionID + `","stale_after_seconds":300}`
		if err := os.WriteFile(sessionDir+"/.moirai-lock", []byte(lockData), 0o644); err != nil {
			t.Fatal(err)
		}
	}
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
	return tmpDir, ctx
}

// WG-E1: Edit timeline entry, no lock → ALLOW with advisory
func TestWriteguard_SectionEdit_TimelineNoLock_Allow(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef01"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, false)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "- 14:30 | SESSION  | created: Add dark mode",
		"new_string": "- 14:30 | SESSION  | created: Add dark mode\n- 14:35 | AGENT    | delegated",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Edit",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("WG-E1: PermissionDecision = %q, want allow", result.HookSpecificOutput.PermissionDecision)
	}
	// Advisory context must be present (E1 advisory reminder)
	if result.HookSpecificOutput.AdditionalContext == "" {
		t.Error("WG-E1: AdditionalContext should be present for timeline allow")
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.AdditionalContext), []byte("Timeline append allowed")) {
		t.Errorf("WG-E1: AdditionalContext should mention 'Timeline append allowed', got: %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

// WG-E2: Edit frontmatter, lock held → ALLOW
func TestWriteguard_SectionEdit_FrontmatterWithLock_Allow(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef02"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, true)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "status: ACTIVE",
		"new_string": "status: PARKED",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Edit",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("WG-E2: PermissionDecision = %q, want allow (frontmatter + lock)", result.HookSpecificOutput.PermissionDecision)
	}
}

// WG-E3: Edit frontmatter, no lock → DENY with frontmatter advisory
func TestWriteguard_SectionEdit_FrontmatterNoLock_Deny(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef03"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, false)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "status: ACTIVE",
		"new_string": "status: PARKED",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Edit",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("WG-E3: PermissionDecision = %q, want deny", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("frontmatter")) {
		t.Errorf("WG-E3: Reason should mention 'frontmatter', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.AdditionalContext), []byte("field-set")) {
		t.Errorf("WG-E3: AdditionalContext should mention 'field-set', got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// WG-E4: Edit Artifacts section, lock held → ALLOW
func TestWriteguard_SectionEdit_ArtifactsWithLock_Allow(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef04"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, true)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "## Artifacts",
		"new_string": "## Artifacts\n- PRD: complete",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Edit",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("WG-E4: PermissionDecision = %q, want allow (Artifacts + lock)", result.HookSpecificOutput.PermissionDecision)
	}
}

// WG-E5: Edit Artifacts section, no lock → DENY with body section advisory
func TestWriteguard_SectionEdit_ArtifactsNoLock_Deny(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef05"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, false)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "## Artifacts",
		"new_string": "## Artifacts\n- PRD: complete",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Edit",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("WG-E5: PermissionDecision = %q, want deny", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("body section")) {
		t.Errorf("WG-E5: Reason should mention 'body section', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.AdditionalContext), []byte("Moirai-managed")) {
		t.Errorf("WG-E5: AdditionalContext should mention 'Moirai-managed', got: %q", result.HookSpecificOutput.AdditionalContext)
	}
}

// WG-E6: Edit mixed timeline + frontmatter → DENY always (regardless of lock)
func TestWriteguard_SectionEdit_Mixed_Deny(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef06"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, true) // lock held — still must deny
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "- 14:30 | SESSION  | created\nstatus: ACTIVE",
		"new_string": "- 14:30 | SESSION  | created\nstatus: PARKED",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Edit",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("WG-E6: PermissionDecision = %q, want deny (mixed edit, fail-closed)", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("multiple")) {
		t.Errorf("WG-E6: Reason should mention 'multiple', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// WG-E7: Edit unknown content → DENY always
func TestWriteguard_SectionEdit_Unknown_Deny(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef07"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, false)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "some random text without any section indicators",
		"new_string": "some modified text",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Edit",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("WG-E7: PermissionDecision = %q, want deny (unknown section, fail-closed)", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("Cannot determine")) {
		t.Errorf("WG-E7: Reason should mention 'Cannot determine', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// WG-W1: Write SESSION_CONTEXT, lock held → ALLOW
func TestWriteguard_SessionContextWrite_WithLock_Allow(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef08"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, true)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path": filePath,
		"content":   "---\nschema_version: \"3.0\"\n---\n\n## Timeline\n",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Write",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("WG-W1: PermissionDecision = %q, want allow (Write + lock)", result.HookSpecificOutput.PermissionDecision)
	}
}

// WG-W2: Write SESSION_CONTEXT, no lock → DENY
func TestWriteguard_SessionContextWrite_NoLock_Deny(t *testing.T) {
	sessionID := "session-20260226-120000-abcdef09"
	tmpDir, ctx := makeSessionCtxWithLock(t, sessionID, false)
	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path": filePath,
		"content":   "---\nschema_version: \"3.0\"\n---\n",
	})

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "PreToolUse",
		ToolName:   "Write",
		ToolInput:  string(toolInputJSON),
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runWriteguardCore(ctx, printer); err != nil {
		t.Fatalf("runWriteguardCore() error = %v", err)
	}
	var result hook.PreToolUseOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("WG-W2: PermissionDecision = %q, want deny (Write without lock)", result.HookSpecificOutput.PermissionDecision)
	}
	if !bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("Moirai")) {
		t.Errorf("WG-W2: Reason should mention 'Moirai', got: %q", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// WG-P1: Edit SPRINT_CONTEXT.md, no lock → DENY (unchanged behavior, no section detection)
func TestWriteguard_SprintContextEdit_NoLock_Deny(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Edit",
		ToolInput: `{"file_path":".sos/sessions/session-abc/sprints/current/SPRINT_CONTEXT.md","old_string":"status: ACTIVE","new_string":"status: DONE"}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("WG-P1: PermissionDecision = %q, want deny (SPRINT_CONTEXT no lock)", result.HookSpecificOutput.PermissionDecision)
	}
	// Should NOT give frontmatter-specific message — uses generic block for non-SESSION_CONTEXT
	if bytes.Contains([]byte(result.HookSpecificOutput.PermissionDecisionReason), []byte("frontmatter")) {
		t.Errorf("WG-P1: SPRINT_CONTEXT should not use frontmatter message, got: %q",
			result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// WG-R1: Edit regular file → ALLOW (regression test: section detection does not apply)
func TestWriteguard_RegularFileEdit_Allow(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Edit",
		ToolInput: `{"file_path":"src/main.go","old_string":"old code","new_string":"new code"}`,
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("WG-R1: PermissionDecision = %q, want allow (regular file)", result.HookSpecificOutput.PermissionDecision)
	}
}

// WG-R2: Write .sos/wip/ file → ALLOW (regression: wip path handled before protected check)
func TestWriteguard_WipWrite_Regression(t *testing.T) {
	result := runWipTest(t, &testutil.HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: makeToolInput(".sos/wip/SPIKE-x.md", "---\ntype: spike\n---\n\n# Spike"),
	})
	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("WG-R2: PermissionDecision = %q, want allow (.sos/wip/ regression)", result.HookSpecificOutput.PermissionDecision)
	}
}

// =============================================================================
// Benchmarks for new section-detection paths
// =============================================================================

// BenchmarkWriteguardHook_TimelineAllow benchmarks the timeline allow path.
func BenchmarkWriteguardHook_TimelineAllow(b *testing.B) {
	sessionID := "session-20260226-120000-bench0001"
	tmpDir := b.TempDir()
	sessionDir := tmpDir + "/.sos/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		b.Fatal(err)
	}

	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "- 14:30 | SESSION  | created: benchmark test",
		"new_string": "- 14:30 | SESSION  | created: benchmark test\n- 14:31 | AGENT    | next",
	})

	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Edit")
	os.Setenv("CLAUDE_TOOL_INPUT", string(toolInputJSON))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_TOOL_NAME")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
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
		b.Errorf("TimelineAllow took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkWriteguardHook_FrontmatterBlock benchmarks the frontmatter block path (no lock).
func BenchmarkWriteguardHook_FrontmatterBlock(b *testing.B) {
	sessionID := "session-20260226-120000-bench0002"
	tmpDir := b.TempDir()
	sessionDir := tmpDir + "/.sos/sessions/" + sessionID
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		b.Fatal(err)
	}

	filePath := ".sos/sessions/" + sessionID + "/SESSION_CONTEXT.md"
	toolInputJSON, _ := json.Marshal(map[string]any{
		"file_path":  filePath,
		"old_string": "status: ACTIVE",
		"new_string": "status: PARKED",
	})

	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	os.Setenv("CLAUDE_TOOL_NAME", "Edit")
	os.Setenv("CLAUDE_TOOL_INPUT", string(toolInputJSON))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_TOOL_NAME")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
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
		b.Errorf("FrontmatterBlock took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}


package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hookpkg "github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestClewOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   ClewOutput
		expected string
	}{
		{
			name: "recorded successfully",
			output: ClewOutput{
				Recorded:   true,
				EventsFile: "/path/to/events.jsonl",
			},
			expected: "Event recorded to /path/to/events.jsonl",
		},
		{
			name: "not recorded with reason",
			output: ClewOutput{
				Recorded: false,
				Reason:   "no active session",
			},
			expected: "Not recorded: no active session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.Text()
			if result != tt.expected {
				t.Errorf("Text() = %q, want %q", result, tt.expected)
			}
		})
	}
}


func TestRunClew_WrongEventType(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
		ToolName:    "Edit",
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:  &outputFlag,
				Verbose: &verboseFlag,
			},
		},
	}

	err := runClewCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Recorded {
		t.Error("Expected Recorded=false for wrong event type")
	}
	if result.Reason != "not a post_tool event" {
		t.Errorf("Reason = %q, want %q", result.Reason, "not a post_tool event")
	}
}

func TestRunClew_NoActiveSession(t *testing.T) {
	// Create temp project dir without session
	tmpDir := t.TempDir()

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Edit",
		ProjectDir:  tmpDir,
		ToolInput:   `{"file_path": "/some/file.go"}`,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runClewCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Recorded {
		t.Error("Expected Recorded=false when no active session")
	}
	if result.Reason != "no active session" {
		t.Errorf("Reason = %q, want %q", result.Reason, "no active session")
	}
}

func TestRunClew_WithActiveSession(t *testing.T) {
	// Create temp project structure with session
	tmpDir := t.TempDir()
	sosDir := filepath.Join(tmpDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	sessionID := "test-session-001"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Create session directory structure
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Edit",
		ProjectDir:  tmpDir,
		ToolInput:   `{"file_path": "/some/file.go", "old_string": "foo", "new_string": "bar"}`,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	err := runClewCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Recorded {
		t.Errorf("Expected Recorded=true, got false with reason: %s", result.Reason)
	}

	// Verify events.jsonl was created
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	if _, err := os.Stat(eventsPath); os.IsNotExist(err) {
		t.Error("events.jsonl was not created")
	}
}

func TestRunClew_OrchestratorStamping(t *testing.T) {
	// Create temp project structure with session
	tmpDir := t.TempDir()
	sosDir := filepath.Join(tmpDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	sessionID := "test-session-orchestrator"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Create session directory structure
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Simulate orchestrator CONSULTATION_RESPONSE with throughline
	orchestratorResult := `request_id: req-001

directive:
  action: invoke_specialist

specialist:
  name: principal-engineer
  prompt: "Complete prompt here"

state_update:
  current_phase: implementation
  next_phases: [validation]
  routing_rationale: "Design approved"

throughline:
  decision: "Route to principal-engineer for implementation"
  rationale: "TDD-user-auth approved with complete API contracts."
`

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Task",
		ProjectDir:  tmpDir,
		ToolInput:   `{"agent": "orchestrator", "prompt": "Analyze requirements"}`,
		ToolResult:  orchestratorResult,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	err := runClewCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Recorded {
		t.Errorf("Expected Recorded=true, got false with reason: %s", result.Reason)
	}

	// Verify events.jsonl was created and contains both tool_call AND decision events
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	eventsContent := string(eventsData)

	// Should have a tool_call event for Task
	if !strings.Contains(eventsContent, `"type":"tool.call"`) {
		t.Error("events.jsonl missing tool_call event")
	}
	if !strings.Contains(eventsContent, `"tool":"Task"`) {
		t.Error("events.jsonl missing Task tool")
	}

	// Should have a decision event from the stamp
	if !strings.Contains(eventsContent, `"type":"agent.decision"`) {
		t.Error("events.jsonl missing decision event from stamp")
	}
	if !strings.Contains(eventsContent, "Route to principal-engineer for implementation") {
		t.Error("events.jsonl missing decision summary from throughline")
	}
	if !strings.Contains(eventsContent, "TDD-user-auth approved with complete API contracts") {
		t.Error("events.jsonl missing rationale from throughline")
	}
}

func TestRunClew_NonOrchestratorTask(t *testing.T) {
	// Create temp project structure with session
	tmpDir := t.TempDir()
	sosDir := filepath.Join(tmpDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	sessionID := "test-session-regular-task"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Create session directory structure
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Regular Task result without throughline (non-orchestrator)
	regularTaskResult := `Task completed successfully.

Files modified:
- /path/to/file.go
`

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Task",
		ProjectDir:  tmpDir,
		ToolInput:   `{"agent": "principal-engineer", "prompt": "Implement feature"}`,
		ToolResult:  regularTaskResult,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	err := runClewCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runThread() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Recorded {
		t.Errorf("Expected Recorded=true, got false with reason: %s", result.Reason)
	}

	// Verify events.jsonl has tool_call but NOT decision event (no throughline)
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	eventsContent := string(eventsData)

	// Should have a tool_call event for Task
	if !strings.Contains(eventsContent, `"type":"tool.call"`) {
		t.Error("events.jsonl missing tool_call event")
	}

	// Should NOT have a decision event (no throughline in non-orchestrator Task)
	if strings.Contains(eventsContent, `"type":"agent.decision"`) {
		t.Error("events.jsonl should NOT have decision event for non-orchestrator Task")
	}
}

func TestClew_StdinIntegration_RecordsToolEvent(t *testing.T) {
	// Test that the full production path works with stdin JSON

	// Create temp project structure with session
	tmpDir := t.TempDir()
	sosDir := filepath.Join(tmpDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	sessionID := "test-stdin-session"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	// Create session directory structure
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Save and restore original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create pipe with CC-format JSON
	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":"/tmp/test.txt","content":"hello"},"tool_response":{"success":true},"session_id":"test-stdin-session","cwd":"` + tmpDir + `"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	err := runClewCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runClew() error = %v", err)
	}

	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Recorded {
		t.Errorf("Expected Recorded=true (stdin should work), got false with reason: %s", result.Reason)
	}

	// Verify events.jsonl was created and contains tool_call event with Write tool
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	eventsContent := string(eventsData)

	// Should have a tool_call event for Write
	if !strings.Contains(eventsContent, `"type":"tool.call"`) {
		t.Error("events.jsonl missing tool_call event")
	}
	if !strings.Contains(eventsContent, `"tool":"Write"`) {
		t.Errorf("events.jsonl should contain Write tool, got: %s", eventsContent)
	}
}

// --- .sos/wip/ artifact detection tests (C1-C6) ---

// makeClewSession creates a temp session dir and returns (tmpDir, sessionDir, ctx).
func makeClewSession(t *testing.T, sessionID string) (string, string, *cmdContext) {
	t.Helper()
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, ".sos", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	outputFlag := "json"
	verboseFlag := false
	projectDir := tmpDir
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDPtr,
		},
	}
	return tmpDir, sessionDir, ctx
}

// runClewWithStdin runs runClewCore with the given JSON payload on stdin and returns parsed output.
func runClewWithStdin(t *testing.T, ctx *cmdContext, payload string) ClewOutput {
	t.Helper()
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	if err := runClewCore(nil, ctx, printer); err != nil {
		t.Fatalf("runClewCore() error = %v", err)
	}
	var result ClewOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}
	return result
}

// C1: .sos/wip/ write with valid frontmatter emits artifact_created with correct metadata.
func TestClew_WipWrite_ValidFrontmatter_EmitsArtifact(t *testing.T) {
	_, sessionDir, ctx := makeClewSession(t, "test-wip-c1")
	content := "---\\ntype: design\\n---\\n\\n# Design doc"
	payload := `{"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":".sos/wip/DESIGN-foo.md","content":"---\ntype: design\n---\n\n# Design doc"},"session_id":"test-wip-c1"}`

	result := runClewWithStdin(t, ctx, payload)
	if !result.Recorded {
		t.Fatalf("Expected Recorded=true, got false: %s", result.Reason)
	}

	_ = content // used in payload above
	eventsData, err := os.ReadFile(filepath.Join(sessionDir, "events.jsonl"))
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}
	eventsContent := string(eventsData)

	if !strings.Contains(eventsContent, `"type":"tool.artifact_created"`) {
		t.Error("events.jsonl missing artifact_created event for .sos/wip/ write")
	}
	if !strings.Contains(eventsContent, `"ephemeral"`) {
		t.Error("events.jsonl missing artifact_type ephemeral")
	}
	if !strings.Contains(eventsContent, `"design"`) {
		t.Error("events.jsonl missing wip_type design")
	}
	if !strings.Contains(eventsContent, `"DESIGN-foo"`) {
		t.Error("events.jsonl missing slug DESIGN-foo")
	}
}

// C2: .sos/wip/ write with missing frontmatter still emits artifact_created with wip_type "unknown".
func TestClew_WipWrite_MissingFrontmatter_EmitsUnknown(t *testing.T) {
	_, sessionDir, ctx := makeClewSession(t, "test-wip-c2")
	payload := `{"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":".sos/wip/BAD.md","content":"# no frontmatter"},"session_id":"test-wip-c2"}`

	result := runClewWithStdin(t, ctx, payload)
	if !result.Recorded {
		t.Fatalf("Expected Recorded=true, got false: %s", result.Reason)
	}

	eventsData, err := os.ReadFile(filepath.Join(sessionDir, "events.jsonl"))
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}
	eventsContent := string(eventsData)

	if !strings.Contains(eventsContent, `"type":"tool.artifact_created"`) {
		t.Error("events.jsonl should emit artifact_created even for missing frontmatter")
	}
	if !strings.Contains(eventsContent, `"ephemeral"`) {
		t.Error("events.jsonl missing artifact_type ephemeral")
	}
	if !strings.Contains(eventsContent, `"unknown"`) {
		t.Error("events.jsonl missing wip_type unknown for missing frontmatter")
	}
}

// C3: Non-.sos/wip/ Write does NOT emit artifact_created (regression).
func TestClew_NonWipWrite_NoArtifactCreated(t *testing.T) {
	_, sessionDir, ctx := makeClewSession(t, "test-wip-c3")
	payload := `{"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":"src/main.go","content":"package main"},"session_id":"test-wip-c3"}`

	result := runClewWithStdin(t, ctx, payload)
	if !result.Recorded {
		t.Fatalf("Expected Recorded=true, got false: %s", result.Reason)
	}

	eventsData, err := os.ReadFile(filepath.Join(sessionDir, "events.jsonl"))
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}
	eventsContent := string(eventsData)

	// Should NOT contain artifact_created for a regular source file
	if strings.Contains(eventsContent, `"type":"tool.artifact_created"`) {
		t.Error("events.jsonl should NOT have artifact_created for non-.sos/wip/ write")
	}
}

// C4: Existing PRD pattern unchanged — artifact_created with artifact_type "prd" (regression).
func TestClew_PRDPattern_Unchanged(t *testing.T) {
	_, sessionDir, ctx := makeClewSession(t, "test-wip-c4")
	payload := `{"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":"PRD-foo.md","content":"# PRD content"},"session_id":"test-wip-c4"}`

	result := runClewWithStdin(t, ctx, payload)
	if !result.Recorded {
		t.Fatalf("Expected Recorded=true, got false: %s", result.Reason)
	}

	eventsData, err := os.ReadFile(filepath.Join(sessionDir, "events.jsonl"))
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}
	eventsContent := string(eventsData)

	if !strings.Contains(eventsContent, `"type":"tool.artifact_created"`) {
		t.Error("events.jsonl missing artifact_created for PRD file")
	}
	if !strings.Contains(eventsContent, `"prd"`) {
		t.Error("events.jsonl missing artifact_type prd for PRD file")
	}
}

// C5: .sos/wip/ Edit does NOT emit artifact_created — only file_change.
func TestClew_WipEdit_NoArtifactCreated(t *testing.T) {
	_, sessionDir, ctx := makeClewSession(t, "test-wip-c5")
	payload := `{"hook_event_name":"PostToolUse","tool_name":"Edit","tool_input":{"file_path":".sos/wip/existing.md","old_string":"old","new_string":"new"},"session_id":"test-wip-c5"}`

	result := runClewWithStdin(t, ctx, payload)
	if !result.Recorded {
		t.Fatalf("Expected Recorded=true, got false: %s", result.Reason)
	}

	eventsData, err := os.ReadFile(filepath.Join(sessionDir, "events.jsonl"))
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}
	eventsContent := string(eventsData)

	// Edit tool: file_change emitted, artifact_created NOT emitted
	if !strings.Contains(eventsContent, `"type":"tool.file_change"`) {
		t.Error("events.jsonl missing file_change event for Edit tool")
	}
	if strings.Contains(eventsContent, `"type":"tool.artifact_created"`) {
		t.Error("events.jsonl should NOT have artifact_created for Edit tool (only Write triggers it)")
	}
}

// C6: Slug derivation — strips only final extension.
func TestWipSlug(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{".sos/wip/DESIGN-ephemeral-artifacts.md", "DESIGN-ephemeral-artifacts"},
		{".sos/wip/SPIKE-complex-name.analysis.md", "SPIKE-complex-name.analysis"},
		{".sos/wip/scratch.md", "scratch"},
		{"/home/user/.sos/wip/TRIAGE-foo.md", "TRIAGE-foo"},
	}
	for _, tt := range tests {
		got := wipSlug(tt.path)
		if got != tt.want {
			t.Errorf("wipSlug(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestMatchWipArtifact covers matchWipArtifact directly.
func TestMatchWipArtifact(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		content     string
		wantArtType clewcontract.ArtifactType
		wantWipType string
	}{
		{
			name:        "valid design frontmatter",
			path:        ".sos/wip/DESIGN-foo.md",
			content:     "---\ntype: design\n---\n\nbody",
			wantArtType: clewcontract.ArtifactTypeEphemeral,
			wantWipType: "design",
		},
		{
			name:        "missing frontmatter yields unknown",
			path:        ".sos/wip/BAD.md",
			content:     "# just markdown",
			wantArtType: clewcontract.ArtifactTypeEphemeral,
			wantWipType: "unknown",
		},
		{
			name:        "invalid type yields unknown",
			path:        ".sos/wip/BAD.md",
			content:     "---\ntype: memo\n---\n",
			wantArtType: clewcontract.ArtifactTypeEphemeral,
			wantWipType: "unknown",
		},
		{
			name:        "non-wip path returns empty",
			path:        "src/main.go",
			content:     "package main",
			wantArtType: "",
			wantWipType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &hookpkg.ToolInput{}
			input.Content = tt.content
			artType, wipType := matchWipArtifact(tt.path, input)
			if artType != tt.wantArtType {
				t.Errorf("artType = %q, want %q", artType, tt.wantArtType)
			}
			if wipType != tt.wantWipType {
				t.Errorf("wipType = %q, want %q", wipType, tt.wantWipType)
			}
		})
	}
}

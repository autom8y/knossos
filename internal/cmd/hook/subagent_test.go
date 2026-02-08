package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
)

func TestSubagentStart_NoSession(t *testing.T) {
	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSubagentStartCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result subagentResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Recorded {
		t.Error("expected recorded=false for no session")
	}
	if result.Reason != "no active session" {
		t.Errorf("expected reason 'no active session', got: %s", result.Reason)
	}
}

func TestSubagentStop_NoSession(t *testing.T) {
	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSubagentStopCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result subagentResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Recorded {
		t.Error("expected recorded=false for no session")
	}
}

func TestSubagentStart_WrongEvent(t *testing.T) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PostToolUse")
	defer os.Unsetenv("CLAUDE_HOOK_EVENT")

	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSubagentStartCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result subagentResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Recorded {
		t.Error("expected recorded=false for wrong event")
	}
	if result.Reason != "not a SubagentStart event" {
		t.Errorf("expected reason 'not a SubagentStart event', got: %s", result.Reason)
	}
}

func TestSubagentStop_WrongEvent(t *testing.T) {
	os.Setenv("CLAUDE_HOOK_EVENT", "PreToolUse")
	defer os.Unsetenv("CLAUDE_HOOK_EVENT")

	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSubagentStopCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result subagentResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result.Recorded {
		t.Error("expected recorded=false for wrong event")
	}
	if result.Reason != "not a SubagentStop event" {
		t.Errorf("expected reason 'not a SubagentStop event', got: %s", result.Reason)
	}
}

func TestSubagentStart_LogsToClew(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-subagent-start"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	os.WriteFile(currentSessionFile, []byte(sessionID), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventSubagentStart))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	os.Setenv("CLAUDE_TOOL_INPUT", `{"agent_name":"integration-engineer","agent_type":"specialist","task_id":"task-016"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
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

	if err := runSubagentStartCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result subagentResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if !result.Recorded {
		t.Errorf("expected recorded=true, got reason: %s", result.Reason)
	}

	// Verify clew event was written
	eventsPath := filepath.Join(sessionDir, clewcontract.EventsFileName)
	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("events.jsonl not created: %v", err)
	}

	content := string(eventsData)
	if !strings.Contains(content, "Subagent started: integration-engineer") {
		t.Error("events.jsonl missing subagent start summary")
	}
	if !strings.Contains(content, "task_start") {
		t.Error("events.jsonl missing task_start event type")
	}
}

func TestSubagentStop_LogsToClew(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-subagent-stop"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	os.WriteFile(currentSessionFile, []byte(sessionID), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventSubagentStop))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	os.Setenv("CLAUDE_TOOL_INPUT", `{"agent_name":"context-architect","type":"specialist"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
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

	if err := runSubagentStopCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result subagentResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if !result.Recorded {
		t.Errorf("expected recorded=true, got reason: %s", result.Reason)
	}

	// Verify clew event was written
	eventsPath := filepath.Join(sessionDir, clewcontract.EventsFileName)
	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("events.jsonl not created: %v", err)
	}

	content := string(eventsData)
	if !strings.Contains(content, "Subagent stopped: context-architect") {
		t.Error("events.jsonl missing subagent stop summary")
	}
	if !strings.Contains(content, "task_end") {
		t.Error("events.jsonl missing task_end event type")
	}
}

func TestParseSubagentInfo_ValidJSON(t *testing.T) {
	info := parseSubagentInfo(`{"agent_name":"integration-engineer","agent_type":"specialist","task_id":"task-016"}`)
	if info.AgentName != "integration-engineer" {
		t.Errorf("AgentName = %q, want %q", info.AgentName, "integration-engineer")
	}
	if info.AgentType != "specialist" {
		t.Errorf("AgentType = %q, want %q", info.AgentType, "specialist")
	}
	if info.TaskID != "task-016" {
		t.Errorf("TaskID = %q, want %q", info.TaskID, "task-016")
	}
}

func TestParseSubagentInfo_NameFallback(t *testing.T) {
	// Falls back to "name" field if "agent_name" is not present
	info := parseSubagentInfo(`{"name":"my-agent"}`)
	if info.AgentName != "my-agent" {
		t.Errorf("AgentName = %q, want %q", info.AgentName, "my-agent")
	}
}

func TestParseSubagentInfo_EmptyJSON(t *testing.T) {
	info := parseSubagentInfo("")
	if info.AgentName != "unknown" {
		t.Errorf("AgentName = %q, want %q", info.AgentName, "unknown")
	}
}

func TestParseSubagentInfo_InvalidJSON(t *testing.T) {
	info := parseSubagentInfo("not json")
	if info.AgentName != "unknown" {
		t.Errorf("AgentName = %q, want %q", info.AgentName, "unknown")
	}
}

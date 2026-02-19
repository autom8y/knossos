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

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventSubagentStart))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	os.Setenv("CLAUDE_TOOL_INPUT", `{"agent_name":"integration-engineer","agent_type":"specialist","task_id":"task-016","agent_id":"agent-abc123"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
	}()

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
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
	if !strings.Contains(content, "agent.task_start") {
		t.Error("events.jsonl missing agent.task_start event type")
	}
}

func TestSubagentStop_LogsToClew(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-subagent-stop"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventSubagentStop))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	os.Setenv("CLAUDE_TOOL_INPUT", `{"agent_name":"context-architect","type":"specialist","agent_id":"agent-def456"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
	}()

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
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
	if !strings.Contains(content, "agent.task_end") {
		t.Error("events.jsonl missing agent.task_end event type")
	}
}

func TestParseSubagentInfo_ValidJSON(t *testing.T) {
	info := parseSubagentInfo(`{"agent_name":"integration-engineer","agent_type":"specialist","task_id":"task-016","agent_id":"agent-abc123"}`)
	if info.AgentName != "integration-engineer" {
		t.Errorf("AgentName = %q, want %q", info.AgentName, "integration-engineer")
	}
	if info.AgentType != "specialist" {
		t.Errorf("AgentType = %q, want %q", info.AgentType, "specialist")
	}
	if info.TaskID != "task-016" {
		t.Errorf("TaskID = %q, want %q", info.TaskID, "task-016")
	}
	if info.AgentID != "agent-abc123" {
		t.Errorf("AgentID = %q, want %q", info.AgentID, "agent-abc123")
	}
}

func TestParseSubagentInfo_AgentIDFallback(t *testing.T) {
	// Falls back to "id" field if "agent_id" is not present
	info := parseSubagentInfo(`{"agent_name":"pythia","id":"agent-fallback-id"}`)
	if info.AgentID != "agent-fallback-id" {
		t.Errorf("AgentID = %q, want %q", info.AgentID, "agent-fallback-id")
	}
}

func TestParseSubagentInfo_AgentIDMissing(t *testing.T) {
	// No agent_id field — AgentID should be empty string
	info := parseSubagentInfo(`{"agent_name":"integration-engineer","agent_type":"specialist"}`)
	if info.AgentID != "" {
		t.Errorf("AgentID = %q, want empty string", info.AgentID)
	}
}

func TestSubagentStart_PersistsThroughlineID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-throughline-persist"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventSubagentStart))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	// pythia is a throughline agent — its ID should be persisted
	os.Setenv("CLAUDE_TOOL_INPUT", `{"agent_name":"pythia","agent_type":"orchestrator","task_id":"task-001","agent_id":"agent-pythia-xyz"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
	}()

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
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

	// Verify .throughline-ids.json was created with pythia's ID
	idFile := filepath.Join(sessionDir, ThroughlineIDsFile)
	data, err := os.ReadFile(idFile)
	if err != nil {
		t.Fatalf(".throughline-ids.json was not created: %v", err)
	}

	var ids map[string]string
	if err := json.Unmarshal(data, &ids); err != nil {
		t.Fatalf("failed to parse .throughline-ids.json: %v", err)
	}
	if ids["pythia"] != "agent-pythia-xyz" {
		t.Errorf("pythia ID = %q, want %q", ids["pythia"], "agent-pythia-xyz")
	}
}

func TestSubagentStart_NonThroughlineAgentNotPersisted(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-non-throughline"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventSubagentStart))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	// integration-engineer is NOT a throughline agent
	os.Setenv("CLAUDE_TOOL_INPUT", `{"agent_name":"integration-engineer","agent_type":"specialist","agent_id":"agent-ie-xyz"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
	}()

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runSubagentStartCore(ctx, printer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// .throughline-ids.json should NOT be created for non-throughline agents
	idFile := filepath.Join(sessionDir, ThroughlineIDsFile)
	if _, err := os.Stat(idFile); err == nil {
		t.Error(".throughline-ids.json should not be created for non-throughline agents")
	}
}

func TestSubagentStart_ThroughlineIDUpsert(t *testing.T) {
	// Verify that a second call updates the entry rather than overwriting others
	tmpDir := t.TempDir()
	sessionID := "session-throughline-upsert"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Pre-seed with moirai entry
	existingIDs := map[string]string{"moirai": "agent-moirai-existing"}
	existingData, _ := json.Marshal(existingIDs)
	os.WriteFile(filepath.Join(sessionDir, ThroughlineIDsFile), existingData, 0644)

	// Add pythia entry via hook
	if err := upsertThroughlineID(sessionDir, "pythia", "agent-pythia-new"); err != nil {
		t.Fatalf("upsertThroughlineID error: %v", err)
	}

	ids := readThroughlineIDs(sessionDir)
	if ids == nil {
		t.Fatal("readThroughlineIDs returned nil")
	}
	if ids["moirai"] != "agent-moirai-existing" {
		t.Errorf("moirai entry was overwritten: %q", ids["moirai"])
	}
	if ids["pythia"] != "agent-pythia-new" {
		t.Errorf("pythia ID = %q, want %q", ids["pythia"], "agent-pythia-new")
	}
}

func TestReadThroughlineIDs_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	ids := readThroughlineIDs(tmpDir)
	if ids != nil {
		t.Errorf("readThroughlineIDs with no file = %v, want nil", ids)
	}
}

func TestSubagentStart_NoAgentIDSkipsPersistence(t *testing.T) {
	// When agent_id is missing, hook must still succeed and NOT write the IDs file
	tmpDir := t.TempDir()
	sessionID := "session-no-agent-id"
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	os.Setenv("CLAUDE_HOOK_EVENT", string(hook.EventSubagentStart))
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	// pythia with no agent_id
	os.Setenv("CLAUDE_TOOL_INPUT", `{"agent_name":"pythia","agent_type":"orchestrator"}`)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
		os.Unsetenv("CLAUDE_TOOL_INPUT")
	}()

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
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
		t.Errorf("expected recorded=true even with no agent_id, got reason: %s", result.Reason)
	}

	// No .throughline-ids.json should be written when agent_id is empty
	idFile := filepath.Join(sessionDir, ThroughlineIDsFile)
	if _, err := os.Stat(idFile); err == nil {
		t.Error(".throughline-ids.json should not be created when agent_id is empty")
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

package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	sess "github.com/autom8y/knossos/internal/session"
)

// setupLogTestSession creates a minimal session environment for log command tests.
// Returns (ctx, sessionsDir, sessionID).
func setupLogTestSession(t *testing.T) (*cmdContext, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectDir := tmpDir
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	sessionID := "session-20260226-100000-logtestxx"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	for _, dir := range []string{sessionDir, locksDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Write SESSION_CONTEXT.md with a ## Timeline section.
	contextContent := "---\n" +
		"schema_version: \"2.1\"\n" +
		"session_id: " + sessionID + "\n" +
		"status: ACTIVE\n" +
		"initiative: test initiative\n" +
		"complexity: MODULE\n" +
		"active_rite: ecosystem\n" +
		"current_phase: requirements\n" +
		"created_at: 2026-02-26T10:00:00Z\n" +
		"---\n\n# Session: test initiative\n\n## Timeline\n\n## Artifacts\n- PRD: pending\n"

	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(ctxPath, []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Write .current-session marker.
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	return ctx, sessionsDir, sessionID
}

// readTypedEventsJSONL reads all TypedEvents from a session's events.jsonl.
func readTypedEventsJSONL(t *testing.T, sessionDir string) []clewcontract.TypedEvent {
	t.Helper()
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	f, err := os.Open(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("failed to open events.jsonl: %v", err)
	}
	defer f.Close()

	var events []clewcontract.TypedEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		// Only decode lines with "data" field (v3 TypedEvent).
		if !strings.Contains(string(line), `"data"`) {
			continue
		}
		var event clewcontract.TypedEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events
}

// TestRunLog_GeneralType verifies that --type=general (default) writes an event and timeline entry.
func TestRunLog_GeneralType(t *testing.T) {
	ctx, sessionsDir, sessionID := setupLogTestSession(t)
	sessionDir := filepath.Join(sessionsDir, sessionID)

	err := runLog(ctx, "started architect handoff", logOptions{eventType: "general"})
	if err != nil {
		t.Fatalf("runLog failed: %v", err)
	}

	// Verify event was written to events.jsonl.
	events := readTypedEventsJSONL(t, sessionDir)
	if len(events) == 0 {
		t.Fatal("expected at least one typed event in events.jsonl")
	}
	// general maps to decision.recorded per spec.
	found := false
	for _, e := range events {
		if e.Type == clewcontract.EventTypeDecisionRecorded {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected decision.recorded event for --type=general; events: %v", events)
	}

	// Verify timeline entry was appended to SESSION_CONTEXT.md.
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	timeline, err := sess.ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}
	if len(timeline) == 0 {
		t.Fatal("expected at least one timeline entry after runLog")
	}
}

// TestRunLog_DecisionType verifies --type=decision produces a decision.recorded event.
func TestRunLog_DecisionType(t *testing.T) {
	ctx, sessionsDir, sessionID := setupLogTestSession(t)
	sessionDir := filepath.Join(sessionsDir, sessionID)

	err := runLog(ctx, "chose CSS vars over styled-components", logOptions{
		eventType: "decision",
		rationale: "runtime perf",
	})
	if err != nil {
		t.Fatalf("runLog failed: %v", err)
	}

	events := readTypedEventsJSONL(t, sessionDir)
	found := false
	for _, e := range events {
		if e.Type == clewcontract.EventTypeDecisionRecorded {
			var d clewcontract.DecisionRecordedData
			if err := json.Unmarshal(e.Data, &d); err != nil {
				t.Fatalf("failed to unmarshal decision data: %v", err)
			}
			if d.Decision == "chose CSS vars over styled-components" && d.Rationale == "runtime perf" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected decision.recorded event with correct decision and rationale")
	}

	// Verify timeline entry.
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	timeline, err := sess.ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}
	if len(timeline) == 0 {
		t.Fatal("expected timeline entry for decision event")
	}
	if !strings.Contains(timeline[0].Raw, "DECISION") {
		t.Errorf("timeline entry = %q, want DECISION category", timeline[0].Raw)
	}
}

// TestRunLog_AgentType verifies --type=agent produces an agent.delegated event.
func TestRunLog_AgentType(t *testing.T) {
	ctx, sessionsDir, sessionID := setupLogTestSession(t)
	sessionDir := filepath.Join(sessionsDir, sessionID)

	err := runLog(ctx, "designing components", logOptions{
		eventType: "agent",
		agent:     "architect",
	})
	if err != nil {
		t.Fatalf("runLog failed: %v", err)
	}

	events := readTypedEventsJSONL(t, sessionDir)
	found := false
	for _, e := range events {
		if e.Type == clewcontract.EventTypeAgentDelegated {
			var d clewcontract.AgentDelegatedData
			if err := json.Unmarshal(e.Data, &d); err != nil {
				t.Fatalf("failed to unmarshal agent data: %v", err)
			}
			if d.AgentName == "architect" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected agent.delegated event with agent_name=architect")
	}

	// Verify timeline entry contains AGENT category.
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	timeline, err := sess.ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}
	if len(timeline) == 0 {
		t.Fatal("expected timeline entry for agent event")
	}
	if !strings.Contains(timeline[0].Raw, "AGENT") {
		t.Errorf("timeline entry = %q, want AGENT category", timeline[0].Raw)
	}
}

// TestRunLog_CommitType verifies --type=commit produces a commit.created event.
func TestRunLog_CommitType(t *testing.T) {
	ctx, sessionsDir, sessionID := setupLogTestSession(t)
	sessionDir := filepath.Join(sessionsDir, sessionID)

	err := runLog(ctx, "feat: theme provider", logOptions{
		eventType: "commit",
		sha:       "abc123f",
	})
	if err != nil {
		t.Fatalf("runLog failed: %v", err)
	}

	events := readTypedEventsJSONL(t, sessionDir)
	found := false
	for _, e := range events {
		if e.Type == clewcontract.EventTypeCommitCreated {
			var d clewcontract.CommitCreatedData
			if err := json.Unmarshal(e.Data, &d); err != nil {
				t.Fatalf("failed to unmarshal commit data: %v", err)
			}
			if d.SHA == "abc123f" && d.Message == "feat: theme provider" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected commit.created event with correct SHA and message")
	}

	// Verify timeline entry.
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	timeline, err := sess.ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}
	if len(timeline) == 0 {
		t.Fatal("expected timeline entry for commit event")
	}
	if !strings.Contains(timeline[0].Raw, "COMMIT") {
		t.Errorf("timeline entry = %q, want COMMIT category", timeline[0].Raw)
	}
	if !strings.Contains(timeline[0].Raw, "abc123f") {
		t.Errorf("timeline entry = %q, want SHA abc123f in summary", timeline[0].Raw)
	}
}

// TestRunLog_CommandType verifies --type=command produces a command.invoked event.
func TestRunLog_CommandType(t *testing.T) {
	ctx, sessionsDir, sessionID := setupLogTestSession(t)
	sessionDir := filepath.Join(sessionsDir, sessionID)

	err := runLog(ctx, "/consult", logOptions{eventType: "command"})
	if err != nil {
		t.Fatalf("runLog failed: %v", err)
	}

	events := readTypedEventsJSONL(t, sessionDir)
	found := false
	for _, e := range events {
		if e.Type == clewcontract.EventTypeCommandInvoked {
			var d clewcontract.CommandInvokedData
			if err := json.Unmarshal(e.Data, &d); err != nil {
				t.Fatalf("failed to unmarshal command data: %v", err)
			}
			if d.Command == "/consult" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected command.invoked event with command=/consult")
	}
}

// TestRunLog_AgentTypeMissingFlag verifies --type=agent without --agent returns error.
func TestRunLog_AgentTypeMissingFlag(t *testing.T) {
	ctx, _, _ := setupLogTestSession(t)

	err := runLog(ctx, "doing something", logOptions{eventType: "agent"})
	if err == nil {
		t.Fatal("expected error when --type=agent without --agent flag")
	}
	if !strings.Contains(err.Error(), "--agent") {
		t.Errorf("error = %q, want mention of --agent flag", err.Error())
	}
}

// TestRunLog_CommitTypeMissingFlag verifies --type=commit without --sha returns error.
func TestRunLog_CommitTypeMissingFlag(t *testing.T) {
	ctx, _, _ := setupLogTestSession(t)

	err := runLog(ctx, "feat: something", logOptions{eventType: "commit"})
	if err == nil {
		t.Fatal("expected error when --type=commit without --sha flag")
	}
	if !strings.Contains(err.Error(), "--sha") {
		t.Errorf("error = %q, want mention of --sha flag", err.Error())
	}
}

// TestRunLog_InvalidType verifies an unrecognized --type returns error.
func TestRunLog_InvalidType(t *testing.T) {
	ctx, _, _ := setupLogTestSession(t)

	err := runLog(ctx, "message", logOptions{eventType: "bogus"})
	if err == nil {
		t.Fatal("expected error for invalid --type value")
	}
}

// TestRunLog_NoActiveSession verifies error when no session exists.
func TestRunLog_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("failed to create locks dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runLog(ctx, "test message", logOptions{eventType: "general"})
	if err == nil {
		t.Fatal("expected error when no active session")
	}
}

// TestRunLog_TimelinePreservedAfterLog verifies existing timeline entries are not overwritten.
func TestRunLog_TimelinePreservedAfterLog(t *testing.T) {
	ctx, sessionsDir, sessionID := setupLogTestSession(t)
	sessionDir := filepath.Join(sessionsDir, sessionID)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

	// Add a pre-existing timeline entry.
	existingEntry := "- 14:00 | SESSION  | created: test initiative (MODULE)"
	if err := sess.AppendEntry(ctxPath, existingEntry); err != nil {
		t.Fatalf("failed to pre-populate timeline: %v", err)
	}

	// Run log to add a second entry.
	err := runLog(ctx, "chose X over Y", logOptions{
		eventType: "decision",
		rationale: "perf",
	})
	if err != nil {
		t.Fatalf("runLog failed: %v", err)
	}

	timeline, err := sess.ReadTimeline(ctxPath)
	if err != nil {
		t.Fatalf("ReadTimeline failed: %v", err)
	}
	if len(timeline) != 2 {
		t.Errorf("len(timeline) = %d, want 2 (original + new)", len(timeline))
	}
	// Original entry should still be there.
	if !strings.Contains(timeline[0].Raw, "created: test initiative (MODULE)") {
		t.Errorf("original entry not preserved: %q", timeline[0].Raw)
	}
	// New entry should be there.
	if !strings.Contains(timeline[1].Raw, "DECISION") {
		t.Errorf("new entry should have DECISION category: %q", timeline[1].Raw)
	}
}

// TestBuildLogEvent_CorrectEventTypes verifies each --type flag maps to the correct event type.
func TestBuildLogEvent_CorrectEventTypes(t *testing.T) {
	cases := []struct {
		opts      logOptions
		wantType  clewcontract.EventType
		message   string
	}{
		{logOptions{eventType: "general"}, clewcontract.EventTypeDecisionRecorded, "note"},
		{logOptions{eventType: "decision", rationale: "r"}, clewcontract.EventTypeDecisionRecorded, "decision text"},
		{logOptions{eventType: "agent", agent: "architect"}, clewcontract.EventTypeAgentDelegated, "task"},
		{logOptions{eventType: "commit", sha: "abc1234"}, clewcontract.EventTypeCommitCreated, "feat: thing"},
		{logOptions{eventType: "command"}, clewcontract.EventTypeCommandInvoked, "/consult"},
	}

	for _, tc := range cases {
		event := buildLogEvent(tc.message, tc.opts)
		if event.Type != tc.wantType {
			t.Errorf("buildLogEvent(%q) type = %q, want %q", tc.opts.eventType, event.Type, tc.wantType)
		}
	}
}

// TestValidateLogFlags_ValidCases verifies valid flag combinations return no error.
func TestValidateLogFlags_ValidCases(t *testing.T) {
	valid := []logOptions{
		{eventType: "general"},
		{eventType: "decision"},
		{eventType: "decision", rationale: "some reason"},
		{eventType: "agent", agent: "architect"},
		{eventType: "commit", sha: "abc1234"},
		{eventType: "command"},
	}
	for _, opts := range valid {
		if err := validateLogFlags(opts); err != nil {
			t.Errorf("validateLogFlags(%+v) = %v, want nil", opts, err)
		}
	}
}

// TestValidateLogFlags_InvalidCases verifies invalid flag combinations return errors.
func TestValidateLogFlags_InvalidCases(t *testing.T) {
	invalid := []logOptions{
		{eventType: "agent"},          // missing --agent
		{eventType: "commit"},         // missing --sha
		{eventType: "badtype"},        // unknown type
	}
	for _, opts := range invalid {
		if err := validateLogFlags(opts); err == nil {
			t.Errorf("validateLogFlags(%+v) = nil, want error", opts)
		}
	}
}

// TestRunLog_EventSourceIsAgent verifies events are tagged with source=agent.
func TestRunLog_EventSourceIsAgent(t *testing.T) {
	ctx, sessionsDir, sessionID := setupLogTestSession(t)
	sessionDir := filepath.Join(sessionsDir, sessionID)

	err := runLog(ctx, "test decision", logOptions{eventType: "decision", rationale: "test"})
	if err != nil {
		t.Fatalf("runLog failed: %v", err)
	}

	events := readTypedEventsJSONL(t, sessionDir)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	for _, e := range events {
		if e.Type == clewcontract.EventTypeDecisionRecorded {
			if e.Source != clewcontract.SourceAgent {
				t.Errorf("event source = %q, want %q", e.Source, clewcontract.SourceAgent)
			}
		}
	}
}

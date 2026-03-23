package clewcontract

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// --- EventSource constants ---

func TestEventSource_Constants(t *testing.T) {
	tests := []struct {
		src      EventSource
		expected string
	}{
		{SourceCLI, "cli"},
		{SourceHook, "hook"},
		{SourceAgent, "agent"},
	}
	for _, tt := range tests {
		if string(tt.src) != tt.expected {
			t.Errorf("EventSource %v = %q, want %q", tt.src, tt.src, tt.expected)
		}
	}
}

// --- New EventType constants ---

func TestNewEventTypeConstants(t *testing.T) {
	tests := []struct {
		constant EventType
		expected string
	}{
		{EventTypeToolInvoked, "tool.invoked"},
		{EventTypeFileModified, "file.modified"},
		{EventTypeDecisionRecorded, "decision.recorded"},
		{EventTypeAgentDelegated, "agent.delegated"},
		{EventTypeAgentCompleted, "agent.completed"},
		{EventTypeArtifactCreatedV3, "artifact.created"},
		{EventTypeErrorOccurred, "error.occurred"},
		{EventTypeSessionWrapped, "session.wrapped"},
		{EventTypeCommitCreated, "commit.created"},
		{EventTypeCommandInvoked, "command.invoked"},
		{EventTypeFieldUpdated, "field.updated"},
		{EventTypeHookFired, "hook.fired"},
	}
	for _, tt := range tests {
		if string(tt.constant) != tt.expected {
			t.Errorf("EventType constant = %q, want %q", tt.constant, tt.expected)
		}
	}
}

// --- TypedEvent serialization ---

func TestTypedEvent_JSONFields(t *testing.T) {
	// All four envelope fields must be present in serialized output.
	event := newTypedEvent(EventTypeSessionCreated, SourceCLI, "", SessionCreatedData{
		SessionID:  "session-001",
		Initiative: "Add dark mode",
		Complexity: "MODULE",
		Rite:       "ecosystem",
	})

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal TypedEvent: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify all four envelope fields are present
	if _, ok := raw["ts"]; !ok {
		t.Error("ts field missing from TypedEvent JSON")
	}
	if _, ok := raw["type"]; !ok {
		t.Error("type field missing from TypedEvent JSON")
	}
	if _, ok := raw["source"]; !ok {
		t.Error("source field missing from TypedEvent JSON")
	}
	if _, ok := raw["data"]; !ok {
		t.Error("data field missing from TypedEvent JSON")
	}

	// Verify no extraneous fields (e.g., v2 "meta", "tool", "path", "summary")
	for _, field := range []string{"meta", "tool", "path", "summary"} {
		if _, ok := raw[field]; ok {
			t.Errorf("unexpected field %q in TypedEvent JSON", field)
		}
	}
}

func TestTypedEvent_SourceField(t *testing.T) {
	// Source must always be set in the JSON output
	tests := []struct {
		name     string
		source   EventSource
		expected string
	}{
		{"cli source", SourceCLI, "cli"},
		{"hook source", SourceHook, "hook"},
		{"agent source", SourceAgent, "agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := newTypedEvent(EventTypeSessionCreated, tt.source, "", SessionCreatedData{
				SessionID:  "s-001",
				Initiative: "test",
				Complexity: "PATCH",
			})

			data, err := json.Marshal(event)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if raw["source"] != tt.expected {
				t.Errorf("source = %v, want %q", raw["source"], tt.expected)
			}
		})
	}
}

func TestTypedEvent_DataNeverNull(t *testing.T) {
	// Data must be a JSON object, never null, even for minimal events.
	event := NewTypedSessionResumedEvent("", "session-001")

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	dataField, ok := raw["data"]
	if !ok {
		t.Fatal("data field missing")
	}
	if dataField == nil {
		t.Error("data field is null, want a JSON object")
	}
	if _, isMap := dataField.(map[string]any); !isMap {
		t.Errorf("data field is %T, want map (JSON object)", dataField)
	}
}

func TestTypedEvent_TimestampFormat(t *testing.T) {
	event := NewTypedSessionCreatedEvent("", "s-001", "initiative", "MODULE", "ecosystem")

	if event.Ts == "" {
		t.Error("Ts should not be empty")
	}

	// Must match TypedEvent timestamp format (same as v2)
	if len(event.Ts) < 24 || event.Ts[len(event.Ts)-1] != 'Z' {
		t.Errorf("Ts %q does not match expected format 2006-01-02T15:04:05.000Z", event.Ts)
	}
}

// --- Per-type data struct round-trips ---

func TestSessionCreatedData_RoundTrip(t *testing.T) {
	original := SessionCreatedData{
		SessionID:  "session-20260226-140300-abc12345",
		Initiative: "Add dark mode",
		Complexity: "MODULE",
		Rite:       "ecosystem",
	}

	data, _ := json.Marshal(original)
	var decoded SessionCreatedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.SessionID != original.SessionID {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, original.SessionID)
	}
	if decoded.Initiative != original.Initiative {
		t.Errorf("Initiative = %q, want %q", decoded.Initiative, original.Initiative)
	}
	if decoded.Complexity != original.Complexity {
		t.Errorf("Complexity = %q, want %q", decoded.Complexity, original.Complexity)
	}
	if decoded.Rite != original.Rite {
		t.Errorf("Rite = %q, want %q", decoded.Rite, original.Rite)
	}
}

func TestSessionCreatedData_OmitEmptyRite(t *testing.T) {
	// Rite is optional; omitempty should suppress it when empty.
	original := SessionCreatedData{
		SessionID:  "s-001",
		Initiative: "cross-cutting",
		Complexity: "PATCH",
		// Rite intentionally left empty
	}

	data, _ := json.Marshal(original)
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := raw["rite"]; ok {
		t.Error("rite should be omitted when empty")
	}
}

func TestSessionWrappedData_RoundTrip(t *testing.T) {
	original := SessionWrappedData{
		SessionID:  "s-001",
		SailsColor: "WHITE",
		DurationMs: 3600000,
	}

	data, _ := json.Marshal(original)
	var decoded SessionWrappedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.SessionID != original.SessionID {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, original.SessionID)
	}
	if decoded.SailsColor != original.SailsColor {
		t.Errorf("SailsColor = %q, want %q", decoded.SailsColor, original.SailsColor)
	}
	if decoded.DurationMs != original.DurationMs {
		t.Errorf("DurationMs = %d, want %d", decoded.DurationMs, original.DurationMs)
	}
}

func TestAgentDelegatedData_RoundTrip(t *testing.T) {
	original := AgentDelegatedData{
		AgentName: "context-architect",
		AgentType: "specialist",
		TaskID:    "task-001",
		AgentID:   "agent-abc123",
	}

	data, _ := json.Marshal(original)
	var decoded AgentDelegatedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.AgentName != original.AgentName {
		t.Errorf("AgentName = %q, want %q", decoded.AgentName, original.AgentName)
	}
	if decoded.AgentType != original.AgentType {
		t.Errorf("AgentType = %q, want %q", decoded.AgentType, original.AgentType)
	}
	if decoded.TaskID != original.TaskID {
		t.Errorf("TaskID = %q, want %q", decoded.TaskID, original.TaskID)
	}
}

func TestAgentCompletedData_RoundTrip(t *testing.T) {
	original := AgentCompletedData{
		AgentName:  "integration-engineer",
		AgentType:  "specialist",
		Outcome:    "success",
		DurationMs: 15000,
		Artifacts:  []string{"/path/to/artifact.go", "/path/to/test.go"},
	}

	data, _ := json.Marshal(original)
	var decoded AgentCompletedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Outcome != original.Outcome {
		t.Errorf("Outcome = %q, want %q", decoded.Outcome, original.Outcome)
	}
	if decoded.DurationMs != original.DurationMs {
		t.Errorf("DurationMs = %d, want %d", decoded.DurationMs, original.DurationMs)
	}
	if len(decoded.Artifacts) != len(original.Artifacts) {
		t.Errorf("Artifacts len = %d, want %d", len(decoded.Artifacts), len(original.Artifacts))
	}
}

func TestCommitCreatedData_RoundTrip(t *testing.T) {
	original := CommitCreatedData{
		SHA:     "abc123f",
		Message: "feat: add theme provider",
	}

	data, _ := json.Marshal(original)
	var decoded CommitCreatedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.SHA != original.SHA {
		t.Errorf("SHA = %q, want %q", decoded.SHA, original.SHA)
	}
	if decoded.Message != original.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, original.Message)
	}
}

func TestDecisionRecordedData_RoundTrip(t *testing.T) {
	original := DecisionRecordedData{
		Decision:  "Use PostgreSQL over MongoDB",
		Rationale: "Better ACID compliance",
		Rejected:  []string{"MongoDB", "CockroachDB"},
	}

	data, _ := json.Marshal(original)
	var decoded DecisionRecordedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Decision != original.Decision {
		t.Errorf("Decision = %q, want %q", decoded.Decision, original.Decision)
	}
	if len(decoded.Rejected) != len(original.Rejected) {
		t.Errorf("Rejected len = %d, want %d", len(decoded.Rejected), len(original.Rejected))
	}
}

func TestDecisionRecordedData_OmitRejected(t *testing.T) {
	// Rejected is optional and should be omitted when nil.
	original := DecisionRecordedData{
		Decision:  "Keep current approach",
		Rationale: "No better alternatives",
	}

	data, _ := json.Marshal(original)
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := raw["rejected"]; ok {
		t.Error("rejected should be omitted when nil")
	}
}

func TestPhaseTransitionedData_RoundTrip(t *testing.T) {
	original := PhaseTransitionedData{
		SessionID: "s-001",
		From:      "design",
		To:        "implementation",
	}

	data, _ := json.Marshal(original)
	var decoded PhaseTransitionedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.SessionID != original.SessionID {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, original.SessionID)
	}
	if decoded.From != original.From {
		t.Errorf("From = %q, want %q", decoded.From, original.From)
	}
	if decoded.To != original.To {
		t.Errorf("To = %q, want %q", decoded.To, original.To)
	}
}

func TestErrorOccurredData_RoundTrip(t *testing.T) {
	original := ErrorOccurredData{
		ErrorCode:       "VALIDATION_FAILED",
		Message:         "Schema check failed",
		Context:         "session.wrap",
		Recoverable:     false,
		SuggestedAction: "Fix the schema and retry",
	}

	data, _ := json.Marshal(original)
	var decoded ErrorOccurredData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ErrorCode != original.ErrorCode {
		t.Errorf("ErrorCode = %q, want %q", decoded.ErrorCode, original.ErrorCode)
	}
	if decoded.Recoverable != original.Recoverable {
		t.Errorf("Recoverable = %v, want %v", decoded.Recoverable, original.Recoverable)
	}
}

func TestSailsGeneratedTypedData_RoundTrip(t *testing.T) {
	original := SailsGeneratedTypedData{
		SessionID:    "s-001",
		Color:        "WHITE",
		ComputedBase: "WHITE",
		Reasons:      []string{"all proofs passing"},
		FilePath:     ".sos/sessions/s-001/WHITE_SAILS.yaml",
		Evidence: map[string]string{
			"tests": ".sos/sessions/s-001/test-output.log",
			"build": ".sos/sessions/s-001/build-output.log",
		},
	}

	data, _ := json.Marshal(original)
	var decoded SailsGeneratedTypedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Color != original.Color {
		t.Errorf("Color = %q, want %q", decoded.Color, original.Color)
	}
	if len(decoded.Reasons) != len(original.Reasons) {
		t.Errorf("Reasons len = %d, want %d", len(decoded.Reasons), len(original.Reasons))
	}
	if decoded.Evidence["tests"] != original.Evidence["tests"] {
		t.Errorf("Evidence.tests = %q, want %q", decoded.Evidence["tests"], original.Evidence["tests"])
	}
}

func TestHandoffExecutedData_ArtifactsAlwaysPresent(t *testing.T) {
	// Artifacts field in HandoffExecutedData is required (not omitempty).
	original := HandoffExecutedData{
		FromAgent: "context-architect",
		ToAgent:   "integration-engineer",
		SessionID: "s-001",
		Artifacts: []string{},
	}

	data, _ := json.Marshal(original)
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := raw["artifacts"]; !ok {
		t.Error("artifacts field should always be present in HandoffExecutedData")
	}
}

// --- Constructor functions ---

func TestNewTypedSessionCreatedEvent(t *testing.T) {
	event := NewTypedSessionCreatedEvent("", "s-001", "Add dark mode", "MODULE", "ecosystem")

	if event.Type != EventTypeSessionCreated {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionCreated)
	}
	if event.Source != SourceCLI {
		t.Errorf("Source = %q, want %q", event.Source, SourceCLI)
	}
	if event.Ts == "" {
		t.Error("Ts should not be empty")
	}

	// Verify data field contents
	var d SessionCreatedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.SessionID != "s-001" {
		t.Errorf("data.session_id = %q, want %q", d.SessionID, "s-001")
	}
	if d.Initiative != "Add dark mode" {
		t.Errorf("data.initiative = %q, want %q", d.Initiative, "Add dark mode")
	}
	if d.Complexity != "MODULE" {
		t.Errorf("data.complexity = %q, want %q", d.Complexity, "MODULE")
	}
	if d.Rite != "ecosystem" {
		t.Errorf("data.rite = %q, want %q", d.Rite, "ecosystem")
	}
}

func TestNewTypedSessionParkedEvent(t *testing.T) {
	event := NewTypedSessionParkedEvent("", "s-001", "lunch break")

	if event.Type != EventTypeSessionParked {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionParked)
	}
	if event.Source != SourceCLI {
		t.Errorf("Source = %q, want %q", event.Source, SourceCLI)
	}

	var d SessionParkedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.SessionID != "s-001" {
		t.Errorf("data.session_id = %q, want %q", d.SessionID, "s-001")
	}
	if d.Reason != "lunch break" {
		t.Errorf("data.reason = %q, want %q", d.Reason, "lunch break")
	}
}

func TestNewTypedSessionResumedEvent(t *testing.T) {
	event := NewTypedSessionResumedEvent("", "s-001")

	if event.Type != EventTypeSessionResumed {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionResumed)
	}
	if event.Source != SourceCLI {
		t.Errorf("Source = %q, want %q", event.Source, SourceCLI)
	}

	var d SessionResumedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.SessionID != "s-001" {
		t.Errorf("data.session_id = %q, want %q", d.SessionID, "s-001")
	}
}

func TestNewTypedSessionWrappedEvent(t *testing.T) {
	event := NewTypedSessionWrappedEvent("", "s-001", "WHITE", 3600000)

	if event.Type != EventTypeSessionWrapped {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionWrapped)
	}
	if event.Source != SourceCLI {
		t.Errorf("Source = %q, want %q", event.Source, SourceCLI)
	}

	var d SessionWrappedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.SailsColor != "WHITE" {
		t.Errorf("data.sails_color = %q, want WHITE", d.SailsColor)
	}
	if d.DurationMs != 3600000 {
		t.Errorf("data.duration_ms = %d, want 3600000", d.DurationMs)
	}
}

func TestNewTypedAgentDelegatedEvent_HookSource(t *testing.T) {
	event := NewTypedAgentDelegatedEvent(SourceHook, "", "context-architect", "specialist", "task-001", "agent-abc")

	if event.Type != EventTypeAgentDelegated {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeAgentDelegated)
	}
	if event.Source != SourceHook {
		t.Errorf("Source = %q, want %q", event.Source, SourceHook)
	}

	var d AgentDelegatedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.AgentName != "context-architect" {
		t.Errorf("data.agent_name = %q, want context-architect", d.AgentName)
	}
	if d.AgentType != "specialist" {
		t.Errorf("data.agent_type = %q, want specialist", d.AgentType)
	}
}

func TestNewTypedAgentDelegatedEvent_CLISource(t *testing.T) {
	// agent.delegated can also come from CLI (ari handoff execute)
	event := NewTypedAgentDelegatedEvent(SourceCLI, "", "integration-engineer", "", "", "")

	if event.Source != SourceCLI {
		t.Errorf("Source = %q, want %q", event.Source, SourceCLI)
	}
}

func TestNewTypedAgentCompletedEvent(t *testing.T) {
	artifacts := []string{"/path/to/typed_event.go", "/path/to/typed_event_test.go"}
	event := NewTypedAgentCompletedEvent(SourceHook, "", "integration-engineer", "specialist", "task-001", "agent-xyz", "success", 15000, artifacts)

	if event.Type != EventTypeAgentCompleted {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeAgentCompleted)
	}
	if event.Source != SourceHook {
		t.Errorf("Source = %q, want %q", event.Source, SourceHook)
	}

	var d AgentCompletedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.Outcome != "success" {
		t.Errorf("data.outcome = %q, want success", d.Outcome)
	}
	if d.DurationMs != 15000 {
		t.Errorf("data.duration_ms = %d, want 15000", d.DurationMs)
	}
	if len(d.Artifacts) != 2 {
		t.Errorf("data.artifacts len = %d, want 2", len(d.Artifacts))
	}
}

func TestNewTypedCommitCreatedEvent(t *testing.T) {
	event := NewTypedCommitCreatedEvent("", "abc123f", "feat: add theme provider")

	if event.Type != EventTypeCommitCreated {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeCommitCreated)
	}
	if event.Source != SourceHook {
		t.Errorf("Source = %q, want %q", event.Source, SourceHook)
	}

	var d CommitCreatedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.SHA != "abc123f" {
		t.Errorf("data.sha = %q, want abc123f", d.SHA)
	}
	if d.Message != "feat: add theme provider" {
		t.Errorf("data.message = %q, want 'feat: add theme provider'", d.Message)
	}
}

func TestNewTypedDecisionRecordedEvent(t *testing.T) {
	event := NewTypedDecisionRecordedEvent("", "Use PostgreSQL", "Better ACID", []string{"MongoDB", "SQLite"})

	if event.Type != EventTypeDecisionRecorded {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeDecisionRecorded)
	}
	if event.Source != SourceAgent {
		t.Errorf("Source = %q, want %q", event.Source, SourceAgent)
	}

	var d DecisionRecordedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.Decision != "Use PostgreSQL" {
		t.Errorf("data.decision = %q, want 'Use PostgreSQL'", d.Decision)
	}
	if len(d.Rejected) != 2 {
		t.Errorf("data.rejected len = %d, want 2", len(d.Rejected))
	}
}

func TestNewTypedCommandInvokedEvent(t *testing.T) {
	event := NewTypedCommandInvokedEvent("", "ecosystem-ref", "skill")

	if event.Type != EventTypeCommandInvoked {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeCommandInvoked)
	}
	if event.Source != SourceHook {
		t.Errorf("Source = %q, want %q", event.Source, SourceHook)
	}

	var d CommandInvokedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.Command != "ecosystem-ref" {
		t.Errorf("data.command = %q, want ecosystem-ref", d.Command)
	}
	if d.Type != "skill" {
		t.Errorf("data.type = %q, want skill", d.Type)
	}
}

func TestNewTypedHandoffExecutedEvent_NilArtifacts(t *testing.T) {
	// nil artifacts should be normalized to empty slice, not serialized as null.
	event := NewTypedHandoffExecutedEvent("", "agent-a", "agent-b", "s-001", nil)

	var d HandoffExecutedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if d.Artifacts == nil {
		t.Error("Artifacts should not be nil after normalization")
	}
	if len(d.Artifacts) != 0 {
		t.Errorf("Artifacts len = %d, want 0", len(d.Artifacts))
	}
}

func TestNewTypedFieldUpdatedEvent(t *testing.T) {
	event := NewTypedFieldUpdatedEvent("", "s-001", "complexity", "PATCH", "MODULE")

	if event.Type != EventTypeFieldUpdated {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeFieldUpdated)
	}
	if event.Source != SourceCLI {
		t.Errorf("Source = %q, want %q", event.Source, SourceCLI)
	}

	var d FieldUpdatedData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.Key != "complexity" {
		t.Errorf("data.key = %q, want complexity", d.Key)
	}
}

func TestNewTypedHookFiredEvent(t *testing.T) {
	event := NewTypedHookFiredEvent("", "clew", "PostToolUse")

	if event.Type != EventTypeHookFired {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeHookFired)
	}
	if event.Source != SourceHook {
		t.Errorf("Source = %q, want %q", event.Source, SourceHook)
	}

	var d HookFiredData
	if err := json.Unmarshal(event.Data, &d); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if d.HookName != "clew" {
		t.Errorf("data.hook_name = %q, want clew", d.HookName)
	}
	if d.EventType != "PostToolUse" {
		t.Errorf("data.event_type = %q, want PostToolUse", d.EventType)
	}
}

// --- JSONL format verification ---

func TestTypedEvent_JSONLFormat(t *testing.T) {
	// TypedEvent should serialize to a valid single-line JSON object.
	event := NewTypedSessionCreatedEvent("", "s-001", "dark mode", "MODULE", "ecosystem")

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Must be valid JSON
	var check map[string]any
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatalf("JSONL line is not valid JSON: %v", err)
	}

	// Must have all required fields for format discrimination
	if _, ok := check["ts"]; !ok {
		t.Error("ts missing from JSONL")
	}
	if _, ok := check["type"]; !ok {
		t.Error("type missing from JSONL")
	}
	if _, ok := check["source"]; !ok {
		t.Error("source missing from JSONL")
	}
	if _, ok := check["data"]; !ok {
		t.Error("data missing from JSONL -- required for v3 format detection")
	}
}

// --- Source inference ---

func TestInferSource_AllKnownTypes(t *testing.T) {
	tests := []struct {
		eventType string
		expected  EventSource
	}{
		{"session.created", SourceCLI},
		{"session.parked", SourceCLI},
		{"session.resumed", SourceCLI},
		{"session.archived", SourceCLI},
		{"session.frayed", SourceCLI},
		{"session.strand_resolved", SourceCLI},
		{"session.schema_migrated", SourceCLI},
		{"session.started", SourceHook},
		{"session.ended", SourceHook},
		{"phase.transitioned", SourceCLI},
		{"tool.call", SourceHook},
		{"tool.file_change", SourceHook},
		{"tool.artifact_created", SourceHook},
		{"tool.error", SourceHook},
		{"tool.invoked", SourceHook},
		{"file.modified", SourceHook},
		{"artifact.created", SourceHook},
		{"error.occurred", SourceHook},
		{"agent.decision", SourceAgent},
		{"decision.recorded", SourceAgent},
		{"agent.task_start", SourceHook},
		{"agent.task_end", SourceHook},
		{"agent.delegated", SourceHook},
		{"agent.completed", SourceHook},
		{"agent.handoff_prepared", SourceCLI},
		{"agent.handoff_executed", SourceCLI},
		{"quality.sails_generated", SourceCLI},
		{"lock.acquired", SourceCLI},
		{"lock.released", SourceCLI},
		{"context_switch", SourceHook},
		{"commit.created", SourceHook},
		{"command.invoked", SourceHook},
		{"hook.fired", SourceHook},
		{"field.updated", SourceCLI},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			got := InferSource(EventType(tt.eventType))
			if got != tt.expected {
				t.Errorf("InferSource(%q) = %q, want %q", tt.eventType, got, tt.expected)
			}
		})
	}
}

func TestInferSource_UnknownType_DefaultsToHook(t *testing.T) {
	// Unknown types default to SourceHook (safe default per spec).
	got := InferSource(EventType("releaser.package.published"))
	if got != SourceHook {
		t.Errorf("InferSource(unknown) = %q, want %q", got, SourceHook)
	}
}

// --- Type rename map ---

func TestRenameV2Type_AllRenames(t *testing.T) {
	tests := []struct {
		v2Type   string
		expected string
	}{
		{"tool.call", "tool.invoked"},
		{"tool.file_change", "file.modified"},
		{"tool.artifact_created", "artifact.created"},
		{"tool.error", "error.occurred"},
		{"agent.decision", "decision.recorded"},
		{"agent.task_start", "agent.delegated"},
		{"agent.task_end", "agent.completed"},
	}

	for _, tt := range tests {
		got := RenameV2Type(tt.v2Type)
		if got != tt.expected {
			t.Errorf("RenameV2Type(%q) = %q, want %q", tt.v2Type, got, tt.expected)
		}
	}
}

func TestRenameV2Type_UnchangedTypes(t *testing.T) {
	// Types not in the rename map pass through unchanged.
	unchanged := []string{
		"session.created",
		"session.parked",
		"session.ended",
		"session.started",
		"phase.transitioned",
		"quality.sails_generated",
		"lock.acquired",
		"lock.released",
		"agent.handoff_prepared",
		"agent.handoff_executed",
		"session.frayed",
		"session.strand_resolved",
		"context_switch",
		"session.schema_migrated",
		"session.archived",
	}

	for _, typ := range unchanged {
		t.Run(typ, func(t *testing.T) {
			got := RenameV2Type(typ)
			if got != typ {
				t.Errorf("RenameV2Type(%q) = %q, want %q (unchanged)", typ, got, typ)
			}
		})
	}
}

func TestRenameV2Type_V3TypesPassthrough(t *testing.T) {
	// v3 type strings that are already canonical pass through unchanged.
	v3Types := []string{
		"tool.invoked",
		"file.modified",
		"artifact.created",
		"error.occurred",
		"decision.recorded",
		"agent.delegated",
		"agent.completed",
	}

	for _, typ := range v3Types {
		t.Run(typ, func(t *testing.T) {
			got := RenameV2Type(typ)
			if got != typ {
				t.Errorf("RenameV2Type(%q) = %q, want same (v3 types should pass through)", typ, got)
			}
		})
	}
}

// --- Writer integration tests ---

func TestEventWriter_WriteTyped(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer func() { _ = writer.Close() }()

	event := NewTypedSessionCreatedEvent("", "s-001", "test initiative", "MODULE", "ecosystem")
	if err := writer.WriteTyped(event); err != nil {
		t.Fatalf("WriteTyped failed: %v", err)
	}

	// Read back the file and verify it's valid JSONL with "data" field.
	eventsPath := writer.Path()
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("failed to read events file: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data[:len(data)-1], &raw); err != nil { // strip trailing newline
		t.Fatalf("events.jsonl line is not valid JSON: %v", err)
	}

	if raw["type"] != "session.created" {
		t.Errorf("type = %v, want session.created", raw["type"])
	}
	if raw["source"] != "cli" {
		t.Errorf("source = %v, want cli", raw["source"])
	}
	if _, ok := raw["data"]; !ok {
		t.Error("data field missing from written TypedEvent")
	}
	// v2 fields must not be present in v3 events
	if _, ok := raw["meta"]; ok {
		t.Error("meta field must not be present in TypedEvent JSONL")
	}
	if _, ok := raw["summary"]; ok {
		t.Error("summary field must not be present in TypedEvent JSONL")
	}
}

func TestEventWriter_WriteTyped_InterleavedWithV2(t *testing.T) {
	// Write v2 flat events and v3 typed events to the same file.
	// The reader must distinguish them by field presence.
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer func() { _ = writer.Close() }()

	// v2 flat event
	v2Event := NewSessionCreatedEvent("s-001", "init", "PATCH", "ecosystem")
	if err := writer.Write(v2Event); err != nil {
		t.Fatalf("Write(v2) failed: %v", err)
	}

	// v3 typed event
	v3Event := NewTypedAgentDelegatedEvent(SourceHook, "", "architect", "specialist", "task-001", "")
	if err := writer.WriteTyped(v3Event); err != nil {
		t.Fatalf("WriteTyped(v3) failed: %v", err)
	}

	// Read back all lines
	eventsPath := writer.Path()
	f, err := os.Open(eventsPath)
	if err != nil {
		t.Fatalf("failed to open events file: %v", err)
	}
	defer f.Close()

	var lines []map[string]any
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var raw map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			t.Fatalf("events.jsonl line is not valid JSON: %v", err)
		}
		lines = append(lines, raw)
	}

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Line 1: v2 event -- has "meta", no "data"
	if _, ok := lines[0]["meta"]; !ok {
		t.Error("v2 line should have meta field")
	}
	if _, ok := lines[0]["data"]; ok {
		t.Error("v2 line should NOT have data field")
	}

	// Line 2: v3 event -- has "data" and "source", no "meta"
	if _, ok := lines[1]["data"]; !ok {
		t.Error("v3 line should have data field")
	}
	if _, ok := lines[1]["source"]; !ok {
		t.Error("v3 line should have source field")
	}
	if _, ok := lines[1]["meta"]; ok {
		t.Error("v3 line should NOT have meta field")
	}
}

func TestBufferedEventWriter_WriteTyped(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewBufferedEventWriter(tmpDir, 100*time.Millisecond)

	// Write a v3 typed event
	event := NewTypedCommitCreatedEvent("", "abc123", "feat: test commit")
	w.WriteTyped(event)

	if w.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after WriteTyped", w.Len())
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify the event was written to disk
	eventsPath := w.Path()
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("failed to read events file: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data[:len(data)-1], &raw); err != nil {
		t.Fatalf("events.jsonl line is not valid JSON: %v", err)
	}

	if raw["type"] != "commit.created" {
		t.Errorf("type = %v, want commit.created", raw["type"])
	}
	if _, ok := raw["data"]; !ok {
		t.Error("data field missing from written TypedEvent")
	}
}

func TestBufferedEventWriter_WriteTyped_InterleavedWithV2(t *testing.T) {
	// Verify that both v2 and v3 events can be buffered and flushed together.
	tmpDir := t.TempDir()
	w := NewBufferedEventWriter(tmpDir, 100*time.Millisecond)

	// Mix v2 and v3 events
	w.Write(NewSessionCreatedEvent("s-001", "init", "PATCH", ""))
	w.WriteTyped(NewTypedCommitCreatedEvent("", "abc123", "feat: first commit"))
	w.Write(NewSessionParkedEvent("s-001", "lunch"))
	w.WriteTyped(NewTypedAgentDelegatedEvent(SourceHook, "", "architect", "", "", ""))

	if w.Len() != 4 {
		t.Errorf("Len() = %d, want 4 (2 v2 + 2 v3)", w.Len())
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Read back and count lines
	eventsPath := w.Path()
	f, err := os.Open(eventsPath)
	if err != nil {
		t.Fatalf("failed to open events file: %v", err)
	}
	defer f.Close()

	var lineCount int
	var v2Count, v3Count int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineCount++
		var raw map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			t.Fatalf("invalid JSON on line %d: %v", lineCount, err)
		}
		if _, hasData := raw["data"]; hasData {
			v3Count++
		} else {
			v2Count++
		}
	}

	if lineCount != 4 {
		t.Errorf("expected 4 lines written, got %d", lineCount)
	}
	if v2Count != 2 {
		t.Errorf("expected 2 v2 events, got %d", v2Count)
	}
	if v3Count != 2 {
		t.Errorf("expected 2 v3 events, got %d", v3Count)
	}
}

func TestBufferedEventWriter_Len_CountsBothTypes(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewBufferedEventWriter(tmpDir, 10*time.Second) // long interval so no auto-flush

	w.Write(NewSessionCreatedEvent("s-001", "init", "PATCH", ""))
	w.WriteTyped(NewTypedCommitCreatedEvent("", "abc123", "test"))
	w.WriteTyped(NewTypedAgentDelegatedEvent(SourceHook, "", "arch", "", "", ""))

	if w.Len() != 3 {
		t.Errorf("Len() = %d, want 3 (1 v2 + 2 v3)", w.Len())
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

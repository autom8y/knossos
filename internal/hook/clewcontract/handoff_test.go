package clewcontract

import (
	"encoding/json"
	"testing"
)

func TestEventTypeHandoffPrepared_Constant(t *testing.T) {
	if EventTypeHandoffPrepared != "agent.handoff_prepared" {
		t.Errorf("EventTypeHandoffPrepared = %q, want %q", EventTypeHandoffPrepared, "agent.handoff_prepared")
	}
}

func TestEventTypeHandoffExecuted_Constant(t *testing.T) {
	if EventTypeHandoffExecuted != "agent.handoff_executed" {
		t.Errorf("EventTypeHandoffExecuted = %q, want %q", EventTypeHandoffExecuted, "agent.handoff_executed")
	}
}

func TestNewHandoffPreparedEvent(t *testing.T) {
	event := NewHandoffPreparedEvent("architect", "integration-engineer", "session-123")

	if event.Type != EventTypeHandoffPrepared {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeHandoffPrepared)
	}

	if event.Summary != "Handoff prepared: architect -> integration-engineer" {
		t.Errorf("Summary = %q, want %q", event.Summary, "Handoff prepared: architect -> integration-engineer")
	}

	// Check meta fields
	if fromAgent, ok := event.Meta["from_agent"].(string); !ok || fromAgent != "architect" {
		t.Errorf("Meta[from_agent] = %v, want %q", event.Meta["from_agent"], "architect")
	}

	if toAgent, ok := event.Meta["to_agent"].(string); !ok || toAgent != "integration-engineer" {
		t.Errorf("Meta[to_agent] = %v, want %q", event.Meta["to_agent"], "integration-engineer")
	}

	if sessionID, ok := event.Meta["session_id"].(string); !ok || sessionID != "session-123" {
		t.Errorf("Meta[session_id] = %v, want %q", event.Meta["session_id"], "session-123")
	}

	// Check timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestNewHandoffExecutedEvent(t *testing.T) {
	artifacts := []string{
		"/path/to/CONTEXT-DESIGN-foo.md",
		"/path/to/sync-core.sh",
	}
	event := NewHandoffExecutedEvent("architect", "integration-engineer", "session-123", artifacts)

	if event.Type != EventTypeHandoffExecuted {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeHandoffExecuted)
	}

	expectedSummary := "Handoff executed: architect -> integration-engineer (2 artifacts)"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %q, want %q", event.Summary, expectedSummary)
	}

	// Check meta fields
	if fromAgent, ok := event.Meta["from_agent"].(string); !ok || fromAgent != "architect" {
		t.Errorf("Meta[from_agent] = %v, want %q", event.Meta["from_agent"], "architect")
	}

	if toAgent, ok := event.Meta["to_agent"].(string); !ok || toAgent != "integration-engineer" {
		t.Errorf("Meta[to_agent] = %v, want %q", event.Meta["to_agent"], "integration-engineer")
	}

	if sessionID, ok := event.Meta["session_id"].(string); !ok || sessionID != "session-123" {
		t.Errorf("Meta[session_id] = %v, want %q", event.Meta["session_id"], "session-123")
	}

	// Check artifacts array
	if artifactsMeta, ok := event.Meta["artifacts"].([]string); !ok {
		t.Errorf("Meta[artifacts] type = %T, want []string", event.Meta["artifacts"])
	} else if len(artifactsMeta) != 2 {
		t.Errorf("len(Meta[artifacts]) = %d, want 2", len(artifactsMeta))
	} else if artifactsMeta[0] != artifacts[0] || artifactsMeta[1] != artifacts[1] {
		t.Errorf("Meta[artifacts] = %v, want %v", artifactsMeta, artifacts)
	}

	// Check timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestNewHandoffExecutedEvent_EmptyArtifacts(t *testing.T) {
	event := NewHandoffExecutedEvent("architect", "integration-engineer", "session-123", []string{})

	expectedSummary := "Handoff executed: architect -> integration-engineer (0 artifacts)"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %q, want %q", event.Summary, expectedSummary)
	}

	// Check artifacts array exists but is empty
	if artifactsMeta, ok := event.Meta["artifacts"].([]string); !ok {
		t.Errorf("Meta[artifacts] type = %T, want []string", event.Meta["artifacts"])
	} else if len(artifactsMeta) != 0 {
		t.Errorf("len(Meta[artifacts]) = %d, want 0", len(artifactsMeta))
	}
}

func TestHandoffPreparedEvent_JSONMarshaling(t *testing.T) {
	event := NewHandoffPreparedEvent("architect", "integration-engineer", "session-123")

	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Check required fields
	if result["type"] != "agent.handoff_prepared" {
		t.Errorf("type = %v, want %q", result["type"], "agent.handoff_prepared")
	}

	if result["summary"] != "Handoff prepared: architect -> integration-engineer" {
		t.Errorf("summary = %v", result["summary"])
	}

	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta is not a map: %T", result["meta"])
	}

	if meta["from_agent"] != "architect" {
		t.Errorf("meta.from_agent = %v, want %q", meta["from_agent"], "architect")
	}

	if meta["to_agent"] != "integration-engineer" {
		t.Errorf("meta.to_agent = %v, want %q", meta["to_agent"], "integration-engineer")
	}

	if meta["session_id"] != "session-123" {
		t.Errorf("meta.session_id = %v, want %q", meta["session_id"], "session-123")
	}
}

func TestHandoffExecutedEvent_JSONMarshaling(t *testing.T) {
	artifacts := []string{
		"/path/to/CONTEXT-DESIGN-foo.md",
		"/path/to/sync-core.sh",
	}
	event := NewHandoffExecutedEvent("architect", "integration-engineer", "session-123", artifacts)

	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Check required fields
	if result["type"] != "agent.handoff_executed" {
		t.Errorf("type = %v, want %q", result["type"], "agent.handoff_executed")
	}

	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta is not a map: %T", result["meta"])
	}

	artifactsMeta, ok := meta["artifacts"].([]interface{})
	if !ok {
		t.Fatalf("meta.artifacts is not an array: %T", meta["artifacts"])
	}

	if len(artifactsMeta) != 2 {
		t.Errorf("len(meta.artifacts) = %d, want 2", len(artifactsMeta))
	}

	if artifactsMeta[0] != artifacts[0] {
		t.Errorf("meta.artifacts[0] = %v, want %q", artifactsMeta[0], artifacts[0])
	}

	if artifactsMeta[1] != artifacts[1] {
		t.Errorf("meta.artifacts[1] = %v, want %q", artifactsMeta[1], artifacts[1])
	}
}

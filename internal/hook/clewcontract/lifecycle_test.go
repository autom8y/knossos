package clewcontract

import (
	"testing"
)

// TestNewSessionCreatedEvent verifies the session.created event constructor.
func TestNewSessionCreatedEvent(t *testing.T) {
	sessionID := "session-test-123"
	initiative := "test-initiative"
	complexity := "standard"
	rite := "10x-dev"

	event := NewSessionCreatedEvent(sessionID, initiative, complexity, rite)

	if event.Type != EventTypeSessionCreated {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionCreated)
	}

	expectedSummary := "Session created: " + sessionID
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %q, want %q", event.Summary, expectedSummary)
	}

	// Verify timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Verify metadata keys
	requiredKeys := []string{"session_id", "initiative", "complexity", "rite", "from", "to"}
	for _, key := range requiredKeys {
		if _, ok := event.Meta[key]; !ok {
			t.Errorf("Meta missing required key: %s", key)
		}
	}

	// Verify metadata values
	if event.Meta["session_id"] != sessionID {
		t.Errorf("Meta[session_id] = %q, want %q", event.Meta["session_id"], sessionID)
	}
	if event.Meta["initiative"] != initiative {
		t.Errorf("Meta[initiative] = %q, want %q", event.Meta["initiative"], initiative)
	}
	if event.Meta["complexity"] != complexity {
		t.Errorf("Meta[complexity] = %q, want %q", event.Meta["complexity"], complexity)
	}
	if event.Meta["rite"] != rite {
		t.Errorf("Meta[rite] = %q, want %q", event.Meta["rite"], rite)
	}
	if event.Meta["from"] != "NONE" {
		t.Errorf("Meta[from] = %q, want %q", event.Meta["from"], "NONE")
	}
	if event.Meta["to"] != "ACTIVE" {
		t.Errorf("Meta[to] = %q, want %q", event.Meta["to"], "ACTIVE")
	}
}

// TestNewSessionParkedEvent verifies the session.parked event constructor.
func TestNewSessionParkedEvent(t *testing.T) {
	sessionID := "session-test-456"
	reason := "manual park"

	event := NewSessionParkedEvent(sessionID, reason)

	if event.Type != EventTypeSessionParked {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionParked)
	}

	expectedSummary := "Session parked: " + sessionID
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %q, want %q", event.Summary, expectedSummary)
	}

	// Verify timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Verify metadata keys
	requiredKeys := []string{"session_id", "reason", "from", "to"}
	for _, key := range requiredKeys {
		if _, ok := event.Meta[key]; !ok {
			t.Errorf("Meta missing required key: %s", key)
		}
	}

	// Verify metadata values
	if event.Meta["session_id"] != sessionID {
		t.Errorf("Meta[session_id] = %q, want %q", event.Meta["session_id"], sessionID)
	}
	if event.Meta["reason"] != reason {
		t.Errorf("Meta[reason] = %q, want %q", event.Meta["reason"], reason)
	}
	if event.Meta["from"] != "ACTIVE" {
		t.Errorf("Meta[from] = %q, want %q", event.Meta["from"], "ACTIVE")
	}
	if event.Meta["to"] != "PARKED" {
		t.Errorf("Meta[to] = %q, want %q", event.Meta["to"], "PARKED")
	}
}

// TestNewSessionResumedEvent verifies the session.resumed event constructor.
func TestNewSessionResumedEvent(t *testing.T) {
	sessionID := "session-test-789"

	event := NewSessionResumedEvent(sessionID)

	if event.Type != EventTypeSessionResumed {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionResumed)
	}

	expectedSummary := "Session resumed: " + sessionID
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %q, want %q", event.Summary, expectedSummary)
	}

	// Verify timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Verify metadata keys
	requiredKeys := []string{"session_id", "from", "to"}
	for _, key := range requiredKeys {
		if _, ok := event.Meta[key]; !ok {
			t.Errorf("Meta missing required key: %s", key)
		}
	}

	// Verify metadata values
	if event.Meta["session_id"] != sessionID {
		t.Errorf("Meta[session_id] = %q, want %q", event.Meta["session_id"], sessionID)
	}
	if event.Meta["from"] != "PARKED" {
		t.Errorf("Meta[from] = %q, want %q", event.Meta["from"], "PARKED")
	}
	if event.Meta["to"] != "ACTIVE" {
		t.Errorf("Meta[to] = %q, want %q", event.Meta["to"], "ACTIVE")
	}
}

// TestNewSessionArchivedEvent verifies the session.archived event constructor.
func TestNewSessionArchivedEvent(t *testing.T) {
	sessionID := "session-test-abc"
	fromStatus := "PARKED"

	event := NewSessionArchivedEvent(sessionID, fromStatus)

	if event.Type != EventTypeSessionArchived {
		t.Errorf("Type = %q, want %q", event.Type, EventTypeSessionArchived)
	}

	expectedSummary := "Session archived: " + sessionID
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %q, want %q", event.Summary, expectedSummary)
	}

	// Verify timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Verify metadata keys
	requiredKeys := []string{"session_id", "from", "to"}
	for _, key := range requiredKeys {
		if _, ok := event.Meta[key]; !ok {
			t.Errorf("Meta missing required key: %s", key)
		}
	}

	// Verify metadata values
	if event.Meta["session_id"] != sessionID {
		t.Errorf("Meta[session_id] = %q, want %q", event.Meta["session_id"], sessionID)
	}
	if event.Meta["from"] != fromStatus {
		t.Errorf("Meta[from] = %q, want %q", event.Meta["from"], fromStatus)
	}
	if event.Meta["to"] != "ARCHIVED" {
		t.Errorf("Meta[to] = %q, want %q", event.Meta["to"], "ARCHIVED")
	}

	// Test with different fromStatus value
	event2 := NewSessionArchivedEvent(sessionID, "ACTIVE")
	if event2.Meta["from"] != "ACTIVE" {
		t.Errorf("Meta[from] with ACTIVE = %q, want %q", event2.Meta["from"], "ACTIVE")
	}
}

// TestNewPhaseTransitionedEvent verifies the phase.transitioned event constructor.
func TestNewPhaseTransitionedEvent(t *testing.T) {
	sessionID := "session-test-def"
	fromPhase := "design"
	toPhase := "implementation"

	event := NewPhaseTransitionedEvent(sessionID, fromPhase, toPhase)

	if event.Type != EventTypePhaseTransitioned {
		t.Errorf("Type = %q, want %q", event.Type, EventTypePhaseTransitioned)
	}

	expectedSummary := "Phase transitioned: design -> implementation"
	if event.Summary != expectedSummary {
		t.Errorf("Summary = %q, want %q", event.Summary, expectedSummary)
	}

	// Verify timestamp is set
	if event.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Verify metadata keys
	requiredKeys := []string{"session_id", "from_phase", "to_phase"}
	for _, key := range requiredKeys {
		if _, ok := event.Meta[key]; !ok {
			t.Errorf("Meta missing required key: %s", key)
		}
	}

	// Verify metadata values
	if event.Meta["session_id"] != sessionID {
		t.Errorf("Meta[session_id] = %q, want %q", event.Meta["session_id"], sessionID)
	}
	if event.Meta["from_phase"] != fromPhase {
		t.Errorf("Meta[from_phase] = %q, want %q", event.Meta["from_phase"], fromPhase)
	}
	if event.Meta["to_phase"] != toPhase {
		t.Errorf("Meta[to_phase] = %q, want %q", event.Meta["to_phase"], toPhase)
	}
}

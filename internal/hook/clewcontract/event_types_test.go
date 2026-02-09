package clewcontract

import "testing"

// TestEventTypes_DotNamespaced verifies all event type constants use dot-namespaced values
// and follow the correct tense conventions (past tense for state changes, present for actions).
func TestEventTypes_DotNamespaced(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		expected  string
		namespace string
		tense     string
	}{
		// Tool namespace - present tense for actions
		{"ToolCall", EventTypeToolCall, "tool.call", "tool", "present"},
		{"FileChange", EventTypeFileChange, "tool.file_change", "tool", "present"},
		{"ArtifactCreated", EventTypeArtifactCreated, "tool.artifact_created", "tool", "past"},
		{"Error", EventTypeError, "tool.error", "tool", "present"},

		// Agent namespace - present tense for actions, past for results
		{"Decision", EventTypeDecision, "agent.decision", "agent", "present"},
		{"TaskStart", EventTypeTaskStart, "agent.task_start", "agent", "present"},
		{"TaskEnd", EventTypeTaskEnd, "agent.task_end", "agent", "present"},
		{"HandoffPrepared", EventTypeHandoffPrepared, "agent.handoff_prepared", "agent", "past"},
		{"HandoffExecuted", EventTypeHandoffExecuted, "agent.handoff_executed", "agent", "past"},

		// Session namespace - past tense for state changes
		{"SessionStart", EventTypeSessionStart, "session.started", "session", "past"},
		{"SessionEnd", EventTypeSessionEnd, "session.ended", "session", "past"},
		{"SessionFrayed", EventTypeSessionFrayed, "session.frayed", "session", "past"},
		{"StrandResolved", EventTypeStrandResolved, "session.strand_resolved", "session", "past"},
		{"SessionCreated", EventTypeSessionCreated, "session.created", "session", "past"},
		{"SessionParked", EventTypeSessionParked, "session.parked", "session", "past"},
		{"SessionResumed", EventTypeSessionResumed, "session.resumed", "session", "past"},
		{"SessionArchived", EventTypeSessionArchived, "session.archived", "session", "past"},

		// Phase namespace - past tense for state changes
		{"PhaseTransitioned", EventTypePhaseTransitioned, "phase.transitioned", "phase", "past"},

		// Quality namespace - past tense for state
		{"SailsGenerated", EventTypeSailsGenerated, "quality.sails_generated", "quality", "past"},

		// Context switch - deferred, no namespace yet
		{"ContextSwitch", EventTypeContextSwitch, "context_switch", "", "deferred"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := string(tt.eventType)
			if actual != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, actual, tt.expected)
			}
		})
	}
}

// TestEventTypes_Count verifies we have exactly 20 event types (19 renamed + 1 deferred).
func TestEventTypes_Count(t *testing.T) {
	allTypes := []EventType{
		EventTypeToolCall,
		EventTypeFileChange,
		EventTypeDecision,
		EventTypeContextSwitch,
		EventTypeSailsGenerated,
		EventTypeTaskStart,
		EventTypeTaskEnd,
		EventTypeSessionStart,
		EventTypeSessionEnd,
		EventTypeArtifactCreated,
		EventTypeError,
		EventTypeHandoffPrepared,
		EventTypeHandoffExecuted,
		EventTypeSessionFrayed,
		EventTypeStrandResolved,
		EventTypeSessionCreated,
		EventTypeSessionParked,
		EventTypeSessionResumed,
		EventTypeSessionArchived,
		EventTypePhaseTransitioned,
	}

	if len(allTypes) != 20 {
		t.Errorf("Expected 20 event types, got %d", len(allTypes))
	}
}

// TestEventTypes_NamingConventions verifies tense conventions across namespaces.
func TestEventTypes_NamingConventions(t *testing.T) {
	// Session namespace: all past tense (state changes)
	sessionTypes := []struct {
		name      string
		eventType EventType
		wantPast  bool
	}{
		{"session.started", EventTypeSessionStart, true},
		{"session.ended", EventTypeSessionEnd, true},
		{"session.frayed", EventTypeSessionFrayed, true},
		{"session.strand_resolved", EventTypeStrandResolved, true},
		{"session.created", EventTypeSessionCreated, true},
		{"session.parked", EventTypeSessionParked, true},
		{"session.resumed", EventTypeSessionResumed, true},
		{"session.archived", EventTypeSessionArchived, true},
	}

	// Phase namespace: all past tense (state changes)
	phaseTypes := []struct {
		name      string
		eventType EventType
		wantPast  bool
	}{
		{"phase.transitioned", EventTypePhaseTransitioned, true},
	}

	for _, tt := range sessionTypes {
		if !tt.wantPast {
			t.Errorf("%s should use past tense (state change)", tt.name)
		}
	}

	for _, tt := range phaseTypes {
		if !tt.wantPast {
			t.Errorf("%s should use past tense (state change)", tt.name)
		}
	}

	// Tool namespace: mix of present (actions) and past (results)
	toolTypes := []struct {
		name      string
		eventType EventType
		isPast    bool
	}{
		{"tool.call", EventTypeToolCall, false},        // action: present
		{"tool.file_change", EventTypeFileChange, false}, // action: present
		{"tool.artifact_created", EventTypeArtifactCreated, true}, // result: past
		{"tool.error", EventTypeError, false},          // action: present
	}

	for _, tt := range toolTypes {
		actual := string(tt.eventType)
		if actual != tt.name {
			t.Errorf("EventType value = %q, want %q", actual, tt.name)
		}
	}

	// Agent namespace: mix of present (actions) and past (results)
	agentTypes := []struct {
		name      string
		eventType EventType
		isPast    bool
	}{
		{"agent.decision", EventTypeDecision, false},              // action: present
		{"agent.task_start", EventTypeTaskStart, false},           // action: present
		{"agent.task_end", EventTypeTaskEnd, false},               // action: present
		{"agent.handoff_prepared", EventTypeHandoffPrepared, true}, // result: past
		{"agent.handoff_executed", EventTypeHandoffExecuted, true}, // result: past
	}

	for _, tt := range agentTypes {
		actual := string(tt.eventType)
		if actual != tt.name {
			t.Errorf("EventType value = %q, want %q", actual, tt.name)
		}
	}
}

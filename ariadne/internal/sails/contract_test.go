package sails

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/ariadne/internal/hook/clewcontract"
)

// TestValidateThreadContract_ValidSequences tests that valid event sequences pass validation.
func TestValidateThreadContract_ValidSequences(t *testing.T) {
	tests := []struct {
		name   string
		events []clewcontract.Event
	}{
		{
			name: "valid handoff sequence",
			events: []clewcontract.Event{
				clewcontract.NewHandoffPreparedEvent("agent-a", "agent-b", "session-123"),
				clewcontract.NewHandoffExecutedEvent("agent-a", "agent-b", "session-123", []string{}),
			},
		},
		{
			name: "valid task lifecycle",
			events: []clewcontract.Event{
				clewcontract.NewTaskStartEvent("task-001", "agent-a", "design", "session-123"),
				clewcontract.NewTaskEndEvent("task-001", "agent-a", "success", "session-123", 1000, []string{}),
			},
		},
		{
			name: "multiple valid handoffs",
			events: []clewcontract.Event{
				clewcontract.NewHandoffPreparedEvent("agent-a", "agent-b", "session-123"),
				clewcontract.NewHandoffExecutedEvent("agent-a", "agent-b", "session-123", []string{}),
				clewcontract.NewHandoffPreparedEvent("agent-b", "agent-c", "session-123"),
				clewcontract.NewHandoffExecutedEvent("agent-b", "agent-c", "session-123", []string{}),
			},
		},
		{
			name: "multiple valid tasks",
			events: []clewcontract.Event{
				clewcontract.NewTaskStartEvent("task-001", "agent-a", "design", "session-123"),
				clewcontract.NewTaskEndEvent("task-001", "agent-a", "success", "session-123", 1000, []string{}),
				clewcontract.NewTaskStartEvent("task-002", "agent-b", "implementation", "session-123"),
				clewcontract.NewTaskEndEvent("task-002", "agent-b", "success", "session-123", 2000, []string{}),
			},
		},
		{
			name:   "empty events file",
			events: []clewcontract.Event{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory with events.jsonl
			tempDir := t.TempDir()
			eventsPath := filepath.Join(tempDir, "events.jsonl")
			writeEventsFile(t, eventsPath, tt.events)

			// Validate
			violations, err := ValidateThreadContract(tempDir)
			if err != nil {
				t.Fatalf("ValidateThreadContract() error = %v", err)
			}

			if len(violations) > 0 {
				t.Errorf("ValidateThreadContract() expected no violations, got %d: %+v", len(violations), violations)
			}
		})
	}
}

// TestValidateThreadContract_HandoffViolations tests handoff-related violations.
func TestValidateThreadContract_HandoffViolations(t *testing.T) {
	tests := []struct {
		name              string
		events            []clewcontract.Event
		expectedViolation string
	}{
		{
			name: "handoff_executed without handoff_prepared",
			events: []clewcontract.Event{
				clewcontract.NewHandoffExecutedEvent("agent-a", "agent-b", "session-123", []string{}),
			},
			expectedViolation: "handoff_unprepared",
		},
		{
			name: "handoff_executed with different agent pair",
			events: []clewcontract.Event{
				clewcontract.NewHandoffPreparedEvent("agent-a", "agent-b", "session-123"),
				clewcontract.NewHandoffExecutedEvent("agent-a", "agent-c", "session-123", []string{}),
			},
			expectedViolation: "handoff_unprepared",
		},
		{
			name: "handoff_executed missing metadata",
			events: []clewcontract.Event{
				{
					Type:    clewcontract.EventTypeHandoffExecuted,
					Summary: "Handoff executed",
					Meta:    map[string]interface{}{},
				},
			},
			expectedViolation: "handoff_missing_metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory with events.jsonl
			tempDir := t.TempDir()
			eventsPath := filepath.Join(tempDir, "events.jsonl")
			writeEventsFile(t, eventsPath, tt.events)

			// Validate
			violations, err := ValidateThreadContract(tempDir)
			if err != nil {
				t.Fatalf("ValidateThreadContract() error = %v", err)
			}

			if len(violations) == 0 {
				t.Fatalf("ValidateThreadContract() expected violations, got none")
			}

			// Check that the expected violation type is present
			found := false
			for _, v := range violations {
				if v.Type == tt.expectedViolation {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("ValidateThreadContract() expected violation type %q, got violations: %+v", tt.expectedViolation, violations)
			}
		})
	}
}

// TestValidateThreadContract_TaskViolations tests task lifecycle violations.
func TestValidateThreadContract_TaskViolations(t *testing.T) {
	tests := []struct {
		name              string
		events            []clewcontract.Event
		expectedViolation string
	}{
		{
			name: "task_end without task_start",
			events: []clewcontract.Event{
				clewcontract.NewTaskEndEvent("task-001", "agent-a", "success", "session-123", 1000, []string{}),
			},
			expectedViolation: "task_orphaned_end",
		},
		{
			name: "task_end with different task_id",
			events: []clewcontract.Event{
				clewcontract.NewTaskStartEvent("task-001", "agent-a", "design", "session-123"),
				clewcontract.NewTaskEndEvent("task-002", "agent-a", "success", "session-123", 1000, []string{}),
			},
			expectedViolation: "task_orphaned_end",
		},
		{
			name: "task_start missing task_id",
			events: []clewcontract.Event{
				{
					Type:    clewcontract.EventTypeTaskStart,
					Summary: "Task started",
					Meta:    map[string]interface{}{},
				},
			},
			expectedViolation: "task_missing_id",
		},
		{
			name: "task_end missing task_id",
			events: []clewcontract.Event{
				{
					Type:    clewcontract.EventTypeTaskEnd,
					Summary: "Task ended",
					Meta:    map[string]interface{}{},
				},
			},
			expectedViolation: "task_missing_id",
		},
		{
			name: "duplicate task_start",
			events: []clewcontract.Event{
				clewcontract.NewTaskStartEvent("task-001", "agent-a", "design", "session-123"),
				clewcontract.NewTaskStartEvent("task-001", "agent-a", "design", "session-123"),
			},
			expectedViolation: "task_duplicate_start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory with events.jsonl
			tempDir := t.TempDir()
			eventsPath := filepath.Join(tempDir, "events.jsonl")
			writeEventsFile(t, eventsPath, tt.events)

			// Validate
			violations, err := ValidateThreadContract(tempDir)
			if err != nil {
				t.Fatalf("ValidateThreadContract() error = %v", err)
			}

			if len(violations) == 0 {
				t.Fatalf("ValidateThreadContract() expected violations, got none")
			}

			// Check that the expected violation type is present
			found := false
			for _, v := range violations {
				if v.Type == tt.expectedViolation {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("ValidateThreadContract() expected violation type %q, got violations: %+v", tt.expectedViolation, violations)
			}
		})
	}
}

// TestValidateThreadContract_NoEventsFile tests behavior when events.jsonl doesn't exist.
func TestValidateThreadContract_NoEventsFile(t *testing.T) {
	// Create temporary directory without events.jsonl
	tempDir := t.TempDir()

	// Validate
	violations, err := ValidateThreadContract(tempDir)
	if err != nil {
		t.Fatalf("ValidateThreadContract() error = %v", err)
	}

	// No events file should not be a violation
	if len(violations) > 0 {
		t.Errorf("ValidateThreadContract() expected no violations for missing events.jsonl, got %d: %+v", len(violations), violations)
	}
}

// TestValidateThreadContract_MalformedJSON tests behavior with malformed events.jsonl.
func TestValidateThreadContract_MalformedJSON(t *testing.T) {
	// Create temporary directory with malformed events.jsonl
	tempDir := t.TempDir()
	eventsPath := filepath.Join(tempDir, "events.jsonl")

	// Write malformed JSON
	err := os.WriteFile(eventsPath, []byte(`{"type": "task_start", "ts": "invalid`), 0644)
	if err != nil {
		t.Fatalf("Failed to write malformed events.jsonl: %v", err)
	}

	// Validate
	_, err = ValidateThreadContract(tempDir)
	if err == nil {
		t.Error("ValidateThreadContract() expected error for malformed JSON, got nil")
	}
}

// TestCheckGate_ContractViolationsDowngradeWhite tests that violations downgrade WHITE to GRAY.
func TestCheckGate_ContractViolationsDowngradeWhite(t *testing.T) {
	// Create temporary session directory
	tempDir := t.TempDir()

	// Write WHITE_SAILS.yaml with WHITE color
	sailsPath := filepath.Join(tempDir, "WHITE_SAILS.yaml")
	sailsContent := `version: "1.0"
session_id: "session-123"
color: WHITE
computed_base: WHITE
type: standard
complexity: MODULE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
`
	err := os.WriteFile(sailsPath, []byte(sailsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
	}

	// Write events.jsonl with a violation (orphaned task_end)
	eventsPath := filepath.Join(tempDir, "events.jsonl")
	events := []clewcontract.Event{
		clewcontract.NewTaskEndEvent("task-001", "agent-a", "success", "session-123", 1000, []string{}),
	}
	writeEventsFile(t, eventsPath, events)

	// Check gate
	result, err := CheckGate(tempDir)
	if err != nil {
		t.Fatalf("CheckGate() error = %v", err)
	}

	// Verify that color was downgraded to GRAY
	if result.Color != ColorGray {
		t.Errorf("CheckGate() expected color GRAY due to violations, got %s", result.Color)
	}

	// Verify that pass is false
	if result.Pass {
		t.Error("CheckGate() expected pass=false due to violations, got true")
	}

	// Verify that violations are present
	if len(result.ContractViolations) == 0 {
		t.Error("CheckGate() expected contract violations, got none")
	}

	// Verify that reasons mention the downgrade
	foundDowngradeReason := false
	for _, reason := range result.Reasons {
		if reason == "thread contract violations present: downgraded to GRAY" {
			foundDowngradeReason = true
			break
		}
	}
	if !foundDowngradeReason {
		t.Errorf("CheckGate() expected downgrade reason in reasons, got: %+v", result.Reasons)
	}
}

// TestCheckGate_ContractViolationsPreserveGray tests that violations don't affect already GRAY sails.
func TestCheckGate_ContractViolationsPreserveGray(t *testing.T) {
	// Create temporary session directory
	tempDir := t.TempDir()

	// Write WHITE_SAILS.yaml with GRAY color
	sailsPath := filepath.Join(tempDir, "WHITE_SAILS.yaml")
	sailsContent := `version: "1.0"
session_id: "session-123"
color: GRAY
computed_base: GRAY
type: standard
complexity: MODULE
open_questions:
  - "What is the performance impact?"
`
	err := os.WriteFile(sailsPath, []byte(sailsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
	}

	// Write events.jsonl with a violation (orphaned task_end)
	eventsPath := filepath.Join(tempDir, "events.jsonl")
	events := []clewcontract.Event{
		clewcontract.NewTaskEndEvent("task-001", "agent-a", "success", "session-123", 1000, []string{}),
	}
	writeEventsFile(t, eventsPath, events)

	// Check gate
	result, err := CheckGate(tempDir)
	if err != nil {
		t.Fatalf("CheckGate() error = %v", err)
	}

	// Verify that color remains GRAY
	if result.Color != ColorGray {
		t.Errorf("CheckGate() expected color GRAY, got %s", result.Color)
	}

	// Verify that violations are present
	if len(result.ContractViolations) == 0 {
		t.Error("CheckGate() expected contract violations, got none")
	}
}

// TestCheckGate_ContractViolationsPreserveBlack tests that violations don't affect BLACK sails.
func TestCheckGate_ContractViolationsPreserveBlack(t *testing.T) {
	// Create temporary session directory
	tempDir := t.TempDir()

	// Write WHITE_SAILS.yaml with BLACK color
	sailsPath := filepath.Join(tempDir, "WHITE_SAILS.yaml")
	sailsContent := `version: "1.0"
session_id: "session-123"
color: BLACK
computed_base: BLACK
type: standard
complexity: MODULE
proofs:
  tests:
    status: FAIL
`
	err := os.WriteFile(sailsPath, []byte(sailsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
	}

	// Write events.jsonl with a violation (orphaned task_end)
	eventsPath := filepath.Join(tempDir, "events.jsonl")
	events := []clewcontract.Event{
		clewcontract.NewTaskEndEvent("task-001", "agent-a", "success", "session-123", 1000, []string{}),
	}
	writeEventsFile(t, eventsPath, events)

	// Check gate
	result, err := CheckGate(tempDir)
	if err != nil {
		t.Fatalf("CheckGate() error = %v", err)
	}

	// Verify that color remains BLACK
	if result.Color != ColorBlack {
		t.Errorf("CheckGate() expected color BLACK, got %s", result.Color)
	}

	// Verify that violations are present
	if len(result.ContractViolations) == 0 {
		t.Error("CheckGate() expected contract violations, got none")
	}
}

// writeEventsFile writes events to events.jsonl in JSON Lines format.
func writeEventsFile(t *testing.T, path string, events []clewcontract.Event) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create events.jsonl: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			t.Fatalf("Failed to write event: %v", err)
		}
	}
}

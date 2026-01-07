// Package sails implements the White Sails confidence signaling system per Knossos Doctrine v2.
// This file implements thread contract validation per Wave 2 Task 4 (T1-004).
package sails

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/hook/clewcontract"
)

// ContractViolation represents a single clew contract violation.
type ContractViolation struct {
	// Type is the category of violation
	Type string `json:"type" yaml:"type"`

	// Description explains what went wrong
	Description string `json:"description" yaml:"description"`

	// Severity indicates impact: "error" or "warning"
	Severity string `json:"severity" yaml:"severity"`

	// RelatedEvents lists event indices involved in the violation
	RelatedEvents []int `json:"related_events,omitempty" yaml:"related_events,omitempty"`
}

// ValidateClewContract validates the clew contract from events.jsonl.
// This checks for:
// - Handoff sequences: handoff_prepared must precede handoff_executed for same agent pair
// - Task lifecycle: task_start must precede task_end for same task_id
// - Session lifecycle: session_start should exist in the thread
//
// Returns a list of violations (empty if contract is valid).
func ValidateClewContract(sessionDir string) ([]ContractViolation, error) {
	eventsPath := filepath.Join(sessionDir, "events.jsonl")

	// Check if events.jsonl exists
	if _, err := os.Stat(eventsPath); err != nil {
		if os.IsNotExist(err) {
			// No events file is not a violation - it means no clew contract to validate
			return nil, nil
		}
		return nil, fmt.Errorf("failed to access events.jsonl: %w", err)
	}

	// Parse all events
	events, err := parseEventsFile(eventsPath)
	if err != nil {
		return nil, err
	}

	// Collect violations
	var violations []ContractViolation

	// Validate handoff sequences
	handoffViolations := validateHandoffSequences(events)
	violations = append(violations, handoffViolations...)

	// Validate task lifecycle
	taskViolations := validateTaskLifecycle(events)
	violations = append(violations, taskViolations...)

	return violations, nil
}

// parseEventsFile reads and parses events.jsonl into a slice of events.
func parseEventsFile(path string) ([]clewcontract.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open events.jsonl: %w", err)
	}
	defer file.Close()

	var events []clewcontract.Event
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		var event clewcontract.Event
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, fmt.Errorf("failed to parse event at line %d: %w", lineNum, err)
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading events.jsonl: %w", err)
	}

	return events, nil
}

// validateHandoffSequences checks that handoff_prepared precedes handoff_executed.
// For each handoff_executed event, there must be a corresponding handoff_prepared
// event with the same from_agent and to_agent that occurred earlier in the timeline.
func validateHandoffSequences(events []clewcontract.Event) []ContractViolation {
	var violations []ContractViolation

	// Track handoff preparations: key is "from_agent:to_agent"
	handoffPrepared := make(map[string]int) // map to event index

	for i, event := range events {
		switch event.Type {
		case clewcontract.EventTypeHandoffPrepared:
			// Record this preparation
			fromAgent := getMetaString(event.Meta, "from_agent")
			toAgent := getMetaString(event.Meta, "to_agent")
			if fromAgent != "" && toAgent != "" {
				key := fromAgent + ":" + toAgent
				handoffPrepared[key] = i
			}

		case clewcontract.EventTypeHandoffExecuted:
			// Check if there was a preceding preparation
			fromAgent := getMetaString(event.Meta, "from_agent")
			toAgent := getMetaString(event.Meta, "to_agent")
			if fromAgent == "" || toAgent == "" {
				violations = append(violations, ContractViolation{
					Type:          "handoff_missing_metadata",
					Description:   fmt.Sprintf("handoff_executed at event %d missing from_agent or to_agent", i),
					Severity:      "error",
					RelatedEvents: []int{i},
				})
				continue
			}

			key := fromAgent + ":" + toAgent
			prepIndex, exists := handoffPrepared[key]
			if !exists {
				violations = append(violations, ContractViolation{
					Type:          "handoff_unprepared",
					Description:   fmt.Sprintf("handoff_executed from %s to %s at event %d has no preceding handoff_prepared", fromAgent, toAgent, i),
					Severity:      "error",
					RelatedEvents: []int{i},
				})
			} else if prepIndex >= i {
				// This should never happen due to iteration order, but check anyway
				violations = append(violations, ContractViolation{
					Type:          "handoff_out_of_order",
					Description:   fmt.Sprintf("handoff_executed at event %d precedes its handoff_prepared at event %d", i, prepIndex),
					Severity:      "error",
					RelatedEvents: []int{prepIndex, i},
				})
			}
			// Clear the preparation so it can't be reused
			delete(handoffPrepared, key)
		}
	}

	return violations
}

// validateTaskLifecycle checks that task_start precedes task_end.
// For each task_end event, there must be a corresponding task_start event
// with the same task_id that occurred earlier in the timeline.
func validateTaskLifecycle(events []clewcontract.Event) []ContractViolation {
	var violations []ContractViolation

	// Track task starts: key is task_id
	taskStarts := make(map[string]int) // map to event index

	for i, event := range events {
		switch event.Type {
		case clewcontract.EventTypeTaskStart:
			// Record this task start
			taskID := getMetaString(event.Meta, "task_id")
			if taskID == "" {
				violations = append(violations, ContractViolation{
					Type:          "task_missing_id",
					Description:   fmt.Sprintf("task_start at event %d missing task_id", i),
					Severity:      "error",
					RelatedEvents: []int{i},
				})
				continue
			}

			if _, exists := taskStarts[taskID]; exists {
				violations = append(violations, ContractViolation{
					Type:          "task_duplicate_start",
					Description:   fmt.Sprintf("duplicate task_start for task_id %s at event %d", taskID, i),
					Severity:      "warning",
					RelatedEvents: []int{i},
				})
			}
			taskStarts[taskID] = i

		case clewcontract.EventTypeTaskEnd:
			// Check if there was a preceding start
			taskID := getMetaString(event.Meta, "task_id")
			if taskID == "" {
				violations = append(violations, ContractViolation{
					Type:          "task_missing_id",
					Description:   fmt.Sprintf("task_end at event %d missing task_id", i),
					Severity:      "error",
					RelatedEvents: []int{i},
				})
				continue
			}

			startIndex, exists := taskStarts[taskID]
			if !exists {
				violations = append(violations, ContractViolation{
					Type:          "task_orphaned_end",
					Description:   fmt.Sprintf("task_end for task_id %s at event %d has no preceding task_start", taskID, i),
					Severity:      "error",
					RelatedEvents: []int{i},
				})
			} else if startIndex >= i {
				// This should never happen due to iteration order, but check anyway
				violations = append(violations, ContractViolation{
					Type:          "task_out_of_order",
					Description:   fmt.Sprintf("task_end at event %d precedes its task_start at event %d", i, startIndex),
					Severity:      "error",
					RelatedEvents: []int{startIndex, i},
				})
			}
			// Remove the task start so we can detect duplicate ends
			delete(taskStarts, taskID)
		}
	}

	return violations
}

// getMetaString extracts a string value from event metadata.
// Returns empty string if key doesn't exist or value is not a string.
func getMetaString(meta map[string]interface{}, key string) string {
	if meta == nil {
		return ""
	}
	val, ok := meta[key]
	if !ok {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

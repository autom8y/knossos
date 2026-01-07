// Package clewcontract provides Clew Contract v2 event recording for Claude Code hooks.
// "Theseus has amnesia; the Clew remembers" - events.jsonl provides the factual route through decisions.
package clewcontract

import (
	"fmt"
	"time"
)

// EventType represents the type of clew event.
type EventType string

// Clew event types for tracking Claude Code activity.
const (
	EventTypeToolCall        EventType = "tool_call"
	EventTypeFileChange      EventType = "file_change"
	EventTypeCommand         EventType = "command"
	EventTypeDecision        EventType = "decision"
	EventTypeContextSwitch   EventType = "context_switch"
	EventTypeSailsGenerated  EventType = "sails_generated"
	EventTypeTaskStart       EventType = "task_start"
	EventTypeTaskEnd         EventType = "task_end"
	EventTypeSessionStart    EventType = "session_start"
	EventTypeSessionEnd      EventType = "session_end"
	EventTypeArtifactCreated EventType = "artifact_created"
	EventTypeError           EventType = "error"
	EventTypeHandoffPrepared EventType = "handoff_prepared"
	EventTypeHandoffExecuted EventType = "handoff_executed"
)

// ArtifactType represents the type of artifact created during a session.
type ArtifactType string

// Artifact types for tracking deliverables produced during Claude Code sessions.
const (
	ArtifactTypePRD        ArtifactType = "prd"
	ArtifactTypeTDD        ArtifactType = "tdd"
	ArtifactTypeADR        ArtifactType = "adr"
	ArtifactTypeTestPlan   ArtifactType = "test_plan"
	ArtifactTypeCode       ArtifactType = "code"
	ArtifactTypeWhiteSails ArtifactType = "white_sails"
)

// Event represents a clew event in the events.jsonl log.
// Each event captures a discrete action during a Claude Code session.
type Event struct {
	// Timestamp in RFC3339 format with milliseconds
	Timestamp string `json:"ts"`

	// Type of event (tool_call, file_change, command, decision, context_switch)
	Type EventType `json:"type"`

	// Tool name when Type is tool_call (Edit, Write, Bash, Read, Glob, Grep, Task)
	Tool string `json:"tool,omitempty"`

	// Absolute path if applicable
	Path string `json:"path,omitempty"`

	// One-line summary of the event
	Summary string `json:"summary"`

	// Additional metadata (lines_changed, exit_code, duration_ms, etc.)
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// timestamp returns the current time in RFC3339 format with milliseconds.
func timestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

// NewToolCallEvent creates an event for a Claude Code tool invocation.
func NewToolCallEvent(tool, path string, meta map[string]interface{}) Event {
	summary := "Tool: " + tool
	if path != "" {
		summary += " on " + path
	}
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeToolCall,
		Tool:      tool,
		Path:      path,
		Summary:   summary,
		Meta:      meta,
	}
}

// NewFileChangeEvent creates an event for a file modification.
func NewFileChangeEvent(path string, linesChanged int) Event {
	meta := map[string]interface{}{
		"lines_changed": linesChanged,
	}
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeFileChange,
		Path:      path,
		Summary:   "Changed " + path,
		Meta:      meta,
	}
}

// NewCommandEvent creates an event for a shell command execution.
func NewCommandEvent(command string, exitCode int, durationMs int64) Event {
	// Truncate long commands for summary
	summary := command
	if len(summary) > 80 {
		summary = summary[:77] + "..."
	}
	meta := map[string]interface{}{
		"exit_code":   exitCode,
		"duration_ms": durationMs,
	}
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeCommand,
		Summary:   summary,
		Meta:      meta,
	}
}

// NewDecisionEvent creates an event for a workflow decision.
func NewDecisionEvent(summary string, meta map[string]interface{}) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeDecision,
		Summary:   summary,
		Meta:      meta,
	}
}

// NewContextSwitchEvent creates an event for a context change (e.g., new file, new task).
func NewContextSwitchEvent(summary string, path string, meta map[string]interface{}) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeContextSwitch,
		Path:      path,
		Summary:   summary,
		Meta:      meta,
	}
}

// EvidencePaths contains paths to proof evidence files from WHITE_SAILS.yaml.
// These paths link the sails_generated event to the actual proof artifacts.
type EvidencePaths struct {
	// Tests is the path to test output evidence (e.g., test-output.log).
	Tests string `json:"tests,omitempty"`

	// Build is the path to build output evidence (e.g., build-output.log).
	Build string `json:"build,omitempty"`

	// Lint is the path to lint output evidence (e.g., lint-output.log).
	Lint string `json:"lint,omitempty"`

	// Adversarial is the path to adversarial test evidence (optional).
	Adversarial string `json:"adversarial,omitempty"`

	// Integration is the path to integration test evidence (optional).
	Integration string `json:"integration,omitempty"`
}

// SailsGeneratedData contains the data for a SAILS_GENERATED event.
// This captures the White Sails confidence signal generation per TDD 5.4.
type SailsGeneratedData struct {
	// Color is the final confidence signal (WHITE, GRAY, BLACK).
	Color string

	// ComputedBase is the computed color before human modifiers.
	ComputedBase string

	// Reasons explains why this color was computed.
	Reasons []string

	// FilePath is the path to the generated WHITE_SAILS.yaml.
	FilePath string

	// EvidencePaths contains paths to the proof evidence files.
	// These reflect the WHITE_SAILS.yaml proof file paths.
	EvidencePaths *EvidencePaths
}

// NewSailsGeneratedEvent creates an event for White Sails generation.
// This event is emitted when a session's confidence signal (WHITE_SAILS.yaml) is generated.
//
// Per TDD Section 5.4, the event captures:
//   - Session ID
//   - Final color (WHITE/GRAY/BLACK)
//   - Computed base color (before modifiers)
//   - Reasons array explaining the color determination
//   - File path to WHITE_SAILS.yaml
//   - Evidence paths from WHITE_SAILS.yaml proofs (tests, build, lint)
func NewSailsGeneratedEvent(sessionID string, data SailsGeneratedData) Event {
	meta := map[string]interface{}{
		"session_id":    sessionID,
		"color":         data.Color,
		"computed_base": data.ComputedBase,
		"reasons":       data.Reasons,
		"file_path":     data.FilePath,
	}

	// Add evidence_paths to meta if provided
	if data.EvidencePaths != nil {
		evidencePaths := make(map[string]string)
		if data.EvidencePaths.Tests != "" {
			evidencePaths["tests"] = data.EvidencePaths.Tests
		}
		if data.EvidencePaths.Build != "" {
			evidencePaths["build"] = data.EvidencePaths.Build
		}
		if data.EvidencePaths.Lint != "" {
			evidencePaths["lint"] = data.EvidencePaths.Lint
		}
		if data.EvidencePaths.Adversarial != "" {
			evidencePaths["adversarial"] = data.EvidencePaths.Adversarial
		}
		if data.EvidencePaths.Integration != "" {
			evidencePaths["integration"] = data.EvidencePaths.Integration
		}
		if len(evidencePaths) > 0 {
			meta["evidence_paths"] = evidencePaths
		}
	}

	summary := "Generated WHITE_SAILS: " + data.Color
	if data.Color != data.ComputedBase {
		summary += " (base: " + data.ComputedBase + ")"
	}

	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeSailsGenerated,
		Path:      data.FilePath,
		Summary:   summary,
		Meta:      meta,
	}
}

// NewTaskStartEvent creates a task_start event for cognitive budget tracking.
// This marks the beginning of a task within a workflow phase.
//
// Parameters:
//   - taskID: Unique task identifier (e.g., "task-001")
//   - agent: The agent executing the task
//   - phase: The workflow phase (e.g., "design", "implementation", "validation")
//   - sessionID: The session ID context for the task
func NewTaskStartEvent(taskID, agent, phase, sessionID string) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeTaskStart,
		Summary:   fmt.Sprintf("Task started: %s by %s in %s phase", taskID, agent, phase),
		Meta: map[string]interface{}{
			"task_id":    taskID,
			"agent":      agent,
			"phase":      phase,
			"session_id": sessionID,
		},
	}
}

// NewTaskEndEvent creates a task_end event with outcome for cognitive budget tracking.
// This marks the end of a task and captures completion metrics.
//
// Parameters:
//   - taskID: Unique task identifier matching the task_start event
//   - agent: The agent that executed the task
//   - outcome: Task completion outcome (e.g., "success", "failed", "blocked")
//   - sessionID: The session ID context for the task
//   - durationMs: Task execution duration in milliseconds
//   - artifacts: List of artifact paths produced by the task
func NewTaskEndEvent(taskID, agent, outcome, sessionID string, durationMs int64, artifacts []string) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeTaskEnd,
		Summary:   fmt.Sprintf("Task ended: %s by %s - %s (%dms)", taskID, agent, outcome, durationMs),
		Meta: map[string]interface{}{
			"task_id":     taskID,
			"agent":       agent,
			"outcome":     outcome,
			"session_id":  sessionID,
			"duration_ms": durationMs,
			"artifacts":   artifacts,
		},
	}
}

// NewSessionStartEvent creates an event for session initialization.
// This marks the beginning of a tracked Claude Code session.
//
// Parameters:
//   - sessionID: The unique session identifier
//   - initiative: The initiative or goal for this session
//   - complexity: Complexity rating (trivial, standard, complex, critical)
//   - team: The active rite (e.g., "10x-dev")
func NewSessionStartEvent(sessionID, initiative, complexity, team string) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeSessionStart,
		Summary:   "Session started: " + initiative + " (" + complexity + ")",
		Meta: map[string]interface{}{
			"session_id": sessionID,
			"initiative": initiative,
			"complexity": complexity,
			"team":       team,
		},
	}
}

// NewSessionEndEvent creates an event for session completion.
// This marks the end of a tracked Claude Code session.
//
// Parameters:
//   - sessionID: The unique session identifier
//   - status: Completion status (e.g., "completed", "parked", "abandoned")
//   - durationMs: Total session duration in milliseconds
func NewSessionEndEvent(sessionID, status string, durationMs int64) Event {
	return NewSessionEndEventWithBudget(sessionID, status, durationMs, nil)
}

// NewSessionEndEventWithBudget creates a session_end event with optional cognitive budget metadata.
// Parameters:
//   - sessionID: Session identifier
//   - status: Completion status (e.g., "completed", "parked", "abandoned")
//   - durationMs: Total session duration in milliseconds
//   - budget: Optional cognitive budget data (tool calls, message count, etc.)
func NewSessionEndEventWithBudget(sessionID, status string, durationMs int64, budget map[string]interface{}) Event {
	meta := map[string]interface{}{
		"session_id":  sessionID,
		"status":      status,
		"duration_ms": durationMs,
	}

	// Add budget metadata if provided
	if budget != nil && len(budget) > 0 {
		meta["cognitive_budget"] = budget
	}

	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeSessionEnd,
		Summary:   "Session ended: " + status,
		Meta:      meta,
	}
}

// NewArtifactCreatedEvent creates an event for artifact creation.
// This captures semantic artifact creation for handoff validation, distinct from file_change
// which only tracks raw file modifications.
//
// Parameters:
//   - artifactType: The type of artifact (prd, tdd, adr, test_plan, code)
//   - path: Absolute path to the created artifact
//   - phase: The workflow phase during which the artifact was created
//   - validatesAgainst: Reference to the artifact this validates against (e.g., PRD path for TDD)
func NewArtifactCreatedEvent(artifactType ArtifactType, path, phase string, validatesAgainst string) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeArtifactCreated,
		Path:      path,
		Summary:   fmt.Sprintf("Artifact created: %s (%s)", artifactType, phase),
		Meta: map[string]interface{}{
			"artifact_type":     string(artifactType),
			"phase":             phase,
			"validates_against": validatesAgainst,
		},
	}
}

// NewErrorEvent creates an event for structured error capture.
// This provides dedicated error tracking with recovery guidance for handoff protocols.
//
// Parameters:
//   - errorCode: A structured error code (e.g., "VALIDATION_FAILED", "DEPENDENCY_MISSING")
//   - message: Human-readable error message
//   - context: Additional context about where/why the error occurred
//   - recoverable: Whether the error is recoverable without human intervention
//   - suggestedAction: Recommended action to resolve the error
func NewErrorEvent(errorCode, message, context string, recoverable bool, suggestedAction string) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeError,
		Summary:   fmt.Sprintf("Error: %s - %s", errorCode, message),
		Meta: map[string]interface{}{
			"error_code":       errorCode,
			"message":          message,
			"context":          context,
			"recoverable":      recoverable,
			"suggested_action": suggestedAction,
		},
	}
}

// NewHandoffPreparedEvent creates an event for handoff preparation.
// This marks the preparation phase of agent handoff per Knossos doctrine section VI.
//
// Parameters:
//   - fromAgent: The agent transferring work
//   - toAgent: The agent receiving work
//   - sessionID: The session ID context for the handoff
func NewHandoffPreparedEvent(fromAgent, toAgent, sessionID string) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeHandoffPrepared,
		Summary:   fmt.Sprintf("Handoff prepared: %s -> %s", fromAgent, toAgent),
		Meta: map[string]interface{}{
			"from_agent": fromAgent,
			"to_agent":   toAgent,
			"session_id": sessionID,
		},
	}
}

// NewHandoffExecutedEvent creates an event for handoff execution.
// This marks the execution phase of agent handoff per Knossos doctrine section VI.
//
// Parameters:
//   - fromAgent: The agent that transferred work
//   - toAgent: The agent that received work
//   - sessionID: The session ID context for the handoff
//   - artifacts: List of artifact paths transferred during handoff
func NewHandoffExecutedEvent(fromAgent, toAgent, sessionID string, artifacts []string) Event {
	return Event{
		Timestamp: timestamp(),
		Type:      EventTypeHandoffExecuted,
		Summary:   fmt.Sprintf("Handoff executed: %s -> %s (%d artifacts)", fromAgent, toAgent, len(artifacts)),
		Meta: map[string]interface{}{
			"from_agent": fromAgent,
			"to_agent":   toAgent,
			"session_id": sessionID,
			"artifacts":  artifacts,
		},
	}
}

// Stamp represents a decision stamp - "Why we chose A not B" at a fork.
// Stamps provide structured rationale for significant decisions during a session.
type Stamp struct {
	// Timestamp in RFC3339 format with milliseconds
	Ts time.Time `json:"ts"`

	// Decision describes what choice was made
	Decision string `json:"decision"`

	// Rationale explains why (1-3 lines)
	Rationale string `json:"rationale"`

	// Rejected lists alternatives that were NOT chosen (optional)
	Rejected []string `json:"rejected,omitempty"`

	// Context provides additional metadata
	Context map[string]any `json:"context,omitempty"`
}

// NewStamp creates a new Stamp with the current timestamp.
func NewStamp(decision, rationale string, rejected []string, context map[string]any) Stamp {
	return Stamp{
		Ts:        time.Now().UTC(),
		Decision:  decision,
		Rationale: rationale,
		Rejected:  rejected,
		Context:   context,
	}
}

// ToEvent converts a Stamp to an Event with type="decision".
// This allows stamps to be written to events.jsonl in a consistent format.
func (s Stamp) ToEvent() Event {
	meta := make(map[string]interface{})

	// Add rationale to meta
	meta["rationale"] = s.Rationale

	// Add rejected alternatives if present
	if len(s.Rejected) > 0 {
		meta["rejected"] = s.Rejected
	}

	// Merge stamp context into meta
	for k, v := range s.Context {
		meta[k] = v
	}

	return Event{
		Timestamp: s.Ts.Format("2006-01-02T15:04:05.000Z"),
		Type:      EventTypeDecision,
		Summary:   s.Decision,
		Meta:      meta,
	}
}

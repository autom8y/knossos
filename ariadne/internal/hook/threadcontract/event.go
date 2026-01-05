// Package threadcontract provides Thread Contract v2 event recording for Claude Code hooks.
// "Theseus has amnesia; the Thread remembers" - events.jsonl provides the factual route through decisions.
package threadcontract

import "time"

// EventType represents the type of thread event.
type EventType string

// Thread event types for tracking Claude Code activity.
const (
	EventTypeToolCall       EventType = "tool_call"
	EventTypeFileChange     EventType = "file_change"
	EventTypeCommand        EventType = "command"
	EventTypeDecision       EventType = "decision"
	EventTypeContextSwitch  EventType = "context_switch"
	EventTypeSailsGenerated EventType = "sails_generated"
)

// Event represents a thread event in the events.jsonl log.
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
func NewSailsGeneratedEvent(sessionID string, data SailsGeneratedData) Event {
	meta := map[string]interface{}{
		"session_id":    sessionID,
		"color":         data.Color,
		"computed_base": data.ComputedBase,
		"reasons":       data.Reasons,
		"file_path":     data.FilePath,
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

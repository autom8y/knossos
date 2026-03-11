package clewcontract

import (
	"encoding/json"
	"time"
)

// EventSource distinguishes who produced a v3 typed event.
// This allows consumers to filter by origin without parsing the type string.
type EventSource string

const (
	// SourceCLI indicates the event was produced by an `ari` CLI subcommand.
	SourceCLI EventSource = "cli"

	// SourceHook indicates the event was produced by a CC lifecycle hook
	// (PostToolUse, SessionStart, SessionEnd, SubagentStart, SubagentStop).
	SourceHook EventSource = "hook"

	// SourceAgent indicates the event was produced by an agent calling `ari session log`.
	SourceAgent EventSource = "agent"
)

// TypedEvent is the v3 event envelope with structured, per-type data.
//
// It coexists with the v2 flat Event struct in the same events.jsonl file.
// Format detection: a JSON line has a "data" field -> TypedEvent; otherwise -> Event (v2).
//
// Serialization: one JSON object per line (JSONL), newline-terminated.
// The Data field is always present and always a JSON object (never null, never array).
type TypedEvent struct {
	// Ts is the event timestamp in RFC3339 with milliseconds, always UTC.
	// Format: "2006-01-02T15:04:05.000Z"
	Ts string `json:"ts"`

	// Type is the event type in dotted namespace format.
	// Pattern: ^[a-z][a-z0-9]*(\.[a-z][a-z0-9_]*)+$
	Type EventType `json:"type"`

	// Source identifies who produced this event.
	// One of: cli, hook, agent.
	Source EventSource `json:"source"`

	// Channel identifies which AI assistant triggered the event.
	// Typically "claude" or "gemini", defaults to "claude" if omitted.
	Channel string `json:"channel,omitempty"`

	// Data is the per-type structured payload. Always a valid JSON object.
	// Use the concrete typed constructors to produce correctly-typed Data values.
	Data json.RawMessage `json:"data"`
}

// typedEventTimestamp returns the current time in TypedEvent format.
// Same format as Event.Timestamp for consistency.
func typedEventTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

// newTypedEvent constructs a TypedEvent by marshaling the data payload.
// data must be a value that serializes as a JSON object (struct with json tags).
// If marshaling fails, Data is set to the empty object `{}` to preserve invariant.
// Channel is only set when non-empty and not "claude" (the default).
func newTypedEvent(eventType EventType, source EventSource, channel string, data any) TypedEvent {
	raw, err := json.Marshal(data)
	if err != nil {
		// Invariant: Data is always a valid JSON object. Fall back to empty object.
		raw = json.RawMessage(`{}`)
	}
	te := TypedEvent{
		Ts:     typedEventTimestamp(),
		Type:   eventType,
		Source: source,
		Data:   raw,
	}
	if channel != "" && channel != "claude" {
		te.Channel = channel
	}
	return te
}

// New v3 EventType constants for renamed and new types.
// All existing v2 constants in event.go are retained unchanged.
const (
	// v3 renames of v2 types -- new preferred constants.
	// Deprecated v2 constants are marked in event.go.

	// EventTypeToolInvoked is the v3 rename of EventTypeToolCall ("tool.call").
	EventTypeToolInvoked EventType = "tool.invoked"

	// EventTypeFileModified is the v3 rename of EventTypeFileChange ("tool.file_change").
	EventTypeFileModified EventType = "file.modified"

	// EventTypeDecisionRecorded is the v3 rename of EventTypeDecision ("agent.decision").
	EventTypeDecisionRecorded EventType = "decision.recorded"

	// EventTypeAgentDelegated is the v3 rename of EventTypeTaskStart ("agent.task_start").
	EventTypeAgentDelegated EventType = "agent.delegated"

	// EventTypeAgentCompleted is the v3 rename of EventTypeTaskEnd ("agent.task_end").
	EventTypeAgentCompleted EventType = "agent.completed"

	// EventTypeArtifactCreatedV3 is the v3 rename of EventTypeArtifactCreated ("tool.artifact_created").
	// The V3 suffix resolves the naming collision with the existing EventTypeArtifactCreated constant.
	EventTypeArtifactCreatedV3 EventType = "artifact.created"

	// EventTypeErrorOccurred is the v3 rename of EventTypeError ("tool.error").
	EventTypeErrorOccurred EventType = "error.occurred"

	// Entirely new v3 types with no v2 equivalent.

	// EventTypeSessionWrapped is emitted by `ari session wrap` as the curated "session intentionally concluded" signal.
	// Distinct from EventTypeSessionEnd which is a backplane lifecycle event.
	EventTypeSessionWrapped EventType = "session.wrapped"

	// EventTypeCommitCreated is emitted when a git commit is detected in a PostToolUse hook.
	// No v2 equivalent; git commits were previously captured as generic tool.call events.
	EventTypeCommitCreated EventType = "commit.created"

	// EventTypeCommandInvoked is emitted when a Skill tool call is detected in a PostToolUse hook.
	// No v2 equivalent; Skill tool invocations were not previously tracked.
	EventTypeCommandInvoked EventType = "command.invoked"

	// EventTypeFieldUpdated is emitted by the future `ari session field-set` command.
	// No current producer; declared here per spec for future implementation.
	EventTypeFieldUpdated EventType = "field.updated"

	// EventTypeHookFired is emitted by the hook runner for observability.
	// No current producer; declared here per spec for future implementation.
	EventTypeHookFired EventType = "hook.fired"
)

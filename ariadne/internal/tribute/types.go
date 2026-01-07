// Package tribute implements TRIBUTE.md auto-generation for session summaries.
// In the Knossos mythology, King Minos demanded tribute from Athens; in our system,
// TRIBUTE.md is the "payment" for navigating the labyrinth--a comprehensive record
// of what was accomplished, what decisions were made, and what artifacts were produced.
package tribute

import (
	"time"
)

// SchemaVersion is the current TRIBUTE.md schema version.
const SchemaVersion = "1.0"

// GenerateResult contains the output of tribute generation.
type GenerateResult struct {
	// FilePath is the path to the generated TRIBUTE.md.
	FilePath string

	// SessionID is the session identifier.
	SessionID string

	// Initiative is the session initiative name.
	Initiative string

	// Complexity is the complexity tier (SCRIPT/MODULE/SERVICE/SYSTEM).
	Complexity string

	// Duration is the session duration.
	Duration time.Duration

	// Rite is the active rite for the session.
	Rite string

	// FinalPhase is the phase at wrap time.
	FinalPhase string

	// StartedAt is when the session started.
	StartedAt time.Time

	// EndedAt is when the session ended (archived).
	EndedAt time.Time

	// Artifacts contains produced artifacts.
	Artifacts []Artifact

	// Decisions contains recorded decisions.
	Decisions []Decision

	// Phases contains phase progression records.
	Phases []PhaseRecord

	// Handoffs contains agent handoff records.
	Handoffs []Handoff

	// Commits contains git commits (Phase 2 - placeholder for now).
	Commits []Commit

	// SailsColor is the confidence signal color.
	SailsColor string

	// SailsBase is the computed base color before modifiers.
	SailsBase string

	// SailsProofs contains the proof status map.
	SailsProofs map[string]ProofStatus

	// Metrics contains session metrics.
	Metrics Metrics

	// Notes contains any additional notes from SESSION_CONTEXT body.
	Notes string

	// GeneratedAt is when the tribute was generated.
	GeneratedAt time.Time
}

// Artifact represents a produced artifact.
type Artifact struct {
	// Type is the artifact type (PRD, TDD, ADR, Code, Tests, etc.)
	Type string

	// Path is the artifact file path.
	Path string

	// Status is the artifact status (Created, Approved, Implemented, Passing).
	Status string

	// Timestamp is when the artifact was created.
	Timestamp time.Time
}

// Decision represents a recorded decision.
type Decision struct {
	// Timestamp is when the decision was made.
	Timestamp time.Time

	// Decision is the decision text.
	Decision string

	// Rationale is the reasoning behind the decision.
	Rationale string

	// Rejected contains rejected alternatives.
	Rejected []string

	// Context provides additional context.
	Context string
}

// PhaseRecord represents a workflow phase.
type PhaseRecord struct {
	// Phase is the phase name.
	Phase string

	// StartedAt is when the phase started.
	StartedAt time.Time

	// Duration is how long the phase lasted.
	Duration time.Duration

	// Agent is the responsible agent.
	Agent string
}

// Handoff represents an agent handoff.
type Handoff struct {
	// From is the source agent.
	From string

	// To is the target agent.
	To string

	// Timestamp is when the handoff occurred.
	Timestamp time.Time

	// Notes contains handoff notes.
	Notes string
}

// Commit represents a git commit (Phase 2 - placeholder).
type Commit struct {
	// Hash is the full commit hash.
	Hash string

	// ShortHash is the abbreviated hash.
	ShortHash string

	// Message is the commit message.
	Message string

	// FilesChanged is the number of files changed.
	FilesChanged int

	// Timestamp is when the commit was made.
	Timestamp time.Time
}

// Metrics contains session metrics.
type Metrics struct {
	// ToolCalls is the number of tool calls.
	ToolCalls int

	// EventsRecorded is the number of events in events.jsonl.
	EventsRecorded int

	// FilesModified is the number of files modified.
	FilesModified int

	// LinesAdded is the number of lines added.
	LinesAdded int

	// LinesRemoved is the number of lines removed.
	LinesRemoved int
}

// ProofStatus represents a proof item status for sails.
type ProofStatus struct {
	// Status is the proof status (PASS, FAIL, SKIP, UNKNOWN).
	Status string

	// Summary provides additional context.
	Summary string
}

// TributeFrontmatter represents the YAML frontmatter for TRIBUTE.md.
type TributeFrontmatter struct {
	SchemaVersion string  `yaml:"schema_version"`
	SessionID     string  `yaml:"session_id"`
	Initiative    string  `yaml:"initiative"`
	Complexity    string  `yaml:"complexity"`
	GeneratedAt   string  `yaml:"generated_at"`
	DurationHours float64 `yaml:"duration_hours,omitempty"`
}

// WhiteSailsData represents parsed WHITE_SAILS.yaml data.
type WhiteSailsData struct {
	Color        string
	ComputedBase string
	Proofs       map[string]ProofStatus
}

// EventData represents a generic event from events.jsonl.
// Events can have different schemas, so we use interface{} for flexible parsing.
type EventData struct {
	// Common fields across event types
	Timestamp string                 `json:"timestamp,omitempty"` // Standard events
	Ts        string                 `json:"ts,omitempty"`        // Hook-generated events
	Event     string                 `json:"event,omitempty"`     // Standard event type
	Type      string                 `json:"type,omitempty"`      // Hook event type
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	FromPhase string                 `json:"from_phase,omitempty"`
	ToPhase   string                 `json:"to_phase,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Summary   string                 `json:"summary,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"` // Hook events use "meta"

	// Decision-specific fields
	Decision  string   `json:"decision,omitempty"`
	Rationale string   `json:"rationale,omitempty"`
	Rejected  []string `json:"rejected,omitempty"`
	Context   string   `json:"context,omitempty"`

	// Artifact-specific fields
	ArtifactType string `json:"artifact_type,omitempty"`

	// Handoff-specific fields
	Notes string `json:"notes,omitempty"`
}

// GetTimestamp returns the timestamp from either format.
func (e *EventData) GetTimestamp() string {
	if e.Timestamp != "" {
		return e.Timestamp
	}
	return e.Ts
}

// GetEventType returns the event type from either format.
func (e *EventData) GetEventType() string {
	if e.Event != "" {
		return e.Event
	}
	return e.Type
}

// GetMetadata returns metadata from either format.
func (e *EventData) GetMetadata() map[string]interface{} {
	if e.Metadata != nil {
		return e.Metadata
	}
	return e.Meta
}

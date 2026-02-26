package clewcontract

// typed_data.go -- Per-type data structs for all v3 TypedEvent payloads.
//
// Each struct here is the Go representation of the "data" field in a TypedEvent.
// These structs are serialized to json.RawMessage inside TypedEvent.Data.
//
// Naming convention: {EventNameInCamelCase}Data
// All fields use json tags matching the SESSION-1 spec exactly.

// --- Curated types (projected to SESSION_CONTEXT.md timeline) ---

// SessionCreatedData is the data payload for the "session.created" event.
// Emitted by `ari session create`.
type SessionCreatedData struct {
	SessionID  string `json:"session_id"`          // required
	Initiative string `json:"initiative"`           // required
	Complexity string `json:"complexity"`           // required: PATCH|MODULE|INITIATIVE
	Rite       string `json:"rite,omitempty"`        // optional, empty for cross-cutting
}

// SessionParkedData is the data payload for the "session.parked" event.
// Emitted by `ari session park`.
type SessionParkedData struct {
	SessionID string `json:"session_id"` // required
	Reason    string `json:"reason"`     // required
}

// SessionResumedData is the data payload for the "session.resumed" event.
// Emitted by `ari session resume`.
type SessionResumedData struct {
	SessionID string `json:"session_id"` // required
}

// SessionWrappedData is the data payload for the "session.wrapped" event.
// Emitted by `ari session wrap`. This is the curated "session intentionally concluded" signal.
// Distinct from SessionEndedData which is a backplane lifecycle event.
type SessionWrappedData struct {
	SessionID  string `json:"session_id"`               // required
	SailsColor string `json:"sails_color,omitempty"`    // optional: WHITE, GRAY, BLACK
	DurationMs int64  `json:"duration_ms"`              // required
}

// SessionFrayedData is the data payload for the "session.frayed" event.
// Emitted by `ari session fray`.
type SessionFrayedData struct {
	ParentID  string `json:"parent_id"`  // required
	ChildID   string `json:"child_id"`   // required
	FrayPoint string `json:"fray_point"` // required, phase at fork
}

// PhaseTransitionedData is the data payload for the "phase.transitioned" event.
// Emitted by `ari session transition`.
type PhaseTransitionedData struct {
	SessionID string `json:"session_id"` // required
	From      string `json:"from"`       // required, phase name
	To        string `json:"to"`         // required, phase name
}

// AgentDelegatedData is the data payload for the "agent.delegated" event.
// Emitted by SubagentStart hook and `ari handoff execute`.
// v3 rename of agent.task_start; v2 constant EventTypeTaskStart retained for backward compat.
type AgentDelegatedData struct {
	AgentName string `json:"agent_name"`           // required
	AgentType string `json:"agent_type,omitempty"` // optional: specialist, orchestrator
	TaskID    string `json:"task_id,omitempty"`    // optional
	AgentID   string `json:"agent_id,omitempty"`   // optional, CC-assigned subagent ID
}

// AgentCompletedData is the data payload for the "agent.completed" event.
// Emitted by SubagentStop hook and `ari handoff prepare`.
// v3 rename of agent.task_end; v2 constant EventTypeTaskEnd retained for backward compat.
type AgentCompletedData struct {
	AgentName  string   `json:"agent_name"`            // required
	AgentType  string   `json:"agent_type,omitempty"`  // optional
	TaskID     string   `json:"task_id,omitempty"`     // optional
	AgentID    string   `json:"agent_id,omitempty"`    // optional
	Outcome    string   `json:"outcome,omitempty"`     // optional: success, failed, blocked
	DurationMs int64    `json:"duration_ms,omitempty"` // optional
	Artifacts  []string `json:"artifacts,omitempty"`   // optional
}

// CommitCreatedData is the data payload for the "commit.created" event.
// Emitted by the PostToolUse hook when a git commit is detected in a Bash tool call.
// This is a NEW event type with no v2 equivalent.
type CommitCreatedData struct {
	SHA     string `json:"sha"`     // required, short or full SHA
	Message string `json:"message"` // required, first line of commit message
}

// DecisionRecordedData is the data payload for the "decision.recorded" event.
// Emitted by agents via `ari session log --type=decision`, also via Stamp.ToEvent().
// v3 rename of agent.decision; v2 constant EventTypeDecision retained for backward compat.
type DecisionRecordedData struct {
	Decision  string   `json:"decision"`           // required
	Rationale string   `json:"rationale"`          // required
	Rejected  []string `json:"rejected,omitempty"` // optional, alternatives not chosen
}

// CommandInvokedData is the data payload for the "command.invoked" event.
// Emitted by the PostToolUse hook when a Skill tool call is detected.
// This is a NEW event type with no v2 equivalent.
type CommandInvokedData struct {
	Command string `json:"command"` // required, skill/command name
	Type    string `json:"type"`    // required: "skill" or "command"
}

// --- Backplane-only types (events.jsonl only, not curated to timeline) ---

// ToolInvokedData is the data payload for the "tool.invoked" event.
// Emitted by the PostToolUse hook for every tool call.
// v3 rename of tool.call; Meta field retained as escape hatch for tool-specific fields.
type ToolInvokedData struct {
	Tool    string                 `json:"tool"`           // required
	Path    string                 `json:"path,omitempty"` // optional
	Summary string                 `json:"summary"`        // required
	Meta    map[string]any `json:"meta,omitempty"` // optional, tool-specific fields
}

// FileModifiedData is the data payload for the "file.modified" event.
// Emitted by the PostToolUse hook (emitSupplementalEvents).
// v3 rename of tool.file_change.
type FileModifiedData struct {
	Path         string `json:"path"`          // required, absolute path
	LinesChanged int    `json:"lines_changed"` // required
}

// ArtifactCreatedData is the data payload for the "artifact.created" event.
// Emitted by the PostToolUse hook (emitSupplementalEvents).
// v3 rename of tool.artifact_created; promoted from tool.* to first-class namespace.
type ArtifactCreatedData struct {
	ArtifactType     string `json:"artifact_type"`               // required: prd, tdd, adr, test_plan, ephemeral
	Path             string `json:"path"`                        // required, absolute path
	Phase            string `json:"phase"`                       // required, workflow phase
	ValidatesAgainst string `json:"validates_against,omitempty"` // optional, path to upstream artifact
	WipType          string `json:"wip_type,omitempty"`          // optional, for .wip/ artifacts
	Slug             string `json:"slug,omitempty"`              // optional, for .wip/ artifacts
}

// ErrorOccurredData is the data payload for the "error.occurred" event.
// Emitted by the PostToolUse hook (emitErrorEvent) or CLI commands.
// v3 rename of tool.error; promoted from tool.* to first-class namespace.
type ErrorOccurredData struct {
	ErrorCode       string `json:"error_code"`       // required
	Message         string `json:"message"`          // required
	Context         string `json:"context"`          // required
	Recoverable     bool   `json:"recoverable"`      // required
	SuggestedAction string `json:"suggested_action"` // required
}

// SessionStartedData is the data payload for the "session.started" event.
// Emitted by the SessionStart CC lifecycle hook.
// Distinct from SessionCreatedData: "session.started" = CC session began;
// "session.created" = knossos session was created via CLI.
type SessionStartedData struct {
	SessionID  string `json:"session_id"`        // required
	Initiative string `json:"initiative"`         // required
	Complexity string `json:"complexity"`         // required
	Rite       string `json:"rite,omitempty"`     // optional
}

// SessionEndedData is the data payload for the "session.ended" event.
// Emitted by the SessionEnd hook and CLI commands (park, wrap).
// Backplane event; SessionWrappedData and SessionParkedData are the curated equivalents.
type SessionEndedData struct {
	SessionID       string                 `json:"session_id"`                // required
	Status          string                 `json:"status"`                    // required: completed, parked, abandoned
	DurationMs      int64                  `json:"duration_ms"`               // required
	CognitiveBudget map[string]any `json:"cognitive_budget,omitempty"` // optional
}

// SessionArchivedData is the data payload for the "session.archived" event.
// Emitted by `ari session wrap`. Internal lifecycle backplane event.
type SessionArchivedData struct {
	SessionID  string `json:"session_id"` // required
	FromStatus string `json:"from"`       // required: ACTIVE, PARKED
}

// StrandResolvedData is the data payload for the "session.strand_resolved" event.
// Emitted by `ari session wrap` when a frayed child session wraps.
type StrandResolvedData struct {
	ParentID   string `json:"parent_id"`  // required
	ChildID    string `json:"child_id"`   // required
	Resolution string `json:"resolution"` // required: wrapped, abandoned
}

// SchemaMigratedData is the data payload for the "session.schema_migrated" event.
// Emitted by `ari session migrate`.
type SchemaMigratedData struct {
	SessionID   string `json:"session_id"`   // required
	FromVersion string `json:"from_version"` // required
	ToVersion   string `json:"to_version"`   // required
}

// LockAcquiredData is the data payload for the "lock.acquired" event.
// Emitted by session lock operations.
type LockAcquiredData struct {
	SessionID string `json:"session_id"` // required
	Holder    string `json:"holder"`     // required
}

// LockReleasedData is the data payload for the "lock.released" event.
// Emitted by session lock operations (declared but not currently emitted).
type LockReleasedData struct {
	SessionID string `json:"session_id"` // required
	Holder    string `json:"holder"`     // required
}

// SailsGeneratedTypedData is the data payload for the "quality.sails_generated" event (v3 typed form).
// Emitted by `ari session wrap`.
// Named SailsGeneratedTypedData to avoid collision with the existing SailsGeneratedData struct in event.go.
type SailsGeneratedTypedData struct {
	SessionID    string            `json:"session_id"`          // required
	Color        string            `json:"color"`               // required: WHITE, GRAY, BLACK
	ComputedBase string            `json:"computed_base"`       // required
	Reasons      []string          `json:"reasons"`             // required
	FilePath     string            `json:"file_path"`           // required
	Evidence     map[string]string `json:"evidence,omitempty"`  // optional: tests, build, lint, adversarial, integration
}

// ContextSwitchData is the data payload for the "context_switch" event.
// Note: context_switch does not follow the dotted namespace convention.
// This is a legacy type string retained for backward compat with trigger detection.
type ContextSwitchData struct {
	Summary string `json:"summary"`          // required
	Path    string `json:"path,omitempty"`   // optional
}

// HandoffPreparedData is the data payload for the "agent.handoff_prepared" event.
// Emitted by `ari handoff prepare`.
type HandoffPreparedData struct {
	FromAgent string `json:"from_agent"` // required
	ToAgent   string `json:"to_agent"`   // required
	SessionID string `json:"session_id"` // required
}

// HandoffExecutedData is the data payload for the "agent.handoff_executed" event.
// Emitted by `ari handoff execute`.
type HandoffExecutedData struct {
	FromAgent string   `json:"from_agent"` // required
	ToAgent   string   `json:"to_agent"`   // required
	SessionID string   `json:"session_id"` // required
	Artifacts []string `json:"artifacts"`  // required (may be empty array)
}

// --- Future types (no current producers) ---

// FieldUpdatedData is the data payload for the "field.updated" event.
// Will be emitted by the future `ari session field-set` command.
// No current producer; defined here per SESSION-1 spec Section 7.1.
type FieldUpdatedData struct {
	SessionID string `json:"session_id"`  // required
	Key       string `json:"key"`         // required
	OldValue  any    `json:"old_value"`   // required (may be null)
	NewValue  any    `json:"new_value"`   // required
}

// HookFiredData is the data payload for the "hook.fired" event.
// Will be emitted by the hook runner for observability.
// No current producer; defined here per SESSION-1 spec Section 7.2.
type HookFiredData struct {
	HookName  string `json:"hook_name"`  // required
	EventType string `json:"event_type"` // required, CC hook event name (e.g., "PostToolUse")
}

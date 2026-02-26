package clewcontract

// source_infer.go -- Source inference for v2 events that lack the "source" field.
//
// When reading v2 flat events, the source field is absent. InferSource() maps
// the v2 type string to the most likely EventSource based on which component
// produces each event type. See SESSION-1 spec Section 6 for the full table.

// v2TypeToSource maps v2 event type strings to their inferred EventSource.
// Built from SESSION-1 spec Section 6 (EventSource Inference Table).
var v2TypeToSource = map[string]EventSource{
	// Session lifecycle -- CLI producers
	"session.created":         SourceCLI,
	"session.parked":          SourceCLI,
	"session.resumed":         SourceCLI,
	"session.archived":        SourceCLI,
	"session.frayed":          SourceCLI,
	"session.strand_resolved": SourceCLI,
	"session.schema_migrated": SourceCLI,
	"session.wrapped":         SourceCLI, // new v3 type, always CLI

	// Hook-emitted lifecycle
	"session.started": SourceHook,

	// session.ended is ambiguous (both hook and CLI produce it).
	// Default to hook per SESSION-1 spec note in Section 6.
	"session.ended": SourceHook,

	// Phase -- CLI producer
	"phase.transitioned": SourceCLI,

	// Tool events -- hook producers
	"tool.call":            SourceHook,
	"tool.file_change":     SourceHook,
	"tool.artifact_created": SourceHook,
	"tool.error":           SourceHook,

	// v3 renames of tool events -- hook producers
	"tool.invoked":     SourceHook,
	"file.modified":    SourceHook,
	"artifact.created": SourceHook,
	"error.occurred":   SourceHook,

	// Decision -- agent producer
	"agent.decision":     SourceAgent,
	"decision.recorded":  SourceAgent, // v3 rename

	// Agent coordination -- hook producers (SubagentStart/Stop)
	"agent.task_start":  SourceHook,
	"agent.task_end":    SourceHook,
	"agent.delegated":   SourceHook, // v3 rename
	"agent.completed":   SourceHook, // v3 rename

	// Handoff -- CLI producers
	"agent.handoff_prepared": SourceCLI,
	"agent.handoff_executed": SourceCLI,

	// Quality -- CLI producer
	"quality.sails_generated": SourceCLI,

	// Lock -- CLI producer
	"lock.acquired": SourceCLI,
	"lock.released": SourceCLI,

	// Trigger detection -- hook (legacy non-dotted name)
	"context_switch": SourceHook,

	// New v3-only types
	"commit.created":  SourceHook,
	"command.invoked": SourceHook,
	"hook.fired":      SourceHook,
	"field.updated":   SourceCLI,
}

// InferSource returns the EventSource for a given v2 event type string.
// This is used when reading v2 flat events that lack a "source" field.
//
// Returns SourceHook for unknown types as a safe default (hooks are the most
// prolific producers and unknown types are most likely hook-emitted).
func InferSource(eventType EventType) EventSource {
	if src, ok := v2TypeToSource[string(eventType)]; ok {
		return src
	}
	// Unknown type: default to hook (most common producer category)
	return SourceHook
}

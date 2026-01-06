package clewcontract

import (
	"path/filepath"

	"github.com/autom8y/ariadne/internal/hook"
)

// RecordToolEvent records a tool event from a Claude Code hook invocation.
// This is the primary integration point for the `ari hook clew` command.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - env: Parsed hook environment variables
//   - toolInput: Parsed tool input JSON
//
// The function extracts relevant information from the hook context and
// writes an appropriate event to the session's events.jsonl file.
func RecordToolEvent(sessionDir string, env *hook.Env, toolInput *hook.ToolInput) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Build event based on tool type
	event := BuildEventFromToolInput(env, toolInput)

	return writer.Write(event)
}

// BuildEventFromToolInput creates an Event from hook context.
// This is exported for trigger checking in the clew command.
func BuildEventFromToolInput(env *hook.Env, toolInput *hook.ToolInput) Event {
	tool := env.ToolName
	path := toolInput.GetEffectivePath()
	meta := make(map[string]interface{})

	// Add tool-specific metadata
	switch tool {
	case "Bash":
		command := toolInput.Command
		if command == "" {
			command = toolInput.GetString("command")
		}
		if command != "" {
			meta["command"] = truncateString(command, 200)
		}
		desc := toolInput.Description
		if desc == "" {
			desc = toolInput.GetString("description")
		}
		if desc != "" {
			meta["description"] = desc
		}

	case "Edit":
		if toolInput.OldString != "" {
			meta["has_old_string"] = true
		}
		if toolInput.NewString != "" {
			meta["has_new_string"] = true
		}

	case "Write":
		if toolInput.Content != "" {
			meta["content_length"] = len(toolInput.Content)
		}

	case "Read":
		// Read is primarily informational
		if limit := toolInput.GetInt("limit"); limit > 0 {
			meta["limit"] = limit
		}
		if offset := toolInput.GetInt("offset"); offset > 0 {
			meta["offset"] = offset
		}

	case "Glob":
		pattern := toolInput.Pattern
		if pattern == "" {
			pattern = toolInput.GetString("pattern")
		}
		if pattern != "" {
			meta["pattern"] = pattern
		}

	case "Grep":
		pattern := toolInput.Pattern
		if pattern == "" {
			pattern = toolInput.GetString("pattern")
		}
		if pattern != "" {
			meta["pattern"] = pattern
		}
		if query := toolInput.Query; query != "" {
			meta["query"] = query
		}

	case "Task":
		// Task tool delegates to sub-agents
		meta["delegation"] = true
	}

	return NewToolCallEvent(tool, path, meta)
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// RecordFileChange records a file change event.
// Use when tracking file modifications outside of tool calls.
func RecordFileChange(sessionDir, path string, linesChanged int) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewFileChangeEvent(path, linesChanged)
	return writer.Write(event)
}

// RecordCommand records a command execution event.
// Use when tracking shell commands with their results.
func RecordCommand(sessionDir, command string, exitCode int, durationMs int64) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewCommandEvent(command, exitCode, durationMs)
	return writer.Write(event)
}

// RecordDecision records a workflow decision event.
// Use when tracking significant decisions during the session.
func RecordDecision(sessionDir, summary string, meta map[string]interface{}) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewDecisionEvent(summary, meta)
	return writer.Write(event)
}

// RecordContextSwitch records a context change event.
// Use when tracking transitions between files, tasks, or focus areas.
func RecordContextSwitch(sessionDir, summary, path string, meta map[string]interface{}) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewContextSwitchEvent(summary, path, meta)
	return writer.Write(event)
}

// GetEventsPath returns the path to events.jsonl for a session directory.
func GetEventsPath(sessionDir string) string {
	return filepath.Join(sessionDir, EventsFileName)
}

// RecordStamp records a decision stamp to events.jsonl.
// This is the primary integration point for the /stamp skill.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - decision: What choice was made (the decision summary)
//   - rationale: Why this choice was made (1-3 lines)
//   - rejected: Alternatives that were NOT chosen (optional, can be nil)
//
// The stamp is converted to an Event with type="decision" and appended
// to the session's events.jsonl file.
func RecordStamp(sessionDir, decision, rationale string, rejected []string) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	stamp := NewStamp(decision, rationale, rejected, nil)
	event := stamp.ToEvent()
	return writer.Write(event)
}

// RecordStampWithContext records a decision stamp with additional context metadata.
// Use this variant when extra contextual information should be captured alongside the decision.
func RecordStampWithContext(sessionDir, decision, rationale string, rejected []string, context map[string]any) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	stamp := NewStamp(decision, rationale, rejected, context)
	event := stamp.ToEvent()
	return writer.Write(event)
}

// RecordTaskStart records a task start event for cognitive budget tracking.
// Use when a task begins within a workflow phase.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - taskID: Unique task identifier (e.g., "task-001")
//   - agent: The agent executing the task
//   - phase: The workflow phase (e.g., "design", "implementation", "validation")
//   - sessionID: The session ID context for the task
func RecordTaskStart(sessionDir, taskID, agent, phase, sessionID string) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewTaskStartEvent(taskID, agent, phase, sessionID)
	return writer.Write(event)
}

// RecordTaskEnd records a task completion event for cognitive budget tracking.
// Use when a task completes and captures completion metrics.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - taskID: Unique task identifier matching the task_start event
//   - agent: The agent that executed the task
//   - outcome: Task completion outcome (e.g., "success", "failed", "blocked")
//   - sessionID: The session ID context for the task
//   - durationMs: Task execution duration in milliseconds
//   - artifacts: List of artifact paths produced by the task
func RecordTaskEnd(sessionDir, taskID, agent, outcome, sessionID string, durationMs int64, artifacts []string) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewTaskEndEvent(taskID, agent, outcome, sessionID, durationMs, artifacts)
	return writer.Write(event)
}

// RecordSessionStart records a session initialization event.
// Use when a new tracked session is started via /start or session manager.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - sessionID: The unique session identifier
//   - initiative: The initiative or goal for this session
//   - complexity: Complexity rating (trivial, standard, complex, critical)
//   - team: The active team pack (e.g., "10x-dev-pack")
func RecordSessionStart(sessionDir, sessionID, initiative, complexity, team string) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewSessionStartEvent(sessionID, initiative, complexity, team)
	return writer.Write(event)
}

// RecordSessionEnd records a session completion event.
// Use when a tracked session ends via /wrap, /park, or session manager.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - sessionID: The unique session identifier
//   - status: Completion status (e.g., "completed", "parked", "abandoned")
//   - durationMs: Total session duration in milliseconds
func RecordSessionEnd(sessionDir, sessionID, status string, durationMs int64) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewSessionEndEvent(sessionID, status, durationMs)
	return writer.Write(event)
}

// RecordArtifactCreated records an artifact creation event.
// Use when a semantic artifact is created during a session, distinct from raw file changes.
// This enables handoff validation by tracking deliverables with their validation relationships.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - artifactType: The type of artifact (prd, tdd, adr, test_plan, code)
//   - path: Absolute path to the created artifact
//   - phase: The workflow phase during which the artifact was created
//   - validatesAgainst: Reference to the artifact this validates against (e.g., PRD path for TDD)
func RecordArtifactCreated(sessionDir string, artifactType ArtifactType, path, phase, validatesAgainst string) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewArtifactCreatedEvent(artifactType, path, phase, validatesAgainst)
	return writer.Write(event)
}

// RecordError records an error event with structured metadata.
// Use for capturing errors with recovery guidance during session execution.
// This provides dedicated error tracking for handoff protocols and debugging.
//
// Parameters:
//   - sessionDir: Path to the session directory containing events.jsonl
//   - errorCode: A structured error code (e.g., "VALIDATION_FAILED", "DEPENDENCY_MISSING")
//   - message: Human-readable error message
//   - context: Additional context about where/why the error occurred
//   - recoverable: Whether the error is recoverable without human intervention
//   - suggestedAction: Recommended action to resolve the error
func RecordError(sessionDir, errorCode, message, context string, recoverable bool, suggestedAction string) error {
	writer, err := NewEventWriter(sessionDir)
	if err != nil {
		return err
	}
	defer writer.Close()

	event := NewErrorEvent(errorCode, message, context, recoverable, suggestedAction)
	return writer.Write(event)
}

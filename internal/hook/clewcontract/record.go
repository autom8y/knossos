package clewcontract

import (
	"path/filepath"

	"github.com/autom8y/knossos/internal/hook"
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

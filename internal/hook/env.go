// Package hook provides hook infrastructure for Claude Code integration.
// It handles parsing hook environment variables, formatting output, and loading session context.
package hook

import (
	"encoding/json"
	"io"
	"os"
)

// EnvProjectDir is the only hook environment variable still set by Claude Code.
// All other hook data arrives via stdin JSON (see StdinPayload).
const EnvProjectDir = "CLAUDE_PROJECT_DIR"

// HookEvent represents the type of hook event.
type HookEvent string

// Known hook events from Claude Code.
// See: https://code.claude.com/docs/en/hooks
const (
	EventPreToolUse         HookEvent = "PreToolUse"
	EventPostToolUse        HookEvent = "PostToolUse"
	EventPostToolUseFailure HookEvent = "PostToolUseFailure"
	EventPermissionRequest  HookEvent = "PermissionRequest"
	EventStop               HookEvent = "Stop"
	EventSessionStart       HookEvent = "SessionStart"
	EventSessionEnd         HookEvent = "SessionEnd"
	EventUserPromptSubmit   HookEvent = "UserPromptSubmit"
	EventPreCompact         HookEvent = "PreCompact"
	EventSubagentStart      HookEvent = "SubagentStart"
	EventSubagentStop       HookEvent = "SubagentStop"
	EventNotification       HookEvent = "Notification"
	EventTeammateIdle       HookEvent = "TeammateIdle"
	EventTaskCompleted      HookEvent = "TaskCompleted"
)

// StdinPayload represents the JSON data Claude Code sends to hooks via stdin.
type StdinPayload struct {
	SessionID      string          `json:"session_id"`
	ConversationID string          `json:"conversation_id"`
	TranscriptPath string          `json:"transcript_path"`
	CWD            string          `json:"cwd"`
	PermissionMode string          `json:"permission_mode"`
	HookEventName  string          `json:"hook_event_name"`
	ToolName       string          `json:"tool_name"`
	ToolInput      json.RawMessage `json:"tool_input"`
	ToolResponse   json.RawMessage `json:"tool_response"`
	ToolUseID      string          `json:"tool_use_id"`
	Prompt         string          `json:"prompt"`
	Source         string          `json:"source"`
	StopHookActive bool            `json:"stop_hook_active"`
	Trigger        string          `json:"trigger"`
}

// Env holds parsed hook environment variables.
type Env struct {
	// Event type that triggered this hook
	Event HookEvent

	// Tool information (for PreToolUse/PostToolUse)
	ToolName   string
	ToolInput  string
	ToolResult string // Tool output/result (PostToolUse only)

	// Session context
	SessionID      string
	ProjectDir     string
	CWD            string // Working directory from CC stdin payload (distinct from ProjectDir)
	ConversationID string

	// Message context
	UserMessage string
}

// parseStdin reads and parses the JSON payload from stdin.
// Returns nil if stdin is a terminal or empty.
func parseStdin() *StdinPayload {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil
	}
	// If stdin is a terminal (no pipe), return nil
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		return nil
	}
	var payload StdinPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}
	return &payload
}

// ParseEnv reads hook data from stdin JSON and returns an Env.
// CLAUDE_PROJECT_DIR is the only value still read from environment.
func ParseEnv() *Env {
	stdin := parseStdin()

	projectDir := os.Getenv(EnvProjectDir)

	if stdin == nil {
		return &Env{
			ProjectDir: projectDir,
		}
	}

	event := HookEvent(stdin.HookEventName)
	if event != "" && !isValidHookEvent(event) {
		event = ""
	}

	var cwd string
	if stdin.CWD != "" {
		cwd = stdin.CWD
		if projectDir == "" {
			projectDir = stdin.CWD
		}
	}

	var toolInput string
	if len(stdin.ToolInput) > 0 && string(stdin.ToolInput) != "null" {
		toolInput = unwrapJSONValue(stdin.ToolInput)
	}

	var toolResult string
	if len(stdin.ToolResponse) > 0 && string(stdin.ToolResponse) != "null" {
		toolResult = unwrapJSONValue(stdin.ToolResponse)
	}

	return &Env{
		Event:          event,
		ToolName:       stdin.ToolName,
		ToolInput:      toolInput,
		ToolResult:     toolResult,
		SessionID:      stdin.SessionID,
		ProjectDir:     projectDir,
		CWD:            cwd,
		ConversationID: stdin.ConversationID,
		UserMessage:    stdin.Prompt,
	}
}

// isValidHookEvent checks if the provided event is a known HookEvent.
func isValidHookEvent(event HookEvent) bool {
	switch event {
	case EventPreToolUse, EventPostToolUse, EventPostToolUseFailure, EventPermissionRequest,
		EventStop, EventSessionStart, EventSessionEnd, EventUserPromptSubmit,
		EventPreCompact, EventSubagentStart, EventSubagentStop, EventNotification,
		EventTeammateIdle, EventTaskCompleted:
		return true
	default:
		return false
	}
}

// unwrapJSONValue converts a json.RawMessage to a string.
// If the value is a JSON string, it is unwrapped (unquoted).
// If it's an object or array, it is returned as raw JSON text.
func unwrapJSONValue(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}

// IsPreToolUse returns true if this is a PreToolUse event.
func (e *Env) IsPreToolUse() bool {
	return e.Event == EventPreToolUse
}

// IsPostToolUse returns true if this is a PostToolUse event.
func (e *Env) IsPostToolUse() bool {
	return e.Event == EventPostToolUse
}

// IsStop returns true if this is a Stop event.
func (e *Env) IsStop() bool {
	return e.Event == EventStop
}

// IsSessionStart returns true if this is a SessionStart event.
func (e *Env) IsSessionStart() bool {
	return e.Event == EventSessionStart
}

// IsPreCompact returns true if this is a PreCompact event.
func (e *Env) IsPreCompact() bool {
	return e.Event == EventPreCompact
}

// IsSubagentStart returns true if this is a SubagentStart event.
func (e *Env) IsSubagentStart() bool {
	return e.Event == EventSubagentStart
}

// HasTool returns true if tool information is available.
func (e *Env) HasTool() bool {
	return e.ToolName != ""
}

// HasSession returns true if session information is available.
func (e *Env) HasSession() bool {
	return e.SessionID != "" || e.ProjectDir != ""
}

// GetProjectDir returns the project directory, falling back to current working directory.
func (e *Env) GetProjectDir() string {
	if e.ProjectDir != "" {
		return e.ProjectDir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

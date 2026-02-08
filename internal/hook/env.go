// Package hook provides hook infrastructure for Claude Code integration.
// It handles parsing hook environment variables, formatting output, and loading session context.
package hook

import (
	"encoding/json"
	"io"
	"os"
)

// Claude Code hook environment variable names.
// Deprecated: CC sends these via stdin JSON, not env vars.
// Kept for backwards compatibility with direct CLI invocation.
const (
	EnvHookEvent     = "CLAUDE_HOOK_EVENT"       // Hook event type (PreToolUse, PostToolUse, etc.)
	EnvToolName      = "CLAUDE_TOOL_NAME"        // Name of the tool being used
	EnvToolInput     = "CLAUDE_TOOL_INPUT"       // JSON input to the tool
	EnvToolResult    = "CLAUDE_HOOK_TOOL_RESULT" // Tool result/output (PostToolUse only)
	EnvSessionID     = "CLAUDE_SESSION_ID"       // Claude session identifier
	EnvProjectDir    = "CLAUDE_PROJECT_DIR"      // Project root directory
	EnvConversation  = "CLAUDE_CONVERSATION_ID"  // Conversation identifier
	EnvUserMessage   = "CLAUDE_USER_MESSAGE"     // User message that triggered the action
	EnvAssistantText = "CLAUDE_ASSISTANT_TEXT"   // Assistant response text
)

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
	ConversationID string

	// Message context
	UserMessage   string
	AssistantText string
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

// ParseEnv reads hook-related environment variables and returns an Env.
func ParseEnv() *Env {
	// Read stdin first (primary source from CC)
	stdin := parseStdin()

	// Start with env var values (existing behavior)
	event := HookEvent(os.Getenv(EnvHookEvent))
	toolName := os.Getenv(EnvToolName)
	toolInput := os.Getenv(EnvToolInput)
	toolResult := os.Getenv(EnvToolResult)
	sessionID := os.Getenv(EnvSessionID)
	projectDir := os.Getenv(EnvProjectDir)
	userMessage := os.Getenv(EnvUserMessage)
	assistantText := os.Getenv(EnvAssistantText)
	conversationID := os.Getenv(EnvConversation)

	// Override with stdin values if available (CC's actual data)
	if stdin != nil {
		if stdin.HookEventName != "" {
			event = HookEvent(stdin.HookEventName)
		}
		if stdin.ToolName != "" {
			toolName = stdin.ToolName
		}
		if len(stdin.ToolInput) > 0 && string(stdin.ToolInput) != "null" {
			toolInput = string(stdin.ToolInput)
		}
		if len(stdin.ToolResponse) > 0 && string(stdin.ToolResponse) != "null" {
			toolResult = string(stdin.ToolResponse)
		}
		if stdin.SessionID != "" {
			sessionID = stdin.SessionID
		}
		if stdin.CWD != "" && projectDir == "" {
			projectDir = stdin.CWD
		}
		if stdin.Prompt != "" {
			userMessage = stdin.Prompt
		}
	}

	// Validate the event type if non-empty
	if event != "" && !isValidHookEvent(event) {
		// Log warning for invalid event type but don't fail
		// Empty event is valid (used for testing/direct invocation)
		event = ""
	}

	return &Env{
		Event:          event,
		ToolName:       toolName,
		ToolInput:      toolInput,
		ToolResult:     toolResult,
		SessionID:      sessionID,
		ProjectDir:     projectDir,
		ConversationID: conversationID,
		UserMessage:    userMessage,
		AssistantText:  assistantText,
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

// IsPreToolUse returns true if this is a PreToolUse event.
func (e *Env) IsPreToolUse() bool {
	return e.Event == EventPreToolUse
}

// IsPostToolUse returns true if this is a PostToolUse event.
func (e *Env) IsPostToolUse() bool {
	return e.Event == EventPostToolUse
}

// IsPostToolUseFailure returns true if this is a PostToolUseFailure event.
func (e *Env) IsPostToolUseFailure() bool {
	return e.Event == EventPostToolUseFailure
}

// IsPermissionRequest returns true if this is a PermissionRequest event.
func (e *Env) IsPermissionRequest() bool {
	return e.Event == EventPermissionRequest
}

// IsStop returns true if this is a Stop event.
func (e *Env) IsStop() bool {
	return e.Event == EventStop
}

// IsSessionStart returns true if this is a SessionStart event.
func (e *Env) IsSessionStart() bool {
	return e.Event == EventSessionStart
}

// IsSessionEnd returns true if this is a SessionEnd event.
func (e *Env) IsSessionEnd() bool {
	return e.Event == EventSessionEnd
}

// IsPreCompact returns true if this is a PreCompact event.
func (e *Env) IsPreCompact() bool {
	return e.Event == EventPreCompact
}

// IsSubagentStart returns true if this is a SubagentStart event.
func (e *Env) IsSubagentStart() bool {
	return e.Event == EventSubagentStart
}

// IsSubagentStop returns true if this is a SubagentStop event.
func (e *Env) IsSubagentStop() bool {
	return e.Event == EventSubagentStop
}

// IsNotification returns true if this is a Notification event.
func (e *Env) IsNotification() bool {
	return e.Event == EventNotification
}

// IsTeammateIdle returns true if this is a TeammateIdle event.
func (e *Env) IsTeammateIdle() bool {
	return e.Event == EventTeammateIdle
}

// IsTaskCompleted returns true if this is a TaskCompleted event.
func (e *Env) IsTaskCompleted() bool {
	return e.Event == EventTaskCompleted
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

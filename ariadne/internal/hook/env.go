// Package hook provides hook infrastructure for Claude Code integration.
// It handles parsing hook environment variables, formatting output, and loading session context.
package hook

import (
	"os"
	"strings"
)

// FeatureFlagEnvVar is the environment variable that enables ari hooks.
const FeatureFlagEnvVar = "USE_ARI_HOOKS"

// Claude Code hook environment variable names.
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
const (
	EventPreToolUse        HookEvent = "PreToolUse"
	EventPostToolUse       HookEvent = "PostToolUse"
	EventStop              HookEvent = "Stop"
	EventSessionStart      HookEvent = "SessionStart"
	EventUserPromptSubmit  HookEvent = "UserPromptSubmit"
)

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

// ParseEnv reads hook-related environment variables and returns an Env.
func ParseEnv() *Env {
	return &Env{
		Event:          HookEvent(os.Getenv(EnvHookEvent)),
		ToolName:       os.Getenv(EnvToolName),
		ToolInput:      os.Getenv(EnvToolInput),
		ToolResult:     os.Getenv(EnvToolResult),
		SessionID:      os.Getenv(EnvSessionID),
		ProjectDir:     os.Getenv(EnvProjectDir),
		ConversationID: os.Getenv(EnvConversation),
		UserMessage:    os.Getenv(EnvUserMessage),
		AssistantText:  os.Getenv(EnvAssistantText),
	}
}

// IsEnabled returns true if ari hooks are enabled via the feature flag.
func IsEnabled() bool {
	val := os.Getenv(FeatureFlagEnvVar)
	return val == "1" || strings.ToLower(val) == "true"
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

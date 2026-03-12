// Package hook provides hook infrastructure for Claude Code integration.
// It handles parsing hook environment variables, formatting output, and loading session context.
package hook

import (
	"encoding/json"
	"os"
)

// EnvProjectDir is the only hook environment variable still set by Claude Code.
// All other hook data arrives via stdin JSON (see StdinPayload).
const EnvProjectDir = "CLAUDE_PROJECT_DIR"

// HookEvent represents the type of hook event.
type HookEvent string

// Knossos canonical hook events (ADR-0032).
// String values are snake_case canonical names; adapters translate to/from wire format.
const (
	EventPreTool           HookEvent = "pre_tool"
	EventPostTool          HookEvent = "post_tool"
	EventPostToolFailure   HookEvent = "post_tool_failure"
	EventPermissionRequest HookEvent = "permission_request"
	EventStop              HookEvent = "stop"
	EventSessionStart      HookEvent = "session_start"
	EventSessionEnd        HookEvent = "session_end"
	EventPrePrompt         HookEvent = "pre_prompt"
	EventPreCompact        HookEvent = "pre_compact"
	EventSubagentStart     HookEvent = "subagent_start"
	EventSubagentStop      HookEvent = "subagent_stop"
	EventNotification      HookEvent = "notification"
	EventTeammateIdle      HookEvent = "teammate_idle"
	EventTaskCompleted     HookEvent = "task_completed"
	// Gemini-exclusive events (no CC wire equivalent)
	EventPreModel  HookEvent = "pre_model"
	EventPostModel HookEvent = "post_model"
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

	// XKnossosSignature is the HMAC-SHA256 signature of the payload.
	// Injected by Knossos client when calling hooks.
	XKnossosSignature string `json:"x_knossos_signature"`
}

// Env holds parsed hook environment variables.
type Env struct {
	// Event type that triggered this hook
	Event HookEvent

	// Tool information (for pre_tool/post_tool events)
	ToolName   string
	ToolInput  string
	ToolResult string // Tool output/result (post_tool only)

	// Session context
	SessionID      string
	ProjectDir     string
	CWD            string // Working directory from CC stdin payload (distinct from ProjectDir)
	ConversationID string

	// Message context
	UserMessage string

	// Authentication
	Signature  string // HMAC signature from payload
	RawPayload []byte // Full raw stdin payload for verification
}

// GetAdapter returns the appropriate LifecycleAdapter based on the KNOSSOS_CHANNEL env var.
// NOTE (HA-1-011): When KNOSSOS_CHANNEL is unset, this implicitly defaults to ClaudeAdapter.
// This is intentional for backward compatibility. Do NOT change the default behavior here;
// it is tracked as a behavioral coupling item in the harness-agnosticism initiative (PKG-010).
func GetAdapter() LifecycleAdapter {
	if os.Getenv("KNOSSOS_CHANNEL") == "gemini" {
		return &GeminiAdapter{}
	}
	return &ClaudeAdapter{}
}

// ParseEnv reads hook data from stdin JSON and returns an Env.
// CLAUDE_PROJECT_DIR is the only value still read from environment.
func ParseEnv() *Env {
	adapter := GetAdapter()

	// We must check if stdin is a terminal before reading
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return &Env{ProjectDir: os.Getenv(EnvProjectDir)}
	}

	env, err := adapter.ParsePayload(os.Stdin)
	if err != nil || env == nil {
		return &Env{ProjectDir: os.Getenv(EnvProjectDir)}
	}
	return env
}

// isValidHookEvent checks if the provided event is a known HookEvent.
func isValidHookEvent(event HookEvent) bool {
	switch event {
	case EventPreTool, EventPostTool, EventPostToolFailure, EventPermissionRequest,
		EventStop, EventSessionStart, EventSessionEnd, EventPrePrompt,
		EventPreCompact, EventSubagentStart, EventSubagentStop, EventNotification,
		EventTeammateIdle, EventTaskCompleted, EventPreModel, EventPostModel:
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

// IsPreTool returns true if this is a pre_tool event.
func (e *Env) IsPreTool() bool {
	return e.Event == EventPreTool
}

// IsPostTool returns true if this is a post_tool event.
func (e *Env) IsPostTool() bool {
	return e.Event == EventPostTool
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

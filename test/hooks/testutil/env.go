// Package testutil provides test utilities for hook testing.
// It enables testing hooks without running Claude Code.
package testutil

import (
	"os"
	"testing"
)

// EnvSetup holds original environment values for restoration.
type EnvSetup struct {
	original map[string]string
	t        *testing.T
}

// HookEnv represents the standard Claude Code hook environment variables.
type HookEnv struct {
	Event          string
	ToolName       string
	ToolInput      string
	ToolResult     string // Tool result/output (PostToolUse only)
	SessionID      string
	ProjectDir     string
	ConversationID string
	UserMessage    string
	AssistantText  string
}

// SetupEnv creates a test environment with Claude Code hook variables.
// It automatically cleans up when the test completes.
func SetupEnv(t *testing.T, env *HookEnv) *EnvSetup {
	t.Helper()

	setup := &EnvSetup{
		original: make(map[string]string),
		t:        t,
	}

	// Capture and set each environment variable
	vars := map[string]string{
		"CLAUDE_HOOK_EVENT":       env.Event,
		"CLAUDE_TOOL_NAME":        env.ToolName,
		"CLAUDE_TOOL_INPUT":       env.ToolInput,
		"CLAUDE_HOOK_TOOL_RESULT": env.ToolResult,
		"CLAUDE_SESSION_ID":       env.SessionID,
		"CLAUDE_PROJECT_DIR":      env.ProjectDir,
		"CLAUDE_CONVERSATION_ID":  env.ConversationID,
		"CLAUDE_USER_MESSAGE":     env.UserMessage,
		"CLAUDE_ASSISTANT_TEXT":   env.AssistantText,
	}

	for key, value := range vars {
		setup.original[key] = os.Getenv(key)
		if value != "" {
			if err := os.Setenv(key, value); err != nil {
				t.Fatalf("SetupEnv: os.Setenv(%q): %v", key, err)
			}
		} else {
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("SetupEnv: os.Unsetenv(%q): %v", key, err)
			}
		}
	}

	// Register cleanup
	t.Cleanup(func() {
		setup.Restore()
	})

	return setup
}

// Restore restores the original environment.
func (s *EnvSetup) Restore() {
	for key, value := range s.original {
		if value == "" {
			if err := os.Unsetenv(key); err != nil {
				s.t.Logf("Restore: os.Unsetenv(%q): %v", key, err)
			}
		} else {
			if err := os.Setenv(key, value); err != nil {
				s.t.Logf("Restore: os.Setenv(%q): %v", key, err)
			}
		}
	}
}

// SetVar sets an additional environment variable (captured for restoration).
func (s *EnvSetup) SetVar(key, value string) {
	if _, exists := s.original[key]; !exists {
		s.original[key] = os.Getenv(key)
	}
	if err := os.Setenv(key, value); err != nil {
		s.t.Fatalf("SetVar: os.Setenv(%q): %v", key, err)
	}
}

// UnsetVar unsets an environment variable (captured for restoration).
func (s *EnvSetup) UnsetVar(key string) {
	if _, exists := s.original[key]; !exists {
		s.original[key] = os.Getenv(key)
	}
	if err := os.Unsetenv(key); err != nil {
		s.t.Fatalf("UnsetVar: os.Unsetenv(%q): %v", key, err)
	}
}

// PresetEnvs provides common hook environment configurations for testing.
var PresetEnvs = struct {
	// PreToolUseBash simulates a PreToolUse event for Bash tool
	PreToolUseBash HookEnv
	// PreToolUseWrite simulates a PreToolUse event for Write tool
	PreToolUseWrite HookEnv
	// PreToolUseEdit simulates a PreToolUse event for Edit tool
	PreToolUseEdit HookEnv
	// PostToolUseBash simulates a PostToolUse event for Bash tool
	PostToolUseBash HookEnv
	// SessionStart simulates a SessionStart event
	SessionStart HookEnv
	// Stop simulates a Stop event
	Stop HookEnv
}{
	PreToolUseBash: HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Bash",
		ToolInput: `{"command":"ls -la","description":"List files"}`,
	},
	PreToolUseWrite: HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Write",
		ToolInput: `{"file_path":"/tmp/test.txt","content":"hello world"}`,
	},
	PreToolUseEdit: HookEnv{
		Event:     "PreToolUse",
		ToolName:  "Edit",
		ToolInput: `{"file_path":"/tmp/test.txt","old_string":"hello","new_string":"goodbye"}`,
	},
	PostToolUseBash: HookEnv{
		Event:    "PostToolUse",
		ToolName: "Bash",
	},
	SessionStart: HookEnv{
		Event: "SessionStart",
	},
	Stop: HookEnv{
		Event: "Stop",
	},
}

// WithSession adds session context to a HookEnv.
func (h HookEnv) WithSession(sessionID, projectDir string) HookEnv {
	h.SessionID = sessionID
	h.ProjectDir = projectDir
	return h
}

// WithToolInput sets the tool input JSON.
func (h HookEnv) WithToolInput(input string) HookEnv {
	h.ToolInput = input
	return h
}

// WithToolResult sets the tool result output.
func (h HookEnv) WithToolResult(result string) HookEnv {
	h.ToolResult = result
	return h
}

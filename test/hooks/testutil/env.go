// Package testutil provides test utilities for hook testing.
// It enables testing hooks without running Claude Code by piping
// JSON payloads to stdin (matching CC's actual transport).
package testutil

import (
	"encoding/json"
	"os"
	"testing"
)

// EnvSetup holds original state for restoration.
type EnvSetup struct {
	originalStdin *os.File
	originalPD    string
	hadPD         bool
	t             *testing.T
}

// HookEnv represents the hook data to inject for testing.
type HookEnv struct {
	Event          string
	ToolName       string
	ToolInput      string
	ToolResult     string
	SessionID      string
	ProjectDir     string
	ConversationID string
	UserMessage    string
}

// SetupEnv creates a test environment by piping hook data as JSON to stdin.
// CLAUDE_PROJECT_DIR is set as an env var (matching CC's actual behavior).
// All other hook data goes through stdin JSON.
//
// ToolInput and ToolResult are embedded as-is if they are valid JSON,
// or escaped as JSON strings if they are not (for testing graceful degradation).
func SetupEnv(t *testing.T, env *HookEnv) *EnvSetup {
	t.Helper()

	setup := &EnvSetup{
		originalStdin: os.Stdin,
		originalPD:    os.Getenv("CLAUDE_PROJECT_DIR"),
		hadPD:         os.Getenv("CLAUDE_PROJECT_DIR") != "",
		t:             t,
	}

	// Set CLAUDE_PROJECT_DIR (the one env var CC still sends)
	if env.ProjectDir != "" {
		if err := os.Setenv("CLAUDE_PROJECT_DIR", env.ProjectDir); err != nil {
			t.Fatalf("SetupEnv: os.Setenv: %v", err)
		}
	}

	// Build payload as a map so we can handle non-JSON ToolInput/ToolResult
	payload := map[string]any{
		"hook_event_name": env.Event,
		"tool_name":       env.ToolName,
		"session_id":      env.SessionID,
		"conversation_id": env.ConversationID,
		"cwd":             env.ProjectDir,
		"prompt":          env.UserMessage,
	}
	if env.ToolInput != "" {
		if json.Valid([]byte(env.ToolInput)) {
			payload["tool_input"] = json.RawMessage(env.ToolInput)
		} else {
			payload["tool_input"] = env.ToolInput // marshals as JSON string
		}
	}
	if env.ToolResult != "" {
		if json.Valid([]byte(env.ToolResult)) {
			payload["tool_response"] = json.RawMessage(env.ToolResult)
		} else {
			payload["tool_response"] = env.ToolResult // marshals as JSON string
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("SetupEnv: json.Marshal: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("SetupEnv: os.Pipe: %v", err)
	}
	go func() {
		_, _ = w.Write(data)
		_ = w.Close()
	}()
	os.Stdin = r

	t.Cleanup(func() {
		setup.Restore()
	})

	return setup
}

// Restore restores the original stdin and environment.
func (s *EnvSetup) Restore() {
	os.Stdin = s.originalStdin
	if s.hadPD {
		_ = os.Setenv("CLAUDE_PROJECT_DIR", s.originalPD)
	} else {
		_ = os.Unsetenv("CLAUDE_PROJECT_DIR")
	}
}

// SetVar sets an additional environment variable (captured for restoration).
func (s *EnvSetup) SetVar(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		s.t.Fatalf("SetVar: os.Setenv(%q): %v", key, err)
	}
	s.t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})
}

// UnsetVar unsets an environment variable.
func (s *EnvSetup) UnsetVar(key string) {
	if err := os.Unsetenv(key); err != nil {
		s.t.Fatalf("UnsetVar: os.Unsetenv(%q): %v", key, err)
	}
}

// PresetEnvs provides common hook environment configurations for testing.
var PresetEnvs = struct {
	PreToolUseBash  HookEnv
	PreToolUseWrite HookEnv
	PreToolUseEdit  HookEnv
	PostToolUseBash HookEnv
	SessionStart    HookEnv
	Stop            HookEnv
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

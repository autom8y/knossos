package hook

import (
	"os"
	"testing"
)

func TestParseEnv(t *testing.T) {
	// Save and restore environment
	originalEnv := map[string]string{
		EnvHookEvent:     os.Getenv(EnvHookEvent),
		EnvToolName:      os.Getenv(EnvToolName),
		EnvToolInput:     os.Getenv(EnvToolInput),
		EnvSessionID:     os.Getenv(EnvSessionID),
		EnvProjectDir:    os.Getenv(EnvProjectDir),
		EnvConversation:  os.Getenv(EnvConversation),
		EnvUserMessage:   os.Getenv(EnvUserMessage),
		EnvAssistantText: os.Getenv(EnvAssistantText),
	}
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Set test values
	os.Setenv(EnvHookEvent, "PreToolUse")
	os.Setenv(EnvToolName, "Bash")
	os.Setenv(EnvToolInput, `{"command":"ls"}`)
	os.Setenv(EnvSessionID, "test-session-123")
	os.Setenv(EnvProjectDir, "/test/project")
	os.Setenv(EnvConversation, "conv-456")
	os.Setenv(EnvUserMessage, "run ls")
	os.Setenv(EnvAssistantText, "I'll list files")

	env := ParseEnv()

	if env.Event != EventPreToolUse {
		t.Errorf("Expected event %s, got %s", EventPreToolUse, env.Event)
	}
	if env.ToolName != "Bash" {
		t.Errorf("Expected tool name Bash, got %s", env.ToolName)
	}
	if env.ToolInput != `{"command":"ls"}` {
		t.Errorf("Expected tool input, got %s", env.ToolInput)
	}
	if env.SessionID != "test-session-123" {
		t.Errorf("Expected session ID test-session-123, got %s", env.SessionID)
	}
	if env.ProjectDir != "/test/project" {
		t.Errorf("Expected project dir /test/project, got %s", env.ProjectDir)
	}
	if env.ConversationID != "conv-456" {
		t.Errorf("Expected conversation ID conv-456, got %s", env.ConversationID)
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty", "", false},
		{"zero", "0", false},
		{"one", "1", true},
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"True mixed", "True", true},
		{"false", "false", false},
		{"random", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv(FeatureFlagEnvVar)
			defer func() {
				if original == "" {
					os.Unsetenv(FeatureFlagEnvVar)
				} else {
					os.Setenv(FeatureFlagEnvVar, original)
				}
			}()

			if tt.value == "" {
				os.Unsetenv(FeatureFlagEnvVar)
			} else {
				os.Setenv(FeatureFlagEnvVar, tt.value)
			}

			if got := IsEnabled(); got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEnvEventChecks(t *testing.T) {
	tests := []struct {
		event        HookEvent
		isPreTool    bool
		isPostTool   bool
		isStop       bool
		isSessionStart bool
	}{
		{EventPreToolUse, true, false, false, false},
		{EventPostToolUse, false, true, false, false},
		{EventStop, false, false, true, false},
		{EventSessionStart, false, false, false, true},
		{"UnknownEvent", false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			env := &Env{Event: tt.event}

			if got := env.IsPreToolUse(); got != tt.isPreTool {
				t.Errorf("IsPreToolUse() = %v, want %v", got, tt.isPreTool)
			}
			if got := env.IsPostToolUse(); got != tt.isPostTool {
				t.Errorf("IsPostToolUse() = %v, want %v", got, tt.isPostTool)
			}
			if got := env.IsStop(); got != tt.isStop {
				t.Errorf("IsStop() = %v, want %v", got, tt.isStop)
			}
			if got := env.IsSessionStart(); got != tt.isSessionStart {
				t.Errorf("IsSessionStart() = %v, want %v", got, tt.isSessionStart)
			}
		})
	}
}

func TestEnvHasTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		expected bool
	}{
		{"with tool", "Bash", true},
		{"empty tool", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Env{ToolName: tt.toolName}
			if got := env.HasTool(); got != tt.expected {
				t.Errorf("HasTool() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEnvHasSession(t *testing.T) {
	tests := []struct {
		name       string
		sessionID  string
		projectDir string
		expected   bool
	}{
		{"both set", "sess-123", "/project", true},
		{"only session", "sess-123", "", true},
		{"only project", "", "/project", true},
		{"neither", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Env{SessionID: tt.sessionID, ProjectDir: tt.projectDir}
			if got := env.HasSession(); got != tt.expected {
				t.Errorf("HasSession() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEnvGetProjectDir(t *testing.T) {
	t.Run("returns set project dir", func(t *testing.T) {
		env := &Env{ProjectDir: "/my/project"}
		if got := env.GetProjectDir(); got != "/my/project" {
			t.Errorf("GetProjectDir() = %v, want /my/project", got)
		}
	})

	t.Run("returns cwd when empty", func(t *testing.T) {
		env := &Env{ProjectDir: ""}
		got := env.GetProjectDir()
		// Should return cwd, not empty
		if got == "" {
			t.Error("GetProjectDir() returned empty, expected cwd")
		}
	})
}

package hook

import (
	"os"
	"strings"
	"testing"
)

func TestParseEnv_StdinOnly(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"ls"},"session_id":"test-session-123","cwd":"/test/project","conversation_id":"conv-456","prompt":"run ls"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

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
	if env.UserMessage != "run ls" {
		t.Errorf("Expected user message 'run ls', got %s", env.UserMessage)
	}
}

func TestParseEnv_EmptyStdin_FallsBackToProjectDir(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Set CLAUDE_PROJECT_DIR (the one env var CC still sets)
	origPD := os.Getenv(EnvProjectDir)
	os.Setenv(EnvProjectDir, "/env/project")
	defer func() {
		if origPD == "" {
			os.Unsetenv(EnvProjectDir)
		} else {
			os.Setenv(EnvProjectDir, origPD)
		}
	}()

	// Empty stdin
	r, w, _ := os.Pipe()
	w.Close()
	os.Stdin = r

	env := ParseEnv()

	// Only ProjectDir should be populated from env
	if env.ProjectDir != "/env/project" {
		t.Errorf("ProjectDir = %q, want /env/project", env.ProjectDir)
	}
	if env.Event != "" {
		t.Errorf("Event should be empty without stdin, got %q", env.Event)
	}
	if env.ToolName != "" {
		t.Errorf("ToolName should be empty without stdin, got %q", env.ToolName)
	}
	if env.SessionID != "" {
		t.Errorf("SessionID should be empty without stdin, got %q", env.SessionID)
	}
}

func TestEnvEventChecks(t *testing.T) {
	tests := []struct {
		event           HookEvent
		isPreTool       bool
		isPostTool      bool
		isStop          bool
		isSessionStart  bool
		isPreCompact    bool
		isSubagentStart bool
	}{
		{EventPreToolUse, true, false, false, false, false, false},
		{EventPostToolUse, false, true, false, false, false, false},
		{EventStop, false, false, true, false, false, false},
		{EventSessionStart, false, false, false, true, false, false},
		{EventPreCompact, false, false, false, false, true, false},
		{EventSubagentStart, false, false, false, false, false, true},
		{"UnknownEvent", false, false, false, false, false, false},
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
			if got := env.IsPreCompact(); got != tt.isPreCompact {
				t.Errorf("IsPreCompact() = %v, want %v", got, tt.isPreCompact)
			}
			if got := env.IsSubagentStart(); got != tt.isSubagentStart {
				t.Errorf("IsSubagentStart() = %v, want %v", got, tt.isSubagentStart)
			}
		})
	}
}

func TestValidHookEvents(t *testing.T) {
	validEvents := []HookEvent{
		EventPreToolUse,
		EventPostToolUse,
		EventPostToolUseFailure,
		EventPermissionRequest,
		EventStop,
		EventSessionStart,
		EventSessionEnd,
		EventUserPromptSubmit,
		EventPreCompact,
		EventSubagentStart,
		EventSubagentStop,
		EventNotification,
		EventTeammateIdle,
		EventTaskCompleted,
	}

	for _, event := range validEvents {
		t.Run(string(event), func(t *testing.T) {
			if !isValidHookEvent(event) {
				t.Errorf("isValidHookEvent(%q) = false, want true", event)
			}
		})
	}

	invalidEvents := []HookEvent{
		"UnknownEvent",
		"InvalidEvent",
		"",
	}

	for _, event := range invalidEvents {
		t.Run(string(event), func(t *testing.T) {
			if isValidHookEvent(event) {
				t.Errorf("isValidHookEvent(%q) = true, want false", event)
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
		if got == "" {
			t.Error("GetProjectDir() returned empty, expected cwd")
		}
	})
}

func TestParseStdin_PopulatesEnv(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":"/tmp/test.txt","content":"hello"},"tool_response":{"success":true},"session_id":"test-123","cwd":"/home/user/project"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

	env := ParseEnv()

	if env.Event != EventPostToolUse {
		t.Errorf("Event = %q, want %q", env.Event, EventPostToolUse)
	}
	if env.ToolName != "Write" {
		t.Errorf("ToolName = %q, want %q", env.ToolName, "Write")
	}
	if env.SessionID != "test-123" {
		t.Errorf("SessionID = %q, want %q", env.SessionID, "test-123")
	}
	if env.ProjectDir != "/home/user/project" {
		t.Errorf("ProjectDir = %q, want %q", env.ProjectDir, "/home/user/project")
	}
	if !strings.Contains(env.ToolInput, "file_path") {
		t.Errorf("ToolInput should contain file_path, got %q", env.ToolInput)
	}
	if !strings.Contains(env.ToolInput, "/tmp/test.txt") {
		t.Errorf("ToolInput should contain /tmp/test.txt, got %q", env.ToolInput)
	}
	if !strings.Contains(env.ToolResult, "success") {
		t.Errorf("ToolResult should contain success, got %q", env.ToolResult)
	}
}

func TestParseStdin_ToolInputReserializedAsString(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/src/main.go","old_string":"foo","new_string":"bar"}}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

	env := ParseEnv()

	if env.ToolInput == "" {
		t.Error("ToolInput should not be empty")
	}
	if !strings.HasPrefix(env.ToolInput, "{") {
		t.Errorf("ToolInput should be JSON string starting with {, got %q", env.ToolInput)
	}
	if !strings.Contains(env.ToolInput, "file_path") {
		t.Errorf("ToolInput should contain file_path, got %q", env.ToolInput)
	}
	if !strings.Contains(env.ToolInput, "old_string") {
		t.Errorf("ToolInput should contain old_string, got %q", env.ToolInput)
	}
	if !strings.Contains(env.ToolInput, "new_string") {
		t.Errorf("ToolInput should contain new_string, got %q", env.ToolInput)
	}
}

func TestParseStdin_UserPromptSubmit(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	payload := `{"hook_event_name":"UserPromptSubmit","session_id":"prompt-session","prompt":"/sos start my-initiative"}`
	go func() {
		w.Write([]byte(payload))
		w.Close()
	}()
	os.Stdin = r

	env := ParseEnv()

	if env.Event != EventUserPromptSubmit {
		t.Errorf("Event = %q, want %q", env.Event, EventUserPromptSubmit)
	}
	if env.UserMessage != "/sos start my-initiative" {
		t.Errorf("UserMessage = %q, want %q", env.UserMessage, "/sos start my-initiative")
	}
	if env.SessionID != "prompt-session" {
		t.Errorf("SessionID = %q, want %q", env.SessionID, "prompt-session")
	}
}

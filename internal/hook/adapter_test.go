package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestClaudeAdapter_ParsePayload(t *testing.T) {
	adapter := &ClaudeAdapter{}
	
	payload := `{"session_id": "sess-123", "hook_event_name": "PreToolUse", "tool_name": "ReadFiles"}`
	reader := bytes.NewReader([]byte(payload))

	env, err := adapter.ParsePayload(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.SessionID != "sess-123" {
		t.Errorf("expected session_id sess-123, got %s", env.SessionID)
	}
	if env.Event != EventPreToolUse {
		t.Errorf("expected EventPreToolUse, got %s", env.Event)
	}
}

func TestGeminiAdapter_ParsePayload(t *testing.T) {
	adapter := &GeminiAdapter{}
	
	payload := `{"sessionId": "sess-gemini", "event": "PreToolUse", "toolName": "WriteFile"}`
	reader := bytes.NewReader([]byte(payload))

	env, err := adapter.ParsePayload(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.SessionID != "sess-gemini" {
		t.Errorf("expected sessionId sess-gemini, got %s", env.SessionID)
	}
	if env.Event != EventPreToolUse {
		t.Errorf("expected EventPreToolUse, got %s", env.Event)
	}
}

func TestGeminiAdapter_FormatResponse(t *testing.T) {
	// Note: testing os.Exit is tricky, so we just test the allow case which doesn't exit
	adapter := &GeminiAdapter{}
	
	resp, err := adapter.FormatResponse("allow", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty response for Gemini allow, got %s", string(resp))
	}
}

func TestChannelDetection_DefaultClaude(t *testing.T) {
	os.Unsetenv("KNOSSOS_CHANNEL")
	
	adapter := GetAdapter()
	if adapter.ChannelName() != "claude" {
		t.Errorf("expected claude, got %s", adapter.ChannelName())
	}
}

func TestChannelDetection_GeminiEnvVar(t *testing.T) {
	os.Setenv("KNOSSOS_CHANNEL", "gemini")
	defer os.Unsetenv("KNOSSOS_CHANNEL")
	
	adapter := GetAdapter()
	if adapter.ChannelName() != "gemini" {
		t.Errorf("expected gemini, got %s", adapter.ChannelName())
	}
}

func TestClaudeAdapter_FormatResponse(t *testing.T) {
	adapter := &ClaudeAdapter{}
	resp, err := adapter.FormatResponse("deny", "not allowed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	var m map[string]string
	if err := json.Unmarshal(resp, &m); err != nil {
		t.Fatal(err)
	}
	if m["decision"] != "deny" {
		t.Errorf("expected deny, got %v", m["decision"])
	}
	if m["reason"] != "not allowed" {
		t.Errorf("expected not allowed, got %v", m["reason"])
	}
}

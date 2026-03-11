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

	// Gemini sends snake_case fields with Gemini wire event names
	payload := `{"session_id": "sess-gemini", "hook_event_name": "BeforeTool", "tool_name": "replace"}`
	reader := bytes.NewReader([]byte(payload))

	env, err := adapter.ParsePayload(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.SessionID != "sess-gemini" {
		t.Errorf("expected session_id sess-gemini, got %s", env.SessionID)
	}
	// BeforeTool should be translated to PreToolUse
	if env.Event != EventPreToolUse {
		t.Errorf("expected EventPreToolUse after translation, got %s", env.Event)
	}
}

func TestGeminiAdapter_EventTranslation(t *testing.T) {
	adapter := &GeminiAdapter{}

	payload := `{"session_id": "sess-1", "hook_event_name": "BeforeTool", "tool_name": "replace"}`
	reader := bytes.NewReader([]byte(payload))

	env, err := adapter.ParsePayload(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.Event != EventPreToolUse {
		t.Errorf("BeforeTool should translate to EventPreToolUse, got %s", env.Event)
	}
}

func TestGeminiAdapter_UnknownEventDropped(t *testing.T) {
	adapter := &GeminiAdapter{}

	// BeforeModel is a Gemini-only event with no CC equivalent
	payload := `{"session_id": "sess-1", "hook_event_name": "BeforeModel"}`
	reader := bytes.NewReader([]byte(payload))

	env, err := adapter.ParsePayload(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Unknown Gemini-only event should result in empty event
	if env.Event != "" {
		t.Errorf("BeforeModel should produce empty event, got %s", env.Event)
	}
}

func TestGeminiAdapter_CCFieldsEmpty(t *testing.T) {
	adapter := &GeminiAdapter{}

	// Gemini payload has no conversation_id -- should parse without error
	payload := `{"session_id": "sess-gemini", "hook_event_name": "SessionStart"}`
	reader := bytes.NewReader([]byte(payload))

	env, err := adapter.ParsePayload(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.ConversationID != "" {
		t.Errorf("ConversationID should be empty for Gemini payload, got %s", env.ConversationID)
	}
	if env.SessionID != "sess-gemini" {
		t.Errorf("SessionID = %q, want sess-gemini", env.SessionID)
	}
}

func TestGeminiAdapter_FormatResponse_Allow(t *testing.T) {
	adapter := &GeminiAdapter{}

	resp, err := adapter.FormatResponse("allow", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(resp, &m); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if m["decision"] != "allow" {
		t.Errorf("decision = %v, want allow", m["decision"])
	}
	if _, hasReason := m["reason"]; hasReason {
		t.Error("reason should be absent when empty")
	}
}

func TestGeminiAdapter_FormatResponse_Block(t *testing.T) {
	adapter := &GeminiAdapter{}

	resp, err := adapter.FormatResponse("deny", "blocked by writeguard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(resp, &m); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if m["decision"] != "deny" {
		t.Errorf("decision = %v, want deny", m["decision"])
	}
	if m["reason"] != "blocked by writeguard" {
		t.Errorf("reason = %v, want blocked by writeguard", m["reason"])
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

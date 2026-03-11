package clewcontract

import (
	"encoding/json"
	"testing"
)

func TestEvent_WithChannel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		channel string
		want    string
	}{
		{"gemini sets field", "gemini", "gemini"},
		{"arbitrary channel sets field", "cursor", "cursor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := NewToolCallEvent("Bash", "/tmp/test", nil)
			e = e.WithChannel(tt.channel)
			if e.Channel != tt.want {
				t.Errorf("WithChannel(%q).Channel = %q, want %q", tt.channel, e.Channel, tt.want)
			}
		})
	}
}

func TestEvent_WithChannel_Claude(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		channel string
	}{
		{"claude is default, no change", "claude"},
		{"empty is default, no change", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := NewToolCallEvent("Bash", "/tmp/test", nil)
			original := e.Channel
			e = e.WithChannel(tt.channel)
			if e.Channel != original {
				t.Errorf("WithChannel(%q).Channel = %q, want %q (unchanged)", tt.channel, e.Channel, original)
			}
		})
	}
}

func TestTypedEvent_ChannelField(t *testing.T) {
	t.Parallel()
	event := NewTypedSessionCreatedEvent("gemini", "s-001", "dark mode", "MODULE", "ecosystem")

	if event.Channel != "gemini" {
		t.Errorf("Channel = %q, want %q", event.Channel, "gemini")
	}

	// Verify serialization includes channel
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if raw["channel"] != "gemini" {
		t.Errorf("serialized channel = %v, want %q", raw["channel"], "gemini")
	}
}

func TestTypedEvent_ChannelOmitted(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		channel string
	}{
		{"claude channel omitted", "claude"},
		{"empty channel omitted", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event := NewTypedSessionCreatedEvent(tt.channel, "s-001", "dark mode", "MODULE", "ecosystem")

			if event.Channel != "" {
				t.Errorf("Channel = %q, want empty (omitted for %q)", event.Channel, tt.channel)
			}

			// Verify serialization omits channel
			data, err := json.Marshal(event)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if _, ok := raw["channel"]; ok {
				t.Errorf("channel should be omitted for %q, but was present", tt.channel)
			}
		})
	}
}

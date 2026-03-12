package hook

import "testing"

func TestTranslateEventForChannel_ClaudeTranslated(t *testing.T) {
	t.Parallel()
	// Canonical -> CC wire
	event, skip := TranslateEventForChannel("pre_tool", "claude")
	if skip {
		t.Error("claude channel should never skip pre_tool")
	}
	if event != "PreToolUse" {
		t.Errorf("got %q, want %q", event, "PreToolUse")
	}
}

func TestTranslateEventForChannel_GeminiTranslated(t *testing.T) {
	t.Parallel()
	event, skip := TranslateEventForChannel("pre_tool", "gemini")
	if skip {
		t.Error("pre_tool is translatable, should not be skipped")
	}
	if event != "BeforeTool" {
		t.Errorf("got %q, want %q", event, "BeforeTool")
	}
}

func TestTranslateEventForChannel_GeminiSkipsUnmappable(t *testing.T) {
	t.Parallel()
	event, skip := TranslateEventForChannel("stop", "gemini")
	if !skip {
		t.Error("stop has no Gemini equivalent, should be skipped")
	}
	if event != "" {
		t.Errorf("skipped event should be empty string, got %q", event)
	}
}

func TestTranslateEventForChannel_GeminiIdentityMapping(t *testing.T) {
	t.Parallel()
	event, skip := TranslateEventForChannel("session_start", "gemini")
	if skip {
		t.Error("session_start is translatable, should not be skipped")
	}
	if event != "SessionStart" {
		t.Errorf("got %q, want %q", event, "SessionStart")
	}
}

func TestTranslateEventForChannel_GeminiExclusive(t *testing.T) {
	t.Parallel()
	// pre_model is Gemini-exclusive; claude should have no wire
	event, skip := TranslateEventForChannel("pre_model", "claude")
	if !skip {
		t.Error("pre_model has no CC equivalent, should be skipped for claude")
	}
	if event != "" {
		t.Errorf("skipped event should be empty string, got %q", event)
	}

	// Gemini should get BeforeModel
	event, skip = TranslateEventForChannel("pre_model", "gemini")
	if skip {
		t.Error("pre_model should translate for gemini")
	}
	if event != "BeforeModel" {
		t.Errorf("got %q, want %q", event, "BeforeModel")
	}
}

func TestWireToCanonical_CCWireName(t *testing.T) {
	t.Parallel()
	result := WireToCanonical("PreToolUse")
	if result != "pre_tool" {
		t.Errorf("got %q, want %q", result, "pre_tool")
	}
}

func TestWireToCanonical_GeminiWireName(t *testing.T) {
	t.Parallel()
	result := WireToCanonical("BeforeTool")
	if result != "pre_tool" {
		t.Errorf("got %q, want %q", result, "pre_tool")
	}
}

func TestWireToCanonical_CanonicalPassthrough(t *testing.T) {
	t.Parallel()
	// Canonical names are not in wireToCanonical -- pass through unchanged
	result := WireToCanonical("pre_tool")
	if result != "pre_tool" {
		t.Errorf("got %q, want %q", result, "pre_tool")
	}
}

func TestWireToCanonical_UnknownPassthrough(t *testing.T) {
	t.Parallel()
	result := WireToCanonical("UnknownEvent")
	if result != "UnknownEvent" {
		t.Errorf("got %q, want %q", result, "UnknownEvent")
	}
}

func TestTranslateInboundEvent_GeminiWireName(t *testing.T) {
	t.Parallel()
	result := TranslateInboundEvent("BeforeTool")
	if result != "pre_tool" {
		t.Errorf("got %q, want %q", result, "pre_tool")
	}
}

func TestTranslateInboundEvent_CCWireName(t *testing.T) {
	t.Parallel()
	result := TranslateInboundEvent("PostToolUse")
	if result != "post_tool" {
		t.Errorf("got %q, want %q", result, "post_tool")
	}
}

func TestTranslateInboundEvent_UnknownEventPassthrough(t *testing.T) {
	t.Parallel()
	result := TranslateInboundEvent("UnknownEvent")
	if result != "UnknownEvent" {
		t.Errorf("got %q, want %q", result, "UnknownEvent")
	}
}

func TestTranslateMatcherForChannel_GeminiTwoTools(t *testing.T) {
	t.Parallel()
	result := TranslateMatcherForChannel("Edit|Write", "gemini")
	if result != "replace|write_file" {
		t.Errorf("got %q, want %q", result, "replace|write_file")
	}
}

func TestTranslateMatcherForChannel_GeminiSingleTool(t *testing.T) {
	t.Parallel()
	result := TranslateMatcherForChannel("Bash", "gemini")
	if result != "run_shell_command" {
		t.Errorf("got %q, want %q", result, "run_shell_command")
	}
}

func TestTranslateMatcherForChannel_GeminiThreeTools(t *testing.T) {
	t.Parallel()
	result := TranslateMatcherForChannel("Edit|Write|Bash", "gemini")
	if result != "replace|write_file|run_shell_command" {
		t.Errorf("got %q, want %q", result, "replace|write_file|run_shell_command")
	}
}

func TestTranslateMatcherForChannel_ClaudePassthrough(t *testing.T) {
	t.Parallel()
	result := TranslateMatcherForChannel("Edit|Write", "claude")
	if result != "Edit|Write" {
		t.Errorf("got %q, want %q", result, "Edit|Write")
	}
}

func TestTranslateMatcherForChannel_UnknownToolPassthrough(t *testing.T) {
	t.Parallel()
	result := TranslateMatcherForChannel("CustomTool", "gemini")
	if result != "CustomTool" {
		t.Errorf("got %q, want %q", result, "CustomTool")
	}
}

func TestCanonicalToWire_AllBidirectionalEvents(t *testing.T) {
	t.Parallel()
	// Table of all bidirectional events per ADR-0032
	tests := []struct {
		canonical string
		ccWire    string
		gemWire   string
	}{
		{"pre_tool", "PreToolUse", "BeforeTool"},
		{"post_tool", "PostToolUse", "AfterTool"},
		{"session_start", "SessionStart", "SessionStart"},
		{"session_end", "SessionEnd", "SessionEnd"},
		{"pre_prompt", "UserPromptSubmit", "BeforeAgent"},
		{"pre_compact", "PreCompact", "PreCompress"},
		{"notification", "Notification", "Notification"},
	}
	for _, tt := range tests {
		t.Run(tt.canonical, func(t *testing.T) {
			t.Parallel()
			ccWire, skip := CanonicalToWire(tt.canonical, "claude")
			if skip {
				t.Errorf("canonical %q should not be skipped for claude", tt.canonical)
			}
			if ccWire != tt.ccWire {
				t.Errorf("CC wire = %q, want %q", ccWire, tt.ccWire)
			}

			gemWire, skip := CanonicalToWire(tt.canonical, "gemini")
			if skip {
				t.Errorf("canonical %q should not be skipped for gemini", tt.canonical)
			}
			if gemWire != tt.gemWire {
				t.Errorf("Gemini wire = %q, want %q", gemWire, tt.gemWire)
			}
		})
	}
}

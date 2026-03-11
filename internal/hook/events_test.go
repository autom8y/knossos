package hook

import "testing"

func TestTranslateEventForChannel_ClaudePassthrough(t *testing.T) {
	t.Parallel()
	event, skip := TranslateEventForChannel("PreToolUse", "claude")
	if skip {
		t.Error("claude channel should never skip events")
	}
	if event != "PreToolUse" {
		t.Errorf("got %q, want %q", event, "PreToolUse")
	}
}

func TestTranslateEventForChannel_GeminiTranslated(t *testing.T) {
	t.Parallel()
	event, skip := TranslateEventForChannel("PreToolUse", "gemini")
	if skip {
		t.Error("PreToolUse is translatable, should not be skipped")
	}
	if event != "BeforeTool" {
		t.Errorf("got %q, want %q", event, "BeforeTool")
	}
}

func TestTranslateEventForChannel_GeminiSkipsUnmappable(t *testing.T) {
	t.Parallel()
	event, skip := TranslateEventForChannel("Stop", "gemini")
	if !skip {
		t.Error("Stop has no Gemini equivalent, should be skipped")
	}
	if event != "" {
		t.Errorf("skipped event should be empty string, got %q", event)
	}
}

func TestTranslateEventForChannel_GeminiIdentityMapping(t *testing.T) {
	t.Parallel()
	event, skip := TranslateEventForChannel("SessionStart", "gemini")
	if skip {
		t.Error("SessionStart is translatable (identity), should not be skipped")
	}
	if event != "SessionStart" {
		t.Errorf("got %q, want %q", event, "SessionStart")
	}
}

func TestTranslateInboundEvent_GeminiWireName(t *testing.T) {
	t.Parallel()
	result := TranslateInboundEvent("BeforeTool")
	if result != "PreToolUse" {
		t.Errorf("got %q, want %q", result, "PreToolUse")
	}
}

func TestTranslateInboundEvent_CCCanonicalPassthrough(t *testing.T) {
	t.Parallel()
	// CC canonical names are not Gemini wire names -- pass through unchanged
	result := TranslateInboundEvent("PreToolUse")
	if result != "PreToolUse" {
		t.Errorf("got %q, want %q", result, "PreToolUse")
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

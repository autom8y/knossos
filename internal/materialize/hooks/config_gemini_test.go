package hooks

import (
	"encoding/json"
	"testing"
)

// geminiTestConfig returns a representative HooksConfig covering all skip/translate scenarios.
// Event names use knossos canonical vocabulary (ADR-0032).
func geminiTestConfig() *HooksConfig {
	return &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			// Translatable: pre_tool -> BeforeTool
			{Event: "pre_tool", Matcher: "Edit|Write", Command: "ari hook writeguard --output json", Timeout: 3, Priority: 3},
			{Event: "pre_tool", Matcher: "Bash", Command: "ari hook validate --output json", Timeout: 5, Priority: 5},
			// Translatable: post_tool -> AfterTool
			{Event: "post_tool", Matcher: "Edit|Write|Bash", Command: "ari hook clew --output json", Timeout: 5, Async: true, Priority: 5},
			{Event: "post_tool", Command: "ari hook budget --output json", Timeout: 3, Priority: 90},
			// Translatable: session_start -> SessionStart (identity)
			{Event: "session_start", Command: "ari hook context --output json", Timeout: 10, Priority: 5},
			// Translatable: session_end -> SessionEnd (identity)
			{Event: "session_end", Command: "ari hook sessionend --output json", Timeout: 5, Priority: 5},
			// Translatable: pre_compact -> PreCompress
			{Event: "pre_compact", Command: "ari hook precompact --output json", Timeout: 5, Priority: 5},
			// No Gemini equivalent -- must be skipped
			{Event: "stop", Command: "ari hook autopark --output json", Timeout: 5, Priority: 5},
			{Event: "worktree_create", Command: "ari hook worktree-seed --output json", Timeout: 30, Priority: 5},
			{Event: "worktree_remove", Command: "ari hook worktree-remove --output json", Timeout: 10, Priority: 5},
			{Event: "subagent_start", Command: "ari hook subagent-start --output json", Timeout: 5, Priority: 5},
			{Event: "subagent_stop", Command: "ari hook subagent-stop --output json", Timeout: 5, Priority: 5},
		},
	}
}

// TestBuildHooksSettings_GeminiEventTranslation verifies that canonical event names are
// translated to Gemini wire names when channel == "gemini".
func TestBuildHooksSettings_GeminiEventTranslation(t *testing.T) {
	t.Parallel()

	hooks := BuildHooksSettings(geminiTestConfig(), "gemini")

	// Events that should be present with Gemini wire names
	expectedPresent := []string{"BeforeTool", "AfterTool", "SessionStart", "SessionEnd", "PreCompress"}
	for _, key := range expectedPresent {
		if hooks[key] == nil {
			t.Errorf("expected Gemini event key %q in output, but it is absent", key)
		}
	}

	// Canonical names must NOT appear as keys (they were translated)
	canonicalNames := []string{"pre_tool", "post_tool", "pre_compact", "session_start", "session_end"}
	for _, key := range canonicalNames {
		if hooks[key] != nil {
			t.Errorf("canonical key %q should not appear in Gemini output", key)
		}
	}
}

// TestBuildHooksSettings_GeminiSkipsUnmappable verifies that events with no Gemini
// equivalent are silently omitted from the output.
func TestBuildHooksSettings_GeminiSkipsUnmappable(t *testing.T) {
	t.Parallel()

	hooks := BuildHooksSettings(geminiTestConfig(), "gemini")

	// These canonical events have no Gemini wire equivalent
	skippedEvents := []string{"stop", "worktree_create", "worktree_remove", "subagent_start", "subagent_stop",
		// Also verify the CC wire names are not present
		"Stop", "WorktreeCreate", "WorktreeRemove", "SubagentStart", "SubagentStop"}
	for _, event := range skippedEvents {
		if hooks[event] != nil {
			t.Errorf("event %q has no Gemini equivalent and should be absent, but it is present", event)
		}
	}
}

// TestBuildHooksSettings_GeminiToolTranslation verifies that matcher patterns
// with CC tool names are rewritten to Gemini tool names.
func TestBuildHooksSettings_GeminiToolTranslation(t *testing.T) {
	t.Parallel()

	hooks := BuildHooksSettings(geminiTestConfig(), "gemini")

	beforeTool, ok := hooks["BeforeTool"].([]map[string]any)
	if !ok {
		t.Fatalf("BeforeTool is not []map[string]any")
	}

	// First entry: Edit|Write -> replace|write_file
	if len(beforeTool) < 1 {
		t.Fatalf("expected at least 1 entry in BeforeTool, got %d", len(beforeTool))
	}
	writeguardMatcher, _ := beforeTool[0]["matcher"].(string)
	if writeguardMatcher != "replace|write_file" {
		t.Errorf("writeguard matcher = %q, want %q", writeguardMatcher, "replace|write_file")
	}

	// Second entry: Bash -> run_shell_command
	if len(beforeTool) < 2 {
		t.Fatalf("expected at least 2 entries in BeforeTool, got %d", len(beforeTool))
	}
	validateMatcher, _ := beforeTool[1]["matcher"].(string)
	if validateMatcher != "run_shell_command" {
		t.Errorf("validate matcher = %q, want %q", validateMatcher, "run_shell_command")
	}

	// AfterTool: Edit|Write|Bash -> replace|write_file|run_shell_command
	afterTool, ok := hooks["AfterTool"].([]map[string]any)
	if !ok {
		t.Fatalf("AfterTool is not []map[string]any")
	}
	if len(afterTool) < 1 {
		t.Fatalf("expected at least 1 entry in AfterTool, got %d", len(afterTool))
	}
	clewMatcher, _ := afterTool[0]["matcher"].(string)
	if clewMatcher != "replace|write_file|run_shell_command" {
		t.Errorf("clew matcher = %q, want %q", clewMatcher, "replace|write_file|run_shell_command")
	}
}

// TestBuildHooksSettings_GeminiPreservesEnv verifies that the KNOSSOS_CHANNEL env
// var is injected into every hook handler for the gemini channel.
func TestBuildHooksSettings_GeminiPreservesEnv(t *testing.T) {
	t.Parallel()

	hooks := BuildHooksSettings(geminiTestConfig(), "gemini")

	// Check SessionStart hooks carry the env var
	sessionStart, ok := hooks["SessionStart"].([]map[string]any)
	if !ok {
		t.Fatalf("SessionStart is not []map[string]any")
	}
	if len(sessionStart) == 0 {
		t.Fatal("SessionStart has no entries")
	}

	hooksArr, ok := sessionStart[0]["hooks"].([]map[string]any)
	if !ok || len(hooksArr) == 0 {
		t.Fatal("SessionStart[0].hooks is empty")
	}

	envMap, ok := hooksArr[0]["env"].(map[string]string)
	if !ok {
		t.Fatalf("env is not map[string]string, got %T", hooksArr[0]["env"])
	}
	if envMap["KNOSSOS_CHANNEL"] != "gemini" {
		t.Errorf("KNOSSOS_CHANNEL = %q, want gemini", envMap["KNOSSOS_CHANNEL"])
	}
}

// TestBuildHooksSettings_ClaudeTranslation verifies that canonical event names are
// translated to CC wire names when channel == "claude".
func TestBuildHooksSettings_ClaudeTranslation(t *testing.T) {
	t.Parallel()

	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "pre_tool", Matcher: "Edit|Write", Command: "ari hook writeguard --output json", Timeout: 3, Priority: 3},
			{Event: "stop", Command: "ari hook autopark --output json", Timeout: 5, Priority: 5},
			{Event: "post_tool", Matcher: "Edit|Write|Bash", Command: "ari hook clew --output json", Timeout: 5, Async: true, Priority: 5},
		},
	}

	hooks := BuildHooksSettings(cfg, "claude")

	// CC wire event names must be present
	if hooks["PreToolUse"] == nil {
		t.Error("claude output should have PreToolUse key")
	}
	if hooks["Stop"] == nil {
		t.Error("claude output should have Stop key")
	}
	if hooks["PostToolUse"] == nil {
		t.Error("claude output should have PostToolUse key")
	}

	// Canonical names must NOT appear as keys
	if hooks["pre_tool"] != nil {
		t.Error("canonical name pre_tool should not appear in claude output")
	}

	// Matcher must be CC tool names unchanged
	preToolUse, ok := hooks["PreToolUse"].([]map[string]any)
	if !ok || len(preToolUse) == 0 {
		t.Fatalf("PreToolUse is not []map[string]any or empty")
	}
	matcher, _ := preToolUse[0]["matcher"].(string)
	if matcher != "Edit|Write" {
		t.Errorf("claude matcher = %q, want %q", matcher, "Edit|Write")
	}

	// No KNOSSOS_CHANNEL env var for claude channel
	hooksArr, ok := preToolUse[0]["hooks"].([]map[string]any)
	if !ok || len(hooksArr) == 0 {
		t.Fatal("PreToolUse[0].hooks is empty")
	}
	if _, hasEnv := hooksArr[0]["env"]; hasEnv {
		t.Error("claude channel hooks should not have KNOSSOS_CHANNEL env var")
	}
}

// TestMergeHooksSettings_GeminiIdempotent verifies that merging twice with
// channel=="gemini" produces identical output (idempotency invariant).
func TestMergeHooksSettings_GeminiIdempotent(t *testing.T) {
	t.Parallel()

	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "pre_tool", Matcher: "Edit|Write", Command: "ari hook writeguard --output json", Priority: 3},
			{Event: "post_tool", Command: "ari hook budget --output json", Priority: 90},
			{Event: "session_start", Command: "ari hook context --output json", Priority: 5},
			// This event should be skipped for gemini
			{Event: "stop", Command: "ari hook autopark --output json", Priority: 5},
		},
	}

	settings := make(map[string]any)
	result1 := MergeHooksSettings(settings, cfg, "gemini")

	// Serialize to JSON and back (simulates load/save cycle)
	data, err := json.Marshal(result1)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	var settings2 map[string]any
	if err := json.Unmarshal(data, &settings2); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	result2 := MergeHooksSettings(settings2, cfg, "gemini")

	data1, _ := json.MarshalIndent(result1, "", "  ")
	data2, _ := json.MarshalIndent(result2, "", "  ")

	if string(data1) != string(data2) {
		t.Errorf("Gemini merge is not idempotent.\nFirst:\n%s\nSecond:\n%s", data1, data2)
	}
}

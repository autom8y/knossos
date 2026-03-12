package hook

import (
	"strings"

	"github.com/autom8y/knossos/internal/channel"
)

// canonicalToWire maps knossos canonical event names to channel wire names.
// Each canonical event maps to a per-channel wire name.
// Absent channel keys mean the channel doesn't support that event.
var canonicalToWire = map[string]map[string]string{
	"pre_tool":           {"claude": "PreToolUse", "gemini": "BeforeTool"},
	"post_tool":          {"claude": "PostToolUse", "gemini": "AfterTool"},
	"post_tool_failure":  {"claude": "PostToolUseFailure"},
	"permission_request": {"claude": "PermissionRequest"},
	"stop":               {"claude": "Stop"},
	"session_start":      {"claude": "SessionStart", "gemini": "SessionStart"},
	"session_end":        {"claude": "SessionEnd", "gemini": "SessionEnd"},
	"pre_prompt":         {"claude": "UserPromptSubmit", "gemini": "BeforeAgent"},
	"pre_compact":        {"claude": "PreCompact", "gemini": "PreCompress"},
	"subagent_start":     {"claude": "SubagentStart"},
	"subagent_stop":      {"claude": "SubagentStop"},
	"notification":       {"claude": "Notification", "gemini": "Notification"},
	"teammate_idle":      {"claude": "TeammateIdle"},
	"task_completed":     {"claude": "TaskCompleted"},
	"pre_model":          {"gemini": "BeforeModel"},
	"post_model":         {"gemini": "AfterModel"},
	// CC-specific lifecycle events (not yet in ADR-0032 canonical set)
	"worktree_create": {"claude": "WorktreeCreate"},
	"worktree_remove": {"claude": "WorktreeRemove"},
}

// wireToCanonical is the reverse mapping: wire event names (from any channel) -> canonical.
// Computed from canonicalToWire at init time.
var wireToCanonical map[string]string

func init() {
	wireToCanonical = make(map[string]string)
	for canonical, wires := range canonicalToWire {
		for _, wireName := range wires {
			wireToCanonical[wireName] = canonical
		}
	}
}

// CanonicalToWire returns the wire event name for a given channel.
// Returns ("", true) if the event has no wire equivalent for that channel (skip it).
// Returns (wireName, false) on success.
// Returns (canonical, false) if the canonical name is unknown (passthrough).
func CanonicalToWire(canonical string, channel string) (string, bool) {
	wires, ok := canonicalToWire[canonical]
	if !ok {
		return canonical, false // unknown canonical -- pass through
	}
	wire, hasWire := wires[channel]
	if !hasWire {
		return "", true // no equivalent for this channel -- skip
	}
	return wire, false
}

// WireToCanonical converts any channel's wire event name to knossos canonical.
// Returns the input unchanged if the wire name is not recognized.
func WireToCanonical(wireEvent string) string {
	if canonical, ok := wireToCanonical[wireEvent]; ok {
		return canonical
	}
	return wireEvent // pass through unknown
}

// TranslateEventForChannel returns the channel-appropriate wire event name.
// For "claude", translates canonical to CC wire names.
// For "gemini", translates canonical to Gemini wire names.
// Returns ("", true) if the event has no equivalent for the target channel (skip it).
func TranslateEventForChannel(canonicalEvent, targetChannel string) (string, bool) {
	return CanonicalToWire(canonicalEvent, targetChannel)
}

// TranslateInboundEvent converts a wire event name (from any channel) to canonical.
// Returns the input unchanged if the wire name is not recognized.
func TranslateInboundEvent(wireEvent string) string {
	return WireToCanonical(wireEvent)
}

// TranslateMatcherForChannel rewrites a pipe-delimited matcher pattern
// for the target channel.
// For "claude", returns the matcher unchanged.
// For "gemini", translates each pipe-delimited tool name segment using the
// channel.CanonicalTool mapping. Unknown tool names pass through unchanged.
func TranslateMatcherForChannel(matcher, targetChannel string) string {
	if targetChannel != "gemini" || matcher == "" {
		return matcher
	}

	segments := strings.Split(matcher, "|")
	translated := make([]string, len(segments))
	for i, seg := range segments {
		// Use ccWireToGeminiWire (not TranslateTool) so that CC-only tools appearing
		// in hook matchers pass through unchanged rather than being dropped.
		// A matcher for a nonexistent Gemini tool is harmless — the hook never fires.
		geminiTool, _ := channel.CCWireToGeminiWire(seg)
		translated[i] = geminiTool
	}
	return strings.Join(translated, "|")
}

package hook

import (
	"strings"

	"github.com/autom8y/knossos/internal/channel"
)

// ccToGeminiEvent maps CC-canonical event names to Gemini wire names.
// Events not in this map have no Gemini equivalent and must be skipped.
var ccToGeminiEvent = map[string]string{
	"PreToolUse":       "BeforeTool",
	"PostToolUse":      "AfterTool",
	"SessionStart":     "SessionStart",
	"SessionEnd":       "SessionEnd",
	"UserPromptSubmit": "BeforeAgent",
	"PreCompact":       "PreCompress",
	"Notification":     "Notification",
}

// geminiToCCEvent is the reverse mapping: Gemini wire names -> CC canonical.
// Computed from ccToGeminiEvent at init time.
var geminiToCCEvent map[string]string

func init() {
	geminiToCCEvent = make(map[string]string, len(ccToGeminiEvent))
	for cc, gemini := range ccToGeminiEvent {
		geminiToCCEvent[gemini] = cc
	}
}

// Tool name mapping is provided by internal/channel.CCToGeminiTool (single source of truth).
// Both hook wire-protocol names (ReadFiles) and agent frontmatter names (Read) are covered.

// TranslateEventForChannel returns the channel-appropriate event name.
// For "claude", returns (ccEvent, false) -- passthrough, no translation needed.
// For "gemini", returns (geminiEvent, false) if the event has a Gemini equivalent,
// or ("", true) if the event has no Gemini equivalent (caller should skip it).
func TranslateEventForChannel(ccEvent, targetChannel string) (string, bool) {
	if targetChannel != "gemini" {
		return ccEvent, false
	}
	geminiEvent, ok := ccToGeminiEvent[ccEvent]
	if !ok {
		// No Gemini equivalent -- caller must skip this event
		return "", true
	}
	return geminiEvent, false
}

// TranslateInboundEvent converts a Gemini wire event name to CC canonical.
// Returns the input unchanged if the event is not a known Gemini wire name,
// which handles CC-canonical passthrough and unknown future events gracefully.
func TranslateInboundEvent(wireEvent string) string {
	if cc, ok := geminiToCCEvent[wireEvent]; ok {
		return cc
	}
	// Not a known Gemini wire name -- return as-is (may already be CC canonical
	// or may be an unknown Gemini-only event like BeforeModel).
	return wireEvent
}

// TranslateMatcherForChannel rewrites a pipe-delimited matcher pattern
// for the target channel.
// For "claude", returns the matcher unchanged.
// For "gemini", translates each pipe-delimited tool name segment using the
// channel.CCToGeminiTool mapping. Unknown tool names pass through unchanged.
func TranslateMatcherForChannel(matcher, targetChannel string) string {
	if targetChannel != "gemini" || matcher == "" {
		return matcher
	}

	segments := strings.Split(matcher, "|")
	translated := make([]string, len(segments))
	for i, seg := range segments {
		// Use the shared channel map directly (not TranslateTool) so that CC-only
		// tools appearing in hook matchers pass through unchanged rather than being
		// dropped. A matcher for a nonexistent tool is harmless — the hook never fires.
		if geminiTool, ok := channel.CCToGeminiTool[seg]; ok {
			translated[i] = geminiTool
		} else {
			// Unknown tool name passes through unchanged (defensive)
			translated[i] = seg
		}
	}
	return strings.Join(translated, "|")
}

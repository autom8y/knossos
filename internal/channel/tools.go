package channel

// CanonicalTool maps knossos canonical tool names to per-channel wire names.
// This is the harness-agnostic vocabulary: templates and platform code reference
// canonical names; channel compilers resolve to wire names at projection time.
var CanonicalTool = map[string]map[string]string{
	"read_file":      {"claude": "Read", "gemini": "read_file"},
	"edit_file":      {"claude": "Edit", "gemini": "replace"},
	"write_file":     {"claude": "Write", "gemini": "write_file"},
	"list_files":     {"claude": "Glob", "gemini": "glob"},
	"search_content": {"claude": "Grep", "gemini": "grep_search"},
	"run_shell":      {"claude": "Bash", "gemini": "run_shell_command"},
	"web_search":     {"claude": "WebSearch", "gemini": "google_web_search"},
	"web_fetch":      {"claude": "WebFetch", "gemini": "web_fetch"},
	"write_todos":    {"claude": "TodoWrite", "gemini": "write_todos"},
	"activate_skill": {"claude": "Skill", "gemini": "activate_skill"},
	"delegate":       {"claude": "Task"},
}

// wireToCanonical is the reverse index: wire name -> canonical name.
// Computed once at init from CanonicalTool.
var wireToCanonical map[string]string

func init() {
	wireToCanonical = make(map[string]string, len(CanonicalTool)*2)
	for canonical, wires := range CanonicalTool {
		for _, wire := range wires {
			wireToCanonical[wire] = canonical
		}
	}
}

// CanonicalToWireTool returns the wire tool name for a given channel.
//
// Return semantics:
//   - (wireName, true): resolved successfully.
//   - (canonical, true): unknown canonical name -- passed through as-is.
//   - ("", false): canonical is known but has no equivalent for this channel.
func CanonicalToWireTool(canonical, channel string) (string, bool) {
	wires, ok := CanonicalTool[canonical]
	if !ok {
		return canonical, true // unknown canonical -- pass through
	}
	wire, hasWire := wires[channel]
	if !hasWire {
		return "", false // no equivalent for this channel
	}
	return wire, true
}

// WireToCanonicalTool converts a channel-specific wire tool name to its
// canonical knossos name. Returns (canonical, true) on hit, or
// (wireName, false) if the wire name is not in any canonical mapping.
func WireToCanonicalTool(wireName string) (string, bool) {
	if canonical, ok := wireToCanonical[wireName]; ok {
		return canonical, true
	}
	return wireName, false
}

// CCWireToGeminiWire translates a CC wire tool name to its Gemini wire equivalent.
// Unlike TranslateTool, this function does NOT drop CC-only tools — they pass through
// unchanged. This is the correct semantics for hook matchers, where a matcher for a
// nonexistent Gemini tool is harmless (the hook never fires).
//
// Return semantics:
//   - (geminiName, true): translated successfully.
//   - (ccTool, false): no Gemini equivalent found — caller receives the original name.
func CCWireToGeminiWire(ccTool string) (string, bool) {
	for _, wires := range CanonicalTool {
		if wires["claude"] == ccTool {
			if gemini, ok := wires["gemini"]; ok {
				return gemini, true
			}
			// CC-only canonical entry — no Gemini equivalent; pass through.
			return ccTool, false
		}
	}
	// Wire aliases not in canonical map (e.g., "ReadFiles" -> "read_file").
	// Fall through: look for a direct gemini entry matching the CC name.
	// Unknown tool passes through unchanged.
	return ccTool, false
}

// TranslateTool returns the Gemini equivalent for a CC tool name.
//
// Return semantics:
//   - ("", false): the tool is CC-only — caller should drop it entirely.
//   - (geminiName, true): translated successfully.
//   - (ccTool, true): unknown tool — passes through unchanged (forward compat).
func TranslateTool(ccTool string) (string, bool) {
	for _, wires := range CanonicalTool {
		if wires["claude"] == ccTool {
			gemini, hasGemini := wires["gemini"]
			if !hasGemini {
				// Exists in canonical map with a claude entry but no gemini entry:
				// this is a claude-only tool — drop it.
				return "", false
			}
			return gemini, true
		}
	}
	// Wire aliases not in canonical map (e.g., "ReadFiles").
	// Check wireToCanonical to resolve aliases to their canonical entry.
	if canonical, ok := wireToCanonical[ccTool]; ok {
		// Resolved to canonical: look up the gemini wire name.
		wires := CanonicalTool[canonical]
		if gemini, ok := wires["gemini"]; ok {
			return gemini, true
		}
		return "", false
	}
	// Unknown tool passes through unchanged (defensive forward compatibility).
	return ccTool, true
}

// TranslateFrontmatterTools translates a slice of CC tool names to Gemini equivalents.
// CC-only tools are silently dropped. Unknown tools pass through unchanged.
// Returns nil if the input is nil; returns an empty slice if all tools are dropped.
func TranslateFrontmatterTools(ccTools []string) []string {
	if ccTools == nil {
		return nil
	}
	result := make([]string, 0, len(ccTools))
	for _, tool := range ccTools {
		if gemini, ok := TranslateTool(tool); ok {
			result = append(result, gemini)
		}
	}
	return result
}

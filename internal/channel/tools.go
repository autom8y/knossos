package channel

// CCToGeminiTool maps CC-canonical tool names to Gemini equivalents.
// Both agent-frontmatter names (Read) and hook wire-protocol names (ReadFiles)
// are included so this is the single source of truth per ADR-0031.
//
// Source of truth for Gemini tool names:
//
//	google-gemini/gemini-cli → packages/core/src/tools/definitions/base-declarations.ts
var CCToGeminiTool = map[string]string{
	// File system tools
	"Read":      "read_file",          // agent frontmatter name
	"ReadFiles": "read_file",          // hook wire-protocol name
	"Edit":      "replace",            // EDIT_TOOL_NAME
	"Write":     "write_file",         // WRITE_FILE_TOOL_NAME
	"Glob":      "glob",              // GLOB_TOOL_NAME
	"Grep":      "grep_search",       // GREP_TOOL_NAME
	// Shell
	"Bash": "run_shell_command",       // SHELL_TOOL_NAME
	// Web
	"WebSearch": "google_web_search",  // WEB_SEARCH_TOOL_NAME
	"WebFetch":  "web_fetch",          // WEB_FETCH_TOOL_NAME
	// Task management
	"TodoWrite": "write_todos",        // WRITE_TODOS_TOOL_NAME
	// Skills
	"Skill": "activate_skill",         // ACTIVATE_SKILL_TOOL_NAME
}

// CCOnlyTools lists CC tools that have no Gemini equivalent.
// These are silently dropped from Gemini agent frontmatter (tools and disallowedTools).
var CCOnlyTools = map[string]bool{
	"Task":         true, // Gemini uses implicit description-based agent routing
	"NotebookEdit": true, // No Gemini equivalent
}

// CanonicalTool maps knossos canonical tool names to per-channel wire names.
// This is the harness-agnostic vocabulary: templates and platform code reference
// canonical names; channel compilers resolve to wire names at projection time.
//
// The existing CCToGeminiTool map is retained for backward compatibility.
// New code should prefer CanonicalToWireTool / WireToCanonicalTool.
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

// TranslateTool returns the Gemini equivalent for a CC tool name.
//
// Return semantics:
//   - ("", false): the tool is CC-only — caller should drop it entirely.
//   - (geminiName, true): translated successfully.
//   - (ccTool, true): unknown tool — passes through unchanged (forward compat).
func TranslateTool(ccTool string) (string, bool) {
	if CCOnlyTools[ccTool] {
		return "", false
	}
	if gemini, ok := CCToGeminiTool[ccTool]; ok {
		return gemini, true
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

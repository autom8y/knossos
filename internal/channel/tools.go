package channel

// CCToGeminiTool maps CC-canonical tool names to Gemini equivalents.
// Both agent-frontmatter names (Read) and hook wire-protocol names (ReadFiles)
// are included so this is the single source of truth per ADR-0031.
// Neither key set includes CC-only tools (Task, TodoWrite, etc.) — those are
// in CCOnlyTools and have no Gemini equivalent.
var CCToGeminiTool = map[string]string{
	"Read":      "read_file",       // agent frontmatter name
	"ReadFiles": "read_file",       // hook wire-protocol name
	"Edit":      "replace",
	"Write":     "write_file",
	"Bash":      "run_shell_command",
	"Glob":      "glob",
	"Grep":      "search_files",
}

// CCOnlyTools lists CC tools that have no Gemini equivalent.
// These are silently dropped from Gemini agent frontmatter (tools and disallowedTools).
// Dropping is correct: you cannot allow or disallow a tool that does not exist.
var CCOnlyTools = map[string]bool{
	"Task":         true,
	"TodoWrite":    true,
	"Skill":        true,
	"WebSearch":    true,
	"WebFetch":     true,
	"NotebookEdit": true,
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

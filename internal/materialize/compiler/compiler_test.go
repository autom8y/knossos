package compiler_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/materialize/compiler"
	"github.com/pelletier/go-toml/v2"
)

func TestClaudeCompiler_IdentityTransform(t *testing.T) {
	t.Parallel()
	c := &compiler.ClaudeCompiler{}

	// Test Command
	name, content, err := c.CompileCommand("test-cmd", "desc", "<arg>", "# body\ncontent")
	if err != nil {
		t.Fatal(err)
	}
	if name != "test-cmd.md" {
		t.Errorf("expected test-cmd.md, got %s", name)
	}
	if string(content) != "# body\ncontent" {
		t.Errorf("unexpected content: %s", string(content))
	}

	// Test Skill
	dir, filename, content, err := c.CompileSkill("test-skill", "desc", "# body\ncontent")
	if err != nil {
		t.Fatal(err)
	}
	if dir != "test-skill" || filename != "SKILL.md" {
		t.Errorf("unexpected dir/file: %s/%s", dir, filename)
	}
	if string(content) != "# body\ncontent" {
		t.Errorf("unexpected content: %s", string(content))
	}

	// Test Agent
	fm := map[string]any{"name": "test-agent", "role": "tester"}
	agentContent, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	
	if !bytes.Contains(agentContent, []byte("name: test-agent")) || !bytes.Contains(agentContent, []byte("# body")) {
		t.Errorf("unexpected agent content: %s", string(agentContent))
	}
}

func TestGeminiCompiler_CommandToTOML(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	name, content, err := c.CompileCommand("test-cmd", "A test", "<arg>", "# body")
	if err != nil {
		t.Fatal(err)
	}

	if name != "test-cmd.toml" {
		t.Errorf("expected test-cmd.toml, got %s", name)
	}

	var tomlData map[string]any
	if err := toml.Unmarshal(content, &tomlData); err != nil {
		t.Fatal(err)
	}

	if tomlData["name"] != "test-cmd" {
		t.Errorf("unexpected name in toml")
	}
	if tomlData["description"] != "A test" {
		t.Errorf("unexpected desc in toml")
	}
	
	prompt := tomlData["prompt"].(string)
	if !strings.Contains(prompt, "# body") {
		t.Errorf("missing body in prompt: %s", prompt)
	}
	if !strings.Contains(prompt, "User arguments: {{args}}") {
		t.Errorf("missing arg hint in prompt: %s", prompt)
	}
}

func TestGeminiCompiler_SkillFormat(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	dir, filename, content, err := c.CompileSkill("test-skill", "A test", "# body")
	if err != nil {
		t.Fatal(err)
	}

	if dir != "test-skill" || filename != "SKILL.md" {
		t.Errorf("unexpected dir/file")
	}

	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		t.Errorf("expected frontmatter")
	}
	if !strings.Contains(contentStr, "name: test-skill") || !strings.Contains(contentStr, "version: \"1.0\"") {
		t.Errorf("missing fields in frontmatter: %s", contentStr)
	}
	if !strings.Contains(contentStr, "# body") {
		t.Errorf("missing body")
	}
}

func TestGeminiCompiler_AgentFormat(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	fm := map[string]any{"name": "test-agent", "role": "tester"}
	agentContent, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(agentContent, []byte("name: test-agent")) || !bytes.Contains(agentContent, []byte("# body")) {
		t.Errorf("unexpected agent content: %s", string(agentContent))
	}
}

// --- Gemini agent tool translation tests (Layer 1 of agent parity fix) ---

func TestGeminiCompiler_CompileAgent_ToolTranslation(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	fm := map[string]any{
		"name":  "test-agent",
		"tools": []any{"Read", "Bash", "Edit", "Write", "Glob", "Grep"},
	}
	content, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	// Verify Gemini tool names are present
	for _, gemini := range []string{"read_file", "run_shell_command", "replace", "write_file", "glob", "grep_search"} {
		if !strings.Contains(s, gemini) {
			t.Errorf("expected Gemini tool %q in output:\n%s", gemini, s)
		}
	}

	// Verify CC tool names are NOT present in the tools list
	for _, cc := range []string{"- Read\n", "- Bash\n", "- Edit\n", "- Write\n", "- Glob\n", "- Grep\n"} {
		if strings.Contains(s, cc) {
			t.Errorf("unexpected CC tool %q still in output:\n%s", cc, s)
		}
	}
}

func TestGeminiCompiler_CompileAgent_CCOnlyToolsDropped(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	fm := map[string]any{
		"name":            "potnia",
		"tools":           []any{"Read", "Skill", "Task"},
		"disallowedTools": []any{"Bash", "Write", "Edit", "Glob", "Grep", "Task"},
	}
	content, err := c.CompileAgent("potnia", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	// tools: Read->read_file, Skill->activate_skill, Task dropped
	if !strings.Contains(s, "read_file") {
		t.Errorf("expected read_file in tools: %s", s)
	}
	if !strings.Contains(s, "activate_skill") {
		t.Errorf("expected activate_skill in tools (Skill is now mapped): %s", s)
	}
	// Task is the only truly CC-only tool in this list
	if strings.Contains(s, "- Task") {
		t.Errorf("CC-only tool Task still present in output:\n%s", s)
	}

	// disallowedTools is stripped entirely — Gemini CLI doesn't accept it as a key.
	if strings.Contains(s, "disallowedTools") {
		t.Errorf("disallowedTools should be stripped for Gemini (unrecognized key): %s", s)
	}
}

func TestGeminiCompiler_CompileAgent_UnknownToolPassthrough(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	fm := map[string]any{
		"name":  "test-agent",
		"tools": []any{"Read", "CustomTool"},
	}
	content, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	if !strings.Contains(s, "read_file") {
		t.Errorf("expected read_file in output: %s", s)
	}
	if !strings.Contains(s, "CustomTool") {
		t.Errorf("expected CustomTool to pass through: %s", s)
	}
}

func TestGeminiCompiler_CompileAgent_EmptyToolsField(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	fm := map[string]any{
		"name": "test-agent",
		// No tools field
	}
	content, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	// Should not have a tools key since none was specified
	if strings.Contains(s, "tools:") {
		t.Errorf("unexpected tools field in output: %s", s)
	}
}

func TestGeminiCompiler_CompileAgent_AllCCOnlyToolsDropped(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	// Only Task and NotebookEdit are truly CC-only now
	fm := map[string]any{
		"name":            "test-agent",
		"disallowedTools": []any{"Task", "NotebookEdit"},
	}
	content, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	// disallowedTools is stripped by Gemini key whitelist regardless
	if strings.Contains(s, "disallowedTools:") {
		t.Errorf("expected disallowedTools to be absent (stripped by Gemini key whitelist): %s", s)
	}
}

func TestGeminiCompiler_CompileAgent_SingleStringTools(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	// YAML parses "tools: Read" as a bare string, not a list.
	// CompileAgent must normalize to a proper YAML array.
	fm := map[string]any{
		"name":  "test-agent",
		"tools": "Read",
	}
	content, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	if !strings.Contains(s, "- read_file") {
		t.Errorf("expected tools to be YAML array with read_file, got: %s", s)
	}
}

func TestGeminiCompiler_CompileAgent_StripsUnrecognizedKeys(t *testing.T) {
	t.Parallel()
	c := &compiler.GeminiCompiler{}

	fm := map[string]any{
		"name":        "test-agent",
		"description": "A test agent",
		"tools":       []any{"Read"},
		"color":       "cyan",
		"maxTurns":    150,
		"model":       "opus",
		"skills":      []any{"security-ref"},
		"hooks":       map[string]any{"PreToolUse": "check"},
		"memory":      map[string]any{"type": "user"},
	}
	content, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	// Only name, description, tools should survive
	for _, kept := range []string{"name:", "description:", "tools:"} {
		if !strings.Contains(s, kept) {
			t.Errorf("expected key %q in output: %s", kept, s)
		}
	}
	for _, stripped := range []string{"color:", "maxTurns:", "model:", "skills:", "hooks:", "memory:"} {
		if strings.Contains(s, stripped) {
			t.Errorf("CC-only key %q should be stripped for Gemini: %s", stripped, s)
		}
	}
}

func TestClaudeCompiler_CompileAgent_Passthrough(t *testing.T) {
	t.Parallel()
	c := &compiler.ClaudeCompiler{}

	// Claude compiler must NOT translate tool names
	fm := map[string]any{
		"name":            "test-agent",
		"tools":           []any{"Read", "Bash", "Task"},
		"disallowedTools": []any{"Write", "Edit"},
	}
	content, err := c.CompileAgent("test-agent", fm, "# body")
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	// CC names must be present unchanged
	for _, cc := range []string{"Read", "Bash", "Task", "Write", "Edit"} {
		if !strings.Contains(s, cc) {
			t.Errorf("CC tool %q should be preserved by ClaudeCompiler: %s", cc, s)
		}
	}
	// Gemini names must NOT appear
	for _, gemini := range []string{"read_file", "run_shell_command", "write_file", "replace"} {
		if strings.Contains(s, gemini) {
			t.Errorf("Gemini tool %q should NOT appear from ClaudeCompiler: %s", gemini, s)
		}
	}
}

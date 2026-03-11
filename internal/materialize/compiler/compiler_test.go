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

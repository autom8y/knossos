package compiler

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ClaudeCompiler struct{}

func (c *ClaudeCompiler) CompileCommand(name, description, argHint, body string) (string, []byte, error) {
	return name + ".md", []byte(body), nil
}

func (c *ClaudeCompiler) CompileSkill(name, description, body string) (string, string, []byte, error) {
	return name, "SKILL.md", []byte(body), nil
}

func (c *ClaudeCompiler) CompileAgent(name string, frontmatter map[string]any, body string) ([]byte, error) {
	yamlOut, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	result := []byte("---\n")
	result = append(result, yamlOut...)
	result = append(result, []byte("---\n")...)
	result = append(result, []byte(body)...)
	return result, nil
}

func (c *ClaudeCompiler) ContextFilename() string {
	return "CLAUDE.md"
}

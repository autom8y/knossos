package compiler

import (
	"bytes"
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
	var buf bytes.Buffer
	buf.WriteString("---\n")
	
	encoder := yaml.NewEncoder(&buf)
	if err := encoder.Encode(frontmatter); err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}
	encoder.Close()
	
	buf.WriteString("---\n")
	buf.WriteString(body)
	return buf.Bytes(), nil
}

func (c *ClaudeCompiler) ContextFilename() string {
	return "CLAUDE.md"
}

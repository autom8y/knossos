package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type GeminiCompiler struct{}

type GeminiCommand struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Prompt      string `toml:"prompt"`
}

func (c *GeminiCompiler) CompileCommand(name, description, argHint, body string) (string, []byte, error) {
	prompt := strings.TrimSpace(body)
	if argHint != "" {
		prompt += "\nUser arguments: {{args}}"
	}

	cmd := GeminiCommand{
		Name:        name,
		Description: description,
		Prompt:      prompt,
	}

	out, err := toml.Marshal(cmd)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal toml: %w", err)
	}

	return name + ".toml", out, nil
}

func (c *GeminiCompiler) CompileSkill(name, description, body string) (string, string, []byte, error) {
	var buf bytes.Buffer
	buf.WriteString("---\n")
	
	fm := map[string]string{
		"name":    name,
		"version": "1.0",
	}
	if description != "" {
		fm["description"] = description
	}
	
	encoder := yaml.NewEncoder(&buf)
	if err := encoder.Encode(fm); err != nil {
		return "", "", nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}
	encoder.Close()
	
	buf.WriteString("---\n")
	buf.WriteString(strings.TrimLeft(body, " \t\n\r"))

	return name, "SKILL.md", buf.Bytes(), nil
}

func (c *GeminiCompiler) CompileAgent(name string, frontmatter map[string]any, body string) ([]byte, error) {
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

func (c *GeminiCompiler) ContextFilename() string {
	return "GEMINI.md"
}

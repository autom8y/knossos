package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/autom8y/knossos/internal/channel"
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

// geminiAgentKeys are the only frontmatter keys Gemini CLI accepts for local agents.
// All other keys (color, maxTurns, model, skills, hooks, memory, disallowedTools)
// must be stripped or Gemini CLI rejects the agent with "Unrecognized key(s)".
var geminiAgentKeys = map[string]bool{
	"name":        true,
	"description": true,
	"tools":       true,
}

// CompileAgent translates CC agent frontmatter for Gemini consumption.
//
// Transformations applied:
//  1. tools field: CC names translated to Gemini equivalents; CC-only tools dropped.
//  2. All CC-specific frontmatter keys stripped (Gemini only accepts name, description, tools).
//
// The agent body is passed through unchanged — body-level substitutions are
// handled upstream by transformAgentContent() before CompileAgent is called.
func (c *GeminiCompiler) CompileAgent(name string, frontmatter map[string]any, body string) ([]byte, error) {
	// Translate tools field
	if tools, ok := extractStringSlice(frontmatter, "tools"); ok {
		translated := channel.TranslateFrontmatterTools(tools)
		if len(translated) == 0 {
			delete(frontmatter, "tools")
		} else {
			frontmatter["tools"] = translated
		}
	}

	// Strip all keys Gemini doesn't recognize. CC-specific keys like color,
	// maxTurns, model, skills, hooks, disallowedTools, memory cause Gemini CLI
	// to reject the agent with "Unrecognized key(s) in object".
	for key := range frontmatter {
		if !geminiAgentKeys[key] {
			delete(frontmatter, key)
		}
	}

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

// extractStringSlice extracts the named field from a frontmatter map as a []string.
// Handles []string (direct assignment), []any (YAML unmarshal of list), and string
// (YAML unmarshal of single value like "tools: Read"). The single-string case is
// critical: YAML parses "tools: Read" as string, not []string, causing Gemini CLI
// to reject with "Expected array, received string" if not normalized here.
// Returns (nil, false) if the field is absent or not a string-containing type.
func extractStringSlice(fm map[string]any, key string) ([]string, bool) {
	raw, ok := fm[key]
	if !ok {
		return nil, false
	}
	switch v := raw.(type) {
	case string:
		return []string{v}, true
	case []string:
		return v, true
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result, true
	default:
		return nil, false
	}
}

func (c *GeminiCompiler) ContextFilename() string {
	return "GEMINI.md"
}

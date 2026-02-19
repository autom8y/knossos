// Package agent provides agent frontmatter parsing, validation, and management.
package agent

import (
	"encoding/json"
	"fmt"

	"github.com/autom8y/knossos/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// FlexibleStringSlice is an alias for the shared frontmatter.FlexibleStringSlice type.
// It accepts both comma-separated strings and YAML lists.
type FlexibleStringSlice = frontmatter.FlexibleStringSlice

// MemoryField represents the CC agent memory configuration.
// It accepts both boolean (true → "project", false → disabled) and
// string enum ("user", "project", "local") values from YAML/JSON.
// Internally stored as a string; zero value "" means memory is disabled.
type MemoryField string

// validMemoryScopes lists the CC-recognized memory scope values.
var validMemoryScopes = map[string]bool{
	"user":    true,
	"project": true,
	"local":   true,
}

// UnmarshalYAML implements yaml.Unmarshaler.
// Accepts:
//   - bool true  → normalizes to "project"
//   - bool false → normalizes to "" (disabled)
//   - string     → stores as-is (validated later by validateCore)
func (m *MemoryField) UnmarshalYAML(value *yaml.Node) error {
	// Try boolean first (YAML tag !!bool)
	if value.Kind == yaml.ScalarNode && value.Tag == "!!bool" {
		var b bool
		if err := value.Decode(&b); err != nil {
			return err
		}
		if b {
			*m = "project"
		} else {
			*m = ""
		}
		return nil
	}

	// Try string
	if value.Kind == yaml.ScalarNode {
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		*m = MemoryField(s)
		return nil
	}

	return fmt.Errorf("memory must be a boolean or string, got %v", value.Tag)
}

// UnmarshalJSON implements json.Unmarshaler.
// Same normalization rules as UnmarshalYAML.
func (m *MemoryField) UnmarshalJSON(data []byte) error {
	// Try boolean
	if string(data) == "true" {
		*m = "project"
		return nil
	}
	if string(data) == "false" {
		*m = ""
		return nil
	}

	// Try string (must be quoted)
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("memory must be a boolean or string: %w", err)
	}
	*m = MemoryField(s)
	return nil
}

// MarshalJSON implements json.Marshaler.
// Serializes as the normalized string scope value.
// Empty string serializes as empty string (omitted by omitempty tag).
func (m MemoryField) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(m))
}

// IsEnabled returns true if memory is configured with any scope.
func (m MemoryField) IsEnabled() bool {
	return m != ""
}

// Scope returns the memory scope string ("user", "project", "local")
// or empty string if memory is disabled.
func (m MemoryField) Scope() string {
	return string(m)
}

// UpstreamRef describes an agent or source that feeds into this agent.
type UpstreamRef struct {
	// Source is the upstream agent name or external input description.
	Source string `yaml:"source" json:"source"`
	// Artifact is the expected artifact from upstream.
	Artifact string `yaml:"artifact,omitempty" json:"artifact,omitempty"`
}

// DownstreamRef describes an agent this agent routes work to.
type DownstreamRef struct {
	// Agent is the target agent name.
	Agent string `yaml:"agent" json:"agent"`
	// Condition describes when to route to this agent.
	Condition string `yaml:"condition,omitempty" json:"condition,omitempty"`
	// Artifact is the artifact passed to the downstream agent.
	Artifact string `yaml:"artifact,omitempty" json:"artifact,omitempty"`
}

// ArtifactDecl declares an artifact this agent produces.
type ArtifactDecl struct {
	// Artifact is the artifact name or type.
	Artifact string `yaml:"artifact" json:"artifact"`
	// Format is the expected format (e.g., "markdown", "yaml", "json").
	Format string `yaml:"format,omitempty" json:"format,omitempty"`
}

// McpServerConfig describes an MCP server configuration for CC pass-through.
type McpServerConfig struct {
	// Name is the MCP server identifier.
	Name string `yaml:"name" json:"name"`
	// URL is the MCP server endpoint.
	URL string `yaml:"url" json:"url"`
}

// BehavioralContract defines behavioral constraints and requirements.
type BehavioralContract struct {
	// MustUse lists tools or patterns the agent must use.
	MustUse []string `yaml:"must_use,omitempty" json:"must_use,omitempty"`
	// MustProduce lists artifacts the agent must produce.
	MustProduce []string `yaml:"must_produce,omitempty" json:"must_produce,omitempty"`
	// MustNot lists actions the agent must not take.
	MustNot []string `yaml:"must_not,omitempty" json:"must_not,omitempty"`
	// MaxTurns is the maximum conversation turns before requiring handoff.
	MaxTurns int `yaml:"max_turns,omitempty" json:"max_turns,omitempty"`
}

// Package agent provides agent frontmatter parsing, validation, and management.
package agent

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// FlexibleStringSlice is a YAML type that accepts both a comma-separated string
// (e.g., "Bash, Read, Glob") and a proper YAML list (e.g., [Bash, Read, Glob]).
// This is a local copy of materialize.FlexibleStringSlice to avoid coupling
// the agent package to the materialize package for a single utility type.
type FlexibleStringSlice []string

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (f *FlexibleStringSlice) UnmarshalYAML(value *yaml.Node) error {
	// Try as a sequence first
	if value.Kind == yaml.SequenceNode {
		var slice []string
		if err := value.Decode(&slice); err != nil {
			return err
		}
		*f = slice
		return nil
	}

	// Fall back to comma-separated string
	var str string
	if err := value.Decode(&str); err != nil {
		return err
	}

	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	*f = result
	return nil
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

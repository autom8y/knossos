// Package agent provides agent frontmatter parsing, validation, and management.
package agent

import (
	"github.com/autom8y/knossos/internal/frontmatter"
)

// FlexibleStringSlice is an alias for the shared frontmatter.FlexibleStringSlice type.
// It accepts both comma-separated strings and YAML lists.
type FlexibleStringSlice = frontmatter.FlexibleStringSlice

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

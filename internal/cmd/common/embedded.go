package common

import "io/fs"

// Package-level storage for embedded assets, set by main.go via SetEmbeddedAssets.
// These are accessed by any command that creates a Materializer.
var (
	embeddedRites     fs.FS
	embeddedTemplates fs.FS
	embeddedHooksYAML []byte
	embeddedAgents    fs.FS
	embeddedMena      fs.FS
)

// SetEmbeddedAssets stores embedded rite definitions, templates, and hooks
// configuration for use by commands that create Materializers.
func SetEmbeddedAssets(rites, templates fs.FS, hooksYAML []byte) {
	embeddedRites = rites
	embeddedTemplates = templates
	embeddedHooksYAML = hooksYAML
}

// EmbeddedRites returns the embedded rites filesystem, or nil if not set.
func EmbeddedRites() fs.FS { return embeddedRites }

// EmbeddedTemplates returns the embedded templates filesystem, or nil if not set.
func EmbeddedTemplates() fs.FS { return embeddedTemplates }

// EmbeddedHooksYAML returns the embedded hooks.yaml bytes, or nil if not set.
// Used by "ari init" to bootstrap config/hooks.yaml in new projects.
func EmbeddedHooksYAML() []byte { return embeddedHooksYAML }

// SetEmbeddedUserAssets stores embedded agent and mena definitions
// for use as fallback when KNOSSOS_HOME is unavailable.
func SetEmbeddedUserAssets(agents, mena fs.FS) {
	embeddedAgents = agents
	embeddedMena = mena
}

// EmbeddedAgents returns the embedded agents filesystem, or nil if not set.
func EmbeddedAgents() fs.FS { return embeddedAgents }

// EmbeddedMena returns the embedded mena filesystem, or nil if not set.
func EmbeddedMena() fs.FS { return embeddedMena }

package common

import (
	"io/fs"

	"github.com/autom8y/knossos/internal/assets"
)

// SetEmbeddedAssets stores embedded rite definitions, templates, and hooks
// configuration for use by commands that create Materializers.
func SetEmbeddedAssets(rites, templates fs.FS, hooksYAML []byte) {
	assets.SetEmbedded(rites, templates, hooksYAML)
}

// EmbeddedRites returns the embedded rites filesystem, or nil if not set.
func EmbeddedRites() fs.FS { return assets.Rites() }

// EmbeddedTemplates returns the embedded templates filesystem, or nil if not set.
func EmbeddedTemplates() fs.FS { return assets.Templates() }

// EmbeddedHooksYAML returns the embedded hooks.yaml bytes, or nil if not set.
// Used by "ari init" to bootstrap config/hooks.yaml in new projects.
func EmbeddedHooksYAML() []byte { return assets.HooksYAML() }

// SetEmbeddedUserAssets stores embedded agent and mena definitions
// for use as fallback when KNOSSOS_HOME is unavailable.
func SetEmbeddedUserAssets(agents, mena fs.FS) {
	assets.SetUserAssets(agents, mena)
}

// EmbeddedAgents returns the embedded agents filesystem, or nil if not set.
func EmbeddedAgents() fs.FS { return assets.Agents() }

// EmbeddedMena returns the embedded mena filesystem, or nil if not set.
func EmbeddedMena() fs.FS { return assets.Mena() }

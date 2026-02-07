package common

import "io/fs"

// Package-level storage for embedded assets, set by main.go via SetEmbeddedAssets.
// These are accessed by any command that creates a Materializer.
var (
	embeddedRites     fs.FS
	embeddedTemplates fs.FS
	embeddedHooks     fs.FS
)

// SetEmbeddedAssets stores embedded rite definitions, templates, and hooks
// configuration for use by commands that create Materializers.
func SetEmbeddedAssets(rites, templates, hooks fs.FS) {
	embeddedRites = rites
	embeddedTemplates = templates
	embeddedHooks = hooks
}

// EmbeddedRites returns the embedded rites filesystem, or nil if not set.
func EmbeddedRites() fs.FS { return embeddedRites }

// EmbeddedTemplates returns the embedded templates filesystem, or nil if not set.
func EmbeddedTemplates() fs.FS { return embeddedTemplates }

// EmbeddedHooks returns the embedded hooks filesystem, or nil if not set.
func EmbeddedHooks() fs.FS { return embeddedHooks }

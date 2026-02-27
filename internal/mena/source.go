// Package mena provides shared mena resolution primitives used by both
// the registry validator and the materialize projection pipeline.
// This is a LEAF package — it imports only stdlib (os, path/filepath,
// io/fs, strings). No internal/ imports.
package mena

import (
	"io/fs"
	"path/filepath"
)

// MenaSource represents a source for mena files. It can be either a
// filesystem path or an embedded FS path.
type MenaSource struct {
	Path       string // Filesystem path (for os-based sources)
	Fsys       fs.FS  // Embedded filesystem (nil for os-based sources)
	FsysPath   string // Path within Fsys (e.g., "rites/shared/mena")
	IsEmbedded bool
}

// SourceChainOptions configures source chain construction.
type SourceChainOptions struct {
	// RitePath is the absolute path to the active rite directory.
	RitePath string

	// RitesBase is the parent directory containing all rites (e.g., "rites/").
	// When empty, only rite-local mena is included.
	RitesBase string

	// Dependencies is the list of dependency rite names from the manifest.
	// "shared" is handled implicitly and filtered from this list.
	Dependencies []string

	// PlatformMenaDir is the resolved platform mena directory path.
	// When non-empty, added as the lowest-priority source (position 0).
	// Comes from the caller's getMenaDir() resolution or equivalent.
	PlatformMenaDir string
}

// BuildSourceChain constructs a priority-ordered list of MenaSource entries.
// Order: platform (lowest) -> shared -> dependencies -> rite-local (highest).
// Later entries in the returned slice have higher priority.
//
// Sources with empty paths are included (the caller's exists-check or
// collection logic handles nonexistent directories gracefully).
func BuildSourceChain(opts SourceChainOptions) []MenaSource {
	var sources []MenaSource

	// 1. Platform mena (lowest priority)
	if opts.PlatformMenaDir != "" {
		sources = append(sources, MenaSource{Path: opts.PlatformMenaDir})
	}

	// 2. Shared and dependency mena (only when RitesBase is set)
	if opts.RitesBase != "" {
		// Shared rite mena
		sources = append(sources, MenaSource{Path: filepath.Join(opts.RitesBase, "shared", "mena")})

		// Dependency rite mena (in manifest order); "shared" is handled above, skip it
		for _, dep := range opts.Dependencies {
			if dep != "shared" {
				sources = append(sources, MenaSource{Path: filepath.Join(opts.RitesBase, dep, "mena")})
			}
		}
	}

	// 3. Rite-local mena (highest priority)
	sources = append(sources, MenaSource{Path: filepath.Join(opts.RitePath, "mena")})

	return sources
}

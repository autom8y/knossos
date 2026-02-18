// Package materialize re-exports frontmatter types from the mena sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/mena"
)

// Type aliases for backward compatibility.
type (
	MenaFrontmatter    = mena.MenaFrontmatter
	FlexibleStringSlice = mena.FlexibleStringSlice
)

// Re-export the frontmatter parser (used by core tests).
var parseMenaFrontmatterBytes = mena.ParseMenaFrontmatterBytes

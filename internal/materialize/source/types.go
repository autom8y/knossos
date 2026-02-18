// Package source provides rite source resolution for the materialize pipeline.
package source

import "fmt"

// SourceType represents the type of rite source.
type SourceType string

const (
	// SourceProject represents rites from the current project's rites/ directory.
	SourceProject SourceType = "project"
	// SourceUser represents rites from user-level ~/.local/share/knossos/rites/.
	SourceUser SourceType = "user"
	// SourceKnossos represents rites from the Knossos platform ($KNOSSOS_HOME/rites).
	SourceKnossos SourceType = "knossos"
	// SourceExplicit represents an explicitly specified path via --source flag.
	SourceExplicit SourceType = "explicit"
	// SourceEmbedded represents rites compiled into the binary via //go:embed.
	SourceEmbedded SourceType = "embedded"
)

// RiteSource represents a location where rites can be found.
type RiteSource struct {
	Type        SourceType `json:"type"`
	Path        string     `json:"path"`
	Description string     `json:"description,omitempty"`
}

// String returns a human-readable representation of the source.
func (s RiteSource) String() string {
	return fmt.Sprintf("%s:%s", s.Type, s.Path)
}

// ResolvedRite contains the resolution result for a rite.
type ResolvedRite struct {
	Name         string     `json:"name"`
	Source       RiteSource `json:"source"`
	RitePath     string     `json:"rite_path"`
	ManifestPath string     `json:"manifest_path"`
	TemplatesDir string     `json:"templates_dir,omitempty"`
}

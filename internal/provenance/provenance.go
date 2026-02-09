// Package provenance provides unified file-level provenance tracking for .claude/.
// The provenance manifest records the origin and ownership state of all files
// materialized by the knossos pipeline, enabling divergence detection and safe
// ownership transitions.
package provenance

import (
	"path/filepath"
	"time"
)

// ManifestFileName is the provenance manifest filename within .claude/.
const ManifestFileName = "PROVENANCE_MANIFEST.yaml"

// UserManifestFileName is the user-level provenance manifest filename.
const UserManifestFileName = "USER_PROVENANCE_MANIFEST.yaml"

// CurrentSchemaVersion is the current manifest schema version.
// Starts at "2.0" for the provenance manifest (independent of inscription's "1.0").
const CurrentSchemaVersion = "2.0"

// ProvenanceManifest is the unified file-level provenance tracker for .claude/.
// Stored at .claude/PROVENANCE_MANIFEST.yaml.
type ProvenanceManifest struct {
	// SchemaVersion is the manifest format version. Currently "1.0".
	SchemaVersion string `yaml:"schema_version"`

	// LastSync is the UTC timestamp of the most recent materialization.
	LastSync time.Time `yaml:"last_sync"`

	// ActiveRite is the rite name that produced this manifest.
	// Empty string for minimal (cross-cutting) materializations.
	ActiveRite string `yaml:"active_rite,omitempty"`

	// Entries maps relative paths within .claude/ to their provenance records.
	// Keys use forward slashes. Directory entries end with "/" (mena only).
	// Examples: "agents/orchestrator.md", "commands/commit/", "CLAUDE.md"
	Entries map[string]*ProvenanceEntry `yaml:"entries"`
}

// ProvenanceEntry tracks the origin and state of a single file or directory in .claude/.
type ProvenanceEntry struct {
	// Owner determines sync behavior for this entry.
	Owner OwnerType `yaml:"owner"`

	// Scope indicates whether this entry belongs to rite or user provenance.
	Scope ScopeType `yaml:"scope"`

	// SourcePath is the relative path (from project root) to the source file.
	// Empty for user-created files.
	// Examples: "rites/ecosystem/agents/orchestrator.md",
	//           "mena/operations/commit/INDEX.dro.md",
	//           "knossos/templates/rules/internal-hook.md"
	SourcePath string `yaml:"source_path,omitempty"`

	// SourceType records which tier of the source resolution chain provided the file.
	// Values match materialize.SourceType: "project", "user", "knossos", "explicit", "embedded".
	// Additional values for mena provenance: "template", "shared", "dependency".
	SourceType string `yaml:"source_type,omitempty"`

	// Checksum is the SHA256 hash of the file (or directory for mena) at write time.
	// Uses the "sha256:" prefix per ADR-0026 and internal/checksum convention.
	Checksum string `yaml:"checksum"`

	// LastSynced is the UTC timestamp when this entry was last written by the pipeline.
	LastSynced time.Time `yaml:"last_synced"`
}

// OwnerType represents who owns a file in .claude/.
type OwnerType string

const (
	// OwnerKnossos indicates files managed by Knossos.
	// These are safe to overwrite on sync.
	OwnerKnossos OwnerType = "knossos"

	// OwnerUser indicates files created or modified by the user.
	// These are NEVER overwritten by the pipeline.
	OwnerUser OwnerType = "user"

	// OwnerUntracked indicates pre-existing files discovered during bootstrap.
	// Treated as user-owned for safety. Promoted to OwnerUser or OwnerKnossos
	// on the next sync that interacts with the file.
	OwnerUntracked OwnerType = "untracked"
)

// IsValid returns true if the owner type is a recognized value.
func (o OwnerType) IsValid() bool {
	switch o {
	case OwnerKnossos, OwnerUser, OwnerUntracked:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (o OwnerType) String() string {
	return string(o)
}

// ScopeType represents the provenance scope (rite or user).
type ScopeType string

const (
	// ScopeRite indicates entries tracked in project-level PROVENANCE_MANIFEST.yaml.
	ScopeRite ScopeType = "rite"

	// ScopeUser indicates entries tracked in user-level USER_PROVENANCE_MANIFEST.yaml.
	ScopeUser ScopeType = "user"
)

// IsValid returns true if the scope type is a recognized value.
func (s ScopeType) IsValid() bool {
	switch s {
	case ScopeRite, ScopeUser:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s ScopeType) String() string {
	return string(s)
}

// UserManifestPath returns the full path to USER_PROVENANCE_MANIFEST.yaml within the user .claude directory.
func UserManifestPath(userClaudeDir string) string {
	return filepath.Join(userClaudeDir, UserManifestFileName)
}

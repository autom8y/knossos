// Package provenance provides unified file-level provenance tracking for .claude/.
// The provenance manifest records the origin and ownership state of all files
// materialized by the knossos pipeline, enabling divergence detection and safe
// ownership transitions.
package provenance

import "time"

// ManifestFileName is the provenance manifest filename within .claude/.
const ManifestFileName = "PROVENANCE_MANIFEST.yaml"

// CurrentSchemaVersion is the current manifest schema version.
// Starts at "1.0" for the provenance manifest (independent of inscription's "1.0").
const CurrentSchemaVersion = "1.0"

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

	// SourcePipeline identifies which pipeline placed this file.
	// "materialize" for project-level, empty string for user-created files.
	SourcePipeline string `yaml:"source_pipeline,omitempty"`

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

	// OwnerUnknown indicates pre-existing files discovered during bootstrap.
	// Treated as user-owned for safety. Promoted to OwnerUser or OwnerKnossos
	// on the next sync that interacts with the file.
	OwnerUnknown OwnerType = "unknown"
)

// IsValid returns true if the owner type is a recognized value.
func (o OwnerType) IsValid() bool {
	switch o {
	case OwnerKnossos, OwnerUser, OwnerUnknown:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (o OwnerType) String() string {
	return string(o)
}

package provenance

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
	"gopkg.in/yaml.v3"
)

// Load reads and parses the PROVENANCE_MANIFEST.yaml file from the given path.
// Returns an error if the file cannot be read or parsed.
func Load(path string) (*ProvenanceManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"provenance manifest not found",
				map[string]any{"path": path})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read provenance manifest", err)
	}

	var manifest ProvenanceManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, errors.NewWithDetails(errors.CodeParseError,
			"failed to parse provenance manifest YAML",
			map[string]any{
				"path":  path,
				"cause": err.Error(),
			})
	}

	// Migrate v1 to v2 if needed
	migrateV1ToV2(&manifest)

	// Validate manifest
	if err := validateManifest(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// Save writes the manifest to disk, but only if structural content has changed.
// Uses a stable comparison (excluding LastSync timestamp) to avoid triggering
// CC's file watcher on no-op syncs. The watcher crash is a known issue when
// atomic writes (temp file + rename) happen in .claude/ unnecessarily.
func Save(path string, manifest *ProvenanceManifest) error {
	// Validate before saving
	if err := validateManifest(manifest); err != nil {
		return err
	}

	// Marshal the full manifest (with current timestamp)
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal provenance manifest", err)
	}

	// Compare structural content only (exclude volatile last_sync timestamp).
	// This prevents timestamp-only changes from triggering atomic writes
	// which create temp files in .claude/ and trigger CC's file watcher.
	existing, readErr := os.ReadFile(path)
	if readErr == nil {
		// Parse existing manifest to compare structurally
		var existingManifest ProvenanceManifest
		if parseErr := yaml.Unmarshal(existing, &existingManifest); parseErr == nil {
			if structurallyEqual(manifest, &existingManifest) {
				return nil // No structural change, skip write entirely
			}
		}
	}

	// Structural change detected (or first write) — write atomically
	if err := fileutil.AtomicWriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write provenance manifest", err)
	}

	return nil
}

// structurallyEqual compares two manifests ignoring volatile fields (LastSync,
// per-entry LastSynced timestamps). Returns true if all structural content
// (schema version, active rite, entries with owners/sources/checksums) is identical.
func structurallyEqual(a, b *ProvenanceManifest) bool {
	if a.SchemaVersion != b.SchemaVersion || a.ActiveRite != b.ActiveRite {
		return false
	}
	if len(a.Entries) != len(b.Entries) {
		return false
	}
	for path, entryA := range a.Entries {
		entryB, ok := b.Entries[path]
		if !ok {
			return false
		}
		if entryA.Owner != entryB.Owner ||
			entryA.Scope != entryB.Scope ||
			entryA.SourcePath != entryB.SourcePath ||
			entryA.SourceType != entryB.SourceType ||
			entryA.Checksum != entryB.Checksum {
			return false
		}
	}
	return true
}

// LoadOrBootstrap loads the manifest from the given path.
// If the file doesn't exist, returns an empty manifest with the current schema version.
func LoadOrBootstrap(path string) (*ProvenanceManifest, error) {
	manifest, err := Load(path)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return empty manifest for bootstrap
			return &ProvenanceManifest{
				SchemaVersion: CurrentSchemaVersion,
				LastSync:      time.Time{},
				ActiveRite:    "",
				Entries:       make(map[string]*ProvenanceEntry),
			}, nil
		}
		return nil, err
	}
	return manifest, nil
}

// ManifestPath returns the full path to PROVENANCE_MANIFEST.yaml within the .claude directory.
func ManifestPath(claudeDir string) string {
	return filepath.Join(claudeDir, ManifestFileName)
}

// validateManifest validates the manifest structure and content per TDD Section 1.
func validateManifest(manifest *ProvenanceManifest) error {
	var issues []string

	// Required: SchemaVersion must match ^[0-9]+\.[0-9]+$
	if manifest.SchemaVersion == "" {
		issues = append(issues, "missing required field: schema_version")
	} else if !isValidSchemaVersion(manifest.SchemaVersion) {
		issues = append(issues, "invalid schema_version format: "+manifest.SchemaVersion+" (expected N.N)")
	}

	// Required: LastSync must be non-zero time
	if manifest.LastSync.IsZero() {
		issues = append(issues, "missing required field: last_sync (must be non-zero time)")
	}

	// Required: Entries must be non-nil map (may be empty)
	if manifest.Entries == nil {
		issues = append(issues, "missing required field: entries (must be non-nil map)")
	}

	// Validate each entry
	for path, entry := range manifest.Entries {
		// Required: Entry.Owner
		if entry.Owner == "" {
			issues = append(issues, "entry '"+path+"' missing required field: owner")
		} else if !entry.Owner.IsValid() {
			issues = append(issues, "entry '"+path+"' has invalid owner: "+string(entry.Owner))
		}

		// If Owner is knossos, SourcePath must be non-empty
		if entry.Owner == OwnerKnossos && entry.SourcePath == "" {
			issues = append(issues, "entry '"+path+"' with owner 'knossos' requires non-empty source_path")
		}

		// If Owner is knossos, SourceType must be non-empty
		if entry.Owner == OwnerKnossos && entry.SourceType == "" {
			issues = append(issues, "entry '"+path+"' with owner 'knossos' requires non-empty source_type")
		}

		// Required: Entry.Scope must be present and valid
		if entry.Scope == "" {
			issues = append(issues, "entry '"+path+"' missing required field: scope")
		} else if !entry.Scope.IsValid() {
			issues = append(issues, "entry '"+path+"' has invalid scope: "+string(entry.Scope))
		}

		// Required: Entry.Checksum must match ^sha256:[0-9a-f]{64}$
		if entry.Checksum == "" {
			issues = append(issues, "entry '"+path+"' missing required field: checksum")
		} else if !isValidChecksum(entry.Checksum) {
			issues = append(issues, "entry '"+path+"' has invalid checksum format: "+entry.Checksum+" (expected sha256:HASH)")
		}

		// Required: Entry.LastSynced must be non-zero time
		if entry.LastSynced.IsZero() {
			issues = append(issues, "entry '"+path+"' missing required field: last_synced (must be non-zero time)")
		}
	}

	if len(issues) > 0 {
		return errors.ErrSchemaInvalid("provenance manifest", issues)
	}

	return nil
}

// isValidSchemaVersion checks if a version string matches ^[0-9]+\.[0-9]+$ format.
func isValidSchemaVersion(version string) bool {
	matched, _ := regexp.MatchString(`^[0-9]+\.[0-9]+$`, version)
	return matched
}

// isValidChecksum checks if a checksum string matches ^sha256:[0-9a-f]{64}$ format.
func isValidChecksum(checksum string) bool {
	matched, _ := regexp.MatchString(`^sha256:[0-9a-f]{64}$`, checksum)
	return matched
}

// migrateV1ToV2 migrates a v1.0 manifest to v2.0 schema in-place.
// Converts SourcePipeline to Scope and "unknown" owner to "untracked".
func migrateV1ToV2(m *ProvenanceManifest) {
	// Only migrate if it's v1.0 (or missing version, treat as v1.0)
	if m.SchemaVersion != "1.0" && m.SchemaVersion != "" {
		return
	}

	// Update schema version
	m.SchemaVersion = "2.0"

	// Migrate entries
	for _, entry := range m.Entries {
		// Convert empty Scope to ScopeRite (all v1.0 entries were rite-scoped)
		if entry.Scope == "" {
			entry.Scope = ScopeRite
		}

		// Convert "unknown" owner to "untracked"
		if entry.Owner == "unknown" {
			entry.Owner = OwnerUntracked
		}
	}
}

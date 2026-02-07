package usersync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ManifestVersion is the current manifest schema version.
// Version 2.0 adds mena_type and target fields for ResourceMena entries.
const ManifestVersion = "2.0"

// Manifest represents a user resource manifest.
type Manifest struct {
	Version  string           `json:"manifest_version"`
	LastSync time.Time        `json:"last_sync"`
	Entries  map[string]Entry `json:"-"` // Entries are stored dynamically based on resource type
}

// Entry represents a single resource entry in the manifest.
type Entry struct {
	Source      SourceType `json:"source"`
	InstalledAt time.Time  `json:"installed_at"`
	Checksum    string     `json:"checksum"`
	MenaType    string     `json:"mena_type,omitempty"` // "dro" or "lego" (mena entries only)
	Target      string     `json:"target,omitempty"`     // "commands" or "skills" (mena entries only)
}

// manifestJSON is the on-disk manifest format.
// The entry key varies by resource type.
type manifestJSON struct {
	Version  string               `json:"manifest_version"`
	LastSync string               `json:"last_sync"`
	Agents   map[string]entryJSON `json:"agents,omitempty"`
	Mena     map[string]entryJSON `json:"mena,omitempty"`
	Hooks    map[string]entryJSON `json:"hooks,omitempty"`
}

type entryJSON struct {
	Source      string `json:"source"`
	InstalledAt string `json:"installed_at"`
	Checksum    string `json:"checksum"`
	MenaType    string `json:"mena_type,omitempty"` // "dro" or "lego" (mena entries only)
	Target      string `json:"target,omitempty"`     // "commands" or "skills" (mena entries only)
}

// loadManifest reads the manifest from disk.
func (s *Syncer) loadManifest() (*Manifest, error) {
	data, err := os.ReadFile(s.manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty manifest
			return &Manifest{
				Version:  ManifestVersion,
				LastSync: time.Time{},
				Entries:  make(map[string]Entry),
			}, nil
		}
		return nil, ErrManifestRead(s.manifestPath, err)
	}

	var mj manifestJSON
	if err := json.Unmarshal(data, &mj); err != nil {
		// Manifest corrupt - backup and create new
		backupPath := s.manifestPath + ".corrupt"
		os.Rename(s.manifestPath, backupPath)
		return &Manifest{
			Version:  ManifestVersion,
			LastSync: time.Time{},
			Entries:  make(map[string]Entry),
		}, nil
	}

	// Version mismatch: backup old manifest and start fresh.
	// This handles migration from v1.0 to v2.0 manifests (wipe-and-resync per D10).
	if mj.Version != ManifestVersion {
		backupPath := s.manifestPath + ".v1-backup"
		os.WriteFile(backupPath, data, 0644) // Best effort backup
		return &Manifest{
			Version:  ManifestVersion,
			LastSync: time.Time{},
			Entries:  make(map[string]Entry),
		}, nil
	}

	// Convert to internal format
	manifest := &Manifest{
		Version: mj.Version,
		Entries: make(map[string]Entry),
	}

	if t, err := time.Parse(time.RFC3339, mj.LastSync); err == nil {
		manifest.LastSync = t
	}

	// Get entries based on resource type
	var entries map[string]entryJSON
	switch s.resourceType {
	case ResourceAgents:
		entries = mj.Agents
	case ResourceMena:
		entries = mj.Mena
	case ResourceHooks:
		entries = mj.Hooks
	}

	for name, ej := range entries {
		entry := Entry{
			Source:   SourceType(ej.Source),
			Checksum: ej.Checksum,
			MenaType: ej.MenaType,
			Target:   ej.Target,
		}
		if t, err := time.Parse(time.RFC3339, ej.InstalledAt); err == nil {
			entry.InstalledAt = t
		}
		manifest.Entries[name] = entry
	}

	return manifest, nil
}

// saveManifest writes the manifest to disk.
func (s *Syncer) saveManifest(manifest *Manifest) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.manifestPath), 0755); err != nil {
		return ErrManifestWrite(s.manifestPath, err)
	}

	// Convert to JSON format
	mj := manifestJSON{
		Version:  manifest.Version,
		LastSync: manifest.LastSync.Format(time.RFC3339),
	}

	entries := make(map[string]entryJSON)
	for name, entry := range manifest.Entries {
		entries[name] = entryJSON{
			Source:      string(entry.Source),
			InstalledAt: entry.InstalledAt.Format(time.RFC3339),
			Checksum:    entry.Checksum,
			MenaType:    entry.MenaType,
			Target:      entry.Target,
		}
	}

	// Set entries in correct field
	switch s.resourceType {
	case ResourceAgents:
		mj.Agents = entries
	case ResourceMena:
		mj.Mena = entries
	case ResourceHooks:
		mj.Hooks = entries
	}

	data, err := json.MarshalIndent(mj, "", "  ")
	if err != nil {
		return ErrManifestWrite(s.manifestPath, err)
	}

	// Write atomically using temp file + rename
	tmpPath := s.manifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return ErrManifestWrite(s.manifestPath, err)
	}

	if err := os.Rename(tmpPath, s.manifestPath); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return ErrManifestWrite(s.manifestPath, err)
	}

	return nil
}

// cleanupOldManifests removes legacy manifest files after a successful
// unified mena manifest save. Only applies to ResourceMena.
func (s *Syncer) cleanupOldManifests() {
	if s.resourceType != ResourceMena {
		return
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	oldManifests := []string{
		filepath.Join(homeDir, ".claude", "USER_COMMAND_MANIFEST.json"),
		filepath.Join(homeDir, ".claude", "USER_SKILL_MANIFEST.json"),
	}
	for _, path := range oldManifests {
		os.Remove(path) // Ignore errors -- they may not exist
	}
}

// LoadManifest loads a manifest from a file path directly.
// This is a convenience function for testing and external use.
func LoadManifest(path string, resourceType ResourceType) (*Manifest, error) {
	s := &Syncer{
		manifestPath: path,
		resourceType: resourceType,
	}
	return s.loadManifest()
}

// SaveManifest saves a manifest to a file path directly.
// This is a convenience function for testing and external use.
func SaveManifest(path string, resourceType ResourceType, manifest *Manifest) error {
	s := &Syncer{
		manifestPath: path,
		resourceType: resourceType,
	}
	return s.saveManifest(manifest)
}

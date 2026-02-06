package usersync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ManifestVersion is the current manifest schema version.
const ManifestVersion = "1.0"

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
}

// manifestJSON is the on-disk manifest format (for backward compatibility).
// The entry key varies by resource type.
type manifestJSON struct {
	Version  string            `json:"manifest_version"`
	LastSync string            `json:"last_sync"`
	Agents   map[string]entryJSON `json:"agents,omitempty"`
	Skills   map[string]entryJSON `json:"skills,omitempty"`
	Commands map[string]entryJSON `json:"commands,omitempty"`
	Hooks    map[string]entryJSON `json:"hooks,omitempty"`
}

type entryJSON struct {
	Source      string `json:"source"`
	InstalledAt string `json:"installed_at"`
	Checksum    string `json:"checksum"`
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
	case ResourceSkills:
		entries = mj.Skills
	case ResourceCommands:
		entries = mj.Commands
	case ResourceHooks:
		entries = mj.Hooks
	}

	for name, ej := range entries {
		entry := Entry{
			Source:   NormalizeSource(SourceType(ej.Source)),
			Checksum: ej.Checksum,
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
		}
	}

	// Set entries in correct field
	switch s.resourceType {
	case ResourceAgents:
		mj.Agents = entries
	case ResourceSkills:
		mj.Skills = entries
	case ResourceCommands:
		mj.Commands = entries
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

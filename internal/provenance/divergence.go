package provenance

import (
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
)

// DivergenceReport contains the results of divergence detection.
type DivergenceReport struct {
	// Promoted contains entries that were promoted from knossos to user ownership.
	Promoted map[string]*ProvenanceEntry

	// CarriedForward contains user and unknown entries from the previous manifest.
	CarriedForward map[string]*ProvenanceEntry

	// Removed contains entries that no longer exist on disk.
	Removed []string
}

// DetectDivergence compares the previous manifest against the current filesystem state
// and identifies knossos-owned files that have been modified by the user.
// Returns a DivergenceReport containing promoted entries, carried-forward entries, and removed files.
// Algorithm per TDD Section 6.
func DetectDivergence(previous *ProvenanceManifest, current map[string]*ProvenanceEntry, channelDir string) (*DivergenceReport, error) {
	report := &DivergenceReport{
		Promoted:       make(map[string]*ProvenanceEntry),
		CarriedForward: make(map[string]*ProvenanceEntry),
		Removed:        []string{},
	}

	// If no previous manifest, no divergence possible
	if previous == nil {
		return report, nil
	}

	// Process each entry in previous manifest
	for path, entry := range previous.Entries {
		// Non-knossos entries are carried forward unchanged
		if entry.Owner != OwnerKnossos {
			report.CarriedForward[path] = entry
			continue
		}

		// Knossos-owned entries: check for divergence
		currentChecksum, err := computeCurrentChecksum(channelDir, path)
		if err != nil {
			// Error computing checksum: treat as missing/removed
			promotedEntry := *entry
			promotedEntry.Owner = OwnerUser
			promotedEntry.Checksum = ""
			report.Promoted[path] = &promotedEntry
			report.Removed = append(report.Removed, path)
			continue
		}

		if currentChecksum == "" {
			// File was deleted by user. Promote to user-owned so pipeline
			// does not recreate it (respects user intent to remove).
			promotedEntry := *entry
			promotedEntry.Owner = OwnerUser
			promotedEntry.Checksum = ""
			report.Promoted[path] = &promotedEntry
			report.Removed = append(report.Removed, path)
			continue
		}

		if currentChecksum != entry.Checksum {
			// User modified a knossos file. Promote to user-owned.
			promotedEntry := *entry
			promotedEntry.Owner = OwnerUser
			promotedEntry.Checksum = currentChecksum
			// Scope, SourcePath, SourceType retained for provenance history
			report.Promoted[path] = &promotedEntry
			continue
		}

		// Checksum matches: file unchanged, will be handled by pipeline
	}

	return report, nil
}

// computeCurrentChecksum computes the checksum of the file or directory at the given path.
// relativePath is relative to channelDir.
// Returns empty string if the file/directory doesn't exist or can't be read.
func computeCurrentChecksum(channelDir string, relativePath string) (string, error) {
	fullPath := filepath.Join(channelDir, relativePath)

	if strings.HasSuffix(relativePath, "/") {
		// Directory-level entry (mena). Use checksum.Dir().
		dirPath := strings.TrimSuffix(fullPath, "/")
		hash, err := checksum.Dir(dirPath)
		if err != nil {
			return "", err
		}
		return hash, nil
	} else {
		// File-level entry. Use checksum.File().
		hash, err := checksum.File(fullPath)
		if err != nil {
			return "", err
		}
		return hash, nil
	}
}

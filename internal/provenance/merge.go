package provenance

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Merge combines collector entries, divergence report, and previous manifest into a final manifest.
// This is the 4-step merge algorithm that determines provenance ownership in the channel dir.
//
// knossosDir is the .knossos/ sibling directory — some tracked files (e.g. ACTIVE_WORKFLOW.yaml)
// live there rather than in channelDir. When checking file existence in Step 0, both directories
// are searched.
//
// Steps:
//
//  0. Carry forward knossos entries from prev that still exist on disk (idempotency)
//  1. Layer promoted + carried-forward entries from divergence detection (skip empty checksums)
//  2. Layer collector entries (pipeline-written), EXCEPT where divergence promoted to user
//  3. Promote prev untracked entries not written this sync → user
func Merge(
	channelDir string,
	knossosDir string,
	activeRite string,
	collector Collector,
	divergenceReport *DivergenceReport,
	prevManifest *ProvenanceManifest,
	overwriteDiverged bool,
) *ProvenanceManifest {
	finalEntries := make(map[string]*ProvenanceEntry)

	// fileExists checks whether a manifest entry path exists on disk.
	// Files may live in channelDir (most entries) or knossosDir (e.g. ACTIVE_WORKFLOW.yaml).
	fileExists := func(path string) bool {
		// Normalise directory entries (strip trailing slash before stat)
		cleanPath := strings.TrimSuffix(path, "/")
		if _, err := os.Stat(filepath.Join(channelDir, cleanPath)); err == nil {
			return true
		}
		if knossosDir != "" {
			if _, err := os.Stat(filepath.Join(knossosDir, cleanPath)); err == nil {
				return true
			}
		}
		return false
	}

	// Step 0: Carry forward knossos entries from previous manifest that still exist on disk
	// but weren't re-written this sync (idempotency - files that didn't change)
	if prevManifest != nil {
		for path, entry := range prevManifest.Entries {
			if entry.Owner == OwnerKnossos {
				if fileExists(path) {
					// File/directory exists, carry forward entry (will be overwritten if rewritten in Step 2)
					finalEntries[path] = entry
				}
			}
		}
	}

	// Step 1: Carry forward promoted entries (user-owned + unknown) from divergence detection
	// Skip entries with empty checksums (deleted files) as they fail validation
	if divergenceReport != nil {
		for path, entry := range divergenceReport.Promoted {
			if entry.Checksum != "" {
				finalEntries[path] = entry
			}
		}
		for path, entry := range divergenceReport.CarriedForward {
			if entry.Checksum != "" {
				finalEntries[path] = entry
			}
		}
	}

	// Step 2: Layer current sync entries on top
	// Pipeline-written files take precedence unless path was promoted to user-owned in Step 1.
	// When overwriteDiverged is true, collector entries reclaim user-promoted entries —
	// this matches the --overwrite-diverged flag which overwrites files on disk.
	for path, entry := range collector.Entries() {
		if existing, ok := finalEntries[path]; ok {
			if existing.Owner == OwnerUser && !overwriteDiverged {
				// User promoted this file via divergence detection.
				// Do NOT overwrite with the pipeline entry.
				continue
			}
		}
		finalEntries[path] = entry
	}

	// Step 3: Resolve unknown entries from previous manifest
	// Files with owner:untracked that the pipeline did NOT write this sync are promoted to owner:user
	if prevManifest != nil {
		for path, entry := range prevManifest.Entries {
			if entry.Owner == OwnerUntracked {
				if _, writtenThisSync := collector.Entries()[path]; !writtenThisSync {
					promotedEntry := *entry
					promotedEntry.Owner = OwnerUser
					if _, alreadyInFinal := finalEntries[path]; !alreadyInFinal {
						finalEntries[path] = &promotedEntry
					}
				}
			}
		}
	}

	// Build final manifest
	return &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    activeRite,
		Entries:       finalEntries,
	}
}

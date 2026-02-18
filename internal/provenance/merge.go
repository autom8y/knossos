package provenance

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Merge combines collector entries, divergence report, and previous manifest into a final manifest.
// This is the 4-step merge algorithm that determines provenance ownership in .claude/.
//
// Steps:
//
//	0. Carry forward knossos entries from prev that still exist on disk (idempotency)
//	1. Layer promoted + carried-forward entries from divergence detection (skip empty checksums)
//	2. Layer collector entries (pipeline-written), EXCEPT where divergence promoted to user
//	3. Promote prev untracked entries not written this sync → user
func Merge(
	claudeDir string,
	activeRite string,
	collector Collector,
	divergenceReport *DivergenceReport,
	prevManifest *ProvenanceManifest,
) *ProvenanceManifest {
	finalEntries := make(map[string]*ProvenanceEntry)

	// Step 0: Carry forward knossos entries from previous manifest that still exist on disk
	// but weren't re-written this sync (idempotency - files that didn't change)
	if prevManifest != nil {
		for path, entry := range prevManifest.Entries {
			if entry.Owner == OwnerKnossos {
				// Check if file/directory still exists on disk (not removed)
				fullPath := filepath.Join(claudeDir, path)
				// For directory entries (mena), remove trailing slash before stat
				if strings.HasSuffix(path, "/") {
					fullPath = strings.TrimSuffix(fullPath, "/")
				}
				if _, err := os.Stat(fullPath); err == nil {
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
	// Pipeline-written files take precedence unless path was promoted to user-owned in Step 1
	for path, entry := range collector.Entries() {
		if existing, ok := finalEntries[path]; ok {
			if existing.Owner == OwnerUser {
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

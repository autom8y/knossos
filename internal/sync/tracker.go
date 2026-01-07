// Package sync provides checksum tracking for Ariadne.
package sync

import (
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/knossos/internal/paths"
)

// DefaultTrackedPaths lists the default paths to track for sync.
var DefaultTrackedPaths = []string{
	".claude/CLAUDE.md",
	".claude/manifest.json",
	".claude/ACTIVE_RITE",
	".claude/settings.json",
}

// Tracker handles tracking files for sync.
type Tracker struct {
	resolver *paths.Resolver
	state    *StateManager
}

// NewTracker creates a new tracker.
func NewTracker(resolver *paths.Resolver, state *StateManager) *Tracker {
	return &Tracker{
		resolver: resolver,
		state:    state,
	}
}

// TrackFile adds a file to be tracked.
func (t *Tracker) TrackFile(state *State, path string) error {
	fullPath := filepath.Join(t.resolver.ProjectRoot(), path)

	hash, err := ComputeFileHash(fullPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(fullPath)
	var modTime time.Time
	if err == nil {
		modTime = info.ModTime()
	} else {
		modTime = time.Now().UTC()
	}

	state.TrackedFiles[path] = TrackedFile{
		Path:         path,
		LocalHash:    hash,
		RemoteHash:   hash, // Initially same as local
		BaseHash:     hash, // Base is current state
		LastModified: modTime,
		Status:       "synced",
	}

	return nil
}

// UntrackFile removes a file from tracking.
func (t *Tracker) UntrackFile(state *State, path string) {
	delete(state.TrackedFiles, path)
}

// RefreshAll updates hashes for all tracked files.
func (t *Tracker) RefreshAll(state *State) error {
	for path, tracked := range state.TrackedFiles {
		fullPath := filepath.Join(t.resolver.ProjectRoot(), path)

		hash, err := ComputeFileHash(fullPath)
		if err != nil {
			continue // Skip files that can't be hashed
		}

		if hash == "" {
			// File was deleted
			tracked.Status = "deleted"
			tracked.LocalHash = ""
		} else if hash != tracked.LocalHash {
			// File was modified locally
			tracked.LocalHash = hash
			if tracked.RemoteHash != hash {
				tracked.Status = "modified"
			} else {
				tracked.Status = "synced"
			}
		}

		info, _ := os.Stat(fullPath)
		if info != nil {
			tracked.LastModified = info.ModTime()
		}

		state.TrackedFiles[path] = tracked
	}

	return nil
}

// DetectLocalChanges returns paths that have changed locally.
func (t *Tracker) DetectLocalChanges(state *State) ([]string, error) {
	var changed []string

	for path, tracked := range state.TrackedFiles {
		fullPath := filepath.Join(t.resolver.ProjectRoot(), path)

		hash, err := ComputeFileHash(fullPath)
		if err != nil {
			continue
		}

		if hash != tracked.LocalHash {
			changed = append(changed, path)
		}
	}

	return changed, nil
}

// CompareWithRemote compares local files with remote versions.
// Returns maps of added, modified, and deleted files.
func (t *Tracker) CompareWithRemote(state *State, remoteFiles map[string][]byte) (added, modified, deleted []string) {
	// Check for modified and deleted files
	for path, tracked := range state.TrackedFiles {
		remoteContent, exists := remoteFiles[path]
		if !exists {
			// File deleted in remote
			deleted = append(deleted, path)
			continue
		}

		remoteHash := ComputeContentHash(remoteContent)
		if remoteHash != tracked.RemoteHash {
			// File modified in remote
			modified = append(modified, path)
		}
	}

	// Check for added files
	for path := range remoteFiles {
		if _, exists := state.TrackedFiles[path]; !exists {
			added = append(added, path)
		}
	}

	return added, modified, deleted
}

// DetectConflicts checks for three-way merge conflicts.
// A conflict exists when:
// - Local has changed from base
// - Remote has changed from base
// - Local and remote changes are different
func (t *Tracker) DetectConflicts(state *State, remotePath string, remoteContent []byte) bool {
	tracked, exists := state.TrackedFiles[remotePath]
	if !exists {
		return false
	}

	localPath := filepath.Join(t.resolver.ProjectRoot(), remotePath)
	localHash, _ := ComputeFileHash(localPath)
	remoteHash := ComputeContentHash(remoteContent)

	// If local matches remote, no conflict
	if localHash == remoteHash {
		return false
	}

	// If local hasn't changed from base, no conflict (safe to update)
	if localHash == tracked.BaseHash {
		return false
	}

	// If remote hasn't changed from base, no conflict (local changes take precedence)
	if remoteHash == tracked.BaseHash {
		return false
	}

	// Both local and remote changed differently from base - conflict!
	return true
}

// GetFileStatus returns the sync status for a file.
func (t *Tracker) GetFileStatus(state *State, path string) string {
	tracked, exists := state.TrackedFiles[path]
	if !exists {
		return "untracked"
	}
	return tracked.Status
}

// DiscoverTrackedFiles finds files that should be tracked based on defaults.
func (t *Tracker) DiscoverTrackedFiles() []string {
	var found []string

	for _, path := range DefaultTrackedPaths {
		fullPath := filepath.Join(t.resolver.ProjectRoot(), path)
		if _, err := os.Stat(fullPath); err == nil {
			found = append(found, path)
		}
	}

	return found
}

// UpdateFromRemote updates local hash to match remote after a pull.
func (t *Tracker) UpdateFromRemote(state *State, path string, remoteContent []byte) {
	hash := ComputeContentHash(remoteContent)

	tracked, exists := state.TrackedFiles[path]
	if exists {
		tracked.LocalHash = hash
		tracked.RemoteHash = hash
		tracked.BaseHash = hash // Update base after successful sync
		tracked.Status = "synced"
		tracked.LastModified = time.Now().UTC()
		state.TrackedFiles[path] = tracked
	} else {
		// New file
		state.TrackedFiles[path] = TrackedFile{
			Path:         path,
			LocalHash:    hash,
			RemoteHash:   hash,
			BaseHash:     hash,
			Status:       "synced",
			LastModified: time.Now().UTC(),
		}
	}
}

// MarkPushed marks a file as pushed (update remote hash).
func (t *Tracker) MarkPushed(state *State, path string) {
	tracked, exists := state.TrackedFiles[path]
	if !exists {
		return
	}

	tracked.RemoteHash = tracked.LocalHash
	tracked.BaseHash = tracked.LocalHash // Update base after successful push
	tracked.Status = "synced"
	state.TrackedFiles[path] = tracked
}

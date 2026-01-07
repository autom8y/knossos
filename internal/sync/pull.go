// Package sync provides pull operations for Ariadne.
package sync

import (
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// PullOptions configures pull behavior.
type PullOptions struct {
	Force  bool   // Force overwrite even with conflicts
	DryRun bool   // Don't actually write files
	Paths  []string // Specific paths to pull (empty = all tracked)
}

// PullResult represents the result of a pull operation.
type PullResult struct {
	Remote        string
	FilesUpdated  []FileChange
	FilesConflict []Conflict
	UpdatedCount  int
	ConflictCount int
	Success       bool
}

// FileChange represents a change made during sync.
type FileChange struct {
	Path    string
	Action  string // added, updated, deleted
	OldHash string
	NewHash string
}

// Puller handles pull operations.
type Puller struct {
	resolver *paths.Resolver
	state    *StateManager
	fetcher  *RemoteFetcher
	tracker  *Tracker
}

// NewPuller creates a new puller.
func NewPuller(resolver *paths.Resolver) *Puller {
	stateManager := NewStateManager(resolver)
	return &Puller{
		resolver: resolver,
		state:    stateManager,
		fetcher:  NewRemoteFetcher(),
		tracker:  NewTracker(resolver, stateManager),
	}
}

// Pull performs a pull from the remote.
func (p *Puller) Pull(remoteURL string, opts PullOptions) (*PullResult, error) {
	remote, err := ParseRemote(remoteURL)
	if err != nil {
		return nil, err
	}

	result := &PullResult{
		Remote:        remoteURL,
		FilesUpdated:  []FileChange{},
		FilesConflict: []Conflict{},
		Success:       true,
	}

	// Load or initialize state
	state, err := p.state.Load()
	if err != nil {
		return nil, err
	}

	if state == nil {
		// First pull - initialize state
		state, err = p.state.Initialize(remoteURL)
		if err != nil {
			return nil, err
		}
	}

	// Determine paths to pull
	pullPaths := opts.Paths
	if len(pullPaths) == 0 {
		// If state has tracked files, use those
		if len(state.TrackedFiles) > 0 {
			for path := range state.TrackedFiles {
				pullPaths = append(pullPaths, path)
			}
		} else {
			// Initial pull - discover from both local AND remote (D-004 fix)
			localPaths := p.tracker.DiscoverTrackedFiles()
			pathSet := make(map[string]bool)
			for _, path := range localPaths {
				pathSet[path] = true
			}

			// Also check remote for default tracked paths
			for _, path := range DefaultTrackedPaths {
				exists, _ := p.fetcher.Exists(remote, path)
				if exists {
					pathSet[path] = true
				}
			}

			for path := range pathSet {
				pullPaths = append(pullPaths, path)
			}
		}
	}

	// Fetch and process each path
	for _, path := range pullPaths {
		change, conflict, err := p.pullFile(state, remote, path, opts)
		if err != nil {
			if errors.IsRemoteNotFound(err) {
				// File doesn't exist on remote, skip
				continue
			}
			return nil, err
		}

		if conflict != nil {
			result.FilesConflict = append(result.FilesConflict, *conflict)
			result.ConflictCount++
			result.Success = false
		} else if change != nil {
			result.FilesUpdated = append(result.FilesUpdated, *change)
			result.UpdatedCount++
		}
	}

	// Update state
	state.LastSync = time.Now().UTC()
	state.Remote = remoteURL
	if err := p.state.Save(state); err != nil {
		return nil, err
	}

	return result, nil
}

// pullFile pulls a single file from the remote.
func (p *Puller) pullFile(state *State, remote *Remote, path string, opts PullOptions) (*FileChange, *Conflict, error) {
	// Fetch remote content
	remoteContent, err := p.fetcher.FetchFile(remote, path)
	if err != nil {
		return nil, nil, err
	}

	remoteHash := ComputeContentHash(remoteContent)
	localPath := filepath.Join(p.resolver.ProjectRoot(), path)
	localHash, _ := ComputeFileHash(localPath)

	tracked, exists := state.TrackedFiles[path]

	// Determine action
	action := "updated"
	if !exists || tracked.LocalHash == "" {
		action = "added"
	}

	// Check for conflicts
	if !opts.Force {
		if exists {
			// Tracked file - use three-way merge conflict detection
			// Has local changed from base?
			localChanged := localHash != "" && localHash != tracked.BaseHash

			// Has remote changed from base?
			remoteChanged := remoteHash != tracked.BaseHash

			if localChanged && remoteChanged && localHash != remoteHash {
				// Three-way conflict
				conflict := Conflict{
					Path:        path,
					Description: "Both local and remote have changed",
					LocalHash:   localHash,
					RemoteHash:  remoteHash,
					BaseHash:    tracked.BaseHash,
					DetectedAt:  time.Now().UTC().Format(time.RFC3339),
				}

				// Add to state conflicts
				p.state.AddConflict(state, path, localHash, remoteHash, tracked.BaseHash, conflict.Description)

				return nil, &conflict, nil
			}

			// If local hasn't changed, safe to update
			if !localChanged {
				// No conflict, proceed with update
			} else if !remoteChanged {
				// Remote hasn't changed, local wins - no update needed
				return nil, nil, nil
			}
		} else if localHash != "" && localHash != remoteHash {
			// D-005 fix: Untracked file exists locally with different content
			// This is a potential conflict on first sync - don't silently overwrite
			conflict := Conflict{
				Path:        path,
				Description: "Local file exists with different content (first sync)",
				LocalHash:   localHash,
				RemoteHash:  remoteHash,
				BaseHash:    "", // No base on first sync
				DetectedAt:  time.Now().UTC().Format(time.RFC3339),
			}

			// Add to state conflicts
			p.state.AddConflict(state, path, localHash, remoteHash, "", conflict.Description)

			return nil, &conflict, nil
		}
	}

	// If remote hash matches local, no update needed
	if remoteHash == localHash {
		// Just ensure tracking is up to date
		p.tracker.UpdateFromRemote(state, path, remoteContent)
		return nil, nil, nil
	}

	// Apply the update
	if !opts.DryRun {
		// Ensure directory exists
		dir := filepath.Dir(localPath)
		if err := paths.EnsureDir(dir); err != nil {
			return nil, nil, errors.Wrap(errors.CodeGeneralError, "failed to create directory", err)
		}

		// Write the file
		if err := os.WriteFile(localPath, remoteContent, 0644); err != nil {
			return nil, nil, errors.Wrap(errors.CodeGeneralError, "failed to write file", err)
		}
	}

	// Update tracking
	p.tracker.UpdateFromRemote(state, path, remoteContent)

	change := &FileChange{
		Path:    path,
		Action:  action,
		OldHash: localHash,
		NewHash: remoteHash,
	}

	return change, nil, nil
}

// InitializeTracking sets up initial tracking for a project.
func (p *Puller) InitializeTracking(remoteURL string) (*State, error) {
	remote, err := ParseRemote(remoteURL)
	if err != nil {
		return nil, err
	}

	state, err := p.state.Initialize(remoteURL)
	if err != nil {
		return nil, err
	}

	// Discover and track local files
	for _, path := range p.tracker.DiscoverTrackedFiles() {
		if err := p.tracker.TrackFile(state, path); err != nil {
			continue // Skip files that can't be tracked
		}
	}

	// Also check remote for files
	if remote.Type == RemoteTypeLocal {
		for _, path := range DefaultTrackedPaths {
			exists, _ := p.fetcher.Exists(remote, path)
			if exists {
				if _, tracked := state.TrackedFiles[path]; !tracked {
					// Add placeholder tracking for remote-only files
					state.TrackedFiles[path] = TrackedFile{
						Path:         path,
						LocalHash:    "",
						RemoteHash:   "",
						BaseHash:     "",
						Status:       "untracked",
						LastModified: time.Now().UTC(),
					}
				}
			}
		}
	}

	if err := p.state.Save(state); err != nil {
		return nil, err
	}

	return state, nil
}

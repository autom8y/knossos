// Package sync provides push operations for Ariadne.
package sync

import (
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/paths"
)

// PushOptions configures push behavior.
type PushOptions struct {
	Force  bool     // Force push even if remote has changes
	DryRun bool     // Don't actually push
	Paths  []string // Specific paths to push (empty = all modified)
}

// PushResult represents the result of a push operation.
type PushResult struct {
	Remote       string
	FilesPushed  []FileChange
	PushedCount  int
	Rejected     bool
	RejectReason string
	Success      bool
}

// Pusher handles push operations.
type Pusher struct {
	resolver *paths.Resolver
	state    *StateManager
	fetcher  *RemoteFetcher
	tracker  *Tracker
}

// NewPusher creates a new pusher.
func NewPusher(resolver *paths.Resolver) *Pusher {
	stateManager := NewStateManager(resolver)
	return &Pusher{
		resolver: resolver,
		state:    stateManager,
		fetcher:  NewRemoteFetcher(),
		tracker:  NewTracker(resolver, stateManager),
	}
}

// Push performs a push to the remote.
func (p *Pusher) Push(opts PushOptions) (*PushResult, error) {
	// Load state
	state, err := p.state.Load()
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, errors.New(errors.CodeSyncStateCorrupt, "Sync not initialized. Run 'ari sync pull <remote>' first.")
	}

	// Check for unresolved conflicts
	if state.HasConflicts() && !opts.Force {
		return &PushResult{
			Remote:       state.Remote,
			Rejected:     true,
			RejectReason: "Unresolved conflicts exist. Run 'ari sync resolve' or use --force.",
			Success:      false,
		}, nil
	}

	remote, err := ParseRemote(state.Remote)
	if err != nil {
		return nil, err
	}

	// Only local remotes support push currently
	if remote.Type != RemoteTypeLocal {
		return &PushResult{
			Remote:       state.Remote,
			Rejected:     true,
			RejectReason: "Push only supported for local remotes",
			Success:      false,
		}, nil
	}

	result := &PushResult{
		Remote:      state.Remote,
		FilesPushed: []FileChange{},
		Success:     true,
	}

	// Refresh local state
	if err := p.tracker.RefreshAll(state); err != nil {
		return nil, err
	}

	// Determine paths to push
	pushPaths := opts.Paths
	if len(pushPaths) == 0 {
		// Find all modified files
		for path, tracked := range state.TrackedFiles {
			if tracked.Status == "modified" || tracked.LocalHash != tracked.RemoteHash {
				pushPaths = append(pushPaths, path)
			}
		}
	}

	// Push each file
	for _, path := range pushPaths {
		change, err := p.pushFile(state, remote, path, opts)
		if err != nil {
			if errors.IsRemoteRejected(err) {
				result.Rejected = true
				result.RejectReason = err.Error()
				result.Success = false
				break
			}
			return nil, err
		}

		if change != nil {
			result.FilesPushed = append(result.FilesPushed, *change)
			result.PushedCount++
		}
	}

	// Update state
	state.LastSync = time.Now().UTC()
	if err := p.state.Save(state); err != nil {
		return nil, err
	}

	return result, nil
}

// pushFile pushes a single file to the remote.
func (p *Pusher) pushFile(state *State, remote *Remote, path string, opts PushOptions) (*FileChange, error) {
	localPath := filepath.Join(p.resolver.ProjectRoot(), path)

	// Read local content
	localContent, err := os.ReadFile(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File deleted locally - handle as delete push
			return p.pushDelete(state, remote, path, opts)
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read local file", err)
	}

	localHash := ComputeContentHash(localContent)

	tracked, exists := state.TrackedFiles[path]

	// Check if push is needed
	if exists && localHash == tracked.RemoteHash {
		return nil, nil // Already in sync
	}

	// Check remote for conflicts (if not forcing)
	if exists && !opts.Force && remote.Type == RemoteTypeLocal {
		remoteContent, err := p.fetcher.FetchFile(remote, path)
		if err != nil && !errors.IsRemoteNotFound(err) {
			return nil, err
		}

		if remoteContent != nil {
			remoteHash := ComputeContentHash(remoteContent)
			if remoteHash != tracked.RemoteHash {
				// Remote changed since last sync
				return nil, errors.ErrRemoteRejected(state.Remote, "Remote has newer changes. Pull first or use --force.")
			}
		}
	}

	// Determine action
	action := "updated"
	if !exists {
		action = "added"
	}

	// Perform the push
	if !opts.DryRun {
		remotePath := filepath.Join(remote.URL, path)

		// Ensure directory exists
		dir := filepath.Dir(remotePath)
		if err := paths.EnsureDir(dir); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to create remote directory", err)
		}

		// Write to remote
		if err := os.WriteFile(remotePath, localContent, 0644); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to write to remote", err)
		}

		// Update tracking only if actually pushed (not dry-run)
		p.tracker.MarkPushed(state, path)
	}

	oldHash := ""
	if exists {
		oldHash = tracked.RemoteHash
	}

	return &FileChange{
		Path:    path,
		Action:  action,
		OldHash: oldHash,
		NewHash: localHash,
	}, nil
}

// pushDelete handles pushing a delete to remote.
func (p *Pusher) pushDelete(state *State, remote *Remote, path string, opts PushOptions) (*FileChange, error) {
	tracked, exists := state.TrackedFiles[path]
	if !exists {
		return nil, nil // Not tracked, nothing to do
	}

	if !opts.DryRun {
		remotePath := filepath.Join(remote.URL, path)
		if err := os.Remove(remotePath); err != nil && !os.IsNotExist(err) {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to delete from remote", err)
		}

		// Remove from tracking only if actually deleted (not dry-run)
		delete(state.TrackedFiles, path)
	}

	return &FileChange{
		Path:    path,
		Action:  "deleted",
		OldHash: tracked.RemoteHash,
		NewHash: "",
	}, nil
}

// ListPendingPush returns files that have local changes to push.
func (p *Pusher) ListPendingPush() ([]string, error) {
	state, err := p.state.Load()
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, nil
	}

	// Refresh local state
	if err := p.tracker.RefreshAll(state); err != nil {
		return nil, err
	}

	var pending []string
	for path, tracked := range state.TrackedFiles {
		if tracked.LocalHash != tracked.RemoteHash {
			pending = append(pending, path)
		}
	}

	return pending, nil
}

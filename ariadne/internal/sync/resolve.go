// Package sync provides conflict resolution for Ariadne.
package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/manifest"
	"github.com/autom8y/ariadne/internal/paths"
)

// ResolveStrategy defines how to resolve conflicts.
type ResolveStrategy string

const (
	// ResolveOurs uses local changes.
	ResolveOurs ResolveStrategy = "ours"
	// ResolveTheirs uses remote changes.
	ResolveTheirs ResolveStrategy = "theirs"
	// ResolveMerge attempts three-way merge.
	ResolveMerge ResolveStrategy = "merge"
)

// ResolveOptions configures resolution behavior.
type ResolveOptions struct {
	Strategy ResolveStrategy
	Path     string // Specific path to resolve (empty = all)
	DryRun   bool
}

// ResolveResult represents the result of a resolve operation.
type ResolveResult struct {
	Path           string
	Strategy       string
	Resolved       bool
	RemainingCount int
	Remaining      []string
}

// Resolver handles conflict resolution.
type Resolver struct {
	resolver *paths.Resolver
	state    *StateManager
	fetcher  *RemoteFetcher
	tracker  *Tracker
}

// NewResolver creates a new resolver.
func NewResolver(pathResolver *paths.Resolver) *Resolver {
	stateManager := NewStateManager(pathResolver)
	return &Resolver{
		resolver: pathResolver,
		state:    stateManager,
		fetcher:  NewRemoteFetcher(),
		tracker:  NewTracker(pathResolver, stateManager),
	}
}

// Resolve resolves conflicts.
func (r *Resolver) Resolve(opts ResolveOptions) (*ResolveResult, error) {
	state, err := r.state.Load()
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, errors.New(errors.CodeSyncStateCorrupt, "Sync not initialized")
	}

	result := &ResolveResult{
		Strategy: string(opts.Strategy),
		Resolved: true,
	}

	if opts.Path != "" {
		// Resolve single conflict
		return r.resolveSingle(state, opts)
	}

	// Resolve all conflicts
	for _, conflict := range state.Conflicts {
		singleOpts := ResolveOptions{
			Strategy: opts.Strategy,
			Path:     conflict.Path,
			DryRun:   opts.DryRun,
		}

		_, err := r.resolveSingle(state, singleOpts)
		if err != nil {
			return nil, err
		}
	}

	// Save state
	if !opts.DryRun {
		if err := r.state.Save(state); err != nil {
			return nil, err
		}
	}

	result.RemainingCount = len(state.Conflicts)
	for _, c := range state.Conflicts {
		result.Remaining = append(result.Remaining, c.Path)
	}

	return result, nil
}

// resolveSingle resolves a single conflict.
func (r *Resolver) resolveSingle(state *State, opts ResolveOptions) (*ResolveResult, error) {
	result := &ResolveResult{
		Path:     opts.Path,
		Strategy: string(opts.Strategy),
		Resolved: false,
	}

	conflict := state.GetConflict(opts.Path)
	if conflict == nil {
		return nil, errors.NewWithDetails(errors.CodeFileNotFound,
			"No conflict found for path",
			map[string]interface{}{"path": opts.Path})
	}

	remote, err := ParseRemote(state.Remote)
	if err != nil {
		return nil, err
	}

	localPath := filepath.Join(r.resolver.ProjectRoot(), opts.Path)

	switch opts.Strategy {
	case ResolveOurs:
		// Keep local, update tracking
		localHash, _ := ComputeFileHash(localPath)
		if !opts.DryRun {
			r.state.UpdateTrackedFile(state, opts.Path, localHash, localHash, localHash, "synced")
			r.state.RemoveConflict(state, opts.Path)
		}
		result.Resolved = true

	case ResolveTheirs:
		// Use remote content
		remoteContent, err := r.fetcher.FetchFile(remote, opts.Path)
		if err != nil {
			return nil, err
		}

		if !opts.DryRun {
			// Write remote content to local
			if err := os.WriteFile(localPath, remoteContent, 0644); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to write file", err)
			}
			r.tracker.UpdateFromRemote(state, opts.Path, remoteContent)
			r.state.RemoveConflict(state, opts.Path)
		}
		result.Resolved = true

	case ResolveMerge:
		// Attempt three-way merge using manifest package
		result.Resolved, err = r.attemptMerge(state, remote, opts)
		if err != nil {
			return nil, err
		}
	}

	if !opts.DryRun && result.Resolved {
		if err := r.state.Save(state); err != nil {
			return nil, err
		}
	}

	// Update remaining count
	result.RemainingCount = len(state.Conflicts)
	for _, c := range state.Conflicts {
		result.Remaining = append(result.Remaining, c.Path)
	}

	return result, nil
}

// attemptMerge attempts a three-way merge.
func (r *Resolver) attemptMerge(state *State, remote *Remote, opts ResolveOptions) (bool, error) {
	conflict := state.GetConflict(opts.Path)
	if conflict == nil {
		return false, nil
	}

	localPath := filepath.Join(r.resolver.ProjectRoot(), opts.Path)

	// Read local content
	localContent, err := os.ReadFile(localPath)
	if err != nil {
		return false, errors.Wrap(errors.CodeGeneralError, "failed to read local file", err)
	}

	// Fetch remote content
	remoteContent, err := r.fetcher.FetchFile(remote, opts.Path)
	if err != nil {
		return false, err
	}

	// For JSON files, use the manifest merge logic
	if isJSONFile(opts.Path) {
		return r.mergeJSON(state, opts, localContent, remoteContent)
	}

	// For non-JSON, we can't auto-merge - fall back to ours
	if !opts.DryRun {
		localHash, _ := ComputeFileHash(localPath)
		r.state.UpdateTrackedFile(state, opts.Path, localHash, localHash, localHash, "synced")
		r.state.RemoveConflict(state, opts.Path)
	}

	return true, nil
}

// mergeJSON merges JSON files using three-way merge.
func (r *Resolver) mergeJSON(state *State, opts ResolveOptions, localContent, remoteContent []byte) (bool, error) {
	localPath := filepath.Join(r.resolver.ProjectRoot(), opts.Path)

	// Parse as JSON
	var localData, remoteData, baseData map[string]interface{}

	if err := json.Unmarshal(localContent, &localData); err != nil {
		return false, errors.NewWithDetails(errors.CodeParseError,
			"failed to parse local JSON",
			map[string]interface{}{"path": opts.Path})
	}

	if err := json.Unmarshal(remoteContent, &remoteData); err != nil {
		return false, errors.NewWithDetails(errors.CodeParseError,
			"failed to parse remote JSON",
			map[string]interface{}{"path": opts.Path})
	}

	// For base, we use an empty object if we don't have the original
	// In a full implementation, we'd fetch the base from history
	baseData = make(map[string]interface{})

	// Use manifest merge logic
	baseManifest := &manifest.Manifest{Path: "base", Content: baseData}
	oursManifest := &manifest.Manifest{Path: "ours", Content: localData}
	theirsManifest := &manifest.Manifest{Path: "theirs", Content: remoteData}

	mergeOpts := manifest.MergeOptions{
		Strategy: manifest.StrategySmart,
		DryRun:   opts.DryRun,
	}

	result, err := manifest.Merge(baseManifest, oursManifest, theirsManifest, mergeOpts)
	if err != nil {
		return false, err
	}

	if result.HasConflicts {
		// Still has conflicts, can't auto-resolve
		return false, nil
	}

	if !opts.DryRun {
		// Write merged content
		mergedContent, err := json.MarshalIndent(result.Merged, "", "  ")
		if err != nil {
			return false, errors.Wrap(errors.CodeGeneralError, "failed to marshal merged content", err)
		}

		if err := os.WriteFile(localPath, mergedContent, 0644); err != nil {
			return false, errors.Wrap(errors.CodeGeneralError, "failed to write merged file", err)
		}

		mergedHash := ComputeContentHash(mergedContent)
		r.state.UpdateTrackedFile(state, opts.Path, mergedHash, mergedHash, mergedHash, "synced")
		r.state.RemoveConflict(state, opts.Path)
	}

	return true, nil
}

// isJSONFile checks if a path is a JSON file.
func isJSONFile(path string) bool {
	return filepath.Ext(path) == ".json"
}

// ListConflicts returns all current conflicts.
func (r *Resolver) ListConflicts() ([]Conflict, error) {
	state, err := r.state.Load()
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, nil
	}

	return state.Conflicts, nil
}

// HasConflicts checks if there are any unresolved conflicts.
func (r *Resolver) HasConflicts() (bool, error) {
	state, err := r.state.Load()
	if err != nil {
		return false, err
	}

	if state == nil {
		return false, nil
	}

	return state.HasConflicts(), nil
}

// MarkResolved marks a conflict as resolved without changing files.
func (r *Resolver) MarkResolved(path string) error {
	state, err := r.state.Load()
	if err != nil {
		return err
	}

	if state == nil {
		return errors.New(errors.CodeSyncStateCorrupt, "Sync not initialized")
	}

	if !r.state.RemoveConflict(state, path) {
		return errors.NewWithDetails(errors.CodeFileNotFound,
			"No conflict found for path",
			map[string]interface{}{"path": path})
	}

	// Update tracking
	localPath := filepath.Join(r.resolver.ProjectRoot(), path)
	localHash, _ := ComputeFileHash(localPath)

	tracked, exists := state.TrackedFiles[path]
	if exists {
		tracked.LocalHash = localHash
		tracked.BaseHash = localHash
		tracked.Status = "synced"
		tracked.LastModified = time.Now().UTC()
		state.TrackedFiles[path] = tracked
	}

	return r.state.Save(state)
}

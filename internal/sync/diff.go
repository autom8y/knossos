// Package sync provides diff operations for Ariadne.
package sync

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// DiffOptions configures diff behavior.
type DiffOptions struct {
	Path     string // Specific path to diff (empty = all tracked)
	ShowFull bool   // Show full content, not just summary
}

// DiffResult represents the result of a diff operation.
type DiffResult struct {
	Path          string
	HasChanges    bool
	LocalContent  string
	RemoteContent string
	UnifiedDiff   string
	Additions     int
	Deletions     int
	TotalFiles    int
	ChangedFiles  int
	FileResults   []FileDiffResult
}

// FileDiffResult represents diff for a single file.
type FileDiffResult struct {
	Path       string
	HasChanges bool
	Status     string // synced, modified, added, deleted
	LocalHash  string
	RemoteHash string
}

// Differ handles diff operations.
type Differ struct {
	resolver *paths.Resolver
	state    *StateManager
	fetcher  *RemoteFetcher
	tracker  *Tracker
}

// NewDiffer creates a new differ.
func NewDiffer(resolver *paths.Resolver) *Differ {
	stateManager := NewStateManager(resolver)
	return &Differ{
		resolver: resolver,
		state:    stateManager,
		fetcher:  NewRemoteFetcher(),
		tracker:  NewTracker(resolver, stateManager),
	}
}

// Diff computes differences between local and remote.
func (d *Differ) Diff(opts DiffOptions) (*DiffResult, error) {
	state, err := d.state.Load()
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, errors.New(errors.CodeSyncStateCorrupt, "Sync not initialized. Run 'ari sync pull <remote>' first.")
	}

	remote, err := ParseRemote(state.Remote)
	if err != nil {
		return nil, err
	}

	result := &DiffResult{
		FileResults: []FileDiffResult{},
	}

	// Refresh local state
	if err := d.tracker.RefreshAll(state); err != nil {
		return nil, err
	}

	// Single file diff
	if opts.Path != "" {
		return d.diffSingleFile(state, remote, opts.Path, opts.ShowFull)
	}

	// All files diff
	for path, tracked := range state.TrackedFiles {
		fileResult := FileDiffResult{
			Path:       path,
			LocalHash:  tracked.LocalHash,
			RemoteHash: tracked.RemoteHash,
		}

		localPath := filepath.Join(d.resolver.ProjectRoot(), path)
		currentHash, _ := ComputeFileHash(localPath)

		// Check local changes
		if currentHash != tracked.RemoteHash {
			fileResult.HasChanges = true
			result.ChangedFiles++
			result.HasChanges = true

			if currentHash == "" {
				fileResult.Status = "deleted"
			} else if tracked.RemoteHash == "" {
				fileResult.Status = "added"
			} else {
				fileResult.Status = "modified"
			}
		} else {
			fileResult.Status = "synced"
		}

		result.FileResults = append(result.FileResults, fileResult)
		result.TotalFiles++
	}

	return result, nil
}

// diffSingleFile computes detailed diff for a single file.
func (d *Differ) diffSingleFile(state *State, remote *Remote, path string, showFull bool) (*DiffResult, error) {
	result := &DiffResult{
		Path:        path,
		TotalFiles:  1,
		FileResults: []FileDiffResult{},
	}

	localPath := filepath.Join(d.resolver.ProjectRoot(), path)

	// Read local content
	localContent, err := os.ReadFile(localPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read local file", err)
	}

	// Fetch remote content
	remoteContent, err := d.fetcher.FetchFile(remote, path)
	if err != nil && !errors.IsRemoteNotFound(err) {
		return nil, err
	}

	localStr := string(localContent)
	remoteStr := string(remoteContent)

	if localStr == remoteStr {
		result.HasChanges = false
		return result, nil
	}

	result.HasChanges = true
	result.ChangedFiles = 1

	if showFull {
		result.LocalContent = localStr
		result.RemoteContent = remoteStr
	}

	// Generate unified diff
	result.UnifiedDiff = generateUnifiedDiff(path, remoteStr, localStr)

	// Count additions and deletions
	result.Additions, result.Deletions = countDiffStats(remoteStr, localStr)

	return result, nil
}

// generateUnifiedDiff creates a unified diff output.
func generateUnifiedDiff(path, oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	var b strings.Builder
	b.WriteString("--- a/" + path + " (remote)\n")
	b.WriteString("+++ b/" + path + " (local)\n")

	// Simple diff - in production use a proper diff algorithm
	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	inHunk := false
	hunkStart := 0

	for i := 0; i < maxLines; i++ {
		oldLine := ""
		newLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			if !inHunk {
				// Start new hunk
				hunkStart = i + 1
				b.WriteString("@@ " + formatHunkHeader(hunkStart, oldLines, newLines) + " @@\n")
				inHunk = true
			}

			if oldLine != "" && (i >= len(newLines) || oldLine != newLine) {
				b.WriteString("-" + oldLine + "\n")
			}
			if newLine != "" && (i >= len(oldLines) || oldLine != newLine) {
				b.WriteString("+" + newLine + "\n")
			}
		} else if inHunk {
			// Context line
			b.WriteString(" " + oldLine + "\n")
		}
	}

	return b.String()
}

// formatHunkHeader formats a hunk header.
func formatHunkHeader(start int, oldLines, newLines []string) string {
	oldLen := len(oldLines)
	newLen := len(newLines)
	return "-" + itoa(start) + "," + itoa(oldLen) + " +" + itoa(start) + "," + itoa(newLen)
}

// itoa is a simple int to string converter.
func itoa(i int) string {
	return strconv.Itoa(i)
}

// countDiffStats counts additions and deletions.
func countDiffStats(oldContent, newContent string) (additions, deletions int) {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	oldSet := make(map[string]int)
	for _, line := range oldLines {
		oldSet[line]++
	}

	newSet := make(map[string]int)
	for _, line := range newLines {
		newSet[line]++
	}

	for line, count := range newSet {
		oldCount := oldSet[line]
		if count > oldCount {
			additions += count - oldCount
		}
	}

	for line, count := range oldSet {
		newCount := newSet[line]
		if count > newCount {
			deletions += count - newCount
		}
	}

	return additions, deletions
}

// QuickCheck performs a quick check for any changes.
func (d *Differ) QuickCheck() (bool, error) {
	state, err := d.state.Load()
	if err != nil {
		return false, err
	}

	if state == nil {
		return false, nil
	}

	for path, tracked := range state.TrackedFiles {
		localPath := filepath.Join(d.resolver.ProjectRoot(), path)
		currentHash, _ := ComputeFileHash(localPath)

		if currentHash != tracked.LocalHash {
			return true, nil
		}
	}

	return false, nil
}

package know

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ChangeManifest describes what changed between two git states within source scope.
type ChangeManifest struct {
	FromHash      string        `json:"from_hash"`
	ToHash        string        `json:"to_hash"`
	NewFiles      []string      `json:"new_files"`
	ModifiedFiles []string      `json:"modified_files"`
	DeletedFiles  []string      `json:"deleted_files"`
	RenamedFiles  []RenamedFile `json:"renamed_files"`
	CommitLog     string        `json:"commit_log"`
	DeltaLines    int           `json:"delta_lines"`
	DeltaRatio    float64       `json:"delta_ratio"`
	TotalFiles    int           `json:"total_files"`
}

// RenamedFile represents a file that was renamed between two git states.
type RenamedFile struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

// gitDiffFiltered gets files matching a specific diff filter (A=added, D=deleted, M=modified, R=renamed).
// Replaceable in tests to avoid real git invocations.
var gitDiffFiltered = defaultGitDiffFiltered

// gitDiffNumstat gets added/deleted line counts between two commits.
// Replaceable in tests to avoid real git invocations.
var gitDiffNumstat = defaultGitDiffNumstat

// gitLogOneline gets compact commit log between two hashes.
// Replaceable in tests to avoid real git invocations.
var gitLogOneline = defaultGitLogOneline

// gitLsFiles counts tracked files matching scope patterns.
// Replaceable in tests to avoid real git invocations.
var gitLsFiles = defaultGitLsFiles

// gitDiffNameStatus gets tab-separated name-status output for a specific diff filter.
// Used internally for rename detection (R\toldpath\tnewpath format).
// Replaceable in tests to avoid real git invocations.
var gitDiffNameStatus = defaultGitDiffNameStatus

func defaultGitDiffFiltered(fromHash, toHash, filter string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "diff", "--name-only", "--diff-filter="+filter, fromHash+".."+toHash).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

func defaultGitDiffNumstat(fromHash, toHash string) (added int, deleted int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, execErr := exec.CommandContext(ctx, "git", "diff", "--numstat", fromHash+".."+toHash).Output()
	if execErr != nil {
		return 0, 0, execErr
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// numstat format: "<added>\t<deleted>\t<filename>"
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		// Binary files show "-" instead of numbers; skip those.
		a, aErr := strconv.Atoi(parts[0])
		d, dErr := strconv.Atoi(parts[1])
		if aErr != nil || dErr != nil {
			continue
		}
		added += a
		deleted += d
	}
	return added, deleted, nil
}

func defaultGitLogOneline(fromHash, toHash string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "log", "--oneline", fromHash+".."+toHash).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func defaultGitLsFiles(patterns []string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "ls-files").Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}

	// If no scope patterns, count all tracked files.
	if len(patterns) == 0 {
		return len(lines), nil
	}

	// Filter tracked files by scope patterns using matchScope (same function as scopedStaleness).
	count := 0
	for _, f := range lines {
		if f == "" {
			continue
		}
		for _, pattern := range patterns {
			if matchScope(pattern, f) {
				count++
				break // count each file at most once
			}
		}
	}
	return count, nil
}

func defaultGitDiffNameStatus(fromHash, toHash string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "diff", "--name-status", "--diff-filter=R", fromHash+".."+toHash).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

// ComputeChangeManifest computes what changed between fromHash and toHash,
// optionally filtered by sourceScope glob patterns.
//
// Returns an empty manifest (no files in any list) when fromHash == toHash.
// Returns an error if any git command fails — callers should fall back to full mode.
// When sourceScope is empty, all changed files are included without filtering.
func ComputeChangeManifest(fromHash, toHash string, sourceScope []string) (*ChangeManifest, error) {
	manifest := &ChangeManifest{
		FromHash: fromHash,
		ToHash:   toHash,
	}

	// Short-circuit: identical hashes mean no changes.
	if fromHash == toHash {
		return manifest, nil
	}

	// scopeFilter wraps matchScope for a file against the source scope list.
	// Returns true if no scope defined (include all) or if any pattern matches.
	inScope := func(path string) bool {
		if len(sourceScope) == 0 {
			return true
		}
		for _, pattern := range sourceScope {
			if matchScope(pattern, path) {
				return true
			}
		}
		return false
	}

	filterFiles := func(files []string) []string {
		if len(files) == 0 {
			return nil
		}
		var result []string
		for _, f := range files {
			if f != "" && inScope(f) {
				result = append(result, f)
			}
		}
		return result
	}

	// Added files.
	added, err := gitDiffFiltered(fromHash, toHash, "A")
	if err != nil {
		return nil, fmt.Errorf("git diff added: %w", err)
	}
	manifest.NewFiles = filterFiles(added)

	// Modified files.
	modified, err := gitDiffFiltered(fromHash, toHash, "M")
	if err != nil {
		return nil, fmt.Errorf("git diff modified: %w", err)
	}
	manifest.ModifiedFiles = filterFiles(modified)

	// Deleted files.
	deleted, err := gitDiffFiltered(fromHash, toHash, "D")
	if err != nil {
		return nil, fmt.Errorf("git diff deleted: %w", err)
	}
	manifest.DeletedFiles = filterFiles(deleted)

	// Renamed files: parse name-status output (tab-separated: R<score>\toldpath\tnewpath).
	renameLines, err := gitDiffNameStatus(fromHash, toHash)
	if err != nil {
		return nil, fmt.Errorf("git diff renames: %w", err)
	}
	for _, line := range renameLines {
		if line == "" {
			continue
		}
		// Format: "R100\told/path\tnew/path" (similarity score may vary)
		parts := strings.Split(line, "\t")
		if len(parts) < 3 || !strings.HasPrefix(parts[0], "R") {
			continue
		}
		oldPath := parts[1]
		newPath := parts[2]
		// Apply scope filter on the new path (what the file is now called).
		if inScope(newPath) {
			manifest.RenamedFiles = append(manifest.RenamedFiles, RenamedFile{
				OldPath: oldPath,
				NewPath: newPath,
			})
		}
	}

	// Commit log summary.
	log, err := gitLogOneline(fromHash, toHash)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	manifest.CommitLog = log

	// Delta line count: sum added+deleted lines from numstat.
	// We count all lines in the diff range, not just scoped lines, because
	// numstat parsing by file would require a second pass. This is a best-effort
	// signal used for threshold comparisons; minor over-counting is acceptable.
	addedLines, deletedLines, err := gitDiffNumstat(fromHash, toHash)
	if err != nil {
		return nil, fmt.Errorf("git diff numstat: %w", err)
	}
	manifest.DeltaLines = addedLines + deletedLines

	// Total files in scope (used for DeltaRatio denominator).
	totalFiles, err := gitLsFiles(sourceScope)
	if err != nil {
		// Non-fatal: fall back to ratio 1.0 to force full mode.
		totalFiles = 0
	}
	manifest.TotalFiles = totalFiles

	// DeltaRatio = changed_files / total_files.
	// When TotalFiles == 0, set to 1.0 to force full mode (cannot divide).
	changedCount := len(manifest.NewFiles) + len(manifest.ModifiedFiles) +
		len(manifest.DeletedFiles) + len(manifest.RenamedFiles)
	if totalFiles == 0 {
		manifest.DeltaRatio = 1.0
	} else {
		manifest.DeltaRatio = float64(changedCount) / float64(totalFiles)
	}

	return manifest, nil
}

// FilterChangeManifest returns a copy of the manifest with file lists filtered
// by sourceScope. If sourceScope is empty, returns the original manifest unchanged.
// DeltaRatio is recalculated based on filtered file counts. DeltaLines and CommitLog
// are preserved as-is (they are global signals, not scope-filtered).
func FilterChangeManifest(m *ChangeManifest, sourceScope []string) *ChangeManifest {
	if len(sourceScope) == 0 || m == nil {
		return m
	}

	inScope := func(path string) bool {
		for _, pattern := range sourceScope {
			if matchScope(pattern, path) {
				return true
			}
		}
		return false
	}

	filterFiles := func(files []string) []string {
		var result []string
		for _, f := range files {
			if f != "" && inScope(f) {
				result = append(result, f)
			}
		}
		return result
	}

	filtered := &ChangeManifest{
		FromHash:      m.FromHash,
		ToHash:        m.ToHash,
		NewFiles:      filterFiles(m.NewFiles),
		ModifiedFiles: filterFiles(m.ModifiedFiles),
		DeletedFiles:  filterFiles(m.DeletedFiles),
		CommitLog:     m.CommitLog,
		DeltaLines:    m.DeltaLines,
		TotalFiles:    m.TotalFiles,
	}

	for _, r := range m.RenamedFiles {
		if inScope(r.NewPath) {
			filtered.RenamedFiles = append(filtered.RenamedFiles, r)
		}
	}

	changedCount := len(filtered.NewFiles) + len(filtered.ModifiedFiles) +
		len(filtered.DeletedFiles) + len(filtered.RenamedFiles)
	if filtered.TotalFiles == 0 {
		filtered.DeltaRatio = 1.0
	} else {
		filtered.DeltaRatio = float64(changedCount) / float64(filtered.TotalFiles)
	}

	return filtered
}

// RecommendedMode returns the recommended update mode based on the change manifest
// and domain metadata.
//
// Return values:
//   - "skip"        — never returned by this function (reserved for callers)
//   - "time-only"   — manifest is nil or has no file changes; only timestamp needs bumping
//   - "full"        — delta is too large, cycle limit hit, or manifest is nil with no changes
//   - "incremental" — delta is small enough to update incrementally
func RecommendedMode(manifest *ChangeManifest, meta *Meta) string {
	// Nil manifest or no file changes at all: only timestamps need updating.
	if manifest == nil {
		return "time-only"
	}

	hasChanges := len(manifest.NewFiles) > 0 ||
		len(manifest.ModifiedFiles) > 0 ||
		len(manifest.DeletedFiles) > 0 ||
		len(manifest.RenamedFiles) > 0

	if !hasChanges {
		return "time-only"
	}

	// Cycle limit: if we've been doing incremental updates too long, force a full rebuild.
	if meta != nil && meta.MaxIncrementalCycles > 0 && meta.IncrementalCycle >= meta.MaxIncrementalCycles {
		return "full"
	}

	// Large delta thresholds: too many files or too many line changes.
	if manifest.DeltaRatio >= 0.5 || manifest.DeltaLines >= 5000 {
		return "full"
	}

	return "incremental"
}

package worktree

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
)

// SyncResult provides detailed status about worktree synchronization with upstream.
type SyncResult struct {
	Worktree  Worktree `json:"worktree"`
	Ahead     int      `json:"ahead"`      // Commits ahead of upstream
	Behind    int      `json:"behind"`     // Commits behind upstream
	Diverged  bool     `json:"diverged"`   // Both ahead and behind
	UpToDate  bool     `json:"up_to_date"` // Exactly at upstream
	Conflicts []string `json:"conflicts,omitempty"`
	Pulled    bool     `json:"pulled"`    // If pull was performed
	PullError string   `json:"pull_error,omitempty"`
}

// WorktreeSwitchOptions specifies options for switching to a worktree.
type WorktreeSwitchOptions struct {
	UpdateRite    bool // Update ACTIVE_RITE to match worktree's rite
	CreateSession bool // Create a new session in the worktree
}

// CloneOptions specifies options for cloning a worktree.
type CloneOptions struct {
	Rite        string // Override rite (empty = copy from source)
	CopySession bool   // Copy session context from source
}

// WorktreeSyncOptions specifies options for syncing a worktree.
type WorktreeSyncOptions struct {
	Pull bool // Actually pull changes (vs. just check status)
}

// ExportArchive contains metadata about an exported worktree archive.
type ExportArchive struct {
	WorktreeID string    `json:"worktree_id"`
	Name       string    `json:"name"`
	Rite       string    `json:"rite"`
	Complexity string    `json:"complexity"`
	FromRef    string    `json:"from_ref"`
	GitRef     string    `json:"git_ref"` // HEAD commit at export time
	ExportedAt time.Time `json:"exported_at"`
	Version    string    `json:"version"` // Archive format version
}

const (
	archiveMetaFile = ".worktree-export.json"
	archiveVersion  = "1.0"
)

// Switch updates the session context to point to a different worktree.
// This enables switching context without changing the shell's working directory.
func (m *Manager) Switch(idOrName string, opts WorktreeSwitchOptions) (*Worktree, error) {
	// Resolve worktree
	wt, err := m.resolveWorktree(idOrName)
	if err != nil {
		return nil, err
	}

	// Verify worktree still exists
	if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
		return nil, errors.NewWithDetails(errors.CodeFileNotFound,
			"worktree path no longer exists",
			map[string]interface{}{
				"worktree_id": wt.ID,
				"path":        wt.Path,
			})
	}

	// Update ACTIVE_RITE if requested and rite differs
	if opts.UpdateRite && wt.Rite != "" {
		activeRitePath := filepath.Join(wt.Path, ".claude", "ACTIVE_RITE")
		if err := os.MkdirAll(filepath.Dir(activeRitePath), 0755); err == nil {
			os.WriteFile(activeRitePath, []byte(wt.Rite+"\n"), 0644)
		}

		// Also try to run swap-rite.sh if available
		knossosHome := config.KnossosHome()
		if knossosHome != "" {
			swapRitePath := filepath.Join(knossosHome, "swap-rite.sh")
			if _, err := os.Stat(swapRitePath); err == nil {
				cmd := exec.Command(swapRitePath, wt.Rite)
				cmd.Dir = wt.Path
				cmd.Run() // Ignore errors
			}
		}
	}

	return wt, nil
}

// Clone creates a new worktree as a copy of an existing one, preserving metadata.
func (m *Manager) Clone(sourceIDOrName, newName string, opts CloneOptions) (*Worktree, error) {
	// Prevent nested worktree creation
	if m.git.IsWorktree() {
		return nil, errors.New(errors.CodeGeneralError,
			"Cannot clone worktree from within a worktree. Navigate to main project first.")
	}

	// Resolve source worktree
	source, err := m.resolveWorktree(sourceIDOrName)
	if err != nil {
		return nil, errors.NewWithDetails(errors.CodeFileNotFound,
			"source worktree not found",
			map[string]interface{}{"source": sourceIDOrName})
	}

	// Get current HEAD of source worktree
	sourceHead, err := m.git.GetHead(source.Path)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to get source HEAD", err)
	}

	// Determine rite
	rite := opts.Rite
	if rite == "" {
		rite = source.Rite
	}

	// Generate new worktree ID
	id := GenerateWorktreeID()
	worktreesDir, err := m.git.GetWorktreesDir()
	if err != nil {
		return nil, err
	}
	wtPath := filepath.Join(worktreesDir, id)

	// Create git worktree from source HEAD
	if err := m.git.WorktreeAdd(wtPath, sourceHead, true); err != nil {
		return nil, err
	}

	// Create new worktree record with copied metadata
	wt := Worktree{
		ID:         id,
		Name:       newName,
		Path:       wtPath,
		Rite:       rite,
		CreatedAt:  time.Now().UTC(),
		BaseBranch: source.BaseBranch,
		FromRef:    sourceHead, // Created from source's HEAD
		Complexity: source.Complexity,
	}

	// Save per-worktree metadata
	if err := SavePerWorktreeMeta(wtPath, wt, m.rootDir); err != nil {
		m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Add to registry
	if err := m.metadata.Add(wt); err != nil {
		m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Copy session context if requested
	if opts.CopySession {
		if err := copySessionContext(source.Path, wtPath); err != nil {
			// Non-fatal, just log
		}
	}

	// Try to run roster-sync and swap-rite
	m.setupWorktreeEcosystem(wtPath, rite)

	return &wt, nil
}

// Sync checks synchronization status between worktree and upstream,
// optionally pulling changes.
func (m *Manager) Sync(idOrName string, opts WorktreeSyncOptions) (*SyncResult, error) {
	// Resolve worktree
	wt, err := m.resolveWorktree(idOrName)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		Worktree: *wt,
	}

	// Fetch from remote to update refs
	if err := m.gitFetch(wt.Path); err != nil {
		// Non-fatal, continue with local comparison
	}

	// Get sync status against base branch
	ahead, behind, err := m.git.GetCommitDiff(wt.Path, wt.BaseBranch)
	if err != nil {
		// Try against origin/base
		ahead, behind, _ = m.git.GetCommitDiff(wt.Path, "origin/"+wt.BaseBranch)
	}

	result.Ahead = ahead
	result.Behind = behind
	result.Diverged = ahead > 0 && behind > 0
	result.UpToDate = ahead == 0 && behind == 0

	// If pull requested and behind, try to pull
	if opts.Pull && behind > 0 {
		pullErr := m.gitPull(wt.Path)
		if pullErr != nil {
			result.PullError = pullErr.Error()
			// Check for merge conflicts
			conflicts := m.detectConflicts(wt.Path)
			if len(conflicts) > 0 {
				result.Conflicts = conflicts
			}
		} else {
			result.Pulled = true
			// Re-check status after pull
			ahead, behind, _ = m.git.GetCommitDiff(wt.Path, wt.BaseBranch)
			result.Ahead = ahead
			result.Behind = behind
			result.UpToDate = ahead == 0 && behind == 0
			result.Diverged = ahead > 0 && behind > 0
		}
	}

	return result, nil
}

// Export creates a tar.gz archive of a worktree including metadata.
func (m *Manager) Export(idOrName, targetPath string) error {
	// Resolve worktree
	wt, err := m.resolveWorktree(idOrName)
	if err != nil {
		return err
	}

	// Get current HEAD
	gitRef, err := m.git.GetHead(wt.Path)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to get worktree HEAD", err)
	}

	// Create export metadata
	meta := ExportArchive{
		WorktreeID: wt.ID,
		Name:       wt.Name,
		Rite:       wt.Rite,
		Complexity: wt.Complexity,
		FromRef:    wt.FromRef,
		GitRef:     gitRef,
		ExportedAt: time.Now().UTC(),
		Version:    archiveVersion,
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create target directory", err)
	}

	// Create archive file
	archiveFile, err := os.Create(targetPath)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create archive", err)
	}
	defer archiveFile.Close()

	gzWriter := gzip.NewWriter(archiveFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Write metadata first
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal metadata", err)
	}

	metaHeader := &tar.Header{
		Name:    archiveMetaFile,
		Size:    int64(len(metaData)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tarWriter.WriteHeader(metaHeader); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write metadata header", err)
	}
	if _, err := tarWriter.Write(metaData); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write metadata", err)
	}

	// Walk worktree directory and add files
	err = filepath.Walk(wt.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory (it's a file in worktrees pointing to main repo)
		if info.Name() == ".git" && !info.IsDir() {
			return nil
		}

		// Create relative path
		relPath, err := filepath.Rel(wt.Path, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Join("worktree", relPath)

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Linkname = link
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		if !info.IsDir() && info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to archive worktree", err)
	}

	return nil
}

// Import creates a worktree from an exported archive.
func (m *Manager) Import(archivePath string) (*Worktree, error) {
	// Prevent nested worktree creation
	if m.git.IsWorktree() {
		return nil, errors.New(errors.CodeGeneralError,
			"Cannot import worktree from within a worktree. Navigate to main project first.")
	}

	// Open archive
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to open archive", err)
	}
	defer archiveFile.Close()

	gzReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to read gzip archive", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// Read metadata first
	var meta ExportArchive
	metaFound := false

	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "worktree-import-*")
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create temp directory", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(errors.CodeParseError, "failed to read archive", err)
		}

		// Handle metadata file
		if header.Name == archiveMetaFile {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, errors.Wrap(errors.CodeParseError, "failed to read metadata", err)
			}
			if err := json.Unmarshal(data, &meta); err != nil {
				return nil, errors.Wrap(errors.CodeParseError, "failed to parse metadata", err)
			}
			metaFound = true
			continue
		}

		// Extract to temp directory
		targetPath := filepath.Join(tempDir, header.Name)
		targetDir := filepath.Dir(targetPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to create directory", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to create directory", err)
			}
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to create file", err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to write file", err)
			}
			file.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to create directory", err)
			}
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to create symlink", err)
			}
		}
	}

	if !metaFound {
		return nil, errors.New(errors.CodeParseError, "archive missing metadata file")
	}

	// Validate archive version
	if meta.Version != archiveVersion {
		return nil, errors.NewWithDetails(errors.CodeParseError,
			"unsupported archive version",
			map[string]interface{}{
				"expected": archiveVersion,
				"actual":   meta.Version,
			})
	}

	// Verify git ref exists
	if !m.git.RefExists(meta.GitRef) {
		// Try to fetch it
		m.gitFetch(m.rootDir)
		if !m.git.RefExists(meta.GitRef) {
			return nil, errors.NewWithDetails(errors.CodeGeneralError,
				"git ref from archive not found",
				map[string]interface{}{"ref": meta.GitRef})
		}
	}

	// Generate new worktree ID
	id := GenerateWorktreeID()
	worktreesDir, err := m.git.GetWorktreesDir()
	if err != nil {
		return nil, err
	}
	wtPath := filepath.Join(worktreesDir, id)

	// Create git worktree at the archived ref
	if err := m.git.WorktreeAdd(wtPath, meta.GitRef, true); err != nil {
		return nil, err
	}

	// Copy extracted files (skip .git)
	extractedWorktree := filepath.Join(tempDir, "worktree")
	if _, err := os.Stat(extractedWorktree); err == nil {
		err = filepath.Walk(extractedWorktree, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(extractedWorktree, path)
			if err != nil {
				return err
			}
			if relPath == "." {
				return nil
			}

			// Skip .git file
			if info.Name() == ".git" {
				return nil
			}

			destPath := filepath.Join(wtPath, relPath)

			if info.IsDir() {
				return os.MkdirAll(destPath, info.Mode())
			}

			// Copy file
			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			return err
		})
		if err != nil {
			m.git.WorktreeRemove(wtPath, true)
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to restore files", err)
		}
	}

	// Create worktree record
	wt := Worktree{
		ID:         id,
		Name:       meta.Name,
		Path:       wtPath,
		Rite:       meta.Rite,
		CreatedAt:  time.Now().UTC(),
		BaseBranch: m.git.GetDefaultBranch(),
		FromRef:    meta.GitRef,
		Complexity: meta.Complexity,
	}

	// Update per-worktree metadata with new ID
	if err := SavePerWorktreeMeta(wtPath, wt, m.rootDir); err != nil {
		m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Add to registry
	if err := m.metadata.Add(wt); err != nil {
		m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Setup ecosystem
	m.setupWorktreeEcosystem(wtPath, wt.Rite)

	return &wt, nil
}

// resolveWorktree looks up a worktree by ID or name.
func (m *Manager) resolveWorktree(idOrName string) (*Worktree, error) {
	var wt *Worktree
	var err error

	if IsValidWorktreeID(idOrName) {
		wt, err = m.metadata.Get(idOrName)
	} else {
		wt, err = m.metadata.GetByName(idOrName)
	}

	if err != nil {
		return nil, err
	}

	return wt, nil
}

// gitFetch runs git fetch in the specified directory.
func (m *Manager) gitFetch(path string) error {
	cmd := exec.Command("git", "fetch", "--all", "--prune")
	cmd.Dir = path
	return cmd.Run()
}

// gitPull runs git pull in the specified directory.
func (m *Manager) gitPull(path string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = path
	return cmd.Run()
}

// detectConflicts checks for merge conflicts in the working directory.
func (m *Manager) detectConflicts(path string) []string {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var conflicts []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			conflicts = append(conflicts, line)
		}
	}
	return conflicts
}

// copySessionContext copies session context from source to target worktree.
func copySessionContext(sourcePath, targetPath string) error {
	sourceSessionsDir := filepath.Join(sourcePath, ".claude", "sessions")
	targetSessionsDir := filepath.Join(targetPath, ".claude", "sessions")

	entries, err := os.ReadDir(sourceSessionsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "session-") {
			continue
		}

		// Copy session directory
		sourceDir := filepath.Join(sourceSessionsDir, entry.Name())
		targetDir := filepath.Join(targetSessionsDir, entry.Name())

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			continue
		}

		// Copy SESSION_CONTEXT.md
		contextFile := "SESSION_CONTEXT.md"
		sourceContext := filepath.Join(sourceDir, contextFile)
		targetContext := filepath.Join(targetDir, contextFile)

		data, err := os.ReadFile(sourceContext)
		if err == nil {
			os.WriteFile(targetContext, data, 0644)
		}
	}

	return nil
}

// setupWorktreeEcosystem runs roster-sync and swap-rite for a worktree.
func (m *Manager) setupWorktreeEcosystem(wtPath, rite string) {
	knossosHome := config.KnossosHome()
	if knossosHome == "" {
		return
	}

	// Run roster-sync init or sync
	syncPath := filepath.Join(knossosHome, "roster-sync")
	if _, err := os.Stat(syncPath); err == nil {
		manifestPath := filepath.Join(wtPath, ".claude", ".cem", "manifest.json")
		if _, err := os.Stat(manifestPath); err == nil {
			cmd := exec.Command(syncPath, "sync")
			cmd.Dir = wtPath
			cmd.Run()
		} else {
			cmd := exec.Command(syncPath, "init")
			cmd.Dir = wtPath
			cmd.Run()
		}
	}

	// Run swap-rite if rite specified
	if rite != "" && rite != "none" {
		swapRitePath := filepath.Join(knossosHome, "swap-rite.sh")
		if _, err := os.Stat(swapRitePath); err == nil {
			cmd := exec.Command(swapRitePath, rite)
			cmd.Dir = wtPath
			cmd.Run()
		}
	}
}

package worktree

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// Manager handles worktree lifecycle operations.
type Manager struct {
	git      *GitOperations
	metadata *MetadataManager
	rootDir  string // Main project root
}

// NewManager creates a new worktree Manager.
func NewManager(workDir string) (*Manager, error) {
	git := NewGitOperations(workDir)

	if !git.IsGitRepo() {
		return nil, errors.New(errors.CodeProjectNotFound, "not a git repository")
	}

	// Get main worktree root
	rootDir, err := git.GetMainWorktree()
	if err != nil {
		return nil, err
	}

	// Initialize worktrees directory
	worktreesDir := filepath.Join(rootDir, ".worktrees")
	metadata := NewMetadataManager(worktreesDir)

	return &Manager{
		git:      git,
		metadata: metadata,
		rootDir:  rootDir,
	}, nil
}

// Create creates a new worktree with the given options.
func (m *Manager) Create(opts CreateOptions) (*Worktree, error) {
	// Prevent nested worktree creation
	if m.git.IsWorktree() {
		return nil, errors.New(errors.CodeGeneralError,
			"Cannot create worktree from within a worktree. Navigate to main project first.")
	}

	// Set defaults
	if opts.Name == "" {
		opts.Name = "unnamed"
	}
	if opts.FromRef == "" {
		opts.FromRef = "HEAD"
	}
	if opts.Complexity == "" {
		opts.Complexity = "MODULE"
	}

	// Validate ref exists
	if !m.git.RefExists(opts.FromRef) {
		return nil, errors.NewWithDetails(errors.CodeGeneralError,
			"invalid git ref",
			map[string]any{"ref": opts.FromRef})
	}

	// Generate worktree ID
	id := GenerateWorktreeID()
	worktreesDir, err := m.git.GetWorktreesDir()
	if err != nil {
		return nil, err
	}
	wtPath := filepath.Join(worktreesDir, id)

	// Ensure worktrees directory exists with .gitignore
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create worktrees directory", err)
	}
	gitignorePath := filepath.Join(worktreesDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		_ = os.WriteFile(gitignorePath, []byte("*\n"), 0644)
	}

	// Create git worktree with detached HEAD
	if err := m.git.WorktreeAdd(wtPath, opts.FromRef, true); err != nil {
		return nil, err
	}

	// Get rite from flag or detect from ACTIVE_RITE
	rite := opts.Rite
	if rite == "" {
		rite = paths.NewResolver(m.rootDir).ReadActiveRite()
	}

	// Determine base branch
	baseBranch := m.git.GetCurrentBranch()
	if baseBranch == "" {
		baseBranch = m.git.GetDefaultBranch()
	}

	// Create worktree record
	wt := Worktree{
		ID:         id,
		Name:       opts.Name,
		Path:       wtPath,
		Rite:       rite,
		CreatedAt:  time.Now().UTC(),
		BaseBranch: baseBranch,
		FromRef:    opts.FromRef,
		Complexity: opts.Complexity,
	}

	// Save per-worktree metadata
	if err := SavePerWorktreeMeta(wtPath, wt, m.rootDir); err != nil {
		// Cleanup on failure
		_ = m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Add to registry
	if err := m.metadata.Add(wt); err != nil {
		// Cleanup on failure
		_ = m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Materialize channel dir and set up rite for the new worktree
	m.setupWorktreeEcosystem(wtPath, rite)

	return &wt, nil
}

// List returns all worktrees with their status.
func (m *Manager) List() ([]WorktreeStatus, error) {
	// Sync metadata with filesystem first
	_ = m.metadata.SyncMetadataFromFilesystem() // Non-fatal, continue with what we have

	worktrees, err := m.metadata.List()
	if err != nil {
		return nil, err
	}

	var results []WorktreeStatus
	for _, wt := range worktrees {
		status, err := m.getStatus(wt)
		if err != nil {
			// Include with minimal info if status check fails
			status = &WorktreeStatus{
				Worktree: wt,
				Age:      FormatAge(wt.CreatedAt),
			}
		}
		results = append(results, *status)
	}

	return results, nil
}

// Status returns detailed status for a specific worktree.
func (m *Manager) Status(idOrName string) (*WorktreeStatus, error) {
	var wt *Worktree
	var err error

	// Try as ID first
	if IsValidWorktreeID(idOrName) {
		wt, err = m.metadata.Get(idOrName)
	} else {
		// Try as name
		wt, err = m.metadata.GetByName(idOrName)
	}

	if err != nil {
		return nil, err
	}

	return m.getStatus(*wt)
}

// getStatus builds WorktreeStatus for a worktree.
func (m *Manager) getStatus(wt Worktree) (*WorktreeStatus, error) {
	status := &WorktreeStatus{
		Worktree:      wt,
		Age:           FormatAge(wt.CreatedAt),
		SessionStatus: "none",
	}

	// Get git status
	gitStatus, err := m.git.Status(wt.Path)
	if err == nil {
		status.IsDirty = gitStatus.IsDirty
		status.HasUntracked = gitStatus.HasUntracked
		status.ChangedFiles = gitStatus.ChangedFiles
		status.UntrackedCount = gitStatus.UntrackedCount
	}

	// Get commits ahead/behind
	if wt.BaseBranch != "" {
		ahead, behind, _ := m.git.GetCommitDiff(wt.Path, wt.BaseBranch)
		status.CommitsAhead = ahead
		status.CommitsBehind = behind
	}

	// Get session status
	sessionsDir := filepath.Join(wt.Path, ".sos", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "session-") {
				status.CurrentSession = entry.Name()
				status.SessionStatus = "active"

				// Check if parked
				contextPath := filepath.Join(sessionsDir, entry.Name(), "SESSION_CONTEXT.md")
				data, err := os.ReadFile(contextPath)
				if err == nil {
					content := string(data)
					if strings.Contains(content, "parked_at:") || strings.Contains(content, "auto_parked_at:") {
						status.SessionStatus = "parked"
					}
				}
				break
			}
		}
	}

	// Get branch
	status.Branch = m.git.GetBranchForPath(wt.Path)

	return status, nil
}

// Remove removes a worktree.
func (m *Manager) Remove(idOrName string, force bool) error {
	var wt *Worktree
	var err error

	// Try as ID first
	if IsValidWorktreeID(idOrName) {
		wt, err = m.metadata.Get(idOrName)
	} else {
		// Try as name
		wt, err = m.metadata.GetByName(idOrName)
	}

	if err != nil {
		return err
	}

	// Check for uncommitted changes unless force
	if !force {
		gitStatus, err := m.git.Status(wt.Path)
		if err == nil && (gitStatus.IsDirty || gitStatus.HasUntracked) {
			return errors.NewWithDetails(errors.CodeGeneralError,
				"worktree has uncommitted changes, use --force to override",
				map[string]any{
					"worktree_id":     wt.ID,
					"changed_files":   gitStatus.ChangedFiles,
					"untracked_files": gitStatus.UntrackedCount,
				})
		}
	}

	// Clean up worktree-local directories before git worktree removal.
	// These are best-effort; don't fail the Remove operation on cleanup errors.
	// Symlinked dirs (.knossos/, .know/) are just symlinks and get removed
	// with the worktree directory -- do NOT follow/delete symlink targets.
	_ = os.RemoveAll(filepath.Join(wt.Path, ".sos"))
	_ = os.RemoveAll(filepath.Join(wt.Path, ".ledge"))

	// Remove git worktree
	if err := m.git.WorktreeRemove(wt.Path, force); err != nil {
		// Try manual removal if git worktree remove fails
		if force {
			_ = os.RemoveAll(wt.Path)
		} else {
			return err
		}
	}

	// Remove from metadata
	if err := m.metadata.Remove(wt.ID); err != nil {
		return err
	}

	return nil
}

// CleanupResult represents the result of a cleanup operation.
type CleanupResult struct {
	Removed     []string          `json:"removed"`
	Skipped     []string          `json:"skipped"`
	SkipReasons map[string]string `json:"skip_reasons,omitempty"`
	DryRun      bool              `json:"dry_run"`
}

// Cleanup removes stale worktrees.
func (m *Manager) Cleanup(opts CleanupOptions) (*CleanupResult, error) {
	// Default to 7 days
	if opts.OlderThan == 0 {
		opts.OlderThan = 7 * 24 * time.Hour
	}

	// Sync metadata first
	_ = m.metadata.SyncMetadataFromFilesystem() // Non-fatal

	oldWorktrees, err := m.metadata.GetOlderThan(opts.OlderThan)
	if err != nil {
		return nil, err
	}

	result := &CleanupResult{
		Removed:     []string{},
		Skipped:     []string{},
		SkipReasons: make(map[string]string),
		DryRun:      opts.DryRun,
	}

	for _, wt := range oldWorktrees {
		skipReason := ""

		// Check for uncommitted changes
		if !opts.Force {
			gitStatus, err := m.git.Status(wt.Path)
			if err == nil && gitStatus.IsDirty {
				skipReason = "uncommitted changes"
			} else if err == nil && gitStatus.HasUntracked {
				skipReason = "untracked files"
			}

			// Check for active sessions
			if skipReason == "" {
				status, _ := m.getStatus(wt)
				if status != nil && status.SessionStatus == "active" {
					skipReason = "active session"
				}
			}
		}

		if skipReason != "" && !opts.Force {
			result.Skipped = append(result.Skipped, wt.ID)
			result.SkipReasons[wt.ID] = skipReason
			continue
		}

		if opts.DryRun {
			result.Removed = append(result.Removed, wt.ID)
			continue
		}

		// Actually remove
		if err := m.Remove(wt.ID, opts.Force); err != nil {
			result.Skipped = append(result.Skipped, wt.ID)
			result.SkipReasons[wt.ID] = err.Error()
		} else {
			result.Removed = append(result.Removed, wt.ID)
		}
	}

	// Prune orphaned refs
	if !opts.DryRun {
		_ = m.git.WorktreePrune()
	}

	return result, nil
}

// GetWorktreesDir returns the worktrees directory path.
func (m *Manager) GetWorktreesDir() string {
	return filepath.Join(m.rootDir, ".worktrees")
}

// GetRootDir returns the main project root directory.
func (m *Manager) GetRootDir() string {
	return m.rootDir
}

// CurrentWorktree returns the current worktree if in one, nil otherwise.
func (m *Manager) CurrentWorktree() (*Worktree, error) {
	if !m.git.IsWorktree() {
		return nil, nil
	}

	// We're in a worktree, try to find it
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Check for per-worktree metadata
	meta, err := LoadPerWorktreeMeta(cwd)
	if err != nil {
		return nil, err
	}

	return m.metadata.Get(meta.WorktreeID)
}

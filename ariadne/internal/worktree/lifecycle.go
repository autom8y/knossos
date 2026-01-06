package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
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
			map[string]interface{}{"ref": opts.FromRef})
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
		os.WriteFile(gitignorePath, []byte("*\n"), 0644)
	}

	// Create git worktree with detached HEAD
	if err := m.git.WorktreeAdd(wtPath, opts.FromRef, true); err != nil {
		return nil, err
	}

	// Get team from flag or detect from ACTIVE_RITE
	team := opts.Team
	if team == "" {
		activeRitePath := filepath.Join(m.rootDir, ".claude", "ACTIVE_RITE")
		data, err := os.ReadFile(activeRitePath)
		if err == nil {
			team = strings.TrimSpace(string(data))
		}
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
		Team:       team,
		CreatedAt:  time.Now().UTC(),
		BaseBranch: baseBranch,
		FromRef:    opts.FromRef,
		Complexity: opts.Complexity,
	}

	// Save per-worktree metadata
	if err := SavePerWorktreeMeta(wtPath, wt, m.rootDir); err != nil {
		// Cleanup on failure
		m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Add to registry
	if err := m.metadata.Add(wt); err != nil {
		// Cleanup on failure
		m.git.WorktreeRemove(wtPath, true)
		return nil, err
	}

	// Try to run roster-sync if available
	rosterHome := os.Getenv("ROSTER_HOME")
	if rosterHome != "" {
		syncPath := filepath.Join(rosterHome, "roster-sync")
		if _, err := os.Stat(syncPath); err == nil {
			// First check if .claude/.cem/manifest.json exists
			manifestPath := filepath.Join(wtPath, ".claude", ".cem", "manifest.json")
			if _, err := os.Stat(manifestPath); err == nil {
				// Already initialized, run sync
				cmd := exec.Command(syncPath, "sync")
				cmd.Dir = wtPath
				cmd.Run() // Ignore errors, sync is optional
			} else {
				// Not initialized, run init
				cmd := exec.Command(syncPath, "init")
				cmd.Dir = wtPath
				cmd.Run() // Ignore errors, init is optional
			}
		}
	}

	// Try to set team if specified
	if team != "" && team != "none" {
		if rosterHome != "" {
			swapTeamPath := filepath.Join(rosterHome, "swap-team.sh")
			if _, err := os.Stat(swapTeamPath); err == nil {
				cmd := exec.Command(swapTeamPath, team)
				cmd.Dir = wtPath
				cmd.Run() // Ignore errors, team setup is optional
			}
		}
	}

	return &wt, nil
}

// List returns all worktrees with their status.
func (m *Manager) List() ([]WorktreeStatus, error) {
	// Sync metadata with filesystem first
	if err := m.metadata.SyncMetadataFromFilesystem(); err != nil {
		// Non-fatal, continue with what we have
	}

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
	sessionsDir := filepath.Join(wt.Path, ".claude", "sessions")
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
				map[string]interface{}{
					"worktree_id":     wt.ID,
					"changed_files":   gitStatus.ChangedFiles,
					"untracked_files": gitStatus.UntrackedCount,
				})
		}
	}

	// Remove git worktree
	if err := m.git.WorktreeRemove(wt.Path, force); err != nil {
		// Try manual removal if git worktree remove fails
		if force {
			os.RemoveAll(wt.Path)
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
	Removed     []string `json:"removed"`
	Skipped     []string `json:"skipped"`
	SkipReasons map[string]string `json:"skip_reasons,omitempty"`
	DryRun      bool     `json:"dry_run"`
}

// Cleanup removes stale worktrees.
func (m *Manager) Cleanup(opts CleanupOptions) (*CleanupResult, error) {
	// Default to 7 days
	if opts.OlderThan == 0 {
		opts.OlderThan = 7 * 24 * time.Hour
	}

	// Sync metadata first
	if err := m.metadata.SyncMetadataFromFilesystem(); err != nil {
		// Non-fatal
	}

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
		m.git.WorktreePrune()
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

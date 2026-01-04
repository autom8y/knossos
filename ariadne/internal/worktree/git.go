package worktree

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/autom8y/ariadne/internal/errors"
)

// GitOperations provides git worktree operations.
type GitOperations struct {
	workDir string // Working directory for git commands
}

// NewGitOperations creates a new GitOperations instance.
func NewGitOperations(workDir string) *GitOperations {
	return &GitOperations{workDir: workDir}
}

// IsGitRepo checks if the working directory is a git repository.
func (g *GitOperations) IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.workDir
	return cmd.Run() == nil
}

// IsWorktree checks if the working directory is a git worktree (not main repo).
func (g *GitOperations) IsWorktree() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	// Worktrees have a .git file pointing to the real git dir, not a .git directory
	gitDir := strings.TrimSpace(string(out))
	return strings.Contains(gitDir, "worktrees")
}

// GetProjectRoot returns the git repository root directory.
func (g *GitOperations) GetProjectRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to get project root", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetMainWorktree returns the path to the main worktree (not a linked worktree).
func (g *GitOperations) GetMainWorktree() (string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to list worktrees", err)
	}

	// First worktree in list is the main worktree
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			return strings.TrimPrefix(line, "worktree "), nil
		}
	}
	return "", errors.New(errors.CodeGeneralError, "could not determine main worktree")
}

// RefExists checks if a git ref (branch, tag, commit) exists.
func (g *GitOperations) RefExists(ref string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	cmd.Dir = g.workDir
	return cmd.Run() == nil
}

// GetCurrentBranch returns the current branch name, or empty string if detached.
func (g *GitOperations) GetCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetDefaultBranch returns the default branch (main or master).
func (g *GitOperations) GetDefaultBranch() string {
	// Try main first
	if g.RefExists("main") {
		return "main"
	}
	if g.RefExists("master") {
		return "master"
	}
	return "main" // Default fallback
}

// WorktreeAdd creates a new git worktree.
func (g *GitOperations) WorktreeAdd(path, ref string, detach bool) error {
	args := []string{"worktree", "add"}
	if detach {
		args = append(args, "--detach")
	}
	args = append(args, path, ref)

	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.NewWithDetails(errors.CodeGeneralError,
			"failed to create git worktree",
			map[string]interface{}{
				"path":   path,
				"ref":    ref,
				"stderr": stderr.String(),
			})
	}
	return nil
}

// WorktreeRemove removes a git worktree.
func (g *GitOperations) WorktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.NewWithDetails(errors.CodeGeneralError,
			"failed to remove git worktree",
			map[string]interface{}{
				"path":   path,
				"stderr": stderr.String(),
			})
	}
	return nil
}

// WorktreePrune removes stale worktree references.
func (g *GitOperations) WorktreePrune() error {
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = g.workDir
	return cmd.Run()
}

// GitWorktreeEntry represents a parsed git worktree list entry.
type GitWorktreeEntry struct {
	Path     string
	Head     string
	Branch   string
	Detached bool
}

// WorktreeList returns all git worktrees.
func (g *GitOperations) WorktreeList() ([]GitWorktreeEntry, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to list worktrees", err)
	}

	var entries []GitWorktreeEntry
	var current *GitWorktreeEntry

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "worktree "):
			current = &GitWorktreeEntry{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		case strings.HasPrefix(line, "HEAD "):
			if current != nil {
				current.Head = strings.TrimPrefix(line, "HEAD ")
			}
		case strings.HasPrefix(line, "branch "):
			if current != nil {
				current.Branch = strings.TrimPrefix(line, "branch ")
			}
		case line == "detached":
			if current != nil {
				current.Detached = true
			}
		}
	}

	// Don't forget the last entry
	if current != nil {
		entries = append(entries, *current)
	}

	return entries, nil
}

// GitStatus represents the status of a working directory.
type GitStatus struct {
	IsDirty        bool
	HasUntracked   bool
	ChangedFiles   int
	UntrackedCount int
}

// Status returns the git status for a directory.
func (g *GitOperations) Status(path string) (*GitStatus, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to get git status", err)
	}

	status := &GitStatus{}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			status.HasUntracked = true
			status.UntrackedCount++
		} else {
			status.IsDirty = true
			status.ChangedFiles++
		}
	}

	return status, nil
}

// GetCommitDiff returns commits ahead/behind compared to a ref.
func (g *GitOperations) GetCommitDiff(path, baseRef string) (ahead, behind int, err error) {
	// Get current HEAD
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	headOut, err := cmd.Output()
	if err != nil {
		return 0, 0, errors.Wrap(errors.CodeGeneralError, "failed to get HEAD", err)
	}
	head := strings.TrimSpace(string(headOut))

	// Get merge base
	cmd = exec.Command("git", "merge-base", head, baseRef)
	cmd.Dir = path
	_, err = cmd.Output()
	if err != nil {
		// Can't compare, probably no common ancestor
		return 0, 0, nil
	}

	// Count ahead
	cmd = exec.Command("git", "rev-list", "--count", baseRef+"..HEAD")
	cmd.Dir = path
	aheadOut, err := cmd.Output()
	if err == nil {
		ahead, _ = strconv.Atoi(strings.TrimSpace(string(aheadOut)))
	}

	// Count behind
	cmd = exec.Command("git", "rev-list", "--count", "HEAD.."+baseRef)
	cmd.Dir = path
	behindOut, err := cmd.Output()
	if err == nil {
		behind, _ = strconv.Atoi(strings.TrimSpace(string(behindOut)))
	}

	return ahead, behind, nil
}

// GetWorktreesDir returns the conventional worktrees directory path.
func (g *GitOperations) GetWorktreesDir() (string, error) {
	root, err := g.GetMainWorktree()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".worktrees"), nil
}

// GetHead returns the HEAD commit for a path.
func (g *GitOperations) GetHead(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to get HEAD", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetBranchForPath returns the branch name for a worktree path.
func (g *GitOperations) GetBranchForPath(path string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

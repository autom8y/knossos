// Package materialize generates channel directories from templates and rite manifests.
package materialize

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/paths"
)

// isGitWorktree reports whether projectDir is a linked git worktree (not the main worktree).
//
// Uses `git rev-parse --git-common-dir` which returns:
//   - A relative path ".git" when called from the main worktree
//   - An absolute path like "/path/to/repo/.git/worktrees/name" when called from a linked worktree
//
// When --git-common-dir returns ".git" (relative), we are in the main worktree.
// When it returns an absolute path that does not end in "/.git", we are in a linked worktree.
func isGitWorktree(projectDir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return false
	}
	commonDir := strings.TrimSpace(string(out))
	if commonDir == "" {
		return false
	}
	// Relative path ".git" means we are in the main worktree.
	if !filepath.IsAbs(commonDir) {
		return false
	}
	// Absolute path: linked worktrees have commonDir like "/repo/.git"
	// (shared .git dir), while their own .git file points elsewhere.
	// Verify we are NOT the main worktree by checking that the commonDir
	// is not the same as projectDir/.git.
	mainDotGit := filepath.Join(projectDir, ".git")
	info, err := os.Stat(mainDotGit)
	if err == nil && info.IsDir() {
		// projectDir has a real .git directory → this IS the main worktree.
		return false
	}
	// projectDir has a .git file (not a dir) → linked worktree.
	return true
}

// getMainWorktreeDir returns the main worktree path for a linked worktree.
//
// `git rev-parse --git-common-dir` returns the shared .git directory, which
// lives at <main-worktree>/.git. The main worktree is therefore its parent.
//
// Returns an error if the command fails or we are not in a linked worktree.
func getMainWorktreeDir(projectDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", err
	}
	commonDir := strings.TrimSpace(string(out))
	if commonDir == "" {
		return "", os.ErrInvalid
	}
	if !filepath.IsAbs(commonDir) {
		// Relative ".git" means main worktree — not a linked worktree.
		return "", os.ErrInvalid
	}
	// commonDir is e.g. "/path/to/repo/.git"
	// Main worktree is its parent.
	return filepath.Dir(commonDir), nil
}

// inheritRiteFromMainWorktree reads the ACTIVE_RITE file from the main worktree's .knossos/
// directory. Returns empty string if no ACTIVE_RITE file exists or it is unreadable.
func inheritRiteFromMainWorktree(mainDir string) string {
	return paths.NewResolver(mainDir).ReadActiveRite()
}

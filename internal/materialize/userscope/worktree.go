package userscope

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// worktreeMainDir returns the main worktree directory for a linked git worktree.
// Uses `git rev-parse --git-common-dir` (~5ms) rather than `git worktree list`.
// Returns an error if projectDir is the main worktree, not in a git repo, or
// the git command fails.
//
// This is a local copy of the detection logic from internal/materialize/worktree.go
// to avoid a circular import (userscope → materialize → userscope).
func worktreeMainDir(projectDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", projectDir, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", err
	}
	commonDir := strings.TrimSpace(string(out))
	if commonDir == "" || !filepath.IsAbs(commonDir) {
		// Relative path ".git" means main worktree, not linked.
		return "", os.ErrInvalid
	}
	// Verify projectDir is a linked worktree by checking that its .git entry
	// is a file (pointer), not a directory (real .git).
	mainDotGit := filepath.Join(projectDir, ".git")
	info, statErr := os.Stat(mainDotGit)
	if statErr == nil && info.IsDir() {
		// Main worktree has a real .git directory.
		return "", os.ErrInvalid
	}
	// commonDir is e.g. "/path/to/repo/.git"; main worktree is its parent.
	return filepath.Dir(commonDir), nil
}

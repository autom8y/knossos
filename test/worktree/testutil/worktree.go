// Package testutil provides shared test fixtures for worktree scenarios.
// These helpers create git repositories with linked worktrees so that
// worktree-aware code paths in the materialize and hook packages can be
// exercised without shell scripts or external fixtures.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

// WorktreeFixture holds paths created by SetupWorktreeTestFixture.
type WorktreeFixture struct {
	// MainDir is the main (primary) worktree directory. It contains a real
	// .git/ directory and a populated .claude/ with ACTIVE_RITE.
	MainDir string

	// WorktreeDir is the linked worktree directory. It contains a .git file
	// (pointer) but no .claude/ — matching the state after `git worktree add`.
	WorktreeDir string
}

// SetupWorktreeTestFixture creates a git repository with a linked worktree and
// returns a WorktreeFixture. The fixture simulates the real-world state where:
//
//   - The main worktree has .claude/ACTIVE_RITE and a valid PROVENANCE_MANIFEST.yaml
//   - The linked worktree has none of these (gitignored directories are absent)
//
// Cleanup is registered via t.Cleanup and handles both worktree removal and
// temp directory teardown. Tests must set KNOSSOS_HOME isolation separately if
// they call code that reads from KNOSSOS_HOME.
func SetupWorktreeTestFixture(t *testing.T) WorktreeFixture {
	t.Helper()

	// 1. Create temp dir and run git init
	mainDir := t.TempDir()
	gitInit(t, mainDir, "Test User", "test@example.com")

	// 2. Create an initial commit so `git worktree add` succeeds.
	//    git worktree add requires at least one commit.
	initFile := filepath.Join(mainDir, "README.md")
	writeFile(t, initFile, "# Test Repository\n")
	gitRun(t, mainDir, "add", "README.md")
	gitRun(t, mainDir, "commit", "-m", "initial commit")

	// 3. Set up .claude/ACTIVE_RITE in main (simulates a post-sync main worktree).
	claudeDir := filepath.Join(mainDir, ".claude")
	mkdirAll(t, claudeDir)
	writeFile(t, filepath.Join(claudeDir, "ACTIVE_RITE"), "test-rite\n")

	// 4. Write a minimal PROVENANCE_MANIFEST.yaml so the collision checker has
	//    a valid manifest to load from the main worktree's .claude/.
	//    Validation rules: knossos owner requires non-empty SourcePath and SourceType;
	//    Checksum must match sha256:[0-9a-f]{64}.
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    "test-rite",
		Entries: map[string]*provenance.ProvenanceEntry{
			"ACTIVE_RITE": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeRite,
				SourcePath: "rites/test-rite/ACTIVE_RITE",
				SourceType: "project",
				Checksum:   "sha256:a000000000000000000000000000000000000000000000000000000000000001",
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.ManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("SetupWorktreeTestFixture: failed to write provenance manifest: %v", err)
	}

	// 5. Create .sos/sessions/ and .knossos/ in main.
	mkdirAll(t, filepath.Join(mainDir, ".sos", "sessions"))
	mkdirAll(t, filepath.Join(mainDir, ".knossos"))

	// 6. Create the linked worktree in a sibling directory.
	//    The worktree directory itself is inside a parent temp dir so it is
	//    not a subdirectory of mainDir (git does not allow worktrees inside
	//    the repo).
	worktreeParent := t.TempDir()
	worktreeDir := filepath.Join(worktreeParent, "linked-worktree")

	// Create a new branch for the worktree to avoid "already checked out" errors.
	gitRun(t, mainDir, "branch", "worktree-branch")
	gitRun(t, mainDir, "worktree", "add", worktreeDir, "worktree-branch")

	// Register cleanup: remove the worktree before the temp dirs are deleted,
	// otherwise git leaves stale worktree metadata.
	t.Cleanup(func() {
		// Best-effort cleanup — ignore errors so the test suite does not
		// fail on teardown if the fixture was already partially cleaned.
		cmd := exec.Command("git", "-C", mainDir, "worktree", "remove", "--force", worktreeDir)
		cmd.Run() //nolint:errcheck
	})

	return WorktreeFixture{
		MainDir:     mainDir,
		WorktreeDir: worktreeDir,
	}
}

// gitInit initialises a git repository with a local user identity so commits
// work even in environments without a global git config (e.g., CI).
func gitInit(t *testing.T, dir, name, email string) {
	t.Helper()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.name", name)
	gitRun(t, dir, "config", "user.email", email)
}

// gitRun runs a git command in dir and fails the test on error.
func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, string(out))
	}
}

// writeFile writes content to path, failing the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}

// mkdirAll creates the directory hierarchy, failing the test on error.
func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdirAll(%s): %v", path, err)
	}
}

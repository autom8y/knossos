package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	worktreefixture "github.com/autom8y/knossos/test/worktree/testutil"
)

// TestIsGitWorktree_MainWorktree verifies that a freshly initialised git repository
// (the main worktree) is not reported as a linked worktree.
func TestIsGitWorktree_MainWorktree(t *testing.T) {
	t.Parallel()
	fix := worktreefixture.SetupWorktreeTestFixture(t)

	if isGitWorktree(fix.MainDir) {
		t.Errorf("isGitWorktree(%s) = true; main worktree should return false", fix.MainDir)
	}
}

// TestIsGitWorktree_LinkedWorktree verifies that a linked worktree (created via
// `git worktree add`) is correctly identified.
func TestIsGitWorktree_LinkedWorktree(t *testing.T) {
	t.Parallel()
	fix := worktreefixture.SetupWorktreeTestFixture(t)

	if !isGitWorktree(fix.WorktreeDir) {
		t.Errorf("isGitWorktree(%s) = false; linked worktree should return true", fix.WorktreeDir)
	}
}

// TestIsGitWorktree_NonGitDir verifies that a plain directory (not a git repo)
// returns false without panicking.
func TestIsGitWorktree_NonGitDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	if isGitWorktree(tmpDir) {
		t.Error("isGitWorktree() on non-git dir should return false")
	}
}

// TestGetMainWorktreeDir_FromLinkedWorktree verifies that getMainWorktreeDir
// returns the main worktree directory when called from a linked worktree.
func TestGetMainWorktreeDir_FromLinkedWorktree(t *testing.T) {
	t.Parallel()
	fix := worktreefixture.SetupWorktreeTestFixture(t)

	mainDir, err := getMainWorktreeDir(fix.WorktreeDir)
	if err != nil {
		t.Fatalf("getMainWorktreeDir(%s) error = %v", fix.WorktreeDir, err)
	}

	// Resolve symlinks for comparison since temp dirs may use symlinks on macOS.
	got, err := filepath.EvalSymlinks(mainDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", mainDir, err)
	}
	want, err := filepath.EvalSymlinks(fix.MainDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", fix.MainDir, err)
	}

	if got != want {
		t.Errorf("getMainWorktreeDir() = %q, want %q", got, want)
	}
}

// TestGetMainWorktreeDir_FromMainWorktree verifies that getMainWorktreeDir
// returns an error when called from the main worktree (it is not a linked worktree).
func TestGetMainWorktreeDir_FromMainWorktree(t *testing.T) {
	t.Parallel()
	fix := worktreefixture.SetupWorktreeTestFixture(t)

	_, err := getMainWorktreeDir(fix.MainDir)
	if err == nil {
		t.Error("getMainWorktreeDir() from main worktree should return an error")
	}
}

// TestSyncRiteScope_InWorktree_InheritsRite is a P0 test verifying that when
// `ari sync` is run from a linked worktree (no ACTIVE_RITE present), it
// inherits the rite from the main worktree's .knossos/ACTIVE_RITE.
func TestSyncRiteScope_InWorktree_InheritsRite(t *testing.T) {
	t.Parallel()
	// Use an isolated dir for knossos home so only test rites are visible.
	isolatedHome := t.TempDir()

	fix := worktreefixture.SetupWorktreeTestFixture(t)

	// The main worktree already has .knossos/ACTIVE_RITE = "test-rite" (set by fixture).
	// Create a minimal rite in the isolated home's rites/ dir so the SourceResolver
	// can find it via the knossos tier.
	riteDir := filepath.Join(isolatedHome, "rites", "test-rite")
	if err := os.MkdirAll(filepath.Join(riteDir, "agents"), 0755); err != nil {
		t.Fatalf("failed to create rite agents dir: %v", err)
	}
	manifest := `name: test-rite
version: "1.0"
description: "Test rite for worktree inheritance"
entry_agent: worker
agents:
  - name: worker
    role: "Worker agent"
dromena: []
legomena: []
hooks: []
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(riteDir, "agents", "worker.md"), []byte("# Worker\n"), 0644); err != nil {
		t.Fatalf("failed to write agent: %v", err)
	}

	// The worktree has no .claude/ at all — ensureProjectDirs() will create it.
	worktreeResolver := paths.NewResolver(fix.WorktreeDir)
	// Inject source resolver scoped to project + isolated knossos home only.
	sr := NewSourceResolverWithPaths(fix.WorktreeDir, "", "", isolatedHome)
	m := NewMaterializer(worktreeResolver).WithSourceResolver(sr)

	// Run rite-scope sync with no --rite flag (simulates running `ari sync` in worktree).
	result, err := m.Sync(SyncOptions{Scope: ScopeRite})
	if err != nil {
		t.Fatalf("Sync() from worktree error = %v", err)
	}

	if result.RiteResult == nil {
		t.Fatal("RiteResult should not be nil")
	}

	// The worktree sync should succeed (not "minimal") with the inherited rite.
	if result.RiteResult.Status != "success" {
		t.Errorf("RiteResult.Status = %q, want %q", result.RiteResult.Status, "success")
	}

	// ACTIVE_RITE must be written into the worktree's .knossos/ so subsequent
	// hook calls can find the rite without re-inheriting.
	worktreeActiveRite := filepath.Join(fix.WorktreeDir, ".knossos", "ACTIVE_RITE")
	data, err := os.ReadFile(worktreeActiveRite)
	if err != nil {
		t.Fatalf("ACTIVE_RITE not created in worktree: %v", err)
	}
	if got := trimNewline(string(data)); got != "test-rite" {
		t.Errorf("worktree ACTIVE_RITE = %q, want %q", got, "test-rite")
	}
}

// TestSyncRiteScope_InWorktree_NoMainRite_FallsToMinimal is a P0 test verifying
// that when neither the worktree nor the main worktree has an ACTIVE_RITE, sync
// falls through to minimal mode.
func TestSyncRiteScope_InWorktree_NoMainRite_FallsToMinimal(t *testing.T) {
	t.Parallel()
	isolatedHome := t.TempDir()

	fix := worktreefixture.SetupWorktreeTestFixture(t)

	// Remove ACTIVE_RITE from the main worktree to simulate a fresh main repo
	// that has never been synced with a rite.
	if err := os.Remove(filepath.Join(fix.MainDir, ".knossos", "ACTIVE_RITE")); err != nil {
		t.Fatalf("failed to remove main ACTIVE_RITE: %v", err)
	}

	// The worktree has no .knossos/ and the main has no ACTIVE_RITE.
	worktreeResolver := paths.NewResolver(fix.WorktreeDir)
	sr := NewSourceResolverWithPaths(fix.WorktreeDir, "", "", isolatedHome)
	m := NewMaterializer(worktreeResolver).WithSourceResolver(sr)

	// scope=all: should fall through to minimal, not return an error.
	result, err := m.Sync(SyncOptions{Scope: ScopeAll})
	if err != nil {
		t.Fatalf("Sync() with no ACTIVE_RITE anywhere should not error: %v", err)
	}

	if result.RiteResult == nil {
		t.Fatal("RiteResult should not be nil even for minimal sync")
	}

	// Without an ACTIVE_RITE anywhere, the sync should produce a minimal result.
	if result.RiteResult.Status != "minimal" {
		t.Errorf("RiteResult.Status = %q, want %q (no rite should fall to minimal)", result.RiteResult.Status, "minimal")
	}
}

// trimNewline removes trailing newline characters for assertion comparisons.
func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

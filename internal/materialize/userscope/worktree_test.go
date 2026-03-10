package userscope

import (
	"os"
	"path/filepath"
	"testing"

	worktreefixture "github.com/autom8y/knossos/test/worktree/testutil"
)

// TestCollisionChecker_InWorktree_FallsBackToMainProvenance is a P1 test
// verifying that the collision checker falls back to the main worktree's
// PROVENANCE_MANIFEST.yaml when the local (worktree) .knossos/ has none.
//
// The fixture sets up main with ACTIVE_RITE and a valid provenance manifest.
// The worktree has no .knossos/ at all. worktreeMainDir() should resolve to
// the main worktree, and NewCollisionChecker(mainKnossosDir) should succeed.
func TestCollisionChecker_InWorktree_FallsBackToMainProvenance(t *testing.T) {
	t.Parallel()
	fix := worktreefixture.SetupWorktreeTestFixture(t)

	// Worktree has no .knossos/ — the directory does not exist yet.
	// Verify that attempting to create a checker directly from the worktree's
	// (non-existent) knossos dir reports IsEffective() = false.
	worktreeKnossosDir := filepath.Join(fix.WorktreeDir, ".knossos")
	localChecker := NewCollisionChecker(worktreeKnossosDir)
	if localChecker.IsEffective() {
		t.Fatal("expected local collision checker to be ineffective (no .knossos/ in worktree)")
	}

	// Simulate the fallback logic in sync.go:
	//   if !collisionChecker.IsEffective() {
	//       if mainDir, err := worktreeMainDir(...); err == nil {
	//           mainChecker := NewCollisionChecker(mainKnossosDir)
	//           if mainChecker.IsEffective() { use it }
	//       }
	//   }
	mainDir, err := worktreeMainDir(fix.WorktreeDir)
	if err != nil {
		t.Fatalf("worktreeMainDir(%s) error = %v", fix.WorktreeDir, err)
	}

	mainKnossosDir := filepath.Join(mainDir, ".knossos")
	mainChecker := NewCollisionChecker(mainKnossosDir)
	if !mainChecker.IsEffective() {
		t.Error("main worktree collision checker should be effective (has PROVENANCE_MANIFEST.yaml from fixture)")
	}

	// The main worktree's manifest contains a rite-scoped entry for "ACTIVE_RITE"
	// (written by the fixture). Verify collision detection works via the fallback checker.
	collides, reason := mainChecker.CheckCollision("ACTIVE_RITE")
	if !collides {
		t.Errorf("expected collision for 'ACTIVE_RITE' (rite-owned in main manifest), reason: %q", reason)
	}
}

// TestCollisionChecker_InWorktree_NoMainProvenance_FailsClosed is a P1 test
// verifying that when neither the worktree nor the main worktree has a
// provenance manifest, the fallback fails closed: worktreeMainDir() succeeds
// but the main checker is also ineffective, so the caller must treat
// IsEffective() = false as a "skip all writes" signal.
func TestCollisionChecker_InWorktree_NoMainProvenance_FailsClosed(t *testing.T) {
	t.Parallel()
	fix := worktreefixture.SetupWorktreeTestFixture(t)

	// Remove the provenance manifest from the main worktree to simulate a
	// scenario where the main was never synced with a rite.
	mainManifest := filepath.Join(fix.MainDir, ".knossos", "PROVENANCE_MANIFEST.yaml")
	if err := os.Remove(mainManifest); err != nil {
		t.Fatalf("failed to remove main provenance manifest: %v", err)
	}

	// Local checker for the worktree (no .knossos/ at all).
	worktreeKnossosDir := filepath.Join(fix.WorktreeDir, ".knossos")
	localChecker := NewCollisionChecker(worktreeKnossosDir)
	if localChecker.IsEffective() {
		t.Fatal("expected local collision checker to be ineffective (no .knossos/ in worktree)")
	}

	// Fallback to main — main also has no manifest.
	mainDir, err := worktreeMainDir(fix.WorktreeDir)
	if err != nil {
		t.Fatalf("worktreeMainDir(%s) error = %v", fix.WorktreeDir, err)
	}

	mainKnossosDir := filepath.Join(mainDir, ".knossos")
	mainChecker := NewCollisionChecker(mainKnossosDir)

	// Both checkers are ineffective — fail-closed.
	if mainChecker.IsEffective() {
		t.Error("main worktree collision checker should also be ineffective (manifest was removed)")
	}

	// When both are ineffective the calling code (sync.go) checks for ACTIVE_RITE
	// and skips user-scope writes if one is set. We verify that IsEffective()=false
	// on both ends of the chain, so the caller has a correct signal.
	if localChecker.IsEffective() || mainChecker.IsEffective() {
		t.Error("fail-closed: both checkers should report IsEffective()=false with no manifests")
	}
}

// TestWorktreeMainDir_NotInWorktree verifies that worktreeMainDir returns an
// error when called from a plain directory (not a git repo or main worktree).
func TestWorktreeMainDir_NotInWorktree(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, err := worktreeMainDir(tmpDir)
	if err == nil {
		t.Error("worktreeMainDir() on non-git dir should return an error")
	}
}

// TestWorktreeMainDir_FromLinkedWorktree verifies that worktreeMainDir correctly
// returns the main worktree path when called from a linked worktree.
func TestWorktreeMainDir_FromLinkedWorktree(t *testing.T) {
	t.Parallel()
	fix := worktreefixture.SetupWorktreeTestFixture(t)

	got, err := worktreeMainDir(fix.WorktreeDir)
	if err != nil {
		t.Fatalf("worktreeMainDir(%s) error = %v", fix.WorktreeDir, err)
	}

	// Resolve symlinks for comparison (macOS uses /private/var/folders/... for TempDir).
	gotResolved, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", got, err)
	}
	wantResolved, err := filepath.EvalSymlinks(fix.MainDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", fix.MainDir, err)
	}

	if gotResolved != wantResolved {
		t.Errorf("worktreeMainDir() = %q, want %q", gotResolved, wantResolved)
	}
}

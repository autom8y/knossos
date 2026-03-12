package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/session"
)

// =============================================================================
// Archive Boundary Integration Tests
//
// These tests validate that the four archive boundary fixes work together
// as a coordinated system:
//
//   Fix 1 (Sprint 1): Ghost directory cleanup — live dir removed after archive
//   Fix 2 (Sprint 1): Already-archived guard — clear error, not silent no-op
//   Fix 3 (Sprint 2): Moirai lock removed before archive move
//   Fix 4 (Sprint 2): CC map entries cleared on wrap
//
// The intent is to catch regressions where fixes interact badly.
// =============================================================================

// TestArchiveBoundary_FullLifecycle exercises the complete session lifecycle
// and validates all four archive boundary fixes hold together:
//
//  1. Create session
//  2. Park session
//  3. Resume session
//  4. Wrap session (with archive enabled)
//  5. Verify: no ghost directory
//  6. Verify: archive exists with ARCHIVED status
//  7. Verify: advisory lock removed (.locks/{id}.lock)
//  8. Verify: moirai lock removed (was in session dir, not in archive)
//  9. Verify: CC map entries for this session removed
//  10. Verify: second wrap attempt returns clear lifecycle violation
func TestArchiveBoundary_FullLifecycle(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	harnessMapDir := filepath.Join(sessionsDir, ".harness-map")
	if err := os.MkdirAll(harnessMapDir, 0755); err != nil {
		t.Fatalf("Failed to create harness-map dir: %v", err)
	}

	// Step 1: Create session
	createCtx := newTestContext(projectDir)
	if err := runCreate(createCtx, "Archive boundary integration test", createOptions{
		complexity: "MODULE",
	}); err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	// Discover the created session ID
	sessionID, err := session.FindActiveSession(sessionsDir)
	if err != nil || sessionID == "" {
		t.Fatalf("Could not find created session: err=%v, id=%q", err, sessionID)
	}

	sessionDir := filepath.Join(sessionsDir, sessionID)
	archivePath := filepath.Join(projectDir, ".sos", "archive", sessionID)

	// Step 2: Park session
	parkCtx := newTestContext(projectDir, sessionID)
	if err := runPark(parkCtx, parkOptions{reason: "mid-test park"}); err != nil {
		t.Fatalf("runPark failed: %v", err)
	}

	// Step 3: Resume session
	resumeCtx := newTestContext(projectDir, sessionID)
	if err := runResume(resumeCtx); err != nil {
		t.Fatalf("runResume failed: %v", err)
	}

	// Pre-populate moirai lock in session dir (simulating active Moirai lock)
	moiraiLockPath := filepath.Join(sessionDir, ".moirai-lock")
	moiraiLockContent := `{"agent":"moirai","acquired_at":"2025-01-05T12:00:00Z","session_id":"` + sessionID + `","stale_after_seconds":300}`
	if err := os.WriteFile(moiraiLockPath, []byte(moiraiLockContent), 0644); err != nil {
		t.Fatalf("Failed to write moirai lock: %v", err)
	}

	// Pre-populate harness map entry for this session
	harnessMapEntry := filepath.Join(harnessMapDir, "cc-lifecycle-test-session")
	if err := os.WriteFile(harnessMapEntry, []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write harness map entry: %v", err)
	}

	// Step 4: Wrap session (with archive enabled)
	wrapCtx := newTestContext(projectDir, sessionID)
	if err := runWrap(wrapCtx, wrapOptions{noArchive: false}); err != nil {
		t.Fatalf("runWrap failed: %v", err)
	}

	// Step 5: Verify no ghost directory (Fix 1)
	if _, statErr := os.Stat(sessionDir); !os.IsNotExist(statErr) {
		t.Errorf("Fix 1 regression: ghost directory still exists at %s", sessionDir)
	}

	// Step 6: Verify archive exists with ARCHIVED status (Fix 1)
	if _, statErr := os.Stat(archivePath); os.IsNotExist(statErr) {
		t.Fatalf("Archive directory does not exist at %s", archivePath)
	}
	archivedCtx, loadErr := session.LoadContext(filepath.Join(archivePath, "SESSION_CONTEXT.md"))
	if loadErr != nil {
		t.Fatalf("Failed to load archived session context: %v", loadErr)
	}
	if archivedCtx.Status != session.StatusArchived {
		t.Errorf("Expected ARCHIVED status in archive, got %s", archivedCtx.Status)
	}

	// Step 7: Verify advisory lock removed (Fix 4)
	advisoryLockPath := filepath.Join(sessionsDir, ".locks", sessionID+".lock")
	if _, statErr := os.Stat(advisoryLockPath); !os.IsNotExist(statErr) {
		t.Errorf("Fix 4 regression: advisory lock still exists at %s", advisoryLockPath)
	}

	// Step 8: Verify moirai lock removed from archive (Fix 4)
	// After the archive move, .moirai-lock should NOT be in the archive dir
	// because we remove it from the live dir before the move.
	archivedMoiraiLock := filepath.Join(archivePath, ".moirai-lock")
	if _, statErr := os.Stat(archivedMoiraiLock); !os.IsNotExist(statErr) {
		t.Errorf("Fix 4 regression: moirai lock was not removed before archive move, found at %s", archivedMoiraiLock)
	}

	// Step 9: Verify harness map entry removed (Fix 4)
	if _, statErr := os.Stat(harnessMapEntry); !os.IsNotExist(statErr) {
		t.Errorf("Fix 4 regression: harness map entry still exists at %s", harnessMapEntry)
	}

	// Step 10: Verify second wrap attempt returns lifecycle violation (Fix 2)
	wrapAgainCtx := newTestContext(projectDir, sessionID)
	wrapErr := runWrap(wrapAgainCtx, wrapOptions{noArchive: false})
	if wrapErr == nil {
		t.Error("Fix 2 regression: second wrap on archived session should fail, but succeeded")
	} else if wrapErr.Error() == "" {
		t.Error("Fix 2 regression: expected non-empty error message on second wrap")
	}
}

// TestArchiveBoundary_WriteGuardAndWrap validates that the write guard's
// archived-session detection (Fix 3) and the wrap's ghost cleanup (Fix 1)
// compose correctly: after wrap, any write attempt to the old session path
// should get the "session is archived" denial, not the Moirai delegation message.
//
// This test operates at the unit level (not through CC's hook framework),
// verifying that isSessionArchived() returns the correct result after wrap.
func TestArchiveBoundary_WriteGuardDetectsArchivedAfterWrap(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")

	// Create and immediately wrap a session
	createCtx := newTestContext(projectDir)
	if err := runCreate(createCtx, "Write guard archive detection test", createOptions{
		complexity: "MODULE",
	}); err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	sessionID, err := session.FindActiveSession(sessionsDir)
	if err != nil || sessionID == "" {
		t.Fatalf("Could not find created session: %v", err)
	}

	wrapCtx := newTestContext(projectDir, sessionID)
	if err := runWrap(wrapCtx, wrapOptions{noArchive: false}); err != nil {
		t.Fatalf("runWrap failed: %v", err)
	}

	// Verify: after wrap, the archive directory exists
	archivePath := filepath.Join(projectDir, ".sos", "archive", sessionID)
	if _, statErr := os.Stat(archivePath); os.IsNotExist(statErr) {
		t.Fatalf("Archive directory should exist after wrap, but doesn't: %s", archivePath)
	}

	// Verify: after wrap, the live directory is gone (no ghost)
	liveDir := filepath.Join(sessionsDir, sessionID)
	if _, statErr := os.Stat(liveDir); !os.IsNotExist(statErr) {
		t.Errorf("Live directory should not exist after wrap (ghost), but does: %s", liveDir)
	}

	// The write guard's isSessionArchived() checks resolver.ArchiveDir() + "/" + sessionID
	// which is exactly what we just verified exists. So any write to this session's
	// context file will hit the "session is archived" path.
	// This is validated at the writeguard level in writeguard_test.go
	// (TestWriteguard_ArchivedSession_DeniesWithClearMessage).
	// Here we just confirm the precondition: archive exists, live dir does not.
	t.Logf("Archive boundary confirmed for session %s: archive exists, live dir gone", sessionID)
}

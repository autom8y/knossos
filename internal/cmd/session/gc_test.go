package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/session"
)

// TestGc_NoStaleSessions verifies that gc exits cleanly with no sessions to archive.
func TestGc_NoStaleSessions(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")

	// Create a PARKED session but with a very recent parked_at (not stale)
	createCtx := newTestContext(projectDir)
	if err := runCreate(createCtx, "Fresh session", createOptions{complexity: "PATCH"}); err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
	sessionID, err := session.FindActiveSession(sessionsDir)
	if err != nil || sessionID == "" {
		t.Fatalf("Could not find created session: %v", err)
	}
	parkCtx := newTestContext(projectDir, sessionID)
	if err := runPark(parkCtx, parkOptions{reason: "test"}); err != nil {
		t.Fatalf("runPark failed: %v", err)
	}

	// Use a threshold larger than the session age (7 days — session was just parked)
	gcCtx := newTestContext(projectDir)
	if err := runGc(gcCtx, gcOptions{staleDays: 7}); err != nil {
		t.Errorf("runGc should succeed with no stale sessions, got: %v", err)
	}

	// Session should still be PARKED (not archived)
	ctxPath := filepath.Join(sessionsDir, sessionID, "SESSION_CONTEXT.md")
	sessCtx, loadErr := session.LoadContext(ctxPath)
	if loadErr != nil {
		t.Fatalf("Failed to load session context: %v", loadErr)
	}
	if sessCtx.Status != session.StatusParked {
		t.Errorf("Status = %v, want PARKED (session should not be archived)", sessCtx.Status)
	}
}

// TestGc_DryRun verifies that --dry-run lists stale sessions but does not archive them.
func TestGc_DryRun(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")

	// Create a session and park it, then manually backdate its parked_at to make it stale
	createCtx := newTestContext(projectDir)
	if err := runCreate(createCtx, "Stale candidate", createOptions{complexity: "PATCH"}); err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
	sessionID, err := session.FindActiveSession(sessionsDir)
	if err != nil || sessionID == "" {
		t.Fatalf("Could not find created session: %v", err)
	}
	parkCtx := newTestContext(projectDir, sessionID)
	if err := runPark(parkCtx, parkOptions{reason: "going stale"}); err != nil {
		t.Fatalf("runPark failed: %v", err)
	}

	// Backdate parked_at to 10 days ago
	ctxPath := filepath.Join(sessionsDir, sessionID, "SESSION_CONTEXT.md")
	backdateParkedAt(t, ctxPath, -10*24*time.Hour)

	// Run gc in dry-run mode with 7-day threshold
	gcCtx := newTestContext(projectDir)
	if err := runGc(gcCtx, gcOptions{staleDays: 7, dryRun: true}); err != nil {
		t.Errorf("runGc --dry-run should succeed, got: %v", err)
	}

	// Session should still be PARKED — dry-run must not archive
	sessCtx, loadErr := session.LoadContext(ctxPath)
	if loadErr != nil {
		t.Fatalf("Failed to load session context: %v", loadErr)
	}
	if sessCtx.Status != session.StatusParked {
		t.Errorf("Status = %v, want PARKED (dry-run must not archive)", sessCtx.Status)
	}
}

// TestGc_Force verifies that --force archives stale sessions without prompting.
func TestGc_Force(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	archiveDir := filepath.Join(projectDir, ".claude", ".archive", "sessions")

	// Create and park a session
	createCtx := newTestContext(projectDir)
	if err := runCreate(createCtx, "Force gc candidate", createOptions{complexity: "PATCH"}); err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}
	sessionID, err := session.FindActiveSession(sessionsDir)
	if err != nil || sessionID == "" {
		t.Fatalf("Could not find created session: %v", err)
	}
	parkCtx := newTestContext(projectDir, sessionID)
	if err := runPark(parkCtx, parkOptions{reason: "force gc test"}); err != nil {
		t.Fatalf("runPark failed: %v", err)
	}

	// Backdate to 10 days ago
	ctxPath := filepath.Join(sessionsDir, sessionID, "SESSION_CONTEXT.md")
	backdateParkedAt(t, ctxPath, -10*24*time.Hour)

	// Run gc with --force and 7-day threshold
	gcCtx := newTestContext(projectDir)
	if err := runGc(gcCtx, gcOptions{staleDays: 7, force: true}); err != nil {
		t.Errorf("runGc --force should succeed, got: %v", err)
	}

	// Session must be archived: archive exists, live dir gone
	archivePath := filepath.Join(archiveDir, sessionID)
	if _, statErr := os.Stat(archivePath); os.IsNotExist(statErr) {
		t.Errorf("Archive does not exist at %s after gc --force", archivePath)
	}
	liveDir := filepath.Join(sessionsDir, sessionID)
	if _, statErr := os.Stat(liveDir); !os.IsNotExist(statErr) {
		t.Errorf("Live directory still exists at %s after gc --force (ghost)", liveDir)
	}
}

// TestGc_ForceRespectsBoundaryFixes verifies that gc delegates to runWrap()
// and therefore inherits the Session A archive boundary fixes:
// - No ghost directory after archive
// - Archive has ARCHIVED status
func TestGc_ForceRespectsBoundaryFixes(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	archiveDir := filepath.Join(projectDir, ".claude", ".archive", "sessions")

	// Create, park, and backdate two sessions
	ids := make([]string, 2)
	for i, initiative := range []string{"gc boundary test 1", "gc boundary test 2"} {
		createCtx := newTestContext(projectDir)
		if err := runCreate(createCtx, initiative, createOptions{complexity: "PATCH"}); err != nil {
			t.Fatalf("runCreate failed: %v", err)
		}
		sessionID, err := session.FindActiveSession(sessionsDir)
		if err != nil || sessionID == "" {
			t.Fatalf("Could not find created session: %v", err)
		}
		parkCtx := newTestContext(projectDir, sessionID)
		if err := runPark(parkCtx, parkOptions{reason: "gc boundary test"}); err != nil {
			t.Fatalf("runPark failed: %v", err)
		}
		ctxPath := filepath.Join(sessionsDir, sessionID, "SESSION_CONTEXT.md")
		backdateParkedAt(t, ctxPath, -10*24*time.Hour)
		ids[i] = sessionID
	}

	// Run gc --force
	gcCtx := newTestContext(projectDir)
	if err := runGc(gcCtx, gcOptions{staleDays: 7, force: true}); err != nil {
		t.Errorf("runGc failed: %v", err)
	}

	// Verify both sessions: archive has ARCHIVED status, live dir gone
	for _, id := range ids {
		archivePath := filepath.Join(archiveDir, id)
		if _, statErr := os.Stat(archivePath); os.IsNotExist(statErr) {
			t.Errorf("Archive does not exist at %s", archivePath)
			continue
		}
		archivedCtx, loadErr := session.LoadContext(filepath.Join(archivePath, "SESSION_CONTEXT.md"))
		if loadErr != nil {
			t.Errorf("Failed to load archived context for %s: %v", id, loadErr)
			continue
		}
		if archivedCtx.Status != session.StatusArchived {
			t.Errorf("Session %s: Status = %v, want ARCHIVED", id, archivedCtx.Status)
		}
		liveDir := filepath.Join(sessionsDir, id)
		if _, statErr := os.Stat(liveDir); !os.IsNotExist(statErr) {
			t.Errorf("Session %s: live dir still exists (ghost)", id)
		}
	}
}

// backdateParkedAt modifies a session context file to set parked_at in the past.
// Used to simulate stale sessions in tests without waiting.
func backdateParkedAt(t *testing.T, ctxPath string, offset time.Duration) {
	t.Helper()
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("backdateParkedAt: failed to load context: %v", err)
	}
	backdated := time.Now().Add(offset).UTC()
	sessCtx.ParkedAt = &backdated
	if err := sessCtx.Save(ctxPath); err != nil {
		t.Fatalf("backdateParkedAt: failed to save context: %v", err)
	}
}

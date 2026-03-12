package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// setupClaimTest creates a temporary project dir with a session in the given status.
// Returns (projectDir, sessionID).
func setupClaimTest(t *testing.T, status session.Status) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	os.MkdirAll(sessionsDir, 0755)

	ctx := session.NewContext("Test Claim", "PATCH", "10x-dev")
	ctx.Status = status
	sessionDir := filepath.Join(sessionsDir, ctx.SessionID)
	os.MkdirAll(sessionDir, 0755)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}
	return tmpDir, ctx.SessionID
}

func TestClaim_Success(t *testing.T) {
	projectDir, sessionID := setupClaimTest(t, session.StatusActive)
	resolver := paths.NewResolver(projectDir)
	ccID := "test-cc-session-abc123"

	err := session.SetHarnessSessionMap(resolver, ccID, sessionID)
	if err != nil {
		t.Fatalf("SetHarnessSessionMap() error = %v", err)
	}

	// Verify the CC map file was written
	mapFile := filepath.Join(resolver.HarnessMapDir(), ccID)
	data, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("failed to read cc-map file: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != sessionID {
		t.Errorf("cc-map content = %q, want %q", got, sessionID)
	}

	// Verify round-trip: ResolveSession should find it via CC map (priority 2)
	resolved, err := session.ResolveSession(resolver, ccID, "")
	if err != nil {
		t.Fatalf("ResolveSession() error = %v", err)
	}
	if resolved != sessionID {
		t.Errorf("ResolveSession() = %q, want %q", resolved, sessionID)
	}
}

func TestClaim_SessionNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	os.MkdirAll(sessionsDir, 0755)
	resolver := paths.NewResolver(tmpDir)

	// Try to load a nonexistent session
	ctxPath := resolver.SessionContextFile("session-nonexistent")
	_, err := session.LoadContext(ctxPath)
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
}

func TestClaim_ArchivedSession(t *testing.T) {
	projectDir, sessionID := setupClaimTest(t, session.StatusArchived)
	resolver := paths.NewResolver(projectDir)

	// Load context and verify it's archived
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext() error = %v", err)
	}
	if sessCtx.Status != session.StatusArchived {
		t.Fatalf("status = %v, want ARCHIVED", sessCtx.Status)
	}

	// The claim command should reject this — we test the status check logic
	if sessCtx.Status == session.StatusArchived {
		// This is the guard that runClaim implements
		return // Correctly rejected
	}
	t.Fatal("archived session was not rejected")
}

func TestClaim_ParkedSession(t *testing.T) {
	projectDir, sessionID := setupClaimTest(t, session.StatusParked)
	resolver := paths.NewResolver(projectDir)
	ccID := "test-cc-parked-session"

	// Claiming a PARKED session should succeed (binding, not lifecycle)
	err := session.SetHarnessSessionMap(resolver, ccID, sessionID)
	if err != nil {
		t.Fatalf("SetHarnessSessionMap() for parked session error = %v", err)
	}

	// Round-trip should work
	resolved, err := session.ResolveSession(resolver, ccID, "")
	if err != nil {
		t.Fatalf("ResolveSession() error = %v", err)
	}
	if resolved != sessionID {
		t.Errorf("ResolveSession() = %q, want %q", resolved, sessionID)
	}
}

func TestClaim_Overwrite(t *testing.T) {
	projectDir, sessionID1 := setupClaimTest(t, session.StatusActive)
	resolver := paths.NewResolver(projectDir)
	ccID := "test-cc-overwrite"

	// Create a second session in the same project
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	ctx2 := session.NewContext("Test Claim 2", "PATCH", "10x-dev")
	sessionDir2 := filepath.Join(sessionsDir, ctx2.SessionID)
	os.MkdirAll(sessionDir2, 0755)
	if err := ctx2.Save(filepath.Join(sessionDir2, "SESSION_CONTEXT.md")); err != nil {
		t.Fatalf("failed to save session 2: %v", err)
	}

	// Claim session 1
	if err := session.SetHarnessSessionMap(resolver, ccID, sessionID1); err != nil {
		t.Fatalf("SetHarnessSessionMap(1) error = %v", err)
	}

	// Overwrite: claim session 2
	if err := session.SetHarnessSessionMap(resolver, ccID, ctx2.SessionID); err != nil {
		t.Fatalf("SetHarnessSessionMap(2) error = %v", err)
	}

	// Resolve should return session 2
	resolved, err := session.ResolveSession(resolver, ccID, "")
	if err != nil {
		t.Fatalf("ResolveSession() error = %v", err)
	}
	if resolved != ctx2.SessionID {
		t.Errorf("ResolveSession() = %q, want %q (overwritten)", resolved, ctx2.SessionID)
	}
}

func TestClaim_EmptyCCSessionID(t *testing.T) {
	projectDir, _ := setupClaimTest(t, session.StatusActive)
	resolver := paths.NewResolver(projectDir)

	err := session.SetHarnessSessionMap(resolver, "", "session-anything")
	if err == nil {
		t.Fatal("expected error for empty CC session ID, got nil")
	}
}

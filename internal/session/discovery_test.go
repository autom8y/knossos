package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeSessionContext(t *testing.T, sessionsDir, sessionID, status string) {
	t.Helper()
	dir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	content := fmt.Sprintf(`---
schema_version: "2.1"
session_id: %s
status: %s
created_at: %s
initiative: test
complexity: PATCH
active_rite: none
current_phase: requirements
---

# Test Session
`, sessionID, status, time.Now().UTC().Format(time.RFC3339))
	ctxPath := filepath.Join(dir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write session context: %v", err)
	}
}

func TestFindActiveSession_NoSessions(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}
	if result != "" {
		t.Errorf("FindActiveSession() = %q, want empty string", result)
	}
}

func TestFindActiveSession_OneActive(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")
	sessionID := "session-20260205-160414-abc12345"
	writeSessionContext(t, sessionsDir, sessionID, "ACTIVE")

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}
	if result != sessionID {
		t.Errorf("FindActiveSession() = %q, want %q", result, sessionID)
	}
}

func TestFindActiveSession_OneParked(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")
	writeSessionContext(t, sessionsDir, "session-20260205-160414-abc12345", "PARKED")

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}
	if result != "" {
		t.Errorf("FindActiveSession() = %q, want empty (parked is not active)", result)
	}
}

func TestFindActiveSession_MultipleParked_OneActive(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")
	writeSessionContext(t, sessionsDir, "session-20260201-100000-parked01", "PARKED")
	writeSessionContext(t, sessionsDir, "session-20260202-100000-parked02", "PARKED")
	writeSessionContext(t, sessionsDir, "session-20260203-100000-active01", "ACTIVE")

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}
	if result != "session-20260203-100000-active01" {
		t.Errorf("FindActiveSession() = %q, want %q", result, "session-20260203-100000-active01")
	}
}

func TestFindActiveSession_NoActiveDir(t *testing.T) {
	// Sessions directory doesn't exist at all
	result, err := FindActiveSession("/nonexistent/path/sessions")
	if err != nil {
		t.Fatalf("FindActiveSession() should not error on missing dir: %v", err)
	}
	if result != "" {
		t.Errorf("FindActiveSession() = %q, want empty", result)
	}
}

func TestFindActiveSession_CorruptContext(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")

	// Create a corrupt session (no valid frontmatter)
	corruptDir := filepath.Join(sessionsDir, "session-20260201-100000-corrupt1")
	if err := os.MkdirAll(corruptDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(corruptDir, "SESSION_CONTEXT.md"), []byte("garbage"), 0644); err != nil {
		t.Fatalf("Failed to write corrupt context: %v", err)
	}

	// Create a valid ACTIVE session
	writeSessionContext(t, sessionsDir, "session-20260202-100000-valid001", "ACTIVE")

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}
	if result != "session-20260202-100000-valid001" {
		t.Errorf("FindActiveSession() = %q, want %q (should skip corrupt and find valid)", result, "session-20260202-100000-valid001")
	}
}

func TestFindActiveSession_SkipsNonSessionDirs(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	// Create non-session directories that should be skipped
	for _, name := range []string{".locks", ".audit", ".current-session", "not-a-session"} {
		path := filepath.Join(sessionsDir, name)
		os.MkdirAll(path, 0755)
	}

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}
	if result != "" {
		t.Errorf("FindActiveSession() = %q, want empty", result)
	}
}

func TestFindActiveSession_PhantomStatusNotActive(t *testing.T) {
	// A session with phantom status "COMPLETED" should NOT appear as active.
	// After normalization, COMPLETED → ARCHIVED, which != ACTIVE.
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")
	writeSessionContext(t, sessionsDir, "session-20260201-100000-complet1", "COMPLETED")
	writeSessionContext(t, sessionsDir, "session-20260202-100000-complet2", "COMPLETE")

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}
	if result != "" {
		t.Errorf("FindActiveSession() = %q, want empty (phantom statuses should not match ACTIVE)", result)
	}
}

// --- Dual-ACTIVE Session Scenario ---
//
// KNOWN BEHAVIOR: After a crash, two sessions could both have status ACTIVE.
// FindActiveSession enforces single-ACTIVE invariant by detecting multiple
// ACTIVE sessions and returning an error. This prevents silent data corruption
// from race conditions.

func TestFindActiveSession_DualActive(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")

	// Create two sessions both with status ACTIVE
	session1 := "session-20260201-100000-active01"
	session2 := "session-20260202-100000-active02"
	writeSessionContext(t, sessionsDir, session1, "ACTIVE")
	writeSessionContext(t, sessionsDir, session2, "ACTIVE")

	result, err := FindActiveSession(sessionsDir)

	// Must return an error for dual-ACTIVE
	if err == nil {
		t.Fatalf("FindActiveSession() expected error for dual-ACTIVE, got nil")
	}

	// Error message must mention both sessions
	errMsg := err.Error()
	if !strings.Contains(errMsg, "multiple active sessions") {
		t.Errorf("Error message %q should contain 'multiple active sessions'", errMsg)
	}
	if !strings.Contains(errMsg, session1) || !strings.Contains(errMsg, session2) {
		t.Errorf("Error message %q should mention both session IDs: %q, %q", errMsg, session1, session2)
	}

	// Result must be empty string on error
	if result != "" {
		t.Errorf("FindActiveSession() returned %q on error, want empty string", result)
	}

	t.Logf("Dual-ACTIVE correctly detected: %v", err)
}

// --- Cache-Then-Scan Performance Sanity Check ---
//
// The GetCurrentSessionID() in internal/cmd/common/context.go implements
// cache-then-scan with a 5s TTL. This test verifies the scan component
// (FindActiveSession) completes in a reasonable time even with multiple
// sessions. This is a sanity check, not a benchmark.

func TestFindActiveSession_Performance(t *testing.T) {
	tmpDir := t.TempDir()

	sessionsDir := filepath.Join(tmpDir, "sessions")

	// Create 9 PARKED sessions and 1 ACTIVE session (last alphabetically)
	for i := 0; i < 9; i++ {
		sessionID := fmt.Sprintf("session-20260201-10%02d00-parked%02d", i, i)
		writeSessionContext(t, sessionsDir, sessionID, "PARKED")
	}
	activeID := "session-20260201-109900-theactive"
	writeSessionContext(t, sessionsDir, activeID, "ACTIVE")

	// Time the scan
	start := time.Now()
	result, err := FindActiveSession(sessionsDir)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}

	if result != activeID {
		t.Errorf("FindActiveSession() = %q, want %q", result, activeID)
	}

	// Sanity check: 10 sessions should scan in well under 100ms
	// even on slow filesystems
	if elapsed > 100*time.Millisecond {
		t.Errorf("FindActiveSession() took %v for 10 sessions, expected < 100ms", elapsed)
	}

	t.Logf("Performance: scanned 10 sessions in %v", elapsed)
}

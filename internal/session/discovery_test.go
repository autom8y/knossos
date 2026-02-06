package session

import (
	"fmt"
	"os"
	"path/filepath"
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
	tmpDir, err := os.MkdirTemp("", "discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
	tmpDir, err := os.MkdirTemp("", "discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
	tmpDir, err := os.MkdirTemp("", "discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
	tmpDir, err := os.MkdirTemp("", "discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
	tmpDir, err := os.MkdirTemp("", "discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
	tmpDir, err := os.MkdirTemp("", "discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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

// --- Dual-ACTIVE Session Scenario ---
//
// KNOWN BEHAVIOR: After a crash, two sessions could both have status ACTIVE.
// FindActiveSession scans via os.ReadDir (alphabetical order on most OSes)
// and returns the FIRST active session found. There is no conflict detection.
//
// This is documented and accepted behavior for now. The Fray initiative
// (future work) will need to add conflict detection to handle this case
// by either prompting the user or using timestamp-based resolution.

func TestFindActiveSession_DualActive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "discovery-dual-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sessionsDir := filepath.Join(tmpDir, "sessions")

	// Create two sessions both with status ACTIVE
	session1 := "session-20260201-100000-active01"
	session2 := "session-20260202-100000-active02"
	writeSessionContext(t, sessionsDir, session1, "ACTIVE")
	writeSessionContext(t, sessionsDir, session2, "ACTIVE")

	result, err := FindActiveSession(sessionsDir)
	if err != nil {
		t.Fatalf("FindActiveSession() error = %v", err)
	}

	// Must return one of the two — not empty
	if result == "" {
		t.Fatal("FindActiveSession() returned empty string, expected one of the two ACTIVE sessions")
	}

	// Must be one of the two sessions we created
	if result != session1 && result != session2 {
		t.Errorf("FindActiveSession() = %q, want either %q or %q", result, session1, session2)
	}

	// KNOWN BEHAVIOR: scan returns first-found by os.ReadDir order.
	// On most systems, os.ReadDir returns entries sorted alphabetically,
	// so session1 (active01) would be returned first. But we do NOT assert
	// which specific one is returned — only that it IS one of them.
	// The Fray initiative will need to add conflict detection here.
	t.Logf("Dual-ACTIVE: FindActiveSession returned %q (first-found behavior)", result)
}

// --- Cache-Then-Scan Performance Sanity Check ---
//
// The GetCurrentSessionID() in internal/cmd/common/context.go implements
// cache-then-scan with a 5s TTL. This test verifies the scan component
// (FindActiveSession) completes in a reasonable time even with multiple
// sessions. This is a sanity check, not a benchmark.

func TestFindActiveSession_Performance(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "discovery-perf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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

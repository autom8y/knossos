package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/lock"
)

func TestIsLockStale_EmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recover-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	if err := os.WriteFile(lockPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if !isAdvisoryLockStale(lockPath) {
		t.Error("Expected empty lock file to be stale")
	}
}

func TestIsLockStale_FreshJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recover-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	meta := lock.LockMetadata{
		Session:  "session-20260205-160414-abc12345",
		Acquired: time.Now().Unix(),
		Holder:   "ari-session-create",
		Version:  "2",
	}
	data, _ := json.Marshal(meta)

	lockPath := filepath.Join(tmpDir, "test.lock")
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if isAdvisoryLockStale(lockPath) {
		t.Error("Expected fresh JSON lock to NOT be stale")
	}
}

func TestIsLockStale_OldJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recover-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	meta := lock.LockMetadata{
		Session:  "session-20260205-160414-abc12345",
		Acquired: time.Now().Add(-10 * time.Minute).Unix(),
		Holder:   "ari-session-create",
		Version:  "2",
	}
	data, _ := json.Marshal(meta)

	lockPath := filepath.Join(tmpDir, "test.lock")
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if !isAdvisoryLockStale(lockPath) {
		t.Error("Expected old JSON lock (10 min) to be stale")
	}
}

func TestIsLockStale_LegacyPID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recover-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	if err := os.WriteFile(lockPath, []byte("12345\n"), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if !isAdvisoryLockStale(lockPath) {
		t.Error("Expected legacy PID lock to be stale (all legacy locks treated as stale)")
	}
}

func TestIsLockStale_Unparseable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recover-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	if err := os.WriteFile(lockPath, []byte("garbage data here"), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if !isAdvisoryLockStale(lockPath) {
		t.Error("Expected unparseable lock to be stale")
	}
}

func TestIsLockStale_NonexistentFile(t *testing.T) {
	if isAdvisoryLockStale("/nonexistent/path/test.lock") {
		t.Error("Expected nonexistent file to NOT be reported as stale")
	}
}

// --- Intentional Stale Divergence Tests ---
//
// There are TWO isStale-like functions with INTENTIONALLY different behavior:
//
// 1. lock.Manager.isStale() in internal/lock/lock.go
//    - Runs during lock ACQUISITION (hot path)
//    - CONSERVATIVE: avoids breaking a lock held by a live process
//    - Empty file       -> NOT stale (returns false; treats as potentially held)
//    - JSON v2 old      -> stale
//    - Legacy PID alive -> NOT stale (preserves running process's lock)
//    - Legacy PID dead  -> stale
//
// 2. isAdvisoryLockStale() in internal/cmd/session/recover.go
//    - Runs during RECOVERY (explicit user action: "ari session recover")
//    - AGGRESSIVE: user asked for cleanup, so be thorough
//    - Empty file       -> stale (empty lock is garbage; clean it up)
//    - JSON v2 old      -> stale
//    - Legacy PID alive -> stale (legacy format should have been migrated)
//    - Legacy PID dead  -> stale
//
// The divergence is intentional. During normal operation (lock.isStale), we
// must not break a lock held by a running process even if it uses the legacy
// format. During recovery (isLockStale), the user has explicitly requested
// cleanup, and ALL legacy-format locks are treated as stale because they
// should have been migrated to JSON v2 format by now.

func TestStaleDivergence_LegacyPIDAlive(t *testing.T) {
	// This test documents the ONE case where lock.isStale and recover.isLockStale
	// intentionally DISAGREE: a legacy PID lock where the process is still alive.
	//
	// lock.isStale:     returns false (conservative — don't break live process lock)
	// isLockStale:      returns true  (aggressive — legacy format is stale by definition)
	//
	// We use the current process PID (os.Getpid()) as a guaranteed-alive PID.

	tmpDir, err := os.MkdirTemp("", "divergence-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a legacy PID lock file using the current process PID (guaranteed alive)
	alivePID := os.Getpid()
	lockPath := filepath.Join(tmpDir, "divergence.lock")
	if err := os.WriteFile(lockPath, []byte(fmt.Sprintf("%d\n", alivePID)), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// lock.Manager.isStale — CONSERVATIVE: alive PID -> NOT stale
	lockMgr := lock.NewManager(tmpDir)
	lockIsStale := lockMgr.IsStale(lockPath, false)
	if lockIsStale {
		t.Error("lock.isStale should return false for legacy PID with alive process (conservative)")
	}

	// recover.isLockStale — AGGRESSIVE: all legacy -> stale
	recoverIsStale := isAdvisoryLockStale(lockPath)
	if !recoverIsStale {
		t.Error("recover.isLockStale should return true for legacy PID (aggressive: all legacy is stale)")
	}

	// Document the divergence explicitly
	if lockIsStale == recoverIsStale {
		t.Error("Expected lock.isStale and recover.isLockStale to DISAGREE on legacy alive PID — " +
			"this is the intentional divergence point between conservative (acquisition) " +
			"and aggressive (recovery) stale detection")
	}
}

func TestStaleDivergence_EmptyFile(t *testing.T) {
	// Empty lock files are stale in all contexts (stakeholder decision).

	tmpDir, err := os.MkdirTemp("", "divergence-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "empty.lock")
	if err := os.WriteFile(lockPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// lock.Manager.IsStale — empty -> stale
	lockMgr := lock.NewManager(tmpDir)
	lockIsStale := lockMgr.IsStale(lockPath, false)
	if !lockIsStale {
		t.Error("lock.IsStale should return true for empty file")
	}

	// recover.isLockStale — empty -> stale
	recoverIsStale := isAdvisoryLockStale(lockPath)
	if !recoverIsStale {
		t.Error("recover.isLockStale should return true for empty file")
	}
}

func TestStaleDivergence_AgreementCases(t *testing.T) {
	// Documents the cases where both functions AGREE.
	// This ensures the divergence is scoped and not accidental.

	tmpDir, err := os.MkdirTemp("", "divergence-agree-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockMgr := lock.NewManager(tmpDir)

	t.Run("fresh_JSON_both_not_stale", func(t *testing.T) {
		meta := lock.LockMetadata{
			Session:  "session-20260205-160414-abc12345",
			Acquired: time.Now().Unix(),
			Holder:   "ari-session-create",
			Version:  "2",
		}
		data, _ := json.Marshal(meta)
		lockPath := filepath.Join(tmpDir, "fresh.lock")
		if err := os.WriteFile(lockPath, data, 0644); err != nil {
			t.Fatalf("Failed to write lock file: %v", err)
		}

		if lockMgr.IsStale(lockPath, false) {
			t.Error("lock.isStale should return false for fresh JSON")
		}
		if isAdvisoryLockStale(lockPath) {
			t.Error("recover.isLockStale should return false for fresh JSON")
		}
	})

	t.Run("old_JSON_both_stale", func(t *testing.T) {
		meta := lock.LockMetadata{
			Session:  "session-20260205-160414-abc12345",
			Acquired: time.Now().Add(-10 * time.Minute).Unix(),
			Holder:   "ari-session-create",
			Version:  "2",
		}
		data, _ := json.Marshal(meta)
		lockPath := filepath.Join(tmpDir, "old.lock")
		if err := os.WriteFile(lockPath, data, 0644); err != nil {
			t.Fatalf("Failed to write lock file: %v", err)
		}

		if !lockMgr.IsStale(lockPath, false) {
			t.Error("lock.isStale should return true for old JSON")
		}
		if !isAdvisoryLockStale(lockPath) {
			t.Error("recover.isLockStale should return true for old JSON")
		}
	})

	t.Run("dead_PID_both_stale", func(t *testing.T) {
		// Use a very high PID that is almost certainly not running
		lockPath := filepath.Join(tmpDir, "dead.lock")
		if err := os.WriteFile(lockPath, []byte("99999999\n"), 0644); err != nil {
			t.Fatalf("Failed to write lock file: %v", err)
		}

		if !lockMgr.IsStale(lockPath, false) {
			t.Error("lock.isStale should return true for dead PID")
		}
		if !isAdvisoryLockStale(lockPath) {
			t.Error("recover.isLockStale should return true for dead PID (all legacy is stale)")
		}
	})
}

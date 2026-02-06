package session

import (
	"encoding/json"
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

	if !isLockStale(lockPath) {
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

	if isLockStale(lockPath) {
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

	if !isLockStale(lockPath) {
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

	if !isLockStale(lockPath) {
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

	if !isLockStale(lockPath) {
		t.Error("Expected unparseable lock to be stale")
	}
}

func TestIsLockStale_NonexistentFile(t *testing.T) {
	if isLockStale("/nonexistent/path/test.lock") {
		t.Error("Expected nonexistent file to NOT be reported as stale")
	}
}

package lock

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadMoiraiLock(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		lockPath := filepath.Join(tmpDir, ".moirai-lock")

		ml := MoiraiLock{
			Agent:             "moirai",
			AcquiredAt:        time.Now().UTC(),
			SessionID:         "session-123",
			StaleAfterSeconds: 300,
		}
		data, _ := json.Marshal(ml)
		if err := os.WriteFile(lockPath, data, 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ReadMoiraiLock(lockPath)
		if err != nil {
			t.Fatalf("ReadMoiraiLock() unexpected error: %v", err)
		}
		if got.Agent != "moirai" {
			t.Errorf("Agent = %q, want %q", got.Agent, "moirai")
		}
		if got.SessionID != "session-123" {
			t.Errorf("SessionID = %q, want %q", got.SessionID, "session-123")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		lockPath := filepath.Join(tmpDir, ".moirai-lock")

		if err := os.WriteFile(lockPath, []byte("not json"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ReadMoiraiLock(lockPath)
		if err == nil {
			t.Error("ReadMoiraiLock() expected error for invalid JSON")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := ReadMoiraiLock("/nonexistent/path/.moirai-lock")
		if err == nil {
			t.Error("ReadMoiraiLock() expected error for missing file")
		}
	})
}

func TestIsMoiraiLockStale(t *testing.T) {
	t.Run("fresh lock", func(t *testing.T) {
		ml := &MoiraiLock{
			AcquiredAt:        time.Now().UTC(),
			StaleAfterSeconds: 300,
		}
		if IsMoiraiLockStale(ml) {
			t.Error("IsMoiraiLockStale() = true for fresh lock, want false")
		}
	})

	t.Run("expired lock", func(t *testing.T) {
		ml := &MoiraiLock{
			AcquiredAt:        time.Now().UTC().Add(-10 * time.Minute),
			StaleAfterSeconds: 300,
		}
		if !IsMoiraiLockStale(ml) {
			t.Error("IsMoiraiLockStale() = false for expired lock, want true")
		}
	})
}

func TestIsStale_TreatLegacyAsStale(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	lockPath := filepath.Join(tmpDir, "legacy.lock")
	// Write a legacy PID-format lock (current process PID, which is alive)
	if err := os.WriteFile(lockPath, []byte("1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// With treatLegacyAsStale=true, should be stale regardless of process liveness
	if !mgr.IsStale(lockPath, true) {
		t.Error("IsStale(treatLegacyAsStale=true) should return true for legacy PID lock")
	}
}

func TestIsStale_LegacyProcessLiveness(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Use a PID that is almost certainly dead
	lockPath := filepath.Join(tmpDir, "dead.lock")
	if err := os.WriteFile(lockPath, []byte("99999999\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// With treatLegacyAsStale=false, dead process should be stale
	if !mgr.IsStale(lockPath, false) {
		t.Error("IsStale(treatLegacyAsStale=false) should return true for dead PID")
	}
}

func TestIsStale_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	lockPath := filepath.Join(tmpDir, "empty.lock")
	if err := os.WriteFile(lockPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Empty files should always be stale
	if !mgr.IsStale(lockPath, false) {
		t.Error("IsStale() should return true for empty file")
	}
	if !mgr.IsStale(lockPath, true) {
		t.Error("IsStale(treatLegacyAsStale=true) should return true for empty file")
	}
}

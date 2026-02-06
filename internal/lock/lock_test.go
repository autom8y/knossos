package lock

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestManager_AcquireRelease(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "test-session"

	lock, err := mgr.Acquire(sessionID, Exclusive, 5*time.Second, "test-acquire")
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	// Check lock file exists
	lockPath := filepath.Join(tmpDir, sessionID+".lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file should exist after acquire")
	}

	if err := lock.Release(); err != nil {
		t.Errorf("Release() error = %v", err)
	}
}

func TestManager_ExclusiveLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "exclusive-test"

	lock1, err := mgr.Acquire(sessionID, Exclusive, 5*time.Second, "test-lock1")
	if err != nil {
		t.Fatalf("First Acquire() error = %v", err)
	}
	defer lock1.Release()

	// Second lock should timeout
	_, err = mgr.Acquire(sessionID, Exclusive, 100*time.Millisecond, "test-lock2")
	if err == nil {
		t.Error("Second Acquire() should fail while first lock is held")
	}
}

func TestManager_SharedLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "shared-test"

	lock1, err := mgr.Acquire(sessionID, Shared, 5*time.Second, "test-shared1")
	if err != nil {
		t.Fatalf("First shared Acquire() error = %v", err)
	}
	defer lock1.Release()

	lock2, err := mgr.Acquire(sessionID, Shared, 5*time.Second, "test-shared2")
	if err != nil {
		t.Errorf("Second shared Acquire() should succeed: %v", err)
	} else {
		lock2.Release()
	}
}

func TestManager_IsLocked(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "islock-test"

	if mgr.IsLocked(sessionID) {
		t.Error("IsLocked() should return false when not locked")
	}

	lock, err := mgr.Acquire(sessionID, Exclusive, 5*time.Second, "test-islock")
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	lock.Release()
}

func TestManager_ConcurrentPark(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "concurrent-park"

	var wg sync.WaitGroup
	results := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			lock, err := mgr.Acquire(sessionID, Exclusive, 5*time.Second, "test-concurrent")
			if err != nil {
				results <- err
				return
			}

			time.Sleep(100 * time.Millisecond)

			lock.Release()
			results <- nil
		}(i)
	}

	wg.Wait()
	close(results)

	var successes, failures int
	for err := range results {
		if err == nil {
			successes++
		} else {
			failures++
		}
	}

	if successes != 2 {
		t.Errorf("Expected 2 successes, got %d (failures: %d)", successes, failures)
	}
}

func TestManager_RaceCondition(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "race-test"

	var wg sync.WaitGroup
	var mu sync.Mutex
	counter := 0
	overlaps := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			lock, err := mgr.Acquire(sessionID, Exclusive, 10*time.Second, "test-race")
			if err != nil {
				return
			}

			mu.Lock()
			current := counter
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			if counter != current {
				overlaps++
			}
			counter = current + 1
			mu.Unlock()

			lock.Release()
		}()
	}

	wg.Wait()

	mu.Lock()
	finalCounter := counter
	finalOverlaps := overlaps
	mu.Unlock()

	if finalCounter != 10 {
		t.Errorf("Counter = %d, expected 10", finalCounter)
	}
	if finalOverlaps > 0 {
		t.Errorf("Overlaps = %d, expected 0 (lock not serializing properly)", finalOverlaps)
	}
}

func TestLock_Metadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "props-test"

	lock, err := mgr.Acquire(sessionID, Exclusive, 5*time.Second, "test-props")
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()

	if lock.SessionID() != sessionID {
		t.Errorf("SessionID() = %q, want %q", lock.SessionID(), sessionID)
	}

	if lock.Path() == "" {
		t.Error("Path() should not be empty")
	}

	meta := lock.Metadata()
	if meta == nil {
		t.Fatal("Metadata() should not be nil")
	}
	if meta.Holder != "test-props" {
		t.Errorf("Metadata().Holder = %q, want %q", meta.Holder, "test-props")
	}
	if meta.Session != sessionID {
		t.Errorf("Metadata().Session = %q, want %q", meta.Session, sessionID)
	}
	if meta.Version != "2" {
		t.Errorf("Metadata().Version = %q, want %q", meta.Version, "2")
	}
}

func TestManager_ForceRelease(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "force-release"

	lockPath := filepath.Join(tmpDir, sessionID+".lock")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	if err := os.WriteFile(lockPath, []byte("12345\n"), 0644); err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}

	if err := mgr.ForceRelease(sessionID); err != nil {
		t.Errorf("ForceRelease() error = %v", err)
	}

	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("Lock file should be removed after ForceRelease")
	}
}

// --- TDD: JSON Lock Format Tests ---

func TestLockFileFormat_JSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "json-format-test"

	lock, err := mgr.Acquire(sessionID, Exclusive, 5*time.Second, "ari-session-create")
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()

	// Read lock file and verify JSON content
	lockPath := filepath.Join(tmpDir, sessionID+".lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	var meta LockMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Lock file is not valid JSON: %v", err)
	}

	if meta.Session != sessionID {
		t.Errorf("session = %q, want %q", meta.Session, sessionID)
	}
	if meta.Holder != "ari-session-create" {
		t.Errorf("holder = %q, want %q", meta.Holder, "ari-session-create")
	}
	if meta.Version != "2" {
		t.Errorf("version = %q, want %q", meta.Version, "2")
	}
	if meta.Acquired <= 0 {
		t.Errorf("acquired should be a positive unix timestamp, got %d", meta.Acquired)
	}

	// Verify acquired is approximately now
	now := time.Now().Unix()
	if meta.Acquired < now-5 || meta.Acquired > now+5 {
		t.Errorf("acquired timestamp %d is not near now %d", meta.Acquired, now)
	}
}

func TestLockFileFormat_ParseLegacy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "legacy-test"

	// Write a legacy PID-only lock file
	lockPath := filepath.Join(tmpDir, sessionID+".lock")
	if err := os.WriteFile(lockPath, []byte("12345\n"), 0644); err != nil {
		t.Fatalf("Failed to write legacy lock: %v", err)
	}

	// GetLockInfo should parse legacy format and synthesize metadata
	meta, err := mgr.GetLockInfo(sessionID)
	if err != nil {
		t.Fatalf("GetLockInfo() error = %v", err)
	}

	if meta.Version != "1" {
		t.Errorf("Version = %q, want %q for legacy lock", meta.Version, "1")
	}
	if meta.Holder != "legacy-pid-12345" {
		t.Errorf("Holder = %q, want %q", meta.Holder, "legacy-pid-12345")
	}
}

func TestLockFileFormat_BackwardCompat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "compat-test"

	// Write a legacy PID lock with a dead PID
	lockPath := filepath.Join(tmpDir, sessionID+".lock")
	if err := os.WriteFile(lockPath, []byte("99999999\n"), 0644); err != nil {
		t.Fatalf("Failed to write legacy lock: %v", err)
	}

	// isStale should detect dead PID in legacy format
	stale := mgr.isStale(lockPath)
	if !stale {
		t.Error("Legacy lock with dead PID should be detected as stale")
	}
}

func TestLockStale_AgeBased(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "stale-age-test"

	// Write a JSON lock file with old timestamp (10 minutes ago)
	lockPath := filepath.Join(tmpDir, sessionID+".lock")
	meta := LockMetadata{
		Session:  sessionID,
		Acquired: time.Now().Add(-10 * time.Minute).Unix(),
		Holder:   "ari-session-create",
		Version:  "2",
	}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if !mgr.isStale(lockPath) {
		t.Error("Lock older than StaleThreshold should be stale")
	}

	// Write a JSON lock with recent timestamp (1 second ago)
	meta.Acquired = time.Now().Add(-1 * time.Second).Unix()
	data, _ = json.Marshal(meta)
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if mgr.isStale(lockPath) {
		t.Error("Recent lock should NOT be stale")
	}
}

func TestLockStale_Unparseable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	lockPath := filepath.Join(tmpDir, "corrupt.lock")

	// Write garbage content
	if err := os.WriteFile(lockPath, []byte("not-a-number-not-json"), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	if !mgr.isStale(lockPath) {
		t.Error("Unparseable lock file should be treated as stale")
	}
}

func TestGetLockInfo_JSONFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	sessionID := "info-test"

	// Write a v2 JSON lock file
	lockPath := filepath.Join(tmpDir, sessionID+".lock")
	expected := LockMetadata{
		Session:  sessionID,
		Acquired: time.Now().Unix(),
		Holder:   "ari-session-park",
		Version:  "2",
	}
	data, _ := json.Marshal(expected)
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	meta, err := mgr.GetLockInfo(sessionID)
	if err != nil {
		t.Fatalf("GetLockInfo() error = %v", err)
	}

	if meta.Session != expected.Session {
		t.Errorf("Session = %q, want %q", meta.Session, expected.Session)
	}
	if meta.Holder != expected.Holder {
		t.Errorf("Holder = %q, want %q", meta.Holder, expected.Holder)
	}
	if meta.Version != "2" {
		t.Errorf("Version = %q, want %q", meta.Version, "2")
	}
}

func TestGetLockInfo_NoLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	_, err = mgr.GetLockInfo("nonexistent")
	if err == nil {
		t.Error("GetLockInfo() should return error for nonexistent lock")
	}
}

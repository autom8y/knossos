package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/session"
)

func TestLockCreation(t *testing.T) {
	td := setupTestEnv(t)
	defer td.cleanup()

	// Create active session
	sessionID := td.createSession("test-lock")

	// Run lock command
	ctx := td.newContext()
	opts := lockOptions{agent: "moirai"}
	err := runLock(ctx, opts)
	if err != nil {
		t.Fatalf("runLock failed: %v", err)
	}

	// Verify lock file exists
	lockPath := filepath.Join(td.sessionDir(sessionID), moiraiLockFilename)
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Fatalf("lock file not created at %s", lockPath)
	}

	// Verify lock contents
	lock, err := readMoiraiLock(lockPath)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}

	if lock.Agent != "moirai" {
		t.Errorf("expected agent 'moirai', got %s", lock.Agent)
	}
	if lock.SessionID != sessionID {
		t.Errorf("expected session_id %s, got %s", sessionID, lock.SessionID)
	}
	if lock.StaleAfterSeconds != staleAfterSeconds {
		t.Errorf("expected stale_after_seconds %d, got %d", staleAfterSeconds, lock.StaleAfterSeconds)
	}
	if lock.AcquiredAt.IsZero() {
		t.Error("acquired_at is zero")
	}
}

func TestLockRemoval(t *testing.T) {
	td := setupTestEnv(t)
	defer td.cleanup()

	// Create active session and lock
	sessionID := td.createSession("test-unlock")
	lockPath := filepath.Join(td.sessionDir(sessionID), moiraiLockFilename)

	lock := MoiraiLock{
		Agent:             "moirai",
		AcquiredAt:        time.Now().UTC(),
		SessionID:         sessionID,
		StaleAfterSeconds: staleAfterSeconds,
	}
	data, _ := json.Marshal(lock)
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("failed to create test lock: %v", err)
	}

	// Run unlock command
	ctx := td.newContext()
	opts := lockOptions{agent: "moirai"}
	err := runUnlock(ctx, opts)
	if err != nil {
		t.Fatalf("runUnlock failed: %v", err)
	}

	// Verify lock file removed
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatal("lock file still exists after unlock")
	}
}

func TestDoubleLockError(t *testing.T) {
	td := setupTestEnv(t)
	defer td.cleanup()

	sessionID := td.createSession("test-double-lock")
	lockPath := filepath.Join(td.sessionDir(sessionID), moiraiLockFilename)

	// Create existing lock
	lock := MoiraiLock{
		Agent:             "moirai",
		AcquiredAt:        time.Now().UTC(),
		SessionID:         sessionID,
		StaleAfterSeconds: staleAfterSeconds,
	}
	data, _ := json.Marshal(lock)
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("failed to create initial lock: %v", err)
	}

	// Try to acquire lock again
	ctx := td.newContext()
	opts := lockOptions{agent: "moirai"}
	err := runLock(ctx, opts)
	if err == nil {
		t.Fatal("expected error when acquiring existing lock, got nil")
	}
}

func TestStaleLockDetection(t *testing.T) {
	td := setupTestEnv(t)
	defer td.cleanup()

	sessionID := td.createSession("test-stale-lock")
	lockPath := filepath.Join(td.sessionDir(sessionID), moiraiLockFilename)

	// Create stale lock (>300s old)
	staleLock := MoiraiLock{
		Agent:             "moirai",
		AcquiredAt:        time.Now().UTC().Add(-400 * time.Second),
		SessionID:         sessionID,
		StaleAfterSeconds: staleAfterSeconds,
	}
	data, _ := json.Marshal(staleLock)
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("failed to create stale lock: %v", err)
	}

	// Acquire lock should succeed with stale lock (overwrites with warning)
	ctx := td.newContext()
	opts := lockOptions{agent: "moirai"}
	err := runLock(ctx, opts)
	if err != nil {
		t.Fatalf("expected stale lock to be overwritten, got error: %v", err)
	}

	// Verify new lock has recent timestamp
	newLock, err := readMoiraiLock(lockPath)
	if err != nil {
		t.Fatalf("failed to read new lock: %v", err)
	}
	age := time.Since(newLock.AcquiredAt)
	if age > 5*time.Second {
		t.Errorf("new lock timestamp is not recent: %v old", age)
	}
}

func TestAgentValidation(t *testing.T) {
	td := setupTestEnv(t)
	defer td.cleanup()

	td.createSession("test-agent-validation")

	// Try invalid agent name
	ctx := td.newContext()
	opts := lockOptions{agent: "invalid"}
	err := runLock(ctx, opts)
	if err == nil {
		t.Fatal("expected error with invalid agent name, got nil")
	}
}

func TestUnlockWithNonMoiraiLock(t *testing.T) {
	td := setupTestEnv(t)
	defer td.cleanup()

	sessionID := td.createSession("test-wrong-agent")
	lockPath := filepath.Join(td.sessionDir(sessionID), moiraiLockFilename)

	// Create lock with different agent (simulating future multi-agent support)
	lock := MoiraiLock{
		Agent:             "other",
		AcquiredAt:        time.Now().UTC(),
		SessionID:         sessionID,
		StaleAfterSeconds: staleAfterSeconds,
	}
	data, _ := json.Marshal(lock)
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("failed to create test lock: %v", err)
	}

	// Try to unlock with moirai agent
	ctx := td.newContext()
	opts := lockOptions{agent: "moirai"}
	err := runUnlock(ctx, opts)
	if err == nil {
		t.Fatal("expected error when unlocking with wrong agent, got nil")
	}
}

func TestUnlockWithNoLock(t *testing.T) {
	td := setupTestEnv(t)
	defer td.cleanup()

	td.createSession("test-no-lock")

	// Try to unlock when no lock exists
	ctx := td.newContext()
	opts := lockOptions{agent: "moirai"}
	err := runUnlock(ctx, opts)
	if err == nil {
		t.Fatal("expected error when no lock exists, got nil")
	}
}

// Test helper for session lock tests
type lockTestData struct {
	t              *testing.T
	projectDir     string
	outputFlag     string
	verboseFlag    bool
	sessionIDValue string
}

func setupTestEnv(t *testing.T) *lockTestData {
	projectDir := t.TempDir()

	// Create .claude/sessions directory
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	return &lockTestData{
		t:           t,
		projectDir:  projectDir,
		outputFlag:  "json",
		verboseFlag: false,
	}
}

func (td *lockTestData) cleanup() {
	os.RemoveAll(td.projectDir)
}

func (td *lockTestData) newContext() *cmdContext {
	return &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &td.outputFlag,
				Verbose:    &td.verboseFlag,
				ProjectDir: &td.projectDir,
			},
			SessionID: &td.sessionIDValue,
		},
	}
}

func (td *lockTestData) sessionDir(sessionID string) string {
	return filepath.Join(td.projectDir, ".claude", "sessions", sessionID)
}

func (td *lockTestData) createSession(initiative string) string {
	sessionID := "session-test-" + time.Now().Format("20060102-150405")
	td.sessionIDValue = sessionID

	sessionDir := td.sessionDir(sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		td.t.Fatalf("failed to create session dir: %v", err)
	}

	// Create minimal SESSION_CONTEXT.md
	ctx := &session.Context{
		SessionID:  sessionID,
		Initiative: initiative,
		Status:     session.StatusActive,
		CreatedAt:  time.Now().UTC(),
	}
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := ctx.Save(ctxPath); err != nil {
		td.t.Fatalf("failed to save session context: %v", err)
	}

	// Set as current session
	currentPath := filepath.Join(td.projectDir, ".claude", "sessions", ".current-session")
	if err := os.WriteFile(currentPath, []byte(sessionID), 0644); err != nil {
		td.t.Fatalf("failed to write current session: %v", err)
	}

	return sessionID
}

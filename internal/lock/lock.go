// Package lock provides advisory file locking with stale detection.
// It implements the concurrency model from TDD Section 5.
package lock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/autom8y/knossos/internal/errors"
)

// LockType represents the type of lock (shared or exclusive).
type LockType int

const (
	// Shared allows multiple readers.
	Shared LockType = syscall.LOCK_SH
	// Exclusive allows only one writer.
	Exclusive LockType = syscall.LOCK_EX
)

// DefaultTimeout is the default lock acquisition timeout.
const DefaultTimeout = 10 * time.Second

// StaleThreshold is the age after which a lock is considered stale.
const StaleThreshold = 5 * time.Minute

// LockMetadata is the JSON structure written to lock files.
type LockMetadata struct {
	Session  string `json:"session"`           // Session ID being operated on
	Acquired int64  `json:"acquired"`          // Unix timestamp of acquisition
	Holder   string `json:"holder"`            // Command name (e.g., "ari-session-create")
	Version  string `json:"version,omitempty"` // Lock format version "2"
}

// Lock represents an acquired file lock.
type Lock struct {
	sessionID string
	file      *os.File
	lockType  LockType
	lockPath  string
	metadata  *LockMetadata
}

// Manager handles lock operations for sessions.
type Manager struct {
	locksDir string
}

// NewManager creates a new lock manager for the given locks directory.
func NewManager(locksDir string) *Manager {
	return &Manager{locksDir: locksDir}
}

// lockFilePath returns the path to a session's lock file.
func (m *Manager) lockFilePath(sessionID string) string {
	return filepath.Join(m.locksDir, sessionID+".lock")
}

// Acquire attempts to acquire a lock with the given timeout.
// The holder parameter identifies the command acquiring the lock (e.g., "ari-session-create").
// Returns a Lock that must be released when done.
func (m *Manager) Acquire(sessionID string, lockType LockType, timeout time.Duration, holder string) (*Lock, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	lockPath := m.lockFilePath(sessionID)

	// Ensure lock directory exists
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create lock directory", err)
	}

	// Open or create lock file
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to open lock file", err)
	}

	// Attempt lock with timeout
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Try non-blocking lock
		err := syscall.Flock(int(file.Fd()), int(lockType)|syscall.LOCK_NB)
		if err == nil {
			// Lock acquired successfully
			meta := &LockMetadata{
				Session:  sessionID,
				Acquired: time.Now().Unix(),
				Holder:   holder,
				Version:  "2",
			}
			if lockType == Exclusive {
				// Write JSON metadata
				file.Truncate(0)
				file.Seek(0, 0)
				data, _ := json.Marshal(meta)
				file.Write(data)
				file.Write([]byte("\n"))
			}
			return &Lock{
				sessionID: sessionID,
				file:      file,
				lockType:  lockType,
				lockPath:  lockPath,
				metadata:  meta,
			}, nil
		}

		// Check for stale lock — reclaim atomically via flock on existing fd.
		// This eliminates the TOCTOU race in the old remove+reopen pattern:
		// we already hold `file` open, so try to flock IT directly.
		if m.IsStale(lockPath, false) {
			if flockErr := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); flockErr == nil {
				// We acquired the flock — reclaim by rewriting metadata
				meta := &LockMetadata{
					Session:  sessionID,
					Acquired: time.Now().Unix(),
					Holder:   holder,
					Version:  "2",
				}
				file.Truncate(0)
				file.Seek(0, 0)
				data, _ := json.Marshal(meta)
				file.Write(data)
				file.Write([]byte("\n"))
				return &Lock{
					sessionID: sessionID,
					file:      file,
					lockType:  lockType,
					lockPath:  lockPath,
					metadata:  meta,
				}, nil
			}
			// Another process beat us to the reclaim — fall through to retry
		}

		// Wait before retry
		time.Sleep(100 * time.Millisecond)
	}

	// Timeout - get holder info for error message
	file.Close()
	meta, _ := m.GetLockInfo(sessionID)
	return nil, errors.ErrLockTimeout(lockPath, meta)
}

// Release releases the lock.
func (l *Lock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}

	// Release flock
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	l.file.Close()
	l.file = nil

	return nil
}

// IsStale checks if an advisory lock file should be considered stale.
// When treatLegacyAsStale is true, all legacy PID-format locks are
// treated as stale (suitable for recovery operations).
// When false, legacy locks are checked for process liveness.
// Empty lock files are always considered stale.
func (m *Manager) IsStale(lockPath string, treatLegacyAsStale bool) bool {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return false
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return true // Empty lock file is always stale
	}

	// Try JSON format first (v2)
	var meta LockMetadata
	if json.Unmarshal(data, &meta) == nil && meta.Version == "2" {
		// Age-based stale check
		acquired := time.Unix(meta.Acquired, 0)
		return time.Since(acquired) > StaleThreshold
	}

	// Legacy PID format
	if treatLegacyAsStale {
		return true // Recovery mode: treat all legacy locks as stale
	}

	// Try legacy PID format with process liveness check
	pid, err := strconv.Atoi(content)
	if err != nil {
		// Unparseable — treat as stale
		return true
	}

	if pid <= 0 {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return true
	}

	// On Unix, FindProcess always succeeds; check with signal 0
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return true // Process doesn't exist
	}

	return false
}

// getLockMetadata reads and parses lock metadata from a lock file.
func (m *Manager) getLockMetadata(lockPath string) (*LockMetadata, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, fmt.Errorf("empty lock file")
	}

	// Try JSON format first (v2)
	var meta LockMetadata
	if json.Unmarshal(data, &meta) == nil && meta.Version == "2" {
		return &meta, nil
	}

	// Try legacy PID format — synthesize metadata
	pid, err := strconv.Atoi(content)
	if err != nil {
		return nil, fmt.Errorf("unparseable lock file")
	}

	// Get file mod time as approximate acquired time
	info, _ := os.Stat(lockPath)
	acquired := time.Now().Unix()
	if info != nil {
		acquired = info.ModTime().Unix()
	}

	return &LockMetadata{
		Holder:   fmt.Sprintf("legacy-pid-%d", pid),
		Acquired: acquired,
		Version:  "1",
	}, nil
}

// IsLocked checks if a session is currently locked.
func (m *Manager) IsLocked(sessionID string) bool {
	lockPath := m.lockFilePath(sessionID)

	// Try to acquire a non-blocking exclusive lock
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return false
	}
	defer file.Close()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		// Could not acquire lock - it's held by someone
		return true
	}

	// Release immediately
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	return false
}

// GetLockInfo returns the metadata for a session's lock, if any.
func (m *Manager) GetLockInfo(sessionID string) (*LockMetadata, error) {
	lockPath := m.lockFilePath(sessionID)
	return m.getLockMetadata(lockPath)
}

// ForceRelease forcibly removes a lock file (for stale lock cleanup).
func (m *Manager) ForceRelease(sessionID string) error {
	lockPath := m.lockFilePath(sessionID)
	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.CodeGeneralError, "failed to remove lock file", err)
	}
	return nil
}

// IsStaleFile checks if a lock file at the given path is stale.
// This is a convenience function that does not require a Manager instance.
// When treatLegacyAsStale is true, all legacy PID-format locks are treated as stale.
func IsStaleFile(lockPath string, treatLegacyAsStale bool) bool {
	m := &Manager{}
	return m.IsStale(lockPath, treatLegacyAsStale)
}

// LocksDir returns the locks directory path.
func (m *Manager) LocksDir() string {
	return m.locksDir
}

// SessionID returns the session ID this lock is for.
func (l *Lock) SessionID() string {
	return l.sessionID
}

// Path returns the path to the lock file.
func (l *Lock) Path() string {
	return l.lockPath
}

// Metadata returns the lock metadata.
func (l *Lock) Metadata() *LockMetadata {
	return l.metadata
}

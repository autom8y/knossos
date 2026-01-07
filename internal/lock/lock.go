// Package lock provides advisory file locking with stale detection.
// It implements the concurrency model from TDD Section 5.
package lock

import (
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

// Lock represents an acquired file lock.
type Lock struct {
	sessionID string
	file      *os.File
	lockType  LockType
	lockPath  string
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
// Returns a Lock that must be released when done.
func (m *Manager) Acquire(sessionID string, lockType LockType, timeout time.Duration) (*Lock, error) {
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
			if lockType == Exclusive {
				// Write PID for debugging and stale detection
				file.Truncate(0)
				file.Seek(0, 0)
				fmt.Fprintf(file, "%d\n", os.Getpid())
			}
			return &Lock{
				sessionID: sessionID,
				file:      file,
				lockType:  lockType,
				lockPath:  lockPath,
			}, nil
		}

		// Check for stale lock
		if m.isStale(lockPath) {
			// Force remove stale lock and retry
			os.Remove(lockPath)
			file.Close()
			file, err = os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to reopen lock file", err)
			}
			continue
		}

		// Wait before retry
		time.Sleep(100 * time.Millisecond)
	}

	// Timeout - get holder info for error message
	file.Close()
	holderPID := m.getHolderPID(lockPath)
	return nil, errors.ErrLockTimeout(lockPath, holderPID)
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

// isStale checks if the lock holder process is dead.
func (m *Manager) isStale(lockPath string) bool {
	pid := m.getHolderPID(lockPath)
	if pid <= 0 {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return true // Can't find process
	}

	// On Unix, FindProcess always succeeds; check with signal 0
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return true // Process doesn't exist
	}

	return false
}

// getHolderPID reads the PID from a lock file.
func (m *Manager) getHolderPID(lockPath string) int {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return 0
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}

	return pid
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

// GetHolder returns the PID of the lock holder, if any.
func (m *Manager) GetHolder(sessionID string) (int, error) {
	lockPath := m.lockFilePath(sessionID)
	pid := m.getHolderPID(lockPath)
	if pid <= 0 {
		return 0, errors.New(errors.CodeFileNotFound, "no lock holder found")
	}
	return pid, nil
}

// ForceRelease forcibly removes a lock file (for stale lock cleanup).
func (m *Manager) ForceRelease(sessionID string) error {
	lockPath := m.lockFilePath(sessionID)
	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.CodeGeneralError, "failed to remove lock file", err)
	}
	return nil
}

// SessionID returns the session ID this lock is for.
func (l *Lock) SessionID() string {
	return l.sessionID
}

// Path returns the path to the lock file.
func (l *Lock) Path() string {
	return l.lockPath
}

// HolderPID returns the PID of the lock holder (this process for exclusive locks).
func (l *Lock) HolderPID() int {
	if l.lockType == Exclusive {
		return os.Getpid()
	}
	return 0
}

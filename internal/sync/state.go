// Package sync provides sync state management for Ariadne.
// It handles tracking of synced files, checksums, and conflict detection.
package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// State represents the sync state stored in .claude/sync/state.json.
type State struct {
	SchemaVersion string                 `json:"schema_version"`
	Remote        string                 `json:"remote"`
	ActiveRite    string                 `json:"active_rite,omitempty"`
	LastSync      time.Time              `json:"last_sync"`
	TrackedFiles  map[string]TrackedFile `json:"tracked_files"`
	Conflicts     []Conflict             `json:"conflicts,omitempty"`
}

// TrackedFile represents a file being tracked for sync.
type TrackedFile struct {
	Path         string    `json:"path"`
	LocalHash    string    `json:"local_hash"`
	RemoteHash   string    `json:"remote_hash"`
	BaseHash     string    `json:"base_hash"` // Common ancestor hash for conflict detection
	LastModified time.Time `json:"last_modified"`
	Status       string    `json:"status"` // synced, modified, conflict
}

// Conflict represents a sync conflict.
type Conflict struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	LocalHash   string `json:"local_hash"`
	RemoteHash  string `json:"remote_hash"`
	BaseHash    string `json:"base_hash"`
	DetectedAt  string `json:"detected_at"`
}

// CurrentSchemaVersion is the current state schema version.
const CurrentSchemaVersion = "1.0"

// StateManager handles sync state operations.
type StateManager struct {
	resolver    *paths.Resolver
	syncDirPath string // optional override for staging
}

// NewStateManager creates a new state manager.
func NewStateManager(resolver *paths.Resolver) *StateManager {
	return &StateManager{resolver: resolver}
}

// SetSyncDir overrides the sync directory path (used during staged materialization).
func (m *StateManager) SetSyncDir(dir string) {
	m.syncDirPath = dir
}

// SyncDir returns the path to the .claude/sync directory.
func (m *StateManager) SyncDir() string {
	if m.syncDirPath != "" {
		return m.syncDirPath
	}
	return filepath.Join(m.resolver.ClaudeDir(), "sync")
}

// StatePath returns the path to state.json.
func (m *StateManager) StatePath() string {
	return filepath.Join(m.SyncDir(), "state.json")
}

// HistoryPath returns the path to history.json.
func (m *StateManager) HistoryPath() string {
	return filepath.Join(m.SyncDir(), "history.json")
}

// EnsureSyncDir creates the sync directory if it doesn't exist.
func (m *StateManager) EnsureSyncDir() error {
	return paths.EnsureDir(m.SyncDir())
}

// Load reads the sync state from disk.
func (m *StateManager) Load() (*State, error) {
	statePath := m.StatePath()
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if not initialized
			return nil, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read sync state", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, errors.ErrSyncStateCorrupt(statePath, err.Error())
	}

	// Validate schema version
	if state.SchemaVersion == "" {
		return nil, errors.ErrSyncStateCorrupt(statePath, "missing schema_version")
	}

	return &state, nil
}

// Save writes the sync state to disk.
func (m *StateManager) Save(state *State) error {
	if err := m.EnsureSyncDir(); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create sync directory", err)
	}

	state.SchemaVersion = CurrentSchemaVersion

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal sync state", err)
	}

	statePath := m.StatePath()
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write sync state", err)
	}

	return nil
}

// Initialize creates a new sync state for the given remote.
func (m *StateManager) Initialize(remote string) (*State, error) {
	state := &State{
		SchemaVersion: CurrentSchemaVersion,
		Remote:        remote,
		LastSync:      time.Now().UTC(),
		TrackedFiles:  make(map[string]TrackedFile),
		Conflicts:     []Conflict{},
	}

	if err := m.Save(state); err != nil {
		return nil, err
	}

	return state, nil
}

// Reset clears the sync state.
func (m *StateManager) Reset() error {
	statePath := m.StatePath()
	if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.CodeGeneralError, "failed to remove sync state", err)
	}
	return nil
}

// IsInitialized checks if sync has been initialized.
func (m *StateManager) IsInitialized() bool {
	_, err := os.Stat(m.StatePath())
	return err == nil
}

// ComputeFileHash computes the SHA-256 hash of a file.
func ComputeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Return empty hash for missing files
		}
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeContentHash computes the SHA-256 hash of content bytes.
func ComputeContentHash(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

// DetectChanges checks if a file has changed since last sync.
func (m *StateManager) DetectChanges(state *State, path string) (bool, string, error) {
	tracked, exists := state.TrackedFiles[path]
	if !exists {
		return true, "untracked", nil
	}

	currentHash, err := ComputeFileHash(filepath.Join(m.resolver.ProjectRoot(), path))
	if err != nil {
		return false, "", err
	}

	if currentHash == "" {
		return true, "deleted", nil
	}

	if currentHash != tracked.LocalHash {
		return true, "modified", nil
	}

	return false, "synced", nil
}

// AddConflict records a new conflict.
func (m *StateManager) AddConflict(state *State, path, localHash, remoteHash, baseHash, description string) {
	conflict := Conflict{
		Path:        path,
		Description: description,
		LocalHash:   localHash,
		RemoteHash:  remoteHash,
		BaseHash:    baseHash,
		DetectedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	state.Conflicts = append(state.Conflicts, conflict)
}

// RemoveConflict removes a conflict by path.
func (m *StateManager) RemoveConflict(state *State, path string) bool {
	for i, c := range state.Conflicts {
		if c.Path == path {
			state.Conflicts = append(state.Conflicts[:i], state.Conflicts[i+1:]...)
			return true
		}
	}
	return false
}

// HasConflicts returns true if there are unresolved conflicts.
func (state *State) HasConflicts() bool {
	return len(state.Conflicts) > 0
}

// GetConflict returns the conflict for a path, if any.
func (state *State) GetConflict(path string) *Conflict {
	for i := range state.Conflicts {
		if state.Conflicts[i].Path == path {
			return &state.Conflicts[i]
		}
	}
	return nil
}

// UpdateTrackedFile updates or adds a tracked file entry.
func (m *StateManager) UpdateTrackedFile(state *State, path, localHash, remoteHash, baseHash, status string) {
	state.TrackedFiles[path] = TrackedFile{
		Path:         path,
		LocalHash:    localHash,
		RemoteHash:   remoteHash,
		BaseHash:     baseHash,
		LastModified: time.Now().UTC(),
		Status:       status,
	}
}

// Package sync provides sync state management for Ariadne.
// It handles tracking of synced files, checksums, and conflict detection.
package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/knossos/internal/checksum"
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

// IsInitialized checks if sync has been initialized.
func (m *StateManager) IsInitialized() bool {
	_, err := os.Stat(m.StatePath())
	return err == nil
}

// ComputeFileHash computes the SHA-256 hash of a file with "sha256:" prefix.
func ComputeFileHash(path string) (string, error) {
	return checksum.File(path)
}

// ComputeContentHash computes the SHA-256 hash of content bytes with "sha256:" prefix.
func ComputeContentHash(content []byte) string {
	return checksum.Bytes(content)
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

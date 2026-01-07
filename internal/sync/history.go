// Package sync provides history/audit log for Ariadne.
package sync

import (
	"encoding/json"
	"os"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// HistoryEntry represents a sync history entry.
type HistoryEntry struct {
	Timestamp string                 `json:"timestamp"`
	Operation string                 `json:"operation"` // pull, push, resolve, reset
	Remote    string                 `json:"remote,omitempty"`
	Files     []string               `json:"files,omitempty"`
	FileCount int                    `json:"file_count"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// History represents the sync history file.
type History struct {
	SchemaVersion string         `json:"schema_version"`
	Entries       []HistoryEntry `json:"entries"`
}

// HistoryManager handles sync history operations.
type HistoryManager struct {
	resolver *paths.Resolver
	state    *StateManager
}

// NewHistoryManager creates a new history manager.
func NewHistoryManager(resolver *paths.Resolver) *HistoryManager {
	return &HistoryManager{
		resolver: resolver,
		state:    NewStateManager(resolver),
	}
}

// Load reads the sync history from disk.
func (m *HistoryManager) Load() (*History, error) {
	historyPath := m.state.HistoryPath()

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &History{
				SchemaVersion: "1.0",
				Entries:       []HistoryEntry{},
			}, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read history", err)
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, errors.ErrSyncStateCorrupt(historyPath, err.Error())
	}

	return &history, nil
}

// Save writes the sync history to disk.
func (m *HistoryManager) Save(history *History) error {
	if err := m.state.EnsureSyncDir(); err != nil {
		return err
	}

	history.SchemaVersion = "1.0"

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal history", err)
	}

	historyPath := m.state.HistoryPath()
	if err := os.WriteFile(historyPath, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write history", err)
	}

	return nil
}

// Add records a new history entry.
func (m *HistoryManager) Add(entry HistoryEntry) error {
	history, err := m.Load()
	if err != nil {
		return err
	}

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	if entry.Files != nil {
		entry.FileCount = len(entry.Files)
	}

	history.Entries = append(history.Entries, entry)

	return m.Save(history)
}

// RecordPull records a pull operation.
func (m *HistoryManager) RecordPull(remote string, files []string, success bool, errMsg string) error {
	return m.Add(HistoryEntry{
		Operation: "pull",
		Remote:    remote,
		Files:     files,
		FileCount: len(files),
		Success:   success,
		Error:     errMsg,
	})
}

// RecordPush records a push operation.
func (m *HistoryManager) RecordPush(remote string, files []string, success bool, errMsg string) error {
	return m.Add(HistoryEntry{
		Operation: "push",
		Remote:    remote,
		Files:     files,
		FileCount: len(files),
		Success:   success,
		Error:     errMsg,
	})
}

// RecordResolve records a resolve operation.
func (m *HistoryManager) RecordResolve(path, strategy string, success bool) error {
	return m.Add(HistoryEntry{
		Operation: "resolve",
		Files:     []string{path},
		FileCount: 1,
		Success:   success,
		Metadata: map[string]interface{}{
			"strategy": strategy,
		},
	})
}

// RecordReset records a reset operation.
func (m *HistoryManager) RecordReset(hard bool, filesReset []string) error {
	return m.Add(HistoryEntry{
		Operation: "reset",
		Files:     filesReset,
		FileCount: len(filesReset),
		Success:   true,
		Metadata: map[string]interface{}{
			"hard": hard,
		},
	})
}

// ListOptions configures history listing.
type ListOptions struct {
	Limit     int
	Operation string // Filter by operation
	Since     string // Filter by timestamp (RFC3339)
}

// List returns history entries with optional filtering.
func (m *HistoryManager) List(opts ListOptions) ([]HistoryEntry, error) {
	history, err := m.Load()
	if err != nil {
		return nil, err
	}

	entries := history.Entries

	// Apply filters
	if opts.Operation != "" {
		var filtered []HistoryEntry
		for _, e := range entries {
			if e.Operation == opts.Operation {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	if opts.Since != "" {
		sinceTime, err := time.Parse(time.RFC3339, opts.Since)
		if err == nil {
			var filtered []HistoryEntry
			for _, e := range entries {
				entryTime, err := time.Parse(time.RFC3339, e.Timestamp)
				if err == nil && entryTime.After(sinceTime) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
	}

	// Apply limit (most recent first)
	if opts.Limit > 0 && len(entries) > opts.Limit {
		// Return most recent entries
		start := len(entries) - opts.Limit
		entries = entries[start:]
	}

	// Reverse to show most recent first
	reversed := make([]HistoryEntry, len(entries))
	for i, e := range entries {
		reversed[len(entries)-1-i] = e
	}

	return reversed, nil
}

// Clear removes all history entries.
func (m *HistoryManager) Clear() error {
	history := &History{
		SchemaVersion: "1.0",
		Entries:       []HistoryEntry{},
	}
	return m.Save(history)
}

// Count returns the total number of history entries.
func (m *HistoryManager) Count() (int, error) {
	history, err := m.Load()
	if err != nil {
		return 0, err
	}
	return len(history.Entries), nil
}

package worktree

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
)

const (
	// MetadataFileName is the name of the worktree registry file.
	MetadataFileName = "metadata.json"

	// WorktreeMetaFileName is the name of per-worktree metadata file.
	WorktreeMetaFileName = ".worktree-meta.json"
)

// MetadataManager handles worktree registry persistence.
type MetadataManager struct {
	worktreesDir string
	mu           sync.RWMutex
}

// NewMetadataManager creates a new MetadataManager.
func NewMetadataManager(worktreesDir string) *MetadataManager {
	return &MetadataManager{
		worktreesDir: worktreesDir,
	}
}

// metadataPath returns the path to metadata.json.
func (m *MetadataManager) metadataPath() string {
	return filepath.Join(m.worktreesDir, MetadataFileName)
}

// Load loads the worktree metadata from disk.
func (m *MetadataManager) Load() (*WorktreeMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.metadataPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata if file doesn't exist
			return &WorktreeMetadata{
				Worktrees: []Worktree{},
				UpdatedAt: time.Now().UTC(),
			}, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read metadata", err)
	}

	var meta WorktreeMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse metadata", err)
	}

	return &meta, nil
}

// Save persists the worktree metadata to disk.
func (m *MetadataManager) Save(meta *WorktreeMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(m.worktreesDir, 0755); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create worktrees directory", err)
	}

	// Create .gitignore if not exists
	gitignorePath := filepath.Join(m.worktreesDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte("*\n"), 0644); err != nil {
			// Non-fatal, but log
		}
	}

	meta.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal metadata", err)
	}

	path := m.metadataPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write metadata", err)
	}

	return nil
}

// Add adds a worktree to the registry.
func (m *MetadataManager) Add(wt Worktree) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	// Check for duplicate
	for _, existing := range meta.Worktrees {
		if existing.ID == wt.ID {
			return errors.New(errors.CodeGeneralError, "worktree already exists: "+wt.ID)
		}
	}

	meta.Worktrees = append(meta.Worktrees, wt)
	return m.Save(meta)
}

// Remove removes a worktree from the registry.
func (m *MetadataManager) Remove(id string) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	found := false
	newWorktrees := make([]Worktree, 0, len(meta.Worktrees))
	for _, wt := range meta.Worktrees {
		if wt.ID != id {
			newWorktrees = append(newWorktrees, wt)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New(errors.CodeFileNotFound, "worktree not found: "+id)
	}

	meta.Worktrees = newWorktrees
	return m.Save(meta)
}

// Get retrieves a worktree by ID.
func (m *MetadataManager) Get(id string) (*Worktree, error) {
	meta, err := m.Load()
	if err != nil {
		return nil, err
	}

	for _, wt := range meta.Worktrees {
		if wt.ID == id {
			return &wt, nil
		}
	}

	return nil, errors.New(errors.CodeFileNotFound, "worktree not found: "+id)
}

// GetByName retrieves a worktree by name.
func (m *MetadataManager) GetByName(name string) (*Worktree, error) {
	meta, err := m.Load()
	if err != nil {
		return nil, err
	}

	for _, wt := range meta.Worktrees {
		if wt.Name == name {
			return &wt, nil
		}
	}

	return nil, errors.New(errors.CodeFileNotFound, "worktree not found with name: "+name)
}

// List returns all worktrees.
func (m *MetadataManager) List() ([]Worktree, error) {
	meta, err := m.Load()
	if err != nil {
		return nil, err
	}
	return meta.Worktrees, nil
}

// Update updates a worktree in the registry.
func (m *MetadataManager) Update(wt Worktree) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	found := false
	for i, existing := range meta.Worktrees {
		if existing.ID == wt.ID {
			meta.Worktrees[i] = wt
			found = true
			break
		}
	}

	if !found {
		return errors.New(errors.CodeFileNotFound, "worktree not found: "+wt.ID)
	}

	return m.Save(meta)
}

// GetOlderThan returns worktrees older than the specified duration.
func (m *MetadataManager) GetOlderThan(d time.Duration) ([]Worktree, error) {
	meta, err := m.Load()
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-d)
	var old []Worktree
	for _, wt := range meta.Worktrees {
		if wt.CreatedAt.Before(cutoff) {
			old = append(old, wt)
		}
	}

	return old, nil
}

// PerWorktreeMeta represents metadata stored in each worktree.
type PerWorktreeMeta struct {
	WorktreeID    string `json:"worktree_id"`
	CreatedAt     string `json:"created_at"`
	Name          string `json:"name"`
	FromRef       string `json:"from_ref"`
	Rite          string `json:"rite"`
	Complexity    string `json:"complexity"`
	ParentProject string `json:"parent_project"`
}

// SavePerWorktreeMeta saves metadata to the worktree's .claude directory.
func SavePerWorktreeMeta(worktreePath string, wt Worktree, parentProject string) error {
	claudeDir := filepath.Join(worktreePath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create .claude directory", err)
	}

	meta := PerWorktreeMeta{
		WorktreeID:    wt.ID,
		CreatedAt:     wt.CreatedAt.UTC().Format(time.RFC3339),
		Name:          wt.Name,
		FromRef:       wt.FromRef,
		Rite:          wt.Rite,
		Complexity:    wt.Complexity,
		ParentProject: parentProject,
	}

	data, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal per-worktree metadata", err)
	}

	path := filepath.Join(claudeDir, WorktreeMetaFileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write per-worktree metadata", err)
	}

	return nil
}

// LoadPerWorktreeMeta loads metadata from a worktree's .claude directory.
func LoadPerWorktreeMeta(worktreePath string) (*PerWorktreeMeta, error) {
	path := filepath.Join(worktreePath, ".claude", WorktreeMetaFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.CodeFileNotFound, "worktree metadata not found")
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read per-worktree metadata", err)
	}

	var meta PerWorktreeMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse per-worktree metadata", err)
	}

	return &meta, nil
}

// SyncMetadataFromFilesystem scans the worktrees directory and updates metadata.
// This handles cases where metadata.json might be out of sync with actual worktrees.
func (m *MetadataManager) SyncMetadataFromFilesystem() error {
	entries, err := os.ReadDir(m.worktreesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(errors.CodeGeneralError, "failed to read worktrees directory", err)
	}

	meta, err := m.Load()
	if err != nil {
		return err
	}

	// Build map of existing worktrees
	existingByID := make(map[string]bool)
	for _, wt := range meta.Worktrees {
		existingByID[wt.ID] = true
	}

	// Scan for worktrees on disk
	for _, entry := range entries {
		if !entry.IsDir() || !IsValidWorktreeID(entry.Name()) {
			continue
		}

		id := entry.Name()
		if existingByID[id] {
			continue
		}

		// Found worktree not in metadata - try to recover info from per-worktree meta
		wtPath := filepath.Join(m.worktreesDir, id)
		perMeta, err := LoadPerWorktreeMeta(wtPath)
		if err != nil {
			// Can't recover metadata, create minimal entry
			meta.Worktrees = append(meta.Worktrees, Worktree{
				ID:        id,
				Name:      "unknown",
				Path:      wtPath,
				CreatedAt: ParseWorktreeTimestamp(id),
			})
		} else {
			createdAt, _ := time.Parse(time.RFC3339, perMeta.CreatedAt)
			meta.Worktrees = append(meta.Worktrees, Worktree{
				ID:         id,
				Name:       perMeta.Name,
				Path:       wtPath,
				Rite:       perMeta.Rite,
				Complexity: perMeta.Complexity,
				FromRef:    perMeta.FromRef,
				CreatedAt:  createdAt,
			})
		}
	}

	// Remove entries for worktrees that no longer exist on disk
	validWorktrees := make([]Worktree, 0, len(meta.Worktrees))
	for _, wt := range meta.Worktrees {
		if _, err := os.Stat(wt.Path); err == nil {
			validWorktrees = append(validWorktrees, wt)
		}
	}
	meta.Worktrees = validWorktrees

	return m.Save(meta)
}

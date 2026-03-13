package inscription

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
)

// BackupManager handles backup and rollback operations for CLAUDE.md.
type BackupManager struct {
	// BackupDir is the directory where backups are stored.
	// Default: .knossos/backups
	BackupDir string

	// TargetPath is the path to the file being backed up.
	TargetPath string

	// MaxBackups is the maximum number of backups to retain.
	// Default: 5
	MaxBackups int
}

// BackupInfo contains metadata about a backup.
type BackupInfo struct {
	// Path is the full path to the backup file.
	Path string

	// Timestamp is when the backup was created.
	Timestamp time.Time

	// Size is the file size in bytes.
	Size int64

	// Name is the backup filename.
	Name string
}

// NewBackupManager creates a new backup manager.
// HA-FS: TargetPath targets the actual CC channel context file (SCAR-002: never rename .claude/)
func NewBackupManager(projectRoot string) *BackupManager {
	return &BackupManager{
		BackupDir:  filepath.Join(projectRoot, ".knossos", "backups"),
		TargetPath: filepath.Join(projectRoot, ".claude", "CLAUDE.md"),
		MaxBackups: 5,
	}
}

// NewBackupManagerWithTarget creates a backup manager for a specific file.
func NewBackupManagerWithTarget(backupDir, targetPath string) *BackupManager {
	return &BackupManager{
		BackupDir:  backupDir,
		TargetPath: targetPath,
		MaxBackups: 5,
	}
}

// CreateBackup creates a backup of the target file with timestamp naming.
// Returns the path to the backup file.
func (b *BackupManager) CreateBackup() (string, error) {
	// Check if target exists
	_, err := os.Stat(b.TargetPath)
	if os.IsNotExist(err) {
		return "", errors.NewWithDetails(errors.CodeFileNotFound,
			"target file not found for backup",
			map[string]any{"path": b.TargetPath})
	}
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to stat target file", err)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(b.BackupDir, 0755); err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to create backup directory", err)
	}

	// Read target content
	content, err := os.ReadFile(b.TargetPath)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to read target file", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().UTC()
	backupName := b.generateBackupName(timestamp)
	backupPath := filepath.Join(b.BackupDir, backupName)

	// Write backup atomically
	if err := b.atomicWrite(backupPath, content); err != nil {
		return "", err
	}

	// Clean up old backups
	if err := b.CleanupOldBackups(); err != nil {
		// Log warning but don't fail - backup was successful
		// In production, this would be logged
		_ = err
	}

	return backupPath, nil
}

// generateBackupName creates a backup filename with timestamp.
// Format: CLAUDE.md.2026-01-06T10-30-00Z
func (b *BackupManager) generateBackupName(t time.Time) string {
	// Use RFC3339 format but replace colons with hyphens for filesystem compatibility
	formatted := t.Format("2006-01-02T15-04-05Z")
	baseName := filepath.Base(b.TargetPath)
	return baseName + "." + formatted
}

// parseBackupTimestamp extracts the timestamp from a backup filename.
func (b *BackupManager) parseBackupTimestamp(name string) (time.Time, error) {
	baseName := filepath.Base(b.TargetPath)
	if !strings.HasPrefix(name, baseName+".") {
		return time.Time{}, errors.New(errors.CodeParseError, "invalid backup filename")
	}

	timestampStr := strings.TrimPrefix(name, baseName+".")
	// Parse the timestamp format: 2006-01-02T15-04-05Z
	t, err := time.Parse("2006-01-02T15-04-05Z", timestampStr)
	if err != nil {
		return time.Time{}, errors.Wrap(errors.CodeParseError, "failed to parse backup timestamp", err)
	}

	return t, nil
}

// RestoreBackup restores the target file from a backup.
// If timestamp is empty, restores the most recent backup.
func (b *BackupManager) RestoreBackup(timestamp string) error {
	var backupPath string

	if timestamp == "" {
		// Find most recent backup
		backups, err := b.ListBackups()
		if err != nil {
			return err
		}
		if len(backups) == 0 {
			return errors.New(errors.CodeFileNotFound, "no backups found")
		}
		backupPath = backups[0].Path // Already sorted newest first
	} else {
		// Find backup with specific timestamp
		backups, err := b.ListBackups()
		if err != nil {
			return err
		}

		for _, backup := range backups {
			if strings.Contains(backup.Name, timestamp) {
				backupPath = backup.Path
				break
			}
		}

		if backupPath == "" {
			return errors.NewWithDetails(errors.CodeFileNotFound,
				"backup not found",
				map[string]any{"timestamp": timestamp})
		}
	}

	// Read backup content
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to read backup file", err)
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(b.TargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create target directory", err)
	}

	// Write to target atomically
	return b.atomicWrite(b.TargetPath, content)
}

// CleanupOldBackups removes backups exceeding the retention limit.
// Keeps the most recent MaxBackups backups.
func (b *BackupManager) CleanupOldBackups() error {
	backups, err := b.ListBackups()
	if err != nil {
		return err
	}

	if len(backups) <= b.MaxBackups {
		return nil
	}

	// Remove oldest backups (list is sorted newest first)
	for i := b.MaxBackups; i < len(backups); i++ {
		if err := os.Remove(backups[i].Path); err != nil {
			// Log but continue - try to remove other old backups
			_ = err
		}
	}

	return nil
}

// ListBackups returns all backups sorted by timestamp (newest first).
func (b *BackupManager) ListBackups() ([]BackupInfo, error) {
	// Check if backup directory exists
	if _, err := os.Stat(b.BackupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	entries, err := os.ReadDir(b.BackupDir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read backup directory", err)
	}

	baseName := filepath.Base(b.TargetPath)
	var backups []BackupInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, baseName+".") {
			continue
		}

		timestamp, err := b.parseBackupTimestamp(name)
		if err != nil {
			continue // Skip invalid backup files
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      filepath.Join(b.BackupDir, name),
			Timestamp: timestamp,
			Size:      info.Size(),
			Name:      name,
		})
	}

	// Sort by timestamp, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// atomicWrite writes content to a file atomically via a temp file.
func (b *BackupManager) atomicWrite(path string, content []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create directory", err)
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write temp file", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath) // Clean up on failure
		return errors.Wrap(errors.CodeGeneralError, "failed to rename temp file", err)
	}

	return nil
}

// HasBackups returns true if any backups exist.
func (b *BackupManager) HasBackups() bool {
	backups, err := b.ListBackups()
	if err != nil {
		return false
	}
	return len(backups) > 0
}

// GetLatestBackup returns the most recent backup info, or nil if none exist.
func (b *BackupManager) GetLatestBackup() *BackupInfo {
	backups, err := b.ListBackups()
	if err != nil || len(backups) == 0 {
		return nil
	}
	return &backups[0]
}

// GetBackupByTimestamp returns backup info for a specific timestamp.
func (b *BackupManager) GetBackupByTimestamp(timestamp string) *BackupInfo {
	backups, err := b.ListBackups()
	if err != nil {
		return nil
	}

	for _, backup := range backups {
		if strings.Contains(backup.Name, timestamp) {
			return &backup
		}
	}

	return nil
}

// DeleteBackup removes a specific backup.
func (b *BackupManager) DeleteBackup(timestamp string) error {
	backup := b.GetBackupByTimestamp(timestamp)
	if backup == nil {
		return errors.NewWithDetails(errors.CodeFileNotFound,
			"backup not found",
			map[string]any{"timestamp": timestamp})
	}

	if err := os.Remove(backup.Path); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to delete backup", err)
	}

	return nil
}

// DeleteAllBackups removes all backups.
func (b *BackupManager) DeleteAllBackups() error {
	backups, err := b.ListBackups()
	if err != nil {
		return err
	}

	for _, backup := range backups {
		if err := os.Remove(backup.Path); err != nil {
			// Continue trying to delete other backups
			_ = err
		}
	}

	return nil
}

// BackupAndWrite creates a backup then writes new content atomically.
// Returns the backup path if successful.
func (b *BackupManager) BackupAndWrite(newContent []byte) (string, error) {
	// Check if target exists for backup
	if _, err := os.Stat(b.TargetPath); err == nil {
		// Create backup first
		backupPath, err := b.CreateBackup()
		if err != nil {
			return "", err
		}

		// Write new content
		if err := b.atomicWrite(b.TargetPath, newContent); err != nil {
			return backupPath, err
		}

		return backupPath, nil
	}

	// No existing file to backup, just write
	if err := b.atomicWrite(b.TargetPath, newContent); err != nil {
		return "", err
	}

	return "", nil
}

// VerifyBackup checks that a backup is valid and matches expected hash.
func (b *BackupManager) VerifyBackup(timestamp, expectedHash string) error {
	backup := b.GetBackupByTimestamp(timestamp)
	if backup == nil {
		return errors.NewWithDetails(errors.CodeFileNotFound,
			"backup not found",
			map[string]any{"timestamp": timestamp})
	}

	content, err := os.ReadFile(backup.Path)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to read backup", err)
	}

	actualHash := ComputeContentHash(string(content))
	if actualHash != expectedHash {
		return errors.NewWithDetails(errors.CodeGeneralError,
			"backup hash mismatch",
			map[string]any{
				"expected": expectedHash,
				"actual":   actualHash,
			})
	}

	return nil
}

// GetBackupContent reads and returns the content of a backup.
func (b *BackupManager) GetBackupContent(timestamp string) (string, error) {
	backup := b.GetBackupByTimestamp(timestamp)
	if backup == nil {
		return "", errors.NewWithDetails(errors.CodeFileNotFound,
			"backup not found",
			map[string]any{"timestamp": timestamp})
	}

	content, err := os.ReadFile(backup.Path)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to read backup", err)
	}

	return string(content), nil
}

// GetLatestBackupContent reads and returns the content of the most recent backup.
func (b *BackupManager) GetLatestBackupContent() (string, error) {
	backup := b.GetLatestBackup()
	if backup == nil {
		return "", errors.New(errors.CodeFileNotFound, "no backups found")
	}

	content, err := os.ReadFile(backup.Path)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to read backup", err)
	}

	return string(content), nil
}

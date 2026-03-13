package inscription

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/paths"
)

func TestNewBackupManager(t *testing.T) {
	bm := NewBackupManager("/project")

	if bm.BackupDir != "/project/.knossos/backups" {
		t.Errorf("NewBackupManager() BackupDir = %q", bm.BackupDir)
	}
	wantTarget := "/project/" + paths.ClaudeChannel{}.DirName() + "/" + paths.ClaudeChannel{}.ContextFile()
	if bm.TargetPath != wantTarget {
		t.Errorf("NewBackupManager() TargetPath = %q, want %q", bm.TargetPath, wantTarget)
	}
	if bm.MaxBackups != 5 {
		t.Errorf("NewBackupManager() MaxBackups = %d", bm.MaxBackups)
	}
}

func TestNewBackupManagerWithTarget(t *testing.T) {
	bm := NewBackupManagerWithTarget("/backups", "/target/file.md")

	if bm.BackupDir != "/backups" {
		t.Errorf("NewBackupManagerWithTarget() BackupDir = %q", bm.BackupDir)
	}
	if bm.TargetPath != "/target/file.md" {
		t.Errorf("NewBackupManagerWithTarget() TargetPath = %q", bm.TargetPath)
	}
}

func TestBackupManager_CreateBackup(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	os.MkdirAll(targetDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")
	content := "# Test Content\n\nThis is test content."
	os.WriteFile(targetPath, []byte(content), 0644)

	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, ".claude", "backups"),
		targetPath,
	)

	backupPath, err := bm.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if backupPath == "" {
		t.Error("CreateBackup() returned empty path")
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("CreateBackup() backup file does not exist")
	}

	// Verify backup content
	backupContent, _ := os.ReadFile(backupPath)
	if string(backupContent) != content {
		t.Errorf("CreateBackup() content mismatch, got %q", string(backupContent))
	}

	// Verify filename format
	if !strings.Contains(filepath.Base(backupPath), "CLAUDE.md.") {
		t.Errorf("CreateBackup() filename format incorrect: %s", filepath.Base(backupPath))
	}
}

func TestBackupManager_CreateBackup_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "nonexistent.md"),
	)

	_, err := bm.CreateBackup()
	if err == nil {
		t.Error("CreateBackup() expected error for nonexistent file")
	}
}

func TestBackupManager_RestoreBackup_Latest(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	backupDir := filepath.Join(targetDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")

	// Create original file
	os.WriteFile(targetPath, []byte("Original content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)

	// Create backup
	backupPath, _ := bm.CreateBackup()

	// Modify target
	os.WriteFile(targetPath, []byte("Modified content"), 0644)

	// Restore latest backup
	err := bm.RestoreBackup("")
	if err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}

	// Verify restored content
	content, _ := os.ReadFile(targetPath)
	if string(content) != "Original content" {
		t.Errorf("RestoreBackup() content = %q, want 'Original content'", string(content))
	}

	_ = backupPath // silence unused variable
}

func TestBackupManager_RestoreBackup_ByTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	backupDir := filepath.Join(targetDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")

	// Create first backup
	os.WriteFile(targetPath, []byte("First content"), 0644)
	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	bm.MaxBackups = 10 // Allow more backups for this test
	bm.CreateBackup()

	// Wait a full second and create second backup (timestamp has 1-second resolution)
	time.Sleep(1100 * time.Millisecond)
	os.WriteFile(targetPath, []byte("Second content"), 0644)
	bm.CreateBackup()

	// Get list of backups
	backups, _ := bm.ListBackups()
	if len(backups) < 2 {
		t.Skip("Test requires timestamp resolution to create distinct backups")
	}

	// Restore older backup (second in list since sorted newest first)
	olderBackup := backups[1]
	timestamp := strings.TrimPrefix(olderBackup.Name, "CLAUDE.md.")

	err := bm.RestoreBackup(timestamp)
	if err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}

	content, _ := os.ReadFile(targetPath)
	if string(content) != "First content" {
		t.Errorf("RestoreBackup() content = %q, want 'First content'", string(content))
	}
}

func TestBackupManager_RestoreBackup_NoBackups(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "target.md"),
	)

	err := bm.RestoreBackup("")
	if err == nil {
		t.Error("RestoreBackup() expected error when no backups exist")
	}
}

func TestBackupManager_RestoreBackup_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	backupDir := filepath.Join(targetDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	bm.CreateBackup()

	err := bm.RestoreBackup("nonexistent-timestamp")
	if err == nil {
		t.Error("RestoreBackup() expected error for nonexistent timestamp")
	}
}

func TestBackupManager_CleanupOldBackups(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	backupDir := filepath.Join(targetDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	bm.MaxBackups = 2

	// Create more backups than limit
	for range 5 {
		bm.CreateBackup()
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Cleanup should have already run in CreateBackup
	backups, _ := bm.ListBackups()
	if len(backups) > 2 {
		t.Errorf("CleanupOldBackups() left %d backups, want <= 2", len(backups))
	}
}

func TestBackupManager_ListBackups(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	backupDir := filepath.Join(targetDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	bm.MaxBackups = 10 // Allow more for this test

	// Create multiple backups with 1-second delays (timestamp resolution is 1 second)
	for i := range 3 {
		bm.CreateBackup()
		if i < 2 {
			time.Sleep(1100 * time.Millisecond)
		}
	}

	backups, err := bm.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}

	// Due to timestamp collision, we may get fewer backups
	if len(backups) == 0 {
		t.Error("ListBackups() got 0 backups, want at least 1")
	}

	// Verify sorted newest first
	for i := 1; i < len(backups); i++ {
		if !backups[i-1].Timestamp.After(backups[i].Timestamp) {
			t.Error("ListBackups() not sorted newest first")
		}
	}
}

func TestBackupManager_ListBackups_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "target.md"),
	)

	backups, err := bm.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("ListBackups() empty dir got %d backups, want 0", len(backups))
	}
}

func TestBackupManager_ListBackups_IgnoresOtherFiles(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	bm.CreateBackup()

	// Create some other files in backup dir
	os.WriteFile(filepath.Join(backupDir, "OTHER.md.2026-01-06T10-00-00Z"), []byte("other"), 0644)
	os.WriteFile(filepath.Join(backupDir, "random.txt"), []byte("random"), 0644)

	backups, _ := bm.ListBackups()
	if len(backups) != 1 {
		t.Errorf("ListBackups() should only list target backups, got %d", len(backups))
	}
}

func TestBackupManager_HasBackups(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)

	if bm.HasBackups() {
		t.Error("HasBackups() should be false initially")
	}

	bm.CreateBackup()

	if !bm.HasBackups() {
		t.Error("HasBackups() should be true after creating backup")
	}
}

func TestBackupManager_GetLatestBackup(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)

	// No backups initially
	if bm.GetLatestBackup() != nil {
		t.Error("GetLatestBackup() should be nil initially")
	}

	bm.MaxBackups = 10
	bm.CreateBackup()
	time.Sleep(10 * time.Millisecond)

	os.WriteFile(targetPath, []byte("Updated"), 0644)
	bm.CreateBackup()

	latest := bm.GetLatestBackup()
	if latest == nil {
		t.Fatal("GetLatestBackup() should not be nil after creating backups")
	}

	// Verify it's the most recent
	backups, _ := bm.ListBackups()
	if latest.Path != backups[0].Path {
		t.Error("GetLatestBackup() should return most recent backup")
	}
}

func TestBackupManager_GetBackupByTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	backupPath, _ := bm.CreateBackup()

	// Extract timestamp from path
	name := filepath.Base(backupPath)
	timestamp := strings.TrimPrefix(name, "CLAUDE.md.")

	backup := bm.GetBackupByTimestamp(timestamp)
	if backup == nil {
		t.Fatal("GetBackupByTimestamp() should find backup")
	}
	if backup.Path != backupPath {
		t.Errorf("GetBackupByTimestamp() path = %q, want %q", backup.Path, backupPath)
	}
}

func TestBackupManager_GetBackupByTimestamp_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "target.md"),
	)

	if bm.GetBackupByTimestamp("nonexistent") != nil {
		t.Error("GetBackupByTimestamp() should return nil for nonexistent")
	}
}

func TestBackupManager_DeleteBackup(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	backupPath, _ := bm.CreateBackup()

	name := filepath.Base(backupPath)
	timestamp := strings.TrimPrefix(name, "CLAUDE.md.")

	err := bm.DeleteBackup(timestamp)
	if err != nil {
		t.Fatalf("DeleteBackup() error = %v", err)
	}

	// Verify deleted
	if bm.HasBackups() {
		t.Error("DeleteBackup() backup still exists")
	}
}

func TestBackupManager_DeleteBackup_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "target.md"),
	)

	err := bm.DeleteBackup("nonexistent")
	if err == nil {
		t.Error("DeleteBackup() expected error for nonexistent")
	}
}

func TestBackupManager_DeleteAllBackups(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	bm.MaxBackups = 10

	// Create multiple backups
	for range 3 {
		bm.CreateBackup()
		time.Sleep(10 * time.Millisecond)
	}

	if !bm.HasBackups() {
		t.Fatal("Test setup failed - no backups created")
	}

	err := bm.DeleteAllBackups()
	if err != nil {
		t.Fatalf("DeleteAllBackups() error = %v", err)
	}

	if bm.HasBackups() {
		t.Error("DeleteAllBackups() backups still exist")
	}
}

func TestBackupManager_BackupAndWrite(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	backupDir := filepath.Join(targetDir, "backups")
	os.MkdirAll(targetDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("Original content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)

	backupPath, err := bm.BackupAndWrite([]byte("New content"))
	if err != nil {
		t.Fatalf("BackupAndWrite() error = %v", err)
	}

	// Backup should exist
	if backupPath == "" {
		t.Error("BackupAndWrite() should return backup path")
	}
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("BackupAndWrite() backup file should exist")
	}

	// Target should have new content
	content, _ := os.ReadFile(targetPath)
	if string(content) != "New content" {
		t.Errorf("BackupAndWrite() target content = %q", string(content))
	}

	// Backup should have old content
	backupContent, _ := os.ReadFile(backupPath)
	if string(backupContent) != "Original content" {
		t.Errorf("BackupAndWrite() backup content = %q", string(backupContent))
	}
}

func TestBackupManager_BackupAndWrite_NoExisting(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName())
	os.MkdirAll(targetDir, 0755)

	targetPath := filepath.Join(targetDir, "CLAUDE.md")
	backupDir := filepath.Join(targetDir, "backups")

	bm := NewBackupManagerWithTarget(backupDir, targetPath)

	backupPath, err := bm.BackupAndWrite([]byte("New content"))
	if err != nil {
		t.Fatalf("BackupAndWrite() error = %v", err)
	}

	// No backup when file didn't exist
	if backupPath != "" {
		t.Error("BackupAndWrite() should not create backup for new file")
	}

	// Target should exist with new content
	content, _ := os.ReadFile(targetPath)
	if string(content) != "New content" {
		t.Errorf("BackupAndWrite() target content = %q", string(content))
	}
}

func TestBackupManager_VerifyBackup(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	content := "Test content for hash"
	os.WriteFile(targetPath, []byte(content), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	backupPath, _ := bm.CreateBackup()

	name := filepath.Base(backupPath)
	timestamp := strings.TrimPrefix(name, "CLAUDE.md.")
	expectedHash := ComputeContentHash(content)

	// Valid hash should succeed
	err := bm.VerifyBackup(timestamp, expectedHash)
	if err != nil {
		t.Errorf("VerifyBackup() valid hash error = %v", err)
	}

	// Invalid hash should fail
	err = bm.VerifyBackup(timestamp, "invalid-hash")
	if err == nil {
		t.Error("VerifyBackup() should fail for invalid hash")
	}
}

func TestBackupManager_VerifyBackup_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "target.md"),
	)

	err := bm.VerifyBackup("nonexistent", "hash")
	if err == nil {
		t.Error("VerifyBackup() should fail for nonexistent backup")
	}
}

func TestBackupManager_GetBackupContent(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	content := "Backup content here"
	os.WriteFile(targetPath, []byte(content), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	backupPath, _ := bm.CreateBackup()

	name := filepath.Base(backupPath)
	timestamp := strings.TrimPrefix(name, "CLAUDE.md.")

	got, err := bm.GetBackupContent(timestamp)
	if err != nil {
		t.Fatalf("GetBackupContent() error = %v", err)
	}
	if got != content {
		t.Errorf("GetBackupContent() = %q, want %q", got, content)
	}
}

func TestBackupManager_GetBackupContent_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "target.md"),
	)

	_, err := bm.GetBackupContent("nonexistent")
	if err == nil {
		t.Error("GetBackupContent() should fail for nonexistent")
	}
}

func TestBackupManager_GetLatestBackupContent(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	targetPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(targetPath, []byte("First content"), 0644)

	bm := NewBackupManagerWithTarget(backupDir, targetPath)
	bm.MaxBackups = 10
	bm.CreateBackup()

	time.Sleep(10 * time.Millisecond)
	os.WriteFile(targetPath, []byte("Second content"), 0644)
	bm.CreateBackup()

	content, err := bm.GetLatestBackupContent()
	if err != nil {
		t.Fatalf("GetLatestBackupContent() error = %v", err)
	}
	if content != "Second content" {
		t.Errorf("GetLatestBackupContent() = %q, want 'Second content'", content)
	}
}

func TestBackupManager_GetLatestBackupContent_NoBackups(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManagerWithTarget(
		filepath.Join(tmpDir, "backups"),
		filepath.Join(tmpDir, "target.md"),
	)

	_, err := bm.GetLatestBackupContent()
	if err == nil {
		t.Error("GetLatestBackupContent() should fail when no backups")
	}
}

func TestBackupInfo_Fields(t *testing.T) {
	now := time.Now()
	info := BackupInfo{
		Path:      "/path/to/backup",
		Timestamp: now,
		Size:      1024,
		Name:      "CLAUDE.md.2026-01-06T10-00-00Z",
	}

	if info.Path != "/path/to/backup" {
		t.Error("BackupInfo Path not set")
	}
	if !info.Timestamp.Equal(now) {
		t.Error("BackupInfo Timestamp not set")
	}
	if info.Size != 1024 {
		t.Error("BackupInfo Size not set")
	}
	if info.Name != "CLAUDE.md.2026-01-06T10-00-00Z" {
		t.Error("BackupInfo Name not set")
	}
}

func TestBackupManager_GenerateBackupName(t *testing.T) {
	bm := NewBackupManagerWithTarget("/backups", "/target/CLAUDE.md")

	timestamp := time.Date(2026, 1, 6, 10, 30, 45, 0, time.UTC)
	name := bm.generateBackupName(timestamp)

	if name != "CLAUDE.md.2026-01-06T10-30-45Z" {
		t.Errorf("generateBackupName() = %q", name)
	}
}

func TestBackupManager_ParseBackupTimestamp(t *testing.T) {
	bm := NewBackupManagerWithTarget("/backups", "/target/CLAUDE.md")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "CLAUDE.md.2026-01-06T10-30-45Z", false},
		{"invalid prefix", "OTHER.md.2026-01-06T10-30-45Z", true},
		{"invalid timestamp", "CLAUDE.md.invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bm.parseBackupTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBackupTimestamp(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

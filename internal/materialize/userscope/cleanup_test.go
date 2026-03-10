package userscope

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// manifestNames returns the five legacy JSON manifest base names.
func manifestNames() []string {
	return []string{
		"USER_AGENT_MANIFEST.json",
		"USER_MENA_MANIFEST.json",
		"USER_HOOKS_MANIFEST.json",
		"USER_COMMAND_MANIFEST.json",
		"USER_SKILL_MANIFEST.json",
	}
}

func TestCleanupOldManifests_RemovesBackups(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create only .v2-backup files (simulates a satellite that ran previous migration)
	for _, name := range manifestNames() {
		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		require.NoError(t, os.WriteFile(backupPath, []byte(`{"version":1}`), 0644))
	}

	cleanupOldManifests(tmpDir)

	for _, name := range manifestNames() {
		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		_, err := os.Stat(backupPath)
		assert.True(t, os.IsNotExist(err), "v2-backup file should be removed: %s", name)
	}
}

func TestCleanupOldManifests_RemovesOriginals(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create only the original JSON manifests (no backups)
	for _, name := range manifestNames() {
		origPath := filepath.Join(tmpDir, name)
		require.NoError(t, os.WriteFile(origPath, []byte(`{"agents":[]}`), 0644))
	}

	cleanupOldManifests(tmpDir)

	for _, name := range manifestNames() {
		origPath := filepath.Join(tmpDir, name)
		_, err := os.Stat(origPath)
		assert.True(t, os.IsNotExist(err), "original manifest should be removed: %s", name)

		// No backup should have been created
		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		_, err = os.Stat(backupPath)
		assert.True(t, os.IsNotExist(err), "no v2-backup should be created: %s", name)
	}
}

func TestCleanupOldManifests_BothPresent(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create both originals and .v2-backup files
	for _, name := range manifestNames() {
		origPath := filepath.Join(tmpDir, name)
		require.NoError(t, os.WriteFile(origPath, []byte(`{"agents":[]}`), 0644))
		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		require.NoError(t, os.WriteFile(backupPath, []byte(`{"agents":[]}`), 0644))
	}

	cleanupOldManifests(tmpDir)

	for _, name := range manifestNames() {
		origPath := filepath.Join(tmpDir, name)
		_, err := os.Stat(origPath)
		assert.True(t, os.IsNotExist(err), "original should be removed: %s", name)

		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		_, err = os.Stat(backupPath)
		assert.True(t, os.IsNotExist(err), "v2-backup should be removed: %s", name)
	}
}

func TestCleanupOldManifests_NonePresent(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Should complete without errors when no files exist
	assert.NotPanics(t, func() {
		cleanupOldManifests(tmpDir)
	})
}

func TestCleanupOldManifests_PartialPresence(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	names := manifestNames()

	// Create only the first 2 backup files
	for _, name := range names[:2] {
		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		require.NoError(t, os.WriteFile(backupPath, []byte(`{}`), 0644))
	}

	cleanupOldManifests(tmpDir)

	// The 2 that existed should be gone
	for _, name := range names[:2] {
		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		_, err := os.Stat(backupPath)
		assert.True(t, os.IsNotExist(err), "v2-backup should be removed: %s", name)
	}

	// The other 3 were never there — should not exist (no panic)
	for _, name := range names[2:] {
		backupPath := filepath.Join(tmpDir, name+".v2-backup")
		_, err := os.Stat(backupPath)
		assert.True(t, os.IsNotExist(err), "should not exist: %s", name)
	}
}

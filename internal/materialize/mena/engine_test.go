package mena

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCleanEmptyDirs_NonExistentRoot verifies that CleanEmptyDirs returns nil
// errors when called with a path that does not exist on disk.
func TestCleanEmptyDirs_NonExistentRoot(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	errs := CleanEmptyDirs(nonExistent)
	if len(errs) != 0 {
		t.Errorf("CleanEmptyDirs(%q) returned %d errors, want 0: %v", nonExistent, len(errs), errs)
	}
}

// TestCleanEmptyDirs_ExistingEmptySubdir verifies normal behavior:
// empty subdirectories are removed successfully.
func TestCleanEmptyDirs_ExistingEmptySubdir(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, "root")
	subdir := filepath.Join(root, "empty-child")

	if err := mkdirAll(subdir); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	errs := CleanEmptyDirs(root)
	if len(errs) != 0 {
		t.Errorf("CleanEmptyDirs returned errors: %v", errs)
	}

	// The empty subdir should have been removed
	if exists(subdir) {
		t.Errorf("expected empty subdirectory %q to be removed", subdir)
	}
}

// helpers

func mkdirAll(path string) error {
	return mkdirAllMode(path, 0755)
}

func mkdirAllMode(path string, mode uint32) error {
	return os.MkdirAll(path, os.FileMode(mode))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

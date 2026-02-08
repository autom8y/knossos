// Package fileutil provides canonical file write utilities for the Knossos platform.
package fileutil

import (
	"bytes"
	"os"
	"path/filepath"
)

// AtomicWriteFile writes content to path atomically using temp-file-then-rename.
// Uses os.CreateTemp for safe temp file names, calls Sync() before rename,
// and creates parent directories as needed. The file permission is set to perm.
func AtomicWriteFile(path string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	// Ensure parent directories exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create temp file in the same directory
	tmpFile, err := os.CreateTemp(dir, base+".tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write data
	if _, err := tmpFile.Write(content); err != nil {
		return err
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		return err
	}

	// Close before rename
	if err := tmpFile.Close(); err != nil {
		return err
	}
	tmpFile = nil // Prevent defer cleanup

	// Set permissions
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath) // best-effort cleanup
		return err
	}

	// Atomic rename
	return os.Rename(tmpPath, path)
}

// WriteIfChanged writes content to path only if it differs from the existing file.
// Returns true if a write occurred, false if content was identical.
// Uses AtomicWriteFile for safe writes.
func WriteIfChanged(path string, content []byte, perm os.FileMode) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, content) {
		return false, nil
	}
	return true, AtomicWriteFile(path, content, perm)
}

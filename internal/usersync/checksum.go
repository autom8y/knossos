package usersync

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ChecksumPrefix is the prefix used for SHA256 checksums.
const ChecksumPrefix = "sha256:"

// ComputeFileChecksum calculates SHA256 checksum of a file.
// Returns empty string if file doesn't exist or cannot be read.
func ComputeFileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return ChecksumPrefix + hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeContentChecksum calculates SHA256 checksum of content.
func ComputeContentChecksum(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return ChecksumPrefix + hex.EncodeToString(h.Sum(nil))
}

// ComputeDirChecksum calculates a composite checksum for a directory.
// It hashes all file contents in sorted order (by relative path).
// Used for tracking nested directory changes.
func ComputeDirChecksum(dirPath string) (string, error) {
	h := sha256.New()

	// Collect all file paths first
	var files []string
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		files = append(files, relPath)
		return nil
	})
	if err != nil {
		return "", err
	}

	// Sort for deterministic ordering
	sort.Strings(files)

	// Hash each file's path and content
	for _, relPath := range files {
		fullPath := filepath.Join(dirPath, relPath)

		// Include path in hash for structure awareness
		h.Write([]byte(relPath))

		// Include file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", err
		}
		h.Write(content)
	}

	return ChecksumPrefix + hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyChecksum checks if a file's current checksum matches the expected value.
func VerifyChecksum(path, expected string) (bool, error) {
	actual, err := ComputeFileChecksum(path)
	if err != nil {
		return false, err
	}
	return actual == expected, nil
}

package usersync

import (
	"os"

	"github.com/autom8y/knossos/internal/checksum"
)

// ChecksumPrefix is the prefix used for SHA256 checksums.
// Kept for backward compatibility; delegates to checksum.Prefix.
const ChecksumPrefix = checksum.Prefix

// ComputeFileChecksum calculates SHA256 checksum of a file with "sha256:" prefix.
// Returns error if file doesn't exist or cannot be read. This preserves the
// original usersync behavior where non-existent files return an error (unlike
// checksum.File which returns empty string + nil for missing files).
func ComputeFileChecksum(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return checksum.File(path)
}

// ComputeContentChecksum calculates SHA256 checksum of content with "sha256:" prefix.
func ComputeContentChecksum(content []byte) string {
	return checksum.Bytes(content)
}

// ComputeDirChecksum calculates a composite checksum for a directory with "sha256:" prefix.
// It hashes all file contents in sorted order (by relative path).
func ComputeDirChecksum(dirPath string) (string, error) {
	return checksum.Dir(dirPath)
}

// VerifyChecksum checks if a file's current checksum matches the expected value.
func VerifyChecksum(path, expected string) (bool, error) {
	actual, err := ComputeFileChecksum(path)
	if err != nil {
		return false, err
	}
	return actual == expected, nil
}

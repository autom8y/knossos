// Package checksum provides unified SHA256 hashing with a standard "sha256:"
// prefix format per ADR-0026. All hash functions in this package return strings
// prefixed with "sha256:" followed by the lowercase hex-encoded digest.
package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// Prefix is the standard prefix for SHA256 checksums.
const Prefix = "sha256:"

// Content computes the SHA256 hash of a string and returns it with the
// "sha256:" prefix.
func Content(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return Prefix + hex.EncodeToString(h.Sum(nil))
}

// Bytes computes the SHA256 hash of a byte slice and returns it with the
// "sha256:" prefix.
func Bytes(content []byte) string {
	h := sha256.Sum256(content)
	return Prefix + hex.EncodeToString(h[:])
}

// File computes the SHA256 hash of a file and returns it with the "sha256:"
// prefix. Returns empty string and nil error if the file does not exist.
func File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return Prefix + hex.EncodeToString(h.Sum(nil)), nil
}

// Dir computes a composite SHA256 hash for a directory by hashing all file
// paths and contents in sorted order. Returns the result with the "sha256:"
// prefix.
func Dir(dirPath string) (string, error) {
	h := sha256.New()

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

	sort.Strings(files)

	for _, relPath := range files {
		fullPath := filepath.Join(dirPath, relPath)
		h.Write([]byte(relPath))
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", err
		}
		h.Write(content)
	}

	return Prefix + hex.EncodeToString(h.Sum(nil)), nil
}

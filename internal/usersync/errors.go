// Package usersync syncs user-level resources to ~/.claude/.
package usersync

import (
	"fmt"

	"github.com/autom8y/knossos/internal/errors"
)

// Error codes for usersync package.
const (
	CodeKnossosHomeNotSet   = "KNOSSOS_HOME_NOT_SET"
	CodeInvalidResourceType = "INVALID_RESOURCE_TYPE"
	CodeSourceNotFound      = "SOURCE_NOT_FOUND"
	CodeTargetCreateFailed  = "TARGET_CREATE_FAILED"
	CodeManifestReadError   = "MANIFEST_READ_ERROR"
	CodeManifestWriteError  = "MANIFEST_WRITE_ERROR"
	CodeChecksumError       = "CHECKSUM_ERROR"
	CodeCopyError           = "COPY_ERROR"
)

// Exit codes for usersync.
const (
	ExitSuccess        = 0 // Sync completed successfully
	ExitCollisions     = 1 // Sync completed with collisions detected
	ExitUsageError     = 2 // Invalid arguments
	ExitSourceNotFound = 3 // Source directory not found
	ExitTargetError    = 4 // Target directory creation failed
	ExitManifestError  = 5 // Manifest read/write error
	ExitKnossosNotSet  = 6 // KNOSSOS_HOME not configured
)

// Package-level errors.
var (
	ErrKnossosHomeNotSet   = errors.New(CodeKnossosHomeNotSet, "KNOSSOS_HOME environment variable not set")
	ErrInvalidResourceType = errors.New(CodeInvalidResourceType, "invalid resource type")
)

// ErrSourceNotFound returns an error for missing source directory.
func ErrSourceNotFound(path string) *errors.Error {
	return errors.NewWithDetails(CodeSourceNotFound,
		fmt.Sprintf("source directory not found: %s", path),
		map[string]any{"path": path})
}

// ErrTargetCreateFailed returns an error for target directory creation failure.
func ErrTargetCreateFailed(path string, cause error) *errors.Error {
	details := map[string]any{"path": path}
	if cause != nil {
		details["cause"] = cause.Error()
	}
	return errors.NewWithDetails(CodeTargetCreateFailed,
		fmt.Sprintf("failed to create target directory: %s", path),
		details)
}

// ErrManifestRead returns an error for manifest read failure.
func ErrManifestRead(path string, cause error) *errors.Error {
	details := map[string]any{"path": path}
	if cause != nil {
		details["cause"] = cause.Error()
	}
	return errors.NewWithDetails(CodeManifestReadError,
		fmt.Sprintf("failed to read manifest: %s", path),
		details)
}

// ErrManifestWrite returns an error for manifest write failure.
func ErrManifestWrite(path string, cause error) *errors.Error {
	details := map[string]any{"path": path}
	if cause != nil {
		details["cause"] = cause.Error()
	}
	return errors.NewWithDetails(CodeManifestWriteError,
		fmt.Sprintf("failed to write manifest: %s", path),
		details)
}

// ErrChecksum returns an error for checksum computation failure.
func ErrChecksum(path string, cause error) *errors.Error {
	details := map[string]any{"path": path}
	if cause != nil {
		details["cause"] = cause.Error()
	}
	return errors.NewWithDetails(CodeChecksumError,
		fmt.Sprintf("failed to compute checksum: %s", path),
		details)
}

// ErrCopy returns an error for file copy failure.
func ErrCopy(src, dst string, cause error) *errors.Error {
	details := map[string]any{
		"source":      src,
		"destination": dst,
	}
	if cause != nil {
		details["cause"] = cause.Error()
	}
	return errors.NewWithDetails(CodeCopyError,
		fmt.Sprintf("failed to copy file: %s -> %s", src, dst),
		details)
}

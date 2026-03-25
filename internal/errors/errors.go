// Package errors provides domain-specific error types for Ariadne.
// Exit codes follow the TDD specification (Section 4.1).
package errors

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"slices"
)

// Exit codes per TDD Section 4.1 (GAP-5 Resolution)
const (
	ExitSuccess          = 0  // Operation completed successfully
	ExitGeneralError     = 1  // Unspecified error
	ExitUsageError       = 2  // Invalid arguments or flags
	ExitLockTimeout      = 3  // Could not acquire lock within timeout
	ExitSchemaInvalid    = 4  // Data failed schema validation
	ExitLifecycleError   = 5  // Invalid state transition (also: orphan conflict)
	ExitFileNotFound     = 6  // Required file, session, or rite not found
	ExitPermissionDenied = 7  // Cannot read/write file
	ExitMergeConflict    = 8  // Three-way merge has conflicts
	ExitProjectNotFound  = 9  // No channel directory found
	ExitSessionExists    = 10 // Session already active (for create)
	ExitMigrationFailed  = 11 // Schema migration failed
	ExitValidationFailed = 12 // Rite validation checks failed
	ExitSwitchAborted    = 13 // Rite switch rolled back due to error
	ExitSchemaNotFound   = 14 // Specified schema not available
	ExitParseError       = 15 // JSON/YAML parsing failed
	// Sync-domain exit codes
	ExitSyncStateCorrupt  = 16 // state.json is invalid or corrupt
	ExitRemoteRejected    = 17 // Push rejected by remote
	ExitNetworkError      = 18 // Failed to fetch from remote
	ExitSyncNotConfigured = 21 // Sync not initialized/no remote configured
)

// Error codes for JSON output
const (
	CodeGeneralError       = "GENERAL_ERROR"
	CodeUsageError         = "USAGE_ERROR"
	CodeLockTimeout        = "LOCK_TIMEOUT"
	CodeLockStale          = "LOCK_STALE"
	CodeSchemaInvalid      = "SCHEMA_INVALID"
	CodeLifecycleViolation = "LIFECYCLE_VIOLATION"
	CodeFileNotFound       = "FILE_NOT_FOUND"
	CodeSessionNotFound    = "SESSION_NOT_FOUND"
	CodePermissionDenied   = "PERMISSION_DENIED"
	CodeMergeConflict      = "MERGE_CONFLICT"
	CodeProjectNotFound    = "PROJECT_NOT_FOUND"
	CodeSessionExists      = "SESSION_EXISTS"
	CodeMigrationFailed    = "MIGRATION_FAILED"
	// Rite-domain error codes
	CodeOrphanConflict   = "ORPHAN_CONFLICT"
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeSwitchAborted    = "SWITCH_ABORTED"
	// Manifest-domain error codes
	CodeSchemaNotFound = "SCHEMA_NOT_FOUND"
	CodeParseError     = "PARSE_ERROR"
	// Sync-domain error codes
	CodeSyncStateCorrupt  = "SYNC_STATE_CORRUPT"
	CodeRemoteRejected    = "REMOTE_REJECTED"
	CodeNetworkError      = "NETWORK_ERROR"
	CodeRemoteNotFound    = "REMOTE_NOT_FOUND"
	CodeSyncNotConfigured = "SYNC_NOT_CONFIGURED"
)

// Error represents a structured error with code and details.
type Error struct {
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Details  map[string]any `json:"details,omitempty"`
	ExitCode int            `json:"-"`
	cause    error          // unexported: enables Go error chain traversal, not serialized
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause error, enabling errors.Is and errors.As
// to traverse through Wrap-ed errors.
func (e *Error) Unwrap() error {
	return e.cause
}

// JSON returns the error as a JSON object with "error" wrapper.
func (e *Error) JSON() string {
	wrapper := struct {
		Error *Error `json:"error"`
	}{Error: e}
	data, _ := json.MarshalIndent(wrapper, "", "  ")
	return string(data)
}

// New creates a new Error with the given code and message.
func New(code string, message string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		ExitCode: exitCodeForCode(code),
	}
}

// NewWithDetails creates a new Error with details.
func NewWithDetails(code string, message string, details map[string]any) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Details:  details,
		ExitCode: exitCodeForCode(code),
	}
}

// Wrap wraps an existing error with additional context.
// The cause is stored both as a string in Details["cause"] (for JSON serialization)
// and as the unexported cause field (for Go error chain traversal via Unwrap).
func Wrap(code string, message string, cause error) *Error {
	details := make(map[string]any)
	if cause != nil {
		details["cause"] = cause.Error()
	}
	return &Error{
		Code:     code,
		Message:  message,
		Details:  details,
		ExitCode: exitCodeForCode(code),
		cause:    cause,
	}
}

// exitCodeForCode maps error codes to exit codes.
func exitCodeForCode(code string) int {
	switch code {
	case CodeUsageError:
		return ExitUsageError
	case CodeLockTimeout, CodeLockStale:
		return ExitLockTimeout
	case CodeSchemaInvalid:
		return ExitSchemaInvalid
	case CodeLifecycleViolation, CodeOrphanConflict:
		return ExitLifecycleError
	case CodeFileNotFound, CodeSessionNotFound, CodeRiteNotFound:
		return ExitFileNotFound
	case CodePermissionDenied:
		return ExitPermissionDenied
	case CodeMergeConflict:
		return ExitMergeConflict
	case CodeProjectNotFound:
		return ExitProjectNotFound
	case CodeSessionExists:
		return ExitSessionExists
	case CodeMigrationFailed:
		return ExitMigrationFailed
	case CodeValidationFailed:
		return ExitValidationFailed
	case CodeSwitchAborted:
		return ExitSwitchAborted
	case CodeSchemaNotFound:
		return ExitSchemaNotFound
	case CodeParseError:
		return ExitParseError
	case CodeSyncStateCorrupt:
		return ExitSyncStateCorrupt
	case CodeRemoteRejected:
		return ExitRemoteRejected
	case CodeNetworkError:
		return ExitNetworkError
	case CodeRemoteNotFound:
		return ExitFileNotFound // Reuse FILE_NOT_FOUND exit code
	case CodeSyncNotConfigured:
		return ExitSyncNotConfigured
	case CodeBorrowConflict:
		return ExitLifecycleError
	case CodeBudgetExceeded:
		return ExitBudgetExceeded
	case CodeInvalidRiteForm:
		return ExitUsageError
	case CodeInvocationNotFound:
		return ExitFileNotFound
	case CodeQualityGateFailed:
		return ExitQualityGateFailed
	case CodeSignatureInvalid, CodeTimestampExpired:
		return ExitPermissionDenied
	case CodeServerStartFailed:
		return ExitGeneralError
	default:
		return ExitGeneralError
	}
}

// Common error constructors for convenience

// ErrProjectNotFound returns an error for missing channel directory.
func ErrProjectNotFound() *Error {
	return New(CodeProjectNotFound, "No channel directory found. Run from within a project or use --project-dir.")
}

// ErrSessionNotFound returns an error for missing session.
func ErrSessionNotFound(sessionID string) *Error {
	return NewWithDetails(CodeSessionNotFound,
		fmt.Sprintf("Session not found: %s", sessionID),
		map[string]any{"session_id": sessionID})
}

// ErrSessionExists returns an error when trying to create a session that exists.
func ErrSessionExists(existingID string, status string) *Error {
	return NewWithDetails(CodeSessionExists,
		"Session already active. Use 'ari session park' first or 'ari session wrap' to finalize.",
		map[string]any{
			"existing_session": existingID,
			"status":           status,
		})
}

// ErrLifecycleViolation returns an error for invalid state transitions.
func ErrLifecycleViolation(from, to string, reason string) *Error {
	return NewWithDetails(CodeLifecycleViolation,
		fmt.Sprintf("Cannot transition: %s", reason),
		map[string]any{
			"current_status":       from,
			"requested_transition": fmt.Sprintf("%s -> %s", from, to),
		})
}

// ErrLockTimeout returns an error when lock acquisition times out.
// The lockMeta parameter accepts structured metadata about the lock holder (or nil).
func ErrLockTimeout(lockPath string, lockMeta any) *Error {
	details := map[string]any{"lock_path": lockPath}
	if lockMeta != nil {
		details["lock_holder"] = lockMeta
	}
	return NewWithDetails(CodeLockTimeout,
		"Could not acquire lock within timeout. Run 'ari session recover' to clean up stale locks.",
		details)
}

// ErrSchemaInvalid returns an error for schema validation failures.
func ErrSchemaInvalid(path string, issues []string) *Error {
	return NewWithDetails(CodeSchemaInvalid,
		"Schema validation failed",
		map[string]any{
			"path":   path,
			"issues": issues,
		})
}

// ErrMigrationFailed returns an error for migration failures.
func ErrMigrationFailed(sessionID string, reason string) *Error {
	return NewWithDetails(CodeMigrationFailed,
		fmt.Sprintf("Migration failed: %s", reason),
		map[string]any{"session_id": sessionID})
}

// isCode checks whether err contains an *Error with one of the given codes.
// Uses errors.As for chain traversal through fmt.Errorf("%w", ...) wrappers.
func isCode(err error, codes ...string) bool {
	var e *Error
	if !stderrors.As(err, &e) {
		return false
	}
	return slices.Contains(codes, e.Code)
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	return isCode(err, CodeFileNotFound, CodeSessionNotFound, CodeProjectNotFound)
}

// IsLifecycleError returns true if the error is a lifecycle violation.
func IsLifecycleError(err error) bool {
	return isCode(err, CodeLifecycleViolation)
}

// GetExitCode extracts the exit code from an error.
// Uses errors.As to traverse wrapped errors (e.g. fmt.Errorf("%w", domainErr)).
// Returns ExitGeneralError if no *Error is found in the chain.
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	var e *Error
	if stderrors.As(err, &e) {
		return e.ExitCode
	}
	return ExitGeneralError
}

// --- Rite-domain error constructors ---

// ErrOrphanConflict returns an error when orphaned agents detected without strategy.
func ErrOrphanConflict(orphans []string, currentRite, targetRite string) *Error {
	return NewWithDetails(CodeOrphanConflict,
		"Orphaned agents detected. Use --keep-orphans to preserve, or --overwrite-diverged to force update",
		map[string]any{
			"orphans":      orphans,
			"current_rite": currentRite,
			"target_rite":  targetRite,
		})
}

// ErrValidationFailed returns an error for rite validation failures.
func ErrValidationFailed(riteName string, errorCount int, issues []string) *Error {
	return NewWithDetails(CodeValidationFailed,
		fmt.Sprintf("Rite validation failed with %d errors", errorCount),
		map[string]any{
			"rite":   riteName,
			"issues": issues,
		})
}

// ErrSwitchAborted returns an error when rite switch is rolled back.
func ErrSwitchAborted(riteName string, reason string) *Error {
	return NewWithDetails(CodeSwitchAborted,
		fmt.Sprintf("Rite switch aborted: %s", reason),
		map[string]any{"rite": riteName})
}

// IsOrphanConflict returns true if the error is an orphan conflict.
func IsOrphanConflict(err error) bool {
	return isCode(err, CodeOrphanConflict)
}

// --- Manifest-domain error constructors ---

// ErrMergeConflict returns an error for merge conflicts.
func ErrMergeConflict(conflictPaths []string, outputPath string) *Error {
	return NewWithDetails(CodeMergeConflict,
		"Three-way merge has unresolved conflicts",
		map[string]any{
			"conflict_count": len(conflictPaths),
			"conflicts":      conflictPaths,
			"output_path":    outputPath,
		})
}

// ErrSchemaNotFound returns an error for missing schema.
func ErrSchemaNotFound(schemaName string) *Error {
	return NewWithDetails(CodeSchemaNotFound,
		fmt.Sprintf("Schema not found: %s", schemaName),
		map[string]any{"schema": schemaName})
}

// ErrParseError returns an error for parsing failures.
func ErrParseError(path string, format string, cause error) *Error {
	details := map[string]any{
		"path":   path,
		"format": format,
	}
	if cause != nil {
		details["cause"] = cause.Error()
	}
	e := NewWithDetails(CodeParseError,
		fmt.Sprintf("Failed to parse %s file: %s", format, path),
		details)
	e.cause = cause
	return e
}

// IsMergeConflict returns true if the error is a merge conflict.
func IsMergeConflict(err error) bool {
	return isCode(err, CodeMergeConflict)
}

// IsSchemaNotFound returns true if the error is a schema not found error.
func IsSchemaNotFound(err error) bool {
	return isCode(err, CodeSchemaNotFound)
}

// IsParseError returns true if the error is a parse error.
func IsParseError(err error) bool {
	return isCode(err, CodeParseError)
}

// --- Sync-domain error constructors ---

// ErrRemoteNotFound returns an error for missing remote.
func ErrRemoteNotFound(remoteName string) *Error {
	return NewWithDetails(CodeRemoteNotFound,
		fmt.Sprintf("Remote not found: %s", remoteName),
		map[string]any{"remote": remoteName})
}

// ErrSyncConflict returns an error for sync conflicts.
func ErrSyncConflict(conflicts []string) *Error {
	return NewWithDetails(CodeMergeConflict,
		"Sync pull has unresolved conflicts",
		map[string]any{
			"conflict_count":   len(conflicts),
			"conflicts":        conflicts,
			"resolution_hint": "Resolve conflicts manually or re-run sync with --overwrite-diverged",
		})
}

// ErrSyncStateCorrupt returns an error for corrupt sync state.
func ErrSyncStateCorrupt(path string, reason string) *Error {
	return NewWithDetails(CodeSyncStateCorrupt,
		fmt.Sprintf("Sync state corrupt: %s", reason),
		map[string]any{
			"path":   path,
			"reason": reason,
		})
}

// ErrNetworkError returns an error for network failures.
func ErrNetworkError(url string, cause error) *Error {
	details := map[string]any{"url": url}
	if cause != nil {
		details["cause"] = cause.Error()
	}
	e := NewWithDetails(CodeNetworkError,
		fmt.Sprintf("Network error fetching %s", url),
		details)
	e.cause = cause
	return e
}

// ErrRemoteRejected returns an error when push is rejected.
func ErrRemoteRejected(remote string, reason string) *Error {
	return NewWithDetails(CodeRemoteRejected,
		fmt.Sprintf("Push rejected by %s: %s", remote, reason),
		map[string]any{
			"remote": remote,
			"reason": reason,
		})
}

// IsSyncStateCorrupt returns true if the error is a sync state corrupt error.
func IsSyncStateCorrupt(err error) bool {
	return isCode(err, CodeSyncStateCorrupt)
}

// IsNetworkError returns true if the error is a network error.
func IsNetworkError(err error) bool {
	return isCode(err, CodeNetworkError)
}

// IsRemoteRejected returns true if the error is a remote rejected error.
func IsRemoteRejected(err error) bool {
	return isCode(err, CodeRemoteRejected)
}

// IsRemoteNotFound returns true if the error is a remote not found error.
func IsRemoteNotFound(err error) bool {
	return isCode(err, CodeRemoteNotFound)
}

// IsSyncNotConfigured returns true if the error is a sync not configured error.
func IsSyncNotConfigured(err error) bool {
	return isCode(err, CodeSyncNotConfigured)
}

// --- Serve-domain error codes ---

const (
	CodeSignatureInvalid  = "SIGNATURE_INVALID"
	CodeTimestampExpired  = "TIMESTAMP_EXPIRED"
	CodeServerStartFailed = "SERVER_START_FAILED"
)

// --- Rite-domain error codes ---

const (
	// Rite-specific error codes
	CodeRiteNotFound       = "RITE_NOT_FOUND"
	CodeBorrowConflict     = "BORROW_CONFLICT"
	CodeBudgetExceeded     = "BUDGET_EXCEEDED"
	CodeInvalidRiteForm    = "INVALID_RITE_FORM"
	CodeInvocationNotFound = "INVOCATION_NOT_FOUND"

	// Quality gate error codes
	CodeQualityGateFailed = "QUALITY_GATE_FAILED"
)

// Exit codes for rite errors
const (
	ExitBudgetExceeded    = 19 // Context budget would be exceeded
	ExitQualityGateFailed = 20 // Quality gate check failed (BLACK sails)
)

// --- Rite-domain error constructors ---

// ErrRiteNotFound returns an error for missing rite.
func ErrRiteNotFound(riteName string) *Error {
	return NewWithDetails(CodeRiteNotFound,
		fmt.Sprintf("Rite not found: %s", riteName),
		map[string]any{"rite": riteName})
}

// ErrBorrowConflict returns an error when borrowing would conflict with existing invocations.
func ErrBorrowConflict(conflicts []string) *Error {
	return NewWithDetails(CodeBorrowConflict,
		"Borrowing would conflict with existing invocations",
		map[string]any{"conflicts": conflicts})
}

// ErrBudgetExceeded returns an error when context budget would be exceeded.
func ErrBudgetExceeded(current, requested, limit int) *Error {
	return NewWithDetails(CodeBudgetExceeded,
		fmt.Sprintf("Context budget exceeded: %d + %d > %d", current, requested, limit),
		map[string]any{
			"current":   current,
			"requested": requested,
			"limit":     limit,
		})
}

// ErrInvalidRiteForm returns an error when rite form doesn't support requested component.
func ErrInvalidRiteForm(form, required string) *Error {
	return NewWithDetails(CodeInvalidRiteForm,
		fmt.Sprintf("Rite form '%s' does not support requested component '%s'", form, required),
		map[string]any{
			"form":     form,
			"required": required,
		})
}

// ErrInvocationNotFound returns an error for missing invocation.
func ErrInvocationNotFound(id string) *Error {
	return NewWithDetails(CodeInvocationNotFound,
		fmt.Sprintf("Invocation not found: %s", id),
		map[string]any{"invocation_id": id})
}

// IsRiteNotFound returns true if the error is a rite not found error.
func IsRiteNotFound(err error) bool {
	return isCode(err, CodeRiteNotFound)
}

// IsBorrowConflict returns true if the error is a borrow conflict.
func IsBorrowConflict(err error) bool {
	return isCode(err, CodeBorrowConflict)
}

// IsBudgetExceeded returns true if the error is a budget exceeded error.
func IsBudgetExceeded(err error) bool {
	return isCode(err, CodeBudgetExceeded)
}

// IsInvocationNotFound returns true if the error is an invocation not found error.
func IsInvocationNotFound(err error) bool {
	return isCode(err, CodeInvocationNotFound)
}

// --- Handled error sentinel ---

// handledError wraps an error that has already been printed to the user.
// main.go should extract the exit code but NOT print it again.
type handledError struct {
	cause error
}

func (e *handledError) Error() string { return e.cause.Error() }
func (e *handledError) Unwrap() error { return e.cause }

// Handled wraps an error to indicate it was already printed to the user.
// GetExitCode still works through the wrapper via errors.As traversal.
func Handled(err error) error {
	if err == nil {
		return nil
	}
	return &handledError{cause: err}
}

// IsHandled returns true if the error has already been printed to the user.
func IsHandled(err error) bool {
	var h *handledError
	return stderrors.As(err, &h)
}

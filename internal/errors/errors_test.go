package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// --- Error struct construction and interface ---

func TestNew_SetsCodeMessageAndExitCode(t *testing.T) {
	t.Parallel()
	err := New(CodeFileNotFound, "file not found")
	if err.Code != CodeFileNotFound {
		t.Errorf("Code = %q, want %q", err.Code, CodeFileNotFound)
	}
	if err.Message != "file not found" {
		t.Errorf("Message = %q, want %q", err.Message, "file not found")
	}
	if err.ExitCode != ExitFileNotFound {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitFileNotFound)
	}
	if err.Details != nil {
		t.Errorf("Details = %v, want nil", err.Details)
	}
}

func TestNewWithDetails_SetsAllFields(t *testing.T) {
	t.Parallel()
	details := map[string]any{"path": "/some/path"}
	err := NewWithDetails(CodeParseError, "parse failed", details)
	if err.Code != CodeParseError {
		t.Errorf("Code = %q, want %q", err.Code, CodeParseError)
	}
	if err.Message != "parse failed" {
		t.Errorf("Message = %q, want %q", err.Message, "parse failed")
	}
	if err.Details["path"] != "/some/path" {
		t.Errorf("Details[path] = %v, want /some/path", err.Details["path"])
	}
	if err.ExitCode != ExitParseError {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitParseError)
	}
}

func TestWrap_StoresCauseInDetailsAsString(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("underlying io error")
	err := Wrap(CodeGeneralError, "operation failed", cause)
	if err.Code != CodeGeneralError {
		t.Errorf("Code = %q, want %q", err.Code, CodeGeneralError)
	}
	if err.Message != "operation failed" {
		t.Errorf("Message = %q, want %q", err.Message, "operation failed")
	}
	causeVal, ok := err.Details["cause"]
	if !ok {
		t.Fatal("Details[cause] not set")
	}
	if causeVal != "underlying io error" {
		t.Errorf("Details[cause] = %q, want %q", causeVal, "underlying io error")
	}
}

func TestWrap_NilCause_EmptyDetails(t *testing.T) {
	t.Parallel()
	err := Wrap(CodeGeneralError, "no cause", nil)
	if _, ok := err.Details["cause"]; ok {
		t.Errorf("Details[cause] set for nil cause, want absent")
	}
	// Details map is still allocated, just empty
	if err.Details == nil {
		t.Errorf("Details map should be non-nil even for nil cause")
	}
}

// Wrap preserves the cause for Go error chain traversal via Unwrap().
func TestWrap_UnwrapsFindsSentinelThroughChain(t *testing.T) {
	t.Parallel()
	sentinel := fmt.Errorf("sentinel error")
	wrapped := Wrap(CodeGeneralError, "wrapping", sentinel)
	// Unwrap returns the cause, so isStdlibIs can find the sentinel
	if !isStdlibIs(wrapped, sentinel) {
		t.Errorf("stdlib errors.Is should find sentinel through *Error Unwrap chain")
	}
}

// When *Error itself is wrapped by fmt.Errorf("%w"), errors.As can find it.
func TestErrorsAs_FindsErrorThroughFmtWrapping(t *testing.T) {
	t.Parallel()
	inner := New(CodeFileNotFound, "not found")
	wrapped := fmt.Errorf("context: %w", inner)
	var target *Error
	if !isErrorsAs(wrapped, &target) {
		t.Fatal("errors.As should find *Error through fmt.Errorf wrapping")
	}
	if target.Code != CodeFileNotFound {
		t.Errorf("Code = %q, want %q", target.Code, CodeFileNotFound)
	}
}

// IsNotFound works through fmt.Errorf("%w") wrapping.
func TestIsNotFound_ThroughFmtWrapping(t *testing.T) {
	t.Parallel()
	inner := New(CodeFileNotFound, "not found")
	wrapped := fmt.Errorf("loading config: %w", inner)
	if !IsNotFound(wrapped) {
		t.Errorf("IsNotFound should detect *Error through fmt.Errorf wrapping")
	}
}

// isErrorsAs performs stdlib errors.As without importing "errors" (which would shadow the package).
func isErrorsAs(err error, target **Error) bool {
	type unwrapper interface{ Unwrap() error }
	for e := err; e != nil; {
		if ce, ok := e.(*Error); ok {
			*target = ce
			return true
		}
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

// isStdlibIs walks the error chain to check identity, avoiding importing "errors".
func isStdlibIs(err error, target error) bool {
	// Walk the error chain manually using Unwrap (stdlib pattern)
	for e := err; e != nil; {
		if e == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

// --- Error() string output ---

func TestError_ReturnsMessage(t *testing.T) {
	t.Parallel()
	err := New(CodeUsageError, "invalid argument")
	if err.Error() != "invalid argument" {
		t.Errorf("Error() = %q, want %q", err.Error(), "invalid argument")
	}
}

// --- JSON() output format ---

func TestJSON_ProducesWrappedEnvelope(t *testing.T) {
	t.Parallel()
	err := NewWithDetails(CodeFileNotFound, "not found", map[string]any{"path": "/x"})
	raw := err.JSON()

	var envelope struct {
		Error struct {
			Code    string                 `json:"code"`
			Message string                 `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal([]byte(raw), &envelope); jsonErr != nil {
		t.Fatalf("JSON() output is not valid JSON: %v\nOutput: %s", jsonErr, raw)
	}
	if envelope.Error.Code != CodeFileNotFound {
		t.Errorf("JSON code = %q, want %q", envelope.Error.Code, CodeFileNotFound)
	}
	if envelope.Error.Message != "not found" {
		t.Errorf("JSON message = %q, want %q", envelope.Error.Message, "not found")
	}
	if envelope.Error.Details["path"] != "/x" {
		t.Errorf("JSON details.path = %v, want /x", envelope.Error.Details["path"])
	}
}

func TestJSON_ExitCodeOmitted(t *testing.T) {
	t.Parallel()
	err := New(CodeUsageError, "bad flag")
	raw := err.JSON()
	if strings.Contains(raw, "exit_code") || strings.Contains(raw, "ExitCode") {
		t.Errorf("JSON() should not include ExitCode, got: %s", raw)
	}
}

func TestJSON_NilDetails_OmitsDetailsKey(t *testing.T) {
	t.Parallel()
	err := New(CodeGeneralError, "bare error")
	raw := err.JSON()
	if strings.Contains(raw, `"details"`) {
		t.Errorf("JSON() should omit details when nil, got: %s", raw)
	}
}

// --- exitCodeForCode mapping table ---

func TestExitCodeMapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code     string
		wantExit int
	}{
		{CodeGeneralError, ExitGeneralError},
		{CodeUsageError, ExitUsageError},
		{CodeLockTimeout, ExitLockTimeout},
		{CodeLockStale, ExitLockTimeout}, // shares exit code with LockTimeout
		{CodeSchemaInvalid, ExitSchemaInvalid},
		{CodeLifecycleViolation, ExitLifecycleError},
		{CodeOrphanConflict, ExitLifecycleError}, // shares exit code with LifecycleViolation
		{CodeFileNotFound, ExitFileNotFound},
		{CodeSessionNotFound, ExitFileNotFound},    // shares exit code
		{CodeRiteNotFound, ExitFileNotFound},       // shares exit code
		{CodeRemoteNotFound, ExitFileNotFound},     // shares exit code (reuse FILE_NOT_FOUND)
		{CodeInvocationNotFound, ExitFileNotFound}, // shares exit code
		{CodePermissionDenied, ExitPermissionDenied},
		{CodeMergeConflict, ExitMergeConflict},
		{CodeProjectNotFound, ExitProjectNotFound},
		{CodeSessionExists, ExitSessionExists},
		{CodeMigrationFailed, ExitMigrationFailed},
		{CodeValidationFailed, ExitValidationFailed},
		{CodeSwitchAborted, ExitSwitchAborted},
		{CodeSchemaNotFound, ExitSchemaNotFound},
		{CodeParseError, ExitParseError},
		{CodeSyncStateCorrupt, ExitSyncStateCorrupt},
		{CodeRemoteRejected, ExitRemoteRejected},
		{CodeNetworkError, ExitNetworkError},
		{CodeSyncNotConfigured, ExitSyncNotConfigured},
		{CodeBorrowConflict, ExitLifecycleError},
		{CodeBudgetExceeded, ExitBudgetExceeded},
		{CodeInvalidRiteForm, ExitUsageError},
		{CodeQualityGateFailed, ExitQualityGateFailed},
		{"UNKNOWN_CODE_XYZ", ExitGeneralError}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()
			err := New(tt.code, "msg")
			if err.ExitCode != tt.wantExit {
				t.Errorf("New(%q).ExitCode = %d, want %d", tt.code, err.ExitCode, tt.wantExit)
			}
		})
	}
}

// --- GetExitCode ---

func TestGetExitCode_NilError_ReturnsSuccess(t *testing.T) {
	t.Parallel()
	if got := GetExitCode(nil); got != ExitSuccess {
		t.Errorf("GetExitCode(nil) = %d, want %d", got, ExitSuccess)
	}
}

func TestGetExitCode_CustomError_ReturnsCorrectCode(t *testing.T) {
	t.Parallel()
	err := New(CodeValidationFailed, "failed")
	if got := GetExitCode(err); got != ExitValidationFailed {
		t.Errorf("GetExitCode = %d, want %d", got, ExitValidationFailed)
	}
}

func TestGetExitCode_StdlibError_ReturnsGeneralError(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("plain error")
	if got := GetExitCode(err); got != ExitGeneralError {
		t.Errorf("GetExitCode(stdlib err) = %d, want %d", got, ExitGeneralError)
	}
}

func TestGetExitCode_FmtWrappedCustomError_TraversesChain(t *testing.T) {
	t.Parallel()
	// errors.As traverses through fmt.Errorf("%w") to find the inner *Error
	inner := New(CodeFileNotFound, "not found")
	wrapped := fmt.Errorf("outer: %w", inner)
	if got := GetExitCode(wrapped); got != ExitFileNotFound {
		t.Errorf("GetExitCode(fmt.Errorf wrapping) = %d, want %d (chain traversal)", got, ExitFileNotFound)
	}
}

// --- Domain-specific error constructors ---

func TestErrProjectNotFound(t *testing.T) {
	t.Parallel()
	err := ErrProjectNotFound()
	if err.Code != CodeProjectNotFound {
		t.Errorf("Code = %q, want %q", err.Code, CodeProjectNotFound)
	}
	if err.ExitCode != ExitProjectNotFound {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitProjectNotFound)
	}
	if !strings.Contains(err.Message, ".claude/") {
		t.Errorf("Message should mention .claude/, got: %q", err.Message)
	}
}

func TestErrSessionNotFound(t *testing.T) {
	t.Parallel()
	err := ErrSessionNotFound("sess-abc-123")
	if err.Code != CodeSessionNotFound {
		t.Errorf("Code = %q, want %q", err.Code, CodeSessionNotFound)
	}
	if !strings.Contains(err.Message, "sess-abc-123") {
		t.Errorf("Message should contain session ID, got: %q", err.Message)
	}
	if err.Details["session_id"] != "sess-abc-123" {
		t.Errorf("Details[session_id] = %v, want sess-abc-123", err.Details["session_id"])
	}
	if err.ExitCode != ExitFileNotFound {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitFileNotFound)
	}
}

func TestErrSessionExists(t *testing.T) {
	t.Parallel()
	err := ErrSessionExists("existing-id", "ACTIVE")
	if err.Code != CodeSessionExists {
		t.Errorf("Code = %q, want %q", err.Code, CodeSessionExists)
	}
	if err.ExitCode != ExitSessionExists {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitSessionExists)
	}
	if err.Details["existing_session"] != "existing-id" {
		t.Errorf("Details[existing_session] = %v, want existing-id", err.Details["existing_session"])
	}
	if err.Details["status"] != "ACTIVE" {
		t.Errorf("Details[status] = %v, want ACTIVE", err.Details["status"])
	}
}

func TestErrLifecycleViolation(t *testing.T) {
	t.Parallel()
	err := ErrLifecycleViolation("ACTIVE", "ARCHIVED", "cannot archive from active")
	if err.Code != CodeLifecycleViolation {
		t.Errorf("Code = %q, want %q", err.Code, CodeLifecycleViolation)
	}
	if err.ExitCode != ExitLifecycleError {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitLifecycleError)
	}
	if !strings.Contains(err.Message, "cannot archive from active") {
		t.Errorf("Message should contain reason, got: %q", err.Message)
	}
	if err.Details["current_status"] != "ACTIVE" {
		t.Errorf("Details[current_status] = %v, want ACTIVE", err.Details["current_status"])
	}
	transition, _ := err.Details["requested_transition"].(string)
	if !strings.Contains(transition, "ACTIVE") || !strings.Contains(transition, "ARCHIVED") {
		t.Errorf("Details[requested_transition] should show from->to, got: %q", transition)
	}
}

func TestErrLockTimeout_WithMeta(t *testing.T) {
	t.Parallel()
	meta := map[string]string{"pid": "1234"}
	err := ErrLockTimeout("/var/lock/ari.lock", meta)
	if err.Code != CodeLockTimeout {
		t.Errorf("Code = %q, want %q", err.Code, CodeLockTimeout)
	}
	if err.ExitCode != ExitLockTimeout {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitLockTimeout)
	}
	if err.Details["lock_path"] != "/var/lock/ari.lock" {
		t.Errorf("Details[lock_path] = %v", err.Details["lock_path"])
	}
	if err.Details["lock_holder"] == nil {
		t.Errorf("Details[lock_holder] should be set when meta is non-nil")
	}
}

func TestErrLockTimeout_NilMeta_NoLockHolder(t *testing.T) {
	t.Parallel()
	err := ErrLockTimeout("/var/lock/ari.lock", nil)
	if _, ok := err.Details["lock_holder"]; ok {
		t.Errorf("Details[lock_holder] should be absent when meta is nil")
	}
}

func TestErrSchemaInvalid(t *testing.T) {
	t.Parallel()
	issues := []string{"field 'x' required", "field 'y' invalid"}
	err := ErrSchemaInvalid("/schema/path.json", issues)
	if err.Code != CodeSchemaInvalid {
		t.Errorf("Code = %q, want %q", err.Code, CodeSchemaInvalid)
	}
	if err.ExitCode != ExitSchemaInvalid {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitSchemaInvalid)
	}
	if err.Details["path"] != "/schema/path.json" {
		t.Errorf("Details[path] = %v", err.Details["path"])
	}
}

func TestErrMigrationFailed(t *testing.T) {
	t.Parallel()
	err := ErrMigrationFailed("sess-xyz", "schema version mismatch")
	if err.Code != CodeMigrationFailed {
		t.Errorf("Code = %q, want %q", err.Code, CodeMigrationFailed)
	}
	if err.ExitCode != ExitMigrationFailed {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitMigrationFailed)
	}
	if !strings.Contains(err.Message, "schema version mismatch") {
		t.Errorf("Message should contain reason, got: %q", err.Message)
	}
	if err.Details["session_id"] != "sess-xyz" {
		t.Errorf("Details[session_id] = %v", err.Details["session_id"])
	}
}

func TestErrOrphanConflict(t *testing.T) {
	t.Parallel()
	orphans := []string{"agent-a.md", "agent-b.md"}
	err := ErrOrphanConflict(orphans, "rite-old", "rite-new")
	if err.Code != CodeOrphanConflict {
		t.Errorf("Code = %q, want %q", err.Code, CodeOrphanConflict)
	}
	if err.ExitCode != ExitLifecycleError {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitLifecycleError)
	}
	if err.Details["current_rite"] != "rite-old" {
		t.Errorf("Details[current_rite] = %v, want rite-old", err.Details["current_rite"])
	}
	if err.Details["target_rite"] != "rite-new" {
		t.Errorf("Details[target_rite] = %v, want rite-new", err.Details["target_rite"])
	}
}

func TestErrValidationFailed(t *testing.T) {
	t.Parallel()
	issues := []string{"missing field x", "invalid value for y"}
	err := ErrValidationFailed("ecosystem", 2, issues)
	if err.Code != CodeValidationFailed {
		t.Errorf("Code = %q, want %q", err.Code, CodeValidationFailed)
	}
	if err.ExitCode != ExitValidationFailed {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitValidationFailed)
	}
	if !strings.Contains(err.Message, "2") {
		t.Errorf("Message should mention error count, got: %q", err.Message)
	}
	if err.Details["rite"] != "ecosystem" {
		t.Errorf("Details[rite] = %v", err.Details["rite"])
	}
}

func TestErrSwitchAborted(t *testing.T) {
	t.Parallel()
	err := ErrSwitchAborted("my-rite", "validation failed")
	if err.Code != CodeSwitchAborted {
		t.Errorf("Code = %q, want %q", err.Code, CodeSwitchAborted)
	}
	if err.ExitCode != ExitSwitchAborted {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitSwitchAborted)
	}
	if !strings.Contains(err.Message, "validation failed") {
		t.Errorf("Message should contain reason, got: %q", err.Message)
	}
}

func TestErrMergeConflict(t *testing.T) {
	t.Parallel()
	conflicts := []string{"file1.md", "file2.md"}
	err := ErrMergeConflict(conflicts, "/output/path.md")
	if err.Code != CodeMergeConflict {
		t.Errorf("Code = %q, want %q", err.Code, CodeMergeConflict)
	}
	if err.ExitCode != ExitMergeConflict {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitMergeConflict)
	}
	if err.Details["conflict_count"] != 2 {
		t.Errorf("Details[conflict_count] = %v, want 2", err.Details["conflict_count"])
	}
	if err.Details["output_path"] != "/output/path.md" {
		t.Errorf("Details[output_path] = %v", err.Details["output_path"])
	}
}

func TestErrSchemaNotFound(t *testing.T) {
	t.Parallel()
	err := ErrSchemaNotFound("session-v2")
	if err.Code != CodeSchemaNotFound {
		t.Errorf("Code = %q, want %q", err.Code, CodeSchemaNotFound)
	}
	if err.ExitCode != ExitSchemaNotFound {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitSchemaNotFound)
	}
	if !strings.Contains(err.Message, "session-v2") {
		t.Errorf("Message should contain schema name, got: %q", err.Message)
	}
	if err.Details["schema"] != "session-v2" {
		t.Errorf("Details[schema] = %v", err.Details["schema"])
	}
}

func TestErrParseError_WithCause(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("unexpected EOF")
	err := ErrParseError("/path/to/file.yaml", "YAML", cause)
	if err.Code != CodeParseError {
		t.Errorf("Code = %q, want %q", err.Code, CodeParseError)
	}
	if err.ExitCode != ExitParseError {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitParseError)
	}
	if !strings.Contains(err.Message, "YAML") {
		t.Errorf("Message should contain format, got: %q", err.Message)
	}
	if !strings.Contains(err.Message, "/path/to/file.yaml") {
		t.Errorf("Message should contain path, got: %q", err.Message)
	}
	if err.Details["cause"] != "unexpected EOF" {
		t.Errorf("Details[cause] = %v", err.Details["cause"])
	}
}

func TestErrParseError_NilCause(t *testing.T) {
	t.Parallel()
	err := ErrParseError("/path/to/file.json", "JSON", nil)
	if _, ok := err.Details["cause"]; ok {
		t.Errorf("Details[cause] should be absent when cause is nil")
	}
}

func TestErrRemoteNotFound(t *testing.T) {
	t.Parallel()
	err := ErrRemoteNotFound("origin")
	if err.Code != CodeRemoteNotFound {
		t.Errorf("Code = %q, want %q", err.Code, CodeRemoteNotFound)
	}
	// CodeRemoteNotFound reuses ExitFileNotFound per exitCodeForCode
	if err.ExitCode != ExitFileNotFound {
		t.Errorf("ExitCode = %d, want %d (reuses FILE_NOT_FOUND exit)", err.ExitCode, ExitFileNotFound)
	}
	if err.Details["remote"] != "origin" {
		t.Errorf("Details[remote] = %v", err.Details["remote"])
	}
}

func TestErrSyncConflict(t *testing.T) {
	t.Parallel()
	conflicts := []string{"agents/x.md"}
	err := ErrSyncConflict(conflicts)
	// ErrSyncConflict uses CodeMergeConflict internally
	if err.Code != CodeMergeConflict {
		t.Errorf("Code = %q, want %q", err.Code, CodeMergeConflict)
	}
	if err.ExitCode != ExitMergeConflict {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitMergeConflict)
	}
	if err.Details["conflict_count"] != 1 {
		t.Errorf("Details[conflict_count] = %v, want 1", err.Details["conflict_count"])
	}
}

func TestErrSyncStateCorrupt(t *testing.T) {
	t.Parallel()
	err := ErrSyncStateCorrupt("/path/state.json", "unexpected null")
	if err.Code != CodeSyncStateCorrupt {
		t.Errorf("Code = %q, want %q", err.Code, CodeSyncStateCorrupt)
	}
	if err.ExitCode != ExitSyncStateCorrupt {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitSyncStateCorrupt)
	}
	if !strings.Contains(err.Message, "unexpected null") {
		t.Errorf("Message should contain reason, got: %q", err.Message)
	}
}

func TestErrNetworkError_WithCause(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("connection refused")
	err := ErrNetworkError("https://example.com/api", cause)
	if err.Code != CodeNetworkError {
		t.Errorf("Code = %q, want %q", err.Code, CodeNetworkError)
	}
	if err.ExitCode != ExitNetworkError {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitNetworkError)
	}
	if err.Details["url"] != "https://example.com/api" {
		t.Errorf("Details[url] = %v", err.Details["url"])
	}
	if err.Details["cause"] != "connection refused" {
		t.Errorf("Details[cause] = %v", err.Details["cause"])
	}
}

func TestErrNetworkError_NilCause(t *testing.T) {
	t.Parallel()
	err := ErrNetworkError("https://example.com", nil)
	if _, ok := err.Details["cause"]; ok {
		t.Errorf("Details[cause] should be absent for nil cause")
	}
}

func TestErrRemoteRejected(t *testing.T) {
	t.Parallel()
	err := ErrRemoteRejected("origin", "non-fast-forward")
	if err.Code != CodeRemoteRejected {
		t.Errorf("Code = %q, want %q", err.Code, CodeRemoteRejected)
	}
	if err.ExitCode != ExitRemoteRejected {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitRemoteRejected)
	}
	if !strings.Contains(err.Message, "non-fast-forward") {
		t.Errorf("Message should contain reason, got: %q", err.Message)
	}
}

func TestErrRiteNotFound(t *testing.T) {
	t.Parallel()
	err := ErrRiteNotFound("my-rite")
	if err.Code != CodeRiteNotFound {
		t.Errorf("Code = %q, want %q", err.Code, CodeRiteNotFound)
	}
	if err.ExitCode != ExitFileNotFound {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitFileNotFound)
	}
	if !strings.Contains(err.Message, "my-rite") {
		t.Errorf("Message should contain rite name, got: %q", err.Message)
	}
	if err.Details["rite"] != "my-rite" {
		t.Errorf("Details[rite] = %v", err.Details["rite"])
	}
}

func TestErrBorrowConflict(t *testing.T) {
	t.Parallel()
	conflicts := []string{"invocation-1", "invocation-2"}
	err := ErrBorrowConflict(conflicts)
	if err.Code != CodeBorrowConflict {
		t.Errorf("Code = %q, want %q", err.Code, CodeBorrowConflict)
	}
	if err.ExitCode != ExitLifecycleError {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitLifecycleError)
	}
}

func TestErrBudgetExceeded(t *testing.T) {
	t.Parallel()
	err := ErrBudgetExceeded(80000, 30000, 100000)
	if err.Code != CodeBudgetExceeded {
		t.Errorf("Code = %q, want %q", err.Code, CodeBudgetExceeded)
	}
	if err.ExitCode != ExitBudgetExceeded {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitBudgetExceeded)
	}
	if !strings.Contains(err.Message, "80000") {
		t.Errorf("Message should contain current value, got: %q", err.Message)
	}
	if err.Details["current"] != 80000 {
		t.Errorf("Details[current] = %v, want 80000", err.Details["current"])
	}
	if err.Details["requested"] != 30000 {
		t.Errorf("Details[requested] = %v, want 30000", err.Details["requested"])
	}
	if err.Details["limit"] != 100000 {
		t.Errorf("Details[limit] = %v, want 100000", err.Details["limit"])
	}
}

func TestErrInvalidRiteForm(t *testing.T) {
	t.Parallel()
	err := ErrInvalidRiteForm("solo", "agents")
	if err.Code != CodeInvalidRiteForm {
		t.Errorf("Code = %q, want %q", err.Code, CodeInvalidRiteForm)
	}
	if err.ExitCode != ExitUsageError {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitUsageError)
	}
	if err.Details["form"] != "solo" {
		t.Errorf("Details[form] = %v", err.Details["form"])
	}
	if err.Details["required"] != "agents" {
		t.Errorf("Details[required] = %v", err.Details["required"])
	}
}

func TestErrInvocationNotFound(t *testing.T) {
	t.Parallel()
	err := ErrInvocationNotFound("inv-abc-123")
	if err.Code != CodeInvocationNotFound {
		t.Errorf("Code = %q, want %q", err.Code, CodeInvocationNotFound)
	}
	if err.ExitCode != ExitFileNotFound {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitFileNotFound)
	}
	if err.Details["invocation_id"] != "inv-abc-123" {
		t.Errorf("Details[invocation_id] = %v", err.Details["invocation_id"])
	}
}

// --- Predicate helpers: IsNotFound ---

func TestIsNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeFileNotFound", New(CodeFileNotFound, "m"), true},
		{"CodeSessionNotFound", New(CodeSessionNotFound, "m"), true},
		{"CodeProjectNotFound", New(CodeProjectNotFound, "m"), true},
		{"CodeGeneralError", New(CodeGeneralError, "m"), false},
		{"CodeUsageError", New(CodeUsageError, "m"), false},
		{"CodeRiteNotFound", New(CodeRiteNotFound, "m"), false}, // RiteNotFound is NOT in IsNotFound
		{"stdlib error", fmt.Errorf("not found"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsLifecycleError ---

func TestIsLifecycleError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeLifecycleViolation", New(CodeLifecycleViolation, "m"), true},
		{"CodeOrphanConflict", New(CodeOrphanConflict, "m"), false}, // OrphanConflict is NOT in IsLifecycleError
		{"CodeGeneralError", New(CodeGeneralError, "m"), false},
		{"stdlib error", fmt.Errorf("lifecycle"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsLifecycleError(tt.err); got != tt.want {
				t.Errorf("IsLifecycleError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsOrphanConflict ---

func TestIsOrphanConflict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeOrphanConflict", New(CodeOrphanConflict, "m"), true},
		{"CodeLifecycleViolation", New(CodeLifecycleViolation, "m"), false},
		{"stdlib error", fmt.Errorf("orphan"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsOrphanConflict(tt.err); got != tt.want {
				t.Errorf("IsOrphanConflict(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsMergeConflict ---

func TestIsMergeConflict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeMergeConflict", New(CodeMergeConflict, "m"), true},
		{"ErrSyncConflict also uses CodeMergeConflict", ErrSyncConflict([]string{"f"}), true},
		{"CodeGeneralError", New(CodeGeneralError, "m"), false},
		{"stdlib error", fmt.Errorf("conflict"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsMergeConflict(tt.err); got != tt.want {
				t.Errorf("IsMergeConflict(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsSchemaNotFound ---

func TestIsSchemaNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeSchemaNotFound", New(CodeSchemaNotFound, "m"), true},
		{"CodeSchemaInvalid", New(CodeSchemaInvalid, "m"), false},
		{"stdlib error", fmt.Errorf("schema"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsSchemaNotFound(tt.err); got != tt.want {
				t.Errorf("IsSchemaNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsParseError ---

func TestIsParseError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeParseError", New(CodeParseError, "m"), true},
		{"CodeSchemaInvalid", New(CodeSchemaInvalid, "m"), false},
		{"stdlib error", fmt.Errorf("parse"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsParseError(tt.err); got != tt.want {
				t.Errorf("IsParseError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsSyncStateCorrupt ---

func TestIsSyncStateCorrupt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeSyncStateCorrupt", New(CodeSyncStateCorrupt, "m"), true},
		{"CodeGeneralError", New(CodeGeneralError, "m"), false},
		{"stdlib error", fmt.Errorf("corrupt"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsSyncStateCorrupt(tt.err); got != tt.want {
				t.Errorf("IsSyncStateCorrupt(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsNetworkError ---

func TestIsNetworkError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeNetworkError", New(CodeNetworkError, "m"), true},
		{"CodeRemoteRejected", New(CodeRemoteRejected, "m"), false},
		{"stdlib error", fmt.Errorf("network"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsNetworkError(tt.err); got != tt.want {
				t.Errorf("IsNetworkError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsRemoteRejected ---

func TestIsRemoteRejected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeRemoteRejected", New(CodeRemoteRejected, "m"), true},
		{"CodeNetworkError", New(CodeNetworkError, "m"), false},
		{"stdlib error", fmt.Errorf("rejected"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRemoteRejected(tt.err); got != tt.want {
				t.Errorf("IsRemoteRejected(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsRemoteNotFound ---

func TestIsRemoteNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeRemoteNotFound", New(CodeRemoteNotFound, "m"), true},
		{"CodeFileNotFound", New(CodeFileNotFound, "m"), false},
		{"stdlib error", fmt.Errorf("remote"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRemoteNotFound(tt.err); got != tt.want {
				t.Errorf("IsRemoteNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsSyncNotConfigured ---

func TestIsSyncNotConfigured(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeSyncNotConfigured", New(CodeSyncNotConfigured, "m"), true},
		{"CodeGeneralError", New(CodeGeneralError, "m"), false},
		{"stdlib error", fmt.Errorf("not configured"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsSyncNotConfigured(tt.err); got != tt.want {
				t.Errorf("IsSyncNotConfigured(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsRiteNotFound ---

func TestIsRiteNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeRiteNotFound", New(CodeRiteNotFound, "m"), true},
		{"CodeFileNotFound", New(CodeFileNotFound, "m"), false},
		{"stdlib error", fmt.Errorf("rite"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRiteNotFound(tt.err); got != tt.want {
				t.Errorf("IsRiteNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsBorrowConflict ---

func TestIsBorrowConflict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeBorrowConflict", New(CodeBorrowConflict, "m"), true},
		{"CodeMergeConflict", New(CodeMergeConflict, "m"), false},
		{"stdlib error", fmt.Errorf("borrow"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsBorrowConflict(tt.err); got != tt.want {
				t.Errorf("IsBorrowConflict(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsBudgetExceeded ---

func TestIsBudgetExceeded(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeBudgetExceeded", New(CodeBudgetExceeded, "m"), true},
		{"CodeGeneralError", New(CodeGeneralError, "m"), false},
		{"stdlib error", fmt.Errorf("budget"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsBudgetExceeded(tt.err); got != tt.want {
				t.Errorf("IsBudgetExceeded(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Predicate helpers: IsInvocationNotFound ---

func TestIsInvocationNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"CodeInvocationNotFound", New(CodeInvocationNotFound, "m"), true},
		{"CodeFileNotFound", New(CodeFileNotFound, "m"), false},
		{"stdlib error", fmt.Errorf("invocation"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsInvocationNotFound(tt.err); got != tt.want {
				t.Errorf("IsInvocationNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Error chain traversal ---

func TestWrap_CauseStoredAsStringAndTraversable(t *testing.T) {
	t.Parallel()
	// Wrap stores cause both as string in Details (for JSON) and via Unwrap (for Go chain).
	inner := fmt.Errorf("inner io error")
	outer := Wrap(CodeFileNotFound, "outer message", inner)

	// The outer *Error is detectable by its own predicate
	if !IsNotFound(outer) {
		t.Errorf("IsNotFound should detect outer *Error with CodeFileNotFound")
	}

	// The cause is accessible as a string in details
	if outer.Details["cause"] != "inner io error" {
		t.Errorf("cause detail = %v, want %q", outer.Details["cause"], "inner io error")
	}

	// The inner error IS traversable via Unwrap chain
	if !isStdlibIs(outer, inner) {
		t.Errorf("inner error should be findable via stdlib errors chain traversal")
	}
}

func TestWrap_DoubleWrapped_InnerCauseIsSecondWrapMessage(t *testing.T) {
	t.Parallel()
	// When wrapping a *Error with Wrap, the cause is the wrapped error's message string.
	inner := New(CodeFileNotFound, "inner not found")
	outer := Wrap(CodeGeneralError, "outer context", inner)

	// outer.Details["cause"] is inner.Error() == inner.Message
	if outer.Details["cause"] != "inner not found" {
		t.Errorf("cause detail = %v, want %q", outer.Details["cause"], "inner not found")
	}

	// IsNotFound only sees the outer code
	if IsNotFound(outer) {
		t.Errorf("IsNotFound should NOT match outer with CodeGeneralError")
	}
}

// --- Error satisfies standard error interface ---

func TestError_ImplementsErrorInterface(t *testing.T) {
	t.Parallel()
	var err error = New(CodeGeneralError, "test")
	if err.Error() != "test" {
		t.Errorf("error.Error() = %q, want %q", err.Error(), "test")
	}
}

// --- Structured field access ---

func TestError_FieldsDirectlyAccessible(t *testing.T) {
	t.Parallel()
	details := map[string]any{"key": "value"}
	err := NewWithDetails(CodeUsageError, "bad flag --foo", details)

	if err.Code != CodeUsageError {
		t.Errorf("err.Code = %q", err.Code)
	}
	if err.Message != "bad flag --foo" {
		t.Errorf("err.Message = %q", err.Message)
	}
	if err.ExitCode != ExitUsageError {
		t.Errorf("err.ExitCode = %d", err.ExitCode)
	}
	if err.Details["key"] != "value" {
		t.Errorf("err.Details[key] = %v", err.Details["key"])
	}
}

// --- Exit code constants have expected values ---

func TestExitCodeConstants(t *testing.T) {
	t.Parallel()
	// Verify the constants match their documented values from TDD Section 4.1
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitGeneralError", ExitGeneralError, 1},
		{"ExitUsageError", ExitUsageError, 2},
		{"ExitLockTimeout", ExitLockTimeout, 3},
		{"ExitSchemaInvalid", ExitSchemaInvalid, 4},
		{"ExitLifecycleError", ExitLifecycleError, 5},
		{"ExitFileNotFound", ExitFileNotFound, 6},
		{"ExitPermissionDenied", ExitPermissionDenied, 7},
		{"ExitMergeConflict", ExitMergeConflict, 8},
		{"ExitProjectNotFound", ExitProjectNotFound, 9},
		{"ExitSessionExists", ExitSessionExists, 10},
		{"ExitMigrationFailed", ExitMigrationFailed, 11},
		{"ExitValidationFailed", ExitValidationFailed, 12},
		{"ExitSwitchAborted", ExitSwitchAborted, 13},
		{"ExitSchemaNotFound", ExitSchemaNotFound, 14},
		{"ExitParseError", ExitParseError, 15},
		{"ExitSyncStateCorrupt", ExitSyncStateCorrupt, 16},
		{"ExitRemoteRejected", ExitRemoteRejected, 17},
		{"ExitNetworkError", ExitNetworkError, 18},
		{"ExitBudgetExceeded", ExitBudgetExceeded, 19},
		{"ExitQualityGateFailed", ExitQualityGateFailed, 20},
		{"ExitSyncNotConfigured", ExitSyncNotConfigured, 21},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

// --- Error code string constants equal their own names (SCREAMING_SNAKE_CASE) ---

func TestErrorCodeStrings_SelfNamed(t *testing.T) {
	t.Parallel()
	// Error codes are self-naming: the constant value equals the constant name.
	tests := []struct {
		name string
		code string
	}{
		{"CodeGeneralError", CodeGeneralError},
		{"CodeUsageError", CodeUsageError},
		{"CodeLockTimeout", CodeLockTimeout},
		{"CodeLockStale", CodeLockStale},
		{"CodeSchemaInvalid", CodeSchemaInvalid},
		{"CodeLifecycleViolation", CodeLifecycleViolation},
		{"CodeFileNotFound", CodeFileNotFound},
		{"CodeSessionNotFound", CodeSessionNotFound},
		{"CodePermissionDenied", CodePermissionDenied},
		{"CodeMergeConflict", CodeMergeConflict},
		{"CodeProjectNotFound", CodeProjectNotFound},
		{"CodeSessionExists", CodeSessionExists},
		{"CodeMigrationFailed", CodeMigrationFailed},
		{"CodeOrphanConflict", CodeOrphanConflict},
		{"CodeValidationFailed", CodeValidationFailed},
		{"CodeSwitchAborted", CodeSwitchAborted},
		{"CodeSchemaNotFound", CodeSchemaNotFound},
		{"CodeParseError", CodeParseError},
		{"CodeSyncStateCorrupt", CodeSyncStateCorrupt},
		{"CodeRemoteRejected", CodeRemoteRejected},
		{"CodeNetworkError", CodeNetworkError},
		{"CodeRemoteNotFound", CodeRemoteNotFound},
		{"CodeSyncNotConfigured", CodeSyncNotConfigured},
		{"CodeRiteNotFound", CodeRiteNotFound},
		{"CodeBorrowConflict", CodeBorrowConflict},
		{"CodeBudgetExceeded", CodeBudgetExceeded},
		{"CodeInvalidRiteForm", CodeInvalidRiteForm},
		{"CodeInvocationNotFound", CodeInvocationNotFound},
		{"CodeQualityGateFailed", CodeQualityGateFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Verify the value is non-empty and all uppercase/underscore (SCREAMING_SNAKE_CASE)
			if tt.code == "" {
				t.Errorf("%s constant is empty string", tt.name)
			}
			for _, ch := range tt.code {
				if (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') && ch != '_' {
					t.Errorf("%s = %q contains non-SCREAMING_SNAKE_CASE char %q", tt.name, tt.code, ch)
				}
			}
		})
	}
}

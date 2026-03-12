package session

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
)

// createTestSession creates a session directory with SESSION_CONTEXT.md frontmatter.
// Session IDs follow pattern: session-YYYYMMDD-HHMMSS-{8-hex} (32+ chars)
func createTestSession(t *testing.T, sessionsDir, sessionID, status string) {
	t.Helper()
	// Ensure session ID matches expected pattern (32+ chars, starts with "session-")
	if len(sessionID) < 32 || sessionID[:8] != "session-" {
		t.Fatalf("invalid test session ID %q: must be 32+ chars starting with 'session-'", sessionID)
	}

	dir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	content := fmt.Sprintf("---\nstatus: %s\n---\n", status)
	contextFile := filepath.Join(dir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(contextFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}
}

// --- ResolveSession Tests ---

func TestResolveSession_ExplicitIDTakesPriority(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	// Setup: active session + cc-map entry
	createTestSession(t, sessionsDir, "session-20260209-120000-aaaaaaaa", "ACTIVE")

	ccMapDir := resolver.CCMapDir()
	os.MkdirAll(ccMapDir, 0755)
	os.WriteFile(filepath.Join(ccMapDir, "cc-test-id"), []byte("session-from-map"), 0644)

	// Test: explicit ID takes priority
	result, err := ResolveSession(resolver, "cc-test-id", "session-explicit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "session-explicit" {
		t.Errorf("expected explicit ID, got %q", result)
	}
}

func TestResolveSession_CCMapLookup(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	ccMapDir := resolver.CCMapDir()
	os.MkdirAll(ccMapDir, 0755)
	os.WriteFile(filepath.Join(ccMapDir, "cc-session-123"), []byte("session-mapped"), 0644)

	result, err := ResolveSession(resolver, "cc-session-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "session-mapped" {
		t.Errorf("expected session-mapped, got %q", result)
	}
}

func TestResolveSession_CCMapMissing_FallsToScan(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	// Create active session for scan fallback
	createTestSession(t, sessionsDir, "session-20260209-120000-bbbbbbbb", "ACTIVE")

	// Test: cc-map missing, falls through to scan
	result, err := ResolveSession(resolver, "nonexistent-cc-id", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "session-20260209-120000-bbbbbbbb" {
		t.Errorf("expected session-20260209-120000-bbbbbbbb from scan fallback, got %q", result)
	}
}

func TestResolveSession_ScanSingleActive(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	createTestSession(t, sessionsDir, "session-20260209-120000-bbbbbbbb", "ACTIVE")
	createTestSession(t, sessionsDir, "session-20260209-110000-cccccccc", "PARKED")

	result, err := ResolveSession(resolver, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "session-20260209-120000-bbbbbbbb" {
		t.Errorf("expected session-20260209-120000-bbbbbbbb, got %q", result)
	}
}

func TestResolveSession_ScanNoActive(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	createTestSession(t, sessionsDir, "session-20260209-110000-cccccccc", "PARKED")
	createTestSession(t, sessionsDir, "session-20260209-100000-dddddddd", "ARCHIVED")

	result, err := ResolveSession(resolver, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestResolveSession_ScanMultipleActive(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	createTestSession(t, sessionsDir, "session-20260209-120001-11111111", "ACTIVE")
	createTestSession(t, sessionsDir, "session-20260209-120002-22222222", "ACTIVE")

	result, err := ResolveSession(resolver, "", "")
	if err == nil {
		t.Fatalf("expected error for multiple active sessions, got result: %q", result)
	}
	if result != "" {
		t.Errorf("expected empty string on error, got %q", result)
	}
}

func TestResolveSession_CCMapStaleEntry(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	ccMapDir := resolver.CCMapDir()
	os.MkdirAll(ccMapDir, 0755)
	// Map points to non-existent session
	os.WriteFile(filepath.Join(ccMapDir, "cc-stale"), []byte("session-nonexistent"), 0644)

	// Should still return the mapped ID — caller decides validity
	result, err := ResolveSession(resolver, "cc-stale", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "session-nonexistent" {
		t.Errorf("expected session-nonexistent, got %q", result)
	}
}

// --- SetCCMap Tests ---

func TestSetCCMap_CreatesDirectoryAndFile(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	err := SetCCMap(resolver, "cc-new-session", "session-knossos-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mapFile := filepath.Join(resolver.CCMapDir(), "cc-new-session")
	data, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("failed to read map file: %v", err)
	}
	if string(data) != "session-knossos-123" {
		t.Errorf("expected session-knossos-123, got %q", string(data))
	}
}

func TestSetCCMap_OverwritesExisting(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	SetCCMap(resolver, "cc-test", "session-old")
	SetCCMap(resolver, "cc-test", "session-new")

	mapFile := filepath.Join(resolver.CCMapDir(), "cc-test")
	data, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("failed to read map file: %v", err)
	}
	if string(data) != "session-new" {
		t.Errorf("expected session-new, got %q", string(data))
	}
}

func TestSetCCMap_RejectsEmptyID(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	err := SetCCMap(resolver, "", "session-knossos-123")
	if err == nil {
		t.Fatal("expected error for empty CC session ID")
	}
}

func TestSetCCMap_SanitizesPathTraversal(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	// Attempt path traversal attack
	err := SetCCMap(resolver, "../../../etc/passwd", "session-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should sanitize to just "passwd"
	mapFile := filepath.Join(resolver.CCMapDir(), "passwd")
	data, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("failed to read map file: %v", err)
	}
	if string(data) != "session-test" {
		t.Errorf("expected session-test, got %q", string(data))
	}

	// Verify no files created outside cc-map dir
	attackPath := filepath.Join(projectRoot, "..", "..", "etc", "passwd")
	if _, err := os.Stat(attackPath); err == nil {
		t.Fatal("path traversal attack succeeded — file created outside cc-map")
	}
}

// --- ClearCCMap Tests ---

func TestClearCCMap_RemovesFile(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	SetCCMap(resolver, "cc-test", "session-123")

	err := ClearCCMap(resolver, "cc-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mapFile := filepath.Join(resolver.CCMapDir(), "cc-test")
	if _, err := os.Stat(mapFile); !os.IsNotExist(err) {
		t.Fatal("map file still exists after clear")
	}
}

func TestClearCCMap_NoErrorOnMissing(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)

	err := ClearCCMap(resolver, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error on missing file: %v", err)
	}
}

// --- FindActiveSessions Tests ---

func TestFindActiveSessions_Empty(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	result, err := FindActiveSessions(sessionsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestFindActiveSessions_OneActive(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	createTestSession(t, sessionsDir, "session-20260209-120000-bbbbbbbb", "ACTIVE")
	createTestSession(t, sessionsDir, "session-20260209-110000-cccccccc", "PARKED")

	result, err := FindActiveSessions(sessionsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(result))
	}
	if result[0] != "session-20260209-120000-bbbbbbbb" {
		t.Errorf("expected session-20260209-120000-bbbbbbbb, got %q", result[0])
	}
}

func TestFindActiveSessions_MultipleActive(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	createTestSession(t, sessionsDir, "session-20260209-120001-11111111", "ACTIVE")
	createTestSession(t, sessionsDir, "session-20260209-120002-22222222", "ACTIVE")

	result, err := FindActiveSessions(sessionsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(result))
	}

	// Check both IDs present (order not guaranteed)
	found := make(map[string]bool)
	for _, id := range result {
		found[id] = true
	}
	if !found["session-20260209-120001-11111111"] || !found["session-20260209-120002-22222222"] {
		t.Errorf("expected both session-20260209-120001-11111111 and session-20260209-120002-22222222, got %v", result)
	}
}

func TestFindActiveSessions_MixedStatuses(t *testing.T) {
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sessionsDir := resolver.SessionsDir()
	os.MkdirAll(sessionsDir, 0755)

	createTestSession(t, sessionsDir, "session-20260209-120000-bbbbbbbb", "ACTIVE")
	createTestSession(t, sessionsDir, "session-20260209-110000-cccccccc", "PARKED")
	createTestSession(t, sessionsDir, "session-20260209-100000-dddddddd", "ARCHIVED")

	result, err := FindActiveSessions(sessionsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 active session, got %d: %v", len(result), result)
	}
	if result[0] != "session-20260209-120000-bbbbbbbb" {
		t.Errorf("expected session-20260209-120000-bbbbbbbb, got %q", result[0])
	}
}

// --- sanitizeHarnessSessionID Tests ---

func TestSanitizeCCSessionID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal ID",
			input:    "cc-session-123",
			expected: "cc-session-123",
		},
		{
			name:     "with whitespace",
			input:    "  cc-session-456  ",
			expected: "cc-session-456",
		},
		{
			name:     "path traversal",
			input:    "../../../etc/passwd",
			expected: "passwd",
		},
		{
			name:     "complex path traversal",
			input:    "../../.ssh/id_rsa",
			expected: "id_rsa",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "single dot",
			input:    ".",
			expected: "",
		},
		{
			name:     "double dot",
			input:    "..",
			expected: "",
		},
		{
			name:     "with slash",
			input:    "foo/bar",
			expected: "bar",
		},
		{
			name:     "absolute path",
			input:    "/etc/passwd",
			expected: "passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeHarnessSessionID(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

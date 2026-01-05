package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/session"
)

// TestWrapGeneratesWhiteSails verifies that /wrap generates WHITE_SAILS.yaml.
func TestWrapGeneratesWhiteSails(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120000-abc12345"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create SESSION_CONTEXT.md
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Initiative
complexity: MODULE
created_at: 2025-01-05T12:00:00Z
---

# Session Context

## Session Type
standard

## Open Questions
- None
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Create proof log files with passing tests
	// Proof collector expects test-output.log, build-output.log, lint-output.log
	testLog := `=== RUN   TestExample
--- PASS: TestExample (0.00s)
PASS
ok  	example.com/pkg	0.123s
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "test-output.log"), []byte(testLog), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	buildLog := `Building example.com/pkg...
Build successful
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "build-output.log"), []byte(buildLog), 0644); err != nil {
		t.Fatalf("Failed to write build log: %v", err)
	}

	lintLog := `Linting example.com/pkg...
No issues found
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "lint-output.log"), []byte(lintLog), 0644); err != nil {
		t.Fatalf("Failed to write lint log: %v", err)
	}

	// Run wrap command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := wrapOptions{
		noArchive: true, // Don't move to archive for easier testing
	}

	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("runWrap failed: %v", err)
	}

	// Verify WHITE_SAILS.yaml was created
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	if _, err := os.Stat(sailsPath); os.IsNotExist(err) {
		t.Error("WHITE_SAILS.yaml was not created")
	}

	// Read and verify sails content
	sailsContent, err := os.ReadFile(sailsPath)
	if err != nil {
		t.Fatalf("Failed to read WHITE_SAILS.yaml: %v", err)
	}

	if !strings.Contains(string(sailsContent), "session_id:") {
		t.Error("WHITE_SAILS.yaml missing session_id")
	}
	if !strings.Contains(string(sailsContent), "color:") {
		t.Error("WHITE_SAILS.yaml missing color field")
	}

	// Verify events.jsonl contains SAILS_GENERATED event
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsContent, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	if !strings.Contains(string(eventsContent), "sails_generated") {
		t.Error("events.jsonl missing sails_generated event")
	}
	if !strings.Contains(string(eventsContent), "WHITE_SAILS") {
		t.Error("events.jsonl missing WHITE_SAILS reference")
	}
}

// TestWrapWithFailingTests verifies BLACK sails for failing tests.
func TestWrapWithFailingTests(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120001-def67890"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create SESSION_CONTEXT.md
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Initiative
complexity: MODULE
created_at: 2025-01-05T12:00:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Create failing test proof (exit_code != 0)
	testLog := `=== RUN   TestExample
--- FAIL: TestExample (0.00s)
    example_test.go:15: assertion failed
FAIL
FAIL	example.com/pkg	0.123s
exit code: 1`
	if err := os.WriteFile(filepath.Join(sessionDir, "test-output.log"), []byte(testLog), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Create passing build and lint logs
	buildLog := `Building example.com/pkg...
Build successful
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "build-output.log"), []byte(buildLog), 0644); err != nil {
		t.Fatalf("Failed to write build log: %v", err)
	}

	lintLog := `Linting example.com/pkg...
No issues found
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "lint-output.log"), []byte(lintLog), 0644); err != nil {
		t.Fatalf("Failed to write lint log: %v", err)
	}

	// Run wrap command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := wrapOptions{
		noArchive: true,
	}

	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("runWrap failed: %v", err)
	}

	// Verify WHITE_SAILS.yaml was created with BLACK color
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	sailsContent, err := os.ReadFile(sailsPath)
	if err != nil {
		t.Fatalf("Failed to read WHITE_SAILS.yaml: %v", err)
	}

	if !strings.Contains(string(sailsContent), "color: BLACK") {
		t.Errorf("Expected BLACK sails for failing tests, got: %s", string(sailsContent))
	}
}

// TestTransitionOutputText verifies the Text() method for TransitionOutput.
func TestTransitionOutputText(t *testing.T) {
	tests := []struct {
		name     string
		output   output.TransitionOutput
		contains []string
	}{
		{
			name: "WHITE sails",
			output: output.TransitionOutput{
				SessionID:  "session-123",
				Status:     "ARCHIVED",
				SailsColor: "WHITE",
				SailsBase:  "WHITE",
			},
			contains: []string{"Session session-123 archived", "Sails: WHITE", "Ship with confidence"},
		},
		{
			name: "GRAY sails",
			output: output.TransitionOutput{
				SessionID:  "session-456",
				Status:     "ARCHIVED",
				SailsColor: "GRAY",
				SailsBase:  "GRAY",
			},
			contains: []string{"Session session-456 archived", "Sails: GRAY", "Consider QA review", "/qa"},
		},
		{
			name: "BLACK sails",
			output: output.TransitionOutput{
				SessionID:    "session-789",
				Status:       "ARCHIVED",
				SailsColor:   "BLACK",
				SailsBase:    "BLACK",
				SailsReasons: []string{"proof 'tests' has status FAIL"},
			},
			contains: []string{"Session session-789 archived", "Sails: BLACK", "WARNING", "Do NOT ship", "proof 'tests' has status FAIL"},
		},
		{
			name: "GRAY with WHITE base (downgraded)",
			output: output.TransitionOutput{
				SessionID:  "session-abc",
				Status:     "ARCHIVED",
				SailsColor: "GRAY",
				SailsBase:  "WHITE",
			},
			contains: []string{"Sails: GRAY", "(base: WHITE)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := tt.output.Text()
			for _, s := range tt.contains {
				if !strings.Contains(text, s) {
					t.Errorf("Expected text to contain %q, got: %s", s, text)
				}
			}
		})
	}
}

// TestWrapContinuesOnSailsError verifies wrap doesn't fail if sails generation fails.
func TestWrapContinuesOnSailsError(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120002-ghi11111"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create minimal SESSION_CONTEXT.md (no proofs, should still work)
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Initiative
complexity: MODULE
created_at: 2025-01-05T12:00:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Run wrap command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := wrapOptions{
		noArchive: true,
	}

	// This should succeed even with no proofs (sails will be GRAY due to missing proofs)
	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("runWrap should not fail even without proofs: %v", err)
	}

	// Verify session was archived
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	ctx2, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}

	if ctx2.Status != session.StatusArchived {
		t.Errorf("Expected status ARCHIVED, got %s", ctx2.Status)
	}
}

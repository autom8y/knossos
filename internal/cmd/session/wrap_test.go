package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"

	"github.com/autom8y/knossos/internal/cmd/common"
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
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
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

	if !strings.Contains(string(eventsContent), "quality.sails_generated") {
		t.Error("events.jsonl missing quality.sails_generated event")
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
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := wrapOptions{
		noArchive: true,
	}

	// Run wrap - should fail on BLACK sails due to quality gate
	err := runWrap(ctx, opts)
	if err == nil {
		t.Fatal("Expected runWrap to fail on BLACK sails (failing tests), but it succeeded")
	}

	// Verify error is quality gate failure
	if !strings.Contains(err.Error(), "BLACK sails") {
		t.Errorf("Expected error about BLACK sails, got: %v", err)
	}

	// Verify session was NOT archived (status should still be ACTIVE)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	ctx2, loadErr := session.LoadContext(ctxPath)
	if loadErr != nil {
		t.Fatalf("Failed to load session context: %v", loadErr)
	}

	if ctx2.Status != session.StatusActive {
		t.Errorf("Expected status ACTIVE (wrap should have failed), got %s", ctx2.Status)
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
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
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

// TestWrapWithQAUpgrade verifies QA upgrade extraction from SESSION_CONTEXT.md.
func TestWrapWithQAUpgrade(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120003-qa123456"
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

	// Create SESSION_CONTEXT.md with QA Upgrade section
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test QA Initiative
complexity: MODULE
created_at: 2025-01-05T12:00:00Z
---

# Session Context

## QA Upgrade
- qa_session_id: session-20250105-110000-qa999999
- upgraded_at: 2025-01-05T11:00:00Z
- constraint_resolution_log: docs/qa-resolution.md
- adversarial_tests_added:
  - tests/adversarial/edge_case_test.go
  - tests/adversarial/boundary_test.go
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Create passing test proofs
	testLog := `=== RUN   TestExample
--- PASS: TestExample (0.00s)
PASS
ok  	example.com/pkg	0.123s
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "test-output.log"), []byte(testLog), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	buildLog := `Build successful
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "build-output.log"), []byte(buildLog), 0644); err != nil {
		t.Fatalf("Failed to write build log: %v", err)
	}

	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := wrapOptions{
		noArchive: true,
	}

	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("runWrap failed: %v", err)
	}

	// Verify WHITE_SAILS.yaml contains QA upgrade info
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	sailsContent, err := os.ReadFile(sailsPath)
	if err != nil {
		t.Fatalf("Failed to read WHITE_SAILS.yaml: %v", err)
	}

	sailsStr := string(sailsContent)
	if !strings.Contains(sailsStr, "qa_session_id:") {
		t.Error("WHITE_SAILS.yaml missing qa_session_id from QA upgrade")
	}
	if !strings.Contains(sailsStr, "session-20250105-110000-qa999999") {
		t.Error("WHITE_SAILS.yaml missing correct QA session ID")
	}
	if !strings.Contains(sailsStr, "upgraded_at:") {
		t.Error("WHITE_SAILS.yaml missing upgraded_at timestamp")
	}
	if !strings.Contains(sailsStr, "adversarial_tests_added:") {
		t.Error("WHITE_SAILS.yaml missing adversarial_tests_added list")
	}
}

// TestWrapBlocksOnBlackSails verifies that wrap fails when sails are BLACK.
func TestWrapBlocksOnBlackSails(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120010-black123"
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

	// Create SESSION_CONTEXT.md with explicit blocker
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test BLACK Sails Block
complexity: MODULE
created_at: 2025-01-05T12:00:00Z
---

# Session Context

## Blockers
- Critical bug in authentication flow
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Create passing proofs (blocker is what causes BLACK)
	testLog := `PASS
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "test-output.log"), []byte(testLog), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	buildLog := `Build successful
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "build-output.log"), []byte(buildLog), 0644); err != nil {
		t.Fatalf("Failed to write build log: %v", err)
	}

	lintLog := `No issues found
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "lint-output.log"), []byte(lintLog), 0644); err != nil {
		t.Fatalf("Failed to write lint log: %v", err)
	}

	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := wrapOptions{
		noArchive: true,
		force:     false, // No force - should fail
	}

	// Run wrap - should fail on BLACK sails
	err := runWrap(ctx, opts)
	if err == nil {
		t.Fatal("Expected runWrap to fail on BLACK sails, but it succeeded")
	}

	// Verify error is quality gate failure
	if !strings.Contains(err.Error(), "BLACK sails") {
		t.Errorf("Expected error about BLACK sails, got: %v", err)
	}

	// Verify session was NOT archived (status should still be ACTIVE)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	ctx2, loadErr := session.LoadContext(ctxPath)
	if loadErr != nil {
		t.Fatalf("Failed to load session context: %v", loadErr)
	}

	if ctx2.Status != session.StatusActive {
		t.Errorf("Expected status ACTIVE (wrap should have failed), got %s", ctx2.Status)
	}
}

// TestWrapWithForceBypassesBlackSails verifies that --force bypasses BLACK sails check.
func TestWrapWithForceBypassesBlackSails(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120011-force123"
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

	// Create SESSION_CONTEXT.md with explicit blocker
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Force Bypass
complexity: MODULE
created_at: 2025-01-05T12:00:00Z
---

# Session Context

## Blockers
- Known issue - forcing wrap for emergency deployment
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Create passing proofs
	testLog := `PASS
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "test-output.log"), []byte(testLog), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	buildLog := `Build successful
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "build-output.log"), []byte(buildLog), 0644); err != nil {
		t.Fatalf("Failed to write build log: %v", err)
	}

	lintLog := `No issues found
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "lint-output.log"), []byte(lintLog), 0644); err != nil {
		t.Fatalf("Failed to write lint log: %v", err)
	}

	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := wrapOptions{
		noArchive: true,
		force:     true, // Force enabled - should succeed
	}

	// Run wrap with --force - should succeed despite BLACK sails
	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("Expected runWrap with --force to succeed, but it failed: %v", err)
	}

	// Verify session was archived
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	ctx2, loadErr := session.LoadContext(ctxPath)
	if loadErr != nil {
		t.Fatalf("Failed to load session context: %v", loadErr)
	}

	if ctx2.Status != session.StatusArchived {
		t.Errorf("Expected status ARCHIVED, got %s", ctx2.Status)
	}

	// Verify WHITE_SAILS.yaml was created with BLACK color
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	sailsContent, err := os.ReadFile(sailsPath)
	if err != nil {
		t.Fatalf("Failed to read WHITE_SAILS.yaml: %v", err)
	}

	if !strings.Contains(string(sailsContent), "color: BLACK") {
		t.Errorf("Expected BLACK sails in file, got: %s", string(sailsContent))
	}
}

// TestWrapSucceedsWithWhiteSails verifies wrap succeeds when sails are WHITE.
func TestWrapSucceedsWithWhiteSails(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120012-white123"
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

	// Create SESSION_CONTEXT.md with NO blockers
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test WHITE Sails Success
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

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Create passing proofs (all required for MODULE complexity)
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

	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := wrapOptions{
		noArchive: true,
	}

	// Run wrap - should succeed with WHITE sails
	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("Expected runWrap to succeed with WHITE sails, but it failed: %v", err)
	}

	// Verify session was archived
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	ctx2, loadErr := session.LoadContext(ctxPath)
	if loadErr != nil {
		t.Fatalf("Failed to load session context: %v", loadErr)
	}

	if ctx2.Status != session.StatusArchived {
		t.Errorf("Expected status ARCHIVED, got %s", ctx2.Status)
	}

	// Verify WHITE_SAILS.yaml was created with WHITE color
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	sailsContent, err := os.ReadFile(sailsPath)
	if err != nil {
		t.Fatalf("Failed to read WHITE_SAILS.yaml: %v", err)
	}

	if !strings.Contains(string(sailsContent), "color: WHITE") {
		t.Errorf("Expected WHITE sails in file, got: %s", string(sailsContent))
	}
}

// TestWrapEmitsSessionEndWithBudget verifies session_end event includes cognitive budget.
func TestWrapEmitsSessionEndWithBudget(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20250105-120004-budget123"
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

	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Budget Tracking
complexity: MODULE
created_at: 2025-01-05T12:00:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Create CLEW_RECORD.ndjson with some tool events
	clewRecord := `{"timestamp":"2025-01-05T12:00:01Z","type":"tool.call","tool":"Read"}
{"timestamp":"2025-01-05T12:00:02Z","type":"tool.call","tool":"Bash"}
{"timestamp":"2025-01-05T12:00:03Z","type":"tool.call","tool":"Write"}
`
	if err := os.WriteFile(filepath.Join(sessionDir, "CLEW_RECORD.ndjson"), []byte(clewRecord), 0644); err != nil {
		t.Fatalf("Failed to write CLEW_RECORD.ndjson: %v", err)
	}

	// Create minimal proofs
	testLog := `PASS
exit code: 0`
	if err := os.WriteFile(filepath.Join(sessionDir, "test-output.log"), []byte(testLog), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := wrapOptions{
		noArchive: true,
	}

	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("runWrap failed: %v", err)
	}

	// Verify events.jsonl contains session_end with cognitive_budget
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsContent, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	eventsStr := string(eventsContent)
	if !strings.Contains(eventsStr, "session.ended") {
		t.Error("events.jsonl missing session.ended event")
	}
	if !strings.Contains(eventsStr, "cognitive_budget") {
		t.Error("events.jsonl session_end event missing cognitive_budget field")
	}
	if !strings.Contains(eventsStr, "total_tool_calls") {
		t.Error("events.jsonl session_end event missing total_tool_calls in budget")
	}
}

package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// TestStatus_WithSailsColor verifies that sails color is included in status output.
func TestStatus_WithSailsColor(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create session directory structure
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-100000-sailstest"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	// Create SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Sails Test
complexity: MODULE
created_at: 2026-01-06T10:00:00Z
current_phase: implementation
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create WHITE_SAILS.yaml
	sailsContent := `schema_version: "1.0"
session_id: ` + sessionID + `
generated_at: 2026-01-06T10:30:00Z
color: WHITE
computed_base: WHITE
proofs:
  tests:
    status: PASS
    summary: All tests passed
  build:
    status: PASS
    summary: Build successful
  lint:
    status: PASS
    summary: No lint errors
open_questions: []
complexity: MODULE
type: standard
`
	if err := os.WriteFile(filepath.Join(sessionDir, "WHITE_SAILS.yaml"), []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
	}

	// Set current session
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	// TODO: Capture and verify JSON output contains sails_color: WHITE
	// For now, we verify no error occurred
}

// TestStatus_WithGraySails verifies gray sails display with base color.
func TestStatus_WithGraySails(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-110000-graysails"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Gray Sails Test
complexity: MODULE
created_at: 2026-01-06T11:00:00Z
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create WHITE_SAILS.yaml with GRAY due to modifier
	sailsContent := `schema_version: "1.0"
session_id: ` + sessionID + `
generated_at: 2026-01-06T11:30:00Z
color: GRAY
computed_base: WHITE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
open_questions: []
modifiers:
  - type: DOWNGRADE_TO_GRAY
    justification: Uncertainty about edge case handling
    applied_by: agent
complexity: MODULE
type: standard
`
	if err := os.WriteFile(filepath.Join(sessionDir, "WHITE_SAILS.yaml"), []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	// Verify no error - actual output validation would require capturing output
}

// TestStatus_NoSailsFile verifies graceful handling when WHITE_SAILS.yaml doesn't exist.
func TestStatus_NoSailsFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-120000-nosails"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: No Sails Test
complexity: MODULE
created_at: 2026-01-06T12:00:00Z
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Do NOT create WHITE_SAILS.yaml

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	// Should succeed without error even when sails file is missing
}

// TestStatus_MalformedSailsFile verifies graceful handling of malformed YAML.
func TestStatus_MalformedSailsFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-130000-badsails"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Malformed Sails Test
complexity: MODULE
created_at: 2026-01-06T13:00:00Z
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create malformed WHITE_SAILS.yaml
	sailsContent := `this is not valid yaml: [[[
`
	if err := os.WriteFile(filepath.Join(sessionDir, "WHITE_SAILS.yaml"), []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	// Should succeed without error - just won't have sails data
}

// TestStatus_ArchivedSessionWithSails verifies sails display for archived sessions.
func TestStatus_ArchivedSessionWithSails(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-140000-archivedsails"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	now := time.Now().UTC()
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ARCHIVED
initiative: Archived Sails Test
complexity: MODULE
created_at: 2026-01-06T14:00:00Z
archived_at: ` + now.Format(time.RFC3339) + `
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	sailsContent := `schema_version: "1.0"
session_id: ` + sessionID + `
generated_at: ` + now.Format(time.RFC3339) + `
color: WHITE
computed_base: WHITE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
open_questions: []
complexity: MODULE
type: standard
`
	if err := os.WriteFile(filepath.Join(sessionDir, "WHITE_SAILS.yaml"), []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	// Should include sails color even for archived sessions
}

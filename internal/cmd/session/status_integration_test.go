// +build integration

package session

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/output"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// TestStatusIntegration_WithSailsColor is an integration test verifying
// that sails color appears in status output for sessions with WHITE_SAILS.yaml.
func TestStatusIntegration_WithSailsColor(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create session directory structure
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-integration-sails"
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
initiative: Integration Sails Test
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

	// Create context with captured output
	var outBuf bytes.Buffer
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

	// Override printer to capture output
	printer := output.NewPrinter(output.FormatJSON, &outBuf, os.Stderr, false)
	ctx.printer = printer

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	// Parse JSON output
	var result output.StatusOutput
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, outBuf.String())
	}

	// Verify sails color is present
	if result.SailsColor != "WHITE" {
		t.Errorf("SailsColor = %q, want %q", result.SailsColor, "WHITE")
	}

	if result.SailsBase != "WHITE" {
		t.Errorf("SailsBase = %q, want %q", result.SailsBase, "WHITE")
	}
}

// TestStatusIntegration_NoSailsFile verifies that status works when sails file is missing.
func TestStatusIntegration_NoSailsFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-integration-nosails"
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
initiative: No Sails Integration Test
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

	var outBuf bytes.Buffer
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

	printer := output.NewPrinter(output.FormatJSON, &outBuf, os.Stderr, false)
	ctx.printer = printer

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	var result output.StatusOutput
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify sails fields are empty when file doesn't exist
	if result.SailsColor != "" {
		t.Errorf("SailsColor = %q, want empty string", result.SailsColor)
	}
}

// TestStatusIntegration_TextOutput verifies text output includes sails info.
func TestStatusIntegration_TextOutput(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionID := "session-20260106-integration-text"
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
initiative: Text Output Test
complexity: MODULE
created_at: 2026-01-06T13:00:00Z
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	sailsContent := `schema_version: "1.0"
session_id: ` + sessionID + `
generated_at: 2026-01-06T13:30:00Z
color: GRAY
computed_base: WHITE
proofs:
  tests:
    status: PASS
open_questions: []
modifiers:
  - type: DOWNGRADE_TO_GRAY
    justification: Uncertainty
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

	var outBuf bytes.Buffer
	outputFormat := "text"
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

	printer := output.NewPrinter(output.FormatText, &outBuf, os.Stderr, false)
	ctx.printer = printer

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	outputStr := outBuf.String()

	// Verify text output contains sails info
	if !bytes.Contains(outBuf.Bytes(), []byte("Sails: GRAY (base: WHITE)")) {
		t.Errorf("Text output missing expected sails info.\nGot:\n%s", outputStr)
	}
}

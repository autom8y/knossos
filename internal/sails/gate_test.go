package sails

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckGate_WhiteColor(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a WHITE_SAILS.yaml with WHITE color
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-abc12345"
generated_at: "2026-01-05T15:30:00Z"
color: "WHITE"
computed_base: "WHITE"
complexity: "MODULE"
type: "standard"
proofs:
  tests:
    status: "PASS"
    summary: "47 tests passed"
    exit_code: 0
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
open_questions: []
modifiers: []
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test with directory path
	result, err := CheckGate(tmpDir)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	// Verify result
	if !result.Pass {
		t.Errorf("Expected Pass=true for WHITE color, got false")
	}
	if result.Color != ColorWhite {
		t.Errorf("Expected color=WHITE, got %s", result.Color)
	}
	if result.SessionID != "session-20260105-143000-abc12345" {
		t.Errorf("Expected session_id=session-20260105-143000-abc12345, got %s", result.SessionID)
	}
	if result.FilePath != sailsPath {
		t.Errorf("Expected file_path=%s, got %s", sailsPath, result.FilePath)
	}
	if len(result.Reasons) == 0 {
		t.Error("Expected at least one reason")
	}
}

func TestCheckGate_GrayColor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a WHITE_SAILS.yaml with GRAY color
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-def67890"
generated_at: "2026-01-05T16:00:00Z"
color: "GRAY"
computed_base: "GRAY"
complexity: "SERVICE"
type: "standard"
proofs:
  tests:
    status: "PASS"
    summary: "89 tests passed"
    exit_code: 0
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
open_questions:
  - "How should rate limiting behave under cluster failover?"
  - "Need to validate with Production DBA on index strategy"
modifiers: []
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := CheckGate(tmpDir)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	// Verify result
	if result.Pass {
		t.Errorf("Expected Pass=false for GRAY color, got true")
	}
	if result.Color != ColorGray {
		t.Errorf("Expected color=GRAY, got %s", result.Color)
	}
	if len(result.OpenQuestions) != 2 {
		t.Errorf("Expected 2 open questions, got %d", len(result.OpenQuestions))
	}
}

func TestCheckGate_BlackColor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a WHITE_SAILS.yaml with BLACK color
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-failed01"
generated_at: "2026-01-05T17:00:00Z"
color: "BLACK"
computed_base: "BLACK"
complexity: "MODULE"
type: "standard"
proofs:
  tests:
    status: "FAIL"
    summary: "10 tests failed"
    exit_code: 1
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
open_questions: []
modifiers: []
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := CheckGate(tmpDir)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	// Verify result
	if result.Pass {
		t.Errorf("Expected Pass=false for BLACK color, got true")
	}
	if result.Color != ColorBlack {
		t.Errorf("Expected color=BLACK, got %s", result.Color)
	}
}

func TestCheckGate_SpikeSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a WHITE_SAILS.yaml for a spike session (always GRAY)
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-spike001"
generated_at: "2026-01-05T17:00:00Z"
color: "GRAY"
computed_base: "GRAY"
complexity: "MODULE"
type: "spike"
proofs:
  tests:
    status: "PASS"
    summary: "tests passed"
    exit_code: 0
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
open_questions: []
modifiers: []
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := CheckGate(tmpDir)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	if result.Pass {
		t.Errorf("Expected Pass=false for spike session, got true")
	}
	if result.Color != ColorGray {
		t.Errorf("Expected color=GRAY for spike, got %s", result.Color)
	}

	// Check that spike is mentioned in reasons
	hasSpkeReason := false
	for _, reason := range result.Reasons {
		if reason == "session type is spike (gray ceiling)" {
			hasSpkeReason = true
			break
		}
	}
	if !hasSpkeReason {
		t.Errorf("Expected spike reason in reasons, got: %v", result.Reasons)
	}
}

func TestCheckGate_HotfixSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a WHITE_SAILS.yaml for a hotfix session (always GRAY)
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-hotfix01"
generated_at: "2026-01-05T17:00:00Z"
color: "GRAY"
computed_base: "GRAY"
complexity: "PATCH"
type: "hotfix"
proofs:
  tests:
    status: "PASS"
    summary: "tests passed"
    exit_code: 0
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
open_questions: []
modifiers: []
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := CheckGate(tmpDir)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	if result.Pass {
		t.Errorf("Expected Pass=false for hotfix session, got true")
	}

	// Check that hotfix is mentioned in reasons
	hasHotfixReason := false
	for _, reason := range result.Reasons {
		if reason == "session type is hotfix (expedited gray)" {
			hasHotfixReason = true
			break
		}
	}
	if !hasHotfixReason {
		t.Errorf("Expected hotfix reason in reasons, got: %v", result.Reasons)
	}
}

func TestCheckGate_DirectFilePath(t *testing.T) {
	tmpDir := t.TempDir()

	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-abc12345"
generated_at: "2026-01-05T15:30:00Z"
color: "WHITE"
computed_base: "WHITE"
proofs:
  tests:
    status: "PASS"
  build:
    status: "PASS"
  lint:
    status: "PASS"
open_questions: []
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test with direct file path
	result, err := CheckGate(sailsPath)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	if !result.Pass {
		t.Errorf("Expected Pass=true, got false")
	}
}

func TestCheckGate_MissingSailsFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create WHITE_SAILS.yaml
	_, err := CheckGate(tmpDir)
	if err == nil {
		t.Error("Expected error for missing WHITE_SAILS.yaml")
	}
}

func TestCheckGate_InvalidPath(t *testing.T) {
	_, err := CheckGate("/nonexistent/path/to/session")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestCheckGate_EmptyPath(t *testing.T) {
	_, err := CheckGate("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestCheckGate_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid YAML
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := CheckGate(tmpDir)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestGateExitCode_Pass(t *testing.T) {
	result := &GateResult{
		Pass:  true,
		Color: ColorWhite,
	}

	exitCode := GateExitCode(result)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for passing gate, got %d", exitCode)
	}
}

func TestGateExitCode_Fail(t *testing.T) {
	result := &GateResult{
		Pass:  false,
		Color: ColorGray,
	}

	exitCode := GateExitCode(result)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for failing gate")
	}
}

func TestGateExitCode_NilResult(t *testing.T) {
	exitCode := GateExitCode(nil)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for nil result")
	}
}

func TestCheckGate_QAUpgradedSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a WHITE_SAILS.yaml that was upgraded by QA
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-upgraded"
generated_at: "2026-01-05T18:00:00Z"
color: "WHITE"
computed_base: "GRAY"
complexity: "SERVICE"
type: "standard"
proofs:
  tests:
    status: "PASS"
    summary: "all tests pass"
    exit_code: 0
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
open_questions: []
modifiers: []
qa_upgrade:
  upgraded_at: "2026-01-06T12:00:00Z"
  qa_session_id: "session-20260106-100000-qa123456"
  constraint_resolution_log: ".ledge/specs/TP-qa-original-session.md"
  adversarial_tests_added:
    - "tests/integration/rate_limit_failover_test.go"
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := CheckGate(tmpDir)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	// Should pass because final color is WHITE (upgraded by QA)
	if !result.Pass {
		t.Errorf("Expected Pass=true for QA-upgraded session, got false")
	}
	if result.Color != ColorWhite {
		t.Errorf("Expected color=WHITE, got %s", result.Color)
	}
	// Computed base was GRAY before QA upgrade
	if result.ComputedBase != ColorGray {
		t.Errorf("Expected computed_base=GRAY, got %s", result.ComputedBase)
	}
}

func TestCheckGate_WithModifiers(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a WHITE_SAILS.yaml with a downgrade modifier
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-downgraded"
generated_at: "2026-01-05T17:00:00Z"
color: "GRAY"
computed_base: "WHITE"
complexity: "PATCH"
type: "standard"
proofs:
  tests:
    status: "PASS"
    summary: "all tests pass"
    exit_code: 0
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
open_questions: []
modifiers:
  - type: "DOWNGRADE_TO_GRAY"
    justification: "Changed payment flow; want senior review despite passing tests"
    applied_by: "human"
`
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
	if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := CheckGate(tmpDir)
	if err != nil {
		t.Fatalf("CheckGate failed: %v", err)
	}

	// Should fail because final color is GRAY (downgraded)
	if result.Pass {
		t.Errorf("Expected Pass=false for downgraded session, got true")
	}
	if result.Color != ColorGray {
		t.Errorf("Expected color=GRAY, got %s", result.Color)
	}
	// Computed base was WHITE before modifier
	if result.ComputedBase != ColorWhite {
		t.Errorf("Expected computed_base=WHITE, got %s", result.ComputedBase)
	}
}

func TestTrimWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"  hello  ", "hello"},
		{"\n\thello\n\t", "hello"},
		{"", ""},
		{"   ", ""},
		{"\n\n", ""},
		{"hello world", "hello world"},
		{"  hello world  ", "hello world"},
	}

	for _, tc := range tests {
		result := trimWhitespace(tc.input)
		if result != tc.expected {
			t.Errorf("trimWhitespace(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

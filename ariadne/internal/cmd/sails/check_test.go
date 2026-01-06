// Package sails integration tests for the ari sails commands.
// Tests cover check command, confidence level computation, and session wrap integration.
package sails

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/ariadne/internal/sails"
)

// =============================================================================
// Test Helpers
// =============================================================================

// createTestProject creates a temporary project structure with sessions directory.
func createTestProject(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0755))
	return tmpDir
}

// createTestSession creates a session directory with the given ID.
func createTestSession(t *testing.T, projectDir, sessionID string) string {
	t.Helper()
	sessionDir := filepath.Join(projectDir, ".claude", "sessions", sessionID)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	return sessionDir
}

// setCurrentSession sets the current session for the project.
func setCurrentSession(t *testing.T, projectDir, sessionID string) {
	t.Helper()
	currentPath := filepath.Join(projectDir, ".claude", "sessions", ".current-session")
	require.NoError(t, os.WriteFile(currentPath, []byte(sessionID), 0644))
}

// writeWhiteSails writes a WHITE_SAILS.yaml file to the session directory.
func writeWhiteSails(t *testing.T, sessionDir string, content string) {
	t.Helper()
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	require.NoError(t, os.WriteFile(sailsPath, []byte(content), 0644))
}

// =============================================================================
// Check Command Tests
// =============================================================================

// TestCheckCmd_WhiteColorPasses verifies check command passes for WHITE sails.
func TestCheckCmd_WhiteColorPasses(t *testing.T) {
	projectDir := createTestProject(t)
	sessionID := "session-20260105-143000-white123"
	sessionDir := createTestSession(t, projectDir, sessionID)

	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-white123"
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
	writeWhiteSails(t, sessionDir, sailsContent)

	// Test CheckGate directly (since command calls os.Exit)
	result, err := sails.CheckGate(sessionDir)
	require.NoError(t, err)

	assert.True(t, result.Pass, "WHITE sails should pass gate")
	assert.Equal(t, sails.ColorWhite, result.Color)
	assert.Equal(t, sessionID, result.SessionID)
	assert.Contains(t, result.FilePath, "WHITE_SAILS.yaml")

	// Verify exit code
	exitCode := sails.GateExitCode(result)
	assert.Equal(t, 0, exitCode, "WHITE should have exit code 0")
}

// TestCheckCmd_GrayColorFails verifies check command fails for GRAY sails.
func TestCheckCmd_GrayColorFails(t *testing.T) {
	projectDir := createTestProject(t)
	sessionID := "session-20260105-143000-gray123"
	sessionDir := createTestSession(t, projectDir, sessionID)

	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-gray123"
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
	writeWhiteSails(t, sessionDir, sailsContent)

	result, err := sails.CheckGate(sessionDir)
	require.NoError(t, err)

	assert.False(t, result.Pass, "GRAY sails should fail gate")
	assert.Equal(t, sails.ColorGray, result.Color)
	assert.Len(t, result.OpenQuestions, 2)

	// Verify non-zero exit code
	exitCode := sails.GateExitCode(result)
	assert.NotEqual(t, 0, exitCode, "GRAY should have non-zero exit code")
}

// TestCheckCmd_BlackColorFails verifies check command fails for BLACK sails.
func TestCheckCmd_BlackColorFails(t *testing.T) {
	projectDir := createTestProject(t)
	sessionID := "session-20260105-143000-black123"
	sessionDir := createTestSession(t, projectDir, sessionID)

	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-black123"
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
	writeWhiteSails(t, sessionDir, sailsContent)

	result, err := sails.CheckGate(sessionDir)
	require.NoError(t, err)

	assert.False(t, result.Pass, "BLACK sails should fail gate")
	assert.Equal(t, sails.ColorBlack, result.Color)

	// Verify non-zero exit code
	exitCode := sails.GateExitCode(result)
	assert.NotEqual(t, 0, exitCode, "BLACK should have non-zero exit code")
}

// TestCheckCmd_DirectFilePath verifies check works with direct file path.
func TestCheckCmd_DirectFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")

	sailsContent := `schema_version: "1.0"
session_id: "session-test-direct"
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
	require.NoError(t, os.WriteFile(sailsPath, []byte(sailsContent), 0644))

	// Check with direct file path
	result, err := sails.CheckGate(sailsPath)
	require.NoError(t, err)
	assert.True(t, result.Pass)
	assert.Equal(t, sails.ColorWhite, result.Color)
}

// TestCheckCmd_CurrentSession verifies check works for current session.
func TestCheckCmd_CurrentSession(t *testing.T) {
	projectDir := createTestProject(t)
	sessionID := "session-20260105-143000-current"
	sessionDir := createTestSession(t, projectDir, sessionID)
	setCurrentSession(t, projectDir, sessionID)

	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-current"
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
	writeWhiteSails(t, sessionDir, sailsContent)

	// Check current session via project root
	result, err := sails.CheckGateForCurrentSession(projectDir)
	require.NoError(t, err)
	assert.True(t, result.Pass)
	assert.Equal(t, sessionID, result.SessionID)
}

// TestCheckCmd_ErrorCases verifies error handling.
func TestCheckCmd_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		errContains string
	}{
		{
			name: "empty path",
			setup: func(t *testing.T) string {
				return ""
			},
			errContains: "required",
		},
		{
			name: "nonexistent path",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/to/session"
			},
			errContains: "not found",
		},
		{
			name: "missing sails file in directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			errContains: "WHITE_SAILS.yaml not found",
		},
		{
			name: "invalid YAML content",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				sailsPath := filepath.Join(tmpDir, "WHITE_SAILS.yaml")
				require.NoError(t, os.WriteFile(sailsPath, []byte("invalid: yaml: content:"), 0644))
				return tmpDir
			},
			errContains: "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			_, err := sails.CheckGate(path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// =============================================================================
// Confidence Level Computation Tests
// =============================================================================

// TestConfidenceLevel_WHITE_AllProofsPassing verifies WHITE for all passing proofs.
func TestConfidenceLevel_WHITE_AllProofsPassing(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofPass},
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: nil,
	}

	result := sails.ComputeColor(input)

	assert.Equal(t, sails.ColorWhite, result.Color)
	assert.Equal(t, sails.ColorWhite, result.ComputedBase)
	assert.Contains(t, result.Reasons, "all required proofs present and passing")
}

// TestConfidenceLevel_GRAY_OpenQuestions verifies GRAY ceiling for open questions.
func TestConfidenceLevel_GRAY_OpenQuestions(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofPass},
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: []string{"What about edge case X?"},
	}

	result := sails.ComputeColor(input)

	assert.Equal(t, sails.ColorGray, result.Color)
	assert.Equal(t, sails.ColorGray, result.ComputedBase)
	assert.Contains(t, result.Reasons, "open questions present: gray ceiling applied")
}

// TestConfidenceLevel_GRAY_MissingProofs verifies GRAY for missing required proofs.
func TestConfidenceLevel_GRAY_MissingProofs(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			// Missing build and lint
		},
		OpenQuestions: nil,
	}

	result := sails.ComputeColor(input)

	assert.Equal(t, sails.ColorGray, result.Color)
	assert.Equal(t, sails.ColorGray, result.ComputedBase)
	// Should mention missing proof
	hasReason := false
	for _, r := range result.Reasons {
		if strings.Contains(r, "missing") || strings.Contains(r, "required") {
			hasReason = true
			break
		}
	}
	assert.True(t, hasReason, "Expected reason mentioning missing proof")
}

// TestConfidenceLevel_BLACK_FailingTests verifies BLACK for failing tests.
func TestConfidenceLevel_BLACK_FailingTests(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofFail},
			"build": {Status: sails.ProofPass},
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: nil,
	}

	result := sails.ComputeColor(input)

	assert.Equal(t, sails.ColorBlack, result.Color)
	assert.Equal(t, sails.ColorBlack, result.ComputedBase)
	assert.Contains(t, result.Reasons, "proof 'tests' has status FAIL")
}

// TestConfidenceLevel_GRAY_SpikeCeiling verifies spike sessions have GRAY ceiling.
func TestConfidenceLevel_GRAY_SpikeCeiling(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "spike",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofPass},
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: nil,
	}

	result := sails.ComputeColor(input)

	assert.Equal(t, sails.ColorGray, result.Color)
	assert.Equal(t, sails.ColorGray, result.ComputedBase)
	assert.Contains(t, result.Reasons, "session type 'spike' has gray ceiling (spikes never white)")
}

// TestConfidenceLevel_GRAY_HotfixCeiling verifies hotfix sessions have GRAY ceiling.
func TestConfidenceLevel_GRAY_HotfixCeiling(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "hotfix",
		Complexity:  "PATCH",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofPass},
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: nil,
	}

	result := sails.ComputeColor(input)

	assert.Equal(t, sails.ColorGray, result.Color)
	assert.Equal(t, sails.ColorGray, result.ComputedBase)
	assert.Contains(t, result.Reasons, "session type 'hotfix' has gray ceiling (expedited gray)")
}

// TestConfidenceLevel_ModifierDowngrade verifies modifiers can downgrade color.
func TestConfidenceLevel_ModifierDowngrade(t *testing.T) {
	tests := []struct {
		name         string
		modifier     sails.ModifierType
		expectedFinal sails.Color
		expectedBase  sails.Color
	}{
		{
			name:         "DOWNGRADE_TO_GRAY",
			modifier:     sails.ModifierDowngradeToGray,
			expectedFinal: sails.ColorGray,
			expectedBase:  sails.ColorWhite,
		},
		{
			name:         "DOWNGRADE_TO_BLACK",
			modifier:     sails.ModifierDowngradeToBlack,
			expectedFinal: sails.ColorBlack,
			expectedBase:  sails.ColorWhite,
		},
		{
			name:         "HUMAN_OVERRIDE_GRAY",
			modifier:     sails.ModifierHumanOverrideGray,
			expectedFinal: sails.ColorGray,
			expectedBase:  sails.ColorWhite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := sails.ColorInput{
				SessionType: "standard",
				Complexity:  "MODULE",
				Proofs: map[string]sails.ColorProof{
					"tests": {Status: sails.ProofPass},
					"build": {Status: sails.ProofPass},
					"lint":  {Status: sails.ProofPass},
				},
				OpenQuestions: nil,
				Modifiers: []sails.Modifier{
					{
						Type:          tt.modifier,
						Justification: "Test justification",
						AppliedBy:     "human",
					},
				},
			}

			result := sails.ComputeColor(input)

			assert.Equal(t, tt.expectedFinal, result.Color)
			assert.Equal(t, tt.expectedBase, result.ComputedBase)
		})
	}
}

// TestConfidenceLevel_QAUpgrade verifies QA upgrade from GRAY to WHITE.
func TestConfidenceLevel_QAUpgrade(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofPass},
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: []string{"Originally unresolved question"},
		QAUpgrade: &sails.QAUpgrade{
			QASessionID:             "session-qa-upgrade",
			ConstraintResolutionLog: "docs/qa-resolution.md",
			AdversarialTestsAdded:   []string{"tests/edge_case_test.go"},
		},
	}

	result := sails.ComputeColor(input)

	assert.Equal(t, sails.ColorWhite, result.Color, "QA upgrade should upgrade GRAY to WHITE")
	assert.Equal(t, sails.ColorGray, result.ComputedBase, "Computed base should remain GRAY")
	assert.Contains(t, result.Reasons, "QA upgrade applied: gray -> white via QA session session-qa-upgrade")
}

// TestConfidenceLevel_QAUpgradeRequirements verifies QA upgrade conditions.
func TestConfidenceLevel_QAUpgradeRequirements(t *testing.T) {
	tests := []struct {
		name          string
		qaUpgrade     *sails.QAUpgrade
		expectedColor sails.Color
		expectedReason string
	}{
		{
			name: "missing constraint log",
			qaUpgrade: &sails.QAUpgrade{
				QASessionID:           "qa-session",
				AdversarialTestsAdded: []string{"test.go"},
				// Missing ConstraintResolutionLog
			},
			expectedColor:  sails.ColorGray,
			expectedReason: "QA upgrade missing constraint_resolution_log: cannot upgrade",
		},
		{
			name: "missing adversarial tests",
			qaUpgrade: &sails.QAUpgrade{
				QASessionID:             "qa-session",
				ConstraintResolutionLog: "docs/log.md",
				// Missing AdversarialTestsAdded
			},
			expectedColor:  sails.ColorGray,
			expectedReason: "QA upgrade has no adversarial_tests_added: cannot upgrade",
		},
		{
			name:          "nil QA upgrade",
			qaUpgrade:     nil,
			expectedColor: sails.ColorGray,
			// No specific reason for nil upgrade
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := sails.ColorInput{
				SessionType: "standard",
				Complexity:  "MODULE",
				Proofs: map[string]sails.ColorProof{
					"tests": {Status: sails.ProofPass},
					"build": {Status: sails.ProofPass},
					"lint":  {Status: sails.ProofPass},
				},
				OpenQuestions: []string{"Question causing GRAY"},
				QAUpgrade:     tt.qaUpgrade,
			}

			result := sails.ComputeColor(input)

			assert.Equal(t, tt.expectedColor, result.Color)
			if tt.expectedReason != "" {
				assert.Contains(t, result.Reasons, tt.expectedReason)
			}
		})
	}
}

// =============================================================================
// Complexity Threshold Tests
// =============================================================================

// TestComplexityThresholds_RequiredProofs verifies proof requirements by complexity.
func TestComplexityThresholds_RequiredProofs(t *testing.T) {
	tests := []struct {
		complexity string
		required   []string
		notRequired []string
	}{
		{
			complexity:  "PATCH",
			required:    []string{"tests", "build", "lint"},
			notRequired: []string{"adversarial", "integration"},
		},
		{
			complexity:  "MODULE",
			required:    []string{"tests", "build", "lint"},
			notRequired: []string{"adversarial", "integration"},
		},
		{
			complexity:  "SERVICE",
			required:    []string{"tests", "build", "lint"},
			notRequired: []string{"adversarial", "integration"}, // recommended, not required
		},
		{
			complexity:  "INITIATIVE",
			required:    []string{"tests", "build", "lint", "adversarial", "integration"},
			notRequired: []string{},
		},
		{
			complexity:  "MIGRATION",
			required:    []string{"tests", "build", "lint", "adversarial", "integration"},
			notRequired: []string{},
		},
		{
			complexity:  "PLATFORM",
			required:    []string{"tests", "build", "lint", "adversarial", "integration"},
			notRequired: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.complexity, func(t *testing.T) {
			requiredProofs := sails.GetRequiredProofs(tt.complexity)

			for _, proof := range tt.required {
				assert.Contains(t, requiredProofs, proof,
					"Complexity %s should require %s", tt.complexity, proof)
			}

			for _, proof := range tt.notRequired {
				assert.NotContains(t, requiredProofs, proof,
					"Complexity %s should not require %s", tt.complexity, proof)
			}
		})
	}
}

// TestComplexityThresholds_INITIATIVERequiresAllProofs verifies INITIATIVE complexity.
func TestComplexityThresholds_INITIATIVERequiresAllProofs(t *testing.T) {
	// INITIATIVE with only basic proofs should be GRAY
	inputMissing := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "INITIATIVE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofPass},
			"lint":  {Status: sails.ProofPass},
			// Missing adversarial and integration
		},
		OpenQuestions: nil,
	}

	resultMissing := sails.ComputeColor(inputMissing)
	assert.Equal(t, sails.ColorGray, resultMissing.Color, "INITIATIVE missing proofs should be GRAY")

	// INITIATIVE with all proofs should be WHITE
	inputComplete := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "INITIATIVE",
		Proofs: map[string]sails.ColorProof{
			"tests":       {Status: sails.ProofPass},
			"build":       {Status: sails.ProofPass},
			"lint":        {Status: sails.ProofPass},
			"adversarial": {Status: sails.ProofPass},
			"integration": {Status: sails.ProofPass},
		},
		OpenQuestions: nil,
	}

	resultComplete := sails.ComputeColor(inputComplete)
	assert.Equal(t, sails.ColorWhite, resultComplete.Color, "INITIATIVE with all proofs should be WHITE")
}

// =============================================================================
// Gate Output Format Tests
// =============================================================================

// TestGateOutput_Formatting verifies gateOutput String() method.
func TestGateOutput_Formatting(t *testing.T) {
	tests := []struct {
		name     string
		output   gateOutput
		contains []string
	}{
		{
			name: "WHITE pass",
			output: gateOutput{
				Pass:      true,
				Color:     "WHITE",
				SessionID: "session-123",
				FilePath:  "/path/to/WHITE_SAILS.yaml",
				Reasons:   []string{"all proofs passing"},
				Summary:   "WHITE sails: high confidence",
			},
			contains: []string{"PASS", "WHITE", "session-123"},
		},
		{
			name: "GRAY fail with open questions",
			output: gateOutput{
				Pass:          false,
				Color:         "GRAY",
				SessionID:     "session-456",
				FilePath:      "/path/to/WHITE_SAILS.yaml",
				Reasons:       []string{"open questions present"},
				OpenQuestions: []string{"Question 1?", "Question 2?"},
				Summary:       "GRAY sails: needs QA",
			},
			contains: []string{"FAIL", "GRAY", "Open Questions", "Question 1"},
		},
		{
			name: "GRAY with WHITE base (downgraded)",
			output: gateOutput{
				Pass:         false,
				Color:        "GRAY",
				ComputedBase: "WHITE",
				SessionID:    "session-789",
				FilePath:     "/path/to/WHITE_SAILS.yaml",
				Reasons:      []string{"modifier applied"},
			},
			contains: []string{"FAIL", "GRAY", "WHITE", "Computed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := tt.output.String()
			for _, s := range tt.contains {
				assert.Contains(t, text, s, "Output should contain %q", s)
			}
		})
	}
}

// TestBuildSummary verifies summary generation for gate results.
func TestBuildSummary(t *testing.T) {
	tests := []struct {
		name     string
		result   *sails.GateResult
		expected string
	}{
		{
			name:     "WHITE pass",
			result:   &sails.GateResult{Pass: true, Color: sails.ColorWhite},
			expected: "WHITE sails: high confidence, ship without QA",
		},
		{
			name:     "GRAY with open questions",
			result:   &sails.GateResult{Pass: false, Color: sails.ColorGray, OpenQuestions: []string{"Q1", "Q2"}},
			expected: "GRAY sails: 2 open question(s), needs QA review",
		},
		{
			name:     "GRAY without open questions",
			result:   &sails.GateResult{Pass: false, Color: sails.ColorGray},
			expected: "GRAY sails: unknown confidence, needs QA review",
		},
		{
			name:     "BLACK failure",
			result:   &sails.GateResult{Pass: false, Color: sails.ColorBlack},
			expected: "BLACK sails: known failure, do not ship",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := buildSummary(tt.result)
			assert.Equal(t, tt.expected, summary)
		})
	}
}

// =============================================================================
// Check Command Context Tests
// =============================================================================

// TestCmdContext_GetPrinter verifies printer creation from context.
func TestCmdContext_GetPrinter(t *testing.T) {
	outputFormat := "json"
	verbose := true

	ctx := &cmdContext{
		output:  &outputFormat,
		verbose: &verbose,
	}

	printer := ctx.getPrinter()
	assert.NotNil(t, printer)
}

// TestCmdContext_NilValues handles nil context values gracefully.
func TestCmdContext_NilValues(t *testing.T) {
	ctx := &cmdContext{
		output:  nil,
		verbose: nil,
	}

	// Should not panic
	printer := ctx.getPrinter()
	assert.NotNil(t, printer)
}

// =============================================================================
// Integration: Check with Session Wrap (C4 Coordination)
// =============================================================================

// TestCheckAfterWrap_Integration verifies check works on wrapped session sails.
// This tests the C4 sails/moirai coordination pattern.
func TestCheckAfterWrap_Integration(t *testing.T) {
	// Simulate a session that was wrapped with WHITE_SAILS.yaml generated
	projectDir := createTestProject(t)
	sessionID := "session-20260105-143000-wrapped"
	sessionDir := createTestSession(t, projectDir, sessionID)

	// Create SESSION_CONTEXT.md in ARCHIVED state (post-wrap)
	contextContent := `---
schema_version: "1.0"
session_id: ` + sessionID + `
status: ARCHIVED
initiative: Test Initiative
complexity: MODULE
created_at: 2026-01-05T12:00:00Z
archived_at: 2026-01-05T14:30:00Z
---

# Session Context

## Session Type
standard

## Open Questions
None.
`
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "SESSION_CONTEXT.md"),
		[]byte(contextContent),
		0644,
	))

	// Create WHITE_SAILS.yaml as would be generated by wrap command
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-wrapped"
generated_at: "2026-01-05T14:30:00Z"
color: "WHITE"
computed_base: "WHITE"
complexity: "MODULE"
type: "standard"
proofs:
  tests:
    status: "PASS"
    summary: "47 tests passed"
    exit_code: 0
    evidence_path: "test-output.log"
    timestamp: "2026-01-05T14:29:00Z"
  build:
    status: "PASS"
    summary: "build succeeded"
    exit_code: 0
    evidence_path: "build-output.log"
    timestamp: "2026-01-05T14:29:30Z"
  lint:
    status: "PASS"
    summary: "lint clean"
    exit_code: 0
    evidence_path: "lint-output.log"
    timestamp: "2026-01-05T14:29:45Z"
open_questions: []
modifiers: []
`
	writeWhiteSails(t, sessionDir, sailsContent)

	// Verify check command can read the wrapped session's sails
	result, err := sails.CheckGate(sessionDir)
	require.NoError(t, err)

	assert.True(t, result.Pass, "Wrapped session with WHITE sails should pass")
	assert.Equal(t, sails.ColorWhite, result.Color)
	assert.Equal(t, sessionID, result.SessionID)

	// Verify proofs are populated
	assert.NotEmpty(t, result.FilePath)
}

// TestCheckWithEvents_Integration verifies check works with event trail.
// This validates C4 coordination where sails_generated event is emitted.
func TestCheckWithEvents_Integration(t *testing.T) {
	projectDir := createTestProject(t)
	sessionID := "session-20260105-143000-events"
	sessionDir := createTestSession(t, projectDir, sessionID)

	// Create WHITE_SAILS.yaml
	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-events"
generated_at: "2026-01-05T14:30:00Z"
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
	writeWhiteSails(t, sessionDir, sailsContent)

	// Create events.jsonl with sails_generated event (as wrap command would)
	eventsContent := `{"timestamp":"2026-01-05T14:30:00Z","type":"sails_generated","session_id":"session-20260105-143000-events","data":{"color":"WHITE","computed_base":"WHITE","file_path":"WHITE_SAILS.yaml"}}
{"timestamp":"2026-01-05T14:30:01Z","type":"session_end","session_id":"session-20260105-143000-events","data":{"reason":"completed"}}
`
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "events.jsonl"),
		[]byte(eventsContent),
		0644,
	))

	// Check gate still works regardless of events
	result, err := sails.CheckGate(sessionDir)
	require.NoError(t, err)
	assert.True(t, result.Pass)
	assert.Equal(t, sails.ColorWhite, result.Color)
}

// =============================================================================
// Printer Output Tests (for formatGateResult)
// =============================================================================

// TestFormatGateResult verifies gate result formatting.
func TestFormatGateResult(t *testing.T) {
	result := &sails.GateResult{
		Pass:          true,
		Color:         sails.ColorWhite,
		SessionID:     "test-session",
		Reasons:       []string{"all proofs passing"},
		FilePath:      "/path/to/WHITE_SAILS.yaml",
		ComputedBase:  sails.ColorWhite,
		OpenQuestions: nil,
	}

	output := formatGateResult(result)
	gateOut, ok := output.(*gateOutput)
	require.True(t, ok, "Expected *gateOutput type")

	assert.True(t, gateOut.Pass)
	assert.Equal(t, "WHITE", gateOut.Color)
	assert.Equal(t, "test-session", gateOut.SessionID)
	assert.Contains(t, gateOut.Reasons, "all proofs passing")
	assert.Equal(t, "WHITE sails: high confidence, ship without QA", gateOut.Summary)
}

// =============================================================================
// Edge Cases and Error Recovery
// =============================================================================

// TestCheckCmd_WhitespaceInSessionID verifies handling of whitespace.
func TestCheckCmd_WhitespaceInSessionID(t *testing.T) {
	projectDir := createTestProject(t)
	sessionID := "session-20260105-143000-whitespace"
	sessionDir := createTestSession(t, projectDir, sessionID)

	// Set current session with trailing whitespace
	currentPath := filepath.Join(projectDir, ".claude", "sessions", ".current-session")
	require.NoError(t, os.WriteFile(currentPath, []byte(sessionID+"\n  "), 0644))

	sailsContent := `schema_version: "1.0"
session_id: "session-20260105-143000-whitespace"
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
	writeWhiteSails(t, sessionDir, sailsContent)

	// Should handle whitespace in session ID
	result, err := sails.CheckGateForCurrentSession(projectDir)
	require.NoError(t, err)
	assert.True(t, result.Pass)
}

// TestCheckCmd_ProofSkipStatus verifies SKIP status is treated as passing.
func TestCheckCmd_ProofSkipStatus(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofSkip}, // Intentionally skipped
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: nil,
	}

	result := sails.ComputeColor(input)

	// SKIP should be treated as passing
	assert.Equal(t, sails.ColorWhite, result.Color, "SKIP should be treated as passing")
}

// TestCheckCmd_UnknownProofStatus verifies UNKNOWN status handling.
func TestCheckCmd_UnknownProofStatus(t *testing.T) {
	input := sails.ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]sails.ColorProof{
			"tests": {Status: sails.ProofPass},
			"build": {Status: sails.ProofUnknown}, // Unknown - missing proof
			"lint":  {Status: sails.ProofPass},
		},
		OpenQuestions: nil,
	}

	result := sails.ComputeColor(input)

	// UNKNOWN should cause GRAY (not passing)
	assert.Equal(t, sails.ColorGray, result.Color, "UNKNOWN status should cause GRAY")
}

// =============================================================================
// Concurrency Safety Test (simulated via buffer)
// =============================================================================

// TestCmdContext_ConcurrentPrinterAccess verifies printer is safe for concurrent use.
func TestCmdContext_ConcurrentPrinterAccess(t *testing.T) {
	outputFormat := "json"
	verbose := true

	ctx := &cmdContext{
		output:  &outputFormat,
		verbose: &verbose,
	}

	// Create multiple printers concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			printer := ctx.getPrinter()
			assert.NotNil(t, printer)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

// TestValidateColorInput verifies color input validation.
func TestValidateColorInput(t *testing.T) {
	tests := []struct {
		name       string
		input      sails.ColorInput
		hasErrors  bool
	}{
		{
			name: "valid input",
			input: sails.ColorInput{
				SessionType: "standard",
				Complexity:  "MODULE",
				Proofs: map[string]sails.ColorProof{
					"tests": {Status: sails.ProofPass},
				},
			},
			hasErrors: false,
		},
		{
			name: "invalid proof status",
			input: sails.ColorInput{
				SessionType: "standard",
				Complexity:  "MODULE",
				Proofs: map[string]sails.ColorProof{
					"tests": {Status: "INVALID"},
				},
			},
			hasErrors: true,
		},
		{
			name: "modifier missing justification",
			input: sails.ColorInput{
				SessionType: "standard",
				Complexity:  "MODULE",
				Proofs: map[string]sails.ColorProof{
					"tests": {Status: sails.ProofPass},
				},
				Modifiers: []sails.Modifier{
					{
						Type:      sails.ModifierDowngradeToGray,
						AppliedBy: "agent",
						// Missing Justification
					},
				},
			},
			hasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := sails.ValidateColorInput(tt.input)
			if tt.hasErrors {
				assert.NotEmpty(t, errors, "Expected validation errors")
			} else {
				assert.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

// Helper to capture stdout/stderr for testing (not used but available)
func captureOutput(f func()) (string, string) {
	var stdout, stderr bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	f()

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout.ReadFrom(rOut)
	stderr.ReadFrom(rErr)

	return stdout.String(), stderr.String()
}

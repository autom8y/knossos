// Package sails integration tests per TDD Section 11.2.
// These tests verify the full sails generation flow with real Generator implementation.
package sails

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Test Helpers
// =============================================================================

// writeProofLog creates a proof log file in the session directory.
func writeProofLog(t *testing.T, sessionDir, filename, content string) {
	t.Helper()
	path := filepath.Join(sessionDir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "failed to write proof log %s", filename)
}

// SessionContextInput provides structured input for session context creation.
type SessionContextInput struct {
	SessionID     string
	Complexity    string
	Type          string // "standard", "spike", "hotfix"
	OpenQuestions []string
	Modifiers     []ModifierInput
}

// ModifierInput represents a modifier for test setup.
type ModifierInput struct {
	Type          string // "DOWNGRADE_TO_GRAY", "DOWNGRADE_TO_BLACK", "HUMAN_OVERRIDE_GRAY"
	Justification string
	AppliedBy     string // "agent" or "human"
}

// writeSessionContext creates a SESSION_CONTEXT.md file with the given configuration.
func writeSessionContext(t *testing.T, sessionDir string, ctx SessionContextInput) {
	t.Helper()

	// Set defaults
	if ctx.SessionID == "" {
		ctx.SessionID = filepath.Base(sessionDir)
	}
	if ctx.Complexity == "" {
		ctx.Complexity = "MODULE"
	}
	if ctx.Type == "" {
		ctx.Type = "standard"
	}

	// Build frontmatter
	content := `---
schema_version: "2.1"
session_id: "` + ctx.SessionID + `"
status: ACTIVE
created_at: "2026-01-05T12:00:00Z"
initiative: "Test Initiative"
complexity: ` + ctx.Complexity + `
active_rite: 10x-dev-pack
current_phase: implementation
---

# Session: Test Initiative

`

	// Add session type section if not standard
	if ctx.Type != "standard" {
		content += `## Session Type
` + ctx.Type + `

`
	}

	// Add open questions section
	content += `## Open Questions
`
	if len(ctx.OpenQuestions) == 0 {
		content += "None.\n\n"
	} else {
		for _, q := range ctx.OpenQuestions {
			content += "- " + q + "\n"
		}
		content += "\n"
	}

	// Add modifiers section
	content += `## Modifiers
`
	if len(ctx.Modifiers) == 0 {
		content += "None.\n\n"
	} else {
		for _, m := range ctx.Modifiers {
			content += "- " + m.Type + ": " + m.Justification
			if m.AppliedBy != "" {
				content += " (applied_by: " + m.AppliedBy + ")"
			}
			content += "\n"
		}
		content += "\n"
	}

	content += `## Blockers
None.
`

	path := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "failed to write SESSION_CONTEXT.md")
}

// readWhiteSails reads and parses the WHITE_SAILS.yaml file from the session directory.
func readWhiteSails(t *testing.T, sessionDir string) WhiteSailsYAML {
	t.Helper()
	path := filepath.Join(sessionDir, "WHITE_SAILS.yaml")

	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read WHITE_SAILS.yaml")

	var sails WhiteSailsYAML
	err = yaml.Unmarshal(content, &sails)
	require.NoError(t, err, "failed to parse WHITE_SAILS.yaml")

	return sails
}

// createTestSession creates a temp session directory and returns its path.
func createTestSession(t *testing.T) string {
	t.Helper()
	sessionDir := filepath.Join(t.TempDir(), "session-20260105-120000-abc12345")
	err := os.MkdirAll(sessionDir, 0755)
	require.NoError(t, err, "failed to create session directory")
	return sessionDir
}

// writePassingProofs writes passing proof logs for tests, build, and lint.
func writePassingProofs(t *testing.T, sessionDir string) {
	t.Helper()
	writeProofLog(t, sessionDir, "test-output.log", "ok  github.com/test/pkg 0.123s\n47 tests passed\nexit code: 0")
	writeProofLog(t, sessionDir, "build-output.log", "go build succeeded\nexit code: 0")
	writeProofLog(t, sessionDir, "lint-output.log", "golangci-lint: no issues found\nexit code: 0")
}

// writeAllPassingProofs writes passing proof logs for all proof types (including optional).
func writeAllPassingProofs(t *testing.T, sessionDir string) {
	t.Helper()
	writePassingProofs(t, sessionDir)
	writeProofLog(t, sessionDir, "adversarial-output.log", "adversarial tests passed\nexit code: 0")
	writeProofLog(t, sessionDir, "integration-output.log", "integration tests passed\nexit code: 0")
}

// =============================================================================
// Integration Tests per TDD Section 11.2
// =============================================================================

// TestIntegration_sails_001_WhiteWithAllProofsPassing verifies that a session with
// all proofs passing and no open questions produces WHITE sails.
// TLA+ Property: ColorComputation
func TestIntegration_sails_001_WhiteWithAllProofsPassing(t *testing.T) {
	// Setup
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity:    "MODULE",
		Type:          "standard",
		OpenQuestions: nil,
	})

	// Execute
	g := NewGenerator(sessionDir)
	fixedTime := time.Date(2026, 1, 5, 14, 30, 22, 0, time.UTC)
	g.Now = func() time.Time { return fixedTime }

	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorWhite, result.Color, "expected WHITE for all proofs passing")
	assert.Equal(t, ColorWhite, result.ComputedBase, "expected WHITE computed base")
	assert.FileExists(t, filepath.Join(sessionDir, "WHITE_SAILS.yaml"))

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "WHITE", sails.Color)
	assert.Equal(t, "WHITE", sails.ComputedBase)
	assert.Equal(t, "PASS", sails.Proofs["tests"].Status)
	assert.Equal(t, "PASS", sails.Proofs["build"].Status)
	assert.Equal(t, "PASS", sails.Proofs["lint"].Status)
	assert.Empty(t, sails.OpenQuestions, "expected no open questions")
	assert.Contains(t, result.Reasons, "all required proofs present and passing")
}

// TestIntegration_sails_002_GrayWithOpenQuestions verifies that a session with
// open questions produces GRAY sails even when all proofs pass.
// TLA+ Property: GrayCeiling
func TestIntegration_sails_002_GrayWithOpenQuestions(t *testing.T) {
	// Setup
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
		OpenQuestions: []string{
			"How should we handle edge case X?",
			"What is the expected behavior for Y?",
		},
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY due to open questions")
	assert.Equal(t, ColorGray, result.ComputedBase, "expected GRAY computed base")
	assert.FileExists(t, filepath.Join(sessionDir, "WHITE_SAILS.yaml"))

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "GRAY", sails.Color)
	assert.Equal(t, "GRAY", sails.ComputedBase)
	assert.Len(t, sails.OpenQuestions, 2, "expected 2 open questions")
	assert.Contains(t, result.Reasons, "open questions present: gray ceiling applied")
}

// TestIntegration_sails_003_GrayWithMissingProofs verifies that a session with
// missing required proofs produces GRAY sails.
// TLA+ Property: ProofRequirement
func TestIntegration_sails_003_GrayWithMissingProofs(t *testing.T) {
	// Setup - only write test output, missing build and lint
	sessionDir := createTestSession(t)
	writeProofLog(t, sessionDir, "test-output.log", "47 tests passed\nexit code: 0")
	// Intentionally NOT writing build-output.log and lint-output.log
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY due to missing proofs")
	assert.Equal(t, ColorGray, result.ComputedBase, "expected GRAY computed base")
	assert.FileExists(t, filepath.Join(sessionDir, "WHITE_SAILS.yaml"))

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "GRAY", sails.Color)

	// Build proof should be UNKNOWN (missing file)
	assert.Equal(t, "UNKNOWN", sails.Proofs["build"].Status, "expected UNKNOWN for missing build proof")
}

// TestIntegration_sails_004_BlackWithFailingTests verifies that a session with
// failing tests produces BLACK sails.
// TLA+ Property: FailureDetection
func TestIntegration_sails_004_BlackWithFailingTests(t *testing.T) {
	// Setup
	sessionDir := createTestSession(t)
	writeProofLog(t, sessionDir, "test-output.log", "FAIL github.com/test/pkg 0.123s\n5 tests failed\nexit status 1")
	writeProofLog(t, sessionDir, "build-output.log", "go build succeeded\nexit code: 0")
	writeProofLog(t, sessionDir, "lint-output.log", "golangci-lint: no issues found\nexit code: 0")
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorBlack, result.Color, "expected BLACK due to failing tests")
	assert.Equal(t, ColorBlack, result.ComputedBase, "expected BLACK computed base")
	assert.FileExists(t, filepath.Join(sessionDir, "WHITE_SAILS.yaml"))

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "BLACK", sails.Color)
	assert.Equal(t, "BLACK", sails.ComputedBase)
	assert.Equal(t, "FAIL", sails.Proofs["tests"].Status, "expected FAIL for tests")
	assert.Contains(t, result.Reasons, "proof 'tests' has status FAIL")
}

// TestIntegration_sails_005_SpikeAlwaysGray verifies that spike sessions always
// produce GRAY sails regardless of proof status.
// TLA+ Property: TypeCeiling
func TestIntegration_sails_005_SpikeAlwaysGray(t *testing.T) {
	// Setup - spike with all passing proofs
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "spike",
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY for spike session")
	assert.Equal(t, ColorGray, result.ComputedBase, "expected GRAY computed base for spike")
	assert.FileExists(t, filepath.Join(sessionDir, "WHITE_SAILS.yaml"))

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "GRAY", sails.Color)
	assert.Equal(t, "GRAY", sails.ComputedBase)
	assert.Contains(t, result.Reasons, "session type 'spike' has gray ceiling (spikes never white)")
}

// TestIntegration_sails_006_HotfixAlwaysGray verifies that hotfix sessions always
// produce GRAY sails regardless of proof status.
// TLA+ Property: TypeCeiling
func TestIntegration_sails_006_HotfixAlwaysGray(t *testing.T) {
	// Setup - hotfix with all passing proofs
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "hotfix",
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY for hotfix session")
	assert.Equal(t, ColorGray, result.ComputedBase, "expected GRAY computed base for hotfix")
	assert.FileExists(t, filepath.Join(sessionDir, "WHITE_SAILS.yaml"))

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "GRAY", sails.Color)
	assert.Equal(t, "GRAY", sails.ComputedBase)
	assert.Contains(t, result.Reasons, "session type 'hotfix' has gray ceiling (expedited gray)")
}

// TestIntegration_sails_007_HumanDowngradeOverride verifies that human modifiers
// can downgrade WHITE to GRAY.
// TLA+ Property: ModifierApplication
func TestIntegration_sails_007_HumanDowngradeOverride(t *testing.T) {
	// Setup - all passing proofs, but human adds downgrade modifier
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
		Modifiers: []ModifierInput{
			{
				Type:          "HUMAN_OVERRIDE_GRAY",
				Justification: "Need senior review before shipping despite passing tests",
				AppliedBy:     "human",
			},
		},
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY after human override")
	assert.Equal(t, ColorWhite, result.ComputedBase, "expected WHITE computed base before modifier")
	assert.FileExists(t, filepath.Join(sessionDir, "WHITE_SAILS.yaml"))

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "GRAY", sails.Color)
	assert.Equal(t, "WHITE", sails.ComputedBase)
	require.Len(t, sails.Modifiers, 1, "expected 1 modifier")
	assert.Equal(t, "HUMAN_OVERRIDE_GRAY", sails.Modifiers[0].Type)
	assert.Contains(t, sails.Modifiers[0].Justification, "senior review")
}

// TestIntegration_sails_008_QAUpgradeGrayToWhite verifies that a QA session can
// upgrade GRAY to WHITE when proper conditions are met.
// TLA+ Property: QAUpgrade
func TestIntegration_sails_008_QAUpgradeGrayToWhite(t *testing.T) {
	// Setup - session with open questions (normally GRAY), but with QA upgrade
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)

	// Write SESSION_CONTEXT.md with open questions
	contextContent := `---
schema_version: "2.1"
session_id: "session-20260105-120000-abc12345"
status: ACTIVE
created_at: "2026-01-05T12:00:00Z"
initiative: "Test Initiative"
complexity: MODULE
active_rite: 10x-dev-pack
current_phase: implementation
---

# Session: Test Initiative

## Open Questions
- Original question that was resolved by QA?

## Blockers
None.
`
	path := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	err := os.WriteFile(path, []byte(contextContent), 0644)
	require.NoError(t, err)

	// Create generator with QA upgrade capability
	// For this test, we'll test the color computation directly with QA upgrade
	g := NewGenerator(sessionDir)
	fixedTime := time.Date(2026, 1, 5, 14, 30, 22, 0, time.UTC)
	g.Now = func() time.Time { return fixedTime }

	// First, generate without QA upgrade to get GRAY
	result, err := g.Generate()
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY before QA upgrade")

	// Now test the color computation directly with QA upgrade
	qaUpgradeTime := time.Date(2026, 1, 6, 10, 0, 0, 0, time.UTC)
	qaUpgrade := &QAUpgrade{
		UpgradedAt:              &qaUpgradeTime,
		QASessionID:             "session-20260106-100000-qa123456",
		ConstraintResolutionLog: "docs/testing/TP-qa-findings.md",
		AdversarialTestsAdded: []string{
			"tests/integration/edge_case_test.go",
		},
	}

	// Test color computation with QA upgrade
	colorInput := ColorInput{
		SessionType:   "standard",
		Complexity:    "MODULE",
		Proofs:        result.Proofs,
		OpenQuestions: []string{"Original question that was resolved by QA?"},
		QAUpgrade:     qaUpgrade,
	}

	colorResult := ComputeColor(colorInput)

	// Verify QA upgrade applied
	assert.Equal(t, ColorWhite, colorResult.Color, "expected WHITE after QA upgrade")
	assert.Equal(t, ColorGray, colorResult.ComputedBase, "expected GRAY computed base (before QA upgrade)")
	assert.Contains(t, colorResult.Reasons, "QA upgrade applied: gray -> white via QA session session-20260106-100000-qa123456")
}

// TestIntegration_sails_009_CannotSelfUpgrade verifies that a session cannot
// self-upgrade from GRAY to WHITE without valid QA upgrade conditions.
// TLA+ Property: NoSelfUpgrade
func TestIntegration_sails_009_CannotSelfUpgrade(t *testing.T) {
	// Setup - session with open questions, attempting to self-upgrade
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
		OpenQuestions: []string{
			"Unresolved question?",
		},
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify initial state is GRAY
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY with open questions")

	// Now test various invalid QA upgrade scenarios

	// Scenario A: QA upgrade without constraint resolution log
	colorInputA := ColorInput{
		SessionType:   "standard",
		Complexity:    "MODULE",
		Proofs:        result.Proofs,
		OpenQuestions: []string{"Unresolved question?"},
		QAUpgrade: &QAUpgrade{
			QASessionID:           "session-20260106-100000-qa123456",
			AdversarialTestsAdded: []string{"some_test.go"},
			// Missing ConstraintResolutionLog
		},
	}
	colorResultA := ComputeColor(colorInputA)
	assert.Equal(t, ColorGray, colorResultA.Color, "expected GRAY - cannot upgrade without constraint resolution log")
	assert.Contains(t, colorResultA.Reasons, "QA upgrade missing constraint_resolution_log: cannot upgrade")

	// Scenario B: QA upgrade without adversarial tests
	colorInputB := ColorInput{
		SessionType:   "standard",
		Complexity:    "MODULE",
		Proofs:        result.Proofs,
		OpenQuestions: []string{"Unresolved question?"},
		QAUpgrade: &QAUpgrade{
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "docs/testing/TP-qa-findings.md",
			// Missing AdversarialTestsAdded
		},
	}
	colorResultB := ComputeColor(colorInputB)
	assert.Equal(t, ColorGray, colorResultB.Color, "expected GRAY - cannot upgrade without adversarial tests")
	assert.Contains(t, colorResultB.Reasons, "QA upgrade has no adversarial_tests_added: cannot upgrade")

	// Scenario C: QA upgrade on WHITE base (should stay WHITE, not change)
	colorInputC := ColorInput{
		SessionType:   "standard",
		Complexity:    "MODULE",
		Proofs:        result.Proofs,
		OpenQuestions: nil, // No open questions = WHITE base
		QAUpgrade: &QAUpgrade{
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "docs/testing/TP-qa-findings.md",
			AdversarialTestsAdded:   []string{"some_test.go"},
		},
	}
	colorResultC := ComputeColor(colorInputC)
	assert.Equal(t, ColorWhite, colorResultC.Color, "expected WHITE - QA upgrade doesn't change WHITE base")
	assert.Equal(t, ColorWhite, colorResultC.ComputedBase, "expected WHITE computed base")
}

// =============================================================================
// Additional Integration Tests for Edge Cases
// =============================================================================

// TestIntegration_ComplexityThresholds_INITIATIVE verifies that INITIATIVE
// complexity requires adversarial and integration proofs.
func TestIntegration_ComplexityThresholds_INITIATIVE(t *testing.T) {
	// Setup - INITIATIVE with only basic proofs (missing adversarial, integration)
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir) // Only tests, build, lint
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "INITIATIVE",
		Type:       "standard",
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify - should be GRAY due to missing required proofs
	require.NoError(t, err)
	assert.Equal(t, ColorGray, result.Color, "expected GRAY for INITIATIVE missing adversarial/integration")

	// Now test with all proofs
	sessionDir2 := createTestSession(t)
	writeAllPassingProofs(t, sessionDir2)
	writeSessionContext(t, sessionDir2, SessionContextInput{
		Complexity: "INITIATIVE",
		Type:       "standard",
	})

	g2 := NewGenerator(sessionDir2)
	result2, err := g2.Generate()

	require.NoError(t, err)
	assert.Equal(t, ColorWhite, result2.Color, "expected WHITE for INITIATIVE with all proofs")
}

// TestIntegration_DowngradeToBlackModifier verifies DOWNGRADE_TO_BLACK modifier.
func TestIntegration_DowngradeToBlackModifier(t *testing.T) {
	// Setup - all passing proofs with DOWNGRADE_TO_BLACK modifier
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
		Modifiers: []ModifierInput{
			{
				Type:          "DOWNGRADE_TO_BLACK",
				Justification: "Critical security vulnerability discovered post-implementation",
				AppliedBy:     "agent",
			},
		},
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, ColorBlack, result.Color, "expected BLACK after DOWNGRADE_TO_BLACK")
	assert.Equal(t, ColorWhite, result.ComputedBase, "expected WHITE computed base before modifier")

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "BLACK", sails.Color)
	assert.Equal(t, "WHITE", sails.ComputedBase)
	require.Len(t, sails.Modifiers, 1)
	assert.Equal(t, "DOWNGRADE_TO_BLACK", sails.Modifiers[0].Type)
}

// TestIntegration_MultipleModifiers verifies multiple modifiers are applied in order.
func TestIntegration_MultipleModifiers(t *testing.T) {
	// Setup - all passing proofs with multiple modifiers
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
		Modifiers: []ModifierInput{
			{
				Type:          "DOWNGRADE_TO_GRAY",
				Justification: "First concern - needs review",
				AppliedBy:     "agent",
			},
			{
				Type:          "DOWNGRADE_TO_BLACK",
				Justification: "Second concern - blocking issue",
				AppliedBy:     "human",
			},
		},
	})

	// Execute
	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	// Verify - BLACK should be the final color due to second modifier
	require.NoError(t, err)
	assert.Equal(t, ColorBlack, result.Color, "expected BLACK after multiple modifiers")
	assert.Equal(t, ColorWhite, result.ComputedBase, "expected WHITE computed base")

	// Verify YAML content
	sails := readWhiteSails(t, sessionDir)
	assert.Len(t, sails.Modifiers, 2)
}

// TestIntegration_BuildFailure verifies that build failure produces BLACK.
func TestIntegration_BuildFailure(t *testing.T) {
	sessionDir := createTestSession(t)
	writeProofLog(t, sessionDir, "test-output.log", "47 tests passed\nexit code: 0")
	writeProofLog(t, sessionDir, "build-output.log", "build failed: compilation error\nexit status 1")
	writeProofLog(t, sessionDir, "lint-output.log", "exit code: 0")
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
	})

	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	require.NoError(t, err)
	assert.Equal(t, ColorBlack, result.Color, "expected BLACK due to build failure")
	assert.Contains(t, result.Reasons, "proof 'build' has status FAIL")
}

// TestIntegration_LintFailure verifies that lint failure produces BLACK.
func TestIntegration_LintFailure(t *testing.T) {
	sessionDir := createTestSession(t)
	writeProofLog(t, sessionDir, "test-output.log", "47 tests passed\nexit code: 0")
	writeProofLog(t, sessionDir, "build-output.log", "build succeeded\nexit code: 0")
	writeProofLog(t, sessionDir, "lint-output.log", "3 errors, 5 warnings\nexit status 1")
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
	})

	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	require.NoError(t, err)
	assert.Equal(t, ColorBlack, result.Color, "expected BLACK due to lint failure")
}

// TestIntegration_SessionIDExtraction verifies session ID is correctly extracted.
func TestIntegration_SessionIDExtraction(t *testing.T) {
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		SessionID:  "session-20260105-120000-abc12345",
		Complexity: "MODULE",
		Type:       "standard",
	})

	g := NewGenerator(sessionDir)
	result, err := g.Generate()

	require.NoError(t, err)
	assert.Equal(t, "session-20260105-120000-abc12345", result.SessionID)

	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "session-20260105-120000-abc12345", sails.SessionID)
}

// TestIntegration_SchemaVersion verifies schema version is set correctly.
func TestIntegration_SchemaVersion(t *testing.T) {
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity: "MODULE",
		Type:       "standard",
	})

	g := NewGenerator(sessionDir)
	_, err := g.Generate()

	require.NoError(t, err)
	sails := readWhiteSails(t, sessionDir)
	assert.Equal(t, "1.0", sails.SchemaVersion)
}

// TestIntegration_OpenQuestionsNullVsEmpty verifies open_questions is array not null.
func TestIntegration_OpenQuestionsNullVsEmpty(t *testing.T) {
	sessionDir := createTestSession(t)
	writePassingProofs(t, sessionDir)
	writeSessionContext(t, sessionDir, SessionContextInput{
		Complexity:    "MODULE",
		Type:          "standard",
		OpenQuestions: nil,
	})

	g := NewGenerator(sessionDir)
	_, err := g.Generate()

	require.NoError(t, err)
	sails := readWhiteSails(t, sessionDir)

	// Should be empty array, not nil
	assert.NotNil(t, sails.OpenQuestions, "open_questions should be empty array, not nil")
	assert.Empty(t, sails.OpenQuestions)
}

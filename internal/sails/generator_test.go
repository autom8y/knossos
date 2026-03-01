package sails

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator("/path/to/session")
	if gen.SessionPath != "/path/to/session" {
		t.Errorf("Expected session path '/path/to/session', got '%s'", gen.SessionPath)
	}
	if gen.Validator != nil {
		t.Error("Expected nil validator for basic generator")
	}
	if gen.Now == nil {
		t.Error("Expected Now function to be set")
	}
}

func TestGenerator_Generate_EmptyPath(t *testing.T) {
	gen := NewGenerator("")
	_, err := gen.Generate()
	if err == nil {
		t.Fatal("Expected error for empty session path")
	}
	if !strings.Contains(err.Error(), "session path is required") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestGenerator_Generate_NonexistentPath(t *testing.T) {
	gen := NewGenerator("/nonexistent/path/to/session")
	_, err := gen.Generate()
	if err == nil {
		t.Fatal("Expected error for nonexistent session path")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestGenerator_Generate_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gen := NewGenerator(filePath)
	_, err := gen.Generate()
	if err == nil {
		t.Fatal("Expected error when path is not a directory")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestGenerator_Generate_EmptySession(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-143022-abc12345")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	gen := NewGenerator(sessionDir)
	// Use a fixed time for testing
	fixedTime := time.Date(2026, 1, 5, 14, 30, 22, 0, time.UTC)
	gen.Now = func() time.Time { return fixedTime }

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// With no proof logs, all proofs should be UNKNOWN, leading to GRAY
	if result.Color != ColorGray {
		t.Errorf("Expected GRAY color for empty session, got %s", result.Color)
	}
	if result.ComputedBase != ColorGray {
		t.Errorf("Expected GRAY computed base, got %s", result.ComputedBase)
	}

	// Verify WHITE_SAILS.yaml was created
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	if _, err := os.Stat(sailsPath); os.IsNotExist(err) {
		t.Error("WHITE_SAILS.yaml was not created")
	}

	// Verify the session ID was extracted from path
	if result.SessionID != "session-20260105-143022-abc12345" {
		t.Errorf("Expected session ID from path, got '%s'", result.SessionID)
	}
}

func TestGenerator_Generate_AllProofsPassing(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-150000-def45678")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Create passing proof logs
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("ok  github.com/test/pkg 0.123s\nexit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("go build succeeded\nexit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("golangci-lint run: no issues\nexit code: 0"))

	gen := NewGenerator(sessionDir)
	fixedTime := time.Date(2026, 1, 5, 15, 0, 0, 0, time.UTC)
	gen.Now = func() time.Time { return fixedTime }

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Color != ColorWhite {
		t.Errorf("Expected WHITE color for all passing proofs, got %s", result.Color)
	}
	if result.ComputedBase != ColorWhite {
		t.Errorf("Expected WHITE computed base, got %s", result.ComputedBase)
	}

	// Verify proofs
	if result.Proofs["tests"].Status != ProofPass {
		t.Errorf("Expected PASS for tests, got %s", result.Proofs["tests"].Status)
	}
	if result.Proofs["build"].Status != ProofPass {
		t.Errorf("Expected PASS for build, got %s", result.Proofs["build"].Status)
	}
	if result.Proofs["lint"].Status != ProofPass {
		t.Errorf("Expected PASS for lint, got %s", result.Proofs["lint"].Status)
	}
}

func TestGenerator_Generate_TestsFailing(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-160000-ghi78901")
	mustMkdirAll(t, sessionDir)

	// Create failing test log
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("FAIL github.com/test/pkg 0.123s\nexit status 1"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("exit code: 0"))

	gen := NewGenerator(sessionDir)

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Failing tests should result in BLACK
	if result.Color != ColorBlack {
		t.Errorf("Expected BLACK color for failing tests, got %s", result.Color)
	}
	if result.ComputedBase != ColorBlack {
		t.Errorf("Expected BLACK computed base, got %s", result.ComputedBase)
	}
}

func TestGenerator_Generate_WithSessionContext(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-170000-jkl23456")
	mustMkdirAll(t, sessionDir)

	// Create SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: session-20260105-170000-jkl23456
status: ACTIVE
created_at: "2026-01-05T17:00:00Z"
initiative: "Test Initiative"
complexity: MODULE
active_rite: 10x-dev
current_phase: implementation
---

# Session: Test Initiative

## Open Questions
- What about edge case X?
- How to handle Y scenario?

## Blockers
None.
`
	mustWriteFile(t, filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent))

	// Create passing proof logs
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("exit code: 0"))

	gen := NewGenerator(sessionDir)

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Open questions should cause GRAY ceiling
	if result.Color != ColorGray {
		t.Errorf("Expected GRAY color due to open questions, got %s", result.Color)
	}

	// Verify session ID was extracted from context
	if result.SessionID != "session-20260105-170000-jkl23456" {
		t.Errorf("Expected session ID from context, got '%s'", result.SessionID)
	}

	// Verify open questions were extracted
	if len(result.OpenQuestions) != 2 {
		t.Errorf("Expected 2 open questions, got %d", len(result.OpenQuestions))
	}
}

func TestGenerator_Generate_WithModifiers(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-180000-mno56789")
	mustMkdirAll(t, sessionDir)

	// Create SESSION_CONTEXT.md with modifier
	contextContent := `---
schema_version: "2.1"
session_id: session-20260105-180000-mno56789
status: ACTIVE
created_at: "2026-01-05T18:00:00Z"
initiative: "Test Initiative"
complexity: MODULE
active_rite: 10x-dev
current_phase: implementation
---

# Session: Test Initiative

## Open Questions
None.

## Modifiers
- DOWNGRADE_TO_GRAY: Uncertainty about production impact (applied_by: agent)
`
	mustWriteFile(t, filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent))

	// Create passing proof logs
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("10 passed\nexit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("exit code: 0"))

	gen := NewGenerator(sessionDir)

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Modifier should downgrade WHITE to GRAY
	if result.Color != ColorGray {
		t.Errorf("Expected GRAY color due to modifier, got %s", result.Color)
	}
	if result.ComputedBase != ColorWhite {
		t.Errorf("Expected WHITE computed base before modifier, got %s", result.ComputedBase)
	}

	// Verify modifier was extracted
	if len(result.Modifiers) != 1 {
		t.Errorf("Expected 1 modifier, got %d", len(result.Modifiers))
	}
	if len(result.Modifiers) > 0 && result.Modifiers[0].Type != ModifierDowngradeToGray {
		t.Errorf("Expected DOWNGRADE_TO_GRAY modifier, got %s", result.Modifiers[0].Type)
	}
}

func TestGenerator_Generate_YAMLContent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-190000-pqr89012")
	mustMkdirAll(t, sessionDir)

	// Create passing proof logs
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("47 tests passed\nexit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("go build succeeded\nexit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("golangci-lint clean\nexit code: 0"))

	gen := NewGenerator(sessionDir)
	fixedTime := time.Date(2026, 1, 5, 19, 0, 0, 0, time.UTC)
	gen.Now = func() time.Time { return fixedTime }

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read and parse the generated YAML
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	content, err := os.ReadFile(sailsPath)
	if err != nil {
		t.Fatalf("Failed to read WHITE_SAILS.yaml: %v", err)
	}

	var sails WhiteSailsYAML
	if err := yaml.Unmarshal(content, &sails); err != nil {
		t.Fatalf("Failed to parse WHITE_SAILS.yaml: %v", err)
	}

	// Verify required fields
	if sails.SchemaVersion != "1.0" {
		t.Errorf("Expected schema_version '1.0', got '%s'", sails.SchemaVersion)
	}
	if sails.SessionID != "session-20260105-190000-pqr89012" {
		t.Errorf("Unexpected session_id: '%s'", sails.SessionID)
	}
	if sails.Color != "WHITE" {
		t.Errorf("Expected color 'WHITE', got '%s'", sails.Color)
	}
	if sails.ComputedBase != "WHITE" {
		t.Errorf("Expected computed_base 'WHITE', got '%s'", sails.ComputedBase)
	}

	// Verify proofs
	if sails.Proofs["tests"].Status != "PASS" {
		t.Errorf("Expected tests status 'PASS', got '%s'", sails.Proofs["tests"].Status)
	}
	if sails.Proofs["build"].Status != "PASS" {
		t.Errorf("Expected build status 'PASS', got '%s'", sails.Proofs["build"].Status)
	}
	if sails.Proofs["lint"].Status != "PASS" {
		t.Errorf("Expected lint status 'PASS', got '%s'", sails.Proofs["lint"].Status)
	}

	// Verify open_questions is an empty array, not null
	if sails.OpenQuestions == nil {
		t.Error("Expected open_questions to be an empty array, not nil")
	}

	// Verify file path in result
	if result.FilePath != sailsPath {
		t.Errorf("Expected file path '%s', got '%s'", sailsPath, result.FilePath)
	}
}

func TestExtractOpenQuestions(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "no open questions section",
			body:     "# Session\n\n## Blockers\nNone.",
			expected: nil,
		},
		{
			name: "empty open questions",
			body: `# Session

## Open Questions
None.

## Blockers
None.`,
			expected: nil,
		},
		{
			name: "single question",
			body: `## Open Questions
- What about edge case X?

## Blockers`,
			expected: []string{"What about edge case X?"},
		},
		{
			name: "multiple questions",
			body: `## Open Questions
- First question?
- Second question?
* Third question (with asterisk)?

## Next Section`,
			expected: []string{"First question?", "Second question?", "Third question (with asterisk)?"},
		},
		{
			name: "case insensitive header",
			body: `## open questions
- A question?`,
			expected: []string{"A question?"},
		},
		{
			name: "N/A should be filtered",
			body: `## Open Questions
- N/A
- Real question?`,
			expected: []string{"Real question?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOpenQuestions(tt.body)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d questions, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, q := range tt.expected {
				if result[i] != q {
					t.Errorf("Question %d: expected '%s', got '%s'", i, q, result[i])
				}
			}
		})
	}
}

func TestExtractBlockers(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "no blockers section",
			body:     "# Session\n\n## Open Questions\nNone.",
			expected: nil,
		},
		{
			name: "blockers section with None",
			body: `# Session

## Blockers
None.

## Next Steps`,
			expected: nil,
		},
		{
			name: "blockers section with None yet",
			body: `# Session

## Blockers
None yet.`,
			expected: nil,
		},
		{
			name: "single blocker",
			body: `## Blockers
- Waiting for database migration approval

## Next Steps`,
			expected: []string{"Waiting for database migration approval"},
		},
		{
			name: "multiple blockers",
			body: `## Blockers
- Security review pending
- Waiting for external API access
* Third blocker with asterisk

## Open Questions`,
			expected: []string{"Security review pending", "Waiting for external API access", "Third blocker with asterisk"},
		},
		{
			name: "case insensitive header",
			body: `## blockers
- A blocker item`,
			expected: []string{"A blocker item"},
		},
		{
			name: "N/A should be filtered",
			body: `## Blockers
- N/A
- Real blocker
- none
- Another real blocker`,
			expected: []string{"Real blocker", "Another real blocker"},
		},
		{
			name: "singular form (Blocker) should work",
			body: `## Blocker
- Some blocker`,
			expected: []string{"Some blocker"},
		},
		{
			name: "None at start of text should be filtered",
			body: `## Blockers
- None at this time
- Real blocker`,
			expected: []string{"Real blocker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBlockers(tt.body)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d blockers, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, b := range tt.expected {
				if result[i] != b {
					t.Errorf("Blocker %d: expected '%s', got '%s'", i, b, result[i])
				}
			}
		})
	}
}

func TestExtractModifiers(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []Modifier
	}{
		{
			name:     "no modifiers section",
			body:     "# Session\n\n## Blockers\nNone.",
			expected: nil,
		},
		{
			name: "single modifier with applied_by",
			body: `## Modifiers
- DOWNGRADE_TO_GRAY: Uncertainty about impact (applied_by: agent)`,
			expected: []Modifier{
				{Type: ModifierDowngradeToGray, Justification: "Uncertainty about impact", AppliedBy: "agent"},
			},
		},
		{
			name: "modifier without applied_by defaults to agent",
			body: `## Modifiers
- DOWNGRADE_TO_BLACK: Critical issue found`,
			expected: []Modifier{
				{Type: ModifierDowngradeToBlack, Justification: "Critical issue found", AppliedBy: "agent"},
			},
		},
		{
			name: "human override modifier",
			body: `## Modifiers
- HUMAN_OVERRIDE_GRAY: Not ready for production (applied_by: human)`,
			expected: []Modifier{
				{Type: ModifierHumanOverrideGray, Justification: "Not ready for production", AppliedBy: "human"},
			},
		},
		{
			name: "multiple modifiers",
			body: `## Modifiers
- DOWNGRADE_TO_GRAY: First reason (applied_by: agent)
- DOWNGRADE_TO_BLACK: Second reason (applied_by: human)`,
			expected: []Modifier{
				{Type: ModifierDowngradeToGray, Justification: "First reason", AppliedBy: "agent"},
				{Type: ModifierDowngradeToBlack, Justification: "Second reason", AppliedBy: "human"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractModifiers(tt.body)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d modifiers, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, m := range tt.expected {
				if result[i].Type != m.Type {
					t.Errorf("Modifier %d type: expected '%s', got '%s'", i, m.Type, result[i].Type)
				}
				if result[i].Justification != m.Justification {
					t.Errorf("Modifier %d justification: expected '%s', got '%s'", i, m.Justification, result[i].Justification)
				}
				if result[i].AppliedBy != m.AppliedBy {
					t.Errorf("Modifier %d applied_by: expected '%s', got '%s'", i, m.AppliedBy, result[i].AppliedBy)
				}
			}
		})
	}
}

func TestExtractSessionType(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "no session type section defaults to standard",
			body:     "# Session\n\n## Blockers\nNone.",
			expected: "standard",
		},
		{
			name: "spike session type",
			body: `## Session Type
spike

## Blockers`,
			expected: "spike",
		},
		{
			name: "hotfix session type",
			body: `## Session Type
hotfix

## Blockers`,
			expected: "hotfix",
		},
		{
			name: "standard session type",
			body: `## Session Type
standard

## Blockers`,
			expected: "standard",
		},
		{
			name: "spike with description",
			body: `## Session Type
spike - exploring implementation options

## Next`,
			expected: "spike",
		},
		{
			name: "hotfix with description",
			body: `## Session Type
hotfix - urgent production fix

## Next`,
			expected: "hotfix",
		},
		{
			name: "case insensitive header",
			body: `## session type
spike

## Next`,
			expected: "spike",
		},
		{
			name: "unknown type defaults to standard",
			body: `## Session Type
experimental

## Next`,
			expected: "standard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSessionType(tt.body)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerator_Generate_WithOptionalProofs(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-200000-stu34567")
	mustMkdirAll(t, sessionDir)

	// Create SESSION_CONTEXT.md with INITIATIVE complexity (requires adversarial/integration)
	contextContent := `---
schema_version: "2.1"
session_id: session-20260105-200000-stu34567
status: ACTIVE
created_at: "2026-01-05T20:00:00Z"
initiative: "Big Initiative"
complexity: INITIATIVE
active_rite: 10x-dev
current_phase: implementation
---

# Session: Big Initiative

## Open Questions
None.

## Blockers
None.
`
	mustWriteFile(t, filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent))

	// Create all proof logs (including optional ones)
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "adversarial-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "integration-output.log"),
		[]byte("exit code: 0"))

	gen := NewGenerator(sessionDir)

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// All proofs present and passing for INITIATIVE = WHITE
	if result.Color != ColorWhite {
		t.Errorf("Expected WHITE color with all proofs, got %s", result.Color)
	}

	// Verify all proofs were collected
	if result.Proofs["adversarial"].Status != ProofPass {
		t.Errorf("Expected PASS for adversarial, got %s", result.Proofs["adversarial"].Status)
	}
	if result.Proofs["integration"].Status != ProofPass {
		t.Errorf("Expected PASS for integration, got %s", result.Proofs["integration"].Status)
	}
}

func TestGenerator_Generate_MissingRequiredProofsForComplexity(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-210000-vwx67890")
	mustMkdirAll(t, sessionDir)

	// Create SESSION_CONTEXT.md with INITIATIVE complexity
	contextContent := `---
schema_version: "2.1"
session_id: session-20260105-210000-vwx67890
status: ACTIVE
created_at: "2026-01-05T21:00:00Z"
initiative: "Another Initiative"
complexity: INITIATIVE
active_rite: 10x-dev
current_phase: implementation
---

# Session: Another Initiative
`
	mustWriteFile(t, filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent))

	// Create only basic proof logs (missing adversarial/integration for INITIATIVE)
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("exit code: 0"))

	gen := NewGenerator(sessionDir)

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Missing required proofs for INITIATIVE should result in GRAY
	if result.Color != ColorGray {
		t.Errorf("Expected GRAY color for missing required proofs, got %s", result.Color)
	}
}

func TestGeneratorFromProject_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .sos directory structure but no sessions
	mustMkdirAll(t, filepath.Join(tmpDir, ".sos", "sessions"))

	_, err := GeneratorFromProject(tmpDir, "")
	if err == nil {
		t.Fatal("Expected error for no active session")
	}
	if !strings.Contains(err.Error(), "no session ID") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestGeneratorFromProject_WithActiveSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .sos directory structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	mustMkdirAll(t, sessionsDir)

	// Create a session directory
	sessionID := "session-20260105-220000-yz012345"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	mustMkdirAll(t, sessionDir)

	gen, err := GeneratorFromProject(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPath := filepath.Join(sessionsDir, sessionID)
	if gen.SessionPath != expectedPath {
		t.Errorf("Expected session path '%s', got '%s'", expectedPath, gen.SessionPath)
	}
}

func TestGenerator_proofSetToColorProofs(t *testing.T) {
	gen := NewGenerator("/test")
	now := time.Now()

	proofSet := &ProofSet{
		Tests: &ProofItem{
			Status:       ProofPass,
			EvidencePath: "/path/to/test.log",
			Summary:      "10 passed",
			ExitCode:     0,
			Timestamp:    now,
		},
		Build: &ProofItem{
			Status:       ProofPass,
			EvidencePath: "/path/to/build.log",
			Summary:      "build succeeded",
			ExitCode:     0,
			Timestamp:    now,
		},
		Lint: nil, // No lint proof
	}

	proofs := gen.proofSetToColorProofs(proofSet)

	if len(proofs) != 2 {
		t.Errorf("Expected 2 proofs, got %d", len(proofs))
	}

	if proofs["tests"].Status != ProofPass {
		t.Errorf("Expected tests PASS, got %s", proofs["tests"].Status)
	}
	if proofs["build"].Status != ProofPass {
		t.Errorf("Expected build PASS, got %s", proofs["build"].Status)
	}
	if _, ok := proofs["lint"]; ok {
		t.Error("Lint proof should not be present when nil")
	}
}

func TestGenerator_extractSessionIDFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{
			path:     "/path/to/sessions/session-20260105-143022-abc12345",
			expected: "session-20260105-143022-abc12345",
		},
		{
			path:     "session-20260105-143022-abc12345",
			expected: "session-20260105-143022-abc12345",
		},
		{
			path:     "/single",
			expected: "single",
		},
	}

	for _, tt := range tests {
		gen := NewGenerator(tt.path)
		result := gen.extractSessionIDFromPath()
		if result != tt.expected {
			t.Errorf("For path '%s': expected '%s', got '%s'", tt.path, tt.expected, result)
		}
	}
}

func TestGenerator_Generate_ReturnsReasons(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260105-230000-abc11111")
	mustMkdirAll(t, sessionDir)

	// Create passing proof logs
	mustWriteFile(t, filepath.Join(sessionDir, "test-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "build-output.log"),
		[]byte("exit code: 0"))
	mustWriteFile(t, filepath.Join(sessionDir, "lint-output.log"),
		[]byte("exit code: 0"))

	gen := NewGenerator(sessionDir)

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have at least one reason
	if len(result.Reasons) == 0 {
		t.Error("Expected at least one reason in result")
	}
}

// mustWriteFile is a test helper that writes a file and fails the test on error.
func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// mustMkdirAll is a test helper that creates directories and fails the test on error.
func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", path, err)
	}
}

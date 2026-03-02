package tribute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerator_Generate_MinimalContext(t *testing.T) {
	// Create temporary session directory
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-test-123")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create minimal SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: session-test-123
status: ACTIVE
created_at: "2026-01-06T10:00:00Z"
initiative: test-initiative
complexity: MODULE
active_rite: test-rite
current_phase: implementation
---

# Session: test-initiative
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Generate tribute
	gen := NewGenerator(sessionDir)
	gen.Now = func() time.Time {
		return time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)
	}

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify result
	if result.SessionID != "session-test-123" {
		t.Errorf("SessionID = %q, want %q", result.SessionID, "session-test-123")
	}
	if result.Initiative != "test-initiative" {
		t.Errorf("Initiative = %q, want %q", result.Initiative, "test-initiative")
	}
	if result.Complexity != "MODULE" {
		t.Errorf("Complexity = %q, want %q", result.Complexity, "MODULE")
	}
	if result.Rite != "test-rite" {
		t.Errorf("Rite = %q, want %q", result.Rite, "test-rite")
	}
	if result.FinalPhase != "implementation" {
		t.Errorf("FinalPhase = %q, want %q", result.FinalPhase, "implementation")
	}

	// Verify file was created
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Errorf("TRIBUTE.md not created at %s", result.FilePath)
	}

	// Read and verify content
	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("failed to read TRIBUTE.md: %v", err)
	}

	// Check for key sections
	contentStr := string(content)
	expectedSections := []string{
		"schema_version: \"1.0\"",
		"session_id: session-test-123",
		"# Tribute: test-initiative",
		"## Summary",
		"**Initiative**: test-initiative",
		"**Complexity**: MODULE",
		"## Artifacts Produced",
		"## Metrics",
	}

	for _, section := range expectedSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("TRIBUTE.md missing section: %q", section)
		}
	}
}

func TestGenerator_Generate_WithEvents(t *testing.T) {
	// Create temporary session directory
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-events-test")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: session-events-test
status: ARCHIVED
created_at: "2026-01-06T10:00:00Z"
initiative: events-test
complexity: MODULE
active_rite: 10x-dev
current_phase: validation
archived_at: "2026-01-06T14:00:00Z"
---

# Session: events-test
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create events.jsonl with various events
	eventsContent := `{"timestamp":"2026-01-06T10:00:00Z","event":"SESSION_CREATED","from":"NONE","to":"ACTIVE","metadata":{"initiative":"events-test","complexity":"MODULE","rite":"10x-dev"}}
{"ts":"2026-01-06T11:00:00Z","type":"tool.artifact_created","path":".ledge/specs/PRD-test.md","artifact_type":"PRD"}
{"ts":"2026-01-06T12:00:00Z","type":"agent.decision","decision":"Use Go for implementation","rationale":"Type safety and performance"}
{"ts":"2026-01-06T13:00:00Z","type":"agent.handoff_executed","from":"architect","to":"principal-engineer","notes":"TDD approved"}
{"timestamp":"2026-01-06T14:00:00Z","event":"SESSION_ARCHIVED","from":"ACTIVE","to":"ARCHIVED"}
`
	if err := os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(eventsContent), 0644); err != nil {
		t.Fatalf("failed to write events.jsonl: %v", err)
	}

	// Generate tribute
	gen := NewGenerator(sessionDir)
	gen.Now = func() time.Time {
		return time.Date(2026, 1, 6, 14, 30, 0, 0, time.UTC)
	}

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify extracted data
	if len(result.Artifacts) != 1 {
		t.Errorf("Artifacts count = %d, want 1", len(result.Artifacts))
	}
	if len(result.Decisions) != 1 {
		t.Errorf("Decisions count = %d, want 1", len(result.Decisions))
	}
	if len(result.Handoffs) != 1 {
		t.Errorf("Handoffs count = %d, want 1", len(result.Handoffs))
	}

	// Verify duration calculation
	expectedDuration := 4 * time.Hour
	if result.Duration != expectedDuration {
		t.Errorf("Duration = %v, want %v", result.Duration, expectedDuration)
	}

	// Verify content includes the data
	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("failed to read TRIBUTE.md: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "## Decisions Made") {
		t.Error("TRIBUTE.md missing Decisions Made section")
	}
	if !strings.Contains(contentStr, "## Handoffs") {
		t.Error("TRIBUTE.md missing Handoffs section")
	}
	if !strings.Contains(contentStr, "PRD") {
		t.Error("TRIBUTE.md missing PRD artifact")
	}
}

func TestGenerator_Generate_MissingEventsFile(t *testing.T) {
	// Create temporary session directory without events.jsonl
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-no-events")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create minimal SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: session-no-events
status: ACTIVE
created_at: "2026-01-06T10:00:00Z"
initiative: no-events-test
complexity: SCRIPT
active_rite: test-rite
current_phase: requirements
---

# Session: no-events-test
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Generate tribute - should succeed with graceful degradation
	gen := NewGenerator(sessionDir)
	gen.Now = func() time.Time {
		return time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)
	}

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() should succeed with missing events.jsonl: %v", err)
	}

	// Verify empty collections
	if len(result.Artifacts) != 0 {
		t.Errorf("Artifacts should be empty, got %d", len(result.Artifacts))
	}
	if len(result.Decisions) != 0 {
		t.Errorf("Decisions should be empty, got %d", len(result.Decisions))
	}

	// Verify TRIBUTE.md was created
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Error("TRIBUTE.md should still be created")
	}
}

func TestGenerator_Generate_Idempotent(t *testing.T) {
	// Create temporary session directory
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-idempotent")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create minimal SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: session-idempotent
status: ACTIVE
created_at: "2026-01-06T10:00:00Z"
initiative: idempotent-test
complexity: MODULE
active_rite: test-rite
current_phase: implementation
---

# Session: idempotent-test
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Fixed time for consistency
	fixedTime := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	// Generate first time
	gen1 := NewGenerator(sessionDir)
	gen1.Now = func() time.Time { return fixedTime }
	result1, err := gen1.Generate()
	if err != nil {
		t.Fatalf("First Generate() failed: %v", err)
	}

	content1, err := os.ReadFile(result1.FilePath)
	if err != nil {
		t.Fatalf("failed to read first TRIBUTE.md: %v", err)
	}

	// Generate second time with same timestamp
	gen2 := NewGenerator(sessionDir)
	gen2.Now = func() time.Time { return fixedTime }
	result2, err := gen2.Generate()
	if err != nil {
		t.Fatalf("Second Generate() failed: %v", err)
	}

	content2, err := os.ReadFile(result2.FilePath)
	if err != nil {
		t.Fatalf("failed to read second TRIBUTE.md: %v", err)
	}

	// Verify idempotent output
	if string(content1) != string(content2) {
		t.Error("Generate() is not idempotent - content differs between runs")
	}
}

func TestGenerator_Generate_MissingSessionContext(t *testing.T) {
	// Create empty session directory
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-missing-context")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Generate tribute - should fail
	gen := NewGenerator(sessionDir)
	_, err := gen.Generate()
	if err == nil {
		t.Fatal("Generate() should fail with missing SESSION_CONTEXT.md")
	}
}

func TestGenerator_Generate_InvalidPath(t *testing.T) {
	// Test with non-existent directory
	gen := NewGenerator("/nonexistent/path/session")
	_, err := gen.Generate()
	if err == nil {
		t.Fatal("Generate() should fail with non-existent path")
	}
}

func TestGenerator_Generate_WithWhiteSails(t *testing.T) {
	// Create temporary session directory
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-with-sails")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: session-with-sails
status: ARCHIVED
created_at: "2026-01-06T10:00:00Z"
initiative: sails-test
complexity: MODULE
active_rite: 10x-dev
current_phase: validation
archived_at: "2026-01-06T14:00:00Z"
---

# Session: sails-test
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create WHITE_SAILS.yaml
	sailsContent := `schema_version: "1.0"
session_id: session-with-sails
generated_at: "2026-01-06T14:00:00Z"
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
`
	if err := os.WriteFile(filepath.Join(sessionDir, "WHITE_SAILS.yaml"), []byte(sailsContent), 0644); err != nil {
		t.Fatalf("failed to write WHITE_SAILS.yaml: %v", err)
	}

	// Generate tribute
	gen := NewGenerator(sessionDir)
	gen.Now = func() time.Time {
		return time.Date(2026, 1, 6, 14, 30, 0, 0, time.UTC)
	}

	result, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify sails data
	if result.SailsColor != "WHITE" {
		t.Errorf("SailsColor = %q, want %q", result.SailsColor, "WHITE")
	}
	if result.SailsBase != "WHITE" {
		t.Errorf("SailsBase = %q, want %q", result.SailsBase, "WHITE")
	}
	if len(result.SailsProofs) != 3 {
		t.Errorf("SailsProofs count = %d, want 3", len(result.SailsProofs))
	}

	// Verify content includes White Sails section
	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("failed to read TRIBUTE.md: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "## White Sails Attestation") {
		t.Error("TRIBUTE.md missing White Sails Attestation section")
	}
	if !strings.Contains(contentStr, "**Color**: WHITE") {
		t.Error("TRIBUTE.md missing sails color")
	}
}

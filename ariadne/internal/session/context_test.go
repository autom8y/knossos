package session

import (
	"strings"
	"testing"
	"time"
)

func TestParseContext(t *testing.T) {
	content := `---
schema_version: "2.1"
session_id: "session-20260104-160414-563c681e"
status: "ACTIVE"
created_at: "2026-01-04T16:04:14Z"
initiative: "Test Initiative"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "design"
---

# Session: Test Initiative

## Artifacts
- PRD: pending
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	// Check fields
	if ctx.SchemaVersion != "2.1" {
		t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "2.1")
	}
	if ctx.SessionID != "session-20260104-160414-563c681e" {
		t.Errorf("SessionID = %q, want %q", ctx.SessionID, "session-20260104-160414-563c681e")
	}
	if ctx.Status != StatusActive {
		t.Errorf("Status = %v, want %v", ctx.Status, StatusActive)
	}
	if ctx.Initiative != "Test Initiative" {
		t.Errorf("Initiative = %q, want %q", ctx.Initiative, "Test Initiative")
	}
	if ctx.Complexity != "MODULE" {
		t.Errorf("Complexity = %q, want %q", ctx.Complexity, "MODULE")
	}
	if ctx.ActiveRite != "10x-dev" {
		t.Errorf("ActiveTeam = %q, want %q", ctx.ActiveRite, "10x-dev")
	}
	if ctx.CurrentPhase != "design" {
		t.Errorf("CurrentPhase = %q, want %q", ctx.CurrentPhase, "design")
	}
}

func TestParseContext_WithParkedFields(t *testing.T) {
	content := `---
schema_version: "2.1"
session_id: "session-20260104-160414-563c681e"
status: "PARKED"
created_at: "2026-01-04T16:04:14Z"
initiative: "Test"
complexity: "PATCH"
active_rite: "none"
current_phase: "requirements"
parked_at: "2026-01-04T18:00:00Z"
parked_reason: "End of day"
---
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if ctx.Status != StatusParked {
		t.Errorf("Status = %v, want %v", ctx.Status, StatusParked)
	}
	if ctx.ParkedAt == nil {
		t.Fatal("ParkedAt should not be nil")
	}
	if ctx.ParkedReason != "End of day" {
		t.Errorf("ParkedReason = %q, want %q", ctx.ParkedReason, "End of day")
	}
}

func TestParseContext_NoFrontmatter(t *testing.T) {
	content := `# Just markdown, no frontmatter`

	_, err := ParseContext([]byte(content))
	if err == nil {
		t.Error("ParseContext() should error on missing frontmatter")
	}
}

func TestParseContext_UnclosedFrontmatter(t *testing.T) {
	content := `---
schema_version: "2.1"
session_id: "test"
# Never closed
`

	_, err := ParseContext([]byte(content))
	if err == nil {
		t.Error("ParseContext() should error on unclosed frontmatter")
	}
}

func TestContext_Serialize(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	ctx := &Context{
		SchemaVersion: "2.1",
		SessionID:     "session-20260104-160414-563c681e",
		Status:        StatusActive,
		CreatedAt:     now,
		Initiative:    "Test Initiative",
		Complexity:    "MODULE",
		ActiveRite:    "test-pack",
		CurrentPhase:  "requirements",
		Body:          "\n# Test\n",
	}

	data, err := ctx.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	str := string(data)

	// Check contains expected content
	if !strings.Contains(str, "schema_version: \"2.1\"") {
		t.Error("Serialized content should contain schema_version")
	}
	if !strings.Contains(str, "status: ACTIVE") {
		t.Error("Serialized content should contain status")
	}
	if !strings.Contains(str, "session_id:") {
		t.Error("Serialized content should contain session_id")
	}
	if !strings.HasPrefix(str, "---\n") {
		t.Error("Serialized content should start with frontmatter delimiter")
	}
	if !strings.Contains(str, "---\n\n# Test") {
		t.Error("Serialized content should contain body after frontmatter")
	}
}

func TestContext_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	original := &Context{
		SchemaVersion: "2.1",
		SessionID:     "session-20260104-160414-563c681e",
		Status:        StatusActive,
		CreatedAt:     now,
		Initiative:    "Round Trip Test",
		Complexity:    "SYSTEM",
		ActiveRite:    "test-pack",
		CurrentPhase:  "design",
		Body:          "\n# Test Body\n",
	}

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Parse back
	parsed, err := ParseContext(data)
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	// Compare
	if parsed.SessionID != original.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", parsed.SessionID, original.SessionID)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status mismatch: got %v, want %v", parsed.Status, original.Status)
	}
	if parsed.Initiative != original.Initiative {
		t.Errorf("Initiative mismatch: got %q, want %q", parsed.Initiative, original.Initiative)
	}
}

func TestNewContext(t *testing.T) {
	ctx := NewContext("Test Initiative", "MODULE", "10x-dev")

	// Check required fields
	if ctx.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if !IsValidSessionID(ctx.SessionID) {
		t.Errorf("SessionID %q is not valid", ctx.SessionID)
	}
	if ctx.SchemaVersion != "2.1" {
		t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "2.1")
	}
	if ctx.Status != StatusActive {
		t.Errorf("Status = %v, want %v", ctx.Status, StatusActive)
	}
	if ctx.Initiative != "Test Initiative" {
		t.Errorf("Initiative = %q, want %q", ctx.Initiative, "Test Initiative")
	}
	if ctx.Complexity != "MODULE" {
		t.Errorf("Complexity = %q, want %q", ctx.Complexity, "MODULE")
	}
	if ctx.ActiveRite != "10x-dev" {
		t.Errorf("ActiveTeam = %q, want %q", ctx.ActiveRite, "10x-dev")
	}
	if ctx.CurrentPhase != "requirements" {
		t.Errorf("CurrentPhase = %q, want %q", ctx.CurrentPhase, "requirements")
	}
}

func TestContext_Validate(t *testing.T) {
	// Valid context
	ctx := NewContext("Test", "MODULE", "test-pack")
	issues := ctx.Validate()
	if len(issues) > 0 {
		t.Errorf("Validate() returned issues for valid context: %v", issues)
	}

	// Invalid context
	invalid := &Context{
		SessionID:    "invalid-id",
		Status:       Status("INVALID"),
		Initiative:   "Test",
		Complexity:   "MODULE",
		ActiveRite:   "test",
		CurrentPhase: "design",
	}
	issues = invalid.Validate()
	if len(issues) == 0 {
		t.Error("Validate() should return issues for invalid context")
	}
}

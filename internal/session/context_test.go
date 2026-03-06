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
		t.Errorf("ActiveRite = %q, want %q", ctx.ActiveRite, "10x-dev")
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
		ActiveRite:    "test-rite",
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
		ActiveRite:    "test-rite",
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
	if ctx.SchemaVersion != "2.3" {
		t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "2.3")
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
		t.Errorf("ActiveRite = %q, want %q", ctx.ActiveRite, "10x-dev")
	}
	if ctx.CurrentPhase != "requirements" {
		t.Errorf("CurrentPhase = %q, want %q", ctx.CurrentPhase, "requirements")
	}
}

func TestContext_Validate(t *testing.T) {
	// Valid context
	ctx := NewContext("Test", "MODULE", "test-rite")
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

func TestParseContext_NormalizesPhantomStatus(t *testing.T) {
	// "COMPLETED" is a phantom status — not in the FSM, but written by
	// legacy scripts. ParseContext should normalize it to ARCHIVED.
	content := `---
schema_version: "2.1"
session_id: "session-20260104-160414-563c681e"
status: "COMPLETED"
created_at: "2026-01-04T16:04:14Z"
initiative: "Phantom status test"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "complete"
---
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if ctx.Status != StatusArchived {
		t.Errorf("Status = %q, want %q (COMPLETED should normalize to ARCHIVED)", ctx.Status, StatusArchived)
	}
}

func TestParseContext_NormalizesComplete(t *testing.T) {
	content := `---
schema_version: "2.1"
session_id: "session-20260104-160414-563c681e"
status: "COMPLETE"
created_at: "2026-01-04T16:04:14Z"
initiative: "Phantom status test"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "complete"
---
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if ctx.Status != StatusArchived {
		t.Errorf("Status = %q, want %q (COMPLETE should normalize to ARCHIVED)", ctx.Status, StatusArchived)
	}
}

func TestContext_FrayFields_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	original := &Context{
		SchemaVersion: "2.3",
		SessionID:     "session-20260206-120000-abcdef01",
		Status:        StatusActive,
		CreatedAt:     now,
		Initiative:    "Fray Test",
		Complexity:    "MODULE",
		ActiveRite:    "test-rite",
		CurrentPhase:  "design",
		FrayedFrom:    "session-20260101-120000-abcdef01",
		FrayPoint:     "design",
		Strands:       []Strand{{SessionID: "session-20260101-130000-bcdef012", Status: "ACTIVE"}},
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

	// Verify fray fields survived the round trip
	if parsed.FrayedFrom != original.FrayedFrom {
		t.Errorf("FrayedFrom mismatch: got %q, want %q", parsed.FrayedFrom, original.FrayedFrom)
	}
	if parsed.FrayPoint != original.FrayPoint {
		t.Errorf("FrayPoint mismatch: got %q, want %q", parsed.FrayPoint, original.FrayPoint)
	}
	if len(parsed.Strands) != len(original.Strands) {
		t.Fatalf("Strands length mismatch: got %d, want %d", len(parsed.Strands), len(original.Strands))
	}
	if parsed.Strands[0].SessionID != original.Strands[0].SessionID {
		t.Errorf("Strands[0].SessionID mismatch: got %q, want %q", parsed.Strands[0].SessionID, original.Strands[0].SessionID)
	}
	if parsed.Strands[0].Status != original.Strands[0].Status {
		t.Errorf("Strands[0].Status mismatch: got %q, want %q", parsed.Strands[0].Status, original.Strands[0].Status)
	}
}

func TestContext_FrayFields_Optional(t *testing.T) {
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

	// Verify fray fields are empty (backward compatibility)
	if ctx.FrayedFrom != "" {
		t.Errorf("FrayedFrom should be empty, got %q", ctx.FrayedFrom)
	}
	if ctx.FrayPoint != "" {
		t.Errorf("FrayPoint should be empty, got %q", ctx.FrayPoint)
	}
	if ctx.Strands != nil {
		t.Errorf("Strands should be nil, got %v", ctx.Strands)
	}
}

func TestContext_FrayFields_StrandsArray(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	original := &Context{
		SchemaVersion: "2.3",
		SessionID:     "session-20260206-120000-abcdef01",
		Status:        StatusActive,
		CreatedAt:     now,
		Initiative:    "Strands Test",
		Complexity:    "MODULE",
		ActiveRite:    "test-rite",
		CurrentPhase:  "design",
		Strands: []Strand{
			{SessionID: "session-20260201-100000-aaa11111", Status: "ACTIVE"},
			{SessionID: "session-20260202-110000-bbb22222", Status: "LANDED", LandedAt: "2026-02-03T10:00:00Z"},
		},
		Body: "\n# Test\n",
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

	// Verify both strands are present and in order
	if len(parsed.Strands) != 2 {
		t.Fatalf("Strands length mismatch: got %d, want 2", len(parsed.Strands))
	}
	if parsed.Strands[0].SessionID != "session-20260201-100000-aaa11111" {
		t.Errorf("Strands[0].SessionID mismatch: got %q", parsed.Strands[0].SessionID)
	}
	if parsed.Strands[0].Status != "ACTIVE" {
		t.Errorf("Strands[0].Status mismatch: got %q", parsed.Strands[0].Status)
	}
	if parsed.Strands[1].SessionID != "session-20260202-110000-bbb22222" {
		t.Errorf("Strands[1].SessionID mismatch: got %q", parsed.Strands[1].SessionID)
	}
	if parsed.Strands[1].Status != "LANDED" {
		t.Errorf("Strands[1].Status mismatch: got %q", parsed.Strands[1].Status)
	}
	if parsed.Strands[1].LandedAt != "2026-02-03T10:00:00Z" {
		t.Errorf("Strands[1].LandedAt mismatch: got %q", parsed.Strands[1].LandedAt)
	}
}

func TestParseContext_StrandStructRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	original := &Context{
		SchemaVersion: "2.3",
		SessionID:     "session-20260306-120000-abcdef01",
		Status:        StatusActive,
		CreatedAt:     now,
		Initiative:    "Strand struct round trip",
		Complexity:    "MODULE",
		ActiveRite:    "test-rite",
		CurrentPhase:  "implementation",
		Strands: []Strand{
			{SessionID: "session-20260306-130000-aaa11111", Status: "ACTIVE"},
			{SessionID: "session-20260306-140000-bbb22222", Status: "LANDED", FrameRef: "frames/auth.md", LandedAt: "2026-03-06T15:00:00Z"},
		},
		Body: "\n# Test\n",
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	parsed, err := ParseContext(data)
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if len(parsed.Strands) != 2 {
		t.Fatalf("Strands length = %d, want 2", len(parsed.Strands))
	}

	// Verify all Strand fields preserved
	s0 := parsed.Strands[0]
	if s0.SessionID != "session-20260306-130000-aaa11111" {
		t.Errorf("Strands[0].SessionID = %q", s0.SessionID)
	}
	if s0.Status != "ACTIVE" {
		t.Errorf("Strands[0].Status = %q", s0.Status)
	}
	if s0.FrameRef != "" {
		t.Errorf("Strands[0].FrameRef should be empty, got %q", s0.FrameRef)
	}
	if s0.LandedAt != "" {
		t.Errorf("Strands[0].LandedAt should be empty, got %q", s0.LandedAt)
	}

	s1 := parsed.Strands[1]
	if s1.SessionID != "session-20260306-140000-bbb22222" {
		t.Errorf("Strands[1].SessionID = %q", s1.SessionID)
	}
	if s1.Status != "LANDED" {
		t.Errorf("Strands[1].Status = %q", s1.Status)
	}
	if s1.FrameRef != "frames/auth.md" {
		t.Errorf("Strands[1].FrameRef = %q, want %q", s1.FrameRef, "frames/auth.md")
	}
	if s1.LandedAt != "2026-03-06T15:00:00Z" {
		t.Errorf("Strands[1].LandedAt = %q", s1.LandedAt)
	}
}

func TestParseContext_StrandPolymorphic_OldFormat(t *testing.T) {
	content := `---
schema_version: "2.1"
session_id: "session-20260104-160414-563c681e"
status: "ACTIVE"
created_at: "2026-01-04T16:04:14Z"
initiative: "Old format strands"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "design"
strands:
  - session-20260104-170000-abc12345
  - session-20260104-180000-def67890
---
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if len(ctx.Strands) != 2 {
		t.Fatalf("Strands length = %d, want 2", len(ctx.Strands))
	}
	if ctx.Strands[0].SessionID != "session-20260104-170000-abc12345" {
		t.Errorf("Strands[0].SessionID = %q", ctx.Strands[0].SessionID)
	}
	if ctx.Strands[0].Status != "ACTIVE" {
		t.Errorf("Strands[0].Status = %q, want %q (auto-converted)", ctx.Strands[0].Status, "ACTIVE")
	}
	if ctx.Strands[1].SessionID != "session-20260104-180000-def67890" {
		t.Errorf("Strands[1].SessionID = %q", ctx.Strands[1].SessionID)
	}
	if ctx.Strands[1].Status != "ACTIVE" {
		t.Errorf("Strands[1].Status = %q, want %q", ctx.Strands[1].Status, "ACTIVE")
	}
}

func TestParseContext_StrandPolymorphic_NewFormat(t *testing.T) {
	content := `---
schema_version: "2.3"
session_id: "session-20260306-120000-abcdef01"
status: "ACTIVE"
created_at: "2026-03-06T12:00:00Z"
initiative: "New format strands"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "design"
strands:
  - session_id: session-20260306-130000-aaa11111
    status: ACTIVE
  - session_id: session-20260306-140000-bbb22222
    status: LANDED
    frame_ref: frames/auth.md
    landed_at: "2026-03-06T15:00:00Z"
---
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if len(ctx.Strands) != 2 {
		t.Fatalf("Strands length = %d, want 2", len(ctx.Strands))
	}
	if ctx.Strands[0].SessionID != "session-20260306-130000-aaa11111" {
		t.Errorf("Strands[0].SessionID = %q", ctx.Strands[0].SessionID)
	}
	if ctx.Strands[0].Status != "ACTIVE" {
		t.Errorf("Strands[0].Status = %q", ctx.Strands[0].Status)
	}
	if ctx.Strands[1].SessionID != "session-20260306-140000-bbb22222" {
		t.Errorf("Strands[1].SessionID = %q", ctx.Strands[1].SessionID)
	}
	if ctx.Strands[1].Status != "LANDED" {
		t.Errorf("Strands[1].Status = %q", ctx.Strands[1].Status)
	}
	if ctx.Strands[1].FrameRef != "frames/auth.md" {
		t.Errorf("Strands[1].FrameRef = %q", ctx.Strands[1].FrameRef)
	}
	if ctx.Strands[1].LandedAt != "2026-03-06T15:00:00Z" {
		t.Errorf("Strands[1].LandedAt = %q", ctx.Strands[1].LandedAt)
	}
}

func TestParseContext_StrandPolymorphic_Empty(t *testing.T) {
	content := `---
schema_version: "2.3"
session_id: "session-20260306-120000-abcdef01"
status: "ACTIVE"
created_at: "2026-03-06T12:00:00Z"
initiative: "Empty strands"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "design"
strands: []
---
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if len(ctx.Strands) != 0 {
		t.Errorf("Strands length = %d, want 0", len(ctx.Strands))
	}
}

func TestParseContext_NewFieldsRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	original := &Context{
		SchemaVersion: "2.3",
		SessionID:     "session-20260306-120000-abcdef01",
		Status:        StatusActive,
		CreatedAt:     now,
		Initiative:    "New fields round trip",
		Complexity:    "MODULE",
		ActiveRite:    "test-rite",
		CurrentPhase:  "implementation",
		FrameRef:      ".sos/wip/frames/auth-flow.md",
		ClaimedBy:     "cc-session-abc123",
		ParkSource:    "fray",
		Body:          "\n# Test\n",
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	parsed, err := ParseContext(data)
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if parsed.FrameRef != original.FrameRef {
		t.Errorf("FrameRef = %q, want %q", parsed.FrameRef, original.FrameRef)
	}
	if parsed.ClaimedBy != original.ClaimedBy {
		t.Errorf("ClaimedBy = %q, want %q", parsed.ClaimedBy, original.ClaimedBy)
	}
	if parsed.ParkSource != original.ParkSource {
		t.Errorf("ParkSource = %q, want %q", parsed.ParkSource, original.ParkSource)
	}
}

func TestParseContext_NewFieldsOptional(t *testing.T) {
	content := `---
schema_version: "2.1"
session_id: "session-20260104-160414-563c681e"
status: "ACTIVE"
created_at: "2026-01-04T16:04:14Z"
initiative: "V2.1 session without new fields"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "design"
---
`

	ctx, err := ParseContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseContext() error = %v", err)
	}

	if ctx.FrameRef != "" {
		t.Errorf("FrameRef should be empty, got %q", ctx.FrameRef)
	}
	if ctx.ClaimedBy != "" {
		t.Errorf("ClaimedBy should be empty, got %q", ctx.ClaimedBy)
	}
	if ctx.ParkSource != "" {
		t.Errorf("ParkSource should be empty, got %q", ctx.ParkSource)
	}
}

func TestContext_FindStrand(t *testing.T) {
	ctx := &Context{
		Strands: []Strand{
			{SessionID: "session-aaa", Status: "ACTIVE"},
			{SessionID: "session-bbb", Status: "LANDED"},
		},
	}

	// Found case
	s := ctx.FindStrand("session-bbb")
	if s == nil {
		t.Fatal("FindStrand should find session-bbb")
	}
	if s.Status != "LANDED" {
		t.Errorf("found strand Status = %q, want %q", s.Status, "LANDED")
	}

	// Not found case
	s = ctx.FindStrand("session-nonexistent")
	if s != nil {
		t.Errorf("FindStrand should return nil for nonexistent ID, got %+v", s)
	}
}

func TestContext_FindStrand_MutablePointer(t *testing.T) {
	ctx := &Context{
		Strands: []Strand{
			{SessionID: "session-aaa", Status: "ACTIVE"},
		},
	}

	s := ctx.FindStrand("session-aaa")
	if s == nil {
		t.Fatal("FindStrand should find session-aaa")
	}

	// Mutate via pointer
	s.Status = "LANDED"
	s.LandedAt = "2026-03-06T15:00:00Z"

	// Verify mutation is reflected in original
	if ctx.Strands[0].Status != "LANDED" {
		t.Errorf("mutation not reflected: Status = %q, want %q", ctx.Strands[0].Status, "LANDED")
	}
	if ctx.Strands[0].LandedAt != "2026-03-06T15:00:00Z" {
		t.Errorf("mutation not reflected: LandedAt = %q", ctx.Strands[0].LandedAt)
	}
}

func TestNewContext_SchemaVersion23(t *testing.T) {
	ctx := NewContext("Test", "PATCH", "10x-dev")
	if ctx.SchemaVersion != "2.3" {
		t.Errorf("NewContext SchemaVersion = %q, want %q", ctx.SchemaVersion, "2.3")
	}
}

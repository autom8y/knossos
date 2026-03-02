package tribute

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractor_ExtractArtifacts(t *testing.T) {
	events := []EventData{
		{
			Ts:           "2026-01-06T11:00:00Z",
			Type:         "tool.artifact_created",
			Path:         ".ledge/specs/PRD-test.md",
			ArtifactType: "PRD",
		},
		{
			Ts:   "2026-01-06T12:00:00Z",
			Type: "tool.artifact_created",
			Path: ".ledge/specs/TDD-test.md",
			// ArtifactType inferred from path
		},
		{
			Ts:   "2026-01-06T13:00:00Z",
			Type: "tool.call", // Not an artifact event
		},
	}

	extractor := NewExtractor("/tmp")
	artifacts := extractor.ExtractArtifacts(events)

	if len(artifacts) != 2 {
		t.Fatalf("ExtractArtifacts() = %d artifacts, want 2", len(artifacts))
	}

	// First artifact
	if artifacts[0].Type != "PRD" {
		t.Errorf("artifacts[0].Type = %q, want %q", artifacts[0].Type, "PRD")
	}
	if artifacts[0].Path != ".ledge/specs/PRD-test.md" {
		t.Errorf("artifacts[0].Path = %q, want %q", artifacts[0].Path, ".ledge/specs/PRD-test.md")
	}

	// Second artifact - type inferred
	if artifacts[1].Type != "TDD" {
		t.Errorf("artifacts[1].Type = %q, want %q", artifacts[1].Type, "TDD")
	}
}

func TestExtractor_ExtractDecisions(t *testing.T) {
	events := []EventData{
		{
			Ts:        "2026-01-06T11:00:00Z",
			Type:      "agent.decision",
			Decision:  "Use PostgreSQL over MySQL",
			Rationale: "Better JSON support and performance",
			Rejected:  []string{"MySQL", "SQLite"},
		},
		{
			Ts:   "2026-01-06T12:00:00Z",
			Type: "tool.call", // Not a decision event
		},
	}

	extractor := NewExtractor("/tmp")
	decisions := extractor.ExtractDecisions(events)

	if len(decisions) != 1 {
		t.Fatalf("ExtractDecisions() = %d decisions, want 1", len(decisions))
	}

	if decisions[0].Decision != "Use PostgreSQL over MySQL" {
		t.Errorf("Decision = %q, want %q", decisions[0].Decision, "Use PostgreSQL over MySQL")
	}
	if decisions[0].Rationale != "Better JSON support and performance" {
		t.Errorf("Rationale = %q", decisions[0].Rationale)
	}
	if len(decisions[0].Rejected) != 2 {
		t.Errorf("Rejected count = %d, want 2", len(decisions[0].Rejected))
	}
}

func TestExtractor_ExtractHandoffs(t *testing.T) {
	events := []EventData{
		{
			Ts:    "2026-01-06T11:00:00Z",
			Type:  "agent.handoff_prepared",
			From:  "architect",
			To:    "principal-engineer",
			Notes: "TDD approved",
		},
		{
			Ts:   "2026-01-06T11:00:01Z",
			Type: "agent.handoff_executed",
			From: "architect",
			To:   "principal-engineer",
		},
	}

	extractor := NewExtractor("/tmp")
	handoffs := extractor.ExtractHandoffs(events)

	if len(handoffs) != 1 {
		t.Fatalf("ExtractHandoffs() = %d handoffs, want 1", len(handoffs))
	}

	if handoffs[0].From != "architect" {
		t.Errorf("From = %q, want %q", handoffs[0].From, "architect")
	}
	if handoffs[0].To != "principal-engineer" {
		t.Errorf("To = %q, want %q", handoffs[0].To, "principal-engineer")
	}
	if handoffs[0].Notes != "TDD approved" {
		t.Errorf("Notes = %q, want %q", handoffs[0].Notes, "TDD approved")
	}
}

func TestExtractor_ExtractPhases(t *testing.T) {
	events := []EventData{
		{
			Timestamp: "2026-01-06T10:00:00Z",
			Event:     "PHASE_TRANSITIONED",
			FromPhase: "",
			ToPhase:   "requirements",
			Metadata:  map[string]interface{}{"agent": "requirements-analyst"},
		},
		{
			Timestamp: "2026-01-06T12:00:00Z",
			Event:     "PHASE_TRANSITIONED",
			FromPhase: "requirements",
			ToPhase:   "design",
			Metadata:  map[string]interface{}{"agent": "architect"},
		},
	}

	extractor := NewExtractor("/tmp")
	phases := extractor.ExtractPhases(events)

	if len(phases) != 2 {
		t.Fatalf("ExtractPhases() = %d phases, want 2", len(phases))
	}

	// First phase
	if phases[0].Phase != "requirements" {
		t.Errorf("phases[0].Phase = %q, want %q", phases[0].Phase, "requirements")
	}
	if phases[0].Agent != "requirements-analyst" {
		t.Errorf("phases[0].Agent = %q, want %q", phases[0].Agent, "requirements-analyst")
	}
	// Duration should be 2 hours
	expectedDuration := 2 * time.Hour
	if phases[0].Duration != expectedDuration {
		t.Errorf("phases[0].Duration = %v, want %v", phases[0].Duration, expectedDuration)
	}

	// Second phase
	if phases[1].Phase != "design" {
		t.Errorf("phases[1].Phase = %q, want %q", phases[1].Phase, "design")
	}
}

func TestExtractor_ExtractMetrics(t *testing.T) {
	events := []EventData{
		{Type: "tool.call"},
		{Type: "tool.call"},
		{Type: "tool.call"},
		{
			Type: "tool.file_change",
			Meta: map[string]interface{}{
				"lines_added":   float64(100),
				"lines_removed": float64(20),
			},
		},
		{
			Type: "tool.file_change",
			Meta: map[string]interface{}{
				"lines_added":   float64(50),
				"lines_removed": float64(10),
			},
		},
	}

	extractor := NewExtractor("/tmp")
	metrics := extractor.ExtractMetrics(events)

	if metrics.ToolCalls != 3 {
		t.Errorf("ToolCalls = %d, want 3", metrics.ToolCalls)
	}
	if metrics.FilesModified != 2 {
		t.Errorf("FilesModified = %d, want 2", metrics.FilesModified)
	}
	if metrics.LinesAdded != 150 {
		t.Errorf("LinesAdded = %d, want 150", metrics.LinesAdded)
	}
	if metrics.LinesRemoved != 30 {
		t.Errorf("LinesRemoved = %d, want 30", metrics.LinesRemoved)
	}
	if metrics.EventsRecorded != 5 {
		t.Errorf("EventsRecorded = %d, want 5", metrics.EventsRecorded)
	}
}

func TestExtractor_ExtractWhiteSails(t *testing.T) {
	tmpDir := t.TempDir()

	// Create WHITE_SAILS.yaml
	sailsContent := `schema_version: "1.0"
session_id: test-session
color: WHITE
computed_base: GRAY
proofs:
  tests:
    status: PASS
    summary: All tests passed
  build:
    status: PASS
    summary: Build successful
`
	if err := os.WriteFile(filepath.Join(tmpDir, "WHITE_SAILS.yaml"), []byte(sailsContent), 0644); err != nil {
		t.Fatalf("failed to write WHITE_SAILS.yaml: %v", err)
	}

	extractor := NewExtractor(tmpDir)
	sails, err := extractor.ExtractWhiteSails()
	if err != nil {
		t.Fatalf("ExtractWhiteSails() error: %v", err)
	}

	if sails == nil {
		t.Fatal("ExtractWhiteSails() returned nil")
	}
	if sails.Color != "WHITE" {
		t.Errorf("Color = %q, want %q", sails.Color, "WHITE")
	}
	if sails.ComputedBase != "GRAY" {
		t.Errorf("ComputedBase = %q, want %q", sails.ComputedBase, "GRAY")
	}
	if len(sails.Proofs) != 2 {
		t.Errorf("Proofs count = %d, want 2", len(sails.Proofs))
	}
	if sails.Proofs["tests"].Status != "PASS" {
		t.Errorf("tests proof status = %q, want %q", sails.Proofs["tests"].Status, "PASS")
	}
}

func TestExtractor_ExtractWhiteSails_Missing(t *testing.T) {
	tmpDir := t.TempDir()

	extractor := NewExtractor(tmpDir)
	sails, err := extractor.ExtractWhiteSails()

	// Should gracefully return nil without error
	if err != nil {
		t.Errorf("ExtractWhiteSails() error = %v, want nil", err)
	}
	if sails != nil {
		t.Errorf("ExtractWhiteSails() = %v, want nil for missing file", sails)
	}
}

func TestExtractor_ExtractNotes_Boilerplate(t *testing.T) {
	extractor := NewExtractor("/tmp")

	// Default boilerplate should return empty
	boilerplate := `
# Session: test

## Artifacts
- PRD: pending
- TDD: pending

## Blockers
None yet.

## Next Steps
1. Complete requirements gathering
`
	notes := extractor.ExtractNotes(boilerplate)
	if notes != "" {
		t.Errorf("ExtractNotes(boilerplate) = %q, want empty string", notes)
	}
}

func TestExtractor_ExtractNotes_Custom(t *testing.T) {
	extractor := NewExtractor("/tmp")

	// Custom content should be preserved
	custom := `
# Session: test

## Implementation Notes

This session implemented a complex caching layer with the following design decisions:
- LRU eviction policy chosen for simplicity
- Redis selected as backing store for horizontal scaling
`
	notes := extractor.ExtractNotes(custom)
	if notes == "" {
		t.Error("ExtractNotes(custom) should return non-empty string")
	}
}

func TestInferArtifactType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{".ledge/specs/PRD-feature.md", "PRD"},
		{".ledge/specs/TDD-feature.md", "TDD"},
		{".ledge/decisions/ADR-0001-approach.md", "ADR"},
		{"src/feature_test.go", "Tests"},
		{"tests/integration/test_api.py", "Tests"},
		{"src/main.go", "Code"},
		{"lib/utils.js", "Code"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := inferArtifactType(tt.path)
			if result != tt.expected {
				t.Errorf("inferArtifactType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2026-01-06T10:00:00Z", time.Date(2026, 1, 6, 10, 0, 0, 0, time.UTC)},
		{"2026-01-06T10:00:00.123Z", time.Date(2026, 1, 6, 10, 0, 0, 123000000, time.UTC)},
		{"", time.Time{}},
		{"invalid", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseTimestamp(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("parseTimestamp(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEventData_GetTimestamp(t *testing.T) {
	// Test with Timestamp field
	e1 := EventData{Timestamp: "2026-01-06T10:00:00Z"}
	if e1.GetTimestamp() != "2026-01-06T10:00:00Z" {
		t.Errorf("GetTimestamp() with Timestamp field failed")
	}

	// Test with Ts field
	e2 := EventData{Ts: "2026-01-06T11:00:00Z"}
	if e2.GetTimestamp() != "2026-01-06T11:00:00Z" {
		t.Errorf("GetTimestamp() with Ts field failed")
	}

	// Test with both (Timestamp takes precedence)
	e3 := EventData{Timestamp: "2026-01-06T10:00:00Z", Ts: "2026-01-06T11:00:00Z"}
	if e3.GetTimestamp() != "2026-01-06T10:00:00Z" {
		t.Errorf("GetTimestamp() with both fields should prefer Timestamp")
	}
}

func TestEventData_GetEventType(t *testing.T) {
	// Test with Event field
	e1 := EventData{Event: "SESSION_CREATED"}
	if e1.GetEventType() != "SESSION_CREATED" {
		t.Errorf("GetEventType() with Event field failed")
	}

	// Test with Type field
	e2 := EventData{Type: "tool.artifact_created"}
	if e2.GetEventType() != "tool.artifact_created" {
		t.Errorf("GetEventType() with Type field failed")
	}
}

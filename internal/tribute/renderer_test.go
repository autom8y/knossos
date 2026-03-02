package tribute

import (
	"strings"
	"testing"
	"time"
)

func TestRenderer_Render(t *testing.T) {
	result := &GenerateResult{
		SessionID:   "session-test-123",
		Initiative:  "Test Initiative",
		Complexity:  "MODULE",
		Duration:    4*time.Hour + 30*time.Minute,
		Rite:        "10x-dev",
		FinalPhase:  "validation",
		StartedAt:   time.Date(2026, 1, 6, 10, 0, 0, 0, time.UTC),
		EndedAt:     time.Date(2026, 1, 6, 14, 30, 0, 0, time.UTC),
		GeneratedAt: time.Date(2026, 1, 6, 14, 30, 0, 0, time.UTC),
		Artifacts: []Artifact{
			{Type: "PRD", Path: ".ledge/specs/PRD-test.md", Status: "Approved"},
		},
		Decisions: []Decision{
			{
				Timestamp: time.Date(2026, 1, 6, 11, 0, 0, 0, time.UTC),
				Decision:  "Use Go",
				Rationale: "Type safety",
			},
		},
		SailsColor: "WHITE",
		SailsBase:  "WHITE",
		SailsProofs: map[string]ProofStatus{
			"tests": {Status: "PASS", Summary: "All passed"},
		},
		Metrics: Metrics{
			ToolCalls:      100,
			EventsRecorded: 50,
			FilesModified:  10,
		},
	}

	renderer := NewRenderer()
	content, err := renderer.Render(result)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	contentStr := string(content)

	// Check frontmatter
	if !strings.Contains(contentStr, "schema_version: \"1.0\"") {
		t.Error("Missing schema_version in frontmatter")
	}
	if !strings.Contains(contentStr, "session_id: session-test-123") {
		t.Error("Missing session_id in frontmatter")
	}

	// Check title
	if !strings.Contains(contentStr, "# Tribute: Test Initiative") {
		t.Error("Missing title")
	}

	// Check summary section
	if !strings.Contains(contentStr, "## Summary") {
		t.Error("Missing Summary section")
	}
	if !strings.Contains(contentStr, "**Initiative**: Test Initiative") {
		t.Error("Missing initiative in summary")
	}
	if !strings.Contains(contentStr, "**Complexity**: MODULE") {
		t.Error("Missing complexity in summary")
	}

	// Check artifacts section
	if !strings.Contains(contentStr, "## Artifacts Produced") {
		t.Error("Missing Artifacts Produced section")
	}
	if !strings.Contains(contentStr, "PRD") {
		t.Error("Missing PRD artifact")
	}

	// Check decisions section
	if !strings.Contains(contentStr, "## Decisions Made") {
		t.Error("Missing Decisions Made section")
	}

	// Check sails section
	if !strings.Contains(contentStr, "## White Sails Attestation") {
		t.Error("Missing White Sails Attestation section")
	}
	if !strings.Contains(contentStr, "**Color**: WHITE") {
		t.Error("Missing sails color")
	}

	// Check metrics section
	if !strings.Contains(contentStr, "## Metrics") {
		t.Error("Missing Metrics section")
	}
	if !strings.Contains(contentStr, "| Tool Calls | 100 |") {
		t.Error("Missing tool calls metric")
	}
}

func TestRenderer_RenderConditionalSections(t *testing.T) {
	// Minimal result without optional data
	result := &GenerateResult{
		SessionID:   "session-minimal",
		Initiative:  "Minimal Test",
		Complexity:  "SCRIPT",
		GeneratedAt: time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC),
		// No Decisions, Phases, Handoffs, Commits, SailsColor
	}

	renderer := NewRenderer()
	content, err := renderer.Render(result)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	contentStr := string(content)

	// Should NOT have conditional sections
	if strings.Contains(contentStr, "## Decisions Made") {
		t.Error("Should not have Decisions Made section when no decisions")
	}
	if strings.Contains(contentStr, "## Phase Progression") {
		t.Error("Should not have Phase Progression section when no phases")
	}
	if strings.Contains(contentStr, "## Handoffs") {
		t.Error("Should not have Handoffs section when no handoffs")
	}
	if strings.Contains(contentStr, "## Git Commits") {
		t.Error("Should not have Git Commits section when no commits")
	}
	if strings.Contains(contentStr, "## White Sails Attestation") {
		t.Error("Should not have White Sails section when no sails color")
	}

	// Should still have required sections
	if !strings.Contains(contentStr, "## Summary") {
		t.Error("Missing required Summary section")
	}
	if !strings.Contains(contentStr, "## Artifacts Produced") {
		t.Error("Missing required Artifacts Produced section")
	}
	if !strings.Contains(contentStr, "## Metrics") {
		t.Error("Missing required Metrics section")
	}
}

func TestRenderer_RenderPhases(t *testing.T) {
	result := &GenerateResult{
		SessionID:   "session-phases",
		Initiative:  "Phase Test",
		Complexity:  "MODULE",
		GeneratedAt: time.Date(2026, 1, 6, 14, 0, 0, 0, time.UTC),
		Phases: []PhaseRecord{
			{Phase: "requirements", StartedAt: time.Date(2026, 1, 6, 10, 0, 0, 0, time.UTC), Duration: 2 * time.Hour, Agent: "requirements-analyst"},
			{Phase: "design", StartedAt: time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC), Duration: 1 * time.Hour, Agent: "architect"},
			{Phase: "implementation", StartedAt: time.Date(2026, 1, 6, 13, 0, 0, 0, time.UTC), Duration: 1 * time.Hour, Agent: "principal-engineer"},
		},
	}

	renderer := NewRenderer()
	content, err := renderer.Render(result)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	contentStr := string(content)

	// Check phase progression section
	if !strings.Contains(contentStr, "## Phase Progression") {
		t.Error("Missing Phase Progression section")
	}

	// Check ASCII timeline
	if !strings.Contains(contentStr, "requirements") {
		t.Error("Missing requirements phase in timeline")
	}
	if !strings.Contains(contentStr, "design") {
		t.Error("Missing design phase in timeline")
	}
	if !strings.Contains(contentStr, "implementation") {
		t.Error("Missing implementation phase in timeline")
	}

	// Check phase table
	if !strings.Contains(contentStr, "| Phase | Started | Duration | Agent |") {
		t.Error("Missing phase table header")
	}
	if !strings.Contains(contentStr, "requirements-analyst") {
		t.Error("Missing requirements-analyst agent")
	}
}

func TestRenderer_RenderHandoffs(t *testing.T) {
	result := &GenerateResult{
		SessionID:   "session-handoffs",
		Initiative:  "Handoff Test",
		Complexity:  "MODULE",
		GeneratedAt: time.Date(2026, 1, 6, 14, 0, 0, 0, time.UTC),
		Handoffs: []Handoff{
			{From: "architect", To: "principal-engineer", Timestamp: time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC), Notes: "TDD approved"},
		},
	}

	renderer := NewRenderer()
	content, err := renderer.Render(result)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	contentStr := string(content)

	// Check handoffs section
	if !strings.Contains(contentStr, "## Handoffs") {
		t.Error("Missing Handoffs section")
	}
	if !strings.Contains(contentStr, "| From | To | Timestamp | Notes |") {
		t.Error("Missing handoffs table header")
	}
	if !strings.Contains(contentStr, "architect") {
		t.Error("Missing architect in handoffs")
	}
	if !strings.Contains(contentStr, "principal-engineer") {
		t.Error("Missing principal-engineer in handoffs")
	}
}

func TestRenderer_RenderNotes(t *testing.T) {
	result := &GenerateResult{
		SessionID:   "session-notes",
		Initiative:  "Notes Test",
		Complexity:  "MODULE",
		GeneratedAt: time.Date(2026, 1, 6, 14, 0, 0, 0, time.UTC),
		Notes:       "Custom implementation notes here.",
	}

	renderer := NewRenderer()
	content, err := renderer.Render(result)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	contentStr := string(content)

	// Check notes section
	if !strings.Contains(contentStr, "## Notes") {
		t.Error("Missing Notes section")
	}
	if !strings.Contains(contentStr, "Custom implementation notes here.") {
		t.Error("Missing notes content")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0m"},
		{30 * time.Minute, "30m"},
		{1 * time.Hour, "1h 00m"},
		{1*time.Hour + 30*time.Minute, "1h 30m"},
		{4*time.Hour + 45*time.Minute, "4h 45m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestComplexityDescription(t *testing.T) {
	tests := []struct {
		complexity string
		expected   string
	}{
		{"SCRIPT", " (estimated 1-2 hours)"},
		{"MODULE", " (estimated 4-8 hours)"},
		{"SERVICE", " (estimated 1-2 days)"},
		{"SYSTEM", " (estimated 1+ weeks)"},
		{"UNKNOWN", ""},
	}

	for _, tt := range tests {
		t.Run(tt.complexity, func(t *testing.T) {
			result := complexityDescription(tt.complexity)
			if result != tt.expected {
				t.Errorf("complexityDescription(%q) = %q, want %q", tt.complexity, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

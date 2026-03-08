package suggest

import (
	"testing"
)

func TestSessionStartSuggestions(t *testing.T) {
	tests := []struct {
		name      string
		input     *SessionInput
		wantLen   int
		wantKind  Kind
		wantTexts []string // substrings to check in Text fields
	}{
		{
			name:    "nil input",
			input:   nil,
			wantLen: 0,
		},
		{
			name:    "no session (empty session ID)",
			input:   &SessionInput{},
			wantLen: 0,
		},
		{
			name: "task complexity with initiative",
			input: &SessionInput{
				SessionID:  "session-123",
				Initiative: "Fix bug",
				Complexity: "TASK",
			},
			wantLen:   1,
			wantKind:  KindSessionStart,
			wantTexts: []string{"Fix bug", "/task"},
		},
		{
			name: "module complexity with initiative",
			input: &SessionInput{
				SessionID:  "session-123",
				Initiative: "Add feature",
				Complexity: "MODULE",
			},
			wantLen:   1,
			wantKind:  KindSessionStart,
			wantTexts: []string{"Add feature", "/sprint"},
		},
		{
			name: "initiative complexity",
			input: &SessionInput{
				SessionID:  "session-123",
				Initiative: "Big thing",
				Complexity: "INITIATIVE",
			},
			wantLen:   1,
			wantKind:  KindSessionStart,
			wantTexts: []string{"Big thing", "/10x"},
		},
		{
			name: "resumed from park with complexity",
			input: &SessionInput{
				SessionID:  "session-123",
				Initiative: "X",
				Complexity: "TASK",
				ParkSource: "auto",
			},
			wantLen:   2,
			wantKind:  KindSessionStart,
			wantTexts: []string{"Resuming parked session"},
		},
		{
			name: "resumed from park without initiative",
			input: &SessionInput{
				SessionID:  "session-123",
				ParkSource: "auto",
			},
			wantLen:   1,
			wantKind:  KindSessionStart,
			wantTexts: []string{"Resuming parked session"},
		},
		{
			name: "with strands and no initiative",
			input: &SessionInput{
				SessionID:   "session-123",
				StrandCount: 3,
			},
			wantLen:   1,
			wantKind:  KindSessionStart,
			wantTexts: []string{"3 strand(s)", "/fray"},
		},
		{
			name: "strands capped at max suggestions",
			input: &SessionInput{
				SessionID:   "session-123",
				ParkSource:  "manual",
				Initiative:  "Something",
				Complexity:  "TASK",
				StrandCount: 2,
			},
			wantLen: 2, // park + initiative, strands dropped due to cap
		},
		{
			name: "session ID only (no enrichment)",
			input: &SessionInput{
				SessionID: "session-123",
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SessionStartSuggestions(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d; suggestions: %+v", len(got), tt.wantLen, got)
				return
			}
			if tt.wantLen > 0 && tt.wantKind != "" {
				if got[0].Kind != tt.wantKind {
					t.Errorf("Kind = %q, want %q", got[0].Kind, tt.wantKind)
				}
			}
			for _, substr := range tt.wantTexts {
				found := false
				for _, s := range got {
					if contains(s.Text, substr) || contains(s.Action, substr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected substring %q in suggestions, got: %+v", substr, got)
				}
			}
		})
	}
}

func TestPhaseTransitionSuggestions(t *testing.T) {
	tests := []struct {
		name      string
		input     *PhaseTransitionInput
		wantLen   int
		wantTexts []string
	}{
		{
			name:    "nil input",
			input:   nil,
			wantLen: 0,
		},
		{
			name:    "empty phases",
			input:   &PhaseTransitionInput{},
			wantLen: 0,
		},
		{
			name: "same phase (no transition)",
			input: &PhaseTransitionInput{
				PreviousPhase: "design",
				CurrentPhase:  "design",
			},
			wantLen: 0,
		},
		{
			name: "empty previous phase",
			input: &PhaseTransitionInput{
				PreviousPhase: "",
				CurrentPhase:  "design",
			},
			wantLen: 0,
		},
		{
			name: "requirements to design",
			input: &PhaseTransitionInput{
				PreviousPhase: "requirements",
				CurrentPhase:  "design",
			},
			wantLen:   1,
			wantTexts: []string{"Requirements phase complete", "architect"},
		},
		{
			name: "design to implementation",
			input: &PhaseTransitionInput{
				PreviousPhase: "design",
				CurrentPhase:  "implementation",
			},
			wantLen:   1,
			wantTexts: []string{"Design phase complete", "principal-engineer"},
		},
		{
			name: "implementation to validation",
			input: &PhaseTransitionInput{
				PreviousPhase: "implementation",
				CurrentPhase:  "validation",
			},
			wantLen:   1,
			wantTexts: []string{"Implementation phase complete", "qa-adversary"},
		},
		{
			name: "validation to any",
			input: &PhaseTransitionInput{
				PreviousPhase: "validation",
				CurrentPhase:  "done",
			},
			wantLen:   1,
			wantTexts: []string{"Validation phase complete", "/wrap"},
		},
		{
			name: "non-standard transition",
			input: &PhaseTransitionInput{
				PreviousPhase: "design",
				CurrentPhase:  "validation",
			},
			wantLen:   1,
			wantTexts: []string{"Phase transitioned"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PhaseTransitionSuggestions(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d; suggestions: %+v", len(got), tt.wantLen, got)
				return
			}
			if tt.wantLen > 0 && got[0].Kind != KindPhaseTransition {
				t.Errorf("Kind = %q, want %q", got[0].Kind, KindPhaseTransition)
			}
			for _, substr := range tt.wantTexts {
				found := false
				for _, s := range got {
					if contains(s.Text, substr) || contains(s.Action, substr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected substring %q in suggestions, got: %+v", substr, got)
				}
			}
		})
	}
}

func TestSubagentStopSuggestions(t *testing.T) {
	tests := []struct {
		name      string
		input     *SubagentInput
		wantLen   int
		wantTexts []string
	}{
		{
			name:    "nil input",
			input:   nil,
			wantLen: 0,
		},
		{
			name:    "empty agent name",
			input:   &SubagentInput{},
			wantLen: 0,
		},
		{
			name:      "qa-adversary completed",
			input:     &SubagentInput{AgentName: "qa-adversary"},
			wantLen:   1,
			wantTexts: []string{"QA completed", "/wrap"},
		},
		{
			name:      "architect completed",
			input:     &SubagentInput{AgentName: "architect"},
			wantLen:   1,
			wantTexts: []string{"Architecture design complete", "TDD"},
		},
		{
			name:      "requirements-analyst completed",
			input:     &SubagentInput{AgentName: "requirements-analyst"},
			wantLen:   1,
			wantTexts: []string{"Requirements gathered", "PRD"},
		},
		{
			name:      "principal-engineer completed",
			input:     &SubagentInput{AgentName: "principal-engineer"},
			wantLen:   1,
			wantTexts: []string{"Implementation complete", "qa-adversary"},
		},
		{
			name:    "unknown agent (no suggestion)",
			input:   &SubagentInput{AgentName: "custom-agent"},
			wantLen: 0,
		},
		{
			name:    "potnia (no suggestion)",
			input:   &SubagentInput{AgentName: "potnia"},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubagentStopSuggestions(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d; suggestions: %+v", len(got), tt.wantLen, got)
				return
			}
			if tt.wantLen > 0 && got[0].Kind != KindSubagentComplete {
				t.Errorf("Kind = %q, want %q", got[0].Kind, KindSubagentComplete)
			}
			for _, substr := range tt.wantTexts {
				found := false
				for _, s := range got {
					if contains(s.Text, substr) || contains(s.Action, substr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected substring %q in suggestions, got: %+v", substr, got)
				}
			}
		})
	}
}

func TestBudgetWarningSuggestions(t *testing.T) {
	tests := []struct {
		name      string
		input     *SessionInput
		wantLen   int
		wantTexts []string
	}{
		{
			name:    "nil input",
			input:   nil,
			wantLen: 0,
		},
		{
			name: "below all thresholds",
			input: &SessionInput{
				ToolCount:     10,
				WarnThreshold: 250,
				ParkThreshold: 400,
			},
			wantLen: 0,
		},
		{
			name: "at warn threshold",
			input: &SessionInput{
				ToolCount:     250,
				WarnThreshold: 250,
			},
			wantLen:   1,
			wantTexts: []string{"Consider /park", "preserve session state"},
		},
		{
			name: "above warn threshold",
			input: &SessionInput{
				ToolCount:     300,
				WarnThreshold: 250,
			},
			wantLen:   1,
			wantTexts: []string{"Consider /park"},
		},
		{
			name: "at park threshold (overrides warn)",
			input: &SessionInput{
				ToolCount:     400,
				WarnThreshold: 250,
				ParkThreshold: 400,
			},
			wantLen:   1,
			wantTexts: []string{"Session is deep", "/park", "/handoff"},
		},
		{
			name: "above park threshold",
			input: &SessionInput{
				ToolCount:     500,
				WarnThreshold: 250,
				ParkThreshold: 400,
			},
			wantLen:   1,
			wantTexts: []string{"Session is deep"},
		},
		{
			name: "zero thresholds (disabled)",
			input: &SessionInput{
				ToolCount:     100,
				WarnThreshold: 0,
				ParkThreshold: 0,
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BudgetWarningSuggestions(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d; suggestions: %+v", len(got), tt.wantLen, got)
				return
			}
			if tt.wantLen > 0 && got[0].Kind != KindBudgetWarning {
				t.Errorf("Kind = %q, want %q", got[0].Kind, KindBudgetWarning)
			}
			for _, substr := range tt.wantTexts {
				found := false
				for _, s := range got {
					if contains(s.Text, substr) || contains(s.Action, substr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected substring %q in suggestions, got: %+v", substr, got)
				}
			}
		})
	}
}

// contains is a test helper that checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

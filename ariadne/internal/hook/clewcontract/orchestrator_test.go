package threadcontract

import (
	"testing"
)

func TestExtractThroughline(t *testing.T) {
	tests := []struct {
		name             string
		toolResult       string
		wantDecision     string
		wantRationale    string
		wantNil          bool
	}{
		{
			name: "valid YAML throughline with quotes",
			toolResult: `request_id: req-001

directive:
  action: invoke_specialist

specialist:
  name: principal-engineer
  prompt: "Complete prompt here"

throughline:
  decision: "Route to principal-engineer for implementation"
  rationale: "TDD-user-auth approved with complete API contracts."
`,
			wantDecision:  "Route to principal-engineer for implementation",
			wantRationale: "TDD-user-auth approved with complete API contracts.",
			wantNil:       false,
		},
		{
			name: "valid YAML throughline without quotes",
			toolResult: `throughline:
  decision: Route to QA for validation
  rationale: Implementation complete, needs testing
`,
			wantDecision:  "Route to QA for validation",
			wantRationale: "Implementation complete, needs testing",
			wantNil:       false,
		},
		{
			name: "throughline with only decision",
			toolResult: `throughline:
  decision: "Initiative complete"
`,
			wantDecision:  "Initiative complete",
			wantRationale: "",
			wantNil:       false,
		},
		{
			name: "no throughline present",
			toolResult: `directive:
  action: complete

state_update:
  current_phase: null
`,
			wantNil: true,
		},
		{
			name:       "empty tool result",
			toolResult: "",
			wantNil:    true,
		},
		{
			name: "throughline keyword but no decision field",
			toolResult: `some_field:
  throughline: not what we want
  other: value
`,
			wantNil: true,
		},
		{
			name: "multi-line rationale (first line only)",
			toolResult: `throughline:
  decision: "Escalate to user"
  rationale: "SC-003 says 'rate limited' but doesn't specify thresholds."
`,
			wantDecision:  "Escalate to user",
			wantRationale: "SC-003 says 'rate limited' but doesn't specify thresholds.",
			wantNil:       false,
		},
		{
			name: "indented YAML",
			toolResult: `    throughline:
      decision: "Proceed with standard implementation"
      rationale: "No blockers identified"
`,
			wantDecision:  "Proceed with standard implementation",
			wantRationale: "No blockers identified",
			wantNil:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractThroughline(tt.toolResult)

			if tt.wantNil {
				if result != nil {
					t.Errorf("ExtractThroughline() = %+v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("ExtractThroughline() = nil, want non-nil")
			}

			if result.Decision != tt.wantDecision {
				t.Errorf("Decision = %q, want %q", result.Decision, tt.wantDecision)
			}

			if result.Rationale != tt.wantRationale {
				t.Errorf("Rationale = %q, want %q", result.Rationale, tt.wantRationale)
			}
		})
	}
}

func TestIsOrchestratorAgent(t *testing.T) {
	tests := []struct {
		name       string
		toolResult string
		want       bool
	}{
		{
			name: "has throughline",
			toolResult: `throughline:
  decision: "Something"
`,
			want: true,
		},
		{
			name: "has directive",
			toolResult: `directive:
  action: invoke_specialist
`,
			want: true,
		},
		{
			name: "has CONSULTATION_RESPONSE",
			toolResult: `# CONSULTATION_RESPONSE

directive:
  action: complete
`,
			want: true,
		},
		{
			name:       "plain text output",
			toolResult: "This is just regular tool output",
			want:       false,
		},
		{
			name:       "empty string",
			toolResult: "",
			want:       false,
		},
		{
			name: "has throughline word in different context",
			toolResult: `notes:
  - The throughline of this story is unclear
`,
			want: false, // Does not match - "throughline" needs to be a YAML key (followed by :)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOrchestratorAgent(tt.toolResult)
			if got != tt.want {
				t.Errorf("IsOrchestratorAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractThroughline_EdgeCases(t *testing.T) {
	// Test with special characters in decision/rationale
	toolResult := `throughline:
  decision: "Fix bug #123: Handle null values in API"
  rationale: "Users reported 500 errors when field=null"
`
	result := ExtractThroughline(toolResult)
	if result == nil {
		t.Fatal("Expected non-nil result for special characters")
	}
	if result.Decision != "Fix bug #123: Handle null values in API" {
		t.Errorf("Decision with special chars = %q", result.Decision)
	}
}

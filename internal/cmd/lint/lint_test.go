package lint

import (
	"testing"
)

func TestSkillAtPattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantN   int // expected number of raw regex matches (before exclusions)
	}{
		{
			name:  "bare @skill in bullet",
			input: "- @standards for code conventions",
			wantN: 1,
		},
		{
			name:  "backticked @skill",
			input: "Use `@doc-security` for templates",
			wantN: 1,
		},
		{
			name:  "fragment ref",
			input: "Produce using `@doc-sre#postmortem-template`.",
			wantN: 1,
		},
		{
			name:  "path ref",
			input: "See `@orchestrator-templates/schemas/request.md`",
			wantN: 1,
		},
		{
			name:  "multiple refs",
			input: "- @standards for code\n- @cross-rite for handoffs\n- @doc-artifacts for templates",
			wantN: 3,
		},
		{
			name:  "email address not matched",
			input: "Contact noreply@anthropic.com for help",
			wantN: 0,
		},
		{
			name:  "plain skill name clean",
			input: "- standards for code conventions\n- cross-rite for handoffs",
			wantN: 0,
		},
		{
			name:  "at sign mid-word not matched",
			input: "Use user@domain patterns",
			wantN: 0,
		},
		{
			name:  "start of line",
			input: "@smell-detection for patterns",
			wantN: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := skillAtPattern.FindAllStringIndex(tt.input, -1)
			if len(matches) != tt.wantN {
				t.Errorf("skillAtPattern.FindAll(%q) = %d matches, want %d", tt.input, len(matches), tt.wantN)
			}
		})
	}
}

func TestCheckSkillAtRefs(t *testing.T) {
	t.Run("no matches produces no findings", func(t *testing.T) {
		var findings []Finding
		checkSkillAtRefs("plain skill references only", "test.md", &findings)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("matches produce single finding with count", func(t *testing.T) {
		var findings []Finding
		body := "- @standards for code\n- @cross-rite for handoffs"
		checkSkillAtRefs(body, "test.md", &findings)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "skill-at-syntax" {
			t.Errorf("expected rule skill-at-syntax, got %s", findings[0].Rule)
		}
		if findings[0].Severity != SevHigh {
			t.Errorf("expected severity HIGH, got %s", findings[0].Severity)
		}
	})

	t.Run("excluded handles produce no findings", func(t *testing.T) {
		var findings []Finding
		body := "**Owner**: @api-team\n**Lead**: @product-lead"
		checkSkillAtRefs(body, "test.md", &findings)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for excluded handles, got %d", len(findings))
		}
	})

	t.Run("@skill-name in docs excluded", func(t *testing.T) {
		var findings []Finding
		body := "| `@skill-name` | CC has no resolution |"
		checkSkillAtRefs(body, "test.md", &findings)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for @skill-name documentation, got %d", len(findings))
		}
	})

	t.Run("mixed excluded and real refs counts correctly", func(t *testing.T) {
		var findings []Finding
		body := "- @standards for code\n**Owner**: @api-team"
		checkSkillAtRefs(body, "test.md", &findings)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Message != "body contains 1 @skill-name reference(s) — use plain skill name instead (see lexicon anti-patterns)" {
			t.Errorf("unexpected message: %s", findings[0].Message)
		}
	})
}

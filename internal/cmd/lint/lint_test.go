package lint

import (
	"os"
	"path/filepath"
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

func TestCheckSourcePathLeaks(t *testing.T) {
	t.Run("clean body produces no findings", func(t *testing.T) {
		var findings []Finding
		body := "Read the domain criteria:\n```\nRead(\".claude/skills/pinakes/domains/arch.md\")\n```"
		checkSourcePathLeaks(body, "test.md", &findings)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("Read with rites path produces HIGH finding", func(t *testing.T) {
		var findings []Finding
		body := "Load criteria:\n```\nRead(\"rites/shared/mena/pinakes/domains/arch.lego.md\")\n```"
		checkSourcePathLeaks(body, "test.md", &findings)
		var hasRead bool
		for _, f := range findings {
			if f.Rule == "source-path-read" {
				hasRead = true
				if f.Severity != SevHigh {
					t.Errorf("expected severity HIGH, got %s", f.Severity)
				}
			}
		}
		if !hasRead {
			t.Error("expected source-path-read finding")
		}
	})

	t.Run("rites/*/mena/ reference produces MEDIUM finding", func(t *testing.T) {
		var findings []Finding
		body := "Full documentation: `rites/strategy/mena/strategy-ref/INDEX.lego.md`"
		checkSourcePathLeaks(body, "test.md", &findings)
		var hasRef bool
		for _, f := range findings {
			if f.Rule == "source-path-ref" {
				hasRef = true
				if f.Severity != SevMedium {
					t.Errorf("expected severity MED, got %s", f.Severity)
				}
			}
		}
		if !hasRef {
			t.Error("expected source-path-ref finding")
		}
	})

	t.Run("source extension produces LOW finding", func(t *testing.T) {
		var findings []Finding
		body := "- [documentation](../templates/documentation/INDEX.lego.md) - templates"
		checkSourcePathLeaks(body, "test.md", &findings)
		var hasExt bool
		for _, f := range findings {
			if f.Rule == "source-path-ext" {
				hasExt = true
				if f.Severity != SevLow {
					t.Errorf("expected severity LOW, got %s", f.Severity)
				}
			}
		}
		if !hasExt {
			t.Error("expected source-path-ext finding")
		}
	})

	t.Run("materialization documentation excluded", func(t *testing.T) {
		var findings []Finding
		body := "**Target files**: .claude/skills/**/*.md (projected from rites/*/mena/**/*.lego.md)"
		checkSourcePathLeaks(body, "test.md", &findings)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for materialization docs, got %d", len(findings))
		}
	})
}

func TestLintSessionArtifactsInSharedMena(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantFound bool
	}{
		{
			name: "session_id triggers finding",
			content: "---\nname: bad-artifact\ndescription: test\nsession_id: session-20260302-123456-abcdef01\n---\n# Bad\n",
			wantFound: true,
		},
		{
			name: "throughline triggers finding",
			content: "---\nname: throughline-doc\ndescription: test\nthroughline: some-initiative-ref\n---\n# Bad\n",
			wantFound: true,
		},
		{
			name: "session-ref triggers finding",
			content: "---\nname: session-doc\ndescription: test\nsession-ref: session-20260301\n---\n# Bad\n",
			wantFound: true,
		},
		{
			name: "clean legomena passes",
			content: "---\nname: good-skill\ndescription: Permanent platform knowledge. Triggers: domain reference.\n---\n# Good\n",
			wantFound: false,
		},
		{
			name: "no frontmatter passes",
			content: "# Just a plain markdown file\nNo frontmatter here.\n",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := t.TempDir()
			sharedMenaDir := filepath.Join(projectRoot, "rites", "shared", "mena", "test-entry")
			if err := os.MkdirAll(sharedMenaDir, 0755); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			if err := os.WriteFile(filepath.Join(sharedMenaDir, "INDEX.lego.md"), []byte(tt.content), 0644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			report := &LintReport{}
			lintSessionArtifactsInSharedMena(projectRoot, report)

			found := false
			for _, f := range report.Legomena {
				if f.Rule == "session-artifact-in-shared-mena" {
					found = true
				}
			}
			if found != tt.wantFound {
				t.Errorf("session-artifact-in-shared-mena finding: got %v, want %v", found, tt.wantFound)
				if len(report.Legomena) > 0 {
					t.Logf("findings: %+v", report.Legomena)
				}
			}
		})
	}
}

func TestLintSessionArtifactsInSharedMena_NoSharedDir(t *testing.T) {
	projectRoot := t.TempDir()
	// No rites/shared/mena/ directory exists

	report := &LintReport{}
	lintSessionArtifactsInSharedMena(projectRoot, report)

	if len(report.Legomena) != 0 {
		t.Errorf("expected 0 findings when shared mena dir absent, got %d", len(report.Legomena))
	}
}

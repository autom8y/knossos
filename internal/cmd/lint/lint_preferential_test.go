package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestLintPreferentialLanguageGo(t *testing.T) {
	t.Run("clean file produces no findings", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "clean.go"),
			[]byte("package example\n\nfunc Hello() string { return \"hello\" }\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("expected 0 findings, got %d: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("hardcoded claude path flagged", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "bad.go"),
			[]byte("package example\n\nvar dir = \".claude/agents\"\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 1 {
			t.Fatalf("expected 1 finding, got %d: %+v", len(prefFindings), prefFindings)
		}
		if prefFindings[0].Severity != SevMedium {
			t.Errorf("expected severity MED, got %s", prefFindings[0].Severity)
		}
	})

	t.Run("hardcoded gemini reference flagged", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "bad.go"),
			[]byte("package example\n\nvar ch = \"gemini\"\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 1 {
			t.Fatalf("expected 1 finding, got %d: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("HA-tagged line excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "tagged.go"),
			[]byte("package example\n\nvar defaultChannel = \"claude\" // HA-001: Default channel\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("HA-tagged line should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("ClaudeChannel identifier excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "channel.go"),
			[]byte("package example\n\nvar ch = ClaudeChannel{}\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("ClaudeChannel should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("ChannelByName identifier excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "channel.go"),
			[]byte("package example\n\nvar ch = paths.ChannelByName(\"claude\")\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("ChannelByName should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("adapter whitelisted file excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		adapterDir := filepath.Join(projectRoot, "internal", "compiler", "claude") // HA-TEST: lint rule test fixture
		os.MkdirAll(adapterDir, 0755)
		os.WriteFile(
			filepath.Join(adapterDir, "compiler.go"),
			[]byte("package claude\n\nvar name = \"claude\"\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("whitelisted adapter file should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("paths/channel.go excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		channelDir := filepath.Join(projectRoot, "internal", "paths")
		os.MkdirAll(channelDir, 0755)
		os.WriteFile(
			filepath.Join(channelDir, "channel.go"),
			[]byte("package paths\n\nvar claude = \"claude\"\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("paths/channel.go should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("test files excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "foo_test.go"),
			[]byte("package example\n\nvar testDir = \".claude/agents\"\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("test files should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("multiple violations in one file", func(t *testing.T) {
		projectRoot := t.TempDir()
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "multi.go"),
			[]byte("package example\n\nvar a = \"claude\"\nvar b = \"gemini\"\nvar c = \".claude/settings\"\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageGo(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 3 {
			t.Errorf("expected 3 findings, got %d: %+v", len(prefFindings), prefFindings)
		}
	})
}

func TestIsWhitelisted(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"internal/compiler/claude/compiler.go", true},
		{"internal/compiler/gemini/compiler.go", true},
		{"internal/adapter_claude/adapter.go", true},
		{"internal/adapter_gemini/adapter.go", true},
		{"internal/channel/tools/tools.go", true},
		{"internal/hook/events/events.go", true},
		{"internal/paths/channel.go", true},
		{"internal/cmd/sync/sync.go", false},
		{"internal/materialize/materialize.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isWhitelisted(tt.path); got != tt.want {
				t.Errorf("isWhitelisted(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestHasGoIdentifierExclusion(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"ch = ChannelByName(\"claude\")", true},
		{"channels := AllChannels()", true},
		{"var ch = ClaudeChannel{}", true},
		{"var ch = GeminiChannel{}", true},
		{"var name = \"claude\"", false},
		{"filepath.Join(\".claude\")", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if got := hasGoIdentifierExclusion(tt.line); got != tt.want {
				t.Errorf("hasGoIdentifierExclusion(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestLintPreferentialLanguageMena(t *testing.T) {
	t.Run("clean .channel/ reference passes", func(t *testing.T) {
		projectRoot := t.TempDir()
		menaDir := filepath.Join(projectRoot, "mena", "example")
		os.MkdirAll(menaDir, 0755)
		os.WriteFile(
			filepath.Join(menaDir, "INDEX.lego.md"),
			[]byte("---\nname: example\n---\n\nRead(\".channel/skills/foo\")\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageMena(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("expected 0 findings, got %d: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run(".claude/ path flagged", func(t *testing.T) {
		projectRoot := t.TempDir()
		menaDir := filepath.Join(projectRoot, "mena", "example")
		os.MkdirAll(menaDir, 0755)
		os.WriteFile(
			filepath.Join(menaDir, "INDEX.lego.md"),
			[]byte("---\nname: example\n---\n\nRead(\".claude/skills/foo\")\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageMena(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 1 {
			t.Fatalf("expected 1 finding, got %d: %+v", len(prefFindings), prefFindings)
		}
		if prefFindings[0].Severity != SevMedium {
			t.Errorf("expected severity MED, got %s", prefFindings[0].Severity)
		}
	})

	t.Run(".gemini/ path flagged", func(t *testing.T) {
		projectRoot := t.TempDir()
		menaDir := filepath.Join(projectRoot, "mena", "example")
		os.MkdirAll(menaDir, 0755)
		os.WriteFile(
			filepath.Join(menaDir, "INDEX.lego.md"),
			[]byte("---\nname: example\n---\n\nSee .gemini/agents/ for details\n"),
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageMena(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 1 {
			t.Fatalf("expected 1 finding, got %d: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("HA-tagged markdown excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		menaDir := filepath.Join(projectRoot, "mena", "example")
		os.MkdirAll(menaDir, 0755)
		os.WriteFile(
			filepath.Join(menaDir, "INDEX.lego.md"),
			[]byte("---\nname: example\n---\n\n<!-- HA-005 --> .claude/settings.local.json is Claude-only\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageMena(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("HA-tagged markdown should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("line with both .channel/ and .claude/ excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		menaDir := filepath.Join(projectRoot, "mena", "example")
		os.MkdirAll(menaDir, 0755)
		os.WriteFile(
			filepath.Join(menaDir, "INDEX.lego.md"),
			[]byte("---\nname: example\n---\n\n.channel/ (was .claude/) is the canonical placeholder\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageMena(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("line with .channel/ should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("non-.md file excluded", func(t *testing.T) {
		projectRoot := t.TempDir()
		menaDir := filepath.Join(projectRoot, "mena", "example")
		os.MkdirAll(menaDir, 0755)
		os.WriteFile(
			filepath.Join(menaDir, "config.yaml"),
			[]byte("path: .claude/settings\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageMena(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 0 {
			t.Errorf("non-.md files should be excluded, got %d findings: %+v", len(prefFindings), prefFindings)
		}
	})

	t.Run("rites/*/mena/ scanned", func(t *testing.T) {
		projectRoot := t.TempDir()
		riteMenaDir := filepath.Join(projectRoot, "rites", "10x-dev", "mena", "example")
		os.MkdirAll(riteMenaDir, 0755)
		os.WriteFile(
			filepath.Join(riteMenaDir, "INDEX.lego.md"),
			[]byte("---\nname: example\n---\n\nSee .claude/commands/ for details\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		report := &LintReport{}
		lintPreferentialLanguageMena(projectRoot, report)
		prefFindings := filterByRule(report.Legomena, rulePreferentialLanguage)
		if len(prefFindings) != 1 {
			t.Fatalf("expected 1 finding from rites/*/mena/, got %d: %+v", len(prefFindings), prefFindings)
		}
	})
}

func TestRunLint_CheckFlag(t *testing.T) {
	t.Run("unknown check value returns error", func(t *testing.T) {
		projectRoot := t.TempDir()
		outputFormat := "text"
		verbose := false
		ctx := &cmdContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectRoot,
			},
		}
		err := runLint(ctx, "", "unknown-rule")
		if err == nil {
			t.Fatal("expected error for unknown check value")
		}
		if !strings.Contains(err.Error(), "unknown check") {
			t.Errorf("expected 'unknown check' in error, got: %s", err.Error())
		}
	})

	t.Run("check=preferential-language with no violations returns nil", func(t *testing.T) {
		projectRoot := t.TempDir()
		// Create internal/ with a clean file
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "clean.go"),
			[]byte("package example\n\nfunc Hello() string { return \"hello\" }\n"),
			0644,
		)
		outputFormat := "text"
		verbose := false
		ctx := &cmdContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectRoot,
			},
		}
		err := runLint(ctx, "", rulePreferentialLanguage)
		if err != nil {
			t.Errorf("expected nil error for clean project, got: %v", err)
		}
	})

	t.Run("check=preferential-language with violations returns error", func(t *testing.T) {
		projectRoot := t.TempDir()
		// Create internal/ with a violating file
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "bad.go"),
			[]byte("package example\n\nvar dir = \".claude/agents\"\n"), // HA-TEST: lint rule test fixture
			0644,
		)
		outputFormat := "text"
		verbose := false
		ctx := &cmdContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectRoot,
			},
		}
		err := runLint(ctx, "", rulePreferentialLanguage)
		if err == nil {
			t.Fatal("expected error for project with violations")
		}
		if !strings.Contains(err.Error(), "violations found") {
			t.Errorf("expected 'violations found' in error, got: %s", err.Error())
		}
	})

	t.Run("check=preferential-language only runs this rule", func(t *testing.T) {
		projectRoot := t.TempDir()
		// Create an agent file with missing frontmatter (would trigger agent rules)
		agentDir := filepath.Join(projectRoot, "rites", "test", "agents")
		os.MkdirAll(agentDir, 0755)
		os.WriteFile(
			filepath.Join(agentDir, "bad-agent.md"),
			[]byte("# No frontmatter agent\nThis has no YAML frontmatter.\n"),
			0644,
		)
		// Also create internal/ with a clean file
		internalDir := filepath.Join(projectRoot, "internal", "example")
		os.MkdirAll(internalDir, 0755)
		os.WriteFile(
			filepath.Join(internalDir, "clean.go"),
			[]byte("package example\n\nfunc Hello() string { return \"hello\" }\n"),
			0644,
		)
		outputFormat := "text"
		verbose := false
		ctx := &cmdContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectRoot,
			},
		}
		// With --check=preferential-language, agent issues should NOT appear
		err := runLint(ctx, "", rulePreferentialLanguage)
		if err != nil {
			t.Errorf("expected nil error (no preferential-language violations), got: %v", err)
		}
	})
}

// filterByRule extracts findings with the specified rule name.
func filterByRule(findings []Finding, rule string) []Finding {
	var result []Finding
	for _, f := range findings {
		if f.Rule == rule {
			result = append(result, f)
		}
	}
	return result
}

package lint

import (
	"os"
	"path/filepath"
	"testing"
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
			[]byte("package example\n\nvar dir = \".claude/agents\"\n"),
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
		adapterDir := filepath.Join(projectRoot, "internal", "compiler", "claude")
		os.MkdirAll(adapterDir, 0755)
		os.WriteFile(
			filepath.Join(adapterDir, "compiler.go"),
			[]byte("package claude\n\nvar name = \"claude\"\n"),
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
			[]byte("package example\n\nvar testDir = \".claude/agents\"\n"),
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
			[]byte("package example\n\nvar a = \"claude\"\nvar b = \"gemini\"\nvar c = \".claude/settings\"\n"),
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

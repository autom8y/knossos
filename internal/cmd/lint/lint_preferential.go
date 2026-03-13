package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Rule name constant.
const rulePreferentialLanguage = "preferential-language"

// Severity for all findings from this rule.
const prefLangSeverity = SevMedium // WARNING, not blocking

// adapterWhitelist contains path suffixes/patterns that are LEGITIMATE
// adapter code and should never be flagged.
var adapterWhitelist = []string{
	"adapter_claude",
	"adapter_gemini",
	"compiler/claude",
	"compiler/gemini",
	"channel/tools",
	"hook/events",
	"paths/channel.go",
}

// goIdentifierExclusions are Go identifiers that legitimately reference
// channel names as part of the channel abstraction layer.
var goIdentifierExclusions = []string{
	"ChannelByName",
	"AllChannels",
	"ClaudeChannel",
	"GeminiChannel",
}

// goHarnessPattern matches harness-specific identifiers in Go source.
// Case-insensitive to catch "claude", "Claude", "CLAUDE", etc.
var goHarnessPattern = regexp.MustCompile(`(?i)\bclaude\b|\bgemini\b`)

// haTagPattern matches HA-NNN tags in comments that document
// legitimate harness-specific references.
var haTagPattern = regexp.MustCompile(`//\s*HA-\d+`)

// haTagMdPattern matches HA-NNN tags in markdown comments.
var haTagMdPattern = regexp.MustCompile(`<!--\s*HA-\d+`)

// menaChannelPathPattern matches .claude/ or .gemini/ path references
// that should use .channel/ placeholder.
var menaChannelPathPattern = regexp.MustCompile(`\.(claude|gemini)/`)

func isWhitelisted(path string) bool {
	for _, pattern := range adapterWhitelist {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func hasGoIdentifierExclusion(line string) bool {
	for _, id := range goIdentifierExclusions {
		if strings.Contains(line, id) {
			return true
		}
	}
	return false
}

// lintPreferentialLanguageGo scans Go source files under internal/ for
// harness-specific references (claude/gemini) that are not in adapter code,
// not HA-tagged, and not excluded identifiers.
func lintPreferentialLanguageGo(projectRoot string, report *LintReport) {
	internalDir := filepath.Join(projectRoot, "internal")
	_ = filepath.WalkDir(internalDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Exclude test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relPath := mustRel(projectRoot, path)

		// Exclude whitelisted adapter files
		if isWhitelisted(relPath) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		report.Summary.Files++
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if !goHarnessPattern.MatchString(line) {
				continue
			}
			// Exclude HA-tagged lines
			if haTagPattern.MatchString(line) {
				continue
			}
			// Exclude Go identifier exclusions
			if hasGoIdentifierExclusion(line) {
				continue
			}

			report.Legomena = append(report.Legomena, Finding{
				File:     fmt.Sprintf("%s:%d", relPath, i+1),
				Severity: prefLangSeverity,
				Rule:     rulePreferentialLanguage,
				Message:  fmt.Sprintf("harness-specific reference: %s", strings.TrimSpace(line)),
			})
		}
		return nil
	})
}

// lintPreferentialLanguage runs both Go source and mena content scans.
func lintPreferentialLanguage(projectRoot string, report *LintReport) {
	lintPreferentialLanguageGo(projectRoot, report)
}

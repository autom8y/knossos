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

// lintPreferentialLanguageMena scans mena content (.md files in mena/ and
// rites/*/mena/) for channel-specific path references (.claude/ or .gemini/)
// that should use the .channel/ placeholder.
func lintPreferentialLanguageMena(projectRoot string, report *LintReport) {
	var dirs []string

	// Platform mena
	dirs = append(dirs, filepath.Join(projectRoot, "mena"))

	// All rites/*/mena/ directories and rites/*/ for agent files
	riteDir := filepath.Join(projectRoot, "rites")
	rites, _ := os.ReadDir(riteDir)
	for _, r := range rites {
		if r.IsDir() {
			dirs = append(dirs, filepath.Join(riteDir, r.Name(), "mena"))
			dirs = append(dirs, filepath.Join(riteDir, r.Name()))
		}
	}

	// Track mena dirs to avoid double-scanning
	menaDirs := make(map[string]bool)
	for _, dir := range dirs {
		if strings.HasSuffix(dir, "mena") {
			menaDirs[dir] = true
		}
	}

	for _, dir := range dirs {
		isMenaRoot := menaDirs[dir]
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				// When walking a rite root (not a mena dir), skip mena/
				// subdirectory since it is walked separately.
				if !isMenaRoot && d.Name() == "mena" && path != dir {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}

			relPath := mustRel(projectRoot, path)
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			report.Summary.Files++
			lines := strings.Split(string(data), "\n")
			for i, line := range lines {
				if !menaChannelPathPattern.MatchString(line) {
					continue
				}
				// Exclude HA-tagged lines (markdown comment style)
				if haTagMdPattern.MatchString(line) {
					continue
				}
				// Exclude lines already using .channel/ (may have both in explanation)
				if strings.Contains(line, ".channel/") {
					continue
				}

				report.Legomena = append(report.Legomena, Finding{
					File:     fmt.Sprintf("%s:%d", relPath, i+1),
					Severity: prefLangSeverity,
					Rule:     rulePreferentialLanguage,
					Message:  fmt.Sprintf("channel-specific path reference: %s", strings.TrimSpace(line)),
				})
			}
			return nil
		})
	}
}

// lintPreferentialLanguage runs both Go source and mena content scans.
func lintPreferentialLanguage(projectRoot string, report *LintReport) {
	lintPreferentialLanguageGo(projectRoot, report)
	lintPreferentialLanguageMena(projectRoot, report)
}

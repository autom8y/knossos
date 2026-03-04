// Package tokenizer provides token counting for context budget estimation.
// Uses tiktoken-go with cl100k_base encoding (closest approximation to Claude's tokenizer).
package tokenizer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tiktoken-go/tokenizer"
)

// Counter counts tokens in text using cl100k_base encoding.
type Counter struct {
	codec tokenizer.Codec
}

// New creates a new token counter.
func New() (*Counter, error) {
	codec, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return nil, err
	}
	return &Counter{codec: codec}, nil
}

// Count returns the token count for a string.
func (c *Counter) Count(text string) int {
	n, err := c.codec.Count(text)
	if err != nil {
		// Fallback: rough estimate of 4 chars per token
		return len(text) / 4
	}
	return n
}

// CountFile reads a file and returns its token count.
func (c *Counter) CountFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return c.Count(string(data)), nil
}

// BudgetReport contains token counts for a .claude/ directory.
type BudgetReport struct {
	TotalTokens int                    `json:"total_tokens"`
	Categories  map[string]int         `json:"categories"`
	Files       []FileTokenCount       `json:"files"`
	Sections    []SectionTokenCount    `json:"sections,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
}

// FileTokenCount is a per-file token count.
type FileTokenCount struct {
	Path   string `json:"path"`
	Tokens int    `json:"tokens"`
}

// SectionTokenCount is a per-section token count for CLAUDE.md.
type SectionTokenCount struct {
	Name   string `json:"name"`
	Tokens int    `json:"tokens"`
}

// CalculateBudget walks a .claude/ directory and counts tokens for all context files.
func (c *Counter) CalculateBudget(claudeDir string) (*BudgetReport, error) {
	report := &BudgetReport{
		Categories: make(map[string]int),
	}

	// Walk relevant subdirectories and files
	entries := []struct {
		category string
		path     string
		isDir    bool
	}{
		{"CLAUDE.md", filepath.Join(claudeDir, "CLAUDE.md"), false},
		{"agents", filepath.Join(claudeDir, "agents"), true},
		{"commands", filepath.Join(claudeDir, "commands"), true},
		{"skills", filepath.Join(claudeDir, "skills"), true},
		{"rules", filepath.Join(claudeDir, "rules"), true},
		{"settings", filepath.Join(claudeDir, "settings.local.json"), false},
		{"workflow", filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"), false},
	}

	for _, entry := range entries {
		if entry.isDir {
			tokens, files := c.countDir(entry.path, claudeDir)
			if tokens > 0 {
				report.Categories[entry.category] = tokens
				report.Files = append(report.Files, files...)
				report.TotalTokens += tokens
			}
		} else {
			tokens, err := c.CountFile(entry.path)
			if err != nil {
				continue // file doesn't exist, skip
			}
			report.Categories[entry.category] = tokens
			report.Files = append(report.Files, FileTokenCount{
				Path:   relPath(entry.path, claudeDir),
				Tokens: tokens,
			})
			report.TotalTokens += tokens
		}
	}

	// Parse CLAUDE.md sections
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	if sections := c.parseSections(claudeMdPath); len(sections) > 0 {
		report.Sections = sections
	}

	// Sort files by token count descending
	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].Tokens > report.Files[j].Tokens
	})

	// Warnings
	if claudeMd, ok := report.Categories["CLAUDE.md"]; ok && claudeMd > 3000 {
		report.Warnings = append(report.Warnings,
			"CLAUDE.md exceeds recommended 3000 tokens")
	}
	if report.TotalTokens > 10000 {
		report.Warnings = append(report.Warnings,
			"Total context exceeds 10000 tokens — consider compressing skills or extracting reference content")
	}

	return report, nil
}

// countDir walks a directory tree counting .md, .json, .yaml files.
func (c *Counter) countDir(dir, base string) (int, []FileTokenCount) {
	var total int
	var files []FileTokenCount

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}
		tokens, err := c.CountFile(path)
		if err != nil || tokens == 0 {
			return nil
		}
		total += tokens
		files = append(files, FileTokenCount{
			Path:   relPath(path, base),
			Tokens: tokens,
		})
		return nil
	})

	return total, files
}

// parseSections extracts KNOSSOS regions from CLAUDE.md and counts tokens per section.
func (c *Counter) parseSections(claudeMdPath string) []SectionTokenCount {
	data, err := os.ReadFile(claudeMdPath)
	if err != nil {
		return nil
	}

	content := string(data)
	var sections []SectionTokenCount

	// Find <!-- KNOSSOS:START name --> ... <!-- KNOSSOS:END name --> regions
	lines := strings.Split(content, "\n")
	var currentSection string
	var sectionLines []string

	for _, line := range lines {
		if strings.Contains(line, "<!-- KNOSSOS:START") {
			// Extract section name
			start := strings.Index(line, "KNOSSOS:START") + len("KNOSSOS:START")
			rest := strings.TrimSpace(line[start:])
			rest = strings.TrimSuffix(rest, "-->")
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				currentSection = parts[0]
				sectionLines = nil
			}
		} else if strings.Contains(line, "<!-- KNOSSOS:END") {
			if currentSection != "" && len(sectionLines) > 0 {
				text := strings.Join(sectionLines, "\n")
				tokens := c.Count(text)
				sections = append(sections, SectionTokenCount{
					Name:   currentSection,
					Tokens: tokens,
				})
			}
			currentSection = ""
			sectionLines = nil
		} else if currentSection != "" {
			sectionLines = append(sectionLines, line)
		}
	}

	// Sort by token count descending
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Tokens > sections[j].Tokens
	})

	return sections
}

func relPath(path, base string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return filepath.Base(path)
	}
	return rel
}

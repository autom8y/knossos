package sync

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/tokenizer"
)

// formatBudgetReport generates a human-readable budget report and adds structured data to the output map.
func formatBudgetReport(claudeDir string, out map[string]any) error {
	counter, err := tokenizer.New()
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "initializing tokenizer", err)
	}

	report, err := counter.CalculateBudget(claudeDir)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "calculating budget", err)
	}

	// Add structured budget data to output map
	budget := map[string]any{
		"total_tokens": report.TotalTokens,
		"categories":   report.Categories,
	}

	// Top files (limit to 10)
	topFiles := report.Files
	if len(topFiles) > 10 {
		topFiles = topFiles[:10]
	}
	fileList := make([]map[string]any, len(topFiles))
	for i, f := range topFiles {
		fileList[i] = map[string]any{
			"path":   f.Path,
			"tokens": f.Tokens,
		}
	}
	budget["top_files"] = fileList

	if len(report.Sections) > 0 {
		sectionList := make([]map[string]any, len(report.Sections))
		for i, s := range report.Sections {
			sectionList[i] = map[string]any{
				"name":   s.Name,
				"tokens": s.Tokens,
			}
		}
		budget["claude_md_sections"] = sectionList
	}

	if len(report.Warnings) > 0 {
		budget["warnings"] = report.Warnings
	}

	out["budget"] = budget
	return nil
}

// budgetText generates a human-readable text summary of the budget.
func budgetText(claudeDir string) (string, error) {
	counter, err := tokenizer.New()
	if err != nil {
		return "", err
	}

	report, err := counter.CalculateBudget(claudeDir)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("\nContext Budget Report\n")
	b.WriteString(strings.Repeat("=", 50) + "\n")
	b.WriteString(fmt.Sprintf("Total tokens in %s: %s\n\n",
		filepath.Base(claudeDir), formatNum(report.TotalTokens)))

	// Category breakdown
	b.WriteString("By category:\n")
	for _, cat := range []string{"CLAUDE.md", "agents", "skills", "commands", "rules", "settings", "workflow"} {
		if tokens, ok := report.Categories[cat]; ok && tokens > 0 {
			pct := float64(tokens) / float64(report.TotalTokens) * 100
			b.WriteString(fmt.Sprintf("  %-15s %6s tokens  (%4.1f%%)\n", cat, formatNum(tokens), pct))
		}
	}

	// CLAUDE.md sections
	if len(report.Sections) > 0 {
		b.WriteString("\nCLAUDE.md sections:\n")
		for _, s := range report.Sections {
			b.WriteString(fmt.Sprintf("  %-25s %6s tokens\n", s.Name, formatNum(s.Tokens)))
		}
	}

	// Top files
	b.WriteString("\nTop files:\n")
	limit := 10
	if len(report.Files) < limit {
		limit = len(report.Files)
	}
	for i := 0; i < limit; i++ {
		f := report.Files[i]
		b.WriteString(fmt.Sprintf("  %-40s %6s tokens\n", f.Path, formatNum(f.Tokens)))
	}

	// Warnings
	if len(report.Warnings) > 0 {
		b.WriteString("\nWarnings:\n")
		for _, w := range report.Warnings {
			b.WriteString(fmt.Sprintf("  - %s\n", w))
		}
	}

	return b.String(), nil
}

func formatNum(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}

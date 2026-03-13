package sync

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/tokenizer"
)

// formatBudgetReport generates a human-readable budget report and adds structured data to the output map.
func formatBudgetReport(channelDir string, out map[string]any) error {
	counter, err := tokenizer.New()
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "initializing tokenizer", err)
	}

	report, err := counter.CalculateBudget(channelDir)
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
		budget["context_file_sections"] = sectionList
	}

	if len(report.Warnings) > 0 {
		budget["warnings"] = report.Warnings
	}

	out["budget"] = budget
	return nil
}

// budgetText generates a human-readable text summary of the budget.
func budgetText(channelDir string) (string, error) {
	counter, err := tokenizer.New()
	if err != nil {
		return "", err
	}

	report, err := counter.CalculateBudget(channelDir)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("\nContext Budget Report\n")
	b.WriteString(strings.Repeat("=", 50) + "\n")
	fmt.Fprintf(&b, "Total tokens in %s: %s\n\n",
		filepath.Base(channelDir), formatNum(report.TotalTokens))

	// Category breakdown
	b.WriteString("By category:\n")
	// HA-CC: "CLAUDE.md" is the tokenizer category key for the CC context file; display label is channel-neutral.
	categoryLabels := map[string]string{
		"CLAUDE.md": "context file",
		"agents":    "agents",
		"skills":    "skills",
		"commands":  "commands",
		"rules":     "rules",
		"settings":  "settings",
		"workflow":  "workflow",
	}
	for _, cat := range []string{"CLAUDE.md", "agents", "skills", "commands", "rules", "settings", "workflow"} { // HA-CC: "CLAUDE.md" matches tokenizer category key
		if tokens, ok := report.Categories[cat]; ok && tokens > 0 {
			pct := float64(tokens) / float64(report.TotalTokens) * 100
			label := categoryLabels[cat]
			fmt.Fprintf(&b, "  %-15s %6s tokens  (%4.1f%%)\n", label, formatNum(tokens), pct)
		}
	}

	// Context file sections
	if len(report.Sections) > 0 {
		b.WriteString("\nContext file sections:\n")
		for _, s := range report.Sections {
			fmt.Fprintf(&b, "  %-25s %6s tokens\n", s.Name, formatNum(s.Tokens))
		}
	}

	// Top files
	b.WriteString("\nTop files:\n")
	limit := min(len(report.Files), 10)
	for i := range limit {
		f := report.Files[i]
		fmt.Fprintf(&b, "  %-40s %6s tokens\n", f.Path, formatNum(f.Tokens))
	}

	// Warnings
	if len(report.Warnings) > 0 {
		b.WriteString("\nWarnings:\n")
		for _, w := range report.Warnings {
			fmt.Fprintf(&b, "  - %s\n", w)
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

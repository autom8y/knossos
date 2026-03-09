package worktree

import (
	"fmt"
	"github.com/autom8y/knossos/internal/cmd/common"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/worktree"
)

type cleanupOptions struct {
	olderThan string
	dryRun    bool
	force     bool
}

// CleanupOutput represents the output of worktree cleanup.
type CleanupOutput struct {
	Removed     []string          `json:"removed"`
	Skipped     []string          `json:"skipped"`
	SkipReasons map[string]string `json:"skip_reasons,omitempty"`
	DryRun      bool              `json:"dry_run"`
	OlderThan   string            `json:"older_than"`
}

// Text implements output.Textable for CleanupOutput.
func (c CleanupOutput) Text() string {
	var b strings.Builder

	if c.DryRun {
		b.WriteString("DRY RUN - no changes made\n\n")
	}

	fmt.Fprintf(&b, "Worktrees older than: %s\n\n", c.OlderThan)

	if len(c.Removed) > 0 {
		if c.DryRun {
			b.WriteString("Would remove:\n")
		} else {
			b.WriteString("Removed:\n")
		}
		for _, id := range c.Removed {
			fmt.Fprintf(&b, "  - %s\n", id)
		}
	}

	if len(c.Skipped) > 0 {
		b.WriteString("\nSkipped:\n")
		for _, id := range c.Skipped {
			reason := c.SkipReasons[id]
			if reason != "" {
				fmt.Fprintf(&b, "  - %s (%s)\n", id, reason)
			} else {
				fmt.Fprintf(&b, "  - %s\n", id)
			}
		}
	}

	if len(c.Removed) == 0 && len(c.Skipped) == 0 {
		b.WriteString("No worktrees found to clean up\n")
	}

	return b.String()
}

func newCleanupCmd(ctx *cmdContext) *cobra.Command {
	var opts cleanupOptions

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up stale worktrees",
		Long: `Remove worktrees older than a specified duration.

By default, removes worktrees older than 7 days that have no uncommitted changes.
Use --force to remove worktrees even with uncommitted changes.

Examples:
  ari worktree cleanup
  ari worktree cleanup --older-than=7d
  ari worktree cleanup --older-than=1h --dry-run
  ari worktree cleanup --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanup(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.olderThan, "older-than", "7d", "Remove worktrees older than this (e.g., 7d, 24h, 1d)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Show what would be cleaned up without doing it")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force cleanup even with uncommitted changes")

	return cmd
}

func runCleanup(ctx *cmdContext, opts cleanupOptions) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Parse duration
	duration, err := parseDuration(opts.olderThan)
	if err != nil {
		err := errors.NewWithDetails(errors.CodeUsageError, "invalid duration format", map[string]any{
			"value":   opts.olderThan,
			"example": "7d, 24h, 1h",
		})
		return common.PrintAndReturn(printer, err)
	}

	cleanupOpts := worktree.CleanupOptions{
		OlderThan: duration,
		DryRun:    opts.dryRun,
		Force:     opts.force,
	}

	result, err := mgr.Cleanup(cleanupOpts)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	output := CleanupOutput{
		Removed:     result.Removed,
		Skipped:     result.Skipped,
		SkipReasons: result.SkipReasons,
		DryRun:      result.DryRun,
		OlderThan:   opts.olderThan,
	}

	return printer.Print(output)
}

// parseDuration parses a duration string like "7d", "24h", "1h".
func parseDuration(s string) (time.Duration, error) {
	// Handle day suffix specially
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days, err := parseInt(s[:len(s)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// Try standard Go duration parsing
	return time.ParseDuration(s)
}

// parseInt parses an integer from a string.
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}

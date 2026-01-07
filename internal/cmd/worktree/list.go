package worktree

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/worktree"
)

// ListOutput represents the output of worktree list.
type ListOutput struct {
	Worktrees []WorktreeSummary `json:"worktrees"`
	Count     int               `json:"count"`
	InCurrent bool              `json:"in_current"` // True if currently in a worktree
	CurrentID string            `json:"current_id,omitempty"`
}

// WorktreeSummary is a brief worktree entry for listing.
type WorktreeSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	Rite          string `json:"rite"`
	Age           string `json:"age"`
	SessionStatus string `json:"session_status"`
	IsDirty       bool   `json:"is_dirty"`
	Current       bool   `json:"current"`
}

// Headers implements output.Tabular for ListOutput.
func (l ListOutput) Headers() []string {
	return []string{"ID", "NAME", "RITE", "SESSION", "STATUS", "AGE"}
}

// Rows implements output.Tabular for ListOutput.
func (l ListOutput) Rows() [][]string {
	rows := make([][]string, len(l.Worktrees))
	for i, wt := range l.Worktrees {
		prefix := "  "
		if wt.Current {
			prefix = "* "
		}
		status := "clean"
		if wt.IsDirty {
			status = "dirty"
		}
		rite := wt.Rite
		if rite == "" {
			rite = "-"
		}
		rows[i] = []string{
			prefix + wt.ID,
			wt.Name,
			rite,
			wt.SessionStatus,
			status,
			wt.Age,
		}
	}
	return rows
}

// Text implements output.Textable for ListOutput.
func (l ListOutput) Text() string {
	if len(l.Worktrees) == 0 {
		return "No worktrees found"
	}
	// Let tabular handle it
	return ""
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all worktrees",
		Long: `List all git worktrees associated with this repository,
including their session and team status.

Examples:
  ari worktree list
  ari worktree list --output=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx)
		},
	}

	return cmd
}

func runList(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	worktrees, err := mgr.List()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Check if we're in a worktree
	currentWT, _ := mgr.CurrentWorktree()

	summaries := make([]WorktreeSummary, len(worktrees))
	for i, wt := range worktrees {
		isCurrent := currentWT != nil && currentWT.ID == wt.ID
		summaries[i] = WorktreeSummary{
			ID:            wt.ID,
			Name:          wt.Name,
			Path:          wt.Path,
			Rite:          wt.Rite,
			Age:           wt.Age,
			SessionStatus: wt.SessionStatus,
			IsDirty:       wt.IsDirty || wt.HasUntracked,
			Current:       isCurrent,
		}
	}

	result := ListOutput{
		Worktrees: summaries,
		Count:     len(summaries),
		InCurrent: currentWT != nil,
	}
	if currentWT != nil {
		result.CurrentID = currentWT.ID
	}

	if len(summaries) == 0 {
		printer.PrintLine("No worktrees found")
		printer.PrintLine("")
		printer.PrintLine(fmt.Sprintf("Create one with: ari worktree create <name>"))
		return nil
	}

	return printer.Print(result)
}

// formatDirtyStatus returns a human-readable status for dirty state.
func formatDirtyStatus(status worktree.WorktreeStatus) string {
	var parts []string
	if status.ChangedFiles > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", status.ChangedFiles))
	}
	if status.UntrackedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d untracked", status.UntrackedCount))
	}
	if len(parts) == 0 {
		return "clean"
	}
	return strings.Join(parts, ", ")
}

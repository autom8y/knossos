package worktree

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/worktree"
)

// StatusOutput represents the output of worktree status.
type StatusOutput struct {
	WorktreeID     string `json:"worktree_id"`
	Name           string `json:"name"`
	Path           string `json:"path"`
	Rite           string `json:"rite"`
	Branch         string `json:"branch,omitempty"`
	BaseBranch     string `json:"base_branch"`
	FromRef        string `json:"from_ref"`
	Complexity     string `json:"complexity"`
	CreatedAt      string `json:"created_at"`
	Age            string `json:"age"`
	IsDirty        bool   `json:"is_dirty"`
	HasUntracked   bool   `json:"has_untracked"`
	ChangedFiles   int    `json:"changed_files"`
	UntrackedCount int    `json:"untracked_count"`
	CommitsAhead   int    `json:"commits_ahead"`
	CommitsBehind  int    `json:"commits_behind"`
	SessionStatus  string `json:"session_status"`
	CurrentSession string `json:"current_session,omitempty"`
}

// Text implements output.Textable for StatusOutput.
func (s StatusOutput) Text() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Worktree: %s\n", s.WorktreeID))
	b.WriteString(fmt.Sprintf("  Name: %s\n", s.Name))
	b.WriteString(fmt.Sprintf("  Path: %s\n", s.Path))

	if s.Rite != "" && s.Rite != "none" {
		b.WriteString(fmt.Sprintf("  Rite: %s\n", s.Rite))
	}

	if s.Branch != "" {
		b.WriteString(fmt.Sprintf("  Branch: %s\n", s.Branch))
	} else {
		b.WriteString("  Branch: (detached)\n")
	}

	b.WriteString(fmt.Sprintf("  Base: %s\n", s.BaseBranch))
	b.WriteString(fmt.Sprintf("  Age: %s\n", s.Age))

	// Git status
	b.WriteString("\n")
	if s.IsDirty || s.HasUntracked {
		b.WriteString("Git Status: dirty\n")
		if s.ChangedFiles > 0 {
			b.WriteString(fmt.Sprintf("  Modified files: %d\n", s.ChangedFiles))
		}
		if s.UntrackedCount > 0 {
			b.WriteString(fmt.Sprintf("  Untracked files: %d\n", s.UntrackedCount))
		}
	} else {
		b.WriteString("Git Status: clean\n")
	}

	if s.CommitsAhead > 0 || s.CommitsBehind > 0 {
		b.WriteString(fmt.Sprintf("  Ahead: %d, Behind: %d (vs %s)\n", s.CommitsAhead, s.CommitsBehind, s.BaseBranch))
	}

	// Session status
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Session: %s\n", s.SessionStatus))
	if s.CurrentSession != "" {
		b.WriteString(fmt.Sprintf("  ID: %s\n", s.CurrentSession))
	}

	return b.String()
}

func newStatusCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [id]",
		Short: "Show worktree status",
		Long: `Show detailed status for a worktree, including git state,
session state, and rite configuration.

If no ID is specified and you're in a worktree, shows status for that worktree.

Examples:
  ari worktree status
  ari worktree status wt-20260104-143052-a1b2
  ari worktree status feature-auth`,
		Args: common.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idOrName := ""
			if len(args) > 0 {
				idOrName = args[0]
			}
			return runStatus(ctx, idOrName)
		},
	}

	return cmd
}

func runStatus(ctx *cmdContext, idOrName string) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	var status *worktree.WorktreeStatus

	if idOrName == "" {
		// If no ID specified, try current worktree
		currentWT, err := mgr.CurrentWorktree()
		if err != nil || currentWT == nil {
			// Not in a worktree, list all
			return runList(ctx)
		}
		idOrName = currentWT.ID
	}

	status, err = mgr.Status(idOrName)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	result := StatusOutput{
		WorktreeID:     status.ID,
		Name:           status.Name,
		Path:           status.Path,
		Rite:           status.Rite,
		Branch:         status.Branch,
		BaseBranch:     status.BaseBranch,
		FromRef:        status.FromRef,
		Complexity:     status.Complexity,
		CreatedAt:      status.CreatedAt.Format(time.RFC3339),
		Age:            status.Age,
		IsDirty:        status.IsDirty,
		HasUntracked:   status.HasUntracked,
		ChangedFiles:   status.ChangedFiles,
		UntrackedCount: status.UntrackedCount,
		CommitsAhead:   status.CommitsAhead,
		CommitsBehind:  status.CommitsBehind,
		SessionStatus:  status.SessionStatus,
		CurrentSession: status.CurrentSession,
	}

	return printer.Print(result)
}

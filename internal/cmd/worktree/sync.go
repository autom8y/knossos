package worktree

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/worktree"
)

type syncOptions struct {
	pull bool
}

// SyncOutput represents the output of worktree sync.
type SyncOutput struct {
	Success     bool     `json:"success"`
	WorktreeID  string   `json:"worktree_id"`
	Name        string   `json:"name"`
	BaseBranch  string   `json:"base_branch"`
	Ahead       int      `json:"ahead"`
	Behind      int      `json:"behind"`
	Diverged    bool     `json:"diverged"`
	UpToDate    bool     `json:"up_to_date"`
	Pulled      bool     `json:"pulled"`
	PullError   string   `json:"pull_error,omitempty"`
	Conflicts   []string `json:"conflicts,omitempty"`
	Message     string   `json:"message"`
}

// Text implements output.Textable for SyncOutput.
func (s SyncOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Worktree: %s (%s)\n", s.Name, s.WorktreeID))
	b.WriteString(fmt.Sprintf("Base branch: %s\n\n", s.BaseBranch))

	if s.UpToDate {
		b.WriteString("Status: Up to date\n")
	} else if s.Diverged {
		b.WriteString(fmt.Sprintf("Status: Diverged (+%d/-%d)\n", s.Ahead, s.Behind))
		b.WriteString("  Your branch and upstream have diverged.\n")
		b.WriteString("  Consider rebasing or merging upstream changes.\n")
	} else if s.Ahead > 0 {
		b.WriteString(fmt.Sprintf("Status: Ahead by %d commit(s)\n", s.Ahead))
	} else if s.Behind > 0 {
		b.WriteString(fmt.Sprintf("Status: Behind by %d commit(s)\n", s.Behind))
	}

	if s.Pulled {
		b.WriteString("\nPull: Successful\n")
	} else if s.PullError != "" {
		b.WriteString(fmt.Sprintf("\nPull: Failed - %s\n", s.PullError))
	}

	if len(s.Conflicts) > 0 {
		b.WriteString(fmt.Sprintf("\nConflicts (%d):\n", len(s.Conflicts)))
		for _, conflict := range s.Conflicts {
			b.WriteString(fmt.Sprintf("  - %s\n", conflict))
		}
		b.WriteString("\nResolve conflicts and commit to complete the merge.\n")
	}

	return b.String()
}

func newSyncCmd(ctx *cmdContext) *cobra.Command {
	var opts syncOptions

	cmd := &cobra.Command{
		Use:   "sync [id-or-name]",
		Short: "Sync worktree with upstream",
		Long: `Check synchronization status between a worktree and its upstream branch.

Shows commits ahead/behind and optionally pulls changes.
If no worktree is specified and you're in one, syncs that worktree.

Examples:
  ari worktree sync
  ari worktree sync feature-auth
  ari worktree sync feature-auth --pull`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idOrName := ""
			if len(args) > 0 {
				idOrName = args[0]
			}
			return runSync(ctx, idOrName, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.pull, "pull", false, "Pull changes from upstream (fast-forward only)")

	return cmd
}

func runSync(ctx *cmdContext, idOrName string, opts syncOptions) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// If no ID specified, try current worktree
	if idOrName == "" {
		currentWT, err := mgr.CurrentWorktree()
		if err != nil || currentWT == nil {
			printer.PrintLine("Not in a worktree. Specify a worktree ID or name.")
			return fmt.Errorf("not in a worktree")
		}
		idOrName = currentWT.ID
	}

	syncOpts := worktree.SyncOptions{
		Pull: opts.pull,
	}

	result, err := mgr.Sync(idOrName, syncOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Determine message
	message := "up to date"
	if result.Diverged {
		message = "diverged from upstream"
	} else if result.Ahead > 0 && result.Behind > 0 {
		message = fmt.Sprintf("%d ahead, %d behind", result.Ahead, result.Behind)
	} else if result.Ahead > 0 {
		message = fmt.Sprintf("%d ahead", result.Ahead)
	} else if result.Behind > 0 {
		message = fmt.Sprintf("%d behind", result.Behind)
	}

	output := SyncOutput{
		Success:    true,
		WorktreeID: result.Worktree.ID,
		Name:       result.Worktree.Name,
		BaseBranch: result.Worktree.BaseBranch,
		Ahead:      result.Ahead,
		Behind:     result.Behind,
		Diverged:   result.Diverged,
		UpToDate:   result.UpToDate,
		Pulled:     result.Pulled,
		PullError:  result.PullError,
		Conflicts:  result.Conflicts,
		Message:    message,
	}

	return printer.Print(output)
}

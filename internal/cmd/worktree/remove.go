package worktree

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

type removeOptions struct {
	force bool
}

// RemoveOutput represents the output of worktree remove.
type RemoveOutput struct {
	Success    bool   `json:"success"`
	Removed    string `json:"removed"`
	WasForced  bool   `json:"was_forced,omitempty"`
}

// Text implements output.Textable for RemoveOutput.
func (r RemoveOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Removed worktree: %s\n", r.Removed))
	if r.WasForced {
		b.WriteString("  (forced removal)\n")
	}
	return b.String()
}

func newRemoveCmd(ctx *cmdContext) *cobra.Command {
	var opts removeOptions

	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a worktree",
		Long: `Remove a git worktree and its associated metadata.

By default, refuses to remove worktrees with uncommitted changes.
Use --force to override.

Examples:
  ari worktree remove wt-20260104-143052-a1b2
  ari worktree remove feature-auth
  ari worktree remove feature-auth --force`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(ctx, args[0], opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force removal even with uncommitted changes")

	return cmd
}

func runRemove(ctx *cmdContext, idOrName string, opts removeOptions) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	err = mgr.Remove(idOrName, opts.force)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	result := RemoveOutput{
		Success:   true,
		Removed:   idOrName,
		WasForced: opts.force,
	}

	return printer.Print(result)
}

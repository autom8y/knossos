// Package worktree implements the ari worktree commands for git worktree management.
package worktree

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/worktree"
)

// cmdContext holds shared state for worktree commands.
type cmdContext struct {
	common.BaseContext
}

// NewWorktreeCmd creates the worktree command group.
func NewWorktreeCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "worktree",
		Short: "Manage git worktrees for parallel Claude sessions",
		Long: `Manage git worktrees for running parallel Claude Code sessions
with filesystem isolation.

Git worktrees allow multiple sessions to work on different features
simultaneously without branch conflicts or file contention.

Examples:
  ari worktree create feature-auth
  ari worktree list
  ari worktree cleanup --older-than=7d`,
	}

	// Add subcommands
	cmd.AddCommand(newCreateCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newRemoveCmd(ctx))
	cmd.AddCommand(newCleanupCmd(ctx))

	// Advanced operations (Task 6)
	cmd.AddCommand(newSwitchCmd(ctx))
	cmd.AddCommand(newCloneCmd(ctx))
	cmd.AddCommand(newSyncCmd(ctx))
	cmd.AddCommand(newExportCmd(ctx))
	cmd.AddCommand(newImportCmd(ctx))

	// Worktree commands require project context
	common.SetNeedsProject(cmd, true, true)
	common.SetGroupCommand(cmd)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// getManager creates a worktree manager from the context.
func (c *cmdContext) getManager() (*worktree.Manager, error) {
	// Start from project dir if specified, otherwise current dir
	workDir := ""
	if c.ProjectDir != nil && *c.ProjectDir != "" {
		workDir = *c.ProjectDir
	} else {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	return worktree.NewManager(workDir)
}

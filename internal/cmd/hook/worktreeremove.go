// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// worktreeRemovePayload is the stdin JSON payload CC sends for WorktreeRemove events.
// CC sends: {worktree_path} for cleanup.
type worktreeRemovePayload struct {
	WorktreePath  string `json:"worktree_path"`
	HookEventName string `json:"hook_event_name"`
}

// newWorktreeRemoveCmd creates the worktree-remove hook subcommand for WorktreeRemove events.
func newWorktreeRemoveCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree-remove",
		Short: "Handle WorktreeRemove event: remove a git worktree",
		Long: `Handles CC WorktreeRemove cleanup events.

This hook is triggered when Claude Code removes a worktree. It runs
'git worktree remove' to clean up the linked worktree filesystem entry.

Input (stdin JSON from CC):
  {"hook_event_name":"WorktreeRemove","worktree_path":"/absolute/path/to/worktree"}

All output goes to STDERR. Exit 0 = success.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorktreeRemove(ctx)
		},
	}

	return cmd
}

// runWorktreeRemove implements the WorktreeRemove cleanup hook.
func runWorktreeRemove(ctx *cmdContext) error {
	// All output to STDERR.
	stderr := os.Stderr

	// Step 1: Read stdin JSON payload from CC.
	stdinBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(stderr, "worktree-remove: failed to read stdin: %v\n", err)
		return err
	}

	var payload worktreeRemovePayload
	if len(stdinBytes) > 0 {
		if err := json.Unmarshal(stdinBytes, &payload); err != nil {
			fmt.Fprintf(stderr, "worktree-remove: failed to parse stdin JSON: %v\n", err)
			return err
		}
	}

	// Step 2: Validate worktree_path.
	worktreePath := strings.TrimSpace(payload.WorktreePath)
	if worktreePath == "" {
		fmt.Fprintf(stderr, "worktree-remove: stdin JSON missing 'worktree_path' field\n")
		return fmt.Errorf("worktree-remove: missing worktree_path")
	}

	// Step 3: Remove the git worktree.
	fmt.Fprintf(stderr, "worktree-remove: removing git worktree at %s\n", worktreePath)
	gitCmd := exec.Command("git", "worktree", "remove", worktreePath)
	gitCmd.Stderr = stderr
	if err := gitCmd.Run(); err != nil {
		fmt.Fprintf(stderr, "worktree-remove: git worktree remove failed: %v\n", err)
		return err
	}

	fmt.Fprintf(stderr, "worktree-remove: worktree removed successfully\n")
	return nil
}

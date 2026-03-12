// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
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
		Long: `Handles WorktreeRemove cleanup events.

This hook is triggered when the harness removes a worktree. It runs
'git worktree remove' to clean up the linked worktree filesystem entry.

Input (stdin JSON):
  {"hook_event_name":"WorktreeRemove","worktree_path":"/absolute/path/to/worktree"}

All output goes to STDERR. Exit 0 = success.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runWorktreeRemove(cmd, ctx)
			})
		},
	}

	return cmd
}

// runWorktreeRemove implements the WorktreeRemove cleanup hook.
func runWorktreeRemove(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runWorktreeRemoveCore(cmd, ctx, printer)
}

// runWorktreeRemoveCore contains the actual logic with injected printer for testing.
func runWorktreeRemoveCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer) error {
	// All output to STDERR.
	stderr := os.Stderr

	// Step 0: Get hook environment and verify signature.
	hookEnv := ctx.getHookEnv(cmd)
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Step 1: Read stdin JSON payload from CC.
	stdinBytes := hookEnv.RawPayload
	var payload worktreeRemovePayload
	if len(stdinBytes) > 0 {
		if err := json.Unmarshal(stdinBytes, &payload); err != nil {
			fmt.Fprintf(stderr, "worktree-remove: failed to parse stdin JSON: %v\n", err)
			return err
		}
	}

	// Step 2: Event guard — only process WorktreeRemove events (or empty for direct CLI).
	if payload.HookEventName != "" && payload.HookEventName != "WorktreeRemove" {
		fmt.Fprintf(stderr, "worktree-remove: skipping non-WorktreeRemove event %q\n", payload.HookEventName)
		return nil
	}

	// Step 3: Validate worktree_path.
	worktreePath := strings.TrimSpace(payload.WorktreePath)
	if worktreePath == "" {
		fmt.Fprintf(stderr, "worktree-remove: stdin JSON missing 'worktree_path' field\n")
		return errors.New(errors.CodeUsageError, "worktree-remove: missing worktree_path")
	}

	// Step 4: Remove the git worktree.
	fmt.Fprintf(stderr, "worktree-remove: removing git worktree at %s\n", worktreePath)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	gitCmd := exec.CommandContext(timeoutCtx, "git", "worktree", "remove", worktreePath)
	gitCmd.Stderr = stderr
	if err := gitCmd.Run(); err != nil {
		fmt.Fprintf(stderr, "worktree-remove: git worktree remove failed: %v\n", err)
		return err
	}

	fmt.Fprintf(stderr, "worktree-remove: worktree removed successfully\n")
	return nil
}

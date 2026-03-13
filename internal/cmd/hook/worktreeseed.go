// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// worktreeCreatePayload is the stdin JSON payload CC sends for WorktreeCreate events.
// CC sends: {session_id, name, cwd, hook_event_name}
// There is no worktree_path field — the hook must create the worktree itself
// and print the absolute path to stdout.
type worktreeCreatePayload struct {
	Name            string `json:"name"`
	SessionID       string `json:"session_id"`
	CWD             string `json:"cwd"`
	HookEventName   string `json:"hook_event_name"`
}

// newWorktreeSeedCmd creates the worktree-seed hook subcommand for WorktreeCreate events.
func newWorktreeSeedCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree-seed",
		Short: "Handle WorktreeCreate event: create and seed a git worktree",
		Long: `Handles WorktreeCreate delegation events.

This hook is called by the harness when it wants to create a new git worktree.
As a delegation hook, this command MUST:
  1. Create the git worktree (git worktree add)
  2. Print the absolute worktree path to STDOUT (harness reads this)
  3. Send all other output to STDERR

Input (stdin JSON):
  {"hook_event_name":"WorktreeCreate","session_id":"...","name":"feature-x","cwd":"/project"}

Output (stdout -- read by the harness as the worktree path):
  /absolute/path/to/worktree

The worktree is created at $PROJECT_DIR/.knossos/worktrees/{name}.
If the main worktree has an ACTIVE_RITE, ari sync runs in the new worktree
to seed the channel directory with the rite configuration.

Exit 0 = success; non-zero = failure (harness will not proceed).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorktreeSeed(cmd, ctx)
		},
	}

	return cmd
}

// runWorktreeSeed implements the WorktreeCreate delegation hook.
func runWorktreeSeed(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runWorktreeSeedCore(cmd, ctx, printer)
}

// runWorktreeSeedCore contains the actual logic with injected printer for testing.
func runWorktreeSeedCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer) error {
	// All log output must go to STDERR. Stdout is reserved for the worktree path.
	stderr := os.Stderr

	// Step 0: Get hook environment and verify signature.
	hookEnv := ctx.getHookEnv(cmd)
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Step 1: Read stdin JSON payload from CC.
	stdinBytes := hookEnv.RawPayload
	var payload worktreeCreatePayload
	if len(stdinBytes) > 0 {
		if err := json.Unmarshal(stdinBytes, &payload); err != nil {
			fmt.Fprintf(stderr, "worktree-seed: failed to parse stdin JSON: %v\n", err)
			return err
		}
	}

	// HA-CC: CLAUDE_PROJECT_DIR is the CC wire protocol env var for the project root.
	projectRoot := os.Getenv("CLAUDE_PROJECT_DIR")
	if projectRoot == "" {
		// Fallback to CWD from payload, then os.Getwd().
		if payload.CWD != "" {
			projectRoot = payload.CWD
		} else {
			projectRoot, _ = os.Getwd()
			if projectRoot == "" {
				fmt.Fprintf(stderr, "worktree-seed: cannot determine project root\n")
				return errors.New(errors.CodeValidationFailed, "worktree-seed: cannot determine project root")
			}
		}
	}

	// Step 3: Determine worktree slug from name field.
	slug := strings.TrimSpace(payload.Name)
	if slug == "" {
		fmt.Fprintf(stderr, "worktree-seed: stdin JSON missing 'name' field\n")
		return errors.New(errors.CodeUsageError, "worktree-seed: missing worktree name")
	}

	// Step 4: Determine worktree path: projectRoot/.knossos/worktrees/{slug}
	worktreesDir := filepath.Join(projectRoot, ".knossos", "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		fmt.Fprintf(stderr, "worktree-seed: failed to create worktrees dir %s: %v\n", worktreesDir, err)
		return err
	}
	worktreePath := filepath.Join(worktreesDir, slug)

	// Step 5: Create the git worktree.
	fmt.Fprintf(stderr, "worktree-seed: creating git worktree at %s\n", worktreePath)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	gitCmd := exec.CommandContext(timeoutCtx, "git", "-C", projectRoot, "worktree", "add", worktreePath, "HEAD")
	gitCmd.Stderr = stderr
	if err := gitCmd.Run(); err != nil {
		fmt.Fprintf(stderr, "worktree-seed: git worktree add failed: %v\n", err)
		return err
	}

	// Step 6: Read ACTIVE_RITE from the main worktree (best-effort).
	riteName := paths.NewResolver(projectRoot).ReadActiveRite()

	// Step 7: If a rite is active, run ari sync in the new worktree to seed the channel dir.
	if riteName != "" {
		fmt.Fprintf(stderr, "worktree-seed: seeding worktree with rite %q\n", riteName)
		if err := seedWorktreeRite(worktreePath, riteName, ctx); err != nil {
			// Non-fatal: the worktree was created; sync failure is recoverable.
			fmt.Fprintf(stderr, "worktree-seed: sync failed (non-fatal): %v\n", err)
		}
	} else {
		fmt.Fprintf(stderr, "worktree-seed: no ACTIVE_RITE in main worktree; skipping rite sync\n")
	}

	// Step 8: Print absolute worktree path to STDOUT. CC reads this as the result.
	absPath, err := filepath.Abs(worktreePath)
	if err != nil {
		absPath = worktreePath
	}
	fmt.Print(absPath)

	return nil
}

// seedWorktreeRite runs ari sync (rite scope only) in the given worktree directory
// using the materialize package directly (no subprocess). This seeds the worktree's
// channel directory with the specified rite.
func seedWorktreeRite(worktreePath, riteName string, ctx *cmdContext) error {
	resolver := paths.NewResolver(worktreePath)
	m := NewWiredMaterializer(resolver)

	opts := materialize.SyncOptions{
		Scope:    materialize.ScopeRite,
		RiteName: riteName,
	}

	_, err := m.Sync(opts)
	return err
}

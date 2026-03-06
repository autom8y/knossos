// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/materialize"
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
		Long: `Handles CC WorktreeCreate delegation events.

This hook is called by Claude Code when it wants to create a new git worktree.
As a delegation hook, this command MUST:
  1. Create the git worktree (git worktree add)
  2. Print the absolute worktree path to STDOUT (CC reads this)
  3. Send all other output to STDERR

Input (stdin JSON from CC):
  {"hook_event_name":"WorktreeCreate","session_id":"...","name":"feature-x","cwd":"/project"}

Output (stdout -- read by CC as the worktree path):
  /absolute/path/to/worktree

The worktree is created at CLAUDE_PROJECT_DIR/.claude/worktrees/{name}.
If the main worktree has an ACTIVE_RITE, ari sync runs in the new worktree
to seed .claude/ with the rite configuration.

Exit 0 = success; non-zero = failure (CC will not proceed).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorktreeSeed(ctx)
		},
	}

	return cmd
}

// runWorktreeSeed implements the WorktreeCreate delegation hook.
func runWorktreeSeed(ctx *cmdContext) error {
	// All log output must go to STDERR. Stdout is reserved for the worktree path.
	stderr := os.Stderr

	// Step 1: Read stdin JSON payload from CC.
	stdinBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(stderr, "worktree-seed: failed to read stdin: %v\n", err)
		return err
	}

	var payload worktreeCreatePayload
	if len(stdinBytes) > 0 {
		if err := json.Unmarshal(stdinBytes, &payload); err != nil {
			fmt.Fprintf(stderr, "worktree-seed: failed to parse stdin JSON: %v\n", err)
			return err
		}
	}

	// Step 2: Resolve the main project root from CLAUDE_PROJECT_DIR env var.
	// CC sets CLAUDE_PROJECT_DIR to the main project root.
	projectRoot := os.Getenv("CLAUDE_PROJECT_DIR")
	if projectRoot == "" {
		// Fallback to CWD from payload, then os.Getwd().
		if payload.CWD != "" {
			projectRoot = payload.CWD
		} else {
			projectRoot, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(stderr, "worktree-seed: cannot determine project root: %v\n", err)
				return err
			}
		}
	}

	// Step 3: Determine worktree slug from name field.
	slug := strings.TrimSpace(payload.Name)
	if slug == "" {
		fmt.Fprintf(stderr, "worktree-seed: stdin JSON missing 'name' field\n")
		return fmt.Errorf("worktree-seed: missing worktree name")
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

	// Step 7: If a rite is active, run ari sync in the new worktree to seed .claude/.
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
// .claude/ directory with the specified rite.
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

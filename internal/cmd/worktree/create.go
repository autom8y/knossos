package worktree

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/session"
	"github.com/autom8y/knossos/internal/worktree"
)

type createOptions struct {
	rite       string
	fromRef    string
	complexity string
}

// CreateOutput represents the output of worktree create.
type CreateOutput struct {
	Success      bool   `json:"success"`
	WorktreeID   string `json:"worktree_id"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	Rite         string `json:"rite"`
	FromRef      string `json:"from_ref"`
	Complexity   string `json:"complexity"`
	CreatedAt    string `json:"created_at"`
	Instructions string `json:"instructions"`
}

// Text implements output.Textable for CreateOutput.
func (c CreateOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Created worktree: %s\n", c.WorktreeID)
	fmt.Fprintf(&b, "  Name: %s\n", c.Name)
	fmt.Fprintf(&b, "  Path: %s\n", c.Path)
	if c.Rite != "" && c.Rite != "none" {
		fmt.Fprintf(&b, "  Rite: %s\n", c.Rite)
	}
	// "claude" is the CC CLI binary name (not a knossos concept).
	fmt.Fprintf(&b, "\nTo start working: cd %s && claude\n", c.Path)
	return b.String()
}

func newCreateCmd(ctx *cmdContext) *cobra.Command {
	var opts createOptions

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new worktree for parallel session",
		Long: `Create a new git worktree for running a parallel agentic session.

The worktree is created in .worktrees/{id}/ with isolated filesystem,
allowing parallel work without conflicts.

Examples:
  ari worktree create feature-auth
  ari worktree create feature-auth --rite=10x-dev
  ari worktree create bugfix --from=develop
  ari worktree create experiment --complexity=PATCH`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(ctx, args[0], opts)
		},
	}

	cmd.Flags().StringVar(&opts.rite, "rite", "", "Rite (practice bundle) to activate in worktree")
	cmd.Flags().StringVar(&opts.fromRef, "from", "", "Git ref to create from (default: HEAD)")
	cmd.Flags().StringVar(&opts.complexity, "complexity", "MODULE", "Session complexity: PATCH, MODULE, SYSTEM, INITIATIVE, MIGRATION")

	return cmd
}

func runCreate(ctx *cmdContext, name string, opts createOptions) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Validate complexity
	if !session.IsValidComplexity(opts.complexity) {
		err := errors.New(errors.CodeUsageError, "invalid complexity: must be PATCH, MODULE, SYSTEM, INITIATIVE, or MIGRATION")
		return common.PrintAndReturn(printer, err)
	}

	createOpts := worktree.CreateOptions{
		Name:       name,
		Rite:       opts.rite,
		FromRef:    opts.fromRef,
		Complexity: opts.complexity,
	}

	wt, err := mgr.Create(createOpts)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	result := CreateOutput{
		Success:      true,
		WorktreeID:   wt.ID,
		Path:         wt.Path,
		Name:         wt.Name,
		Rite:         wt.Rite,
		FromRef:      wt.FromRef,
		Complexity:   wt.Complexity,
		CreatedAt:    wt.CreatedAt.Format(time.RFC3339),
		Instructions: fmt.Sprintf("cd %s && claude", wt.Path), // "claude" is the CC CLI binary name
	}

	return printer.Print(result)
}

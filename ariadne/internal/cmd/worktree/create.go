package worktree

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/worktree"
)

type createOptions struct {
	team       string
	fromRef    string
	complexity string
}

// CreateOutput represents the output of worktree create.
type CreateOutput struct {
	Success      bool   `json:"success"`
	WorktreeID   string `json:"worktree_id"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	Team         string `json:"team"`
	FromRef      string `json:"from_ref"`
	Complexity   string `json:"complexity"`
	CreatedAt    string `json:"created_at"`
	Instructions string `json:"instructions"`
}

// Text implements output.Textable for CreateOutput.
func (c CreateOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Created worktree: %s\n", c.WorktreeID))
	b.WriteString(fmt.Sprintf("  Name: %s\n", c.Name))
	b.WriteString(fmt.Sprintf("  Path: %s\n", c.Path))
	if c.Team != "" && c.Team != "none" {
		b.WriteString(fmt.Sprintf("  Team: %s\n", c.Team))
	}
	b.WriteString(fmt.Sprintf("\nTo start working: cd %s && claude\n", c.Path))
	return b.String()
}

func newCreateCmd(ctx *cmdContext) *cobra.Command {
	var opts createOptions

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new worktree for parallel session",
		Long: `Create a new git worktree for running a parallel Claude Code session.

The worktree is created in .worktrees/{id}/ with isolated filesystem,
allowing parallel work without conflicts.

Examples:
  ari worktree create feature-auth
  ari worktree create feature-auth --team=10x-dev-pack
  ari worktree create bugfix --from=develop
  ari worktree create experiment --complexity=PATCH`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(ctx, args[0], opts)
		},
	}

	cmd.Flags().StringVar(&opts.team, "team", "", "Team pack to activate in worktree")
	cmd.Flags().StringVar(&opts.fromRef, "from", "", "Git ref to create from (default: HEAD)")
	cmd.Flags().StringVar(&opts.complexity, "complexity", "MODULE", "Session complexity: PATCH, MODULE, SYSTEM, INITIATIVE, MIGRATION")

	return cmd
}

func runCreate(ctx *cmdContext, name string, opts createOptions) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Validate complexity
	if !isValidComplexity(opts.complexity) {
		err := errors.New(errors.CodeUsageError, "invalid complexity: must be PATCH, MODULE, SYSTEM, INITIATIVE, or MIGRATION")
		printer.PrintError(err)
		return err
	}

	createOpts := worktree.CreateOptions{
		Name:       name,
		Team:       opts.team,
		FromRef:    opts.fromRef,
		Complexity: opts.complexity,
	}

	wt, err := mgr.Create(createOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	result := CreateOutput{
		Success:      true,
		WorktreeID:   wt.ID,
		Path:         wt.Path,
		Name:         wt.Name,
		Team:         wt.Team,
		FromRef:      wt.FromRef,
		Complexity:   wt.Complexity,
		CreatedAt:    wt.CreatedAt.Format(time.RFC3339),
		Instructions: fmt.Sprintf("cd %s && claude", wt.Path),
	}

	return printer.Print(result)
}

func isValidComplexity(c string) bool {
	switch c {
	case "PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION":
		return true
	default:
		return false
	}
}

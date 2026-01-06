package worktree

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/worktree"
)

type cloneOptions struct {
	team        string
	copySession bool
}

// CloneOutput represents the output of worktree clone.
type CloneOutput struct {
	Success        bool   `json:"success"`
	WorktreeID     string `json:"worktree_id"`
	Name           string `json:"name"`
	Path           string `json:"path"`
	Team           string `json:"team"`
	SourceID       string `json:"source_id"`
	SourceName     string `json:"source_name"`
	CreatedAt      string `json:"created_at"`
	SessionCopied  bool   `json:"session_copied"`
	Instructions   string `json:"instructions"`
}

// Text implements output.Textable for CloneOutput.
func (c CloneOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Cloned worktree: %s\n", c.WorktreeID))
	b.WriteString(fmt.Sprintf("  Name: %s\n", c.Name))
	b.WriteString(fmt.Sprintf("  Path: %s\n", c.Path))
	b.WriteString(fmt.Sprintf("  Source: %s (%s)\n", c.SourceName, c.SourceID))
	if c.Team != "" && c.Team != "none" {
		b.WriteString(fmt.Sprintf("  Team: %s\n", c.Team))
	}
	if c.SessionCopied {
		b.WriteString("  Session context: copied\n")
	}
	b.WriteString(fmt.Sprintf("\nTo start working: cd %s && claude\n", c.Path))
	return b.String()
}

func newCloneCmd(ctx *cmdContext) *cobra.Command {
	var opts cloneOptions

	cmd := &cobra.Command{
		Use:   "clone <source-id-or-name> <new-name>",
		Short: "Clone a worktree with its metadata",
		Long: `Clone an existing worktree, creating a new one with copied metadata.

The new worktree is created from the source worktree's current HEAD,
preserving team, complexity, and other metadata settings.

Examples:
  ari worktree clone feature-auth feature-auth-v2
  ari worktree clone wt-20260104-143052-a1b2 experiment
  ari worktree clone feature-auth branch-b --rite=10x-dev-pack
  ari worktree clone feature-auth backup --copy-session`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(ctx, args[0], args[1], opts)
		},
	}

	cmd.Flags().StringVar(&opts.team, "rite", "", "Override rite (practice bundle) (default: copy from source)")
	cmd.Flags().StringVar(&opts.team, "team", "", "Deprecated: use --rite instead")
	cmd.Flags().BoolVar(&opts.copySession, "copy-session", false, "Copy session context from source")

	cmd.Flags().MarkDeprecated("team", "use --rite instead")

	return cmd
}

func runClone(ctx *cmdContext, sourceIDOrName, newName string, opts cloneOptions) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Get source info for output
	sourceStatus, _ := mgr.Status(sourceIDOrName)
	sourceName := sourceIDOrName
	sourceID := sourceIDOrName
	if sourceStatus != nil {
		sourceName = sourceStatus.Name
		sourceID = sourceStatus.ID
	}

	cloneOpts := worktree.CloneOptions{
		Team:        opts.team,
		CopySession: opts.copySession,
	}

	wt, err := mgr.Clone(sourceIDOrName, newName, cloneOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	result := CloneOutput{
		Success:       true,
		WorktreeID:    wt.ID,
		Name:          wt.Name,
		Path:          wt.Path,
		Team:          wt.Team,
		SourceID:      sourceID,
		SourceName:    sourceName,
		CreatedAt:     wt.CreatedAt.Format(time.RFC3339),
		SessionCopied: opts.copySession,
		Instructions:  fmt.Sprintf("cd %s && claude", wt.Path),
	}

	return printer.Print(result)
}

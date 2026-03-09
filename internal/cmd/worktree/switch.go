package worktree

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/worktree"
)

type switchOptions struct {
	updateRite bool
}

// SwitchOutput represents the output of worktree switch.
type SwitchOutput struct {
	Success     bool   `json:"success"`
	WorktreeID  string `json:"worktree_id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Rite        string `json:"rite"`
	RiteUpdated bool   `json:"rite_updated"`
	Message     string `json:"message"`
}

// Text implements output.Textable for SwitchOutput.
func (s SwitchOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Switched to worktree: %s\n", s.WorktreeID)
	fmt.Fprintf(&b, "  Name: %s\n", s.Name)
	fmt.Fprintf(&b, "  Path: %s\n", s.Path)
	if s.Rite != "" && s.Rite != "none" {
		fmt.Fprintf(&b, "  Rite: %s", s.Rite)
		if s.RiteUpdated {
			b.WriteString(" (updated)")
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "\nTo navigate: cd %s\n", s.Path)
	return b.String()
}

func newSwitchCmd(ctx *cmdContext) *cobra.Command {
	var opts switchOptions

	cmd := &cobra.Command{
		Use:   "switch <id-or-name>",
		Short: "Switch context to a different worktree",
		Long: `Switch the active context to a different worktree.

This command updates session context and optionally syncs the rite configuration.
Note: This does not change your shell's working directory. Use 'cd' to navigate.

Examples:
  ari worktree switch feature-auth
  ari worktree switch wt-20260104-143052-a1b2
  ari worktree switch feature-auth --update-rite`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSwitch(ctx, args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.updateRite, "update-rite", false, "Update ACTIVE_RITE to match worktree's rite")

	return cmd
}

func runSwitch(ctx *cmdContext, idOrName string, opts switchOptions) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	switchOpts := worktree.WorktreeSwitchOptions{
		UpdateRite: opts.updateRite,
	}

	wt, err := mgr.Switch(idOrName, switchOpts)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	result := SwitchOutput{
		Success:     true,
		WorktreeID:  wt.ID,
		Name:        wt.Name,
		Path:        wt.Path,
		Rite:        wt.Rite,
		RiteUpdated: opts.updateRite && wt.Rite != "",
		Message:     fmt.Sprintf("Switched to worktree %s", wt.Name),
	}

	return printer.Print(result)
}

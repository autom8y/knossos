// Package agent implements the ari agent commands.
package agent

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// cmdContext holds shared state for agent commands.
type cmdContext struct {
	common.BaseContext
}

// NewAgentCmd creates the agent command group.
func NewAgentCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent management commands",
		Long: `Validate and manage agent specifications.

Agent commands work with agent files in rites and agents directories.
Use these commands to validate agent frontmatter, list agents, and check
agent compliance with the agent schema.`,
	}

	// Add subcommands
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newNewCmd(ctx))
	cmd.AddCommand(newUpdateCmd(ctx))

	// Agent commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

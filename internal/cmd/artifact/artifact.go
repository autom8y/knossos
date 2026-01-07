// Package artifact implements the ari artifact commands.
package artifact

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/artifact"
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for artifact commands.
type cmdContext struct {
	common.SessionContext
}

// NewArtifactCmd creates the artifact command group.
func NewArtifactCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     outputFlag,
				Verbose:    verboseFlag,
				ProjectDir: projectDir,
			},
			SessionID: sessionID,
		},
	}

	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage workflow artifacts",
		Long:  `Register, query, and manage workflow artifacts across sessions.`,
	}

	// Add subcommands
	cmd.AddCommand(newRegisterCmd(ctx))
	cmd.AddCommand(newQueryCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newRebuildCmd(ctx))

	// Artifact commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// getRegistry creates an artifact registry from the context.
func (c *cmdContext) getRegistry() *artifact.Registry {
	projectDir := ""
	if c.ProjectDir != nil {
		projectDir = *c.ProjectDir
	}
	return artifact.NewRegistry(projectDir)
}

// getQuerier creates an artifact querier from the context.
func (c *cmdContext) getQuerier() *artifact.Querier {
	return artifact.NewQuerier(c.getRegistry())
}

// getAggregator creates an artifact aggregator from the context.
func (c *cmdContext) getAggregator() *artifact.Aggregator {
	return artifact.NewAggregator(c.getRegistry())
}

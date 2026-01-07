// Package artifact implements the ari artifact commands.
package artifact

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/artifact"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// cmdContext holds shared state for artifact commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
	sessionID  *string
}

// NewArtifactCmd creates the artifact command group.
func NewArtifactCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
		sessionID:  sessionID,
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

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	format := output.FormatText
	if c.output != nil {
		format = output.ParseFormat(*c.output)
	}
	verbose := false
	if c.verbose != nil {
		verbose = *c.verbose
	}
	return output.NewPrinter(format, os.Stdout, os.Stderr, verbose)
}

// getResolver creates a path resolver from the context.
func (c *cmdContext) getResolver() *paths.Resolver {
	projectDir := ""
	if c.projectDir != nil {
		projectDir = *c.projectDir
	}
	return paths.NewResolver(projectDir)
}

// getRegistry creates an artifact registry from the context.
func (c *cmdContext) getRegistry() *artifact.Registry {
	projectDir := ""
	if c.projectDir != nil {
		projectDir = *c.projectDir
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

// getSessionID returns the session ID to use (from flag or current).
func (c *cmdContext) getSessionID() (string, error) {
	if c.sessionID != nil && *c.sessionID != "" {
		return *c.sessionID, nil
	}
	return c.getCurrentSessionID()
}

// getCurrentSessionID reads the current session ID from .current-session file.
func (c *cmdContext) getCurrentSessionID() (string, error) {
	resolver := c.getResolver()
	data, err := os.ReadFile(resolver.CurrentSessionFile())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

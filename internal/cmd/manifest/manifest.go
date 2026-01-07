// Package manifest implements the ari manifest commands.
package manifest

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/manifest"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// cmdContext holds shared state for manifest commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
}

// NewManifestCmd creates the manifest command group.
func NewManifestCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
	}

	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Manage manifest files",
		Long:  `Show, validate, diff, and merge Claude Extension Manifest (CEM) files.`,
	}

	// Add subcommands
	cmd.AddCommand(newShowCmd(ctx))
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newDiffCmd(ctx))
	cmd.AddCommand(newMergeCmd(ctx))

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

// getSchemaValidator creates a manifest schema validator.
func (c *cmdContext) getSchemaValidator() (*manifest.SchemaValidator, error) {
	return manifest.NewSchemaValidator()
}

// defaultManifestPath returns the default manifest path.
func (c *cmdContext) defaultManifestPath() string {
	resolver := c.getResolver()
	return resolver.ClaudeDir() + "/manifest.json"
}

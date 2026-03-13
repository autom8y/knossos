// Package manifest implements the ari manifest commands.
package manifest

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/manifest"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// cmdContext holds shared state for manifest commands.
type cmdContext struct {
	common.BaseContext
}

// NewManifestCmd creates the manifest command group.
func NewManifestCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Manage manifest files",
		Long:  `Show, validate, diff, and merge Knossos project manifest files.`,
	}

	// Add subcommands
	cmd.AddCommand(newShowCmd(ctx))
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newDiffCmd(ctx))
	cmd.AddCommand(newMergeCmd(ctx))

	// Manifest commands require project context
	common.SetNeedsProject(cmd, true, true)
	common.SetGroupCommand(cmd)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// getSchemaValidator creates a manifest schema validator.
func (c *cmdContext) getSchemaValidator() (*manifest.SchemaValidator, error) {
	return manifest.NewSchemaValidator()
}

// defaultManifestPath returns the default manifest path.
// HA-CC: manifest.json is a CC-specific concept; other channels use different manifest formats/locations.
func (c *cmdContext) defaultManifestPath() string {
	resolver := c.GetResolver()
	return resolver.ChannelDir(paths.ClaudeChannel{}) + "/manifest.json"
}

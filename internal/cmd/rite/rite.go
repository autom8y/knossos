// Package rite implements the ari rite commands.
package rite

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/rite"
)

// cmdContext holds shared state for rite commands.
type cmdContext struct {
	common.BaseContext
}

// NewRiteCmd creates the rite command group.
func NewRiteCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "rite",
		Short: "Manage rite invocations and composition",
		Long: `List, invoke, release, and manage rite partial composition.

Rites are composable practice bundles. The invoke operation is additive
(borrow components without switching), while swap is replacement
(same as team switch).`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newInfoCmd(ctx))
	cmd.AddCommand(newCurrentCmd(ctx))
	cmd.AddCommand(newInvokeCmd(ctx))
	cmd.AddCommand(newReleaseCmd(ctx))
	cmd.AddCommand(newContextCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newPantheonCmd(ctx))

	// Rite commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// getDiscovery creates a rite discovery instance.
func (c *cmdContext) getDiscovery() *rite.Discovery {
	return rite.NewDiscovery(c.GetResolver())
}

// getInvoker creates a rite invoker.
func (c *cmdContext) getInvoker() *rite.Invoker {
	return rite.NewInvoker(c.GetResolver())
}

// getValidator creates a rite validator.
func (c *cmdContext) getValidator() *rite.Validator {
	return rite.NewValidator(c.GetResolver())
}

// getActiveRite reads the active rite from ACTIVE_RITE file.
func (c *cmdContext) getActiveRite() string {
	return c.GetResolver().ReadActiveRite()
}

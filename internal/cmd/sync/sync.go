// Package sync implements the ari sync commands.
package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
)

// cmdContext holds shared state for sync commands.
type cmdContext struct {
	common.BaseContext
}

// NewSyncCmd creates the sync command group.
func NewSyncCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize configuration with remotes",
		Long: `Synchronize .claude/ configuration with remote sources.

Sync tracks changes to configuration files and enables pulling updates
from remotes and pushing local changes back.

Supported remotes:
  - Local paths: /path/to/source, ./relative/path
  - HTTP(S) URLs: https://example.com/config
  - GitHub shorthand: org/repo
  - Git refs: HEAD:.claude/path, origin/main:.claude/path`,
	}

	// Add subcommands - each sets its own NeedsProject annotation
	cmd.AddCommand(newMaterializeCmd(ctx)) // NeedsProject=false (can bootstrap)
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newPullCmd(ctx))
	cmd.AddCommand(newPushCmd(ctx))
	cmd.AddCommand(newDiffCmd(ctx))
	cmd.AddCommand(newResolveCmd(ctx))
	cmd.AddCommand(newHistoryCmd(ctx))
	cmd.AddCommand(newResetCmd(ctx))
	cmd.AddCommand(newUserCmd(ctx)) // User sync (NeedsProject=false)

	// Sync parent command requires project (but not recursive - materialize/user overrides)
	common.SetNeedsProject(cmd, true, false)

	// Set NeedsProject for all subcommands except materialize and user
	for _, sub := range cmd.Commands() {
		if sub.Name() != "materialize" && sub.Name() != "user" {
			common.SetNeedsProject(sub, true, false)
		}
	}

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

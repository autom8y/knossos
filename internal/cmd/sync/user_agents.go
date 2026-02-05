package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/usersync"
)

func newUserAgentsCmd(ctx *cmdContext) *cobra.Command {
	var dryRun, recover, force, verbose bool

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Sync user agents to ~/.claude/agents/",
		Long: `Sync agent files from roster user-agents/ to ~/.claude/agents/.

Behavior:
  - Adds new agents from roster
  - Updates roster-managed agents when source changes
  - Preserves user-created agents (never deleted)
  - Skips agents that would shadow rite agents (collision detection)

Examples:
  # Sync agents
  ari sync user agents

  # Preview what would be synced
  ari sync user agents --dry-run

  # Adopt existing files into manifest
  ari sync user agents --recover

  # Force overwrite diverged files
  ari sync user agents --force`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := usersync.Options{
				DryRun:  dryRun,
				Recover: recover,
				Force:   force,
				Verbose: verbose,
			}
			return runUserSync(ctx, usersync.ResourceAgents, opts)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching roster")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	common.SetNeedsProject(cmd, false, false)

	return cmd
}

func runUserSync(ctx *cmdContext, resourceType usersync.ResourceType, opts usersync.Options) error {
	printer := ctx.getPrinter()

	syncer, err := usersync.NewSyncer(resourceType)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	result, err := syncer.Sync(opts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	return printer.Print(result)
}

package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/usersync"
)

func newUserCommandsCmd(ctx *cmdContext) *cobra.Command {
	var dryRun, recover, force, verbose bool

	cmd := &cobra.Command{
		Use:   "commands",
		Short: "Sync user commands to ~/.claude/commands/",
		Long: `Sync command files from roster mena/ to ~/.claude/commands/.

Commands use a nested directory structure (category/).
Manifest keys are relative paths: operations/commit.md

Behavior:
  - Recursively copies command directories
  - Preserves nested category structure
  - Only updates when source changes (checksum-based)
  - Preserves user-created commands (never deleted)
  - Skips commands that would shadow rite commands

Examples:
  # Sync commands
  ari sync user commands

  # Preview what would be synced
  ari sync user commands --dry-run

  # Adopt existing files into manifest
  ari sync user commands --recover

  # Force overwrite diverged files
  ari sync user commands --force`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := usersync.Options{
				DryRun:  dryRun,
				Recover: recover,
				Force:   force,
				Verbose: verbose,
			}
			return runUserSync(ctx, usersync.ResourceCommands, opts)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching roster")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	common.SetNeedsProject(cmd, false, false)

	return cmd
}

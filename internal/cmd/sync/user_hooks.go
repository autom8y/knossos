package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/usersync"
)

func newUserHooksCmd(ctx *cmdContext) *cobra.Command {
	var dryRun, recover, force, verbose bool

	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Sync user hooks to ~/.claude/hooks/",
		Long: `Sync hook files from knossos user-hooks/ to ~/.claude/hooks/.

Hooks have special handling:
  - lib/ directory: Recursive copy of shared libraries
  - *.yaml files: Hook configuration files
  - Shell scripts: Preserve executable permissions (+x)

Manifest keys are relative paths: lib/session-manager.sh, hooks.yaml

Behavior:
  - Recursively copies hook directories
  - Preserves lib/ directory structure
  - Maintains executable permissions on scripts
  - Only updates when source changes (checksum-based)
  - Preserves user-created hooks (never deleted)
  - Skips hooks that would shadow rite hooks

Examples:
  # Sync hooks
  ari sync user hooks

  # Preview what would be synced
  ari sync user hooks --dry-run

  # Adopt existing files into manifest
  ari sync user hooks --recover

  # Force overwrite diverged files
  ari sync user hooks --force`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := usersync.Options{
				DryRun:  dryRun,
				Recover: recover,
				Force:   force,
				Verbose: verbose,
			}
			return runUserSync(ctx, usersync.ResourceHooks, opts)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching knossos")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	common.SetNeedsProject(cmd, false, false)

	return cmd
}

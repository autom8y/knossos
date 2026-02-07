package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/usersync"
)

func newUserMenaCmd(ctx *cmdContext) *cobra.Command {
	var dryRun, recover, force, verbose bool

	cmd := &cobra.Command{
		Use:   "mena",
		Short: "Sync user mena (commands + skills) to ~/.claude/",
		Long: `Sync mena files from knossos mena/ to ~/.claude/commands/ and ~/.claude/skills/.

Mena files are routed by their source extension:
  .dro.md  -> ~/.claude/commands/ (dromena: invokable commands)
  .lego.md -> ~/.claude/skills/   (legomena: reference knowledge)

Extensions are stripped during projection:
  INDEX.dro.md  -> INDEX.md in commands/
  INDEX.lego.md -> INDEX.md in skills/

Behavior:
  - Routes files to commands/ or skills/ based on mena type
  - Strips .dro/.lego extensions from projected filenames
  - Preserves directory structure (progressive disclosure)
  - Only updates when source changes (checksum-based)
  - Preserves user-created content (never deleted)
  - Skips resources that would shadow rite mena

Examples:
  ari sync user mena
  ari sync user mena --dry-run
  ari sync user mena --recover
  ari sync user mena --force`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := usersync.Options{
				DryRun:  dryRun,
				Recover: recover,
				Force:   force,
				Verbose: verbose,
			}
			return runUserSync(ctx, usersync.ResourceMena, opts)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching knossos")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	common.SetNeedsProject(cmd, false, false)
	return cmd
}

package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/usersync"
)

func newUserSkillsCmd(ctx *cmdContext) *cobra.Command {
	var dryRun, recover, force, verbose bool

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Sync user skills to ~/.claude/skills/",
		Long: `Sync skill directories from knossos user-skills/ to ~/.claude/skills/.

Skills use a nested directory structure (category/skill-name/).
Manifest keys are relative paths: documentation/doc-artifacts/SKILL.md

Behavior:
  - Recursively copies skill directories
  - Preserves nested category structure
  - Only updates when source changes (checksum-based)
  - Preserves user-created skills (never deleted)
  - Skips skills that would shadow rite skills

Examples:
  # Sync skills
  ari sync user skills

  # Preview what would be synced
  ari sync user skills --dry-run

  # Adopt existing files into manifest
  ari sync user skills --recover

  # Force overwrite diverged files
  ari sync user skills --force`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := usersync.Options{
				DryRun:  dryRun,
				Recover: recover,
				Force:   force,
				Verbose: verbose,
			}
			return runUserSync(ctx, usersync.ResourceSkills, opts)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching knossos")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	common.SetNeedsProject(cmd, false, false)

	return cmd
}

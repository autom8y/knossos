package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func newUserCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Sync user-level resources to ~/.claude/",
		Long: `Sync user-level resources from roster to ~/.claude/.

User resources are globally available across all projects.
They are stored in ~/.claude/ and synced from $KNOSSOS_HOME/user-{type}/.

Resources:
  agents    - Agent prompts (user-agents/ -> ~/.claude/agents/)
  skills    - Skill references (user-skills/ -> ~/.claude/skills/)
  commands  - Slash commands (mena/ -> ~/.claude/commands/)
  hooks     - Hook scripts (user-hooks/ -> ~/.claude/hooks/)

Sync Behavior:
  - Additive: Never removes user-created content
  - Checksum-based: Only updates when source changes
  - Collision-aware: Skips resources that would shadow rite resources

Source Types:
  - roster          Synced from roster, checksums match
  - roster-diverged Originally from roster but locally modified
  - user            Created by user, not in roster`,
	}

	// Add subcommands
	cmd.AddCommand(newUserAgentsCmd(ctx))
	cmd.AddCommand(newUserSkillsCmd(ctx))
	cmd.AddCommand(newUserCommandsCmd(ctx))
	cmd.AddCommand(newUserHooksCmd(ctx))
	cmd.AddCommand(newUserAllCmd(ctx))

	// User sync doesn't require a project directory
	common.SetNeedsProject(cmd, false, false)

	return cmd
}

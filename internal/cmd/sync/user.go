package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func newUserCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Sync user-level resources to ~/.claude/",
		Long: `Sync user-level resources from knossos to ~/.claude/.

User resources are globally available across all projects.
They are stored in ~/.claude/ and synced from $KNOSSOS_HOME/{type}/.

Resources:
  agents  - Agent prompts (agents/ -> ~/.claude/agents/)
  mena    - Commands and skills (mena/ -> ~/.claude/commands/ + skills/)
  hooks   - Hook scripts (hooks/ -> ~/.claude/hooks/)

Sync Behavior:
  - Additive: Never removes user-created content
  - Checksum-based: Only updates when source changes
  - Collision-aware: Skips resources that would shadow rite resources

Source Types:
  - knossos          Synced from knossos, checksums match
  - knossos-diverged Originally from knossos but locally modified
  - user             Created by user, not from knossos`,
	}

	cmd.AddCommand(newUserAgentsCmd(ctx))
	cmd.AddCommand(newUserMenaCmd(ctx))
	cmd.AddCommand(newUserHooksCmd(ctx))
	cmd.AddCommand(newUserAllCmd(ctx))

	common.SetNeedsProject(cmd, false, false)
	return cmd
}

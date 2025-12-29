---
description: Switch agent team packs or list available teams
argument-hint: [pack-name] [--list] [--force] [--keep-all|--remove-all|--promote-all]
model: claude-sonnet-4-5
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

**Available teams**: !`ls ~/Code/roster/teams/`

## Your Task

Manage agent team packs. $ARGUMENTS

## Behavior

**If no arguments or querying current team:**
1. Read `.claude/ACTIVE_TEAM` and display current team
2. Show: "Active team: {name}" or "No team active"

**If `--list` or `-l`:**
1. Execute: `~/Code/roster/swap-team.sh --list`
2. Display all available team packs

**If `<pack-name>` provided:**
1. Execute: `~/Code/roster/swap-team.sh <pack-name> [flags]`
2. If orphan agents exist (agents in current team but not in target):
   - **Interactive (TTY)**: Prompt user for each orphan agent
   - **Non-interactive**: Require `--keep-all`, `--remove-all`, or `--promote-all` flag
3. Show confirmation with agent count
4. If SESSION_CONTEXT exists, update `active_team` field

## Orphan Agent Handling

When switching teams, agents that exist in the current project but not in the target team are called "orphans". You'll be prompted to choose for each:

| Choice | Key | Effect |
|--------|-----|--------|
| Keep | k | Agent stays in project (survives swap) |
| Promote | p | Agent moves to `~/.claude/agents/` (user-level) |
| Remove | r | Agent removed (available in `.claude/agents.backup/`) |
| Apply to all | a | Apply same choice to remaining orphans |

For CI/scripts (non-interactive), use flags:
- `--force`, `-f`: Re-swap even if already on target team
- `--keep-all`: Preserve all orphan agents in project
- `--remove-all`: Remove all orphans (backup available)
- `--promote-all`: Move all orphans to user-level

## Agent Provenance

Team swaps track agent provenance in `.claude/AGENT_MANIFEST.json`:
- **source**: `team` (from roster) or `user` (project-added)
- **origin**: Which team pack installed this agent
- **installed_at**: Timestamp of installation

**Note**: Team context (phase→agent routing) is automatically injected into every session via the session-context hook.

## Examples

```bash
/team                           # Show current team
/team --list                    # List all teams
/team 10x-dev-pack              # Switch (prompts for orphans)
/team hygiene-pack --keep-all   # Switch, keep all orphans
/team debt-pack --promote-all   # Switch, promote orphans to user-level
/team doc-team-pack --force     # Re-swap even if already on team
```

## Reference

Full documentation: `.claude/skills/team-ref/skill.md`

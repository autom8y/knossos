---
description: Switch rites or list available rites
argument-hint: [pack-name] [--list] [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: sonnet
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

**Available rites**: !`ls ${KNOSSOS_HOME:-~/Code/roster}/rites/`

## Your Task

Manage rites. $ARGUMENTS

## Behavior

**If no arguments or querying current rite:**
1. Read `.claude/ACTIVE_RITE` and display current rite
2. Show: "Active rite: {name}" or "No rite active"

**If `--list` or `-l`:**
1. Execute: `${KNOSSOS_HOME:-~/Code/roster}/swap-rite.sh --list`
2. Display all available rites

**If `<pack-name>` provided:**
1. Execute: `${KNOSSOS_HOME:-~/Code/roster}/swap-rite.sh <pack-name> [flags]`
2. If orphan agents exist (agents in current project but not in target rite):
   - **Interactive (TTY)**: Prompt user for each orphan agent
   - **Non-interactive**: Require `--keep-all`, `--remove-all`, or `--promote-all` flag
3. Show confirmation with agent count
4. If SESSION_CONTEXT exists, update `active_rite` field

## Orphan Agent Handling

When switching rites, agents that exist in the current project but not in the target rite are called "orphans". You'll be prompted to choose for each:

| Choice | Key | Effect |
|--------|-----|--------|
| Keep | k | Agent stays in project (survives swap) |
| Promote | p | Agent moves to `~/.claude/agents/` (user-level) |
| Remove | r | Agent removed (available in `.claude/agents.backup/`) |
| Apply to all | a | Apply same choice to remaining orphans |

For CI/scripts (non-interactive), use flags:
- `--update`, `-u`: Pull latest agent definitions from roster even if already on rite
- `--dry-run`: Preview changes without applying
- `--keep-all`: Preserve all orphan agents in project
- `--remove-all`: Remove all orphans (backup available)
- `--promote-all`: Move all orphans to user-level

## Agent Provenance

Rite swaps track agent provenance in `.claude/AGENT_MANIFEST.json`:
- **source**: `rite` (from roster) or `user` (project-added)
- **origin**: Which rite installed this agent
- **installed_at**: Timestamp of installation

**Note**: Rite context (phase->agent routing) is automatically injected into every session via the session-context hook.

## Quick Switch Commands

Quick-switch commands are derived from rite names:

| Rite | Quick Switch | Domain |
|------|--------------|--------|
| 10x-dev | `/10x` | Full feature development |
| debt-triage | `/debt` | Technical debt management |
| docs | `/docs` | Documentation workflows |
| ecosystem | `/ecosystem` | CEM/skeleton/roster infrastructure |
| forge | `/forge` | Rite creation |
| hygiene | `/hygiene` | Code quality, refactoring |
| intelligence | `/intelligence` | Analytics, research |
| rnd | `/rnd` | Exploration, prototyping |
| security | `/security` | Security assessment |
| sre | `/sre` | Operations, reliability |
| strategy | `/strategy` | Business analysis |

**Note**: Use `team-discovery` skill for programmatic rite metadata access.

## Examples

```bash
/rite                           # Show current rite
/rite --list                    # List all rites
/rite 10x-dev              # Switch (prompts for orphans)
/rite hygiene --keep-all   # Switch, keep all orphans
/rite debt-pack --promote-all   # Switch, promote orphans to user-level
/rite docs --update    # Update even if already on rite
```

## Reference

Full documentation: `.claude/skills/rite-ref/skill.md`

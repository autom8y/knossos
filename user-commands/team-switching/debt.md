---
description: Quick switch to debt-triage-pack (technical debt workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the technical debt triage team pack and display the team roster. $ARGUMENTS

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh debt-triage-pack $ARGUMENTS`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `debt-triage-pack`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on team | swap-team.sh |
| `--dry-run` | - | Preview changes without applying | swap-team.sh |
| `--keep-all` | - | Preserve all orphan agents in project | swap-team.sh |
| `--remove-all` | - | Remove all orphans (backup available) | swap-team.sh |
| `--promote-all` | - | Move all orphans to user-level | swap-team.sh |

## When to Use

- Technical debt inventory
- Prioritizing debt paydown
- Sprint planning for maintenance
- Risk assessment of shortcuts

## Reference

Full documentation: `.claude/skills/debt-ref/skill.md`

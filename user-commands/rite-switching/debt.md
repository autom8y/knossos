---
description: Quick switch to debt-triage (technical debt workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the technical debt triage rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `${KNOSSOS_HOME:-~/Code/roster}/swap-rite.sh debt-triage $ARGUMENTS`
2. Display the roster output from swap-rite.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `debt-triage`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | swap-rite.sh |
| `--dry-run` | - | Preview changes without applying | swap-rite.sh |
| `--keep-all` | - | Preserve all orphan agents in project | swap-rite.sh |
| `--remove-all` | - | Remove all orphans (backup available) | swap-rite.sh |
| `--promote-all` | - | Move all orphans to user-level | swap-rite.sh |

## When to Use

- Technical debt inventory
- Prioritizing debt paydown
- Sprint planning for maintenance
- Risk assessment of shortcuts

## Reference

Full documentation: `.claude/skills/debt-ref/skill.md`

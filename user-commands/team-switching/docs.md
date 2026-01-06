---
description: Quick switch to doc-team-pack (documentation workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the documentation rite and display the rite roster. $ARGUMENTS

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-rite.sh doc-team-pack $ARGUMENTS`
2. Display the roster output from swap-rite.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `doc-team-pack`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on team | swap-rite.sh |
| `--dry-run` | - | Preview changes without applying | swap-rite.sh |
| `--keep-all` | - | Preserve all orphan agents in project | swap-rite.sh |
| `--remove-all` | - | Remove all orphans (backup available) | swap-rite.sh |
| `--promote-all` | - | Move all orphans to user-level | swap-rite.sh |

## When to Use

- Documentation audits and cleanup
- Creating new documentation
- Restructuring doc organization
- Technical writing tasks

## Reference

Full documentation: `.claude/skills/docs-ref/skill.md`

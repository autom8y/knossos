---
description: Quick switch to hygiene (code quality workflow)
argument-hint: [--force] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the code hygiene rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite hygiene $ARGUMENTS`
2. Display the pantheon output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `hygiene`

## Flags

| Flag | Description |
|------|-------------|
| `--force` | Force regeneration even if already on this rite |
| `--dry-run` | Preview changes without applying |
| `--keep-all` | Preserve all orphan agents in project (default) |
| `--remove-all` | Remove orphans (backup in `.claude/.orphan-backup/`) |
| `--promote-all` | Move orphans to user-level (`~/.claude/agents/`) |

## When to Use

- Code quality audits
- Refactoring initiatives
- Reducing technical debt
- Enforcing architectural patterns

## Reference

Full documentation: `.claude/skills/hygiene-ref/skill.md`

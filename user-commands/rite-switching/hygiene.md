---
description: Quick switch to hygiene (code quality workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the code hygiene rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite hygiene $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `hygiene`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | ari sync materialize |
| `--dry-run` | - | Preview changes without applying | ari sync materialize |
| `--keep-all` | - | Preserve all orphan agents in project | ari sync materialize |
| `--remove-all` | - | Remove all orphans (backup available) | ari sync materialize |
| `--promote-all` | - | Move all orphans to user-level | ari sync materialize |

## When to Use

- Code quality audits
- Refactoring initiatives
- Reducing technical debt
- Enforcing architectural patterns

## Reference

Full documentation: `.claude/skills/hygiene-ref/skill.md`

---
description: Quick switch to debt-triage (technical debt workflow)
argument-hint: [--force] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the technical debt triage rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite debt-triage $ARGUMENTS`
2. Display the pantheon output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `debt-triage`

## Flags

| Flag | Description |
|------|-------------|
| `--force` | Force regeneration even if already on this rite |
| `--dry-run` | Preview changes without applying |
| `--keep-all` | Preserve all orphan agents in project (default) |
| `--remove-all` | Remove orphans (backup in `.claude/.orphan-backup/`) |
| `--promote-all` | Move orphans to user-level (`~/.claude/agents/`) |

## When to Use

- Technical debt inventory
- Prioritizing debt paydown
- Sprint planning for maintenance
- Risk assessment of shortcuts

## Reference

Full documentation: `rites/debt-triage/mena/debt-ref/INDEX.lego.md`

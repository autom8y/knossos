---
name: sre
description: Quick switch to sre (reliability workflow)
argument-hint: [--force] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the SRE rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite sre $ARGUMENTS`
2. Display the pantheon output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `sre`

## Flags

| Flag | Description |
|------|-------------|
| `--force` | Force regeneration even if already on this rite |
| `--dry-run` | Preview changes without applying |
| `--keep-all` | Preserve all orphan agents in project (default) |
| `--remove-all` | Remove orphans (backup in `.claude/.orphan-backup/`) |
| `--promote-all` | Move orphans to user-level (`~/.claude/agents/`) |

## When to Use

- System reliability improvements
- Observability and monitoring work
- Incident response preparation
- Chaos engineering experiments
- Platform and infrastructure work

## Reference

Full documentation: `rites/sre/mena/sre-ref/INDEX.lego.md`

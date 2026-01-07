---
description: Quick switch to rnd (innovation lab workflow)
argument-hint: [--force] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Innovation Lab (R&D) rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite rnd $ARGUMENTS`
2. Display the pantheon output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `rnd`

## Flags

| Flag | Description |
|------|-------------|
| `--force` | Force regeneration even if already on this rite |
| `--dry-run` | Preview changes without applying |
| `--keep-all` | Preserve all orphan agents in project (default) |
| `--remove-all` | Remove orphans (backup in `.claude/.orphan-backup/`) |
| `--promote-all` | Move orphans to user-level (`~/.claude/agents/`) |

## When to Use

- Evaluating new technologies
- Building proof-of-concept prototypes
- Long-term architecture planning
- Innovation and R&D exploration

## Workflow

```
scouting -> integration-analysis -> prototyping -> future-architecture
```

## Reference

Full documentation: `.claude/skills/rnd-ref/skill.md`

---
description: Quick switch to strategy (business strategy workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Strategy rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite strategy $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `strategy`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | ari sync materialize |
| `--dry-run` | - | Preview changes without applying | ari sync materialize |
| `--keep-all` | - | Preserve all orphan agents in project | ari sync materialize |
| `--remove-all` | - | Remove all orphans (backup available) | ari sync materialize |
| `--promote-all` | - | Move all orphans to user-level | ari sync materialize |

## When to Use

- Market analysis and sizing
- Competitive intelligence gathering
- Pricing and business model analysis
- Strategic roadmap planning

## Workflow

```
market-research -> competitive-analysis -> business-modeling -> strategic-planning
```

## Reference

Full documentation: `.claude/skills/strategy-ref/skill.md`

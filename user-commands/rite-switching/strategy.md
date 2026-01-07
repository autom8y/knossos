---
description: Quick switch to strategy-pack (business strategy workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Strategy Team pack and display the rite roster. $ARGUMENTS

## Behavior

1. Execute: `${KNOSSOS_HOME:-~/Code/roster}/swap-rite.sh strategy-pack $ARGUMENTS`
2. Display the roster output from swap-rite.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `strategy-pack`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | swap-rite.sh |
| `--dry-run` | - | Preview changes without applying | swap-rite.sh |
| `--keep-all` | - | Preserve all orphan agents in project | swap-rite.sh |
| `--remove-all` | - | Remove all orphans (backup available) | swap-rite.sh |
| `--promote-all` | - | Move all orphans to user-level | swap-rite.sh |

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

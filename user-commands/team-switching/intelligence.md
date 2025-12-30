---
description: Quick switch to intelligence-pack (product analytics workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the Product Intelligence Team pack and display the team roster. $ARGUMENTS

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh intelligence-pack $ARGUMENTS`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `intelligence-pack`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on team | swap-team.sh |
| `--dry-run` | - | Preview changes without applying | swap-team.sh |
| `--keep-all` | - | Preserve all orphan agents in project | swap-team.sh |
| `--remove-all` | - | Remove all orphans (backup available) | swap-team.sh |
| `--promote-all` | - | Move all orphans to user-level | swap-team.sh |

## When to Use

- Instrumenting new features with analytics
- Designing and running A/B tests
- User research and interview planning
- Data-driven product decisions

## Workflow

```
instrumentation -> research -> experimentation -> synthesis
```

## Reference

Full documentation: `.claude/skills/intelligence-ref/skill.md`

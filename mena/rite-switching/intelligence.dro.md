---
name: intelligence
description: Quick switch to intelligence (product analytics workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Product Intelligence rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite intelligence $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `intelligence`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | ari sync materialize |
| `--dry-run` | - | Preview changes without applying | ari sync materialize |
| `--keep-all` | - | Preserve all orphan agents in project | ari sync materialize |
| `--remove-all` | - | Remove all orphans (backup available) | ari sync materialize |
| `--promote-all` | - | Move all orphans to user-level | ari sync materialize |

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

Full documentation: `rites/intelligence/mena/intelligence-ref/INDEX.lego.md`

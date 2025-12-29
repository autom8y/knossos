---
description: Quick switch to intelligence-pack (product analytics workflow)
allowed-tools: Bash, Read
model: claude-sonnet-4-5
---

## Context

Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the Product Intelligence Team pack and display the team roster.

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh intelligence-pack`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `intelligence-pack`

## When to Use

- Instrumenting new features with analytics
- Designing and running A/B tests
- User research and interview planning
- Data-driven product decisions

## Workflow

```
instrumentation → research → experimentation → synthesis
```

## Reference

Full documentation: `.claude/skills/intelligence-ref/skill.md`

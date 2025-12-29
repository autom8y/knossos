---
description: Quick switch to strategy-pack (business strategy workflow)
allowed-tools: Bash, Read
model: sonnet
---

## Context

Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the Strategy Team pack and display the team roster.

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh strategy-pack`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `strategy-pack`

## When to Use

- Market analysis and sizing
- Competitive intelligence gathering
- Pricing and business model analysis
- Strategic roadmap planning

## Workflow

```
market-research → competitive-analysis → business-modeling → strategic-planning
```

## Reference

Full documentation: `.claude/skills/strategy-ref/skill.md`

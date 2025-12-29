---
description: Quick switch to rnd-pack (innovation lab workflow)
allowed-tools: Bash, Read
model: sonnet
---

## Context

Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the Innovation Lab (R&D) pack and display the team roster.

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh rnd-pack`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `rnd-pack`

## When to Use

- Evaluating new technologies
- Building proof-of-concept prototypes
- Long-term architecture planning
- Innovation and R&D exploration

## Workflow

```
scouting → integration-analysis → prototyping → future-architecture
```

## Reference

Full documentation: `.claude/skills/rnd-ref/skill.md`

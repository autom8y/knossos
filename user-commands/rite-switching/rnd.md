---
description: Quick switch to rnd (innovation lab workflow)
argument-hint: [--force]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Innovation Lab (R&D) pack and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite rnd $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `rnd`

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

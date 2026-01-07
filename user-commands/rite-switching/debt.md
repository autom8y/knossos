---
description: Quick switch to debt-triage (technical debt workflow)
argument-hint: [--force]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the technical debt triage rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite debt-triage $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `debt-triage`

## When to Use

- Technical debt inventory
- Prioritizing debt paydown
- Sprint planning for maintenance
- Risk assessment of shortcuts

## Reference

Full documentation: `.claude/skills/debt-ref/skill.md`

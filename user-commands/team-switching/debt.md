---
description: Quick switch to debt-triage-pack (technical debt workflow)
allowed-tools: Bash, Read
model: sonnet
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the technical debt triage team pack and display the team roster.

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh debt-triage-pack`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `debt-triage-pack`

## When to Use

- Technical debt inventory
- Prioritizing debt paydown
- Sprint planning for maintenance
- Risk assessment of shortcuts

## Reference

Full documentation: `.claude/skills/debt-ref/skill.md`

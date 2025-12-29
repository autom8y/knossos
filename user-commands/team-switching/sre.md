---
description: Quick switch to sre-pack (reliability workflow)
allowed-tools: Bash, Read
model: sonnet
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the SRE team pack and display the team roster.

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh sre-pack`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `sre-pack`

## When to Use

- System reliability improvements
- Observability and monitoring work
- Incident response preparation
- Chaos engineering experiments
- Platform and infrastructure work

## Reference

Full documentation: `.claude/skills/sre-ref/skill.md`

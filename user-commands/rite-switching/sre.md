---
description: Quick switch to sre (reliability workflow)
argument-hint: [--force]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the SRE rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite sre $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `sre`

## When to Use

- System reliability improvements
- Observability and monitoring work
- Incident response preparation
- Chaos engineering experiments
- Platform and infrastructure work

## Reference

Full documentation: `.claude/skills/sre-ref/skill.md`

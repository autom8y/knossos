---
description: Quick switch to hygiene (code quality workflow)
argument-hint: [--force]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the code hygiene rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite hygiene $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `hygiene`

## When to Use

- Code quality audits
- Refactoring initiatives
- Reducing technical debt
- Enforcing architectural patterns

## Reference

Full documentation: `.claude/skills/hygiene-ref/skill.md`

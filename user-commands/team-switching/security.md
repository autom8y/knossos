---
description: Quick switch to security-pack (security assessment workflow)
allowed-tools: Bash, Read
model: claude-sonnet-4-5
---

## Context

Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the Security Team pack and display the team roster.

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh security-pack`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `security-pack`

## When to Use

- Security review of new features or changes
- Compliance mapping and audit preparation
- Penetration testing and vulnerability assessment
- Pre-release security signoff

## Workflow

```
threat-modeling → compliance-design → penetration-testing → security-review
```

## Reference

Full documentation: `.claude/skills/security-ref/skill.md`

---
name: security
description: Quick switch to security (security assessment workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
disable-model-invocation: true
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Security rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite security $ARGUMENTS`
2. Display the roster output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `security`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | ari sync materialize |
| `--dry-run` | - | Preview changes without applying | ari sync materialize |
| `--keep-all` | - | Preserve all orphan agents in project | ari sync materialize |
| `--remove-all` | - | Remove all orphans (backup available) | ari sync materialize |
| `--promote-all` | - | Move all orphans to user-level | ari sync materialize |

## When to Use

- Security review of new features or changes
- Compliance mapping and audit preparation
- Penetration testing and vulnerability assessment
- Pre-release security signoff

## Workflow

```
threat-modeling -> compliance-design -> penetration-testing -> security-review
```

## Reference

Full documentation: `rites/security/mena/security-ref/INDEX.lego.md`

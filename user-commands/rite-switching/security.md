---
description: Quick switch to security-pack (security assessment workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Security Team pack and display the rite roster. $ARGUMENTS

## Behavior

1. Execute: `${KNOSSOS_HOME:-~/Code/roster}/swap-rite.sh security-pack $ARGUMENTS`
2. Display the roster output from swap-rite.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `security-pack`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | swap-rite.sh |
| `--dry-run` | - | Preview changes without applying | swap-rite.sh |
| `--keep-all` | - | Preserve all orphan agents in project | swap-rite.sh |
| `--remove-all` | - | Remove all orphans (backup available) | swap-rite.sh |
| `--promote-all` | - | Move all orphans to user-level | swap-rite.sh |

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

Full documentation: `.claude/skills/security-ref/skill.md`

---
name: hygiene
description: "Switch to hygiene rite (code quality workflow). Use when: user says /hygiene, wants code quality, refactoring, quality audit, code cleanup. Triggers: /hygiene, code quality, refactoring rite, quality audit."
context: fork
---

# /hygiene - Switch to Code Hygiene Rite

Switch to hygiene, the code quality and refactoring rite.

## Behavior

### 1. Invoke Rite Switch

Execute via Bash tool:

```bash
ari sync --rite hygiene
```

### 2. Display Knossos

After successful switch, show the agent table:

| Agent | Role |
|-------|------|
| code-smeller | Detects code smells and anti-patterns |
| architect-enforcer | Validates architectural compliance |
| janitor | Cleans up code, refactors for quality |
| audit-lead | Conducts comprehensive quality audits |

### 3. Update Session

If a session is active, update `active_rite` to `hygiene`.

## When to Use

- Code quality audits and health checks
- Refactoring initiatives and complexity reduction
- Architecture compliance enforcement
- Pre-release cleanup, post-implementation cleanup

**Don't use for**: Feature development --> `/10x` | Documentation --> `/docs` | Debt assessment (planning) --> `/debt`

**Hygiene vs Debt**: Use `/debt` to plan and prioritize. Use `/hygiene` to execute cleanup.

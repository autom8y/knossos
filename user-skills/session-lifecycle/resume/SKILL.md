---
name: resume
description: "Resume a parked work session by restoring full context. Use when: returning after /park, continuing after a break, picking up work after blocker resolution. Triggers: /resume, continue session, resume work, unpause session, restore session."
---

# /resume - Resume Parked Session

> Restore context, validate environment, invoke agent to continue work.

## Decision Tree

```
Have a parked session?
├─ Yes, returning to work → /resume
├─ No parked session → /start new session
├─ Want different agent → /resume --agent=NAME
└─ Session is stale (weeks old) → Consider /wrap, then /start fresh
```

## Usage

```bash
/resume [--agent=NAME]
```

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `--agent` | No | `last_agent` | Override which agent to continue with |

## Quick Reference

**Pre-flight**: Session exists and is parked (`parked_at` set)

**Actions**:
1. Load and display session summary
2. Validate team consistency
3. Check git status changes
4. Confirm/select agent
5. Invoke moirai agent for resume mutation (removes park metadata, adds resume metadata)
6. Invoke selected agent with full context

**State Changes**:
| Field | Action |
|-------|--------|
| `parked_*` fields | REMOVED |
| `resumed_at` | ADDED |
| `resume_count` | INCREMENTED |
| `last_agent` | UPDATED |

## Validation Checks

| Check | Warning If | Action |
|-------|-----------|--------|
| Team consistency | `ACTIVE_RITE` ≠ `session.active_rite` | Offer team switch |
| Git status | New changes since park | Offer diff review |
| Agent availability | Agent not in team | Error with valid list |

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Resume after major code changes | Session context stale | Consider /wrap, /start fresh |
| Ignore team mismatch warning | Wrong agents invoked | Switch to session's team |
| Resume stale session (weeks old) | Context drift | Review PRD/TDD, consider fresh start |
| Skip git status review | May conflict with session work | Review diff before continuing |

## Prerequisites

- Parked session exists (`parked_at` field set)
- Target agent exists in team (or team can be switched)

## Success Criteria

- Park metadata removed from SESSION_CONTEXT
- Agent invoked with full context
- User sees summary and next steps
- Team and git issues surfaced

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/start` | No parked session exists |
| `/park` | Pause again after resuming |
| `/status` | Check state without resuming |
| `/handoff` | Switch agents after resuming |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence
- [examples.md](examples.md) - Usage scenarios
- [validation-checks.md](validation-checks.md) - Team and git consistency
- [session-context-schema](../../session-common/session-context-schema.md) - Field definitions
- [session-validation](../../session-common/session-validation.md) - Pre-flight patterns

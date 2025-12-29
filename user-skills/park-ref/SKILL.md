---
name: park-ref
description: "Pause current work session and preserve state for later resumption. Use when: taking a break, waiting for external input, encountering blockers, switching context temporarily. Triggers: /park, pause session, save session, park work, suspend session, pause work."
---

# /park - Pause Work Session

> Preserve session state for later resumption via `/resume`.

## Decision Tree

```
Need to stop working?
├─ Temporary pause (returning soon) → /park
├─ Work complete → /wrap
├─ Switching agents mid-session → /handoff
└─ Abandoning session → /wrap --skip-checks
```

## Usage

```bash
/park [reason]
```

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `reason` | No | "Manual park" | Why work is being paused |

## Quick Reference

**Pre-flight**: Session exists, not already parked

**Actions**:
1. Capture git status, phase, artifacts, blockers
2. Generate parking summary
3. Add park metadata to SESSION_CONTEXT frontmatter
4. Append parking summary to SESSION_CONTEXT body
5. Display confirmation with resume instructions

**State Changes**:
| Field | Value |
|-------|-------|
| `parked_at` | ISO 8601 timestamp |
| `parked_reason` | User-provided or "Manual park" |
| `parked_phase` | Current phase at park time |
| `parked_git_status` | clean \| dirty |

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Park with uncommitted changes indefinitely | Stale work, merge conflicts | Commit or stash before extended park |
| Park without reason | Context loss on resume | Always provide descriptive reason |
| Multiple consecutive parks | Indicates scope/blocker issues | Address blockers before re-parking |
| Park to avoid quality gates | Defeats workflow purpose | Use /wrap --skip-checks if truly needed |

## Prerequisites

- Active session exists
- Session not already parked

## Success Criteria

- SESSION_CONTEXT updated with park metadata
- Parking summary appended
- User receives confirmation with /resume instructions

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/resume` | Continue this parked session |
| `/status` | Check session state |
| `/wrap` | Complete session (not pause) |
| `/start` | Begin new session |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence
- [examples.md](examples.md) - Usage scenarios
- [parking-summary.md](parking-summary.md) - Summary template
- [session-context-schema](../session-common/session-context-schema.md) - Field definitions
- [session-validation](../session-common/session-validation.md) - Pre-flight patterns

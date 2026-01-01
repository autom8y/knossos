---
name: handoff-ref
description: "Transfer work from one agent to another within active session. Use when: transitioning between workflow phases, current agent work is complete, specialist expertise needed. Triggers: /handoff, switch agent, transfer work, change agent, agent handoff."
---

# /handoff - Transfer Work Between Agents

> Generate handoff note, update SESSION_CONTEXT, invoke target agent.

## Decision Tree

```
Need to switch agents?
├─ Phase transition (design → implementation) → /handoff engineer
├─ Session is parked → /resume first, then /handoff
├─ Want to pause work → /park (not /handoff)
└─ Want to switch teams → /team, then new /start
```

## Usage

```bash
/handoff <agent-name> [note]
```

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `agent-name` | Yes | N/A | Target agent to hand off to |
| `note` | No | Auto-generated | Custom handoff context |

### Valid Agent Names

| Name | Aliases | Phase |
|------|---------|-------|
| requirements-analyst | analyst | requirements |
| architect | - | design |
| principal-engineer | engineer | implementation |
| qa-adversary | qa | validation |

## Quick Reference

**Pre-flight**: Session active (not parked), target agent exists, not same agent

**Actions**:
1. Generate handoff note with artifacts, decisions, blockers
2. Append note to SESSION_CONTEXT body
3. Update metadata (last_agent, handoff_count, phase)
4. Invoke target agent via Task tool
5. Display confirmation

**State Changes**:
| Field | Value |
|-------|-------|
| `last_agent` | Target agent name |
| `handoff_count` | Incremented |
| `last_handoff_at` | ISO timestamp |
| `current_phase` | Inferred from agent |

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Handoff to same agent | No-op, wastes tokens | Continue working with current agent |
| Handoff without artifacts | Next agent lacks context | Ensure artifacts exist first |
| Excessive handoffs (>6) | Ping-pong indicates scope issues | Re-scope with analyst |
| Handoff while parked | Agent lacks restoration context | Use /resume first |

## Prerequisites

- Active session (not parked)
- Target agent exists in current team

## Success Criteria

- Handoff note generated with full context
- SESSION_CONTEXT updated with metadata
- Target agent invoked with session context
- User receives confirmation

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/resume` | Required before handoff if parked |
| `/status` | Check current agent and phase |
| `/wrap` | Complete session (not handoff) |
| `/roster` | List available agents |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence
- [examples.md](examples.md) - Usage scenarios
- [handoff-notes.md](handoff-notes.md) - Transition-specific templates
- [session-context-schema](../../session-common/session-context-schema.md) - Field definitions
- [session-phases](../../session-common/session-phases.md) - Phase transitions

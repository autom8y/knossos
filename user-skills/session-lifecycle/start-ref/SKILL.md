---
name: start-ref
description: "Begin a new work session capturing initiative, complexity, and team context. Use when: starting new feature work, initializing tracked workflow, beginning fresh development session. Triggers: /start, new session, begin work, kickoff, start session, initialize session."
---

# /start - Initialize Work Session

> Create SESSION_CONTEXT and invoke Requirements Analyst (+ Architect for MODULE+).

## Decision Tree

```
Starting work?
├─ New feature/task → /start
├─ Continuing parked work → /resume
├─ Complex initiative needing deliberation → initiative-scoping skill first
└─ Quick one-off (no tracking) → Work directly, no session
```

## Usage

```bash
/start [initiative-name] [--complexity=LEVEL] [--team=PACK]
```

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `initiative-name` | No* | Prompted | Name of feature/task |
| `--complexity` | No* | Prompted | SCRIPT \| MODULE \| SERVICE \| PLATFORM |
| `--team` | No | ACTIVE_TEAM | Team pack for session |

*If not provided, user will be prompted interactively.

## Complexity Levels

| Level | Scope | Agents Invoked |
|-------|-------|----------------|
| SCRIPT | Single file, < 200 LOC | Analyst only |
| MODULE | Multiple files, < 2000 LOC | Analyst → Architect |
| SERVICE | APIs, data persistence | Analyst → Architect |
| PLATFORM | Multiple services, infra | Analyst → Architect |

## Quick Reference

**Pre-flight**: No existing session, team pack valid

**Actions**:
1. Gather parameters (initiative, complexity, team)
2. Create SESSION_CONTEXT with initial metadata
3. Invoke Requirements Analyst → PRD
4. Invoke Architect → TDD + ADRs (if MODULE+)
5. Update SESSION_CONTEXT with artifacts
6. Display confirmation with next steps

**Creates**:
- `.claude/sessions/{session_id}/SESSION_CONTEXT.md`
- `/docs/requirements/PRD-{slug}.md`
- `/docs/design/TDD-{slug}.md` (if MODULE+)

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Start without wrapping previous session | Creates orphaned sessions | Run `/wrap` or check `/status` first |
| Under-classify complexity | Skips design phase, causes rework | When uncertain, classify one level higher |
| Start PLATFORM work in single session | Too large, loses context | Break into multiple MODULE sessions |
| Ignore team context | Wrong agents for the work | Verify team with `/team` before starting |

## Prerequisites

- No existing active session
- Target team pack exists (if specified)

## Success Criteria

- SESSION_CONTEXT file created
- PRD produced and saved
- TDD + ADRs produced (if complexity > SCRIPT)
- User receives confirmation with next steps

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/park` | Pause this session |
| `/resume` | Continue parked session (not /start) |
| `/handoff` | Switch agents mid-session |
| `/wrap` | Complete session |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence
- [examples.md](examples.md) - Usage scenarios
- [integration.md](integration.md) - Agent delegation templates
- [session-context-schema](../../session-common/session-context-schema.md) - Field definitions
- [session-phases](../../session-common/session-phases.md) - Phase transitions

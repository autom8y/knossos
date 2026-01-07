---
name: start-ref
description: "Begin a new work session capturing initiative, complexity, and rite context. Use when: starting new feature work, initializing tracked workflow, beginning fresh development session. Triggers: /start, new session, begin work, kickoff, start session, initialize session."
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
/start [initiative-name] [--complexity=LEVEL] [--rite=NAME] [--no-rite]
```

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `initiative-name` | No* | Prompted | Name of feature/task |
| `--complexity` | No* | Prompted | SCRIPT \| MODULE \| SERVICE \| PLATFORM |
| `--rite` | No | ACTIVE_RITE | Rite for session |
| `--no-rite` | No | false | Create cross-cutting session (no orchestration) |

*If not provided, user will be prompted interactively.

**Note**: When `--no-rite` is specified or no rite is active, the session operates in **cross-cutting mode**: direct execution with session tracking, no orchestrator required. See `orchestration/execution-mode.md`.

## Complexity Levels

| Level | Scope | Agents Invoked |
|-------|-------|----------------|
| SCRIPT | Single file, < 200 LOC | Analyst only |
| MODULE | Multiple files, < 2000 LOC | Analyst → Architect |
| SERVICE | APIs, data persistence | Analyst → Architect |
| PLATFORM | Multiple services, infra | Analyst → Architect |

## Quick Reference

**Pre-flight**: No existing session, rite valid

**Actions**:
1. Gather parameters (initiative, complexity, team)
2. Delegate to Moirai (Clotho - the Spinner) to create session
3. Invoke Requirements Analyst → PRD
4. Invoke Architect → TDD + ADRs (if MODULE+)
5. Delegate to Moirai to update session with artifacts
6. Display confirmation with next steps

**Creates** (via Moirai):
- `.claude/sessions/{session_id}/SESSION_CONTEXT.md`
- `/docs/requirements/PRD-{slug}.md`
- `/docs/design/TDD-{slug}.md` (if MODULE+)

## State Mutation

All session state mutations are delegated to the **Moirai** (the Fates). Direct writes to `SESSION_CONTEXT.md` are prohibited.

**Session Creation** (Clotho - the Spinner):
```
Task(moirai, "create_session initiative='...' complexity=... rite=...

Session Context:
- New session requested
- Initiative: {user-provided}
- Complexity: {determined}
- Team: {active-rite}")
```

**Session Updates** (Lachesis - the Measurer):
```
Task(moirai, "update_session session_id=... artifacts=[{type: PRD, path: ...}]

Session Context:
- Session ID: {session-id}
- Artifacts produced: PRD, TDD
- Phase transition: requirements → design")
```

The Moirai enforce:
- Schema validation
- Lifecycle transitions
- Audit trails
- Atomic state changes

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Start without wrapping previous session | Creates orphaned sessions | Run `/wrap` or check `/status` first |
| Under-classify complexity | Skips design phase, causes rework | When uncertain, classify one level higher |
| Start PLATFORM work in single session | Too large, loses context | Break into multiple MODULE sessions |
| Ignore rite context | Wrong agents for the work | Verify team with `/team` before starting |

## Prerequisites

- No existing active session
- Target rite exists (if specified)

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

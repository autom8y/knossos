# Moirai Invocation Pattern

> Centralized session state mutation through the unified Moirai agent.

## Overview

**Moirai** is the unified session lifecycle agent and sole authority for `SESSION_CONTEXT.md` and `SPRINT_CONTEXT.md` mutations. All state transitions and context updates MUST go through this agent via the Task tool.

**Architecture**: Moirai embodies the three Fates as internal skills (not separate agents):
- **Clotho** (creation): `create_sprint`, `start_sprint`
- **Lachesis** (measurement): `mark_complete`, `transition_phase`, `park_session`, etc.
- **Atropos** (termination): `wrap_session`, `generate_sails`, `delete_sprint`

**Agent location**: `.claude/agents/moirai.md` (materialized per rite; may not exist in all rites)
**Skills location**: `.claude/skills/session/moirai/` (INDEX.md, clotho.md, lachesis.md, atropos.md)

## When to Invoke

Use Moirai for all session lifecycle operations:

| Operation | Command | Description |
|-----------|---------|-------------|
| Park session | `park_session` | Pause active work with reason |
| Resume session | `resume_session` | Resume parked session |
| Wrap session | `wrap_session` | Complete and finalize session |
| Update context | `update_field` | Modify session context fields |
| Sprint state | Various | Sprint lifecycle mutations |

## Invocation Template

```markdown
Task(moirai, "{operation} {parameters}

Session Context:
- Session ID: {session-id}
- Session Path: .claude/sessions/{session-id}/SESSION_CONTEXT.md
")
```

### Parameters

| Parameter | Description | Required |
|-----------|-------------|----------|
| `operation` | Moirai operation name | Yes |
| `parameters` | Operation-specific arguments | Varies |
| Session ID | Active session identifier | Yes |
| Session Path | Path to SESSION_CONTEXT.md | Yes |

## Common Operations

### Park Session

```markdown
Task(moirai, "park_session reason='Need to context switch to hotfix'

Session Context:
- Session ID: session-20260106-123456-abc123
- Session Path: .claude/sessions/session-20260106-123456-abc123/SESSION_CONTEXT.md
")
```

### Resume Session

```markdown
Task(moirai, "resume_session

Session Context:
- Session ID: session-20260106-123456-abc123
- Session Path: .claude/sessions/session-20260106-123456-abc123/SESSION_CONTEXT.md
")
```

### Wrap Session

```markdown
Task(moirai, "wrap_session

Session Context:
- Session ID: session-20260106-123456-abc123
- Session Path: .claude/sessions/session-20260106-123456-abc123/SESSION_CONTEXT.md
")
```

### Update Context Field

```markdown
Task(moirai, "update_field current_phase='implementation'

Session Context:
- Session ID: session-20260106-123456-abc123
- Session Path: .claude/sessions/session-20260106-123456-abc123/SESSION_CONTEXT.md
")
```

## Expected Response

Moirai returns structured JSON:

```json
{
  "success": true,
  "operation": "park_session",
  "session_id": "session-20260106-123456-abc123",
  "state_before": "ACTIVE",
  "state_after": "PARKED",
  "timestamp": "2026-01-06T12:34:56Z"
}
```

### Success Response
- `success: true`
- State transition confirmed
- Timestamp recorded
- Context file updated

### Error Response
- `success: false`
- `error_code`: Machine-readable error type
- `error_message`: Human-readable description
- `details`: Additional context

Common error codes:
- `SESSION_NOT_FOUND`: Invalid session ID
- `INVALID_STATE_TRANSITION`: Illegal FSM transition
- `SCHEMA_VALIDATION_FAILED`: Context update violates schema
- `FILE_WRITE_ERROR`: Filesystem operation failed

## Response Handling

### 1. Check Success Flag

```
if response.success:
    proceed with confirmation
else:
    handle error based on error_code
```

### 2. Verify State Transition

Read SESSION_CONTEXT.md to confirm state change:

```markdown
Read(/path/to/SESSION_CONTEXT.md)
```

### 3. Confirm to User

Display operation result:
- What changed (state transition, field update)
- Current session state
- Next steps if applicable

## Duration

Moirai operations complete in < 1 second (typically < 200ms).

## Anti-Patterns

❌ **Direct file writes**
```markdown
# WRONG - blocked by PreToolUse hook
Write(/path/to/SESSION_CONTEXT.md, content)
```

❌ **Skipping Moirai**
```markdown
# WRONG - state mutation without authority
Edit(SESSION_CONTEXT.md, old, new)
```

✅ **Correct delegation**
```markdown
# CORRECT - delegate to Moirai
Task(moirai, "update_field ...")
```

## Implementation Details

### Hook Enforcement

The PreToolUse hook intercepts direct Write/Edit calls to `*_CONTEXT.md` files and redirects to Moirai invocation pattern.

### Schema Validation

Moirai validates all mutations against JSON Schema before applying changes. Invalid updates are rejected with `SCHEMA_VALIDATION_FAILED`.

### Audit Trail

All mutations are logged to `audit/session-mutations.log` with:
- Timestamp
- Operation
- Session ID
- State transition
- Invoking agent

## Backward Compatibility

The `state-mate` alias is supported for backward compatibility:

```markdown
# Both work
Task(moirai, "...")
Task(moirai, "...")
```

**DEPRECATED**: The `state-mate` alias is deprecated as of 2026-01-06. New code should use `moirai`. The alias will be removed in a future major version.

## Cross-References

- [Session State Machine](../session-common/session-state-machine.md) - Valid state transitions
- [Session Context Schema](../session-common/session-context-schema.md) - Field definitions
- [Agent Delegation](../session-common/agent-delegation.md) - General delegation patterns
- ADR-0005-state-mate-centralized-state-authority.md - Design rationale
- [Knossos Doctrine](../../../docs/philosophy/knossos-doctrine.md) - Moirai mythology

## See Also

- `.claude/agents/moirai.md` - Full Moirai agent specification (materialized per rite; may not exist in all rites)
- `.claude/skills/session/moirai/` - Fate skills (INDEX.md, clotho.md, lachesis.md, atropos.md)
- `.claude/hooks/PreToolUse.sh` - Direct write prevention
- `schemas/artifacts/session-context.schema.json` - Validation schema

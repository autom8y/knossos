# Moirai Invocation Pattern

> Delegate session state mutations to Moirai (the Fates) via Task tool.
>
> **Note**: `state-mate` is a backward-compatible alias for `moirai`.

## When to Apply

Commands that mutate SESSION_CONTEXT:
- /park - sets parked_at, parked_reason, etc.
- /resume - clears park fields, sets resumed_at
- /wrap - transitions to ARCHIVED state

## Task Invocation Template

```
Task(moirai, "{operation}

Session Context:
- Session ID: {session_id}
- Session Path: .claude/sessions/{session_id}/SESSION_CONTEXT.md")
```

**Note**: `state-mate` can be used interchangeably with `moirai`.

### Operations and Fate Routing

| Operation | Command | Fate | Mutations |
|-----------|---------|------|-----------|
| `park_session reason='{reason}'` | /park | Lachesis | Set parked_at, parked_reason |
| `resume_session` | /resume | Lachesis | Clear parked_*, set resumed_at |
| `wrap_session` | /wrap | Atropos | Set completed_at, archive |

The Moirai router automatically delegates to the appropriate Fate based on operation semantics:
- **Lachesis** (the Measurer): Session lifecycle transitions (park, resume, handoff)
- **Atropos** (the Cutter): Session termination (wrap, generate_sails)

### WHITE_SAILS Quality Gate (wrap_session only)

When invoking `wrap_session`, Atropos generates a WHITE_SAILS confidence signal and enforces a quality gate:

| Sails Color | Meaning | Wrap Behavior |
|-------------|---------|---------------|
| **WHITE** | All proofs pass, no open questions | Wrap succeeds |
| **GREY** | Some proofs missing or open questions exist | Wrap succeeds with warning |
| **BLACK** | Critical failures or major blockers | **Wrap BLOCKED** |

**To override BLACK sails** (e.g., for emergency hotfixes):
```
Task(moirai, "wrap_session --emergency=reason=\"Hotfix deployment, will revert if unsuccessful\"

Session Context:
- Session ID: {session_id}
- Session Path: {session_path}")
```

See [Atropos agent](../../../user-agents/atropos.md) for full WHITE_SAILS algorithm and proof collection.

## Response Handling

### Success Response

```json
{
  "success": true,
  "operation": "{operation_name}",
  "message": "Session {operation} successfully",
  "state_before": { "session_state": "..." },
  "state_after": { "session_state": "...", ... }
}
```

**Action**: Parse response, display confirmation to user.

### Failure Response

```json
{
  "success": false,
  "error_type": "LIFECYCLE_VIOLATION",
  "message": "Cannot {operation}: {reason}",
  "hint": "Use /{suggested_command} first"
}
```

**Action**: Surface error message and hint to user.

## Error Types

| Error Type | Cause | Recovery |
|------------|-------|----------|
| `LIFECYCLE_VIOLATION` | Invalid state transition | Follow hint (e.g., resume before wrap) |
| `VALIDATION_ERROR` | Missing required field | Provide missing data |
| `UNAVAILABLE` | Moirai not responding | Retry or check agent configuration |

## Implementation

```
1. Get session context:
   session_id=$(session-manager.sh status | jq -r '.session_id')
   session_path=".claude/sessions/${session_id}/SESSION_CONTEXT.md"

2. Invoke Moirai via Task tool:
   Task(moirai, "{operation}

   Session Context:
   - Session ID: {session_id}
   - Session Path: {session_path}")

   # Or using state-mate alias (backward compatible)
   Task(state-mate, "{operation}

   Session Context:
   - Session ID: {session_id}
   - Session Path: {session_path}")

3. Parse JSON response:
   - If success: true → Extract state_after, continue
   - If success: false → Extract message and hint, surface to user

4. Post-operation (command-specific):
   - /park: Display parking summary
   - /resume: Invoke selected agent
   - /wrap: Archive session directory
```

## Customization Points

| Parameter | Description | Commands Using |
|-----------|-------------|----------------|
| `operation` | state-mate operation name | All |
| `reason` | User-provided reason | park |
| `post_action` | Action after successful mutation | All |

## Cross-Reference

- Moirai router agent: `user-agents/moirai.md`
- Lachesis (Measurer): `user-agents/lachesis.md`
- Atropos (Cutter): `user-agents/atropos.md`
- Clotho (Spinner): `user-agents/clotho.md`
- Knossos Doctrine: `docs/philosophy/knossos-doctrine.md`
- ADR: `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md`

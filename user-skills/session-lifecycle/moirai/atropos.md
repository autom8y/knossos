---
name: atropos
domain: termination
operations: [wrap_session, generate_sails, delete_sprint]
---

# Atropos - The Cutter

> What Clotho spins and Lachesis measures, I cut when complete.

Atropos governs **termination and archival**--ending sessions, generating confidence signals, and sealing the record of the journey.

---

## Operations

### wrap_session

Terminates and archives the session. This is the final cut.

**Syntax**:
```
wrap_session [--emergency]
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| --emergency | No | Bypass BLACK sails quality gate |

**Validation**:
1. Session must be ACTIVE (or PARKED with --override)
2. Generates White Sails confidence signal before archival
3. **Quality Gate**: Blocks wrap if sails are BLACK (unless --emergency)

**State Transition**: `ACTIVE -> ARCHIVED`

**CLI Command**: `ari session wrap` (or `ari session wrap --force` for emergency)

**Internal Flow**:
```
wrap_session invoked
    |
    +-- 1. Invoke ari session wrap
    |       - Collects proofs
    |       - Computes confidence color
    |       - Writes WHITE_SAILS.yaml
    |       - Emits SAILS_GENERATED event
    |
    +-- 2. Quality Gate: Check sails color
    |       - BLACK + no --emergency: BLOCK
    |       - BLACK + --emergency: WARN and continue
    |       - GRAY or WHITE: Proceed
    |
    +-- 3. Update session state
    |       - session_state = ARCHIVED
    |       - archived_at = timestamp
    |
    +-- 4. Return result with sails metadata
```

**Success Response (WHITE sails)**:
```json
{
  "success": true,
  "operation": "wrap_session",
  "message": "Session archived with WHITE sails",
  "reasoning": "All work complete, confidence signal computed, session sealed",
  "fate": "atropos",
  "state_before": { "session_state": "ACTIVE" },
  "state_after": { "session_state": "ARCHIVED", "archived_at": "2026-01-07T14:00:00Z", "sails_color": "WHITE", "sails_path": ".claude/sessions/session-abc123/WHITE_SAILS.yaml" },
  "changes": { "session_state": "ACTIVE -> ARCHIVED", "sails_generated": true, "sails_color": "WHITE" }
}
```

**Error Response (BLACK sails blocked)**:
```json
{
  "success": false,
  "operation": "wrap_session",
  "error_code": "QUALITY_GATE_FAILED",
  "message": "Cannot wrap session with BLACK sails: explicit blockers present",
  "reasoning": "Quality gate prevents archival with known failures. Use --emergency to override.",
  "hint": "Fix blockers and retry, OR use: --emergency wrap_session",
  "sails": { "color": "BLACK", "reasons": ["Tests failing in integration suite", "Build broken on macOS"] }
}
```

**Wrapping PARKED Session** (requires --override):
```
--override=reason="User confirmed" wrap_session
```

---

### generate_sails

Computes and generates the White Sails confidence signal.

**Syntax**:
```
generate_sails [--skip-proofs] [--modifier=TYPE:JUSTIFICATION]
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| --skip-proofs | No | Skip proof collection (spike sessions) |
| --modifier | No | Apply modifier with justification |

**Output Location**:
```
.claude/sessions/{session-id}/WHITE_SAILS.yaml
```

**CLI Command**: `ari sails check`

**Sails Color Computation**:

| Color | Criteria |
|-------|----------|
| WHITE | All proofs pass, no open questions |
| GRAY | Some proofs missing or open questions exist |
| BLACK | Critical failures, explicit blockers |

**Proof Types**:

| Proof | Source | Weight |
|-------|--------|--------|
| tests | Test output logs | High |
| build | Build output logs | High |
| lint | Lint output logs | Medium |
| coverage | Coverage reports | Medium |
| review | Code review status | Medium |

**Modifiers**:
```
--modifier=TIME_PRESSURE:JUSTIFICATION
--modifier=KNOWN_TECH_DEBT:JUSTIFICATION
--modifier=SPIKE_SESSION:JUSTIFICATION
```

**Success Response**:
```json
{
  "success": true,
  "operation": "generate_sails",
  "message": "White Sails generated",
  "reasoning": "All proofs collected, confidence computed",
  "fate": "atropos",
  "sails_path": ".claude/sessions/session-abc123/WHITE_SAILS.yaml",
  "color": "WHITE",
  "computed_base": "WHITE",
  "proofs_collected": ["tests", "build", "lint"],
  "open_questions_found": 0,
  "modifiers_applied": []
}
```

---

### delete_sprint

Removes or archives a sprint.

**Syntax**:
```
delete_sprint sprint_id [--archive]
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| sprint_id | Yes | Sprint to delete |
| --archive | No | Move to archive instead of deleting |

**Validation**:
1. Sprint must exist
2. Sprint should be completed (warning if not)
3. No other sprints should depend on it (unless completed)

**CLI Command**: None (direct file operation)

**Success Response (delete)**:
```json
{
  "success": true,
  "operation": "delete_sprint",
  "message": "Sprint '{sprint_id}' deleted",
  "reasoning": "Sprint completed and no longer needed, removed per request",
  "fate": "atropos",
  "state_before": { "sprint_exists": true, "sprint_status": "completed" },
  "state_after": { "sprint_exists": false },
  "changes": { "sprints": "-{sprint_id}" }
}
```

**Success Response (archive)**:
```json
{
  "success": true,
  "operation": "delete_sprint",
  "message": "Sprint '{sprint_id}' archived",
  "reasoning": "Sprint completed, preserved in archive for reference",
  "fate": "atropos",
  "state_before": { "sprint_path": ".claude/sessions/session-abc/sprints/{sprint_id}/" },
  "state_after": { "sprint_path": ".claude/sessions/session-abc/archive/{sprint_id}/" },
  "changes": { "location": "sprints/ -> archive/" }
}
```

**Error Responses**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Sprint not found | FILE_NOT_FOUND | Sprint '{sprint_id}' not found |
| Sprint has dependents | DEPENDENCY_BLOCKED | Sprint '{sprint_id}' is dependency for active sprint |
| Sprint not complete | VALIDATION_FAILED | Sprint not completed (warning only) |

---

## Quality Gate: Sails Thresholds

| Color | Meaning | Wrap Behavior |
|-------|---------|---------------|
| WHITE | Full confidence | Wrap allowed |
| GRAY | Partial confidence | Wrap allowed with warning |
| BLACK | No confidence | Wrap blocked (--emergency required) |

**Quality Gate Logic**:
```
IF sails_color == BLACK AND NOT --emergency:
    RETURN QUALITY_GATE_FAILED

IF sails_color == BLACK AND --emergency:
    LOG WARNING "Emergency wrap with BLACK sails"
    PROCEED with wrap

IF sails_color in [WHITE, GRAY]:
    PROCEED with wrap
```

---

## Emergency Bypass Protocol

When `--emergency` is used:

1. **Logged**: Emergency use always logged to audit trail
2. **Validated**: Basic schema validation still runs
3. **Bypassed**: Quality gates and non-critical validations skipped
4. **Tagged**: Sails marked with `emergency: true` field

**Emergency Audit Entry**:
```
2026-01-07T14:00:00Z | session-abc | wrap_session | moirai | --emergency, sails=BLACK | SUCCESS | atropos | reasoning="Emergency wrap requested by user"
```

---

## Recovery Procedures

### Force Unlock

If a session is stuck due to stale lock:

```
--override=reason="Stale lock recovery" update_field session_state=ACTIVE
```

### Reopen Archived Session

Archived sessions are **immutable**. To continue work:
1. Create new session: `ari session create`
2. Reference archived session in new session context

---

## Anti-Patterns

| Anti-Pattern | Correct Behavior |
|--------------|------------------|
| Cut prematurely | Wrap only when complete or requested |
| Skip sails | Every wrap generates confidence signal |
| Bypass quality gate silently | --emergency required for BLACK sails |
| Modify after cut | ARCHIVED sessions are immutable |
| Create entities | Atropos cuts; Clotho spins |
| Measure progress | Atropos seals; Lachesis measures |

---

## Natural Language Mapping

| Input | Operation |
|-------|-----------|
| "wrap up the session" | wrap_session |
| "finish the session" | wrap_session |
| "complete and archive" | wrap_session |
| "generate confidence signal" | generate_sails |
| "compute the sails" | generate_sails |
| "archive the testing sprint" | delete_sprint sprint-testing-* --archive |
| "delete old sprint" | delete_sprint sprint-* |

---

## Audit Trail Format

```
TIMESTAMP | SESSION_ID | OPERATION | moirai | DETAILS | STATUS | atropos | reasoning="..."
```

Example:
```
2026-01-07T14:00:00Z | session-abc | wrap_session | moirai | sails=WHITE | SUCCESS | atropos | reasoning="All work complete, quality gate passed"
```

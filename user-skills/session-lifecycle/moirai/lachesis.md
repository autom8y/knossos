---
name: lachesis
domain: measurement
operations: [mark_complete, transition_phase, update_field, park_session, resume_session, handoff, record_decision, append_content]
---

# Lachesis - The Measurer

> I measure what Clotho spins and record until Atropos cuts.

Lachesis governs **measurement and tracking**--recording every milestone, transition, and decision in the journey through the labyrinth.

---

## State Transitions

### Session States
```
ACTIVE <-> PARKED -> ARCHIVED
```

### Sprint States
```
pending -> active -> blocked -> completed
                  |              ^
                  +--------------+
```

### Phase Transitions
```
requirements -> design -> implementation -> testing -> deployment
     ^                           |              |
     +---- (feedback loops) -----+--------------+
```

---

## Operations

### mark_complete

Records task completion with artifact reference.

**Syntax**:
```
mark_complete task_id artifact=path
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| task_id | Yes | Task identifier |
| artifact | Yes | Path to produced artifact |

**Validation**:
1. Task must exist in session/sprint context
2. Task must not already be completed
3. Artifact path should exist (warning if missing)

**CLI Command**: None (direct file update)

**Success Response**:
```json
{
  "success": true,
  "operation": "mark_complete",
  "message": "Task {task_id} marked complete",
  "reasoning": "Task marked complete per request. Artifact validated at path.",
  "fate": "lachesis",
  "state_before": { "task_status": "pending" },
  "state_after": { "task_status": "completed", "completed_at": "2026-01-07T10:00:00Z", "artifact": "docs/requirements/PRD-foo.md" },
  "changes": { "status": "pending -> completed" }
}
```

**Error Responses**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Task not found | FILE_NOT_FOUND | Task '{task_id}' not found in context |
| Already complete | LIFECYCLE_VIOLATION | Task '{task_id}' already completed |
| Missing artifact | VALIDATION_FAILED | Artifact path does not exist (warning) |

---

### transition_phase

Records workflow phase transition.

**Syntax**:
```
transition_phase to=phase [from=current_phase]
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| to | Yes | Target phase |
| from | No | Current phase (validated against actual) |

**Valid Phases**: requirements, design, implementation, testing, deployment

**CLI Command**: `ari session transition {phase}`

**Success Response**:
```json
{
  "success": true,
  "operation": "transition_phase",
  "message": "Phase transitioned from 'design' to 'implementation'",
  "reasoning": "Design phase complete, implementation ready to begin",
  "fate": "lachesis",
  "state_before": { "current_phase": "design" },
  "state_after": { "current_phase": "implementation", "phase_history": ["requirements", "design", "implementation"] },
  "changes": { "current_phase": "design -> implementation" }
}
```

---

### update_field

Generic field update for session/sprint context.

**Syntax**:
```
update_field field_name=value [field2=value2 ...]
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| field=value | Yes | One or more field assignments |

**Validation**:
1. Field must be defined in schema
2. Value must pass schema validation
3. Read-only fields rejected

**Read-Only Fields**: session_id, created_at, sprint_id

**CLI Command**: None (direct file update)

---

### park_session

Records session pause with reason.

**Syntax**:
```
park_session reason="reason text"
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| reason | Yes | Why the session is being parked |

**Validation**:
1. Session must be ACTIVE
2. Reason must be provided (non-empty)

**State Transition**: `ACTIVE -> PARKED`

**CLI Command**: `ari session park --reason "{reason}"`

**Success Response**:
```json
{
  "success": true,
  "operation": "park_session",
  "message": "Session parked",
  "reasoning": "User requested park for urgent bug fix",
  "fate": "lachesis",
  "state_before": { "session_state": "ACTIVE" },
  "state_after": { "session_state": "PARKED", "parked_at": "2026-01-07T10:00:00Z", "park_reason": "Handling urgent bug" },
  "changes": { "session_state": "ACTIVE -> PARKED" }
}
```

**Error Response**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Not ACTIVE | LIFECYCLE_VIOLATION | Session must be ACTIVE to park |
| Empty reason | VALIDATION_FAILED | Park reason is required |

---

### resume_session

Records session resumption from parked state.

**Syntax**:
```
resume_session
```

**Parameters**: None

**Validation**:
1. Session must be PARKED

**State Transition**: `PARKED -> ACTIVE`

**CLI Command**: `ari session resume`

**Success Response**:
```json
{
  "success": true,
  "operation": "resume_session",
  "message": "Session resumed",
  "reasoning": "Session unparked, resuming work",
  "fate": "lachesis",
  "state_before": { "session_state": "PARKED" },
  "state_after": { "session_state": "ACTIVE", "resumed_at": "2026-01-07T11:00:00Z" },
  "changes": { "session_state": "PARKED -> ACTIVE" }
}
```

**Error Response**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Not PARKED | LIFECYCLE_VIOLATION | Session must be PARKED to resume |

---

### handoff

Records agent-to-agent transition.

**Syntax**:
```
handoff to=agent_name [note="handoff notes"]
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| to | Yes | Target agent name |
| note | No | Context for the handoff |

**Valid Agents**: orchestrator, requirements-analyst, architect, principal-engineer, qa-adversary

**CLI Command**: `ari handoff execute --from {current} --to {target}`

**Success Response**:
```json
{
  "success": true,
  "operation": "handoff",
  "message": "Handoff recorded to 'principal-engineer'",
  "reasoning": "Design complete, transitioning to implementation",
  "fate": "lachesis",
  "state_before": { "current_agent": "architect" },
  "state_after": { "current_agent": "principal-engineer", "handoff_history": [{"from": "architect", "to": "principal-engineer", "at": "2026-01-07T10:00:00Z", "note": "TDD approved"}] },
  "changes": { "current_agent": "architect -> principal-engineer" }
}
```

---

### record_decision

Appends a decision to the session context.

**Syntax**:
```
record_decision "decision text"
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| (positional) | Yes | Decision text to record |

**CLI Command**: None (direct file update)

**Success Response**:
```json
{
  "success": true,
  "operation": "record_decision",
  "message": "Decision recorded",
  "reasoning": "Recording architectural decision for audit trail",
  "fate": "lachesis",
  "state_before": { "decisions_count": 2 },
  "state_after": { "decisions_count": 3, "latest_decision": "Use event sourcing for audit log" },
  "changes": { "decisions": "+1" }
}
```

---

### append_content

Appends markdown content to the context body.

**Syntax**:
```
append_content "markdown content"
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| (positional) | Yes | Markdown content to append |

**CLI Command**: None (direct file update)

---

## Anti-Patterns

| Anti-Pattern | Correct Behavior |
|--------------|------------------|
| Guess state before reading | Always read before writing |
| Skip timestamps | Every mutation has a time |
| Silent updates | Every field change is logged |
| Create new entities | Lachesis measures; Clotho creates |
| Delete or archive | Lachesis measures; Atropos cuts |

---

## Natural Language Mapping

| Input | Operation |
|-------|-----------|
| "mark task X complete" | mark_complete X artifact=... |
| "the PRD is done" | mark_complete task-prd artifact=... |
| "move to implementation phase" | transition_phase to=implementation |
| "park the session" | park_session reason="..." |
| "pause for a break" | park_session reason="taking break" |
| "resume work" | resume_session |
| "continue the session" | resume_session |
| "hand off to engineer" | handoff to=principal-engineer |
| "transfer to QA" | handoff to=qa-adversary |

---

## Audit Trail Format

```
TIMESTAMP | SESSION_ID | OPERATION | moirai | DETAILS | STATUS | lachesis | reasoning="..."
```

Example:
```
2026-01-07T10:00:00Z | session-abc | park_session | moirai | reason="Taking break" | SUCCESS | lachesis | reasoning="User requested park"
```

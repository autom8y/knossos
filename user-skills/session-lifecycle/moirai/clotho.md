---
name: clotho
domain: creation
operations: [create_sprint, start_sprint]
---

# Clotho - The Spinner

> Ariadne gave Theseus the thread as a gift. I am who spins it.

Clotho governs **creation and initialization**--bringing sprints into existence within sessions.

---

## Operations

### create_sprint

Creates a new sprint within the current session.

**Syntax**:
```
create_sprint name="Sprint Name" [depends_on=sprint-id]
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| name | Yes | Human-readable sprint name |
| depends_on | No | Sprint ID this sprint depends on |

**Validation**:
1. Session must exist and be ACTIVE
2. If depends_on specified, dependency sprint must exist and be completed
3. Sprint name must not duplicate existing sprint in session
4. Sprint ID generated as: `sprint-{sanitized-name}-{YYYYMMDD}`

**File Creation**:
```
.claude/sessions/{session-id}/sprints/{sprint-id}/SPRINT_CONTEXT.md
```

**Sprint Context Schema**:
```yaml
---
sprint_id: "{generated-id}"
name: "{name}"
status: pending
depends_on: "{depends_on | null}"
created_at: "{ISO-8601}"
started_at: null
completed_at: null
tasks: []
---
```

**CLI Command**: None (direct file creation)

**Success Response**:
```json
{
  "success": true,
  "operation": "create_sprint",
  "message": "Sprint '{name}' created",
  "reasoning": "Created new sprint with dependency on completed design sprint",
  "fate": "clotho",
  "state_before": { "sprint_count": 1 },
  "state_after": { "sprint_count": 2, "new_sprint_id": "sprint-impl-20260107" },
  "changes": { "sprints": "+sprint-impl-20260107" }
}
```

**Error Responses**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Session not ACTIVE | LIFECYCLE_VIOLATION | Session must be ACTIVE to create sprints |
| Dependency not found | DEPENDENCY_BLOCKED | Sprint '{depends_on}' not found |
| Dependency not complete | DEPENDENCY_BLOCKED | Sprint '{depends_on}' must be completed first |
| Duplicate name | VALIDATION_FAILED | Sprint with name '{name}' already exists |

**Example**:
```
Input: create_sprint name="API Implementation" depends_on=sprint-design-20260106

Validation:
1. Read SESSION_CONTEXT.md -> session_state: ACTIVE (pass)
2. Read sprint-design-20260106/SPRINT_CONTEXT.md -> status: completed (pass)
3. Glob for existing sprint names -> no duplicate (pass)

Create:
Write .claude/sessions/{session-id}/sprints/sprint-api-impl-20260107/SPRINT_CONTEXT.md

Log:
2026-01-07T10:00:00Z | session-abc | create_sprint | moirai | name="API Implementation" | SUCCESS | clotho
```

---

### start_sprint

Activates a pending sprint, setting started_at timestamp.

**Syntax**:
```
start_sprint sprint_id
```

**Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| sprint_id | Yes | Sprint identifier to start |

**Validation**:
1. Sprint must exist
2. Sprint must be in `pending` status
3. All dependencies must be completed

**State Transition**: `pending -> active`

**CLI Command**: None (direct file update)

**Success Response**:
```json
{
  "success": true,
  "operation": "start_sprint",
  "message": "Sprint '{sprint_id}' started",
  "reasoning": "Sprint activated, dependencies satisfied",
  "fate": "clotho",
  "state_before": { "status": "pending", "started_at": null },
  "state_after": { "status": "active", "started_at": "2026-01-07T10:00:00Z" },
  "changes": { "status": "pending -> active", "started_at": "null -> 2026-01-07T10:00:00Z" }
}
```

**Error Responses**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Sprint not found | FILE_NOT_FOUND | Sprint '{sprint_id}' not found |
| Not pending | LIFECYCLE_VIOLATION | Sprint must be pending to start |
| Dependency incomplete | DEPENDENCY_BLOCKED | Dependency '{depends_on}' not completed |

---

## Anti-Patterns

| Anti-Pattern | Correct Behavior |
|--------------|------------------|
| Create existing sprint | Return VALIDATION_FAILED, not overwrite |
| Modify existing state | Clotho creates; Lachesis measures |
| Delete or archive | Clotho spins; Atropos cuts |
| Create sessions | Sessions created by `ari session create` |

---

## Natural Language Mapping

| Input | Operation |
|-------|-----------|
| "create a new sprint called X" | create_sprint name="X" |
| "new sprint for implementation" | create_sprint name="Implementation" |
| "spin up a testing sprint" | create_sprint name="Testing" |
| "start the implementation sprint" | start_sprint sprint-impl-* |
| "begin sprint X" | start_sprint X |
| "activate the testing sprint" | start_sprint sprint-testing-* |

---

## Audit Trail Format

```
TIMESTAMP | SESSION_ID | OPERATION | moirai | DETAILS | STATUS | clotho | reasoning="..."
```

Example:
```
2026-01-07T10:00:00Z | session-abc | create_sprint | moirai | name="Implementation" | SUCCESS | clotho | reasoning="Creating implementation sprint after design approval"
```

# ADR-0005: State-Mate Centralized State Mutation Authority

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2025-12-31 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The roster session management system tracks session and sprint state in markdown frontmatter files (`SESSION_CONTEXT.md`, `SPRINT_CONTEXT.md`). These files are critical to workflow coordination:

- **Session state** determines valid operations (ACTIVE sessions can be parked, PARKED sessions can be resumed)
- **Sprint state** tracks task progress and dependencies
- **Phase transitions** coordinate handoffs between specialist agents

### The Problem: Uncontrolled State Mutation

Without centralized control, any agent or the main Claude thread could directly modify context files:

```bash
# Dangerous: Direct edit bypasses validation
Edit(SESSION_CONTEXT.md, "status: ACTIVE", "status: ARCHIVED")
```

This creates several failure modes:

| Failure Mode | Description | Impact |
|--------------|-------------|--------|
| **Schema violations** | Invalid field values written | Workflow breaks, hook failures |
| **Lifecycle violations** | Invalid state transitions (e.g., PARKED -> ARCHIVED directly) | Corrupted session state |
| **Lost audit trail** | No record of who changed what or why | Cannot debug or rollback |
| **Race conditions** | Concurrent writes corrupt file | Data loss |
| **Orphaned substates** | Phase transitions without proper cleanup | Inconsistent workflow state |

### ADR-0001 Foundation

ADR-0001 (Session State Machine Redesign) established the formal state machine:

```
                    +--------------+
                    |              |
                    v              |
   (new) ---> ACTIVE ---> PARKED --+---> ARCHIVED
                |                         ^
                +-------------------------+
                     (direct wrap)
```

ADR-0001 specified that `state-mate` would serve as the "policy engine" for mutations, but did not fully define the enforcement mechanism. This ADR completes that architecture.

### Forces

- **Data Integrity**: Context files must remain valid and consistent
- **Auditability**: All state changes must be traceable with reasoning
- **Concurrency Safety**: Multiple hooks and agents may operate simultaneously
- **Developer Experience**: Mutations should have a clear, discoverable interface
- **Fail-Fast**: Invalid operations should be rejected early with helpful messages

## Decision

We establish **state-mate** as the sole authority for all mutations to `*_CONTEXT.md` files, enforced by a PreToolUse hook that blocks direct writes.

### 1. state-mate Agent Authority

The `state-mate` agent (defined in `user-agents/state-mate.md`) is the exclusive interface for context file mutations:

```
Task(state-mate, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md

Session Context:
- Session ID: session-20251231-120000-abcd1234
- Session Path: .claude/sessions/session-20251231-120000-abcd1234/SESSION_CONTEXT.md")
```

state-mate provides:

| Capability | Description |
|------------|-------------|
| **Schema Validation** | All mutations validated against JSON schemas before writing |
| **Lifecycle Enforcement** | State transitions checked against FSM transition matrix |
| **Audit Logging** | Every mutation logged with timestamp, operation, and reasoning |
| **Concurrency Control** | File locking prevents race conditions |
| **Structured Responses** | JSON responses enable programmatic handling |

### 2. PreToolUse Hook Enforcement

The `session-write-guard.sh` hook intercepts all `Write` and `Edit` operations targeting `*_CONTEXT.md` files:

```bash
# Pattern match for context files
if [[ "$FILE_PATH" =~ _CONTEXT\.md$ ]]; then
    # Block operation with helpful error
    cat <<'EOF'
{
  "decision": "block",
  "reason": "Direct writes to *_CONTEXT.md files are blocked"
}
EOF
    exit 1
fi
```

The hook provides context-aware error messages:

- **Active workflow**: Directs user to appropriate slash commands (`/park`, `/wrap`, `/handoff`)
- **No workflow**: Directs user to invoke state-mate via Task tool

### 3. Operations Interface

state-mate accepts both natural language and structured commands:

**Low-Level CRUD Operations:**

| Operation | Syntax | Description |
|-----------|--------|-------------|
| `create_sprint` | `create_sprint name="Sprint Name"` | Create new sprint |
| `update_field` | `update_field field_name=value` | Modify frontmatter field |
| `append_content` | `append_content "content"` | Add to markdown body |

**High-Level Semantic Operations:**

| Operation | Syntax | Description |
|-----------|--------|-------------|
| `mark_complete` | `mark_complete task_id artifact=path` | Mark task complete with artifact |
| `transition_phase` | `transition_phase from=X to=Y` | Move between workflow phases |
| `park_session` | `park_session reason="..."` | Park session with reason |
| `resume_session` | `resume_session` | Resume parked session |
| `wrap_session` | `wrap_session` | Complete and archive session |

### 4. Control Flags

For exceptional circumstances, state-mate provides control flags:

| Flag | Effect | Use Case |
|------|--------|----------|
| `--dry-run` | Preview mutation without applying | Verification before commit |
| `--emergency` | Bypass non-critical validations (logged) | Urgent fixes when validation blocks |
| `--override=reason` | Bypass lifecycle rules | Recovery scenarios requiring explicit reason |

Example:
```
Task(state-mate, "--override=reason='Data recovery from corrupted state' transition_phase from=requirements to=implementation")
```

### 5. Audit Trail

All mutations are logged to `.claude/sessions/.audit/session-mutations.log`:

```
TIMESTAMP | SESSION_ID | OPERATION | SOURCE | DETAILS | STATUS | REASONING
```

Example entries:
```
2025-12-31T23:45:00Z | session-abc123 | mark_complete | state-mate | task=task-001, artifact=docs/PRD.md | SUCCESS | "PRD completed per orchestrator directive"
2025-12-31T23:46:00Z | session-abc123 | transition_phase | state-mate | from=PARKED, to=ARCHIVED | FAILED:LIFECYCLE_VIOLATION | "Direct transition requires --override"
```

### 6. Response Format

state-mate returns structured JSON for all operations:

**Success:**
```json
{
  "success": true,
  "operation": "mark_complete",
  "message": "Task task-001 marked complete",
  "reasoning": "Task marked complete per request. Artifact validated at path.",
  "state_before": { "task_status": "pending" },
  "state_after": { "task_status": "completed", "completed_at": "2025-12-31T23:45:00Z" }
}
```

**Failure:**
```json
{
  "success": false,
  "operation": "transition_phase",
  "error_code": "LIFECYCLE_VIOLATION",
  "message": "Cannot transition from PARKED to ARCHIVED without --override",
  "hint": "Use: resume_session first, OR --override=reason=\"...\""
}
```

## Consequences

### Positive

1. **Schema Integrity**: All mutations validated before write, preventing invalid state
2. **Complete Audit Trail**: Every change logged with reasoning, enabling debugging and rollback analysis
3. **Lifecycle Enforcement**: Invalid state transitions rejected at boundary, not discovered later
4. **Concurrency Safety**: File locking prevents race conditions from concurrent operations
5. **Consistent Interface**: Single point of entry for mutations simplifies documentation and training
6. **Fail-Fast Behavior**: Invalid operations rejected immediately with actionable error messages
7. **Recovery Options**: Control flags enable override for legitimate exceptional cases

### Negative

1. **No Quick Direct Edits**: Cannot manually fix a typo in context files without going through state-mate
2. **Learning Curve**: Developers must learn state-mate invocation pattern
3. **Hook Dependency**: System requires hook to be properly installed and configured
4. **Additional Latency**: Mutations route through Task tool invocation rather than direct file write

### Neutral

1. **Debugging Context Files**: Can still Read context files directly for inspection
2. **Schema Evolution**: Schema changes require coordinated update to state-mate validation
3. **Hook Bypass**: Malicious or misconfigured environments could disable hook (defense in depth, not absolute)

## Implementation

### Components

| Component | Path | Purpose |
|-----------|------|---------|
| state-mate agent | `user-agents/state-mate.md` | Agent definition with operations |
| Write guard hook | `.claude/hooks/session-guards/session-write-guard.sh` | PreToolUse interceptor |
| Session schemas | `schemas/artifacts/session-context.schema.json` | Validation schema |
| Sprint schemas | `schemas/artifacts/sprint-context.schema.json` | Validation schema |
| Audit log | `.claude/sessions/.audit/session-mutations.log` | Mutation history |

### Hook Registration

The write guard is registered in `.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": ["bash .claude/hooks/session-guards/session-write-guard.sh"]
      }
    ]
  }
}
```

### Integration with Session Manager

state-mate complements (does not replace) `session-manager.sh`:

| Concern | session-manager.sh | state-mate |
|---------|-------------------|------------|
| Session creation | `create` command | Read-only (must exist) |
| Session query | `status` command | Read-only |
| TTY/terminal mapping | Handles | Not aware |
| Agent-to-agent mutations | Not used | Primary path |
| Schema validation | Basic (optional) | Full enforcement |
| Lifecycle enforcement | Basic transitions | Full state machine |

## Related Decisions

- **ADR-0001**: Session State Machine Redesign (defines FSM that state-mate enforces)
- **ADR-0002**: Hook Library Resolution Architecture (defines hook installation pattern)

## References

- state-mate agent: `user-agents/state-mate.md`
- Session write guard hook: `.claude/hooks/session-guards/session-write-guard.sh`
- Session schemas: `schemas/artifacts/session-context.schema.json`
- CLAUDE.md state management section: `.claude/CLAUDE.md` (lines 81-120)
- Invocation patterns: `user-skills/session-lifecycle/shared-sections/state-mate-invocation.md`

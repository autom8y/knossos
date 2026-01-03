# state-mate Skill Reference

> Quick reference for state-mate operations, error codes, and common patterns.

## When to Activate

Use state-mate for **any mutation** to session or sprint context files:
- Updating `SESSION_CONTEXT.md` or `SPRINT_CONTEXT.md`
- Marking tasks complete
- Transitioning workflow phases
- Parking, resuming, or wrapping sessions
- Creating or managing sprints
- Recording decisions or handoffs

**Never use direct Write/Edit** on `*_CONTEXT.md` files—PreToolUse hook will block with instruction to use state-mate.

---

## Quick Start

### Invocation via Task Tool

```
Task(state-mate, "operation [params] [flags]")
```

**Natural language supported**:
```
Task(state-mate, "Mark the PRD task complete with artifact at docs/requirements/PRD-foo.md")
```

### Common Operations

| Operation | Syntax | Example |
|-----------|--------|---------|
| **Mark complete** | `mark_complete task_id artifact=path` | `mark_complete task-001 artifact=docs/requirements/PRD-foo.md` |
| **Park session** | `park_session reason="..."` | `park_session reason="urgent bug"` |
| **Resume session** | `resume_session` | `resume_session` |
| **Wrap session** | `wrap_session` | `wrap_session` |
| **Transition phase** | `transition_phase from=X to=Y` | `transition_phase from=design to=implementation` |
| **Create sprint** | `create_sprint name="..." [depends_on=sprint-id]` | `create_sprint name="API Implementation" depends_on=sprint-schema-001` |
| **Update field** | `update_field field=value` | `update_field status=completed` |
| **Record decision** | `record_decision "text"` | `record_decision "Chose PostgreSQL for ACID guarantees"` |
| **Handoff** | `handoff to=agent note="..."` | `handoff to=integration-engineer note="PRD approved"` |

---

## Control Flags

| Flag | Purpose | Usage |
|------|---------|-------|
| `--dry-run` | Preview changes without applying | `--dry-run mark_complete task-001` |
| `--emergency` | Bypass non-critical validations (logged) | `--emergency update_field session_state=ACTIVE` |
| `--override=reason` | Bypass lifecycle rules with explicit reason | `--override="data recovery" transition_phase from=parked to=archived` |

---

## Response Format

All responses are structured JSON:

### Success Response

```json
{
  "success": true,
  "operation": "mark_complete",
  "message": "Task task-001 marked complete",
  "reasoning": "Task marked complete per request. Artifact validated at path.",
  "state_before": { "task_status": "pending" },
  "state_after": { "task_status": "completed", "completed_at": "2025-12-29T23:45:00Z" },
  "changes": {
    "status": "pending -> completed",
    "completed_at": "null -> 2025-12-29T23:45:00Z"
  }
}
```

### Error Response

```json
{
  "success": false,
  "operation": "transition_phase",
  "error_code": "LIFECYCLE_VIOLATION",
  "message": "Cannot transition from PARKED to ARCHIVED without --override",
  "reasoning": "Direct transition from PARKED requires explicit confirmation",
  "hint": "Use: resume_session first, OR --override=reason=\"...\"",
  "allowed_transitions": ["ACTIVE"]
}
```

### Dry-Run Response

```json
{
  "success": true,
  "operation": "park_session",
  "dry_run": true,
  "message": "Preview: Would park session",
  "diff": {
    "session_state": "ACTIVE -> PARKED",
    "parked_at": "null -> 2025-12-29T23:45:00Z",
    "park_reason": "null -> Taking a break"
  }
}
```

---

## Error Codes

| Code | Description | Common Causes | Typical Fix |
|------|-------------|---------------|-------------|
| `SCHEMA_VIOLATION` | Output would not pass schema validation | Invalid field value, missing required field | Check schema: `roster/schemas/artifacts/*.schema.json` |
| `LIFECYCLE_VIOLATION` | State transition not allowed | Invalid session_state transition, incomplete prerequisites | Check allowed transitions in error response |
| `DEPENDENCY_BLOCKED` | Operation blocked by dependency | Sprint depends on incomplete sprint | Complete dependency first |
| `LOCK_TIMEOUT` | Could not acquire file lock | Concurrent operation in progress | Retry in a few seconds |
| `FILE_NOT_FOUND` | Target context file does not exist | Session not created, invalid path | Create session first with `/init` |
| `PERMISSION_DENIED` | Cannot write to target file | File system permissions | Check file permissions with `ls -la` |
| `INVALID_OPERATION` | Unrecognized operation | Typo in command, unsupported operation | See valid operations in this reference |
| `VALIDATION_FAILED` | Pre-mutation validation failed | Missing required artifact, invalid field value | Check error message for specific requirement |
| `CONCURRENT_MODIFICATION` | File changed during operation | Race condition (rare with locking) | Re-read current state and retry |

---

## State Machine Reference

### Session State Transitions

```
                    +--------------+
                    |              |
                    v              |
   (new) ---> ACTIVE ---> PARKED --+---> ARCHIVED
                |                         ^
                +-------------------------+
                     (direct wrap)
```

**Allowed Transitions**:
- `(new) -> ACTIVE`: Session creation
- `ACTIVE -> PARKED`: Park session
- `PARKED -> ACTIVE`: Resume session
- `ACTIVE -> ARCHIVED`: Wrap session
- `PARKED -> ARCHIVED`: Wrap parked session (requires `--override`)

### Sprint Status Transitions

```
   pending ---> active ---> blocked ---> completed ---> archived
                  |            |              ^
                  |            +--------------+
                  |                (unblock)
                  +---------------------------+
                       (can complete directly)
```

**Allowed Transitions**:
- `pending -> active`: Start sprint
- `active -> blocked`: Set blocker
- `blocked -> active`: Clear blocker
- `active -> completed`: Mark complete (all tasks done)
- `blocked -> completed`: Mark complete (blocker resolved + all tasks done)
- `completed -> archived`: Archive sprint

---

## Common Patterns

### Pattern 1: Complete Task with Artifact

```
Task(state-mate, "mark_complete task-prd artifact=docs/requirements/PRD-state-mate.md")
```

**When**: Task produces a deliverable artifact (PRD, TDD, code file).

**Validation**: state-mate verifies artifact path exists.

### Pattern 2: Preview Mutation Before Applying

```
# Preview
Task(state-mate, "--dry-run park_session reason='urgent bug'")

# Review diff, then apply
Task(state-mate, "park_session reason='urgent bug'")
```

**When**: Want to verify changes before committing.

**Returns**: Diff showing old -> new values.

### Pattern 3: Phase Transition with Prerequisites

```
# Will fail if TDD not present
Task(state-mate, "transition_phase from=design to=implementation")

# Error response includes hint
{
  "success": false,
  "error_code": "VALIDATION_FAILED",
  "message": "TDD required before implementation phase",
  "hint": "Create TDD in docs/design/ first"
}
```

**When**: Workflow has artifact prerequisites for phase transitions.

**Extensions**: Loaded from `ACTIVE_WORKFLOW.yaml` → `state_mate_extensions`.

### Pattern 4: Emergency Override for Recovery

```
# Session stuck in invalid state
Task(state-mate, "--emergency update_field session_state=ACTIVE")

# Logged to audit trail with reasoning
```

**When**: Data recovery, fixing corruption, bypassing validation temporarily.

**Caution**: Logged to audit trail. Use sparingly.

### Pattern 5: Parallel Sprints with Dependencies

```
# Create base sprint
Task(state-mate, "create_sprint name='Schema Updates'")

# Create dependent sprint
Task(state-mate, "create_sprint name='API Implementation' depends_on=sprint-schema-20251229")

# Cannot complete dependent until base completes
Task(state-mate, "mark_complete sprint-api-20251229")  # FAILS: DEPENDENCY_BLOCKED
```

**When**: Running parallel work streams with ordering constraints.

**Validation**: state-mate enforces `depends_on` before allowing completion.

### Pattern 6: Record Decision for Audit Trail

```
Task(state-mate, "record_decision 'Chose PostgreSQL over MongoDB for ACID guarantees and relational data model'")
```

**When**: Important architectural decision made during session.

**Result**: Appended to SESSION_CONTEXT.md markdown body with timestamp.

---

## File Paths

### Session-Level Context

```
.claude/sessions/{session-id}/SESSION_CONTEXT.md
```

### Default Sprint

```
.claude/sessions/{session-id}/SPRINT_CONTEXT.md
```

### Named Sprint (Parallel Sprints)

```
.claude/sessions/{session-id}/sprints/{sprint-id}/SPRINT_CONTEXT.md
```

### Audit Log

```
.claude/sessions/.audit/session-mutations.log
```

**Format**: `TIMESTAMP | SESSION_ID | OPERATION | SOURCE | DETAILS | STATUS`

Example:
```
2025-12-29T23:45:00Z | session-abc123 | mark_complete | state-mate | task=task-001, artifact=docs/PRD.md | SUCCESS
```

---

## Integration with session-manager.sh

| Concern | session-manager.sh | state-mate |
|---------|-------------------|------------|
| Session creation | `create` command | Read-only (session must exist) |
| Session query | `status` command | Read-only |
| Agent-to-agent mutations | Not used | Primary path |
| CLI access | Shell entry point | Task tool only |
| Schema validation | Basic (optional) | Full enforcement (required) |

**Workflow**:
1. Create session: `session-manager.sh create` (CLI)
2. Mutate state: `Task(state-mate, "...")` (agents)
3. Query status: `session-manager.sh status` (CLI)

---

## Troubleshooting

### Issue: "Direct writes to *_CONTEXT.md files are not allowed"

**Cause**: Attempted `Write` or `Edit` to `SESSION_CONTEXT.md` or `SPRINT_CONTEXT.md`.

**Solution**: Use state-mate instead:
```
Task(state-mate, "update_field field_name=value")
```

### Issue: "LOCK_TIMEOUT: Could not acquire lock after 10s"

**Cause**: Another operation is mutating the same session.

**Solution**: Wait and retry. If persists, check for stale lock:
```bash
ls -la .claude/sessions/{session-id}/.locks/context.lock
# If older than 60s, state-mate will auto-force-release
```

### Issue: "LIFECYCLE_VIOLATION: Cannot transition from PARKED to ARCHIVED"

**Cause**: Invalid state transition per state machine.

**Solution**: Check `allowed_transitions` in error response. Example:
```
# Resume first
Task(state-mate, "resume_session")
# Then wrap
Task(state-mate, "wrap_session")

# OR use override
Task(state-mate, "--override='session abandoned' wrap_session")
```

### Issue: "DEPENDENCY_BLOCKED: Sprint depends on incomplete sprint"

**Cause**: Attempted to complete sprint with incomplete `depends_on` dependency.

**Solution**: Complete dependency first:
```
# Check blocker status
Read(.claude/sessions/{session-id}/sprints/{blocker-id}/SPRINT_CONTEXT.md)

# Complete blocker
Task(state-mate, "mark_complete {blocker-id}")

# Then complete dependent
Task(state-mate, "mark_complete {dependent-id}")
```

---

## Extension Points

state-mate loads workflow-specific extensions from `ACTIVE_WORKFLOW.yaml`:

```yaml
state_mate_extensions:
  - path: .claude/extensions/state-mate/forge-pack.md
    triggers:
      - transition_phase
      - mark_complete
```

Extensions can add:
- **Pre-mutation hooks**: Validate before operation (e.g., require artifact for mark_complete)
- **Post-mutation hooks**: Trigger actions after success (e.g., auto-archive on sprint completion)
- **Custom operations**: Compound operations (e.g., approve_prd = mark_complete + transition_phase)
- **Custom validators**: Workflow-specific rules (e.g., TDD required before implementation)

---

## References

- **Agent Prompt**: `.claude/agents/state-mate.md`
- **ADR**: `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md`
- **PRD**: `docs/requirements/PRD-state-mate.md`
- **TDD**: `docs/design/TDD-state-mate.md`
- **Hook**: `.claude/hooks/session-write-guard.sh`
- **Session Schema**: `roster/schemas/artifacts/session-context.schema.json`
- **Sprint Schema**: `roster/schemas/artifacts/sprint-context.schema.json`

---

## The Acid Test

Before invoking state-mate, verify:

1. **Am I mutating context state?** (If yes, use state-mate)
2. **Have I checked current state?** (Read `*_CONTEXT.md` first to understand current state)
3. **Do I know which operation?** (See Common Operations table)
4. **Do I need dry-run?** (Use `--dry-run` to preview if uncertain)

If uncertain, default to state-mate. Hook will guide you if you use direct Write/Edit by mistake.

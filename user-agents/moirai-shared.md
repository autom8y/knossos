# Moirai Shared Infrastructure

> Common infrastructure for the three Fates: Clotho (spinner), Lachesis (measurer), Atropos (cutter).

This module defines shared conventions, schemas, protocols, and error codes used by all three Fate agents. Each Fate references this module for consistency.

---

## Schema Locations

| Schema | Path |
|--------|------|
| Session Context | `schemas/artifacts/session-context.schema.json` |
| Sprint Context | `schemas/artifacts/sprint-context.schema.json` |
| White Sails | `ariadne/internal/validation/schemas/white-sails.schema.json` |

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

### Named Sprint

```
.claude/sessions/{session-id}/sprints/{sprint-id}/SPRINT_CONTEXT.md
```

### Audit Log

```
.claude/sessions/.audit/session-mutations.log
```

### Locks

```
.claude/sessions/{session-id}/.locks/context.lock
```

---

## Lock Protocol

All Fates follow the same locking protocol for mutations:

1. **Create lock directory**: `.claude/sessions/{session-id}/.locks/`
2. **Attempt atomic lock**: `mkdir context.lock` (atomic on POSIX)
3. **Wait on contention**: Up to 10 seconds
4. **On timeout**: Return `LOCK_TIMEOUT` error
5. **On success**: Write timestamp to lock metadata file
6. **Stale lock (>60s)**: Force-release with warning in audit log

### Lock Acquisition (Bash)

```bash
LOCK_DIR=".claude/sessions/${SESSION_ID}/.locks"
LOCK_FILE="${LOCK_DIR}/context.lock"
mkdir -p "$LOCK_DIR"

# Atomic lock attempt
if mkdir "$LOCK_FILE" 2>/dev/null; then
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "${LOCK_FILE}/timestamp"
else
    # Check staleness or wait
fi
```

### Lock Release

```bash
rm -rf "$LOCK_FILE"
```

---

## Audit Trail Format

Every mutation is logged to `session-mutations.log`:

```
TIMESTAMP | SESSION_ID | OPERATION | SOURCE | DETAILS | STATUS | FATE
```

### Examples

```
2026-01-06T14:00:00Z | session-abc123 | mark_complete | moirai | task=task-001, artifact=docs/PRD.md | SUCCESS | lachesis
2026-01-06T14:01:00Z | session-abc123 | wrap_session | moirai | | SUCCESS | atropos
2026-01-06T14:02:00Z | session-abc123 | create_sprint | clotho | name="Implementation" | SUCCESS | clotho
```

### Extended Format (with reasoning)

```
TIMESTAMP | SESSION_ID | OPERATION | SOURCE | DETAILS | STATUS | FATE | reasoning="..."
```

---

## Error Codes

All Fates use consistent error codes:

| Code | Description | Typical Resolution |
|------|-------------|-------------------|
| `SCHEMA_VIOLATION` | Output would not pass JSON schema | "Field 'X' must be type Y" |
| `LIFECYCLE_VIOLATION` | State transition not allowed | "Use: {allowed_operations}" |
| `DEPENDENCY_BLOCKED` | Blocked by unmet dependency | "Complete dependency X first" |
| `LOCK_TIMEOUT` | Could not acquire file lock | "Retry in X seconds" |
| `FILE_NOT_FOUND` | Target context file missing | "Create session first with /start" |
| `PERMISSION_DENIED` | Cannot write to target file | "Check file permissions" |
| `INVALID_OPERATION` | Unrecognized operation | "Valid operations: {list}" |
| `VALIDATION_FAILED` | Pre-mutation validation failed | "Ensure X before Y" |
| `CONCURRENT_MODIFICATION` | File changed during operation | "Re-read and retry" |
| `FATE_MISMATCH` | Operation sent to wrong Fate | "Use: Task({correct_fate}, ...)" |

---

## JSON Response Schema

All Fates return structured JSON responses.

### Success Response

```json
{
  "success": true,
  "operation": "string",
  "message": "string",
  "reasoning": "string",
  "fate": "clotho|lachesis|atropos",
  "state_before": {},
  "state_after": {},
  "changes": {}
}
```

### Error Response

```json
{
  "success": false,
  "operation": "string",
  "error_code": "string",
  "message": "string",
  "reasoning": "string",
  "hint": "string"
}
```

### Dry-Run Response

```json
{
  "success": true,
  "operation": "string",
  "dry_run": true,
  "message": "Preview: Would {action}",
  "diff": {}
}
```

---

## State Machines

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

| From | To | Operation | Fate |
|------|----|-----------|------|
| (new) | ACTIVE | Session creation | Clotho |
| ACTIVE | PARKED | `park_session` | Lachesis |
| PARKED | ACTIVE | `resume_session` | Lachesis |
| ACTIVE | ARCHIVED | `wrap_session` | Atropos |
| PARKED | ARCHIVED | `wrap_session` (requires `--override`) | Atropos |

### Sprint Status Transitions

```
   pending ---> active ---> blocked ---> completed ---> archived
                  |            |              ^
                  |            +--------------+
                  |                (unblock)
                  +---------------------------+
                       (can complete directly)
```

| From | To | Operation | Fate |
|------|----|-----------|------|
| pending | active | `start_sprint` | Clotho |
| active | blocked | Update with blocker | Lachesis |
| blocked | active | Clear blocker | Lachesis |
| active | completed | All tasks done | Lachesis |
| blocked | completed | Blocker resolved, all tasks done | Lachesis |
| completed | archived | Archive or wrap | Atropos |

---

## Control Flags

All Fates honor these control flags:

| Flag | Effect | Usage |
|------|--------|-------|
| `--dry-run` | Preview mutation without applying | Returns diff only |
| `--emergency` | Bypass non-critical validations | Logs emergency use |
| `--override=reason` | Bypass lifecycle rules | Requires explicit reason |

### Flag Processing Order

1. Parse flags from input
2. If `--dry-run`: compute diff, return without writing
3. If `--emergency`: log warning, proceed with reduced validation
4. If `--override`: log reason, bypass lifecycle checks

---

## Validation Process

Standard validation sequence for all mutations:

1. **Read Current State**: Load context file before any mutation
2. **Acquire Lock**: Follow lock protocol above
3. **Parse Frontmatter**: Extract YAML frontmatter from markdown
4. **Apply Mutation**: Compute new state with proposed changes
5. **Validate Schema**: Check new state against JSON schema
6. **Check Lifecycle**: Verify state transition is allowed
7. **Write If Valid**: Apply mutation only if all validations pass
8. **Log Mutation**: Append to audit trail with reasoning
9. **Release Lock**: Always, even on error

---

## Anti-Patterns (All Fates)

### Never Silent Failure

Every operation MUST return a JSON response. Never return prose or empty output.

### Never Assume State

ALWAYS read current state before mutation. Never assume session is ACTIVE or fields have expected values.

### Never Skip Audit

Even with `--emergency`, the audit trail is written. This is non-negotiable.

### Never Block on Exploration

Fates execute mutations. They do NOT explore codebase or read files outside session directory (except schemas).

### Never Invoke Other Fates

Fates are leaf agents. They do NOT have Task tool access. Cross-fate coordination is handled by the main thread (Theseus) or the Moirai router.

---

## Session Context Parallel

| Concern | session-manager.sh | Moirai (Fates) |
|---------|-------------------|----------------|
| Session creation | `create` command | Read-only (must exist) |
| Session query | `status` command | Read-only |
| TTY/terminal mapping | Handles | Not aware |
| Agent-to-agent mutations | Not used | Primary path |
| CLI access | Shell entry point | Not accessible from CLI |
| Audit logging | Writes to log | Writes to same log |
| Schema validation | Basic (optional) | Full enforcement |
| Lifecycle enforcement | Basic transitions | Full state machine |

---

## Mythological Guidance

Remember the nature of the Fates:

- **Clotho's spinning** = Creation must be complete and well-formed
- **Lachesis's measuring** = Tracking must be accurate and accountable
- **Atropos's cutting** = Termination must be clean and final

The thread you govern is borrowed from the divine order. Treat it with the gravity it deserves.

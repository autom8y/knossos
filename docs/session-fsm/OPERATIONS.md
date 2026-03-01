# Session State Machine Operations Guide

This guide covers common operations, migration procedures, troubleshooting, and recovery for the session state machine.

## Quick Reference

### Session Commands

| Command                              | Description                     |
|--------------------------------------|---------------------------------|
| `session-manager.sh status`          | Show current session state      |
| `session-manager.sh create <init> <cx>` | Create new session           |
| `session-manager.sh mutate park`     | Park current session            |
| `session-manager.sh mutate resume`   | Resume parked session           |
| `session-manager.sh mutate wrap`     | Archive and complete session    |

### FSM Direct Commands

| Command                                      | Description                     |
|----------------------------------------------|---------------------------------|
| `session-fsm.sh get-state <session_id>`      | Get session state               |
| `session-fsm.sh transition <id> <state>`     | Execute state transition        |
| `session-fsm.sh create <init> <cx> [team]`   | Create session via FSM          |
| `session-fsm.sh is-valid-transition <from> <to>` | Check if transition valid   |
| `session-fsm.sh validate <ctx_file>`         | Validate SESSION_CONTEXT.md     |

### Migration Commands

| Command                                      | Description                     |
|----------------------------------------------|---------------------------------|
| `session-migrate.sh status`                  | Show all sessions' schema status|
| `session-migrate.sh status <session_id>`     | Show specific session status    |
| `session-migrate.sh migrate --dry-run --batch` | Preview migration            |
| `session-migrate.sh migrate --batch`         | Migrate all v1 sessions         |
| `session-migrate.sh migrate <session_id>`    | Migrate specific session        |
| `session-migrate.sh rollback <session_id>`   | Rollback to v1 backup           |

## Common Scenarios

### Creating a New Session

```bash
# Via session-manager (recommended)
./user-hooks/lib/session-manager.sh create "Feature: Dark Mode" "MODULE" "10x-dev"

# Output:
# {"success": true, "session_id": "session-20251231-120000-abc12345"}
```

The session starts in `ACTIVE` state with `current_phase: requirements`.

### Parking a Session

When you need to pause work (lunch, meetings, switching context):

```bash
# Park with reason
./user-hooks/lib/session-manager.sh mutate park "Taking lunch break"

# Output:
# {"success": true, "state": "PARKED"}
```

This:
1. Acquires exclusive lock
2. Validates ACTIVE -> PARKED is allowed
3. Updates status field to PARKED
4. Emits SESSION_PARKED event with reason
5. Releases lock

### Resuming a Parked Session

```bash
./user-hooks/lib/session-manager.sh mutate resume

# Output:
# {"success": true, "state": "ACTIVE"}
```

The session returns to ACTIVE with its previous `current_phase` preserved.

### Archiving/Wrapping a Session

When work is complete:

```bash
./user-hooks/lib/session-manager.sh mutate wrap "true"

# Output:
# {"success": true, "state": "ARCHIVED"}
```

This:
1. Moves session to `.sos/archive/`
2. Sets status to ARCHIVED (terminal state)
3. No further transitions are possible

### Checking Session State

```bash
# Full status JSON
./user-hooks/lib/session-manager.sh status

# Output:
# {
#   "has_session": true,
#   "session_id": "session-20251231-120000-abc12345",
#   "session_state": "ACTIVE",
#   "current_phase": "design",
#   ...
# }

# Just the state
./user-hooks/lib/session-fsm.sh get-state session-20251231-120000-abc12345
# Output: ACTIVE
```

## Migration Procedures

### Checking Migration Status

Before migrating, check current state:

```bash
./user-hooks/lib/session-migrate.sh status

# Output:
# {
#   "total_sessions": 10,
#   "v1_sessions": 3,
#   "v2_sessions": 7,
#   "with_backup": 2,
#   "migration_needed": 3
# }
```

### Dry Run Migration

Always preview before migrating:

```bash
./user-hooks/lib/session-migrate.sh migrate --dry-run --batch

# Output:
# Would migrate session-20251231-100000-aaa: status=ACTIVE
# Would migrate session-20251231-110000-bbb: status=PARKED
# Would migrate session-20251231-120000-ccc: status=ARCHIVED
```

### Migrating All Sessions

```bash
./user-hooks/lib/session-migrate.sh migrate --batch

# Output:
# [2025-12-31T12:00:00Z] [INFO] Starting batch migration...
# [2025-12-31T12:00:00Z] [INFO] Migrating: session-20251231-100000-aaa
# [2025-12-31T12:00:01Z] [INFO] Migrated: session-20251231-100000-aaa (status=ACTIVE)
# ...
# [2025-12-31T12:00:05Z] [INFO] Batch migration complete: 3 succeeded, 0 failed, 7 skipped (already v2)
```

### Migrating a Single Session

```bash
./user-hooks/lib/session-migrate.sh migrate session-20251231-100000-aaa

# Output:
# [2025-12-31T12:00:00Z] [INFO] Migrating: session-20251231-100000-aaa
# [2025-12-31T12:00:01Z] [INFO] Migrated: session-20251231-100000-aaa (status=ACTIVE)
```

### Rolling Back a Migration

If migration causes issues:

```bash
# Rollback specific session
./user-hooks/lib/session-migrate.sh rollback session-20251231-100000-aaa

# Output:
# [2025-12-31T12:00:00Z] [INFO] Rolled back: session-20251231-100000-aaa

# Rollback all sessions with backups
./user-hooks/lib/session-migrate.sh rollback --batch
```

Rollback restores from `.v1.backup` files created during migration.

### Auto-Migration on Access

Sessions are automatically migrated on first access via `session-manager.sh status`:

```bash
# When you run any session command, v1 sessions are auto-migrated
./user-hooks/lib/session-manager.sh status

# Behind the scenes:
# [INFO] Auto-migrating v1 session: session-20251231-100000-aaa
```

## Troubleshooting Guide

### Error: LOCK_TIMEOUT

**Symptom**: Operation fails with `{"error": "LOCK_TIMEOUT"}`

**Cause**: Another process holds the lock, or a stale lock exists.

**Resolution**:

```bash
# 1. Check for running Claude Code instances
ps aux | grep "claude"

# 2. Check lock status
ls -la .sos/sessions/.locks/

# 3. Remove stale lock (if PID is dead)
# First, check the PID
cat .sos/sessions/.locks/<session-id>.lock.d/pid

# If that process is not running:
rm -rf .sos/sessions/.locks/<session-id>.lock.d/

# 4. Increase timeout (temporary)
FSM_LOCK_TIMEOUT=30 ./user-hooks/lib/session-manager.sh mutate resume
```

### Error: INVALID_TRANSITION

**Symptom**: Operation fails with `{"error": "INVALID_TRANSITION", "from": "...", "to": "..."}`

**Cause**: Attempted transition is not allowed by FSM rules.

**Common Cases**:

| Error | Cause | Solution |
|-------|-------|----------|
| PARKED -> PARKED | Trying to park already parked session | Resume first |
| ACTIVE -> ACTIVE | Trying to resume non-parked session | Session is already active |
| ARCHIVED -> * | Trying to modify archived session | Archived is terminal; cannot be changed |

**Resolution**:

```bash
# Check current state
./user-hooks/lib/session-fsm.sh get-state <session_id>

# Check valid transitions
./user-hooks/lib/session-fsm.sh is-valid-transition PARKED ACTIVE
# Output: true

./user-hooks/lib/session-fsm.sh is-valid-transition ARCHIVED ACTIVE
# Output: false
```

### Error: VALIDATION_FAILED

**Symptom**: Operation fails with `{"error": "VALIDATION_FAILED"}`

**Cause**: SESSION_CONTEXT.md does not meet v2 schema requirements.

**Resolution**:

```bash
# 1. Validate the file directly
./user-hooks/lib/session-fsm.sh validate .sos/sessions/<session_id>/SESSION_CONTEXT.md

# 2. Check for missing required fields
grep -E "^(schema_version|session_id|status|created_at|initiative|complexity|active_rite|current_phase):" \
  .sos/sessions/<session_id>/SESSION_CONTEXT.md

# 3. Fix missing fields manually or re-migrate
./user-hooks/lib/session-migrate.sh migrate <session_id>
```

**Required v2 Fields**:
- `schema_version: "2.0"`
- `session_id`
- `status` (ACTIVE, PARKED, or ARCHIVED)
- `created_at`
- `initiative`
- `complexity`
- `active_rite`
- `current_phase`

### Error: Session Not Found

**Symptom**: `{"error": "Session not found: <session_id>"}`

**Cause**: Session directory or context file does not exist.

**Resolution**:

```bash
# 1. List existing sessions
ls -la .sos/sessions/

# 2. Check if session was archived
ls -la .sos/archive/

# 3. Check current session link
cat .sos/sessions/.current 2>/dev/null || echo "No current session"
```

### Error: No Active Session

**Symptom**: `{"error": "No active session"}`

**Cause**: No session is currently selected.

**Resolution**:

```bash
# 1. List available sessions
ls -d .sos/sessions/session-* 2>/dev/null

# 2. Create a new session
./user-hooks/lib/session-manager.sh create "My Initiative" "MODULE"

# 3. Or set an existing session as current
echo "session-20251231-120000-abc" > .sos/sessions/.current
```

### Migration Fails: Missing Required Fields

**Symptom**: Migration fails with "Missing required field: X"

**Cause**: v1 session lacks fields required by v2 schema.

**Resolution**:

```bash
# 1. Check what fields exist
grep "^[a-z_]*:" .sos/sessions/<session_id>/SESSION_CONTEXT.md

# 2. Add missing fields manually
cat >> .sos/sessions/<session_id>/SESSION_CONTEXT.md << 'EOF'
initiative: "Unknown"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "requirements"
EOF

# 3. Re-run migration
./user-hooks/lib/session-migrate.sh migrate <session_id>
```

### Corrupted Session State

**Symptom**: Session behaves unexpectedly, state is inconsistent.

**Resolution**:

```bash
# 1. Check the raw context file
cat .sos/sessions/<session_id>/SESSION_CONTEXT.md

# 2. Check event history
cat .sos/sessions/<session_id>/events.jsonl

# 3. If backup exists, consider rollback
ls -la .sos/sessions/<session_id>/SESSION_CONTEXT.md.v1.backup

# 4. If no backup, manually fix the status field
# Edit the file and ensure status: "ACTIVE" (or appropriate state)
```

## Recovery Procedures

### Restore from Backup

Each migration creates a `.v1.backup` file:

```bash
# Check if backup exists
ls -la .sos/sessions/<session_id>/SESSION_CONTEXT.md.v1.backup

# Restore backup
mv .sos/sessions/<session_id>/SESSION_CONTEXT.md.v1.backup \
   .sos/sessions/<session_id>/SESSION_CONTEXT.md
```

### Manual State Repair

If state is corrupted and no backup exists:

```bash
# 1. Determine correct state from events log
tail -5 .sos/sessions/<session_id>/events.jsonl

# 2. Edit SESSION_CONTEXT.md directly
# Ensure the status field matches the last known good state

# 3. Validate the repair
./user-hooks/lib/session-fsm.sh validate .sos/sessions/<session_id>/SESSION_CONTEXT.md
```

### Clearing Stale Locks

If system crashed while holding lock:

```bash
# Remove all lock files for a session
rm -rf .sos/sessions/.locks/<session_id>.lock*

# Or remove all locks (nuclear option)
rm -rf .sos/sessions/.locks/
```

### Recreating Missing Audit Logs

If audit logs are lost:

```bash
# Recreate from session event logs
mkdir -p .sos/sessions/.audit

for events_file in .sos/sessions/*/events.jsonl; do
    session_id=$(basename $(dirname "$events_file"))
    while read -r line; do
        ts=$(echo "$line" | grep -o '"timestamp":"[^"]*"' | cut -d'"' -f4)
        event=$(echo "$line" | grep -o '"event":"[^"]*"' | cut -d'"' -f4)
        from=$(echo "$line" | grep -o '"from":"[^"]*"' | cut -d'"' -f4)
        to=$(echo "$line" | grep -o '"to":"[^"]*"' | cut -d'"' -f4)
        echo "$ts | $session_id | $event | $from -> $to"
    done < "$events_file"
done >> .sos/sessions/.audit/transitions.log
```

## Error Codes Reference

| Code | Name                 | Exit | Description                       |
|------|----------------------|------|-----------------------------------|
| 0    | Success              | 0    | Operation completed successfully  |
| 1    | General Error        | 1    | Generic failure                   |
| 1    | INVALID_TRANSITION   | 1    | Transition not allowed by FSM     |
| 1    | VALIDATION_FAILED    | 1    | Schema validation failed          |
| 1    | SESSION_NOT_FOUND    | 1    | Session directory not found       |
| 1    | WRITE_FAILED         | 1    | Failed to write context file      |
| 2    | LOCK_TIMEOUT         | 2    | Could not acquire lock in time    |

## Performance Considerations

### Lock Timeout Tuning

Default: 10 seconds. Adjust based on system load:

```bash
# Increase for slow systems
export FSM_LOCK_TIMEOUT=30

# Decrease for fast feedback during development
export FSM_LOCK_TIMEOUT=5
```

### Disabling Validation (Not Recommended)

For debugging only:

```bash
export FSM_VALIDATE_SCHEMA=false
./user-hooks/lib/session-manager.sh mutate park
```

### Disabling Events (Not Recommended)

If event logging causes issues:

```bash
export FSM_EMIT_EVENTS=false
./user-hooks/lib/session-manager.sh mutate park
```

## Monitoring

### Watch State Transitions

```bash
# Real-time transition monitoring
tail -f .sos/sessions/.audit/transitions.log
```

### Check for Errors

```bash
# Recent errors
tail -20 .sos/sessions/.audit/errors.log

# Count errors by type
awk -F'|' '{print $4}' .sos/sessions/.audit/errors.log | sort | uniq -c
```

### Session Statistics

```bash
# Count sessions by state
for dir in .sos/sessions/session-*; do
    if [[ -f "$dir/SESSION_CONTEXT.md" ]]; then
        grep "^status:" "$dir/SESSION_CONTEXT.md" | cut -d: -f2 | tr -d ' "'
    fi
done | sort | uniq -c
```

## Related Documentation

- [Architecture Overview](./ARCHITECTURE.md) - Component design
- [ADR-0001](../decisions/ADR-0001-session-state-machine-redesign.md) - Design rationale
- [TDD](../design/TDD-session-state-machine.md) - Technical specification

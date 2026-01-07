# Session State Machine Architecture

This document describes the architecture of the session state machine (FSM) system, which provides centralized, formally-verified session lifecycle management for the roster ecosystem.

## Overview

The session FSM provides:
- **Single source of truth** for session state via the `status` field
- **Formal verification** through TLA+ specifications with CI integration
- **Advisory locking** to prevent race conditions in multi-process scenarios
- **Event emission** for observability and audit trails
- **Backward-compatible migration** from v1 to v2 schema

## State Machine Model

### States

```
                 +------------------------------------------+
                 |                 ACTIVE                   |
                 |  +------------------------------------+  |
                 |  | Substates (from ACTIVE_WORKFLOW):  |  |
    +------------+  |  requirements -> design ->         |  +------------+
    |            |  |  implementation -> validation      |  |            |
    |            |  +------------------------------------+  |            |
    |            +------------------------------------------+            |
    |                   |                        ^                       |
    |                   | /park                  | /resume               |
    |                   v                        |                       |
    |            +------------------------------------------+            |
    |            |                 PARKED                   |            |
    |            |  (preserves current_phase for resume)   |            |
    +------------+------------------------------------------+------------+
         |                           |                            |
         |                           | /wrap (archive)            |
         |                           v                            |
         |       +------------------------------------------+     |
         +------>|                ARCHIVED                  |<----+
                 |  (terminal state, read-only)             |
                 +------------------------------------------+
```

### Transition Matrix

| From     | To       | Command     | Valid | Guard Conditions           |
|----------|----------|-------------|-------|----------------------------|
| NONE     | ACTIVE   | create      | Yes   | Session ID unique          |
| ACTIVE   | PARKED   | park        | Yes   | None                       |
| ACTIVE   | ARCHIVED | wrap        | Yes   | None                       |
| PARKED   | ACTIVE   | resume      | Yes   | None                       |
| PARKED   | ARCHIVED | wrap        | Yes   | None                       |
| ARCHIVED | *        | *           | No    | Terminal state (immutable) |

### TLA+ Invariants

The implementation enforces these invariants from `docs/specs/session-fsm.tla`:

| Invariant             | Description                                  |
|-----------------------|----------------------------------------------|
| TypeInvariant         | status in {ACTIVE, PARKED, ARCHIVED}         |
| NoInvalidTransitions  | All transitions obey ValidTransition         |
| ArchivedIsTerminal    | No transitions out of ARCHIVED               |
| MutualExclusion       | At most one process holds exclusive lock     |
| PhaseConsistency      | current_phase meaningful only in ACTIVE      |

## Component Architecture

```
                              +---------------------+
                              |  Claude Code CLI    |
                              +----------+----------+
                                         |
          +------------------------------+------------------------------+
          |                              |                              |
          v                              v                              v
+------------------+           +------------------+           +------------------+
|     Hooks        |           |   state-mate     |           |  session-mgr     |
| (Fast-Path)      |           |   (Agent)        |           |   CLI            |
+--------+---------+           +--------+---------+           +--------+---------+
         |                              |                              |
         +------------------------------+------------------------------+
                                        |
                                        v
                       +--------------------------------+
                       |      Session FSM Module        |
                       |  +---------------------------+ |
                       |  | State Transition Engine   | |
                       |  +---------------------------+ |
                       |  | Lock Manager              | |
                       |  +---------------------------+ |
                       |  | Schema Validator          | |
                       |  +---------------------------+ |
                       |  | Event Emitter             | |
                       |  +---------------------------+ |
                       +---------------+----------------+
                                       |
          +----------------------------+----------------------------+
          |                                                         |
          v                                                         v
+------------------+                                       +------------------+
|  SESSION_CONTEXT |                                       |   Audit Log      |
|  .md (v2)        |                                       |   events.jsonl   |
+------------------+                                       +------------------+
```

## File Layout

```
user-hooks/lib/
  session-fsm.sh        # Core state machine (830 lines)
  session-migrate.sh    # v1 -> v2 migration (681 lines)
  session-manager.sh    # CLI interface, integrates FSM
  session-utils.sh      # Shared utilities
  primitives.sh         # Low-level I/O primitives

docs/specs/
  session-fsm.tla       # TLA+ formal specification
  session-fsm.cfg       # TLC model checker config
  session-permissions.als  # Alloy permission model

docs/design/
  TDD-session-state-machine.md  # Technical design (1145 lines)

docs/decisions/
  ADR-0001-session-state-machine-redesign.md  # Architecture decision

tests/session-fsm/
  test_state_transitions.bats   # State transition tests
  test_locking.bats             # Concurrency tests
  test_migration.bats           # Migration tests
  test_helpers.bash             # Test utilities
  fixtures/                     # Test fixtures

.claude/sessions/
  .locks/                       # Advisory lock files
    <session-id>.lock           # flock target
    <session-id>.lock.d/        # mkdir fallback
  .audit/
    transitions.log             # Global state transitions
    errors.log                  # Error events
    migrations.log              # Migration audit trail
  <session-id>/
    SESSION_CONTEXT.md          # Session state (v2 schema)
    events.jsonl                # Session-specific events
```

## Component Responsibilities

### session-fsm.sh (Core State Machine)

The core module provides all FSM operations:

| Function               | Description                              |
|------------------------|------------------------------------------|
| `fsm_get_state`        | Get current session state (with locking) |
| `fsm_transition`       | Execute validated state transition       |
| `fsm_create_session`   | Create new session (NONE -> ACTIVE)      |
| `_fsm_lock_shared`     | Acquire shared lock for reads            |
| `_fsm_lock_exclusive`  | Acquire exclusive lock for writes        |
| `_fsm_unlock`          | Release any held lock                    |
| `_fsm_validate_context`| Validate session against v2 schema       |
| `_fsm_emit_event`      | Emit state transition event              |
| `_fsm_is_valid_transition` | Check transition validity            |

**Configuration (Environment Variables)**:

| Variable            | Default            | Description                    |
|---------------------|--------------------|--------------------------------|
| FSM_SESSIONS_DIR    | .claude/sessions   | Session storage directory      |
| FSM_LOCK_TIMEOUT    | 10                 | Lock timeout in seconds        |
| FSM_VALIDATE_SCHEMA | true               | Enable schema validation       |
| FSM_EMIT_EVENTS     | true               | Enable event emission          |

### session-migrate.sh (Migration Engine)

Handles v1 to v2 schema migration:

| Function              | Description                              |
|-----------------------|------------------------------------------|
| `migrate_session`     | Migrate single session to v2             |
| `migrate_all_sessions`| Batch migrate all v1 sessions            |
| `rollback_session`    | Restore from backup                      |
| `report_status`       | Show migration status                    |
| `auto_migrate_if_needed` | Called by session-manager on access  |

**Migration preserves**:
- All existing fields except legacy duplicates
- Session history via event log extraction
- Rollback capability via `.v1.backup` files

### session-manager.sh (CLI Interface)

Unified CLI for session operations:

| Command              | Description                              |
|----------------------|------------------------------------------|
| `status`             | Output JSON with full session state      |
| `create <init> <cx>` | Create new session                       |
| `mutate park`        | Park current session                     |
| `mutate resume`      | Resume parked session                    |
| `mutate wrap`        | Archive session                          |
| `transition <from> <to>` | Phase transition within ACTIVE       |

## Integration Points

### Hooks Integration

Hooks call into session-fsm.sh for fast-path operations:

```bash
# In hook scripts
source "$SCRIPT_DIR/session-fsm.sh"

# Check state before operation
state=$(fsm_get_state "$session_id")
if [[ "$state" == "PARKED" ]]; then
    echo "Session is parked. Run /resume first."
    exit 1
fi
```

### state-mate Integration

The state-mate agent delegates to FSM for mutations:

```bash
# state-mate uses session-fsm.sh for validated mutations
result=$(fsm_transition "$session_id" "PARKED" '{"reason":"User requested"}')
```

### Auto-Migration

session-manager.sh triggers auto-migration on first access:

```bash
# In cmd_status()
auto_migrate_if_needed "$session_id" 2>/dev/null || true
```

## Locking Strategy

### Primary: flock (Linux/macOS)

```bash
# Shared lock for reads
flock -s -w 10 "$lock_fd" 2>/dev/null

# Exclusive lock for writes
flock -x -w 10 "$lock_fd" 2>/dev/null
```

### Fallback: mkdir (Portable)

```bash
# Atomic directory creation as lock
if mkdir "$lock_marker" 2>/dev/null; then
    echo "$$" > "$lock_marker/pid"
    # Lock acquired
fi
```

### Stale Lock Detection

Locks include PID for detecting dead processes:

```bash
if [[ -f "$lock_marker/pid" ]]; then
    owner_pid=$(cat "$lock_marker/pid")
    if ! kill -0 "$owner_pid" 2>/dev/null; then
        # Owner dead, remove stale lock
        rm -rf "$lock_marker"
    fi
fi
```

## Schema Versions

### v1 Schema (Legacy)

```yaml
---
session_id: "session-..."
session_state: "ACTIVE"          # Or absent
parked_at: "2025-..."            # Park state indicator
park_reason: "..."               # Or parked_reason
git_status_at_park: "..."        # Or parked_git_status
---
```

**Problems**:
- Dual state determination (status vs parked_at presence)
- Duplicate field names
- No schema version tracking

### v2 Schema (Current)

```yaml
---
schema_version: "2.0"
session_id: "session-..."
status: "ACTIVE"                 # Single source of truth
created_at: "2025-..."
initiative: "Feature X"
complexity: "MODULE"
active_team: "10x-dev"
current_phase: "requirements"
---
```

**Improvements**:
- `status` is the ONLY state authority
- Park metadata moves to `events.jsonl`
- All legacy field variants removed
- Schema version enables format detection

## Event System

### Event Types

| Event            | Trigger                    | Payload                      |
|------------------|----------------------------|------------------------------|
| SESSION_CREATED  | NONE -> ACTIVE             | initiative, team             |
| SESSION_PARKED   | ACTIVE -> PARKED           | reason, git_status           |
| SESSION_RESUMED  | PARKED -> ACTIVE           | -                            |
| SESSION_ARCHIVED | * -> ARCHIVED              | -                            |
| SCHEMA_MIGRATED  | v1 -> v2 migration         | from_version, to_version     |

### Event Log Format (JSONL)

```jsonl
{"timestamp":"2025-12-31T12:00:00Z","event":"SESSION_CREATED","from":"NONE","to":"ACTIVE","metadata":{"initiative":"Feature X"}}
{"timestamp":"2025-12-31T14:00:00Z","event":"SESSION_PARKED","from":"ACTIVE","to":"PARKED","metadata":{"reason":"Lunch break"}}
```

### Audit Trail

Global logs in `.claude/sessions/.audit/`:

```
# transitions.log
2025-12-31T12:00:00Z | session-20251231-120000-abc | SESSION_CREATED | NONE -> ACTIVE

# errors.log
2025-12-31T12:00:05Z | ERROR | session-abc | LOCK_TIMEOUT | PARKED
```

## Extension Points

### Adding New States

1. Update TLA+ spec with new state and transitions
2. Run CI verification to ensure invariants hold
3. Update `_fsm_is_valid_transition()` in session-fsm.sh
4. Add event type for new transitions
5. Update schema if new fields required

### Adding New Substates (Phases)

1. Update `ACTIVE_WORKFLOW.yaml` with new phase
2. Add guard conditions in workflow engine
3. No FSM changes needed (phases are workflow-controlled)

### Custom Event Handlers

Events can be consumed by:
- Log aggregators (parse JSONL)
- Monitoring systems (tail audit logs)
- Future: webhooks, metrics emission

## Testing Strategy

### Unit Tests (BATS)

```bash
# Run all session FSM tests
bats tests/session-fsm/

# Run specific test file
bats tests/session-fsm/test_state_transitions.bats
```

### Formal Verification

```bash
# Quick syntax check
./scripts/verify-specs.sh

# Full model checking (requires Java)
./scripts/verify-specs.sh --full
```

### CI Verification

The GitHub Actions workflow `.github/workflows/verify-formal-specs.yml` runs:
- TLC model checker on `session-fsm.tla`
- Alloy Analyzer on `session-permissions.als`

## Related Documentation

- [Operations Guide](./OPERATIONS.md) - Commands and troubleshooting
- [ADR-0001](../decisions/ADR-0001-session-state-machine-redesign.md) - Design rationale
- [TDD](../design/TDD-session-state-machine.md) - Technical specification
- [TLA+ Spec](../specs/session-fsm.tla) - Formal model

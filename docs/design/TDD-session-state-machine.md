# TDD: Session State Machine Redesign

## Overview

This Technical Design Document specifies the implementation of a formal finite state machine (FSM) for session lifecycle management in the roster system. The design eliminates dual-state determination bugs, enforces valid transitions, provides concurrency safety through advisory locking, and enables migration from the legacy schema to a single-source-of-truth model.

## Context

| Reference | Location |
|-----------|----------|
| ADR | `docs/decisions/ADR-0001-session-state-machine-redesign.md` |
| TLA+ Spec | `docs/specs/session-fsm.tla` |
| Alloy Spec | `docs/specs/session-permissions.als` |
| Current Implementation | `user-hooks/lib/session-manager.sh`, `user-hooks/lib/session-state.sh` |
| Schema Definition | `schemas/artifacts/session-context.schema.json` |
| Related ADR | `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` |

### Problem Statement

The current session management system has four critical defects:

1. **Dual State Determination**: `extract_session_fields()` reads the `status` field while `get_session_state()` infers state from `parked_at` presence, causing conflicting results.
2. **Field Naming Chaos**: Three pairs of duplicate fields (`status`/`session_state`, `parked_reason`/`park_reason`, `parked_git_status`/`git_status_at_park`).
3. **Race Conditions**: Inconsistent locking across operations allows stale reads.
4. **Undefined Transitions**: No enforcement of valid state sequences.

### Design Goals

1. Single source of truth for session state
2. Formal state transition enforcement derived from TLA+ specification
3. Advisory locking for all state reads and mutations
4. Clean migration path from v1 to v2 schema
5. Event emission for observability

---

## System Design

### Architecture Diagram

```
                                    +-----------------------+
                                    |   Claude Code CLI     |
                                    +-----------+-----------+
                                                |
                    +---------------------------+---------------------------+
                    |                           |                           |
                    v                           v                           v
          +------------------+        +------------------+        +------------------+
          |   Hooks          |        |   state-mate     |        |   session-mgr    |
          | (Fast-Path)      |        |   (Agent)        |        |   CLI            |
          +--------+---------+        +--------+---------+        +--------+---------+
                   |                           |                           |
                   +---------------------------+---------------------------+
                                               |
                                               v
                              +--------------------------------+
                              |    Session FSM Module          |
                              |  +---------------------------+ |
                              |  | State Transition Engine   | |
                              |  +---------------------------+ |
                              |  | Lock Manager              | |
                              |  +---------------------------+ |
                              |  | Schema Validator          | |
                              |  +---------------------------+ |
                              |  | Event Emitter             | |
                              |  +---------------------------+ |
                              +----------------+---------------+
                                               |
                    +---------------------------+---------------------------+
                    |                                                       |
                    v                                                       v
          +------------------+                                    +------------------+
          |  SESSION_CONTEXT |                                    |   Audit Log      |
          |  .md (v2)        |                                    |                  |
          +------------------+                                    +------------------+
```

### Components

| Component | Responsibility | Technology | Location |
|-----------|---------------|------------|----------|
| **Session FSM Module** | Encapsulates all state machine logic | Bash | `user-hooks/lib/session-fsm.sh` |
| **State Transition Engine** | Validates and executes transitions | Bash | Part of `session-fsm.sh` |
| **Lock Manager** | Advisory locking (flock/mkdir) | Bash | Part of `session-fsm.sh` |
| **Schema Validator** | JSON Schema validation | Bash + ajv-cli (optional) | `user-hooks/lib/session-validator.sh` |
| **Event Emitter** | Publishes state change events | Bash | Part of `session-fsm.sh` |
| **Migration Engine** | v1 to v2 schema migration | Bash | `user-hooks/lib/session-migrate.sh` |

### Module Dependency Graph

```
session-fsm.sh
    |
    +-- session-validator.sh
    |       |
    |       +-- primitives.sh
    |               |
    |               +-- config.sh
    |
    +-- primitives.sh (shared)
```

---

## State Management Implementation

### State Machine Definition

From TLA+ specification (`docs/specs/session-fsm.tla`), the session FSM has:

**Top-Level States**:
| State | Description | Mutable | Valid Transitions Out |
|-------|-------------|---------|----------------------|
| `NONE` | No session exists | N/A | ACTIVE |
| `ACTIVE` | Session in progress | Yes | PARKED, ARCHIVED |
| `PARKED` | Session suspended | Limited | ACTIVE, ARCHIVED |
| `ARCHIVED` | Session complete | No (immutable) | None (terminal) |

**Transition Matrix**:
```
             | NONE | ACTIVE | PARKED | ARCHIVED |
-------------|------|--------|--------|----------|
  NONE       |  -   |   *    |   -    |    -     |
  ACTIVE     |  -   |   -    |   *    |    *     |
  PARKED     |  -   |   *    |   -    |    *     |
  ARCHIVED   |  -   |   -    |   -    |    -     |

* = Valid transition
- = Invalid transition
```

### Single Source of Truth

The `status` field is the ONLY authority for top-level state:

```yaml
# SESSION_CONTEXT.md (v2 Schema)
---
schema_version: "2.0"
session_id: "session-20251231-120000-abcd1234"
status: "ACTIVE"                    # ONLY this field determines state
current_phase: "design"             # Substate within ACTIVE (from workflow)
created_at: "2025-12-31T12:00:00Z"
initiative: "Feature X"
complexity: "MODULE"
active_team: "10x-dev-pack"
---
```

**Removed Fields** (v2 eliminates):
- `session_state` (duplicate of `status`)
- `parked_at`, `auto_parked_at` (status=PARKED is sufficient)
- `park_reason`, `parked_reason` (consolidated to event log)
- `git_status_at_park`, `parked_git_status` (consolidated to event log)

**Preserved Metadata** (moved to event log):
```yaml
# .claude/sessions/<session-id>/events.jsonl
{"timestamp":"2025-12-31T14:00:00Z","event":"PARKED","reason":"Lunch break","git_status":"clean"}
{"timestamp":"2025-12-31T15:00:00Z","event":"RESUMED"}
```

### Substate Derivation

Substates are derived from `ACTIVE_WORKFLOW.yaml`, not stored redundantly:

```bash
# Get current phase (substate) when status=ACTIVE
get_current_phase() {
    local session_id="$1"
    local ctx_file=".claude/sessions/$session_id/SESSION_CONTEXT.md"

    # Only valid if status=ACTIVE
    local status
    status=$(get_status "$session_id")
    [[ "$status" == "ACTIVE" ]] || { echo "none"; return; }

    # Read from context file
    get_yaml_field "$ctx_file" "current_phase"
}
```

---

## Interface Contracts

### Session FSM API

#### `fsm_get_state(session_id) -> state`

Returns current state with proper locking.

```bash
# Returns: NONE | ACTIVE | PARKED | ARCHIVED
# Exit code: 0 on success, 1 on error

fsm_get_state() {
    local session_id="$1"
    local ctx_file=".claude/sessions/$session_id/SESSION_CONTEXT.md"

    # Acquire shared lock for read
    _fsm_lock_shared "$session_id" || return 1

    if [[ ! -f "$ctx_file" ]]; then
        _fsm_unlock "$session_id"
        echo "NONE"
        return 0
    fi

    local status
    status=$(get_yaml_field "$ctx_file" "status")

    _fsm_unlock "$session_id"
    echo "${status:-ACTIVE}"  # Default to ACTIVE for v1 compat
}
```

#### `fsm_transition(session_id, target_state, metadata) -> result`

Executes a state transition with full validation.

```bash
# Returns: JSON result object
# Exit code: 0 on success, 1 on invalid transition, 2 on lock failure

fsm_transition() {
    local session_id="$1"
    local target_state="$2"
    local metadata="$3"  # JSON object with operation-specific data

    # Acquire exclusive lock
    _fsm_lock_exclusive "$session_id" || {
        _emit_error "LOCK_TIMEOUT" "$session_id" "$target_state"
        return 2
    }

    local current_state
    current_state=$(fsm_get_state_unlocked "$session_id")

    # Validate transition
    if ! _fsm_is_valid_transition "$current_state" "$target_state"; then
        _fsm_unlock "$session_id"
        _emit_error "INVALID_TRANSITION" "$session_id" "$current_state" "$target_state"
        return 1
    fi

    # Create backup
    local ctx_file=".claude/sessions/$session_id/SESSION_CONTEXT.md"
    local backup_file="$ctx_file.backup"
    cp "$ctx_file" "$backup_file"

    # Execute transition
    if ! _fsm_execute_transition "$session_id" "$current_state" "$target_state" "$metadata"; then
        mv "$backup_file" "$ctx_file"
        _fsm_unlock "$session_id"
        return 1
    fi

    # Validate result
    if ! _fsm_validate_context "$ctx_file"; then
        mv "$backup_file" "$ctx_file"
        _fsm_unlock "$session_id"
        _emit_error "VALIDATION_FAILED" "$session_id"
        return 1
    fi

    # Emit event
    _fsm_emit_event "$session_id" "$current_state" "$target_state" "$metadata"

    rm -f "$backup_file"
    _fsm_unlock "$session_id"

    echo "{\"success\": true, \"from\": \"$current_state\", \"to\": \"$target_state\"}"
}
```

#### `fsm_create_session(initiative, complexity, team) -> session_id`

Creates a new session (NONE -> ACTIVE transition).

```bash
fsm_create_session() {
    local initiative="$1"
    local complexity="$2"
    local team="$3"

    local session_id
    session_id=$(generate_session_id)

    # Initialize context file
    local session_dir=".claude/sessions/$session_id"
    mkdir -p "$session_dir"

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat > "$session_dir/SESSION_CONTEXT.md" <<CONTEXT
---
schema_version: "2.0"
session_id: "$session_id"
status: "ACTIVE"
created_at: "$timestamp"
initiative: "$initiative"
complexity: "$complexity"
active_team: "$team"
current_phase: "requirements"
---

# Session: $initiative

## Artifacts
- PRD: pending
- TDD: pending

## Blockers
None yet.

## Next Steps
1. Complete requirements gathering
CONTEXT

    # Validate
    if ! _fsm_validate_context "$session_dir/SESSION_CONTEXT.md"; then
        rm -rf "$session_dir"
        return 1
    fi

    # Set as current session
    set_current_session "$session_id"

    # Emit creation event
    _fsm_emit_event "$session_id" "NONE" "ACTIVE" "{\"initiative\":\"$initiative\"}"

    echo "$session_id"
}
```

### Transition Validation Table

| From | To | Guard Conditions | Side Effects |
|------|-----|------------------|--------------|
| NONE | ACTIVE | Session ID unique | Create session dir, emit `SESSION_CREATED` |
| ACTIVE | PARKED | None | Emit `SESSION_PARKED` with reason |
| ACTIVE | ARCHIVED | None | Move to archive, emit `SESSION_ARCHIVED` |
| PARKED | ACTIVE | None | Emit `SESSION_RESUMED` |
| PARKED | ARCHIVED | None | Move to archive, emit `SESSION_ARCHIVED` |

---

## Locking Strategy

### Advisory Lock Implementation

Primary mechanism: `flock` (available via coreutils on macOS, native on Linux)

Fallback mechanism: `mkdir`-based locking (portable, less elegant)

#### Lock File Location

```
.claude/sessions/.locks/
    <session-id>.lock       # flock target file
    <session-id>.lock.d/    # mkdir fallback directory
        pid                 # Owner process ID
```

#### Shared Lock (Read Operations)

```bash
_fsm_lock_shared() {
    local session_id="$1"
    local lock_file=".claude/sessions/.locks/${session_id}.lock"
    local timeout="${FSM_LOCK_TIMEOUT:-10}"

    mkdir -p "$(dirname "$lock_file")" 2>/dev/null

    if command -v flock >/dev/null 2>&1; then
        # Shared lock with timeout
        exec 200>"$lock_file"
        flock -s -w "$timeout" 200 2>/dev/null
    else
        # Fallback: treat shared as exclusive (conservative)
        _fsm_lock_exclusive "$session_id"
    fi
}
```

#### Exclusive Lock (Write Operations)

```bash
_fsm_lock_exclusive() {
    local session_id="$1"
    local lock_file=".claude/sessions/.locks/${session_id}.lock"
    local timeout="${FSM_LOCK_TIMEOUT:-10}"

    mkdir -p "$(dirname "$lock_file")" 2>/dev/null

    if command -v flock >/dev/null 2>&1; then
        # Exclusive lock with timeout
        exec 200>"$lock_file"
        if flock -x -w "$timeout" 200 2>/dev/null; then
            echo "$$" >&200
            return 0
        else
            exec 200>&-
            return 1
        fi
    else
        # Fallback: mkdir-based locking
        local lock_marker="${lock_file}.d"
        local elapsed=0

        while [[ "$elapsed" -lt "$timeout" ]]; do
            if mkdir "$lock_marker" 2>/dev/null; then
                echo "$$" > "$lock_marker/pid"
                return 0
            fi

            # Check for stale lock
            if [[ -f "$lock_marker/pid" ]]; then
                local owner_pid
                owner_pid=$(cat "$lock_marker/pid" 2>/dev/null)
                if [[ -n "$owner_pid" ]] && ! kill -0 "$owner_pid" 2>/dev/null; then
                    rm -rf "$lock_marker" 2>/dev/null
                    continue
                fi
            fi

            sleep 0.1
            ((elapsed++))
        done

        return 1
    fi
}
```

#### Lock Release

```bash
_fsm_unlock() {
    local session_id="$1"
    local lock_file=".claude/sessions/.locks/${session_id}.lock"
    local lock_marker="${lock_file}.d"

    # Release flock
    if command -v flock >/dev/null 2>&1; then
        exec 200>&- 2>/dev/null || true
    fi

    # Remove mkdir lock
    rm -rf "$lock_marker" 2>/dev/null || true
}
```

### Deadlock Prevention

1. **Timeout**: All lock acquisitions have a configurable timeout (default: 10s)
2. **Single Lock Per Session**: Only one lock type per session (no nested locks)
3. **Stale Lock Detection**: Checks if lock owner PID is alive
4. **Lock Ordering**: Not required (single lock per session)

### Liveness Properties (from TLA+)

The implementation satisfies these properties from `docs/specs/session-fsm.tla`:

1. **LockEventuallyGranted**: With fair scheduling, every lock request is eventually granted
2. **NoDeadlock**: System can always make progress (no circular wait)
3. **MutualExclusion**: At most one process holds exclusive lock

---

## Schema Validation

### JSON Schema (v2)

Update `schemas/artifacts/session-context.schema.json`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://roster.local/schemas/artifacts/session-context-v2.schema.json",
  "title": "SESSION_CONTEXT Schema v2",
  "description": "Schema for SESSION_CONTEXT.md YAML frontmatter (v2 - FSM compliant)",
  "type": "object",

  "required": [
    "schema_version",
    "session_id",
    "status",
    "created_at",
    "initiative",
    "complexity",
    "active_team",
    "current_phase"
  ],

  "properties": {
    "schema_version": {
      "const": "2.0",
      "description": "Schema version (must be 2.0 for FSM-compliant sessions)"
    },
    "session_id": {
      "$ref": "common.schema.json#/$defs/session_id"
    },
    "status": {
      "type": "string",
      "enum": ["ACTIVE", "PARKED", "ARCHIVED"],
      "description": "Current session status (single source of truth)"
    },
    "created_at": {
      "$ref": "common.schema.json#/$defs/iso8601_timestamp"
    },
    "initiative": {
      "type": "string",
      "minLength": 1,
      "description": "Initiative or feature being worked on"
    },
    "complexity": {
      "$ref": "common.schema.json#/$defs/complexity_enum"
    },
    "active_team": {
      "$ref": "common.schema.json#/$defs/rite_name"
    },
    "current_phase": {
      "type": "string",
      "minLength": 1,
      "description": "Current workflow phase (e.g., requirements, design)"
    },
    "last_accessed_at": {
      "$ref": "common.schema.json#/$defs/iso8601_timestamp"
    },
    "archived_at": {
      "$ref": "common.schema.json#/$defs/iso8601_timestamp"
    }
  },

  "additionalProperties": true
}
```

### Validation Implementation

```bash
_fsm_validate_context() {
    local ctx_file="$1"

    # Check file exists
    [[ -f "$ctx_file" ]] || return 1

    # Required fields for v2
    local required=("schema_version" "session_id" "status" "created_at"
                    "initiative" "complexity" "active_team" "current_phase")

    for field in "${required[@]}"; do
        if ! grep -q "^${field}:" "$ctx_file" 2>/dev/null; then
            echo "Missing required field: $field" >&2
            return 1
        fi
    done

    # Validate status is a valid state
    local status
    status=$(get_yaml_field "$ctx_file" "status")
    case "$status" in
        ACTIVE|PARKED|ARCHIVED) ;;
        *)
            echo "Invalid status: $status" >&2
            return 1
            ;;
    esac

    # Validate schema version
    local version
    version=$(get_yaml_field "$ctx_file" "schema_version")
    if [[ "$version" != "2.0" ]]; then
        echo "Unsupported schema version: $version" >&2
        return 1
    fi

    # Optional: Full JSON Schema validation with ajv-cli
    if command -v ajv >/dev/null 2>&1; then
        # Extract frontmatter as JSON and validate
        local json
        json=$(_extract_frontmatter_json "$ctx_file")
        echo "$json" | ajv validate -s "schemas/artifacts/session-context-v2.schema.json" --strict=false 2>/dev/null
    fi

    return 0
}
```

---

## Migration Design

### Overview

Migration converts v1 SESSION_CONTEXT.md files to v2 format:

1. **Field Canonicalization**: Unify duplicate fields
2. **State Derivation**: Compute `status` from legacy fields
3. **Metadata Extraction**: Move park metadata to event log
4. **Schema Upgrade**: Set `schema_version: "2.0"`

### Migration Script

Location: `user-hooks/lib/session-migrate.sh`

```bash
#!/bin/bash
# Session schema migration: v1 -> v2

migrate_session() {
    local session_dir="$1"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"
    local backup_file="$session_dir/SESSION_CONTEXT.v1.backup"
    local events_file="$session_dir/events.jsonl"

    # Skip if already v2
    local version
    version=$(get_yaml_field "$ctx_file" "schema_version" 2>/dev/null)
    if [[ "$version" == "2.0" ]]; then
        echo "Already v2: $session_dir"
        return 0
    fi

    # Create backup
    cp "$ctx_file" "$backup_file" || return 1

    # Determine canonical state
    local new_status="ACTIVE"
    if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
        new_status="PARKED"
    fi
    if grep -q "^completed_at:" "$ctx_file" 2>/dev/null; then
        new_status="ARCHIVED"
    fi

    # Extract park metadata for event log
    local parked_at=""
    local parked_reason=""
    local git_status=""

    parked_at=$(get_yaml_field "$ctx_file" "parked_at" 2>/dev/null || \
                get_yaml_field "$ctx_file" "auto_parked_at" 2>/dev/null)
    parked_reason=$(get_yaml_field "$ctx_file" "parked_reason" 2>/dev/null || \
                    get_yaml_field "$ctx_file" "park_reason" 2>/dev/null || \
                    get_yaml_field "$ctx_file" "auto_parked_reason" 2>/dev/null)
    git_status=$(get_yaml_field "$ctx_file" "parked_git_status" 2>/dev/null || \
                 get_yaml_field "$ctx_file" "git_status_at_park" 2>/dev/null)

    # Write park event to event log
    if [[ -n "$parked_at" ]]; then
        local event="{\"timestamp\":\"$parked_at\",\"event\":\"PARKED\""
        [[ -n "$parked_reason" ]] && event+=",\"reason\":\"$parked_reason\""
        [[ -n "$git_status" ]] && event+=",\"git_status\":\"$git_status\""
        event+="}"
        echo "$event" >> "$events_file"
    fi

    # Remove legacy fields and add v2 fields
    local temp_file="${ctx_file}.tmp"

    awk -v status="$new_status" '
    BEGIN { in_frontmatter=0; frontmatter_count=0; status_written=0; version_written=0 }

    /^---$/ {
        frontmatter_count++
        if (frontmatter_count == 1) {
            in_frontmatter = 1
            print
            next
        }
        if (frontmatter_count == 2) {
            # Write v2 fields before closing ---
            if (!version_written) print "schema_version: \"2.0\""
            if (!status_written) print "status: \"" status "\""
            in_frontmatter = 0
            print
            next
        }
    }

    in_frontmatter {
        # Skip legacy fields
        if (/^(session_state|parked_at|auto_parked_at|park_reason|parked_reason|parked_git_status|git_status_at_park|auto_parked_reason):/) {
            next
        }
        # Track if we see v2 fields already
        if (/^schema_version:/) { version_written=1 }
        if (/^status:/) { status_written=1 }
    }

    { print }
    ' "$ctx_file" > "$temp_file"

    mv "$temp_file" "$ctx_file"

    # Validate migrated file
    if ! _fsm_validate_context "$ctx_file"; then
        # Rollback
        mv "$backup_file" "$ctx_file"
        echo "Migration failed validation: $session_dir" >&2
        return 1
    fi

    echo "Migrated: $session_dir (status=$new_status)"
    return 0
}

# Batch migration for all sessions
migrate_all_sessions() {
    local sessions_dir=".claude/sessions"
    local success=0
    local failed=0

    for session_dir in "$sessions_dir"/session-*; do
        [[ -d "$session_dir" ]] || continue

        if migrate_session "$session_dir"; then
            ((success++))
        else
            ((failed++))
        fi
    done

    echo "Migration complete: $success succeeded, $failed failed"
    return $([[ $failed -eq 0 ]] && echo 0 || echo 1)
}
```

### Field Canonicalization Map

| v1 Field(s) | v2 Field | Migration Logic |
|-------------|----------|-----------------|
| `status`, `session_state` | `status` | Use `status` if present, else `session_state`, else derive from presence |
| `parked_at`, `auto_parked_at` | (removed) | Move to `events.jsonl` |
| `park_reason`, `parked_reason`, `auto_parked_reason` | (removed) | Move to `events.jsonl` |
| `git_status_at_park`, `parked_git_status` | (removed) | Move to `events.jsonl` |
| (new) | `schema_version` | Set to `"2.0"` |

### Rollback Procedure

```bash
rollback_session() {
    local session_dir="$1"
    local backup_file="$session_dir/SESSION_CONTEXT.v1.backup"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"

    if [[ -f "$backup_file" ]]; then
        mv "$backup_file" "$ctx_file"
        echo "Rolled back: $session_dir"
        return 0
    else
        echo "No backup found: $session_dir" >&2
        return 1
    fi
}

rollback_all_sessions() {
    local sessions_dir=".claude/sessions"

    for session_dir in "$sessions_dir"/session-*; do
        [[ -d "$session_dir" ]] || continue
        rollback_session "$session_dir"
    done
}
```

### Migration CLI

```bash
# Run migration with dry-run option
./user-hooks/lib/session-migrate.sh migrate --dry-run

# Run actual migration
./user-hooks/lib/session-migrate.sh migrate

# Rollback all migrations
./user-hooks/lib/session-migrate.sh rollback
```

---

## Error Handling

### Error Categories

| Category | Code | Description | Recovery |
|----------|------|-------------|----------|
| `LOCK_TIMEOUT` | 2 | Could not acquire lock within timeout | Retry after delay |
| `INVALID_TRANSITION` | 1 | Requested transition not allowed | Return error to caller |
| `VALIDATION_FAILED` | 1 | Context file fails schema validation | Rollback from backup |
| `SESSION_NOT_FOUND` | 1 | Session directory does not exist | Return NONE state |
| `WRITE_FAILED` | 1 | Could not write to context file | Rollback from backup |

### Structured Error Output

```bash
_emit_error() {
    local error_type="$1"
    local session_id="$2"
    shift 2
    local details="$*"

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat >&2 <<EOF
{
  "error": "$error_type",
  "session_id": "$session_id",
  "timestamp": "$timestamp",
  "details": "$details"
}
EOF

    # Log to audit trail
    echo "$timestamp | ERROR | $session_id | $error_type | $details" >> \
        ".claude/sessions/.audit/errors.log"
}
```

### Partial Write Recovery

All mutations follow the pattern:

1. Create backup of current file
2. Perform mutation
3. Validate result
4. On failure: restore from backup
5. On success: remove backup

```bash
_fsm_safe_mutate() {
    local ctx_file="$1"
    local mutation_func="$2"
    shift 2
    local args=("$@")

    local backup_file="${ctx_file}.backup.$$"

    # Create backup
    cp "$ctx_file" "$backup_file" || {
        _emit_error "BACKUP_FAILED" "$(basename "$(dirname "$ctx_file")")"
        return 1
    }

    # Execute mutation
    if ! "$mutation_func" "$ctx_file" "${args[@]}"; then
        mv "$backup_file" "$ctx_file"
        return 1
    fi

    # Validate
    if ! _fsm_validate_context "$ctx_file"; then
        mv "$backup_file" "$ctx_file"
        _emit_error "VALIDATION_FAILED" "$(basename "$(dirname "$ctx_file")")"
        return 1
    fi

    # Success - remove backup
    rm -f "$backup_file"
    return 0
}
```

---

## Event Emission

### Event Types

| Event | Payload Fields | Trigger |
|-------|---------------|---------|
| `SESSION_CREATED` | session_id, initiative, complexity, team | fsm_create_session |
| `SESSION_PARKED` | session_id, reason, git_status | ACTIVE -> PARKED |
| `SESSION_RESUMED` | session_id | PARKED -> ACTIVE |
| `SESSION_ARCHIVED` | session_id | * -> ARCHIVED |
| `PHASE_CHANGED` | session_id, from_phase, to_phase | current_phase update |
| `VALIDATION_ERROR` | session_id, error_details | Validation failure |
| `LOCK_CONTENTION` | session_id, wait_time | Lock wait > 1s |

### Event Log Format

Events are stored in JSONL format for efficient append and parsing:

```bash
# Location: .claude/sessions/<session-id>/events.jsonl

_fsm_emit_event() {
    local session_id="$1"
    local from_state="$2"
    local to_state="$3"
    local metadata="$4"

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local event_type
    case "$from_state:$to_state" in
        "NONE:ACTIVE")    event_type="SESSION_CREATED" ;;
        "ACTIVE:PARKED")  event_type="SESSION_PARKED" ;;
        "PARKED:ACTIVE")  event_type="SESSION_RESUMED" ;;
        *":ARCHIVED")     event_type="SESSION_ARCHIVED" ;;
        *)                event_type="STATE_CHANGED" ;;
    esac

    local events_file=".claude/sessions/$session_id/events.jsonl"

    # Build event JSON
    local event="{\"timestamp\":\"$timestamp\",\"event\":\"$event_type\",\"from\":\"$from_state\",\"to\":\"$to_state\""
    if [[ -n "$metadata" && "$metadata" != "{}" ]]; then
        event+=",\"metadata\":$metadata"
    fi
    event+="}"

    echo "$event" >> "$events_file"

    # Also log to global audit log
    echo "$timestamp | $session_id | $event_type | $from_state -> $to_state" >> \
        ".claude/sessions/.audit/transitions.log"
}
```

---

## Test Strategy

### Unit Tests

Location: `tests/unit/session-fsm.bats`

| Test ID | Description | TLA+ Invariant |
|---------|-------------|----------------|
| `fsm_001` | Create session sets status to ACTIVE | TypeInvariant |
| `fsm_002` | Park session changes ACTIVE to PARKED | ValidTransition |
| `fsm_003` | Resume session changes PARKED to ACTIVE | ValidTransition |
| `fsm_004` | Archive from ACTIVE succeeds | ValidTransition |
| `fsm_005` | Archive from PARKED succeeds | ValidTransition |
| `fsm_006` | Resume from ACTIVE fails (invalid) | NoInvalidTransitions |
| `fsm_007` | Any transition from ARCHIVED fails | ArchivedIsTerminal |
| `fsm_008` | Status is only source of truth | PhaseConsistency |
| `fsm_009` | Missing required field fails validation | TypeInvariant |
| `fsm_010` | Invalid status value fails validation | TypeInvariant |

```bash
@test "fsm_003: Resume session changes PARKED to ACTIVE" {
    # Setup: Create parked session
    local session_id
    session_id=$(fsm_create_session "Test" "MODULE" "10x-dev-pack")
    fsm_transition "$session_id" "PARKED" '{"reason":"test"}'

    # Act: Resume
    run fsm_transition "$session_id" "ACTIVE" '{}'

    # Assert
    [ "$status" -eq 0 ]
    [[ "$output" == *'"success": true'* ]]

    local current_state
    current_state=$(fsm_get_state "$session_id")
    [ "$current_state" = "ACTIVE" ]
}

@test "fsm_007: Any transition from ARCHIVED fails" {
    # Setup: Create and archive session
    local session_id
    session_id=$(fsm_create_session "Test" "MODULE" "10x-dev-pack")
    fsm_transition "$session_id" "ARCHIVED" '{}'

    # Act: Attempt resume
    run fsm_transition "$session_id" "ACTIVE" '{}'

    # Assert: Invalid transition
    [ "$status" -eq 1 ]
    [[ "$output" == *'INVALID_TRANSITION'* ]]
}
```

### Integration Tests

Location: `tests/integration/session-concurrency.bats`

| Test ID | Description | TLA+ Property |
|---------|-------------|---------------|
| `int_001` | Concurrent reads see consistent state | LockedReadsAreConsistent |
| `int_002` | Concurrent writes are serialized | MutualExclusion |
| `int_003` | Lock timeout returns error, not hang | NoDeadlock |
| `int_004` | Stale lock is cleaned up | HolderNotInQueue |
| `int_005` | Migration preserves state semantics | - |

```bash
@test "int_002: Concurrent writes are serialized" {
    local session_id
    session_id=$(fsm_create_session "Concurrent" "MODULE" "10x-dev-pack")

    # Start two parallel park operations
    (fsm_transition "$session_id" "PARKED" '{"reason":"writer1"}') &
    local pid1=$!
    (fsm_transition "$session_id" "PARKED" '{"reason":"writer2"}') &
    local pid2=$!

    wait $pid1
    local status1=$?
    wait $pid2
    local status2=$?

    # Exactly one should succeed, one should fail (already parked)
    local success_count=$((($status1 == 0 ? 1 : 0) + ($status2 == 0 ? 1 : 0)))
    [ "$success_count" -eq 1 ]

    # Final state should be PARKED
    local final_state
    final_state=$(fsm_get_state "$session_id")
    [ "$final_state" = "PARKED" ]
}
```

### Property-Based Tests (Derived from TLA+)

| Property | Test Approach |
|----------|---------------|
| `NoInvalidTransitions` | Fuzzing: Generate random transition sequences, verify all invalid ones are rejected |
| `MutualExclusion` | Race condition: Spawn N processes, verify lock holder count never > 1 |
| `LockEventuallyGranted` | Stress: Queue 10 lock requests, verify all complete within timeout |
| `ArchivedIsTerminal` | Exhaustive: Try all possible transitions from ARCHIVED |

---

## Implementation Guidance

### Recommended Implementation Order

1. **Phase 1: Lock Manager** (`_fsm_lock_*` functions)
   - Implement flock and mkdir fallback
   - Add stale lock detection
   - Add timeout handling

2. **Phase 2: Schema Validator** (`_fsm_validate_context`)
   - Field presence validation
   - Status enum validation
   - Schema version check

3. **Phase 3: State Transition Engine**
   - Transition matrix implementation
   - Guard condition checks
   - Atomic mutation with backup/rollback

4. **Phase 4: Event Emitter**
   - JSONL event log
   - Global audit log
   - Error logging

5. **Phase 5: Migration Engine**
   - Field canonicalization
   - Metadata extraction to event log
   - Batch migration with dry-run

6. **Phase 6: API Surface**
   - `fsm_get_state`
   - `fsm_transition`
   - `fsm_create_session`
   - Integration with `session-manager.sh`

### Dependency Injection Points

```bash
# For testing, these can be overridden:
FSM_SESSIONS_DIR="${FSM_SESSIONS_DIR:-.claude/sessions}"
FSM_LOCK_TIMEOUT="${FSM_LOCK_TIMEOUT:-10}"
FSM_VALIDATE_SCHEMA="${FSM_VALIDATE_SCHEMA:-true}"
FSM_EMIT_EVENTS="${FSM_EMIT_EVENTS:-true}"
```

### Backward Compatibility Layer

During migration, the legacy API must continue to work:

```bash
# session-state.sh: Shim for backward compatibility
get_session_state() {
    local session_id="${1:-$(get_session_id)}"

    # Check if v2 schema
    local ctx_file=".claude/sessions/$session_id/SESSION_CONTEXT.md"
    local version
    version=$(get_yaml_field "$ctx_file" "schema_version" 2>/dev/null)

    if [[ "$version" == "2.0" ]]; then
        # v2: Use FSM
        fsm_get_state "$session_id"
    else
        # v1: Legacy behavior (infer from parked_at)
        _legacy_get_session_state "$session_id"
    fi
}
```

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Migration corrupts sessions | Low | High | Create backups, validate post-migration, provide rollback |
| flock unavailable on some systems | Medium | Medium | mkdir fallback implementation, document requirements |
| Lock contention under high concurrency | Low | Medium | Short critical sections, configurable timeout |
| Event log grows unbounded | Medium | Low | Add log rotation (future), archive with session |
| Legacy code bypasses FSM | Medium | High | PreToolUse hook intercepts direct writes to SESSION_CONTEXT |

---

## ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-0001 | Accepted | Session State Machine Redesign |
| ADR-0005 | Accepted | state-mate Centralized State Authority |

---

## Open Items

| Item | Status | Owner | Notes |
|------|--------|-------|-------|
| ajv-cli integration | Optional | Principal Engineer | Full JSON Schema validation in CI |
| Event log rotation | Deferred | Future Sprint | Archive events when session archived |
| Metrics emission | Deferred | Future Sprint | Prometheus/StatsD integration |

---

## Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-state-machine.md` | Pending |
| ADR-0001 | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0001-session-state-machine-redesign.md` | Read |
| TLA+ Spec | `/Users/tomtenuta/Code/roster/docs/specs/session-fsm.tla` | Read |
| Alloy Spec | `/Users/tomtenuta/Code/roster/docs/specs/session-permissions.als` | Read |
| Current session-manager.sh | `/Users/tomtenuta/Code/roster/user-hooks/lib/session-manager.sh` | Read |
| Current session-state.sh | `/Users/tomtenuta/Code/roster/user-hooks/lib/session-state.sh` | Read |
| Schema | `/Users/tomtenuta/Code/roster/schemas/artifacts/session-context.schema.json` | Read |

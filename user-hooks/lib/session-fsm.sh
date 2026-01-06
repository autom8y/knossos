#!/bin/bash
# Session Finite State Machine - Centralized state management with formal verification
#
# Implements the session FSM per TDD-session-state-machine.md and session-fsm.tla
# Provides:
#   - Lock Manager: Advisory locking with flock/mkdir fallback
#   - Schema Validator: Required field and status enum validation
#   - State Transition Engine: Enforces valid transitions per TLA+ spec
#   - Event Emitter: JSONL event logging for observability
#   - API Surface: fsm_get_state, fsm_transition, fsm_create_session
#
# TLA+ Invariants Enforced:
#   - TypeInvariant: status in {ACTIVE, PARKED, ARCHIVED}
#   - NoInvalidTransitions: All transitions obey ValidTransition predicate
#   - ArchivedIsTerminal: No transitions out of ARCHIVED
#   - MutualExclusion: At most one process holds exclusive lock
#
# Reference: docs/design/TDD-session-state-machine.md
# Reference: docs/specs/session-fsm.tla

set -euo pipefail

# =============================================================================
# Configuration (Dependency Injection Points)
# =============================================================================

# Override these for testing/isolation
export FSM_SESSIONS_DIR="${FSM_SESSIONS_DIR:-.claude/sessions}"
export FSM_LOCK_TIMEOUT="${FSM_LOCK_TIMEOUT:-10}"
export FSM_VALIDATE_SCHEMA="${FSM_VALIDATE_SCHEMA:-true}"
export FSM_EMIT_EVENTS="${FSM_EMIT_EVENTS:-true}"

# Internal state tracking for flock file descriptors
declare -A _FSM_LOCK_FDS 2>/dev/null || true

# =============================================================================
# Source Dependencies
# =============================================================================

# Get script directory for relative sourcing
_FSM_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source primitives if available (for get_yaml_field, atomic_write)
if [[ -f "$_FSM_SCRIPT_DIR/primitives.sh" ]]; then
    # shellcheck source=primitives.sh
    source "$_FSM_SCRIPT_DIR/primitives.sh"
fi

# =============================================================================
# Phase 1: Lock Manager
# =============================================================================
# Implements advisory locking with flock (preferred) and mkdir fallback (portable)
# TLA+ Properties: MutualExclusion, LockEventuallyGranted, NoDeadlock

# Internal: Get lock file path for a session
_fsm_lock_file() {
    local session_id="$1"
    echo "$FSM_SESSIONS_DIR/.locks/${session_id}.lock"
}

# Internal: Get mkdir lock marker path for a session
_fsm_lock_marker() {
    local session_id="$1"
    echo "$FSM_SESSIONS_DIR/.locks/${session_id}.lock.d"
}

# Acquire shared lock (for read operations)
# Multiple processes can hold shared locks simultaneously
# Usage: _fsm_lock_shared <session_id>
# Returns: 0 on success, 1 on timeout
_fsm_lock_shared() {
    local session_id="$1"
    local lock_file
    lock_file=$(_fsm_lock_file "$session_id")
    local timeout="${FSM_LOCK_TIMEOUT:-10}"

    # Ensure lock directory exists
    mkdir -p "$(dirname "$lock_file")" 2>/dev/null || return 1

    if command -v flock >/dev/null 2>&1; then
        # flock supports shared locks (-s)
        # Use automatic FD allocation (bash 4.1+) - FIXED (LOCK-003)
        local fd
        if exec {fd}>"$lock_file" 2>/dev/null; then
            if flock -s -w "$timeout" "$fd" 2>/dev/null; then
                # Store fd for later release
                _FSM_LOCK_FDS["$session_id"]="$fd"
                return 0
            else
                exec {fd}>&- 2>/dev/null || true
                return 1
            fi
        fi
        return 1
    else
        # Fallback: mkdir-based locking doesn't support shared mode
        # Treat shared as exclusive (conservative but correct)
        _fsm_lock_exclusive "$session_id"
    fi
}

# Acquire exclusive lock (for write operations)
# Only one process can hold an exclusive lock
# Usage: _fsm_lock_exclusive <session_id>
# Returns: 0 on success, 1 on timeout
_fsm_lock_exclusive() {
    local session_id="$1"
    local lock_file
    lock_file=$(_fsm_lock_file "$session_id")
    local lock_marker
    lock_marker=$(_fsm_lock_marker "$session_id")
    local timeout="${FSM_LOCK_TIMEOUT:-10}"

    # Ensure lock directory exists
    mkdir -p "$(dirname "$lock_file")" 2>/dev/null || return 1

    if command -v flock >/dev/null 2>&1; then
        # Use flock for exclusive lock (-x)
        # Use automatic FD allocation (bash 4.1+) - FIXED (LOCK-003)
        local fd
        if exec {fd}>"$lock_file" 2>/dev/null; then
            if flock -x -w "$timeout" "$fd" 2>/dev/null; then
                # Write PID to lock file for debugging
                echo "$$" >&"$fd"
                _FSM_LOCK_FDS["$session_id"]="$fd"
                return 0
            else
                exec {fd}>&- 2>/dev/null || true
                return 1
            fi
        fi
        return 1
    else
        # Fallback: mkdir-based locking (portable)
        local elapsed=0
        local sleep_interval=1

        while [[ "$elapsed" -lt "$timeout" ]]; do
            # mkdir is atomic - if it succeeds, we have the lock
            if mkdir "$lock_marker" 2>/dev/null; then
                # Store PID for stale lock detection
                echo "$$" > "$lock_marker/pid"
                return 0
            fi

            # Check if existing lock is stale (owner process dead)
            if [[ -f "$lock_marker/pid" ]]; then
                local owner_pid
                owner_pid=$(cat "$lock_marker/pid" 2>/dev/null || echo "")
                if [[ -n "$owner_pid" ]] && ! kill -0 "$owner_pid" 2>/dev/null; then
                    # Owner is dead, remove stale lock and retry
                    rm -rf "$lock_marker" 2>/dev/null
                    continue
                fi
            fi

            sleep "$sleep_interval"
            ((elapsed++))
        done

        return 1  # Timeout
    fi
}

# Release lock (both shared and exclusive)
# Usage: _fsm_unlock <session_id>
# Returns: 0 always (idempotent)
_fsm_unlock() {
    local session_id="$1"
    local lock_file
    lock_file=$(_fsm_lock_file "$session_id")
    local lock_marker
    lock_marker=$(_fsm_lock_marker "$session_id")

    # Release flock if we're using file descriptors
    if command -v flock >/dev/null 2>&1; then
        local fd="${_FSM_LOCK_FDS[$session_id]:-}"
        if [[ -n "$fd" ]]; then
            eval "exec $fd>&-" 2>/dev/null || true
            unset "_FSM_LOCK_FDS[$session_id]" 2>/dev/null || true
        fi
    fi

    # Remove mkdir-based lock marker if present
    rm -rf "$lock_marker" 2>/dev/null || true

    return 0
}

# =============================================================================
# Phase 2: Schema Validator
# =============================================================================
# Validates SESSION_CONTEXT.md against v2 schema requirements
# TLA+ Property: TypeInvariant

# Validate session context file against v2 schema
# Usage: _fsm_validate_context <ctx_file>
# Returns: 0 if valid, 1 if invalid (error details on stderr)
_fsm_validate_context() {
    local ctx_file="$1"

    # Skip validation if disabled
    if [[ "${FSM_VALIDATE_SCHEMA:-true}" != "true" ]]; then
        return 0
    fi

    # Check file exists
    if [[ ! -f "$ctx_file" ]]; then
        echo "Validation failed: File not found: $ctx_file" >&2
        return 1
    fi

    # Required fields for v2 schema
    local required_fields=("schema_version" "session_id" "status" "created_at"
                           "initiative" "complexity" "active_team" "current_phase")
    local missing=()

    for field in "${required_fields[@]}"; do
        if ! grep -q "^${field}:" "$ctx_file" 2>/dev/null; then
            missing+=("$field")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "Validation failed: Missing required fields: ${missing[*]}" >&2
        return 1
    fi

    # Validate status is a valid enum value
    local status
    status=$(grep -m1 "^status:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "' || echo "")
    case "$status" in
        ACTIVE|PARKED|ARCHIVED)
            ;;
        *)
            echo "Validation failed: Invalid status value: '$status' (must be ACTIVE, PARKED, or ARCHIVED)" >&2
            return 1
            ;;
    esac

    # Validate schema version
    local version
    version=$(grep -m1 "^schema_version:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "' || echo "")
    case "$version" in
        2.0|2.1) ;;  # Accept both versions
        *)
            echo "Validation failed: Unsupported schema version: '$version' (expected 2.0 or 2.1)" >&2
            return 1
            ;;
    esac

    return 0
}

# =============================================================================
# Phase 3: State Transition Engine
# =============================================================================
# Enforces valid state transitions per TLA+ ValidTransition predicate
# TLA+ Properties: NoInvalidTransitions, ArchivedIsTerminal

# Check if a state transition is valid per TLA+ spec
# Usage: _fsm_is_valid_transition <from_state> <to_state>
# Returns: 0 if valid, 1 if invalid
_fsm_is_valid_transition() {
    local from="$1"
    local to="$2"

    # TLA+ ValidTransition predicate:
    # \/ (from = "NONE" /\ to = "ACTIVE")      - Create session
    # \/ (from = "ACTIVE" /\ to = "PARKED")    - Park session
    # \/ (from = "ACTIVE" /\ to = "ARCHIVED")  - Complete/wrap session
    # \/ (from = "PARKED" /\ to = "ACTIVE")    - Resume session
    # \/ (from = "PARKED" /\ to = "ARCHIVED")  - Archive parked session

    case "$from:$to" in
        NONE:ACTIVE)      return 0 ;;
        ACTIVE:PARKED)    return 0 ;;
        ACTIVE:ARCHIVED)  return 0 ;;
        PARKED:ACTIVE)    return 0 ;;
        PARKED:ARCHIVED)  return 0 ;;
        *)                return 1 ;;
    esac
}

# Execute a state transition with atomic mutation
# Usage: _fsm_execute_transition <session_id> <from_state> <to_state> <metadata>
# Returns: 0 on success, 1 on failure
_fsm_execute_transition() {
    local session_id="$1"
    local from_state="$2"
    local to_state="$3"
    local metadata="${4:-{}}"

    local ctx_file="$FSM_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Update status field in frontmatter
    if grep -q "^status:" "$ctx_file" 2>/dev/null; then
        # Update existing status field
        sed -i.bak "s/^status:.*$/status: \"$to_state\"/" "$ctx_file"
        rm -f "${ctx_file}.bak"
    else
        # Add status field before closing ---
        awk -v status="$to_state" '
            /^---$/ && ++count == 2 {
                print "status: \"" status "\""
            }
            { print }
        ' "$ctx_file" > "${ctx_file}.tmp" && mv "${ctx_file}.tmp" "$ctx_file"
    fi

    # Add state-specific fields based on transition
    case "$to_state" in
        PARKED)
            # Extract reason from metadata if present
            local reason
            reason=$(echo "$metadata" | grep -o '"reason":"[^"]*"' 2>/dev/null | cut -d'"' -f4 || echo "Manual park")

            # Add parked_at timestamp
            awk -v ts="$timestamp" -v reason="$reason" '
                /^---$/ && ++count == 2 {
                    print "parked_at: \"" ts "\""
                    if (reason != "") print "parked_reason: \"" reason "\""
                }
                { print }
            ' "$ctx_file" > "${ctx_file}.tmp" && mv "${ctx_file}.tmp" "$ctx_file"
            ;;
        ARCHIVED)
            # Add archived_at timestamp
            awk -v ts="$timestamp" '
                /^---$/ && ++count == 2 {
                    print "archived_at: \"" ts "\""
                }
                { print }
            ' "$ctx_file" > "${ctx_file}.tmp" && mv "${ctx_file}.tmp" "$ctx_file"
            ;;
        ACTIVE)
            # Remove park fields if resuming from PARKED
            if [[ "$from_state" == "PARKED" ]]; then
                sed -i.bak \
                    -e '/^parked_at:/d' \
                    -e '/^parked_reason:/d' \
                    -e '/^parked_git_status:/d' \
                    -e '/^auto_parked_at:/d' \
                    -e '/^auto_parked_reason:/d' \
                    "$ctx_file" && rm -f "${ctx_file}.bak"

                # Add resumed_at
                awk -v ts="$timestamp" '
                    /^---$/ && ++count == 2 {
                        print "resumed_at: \"" ts "\""
                    }
                    { print }
                ' "$ctx_file" > "${ctx_file}.tmp" && mv "${ctx_file}.tmp" "$ctx_file"
            fi
            ;;
    esac

    return 0
}

# Safe mutation wrapper with backup and rollback
# Usage: _fsm_safe_mutate <session_id> <mutation_func> [args...]
# Returns: 0 on success, 1 on failure (with rollback)
_fsm_safe_mutate() {
    local session_id="$1"
    local mutation_func="$2"
    shift 2
    local args=("$@")

    local ctx_file="$FSM_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    local backup_file="${ctx_file}.backup.$$"

    # Create backup
    if [[ -f "$ctx_file" ]]; then
        cp "$ctx_file" "$backup_file" || {
            _fsm_emit_error "BACKUP_FAILED" "$session_id"
            return 1
        }
    fi

    # Execute mutation
    if ! "$mutation_func" "${args[@]}"; then
        # Rollback on failure
        if [[ -f "$backup_file" ]]; then
            mv "$backup_file" "$ctx_file"
        fi
        return 1
    fi

    # Validate result
    if ! _fsm_validate_context "$ctx_file"; then
        # Rollback on validation failure
        if [[ -f "$backup_file" ]]; then
            mv "$backup_file" "$ctx_file"
        fi
        _fsm_emit_error "VALIDATION_FAILED" "$session_id"
        return 1
    fi

    # Success - remove backup
    rm -f "$backup_file"
    return 0
}

# =============================================================================
# Phase 4: Event Emitter
# =============================================================================
# Publishes state change events to JSONL log for observability
# Event Types: SESSION_CREATED, SESSION_PARKED, SESSION_RESUMED, SESSION_ARCHIVED,
#              PHASE_TRANSITIONED, GUARD_VIOLATION

# Emit a state transition event
# Usage: _fsm_emit_event <session_id> <from_state> <to_state> [metadata]
_fsm_emit_event() {
    local session_id="$1"
    local from_state="$2"
    local to_state="$3"
    local metadata="${4:-{}}"

    # Skip if events disabled
    if [[ "${FSM_EMIT_EVENTS:-true}" != "true" ]]; then
        return 0
    fi

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Determine event type from transition
    local event_type
    case "$from_state:$to_state" in
        NONE:ACTIVE)    event_type="SESSION_CREATED" ;;
        ACTIVE:PARKED)  event_type="SESSION_PARKED" ;;
        PARKED:ACTIVE)  event_type="SESSION_RESUMED" ;;
        *:ARCHIVED)     event_type="SESSION_ARCHIVED" ;;
        *)              event_type="STATE_CHANGED" ;;
    esac

    # Build event JSON
    local event="{\"timestamp\":\"$timestamp\",\"event\":\"$event_type\",\"from\":\"$from_state\",\"to\":\"$to_state\""
    if [[ -n "$metadata" && "$metadata" != "{}" ]]; then
        event+=",\"metadata\":$metadata"
    fi
    event+="}"

    # Write to session-specific event log
    local events_file="$FSM_SESSIONS_DIR/$session_id/events.jsonl"
    mkdir -p "$(dirname "$events_file")" 2>/dev/null
    echo "$event" >> "$events_file"

    # Also write to global audit log
    local audit_dir="$FSM_SESSIONS_DIR/.audit"
    mkdir -p "$audit_dir" 2>/dev/null
    echo "$timestamp | $session_id | $event_type | $from_state -> $to_state" >> "$audit_dir/transitions.log"
}

# Emit an error event
# Usage: _fsm_emit_error <error_type> <session_id> [details...]
_fsm_emit_error() {
    local error_type="$1"
    local session_id="$2"
    shift 2
    local details="$*"

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Write structured error to stderr
    cat >&2 <<EOF
{
  "error": "$error_type",
  "session_id": "$session_id",
  "timestamp": "$timestamp",
  "details": "$details"
}
EOF

    # Log to audit trail
    local audit_dir="$FSM_SESSIONS_DIR/.audit"
    mkdir -p "$audit_dir" 2>/dev/null
    echo "$timestamp | ERROR | $session_id | $error_type | $details" >> "$audit_dir/errors.log"
}

# =============================================================================
# Phase 5: API Surface
# =============================================================================
# Public API: fsm_get_state, fsm_transition, fsm_create_session

# Get current state of a session
# Usage: fsm_get_state <session_id>
# Returns: NONE | ACTIVE | PARKED | ARCHIVED
# Exit code: 0 on success, 1 on error
fsm_get_state() {
    local session_id="$1"

    if [[ -z "$session_id" ]]; then
        echo "NONE"
        return 0
    fi

    local ctx_file="$FSM_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    # Acquire shared lock for consistent read
    _fsm_lock_shared "$session_id" || {
        echo "NONE"
        return 1
    }

    # Check if session exists
    if [[ ! -f "$ctx_file" ]]; then
        _fsm_unlock "$session_id"
        echo "NONE"
        return 0
    fi

    # Read status field (v2 canonical source of truth)
    local status
    status=$(grep -m1 "^status:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')

    # Fallback for v1 sessions: infer from parked_at/completed_at
    if [[ -z "$status" ]]; then
        if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
            status="PARKED"
        elif grep -q "^completed_at:" "$ctx_file" 2>/dev/null; then
            status="ARCHIVED"
        else
            status="ACTIVE"
        fi
    fi

    _fsm_unlock "$session_id"
    echo "${status:-ACTIVE}"
}

# Internal: Get state without locking (for use when lock already held)
_fsm_get_state_unlocked() {
    local session_id="$1"
    local ctx_file="$FSM_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    if [[ ! -f "$ctx_file" ]]; then
        echo "NONE"
        return 0
    fi

    local status
    status=$(grep -m1 "^status:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')

    # Fallback for v1 sessions
    if [[ -z "$status" ]]; then
        if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
            status="PARKED"
        elif grep -q "^completed_at:" "$ctx_file" 2>/dev/null; then
            status="ARCHIVED"
        else
            status="ACTIVE"
        fi
    fi

    echo "${status:-ACTIVE}"
}

# Execute a state transition
# Usage: fsm_transition <session_id> <target_state> [metadata_json]
# Returns: JSON result object
# Exit code: 0 on success, 1 on invalid transition, 2 on lock failure
fsm_transition() {
    local session_id="$1"
    local target_state="$2"
    local metadata="${3:-{}}"

    if [[ -z "$session_id" || -z "$target_state" ]]; then
        echo '{"success": false, "error": "session_id and target_state required"}'
        return 1
    fi

    local ctx_file="$FSM_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    # Acquire exclusive lock
    if ! _fsm_lock_exclusive "$session_id"; then
        _fsm_emit_error "LOCK_TIMEOUT" "$session_id" "$target_state"
        echo '{"success": false, "error": "LOCK_TIMEOUT"}'
        return 2
    fi

    # Get current state (without lock since we already hold it)
    local current_state
    current_state=$(_fsm_get_state_unlocked "$session_id")

    # Validate transition
    if ! _fsm_is_valid_transition "$current_state" "$target_state"; then
        _fsm_unlock "$session_id"
        _fsm_emit_error "INVALID_TRANSITION" "$session_id" "$current_state" "$target_state"
        echo "{\"success\": false, \"error\": \"INVALID_TRANSITION\", \"from\": \"$current_state\", \"to\": \"$target_state\"}"
        return 1
    fi

    # Create backup
    local backup_file="${ctx_file}.backup.$$"
    if [[ -f "$ctx_file" ]]; then
        cp "$ctx_file" "$backup_file"
    fi

    # Execute transition
    if ! _fsm_execute_transition "$session_id" "$current_state" "$target_state" "$metadata"; then
        # Rollback
        if [[ -f "$backup_file" ]]; then
            mv "$backup_file" "$ctx_file"
        fi
        _fsm_unlock "$session_id"
        echo '{"success": false, "error": "TRANSITION_FAILED"}'
        return 1
    fi

    # Validate result (only for v2 sessions)
    local schema_version
    schema_version=$(grep -m1 "^schema_version:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "' || echo "")
    if [[ "$schema_version" == "2.0" || "$schema_version" == "2.1" ]]; then
        if ! _fsm_validate_context "$ctx_file"; then
            # Rollback
            if [[ -f "$backup_file" ]]; then
                mv "$backup_file" "$ctx_file"
            fi
            _fsm_unlock "$session_id"
            _fsm_emit_error "VALIDATION_FAILED" "$session_id"
            echo '{"success": false, "error": "VALIDATION_FAILED"}'
            return 1
        fi
    fi

    # Emit event
    _fsm_emit_event "$session_id" "$current_state" "$target_state" "$metadata"

    # Cleanup
    rm -f "$backup_file"
    _fsm_unlock "$session_id"

    echo "{\"success\": true, \"from\": \"$current_state\", \"to\": \"$target_state\"}"
}

# Create a new session (NONE -> ACTIVE transition)
# Usage: fsm_create_session <initiative> <complexity> [team]
# Returns: session_id on success
# Exit code: 0 on success, 1 on failure
# Note: If team is empty/none/null, creates a cross-cutting session
fsm_create_session() {
    local initiative="$1"
    local complexity="$2"
    local team="${3:-}"

    # If team not specified, check ACTIVE_RITE file
    if [[ -z "$team" ]]; then
        team=$(cat ".claude/ACTIVE_RITE" 2>/dev/null || echo "")
    fi

    # Normalize empty/none to explicit marker for cross-cutting
    if [[ -z "$team" || "$team" == "none" ]]; then
        team="none"
    fi

    if [[ -z "$initiative" || -z "$complexity" ]]; then
        echo '{"success": false, "error": "initiative and complexity required"}' >&2
        return 1
    fi

    # Generate unique session ID
    local session_id
    session_id="session-$(date +%Y%m%d-%H%M%S)-$(openssl rand -hex 4 2>/dev/null || head -c 8 /dev/urandom | xxd -p)"

    local session_dir="$FSM_SESSIONS_DIR/$session_id"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"

    # Create session directory
    mkdir -p "$session_dir" || {
        echo '{"success": false, "error": "Failed to create session directory"}' >&2
        return 1
    }

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Compute team field value (null for cross-cutting, quoted string otherwise)
    local team_field_value
    if [[ "$team" == "none" ]]; then
        team_field_value="null"
    else
        team_field_value="\"$team\""
    fi

    # Create SESSION_CONTEXT.md with v2.1 schema
    cat > "$ctx_file" <<CONTEXT
---
schema_version: "2.1"
session_id: "$session_id"
status: "ACTIVE"
created_at: "$timestamp"
initiative: "$initiative"
complexity: "$complexity"
active_team: "$team"
team: $team_field_value
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

    # Validate the created session
    if ! _fsm_validate_context "$ctx_file"; then
        rm -rf "$session_dir"
        echo '{"success": false, "error": "Failed to validate SESSION_CONTEXT.md"}' >&2
        return 1
    fi

    # Set as current session (if set_current_session is available)
    if declare -F set_current_session >/dev/null 2>&1; then
        set_current_session "$session_id" 2>/dev/null || true
    fi

    # Emit creation event
    _fsm_emit_event "$session_id" "NONE" "ACTIVE" "{\"initiative\":\"$initiative\"}"

    echo "$session_id"
}

# =============================================================================
# Backward Compatibility Layer
# =============================================================================
# These functions provide compatibility with existing code

# Check if session is in v2 schema format
_fsm_is_v2_session() {
    local session_id="$1"
    local ctx_file="$FSM_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    if [[ ! -f "$ctx_file" ]]; then
        return 1
    fi

    local version
    version=$(grep -m1 "^schema_version:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')
    [[ "$version" == "2.0" ]]
}

# Get session state with backward compatibility for v1 sessions
# For v1 sessions, infers state from parked_at/completed_at fields
# For v2 sessions, reads status field directly
fsm_get_state_compat() {
    local session_id="${1:-}"

    if [[ -z "$session_id" ]]; then
        # Try to get current session
        if declare -F get_session_id >/dev/null 2>&1; then
            session_id=$(get_session_id)
        fi
    fi

    if [[ -z "$session_id" ]]; then
        echo "NONE"
        return 0
    fi

    # Use FSM for state lookup
    fsm_get_state "$session_id"
}

# =============================================================================
# CLI Interface (for testing and direct invocation)
# =============================================================================

_fsm_main() {
    local cmd="${1:-help}"
    shift || true

    case "$cmd" in
        get-state)
            fsm_get_state "$@"
            ;;
        transition)
            fsm_transition "$@"
            ;;
        create)
            fsm_create_session "$@"
            ;;
        is-valid-transition)
            if _fsm_is_valid_transition "$1" "$2"; then
                echo "true"
            else
                echo "false"
                return 1
            fi
            ;;
        validate)
            if _fsm_validate_context "$1"; then
                echo '{"valid": true}'
            else
                echo '{"valid": false}'
                return 1
            fi
            ;;
        help|--help|-h)
            cat <<EOF
session-fsm.sh - Session Finite State Machine

Commands:
  get-state <session_id>              Get current session state
  transition <session_id> <state> [metadata]  Execute state transition
  create <initiative> <complexity> [team]     Create new session
  is-valid-transition <from> <to>     Check if transition is valid
  validate <ctx_file>                 Validate SESSION_CONTEXT.md

Environment Variables:
  FSM_SESSIONS_DIR     Session storage directory (default: .claude/sessions)
  FSM_LOCK_TIMEOUT     Lock acquisition timeout in seconds (default: 10)
  FSM_VALIDATE_SCHEMA  Enable schema validation (default: true)
  FSM_EMIT_EVENTS      Enable event emission (default: true)

States:
  NONE      - No session exists
  ACTIVE    - Session in progress
  PARKED    - Session suspended
  ARCHIVED  - Session complete (terminal)

Valid Transitions:
  NONE -> ACTIVE (create)
  ACTIVE -> PARKED (park)
  ACTIVE -> ARCHIVED (wrap)
  PARKED -> ACTIVE (resume)
  PARKED -> ARCHIVED (archive)
EOF
            ;;
        *)
            echo "Unknown command: $cmd" >&2
            echo "Use '$0 help' for usage information" >&2
            return 1
            ;;
    esac
}

# Run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    _fsm_main "$@"
fi

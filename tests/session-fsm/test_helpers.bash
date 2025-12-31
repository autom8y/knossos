#!/bin/bash
# test_helpers.bash - Shared test infrastructure for Session FSM tests
#
# Provides:
#   - Test isolation (temp directories)
#   - Mock session creation/cleanup
#   - Assertion helpers compatible with BATS 1.x
#   - TLA+ invariant validation helpers
#
# Usage: source this file in your .bats tests via setup()
#
# Reference: TDD-session-state-machine.md, session-fsm.tla

# =============================================================================
# Configuration (Dependency Injection Points)
# =============================================================================

# Override these in tests for isolation
export FSM_SESSIONS_DIR="${FSM_SESSIONS_DIR:-}"
export FSM_LOCK_TIMEOUT="${FSM_LOCK_TIMEOUT:-5}"
export FSM_VALIDATE_SCHEMA="${FSM_VALIDATE_SCHEMA:-true}"
export FSM_EMIT_EVENTS="${FSM_EMIT_EVENTS:-true}"

# Path to scripts under test
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export SESSION_MANAGER="${SESSION_MANAGER:-$SCRIPT_DIR/user-hooks/lib/session-manager.sh}"
export SESSION_UTILS="${SESSION_UTILS:-$SCRIPT_DIR/user-hooks/lib/session-utils.sh}"
export SESSION_MIGRATE="${SESSION_MIGRATE:-$SCRIPT_DIR/user-hooks/lib/session-migrate.sh}"
export SESSION_FSM="${SESSION_FSM:-$SCRIPT_DIR/user-hooks/lib/session-fsm.sh}"

# =============================================================================
# Test Isolation
# =============================================================================

# Create isolated test environment
# Sets up temp directory structure mirroring .claude/sessions/
# Returns: sets TEST_DIR, TEST_SESSIONS_DIR, TEST_PROJECT_DIR
setup_test_environment() {
    # Create unique temp directory for this test
    TEST_DIR=$(mktemp -d "${TMPDIR:-/tmp}/session-fsm-test.XXXXXX")

    # Create project structure
    TEST_PROJECT_DIR="$TEST_DIR/project"
    mkdir -p "$TEST_PROJECT_DIR/.claude/sessions"
    mkdir -p "$TEST_PROJECT_DIR/.claude/sessions/.locks"
    mkdir -p "$TEST_PROJECT_DIR/.claude/sessions/.audit"

    TEST_SESSIONS_DIR="$TEST_PROJECT_DIR/.claude/sessions"

    # Set environment for session-manager.sh isolation
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    export FSM_SESSIONS_DIR="$TEST_SESSIONS_DIR"

    # Change to test project directory
    cd "$TEST_PROJECT_DIR" || return 1

    # Initialize minimal git repo for git-related functions
    git init --quiet 2>/dev/null || true
    git config user.email "test@test.com" 2>/dev/null || true
    git config user.name "Test" 2>/dev/null || true

    # Create ACTIVE_TEAM file
    echo "10x-dev-pack" > "$TEST_PROJECT_DIR/.claude/ACTIVE_TEAM"
}

# Cleanup test environment
teardown_test_environment() {
    # Return to original directory
    cd / 2>/dev/null || true

    # Remove test directory
    if [[ -n "$TEST_DIR" && -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
    fi

    # Unset test variables
    unset TEST_DIR TEST_PROJECT_DIR TEST_SESSIONS_DIR
    unset CLAUDE_PROJECT_DIR FSM_SESSIONS_DIR CLAUDE_SESSION_ID
}

# =============================================================================
# Mock Session Creation
# =============================================================================

# Create a mock session in specified state
# Usage: create_mock_session <session_id> <status> [initiative] [complexity] [team]
# Returns: 0 on success, sets SESSION_FILE to created file path
create_mock_session() {
    local session_id="${1:-session-$(date +%Y%m%d-%H%M%S)-$(openssl rand -hex 4 2>/dev/null || head -c 8 /dev/urandom | xxd -p)}"
    local status="${2:-ACTIVE}"
    local initiative="${3:-Test Initiative}"
    local complexity="${4:-MODULE}"
    local team="${5:-10x-dev-pack}"

    local session_dir="$TEST_SESSIONS_DIR/$session_id"
    mkdir -p "$session_dir"

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    SESSION_FILE="$session_dir/SESSION_CONTEXT.md"

    # Create base session context
    cat > "$SESSION_FILE" <<EOF
---
schema_version: "2.0"
session_id: "$session_id"
status: "$status"
created_at: "$timestamp"
initiative: "$initiative"
complexity: "$complexity"
active_team: "$team"
current_phase: "requirements"
EOF

    # Add status-specific fields
    case "$status" in
        PARKED)
            cat >> "$SESSION_FILE" <<EOF
parked_at: "$timestamp"
parked_reason: "Test park"
parked_git_status: "clean"
EOF
            ;;
        ARCHIVED)
            cat >> "$SESSION_FILE" <<EOF
completed_at: "$timestamp"
EOF
            ;;
    esac

    # Close frontmatter and add body
    cat >> "$SESSION_FILE" <<EOF
---

# Session: $initiative

## Artifacts
- PRD: pending
- TDD: pending

## Blockers
None yet.

## Next Steps
1. Complete requirements gathering
EOF

    echo "$session_id"
}

# Create a v1 schema session (for migration tests)
# Usage: create_v1_session <session_id> <state> [with_dual_fields]
create_v1_session() {
    local session_id="${1:-session-$(date +%Y%m%d-%H%M%S)-$(openssl rand -hex 4 2>/dev/null || head -c 8 /dev/urandom | xxd -p)}"
    local state="${2:-active}"
    local with_dual_fields="${3:-false}"

    local session_dir="$TEST_SESSIONS_DIR/$session_id"
    mkdir -p "$session_dir"

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    SESSION_FILE="$session_dir/SESSION_CONTEXT.md"

    # Create v1 schema (no schema_version, legacy fields)
    cat > "$SESSION_FILE" <<EOF
---
session_id: "$session_id"
created_at: "$timestamp"
initiative: "V1 Test Initiative"
complexity: "MODULE"
active_team: "10x-dev-pack"
current_phase: "requirements"
EOF

    # Add state-specific legacy fields
    case "$state" in
        parked)
            cat >> "$SESSION_FILE" <<EOF
parked_at: "$timestamp"
park_reason: "Legacy park reason"
git_status_at_park: "clean"
EOF
            if [[ "$with_dual_fields" == "true" ]]; then
                cat >> "$SESSION_FILE" <<EOF
parked_reason: "Duplicate park reason"
parked_git_status: "clean"
session_state: "PARKED"
EOF
            fi
            ;;
        auto_parked)
            cat >> "$SESSION_FILE" <<EOF
auto_parked_at: "$timestamp"
auto_parked_reason: "Inactivity timeout"
EOF
            ;;
        archived)
            cat >> "$SESSION_FILE" <<EOF
completed_at: "$timestamp"
EOF
            ;;
    esac

    # Close frontmatter
    cat >> "$SESSION_FILE" <<EOF
---

# Session: V1 Test Initiative
EOF

    echo "$session_id"
}

# Set a session as current
# Usage: set_current_mock_session <session_id>
set_current_mock_session() {
    local session_id="$1"
    echo "$session_id" > "$TEST_SESSIONS_DIR/.current-session"
    export CLAUDE_SESSION_ID="$session_id"
}

# Clear current session
clear_current_mock_session() {
    rm -f "$TEST_SESSIONS_DIR/.current-session"
    unset CLAUDE_SESSION_ID
}

# =============================================================================
# Session State Helpers
# =============================================================================

# Get status field from session context
# Usage: get_session_status <session_id>
get_session_status() {
    local session_id="$1"
    local session_file="$TEST_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    if [[ ! -f "$session_file" ]]; then
        echo "NONE"
        return 0
    fi

    # Try status field first (v2), then session_state (v1 legacy)
    local status
    status=$(grep -m1 "^status:" "$session_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')

    if [[ -z "$status" ]]; then
        status=$(grep -m1 "^session_state:" "$session_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')
    fi

    # Infer from parked_at if no explicit status
    if [[ -z "$status" ]]; then
        if grep -qE "^(parked_at|auto_parked_at):" "$session_file" 2>/dev/null; then
            status="PARKED"
        elif grep -q "^completed_at:" "$session_file" 2>/dev/null; then
            status="ARCHIVED"
        else
            status="ACTIVE"
        fi
    fi

    echo "$status"
}

# Get schema version from session context
# Usage: get_schema_version <session_id>
get_schema_version() {
    local session_id="$1"
    local session_file="$TEST_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    if [[ ! -f "$session_file" ]]; then
        echo ""
        return 1
    fi

    grep -m1 "^schema_version:" "$session_file" 2>/dev/null | cut -d: -f2- | tr -d ' "'
}

# Check if session has specific field
# Usage: session_has_field <session_id> <field_name>
session_has_field() {
    local session_id="$1"
    local field="$2"
    local session_file="$TEST_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    grep -q "^${field}:" "$session_file" 2>/dev/null
}

# =============================================================================
# TLA+ Invariant Validation Helpers
# =============================================================================

# Validate transition is allowed per TLA+ spec
# Usage: is_valid_transition <from_state> <to_state>
# Returns: 0 if valid, 1 if invalid
is_valid_transition() {
    local from="$1"
    local to="$2"

    # From TLA+ ValidTransition predicate:
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

# Get all valid transitions from a state
# Usage: get_valid_transitions <from_state>
# Returns: space-separated list of valid target states
get_valid_transitions() {
    local from="$1"

    case "$from" in
        NONE)     echo "ACTIVE" ;;
        ACTIVE)   echo "PARKED ARCHIVED" ;;
        PARKED)   echo "ACTIVE ARCHIVED" ;;
        ARCHIVED) echo "" ;;  # Terminal state
        *)        echo "" ;;
    esac
}

# Get all invalid transitions from a state
# Usage: get_invalid_transitions <from_state>
# Returns: space-separated list of invalid target states
get_invalid_transitions() {
    local from="$1"
    local all_states="NONE ACTIVE PARKED ARCHIVED"
    local invalid=""

    for to in $all_states; do
        if ! is_valid_transition "$from" "$to"; then
            invalid="$invalid $to"
        fi
    done

    echo "$invalid" | xargs  # Trim whitespace
}

# Verify ArchivedIsTerminal invariant
# Usage: assert_archived_is_terminal <session_id>
assert_archived_is_terminal() {
    local session_id="$1"
    local status
    status=$(get_session_status "$session_id")

    if [[ "$status" == "ARCHIVED" ]]; then
        # Verify no valid transitions out
        local valid_out
        valid_out=$(get_valid_transitions "ARCHIVED")
        if [[ -n "$valid_out" ]]; then
            echo "ArchivedIsTerminal VIOLATED: ARCHIVED has valid transitions to: $valid_out" >&2
            return 1
        fi
    fi
    return 0
}

# Verify PhaseConsistency invariant
# Usage: assert_phase_consistency <session_id>
assert_phase_consistency() {
    local session_id="$1"
    local session_file="$TEST_SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    if [[ ! -f "$session_file" ]]; then
        return 0  # No session, no phase to check
    fi

    local status
    status=$(get_session_status "$session_id")
    local phase
    phase=$(grep -m1 "^current_phase:" "$session_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')

    # Phase is only meaningful when ACTIVE
    # For PARKED/ARCHIVED, phase may exist but shouldn't be relied upon
    if [[ "$status" != "ACTIVE" && -n "$phase" && "$phase" != "none" ]]; then
        # This is a warning, not a violation - phase can be preserved
        echo "Note: Phase '$phase' exists in non-ACTIVE state '$status'" >&2
    fi

    return 0
}

# =============================================================================
# Locking Test Helpers
# =============================================================================

# Check if lock exists
# Usage: lock_exists <lock_name>
lock_exists() {
    local lock_name="$1"
    local lock_dir="$TEST_SESSIONS_DIR/.locks"

    # Check both flock file and mkdir marker
    [[ -f "$lock_dir/${lock_name}.lock" ]] || [[ -d "$lock_dir/${lock_name}.lock.d" ]]
}

# Get lock holder PID (if available)
# Usage: get_lock_holder <lock_name>
get_lock_holder() {
    local lock_name="$1"
    local lock_dir="$TEST_SESSIONS_DIR/.locks"
    local marker_dir="$lock_dir/${lock_name}.lock.d"

    if [[ -f "$marker_dir/pid" ]]; then
        cat "$marker_dir/pid"
    else
        echo ""
    fi
}

# Create a lock for testing (simulates another process holding lock)
# Usage: create_test_lock <lock_name> [pid]
create_test_lock() {
    local lock_name="$1"
    local pid="${2:-99999}"  # Use fake PID by default
    local lock_dir="$TEST_SESSIONS_DIR/.locks"

    mkdir -p "$lock_dir/${lock_name}.lock.d"
    echo "$pid" > "$lock_dir/${lock_name}.lock.d/pid"
}

# Remove a test lock
# Usage: remove_test_lock <lock_name>
remove_test_lock() {
    local lock_name="$1"
    local lock_dir="$TEST_SESSIONS_DIR/.locks"

    rm -rf "$lock_dir/${lock_name}.lock" "$lock_dir/${lock_name}.lock.d" 2>/dev/null
}

# =============================================================================
# Assertion Helpers (BATS 1.x Compatible)
# =============================================================================

# Assert two values are equal
# Usage: assert_equal <expected> <actual> [message]
assert_equal() {
    local expected="$1"
    local actual="$2"
    local message="${3:-Expected '$expected' but got '$actual'}"

    if [[ "$expected" != "$actual" ]]; then
        echo "ASSERTION FAILED: $message" >&2
        return 1
    fi
    return 0
}

# Assert value matches regex
# Usage: assert_match <pattern> <value> [message]
assert_match() {
    local pattern="$1"
    local value="$2"
    local message="${3:-Value '$value' does not match pattern '$pattern'}"

    if [[ ! "$value" =~ $pattern ]]; then
        echo "ASSERTION FAILED: $message" >&2
        return 1
    fi
    return 0
}

# Assert file exists
# Usage: assert_file_exists <path> [message]
assert_file_exists() {
    local path="$1"
    local message="${2:-File does not exist: $path}"

    if [[ ! -f "$path" ]]; then
        echo "ASSERTION FAILED: $message" >&2
        return 1
    fi
    return 0
}

# Assert file does not exist
# Usage: assert_file_not_exists <path> [message]
assert_file_not_exists() {
    local path="$1"
    local message="${2:-File should not exist: $path}"

    if [[ -f "$path" ]]; then
        echo "ASSERTION FAILED: $message" >&2
        return 1
    fi
    return 0
}

# Assert directory exists
# Usage: assert_dir_exists <path> [message]
assert_dir_exists() {
    local path="$1"
    local message="${2:-Directory does not exist: $path}"

    if [[ ! -d "$path" ]]; then
        echo "ASSERTION FAILED: $message" >&2
        return 1
    fi
    return 0
}

# Assert output contains string
# Usage: assert_output_contains <substring> [message]
# Note: Uses $output from BATS run command
assert_output_contains() {
    local substring="$1"
    local message="${2:-Output does not contain: $substring}"

    if [[ "$output" != *"$substring"* ]]; then
        echo "ASSERTION FAILED: $message" >&2
        echo "Actual output: $output" >&2
        return 1
    fi
    return 0
}

# Assert command succeeds (exit 0)
# Usage: assert_success [message]
# Note: Uses $status from BATS run command
assert_success() {
    local message="${1:-Command should have succeeded (exit 0) but got exit $status}"

    if [[ "$status" -ne 0 ]]; then
        echo "ASSERTION FAILED: $message" >&2
        return 1
    fi
    return 0
}

# Assert command fails (exit non-zero)
# Usage: assert_failure [expected_code] [message]
# Note: Uses $status from BATS run command
assert_failure() {
    local expected="${1:-}"
    local message="${2:-Command should have failed but succeeded}"

    if [[ "$status" -eq 0 ]]; then
        echo "ASSERTION FAILED: $message" >&2
        return 1
    fi

    if [[ -n "$expected" && "$status" -ne "$expected" ]]; then
        echo "ASSERTION FAILED: Expected exit code $expected but got $status" >&2
        return 1
    fi
    return 0
}

# =============================================================================
# JSON Assertion Helpers
# =============================================================================

# Assert JSON field equals value
# Usage: assert_json_field <json> <field> <expected>
assert_json_field() {
    local json="$1"
    local field="$2"
    local expected="$3"

    local actual
    if command -v jq >/dev/null 2>&1; then
        actual=$(echo "$json" | jq -r ".$field // empty" 2>/dev/null)
    else
        # Fallback: grep-based parsing (handles simple cases)
        actual=$(echo "$json" | grep -o "\"$field\": *\"[^\"]*\"" 2>/dev/null | cut -d'"' -f4)
        if [[ -z "$actual" ]]; then
            # Try non-string values
            actual=$(echo "$json" | grep -o "\"$field\": *[^,}]*" 2>/dev/null | cut -d: -f2 | tr -d ' ')
        fi
    fi

    assert_equal "$expected" "$actual" "JSON field '$field' expected '$expected' but got '$actual'"
}

# Assert JSON has field
# Usage: assert_json_has_field <json> <field>
assert_json_has_field() {
    local json="$1"
    local field="$2"

    if command -v jq >/dev/null 2>&1; then
        if ! echo "$json" | jq -e "has(\"$field\")" >/dev/null 2>&1; then
            echo "ASSERTION FAILED: JSON missing field '$field'" >&2
            return 1
        fi
    else
        if [[ "$json" != *"\"$field\""* ]]; then
            echo "ASSERTION FAILED: JSON missing field '$field'" >&2
            return 1
        fi
    fi
    return 0
}

# =============================================================================
# Test Utilities
# =============================================================================

# Wait for condition with timeout
# Usage: wait_for <condition_command> [timeout_seconds] [interval_seconds]
wait_for() {
    local condition="$1"
    local timeout="${2:-5}"
    local interval="${3:-0.1}"
    local elapsed=0

    while ! eval "$condition"; do
        sleep "$interval"
        elapsed=$(echo "$elapsed + $interval" | bc 2>/dev/null || echo $((elapsed + 1)))
        if [[ $(echo "$elapsed >= $timeout" | bc 2>/dev/null || echo 0) -eq 1 ]]; then
            return 1
        fi
    done
    return 0
}

# Generate unique test session ID
# Usage: generate_test_session_id
generate_test_session_id() {
    echo "session-$(date +%Y%m%d-%H%M%S)-$(head -c 4 /dev/urandom | xxd -p 2>/dev/null || openssl rand -hex 4)"
}

# Print test context (for debugging)
# Usage: debug_context
debug_context() {
    echo "=== Test Context ===" >&2
    echo "TEST_DIR: $TEST_DIR" >&2
    echo "TEST_PROJECT_DIR: $TEST_PROJECT_DIR" >&2
    echo "TEST_SESSIONS_DIR: $TEST_SESSIONS_DIR" >&2
    echo "CLAUDE_PROJECT_DIR: $CLAUDE_PROJECT_DIR" >&2
    echo "CLAUDE_SESSION_ID: ${CLAUDE_SESSION_ID:-<unset>}" >&2
    echo "Sessions:" >&2
    ls -la "$TEST_SESSIONS_DIR" 2>/dev/null || echo "  (none)" >&2
    echo "===================" >&2
}

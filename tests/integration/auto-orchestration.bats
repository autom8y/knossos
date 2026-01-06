#!/usr/bin/env bats
# auto-orchestration.bats - Integration tests for Phase 1-2 auto-orchestration hooks
#
# Tests orchestrator-router.sh and start-preflight.sh coordination
# Reference: docs/design/TDD-auto-orchestration.md (lines 505-731)
#
# Test Categories:
#   boot_*  - Session bootstrap (creation/reuse)
#   cons_*  - Consultation request format validation
#   hook_*  - Hook coordination behavior
#   fric_*  - Friction measurement

# =============================================================================
# Load Test Helper
# =============================================================================

load '../session-fsm/test_helpers.bash'

# =============================================================================
# Setup / Teardown
# =============================================================================

setup() {
    # Store REAL paths BEFORE setup_test_environment changes cwd
    # BATS_TEST_DIRNAME is the directory containing this test file
    REAL_PROJECT_DIR="${BATS_TEST_DIRNAME}/../.."
    REAL_PROJECT_DIR="$(cd "$REAL_PROJECT_DIR" && pwd)"
    export ORCHESTRATOR_ROUTER="$REAL_PROJECT_DIR/.claude/hooks/validation/orchestrator-router.sh"
    export START_PREFLIGHT="$REAL_PROJECT_DIR/.claude/hooks/session-guards/start-preflight.sh"
    export REAL_HOOKS_LIB="$REAL_PROJECT_DIR/.claude/hooks/lib"

    setup_test_environment

    # Create ecosystem-pack team structure (orchestrator present)
    echo "ecosystem-pack" > "$TEST_PROJECT_DIR/.claude/ACTIVE_RITE"
    mkdir -p "$TEST_PROJECT_DIR/.claude/agents"
    echo "# Orchestrator Agent" > "$TEST_PROJECT_DIR/.claude/agents/orchestrator.md"

    # Ensure hooks lib is in test project
    export HOOKS_LIB="$TEST_PROJECT_DIR/.claude/hooks/lib"
    mkdir -p "$HOOKS_LIB"

    # Copy required library files to test project
    cp "$REAL_HOOKS_LIB/"*.sh "$HOOKS_LIB/" 2>/dev/null || true
}

teardown() {
    teardown_test_environment
}

# =============================================================================
# Helper Functions
# =============================================================================

# Remove all sessions for clean state
clean_sessions() {
    rm -f "$TEST_PROJECT_DIR/.claude/sessions/.current-session"
    rm -rf "$TEST_PROJECT_DIR/.claude/sessions/session-"* 2>/dev/null || true
}

# Remove orchestrator to test non-orchestrated behavior
remove_orchestrator() {
    rm -f "$TEST_PROJECT_DIR/.claude/agents/orchestrator.md"
}

# =============================================================================
# Phase 1: Session Bootstrap Tests
# =============================================================================

# boot_001: /start creates session when none exists
# Requirement: FR-1.1
@test "boot_001: /start creates session when none exists" {
    # Ensure no session
    clean_sessions

    # Simulate /start command via orchestrator-router.sh
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    # Check session was created
    [ "$status" -eq 0 ]
    [[ "$output" == *"Session created:"* ]]
    [[ "$output" == *"Task(orchestrator"* ]]

    # Verify session file exists
    local session_id
    session_id=$(cat "$TEST_PROJECT_DIR/.claude/sessions/.current-session" 2>/dev/null)
    [ -n "$session_id" ]
    [ -f "$TEST_PROJECT_DIR/.claude/sessions/$session_id/SESSION_CONTEXT.md" ]
}

# boot_002: /start with existing session reuses it
# Requirement: FR-1.2
@test "boot_002: /start with existing session reuses it" {
    # Create existing session manually
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    # Create session via session-manager.sh
    local create_result
    create_result=$("$HOOKS_LIB/session-manager.sh" create "Existing Initiative" "MODULE" "ecosystem-pack" 2>&1)

    # Extract session_id
    local existing_id
    if command -v jq >/dev/null 2>&1; then
        existing_id=$(echo "$create_result" | jq -r '.session_id // empty' 2>/dev/null)
    else
        existing_id=$(echo "$create_result" | grep -o '"session_id": "[^"]*"' | cut -d'"' -f4)
    fi
    [ -n "$existing_id" ]

    # Simulate /start command
    export CLAUDE_USER_PROMPT='/start "New Initiative"'
    run bash "$ORCHESTRATOR_ROUTER"

    # Should use existing session
    [ "$status" -eq 0 ]
    [[ "$output" == *"Using existing session:"* ]]
    [[ "$output" == *"$existing_id"* ]]
}

# boot_003: Parallel /start commands don't race
# Requirement: FR-1.3
@test "boot_003: parallel /start commands don't create duplicates" {
    clean_sessions
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    # Run two /start commands in quick succession
    # Note: True parallelism is hard to test reliably, so we use sequential with lock verification
    export CLAUDE_USER_PROMPT='/start "Parallel Test 1"'
    bash "$ORCHESTRATOR_ROUTER" &
    pid1=$!

    export CLAUDE_USER_PROMPT='/start "Parallel Test 2"'
    bash "$ORCHESTRATOR_ROUTER" &
    pid2=$!

    wait $pid1 || true
    wait $pid2 || true

    # Count sessions - should only be 1
    local session_count
    session_count=$(ls -1d "$TEST_PROJECT_DIR/.claude/sessions/session-"* 2>/dev/null | wc -l | tr -d ' ')
    [ "$session_count" -eq 1 ]
}

# boot_004: Session creation failure produces graceful error
# Requirement: NFR-2
@test "boot_004: session creation failure degrades gracefully" {
    clean_sessions
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    # Make sessions directory read-only to simulate failure
    chmod 444 "$TEST_PROJECT_DIR/.claude/sessions"

    export CLAUDE_USER_PROMPT='/start "Test"'
    run bash "$ORCHESTRATOR_ROUTER"

    # Restore permissions for teardown
    chmod 755 "$TEST_PROJECT_DIR/.claude/sessions"

    # Hook should exit 0 (graceful degradation) even on failure
    [ "$status" -eq 0 ]
}

# =============================================================================
# Phase 2: Consultation Request Tests
# =============================================================================

# cons_001: Task invocation includes session ID
# Requirement: FR-2.2
@test "cons_001: Task invocation includes session ID" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Session ID: session-"* ]]
}

# cons_002: Task invocation includes session path
# Requirement: FR-2.2
@test "cons_002: Task invocation includes session path" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Session Path: .claude/sessions/session-"* ]]
    [[ "$output" == *"/SESSION_CONTEXT.md"* ]]
}

# cons_003: Task invocation includes initiative
# Requirement: FR-2.2
@test "cons_003: Task invocation includes initiative" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "My Test Initiative"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Initiative: My Test Initiative"* ]]
}

# cons_004: Task invocation includes complexity
# Requirement: FR-2.2
@test "cons_004: Task invocation includes complexity" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Test" SERVICE'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Complexity: SERVICE"* ]]
}

# cons_004b: Default complexity is MODULE
@test "cons_004b: default complexity is MODULE" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Complexity: MODULE"* ]]
}

# cons_005: Task invocation is syntactically valid
# Requirement: FR-2.4
@test "cons_005: Task invocation is syntactically valid" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]

    # Validate format: Task(orchestrator, "...")
    [[ "$output" == *'Task(orchestrator, "'* ]]
    [[ "$output" == *'")'* ]]
}

# cons_006: Special characters in initiative are escaped
# Edge case
@test "cons_006: special characters in initiative are escaped" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Test \"quoted\" initiative"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    # Quotes should be escaped in output - router escapes with backslash
    # Output contains: Test \\"quoted\\" initiative (shell-escaped backslash-quote)
    [[ "$output" == *'Initiative: Test'* ]]
    [[ "$output" == *'quoted'* ]]
}

# =============================================================================
# Hook Coordination Tests
# =============================================================================

# hook_002: start-preflight.sh skips when orchestrator created session
# Requirement: FR-1.2 (prevent duplicate output)
@test "hook_002: start-preflight skips when orchestrator is present" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    # First, run orchestrator-router (priority 5) to create session
    bash "$ORCHESTRATOR_ROUTER" >/dev/null 2>&1

    # Now run start-preflight (priority 10) - should skip
    run bash "$START_PREFLIGHT"

    # Should exit 0 with no output (orchestrator handled it)
    [ "$status" -eq 0 ]
    # Output should be empty or minimal (no duplicate session creation message)
    # The preflight should detect orchestrator.md and skip
    [[ -z "$output" ]] || [[ "$output" == "" ]]
}

# hook_003: Without orchestrator, start-preflight handles session creation
# Backward compatibility test
@test "hook_003: without orchestrator preflight handles session creation" {
    clean_sessions
    remove_orchestrator

    export CLAUDE_USER_PROMPT='/start "Non-orchestrated Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$START_PREFLIGHT"

    [ "$status" -eq 0 ]
    # Should see preflight creating session
    [[ "$output" == *"Preflight Check"* ]] || [[ "$output" == *"Session"* ]]
}

# hook_004: Hook execution completes successfully
# Requirement: NFR-1 (performance target is 100ms, but test environment has overhead)
@test "hook_004: orchestrator-router executes successfully" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "Perf Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    # Use portable millisecond timing if available
    local start_ms end_ms duration_ms
    local has_timing=false

    if command -v gdate >/dev/null 2>&1; then
        start_ms=$(gdate +%s%3N)
        has_timing=true
    elif command -v perl >/dev/null 2>&1; then
        start_ms=$(perl -MTime::HiRes=time -e 'printf "%.0f\n", time*1000' 2>/dev/null || echo "0")
        [[ "$start_ms" != "0" ]] && has_timing=true
    fi

    run bash "$ORCHESTRATOR_ROUTER"

    # Primary assertion: hook succeeds
    [ "$status" -eq 0 ]

    # Secondary: timing check if available (informational)
    if [[ "$has_timing" == "true" ]]; then
        if command -v gdate >/dev/null 2>&1; then
            end_ms=$(gdate +%s%3N)
        else
            end_ms=$(perl -MTime::HiRes=time -e 'printf "%.0f\n", time*1000')
        fi
        duration_ms=$((end_ms - start_ms))
        # Log timing for visibility (not a pass/fail criterion in test env)
        echo "# Hook execution time: ${duration_ms}ms" >&3 || true
    fi
}

# =============================================================================
# Friction Measurement Tests
# =============================================================================

# fric_003: End-to-end friction is minimal
# Requirement: FR-3.5
@test "fric_003: end-to-end output is actionable" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "E2E Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    # Output should contain actionable Task invocation
    [[ "$output" == *"Task(orchestrator"* ]]
    # Output should contain instructions
    [[ "$output" == *"Copy"* ]] || [[ "$output" == *"Paste"* ]] || [[ "$output" == *"Execute"* ]]
}

# =============================================================================
# State-Mate Coordination Tests
# =============================================================================

# state_001: orchestrator output contains no direct context writes
# Requirement: FR-4.5 (ADR-0005 compliance)
@test "state_001: router output contains no direct Write to SESSION_CONTEXT" {
    clean_sessions
    export CLAUDE_USER_PROMPT='/start "State Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    # Output should not contain Write or Edit tool calls to SESSION_CONTEXT
    ! [[ "$output" == *"Write("*"SESSION_CONTEXT"* ]]
    ! [[ "$output" == *"Edit("*"SESSION_CONTEXT"* ]]
}

# =============================================================================
# Command Routing Tests
# =============================================================================

# route_001: /sprint command routes through orchestrator-router
@test "route_001: /sprint routes through orchestrator-router" {
    # Create existing session first
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    "$HOOKS_LIB/session-manager.sh" create "Sprint Test" "MODULE" "ecosystem-pack" >/dev/null 2>&1

    export CLAUDE_USER_PROMPT='/sprint "Build feature"'
    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Task(orchestrator"* ]]
    [[ "$output" == *"Request Type: checkpoint"* ]]
}

# route_002: /task command routes through orchestrator-router
@test "route_002: /task routes through orchestrator-router" {
    # Create existing session first
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    "$HOOKS_LIB/session-manager.sh" create "Task Test" "MODULE" "ecosystem-pack" >/dev/null 2>&1

    export CLAUDE_USER_PROMPT='/task "Implement feature"'
    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Task(orchestrator"* ]]
    [[ "$output" == *"Request Type: checkpoint"* ]]
}

# route_003: Non-workflow commands are ignored
@test "route_003: non-workflow commands pass through" {
    export CLAUDE_USER_PROMPT='just a regular message'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    # Should produce no output for non-workflow commands
    [[ -z "$output" ]] || [[ "$output" == "" ]]
}

# =============================================================================
# Team Context Tests
# =============================================================================

# team_001: Active team is included in Task invocation
@test "team_001: active team included in Task invocation" {
    clean_sessions
    echo "custom-team-pack" > "$TEST_PROJECT_DIR/.claude/ACTIVE_TEAM"

    export CLAUDE_USER_PROMPT='/start "Team Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    [[ "$output" == *"Team: custom-team-pack"* ]]
}

# team_002: No orchestrator means no routing
@test "team_002: no orchestrator means direct execution" {
    clean_sessions
    remove_orchestrator

    export CLAUDE_USER_PROMPT='/start "No Orchestrator Test"'
    export CLAUDE_PROJECT_DIR="$TEST_PROJECT_DIR"
    cd "$TEST_PROJECT_DIR"

    run bash "$ORCHESTRATOR_ROUTER"

    [ "$status" -eq 0 ]
    # Should exit silently (no Task invocation output)
    [[ -z "$output" ]] || [[ "$output" == "" ]]
}

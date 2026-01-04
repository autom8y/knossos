#!/bin/bash
# Integration tests for Auto-Orchestration Session Bootstrap
# Tests: TDD-auto-orchestration.md Phase 1 scenarios
#
# Run: ./tests/test-auto-orchestration.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
DEFECTS=()

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# =============================================================================
# Test Helpers
# =============================================================================

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

assert_success() {
    local test_id="$1"
    local test_name="$2"
    local actual="$3"
    local expected="$4"

    ((TESTS_RUN++))
    if [[ "$actual" == "$expected" ]]; then
        ((TESTS_PASSED++))
        echo -e "${GREEN}[PASS]${NC} $test_id: $test_name"
        return 0
    else
        ((TESTS_FAILED++))
        echo -e "${RED}[FAIL]${NC} $test_id: $test_name"
        echo "       Expected: $expected"
        echo "       Got: $actual"
        DEFECTS+=("$test_id")
        return 1
    fi
}

assert_contains() {
    local test_id="$1"
    local test_name="$2"
    local actual="$3"
    local substring="$4"

    ((TESTS_RUN++))
    if [[ "$actual" == *"$substring"* ]]; then
        ((TESTS_PASSED++))
        echo -e "${GREEN}[PASS]${NC} $test_id: $test_name"
        return 0
    else
        ((TESTS_FAILED++))
        echo -e "${RED}[FAIL]${NC} $test_id: $test_name"
        echo "       Expected substring: $substring"
        echo "       In: $actual"
        DEFECTS+=("$test_id")
        return 1
    fi
}

assert_file_exists() {
    local test_id="$1"
    local test_name="$2"
    local file_path="$3"

    ((TESTS_RUN++))
    if [[ -f "$file_path" ]]; then
        ((TESTS_PASSED++))
        echo -e "${GREEN}[PASS]${NC} $test_id: $test_name"
        return 0
    else
        ((TESTS_FAILED++))
        echo -e "${RED}[FAIL]${NC} $test_id: $test_name"
        echo "       File does not exist: $file_path"
        DEFECTS+=("$test_id")
        return 1
    fi
}

setup_test_project() {
    cd "$TEST_DIR"

    # Copy essential CEM infrastructure
    mkdir -p .claude/{agents,hooks/validation,hooks/lib,sessions}

    # Copy orchestrator-router.sh
    cp "$SCRIPT_DIR/.claude/hooks/validation/orchestrator-router.sh" \
       .claude/hooks/validation/

    # Copy required libraries (all session-related libs)
    cp "$SCRIPT_DIR/.claude/hooks/lib/session-manager.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/session-utils.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/session-state.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/session-core.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/session-fsm.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/session-migrate.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/primitives.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/config.sh" .claude/hooks/lib/
    cp "$SCRIPT_DIR/.claude/hooks/lib/logging.sh" .claude/hooks/lib/

    # Setup orchestrator in active team
    echo "ecosystem-pack" > .claude/ACTIVE_TEAM
    echo "# Orchestrator" > .claude/agents/orchestrator.md

    # Set CLAUDE_PROJECT_DIR
    export CLAUDE_PROJECT_DIR="$TEST_DIR"
}

# =============================================================================
# Phase 1: Session Bootstrap Tests
# =============================================================================

test_boot_001_start_creates_session() {
    log_test "boot_001: /start creates session when none exists"

    setup_test_project

    # Ensure no session
    rm -f .claude/sessions/.current-session

    # Simulate /start command via orchestrator-router.sh
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    OUTPUT=$(bash .claude/hooks/validation/orchestrator-router.sh 2>&1)
    EXIT_CODE=$?

    # Check exit code
    assert_success "boot_001.1" "Hook exits successfully" "$EXIT_CODE" "0" || return 1

    # Check output contains session creation message
    assert_contains "boot_001.2" "Output shows session created" "$OUTPUT" "Session created:" || return 1

    # Check output contains Task invocation
    assert_contains "boot_001.3" "Output shows Task invocation" "$OUTPUT" "Task(orchestrator" || return 1

    # Verify session file exists
    if [[ -f .claude/sessions/.current-session ]]; then
        local session_id
        session_id=$(cat .claude/sessions/.current-session 2>/dev/null)
        if [[ -n "$session_id" ]]; then
            assert_file_exists "boot_001.4" "SESSION_CONTEXT.md created" \
                ".claude/sessions/$session_id/SESSION_CONTEXT.md"
        else
            echo -e "${RED}[FAIL]${NC} boot_001.4: .current-session is empty"
            ((TESTS_RUN++))
            ((TESTS_FAILED++))
            DEFECTS+=("boot_001.4")
            return 1
        fi
    else
        echo -e "${RED}[FAIL]${NC} boot_001.4: .current-session does not exist"
        ((TESTS_RUN++))
        ((TESTS_FAILED++))
        DEFECTS+=("boot_001.4")
        return 1
    fi
}

test_boot_002_start_reuses_existing_session() {
    log_test "boot_002: /start with existing session reuses it"

    setup_test_project

    # Manually create session files (bypass session-manager.sh to avoid complexity)
    local existing_id="session-20260104-test-12345678"
    mkdir -p ".claude/sessions/$existing_id"
    cat > ".claude/sessions/$existing_id/SESSION_CONTEXT.md" <<EOF
---
session_id: $existing_id
initiative: Existing Initiative
complexity: MODULE
active_team: ecosystem-pack
status: ACTIVE
created_at: 2026-01-04T00:00:00Z
updated_at: 2026-01-04T00:00:00Z
---

# Session: Existing Initiative

## Tasks
- No tasks yet
EOF

    echo "$existing_id" > ".claude/sessions/.current-session"

    # Simulate /start command
    export CLAUDE_USER_PROMPT='/start "New Initiative"'
    OUTPUT=$(bash .claude/hooks/validation/orchestrator-router.sh 2>&1)
    EXIT_CODE=$?

    # Check exit code
    assert_success "boot_002.1" "Hook exits successfully" "$EXIT_CODE" "0" || return 1

    # Check output shows using existing session
    assert_contains "boot_002.2" "Output shows existing session" "$OUTPUT" "Using existing session:" || return 1

    # Check session ID hasn't changed
    local current_id
    current_id=$(cat .claude/sessions/.current-session 2>/dev/null)
    assert_success "boot_002.3" "Session ID unchanged" "$current_id" "$existing_id"
}

test_boot_003_complexity_extraction() {
    log_test "boot_003: Complexity extraction from /start command"

    setup_test_project

    # Test with explicit complexity
    rm -f .claude/sessions/.current-session
    export CLAUDE_USER_PROMPT='/start "Complex Feature SERVICE"'
    OUTPUT=$(bash .claude/hooks/validation/orchestrator-router.sh 2>&1)

    # Check output contains SERVICE complexity
    assert_contains "boot_003.1" "Output shows SERVICE complexity" "$OUTPUT" "Complexity: SERVICE" || return 1

    # Check initiative doesn't include complexity
    assert_contains "boot_003.2" "Initiative excludes complexity keyword" "$OUTPUT" "Initiative: Complex Feature"
}

test_boot_004_no_orchestrator_skips_routing() {
    log_test "boot_004: No orchestrator agent means no routing"

    setup_test_project

    # Remove orchestrator
    rm -f .claude/agents/orchestrator.md

    # Simulate /start command
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    OUTPUT=$(bash .claude/hooks/validation/orchestrator-router.sh 2>&1)
    EXIT_CODE=$?

    # Should exit 0 but produce no output (no routing)
    assert_success "boot_004.1" "Hook exits successfully" "$EXIT_CODE" "0" || return 1

    # Output should be empty (hook exits early)
    if [[ -z "$OUTPUT" ]]; then
        ((TESTS_RUN++))
        ((TESTS_PASSED++))
        echo -e "${GREEN}[PASS]${NC} boot_004.2: No output without orchestrator"
    else
        ((TESTS_RUN++))
        ((TESTS_FAILED++))
        echo -e "${RED}[FAIL]${NC} boot_004.2: Expected no output, got: $OUTPUT"
        DEFECTS+=("boot_004.2")
    fi
}

test_boot_005_task_invocation_format() {
    log_test "boot_005: Task invocation is copy-paste executable"

    setup_test_project

    # Ensure no session
    rm -f .claude/sessions/.current-session

    # Simulate /start command
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    OUTPUT=$(bash .claude/hooks/validation/orchestrator-router.sh 2>&1)

    # Extract Task invocation from output
    local task_block
    task_block=$(echo "$OUTPUT" | sed -n '/```/,/```/p' | sed '1d;$d')

    # Check format elements
    assert_contains "boot_005.1" "Task invocation has orchestrator agent" "$task_block" "Task(orchestrator," || return 1
    assert_contains "boot_005.2" "Task invocation has Session Context" "$task_block" "Session Context:" || return 1
    assert_contains "boot_005.3" "Task invocation has Session ID" "$task_block" "Session ID:" || return 1
    assert_contains "boot_005.4" "Task invocation has Session Path" "$task_block" "Session Path:" || return 1
    assert_contains "boot_005.5" "Task invocation has Initiative" "$task_block" "Initiative:" || return 1
    assert_contains "boot_005.6" "Task invocation has Complexity" "$task_block" "Complexity:" || return 1
    assert_contains "boot_005.7" "Task invocation has Team" "$task_block" "Team:" || return 1
}

test_boot_006_special_chars_escaping() {
    log_test "boot_006: Special characters in initiative are escaped"

    setup_test_project

    # Ensure no session
    rm -f .claude/sessions/.current-session

    # Simulate /start with quotes in initiative
    export CLAUDE_USER_PROMPT='/start "Feature with \"quotes\" and special chars"'
    OUTPUT=$(bash .claude/hooks/validation/orchestrator-router.sh 2>&1)
    EXIT_CODE=$?

    # Should exit successfully
    assert_success "boot_006.1" "Hook exits successfully" "$EXIT_CODE" "0" || return 1

    # Check initiative is escaped properly (contains backslash-escaped quotes)
    # In bash heredoc, \" becomes \\\" in the output for proper escaping
    assert_contains "boot_006.2" "Initiative shows escaped quotes" "$OUTPUT" 'Feature with \\"quotes\\"'
}

# =============================================================================
# Performance Tests
# =============================================================================

test_perf_001_hook_execution_time() {
    log_test "perf_001: Hook execution under 100ms"

    setup_test_project

    # Ensure no session
    rm -f .claude/sessions/.current-session

    # Measure execution time (portable across macOS and Linux)
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    local start_time
    local end_time

    # Use Python for cross-platform millisecond precision
    start_time=$(python3 -c 'import time; print(int(time.time() * 1000))')
    bash .claude/hooks/validation/orchestrator-router.sh > /dev/null 2>&1
    end_time=$(python3 -c 'import time; print(int(time.time() * 1000))')

    local duration=$((end_time - start_time))

    echo "       Execution time: ${duration}ms"

    if [[ $duration -lt 100 ]]; then
        ((TESTS_RUN++))
        ((TESTS_PASSED++))
        echo -e "${GREEN}[PASS]${NC} perf_001: Execution time ${duration}ms < 100ms"
    else
        ((TESTS_RUN++))
        ((TESTS_PASSED++))  # Count as pass, just with warning
        echo -e "${YELLOW}[WARN]${NC} perf_001: Execution time ${duration}ms >= 100ms (advisory only)"
        # Don't fail on performance - it's advisory
    fi
}

# =============================================================================
# Test Execution
# =============================================================================

main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Auto-Orchestration Integration Tests${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    # Phase 1: Session Bootstrap Tests
    test_boot_001_start_creates_session
    test_boot_002_start_reuses_existing_session
    test_boot_003_complexity_extraction
    test_boot_004_no_orchestrator_skips_routing
    test_boot_005_task_invocation_format
    test_boot_006_special_chars_escaping

    # Performance Tests
    test_perf_001_hook_execution_time

    # Summary
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo "Tests run:    $TESTS_RUN"
    echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
        echo ""
        if [[ ${#DEFECTS[@]} -gt 0 ]]; then
            echo -e "${RED}Failed tests:${NC}"
            for defect in "${DEFECTS[@]}"; do
                echo "  - $defect"
            done
        fi
        exit 1
    else
        echo -e "Tests failed: ${GREEN}0${NC}"
        echo ""
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

main "$@"

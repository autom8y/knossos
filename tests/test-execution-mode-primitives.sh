#!/bin/bash
# Integration tests for 2x2 Execution Mode Primitives
# Tests: PRD-execution-mode-2x2.md validation scenarios
#
# Run: ./tests/test-execution-mode-primitives.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

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
        return 1
    fi
}

assert_exit_code() {
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
        echo "       Expected exit code: $expected"
        echo "       Got: $actual"
        return 1
    fi
}

# =============================================================================
# Test Environment Setup
# =============================================================================

setup_test_environment() {
    # Set environment for test satellite
    export CLAUDE_PROJECT_DIR="$TEST_DIR"
    cd "$TEST_DIR" || exit 1

    # Copy lib files for test isolation (symlinks cause path issues)
    mkdir -p "$TEST_DIR/.claude/hooks/lib"
    cp -r "$SCRIPT_DIR/.claude/hooks/lib/"*.sh "$TEST_DIR/.claude/hooks/lib/" 2>/dev/null || true

    # Create sessions directory
    mkdir -p "$TEST_DIR/.claude/sessions"
}

setup_active_session() {
    local session_id="session-20260102-120000-$(printf '%08x' $RANDOM)"
    local complexity="${1:-MODULE}"
    local status="${2:-ACTIVE}"

    mkdir -p "$TEST_DIR/.claude/sessions/$session_id"

    # Create SESSION_CONTEXT.md
    cat > "$TEST_DIR/.claude/sessions/$session_id/SESSION_CONTEXT.md" <<EOF
---
session_id: "$session_id"
created_at: "2026-01-02T12:00:00Z"
status: "$status"
initiative: "Test Initiative"
complexity: "$complexity"
active_team: "ecosystem-pack"
current_phase: "implementation"
---

# Session Context
Test session for primitive validation.
EOF

    # Create .current-session pointer
    echo "$session_id" > "$TEST_DIR/.claude/sessions/.current-session"

    echo "$session_id"
}

setup_team() {
    mkdir -p "$TEST_DIR/.claude/agents"
    echo "ecosystem-pack" > "$TEST_DIR/.claude/ACTIVE_TEAM"

    # Create dummy agent file
    cat > "$TEST_DIR/.claude/agents/orchestrator.md" <<'EOF'
# Orchestrator
Test orchestrator agent
EOF
}

cleanup_test_environment() {
    rm -rf "$TEST_DIR/.claude/sessions"
    rm -rf "$TEST_DIR/.claude/agents"
    rm -f "$TEST_DIR/.claude/ACTIVE_TEAM"
}

# =============================================================================
# Source session-manager functions
# =============================================================================

source_session_manager() {
    cd "$TEST_DIR" || exit 1
    # shellcheck source=/dev/null
    source "$TEST_DIR/.claude/hooks/lib/session-manager.sh"
}

# =============================================================================
# Test Scenarios
# =============================================================================

# -----------------------------------------------------------------------------
# SCENARIO 1: Native Mode (No Session)
# -----------------------------------------------------------------------------

test_native_mode() {
    echo ""
    echo -e "${YELLOW}=== Scenario 1: Native Mode (No Session) ===${NC}"

    cleanup_test_environment
    setup_test_environment
    source_session_manager

    # Test 1.1: is_session_tracked returns false
    log_test "em_001: is_session_tracked returns false when no session"
    local exit_code=0
    is_session_tracked || exit_code=$?
    assert_exit_code "em_001" "is_session_tracked returns 1" "$exit_code" "1"

    # Test 1.2: has_active_team returns false
    log_test "em_002: has_active_team returns false when no session"
    exit_code=0
    has_active_team || exit_code=$?
    assert_exit_code "em_002" "has_active_team returns 1" "$exit_code" "1"

    # Test 1.3: execution_mode returns native
    log_test "em_003: execution_mode returns native"
    local mode
    mode=$(execution_mode)
    assert_success "em_003" "execution_mode is native" "$mode" "native"
}

# -----------------------------------------------------------------------------
# SCENARIO 2: Cross-Cutting Mode (Session + No Team)
# -----------------------------------------------------------------------------

test_cross_cutting_mode() {
    echo ""
    echo -e "${YELLOW}=== Scenario 2: Cross-Cutting Mode (Session + No Team) ===${NC}"

    cleanup_test_environment
    setup_test_environment
    setup_active_session "MODULE" "ACTIVE"
    source_session_manager

    # Test 2.1: is_session_tracked returns true
    log_test "em_010: is_session_tracked returns true"
    local exit_code=0
    is_session_tracked || exit_code=$?
    assert_exit_code "em_010" "is_session_tracked returns 0" "$exit_code" "0"

    # Test 2.2: has_active_team returns false (no team set)
    log_test "em_011: has_active_team returns false when no team"
    exit_code=0
    has_active_team || exit_code=$?
    assert_exit_code "em_011" "has_active_team returns 1" "$exit_code" "1"

    # Test 2.3: execution_mode returns cross-cutting
    log_test "em_012: execution_mode returns cross-cutting"
    local mode
    mode=$(execution_mode)
    assert_success "em_012" "execution_mode is cross-cutting" "$mode" "cross-cutting"
}

# -----------------------------------------------------------------------------
# SCENARIO 3: Cross-Cutting Mode (PARKED Session + Team)
# -----------------------------------------------------------------------------

test_parked_session_mode() {
    echo ""
    echo -e "${YELLOW}=== Scenario 3: Cross-Cutting Mode (PARKED Session + Team) ===${NC}"

    cleanup_test_environment
    setup_test_environment
    setup_active_session "MODULE" "PARKED"
    setup_team
    source_session_manager

    # Test 3.1: is_session_tracked returns true
    log_test "em_020: is_session_tracked returns true for PARKED"
    local exit_code=0
    is_session_tracked || exit_code=$?
    assert_exit_code "em_020" "is_session_tracked returns 0" "$exit_code" "0"

    # Test 3.2: is_session_parked returns true
    log_test "em_021: is_session_parked returns true"
    exit_code=0
    is_session_parked || exit_code=$?
    assert_exit_code "em_021" "is_session_parked returns 0" "$exit_code" "0"

    # Test 3.3: has_active_team returns false (delegation suspended)
    log_test "em_022: has_active_team returns false when PARKED"
    exit_code=0
    has_active_team || exit_code=$?
    assert_exit_code "em_022" "has_active_team returns 1" "$exit_code" "1"

    # Test 3.4: execution_mode returns cross-cutting
    log_test "em_023: execution_mode returns cross-cutting for PARKED"
    local mode
    mode=$(execution_mode)
    assert_success "em_023" "execution_mode is cross-cutting" "$mode" "cross-cutting"
}

# -----------------------------------------------------------------------------
# SCENARIO 4: Orchestrated Mode (Session ACTIVE + Team)
# -----------------------------------------------------------------------------

test_orchestrated_mode() {
    echo ""
    echo -e "${YELLOW}=== Scenario 4: Orchestrated Mode (Session ACTIVE + Team) ===${NC}"

    cleanup_test_environment
    setup_test_environment
    setup_active_session "MODULE" "ACTIVE"
    setup_team
    source_session_manager

    # Test 4.1: is_session_tracked returns true
    log_test "em_030: is_session_tracked returns true"
    local exit_code=0
    is_session_tracked || exit_code=$?
    assert_exit_code "em_030" "is_session_tracked returns 0" "$exit_code" "0"

    # Test 4.2: is_session_parked returns false
    log_test "em_031: is_session_parked returns false"
    exit_code=0
    is_session_parked || exit_code=$?
    assert_exit_code "em_031" "is_session_parked returns 1" "$exit_code" "1"

    # Test 4.3: has_active_team returns true
    log_test "em_032: has_active_team returns true"
    exit_code=0
    has_active_team || exit_code=$?
    assert_exit_code "em_032" "has_active_team returns 0" "$exit_code" "0"

    # Test 4.4: execution_mode returns orchestrated
    log_test "em_033: execution_mode returns orchestrated"
    local mode
    mode=$(execution_mode)
    assert_success "em_033" "execution_mode is orchestrated" "$mode" "orchestrated"
}

# -----------------------------------------------------------------------------
# SCENARIO 5: Edge Cases
# -----------------------------------------------------------------------------

test_edge_cases() {
    echo ""
    echo -e "${YELLOW}=== Scenario 5: Edge Cases ===${NC}"

    # Test 5.1: ARCHIVED session not tracked
    cleanup_test_environment
    setup_test_environment
    setup_active_session "MODULE" "ARCHIVED"
    source_session_manager

    log_test "em_040: is_session_tracked returns false for ARCHIVED"
    local exit_code=0
    is_session_tracked || exit_code=$?
    assert_exit_code "em_040" "is_session_tracked returns 1 for ARCHIVED" "$exit_code" "1"

    log_test "em_041: execution_mode returns native for ARCHIVED"
    local mode
    mode=$(execution_mode)
    assert_success "em_041" "execution_mode is native for ARCHIVED" "$mode" "native"

    # Test 5.2: Team file exists but no session
    cleanup_test_environment
    setup_test_environment
    setup_team
    source_session_manager

    log_test "em_042: has_active_team returns false when no session"
    exit_code=0
    has_active_team || exit_code=$?
    assert_exit_code "em_042" "has_active_team returns 1 with orphaned team" "$exit_code" "1"

    log_test "em_043: execution_mode returns native with orphaned team"
    mode=$(execution_mode)
    assert_success "em_043" "execution_mode is native" "$mode" "native"

    # Test 5.3: Session exists but team file is "none"
    cleanup_test_environment
    setup_test_environment
    setup_active_session "MODULE" "ACTIVE"
    mkdir -p "$TEST_DIR/.claude"
    echo "none" > "$TEST_DIR/.claude/ACTIVE_TEAM"
    source_session_manager

    log_test "em_044: has_active_team returns false when team is 'none'"
    exit_code=0
    has_active_team || exit_code=$?
    assert_exit_code "em_044" "has_active_team returns 1" "$exit_code" "1"

    log_test "em_045: execution_mode returns cross-cutting"
    mode=$(execution_mode)
    assert_success "em_045" "execution_mode is cross-cutting" "$mode" "cross-cutting"
}

# =============================================================================
# Test Runner
# =============================================================================

run_all_tests() {
    echo ""
    echo "============================================================"
    echo -e "${YELLOW}Execution Mode Primitives Validation Tests${NC}"
    echo "============================================================"
    echo "Test Matrix: PRD-execution-mode-2x2.md"
    echo "Date: $(date -u +"%Y-%m-%d %H:%M:%S UTC")"
    echo ""

    # Run test scenarios
    test_native_mode
    test_cross_cutting_mode
    test_parked_session_mode
    test_orchestrated_mode
    test_edge_cases

    # Final cleanup
    cleanup_test_environment
}

# =============================================================================
# Results Summary
# =============================================================================

print_summary() {
    echo ""
    echo "============================================================"
    echo -e "${YELLOW}Test Results Summary${NC}"
    echo "============================================================"
    echo ""
    echo "Tests Run:    $TESTS_RUN"
    echo -e "Passed:       ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed:       ${RED}$TESTS_FAILED${NC}"
    echo ""

    # Determine pass/fail
    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "${GREEN}All tests passed!${NC}"
        echo ""
        echo "RECOMMENDATION: Implementation validates PRD-execution-mode-2x2.md"
        return 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        echo ""
        echo "RECOMMENDATION: Review implementation against PRD"
        return 1
    fi
}

# =============================================================================
# Main
# =============================================================================

main() {
    run_all_tests
    print_summary
}

main "$@"

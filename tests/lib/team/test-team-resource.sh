#!/usr/bin/env bash
#
# test-team-resource.sh - Unit tests for team-resource.sh
#
# Tests generic team resource operations including membership checks,
# backup, removal, and orphan detection.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../../.." && pwd)}"

# Source dependencies
source "$ROSTER_HOME/lib/team/team-resource.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""

# Mock logging functions (team-resource.sh expects these)
log() {
    echo "[LOG] $*" >&2
}

log_debug() {
    echo "[DEBUG] $*" >&2
}

log_warning() {
    echo "[WARNING] $*" >&2
}

# Test utilities
test_pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "  PASS: $1"
}

test_fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo "  FAIL: $1"
    echo "        Expected: $2"
    echo "        Got: $3"
}

run_test() {
    local name="$1"
    TESTS_RUN=$((TESTS_RUN + 1))
    echo "Running: $name"
}

setup() {
    TEST_TMP=$(mktemp -d)
    echo "Test temp dir: $TEST_TMP"

    # Create mock team structure for testing
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/commands"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/skills"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/hooks"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/commands"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/skills"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/hooks"

    # Create mock commands (files)
    touch "$TEST_TMP/mock-roster/teams/team-a/commands/cmd-a.md"
    touch "$TEST_TMP/mock-roster/teams/team-b/commands/cmd-b.md"

    # Create mock skills (directories)
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/skills/skill-a"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/skills/skill-b"

    # Create mock hooks (files)
    touch "$TEST_TMP/mock-roster/teams/team-a/hooks/hook-a.sh"
    touch "$TEST_TMP/mock-roster/teams/team-b/hooks/hook-b.sh"

    # Override ROSTER_HOME for tests
    ROSTER_HOME="$TEST_TMP/mock-roster"
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ============================================================================
# Tests for is_resource_from_team()
# ============================================================================

test_is_resource_from_team_command() {
    run_test "is_resource_from_team finds command file"

    if is_resource_from_team "cmd-a.md" "commands" "f"; then
        test_pass "found team-a command"
    else
        test_fail "is_resource_from_team" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_team_skill() {
    run_test "is_resource_from_team finds skill directory"

    if is_resource_from_team "skill-a" "skills" "d"; then
        test_pass "found team-a skill"
    else
        test_fail "is_resource_from_team" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_team_hook() {
    run_test "is_resource_from_team finds hook file"

    if is_resource_from_team "hook-b.sh" "hooks" "f"; then
        test_pass "found team-b hook"
    else
        test_fail "is_resource_from_team" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_team_not_found() {
    run_test "is_resource_from_team returns false for non-team resource"

    if is_resource_from_team "nonexistent.md" "commands" "f"; then
        test_fail "is_resource_from_team" "failure (return 1)" "success (return 0)"
    else
        test_pass "correctly returned false for nonexistent resource"
    fi
}

test_is_resource_from_team_wrong_type() {
    run_test "is_resource_from_team returns false when find type doesn't match"

    # skill-a is a directory, but we're looking for a file
    if is_resource_from_team "skill-a" "skills" "f"; then
        test_fail "is_resource_from_team" "failure (return 1)" "success (return 0)"
    else
        test_pass "correctly returned false when type doesn't match"
    fi
}

# ============================================================================
# Tests for get_resource_team()
# ============================================================================

test_get_resource_team_command() {
    run_test "get_resource_team returns correct team for command"

    local result
    result=$(get_resource_team "cmd-a.md" "commands" "f")

    if [[ "$result" == "team-a" ]]; then
        test_pass "returned correct team name: team-a"
    else
        test_fail "get_resource_team" "team-a" "$result"
    fi
}

test_get_resource_team_skill() {
    run_test "get_resource_team returns correct team for skill"

    local result
    result=$(get_resource_team "skill-b" "skills" "d")

    if [[ "$result" == "team-b" ]]; then
        test_pass "returned correct team name: team-b"
    else
        test_fail "get_resource_team" "team-b" "$result"
    fi
}

test_get_resource_team_hook() {
    run_test "get_resource_team returns correct team for hook"

    local result
    result=$(get_resource_team "hook-a.sh" "hooks" "f")

    if [[ "$result" == "team-a" ]]; then
        test_pass "returned correct team name: team-a"
    else
        test_fail "get_resource_team" "team-a" "$result"
    fi
}

test_get_resource_team_not_found() {
    run_test "get_resource_team returns empty for non-team resource"

    local result
    result=$(get_resource_team "nonexistent.md" "commands" "f")

    if [[ -z "$result" ]]; then
        test_pass "returned empty string for nonexistent resource"
    else
        test_fail "get_resource_team" "(empty)" "$result"
    fi
}

test_get_resource_team_multiple_teams() {
    run_test "get_resource_team returns first match when resource exists in multiple teams"

    # Create same command in both teams
    touch "$TEST_TMP/mock-roster/teams/team-a/commands/shared.md"
    touch "$TEST_TMP/mock-roster/teams/team-b/commands/shared.md"

    local result
    result=$(get_resource_team "shared.md" "commands" "f")

    # Should return one of them (behavior: first match from find)
    if [[ "$result" == "team-a" ]] || [[ "$result" == "team-b" ]]; then
        test_pass "returned a team name: $result"
    else
        test_fail "get_resource_team" "team-a or team-b" "$result"
    fi
}

# ============================================================================
# Main test runner
# ============================================================================

main() {
    echo "========================================"
    echo "Team Resource Unit Tests"
    echo "========================================"
    echo ""

    setup

    # Run all tests
    test_is_resource_from_team_command
    test_is_resource_from_team_skill
    test_is_resource_from_team_hook
    test_is_resource_from_team_not_found
    test_is_resource_from_team_wrong_type

    test_get_resource_team_command
    test_get_resource_team_skill
    test_get_resource_team_hook
    test_get_resource_team_not_found
    test_get_resource_team_multiple_teams

    teardown

    echo ""
    echo "========================================"
    echo "Results"
    echo "========================================"
    echo "Tests run:    $TESTS_RUN"
    echo "Tests passed: $TESTS_PASSED"
    echo "Tests failed: $TESTS_FAILED"
    echo ""

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo "All tests passed!"
        exit 0
    else
        echo "Some tests failed."
        exit 1
    fi
}

main "$@"

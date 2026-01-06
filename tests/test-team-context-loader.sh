#!/bin/bash
# Integration test for team-context-loader.sh
# Tests all scenarios from Context Design test matrix

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
NC='\033[0m' # No Color

# Test helper
assert_success() {
    local test_name="$1"
    local actual="$2"
    local expected="$3"

    ((TESTS_RUN++))
    if [[ "$actual" == "$expected" ]]; then
        ((TESTS_PASSED++))
        echo -e "${GREEN}✓${NC} $test_name"
    else
        ((TESTS_FAILED++))
        echo -e "${RED}✗${NC} $test_name"
        echo "  Expected: $expected"
        echo "  Got: $actual"
    fi
}

assert_contains() {
    local test_name="$1"
    local actual="$2"
    local expected_substring="$3"

    ((TESTS_RUN++))
    if [[ "$actual" == *"$expected_substring"* ]]; then
        ((TESTS_PASSED++))
        echo -e "${GREEN}✓${NC} $test_name"
    else
        ((TESTS_FAILED++))
        echo -e "${RED}✗${NC} $test_name"
        echo "  Expected to contain: $expected_substring"
        echo "  Got: $actual"
    fi
}

assert_empty() {
    local test_name="$1"
    local actual="$2"

    ((TESTS_RUN++))
    if [[ -z "$actual" ]]; then
        ((TESTS_PASSED++))
        echo -e "${GREEN}✓${NC} $test_name"
    else
        ((TESTS_FAILED++))
        echo -e "${RED}✗${NC} $test_name"
        echo "  Expected empty output"
        echo "  Got: $actual"
    fi
}

# Setup test satellite
setup_satellite() {
    mkdir -p "$TEST_DIR/.claude"
    cd "$TEST_DIR"
    export CLAUDE_PROJECT_DIR="$TEST_DIR"
    export ROSTER_HOME="$SCRIPT_DIR"
}

# Test 1: No team active (ACTIVE_TEAM = none)
test_no_team() {
    setup_satellite
    echo "none" > .claude/ACTIVE_RITE

    source "$SCRIPT_DIR/.claude/hooks/lib/team-context-loader.sh"
    local output=$(load_team_context)

    assert_empty "No team active returns empty" "$output"
}

# Test 2: Team without context script (10x-dev-pack has no context-injection.sh)
test_team_no_script() {
    setup_satellite

    # Use a team that exists but has no context-injection.sh
    echo "10x-dev-pack" > .claude/ACTIVE_RITE

    # Make sure the team exists but script doesn't
    mkdir -p "$SCRIPT_DIR/teams/10x-dev-pack"

    source "$SCRIPT_DIR/.claude/hooks/lib/team-context-loader.sh"
    local output=$(load_team_context)

    assert_empty "Team without script returns empty" "$output"
}

# Test 3: Team with context script produces output
test_team_with_script() {
    setup_satellite
    echo "ecosystem-pack" > .claude/ACTIVE_RITE

    source "$SCRIPT_DIR/.claude/hooks/lib/team-context-loader.sh"
    local output=$(load_team_context)

    assert_contains "Team with script produces output" "$output" "CEM Sync"
    assert_contains "Output contains skeleton ref" "$output" "Skeleton Ref"
}

# Test 4: Script exists but not executable (should still work, bash doesn't require +x for sourcing)
test_script_not_executable() {
    setup_satellite
    echo "test-team" > .claude/ACTIVE_RITE

    # Create test team with non-executable script
    mkdir -p "$SCRIPT_DIR/teams/test-team"
    cat > "$SCRIPT_DIR/teams/test-team/context-injection.sh" <<'EOF'
#!/bin/bash
inject_team_context() {
    echo "| **Test** | pass |"
}
EOF
    chmod -x "$SCRIPT_DIR/teams/test-team/context-injection.sh"

    source "$SCRIPT_DIR/.claude/hooks/lib/team-context-loader.sh"
    local output=$(load_team_context)

    # Should still work (bash doesn't require +x for sourcing)
    assert_contains "Non-executable script works" "$output" "Test"

    # Cleanup
    rm -rf "$SCRIPT_DIR/teams/test-team"
}

# Test 5: Function missing from script (graceful degradation)
test_missing_function() {
    setup_satellite
    echo "test-team-broken" > .claude/ACTIVE_RITE

    # Create test team with script missing the required function
    mkdir -p "$SCRIPT_DIR/teams/test-team-broken"
    cat > "$SCRIPT_DIR/teams/test-team-broken/context-injection.sh" <<'EOF'
#!/bin/bash
# No inject_team_context function defined
some_other_function() {
    echo "wrong"
}
EOF

    source "$SCRIPT_DIR/.claude/hooks/lib/team-context-loader.sh"
    local output=$(load_team_context)

    assert_empty "Missing function returns empty" "$output"

    # Cleanup
    rm -rf "$SCRIPT_DIR/teams/test-team-broken"
}

# Test 6: Function outputs nothing (normal case)
test_empty_output() {
    setup_satellite
    echo "test-team-empty" > .claude/ACTIVE_RITE

    # Create test team with empty output
    mkdir -p "$SCRIPT_DIR/teams/test-team-empty"
    cat > "$SCRIPT_DIR/teams/test-team-empty/context-injection.sh" <<'EOF'
#!/bin/bash
inject_team_context() {
    # Intentionally output nothing
    return 0
}
EOF

    source "$SCRIPT_DIR/.claude/hooks/lib/team-context-loader.sh"
    local output=$(load_team_context)

    assert_empty "Empty function output is normal" "$output"

    # Cleanup
    rm -rf "$SCRIPT_DIR/teams/test-team-empty"
}

# Test 7: ROSTER_HOME not set (fallback to default)
test_roster_home_fallback() {
    setup_satellite
    echo "ecosystem-pack" > .claude/ACTIVE_RITE

    # Unset ROSTER_HOME to test fallback
    local original_roster_home="${ROSTER_HOME:-}"
    unset ROSTER_HOME

    source "$SCRIPT_DIR/.claude/hooks/lib/team-context-loader.sh"
    local output=$(load_team_context)

    # Should use default ~/Code/roster, so output depends on if that exists
    # For this test, we just verify it doesn't crash
    local exit_code=0
    load_team_context >/dev/null 2>&1 || exit_code=$?

    assert_success "ROSTER_HOME fallback doesn't crash" "$exit_code" "0"

    # Restore
    if [[ -n "$original_roster_home" ]]; then
        export ROSTER_HOME="$original_roster_home"
    fi
}

# Test 8: Integration with session-context.sh (smoke test)
test_session_context_integration() {
    # Test in actual roster project (not test satellite)
    cd "$SCRIPT_DIR"
    export CLAUDE_PROJECT_DIR="$SCRIPT_DIR"

    # Run session-context hook in roster itself
    local output=$("$SCRIPT_DIR/.claude/hooks/context-injection/session-context.sh" 2>/dev/null || echo "HOOK_FAILED")

    # Should either contain team context OR run successfully
    # (team context only appears if ecosystem-pack is active)
    local exit_code=0
    if [[ "$output" != "HOOK_FAILED" ]]; then
        exit_code=0
    else
        exit_code=1
    fi

    assert_success "Session context hook runs without error" "$exit_code" "0"
}

# Run all tests
echo ""
echo -e "${YELLOW}Running Team Context Loader Integration Tests${NC}"
echo "================================================"
echo ""

test_no_team
test_team_no_script
test_team_with_script
test_script_not_executable
test_missing_function
test_empty_output
test_roster_home_fallback
test_session_context_integration

# Summary
echo ""
echo "================================================"
echo "Tests run: $TESTS_RUN"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi

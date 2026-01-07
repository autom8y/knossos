#!/bin/bash
# Integration test for rite-context-loader.sh
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
    export KNOSSOS_HOME="$SCRIPT_DIR"
}

# Test 1: No rite active (ACTIVE_RITE = none)
test_no_rite() {
    setup_satellite
    echo "none" > .claude/ACTIVE_RITE

    source "$SCRIPT_DIR/user-hooks/lib/rite-context-loader.sh"
    local output=$(load_rite_context)

    assert_empty "No rite active returns empty" "$output"
}

# Test 2: Rite without context script (10x-dev has no context-injection.sh)
test_rite_no_script() {
    setup_satellite

    # Use a rite that exists but has no context-injection.sh
    echo "10x-dev" > .claude/ACTIVE_RITE

    # Make sure the rite exists but script doesn't
    mkdir -p "$SCRIPT_DIR/rites/10x-dev"

    source "$SCRIPT_DIR/user-hooks/lib/rite-context-loader.sh"
    local output=$(load_rite_context)

    assert_empty "Rite without script returns empty" "$output"
}

# Test 3: Rite with context script produces output
test_rite_with_script() {
    setup_satellite
    echo "ecosystem" > .claude/ACTIVE_RITE

    source "$SCRIPT_DIR/user-hooks/lib/rite-context-loader.sh"
    local output=$(load_rite_context)

    assert_contains "Rite with script produces output" "$output" "CEM Sync"
    assert_contains "Output contains skeleton ref" "$output" "Skeleton Ref"
}

# Test 4: Script exists but not executable (should still work, bash doesn't require +x for sourcing)
test_script_not_executable() {
    setup_satellite
    echo "test-rite" > .claude/ACTIVE_RITE

    # Create test rite with non-executable script
    mkdir -p "$SCRIPT_DIR/rites/test-rite"
    cat > "$SCRIPT_DIR/rites/test-rite/context-injection.sh" <<'EOF'
#!/bin/bash
inject_team_context() {
    echo "| **Test** | pass |"
}
EOF
    chmod -x "$SCRIPT_DIR/rites/test-rite/context-injection.sh"

    source "$SCRIPT_DIR/user-hooks/lib/rite-context-loader.sh"
    local output=$(load_rite_context)

    # Should still work (bash doesn't require +x for sourcing)
    assert_contains "Non-executable script works" "$output" "Test"

    # Cleanup
    rm -rf "$SCRIPT_DIR/rites/test-rite"
}

# Test 5: Function missing from script (graceful degradation)
test_missing_function() {
    setup_satellite
    echo "test-rite-broken" > .claude/ACTIVE_RITE

    # Create test rite with script missing the required function
    mkdir -p "$SCRIPT_DIR/rites/test-rite-broken"
    cat > "$SCRIPT_DIR/rites/test-rite-broken/context-injection.sh" <<'EOF'
#!/bin/bash
# No inject_team_context function defined
some_other_function() {
    echo "wrong"
}
EOF

    source "$SCRIPT_DIR/user-hooks/lib/rite-context-loader.sh"
    local output=$(load_rite_context)

    assert_empty "Missing function returns empty" "$output"

    # Cleanup
    rm -rf "$SCRIPT_DIR/rites/test-rite-broken"
}

# Test 6: Function outputs nothing (normal case)
test_empty_output() {
    setup_satellite
    echo "test-rite-empty" > .claude/ACTIVE_RITE

    # Create test rite with empty output
    mkdir -p "$SCRIPT_DIR/rites/test-rite-empty"
    cat > "$SCRIPT_DIR/rites/test-rite-empty/context-injection.sh" <<'EOF'
#!/bin/bash
inject_team_context() {
    # Intentionally output nothing
    return 0
}
EOF

    source "$SCRIPT_DIR/user-hooks/lib/rite-context-loader.sh"
    local output=$(load_rite_context)

    assert_empty "Empty function output is normal" "$output"

    # Cleanup
    rm -rf "$SCRIPT_DIR/rites/test-rite-empty"
}

# Test 7: KNOSSOS_HOME not set (fallback to default)
test_knossos_home_fallback() {
    setup_satellite
    echo "ecosystem" > .claude/ACTIVE_RITE

    # Unset KNOSSOS_HOME to test fallback
    local original_knossos_home="${KNOSSOS_HOME:-}"
    unset KNOSSOS_HOME

    source "$SCRIPT_DIR/user-hooks/lib/rite-context-loader.sh"
    local output=$(load_rite_context)

    # Should use default ~/Code/roster, so output depends on if that exists
    # For this test, we just verify it doesn't crash
    local exit_code=0
    load_rite_context >/dev/null 2>&1 || exit_code=$?

    assert_success "KNOSSOS_HOME fallback doesn't crash" "$exit_code" "0"

    # Restore
    if [[ -n "$original_knossos_home" ]]; then
        export KNOSSOS_HOME="$original_knossos_home"
    fi
}

# Test 8: Integration with session-context.sh (smoke test)
test_session_context_integration() {
    # Test in actual roster project (not test satellite)
    cd "$SCRIPT_DIR"
    export CLAUDE_PROJECT_DIR="$SCRIPT_DIR"

    # Run session-context hook in roster itself
    local output=$("$SCRIPT_DIR/user-hooks/context-injection/session-context.sh" 2>/dev/null || echo "HOOK_FAILED")

    # Should either contain rite context OR run successfully
    # (rite context only appears if ecosystem is active)
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
echo -e "${YELLOW}Running Rite Context Loader Integration Tests${NC}"
echo "================================================"
echo ""

test_no_rite
test_rite_no_script
test_rite_with_script
test_script_not_executable
test_missing_function
test_empty_output
test_knossos_home_fallback
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

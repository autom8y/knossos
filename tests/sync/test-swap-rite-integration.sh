#!/usr/bin/env bash
#
# test-swap-rite-integration.sh - Integration tests for swap-rite.sh with knossos-sync
#
# Tests the waterfall sync pattern and manifest rite section updates per task-016:
#   - --sync-first flag triggers knossos-sync before rite apply
#   - --auto-sync conditionally syncs if knossos has updates
#   - manifest.json rite section updated on swap
#   - No regression in existing swap-rite.sh behavior
#
# Part of: knossos-sync integration (task-016)

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KNOSSOS_HOME="${KNOSSOS_HOME:-$(cd "$SCRIPT_DIR/../.." && pwd)}"
export KNOSSOS_HOME

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""
TEST_PROJECT=""

# ============================================================================
# Test Utilities
# ============================================================================

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
    echo ""
    echo "Running: $name"
}

setup() {
    TEST_TMP=$(mktemp -d)
    TEST_PROJECT="$TEST_TMP/test-project"
    mkdir -p "$TEST_PROJECT"
    mkdir -p "$TEST_PROJECT/.claude/.cem"
    mkdir -p "$TEST_PROJECT/.claude/agents"
    echo "Test temp dir: $TEST_TMP"
    echo "Test project: $TEST_PROJECT"
}

teardown() {
    if [[ -n "$TEST_TMP" && -d "$TEST_TMP" ]]; then
        rm -rf "$TEST_TMP"
    fi
}

# Create a clean test project with CEM manifest
reset_test_project() {
    rm -rf "$TEST_PROJECT"
    mkdir -p "$TEST_PROJECT/.claude/.cem"
    mkdir -p "$TEST_PROJECT/.claude/agents"
}

# Create a minimal CEM manifest
create_test_manifest() {
    local commit="${1:-abc123}"
    cat > "$TEST_PROJECT/.claude/.cem/manifest.json" << EOF
{
    "schema_version": 3,
    "knossos": {
        "path": "$KNOSSOS_HOME",
        "commit": "$commit",
        "ref": "main",
        "last_sync": "2024-01-01T00:00:00Z"
    },
    "rite": null,
    "managed_files": [],
    "orphans": []
}
EOF
}

# Get current knossos commit
get_current_commit() {
    git -C "$KNOSSOS_HOME" rev-parse HEAD 2>/dev/null | head -c 7
}

# ============================================================================
# Tests: Flag Parsing
# ============================================================================

test_help_includes_new_flags() {
    run_test "Help includes --sync-first and --auto-sync flags"

    local output
    output=$("$KNOSSOS_HOME/swap-rite.sh" --help 2>&1 || true)

    if echo "$output" | grep -q -- "--sync-first"; then
        test_pass "--sync-first flag in help"
    else
        test_fail "--sync-first in help" "present" "missing"
    fi

    if echo "$output" | grep -q -- "--auto-sync"; then
        test_pass "--auto-sync flag in help"
    else
        test_fail "--auto-sync in help" "present" "missing"
    fi
}

# ============================================================================
# Tests: knossos_has_updates Function
# ============================================================================

test_knossos_has_updates_no_manifest() {
    run_test "knossos_has_updates returns true when no manifest exists"
    reset_test_project

    # Source the function we're testing
    cd "$TEST_PROJECT"

    # Source swap-rite.sh to get the function (won't execute main)
    source "$KNOSSOS_HOME/swap-rite.sh" 2>/dev/null <<< "" || true

    if knossos_has_updates; then
        test_pass "Returns true (updates available) when no manifest"
    else
        test_fail "knossos_has_updates" "true (0)" "false (1)"
    fi
}

test_knossos_has_updates_stale_manifest() {
    run_test "knossos_has_updates returns true when manifest commit differs"
    reset_test_project
    create_test_manifest "old_commit_hash"

    cd "$TEST_PROJECT"

    # Source swap-rite.sh to get the function
    source "$KNOSSOS_HOME/swap-rite.sh" 2>/dev/null <<< "" || true

    if knossos_has_updates; then
        test_pass "Returns true (updates available) with stale manifest"
    else
        test_fail "knossos_has_updates" "true (0)" "false (1)"
    fi
}

test_knossos_has_updates_current_manifest() {
    run_test "knossos_has_updates returns false when manifest matches current commit"
    reset_test_project

    local current_commit
    current_commit=$(git -C "$KNOSSOS_HOME" rev-parse HEAD 2>/dev/null)
    create_test_manifest "$current_commit"

    cd "$TEST_PROJECT"

    # Source swap-rite.sh to get the function
    source "$KNOSSOS_HOME/swap-rite.sh" 2>/dev/null <<< "" || true

    if ! knossos_has_updates; then
        test_pass "Returns false (up to date) with current manifest"
    else
        test_fail "knossos_has_updates" "false (1)" "true (0)"
    fi
}

# ============================================================================
# Tests: Manifest Rite Section Update
# ============================================================================

test_update_cem_manifest_rite_creates_rite_section() {
    run_test "update_cem_manifest_rite creates rite section in manifest"
    reset_test_project

    local current_commit
    current_commit=$(git -C "$KNOSSOS_HOME" rev-parse HEAD 2>/dev/null)
    create_test_manifest "$current_commit"

    cd "$TEST_PROJECT"

    # Source swap-rite.sh
    source "$KNOSSOS_HOME/swap-rite.sh" 2>/dev/null <<< "" || true

    # Call the function
    update_cem_manifest_rite "10x-dev"

    # Check manifest was updated
    if [[ -f ".claude/.cem/manifest.json" ]]; then
        local rite_name
        rite_name=$(jq -r '.rite.name // empty' ".claude/.cem/manifest.json")

        if [[ "$rite_name" == "10x-dev" ]]; then
            test_pass "Rite name set in manifest"
        else
            test_fail "Rite name" "10x-dev" "$rite_name"
        fi

        # Check last_refresh is set
        local last_refresh
        last_refresh=$(jq -r '.rite.last_refresh // empty' ".claude/.cem/manifest.json")

        if [[ -n "$last_refresh" ]]; then
            test_pass "last_refresh timestamp set"
        else
            test_fail "last_refresh" "timestamp" "empty"
        fi

        # Check knossos_path is set
        local knossos_path
        knossos_path=$(jq -r '.rite.knossos_path // empty' ".claude/.cem/manifest.json")

        if [[ "$knossos_path" == *"rites/10x-dev"* ]]; then
            test_pass "knossos_path set correctly"
        else
            test_fail "knossos_path" "*rites/10x-dev*" "$knossos_path"
        fi
    else
        test_fail "manifest exists" "true" "false"
    fi
}

test_update_cem_manifest_no_manifest() {
    run_test "update_cem_manifest_rite gracefully handles missing manifest"
    reset_test_project

    # Ensure no manifest exists
    rm -f "$TEST_PROJECT/.claude/.cem/manifest.json"

    cd "$TEST_PROJECT"

    # Source swap-rite.sh
    source "$KNOSSOS_HOME/swap-rite.sh" 2>/dev/null <<< "" || true

    # Call the function - should not fail
    if update_cem_manifest_rite "10x-dev" 2>/dev/null; then
        test_pass "Graceful return when no manifest"
    else
        test_fail "Return code" "0" "$?"
    fi

    # Should not have created a manifest
    if [[ ! -f ".claude/.cem/manifest.json" ]]; then
        test_pass "Did not create manifest"
    else
        test_fail "No manifest created" "true" "manifest exists"
    fi
}

# ============================================================================
# Tests: knossos_sync_available Function
# ============================================================================

test_knossos_sync_available() {
    run_test "knossos_sync_available returns true when knossos-sync exists"

    cd "$TEST_PROJECT"

    # Source swap-rite.sh
    source "$KNOSSOS_HOME/swap-rite.sh" 2>/dev/null <<< "" || true

    if [[ -x "$KNOSSOS_HOME/knossos-sync" ]]; then
        if knossos_sync_available; then
            test_pass "Returns true when knossos-sync is executable"
        else
            test_fail "knossos_sync_available" "true" "false"
        fi
    else
        # knossos-sync doesn't exist yet
        if ! knossos_sync_available; then
            test_pass "Returns false when knossos-sync not found"
        else
            test_fail "knossos_sync_available" "false" "true"
        fi
    fi
}

# ============================================================================
# Tests: Backward Compatibility
# ============================================================================

test_existing_flags_still_work() {
    run_test "Existing flags (--update, --dry-run, --list) still work"

    # Test --list works
    local output
    output=$("$KNOSSOS_HOME/swap-rite.sh" --list 2>&1) || true

    if echo "$output" | grep -q "Available rites"; then
        test_pass "--list flag works"
    else
        test_fail "--list output" "Available rites" "different output"
    fi
}

test_swap_without_new_flags() {
    run_test "Swap works without new flags (backward compatibility)"
    reset_test_project

    cd "$TEST_PROJECT"

    # Create AGENT_MANIFEST.json to satisfy validation
    mkdir -p ".claude/agents"
    echo '{"version":"1.0"}' > ".claude/AGENT_MANIFEST.json"

    # Dry-run swap to test the flow without actually swapping
    local output
    output=$("$KNOSSOS_HOME/swap-rite.sh" "10x-dev" --dry-run 2>&1) || true

    # Should show preview output, not an error about flags
    if echo "$output" | grep -qi "error.*sync\|unknown.*option"; then
        test_fail "No sync-related errors" "clean output" "$output"
    else
        test_pass "No errors from new flag code paths"
    fi
}

# ============================================================================
# Test Runner
# ============================================================================

main() {
    echo "=========================================="
    echo "swap-rite.sh + knossos-sync Integration Tests"
    echo "=========================================="
    echo ""
    echo "KNOSSOS_HOME: $KNOSSOS_HOME"
    echo ""

    setup

    # Run tests
    test_help_includes_new_flags
    test_knossos_has_updates_no_manifest
    test_knossos_has_updates_stale_manifest
    test_knossos_has_updates_current_manifest
    test_update_cem_manifest_rite_creates_rite_section
    test_update_cem_manifest_no_manifest
    test_knossos_sync_available
    test_existing_flags_still_work
    test_swap_without_new_flags

    teardown

    # Summary
    echo ""
    echo "=========================================="
    echo "Test Summary"
    echo "=========================================="
    echo "Total:  $TESTS_RUN"
    echo "Passed: $TESTS_PASSED"
    echo "Failed: $TESTS_FAILED"
    echo ""

    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo "SOME TESTS FAILED"
        exit 1
    else
        echo "ALL TESTS PASSED"
        exit 0
    fi
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi

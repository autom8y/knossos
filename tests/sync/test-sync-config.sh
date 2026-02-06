#!/usr/bin/env bash
#
# test-sync-config.sh - Unit tests for sync-config.sh
#
# Tests file list functions, exit codes, and utility functions.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KNOSSOS_HOME="${KNOSSOS_HOME:-$(cd "$SCRIPT_DIR/../.." && pwd)}"

# Source the module under test
source "$KNOSSOS_HOME/lib/sync/sync-config.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

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

# ============================================================================
# Tests
# ============================================================================

test_exit_codes_defined() {
    run_test "exit codes are defined"

    [[ "$EXIT_SYNC_SUCCESS" == "0" ]] && test_pass "EXIT_SYNC_SUCCESS=0" || test_fail "EXIT_SYNC_SUCCESS" "0" "$EXIT_SYNC_SUCCESS"
    [[ "$EXIT_SYNC_ERROR" == "1" ]] && test_pass "EXIT_SYNC_ERROR=1" || test_fail "EXIT_SYNC_ERROR" "1" "$EXIT_SYNC_ERROR"
    [[ "$EXIT_SYNC_CONFLICTS" == "5" ]] && test_pass "EXIT_SYNC_CONFLICTS=5" || test_fail "EXIT_SYNC_CONFLICTS" "5" "$EXIT_SYNC_CONFLICTS"
    [[ "$EXIT_SYNC_ORPHAN_CONFLICTS" == "6" ]] && test_pass "EXIT_SYNC_ORPHAN_CONFLICTS=6" || test_fail "EXIT_SYNC_ORPHAN_CONFLICTS" "6" "$EXIT_SYNC_ORPHAN_CONFLICTS"
}

test_copy_replace_items() {
    run_test "get_copy_replace_items returns expected files"

    local items
    items=$(get_copy_replace_items)

    if echo "$items" | grep -q "COMMAND_REGISTRY.md"; then
        test_pass "COMMAND_REGISTRY.md in copy-replace"
    else
        test_fail "COMMAND_REGISTRY.md in copy-replace" "present" "missing"
    fi

    if echo "$items" | grep -q "forge-workflow.yaml"; then
        test_pass "forge-workflow.yaml in copy-replace"
    else
        test_fail "forge-workflow.yaml in copy-replace" "present" "missing"
    fi
}

test_merge_items() {
    run_test "get_merge_items returns items with strategies"

    local items
    items=$(get_merge_items)

    if echo "$items" | grep -q "settings.local.json:merge-settings"; then
        test_pass "settings.local.json with merge-settings strategy"
    else
        test_fail "settings.local.json:merge-settings" "present" "missing"
    fi

    if echo "$items" | grep -q "CLAUDE.md:merge-docs"; then
        test_pass "CLAUDE.md with merge-docs strategy"
    else
        test_fail "CLAUDE.md:merge-docs" "present" "missing"
    fi
}

test_ignore_items() {
    run_test "get_ignore_items returns expected items"

    local items
    items=$(get_ignore_items)

    if echo "$items" | grep -q "ACTIVE_RITE"; then
        test_pass "ACTIVE_RITE in ignore list"
    else
        test_fail "ACTIVE_RITE in ignore" "present" "missing"
    fi

    if echo "$items" | grep -q "sessions"; then
        test_pass "sessions in ignore list"
    else
        test_fail "sessions in ignore" "present" "missing"
    fi

    if echo "$items" | grep -q ".cem"; then
        test_pass ".cem in ignore list"
    else
        test_fail ".cem in ignore" "present" "missing"
    fi
}

test_is_ignored() {
    run_test "is_ignored function"

    if is_ignored "ACTIVE_RITE"; then
        test_pass "ACTIVE_RITE is ignored"
    else
        test_fail "is_ignored ACTIVE_RITE" "true" "false"
    fi

    if is_ignored "sessions"; then
        test_pass "sessions is ignored"
    else
        test_fail "is_ignored sessions" "true" "false"
    fi

    if ! is_ignored "COMMAND_REGISTRY.md"; then
        test_pass "COMMAND_REGISTRY.md is not ignored"
    else
        test_fail "is_ignored COMMAND_REGISTRY.md" "false" "true"
    fi
}

test_get_merge_strategy() {
    run_test "get_merge_strategy function"

    local strategy

    strategy=$(get_merge_strategy "settings.local.json")
    if [[ "$strategy" == "merge-settings" ]]; then
        test_pass "settings.local.json -> merge-settings"
    else
        test_fail "get_merge_strategy settings.local.json" "merge-settings" "$strategy"
    fi

    strategy=$(get_merge_strategy "CLAUDE.md")
    if [[ "$strategy" == "merge-docs" ]]; then
        test_pass "CLAUDE.md -> merge-docs"
    else
        test_fail "get_merge_strategy CLAUDE.md" "merge-docs" "$strategy"
    fi

    strategy=$(get_merge_strategy "nonexistent.txt" || true)
    if [[ -z "$strategy" ]]; then
        test_pass "nonexistent file returns empty"
    else
        test_fail "get_merge_strategy nonexistent" "empty" "$strategy"
    fi
}

test_is_copy_replace() {
    run_test "is_copy_replace function"

    if is_copy_replace "COMMAND_REGISTRY.md"; then
        test_pass "COMMAND_REGISTRY.md is copy-replace"
    else
        test_fail "is_copy_replace COMMAND_REGISTRY.md" "true" "false"
    fi

    if ! is_copy_replace "CLAUDE.md"; then
        test_pass "CLAUDE.md is not copy-replace"
    else
        test_fail "is_copy_replace CLAUDE.md" "false" "true"
    fi
}

test_get_all_managed_files() {
    run_test "get_all_managed_files function"

    local files
    files=$(get_all_managed_files)

    if echo "$files" | grep -q "COMMAND_REGISTRY.md"; then
        test_pass "includes COMMAND_REGISTRY.md"
    else
        test_fail "get_all_managed_files" "includes COMMAND_REGISTRY.md" "missing"
    fi

    if echo "$files" | grep -q "settings.local.json"; then
        test_pass "includes settings.local.json"
    else
        test_fail "get_all_managed_files" "includes settings.local.json" "missing"
    fi

    if echo "$files" | grep -q "CLAUDE.md"; then
        test_pass "includes CLAUDE.md"
    else
        test_fail "get_all_managed_files" "includes CLAUDE.md" "missing"
    fi
}

test_schema_constants() {
    run_test "schema constants defined"

    if [[ "$SYNC_SCHEMA_VERSION" == "3" ]]; then
        test_pass "SYNC_SCHEMA_VERSION=3"
    else
        test_fail "SYNC_SCHEMA_VERSION" "3" "$SYNC_SCHEMA_VERSION"
    fi

    if [[ -n "$SYNC_MANIFEST_FILE" ]]; then
        test_pass "SYNC_MANIFEST_FILE defined"
    else
        test_fail "SYNC_MANIFEST_FILE" "defined" "empty"
    fi
}

# ============================================================================
# Run Tests
# ============================================================================

echo "=========================================="
echo "Running sync-config.sh tests"
echo "=========================================="
echo ""

test_exit_codes_defined
test_copy_replace_items
test_merge_items
test_ignore_items
test_is_ignored
test_get_merge_strategy
test_is_copy_replace
test_get_all_managed_files
test_schema_constants

echo ""
echo "=========================================="
echo "Results: $TESTS_PASSED/$TESTS_RUN passed"
if [[ $TESTS_FAILED -gt 0 ]]; then
    echo "FAILED: $TESTS_FAILED tests"
    exit 1
else
    echo "All tests passed!"
    exit 0
fi

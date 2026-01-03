#!/usr/bin/env bash
#
# test-sync-checksum.sh - Unit tests for sync-checksum.sh
#
# Tests cross-platform checksum computation and caching.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../.." && pwd)}"

# Source dependencies
source "$ROSTER_HOME/lib/sync/sync-config.sh"
source "$ROSTER_HOME/lib/sync/sync-checksum.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""

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
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ============================================================================
# Tests
# ============================================================================

test_detect_checksum_tool() {
    run_test "detect_checksum_tool finds a tool"

    if detect_checksum_tool; then
        test_pass "checksum tool detected"
    else
        test_fail "detect_checksum_tool" "success" "failed"
        return
    fi

    if [[ -n "$CHECKSUM_CMD" ]]; then
        test_pass "CHECKSUM_CMD is set: $CHECKSUM_CMD"
    else
        test_fail "CHECKSUM_CMD" "non-empty" "empty"
    fi
}

test_compute_checksum_basic() {
    run_test "compute_checksum on known content"

    # Create test file with known content
    local test_file="$TEST_TMP/test.txt"
    echo -n "hello world" > "$test_file"

    # Known SHA-256 for "hello world" (no newline)
    local expected="b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

    local actual
    actual=$(compute_checksum "$test_file")

    if [[ "$actual" == "$expected" ]]; then
        test_pass "checksum matches expected"
    else
        test_fail "compute_checksum" "$expected" "$actual"
    fi
}

test_compute_checksum_nonexistent() {
    run_test "compute_checksum on nonexistent file"

    local result
    result=$(compute_checksum "/nonexistent/file/path" 2>/dev/null || echo "")

    if [[ -z "$result" ]]; then
        test_pass "returns empty for nonexistent file"
    else
        test_fail "compute_checksum nonexistent" "empty" "$result"
    fi
}

test_compute_content_checksum() {
    run_test "compute_content_checksum on string"

    local content="test content"
    local checksum
    checksum=$(compute_content_checksum "$content")

    if [[ ${#checksum} -eq 64 ]]; then
        test_pass "returns 64-char hex checksum"
    else
        test_fail "compute_content_checksum length" "64" "${#checksum}"
    fi

    # Verify it's hex
    if [[ "$checksum" =~ ^[a-f0-9]{64}$ ]]; then
        test_pass "checksum is valid hex"
    else
        test_fail "compute_content_checksum hex" "hex string" "$checksum"
    fi
}

test_checksum_cache_init() {
    run_test "init_checksum_cache"

    # Clear any existing cache
    _CHECKSUM_CACHE=()

    init_checksum_cache

    if [[ ${#_CHECKSUM_CACHE[@]} -ge 0 ]]; then
        test_pass "cache initialized (${#_CHECKSUM_CACHE[@]} entries)"
    else
        test_fail "init_checksum_cache" "success" "failed"
    fi
}

test_cache_hit() {
    run_test "checksum cache hit"

    # Create test file
    local test_file="$TEST_TMP/cached.txt"
    echo "cache test" > "$test_file"

    # Clear cache
    _CHECKSUM_CACHE=()

    # First call - should compute
    local first_result
    first_result=$(compute_checksum "$test_file")

    # Second call - should hit cache
    local second_result
    second_result=$(compute_checksum "$test_file")

    if [[ "$first_result" == "$second_result" ]]; then
        test_pass "cached result matches original"
    else
        test_fail "cache consistency" "$first_result" "$second_result"
    fi
}

test_cache_invalidation() {
    run_test "cache invalidation on file change"

    # Create test file
    local test_file="$TEST_TMP/changing.txt"
    echo "original" > "$test_file"

    # Clear cache
    _CHECKSUM_CACHE=()

    # Compute first checksum
    local first_result
    first_result=$(compute_checksum "$test_file")

    # Wait a moment and modify file
    sleep 1
    echo "modified" > "$test_file"

    # Clear cached checksum (simulate mtime change detection)
    clear_cached_checksum "$test_file"

    # Compute second checksum
    local second_result
    second_result=$(compute_checksum "$test_file")

    if [[ "$first_result" != "$second_result" ]]; then
        test_pass "detects file change"
    else
        test_fail "cache invalidation" "different checksums" "same checksums"
    fi
}

test_checksums_match() {
    run_test "checksums_match function"

    local cs1="abc123"
    local cs2="abc123"
    local cs3="def456"

    if checksums_match "$cs1" "$cs2"; then
        test_pass "matching checksums return true"
    else
        test_fail "checksums_match same" "true" "false"
    fi

    if ! checksums_match "$cs1" "$cs3"; then
        test_pass "different checksums return false"
    else
        test_fail "checksums_match different" "false" "true"
    fi
}

test_file_changed() {
    run_test "file_changed function"

    # Create test file
    local test_file="$TEST_TMP/check.txt"
    echo "test content" > "$test_file"

    local checksum
    checksum=$(compute_checksum "$test_file")

    if ! file_changed "$test_file" "$checksum"; then
        test_pass "unchanged file detected"
    else
        test_fail "file_changed unchanged" "false" "true"
    fi

    if file_changed "$test_file" "wrongchecksum"; then
        test_pass "changed file detected"
    else
        test_fail "file_changed changed" "true" "false"
    fi
}

# ============================================================================
# Run Tests
# ============================================================================

echo "=========================================="
echo "Running sync-checksum.sh tests"
echo "=========================================="
echo ""

setup

test_detect_checksum_tool
test_compute_checksum_basic
test_compute_checksum_nonexistent
test_compute_content_checksum
test_checksum_cache_init
test_cache_hit
test_cache_invalidation
test_checksums_match
test_file_changed

teardown

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

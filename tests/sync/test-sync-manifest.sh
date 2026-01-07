#!/usr/bin/env bash
#
# test-sync-manifest.sh - Unit tests for sync-manifest.sh
#
# Tests manifest reading, writing, and migration.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../.." && pwd)}"

# Source dependencies
source "$ROSTER_HOME/lib/sync/sync-config.sh"
source "$ROSTER_HOME/lib/sync/sync-checksum.sh"
source "$ROSTER_HOME/lib/sync/sync-manifest.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""
ORIGINAL_MANIFEST_FILE=""

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
    # Create test manifest location
    mkdir -p "$TEST_TMP/.cem"
    # Override the SYNC_MANIFEST_FILE variable (unset first since it may be readonly)
    export SYNC_MANIFEST_FILE="$TEST_TMP/.cem/manifest.json"
    echo "Test temp dir: $TEST_TMP"
    echo "Test manifest: $SYNC_MANIFEST_FILE"
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ============================================================================
# Tests
# ============================================================================

test_create_manifest() {
    run_test "create_manifest function"

    local manifest
    manifest=$(create_manifest "/path/to/roster" "abc123" "main")

    local version
    version=$(echo "$manifest" | jq -r '.schema_version')
    if [[ "$version" == "3" ]]; then
        test_pass "schema_version is 3"
    else
        test_fail "schema_version" "3" "$version"
    fi

    local path
    path=$(echo "$manifest" | jq -r '.roster.path')
    if [[ "$path" == "/path/to/roster" ]]; then
        test_pass "roster.path set correctly"
    else
        test_fail "roster.path" "/path/to/roster" "$path"
    fi

    local commit
    commit=$(echo "$manifest" | jq -r '.roster.commit')
    if [[ "$commit" == "abc123" ]]; then
        test_pass "roster.commit set correctly"
    else
        test_fail "roster.commit" "abc123" "$commit"
    fi
}

test_write_read_manifest() {
    run_test "write and read manifest"

    local manifest
    manifest=$(create_manifest "/test/roster" "def456" "main")

    write_manifest "$manifest"

    if [[ -f "$SYNC_MANIFEST_FILE" ]]; then
        test_pass "manifest file created"
    else
        test_fail "manifest file" "exists" "missing"
        return
    fi

    local read_manifest
    read_manifest=$(read_manifest)

    local version
    version=$(echo "$read_manifest" | jq -r '.schema_version')
    if [[ "$version" == "3" ]]; then
        test_pass "read manifest has correct version"
    else
        test_fail "read schema_version" "3" "$version"
    fi
}

test_add_managed_file() {
    run_test "add_managed_file function"

    local manifest
    manifest=$(create_manifest "/test" "abc" "main")

    manifest=$(add_managed_file "$manifest" ".claude/test.md" "copy-replace" "checksum123")

    local count
    count=$(echo "$manifest" | jq '.managed_files | length')
    if [[ "$count" == "1" ]]; then
        test_pass "managed_files has 1 entry"
    else
        test_fail "managed_files count" "1" "$count"
    fi

    local path
    path=$(echo "$manifest" | jq -r '.managed_files[0].path')
    if [[ "$path" == ".claude/test.md" ]]; then
        test_pass "path set correctly"
    else
        test_fail "path" ".claude/test.md" "$path"
    fi

    local strategy
    strategy=$(echo "$manifest" | jq -r '.managed_files[0].strategy')
    if [[ "$strategy" == "copy-replace" ]]; then
        test_pass "strategy set correctly"
    else
        test_fail "strategy" "copy-replace" "$strategy"
    fi
}

test_update_manifest_roster() {
    run_test "update_manifest_roster function"

    local manifest
    manifest=$(create_manifest "/test" "old_commit" "main")
    manifest=$(update_manifest_roster "$manifest" "new_commit" "feature")

    local commit
    commit=$(echo "$manifest" | jq -r '.roster.commit')
    if [[ "$commit" == "new_commit" ]]; then
        test_pass "commit updated"
    else
        test_fail "commit" "new_commit" "$commit"
    fi

    local ref
    ref=$(echo "$manifest" | jq -r '.roster.ref')
    if [[ "$ref" == "feature" ]]; then
        test_pass "ref updated"
    else
        test_fail "ref" "feature" "$ref"
    fi
}

test_get_manifest_version() {
    run_test "get_manifest_version function"

    # Test with v3 manifest
    local manifest
    manifest='{"schema_version": 3}'
    echo "$manifest" > "$SYNC_MANIFEST_FILE"

    local version
    version=$(get_manifest_version)
    if [[ "$version" == "3" ]]; then
        test_pass "detects v3"
    else
        test_fail "detect v3" "3" "$version"
    fi

    # Test with v1 manifest (no schema_version)
    manifest='{"skeleton_path": "/test"}'
    echo "$manifest" > "$SYNC_MANIFEST_FILE"

    version=$(get_manifest_version)
    if [[ "$version" == "1" ]]; then
        test_pass "defaults to v1 when missing"
    else
        test_fail "default to v1" "1" "$version"
    fi
}

test_migrate_v1_to_v3() {
    run_test "migrate v1 to v3"

    # Create v1 manifest
    cat > "$SYNC_MANIFEST_FILE" <<'EOF'
{
    "skeleton_path": "/old/skeleton",
    "skeleton_commit": "old123",
    "skeleton_ref": "master",
    "last_sync": "2025-01-01T00:00:00Z",
    "managed_files": [
        {"path": "test.md", "checksum": "abc"}
    ]
}
EOF

    migrate_v1_to_v3 "$SYNC_MANIFEST_FILE"

    # Check backup created
    if [[ -f "${SYNC_MANIFEST_FILE}.v1.backup" ]]; then
        test_pass "v1 backup created"
    else
        test_fail "v1 backup" "exists" "missing"
    fi

    # Check migrated version
    local manifest
    manifest=$(cat "$SYNC_MANIFEST_FILE")

    local version
    version=$(echo "$manifest" | jq -r '.schema_version')
    if [[ "$version" == "3" ]]; then
        test_pass "schema_version updated to 3"
    else
        test_fail "schema_version" "3" "$version"
    fi

    local roster_path
    roster_path=$(echo "$manifest" | jq -r '.roster.path')
    if [[ "$roster_path" == "/old/skeleton" ]]; then
        test_pass "roster.path migrated from skeleton_path"
    else
        test_fail "roster.path" "/old/skeleton" "$roster_path"
    fi

    local skeleton_path
    skeleton_path=$(echo "$manifest" | jq -r '.migration.skeleton_path')
    if [[ "$skeleton_path" == "/old/skeleton" ]]; then
        test_pass "migration.skeleton_path preserved"
    else
        test_fail "migration.skeleton_path" "/old/skeleton" "$skeleton_path"
    fi
}

test_migrate_v2_to_v3() {
    run_test "migrate v2 to v3"

    # Create v2 manifest
    cat > "$SYNC_MANIFEST_FILE" <<'EOF'
{
    "schema_version": 2,
    "skeleton": {
        "path": "/old/skeleton",
        "commit": "old123",
        "ref": "master",
        "last_sync": "2025-01-01T00:00:00Z"
    },
    "rite": {
        "name": "10x-dev"
    },
    "managed_files": []
}
EOF

    migrate_v2_to_v3 "$SYNC_MANIFEST_FILE"

    # Check backup created
    if [[ -f "${SYNC_MANIFEST_FILE}.v2.backup" ]]; then
        test_pass "v2 backup created"
    else
        test_fail "v2 backup" "exists" "missing"
    fi

    # Check migrated version
    local manifest
    manifest=$(cat "$SYNC_MANIFEST_FILE")

    local version
    version=$(echo "$manifest" | jq -r '.schema_version')
    if [[ "$version" == "3" ]]; then
        test_pass "schema_version updated to 3"
    else
        test_fail "schema_version" "3" "$version"
    fi

    local roster_path
    roster_path=$(echo "$manifest" | jq -r '.roster.path')
    if [[ "$roster_path" == "/old/skeleton" ]]; then
        test_pass "roster.path migrated from skeleton.path"
    else
        test_fail "roster.path" "/old/skeleton" "$roster_path"
    fi

    local rite_name
    rite_name=$(echo "$manifest" | jq -r '.rite.name')
    if [[ "$rite_name" == "10x-dev" ]]; then
        test_pass "rite preserved"
    else
        test_fail "rite.name" "10x-dev" "$rite_name"
    fi

    local migrated_from
    migrated_from=$(echo "$manifest" | jq -r '.migration.migrated_from')
    if [[ "$migrated_from" == "2" ]]; then
        test_pass "migration.migrated_from set to 2"
    else
        test_fail "migration.migrated_from" "2" "$migrated_from"
    fi
}

test_update_manifest_rite() {
    run_test "update_manifest_rite function"

    local manifest
    manifest=$(create_manifest "/test" "abc" "main")
    manifest=$(update_manifest_rite "$manifest" "custom-rite" "ritehash")

    local rite_name
    rite_name=$(echo "$manifest" | jq -r '.rite.name')
    if [[ "$rite_name" == "custom-rite" ]]; then
        test_pass "rite.name set"
    else
        test_fail "rite.name" "custom-rite" "$rite_name"
    fi

    local rite_checksum
    rite_checksum=$(echo "$manifest" | jq -r '.rite.checksum')
    if [[ "$rite_checksum" == "ritehash" ]]; then
        test_pass "rite.checksum set"
    else
        test_fail "rite.checksum" "ritehash" "$rite_checksum"
    fi
}

test_add_manifest_orphan() {
    run_test "add_manifest_orphan function"

    local manifest
    manifest=$(create_manifest "/test" "abc" "main")
    manifest=$(add_manifest_orphan "$manifest" ".claude/old.md" "removed")

    local count
    count=$(echo "$manifest" | jq '.orphans | length')
    if [[ "$count" == "1" ]]; then
        test_pass "orphans has 1 entry"
    else
        test_fail "orphans count" "1" "$count"
    fi

    local path
    path=$(echo "$manifest" | jq -r '.orphans[0].path')
    if [[ "$path" == ".claude/old.md" ]]; then
        test_pass "orphan path set"
    else
        test_fail "orphan path" ".claude/old.md" "$path"
    fi
}

# ============================================================================
# Run Tests
# ============================================================================

echo "=========================================="
echo "Running sync-manifest.sh tests"
echo "=========================================="
echo ""

setup

test_create_manifest
test_write_read_manifest
test_add_managed_file
test_update_manifest_roster
test_get_manifest_version
test_migrate_v1_to_v3
test_migrate_v2_to_v3
test_update_manifest_rite
test_add_manifest_orphan

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

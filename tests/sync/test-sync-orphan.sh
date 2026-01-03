#!/usr/bin/env bash
#
# test-sync-orphan.sh - Unit tests for orphan detection and management
#
# Tests orphan detection, backup creation, and pruning functionality.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../.." && pwd)}"

# Source dependencies
source "$ROSTER_HOME/lib/sync/sync-config.sh"
source "$ROSTER_HOME/lib/sync/sync-checksum.sh"
source "$ROSTER_HOME/lib/sync/sync-manifest.sh"
source "$ROSTER_HOME/lib/sync/sync-core.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""
TEST_ROSTER_DIR=""
TEST_PROJECT_DIR=""

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
    TEST_ROSTER_DIR="$TEST_TMP/roster"
    TEST_PROJECT_DIR="$TEST_TMP/project"

    # Create roster .claude directory (source)
    mkdir -p "$TEST_ROSTER_DIR/.claude"

    # Create project .claude directory (satellite)
    mkdir -p "$TEST_PROJECT_DIR/.claude/.cem"

    # Override the SYNC_MANIFEST_FILE variable
    export SYNC_MANIFEST_FILE="$TEST_PROJECT_DIR/.claude/.cem/manifest.json"
    export SYNC_ORPHAN_BACKUP_DIR="$TEST_PROJECT_DIR/.claude/.cem/orphan-backup"
    export SYNC_CHECKSUM_CACHE="$TEST_PROJECT_DIR/.claude/.cem/checksum-cache.json"

    # Initialize checksum cache
    init_checksum_cache

    # Change to project directory for relative paths
    cd "$TEST_PROJECT_DIR"

    echo "Test temp dir: $TEST_TMP"
    echo "Test roster: $TEST_ROSTER_DIR"
    echo "Test project: $TEST_PROJECT_DIR"
}

teardown() {
    cd /
    rm -rf "$TEST_TMP"
}

create_test_manifest() {
    local manifest
    manifest=$(create_manifest "$TEST_ROSTER_DIR" "test123" "main")
    echo "$manifest"
}

# Clean test state between tests
clean_test_state() {
    # Remove all files in roster .claude/
    rm -rf "$TEST_ROSTER_DIR/.claude/"*
    # Remove all files in project .claude/ except .cem
    find "$TEST_PROJECT_DIR/.claude" -maxdepth 1 -type f -delete 2>/dev/null || true
    find "$TEST_PROJECT_DIR/.claude" -maxdepth 1 -type d ! -name ".claude" ! -name ".cem" -exec rm -rf {} + 2>/dev/null || true
    # Clear manifest
    rm -f "$SYNC_MANIFEST_FILE"
    # Clear backup directory
    rm -rf "$SYNC_ORPHAN_BACKUP_DIR"
}

# ============================================================================
# Tests
# ============================================================================

test_detect_orphans_no_orphans() {
    run_test "detect_orphans with no orphans"
    clean_test_state

    # Create file in both roster and project
    echo "content" > "$TEST_ROSTER_DIR/.claude/COMMAND_REGISTRY.md"
    echo "content" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Detect orphans
    local orphans
    orphans=$(detect_orphans "$TEST_ROSTER_DIR/.claude")

    if [[ -z "$orphans" ]]; then
        test_pass "no orphans detected when file exists in roster"
    else
        test_fail "no orphans" "empty" "$orphans"
    fi
}

test_detect_orphans_finds_orphan() {
    run_test "detect_orphans finds orphan when file removed from roster"
    clean_test_state

    # Create file only in project (not in roster)
    echo "content" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Detect orphans (file not in roster)
    local orphans
    orphans=$(detect_orphans "$TEST_ROSTER_DIR/.claude")

    if [[ "$orphans" == ".claude/COMMAND_REGISTRY.md" ]]; then
        test_pass "orphan detected when file missing from roster"
    else
        test_fail "orphan detection" ".claude/COMMAND_REGISTRY.md" "$orphans"
    fi
}

test_detect_orphans_ignores_non_managed() {
    run_test "detect_orphans ignores non-managed files"
    clean_test_state

    # Create file only in project that uses a non-managed strategy
    echo "content" > "$TEST_PROJECT_DIR/.claude/custom-file.txt"

    # Create manifest with the file using a custom/unknown strategy
    # Files with unknown strategies are not considered orphans
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/custom-file.txt" "satellite-only" "abc123")
    write_manifest "$manifest"

    # Detect orphans
    local orphans
    orphans=$(detect_orphans "$TEST_ROSTER_DIR/.claude")

    # custom-file.txt with "satellite-only" strategy is not an orphan (not a managed strategy)
    if [[ -z "$orphans" ]]; then
        test_pass "non-managed files not reported as orphans"
    else
        test_fail "non-managed filter" "empty" "$orphans"
    fi
}

test_detect_orphans_multiple() {
    run_test "detect_orphans finds multiple orphans"
    clean_test_state

    # Create files only in project
    echo "content1" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"
    echo "content2" > "$TEST_PROJECT_DIR/.claude/forge-workflow.yaml"

    # Create manifest with both files
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    manifest=$(add_managed_file "$manifest" ".claude/forge-workflow.yaml" "copy-replace" "def456")
    write_manifest "$manifest"

    # Detect orphans
    local orphans count
    orphans=$(detect_orphans "$TEST_ROSTER_DIR/.claude")
    count=$(echo "$orphans" | grep -c "." || true)

    if [[ "$count" -eq 2 ]]; then
        test_pass "multiple orphans detected"
    else
        test_fail "multiple orphans" "2" "$count"
    fi
}

test_backup_orphans_creates_timestamped_dir() {
    run_test "backup_orphans creates timestamped directory"
    clean_test_state

    # Create orphan file
    echo "orphan content" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Backup the orphan (capture only last line which is the directory path)
    local backup_output backup_dir
    backup_output=$(echo ".claude/COMMAND_REGISTRY.md" | backup_orphans 2>&1)
    backup_dir=$(echo "$backup_output" | tail -1)

    if [[ -n "$backup_dir" && -d "$backup_dir" ]]; then
        test_pass "backup directory created"
    else
        test_fail "backup directory" "exists" "missing: $backup_dir"
        return
    fi

    # Check directory name format (YYYYMMDD-HHMMSS)
    local dir_name
    dir_name=$(basename "$backup_dir")
    if [[ "$dir_name" =~ ^[0-9]{8}-[0-9]{6}$ ]]; then
        test_pass "backup directory has timestamp format"
    else
        test_fail "backup format" "YYYYMMDD-HHMMSS" "$dir_name"
    fi
}

test_backup_orphans_preserves_content() {
    run_test "backup_orphans preserves file content"
    clean_test_state

    # Create orphan file with specific content
    echo "unique orphan content xyz123" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Backup the orphan (capture only last line which is the directory path)
    local backup_output backup_dir
    backup_output=$(echo ".claude/COMMAND_REGISTRY.md" | backup_orphans 2>&1)
    backup_dir=$(echo "$backup_output" | tail -1)

    # Check backup file exists and has correct content
    local backup_file="$backup_dir/COMMAND_REGISTRY.md"
    if [[ -f "$backup_file" ]]; then
        local content
        content=$(cat "$backup_file")
        if [[ "$content" == "unique orphan content xyz123" ]]; then
            test_pass "backup content preserved"
        else
            test_fail "backup content" "unique orphan content xyz123" "$content"
        fi
    else
        test_fail "backup file" "exists" "missing: $backup_file"
    fi
}

test_backup_orphans_preserves_relative_path() {
    run_test "backup_orphans preserves relative path structure"
    clean_test_state

    # Create orphan file in subdirectory
    mkdir -p "$TEST_PROJECT_DIR/.claude/subdir"
    echo "nested content" > "$TEST_PROJECT_DIR/.claude/subdir/nested.md"

    # Backup the orphan (capture only last line which is the directory path)
    local backup_output backup_dir
    backup_output=$(echo ".claude/subdir/nested.md" | backup_orphans 2>&1)
    backup_dir=$(echo "$backup_output" | tail -1)

    # Check backup preserves path structure
    local backup_file="$backup_dir/subdir/nested.md"
    if [[ -f "$backup_file" ]]; then
        test_pass "relative path structure preserved in backup"
    else
        test_fail "nested backup" "exists at $backup_file" "missing"
    fi
}

test_backup_orphans_no_files() {
    run_test "backup_orphans handles no orphans gracefully"
    clean_test_state

    # Backup with empty input
    local backup_dir
    backup_dir=$(echo "" | backup_orphans)

    if [[ -z "$backup_dir" ]]; then
        test_pass "no backup created for empty input"
    else
        test_fail "empty backup" "no output" "$backup_dir"
    fi
}

test_prune_orphans_removes_file() {
    run_test "prune_orphans removes orphan file"
    clean_test_state

    # Create orphan file
    echo "to be pruned" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Prune the orphan
    echo ".claude/COMMAND_REGISTRY.md" | prune_orphans

    if [[ ! -f "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md" ]]; then
        test_pass "orphan file removed"
    else
        test_fail "file removal" "file deleted" "file still exists"
    fi
}

test_prune_orphans_updates_manifest() {
    run_test "prune_orphans removes entry from manifest"
    clean_test_state

    # Create orphan file
    echo "to be pruned" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Prune the orphan
    echo ".claude/COMMAND_REGISTRY.md" | prune_orphans

    # Check manifest no longer has the file
    manifest=$(read_manifest)
    local count
    count=$(echo "$manifest" | jq '.managed_files | length')

    if [[ "$count" == "0" ]]; then
        test_pass "orphan removed from manifest"
    else
        test_fail "manifest update" "0 managed files" "$count"
    fi
}

test_prune_orphans_tracks_removal() {
    run_test "prune_orphans tracks removal in orphans array"
    clean_test_state

    # Create orphan file
    echo "to be pruned" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Prune the orphan
    echo ".claude/COMMAND_REGISTRY.md" | prune_orphans

    # Check manifest has orphan tracking entry
    manifest=$(read_manifest)
    local orphan_count
    orphan_count=$(echo "$manifest" | jq '.orphans | length')

    if [[ "$orphan_count" == "1" ]]; then
        test_pass "orphan removal tracked"
    else
        test_fail "orphan tracking" "1 orphan entry" "$orphan_count"
    fi
}

test_has_orphans_returns_true() {
    run_test "has_orphans returns true when orphans exist"
    clean_test_state

    # Create orphan file
    echo "orphan" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Check has_orphans (file not in roster)
    if has_orphans "$TEST_ROSTER_DIR/.claude"; then
        test_pass "has_orphans returns true"
    else
        test_fail "has_orphans" "true" "false"
    fi
}

test_has_orphans_returns_false() {
    run_test "has_orphans returns false when no orphans"
    clean_test_state

    # Create file in both roster and project
    echo "content" > "$TEST_ROSTER_DIR/.claude/COMMAND_REGISTRY.md"
    echo "content" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Check has_orphans
    if ! has_orphans "$TEST_ROSTER_DIR/.claude"; then
        test_pass "has_orphans returns false"
    else
        test_fail "has_orphans" "false" "true"
    fi
}

test_handle_orphans_reports_without_prune() {
    run_test "handle_orphans reports orphans without --prune"
    clean_test_state

    # Create orphan file
    echo "orphan" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Handle orphans without prune flag
    local exit_code=0
    handle_orphans "$TEST_ROSTER_DIR/.claude" "0" "0" || exit_code=$?

    if [[ "$exit_code" == "$EXIT_SYNC_ORPHAN_CONFLICTS" ]]; then
        test_pass "returns EXIT_SYNC_ORPHAN_CONFLICTS without prune"
    else
        test_fail "exit code" "$EXIT_SYNC_ORPHAN_CONFLICTS" "$exit_code"
    fi

    # File should still exist
    if [[ -f "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md" ]]; then
        test_pass "file preserved without prune"
    else
        test_fail "file preservation" "file exists" "file removed"
    fi
}

test_handle_orphans_prunes_with_flag() {
    run_test "handle_orphans prunes orphans with --prune"
    clean_test_state

    # Create orphan file
    echo "orphan content" > "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md"

    # Create manifest with the file
    local manifest
    manifest=$(create_test_manifest)
    manifest=$(add_managed_file "$manifest" ".claude/COMMAND_REGISTRY.md" "copy-replace" "abc123")
    write_manifest "$manifest"

    # Handle orphans with prune flag
    local exit_code=0
    handle_orphans "$TEST_ROSTER_DIR/.claude" "1" "0" || exit_code=$?

    if [[ "$exit_code" == "$EXIT_SYNC_SUCCESS" ]]; then
        test_pass "returns EXIT_SYNC_SUCCESS with prune"
    else
        test_fail "exit code" "$EXIT_SYNC_SUCCESS" "$exit_code"
    fi

    # File should be removed
    if [[ ! -f "$TEST_PROJECT_DIR/.claude/COMMAND_REGISTRY.md" ]]; then
        test_pass "file removed with prune"
    else
        test_fail "file removal" "file removed" "file still exists"
    fi

    # Backup should exist
    local backup_count
    backup_count=$(find "$SYNC_ORPHAN_BACKUP_DIR" -type f -name "COMMAND_REGISTRY.md" 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$backup_count" -eq 1 ]]; then
        test_pass "backup created before prune"
    else
        test_fail "backup creation" "1 backup" "$backup_count"
    fi
}

test_detect_untracked_finds_files() {
    run_test "detect_untracked finds files not in manifest"
    clean_test_state

    # Create file not in manifest
    echo "untracked" > "$TEST_PROJECT_DIR/.claude/untracked-file.md"

    # Create empty manifest
    local manifest
    manifest=$(create_test_manifest)
    write_manifest "$manifest"

    # Detect untracked
    local untracked
    untracked=$(detect_untracked ".claude")

    if echo "$untracked" | grep -q "untracked-file.md"; then
        test_pass "untracked file detected"
    else
        test_fail "untracked detection" "untracked-file.md" "$untracked"
    fi
}

test_detect_untracked_excludes_ignored() {
    run_test "detect_untracked excludes ignored directories"
    clean_test_state

    # Create files in ignored directories
    mkdir -p "$TEST_PROJECT_DIR/.claude/sessions"
    echo "session data" > "$TEST_PROJECT_DIR/.claude/sessions/session.json"

    mkdir -p "$TEST_PROJECT_DIR/.claude/.cem"
    echo "manifest" > "$TEST_PROJECT_DIR/.claude/.cem/manifest.json"

    # Create empty manifest (but with those dirs in ignore list)
    local manifest
    manifest=$(create_test_manifest)
    write_manifest "$manifest"

    # Detect untracked
    local untracked
    untracked=$(detect_untracked ".claude")

    # Should not include ignored directories
    if ! echo "$untracked" | grep -q "sessions"; then
        test_pass "sessions directory excluded"
    else
        test_fail "ignore sessions" "excluded" "included"
    fi

    if ! echo "$untracked" | grep -q ".cem"; then
        test_pass ".cem directory excluded"
    else
        test_fail "ignore .cem" "excluded" "included"
    fi
}

test_get_local_claude_files() {
    run_test "get_local_claude_files lists files correctly"
    clean_test_state

    # Create various files
    echo "file1" > "$TEST_PROJECT_DIR/.claude/file1.md"
    echo "file2" > "$TEST_PROJECT_DIR/.claude/file2.yaml"

    # Create ignored files
    mkdir -p "$TEST_PROJECT_DIR/.claude/.cem"
    echo "ignored" > "$TEST_PROJECT_DIR/.claude/.cem/cache.json"

    local files
    files=$(get_local_claude_files ".claude")

    if echo "$files" | grep -q "file1.md"; then
        test_pass "file1.md found"
    else
        test_fail "file1.md" "found" "missing"
    fi

    if echo "$files" | grep -q "file2.yaml"; then
        test_pass "file2.yaml found"
    else
        test_fail "file2.yaml" "found" "missing"
    fi

    if ! echo "$files" | grep -q ".cem"; then
        test_pass ".cem directory excluded"
    else
        test_fail ".cem exclusion" "excluded" "included"
    fi
}

# ============================================================================
# Run Tests
# ============================================================================

echo "=========================================="
echo "Running sync-orphan.sh tests"
echo "=========================================="
echo ""

setup

test_detect_orphans_no_orphans
test_detect_orphans_finds_orphan
test_detect_orphans_ignores_non_managed
test_detect_orphans_multiple
test_backup_orphans_creates_timestamped_dir
test_backup_orphans_preserves_content
test_backup_orphans_preserves_relative_path
test_backup_orphans_no_files
test_prune_orphans_removes_file
test_prune_orphans_updates_manifest
test_prune_orphans_tracks_removal
test_has_orphans_returns_true
test_has_orphans_returns_false
test_handle_orphans_reports_without_prune
test_handle_orphans_prunes_with_flag
test_detect_untracked_finds_files
test_detect_untracked_excludes_ignored
test_get_local_claude_files

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

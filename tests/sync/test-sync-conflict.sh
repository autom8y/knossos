#!/usr/bin/env bash
#
# test-sync-conflict.sh - Unit tests for conflict resolution (TDD 4.3)
#
# Tests three-way classification, conflict detection, backup creation,
# and conflict resolution with --force flag.
#
# Part of: knossos-sync (TDD-cem-replacement task-015)

# Note: We use -uo pipefail but NOT -e because we handle test failures explicitly
set -uo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KNOSSOS_HOME="${KNOSSOS_HOME:-$(cd "$SCRIPT_DIR/../.." && pwd)}"

# Source dependencies in order
source "$KNOSSOS_HOME/lib/sync/sync-config.sh"
source "$KNOSSOS_HOME/lib/sync/sync-checksum.sh"
source "$KNOSSOS_HOME/lib/sync/sync-manifest.sh"
source "$KNOSSOS_HOME/lib/sync/sync-core.sh"

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

    # Create test directory structure
    mkdir -p "$TEST_TMP/knossos/.claude"
    mkdir -p "$TEST_TMP/local/.claude/.cem"

    # Override manifest location for tests
    export SYNC_MANIFEST_FILE="$TEST_TMP/local/.claude/.cem/manifest.json"
    export SYNC_CHECKSUM_CACHE="$TEST_TMP/local/.claude/.cem/checksum-cache.json"

    # Reset conflict tracking globals
    _CONFLICT_BACKUP_DIR=""
    _CONFLICT_COUNT=0
    _CONFLICT_FILES=()

    # Initialize checksum tool
    detect_checksum_tool
}

teardown() {
    rm -rf "$TEST_TMP"
}

# Helper to create a test manifest
create_test_manifest() {
    local knossos_path="$1"

    local manifest
    manifest=$(jq -n \
        --argjson sv 3 \
        --arg rp "$knossos_path" \
        --arg rc "test-commit-hash" \
        --arg rr "main" \
        --arg ts "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" '{
        schema_version: $sv,
        knossos: {
            path: $rp,
            commit: $rc,
            ref: $rr,
            last_sync: $ts
        },
        team: null,
        managed_files: [],
        orphans: []
    }')

    echo "$manifest" > "$SYNC_MANIFEST_FILE"
}

# Helper to add a file to manifest
add_file_to_manifest() {
    local path="$1"
    local checksum="$2"
    local strategy="${3:-copy-replace}"

    local manifest
    manifest=$(cat "$SYNC_MANIFEST_FILE")
    manifest=$(echo "$manifest" | jq \
        --arg p "$path" \
        --arg c "$checksum" \
        --arg s "$strategy" \
        --arg t "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" '
        .managed_files += [{
            path: $p,
            strategy: $s,
            checksum: $c,
            source: "knossos",
            added_at: $t,
            last_sync: $t
        }]')

    echo "$manifest" > "$SYNC_MANIFEST_FILE"
}

# ============================================================================
# Classification Tests (TDD 4.1, 4.2)
# ============================================================================

test_classify_skip_up_to_date() {
    run_test "classify_file: SKIP when up to date (no changes)"

    # Create identical files
    echo "content" > "$TEST_TMP/knossos/.claude/test.md"
    echo "content" > "$TEST_TMP/local/.claude/test.md"

    # Create manifest with matching checksum
    create_test_manifest "$TEST_TMP/knossos"
    local checksum
    checksum=$(compute_checksum "$TEST_TMP/knossos/.claude/test.md")
    add_file_to_manifest ".claude/test.md" "$checksum"

    local result
    result=$(classify_file "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    if [[ "$result" == "$CLASSIFY_SKIP" ]]; then
        test_pass "returns SKIP when no changes"
    else
        test_fail "classify_file up to date" "$CLASSIFY_SKIP" "$result"
    fi
}

test_classify_skip_preserve_local() {
    run_test "classify_file: SKIP when only local changed (preserve local)"

    # Knossos unchanged from manifest, local modified
    echo "original" > "$TEST_TMP/knossos/.claude/test.md"
    echo "local changes" > "$TEST_TMP/local/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    local knossos_checksum
    knossos_checksum=$(compute_checksum "$TEST_TMP/knossos/.claude/test.md")
    add_file_to_manifest ".claude/test.md" "$knossos_checksum"

    local result
    result=$(classify_file "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    if [[ "$result" == "$CLASSIFY_SKIP" ]]; then
        test_pass "returns SKIP to preserve local changes"
    else
        test_fail "classify_file preserve local" "$CLASSIFY_SKIP" "$result"
    fi
}

test_classify_update_knossos_changed() {
    run_test "classify_file: UPDATE when knossos changed, local unchanged"

    # Setup: local matches old manifest, knossos has new content
    echo "old content" > "$TEST_TMP/local/.claude/test.md"
    echo "new content" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    local old_checksum
    old_checksum=$(compute_checksum "$TEST_TMP/local/.claude/test.md")
    add_file_to_manifest ".claude/test.md" "$old_checksum"

    local result
    result=$(classify_file "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    if [[ "$result" == "$CLASSIFY_UPDATE" ]]; then
        test_pass "returns UPDATE when knossos changed"
    else
        test_fail "classify_file knossos updated" "$CLASSIFY_UPDATE" "$result"
    fi
}

test_classify_conflict_both_changed() {
    run_test "classify_file: CONFLICT when both knossos and local changed"

    # Setup: manifest has original, both knossos and local have different changes
    echo "local changes" > "$TEST_TMP/local/.claude/test.md"
    echo "knossos changes" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    # Manifest checksum is different from both
    add_file_to_manifest ".claude/test.md" "fake-original-checksum-12345"

    local result
    result=$(classify_file "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    if [[ "$result" == "$CLASSIFY_CONFLICT" ]]; then
        test_pass "returns CONFLICT when both changed"
    else
        test_fail "classify_file both changed" "$CLASSIFY_CONFLICT" "$result"
    fi
}

test_classify_new_local_missing() {
    run_test "classify_file: NEW when local file doesn't exist"

    echo "knossos content" > "$TEST_TMP/knossos/.claude/new.md"

    create_test_manifest "$TEST_TMP/knossos"

    local result
    result=$(classify_file "new.md" "$TEST_TMP/knossos/.claude/new.md" "$TEST_TMP/local/.claude/new.md")

    if [[ "$result" == "$CLASSIFY_NEW" ]]; then
        test_pass "returns NEW when local missing"
    else
        test_fail "classify_file new file" "$CLASSIFY_NEW" "$result"
    fi
}

test_classify_conflict_no_manifest() {
    run_test "classify_file: CONFLICT when no manifest and files differ"

    echo "local content" > "$TEST_TMP/local/.claude/test.md"
    echo "knossos content" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    # No file entry in manifest

    local result
    result=$(classify_file "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    if [[ "$result" == "$CLASSIFY_CONFLICT" ]]; then
        test_pass "returns CONFLICT when no manifest and files differ"
    else
        test_fail "classify_file no manifest conflict" "$CLASSIFY_CONFLICT" "$result"
    fi
}

test_classify_skip_no_manifest_identical() {
    run_test "classify_file: SKIP when no manifest but files identical"

    echo "same content" > "$TEST_TMP/local/.claude/test.md"
    echo "same content" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    # No file entry in manifest

    local result
    result=$(classify_file "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    if [[ "$result" == "$CLASSIFY_SKIP" ]]; then
        test_pass "returns SKIP when identical but no manifest"
    else
        test_fail "classify_file no manifest identical" "$CLASSIFY_SKIP" "$result"
    fi
}

# ============================================================================
# Detailed Classification Tests
# ============================================================================

test_classify_detailed_returns_json() {
    run_test "classify_file_detailed: returns valid JSON"

    echo "content" > "$TEST_TMP/knossos/.claude/test.md"
    echo "content" > "$TEST_TMP/local/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"

    local result
    result=$(classify_file_detailed "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    if echo "$result" | jq -e . >/dev/null 2>&1; then
        test_pass "returns valid JSON"
    else
        test_fail "classify_file_detailed JSON" "valid JSON" "$result"
    fi
}

test_classify_detailed_includes_checksums() {
    run_test "classify_file_detailed: includes all checksums"

    echo "knossos" > "$TEST_TMP/knossos/.claude/test.md"
    echo "local" > "$TEST_TMP/local/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    add_file_to_manifest ".claude/test.md" "manifest-checksum"

    local result
    result=$(classify_file_detailed "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    local has_knossos has_local
    has_knossos=$(echo "$result" | jq -r '.knossos_checksum')
    has_local=$(echo "$result" | jq -r '.local_checksum')

    if [[ -n "$has_knossos" && "$has_knossos" != "null" && -n "$has_local" && "$has_local" != "null" ]]; then
        test_pass "includes knossos and local checksums"
    else
        test_fail "classify_file_detailed checksums" "non-null checksums" "knossos=$has_knossos local=$has_local"
    fi
}

test_classify_detailed_includes_reason() {
    run_test "classify_file_detailed: includes reason"

    echo "knossos changes" > "$TEST_TMP/knossos/.claude/test.md"
    echo "local changes" > "$TEST_TMP/local/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    add_file_to_manifest ".claude/test.md" "old-checksum"

    local result
    result=$(classify_file_detailed "test.md" "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md")

    local reason
    reason=$(echo "$result" | jq -r '.reason')

    if [[ "$reason" == "both_modified" ]]; then
        test_pass "reason is 'both_modified'"
    else
        test_fail "classify_file_detailed reason" "both_modified" "$reason"
    fi
}

# ============================================================================
# Backup Creation Tests (TDD 4.3)
# ============================================================================

test_backup_creates_timestamped_dir() {
    run_test "create_conflict_backup: creates timestamped directory"

    echo "local content" > "$TEST_TMP/local/.claude/test.md"

    init_conflict_backup_session "$TEST_TMP/backup"

    if [[ "$_CONFLICT_BACKUP_DIR" =~ ^$TEST_TMP/backup/[0-9]{8}-[0-9]{6}$ ]]; then
        test_pass "creates YYYYMMDD-HHMMSS directory"
    else
        test_fail "backup directory format" "YYYYMMDD-HHMMSS" "$_CONFLICT_BACKUP_DIR"
    fi
}

test_backup_preserves_path_structure() {
    run_test "create_conflict_backup: preserves relative path"

    mkdir -p "$TEST_TMP/local/.claude/subdir"
    echo "nested content" > "$TEST_TMP/local/.claude/subdir/nested.md"

    init_conflict_backup_session "$TEST_TMP/backup"
    local backup_file
    # Capture only stdout (backup path), send log messages to /dev/null
    backup_file=$(create_conflict_backup "$TEST_TMP/local/.claude/subdir/nested.md" 2>/dev/null | tail -1)

    # Check that path structure is preserved
    if [[ -f "$backup_file" && "$backup_file" == *"/.claude/subdir/nested.md" ]]; then
        test_pass "preserves directory structure in backup"
    else
        test_fail "backup path structure" "contains .claude/subdir/nested.md" "$backup_file"
    fi
}

test_backup_copies_content() {
    run_test "create_conflict_backup: copies file content correctly"

    local original_content="This is the local content that should be backed up"
    echo "$original_content" > "$TEST_TMP/local/.claude/test.md"

    init_conflict_backup_session "$TEST_TMP/backup"
    local backup_file
    # Use tail -1 to get just the path (last line of output)
    backup_file=$(create_conflict_backup "$TEST_TMP/local/.claude/test.md" 2>/dev/null | tail -1)

    local backup_content
    backup_content=$(cat "$backup_file")

    if [[ "$backup_content" == "$original_content" ]]; then
        test_pass "backup content matches original"
    else
        test_fail "backup content" "$original_content" "$backup_content"
    fi
}

test_backup_increments_count() {
    run_test "create_conflict_backup: increments conflict count"

    echo "file1" > "$TEST_TMP/local/file1.md"
    echo "file2" > "$TEST_TMP/local/file2.md"

    init_conflict_backup_session "$TEST_TMP/backup"

    local initial_count=$_CONFLICT_COUNT
    create_conflict_backup "$TEST_TMP/local/file1.md" >/dev/null
    create_conflict_backup "$TEST_TMP/local/file2.md" >/dev/null

    if [[ $_CONFLICT_COUNT -eq $((initial_count + 2)) ]]; then
        test_pass "conflict count incremented correctly"
    else
        test_fail "conflict count" "$((initial_count + 2))" "$_CONFLICT_COUNT"
    fi
}

test_backup_tracks_files() {
    run_test "create_conflict_backup: tracks conflicting files"

    echo "file1" > "$TEST_TMP/local/file1.md"
    echo "file2" > "$TEST_TMP/local/file2.md"

    init_conflict_backup_session "$TEST_TMP/backup"
    create_conflict_backup "$TEST_TMP/local/file1.md" >/dev/null
    create_conflict_backup "$TEST_TMP/local/file2.md" >/dev/null

    if [[ ${#_CONFLICT_FILES[@]} -eq 2 ]]; then
        test_pass "tracks all conflicting files"
    else
        test_fail "conflict file tracking" "2 files" "${#_CONFLICT_FILES[@]} files"
    fi
}

test_backup_handles_nonexistent() {
    run_test "create_conflict_backup: handles nonexistent file"

    init_conflict_backup_session "$TEST_TMP/backup"
    local result
    result=$(create_conflict_backup "$TEST_TMP/local/nonexistent.md" 2>/dev/null) || true

    if [[ -z "$result" ]]; then
        test_pass "returns empty for nonexistent file"
    else
        test_fail "backup nonexistent" "empty" "$result"
    fi
}

# ============================================================================
# Conflict Resolution Tests (TDD 4.3)
# ============================================================================

test_resolve_without_force_keeps_local() {
    run_test "resolve_conflict: without force keeps local file"

    local original_local="local content"
    echo "$original_local" > "$TEST_TMP/local/.claude/test.md"
    echo "knossos content" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    init_conflict_backup_session "$TEST_TMP/backup"

    resolve_conflict "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md" "test.md" 0

    local current_content
    current_content=$(cat "$TEST_TMP/local/.claude/test.md")

    if [[ "$current_content" == "$original_local" ]]; then
        test_pass "local file unchanged without force"
    else
        test_fail "resolve without force" "$original_local" "$current_content"
    fi
}

test_resolve_without_force_creates_backup() {
    run_test "resolve_conflict: without force creates backup"

    echo "local content" > "$TEST_TMP/local/.claude/test.md"
    echo "knossos content" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    # Reset conflict tracking for this test
    _CONFLICT_BACKUP_DIR=""
    _CONFLICT_COUNT=0
    _CONFLICT_FILES=()
    init_conflict_backup_session "$TEST_TMP/backup"

    resolve_conflict "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md" "test.md" 0 2>/dev/null

    if [[ $_CONFLICT_COUNT -eq 1 ]]; then
        test_pass "backup created without force"
    else
        test_fail "resolve backup without force" "1 backup" "$_CONFLICT_COUNT backups"
    fi
}

test_resolve_with_force_overwrites() {
    run_test "resolve_conflict: with force overwrites local"

    local knossos_content="knossos content"
    echo "local content" > "$TEST_TMP/local/.claude/test.md"
    echo "$knossos_content" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    init_conflict_backup_session "$TEST_TMP/backup"

    resolve_conflict "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md" "test.md" 1

    local current_content
    current_content=$(cat "$TEST_TMP/local/.claude/test.md")

    if [[ "$current_content" == "$knossos_content" ]]; then
        test_pass "local file overwritten with force"
    else
        test_fail "resolve with force" "$knossos_content" "$current_content"
    fi
}

test_resolve_with_force_still_creates_backup() {
    run_test "resolve_conflict: with force still creates backup"

    echo "local content" > "$TEST_TMP/local/.claude/test.md"
    echo "knossos content" > "$TEST_TMP/knossos/.claude/test.md"

    create_test_manifest "$TEST_TMP/knossos"
    # Reset conflict tracking for this test
    _CONFLICT_BACKUP_DIR=""
    _CONFLICT_COUNT=0
    _CONFLICT_FILES=()
    init_conflict_backup_session "$TEST_TMP/backup"

    resolve_conflict "$TEST_TMP/knossos/.claude/test.md" "$TEST_TMP/local/.claude/test.md" "test.md" 1 2>/dev/null

    if [[ $_CONFLICT_COUNT -eq 1 ]]; then
        test_pass "backup created even with force"
    else
        test_fail "resolve backup with force" "1 backup" "$_CONFLICT_COUNT backups"
    fi
}

# ============================================================================
# Conflict Report Tests
# ============================================================================

test_report_generates_file() {
    run_test "generate_conflict_report: creates report file"

    echo "file1" > "$TEST_TMP/local/file1.md"

    # Reset and initialize for this test
    _CONFLICT_BACKUP_DIR=""
    _CONFLICT_COUNT=0
    _CONFLICT_FILES=()
    init_conflict_backup_session "$TEST_TMP/backup"
    create_conflict_backup "$TEST_TMP/local/file1.md" >/dev/null 2>&1

    local report_file
    # get last line of output (the path, not log messages)
    report_file=$(generate_conflict_report 2>/dev/null | tail -1)

    if [[ -f "$report_file" ]]; then
        test_pass "report file created"
    else
        test_fail "report file creation" "file exists" "file not found: $report_file"
    fi
}

test_report_includes_file_list() {
    run_test "generate_conflict_report: includes all conflicting files"

    echo "file1" > "$TEST_TMP/local/file1.md"
    echo "file2" > "$TEST_TMP/local/file2.md"

    # Reset and initialize for this test
    _CONFLICT_BACKUP_DIR=""
    _CONFLICT_COUNT=0
    _CONFLICT_FILES=()
    init_conflict_backup_session "$TEST_TMP/backup"
    create_conflict_backup "$TEST_TMP/local/file1.md" >/dev/null 2>&1
    create_conflict_backup "$TEST_TMP/local/file2.md" >/dev/null 2>&1

    local report_file
    report_file=$(generate_conflict_report 2>/dev/null | tail -1)

    local contains_file1 contains_file2
    contains_file1=$(grep -c "file1.md" "$report_file" 2>/dev/null || echo "0")
    contains_file2=$(grep -c "file2.md" "$report_file" 2>/dev/null || echo "0")

    if [[ $contains_file1 -ge 1 && $contains_file2 -ge 1 ]]; then
        test_pass "report includes all files"
    else
        test_fail "report file list" "both files listed" "file1:$contains_file1 file2:$contains_file2"
    fi
}

test_report_includes_instructions() {
    run_test "generate_conflict_report: includes resolution instructions"

    echo "file1" > "$TEST_TMP/local/file1.md"

    # Reset and initialize for this test
    _CONFLICT_BACKUP_DIR=""
    _CONFLICT_COUNT=0
    _CONFLICT_FILES=()
    init_conflict_backup_session "$TEST_TMP/backup"
    create_conflict_backup "$TEST_TMP/local/file1.md" >/dev/null 2>&1

    local report_file
    report_file=$(generate_conflict_report 2>/dev/null | tail -1)

    local has_force_instruction
    has_force_instruction=$(grep -c "\-\-force" "$report_file" 2>/dev/null || echo "0")

    if [[ $has_force_instruction -ge 1 ]]; then
        test_pass "report includes --force instruction"
    else
        test_fail "report instructions" "--force mentioned" "not found"
    fi
}

# ============================================================================
# Finalize Conflicts Tests
# ============================================================================

test_finalize_returns_zero_no_conflicts() {
    run_test "finalize_conflicts: returns 0 when no conflicts"

    init_conflict_backup_session "$TEST_TMP/backup"
    # No conflicts created

    local exit_code=0
    finalize_conflicts 0 || exit_code=$?

    if [[ $exit_code -eq 0 ]]; then
        test_pass "returns 0 with no conflicts"
    else
        test_fail "finalize no conflicts" "0" "$exit_code"
    fi
}

test_finalize_returns_exit_code_without_force() {
    run_test "finalize_conflicts: returns EXIT_SYNC_CONFLICTS without force"

    echo "file1" > "$TEST_TMP/local/file1.md"

    init_conflict_backup_session "$TEST_TMP/backup"
    create_conflict_backup "$TEST_TMP/local/file1.md" >/dev/null

    local exit_code=0
    finalize_conflicts 0 >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq $EXIT_SYNC_CONFLICTS ]]; then
        test_pass "returns EXIT_SYNC_CONFLICTS (5)"
    else
        test_fail "finalize exit code" "$EXIT_SYNC_CONFLICTS" "$exit_code"
    fi
}

test_finalize_returns_zero_with_force() {
    run_test "finalize_conflicts: returns 0 with force"

    echo "file1" > "$TEST_TMP/local/file1.md"

    init_conflict_backup_session "$TEST_TMP/backup"
    create_conflict_backup "$TEST_TMP/local/file1.md" >/dev/null

    local exit_code=0
    finalize_conflicts 1 >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 0 ]]; then
        test_pass "returns 0 with force"
    else
        test_fail "finalize with force" "0" "$exit_code"
    fi
}

# ============================================================================
# Backup Management Tests
# ============================================================================

test_count_backups_empty() {
    run_test "count_conflict_backups: returns 0 for empty directory"

    local count
    count=$(count_conflict_backups "$TEST_TMP/nonexistent")

    if [[ "$count" == "0" ]]; then
        test_pass "returns 0 for nonexistent directory"
    else
        test_fail "count empty" "0" "$count"
    fi
}

test_count_backups_correct() {
    run_test "count_conflict_backups: counts backup directories"

    # Use a fresh directory for this test
    local test_backup_dir="$TEST_TMP/backup-count-test"
    mkdir -p "$test_backup_dir/20260101-120000"
    mkdir -p "$test_backup_dir/20260102-120000"

    local count
    count=$(count_conflict_backups "$test_backup_dir")

    if [[ "$count" == "2" ]]; then
        test_pass "counts backup directories correctly"
    else
        test_fail "count backups" "2" "$count"
    fi
}

test_list_backups_sorted() {
    run_test "list_conflict_backups: returns backups sorted by date (newest first)"

    # Use a fresh directory for this test
    local test_backup_dir="$TEST_TMP/backup-sort-test"
    mkdir -p "$test_backup_dir/20260101-120000"
    mkdir -p "$test_backup_dir/20260103-120000"
    mkdir -p "$test_backup_dir/20260102-120000"

    local first_result
    first_result=$(list_conflict_backups "$test_backup_dir" | head -1)

    if [[ "$first_result" == *"20260103"* ]]; then
        test_pass "newest backup listed first"
    else
        test_fail "backup sort order" "20260103 first" "$first_result"
    fi
}

test_clean_old_backups() {
    run_test "clean_old_backups: keeps only specified number"

    # Use a fresh directory for this test
    local test_backup_dir="$TEST_TMP/backup-clean-test"
    mkdir -p "$test_backup_dir/20260101-120000"
    mkdir -p "$test_backup_dir/20260102-120000"
    mkdir -p "$test_backup_dir/20260103-120000"
    mkdir -p "$test_backup_dir/20260104-120000"
    mkdir -p "$test_backup_dir/20260105-120000"

    clean_old_backups 3 "$test_backup_dir" 2>/dev/null

    local count
    count=$(count_conflict_backups "$test_backup_dir")

    if [[ "$count" == "3" ]]; then
        test_pass "keeps only 3 most recent backups"
    else
        test_fail "clean old backups" "3" "$count"
    fi
}

test_clean_old_backups_keeps_newest() {
    run_test "clean_old_backups: keeps newest backups"

    # Use a fresh directory for this test
    local test_backup_dir="$TEST_TMP/backup-keeps-newest-test"
    mkdir -p "$test_backup_dir/20260101-120000"
    mkdir -p "$test_backup_dir/20260102-120000"
    mkdir -p "$test_backup_dir/20260103-120000"

    clean_old_backups 2 "$test_backup_dir" 2>/dev/null

    # Check that oldest was removed
    if [[ ! -d "$test_backup_dir/20260101-120000" && -d "$test_backup_dir/20260103-120000" ]]; then
        test_pass "removes oldest, keeps newest"
    else
        test_fail "clean keeps newest" "20260103 exists, 20260101 removed" "incorrect"
    fi
}

# ============================================================================
# Run Tests
# ============================================================================

echo "=========================================="
echo "Running sync-conflict.sh tests (TDD 4.3)"
echo "=========================================="
echo ""

setup

# Classification tests
test_classify_skip_up_to_date
test_classify_skip_preserve_local
test_classify_update_knossos_changed
test_classify_conflict_both_changed
test_classify_new_local_missing
test_classify_conflict_no_manifest
test_classify_skip_no_manifest_identical

# Detailed classification tests
test_classify_detailed_returns_json
test_classify_detailed_includes_checksums
test_classify_detailed_includes_reason

# Backup creation tests
test_backup_creates_timestamped_dir
test_backup_preserves_path_structure
test_backup_copies_content
test_backup_increments_count
test_backup_tracks_files
test_backup_handles_nonexistent

# Conflict resolution tests
test_resolve_without_force_keeps_local
test_resolve_without_force_creates_backup
test_resolve_with_force_overwrites
test_resolve_with_force_still_creates_backup

# Conflict report tests
test_report_generates_file
test_report_includes_file_list
test_report_includes_instructions

# Finalize tests
test_finalize_returns_zero_no_conflicts
test_finalize_returns_exit_code_without_force
test_finalize_returns_zero_with_force

# Backup management tests
test_count_backups_empty
test_count_backups_correct
test_list_backups_sorted
test_clean_old_backups
test_clean_old_backups_keeps_newest

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

#!/usr/bin/env bash
#
# test-team-transaction.sh - Unit tests for team-transaction.sh
#
# Tests transaction infrastructure including atomic writes, journal
# management, staging operations, and backup creation.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../../.." && pwd)}"

# Source dependencies
source "$ROSTER_HOME/lib/team/team-transaction.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""

# Mock logging functions (team-transaction.sh expects these)
log() {
    echo "[LOG] $*" >&2
}

log_debug() {
    echo "[DEBUG] $*" >&2
}

log_warning() {
    echo "[WARNING] $*" >&2
}

log_error() {
    echo "[ERROR] $*" >&2
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

    # Override constants for testing
    JOURNAL_FILE="$TEST_TMP/.swap-journal"
    STAGING_DIR="$TEST_TMP/.swap-staging"
    SWAP_BACKUP_DIR="$TEST_TMP/.swap-backup"
    MANIFEST_FILE="$TEST_TMP/AGENT_MANIFEST.json"

    # Create mock roster structure
    mkdir -p "$TEST_TMP/mock-roster/teams/test-team/agents"
    echo "# Test Agent" > "$TEST_TMP/mock-roster/teams/test-team/agents/test-agent.md"
    echo "# Another Agent" > "$TEST_TMP/mock-roster/teams/test-team/agents/other-agent.md"
    echo "workflow: test" > "$TEST_TMP/mock-roster/teams/test-team/workflow.yaml"

    # Override ROSTER_HOME for tests
    ROSTER_HOME="$TEST_TMP/mock-roster"
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ============================================================================
# Tests for write_atomic()
# ============================================================================

test_write_atomic_new_file() {
    run_test "write_atomic creates new file atomically"

    local target="$TEST_TMP/new-file.txt"
    local content="test content"

    if write_atomic "$target" "$content"; then
        if [[ -f "$target" ]] && [[ "$(cat "$target")" == "$content" ]]; then
            test_pass "created file with correct content"
        else
            test_fail "write_atomic" "file with correct content" "file missing or wrong content"
        fi
    else
        test_fail "write_atomic" "return 0" "non-zero return"
    fi
}

test_write_atomic_overwrite() {
    run_test "write_atomic overwrites existing file"

    local target="$TEST_TMP/existing-file.txt"
    echo "old content" > "$target"

    local new_content="new content"
    if write_atomic "$target" "$new_content"; then
        if [[ "$(cat "$target")" == "$new_content" ]]; then
            test_pass "overwrote file with new content"
        else
            test_fail "write_atomic" "new content" "$(cat "$target")"
        fi
    else
        test_fail "write_atomic" "return 0" "non-zero return"
    fi
}

test_write_atomic_creates_parent() {
    run_test "write_atomic creates parent directory"

    local target="$TEST_TMP/subdir/nested/file.txt"
    local content="nested content"

    if write_atomic "$target" "$content"; then
        if [[ -f "$target" ]] && [[ "$(cat "$target")" == "$content" ]]; then
            test_pass "created parent directories and file"
        else
            test_fail "write_atomic" "parent created, file written" "missing parent or file"
        fi
    else
        test_fail "write_atomic" "return 0" "non-zero return"
    fi
}

test_write_atomic_cleanup_on_fail() {
    run_test "write_atomic cleans up temp file on failure"

    # Create unwritable directory
    local readonly_dir="$TEST_TMP/readonly"
    mkdir -p "$readonly_dir"

    # Count temp files before
    local before_count
    before_count=$(find "$TEST_TMP" -name "*.tmp.*" 2>/dev/null | wc -l | tr -d ' ')

    chmod 000 "$readonly_dir"
    local target="$readonly_dir/file.txt"
    write_atomic "$target" "content" 2>/dev/null || true
    chmod 755 "$readonly_dir"  # Restore immediately for find to work

    # Check no additional temp files left
    local after_count
    after_count=$(find "$TEST_TMP" -name "*.tmp.*" 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$after_count" -eq "$before_count" ]]; then
        test_pass "no temp files left after failure"
    else
        test_fail "write_atomic cleanup" "$before_count temp files" "$after_count temp files"
    fi
}

# ============================================================================
# Tests for Journal Operations
# ============================================================================

test_create_journal() {
    run_test "create_journal creates valid JSON journal"

    if create_journal "source-team" "target-team"; then
        if [[ -f "$JOURNAL_FILE" ]]; then
            local version
            version=$(jq -r '.version' "$JOURNAL_FILE")
            local source
            source=$(jq -r '.source_team' "$JOURNAL_FILE")
            local target
            target=$(jq -r '.target_team' "$JOURNAL_FILE")
            local phase
            phase=$(jq -r '.phase' "$JOURNAL_FILE")

            if [[ "$version" == "1.0" ]] && \
               [[ "$source" == "source-team" ]] && \
               [[ "$target" == "target-team" ]] && \
               [[ "$phase" == "PREPARING" ]]; then
                test_pass "created valid journal with correct fields"
            else
                test_fail "create_journal" "valid journal structure" "incorrect field values"
            fi
        else
            test_fail "create_journal" "journal file created" "file missing"
        fi
    else
        test_fail "create_journal" "return 0" "non-zero return"
    fi

    # Cleanup for next test
    rm -f "$JOURNAL_FILE"
}

test_create_journal_virgin_swap() {
    run_test "create_journal handles virgin swap (empty source)"

    if create_journal "" "target-team"; then
        local source
        source=$(jq -r '.source_team' "$JOURNAL_FILE")
        if [[ "$source" == "null" ]]; then
            test_pass "virgin swap: source_team is null"
        else
            test_fail "create_journal" "source_team: null" "source_team: $source"
        fi
    else
        test_fail "create_journal" "return 0" "non-zero return"
    fi

    # Cleanup for next test
    rm -f "$JOURNAL_FILE"
}

test_create_journal_concurrent() {
    run_test "create_journal fails if journal already exists"

    # Create existing journal
    create_journal "old-team" "new-team" 2>/dev/null

    # Try to create again
    if create_journal "another-team" "yet-another" 2>/dev/null; then
        test_fail "create_journal" "return 1 (concurrent)" "return 0"
    else
        test_pass "correctly prevented concurrent swap"
    fi

    # Cleanup for next test
    rm -f "$JOURNAL_FILE"
}

test_update_journal_phase() {
    run_test "update_journal_phase updates phase correctly"

    create_journal "source" "target" 2>/dev/null

    if update_journal_phase "STAGING"; then
        local phase
        phase=$(jq -r '.phase' "$JOURNAL_FILE")
        if [[ "$phase" == "STAGING" ]]; then
            test_pass "phase updated to STAGING"
        else
            test_fail "update_journal_phase" "STAGING" "$phase"
        fi
    else
        test_fail "update_journal_phase" "return 0" "non-zero return"
    fi

    rm -f "$JOURNAL_FILE"
}

test_update_journal_backups() {
    run_test "update_journal_backups updates backup_location"

    create_journal "source" "target" 2>/dev/null

    if update_journal_backups "commands" "/path/to/commands"; then
        local path
        path=$(jq -r '.backup_location.commands' "$JOURNAL_FILE")
        if [[ "$path" == "/path/to/commands" ]]; then
            test_pass "backup location updated for commands"
        else
            test_fail "update_journal_backups" "/path/to/commands" "$path"
        fi
    else
        test_fail "update_journal_backups" "return 0" "non-zero return"
    fi

    rm -f "$JOURNAL_FILE"
}

test_update_journal_error() {
    run_test "update_journal_error records error message"

    create_journal "source" "target" 2>/dev/null

    if update_journal_error "test error message"; then
        local error
        error=$(jq -r '.error' "$JOURNAL_FILE")
        if [[ "$error" == "test error message" ]]; then
            test_pass "error message recorded"
        else
            test_fail "update_journal_error" "test error message" "$error"
        fi
    else
        test_fail "update_journal_error" "return 0" "non-zero return"
    fi

    rm -f "$JOURNAL_FILE"
}

test_get_journal_field() {
    run_test "get_journal_field returns correct field value"

    create_journal "source-team" "target-team" 2>/dev/null

    local result
    result=$(get_journal_field "target_team")

    if [[ "$result" == "target-team" ]]; then
        test_pass "returned correct field value"
    else
        test_fail "get_journal_field" "target-team" "$result"
    fi

    rm -f "$JOURNAL_FILE"
}

test_get_journal_phase() {
    run_test "get_journal_phase returns phase"

    create_journal "source" "target" 2>/dev/null
    update_journal_phase "COMMITTING" 2>/dev/null

    local result
    result=$(get_journal_phase)

    if [[ "$result" == "COMMITTING" ]]; then
        test_pass "returned correct phase"
    else
        test_fail "get_journal_phase" "COMMITTING" "$result"
    fi

    rm -f "$JOURNAL_FILE"
}

test_delete_journal() {
    run_test "delete_journal removes journal file"

    create_journal "source" "target" 2>/dev/null
    delete_journal

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        test_pass "journal file removed"
    else
        test_fail "delete_journal" "file removed" "file still exists"
    fi
}

test_journal_exists_true() {
    run_test "journal_exists returns 0 when journal exists"

    create_journal "source" "target" 2>/dev/null

    if journal_exists; then
        test_pass "correctly detected existing journal"
    else
        test_fail "journal_exists" "return 0" "return 1"
    fi
}

test_journal_exists_false() {
    run_test "journal_exists returns 1 when journal missing"

    if journal_exists; then
        test_fail "journal_exists" "return 1" "return 0"
    else
        test_pass "correctly detected missing journal"
    fi
}

# ============================================================================
# Tests for Staging Operations
# ============================================================================

test_create_staging() {
    run_test "create_staging creates directory"

    if create_staging; then
        if [[ -d "$STAGING_DIR" ]]; then
            test_pass "staging directory created"
        else
            test_fail "create_staging" "directory exists" "directory missing"
        fi
    else
        test_fail "create_staging" "return 0" "non-zero return"
    fi
}

test_create_staging_cleans_existing() {
    run_test "create_staging removes existing staging"

    # Create old staging with old content
    mkdir -p "$STAGING_DIR/old-content"
    touch "$STAGING_DIR/old-file.txt"

    create_staging

    if [[ -d "$STAGING_DIR" ]] && \
       [[ ! -d "$STAGING_DIR/old-content" ]] && \
       [[ ! -f "$STAGING_DIR/old-file.txt" ]]; then
        test_pass "removed old staging and created fresh"
    else
        test_fail "create_staging" "clean staging directory" "old content still present"
    fi
}

test_cleanup_staging() {
    run_test "cleanup_staging removes staging directory"

    mkdir -p "$STAGING_DIR"
    cleanup_staging

    if [[ ! -d "$STAGING_DIR" ]]; then
        test_pass "staging directory removed"
    else
        test_fail "cleanup_staging" "directory removed" "directory still exists"
    fi
}

test_stage_agents() {
    run_test "stage_agents copies agents from team pack"

    create_staging

    if stage_agents "test-team"; then
        if [[ -f "$STAGING_DIR/agents/test-agent.md" ]] && \
           [[ -f "$STAGING_DIR/agents/other-agent.md" ]]; then
            test_pass "agents copied to staging"
        else
            test_fail "stage_agents" "agents copied" "agents missing"
        fi
    else
        test_fail "stage_agents" "return 0" "non-zero return"
    fi
}

test_stage_workflow() {
    run_test "stage_workflow copies workflow.yaml"

    create_staging

    if stage_workflow "test-team"; then
        if [[ -f "$STAGING_DIR/ACTIVE_WORKFLOW.yaml" ]]; then
            test_pass "workflow copied to staging"
        else
            test_fail "stage_workflow" "workflow copied" "workflow missing"
        fi
    else
        test_fail "stage_workflow" "return 0" "non-zero return"
    fi
}

test_stage_active_rite() {
    run_test "stage_active_rite creates ACTIVE_RITE file"

    create_staging

    if stage_active_rite "test-rite"; then
        local content
        content=$(cat "$STAGING_DIR/ACTIVE_RITE" 2>/dev/null || echo "")
        if [[ "$content" == "test-rite" ]]; then
            test_pass "ACTIVE_RITE file created with correct content"
        else
            test_fail "stage_active_rite" "test-rite" "$content"
        fi
    else
        test_fail "stage_active_rite" "return 0" "non-zero return"
    fi
}

test_verify_staging_success() {
    run_test "verify_staging passes with correct agent count"

    create_staging
    stage_agents "test-team" 2>/dev/null
    stage_active_rite "test-team" 2>/dev/null

    if verify_staging 2; then
        test_pass "staging verified successfully"
    else
        test_fail "verify_staging" "return 0" "return 1"
    fi
}

test_verify_staging_wrong_count() {
    run_test "verify_staging fails on agent count mismatch"

    create_staging
    stage_agents "test-team" 2>/dev/null
    stage_active_team "test-team" 2>/dev/null

    if verify_staging 5 2>/dev/null; then
        test_fail "verify_staging" "return 1 (wrong count)" "return 0"
    else
        test_pass "correctly failed on count mismatch"
    fi
}

test_verify_staging_missing_dir() {
    run_test "verify_staging fails on missing staging directory"

    if verify_staging 2 2>/dev/null; then
        test_fail "verify_staging" "return 1 (missing dir)" "return 0"
    else
        test_pass "correctly failed on missing directory"
    fi
}

test_verify_staging_missing_active_rite() {
    run_test "verify_staging fails on missing ACTIVE_RITE"

    create_staging
    stage_agents "test-team" 2>/dev/null
    # Don't stage ACTIVE_RITE

    if verify_staging 2 2>/dev/null; then
        test_fail "verify_staging" "return 1 (missing ACTIVE_RITE)" "return 0"
    else
        test_pass "correctly failed on missing ACTIVE_RITE"
    fi
}

# ============================================================================
# Tests for Backup Operations
# ============================================================================

test_create_swap_backup() {
    run_test "create_swap_backup creates complete backup"

    # Create project structure to backup
    cd "$TEST_TMP"
    mkdir -p .claude/agents
    echo "# Agent" > .claude/agents/test.md
    echo "test-team" > .claude/ACTIVE_RITE
    echo '{}' > "$MANIFEST_FILE"

    # Create journal for backup to update
    create_journal "source" "target" 2>/dev/null

    if create_swap_backup 2>/dev/null; then
        if [[ -d "$SWAP_BACKUP_DIR/agents" ]] && \
           [[ -f "$SWAP_BACKUP_DIR/ACTIVE_RITE" ]] && \
           [[ -f "$SWAP_BACKUP_DIR/AGENT_MANIFEST.json" ]]; then
            test_pass "created complete backup"
        else
            test_fail "create_swap_backup" "complete backup" "missing files"
        fi
    else
        test_fail "create_swap_backup" "return 0" "non-zero return"
    fi
}

test_create_swap_backup_updates_journal() {
    run_test "create_swap_backup updates journal backup_location for resources"

    cd "$TEST_TMP"
    mkdir -p .claude/commands
    echo "cmd.md" > .claude/commands/.team-commands
    echo "# Command" > .claude/commands/cmd.md

    create_journal "source" "target" 2>/dev/null
    create_swap_backup 2>/dev/null

    local backup_path
    backup_path=$(jq -r '.backup_location.commands' "$JOURNAL_FILE")

    if [[ "$backup_path" == "$SWAP_BACKUP_DIR/commands" ]]; then
        test_pass "journal updated with commands backup location"
    else
        test_fail "create_swap_backup" "$SWAP_BACKUP_DIR/commands" "$backup_path"
    fi
}

test_cleanup_swap_backup() {
    run_test "cleanup_swap_backup removes backup directory"

    mkdir -p "$SWAP_BACKUP_DIR"
    cleanup_swap_backup

    if [[ ! -d "$SWAP_BACKUP_DIR" ]]; then
        test_pass "backup directory removed"
    else
        test_fail "cleanup_swap_backup" "directory removed" "directory still exists"
    fi
}

test_verify_backup_integrity_valid() {
    run_test "verify_backup_integrity returns 0 for valid backup"

    cd "$TEST_TMP"
    mkdir -p "$SWAP_BACKUP_DIR/agents"
    echo "# Agent" > "$SWAP_BACKUP_DIR/agents/test.md"
    echo "test-team" > "$SWAP_BACKUP_DIR/ACTIVE_RITE"

    create_journal "source" "target" 2>/dev/null

    if verify_backup_integrity; then
        test_pass "backup verified as valid"
    else
        test_fail "verify_backup_integrity" "return 0" "return 1"
    fi
}

test_verify_backup_missing() {
    run_test "verify_backup_integrity returns 1 for missing backup"

    if verify_backup_integrity 2>/dev/null; then
        test_fail "verify_backup_integrity" "return 1 (missing)" "return 0"
    else
        test_pass "correctly detected missing backup"
    fi
}

test_verify_backup_virgin_swap() {
    run_test "verify_backup_integrity handles virgin swap case"

    mkdir -p "$SWAP_BACKUP_DIR"
    # Virgin swap: no ACTIVE_RITE in backup

    create_journal "" "target-team" 2>/dev/null

    if verify_backup_integrity; then
        test_pass "correctly handled virgin swap (no ACTIVE_RITE required)"
    else
        test_fail "verify_backup_integrity" "return 0 (virgin)" "return 1"
    fi
}

# ============================================================================
# Integration Tests
# ============================================================================

test_transaction_full_cycle() {
    run_test "Full transaction cycle: journal -> staging -> backup -> cleanup"

    cd "$TEST_TMP"
    mkdir -p .claude/agents
    echo "# Agent" > .claude/agents/old.md

    # Create journal
    if ! create_journal "old-team" "test-team" 2>/dev/null; then
        test_fail "transaction cycle" "journal created" "journal creation failed"
        return
    fi

    # Update phase to STAGING
    update_journal_phase "STAGING" 2>/dev/null

    # Create and populate staging
    if ! create_staging 2>/dev/null; then
        test_fail "transaction cycle" "staging created" "staging creation failed"
        return
    fi

    stage_agents "test-team" 2>/dev/null
    stage_active_team "test-team" 2>/dev/null

    # Verify staging
    if ! verify_staging 2 2>/dev/null; then
        test_fail "transaction cycle" "staging verified" "staging verification failed"
        return
    fi

    # Create backup
    if ! create_swap_backup 2>/dev/null; then
        test_fail "transaction cycle" "backup created" "backup creation failed"
        return
    fi

    # Cleanup
    cleanup_staging
    cleanup_swap_backup
    delete_journal

    if [[ ! -d "$STAGING_DIR" ]] && \
       [[ ! -d "$SWAP_BACKUP_DIR" ]] && \
       [[ ! -f "$JOURNAL_FILE" ]]; then
        test_pass "full transaction cycle completed successfully"
    else
        test_fail "transaction cycle" "all cleaned up" "cleanup incomplete"
    fi
}

# ============================================================================
# Main test runner
# ============================================================================

main() {
    echo "========================================"
    echo "Team Transaction Unit Tests"
    echo "========================================"
    echo ""

    setup

    # Atomic I/O tests
    test_write_atomic_new_file
    test_write_atomic_overwrite
    test_write_atomic_creates_parent
    test_write_atomic_cleanup_on_fail

    # Journal tests
    test_create_journal
    test_create_journal_virgin_swap
    test_create_journal_concurrent
    test_update_journal_phase
    test_update_journal_backups
    test_update_journal_error
    test_get_journal_field
    test_get_journal_phase
    test_delete_journal
    test_journal_exists_true
    test_journal_exists_false

    # Staging tests
    test_create_staging
    test_create_staging_cleans_existing
    test_cleanup_staging
    test_stage_agents
    test_stage_workflow
    test_stage_active_team
    test_verify_staging_success
    test_verify_staging_wrong_count
    test_verify_staging_missing_dir
    test_verify_staging_missing_active_team

    # Backup tests
    test_create_swap_backup
    test_create_swap_backup_updates_journal
    test_cleanup_swap_backup
    test_verify_backup_integrity_valid
    test_verify_backup_missing
    test_verify_backup_virgin_swap

    # Integration tests
    test_transaction_full_cycle

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

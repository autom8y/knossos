#!/usr/bin/env bash
#
# test-init.sh - Unit tests for roster-sync init command
#
# Tests initialization scenarios per TDD Section 3.2:
#   - Fresh project initialization
#   - --force overwrite
#   - Invalid path handling
#   - --team flag
#
# Part of: Skeleton Deprecation & CEM Migration (task-012)

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../.." && pwd)}"
export ROSTER_HOME

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
    echo "Test temp dir: $TEST_TMP"
    echo "Test project: $TEST_PROJECT"
}

teardown() {
    if [[ -n "$TEST_TMP" && -d "$TEST_TMP" ]]; then
        rm -rf "$TEST_TMP"
    fi
}

# Create a clean test project
reset_test_project() {
    rm -rf "$TEST_PROJECT"
    mkdir -p "$TEST_PROJECT"
}

# ============================================================================
# Tests: Fresh Project Initialization
# ============================================================================

test_init_fresh_project() {
    run_test "Initialize fresh project"
    reset_test_project

    # Run init
    local output
    output=$("$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" 2>&1) || {
        test_fail "init command" "exit 0" "exit $?"
        echo "$output"
        return
    }

    # Check .claude directory created
    if [[ -d "$TEST_PROJECT/.claude" ]]; then
        test_pass ".claude directory created"
    else
        test_fail ".claude directory" "exists" "missing"
    fi

    # Check .claude/.cem directory created
    if [[ -d "$TEST_PROJECT/.claude/.cem" ]]; then
        test_pass ".claude/.cem directory created"
    else
        test_fail ".claude/.cem directory" "exists" "missing"
    fi

    # Check manifest created
    if [[ -f "$TEST_PROJECT/.claude/.cem/manifest.json" ]]; then
        test_pass "manifest.json created"
    else
        test_fail "manifest.json" "exists" "missing"
        return
    fi

    # Check manifest is valid JSON with schema version 3
    local schema_version
    schema_version=$(jq -r '.schema_version' "$TEST_PROJECT/.claude/.cem/manifest.json" 2>/dev/null)
    if [[ "$schema_version" == "3" ]]; then
        test_pass "manifest schema version is 3"
    else
        test_fail "schema_version" "3" "$schema_version"
    fi

    # Check roster.path in manifest
    local roster_path
    roster_path=$(jq -r '.roster.path' "$TEST_PROJECT/.claude/.cem/manifest.json" 2>/dev/null)
    if [[ "$roster_path" == "$ROSTER_HOME" ]]; then
        test_pass "roster.path set correctly"
    else
        test_fail "roster.path" "$ROSTER_HOME" "$roster_path"
    fi

    # Check managed_files array exists
    local file_count
    file_count=$(jq '.managed_files | length' "$TEST_PROJECT/.claude/.cem/manifest.json" 2>/dev/null)
    if [[ "$file_count" -gt 0 ]]; then
        test_pass "managed_files array populated ($file_count files)"
    else
        test_fail "managed_files" ">0 files" "$file_count files"
    fi

    # Check settings.json created
    if [[ -f "$TEST_PROJECT/.claude/settings.json" ]]; then
        test_pass "settings.json created"
    else
        test_fail "settings.json" "exists" "missing"
    fi

    # Check hooks directory created
    if [[ -d "$TEST_PROJECT/.claude/hooks" ]]; then
        test_pass "hooks directory created"
    else
        test_fail "hooks directory" "exists" "missing"
    fi

    # Check agents directory created
    if [[ -d "$TEST_PROJECT/.claude/agents" ]]; then
        test_pass "agents directory created"
    else
        test_fail "agents directory" "exists" "missing"
    fi
}

test_init_creates_copy_replace_items() {
    run_test "Init creates copy-replace items"
    reset_test_project

    # Run init
    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || true

    # Check COMMAND_REGISTRY.md (a copy-replace item)
    if [[ -f "$TEST_PROJECT/.claude/COMMAND_REGISTRY.md" ]]; then
        test_pass "COMMAND_REGISTRY.md created"

        # Verify it matches roster version
        if diff -q "$ROSTER_HOME/.claude/COMMAND_REGISTRY.md" "$TEST_PROJECT/.claude/COMMAND_REGISTRY.md" >/dev/null 2>&1; then
            test_pass "COMMAND_REGISTRY.md matches roster"
        else
            test_fail "COMMAND_REGISTRY.md content" "matches roster" "differs"
        fi
    else
        test_fail "COMMAND_REGISTRY.md" "exists" "missing"
    fi
}

test_init_creates_merge_items() {
    run_test "Init creates merge items"
    reset_test_project

    # Run init
    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || true

    # Check CLAUDE.md (a merge item)
    if [[ -f "$TEST_PROJECT/.claude/CLAUDE.md" ]]; then
        test_pass "CLAUDE.md created"
    else
        test_fail "CLAUDE.md" "exists" "missing"
    fi

    # Check settings.local.json (a merge item)
    if [[ -f "$TEST_PROJECT/.claude/settings.local.json" ]]; then
        test_pass "settings.local.json created"
    else
        test_fail "settings.local.json" "exists" "missing"
    fi
}

# ============================================================================
# Tests: Already Initialized (--force)
# ============================================================================

test_init_already_initialized_error() {
    run_test "Init fails if already initialized (without --force)"
    reset_test_project

    # Initialize first
    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || true

    # Try to initialize again
    local exit_code=0
    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 3 ]]; then
        test_pass "exits with code 3 (EXIT_SYNC_INIT_FAILED)"
    else
        test_fail "exit code" "3" "$exit_code"
    fi
}

test_init_force_reinitialize() {
    run_test "Init with --force reinitializes"
    reset_test_project

    # Initialize first
    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || true

    # Modify a file to detect overwrite
    echo "modified content" > "$TEST_PROJECT/.claude/COMMAND_REGISTRY.md"

    # Reinitialize with --force
    local exit_code=0
    "$ROSTER_HOME/roster-sync" --force init "$TEST_PROJECT" >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 0 ]]; then
        test_pass "exits with code 0"
    else
        test_fail "exit code" "0" "$exit_code"
    fi

    # Check file was overwritten
    if diff -q "$ROSTER_HOME/.claude/COMMAND_REGISTRY.md" "$TEST_PROJECT/.claude/COMMAND_REGISTRY.md" >/dev/null 2>&1; then
        test_pass "COMMAND_REGISTRY.md was overwritten"
    else
        test_fail "COMMAND_REGISTRY.md" "matches roster after --force" "still modified"
    fi
}

test_init_force_preserves_existing_merge_items() {
    run_test "Init with --force preserves existing merge items"
    reset_test_project

    # Initialize first
    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || true

    # Modify a merge item
    echo "# Custom CLAUDE.md content" > "$TEST_PROJECT/.claude/CLAUDE.md"
    local original_content
    original_content=$(cat "$TEST_PROJECT/.claude/CLAUDE.md")

    # Reinitialize with --force
    "$ROSTER_HOME/roster-sync" --force init "$TEST_PROJECT" >/dev/null 2>&1 || true

    # Merge items should be preserved (only merged on sync, not init)
    local current_content
    current_content=$(cat "$TEST_PROJECT/.claude/CLAUDE.md")
    if [[ "$current_content" == "$original_content" ]]; then
        test_pass "CLAUDE.md preserved during --force init"
    else
        test_fail "CLAUDE.md" "preserved" "modified"
    fi
}

# ============================================================================
# Tests: Invalid Path Handling
# ============================================================================

test_init_nonexistent_path() {
    run_test "Init fails for non-existent path"

    local fake_path="$TEST_TMP/does-not-exist"
    local exit_code=0

    "$ROSTER_HOME/roster-sync" init "$fake_path" >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 3 ]]; then
        test_pass "exits with code 3 for non-existent path"
    else
        test_fail "exit code" "3" "$exit_code"
    fi
}

test_init_inside_roster_fails() {
    run_test "Init fails inside roster directory"

    local exit_code=0
    "$ROSTER_HOME/roster-sync" init "$ROSTER_HOME" >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 3 ]]; then
        test_pass "exits with code 3 when inside roster"
    else
        test_fail "exit code" "3" "$exit_code"
    fi
}

test_init_current_directory() {
    run_test "Init in current directory (no path argument)"
    reset_test_project

    # Change to test project and run init without path
    (
        cd "$TEST_PROJECT"
        "$ROSTER_HOME/roster-sync" init >/dev/null 2>&1
    ) || {
        test_fail "init in current directory" "exit 0" "exit $?"
        return
    }

    if [[ -f "$TEST_PROJECT/.claude/.cem/manifest.json" ]]; then
        test_pass "initialized current directory"
    else
        test_fail "manifest" "exists" "missing"
    fi
}

# ============================================================================
# Tests: --team Flag
# ============================================================================

test_init_with_team_flag() {
    run_test "Init with --team flag"
    reset_test_project

    # Get first available team
    local team_name
    team_name=$(ls "$ROSTER_HOME/teams/" 2>/dev/null | head -1)

    if [[ -z "$team_name" ]]; then
        echo "  SKIP: No teams available"
        return
    fi

    # Run init with team
    local exit_code=0
    "$ROSTER_HOME/roster-sync" init --team="$team_name" "$TEST_PROJECT" >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 0 ]]; then
        test_pass "exits with code 0"
    else
        test_fail "exit code" "0" "$exit_code"
    fi

    # Check ACTIVE_TEAM file created
    if [[ -f "$TEST_PROJECT/.claude/ACTIVE_RITE" ]]; then
        local active_team
        active_team=$(cat "$TEST_PROJECT/.claude/ACTIVE_RITE")
        if [[ "$active_team" == "$team_name" ]]; then
            test_pass "ACTIVE_TEAM set to $team_name"
        else
            test_fail "ACTIVE_TEAM content" "$team_name" "$active_team"
        fi
    else
        test_fail "ACTIVE_TEAM file" "exists" "missing"
    fi

    # Check manifest has team info
    local manifest_team
    manifest_team=$(jq -r '.team.name // empty' "$TEST_PROJECT/.claude/.cem/manifest.json" 2>/dev/null)
    if [[ "$manifest_team" == "$team_name" ]]; then
        test_pass "manifest team.name set"
    else
        test_fail "manifest team.name" "$team_name" "$manifest_team"
    fi
}

test_init_with_invalid_team() {
    run_test "Init with invalid team fails"
    reset_test_project

    local exit_code=0
    "$ROSTER_HOME/roster-sync" init --team="nonexistent-team" "$TEST_PROJECT" >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 3 ]]; then
        test_pass "exits with code 3 for invalid team"
    else
        test_fail "exit code" "3" "$exit_code"
    fi

    # Check no manifest created
    if [[ ! -f "$TEST_PROJECT/.claude/.cem/manifest.json" ]]; then
        test_pass "no manifest created on failure"
    else
        test_fail "manifest" "not created" "created"
    fi
}

test_init_team_equals_syntax() {
    run_test "Init with --team=value syntax"
    reset_test_project

    local team_name
    team_name=$(ls "$ROSTER_HOME/teams/" 2>/dev/null | head -1)

    if [[ -z "$team_name" ]]; then
        echo "  SKIP: No teams available"
        return
    fi

    "$ROSTER_HOME/roster-sync" init "--team=$team_name" "$TEST_PROJECT" >/dev/null 2>&1 || {
        test_fail "init with --team=value" "exit 0" "exit $?"
        return
    }

    local active_team
    active_team=$(cat "$TEST_PROJECT/.claude/ACTIVE_RITE" 2>/dev/null)
    if [[ "$active_team" == "$team_name" ]]; then
        test_pass "--team=value syntax works"
    else
        test_fail "team" "$team_name" "$active_team"
    fi
}

test_init_team_space_syntax() {
    run_test "Init with --team value syntax"
    reset_test_project

    local team_name
    team_name=$(ls "$ROSTER_HOME/teams/" 2>/dev/null | head -1)

    if [[ -z "$team_name" ]]; then
        echo "  SKIP: No teams available"
        return
    fi

    "$ROSTER_HOME/roster-sync" init --team "$team_name" "$TEST_PROJECT" >/dev/null 2>&1 || {
        test_fail "init with --team value" "exit 0" "exit $?"
        return
    }

    local active_team
    active_team=$(cat "$TEST_PROJECT/.claude/ACTIVE_RITE" 2>/dev/null)
    if [[ "$active_team" == "$team_name" ]]; then
        test_pass "--team value syntax works"
    else
        test_fail "team" "$team_name" "$active_team"
    fi
}

# ============================================================================
# Tests: Dry Run
# ============================================================================

test_init_dry_run() {
    run_test "Init with --dry-run"
    reset_test_project

    local exit_code=0
    "$ROSTER_HOME/roster-sync" --dry-run init "$TEST_PROJECT" >/dev/null 2>&1 || exit_code=$?

    if [[ $exit_code -eq 0 ]]; then
        test_pass "exits with code 0"
    else
        test_fail "exit code" "0" "$exit_code"
    fi

    # Check nothing was created
    if [[ ! -d "$TEST_PROJECT/.claude" ]]; then
        test_pass "no .claude directory created (dry-run)"
    else
        test_fail ".claude directory" "not created" "created"
    fi
}

# ============================================================================
# Tests: Manifest Content
# ============================================================================

test_init_manifest_structure() {
    run_test "Init creates valid manifest structure"
    reset_test_project

    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || true

    local manifest="$TEST_PROJECT/.claude/.cem/manifest.json"

    # Check all required v3 fields
    local schema_version roster_path roster_commit roster_ref roster_last_sync
    schema_version=$(jq -r '.schema_version' "$manifest" 2>/dev/null)
    roster_path=$(jq -r '.roster.path' "$manifest" 2>/dev/null)
    roster_commit=$(jq -r '.roster.commit' "$manifest" 2>/dev/null)
    roster_ref=$(jq -r '.roster.ref' "$manifest" 2>/dev/null)
    roster_last_sync=$(jq -r '.roster.last_sync' "$manifest" 2>/dev/null)

    if [[ "$schema_version" == "3" ]]; then
        test_pass "schema_version is 3"
    else
        test_fail "schema_version" "3" "$schema_version"
    fi

    if [[ -n "$roster_path" && "$roster_path" != "null" ]]; then
        test_pass "roster.path set"
    else
        test_fail "roster.path" "set" "missing"
    fi

    if [[ -n "$roster_commit" && "$roster_commit" != "null" && "$roster_commit" != "" ]]; then
        test_pass "roster.commit set"
    else
        test_fail "roster.commit" "set" "missing"
    fi

    if [[ -n "$roster_ref" && "$roster_ref" != "null" ]]; then
        test_pass "roster.ref set"
    else
        test_fail "roster.ref" "set" "missing"
    fi

    if [[ -n "$roster_last_sync" && "$roster_last_sync" != "null" ]]; then
        test_pass "roster.last_sync set"
    else
        test_fail "roster.last_sync" "set" "missing"
    fi

    # Check managed_files array
    local has_managed_files
    has_managed_files=$(jq 'has("managed_files")' "$manifest" 2>/dev/null)
    if [[ "$has_managed_files" == "true" ]]; then
        test_pass "managed_files array exists"
    else
        test_fail "managed_files" "exists" "missing"
    fi

    # Check orphans array (should be empty)
    local orphans_count
    orphans_count=$(jq '.orphans | length' "$manifest" 2>/dev/null)
    if [[ "$orphans_count" == "0" ]]; then
        test_pass "orphans array is empty"
    else
        test_fail "orphans" "empty" "$orphans_count items"
    fi
}

test_init_managed_files_have_checksums() {
    run_test "Init managed files have checksums"
    reset_test_project

    "$ROSTER_HOME/roster-sync" init "$TEST_PROJECT" >/dev/null 2>&1 || true

    local manifest="$TEST_PROJECT/.claude/.cem/manifest.json"

    # Check first managed file has checksum
    local first_checksum
    first_checksum=$(jq -r '.managed_files[0].checksum // empty' "$manifest" 2>/dev/null)
    if [[ -n "$first_checksum" && ${#first_checksum} -eq 64 ]]; then
        test_pass "managed files have SHA-256 checksums"
    else
        test_fail "checksum" "64-char SHA-256" "$first_checksum"
    fi

    # Check first managed file has strategy
    local first_strategy
    first_strategy=$(jq -r '.managed_files[0].strategy // empty' "$manifest" 2>/dev/null)
    if [[ -n "$first_strategy" ]]; then
        test_pass "managed files have strategy"
    else
        test_fail "strategy" "set" "missing"
    fi
}

# ============================================================================
# Run Tests
# ============================================================================

echo "=========================================="
echo "Running roster-sync init tests"
echo "=========================================="

setup

# Fresh project tests
test_init_fresh_project
test_init_creates_copy_replace_items
test_init_creates_merge_items

# Already initialized tests
test_init_already_initialized_error
test_init_force_reinitialize
test_init_force_preserves_existing_merge_items

# Invalid path tests
test_init_nonexistent_path
test_init_inside_roster_fails
test_init_current_directory

# Team flag tests
test_init_with_team_flag
test_init_with_invalid_team
test_init_team_equals_syntax
test_init_team_space_syntax

# Dry run tests
test_init_dry_run

# Manifest structure tests
test_init_manifest_structure
test_init_managed_files_have_checksums

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

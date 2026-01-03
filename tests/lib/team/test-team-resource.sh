#!/usr/bin/env bash
#
# test-team-resource.sh - Unit tests for team-resource.sh
#
# Tests generic team resource operations including membership checks,
# backup, removal, and orphan detection.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../../.." && pwd)}"

# Source dependencies
source "$ROSTER_HOME/lib/team/team-resource.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""

# Mock logging functions (team-resource.sh expects these)
log() {
    echo "[LOG] $*" >&2
}

log_debug() {
    echo "[DEBUG] $*" >&2
}

log_warning() {
    echo "[WARNING] $*" >&2
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

    # Create mock team structure for testing
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/commands"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/skills"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/hooks"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/commands"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/skills"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/hooks"

    # Create mock commands (files)
    touch "$TEST_TMP/mock-roster/teams/team-a/commands/cmd-a.md"
    touch "$TEST_TMP/mock-roster/teams/team-b/commands/cmd-b.md"

    # Create mock skills (directories)
    mkdir -p "$TEST_TMP/mock-roster/teams/team-a/skills/skill-a"
    mkdir -p "$TEST_TMP/mock-roster/teams/team-b/skills/skill-b"

    # Create mock hooks (files)
    touch "$TEST_TMP/mock-roster/teams/team-a/hooks/hook-a.sh"
    touch "$TEST_TMP/mock-roster/teams/team-b/hooks/hook-b.sh"

    # Override ROSTER_HOME for tests
    ROSTER_HOME="$TEST_TMP/mock-roster"
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ============================================================================
# Tests for is_resource_from_team()
# ============================================================================

test_is_resource_from_team_command() {
    run_test "is_resource_from_team finds command file"

    if is_resource_from_team "cmd-a.md" "commands" "f"; then
        test_pass "found team-a command"
    else
        test_fail "is_resource_from_team" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_team_skill() {
    run_test "is_resource_from_team finds skill directory"

    if is_resource_from_team "skill-a" "skills" "d"; then
        test_pass "found team-a skill"
    else
        test_fail "is_resource_from_team" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_team_hook() {
    run_test "is_resource_from_team finds hook file"

    if is_resource_from_team "hook-b.sh" "hooks" "f"; then
        test_pass "found team-b hook"
    else
        test_fail "is_resource_from_team" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_team_not_found() {
    run_test "is_resource_from_team returns false for non-team resource"

    if is_resource_from_team "nonexistent.md" "commands" "f"; then
        test_fail "is_resource_from_team" "failure (return 1)" "success (return 0)"
    else
        test_pass "correctly returned false for nonexistent resource"
    fi
}

test_is_resource_from_team_wrong_type() {
    run_test "is_resource_from_team returns false when find type doesn't match"

    # skill-a is a directory, but we're looking for a file
    if is_resource_from_team "skill-a" "skills" "f"; then
        test_fail "is_resource_from_team" "failure (return 1)" "success (return 0)"
    else
        test_pass "correctly returned false when type doesn't match"
    fi
}

# ============================================================================
# Tests for get_resource_team()
# ============================================================================

test_get_resource_team_command() {
    run_test "get_resource_team returns correct team for command"

    local result
    result=$(get_resource_team "cmd-a.md" "commands" "f")

    if [[ "$result" == "team-a" ]]; then
        test_pass "returned correct team name: team-a"
    else
        test_fail "get_resource_team" "team-a" "$result"
    fi
}

test_get_resource_team_skill() {
    run_test "get_resource_team returns correct team for skill"

    local result
    result=$(get_resource_team "skill-b" "skills" "d")

    if [[ "$result" == "team-b" ]]; then
        test_pass "returned correct team name: team-b"
    else
        test_fail "get_resource_team" "team-b" "$result"
    fi
}

test_get_resource_team_hook() {
    run_test "get_resource_team returns correct team for hook"

    local result
    result=$(get_resource_team "hook-a.sh" "hooks" "f")

    if [[ "$result" == "team-a" ]]; then
        test_pass "returned correct team name: team-a"
    else
        test_fail "get_resource_team" "team-a" "$result"
    fi
}

test_get_resource_team_not_found() {
    run_test "get_resource_team returns empty for non-team resource"

    local result
    result=$(get_resource_team "nonexistent.md" "commands" "f")

    if [[ -z "$result" ]]; then
        test_pass "returned empty string for nonexistent resource"
    else
        test_fail "get_resource_team" "(empty)" "$result"
    fi
}

test_get_resource_team_multiple_teams() {
    run_test "get_resource_team returns first match when resource exists in multiple teams"

    # Create same command in both teams
    touch "$TEST_TMP/mock-roster/teams/team-a/commands/shared.md"
    touch "$TEST_TMP/mock-roster/teams/team-b/commands/shared.md"

    local result
    result=$(get_resource_team "shared.md" "commands" "f")

    # Should return one of them (behavior: first match from find)
    if [[ "$result" == "team-a" ]] || [[ "$result" == "team-b" ]]; then
        test_pass "returned a team name: $result"
    else
        test_fail "get_resource_team" "team-a or team-b" "$result"
    fi
}

# ============================================================================
# Tests for backup_team_resource() - RF-003
# ============================================================================

test_backup_team_resource_commands() {
    run_test "backup_team_resource backs up commands (files)"

    # Setup: create commands directory with marker
    local cmd_dir="$TEST_TMP/project/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "cmd-a.md" > "$cmd_dir/.team-commands"
    echo "cmd-b.md" >> "$cmd_dir/.team-commands"
    echo "test command content" > "$cmd_dir/cmd-a.md"
    echo "another command" > "$cmd_dir/cmd-b.md"

    # Act: backup commands
    cd "$TEST_TMP/project"
    backup_team_resource "commands" ".claude/commands" ".team-commands" "f"

    # Assert: backup directory exists with files
    if [[ -d "$cmd_dir.backup" ]] && \
       [[ -f "$cmd_dir.backup/cmd-a.md" ]] && \
       [[ -f "$cmd_dir.backup/cmd-b.md" ]]; then
        test_pass "backed up command files to .backup directory"
    else
        test_fail "backup_team_resource" "backup directory with files" "missing files or directory"
    fi
}

test_backup_team_resource_skills() {
    run_test "backup_team_resource backs up skills (directories)"

    # Setup: create skills directory with marker
    local skill_dir="$TEST_TMP/project/.claude/skills"
    mkdir -p "$skill_dir/skill-a/subdir"
    mkdir -p "$skill_dir/skill-b"
    echo "skill-a" > "$skill_dir/.team-skills"
    echo "skill-b" >> "$skill_dir/.team-skills"
    echo "content" > "$skill_dir/skill-a/skill.md"
    echo "nested" > "$skill_dir/skill-a/subdir/file.txt"

    # Act: backup skills
    cd "$TEST_TMP/project"
    backup_team_resource "skills" ".claude/skills" ".team-skills" "d"

    # Assert: backup directory exists with recursive copy
    if [[ -d "$skill_dir.backup/skill-a" ]] && \
       [[ -d "$skill_dir.backup/skill-b" ]] && \
       [[ -f "$skill_dir.backup/skill-a/skill.md" ]] && \
       [[ -f "$skill_dir.backup/skill-a/subdir/file.txt" ]]; then
        test_pass "backed up skill directories recursively"
    else
        test_fail "backup_team_resource" "backup directory with recursive structure" "missing structure"
    fi
}

test_backup_team_resource_hooks() {
    run_test "backup_team_resource backs up hooks (files)"

    # Setup: create hooks directory with marker
    local hook_dir="$TEST_TMP/project/.claude/hooks"
    mkdir -p "$hook_dir"
    echo "hook-a.sh" > "$hook_dir/.team-hooks"
    echo "#!/bin/bash" > "$hook_dir/hook-a.sh"

    # Act: backup hooks
    cd "$TEST_TMP/project"
    backup_team_resource "hooks" ".claude/hooks" ".team-hooks" "f"

    # Assert: backup directory exists with file
    if [[ -d "$hook_dir.backup" ]] && \
       [[ -f "$hook_dir.backup/hook-a.sh" ]]; then
        test_pass "backed up hook files"
    else
        test_fail "backup_team_resource" "backup directory with hook file" "missing file or directory"
    fi
}

test_backup_team_resource_no_marker() {
    run_test "backup_team_resource returns 0 when no marker file exists"

    # Setup: create directory without marker (use unique subdir)
    local project_dir="$TEST_TMP/project-no-marker"
    mkdir -p "$project_dir/.claude/commands"

    # Act: backup with no marker
    cd "$project_dir"
    if backup_team_resource "commands" ".claude/commands" ".team-commands" "f"; then
        # Assert: no backup directory created
        if [[ ! -d "$project_dir/.claude/commands.backup" ]]; then
            test_pass "returned 0 and did not create backup"
        else
            test_fail "backup_team_resource" "no backup directory" "backup directory created"
        fi
    else
        test_fail "backup_team_resource" "return 0" "non-zero return"
    fi
}

test_backup_team_resource_removes_old_backup() {
    run_test "backup_team_resource removes old backup before creating new one"

    # Setup: create old backup with old files
    local cmd_dir="$TEST_TMP/project/.claude/commands"
    mkdir -p "$cmd_dir.backup"
    echo "old content" > "$cmd_dir.backup/old-file.md"

    # Setup: create new commands to backup
    mkdir -p "$cmd_dir"
    echo "new-cmd.md" > "$cmd_dir/.team-commands"
    echo "new content" > "$cmd_dir/new-cmd.md"

    # Act: backup (should remove old backup first)
    cd "$TEST_TMP/project"
    backup_team_resource "commands" ".claude/commands" ".team-commands" "f"

    # Assert: old file gone, new file present
    if [[ ! -f "$cmd_dir.backup/old-file.md" ]] && \
       [[ -f "$cmd_dir.backup/new-cmd.md" ]]; then
        test_pass "removed old backup and created new one"
    else
        test_fail "backup_team_resource" "clean backup directory" "old files still present"
    fi
}

# ============================================================================
# Tests for remove_team_resource() - RF-004
# ============================================================================

test_remove_team_resource_commands() {
    run_test "remove_team_resource removes commands (files)"

    # Setup: create commands with marker
    local project_dir="$TEST_TMP/project-remove-commands"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "cmd-a.md" > "$cmd_dir/.team-commands"
    echo "cmd-b.md" >> "$cmd_dir/.team-commands"
    echo "content" > "$cmd_dir/cmd-a.md"
    echo "content" > "$cmd_dir/cmd-b.md"

    # Act: remove commands
    cd "$project_dir"
    remove_team_resource "commands" ".claude/commands" ".team-commands" "f"

    # Assert: files and marker removed
    if [[ ! -f "$cmd_dir/cmd-a.md" ]] && \
       [[ ! -f "$cmd_dir/cmd-b.md" ]] && \
       [[ ! -f "$cmd_dir/.team-commands" ]]; then
        test_pass "removed command files and marker"
    else
        test_fail "remove_team_resource" "all files removed" "files still present"
    fi
}

test_remove_team_resource_skills() {
    run_test "remove_team_resource removes skills (directories)"

    # Setup: create skills with marker
    local project_dir="$TEST_TMP/project-remove-skills"
    local skill_dir="$project_dir/.claude/skills"
    mkdir -p "$skill_dir/skill-a/subdir"
    mkdir -p "$skill_dir/skill-b"
    echo "skill-a" > "$skill_dir/.team-skills"
    echo "skill-b" >> "$skill_dir/.team-skills"
    echo "content" > "$skill_dir/skill-a/skill.md"

    # Act: remove skills
    cd "$project_dir"
    remove_team_resource "skills" ".claude/skills" ".team-skills" "d"

    # Assert: directories and marker removed
    if [[ ! -d "$skill_dir/skill-a" ]] && \
       [[ ! -d "$skill_dir/skill-b" ]] && \
       [[ ! -f "$skill_dir/.team-skills" ]]; then
        test_pass "removed skill directories and marker"
    else
        test_fail "remove_team_resource" "all directories removed" "directories still present"
    fi
}

test_remove_team_resource_hooks() {
    run_test "remove_team_resource removes hooks (files)"

    # Setup: create hooks with marker
    local project_dir="$TEST_TMP/project-remove-hooks"
    local hook_dir="$project_dir/.claude/hooks"
    mkdir -p "$hook_dir"
    echo "hook-a.sh" > "$hook_dir/.team-hooks"
    echo "#!/bin/bash" > "$hook_dir/hook-a.sh"

    # Act: remove hooks
    cd "$project_dir"
    remove_team_resource "hooks" ".claude/hooks" ".team-hooks" "f"

    # Assert: file and marker removed
    if [[ ! -f "$hook_dir/hook-a.sh" ]] && \
       [[ ! -f "$hook_dir/.team-hooks" ]]; then
        test_pass "removed hook file and marker"
    else
        test_fail "remove_team_resource" "all files removed" "files still present"
    fi
}

test_remove_team_resource_no_marker() {
    run_test "remove_team_resource returns 0 when no marker file exists"

    # Setup: create directory without marker
    local project_dir="$TEST_TMP/project-remove-no-marker"
    mkdir -p "$project_dir/.claude/commands"

    # Act: remove with no marker
    cd "$project_dir"
    if remove_team_resource "commands" ".claude/commands" ".team-commands" "f"; then
        test_pass "returned 0 when no marker present"
    else
        test_fail "remove_team_resource" "return 0" "non-zero return"
    fi
}

test_remove_team_resource_removes_marker() {
    run_test "remove_team_resource removes marker file after resources"

    # Setup: create command with marker
    local project_dir="$TEST_TMP/project-remove-marker"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "cmd-a.md" > "$cmd_dir/.team-commands"
    echo "content" > "$cmd_dir/cmd-a.md"

    # Act: remove
    cd "$project_dir"
    remove_team_resource "commands" ".claude/commands" ".team-commands" "f"

    # Assert: marker file is gone
    if [[ ! -f "$cmd_dir/.team-commands" ]]; then
        test_pass "marker file removed"
    else
        test_fail "remove_team_resource" "marker removed" "marker still present"
    fi
}

# ============================================================================
# Tests for detect_resource_orphans() - RF-005
# ============================================================================

test_detect_resource_orphans_commands() {
    run_test "detect_resource_orphans detects command orphans"

    # Setup: project with orphan command from team-a, swapping to team-b
    local project_dir="$TEST_TMP/project-detect-orphans-cmd"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "orphan from team-a" > "$cmd_dir/cmd-a.md"

    # Act: detect orphans (incoming team is team-b)
    cd "$project_dir"
    local result
    result=$(detect_resource_orphans "commands" ".claude/commands" "team-b" "f" "*.md")

    # Assert: should detect cmd-a.md as orphan from team-a
    if [[ "$result" == "cmd-a.md:team-a" ]]; then
        test_pass "detected command orphan with correct origin team"
    else
        test_fail "detect_resource_orphans" "cmd-a.md:team-a" "$result"
    fi
}

test_detect_resource_orphans_skills() {
    run_test "detect_resource_orphans detects skill orphans"

    # Setup: project with orphan skill from team-b, swapping to team-a
    local project_dir="$TEST_TMP/project-detect-orphans-skill"
    local skill_dir="$project_dir/.claude/skills"
    mkdir -p "$skill_dir/skill-b"
    echo "orphan skill" > "$skill_dir/skill-b/skill.md"

    # Act: detect orphans (incoming team is team-a)
    cd "$project_dir"
    local result
    result=$(detect_resource_orphans "skills" ".claude/skills" "team-a" "d" "*/")

    # Assert: should detect skill-b as orphan from team-b
    if [[ "$result" == "skill-b:team-b" ]]; then
        test_pass "detected skill orphan with correct origin team"
    else
        test_fail "detect_resource_orphans" "skill-b:team-b" "$result"
    fi
}

test_detect_resource_orphans_hooks() {
    run_test "detect_resource_orphans detects hook orphans"

    # Setup: project with orphan hook from team-a, swapping to team-b
    local project_dir="$TEST_TMP/project-detect-orphans-hook"
    local hook_dir="$project_dir/.claude/hooks"
    mkdir -p "$hook_dir"
    echo "orphan hook" > "$hook_dir/hook-a.sh"

    # Act: detect orphans (incoming team is team-b)
    cd "$project_dir"
    local result
    result=$(detect_resource_orphans "hooks" ".claude/hooks" "team-b" "f" "*")

    # Assert: should detect hook-a.sh as orphan from team-a
    if [[ "$result" == "hook-a.sh:team-a" ]]; then
        test_pass "detected hook orphan with correct origin team"
    else
        test_fail "detect_resource_orphans" "hook-a.sh:team-a" "$result"
    fi
}

test_detect_resource_orphans_skips_incoming() {
    run_test "detect_resource_orphans skips resources from incoming team"

    # Setup: project with command from team-a, swapping to team-a (should not be orphan)
    local project_dir="$TEST_TMP/project-detect-no-orphan"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "team-a command" > "$cmd_dir/cmd-a.md"

    # Act: detect orphans (incoming team is team-a - same as resource)
    cd "$project_dir"
    local result
    result=$(detect_resource_orphans "commands" ".claude/commands" "team-a" "f" "*.md")

    # Assert: should return empty (no orphans)
    if [[ -z "$result" ]]; then
        test_pass "correctly skipped resource from incoming team"
    else
        test_fail "detect_resource_orphans" "(empty)" "$result"
    fi
}

test_detect_resource_orphans_skips_non_team() {
    run_test "detect_resource_orphans skips non-team resources"

    # Setup: project with user-created command (not from any team)
    local project_dir="$TEST_TMP/project-detect-user-cmd"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "user command" > "$cmd_dir/user-custom.md"

    # Act: detect orphans (incoming team is team-a)
    cd "$project_dir"
    local result
    result=$(detect_resource_orphans "commands" ".claude/commands" "team-a" "f" "*.md")

    # Assert: should return empty (user commands are not orphans)
    if [[ -z "$result" ]]; then
        test_pass "correctly skipped non-team resource"
    else
        test_fail "detect_resource_orphans" "(empty)" "$result"
    fi
}

test_detect_resource_orphans_multiple() {
    run_test "detect_resource_orphans detects multiple orphans"

    # Setup: project with orphans from both teams, swapping to new team
    local project_dir="$TEST_TMP/project-detect-multiple"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "from team-a" > "$cmd_dir/cmd-a.md"
    echo "from team-b" > "$cmd_dir/cmd-b.md"
    echo "user command" > "$cmd_dir/user.md"

    # Act: detect orphans (incoming team is neither team-a nor team-b)
    # Create a third team for this test
    mkdir -p "$TEST_TMP/mock-roster/teams/team-c/commands"
    cd "$project_dir"
    local result
    result=$(detect_resource_orphans "commands" ".claude/commands" "team-c" "f" "*.md")

    # Assert: should detect both cmd-a.md and cmd-b.md (but not user.md)
    local count
    count=$(echo "$result" | grep -c "^cmd-[ab]\.md:team-[ab]$" || true)
    if [[ $count -eq 2 ]]; then
        test_pass "detected multiple orphans, skipped non-team resource"
    else
        test_fail "detect_resource_orphans" "2 orphans" "$count orphans found: $result"
    fi
}

test_detect_resource_orphans_no_directory() {
    run_test "detect_resource_orphans returns empty when resource directory doesn't exist"

    # Setup: project without commands directory
    local project_dir="$TEST_TMP/project-no-cmd-dir"
    mkdir -p "$project_dir"

    # Act: detect orphans on non-existent directory
    cd "$project_dir"
    local result
    result=$(detect_resource_orphans "commands" ".claude/commands" "team-a" "f" "*.md")

    # Assert: should return empty
    if [[ -z "$result" ]]; then
        test_pass "returned empty for non-existent directory"
    else
        test_fail "detect_resource_orphans" "(empty)" "$result"
    fi
}

# ============================================================================
# Main test runner
# ============================================================================

main() {
    echo "========================================"
    echo "Team Resource Unit Tests"
    echo "========================================"
    echo ""

    setup

    # Run all tests
    test_is_resource_from_team_command
    test_is_resource_from_team_skill
    test_is_resource_from_team_hook
    test_is_resource_from_team_not_found
    test_is_resource_from_team_wrong_type

    test_get_resource_team_command
    test_get_resource_team_skill
    test_get_resource_team_hook
    test_get_resource_team_not_found
    test_get_resource_team_multiple_teams

    # RF-003: backup_team_resource tests
    test_backup_team_resource_commands
    test_backup_team_resource_skills
    test_backup_team_resource_hooks
    test_backup_team_resource_no_marker
    test_backup_team_resource_removes_old_backup

    # RF-004: remove_team_resource tests
    test_remove_team_resource_commands
    test_remove_team_resource_skills
    test_remove_team_resource_hooks
    test_remove_team_resource_no_marker
    test_remove_team_resource_removes_marker

    # RF-005: detect_resource_orphans tests
    test_detect_resource_orphans_commands
    test_detect_resource_orphans_skills
    test_detect_resource_orphans_hooks
    test_detect_resource_orphans_skips_incoming
    test_detect_resource_orphans_skips_non_team
    test_detect_resource_orphans_multiple
    test_detect_resource_orphans_no_directory

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

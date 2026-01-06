#!/usr/bin/env bash
#
# test-rite-resource.sh - Unit tests for rite-resource.sh
#
# Tests generic rite resource operations including membership checks,
# backup, removal, and orphan detection.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KNOSSOS_HOME="${KNOSSOS_HOME:-$(cd "$SCRIPT_DIR/../../.." && pwd)}"

# Source dependencies
source "$KNOSSOS_HOME/lib/rite/rite-resource.sh"

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

    # Create mock rite structure for testing
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-a/commands"
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-a/skills"
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-a/hooks"
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-b/commands"
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-b/skills"
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-b/hooks"

    # Create mock commands (files)
    touch "$TEST_TMP/mock-knossos/rites/team-a/commands/cmd-a.md"
    touch "$TEST_TMP/mock-knossos/rites/team-b/commands/cmd-b.md"

    # Create mock skills (directories)
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-a/skills/skill-a"
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-b/skills/skill-b"

    # Create mock hooks (files)
    touch "$TEST_TMP/mock-knossos/rites/team-a/hooks/hook-a.sh"
    touch "$TEST_TMP/mock-knossos/rites/team-b/hooks/hook-b.sh"

    # Override KNOSSOS_HOME for tests
    KNOSSOS_HOME="$TEST_TMP/mock-knossos"
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ============================================================================
# Tests for is_resource_from_rite()
# ============================================================================

test_is_resource_from_rite_command() {
    run_test "is_resource_from_rite finds command file"

    if is_resource_from_rite "cmd-a.md" "commands" "f"; then
        test_pass "found team-a command"
    else
        test_fail "is_resource_from_rite" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_rite_skill() {
    run_test "is_resource_from_rite finds skill directory"

    if is_resource_from_rite "skill-a" "skills" "d"; then
        test_pass "found team-a skill"
    else
        test_fail "is_resource_from_rite" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_rite_hook() {
    run_test "is_resource_from_rite finds hook file"

    if is_resource_from_rite "hook-b.sh" "hooks" "f"; then
        test_pass "found team-b hook"
    else
        test_fail "is_resource_from_rite" "success (return 0)" "failure (return 1)"
    fi
}

test_is_resource_from_rite_not_found() {
    run_test "is_resource_from_rite returns false for non-team resource"

    if is_resource_from_rite "nonexistent.md" "commands" "f"; then
        test_fail "is_resource_from_rite" "failure (return 1)" "success (return 0)"
    else
        test_pass "correctly returned false for nonexistent resource"
    fi
}

test_is_resource_from_rite_wrong_type() {
    run_test "is_resource_from_rite returns false when find type doesn't match"

    # skill-a is a directory, but we're looking for a file
    if is_resource_from_rite "skill-a" "skills" "f"; then
        test_fail "is_resource_from_rite" "failure (return 1)" "success (return 0)"
    else
        test_pass "correctly returned false when type doesn't match"
    fi
}

# ============================================================================
# Tests for get_resource_rite()
# ============================================================================

test_get_resource_rite_command() {
    run_test "get_resource_rite returns correct team for command"

    local result
    result=$(get_resource_rite "cmd-a.md" "commands" "f")

    if [[ "$result" == "team-a" ]]; then
        test_pass "returned correct team name: team-a"
    else
        test_fail "get_resource_rite" "team-a" "$result"
    fi
}

test_get_resource_rite_skill() {
    run_test "get_resource_rite returns correct team for skill"

    local result
    result=$(get_resource_rite "skill-b" "skills" "d")

    if [[ "$result" == "team-b" ]]; then
        test_pass "returned correct team name: team-b"
    else
        test_fail "get_resource_rite" "team-b" "$result"
    fi
}

test_get_resource_rite_hook() {
    run_test "get_resource_rite returns correct team for hook"

    local result
    result=$(get_resource_rite "hook-a.sh" "hooks" "f")

    if [[ "$result" == "team-a" ]]; then
        test_pass "returned correct team name: team-a"
    else
        test_fail "get_resource_rite" "team-a" "$result"
    fi
}

test_get_resource_rite_not_found() {
    run_test "get_resource_rite returns empty for non-team resource"

    local result
    result=$(get_resource_rite "nonexistent.md" "commands" "f")

    if [[ -z "$result" ]]; then
        test_pass "returned empty string for nonexistent resource"
    else
        test_fail "get_resource_rite" "(empty)" "$result"
    fi
}

test_get_resource_rite_multiple_teams() {
    run_test "get_resource_rite returns first match when resource exists in multiple teams"

    # Create same command in both teams
    touch "$TEST_TMP/mock-knossos/rites/team-a/commands/shared.md"
    touch "$TEST_TMP/mock-knossos/rites/team-b/commands/shared.md"

    local result
    result=$(get_resource_rite "shared.md" "commands" "f")

    # Should return one of them (behavior: first match from find)
    if [[ "$result" == "team-a" ]] || [[ "$result" == "team-b" ]]; then
        test_pass "returned a team name: $result"
    else
        test_fail "get_resource_rite" "team-a or team-b" "$result"
    fi
}

# ============================================================================
# Tests for backup_rite_resource() - RF-003
# ============================================================================

test_backup_rite_resource_commands() {
    run_test "backup_rite_resource backs up commands (files)"

    # Setup: create commands directory with marker
    local cmd_dir="$TEST_TMP/project/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "cmd-a.md" > "$cmd_dir/.rite-commands"
    echo "cmd-b.md" >> "$cmd_dir/.rite-commands"
    echo "test command content" > "$cmd_dir/cmd-a.md"
    echo "another command" > "$cmd_dir/cmd-b.md"

    # Act: backup commands
    cd "$TEST_TMP/project"
    backup_rite_resource "commands" ".claude/commands" ".rite-commands" "f"

    # Assert: backup directory exists with files
    if [[ -d "$cmd_dir.backup" ]] && \
       [[ -f "$cmd_dir.backup/cmd-a.md" ]] && \
       [[ -f "$cmd_dir.backup/cmd-b.md" ]]; then
        test_pass "backed up command files to .backup directory"
    else
        test_fail "backup_rite_resource" "backup directory with files" "missing files or directory"
    fi
}

test_backup_rite_resource_skills() {
    run_test "backup_rite_resource backs up skills (directories)"

    # Setup: create skills directory with marker
    local skill_dir="$TEST_TMP/project/.claude/skills"
    mkdir -p "$skill_dir/skill-a/subdir"
    mkdir -p "$skill_dir/skill-b"
    echo "skill-a" > "$skill_dir/.rite-skills"
    echo "skill-b" >> "$skill_dir/.rite-skills"
    echo "content" > "$skill_dir/skill-a/skill.md"
    echo "nested" > "$skill_dir/skill-a/subdir/file.txt"

    # Act: backup skills
    cd "$TEST_TMP/project"
    backup_rite_resource "skills" ".claude/skills" ".rite-skills" "d"

    # Assert: backup directory exists with recursive copy
    if [[ -d "$skill_dir.backup/skill-a" ]] && \
       [[ -d "$skill_dir.backup/skill-b" ]] && \
       [[ -f "$skill_dir.backup/skill-a/skill.md" ]] && \
       [[ -f "$skill_dir.backup/skill-a/subdir/file.txt" ]]; then
        test_pass "backed up skill directories recursively"
    else
        test_fail "backup_rite_resource" "backup directory with recursive structure" "missing structure"
    fi
}

test_backup_rite_resource_hooks() {
    run_test "backup_rite_resource backs up hooks (files)"

    # Setup: create hooks directory with marker
    local hook_dir="$TEST_TMP/project/.claude/hooks"
    mkdir -p "$hook_dir"
    echo "hook-a.sh" > "$hook_dir/.rite-hooks"
    echo "#!/bin/bash" > "$hook_dir/hook-a.sh"

    # Act: backup hooks
    cd "$TEST_TMP/project"
    backup_rite_resource "hooks" ".claude/hooks" ".rite-hooks" "f"

    # Assert: backup directory exists with file
    if [[ -d "$hook_dir.backup" ]] && \
       [[ -f "$hook_dir.backup/hook-a.sh" ]]; then
        test_pass "backed up hook files"
    else
        test_fail "backup_rite_resource" "backup directory with hook file" "missing file or directory"
    fi
}

test_backup_rite_resource_no_marker() {
    run_test "backup_rite_resource returns 0 when no marker file exists"

    # Setup: create directory without marker (use unique subdir)
    local project_dir="$TEST_TMP/project-no-marker"
    mkdir -p "$project_dir/.claude/commands"

    # Act: backup with no marker
    cd "$project_dir"
    if backup_rite_resource "commands" ".claude/commands" ".rite-commands" "f"; then
        # Assert: no backup directory created
        if [[ ! -d "$project_dir/.claude/commands.backup" ]]; then
            test_pass "returned 0 and did not create backup"
        else
            test_fail "backup_rite_resource" "no backup directory" "backup directory created"
        fi
    else
        test_fail "backup_rite_resource" "return 0" "non-zero return"
    fi
}

test_backup_rite_resource_removes_old_backup() {
    run_test "backup_rite_resource removes old backup before creating new one"

    # Setup: create old backup with old files
    local cmd_dir="$TEST_TMP/project/.claude/commands"
    mkdir -p "$cmd_dir.backup"
    echo "old content" > "$cmd_dir.backup/old-file.md"

    # Setup: create new commands to backup
    mkdir -p "$cmd_dir"
    echo "new-cmd.md" > "$cmd_dir/.rite-commands"
    echo "new content" > "$cmd_dir/new-cmd.md"

    # Act: backup (should remove old backup first)
    cd "$TEST_TMP/project"
    backup_rite_resource "commands" ".claude/commands" ".rite-commands" "f"

    # Assert: old file gone, new file present
    if [[ ! -f "$cmd_dir.backup/old-file.md" ]] && \
       [[ -f "$cmd_dir.backup/new-cmd.md" ]]; then
        test_pass "removed old backup and created new one"
    else
        test_fail "backup_rite_resource" "clean backup directory" "old files still present"
    fi
}

# ============================================================================
# Tests for remove_rite_resource() - RF-004
# ============================================================================

test_remove_rite_resource_commands() {
    run_test "remove_rite_resource removes commands (files)"

    # Setup: create commands with marker
    local project_dir="$TEST_TMP/project-remove-commands"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "cmd-a.md" > "$cmd_dir/.rite-commands"
    echo "cmd-b.md" >> "$cmd_dir/.rite-commands"
    echo "content" > "$cmd_dir/cmd-a.md"
    echo "content" > "$cmd_dir/cmd-b.md"

    # Act: remove commands
    cd "$project_dir"
    remove_rite_resource "commands" ".claude/commands" ".rite-commands" "f"

    # Assert: files and marker removed
    if [[ ! -f "$cmd_dir/cmd-a.md" ]] && \
       [[ ! -f "$cmd_dir/cmd-b.md" ]] && \
       [[ ! -f "$cmd_dir/.rite-commands" ]]; then
        test_pass "removed command files and marker"
    else
        test_fail "remove_rite_resource" "all files removed" "files still present"
    fi
}

test_remove_rite_resource_skills() {
    run_test "remove_rite_resource removes skills (directories)"

    # Setup: create skills with marker
    local project_dir="$TEST_TMP/project-remove-skills"
    local skill_dir="$project_dir/.claude/skills"
    mkdir -p "$skill_dir/skill-a/subdir"
    mkdir -p "$skill_dir/skill-b"
    echo "skill-a" > "$skill_dir/.rite-skills"
    echo "skill-b" >> "$skill_dir/.rite-skills"
    echo "content" > "$skill_dir/skill-a/skill.md"

    # Act: remove skills
    cd "$project_dir"
    remove_rite_resource "skills" ".claude/skills" ".rite-skills" "d"

    # Assert: directories and marker removed
    if [[ ! -d "$skill_dir/skill-a" ]] && \
       [[ ! -d "$skill_dir/skill-b" ]] && \
       [[ ! -f "$skill_dir/.rite-skills" ]]; then
        test_pass "removed skill directories and marker"
    else
        test_fail "remove_rite_resource" "all directories removed" "directories still present"
    fi
}

test_remove_rite_resource_hooks() {
    run_test "remove_rite_resource removes hooks (files)"

    # Setup: create hooks with marker
    local project_dir="$TEST_TMP/project-remove-hooks"
    local hook_dir="$project_dir/.claude/hooks"
    mkdir -p "$hook_dir"
    echo "hook-a.sh" > "$hook_dir/.rite-hooks"
    echo "#!/bin/bash" > "$hook_dir/hook-a.sh"

    # Act: remove hooks
    cd "$project_dir"
    remove_rite_resource "hooks" ".claude/hooks" ".rite-hooks" "f"

    # Assert: file and marker removed
    if [[ ! -f "$hook_dir/hook-a.sh" ]] && \
       [[ ! -f "$hook_dir/.rite-hooks" ]]; then
        test_pass "removed hook file and marker"
    else
        test_fail "remove_rite_resource" "all files removed" "files still present"
    fi
}

test_remove_rite_resource_no_marker() {
    run_test "remove_rite_resource returns 0 when no marker file exists"

    # Setup: create directory without marker
    local project_dir="$TEST_TMP/project-remove-no-marker"
    mkdir -p "$project_dir/.claude/commands"

    # Act: remove with no marker
    cd "$project_dir"
    if remove_rite_resource "commands" ".claude/commands" ".rite-commands" "f"; then
        test_pass "returned 0 when no marker present"
    else
        test_fail "remove_rite_resource" "return 0" "non-zero return"
    fi
}

test_remove_rite_resource_removes_marker() {
    run_test "remove_rite_resource removes marker file after resources"

    # Setup: create command with marker
    local project_dir="$TEST_TMP/project-remove-marker"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "cmd-a.md" > "$cmd_dir/.rite-commands"
    echo "content" > "$cmd_dir/cmd-a.md"

    # Act: remove
    cd "$project_dir"
    remove_rite_resource "commands" ".claude/commands" ".rite-commands" "f"

    # Assert: marker file is gone
    if [[ ! -f "$cmd_dir/.rite-commands" ]]; then
        test_pass "marker file removed"
    else
        test_fail "remove_rite_resource" "marker removed" "marker still present"
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
    mkdir -p "$TEST_TMP/mock-knossos/rites/team-c/commands"
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
# Tests for remove_resource_orphans() - RF-006
# ============================================================================

test_remove_resource_orphans_remove_mode_commands() {
    run_test "remove_resource_orphans removes and backs up commands in remove mode"

    # Setup: project with orphan command
    local project_dir="$TEST_TMP/project-remove-orphan-cmd"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "orphan content" > "$cmd_dir/cmd-a.md"

    # Act: pipe orphan list to remove function (remove mode)
    cd "$project_dir"
    echo "cmd-a.md:team-a" | remove_resource_orphans "commands" ".claude/commands" "remove" "f" 2>/dev/null

    # Assert: file removed, backup created
    if [[ ! -f "$cmd_dir/cmd-a.md" ]] && \
       [[ -f "$cmd_dir.orphan-backup/cmd-a.md" ]]; then
        test_pass "removed command and created backup"
    else
        test_fail "remove_resource_orphans" "file removed, backup exists" "file still present or backup missing"
    fi
}

test_remove_resource_orphans_remove_mode_skills() {
    run_test "remove_resource_orphans removes and backs up skills in remove mode"

    # Setup: project with orphan skill
    local project_dir="$TEST_TMP/project-remove-orphan-skill"
    local skill_dir="$project_dir/.claude/skills"
    mkdir -p "$skill_dir/skill-b/subdir"
    echo "orphan skill content" > "$skill_dir/skill-b/skill.md"
    echo "nested file" > "$skill_dir/skill-b/subdir/nested.txt"

    # Act: pipe orphan list to remove function (remove mode)
    cd "$project_dir"
    echo "skill-b:team-b" | remove_resource_orphans "skills" ".claude/skills" "remove" "d" 2>/dev/null

    # Assert: directory removed, backup created with recursive structure
    if [[ ! -d "$skill_dir/skill-b" ]] && \
       [[ -d "$skill_dir.orphan-backup/skill-b" ]] && \
       [[ -f "$skill_dir.orphan-backup/skill-b/skill.md" ]] && \
       [[ -f "$skill_dir.orphan-backup/skill-b/subdir/nested.txt" ]]; then
        test_pass "removed skill directory and created recursive backup"
    else
        test_fail "remove_resource_orphans" "directory removed, backup with structure" "directory still present or backup incomplete"
    fi
}

test_remove_resource_orphans_remove_mode_hooks() {
    run_test "remove_resource_orphans removes and backs up hooks in remove mode"

    # Setup: project with orphan hook
    local project_dir="$TEST_TMP/project-remove-orphan-hook"
    local hook_dir="$project_dir/.claude/hooks"
    mkdir -p "$hook_dir"
    echo "#!/bin/bash" > "$hook_dir/hook-a.sh"

    # Act: pipe orphan list to remove function (remove mode)
    cd "$project_dir"
    echo "hook-a.sh:team-a" | remove_resource_orphans "hooks" ".claude/hooks" "remove" "f" 2>/dev/null

    # Assert: file removed, backup created
    if [[ ! -f "$hook_dir/hook-a.sh" ]] && \
       [[ -f "$hook_dir.orphan-backup/hook-a.sh" ]]; then
        test_pass "removed hook and created backup"
    else
        test_fail "remove_resource_orphans" "file removed, backup exists" "file still present or backup missing"
    fi
}

test_remove_resource_orphans_keep_mode() {
    run_test "remove_resource_orphans preserves resources in keep mode"

    # Setup: project with orphan command
    local project_dir="$TEST_TMP/project-keep-orphan"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "keep this" > "$cmd_dir/cmd-a.md"

    # Act: pipe orphan list to remove function (keep mode)
    cd "$project_dir"
    echo "cmd-a.md:team-a" | remove_resource_orphans "commands" ".claude/commands" "keep" "f" 2>/dev/null

    # Assert: file still exists, no backup created
    if [[ -f "$cmd_dir/cmd-a.md" ]] && \
       [[ ! -d "$cmd_dir.orphan-backup" ]]; then
        test_pass "preserved resource, no backup created"
    else
        test_fail "remove_resource_orphans" "file preserved, no backup" "file removed or backup created"
    fi
}

test_remove_resource_orphans_empty_mode() {
    run_test "remove_resource_orphans preserves resources when mode is empty"

    # Setup: project with orphan command
    local project_dir="$TEST_TMP/project-empty-mode"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "keep this too" > "$cmd_dir/cmd-b.md"

    # Act: pipe orphan list to remove function (empty mode)
    cd "$project_dir"
    echo "cmd-b.md:team-b" | remove_resource_orphans "commands" ".claude/commands" "" "f" 2>/dev/null

    # Assert: file still exists, no backup created
    if [[ -f "$cmd_dir/cmd-b.md" ]] && \
       [[ ! -d "$cmd_dir.orphan-backup" ]]; then
        test_pass "preserved resource with empty mode"
    else
        test_fail "remove_resource_orphans" "file preserved, no backup" "file removed or backup created"
    fi
}

test_remove_resource_orphans_multiple() {
    run_test "remove_resource_orphans handles multiple orphans"

    # Setup: project with multiple orphan commands
    local project_dir="$TEST_TMP/project-remove-multiple"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "orphan 1" > "$cmd_dir/cmd-a.md"
    echo "orphan 2" > "$cmd_dir/cmd-b.md"
    echo "orphan 3" > "$cmd_dir/other.md"

    # Act: pipe multiple orphans to remove function
    cd "$project_dir"
    {
        echo "cmd-a.md:team-a"
        echo "cmd-b.md:team-b"
        echo "other.md:team-a"
    } | remove_resource_orphans "commands" ".claude/commands" "remove" "f" 2>/dev/null

    # Assert: all files removed, all backed up
    if [[ ! -f "$cmd_dir/cmd-a.md" ]] && \
       [[ ! -f "$cmd_dir/cmd-b.md" ]] && \
       [[ ! -f "$cmd_dir/other.md" ]] && \
       [[ -f "$cmd_dir.orphan-backup/cmd-a.md" ]] && \
       [[ -f "$cmd_dir.orphan-backup/cmd-b.md" ]] && \
       [[ -f "$cmd_dir.orphan-backup/other.md" ]]; then
        test_pass "removed and backed up all orphans"
    else
        test_fail "remove_resource_orphans" "all removed, all backed up" "some files remain or backups missing"
    fi
}

test_remove_resource_orphans_empty_input() {
    run_test "remove_resource_orphans handles empty input gracefully"

    # Setup: project with commands
    local project_dir="$TEST_TMP/project-empty-input"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "content" > "$cmd_dir/cmd-a.md"

    # Act: pipe empty input to remove function
    cd "$project_dir"
    echo "" | remove_resource_orphans "commands" ".claude/commands" "remove" "f" 2>/dev/null

    # Assert: no changes, no backup directory created
    if [[ -f "$cmd_dir/cmd-a.md" ]] && \
       [[ ! -d "$cmd_dir.orphan-backup" ]]; then
        test_pass "handled empty input without creating backup"
    else
        test_fail "remove_resource_orphans" "no changes" "unexpected changes or backup created"
    fi
}

test_remove_resource_orphans_piped_from_detect() {
    run_test "remove_resource_orphans works piped from detect_resource_orphans"

    # Setup: project with orphan from team-a, swapping to team-b
    local project_dir="$TEST_TMP/project-pipe-test"
    local cmd_dir="$project_dir/.claude/commands"
    mkdir -p "$cmd_dir"
    echo "orphan from team-a" > "$cmd_dir/cmd-a.md"

    # Act: pipe detect output directly to remove
    cd "$project_dir"
    detect_resource_orphans "commands" ".claude/commands" "team-b" "f" "*.md" \
        | remove_resource_orphans "commands" ".claude/commands" "remove" "f" 2>/dev/null

    # Assert: orphan detected and removed
    if [[ ! -f "$cmd_dir/cmd-a.md" ]] && \
       [[ -f "$cmd_dir.orphan-backup/cmd-a.md" ]]; then
        test_pass "piped workflow works end-to-end"
    else
        test_fail "remove_resource_orphans" "orphan removed via pipe" "orphan still present or backup missing"
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
    test_is_resource_from_rite_command
    test_is_resource_from_rite_skill
    test_is_resource_from_rite_hook
    test_is_resource_from_rite_not_found
    test_is_resource_from_rite_wrong_type

    test_get_resource_rite_command
    test_get_resource_rite_skill
    test_get_resource_rite_hook
    test_get_resource_rite_not_found
    test_get_resource_rite_multiple_teams

    # RF-003: backup_rite_resource tests
    test_backup_rite_resource_commands
    test_backup_rite_resource_skills
    test_backup_rite_resource_hooks
    test_backup_rite_resource_no_marker
    test_backup_rite_resource_removes_old_backup

    # RF-004: remove_rite_resource tests
    test_remove_rite_resource_commands
    test_remove_rite_resource_skills
    test_remove_rite_resource_hooks
    test_remove_rite_resource_no_marker
    test_remove_rite_resource_removes_marker

    # RF-005: detect_resource_orphans tests
    test_detect_resource_orphans_commands
    test_detect_resource_orphans_skills
    test_detect_resource_orphans_hooks
    test_detect_resource_orphans_skips_incoming
    test_detect_resource_orphans_skips_non_team
    test_detect_resource_orphans_multiple
    test_detect_resource_orphans_no_directory

    # RF-006: remove_resource_orphans tests
    test_remove_resource_orphans_remove_mode_commands
    test_remove_resource_orphans_remove_mode_skills
    test_remove_resource_orphans_remove_mode_hooks
    test_remove_resource_orphans_keep_mode
    test_remove_resource_orphans_empty_mode
    test_remove_resource_orphans_multiple
    test_remove_resource_orphans_empty_input
    test_remove_resource_orphans_piped_from_detect

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

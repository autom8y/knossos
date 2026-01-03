#!/usr/bin/env bash
#
# test-team-hooks-registration.sh - Unit tests for team-hooks-registration.sh
#
# Tests hook registration including YAML parsing, JSON generation,
# and user hook preservation.

set -euo pipefail

# Test setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROSTER_HOME="${ROSTER_HOME:-$(cd "$SCRIPT_DIR/../../.." && pwd)}"

# Source dependencies
source "$ROSTER_HOME/lib/team/team-hooks-registration.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test temp directory
TEST_TMP=""

# Mock logging functions (team-hooks-registration.sh expects these)
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

    # Create test fixtures directory
    mkdir -p "$TEST_TMP/fixtures"

    # Create valid hooks.yaml
    cat > "$TEST_TMP/fixtures/valid-hooks.yaml" <<'EOF'
schema_version: "1.0"
hooks:
  - event: SessionStart
    path: session-start.sh
    timeout: 5
  - event: PostToolUse
    matcher: Write|Edit
    path: post-write.sh
    timeout: 10
  - event: Stop
    path: session-stop.sh
    timeout: 5
EOF

    # Create hooks.yaml with invalid event
    cat > "$TEST_TMP/fixtures/invalid-event.yaml" <<'EOF'
hooks:
  - event: InvalidEvent
    path: invalid.sh
    timeout: 5
EOF

    # Create hooks.yaml with PostToolUse without matcher
    cat > "$TEST_TMP/fixtures/no-matcher.yaml" <<'EOF'
hooks:
  - event: PostToolUse
    path: post-tool.sh
    timeout: 5
EOF

    # Create hooks.yaml with invalid regex
    cat > "$TEST_TMP/fixtures/bad-regex.yaml" <<'EOF'
hooks:
  - event: PostToolUse
    matcher: "[invalid"
    path: bad-regex.sh
    timeout: 5
EOF

    # Create hooks.yaml with timeout exceeding limit
    cat > "$TEST_TMP/fixtures/timeout-exceed.yaml" <<'EOF'
hooks:
  - event: SessionStart
    path: session-start.sh
    timeout: 120
EOF

    # Create settings.local.json with only roster hooks
    cat > "$TEST_TMP/fixtures/settings-roster-only.json" <<'EOF'
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/session-start.sh",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
EOF

    # Create settings.local.json with mixed hooks
    cat > "$TEST_TMP/fixtures/settings-mixed.json" <<'EOF'
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/roster-hook.sh",
            "timeout": 5
          },
          {
            "type": "command",
            "command": "/usr/local/bin/user-hook.sh",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
EOF

    # Create settings.local.json with only user hooks
    cat > "$TEST_TMP/fixtures/settings-user-only.json" <<'EOF'
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/my-hook.sh",
            "timeout": 5
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "~/bin/cleanup.sh",
            "timeout": 15
          }
        ]
      }
    ]
  }
}
EOF

    # Create corrupted JSON
    echo "{ invalid json" > "$TEST_TMP/fixtures/settings-corrupted.json"

    # Override ROSTER_HOME for tests
    ROSTER_HOME="$TEST_TMP"
}

teardown() {
    rm -rf "$TEST_TMP"
}

# ============================================================================
# Tests for require_yq()
# ============================================================================

test_require_yq_installed() {
    run_test "require_yq returns 0 when yq v4+ installed"

    # This test assumes yq is installed in the test environment
    # Skip if not available
    if ! command -v yq &>/dev/null; then
        echo "  SKIP: yq not installed in test environment"
        TESTS_RUN=$((TESTS_RUN - 1))
        return
    fi

    if require_yq 2>/dev/null; then
        test_pass "yq v4+ detected"
    else
        test_fail "require_yq" "return 0" "return 1"
    fi
}

# ============================================================================
# Tests for parse_hooks_yaml()
# ============================================================================

test_parse_hooks_yaml_valid() {
    run_test "parse_hooks_yaml parses valid hooks.yaml"

    local result
    result=$(parse_hooks_yaml "$TEST_TMP/fixtures/valid-hooks.yaml" 2>/dev/null)

    local line_count
    line_count=$(echo "$result" | grep -c '^{' || true)

    if [[ "$line_count" -eq 3 ]]; then
        # Verify first hook
        local first_hook
        first_hook=$(echo "$result" | head -1)
        local event
        event=$(echo "$first_hook" | jq -r '.event')
        if [[ "$event" == "SessionStart" ]]; then
            test_pass "parsed 3 hooks correctly"
        else
            test_fail "parse_hooks_yaml" "SessionStart event" "$event"
        fi
    else
        test_fail "parse_hooks_yaml" "3 hooks" "$line_count hooks"
    fi
}

test_parse_hooks_yaml_missing() {
    run_test "parse_hooks_yaml returns empty for missing file"

    local result
    result=$(parse_hooks_yaml "$TEST_TMP/fixtures/nonexistent.yaml" 2>/dev/null)

    if [[ -z "$result" ]]; then
        test_pass "returned empty for missing file"
    else
        test_fail "parse_hooks_yaml" "empty" "$result"
    fi
}

test_parse_hooks_yaml_invalid_event() {
    run_test "parse_hooks_yaml skips invalid event types"

    local result
    result=$(parse_hooks_yaml "$TEST_TMP/fixtures/invalid-event.yaml" 2>/dev/null)

    if [[ -z "$result" ]]; then
        test_pass "skipped invalid event"
    else
        test_fail "parse_hooks_yaml" "empty (invalid event)" "$result"
    fi
}

test_parse_hooks_yaml_no_matcher() {
    run_test "parse_hooks_yaml skips PostToolUse without matcher"

    local result
    result=$(parse_hooks_yaml "$TEST_TMP/fixtures/no-matcher.yaml" 2>/dev/null)

    if [[ -z "$result" ]]; then
        test_pass "skipped PostToolUse without matcher"
    else
        test_fail "parse_hooks_yaml" "empty (no matcher)" "$result"
    fi
}

test_parse_hooks_yaml_invalid_regex() {
    run_test "parse_hooks_yaml skips invalid matcher regex"

    local result
    result=$(parse_hooks_yaml "$TEST_TMP/fixtures/bad-regex.yaml" 2>/dev/null)

    if [[ -z "$result" ]]; then
        test_pass "skipped invalid regex"
    else
        test_fail "parse_hooks_yaml" "empty (bad regex)" "$result"
    fi
}

test_parse_hooks_yaml_timeout_clamp() {
    run_test "parse_hooks_yaml clamps timeout > 60"

    local result
    result=$(parse_hooks_yaml "$TEST_TMP/fixtures/timeout-exceed.yaml" 2>/dev/null)

    if [[ -n "$result" ]]; then
        local timeout
        timeout=$(echo "$result" | jq -r '.timeout')
        if [[ "$timeout" -eq 60 ]]; then
            test_pass "timeout clamped to 60"
        else
            test_fail "parse_hooks_yaml" "timeout: 60" "timeout: $timeout"
        fi
    else
        test_fail "parse_hooks_yaml" "hook with clamped timeout" "empty result"
    fi
}

# ============================================================================
# Tests for extract_non_roster_hooks()
# ============================================================================

test_extract_non_roster_roster_only() {
    run_test "extract_non_roster_hooks returns {} for roster-only hooks"

    local result
    result=$(extract_non_roster_hooks "$TEST_TMP/fixtures/settings-roster-only.json")

    if [[ "$result" == "{}" ]]; then
        test_pass "returned empty for roster-only hooks"
    else
        test_fail "extract_non_roster_hooks" "{}" "$result"
    fi
}

test_extract_non_roster_mixed() {
    run_test "extract_non_roster_hooks preserves user hooks only"

    local result
    result=$(extract_non_roster_hooks "$TEST_TMP/fixtures/settings-mixed.json")

    # Should have SessionStart with one user hook
    local has_session_start
    has_session_start=$(echo "$result" | jq 'has("SessionStart")')

    if [[ "$has_session_start" == "true" ]]; then
        local hook_count
        hook_count=$(echo "$result" | jq '.SessionStart[0].hooks | length')
        if [[ "$hook_count" -eq 1 ]]; then
            local command
            command=$(echo "$result" | jq -r '.SessionStart[0].hooks[0].command')
            if [[ "$command" == "/usr/local/bin/user-hook.sh" ]]; then
                test_pass "preserved user hook, filtered roster hook"
            else
                test_fail "extract_non_roster_hooks" "/usr/local/bin/user-hook.sh" "$command"
            fi
        else
            test_fail "extract_non_roster_hooks" "1 hook" "$hook_count hooks"
        fi
    else
        test_fail "extract_non_roster_hooks" "SessionStart present" "SessionStart missing"
    fi
}

test_extract_non_roster_user_only() {
    run_test "extract_non_roster_hooks preserves all user hooks"

    local result
    result=$(extract_non_roster_hooks "$TEST_TMP/fixtures/settings-user-only.json")

    # Should have SessionStart and Stop
    local event_count
    event_count=$(echo "$result" | jq 'keys | length')

    if [[ "$event_count" -eq 2 ]]; then
        test_pass "preserved all user hooks for 2 events"
    else
        test_fail "extract_non_roster_hooks" "2 events" "$event_count events"
    fi
}

test_extract_non_roster_missing() {
    run_test "extract_non_roster_hooks returns {} for missing file"

    local result
    result=$(extract_non_roster_hooks "$TEST_TMP/fixtures/nonexistent.json")

    if [[ "$result" == "{}" ]]; then
        test_pass "returned {} for missing file"
    else
        test_fail "extract_non_roster_hooks" "{}" "$result"
    fi
}

# ============================================================================
# Tests for merge_hook_registrations()
# ============================================================================

test_merge_hook_registrations() {
    run_test "merge_hook_registrations combines base and team"

    local base='{"event":"SessionStart","matcher":"","path":"base.sh","timeout":5}'
    local team='{"event":"Stop","matcher":"","path":"team.sh","timeout":5}'

    local result
    result=$(merge_hook_registrations "$base" "$team")

    local line_count
    line_count=$(echo "$result" | grep -c '^{' || true)

    if [[ "$line_count" -eq 2 ]]; then
        # Check order: base first
        local first_path
        first_path=$(echo "$result" | head -1 | jq -r '.path')
        if [[ "$first_path" == "base.sh" ]]; then
            test_pass "merged with base first, team second"
        else
            test_fail "merge_hook_registrations" "base.sh first" "$first_path first"
        fi
    else
        test_fail "merge_hook_registrations" "2 hooks" "$line_count hooks"
    fi
}

test_merge_hook_registrations_empty_base() {
    run_test "merge_hook_registrations handles empty base"

    local team='{"event":"Stop","matcher":"","path":"team.sh","timeout":5}'

    local result
    result=$(merge_hook_registrations "" "$team")

    local line_count
    line_count=$(echo "$result" | grep -c '^{' || true)

    if [[ "$line_count" -eq 1 ]]; then
        test_pass "returned team hooks only"
    else
        test_fail "merge_hook_registrations" "1 hook" "$line_count hooks"
    fi
}

test_merge_hook_registrations_empty_team() {
    run_test "merge_hook_registrations handles empty team"

    local base='{"event":"SessionStart","matcher":"","path":"base.sh","timeout":5}'

    local result
    result=$(merge_hook_registrations "$base" "")

    local line_count
    line_count=$(echo "$result" | grep -c '^{' || true)

    if [[ "$line_count" -eq 1 ]]; then
        test_pass "returned base hooks only"
    else
        test_fail "merge_hook_registrations" "1 hook" "$line_count hooks"
    fi
}

# ============================================================================
# Tests for generate_hooks_json()
# ============================================================================

test_generate_hooks_json_single() {
    run_test "generate_hooks_json generates structure for single hook"

    local registrations='{"event":"SessionStart","matcher":"","path":"session-start.sh","timeout":5}'

    local result
    result=$(generate_hooks_json "$registrations")

    local has_session_start
    has_session_start=$(echo "$result" | jq 'has("SessionStart")')

    if [[ "$has_session_start" == "true" ]]; then
        local command
        command=$(echo "$result" | jq -r '.SessionStart[0].hooks[0].command')
        if [[ "$command" == '$CLAUDE_PROJECT_DIR/.claude/hooks/session-start.sh' ]]; then
            test_pass "generated correct hook structure"
        else
            test_fail "generate_hooks_json" "correct command path" "$command"
        fi
    else
        test_fail "generate_hooks_json" "SessionStart present" "SessionStart missing"
    fi
}

test_generate_hooks_json_grouped() {
    run_test "generate_hooks_json groups hooks by event"

    local registrations
    registrations=$(cat <<'EOF'
{"event":"SessionStart","matcher":"","path":"hook1.sh","timeout":5}
{"event":"SessionStart","matcher":"","path":"hook2.sh","timeout":5}
EOF
)

    local result
    result=$(generate_hooks_json "$registrations")

    local hook_count
    hook_count=$(echo "$result" | jq '.SessionStart[0].hooks | length')

    if [[ "$hook_count" -eq 2 ]]; then
        test_pass "grouped 2 hooks under same event"
    else
        test_fail "generate_hooks_json" "2 hooks" "$hook_count hooks"
    fi
}

test_generate_hooks_json_matchers() {
    run_test "generate_hooks_json separates different matchers"

    local registrations
    registrations=$(cat <<'EOF'
{"event":"PostToolUse","matcher":"Write","path":"write.sh","timeout":5}
{"event":"PostToolUse","matcher":"Edit","path":"edit.sh","timeout":5}
EOF
)

    local result
    result=$(generate_hooks_json "$registrations")

    local entry_count
    entry_count=$(echo "$result" | jq '.PostToolUse | length')

    if [[ "$entry_count" -eq 2 ]]; then
        test_pass "created separate entries for different matchers"
    else
        test_fail "generate_hooks_json" "2 entries" "$entry_count entries"
    fi
}

test_generate_hooks_json_no_matcher() {
    run_test "generate_hooks_json handles hooks without matcher"

    local registrations='{"event":"SessionStart","matcher":"","path":"hook.sh","timeout":5}'

    local result
    result=$(generate_hooks_json "$registrations")

    # Entry should not have matcher field when matcher is empty
    local has_matcher
    has_matcher=$(echo "$result" | jq '.SessionStart[0] | has("matcher")')

    if [[ "$has_matcher" == "false" ]]; then
        test_pass "entry without matcher field for empty matcher"
    else
        test_fail "generate_hooks_json" "no matcher field" "matcher field present"
    fi
}

test_generate_hooks_json_empty() {
    run_test "generate_hooks_json returns {} for empty input"

    local result
    result=$(generate_hooks_json "")

    if [[ "$result" == "{}" ]]; then
        test_pass "returned {} for empty input"
    else
        test_fail "generate_hooks_json" "{}" "$result"
    fi
}

# ============================================================================
# Tests for merge_with_preserved()
# ============================================================================

test_merge_with_preserved_empty() {
    run_test "merge_with_preserved returns generated when preserved is empty"

    local generated='{"SessionStart":[{"hooks":[{"type":"command","command":"test.sh","timeout":5}]}]}'
    local preserved='{}'

    local result
    result=$(merge_with_preserved "$generated" "$preserved")

    if [[ "$result" == "$generated" ]]; then
        test_pass "returned generated hooks unchanged"
    else
        test_fail "merge_with_preserved" "generated unchanged" "different result"
    fi
}

test_merge_with_preserved_append() {
    run_test "merge_with_preserved appends preserved to generated"

    local generated='{"SessionStart":[{"hooks":[{"type":"command","command":"roster.sh","timeout":5}]}]}'
    local preserved='{"SessionStart":[{"hooks":[{"type":"command","command":"user.sh","timeout":10}]}]}'

    local result
    result=$(merge_with_preserved "$generated" "$preserved")

    local entry_count
    entry_count=$(echo "$result" | jq '.SessionStart | length')

    if [[ "$entry_count" -eq 2 ]]; then
        # Verify order: generated first, preserved second
        local first_command
        first_command=$(echo "$result" | jq -r '.SessionStart[0].hooks[0].command')
        local second_command
        second_command=$(echo "$result" | jq -r '.SessionStart[1].hooks[0].command')

        if [[ "$first_command" == "roster.sh" ]] && [[ "$second_command" == "user.sh" ]]; then
            test_pass "appended preserved after generated"
        else
            test_fail "merge_with_preserved" "roster.sh then user.sh" "$first_command then $second_command"
        fi
    else
        test_fail "merge_with_preserved" "2 entries" "$entry_count entries"
    fi
}

test_merge_with_preserved_different_events() {
    run_test "merge_with_preserved preserves hooks for different events"

    local generated='{"SessionStart":[{"hooks":[{"type":"command","command":"start.sh","timeout":5}]}]}'
    local preserved='{"Stop":[{"hooks":[{"type":"command","command":"stop.sh","timeout":5}]}]}'

    local result
    result=$(merge_with_preserved "$generated" "$preserved")

    local event_count
    event_count=$(echo "$result" | jq 'keys | length')

    if [[ "$event_count" -eq 2 ]]; then
        test_pass "both events present in result"
    else
        test_fail "merge_with_preserved" "2 events" "$event_count events"
    fi
}

# ============================================================================
# Tests for swap_hook_registrations()
# ============================================================================

test_swap_hook_registrations_creates_settings() {
    run_test "swap_hook_registrations creates settings.local.json if missing"

    # Skip if yq not installed
    if ! command -v yq &>/dev/null; then
        echo "  SKIP: yq not installed"
        TESTS_RUN=$((TESTS_RUN - 1))
        return
    fi

    cd "$TEST_TMP"
    mkdir -p .claude
    mkdir -p user-hooks
    mkdir -p teams/test-team

    # Create minimal base_hooks.yaml
    cat > user-hooks/base_hooks.yaml <<'EOF'
schema_version: "1.0"
hooks:
  - event: SessionStart
    path: session-start.sh
    timeout: 5
EOF

    # Run swap_hook_registrations
    if swap_hook_registrations "test-team" 2>/dev/null; then
        if [[ -f ".claude/settings.local.json" ]]; then
            test_pass "created settings.local.json"
        else
            test_fail "swap_hook_registrations" "settings.local.json created" "file missing"
        fi
    else
        test_fail "swap_hook_registrations" "return 0" "non-zero return"
    fi
}

test_swap_hook_registrations_dry_run() {
    run_test "swap_hook_registrations shows preview in dry-run mode"

    # Skip if yq not installed
    if ! command -v yq &>/dev/null; then
        echo "  SKIP: yq not installed"
        TESTS_RUN=$((TESTS_RUN - 1))
        return
    fi

    cd "$TEST_TMP"
    mkdir -p .claude user-hooks teams/test-team
    echo '{}' > .claude/settings.local.json

    cat > user-hooks/base_hooks.yaml <<'EOF'
hooks:
  - event: SessionStart
    path: session-start.sh
    timeout: 5
EOF

    # Set dry-run mode
    DRY_RUN_MODE=1

    local output
    output=$(swap_hook_registrations "test-team" 2>&1)

    DRY_RUN_MODE=0

    # Should contain "dry-run" or preview indication
    if echo "$output" | grep -q "dry-run\|preview"; then
        test_pass "showed dry-run preview"
    else
        test_fail "swap_hook_registrations" "dry-run output" "no preview shown"
    fi
}

# ============================================================================
# Integration Test
# ============================================================================

test_hooks_registration_end_to_end() {
    run_test "Full hook registration flow with base and team hooks"

    # Skip if yq not installed
    if ! command -v yq &>/dev/null; then
        echo "  SKIP: yq not installed"
        TESTS_RUN=$((TESTS_RUN - 1))
        return
    fi

    cd "$TEST_TMP"
    mkdir -p .claude user-hooks teams/test-team

    # Create base hooks
    cat > user-hooks/base_hooks.yaml <<'EOF'
hooks:
  - event: SessionStart
    path: base-session-start.sh
    timeout: 5
EOF

    # Create team hooks
    cat > teams/test-team/hooks.yaml <<'EOF'
hooks:
  - event: PostToolUse
    matcher: Write|Edit
    path: team-post-write.sh
    timeout: 10
EOF

    # Create settings with user hook
    cat > .claude/settings.local.json <<'EOF'
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/user-stop.sh",
            "timeout": 15
          }
        ]
      }
    ]
  }
}
EOF

    if swap_hook_registrations "test-team" 2>/dev/null; then
        local settings
        settings=$(cat .claude/settings.local.json)

        # Should have all three: SessionStart (base), PostToolUse (team), Stop (user)
        local has_session_start
        has_session_start=$(echo "$settings" | jq 'has("hooks") and .hooks | has("SessionStart")')
        local has_post_tool
        has_post_tool=$(echo "$settings" | jq '.hooks | has("PostToolUse")')
        local has_stop
        has_stop=$(echo "$settings" | jq '.hooks | has("Stop")')

        if [[ "$has_session_start" == "true" ]] && \
           [[ "$has_post_tool" == "true" ]] && \
           [[ "$has_stop" == "true" ]]; then
            # Verify user hook preserved
            local user_command
            user_command=$(echo "$settings" | jq -r '.hooks.Stop[0].hooks[0].command')
            if [[ "$user_command" == "/usr/local/bin/user-stop.sh" ]]; then
                test_pass "registered base, team, and preserved user hooks"
            else
                test_fail "end-to-end" "user hook preserved" "user hook: $user_command"
            fi
        else
            test_fail "end-to-end" "all 3 events" "start:$has_session_start post:$has_post_tool stop:$has_stop"
        fi
    else
        test_fail "swap_hook_registrations" "return 0" "non-zero return"
    fi
}

# ============================================================================
# Main test runner
# ============================================================================

main() {
    echo "========================================"
    echo "Team Hooks Registration Unit Tests"
    echo "========================================"
    echo ""

    setup

    # Validation tests
    test_require_yq_installed

    # YAML parsing tests
    test_parse_hooks_yaml_valid
    test_parse_hooks_yaml_missing
    test_parse_hooks_yaml_invalid_event
    test_parse_hooks_yaml_no_matcher
    test_parse_hooks_yaml_invalid_regex
    test_parse_hooks_yaml_timeout_clamp

    # JSON extraction tests
    test_extract_non_roster_roster_only
    test_extract_non_roster_mixed
    test_extract_non_roster_user_only
    test_extract_non_roster_missing

    # Merge tests
    test_merge_hook_registrations
    test_merge_hook_registrations_empty_base
    test_merge_hook_registrations_empty_team

    # JSON generation tests
    test_generate_hooks_json_single
    test_generate_hooks_json_grouped
    test_generate_hooks_json_matchers
    test_generate_hooks_json_no_matcher
    test_generate_hooks_json_empty

    # Merge with preserved tests
    test_merge_with_preserved_empty
    test_merge_with_preserved_append
    test_merge_with_preserved_different_events

    # Orchestration tests
    test_swap_hook_registrations_creates_settings
    test_swap_hook_registrations_dry_run

    # Integration test
    test_hooks_registration_end_to_end

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

#!/bin/bash
# Integration test for thread contract validation in sails check
# Tests that contract violations are detected and degrade WHITE to GRAY

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARI="${PROJECT_ROOT}/ariadne/ari"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

pass_count=0
fail_count=0

# Test helper functions
assert_exit_code() {
    local expected=$1
    local actual=$2
    local msg=$3

    if [ "$expected" -eq "$actual" ]; then
        echo -e "${GREEN}PASS${NC}: $msg (exit code $actual)"
        ((pass_count++))
    else
        echo -e "${RED}FAIL${NC}: $msg (expected exit code $expected, got $actual)"
        ((fail_count++))
    fi
}

assert_contains() {
    local haystack=$1
    local needle=$2
    local msg=$3

    if echo "$haystack" | grep -q "$needle"; then
        echo -e "${GREEN}PASS${NC}: $msg"
        ((pass_count++))
    else
        echo -e "${RED}FAIL${NC}: $msg (output does not contain '$needle')"
        echo "Output was:"
        echo "$haystack"
        ((fail_count++))
    fi
}

# Build ariadne if needed
if [ ! -f "$ARI" ]; then
    echo "Building ariadne..."
    cd "$PROJECT_ROOT/ariadne"
    just build
fi

echo "=== Thread Contract Validation Integration Tests ==="
echo

# Test 1: WHITE sails with valid events should pass
echo "Test 1: WHITE sails with valid events"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

cat > "$TEST_DIR/WHITE_SAILS.yaml" << 'EOF'
version: "1.0"
session_id: "test-session-001"
color: WHITE
computed_base: WHITE
type: standard
complexity: MODULE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
EOF

cat > "$TEST_DIR/events.jsonl" << 'EOF'
{"ts":"2026-01-06T12:00:00.000Z","type":"task_start","summary":"Task started: task-001 by agent-a in design phase","meta":{"task_id":"task-001","agent":"agent-a","phase":"design","session_id":"test-session-001"}}
{"ts":"2026-01-06T12:05:00.000Z","type":"task_end","summary":"Task ended: task-001 by agent-a - success (300000ms)","meta":{"task_id":"task-001","agent":"agent-a","outcome":"success","session_id":"test-session-001","duration_ms":300000,"artifacts":[]}}
EOF

output=$("$ARI" sails check "$TEST_DIR" 2>&1 || true)
exit_code=$?
assert_exit_code 0 $exit_code "WHITE sails with valid events should pass"
assert_contains "$output" "PASS: Quality gate passed" "Output should show PASS"
echo

# Test 2: WHITE sails with orphaned task_end should downgrade to GRAY
echo "Test 2: WHITE sails with orphaned task_end violation"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

cat > "$TEST_DIR/WHITE_SAILS.yaml" << 'EOF'
version: "1.0"
session_id: "test-session-002"
color: WHITE
computed_base: WHITE
type: standard
complexity: MODULE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
EOF

cat > "$TEST_DIR/events.jsonl" << 'EOF'
{"ts":"2026-01-06T12:05:00.000Z","type":"task_end","summary":"Task ended: task-001 by agent-a - success (300000ms)","meta":{"task_id":"task-001","agent":"agent-a","outcome":"success","session_id":"test-session-002","duration_ms":300000,"artifacts":[]}}
EOF

output=$("$ARI" sails check "$TEST_DIR" 2>&1 || true)
exit_code=$?
assert_exit_code 134 $exit_code "WHITE sails with violations should fail gate (non-zero exit)"
assert_contains "$output" "FAIL: Quality gate failed" "Output should show FAIL"
assert_contains "$output" "Color:.*GRAY" "Color should be downgraded to GRAY"
assert_contains "$output" "Thread Contract Violations" "Output should show contract violations section"
assert_contains "$output" "task_orphaned_end" "Output should show task_orphaned_end violation"
echo

# Test 3: WHITE sails with unprepared handoff should downgrade to GRAY
echo "Test 3: WHITE sails with unprepared handoff violation"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

cat > "$TEST_DIR/WHITE_SAILS.yaml" << 'EOF'
version: "1.0"
session_id: "test-session-003"
color: WHITE
computed_base: WHITE
type: standard
complexity: MODULE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
EOF

cat > "$TEST_DIR/events.jsonl" << 'EOF'
{"ts":"2026-01-06T12:00:00.000Z","type":"handoff_executed","summary":"Handoff executed: agent-a -> agent-b (0 artifacts)","meta":{"from_agent":"agent-a","to_agent":"agent-b","session_id":"test-session-003","artifacts":[]}}
EOF

output=$("$ARI" sails check "$TEST_DIR" 2>&1 || true)
exit_code=$?
assert_exit_code 134 $exit_code "WHITE sails with handoff violation should fail gate"
assert_contains "$output" "FAIL: Quality gate failed" "Output should show FAIL"
assert_contains "$output" "Color:.*GRAY" "Color should be downgraded to GRAY"
assert_contains "$output" "handoff_unprepared" "Output should show handoff_unprepared violation"
echo

# Test 4: GRAY sails with violations should remain GRAY
echo "Test 4: GRAY sails with violations remain GRAY"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

cat > "$TEST_DIR/WHITE_SAILS.yaml" << 'EOF'
version: "1.0"
session_id: "test-session-004"
color: GRAY
computed_base: GRAY
type: standard
complexity: MODULE
open_questions:
  - "What is the performance impact?"
EOF

cat > "$TEST_DIR/events.jsonl" << 'EOF'
{"ts":"2026-01-06T12:05:00.000Z","type":"task_end","summary":"Task ended: task-001 by agent-a - success (300000ms)","meta":{"task_id":"task-001","agent":"agent-a","outcome":"success","session_id":"test-session-004","duration_ms":300000,"artifacts":[]}}
EOF

output=$("$ARI" sails check "$TEST_DIR" 2>&1 || true)
exit_code=$?
assert_exit_code 134 $exit_code "GRAY sails should fail gate"
assert_contains "$output" "Color:.*GRAY" "Color should remain GRAY"
assert_contains "$output" "Thread Contract Violations" "Output should show contract violations"
echo

# Test 5: No events.jsonl should not affect sails
echo "Test 5: WHITE sails without events.jsonl"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

cat > "$TEST_DIR/WHITE_SAILS.yaml" << 'EOF'
version: "1.0"
session_id: "test-session-005"
color: WHITE
computed_base: WHITE
type: standard
complexity: MODULE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
EOF

# No events.jsonl file

output=$("$ARI" sails check "$TEST_DIR" 2>&1 || true)
exit_code=$?
assert_exit_code 0 $exit_code "WHITE sails without events.jsonl should pass"
assert_contains "$output" "PASS: Quality gate passed" "Output should show PASS"
echo

# Summary
echo "==================================="
echo "Test Summary:"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo "==================================="

if [ $fail_count -gt 0 ]; then
    exit 1
fi

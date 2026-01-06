#!/bin/bash
# test-start-orchestrator-skip.sh - Integration test for start-preflight.sh orchestrator skip logic
# Tests hook_002 and hook_003 from TDD-auto-orchestration.md
set -euo pipefail

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Test: start-preflight.sh orchestrator skip logic"
echo "================================================"

# Setup: Create minimal test satellite
mkdir -p "$TEST_DIR/.claude/hooks/lib"
mkdir -p "$TEST_DIR/.claude/agents"
mkdir -p "$TEST_DIR/.claude/sessions"

# Copy required library files
cp -r "$PROJECT_ROOT/.claude/hooks/lib/"* "$TEST_DIR/.claude/hooks/lib/"
cp "$PROJECT_ROOT/.claude/hooks/session-guards/start-preflight.sh" "$TEST_DIR/.claude/hooks/"

cd "$TEST_DIR"

# Test 1: hook_002 - With orchestrator, preflight skips when session exists
echo ""
echo "Test 1: hook_002 - Orchestrator present, active session -> preflight skips"
echo "--------------------------------------------------------------------------"

# Create orchestrator.md
echo "# Orchestrator" > .claude/agents/orchestrator.md

# Create ACTIVE_RITE
echo "ecosystem-pack" > .claude/ACTIVE_RITE

# Create session via session-manager
SESSION_RESULT=$(.claude/hooks/lib/session-manager.sh create "Test Initiative" "MODULE" "ecosystem-pack" 2>&1)
SESSION_ID=$(echo "$SESSION_RESULT" | grep -o '"session_id": *"[^"]*"' | cut -d'"' -f4)
if [[ -z "$SESSION_ID" ]]; then
    echo "FAIL: Could not create session for test 1"
    echo "Result: $SESSION_RESULT"
    exit 1
fi

# Run start-preflight.sh with /start command and existing session
export CLAUDE_PROJECT_DIR="$TEST_DIR"
export CLAUDE_USER_PROMPT="/start Test Initiative"

OUTPUT=$(.claude/hooks/start-preflight.sh 2>&1 || true)

if [[ -z "$OUTPUT" ]]; then
    echo "PASS: Preflight produced no output (orchestrator handled it)"
else
    echo "FAIL: Preflight should have exited silently but produced output:"
    echo "$OUTPUT"
    exit 1
fi

# Test 2: hook_003 - Without orchestrator, preflight handles creation
echo ""
echo "Test 2: hook_003 - No orchestrator -> preflight handles creation"
echo "----------------------------------------------------------------"

# Clean up for fresh test
rm -rf .claude/sessions/*
rm -f .claude/agents/orchestrator.md

# Run start-preflight.sh with /start command and no session
export CLAUDE_PROJECT_DIR="$TEST_DIR"
export CLAUDE_USER_PROMPT="/start New Initiative MODULE"

OUTPUT=$(.claude/hooks/start-preflight.sh 2>&1 || true)

if [[ "$OUTPUT" == *"Preflight Check"* ]] && [[ "$OUTPUT" == *"Session Created"* ]]; then
    echo "PASS: Preflight created session and output status"
else
    echo "FAIL: Preflight should have created session and output status"
    echo "Output: $OUTPUT"
    exit 1
fi

# Verify session was created
SESSION_COUNT=$(find .claude/sessions -maxdepth 1 -type d -name "session-*" 2>/dev/null | wc -l)
if [[ $SESSION_COUNT -gt 0 ]]; then
    echo "PASS: Session directory created"
else
    echo "FAIL: Session directory not created"
    ls -la .claude/sessions/
    exit 1
fi

# Test 3: With orchestrator, no session -> preflight skips (router should have created)
echo ""
echo "Test 3: Orchestrator present, no session -> preflight skips"
echo "----------------------------------------------------------"

# Clean up
rm -rf .claude/sessions/*

# Create orchestrator.md again
echo "# Orchestrator" > .claude/agents/orchestrator.md

# Run start-preflight.sh
export CLAUDE_USER_PROMPT="/start Another Initiative"

OUTPUT=$(.claude/hooks/start-preflight.sh 2>&1 || true)

if [[ -z "$OUTPUT" ]]; then
    echo "PASS: Preflight skipped (router should have handled creation)"
else
    echo "FAIL: Preflight should skip when orchestrator present but no session"
    echo "Output: $OUTPUT"
    exit 1
fi

# Test 4: Parked session with orchestrator -> shows options
echo ""
echo "Test 4: Parked session with orchestrator -> shows options"
echo "--------------------------------------------------------"

# Create parked session
SESSION_RESULT=$(.claude/hooks/lib/session-manager.sh create "Parked Initiative" "MODULE" "ecosystem-pack" 2>&1)
SESSION_ID=$(echo "$SESSION_RESULT" | grep -o '"session_id": *"[^"]*"' | cut -d'"' -f4)
if [[ -z "$SESSION_ID" ]]; then
    echo "FAIL: Could not create session for test 4"
    echo "Result: $SESSION_RESULT"
    exit 1
fi

# Mark as parked
cat >> ".claude/sessions/$SESSION_ID/SESSION_CONTEXT.md" <<EOF
parked_at: 2026-01-04T00:00:00Z
EOF

export CLAUDE_USER_PROMPT="/start New Work"

OUTPUT=$(.claude/hooks/start-preflight.sh 2>&1 || true)

if [[ "$OUTPUT" == *"Session exists (parked)"* ]] && [[ "$OUTPUT" == *"/continue"* ]]; then
    echo "PASS: Preflight shows parked session options"
else
    echo "FAIL: Preflight should show parked session options"
    echo "Output: $OUTPUT"
    exit 1
fi

echo ""
echo "================================================"
echo "All tests passed!"

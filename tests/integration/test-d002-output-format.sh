#!/bin/bash
# test-d002-output-format.sh - Verify D002 fix: Task invocation format (not YAML)
# Tests that orchestrator-router.sh outputs Task(orchestrator...) not CONSULTATION_REQUEST YAML

set -euo pipefail

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

echo "=== D002 Output Format Verification ==="
echo ""

# Setup: Create minimal project structure
mkdir -p "$TEST_DIR/.claude/"{agents,hooks/lib,hooks/validation,sessions}
mkdir -p "$TEST_DIR/user-hooks/validation"

# Create orchestrator.md (indicates orchestrator is present)
echo "# Orchestrator" > "$TEST_DIR/.claude/agents/orchestrator.md"

# Create ACTIVE_RITE
echo "ecosystem" > "$TEST_DIR/.claude/ACTIVE_RITE"

# Copy library dependencies
cp /Users/tomtenuta/Code/knossos/.claude/hooks/lib/{logging.sh,session-utils.sh,session-manager.sh} \
   "$TEST_DIR/.claude/hooks/lib/"

# Copy the fixed orchestrator-router.sh
cp /Users/tomtenuta/Code/knossos/user-hooks/validation/orchestrator-router.sh \
   "$TEST_DIR/.claude/hooks/validation/orchestrator-router.sh"

chmod +x "$TEST_DIR/.claude/hooks/validation/orchestrator-router.sh"
chmod +x "$TEST_DIR/.claude/hooks/lib/session-manager.sh"

# Execute the hook
cd "$TEST_DIR"
export CLAUDE_PROJECT_DIR="$TEST_DIR"
export CLAUDE_USER_PROMPT="/start Test Initiative"

OUTPUT=$(.claude/hooks/validation/orchestrator-router.sh 2>&1)

echo "=== Hook Output ==="
echo "$OUTPUT"
echo ""

# Test 1: Output should contain Task invocation (NEW FORMAT)
if echo "$OUTPUT" | grep -q "Task(orchestrator"; then
    echo "✓ PASS: Output contains Task(orchestrator invocation (NEW FORMAT)"
else
    echo "✗ FAIL: Output missing Task(orchestrator invocation"
    exit 1
fi

# Test 2: Output should NOT contain YAML CONSULTATION_REQUEST (OLD FORMAT)
if echo "$OUTPUT" | grep -q "^type: initial$"; then
    echo "✗ FAIL: Output still contains YAML format (OLD FORMAT - D002 NOT FIXED)"
    exit 1
else
    echo "✓ PASS: Output does not contain YAML format (D002 FIXED)"
fi

# Test 3: Output should contain session context fields
CHECKS=(
    "Session ID:"
    "Session Path:"
    "Initiative:"
    "Complexity:"
    "Team:"
    "Request Type:"
)

for check in "${CHECKS[@]}"; do
    if echo "$OUTPUT" | grep -q "$check"; then
        echo "✓ PASS: Output contains '$check'"
    else
        echo "✗ FAIL: Output missing '$check'"
        exit 1
    fi
done

# Test 4: Verify syntactic structure
if echo "$OUTPUT" | grep -E 'Task\(orchestrator, ".*' > /dev/null; then
    echo "✓ PASS: Task invocation is syntactically valid"
else
    echo "✗ FAIL: Task invocation syntax invalid"
    exit 1
fi

echo ""
echo "=== D002 RESOLVED ==="
echo "orchestrator-router.sh now outputs Task invocation format (not YAML)"
echo "All tests passed."

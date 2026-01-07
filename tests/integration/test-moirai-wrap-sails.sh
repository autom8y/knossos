#!/bin/bash
# test-moirai-wrap-sails.sh - Integration test for Moirai wrap_session sails validation
# Tests T3-005: Moirai wrap_session validates sails (gate task for Track 4)
set -euo pipefail

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Test: Moirai wrap_session sails validation (T3-005)"
echo "===================================================="

# Check if ari binary exists
ARI_BIN="$PROJECT_ROOT/ariadne/ari"
if [[ ! -f "$ARI_BIN" ]]; then
    echo "SKIP: ari binary not found at $ARI_BIN"
    echo "Run: cd ariadne && just build"
    exit 0
fi

# Setup: Create minimal test satellite
mkdir -p "$TEST_DIR/.claude/sessions"
cd "$TEST_DIR"

# Test 1: Wrap with WHITE sails - should succeed
echo ""
echo "Test 1: Wrap with WHITE sails - should succeed"
echo "----------------------------------------------"

# Create session
SESSION_ID="session-white-$(date +%s)"
SESSION_DIR=".claude/sessions/$SESSION_ID"
mkdir -p "$SESSION_DIR"

# Create SESSION_CONTEXT.md with no blockers
cat > "$SESSION_DIR/SESSION_CONTEXT.md" <<EOF
---
schema_version: "2.1"
session_id: $SESSION_ID
status: ACTIVE
created_at: "2026-01-06T10:00:00Z"
initiative: "Test wrap with WHITE sails"
complexity: PATCH
active_rite: ecosystem
rite: ecosystem
current_phase: implementation
---

# Session: Test wrap with WHITE sails

## Initiative
Test wrap with WHITE sails.

## Blockers
- None

## Open Questions
- None
EOF

# Create proof evidence (all passing) - using correct filenames
cat > "$SESSION_DIR/test-output.log" <<'EOF'
Running tests...
ok  github.com/test/pkg 0.123s
5 tests passed
EOF

cat > "$SESSION_DIR/build-output.log" <<'EOF'
Building...
Build succeeded
exit code: 0
EOF

cat > "$SESSION_DIR/lint-output.log" <<'EOF'
Linting...
No issues found
exit code: 0
EOF

# Set current session
echo "$SESSION_ID" > .claude/CURRENT_SESSION

# Run wrap (ari needs CLAUDE_PROJECT_DIR environment variable)
export CLAUDE_PROJECT_DIR="$TEST_DIR"
WRAP_OUTPUT=$("$ARI_BIN" session wrap 2>&1 || true)
WRAP_EXIT_CODE=$?

echo "DEBUG: Wrap output:"
echo "$WRAP_OUTPUT"
echo "DEBUG: Exit code: $WRAP_EXIT_CODE"

if [[ $WRAP_EXIT_CODE -eq 0 ]]; then
    echo "PASS: Wrap succeeded"
else
    echo "FAIL: Wrap should have succeeded but failed"
    echo "Exit code: $WRAP_EXIT_CODE"
    echo "Output: $WRAP_OUTPUT"
    exit 1
fi

# Debug: Show archive location
echo "DEBUG: Archive check..."
ls -la .claude/archive/ 2>/dev/null || echo "No archive directory created"
ls -la "$SESSION_DIR" 2>/dev/null || echo "Session dir moved"

# Verify WHITE_SAILS.yaml was generated (check both session dir and archive)
SAILS_FILE=""
if [[ -f "$SESSION_DIR/WHITE_SAILS.yaml" ]]; then
    SAILS_FILE="$SESSION_DIR/WHITE_SAILS.yaml"
elif [[ -f ".claude/archive/$SESSION_ID/WHITE_SAILS.yaml" ]]; then
    SAILS_FILE=".claude/archive/$SESSION_ID/WHITE_SAILS.yaml"
fi

if [[ -n "$SAILS_FILE" ]]; then
    SAILS_COLOR=$(grep "^color:" "$SAILS_FILE" | awk '{print $2}')
    echo "PASS: WHITE_SAILS.yaml generated with color: $SAILS_COLOR"
else
    echo "FAIL: WHITE_SAILS.yaml not found in session or archive"
    ls -la "$SESSION_DIR/" || true
    ls -la ".claude/archive/$SESSION_ID/" || true
    exit 1
fi

# Verify session was archived (check both locations)
CONTEXT_FILE=""
if [[ -f "$SESSION_DIR/SESSION_CONTEXT.md" ]]; then
    CONTEXT_FILE="$SESSION_DIR/SESSION_CONTEXT.md"
elif [[ -f ".claude/archive/$SESSION_ID/SESSION_CONTEXT.md" ]]; then
    CONTEXT_FILE=".claude/archive/$SESSION_ID/SESSION_CONTEXT.md"
fi

if [[ -n "$CONTEXT_FILE" ]]; then
    CONTEXT_STATE=$(grep "^status:" "$CONTEXT_FILE" | awk '{print $2}' | tr -d '"')
    if [[ "$CONTEXT_STATE" == "ARCHIVED" ]]; then
        echo "PASS: Session status is ARCHIVED"
    else
        echo "FAIL: Session status should be ARCHIVED but is: $CONTEXT_STATE"
        cat "$CONTEXT_FILE"
        exit 1
    fi
else
    echo "FAIL: SESSION_CONTEXT.md not found"
    exit 1
fi

# Test 2: Wrap with BLACK sails - should block
echo ""
echo "Test 2: Wrap with BLACK sails - should block"
echo "--------------------------------------------"

# Create session with blockers
SESSION_ID_BLACK="session-black-$(date +%s)"
SESSION_DIR_BLACK=".claude/sessions/$SESSION_ID_BLACK"
mkdir -p "$SESSION_DIR_BLACK"

cat > "$SESSION_DIR_BLACK/SESSION_CONTEXT.md" <<EOF
---
schema_version: "2.1"
session_id: $SESSION_ID_BLACK
status: ACTIVE
created_at: "2026-01-06T10:00:00Z"
initiative: "Test wrap with BLACK sails"
complexity: PATCH
active_rite: ecosystem
rite: ecosystem
current_phase: implementation
---

# Session: Test wrap with BLACK sails

## Initiative
Test wrap with BLACK sails.

## Blockers
- Tests failing in integration suite
- Build broken on macOS

## Open Questions
- None
EOF

# Create proof evidence (with failures) - using correct filenames
cat > "$SESSION_DIR_BLACK/test-output.log" <<'EOF'
Running tests...
FAIL  github.com/test/pkg 0.123s
3 tests failed
exit code: 1
EOF

cat > "$SESSION_DIR_BLACK/build-output.log" <<'EOF'
Building...
Build failed
exit code: 1
EOF

cat > "$SESSION_DIR_BLACK/lint-output.log" <<'EOF'
Linting...
Found 5 issues
exit code: 0
EOF

# Set current session
echo "$SESSION_ID_BLACK" > .claude/CURRENT_SESSION

# Run wrap (should fail)
export CLAUDE_PROJECT_DIR="$TEST_DIR"
WRAP_OUTPUT_BLACK=$("$ARI_BIN" session wrap 2>&1 || true)
WRAP_EXIT_CODE_BLACK=$?

if [[ $WRAP_EXIT_CODE_BLACK -ne 0 ]]; then
    echo "PASS: Wrap blocked on BLACK sails (exit code: $WRAP_EXIT_CODE_BLACK)"
else
    echo "FAIL: Wrap should have blocked on BLACK sails but succeeded"
    echo "Output: $WRAP_OUTPUT_BLACK"
    exit 1
fi

# Verify error message mentions BLACK sails
if [[ "$WRAP_OUTPUT_BLACK" == *"BLACK sails"* ]] || [[ "$WRAP_OUTPUT_BLACK" == *"blockers present"* ]]; then
    echo "PASS: Error message mentions BLACK sails or blockers"
else
    echo "FAIL: Error message should mention BLACK sails or blockers"
    echo "Output: $WRAP_OUTPUT_BLACK"
    exit 1
fi

# Verify session was NOT archived
CONTEXT_STATE_BLACK=$(grep "^status:" "$SESSION_DIR_BLACK/SESSION_CONTEXT.md" | awk '{print $2}' | tr -d '"')
if [[ "$CONTEXT_STATE_BLACK" == "ACTIVE" ]]; then
    echo "PASS: Session status remains ACTIVE (wrap blocked)"
else
    echo "FAIL: Session status should remain ACTIVE but is: $CONTEXT_STATE_BLACK"
    cat "$SESSION_DIR_BLACK/SESSION_CONTEXT.md"
    exit 1
fi

# Test 3: Wrap with BLACK sails + --force - should succeed with warning
echo ""
echo "Test 3: Wrap with BLACK sails + --force - should succeed with warning"
echo "---------------------------------------------------------------------"

# Run wrap with --force
export CLAUDE_PROJECT_DIR="$TEST_DIR"
WRAP_OUTPUT_FORCE=$("$ARI_BIN" session wrap --force 2>&1 || true)
WRAP_EXIT_CODE_FORCE=$?

if [[ $WRAP_EXIT_CODE_FORCE -eq 0 ]]; then
    echo "PASS: Wrap succeeded with --force despite BLACK sails"
else
    echo "FAIL: Wrap with --force should have succeeded"
    echo "Exit code: $WRAP_EXIT_CODE_FORCE"
    echo "Output: $WRAP_OUTPUT_FORCE"
    exit 1
fi

# Verify WHITE_SAILS.yaml shows BLACK (check both locations)
SAILS_FILE_BLACK=""
if [[ -f "$SESSION_DIR_BLACK/WHITE_SAILS.yaml" ]]; then
    SAILS_FILE_BLACK="$SESSION_DIR_BLACK/WHITE_SAILS.yaml"
elif [[ -f ".claude/archive/$SESSION_ID_BLACK/WHITE_SAILS.yaml" ]]; then
    SAILS_FILE_BLACK=".claude/archive/$SESSION_ID_BLACK/WHITE_SAILS.yaml"
fi

if [[ -n "$SAILS_FILE_BLACK" ]]; then
    SAILS_COLOR_BLACK=$(grep "^color:" "$SAILS_FILE_BLACK" | awk '{print $2}')
    if [[ "$SAILS_COLOR_BLACK" == "BLACK" ]]; then
        echo "PASS: WHITE_SAILS.yaml correctly shows BLACK color"
    else
        echo "FAIL: WHITE_SAILS.yaml should show BLACK but shows: $SAILS_COLOR_BLACK"
        cat "$SAILS_FILE_BLACK"
        exit 1
    fi
else
    echo "FAIL: WHITE_SAILS.yaml not found after --force wrap"
    ls -la "$SESSION_DIR_BLACK/" || true
    ls -la ".claude/archive/$SESSION_ID_BLACK/" || true
    exit 1
fi

# Verify session was archived despite BLACK sails (check both locations)
CONTEXT_FILE_FORCE=""
if [[ -f "$SESSION_DIR_BLACK/SESSION_CONTEXT.md" ]]; then
    CONTEXT_FILE_FORCE="$SESSION_DIR_BLACK/SESSION_CONTEXT.md"
elif [[ -f ".claude/archive/$SESSION_ID_BLACK/SESSION_CONTEXT.md" ]]; then
    CONTEXT_FILE_FORCE=".claude/archive/$SESSION_ID_BLACK/SESSION_CONTEXT.md"
fi

if [[ -n "$CONTEXT_FILE_FORCE" ]]; then
    CONTEXT_STATE_FORCE=$(grep "^status:" "$CONTEXT_FILE_FORCE" | awk '{print $2}' | tr -d '"')
    if [[ "$CONTEXT_STATE_FORCE" == "ARCHIVED" ]]; then
        echo "PASS: Session archived with --force override"
    else
        echo "FAIL: Session should be ARCHIVED with --force but is: $CONTEXT_STATE_FORCE"
        cat "$CONTEXT_FILE_FORCE"
        exit 1
    fi
else
    echo "FAIL: SESSION_CONTEXT.md not found after --force wrap"
    exit 1
fi

echo ""
echo "===================================================="
echo "All tests passed: Moirai wrap_session validates sails"
echo "T3-005 COMPLETE: Gate task for Track 4 ready"
echo "===================================================="

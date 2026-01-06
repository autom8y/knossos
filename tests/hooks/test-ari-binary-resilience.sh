#!/bin/bash
# Integration test: Validate ari hook binary resilience
# Tests graceful degradation when ari binary is missing or unavailable
set -euo pipefail

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ARI_HOOKS_DIR="$PROJECT_DIR/user-hooks/ari"

echo "=== Ari Hook Binary Resilience Test ==="
echo "Testing graceful degradation when ari binary is unavailable"
echo ""

# Helper function to run a hook with controlled environment
run_hook() {
    local hook="$1"
    shift
    (
        # Isolate environment
        export ARIADNE_BIN="/nonexistent/path/to/ari"
        export PATH="/usr/bin:/bin"  # Remove ari from PATH
        export CLAUDE_PROJECT_DIR="$TEST_DIR"
        export CLAUDE_HOOK_TOOL_NAME="${CLAUDE_HOOK_TOOL_NAME:-Bash}"
        export CLAUDE_HOOK_TOOL_INPUT="${CLAUDE_HOOK_TOOL_INPUT:-{}}"
        export CLAUDE_HOOK_USER_PROMPT="${CLAUDE_HOOK_USER_PROMPT:-/test}"
        export CLAUDE_SESSION_DIR="$TEST_DIR/.claude/sessions"
        export USE_ARI_HOOKS=1
        "$@"
        "$ARI_HOOKS_DIR/$hook"
    )
}

# Setup test environment
mkdir -p "$TEST_DIR/.claude/sessions/test-session"
echo "status: ACTIVE" > "$TEST_DIR/.claude/sessions/test-session/SESSION_CONTEXT.md"

PASS_COUNT=0
FAIL_COUNT=0

pass() {
    echo "✓ PASS: $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo "✗ FAIL: $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

# =============================================================================
# Test 1: context.sh exits 0 with missing binary
# =============================================================================
echo "Test 1: context.sh exits 0 when ari binary missing"
if run_hook "context.sh" 2>/dev/null; then
    pass "context.sh exits cleanly (exit 0) when binary missing"
else
    fail "context.sh should exit 0, but exited with non-zero"
fi
echo ""

# =============================================================================
# Test 2: validate.sh exits 0 with missing binary
# =============================================================================
echo "Test 2: validate.sh exits 0 when ari binary missing"
if CLAUDE_HOOK_TOOL_NAME=Bash run_hook "validate.sh" 2>/dev/null; then
    pass "validate.sh exits cleanly when binary missing"
else
    fail "validate.sh should exit 0, but exited with non-zero"
fi
echo ""

# =============================================================================
# Test 3: writeguard.sh exits 0 with missing binary
# =============================================================================
echo "Test 3: writeguard.sh exits 0 when ari binary missing"
if CLAUDE_HOOK_TOOL_NAME=Write CLAUDE_HOOK_TOOL_INPUT='{"file_path":"SESSION_CONTEXT.md"}' run_hook "writeguard.sh" 2>/dev/null; then
    pass "writeguard.sh exits cleanly when binary missing"
else
    fail "writeguard.sh should exit 0, but exited with non-zero"
fi
echo ""

# =============================================================================
# Test 4: route.sh exits 0 with missing binary
# =============================================================================
echo "Test 4: route.sh exits 0 when ari binary missing"
if CLAUDE_HOOK_USER_PROMPT="/test" run_hook "route.sh" 2>/dev/null; then
    pass "route.sh exits cleanly when binary missing"
else
    fail "route.sh should exit 0, but exited with non-zero"
fi
echo ""

# =============================================================================
# Test 5: clew.sh exits 0 with missing binary
# =============================================================================
echo "Test 5: clew.sh exits 0 when ari binary missing"
if CLAUDE_HOOK_TOOL_NAME=Edit run_hook "clew.sh" 2>/dev/null; then
    pass "clew.sh exits cleanly when binary missing"
else
    fail "clew.sh should exit 0, but exited with non-zero"
fi
echo ""

# =============================================================================
# Test 6: autopark.sh exits 0 with missing binary
# =============================================================================
echo "Test 6: autopark.sh exits 0 when ari binary missing"
if run_hook "autopark.sh" 2>/dev/null; then
    pass "autopark.sh exits cleanly when binary missing"
else
    fail "autopark.sh should exit 0, but exited with non-zero"
fi
echo ""

# =============================================================================
# Test 7: USE_ARI_HOOKS=0 skips ari dispatch
# =============================================================================
echo "Test 7: USE_ARI_HOOKS=0 skips ari dispatch entirely"
(
    export USE_ARI_HOOKS=0
    export ARIADNE_BIN="$PROJECT_DIR/ariadne/ari"  # Real path (if exists)
    if "$ARI_HOOKS_DIR/context.sh" 2>/dev/null; then
        pass "USE_ARI_HOOKS=0 causes early exit"
    else
        fail "USE_ARI_HOOKS=0 should cause exit 0"
    fi
)
echo ""

# =============================================================================
# Test 8: ARIADNE_BIN override is respected
# =============================================================================
echo "Test 8: ARIADNE_BIN environment variable is respected"
# Create a fake ari binary that we control
FAKE_ARI="$TEST_DIR/fake-ari"
cat > "$FAKE_ARI" << 'EOF'
#!/bin/bash
echo "FAKE_ARI_CALLED"
exit 0
EOF
chmod +x "$FAKE_ARI"

OUTPUT=$(
    export ARIADNE_BIN="$FAKE_ARI"
    export USE_ARI_HOOKS=1
    "$ARI_HOOKS_DIR/context.sh" 2>&1
) || true

if echo "$OUTPUT" | grep -q "FAKE_ARI_CALLED"; then
    pass "ARIADNE_BIN override is respected"
else
    fail "ARIADNE_BIN was not used (output: $OUTPUT)"
fi
echo ""

# =============================================================================
# Test 9: PATH lookup works when ARIADNE_BIN not set
# =============================================================================
echo "Test 9: PATH lookup works when ARIADNE_BIN not set"
# Create a fake ari in a temp PATH directory
FAKE_PATH="$TEST_DIR/bin"
mkdir -p "$FAKE_PATH"
cat > "$FAKE_PATH/ari" << 'EOF'
#!/bin/bash
echo "PATH_ARI_CALLED"
exit 0
EOF
chmod +x "$FAKE_PATH/ari"

OUTPUT=$(
    unset ARIADNE_BIN
    export PATH="$FAKE_PATH:$PATH"
    export USE_ARI_HOOKS=1
    export CLAUDE_PROJECT_DIR="$TEST_DIR"
    "$ARI_HOOKS_DIR/context.sh" 2>&1
) || true

if echo "$OUTPUT" | grep -q "PATH_ARI_CALLED"; then
    pass "PATH lookup finds ari binary"
else
    fail "PATH lookup did not work (output: $OUTPUT)"
fi
echo ""

# =============================================================================
# Test 10: Project-relative fallback works
# =============================================================================
echo "Test 10: Project-relative fallback (CLAUDE_PROJECT_DIR/ariadne/ari)"
# Create fake ari in project-relative location
PROJECT_ARI_DIR="$TEST_DIR/ariadne"
mkdir -p "$PROJECT_ARI_DIR"
cat > "$PROJECT_ARI_DIR/ari" << 'EOF'
#!/bin/bash
echo "PROJECT_ARI_CALLED"
exit 0
EOF
chmod +x "$PROJECT_ARI_DIR/ari"

OUTPUT=$(
    unset ARIADNE_BIN
    export PATH="/usr/bin:/bin"  # No ari in PATH
    export USE_ARI_HOOKS=1
    export CLAUDE_PROJECT_DIR="$TEST_DIR"
    "$ARI_HOOKS_DIR/context.sh" 2>&1
) || true

if echo "$OUTPUT" | grep -q "PROJECT_ARI_CALLED"; then
    pass "Project-relative fallback works"
else
    fail "Project-relative fallback did not work (output: $OUTPUT)"
fi
echo ""

# =============================================================================
# Summary
# =============================================================================
echo "==================================="
echo "Results: $PASS_COUNT passed, $FAIL_COUNT failed"
echo "==================================="

if [[ $FAIL_COUNT -gt 0 ]]; then
    exit 1
fi

echo ""
echo "=== All Tests Passed ==="

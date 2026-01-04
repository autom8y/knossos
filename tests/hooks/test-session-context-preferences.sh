#!/bin/bash
# Integration test: Validate preferences integration in session-context hook
set -euo pipefail

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo "=== Session Context Preferences Integration Test ==="
echo ""

# Test 1: Condensed output contains preference rows
echo "Test 1: Condensed output contains Autonomy and On Failure rows"
OUTPUT=$("$PROJECT_DIR/.claude/hooks/context-injection/session-context.sh" 2>/dev/null)
if echo "$OUTPUT" | grep -q "| \*\*Autonomy\*\*" && echo "$OUTPUT" | grep -q "| \*\*On Failure\*\*"; then
    echo "✓ PASS: Condensed output contains preference rows"
else
    echo "✗ FAIL: Condensed output missing preference rows"
    exit 1
fi
echo ""

# Test 2: Verbose output contains User Preferences section
echo "Test 2: Verbose output contains User Preferences section"
VERBOSE_OUTPUT=$("$PROJECT_DIR/.claude/hooks/context-injection/session-context.sh" --verbose 2>/dev/null)
if echo "$VERBOSE_OUTPUT" | grep -q "### User Preferences"; then
    echo "✓ PASS: Verbose output contains User Preferences section"
else
    echo "✗ FAIL: Verbose output missing User Preferences section"
    exit 1
fi
echo ""

# Test 3: Verbose output shows all 5 key preferences
echo "Test 3: Verbose output shows all 5 key preferences"
PREF_COUNT=$(echo "$VERBOSE_OUTPUT" | grep -A10 "### User Preferences" | grep "| \*\*" | wc -l | tr -d ' ')
if [[ "$PREF_COUNT" -ge 5 ]]; then
    echo "✓ PASS: Verbose output shows $PREF_COUNT preference rows"
else
    echo "✗ FAIL: Expected 5+ preference rows, got $PREF_COUNT"
    exit 1
fi
echo ""

# Test 4: Graceful degradation without preferences file
echo "Test 4: Hook works without preferences file"
TEST_PROJECT="$TEST_DIR/test-project"
mkdir -p "$TEST_PROJECT/.claude/hooks/lib"
mkdir -p "$TEST_PROJECT/.claude/hooks/context-injection"

# Copy required files
cp -r "$PROJECT_DIR/.claude/hooks/lib/"*.sh "$TEST_PROJECT/.claude/hooks/lib/" 2>/dev/null || true
cp "$PROJECT_DIR/.claude/hooks/context-injection/session-context.sh" "$TEST_PROJECT/.claude/hooks/context-injection/"

cd "$TEST_PROJECT"
FALLBACK_OUTPUT=$(.claude/hooks/context-injection/session-context.sh 2>/dev/null || true)
if [[ -n "$FALLBACK_OUTPUT" ]]; then
    echo "✓ PASS: Hook executes without preferences file"
else
    echo "✗ FAIL: Hook failed without preferences file"
    exit 1
fi
echo ""

# Test 5: Default values used when preferences file missing
echo "Test 5: Default preference values used when file missing"
if echo "$FALLBACK_OUTPUT" | grep -q "| \*\*Autonomy\*\* | interactive |" && \
   echo "$FALLBACK_OUTPUT" | grep -q "| \*\*On Failure\*\* | ask |"; then
    echo "✓ PASS: Default preference values (interactive, ask) displayed"
else
    echo "✗ FAIL: Default values not correctly applied"
    echo "Output: $FALLBACK_OUTPUT"
    exit 1
fi
echo ""

# Test 6: Environment variables exported
echo "Test 6: ROSTER_PREF_* environment variables exported"
cd "$PROJECT_DIR"

# Check bash version first
if ((BASH_VERSINFO[0] < 4)); then
    echo "⚠ SKIP: bash 4+ required for preferences-loader (found: $BASH_VERSION)"
    echo "   Note: Hook still works with graceful degradation on bash 3.x"
else
    export CLAUDE_PROJECT_DIR="$PROJECT_DIR"
    source .claude/hooks/lib/hooks-init.sh
    hooks_init "test" "RECOVERABLE"
    source .claude/hooks/lib/session-utils.sh
    source .claude/hooks/lib/preferences-loader.sh
    load_user_preferences
    export_preferences_env

    if [[ -n "${ROSTER_PREF_AUTONOMY_LEVEL:-}" ]] && \
       [[ -n "${ROSTER_PREF_FAILURE_HANDLING:-}" ]] && \
       [[ -n "${ROSTER_PREF_OUTPUT_FORMAT:-}" ]]; then
        echo "✓ PASS: Environment variables exported (AUTONOMY=$ROSTER_PREF_AUTONOMY_LEVEL, FAILURE=$ROSTER_PREF_FAILURE_HANDLING, OUTPUT=$ROSTER_PREF_OUTPUT_FORMAT)"
    else
        echo "✗ FAIL: Environment variables not exported"
        exit 1
    fi
fi
echo ""

echo "=== All Tests Passed ==="

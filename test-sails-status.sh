#!/bin/bash
# Test script to verify sails color appears in session status output

set -euo pipefail

# Create test directory
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

echo "Test directory: $TEST_DIR"

# Create session structure
SESSION_ID="session-20260106-test-sails"
SESSION_DIR="$TEST_DIR/.claude/sessions/$SESSION_ID"
mkdir -p "$SESSION_DIR"
mkdir -p "$TEST_DIR/.claude/sessions/.locks"

# Create SESSION_CONTEXT.md
cat > "$SESSION_DIR/SESSION_CONTEXT.md" <<'EOF'
---
schema_version: "2.1"
session_id: session-20260106-test-sails
status: ACTIVE
initiative: Test Sails Display
complexity: MODULE
created_at: 2026-01-06T10:00:00Z
current_phase: implementation
---

# Session Context
EOF

# Create WHITE_SAILS.yaml
cat > "$SESSION_DIR/WHITE_SAILS.yaml" <<'EOF'
schema_version: "1.0"
session_id: session-20260106-test-sails
generated_at: 2026-01-06T10:30:00Z
color: WHITE
computed_base: WHITE
proofs:
  tests:
    status: PASS
    summary: All tests passed
  build:
    status: PASS
    summary: Build successful
  lint:
    status: PASS
    summary: No lint errors
open_questions: []
complexity: MODULE
type: standard
EOF

# Set current session
echo "$SESSION_ID" > "$TEST_DIR/.claude/sessions/.current-session"

echo ""
echo "=== Testing with WHITE sails ==="
echo ""
echo "Session directory structure:"
tree "$TEST_DIR/.claude" 2>/dev/null || find "$TEST_DIR/.claude" -type f

echo ""
echo "Running: ari session status (JSON format)"
cd "$TEST_DIR"
# Note: This would require the ari binary to be built and in PATH
# For now, just show what the structure looks like
echo "Would execute: ari session status --output json"
echo ""
echo "Expected JSON output should include:"
echo '  "sails_color": "WHITE",'
echo '  "sails_base": "WHITE"'

echo ""
echo "=== Testing with GRAY sails (downgraded) ==="
# Update WHITE_SAILS.yaml with GRAY
cat > "$SESSION_DIR/WHITE_SAILS.yaml" <<'EOF'
schema_version: "1.0"
session_id: session-20260106-test-sails
generated_at: 2026-01-06T10:30:00Z
color: GRAY
computed_base: WHITE
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
open_questions: []
modifiers:
  - type: DOWNGRADE_TO_GRAY
    justification: Uncertainty about edge cases
    applied_by: agent
complexity: MODULE
type: standard
EOF

echo ""
echo "Expected JSON output should include:"
echo '  "sails_color": "GRAY",'
echo '  "sails_base": "WHITE"'

echo ""
echo "Expected text output should include:"
echo "  Sails: GRAY (base: WHITE)"

echo ""
echo "=== Testing without WHITE_SAILS.yaml ==="
rm "$SESSION_DIR/WHITE_SAILS.yaml"

echo ""
echo "Expected JSON output should NOT include sails_color field (omitempty)"
echo "Expected text output should include:"
echo "  Sails: not generated"

echo ""
echo "Test structure created successfully!"
echo "To test manually:"
echo "  cd $TEST_DIR"
echo "  ari session status --output json"
echo "  ari session status --output text"

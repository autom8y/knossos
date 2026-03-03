#!/usr/bin/env bash
# e2e-validate.sh -- E2E distribution validation for ari CLI
#
# Validates the full install-to-use pipeline from a pristine environment:
#   brew tap autom8y/tap
#   brew install ari (or autom8y/tap/ari)
#   ari version
#   ari init
#   ari sync --rite 10x-dev
#   .claude/ structure assertions
#
# DELIBERATE SHORTCUTS (prototype -- not production):
#   - No test framework; plain bash with pass/fail output
#   - Exit 1 on first failure (no "run all, report all" mode)
#   - Version auto-detection falls back to "any non-dev version" if gh CLI unavailable
#   - --skip-install skips tap+install entirely; assumes ari is already on PATH
#   - Linuxbrew path handling is best-effort (checked /home/linuxbrew/.linuxbrew/bin)
#
# Timeout Handling:
#   --brew-timeout SECS   Timeout for brew tap + install (default: 300s / 5 min)
#   --ari-timeout SECS    Timeout for ari init + sync (default: 60s / 1 min)
#   Requires GNU coreutils `timeout` on macOS (brew install coreutils) or Linux timeout.
#   If `timeout` is not available, brew operations run without time limit.
#
# Usage:
#   ./scripts/e2e-validate.sh
#   ./scripts/e2e-validate.sh --version v0.3.0
#   ./scripts/e2e-validate.sh --skip-install
#   ./scripts/e2e-validate.sh --version v0.3.0 --skip-install
#   ./scripts/e2e-validate.sh --brew-timeout 600 --ari-timeout 120

set -euo pipefail

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------
EXPECTED_VERSION=""
SKIP_INSTALL=false
REPO="autom8y/knossos"
TAP="autom8y/tap"
FORMULA="autom8y/tap/ari"
BREW_TIMEOUT=300   # seconds for brew tap + install (can stall in CI)
ARI_TIMEOUT=60     # seconds for ari init + sync

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)
            EXPECTED_VERSION="${2:?--version requires a value}"
            shift 2
            ;;
        --skip-install)
            SKIP_INSTALL=true
            shift
            ;;
        --brew-timeout)
            BREW_TIMEOUT="${2:?--brew-timeout requires a value}"
            shift 2
            ;;
        --ari-timeout)
            ARI_TIMEOUT="${2:?--ari-timeout requires a value}"
            shift 2
            ;;
        -h|--help)
            sed -n '/^# e2e-validate/,/^$/p' "$0"
            exit 0
            ;;
        *)
            echo "Unknown argument: $1" >&2
            exit 1
            ;;
    esac
done

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
PASS_COUNT=0
FAIL_COUNT=0

pass() {
    PASS_COUNT=$((PASS_COUNT + 1))
    echo "  PASS: $1"
}

fail() {
    FAIL_COUNT=$((FAIL_COUNT + 1))
    echo "  FAIL: $1" >&2
    echo ""
    echo "Assertion failed. Stopping." >&2
    exit 1
}

assert_exit0() {
    local label="$1"
    shift
    local output
    if output=$("$@" 2>&1); then
        pass "$label"
    else
        echo "  CMD:  $*" >&2
        echo "  OUT:  $output" >&2
        fail "$label -- command exited non-zero"
    fi
}

# run_with_timeout SECS LABEL cmd [args...]
# Runs cmd with a hard wall-clock timeout. If the timeout binary is not available,
# runs without time limit and prints a warning.
# Exits 1 (via fail) if cmd times out or returns non-zero.
run_with_timeout() {
    local secs="$1"
    local label="$2"
    shift 2

    if [[ -n "$TIMEOUT_BIN" ]]; then
        local output exit_code
        output=$("$TIMEOUT_BIN" "$secs" "$@" 2>&1)
        exit_code=$?
        if [[ $exit_code -eq 124 ]]; then
            echo "  CMD:     $*" >&2
            echo "  TIMEOUT: exceeded ${secs}s" >&2
            fail "$label -- timed out after ${secs}s (increase --brew-timeout or --ari-timeout)"
        elif [[ $exit_code -ne 0 ]]; then
            echo "  CMD:  $*" >&2
            echo "  OUT:  $output" >&2
            fail "$label -- command exited non-zero (exit $exit_code)"
        fi
        pass "$label"
    else
        # Fallback: no timeout binary; run without time limit
        assert_exit0 "$label (no timeout binary -- running without limit)" "$@"
    fi
}

assert_contains() {
    local label="$1"
    local text="$2"
    local pattern="$3"
    if echo "$text" | grep -q "$pattern"; then
        pass "$label"
    else
        echo "  EXPECTED: pattern '$pattern'" >&2
        echo "  GOT:      $text" >&2
        fail "$label -- pattern not found"
    fi
}

assert_file_exists_nonempty() {
    local label="$1"
    local path="$2"
    if [[ -f "$path" && -s "$path" ]]; then
        pass "$label"
    elif [[ ! -f "$path" ]]; then
        fail "$label -- file does not exist: $path"
    else
        fail "$label -- file exists but is empty: $path"
    fi
}

assert_dir_exists() {
    local label="$1"
    local path="$2"
    if [[ -d "$path" ]]; then
        pass "$label"
    else
        fail "$label -- directory does not exist: $path"
    fi
}

# ---------------------------------------------------------------------------
# Environment detection
# ---------------------------------------------------------------------------
echo ""
echo "========================================================"
echo "  ari CLI -- E2E Distribution Validation"
echo "========================================================"
echo ""
echo "Platform:  $(uname -s)/$(uname -m)"
echo "Date:      $(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Detect timeout binary (gtimeout on macOS via coreutils, timeout on Linux).
TIMEOUT_BIN=""
if command -v gtimeout &>/dev/null; then
    TIMEOUT_BIN="gtimeout"
elif command -v timeout &>/dev/null; then
    TIMEOUT_BIN="timeout"
fi
if [[ -n "$TIMEOUT_BIN" ]]; then
    echo "Timeout:   $TIMEOUT_BIN (brew=${BREW_TIMEOUT}s, ari=${ARI_TIMEOUT}s)"
else
    echo "Timeout:   NONE (install coreutils for timeout enforcement: brew install coreutils)"
fi

# Detect brew binary (Linuxbrew installs to /home/linuxbrew/.linuxbrew/bin)
BREW_BIN="brew"
if ! command -v brew &>/dev/null; then
    if [[ -x "/home/linuxbrew/.linuxbrew/bin/brew" ]]; then
        BREW_BIN="/home/linuxbrew/.linuxbrew/bin/brew"
        export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"
    else
        echo "ERROR: brew not found on PATH and Linuxbrew not at default location." >&2
        echo "       Install Homebrew first: https://brew.sh" >&2
        exit 1
    fi
fi

echo "Brew:      $($BREW_BIN --version 2>&1 | head -1)"
echo "Skip install: $SKIP_INSTALL"
echo ""

# ---------------------------------------------------------------------------
# Resolve expected version
# ---------------------------------------------------------------------------
if [[ -z "$EXPECTED_VERSION" ]]; then
    if command -v gh &>/dev/null; then
        EXPECTED_VERSION=$(gh release view --repo "$REPO" --json tagName --jq '.tagName' 2>/dev/null || echo "")
    fi
fi

if [[ -n "$EXPECTED_VERSION" ]]; then
    # Strip leading 'v' for comparison (ari version may output "0.3.0" or "v0.3.0")
    VERSION_BARE="${EXPECTED_VERSION#v}"
    echo "Expected version: $EXPECTED_VERSION (bare: $VERSION_BARE)"
else
    echo "Expected version: (any non-dev version -- gh CLI not available)"
fi
echo ""

# ---------------------------------------------------------------------------
# ASSERTION 1: brew tap autom8y/tap
# ---------------------------------------------------------------------------
echo "--- Assertion 1: brew tap $TAP ---"
if [[ "$SKIP_INSTALL" == "true" ]]; then
    echo "  SKIP: --skip-install set"
else
    run_with_timeout "$BREW_TIMEOUT" "brew tap $TAP exits 0 within ${BREW_TIMEOUT}s" \
        "$BREW_BIN" tap "$TAP"
fi

# ---------------------------------------------------------------------------
# ASSERTION 2: brew install ari
# ---------------------------------------------------------------------------
echo "--- Assertion 2: brew install $FORMULA ---"
if [[ "$SKIP_INSTALL" == "true" ]]; then
    echo "  SKIP: --skip-install set"
else
    run_with_timeout "$BREW_TIMEOUT" "brew install $FORMULA exits 0 within ${BREW_TIMEOUT}s" \
        "$BREW_BIN" install "$FORMULA"
fi

# Ensure ari is on PATH after install
if ! command -v ari &>/dev/null; then
    # Linuxbrew may not have updated PATH yet
    ARI_PATH=$("$BREW_BIN" --prefix "$FORMULA" 2>/dev/null)/bin/ari
    if [[ -x "$ARI_PATH" ]]; then
        export PATH="$(dirname "$ARI_PATH"):$PATH"
    else
        fail "ari not found on PATH after brew install"
    fi
fi

# ---------------------------------------------------------------------------
# ASSERTION 3: ari version output contains expected version string
# ---------------------------------------------------------------------------
echo "--- Assertion 3: ari version ---"
VERSION_OUTPUT=$(ari version 2>&1 || true)
if [[ -n "$EXPECTED_VERSION" ]]; then
    assert_contains "ari version contains $EXPECTED_VERSION" "$VERSION_OUTPUT" "$VERSION_BARE"
else
    # When no expected version is known, just verify it doesn't say "dev"
    if echo "$VERSION_OUTPUT" | grep -q "dev"; then
        fail "ari version contains 'dev' -- release binary expected"
    fi
    pass "ari version does not contain 'dev' (version: $VERSION_OUTPUT)"
fi

# ---------------------------------------------------------------------------
# ASSERTION 4: ari init in a temp directory exits 0
# ---------------------------------------------------------------------------
echo "--- Assertion 4: ari init ---"
E2E_TMPDIR=$(mktemp -d)
trap 'rm -rf "$E2E_TMPDIR"' EXIT

pushd "$E2E_TMPDIR" > /dev/null
run_with_timeout "$ARI_TIMEOUT" "ari init exits 0 within ${ARI_TIMEOUT}s" ari init
popd > /dev/null

# ---------------------------------------------------------------------------
# ASSERTION 5: ari sync --rite 10x-dev exits 0
# ---------------------------------------------------------------------------
echo "--- Assertion 5: ari sync --rite 10x-dev ---"
pushd "$E2E_TMPDIR" > /dev/null
run_with_timeout "$ARI_TIMEOUT" "ari sync --rite 10x-dev exits 0 within ${ARI_TIMEOUT}s" \
    ari sync --rite 10x-dev
popd > /dev/null

# ---------------------------------------------------------------------------
# ASSERTION 6: .claude/CLAUDE.md exists and is non-empty
# ---------------------------------------------------------------------------
echo "--- Assertion 6: .claude/CLAUDE.md exists and non-empty ---"
assert_file_exists_nonempty ".claude/CLAUDE.md exists and non-empty" "$E2E_TMPDIR/.claude/CLAUDE.md"

# Bonus: verify it contains KNOSSOS section markers (smoke test for real content)
CLAUDE_MD_CONTENT=$(cat "$E2E_TMPDIR/.claude/CLAUDE.md")
assert_contains ".claude/CLAUDE.md contains KNOSSOS section markers" "$CLAUDE_MD_CONTENT" "KNOSSOS:START"

# ---------------------------------------------------------------------------
# ASSERTION 7: .claude/ contains expected structure
# ---------------------------------------------------------------------------
echo "--- Assertion 7: .claude/ directory structure ---"
assert_dir_exists ".claude/agents/ exists" "$E2E_TMPDIR/.claude/agents"
assert_dir_exists ".claude/commands/ exists" "$E2E_TMPDIR/.claude/commands"
assert_dir_exists ".claude/skills/ exists" "$E2E_TMPDIR/.claude/skills"
assert_file_exists_nonempty ".claude/settings.local.json exists and non-empty" "$E2E_TMPDIR/.claude/settings.local.json"

# Verify at least one agent file exists (ensures rite content actually materialized)
AGENT_COUNT=$(find "$E2E_TMPDIR/.claude/agents" -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
if [[ "$AGENT_COUNT" -gt 0 ]]; then
    pass ".claude/agents/ contains $AGENT_COUNT agent file(s)"
else
    fail ".claude/agents/ is empty -- rite sync did not materialize agents"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "========================================================"
echo "  Results: $PASS_COUNT passed, $FAIL_COUNT failed"
echo "========================================================"
echo ""

if [[ "$FAIL_COUNT" -gt 0 ]]; then
    exit 1
fi

echo "All assertions passed. Distribution pipeline is healthy."
echo ""

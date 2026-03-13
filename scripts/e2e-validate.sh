#!/usr/bin/env bash
# e2e-validate.sh -- E2E distribution validation for ari CLI
#
# Validates the full install-to-use pipeline from a pristine environment:
#   brew tap autom8y/tap
#   brew install ari (or autom8y/tap/ari)
#   ari version
#   ari init
#   ari sync --rite 10x-dev
#   channel directory structure assertions
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
# Channel Handling:
#   --channel claude      Validate Claude channel output (default)
#   --channel gemini      Validate Gemini channel output
#   --channel all         Validate both channels + structural parity
#
# Usage:
#   ./scripts/e2e-validate.sh
#   ./scripts/e2e-validate.sh --version v0.3.0
#   ./scripts/e2e-validate.sh --skip-install
#   ./scripts/e2e-validate.sh --version v0.3.0 --skip-install
#   ./scripts/e2e-validate.sh --brew-timeout 600 --ari-timeout 120
#   ./scripts/e2e-validate.sh --channel claude
#   ./scripts/e2e-validate.sh --channel gemini
#   ./scripts/e2e-validate.sh --channel all

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
CHANNEL="claude"   # default: backward-compatible

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
        --channel)
            CHANNEL="${2:?--channel requires a value (claude|gemini|all)}"
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

# Validate channel value
case "$CHANNEL" in
    claude|gemini|all) ;;
    *) echo "ERROR: --channel must be claude, gemini, or all" >&2; exit 1 ;;
esac

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
# Channel helper functions
# ---------------------------------------------------------------------------

# channel_dir returns the channel directory name
channel_dir() {
    case "$1" in
        claude) echo ".claude" ;;
        gemini) echo ".gemini" ;;
    esac
}

# context_file returns the context file name (CLAUDE.md or GEMINI.md)
context_file() {
    case "$1" in
        claude) echo "CLAUDE.md" ;;
        gemini) echo "GEMINI.md" ;;
    esac
}

# has_settings_local returns 0 (true) if the channel has settings.local.json
has_settings_local() {
    case "$1" in
        claude) return 0 ;;
        gemini) return 1 ;;  # Gemini does NOT have settings.local.json
    esac
}

# validate_sync runs ari sync for a specific channel
validate_sync() {
    local ch="$1"
    echo "--- Assertion 5($ch): ari sync --rite 10x-dev --channel $ch ---"
    pushd "$E2E_TMPDIR" > /dev/null
    run_with_timeout "$ARI_TIMEOUT" \
        "ari sync --rite 10x-dev --channel $ch exits 0 within ${ARI_TIMEOUT}s" \
        ari sync --rite 10x-dev --channel "$ch"
    popd > /dev/null
}

# validate_channel_output checks assertions 6-7 for a specific channel
validate_channel_output() {
    local ch="$1"
    local chdir
    chdir=$(channel_dir "$ch")
    local ctxfile
    ctxfile=$(context_file "$ch")

    echo "--- Assertion 6($ch): $chdir/$ctxfile exists and non-empty ---"
    assert_file_exists_nonempty \
        "$chdir/$ctxfile exists and non-empty" \
        "$E2E_TMPDIR/$chdir/$ctxfile"

    # Verify KNOSSOS section markers
    local content
    content=$(cat "$E2E_TMPDIR/$chdir/$ctxfile")
    assert_contains "$chdir/$ctxfile contains KNOSSOS section markers" \
        "$content" "KNOSSOS:START"

    echo "--- Assertion 7($ch): $chdir/ directory structure ---"
    assert_dir_exists "$chdir/agents/ exists" "$E2E_TMPDIR/$chdir/agents"
    assert_dir_exists "$chdir/commands/ exists" "$E2E_TMPDIR/$chdir/commands"
    assert_dir_exists "$chdir/skills/ exists" "$E2E_TMPDIR/$chdir/skills"

    # settings.local.json: Claude-only
    if has_settings_local "$ch"; then
        assert_file_exists_nonempty \
            "$chdir/settings.local.json exists and non-empty" \
            "$E2E_TMPDIR/$chdir/settings.local.json"
    fi

    # Agent count
    local agent_count
    agent_count=$(find "$E2E_TMPDIR/$chdir/agents" -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$agent_count" -gt 0 ]]; then
        pass "$chdir/agents/ contains $agent_count agent file(s)"
    else
        fail "$chdir/agents/ is empty -- rite sync did not materialize agents"
    fi
}

# validate_structural_parity checks assertion 8 (--channel all only):
# agent, command, and skill counts between channels must be within tolerance,
# and no cross-contamination of channel-specific path literals.
validate_structural_parity() {
    echo "--- Assertion 8: Structural parity between channels ---"
    local claude_agents gemini_agents claude_cmds gemini_cmds claude_skills gemini_skills

    claude_agents=$(find "$E2E_TMPDIR/.claude/agents" -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
    gemini_agents=$(find "$E2E_TMPDIR/.gemini/agents" -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
    claude_cmds=$(find "$E2E_TMPDIR/.claude/commands" -type f 2>/dev/null | wc -l | tr -d ' ')
    gemini_cmds=$(find "$E2E_TMPDIR/.gemini/commands" -type f 2>/dev/null | wc -l | tr -d ' ')
    claude_skills=$(find "$E2E_TMPDIR/.claude/skills" -type f 2>/dev/null | wc -l | tr -d ' ')
    gemini_skills=$(find "$E2E_TMPDIR/.gemini/skills" -type f 2>/dev/null | wc -l | tr -d ' ')

    # Tolerance: allow up to 2 file difference (for channel-specific items)
    local tolerance=2

    local agent_diff=$((claude_agents - gemini_agents))
    if [[ ${agent_diff#-} -le $tolerance ]]; then
        pass "Agent parity: claude=$claude_agents, gemini=$gemini_agents (within tolerance=$tolerance)"
    else
        fail "Agent parity violation: claude=$claude_agents, gemini=$gemini_agents (tolerance=$tolerance)"
    fi

    local cmd_diff=$((claude_cmds - gemini_cmds))
    if [[ ${cmd_diff#-} -le $tolerance ]]; then
        pass "Command parity: claude=$claude_cmds, gemini=$gemini_cmds (within tolerance=$tolerance)"
    else
        fail "Command parity violation: claude=$claude_cmds, gemini=$gemini_cmds (tolerance=$tolerance)"
    fi

    local skill_diff=$((claude_skills - gemini_skills))
    if [[ ${skill_diff#-} -le $tolerance ]]; then
        pass "Skill parity: claude=$claude_skills, gemini=$gemini_skills (within tolerance=$tolerance)"
    else
        fail "Skill parity violation: claude=$claude_skills, gemini=$gemini_skills (tolerance=$tolerance)"
    fi

    # Cross-contamination check
    if grep -r '\.claude/' "$E2E_TMPDIR/.gemini/" --include="*.md" -l 2>/dev/null | head -1 | grep -q .; then
        fail "Cross-contamination: .claude/ path literal found in .gemini/ output"
    else
        pass "No .claude/ path literals in .gemini/ output"
    fi
    if grep -r '\.gemini/' "$E2E_TMPDIR/.claude/" --include="*.md" -l 2>/dev/null | head -1 | grep -q .; then
        fail "Cross-contamination: .gemini/ path literal found in .claude/ output"
    else
        pass "No .gemini/ path literals in .claude/ output"
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
echo "Channel:   $CHANNEL"

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
# ASSERTION 5-7: Channel-specific validation
# ---------------------------------------------------------------------------
case "$CHANNEL" in
    claude)
        validate_sync "claude"
        validate_channel_output "claude"
        ;;
    gemini)
        validate_sync "gemini"
        validate_channel_output "gemini"
        ;;
    all)
        validate_sync "claude"
        validate_sync "gemini"
        validate_channel_output "claude"
        validate_channel_output "gemini"
        validate_structural_parity
        ;;
esac

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

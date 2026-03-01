# TDD: Context Noise Reduction (Hook Ecosystem Parity - Scope 3)

## Overview

This Technical Design Document specifies the implementation of context noise reduction across the roster hook ecosystem. The design condenses verbose hook outputs to essential information, adds targeted matchers to reduce unnecessary hook invocations, and provides a `--verbose` flag for expanded output when needed.

## Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-hook-ecosystem-parity.md` |
| Task | `task-004` (Sprint: sprint-hook-parity-20251231) |
| Requirements | FR-3.1 through FR-3.4, FR-S.2 |

### Problem Statement

The current hook ecosystem emits excessive context, consuming valuable Claude context window tokens:

1. **session-context.sh**: Outputs 53 lines per session start, including verbose tables, artifact counts, and pre-computed values that are rarely needed
2. **delegation-check.sh**: Emits 15-line warnings with redundant formatting when a simple notification suffices
3. **session-write-guard.sh**: Returns JSON with 5 examples when 1 representative example is sufficient
4. **UserPromptSubmit**: Fires `start-preflight.sh` on every prompt, even when not a slash command

### Design Goals

1. Reduce session-context.sh output from 53 lines to maximum 15 lines
2. Condense delegation-check.sh warning from 15 lines to 3 lines
3. Reduce session-write-guard.sh JSON examples from 5 to 1
4. Fire UserPromptSubmit hooks only on slash commands (`^/` prefix)
5. Preserve access to full information via `--verbose` flag or `/status` command

---

## System Design

### Information Hierarchy

The design applies a three-tier information hierarchy:

| Tier | Content | Display |
|------|---------|---------|
| **Essential** | Team, Session, Git status, Available commands | Always shown |
| **Contextual** | Workflow phase, Worktree info | Shown when relevant |
| **Verbose** | Artifact counts, Pre-computed values, Full property tables | `--verbose` only |

### Architecture

```
SessionStart Hook
       |
       v
+------------------+
| session-context  |
| (condensed)      |-----> Essential context (15 lines max)
|                  |
| --verbose flag   |-----> Full context (current behavior)
+------------------+

UserPromptSubmit Hook
       |
       v
+------------------+
| Matcher: ^/      |-----> Only slash commands
+------------------+
       |
       v
+------------------+
| start-preflight  |-----> Preflight checks
+------------------+

PreToolUse Hook (Edit|Write)
       |
       v
+------------------+       +--------------------+
| session-write-   |  or   | delegation-check   |
| guard (condensed)|       | (3-line warning)   |
+------------------+       +--------------------+
```

---

## Component Designs

### 1. session-context.sh Refactor (FR-3.1, FR-S.2)

#### Current Output (53 lines)

```markdown
## Project Context (auto-loaded)

| Property | Value |
|----------|-------|
| **Project** | /Users/tomtenuta/code/roster |
| **Worktree** | main project |
| **Active Team** | 10x-dev |
| **Workflow** | 10x-dev (entry: requirements-analyst) |
| **Git** | main (27 uncommitted) |

### Session Status

| Property | Value |
|----------|-------|
| **Has Session** | true |
| **Session State** | IDLE |
| **Session ID** | session-20251231-012324-c9fcc2d7 |
| **Initiative** | HarnessEcosystemAuditRemediation&Improvement |
| **Complexity** | SERVICE |
| **Current Phase** | implementation |
| **Parked** | false |
| **Workflow Active** | false |
| **Workflow Mode** | none |

### Artifacts
- **PRDs**: 1
- **TDDs**: 1
- **ADRs**: 0

### Pre-computed Values (for /start)
- **Suggested Session ID**: `session-20251231-130851-31a99f4c`
- **Entry Agent**: requirements-analyst
- **Sessions Directory**: `.sos/sessions/`

---

**Session Commands**:
- `/park` - Pause current session
- `/handoff <agent>` - Transfer to another agent
- `/wrap` - Finalize session

## Rite Context: 10x-dev
... (additional rite routing context)
```

#### Condensed Output (15 lines max)

```markdown
## Session Context

| | |
|---|---|
| **Rite** | 10x-dev |
| **Session** | session-20251231-012324 (ACTIVE) |
| **Initiative** | Hook Ecosystem Parity |
| **Git** | main (27 uncommitted) |

**Commands**: `/park` | `/handoff <agent>` | `/wrap` | `/status`
```

#### Design Decisions

**D1: Single Table Format**
- Combine project and session info into one compact table
- Remove redundant "Property" column header (self-evident from bold keys)
- Truncate session ID to first 23 chars (sufficient for identification)

**D2: Conditional Rows**
- Omit worktree row unless in a worktree
- Omit complexity row (rarely actionable at session start)
- Omit workflow mode/active (internal state, not user-facing)

**D3: Commands as Inline List**
- Replace multi-line command table with single inline list
- Add `/status` for verbose output access

**D4: Team Context Deferral**
- Move team routing table to `/status` or `--verbose`
- Essential team info is in the "Team" row

#### Implementation: Verbose Flag

```bash
# Parse arguments for --verbose
VERBOSE=false
for arg in "$@"; do
    [[ "$arg" == "--verbose" ]] && VERBOSE=true
done

# Also check environment variable (for hooks that can't pass args)
[[ "${ROSTER_VERBOSE:-}" == "true" ]] && VERBOSE=true

if [[ "$VERBOSE" == "true" ]]; then
    # Current full output
    output_full_context
else
    # New condensed output
    output_condensed_context
fi
```

#### Implementation: Condensed Output Function

```bash
output_condensed_context() {
    # Build compact session display
    local session_display="none"
    if [[ "$HAS_SESSION" == "true" ]]; then
        local short_id="${SESSION_ID:0:23}"
        local state_badge="$SESSION_STATE"
        [[ "$PARKED" == "true" ]] && state_badge="PARKED"
        session_display="$short_id ($state_badge)"
    fi

    # Truncate initiative to 30 chars
    local initiative_display="${INITIATIVE:0:30}"
    [[ ${#INITIATIVE} -gt 30 ]] && initiative_display="${initiative_display}..."

    cat <<EOF
## Session Context

| | |
|---|---|
| **Team** | $ACTIVE_RITE |
| **Session** | $session_display |
| **Initiative** | $initiative_display |
| **Git** | $GIT_DISPLAY |
EOF

    # Add worktree row only if in worktree
    if [[ "$IS_WORKTREE" == "true" ]]; then
        echo "| **Worktree** | $WORKTREE_ID |"
    fi

    echo ""

    # Inline commands based on state
    if [[ "$HAS_SESSION" == "true" ]]; then
        if [[ "$PARKED" == "true" ]]; then
            echo "**Commands**: \`/continue\` | \`/wrap\` | \`/status\`"
        else
            echo "**Commands**: \`/park\` | \`/handoff <agent>\` | \`/wrap\` | \`/status\`"
        fi
    else
        echo "**Commands**: \`/start <initiative>\` | \`/status\`"
    fi
}
```

#### Line Count Verification

| Section | Lines |
|---------|-------|
| Header | 1 |
| Blank | 1 |
| Table header | 2 |
| Table rows (4-5) | 5 |
| Blank | 1 |
| Commands line | 1 |
| **Total** | **11-12** |

With worktree: 12-13 lines. Target of max 15 met.

---

### 2. UserPromptSubmit Matcher (FR-3.2)

#### Current Configuration

```json
"UserPromptSubmit": [
  {
    "hooks": [
      {
        "type": "command",
        "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/start-preflight.sh",
        "timeout": 5
      }
    ]
  }
]
```

No matcher means `start-preflight.sh` fires on every user prompt.

#### Designed Configuration

```json
"UserPromptSubmit": [
  {
    "matcher": "^/",
    "hooks": [
      {
        "type": "command",
        "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/start-preflight.sh",
        "timeout": 5
      }
    ]
  }
]
```

The `^/` regex matcher ensures the hook only fires when the user prompt starts with `/`.

#### Matcher Behavior

| User Input | Matches `^/` | Hook Fires |
|------------|--------------|------------|
| `/start auth-feature` | Yes | Yes |
| `/wrap` | Yes | Yes |
| `/park` | Yes | Yes |
| `Please implement auth` | No | No |
| `What does /start do?` | No | No |
| `/ (space)` | Yes | Yes (edge case, harmless) |

#### Implementation Location

File: `.claude/settings.local.json`

The change is a JSON configuration update, not code change. This will be handled by the `swap-rite.sh` settings generation (Scope 2) or can be manually updated.

#### Fallback Behavior

The `start-preflight.sh` script already has internal pattern matching:

```bash
# Only act on session lifecycle commands
if [[ ! "$USER_PROMPT" =~ ^/(start|continue|park|wrap|sessions|worktree) ]]; then
    exit 0
fi
```

Adding the matcher at the JSON level:
1. Eliminates unnecessary shell process spawning
2. Reduces overhead from ~50ms (process spawn + bash init) to 0ms for non-slash prompts
3. Keeps script-level check as defense in depth

---

### 3. delegation-check.sh Warning Condensation (FR-3.3)

#### Current Warning (15 lines)

```
[DELEGATION WARNING]
====================
Active workflow detected: 10x-dev
Tool attempted: Edit on /tmp/test.py

The main thread should delegate implementation to specialists via Task tool.
Direct Edit/Write of code files during active workflow violates the Coach pattern.

If this is intentional (user override), proceed.
If accidental, cancel and use Task tool to invoke the appropriate specialist.

See: .claude/skills/orchestration/main-thread-guide.md
====================
```

#### Condensed Warning (3 lines)

```
[DELEGATION] Workflow active (10x-dev): Edit on /tmp/test.py
  -> Use Task tool to delegate, or proceed if intentional override.
  -> See: .claude/skills/orchestration/main-thread-guide.md
```

#### Design Decisions

**D1: Single-Line Title with Context**
- Combine warning type, workflow name, tool, and file into one line
- Use `[DELEGATION]` tag for grep-ability

**D2: Arrow Prefix for Instructions**
- `->` prefix distinguishes actionable lines from title
- Single remediation line (no conditional override explanation)

**D3: Documentation Reference Preserved**
- Keep doc link for those who need more context
- Experienced users can ignore, new users can learn

#### Implementation

```bash
# Emit condensed warning to stderr
cat >&2 <<EOF
[DELEGATION] Workflow active ($WORKFLOW_NAME): $TOOL_NAME on $FILE_PATH
  -> Use Task tool to delegate, or proceed if intentional override.
  -> See: .claude/skills/orchestration/main-thread-guide.md
EOF
```

---

### 4. session-write-guard.sh JSON Condensation (FR-3.4)

#### Current JSON Output (5 examples)

```json
{
  "decision": "block",
  "reason": "Direct writes to *_CONTEXT.md files are not allowed. Use state-mate agent for all session/sprint state mutations.",
  "instruction": "Use the Task tool to invoke state-mate: Task(moirai, 'your mutation request')",
  "examples": [
    "Task(moirai, 'update_field status=completed')",
    "Task(moirai, 'mark_complete task-001 artifact=docs/design/TDD-foo.md')",
    "Task(moirai, 'park_session reason=\"Taking a break\"')",
    "Task(moirai, 'transition_phase from=design to=implementation')",
    "Task(moirai, '--dry-run mark_complete task-001 artifact=...')"
  ],
  "documentation": "user-agents/state-mate.md"
}
```

#### Condensed JSON Output (1 example)

```json
{
  "decision": "block",
  "reason": "Direct writes to *_CONTEXT.md files are blocked. Use state-mate for state mutations.",
  "instruction": "Task(moirai, 'your mutation request')",
  "example": "Task(moirai, 'mark_complete task-001 artifact=docs/design/TDD-foo.md')",
  "documentation": "user-agents/state-mate.md"
}
```

#### Design Decisions

**D1: Shorten Reason**
- Remove redundant "not allowed" (implicit in "blocked")
- Remove explanation of what state-mate is (link provides that)

**D2: Single Example**
- Change `examples` array to singular `example`
- Use `mark_complete` as representative (most common operation)
- Other examples are in documentation

**D3: Simplify Instruction**
- Remove preamble "Use the Task tool to invoke state-mate:"
- Direct invocation pattern is sufficient

#### Line/Character Reduction

| Metric | Before | After | Reduction |
|--------|--------|-------|-----------|
| JSON lines | 12 | 7 | 42% |
| Characters | 548 | 289 | 47% |
| Examples | 5 | 1 | 80% |

---

## Settings Configuration

### Updated settings.local.json Hooks Section

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|resume",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/session-context.sh",
            "timeout": 10
          },
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/coach-mode.sh",
            "timeout": 5
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "^/",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/start-preflight.sh",
            "timeout": 5
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/session-write-guard.sh",
            "timeout": 3
          },
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/delegation-check.sh",
            "timeout": 3
          }
        ]
      }
    ]
  }
}
```

Key change: `UserPromptSubmit` now has `"matcher": "^/"`.

---

## Non-Functional Considerations

### Performance

| Hook | Before | After | Improvement |
|------|--------|-------|-------------|
| session-context.sh | ~150ms | ~80ms | 47% (fewer operations) |
| start-preflight.sh | ~50ms/prompt | 0ms (non-slash) | 100% for typical prompts |
| delegation-check.sh | ~20ms | ~15ms | 25% (less output) |
| session-write-guard.sh | ~5ms | ~3ms | 40% (less output) |

### Context Token Impact

| Hook | Before (tokens) | After (tokens) | Savings |
|------|-----------------|----------------|---------|
| session-context.sh | ~400 | ~80 | 320 tokens |
| delegation-check.sh | ~100 | ~40 | 60 tokens |
| session-write-guard.sh | ~120 | ~60 | 60 tokens |

Estimated total savings per session: **~440 tokens** (assuming one warning each).

### Backward Compatibility

1. **--verbose flag**: Full output accessible for scripts/users expecting current format
2. **ROSTER_VERBOSE env var**: Alternative for hooks that cannot pass arguments
3. **/status command**: Reference to expanded information in condensed output
4. **Internal checks preserved**: start-preflight.sh still has pattern matching as fallback

---

## Implementation Guidance

### Recommended Implementation Order

1. **Phase 1: session-context.sh refactor**
   - Add `--verbose` flag parsing
   - Implement `output_condensed_context()` function
   - Add ROSTER_VERBOSE env var support
   - Maintain `output_full_context()` as current behavior

2. **Phase 2: delegation-check.sh condensation**
   - Replace heredoc with 3-line format
   - No flag needed (warning is always shown when triggered)

3. **Phase 3: session-write-guard.sh condensation**
   - Update JSON heredoc
   - Change `examples` array to `example` string

4. **Phase 4: settings.local.json matcher update**
   - Add `"matcher": "^/"` to UserPromptSubmit
   - This can be done during Scope 2 settings generation or as standalone change

### Testing Strategy

| Test | Method | Validation |
|------|--------|------------|
| session-context line count | `session-context.sh \| wc -l` | <= 15 |
| session-context verbose | `session-context.sh --verbose \| wc -l` | >= 50 |
| delegation-check line count | Trigger warning, count lines | = 3 |
| session-write-guard JSON | Parse output, count examples | = 1 |
| UserPromptSubmit matcher | Log hook invocations, verify slash-only | No fires on normal prompts |

### Verification Commands

```bash
# Test condensed output
bash hooks/session-context.sh | wc -l
# Expected: <= 15

# Test verbose output
bash hooks/session-context.sh --verbose | wc -l
# Expected: ~53 (current)

# Test delegation warning (requires active workflow)
# Verify 3 lines in hook output

# Test write guard JSON
bash hooks/session-write-guard.sh <<< '{"tool_name":"Write"}' | jq '.example'
# Expected: single string, not array
```

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Essential info omitted | Low | Medium | User testing, --verbose fallback |
| Regex matcher too restrictive | Low | Low | Keeps script-level check as backup |
| Existing scripts depend on output format | Low | Medium | --verbose preserves current format |
| /status command not implemented | N/A | Low | Out of scope; commands inline suffice |

---

## Success Criteria

- [ ] `session-context.sh` output is 15 lines or fewer (measured)
- [ ] `session-context.sh --verbose` produces full output (current behavior)
- [ ] `delegation-check.sh` warning is exactly 3 lines
- [ ] `session-write-guard.sh` JSON has 1 example (not array)
- [ ] UserPromptSubmit only fires on `/` prefixed prompts
- [ ] All existing hook functionality preserved
- [ ] Context token savings verified (~440 tokens)

---

## ADRs

No new ADRs required for Scope 3. This is implementation detail within the existing hook architecture defined in ADR-0002.

---

## Open Items

| Item | Status | Owner | Notes |
|------|--------|-------|-------|
| `/status` command implementation | Deferred | Future Sprint | Referenced in condensed output but not in Scope 3 |
| ROSTER_VERBOSE documentation | Pending | Principal Engineer | Document env var in README |

---

## Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-hook-parity-scope3.md` | Created |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hook-ecosystem-parity.md` | Read |
| session-context.sh | `/Users/tomtenuta/Code/roster/hooks/session-context.sh` | Read |
| delegation-check.sh | `/Users/tomtenuta/Code/roster/hooks/delegation-check.sh` | Read |
| session-write-guard.sh | `/Users/tomtenuta/Code/roster/hooks/session-write-guard.sh` | Read |
| settings.local.json | `/Users/tomtenuta/Code/roster/.claude/settings.local.json` | Read |

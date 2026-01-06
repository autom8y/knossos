# TDD: Orchestrator Enforcement with Complexity Gating

## Overview

This Technical Design Document specifies complexity-aware orchestrator enforcement that gates bypass behavior based on session complexity level. The design extends the existing enforcement infrastructure (delegation-check.sh, orchestrator-bypass-check.sh) to differentiate between SCRIPT/MODULE and SERVICE/PLATFORM complexity, providing stronger guardrails for complex work while maintaining lightweight behavior for simpler tasks.

## Context

| Reference | Location |
|-----------|----------|
| Problem Statement | Main thread bypasses orchestrator during MODULE/SERVICE work, breaking session coherence |
| Existing delegation-check.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/delegation-check.sh` |
| Existing orchestrator-bypass-check.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-bypass-check.sh` |
| Existing orchestration-audit.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/orchestration-audit.sh` |
| Session complexity field | `SESSION_CONTEXT.md` frontmatter `complexity: enum` |
| Complexity levels | SCRIPT, MODULE, SERVICE, PLATFORM (ordered low to high) |
| Prior art | `docs/ecosystem/CONTEXT-DESIGN-orchestration-mode-consolidation.md` |

### Problem Statement

Current enforcement hooks treat all complexity levels equally:
1. **delegation-check.sh** (lines 60-72): Warns on Edit/Write during workflow, no complexity awareness
2. **orchestrator-bypass-check.sh** (lines 94-116): Warns on Task without orchestrator consultation, same for all complexity levels

This creates two issues:
1. **Under-enforcement for complex work**: SERVICE/PLATFORM work can bypass orchestration with only a warning, losing quality gate benefits
2. **Over-enforcement for simple work**: SCRIPT/PATCH work gets the same friction as complex work, adding unnecessary ceremony

### Design Goals

1. **Complexity-appropriate enforcement**: Stronger guardrails for SERVICE/PLATFORM, lighter for SCRIPT/PATCH
2. **Clear override mechanism**: Explicit, auditable way to bypass enforcement when intentional
3. **Enhanced audit trail**: Log complexity level in all bypass events
4. **Backward compatibility**: Existing sessions without complexity field degrade gracefully

---

## Design Decisions

### Decision 1: Enforcement Tiers by Complexity

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Binary (simple/complex) | SCRIPT/MODULE = warn, SERVICE/PLATFORM = block | Simple, clear boundary | MODULE is often substantial |
| B. Three tiers | SCRIPT = skip, MODULE = warn, SERVICE+ = block | Fine-grained | More states to test |
| C. Graduated (warn/acknowledge/block) | SCRIPT = warn, MODULE = acknowledge, SERVICE+ = block | Progressive friction | Acknowledge mechanism complex |
| D. Custom per-level | Each complexity has unique behavior | Maximum flexibility | Harder to reason about |

**Selected**: Option C - Graduated enforcement with three tiers

**Rationale**: The three-tier model maps naturally to user expectations:
- SCRIPT: Quick tasks that benefit from minimal friction - warn only (current behavior)
- MODULE: Substantial work where bypass should be deliberate - stronger warning with acknowledgment prompt
- SERVICE/PLATFORM: Complex work where orchestration provides most value - block by default, require explicit override

**Enforcement Matrix**:

| Complexity | Edit/Write (delegation-check) | Task without consultation (bypass-check) | Override Required |
|------------|-------------------------------|------------------------------------------|-------------------|
| SCRIPT | Warn (current behavior) | Warn (current behavior) | No |
| MODULE | Warn + acknowledgment prompt | Warn + acknowledgment prompt | No (but prompted) |
| SERVICE | Block by default | Block by default | Yes |
| PLATFORM | Block by default | Block by default | Yes |

---

### Decision 2: Override Mechanism

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Environment variable | `CLAUDE_BYPASS_ORCHESTRATOR=1` | Standard pattern | Session-wide, not targeted |
| B. Session-level setting | `SESSION_CONTEXT.md` field `orchestrator_override: true` | Persists in session | Requires state-mate to set |
| C. User message pattern | "Intentional override: [reason]" in prompt | Natural, in-band | Cannot check before tool use |
| D. Hook input parameter | Override flag in tool input | Targeted | Claude cannot set hook params |
| E. Hybrid: Session setting + environment | Setting for session, env for emergency | Flexible | Two mechanisms |

**Selected**: Option E - Hybrid approach with session setting as primary and environment variable as emergency fallback

**Rationale**:
1. **Session setting** (`orchestrator_override: reason`) is the proper mechanism:
   - Set via state-mate for audit trail
   - Persists across tool invocations
   - Can be scoped (e.g., next operation only vs. rest of session)
   - Requires explicit action, not accidental

2. **Environment variable** (`CLAUDE_BYPASS_ORCHESTRATOR=1`) is emergency fallback:
   - For cases where session mutation is problematic
   - Logged but not recommended
   - Terminal-scoped, not session-scoped

**Override Interface**:

```yaml
# In SESSION_CONTEXT.md (set via state-mate)
orchestrator_override:
  enabled: true|false
  reason: "Emergency hotfix for production incident"
  scope: "next"|"session"  # next = one operation, session = rest of session
  set_at: "2026-01-02T12:00:00Z"
  set_by: "user"
```

**state-mate Command**:
```
Task(moirai, "enable orchestrator override reason='Emergency hotfix' scope=next")
Task(moirai, "disable orchestrator override")
```

---

### Decision 3: Complexity Detection Mechanism

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Direct YAML parsing | grep/sed from SESSION_CONTEXT.md | Simple, self-contained | Duplicate logic |
| B. Use get_complexity() | Existing function in session-state.sh | DRY, single source | Requires sourcing library |
| C. Cache in environment | Set CLAUDE_SESSION_COMPLEXITY at session start | Fast reads | Stale if session changes |

**Selected**: Option B - Use existing `get_complexity()` from session-state.sh

**Rationale**:
- `get_complexity()` already exists in session-state.sh (line 127-131)
- Uses `get_session_field()` which properly handles YAML parsing
- Single source of truth for complexity detection
- Already sourced via session-utils.sh in both hooks

**Implementation**:
```bash
# In hooks needing complexity
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || { exit 0; }
COMPLEXITY=$(get_complexity)
```

---

### Decision 4: Backward Compatibility for Missing Complexity

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Default to SCRIPT | Missing = lightest enforcement | Non-breaking, permissive | May under-enforce |
| B. Default to MODULE | Missing = middle ground | Balanced | Unexpected prompts for legacy |
| C. Default to warn-only | Missing = current behavior exactly | Truly backward compatible | Doesn't improve enforcement |

**Selected**: Option C - Default to warn-only (current behavior)

**Rationale**: Sessions created before this feature should behave exactly as they do now. Users who want stronger enforcement should create new sessions with explicit complexity levels. This ensures no surprise blocks for existing workflows.

**Implementation**:
```bash
COMPLEXITY=$(get_complexity)
if [[ -z "$COMPLEXITY" ]]; then
    # No complexity = legacy session, use warn-only (current behavior)
    ENFORCEMENT_TIER="warn"
fi
```

---

### Decision 5: Audit Event Schema Extension

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Add complexity field | Add to existing details JSON | Minimal change | Loses override info |
| B. New event types | COMPLEXITY_BLOCKED, OVERRIDE_USED | Explicit events | More event types |
| C. Extended details | Add complexity, override, outcome to details | All info in one place | Larger event |

**Selected**: Option C - Extended details with all relevant context

**Rationale**: A single event with complete context is easier to analyze than correlating multiple events. The additional fields are small and valuable for post-session analysis.

**Event Schema Extension**:

```json
{
  "timestamp": "2026-01-02T12:00:00Z",
  "event": "DELEGATION_WARNING",
  "hook": "delegation-check.sh",
  "details": {
    "tool": "Edit",
    "file_path": "/path/to/file.ts",
    "mode": "orchestrated",
    "complexity": "SERVICE",
    "enforcement_tier": "block",
    "override_active": true,
    "override_reason": "Emergency hotfix"
  },
  "outcome": "CONTINUED_WITH_OVERRIDE"
}
```

**Outcome Values**:
- `CONTINUED` - Warn tier, operation proceeded
- `BLOCKED` - Block tier, no override, operation blocked
- `CONTINUED_WITH_OVERRIDE` - Block tier, override active, operation proceeded
- `ACKNOWLEDGED` - Acknowledge tier, user acknowledged, operation proceeded

---

## Technical Design

### Complexity Detection Mechanism

Complexity is read from SESSION_CONTEXT.md frontmatter via existing infrastructure:

```bash
# File: .claude/hooks/lib/session-state.sh (existing, lines 127-131)
get_complexity() {
    get_session_field "complexity" "$@"
}
```

**Complexity Level Ordering**:
```bash
# Numeric mapping for comparison
declare -A COMPLEXITY_ORDER=(
    ["SCRIPT"]=0
    ["PATCH"]=0      # Alias for SCRIPT
    ["MODULE"]=1
    ["SERVICE"]=2
    ["PLATFORM"]=3
)
```

### Gating Logic Decision Tree

```
Start: Hook receives tool invocation
  |
  +-- Is there an active session?
  |     |
  |     +-- No --> Allow (native mode, no enforcement)
  |     |
  |     +-- Yes --> Continue
  |
  +-- Is execution_mode == "orchestrated"?
  |     |
  |     +-- No --> Allow (cross-cutting mode, no enforcement)
  |     |
  |     +-- Yes --> Continue
  |
  +-- Get complexity level
  |     |
  |     +-- Empty/Unknown --> ENFORCEMENT_TIER = "warn" (backward compat)
  |     +-- SCRIPT/PATCH --> ENFORCEMENT_TIER = "warn"
  |     +-- MODULE --> ENFORCEMENT_TIER = "acknowledge"
  |     +-- SERVICE/PLATFORM --> ENFORCEMENT_TIER = "block"
  |
  +-- Check override status
  |     |
  |     +-- Session override active? --> OVERRIDE = true
  |     +-- Environment CLAUDE_BYPASS_ORCHESTRATOR=1? --> OVERRIDE = true (log warning)
  |     +-- Otherwise --> OVERRIDE = false
  |
  +-- Apply enforcement tier
        |
        +-- TIER = "warn"
        |     |
        |     +-- Emit warning to stderr
        |     +-- Log DELEGATION_WARNING with outcome=CONTINUED
        |     +-- Allow operation
        |
        +-- TIER = "acknowledge"
        |     |
        |     +-- Emit warning with acknowledgment prompt to stderr
        |     +-- Log with outcome=CONTINUED (hook cannot block for acknowledgment)
        |     +-- Allow operation (acknowledgment is advisory)
        |
        +-- TIER = "block"
              |
              +-- OVERRIDE = true?
              |     |
              |     +-- Emit notice about override usage
              |     +-- Log with outcome=CONTINUED_WITH_OVERRIDE
              |     +-- Allow operation
              |
              +-- OVERRIDE = false?
                    |
                    +-- Emit block message with override instructions
                    +-- Log with outcome=BLOCKED
                    +-- Return non-zero (block operation)
```

### Override Interface Specification

**Session-Level Override** (via state-mate):

```yaml
# SESSION_CONTEXT.md addition
orchestrator_override:
  enabled: boolean           # true/false
  reason: string             # Required when enabled
  scope: enum                # "next" | "session"
  set_at: iso8601_timestamp  # When override was set
  set_by: string             # "user" or agent name
```

**state-mate Commands**:

| Command | Effect |
|---------|--------|
| `enable orchestrator override reason='...' scope=next` | Enable for next blocked operation only |
| `enable orchestrator override reason='...' scope=session` | Enable for rest of session |
| `disable orchestrator override` | Remove override |

**Environment Override** (emergency):

```bash
# In terminal before starting Claude
export CLAUDE_BYPASS_ORCHESTRATOR=1

# Or for single command
CLAUDE_BYPASS_ORCHESTRATOR=1 claude ...
```

**Override Detection in Hook**:

```bash
check_override_active() {
    # Check session-level override
    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    if [[ -n "$session_dir" ]] && [[ -f "$session_dir/SESSION_CONTEXT.md" ]]; then
        local override_enabled
        override_enabled=$(get_session_field "orchestrator_override.enabled" 2>/dev/null || echo "")
        if [[ "$override_enabled" == "true" ]]; then
            # Check scope
            local scope
            scope=$(get_session_field "orchestrator_override.scope" 2>/dev/null || echo "session")
            OVERRIDE_REASON=$(get_session_field "orchestrator_override.reason" 2>/dev/null || echo "")
            OVERRIDE_SOURCE="session"

            # If scope is "next", clear it after this check
            if [[ "$scope" == "next" ]]; then
                # Note: Actual clearing happens after operation completes
                OVERRIDE_SCOPE="next"
            else
                OVERRIDE_SCOPE="session"
            fi
            return 0
        fi
    fi

    # Check environment override (emergency fallback)
    if [[ "${CLAUDE_BYPASS_ORCHESTRATOR:-}" == "1" ]]; then
        OVERRIDE_REASON="Environment variable CLAUDE_BYPASS_ORCHESTRATOR=1"
        OVERRIDE_SOURCE="environment"
        OVERRIDE_SCOPE="terminal"
        log_warn "Environment override detected - prefer session-level override" >&2
        return 0
    fi

    return 1
}
```

### Audit Event Schema Changes

**Extended Event Type Definitions**:

```json
{
  "$defs": {
    "orchestration_event": {
      "type": "object",
      "required": ["timestamp", "event", "hook", "details", "outcome"],
      "properties": {
        "timestamp": { "$ref": "common.schema.json#/$defs/iso8601_timestamp" },
        "event": {
          "type": "string",
          "enum": [
            "DELEGATION_WARNING",
            "BYPASS_WARNING",
            "ORCHESTRATOR_CONSULTED",
            "MODE_TRANSITION",
            "OVERRIDE_ACTIVATED",
            "OVERRIDE_DEACTIVATED",
            "OVERRIDE_CONSUMED"
          ]
        },
        "hook": { "type": "string" },
        "details": {
          "type": "object",
          "properties": {
            "tool": { "type": "string" },
            "file_path": { "type": "string" },
            "specialist": { "type": "string" },
            "mode": { "type": "string" },
            "complexity": {
              "type": "string",
              "enum": ["SCRIPT", "PATCH", "MODULE", "SERVICE", "PLATFORM", "unknown"]
            },
            "enforcement_tier": {
              "type": "string",
              "enum": ["warn", "acknowledge", "block"]
            },
            "override_active": { "type": "boolean" },
            "override_reason": { "type": "string" },
            "override_source": {
              "type": "string",
              "enum": ["session", "environment"]
            }
          }
        },
        "outcome": {
          "type": "string",
          "enum": ["CONTINUED", "BLOCKED", "CONTINUED_WITH_OVERRIDE", "ACKNOWLEDGED"]
        }
      }
    }
  }
}
```

---

## Implementation Plan

### Hook Modification List

#### 1. `.claude/hooks/validation/delegation-check.sh`

**Changes Required**:

| Section | Current (Lines) | Change |
|---------|-----------------|--------|
| Library sourcing | 16 | Add orchestration-audit.sh source |
| Mode detection | 44-48 | Already uses workflow.active (update to execution_mode per prior design) |
| **NEW** Complexity detection | - | Add after mode detection |
| **NEW** Override check | - | Add check_override_active() call |
| **NEW** Enforcement logic | 67-72 | Replace single warning with tiered logic |
| **NEW** Audit logging | - | Add log_delegation_warning() with extended details |

**New Function**: `get_enforcement_tier()`
```bash
get_enforcement_tier() {
    local complexity="$1"
    case "$complexity" in
        SCRIPT|PATCH|"") echo "warn" ;;
        MODULE) echo "acknowledge" ;;
        SERVICE|PLATFORM) echo "block" ;;
        *) echo "warn" ;;  # Unknown defaults to warn
    esac
}
```

#### 2. `.claude/hooks/validation/orchestrator-bypass-check.sh`

**Changes Required**:

| Section | Current (Lines) | Change |
|---------|-----------------|--------|
| Library sourcing | 17-18 | Add orchestration-audit.sh source |
| has_active_workflow | 43-57 | Keep but add complexity detection |
| **NEW** Complexity detection | - | Add after workflow check |
| **NEW** Override check | - | Add check_override_active() call |
| **NEW** Enforcement logic | 99-116 | Replace single warning with tiered logic |
| **NEW** Audit logging | - | Add log_bypass_warning() with extended details |

#### 3. `.claude/hooks/lib/orchestration-audit.sh`

**Changes Required**:

| Function | Current | Change |
|----------|---------|--------|
| log_orchestration_event | Lines 26-46 | No change (base function) |
| log_delegation_warning | Lines 48-61 | Add complexity, tier, override params |
| log_bypass_warning | Lines 63-72 | Add complexity, tier, override params |
| **NEW** log_override_activated | - | New event for override enable |
| **NEW** log_override_consumed | - | New event for next-scope override use |

**Updated Function Signatures**:

```bash
# Updated: log_delegation_warning
log_delegation_warning() {
    local tool="$1"
    local file_path="$2"
    local mode="$3"
    local complexity="${4:-unknown}"
    local enforcement_tier="${5:-warn}"
    local override_active="${6:-false}"
    local override_reason="${7:-}"
    local outcome="${8:-CONTINUED}"

    log_orchestration_event "DELEGATION_WARNING" \
        "{\"tool\":\"$tool\",\"file_path\":\"$file_path\",\"mode\":\"$mode\",\"complexity\":\"$complexity\",\"enforcement_tier\":\"$enforcement_tier\",\"override_active\":$override_active,\"override_reason\":\"$override_reason\"}" \
        "$outcome" "delegation-check.sh"
}

# Updated: log_bypass_warning
log_bypass_warning() {
    local specialist="$1"
    local complexity="${2:-unknown}"
    local enforcement_tier="${3:-warn}"
    local override_active="${4:-false}"
    local override_reason="${5:-}"
    local outcome="${6:-CONTINUED}"

    log_orchestration_event "BYPASS_WARNING" \
        "{\"specialist\":\"$specialist\",\"complexity\":\"$complexity\",\"enforcement_tier\":\"$enforcement_tier\",\"override_active\":$override_active,\"override_reason\":\"$override_reason\"}" \
        "$outcome" "orchestrator-bypass-check.sh"
}

# New: log_override_activated
log_override_activated() {
    local reason="$1"
    local scope="$2"
    local source="${3:-session}"

    log_orchestration_event "OVERRIDE_ACTIVATED" \
        "{\"reason\":\"$reason\",\"scope\":\"$scope\",\"source\":\"$source\"}" \
        "SUCCESS" "state-mate"
}

# New: log_override_consumed
log_override_consumed() {
    local reason="$1"
    local hook="$2"

    log_orchestration_event "OVERRIDE_CONSUMED" \
        "{\"reason\":\"$reason\",\"consumed_by\":\"$hook\"}" \
        "SUCCESS" "$hook"
}
```

#### 4. `.claude/hooks/lib/session-utils.sh` (via session-state.sh)

**Changes Required**: None - `get_complexity()` already exists at line 127-131

#### 5. Documentation Updates

| File | Change |
|------|--------|
| `.claude/skills/orchestration/execution-mode.md` | Add Complexity Gating section |
| `user-agents/state-mate.md` | Add orchestrator override commands (syncs to ~/.claude/agents/) |
| `user-skills/session-common/session-context-schema.md` | Add orchestrator_override field |

### New Functions/Exports

| Function | File | Purpose |
|----------|------|---------|
| `get_enforcement_tier()` | delegation-check.sh | Map complexity to tier |
| `check_override_active()` | Shared (orchestration-audit.sh) | Check session/env override |
| `apply_enforcement()` | Shared or inline | Apply tier-appropriate behavior |
| `log_override_activated()` | orchestration-audit.sh | Audit override enable |
| `log_override_consumed()` | orchestration-audit.sh | Audit override use |

---

## Test Plan

### Test Scenarios for Each Complexity Level

#### SCRIPT/PATCH Complexity (Warn Tier)

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `ce_001` | Edit file in SCRIPT session | Warning to stderr, operation proceeds |
| `ce_002` | Task(specialist) without consultation in SCRIPT | Warning to stderr, operation proceeds |
| `ce_003` | SCRIPT with override enabled | Same as without (override not needed) |
| `ce_004` | PATCH alias behaves as SCRIPT | Same behavior as SCRIPT |

#### MODULE Complexity (Acknowledge Tier)

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `ce_010` | Edit file in MODULE session | Warning + acknowledgment prompt, operation proceeds |
| `ce_011` | Task(specialist) without consultation in MODULE | Warning + acknowledgment prompt, operation proceeds |
| `ce_012` | Audit log shows ACKNOWLEDGED outcome | Event logged with acknowledgment |

#### SERVICE Complexity (Block Tier)

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `ce_020` | Edit file in SERVICE session, no override | Block message, operation blocked |
| `ce_021` | Task(specialist) without consultation in SERVICE, no override | Block message, operation blocked |
| `ce_022` | Edit in SERVICE with session override | Notice about override, operation proceeds |
| `ce_023` | Edit in SERVICE with env override | Warning about env usage, operation proceeds |
| `ce_024` | Audit log shows BLOCKED for blocked ops | Correct outcome logged |
| `ce_025` | Audit log shows CONTINUED_WITH_OVERRIDE | Correct outcome with override reason |

#### PLATFORM Complexity (Block Tier)

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `ce_030` | Edit file in PLATFORM session, no override | Block message (same as SERVICE) |
| `ce_031` | Override with scope=next, two operations | First proceeds, second blocked |
| `ce_032` | Override with scope=session | All subsequent operations proceed |

### Override Mechanism Tests

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `ce_040` | state-mate enable override | SESSION_CONTEXT updated, OVERRIDE_ACTIVATED logged |
| `ce_041` | state-mate disable override | SESSION_CONTEXT updated, field removed |
| `ce_042` | scope=next consumed after one op | Override cleared after use |
| `ce_043` | CLAUDE_BYPASS_ORCHESTRATOR=1 | Override active, warning about env usage |
| `ce_044` | Both session and env override | Session takes precedence |
| `ce_045` | Invalid override reason (empty) | state-mate rejects |

### Backward Compatibility Tests

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `ce_050` | Session without complexity field | Warn tier (current behavior) |
| `ce_051` | Session with unknown complexity value | Warn tier (graceful fallback) |
| `ce_052` | No session (native mode) | No enforcement, operation proceeds |
| `ce_053` | Parked session (cross-cutting mode) | No enforcement, operation proceeds |
| `ce_054` | Session without orchestrator.md | No bypass warning (already handled) |

### Edge Cases

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `ce_060` | Hook library sourcing fails | Exit 0, no crash (defensive) |
| `ce_061` | get_complexity returns error | Default to warn tier |
| `ce_062` | SESSION_CONTEXT.md unreadable | Exit 0, no crash |
| `ce_063` | Concurrent override set/use | Atomic read via session lock |

---

## Backward Compatibility

**Classification**: COMPATIBLE with graceful degradation

### Compatibility Matrix

| Scenario | Behavior |
|----------|----------|
| Session without complexity field | Warn tier (current behavior exactly) |
| Session with complexity field | New tiered behavior |
| Hooks without new code | Will not call new functions, works as before |
| Environment variable not set | No override, default behavior |
| state-mate without override commands | Commands fail gracefully, no session change |

### Migration Path

**Phase 1: Deploy New Hook Logic**
1. Update orchestration-audit.sh with new functions
2. Update delegation-check.sh with complexity awareness
3. Update orchestrator-bypass-check.sh with complexity awareness
4. All changes backward compatible (empty complexity = warn)

**Phase 2: Add state-mate Override Commands**
1. Add orchestrator override commands to state-mate
2. Update SESSION_CONTEXT schema with new field
3. Document override workflow

**Phase 3: Update Documentation**
1. Update execution-mode.md with complexity gating section
2. Add complexity selection guidance to start-ref skill
3. Document override patterns

---

## Schema Changes

### SESSION_CONTEXT.md Extension

```yaml
# New optional field in frontmatter
orchestrator_override:
  enabled: boolean           # Required when block present
  reason: string             # Required when enabled
  scope: enum                # "next" | "session", default "session"
  set_at: iso8601_timestamp  # When override was set
  set_by: string             # Who set it (user/agent)
```

**Validation Rules**:
1. If `orchestrator_override` exists, `enabled`, `reason`, `set_at` required
2. `scope` defaults to "session" if not specified
3. `reason` must be non-empty when `enabled: true`

### JSON Schema Update

Add to `schemas/artifacts/session-context.schema.json`:

```json
{
  "orchestrator_override": {
    "type": "object",
    "properties": {
      "enabled": { "type": "boolean" },
      "reason": { "type": "string", "minLength": 1 },
      "scope": {
        "type": "string",
        "enum": ["next", "session"],
        "default": "session"
      },
      "set_at": { "$ref": "common.schema.json#/$defs/iso8601_timestamp" },
      "set_by": { "type": "string" }
    },
    "required": ["enabled", "reason", "set_at"],
    "if": {
      "properties": { "enabled": { "const": true } }
    },
    "then": {
      "required": ["enabled", "reason", "set_at", "set_by"]
    }
  }
}
```

---

## Work Packages

### WP1: Orchestration Audit Extension

**Objective**: Extend audit logging with complexity and override information

**Files Changed**:

| File | Action | Description |
|------|--------|-------------|
| `.claude/hooks/lib/orchestration-audit.sh` | modify | Add complexity/override params to existing functions, add new functions |

**Estimated Effort**: 1 hour

### WP2: Delegation Check Enhancement

**Objective**: Add complexity-aware gating to delegation-check.sh

**Files Changed**:

| File | Action | Description |
|------|--------|-------------|
| `.claude/hooks/validation/delegation-check.sh` | modify | Add complexity detection, tiered enforcement |

**Dependencies**: WP1

**Estimated Effort**: 2 hours

### WP3: Bypass Check Enhancement

**Objective**: Add complexity-aware gating to orchestrator-bypass-check.sh

**Files Changed**:

| File | Action | Description |
|------|--------|-------------|
| `.claude/hooks/validation/orchestrator-bypass-check.sh` | modify | Add complexity detection, tiered enforcement |

**Dependencies**: WP1

**Estimated Effort**: 2 hours

### WP4: Override Command Support

**Objective**: Add orchestrator override commands to state-mate

**Files Changed**:

| File | Action | Description |
|------|--------|-------------|
| `user-agents/state-mate.md` | modify | Add override enable/disable commands (syncs to ~/.claude/agents/) |
| `schemas/artifacts/session-context.schema.json` | modify | Add orchestrator_override field |
| `user-skills/session-common/session-context-schema.md` | modify | Document new field |

**Dependencies**: WP1

**Estimated Effort**: 2 hours

### WP5: Documentation

**Objective**: Document complexity gating and override mechanism

**Files Changed**:

| File | Action | Description |
|------|--------|-------------|
| `.claude/skills/orchestration/execution-mode.md` | modify | Add Complexity Gating section |
| `user-skills/session-lifecycle/start-ref/SKILL.md` | modify | Add complexity selection guidance |

**Dependencies**: WP2, WP3, WP4

**Estimated Effort**: 1 hour

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Override abused, defeats purpose | Medium | Medium | Audit all overrides, post-session review |
| Blocking breaks legitimate workflows | Low | High | Conservative defaults (missing = warn), easy override |
| Complexity detection fails | Low | Medium | Default to warn tier on any error |
| State-mate race conditions | Low | Low | Use session locks for override mutations |
| User confusion about tiers | Medium | Low | Clear documentation, consistent messaging |

---

## Success Criteria

- [ ] Hooks read complexity from SESSION_CONTEXT.md
- [ ] SCRIPT/PATCH sessions: warn only (current behavior)
- [ ] MODULE sessions: warn with acknowledgment prompt
- [ ] SERVICE/PLATFORM sessions: block by default
- [ ] Session-level override via state-mate works
- [ ] Environment override works as fallback
- [ ] All bypass events include complexity in audit log
- [ ] Override usage logged with reason
- [ ] Sessions without complexity field behave as before
- [ ] All tests in test matrix pass
- [ ] Documentation updated

---

## Handoff Criteria

- [x] Solution architecture documented with rationale
- [x] Complexity detection mechanism specified
- [x] Gating logic per complexity level defined (decision tree)
- [x] Override interface specification complete
- [x] Audit event schema changes documented
- [x] Hook modification list with file/function level detail
- [x] Test scenarios for each complexity level
- [x] Override mechanism tests specified
- [x] Backward compatibility tests specified
- [x] No unresolved design decisions

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-orchestrator-enforcement.md` | Created |
| delegation-check.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/delegation-check.sh` | Read |
| orchestrator-bypass-check.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-bypass-check.sh` | Read |
| orchestration-audit.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/orchestration-audit.sh` | Read |
| session-state.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-state.sh` | Read |
| session-context.schema.json | `/Users/tomtenuta/Code/roster/schemas/artifacts/session-context.schema.json` | Read |
| CONTEXT-DESIGN-orchestration-mode-consolidation.md | `/Users/tomtenuta/Code/roster/docs/ecosystem/CONTEXT-DESIGN-orchestration-mode-consolidation.md` | Read |
| TDD-hooks-init.md | `/Users/tomtenuta/Code/roster/docs/design/TDD-hooks-init.md` | Read (reference) |

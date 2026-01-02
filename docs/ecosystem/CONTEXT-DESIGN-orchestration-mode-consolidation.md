---
title: "Context Design: Orchestration Mode Consolidation"
type: context-design
complexity: MODULE
created_at: "2026-01-02T12:00:00Z"
status: ready-for-implementation
gap_analysis: spike-coach-mode-analysis.md
affected_systems:
  - skeleton
  - roster
author: context-architect
backward_compatible: false
migration_required: true
work_packages:
  - id: WP1
    name: "Retire Coach Mode Terminology"
    description: "Replace all Coach Mode references with Orchestrated Mode"
    files:
      - path: ".claude/hooks/context-injection/coach-mode.sh"
        action: modify
        description: "Rename to orchestrated-mode.sh, update output messaging"
      - path: ".claude/skills/orchestration/main-thread-guide.md"
        action: modify
        description: "Replace Coach terminology with Orchestrated"
      - path: ".claude/skills/orchestration/execution-mode.md"
        action: modify
        description: "Consolidate mode terminology"
      - path: ".claude/hooks/base_hooks.yaml"
        action: modify
        description: "Update hook registration reference"
  - id: WP2
    name: "Unify Detection Mechanism"
    description: "Consolidate on execution_mode() as single source of truth, deprecate workflow.active"
    files:
      - path: ".claude/hooks/context-injection/orchestrated-mode.sh"
        action: modify
        description: "Use execution_mode() instead of workflow.active field"
      - path: ".claude/hooks/validation/delegation-check.sh"
        action: modify
        description: "Align detection with orchestrated-mode.sh"
      - path: ".claude/hooks/lib/session-manager.sh"
        action: modify
        description: "Add is_orchestrated() convenience function"
    dependencies: [WP1]
  - id: WP3
    name: "Implement Enforcement Behavior"
    description: "Define warn/block/guide pattern for orchestration violations"
    files:
      - path: ".claude/hooks/validation/orchestrator-router.sh"
        action: modify
        description: "Inject consultation template (already exists, verify behavior)"
      - path: ".claude/hooks/validation/orchestrator-bypass-check.sh"
        action: modify
        description: "Add audit logging, clarify guidance message"
      - path: ".claude/hooks/validation/delegation-check.sh"
        action: modify
        description: "Add audit logging, document intentional override"
    dependencies: [WP2]
  - id: WP4
    name: "Add Audit Trail"
    description: "Log all orchestration violations for post-session analysis"
    files:
      - path: ".claude/hooks/lib/orchestration-audit.sh"
        action: create
        description: "Create centralized audit logging for orchestration events"
      - path: ".claude/hooks/validation/delegation-check.sh"
        action: modify
        description: "Integrate audit logging"
      - path: ".claude/hooks/validation/orchestrator-bypass-check.sh"
        action: modify
        description: "Integrate audit logging"
    dependencies: [WP3]
  - id: WP5
    name: "Document Override Mechanism"
    description: "Formalize intentional override procedure"
    files:
      - path: ".claude/skills/orchestration/execution-mode.md"
        action: modify
        description: "Add Override section with documented procedure"
      - path: ".claude/skills/orchestration/entry-pattern.md"
        action: modify
        description: "Add Anti-patterns section with override guidance"
    dependencies: [WP3]
  - id: WP6
    name: "Handle Mode Transitions"
    description: "Define behavior for mid-session team activation/removal"
    files:
      - path: ".claude/hooks/lib/session-manager.sh"
        action: modify
        description: "Add mode transition event logging"
      - path: ".claude/skills/orchestration/execution-mode.md"
        action: modify
        description: "Document mode transition behavior"
    dependencies: [WP2]
schema_version: "1.0"
---

## Executive Summary

This Context Design addresses the fragmented orchestration mode implementation identified in `spike-coach-mode-analysis.md`. The core issues are: (1) inconsistent terminology ("Coach Mode" vs "Orchestrated Mode"), (2) disagreement between `workflow.active` field and `execution_mode()` function, (3) missing or incomplete hook implementations, and (4) undocumented override procedures. The solution unifies on "Orchestrated Mode" terminology, consolidates detection on `execution_mode()` as single source of truth, and establishes a clear enforcement pattern with audit trail.

## Design Decisions

### Decision 1: Terminology Unification

**Options Considered**:

1. **Keep "Coach Mode"** - Familiar metaphor, already in use
   - Rejected: Creates fourth term alongside native/cross-cutting/orchestrated
   - Rejected: Not used in canonical PRD-hybrid-session-model
   - Rejected: Metaphor explains behavior but obscures execution mode relationship

2. **Unify on "Orchestrated Mode"** - Canonical term from three-mode model
   - Selected: Aligns with PRD-hybrid-session-model terminology
   - Selected: Maps directly to execution_mode() return value
   - Selected: Clearer relationship between mode and behavior

3. **Introduce new term "Delegation Mode"**
   - Rejected: Adds fifth term, increases confusion
   - Rejected: Not established in any existing documentation

**Selected**: Unify on "Orchestrated Mode"

**Rationale**: PRD-hybrid-session-model establishes the canonical three-mode model: native, cross-cutting, orchestrated. The "Coach Mode" term is a behavioral description that should be documentation/guidance, not a mode name. By unifying terminology, hooks and skills can use consistent language that maps directly to `execution_mode()` return values.

**Migration Impact**:
- `coach-mode.sh` renamed to `orchestrated-mode.sh`
- `main-thread-guide.md` section "You Are the Coach" becomes "Orchestrated Mode Behavior"
- Hook output changes from "COACH MODE ACTIVE" to "ORCHESTRATED MODE"

---

### Decision 2: Single Detection Mechanism

**Options Considered**:

1. **Keep both `workflow.active` and `execution_mode()`** - Maintain backward compatibility
   - Rejected: Root cause of current disagreement
   - Rejected: Two sources of truth for same concept
   - Rejected: `workflow.active` is not reliably set during session creation

2. **Consolidate on `execution_mode()` only** - Single source of truth
   - Selected: Already handles all edge cases per PRD-hybrid-session-model
   - Selected: Considers session state, team configuration, and pack existence
   - Selected: Graceful fallback to cross-cutting on errors

3. **Consolidate on `workflow.active` only** - Simpler field-based check
   - Rejected: Requires ensuring field is set correctly at all state transitions
   - Rejected: Does not consider team pack existence
   - Rejected: Cannot detect PARKED session correctly

**Selected**: Consolidate on `execution_mode()`

**Rationale**: `execution_mode()` in session-manager.sh already implements the complete decision tree from PRD-hybrid-session-model. It checks session existence, status (ACTIVE/PARKED/ARCHIVED), team configuration, and team pack existence. The `workflow.active` field is redundant and prone to staleness.

**Implementation**:

```bash
# In orchestrated-mode.sh (renamed from coach-mode.sh)
# BEFORE:
WORKFLOW_ACTIVE=$(grep -A5 "^workflow:" "$SESSION_CTX" | grep "active:" | grep -o "true\|false")
if [[ "$WORKFLOW_ACTIVE" == "true" ]]; then ...

# AFTER:
source "$HOOKS_LIB/session-manager.sh"
MODE=$(execution_mode)
if [[ "$MODE" == "orchestrated" ]]; then ...
```

**Deprecation**: The `workflow.active` field in SESSION_CONTEXT.md is deprecated but not removed for backward compatibility. Hooks no longer read it. Future sessions may omit it.

---

### Decision 3: Enforcement Behavior Pattern

**Options Considered**:

1. **Block** - Hard enforcement, prevent unauthorized operations
   - Rejected: Removes human override capability
   - Rejected: Can cause workflow lockout on edge cases
   - Rejected: Violates principle of graceful degradation

2. **Warn** - Soft enforcement, emit warning and continue
   - Partially Selected: Appropriate for delegation-check (Edit/Write)
   - Partially Selected: Appropriate for orchestrator-bypass-check (Task)
   - Preserves human agency for intentional overrides

3. **Guide** - Inject routing context, do not warn/block
   - Partially Selected: Appropriate for orchestrator-router (UserPromptSubmit)
   - Proactive rather than reactive
   - Helps main agent do the right thing without friction

**Selected**: Hybrid Pattern - Guide + Warn + Never Block

| Hook | Event | Enforcement | Rationale |
|------|-------|-------------|-----------|
| `orchestrator-router.sh` | UserPromptSubmit | **Guide** | Inject CONSULTATION_REQUEST template |
| `orchestrator-bypass-check.sh` | PreToolUse(Task) | **Warn** | Specialist without orchestrator consultation |
| `delegation-check.sh` | PreToolUse(Edit/Write) | **Warn** | Direct implementation in orchestrated mode |

**No hooks block operations**. All enforcement is advisory. Human/agent can proceed if intentional.

**Rationale**: The roster philosophy is "guardrails, not gates". Hooks guide toward correct behavior and log deviations, but never prevent work from proceeding. This supports the principle that the human operator has final authority.

---

### Decision 4: Audit Trail Design

**Options Considered**:

1. **Per-session event log** - Store in `sessions/{id}/orchestration-audit.jsonl`
   - Selected: Scoped to session lifecycle
   - Selected: Easy to analyze per-session
   - Selected: Automatically archived with session

2. **Global audit log** - Store in `.claude/sessions/.audit/orchestration.log`
   - Rejected: Grows unbounded
   - Rejected: Harder to correlate with sessions

3. **No persistent audit** - Just stderr warnings
   - Rejected: Cannot do post-session analysis
   - Rejected: Warnings may be swallowed in context

**Selected**: Per-session event log at `sessions/{id}/orchestration-audit.jsonl`

**Event Schema**:

```json
{
  "timestamp": "2026-01-02T12:00:00Z",
  "event": "DELEGATION_WARNING",
  "hook": "delegation-check.sh",
  "details": {
    "tool": "Edit",
    "file_path": "/path/to/file.ts",
    "mode": "orchestrated"
  },
  "outcome": "CONTINUED"
}
```

**Event Types**:
- `ORCHESTRATOR_CONSULTED` - Main agent consulted orchestrator (success path)
- `DELEGATION_WARNING` - Edit/Write in orchestrated mode
- `BYPASS_WARNING` - Specialist invoked without orchestrator consultation
- `OVERRIDE_SIGNALED` - User explicitly overrode (see Decision 5)

---

### Decision 5: Override Mechanism

**Options Considered**:

1. **Magic comment** - `# @override-orchestration` in prompt
   - Rejected: Pollutes user prompts
   - Rejected: Easy to forget, hard to discover

2. **Environment variable** - `ROSTER_ORCHESTRATION_OVERRIDE=true`
   - Rejected: Requires terminal restart
   - Rejected: Affects all operations, not targeted

3. **Explicit continuation** - Just proceed after warning
   - Selected: Current implicit behavior made explicit
   - Selected: No ceremony required
   - Selected: Audit trail logs "CONTINUED" outcome

4. **Inline acknowledgment** - Main agent says "proceeding with override"
   - Partially Selected: Encourages explicit acknowledgment
   - Documents reasoning in conversation

**Selected**: Implicit continuation + optional acknowledgment

**Mechanism**:
1. Hook emits warning to stderr (becomes context)
2. Main agent sees warning in context
3. If proceeding anyway, audit logs `outcome: "CONTINUED"`
4. For clear documentation, agent can state: "Intentional override: [reason]"

**Documentation Pattern** (for execution-mode.md):

```markdown
## Intentional Override

When you see a delegation warning and choose to proceed:

1. The operation is **not blocked** - you may continue
2. State your reasoning: "Intentional override: [reason]"
3. The override is logged for post-session analysis

Valid reasons for override:
- Emergency hotfix requiring immediate action
- Cross-team work touching multiple domains
- Documentation or artifact files (not implementation code)
```

---

### Decision 6: Mode Transition Handling

**Options Considered**:

1. **Immediate effect** - Mode changes instantly on `/team` or `/park`
   - Selected: Consistent with current behavior
   - Selected: `execution_mode()` reads current state, not cached

2. **Deferred effect** - Mode changes at next SessionStart
   - Rejected: Confusing gap between command and behavior
   - Rejected: Would require manual session restart

3. **Confirmation required** - "Switching modes, are you sure?"
   - Rejected: Adds friction to intentional transitions
   - Rejected: `/team` and `/park` are explicit commands

**Selected**: Immediate effect with logged transition

**Transition Matrix**:

| Action | From Mode | To Mode | Log Event |
|--------|-----------|---------|-----------|
| `/team <pack>` | cross-cutting | orchestrated | MODE_TRANSITION |
| `/team <pack>` | native | native (session needed first) | N/A |
| `/team --remove` | orchestrated | cross-cutting | MODE_TRANSITION |
| `/park` | orchestrated | cross-cutting (parked sessions not orchestrated) | MODE_TRANSITION |
| `/resume` | cross-cutting (parked) | orchestrated (if team still configured) | MODE_TRANSITION |

**Implementation**: Add `log_mode_transition()` call in session-manager.sh after state changes.

---

## Work Package Details

### WP1: Retire Coach Mode Terminology

**Objective**: Replace "Coach Mode" with "Orchestrated Mode" across all references

**Files Changed**:

| File | Before | After |
|------|--------|-------|
| `coach-mode.sh` | "COACH MODE ACTIVE" | "ORCHESTRATED MODE" |
| `coach-mode.sh` | "You are the Coach" | "You are the coordinator" |
| `main-thread-guide.md` | "You Are the Coach" section | "Orchestrated Mode Behavior" section |
| `execution-mode.md` | N/A | Consolidate terminology section |
| `base_hooks.yaml` | `path: context-injection/coach-mode.sh` | `path: context-injection/orchestrated-mode.sh` |

**Rename Operation**:
```bash
git mv .claude/hooks/context-injection/coach-mode.sh \
       .claude/hooks/context-injection/orchestrated-mode.sh
```

**Output Message Change** (orchestrated-mode.sh):

```
# BEFORE:
**COACH MODE ACTIVE**
You are the Coach. Delegate all implementation via Task tool.
Do NOT use Edit/Write directly on code files.
See: .claude/skills/orchestration/main-thread-guide.md

# AFTER:
**ORCHESTRATED MODE**
You are the coordinator. Delegate implementation to specialists via Task tool.
Do NOT use Edit/Write directly on implementation files.
See: .claude/skills/orchestration/execution-mode.md
```

---

### WP2: Unify Detection Mechanism

**Objective**: Consolidate on `execution_mode()` as single source of truth

**orchestrated-mode.sh Changes** (lines 36-48):

```bash
# BEFORE:
WORKFLOW_ACTIVE=$(grep -A5 "^workflow:" "$SESSION_CTX" 2>/dev/null | grep "active:" | grep -o "true\|false" | head -1) || WORKFLOW_ACTIVE=""

if [[ "$WORKFLOW_ACTIVE" == "true" ]]; then
    # Output reminder...

# AFTER:
# Source session utilities for execution_mode()
source "$HOOKS_LIB/session-manager.sh" 2>/dev/null || { exit 0; }

MODE=$(execution_mode)

if [[ "$MODE" == "orchestrated" ]]; then
    # Output reminder...
```

**delegation-check.sh Changes** (lines 42-48):

```bash
# BEFORE:
WORKFLOW_ACTIVE=$(grep -A5 "^workflow:" "$SESSION_CTX" 2>/dev/null | grep "active:" | grep -o "true\|false" | head -1) || WORKFLOW_ACTIVE=""

if [[ "$WORKFLOW_ACTIVE" != "true" ]]; then
  hooks_finalize 0
  exit 0
fi

# AFTER:
source "$HOOKS_LIB/session-manager.sh" 2>/dev/null || { hooks_finalize 0; exit 0; }

MODE=$(execution_mode)

if [[ "$MODE" != "orchestrated" ]]; then
  hooks_finalize 0
  exit 0
fi
```

**Convenience Function** (session-manager.sh, after execution_mode):

```bash
# Convenience function for hook conditionals
# Returns: 0 if orchestrated, 1 otherwise
is_orchestrated() {
    [[ "$(execution_mode)" == "orchestrated" ]]
}
```

---

### WP3: Implement Enforcement Behavior

**Objective**: Standardize warn/guide pattern across enforcement hooks

**orchestrator-router.sh** (already implemented, verify):
- Injects CONSULTATION_REQUEST template on `/start`, `/sprint`, `/task`
- Does NOT block, only adds context
- Skips if no orchestrator.md present

**orchestrator-bypass-check.sh** (modify warning message):

```bash
# BEFORE (line 99-116):
cat >&2 <<EOF
## Warning: Orchestrator Consultation Recommended
...
*This is a warning only - proceeding with specialist invocation.*
EOF

# AFTER:
cat >&2 <<EOF
## Orchestration Guidance

You are invoking specialist **$AGENT** without recent orchestrator consultation.

**During orchestrated workflows**, best practice is:
1. Consult orchestrator: \`Task(orchestrator, "CONSULTATION_REQUEST...")\`
2. Parse the directive returned
3. Invoke specialist per directive

**If this is intentional**, you may proceed. State "Intentional override: [reason]" for clarity.

See: .claude/skills/orchestration/execution-mode.md#intentional-override
EOF
```

**delegation-check.sh** (modify warning message):

```bash
# BEFORE (line 68-72):
cat >&2 <<EOF
[DELEGATION] Workflow active ($WORKFLOW_NAME): $TOOL_NAME on $FILE_PATH
  -> Use Task tool to delegate, or proceed if intentional override.
  -> See: .claude/skills/orchestration/main-thread-guide.md
EOF

# AFTER:
cat >&2 <<EOF
## Delegation Guidance

Mode: **orchestrated**
Tool: $TOOL_NAME
File: $FILE_PATH

In orchestrated mode, implementation should be delegated to specialists via Task tool.

**If this is intentional** (e.g., artifact management, emergency), you may proceed.
State "Intentional override: [reason]" for clarity.

See: .claude/skills/orchestration/execution-mode.md#intentional-override
EOF
```

---

### WP4: Add Audit Trail

**Objective**: Create centralized audit logging for orchestration events

**New File**: `.claude/hooks/lib/orchestration-audit.sh`

```bash
#!/bin/bash
# orchestration-audit.sh - Centralized orchestration event logging
#
# Usage: log_orchestration_event <event_type> <details_json>
# Events are logged to session's orchestration-audit.jsonl

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"

# Source session utilities
source "$SCRIPT_DIR/session-utils.sh" 2>/dev/null || return 0

log_orchestration_event() {
    local event_type="$1"
    local details_json="$2"
    local outcome="${3:-CONTINUED}"
    local hook_name="${4:-unknown}"

    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    [[ -z "$session_dir" ]] && return 0

    local audit_file="$session_dir/orchestration-audit.jsonl"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Append event to audit log (create if needed)
    cat >> "$audit_file" <<EOF
{"timestamp":"$timestamp","event":"$event_type","hook":"$hook_name","details":$details_json,"outcome":"$outcome"}
EOF
}

# Convenience functions for common events
log_delegation_warning() {
    local tool="$1"
    local file_path="$2"
    local mode="$3"
    log_orchestration_event "DELEGATION_WARNING" \
        "{\"tool\":\"$tool\",\"file_path\":\"$file_path\",\"mode\":\"$mode\"}" \
        "CONTINUED" "delegation-check.sh"
}

log_bypass_warning() {
    local specialist="$1"
    log_orchestration_event "BYPASS_WARNING" \
        "{\"specialist\":\"$specialist\"}" \
        "CONTINUED" "orchestrator-bypass-check.sh"
}

log_orchestrator_consulted() {
    local request_type="$1"
    log_orchestration_event "ORCHESTRATOR_CONSULTED" \
        "{\"request_type\":\"$request_type\"}" \
        "SUCCESS" "main-thread"
}
```

**Integration in delegation-check.sh** (after warning):

```bash
# After emitting warning, log to audit
source "$HOOKS_LIB/orchestration-audit.sh" 2>/dev/null || true
log_delegation_warning "$TOOL_NAME" "$FILE_PATH" "$MODE"
```

**Integration in orchestrator-bypass-check.sh** (after warning):

```bash
# After emitting warning, log to audit
source "$HOOKS_LIB/orchestration-audit.sh" 2>/dev/null || true
log_bypass_warning "$AGENT"
```

---

### WP5: Document Override Mechanism

**Objective**: Formalize intentional override procedure in documentation

**execution-mode.md Addition** (new section after "Mode-Aware Hooks"):

```markdown
## Intentional Override

When enforcement hooks warn about orchestration violations, you may proceed intentionally.

### When to Override

Valid reasons for override:
- **Emergency hotfix** - Production issue requiring immediate action
- **Artifact management** - Updating documentation, configs, or non-code files
- **Cross-domain work** - Task spanning team boundaries
- **Debugging** - Investigating issue requires direct access

### How to Override

1. **Acknowledge the warning** - State: "Intentional override: [your reason]"
2. **Proceed with operation** - The warning does not block
3. **Review audit trail** - Post-session analysis available

### Audit Trail

All warnings are logged to `sessions/{id}/orchestration-audit.jsonl`:
- `DELEGATION_WARNING` - Edit/Write in orchestrated mode
- `BYPASS_WARNING` - Specialist invoked without orchestrator

Review with: `cat .claude/sessions/{session-id}/orchestration-audit.jsonl | jq .`

### Important

Overrides are **logged, not blocked**. The roster philosophy is guardrails, not gates.
If you find yourself overriding frequently, consider switching to cross-cutting mode:
`/team --remove`
```

**entry-pattern.md Addition** (new section):

```markdown
## Anti-Patterns and Overrides

### DO NOT: Bypass orchestrator habitually

If you find yourself bypassing orchestrator consultation frequently, the workflow may not be appropriate. Consider:
- Switching to cross-cutting mode (`/team --remove`)
- Using native mode for quick tasks

### When Override is Appropriate

Single emergency or edge case: Override and proceed.
Repeated pattern: Reconsider workflow choice.

### Logging

All violations are logged for post-session analysis. Review with:
\`cat .claude/sessions/{session-id}/orchestration-audit.jsonl | jq .\`
```

---

### WP6: Handle Mode Transitions

**Objective**: Define and log mode transitions

**session-manager.sh Addition** (in cmd_mutate section):

```bash
# Log mode transition for audit trail
log_mode_transition() {
    local from_mode="$1"
    local to_mode="$2"
    local trigger="$3"

    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    [[ -z "$session_dir" ]] && return 0

    local audit_file="$session_dir/orchestration-audit.jsonl"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat >> "$audit_file" <<EOF
{"timestamp":"$timestamp","event":"MODE_TRANSITION","details":{"from":"$from_mode","to":"$to_mode","trigger":"$trigger"}}
EOF
}
```

**Integration Points**:

| Action | Where | Call |
|--------|-------|------|
| `/team <pack>` | team-ref skill | `log_mode_transition "cross-cutting" "orchestrated" "team-activation"` |
| `/team --remove` | team-ref skill | `log_mode_transition "orchestrated" "cross-cutting" "team-removal"` |
| `/park` | mutate_park_fsm | `log_mode_transition "orchestrated" "cross-cutting" "session-parked"` |
| `/resume` | mutate_resume_fsm | Check if team exists, log if transitioning to orchestrated |

**execution-mode.md Addition** (mode transition section):

```markdown
## Mode Transitions

Mode changes take effect immediately. No session restart required.

| Action | From | To | Notes |
|--------|------|-----|-------|
| `/team <pack>` | cross-cutting | orchestrated | Requires active session |
| `/team --remove` | orchestrated | cross-cutting | Removes team, keeps session |
| `/park` | orchestrated | cross-cutting | Parked sessions are never orchestrated |
| `/resume` | cross-cutting | orchestrated | If team still configured |
| `/wrap` | any | native | Session ends |

### Transition Audit

All mode transitions are logged to `orchestration-audit.jsonl` as `MODE_TRANSITION` events.
```

---

## Backward Compatibility

**Classification**: BREAKING (with migration path)

### Breaking Changes

1. **Hook rename**: `coach-mode.sh` -> `orchestrated-mode.sh`
   - Impact: Satellite hooks.yaml referencing old path will fail
   - Migration: Update `base_hooks.yaml` reference

2. **workflow.active deprecated**: No longer read by hooks
   - Impact: None (field was not reliably set)
   - Migration: No action required; field may remain for legacy sessions

3. **Output message changes**: Warning text changes
   - Impact: Any automation parsing stderr for "COACH MODE" will break
   - Migration: Update to match new "ORCHESTRATED MODE" text

### Non-Breaking Changes

1. **New audit file**: `orchestration-audit.jsonl` created per-session
   - Impact: None (additive)

2. **New convenience function**: `is_orchestrated()` added
   - Impact: None (additive)

3. **Updated documentation**: Terminology changes
   - Impact: None (documentation only)

### Migration Path

1. **Rename hook file** before updating base_hooks.yaml
2. **Update base_hooks.yaml** with new path
3. **Update documentation** (main-thread-guide.md, execution-mode.md)
4. **Test with existing sessions** - should continue working

No session data migration required. Existing sessions will work because:
- `execution_mode()` reads ACTIVE_TEAM and session status, not workflow.active
- New hooks are backward compatible with v2 session schema

---

## Test Matrix

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `omc_001` | Hook loads without workflow.active field | Uses execution_mode(), outputs correctly |
| `omc_002` | orchestrated-mode.sh fires in orchestrated mode | Outputs "ORCHESTRATED MODE" message |
| `omc_003` | orchestrated-mode.sh silent in cross-cutting mode | No output |
| `omc_004` | delegation-check uses execution_mode() | Warns only in orchestrated mode |
| `omc_005` | orchestrator-bypass-check logs to audit | Event in orchestration-audit.jsonl |
| `omc_006` | delegation-check logs to audit | Event in orchestration-audit.jsonl |
| `omc_007` | Mode transition logged on /team activation | MODE_TRANSITION event in audit |
| `omc_008` | Mode transition logged on /park | MODE_TRANSITION event in audit |
| `omc_009` | Override proceeds without block | Operation completes, outcome: CONTINUED |
| `omc_010` | is_orchestrated() returns correctly | 0 in orchestrated, 1 otherwise |

### Satellite Diversity Coverage

| Satellite Type | Test IDs | Notes |
|----------------|----------|-------|
| Minimal (no session) | omc_003 | Silent in native mode |
| Standard (10x-dev-pack) | omc_002, omc_004, omc_005, omc_009 | Full enforcement |
| Complex (ecosystem-pack) | All | Full coverage |
| Legacy (v1 session) | omc_001 | Backward compatibility |

---

## Handoff Criteria

- [x] All design decisions have documented rationale
- [x] No TBD, TODO, or unresolved items
- [x] Work packages specify file-level changes
- [x] Backward compatibility assessed (BREAKING with migration)
- [x] Migration path documented
- [x] Test matrix defined with satellite diversity
- [x] Detection mechanism unified on execution_mode()
- [x] Enforcement behavior specified (warn/guide/never block)
- [x] Audit trail design complete
- [x] Override mechanism documented
- [x] Mode transition handling specified

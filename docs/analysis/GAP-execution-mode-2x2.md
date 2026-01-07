# Gap Analysis: Execution Mode 2x2 Simplification

> Analysis of replacing 3-mode enum (native|cross-cutting|orchestrated) with orthogonal 2x2 matrix.

**Analyst**: Ecosystem Analyst
**Date**: 2026-01-02
**Complexity**: MODULE
**Status**: Analysis Complete

---

## 1. Current State Analysis (3-Mode Model)

### 1.1 Authoritative Detection Logic

**File**: `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh`
**Function**: `execution_mode()` (lines 48-114)

The current detection returns a **3-value enum**: `native | orchestrated | cross-cutting`

```
Detection Flow:
1. get_session_id() fails -> "native"
2. session_id empty -> "native"
3. SESSION_CONTEXT.md missing -> "native"
4. status == PARKED -> "cross-cutting"
5. status == ARCHIVED -> "native"
6. ACTIVE_RITE == none/null -> "cross-cutting"
7. rite directory missing -> "cross-cutting"
8. All conditions pass -> "orchestrated"
```

### 1.2 Three-Mode Semantics (from execution-mode.md)

| Mode | Session | Team | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no session tracking |
| **Cross-Cutting** | Yes | No (or parked) | Direct execution with session tracking |
| **Orchestrated** | Yes (ACTIVE) | Yes | Coordinator pattern, delegate via Task tool |

### 1.3 Components Using the 3-Mode Enum

| Component | File | Usage |
|-----------|------|-------|
| `delegation-check.sh` | `.claude/hooks/validation/` | Warns/blocks on Edit/Write in orchestrated mode |
| `orchestrated-mode.sh` | `.claude/hooks/context-injection/` | Injects coordinator reminder when mode == "orchestrated" |
| `session-context.sh` | `.claude/hooks/context-injection/` | Displays mode in session context table |
| `orchestration-audit.sh` | `.claude/hooks/lib/` | Logs mode value in DELEGATION_WARNING events |
| `handoff-ref/skill.md` | `.claude/skills/` | Guards handoff with mode check |
| `CLAUDE.md` | `.claude/` | Documents 3-mode table |
| `execution-mode.md` | `.claude/skills/orchestration/` | Full documentation of 3-mode model |

---

## 2. Proposed State (2x2 Model)

### 2.1 Orthogonal Dimensions

| Dimension | Values | Description |
|-----------|--------|-------------|
| **Session Tracking** | `tracked` / `untracked` | Is there an active session? |
| **Team Delegation** | `delegated` / `direct` | Is team orchestration active? |

### 2.2 2x2 Matrix

|                    | **Untracked** | **Tracked** |
|--------------------|---------------|-------------|
| **Direct**         | Native        | Cross-Cutting |
| **Delegated**      | Invalid*      | Orchestrated |

*Team without session is not a valid state (see Section 3.1)

### 2.3 Behavioral Mapping

| Current Mode | Session? | Team? | Proposed (Session, Delegation) |
|--------------|----------|-------|--------------------------------|
| Native | No | - | (untracked, direct) |
| Cross-Cutting | Yes | No | (tracked, direct) |
| Orchestrated | Yes | Yes | (tracked, delegated) |

---

## 3. Semantic Entanglement Points

### 3.1 Is "Team Active + No Session" a Valid State?

**Current Implementation**: No. The `execution_mode()` function checks session existence FIRST (lines 50-59). If no session, it returns `native` immediately without checking team.

**Enforcement**: ACTIVE_RITE file can exist without session, but:
- `orchestrator-router.sh` requires orchestrator.md file (agent presence, not team activation)
- Session creation (`fsm_create_session`) reads ACTIVE_RITE but creates session with whatever team exists
- No validation prevents team activation without session

**Recommendation**: Keep this constraint. Teams provide coordination for session-tracked work. A team without a session is semantically undefined - there's nothing to coordinate.

### 3.2 What is the Semantic Difference Between Cross-Cutting and Native?

| Aspect | Native | Cross-Cutting |
|--------|--------|---------------|
| Session tracking | None | Active |
| Artifact recording | No | Yes |
| Blocker/next_steps | No | Yes |
| `/park`, `/wrap` available | No | Yes |
| `/consult` routing | Works | Works |
| delegation-check.sh | Silent | Silent |
| orchestrated-mode.sh | Silent | Silent |

**Key Insight**: The behavioral difference is **session tracking**, not execution style. Both modes allow direct Edit/Write. The distinction matters for:
- Audit trail continuity
- Artifact persistence
- Session lifecycle commands

**Is this worth distinguishing?** Yes, but the distinction is **session tracking**, not **mode**. This supports the 2x2 refactoring.

### 3.3 Is PARKED Status Orthogonal to Team Presence?

**Current Implementation**: Entangled. Line 76-79 of `execution_mode()`:
```bash
if [[ "$status" == "PARKED" ]]; then
    echo "cross-cutting"
    return 0
fi
```

PARKED sessions force cross-cutting **regardless of team configuration**.

**Analysis**:
- PARKED is a **lifecycle state** (ACTIVE -> PARKED -> ACTIVE/ARCHIVED)
- Team presence is a **structural property** of the session
- Current design treats PARKED as "temporarily de-orchestrated"

**Semantic Question**: Should a PARKED session with a team be orchestrated when resumed?

**Current Answer**: Yes. Resuming (PARKED -> ACTIVE) restores orchestration if team still configured.

**2x2 Implication**: PARKED is orthogonal. It's a session lifecycle state, not a mode dimension. The 2x2 should be:
- Session dimension: `untracked | tracked | parked` (3 values, not 2)
- OR: Keep session as binary, add `parked` as a modifier flag

**Recommendation**: PARKED should be modeled as `(tracked, direct, parked=true)`. When resumed, it becomes `(tracked, delegated)` if team present.

### 3.4 Fallback Behavior Analysis

Current fallback chain (NFR-2: "graceful fallback to cross-cutting"):

```
Session corrupted -> native
Team pack missing -> cross-cutting
Any detection error -> cross-cutting
```

**Semantic Issue**: Why does "rite missing" fall to cross-cutting but "session corrupted" falls to native?

**Rationale**: Session corruption means we can't trust ANY session state, so drop to native. Team pack missing means session is valid but team isn't, so preserve session tracking.

**2x2 Mapping**: This actually clarifies in 2x2 model:
- Session corrupted -> `(untracked, direct)` - no session tracking
- Team pack missing -> `(tracked, direct)` - session tracking, no delegation

---

## 4. Migration Impact Assessment

### 4.1 API Changes

Current API:
```bash
execution_mode() -> "native" | "orchestrated" | "cross-cutting"
```

Proposed API Option A (minimal change):
```bash
execution_mode() -> "native" | "orchestrated" | "cross-cutting"  # unchanged
has_session() -> bool
has_team() -> bool  # NEW
```

Proposed API Option B (2x2 exposed):
```bash
session_status() -> "untracked" | "tracked" | "parked"
delegation_status() -> "direct" | "delegated"
# execution_mode() deprecated or computed from above
```

### 4.2 Component-by-Component Migration

| Component | Current Usage | Migration Effort |
|-----------|---------------|------------------|
| `delegation-check.sh` | Checks workflow.active, not mode | None needed |
| `orchestrated-mode.sh` | `$MODE == "orchestrated"` | Check delegation_status or has_team+ACTIVE |
| `session-context.sh` | Displays `$EXECUTION_MODE` | Display both dimensions |
| `orchestration-audit.sh` | Logs mode string | Log both dimensions |
| `handoff-ref/skill.md` | `$MODE != "orchestrated"` | Check delegation status |
| `CLAUDE.md` | 3-column table | 2x2 table |
| `execution-mode.md` | Full 3-mode docs | Rewrite for 2x2 |

### 4.3 Breaking Changes

**Low risk**: The 3-mode enum is consumed as string comparison. All existing checks would continue working if `execution_mode()` remains as backward-compatible wrapper:

```bash
execution_mode() {
    if ! has_session; then echo "native"; return; fi
    if has_team && status == ACTIVE; then echo "orchestrated"; return; fi
    echo "cross-cutting"
}
```

---

## 5. Root Cause Summary

### What's the actual problem with the 3-mode model?

1. **Conceptual entanglement**: "Cross-cutting" conflates two orthogonal properties:
   - Session without team
   - Session with team but PARKED

2. **PARKED special-casing**: Lines 76-79 force cross-cutting for PARKED, making PARKED appear as a mode rather than a lifecycle state.

3. **Naming confusion**: "Cross-cutting" implies work spanning teams, but the mode actually means "session tracking without delegation."

### Why does the 2x2 model help?

1. **Orthogonality**: Session tracking and team delegation are independent concerns.
2. **Clarity**: PARKED is explicitly a lifecycle modifier, not a mode.
3. **Extensibility**: Adding new dimensions (e.g., complexity-based enforcement) doesn't require new modes.

---

## 6. Recommendation

### Verdict: 2x2 IS Cleaner

The 2x2 model correctly separates orthogonal concerns:
- **Session tracking**: Whether artifacts, blockers, and lifecycle commands are available
- **Team delegation**: Whether the main thread should delegate via Task tool

### Suggested Implementation

1. **Keep `execution_mode()` for backward compatibility** - compute from 2x2 dimensions
2. **Add primitive functions**:
   - `has_session()` - already exists
   - `has_active_rite()` - NEW, checks team AND session ACTIVE
3. **Model PARKED as a lifecycle flag**, not a mode dimension
4. **Update documentation** to present 2x2 conceptually, even if API remains compatible

### Complexity Classification: MODULE

Requires changes to:
- session-manager.sh (add `has_active_rite()`)
- execution-mode.md (full rewrite)
- CLAUDE.md (update table)
- Session context display
- No breaking changes to existing hook behavior

### Test Satellites

For verification, test against:
- **skeleton** (baseline, no session)
- **test-satellite-minimal** (session, no team)
- **test-satellite-complex** (session with team, PARKED state)

---

## 7. Traceability

| Analysis Point | Source |
|----------------|--------|
| 3-mode detection logic | `.claude/hooks/lib/session-manager.sh:48-114` |
| PARKED entanglement | `.claude/hooks/lib/session-manager.sh:76-79` |
| Component usage | Grep results across `.claude/` |
| Original requirements | `docs/requirements/PRD-hybrid-session-model.md` |
| Current documentation | `.claude/skills/orchestration/execution-mode.md` |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This Gap Analysis | `/Users/tomtenuta/Code/roster/docs/analysis/GAP-execution-mode-2x2.md` | Created |

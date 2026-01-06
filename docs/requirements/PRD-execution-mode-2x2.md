# PRD: 2x2 Execution Mode Model

## Overview

This PRD specifies a conceptual and implementation refactoring of execution mode detection from a 3-value enum (`native | cross-cutting | orchestrated`) to an orthogonal 2x2 matrix model. The 2x2 model separates session tracking from team delegation, clarifying that PARKED is a lifecycle modifier rather than a mode dimension. Backward compatibility is preserved through computed mode values.

## Background

### Current State (3-Mode Model)

The roster ecosystem determines execution mode via `execution_mode()` in `session-manager.sh`, which returns one of three string values:

| Mode | Session | Team | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no session tracking |
| **Cross-Cutting** | Yes | No (or parked) | Direct execution with session tracking |
| **Orchestrated** | Yes (ACTIVE) | Yes | Coordinator pattern, delegate via Task tool |

**Detection Logic (current)**:
```
1. get_session_id() fails -> "native"
2. session_id empty -> "native"
3. SESSION_CONTEXT.md missing -> "native"
4. status == PARKED -> "cross-cutting"
5. status == ARCHIVED -> "native"
6. ACTIVE_RITE == none/null -> "cross-cutting"
7. team pack directory missing -> "cross-cutting"
8. All conditions pass -> "orchestrated"
```

### Problem Statement

The current 3-mode model has semantic entanglement issues identified in the Gap Analysis:

1. **"Cross-cutting" conflates two orthogonal properties**: A session without team AND a session with team but PARKED both resolve to "cross-cutting". These are semantically different states with different resumption behaviors.

2. **PARKED appears as a mode rather than a lifecycle state**: Lines 76-79 of `execution_mode()` force cross-cutting for PARKED sessions regardless of team configuration, implying PARKED is a mode dimension when it's actually a lifecycle state (ACTIVE -> PARKED -> ACTIVE/ARCHIVED).

3. **Naming confusion**: "Cross-cutting" implies work spanning teams, but the mode actually means "session tracking without delegation." The name obscures the underlying model.

4. **Hidden constraint**: "Team Active + No Session" is silently invalid. The function returns `native` if no session exists, even if ACTIVE_RITE is set. This constraint is not explicit in the model.

### Gap Analysis Reference

Full analysis available at: `/Users/tomtenuta/Code/roster/docs/analysis/GAP-execution-mode-2x2.md`

## Goals

### Goals

1. **Separate orthogonal concerns**: Session tracking and team delegation are independent dimensions that should be modeled independently.

2. **Explicit lifecycle modeling**: PARKED is a modifier on session state, not a mode dimension. The model should reflect this.

3. **Backward compatibility**: Existing `execution_mode()` consumers continue working unchanged.

4. **Improved clarity**: Developers and agents can reason about mode using primitives (`is_session_tracked()`, `has_active_team()`) rather than memorizing a 3-value enum.

5. **Extensibility**: Future dimensions (complexity-based enforcement, environment-specific rules) can be added without creating new mode enum values.

### Non-Goals

- **Breaking existing API**: The `execution_mode()` function continues to return the 3-value enum for all current consumers.
- **Changing hook behavior**: Hooks continue working with the same inputs; only internal implementation changes.
- **UI/UX changes**: No changes to `/consult`, `/start`, or other user-facing commands.
- **Session schema changes**: No modifications to SESSION_CONTEXT.md schema.

## Functional Requirements

### Must Have

#### FR-1: Primitive Detection Functions

- **FR-1.1**: Add `is_session_tracked()` function to `session-manager.sh` returning boolean (`true`/`false` exit code).
  - Returns `true` when: Session ID exists AND session directory exists AND session status is ACTIVE or PARKED
  - Returns `false` otherwise

- **FR-1.2**: Add `has_active_team()` function to `session-manager.sh` returning boolean (`true`/`false` exit code).
  - Returns `true` when: Session is tracked AND status is ACTIVE AND ACTIVE_RITE is set (not "none" or empty) AND team pack directory exists
  - Returns `false` otherwise
  - Note: PARKED sessions return `false` for `has_active_team()` because delegation is suspended

- **FR-1.3**: Add `is_session_parked()` function to `session-manager.sh` returning boolean (`true`/`false` exit code).
  - Returns `true` when: Session is tracked AND status is PARKED
  - Returns `false` otherwise

#### FR-2: Computed Mode (Backward Compatibility)

- **FR-2.1**: Refactor `execution_mode()` to compute mode from primitives:
  ```bash
  execution_mode() {
      if ! is_session_tracked; then
          echo "native"
      elif has_active_team; then
          echo "orchestrated"
      else
          echo "cross-cutting"
      fi
  }
  ```

- **FR-2.2**: All existing `execution_mode()` callers continue working without modification. The function signature and return values remain unchanged.

- **FR-2.3**: Document the computed relationship:
  | Session Tracked | Team Active | Computed Mode |
  |-----------------|-------------|---------------|
  | No | - | native |
  | Yes | No | cross-cutting |
  | Yes | Yes | orchestrated |

#### FR-3: 2x2 Matrix Validation

- **FR-3.1**: Formalize the constraint that "Team Active + No Session" is invalid:
  ```
                | No Session   | Session Active |
  |-------------|--------------|----------------|
  | No Team     | Native       | Cross-Cutting  |
  | Team Active | (INVALID)    | Orchestrated   |
  ```

- **FR-3.2**: Attempting to activate a team without a session should produce an error with guidance: "Cannot activate team without a session. Use /start first."

- **FR-3.3**: If ACTIVE_RITE exists but no session is tracked, `has_active_team()` returns `false` (the orphaned ACTIVE_RITE is ignored).

#### FR-4: PARKED as Lifecycle Modifier

- **FR-4.1**: PARKED suspends the "Team Active" dimension. When a session is PARKED:
  - `is_session_tracked()` returns `true`
  - `has_active_team()` returns `false` (delegation suspended)
  - `is_session_parked()` returns `true`
  - `execution_mode()` returns `cross-cutting`

- **FR-4.2**: When PARKED session is resumed:
  - If team still configured: `has_active_team()` returns `true`, mode becomes `orchestrated`
  - If team removed while parked: `has_active_team()` returns `false`, mode remains `cross-cutting`

- **FR-4.3**: Document PARKED as orthogonal to mode:
  ```
  PARKED is not a mode; it is a lifecycle state that temporarily suspends delegation.
  Mode = f(session_tracked, team_active)
  PARKED = modifier that sets team_active = false while preserving session_tracked
  ```

#### FR-5: Documentation Updates

- **FR-5.1**: Update `execution-mode.md` skill to present the 2x2 conceptual model alongside the 3-mode enum, explaining their equivalence.

- **FR-5.2**: Update `.claude/CLAUDE.md` execution mode table to include the 2x2 dimensions as a reference note.

- **FR-5.3**: Add inline comments to `session-manager.sh` explaining the 2x2 model and how `execution_mode()` is computed from primitives.

### Should Have

- **FR-S.1**: Expose primitives in `cmd_status()` JSON output:
  ```json
  {
    "session_tracked": true,
    "team_active": false,
    "parked": true,
    "execution_mode": "cross-cutting"
  }
  ```

- **FR-S.2**: Update hook context injection to include primitives in addition to mode string.

- **FR-S.3**: Add `primitives` subcommand to `session-manager.sh`:
  ```bash
  session-manager.sh primitives
  # Output: {"session_tracked": true, "team_active": false, "parked": false}
  ```

### Could Have

- **FR-C.1**: Create ADR documenting the evolution from 3-mode enum to 2x2 model.

- **FR-C.2**: Deprecation warnings (logged, not displayed) when code uses mode string comparison instead of primitives.

## Non-Functional Requirements

- **NFR-1**: Performance - Primitive functions must complete in <10ms each. Combined primitives for `execution_mode()` must stay under 50ms total.

- **NFR-2**: Reliability - Primitive functions must never throw errors; return `false` on any detection failure.

- **NFR-3**: Testability - Each primitive is independently testable. Unit tests can mock session state to verify primitive behavior.

- **NFR-4**: Backward Compatibility - All existing hooks, skills, and documentation using `execution_mode()` string comparison continue working.

## Technical Approach

### Implementation Overview

The implementation preserves backward compatibility by introducing primitives as internal functions, then refactoring `execution_mode()` to use them.

### Phase 1: Add Primitive Functions

Add to `session-manager.sh` after `has_session()` function:

```bash
# Check if session is being tracked (exists and not archived)
is_session_tracked() {
    local session_id
    session_id=$(get_session_id 2>/dev/null) || return 1
    [[ -z "$session_id" ]] && return 1

    local session_dir="$SESSIONS_DIR/$session_id"
    [[ ! -f "$session_dir/SESSION_CONTEXT.md" ]] && return 1

    local status
    status=$(fsm_get_state "$session_id" 2>/dev/null) || status="NONE"
    [[ "$status" == "NONE" || "$status" == "ARCHIVED" ]] && return 1

    return 0
}

# Check if session is parked
is_session_parked() {
    is_session_tracked || return 1
    local session_id
    session_id=$(get_session_id)
    local status
    status=$(fsm_get_state "$session_id" 2>/dev/null)
    [[ "$status" == "PARKED" ]]
}

# Check if team delegation is active (tracked session + ACTIVE status + valid team)
has_active_team() {
    is_session_tracked || return 1
    is_session_parked && return 1  # PARKED suspends delegation

    local session_id
    session_id=$(get_session_id)
    local status
    status=$(fsm_get_state "$session_id" 2>/dev/null)
    [[ "$status" != "ACTIVE" ]] && return 1

    local active_team
    active_team=$(cat ".claude/ACTIVE_RITE" 2>/dev/null || echo "none")
    [[ -z "$active_team" || "$active_team" == "none" || "$active_team" == "null" ]] && return 1

    local team_pack_dir="${ROSTER_HOME:-$HOME/.config/roster}/teams/$active_team"
    [[ ! -d "$team_pack_dir" ]] && return 1

    return 0
}
```

### Phase 2: Refactor `execution_mode()`

Simplify `execution_mode()` to use primitives:

```bash
execution_mode() {
    if ! is_session_tracked; then
        echo "native"
    elif has_active_team; then
        echo "orchestrated"
    else
        echo "cross-cutting"
    fi
}
```

### Phase 3: Update Documentation

Update `execution-mode.md` to include:
- 2x2 conceptual model explanation
- Mapping between 2x2 and 3-mode enum
- PARKED as lifecycle modifier explanation
- New primitive functions reference

### Phase 4: Extend Status Output

Update `cmd_status()` to include primitives in JSON output for consumers that want richer information.

## Migration Plan

### Approach: Additive, Non-Breaking

This migration is entirely additive. No existing functionality is removed or changed in behavior.

| Phase | Change | Risk | Rollback |
|-------|--------|------|----------|
| 1 | Add primitive functions | None - new code | Delete functions |
| 2 | Refactor `execution_mode()` internals | Low - same output | Restore original implementation |
| 3 | Update documentation | None | Revert docs |
| 4 | Extend status output | None - additive fields | Remove fields |

### Deprecation Strategy (Future)

After primitives are stable and hooks have migrated to use them (optional):

1. Mark direct mode string comparison as "legacy pattern" in documentation
2. Log deprecation notice (not user-visible) when mode strings are used
3. No forced migration - 3-mode enum remains valid indefinitely

## Success Criteria

### Functional Verification

- [ ] `is_session_tracked()` returns correct boolean for all session states
- [ ] `has_active_team()` returns correct boolean for all team configurations
- [ ] `is_session_parked()` returns correct boolean for PARKED vs non-PARKED
- [ ] `execution_mode()` returns identical values to current implementation for all inputs
- [ ] PARKED sessions with team configured return `cross-cutting` mode
- [ ] Resumed PARKED sessions with team return `orchestrated` mode
- [ ] All existing tests pass without modification

### Integration Verification

- [ ] `delegation-check.sh` works unchanged
- [ ] `orchestrated-mode.sh` works unchanged
- [ ] `session-context.sh` works unchanged
- [ ] `/consult` routing works unchanged
- [ ] `/start`, `/park`, `/resume`, `/wrap` work unchanged

### Documentation Verification

- [ ] `execution-mode.md` updated with 2x2 model
- [ ] `.claude/CLAUDE.md` includes 2x2 reference
- [ ] `session-manager.sh` includes inline documentation

## Edge Cases

| Case | Expected Behavior |
|------|-------------------|
| ACTIVE_RITE exists but no session | `has_active_team()` returns false, orphaned file ignored |
| Session ACTIVE, team file missing | `has_active_team()` returns false, `execution_mode()` returns cross-cutting |
| Session PARKED, team configured | `has_active_team()` returns false, `execution_mode()` returns cross-cutting |
| Session resumed, team still configured | `has_active_team()` returns true, `execution_mode()` returns orchestrated |
| Session resumed, team removed while parked | `has_active_team()` returns false, `execution_mode()` returns cross-cutting |
| fsm_get_state() fails | Primitives return false, `execution_mode()` falls back to native |
| Multiple primitives called in sequence | Each reads fresh state (no caching race conditions) |

## Dependencies and Risks

### Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| `session-manager.sh` | Internal | roster | Ready |
| `session-fsm.sh` | Internal | roster | Ready |
| `execution-mode.md` | Internal | roster | Ready |
| `.claude/CLAUDE.md` | Internal | roster | Ready |

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Subtle behavior difference in refactored `execution_mode()` | Low | High | Exhaustive test matrix comparing old vs new |
| Performance regression from multiple primitive calls | Low | Medium | Each primitive is fast; total stays under 50ms |
| Confusion from having both models documented | Medium | Low | Clear mapping, emphasize equivalence |
| Hooks start using primitives inconsistently | Low | Low | Primitives are optional; mode string remains canonical |

## Out of Scope

- Changing the 3-mode enum values (native, cross-cutting, orchestrated remain unchanged)
- Modifying session lifecycle (ACTIVE, PARKED, ARCHIVED unchanged)
- Changing hook behavior or warning thresholds
- Adding new execution modes
- Modifying `/start`, `/team`, `/park`, `/resume`, `/wrap` commands
- UI changes or new commands

---

## Traceability

| Requirement | Source |
|-------------|--------|
| FR-1.x (Primitive Functions) | Gap Analysis: "Add primitive functions: has_session(), has_active_team()" |
| FR-2.x (Backward Compatibility) | Gap Analysis: "Keep execution_mode() for backward compatibility" |
| FR-3.x (2x2 Validation) | Gap Analysis: "Team Active + No Session is invalid state" |
| FR-4.x (PARKED Lifecycle) | Gap Analysis: "PARKED is orthogonal... lifecycle state, not mode dimension" |
| FR-5.x (Documentation) | Gap Analysis: "Update documentation to present 2x2 conceptually" |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-execution-mode-2x2.md` | Created |
| Gap Analysis | `/Users/tomtenuta/Code/roster/docs/analysis/GAP-execution-mode-2x2.md` | Reference |
| Current Implementation | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Reference |
| Current Documentation | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/execution-mode.md` | Reference |

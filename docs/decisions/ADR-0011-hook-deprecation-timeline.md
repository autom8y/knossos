# ADR-0011: Hook Deprecation Timeline

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-01-05 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The Knossos platform has completed migration from shell-based hooks to Go-based hooks via the Ariadne CLI (`ari`). This migration, documented in the MANIFEST.md, consolidated 13+ bash hooks and 16 library files into 6 purpose-built Go hooks implementing Thread Contract v2.

### Current State

**Deprecated hooks location**: `.claude/hooks/.deprecated/`

The deprecated directory contains 29 files across 4 categories:

| Category | Files | Line Count |
|----------|-------|------------|
| context-injection/ | 3 | ~400 LOC |
| session-guards/ | 3 | ~600 LOC |
| tracking/ | 3 | ~500 LOC |
| validation/ | 4 | ~700 LOC |
| lib/ | 16 | ~3,500 LOC |
| **Total** | **29** | **~5,700 LOC** |

**New hooks location**: `.claude/hooks/ari/`

The new system provides 7 thin bash wrappers that dispatch to the Go binary:

| Hook | Event | Description |
|------|-------|-------------|
| context.sh | SessionStart | Injects session context via `ari hook context` |
| autopark.sh | Stop | Auto-parks session via `ari hook autopark` |
| writeguard.sh | PreToolUse (Edit/Write) | Guards session context writes via `ari hook writeguard` |
| validate.sh | PreToolUse (Bash) | Validates bash commands via `ari hook validate` |
| thread.sh | PostToolUse | Tracks artifacts and commits via `ari hook clew` |
| route.sh | UserPromptSubmit | Routes slash commands via `ari hook route` |
| cognitive-budget.sh | PostToolUse | Tracks cognitive budget via `ari hook budget` |

### Migration Mapping

The MANIFEST.md documents the complete replacement mapping:

| Old (Bash) | New (Go) | Status |
|------------|----------|--------|
| session-context.sh + orchestrated-mode.sh | `ari hook context` | Replaced |
| auto-park.sh | `ari hook autopark` | Replaced |
| session-write-guard.sh | `ari hook writeguard` | Replaced |
| command-validator.sh | `ari hook validate` | Replaced |
| artifact-tracker.sh + commit-tracker.sh + session-audit.sh | `ari hook clew` | Replaced |
| start-preflight.sh + orchestrator-router.sh | `ari hook route` | Replaced |
| delegation-check.sh | REMOVED | Deprecated (poor workflow enforcement) |
| orchestrator-bypass-check.sh | REMOVED | Deprecated (rearchitected) |
| coach-mode.sh | REMOVED | Deprecated (coach framework deprecated) |

### Library Replacement

All library functionality has been reimplemented in Go:

| Old Library (Bash) | New Location (Go) |
|-------------------|-------------------|
| session-manager.sh (829 LOC) | `ariadne/internal/session/` |
| session-fsm.sh (853 LOC) | `ariadne/internal/session/fsm.go` |
| hooks-init.sh (234 LOC) | No longer needed (Go binary) |
| session-core.sh | `ariadne/internal/session/core.go` |
| session-state.sh | `ariadne/internal/session/state.go` |
| session-utils.sh | `ariadne/internal/session/utils.go` |
| session-migrate.sh | `ariadne/internal/session/migrate.go` |
| logging.sh | `ariadne/internal/logging/` |
| config.sh | `ariadne/internal/config/` |
| primitives.sh | Go standard library |
| preferences-loader.sh | `ariadne/internal/config/preferences.go` |
| team-context-loader.sh | `ariadne/internal/team/` |
| worktree-manager.sh | `ariadne/internal/worktree/` |
| artifact-validation.sh | `ariadne/internal/artifact/` |
| handoff-validator.sh | `ariadne/internal/handoff/` |
| orchestration-audit.sh | `ariadne/internal/audit/` |

### Forces

- **Technical debt**: Maintaining parallel implementations increases cognitive load
- **Rollback capability**: The deprecated hooks provide fallback if Go implementation has issues
- **Clean repository**: Dead code creates confusion for new contributors
- **Feature flag**: `USE_ARI_HOOKS=0` currently enables fallback to legacy hooks
- **Testing confidence**: Go hooks have been operational since 2026-01-04

## Decision

Implement a three-phase deprecation timeline with explicit gates:

### Phase 1: Documentation (Immediate - T+0)

**Timeline**: Effective immediately upon ADR acceptance

**Actions**:
1. Mark all `.deprecated/` hooks as deprecated in documentation
2. Add deprecation warnings to CLAUDE.md and related guides
3. Update ADR-0002 (Hook Library Resolution) to reference this ADR
4. Ensure MANIFEST.md is complete and accurate

**Exit criteria**:
- [ ] MANIFEST.md reviewed and current
- [ ] All documentation references updated
- [ ] No new development on deprecated hooks

### Phase 2: Sync Removal (T+30 days - 2026-02-04)

**Timeline**: 30 days after Phase 1

**Actions**:
1. Remove deprecated hooks from `sync-user-hooks.sh` materialization
2. Remove deprecated hooks from `install-hooks.sh` distribution
3. Update settings.json templates to exclude deprecated hooks
4. Remove `USE_ARI_HOOKS` feature flag (Go hooks become mandatory)

**Gate criteria** (all must be true):
- [ ] No production incidents attributed to Go hooks in 30-day window
- [ ] `ari hook` commands tested across all supported platforms
- [ ] No open issues blocking Go hook adoption
- [ ] Rollback procedure documented and tested

**Exit criteria**:
- [ ] Sync scripts no longer reference deprecated hooks
- [ ] Feature flag removed from codebase
- [ ] User documentation updated for mandatory ari hooks

### Phase 3: Deletion (T+60 days - 2026-03-06)

**Timeline**: 60 days after Phase 1 (30 days after Phase 2)

**Actions**:
1. Delete `.claude/hooks/.deprecated/` directory entirely
2. Delete `roster/hooks/.deprecated/` if it exists
3. Archive MANIFEST.md to `docs/archive/` for historical reference
4. Update this ADR status to "Implemented"

**Gate criteria** (all must be true):
- [ ] Phase 2 complete for at least 30 days
- [ ] No requests to restore deprecated functionality
- [ ] All team members acknowledge deprecation

**Exit criteria**:
- [ ] No deprecated hook files in repository
- [ ] MANIFEST.md archived with deprecation history
- [ ] This ADR marked "Implemented"

## Migration Guidance

### For Users with Legacy Hooks

If you have customized the legacy bash hooks:

1. **Identify customizations**: Compare your hooks against the archived MANIFEST.md
2. **Port to Go**: Custom logic should be added to `ariadne/internal/cmd/hook/`
3. **Extend via wrappers**: For simple customizations, extend the bash wrappers in `.claude/hooks/ari/`

### For Users with `USE_ARI_HOOKS=0`

After Phase 2, this environment variable will have no effect:

1. **Test now**: Run `USE_ARI_HOOKS=1` in your environment to test Go hooks
2. **Report issues**: File issues before Phase 2 gate evaluation
3. **Plan migration**: Custom bash hooks must be ported before Phase 2

### Edge Cases

| Scenario | Guidance |
|----------|----------|
| Custom session-manager.sh extensions | Port to `ari session` commands |
| Custom logging.sh formats | Use `ariadne/internal/logging/` configuration |
| Custom validation rules | Add to `ari hook validate` via config |
| Orchestration customizations | Consult Architecture Team before Phase 2 |

## Consequences

### Positive

1. **Reduced maintenance burden**: Single implementation (Go) instead of dual (bash + Go)
2. **Cleaner repository**: ~5,700 LOC of deprecated code removed
3. **Faster hooks**: Go implementation is consistently faster than bash
4. **Better testing**: Go hooks have unit tests; bash hooks did not
5. **Clew Contract v2**: Unified event tracking architecture

### Negative

1. **Forced migration**: Users with custom bash hooks must port them
2. **Loss of fallback**: After Phase 2, Go hooks are the only option
3. **Binary dependency**: `ari` binary must be built and available

### Neutral

1. **Historical context preserved**: MANIFEST.md archived, not deleted
2. **ADR chain maintained**: This ADR references ADR-0002, ADR-0009
3. **Feature parity**: All deprecated functionality has Go equivalent

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Go hooks have undiscovered bugs | 30-day observation window in Phase 1-2 gap |
| Users miss deprecation notice | Multiple documentation touchpoints, gate checks |
| Rollback needed after Phase 3 | Git history preserves all deleted files |
| Custom hooks not ported | Clear migration guidance, Architecture Team support |

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This ADR | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0011-hook-deprecation-timeline.md` | Created |
| Deprecated hooks manifest | `/Users/tomtenuta/Code/roster/.claude/hooks/.deprecated/MANIFEST.md` | Existing |
| New hooks configuration | `/Users/tomtenuta/Code/roster/.claude/hooks/ari/hooks.yaml` | Existing |
| Related ADR (hooks architecture) | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0002-hook-library-resolution-architecture.md` | Existing |

## Related Decisions

- **ADR-0002**: Hook Library Resolution Architecture (original hook system design)
- **ADR-0009**: Knossos-Roster Identity (Ariadne naming and integration context)

## References

- MANIFEST.md: `.claude/hooks/.deprecated/MANIFEST.md`
- Clew Contract v2: PRD-ariadne.md Section 3.2
- Ariadne implementation: `ariadne/internal/cmd/hook/`
- hooks.yaml configuration: `.claude/hooks/ari/hooks.yaml`

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-01-05 | Claude Code (Architect) | Initial acceptance - hook deprecation timeline |

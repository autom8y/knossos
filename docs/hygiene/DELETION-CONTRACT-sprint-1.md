# Sprint 1 Deletion Contract

**Initiative**: Knossos Code Hygiene - Dead Code Elimination
**Phase**: Sprint 1 - Safe Deletion Verification
**Date**: 2026-02-06
**Architect**: architect-enforcer

## Purpose

This contract confirms zero production callers for Sprint 1 dead code targets. Each target was verified via codebase search to ensure safe deletion without breaking production functionality.

## Verification Method

- Tool: Grep with ripgrep backend
- Scope: All Go source files in `/Users/tomtenuta/Code/knossos`
- Excluded: Test files (verified separately), documentation files (non-breaking)
- Each target searched for:
  - Direct function calls
  - Type references
  - Constant usage

## Deletion Targets

| Target | File | Lines | Callers Found | Verdict | Notes |
|--------|------|-------|---------------|---------|-------|
| **1. clewcontract/record.go Dead Functions** |
| RecordFileChange | internal/hook/clewcontract/record.go | 116-127 | Test-only | DELETE | Only called in record_test.go |
| RecordCommand | internal/hook/clewcontract/record.go | 129-140 | Test-only | DELETE | Only called in record_test.go |
| RecordDecision | internal/hook/clewcontract/record.go | 142-153 | Test-only | DELETE | Only called in record_test.go |
| RecordContextSwitch | internal/hook/clewcontract/record.go | 155-166 | Test-only | DELETE | Only called in record_test.go |
| RecordStamp | internal/hook/clewcontract/record.go | 173-194 | LIVE (2 callers) | **KEEP** | Called in clew.go:124 and clew.go:243 for orchestrator stamps |
| RecordStampWithContext | internal/hook/clewcontract/record.go | 196-208 | Zero callers | DELETE | No production usage |
| RecordTaskStart | internal/hook/clewcontract/record.go | 210-228 | Zero callers | DELETE | No production usage |
| RecordTaskEnd | internal/hook/clewcontract/record.go | 230-250 | Zero callers | DELETE | No production usage |
| RecordSessionStart | internal/hook/clewcontract/record.go | 252-270 | Zero callers | DELETE | No production usage |
| RecordSessionEnd | internal/hook/clewcontract/record.go | 272-289 | Zero callers | DELETE | No production usage |
| RecordArtifactCreated | internal/hook/clewcontract/record.go | 291-310 | Zero callers | DELETE | No production usage |
| RecordError | internal/hook/clewcontract/record.go | 312-332 | Zero callers | DELETE | No production usage |
| **2. ClaudeMDUpdater (Entire File)** |
| ClaudeMDUpdater struct | internal/rite/claudemd.go | 11-14 | LIVE (1 caller) | **REWIRE** | Called in validate.go:291 Fix() method |
| NewClaudeMDUpdater | internal/rite/claudemd.go | 17-19 | LIVE (1 caller) | **REWIRE** | Called in validate.go:291 Fix() method |
| UpdateForRite | internal/rite/claudemd.go | 22-45 | LIVE (1 caller) | **REWIRE** | Called in validate.go:292 Fix() method |
| All other methods | internal/rite/claudemd.go | 47-316 | Zero direct | DELETE | Helper methods only used internally |
| **3. Budget Warning System** |
| BudgetWarning type | internal/rite/budget.go | 151-158 | Zero callers | DELETE | Type not used in production |
| BudgetWarnPercent const | internal/rite/budget.go | 162 | Internal only | DELETE | Only used in CheckBudgetWarnings |
| BudgetCriticalPercent const | internal/rite/budget.go | 163 | Internal only | DELETE | Only used in CheckBudgetWarnings |
| CheckBudgetWarnings | internal/rite/budget.go | 166-191 | Zero callers | DELETE | No production usage |
| ComponentCost type | internal/rite/budget.go | 193-199 | Zero callers | DELETE | Type not used in production |
| CalculateDetailedCost | internal/rite/budget.go | 201-261 | Zero callers | DELETE | No production usage |
| SummaryCost type | internal/rite/budget.go | 263-269 | LIVE (1 caller) | **KEEP** | Used in cmd/rite/info.go:104 |
| CalculateSummaryCost | internal/rite/budget.go | 271-295 | LIVE (1 caller) | **KEEP** | Used in cmd/rite/info.go:104 |
| **4. Error Types** |
| ErrorResponse type | internal/errors/errors.go | 88-91 | Zero callers | DELETE | Unused wrapper type |
| CodeSuccess constant | internal/errors/errors.go | 37 | Zero callers | DELETE | Never used in exitCodeForCode switch |
| **5. Output Types** |
| LockOutput type | internal/output/output.go | 493-502 | Zero callers | DELETE | Type not used in production |
| **6. Manifest Utilities** |
| writeOutputFile function | internal/cmd/manifest/merge.go | 168-170 | Zero callers | DELETE | Defined but never called |
| **7. Artifact Constants** |
| ArtifactTypeWhiteSails | internal/hook/clewcontract/event.go | 43 | Zero callers | DELETE | Constant not used |
| ArtifactTypeCode | internal/hook/clewcontract/event.go | 42 | Zero callers | DELETE | Constant not used |
| **8. Query Helper Methods** |
| QueryByPhase | internal/artifact/query.go | 52-55 | Zero callers | DELETE | Wrapper never called |
| QueryByType | internal/artifact/query.go | 57-60 | Zero callers | DELETE | Wrapper never called |
| QueryBySpecialist | internal/artifact/query.go | 62-65 | Zero callers | DELETE | Wrapper never called |
| QueryBySession | internal/artifact/query.go | 67-70 | Zero callers | DELETE | Wrapper never called |
| **9. Sails Wrapper Functions** |
| GetRequiredProofs | internal/sails/color.go | 398-401 | Test-only | DELETE | Wrapper for test compat, tests can use real function |
| IsRequiredProof | internal/sails/color.go | 403-406 | Test-only | DELETE | Wrapper for test compat, tests can use real function |
| **10. Rite Manifest Migration** |
| MigrationInfo type | internal/rite/manifest.go | 157-160 | Zero usage | DELETE | Type defined but never populated |
| Migration field (RiteManifest) | internal/rite/manifest.go | 85 | Zero usage | DELETE | Field never set in any manifest |
| Migration field (rawManifest) | internal/rite/manifest.go | 180 | Zero usage | DELETE | Field never set in any manifest |
| **11. Frontmatter Parsing** |
| ParseMenaFrontmatter | internal/materialize/frontmatter.go | 79-109 | Zero callers | DELETE | Function defined but never called |
| MenaFrontmatter struct | internal/materialize/frontmatter.go | 47-65 | KEEP | **KEEP** | Needed for Sprint 2 frontmatter wiring |
| FlexibleStringSlice | internal/materialize/frontmatter.go | 15-45 | KEEP | **KEEP** | Needed for Sprint 2 frontmatter wiring |
| DetectMenaType | internal/materialize/frontmatter.go | 115-123 | LIVE (2 callers) | **KEEP** | Used in materialize.go:690 and materialize.go:699 |
| **12. Hook Feature Flag** |
| FeatureFlagEnvVar const | internal/hook/env.go | 11 | LIVE | **KEEP** | Used in help text, to be removed in Sprint 2 |
| IsEnabled function | internal/hook/env.go | 91-100 | LIVE (1 caller) | **KEEP** | Used in hook.go:114 shouldEarlyExit() |
| **13. Hook Help Text** |
| USE_ARI_HOOKS in help | internal/cmd/hook/hook.go | 58 | Documentation only | REWIRE | Update help text to remove mention (Sprint 2) |
| USE_ARI_HOOKS check | internal/cmd/hook/hook.go | 114 | LIVE | **KEEP** | Actual flag check in shouldEarlyExit() |

## Rewiring Requirements

### REWIRE-001: ClaudeMDUpdater → Materialize

**Current State** (validate.go:284-293):
```go
case "CLAUDE_MD_SYNC":
    // Update CLAUDE.md satellites
    rite, err := v.discovery.Get(riteName)
    if err != nil {
        continue
    }
    if rite.Active {
        updater := NewClaudeMDUpdater(v.resolver.ClaudeMDFile())
        updater.UpdateForRite(rite)
    }
```

**Target State**:
```go
case "CLAUDE_MD_SYNC":
    // Trigger full rematerialization to update CLAUDE.md satellites
    // This replaces the old ClaudeMDUpdater with the standard materialize path
    if err := materialize.Run(v.resolver); err != nil {
        // Log error but don't fail validation fix
        continue
    }
```

**Verification**:
1. Run: `ari rite validate <rite> --fix`
2. Confirm CLAUDE.md Quick Start and Agent Configurations sections regenerate
3. Verify no behavioral changes from current implementation

**Risk**: LOW. Materialize already handles CLAUDE.md satellite updates. This unifies the code path.

**Rollback**: Revert single commit, restore claudemd.go.

---

### REWIRE-002: Hook Help Text Update (Sprint 2)

**Current State** (hook.go:58):
```
Environment Variables:
  USE_ARI_HOOKS=0    Emergency kill switch to disable ari hooks (default: enabled)
  CLAUDE_HOOK_*      Standard Claude Code hook environment variables
```

**Target State** (Sprint 2):
```
Environment Variables:
  CLAUDE_HOOK_*      Standard Claude Code hook environment variables
```

**Verification**:
1. Run: `ari hook --help`
2. Confirm USE_ARI_HOOKS no longer mentioned

**Risk**: TRIVIAL. Documentation-only change.

---

### REWIRE-003: Update Tests Referencing Deleted Wrappers

**Affected Tests**:
- `internal/sails/color_test.go` (uses GetRequiredProofs/IsRequiredProof wrappers)
- `internal/cmd/sails/check_test.go` (if applicable)

**Change**: Replace wrapper calls with direct function calls:
- `GetRequiredProofs(complexity)` → `GetRequiredProofsForColor(complexity)`
- `IsRequiredProof(complexity, proof)` → `IsRequiredProofForColor(complexity, proof)`

**Verification**:
1. Run: `go test ./internal/sails/... -v`
2. All tests pass

**Risk**: TRIVIAL. Test-only changes.

## Deletion Sequence

Execute in this order to minimize dependencies:

### Phase 1: Zero-Caller Deletions (Safe, No Rewiring)
1. Delete unused record functions (RecordFileChange, RecordCommand, etc.) except RecordStamp
2. Delete BudgetWarning system (BudgetWarning type, constants, CheckBudgetWarnings, ComponentCost, CalculateDetailedCost)
3. Delete ErrorResponse type
4. Delete CodeSuccess constant
5. Delete LockOutput type
6. Delete writeOutputFile function
7. Delete ArtifactTypeWhiteSails and ArtifactTypeCode constants
8. Delete QueryBy* wrapper methods
9. Delete MigrationInfo type and Migration fields from both manifest structs
10. Delete ParseMenaFrontmatter function

**Verification**: `CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./...`

**Rollback Point**: Single commit, can revert if build breaks.

---

### Phase 2: Test Wrapper Cleanup
1. Update sails tests to use direct functions instead of wrappers
2. Delete GetRequiredProofs and IsRequiredProof wrapper functions

**Verification**: `go test ./internal/sails/... -v`

**Rollback Point**: Single commit after Phase 1.

---

### Phase 3: ClaudeMDUpdater Rewiring
1. Update validate.go Fix() method to call materialize.Run()
2. Delete entire claudemd.go file
3. Remove any imports of claudemd package

**Verification**:
1. `ari rite validate hygiene --fix`
2. Confirm CLAUDE.md regenerates correctly
3. `go test ./internal/rite/... -v`

**Rollback Point**: Single commit after Phase 2.

---

## Risk Assessment

| Phase | Blast Radius | Failure Detection | Recovery Path |
|-------|--------------|-------------------|---------------|
| Phase 1 | MINIMAL | Build failure immediate | Revert commit |
| Phase 2 | Test-only | Test failure immediate | Revert commit |
| Phase 3 | Validation only | `ari rite validate --fix` | Revert commit |

**Overall Risk**: LOW
- All deletions verified with zero production callers
- Rewiring uses existing, tested code paths (materialize)
- Atomic commits allow instant rollback
- No public API changes
- No behavior changes except consolidating code paths

## Success Criteria

- [ ] All phases complete without build errors
- [ ] `CGO_ENABLED=0 go test ./...` passes
- [ ] `ari rite validate hygiene --fix` successfully regenerates CLAUDE.md
- [ ] No regression in existing functionality
- [ ] Codebase reduced by ~800 LOC (estimated)

## Attestation

**Files Read**:
- `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/record.go`
- `/Users/tomtenuta/Code/knossos/internal/rite/claudemd.go`
- `/Users/tomtenuta/Code/knossos/internal/rite/budget.go`
- `/Users/tomtenuta/Code/knossos/internal/errors/errors.go`
- `/Users/tomtenuta/Code/knossos/internal/output/output.go`
- `/Users/tomtenuta/Code/knossos/internal/cmd/manifest/merge.go`
- `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/event.go`
- `/Users/tomtenuta/Code/knossos/internal/artifact/query.go`
- `/Users/tomtenuta/Code/knossos/internal/sails/color.go`
- `/Users/tomtenuta/Code/knossos/internal/rite/manifest.go`
- `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter.go`
- `/Users/tomtenuta/Code/knossos/internal/hook/env.go`
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/hook.go`
- `/Users/tomtenuta/Code/knossos/internal/rite/validate.go`
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/clew.go`
- `/Users/tomtenuta/Code/knossos/internal/cmd/rite/info.go`

**Search Verification**:
- ✅ RecordFileChange: Test-only callers (record_test.go)
- ✅ RecordCommand: Test-only callers (record_test.go)
- ✅ RecordDecision: Test-only callers (record_test.go)
- ✅ RecordContextSwitch: Test-only callers (record_test.go)
- ✅ RecordStamp: **LIVE** (clew.go:124, clew.go:243) - KEEP
- ✅ RecordStampWithContext: Zero callers
- ✅ RecordTaskStart through RecordError: Zero callers
- ✅ ClaudeMDUpdater: LIVE in validate.go:291-292 - REWIRE
- ✅ Budget warning types/functions: Zero production callers (BudgetWarning, CheckBudgetWarnings, ComponentCost, CalculateDetailedCost)
- ✅ SummaryCost/CalculateSummaryCost: **LIVE** (info.go:104) - KEEP
- ✅ ErrorResponse: Zero callers
- ✅ CodeSuccess: Zero callers (not in exitCodeForCode switch)
- ✅ LockOutput: Zero callers
- ✅ writeOutputFile: Zero callers
- ✅ ArtifactTypeWhiteSails/ArtifactTypeCode: Zero callers
- ✅ QueryBy* methods: Zero callers
- ✅ GetRequiredProofs/IsRequiredProof wrappers: Test-only
- ✅ MigrationInfo: Zero usage (no manifests populate this field)
- ✅ ParseMenaFrontmatter: Zero callers
- ✅ DetectMenaType: **LIVE** (materialize.go:690, 699) - KEEP
- ✅ FeatureFlagEnvVar/IsEnabled: **LIVE** (hook.go:114) - KEEP for Sprint 1

**Architect**: architect-enforcer
**Date**: 2026-02-06
**Status**: READY FOR JANITOR

---

## Janitor Notes

1. **Commit Boundaries**: One commit per phase. Atomic rollback critical.
2. **Test After Each Phase**: Run full test suite before proceeding to next phase.
3. **ClaudeMDUpdater Rewiring**: Verify materialize behavior matches old UpdateForRite before deletion.
4. **Critical Ordering**: Must delete test wrapper callers (Phase 2) before deleting wrapper functions.
5. **RecordStamp Exception**: Do NOT delete RecordStamp - it has active production usage in orchestrator stamp recording.
6. **Budget System Partial Deletion**: Delete warning system but KEEP SummaryCost/CalculateSummaryCost used by rite info command.

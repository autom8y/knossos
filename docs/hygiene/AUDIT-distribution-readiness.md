# Audit Report: Distribution Readiness

> **Auditor**: audit-lead (hygiene rite)
> **Date**: 2026-02-08
> **Plan**: `docs/hygiene/PLAN-distribution-readiness.md`
> **Verdict**: **APPROVED WITH NOTES**

---

## Executive Summary

| Metric | Value |
|--------|-------|
| Work units planned | 18 (17 active + 1 deferred) |
| Work units executed | 16 (WU-015 deferred per stakeholder decision) |
| Hygiene commits | 10 |
| Interleaved commits (other session) | 4 |
| Build status | PASS (`go build ./...`) |
| Test status | PASS (`go test ./...` -- all packages pass) |
| Smells addressed | 20 of 22 (2 were "no action" in plan) |
| New packages created | 2 (`internal/fileutil/`, `internal/lock/moirai.go`) |
| Lines deleted (est.) | ~700+ (dead code, duplication, legacy types) |
| ADR created | ADR-0026 (Unified Provenance) |

All 16 active work units were executed. Build compiles cleanly. Full test suite passes. Behavior is preserved across all refactored code paths.

---

## 1. Build and Test Verification

**Build**: `CGO_ENABLED=0 go build ./...` -- PASS (zero errors, zero warnings)

**Tests**: `CGO_ENABLED=0 go test ./...` -- PASS (all 43 packages, zero failures)

Packages with tests that exercised refactored code all pass:
- `internal/fileutil` -- 7 tests (new package)
- `internal/lock` -- 5 new tests (MoiraiLock, IsStale variants)
- `internal/cmd/hook` -- precompact + writeguard tests pass
- `internal/cmd/session` -- session recovery tests pass
- `internal/materialize` -- all materialize tests pass
- `internal/inscription` -- marker/backup tests pass
- `internal/session` -- rotation + context tests pass
- `internal/usersync` -- sync tests pass

---

## 2. Plan Coverage Verification

| WU | Status | Verified |
|----|--------|----------|
| WU-001: Delete StagedMaterialize + cloneDir | DONE | `StagedMaterialize`, `cloneDir` absent from `materialize.go`. `staging_test.go` deleted. Write tests preserved in `write_test.go` (later consolidated into fileutil). |
| WU-002: Commit untracked files | DONE | `rotation.go`, `rotation_test.go`, `precompact.go`, `precompact_test.go`, `hooks.yaml` committed in `314641e`. |
| WU-003: Consolidate lock reading | DONE | `internal/lock/moirai.go` created with `MoiraiLock`, `ReadMoiraiLock`, `IsMoiraiLockStale`. `IsStale` exported with `treatLegacyAsStale` parameter. `IsStaleForTest` deleted. `IsStaleFile` convenience function added. Tests in `moirai_test.go`. |
| WU-004: Create internal/fileutil/ | DONE | `fileutil.go` with `AtomicWriteFile` (CreateTemp + Sync + Rename) and `WriteIfChanged`. 7 tests in `fileutil_test.go`. |
| WU-005: Rewrite writeguard.go | DONE | `isMoiraiLockHeld` now calls `lock.ReadMoiraiLock` and `lock.IsMoiraiLockStale`. Same fail-closed semantics preserved. |
| WU-006: Delete dead code | DONE | `materializeSettings`, `getCurrentRite`, `GetTemplatesDir` -- all absent from codebase. |
| WU-007: Delete Materialize() wrapper | DONE | No `Materialize(` method exists on Materializer (only `MaterializeWithOptions` and `MaterializeMinimal`). |
| WU-008: Rewrite recover.go | DONE | `isAdvisoryLockStale` is a one-liner: `return lock.IsStaleFile(lockPath, true)`. Preserves `treatLegacyAsStale=true` for recovery. |
| WU-009: Replace atomicWriteFile callers | DONE | `materialize.go:writeIfChanged` delegates to `fileutil.WriteIfChanged`. `rotation.go` local `atomicWriteFile` deleted. `inscription/backup.go` exported `AtomicWriteFile` deleted. |
| WU-010: Atomic write for Context.Save() | DONE | `context.go:Save()` uses `fileutil.AtomicWriteFile(path, data, 0644)`. |
| WU-011: Delete legacy templates | DONE | `templates/base-orchestrator.md` and `templates/orchestrator-base.md.tpl` -- both absent from filesystem. |
| WU-012: Delete ParseLegacyMarkers | DONE | `ParseLegacyMarkers`, `suggestRegionName`, `LegacyMarkerType`, `LegacyMarker`, `legacyPreserveRegex`, `legacySyncRegex` -- all absent from inscription package. |
| WU-013: Standardize precompact output | DONE (modified by interleaved commit) | See Section 5 below. |
| WU-014: Remove scope infrastructure | DONE | `MenaScope`, `MenaScopeBoth`, `MenaScopeUser`, `MenaScopeProject`, `scopeIncludesPipeline`, `PipelineScope` -- all absent from codebase. `Scope` field removed from `MenaFrontmatter`. |
| WU-015: Phase-specific error codes | DEFERRED | Per stakeholder decision. Not in scope. |
| WU-016: Complete ritesDir migration | DONE | `ritesDir` field absent from `Materializer` struct. `materializeMena` derives paths from `resolved.RitePath` via `filepath.Dir(resolved.RitePath)`. |
| WU-017: Warning log on parse failure | DONE | `log.Printf("Warning: malformed YAML frontmatter, treating as unscoped: %v", err)` at `project_mena.go:169`. |
| WU-018: Extract provenance to ADR | DONE | `docs/decisions/ADR-0026-unified-provenance.md` created (numbered 0026, not 024 as originally planned -- correct since ADR-0025 already existed). |

**Result**: 16/16 active work units verified. WU-015 deferred as planned.

---

## 3. Contract Verification

### WU-001: StagedMaterialize Removal
- **Before**: `StagedMaterialize` method + `cloneDir` helper in `materialize.go`, `staging_test.go` with 8 tests
- **After**: Both functions deleted, `staging_test.go` deleted, `claudeDirOverride` field retained (used by `trackState`)
- **Contract met**: Yes. Four write tests were moved to `write_test.go`, then consolidated into `fileutil_test.go` in WU-009.

### WU-003: Lock Consolidation
- **Before**: Two independent MoiraiLock implementations (writeguard.go, session/lock.go), two stale detection implementations
- **After**: Single `MoiraiLock` type in `internal/lock/moirai.go`, `IsStale` exported with `treatLegacyAsStale` parameter, `IsStaleFile` convenience function
- **Contract met**: Yes. Empty lock files return `true` (stale) per stakeholder decision. Tests verify this explicitly (`TestIsStale_EmptyFile`).

### WU-004: fileutil Package
- **Before**: Three `atomicWriteFile` implementations with divergent behavior
- **After**: Single `AtomicWriteFile` with CreateTemp + Sync + Rename (most robust variant), parent dir creation, and explicit permissions
- **Contract met**: Yes. 7 tests cover new file, overwrite, parent dir creation, no temp leftovers, permissions, and WriteIfChanged semantics.

### WU-005/008: Lock Consumer Rewrites
- **writeguard.go**: `isMoiraiLockHeld` now uses `lock.ReadMoiraiLock` + `lock.IsMoiraiLockStale`. Same fail-closed semantics (returns false on any error). Same agent check ("moirai" only).
- **recover.go**: `isAdvisoryLockStale` delegates to `lock.IsStaleFile(lockPath, true)`. Preserves intentional `treatLegacyAsStale=true` for recovery operations.
- **Contract met**: Yes for both.

### WU-009/010: fileutil Consumer Rewrites
- `materialize.go:writeIfChanged` delegates to `fileutil.WriteIfChanged`
- `rotation.go` local `atomicWriteFile` deleted (rotation now uses `appendToArchive` for archive writes, which was always separate)
- `inscription/backup.go` exported `AtomicWriteFile` deleted; internal `atomicWrite` method retained (it has different error wrapping semantics for the BackupManager)
- `context.go:Save()` upgraded from `os.WriteFile` to `fileutil.AtomicWriteFile`
- **Contract met**: Yes. The `inscription/backup.go` private `atomicWrite` method is intentionally retained -- it is a BackupManager implementation detail with different error wrapping, not part of the plan's target.

### WU-014/016: Scope + ritesDir Removal
- `MenaScope` type and all related infrastructure completely absent from codebase
- `ritesDir` field absent from `Materializer` struct
- `materializeMena` now derives `ritesBase` from `filepath.Dir(resolved.RitePath)` -- uses the resolution abstraction as stakeholder requested
- **Contract met**: Yes.

### WU-018: ADR-0026
- Numbered ADR-0026 (correct -- ADR-0025 for MenaScope already exists)
- References real types: `SourceResolver`, `ResolvedRite`, `Entry`, `Manifest`, `Region`, `OwnerType`
- Follows existing ADR format (Status table, Context, Decision, Consequences, Alternatives, References)
- Contains coherent migration path (4 phases)
- **Contract met**: Yes.

---

## 4. Commit Quality Assessment

| Commit | Atomicity | Message Quality | Reversibility |
|--------|-----------|-----------------|---------------|
| `6d84ac8` | Good -- batches all Phase 1 dead code deletion per stakeholder request | Clear conventional commit format | Single revert restores all dead code |
| `314641e` | Good -- commits untracked feature files as a unit | Correct `feat(session)` prefix | Single revert |
| `bc7da13` | Good -- one concern (lock consolidation) | Descriptive | Single revert |
| `e0f2375` | Good -- new package, no consumers yet | Correct `feat(fileutil)` prefix | Delete two files |
| `263e126` | Good -- consumer rewrites depend on WU-003 | Groups two related consumer rewrites | Revert safely (WU-003 remains) |
| `8033cd9` | Good -- consumer rewrites depend on WU-004 | Groups atomicWriteFile + Context.Save | Revert safely (WU-004 remains) |
| `88a2f6a` | Good -- single hook format concern | Correct `fix(hook)` prefix | Single revert |
| `96a15a5` | Good -- groups related cleanup (scope + ritesDir) | Descriptive | Single revert |
| `a10630f` | Good -- minimal, single concern (warning log) | Correct `fix(materialize)` prefix | Single revert |
| `265ba74` | Good -- documentation only | Correct `docs:` prefix | Single revert |

All commits follow the existing `type(scope): description` convention. All include `Co-Authored-By` attribution. Commit batching follows the stakeholder-approved plan (~10 commits).

---

## 5. Behavior Preservation Checklist

| Check | Result | Evidence |
|-------|--------|----------|
| Public API signatures unchanged | PASS | `MaterializeWithOptions`, `MaterializeMinimal` unchanged. `Materialize()` deleted (zero callers confirmed). |
| Return types preserved | PASS | All functions return same types. `Context.Save()` still returns `error`. |
| Error semantics preserved | PASS | writeguard fail-closed (returns false on error). Recovery treats legacy as stale. |
| Documented contracts preserved | PASS | No documented behavior changes. |
| Internal logging change (allowed) | PASS | `log.Printf` warning added on frontmatter parse failure (WU-017). |
| Test coverage maintained | PASS | Tests moved (staging_test.go -> write_test.go -> fileutil_test.go), not deleted. 12+ new tests added. |

### Regression Scan Results

| Pattern | Occurrences | Assessment |
|---------|-------------|------------|
| `StagedMaterialize` | 1 (comment in `cmd/sync/materialize.go:140`) | Advisory only -- explains why it is NOT used. Acceptable. |
| `cloneDir` | 0 in materialize package | Clean removal. |
| `MenaScope` | 0 across all Go files | Clean removal. |
| `ritesDir` | 0 in `materialize/materialize.go`; still exists in `internal/rite/`, `internal/usersync/`, `internal/cmd/agent/` | Correct -- those are separate `ritesDir` variables in other packages, unrelated to the Materializer struct field. |
| `atomicWriteFile` (as function def) | 1 (`fileutil.AtomicWriteFile`) | Only the canonical implementation remains. |
| `IsStaleForTest` | 0 | Clean removal. |
| `ParseLegacyMarkers` | 0 | Clean removal. |

---

## 6. Interleaved Commit Impact

Four commits from another session were interleaved between hygiene commits:
- `580f47e` feat(materialize): support async field in hook materialization pipeline
- `a7a6bd2` feat(hooks): enable async for clew and budget PostToolUse hooks
- `2ed353c` refactor(hook): align event names and remove legacy output code
- `662fb0e` feat(hooks): enable async for route UserPromptSubmit hook

**Impact on WU-013**: Commit `2ed353c` removed the `PreCompactOutput` type and `Result` type that WU-013 (`88a2f6a`) had just added/deprecated. The current precompact output uses a local `precompactResult` struct with plain JSON output. This is a valid simplification -- CC has no hookSpecificOutput schema for PreCompact events, so the envelope was unnecessary. The code comment at `precompact.go:18` accurately documents this: "CC has no hookSpecificOutput for PreCompact, so we emit plain JSON."

**Assessment**: The interleaved commits improved the outcome of WU-013 by removing dead code more aggressively than the plan specified. The current state is cleaner than what the plan called for. No regressions introduced.

---

## 7. Improvement Assessment

### Before
- 3 duplicate `atomicWriteFile` implementations (divergent safety: predictable .tmp vs CreateTemp, with/without Sync)
- 2 duplicate `MoiraiLock` parsing implementations (divergent time handling: string parse vs json.Unmarshal)
- 2 duplicate stale detection implementations (divergent legacy handling)
- ~200 lines of dead code (StagedMaterialize, cloneDir, materializeSettings, getCurrentRite, GetTemplatesDir, Materialize wrapper)
- ~130 lines of unused infrastructure (MenaScope, scopeIncludesPipeline, ParseLegacyMarkers, legacy templates)
- Silent failures on frontmatter parse errors
- Context.Save() used non-atomic writes
- Provenance design trapped in plan document

### After
- Single canonical `AtomicWriteFile` with best-practice safety (CreateTemp + Sync + Rename + parent dir creation)
- Single `MoiraiLock` type with exported `ReadMoiraiLock` and `IsMoiraiLockStale`
- Single `IsStale` with parameterized legacy handling
- All dead code removed
- All unused infrastructure removed
- Warning log on parse failures
- Atomic writes for session context
- Provenance design extracted to ADR-0026

**Net improvement**: Significant reduction in duplication, cleaner package boundaries, better safety guarantees for file writes, and improved observability. No new code smells introduced.

---

## 8. Advisory Notes (Non-Blocking)

1. **inscription/backup.go private `atomicWrite`**: The `BackupManager.atomicWrite` method still uses predictable `.tmp` suffix without `Sync()`. This is a different function from the standalone `AtomicWriteFile` that was deleted, and serves a different purpose (backup creation with error wrapping). However, it could benefit from the same CreateTemp + Sync pattern in a future cleanup pass. Not blocking because backup writes are not performance-critical and the predictable temp name has no security implications in this context.

2. **ADR numbering**: The plan specified ADR-024, but the janitor correctly created ADR-0026 (since ADR-0024 and ADR-0025 already exist). The plan document still references "ADR-024" in WU-018 text. Consider updating the plan reference for consistency.

3. **PreCompact output format**: The current implementation emits plain JSON (`{"reason":"..."}`) instead of the hookSpecificOutput envelope the plan specified. This is correct per the interleaved commit's rationale (CC has no hookSpecificOutput schema for PreCompact). The plan's assumption that all hooks should use the envelope was invalidated by deeper CC alignment work happening in parallel.

---

## 9. Verdict

### APPROVED WITH NOTES

All 16 active work units have been executed and verified. The build compiles cleanly. The full test suite passes. Behavior is preserved across all refactored paths. Commits are atomic, well-documented, and independently reversible. The codebase is measurably cleaner: duplication eliminated, dead code removed, shared utilities extracted, and a design ADR created for future provenance work.

The three advisory notes above are non-blocking observations for future consideration.

### Sign-Off

I have verified every contract against the plan, confirmed behavior preservation through test results and code review, and validated that no regressions were introduced. The interleaved commits from another session improved rather than degraded the outcome. This refactoring is ready to ship.

---

## Verification Attestation

| Artifact | Path | Verified |
|----------|------|----------|
| Plan | `/Users/tomtenuta/Code/knossos/docs/hygiene/PLAN-distribution-readiness.md` | Read, all 18 WUs reviewed |
| fileutil.go | `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil.go` | Read, verified CreateTemp + Sync + Rename + parent dirs |
| fileutil_test.go | `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil_test.go` | Read, 7 tests covering all contract points |
| moirai.go | `/Users/tomtenuta/Code/knossos/internal/lock/moirai.go` | Read, verified MoiraiLock type + ReadMoiraiLock + IsMoiraiLockStale |
| moirai_test.go | `/Users/tomtenuta/Code/knossos/internal/lock/moirai_test.go` | Read, 5 tests including empty file and treatLegacyAsStale |
| lock.go | `/Users/tomtenuta/Code/knossos/internal/lock/lock.go` | Read, verified IsStale with treatLegacyAsStale parameter, IsStaleFile convenience function |
| output.go | `/Users/tomtenuta/Code/knossos/internal/hook/output.go` | Read, verified PreToolUseOutput only (Result/PreCompactOutput removed by interleaved commit) |
| precompact.go | `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go` | Read, verified plain JSON output with precompactResult |
| writeguard.go | `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go` | Read, verified lock.ReadMoiraiLock + lock.IsMoiraiLockStale usage |
| recover.go | `/Users/tomtenuta/Code/knossos/internal/cmd/session/recover.go` | Read, verified lock.IsStaleFile(lockPath, true) |
| materialize.go | `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` | Read, verified ritesDir absent, writeIfChanged delegates to fileutil, no dead code |
| frontmatter.go | `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter.go` | Read, verified MenaScope fully removed, Scope field absent from MenaFrontmatter |
| project_mena.go | `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go` | Read, verified warning log at line 169, scopeIncludesPipeline absent |
| context.go | `/Users/tomtenuta/Code/knossos/internal/session/context.go` | Read, verified fileutil.AtomicWriteFile in Save() |
| rotation.go | `/Users/tomtenuta/Code/knossos/internal/session/rotation.go` | Read, verified local atomicWriteFile deleted |
| backup.go | `/Users/tomtenuta/Code/knossos/internal/inscription/backup.go` | Read, verified exported AtomicWriteFile deleted |
| usersync.go | `/Users/tomtenuta/Code/knossos/internal/usersync/usersync.go` | Read, verified no MenaScope references |
| ADR-0026 | `/Users/tomtenuta/Code/knossos/docs/decisions/ADR-0026-unified-provenance.md` | Read, verified coherent content, real type references, correct format |
| Build output | `CGO_ENABLED=0 go build ./...` | Zero errors |
| Test output | `CGO_ENABLED=0 go test ./...` | All packages pass |
| Regression grep | `MenaScope`, `StagedMaterialize`, `cloneDir`, `IsStaleForTest`, `ParseLegacyMarkers` | Zero occurrences in active code |

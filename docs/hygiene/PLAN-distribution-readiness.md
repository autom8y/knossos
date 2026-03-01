# Refactoring Plan: Distribution Readiness

> **Author**: architect-enforcer (hygiene rite)
> **Date**: 2026-02-08
> **Inputs**: SMELL-distribution-readiness.md (22 findings), STAKEHOLDER-PREFERENCES-distribution-readiness.md (Section 10), CODEBASE-CONTEXT.md
> **Downstream**: Janitor agent executes this plan literally.
> **Stakeholder Review**: 2026-02-08 — 8 clarifications resolved (see Stakeholder Amendments below).

---

## Architectural Assessment

The codebase is structurally sound. The 22 smells cluster into four root causes:

1. **Incomplete cleanup**: StagedMaterialize, Materialize wrapper, dead helpers, and legacy templates were deprecated but never deleted. (SMELL-001, 006, 007, 008, 009, 016, 017)
2. **Missing shared utilities**: AtomicWriteFile duplicated 3x, lock reading duplicated 2x, stale detection duplicated 2x. (SMELL-003, 004, 005, 012)
3. **Format inconsistency**: Hook output uses three contradictory patterns. (SMELL-002)
4. **Premature infrastructure**: Scope filtering built but unused. (SMELL-013, 022)

No boundary violations exist. All proposed changes strengthen encapsulation (shared utilities in `internal/lock/` and `internal/fileutil/`) or reduce API surface (dead code removal). No public API contracts change -- all removed symbols are either unexported, have zero external callers, or are explicitly deprecated.

---

## Dependency Graph (DAG)

```
Phase 1 (Independent -- no dependencies):
  WU-001: Delete StagedMaterialize + cloneDir
  WU-002: Commit untracked files (precompact, rotation, hooks.yaml)
  WU-006: Delete dead code (materializeSettings, getCurrentRite, GetTemplatesDir)
  WU-007: Delete Materialize() wrapper
  WU-011: Delete legacy template files
  WU-012: Delete ParseLegacyMarkers

Phase 2 (Foundation packages):
  WU-003: Consolidate lock reading into internal/lock/       [no deps]
  WU-004: Create internal/fileutil/ package                   [no deps]

Phase 3 (Consumers of Phase 2):
  WU-005: Rewrite writeguard.go lock check    [depends on WU-003]
  WU-008: Rewrite recover.go stale check      [depends on WU-003]
  WU-009: Replace atomicWriteFile callers      [depends on WU-004]
  WU-010: Atomic write for Context.Save()      [depends on WU-004]

Phase 4 (Remaining P2 work):
  WU-013: Migrate precompact to CC-native output + deprecate Result  [depends on WU-002]
  WU-014: Remove scope infrastructure
  WU-016: Complete ritesDir -> sourceResolver migration
  WU-017: Add warning log on frontmatter parse failure
  WU-018: Extract provenance design to ADR-024                      [no deps]

DEFERRED (stakeholder decision):
  WU-015: Phase-specific error codes -- DEFERRED to future sprint
```

### Rollback Points

- **After Phase 1**: All dead code removed. Safe checkpoint. Revert = restore individual files.
- **After Phase 2**: Foundation packages created. Safe checkpoint. Reverting WU-003 or WU-004 requires also reverting their Phase 3 consumers.
- **After Phase 3**: All duplication consolidated. Full test pass required before proceeding.
- **After Phase 4**: Remaining hygiene. Each unit independently revertable.

---

## Execution Order

| Order | WU | Phase | Addresses | Priority | Risk | Commit Batch |
|-------|-----|-------|-----------|----------|------|-------------|
| 1 | WU-001 | 1 | SMELL-001 | P0 | Low | Batch A: dead code |
| 2 | WU-006 | 1 | SMELL-006, 007, 009 | P1 | Low | Batch A: dead code |
| 3 | WU-007 | 1 | SMELL-008 | P1 | Low | Batch A: dead code |
| 4 | WU-011 | 1 | SMELL-016 | P2 | Low | Batch A: dead code |
| 5 | WU-012 | 1 | SMELL-017 | P2 | Low | Batch A: dead code |
| 6 | WU-002 | 1 | SMELL-010 | P1 | Low | Batch B: commit untracked |
| 7 | WU-003 | 2 | SMELL-003, 004 | P0 | Medium | Commit C: lock consolidation |
| 8 | WU-004 | 2 | SMELL-005 | P1 | Low | Commit D: fileutil |
| 9 | WU-005 | 3 | SMELL-003 | P0 | Medium | Commit E: lock consumers |
| 10 | WU-008 | 3 | SMELL-004 | P1 | Low | Commit E: lock consumers |
| 11 | WU-009 | 3 | SMELL-005 | P1 | Low | Commit F: fileutil consumers |
| 12 | WU-010 | 3 | SMELL-012 | P2 | Low | Commit F: fileutil consumers |
| 13 | WU-013 | 4 | SMELL-002 | P0 | Medium | Commit G: hook format |
| 14 | WU-014 | 4 | SMELL-013, 022 | P2 | Low | Commit H: scope + ritesDir |
| 15 | WU-016 | 4 | SMELL-019 | P2 | Medium | Commit H: scope + ritesDir |
| 16 | WU-017 | 4 | SMELL-020 | P2 | Low | Commit I: warning log |
| 17 | WU-018 | 4 | SMELL-014 | Design | Low | Commit J: ADR-024 |
| -- | WU-015 | -- | SMELL-011 | DEFERRED | -- | -- |

### Commit Batching Strategy (Stakeholder Decision)

Phase 1 items batched into 2 commits instead of 6:
- **Batch A**: WU-001 + WU-006 + WU-007 + WU-011 + WU-012 (all dead code deletion)
- **Batch B**: WU-002 (commit untracked files — separate because it's a feature commit, not cleanup)

Phases 2-4: ~8 commits grouped by theme (lock, fileutil, consumers, hook format, scope/ritesDir, ADR).

**Total: ~10 commits** (down from 17).

---

## Work Units

---

### WU-001: Delete StagedMaterialize and cloneDir

**Addresses**: SMELL-001
**Priority**: P0
**Phase**: 1

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`

**Files Deleted**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/staging_test.go`

#### Before Contract

- `materialize.go:146-210`: Exported method `(m *Materializer) StagedMaterialize(materializeFn func(m *Materializer) (*Result, error)) (*Result, error)` -- 65 lines, deprecated, causes CC file watcher freeze.
- `materialize.go:213-243`: Unexported function `cloneDir(src, dst string) error` -- 31 lines, only called by StagedMaterialize and tests.
- `staging_test.go`: 204 lines containing `TestAtomicWriteFile`, `TestAtomicWriteFile_Overwrites`, `TestWriteIfChanged_SkipsIdentical`, `TestWriteIfChanged_WritesWhenDifferent`, `TestCloneDir`, `TestStagedMaterialize_SwapsDirectories`, `TestStagedMaterialize_RollbackOnError`, `TestStagedMaterialize_NoExistingClaudeDir`.

#### After Contract

- `StagedMaterialize` method: deleted entirely.
- `cloneDir` function: deleted entirely.
- `staging_test.go`: deleted entirely.
- The `Materializer.claudeDirOverride` field remains (used by `getClaudeDir()`, which is also called from `trackState` at line 1269 for the `SetSyncDir` path).

**Important**: `staging_test.go` also contains tests for `atomicWriteFile` and `writeIfChanged` (lines 13-65). These tests exercise package-private functions that are still live code. The janitor must **move** `TestAtomicWriteFile`, `TestAtomicWriteFile_Overwrites`, `TestWriteIfChanged_SkipsIdentical`, and `TestWriteIfChanged_WritesWhenDifferent` into a new or existing test file before deleting `staging_test.go`.

#### Steps

1. Read `staging_test.go` completely.
2. Create `/Users/tomtenuta/Code/knossos/internal/materialize/write_test.go` containing the four `writeIfChanged`/`atomicWriteFile` tests (lines 13-65 of staging_test.go). Keep identical imports.
3. Delete `TestCloneDir`, `TestStagedMaterialize_SwapsDirectories`, `TestStagedMaterialize_RollbackOnError`, `TestStagedMaterialize_NoExistingClaudeDir` by deleting `staging_test.go`.
4. In `materialize.go`, delete lines 146-243 (`StagedMaterialize` + `cloneDir`).
5. Run `CGO_ENABLED=0 go test ./internal/materialize/...`.
6. Verify the four write tests still pass in `write_test.go`.

#### Test Impact

- **Tests to delete**: `TestCloneDir`, `TestStagedMaterialize_SwapsDirectories`, `TestStagedMaterialize_RollbackOnError`, `TestStagedMaterialize_NoExistingClaudeDir` (4 tests)
- **Tests to move**: `TestAtomicWriteFile`, `TestAtomicWriteFile_Overwrites`, `TestWriteIfChanged_SkipsIdentical`, `TestWriteIfChanged_WritesWhenDifferent` (4 tests, from staging_test.go to write_test.go)
- **Tests to add**: None
- **Verify**: `CGO_ENABLED=0 go test ./internal/materialize/...`

#### Risk Assessment

- **What could break**: Nothing. StagedMaterialize has zero production callers. The `claudeDirOverride` field is still used by `trackState()` for staged materialization bookkeeping, but `trackState` is called from `MaterializeWithOptions`, not from `StagedMaterialize`.
- **How to verify**: `CGO_ENABLED=0 go test ./...` -- full suite pass.
- **Rollback**: Single commit revert.

---

### WU-002: Commit Untracked Rotation and Precompact Files

**Addresses**: SMELL-010
**Priority**: P1
**Phase**: 1

**Files Committed (already exist, currently untracked/modified)**:
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go`
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact_test.go`
- `/Users/tomtenuta/Code/knossos/internal/session/rotation.go`
- `/Users/tomtenuta/Code/knossos/internal/session/rotation_test.go`
- `/Users/tomtenuta/Code/knossos/hooks/hooks.yaml`

#### Before Contract

- These files exist on disk but are untracked (`??`) or modified (` M`) per `git status`.
- `rotation.go` (236 lines): Fully implemented `RotateSessionContext()` with 7 test cases.
- `precompact.go` (128 lines): Hook that calls `RotateSessionContext` on PreCompact events.
- `hooks.yaml`: Updated to register the precompact hook.

#### After Contract

- Identical files, now tracked in git. No code changes.

#### Steps

1. Run `CGO_ENABLED=0 go test ./internal/session/... ./internal/cmd/hook/...` to verify existing code compiles and tests pass.
2. `git add` the five files listed above.
3. Commit with message: `feat(session): wire rotation and precompact hook for SESSION_CONTEXT management`

#### Test Impact

- **Tests to delete**: None
- **Tests to modify**: None
- **Tests to add**: None (tests already exist in the untracked files)
- **Verify**: `CGO_ENABLED=0 go test ./internal/session/... ./internal/cmd/hook/...`

#### Risk Assessment

- **What could break**: Nothing. This is a commit of existing working code.
- **Rollback**: `git revert` the commit.

---

### WU-003: Consolidate Lock Reading and Stale Detection into internal/lock/

**Addresses**: SMELL-003, SMELL-004
**Priority**: P0
**Phase**: 2

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/lock/lock.go`

**Files Created**:
- `/Users/tomtenuta/Code/knossos/internal/lock/moirai.go`
- `/Users/tomtenuta/Code/knossos/internal/lock/moirai_test.go`

#### Before Contract

**Moirai lock reading** -- two independent implementations:

1. `internal/cmd/hook/writeguard.go:141-190` -- `isMoiraiLockHeld(projectDir string) bool`:
   - Reads `.sos/sessions/{id}/.moirai-lock`
   - Parses with anonymous struct: `Agent string`, `AcquiredAt string`, `StaleAfterSeconds int`
   - Time parse: `time.Parse(time.RFC3339, lock.AcquiredAt)`
   - Stale check: `acquiredAt.Add(Duration(StaleAfterSeconds) * Second)` vs `time.Now().UTC()`
   - Returns `false` on any error (fail-closed)

2. `internal/cmd/session/lock.go:20-26, 222-242` -- `MoiraiLock` struct + `readMoiraiLock(lockPath string) (*MoiraiLock, error)` + `isLockStale(lock *MoiraiLock) bool`:
   - Named struct with `AcquiredAt time.Time` (relies on json.Unmarshal parsing time.Time)
   - Stale check: `time.Since(lock.AcquiredAt) > Duration(StaleAfterSeconds) * Second`
   - Returns error to caller

**Advisory lock stale detection** -- two implementations:

1. `internal/lock/lock.go:159-201` -- `(m *Manager) isStale(lockPath string) bool`:
   - Reads `LockMetadata` (JSON v2 with Unix timestamp `Acquired int64`)
   - Legacy PID: checks process liveness with `Signal(0)`
   - Private method, exposed only via `IsStaleForTest()`

2. `internal/cmd/session/recover.go:126-147` -- `isAdvisoryLockStale(lockPath string) bool`:
   - Same algorithm but treats ALL legacy PID locks as stale (intentional for recovery)
   - Uses `lock.LockMetadata` and `lock.StaleThreshold` from `internal/lock/`

#### After Contract

New file `internal/lock/moirai.go` exports two functions and one type:

```go
// MoiraiLock represents the Moirai agent's lock file structure.
// This is distinct from LockMetadata (advisory session locks).
type MoiraiLock struct {
    Agent             string    `json:"agent"`
    AcquiredAt        time.Time `json:"acquired_at"`
    SessionID         string    `json:"session_id"`
    StaleAfterSeconds int       `json:"stale_after_seconds"`
}

// ReadMoiraiLock reads and parses a Moirai lock file at the given path.
// Returns the parsed lock or an error if the file cannot be read or parsed.
func ReadMoiraiLock(lockPath string) (*MoiraiLock, error)

// IsMoiraiLockStale returns true if the lock has exceeded its stale threshold.
func IsMoiraiLockStale(lock *MoiraiLock) bool
```

Existing `isStale` on Manager becomes exported with a parameter to control legacy behavior:

```go
// IsStale checks if an advisory lock file should be considered stale.
// When treatLegacyAsStale is true, all legacy PID-format locks are
// treated as stale (suitable for recovery operations).
// When false, legacy locks are checked for process liveness.
func (m *Manager) IsStale(lockPath string, treatLegacyAsStale bool) bool
```

`IsStaleForTest` is **deleted** -- its only purpose was to expose `isStale`, which is now exported directly.

#### Steps

1. **Move** `MoiraiLock` struct, `readMoiraiLock()`, and `isLockStale()` from `internal/cmd/session/lock.go` to new file `internal/lock/moirai.go`:
   - Move `MoiraiLock` struct (lines 20-26) — export as-is (already has exported fields)
   - Move `readMoiraiLock()` → export as `ReadMoiraiLock(lockPath string) (*MoiraiLock, error)`
   - Move `isLockStale()` → export as `IsMoiraiLockStale(lock *MoiraiLock) bool`
   - Update all callers in `internal/cmd/session/lock.go` to import from `internal/lock/`
   - **Stakeholder decision**: Move, not copy. Do NOT create two MoiraiLock types.

2. In `/Users/tomtenuta/Code/knossos/internal/lock/lock.go`:
   - Rename `isStale` to `IsStale` and add `treatLegacyAsStale bool` parameter.
   - When `treatLegacyAsStale` is true AND the content is a legacy PID format, return `true` immediately (skip process liveness check).
   - **Empty file handling (stakeholder decision)**: `IsStale` returns `true` for empty lock files regardless of `treatLegacyAsStale` flag. Empty lock = stale in all contexts.
   - Update the single call site at line 118 (`m.isStale(lockPath)`) to `m.IsStale(lockPath, false)`.
   - Delete `IsStaleForTest()` (lines 283-288).

3. Create `/Users/tomtenuta/Code/knossos/internal/lock/moirai_test.go` with tests:
   - `TestReadMoiraiLock` -- valid JSON, invalid JSON, missing file
   - `TestIsMoiraiLockStale` -- fresh lock (not stale), expired lock (stale)
   - `TestIsStale_TreatLegacyAsStale` -- legacy PID returns true when flag is true
   - `TestIsStale_LegacyProcessLiveness` -- legacy PID checks process when flag is false

4. Run `CGO_ENABLED=0 go test ./internal/lock/...`.

**Note**: Do NOT modify `writeguard.go` or `recover.go` in this WU. Those rewrites happen in WU-005 and WU-008 respectively. This WU only creates the shared implementation.

#### Test Impact

- **Tests to delete**: Tests referencing `IsStaleForTest()` -- search for callers and update to use `IsStale(path, false)`.
- **Tests to modify**: Any test calling `m.IsStaleForTest(path)` becomes `m.IsStale(path, false)`.
- **Tests to add**: `TestReadMoiraiLock`, `TestIsMoiraiLockStale`, `TestIsStale_TreatLegacyAsStale`, `TestIsStale_LegacyProcessLiveness`.
- **Verify**: `CGO_ENABLED=0 go test ./internal/lock/...`

#### Risk Assessment

- **What could break**: The `isStale` -> `IsStale` rename changes the method's visibility. All call sites are within the same package (line 118 of lock.go) so this is safe. `IsStaleForTest` callers must be found and updated.
- **How to verify**: Grep for `IsStaleForTest` across the codebase. Update all callers before committing.
- **Rollback**: Single commit revert. No downstream changes yet (WU-005 and WU-008 are separate).

---

### WU-004: Create internal/fileutil/ Package

**Addresses**: SMELL-005
**Priority**: P1
**Phase**: 2

**Files Created**:
- `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil.go`
- `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil_test.go`

#### Before Contract

Three independent `atomicWriteFile` implementations:

1. `materialize/materialize.go:287-297` -- `atomicWriteFile(path string, content []byte, perm os.FileMode) error`: predictable `.tmp` suffix, no `Sync()`, 3 params.
2. `session/rotation.go:195-236` -- `atomicWriteFile(path string, data []byte) error`: `os.CreateTemp`, `Sync()`, defer cleanup, 2 params. **Most robust**.
3. `inscription/backup.go:349-369` -- `AtomicWriteFile(path string, content []byte) error`: predictable `.tmp` suffix, no `Sync()`, 2 params, creates parent dirs, wraps errors.

Plus `materialize/materialize.go:276-282` -- `writeIfChanged(path string, content []byte, perm os.FileMode) (bool, error)`: reads existing, compares with `bytes.Equal`, calls `atomicWriteFile` if different.

#### After Contract

New package `internal/fileutil/` with two exported functions:

```go
package fileutil

// AtomicWriteFile writes content to path atomically using temp-file-then-rename.
// Uses os.CreateTemp for safe temp file names, calls Sync() before rename,
// and creates parent directories as needed. The file permission is set to perm.
func AtomicWriteFile(path string, content []byte, perm os.FileMode) error

// WriteIfChanged writes content to path only if it differs from the existing file.
// Returns true if a write occurred, false if content was identical.
// Uses AtomicWriteFile for safe writes.
func WriteIfChanged(path string, content []byte, perm os.FileMode) (bool, error)
```

The implementation is based on `session/rotation.go:195-236` (the most robust version) with `perm` parameter added and parent directory creation from `inscription/backup.go`.

#### Steps

1. Create `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil.go` with:
   - `AtomicWriteFile(path string, content []byte, perm os.FileMode) error`:
     - `os.MkdirAll(filepath.Dir(path), 0755)` for parent dirs
     - `os.CreateTemp(dir, base+".tmp.*")` for unique temp name
     - Write content, `Sync()`, close, `os.Chmod(tmpPath, perm)`, `os.Rename(tmpPath, path)`
     - Defer cleanup on error
   - `WriteIfChanged(path string, content []byte, perm os.FileMode) (bool, error)`:
     - Read existing with `os.ReadFile`; if equal via `bytes.Equal`, return `false, nil`
     - Otherwise call `AtomicWriteFile` and return `true, nil`

2. Create `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil_test.go` with:
   - `TestAtomicWriteFile_NewFile` -- writes to nonexistent path
   - `TestAtomicWriteFile_Overwrite` -- overwrites existing file
   - `TestAtomicWriteFile_CreatesParentDirs` -- parent dir does not exist
   - `TestAtomicWriteFile_NoTmpLeftBehind` -- no `.tmp` files remain
   - `TestAtomicWriteFile_PermissionsApplied` -- verifies file mode
   - `TestWriteIfChanged_SkipsIdentical` -- same content returns false
   - `TestWriteIfChanged_WritesDifferent` -- different content returns true

3. Run `CGO_ENABLED=0 go test ./internal/fileutil/...`.

**Note**: Do NOT replace callers in this WU. That happens in WU-009 and WU-010.

#### Test Impact

- **Tests to delete**: None (in this WU)
- **Tests to modify**: None (in this WU)
- **Tests to add**: 7 tests listed above
- **Verify**: `CGO_ENABLED=0 go test ./internal/fileutil/...`

#### Risk Assessment

- **What could break**: Nothing. This is a new package with no consumers yet.
- **Rollback**: Delete the two files.

---

### WU-005: Rewrite writeguard.go to Use Shared Lock Functions

**Addresses**: SMELL-003
**Priority**: P0
**Phase**: 3 (depends on WU-003)

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go`

#### Before Contract

`writeguard.go:141-190` -- `isMoiraiLockHeld(projectDir string) bool`:
- Resolves current session from `.current-session` file
- Reads `.moirai-lock` file
- Parses JSON with anonymous struct (string AcquiredAt)
- Checks agent == "moirai"
- Parses time with `time.Parse(time.RFC3339, ...)`
- Computes stale threshold manually
- Returns false on any error

#### After Contract

`isMoiraiLockHeld(projectDir string) bool`:
- Resolves current session from `.current-session` file (unchanged)
- Calls `lock.ReadMoiraiLock(lockPath)` instead of inline JSON parse
- Checks `lock.Agent != "moirai"` (unchanged semantics)
- Calls `lock.IsMoiraiLockStale(lock)` instead of inline stale check
- Returns false on any error (unchanged semantics)
- Adds import `"github.com/autom8y/knossos/internal/lock"`

The function still resolves the session path and constructs `lockPath` -- only the JSON parsing and stale detection are delegated.

#### Steps

1. Add `"github.com/autom8y/knossos/internal/lock"` to the imports in `writeguard.go`.
2. Replace lines 159-189 (from `lockData, err := os.ReadFile(lockPath)` through `return true`) with:
   ```go
   // Parse lock file using shared implementation
   moiraiLock, err := lock.ReadMoiraiLock(lockPath)
   if err != nil {
       return false
   }

   // Verify agent is moirai
   if moiraiLock.Agent != "moirai" {
       return false
   }

   // Check if stale
   if lock.IsMoiraiLockStale(moiraiLock) {
       return false
   }

   return true
   ```
3. Remove the `"encoding/json"` and `"time"` imports if no longer needed by other code in the file. (Check before removing -- `time` may be used elsewhere in writeguard.go.)
4. Run `CGO_ENABLED=0 go test ./internal/cmd/hook/...`.

#### Invariants

- **Same fail-closed behavior**: Any error reading or parsing the lock returns `false` (deny).
- **Same agent check**: Only `"moirai"` is accepted.
- **Same stale semantics**: Lock is stale when `AcquiredAt + StaleAfterSeconds < now`. The shared implementation uses `time.Since(lock.AcquiredAt)` which is equivalent to the writeguard's `time.Now().UTC().After(staleThreshold)`.

#### Test Impact

- **Tests to delete**: None
- **Tests to modify**: None (behavior unchanged, tests should pass as-is)
- **Tests to add**: None (covered by WU-003 tests and existing writeguard tests)
- **Verify**: `CGO_ENABLED=0 go test ./internal/cmd/hook/...`

#### Risk Assessment

- **What could break**: Time parsing difference. The old code parsed `AcquiredAt` as a string with `time.Parse(time.RFC3339, ...)`. The new code uses `time.Time` with `json.Unmarshal`, which also expects RFC3339. If the lock file contains a non-RFC3339 time string that `time.Parse` could handle but `json.Unmarshal` cannot, behavior would change. **Mitigation**: Verify that Moirai writes RFC3339 timestamps (it does -- `time.Time` marshals to RFC3339 in Go's `encoding/json`).
- **How to verify**: Run existing writeguard tests. They exercise the lock-checking path.
- **Rollback**: Single commit revert.

---

### WU-006: Delete Dead Code (materializeSettings, getCurrentRite, GetTemplatesDir)

**Addresses**: SMELL-006, SMELL-007, SMELL-009
**Priority**: P1
**Phase**: 1

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`
- `/Users/tomtenuta/Code/knossos/internal/materialize/source.go`

#### Before Contract

1. `materialize.go:1228-1230` -- `(m *Materializer) materializeSettings(claudeDir string) error`: unexported one-line wrapper calling `materializeSettingsWithManifest(claudeDir, nil)`. Zero callers.
2. `materialize.go:1329-1336` -- `(m *Materializer) getCurrentRite(claudeDir string) (string, error)`: reads ACTIVE_RITE file. Zero callers.
3. `source.go:379-398` -- `(r *SourceResolver) GetTemplatesDir(source RiteSource) string`: exported method. Zero callers across entire codebase.

#### After Contract

All three functions deleted. No replacement needed.

#### Steps

1. Delete `materializeSettings` (lines 1227-1230 in materialize.go, including the comment on line 1227).
2. Delete `getCurrentRite` (lines 1328-1336 in materialize.go, including the comment on line 1328).
3. Delete `GetTemplatesDir` (lines 379-398 in source.go, including the comment on line 379).
4. Run `CGO_ENABLED=0 go test ./internal/materialize/...`.

#### Test Impact

- **Tests to delete**: None (no tests exist for these functions)
- **Tests to modify**: None
- **Tests to add**: None
- **Verify**: `CGO_ENABLED=0 go test ./internal/materialize/...`

#### Risk Assessment

- **What could break**: Nothing. Zero callers confirmed by grep.
- **Rollback**: Single commit revert.

---

### WU-007: Delete Legacy Materialize() Wrapper

**Addresses**: SMELL-008
**Priority**: P1
**Phase**: 1

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`

#### Before Contract

`materialize.go:321-326`:
```go
// Materialize generates the .claude/ directory from templates and the active rite.
// This is the legacy method that uses default options (keep orphans).
func (m *Materializer) Materialize(activeRiteName string) error {
    _, err := m.MaterializeWithOptions(activeRiteName, Options{KeepAll: true})
    return err
}
```

Exported one-line wrapper. Stakeholder confirmed no external consumers.

#### After Contract

Function deleted. All callers should use `MaterializeWithOptions` directly.

#### Steps

1. Grep for `\.Materialize(` (with the dot prefix to distinguish from `MaterializeWithOptions`, `MaterializeMinimal`, etc.) across the entire codebase to confirm zero callers outside tests.
2. If any test callers exist, update them to use `MaterializeWithOptions(riteName, Options{KeepAll: true})`.
3. Delete lines 321-326 from `materialize.go`.
4. Run `CGO_ENABLED=0 go test ./internal/materialize/...`.

#### Test Impact

- **Tests to delete**: None expected (but verify with grep)
- **Tests to modify**: Any tests calling `m.Materialize(name)` become `m.MaterializeWithOptions(name, Options{KeepAll: true})`
- **Tests to add**: None
- **Verify**: `CGO_ENABLED=0 go test ./...` (full suite since this is an exported symbol)

#### Risk Assessment

- **What could break**: Any caller using `Materialize()` will get a compile error. This is the desired behavior -- compile-time detection, not runtime surprise.
- **How to verify**: Full `go build ./...` followed by `go test ./...`.
- **Rollback**: Single commit revert.

---

### WU-008: Rewrite recover.go to Use Shared IsStale

**Addresses**: SMELL-004
**Priority**: P1
**Phase**: 3 (depends on WU-003)

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/cmd/session/recover.go`

#### Before Contract

`recover.go:126-147` -- `isAdvisoryLockStale(lockPath string) bool`:
- Reads lock file, tries JSON parse with `lock.LockMetadata`
- Age check: `time.Since(acquired) > lock.StaleThreshold`
- Legacy PID: returns `true` unconditionally (treats all legacy as stale for recovery)

#### After Contract

Function body replaced with a single call:

```go
func isAdvisoryLockStale(lockPath string) bool {
    return lockMgr.IsStale(lockPath, true) // treatLegacyAsStale=true for recovery
}
```

Where `lockMgr` is a `*lock.Manager` constructed with the appropriate locks directory. The janitor must determine how to obtain the Manager instance -- either construct one inline or pass it through the recovery function's call chain.

**Alternative** (if Manager construction is awkward): Add a standalone function to `internal/lock/`:

```go
// IsStaleFile checks if a lock file at the given path is stale.
// This is a convenience function that does not require a Manager instance.
func IsStaleFile(lockPath string, treatLegacyAsStale bool) bool
```

This extracts the logic from `Manager.IsStale` into a standalone function that the Manager also calls. The janitor should choose whichever approach is simpler.

#### Steps

1. Determine whether `recover.go` has access to a `lock.Manager` instance or can construct one.
2. If yes: replace `isAdvisoryLockStale` body with `lockMgr.IsStale(lockPath, true)`.
3. If no: add `IsStaleFile(lockPath string, treatLegacyAsStale bool) bool` to `internal/lock/lock.go` (extract from Manager.IsStale), then call `lock.IsStaleFile(lockPath, true)` from recover.go.
4. Remove the `encoding/json`, `strings`, and `time` imports from recover.go if no longer needed.
5. Run `CGO_ENABLED=0 go test ./internal/cmd/session/...`.

#### Invariants

- **Same recovery behavior**: Legacy PID locks are still treated as stale (the `treatLegacyAsStale=true` flag preserves this intentional divergence).
- **Same JSON v2 behavior**: Age check is identical.

#### Test Impact

- **Tests to delete**: None
- **Tests to modify**: None (behavior preserved)
- **Tests to add**: None (covered by WU-003 tests)
- **Verify**: `CGO_ENABLED=0 go test ./internal/cmd/session/...`

#### Risk Assessment

- **What could break**: The `recover.go` version treats empty lock files as stale (`return true`). The `lock.isStale` version treats empty lock files as NOT stale (`return false`). The janitor must verify this edge case is handled. If `IsStale` does not match, add the empty-file check before calling.
- **How to verify**: Read both implementations carefully. The divergence on empty files must be preserved in the new code.
- **Rollback**: Single commit revert.

---

### WU-009: Replace atomicWriteFile Callers with fileutil Package

**Addresses**: SMELL-005
**Priority**: P1
**Phase**: 3 (depends on WU-004)

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`
- `/Users/tomtenuta/Code/knossos/internal/session/rotation.go`
- `/Users/tomtenuta/Code/knossos/internal/inscription/backup.go`

#### Before Contract

Three private/package-level implementations:
1. `materialize.go:287-297`: `atomicWriteFile(path, content, perm)` -- predictable `.tmp` suffix
2. `rotation.go:195-236`: `atomicWriteFile(path, data)` -- CreateTemp, Sync, defer cleanup
3. `inscription/backup.go:349-369`: `AtomicWriteFile(path, content)` -- predictable `.tmp` suffix, creates dirs

And: `materialize.go:276-282`: `writeIfChanged(path, content, perm)` calls local `atomicWriteFile`.

#### After Contract

1. `materialize.go`: Delete local `atomicWriteFile`. Replace `writeIfChanged` to call `fileutil.AtomicWriteFile`. Or delete `writeIfChanged` and replace with `fileutil.WriteIfChanged`.
2. `rotation.go`: Delete local `atomicWriteFile`. Call `fileutil.AtomicWriteFile(path, data, 0644)`.
3. `inscription/backup.go`: Delete `AtomicWriteFile`. Call `fileutil.AtomicWriteFile(path, content, 0644)`.

All three packages add import `"github.com/autom8y/knossos/internal/fileutil"`.

#### Steps

1. In `materialize.go`:
   - Add `"github.com/autom8y/knossos/internal/fileutil"` import.
   - Replace `writeIfChanged` body to call `fileutil.WriteIfChanged(path, content, perm)`.
   - Delete local `atomicWriteFile` function (lines 284-297).
2. In `rotation.go`:
   - Add `"github.com/autom8y/knossos/internal/fileutil"` import.
   - Replace call `atomicWriteFile(contextPath, newContent)` with `fileutil.AtomicWriteFile(contextPath, newContent, 0644)`.
   - Replace call `atomicWriteFile(archivePath, archiveContent)` with `fileutil.AtomicWriteFile(archivePath, archiveContent, 0644)`.
   - Delete local `atomicWriteFile` function (lines 194-236).
3. In `inscription/backup.go`:
   - Add `"github.com/autom8y/knossos/internal/fileutil"` import.
   - Find all callers of `AtomicWriteFile` within the inscription package. Replace them with `fileutil.AtomicWriteFile(path, content, 0644)`.
   - Delete the local `AtomicWriteFile` function (lines 347-369).
   - If `inscription.AtomicWriteFile` is exported and has external callers, verify none exist (grep confirmed zero external callers).
4. Update tests in `materialize/write_test.go` (created in WU-001) to test `fileutil.WriteIfChanged` behavior through the materialize package's `writeIfChanged` wrapper.
5. Run `CGO_ENABLED=0 go test ./internal/materialize/... ./internal/session/... ./internal/inscription/... ./internal/fileutil/...`.

#### Invariants

- **Same write semantics**: Files are written atomically via temp-file-then-rename.
- **Improved safety**: All callers now get `Sync()` and unique temp filenames (from the rotation.go implementation).
- **Same writeIfChanged semantics**: Content comparison before write is preserved.

#### Test Impact

- **Tests to delete**: `TestAtomicWriteFile` and `TestAtomicWriteFile_Overwrites` from `write_test.go` (now covered by `fileutil_test.go`).
- **Tests to modify**: `TestWriteIfChanged_SkipsIdentical` and `TestWriteIfChanged_WritesDifferent` -- keep as integration tests of materialize's writeIfChanged calling fileutil.
- **Tests to add**: None beyond WU-004's tests.
- **Verify**: `CGO_ENABLED=0 go test ./...`

#### Risk Assessment

- **What could break**: The materialize.go `atomicWriteFile` took a `perm os.FileMode` parameter while rotation.go did not. The fileutil version takes `perm`, so callers in rotation.go must pass `0644` explicitly. If rotation.go's original implementation used a different default permission, verify.
- **Dependency graph**: `materialize` -> `fileutil` is safe. `session` -> `fileutil` is safe. `inscription` -> `fileutil` is safe. No circular imports possible (fileutil has no knossos dependencies).
- **Rollback**: Single commit revert. All three packages fall back to their local implementations.

---

### WU-010: Atomic Write for Context.Save()

**Addresses**: SMELL-012
**Priority**: P2
**Phase**: 3 (depends on WU-004)

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/session/context.go`

#### Before Contract

`context.go:207-216`:
```go
func (c *Context) Save(path string) error {
    data, err := c.Serialize()
    if err != nil {
        return err
    }
    if err := os.WriteFile(path, data, 0644); err != nil {
        return errors.Wrap(errors.CodeGeneralError, "failed to write session context", err)
    }
    return nil
}
```

Direct `os.WriteFile` -- partial write visible to CC file watcher on crash/interrupt.

#### After Contract

```go
func (c *Context) Save(path string) error {
    data, err := c.Serialize()
    if err != nil {
        return err
    }
    if err := fileutil.AtomicWriteFile(path, data, 0644); err != nil {
        return errors.Wrap(errors.CodeGeneralError, "failed to write session context", err)
    }
    return nil
}
```

Adds import `"github.com/autom8y/knossos/internal/fileutil"`.

#### Steps

1. Add `"github.com/autom8y/knossos/internal/fileutil"` to imports in `context.go`.
2. Replace `os.WriteFile(path, data, 0644)` with `fileutil.AtomicWriteFile(path, data, 0644)`.
3. Run `CGO_ENABLED=0 go test ./internal/session/...`.

#### Test Impact

- **Tests to delete**: None
- **Tests to modify**: None (behavior unchanged from caller's perspective)
- **Tests to add**: None (atomic write is tested in fileutil_test.go)
- **Verify**: `CGO_ENABLED=0 go test ./internal/session/...`

#### Risk Assessment

- **What could break**: `fileutil.AtomicWriteFile` creates parent directories. `os.WriteFile` does not. If `Save` is ever called with a path whose parent does not exist and the caller expects an error, the behavior changes. **Mitigation**: `Save` is always called with session directories that already exist.
- **Rollback**: Single commit revert.

---

### WU-011: Delete Legacy Template Files

**Addresses**: SMELL-016
**Priority**: P2
**Phase**: 1

**Files Deleted**:
- `/Users/tomtenuta/Code/knossos/templates/base-orchestrator.md`
- `/Users/tomtenuta/Code/knossos/templates/orchestrator-base.md.tpl`

#### Before Contract

- `templates/base-orchestrator.md` (164 lines): Legacy orchestrator template with `{{TEAM_DESCRIPTION}}` placeholders.
- `templates/orchestrator-base.md.tpl` (49 lines): Different legacy template format.
- Neither file is referenced by any Go code. Active templates live in `knossos/templates/`.

#### After Contract

Both files deleted. No replacement needed.

#### Steps

1. Verify no callers: grep for `base-orchestrator` and `orchestrator-base` across all `.go` and `.md` files.
2. Delete both files.
3. Run `CGO_ENABLED=0 go test ./...` (sanity check).

#### Test Impact

- None.

#### Risk Assessment

- **What could break**: Nothing. Zero references confirmed.
- **Rollback**: `git checkout` the two files.

---

### WU-012: Delete ParseLegacyMarkers and Related Code

**Addresses**: SMELL-017
**Priority**: P2
**Phase**: 1

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/inscription/marker.go`
- `/Users/tomtenuta/Code/knossos/internal/inscription/marker_test.go`
- `/Users/tomtenuta/Code/knossos/internal/inscription/types.go`

#### Before Contract

- `marker.go:248-298`: `ParseLegacyMarkers(content string) []*LegacyMarker` -- exported, only called in tests.
- `marker.go:300-324`: `suggestRegionName(content string, lineNum int) string` -- private helper, only called by ParseLegacyMarkers.
- `marker.go:33-37`: `legacyPreserveRegex` and `legacySyncRegex` -- package-level vars, only used by ParseLegacyMarkers.
- `marker_test.go:318-352`: `TestMarkerParser_ParseLegacyMarkers` and `TestMarkerParser_ParseLegacyMarkers_InCodeBlock`.
- `types.go:314-331`: `LegacyMarkerType` type, `LegacyPreserve`/`LegacySync` constants, `LegacyMarker` struct.

#### After Contract

All of the above deleted. The inscription pipeline uses KNOSSOS markers exclusively. The migration from PRESERVE/SYNC markers is complete.

#### Steps

1. Delete `legacyPreserveRegex` and `legacySyncRegex` from `marker.go` (lines 33-37).
2. Delete `ParseLegacyMarkers` from `marker.go` (lines 248-298).
3. Delete `suggestRegionName` from `marker.go` (lines 300-end of function).
4. Delete `LegacyMarkerType`, `LegacyPreserve`, `LegacySync`, `LegacyMarker` from `types.go` (lines 314-331).
5. Delete `TestMarkerParser_ParseLegacyMarkers` and `TestMarkerParser_ParseLegacyMarkers_InCodeBlock` from `marker_test.go` (lines 318-352).
6. Remove unused imports from all modified files.
7. Run `CGO_ENABLED=0 go test ./internal/inscription/...`.

#### Test Impact

- **Tests to delete**: `TestMarkerParser_ParseLegacyMarkers`, `TestMarkerParser_ParseLegacyMarkers_InCodeBlock` (2 tests)
- **Tests to modify**: None
- **Tests to add**: None
- **Verify**: `CGO_ENABLED=0 go test ./internal/inscription/...`

#### Risk Assessment

- **What could break**: If any code outside the test files calls `ParseLegacyMarkers` or uses `LegacyMarker` types. Grep confirmed zero non-test callers.
- **Rollback**: Single commit revert.

---

### WU-013: Standardize Precompact on CC-Native Output + Deprecate Result

**Addresses**: SMELL-002
**Priority**: P0
**Phase**: 4 (depends on WU-002 -- precompact must be committed first)

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go`
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact_test.go`
- `/Users/tomtenuta/Code/knossos/internal/hook/output.go`

#### Before Contract

**precompact.go:16-21** -- custom flat struct:
```go
type PrecompactDecision struct {
    Decision           string `json:"decision"`
    PermissionDecision string `json:"permissionDecision"`
    Reason             string `json:"reason,omitempty"`
}
```

**precompact.go:111-118** -- `outputAllowPrecompact` creates this custom struct.

**output.go:34-60** -- `Result` struct: legacy dual-format with `Decision` + auto-populated `PermissionDecision`. Still exported, still has helper functions (`WriteAllow`, `WriteBlock`, etc.), but zero production callers (writeguard and validate use `PreToolUseOutput`).

Note: PreCompact is NOT a PreToolUse event, so it does not strictly need `hookSpecificOutput`. However, the stakeholder decision is to standardize all hooks on the CC-native format for consistency. For non-PreToolUse hooks, the output should still use a consistent envelope structure even though CC may not enforce it.

#### After Contract

**precompact.go**:
- Delete `PrecompactDecision` struct.
- `outputAllowPrecompact` uses `hook.PreToolUseOutput` (or a new `hook.PreCompactOutput` -- see design note below) with the same allow semantics.

**Stakeholder Decision**: Create a **separate PreCompactOutput type** in `hook/output.go`. CC models each event with its own Input/Output pair (PreToolUseInput/PreToolUseOutput, PreCompactInput/PreCompactOutput, etc.). Reusing PreToolUseOutput would carry semantically meaningless fields (permissionDecision, updatedInput) for a non-tool event.

New type in `output.go`:
```go
// PreCompactOutput is the CC-native output envelope for PreCompact hooks.
// PreCompact fires before context compaction. Unlike PreToolUse, it has no
// permission decision semantics — it is a side-effect hook (e.g., rotation).
type PreCompactOutput struct {
    HookSpecificOutput PreCompactHookOutput `json:"hookSpecificOutput"`
}

type PreCompactHookOutput struct {
    HookEventName string `json:"hookEventName"` // Always "PreCompact"
    Decision      string `json:"decision"`       // "allow" (always, informational)
    Reason        string `json:"reason,omitempty"`
}
```

**output.go**:
- Add deprecation comment to `Result` struct: `// Deprecated: Use PreToolUseOutput for PreToolUse hooks. For other hook types, use HookSpecificOutput directly. This type produces flat JSON without the hookSpecificOutput envelope that CC expects.`
- Add deprecation comments to `WriteAllow`, `WriteBlock`, `WriteModify`, `WriteError`, `WriteContext`, `Allow`, `Block`, `Modify`, `WithContext`, `WithDuration` helper functions.
- Do NOT delete `Result` -- other hooks or tests may still reference it. Deprecation is sufficient for this sprint.

#### Steps

1. In `precompact.go`, delete `PrecompactDecision` struct (lines 16-21).
2. Rewrite `outputAllowPrecompact` to use the new `hook.PreCompactOutput`:
   ```go
   func outputAllowPrecompact(printer *output.Printer, reason string) error {
       result := hook.PreCompactOutput{
           HookSpecificOutput: hook.PreCompactHookOutput{
               HookEventName: "PreCompact",
               Decision:      "allow",
               Reason:        reason,
           },
       }
       return printer.Print(result)
   }
   ```
3. Update `precompact_test.go`: test assertions that check for `PrecompactDecision` fields must now check for `hookSpecificOutput` envelope.
4. In `output.go`, add `// Deprecated:` comments to `Result` struct and all its helper functions/methods.
5. Run `CGO_ENABLED=0 go test ./internal/cmd/hook/... ./internal/hook/...`.

#### Invariants

- **Precompact always returns allow**: This hook is a side-effect hook (rotation), never blocks. The `permissionDecision: "allow"` is preserved.
- **Reason field preserved**: The rotation reason string is still conveyed via `PermissionDecisionReason`.

#### Test Impact

- **Tests to delete**: None
- **Tests to modify**: All precompact tests that assert on `PrecompactDecision` struct shape
- **Tests to add**: None
- **Verify**: `CGO_ENABLED=0 go test ./internal/cmd/hook/...`

#### Risk Assessment

- **What could break**: If any code parses precompact's stdout and expects the flat `{"decision":"allow"}` format, it will break. **Mitigation**: Precompact's output is consumed by CC, which reads `hookSpecificOutput`. No other code parses it.
- **Rollback**: Single commit revert.

---

### WU-014: Remove Scope Infrastructure

**Addresses**: SMELL-013, SMELL-022
**Priority**: P2
**Phase**: 4

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter.go`
- `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter_test.go`
- `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go`
- `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena_test.go`

#### Before Contract

- `frontmatter.go:47-79`: `MenaScope` type, `MenaScopeBoth`/`MenaScopeUser`/`MenaScopeProject` constants, `ValidScope()` method, `String()` method.
- `frontmatter.go:89`: `Scope MenaScope` field on `MenaFrontmatter` struct.
- `project_mena.go:112-124`: `scopeIncludesPipeline(entryScope, pipelineScope MenaScope) bool`.
- `frontmatter_test.go:142-208`: `TestMenaScope_ValidScope`, `TestMenaScope_String`, scope-related test cases.
- `project_mena_test.go`: Tests referencing `MenaScope`, `scopeIncludesPipeline`, or `PipelineScope`.

Total: ~60 lines of production code, ~66 lines of tests. Zero mena files use `scope:`.

#### After Contract

- `MenaScope` type: deleted.
- Constants `MenaScopeBoth`, `MenaScopeUser`, `MenaScopeProject`: deleted.
- `ValidScope()` method: deleted.
- `String()` method on MenaScope: deleted.
- `Scope` field on `MenaFrontmatter`: deleted.
- `scopeIncludesPipeline` function: deleted.
- `PipelineScope` field on `MenaProjectionOptions` (if it exists): deleted.
- All callers of `scopeIncludesPipeline` in `project_mena.go`: simplified to unconditionally include entries (since scope was always empty/"both").

#### Steps

1. Delete `MenaScope` type, constants, `ValidScope`, and `String` from `frontmatter.go` (lines 47-79).
2. Remove `Scope MenaScope` field from `MenaFrontmatter` struct (line 89).
3. Delete `scopeIncludesPipeline` from `project_mena.go` (lines 112-124).
4. Find all call sites of `scopeIncludesPipeline` in `project_mena.go` and remove the filtering. Where the code does `if !scopeIncludesPipeline(fm.Scope, opts.PipelineScope) { continue }`, delete the conditional.
5. If `MenaProjectionOptions` has a `PipelineScope MenaScope` field, remove it.
6. Delete scope-related tests from `frontmatter_test.go` and `project_mena_test.go`.
7. Run `CGO_ENABLED=0 go test ./internal/materialize/...`.

#### Test Impact

- **Tests to delete**: `TestMenaScope_ValidScope`, `TestMenaScope_String`, any test cases exercising scope filtering
- **Tests to modify**: Any test creating `MenaFrontmatter` with `Scope` field or `MenaProjectionOptions` with `PipelineScope`
- **Tests to add**: None
- **Verify**: `CGO_ENABLED=0 go test ./internal/materialize/...`

#### Risk Assessment

- **What could break**: If any code outside `internal/materialize/` references `MenaScope`, `MenaScopeBoth`, etc. Grep the full codebase before deleting.
- **Rollback**: Single commit revert.

---

### WU-015: DEFERRED — Phase-Specific Error Codes

**Addresses**: SMELL-011
**Priority**: P2
**Status**: **DEFERRED** (stakeholder decision 2026-02-08)

**Rationale**: Error code differentiation is nice-to-have. Janitor time should focus on P0/P1 items. Defer to a future sprint.

---

### WU-018: Extract Provenance Design to ADR-024

**Addresses**: SMELL-014
**Priority**: Design artifact
**Phase**: 4

**Files Created**:
- `/Users/tomtenuta/Code/knossos/docs/decisions/ADR-024-unified-provenance.md`

#### Before Contract

The provenance design exists inline in this plan document (see "Provenance Design" section below). It is not discoverable as a decision record.

#### After Contract

The provenance design is extracted to `docs/decisions/ADR-024-unified-provenance.md` following the existing ADR format in that directory. The inline section in this plan document is replaced with a reference to the ADR.

#### Steps

1. Read an existing ADR in `docs/decisions/` to understand the format.
2. Create `ADR-024-unified-provenance.md` with the content from the Provenance Design section.
3. Replace the inline Provenance Design section in this plan with: `See docs/decisions/ADR-024-unified-provenance.md`

#### Test Impact

- None (documentation only).

#### Risk Assessment

- None.

---

### WU-016: Complete ritesDir to sourceResolver Migration

**Addresses**: SMELL-019
**Priority**: P2
**Phase**: 4

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`

#### Before Contract

`materialize.go:87`:
```go
ritesDir      string // Deprecated: use sourceResolver
```

But `ritesDir` is still used at lines 762-774 in `materializeMena()`:
```go
sharedMenaDir := filepath.Join(m.ritesDir, "shared", "mena")
sources = append(sources, MenaSource{Path: filepath.Join(m.ritesDir, dep, "mena")})
currentRiteMenaDir := filepath.Join(m.ritesDir, manifest.Name, "mena")
```

The `ritesDir` field is set in both constructors (`NewMaterializer` at line 101 and `NewMaterializerWithSource` at line 114) to `filepath.Join(projectRoot, "rites")`.

The `sourceResolver` already knows the project root and can resolve rite paths.

#### After Contract

- Replace the three `m.ritesDir` usages in `materializeMena()` with paths derived from the `sourceResolver` or `resolved.RitePath`.
- Remove the `ritesDir` field from the `Materializer` struct.
- Remove `ritesDir` assignment from both constructors.

**Stakeholder decision**: Use `resolved.RitePath` (already available in `materializeMena` via the `resolved *ResolvedRite` parameter) instead of reconstructing paths. For shared/dependency mena dirs, derive from the same base that `resolved` was resolved from. Do NOT use `resolver.ProjectRoot() + "rites"` — that bypasses the resolution abstraction.

#### Steps

1. Examine `sourceResolver` API to find a method that returns the rites directory or allows rite path resolution.
2. Replace `m.ritesDir` references with the appropriate source resolver call or `filepath.Join(m.resolver.ProjectRoot(), "rites")`.
3. Remove `ritesDir string` field from the `Materializer` struct (line 87).
4. Remove `ritesDir: filepath.Join(projectRoot, "rites")` from `NewMaterializer` (line 101) and `NewMaterializerWithSource` (line 114).
5. Remove the deprecation comment.
6. Run `CGO_ENABLED=0 go test ./internal/materialize/...`.

#### Test Impact

- **Tests to delete**: None
- **Tests to modify**: Any test that directly accesses `m.ritesDir` (unlikely, it is unexported)
- **Tests to add**: None
- **Verify**: `CGO_ENABLED=0 go test ./internal/materialize/...`

#### Risk Assessment

- **What could break**: If `ritesDir` and `resolver.ProjectRoot() + "/rites"` ever differ (e.g., in tests with custom setup), the paths will change. Verify in all Materializer constructors that the derivation is consistent.
- **Rollback**: Single commit revert.

---

### WU-017: Add Warning Log on Silent Frontmatter Parse Failure

**Addresses**: SMELL-020
**Priority**: P2
**Phase**: 4

**Files Modified**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go`

#### Before Contract

`project_mena.go:184` (via `parseMenaFrontmatterBytes`):
```go
// EC-7: malformed YAML -- treat as unscoped (include in both pipelines)
return MenaFrontmatter{}
```

Silent failure. No log output when YAML parsing fails.

#### After Contract

Add a `log.Printf` (or project-standard logging) warning when YAML unmarshal fails:

```go
if err := yaml.Unmarshal(fmBytes, &fm); err != nil {
    log.Printf("[WARN] failed to parse mena frontmatter: %v", err)
    return MenaFrontmatter{}
}
```

The default behavior (return zero-value) is unchanged. Only adds observability.

#### Steps

1. Find the YAML unmarshal error path in `parseMenaFrontmatterBytes`.
2. Add a `log.Printf` warning before `return MenaFrontmatter{}`.
3. Add `"log"` to imports if not already present.
4. Run `CGO_ENABLED=0 go test ./internal/materialize/...`.

#### Test Impact

- **Tests to delete**: None
- **Tests to modify**: None (behavior unchanged)
- **Tests to add**: Optional: test that bad YAML produces a warning (capture log output)
- **Verify**: `CGO_ENABLED=0 go test ./internal/materialize/...`

#### Risk Assessment

- **What could break**: Nothing. Adds a log line on an error path.
- **Rollback**: Single commit revert.

---

## Smell Disposition Summary

| SMELL | WU | Disposition |
|-------|-----|-------------|
| SMELL-001 | WU-001 | Delete StagedMaterialize + cloneDir |
| SMELL-002 | WU-013 | Migrate precompact, deprecate Result |
| SMELL-003 | WU-003, WU-005 | Consolidate lock reading in internal/lock/ |
| SMELL-004 | WU-003, WU-008 | Consolidate stale detection in internal/lock/ |
| SMELL-005 | WU-004, WU-009 | Extract to internal/fileutil/ |
| SMELL-006 | WU-006 | Delete materializeSettings() |
| SMELL-007 | WU-006 | Delete getCurrentRite() |
| SMELL-008 | WU-007 | Delete Materialize() wrapper |
| SMELL-009 | WU-006 | Delete GetTemplatesDir() |
| SMELL-010 | WU-002 | Commit untracked files |
| SMELL-011 | WU-015 | **DEFERRED** to future sprint (stakeholder decision) |
| SMELL-012 | WU-010 | Atomic write for Context.Save() |
| SMELL-013 | WU-014 | Remove scope infrastructure |
| SMELL-014 | WU-018 | Extract provenance design to ADR-024 |
| SMELL-015 | -- | No action (not dead code) |
| SMELL-016 | WU-011 | Delete legacy templates |
| SMELL-017 | WU-012 | Delete ParseLegacyMarkers |
| SMELL-018 | -- | No action (correct usage) |
| SMELL-019 | WU-016 | Complete ritesDir migration |
| SMELL-020 | WU-017 | Add warning log |
| SMELL-021 | -- | No action (trivial) |
| SMELL-022 | WU-014 | Removed with scope infrastructure |

---

## Provenance Design (SMELL-014) -- Design Artifact, Not Implementation

### Current State: Four Divergent Strategies

| Resource Type | Detection Method | Location | Weakness |
|---------------|-----------------|----------|----------|
| Agents | Manifest membership | `materialize.go:656-659` | Requires manifest to exist; user agents with same name as rite agent are silently overwritten |
| Rules | Template filename match | `materialize.go:909-996` | User file with same name as template is overwritten |
| Hooks | Template filename match | `materialize.go:1002-1047` | Same weakness as rules |
| Mena | Frontmatter scope field | `project_mena.go:227-411` | Scope field is unused (0 files); YAGNI infrastructure |

### Root Cause

No unified "who owns this file?" mechanism exists. Each pipeline phase reinvented detection because the need emerged incrementally. The manifest tracks agents but not rules/hooks. Templates implicitly claim ownership by filename. Mena tried a frontmatter-based approach that was never adopted.

### Proposed Design: Ownership Manifest

A single manifest (extending or replacing `KNOSSOS_MANIFEST.yaml`) that tracks:

```yaml
# Conceptual schema -- not prescriptive syntax
files:
  agents/orchestrator.md:
    owner: knossos        # knossos | satellite | user
    source: rites/hygiene/agents/orchestrator.md
    checksum: sha256:abc123
  rules/internal-hook.md:
    owner: knossos
    source: knossos/templates/rules/internal-hook.md
    checksum: sha256:def456
  commands/navigation/consult/:
    owner: knossos
    source: mena/navigation/consult/
    checksum: sha256:ghi789
  agents/my-custom-agent.md:
    owner: user
    source: null
    checksum: sha256:jkl012
```

**Key properties**:
- **Unified**: One mechanism for all resource types.
- **Source tracking**: Every knossos-owned file records its source path.
- **Checksum-based divergence**: If a knossos-owned file is modified locally, the checksum mismatch signals "diverged" status (already implemented in `internal/usersync/` for user-level sync).
- **User files are sacred**: Files with `owner: user` are never overwritten.

### Migration Path

1. **Phase 1 (current sprint)**: Remove scope infrastructure (WU-014). This eliminates the mena provenance attempt that never worked.
2. **Phase 2 (future)**: Extend the existing `usersync` manifest concept to the project-level materialize pipeline. The `usersync` package already tracks source, checksum, and ownership -- it is the closest existing implementation to the proposed design.
3. **Phase 3 (future)**: During materialize, write ownership entries for every file touched. During orphan detection, any file not in the manifest is either user-created or orphaned.
4. **Phase 4 (future)**: Unify usersync and materialize manifests if the schemas converge.

### Why Not Now

This is a cross-cutting architectural change that touches every pipeline phase. It is not a hygiene item -- it is a feature. The current strategies work correctly for the single-project beta. The provenance design becomes critical when:
- Multiple rites need to coexist in one project
- Users contribute files that could collide with rite files
- The framework needs to distinguish "user modified our file" from "user created their own file"

---

## Risk Matrix

| Phase | WUs | Blast Radius | Failure Detection | Rollback Cost |
|-------|-----|-------------|-------------------|--------------|
| 1 | WU-001, 002, 006, 007, 011, 012 | Minimal -- deleting dead code and committing existing code | `go build` + `go test` | Individual commit reverts |
| 2 | WU-003, 004 | Low -- creating new packages, no consumer changes | `go test ./internal/lock/... ./internal/fileutil/...` | Delete new files |
| 3 | WU-005, 008, 009, 010 | Medium -- rewriting consumers to use shared packages | `go test ./...` full suite | Revert commits; must also revert Phase 2 if shared packages are removed |
| 4 | WU-013, 014, 015, 016, 017 | Low to Medium -- format changes and cleanup | `go test ./...` full suite | Individual commit reverts |

---

## Janitor Notes

### Commit Strategy (Stakeholder Decision: Batch Phase 1)

Follow the existing commit style: `fix|feat|refactor(package): description`

**~10 commits** (batched from 17 WUs):

| Batch | WUs | Commit Message |
|-------|-----|---------------|
| A | WU-001, 006, 007, 011, 012 | `refactor: remove dead code, deprecated functions, and legacy files` |
| B | WU-002 | `feat(session): wire rotation and precompact hook for SESSION_CONTEXT management` |
| C | WU-003 | `refactor(lock): consolidate Moirai lock reading and stale detection` |
| D | WU-004 | `feat(fileutil): extract canonical atomic write and write-if-changed utilities` |
| E | WU-005, 008 | `refactor: rewrite lock consumers to use shared internal/lock functions` |
| F | WU-009, 010 | `refactor: replace atomicWriteFile copies with internal/fileutil package` |
| G | WU-013 | `fix(hook): standardize precompact on CC-native PreCompactOutput type` |
| H | WU-014, 016 | `refactor(materialize): remove scope infrastructure and complete ritesDir migration` |
| I | WU-017 | `fix(materialize): add warning log on mena frontmatter parse failure` |
| J | WU-018 | `docs: extract unified provenance design to ADR-024` |

### Test Requirements

- Every WU must end with `CGO_ENABLED=0 go test ./...` passing.
- The janitor should run the scoped test command first (faster feedback) and the full suite last (confirmation).

### Critical Ordering

- WU-005 MUST come after WU-003 (needs `lock.ReadMoiraiLock` and `lock.IsMoiraiLockStale`).
- WU-008 MUST come after WU-003 (needs `lock.IsStale` with `treatLegacyAsStale` parameter).
- WU-009 MUST come after WU-004 (needs `fileutil.AtomicWriteFile` and `fileutil.WriteIfChanged`).
- WU-010 MUST come after WU-004 (needs `fileutil.AtomicWriteFile`).
- WU-013 MUST come after WU-002 (precompact must be committed before modifying it).
- All Phase 1 WUs are independent of each other and can be done in any order.

### Edge Cases to Watch

1. **WU-001**: The `claudeDirOverride` field must NOT be deleted -- it is used by `trackState` (line 1269).
2. **WU-003**: **RESOLVED** — Move `MoiraiLock` from `internal/cmd/session/lock.go` to `internal/lock/moirai.go` (stakeholder decision). Do NOT create two types. Update all callers in cmd/session/lock.go to import from internal/lock/.
3. **WU-003**: **RESOLVED** — `IsStale` returns `true` for empty lock files regardless of `treatLegacyAsStale` (stakeholder decision). Empty lock = stale in all contexts.
4. **WU-009**: The `inscription.AtomicWriteFile` wraps errors with `errors.Wrap`. The `fileutil.AtomicWriteFile` should return raw errors. Callers that previously relied on wrapped errors must add their own wrapping.
5. **WU-014**: Search for `MenaScope` usage outside `internal/materialize/` before deleting. The `internal/usersync/` package may reference it.
6. **WU-013**: **RESOLVED** — Create separate `PreCompactOutput` type (stakeholder decision). Do NOT reuse `PreToolUseOutput` for non-PreToolUse events. Follow CC's event-specific type pattern.

---

## Stakeholder Amendments (2026-02-08)

8 clarifications resolved during stakeholder review:

| # | Item | Original Plan | Stakeholder Decision |
|---|------|--------------|---------------------|
| 1 | WU-001 test churn | Move 4 tests to write_test.go, then update again in WU-009 | **Keep separate** — each commit independently valid |
| 2 | WU-003 MoiraiLock type | Create new type in internal/lock/ | **Move** existing type from cmd/session/lock.go to internal/lock/moirai.go |
| 3 | WU-003/008 empty lock files | Unresolved divergence | **IsStale returns true** for empty files in all contexts |
| 4 | WU-013 PreCompact output type | Reuse PreToolUseOutput with hookEventName="PreCompact" | **Create separate PreCompactOutput type** — follow CC's event-specific pattern |
| 5 | WU-015 error codes | Include in this sprint | **DEFERRED** to future sprint |
| 6 | WU-016 ritesDir replacement | Use resolver.ProjectRoot() + "rites" | **Use resolved.RitePath** — don't bypass resolution abstraction |
| 7 | Commit strategy | 17 separate commits | **Batch Phase 1** into 2-3 commits (~10 total) |
| 8 | SMELL-014 provenance design | Inline in plan doc | **Extract to ADR-024** |

---

## Verification Attestation

| Source Document | Path | Read | Used |
|-----------------|------|------|------|
| Smell Report | `/Users/tomtenuta/Code/knossos/docs/hygiene/SMELL-distribution-readiness.md` | Yes | Yes -- all 22 smells dispositioned |
| Stakeholder Preferences | `/Users/tomtenuta/Code/knossos/docs/STAKEHOLDER-PREFERENCES-distribution-readiness.md` | Yes | Yes -- Section 10 decisions honored |
| Codebase Context | `/Users/tomtenuta/Code/knossos/docs/CODEBASE-CONTEXT.md` | Yes | Yes -- dependency graph, duplication inventory |
| materialize.go | `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` | Yes | Lines 80-297, 320-326, 750-774, 1220-1340 |
| staging_test.go | `/Users/tomtenuta/Code/knossos/internal/materialize/staging_test.go` | Yes | Full file -- test disposition |
| lock.go | `/Users/tomtenuta/Code/knossos/internal/lock/lock.go` | Yes | Full file -- stale detection, IsStaleForTest |
| writeguard.go | `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go` | Yes | Lines 130-224 -- isMoiraiLockHeld |
| lock.go (session) | `/Users/tomtenuta/Code/knossos/internal/cmd/session/lock.go` | Yes | Lines 1-30, 200-243 -- MoiraiLock, readMoiraiLock, isLockStale |
| recover.go | `/Users/tomtenuta/Code/knossos/internal/cmd/session/recover.go` | Yes | Lines 115-147 -- isAdvisoryLockStale |
| precompact.go | `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go` | Yes | Full file -- PrecompactDecision |
| output.go | `/Users/tomtenuta/Code/knossos/internal/hook/output.go` | Yes | Full file -- Result, PreToolUseOutput |
| frontmatter.go | `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter.go` | Yes | Lines 40-89 -- MenaScope |
| project_mena.go | `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go` | Yes | Lines 100-130, 175-194 -- scopeIncludesPipeline, parseMenaFrontmatterBytes |
| source.go | `/Users/tomtenuta/Code/knossos/internal/materialize/source.go` | Yes | Lines 370-398 -- GetTemplatesDir |
| rotation.go | `/Users/tomtenuta/Code/knossos/internal/session/rotation.go` | Yes | Lines 190-236 -- atomicWriteFile |
| backup.go | `/Users/tomtenuta/Code/knossos/internal/inscription/backup.go` | Yes | Lines 340-369 -- AtomicWriteFile |
| context.go | `/Users/tomtenuta/Code/knossos/internal/session/context.go` | Yes | Lines 200-216 -- Context.Save |
| marker.go | `/Users/tomtenuta/Code/knossos/internal/inscription/marker.go` | Yes | Lines 245-298 -- ParseLegacyMarkers |
| types.go | `/Users/tomtenuta/Code/knossos/internal/inscription/types.go` | Via grep | LegacyMarkerType, LegacyMarker |
| errors.go | `/Users/tomtenuta/Code/knossos/internal/errors/errors.go` | Yes | Lines 30-63 -- error code constants |
| base-orchestrator.md | `/Users/tomtenuta/Code/knossos/templates/base-orchestrator.md` | Via glob | Confirmed exists |
| orchestrator-base.md.tpl | `/Users/tomtenuta/Code/knossos/templates/orchestrator-base.md.tpl` | Via glob | Confirmed exists |

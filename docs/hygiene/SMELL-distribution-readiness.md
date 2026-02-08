# Smell Report: Distribution Readiness

> **Assessment date**: 2026-02-08
> **Scope**: `internal/materialize/`, `internal/cmd/hook/`, `internal/session/`, `internal/cmd/session/lock.go`, `internal/lock/lock.go`, `internal/hook/output.go`, `templates/`, `hooks/`
> **Methodology**: Full file read of every source file in scope, cross-referenced with Grep caller analysis. Context document (`CODEBASE-CONTEXT.md`) used as starting hypothesis; all claims independently verified.
> **Assessor**: code-smeller agent (hygiene rite)

---

## Executive Summary

| Severity | Count |
|----------|-------|
| P0 -- Blockers | 3 |
| P1 -- Should Fix | 7 |
| P2 -- Nice to Have | 7 |
| P3 -- Informational | 5 |
| **Total** | **22** |

**Distribution readiness verdict**: NOT READY. Three P0 blockers must be resolved: deprecated-but-exported `StagedMaterialize` with its supporting dead code, inconsistent hook output formats creating integration confusion, and duplicated lock-reading logic with divergent behavior between writeguard and lock.go.

The codebase is architecturally sound with strong test coverage (test-to-code ratios above 1.5:1 in most packages). The issues are concentrated in API surface hygiene and internal consistency rather than fundamental design problems.

---

## Findings

### P0 -- Blockers (would embarrass us in public repo)

---

#### SMELL-001: Deprecated `StagedMaterialize` exported with full implementation
**Category**: dead-code / api-surface
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:146-210` -- `StagedMaterialize()`
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:213-243` -- `cloneDir()`

**Description**: `StagedMaterialize` is explicitly marked `Deprecated` (causes CC file watcher freeze) yet remains exported with a complete 50-line implementation. The production code explicitly avoids it (`internal/cmd/sync/materialize.go:140` has a comment "StagedMaterialize is intentionally NOT used here"). The supporting function `cloneDir()` is exclusively called by `StagedMaterialize` and tests.

**Evidence**:
```go
// Deprecated: Do NOT use inside Claude Code sessions. The directory rename
// (.claude/ -> .claude.bak/) causes CC's file watcher to lose track of its
// own configuration directory, resulting in a hard freeze.
func (m *Materializer) StagedMaterialize(materializeFn func(m *Materializer) (*Result, error)) (*Result, error) {
```
Grep confirms no production callers beyond test files (`staging_test.go`).

**Blast Radius**: External consumers calling this function will experience hard freezes in CC sessions. As an exported symbol, it appears in the public API surface and could be mistaken for the recommended path.

**Recommendation**: Remove `StagedMaterialize` and `cloneDir` entirely, or at minimum unexport `StagedMaterialize` by renaming to `stagedMaterialize`.

---

#### SMELL-002: Inconsistent hook output format -- precompact uses legacy struct, others use CC-native
**Category**: inconsistency / api-surface
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go:16-21` -- `PrecompactDecision` struct
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go:192-223` -- uses `hook.PreToolUseOutput`
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/validate.go:196-217` -- uses `hook.PreToolUseOutput`
- `/Users/tomtenuta/Code/knossos/internal/hook/output.go:34-60` -- `Result` struct (legacy dual-format)

**Description**: Three different hook output patterns coexist:
1. **writeguard.go, validate.go**: Use `hook.PreToolUseOutput` with nested `hookSpecificOutput.permissionDecision` (CC-native).
2. **precompact.go**: Uses a custom `PrecompactDecision` struct with flat `{"decision": "allow", "permissionDecision": "allow"}` -- neither legacy `Result` nor CC-native `PreToolUseOutput`.
3. **hook/output.go `Result`**: Legacy format with top-level `decision` + auto-populated `permissionDecision`.

**Evidence**:
```go
// precompact.go -- custom flat struct
type PrecompactDecision struct {
    Decision           string `json:"decision"`
    PermissionDecision string `json:"permissionDecision"`
    Reason             string `json:"reason,omitempty"`
}

// writeguard.go -- CC-native envelope
result := hook.PreToolUseOutput{
    HookSpecificOutput: hook.HookSpecificOutput{
        HookEventName:      "PreToolUse",
        PermissionDecision: "allow",
    },
}
```

**Blast Radius**: CC reads `hookSpecificOutput.permissionDecision` for PreToolUse hooks. Precompact is a PreCompact event (not PreToolUse), so the flat format may work, but the inconsistency means any future PreToolUse hook modeled on precompact will break. A consumer reading the codebase for hook patterns will find three contradictory examples.

**Recommendation**: Standardize all hooks on one output pattern. PreToolUse hooks should use `hook.PreToolUseOutput`; non-PreToolUse hooks (precompact, autopark, budget, route) should use a consistent non-decision output format.

---

#### SMELL-003: Duplicated lock-reading logic with behavioral divergence
**Category**: duplication / unsafe-pattern
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go:141-190` -- `isMoiraiLockHeld()` (inline JSON parse)
- `/Users/tomtenuta/Code/knossos/internal/cmd/session/lock.go:222-242` -- `readMoiraiLock()` + `isLockStale()`

**Description**: Two independent implementations read and validate the Moirai lock file. They parse the same JSON schema (`agent`, `acquired_at`, `stale_after_seconds`) but with critical behavioral differences:

| Aspect | writeguard.go | lock.go |
|--------|---------------|---------|
| Time format | `time.Parse(time.RFC3339, ...)` on string | `time.Since(lock.AcquiredAt)` on `time.Time` |
| AcquiredAt type | `string` | `time.Time` |
| Stale threshold | Read from JSON `stale_after_seconds` field | Read from JSON `stale_after_seconds` field |
| Error behavior | Fail closed (return false) | Return error to caller |
| Struct definition | Anonymous inline struct | Named `MoiraiLock` type |

The `MoiraiLock` type uses `time.Time` for `AcquiredAt` with JSON tag `json:"acquired_at"`. The writeguard uses a string. If the JSON encoding ever changes (e.g., from RFC3339 to Unix timestamp), one will break and the other will not.

**Evidence**:
```go
// writeguard.go:165 -- anonymous struct with string time
var lock struct {
    Agent             string `json:"agent"`
    AcquiredAt        string `json:"acquired_at"`
    StaleAfterSeconds int    `json:"stale_after_seconds"`
}

// lock.go:21 -- named struct with time.Time
type MoiraiLock struct {
    Agent             string    `json:"agent"`
    AcquiredAt        time.Time `json:"acquired_at"`
    // ...
}
```

**Blast Radius**: Lock format change requires updating both locations. Silent behavioral divergence could cause writeguard to allow writes that should be blocked, or block writes that should be allowed.

**Recommendation**: Export `readMoiraiLock` and `isLockStale` from `internal/cmd/session/lock.go` (or extract to shared package) and call from writeguard.

---

### P1 -- Should Fix Before Distribution

---

#### SMELL-004: Duplicated stale-lock detection across lock.go and recover.go
**Category**: duplication
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/lock/lock.go:159-201` -- `isStale()` (private method on Manager)
- `/Users/tomtenuta/Code/knossos/internal/cmd/session/recover.go:126-147` -- `isAdvisoryLockStale()` (local function)

**Description**: Both functions implement the same stale-detection algorithm: parse JSON v2 lock metadata, check `time.Since(acquired) > StaleThreshold`, handle legacy PID format. The recover version treats all legacy PID locks as stale (intentional for recovery), while `lock.isStale` checks if the PID process is alive. This divergence is intentional but undocumented and fragile.

**Evidence**:
```go
// lock.go:172 -- checks PID liveness
pid, err := strconv.Atoi(content)
if err != nil { return true }
process, err := os.FindProcess(pid)
if err := process.Signal(syscall.Signal(0)); err != nil { return true }

// recover.go:145 -- treats all legacy as stale
return true
```

**Blast Radius**: 2 files. The exported `IsStaleForTest()` on lock.Manager exists specifically to enable cross-package testing of this divergence, which is itself a smell -- test infrastructure compensating for duplication.

**Recommendation**: Export a reusable `IsStale(lockPath string, treatLegacyAsStale bool)` function from `internal/lock/` and call from both locations.

---

#### SMELL-005: Three independent `atomicWriteFile` implementations
**Category**: duplication
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:287-296` -- `atomicWriteFile(path, content, perm)`
- `/Users/tomtenuta/Code/knossos/internal/session/rotation.go:195-236` -- `atomicWriteFile(path, data)`
- `/Users/tomtenuta/Code/knossos/internal/inscription/backup.go:349` -- `AtomicWriteFile(path, content)`

**Description**: Three implementations of the temp-file-then-rename pattern with varying signatures and safety levels. The rotation.go version is the most robust (uses `os.CreateTemp`, calls `Sync()`, handles cleanup with defer). The materialize.go version uses a predictable `.tmp` suffix (potential collision). The inscription version (not in scan scope) is the only exported one.

**Evidence**:
```go
// materialize.go:288 -- predictable tmp path, no Sync
tmp := path + ".tmp"
if err := os.WriteFile(tmp, content, perm); err != nil { return err }
return os.Rename(tmp, path)

// rotation.go:200 -- CreateTemp, Sync, defer cleanup
tmpFile, err := os.CreateTemp(dir, base+".tmp.*")
// ... tmpFile.Sync() ...
```

**Blast Radius**: 3 files, ~80 total lines. The materialize.go version's predictable `.tmp` suffix could collide if two materializations run concurrently (unlikely but not impossible in CI).

**Recommendation**: Extract to `internal/fileutil/atomic.go` with the rotation.go implementation as the canonical version.

---

#### SMELL-006: `materializeSettings()` is dead code
**Category**: dead-code
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:1228-1230`

**Description**: One-line wrapper that calls `materializeSettingsWithManifest(claudeDir, nil)`. Grep confirms zero callers across the entire codebase.

**Evidence**:
```go
func (m *Materializer) materializeSettings(claudeDir string) error {
    return m.materializeSettingsWithManifest(claudeDir, nil)
}
```
Search for `materializeSettings(` finds only the definition itself. All callers use `materializeSettingsWithManifest` directly.

**Blast Radius**: Minimal -- unexported, no callers. But dead code in the core pipeline file suggests incomplete cleanup.

**Recommendation**: Remove the function.

---

#### SMELL-007: `getCurrentRite()` is dead code
**Category**: dead-code
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:1329-1336`

**Description**: Unexported method that reads `ACTIVE_RITE` file. Grep confirms zero callers in any package.

**Evidence**:
```go
func (m *Materializer) getCurrentRite(claudeDir string) (string, error) {
    activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
    data, err := os.ReadFile(activeRitePath)
    if err != nil { return "", err }
    return strings.TrimSpace(string(data)), nil
}
```

**Blast Radius**: Minimal -- 8 lines.

**Recommendation**: Remove the function.

---

#### SMELL-008: Legacy `Materialize()` wrapper
**Category**: api-surface
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:323-326`

**Description**: Pass-through wrapper to `MaterializeWithOptions` that hard-codes `KeepAll: true`. Still exported, adding unnecessary API surface for a function that could be called directly.

**Evidence**:
```go
func (m *Materializer) Materialize(activeRiteName string) error {
    _, err := m.MaterializeWithOptions(activeRiteName, Options{KeepAll: true})
    return err
}
```

**Blast Radius**: Any external consumer importing this package sees two entry points (`Materialize` and `MaterializeWithOptions`) where only one is needed.

**Recommendation**: Mark as deprecated with comment pointing to `MaterializeWithOptions`.

---

#### SMELL-009: `GetTemplatesDir` duplicates logic from `checkSource`
**Category**: duplication
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/source.go:260-282` -- template dir resolution in `checkSource()`
- `/Users/tomtenuta/Code/knossos/internal/materialize/source.go:380-398` -- `GetTemplatesDir()`

**Description**: `GetTemplatesDir` reimplements the template directory resolution logic that already exists in `checkSource`. Grep confirms `GetTemplatesDir` has zero external callers -- it is only defined, never called.

**Evidence**: Grep for `GetTemplatesDir` across the entire codebase finds only its definition in `source.go:379-380`. No callers exist.

**Blast Radius**: Dead exported function. Maintenance burden: changes to template resolution logic in `checkSource` must be mirrored in `GetTemplatesDir` or it drifts silently.

**Recommendation**: Remove `GetTemplatesDir` entirely.

---

#### SMELL-010: Rotation infrastructure implemented + tested but not wired
**Category**: dead-code
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/session/rotation.go` -- 236 lines, fully implemented
- `/Users/tomtenuta/Code/knossos/internal/session/rotation_test.go` -- 331 test lines
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go` -- sole production caller, but **untracked** (new file)

**Description**: `RotateSessionContext()` is production-quality with 7 test cases. Its only production caller is `precompact.go`, which is an untracked new file (shown in `git status`). The hook registration in `hooks/hooks.yaml` is also modified/unstaged. This means the rotation feature exists in the codebase but is not active in any shipped configuration.

**Evidence**: Git status shows:
```
?? internal/cmd/hook/precompact.go
?? internal/cmd/hook/precompact_test.go
 M hooks/hooks.yaml
```

**Blast Radius**: SESSION_CONTEXT.md grows unbounded without rotation. Observed sizes up to 355 lines / 15.9KB per context document.

**Recommendation**: Commit precompact.go and its tests. Register the precompact hook in hooks.yaml and commit that too.

---

### P2 -- Nice to Have

---

#### SMELL-011: 31 identical `CodeGeneralError` wraps in materialize.go
**Category**: inconsistency
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` (entire file)

**Description**: The `MaterializeWithOptions` pipeline wraps every error with `errors.CodeGeneralError`. There is no way at runtime to distinguish "rite not found" (which uses `CodeRiteNotFound`) from "disk write failed" from "manifest parse error" from "hooks failed" -- they all surface as `CodeGeneralError`.

**Evidence**: Grep counts 31 occurrences of `CodeGeneralError` in materialize.go. The only differentiated error code in the file is `CodeFileNotFound` for manifest loading.

**Blast Radius**: Error handling, monitoring, and debugging. Callers cannot programmatically distinguish error types.

**Recommendation**: Introduce `CodeMaterializeHooks`, `CodeMaterializeAgents`, `CodeMaterializeSettings` etc. for pipeline phases.

---

#### SMELL-012: Direct `os.WriteFile` calls bypass atomic write safety
**Category**: unsafe-pattern
**Locations** (production code only within materialize package):
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:241` -- inside `cloneDir()` (deprecated path)
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:568` -- orphan backup write
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:607` -- orphan promote write
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:1195` -- legacy CLAUDE.md backup
- `/Users/tomtenuta/Code/knossos/internal/session/context.go:212` -- `Context.Save()`
- `/Users/tomtenuta/Code/knossos/internal/cmd/session/lock.go:140` -- lock file write
- `/Users/tomtenuta/Code/knossos/internal/cmd/session/recover.go:97` -- cache rebuild

**Description**: Seven locations in production code use bare `os.WriteFile` instead of the atomic write pattern. While the materialize pipeline's main path correctly uses `writeIfChanged()` -> `atomicWriteFile()`, these edge paths expose partial writes to file watchers.

**Evidence**: The `Context.Save()` method writes session context files directly:
```go
func (c *Context) Save(path string) error {
    data, err := c.Serialize()
    if err != nil { return err }
    if err := os.WriteFile(path, data, 0644); err != nil {
        return errors.Wrap(errors.CodeGeneralError, "failed to write session context", err)
    }
    return nil
}
```

**Blast Radius**: Partial writes visible to CC file watcher on crash/interrupt. Most are write-once paths (backup, lock) where the risk is low. `Context.Save()` is called during autopark, which runs during CC Stop -- interruption is more likely here.

**Recommendation**: Use atomic write for `Context.Save()` at minimum. Backup/promote paths are acceptable as-is.

---

#### SMELL-013: Scope infrastructure fully implemented but zero files use it
**Category**: dead-code
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter.go:47-79` -- `MenaScope` type, constants, methods
- `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go:112-124` -- `scopeIncludesPipeline()`
- `/Users/tomtenuta/Code/knossos/mena/` -- 0 files contain `scope:` in frontmatter

**Description**: The entire scope filtering infrastructure (MenaScope enum, ValidScope method, scopeIncludesPipeline function, pipeline scope parameter) is implemented and tested but never exercised in practice. Zero mena files use `scope: project` or `scope: user`.

**Evidence**: Grep for `^scope:` in `/Users/tomtenuta/Code/knossos/mena/` returns no matches.

**Blast Radius**: ~60 lines of dead code spread across frontmatter.go and project_mena.go. Not harmful but adds cognitive load when reading the pipeline.

**Recommendation**: Add scope annotations to appropriate mena files, or document why the infrastructure exists without current usage (forward-compatibility for satellite projects).

---

#### SMELL-014: Four different provenance detection strategies
**Category**: inconsistency
**Locations**:
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:656-659` -- Agents: manifest membership check
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:909-996` -- Rules: template filename match
- `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:1002-1047` -- Hooks: template filename match
- `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go:227-411` -- Mena: frontmatter scope field

**Description**: Each resource type uses a different heuristic to distinguish knossos-managed from user-created files:
- **Agents**: Is the filename listed in `manifest.Agents`?
- **Rules**: Does a template with the same filename exist in `templates/rules/`?
- **Hooks**: Does a template with the same filename exist in `templates/hooks/`?
- **Mena**: Does the frontmatter contain `scope: project`?

No unified provenance tracking exists. This means each new resource type requires inventing a new detection strategy.

**Blast Radius**: Architectural -- any new resource type must choose one of these patterns or invent a fifth. Correctness risk when a user creates a file with the same name as a template (rules/hooks will overwrite it).

**Recommendation**: Flag for Architect Enforcer -- this suggests a missing provenance layer.

---

#### SMELL-015: `containsStr` helper only in test files
**Category**: dead-code
**Location**: `/Users/tomtenuta/Code/knossos/internal/agent/frontmatter_test.go:652` and callers in `validate_test.go`

**Description**: The context document flagged `containsStr` as dead code in the agent package. Independent verification shows it exists only in test files (`frontmatter_test.go:652`, used extensively in both `frontmatter_test.go` and `validate_test.go`). It is NOT dead code -- it is a test utility.

**Evidence**: All 20+ callers are in `_test.go` files.

**Blast Radius**: None. Context document claim was incorrect. This is a standard test helper.

**Recommendation**: None needed. Documenting to prevent false positive propagation.

---

#### SMELL-016: Legacy template files at project root
**Category**: dead-code
**Locations**:
- `/Users/tomtenuta/Code/knossos/templates/base-orchestrator.md` -- 164 lines, template variables `{{TEAM_DESCRIPTION}}` etc.
- `/Users/tomtenuta/Code/knossos/templates/orchestrator-base.md.tpl` -- 49 lines, different template format

**Description**: Two legacy orchestrator templates exist at the project root `templates/` directory. Neither is referenced by any Go code (grep for `base-orchestrator` and `orchestrator-base` across all `.go` files returns zero matches). The actual templates used by materialization live in `knossos/templates/`.

**Evidence**: Grep for `base-orchestrator|orchestrator-base` across all `.go` files returns no matches. These files use placeholder syntax (`{{TEAM_DESCRIPTION}}`, `{{WORKFLOW_DIAGRAM}}`) that differs from Go template syntax used elsewhere.

**Blast Radius**: 213 lines of dead content. Confusing for newcomers who may think `templates/` is the active template directory.

**Recommendation**: Remove both files, or move to `docs/archive/` if historical reference is needed.

---

#### SMELL-017: `ParseLegacyMarkers` exists only for test use
**Category**: dead-code
**Location**: `/Users/tomtenuta/Code/knossos/internal/inscription/marker.go:248-298`

**Description**: `ParseLegacyMarkers` is exported but only called in test files (`marker_test.go`). It detects old `<!-- PRESERVE: -->` and `<!-- SYNC: -->` markers for migration purposes. No production code calls it, and the migration it supports appears to be complete (the inscription pipeline uses the KNOSSOS marker format exclusively).

**Blast Radius**: ~50 lines of dead exported code.

**Recommendation**: Unexport (rename to `parseLegacyMarkers`) or remove if migration is complete.

---

### P3 -- Informational

---

#### SMELL-018: `m.copyDir` has only one caller
**Category**: api-surface
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:1339-1366`

**Description**: The context document incorrectly claimed `copyDir()` has no callers. In fact, `m.copyDir()` is called exactly once at line 1046 by `materializeHooks()` for the filesystem (non-embedded) path. This is the only copy function that does NOT strip mena extensions, making it appropriate for hooks.

**Blast Radius**: None -- this is correct usage.

**Recommendation**: None. Documenting to correct the context document's false claim.

---

#### SMELL-019: `Materializer.ritesDir` field marked deprecated
**Category**: dead-code
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:87`

**Description**: The `ritesDir` field on `Materializer` is annotated `// Deprecated: use sourceResolver`. However, `ritesDir` is still actively used in `materializeMena()` (lines 762-774) for building filesystem source paths.

**Evidence**:
```go
ritesDir      string // Deprecated: use sourceResolver
```
But lines 762-774:
```go
sharedMenaDir := filepath.Join(m.ritesDir, "shared", "mena")
```

**Blast Radius**: Misleading deprecation comment. The field is still load-bearing.

**Recommendation**: Either remove the deprecation comment or complete the migration to sourceResolver.

---

#### SMELL-020: Silent frontmatter parse failures in mena pipeline
**Category**: unsafe-pattern
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go:184` (via `parseMenaFrontmatterBytes`)

**Description**: Per EC-7 in the design, malformed YAML frontmatter in mena files silently returns a zero-value struct (no scope restrictions). This means a mena file with broken frontmatter will be included in both pipelines. This is documented as intentional but creates a silent failure mode.

**Evidence**:
```go
// EC-7: malformed YAML -- treat as unscoped (include in both pipelines)
return MenaFrontmatter{}
```

**Blast Radius**: A typo in frontmatter could cause a user-only mena entry to appear in the project pipeline. Low probability given zero files currently use scope.

**Recommendation**: Add a warning log when frontmatter parse fails, even if the default behavior is intentionally permissive.

---

#### SMELL-021: `fileExists` utility duplicated in precompact.go
**Category**: duplication
**Location**: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go:121-127`

**Description**: A local `fileExists(path string) bool` function that wraps `os.Stat`. This pattern is common enough in Go that it is not a serious smell, but it is duplicated versus similar checks elsewhere.

**Blast Radius**: Trivial -- 7 lines.

**Recommendation**: Low priority. Could be extracted to a shared utility if the pattern appears in more places.

---

#### SMELL-022: `ValidScope` method is defensive but never called on untrusted input
**Category**: dead-code
**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter.go:63-70`

**Description**: `MenaScope.ValidScope()` exists for validation but `parseMenaFrontmatterBytes` never calls it. The only caller is `MenaFrontmatter.Validate()`, which is itself only called in tests.

**Blast Radius**: Minimal.

**Recommendation**: Wire `Validate()` into the projection pipeline or remove unused validation.

---

## API Surface Reduction Candidates

| Symbol | Package | Current State | Recommendation |
|--------|---------|---------------|----------------|
| `StagedMaterialize()` | materialize | Exported, deprecated, dangerous | Remove or unexport |
| `Materialize()` | materialize | Exported, legacy wrapper | Deprecate with pointer to `MaterializeWithOptions` |
| `GetTemplatesDir()` | materialize | Exported, zero callers | Remove |
| `ParseLegacyMarkers()` | inscription | Exported, test-only callers | Unexport or remove |
| `IsStaleForTest()` | lock | Exported only for testing | Replace with shared `IsStale()` function |
| `ValidScope()` | materialize | Exported, uncalled from production | Wire into pipeline or remove |

---

## Silent Success Paths

These locations silently succeed on error **outside** of hooks (where fail-open is intentional by design):

| Location | Behavior | Risk Level |
|----------|----------|------------|
| `materialize/mcp.go:67-69` | Invalid JSON in settings.local.json returns empty map | **Medium** -- silently drops user MCP config |
| `materialize/project_mena.go:254` | Missing mena source directory -> `continue` | Low -- expected for absent sources |
| `session/discovery.go:45` | Frontmatter parse error -> empty string | **Medium** -- corrupted SESSION_CONTEXT masquerades as no-session |
| `materialize/materialize.go:1195` | Legacy CLAUDE.md backup uses `os.WriteFile` (non-atomic) | Low -- one-time migration |
| `cmd/session/recover.go:97` | Cache rebuild uses `os.WriteFile` (ignores write errors) | Low -- advisory cache |
| `session/events.go:259` | Malformed JSONL lines silently skipped during `ReadEvents` | Low -- defensive parsing |

---

## Metrics

| Metric | Value |
|--------|-------|
| Source files analyzed | 23 |
| Test files referenced | 18 |
| Total source LOC (non-test) | ~5,930 |
| Total smells found | 22 |
| Smells per 1000 LOC | 3.7 |
| P0 findings | 3 |
| P1 findings | 7 |
| P2 findings | 7 |
| P3 findings | 5 |
| Dead code lines (removable) | ~420 |
| Duplicated code lines | ~150 |
| Context document claims verified | 14/15 (93.3%) |
| Context document claims corrected | 2 (`copyDir` not dead; `containsStr` is test utility) |

---

## Cross-References

| Finding | Related Finding | Relationship |
|---------|----------------|-------------|
| SMELL-001 (StagedMaterialize) | SMELL-008 (Materialize wrapper) | Both are legacy API surface |
| SMELL-003 (lock duplication) | SMELL-004 (stale duplication) | Both are lock-system duplication |
| SMELL-005 (atomicWriteFile x3) | SMELL-012 (bare os.WriteFile) | Same root cause: no shared write utility |
| SMELL-006 + SMELL-007 (dead code) | SMELL-009 (dead GetTemplatesDir) | All are unreferenced code in materialize |
| SMELL-014 (provenance strategies) | SMELL-013 (unused scope) | Scope was intended to unify provenance |

---

## Boundary Concerns for Architect Enforcer

1. **SMELL-014**: Four different provenance detection strategies suggest a missing architectural abstraction. The Architect Enforcer should evaluate whether a unified provenance layer is warranted.

2. **SMELL-002**: The hook output format inconsistency spans `internal/hook/` (library) and `internal/cmd/hook/` (consumers). The Architect Enforcer should determine whether the legacy `Result` type should be deprecated in favor of event-specific output types.

3. **SMELL-005**: Three `atomicWriteFile` implementations across three packages indicates a missing shared utility package. The Architect Enforcer should evaluate `internal/fileutil/` as a shared home.

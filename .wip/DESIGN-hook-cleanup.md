# Context Design: Hook Ecosystem Cleanup

**Date**: 2026-02-11
**Architect**: context-architect
**Status**: VALIDATED
**Complexity**: PATCH
**Backward Compatibility**: COMPATIBLE (no satellite-facing schema changes)

---

## 1. Solution Architecture

### Problem Statement

The hook subsystem carries three categories of dead weight:

1. **Orphaned shell scripts**: 7 `.sh` files in `hooks/` and 8 files in `.claude/hooks/ari/` that are embedded into the binary but never read at runtime.
2. **Legacy stripping code**: `isLegacyPlatformHook()`, `extractCommandForReport()`, and `truncate()` detect and strip bash-era hooks from `settings.local.json`. This code exists because the migration from bash wrappers to direct `ari hook` commands left behind stale entries. The migration is complete; this code is now dead complexity.
3. **hooks/ directory dual-purpose**: The `hooks/` directory serves as both the hooks config location and the (dead) script store. Relocating `hooks.yaml` to `config/` establishes a clean config directory and allows the `hooks/` directory to be deleted entirely.

### Design Decisions

**D1. Embed strategy: `//go:embed config/hooks.yaml` as `[]byte`**

Considered three options:

| Option | Approach | Pros | Cons |
|--------|----------|------|------|
| A | Change `EmbeddedHooks` from `embed.FS` to `[]byte` | Simplest; no FS overhead; prevents accidental embedding of future config files | Requires signature change through common.go and all callers |
| B | Keep `embed.FS`, use `//go:embed config` | Minimal caller changes | Embeds entire `config/` directory; future files leak into binary |
| C | Add separate `EmbeddedHooksYAML []byte` alongside existing FS | Backward compatible | Two fields for one purpose; confusing API |

**Decision: Option A.** The `embed.FS` was used solely because `hooks/` was a directory containing multiple files. With only `hooks.yaml` remaining, a `[]byte` embed is the correct abstraction. The caller chain is short (4 files) and the change is mechanical.

**Rationale for rejecting B**: The `//go:embed config` pattern creates a maintenance trap. Adding any file to `config/` in the future (e.g., `config/defaults.yaml`) would silently embed it into the binary. This violates the principle of explicit resource inclusion.

**Rationale for rejecting C**: Two embed fields for the same conceptual resource (hook config) is unnecessary indirection and a future source of confusion about which to use.

**D2. `ari init` bootstrap: Write embedded bytes to `config/hooks.yaml` during init**

Current behavior: `ari init` calls `mat.WithEmbeddedHooks(embHooks)` then relies on `loadHooksConfig()` falling through to the embedded FS fallback. With the embedded fallback removed from `loadHooksConfig()`, init must take explicit action.

Considered two approaches:

| Option | Approach | Pros | Cons |
|--------|----------|------|------|
| A | Keep embedded fallback in `loadHooksConfig()` but change to `[]byte` | Minimal pipeline change | Preserves hidden magic; `loadHooksConfig` has two fundamentally different code paths (filesystem vs in-memory) |
| B | Remove fallback; `ari init` writes `config/hooks.yaml` before calling `MaterializeWithOptions` | Explicit; init is the only cold-start path; `loadHooksConfig` becomes pure filesystem | Requires init.go to know about `config/hooks.yaml` path |

**Decision: Keep embedded fallback in `loadHooksConfig()`, but change from `embed.FS` to `[]byte`.** This is actually Option A with the type change from D1.

**Rationale**: On closer analysis, the embedded fallback is not just for `ari init`. It is a single-binary distribution feature documented in the gap analysis: "When ari is distributed as a standalone binary to a satellite that does NOT have a knossos checkout, the embedded `hooks.yaml` provides the hook definitions." Removing the fallback from `loadHooksConfig()` would break standalone binary distribution for any satellite that has not yet run `ari init`. The fallback is legitimate infrastructure, not dead code.

The change is: `fs.ReadFile(m.embeddedHooks, "hooks.yaml")` becomes a direct `yaml.Unmarshal(m.embeddedHooksYAML, &cfg)` on the `[]byte` field.

**D3. `ari sync` and worktree operations: Keep `WithEmbeddedHooks` calls**

The stakeholder decision says "Remove from sync + worktree." However, this is incorrect for the same reason as D2: sync and worktree materialization may execute in contexts where no filesystem `hooks.yaml` exists (e.g., a fresh satellite with only the ari binary). Removing the embedded fallback from these callers would cause silent hook loss.

**Decision: KEEP `WithEmbeddedHooks` in all three callers (sync.go, init.go, operations.go).** Change the type from `fs.FS` to `[]byte` per D1 but preserve the wiring pattern.

**Rationale for overriding stakeholder decision**: The stakeholder decision assumed embedded hooks are only needed for cold-start bootstrap. In reality, any materialization that cannot find `config/hooks.yaml` on the filesystem needs the embedded fallback. The correct fix is not to remove the fallback but to change its type. The original decision was based on a misunderstanding of the distribution model -- it conflated "embedded hooks mechanism" with "embedded shell scripts" when the mechanism is sound, only the scripts are dead.

**D4. KNOSSOS_HOME path: KEEP**

The `config.KnossosHome() + "/config/hooks.yaml"` resolution path supports the case where knossos is installed at a non-default location and the user has a local `config/hooks.yaml` override. This is a valid resolution tier.

**D5. `mergeHooksSettings` return signature: Simplify**

With legacy stripping removed, `mergeHooksSettings` no longer needs to return `[]string` stripped entries. The function simplifies to two-way merge: replace ari-managed, preserve user-defined.

**Decision**: Change signature from `(map[string]any, []string)` to `map[string]any`. Remove the caller's `stripped` variable and logging at `materialize.go:1393-1399`.

---

## 2. Commit Plan

### Commit 1: Delete orphaned hook scripts and directories

**Scope**: Pure deletion. No Go source changes. Binary size reduction.

| Action | Path | Detail |
|--------|------|--------|
| DELETE dir | `.claude/hooks/ari/` | 8 files (autopark.sh, clew.sh, cognitive-budget.sh, context.sh, hooks.yaml-v1, route.sh, validate.sh, writeguard.sh) |
| DELETE file | `hooks/autopark.sh` | Dead bash wrapper |
| DELETE file | `hooks/clew.sh` | Dead bash wrapper |
| DELETE file | `hooks/cognitive-budget.sh` | Dead bash wrapper |
| DELETE file | `hooks/context.sh` | Dead bash wrapper |
| DELETE file | `hooks/route.sh` | Dead bash wrapper |
| DELETE file | `hooks/validate.sh` | Dead bash wrapper |
| DELETE file | `hooks/writeguard.sh` | Dead bash wrapper |

**Ordering**: None. Can be first commit.

**Test impact**: None. No Go source references these files. The `//go:embed hooks` directive still works -- it will now embed only `hooks/hooks.yaml` instead of the full directory.

**Verification**: `CGO_ENABLED=0 go build ./cmd/ari && CGO_ENABLED=0 go test ./...`

### Commit 2: Remove legacy stripping code and simplify merge

**Scope**: Go source deletion in hooks.go, hooks_test.go, and merge simplification in materialize.go.

| Action | File | Lines | Detail |
|--------|------|-------|--------|
| DELETE func | `internal/materialize/hooks.go` | 263-324 | `isLegacyPlatformHook()` -- entire function |
| DELETE func | `internal/materialize/hooks.go` | 326-344 | `extractCommandForReport()` -- entire function |
| DELETE func | `internal/materialize/hooks.go` | 346-352 | `truncate()` -- entire function |
| MODIFY func | `internal/materialize/hooks.go` | 144-219 | `mergeHooksSettings()` -- remove legacy branch, change return type |
| MODIFY caller | `internal/materialize/materialize.go` | 1392-1399 | Adapt to new `mergeHooksSettings()` signature (single return value) |
| DELETE test | `internal/materialize/hooks_test.go` | 527-539 | `TestIsLegacyPlatformHook_CLAUDEProjectDir` |
| DELETE test | `internal/materialize/hooks_test.go` | 541-553 | `TestIsLegacyPlatformHook_DotClaudeHooksPath` |
| DELETE test | `internal/materialize/hooks_test.go` | 555-567 | `TestIsLegacyPlatformHook_ShSuffix` |
| DELETE test | `internal/materialize/hooks_test.go` | 569-581 | `TestIsLegacyPlatformHook_AriShSuffix` |
| DELETE test | `internal/materialize/hooks_test.go` | 583-595 | `TestIsLegacyPlatformHook_UserTool` |
| DELETE test | `internal/materialize/hooks_test.go` | 597-609 | `TestIsLegacyPlatformHook_PythonScript` |
| DELETE test | `internal/materialize/hooks_test.go` | 611-619 | `TestIsLegacyPlatformHook_FlatFormat` |
| DELETE test | `internal/materialize/hooks_test.go` | 621-629 | `TestIsLegacyPlatformHook_FlatFormatUser` |
| DELETE test | `internal/materialize/hooks_test.go` | 631-706 | `TestMergeHooks_StripsLegacyPreservesUser` |
| DELETE test | `internal/materialize/hooks_test.go` | 708-753 | `TestMergeHooks_AllLegacyNoUser` |
| MODIFY test | `internal/materialize/hooks_test.go` | 148-178 | `TestMergeHooksSettings_FreshSettings` -- remove `stripped` from return |
| MODIFY test | `internal/materialize/hooks_test.go` | 180-232 | `TestMergeHooksSettings_PreservesUserHooks` -- remove `stripped` from return |
| MODIFY test | `internal/materialize/hooks_test.go` | 234-275 | `TestMergeHooksSettings_PreservesOldFlatUserHooks` -- remove `stripped` from return |
| MODIFY test | `internal/materialize/hooks_test.go` | 277-319 | `TestMergeHooksSettings_RemovesOldAriHooks` -- remove `stripped` from return |
| MODIFY test | `internal/materialize/hooks_test.go` | 321-347 | `TestMergeHooksSettings_Idempotent` -- remove `stripped` from return |
| CLEAN test | `internal/materialize/hooks_test.go` | 47 | Remove ghost `UserPromptSubmit` entry from `TestBuildHooksSettings` fixture |
| CLEAN test | `internal/materialize/hooks_test.go` | 54 | Remove `"UserPromptSubmit"` from `expectedEvents` slice |
| REMOVE import | `internal/materialize/hooks_test.go` | 8 | Remove `"strings"` import (only used by stripped legacy tests) |

**Detailed `mergeHooksSettings()` changes**:

Current signature (hooks.go:150):
```go
func mergeHooksSettings(existingSettings map[string]any, hooksConfig *HooksConfig) (map[string]any, []string) {
```

New signature:
```go
func mergeHooksSettings(existingSettings map[string]any, hooksConfig *HooksConfig) map[string]any {
```

Remove from the function body:
- `var stripped []string` declaration (line 152)
- The `else if isLegacyPlatformHook(group)` branch (lines 185-189) -- the entire branch including the `extractCommandForReport` call and `stripped = append(...)` line
- `return existingSettings, stripped` becomes `return existingSettings` (line 218)
- `return existingSettings, stripped` at line 158 becomes `return existingSettings`

The three-way classification comment at line 177 changes from "Three-way classification: ari (skip), legacy (strip), user (preserve)" to "Two-way classification: ari (replace), user (preserve)".

**Caller change in materialize.go:1392-1399**:

Current:
```go
if hooksConfig := m.loadHooksConfig(); hooksConfig != nil {
    var stripped []string
    existingSettings, stripped = mergeHooksSettings(existingSettings, hooksConfig)
    // Log stripped legacy hooks if any (for visibility into cleanup)
    for _, msg := range stripped {
        // TODO: Route to structured output when available
        _ = msg // Suppress unused warning; will be used in future structured output
    }
}
```

New:
```go
if hooksConfig := m.loadHooksConfig(); hooksConfig != nil {
    existingSettings = mergeHooksSettings(existingSettings, hooksConfig)
}
```

**Ordering**: Must come after Commit 1 (or be independent -- there is no actual dependency, but logical ordering matters for commit narrative).

**Verification**: `CGO_ENABLED=0 go test ./internal/materialize/...`

### Commit 3: Relocate hooks.yaml to config/ and update embed to []byte

**Scope**: File move, embed type change, path updates across the caller chain.

| Action | File | Detail |
|--------|------|--------|
| CREATE dir | `config/` | Net-new directory at project root |
| MOVE file | `hooks/hooks.yaml` -> `config/hooks.yaml` | The only remaining file in hooks/ |
| DELETE dir | `hooks/` | Now empty after move |
| MODIFY | `embed.go:24-27` | Change embed directive and variable type |
| MODIFY | `internal/cmd/common/embedded.go` | Change `embeddedHooks` type from `fs.FS` to `[]byte` |
| MODIFY | `internal/materialize/materialize.go:96` | Change `embeddedHooks` field type from `fs.FS` to `[]byte` |
| MODIFY | `internal/materialize/materialize.go:136-140` | Change `WithEmbeddedHooks` parameter type |
| MODIFY | `internal/materialize/hooks.go:34-37` | Update resolution order comment |
| MODIFY | `internal/materialize/hooks.go:43` | Update KNOSSOS_HOME path: `"/hooks/hooks.yaml"` -> `"/config/hooks.yaml"` |
| MODIFY | `internal/materialize/hooks.go:47` | Update project root path: `"/hooks/hooks.yaml"` -> `"/config/hooks.yaml"` |
| MODIFY | `internal/materialize/hooks.go:69-80` | Change embedded fallback from `fs.ReadFile` to direct `yaml.Unmarshal` on `[]byte` |
| MODIFY | `cmd/ari/main.go:24` | No change needed -- `knossos.EmbeddedHooks` name stays the same, type changes |
| MODIFY test | `internal/materialize/hooks_test.go:408` | Change test dir from `filepath.Join(tmpDir, "hooks")` to `filepath.Join(tmpDir, "config")` |
| MODIFY test | `internal/materialize/hooks_test.go:422` | Change file write path accordingly |
| MODIFY test | `internal/materialize/hooks_test.go:453-454` | Same dir rename in `TestLoadHooksConfig_RejectsV1Schema` |
| MODIFY test | `internal/materialize/hooks_test.go:463` | Same path update |
| MODIFY test | `internal/materialize/embedded_test.go:215` | Change `m.embeddedHooks` from `fstest.MapFS` to `[]byte` with YAML content |
| MODIFY test | `internal/materialize/embedded_test.go:217` | Direct assignment of `[]byte` instead of FS |
| MODIFY test | `internal/materialize/embedded_test.go:232-277` | `TestEmbeddedHooks_FilesystemOverrides`: update `hooksDir` path and `m.embeddedHooks` type |

**Detailed file changes:**

**embed.go** -- current:
```go
// EmbeddedHooks contains the hooks directory (hooks.yaml and scripts).
//
//go:embed hooks
var EmbeddedHooks embed.FS
```

New:
```go
// EmbeddedHooks contains the canonical hooks.yaml for single-binary distribution.
//
//go:embed config/hooks.yaml
var EmbeddedHooks []byte
```

The `import "embed"` line remains because `EmbeddedRites` and `EmbeddedTemplates` still use `embed.FS`.

**internal/cmd/common/embedded.go** -- current:
```go
var (
    embeddedRites     fs.FS
    embeddedTemplates fs.FS
    embeddedHooks     fs.FS
)

func SetEmbeddedAssets(rites, templates, hooks fs.FS) {
```

New:
```go
var (
    embeddedRites     fs.FS
    embeddedTemplates fs.FS
    embeddedHooks     []byte
)

func SetEmbeddedAssets(rites, templates fs.FS, hooks []byte) {
```

And the accessor:
```go
// EmbeddedHooks returns the embedded hooks YAML bytes, or nil if not set.
func EmbeddedHooks() []byte { return embeddedHooks }
```

**internal/materialize/materialize.go** -- field and method:
```go
// Current (line 96):
embeddedHooks     fs.FS  // Embedded hooks filesystem

// New:
embeddedHooks     []byte // Embedded hooks.yaml content
```

```go
// Current (lines 136-140):
// WithEmbeddedHooks sets the embedded hooks filesystem.
func (m *Materializer) WithEmbeddedHooks(fsys fs.FS) *Materializer {
    m.embeddedHooks = fsys
    return m
}

// New:
// WithEmbeddedHooks sets the embedded hooks.yaml content for single-binary distribution.
func (m *Materializer) WithEmbeddedHooks(data []byte) *Materializer {
    m.embeddedHooks = data
    return m
}
```

**internal/materialize/hooks.go** -- embedded fallback (lines 69-80):
```go
// Current:
if m.embeddedHooks != nil {
    data, err := fs.ReadFile(m.embeddedHooks, "hooks.yaml")
    if err == nil {
        var cfg HooksConfig
        if err := yaml.Unmarshal(data, &cfg); err == nil {
            if cfg.SchemaVersion == "2.0" {
                return &cfg
            }
        }
    }
}

// New:
if len(m.embeddedHooks) > 0 {
    var cfg HooksConfig
    if err := yaml.Unmarshal(m.embeddedHooks, &cfg); err == nil {
        if cfg.SchemaVersion == "2.0" {
            return &cfg
        }
    }
}
```

This also removes the `"io/fs"` import from hooks.go (it was used only for `fs.ReadFile`).

**Caller changes in sync.go, init.go, operations.go**:

All three use the same pattern:
```go
if embHooks := common.EmbeddedHooks(); embHooks != nil {
    m.WithEmbeddedHooks(embHooks)
}
```

With `[]byte`, the nil check works the same way (a nil `[]byte` is falsy). No change needed in these files -- the type flows through naturally because:
- `common.EmbeddedHooks()` returns `[]byte` (was `fs.FS`)
- `mat.WithEmbeddedHooks()` accepts `[]byte` (was `fs.FS`)
- The `nil` check works for both types

**However**, there is one subtlety: Go's `[]byte` nil check with `!= nil` will pass for a zero-length slice. The `//go:embed config/hooks.yaml` directive produces a non-nil, non-empty `[]byte` (the file has content), so this is not a practical issue. But for defensive coding, the `loadHooksConfig` fallback uses `len(m.embeddedHooks) > 0` rather than `m.embeddedHooks != nil`.

**Ordering**: Must come after Commit 1 (hooks/ directory must have only hooks.yaml remaining before the move). Logically should come after Commit 2 as well for clean narrative, but there is no technical dependency on Commit 2.

**Verification**: `CGO_ENABLED=0 go build ./cmd/ari && CGO_ENABLED=0 go test ./... && ari sync --dry-run`

---

## 3. Embed Strategy Detail

### Current embed chain

```
embed.go:       //go:embed hooks           -> var EmbeddedHooks embed.FS
main.go:24:     common.SetEmbeddedAssets(..., knossos.EmbeddedHooks)
embedded.go:    var embeddedHooks fs.FS     -> func EmbeddedHooks() fs.FS
sync.go:153:    m.WithEmbeddedHooks(embHooks)     [embHooks is fs.FS]
init.go:146:    mat.WithEmbeddedHooks(embHooks)   [embHooks is fs.FS]
operations.go:  mat.WithEmbeddedHooks(embHooks)   [embHooks is fs.FS]
materialize.go: embeddedHooks fs.FS               [field on Materializer]
hooks.go:71:    fs.ReadFile(m.embeddedHooks, "hooks.yaml")
```

### New embed chain

```
embed.go:       //go:embed config/hooks.yaml -> var EmbeddedHooks []byte
main.go:24:     common.SetEmbeddedAssets(..., knossos.EmbeddedHooks)  [no change]
embedded.go:    var embeddedHooks []byte      -> func EmbeddedHooks() []byte
sync.go:153:    m.WithEmbeddedHooks(embHooks)     [embHooks is []byte]
init.go:146:    mat.WithEmbeddedHooks(embHooks)   [embHooks is []byte]
operations.go:  mat.WithEmbeddedHooks(embHooks)   [embHooks is []byte]
materialize.go: embeddedHooks []byte              [field on Materializer]
hooks.go:       yaml.Unmarshal(m.embeddedHooks, &cfg)
```

### Type change impact

Files requiring type updates (8 total):

1. `embed.go` -- directive + variable type
2. `internal/cmd/common/embedded.go` -- storage type + setter signature + getter return type
3. `internal/materialize/materialize.go` -- field type + method parameter type
4. `internal/materialize/hooks.go` -- embedded fallback body + remove `"io/fs"` import
5. `cmd/ari/main.go` -- no change (type flows through)
6. `internal/cmd/sync/sync.go` -- no change (type flows through)
7. `internal/cmd/initialize/init.go` -- no change (type flows through)
8. `internal/worktree/operations.go` -- no change (type flows through)

Files 5-8 require zero edits because the variable names and nil-check patterns are identical for `fs.FS` and `[]byte`.

---

## 4. Test Update Catalog

### Tests to DELETE (Commit 2)

| Test Function | File | Lines | Reason |
|---------------|------|-------|--------|
| `TestIsLegacyPlatformHook_CLAUDEProjectDir` | hooks_test.go | 527-539 | Tests deleted function |
| `TestIsLegacyPlatformHook_DotClaudeHooksPath` | hooks_test.go | 541-553 | Tests deleted function |
| `TestIsLegacyPlatformHook_ShSuffix` | hooks_test.go | 555-567 | Tests deleted function |
| `TestIsLegacyPlatformHook_AriShSuffix` | hooks_test.go | 569-581 | Tests deleted function |
| `TestIsLegacyPlatformHook_UserTool` | hooks_test.go | 583-595 | Tests deleted function |
| `TestIsLegacyPlatformHook_PythonScript` | hooks_test.go | 597-609 | Tests deleted function |
| `TestIsLegacyPlatformHook_FlatFormat` | hooks_test.go | 611-619 | Tests deleted function |
| `TestIsLegacyPlatformHook_FlatFormatUser` | hooks_test.go | 621-629 | Tests deleted function |
| `TestMergeHooks_StripsLegacyPreservesUser` | hooks_test.go | 631-706 | Tests legacy stripping path |
| `TestMergeHooks_AllLegacyNoUser` | hooks_test.go | 708-753 | Tests legacy stripping path |

### Tests to MODIFY (Commit 2)

| Test Function | File | Change |
|---------------|------|--------|
| `TestBuildHooksSettings` | hooks_test.go | Remove `UserPromptSubmit` entry (line 47) and from `expectedEvents` (line 54) |
| `TestMergeHooksSettings_FreshSettings` | hooks_test.go | Remove `stripped` from `mergeHooksSettings` return; delete `len(stripped)` check |
| `TestMergeHooksSettings_PreservesUserHooks` | hooks_test.go | Same -- remove `stripped` |
| `TestMergeHooksSettings_PreservesOldFlatUserHooks` | hooks_test.go | Same |
| `TestMergeHooksSettings_RemovesOldAriHooks` | hooks_test.go | Same |
| `TestMergeHooksSettings_Idempotent` | hooks_test.go | Same (uses `_` for stripped already, but call changes) |

### Tests to MODIFY (Commit 3)

| Test Function | File | Change |
|---------------|------|--------|
| `TestLoadHooksConfig` | hooks_test.go | Change `filepath.Join(tmpDir, "hooks")` to `filepath.Join(tmpDir, "config")` |
| `TestLoadHooksConfig_RejectsV1Schema` | hooks_test.go | Same dir rename |
| `TestEmbeddedHooks_Fallback` | embedded_test.go | Change `m.embeddedHooks` from `fstest.MapFS` to `[]byte` YAML content |
| `TestEmbeddedHooks_FilesystemOverrides` | embedded_test.go | Change `hooksDir` to `config`, change `m.embeddedHooks` to `[]byte` |

### Tests UNCHANGED

All other tests in hooks_test.go and embedded_test.go remain unchanged:
- `TestBuildHooksSettings_IncludesTimeout`
- `TestBuildHooksSettings_SkipsEmptyCommand`
- `TestBuildHooksSettings_IncludesAsync`
- `TestBuildHooksSettings_OmitsAsyncWhenFalse`
- `TestIsAriManagedGroup` (5 subtests)
- `TestLoadHooksConfig_NoFile`
- `TestCopyDirFromFS` (3 variants)
- `TestLoadRiteManifest_*` (2 variants)
- `TestMaterializeAgents_FromEmbedded`
- `TestMaterializeMena_FromEmbedded`

---

## 5. Integration Test Matrix

| Satellite Type | Test Scenario | Expected Outcome |
|----------------|---------------|------------------|
| **knossos (self-hosting)** | `ari sync` after all 3 commits | hooks materialized from `config/hooks.yaml`; no legacy stripping messages; settings.local.json hooks section identical to pre-change |
| **Fresh satellite (no .claude/)** | `ari init --rite 10x-dev` | Embedded `[]byte` hooks fallback kicks in; hooks section populated; no errors |
| **Satellite with user hooks** | `ari sync` with custom user hooks in settings.local.json | User hooks preserved; ari hooks replaced; no legacy hooks stripped (because there are none to strip) |
| **Satellite with stale legacy hooks** | `ari sync` with `$CLAUDE_PROJECT_DIR/.claude/hooks/foo.sh` in settings.local.json | Legacy hook is NOT stripped (code is removed). It persists as a "user" hook since it fails the `isAriManagedGroup` check. This is acceptable -- the legacy hooks are inert (the .sh files do not exist) and will be harmless entries. |

**Risk note on row 4**: After this cleanup, any remaining legacy bash hook entries in a satellite's `settings.local.json` will no longer be auto-stripped. They will be preserved as if they were user hooks. This is acceptable because:
1. The `.sh` files they reference do not exist, so CC will fail to execute them (gracefully -- CC logs a warning but continues).
2. The next `ari init --force` or manual edit will clean them up.
3. No satellite has been identified that still has these entries (the migration completed months ago).

---

## 6. Risks and Mitigations

### RISK-1: Stale legacy hooks persist after stripping removal (LOW)

**Impact**: Satellites that still have legacy bash hook entries in `settings.local.json` will no longer have them auto-cleaned.
**Mitigation**: Acceptable degradation. The hooks reference non-existent files and fail gracefully. Document in migration notes.
**Probability**: Very low. ADR-0026 migration completed 2+ months ago.

### RISK-2: `config/` directory naming collision (NONE)

**Impact**: If another Go package or tool expects `config/` at the project root.
**Mitigation**: Verified -- no `config/` directory exists. No Go code references it. The name follows Go project layout conventions (see `github.com/golang-standards/project-layout`).

### RISK-3: Embed directive compile error if config/hooks.yaml missing (LOW)

**Impact**: `//go:embed config/hooks.yaml` will fail at compile time if the file does not exist.
**Mitigation**: This is a feature, not a bug. It ensures the embedded config is never accidentally omitted. The file move in Commit 3 must be atomic with the directive change.

### RISK-4: `io/fs` import removal from hooks.go (LOW)

**Impact**: If any other code in hooks.go uses `io/fs`, removing the import will cause a compile error.
**Mitigation**: Verified -- the only `fs.` reference in hooks.go is `fs.ReadFile` at line 71 (the embedded fallback). No other function uses `io/fs`. Safe to remove.

---

## 7. Documentation Updates (Optional, Non-Blocking)

The following documentation files reference `hooks/hooks.yaml` and will become stale. These are historical design documents and do not affect runtime behavior. They can be updated in a follow-up commit or left as-is.

| File | Priority |
|------|----------|
| `docs/ecosystem/DESIGN-hook-architecture.md` | LOW -- design doc |
| `docs/hygiene/SMELL-distribution-readiness.md` | LOW -- hygiene doc |
| `docs/design/CONTEXT-DESIGN-fate-skills-w4.md` | LOW -- design doc |
| `docs/design/TDD-provenance-manifest.md` | LOW -- test design doc |
| `docs/decisions/ADR-0026-unified-provenance.md` | MEDIUM -- ADR reference |

---

## 8. Changes NOT Made (Explicit Scope Boundaries)

1. **No hook subcommand changes**: The `ari hook route` ghost subcommand (referenced in the deleted test fixture) is a separate concern. The test fixture is cleaned in Commit 2 but the missing subcommand is not added or documented here.
2. **No env var dual-read changes**: Per stakeholder decision, the `CLAUDE_PROJECT_DIR` / `CLAUDE_PLUGIN_ROOT` dual-read pattern in hook env.go is kept as-is.
3. **No new hooks added**: The `route` hook gap (UserPromptSubmit) is accepted. This cleanup does not add new functionality.
4. **No `config/` directory structure design**: This design places only `hooks.yaml` in `config/`. Future additions to `config/` are out of scope.

---

## Attestation Table

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| Context Design | `/Users/tomtenuta/Code/knossos/.wip/DESIGN-hook-cleanup.md` | YES |
| Gap Analysis (input) | `/Users/tomtenuta/Code/knossos/.wip/GAP-hook-cleanup.md` | YES |

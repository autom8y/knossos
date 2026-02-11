# Gap Analysis: Hook Ecosystem Cleanup

**Date**: 2026-02-11
**Analyst**: ecosystem-analyst
**Status**: VERIFIED

---

## 1. Confirmed Safe to Delete

### `.claude/hooks/ari/` (8 files) -- SAFE

| File | Evidence |
|------|----------|
| `autopark.sh` | Zero references outside self and `.wip/` spike docs |
| `clew.sh` | Zero references outside self and `.wip/` spike docs |
| `cognitive-budget.sh` | Zero references outside self and `.wip/` spike docs |
| `context.sh` | Zero references outside self and `.wip/` spike docs |
| `route.sh` | Zero references outside self and `.wip/` spike docs |
| `validate.sh` | Zero references outside self and `.wip/` spike docs |
| `writeguard.sh` | Zero references outside self and `.wip/` spike docs |
| `hooks.yaml` (v1) | Schema 1.0, uses `path:` field. `loadHooksConfig()` rejects v1 at `hooks.go:62` |

**Verification**: Grep for each `.sh` filename across all Go, YAML, MD, and shell files returns hits ONLY in:
- The files themselves (self-referential)
- `.wip/SPIKE-hook-ecosystem-audit.md` (audit document, not code)
- `.wip/CE-AUDIT-hooks.md` (audit document, not code)
- `docs/` reference documentation (historical, not runtime references)

No Go source, no template, no YAML config references these files at runtime.

### `hooks/*.sh` (7 files) -- SAFE with ONE CAVEAT

| File | Evidence |
|------|----------|
| `hooks/autopark.sh` | Dead wrapper. `hooks/hooks.yaml` v2 uses direct `ari hook autopark` |
| `hooks/clew.sh` | Dead wrapper. v2 uses `ari hook clew` |
| `hooks/cognitive-budget.sh` | Dead wrapper. v2 uses `ari hook budget` |
| `hooks/context.sh` | Dead wrapper. v2 uses `ari hook context` |
| `hooks/route.sh` | Dead wrapper. Dispatches to `ari hook route` which does not exist as a subcommand |
| `hooks/validate.sh` | Dead wrapper. v2 uses `ari hook validate` |
| `hooks/writeguard.sh` | Dead wrapper. v2 uses `ari hook writeguard` |

**CAVEAT -- Embed directive**: `/Users/tomtenuta/Code/knossos/embed.go:26-27`:
```go
//go:embed hooks
var EmbeddedHooks embed.FS
```
This embeds the ENTIRE `hooks/` directory into the ari binary, including all 7 dead `.sh` scripts. The embedded hooks fallback (`hooks.go:70-80`) reads `hooks.yaml` from this FS, but the `.sh` files are embedded as dead weight (~7KB unnecessary binary bloat).

**Impact**: Deleting the `.sh` files from `hooks/` will reduce the embedded FS to just `hooks.yaml`, which is the only file the code actually reads. No functional change -- the `.sh` files are never accessed via `fs.ReadFile()`.

---

## 2. References Requiring Update (hooks/hooks.yaml path)

If `hooks/hooks.yaml` moves to `config/hooks.yaml`, the following locations need updating:

### Go Source (MUST update -- runtime paths)

| File | Line(s) | Reference | Type |
|------|---------|-----------|------|
| `internal/materialize/hooks.go` | 36-37 | Comment: resolution order mentions `hooks/hooks.yaml` | Comment |
| `internal/materialize/hooks.go` | 43 | `config.KnossosHome() + "/hooks/hooks.yaml"` | **RUNTIME PATH** |
| `internal/materialize/hooks.go` | 47 | `m.resolver.ProjectRoot()+"/hooks/hooks.yaml"` | **RUNTIME PATH** |
| `internal/materialize/hooks.go` | 71 | `fs.ReadFile(m.embeddedHooks, "hooks.yaml")` | **EMBEDDED FS PATH** |
| `embed.go` | 26-27 | `//go:embed hooks` directive | **EMBED DIRECTIVE** |

### Go Tests (MUST update -- will fail)

| File | Line(s) | Reference |
|------|---------|-----------|
| `internal/materialize/hooks_test.go` | 408-422 | `filepath.Join(tmpDir, "hooks")` creates test dir structure |
| `internal/materialize/embedded_test.go` | 235 | `hooksDir := filepath.Join(tmpDir, "hooks")` |
| `internal/materialize/embedded_test.go` | 215, 245-263 | `m.embeddedHooks` assigned with MapFS using `"hooks.yaml"` key |

### Documentation (SHOULD update -- will be stale but non-breaking)

| File | Line(s) | Reference |
|------|---------|-----------|
| `docs/hygiene/SMELL-distribution-readiness.md` | 275, 281 | References `hooks/hooks.yaml` |
| `docs/design/CONTEXT-DESIGN-fate-skills-w4.md` | 449, 544 | References `hooks/hooks.yaml` |
| `docs/design/TDD-provenance-manifest.md` | 267, 270, 344, 425, 599 | Provenance manifest examples |
| `docs/design/TDD-MENA-SCOPE-PR1.md` | 744, 886 | Hook path references |
| `docs/design/CONTEXT-DESIGN-satellite-hooks-provenance.md` | 521, 543 | Hook definitions reference |
| `docs/decisions/ADR-0026-unified-provenance.md` | 123, 126 | Provenance manifest example |
| `docs/ecosystem/DESIGN-hook-architecture.md` | Multiple | Hook architecture references |
| `.wip/CE-AUDIT-hooks.md` | 5, 22, 348, 439 | Audit references |

### Callers Wiring embeddedHooks (MUST understand for relocation)

| File | Line | Code |
|------|------|------|
| `cmd/ari/main.go` | 24 | `common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooks)` |
| `internal/cmd/sync/sync.go` | 153 | `m.WithEmbeddedHooks(embHooks)` |
| `internal/cmd/initialize/init.go` | 146 | `mat.WithEmbeddedHooks(embHooks)` |
| `internal/worktree/operations.go` | 667 | `mat.WithEmbeddedHooks(embHooks)` |
| `internal/cmd/common/embedded.go` | 10-18, 28 | Storage and accessors for `embeddedHooks` |

---

## 3. Embedded Hooks Verdict

### Current Usage

The `embeddedHooks` mechanism is an `fs.FS` set via `WithEmbeddedHooks()` on the `Materializer`. It serves as a **last-resort fallback** in `loadHooksConfig()` (hooks.go:69-80): if neither `$KNOSSOS_HOME/hooks/hooks.yaml` nor `$PROJECT_ROOT/hooks/hooks.yaml` is found on the filesystem, the embedded copy is used.

### Production Callers

Three production callers wire `embeddedHooks`:
1. `internal/cmd/sync/sync.go:153` -- `ari sync` command
2. `internal/cmd/initialize/init.go:146` -- `ari init` command
3. `internal/worktree/operations.go:667` -- worktree materialization

All three use the identical pattern:
```go
if embHooks := common.EmbeddedHooks(); embHooks != nil {
    m.WithEmbeddedHooks(embHooks)
}
```

### Embed Source

`/Users/tomtenuta/Code/knossos/embed.go:26-27`:
```go
//go:embed hooks
var EmbeddedHooks embed.FS
```

The `knossos.EmbeddedHooks` is set at `cmd/ari/main.go:24` via `common.SetEmbeddedAssets()`.

### Verdict: KEEP the mechanism, but consider relocation impact

The embedded hooks fallback is a **legitimate single-binary distribution feature**. When ari is distributed as a standalone binary to a satellite that does NOT have a knossos checkout, the embedded `hooks.yaml` provides the hook definitions.

If `hooks/hooks.yaml` moves to `config/hooks.yaml`, the embed directive must change:
```go
//go:embed config
var EmbeddedHooks embed.FS
```

And `hooks.go:71` must change from:
```go
fs.ReadFile(m.embeddedHooks, "hooks.yaml")
```
to:
```go
fs.ReadFile(m.embeddedHooks, "hooks.yaml")  // unchanged if config/ has hooks.yaml at root
```

**IMPORTANT**: The `//go:embed config` directive will embed ALL files in `config/`. If other files are added to `config/` later, they will be embedded too. Consider whether a more targeted embed is preferable:
```go
//go:embed config/hooks.yaml
var EmbeddedHooksYAML []byte
```
This would avoid embedding future config files and simplify the FS-to-bytes conversion.

### Test Usage

Two test functions directly set `m.embeddedHooks`:
- `embedded_test.go:215` (`TestEmbeddedHooks_Fallback`)
- `embedded_test.go:263` (`TestEmbeddedHooks_FilesystemOverrides`)

Both use `fstest.MapFS` with key `"hooks.yaml"`. These will NOT need path changes if `hooks.yaml` remains the filename within its parent directory.

---

## 4. Risk Assessment

### RISK-1: Test fixture references ghost subcommand (MEDIUM)

**File**: `internal/materialize/hooks_test.go:47`
```go
{Event: "UserPromptSubmit", Matcher: "^/", Command: "ari hook route --output json", Priority: 3},
```

The test `TestBuildHooksSettings` includes a `UserPromptSubmit` hook entry referencing `ari hook route`, but:
- `ari hook route` is NOT a registered subcommand in `hook.go` (10 subcommands registered, none is `route`)
- `hooks/hooks.yaml` (v2 canonical) does NOT include a UserPromptSubmit entry
- No `route.go` file exists in `internal/cmd/hook/`

This test passes because `buildHooksSettings()` does not validate command existence -- it just formats the config. But it tests a configuration that could never work in production. This is a **test hygiene issue**, not a blocker.

**Recommendation**: Remove the UserPromptSubmit entry from the test fixture when cleaning up, or flag for separate test cleanup.

### RISK-2: Embed directive scope (LOW)

The `//go:embed hooks` directive embeds the entire `hooks/` directory tree. After deleting the `.sh` files, only `hooks.yaml` remains. If the directory is renamed to `config/`, and other config files are added later, they would all be embedded into the binary.

**Mitigation**: Use targeted embed (`//go:embed config/hooks.yaml`) or accept the broader embed scope.

### RISK-3: config/ directory does not exist (LOW)

The `config/` directory does not currently exist at the project root. No Go code references `config/` as a directory path. Creating it requires no migration -- it is a net-new directory.

### RISK-4: Legacy stripping code has active test coverage (LOW)

The functions `isLegacyPlatformHook()` and `extractCommandForReport()` in hooks.go have 7 test cases:
- `TestIsLegacyPlatformHook_CLAUDEProjectDir`
- `TestIsLegacyPlatformHook_DotClaudeHooksPath`
- `TestIsLegacyPlatformHook_ShSuffix`
- `TestIsLegacyPlatformHook_AriShSuffix`
- `TestIsLegacyPlatformHook_UserTool`
- `TestIsLegacyPlatformHook_PythonScript`
- `TestIsLegacyPlatformHook_FlatFormat`
- `TestIsLegacyPlatformHook_FlatFormatUser`
- `TestMergeHooks_StripsLegacyPreservesUser`
- `TestMergeHooks_AllLegacyNoUser`

Removing the legacy stripping code requires removing these tests simultaneously. This is straightforward but must be atomic.

### RISK-5: No hidden dependencies found (CONFIRMED SAFE)

Exhaustive grep confirms:
- No Go source imports or references the `.sh` files
- No template references the `.sh` files
- No YAML config (other than the v1 `hooks.yaml` being deleted) references the `.sh` files
- The `isLegacyPlatformHook()` code only STRIPS legacy hooks from `settings.local.json` -- it does not depend on the `.sh` files existing
- No code path constructs a path to `.claude/hooks/ari/` for reading

---

## 5. Blockers

**None identified.** The 3-commit plan can execute cleanly with these adjustments:

### Commit 1 (Delete orphans): No blockers
- Delete `.claude/hooks/ari/` (8 files)
- Delete `hooks/*.sh` (7 files)
- Both sets have zero external runtime dependencies

### Commit 2 (Remove legacy stripping): No blockers
- Remove `isLegacyPlatformHook()`, `extractCommandForReport()`, `truncate()` from hooks.go
- Simplify `mergeHooksSettings()` to remove the three-way classification (keep only ari vs. user)
- Remove all legacy stripping tests from hooks_test.go
- This is a pure code deletion -- no callers outside hooks.go use these functions

### Commit 3 (Relocate hooks.yaml): Requires coordinated updates
- Move `hooks/hooks.yaml` to `config/hooks.yaml`
- Update `embed.go:26-27`: `//go:embed hooks` to new embed path
- Update `hooks.go:43`: KnossosHome path
- Update `hooks.go:47`: ProjectRoot path
- Update `hooks.go:36-37`: comments
- Verify `hooks.go:71` `fs.ReadFile` key still works (depends on embed structure)
- Update test directory structures in `hooks_test.go:408` and `embedded_test.go:235`
- Optional: clean up test fixture at `hooks_test.go:47` (ghost `ari hook route` entry)
- Optional: update `docs/` references (non-breaking, historical)

---

## 6. Complexity: PATCH

**Rationale**: All three commits are mechanical deletions and path updates. No new logic, no architectural changes, no schema migrations. The legacy stripping removal is subtraction-only. The relocation is a path rename with grep-and-replace updates. Total estimated lines removed: ~200+. Total lines added: ~10 (path changes).

---

## 7. Test Satellite Matrix

| Satellite | Purpose |
|-----------|---------|
| knossos (self-hosting) | Verify `ari sync` still materializes hooks from `config/hooks.yaml` |
| Any satellite with custom user hooks | Verify user hooks preserved after legacy stripping removal |
| Fresh satellite (no existing `.claude/`) | Verify `ari init` uses embedded hooks fallback correctly |

---

## Attestation Table

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| Gap Analysis | `/Users/tomtenuta/Code/knossos/.wip/GAP-hook-cleanup.md` | YES |

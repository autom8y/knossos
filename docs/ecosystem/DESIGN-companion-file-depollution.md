# Context Design: Companion File De-Pollution and Command Namespace Cleanup

**Date**: 2026-02-10
**Author**: Context Architect (ecosystem rite)
**Status**: READY FOR IMPLEMENTATION

## Problem Statement

Three systemic problems discovered in the mena-to-CC projection pipeline:

1. **Companion File Pollution**: 44 companion `.md` files in dromena directories project to `.claude/commands/` where CC discovers each one as a separate slash command. Files like `behavior.md`, `examples.md`, `parking-summary.md` appear as phantom commands (`/session:park:behavior`, `/session:park:examples`) that are not invocable and confuse users.

2. **Deep Path Naming**: Commands at `commands/session/park/INDEX.md` become `/session:park:INDEX` instead of `/park`. The `:INDEX` suffix is unnecessary noise and the `session:` prefix adds verbosity.

3. **context:fork on Session Commands**: All 5 session commands (`/start`, `/park`, `/continue`, `/wrap`, `/fray`) and all 6 operations commands use `context: fork`. For session commands, fork creates an isolated subagent that cannot read the hook-injected session context table, defeating the purpose of these commands which must operate on current session state.

### Scale

| Metric | Count |
|--------|-------|
| INDEX.md files in commands/ (actual commands) | 14 |
| Non-INDEX .md files in commands/ (phantom commands) | 44 |
| Pollution ratio | 3.1x more phantom than real commands |
| Session commands with problematic fork | 5 (/start, /park, /continue, /wrap, /fray) |
| Total commands with fork | 14 (all dromena) |

## Options Considered

### Option A: Flatten output + move companions to skills
**Approach**: Change `ari sync` to project `commands/park.md` (flat, single file) and route companion files to `skills/` instead of `commands/`.

**Pros**: Clean namespace. Users type `/park` not `/session:park:INDEX`.
**Cons**: Requires restructuring mena/ source tree or adding complex path rewriting to the Go pipeline. Companion files in `skills/` break the progressive disclosure pattern -- CC loads skills autonomously but companions are only useful when the parent command is active.

**Verdict**: Rejected. Path rewriting adds fragile complexity and skills routing breaks the content model.

### Option B: Add `user-invocable: false` frontmatter to companions
**Approach**: Add YAML frontmatter with `user-invocable: false` to every companion file. CC respects this field and hides the file from the `/` command menu.

**Pros**: No pipeline changes. No source restructure. Backward compatible. CC-native solution.
**Cons**: 44 files need frontmatter added. Companion files currently have no frontmatter (by design -- they are reference content). Adding frontmatter to non-INDEX files feels like a workaround.

**Verdict**: Rejected as standalone. Adding frontmatter to 44 files purely to suppress CC discovery is maintenance overhead and pollutes reference content with platform concerns.

### Option C: Pipeline injects `user-invocable: false` during projection
**Approach**: Modify `copyDirWithStripping()` in the Go pipeline to detect non-INDEX `.md` files within dromena directories and prepend `user-invocable: false` frontmatter during projection. Source files stay clean.

**Pros**: Source files unchanged. Single point of control. Declarative.
**Cons**: Pipeline modifies content during copy, which violates the current "content passthrough" contract. Makes debugging harder (source differs from output).

**Verdict**: Partially adopted. The pipeline already strips extensions -- injecting frontmatter is a logical extension, but we can do better.

### Option D: Redirect dromena companions to skills (pipeline-level)
**Approach**: Modify `ProjectMena()` to split dromena directories: route `INDEX.dro.md` to `commands/`, route non-INDEX companions to `skills/` under a parallel path. For `mena/session/park/`, the INDEX goes to `commands/session/park/INDEX.md` and `behavior.md` + `examples.md` go to `skills/session/park/behavior.md` + `skills/session/park/examples.md`.

**Pros**: Clean separation. Commands are commands, reference is reference. CC only discovers INDEX files as commands. Companion content remains available as skills for model autonomous loading.
**Cons**: Breaks the "directory = atomic unit" model. INDEX.md loses its `../behavior.md` relative references. Progressive disclosure breaks because the forked subagent loads command context, not separate skills.

**Verdict**: Rejected. Breaking the directory cohesion model has cascading effects on how commands reference their companions.

### Option E (Selected): Pipeline frontmatter injection + namespace flattening + selective fork removal

**Approach**: Three coordinated changes:

1. **Frontmatter injection**: During `copyDirWithStripping()`, detect non-INDEX `.md` files in dromena directories and prepend `---\nuser-invocable: false\n---\n` to their content. Source files stay clean.

2. **Namespace flattening**: For dromena directories whose INDEX has a `name` field in frontmatter, project to `commands/{name}/` instead of `commands/{source-path}/`. This means `mena/session/park/INDEX.dro.md` (with `name: park`) projects to `commands/park/INDEX.md` instead of `commands/session/park/INDEX.md`, giving users `/park` instead of `/session:park:INDEX`.

3. **Selective fork removal**: Remove `context: fork` from session management commands (`/park`, `/continue`, `/wrap`, `/fray`) that need access to hook-injected session context. Keep `context: fork` on commands that genuinely benefit from isolation (`/start`, `/commit`, `/pr`, `/code-review`, `/qa`, `/consult`, `/hotfix`, `/spike`, `/worktree`).

**Pros**: Source files unchanged (injection at pipeline). Clean command namespace. Session commands work correctly with hook context. Each change is independently deployable.
**Cons**: Pipeline now modifies content during copy (a new responsibility). Namespace flattening could cause collisions if two dromena have the same `name`.

**Verdict**: Selected. Best combination of correctness, backward compatibility, and implementation simplicity.

## Detailed Design

### Component 1: Frontmatter Injection for Companion Files

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go`

**Function changes**:

`copyDirWithStripping()` (line 427) -- Add content transformation for non-INDEX `.md` files in dromena directories. The function already walks files and strips extensions. Add a check: if the file is `.md`, is NOT an INDEX file, and is inside a dromena directory (determined by the caller context), prepend `user-invocable: false` frontmatter.

Implementation approach: Add a new parameter `injectHideFrontmatter bool` to `copyDirWithStripping()`. When true and the file is a non-INDEX `.md` file that lacks existing frontmatter, prepend:

```
---
user-invocable: false
---

```

If the file already has frontmatter (starts with `---\n`), inject `user-invocable: false` into the existing frontmatter block instead of adding a new one.

**Callers to update**: The call site in `ProjectMena()` (line 291) passes `injectHideFrontmatter: true` for dromena directories and `false` for legomena directories. Skills companion files are NOT hidden because CC does not discover them as commands.

**Embedded FS variant**: `copyDirFromFSWithStripping()` (line 462) receives the same parameter and applies the same transformation.

**Schema for injected frontmatter**:

```yaml
# Injected by pipeline for non-INDEX .md files in dromena directories.
# CC reads this field and hides the file from the / command menu.
user-invocable: false
```

**Validation**: `user-invocable` is a CC-native field. No custom schema needed. CC treats any `.md` file without this field (or with `user-invocable: true`) as discoverable.

### Component 2: Namespace Flattening

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go`

**Change**: In the `ProjectMena()` function, after collecting entries (Pass 1) and before routing (Pass 2), add a namespace resolution pass for dromena entries.

For each dromena entry:
1. Read the INDEX file's frontmatter using `ReadMenaFrontmatterFromDir()`
2. If `name` field is present and non-empty, use `name` as the destination directory name instead of the source path
3. Check for collisions: if two dromena resolve to the same `name`, log a warning and fall back to the full path for the colliding entry

**Example transformation**:

| Source path | Frontmatter name | Output path (before) | Output path (after) |
|-------------|-------------------|----------------------|---------------------|
| `mena/session/park/` | `park` | `commands/session/park/` | `commands/park/` |
| `mena/operations/commit/` | `commit` | `commands/operations/commit/` | `commands/commit/` |
| `mena/rite-switching/10x` | (standalone) | `commands/rite-switching/10x.md` | `commands/rite-switching/10x.md` |
| `mena/navigation/rite` | (standalone) | `commands/navigation/rite.md` | `commands/navigation/rite.md` |

Standalone files (not in INDEX directories) are NOT flattened because they have no frontmatter `name` field -- their filename IS the command name already.

**Collision detection algorithm**:

```
nameMap := map[string][]string{}  // name -> list of source paths
for name, entry := range collected {
    if menaType == "dro" {
        fm := ReadMenaFrontmatterFromDir(entry.source.Path)
        if fm.Name != "" {
            nameMap[fm.Name] = append(nameMap[fm.Name], name)
        }
    }
}
// Any name with len > 1 is a collision -> fall back to source path
```

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/frontmatter.go`

No changes needed. `ReadMenaFrontmatterFromDir()` already reads and parses frontmatter from INDEX files.

### Component 3: Selective Fork Removal

**Files**: Source mena/ files (frontmatter changes only)

Remove `context: fork` from these session commands where fork prevents hook context visibility:

| File | Current | Change | Rationale |
|------|---------|--------|-----------|
| `mena/session/park/INDEX.dro.md` | `context: fork` | Remove line | Must read hook-injected session context table |
| `mena/session/continue/INDEX.dro.md` | `context: fork` | Remove line | Must read hook-injected session context table |
| `mena/session/wrap/INDEX.dro.md` | `context: fork` | Remove line | Must read hook-injected session context table |
| `mena/session/fray/INDEX.dro.md` | `context: fork` | Remove line | Must read hook-injected session context table |
| `mena/session/handoff/INDEX.dro.md` | `context: fork` | Remove line | Must read hook-injected session context table |

Keep `context: fork` on these commands where isolation IS beneficial:

| File | Rationale for keeping fork |
|------|---------------------------|
| `mena/session/start/INDEX.dro.md` | Spawns new session; isolation prevents contamination from prior context |
| `mena/operations/commit/INDEX.dro.md` | Git operations benefit from clean context |
| `mena/operations/pr/INDEX.dro.md` | GitHub operations benefit from clean context |
| `mena/operations/code-review/INDEX.dro.md` | Needs fresh analysis without prior conversation bias |
| `mena/operations/qa/INDEX.dro.md` | Adversarial testing requires isolation from implementation context |
| `mena/operations/spike/INDEX.dro.md` | Exploration benefits from clean slate |
| `mena/navigation/consult/INDEX.dro.md` | Advisory context should be isolated |
| `mena/navigation/worktree/INDEX.dro.md` | Git worktree operations benefit from isolation |
| `mena/workflow/hotfix/INDEX.dro.md` | Hotfix isolation prevents contamination |

**Design rationale**: The distinction is whether the command READS existing session state (no fork) or CREATES new state from scratch (keep fork). Session lifecycle commands (`park`, `continue`, `wrap`, `fray`, `handoff`) all need to read the session context table injected by the SessionStart hook. That table is only visible in the main conversation context, not in a forked subagent.

`/start` is a special case: it reads the session context table to detect existing sessions, but its primary purpose is creating a NEW session. The fork isolation is more valuable here because `/start` delegates to the entry agent and benefits from a clean context. The session detection is a pre-flight check that can be handled via CLI fallback (`ari session status`).

## Backward Compatibility

### Classification: COMPATIBLE (with namespace shift)

**Frontmatter injection (Component 1)**: Strictly additive. Companion files gain a frontmatter block that hides them from CC command discovery. No existing behavior changes -- these files were never intentionally invoked as commands. The injected frontmatter does not affect file content below the frontmatter block.

**Namespace flattening (Component 2)**: This is a VISIBLE change. Users who type `/session:park:INDEX` will find the command has moved to `/park`. However:
- The old path `/session:park:INDEX` is unlikely to be typed manually (it is ugly and verbose)
- CC auto-discovery uses the new flat path automatically
- No automation scripts reference these paths (commands are interactive)

Migration approach: The old directories are cleaned on `ari sync` (destructive mode wipes per-entry subdirs before writing). No stale paths persist.

**Fork removal (Component 3)**: Behavioral change for 5 session commands. Commands that previously ran in isolated context now run in the main conversation context. This is the DESIRED behavior -- the previous fork was causing breakage, not providing value.

### Impact on satellites

Satellites that have already materialized `.claude/commands/` will see a one-time restructuring on next `ari sync`:
- Old nested paths (`commands/session/park/`) are removed
- New flat paths (`commands/park/`) are created
- Companion files gain `user-invocable: false` frontmatter

This is handled by the existing destructive mode in `ProjectMena()` which calls `os.RemoveAll(destDir)` before writing each entry (line 277-279 of project_mena.go).

**User-created commands**: Unaffected. Destructive mode only removes knossos-managed entries (those present in the collected set). User-created `.md` files in `commands/` are preserved.

## File-Level Change Specification

### Go Pipeline Files

| File | Function | Change |
|------|----------|--------|
| `internal/materialize/project_mena.go` | `copyDirWithStripping()` | Add `injectHideFrontmatter bool` parameter. When true and file is non-INDEX .md, prepend `user-invocable: false` frontmatter. |
| `internal/materialize/project_mena.go` | `copyDirFromFSWithStripping()` | Same `injectHideFrontmatter` parameter for embedded FS variant. |
| `internal/materialize/project_mena.go` | `ProjectMena()` Pass 2 loop | Resolve flat namespace: read frontmatter name, compute `destDir` from name instead of source path for dromena. Add collision detection. Pass `injectHideFrontmatter: true` for dro, `false` for lego. |
| `internal/materialize/project_mena.go` | New function `injectHideFrontmatter()` | Helper: prepend or inject `user-invocable: false` into .md content bytes. |
| `internal/materialize/project_mena.go` | New function `resolveNamespace()` | Helper: build name->path map from frontmatter, detect collisions. |

### Mena Source Files (frontmatter changes only)

| File | Change |
|------|--------|
| `mena/session/park/INDEX.dro.md` | Remove `context: fork` line |
| `mena/session/continue/INDEX.dro.md` | Remove `context: fork` line |
| `mena/session/wrap/INDEX.dro.md` | Remove `context: fork` line |
| `mena/session/fray/INDEX.dro.md` | Remove `context: fork` line |
| `mena/session/handoff/INDEX.dro.md` | Remove `context: fork` line |

### No changes required

| File | Reason |
|------|--------|
| `internal/materialize/frontmatter.go` | `ReadMenaFrontmatterFromDir()` already reads name field |
| `internal/materialize/materialize.go` | No changes to pipeline orchestration |
| `internal/materialize/source.go` | Source resolution unaffected |
| All companion `.md` files in mena/ | Source files stay clean; injection happens at pipeline |

## Integration Test Matrix

### Existing tests that must continue passing

| Test | File | Validates |
|------|------|-----------|
| `TestRoutingDroToCommands` | routing_test.go | Dromena route to commands/ |
| `TestRoutingLegoToSkills` | routing_test.go | Legomena route to skills/ |
| `TestRoutingSupportingFilesFollowIndex` | routing_test.go | Companion files follow INDEX routing |
| `TestProjectMena_Destructive` | project_mena_test.go | Destructive mode with user preservation |
| `TestProjectMena_PriorityOverride` | project_mena_test.go | Source priority ordering |
| `TestProjectMena_EmbeddedFS` | project_mena_test.go | Embedded FS projection |

### New tests to add

| Test | File | Input | Expected Outcome |
|------|------|-------|------------------|
| `TestCompanionFileHideFrontmatter` | project_mena_test.go | Dromena dir with INDEX.dro.md + behavior.md + examples.md | behavior.md and examples.md in output contain `user-invocable: false` frontmatter. INDEX.md does NOT have injected frontmatter. |
| `TestCompanionFilePreservesExistingFrontmatter` | project_mena_test.go | Companion .md file with existing `---\ntitle: Foo\n---` frontmatter | `user-invocable: false` injected into existing frontmatter block (not a second block). |
| `TestLegoCompanionsNotInjected` | project_mena_test.go | Legomena dir with INDEX.lego.md + helper.md | helper.md in skills/ output does NOT have injected frontmatter (CC does not discover skills as commands). |
| `TestNamespaceFlattening` | project_mena_test.go | Dromena at `mena/session/park/INDEX.dro.md` with `name: park` | Output at `commands/park/INDEX.md`, NOT at `commands/session/park/INDEX.md`. |
| `TestNamespaceCollisionFallback` | project_mena_test.go | Two dromena dirs with same `name: park` in frontmatter | Warning logged. Second entry falls back to full source path. |
| `TestNamespaceFlatteningWithCompanions` | project_mena_test.go | Dromena dir with name, INDEX, and companions | All files (INDEX + companions) at flat path. Companions have hide frontmatter. |
| `TestStandaloneFilesNotFlattened` | project_mena_test.go | Standalone `rite.dro.md` in grouping dir | Output preserves grouping path (no frontmatter name to flatten with). |
| `TestEmbeddedFSFlatteningAndInjection` | project_mena_test.go | Embedded FS with dromena INDEX + companion | Flat namespace + hide frontmatter in embedded projection. |

### Satellite diversity tests

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| Minimal (no local settings) | Sync with updated pipeline | Commands flatten correctly. No errors. No orphaned paths. |
| Standard (typical project) | Sync with existing commands/ | Old nested paths cleaned. New flat paths created. User-created commands preserved. |
| Complex (custom user commands + overrides) | Sync with user commands in commands/ | User-created .md files unaffected. Knossos paths flatten. No collisions with user commands. |
| Knossos self-host | Sync on knossos itself | All 14 dromena flatten to top-level names. 44 companions gain hide frontmatter. 5 session commands lose fork. |

## Implementation Sequence

Ordered by dependency:

### Phase 1: Frontmatter injection (Component 1)
**Why first**: Immediately fixes the 44 phantom command problem without any namespace changes. Independently deployable.

1. Add `injectHideFrontmatter()` helper function
2. Modify `copyDirWithStripping()` signature and logic
3. Modify `copyDirFromFSWithStripping()` signature and logic
4. Update callers in `ProjectMena()` to pass dro=true, lego=false
5. Add tests: `TestCompanionFileHideFrontmatter`, `TestCompanionFilePreservesExistingFrontmatter`, `TestLegoCompanionsNotInjected`
6. Verify all existing tests pass

### Phase 2: Fork removal (Component 3)
**Why second**: Pure source file change, no pipeline code. Independent of Phase 1.

1. Edit 5 mena source files to remove `context: fork`
2. Run `ari sync` to verify commands project correctly
3. Manual verification: `/park` in CC sees session context table

### Phase 3: Namespace flattening (Component 2)
**Why third**: Most complex change with collision detection. Depends on Phase 1 being stable.

1. Add `resolveNamespace()` helper function
2. Modify `ProjectMena()` Pass 2 to use resolved names for dromena
3. Add collision detection and fallback
4. Add tests: `TestNamespaceFlattening`, `TestNamespaceCollisionFallback`, `TestNamespaceFlatteningWithCompanions`, `TestStandaloneFilesNotFlattened`
5. Verify all existing tests pass

## Design Decisions Log

| Decision | Rationale | Alternatives Rejected |
|----------|-----------|----------------------|
| Inject frontmatter at pipeline, not source | Source files stay clean. Single point of control. Pipeline already transforms content (extension stripping). | Adding frontmatter to 44 source files: maintenance burden, pollutes reference content. |
| Use `user-invocable: false` not `hidden: true` | `user-invocable` is the CC-native field per Anthropic documentation. `hidden` is not a recognized field. | Custom field name: would be ignored by CC. |
| Flatten by frontmatter `name` not by convention | Explicit is better than implicit. The `name` field already exists in all INDEX files. Convention-based flattening (strip first path segment) is fragile. | Path segment stripping: breaks for commands legitimately nested (e.g., rite-switching/10x where no frontmatter name exists). |
| Remove fork from park/continue/wrap/fray/handoff but keep on start | Park/continue/wrap/fray/handoff must READ session context. Start CREATES context and benefits from isolation. | Remove fork from all: /start delegates to entry agent and benefits from clean context. Keep fork on all: session commands remain broken. |
| Standalone files not flattened | Standalone .dro.md files (`rite.dro.md`, `sessions.dro.md`) have no INDEX frontmatter to read names from. Their filename IS the command name. Flattening them makes no sense. | Flatten everything: would require adding frontmatter to standalone files. |
| Collision falls back to full path (not error) | Collisions are configuration bugs, not runtime errors. The pipeline should degrade gracefully with a warning, not block materialization. | Hard error on collision: too strict for framework. Silent override: data loss risk. |

## Provenance Impact

The provenance system tracks commands/ entries by their relative path within `.claude/`. After namespace flattening:

- Old entry key: `commands/session/park/`
- New entry key: `commands/park/`

On the first sync after this change, the old provenance entry becomes stale (no matching directory on disk). The `saveProvenanceManifest()` carry-forward logic (Step 0, line 1570-1588) checks `os.Stat(fullPath)` and drops entries for directories that no longer exist. The new flat path gets a fresh provenance entry from the collector.

No provenance migration code needed. The existing manifest merge algorithm handles this naturally.

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| CC does not respect `user-invocable: false` on companion files | Low (documented CC feature) | High (phantom commands persist) | Test with CC immediately after Phase 1. Fallback: move companions to skills/ (Option D). |
| Namespace collision between dromena `name` and user-created command | Low (user commands have different names) | Medium (user command shadowed) | Collision detection warns in pipeline output. Provenance tracks ownership. |
| Fork removal causes session commands to pollute conversation context | Medium | Low (commands are short-lived) | Session commands already use `disable-model-invocation: true` which limits auto-invocation. Users invoke these explicitly. |
| Embedded FS path differences break flattening | Low | Medium (embedded builds break) | Test `TestEmbeddedFSFlatteningAndInjection` covers this path. Embedded sources use `ReadMenaFrontmatterFromDir()` equivalent via fs.FS. |

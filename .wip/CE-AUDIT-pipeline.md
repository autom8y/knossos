# CE Audit: Materialization Pipeline

## Summary

| Metric | Value |
|--------|-------|
| Pipeline stages | 10 (resolve, orphan-detect, agents, mena, rules, CLAUDE.md, settings, state, workflow, ACTIVE_RITE) |
| Source types handled | agents (.md), mena (.dro.md/.lego.md), rules (.md), templates (.md.tpl), hooks (.yaml), MCP servers (manifest.yaml), workflow (.yaml) |
| Go files in scope (production) | 24 across 4 packages |
| Production LOC | ~5,450 (materialize: 3,751, inscription: 3,100 (est. excl. tests), provenance: 470, frontmatter: 100) |
| Test LOC | ~11,366 across 25 test files |
| Test-to-production ratio | ~2.1:1 (healthy) |

### Packages

| Package | Role | Production Files | LOC |
|---------|------|-----------------|-----|
| `internal/materialize/` | Orchestration, source resolution, file projection | 9 | 3,751 |
| `internal/inscription/` | CLAUDE.md templating, marker parsing, region merging | 8 | ~3,100 |
| `internal/provenance/` | Ownership tracking, divergence detection, manifest I/O | 4 | 470 |
| `internal/frontmatter/` | YAML frontmatter parsing, FlexibleStringSlice | 2 | 100 |


## Pipeline Flow

### Entry Point: `Materializer.Sync(SyncOptions)`

The unified entry point dispatches to two independent scopes:

```
ari sync
  |
  +-> Phase 1: Rite Scope (syncRiteScope)
  |     |
  |     +-> Read ACTIVE_RITE or --rite flag
  |     +-> MaterializeWithOptions(riteName, opts) -- 10-step pipeline
  |     |     |
  |     |     1. ResolveRite (5-tier: explicit > project > user > knossos > embedded)
  |     |     2. EnsureDir .claude/
  |     |     3. Load prevManifest, detect divergence, create Collector
  |     |     4. clearInvocationState (stale INVOCATION_STATE.yaml)
  |     |     5. detectOrphans + orphan action (keep/remove/promote)
  |     |     6. materializeAgents (selective write: knossos-owned replaced, user preserved)
  |     |     7. materializeMena (4-tier mena source priority -> ProjectMena)
  |     |     8. materializeRules (template rules replace knossos-owned, user rules preserved)
  |     |     9. materializeCLAUDEmd (inscription: generate -> merge -> write)
  |     |    10. materializeSettingsWithManifest (hooks + MCP merge)
  |     |    11. trackState (sync/state.json)
  |     |    12. materializeWorkflow (ACTIVE_WORKFLOW.yaml)
  |     |    13. writeActiveRite (ACTIVE_RITE marker)
  |     |    14. saveProvenanceManifest (4-step merge algorithm)
  |     |
  |     +-> OR MaterializeMinimal (cross-cutting mode, no rite)
  |
  +-> Phase 2: User Scope (syncUserScope)
        |
        +-> Resolve KNOSSOS_HOME
        +-> Load USER_PROVENANCE_MANIFEST.yaml
        +-> Initialize CollisionChecker (reads rite PROVENANCE_MANIFEST)
        +-> For each resource type (agents, mena, hooks):
        |     +-> Snapshot existing knossos-owned entries
        |     +-> Walk source directory, sync files
        |     +-> Phase 3: remove orphaned knossos-owned entries
        +-> Save manifest, cleanup old JSON manifests
```

### Inscription Sub-Pipeline (CLAUDE.md Generation)

```
materializeCLAUDEmd
  |
  +-> Build RenderContext (agents, rite name, vars)
  +-> mergeCLAUDEmd
        |
        +-> Load/create KNOSSOS_MANIFEST.yaml (inscription manifest)
        +-> Create Generator (templates FS or filesystem)
        +-> GenerateAll() -- iterate SectionOrder, render each region
        |     +-> OwnerKnossos: render from template file or Go default
        |     +-> OwnerRegenerate: generate from source (ACTIVE_RITE, agents/)
        |     +-> OwnerSatellite: render template for new files (preserve existing)
        +-> Read existing CLAUDE.md
        +-> Legacy detection: backup if no KNOSSOS markers found
        +-> NewMerger -> MergeRegions
        |     +-> Parse existing with MarkerParser
        |     +-> Process sections in manifest SectionOrder
        |     +-> Per-region merge: satellite preserved, knossos overwritten, regenerate conditional
        |     +-> Append unknown regions as satellite (skip deprecated)
        |     +-> Clean deprecated regions from manifest
        +-> writeIfChanged (atomic, avoids file watcher triggers)
        +-> UpdateManifestHashes -> Save KNOSSOS_MANIFEST.yaml
```

### Provenance Merge Algorithm (saveProvenanceManifest)

```
Step 0: Carry forward knossos entries from previous manifest (still on disk, not rewritten)
Step 1: Layer promoted (user-owned) + carried-forward entries from divergence detection
Step 2: Layer current sync entries (pipeline-written), skip user-promoted paths
Step 3: Resolve untracked entries -> promote to user
Final: Build ProvenanceManifest, Save()
```


## Critical Findings

### C-1: Provenance records only on write, not on no-op

**Severity: CRITICAL**
**Location**: `materialize.go` lines 756-825 (materializeAgents), 896-998 (materializeMena)

The provenance collector only calls `Record()` when `writeIfChanged()` returns `written=true`. On idempotent re-runs where files have not changed, `writeIfChanged()` returns `false` and the collector gets zero entries for unchanged files. The `saveProvenanceManifest()` Step 0 carries forward knossos entries from the previous manifest to compensate, but this creates a dependency on the previous manifest existing and being valid.

**Impact**: If `PROVENANCE_MANIFEST.yaml` is deleted (or never existed on first run after adopting provenance), and then `ari sync` is run twice:
- First run: all files written, all provenance recorded correctly
- If manifest is then deleted manually
- Second run: files unchanged (writeIfChanged=false), collector empty, Step 0 has no previous manifest to carry from. Result: provenance manifest with zero entries despite .claude/ being fully populated.

**Mitigation**: The pipeline compensates via Step 0 carry-forward, and the `LoadOrBootstrap` pattern bootstraps empty manifests. The scenario requires manual manifest deletion which is an edge case. But the architectural coupling between "did we write the file" and "do we track it" is fragile.

### C-2: Two parallel manifest systems with different ownership semantics

**Severity: CRITICAL (architectural)**
**Location**: `internal/inscription/types.go` vs `internal/provenance/provenance.go`

The pipeline maintains two independent manifest systems with overlapping but incompatible ownership models:

| Concern | KNOSSOS_MANIFEST.yaml (inscription) | PROVENANCE_MANIFEST.yaml (provenance) |
|---------|--------------------------------------|---------------------------------------|
| Tracks | CLAUDE.md region ownership | All .claude/ file ownership |
| Owners | knossos / satellite / regenerate | knossos / user / untracked |
| Scope | CLAUDE.md sections only | agents, commands, skills, rules, CLAUDE.md, settings |
| Hash | Per-region content hash | Per-file SHA256 with prefix |
| Written by | inscription ManifestLoader.Save() | provenance.Save() |
| Volatile | inscription_version, last_sync | last_sync, per-entry last_synced |

Both manifests track CLAUDE.md ownership: inscription tracks it at region granularity, provenance tracks it at file granularity. This creates a conceptual overlap where CLAUDE.md has dual provenance tracking. They do not conflict in practice (inscription drives merge behavior, provenance drives file-level ownership), but the two-manifest architecture increases cognitive load and maintenance surface.


## High Findings

### H-1: Mena projection provenance is post-hoc directory scan, not per-file tracking

**Severity: HIGH**
**Location**: `materialize.go` lines 896-998

After `ProjectMena()` returns, the pipeline scans the output `commands/` and `skills/` directories to record provenance at directory granularity (`commands/commit/`, `skills/lexicon/`). The source path detection iterates sources in reverse priority and checks filesystem existence. This is a heuristic that can produce incorrect source attribution when:

1. Multiple sources contribute files to the same command directory (shared + rite-specific)
2. The heuristic breaks on first match, missing that the actual winning source was different
3. Embedded FS sources get generic `sourceType` instead of precise path

The `ProjectMena()` function itself returns `MenaProjectionResult` with lists of projected paths but no source attribution, forcing the caller to reverse-engineer provenance.

**Recommendation**: Thread the provenance collector into `ProjectMena()` so it records provenance at write time with exact source information, rather than guessing after the fact.

### H-2: materializeAgents unconditionally removes and rewrites knossos-managed agents

**Severity: HIGH**
**Location**: `materialize.go` lines 720-728

```go
// Remove only knossos-managed agents (will be rewritten below).
if entries, err := os.ReadDir(agentsDir); err == nil {
    for _, entry := range entries {
        if !entry.IsDir() && managedAgents[entry.Name()] {
            os.Remove(filepath.Join(agentsDir, entry.Name()))
        }
    }
}
```

This deletes knossos-managed agent files before rewriting them. Even though `writeIfChanged()` is used for the rewrite, the delete+write cycle means there is a window where the file does not exist on disk. If the process crashes between delete and write, the agent file is lost. The `writeIfChanged()` pattern used elsewhere avoids this by writing atomically without a prior delete.

**Recommendation**: Remove the pre-deletion loop. Let `writeIfChanged()` handle the overwrite atomically. The delete was presumably added to handle the case where a rite removes an agent, but orphan detection already handles that case.

### H-3: Silent error swallowing in mena provenance scanning

**Severity: HIGH**
**Location**: `materialize.go` lines 907-998

```go
if entries, err := os.ReadDir(commandsDir); err == nil {
    for _, entry := range entries {
        // ...
        hash, err := checksum.Dir(dirPath)
        if err != nil {
            continue  // silently skipped
        }
```

If `checksum.Dir()` fails (permissions, symlink loops, etc.), the entry is silently skipped from provenance tracking. No warning, no error propagation. This means provenance can have missing entries with no diagnostic trail.

### H-4: Inscription Pipeline.Sync() and materializeCLAUDEmd are parallel code paths

**Severity: HIGH**
**Location**: `internal/inscription/pipeline.go` (Sync method) vs `materialize.go` (materializeCLAUDEmd)

The `Pipeline.Sync()` method in the inscription package is a standalone entry point that duplicates the logic in `materializeCLAUDEmd()`. Both do: load manifest -> build context -> generate sections -> merge -> write -> update hashes. But `materializeCLAUDEmd` is the one actually called by the materialization pipeline, while `Pipeline.Sync()` appears to be a standalone entry point (possibly for `ari inscription sync`).

Key differences:
- `Pipeline.Sync()` adds a file header (`# CLAUDE.md\n\n> Entry point...`) via `buildFinalContent()`
- `materializeCLAUDEmd()` does not add this header
- `Pipeline.Sync()` creates backups; `materializeCLAUDEmd()` only backs up legacy format
- `Pipeline.Sync()` builds its own render context from .claude/agents/; `materializeCLAUDEmd()` receives context from the manifest

This duplication risks drift: a fix applied to one path may not reach the other.


## Medium Findings

### M-1: Divergence detection computes checksums for all knossos entries on every sync

**Severity: MEDIUM**
**Location**: `internal/provenance/divergence.go` lines 39-81

`DetectDivergence()` iterates all previous manifest entries and computes fresh checksums for every knossos-owned file. For a .claude/ directory with 30+ entries, this means 30+ file reads and SHA256 computations on every `ari sync`. For mena directory entries, `checksum.Dir()` walks the entire directory tree.

This is O(n) I/O per sync where n is the number of tracked entries. Currently manageable but will scale linearly as the number of tracked entries grows.

### M-2: `copyUserFile` does not use `writeIfChanged()`

**Severity: MEDIUM**
**Location**: `user_scope.go` line 595

The user scope sync uses `os.WriteFile()` directly via `copyUserFile()`, bypassing the `writeIfChanged()` optimization that the rite scope uses. This means user scope syncs will trigger CC's file watcher even when file content is identical.

### M-3: Orphan backup writes directly with os.WriteFile

**Severity: MEDIUM**
**Location**: `materialize.go` line 628

```go
if err := os.WriteFile(dstPath, content, 0644); err != nil {
```

The orphan backup in `backupAndRemoveOrphans` uses `os.WriteFile` instead of atomic write. While this is a backup directory (not .claude/ proper), it does not follow the project's pattern of atomic writes for safety.

### M-4: `isValidSchemaVersion` is defined in two packages

**Severity: MEDIUM**
**Location**: `internal/inscription/manifest.go` and `internal/provenance/manifest.go`

Both packages define their own `isValidSchemaVersion()` function. The inscription version uses a hand-rolled character-by-character parser; the provenance version uses a regex. Both validate "N.N" format but with subtly different implementations. This is a DRY violation that could lead to inconsistent validation.

### M-5: DefaultSectionOrder comment says "Team context" (terminology debt)

**Severity: MEDIUM**
**Location**: `internal/inscription/manifest.go` line 240

```go
// Team context (who is available)
"quick-start",
```

Despite the SL-008 terminology cleanse, this comment still uses "Team context" instead of "Rite context". This is in generated infrastructure code, not a user-facing string, but it contradicts the project's terminology standards.

### M-6: Sprig dependency for template functions

**Severity: MEDIUM (dependency hygiene)**
**Location**: `internal/inscription/generator.go` line 253

The generator imports `github.com/Masterminds/sprig/v3` for template functions. This adds 100+ template functions to every render. The templates in `knossos/templates/sections/` are simple markdown with basic conditionals and variable substitution. The Sprig dependency is heavyweight for the actual usage. If any templates break because of a Sprig version change, it would be a non-obvious failure.

### M-7: provenance.Save validates on every write

**Severity: MEDIUM (performance)**
**Location**: `internal/provenance/manifest.go` line 54

`Save()` calls `validateManifest()` which uses regex matching on every checksum string for every entry. With 30 entries, that is 30 regex compilations per save (the regex is not compiled to a package-level var).


## Low Findings

### L-1: `itoa` reimplemented in inscription/marker.go

The `itoa` function is a hand-rolled integer-to-string converter that avoids importing `strconv`. The same package (`inscription/manifest.go`) already imports `strconv` for `Atoi/Itoa`. The hand-rolled version is used in marker.go and generator.go. This is unnecessary complexity.

### L-2: `simpleDiff` in Pipeline uses O(n*m) line comparison

**Location**: `internal/inscription/pipeline.go` lines 693-737

The `simpleDiff` function checks membership with a linear scan (`contains()`) for every line, making it O(n*m). For large CLAUDE.md files this could be slow, though in practice CLAUDE.md files are small (<200 lines).

### L-3: `parseVersion` in pipeline.go silently accepts non-numeric suffixes

**Location**: `internal/inscription/pipeline.go` lines 761-771

`parseVersion("42abc")` returns 42 with no error. This is lenient parsing that could mask issues.

### L-4: Legacy manifest cleanup is best-effort

**Location**: `user_scope.go` lines 655-673

The `cleanupOldManifests` function creates `.v2-backup` files and deletes originals. If the backup write fails, the original is still deleted. The `os.WriteFile` error is discarded.

### L-5: `codeBlockStartRegex` and `codeBlockEndRegex` are separate but could be one

**Location**: `internal/inscription/marker.go` lines 33-36

Two regex patterns for code blocks where a single pattern with a flag would suffice. Minor, no functional impact.


## Missing Capabilities

### MC-1: No token counting or context budget estimation

The pipeline generates CLAUDE.md and all files that end up in context, but has no awareness of the token budget. There is no mechanism to:
- Estimate total token cost of the generated .claude/ directory
- Warn when CLAUDE.md exceeds a threshold
- Report per-section token cost for optimization decisions
- Validate that agent prompts + CLAUDE.md + skills fit within context constraints

For a context-engineering framework, this is the highest-priority missing capability.

### MC-2: No source file validation before projection

The pipeline does not validate source files before projecting them. Specifically:
- **Agent files**: No validation that they are valid markdown with expected frontmatter
- **Mena files**: `MenaFrontmatter.Validate()` only checks name and description; no validation of `allowed-tools`, `model`, or `triggers` values against known values
- **Template files**: No syntax validation of `.md.tpl` files before rendering (template errors surface at render time)
- **hooks.yaml**: Only validates schema_version="2.0", not individual hook entries

A `ari sync --validate` or `ari lint` command that pre-validates all sources would catch errors before they propagate to .claude/.

### MC-3: No dependency graph or circular dependency detection

Rite manifests declare `dependencies: [shared, ...]` but there is no validation that:
- Named dependencies exist
- The dependency graph is acyclic
- Dependency resolution is deterministic (the current priority ordering handles this implicitly, but there is no explicit cycle detection)

### MC-4: No diff/changelog output for sync operations

`ari sync` returns a structured result but does not produce a human-readable changelog of what changed. The inscription `Pipeline` has `GetDiff()` but it is not wired to the main sync path. Users running `ari sync` cannot easily see what was updated, added, or removed without comparing git diffs.

### MC-5: No rollback for rite scope

The inscription package has `BackupManager` with `CreateBackup`, `RestoreBackup`, and retention policies. But the main materialization pipeline does not create backups for the full .claude/ state (only CLAUDE.md legacy backup). If `ari sync` produces incorrect output, there is no `ari sync --rollback` for agents, commands, skills, or rules.

### MC-6: No parallel execution of independent stages

Stages 6 (agents), 7 (mena), and 8 (rules) are independent of each other but execute sequentially. The provenance collector is thread-safe (uses sync.Mutex), so parallel execution of independent stages would reduce sync time. Not critical now but would matter as the number of files grows.

### MC-7: No resource filtering for rite scope

User scope supports `--resource=agents|mena|hooks` filtering, but rite scope always syncs everything. Adding `--resource` support to rite scope would enable targeted re-syncs (e.g., only regenerate agents after editing agent files).


## Recommendations

### Priority 1: Architectural

1. **Unify provenance recording**: Record provenance for all files regardless of whether they were rewritten, not just when `writeIfChanged()` returns true. This eliminates the carry-forward dependency in `saveProvenanceManifest` Step 0.

2. **Thread collector into ProjectMena()**: Instead of post-hoc directory scanning for mena provenance, pass the collector into `ProjectMena()` so provenance is recorded at exact write time with precise source information.

3. **Consolidate inscription Pipeline.Sync() with materializeCLAUDEmd()**: Either make `Pipeline.Sync()` the single path (called by `materializeCLAUDEmd`) or deprecate it. Two parallel code paths for the same operation is a maintenance hazard.

### Priority 2: Safety

4. **Remove pre-deletion of managed agents**: The delete-before-rewrite pattern in `materializeAgents()` creates a crash-unsafe window. Let `writeIfChanged()` handle overwrites atomically.

5. **Use writeIfChanged in user scope**: Replace `copyUserFile()` with `writeIfChanged()` to avoid unnecessary file watcher triggers.

6. **Surface errors from mena provenance scanning**: Replace `continue` with at least a logged warning when `checksum.Dir()` fails.

### Priority 3: Missing Capabilities

7. **Add token counting**: Implement a `ari sync --budget` or `ari context budget` command that reports estimated token usage for all generated context. This is core to the framework's value proposition.

8. **Add source validation**: Implement `ari lint` that validates all source files (agents, mena, templates, hooks) before projection.

9. **Add human-readable sync changelog**: Wire the diff output from the inscription system to the main sync result, and print a summary of changes after each `ari sync`.

### Priority 4: Cleanup

10. **Deduplicate isValidSchemaVersion**: Move to a shared validation package.

11. **Replace hand-rolled itoa**: Use `strconv.Itoa` (already imported elsewhere in the package).

12. **Fix "Team context" comment**: Update to "Rite context" per SL-008 terminology standards.

13. **Evaluate Sprig dependency**: If only basic functions (join, lower, upper) are used in templates, consider replacing Sprig with a minimal custom funcmap.

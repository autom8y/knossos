---
domain: design-constraints
generated_at: "2026-03-08T21:08:37Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "dbf81b8"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "313c675e38c3e4000caa21dfac68c38f337b5fb95d53f85444ef8d43174f4171"
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Dual RiteManifest Types

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:59` and `/Users/tomtenuta/Code/knossos/internal/rite/manifest.go:42`

Two separate `RiteManifest` struct definitions coexist in the codebase. The `materialize.RiteManifest` is minimal (Name, Version, Agents, Dromena, Legomena, Skills, Hooks, HookDefaults, SkillPolicies). The `rite.RiteManifest` is richer and carries backward-compat planned-schema fields (SchemaVersion, Form, DisplayName, Workflow, Phases, Budget). Both are in active use; neither is a generated output of the other.

**Why it exists**: The `materialize` package predates the `rite` package's ambition to support both the "planned format" and the "actual format." Consolidation would require either making `rite.RiteManifest` the sole type (adding backward-compat to materialize) or generating `materialize.RiteManifest` from `rite.RiteManifest`, either of which is a coordinated multi-file refactor.

**Impact**: Adding a new manifest field requires updating both structs. Any consumer comparing manifest representations across the two packages will see different field sets. The `materialize.RiteManifest.Skills` field is already `Deprecated: use Legomena instead` (`/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:68`).

---

### TENSION-002: Sync() Wraps MaterializeWithOptions() — Two Pipeline Abstractions

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:540` (Sync) and `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:304` (MaterializeWithOptions)

`Sync()` is the unified entry point, but it delegates to `MaterializeWithOptions()` for the rite scope via an adapter method `syncRiteScope()` that manually maps `SyncOptions` back into the legacy `Options` struct. This creates two parallel option types for essentially the same operation. The result struct chain is `SyncResult -> RiteScopeResult`, where `RiteScopeResult` is assembled by mapping all fields out of `legacyResult` (visible at lines 659-669 with the `legacyOpts` / `legacyResult` variable names).

**Why it exists**: `MaterializeWithOptions` was the original API; `Sync` was added as the unified pipeline. Removing `MaterializeWithOptions` is blocked by test infrastructure and the `SyncRite` interface implementation.

**Impact**: Any new field on `Options` must also be threaded through `SyncOptions` and the mapping in `syncRiteScope()`. Forgetting one of the three touch points causes silent behavioral divergence.

---

### TENSION-003: CC File Watcher Atomic-Write Constraint

**Location**: Multiple files: `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil.go`, `/Users/tomtenuta/Code/knossos/internal/provenance/manifest.go:50-66`, `/Users/tomtenuta/Code/knossos/internal/materialize/materialize_agents.go:39`, `/Users/tomtenuta/Code/knossos/internal/materialize/materialize_rules.go:109`

Claude Code's file watcher crashes or disrupts active sessions when it observes DELETE events or rapid temp-file-then-rename patterns in `.claude/`. This creates a pervasive cross-cutting constraint on every write into `.claude/`: writes MUST use `WriteIfChanged()` (not `os.WriteFile()` or delete-then-recreate) to suppress unnecessary disk events. The provenance manifest has a structural equality check (`structurallyEqual()`) to avoid even valid atomic writes when only timestamps change.

**Why it exists**: External system behavior (CC file watcher) constrains internal write patterns. The constraint was identified from scar regressions (see `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go:24`).

**Impact**: Any new code that writes to `.claude/` must use `fileutil.WriteIfChanged()` or `fileutil.AtomicWriteFile()`. Direct `os.WriteFile()` or delete-before-recreate in `.claude/` is a regression risk. Provenance manifest must check `structurallyEqual()` before writing.

---

### TENSION-004: Triple Event Format Legacy Bridge (v1/v2/v3 JSONL)

**Location**: `/Users/tomtenuta/Code/knossos/internal/session/events_read.go:58-127`

Three event format versions coexist in `events.jsonl` files: v1 (pre-ADR-0027 `EventEmitter` format, detected by `"event"` field), v2 (Clew Contract flat, detected by `"type"` field), v3 (TypedEvent with `"data"` field, highest precedence). `ReadEvents()` implements format-sniffing logic to normalize all three. The canonical write path is v3 only, but the read path is a three-way parser.

**Why it exists**: ADR-0027 migrated the write path to v3. Existing archived session logs still contain v1/v2 events. The dual-struct header (`events_read.go:14-28`) explicitly documents the removal trigger: when all pre-ADR-0027 sessions are wrapped and archived.

**Impact**: Changes to `ReadEvents()` must preserve all three format detections. Removing v1 detection before archival is complete would break `ari session audit` for any session with legacy events. The `clewcontract.RenameV2Type()` rename map (`/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/type_rename.go`) is an append-only contract; entries must never be removed until v2 events are gone.

---

### TENSION-005: KNOSSOS_HOME Singleton Cache Poisoning Risk

**Location**: `/Users/tomtenuta/Code/knossos/internal/config/home.go:11-23`

`KnossosHome()` uses `sync.Once` for lazy initialization. This creates a package-level singleton that is poisoned on first call: any test that calls `KnossosHome()` (directly or transitively) before setting `KNOSSOS_HOME` bakes in the default value (`$HOME/Code/knossos`) for all subsequent tests in the same process. The comment at line 91-93 explicitly documents this as `RISK-003`.

**Why it exists**: Performance optimization to avoid repeated env-var lookups. The `sync.Once` pattern is correct for production but hostile to test isolation.

**Impact**: Tests requiring custom `KNOSSOS_HOME` must always call `config.ResetKnossosHome()` before and after. Any package that imports `config` and calls `KnossosHome()` in its init path may silently poison test runs. `ResolveRite()` in `SourceResolver` calls `config.KnossosHome()` at construction time — tests creating `SourceResolver` with a test home must reset first.

---

### TENSION-006: Mena Content Path Rewrite — Ordering Dependency in Regex Application

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/mena/content_rewrite.go:131-141`

The `applyRewrites()` function applies regex substitutions in a strictly ordered pipeline: (1) `INDEX.lego.md -> SKILL.md` first, then (2) general `{name}.lego.md -> {name}.md`, then (3) `{name}.dro.md -> {name}.md`. This ordering is load-bearing: running the general `.lego.md` pattern before the `INDEX.lego.md` pattern would incorrectly transform `INDEX.lego.md` to `INDEX.md` instead of `SKILL.md`.

**Why it exists**: The naming convention for INDEX files (`SKILL.md`) is a special case of the general rule. There is no way to encode this exception into a single regex without ordered passes.

**Impact**: The three passes in `applyRewrites()` must never be reordered. Adding new special-case transforms (e.g., a future `CATALOG.lego.md -> CATALOG.md` exception) requires prepending before the general pattern, not appending.

---

### TENSION-007: search/collectors.go Imports internal/cmd/explain — Downward Layer Violation

**Location**: `/Users/tomtenuta/Code/knossos/internal/search/collectors.go:12`

`internal/search` imports `internal/cmd/explain` to access the concept registry. This is a layering violation: `internal/search` is a core domain library that should not depend on CLI command packages. The correct direction is `internal/cmd/ask` importing `internal/search`, not the reverse.

**Why it exists**: The `explain` package holds concept definitions that the search index needs to enumerate. Refactoring would require extracting the concept registry out of `cmd/explain` into a shared package (e.g., `internal/concepts`).

**Impact**: Any change to `internal/cmd/explain`'s concept API breaks `internal/search`. Tests of `internal/search` indirectly depend on the CLI command package. This makes `internal/search` harder to test in isolation.

---

### TENSION-008: Dual Lock Staleness Interpretation

**Location**: `/Users/tomtenuta/Code/knossos/internal/lock/lock.go:170-198`, `/Users/tomtenuta/Code/knossos/internal/cmd/session/recover.go:170`

The same legacy PID-format lock file is interpreted differently by two code paths: `lock.IsStale(treatLegacyAsStale=false)` checks process liveness for legacy PID locks (conservative), while `recover.isLockStale()` always treats legacy PID locks as stale (aggressive). This intentional disagreement is documented in test comments at `/Users/tomtenuta/Code/knossos/internal/cmd/session/recover_test.go:122-155`.

**Why it exists**: Recovery mode (ari recover) needs to be aggressive to unblock stuck sessions; the normal lock path needs to be conservative to avoid stomping a live process. The divergent policies serve different safety requirements.

**Impact**: The two interpretations must remain separate. Unifying them to the aggressive interpretation would break the normal lock path's safety guarantee. Any new lock staleness check must choose a policy explicitly.

---

### TENSION-009: Org Scope Is a Third Tier Without Full Parity

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/org_scope.go`, `/Users/tomtenuta/Code/knossos/internal/materialize/source/resolver.go:127-139`

Org scope was added as tier 4 in the resolution chain (project > user > org > knossos > embedded). However, org scope only syncs agents and mena (via `orgscope.SyncOrgScope`) — it does not sync hooks, rules, or settings in the same way rite scope does. The `SyncResult` type carries an `OrgResult *OrgScopeResult` field, but `OrgScopeResult` is a simpler struct (only Status/Error/OrgName/Source/Agents/Mena counts).

**Why it exists**: Org scope was added incrementally (session-20260302 sprint sessions). Full parity with rite scope was deferred.

**Impact**: Adding org-level hooks or rules requires extending `orgscope.SyncOrgScope` and `OrgScopeResult` in coordination. Agents working with org scope should not assume hooks or settings are org-configurable.

---

### TENSION-010: .claude/ Is NOT a Pure Cache

**Location**: Documented in CLAUDE.md materialization invariants and golden rules.

`.claude/` contains both knossos-generated platform files AND user state (user-agents, user-hooks, satellite regions in CLAUDE.md). The idempotency invariant requires that running sync twice produces identical output AND never destroys user content. This means the pipeline cannot `rm -rf .claude/` and regenerate — it must use selective writes with provenance tracking to distinguish knossos-owned from user-owned content.

**Why it exists**: Claude Code uses `.claude/` as the project context directory. Users customize it. The platform generates into it. The two uses cannot be separated without breaking CC's discovery mechanism.

**Impact**: Any cleanup operation on `.claude/` must consult `PROVENANCE_MANIFEST.yaml` before deleting. The `cleanStaleMenaEntries()` function demonstrates the correct pattern (check `entry.Owner == OwnerKnossos` before removing). Never use `os.RemoveAll` on the agents/ or skills/ or commands/ directories.

---

## Trade-off Documentation

### Trade-off 1: WriteIfChanged Over Atomic Write for CC Stability

**Chosen**: Use `fileutil.WriteIfChanged()` (read-compare-then-conditional-atomic-write) everywhere in `.claude/`.

**Rejected**: Simple `os.WriteFile()` or `fileutil.AtomicWriteFile()` unconditionally.

**Why**: CC's file watcher observes DELETE events from temp-file-then-rename patterns. Even valid atomic writes that produce identical content still trigger file watcher events. `WriteIfChanged()` adds a read overhead but eliminates watcher churn. The provenance manifest adds a structural equality check on top of byte equality for the same reason.

**Cost**: Every sync reads existing files before potentially writing them. For `.claude/` directories with many agents, this adds disk reads proportional to agent count.

---

### Trade-off 2: Dual Read Path for Events (v1/v2/v3) Instead of Format Migration

**Chosen**: Format-sniffing reader that normalizes all three formats at read time.

**Rejected**: One-time migration script to upgrade all v1/v2 events to v3 in place.

**Why**: Sessions may be resumed across format versions. In-place migration of JSONL files risks data loss on crash during migration. The dual read path is safe to remove once sessions are archived; migration scripts require careful orchestration across all active/parked sessions.

**Cost**: `ReadEvents()` maintains three-way format detection code that must be kept in sync with each format's field schema. The removal trigger (all pre-ADR-0027 sessions archived) creates long-lived technical debt.

---

### Trade-off 3: Provenance Manifest Lives in .knossos/, Not .claude/

**Chosen**: `PROVENANCE_MANIFEST.yaml` in `.knossos/` (outside CC's project context directory).

**Rejected**: Storing provenance in `.claude/` alongside the files it tracks.

**Why**: `.knossos/` is gitignored infrastructure; `.claude/` is CC's project context. Writing provenance into `.claude/` would expose internal platform tracking to CC's context window and require all CC sessions to load it. `.knossos/` is invisible to CC.

**Cost**: Any code needing provenance must derive the `.knossos/` path from the project root. The `cleanStaleMenaEntries()` function has a fallback path derivation (`filepath.Dir(claudeDir), ".knossos/"`) for this reason.

---

### Trade-off 4: Aggressive Orphan Auto-Removal vs. User Safety

**Chosen**: `--remove-all` (default in non-Soft sync) auto-removes knossos-owned orphans after backing them up.

**Rejected**: Always prompt before removing; always keep orphans.

**Why**: Stale agents from old rites pollute the CC agent pool and confuse routing. Auto-removal with backup provides a recoverable cleanup path. `--keep-orphans` flag exists for users who want manual control.

**Cost**: A user who customizes a knossos-managed agent without changing its provenance entry will lose that customization on the next sync. The divergence detection system (`DetectDivergence`) is the mitigation, but it only promotes to user ownership if the checksum changed.

---

### Trade-off 5: Inlining XDG Config Path in config.ActiveOrg() to Avoid Circular Import

**Chosen**: Duplicate the XDG config path logic inside `config.ActiveOrg()` instead of importing `internal/paths`.

**Rejected**: Import `internal/paths` from `internal/config`.

**Why**: `internal/paths` imports `internal/config` (for `config.KnossosHome()`). Importing `internal/paths` from `internal/config` would create a circular dependency. The comment at `/Users/tomtenuta/Code/knossos/internal/config/home.go:60` explicitly documents this.

**Cost**: XDG config path logic is duplicated between `config.ActiveOrg()` and `paths.ConfigDir()`. Changes to macOS detection logic must be made in both places.

---

### Trade-off 6: Two-Manifest Architecture (Rite + User Provenance)

**Chosen**: Separate `PROVENANCE_MANIFEST.yaml` (project `.knossos/`) and `USER_PROVENANCE_MANIFEST.yaml` (`~/.claude/`) for rite and user scope respectively.

**Rejected**: Single unified manifest covering both scopes.

**Why**: Rite scope content is per-project; user scope content is global across all projects. Merging them would require project-keyed namespacing in the manifest and would make the manifest grow proportionally to number of projects.

**Cost**: Any code reconciling rite and user scope must load two manifests. The `CollisionChecker` reads the rite manifest to detect shadowing conflicts at user-scope sync time.

---

## Abstraction Gap Mapping

### Missing Abstraction: Concept Registry

**Location**: `/Users/tomtenuta/Code/knossos/internal/cmd/explain/concepts.go`

`CollectConcepts()` in `internal/search/collectors.go` imports `internal/cmd/explain` to enumerate the concept registry. Concepts are a domain object that belong in an `internal/concepts` (or similar) package. The current location inside a CLI command package couples the search index build to the CLI layer.

**Evidence**: The import in `/Users/tomtenuta/Code/knossos/internal/search/collectors.go:12`.

---

### Missing Abstraction: Mena Source Unification Across Scopes

**Location**: `internal/materialize/mena/` and `internal/materialize/userscope/sync_mena.go`

The rite-scope and user-scope mena pipelines both perform: source collection, namespace collision detection, extension stripping, and destination write. The implementations are separate (`mena.SyncMena` vs `userscope` mena sync). There is no shared `MenaEngine` that both call; each scope re-implements the walk.

**Evidence**: `/Users/tomtenuta/Code/knossos/internal/materialize/userscope/sync_mena.go:169` comment: "CollectMena supports fs.FS sources via MenaSource.Fsys field" — the shared `CollectMena` function is the beginning of an abstraction, but the full sync loop is still duplicated.

---

### Missing Abstraction: Rite-Scope vs. User-Scope Template Availability

**Location**: `SourceResolver.checkSource()` at `/Users/tomtenuta/Code/knossos/internal/materialize/source/resolver.go:267-291`

Template directory resolution has a 5-way switch (Knossos/Project/Explicit/User/Org) with the comment "User rites don't have templates" and "Org rites don't carry templates." This scattered switch should be encapsulated in a `TemplatePolicy` type rather than embedded in resolution logic.

---

### Premature Abstraction: Sync() Over MaterializeWithOptions()

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:540`

The `Sync()` function with `SyncScope` enum (all/rite/org/user) is a correct abstraction for multi-scope dispatch. However, it wraps `MaterializeWithOptions()` via a mapping layer rather than replacing it. The `MaterializeWithOptions()` function's `Options` struct and `Result` struct should have been deprecated when `SyncOptions`/`SyncResult` were introduced. They remain as a "legacy" wrapper surface (visible from the variable naming `legacyOpts`, `legacyResult` in `syncRiteScope()`).

---

### Premature Abstraction: rite.RiteManifest "Legacy Format" Fields

**Location**: `/Users/tomtenuta/Code/knossos/internal/rite/manifest.go:66-78`

`rite.RiteManifest` carries planned-schema fields (SchemaVersion, DisplayName, Form, Workflow) that are only validated for the "legacy format." The dual-path `Validate()` at line 246 checks `if m.SchemaVersion != "" || m.Form != ""` to activate legacy validation. In practice, these fields are never set in actual rite manifests — they exist only for backward-compat with a schema design that was not adopted.

---

## Load-Bearing Code Identification

### LB-001: fileutil.WriteIfChanged() — CC Stability Invariant

**File**: `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil.go:66-72`

Every write to `.claude/` passes through this function. Replacing it with `os.WriteFile()` would trigger CC file watcher crashes for all users running active sessions during sync. This function is called from 30 production sites (confirmed via grep). Changing its semantics (e.g., removing the equality check) would break the CC stability invariant.

**Do not change**: The read-compare-skip logic. The atomic write fallback.

---

### LB-002: provenance.structurallyEqual() — Timestamp-Only Write Suppression

**File**: `/Users/tomtenuta/Code/knossos/internal/provenance/manifest.go:86-110`

`Save()` calls `structurallyEqual()` to avoid writing the provenance manifest when only `LastSync` / `LastSynced` timestamps changed. Without this check, every sync would write the manifest (triggering CC file watcher) even when no managed files changed.

**Do not change**: The field exclusion logic (LastSync/LastSynced are excluded from comparison). Removing any structural field from comparison would cause silent over-skipping.

---

### LB-003: applyRewrites() Ordering in content_rewrite.go

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/mena/content_rewrite.go:131-141`

See TENSION-006. The three-pass ordering is load-bearing. The `INDEX.lego.md -> SKILL.md` pass must precede the general `{name}.lego.md -> {name}.md` pass.

**Do not change**: The order of `reLinkIndexLego`/`reBacktickIndexLego` application relative to the general patterns.

---

### LB-004: ReadEvents() Format Detection Order

**File**: `/Users/tomtenuta/Code/knossos/internal/session/events_read.go:82-117`

Detection order (v3 first, then v1, then v2) is load-bearing. A v3 event contains both `"data"` and `"type"` fields; if v2 detection ran first, v3 events would be mis-classified. The SESSION-1 spec Section 5.1 defines this priority.

**Do not change**: The v3 `"data"` field detection must always run first.

---

### LB-005: RenameV2Type() — Append-Only Event Rename Map

**File**: `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/type_rename.go:14-22`

The v2-to-v3 rename map must remain append-only while any v2 events exist in archived sessions. Removing a rename entry (e.g., removing `"tool.call" -> "tool.invoked"`) would cause v2 events of that type to be reported under the old name instead of the v3 canonical name.

**Do not change**: Existing entries in `v2TypeRenames`. Safe to add; never remove until v2 format is fully retired.

---

### LB-006: config.KnossosHome() — sync.Once Singleton

**File**: `/Users/tomtenuta/Code/knossos/internal/config/home.go:11-23`

The `sync.Once` singleton is the source of `RISK-003` (test cache poisoning). However, the singleton itself is load-bearing: `KnossosHome()` is called from `SourceResolver` construction, which happens at the start of every sync command. Making it non-singleton would add per-call env-var overhead.

**Do not change**: The `sync.Once` caching. Instead, always use `config.ResetKnossosHome()` in tests.

---

### LB-007: materializeAgents() — NO Pre-Delete Before Overwrite

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize_agents.go:37-40`

The explicit comment (lines 37-40) forbids pre-deletion of managed agents before rewriting. Pre-deletion causes CC's file watcher to see DELETE events for files that are immediately recreated, crashing active sessions. `fileutil.WriteIfChanged()` handles the overwrite atomically without interim DELETE events.

**Do not change**: The no-pre-delete pattern. Any refactor that adds a cleanup pass before writing agents must be reviewed against the CC file watcher constraint.

---

### LB-008: Satellite Region Preservation in inscription.Merger

**Location**: `/Users/tomtenuta/Code/knossos/internal/inscription/pipeline.go:286-292` (DryRun path showing OwnerSatellite check)

Regions with `Owner == OwnerSatellite` in `KNOSSOS_MANIFEST.yaml` are never overwritten by the inscription pipeline. This is the mechanism protecting user content in CLAUDE.md. Any refactor of the merger must preserve this invariant.

**Do not change**: The `OwnerSatellite` bypass in the merge/generate pipeline.

---

## Evolution Constraint Documentation

### EC-001: Safe to Extend

- **New rite manifest fields**: Add to both `materialize.RiteManifest` and `rite.RiteManifest` with `omitempty`. Backward-compatible since YAML ignores unknown fields.
- **New SyncScope values**: Add new constant, add case to `Sync()` dispatch, add case to `IsValid()`. No existing callers break.
- **New event types (v3)**: Add to `clewcontract/typed_data.go` with new Data struct. Append `type_rename.go` only if renaming an existing type.
- **New mena extensions beyond .dro.md/.lego.md**: Would require extending `RouteMenaFile()`, `StripMenaExtension()`, and `applyRewrites()` in coordinated passes.
- **New provenance entry owner types**: Add to `OwnerType` constants and update `IsValid()`. Existing manifests remain loadable.

### EC-002: Coordinated Multi-File Effort Required

- **Replacing MaterializeWithOptions() with Sync()**: Must update `SyncRite()` interface in `internal/rite/syncer.go`, all callers in tests, and all cmd-level sync commands. Medium effort.
- **Merging the two RiteManifest structs**: Requires choosing canonical type, updating all import sites in both `materialize` and `rite` packages, and verifying field parity across tests.
- **Moving concept registry out of cmd/explain**: Must create new package, update `search/collectors.go` import, update `cmd/explain` to import the new package. Low risk, medium search scope.
- **Retiring v1/v2 event format reading**: Must verify all pre-ADR-0027 sessions are archived, then remove `Event` struct, `ClewEvent` struct, and the format-sniffing branches from `ReadEvents()`.

### EC-003: Migration Path Required

- **Changing provenance manifest schema version**: Must update `migrateV1ToV2()` pattern — add `migrateV2ToV3()`, update `Load()` to chain migrations.
- **Changing XDG directory layout**: Must update both `config.ActiveOrg()` (inlined XDG logic) and `paths.ConfigDir()`. If macOS convention changes, also update `XDGDataDir()`.
- **Changing .knossos/ directory name**: Load-bearing path used across `paths.Resolver`, provenance, and all ari commands. Requires a migration command to move existing `.knossos/` directories.

### EC-004: Frozen Areas

- **fileutil.WriteIfChanged() semantics**: Must not change. CC stability depends on it.
- **v2TypeRenames entries**: Append-only until v2 sessions fully retired.
- **OwnerSatellite behavior in inscription merger**: Must always skip overwrite.
- **Detection order in ReadEvents()**: v3 (`"data"` field) must always be checked first.
- **applyRewrites() pass order**: INDEX-specific patterns before general patterns.

---

## Risk Zone Mapping

### RZ-001: materialize/ — No Rollback on Partial Write Failure

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:370-518`

The 10-step materialization pipeline (agents, mena, rules, CLAUDE.md, settings, state, workflow, ACTIVE_RITE, provenance) is not transactional. If step 7 (CLAUDE.md generation) fails after steps 4-6 have already written agents and mena to disk, the project is left in partial state. The `prevalidateCLAUDEmd()` at step 0 mitigates this by failing fast on template rendering errors before any disk writes, but it cannot catch all failure modes (e.g., disk-full mid-write).

**Missing defense**: No rollback mechanism. If `materializeCLAUDEmd` fails after agents are written, the operator must re-run sync.

---

### RZ-002: config.KnossosHome() Test Cache Poisoning

**Location**: `/Users/tomtenuta/Code/knossos/internal/config/home.go:81-96`

Tests that import any package calling `KnossosHome()` before setting `KNOSSOS_HOME` will silently bake in the developer's actual home directory as the test fixture path. This can cause tests to pass locally (where `$HOME/Code/knossos` exists) and fail in CI (where it does not).

**Missing defense**: No `init()` test hook that forces reset. The test pattern (`config.ResetKnossosHome()` before `t.Setenv`) is documented but not enforced.

---

### RZ-003: search/collectors.go Circular Layer Risk

**Location**: `/Users/tomtenuta/Code/knossos/internal/search/collectors.go:12`

`internal/search` imports `internal/cmd/explain`. If `internal/cmd/explain` ever imports `internal/search` (e.g., to expose search-powered completions), a circular import would result. The Go compiler would surface this immediately, but the architectural coupling already exists in one direction.

**Missing defense**: No explicit dependency rule or linter guard preventing `internal/cmd` packages from being imported by core libraries.

---

### RZ-004: Mena Namespace Collision Across Rites

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/mena/namespace.go` (implied by Warnings field on MenaProjectionResult)

When two rites (or a rite and shared mena) define mena with the same flat name, the higher-priority source wins silently (with a Warnings append). The warning is collected but its surfacing to the operator depends on whether the caller logs it. A collision between a user-owned mena entry and a rite's mena entry could silently shadow the user's version.

**Missing defense**: No hard-fail option for collision detection. Warnings may be lost if callers do not log them.

---

### RZ-005: Lock Stale Threshold Is Time-Based, Not Process-Based (Non-Atomic)

**Location**: `/Users/tomtenuta/Code/knossos/internal/lock/lock.go:32` (`StaleThreshold = 5 * time.Minute`)

The primary stale detection for v2 JSON locks uses wall clock comparison (`time.Since(meta.Acquired) > StaleThreshold`). If a legitimate long-running ari command (e.g., a large sync) takes longer than 5 minutes, its lock may be incorrectly classified as stale by a concurrent `ari recover`. System clock skew or NTP jumps could also trigger false stale detection.

**Missing defense**: No heartbeat mechanism to refresh the lock's acquired timestamp during long operations.

---

### RZ-006: Org Scope Sync Silently Skips on Error in scope=all

**Location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:578-589`

When `opts.Scope == ScopeAll`, an org-scope sync error logs the error into `OrgScopeResult.Error` but does not abort the overall sync. This means a misconfigured org directory silently produces a partial sync where org agents/mena are missing, but the command returns success overall.

**Missing defense**: No user-visible warning in the `ari sync` output distinguishing "org sync skipped due to error" from "no org configured."

---

## Knowledge Gaps

- **Worktree rite inheritance**: The exact behavior of `inheritRiteFromMainWorktree()` and its failure modes under concurrent worktree operations was not deeply read. The worktree rite-switch cleanup (`cleanupThroughlineIDs()`) was not traced.
- **inscription/sync.go SyncCLAUDEmd**: The full merge pipeline (marker parsing, conflict detection, atomic write) was inferred from `pipeline.go` and the rules file comment; `sync.go` itself was not read in full.
- **sails/generator.go and naxos/scanner.go**: Both import `internal/session` but their structural role was not explored.
- **tribute/extractor.go**: Imports `internal/session`; its constraints were not examined.
- **Hook lifecycle (PreToolUse/PostToolUse/SubagentStart/SubagentStop)**: The `context.go`, `clew.go`, and `writeguard.go` in `internal/cmd/hook/` contain the hook execution logic, but only the constraint comments were read, not the full implementation.
- **Lock upgrade path**: Whether exclusive lock upgrade from shared is possible without TOCTOU risk was not verified.
- **Org scope full inventory**: What org scope syncs vs. does not sync was inferred but not verified by reading `orgscope/sync.go` in full.


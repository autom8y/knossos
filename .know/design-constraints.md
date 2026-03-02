---
domain: design-constraints
generated_at: "2026-03-01T16:08:41Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "89b109c"
confidence: 0.78
format_version: "1.0"
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Dual `OwnerType` Definitions (Naming Mismatch / Type Collision)

**Type:** Naming mismatch, layering violation
**Location:**
- `internal/inscription/types.go` lines 14-42: `inscription.OwnerType` (values: `knossos`, `satellite`, `regenerate`)
- `internal/provenance/provenance.go` lines 69-100: `provenance.OwnerType` (values: `knossos`, `user`, `untracked`)

**Evidence:** Both types share the type name `OwnerType`, both are `string` types, both have a `knossos` constant, but they are otherwise incompatible. The inscription system owns CLAUDE.md *regions*; the provenance system owns *files*.

**Historical reason:** Inscription predates the unified provenance model (ADR-0026). Each system independently designed its ownership enum.

**Ideal resolution:** Extract a shared `ownership` package with a single `OwnerType`, or rename one to avoid the collision (e.g., `inscription.RegionOwner`).

**Resolution cost:** Medium. Requires renaming across two packages and all their consumers in `materialize/` pipeline.

---

### TENSION-002: `materialize.RiteManifest` Dual Mena Fields (Backward Compat Naming Debt)

**Type:** Naming mismatch, accumulated compat shims
**Location:** `internal/materialize/materialize.go` lines 69-72

Four fields do overlapping work: `Dromena`, `Legomena`, `Commands` (backward compat), `Skills` (deprecated). The canonical names (`dromena`/`legomena`) are correct per ADR-0023 but the legacy names remain for backward compatibility.

**Historical reason:** Terminology evolved: `commands`->`dromena`, `skills`->`legomena`. Backward compatibility required keeping old fields parseable.

**Ideal resolution:** Remove `Skills` and `Commands` fields once all manifests are migrated.

**Resolution cost:** Low code change, high operational risk (must audit all satellite manifests before removal).

---

### TENSION-003: `ARIADNE_*` Environment Variables vs. Knossos Branding (RESOLVED)

**Type:** Naming mismatch (brand vs. implementation)
**Status:** RESOLVED. All `ARIADNE_*` env vars renamed to `ARI_*` in legacy cleanup sprint. `LegacyDataDir()` deleted (zero callers).

**Historical reason:** Pre-rename era. The product was originally "Ariadne" before the Knossos/ari rebrand.

**Resolution:** Renamed all env vars to `ARI_*` (e.g., `ARI_BUDGET_DISABLE`, `ARI_MSG_WARN`, `ARI_MSG_PARK`, `ARI_SESSION_KEY`, `ARI_STALE_SESSION_DAYS`). No deprecation shim needed -- zero external consumers confirmed by audit.

**Resolution cost:** Complete -- renamed in legacy cleanup sprint.

---

### TENSION-004: `internal/materialize/materialize.go` Monolith (PARTIALLY RESOLVED)

**Type:** Under-extracted; single file doing too much
**Location:** `internal/materialize/materialize.go` -- 732 lines (reduced from 1,562 via stage file extraction)

**Status:** PARTIALLY RESOLVED. Five stage files have been extracted (`materialize_agents.go`, `materialize_rules.go`, `materialize_settings.go`, etc.), reducing the main orchestration file from 1,562 to 732 lines. Three sub-packages also extracted (`mena/`, `userscope/`, `source/`).

**Remaining:** The 732-line core contains orchestration, types, constructors, and state tracking. Further extraction is possible but the file is now within reasonable bounds.

**Historical reason:** The pipeline grew incrementally. Stage extraction substantially reduced the monolith.

**Resolution cost:** Low remaining. The orchestration core is semantically cohesive at 732 lines.

---

### TENSION-005: Private `writeIfChanged` Wrapper (RESOLVED, PKG-013)

**Status:** RESOLVED. The wrapper was removed in PKG-013 cleanup sprint. All call sites in `materialize.go`, `materialize_agents.go`, `materialize_settings.go`, and `materialize_rules.go` now call `fileutil.WriteIfChanged` directly.

---

### TENSION-006: Two Parallel Load Paths for Shared Manifest (RESOLVED)

**Type:** Duplicated logic
**Status:** RESOLVED. The two parallel functions (`loadSharedHookDefaults`, `loadSharedSkillPolicies`) have been unified into a single `loadSharedManifest()` method on `Materializer`. Both `agent_transform.go` and `skill_policies.go` now delegate to this shared function.

**Resolution:** Single `loadSharedManifest() (*RiteManifest, error)` method extracts the manifest once; callers access the field they need.

---

### TENSION-007: `SourceType` String Constants Duplicated in Both `provenance.go` and `source/types.go`

**Type:** Naming mismatch / dual-definition
**Locations:**
- `internal/materialize/source/types.go` lines 9-20: `SourceProject`, `SourceUser`, `SourceKnossos`, etc.
- `internal/provenance/provenance.go` lines 56-58: `SourceType` values as plain strings

The provenance package uses these as plain strings without importing `source.SourceType`. Types are invisible to the compiler. Alignment is maintained by convention only.

**Historical reason:** ADR-0026 unified provenance but deliberately kept provenance as a leaf package (no internal imports).

**Resolution cost:** Medium if creating a shared package; very low if accepting documented string convention.

---

### TENSION-008: Dual Event Schema in `events.jsonl` (Historical Dual-System)

**Type:** Dual-system pattern (legacy coexistence)
**Location:** Documented in `docs/decisions/ADR-0027-unified-event-system.md`
- System A (`session.Event`): SCREAMING_CASE, `timestamp`/`event` fields
- System B (`clewcontract.Event`): snake_case `category.action`, `ts`/`type` fields

ADR-0027 documents the full tension and the decision to converge on `clewcontract.Event`. The migration is in progress.

**Resolution cost:** High. Requires audit of all `session.ReadEvents()` callers and output format consumers.

---

## Trade-off Documentation

### TENSION-001 (Dual OwnerType): Status Quo Rationale
**Current**: Two incompatible `OwnerType` types in separate packages.
**Ideal**: Single shared type.
**Why persists**: Inscription values (knossos/satellite/regenerate) serve CLAUDE.md region semantics; provenance values (knossos/user/untracked) serve file-level ownership. Their value sets are genuinely different. Merging them would create an enum with 5 values, only subsets of which are valid in each context.
**ADR link**: ADR-0026 implicitly accepts this tension.

### TENSION-002 (Manifest Field Debt): Status Quo Rationale
**Current**: Four overlapping fields in `RiteManifest`.
**Why persists**: Backward compatibility with existing satellite `manifest.yaml` files.
**ADR link**: ADR-0023 (dromena/legomena rename).

### TENSION-004 (Monolith): Status Quo Rationale
**Current**: 732-line `materialize.go` (reduced from 1,562 via stage file extraction).
**Why partially persists**: The remaining 732-line core is the orchestration layer -- now within reasonable bounds and semantically cohesive.
**ADR link**: None explicit. Acknowledged in `.claude/rules/internal-materialize.md`.

### TENSION-008 (Dual Event Schema): Status Quo Rationale
**Current**: Two schemas coexist in `events.jsonl`.
**Why persists**: The migration (ADR-0027) was accepted but not fully completed.
**ADR link**: ADR-0027 explicitly documents this and the decision to converge.

---

## Abstraction Gap Mapping

### Missing Abstraction: Shared Manifest Loader (RESOLVED)
**Status**: RESOLVED. `loadSharedManifest()` was extracted as a unified `Materializer` method (see TENSION-006).

### Missing Abstraction: Source Walk Dispatcher
**Evidence**: `CollectMena` in `internal/materialize/mena/collect.go` has parallel code paths for embedded FS and filesystem sources. `copyDirWithStripping` and `copyDirFromFSWithStripping` are near-identical except for `fs.FS` vs `os.*` calls.
**Recommended abstraction**: A unified `MenaSourceWalker` that dispatches internally between `fs.FS` and filesystem.
**Maintenance burden**: Three copies of the embedded/filesystem dispatch.

### Premature Abstraction: `internal/registry/registry.go`
**Evidence**: The registry contains 10 keys and is consulted in only 2 packages. The `panic`-on-unknown-key design is appropriate for compile-time safety but adds risk if keys are removed.
**Assessment**: Borderline. Solves a real problem at current scale.

### Zombie Abstraction: `internal/sync/state.go` active_rite (RESOLVED)
**Evidence**: `state.json` `active_rite` field had zero runtime read consumers (PKG-008 consumer audit, 2026-03-02). All 18 runtime consumers read from `.claude/ACTIVE_RITE` file.
**Resolution**: `ActiveRite` field removed from `sync.State` struct; `state.ActiveRite = ...` write removed from `trackState()`. Schema bumped to 1.1. Existing state.json files with the old field parse cleanly (Go JSON ignores unknown fields).
**Remaining**: `state.json` `last_sync` is also unread by runtime code -- full state.json removal is DEBT-039 follow-up.

---

## Load-Bearing Code

### LOAD-001: `provenance.Save()` Structural Equality Guard

**Location**: `internal/provenance/manifest.go`

Before writing PROVENANCE_MANIFEST.yaml, `Save()` checks if only timestamps changed and skips the write if content is structurally equal. This prevents triggering CC's file watcher on no-op syncs.

**What depends on it**: Every call site in the materialize pipeline, user scope sync, and tests that assert idempotency.

**What a naive "fix" would break**: Removing the guard would cause CC's file watcher to trigger on every sync, leading to infinite re-sync loops.

**Hot path**: Yes. Called on every `ari sync` and every `ari rite switch`.

---

### LOAD-002: `fileutil.WriteIfChanged()` Atomic Write with Change Detection

**Location**: `internal/fileutil/fileutil.go` lines 63-72

Reads existing file, compares content, and only calls `AtomicWriteFile()` (temp+rename) if content differs.

**What depends on it**: All materialization writes go through this function.

**What a naive "fix" would break**: Switching to direct `os.WriteFile` would eliminate atomic guarantees AND trigger CC file watchers on every sync.

---

### LOAD-003: `inscription.MergeRegions()` Satellite Region Preservation

**Location**: `internal/inscription/merger.go`

During CLAUDE.md sync, merges generated knossos regions with existing content, preserving `satellite` and `regenerate` regions.

**What depends on it**: The entire user-content-preservation guarantee. The materialization invariant "User content NEVER destroyed" depends entirely on this function.

**Load-bearing status**: Documented in rules file. Core invariant of the platform.

---

### LOAD-004: `namespace.resolveNamespace()` Collision Detection Before Flat Name Assignment

**Location**: `internal/materialize/mena/namespace.go` lines 20-187

Before projecting dromena to `.claude/commands/`, reads provenance manifest to distinguish knossos-owned from user-owned entries. If a flat name would collide with a user-owned command directory, knossos yields.

**What a naive "fix" would break**: Removing the provenance-based ownership check would cause knossos to silently overwrite user commands.

---

### LOAD-005: `session.FSM.ValidateTransition()` as State Machine Enforcer

**Location**: `internal/session/fsm.go`

Enforces the 4-state session FSM: 4 states, 5 transitions. `Archived` is terminal.

**What a naive "fix" would break**: Adding a new state without updating all lifecycle command handlers, the writeguard hook, and the event system.

**Hot path**: Every session operation.

---

## Evolution Constraints

### Changeability Matrix

| Area | Rating | Notes |
|------|--------|-------|
| `internal/errors/` | **SAFE** | Leaf package; adding new codes is additive. Do not reorder `ExitCode` constants. |
| `internal/checksum/` | **SAFE** | Pure functions. `sha256:` prefix is a contract -- do not change. |
| `internal/registry/` | **SAFE** | Adding keys is additive; removing keys requires auditing callers. `Ref()` panics on missing key. |
| `internal/paths/` | **COORDINATED** | `FindProjectRoot()` and `Resolver` methods called throughout. |
| `internal/session/fsm.go` | **COORDINATED** | State machine changes require updating all lifecycle handlers + event types + writeguard. |
| `internal/inscription/` | **COORDINATED** | `DefaultSectionOrder()` + `DeprecatedRegions()` govern CLAUDE.md structure; changes affect all satellites. |
| `internal/provenance/` | **MIGRATION** | Schema version `2.0` with `migrateV1ToV2()` shim; version bump requires manifest migration in all satellites. |
| `internal/materialize/materialize.go` | **COORDINATED** | 732 lines (reduced from 1,562); pipeline stage reordering requires checking all `opts.Soft` guards. |
| `internal/materialize/mena/` | **COORDINATED** | `INDEX.md->SKILL.md` rename logic in 3 places; must stay consistent. |
| `internal/hook/env.go` constants | **FROZEN** | `EnvHookEvent` etc. are `Deprecated` but kept for backward compat. |
| `ARI_*` env vars | **SAFE** | Renamed from `ARIADNE_*` in legacy cleanup sprint. Adding new `ARI_*` vars is additive. |

### Deprecated Markers

| Item | Location | Status |
|------|----------|--------|
| `RiteManifest.Skills` field | `internal/materialize/materialize.go:72` | `// Deprecated: use Legomena instead` |
| `hook.Env` env var constants | `internal/hook/env.go:13` | `// Deprecated: CC sends these via stdin JSON` |
| `paths.LegacyDataDir()` | Deleted | Was deprecated; removed (zero callers) |
| `session.EventEmitter` (System A) | Documented in ADR-0027 | Scheduled for removal |
| `inscription.DeprecatedRegions()` | `internal/inscription/manifest.go:302` | Active list of regions to drop on next sync |

---

## Risk Zone Mapping

### RISK-001: Silent Log Fallback on Agent Transform Failure (RESOLVED)

**Location**: `internal/materialize/materialize.go`

**Status:** RESOLVED. `transformAgentContent()` now returns errors that are propagated to the caller. Agent transform failures are no longer silently swallowed -- they surface as errors in the sync result.

---

### RISK-002: `resolveNamespace()` Silent Yield on Collision -- No User Feedback

**Location**: `internal/materialize/mena/namespace.go` lines 177-180

When a mena flat-name collision occurs, knossos silently falls back to the source-path-based name. The user gets a command at the wrong path, with no warning in CLI output.

**Recommended guard**: Surface collision as a `SyncResult` warning field.

---

### RISK-003: `config.KnossosHome()` Once-Cached with No Invalidation

**Location**: `internal/config/home.go` lines 11-23

`KnossosHome()` is cached via `sync.Once`. `ResetKnossosHome()` exists for testing only. Tests that call `KnossosHome()` transitively before setting `KNOSSOS_HOME` will silently use the wrong directory.

**Recommended guard**: Ensure all tests that need custom `KNOSSOS_HOME` call `config.ResetKnossosHome()` in `t.Cleanup()`.

---

### RISK-004: `mena/engine.go` `os.Remove*` Calls Without Error Propagation (PARTIALLY RESOLVED)

**Location**: `internal/materialize/mena/engine.go`

**Status:** PARTIALLY RESOLVED. `CleanEmptyDirs()` now returns `[]error` and callers surface these as `result.Warnings`. The stale-entry removal block in `cleanStaleMenaEntries()` still discards `os.RemoveAll`/`os.Remove` errors (DEBT-143). The `removeStaleFiles()` function also discards `os.Remove` errors for individual stale files.

---

### RISK-005: `provenance.Load()` Swallowed on Warm Path (RESOLVED)

**Location**: `internal/materialize/materialize.go`

**Status:** RESOLVED. `provenance.LoadOrBootstrap()` now propagates corruption errors upward rather than silently falling back to an empty manifest. File-not-found remains acceptable (bootstrap), but parse errors are surfaced.

---

## Shell Scripts Without Go Test Coverage (DEBT-035)

### Current Shell Script Inventory

| Script | Location | Purpose | Coverage |
|--------|----------|---------|---------|
| `e2e-validate.sh` | `scripts/e2e-validate.sh` | Distribution validation (brew install + ari init + sync) | None (integration only) |
| `context-injection.sh` | `rites/ecosystem/context-injection.sh` | DELETED (PKG-000b) — was dead code with zero runtime callers | N/A |
| `validation.sh` | `rites/shared/mena/cross-rite-handoff/validation.sh` | Cross-rite handoff artifact validation | None |

### Port-to-Go Plan (Priority Order)

**PRIORITY 1 — `rites/ecosystem/context-injection.sh`** (REMOVED, PKG-000b)
- **Status**: DELETED. Script had zero runtime callers. The dead call chain (`session-context.sh` -> `rite-context-loader.sh` -> `context-injection.sh`) was fully replaced by Go `ari hook context` implementation.
- 37 documentation references remain; cleanup is a separate future task.

**PRIORITY 2 — `rites/shared/mena/cross-rite-handoff/validation.sh`** (Skill dependency, testable)
- **Port target**: `ari validate handoff [path]` subcommand (extends existing `ari handoff` package)
- **Risk if unported**: Silent failures (bash `|| true` patterns); no structured error output
- **Effort**: LOW-MEDIUM — validates artifact existence and frontmatter. Already partially covered by `internal/validation/` package
- **Prerequisite**: Read script to enumerate validation rules before porting
- **Test plan**: Table-driven tests in `internal/cmd/handoff/validate_test.go`

**PRIORITY 3 — `scripts/e2e-validate.sh`** (CI-only, not in hot path)
- **Port target**: `ari e2e-validate` or keep as bash with timeout guards (already improved in DEBT-037)
- **Risk if unported**: Timeout stalls in CI (DEBT-037 partially mitigated). Script is not on the critical execution path.
- **Effort**: MEDIUM-HIGH — orchestrates brew tap/install; requires subprocess management
- **Recommendation**: Add timeout handling (done in DEBT-037) and leave as bash. Full Go port deferred until `ari release validate` workflow is defined.

### Deferred (Out of Scope)
- **`rites/*/hooks/*.sh`**: Already migrated to `ari hook *` Go subcommands (ADR-0011). No remaining shell hooks in knossos.
- Shell scripts in `.claude/` (auto-generated, not source artifacts)

### ADR Reference
ADR-0011 documents the "hook binary" decision: all hooks must be Go binaries, not shell scripts. The `context-injection.sh` script predates this ADR and is the main remaining violation.

---

## Go Version Configuration (DEBT-038)

`go.mod` declares `go 1.23.0` and `toolchain go1.24.13`.

- `go` directive: minimum required Go version for building knossos. Any Go 1.23+ will work.
- `toolchain` directive: preferred toolchain; Go 1.21+ will auto-download this if needed.
- Running binary: go1.24.13 (confirmed 2026-03-02 via `go version`).

This is the intentional Go 1.21+ pattern. It is NOT a mismatch. The `go` line guarantees compatibility; `toolchain` pins the preferred build environment. The `.know/architecture.md` note "Go 1.23" refers to the minimum `go` directive, not the actual toolchain in use.

No action required. If the minimum compatibility floor needs raising, bump the `go` directive.

---

## Knowledge Gaps

1. **Active Rite Source-of-Truth (DEBT-039 -- RESOLVED, PKG-008 complete)**

   Two stores hold the active rite name after PKG-008 removal. The authoritative store and roles:

   | Store | Path | Written by | Role |
   |-------|------|-----------|------|
   | `ACTIVE_RITE` file | `.claude/ACTIVE_RITE` | `materialize.writeActiveRite()` (step 10 of pipeline) | **Primary authoritative store.** Read at the START of every `syncRiteScope()` to determine the current rite when `--rite` is not passed. Also read by `inheritRiteFromMainWorktree()` for git worktrees. Used by hooks (`ari hook context`) for rite-conditional logic. |
   | `PROVENANCE_MANIFEST.yaml` | `.claude/PROVENANCE_MANIFEST.yaml` | `provenance.MergeManifests()` (called from `saveProvenanceManifest()`) | Secondary. Carries `active_rite` for provenance audit trail only. Not consulted during rite resolution. |

   **Note**: `state.json` `active_rite` was removed (PKG-008, 2026-03-02). Consumer audit found zero runtime readers. All 18 active rite consumers read from `ACTIVE_RITE` file.

   **Authoritative store**: `.claude/ACTIVE_RITE` (the plain text file). This is the ONLY store consulted when `syncRiteScope()` must determine which rite to materialize.

   **Write order within one sync run** (from `MaterializeWithOptions`):
   1. `trackState()` writes `state.json` last_sync only (step 9, line ~460)
   2. `writeActiveRite()` writes `ACTIVE_RITE` file (step 10, line ~477-480)
   3. `saveProvenanceManifest()` writes `PROVENANCE_MANIFEST.yaml` (line ~483)

   **Divergence scenarios and recovery**:

   - *ACTIVE_RITE deleted, state.json intact*: Next `ari sync` with no `--rite` flag will fail with `"no ACTIVE_RITE found, specify --rite"` (for `--scope=rite`) or silently fall to minimal mode (for `--scope=all`). Recovery: run `ari sync --rite <name>` to re-materialize.
   - *state.json deleted, ACTIVE_RITE intact*: `syncRiteScope()` reads ACTIVE_RITE fine; `trackState()` re-initializes state.json from scratch. Transparent self-healing.
   - *PROVENANCE_MANIFEST.yaml has wrong active_rite*: Not a functional issue -- provenance manifest is never consulted for rite selection. Only affects audit trail.

   **Worktree inheritance**: In a linked git worktree with no local ACTIVE_RITE, `inheritRiteFromMainWorktree()` reads `.claude/ACTIVE_RITE` from the main worktree directory. It does NOT consult state.json or PROVENANCE_MANIFEST.yaml.

   **Remaining zombie risk**: state.json `last_sync` is also unread by runtime code (DEBT-039 follow-up). Full state.json elimination is a future cleanup.

2. **Shell hook residue** -- ADR-0011 references shell scripts and a "remaining ports" backlog. Shell scripts were not in scope for this observation.

3. **`internal/sails/` package** -- 13 files implementing quality gate logic were not deeply read.

4. **`internal/naxos/` package** -- 5 files (scanner, report, types). Design constraints not captured.

5. **`internal/tribute/` and `internal/artifact/` packages** -- Not read. Their interaction with session state and provenance is unknown.

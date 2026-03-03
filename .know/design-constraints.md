---
domain: design-constraints
generated_at: "2026-03-03T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "1599813"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Dual `OwnerType` Definitions (Naming Mismatch / Type Collision)

**Type:** Naming mismatch, layering violation
**Location:**
- `internal/inscription/types.go` lines 14-42: `inscription.OwnerType` (values: `knossos`, `satellite`, `regenerate`)
- `internal/provenance/provenance.go` lines 69-100: `provenance.OwnerType` (values: `knossos`, `user`, `untracked`)

**Evidence:** Both types share the name `OwnerType`, both are `string` types, both carry a `knossos` constant, but they are otherwise incompatible. Both files carry cross-reference comments pointing to the other.

**Historical reason:** Inscription predates the unified provenance model (ADR-0026). Each system independently designed its ownership enum.

**Ideal resolution:** Rename one (e.g., `inscription.RegionOwner`), or extract a shared namespace package.

**Resolution cost:** Medium. Value sets are genuinely different (5 combined values, each valid only in its own context).

---

### TENSION-002: `materialize.RiteManifest` Dual Mena Fields (Backward Compat Naming Debt)

**Type:** Naming mismatch, accumulated compat shims
**Location:** `internal/materialize/materialize.go` lines 66-72

Four fields do overlapping work:
- `Dromena []string` — canonical (per ADR-0023)
- `Legomena []string` — canonical (per ADR-0023)
- `Commands []string` — backward compat
- `Skills []string` — `// Deprecated: use Legomena instead`

**Historical reason:** Terminology evolved: `commands` → `dromena`, `skills` → `legomena`.

**Ideal resolution:** Remove `Skills` and `Commands` once all satellite manifests are migrated.

**Resolution cost:** Low code change, high operational risk (must audit all satellite manifests first).

---

### TENSION-003: Dual `RiteManifest` Types (Materialize vs. Rite Package)

**Type:** Duplicate type, naming collision
**Location:**
- `internal/materialize/materialize.go` line 60: `materialize.RiteManifest` (12 fields, pipeline-oriented)
- `internal/rite/manifest.go` line 46: `rite.RiteManifest` (17+ fields, discovery/runtime-oriented)

Two independent structs share the same name and partially overlapping fields but serve different purposes. Neither package imports the other. Schema drift is possible.

**Historical reason:** The two packages evolved independently. Importing across avoided to prevent circular dependencies.

**Ideal resolution:** Extract a shared `manifest` package or add cross-reference comments.

**Resolution cost:** High. Field sets diverge for good reasons.

---

### TENSION-004: `internal/materialize/materialize.go` Monolith (PARTIALLY RESOLVED)

**Type:** Under-extracted; single file doing too much
**Location:** `internal/materialize/materialize.go` — 733 lines (reduced from 1,562 via stage file extraction)

**Status:** PARTIALLY RESOLVED. Five stage files extracted (`materialize_agents.go`, `materialize_rules.go`, `materialize_settings.go`, `materialize_mena.go`, `materialize_claudemd.go`). Three sub-packages also extracted (`mena/`, `userscope/`, `source/`).

**Remaining:** 733-line core with orchestration, types, constructors, state tracking. Within reasonable bounds.

---

### TENSION-005: `SourceType` String Constants Duplicated in Provenance and Source Packages

**Type:** Naming mismatch / dual-definition
**Locations:**
- `internal/materialize/source/types.go` lines 9-20: typed constants
- `internal/provenance/provenance.go` lines 56-68: plain strings with sync-by-convention comment

ADR-0026 deliberately keeps provenance as a leaf package (no internal imports). Alignment maintained by convention only.

**Resolution cost:** Very low if accepting string convention with linting rule. Medium if creating shared constants package.

---

### TENSION-006: ADR-0027 Dual Event Schema Migration In Progress

**Type:** Dual-system pattern (legacy coexistence)
**Locations:**
- `internal/session/events_read.go` lines 13-27: Legacy `session.Event` struct (v1)
- `internal/hook/clewcontract/event.go`: Canonical `clewcontract.Event` (v2)
- `internal/hook/clewcontract/typed_event.go`: New `TypedEvent` envelope (v3)

Three event schema generations coexist in `events.jsonl`. `ReadEvents()` format-sniffs each line. v1 write path removed; read path bridges all formats.

**Resolution cost:** High for full unification. v2-to-v3 coexistence is deliberate.

---

### TENSION-007: `config.KnossosHome()` Once-Cached with No Invalidation in Production

**Type:** Under-engineered (test hazard)
**Location:** `internal/config/home.go` lines 11-23

`KnossosHome()` is cached via `sync.Once`. `ResetKnossosHome()` exists for testing only. Tests that call `KnossosHome()` transitively before setting `KNOSSOS_HOME` silently poison the cache.

**Resolution cost:** Low. `ResetKnossosHome()` already exists. Document the pattern.

---

### TENSION-008: `internal/config` Duplicates XDG Path Logic from `internal/paths`

**Type:** Duplicated logic
**Location:** `internal/config/home.go` lines 61-74 vs. `internal/paths/paths.go`

Both implement macOS `~/Library/Application Support` vs. Linux `~/.config` detection independently. Cannot import across due to circular import risk.

**Ideal resolution:** Extract `internal/platform` or `internal/xdgutil` leaf package.

**Resolution cost:** Low.

---

## Trade-off Documentation

### TENSION-001: Status Quo Rationale
Value sets are genuinely different: inscription has `knossos`/`satellite`/`regenerate` (region ownership); provenance has `knossos`/`user`/`untracked` (file ownership). Merging creates a leaky 5-value enum. Current approach keeps semantics clean at naming collision cost. **ADR link:** ADR-0026 implicitly accepts this.

### TENSION-002: Status Quo Rationale
Backward compatibility with existing satellite `manifest.yaml` files. Removing fields would silently break satellites still using old names. **ADR link:** ADR-0023 documents naming but not removal timeline.

### TENSION-003: Status Quo Rationale
Import graph deliberately isolates `internal/rite` (4 imports) from `internal/materialize` (16 imports). Mixing would violate layer boundaries. **ADR link:** ADR-0014 implicitly separates discovery from materialization.

### TENSION-005: Status Quo Rationale
ADR-0026 explicitly decided provenance is a leaf package. String convention with comment at `provenance.go:67` is the documented workaround.

### TENSION-006: Status Quo Rationale
ADR-0027 migration completed write path. Read path must bridge all historical log files. Removal trigger documented.

---

## Abstraction Gap Mapping

### Missing Abstraction: XDG Platform Detection (Duplicated)
**Location:** `internal/config/home.go:61-74` and `internal/paths/paths.go`
Both perform macOS-vs-Linux XDG directory detection independently. See TENSION-008.
**Recommended:** `internal/platform` or `internal/xdgutil` leaf package.

### Missing Abstraction: Shared `go/ast` Diff Logic
**Location:** `internal/know/astdiff.go`
General-purpose Go AST utility (declaration-level diffing) but package-private to `know`. Currently one consumer; potentially reusable.
**Maintenance burden:** Low at current scale.

### Missing Abstraction: Source Walk Dispatcher
**Location:** `internal/materialize/mena/`
`CollectMena` has parallel code paths for embedded FS and filesystem. Substantially improved after `copyDirFS` extraction. Remaining duplication bounded within package.

### Zombie Abstraction: `state.json` `last_sync` Field
**Location:** `internal/sync/state.go`
`active_rite` field was removed (PKG-008). `last_sync` field has zero confirmed runtime consumers. Full `state.json` elimination is DEBT-039.

### Premature Abstraction: `internal/registry/registry.go`
~10 keys, consulted in 2 packages. `panic`-on-unknown-key appropriate for compile-time safety. Borderline; not a significant burden.

---

## Load-Bearing Code Identification

### LOAD-001: `provenance.Save()` Structural Equality Guard
**Location:** `internal/provenance/manifest.go`
Before writing `PROVENANCE_MANIFEST.yaml`, checks structural equality and skips write if only timestamps changed. Prevents triggering CC file watcher on no-op syncs.
**What a naive fix would break:** Removing guard causes infinite re-sync loops in active CC sessions.
**Hot path:** Every `ari sync` and every `ari rite switch`.

### LOAD-002: `fileutil.WriteIfChanged()` Atomic Write with Change Detection
**Location:** `internal/fileutil/fileutil.go:63-72`
Reads existing file, compares bytes, only calls `AtomicWriteFile` (temp+rename) if content differs.
**What a naive fix would break:** Direct `os.WriteFile` eliminates atomicity AND triggers CC file watcher on every sync.

### LOAD-003: `inscription.MergeRegions()` Satellite Region Preservation
**Location:** `internal/inscription/merger.go`
Merges knossos regions with existing content, preserving `satellite` and `regenerate` regions. The platform invariant "User content NEVER destroyed" depends entirely on this function.
**Safe refactor:** Any change must be tested with existing satellite content across all known CLAUDE.md structures.

### LOAD-004: `namespace.resolveNamespace()` Provenance-Based Collision Detection
**Location:** `internal/materialize/mena/namespace.go:20-204`
Before projecting dromena to `.claude/commands/`, reads provenance to distinguish knossos-owned from user-owned entries. Knossos yields on collision.
**What a naive fix would break:** Removing provenance check causes knossos to silently overwrite user commands.

### LOAD-005: `session.FSM.ValidateTransition()` State Machine Enforcer
**Location:** `internal/session/fsm.go`
Enforces 4-state session FSM: `None → Active`, `Active → Parked`, `Active → Archived`, `Parked → Active`, `Parked → Archived`. `Archived` is terminal.
**What a naive fix would break:** Adding a state without updating all lifecycle handlers, writeguard, CC session map, and event constructors.

### LOAD-006: `rite.Syncer` Interface as Dependency Inversion Boundary
**Location:** `internal/rite/syncer.go`
Single-method interface breaking upward dependency from `internal/rite` → `internal/materialize`. The entire import graph separation depends on this.

---

## Evolution Constraint Documentation

### Changeability Matrix

| Area | Rating | Notes |
|------|--------|-------|
| `internal/errors/` | **SAFE** | Leaf package. Adding codes is additive. Do not reorder `ExitCode` constants. |
| `internal/checksum/` | **SAFE** | `sha256:` prefix is contract with provenance — do not change without migration. |
| `internal/fileutil/` | **SAFE** | `WriteIfChanged` signature change requires updating 20+ call sites. |
| `internal/registry/` | **SAFE** | Adding keys additive; removing keys requires auditing callers (panics on missing). |
| `internal/paths/` | **COORDINATED** | `FindProjectRoot()` and `Resolver` called throughout. Any signature change requires broad sweep. |
| `internal/session/fsm.go` | **COORDINATED** | State machine changes require updating all lifecycle handlers, events, writeguard simultaneously. |
| `internal/inscription/` | **COORDINATED** | `DefaultSectionOrder()` and `DeprecatedRegions()` govern CLAUDE.md for all satellites. |
| `internal/provenance/` | **MIGRATION** | Schema version `2.0` with migration shim. Bump requires manifest migration in all projects. |
| `internal/materialize/materialize.go` | **COORDINATED** | Pipeline stage reordering requires checking all `opts.Soft` guards. |
| `internal/hook/env.go` env var constants | **FROZEN** | Deprecated but kept for backward compat. Do not remove without confirming zero consumers. |
| `internal/config/home.go` | **COORDINATED** | `sync.Once` cache test-hazardous. Reset required in tests setting `KNOSSOS_HOME`. |

### Deprecated Markers

| Item | Location | Status |
|------|----------|--------|
| `materialize.RiteManifest.Skills` | `internal/materialize/materialize.go:69` | `// Deprecated: use Legomena` |
| `materialize.RiteManifest.Commands` | `internal/materialize/materialize.go:68` | Backward compat; no removal plan |
| `hook.Env` env var constants | `internal/hook/env.go:12` | `// Deprecated: CC sends via stdin JSON` |
| Legacy `session.Event` v1 struct | `internal/session/events_read.go:29` | Read-path only; removal after pre-ADR-0027 sessions archived |
| `inscription.DeprecatedRegions()` | `internal/inscription/manifest.go:303` | Active list of regions to drop on next sync |

### External Dependency Constraints

| Dependency | Constraint |
|------------|-----------|
| `CGO_ENABLED=0` | All packages must compile without CGO. Eliminated SQLite as option for CC session map. |
| `go 1.23.0` minimum | `go.mod` minimum; `toolchain go1.24.13` actual build tool. |
| CC hook stdin JSON protocol | CC delivers hook data via stdin JSON, not env vars. New fields must come from `StdinPayload`. |
| XDG directory conventions | Platform path diverges macOS vs. Linux. Both `config` and `paths` implement independently. |

---

## Risk Zone Mapping

### RISK-002: `materializeMena()` Discards Namespace Collision Warnings
**Location:** `internal/materialize/materialize_mena.go` line 114
`_, err := SyncMena(sources, opts)` discards `*MenaProjectionResult`. Namespace collision warnings lost at the `materialize/` boundary. Users get no feedback from `ari sync`.
**Status:** PARTIALLY RESOLVED within `mena/` package; discard at `materialize/` boundary remains.
**Recommended fix:** Change `_` to named result, propagate `menaResult.Warnings` into `RiteScopeResult`.

### RISK-003: `config.KnossosHome()` Once-Cached with No Production Invalidation
**Location:** `internal/config/home.go:11-23`
Tests that call `KnossosHome()` transitively before setting `KNOSSOS_HOME` silently poison the cache for all subsequent tests.
**Guard:** All tests requiring custom `KNOSSOS_HOME` must call `config.ResetKnossosHome()` in `t.Cleanup()`.

### RISK-004: `mena/engine.go` Stale Entry Removal Silently Discards Errors
**Location:** `internal/materialize/mena/engine.go`
`removeStaleFiles()` discards individual removal errors. Stale mena files accumulate silently.
**Status:** PARTIALLY RESOLVED. `CleanEmptyDirs` fixed; stale-entry block remains.

### RISK-006: `internal/mena/` vs `internal/materialize/mena/` Package Naming Ambiguity
**Location:** Two packages share the name `mena` at different path depths. Source-side scanning vs materialization-side pipeline.
**Risk:** Developer confusion (compiler catches wrong import immediately).
**Recommended guard:** Add package doc comment to `internal/mena/` distinguishing from materialization-side.

---

## Knowledge Gaps

1. **`internal/sails/` package** — 13 files implementing quality gate logic. Load-bearing status not captured.
2. **`internal/naxos/` package** — interaction with provenance/materialization not captured.
3. **`internal/tribute/` and `internal/artifact/`** — interaction with session state and provenance not captured.
4. **`internal/know/` incremental cycle design** — `incremental_cycle`/`max_incremental_cycles` fields and `astdiff.go` semantic diffing are recent additions not fully mapped.
5. **DEBT-039 (`state.json` `last_sync` elimination)** — whether `state.json` has remaining consumers beyond write path is unconfirmed.

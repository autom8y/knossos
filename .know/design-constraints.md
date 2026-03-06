---
domain: design-constraints
generated_at: "2026-03-06T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "3847e28"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "d2ffa6c6e09a8852bf5f5169eb68f03dd80a73cf4b057a217bbe9a734503e94b"
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Dual `OwnerType` Definitions (Naming Mismatch / Type Collision)

**Type:** Naming mismatch, layering violation
**Location:**
- `internal/inscription/types.go` lines 14-42: `inscription.OwnerType` (values: `knossos`, `satellite`, `regenerate`)
- `internal/provenance/provenance.go` lines 78-100: `provenance.OwnerType` (values: `knossos`, `user`, `untracked`)

**Evidence:** Both types share the name `OwnerType`, both are `string` types, both carry a `knossos` constant, but they are otherwise incompatible. Both files carry explicit cross-reference comments.

**Historical reason:** Inscription predates the unified provenance model (ADR-0026). Each system independently designed its ownership enum.

**Ideal resolution:** Rename one (e.g., `inscription.RegionOwner`), or extract a shared namespace package.

**Resolution cost:** Medium. Value sets are genuinely different (5 combined values, each valid only in its own context).

---

### TENSION-002: `materialize.RiteManifest` Dual Mena Fields (Backward Compat Naming Debt)

**Type:** Naming mismatch, accumulated compat shims
**Location:** `internal/materialize/materialize.go` lines 65-68

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
- `internal/materialize/materialize.go` line 59: `materialize.RiteManifest` (12 fields, pipeline-oriented)
- `internal/rite/manifest.go` line 42: `rite.RiteManifest` (17+ fields, discovery/runtime-oriented)

Two independent structs share the same name and partially overlapping fields but serve different purposes. Neither package imports the other. Schema drift is possible.

**Historical reason:** The two packages evolved independently. Import isolation avoids circular dependencies.

**Ideal resolution:** Extract a shared `manifest` package or add explicit cross-reference comments.

**Resolution cost:** High. Field sets diverge for good reasons (pipeline vs. runtime).

---

### TENSION-004: `internal/materialize/materialize.go` Monolith (PARTIALLY RESOLVED)

**Type:** Under-extracted; single file doing too much
**Location:** `internal/materialize/materialize.go` — now 733 lines (reduced from 1,562 via stage file extraction)

**Status:** PARTIALLY RESOLVED. Five stage files extracted: `materialize_agents.go`, `materialize_rules.go`, `materialize_settings.go`, `materialize_mena.go`, `materialize_claudemd.go`. Three sub-packages also extracted: `mena/`, `userscope/`, `source/`.

**Remaining:** 733-line core with orchestration, types, constructors, state tracking. Within reasonable bounds.

---

### TENSION-005: `SourceType` String Constants Duplicated in Provenance and Source Packages

**Type:** Naming mismatch / dual-definition
**Locations:**
- `internal/materialize/source/types.go` lines 9-22: typed `SourceType` constants
- `internal/provenance/provenance.go` lines 59-68: plain string constants with sync-by-convention comment

ADR-0026 deliberately keeps provenance as a leaf package (no internal imports). String alignment maintained by convention only.

**Resolution cost:** Very low if accepting string convention with linting rule. Medium if creating shared constants package.

---

### TENSION-006: ADR-0027 Dual Event Schema Migration In Progress

**Type:** Dual-system pattern (legacy coexistence)
**Locations:**
- `internal/session/events_read.go` lines 29-56: Legacy `session.Event` (v1) and `ClewEvent` (v2) structs
- `internal/hook/clewcontract/event.go` line 56: Canonical `clewcontract.Event` (v2)
- `internal/hook/clewcontract/typed_event.go` line 31: `TypedEvent` (v3) envelope

Three event schema generations coexist in `events.jsonl`. `ReadEvents()` format-sniffs each line. v1 write path removed; read path bridges all formats.

**Removal trigger documented** at `events_read.go:24-27`: once all pre-ADR-0027 sessions are archived, the legacy read bridge can be removed.

**Resolution cost:** High for full unification. v2-to-v3 coexistence is deliberate.

---

### TENSION-007: `config.KnossosHome()` Once-Cached with No Invalidation in Production

**Type:** Under-engineered (test hazard)
**Location:** `internal/config/home.go` lines 11-23

`KnossosHome()` is cached via `sync.Once`. `ResetKnossosHome()` exists for testing only. Tests that call `KnossosHome()` transitively before setting `KNOSSOS_HOME` silently poison the cache.

**Resolution cost:** Low. `ResetKnossosHome()` already exists.

---

### TENSION-008: `internal/config` Duplicates XDG Path Logic from `internal/paths`

**Type:** Duplicated logic
**Location:** `internal/config/home.go` lines 61-74 vs. `internal/paths/paths.go`

Both implement macOS `~/Library/Application Support` vs. Linux `~/.config` detection independently. Cannot import across due to circular import risk.

**Ideal resolution:** Extract `internal/platform` or `internal/xdgutil` leaf package.

**Resolution cost:** Low.

---

### TENSION-009: `tribute.EventData` as a Fourth Event-Parsing Struct

**Type:** Duplicated abstraction
**Location:** `internal/tribute/types.go` lines 222-273: `EventData` struct with dual timestamp/type fields

`tribute.EventData` is a fourth independent struct for parsing `events.jsonl`, alongside `session.Event` (v1), `session.ClewEvent` (v2), and `clewcontract.Event` (canonical).

**Historical reason:** `internal/tribute` avoids importing `internal/session` or `internal/hook/clewcontract` to prevent import cycles.

**Ideal resolution:** Import `clewcontract` directly (it is a leaf package with no internal imports).

**Resolution cost:** Medium. `clewcontract` could be directly imported by `tribute`, eliminating `EventData`.

---

### TENSION-010: `IsKnossosProject` Flag Embeds Ecosystem-Specific Content in Platform Generator

**Type:** Framework agnosticism violation (residual)
**Location:** `internal/inscription/generator.go` lines 44-46, 405, 512, 531, 545, 571, 586

The inscription generator has a special `IsKnossosProject` flag that controls whether "Pythia coordinates" language appears in CLAUDE.md content. Platform-layer code contains ecosystem-specific agent names.

**Resolution cost:** Low. Move "Pythia" references to a template variable populated from `ACTIVE_RITE+agents`.

---

## Trade-off Documentation

### TENSION-001: Status Quo Rationale
Value sets are genuinely different: inscription has `knossos`/`satellite`/`regenerate` (region ownership); provenance has `knossos`/`user`/`untracked` (file ownership). Merging creates a leaky 5-value enum. **ADR link:** ADR-0026 implicitly accepts this.

### TENSION-002: Status Quo Rationale
Backward compatibility with existing satellite `manifest.yaml` files. Removing fields would silently break satellites. **ADR link:** ADR-0023 documents naming but not removal timeline.

### TENSION-003: Status Quo Rationale
Import graph deliberately isolates `internal/rite` (minimal imports) from `internal/materialize` (heavy imports). **ADR link:** ADR-0014 implicitly separates discovery from materialization.

### TENSION-005: Status Quo Rationale
ADR-0026 explicitly decided provenance is a leaf package. String convention with comment at `provenance.go:67` is the documented workaround.

### TENSION-006: Status Quo Rationale
ADR-0027 migration completed write path. Read path must bridge all historical formats. Removal trigger documented.

### TENSION-008: Status Quo Rationale
`internal/config` is a leaf package. Importing `internal/paths` would create a cycle. The documentation comment at `config/home.go:61` keeps the two implementations synchronized by convention.

---

## Abstraction Gap Mapping

### Missing Abstraction: XDG Platform Detection (Duplicated)
**Location:** `internal/config/home.go` lines 61-74 and `internal/paths/paths.go`
Both perform macOS vs. Linux XDG directory detection independently. See TENSION-008.
**Recommended:** `internal/platform` or `internal/xdgutil` leaf package.

### Missing Abstraction: Shared `events.jsonl` Parser
**Locations:** `internal/session/events_read.go`, `internal/tribute/types.go` lines 222-273, `internal/cmd/handoff/status.go` line 153
Three packages implement independent `events.jsonl` parsers. `clewcontract` is a leaf package that could serve as a shared reader.

### Missing Abstraction: Shared `go/ast` Diff Logic
**Location:** `internal/know/astdiff.go`
General-purpose Go AST utility but package-private to `know`. Currently one consumer; potentially reusable.

### Zombie Abstraction: `state.json` `last_sync` Field
**Location:** `internal/sync/state.go` line 19
`active_rite` field was removed. `last_sync` field survives with zero confirmed runtime read consumers. Full `state.json` elimination is DEBT-039.

### Premature Abstraction: `internal/registry/registry.go`
~10 keys, consulted by 2 packages. `panic`-on-unknown-key appropriate for compile-time safety. Borderline; not a significant burden.

### Zombie Code: `printAgentList` Function
**Location:** `internal/cmd/agent/list.go` line 199
Explicitly marked unused dead code. No callers. Safe to remove.

---

## Load-Bearing Code Identification

### LOAD-001: `provenance.Save()` Structural Equality Guard
**Location:** `internal/provenance/manifest.go`
Before writing `PROVENANCE_MANIFEST.yaml`, checks structural equality and skips write if only timestamps changed. Prevents triggering CC file watcher on no-op syncs.
**What a naive fix would break:** Removing guard causes infinite re-sync loops in active CC sessions.

### LOAD-002: `fileutil.WriteIfChanged()` Atomic Write with Change Detection
**Location:** `internal/fileutil/fileutil.go`
Reads existing file, compares bytes, only calls `AtomicWriteFile` if content differs.
**What a naive fix would break:** Direct `os.WriteFile` eliminates atomicity AND triggers CC file watcher on every sync.

### LOAD-003: `inscription.MergeRegions()` Satellite Region Preservation
**Location:** `internal/inscription/merger.go`
Merges knossos regions, preserving `satellite` and `regenerate` regions. The platform invariant "User content NEVER destroyed" depends on this function.

### LOAD-004: `namespace.resolveNamespace()` Provenance-Based Collision Detection
**Location:** `internal/materialize/mena/namespace.go` lines 20-204
Before projecting dromena to `.claude/commands/`, reads provenance to distinguish knossos-owned from user-owned entries. Knossos yields on collision.
**What a naive fix would break:** Removing provenance check causes knossos to silently overwrite user commands.

### LOAD-005: `session.FSM.ValidateTransition()` State Machine Enforcer
**Location:** `internal/session/fsm.go`
Enforces 4-state session FSM. `Archived` is terminal.
**What a naive fix would break:** Adding a state without updating all lifecycle handlers, writeguard, CC session map, and event constructors.

### LOAD-006: `rite.Syncer` Interface as Dependency Inversion Boundary
**Location:** `internal/rite/syncer.go`
Single-method interface breaking upward dependency from `internal/rite` → `internal/materialize`. The entire import graph separation depends on this.

### LOAD-007: `sails.CheckGate()` as Session Quality Gate
**Location:** `internal/sails/gate.go`
Called by session wrap, sails check, and naxos scanner. Determines if a session's confidence signal passes the quality gate.
**What a naive fix would break:** Changing color enum or pass logic affects all three consumers.

---

## Evolution Constraint Documentation

### Changeability Matrix

| Area | Rating | Notes |
|------|--------|-------|
| `internal/errors/` | **SAFE** | Leaf package. Adding codes is additive. Do not reorder `ExitCode` constants. |
| `internal/checksum/` | **SAFE** | `sha256:` prefix is contract with provenance — do not change without migration. |
| `internal/fileutil/` | **SAFE** | `WriteIfChanged` signature change requires updating 20+ call sites. |
| `internal/registry/` | **SAFE** | Adding keys additive; removing keys panics on missing. |
| `internal/paths/` | **COORDINATED** | `FindProjectRoot()` and `Resolver` called throughout. |
| `internal/session/fsm.go` | **COORDINATED** | State machine changes require updating all lifecycle handlers simultaneously. |
| `internal/inscription/` | **COORDINATED** | `DefaultSectionOrder()` and `DeprecatedRegions()` govern CLAUDE.md for all satellites. |
| `internal/provenance/` | **MIGRATION** | Schema version `2.0` with migration shim. Bump requires manifest migration in all projects. |
| `internal/materialize/materialize.go` | **COORDINATED** | Pipeline stage reordering requires checking all `opts.Soft` guards. |
| `internal/hook/env.go` stdin-only transport | **SETTLED** | Env var fallback removed. Only `CLAUDE_PROJECT_DIR` read from env. All other hook data via stdin JSON. |
| `internal/config/home.go` | **COORDINATED** | `sync.Once` cache is test-hazardous. Reset required in tests setting `KNOSSOS_HOME`. |
| `internal/sails/color.go` | **COORDINATED** | Color enum is contractual across gate, generator, naxos scanner, tribute, and wrap. |
| `internal/hook/clewcontract/` | **COORDINATED** | Canonical event write path. Schema changes require updating `events_read.go` bridge and `tribute.EventData`. |

### Deprecated Markers

| Item | Location | Status |
|------|----------|--------|
| `materialize.RiteManifest.Skills` | `internal/materialize/materialize.go:68` | `// Deprecated: use Legomena` |
| `materialize.RiteManifest.Commands` | `internal/materialize/materialize.go:67` | Backward compat; no removal plan |
| `hook.Env` env var constants | `internal/hook/env.go` | `// Deprecated: CC sends via stdin JSON` |
| Legacy `session.Event` v1 struct | `internal/session/events_read.go:30` | Read-path only; removal after pre-ADR-0027 sessions archived |
| `inscription.DeprecatedRegions()` | `internal/inscription/manifest.go:303` | Active list of regions dropped on next sync |
| `printAgentList` function | `internal/cmd/agent/list.go:199` | Dead code, safe to remove |
| `validate --schema` flag | `internal/cmd/validate/validate.go:573` | Backward compat, not used |
| `sync.State.LastSync` field | `internal/sync/state.go:19` | Candidate for DEBT-039 elimination |

### External Dependency Constraints

| Dependency | Constraint |
|------------|-----------|
| `CGO_ENABLED=0` | All packages must compile without CGO. Eliminated SQLite as option. |
| `go 1.23.0` minimum | `go.mod` minimum; `toolchain go1.24.13` actual build tool. |
| CC hook stdin JSON protocol | CC delivers hook data via stdin JSON, not env vars. New fields must come from `StdinPayload`. |
| XDG directory conventions | Platform path diverges macOS vs. Linux. Both `config` and `paths` implement independently. |
| `adrg/xdg` library | Used only by `internal/paths`. `internal/config` reimplements to stay leaf. |
| Sprig template functions | `Masterminds/sprig/v3` in `go.mod`. Used in inscription template engine. |

---

## Risk Zone Mapping

### RISK-002: `materializeMena()` Discards Namespace Collision Warnings
**Location:** `internal/materialize/materialize_mena.go` line 114
`_, err := SyncMena(sources, opts)` discards `*MenaProjectionResult`. Namespace collision warnings lost at the `materialize/` boundary.
**Recommended fix:** Change `_` to named result, propagate warnings into `RiteScopeResult`.

### RISK-003: `config.KnossosHome()` Once-Cached with No Production Invalidation
**Location:** `internal/config/home.go` lines 11-23
Tests that call `KnossosHome()` transitively before setting `KNOSSOS_HOME` silently poison the cache.
**Guard:** All tests requiring custom `KNOSSOS_HOME` must call `config.ResetKnossosHome()`.

### RISK-004: `mena/engine.go` Stale Entry Removal Silently Discards Errors
**Location:** `internal/materialize/mena/engine.go`
`removeStaleFiles()` discards individual removal errors. Stale mena files accumulate silently.

### RISK-005: `naxos/scanner.go` Reads `sails.yaml` Instead of `WHITE_SAILS.yaml`
**Location:** `internal/naxos/scanner.go` lines 210-213
The canonical sails file is `WHITE_SAILS.yaml` (used by `sails.CheckGate()`, `cmd/session/status.go`, `cmd/land/synthesize.go`). The naxos scanner reads a different path, returning empty color for sessions using the canonical filename.
**Recommended fix:** Update to read `WHITE_SAILS.yaml`.

### RISK-006: `internal/mena/` vs `internal/materialize/mena/` Package Naming Ambiguity
**Location:** Two packages share the name `mena` at different path depths.
**Risk:** Developer confusion (compiler catches wrong import immediately).
**Recommended guard:** Add package doc comment to `internal/mena/` distinguishing from materialization-side.

## Knowledge Gaps

1. **TENSION-009 import feasibility** — whether `tribute` can import `clewcontract` directly has not been verified against the full import graph.
2. **DEBT-039 (`state.json` `last_sync` elimination)** — whether `State.LastSync` has read consumers beyond initialization is unconfirmed.
3. **RISK-005 root cause** — whether `naxos/scanner.go` reading `sails.yaml` is intentional or a stale rename is unconfirmed.
4. **`internal/know/` incremental cycle design** — the interaction between `DependsOn` graph edges, `LandSources` tracking, and incremental vs. full regeneration modes is partially mapped.
5. **`internal/sails/` threshold configuration** — whether thresholds are configurable per-rite or global is unconfirmed.

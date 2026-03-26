---
domain: design-constraints
generated_at: "2026-03-23T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "78abb186"
confidence: 0.91
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "08704d8d5bd3b4b8b2c6a8a1590586893f35aa5a3c3d6f27c260cee86fca82b0"
---

# Codebase Design Constraints

## Tension Catalog Completeness

### TENSION-001: Dual `OwnerType` Types for Different Ownership Concepts

**Location**: `internal/provenance/provenance.go:91` and `internal/inscription/types.go:14`

**Type**: Naming mismatch / semantic divergence

Two distinct `OwnerType` types exist in separate packages:
- `provenance.OwnerType`: `knossos / user / untracked` — file ownership in the channel dir
- `inscription.OwnerType`: `knossos / satellite / regenerate` — region ownership in CLAUDE.md

Both share the value `"knossos"` for platform-managed content but diverge from there.

**Historical reason**: Inscription and provenance were designed independently. Merging would require provenance (a leaf package per ADR-0026) to import inscription's type definitions, creating an import cycle.

**Ideal resolution**: Define a shared `owner` package that both import. Estimated effort: medium.

---

### TENSION-002: `channelDirOverride` Mutation-Based Dispatch Hack

**Location**: `internal/materialize/materialize.go:118,284-289,385-390`

**Type**: Over-engineering smell / structural hack

The `Materializer` struct has a `channelDirOverride string` field that is mutated-then-deferred to redirect the Gemini channel write target.

**Historical reason**: Documented in ADR-0031 as "a pragmatic hack." Reuse of the existing pipeline without constructor changes.

**Ideal resolution**: Thread `channel string` through the constructor or pass it as a pipeline parameter. Blocked by the scope of callers.

---

### TENSION-003: `perspective` Package Hardcodes `.claude/` Channel Directory

**Location**: `internal/perspective/context.go:60` and `resolvers.go:274`

**Type**: Layering violation / harness-agnosticism gap

The perspective package resolves channel directory unconditionally as `.claude/`. Both lines carry `// HA-FS:` markers. Gemini channel cannot produce a valid perspective document.

**Ideal resolution**: Add `channel string` to `perspective.Options`; use `paths.ChannelByName()`. Medium effort.

---

### TENSION-004: CC-Native Frontmatter Fields Embedded in `AgentFrontmatter`

**Location**: `internal/agent/frontmatter.go:33-40`

**Type**: Abstraction gap / harness-agnosticism

`AgentFrontmatter` uses CC camelCase wire-format field names as its canonical Go struct fields: `maxTurns`, `disallowedTools`, `permissionMode`, `mcpServers`, `hooks`. The Gemini compiler strips most of them. There is no intermediate canonical representation.

**Historical reason**: ADR-0032 defined canonical equivalents (`BehavioralContract.MaxTurns`), but the struct was never migrated.

---

### TENSION-005: `resolution` Package Zero-Import Policy Creates DI Burden

**Location**: `internal/resolution/chain.go:1-6`

**Type**: Under/over-engineering tension

The package explicitly documents: `"This package has ZERO internal imports."` This forces callers to construct `Tier{}` structs with raw string paths.

**Historical reason**: Prevents import cycles from higher-level packages.

**Ideal resolution**: Acceptable as-is. The zero-import discipline is load-bearing.

---

### TENSION-006: `provenance.SourceType` Uses Plain Strings to Avoid Import Cycle

**Location**: `internal/provenance/provenance.go:68-72`

**Type**: Naming mismatch / leaf package constraint

Provenance uses untyped `string` for `SourceType` rather than importing `source.SourceType`. The comment reads: "String values must stay in sync manually."

**Historical reason**: Importing `materialize/source` from `provenance` would break the leaf package guarantee of ADR-0026.

---

### TENSION-007: `GetAdapter()` Defaults to `ClaudeAdapter` When `KNOSSOS_CHANNEL` Unset

**Location**: `internal/hook/env.go:86-94`

**Type**: Harness-agnosticism gap / backward-compatibility constraint

When `KNOSSOS_CHANNEL` is unset, `GetAdapter()` silently returns `&ClaudeAdapter{}`. Tracked as PKG-010 in harness-agnosticism initiative.

---

### TENSION-008: `Soft bool // CC-safe mode` Comment Is Channel-Specific

**Location**: `internal/materialize/sync_types.go:51`; `materialize.go:29`

**Type**: Naming mismatch (cosmetic)

Soft mode is functionally channel-agnostic but the comment labels it `"CC-safe mode"`.

---

### TENSION-009: `userscope/sync.go` Hardcodes `"claude"` Fallback Default

**Location**: `internal/materialize/userscope/sync.go:33`

**Type**: Harness-agnosticism gap

Defaults to `paths.UserChannelDir("claude")` when no channel configured. Gemini user files never synced unless caller explicitly passes `userChannelDir`.

---

### TENSION-010: `knownTools` Validator Is CC-Specific

**Location**: `internal/agent/frontmatter.go:55-70`

**Type**: Harness-agnosticism gap

The `knownTools` map contains CC wire tool names only. Gemini tool names not in the validator.

---

### TENSION-011: `RiteManifest.Skills` and `.Commands` Deprecated Fields Present

**Location**: `internal/materialize/materialize.go:69-70`

**Type**: Zombie abstraction

Both fields must be checked by every manifest reader. No removal date set.

---

### TENSION-012: `PreToolUseOutput.HookEventName = "PreToolUse"` Is Frozen CC Wire Format

**Location**: `internal/hook/output.go:28`

**Type**: Frozen / wire protocol constraint

The value `"PreToolUse"` is the CC protocol value. Changing it would bypass the write guard silently. This is a security boundary.

---

### TENSION-013: `inscription` Package `NewPipeline()` Hardcodes `.claude/CLAUDE.md`

**Location**: `internal/inscription/pipeline.go:156-163`

**Type**: Harness-agnosticism gap

The default constructor hardcodes the CC channel context file path. Gemini requires `NewPipelineWithPaths()`.

---

### TENSION-014: `tribute/Commits` Section Is a Phase 2 Placeholder

**Location**: `internal/tribute/types.go:58,158`; `renderer.go:64,250`

**Type**: Premature abstraction (incomplete)

`Commit` struct and `renderCommits()` exist but are marked as "Phase 2 - placeholder." No commit data is populated.

---

### TENSION-015: `search/collectors.go` Imports `cmd/explain` Across Layer Boundary

**Location**: `internal/search/collectors.go:12`

**Type**: Layering violation (latent circular risk)

`internal/search` imports `internal/cmd/explain`. If `cmd/explain` ever imports `search`, a circular dependency results. Currently safe — risk is latent.

---

### TENSION-016: `config.KnossosHome()` Uses `sync.Once` Singleton — Test Cache Poisoning

**Location**: `internal/config/home.go:11-23`

**Type**: Load-bearing jank / test isolation risk

A test that calls `KnossosHome()` before setting `KNOSSOS_HOME` poisons the cache for all subsequent tests. `ResetKnossosHome()` must be called.

---

### TENSION-017: `.mcp.json` Written to Project Root, Not Channel Dir

**Location**: `internal/materialize/materialize.go:570-577` (SCAR-028)

**Type**: Structural asymmetry

MCP servers in `.mcp.json` at project root; hooks in `{channelDir}/settings.local.json`. Two-location config split.

---

## Trade-off Documentation

### Trade-off 1: WriteIfChanged Over Atomic Write for CC Stability

**Chosen**: `fileutil.WriteIfChanged()` for all writes to `.claude/`
**Rejected**: Plain `os.WriteFile()` or unconditional atomic rename
**Why current persists**: CC's file watcher crashes on DELETE events. `WriteIfChanged` reads first; if identical, no write occurs.

---

### Trade-off 2: Dual Read Path for Events (v1/v2/v3) Instead of Migration

**Chosen**: Format-sniffing reader at `session/events_read.go:82-117`
**Rejected**: One-time migration script
**Why current persists**: Sessions may resume across format versions. Safe to remove once all pre-ADR-0027 sessions are archived.

---

### Trade-off 3: Provenance Manifest Lives in `.knossos/`, Not `.claude/`

**Chosen**: `PROVENANCE_MANIFEST.yaml` in `.knossos/`
**Rejected**: Storing provenance in `.claude/`
**Why current persists**: `.knossos/` is gitignored infrastructure. Writing to `.claude/` exposes internal tracking to CC's context window.

---

### Trade-off 4: CC-Canonical Hook Event Names as Internal Lingua Franca

**Chosen**: CC event names (`PreToolUse`, `PostToolUse`) as internal canonical
**Rejected**: A third neutral vocabulary
**Why current persists**: ADR-0032 explicitly rejected a third vocabulary. Translation at adapter boundary.

---

### Trade-off 5: XDG Config Path Inlined in `config.ActiveOrg()` to Avoid Import Cycle

**Chosen**: Duplicate XDG logic in `config.ActiveOrg()`
**Rejected**: Import `internal/paths` from `internal/config`
**Why current persists**: `paths` imports `config`. Reverse import creates cycle.

---

### Trade-off 6: Two-Manifest Architecture (Rite + User Provenance)

**Chosen**: Separate `PROVENANCE_MANIFEST.yaml` and `USER_PROVENANCE_MANIFEST.yaml`
**Rejected**: Single unified manifest
**Why current persists**: Rite scope is per-project; user scope is global. Merging requires project-keyed namespacing.

---

### Trade-off 7: `channelDirOverride` Save-and-Restore vs. Constructor Threading

**Chosen**: Mutation-with-defer pattern
**Rejected**: Thread channel through constructor
**Why current persists**: ADR-0031 acknowledges "pragmatic hack." Constructor refactor touches 15+ call sites.

---

### Trade-off 8: `ClaudeCompiler` as Pass-Through (No Transformation)

**Chosen**: `ClaudeCompiler` returns content unchanged
**Rejected**: Symmetric pipeline where both compilers transform
**Why current persists**: CC consumes raw markdown. No transformation required.

---

## Abstraction Gap Mapping

### AGM-001: `perspective` Package Has No Channel Abstraction

**Location**: `internal/perspective/context.go:60`, `resolvers.go:274`

Both hardcode `.claude/` paths. Blocks Gemini feature parity for `ari ask`.

---

### AGM-002: `AgentFrontmatter` Mixes Canonical and CC-Wire Fields

**Location**: `internal/agent/frontmatter.go:17-51`

CC-wire `maxTurns` field and canonical `BehavioralContract.MaxTurns` coexist. The canonical path is designed but never connected.

---

### AGM-003: `AGENTS.md` Compilation Target Designed But Not Implemented

ADR-0032 references `AGENTS.md` as a third compilation target. No implementation exists.

---

### AGM-004: `BehavioralContract.MaxTurns` Not Wired to CC Wire Format

**Location**: `internal/agent/frontmatter.go:47-48`

The compiler pipeline reads from the direct CC-wire field, not from `Contract.MaxTurns`.

---

### AGM-005: `RiteManifest.Commands` and `.Skills` Backward-Compat Fields

**Location**: `internal/materialize/materialize.go:68-70`

Every manifest reader must check both deprecated and current field names. No removal date.

---

### AGM-006: Duplicate Resolution Logic in `materialize/source` and `resolution` Packages

Both implement priority-ordered tier traversal. Moderate code duplication.

---

### AGM-007: `tribute/Commits` Placeholder Occupies API Surface Without Implementation

**Location**: `internal/tribute/types.go:58,158`; `renderer.go:250`

Type surface committed, git integration deferred indefinitely.

---

### AGM-008: `fileutil.AtomicWriteFile` vs. User-Scope `os.WriteFile`

**Location**: `internal/materialize/userscope/sync.go:50`

Rite scope uses `WriteIfChanged()`; user scope uses `os.WriteFile`. Documented as intentional but asymmetric.

---

## Load-Bearing Code Identification

### LB-001: `fileutil.WriteIfChanged()` — CC Stability Invariant

**File**: `internal/fileutil/fileutil.go:66-72`

Every write to `.claude/` passes through this function. Removing the equality check triggers CC file watcher crashes.

---

### LB-002: `provenance.structurallyEqual()` — Timestamp-Only Write Suppression

Avoids writing the provenance manifest when only timestamps changed.

---

### LB-003: `applyRewrites()` Three-Pass Ordering in `content_rewrite.go`

**File**: `internal/materialize/mena/content_rewrite.go:128-138`

`INDEX.lego.md -> SKILL.md` must be applied before general `{name}.lego.md -> {name}.md`. Fence-split pass precedes both.

---

### LB-004: `ReadEvents()` Format Detection Order

**File**: `internal/session/events_read.go:82-117`

v3 events must be checked first, before v2 and v1. Wrong order silently misreads all v3 events.

---

### LB-005: `RenameV2Type()` — Append-Only Event Rename Map

**File**: `internal/hook/clewcontract/type_rename.go:14-22`

Must be append-only while any v2 event files exist on disk.

---

### LB-006: `config.KnossosHome()` — sync.Once Singleton

**File**: `internal/config/home.go:11-23`

Load-bearing for performance; creates test cache poisoning risk.

---

### LB-007: `materializeAgents()` — NO Pre-Delete Before Overwrite

**File**: `internal/materialize/materialize_agents.go:42-45`

Pre-deletion causes CC file watcher DELETE events, crashing active sessions.

---

### LB-008: Satellite Region Preservation in `inscription.Merger`

Regions with `Owner == OwnerSatellite` are never overwritten. Core invariant protecting user content in CLAUDE.md.

---

### LB-009: `provenance.LoadOrBootstrap()` — Abort-on-Corrupt Contract

All errors except file-not-found propagate and abort the pipeline.

---

### LB-010: `hook/output.go` PreToolUse Wire Format String

**File**: `internal/hook/output.go:28`

`HookEventName: "PreToolUse"` is the CC wire protocol value. If changed, CC bypasses writeguard. Security boundary.

---

### LB-011: `materialize.go` Pipeline Step Ordering

**File**: `internal/materialize/materialize.go:457-627`

10+ step pipeline with implicit ordering dependencies. Non-transactional — partial failure leaves inconsistent state.

---

## Evolution Constraint Documentation

### SAFE areas (local changes only)

- `internal/errors/` — adding new error codes is contained
- `internal/session/fsm.go` — adding states is self-contained
- `internal/channel/tools.go` — adding tool mappings is additive
- `internal/resolution/chain.go` — purely algorithmic
- `internal/tribute/types.go` `Commit` struct — placeholder, zero-risk
- `internal/lock/lock.go` `DefaultTimeout` — safe; `StaleThreshold` is load-bearing for long commands

### COORDINATED areas (require cross-file changes)

- **Adding a new `TargetChannel`**: Minimum 5 files across 4 packages
- **Changing `provenance.SourceType` values**: Must sync with `materialize/source/types.go`
- **Adding a new inscription region**: Schema version bump, template, manifest migration
- **Adding a new `HookEvent`**: Constant, translation tables, hooks.yaml, switch statements
- **Removing `RiteManifest.Skills` or `.Commands`**: All `manifest.yaml` files across all satellite repos

### MIGRATION areas (on-disk migration required)

- `provenance.CurrentSchemaVersion = "2.0"` — v3 bump requires new migration function
- `inscription.DefaultSchemaVersion = "1.0"` — format changes require manifest migration
- `RiteManifest.Commands` and `.Skills` — removal requires all satellite repo manifests updated
- MCP server location split (SCAR-028): stale entries cleaned by `cleanupStaleBlanketSettings()`

### FROZEN areas (cannot change without breaking wire protocol)

- `hook/output.go` `HookEventName: "PreToolUse"` — CC protocol value
- `.claude/` directory name — CC hardcodes this
- `PROVENANCE_MANIFEST.yaml` filename for Claude channel
- `KNOSSOS_CHANNEL` environment variable name
- `.knossos/` directory name — used by `FindProjectRoot()`

### Deprecated markers (still present)

| Item | Location | Replacement |
|---|---|---|
| `RiteManifest.Skills` | `materialize.go:70` | `Legomena` |
| `RiteManifest.Commands` | `materialize.go:69` | `Dromena` + `Legomena` |
| `paths.AgentsDir()` | `paths/paths.go:169` | `AgentsDirForChannel()` |
| `materialize.ResolveUserResources()` | `materialize/source.go:71` | `ResolveUserResourcesForChannel()` |

### In-progress migrations (parked or active)

- ADR-0032 PKG-010: `GetAdapter()` CC default (TENSION-007, parked)
- `perspective` channel-awareness (TENSION-003, unresolved)
- ADR-0032 `AGENTS.md` compilation target (AGM-003, not implemented)
- `BehavioralContract.MaxTurns` wiring (AGM-004, not connected)
- `tribute` git commits Phase 2 (AGM-007, deferred indefinitely)

---

## Risk Zone Mapping

### RZ-001: `materialize/` — No Rollback on Partial Write Failure

**Location**: `internal/materialize/materialize.go:454-627`

10+ step pipeline is not transactional. Mid-pipeline failure leaves agents on disk with no CLAUDE.md update.

---

### RZ-002: `config.KnossosHome()` Test Cache Poisoning

**Location**: `internal/config/home.go:81-96`

A test calling `KnossosHome()` before setting `KNOSSOS_HOME` poisons `sync.Once` for all subsequent tests.

---

### RZ-003: `search/collectors.go` Circular Layer Risk

**Location**: `internal/search/collectors.go:12`

`internal/search` imports `internal/cmd/explain`. Soft constraint enforced only by discipline.

---

### RZ-004: Mena Namespace Collision Across Rites

Collisions produce warnings but no hard failure. Callers may silently deliver incomplete mena.

---

### RZ-005: Lock Stale Threshold Is Time-Based (5 min), Not Process-Based

**Location**: `internal/lock/lock.go:32`

Long-running commands (>5 min) may have locks classified as stale. No heartbeat mechanism.

---

### RZ-006: Org Scope Sync Silently Non-Fatal on Error

**Location**: `internal/materialize/org_scope.go:63-73`

Per-channel errors accumulated but execution continues. Misconfigured org silently produces partial sync.

---

### RZ-007: `userscope/sync.go` Fallback Channel Default

**Location**: `internal/materialize/userscope/sync.go:33`

Defaults to `"claude"` when no channel configured. No error or warning emitted.

---

### RZ-008: `perspective/context.go` Channel Dir Hardcode

**Location**: `internal/perspective/context.go:60`, `resolvers.go:274`

Gemini perspective assembly resolves against `.claude/` unconditionally. Silently incorrect results.

---

### RZ-009: `procession/resolver.go` Calls Global Config Functions Directly

**Location**: `internal/materialize/procession/resolver.go:36-50`

`ResolveProcessions()` calls global config functions rather than injected parameters. Test isolation requires env var overrides.

---

### RZ-010: `isGitWorktree()` External Process Call With 10s Timeout

**Location**: `internal/materialize/worktree.go:22-49`

Git subprocess in materialization path. On NFS/CI, adds latency. On timeout, returns `false` (safe default) but misses rite inheritance.

---

## Knowledge Gaps

1. **`internal/naxos/triage.go`**: Debt triage scoring heuristics not inspected.
2. **`internal/sails/`**: White sails signaling constraints not traced.
3. **`internal/session/fsm.go`**: State machine transition invariants not fully enumerated.
4. **`internal/procession/` (non-materialize)**: Template schema validation constraints not verified.
5. **`internal/validation/schemas/`**: JSON schema versioning constraints not documented.
6. **`hooks.yaml` canonical structure**: Hook configuration format constraints not inspected end-to-end.
7. **`internal/registry/`**: Registry package constraints on rite registration not examined.
8. **`internal/materialize/mcp_ownership.go`**: MCP ownership model constraints not captured.

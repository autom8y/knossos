---
domain: design-constraints
generated_at: "2026-03-13T10:04:06Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "59a0de2"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "08704d8d5bd3b4b8b2c6a8a1590586893f35aa5a3c3d6f27c260cee86fca82b0"
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Dual `OwnerType` for Different Ownership Concepts

**Location**: `internal/provenance/provenance.go` and `internal/inscription/types.go`

Two distinct `OwnerType` types exist: provenance uses `knossos / user / untracked` (file ownership), inscription uses `knossos / satellite / regenerate` (region ownership). Both use `"knossos"` for platform-managed, but semantics diverge. Both files carry inline TENSION-001 comments.

**Why it exists**: Inscription and provenance are separate concerns with different lifecycle semantics. Merging would couple a leaf package (provenance) to inscription's region model.

---

### TENSION-002: `channelDirOverride` as a Mutation-Based Dispatch Hack

**Location**: `internal/materialize/materialize.go` lines 103, 195-196, 269-273, 370-374

The `Materializer` uses a `channelDirOverride string` field, mutated and restored with `defer`, to re-route materialization to `.gemini/` instead of `.claude/`. Acknowledged as "a pragmatic hack" in ADR-0031.

**Why it exists**: Reuses entire materialization pipeline for Gemini without forking. Threading channel through constructor would require larger refactor.

---

### TENSION-003: `perspective` Package Hardcodes `.claude/` Channel Directory

**Location**: `internal/perspective/context.go` line 60

The perspective package resolves channel directory unconditionally to `.claude/`. Gemini agents cannot produce correct perspective documents. Unresolved gap post-ADR-0031.

---

### TENSION-004: CC-Native Frontmatter Fields Embedded in Agent Struct

**Location**: `internal/agent/frontmatter.go`

`AgentFrontmatter` contains CC wire-format fields (`maxTurns`, `disallowedTools`, `permissionMode` in camelCase). ADR-0032 defined canonical equivalents but the struct still holds CC-wire names as primary. The compiler layer must translate at output time.

---

### TENSION-005: `resolution` Package Zero-Import Policy vs. Import Graph Complexity

**Location**: `internal/resolution/chain.go`

The package documents: "ZERO internal imports. All tier paths are injected via constructor to avoid import cycles." This creates a DI burden for callers who must construct `Tier` structs with raw string paths.

---

### TENSION-006: `provenance.SourceType` Uses Plain Strings to Avoid Import Cycle

**Location**: `internal/provenance/provenance.go` lines 68-72

Provenance is a leaf package (ADR-0026). It uses plain strings rather than importing `source.SourceType`. String values must stay in sync with `internal/materialize/source/types.go` manually.

---

### TENSION-007: `GetAdapter()` Defaults to `ClaudeAdapter` When `KNOSSOS_CHANNEL` Unset

**Location**: `internal/hook/env.go` line 88

When KNOSSOS_CHANNEL is unset, hooks default to CC semantics. Tracked as PKG-010 in harness-agnosticism initiative.

---

### TENSION-008: Soft Mode Comment Still References CC Name

**Location**: `internal/materialize/sync_types.go` line 51; `internal/materialize/materialize.go` line 29

`Soft bool // CC-safe mode` — functionally channel-agnostic but comment is CC-specific.

---

### TENSION-009: `userscope` Hardcodes `"claude"` Fallback Default

**Location**: `internal/materialize/userscope/sync.go` line 33

`userChannelDir = paths.UserChannelDir("claude")` — unparameterized invocations silently target only Claude.

---

### TENSION-010: `AgentFrontmatter.knownTools` Is CC-Specific

**Location**: `internal/agent/frontmatter.go` line 54

The `knownTools` map uses CC wire names (`Bash`, `Grep`, `Read`). Gemini wire names are unknown to this validator.

---

### TENSION-011: `RiteManifest.Skills` Field Is Deprecated But Present

**Location**: `internal/materialize/materialize.go` line 70

`Skills []string` deprecated in favor of `Legomena`. `Commands []string` similarly backward-compat.

---

### TENSION-012: `PreToolUseOutput` Uses CC Wire Format by Decree

**Location**: `internal/hook/output.go` line 27

`HookEventName: "PreToolUse"` is frozen — CC reads it directly. Canonical vocabulary does not apply here. Comment: "CC wire format (SCAR-009) -- do not change to canonical name".

---

## Trade-off Documentation

### Trade-off 1: WriteIfChanged Over Atomic Write for CC Stability

**Chosen**: `fileutil.WriteIfChanged()` everywhere in `.claude/`.
**Rejected**: Simple `os.WriteFile()` or unconditional atomic write.
**Why**: CC's file watcher crashes on DELETE events from temp-file-then-rename patterns. `WriteIfChanged()` adds read overhead but eliminates watcher churn. Provenance manifest adds `structurallyEqual()` check on top.

### Trade-off 2: Dual Read Path for Events (v1/v2/v3) Instead of Format Migration

**Chosen**: Format-sniffing reader normalizing all three formats at read time.
**Rejected**: One-time migration script.
**Why**: Sessions may be resumed across format versions. In-place migration risks data loss. Safe to remove once all pre-ADR-0027 sessions are archived.

### Trade-off 3: Provenance Manifest in .knossos/, Not .claude/

**Chosen**: `PROVENANCE_MANIFEST.yaml` in `.knossos/`.
**Rejected**: Storing provenance in `.claude/`.
**Why**: `.knossos/` is gitignored infrastructure. Writing to `.claude/` would expose internal tracking to CC's context window.

### Trade-off 4: Aggressive Orphan Auto-Removal vs. User Safety

**Chosen**: `--remove-all` auto-removes knossos-owned orphans after backup.
**Rejected**: Always prompt; always keep orphans.
**Why**: Stale agents pollute CC agent pool. Divergence detection (`DetectDivergence`) promotes to user ownership if checksum changed.

### Trade-off 5: Inlining XDG Config Path to Avoid Circular Import

**Chosen**: Duplicate XDG logic in `config.ActiveOrg()`.
**Rejected**: Import `internal/paths` from `internal/config`.
**Why**: `paths` imports `config`. Reverse import would create cycle. Comment at `internal/config/home.go:60` documents this.

### Trade-off 6: Two-Manifest Architecture (Rite + User Provenance)

**Chosen**: Separate `PROVENANCE_MANIFEST.yaml` (project) and `USER_PROVENANCE_MANIFEST.yaml` (user global).
**Rejected**: Single unified manifest.
**Why**: Rite scope is per-project; user scope is global. Merging requires project-keyed namespacing.

---

## Abstraction Gap Mapping

### AGM-001: `perspective` Package Has No Channel Abstraction

The perspective package reads from `~/.claude/` paths hardcoded in `context.go`. No `TargetChannel` parameter. Features like `ari ask` and agent simulation only function for CC.

### AGM-002: `AgentFrontmatter` Mixed Canonical and CC-Wire Fields

The source struct is CC-biased. The compiler exists but the struct still holds CC wire-format fields. Translation to canonical happens at output, not at the model level.

### AGM-003: `AGENTS.md` Compilation Target Designed But Not Implemented

ADR-0032 specifies `AGENTS.md` as a third compilation target. No implementation found. The `ChannelCompiler` interface has no `interop` implementation.

### AGM-004: `BehavioralContract.MaxTurns` Not Wired to CC Wire Format

`AgentFrontmatter.MaxTurns` (CC wire) and `BehavioralContract.MaxTurns` (canonical) are two representations. The compiler does not translate the canonical path. Zombie abstraction gap.

### AGM-005: `RiteManifest.Commands` Backward-Compat Field

Retained solely for old manifests. Every reader must check `Commands` and fall back if `Dromena`/`Legomena` are empty.

### AGM-006: Duplicate Resolution in `materialize` vs `resolution` Package

Both `resolution.Chain` and `source.SourceResolver` implement priority-ordered tier traversal. Mild over-engineering duplication.

---

## Load-Bearing Code Identification

### LB-001: `fileutil.WriteIfChanged()` -- CC Stability Invariant

**File**: `internal/fileutil/fileutil.go:66-72`
Every write to `.claude/` passes through this. Called from 30+ production sites. Removing the equality check would trigger CC file watcher crashes.

### LB-002: `provenance.structurallyEqual()` -- Timestamp-Only Write Suppression

**File**: `internal/provenance/manifest.go:86-110`
Avoids writing manifest when only timestamps changed. Without this, every sync triggers CC file watcher.

### LB-003: `applyRewrites()` Ordering in content_rewrite.go

**File**: `internal/materialize/mena/content_rewrite.go:131-141`
Three-pass ordering is load-bearing. `INDEX.lego.md -> SKILL.md` must precede general `{name}.lego.md -> {name}.md`.

### LB-004: `ReadEvents()` Format Detection Order

**File**: `internal/session/events_read.go:82-117`
v3 (`"data"` field) must always be checked first. v3 events contain both `"data"` and `"type"` fields.

### LB-005: `RenameV2Type()` -- Append-Only Event Rename Map

**File**: `internal/hook/clewcontract/type_rename.go:14-22`
Must remain append-only while any v2 events exist. Never remove entries until v2 format is fully retired.

### LB-006: `config.KnossosHome()` -- sync.Once Singleton

**File**: `internal/config/home.go:11-23`
Load-bearing for performance but creates RISK-003 (test cache poisoning). Always use `config.ResetKnossosHome()` in tests.

### LB-007: `materializeAgents()` -- NO Pre-Delete Before Overwrite

**File**: `internal/materialize/materialize_agents.go:37-40`
Pre-deletion causes CC DELETE events and crashes. `writeIfChanged()` handles overwrite atomically.

### LB-008: Satellite Region Preservation in inscription.Merger

**Location**: `internal/inscription/pipeline.go:286-292`
Regions with `Owner == OwnerSatellite` are never overwritten. This protects user content in CLAUDE.md.

### LB-009: `provenance.LoadOrBootstrap()` -- Abort-on-Corrupt Contract

Callers in `materialize.go` lines 299 and 462: corrupted manifest must propagate error, never silently bootstrap empty. Changing this would mask data corruption and overwrite user files.

### LB-010: `hook/output.go` PreToolUse Wire Format String

**File**: `internal/hook/output.go` line 28
`HookEventName: "PreToolUse"` is CC protocol value. If changed to canonical `"pre_tool"`, CC bypasses write guard. Security boundary.

---

## Evolution Constraint Documentation

### SAFE areas (local changes only)

- `internal/errors/` — adding new error codes is contained
- `internal/session/fsm.go` — adding states requires updating transitions map + tests
- `internal/channel/tools.go` — adding canonical tool mappings is additive
- `internal/resolution/chain.go` — purely algorithmic, safe to enhance

### COORDINATED areas (require cross-file changes)

- Adding a new `TargetChannel`: requires `TargetChannel` interface, `ChannelCompiler`, translation tables in `hook/events.go` and `channel/tools.go`. 4+ files across 3+ packages.
- Changing `provenance.SourceType` values: must stay in sync with `source/types.go` (TENSION-006). Migration required.
- Adding new inscription sections: requires `DefaultSchemaVersion`, `SectionOrder`, template files. 3+ files.
- Adding new `HookEvent` constants: requires constants, translation tables, `hooks.yaml`, switch statements across `internal/cmd/hook/`.

### MIGRATION areas (on-disk migration required)

- `provenance.CurrentSchemaVersion = "2.0"` — migration path exists. v3 bump requires `migrateV2ToV3()`.
- `inscription.DefaultSchemaVersion = "1.0"` — format change requires migration.
- `RiteManifest.Commands` and `.Skills` — removal requires all manifest.yaml files to be updated.

### FROZEN areas (cannot change without breaking wire protocol)

- `hook/output.go` `HookEventName: "PreToolUse"` — CC protocol value (TENSION-012 / LB-010)
- `.claude/` directory name — CC expects it (SCAR-002)
- `PROVENANCE_MANIFEST.yaml` filename for default channel
- `KNOSSOS_CHANNEL` environment variable name

### Deprecated markers

- `RiteManifest.Skills` — `// Deprecated: use Legomena instead`
- `RiteManifest.Commands` — backward compat field
- `paths.AgentsDir()` — `// Deprecated: Use AgentsDirForChannel`
- `materialize.ResolveUserResources()` — `// Deprecated: Use ResolveUserResourcesForChannel`

### In-progress migrations

- **ADR-0032 PKG-010**: `GetAdapter()` CC default (harness-agnosticism initiative, parked)
- **perspective channel-awareness**: Hardcoded `.claude/`, unresolved
- **ADR-0032 AGENTS.md**: Third compilation target designed but not implemented
- **BehavioralContract.MaxTurns -> maxTurns**: Known gap, not wired

---

## Risk Zone Mapping

### RZ-001: materialize/ -- No Rollback on Partial Write Failure

**Location**: `internal/materialize/materialize.go:370-518`
The 10-step pipeline is not transactional. `prevalidateCLAUDEmd()` at step 0 mitigates by failing fast, but disk-full mid-write leaves partial state.

### RZ-002: config.KnossosHome() Test Cache Poisoning

**Location**: `internal/config/home.go:81-96`
Tests calling `KnossosHome()` before setting `KNOSSOS_HOME` bake in the developer's actual home directory. Can cause tests to pass locally, fail in CI.

### RZ-003: search/collectors.go Circular Layer Risk

**Location**: `internal/search/collectors.go:12`
`internal/search` imports `internal/cmd/explain`. If `cmd/explain` ever imports `search`, circular import results.

### RZ-004: Mena Namespace Collision Across Rites

Collisions between user-owned and rite mena entries produce Warnings but no hard fail. Warnings may be lost if callers don't log them.

### RZ-005: Lock Stale Threshold Is Time-Based (5 min), Not Process-Based

**Location**: `internal/lock/lock.go:32`
Long-running commands (>5 min) may have locks incorrectly classified as stale. No heartbeat mechanism.

### RZ-006: Org Scope Sync Silently Skips on Error

**Location**: `internal/materialize/materialize.go:578-589`
Org-scope sync error logs to `OrgScopeResult.Error` but doesn't abort overall sync. Misconfigured org silently produces partial sync.

### RZ-007: `userscope/sync.go` Fallback Channel Default

**Location**: `internal/materialize/userscope/sync.go` line 33
Default to `"claude"` when no channel configured. Gemini user files not synced, no error raised.

### RZ-008: `perspective/context.go` Channel Dir Hardcode

**Location**: `internal/perspective/context.go` line 60
Gemini channel perspective assembly resolves against `.claude/`. Produces silently incorrect data.

---

## Knowledge Gaps

1. **`internal/materialize/procession/`**: Procession mena rendering constraints not captured.
2. **`internal/materialize/compiler/gemini.go`**: Gemini compiler format constraints not documented.
3. **`internal/naxos/`**: Potential design constraints around debt detection unknown.
4. **`internal/lock/`**: Session locking constraints not verified in source.
5. **`internal/tribute/`**: Constraints not examined.
6. **`config/hooks.yaml` canonical structure**: Hook configuration format constraints not inspected.
7. **Worktree rite inheritance**: `inheritRiteFromMainWorktree()` failure modes not traced.

# Gap Analysis: Org Tier Opportunity Catalog and Value Capture

> Operationalized from both `docs/spikes/SPIKE-org-level-best-practices.md` (original spike, 2026-03-01) and `docs/spikes/SPIKE-org-tier-root-cause-analysis.md` (RCA spike, 2026-03-02). Every opportunity from both documents is cataloged, classified, and mapped to value chains and critical path. Line-level code references verified against current `main` (commit `804701a`).

**Date**: 2026-03-02
**Source Spikes**: `docs/spikes/SPIKE-org-level-best-practices.md`, `docs/spikes/SPIKE-org-tier-root-cause-analysis.md`
**Rite**: ecosystem
**Complexity**: SYSTEM (overall initiative); individual sprints are MODULE

---

## Root Cause

The source resolution chain in `internal/materialize/source/resolver.go:51-60` was designed for `{project, user, platform, embedded}` -- a single-developer hierarchy. There is no concept of "shared across projects for a team" anywhere in the codebase. The gap sits between user tier (line 108-118) and knossos tier (line 121-132): org-level rites, agents, mena, and conventions have no resolution path.

## Affected Systems

- `internal/materialize/source/` -- resolver and type constants
- `internal/config/` -- org identity resolution
- `internal/paths/` -- XDG directory functions
- `internal/materialize/` -- sync pipeline types and orchestration
- `internal/materialize/userscope/` -- pattern to mirror for new orgscope
- `internal/provenance/` -- ownership tracking
- `internal/cmd/` -- CLI command wiring

---

## 1. Complete Opportunity Registry

### 1.1 Foundation Items (Type: Foundation)

| ID | Name | Source | Files:Lines | Value | Effort | Risk | Dependencies |
|----|------|--------|-------------|-------|--------|------|-------------|
| QW-1 | Add `SourceOrg` type constant | RCA S4.1 | `internal/materialize/source/types.go:9-20` | Unlocks org tier in resolver dispatch; all downstream resolution, sync, and CLI items depend on this constant existing | XS (15 min) | None -- additive constant, zero behavioral change | None |
| QW-2 | Add `OrgDataDir()`, `OrgRitesDir()`, `OrgAgentsDir()`, `OrgMenaDir()` to paths | RCA S4.1 | `internal/paths/paths.go:260-268` (insert after `UserRitesDir`) | Provides canonical XDG-compliant path functions for all org-level file operations | XS (30 min) | None -- pure functions, no side effects | None |
| QW-3 | Add `ActiveOrg()` to config | RCA S4.1 | `internal/config/home.go:34+` (new function) | Provides `$KNOSSOS_ORG` env + `active-org` file lookup; all org-aware code needs this | XS (30 min) | Circular import risk: `home.go` cannot import `paths` (but `ActiveOrg` can inline XDG path resolution to avoid this -- mirror `resolveKnossosHome()` pattern at line 24-33) | None |
| QW-4 | Document existing hierarchy in `.know/conventions.md` | RCA S4.1 + Orig S9 #1 | `.know/conventions.md` | Clarity for developers and agents about the current 5-tier model before adding a 6th | S (1 hr) | None | None |

### 1.2 Resolution Items (Type: Resolution)

| ID | Name | Source | Files:Lines | Value | Effort | Risk | Dependencies |
|----|------|--------|-------------|-------|--------|------|-------------|
| CI-1 | Insert org tier in `ResolveRite()` | RCA S4.2 | `internal/materialize/source/resolver.go:119` (insert after user block ending at line 119, before knossos block at line 121) | Org rites become resolvable in the standard chain; `ari sync --rite X` finds org-provided rites | S (2 hr) | Cache key collision if `orgRitesDir` changes between invocations (see IMP-4) | QW-1, QW-2, QW-3 |
| CI-2 | Add org to `ListAvailableRites()` | RCA S4.2 | `internal/materialize/source/resolver.go:267-271` (insert in `sources` slice) | `ari rite list` shows org-provided rites; required for discoverability | XS (30 min) | None -- additive to existing source iteration | QW-1, QW-2 |
| CI-7 | Tests for 6-tier resolution chain | RCA S4.2 | `internal/materialize/source/source_test.go` (new test functions) | Regression safety for the entire resolution chain including org tier | S (2 hr) | Test isolation -- must mock org directory without affecting developer XDG paths | CI-1, CI-2 |

### 1.3 Pipeline Items (Type: Pipeline)

| ID | Name | Source | Files:Lines | Value | Effort | Risk | Dependencies |
|----|------|--------|-------------|-------|--------|------|-------------|
| CI-3 | Add `ScopeOrg` to sync pipeline types | RCA S4.2 | `internal/materialize/sync_types.go:11-15` (add to const block), `:18-24` (add to `IsValid`), `:40-50` (add `OrgName` field), `:53-56` (add `OrgResult` field) | `ari sync --scope org` becomes a valid invocation; Phase 1.5 insertion point exists in `Sync()` | S (2 hr) | Must add `OrgScopeResult` type and wire it into JSON output -- breaking change if external consumers parse `SyncResult` | CI-1 |
| CI-4 | Create `orgscope/sync.go` package | RCA S4.2 | `internal/materialize/orgscope/sync.go` (NEW, ~150 lines) | Actual sync logic: copies org agents/mena to `~/.claude/` with org provenance tracking; mirrors `userscope/sync.go` patterns | M (4 hr) | Must not conflict with user-scope provenance; needs clear ownership boundaries | CI-3, IMP-6 |
| FA-8 | Ship org sync pipeline (`ari sync --scope org`) | Orig S9 #8 | `internal/materialize/materialize.go:542` (insert Phase 1.5 between rite phase ending at line 541 and user phase starting at line 543) | End-to-end org sync: `ari sync` with active org flows through rite -> org -> user | M (4 hr) | Phase ordering matters -- org content must be available before user sync runs, so user can override org | CI-3, CI-4 |

### 1.4 Command Items (Type: Command)

| ID | Name | Source | Files:Lines | Value | Effort | Risk | Dependencies |
|----|------|--------|-------------|-------|--------|------|-------------|
| CI-5 | `ari org init {name}` command | RCA S4.2 + Orig S9 #4 | `internal/cmd/org/init.go` (NEW, ~80 lines) | Bootstrap org directory structure at XDG path; `ari org init autom8y` creates `~/.local/share/knossos/orgs/autom8y/` with `org.yaml`, `rites/`, `agents/`, `mena/` | S (3 hr) | Must validate org name (kebab-case, no path traversal) | QW-2, QW-3 |
| CI-6 | `ari org set/list` commands | RCA S4.2 | `internal/cmd/org/set.go` (NEW, ~30 lines), `internal/cmd/org/list.go` (NEW, ~40 lines) | `ari org set autom8y` persists active org; `ari org list` discovers all orgs at XDG path | S (2 hr) | `ari org set` must call `ClearCache()` on resolver (see IMP-4) | CI-5 |
| FA-2 | Validate `ari init` without KNOSSOS_HOME | Orig S9 #2 | `internal/cmd/initialize/init.go` (validate existing code paths) | Ensures external users can bootstrap without `$KNOSSOS_HOME` set; critical for Stage 2 distribution | XS (1 hr) | May reveal hidden `KnossosHome()` assumptions (see IMP-2) | None |

### 1.5 Improvement Items (Type: Improvement)

| ID | Name | Source | Files:Lines | Value | Effort | Risk | Dependencies |
|----|------|--------|-------------|-------|--------|------|-------------|
| IMP-1 | `RiteDir()` inconsistency with resolver | RCA S4.3 | `internal/paths/paths.go:196-206` | `RiteDir()` checks project then user, skipping knossos and embedded tiers. Any caller using `RiteDir()` instead of `SourceResolver.ResolveRite()` gets a different answer -- causes subtle bugs when rites exist only in knossos tier | S (2 hr) | Behavioral change -- callers relying on current 2-tier fallback may break | None (standalone fix) |
| IMP-2 | `KnossosHome()` hardcoded default | RCA S4.3 | `internal/config/home.go:31-32` | Defaults to `$HOME/Code/knossos` -- unusable for external developers who have no knossos checkout. Stage 2 distribution requires empty-by-default with `ari init` seeding it | S (1 hr) | Breaking change for current workflow if not gated behind feature flag or migration | FA-2 |
| IMP-3 | `parseExplicitSource()` missing org alias | RCA S4.3 | `internal/materialize/source/resolver.go:168-181` | Only supports `"knossos"` alias; should support `"org"` and `"org:{name}"` for explicit org source selection via `--source org:autom8y` | XS (30 min) | Must handle `org:` prefix parsing correctly; edge case when no active org is set | QW-3, CI-1 |
| IMP-4 | Resolver cache not org-aware | RCA S4.3 | `internal/materialize/source/resolver.go:24-26` (cache map), `:62-70` (cache check), `:154-158` (cache store) | Cache key is rite name only. If user switches org, cached resolution still returns the old org's rite. Stale cache causes wrong rite to materialize silently | S (1 hr) | Must clear cache on org switch OR include org in cache key -- latter is cleaner but requires `orgRitesDir` in key | QW-3, CI-1 |
| IMP-5 | User sync hardcodes sole source | RCA S4.3 | `internal/materialize/userscope/sync.go` (references `config.KnossosHome()` throughout) | User sync pulls exclusively from `$KNOSSOS_HOME`. Org sync needs its own source chain pulling from `OrgDataDir()`. This is not a bug in user sync itself but means org sync cannot reuse `userscope/sync.go` as-is -- must create parallel `orgscope/` | M (4 hr) | Largest effort item; must mirror user sync patterns without duplicating 1530 lines | CI-3, CI-4 |
| IMP-6 | No org-level provenance tracking | RCA S4.3 | `internal/provenance/provenance.go:22+`, `internal/paths/paths.go:312-316` | Only `PROVENANCE_MANIFEST.yaml` (rite) and `USER_PROVENANCE_MANIFEST.yaml` (user) exist. Org-synced files need `ORG_PROVENANCE_MANIFEST.yaml` for ownership tracking and collision detection | S (2 hr) | Must integrate with existing provenance merge logic; org-owned files must be distinguishable from user-owned files during divergence detection | CI-4 |

### 1.6 Documentation Items (Type: Documentation)

| ID | Name | Source | Files:Lines | Value | Effort | Risk | Dependencies |
|----|------|--------|-------------|-------|--------|------|-------------|
| FA-1 | Document current hierarchy in `.know/conventions.md` | Orig S9 #1 | `.know/conventions.md` | Same scope as QW-4; merged in this catalog | XS (1 hr) | None | None |
| FA-3 | Write ADR for org tier addition | Orig S9 #3 + RCA S9 #1 | `docs/decisions/ADR-00XX-org-tier.md` (NEW) | Permanent design record; gates all implementation work | S (2 hr) | Decision must be accepted before implementation begins | Both spikes accepted |
| FA-6 | Create org template repo | Orig S9 #6 | `knossos-org-template` (NEW repo) | Canonical starting point for `ari org init --from`; reduces org bootstrap friction | M (4 hr) | Premature if org sync pipeline is not yet functional | CI-5, FA-8 |
| FA-7 | Document CC managed-settings.json recommendations | Orig S9 #7 | `docs/` or `.know/` | Prevents teams from reinventing CC's native org controls; clear boundary documentation | S (2 hr) | None | None |
| FA-RCA7 | Update distribution roadmap with org tier | RCA S9 #7 | `docs/strategy/ROADMAP-distribution-readiness.md` | Roadmap reflects the new org tier as a Stage 2 prerequisite | XS (30 min) | None | FA-3 |
| FA-RCA8 | Document `KnossosHome()` default for Stage 2 | RCA S9 #8 | `docs/` or `.know/` | Clarifies migration path for external users who have no `$HOME/Code/knossos` | XS (30 min) | None | IMP-2 |

---

## 2. Critical Path Analysis

### 2.1 Sequential Dependencies (Must Execute In Order)

```
Critical Path A: Foundation --> Resolution --> Sync Pipeline --> Commands
Duration: ~18 hours across 4 sprints

  QW-1 (SourceOrg constant, 15 min)
    |
  QW-2 (Org path functions, 30 min)
    |
  QW-3 (ActiveOrg config, 30 min)
    |
    +---> CI-1 (ResolveRite org insertion, 2 hr)
    |       |
    |     CI-2 (ListAvailableRites org, 30 min)
    |       |
    |     CI-7 (6-tier resolution tests, 2 hr)
    |       |
    +---> CI-3 (ScopeOrg sync types, 2 hr)
            |
          IMP-6 (Org provenance manifest, 2 hr)
            |
          CI-4 (orgscope/sync.go, 4 hr)
            |
          FA-8 (Phase 1.5 in Sync(), 4 hr)
            |
          CI-5 (ari org init, 3 hr)
            |
          CI-6 (ari org set/list, 2 hr)
```

### 2.2 Parallelizable Work

Items with NO dependency on the critical path:

| Item | Can Start | Why Independent |
|------|-----------|-----------------|
| QW-4 / FA-1 | Immediately | Documentation of existing hierarchy |
| FA-2 | Immediately | Validates existing `ari init` code |
| FA-3 (ADR) | Immediately | Design document, gates implementation |
| FA-7 | Immediately | CC best practices documentation |
| IMP-1 | Immediately | `RiteDir()` fix is standalone |
| IMP-2 | Immediately | `KnossosHome()` default is standalone |
| FA-RCA7 | After FA-3 | Roadmap update |
| FA-RCA8 | After IMP-2 | Documents the default change |

Items to bundle INTO their critical-path sprint:

| Item | Sprint | Why Coupled |
|------|--------|-------------|
| IMP-3 (org alias) | Sprint 2 (Resolution) | Modifying the same `resolver.go` -- avoid merge conflicts |
| IMP-4 (cache org-awareness) | Sprint 2 (Resolution) | Modifying cache logic in the same `resolver.go` |
| IMP-5 (user sync source) | Sprint 3 (Sync Pipeline) | Understanding org sync scope requires reading user sync |

### 2.3 Recommended Execution Order

| Order | Items | Sprint | Hours | Parallelizable Side Work |
|-------|-------|--------|-------|--------------------------|
| 1 | FA-3 (ADR) | Pre-sprint | 2 | QW-4, FA-2, FA-7, IMP-1, IMP-2 |
| 2 | QW-1, QW-2, QW-3 | Sprint 1 | 1.25 | FA-RCA7, FA-RCA8 |
| 3 | CI-1, CI-2, CI-7, IMP-3, IMP-4 | Sprint 2 | 6 | -- |
| 4 | CI-3, CI-4, IMP-5, IMP-6, FA-8 | Sprint 3 | 12 | -- |
| 5 | CI-5, CI-6 | Sprint 4 | 5 | FA-6 (org template repo) |

**Total critical path**: ~26.25 hours
**Total with parallel work**: ~34 hours

---

## 3. Value Chain Maps

### Chain 1: Org Rite Resolution (core infrastructure)

```
QW-1 (SourceOrg type)
  |-- Enables: CI-1 (resolver insertion)
       |-- Enables: CI-2 (rite listing)
       |         |-- Delivers: `ari rite list` shows org rites
       |-- Enables: CI-7 (resolution tests)
       |         |-- Delivers: regression safety for 6-tier chain
       |-- Enables: IMP-3 (--source org:name)
                  |-- Delivers: explicit org rite selection
```

**End value**: Developers can reference org-provided rites without copying them into every project.

### Chain 2: Org Sync Pipeline (content distribution)

```
QW-2 (OrgDataDir/OrgRitesDir)
  |-- Enables: CI-3 (ScopeOrg sync types)
       |-- Enables: CI-4 (orgscope/sync.go)
       |    |-- Requires: IMP-6 (org provenance)
       |    |-- Delivers: Org agents/mena synced to ~/.claude/
       |-- Enables: FA-8 (Phase 1.5 in Sync)
            |-- Delivers: `ari sync` with active org flows through rite -> org -> user
```

**End value**: `ari sync` distributes org-level agents, skills, and commands to all developers without manual copying.

### Chain 3: Org Lifecycle Management (developer UX)

```
QW-3 (ActiveOrg config)
  |-- Enables: CI-5 (ari org init)
       |-- Enables: CI-6 (ari org set/list)
       |    |-- Requires: IMP-4 (cache invalidation on org switch)
       |    |-- Delivers: `ari org set autom8y` + `ari org list`
       |-- Enables: FA-6 (org template repo)
            |-- Delivers: `ari org init autom8y --from git@...`
```

**End value**: Developers can create, switch between, and list organizations with a familiar CLI pattern (mirrors `ari rite` UX).

### Chain 4: Code Quality (standalone improvements)

```
IMP-1 (RiteDir consistency)
  |-- Delivers: paths.Resolver.RiteDir() checks all tiers, not just project+user
  |-- Prevents: silent resolution failures when rites exist only in knossos tier

IMP-2 (KnossosHome default)
  |-- Delivers: external users get empty default instead of $HOME/Code/knossos
  |-- Prevents: confusing "directory not found" errors on first `ari sync`
  |-- Enables: FA-RCA8 (documentation of new default)
```

**End value**: Reduces support burden for Stage 2 distribution by eliminating two categories of "works on my machine" bugs.

---

## 4. Gap Analysis Between Spikes

### 4.1 Items in Original Spike NOT Covered by RCA

| Original Item | Status | Analysis |
|---------------|--------|----------|
| Orig S4.2 (org.yaml manifest schema) | **Missing from RCA** | The original spike defines a detailed `org.yaml` schema (schema_version, default_rite, shared resources, CC hint). The RCA assumes org directories exist but never specifies the manifest format. **Action needed**: CI-5 (`ari org init`) must define the `org.yaml` schema -- route to Context Architect. |
| Orig S4.4 (bootstrap flow with `--from git@...`) | **Missing from RCA** | The original spike specifies `ari org init autom8y --from git@github.com:...` for cloning remote org configs. The RCA's CI-5 only covers local directory creation. **Action needed**: Remote clone capability is Stage 3; FA-6 (org template repo) partially addresses this. |
| Orig S4.5 (CC managed-settings.json integration) | **Correctly excluded by RCA** | RCA Section 8 explicitly lists this as "What NOT to Build." Both spikes agree: delegate to CC. |
| Orig S6.2 (Approach A implementation steps 4-5) | **Partially covered** | Step 4 (org template repo) = FA-6. Step 5 (CC managed-settings docs) = FA-7. Both present but lower priority in RCA. |
| Orig S8.1-8.4 (CC-native best practices for today) | **Not actionable items** | These are documentation of what teams can do NOW without code changes. Not opportunities per se, but FA-7 should reference them. |

### 4.2 Items in RCA NOT Covered by Original Spike

| RCA Item | Status | Analysis |
|----------|--------|----------|
| IMP-1 (RiteDir consistency) | **New finding** | Code-level bug found during root cause tracing at `internal/paths/paths.go:196-206`. Original spike was architecture-level, could not have identified this. |
| IMP-3 (parseExplicitSource org alias) | **New finding** | Follows from code-level analysis of `resolver.go:168-181`. |
| IMP-4 (cache invalidation) | **New finding** | Cache at `resolver.go:24-26` was not discussed in original spike. |
| IMP-5 (user sync sole source) | **New finding** | Identified by reading `userscope/sync.go`; original spike estimated effort without reading the code. |
| IMP-6 (org provenance) | **New finding** | Original spike mentions provenance conceptually but does not identify the specific gap in `provenance.go:22+`. |

### 4.3 Contradictions Between Spikes

| Topic | Original Spike | RCA | Resolution |
|-------|---------------|-----|------------|
| **Effort estimate** | 40-60 hours (Orig S6.2) | 18 hours (RCA S7) | RCA is correct for core infrastructure. Original includes enterprise features (template repo, CC config docs) that add ~16 hours. True total with all items: ~34 hours. |
| **Resolution order** | project > org > user > knossos > embedded (Orig S4.1) | project > user > org > knossos > embedded (RCA S6 constraint 2) | **RCA is correct.** User must override org (personal preferences override team defaults). Original spike's ordering was aspirational, RCA's was code-informed. This is the most consequential difference between the two spikes and must be resolved before ADR is written. |
| **Org directory location** | `$XDG_DATA_HOME/knossos/orgs/{org}/` (Orig S4.2) | Same, via `paths.OrgDataDir()` (RCA S2.2) | Consistent. No contradiction. |
| **QW-4 vs FA-1** | Original's Follow-Up #1 covers same ground | QW-4 in RCA (document hierarchy) | Effectively duplicates. Merged in this catalog as QW-4/FA-1. |

### 4.4 Assumptions Needing Validation

| Assumption | Source | Validation Method |
|------------|--------|-------------------|
| `ActiveOrg()` should read `$KNOSSOS_ORG` then fall back to file | RCA S2.2 | Confirm this mirrors `ActiveRite()` pattern (currently `ACTIVE_RITE` is a file in `.claude/`, not XDG); check if env var override is sufficient for CI/CD use cases |
| Org rites should NOT carry templates (use embedded) | RCA S2.1 code snippet | Validate that org rites can function without custom CLAUDE.md templates; if orgs need custom inscription templates, the `SourceOrg` case in `checkSource()` at resolver.go:226-249 needs a real `templatesDir` path |
| One active org at a time (parallels one active rite) | RCA S8 | Validate with multi-team developers -- does anyone need simultaneous org resolution? If yes, this is a scope expansion to MIGRATION complexity |
| `config.ActiveOrg()` can inline XDG path resolution | RCA S2.2 | Verified: `config` does NOT import `paths` and `paths` does NOT import `config` (checked import lists). Safe to add `ActiveOrg()` to config using inline XDG resolution, mirroring the `resolveKnossosHome()` pattern at `home.go:24-33` |

---

## 5. Complexity Classification

Using the 10x-workflow complexity scale:

| Classification | Items | Rationale |
|----------------|-------|-----------|
| **PATCH** | QW-1, QW-2, QW-3, QW-4/FA-1, CI-2, IMP-3, FA-2, FA-7, FA-RCA7, FA-RCA8 | Single-file changes, under 30 lines, no cross-package impact |
| **MODULE** | CI-1, CI-3, CI-7, IMP-1, IMP-2, IMP-4, IMP-6, FA-3 | Multi-file changes within one package, or new test files, behavioral changes requiring careful testing |
| **SYSTEM** | CI-4, CI-5, CI-6, FA-8, IMP-5 | New packages, new CLI command groups, cross-package wiring (sync pipeline -> org scope -> provenance -> commands) |
| **MIGRATION** | FA-6 | New external repository; requires distribution infrastructure decisions |

**Overall initiative complexity: SYSTEM** -- requires coordinated changes across 5 packages, a new sub-package, a new CLI command group, and a new provenance manifest. However, each sprint is individually MODULE-complexity, making phased delivery feasible.

---

## 6. Test Satellite Matrix

| Satellite Config | Purpose | Validates |
|------------------|---------|-----------|
| No org (empty `ActiveOrg()`) | Baseline regression | Existing 5-tier resolution unchanged when no org is configured |
| Org with rites only | Minimal org | Org rites resolvable between user and knossos tiers |
| Org with agents + mena | Full org content | Org sync distributes agents and mena to `~/.claude/` |
| Org + user collision | Override semantics | User-tier agents/mena override same-named org content |
| Org + project satellite | Full 6-tier chain | Project overrides user overrides org overrides knossos |
| Multi-org (switch) | Org lifecycle | `ari org set` clears cache and subsequent sync uses new org |
| Org without `$KNOSSOS_HOME` | Stage 2 distribution | External developer with org but no platform checkout |

---

## 7. Success Criteria

- `ari sync` exits 0 with an active org configured, syncing org agents and mena to `~/.claude/`
- `ari sync` exits 0 with NO org configured (regression: existing behavior unchanged)
- `ari rite list` includes org-provided rites with `source: org` annotation
- `ari org init autom8y` creates well-formed directory at `$XDG_DATA_HOME/knossos/orgs/autom8y/`
- `ari org set autom8y && ari sync` resolves rites from org tier before knossos tier
- User-tier agents override same-named org-tier agents (precedence: project > user > org > knossos > embedded)
- `ORG_PROVENANCE_MANIFEST.yaml` tracks org-synced files distinctly from user-synced files
- 6-tier resolution tests pass in CI with no flakiness

---

## 8. Summary Metrics

| Metric | Value |
|--------|-------|
| Total opportunities cataloged | 26 |
| Foundation items | 4 |
| Resolution items | 3 |
| Pipeline items | 3 |
| Command items | 3 |
| Improvement items | 6 |
| Documentation items | 7 |
| Critical path duration | ~26.25 hours |
| Parallelizable work | ~8 hours |
| Total estimated effort | ~34 hours |
| New files to create | 5 (orgscope/sync.go, cmd/org/init.go, cmd/org/list.go, cmd/org/set.go, ADR) |
| Existing files to modify | 8 (types.go, resolver.go, home.go, paths.go, sync_types.go, materialize.go, cmd/sync/sync.go, cmd/initialize/init.go) |
| Items requiring ADR approval first | 15 (all CI-* and FA-8 items) |
| Items executable immediately | 11 (all QW-*, IMP-1, IMP-2, FA-1/QW-4, FA-2, FA-3, FA-7) |

---

## Handoff

This Gap Analysis is ready for Context Architect to design:
1. The `org.yaml` manifest schema (identified as missing from RCA -- see Section 4.1)
2. The resolution order decision (project > user > org > knossos > embedded vs. original spike's project > org > user -- see Section 4.3)
3. The org scope sync architecture (whether to extract shared logic from `userscope/sync.go` or create a parallel implementation -- see IMP-5)
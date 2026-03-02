# SPIKE: Org Tier Root Cause Analysis & Remediation Surface Area

> Deep spike analyzing the root cause surface area for the missing org tier in Knossos's source resolution chain, with remediation paths and improvement opportunities grounded in actual code.

**Date**: 2026-03-02
**Extends**: `docs/spikes/SPIKE-org-level-best-practices.md`
**Session**: `session-20260302-115457-2233b3ef`
**Rite**: ecosystem

---

## Executive Summary

The missing org tier is not a bug — it's a structural gap from Knossos's origin as a single-developer framework. The codebase was built around two resolution axes (project vs. platform) and later added a user tier. An org tier requires changes across **5 packages, 8 files to modify, 4 new files to create**, but the architecture is well-prepared: every modification follows existing patterns.

**Root Cause**: The source resolution chain was designed for `{project, user, platform, embedded}` — a single-developer hierarchy. There is no concept of "shared across projects for a team" anywhere in the codebase.

**Surface Area**: 3 core systems need modification:
1. **Source Resolution** (1 file, ~30 lines) — insert org tier between user and knossos
2. **Path Resolution** (2 files, ~40 lines) — add org directory functions
3. **Sync Pipeline** (3 files + 1 new package, ~200 lines) — add org scope to sync

---

## 1. Root Cause Analysis

### 1.1 Why There Is No Org Tier

The resolution chain in `internal/materialize/source/resolver.go:51-60` is:

```
1. Explicit (--source flag)
2. Project (.knossos/rites/{rite}/)
3. User (~/.local/share/knossos/rites/{rite}/)
4. Knossos ($KNOSSOS_HOME/rites/{rite}/)
5. Embedded (compiled into binary)
```

This was designed for **the Knossos developer** — someone who has `$KNOSSOS_HOME` pointing at their checkout and uses project-level satellites for overrides. The user tier was a later addition for distribution (Stage 2 readiness). There was never a use case for "multiple projects sharing rites from a team-owned directory."

### 1.2 The Actual Gap

The gap is **between tiers 3 and 4**. Today:
- **User tier** = "my personal rites" → `~/.local/share/knossos/rites/`
- **Knossos tier** = "the platform's rites" → `$KNOSSOS_HOME/rites/`

What's missing:
- **Org tier** = "my team's shared rites" → `~/.local/share/knossos/orgs/{org}/rites/`

This means:
- Teams cannot share custom rites across projects without each project embedding them as satellites
- There's no way to distribute org-specific agents, mena, or conventions without `$KNOSSOS_HOME` pointing to a shared checkout
- The user sync pipeline (`userscope/sync.go`) syncs from `$KNOSSOS_HOME` only — there's no org-scoped source

### 1.3 Why It Matters Now

Per `ROADMAP-distribution-readiness.md`, Stage 2 (Trusted External) is imminent. At that point:
- External teams will use `ari init` to bootstrap projects
- They will want shared rites across their org's projects
- Without an org tier, they must either: (a) copy rites into every project, or (b) point `$KNOSSOS_HOME` at a shared directory (which conflates org content with platform content)

---

## 2. Surface Area Map

### 2.1 Source Resolution (`internal/materialize/source/`)

| File | Lines | Changes Needed | Effort |
|------|-------|---------------|--------|
| `types.go` | 42 | Add `SourceOrg SourceType = "org"` | ~5 lines |
| `resolver.go` | 345 | Add `orgRitesDir` field, insert tier 3.5 in `ResolveRite()`, add org to `ListAvailableRites()` | ~30 lines |
| `source_test.go` | ~200 | Add org tier resolution tests (6-tier chain) | ~40 lines |

**Specific changes to `resolver.go`:**

```go
// SourceResolver — add field:
type SourceResolver struct {
    // ... existing fields ...
    orgRitesDir string  // NEW: org-level rites directory
}

// NewSourceResolver — add parameter or detect from config:
func NewSourceResolver(projectRoot string) *SourceResolver {
    return &SourceResolver{
        // ... existing ...
        orgRitesDir: paths.OrgRitesDir(config.ActiveOrg()),  // NEW
    }
}

// ResolveRite — insert between user (tier 3) and knossos (tier 4):
// Line ~119, after the "3. User rites" block:
//
// 3.5. Org rites
if result == nil && r.orgRitesDir != "" {
    source := RiteSource{
        Type:        SourceOrg,
        Path:        r.orgRitesDir,
        Description: "org-level rites",
    }
    if res, err := r.checkSource(riteName, source); err == nil {
        result = res
    } else {
        checkedPaths = append(checkedPaths, source.Path)
    }
}
```

`checkSource()` (line 216) is **already source-agnostic** — it just checks for `manifest.yaml` at `{source.Path}/{riteName}/`. No changes needed there, but the `switch source.Type` block for templates (line 227-249) needs a new case:

```go
case SourceOrg:
    templatesDir = "" // Org rites don't carry templates (use embedded)
```

`ListAvailableRites()` (line 262-316) needs org added to the sources slice:

```go
sources := []RiteSource{
    {Type: SourceProject, Path: r.projectRitesDir},
    {Type: SourceUser, Path: r.userRitesDir},
    {Type: SourceOrg, Path: r.orgRitesDir},      // NEW
    {Type: SourceKnossos, Path: filepath.Join(r.knossosHome, "rites")},
}
```

### 2.2 Path Resolution (`internal/paths/`, `internal/config/`)

| File | Lines | Changes Needed | Effort |
|------|-------|---------------|--------|
| `config/home.go` | 56 | Add `ActiveOrg()`, `OrgHome()` | ~25 lines |
| `paths/paths.go` | 333 | Add `OrgDataDir()`, `OrgRitesDir()`, `OrgAgentsDir()` | ~20 lines |

**New functions in `config/home.go`:**

```go
// ActiveOrg returns the currently active organization name.
// Reads from $KNOSSOS_ORG env var, then falls back to
// $XDG_CONFIG_HOME/knossos/active-org file.
func ActiveOrg() string {
    if org := os.Getenv("KNOSSOS_ORG"); org != "" {
        return org
    }
    data, err := os.ReadFile(filepath.Join(paths.ConfigDir(), "active-org"))
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(data))
}
```

**New functions in `paths/paths.go`:**

```go
// OrgDataDir returns the data directory for a named org.
// Location: $XDG_DATA_HOME/knossos/orgs/{orgName}/
func OrgDataDir(orgName string) string {
    return filepath.Join(DataDir(), "orgs", orgName)
}

// OrgRitesDir returns the org-level rites directory.
func OrgRitesDir(orgName string) string {
    if orgName == "" {
        return ""
    }
    return filepath.Join(OrgDataDir(orgName), "rites")
}

// OrgAgentsDir returns the org-level agents directory.
func OrgAgentsDir(orgName string) string {
    return filepath.Join(OrgDataDir(orgName), "agents")
}

// OrgMenaDir returns the org-level mena directory.
func OrgMenaDir(orgName string) string {
    return filepath.Join(OrgDataDir(orgName), "mena")
}
```

### 2.3 Sync Pipeline (`internal/materialize/`)

| File | Lines | Changes Needed | Effort |
|------|-------|---------------|--------|
| `sync_types.go` | 85 | Add `ScopeOrg`, `OrgScopeResult` | ~20 lines |
| `materialize.go` | 1000+ | Add org scope phase in `Sync()` | ~30 lines |
| `orgscope/sync.go` | NEW | Org-scope sync (mirror `userscope/sync.go`) | ~150 lines |

**Changes to `sync_types.go`:**

```go
const (
    ScopeAll  SyncScope = "all"
    ScopeRite SyncScope = "rite"
    ScopeOrg  SyncScope = "org"   // NEW
    ScopeUser SyncScope = "user"
)

func (s SyncScope) IsValid() bool {
    switch s {
    case ScopeAll, ScopeRite, ScopeOrg, ScopeUser:  // Add ScopeOrg
        return true
    }
}

type SyncOptions struct {
    // ... existing ...
    OrgName string  // NEW: org name for org-scope sync
}

type SyncResult struct {
    RiteResult *RiteScopeResult `json:"rite,omitempty"`
    OrgResult  *OrgScopeResult  `json:"org,omitempty"`   // NEW
    UserResult *UserScopeResult `json:"user,omitempty"`
}

type OrgScopeResult struct {
    Status   string `json:"status"`
    Error    string `json:"error,omitempty"`
    OrgName  string `json:"org_name"`
    Source   string `json:"source,omitempty"`
}
```

**Changes to `materialize.go` `Sync()` (line 506):**

Insert between Phase 1 (rite) and Phase 2 (user):

```go
// Phase 1.5: Org scope
if opts.Scope == ScopeAll || opts.Scope == ScopeOrg {
    orgResult, err := m.syncOrgScope(opts)
    if err != nil {
        if opts.Scope == ScopeOrg {
            return nil, err
        }
        result.OrgResult = &OrgScopeResult{
            Status: "skipped",
            Error:  err.Error(),
        }
    } else {
        result.OrgResult = orgResult
    }
}
```

### 2.4 New CLI Commands (`internal/cmd/`)

| File | Status | Purpose | Effort |
|------|--------|---------|--------|
| `cmd/org/init.go` | NEW | `ari org init {name}` — create org directory | ~80 lines |
| `cmd/org/list.go` | NEW | `ari org list` — list available orgs | ~40 lines |
| `cmd/org/set.go` | NEW | `ari org set {name}` — set active org | ~30 lines |
| `cmd/sync/sync.go` | MODIFY | Add `--org` flag | ~10 lines |
| `cmd/initialize/init.go` | MODIFY | Add `--org` flag | ~15 lines |

---

## 3. Dependency Graph

```
Phase 1: Foundation (no dependencies)
├── paths/paths.go: OrgDataDir(), OrgRitesDir(), OrgAgentsDir(), OrgMenaDir()
├── config/home.go: ActiveOrg(), OrgHome() (if needed)
└── source/types.go: SourceOrg constant

Phase 2: Resolution (depends on Phase 1)
├── source/resolver.go: orgRitesDir field, tier 3.5 insertion
├── source/resolver.go: ListAvailableRites() org inclusion
└── source/resolver.go: checkSource() SourceOrg template case

Phase 3: Sync (depends on Phase 2)
├── sync_types.go: ScopeOrg, OrgScopeResult, OrgName field
├── materialize.go: Phase 1.5 in Sync()
└── orgscope/sync.go: SyncOrgScope() (new package)

Phase 4: Commands (depends on Phase 3)
├── cmd/org/init.go: ari org init
├── cmd/org/list.go: ari org list
├── cmd/org/set.go: ari org set
└── cmd/sync/sync.go: --org flag
```

---

## 4. Remediation Opportunities

### 4.1 Quick Wins (Stage 1 — Internal, Now)

| # | Opportunity | Files | Effort | Impact |
|---|------------|-------|--------|--------|
| QW-1 | Add `SourceOrg` type constant | `source/types.go` | 5 min | Unlocks everything |
| QW-2 | Add `OrgDataDir()` and `OrgRitesDir()` to paths | `paths/paths.go` | 15 min | Foundation |
| QW-3 | Add `ActiveOrg()` to config | `config/home.go` | 15 min | Foundation |
| QW-4 | Document existing hierarchy in `.know/` | `.know/conventions.md` | 1 hour | Clarity |

### 4.2 Core Implementation (Stage 2 — Trusted External)

| # | Opportunity | Files | Effort | Impact |
|---|------------|-------|--------|--------|
| CI-1 | Insert org tier in `ResolveRite()` | `source/resolver.go` | 2 hours | Org rites resolvable |
| CI-2 | Add org to `ListAvailableRites()` | `source/resolver.go` | 30 min | Org rites discoverable |
| CI-3 | Add `ScopeOrg` to sync pipeline | `sync_types.go`, `materialize.go` | 2 hours | Org sync possible |
| CI-4 | Create `orgscope/sync.go` | New package | 4 hours | Org content synced |
| CI-5 | `ari org init` command | `cmd/org/init.go` | 3 hours | Org bootstrap |
| CI-6 | `ari org set/list` commands | `cmd/org/` | 2 hours | Org management |
| CI-7 | Tests for 6-tier resolution | `source/source_test.go` | 2 hours | Confidence |

### 4.3 Improvements Beyond Org Tier

| # | Opportunity | Root Cause | Remediation |
|---|------------|-----------|-------------|
| IMP-1 | `RiteDir()` in paths.go (line 198-206) only checks project then user — skips knossos and embedded | Inconsistent with source resolver's 5-tier chain | Should delegate to `SourceResolver` or at minimum check org+knossos |
| IMP-2 | `KnossosHome()` defaults to `$HOME/Code/knossos` (line 31-32) — baked-in developer assumption | Internal-only origin | Stage 2 should make this empty-by-default with `ari init` seeding it |
| IMP-3 | `parseExplicitSource()` only supports "knossos" alias | No org alias | Add "org" and "org:{name}" aliases |
| IMP-4 | `SourceResolver` caches by rite name only — no invalidation on org switch | No org concept when cache was designed | Add org-awareness to cache key or clear on org change |
| IMP-5 | User sync hardcodes `config.KnossosHome()` as sole source | Single-source design | Org sync needs its own source chain |
| IMP-6 | No provenance tracking at org level | Only project and user provenance manifests exist | Add `OrgProvenanceManifest()` and collision detection at org level |

---

## 5. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| **Org/user collision** — same agent name in org and user scope | High | Medium | Provenance manifest tracks ownership; org has lower priority than user |
| **Multi-org confusion** — developer in two orgs, wrong one active | Medium | High | `ari org set` persists to config; `ari sync` shows active org in output |
| **Template resolution** — org rites may not have templates | Low | Medium | `SourceOrg` case in `checkSource()` returns empty `templatesDir` |
| **Cache invalidation** — org switch doesn't clear resolver cache | Medium | Medium | Call `ClearCache()` in `ari org set` |
| **Worktree inheritance** — linked worktrees should inherit org | Low | Low | Org is config-level (XDG), not project-level; inherits naturally |

---

## 6. Architectural Constraints

From ADR-0025 (Mena Scope) and ADR-0021 (Two-Axis Context Model):

1. **Mena scope** must extend to org: `scope: org` should be a valid mena scope directive
2. **Resolution precedence** must be strict: project > user > org > knossos > embedded (no ambiguity)
3. **Provenance** must track org ownership: org-synced files must be distinguishable from user-synced files
4. **CC alignment**: Org content that ends up in `~/.claude/` (user scope in CC terms) must not conflict with CC's managed-settings.json policies

---

## 7. Recommended Sequencing

### Sprint 1: Foundation (~4 hours)
- `SourceOrg` constant in `types.go`
- `OrgDataDir()`, `OrgRitesDir()` in `paths.go`
- `ActiveOrg()` in `home.go`
- Unit tests for all new functions

### Sprint 2: Resolution (~4 hours)
- Insert org tier in `ResolveRite()`
- Add org to `ListAvailableRites()`
- Add `SourceOrg` case to `checkSource()`
- 6-tier resolution integration tests

### Sprint 3: Sync Pipeline (~6 hours)
- `ScopeOrg` in sync types
- Phase 1.5 in `Sync()`
- `orgscope/sync.go` (minimal: agents + mena)
- Org provenance manifest

### Sprint 4: Commands (~4 hours)
- `ari org init {name}`
- `ari org list`
- `ari org set {name}`
- `--org` flag on `ari sync` and `ari init`

**Total estimated effort**: 18 hours (vs. 40-60 in original spike — reduced by scoping to core infrastructure without enterprise features)

---

## 8. What NOT to Build

| Feature | Why Not |
|---------|---------|
| Org-level `managed-settings.json` generation | CC's domain — delegate entirely |
| Remote org registry / marketplace | Stage 3 — premature complexity |
| Org-level CLAUDE.md (separate from user) | CC has no concept of this; content goes to `~/.claude/CLAUDE.md` via user sync |
| Org-level hook enforcement | CC's `allowManagedHooksOnly` handles this |
| Multi-org simultaneous resolution | One active org at a time (parallels one active rite) |

---

## 9. Follow-Up Actions

| # | Action | Priority | Depends On |
|---|--------|----------|-----------|
| 1 | Write ADR for org tier addition | P1 | This spike accepted |
| 2 | Fix `RiteDir()` inconsistency (IMP-1) | P1 | None |
| 3 | Implement Sprint 1 (Foundation) | P2 | ADR accepted |
| 4 | Implement Sprint 2 (Resolution) | P2 | Sprint 1 |
| 5 | Implement Sprint 3 (Sync Pipeline) | P2 | Sprint 2 |
| 6 | Implement Sprint 4 (Commands) | P2 | Sprint 3 |
| 7 | Update distribution roadmap with org tier | P2 | ADR accepted |
| 8 | Document `KnossosHome()` default for Stage 2 (IMP-2) | P3 | None |

---

## 10. Key Files Reference

| File | Lines | Role in Org Tier |
|------|-------|-----------------|
| `internal/materialize/source/types.go` | 42 | Add `SourceOrg` constant |
| `internal/materialize/source/resolver.go` | 345 | Insert org tier in resolution chain |
| `internal/config/home.go` | 56 | Add `ActiveOrg()`, org home resolution |
| `internal/paths/paths.go` | 333 | Add org directory path functions |
| `internal/materialize/sync_types.go` | 85 | Add `ScopeOrg`, `OrgScopeResult` |
| `internal/materialize/materialize.go` | 1000+ | Add org phase in `Sync()` |
| `internal/materialize/userscope/sync.go` | 1530 | Pattern to mirror for org sync |
| `internal/cmd/initialize/init.go` | 364 | Add `--org` flag |

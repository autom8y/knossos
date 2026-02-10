# TDD: ADR-0026 Phase 4b -- Pipeline Absorption

> Technical Design Document for absorbing usersync into materialize, unifying the CLI, and deleting dead code.

**Status**: Approved
**Author**: Context Architect
**Date**: 2026-02-09
**ADR**: docs/decisions/ADR-0026-unified-provenance.md
**Scope**: Phase 4b (pipeline absorption). Builds on Phase 4a (schema unification, complete).
**Predecessor**: TDD-provenance-phase4a.md (Phase 4a, complete)

---

## Decision Register

Decisions locked from stakeholder session on 2026-02-09. These are final; do not re-litigate.

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| D1 | Absorption model | Full code merge. Usersync package deleted. | Single pipeline eliminates type duplication, reduces cognitive overhead, and enables unified provenance in one pass |
| D2 | Extra stages | Scope-gated. Stages conditionally execute per scope. | Avoids duplicating the Materializer struct. Rite stages skip when scope=user; user stages skip when scope=rite |
| D3 | CLI entry point | `ari sync [flags]`. Materialize/user subcommands removed. | Single command reduces surface area. Flags replace subcommand routing |
| D4 | Per-resource granularity | `--resource=agents\|mena\|hooks` optional filter flag. | Allows targeted sync without re-syncing everything. Useful for debugging |
| D5 | Default scope | `all` (rite + user). | Most common operation is full sync. Explicit `--scope=rite` or `--scope=user` for targeted use |
| D6 | Remote sync commands | Remove (status/pull/push/diff/resolve/history/reset). | Dead code. Remote sync was designed but never used in production. State tracking via provenance manifest made it obsolete |
| D7 | Execution order | Rite first, then user. | User scope collision checker needs rite provenance manifest. Running rite first ensures manifest is fresh |
| D8 | Force flag | Removed. Sync is always idempotent. | With provenance-based divergence detection, --force is unnecessary. --overwrite-diverged replaces it for the specific case of user-modified files |
| D9 | Minimal mode | Eliminated. | Cross-cutting mode is now `--scope=rite` with no ACTIVE_RITE, which triggers MaterializeMinimal internally |
| D10 | Recovery mode | `--recover` promoted to both scopes. | Recovery adopts untracked files into the provenance manifest for both rite and user scopes |
| D11 | Diverged files | `--overwrite-diverged` flag. Default: skip diverged. | Explicit flag name communicates the destructive nature better than --force |
| D12 | Orphan handling | Auto-remove knossos-owned. `--keep-orphans` override. | Default auto-removal matches the "rite scope owns its files" invariant. Override for safety |
| D13 | Promotion | Dropped entirely. | Promotion (moving files to user-level) is an edge case that adds complexity without demonstrated need |
| D14 | Final CLI signature | `ari sync [--scope] [--rite] [--source] [--resource] [--dry-run] [--recover] [--overwrite-diverged] [--keep-orphans]` | Complete flag set derived from D1-D13 |
| D15 | No active rite | Smart fallback: scope=all skips rite silently, scope=rite errors. | Allows `ari sync` to work in user-only contexts without requiring a rite |
| D16 | Init wiring | Stays separate from sync. | `ari init` bootstraps a project; `ari sync` assumes bootstrap complete |
| D17 | Deprecation strategy | Hard remove. No aliases. | Clean break. Phase 4a already migrated all manifests; no backward compat debt |
| D18 | Package files | Merge usersync into materialize. Wrappers deleted. | usersync types (ResourceType, Syncer, etc.) absorbed into materialize with new names |
| D19 | Collision detection | Keep as manifest read. Runs after rite scope. | CollisionChecker moves from usersync to materialize. Reads rite PROVENANCE_MANIFEST.yaml after rite scope writes it |
| D20 | Tests | Fresh test suite for unified pipeline. | Existing tests are tightly coupled to current two-pipeline structure. New tests validate the unified flow |
| D21 | Output | Single unified Result type. | SyncResult replaces both materialize.Result and usersync.Result/AllResult |
| D22 | internal/sync/ package | Keep with pruned surface area. | StateManager is still used by trackState(). Remote sync types are deleted |
| D23 | Source resolution | Extend SourceResolver for user scope. | SourceResolver already handles rite resolution. Adding user-scope source resolution keeps resolution in one place |
| D24 | Rite inference | Auto-infer from ACTIVE_RITE. --rite overrides. | Same behavior as current materialize command |
| D25 | No-project mode | User scope works, rite scope skipped/errors. | ~/.claude/ is always available. Project .claude/ requires project discovery |
| D26 | Mena content updates | Sprint 5 updates 60+ mena/rites files. | References to `ari sync` and old flags must be updated across all content |

---

## Section 0: Architecture Overview

### Current State (Two Pipelines)

```
ari sync --rite=X    --> Materializer.MaterializeWithOptions()
  Writes to: project/.claude/     --> PROVENANCE_MANIFEST.yaml
  Source: rites/{X}/              --> 4-tier resolution (project > user > knossos > embedded)

ari sync user all                 --> usersync.Syncer.Sync() x3
  Writes to: ~/.claude/           --> USER_PROVENANCE_MANIFEST.yaml
  Source: $KNOSSOS_HOME/{type}/   --> Single source (knossos home)
```

### Target State (Unified Pipeline)

```
ari sync                          --> Materializer.Sync(SyncOptions{Scope: ScopeAll})
  Phase 1 (rite):
    Writes to: project/.claude/   --> PROVENANCE_MANIFEST.yaml
    Source: rites/{X}/            --> 4-tier resolution
  Phase 2 (user):
    Writes to: ~/.claude/         --> USER_PROVENANCE_MANIFEST.yaml
    Source: $KNOSSOS_HOME/{type}/ --> User source resolution (new)
    Collision check reads Phase 1 manifest
```

### Absorption Strategy

The Materializer struct gains a `Sync()` method that dispatches to scope-gated sub-methods. The existing `MaterializeWithOptions()` and `MaterializeMinimal()` become internal helpers called by the rite scope path. The usersync logic is absorbed into a new `syncUserScope()` method. After Sprint 4, all usersync imports are eliminated.

---

## Section 1: New Type Definitions

Package: `internal/materialize/materialize.go`

### SyncScope

```go
// SyncScope determines which scopes to execute during sync.
type SyncScope string

const (
    // ScopeAll syncs both rite and user scopes (default).
    ScopeAll SyncScope = "all"
    // ScopeRite syncs only the rite scope (project .claude/).
    ScopeRite SyncScope = "rite"
    // ScopeUser syncs only the user scope (~/.claude/).
    ScopeUser SyncScope = "user"
)

// IsValid returns true if the scope is a recognized value.
func (s SyncScope) IsValid() bool {
    switch s {
    case ScopeAll, ScopeRite, ScopeUser:
        return true
    default:
        return false
    }
}
```

### SyncResource

```go
// SyncResource identifies a filterable resource type for sync.
type SyncResource string

const (
    // ResourceAll syncs all resource types (default).
    ResourceAll SyncResource = ""
    // ResourceAgents syncs only agents.
    ResourceAgents SyncResource = "agents"
    // ResourceMena syncs only mena (commands + skills).
    ResourceMena SyncResource = "mena"
    // ResourceHooks syncs only hooks.
    ResourceHooks SyncResource = "hooks"
)

// IsValid returns true if the resource is a recognized value or empty (all).
func (r SyncResource) IsValid() bool {
    switch r {
    case ResourceAll, ResourceAgents, ResourceMena, ResourceHooks:
        return true
    default:
        return false
    }
}
```

### SyncOptions

```go
// SyncOptions configures the unified sync pipeline.
type SyncOptions struct {
    // Scope determines which scopes to execute. Default: ScopeAll.
    Scope SyncScope

    // RiteName is the rite to sync for rite scope. Empty = auto-infer from ACTIVE_RITE.
    RiteName string

    // Source is the explicit rite source path or "knossos" alias. Empty = 4-tier resolution.
    Source string

    // Resource filters sync to a specific resource type. Empty = all resources.
    Resource SyncResource

    // DryRun previews changes without applying.
    DryRun bool

    // Recover adopts existing untracked files into the provenance manifest.
    Recover bool

    // OverwriteDiverged overwrites user-modified files with source versions.
    OverwriteDiverged bool

    // KeepOrphans prevents auto-removal of knossos-owned orphan files.
    KeepOrphans bool
}
```

Validation rules:
- `Scope` must pass `IsValid()`. Default `""` is normalized to `ScopeAll` before dispatch.
- `Resource` must pass `IsValid()`. Default `""` means all resources.
- `OverwriteDiverged` and `KeepOrphans` are independent (not mutually exclusive).
- `RiteName` is only used when scope includes rite. Ignored for scope=user.
- `Source` is only used when scope includes rite. Ignored for scope=user.

### SyncResult

```go
// SyncResult contains the unified outcome of a sync operation.
type SyncResult struct {
    // RiteResult contains rite scope outcome. Nil if rite scope was not executed.
    RiteResult *RiteScopeResult `json:"rite,omitempty"`

    // UserResult contains user scope outcome. Nil if user scope was not executed.
    UserResult *UserScopeResult `json:"user,omitempty"`
}

// RiteScopeResult contains rite scope sync outcome (replaces materialize.Result).
type RiteScopeResult struct {
    Status           string   `json:"status"`
    RiteName         string   `json:"rite_name,omitempty"`
    Source           string   `json:"source,omitempty"`
    SourcePath       string   `json:"source_path,omitempty"`
    OrphansDetected  []string `json:"orphans_detected,omitempty"`
    OrphanAction     string   `json:"orphan_action,omitempty"`
    BackupPath       string   `json:"backup_path,omitempty"`
    HooksSkipped     bool     `json:"hooks_skipped,omitempty"`
    LegacyBackupPath string   `json:"legacy_backup_path,omitempty"`
}

// UserScopeResult contains user scope sync outcome (replaces usersync.AllResult).
type UserScopeResult struct {
    Status    string                       `json:"status"`
    Resources map[SyncResource]*UserResourceResult `json:"resources,omitempty"`
    Totals    UserSyncSummary              `json:"totals"`
    Errors    []UserResourceError          `json:"errors,omitempty"`
}

// UserResourceResult contains per-resource sync outcome (replaces usersync.Result).
type UserResourceResult struct {
    Source    string          `json:"source"`
    Target    string          `json:"target"`
    Changes   UserSyncChanges `json:"changes"`
    Summary   UserSyncSummary `json:"summary"`
}

// UserSyncChanges categorizes sync outcomes by file.
type UserSyncChanges struct {
    Added     []string            `json:"added"`
    Updated   []string            `json:"updated"`
    Skipped   []UserSkippedEntry  `json:"skipped"`
    Unchanged []string            `json:"unchanged"`
}

// UserSkippedEntry explains why a file was skipped.
type UserSkippedEntry struct {
    Name   string `json:"name"`
    Reason string `json:"reason"`
}

// UserSyncSummary provides aggregate counts.
type UserSyncSummary struct {
    Added      int `json:"added"`
    Updated    int `json:"updated"`
    Skipped    int `json:"skipped"`
    Unchanged  int `json:"unchanged"`
    Collisions int `json:"collisions"`
}

// UserResourceError captures an error for a specific resource type.
type UserResourceError struct {
    Resource SyncResource `json:"resource"`
    Err      string       `json:"error"`
}
```

### CollisionChecker (moved from usersync)

Package: `internal/materialize/collision.go` (new file)

```go
// CollisionChecker detects collisions between user-scope and rite-scope resources.
// It reads the rite PROVENANCE_MANIFEST.yaml to determine which resources are
// already managed by a rite, preventing user-scope from shadowing them.
type CollisionChecker struct {
    claudeDir      string          // Project .claude/ directory
    riteEntries    map[string]bool // Rite-managed entries from manifest
    manifestLoaded bool
}

// NewCollisionChecker creates a collision checker that reads the rite manifest.
// claudeDir is the project .claude/ directory. Empty string disables collision checking.
func NewCollisionChecker(claudeDir string) *CollisionChecker

// CheckCollision checks if a manifest key collides with a rite-managed entry.
// Returns (hasCollision bool, riteName string).
func (c *CollisionChecker) CheckCollision(manifestKey string) (bool, string)
```

The collision checker interface is simplified from the usersync version:
- Removes `resourceType` and `nested` parameters (manifest keys are already namespaced).
- Removes `knossosHome` and directory scan fallback (manifest-based detection only; if no manifest exists, no collision is possible because rite scope has not run).
- The `riteName` return value comes from the manifest's ActiveRite field.

---

## Section 2: Sync() Method Contract

Package: `internal/materialize/materialize.go`

New method on Materializer:

```
func (m *Materializer) Sync(opts SyncOptions) (*SyncResult, error)
```

### Pseudocode

```
Sync(opts):
    result = new SyncResult

    // Normalize defaults
    if opts.Scope == "":
        opts.Scope = ScopeAll

    // Validate
    if !opts.Scope.IsValid():
        return error("invalid scope")
    if !opts.Resource.IsValid():
        return error("invalid resource")

    // Phase 1: Rite scope
    if opts.Scope == ScopeAll || opts.Scope == ScopeRite:
        riteResult, err = m.syncRiteScope(opts)
        if err != nil:
            if opts.Scope == ScopeRite:
                return nil, err  // Hard error when rite was explicitly requested
            // scope=all: log warning, continue to user scope
            riteResult = &RiteScopeResult{Status: "skipped", ...}
        result.RiteResult = riteResult

    // Phase 2: User scope
    if opts.Scope == ScopeAll || opts.Scope == ScopeUser:
        userResult, err = m.syncUserScope(opts)
        if err != nil:
            return nil, err
        result.UserResult = userResult

    return result, nil
```

### syncRiteScope(opts) Pseudocode

```
syncRiteScope(opts):
    // Infer rite name
    riteName = opts.RiteName
    if riteName == "":
        riteName = readActiveRite()
        if riteName == "":
            if opts.Scope == ScopeRite:
                return error("no ACTIVE_RITE found, specify --rite")
            // scope=all with no ACTIVE_RITE: run minimal (cross-cutting)
            return m.syncRiteScopeMinimal(opts)

    // Convert SyncOptions to legacy Options for existing pipeline
    legacyOpts = Options{
        DryRun:    opts.DryRun,
        RemoveAll: !opts.KeepOrphans,  // Invert: auto-remove is now default
        KeepAll:   opts.KeepOrphans,
    }

    // Delegate to existing MaterializeWithOptions (unchanged)
    legacyResult, err = m.MaterializeWithOptions(riteName, legacyOpts)
    if err != nil:
        return nil, err

    // Map legacy Result to RiteScopeResult
    return &RiteScopeResult{
        Status:          legacyResult.Status,
        RiteName:        riteName,
        Source:          legacyResult.Source,
        SourcePath:      legacyResult.SourcePath,
        OrphansDetected: legacyResult.OrphansDetected,
        OrphanAction:    legacyResult.OrphanAction,
        BackupPath:      legacyResult.BackupPath,
        HooksSkipped:    legacyResult.HooksSkipped,
        LegacyBackupPath: legacyResult.LegacyBackupPath,
    }, nil
```

### syncRiteScopeMinimal(opts) Pseudocode

```
syncRiteScopeMinimal(opts):
    legacyOpts = Options{DryRun: opts.DryRun, Minimal: true}
    legacyResult, err = m.MaterializeMinimal(legacyOpts)
    if err != nil:
        return nil, err
    return &RiteScopeResult{
        Status: legacyResult.Status,
        Source: "minimal",
    }, nil
```

### Error Handling Strategy

| Scenario | scope=rite | scope=user | scope=all |
|----------|-----------|-----------|-----------|
| No ACTIVE_RITE | Error | N/A | Skip rite, run user |
| Rite not found | Error | N/A | Error (user intent was rite+user) |
| KNOSSOS_HOME unset | N/A | Error | Skip user, log warning |
| User source dir missing | N/A | Skip resource, record error | Skip resource, record error |
| File copy failure | Error (abort) | Record in resource errors | Record in resource errors |

---

## Section 3: User Scope Pipeline

Package: `internal/materialize/user_scope.go` (new file)

### syncUserScope(opts) Pseudocode

This absorbs the logic from `usersync.Syncer.Sync()` and `runUserSyncAll()`.

```
syncUserScope(opts):
    result = new UserScopeResult{
        Status:    "success",
        Resources: map[SyncResource]*UserResourceResult{},
        Totals:    UserSyncSummary{},
        Errors:    []UserResourceError{},
    }

    // Resolve KNOSSOS_HOME
    knossosHome = config.KnossosHome()
    if knossosHome == "":
        return nil, error("KNOSSOS_HOME not set")

    // Resolve user .claude/ directory
    userClaudeDir = paths.UserClaudeDir()

    // Load unified user provenance manifest
    manifestPath = provenance.UserManifestPath(userClaudeDir)
    manifest, err = provenance.LoadOrBootstrap(manifestPath)
    if err != nil:
        return nil, err

    // Initialize collision checker using rite manifest (if rite scope ran)
    claudeDir = m.getClaudeDir()
    collisionChecker = NewCollisionChecker(claudeDir)

    // Determine resource types to sync
    resourceTypes = [ResourceAgents, ResourceMena, ResourceHooks]
    if opts.Resource != ResourceAll:
        resourceTypes = [opts.Resource]

    // Sync each resource type
    for _, resourceType in resourceTypes:
        resourceResult, err = m.syncUserResource(
            resourceType, knossosHome, userClaudeDir,
            manifest, collisionChecker, opts,
        )
        if err != nil:
            result.Errors = append(result.Errors, UserResourceError{
                Resource: resourceType,
                Err:      err.Error(),
            })
            continue

        result.Resources[resourceType] = resourceResult

        // Aggregate totals
        result.Totals.Added += resourceResult.Summary.Added
        result.Totals.Updated += resourceResult.Summary.Updated
        result.Totals.Skipped += resourceResult.Summary.Skipped
        result.Totals.Unchanged += resourceResult.Summary.Unchanged
        result.Totals.Collisions += resourceResult.Summary.Collisions

    // Save manifest
    if !opts.DryRun:
        manifest.LastSync = time.Now().UTC()
        provenance.Save(manifestPath, manifest)
        cleanupOldManifests(userClaudeDir)

    return result, nil
```

### syncUserResource() -- Per-Resource Sync Logic

This absorbs the logic from `usersync.Syncer.syncFiles()`.

```
syncUserResource(resourceType, knossosHome, userClaudeDir, manifest, collisionChecker, opts):
    // Determine source and target directories
    sourceDir, targetConfig = resolveUserResourcePaths(resourceType, knossosHome, userClaudeDir)

    // Check source exists
    if !exists(sourceDir):
        return nil, error("source not found: " + sourceDir)

    // Ensure target directories exist
    if !opts.DryRun:
        ensureTargetDirs(targetConfig)

    result = new UserResourceResult{
        Source: sourceDir,
        Target: targetConfig.displayString(),
        Changes: UserSyncChanges{Added: [], Updated: [], Skipped: [], Unchanged: []},
    }

    // Handle recovery mode
    if opts.Recover:
        recoverUserFiles(targetConfig, sourceDir, manifest, result, opts)

    // Phase 1: Snapshot existing knossos-owned keys for orphan detection
    existingKeys = snapshotKnossosKeys(manifest, resourceType)

    // Phase 2: Walk source and sync each file
    walkSourceFiles(sourceDir, resourceType, func(sourcePath, manifestKey, menaTarget):
        // Mark key as seen
        existingKeys[manifestKey] = true

        // Check collision with rite scope
        if collisionChecker.CheckCollision(manifestKey):
            result.Changes.Skipped = append(result.Changes.Skipped, {manifestKey, "collision"})
            return

        // Compute source checksum
        sourceChecksum = checksum.File(sourcePath)

        // Check existing manifest entry
        entry = manifest.Entries[manifestKey]

        if entry == nil:
            // NEW file
            syncNewFile(sourcePath, manifestKey, menaTarget, targetConfig,
                        sourceChecksum, manifest, result, opts)
        else:
            // EXISTING file -- dispatch on owner
            switch entry.Owner:
            case OwnerUser, OwnerUntracked:
                result.Changes.Skipped = append(..., "user-created")
            case OwnerKnossos:
                syncExistingKnossosFile(sourcePath, manifestKey, menaTarget,
                    targetConfig, sourceChecksum, entry, manifest, result, opts)
    )

    // Phase 3: Orphan removal
    if !opts.DryRun && !opts.KeepOrphans:
        for key, seen in existingKeys:
            if !seen:
                removeOrphan(key, manifest, targetConfig)

    // Calculate summary
    result.Summary = UserSyncSummary{
        Added:      len(result.Changes.Added),
        Updated:    len(result.Changes.Updated),
        Skipped:    len(result.Changes.Skipped),
        Unchanged:  len(result.Changes.Unchanged),
        Collisions: countCollisions(result.Changes.Skipped),
    }

    return result, nil
```

### File Sync State Machine

Each source file transitions through one of these states:

| State | Condition | Action | Manifest Update |
|-------|-----------|--------|-----------------|
| **New** | Not in manifest, target absent | Copy source to target | Add knossos-owned entry |
| **New + Untracked Target** | Not in manifest, target exists | Skip (mark user-created) | Add user-owned entry |
| **New + Recover Match** | Not in manifest, target exists, --recover, checksums match | Adopt as knossos | Add knossos-owned entry |
| **New + Recover Mismatch** | Not in manifest, target exists, --recover, checksums differ | Adopt as user | Add user-owned entry |
| **Unchanged** | In manifest (knossos), source checksum == entry checksum | No-op | No change |
| **Updated** | In manifest (knossos), source changed, target == entry checksum | Copy source to target | Update checksum |
| **Diverged** | In manifest (knossos), source changed, target != entry checksum | Skip (unless --overwrite-diverged) | No change (or update if overwritten) |
| **User-owned** | In manifest (user/untracked) | Skip always | No change |
| **Orphaned** | In manifest (knossos), source deleted | Auto-remove file + entry | Delete entry |
| **Collision** | Manifest key exists in rite manifest | Skip | No change |

---

## Section 4: CLI Specification

Package: `internal/cmd/sync/sync.go` (rewritten)

### Command Definition

```
ari sync [flags]

Synchronize .claude/ configuration from rite and user sources.

By default, syncs both rite scope (project .claude/) and user scope (~/.claude/).
Use --scope to target a specific scope.

Flags:
  --scope string        Sync scope: all, rite, user (default "all")
  --rite string         Rite to sync (default: auto-infer from ACTIVE_RITE)
  --source string       Rite source: path or 'knossos' alias for $KNOSSOS_HOME
  --resource string     Filter to resource type: agents, mena, hooks
  --dry-run             Preview changes without applying
  --recover             Adopt existing untracked files into manifest
  --overwrite-diverged  Overwrite user-modified files with source versions
  --keep-orphans        Preserve knossos-owned files whose source was deleted

Global Flags:
  -o, --output string       Output format: text, json, yaml (default "text")
  -p, --project-dir string  Project root directory (overrides discovery)
  -v, --verbose             Enable verbose output
```

### Flag Definitions

| Flag | Type | Default | Applies To | Description |
|------|------|---------|------------|-------------|
| `--scope` | string | `"all"` | Both | Sync scope selection |
| `--rite` | string | `""` | Rite only | Rite name override. Empty = read ACTIVE_RITE |
| `--source` | string | `""` | Rite only | Explicit rite source path |
| `--resource` | string | `""` | User only | Resource type filter. Empty = all resources |
| `--dry-run` | bool | `false` | Both | Preview mode |
| `--recover` | bool | `false` | Both | Adopt untracked files into manifest |
| `--overwrite-diverged` | bool | `false` | Both | Force-overwrite diverged files |
| `--keep-orphans` | bool | `false` | Both | Skip orphan auto-removal |

### Mutual Exclusivity Rules

- `--scope=rite` + `--resource` : Error ("--resource only applies to user scope")
- `--scope=user` + `--rite` : Error ("--rite only applies to rite scope")
- `--scope=user` + `--source` : Error ("--source only applies to rite scope")

### NeedsProject Behavior

The sync command sets `NeedsProject=false` (non-recursive). Behavior by scope:
- `--scope=all` (default): If no project found, silently skip rite scope, run user scope only.
- `--scope=rite`: Requires project. Error if no .claude/ found.
- `--scope=user`: Does not require project. Works from any directory.

---

## Section 5: SourceResolver Extensions

Package: `internal/materialize/source.go`

### New Method: ResolveUserSources

The SourceResolver gains a method for resolving user-scope resource directories from KNOSSOS_HOME.

```go
// UserResourcePaths holds the source and target configuration for a user resource.
type UserResourcePaths struct {
    ResourceType  SyncResource
    SourceDir     string // e.g., $KNOSSOS_HOME/agents
    TargetDir     string // e.g., ~/.claude/agents (flat resources)
    TargetCmdDir  string // e.g., ~/.claude/commands (mena only)
    TargetSkillDir string // e.g., ~/.claude/skills (mena only)
    Nested        bool   // true for mena, hooks
}

// ResolveUserResource resolves source and target paths for a user-scope resource.
// Returns error if KNOSSOS_HOME is not set.
func (r *SourceResolver) ResolveUserResource(resource SyncResource) (*UserResourcePaths, error)
```

Resolution logic (absorbed from `usersync.NewSyncer`):

| Resource | Source | Target | Nested |
|----------|--------|--------|--------|
| agents | `$KNOSSOS_HOME/agents` | `~/.claude/agents` | false |
| mena | `$KNOSSOS_HOME/mena` | `~/.claude/commands` + `~/.claude/skills` | true |
| hooks | `$KNOSSOS_HOME/hooks` | `~/.claude/hooks` | true |

The method reads `config.KnossosHome()` and constructs paths using `paths.UserClaudeDir()`, `paths.UserAgentsDir()`, `paths.UserCommandsDir()`, `paths.UserSkillsDir()`, `paths.UserHooksDir()`.

---

## Section 6: Deletion Manifest

### Sprint 1 -- Moved Files

| Source | Destination | Rationale |
|--------|-------------|-----------|
| `internal/usersync/collision.go` | `internal/materialize/collision.go` | CollisionChecker moves to materialize per D19 |

### Sprint 3 -- CLI Deletions

| File | Rationale |
|------|-----------|
| `internal/cmd/sync/materialize.go` | Replaced by unified sync command (D3) |
| `internal/cmd/sync/user.go` | Replaced by --scope=user flag (D3) |
| `internal/cmd/sync/user_all.go` | Replaced by --scope=user flag (D3) |
| `internal/cmd/sync/user_agents.go` | Replaced by --scope=user --resource=agents (D4) |
| `internal/cmd/sync/user_mena.go` | Replaced by --scope=user --resource=mena (D4) |
| `internal/cmd/sync/user_hooks.go` | Replaced by --scope=user --resource=hooks (D4) |
| `internal/cmd/sync/status.go` | Dead code, remote sync removed (D6) |
| `internal/cmd/sync/pull.go` | Dead code, remote sync removed (D6) |
| `internal/cmd/sync/push.go` | Dead code, remote sync removed (D6) |
| `internal/cmd/sync/diff.go` | Dead code, remote sync removed (D6) |
| `internal/cmd/sync/resolve.go` | Dead code, remote sync removed (D6) |
| `internal/cmd/sync/history.go` | Dead code, remote sync removed (D6) |
| `internal/cmd/sync/reset.go` | Dead code, remote sync removed (D6) |
| `internal/cmd/sync/cmd_test.go` | Tests for deleted commands |
| `internal/cmd/sync/sync_test.go` | Tests for deleted command structure |

### Sprint 4 -- Package Deletions

| File/Directory | Rationale |
|----------------|-----------|
| `internal/usersync/usersync.go` | Logic absorbed into materialize/user_scope.go (D1, D18) |
| `internal/usersync/collision.go` | Moved in Sprint 1 (D19) |
| `internal/usersync/manifest.go` | Logic absorbed into materialize (D18) |
| `internal/usersync/output.go` | Output types replaced by SyncResult (D21) |
| `internal/usersync/errors.go` | Error types absorbed into materialize (D18) |
| `internal/usersync/hooks.go` | Hook-specific helpers absorbed (D18) |
| `internal/usersync/checksum.go` | Already delegates to internal/checksum; calls replaced (D18) |
| `internal/usersync/usersync_test.go` | Replaced by unified integration tests (D20) |
| `internal/usersync/collision_test.go` | Replaced by unified integration tests (D20) |

### Sprint 4 -- internal/sync/ Pruning

| File | Action | Rationale |
|------|--------|-----------|
| `internal/sync/state.go` | Keep (StateManager, State, TrackedFile, Conflict types) | Still used by trackState() in materialize |
| `internal/sync/state_test.go` | Keep | Tests for retained functionality |
| `internal/sync/pull.go` | Delete | Dead code, remote sync removed (D6) |
| `internal/sync/push.go` | Delete | Dead code, remote sync removed (D6) |
| `internal/sync/diff.go` | Delete | Dead code, remote sync removed (D6) |
| `internal/sync/resolve.go` | Delete | Dead code, remote sync removed (D6) |
| `internal/sync/history.go` | Delete | Dead code, remote sync removed (D6) |
| `internal/sync/remote.go` | Delete | Dead code, remote sync removed (D6) |
| `internal/sync/tracker.go` | Delete | Dead code, remote sync removed (D6) |

### Post-Sprint 4 State

After all deletions, the following packages remain:
- `internal/materialize/` -- Unified pipeline (rite + user scopes)
- `internal/sync/` -- state.go + state_test.go only (StateManager for tracking)
- `internal/provenance/` -- Unchanged (DO NOT MODIFY)
- `internal/cmd/sync/` -- sync.go only (unified command)

---

## Section 7: Integration Test Plan

Package: `internal/materialize/sync_integration_test.go` (new file)

All tests use temporary directories and embedded FS fixtures. No network calls.

### Test Matrix (12 Scenarios)

| # | Name | Scope | Setup | Action | Expected Outcome |
|---|------|-------|-------|--------|------------------|
| T1 | **ScopeAll_Fresh** | all | Empty project .claude/, empty ~/.claude/, rite source with 2 agents + 3 mena + hooks | `Sync(ScopeAll)` | Rite: agents, mena, hooks, CLAUDE.md materialized. User: all resources synced. Both manifests created. |
| T2 | **ScopeRite_Only** | rite | Empty project .claude/, rite source | `Sync(ScopeRite)` | Only project .claude/ populated. No user manifest touched. |
| T3 | **ScopeUser_Only** | user | KNOSSOS_HOME with agents + mena, no project required | `Sync(ScopeUser)` | Only ~/.claude/ populated. No project .claude/ touched. |
| T4 | **ScopeAll_NoRite** | all | No ACTIVE_RITE, KNOSSOS_HOME set | `Sync(ScopeAll)` | Rite scope skipped (minimal/cross-cutting). User scope runs. No error. |
| T5 | **ScopeRite_NoRite** | rite | No ACTIVE_RITE | `Sync(ScopeRite)` | Error: "no ACTIVE_RITE found, specify --rite". |
| T6 | **UserScope_Collision** | user | Rite manifest has `agents/orchestrator.md` (knossos, rite scope). KNOSSOS_HOME/agents has `orchestrator.md` | `Sync(ScopeUser)` | orchestrator.md skipped with "collision" reason. Other files synced. |
| T7 | **UserScope_Diverged** | user | Previous sync created `agents/myagent.md` (knossos-owned). User edited file. Source updated. | `Sync(ScopeAll)` | Default: skip diverged. With --overwrite-diverged: overwrite and update manifest. |
| T8 | **UserScope_Orphan** | user | Previous sync created `agents/old.md`. Source no longer has old.md. | `Sync(ScopeUser)` | Default: auto-remove old.md + delete manifest entry. With --keep-orphans: preserved. |
| T9 | **UserScope_Recover** | user | Existing target files not in manifest. | `Sync(ScopeUser, Recover: true)` | Matching checksums: adopted as knossos-owned. Mismatched: adopted as user-owned. |
| T10 | **ResourceFilter** | user | KNOSSOS_HOME with all resource types | `Sync(ScopeUser, Resource: ResourceAgents)` | Only agents synced. Mena and hooks untouched. |
| T11 | **DryRun** | all | Populated sources | `Sync(ScopeAll, DryRun: true)` | Changes reported but no files written. No manifests modified. |
| T12 | **Idempotency** | all | Run Sync(ScopeAll) twice with same sources | `Sync(ScopeAll)` x2 | Second run reports 0 added, 0 updated, N unchanged. Manifests identical. |

### Test Infrastructure

Each test:
1. Creates a temp directory tree for project .claude/, user .claude/, and knossos home.
2. Sets up embedded FS or filesystem rite sources.
3. Overrides `config.KnossosHome()` via env var.
4. Creates Materializer with test paths via `claudeDirOverride`.
5. Calls `Sync()` and asserts on SyncResult fields.
6. Verifies file contents and manifest entries on disk.

---

## Section 8: Migration Guide

### Command Mapping

| Old Command | New Command |
|-------------|-------------|
| `ari sync` | `ari sync` or `ari sync --scope=rite` |
| `ari sync --rite=X` | `ari sync --rite=X` |
| `ari sync --source=knossos` | `ari sync --source=knossos` |
| `ari sync --dry-run` | `ari sync --dry-run` |
| `ari sync --minimal` | `ari sync --scope=rite` (auto-detects no rite) |
| `ari sync --force` | `ari sync --overwrite-diverged` |
| `ari sync --remove-all` | `ari sync` (default behavior) |
| `ari sync --keep-all` | `ari sync --keep-orphans` |
| `ari sync --promote-all` | *Removed* (D13) |
| `ari sync user all` | `ari sync --scope=user` |
| `ari sync user agents` | `ari sync --scope=user --resource=agents` |
| `ari sync user mena` | `ari sync --scope=user --resource=mena` |
| `ari sync user hooks` | `ari sync --scope=user --resource=hooks` |
| `ari sync user all --force` | `ari sync --scope=user --overwrite-diverged` |
| `ari sync user all --recover` | `ari sync --scope=user --recover` |
| `ari sync status` | *Removed* (D6) |
| `ari sync pull` | *Removed* (D6) |
| `ari sync push` | *Removed* (D6) |
| `ari sync diff` | *Removed* (D6) |
| `ari sync resolve` | *Removed* (D6) |
| `ari sync history` | *Removed* (D6) |
| `ari sync reset` | *Removed* (D6) |

### Flag Mapping

| Old Flag | New Flag | Notes |
|----------|----------|-------|
| `--force`, `-f` | `--overwrite-diverged` | Explicit semantics |
| `--remove-all` | *(default behavior)* | Auto-remove is now default |
| `--keep-all` | `--keep-orphans` | Renamed for clarity |
| `--promote-all` | *Removed* | D13 |
| `--minimal` | *Removed* | Auto-detected from missing ACTIVE_RITE |
| `--verbose`, `-v` | *(global flag)* | Already global |

### Mena Content Search Patterns

Sprint 5 must find and replace these patterns across 60+ mena and rites files:

| Search Pattern | Replacement |
|----------------|-------------|
| `ari sync` | `ari sync` |
| `ari sync user all` | `ari sync --scope=user` |
| `ari sync user agents` | `ari sync --scope=user --resource=agents` |
| `ari sync user mena` | `ari sync --scope=user --resource=mena` |
| `ari sync user hooks` | `ari sync --scope=user --resource=hooks` |
| `--keep-all` | `--keep-orphans` |
| `--remove-all` | *(delete flag, now default)* |
| `--promote-all` | *(delete reference)* |
| `--minimal` | *(delete flag, explain auto-detection)* |
| `--force` (in sync context) | `--overwrite-diverged` |

---

## Section 9: Risk Register

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | Rite scope regression during refactor | Medium | High | Sprint 1 wraps existing MaterializeWithOptions() -- no logic change. Existing rite tests continue passing. |
| R2 | User scope behavior drift from usersync | Medium | Medium | Sprint 2 ports logic line-by-line. Integration tests T6-T9 cover all user sync states (collision, diverged, orphan, recover). |
| R3 | Mena dual-target routing breaks during absorption | Low | High | Mena routing logic (DetectMenaType, RouteMenaFile, StripMenaExtension) stays in materialize package unchanged. Only the calling code moves. |
| R4 | CollisionChecker reads stale manifest after rite scope failure | Low | Medium | D7 specifies rite-first ordering. If rite scope fails, collision checker gracefully returns no collisions (no manifest = no collisions). |
| R5 | CLI flag confusion for users of old commands | Medium | Low | D17 specifies hard remove. Sprint 5 updates all mena content. `ari sync --help` is self-documenting. |
| R6 | Orphan auto-removal destroys user work | Low | High | Auto-removal ONLY affects knossos-owned entries (provenance manifest verified). --keep-orphans escape hatch. Backup before removal. |
| R7 | Sprint 4 deletion misses an import | Low | Medium | Verification step: `go build ./...` and `go vet ./...` must pass after all deletions. grep for `usersync` imports. |
| R8 | Scope=all performance regression (two sync passes) | Low | Low | Each pass is already fast (checksum-based skip). Combined time is sum of individual times (no worse than running both commands). |
| R9 | internal/sync/ state.go has hidden dependents | Low | Medium | Sprint 4 verifies via grep before pruning. state.go is only imported by materialize.trackState(). |
| R10 | Embedded FS user scope sources | N/A | N/A | User scope always reads from filesystem ($KNOSSOS_HOME). Embedded FS is rite-scope only. No risk. |

---

## Section 10: Sprint Boundaries

### Sprint 1 -- Unified Types + Scope-Gated Skeleton

**Goal**: Introduce unified types and dispatch method. Existing behavior unchanged.

**IN SCOPE**:
- New types in `internal/materialize/materialize.go`: SyncScope, SyncResource, SyncOptions, SyncResult, RiteScopeResult, UserScopeResult, UserResourceResult, UserSyncChanges, UserSkippedEntry, UserSyncSummary, UserResourceError
- New method: `Materializer.Sync(SyncOptions) (*SyncResult, error)` -- dispatch skeleton
- New method: `Materializer.syncRiteScope(SyncOptions) (*RiteScopeResult, error)` -- wraps existing MaterializeWithOptions/MaterializeMinimal
- New method: `Materializer.syncUserScope(SyncOptions) (*UserScopeResult, error)` -- returns empty stub result
- Move CollisionChecker from `internal/usersync/collision.go` to `internal/materialize/collision.go` with simplified interface
- Extend SourceResolver with `ResolveUserResource()` method in `internal/materialize/source.go`
- New file: `internal/materialize/user_scope.go` (stub with syncUserScope returning empty result)

**OUT OF SCOPE**:
- User scope implementation (Sprint 2)
- CLI changes (Sprint 3)
- File deletions (Sprint 3-4)
- Content updates (Sprint 5)

**Files Modified**:
| File | Change |
|------|--------|
| `internal/materialize/materialize.go` | Add types (SyncScope, SyncResource, SyncOptions, SyncResult, etc.) and Sync() dispatch method |
| `internal/materialize/source.go` | Add UserResourcePaths type and ResolveUserResource() method |
| `internal/materialize/collision.go` | New file: CollisionChecker (moved + simplified from usersync) |
| `internal/materialize/user_scope.go` | New file: syncUserScope() stub |

**DO NOT TOUCH**:
- `internal/provenance/` (any file)
- `internal/cmd/sync/` (any file)
- `internal/usersync/` (any file -- still imported by CLI)

**Verification**: `CGO_ENABLED=0 go build ./...` passes. Existing materialize tests pass. New Sync() method callable but user scope returns empty result.

---

### Sprint 2 -- User Scope Pipeline

**Goal**: Full user scope implementation. `Sync(ScopeUser)` produces correct results.

**IN SCOPE**:
- Implement `syncUserScope()` in `internal/materialize/user_scope.go`
- Implement `syncUserResource()` for all 3 resource types (agents flat, mena nested/dual-target, hooks nested)
- Implement recovery logic in user scope
- Implement orphan detection and removal in user scope
- Implement collision checking against rite manifest
- Port `usersync.Syncer.syncFiles()` state machine (new, unchanged, updated, diverged, orphan, recovered)
- Port mena source-finding logic (`findMenaSource()`)
- Port hook executable detection (`isExecutable()`)

**OUT OF SCOPE**:
- CLI changes (Sprint 3)
- Deleting usersync package (Sprint 4 -- CLI still imports it)
- Content updates (Sprint 5)

**Files Modified**:
| File | Change |
|------|--------|
| `internal/materialize/user_scope.go` | Full implementation of syncUserScope, syncUserResource, helper functions |
| `internal/materialize/collision.go` | Finalize CheckCollision implementation |

**DO NOT TOUCH**:
- `internal/provenance/` (any file)
- `internal/cmd/sync/` (any file)
- `internal/usersync/` (any file)
- `internal/materialize/materialize.go` (beyond Sprint 1 additions)

**Verification**: Integration tests T1-T12 pass against the new pipeline (tests do not go through CLI). Existing materialize tests unchanged.

---

### Sprint 3 -- CLI Rebuild

**Goal**: Single `ari sync` command replaces all sync subcommands.

**IN SCOPE**:
- Rewrite `internal/cmd/sync/sync.go`: Remove subcommand registration, implement flag-based dispatch
- Add all flags per D14
- Implement mutual exclusivity validation
- Implement NeedsProject logic per scope
- Implement text/JSON output formatting for SyncResult
- Delete all removed CLI files (materialize.go, user*.go, status.go, pull.go, push.go, diff.go, resolve.go, history.go, reset.go, cmd_test.go, sync_test.go)
- Update `internal/cmd/root/root.go` if NewSyncCmd signature changes

**OUT OF SCOPE**:
- Deleting usersync package (Sprint 4)
- Content updates (Sprint 5)

**Files Modified**:
| File | Change |
|------|--------|
| `internal/cmd/sync/sync.go` | Rewrite: single command with flags, calls Materializer.Sync() |
| `internal/cmd/root/root.go` | Update NewSyncCmd call if signature changes |

**Files Deleted**:
| File |
|------|
| `internal/cmd/sync/materialize.go` |
| `internal/cmd/sync/user.go` |
| `internal/cmd/sync/user_all.go` |
| `internal/cmd/sync/user_agents.go` |
| `internal/cmd/sync/user_mena.go` |
| `internal/cmd/sync/user_hooks.go` |
| `internal/cmd/sync/status.go` |
| `internal/cmd/sync/pull.go` |
| `internal/cmd/sync/push.go` |
| `internal/cmd/sync/diff.go` |
| `internal/cmd/sync/resolve.go` |
| `internal/cmd/sync/history.go` |
| `internal/cmd/sync/reset.go` |
| `internal/cmd/sync/cmd_test.go` |
| `internal/cmd/sync/sync_test.go` |

**DO NOT TOUCH**:
- `internal/provenance/` (any file)
- `internal/usersync/` (any file -- deleted in Sprint 4)

**Verification**: `ari sync --help` shows new flags. `ari sync --scope=rite --rite=ecosystem` works. `ari sync --scope=user` works. Old subcommands (`ari sync`, `ari sync user all`) return "unknown command" errors. `CGO_ENABLED=0 go build ./...` passes.

---

### Sprint 4 -- Dead Code Deletion + Integration Tests

**Goal**: Delete usersync package and unused sync files. Verify zero usersync imports.

**IN SCOPE**:
- Delete entire `internal/usersync/` directory (9 files)
- Delete remote sync files from `internal/sync/` (7 files: pull.go, push.go, diff.go, resolve.go, history.go, remote.go, tracker.go)
- Remove any remaining usersync imports from materialize (collision.go should already be self-contained)
- Commit integration test file `internal/materialize/sync_integration_test.go` (written in Sprint 2 but committed here to validate against final package state)
- Verify: `grep -r "usersync" internal/` returns zero results
- Verify: `CGO_ENABLED=0 go build ./...` passes
- Verify: `CGO_ENABLED=0 go test ./...` passes

**OUT OF SCOPE**:
- Content updates (Sprint 5)

**Files Deleted**:
| File/Directory |
|----------------|
| `internal/usersync/usersync.go` |
| `internal/usersync/collision.go` |
| `internal/usersync/manifest.go` |
| `internal/usersync/output.go` |
| `internal/usersync/errors.go` |
| `internal/usersync/hooks.go` |
| `internal/usersync/checksum.go` |
| `internal/usersync/usersync_test.go` |
| `internal/usersync/collision_test.go` |
| `internal/sync/pull.go` |
| `internal/sync/push.go` |
| `internal/sync/diff.go` |
| `internal/sync/resolve.go` |
| `internal/sync/history.go` |
| `internal/sync/remote.go` |
| `internal/sync/tracker.go` |

**DO NOT TOUCH**:
- `internal/provenance/` (any file)
- `internal/sync/state.go` (retained)
- `internal/sync/state_test.go` (retained)

**Verification**: `grep -r "usersync" internal/` returns zero. All tests pass. `go vet ./...` clean.

---

### Sprint 5 -- Mena/Rites Content Migration

**Goal**: Update all content files referencing old CLI commands and flags.

**IN SCOPE**:
- Search and replace across `mena/` directory (60+ files)
- Search and replace across `rites/` directory (agent files, workflow files)
- Apply migration guide patterns from Section 8
- Verify no remaining references to old commands

**OUT OF SCOPE**:
- Code changes (all complete by Sprint 4)

**Search Scope**:
| Directory | File Patterns |
|-----------|---------------|
| `mena/` | `*.md`, `*.yaml` |
| `rites/` | `*.md`, `*.yaml` |
| `knossos/templates/` | `*.md.tpl`, `*.md` |
| `docs/` | `*.md` (design docs referencing CLI) |

**Verification**: `grep -r "ari sync" mena/ rites/ knossos/templates/` returns zero. `grep -r "ari sync user" mena/ rites/ knossos/templates/` returns zero. `grep -r "\-\-promote-all\|\-\-keep-all\|\-\-remove-all\|\-\-minimal" mena/ rites/ knossos/templates/` returns zero (in sync context).

---

## Appendix A: Existing Code Reuse Map

This section maps which usersync functions are absorbed, which are dropped, and where they land.

| usersync Function/Type | Disposition | New Location |
|------------------------|-------------|--------------|
| `Syncer` struct | Absorbed | Logic in `user_scope.go` (no struct equivalent) |
| `NewSyncer()` | Absorbed | `SourceResolver.ResolveUserResource()` + inline setup |
| `Syncer.Sync()` | Absorbed | `Materializer.syncUserScope()` + `syncUserResource()` |
| `Syncer.syncFiles()` | Absorbed | `syncUserResource()` |
| `Syncer.recover()` | Absorbed | `recoverUserFiles()` in `user_scope.go` |
| `Syncer.removeOrphan()` | Absorbed | `removeUserOrphan()` in `user_scope.go` |
| `Syncer.copyFile()` | Absorbed | `copyUserFile()` in `user_scope.go` (uses fileutil.WriteIfChanged) |
| `Syncer.findMenaSource()` | Absorbed | `findMenaSource()` in `user_scope.go` |
| `Syncer.Status()` | Dropped | Replaced by `--dry-run` |
| `CollisionChecker` | Moved | `collision.go` in materialize (simplified interface) |
| `NewCollisionChecker()` | Moved | `NewCollisionChecker(claudeDir)` |
| `CheckCollision()` | Moved | Same name, simplified signature |
| `GetRiteForResource()` | Dropped | Only used by CLI output, not needed |
| `ListRiteResources()` | Dropped | Only used by CLI output, not needed |
| `Result` | Replaced | `UserResourceResult` |
| `AllResult` | Replaced | `UserScopeResult` |
| `Changes` | Replaced | `UserSyncChanges` |
| `Summary` | Replaced | `UserSyncSummary` |
| `SkippedEntry` | Replaced | `UserSkippedEntry` |
| `ResourceError` | Replaced | `UserResourceError` |
| `ResourceType` | Replaced | `SyncResource` |
| `Options` | Absorbed | `SyncOptions` (unified) |
| `ComputeFileChecksum()` | Dropped | Use `checksum.File()` directly |
| `ComputeContentChecksum()` | Dropped | Use `checksum.Bytes()` directly |
| `ComputeDirChecksum()` | Dropped | Use `checksum.Dir()` directly |
| `VerifyChecksum()` | Dropped | Inline comparison |
| `isExecutable()` | Absorbed | `isExecutable()` in `user_scope.go` |
| `isHookConfigFile()` | Dropped | Not used in sync logic |
| `HooksSyncer` | Dropped | No special hook syncer needed |
| `HooksAnalysis` / `AnalyzeSource()` | Dropped | Not used in sync logic |
| `EnsureExecutable()` | Dropped | File permissions preserved via copyFile |
| `cleanupOldManifests()` | Absorbed | `cleanupOldManifests()` in `user_scope.go` |
| `loadManifest()` / `saveManifest()` | Absorbed | Direct calls to `provenance.LoadOrBootstrap()` / `provenance.Save()` |
| Error types (ErrSourceNotFound, etc.) | Absorbed | Inline errors or shared error helpers |

## Appendix B: File Count Summary

| Sprint | Files Created | Files Modified | Files Deleted | Net Change |
|--------|--------------|----------------|---------------|------------|
| 1 | 2 (collision.go, user_scope.go) | 2 (materialize.go, source.go) | 0 | +2 |
| 2 | 0 | 2 (user_scope.go, collision.go) | 0 | 0 |
| 3 | 0 | 2 (sync.go, root.go) | 15 | -15 |
| 4 | 1 (sync_integration_test.go) | 0 | 16 | -15 |
| 5 | 0 | ~60 (mena/rites content) | 0 | 0 |
| **Total** | **3** | **~66** | **31** | **-28** |

# TDD: ADR-0026 Phase 4a -- Unified Provenance Schema

> Technical Design Document for unifying rite-scope and user-scope provenance under a single schema.

**Status**: Approved
**Author**: Context Architect
**Date**: 2026-02-09
**ADR**: docs/decisions/ADR-0026-unified-provenance.md
**Scope**: Phase 4a (schema unification + usersync migration + CLI). Phase 4b (pipeline absorption) is OUT OF SCOPE.
**Predecessor**: TDD-provenance-manifest.md (Phase 2-3, complete)

---

## Decision Register

Decisions locked from stakeholder session on 2026-02-09. These are final; do not re-litigate.

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| D1 | Schema approach | Full unification -- one ProvenanceEntry type for both scopes | Eliminates type duplication (usersync.Entry vs provenance.ProvenanceEntry) and enables single-command visibility |
| D2 | Owner enum | `knossos` / `user` / `untracked` (rename `unknown` to `untracked`) | "untracked" is semantically clearer than "unknown" for pre-existing files |
| D3 | Scope field | New `ScopeType` replaces `SourcePipeline`. Values: `rite` / `user` | "rite" is more accurate than "materialize" (not all rite entries come from materialize stage). "user" distinguishes from "rite" |
| D4 | Schema version | Bump from `1.0` to `2.0` | Breaking change to SourcePipeline field justifies major version bump |
| D5 | Manifest count | Single manifest per scope: `PROVENANCE_MANIFEST.yaml` (rite), `USER_PROVENANCE_MANIFEST.yaml` (user) | Two files avoids write contention between rite-sync (project .claude/) and user-sync (~/.claude/). Each scope has its own lifecycle |
| D6 | Mena tracking | Directory-level for both scopes. MenaType/Target derived at sync time, not persisted | Consistent with Phase 2-3 approach. Mena entries are atomic directories |
| D7 | Migration | Clean break: wipe old JSON manifests, resync fresh. No backward compat | Simple, safe. Old JSON manifests have no data that cannot be reconstructed from source |
| D8 | CLI | `ari provenance show` gains SCOPE column and `--scope=rite\|user` filter | Additive change. Default shows both scopes merged |
| D9 | CollisionChecker | Reads rite `PROVENANCE_MANIFEST.yaml` instead of scanning directories. Fallback to directory scan if manifest missing | Faster, more accurate. Directory scan preserved for bootstrap scenarios |
| D10 | KNOSSOS_MANIFEST.yaml | Retained as-is (region-level, separate concern) | No change |
| D11 | Divergence computation | Computed at sync time (checksum comparison), NOT stored. Eliminates `knossos-diverged` source type | Reduces state surface. Divergence is transient -- it exists only at the moment of comparison |
| D12 | Orphan parity | User-level gains auto-removal of knossos-owned orphans matching rite-level behavior | Parity between the two scopes. User-owned entries are never auto-removed |

---

## Section 0: Gap Analysis Summary

**Current state**: Two independent provenance systems that serve the same purpose but share no types.

1. **Rite-scope** (`internal/provenance/`): Tracks files in project `.claude/` written by the materialize pipeline. Uses `ProvenanceManifest` with `ProvenanceEntry` (YAML). Provides divergence detection, collector pattern, and CLI (`ari provenance show`).

2. **User-scope** (`internal/usersync/`): Tracks files in `~/.claude/` written by the usersync pipeline. Uses `Manifest` with `Entry` (JSON). Three separate per-resource JSON manifests (`USER_AGENT_MANIFEST.json`, `USER_MENA_MANIFEST.json`, `USER_HOOKS_MANIFEST.json`). Has its own `SourceType` enum (`knossos`, `knossos-diverged`, `user`) that partially overlaps with provenance's `OwnerType`.

**Gap**: No unified view across scopes. `ari provenance show` only shows rite-scope. Divergence is represented as stored state (`knossos-diverged`) in usersync but computed dynamically in provenance. Two different manifest formats (JSON vs YAML). Two different type systems for the same ownership concept.

**Success criteria**: Single `ProvenanceEntry` type used by both scopes. Single `ari provenance show` command showing all tracked files. Three JSON manifests replaced by one YAML manifest per scope. `knossos-diverged` eliminated as stored state.

---

## Section 1: Schema Changes

Package: `internal/provenance/provenance.go`

### BEFORE (current v1.0)

```go
const ManifestFileName = "PROVENANCE_MANIFEST.yaml"
const CurrentSchemaVersion = "1.0"

type OwnerType string
const (
    OwnerKnossos OwnerType = "knossos"
    OwnerUser    OwnerType = "user"
    OwnerUnknown OwnerType = "unknown"
)

type ProvenanceEntry struct {
    Owner          OwnerType `yaml:"owner"`
    SourcePipeline string    `yaml:"source_pipeline,omitempty"`
    SourcePath     string    `yaml:"source_path,omitempty"`
    SourceType     string    `yaml:"source_type,omitempty"`
    Checksum       string    `yaml:"checksum"`
    LastSynced     time.Time `yaml:"last_synced"`
}
```

### AFTER (v2.0)

```go
const ManifestFileName = "PROVENANCE_MANIFEST.yaml"
const UserManifestFileName = "USER_PROVENANCE_MANIFEST.yaml"
const CurrentSchemaVersion = "2.0"

type OwnerType string
const (
    OwnerKnossos   OwnerType = "knossos"
    OwnerUser      OwnerType = "user"
    OwnerUntracked OwnerType = "untracked"  // renamed from OwnerUnknown
)

type ScopeType string
const (
    ScopeRite ScopeType = "rite"
    ScopeUser ScopeType = "user"
)

type ProvenanceEntry struct {
    Owner      OwnerType `yaml:"owner"`
    Scope      ScopeType `yaml:"scope"`                // NEW: replaces SourcePipeline
    SourcePath string    `yaml:"source_path,omitempty"`
    SourceType string    `yaml:"source_type,omitempty"`
    Checksum   string    `yaml:"checksum"`
    LastSynced time.Time `yaml:"last_synced"`
}
```

### Field-level changes

| Field | Change | Detail |
|-------|--------|--------|
| `SourcePipeline` | REMOVED | Was always `"materialize"` or empty. Replaced by `Scope` |
| `Scope` | ADDED | `ScopeType` -- `"rite"` for project-scope entries, `"user"` for user-scope entries |
| `OwnerUnknown` | RENAMED | Becomes `OwnerUntracked` (value changes from `"unknown"` to `"untracked"`) |
| `CurrentSchemaVersion` | CHANGED | `"1.0"` to `"2.0"` |
| `UserManifestFileName` | ADDED | New const `"USER_PROVENANCE_MANIFEST.yaml"` |

### OwnerType.IsValid() update

```go
func (o OwnerType) IsValid() bool {
    switch o {
    case OwnerKnossos, OwnerUser, OwnerUntracked:
        return true
    default:
        return false
    }
}
```

### ScopeType.IsValid() (new)

```go
func (s ScopeType) IsValid() bool {
    switch s {
    case ScopeRite, ScopeUser:
        return true
    default:
        return false
    }
}

func (s ScopeType) String() string {
    return string(s)
}
```

### ProvenanceManifest struct -- no changes

The `ProvenanceManifest` struct is unchanged. Both rite-scope and user-scope manifests use the same structure. The `ActiveRite` field is only populated in rite-scope manifests; it remains empty in user-scope manifests.

---

## Section 2: Manifest I/O Changes

File: `internal/provenance/manifest.go`

### validateManifest updates

Three changes to the validation function:

1. **Owner enum**: Replace `OwnerUnknown` with `OwnerUntracked` in the `IsValid()` check. Since `IsValid()` is updated in Section 1, the validator calls `entry.Owner.IsValid()` unchanged.

2. **Scope field**: Add validation that `entry.Scope` is non-empty and valid (must be `ScopeRite` or `ScopeUser`). This is a required field.

3. **SourcePipeline removal**: Remove the validation block that checks `entry.SourcePipeline != "" && entry.SourcePipeline != "materialize"`. This field no longer exists.

Specific change to `validateManifest()`:

```
REMOVE:
    // If SourcePipeline is present, must be "materialize"
    if entry.SourcePipeline != "" && entry.SourcePipeline != "materialize" {
        issues = append(issues, ...)
    }

ADD:
    // Required: Entry.Scope must be a valid scope type
    if entry.Scope == "" {
        issues = append(issues, "entry '"+path+"' missing required field: scope")
    } else if !entry.Scope.IsValid() {
        issues = append(issues, "entry '"+path+"' has invalid scope: "+string(entry.Scope))
    }
```

### structurallyEqual updates

Replace `SourcePipeline` comparison with `Scope` comparison:

```
BEFORE:
    if entryA.Owner != entryB.Owner ||
        entryA.SourcePipeline != entryB.SourcePipeline ||
        entryA.SourcePath != entryB.SourcePath ||
        entryA.SourceType != entryB.SourceType ||
        entryA.Checksum != entryB.Checksum {

AFTER:
    if entryA.Owner != entryB.Owner ||
        entryA.Scope != entryB.Scope ||
        entryA.SourcePath != entryB.SourcePath ||
        entryA.SourceType != entryB.SourceType ||
        entryA.Checksum != entryB.Checksum {
```

### UserManifestPath (new helper)

```go
// UserManifestPath returns the full path to USER_PROVENANCE_MANIFEST.yaml
// within the user .claude directory (typically ~/.claude/).
func UserManifestPath(userClaudeDir string) string {
    return filepath.Join(userClaudeDir, UserManifestFileName)
}
```

### Load/Save -- no changes

The `Load()` and `Save()` functions are generic (they accept a path and operate on `*ProvenanceManifest`). They work for both rite-scope and user-scope manifests without modification.

### LoadOrBootstrap -- no changes

The bootstrap version starts at `CurrentSchemaVersion` (which is now `"2.0"`). No other changes needed. New manifests bootstrap with `"2.0"` automatically.

### Schema version migration in Load

When `Load()` encounters a `schema_version: "1.0"` manifest, it must handle the upgrade path. Add a migration step after parsing and before validation:

```go
// After yaml.Unmarshal, before validateManifest:
if manifest.SchemaVersion == "1.0" {
    migrateV1ToV2(&manifest)
}
```

The `migrateV1ToV2` function:

```go
func migrateV1ToV2(m *ProvenanceManifest) {
    m.SchemaVersion = "2.0"
    for _, entry := range m.Entries {
        // SourcePipeline "materialize" -> Scope "rite"
        // (SourcePipeline field will be zero-valued after struct change)
        if entry.Scope == "" {
            entry.Scope = ScopeRite
        }
        // OwnerUnknown -> OwnerUntracked
        if entry.Owner == "unknown" {
            entry.Owner = OwnerUntracked
        }
    }
}
```

Rationale: Existing v1.0 PROVENANCE_MANIFEST.yaml files in project `.claude/` directories must continue to load. The migration is lossless because `SourcePipeline` was always `"materialize"` for all entries.

Note: The `SourcePipeline` field will not appear in YAML after deserialization because the struct field no longer exists. The YAML parser silently ignores unknown fields.

---

## Section 3: Paths Changes

File: `internal/paths/paths.go`

### New function

```go
// UserProvenanceManifest returns the path to the user provenance manifest.
func UserProvenanceManifest() string {
    return filepath.Join(UserClaudeDir(), "USER_PROVENANCE_MANIFEST.yaml")
}
```

### Deprecation plan

The following functions are NOT removed in Phase 4a. They continue to exist because the usersync migration happens incrementally (Sprint 2). They will be removed after Sprint 2 is complete and all callers are updated.

Functions marked for removal (add `// Deprecated: Use UserProvenanceManifest() instead.` comment):

- `UserAgentManifest()` -- returns `USER_AGENT_MANIFEST.json`
- `UserMenaManifest()` -- returns `USER_MENA_MANIFEST.json`
- `UserHooksManifest()` -- returns `USER_HOOKS_MANIFEST.json`
- `UserSkillManifest()` -- returns `USER_SKILL_MANIFEST.json`
- `UserCommandManifest()` -- returns `USER_COMMAND_MANIFEST.json`

These are removed at the end of Sprint 2 after all callers migrate.

---

## Section 4: Usersync Migration

This is the largest change. The usersync package currently has its own manifest type (`Manifest` + `Entry`) stored as JSON. This section migrates to using `provenance.ProvenanceEntry` stored as YAML.

### 4.1 Type Mapping

| usersync (current) | provenance (v2.0) | Notes |
|--------------------|--------------------|-------|
| `usersync.Entry` | `provenance.ProvenanceEntry` | REMOVED. All usersync code uses ProvenanceEntry |
| `usersync.Manifest` | `provenance.ProvenanceManifest` | REMOVED. All usersync code uses ProvenanceManifest |
| `usersync.SourceType` | `provenance.OwnerType` | REMOVED |
| `usersync.SourceKnossos` | `provenance.OwnerKnossos` | Direct mapping |
| `usersync.SourceDiverged` | (eliminated) | Divergence is computed, not stored (D11) |
| `usersync.SourceUser` | `provenance.OwnerUser` | Direct mapping |
| `Entry.Source` | `ProvenanceEntry.Owner` | Field name change |
| `Entry.InstalledAt` | `ProvenanceEntry.LastSynced` | Field name change |
| `Entry.Checksum` | `ProvenanceEntry.Checksum` | Same field |
| `Entry.MenaType` | (not persisted) | Derived at sync time from source filename extension |
| `Entry.Target` | (not persisted) | Derived at sync time from mena type routing |

### 4.2 Manifest Key Namespacing

Currently each resource type has its own manifest file with unnamespaced keys (e.g., `"test.md"` in the agents manifest, `"commit/INDEX.md"` in the mena manifest). With a single manifest, keys are namespaced by resource type's target location:

| Resource | Key format | Examples |
|----------|-----------|----------|
| agents | `agents/{name}` | `agents/orchestrator.md`, `agents/custom-agent.md` |
| mena (commands) | `commands/{dir}/` | `commands/commit/`, `commands/consult/` |
| mena (skills) | `skills/{dir}/` | `skills/prompting/`, `skills/lexicon/` |
| hooks | `hooks/{path}` | `hooks/ari/hooks.yaml`, `hooks/ari/lib/context.sh` |

This matches the key format already used in rite-scope `PROVENANCE_MANIFEST.yaml`, ensuring consistency.

### 4.3 Syncer Changes

File: `internal/usersync/usersync.go`

The `Syncer` struct changes:

```
BEFORE:
    manifestPath string  // path to per-resource JSON manifest

AFTER:
    manifestPath string  // path to unified USER_PROVENANCE_MANIFEST.yaml
```

The `NewSyncer()` function changes all three resource types to use the same manifest path:

```go
// All resource types now point to the same manifest
s.manifestPath = paths.UserProvenanceManifest()
```

The `syncFiles()` function changes to use `provenance.ProvenanceEntry` instead of `usersync.Entry`. All comparisons against `SourceDiverged` are replaced with computed divergence (compare target checksum to manifest checksum).

Key behavioral change in `syncFiles()` for existing entries with `Owner == OwnerKnossos`:

```
BEFORE:
    switch entry.Source {
    case SourceUser:     // skip
    case SourceDiverged: // skip (or force)
    case SourceKnossos:  // update if source changed, detect divergence

AFTER:
    switch entry.Owner {
    case OwnerUser:      // skip
    case OwnerUntracked: // skip (treat as user)
    case OwnerKnossos:
        // Compute divergence: compare target checksum to manifest checksum
        targetChecksum, _ := ComputeFileChecksum(targetPath)
        if targetChecksum != entry.Checksum {
            // Target has been locally modified
            if opts.Force {
                // Overwrite + reset to knossos
            } else {
                // Skip (diverged)
            }
        } else {
            // Target matches manifest -- safe to update from source
            if sourceChecksum != entry.Checksum {
                // Source changed, target unchanged -- update
            } else {
                // Both match -- unchanged
            }
        }
```

### 4.4 Manifest I/O in Syncer

Replace `loadManifest()` and `saveManifest()` methods with calls to `provenance.LoadOrBootstrap()` and `provenance.Save()`:

```go
func (s *Syncer) loadManifest() (*provenance.ProvenanceManifest, error) {
    return provenance.LoadOrBootstrap(s.manifestPath)
}

func (s *Syncer) saveManifest(manifest *provenance.ProvenanceManifest) error {
    manifest.LastSync = time.Now().UTC()
    return provenance.Save(s.manifestPath, manifest)
}
```

The `manifestJSON`, `entryJSON` types, and the `cleanupOldManifests()` method body change to handle migration (see Section 4.6).

The public `LoadManifest()` and `SaveManifest()` convenience functions are REMOVED. No external callers exist outside the usersync package itself and its tests.

### 4.5 Concurrency Constraint

Currently the three resource syncers (`agents`, `mena`, `hooks`) each write to their own manifest file and can safely run in parallel. With a single `USER_PROVENANCE_MANIFEST.yaml`, they share a manifest file.

Constraint: In `user_all.go`, the sequential `for` loop over resource types is already sequential. No change needed. The loop in `runUserSyncAll()` iterates `[ResourceAgents, ResourceMena, ResourceHooks]` one at a time. Each syncer loads the manifest at the start of its `Sync()` call and saves it at the end, ensuring all three resource types' entries accumulate correctly.

The existing sequential loop is preserved as a correctness requirement, not coincidence:

```go
// user_all.go -- this sequential loop is load-bearing.
// DO NOT parallelize: all three resource types write to the same manifest file.
for _, resourceType := range resourceTypes {
    syncer, err := usersync.NewSyncer(resourceType)
    ...
    result, err := syncer.Sync(opts)
    ...
}
```

### 4.6 Migration: Old JSON to New YAML

When usersync runs for the first time after the schema change, old JSON manifests exist and the new YAML manifest does not. The migration path:

1. `loadManifest()` calls `provenance.LoadOrBootstrap(manifestPath)`.
2. `LoadOrBootstrap` finds no file at `USER_PROVENANCE_MANIFEST.yaml` and returns an empty `ProvenanceManifest` with `SchemaVersion: "2.0"`.
3. Empty manifest triggers a full resync for all resource types (every source file is treated as new).
4. After sync, the new YAML manifest is written via `provenance.Save()`.
5. The `cleanupOldManifests()` method (called after each successful sync) removes old JSON files.

Update `cleanupOldManifests()` to handle all legacy manifests:

```go
func (s *Syncer) cleanupOldManifests() {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return
    }
    claudeDir := filepath.Join(homeDir, ".claude")
    oldManifests := []string{
        filepath.Join(claudeDir, "USER_AGENT_MANIFEST.json"),
        filepath.Join(claudeDir, "USER_MENA_MANIFEST.json"),
        filepath.Join(claudeDir, "USER_HOOKS_MANIFEST.json"),
        filepath.Join(claudeDir, "USER_COMMAND_MANIFEST.json"),
        filepath.Join(claudeDir, "USER_SKILL_MANIFEST.json"),
    }
    for _, path := range oldManifests {
        // Backup before removal for safety
        data, err := os.ReadFile(path)
        if err != nil {
            continue // Already gone or unreadable
        }
        backupPath := path + ".v2-backup"
        os.WriteFile(backupPath, data, 0644) // Best effort
        os.Remove(path)
    }
}
```

Change: `cleanupOldManifests()` is now called for ALL resource types (not just `ResourceMena`). Remove the `if s.resourceType != ResourceMena { return }` guard.

### 4.7 Entry Construction in syncFiles

When creating entries in `syncFiles()`, map to the unified type:

```go
// New file added:
manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
    Owner:      provenance.OwnerKnossos,
    Scope:      provenance.ScopeUser,
    SourcePath: sourceRelPath,  // relative path from knossos home
    SourceType: "user-sync",    // identifies the usersync pipeline
    Checksum:   sourceChecksum,
    LastSynced: result.SyncedAt,
}

// User-created file discovered:
manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
    Owner:      provenance.OwnerUser,
    Scope:      provenance.ScopeUser,
    Checksum:   targetChecksum,
    LastSynced: result.SyncedAt,
}
```

The `SourceType` value `"user-sync"` distinguishes entries created by the usersync pipeline from those created by the materialize pipeline (which use values like `"project"`, `"template"`, `"embedded"`).

### 4.8 Manifest Key Prefix Computation

The `manifestKey` in `syncFiles()` must be prefixed by resource type. Add a helper:

```go
func (s *Syncer) prefixManifestKey(key string) string {
    switch s.resourceType {
    case ResourceAgents:
        return "agents/" + key
    case ResourceMena:
        // Mena keys already have the target directory in the path
        // (e.g., key is "commit/INDEX.md", prefix with target "commands/" or "skills/")
        // Handled in the mena-specific code path below
        return key  // Caller must handle mena prefixing
    case ResourceHooks:
        return "hooks/" + key
    default:
        return key
    }
}
```

For mena entries, the manifest key prefix depends on the mena type routing:

```go
if menaTarget == "commands" {
    manifestKey = "commands/" + manifestKey + "/"  // directory entry
} else {
    manifestKey = "skills/" + manifestKey + "/"   // directory entry
}
```

Note: Mena is tracked at the directory level (with trailing `/`), matching rite-scope behavior. The per-file manifest key currently used in usersync changes to per-directory keys.

### 4.9 Files Removed/Modified Summary

| File | Change |
|------|--------|
| `usersync/usersync.go` | Replace `Entry`/`Manifest` usage with `provenance.ProvenanceEntry`/`provenance.ProvenanceManifest`. Remove `SourceType` enum. Update `syncFiles()` divergence logic. Add manifest key prefixing |
| `usersync/manifest.go` | Remove `Manifest`, `Entry`, `manifestJSON`, `entryJSON` types. Remove `loadManifest()`, `saveManifest()` methods (replaced by provenance.Load/Save). Remove `LoadManifest()`, `SaveManifest()` public functions. Keep `cleanupOldManifests()` updated per 4.6 |
| `usersync/checksum.go` | No changes (delegates to `checksum` package) |
| `usersync/collision.go` | Changes per Section 5 |
| `usersync/hooks.go` | No changes (executable bit logic is independent of manifest type) |
| `usersync/errors.go` | No changes |
| `usersync/output.go` | No changes (operates on `Result`, not manifest types) |

---

## Section 5: CollisionChecker Migration

File: `internal/usersync/collision.go`

### Current behavior

`CollisionChecker.CheckCollision()` walks the rite directories on disk to find name conflicts. For each rite, it checks if the resource path exists under `rites/{riteName}/{subDir}/{searchName}`.

### New behavior

The `CollisionChecker` reads the rite-scope `PROVENANCE_MANIFEST.yaml` to determine which resource names are claimed by the rite pipeline. This is faster and more accurate because it uses the manifest of record rather than scanning directories that may contain work-in-progress files.

### Implementation

Add a new field and constructor logic:

```go
type CollisionChecker struct {
    knossosHome    string
    ritesDir       string
    resourceType   ResourceType
    nested         bool
    riteEntries    map[string]bool  // NEW: entries from rite manifest
    manifestLoaded bool             // NEW: whether manifest was loaded
}
```

Add a `loadRiteManifest()` method:

```go
func (c *CollisionChecker) loadRiteManifest(claudeDir string) {
    if c.manifestLoaded {
        return
    }
    c.manifestLoaded = true
    c.riteEntries = make(map[string]bool)

    manifestPath := provenance.ManifestPath(claudeDir)
    manifest, err := provenance.Load(manifestPath)
    if err != nil {
        return // Fallback to directory scan
    }

    for key, entry := range manifest.Entries {
        if entry.Scope == provenance.ScopeRite && entry.Owner == provenance.OwnerKnossos {
            c.riteEntries[key] = true
        }
    }
}
```

Update `CheckCollision()`:

```go
func (c *CollisionChecker) CheckCollision(name string) (bool, string) {
    // If manifest was loaded successfully, use manifest entries
    if c.manifestLoaded && len(c.riteEntries) > 0 {
        prefixedName := c.resourcePrefix() + name
        if c.riteEntries[prefixedName] {
            return true, "(from manifest)"
        }
        return false, ""
    }

    // Fallback: directory scan (original behavior)
    // ... existing code unchanged ...
}
```

The `resourcePrefix()` helper:

```go
func (c *CollisionChecker) resourcePrefix() string {
    switch c.resourceType {
    case ResourceAgents:
        return "agents/"
    case ResourceMena:
        return ""  // Mena keys already include commands/ or skills/ prefix
    case ResourceHooks:
        return "hooks/"
    default:
        return ""
    }
}
```

### Passing claudeDir to CollisionChecker

The `CollisionChecker` needs the project `.claude/` directory path to load the rite manifest. This requires a new parameter in `NewCollisionChecker()`:

```go
func NewCollisionChecker(resourceType ResourceType, nested bool, claudeDir string) *CollisionChecker {
    c := &CollisionChecker{
        knossosHome:  config.KnossosHome(),
        ritesDir:     filepath.Join(config.KnossosHome(), "rites"),
        resourceType: resourceType,
        nested:       nested,
    }
    if claudeDir != "" {
        c.loadRiteManifest(claudeDir)
    }
    return c
}
```

Callers that do not have a `claudeDir` (user-level sync without a project context) pass `""`, triggering the directory scan fallback. This is the expected path for `ari sync user agents` when run outside a project directory.

### Limitation

The collision checker cannot load the rite manifest when run outside a project directory (no `.knossos/PROVENANCE_MANIFEST.yaml`). In this case, the existing directory scan fallback handles collision detection. This is acceptable because collision detection is a best-effort optimization -- the worst case is a false negative (user resource shadows a rite resource), which the user can diagnose via `ari provenance show`.

---

## Section 6: CLI Changes

File: `internal/cmd/provenance/provenance.go`

### ShowEntry gains Scope field

```go
type ShowEntry struct {
    Path       string `json:"path"`
    Owner      string `json:"owner"`
    Scope      string `json:"scope"`      // NEW
    SourcePath string `json:"source_path,omitempty"`
    SourceType string `json:"source_type,omitempty"`
    Status     string `json:"status"`
    Checksum   string `json:"checksum,omitempty"`
}
```

### Table output gains SCOPE column

```go
func (s *ShowOutput) Headers() []string {
    return []string{"PATH", "OWNER", "SCOPE", "SOURCE", "STATUS"}
}

func (s *ShowOutput) Rows() [][]string {
    rows := make([][]string, len(s.Entries))
    for i, e := range s.Entries {
        source := formatSource(e.SourcePath, e.SourceType)
        rows[i] = []string{e.Path, e.Owner, e.Scope, source, e.Status}
    }
    return rows
}
```

### New --scope flag

```go
func newShowCmd(ctx *cmdContext) *cobra.Command {
    var scopeFilter string

    cmd := &cobra.Command{
        Use:   "show",
        Short: "Display provenance manifest",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runShow(ctx, scopeFilter)
        },
    }

    cmd.Flags().StringVar(&scopeFilter, "scope", "",
        "Filter by scope: rite, user (default: show both)")

    return cmd
}
```

### runShow updated to load both manifests

```go
func runShow(ctx *cmdContext, scopeFilter string) error {
    printer := ctx.getPrinter()
    resolver := ctx.GetResolver()

    var allEntries []*ShowEntry

    // Load rite-scope manifest (from project .claude/)
    if scopeFilter == "" || scopeFilter == "rite" {
        claudeDir := filepath.Join(resolver.ProjectRoot(), ".claude")
        manifestPath := provenance.ManifestPath(claudeDir)
        manifest, err := provenance.LoadOrBootstrap(manifestPath)
        if err == nil {
            for path, entry := range manifest.Entries {
                allEntries = append(allEntries, makeShowEntry(
                    path, entry, "rite", claudeDir, *ctx.Verbose))
            }
        }
    }

    // Load user-scope manifest (from ~/.claude/)
    if scopeFilter == "" || scopeFilter == "user" {
        userClaudeDir := paths.UserClaudeDir()
        userManifestPath := provenance.UserManifestPath(userClaudeDir)
        userManifest, err := provenance.LoadOrBootstrap(userManifestPath)
        if err == nil {
            for path, entry := range userManifest.Entries {
                displayPath := "~/" + path  // Prefix for visual distinction
                allEntries = append(allEntries, makeShowEntry(
                    displayPath, entry, "user", userClaudeDir, *ctx.Verbose))
            }
        }
    }

    // ... rest of output logic ...
}
```

### User entries prefixed with "~/"

In the table output, user-scope entries have their path prefixed with `~/` to visually distinguish them from rite-scope entries. This prefix is display-only and does not affect the manifest key.

Example output:

```
PATH                          OWNER     SCOPE  SOURCE                                      STATUS
agents/orchestrator.md        knossos   rite   rites/ecosystem/agents/orchestrator.md       match
commands/commit/              knossos   rite   mena/operations/commit/                      match
CLAUDE.md                     knossos   rite   (generated)                                  match
~/agents/custom-agent.md      knossos   user   agents/custom-agent.md                       match
~/commands/commit/            knossos   user   mena/commit/                                 diverged
~/hooks/ari/hooks.yaml        user      user   (user-created)                               -
```

### JSON/YAML output

For structured output formats (JSON, YAML), the combined manifest is output as a unified structure:

```go
type CombinedOutput struct {
    Rite *provenance.ProvenanceManifest `json:"rite,omitempty" yaml:"rite,omitempty"`
    User *provenance.ProvenanceManifest `json:"user,omitempty" yaml:"user,omitempty"`
}
```

---

## Section 7: Orphan Detection Changes

### Current rite-scope behavior

In `internal/materialize/materialize.go`, orphan detection compares the provenance manifest entries against the current materialization's collector entries. Entries in the manifest that are not in the collector output and have `Owner == OwnerKnossos` are orphans. These orphans are auto-removed (file deleted, entry removed from manifest).

### New user-scope behavior

User-scope orphan detection follows the same pattern. After a successful sync for a resource type, entries in the `USER_PROVENANCE_MANIFEST.yaml` with `Owner == OwnerKnossos` that were not encountered during the current sync pass are orphans.

Implementation within `syncFiles()`:

1. Before the `WalkDir`, snapshot existing manifest keys for the current resource type prefix (e.g., `agents/`, `commands/`, `skills/`, `hooks/`).
2. During `WalkDir`, mark each processed key as "seen".
3. After `WalkDir`, iterate unseen keys. For each:
   - If `entry.Owner == OwnerKnossos`: remove the target file/directory and delete the manifest entry.
   - If `entry.Owner == OwnerUser` or `entry.Owner == OwnerUntracked`: leave unchanged (user-owned entries are never auto-removed).

```go
func (s *Syncer) syncFiles(manifest *provenance.ProvenanceManifest, result *Result, opts Options) error {
    // Phase 1: Snapshot current keys for this resource type
    prefix := s.resourcePrefix()
    existingKeys := make(map[string]bool)
    for key, entry := range manifest.Entries {
        if strings.HasPrefix(key, prefix) && entry.Owner == provenance.OwnerKnossos {
            existingKeys[key] = false  // false = not yet seen
        }
    }

    // Phase 2: Walk source and sync (existing logic, adapted)
    err := filepath.WalkDir(s.sourceDir, func(path string, d os.DirEntry, err error) error {
        // ... existing sync logic ...
        // After processing each file:
        existingKeys[fullManifestKey] = true  // mark as seen
        return nil
    })
    if err != nil {
        return err
    }

    // Phase 3: Orphan removal
    if !opts.DryRun {
        for key, seen := range existingKeys {
            if !seen {
                s.removeOrphan(key, manifest, result)
            }
        }
    }

    return nil
}
```

The `removeOrphan()` helper:

```go
func (s *Syncer) removeOrphan(key string, manifest *provenance.ProvenanceManifest, result *Result) {
    entry := manifest.Entries[key]
    if entry == nil || entry.Owner != provenance.OwnerKnossos {
        return // Safety: only remove knossos-owned orphans
    }

    // Determine target path from key
    targetPath := s.keyToTargetPath(key)
    if targetPath == "" {
        return
    }

    // Remove file or directory
    if strings.HasSuffix(key, "/") {
        os.RemoveAll(targetPath)
    } else {
        os.Remove(targetPath)
    }

    // Remove from manifest
    delete(manifest.Entries, key)
}
```

---

## Section 8: Sprint Decomposition

### Sprint 1: Schema Unification (provenance package only)

**Goal**: Update provenance types and manifest I/O to v2.0 schema. Rite-scope pipeline continues to work. No usersync changes.

**Tasks**:

1. Update `OwnerType` enum: rename `OwnerUnknown` to `OwnerUntracked` (value: `"untracked"`)
2. Add `ScopeType` type with `ScopeRite` and `ScopeUser` constants and `IsValid()`/`String()` methods
3. Replace `SourcePipeline` field with `Scope` field in `ProvenanceEntry`
4. Add `UserManifestFileName` const
5. Add `UserManifestPath()` helper function
6. Bump `CurrentSchemaVersion` from `"1.0"` to `"2.0"`
7. Add `migrateV1ToV2()` function in `manifest.go`; wire into `Load()` after parse
8. Update `validateManifest()`: remove SourcePipeline check, add Scope validation
9. Update `structurallyEqual()`: replace `SourcePipeline` with `Scope`
10. Update all `collector.Record()` calls in `internal/materialize/materialize.go`: replace `SourcePipeline: "materialize"` with `Scope: provenance.ScopeRite`
11. Update `internal/provenance/divergence.go`: remove `SourcePipeline` reference in comment on line 74
12. Update `internal/provenance/provenance_test.go`: replace all `SourcePipeline: "materialize"` with `Scope: provenance.ScopeRite`, replace `OwnerUnknown` with `OwnerUntracked`
13. Update `internal/materialize/provenance_integration_test.go`: replace `SourcePipeline` assertion with `Scope` assertion
14. Add `UserProvenanceManifest()` to `internal/paths/paths.go`
15. Deprecation comments on `UserAgentManifest()`, `UserMenaManifest()`, `UserHooksManifest()`, `UserSkillManifest()`, `UserCommandManifest()`

**Verification**:
- `CGO_ENABLED=0 go build ./cmd/ari` succeeds
- `CGO_ENABLED=0 go vet ./...` clean
- `CGO_ENABLED=0 go test ./internal/provenance/...` passes (all existing tests updated)
- `CGO_ENABLED=0 go test ./internal/materialize/...` passes (provenance integration test updated)
- Existing v1.0 `PROVENANCE_MANIFEST.yaml` files load correctly (migration test)
- `ari sync materialize` produces v2.0 manifest with `scope: rite` entries

**DO NOT**:
- Touch `internal/usersync/` (Sprint 2)
- Touch `internal/cmd/provenance/` CLI (Sprint 3)
- Touch `internal/usersync/collision.go` (Sprint 2)
- Create new packages
- Modify KNOSSOS_MANIFEST.yaml format
- Add any Phase 4b content

---

### Sprint 2: Usersync Migration

**Goal**: Migrate usersync from its own manifest types to `provenance.ProvenanceEntry`. Three resource syncers write to single `USER_PROVENANCE_MANIFEST.yaml`.

**Tasks**:

1. Remove `usersync.SourceType`, `usersync.Entry`, `usersync.Manifest`, `usersync.manifestJSON`, `usersync.entryJSON` types from `usersync/manifest.go`
2. Remove `loadManifest()` and `saveManifest()` methods from `usersync/manifest.go` (replaced by `provenance.LoadOrBootstrap` and `provenance.Save`)
3. Remove `LoadManifest()` and `SaveManifest()` public convenience functions
4. Update `Syncer` struct: all resource types use `paths.UserProvenanceManifest()` for `manifestPath`
5. Update `NewSyncer()`: set `s.manifestPath = paths.UserProvenanceManifest()` for all resource types
6. Update `Syncer.Sync()` method: call `provenance.LoadOrBootstrap()` and `provenance.Save()` instead of `s.loadManifest()` / `s.saveManifest()`
7. Update `syncFiles()`: replace all `Entry` usage with `*provenance.ProvenanceEntry`, add manifest key prefixing, eliminate `SourceDiverged` -- use computed divergence
8. Update `recover()` and `recoverDir()`: use `provenance.ProvenanceEntry` types
9. Update `cleanupOldManifests()`: remove all 5 legacy JSON manifests with `.v2-backup`, call for all resource types
10. Add orphan detection to `syncFiles()` per Section 7
11. Update `CollisionChecker` per Section 5 (manifest-based with fallback)
12. Update `NewSyncerWithPaths()` and `NewMenaSyncerWithPaths()` test helpers
13. Update `usersync/usersync_test.go`: all tests use provenance types, verify YAML manifest
14. Remove `usersync/manifest_test.go` tests for removed types; add new tests for provenance round-trip
15. Delete deprecated path functions: `UserAgentManifest()`, `UserMenaManifest()`, `UserHooksManifest()`, `UserSkillManifest()`, `UserCommandManifest()`

**Verification**:
- `CGO_ENABLED=0 go build ./cmd/ari` succeeds
- `CGO_ENABLED=0 go vet ./...` clean
- `CGO_ENABLED=0 go test ./internal/usersync/...` passes
- `CGO_ENABLED=0 go test ./internal/provenance/...` passes
- `CGO_ENABLED=0 go test ./...` full suite passes
- `ari sync user all` creates `USER_PROVENANCE_MANIFEST.yaml` with all three resource types
- Old JSON manifests are backed up and removed
- `ari sync user agents && ari sync user mena && ari sync user hooks` accumulates entries in single manifest
- Sequential sync-all does not lose entries from prior resource types
- Orphan removal works for knossos-owned entries whose source was deleted

**DO NOT**:
- Touch `internal/cmd/provenance/` CLI (Sprint 3)
- Touch `internal/materialize/` (done in Sprint 1)
- Change the provenance package schema (done in Sprint 1)
- Create new packages
- Modify KNOSSOS_MANIFEST.yaml format

---

### Sprint 3: CLI Unification

**Goal**: `ari provenance show` displays both rite-scope and user-scope entries with scope column and filtering.

**Tasks**:

1. Add `Scope` field to `ShowEntry` struct
2. Update `ShowOutput.Headers()` to include `SCOPE` column
3. Update `ShowOutput.Rows()` to include scope value
4. Add `--scope` flag to `newShowCmd()` (values: `rite`, `user`; default: both)
5. Update `runShow()` to load both manifests: rite-scope from project `.claude/`, user-scope from `~/.claude/`
6. Add `~/` prefix to user-scope entry paths in table output
7. Add `CombinedOutput` struct for JSON/YAML output
8. Update JSON/YAML output path to use `CombinedOutput`
9. Update `computeStatus()` to handle `OwnerUntracked` (same behavior as `OwnerUser`: return `"-"`)
10. Update command help text to document new `--scope` flag and SCOPE column

**Verification**:
- `CGO_ENABLED=0 go build ./cmd/ari` succeeds
- `CGO_ENABLED=0 go vet ./...` clean
- `CGO_ENABLED=0 go test ./...` full suite passes
- `ari provenance show` displays both scopes with SCOPE column
- `ari provenance show --scope=rite` shows only rite entries
- `ari provenance show --scope=user` shows only user entries
- `ari provenance show -o json` outputs combined JSON with rite/user sections
- User entries display with `~/` prefix

**DO NOT**:
- Touch `internal/provenance/` types (done in Sprint 1)
- Touch `internal/usersync/` sync logic (done in Sprint 2)
- Create new packages
- Modify KNOSSOS_MANIFEST.yaml format
- Add Phase 4b content

---

## Section 9: Test Plan

### 9.1 Schema Validation Tests

| Test | File | Expected Outcome |
|------|------|------------------|
| `TestOwnerTypeIsValid_Untracked` | `provenance_test.go` | `OwnerUntracked.IsValid()` returns true |
| `TestOwnerTypeIsValid_UnknownRejected` | `provenance_test.go` | `OwnerType("unknown").IsValid()` returns false |
| `TestScopeTypeIsValid` | `provenance_test.go` | `ScopeRite` and `ScopeUser` return true; `ScopeType("invalid")` returns false |
| `TestValidateManifest_ScopeRequired` | `provenance_test.go` | Entry without Scope field fails validation |
| `TestValidateManifest_ScopeInvalid` | `provenance_test.go` | Entry with `Scope: "materialize"` fails validation |
| `TestValidateManifest_NoSourcePipelineField` | `provenance_test.go` | Entry without SourcePipeline (gone from struct) passes validation |

### 9.2 Migration Tests

| Test | File | Expected Outcome |
|------|------|------------------|
| `TestMigrateV1ToV2_OwnerUnknown` | `provenance_test.go` | v1.0 manifest with `owner: unknown` loads as `owner: untracked` |
| `TestMigrateV1ToV2_SourcePipeline` | `provenance_test.go` | v1.0 manifest with `source_pipeline: materialize` loads with `scope: rite` |
| `TestMigrateV1ToV2_RoundTrip` | `provenance_test.go` | Load v1.0, save, load again -- produces valid v2.0 manifest |
| `TestBootstrap_V2Version` | `provenance_test.go` | `LoadOrBootstrap` on missing file returns `SchemaVersion: "2.0"` |

### 9.3 Rite-Scope Pipeline Tests

| Test | File | Expected Outcome |
|------|------|------------------|
| `TestProvenanceIntegration_BasicMaterialization` | `provenance_integration_test.go` | All entries have `Scope: ScopeRite` (was `SourcePipeline: "materialize"`) |
| `TestProvenanceIntegration_DivergenceDetection` | `provenance_integration_test.go` | Divergence detection works with v2.0 entries |
| `TestProvenanceIntegration_Idempotency` | `provenance_integration_test.go` | Two materializations produce structurally equal manifests |

### 9.4 User-Scope Pipeline Tests

| Test | File | Expected Outcome |
|------|------|------------------|
| `TestSyncer_AddNew_ProvenanceEntry` | `usersync_test.go` | New files produce `ProvenanceEntry` with `Scope: ScopeUser`, `Owner: OwnerKnossos` |
| `TestSyncer_UserCreated_ProvenanceEntry` | `usersync_test.go` | User-created files produce `Owner: OwnerUser`, `Scope: ScopeUser` |
| `TestSyncer_ComputedDivergence` | `usersync_test.go` | Locally modified knossos file detected via checksum comparison (no `SourceDiverged`) |
| `TestSyncer_ForceOverwriteDiverged` | `usersync_test.go` | `--force` overwrites diverged file and resets to `Owner: OwnerKnossos` |
| `TestSyncer_ManifestKeyPrefixed` | `usersync_test.go` | Agents entries keyed as `agents/name.md`, hooks as `hooks/path` |
| `TestSyncer_SharedManifest` | `usersync_test.go` | Sequential sync of agents then mena produces manifest with both `agents/` and `commands/` entries |
| `TestSyncer_OrphanRemoval` | `usersync_test.go` | Knossos-owned entry whose source was deleted is removed from target and manifest |
| `TestSyncer_OrphanPreservesUser` | `usersync_test.go` | User-owned entry whose source was deleted is NOT removed |

### 9.5 Migration Tests (JSON to YAML)

| Test | File | Expected Outcome |
|------|------|------------------|
| `TestMigration_OldJsonBackedUp` | `usersync_test.go` | Old JSON manifests get `.v2-backup` suffix |
| `TestMigration_FreshResync` | `usersync_test.go` | Empty YAML manifest triggers full resync (all files treated as new) |
| `TestMigration_NoDataLoss` | `usersync_test.go` | All resources present in target dir after migration sync |

### 9.6 CLI Tests

| Test | File | Expected Outcome |
|------|------|------------------|
| `TestShowOutput_ScopeColumn` | Manual / integration | Table output has SCOPE column |
| `TestShowOutput_ScopeFilter_Rite` | Manual / integration | `--scope=rite` shows only rite entries |
| `TestShowOutput_ScopeFilter_User` | Manual / integration | `--scope=user` shows only user entries |
| `TestShowOutput_UserPrefix` | Manual / integration | User entries display with `~/` prefix |
| `TestShowOutput_JsonCombined` | Manual / integration | JSON output has `rite` and `user` sections |

### 9.7 Collision Detection Tests

| Test | File | Expected Outcome |
|------|------|------------------|
| `TestCollisionChecker_ManifestBased` | `usersync_test.go` | Collision detected via manifest lookup (no directory scan) |
| `TestCollisionChecker_FallbackToScan` | `usersync_test.go` | When manifest missing, falls back to directory scan |
| `TestCollisionChecker_NoClaudeDir` | `usersync_test.go` | Passing empty claudeDir triggers directory scan fallback |

### 9.8 Build Verification

| Check | Command | Expected |
|-------|---------|----------|
| Build | `CGO_ENABLED=0 go build ./cmd/ari` | Success |
| Vet | `CGO_ENABLED=0 go vet ./...` | Clean |
| Full test suite | `CGO_ENABLED=0 go test ./...` | All pass |

---

## Section 10: Risk Register

| Risk | Severity | Mitigation |
|------|----------|------------|
| **CC file watcher crash**: Writing USER_PROVENANCE_MANIFEST.yaml triggers CC's watcher in `~/.claude/` | HIGH | `provenance.Save()` already uses `structurallyEqual()` to skip writes when only timestamps change. This protection applies to both rite and user manifests |
| **Write contention**: Multiple `ari sync user` commands running in parallel write to the same YAML file | MEDIUM | Sequential `for` loop in `user_all.go` is preserved. Document that parallel `ari sync user agents` and `ari sync user mena` in separate terminals is unsupported. Consider file locking in future work |
| **Scope creep**: Integration engineer adds CLI features in Sprint 1 or schema changes in Sprint 2 | MEDIUM | Explicit DO NOT lists per sprint. Each sprint has verification that out-of-scope files are untouched |
| **v1.0 manifest in the wild**: Existing projects have v1.0 PROVENANCE_MANIFEST.yaml that must continue to load | HIGH | `migrateV1ToV2()` function handles in-memory upgrade. Migration is lossless (SourcePipeline was always "materialize"). Covered by `TestMigrateV1ToV2_RoundTrip` |
| **Orphan false positive**: Mena directory-level tracking across user and rite scopes could conflict on key names like `commands/commit/` | LOW | Keys are only compared within their own scope's manifest. Rite manifest has `commands/commit/` in project `.claude/`; user manifest has `commands/commit/` in `~/.claude/`. Different files, different manifests, no conflict |
| **Test fixture breakage**: Many existing tests reference `SourcePipeline: "materialize"` and `OwnerUnknown` | MEDIUM | Sprint 1 task list explicitly includes updating all test files. Verification step runs full test suite |
| **CollisionChecker performance regression**: Loading and parsing manifest YAML on every collision check | LOW | Manifest is loaded once per `NewCollisionChecker()` call and cached in `riteEntries` map. The map lookup is O(1) per collision check vs O(n rites * filesystem stat) in the old approach |
| **cleanupOldManifests runs on every resource type**: Could attempt to backup+remove already-removed files | LOW | The function handles missing files gracefully (`os.ReadFile` returns error, `continue` skips). Multiple cleanup attempts on the same nonexistent file are harmless |

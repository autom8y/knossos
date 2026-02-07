# TDD: Mena Scope Initiative -- PR1 (Mechanical Refactoring)

## Overview

PR1 delivers the mechanical infrastructure for the Mena Scope Initiative: a unified `ProjectMena()` API in `internal/materialize/`, extension stripping for `.dro`/`.lego` infixes, a merged `USER_MENA_MANIFEST.json` schema, a consolidated `ResourceMena` type replacing `ResourceSkills` + `ResourceCommands`, CLI unification (`ari sync user mena`), and source tree directory renames (`user-agents/` -> `agents/`, `user-hooks/ari/*` -> `hooks/*`). This is the "plumbing" PR that makes the scope field (PR2) possible.

This PR implements binding decisions D1, D5, D6, D7, D8, D9, D10, and D12 from `MENA_SCOPE_DECISIONS.md`. Scope field (D3), advisory enforcement (D4), and scope annotations (D5 Step 5) are deferred to PR2.

---

## 1. ProjectMena API

### 1.1 Package Location

Per D8: `internal/materialize/`. No new package. The function lives in a new file `internal/materialize/project_mena.go` to keep the already-large `materialize.go` from growing further.

### 1.2 Type Definitions

```go
// MenaProjectionMode controls whether projection is additive or destructive.
type MenaProjectionMode int

const (
    // MenaProjectionAdditive adds/updates files without removing unmanaged content.
    // Used by usersync (ari sync user mena).
    MenaProjectionAdditive MenaProjectionMode = iota

    // MenaProjectionDestructive wipes target commands/ and skills/ directories
    // before projecting. Used by materialize (ari rite start).
    MenaProjectionDestructive
)

// MenaFilter controls which mena types to project.
type MenaFilter int

const (
    ProjectDro  MenaFilter = 1 << iota // Project dromena only (commands/)
    ProjectLego                        // Project legomena only (skills/)
    ProjectAll  = ProjectDro | ProjectLego
)

// MenaProjectionOptions configures the projection operation.
type MenaProjectionOptions struct {
    Mode   MenaProjectionMode
    Filter MenaFilter

    // TargetCommandsDir is the absolute path to the commands/ output directory.
    // For materialize: <project>/.claude/commands/
    // For usersync:    ~/.claude/commands/
    TargetCommandsDir string

    // TargetSkillsDir is the absolute path to the skills/ output directory.
    // For materialize: <project>/.claude/skills/
    // For usersync:    ~/.claude/skills/
    TargetSkillsDir string
}

// MenaProjectionResult reports what the projection did.
type MenaProjectionResult struct {
    CommandsProjected []string // Relative paths of files written to commands/
    SkillsProjected   []string // Relative paths of files written to skills/
}
```

### 1.3 Function Signatures

The primary exported API:

```go
// ProjectMena projects mena source files into commands/ and skills/ target
// directories. It handles extension stripping, mena type routing, and supports
// both filesystem and embedded FS sources.
//
// Sources are processed in priority order (later overrides earlier):
//   1. Distribution-level mena/ (from knossosHome or projectRoot)
//   2. rites/shared/mena/
//   3. rites/{dependency}/mena/ (in manifest dependency order)
//   4. rites/{active}/mena/ (highest priority)
//
// In Additive mode, existing files in target directories are preserved.
// In Destructive mode, target directories are wiped before projection.
func ProjectMena(sources []MenaSource, opts MenaProjectionOptions) (*MenaProjectionResult, error)
```

Where `MenaSource` is the existing `menaSource` type, promoted to exported:

```go
// MenaSource represents a source for mena files. It can be either a
// filesystem path or an embedded FS path.
type MenaSource struct {
    Path       string // Filesystem path (for os-based sources)
    Fsys       fs.FS  // Embedded filesystem (nil for os-based sources)
    FsysPath   string // Path within Fsys (e.g., "rites/shared/mena")
    IsEmbedded bool
}
```

### 1.4 Additive Mode (usersync caller)

When `Mode == MenaProjectionAdditive`:

1. Do NOT wipe target directories.
2. Walk all sources in priority order, collect entries (same as current `materializeMena()` Pass 1).
3. For each collected entry, strip extensions (Section 2), determine target dir.
4. For each file to be written, check if the target file already exists:
   - If it does not exist: write it.
   - If it exists: write it (overwrite). The usersync manifest tracking and divergence detection happen in the *caller* (`Syncer.syncFiles()`), not in `ProjectMena()`. The projection function is pure I/O; the caller handles policy.

The usersync `Syncer` continues to own manifest loading, checksum comparison, divergence detection, and collision checking. It calls `ProjectMena()` only for the actual file discovery and copying. The integration works as follows:

```go
// In usersync's Syncer.syncFiles() for ResourceMena:
// 1. Syncer discovers source files from mena/ directory
// 2. For each file, Syncer computes checksum, checks manifest, applies policy
// 3. When Syncer decides to write a file, it calls copyFileWithStripping()
//    which uses StripMenaExtension() to determine the output filename
// 4. Manifest keys use the STRIPPED filename (e.g., "commit/INDEX.md" not "commit/INDEX.dro.md")
```

Actually, given the complexity of integrating manifest tracking with extension stripping, the cleaner design is to split the shared logic into two layers:

**Layer 1: Extension stripping + routing (shared)**
```go
// StripMenaExtension removes .dro or .lego infix from a filename.
// "INDEX.dro.md" -> "INDEX.md", "helper.md" -> "helper.md" (unchanged).
// Only strips from entry-point files (INDEX* and standalone .dro.md/.lego.md).
func StripMenaExtension(filename string) string

// RouteMenaFile determines whether a file routes to commands/ or skills/.
// Returns "commands" or "skills".
func RouteMenaFile(filename string) string
```

**Layer 2: Full projection (materialize caller only)**
```go
// ProjectMena performs full mena projection with multi-source priority
// resolution. This is the materialize-side entry point.
func ProjectMena(sources []MenaSource, opts MenaProjectionOptions) (*MenaProjectionResult, error)
```

The usersync caller uses Layer 1 functions directly within its existing `syncFiles()` walk, applying extension stripping to filenames while maintaining its manifest/checksum/divergence tracking. The materialize caller uses Layer 2, which internally uses Layer 1.

### 1.5 Destructive Mode (materialize caller)

When `Mode == MenaProjectionDestructive`:

1. `os.RemoveAll(opts.TargetCommandsDir)` then `os.MkdirAll(...)`.
2. `os.RemoveAll(opts.TargetSkillsDir)` then `os.MkdirAll(...)`.
3. Full multi-source collection with priority resolution (identical to current `materializeMena()` Pass 1).
4. Route each entry by mena type, strip extensions, copy.

This is the existing `materializeMena()` behavior, refactored to use the shared extension stripping and routing logic.

### 1.6 Extension Stripping Integration

Both the `copyDirFromFS()` and filesystem copy paths must apply `StripMenaExtension()` to each filename during copy. The implementation detail:

```go
// copyDirWithStripping copies all files from src to dst, applying
// StripMenaExtension to filenames of entry-point files.
func copyDirWithStripping(src, dst string) error {
    return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        relPath, _ := filepath.Rel(src, path)
        // Strip extension from the filename component
        dir := filepath.Dir(relPath)
        base := StripMenaExtension(filepath.Base(relPath))
        strippedRel := filepath.Join(dir, base)
        destPath := filepath.Join(dst, strippedRel)

        if d.IsDir() {
            return os.MkdirAll(destPath, 0755)
        }
        content, _ := os.ReadFile(path)
        os.MkdirAll(filepath.Dir(destPath), 0755)
        return os.WriteFile(destPath, content, 0644)
    })
}
```

An equivalent `copyDirFromFSWithStripping()` handles `fs.FS` sources.

### 1.7 Embedded FS Support

The existing `materializeMena()` already supports `embed.FS` via the `menaSource.isEmbedded` flag and `collectMenaEntriesFS()`. The refactored `ProjectMena()` preserves this. The `MenaSource` struct carries either a filesystem path or an `fs.FS` + path-within-FS, exactly as the current `menaSource` does.

Usersync does NOT use `embed.FS` -- it always reads from filesystem (`$KNOSSOS_HOME/mena/`). The `MenaSource.IsEmbedded` field will always be `false` for usersync callers.

### 1.8 Refactoring `materializeMena()`

The existing `materializeMena()` method on `Materializer` is refactored to:

1. Build the `[]MenaSource` list (same priority logic it has today).
2. Build `MenaProjectionOptions` with `MenaProjectionDestructive` mode.
3. Call `ProjectMena(sources, opts)`.
4. Return the error.

The method signature remains identical -- callers in `MaterializeWithOptions()` see no change.

---

## 2. Extension Stripping Rules

### 2.1 What Gets Stripped

The `.dro` and `.lego` infixes are stripped from **all** filenames during projection. The stripping function:

```go
// StripMenaExtension removes the .dro or .lego infix from a filename.
// Examples:
//   "INDEX.dro.md"      -> "INDEX.md"
//   "INDEX.lego.md"     -> "INDEX.md"
//   "commit.dro.md"     -> "commit.md"
//   "prompting.lego.md" -> "prompting.md"
//   "helper.md"         -> "helper.md"    (no infix, unchanged)
//   "README.md"         -> "README.md"    (no infix, unchanged)
//   "data.json"         -> "data.json"    (no infix, unchanged)
func StripMenaExtension(filename string) string {
    if strings.Contains(filename, ".dro.") {
        return strings.Replace(filename, ".dro.", ".", 1)
    }
    if strings.Contains(filename, ".lego.") {
        return strings.Replace(filename, ".lego.", ".", 1)
    }
    return filename
}
```

The logic uses `strings.Replace` with count=1 to handle the (pathological) case of `foo.dro.dro.md` -- strips only the first infix.

### 2.2 What Doesn't Get Stripped

**No exception for supporting files.** The original task description suggested only stripping entry-point files (INDEX + standalone). After analysis, this creates unnecessary complexity:

- Supporting files like `helper.md` already have no `.dro`/`.lego` infix to strip.
- The only files that have `.dro`/`.lego` infixes are INDEX files and standalone mena files.
- A supporting file named `detail.dro.md` inside a leaf directory would be unusual but should still be stripped for consistency -- the infix is a routing signal, not a semantic part of the filename.

Therefore: strip ALL files unconditionally. The function is idempotent on files without the infix.

### 2.3 Collision Handling

**Within a single mena source directory:**
- If `foo/INDEX.dro.md` and `foo/INDEX.lego.md` both exist in the same leaf directory, this is a malformed source. Both would strip to `INDEX.md` but route to different target directories (`commands/foo/INDEX.md` and `skills/foo/INDEX.md`). The `DetectMenaType()` function reads the INDEX file to determine routing. Since only ONE routing decision is made per leaf directory (based on whichever INDEX file is found first), this is implicitly handled -- but it should be detected and warned about.

  **Action**: Add a validation check in the collection pass. If a leaf directory contains both `INDEX.dro.md` and `INDEX.lego.md`, log a warning and use the first one found (deterministic via `os.ReadDir` alphabetical order, which means `.dro` wins over `.lego`).

**Across different target directories:**
- `foo.dro.md` strips to `foo.md` in `commands/`; `foo.lego.md` strips to `foo.md` in `skills/`. These are different directories, so no collision occurs. This is correct behavior.

**Across priority sources:**
- If `mena/commit/INDEX.dro.md` exists and `rites/10x-dev/mena/commit/INDEX.dro.md` also exists, the higher-priority source (rite) overrides the lower (distribution). This is existing behavior, unchanged.

---

## 3. USER_MENA_MANIFEST.json

### 3.1 Schema

The new unified manifest replaces both `USER_COMMAND_MANIFEST.json` and `USER_SKILL_MANIFEST.json`. It also absorbs what was `USER_SKILL_MANIFEST.json` (which tracked skills from `user-skills/`).

**File location**: `~/.claude/USER_MENA_MANIFEST.json`

**JSON schema**:

```json
{
  "manifest_version": "2.0",
  "last_sync": "2026-02-07T12:00:00Z",
  "mena": {
    "commit/INDEX.md": {
      "source": "knossos",
      "installed_at": "2026-02-07T12:00:00Z",
      "checksum": "sha256:abc123...",
      "mena_type": "dro",
      "target": "commands"
    },
    "prompting/INDEX.md": {
      "source": "knossos",
      "installed_at": "2026-02-07T12:00:00Z",
      "checksum": "sha256:def456...",
      "mena_type": "lego",
      "target": "skills"
    }
  }
}
```

**Go types**:

```go
// MenaManifestVersion is the schema version for the unified mena manifest.
const MenaManifestVersion = "2.0"

// menaManifestJSON is the on-disk format for USER_MENA_MANIFEST.json.
type menaManifestJSON struct {
    Version  string                    `json:"manifest_version"`
    LastSync string                    `json:"last_sync"`
    Mena     map[string]menaEntryJSON  `json:"mena"`
}

// menaEntryJSON represents a single entry in the mena manifest.
type menaEntryJSON struct {
    Source      string `json:"source"`       // "knossos", "knossos-diverged", "user"
    InstalledAt string `json:"installed_at"` // RFC3339 timestamp
    Checksum    string `json:"checksum"`     // SHA-256 of source file
    MenaType    string `json:"mena_type"`    // "dro" or "lego"
    Target      string `json:"target"`       // "commands" or "skills"
}
```

**Key design decisions**:

1. **Manifest keys use stripped filenames.** The key `"commit/INDEX.md"` (not `"commit/INDEX.dro.md"`) because the manifest tracks the *projected* state, not the source state. The `mena_type` field preserves the original routing information.

2. **`target` field** records which output directory (`commands` or `skills`) the file was written to. This enables the sync system to know where to check for divergence without re-parsing the source.

3. **Version `"2.0"`** -- a major version bump signals incompatibility with `"1.0"` manifests.

### 3.2 Migration Strategy

Per D10: **wipe-and-resync, no migration logic.**

When the `Syncer` loads a manifest:

```go
func (s *Syncer) loadManifest() (*Manifest, error) {
    data, err := os.ReadFile(s.manifestPath)
    if err != nil {
        if os.IsNotExist(err) {
            return newEmptyManifest(), nil
        }
        return nil, ErrManifestRead(s.manifestPath, err)
    }

    // Quick version check before full parse
    var versionCheck struct {
        Version string `json:"manifest_version"`
    }
    if err := json.Unmarshal(data, &versionCheck); err != nil {
        // Corrupt -- backup and start fresh
        return s.backupAndCreateFresh(data)
    }

    if versionCheck.Version != MenaManifestVersion {
        // Version mismatch -- wipe and start fresh
        return s.backupAndCreateFresh(data)
    }

    // Parse full manifest...
}

func (s *Syncer) backupAndCreateFresh(oldData []byte) (*Manifest, error) {
    backupPath := s.manifestPath + ".v1-backup"
    os.WriteFile(backupPath, oldData, 0644) // Best effort backup
    return newEmptyManifest(), nil
}
```

**Old manifest cleanup**: After a successful sync with the new manifest, delete the old manifest files:

```go
// In the sync operation, after successful save of USER_MENA_MANIFEST.json:
func (s *Syncer) cleanupOldManifests() {
    homeDir, _ := os.UserHomeDir()
    oldManifests := []string{
        filepath.Join(homeDir, ".claude", "USER_COMMAND_MANIFEST.json"),
        filepath.Join(homeDir, ".claude", "USER_SKILL_MANIFEST.json"),
    }
    for _, path := range oldManifests {
        os.Remove(path) // Ignore errors -- they may not exist
    }
}
```

### 3.3 Version Mismatch Behavior

1. Read `manifest_version` field from existing manifest.
2. If version != `"2.0"`: backup old manifest to `*.v1-backup`, create fresh empty manifest, proceed with full sync (all files will be "new" and get written).
3. If version == `"2.0"`: normal operation.
4. If manifest does not exist: create new empty manifest with version `"2.0"`.
5. If manifest is corrupt (JSON parse error): backup to `*.corrupt`, create fresh.

This means the first `ari sync user mena` after upgrade effectively re-syncs everything. This is acceptable per D10 rationale: checksums and paths are all invalidated by the rename.

---

## 4. ResourceMena Type

### 4.1 Type Definition

In `internal/usersync/usersync.go`:

```go
const (
    ResourceAgents ResourceType = "agents"
    ResourceMena   ResourceType = "mena"    // Replaces ResourceSkills + ResourceCommands
    ResourceHooks  ResourceType = "hooks"
)
```

`ResourceSkills` and `ResourceCommands` are **removed entirely**. Any code referencing them is updated or deleted.

### 4.2 Method Changes

```go
// Singular returns the singular form of the resource type.
func (r ResourceType) Singular() string {
    switch r {
    case ResourceAgents:
        return "agent"
    case ResourceMena:
        return "mena"
    case ResourceHooks:
        return "hook"
    default:
        return string(r)
    }
}

// SourceDir returns the source directory name for the resource type.
func (r ResourceType) SourceDir() string {
    return string(r) // "agents", "mena", "hooks" -- all direct names now
}

// RiteSubDir returns the subdirectory name within rites for the resource type.
func (r ResourceType) RiteSubDir() string {
    return string(r) // "agents", "mena", "hooks" -- same
}
```

Note: `SourceDir()` previously returned `"user-" + string(r)` for agents, skills, hooks. Now ALL resource types use their bare name as the source directory, because the `user-` prefix directories are being renamed:
- `user-agents/` -> `agents/`
- `user-hooks/` -> `hooks/`  (with flattening, see Section 6)
- `user-skills/` -> merged into `mena/`

### 4.3 NewSyncer Changes

The `NewSyncer()` function for `ResourceMena` sets up TWO target directories (commands/ + skills/) instead of one:

```go
case ResourceMena:
    s.sourceDir = filepath.Join(knossosHome, "mena")
    s.targetCommandsDir = filepath.Join(homeDir, ".claude", "commands")
    s.targetSkillsDir = filepath.Join(homeDir, ".claude", "skills")
    s.manifestPath = filepath.Join(homeDir, ".claude", "USER_MENA_MANIFEST.json")
    s.nested = true
```

The `Syncer` struct gains two new fields:

```go
type Syncer struct {
    resourceType     ResourceType
    sourceDir        string
    targetDir        string          // Used by agents, hooks (single target)
    targetCommandsDir string         // Used by mena (dromena target)
    targetSkillsDir   string         // Used by mena (legomena target)
    manifestPath     string
    collisionChecker *CollisionChecker
    nested           bool
}
```

For mena, `targetDir` is left empty. The `syncFiles()` method checks `resourceType == ResourceMena` to use the dual-target logic with routing via `RouteMenaFile()`.

### 4.4 Impact on Collision Detection

The `CollisionChecker` for `ResourceMena` checks against `rites/*/mena/` (one directory), not against both `rites/*/commands/` and `rites/*/skills/` separately. This is correct because rite mena sources are always in `mena/`, never in separate `commands/`/`skills/` directories at the source level.

```go
// For ResourceMena:
//   RiteSubDir() returns "mena"
//   Collision check: rites/{riteName}/mena/{name}
```

No cross-type collision checking is needed.

---

## 5. CLI Changes

### 5.1 New: ari sync user mena

New file: `internal/cmd/sync/user_mena.go`

```go
func newUserMenaCmd(ctx *cmdContext) *cobra.Command {
    var dryRun, recover, force, verbose bool

    cmd := &cobra.Command{
        Use:   "mena",
        Short: "Sync user mena (commands + skills) to ~/.claude/",
        Long: `Sync mena files from knossos mena/ to ~/.claude/commands/ and ~/.claude/skills/.

Mena files are routed by their source extension:
  .dro.md  -> ~/.claude/commands/ (dromena: invokable commands)
  .lego.md -> ~/.claude/skills/   (legomena: reference knowledge)

Extensions are stripped during projection:
  INDEX.dro.md  -> INDEX.md in commands/
  INDEX.lego.md -> INDEX.md in skills/

Behavior:
  - Routes files to commands/ or skills/ based on mena type
  - Strips .dro/.lego extensions from projected filenames
  - Preserves directory structure (progressive disclosure)
  - Only updates when source changes (checksum-based)
  - Preserves user-created content (never deleted)
  - Skips resources that would shadow rite mena

Examples:
  ari sync user mena
  ari sync user mena --dry-run
  ari sync user mena --recover
  ari sync user mena --force`,
        SilenceUsage: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            opts := usersync.Options{
                DryRun:  dryRun,
                Recover: recover,
                Force:   force,
                Verbose: verbose,
            }
            return runUserSync(ctx, usersync.ResourceMena, opts)
        },
    }

    cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
    cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching knossos")
    cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
    cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

    common.SetNeedsProject(cmd, false, false)
    return cmd
}
```

### 5.2 Removed Commands

Delete these files entirely:
- `internal/cmd/sync/user_commands.go`
- `internal/cmd/sync/user_skills.go`

### 5.3 Updated: ari sync user all

In `internal/cmd/sync/user_all.go`:

```go
// Before:
resourceTypes := []usersync.ResourceType{
    usersync.ResourceAgents,
    usersync.ResourceSkills,
    usersync.ResourceCommands,
    usersync.ResourceHooks,
}

// After:
resourceTypes := []usersync.ResourceType{
    usersync.ResourceAgents,
    usersync.ResourceMena,
    usersync.ResourceHooks,
}
```

The `Long` description is updated to reflect 3 resource types:

```
Runs sync for all resource types in sequence:
  1. agents
  2. mena (commands + skills)
  3. hooks
```

### 5.4 Updated: ari sync user (parent)

In `internal/cmd/sync/user.go`:

```go
func newUserCmd(ctx *cmdContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "user",
        Short: "Sync user-level resources to ~/.claude/",
        Long: `Sync user-level resources from knossos to ~/.claude/.

User resources are globally available across all projects.
They are stored in ~/.claude/ and synced from $KNOSSOS_HOME/{type}/.

Resources:
  agents  - Agent prompts (agents/ -> ~/.claude/agents/)
  mena    - Commands and skills (mena/ -> ~/.claude/commands/ + skills/)
  hooks   - Hook scripts (hooks/ -> ~/.claude/hooks/)

Sync Behavior:
  - Additive: Never removes user-created content
  - Checksum-based: Only updates when source changes
  - Collision-aware: Skips resources that would shadow rite resources

Source Types:
  - knossos          Synced from knossos, checksums match
  - knossos-diverged Originally from knossos but locally modified
  - user             Created by user, not from knossos`,
    }

    cmd.AddCommand(newUserAgentsCmd(ctx))
    cmd.AddCommand(newUserMenaCmd(ctx))
    cmd.AddCommand(newUserHooksCmd(ctx))
    cmd.AddCommand(newUserAllCmd(ctx))

    common.SetNeedsProject(cmd, false, false)
    return cmd
}
```

### 5.5 Updated: AllResult rendering

In `internal/usersync/output.go`:

```go
// Before:
order := []ResourceType{ResourceAgents, ResourceSkills, ResourceCommands, ResourceHooks}

// After:
order := []ResourceType{ResourceAgents, ResourceMena, ResourceHooks}
```

The `Mena` line in output renders as:

```
Mena: 12 added, 3 updated, 1 skipped, 8 unchanged
```

---

## 6. Directory Renames

### 6.1 Source Tree Changes

All renames use `git mv` to preserve history.

| Before | After | Notes |
|--------|-------|-------|
| `user-agents/` | `agents/` | Direct rename. 3 files (consultant.md, context-engineer.md, moirai.md). Backup file `moirai.md.backup` should be removed first. |
| `user-hooks/ari/*` | `hooks/*` | Flatten: remove the `ari/` nesting level. Files: `autopark.sh`, `clew.sh`, `cognitive-budget.sh`, `context.sh`, `hooks.yaml`, `route.sh`, `validate.sh`, `writeguard.sh`. |
| `user-skills/` | (merged into `mena/`) | Per ADR-0021 this is already done at the mena level. Verify no remaining content in `user-skills/`. If the directory still exists and contains files, move them to `mena/` with `.lego.md` extension. |

**Rename sequence** (shell):
```bash
# 1. Remove stale backup
rm user-agents/moirai.md.backup

# 2. Agents
git mv user-agents agents

# 3. Hooks: flatten user-hooks/ari/* -> hooks/
mkdir -p hooks
git mv user-hooks/ari/autopark.sh hooks/
git mv user-hooks/ari/clew.sh hooks/
git mv user-hooks/ari/cognitive-budget.sh hooks/
git mv user-hooks/ari/context.sh hooks/
git mv user-hooks/ari/hooks.yaml hooks/
git mv user-hooks/ari/route.sh hooks/
git mv user-hooks/ari/validate.sh hooks/
git mv user-hooks/ari/writeguard.sh hooks/
# Remove empty directory
rm -rf user-hooks

# 4. user-skills (verify empty or merge into mena/)
rm -rf user-skills  # If empty
```

### 6.2 embed.go Changes

In `/Users/tomtenuta/Code/knossos/embed.go`:

**Before**:
```go
//go:embed user-hooks/ari/hooks.yaml
var EmbeddedHooksYAML []byte
```

**After**:
```go
//go:embed hooks
var EmbeddedHooks embed.FS
```

The type changes from `[]byte` to `embed.FS` because the hooks directory may contain more than `hooks.yaml` in the future (per task description). This is a forward-looking change.

**Impact on `Materializer`**: The `embeddedHooksYAML []byte` field on `Materializer` becomes `embeddedHooks fs.FS`:

```go
type Materializer struct {
    resolver          *paths.Resolver
    sourceResolver    *SourceResolver
    explicitSource    string
    ritesDir          string
    templatesDir      string
    embeddedTemplates fs.FS
    embeddedHooks     fs.FS  // Changed from embeddedHooksYAML []byte
}
```

The `WithEmbeddedHooks` method changes signature:
```go
// Before:
func (m *Materializer) WithEmbeddedHooks(data []byte) *Materializer

// After:
func (m *Materializer) WithEmbeddedHooks(fsys fs.FS) *Materializer
```

The `loadHooksConfig()` method in `hooks.go` changes its embedded fallback:

```go
// Before:
if m.embeddedHooksYAML != nil {
    var cfg HooksConfig
    if err := yaml.Unmarshal(m.embeddedHooksYAML, &cfg); err == nil {
        ...
    }
}

// After:
if m.embeddedHooks != nil {
    data, err := fs.ReadFile(m.embeddedHooks, "hooks.yaml")
    if err == nil {
        var cfg HooksConfig
        if err := yaml.Unmarshal(data, &cfg); err == nil {
            ...
        }
    }
}
```

### 6.3 Path Reference Updates

Every file that references the old directory names must be updated. Complete list based on code analysis:

| File | Line(s) | Change |
|------|---------|--------|
| `internal/materialize/hooks.go` | 34-35, 41, 45 | `user-hooks/ari/hooks.yaml` -> `hooks/hooks.yaml` |
| `internal/usersync/usersync.go` | 134 | `"user-agents"` -> `"agents"` |
| `internal/usersync/usersync.go` | 139 | `"user-skills"` -> DELETED (ResourceSkills removed) |
| `internal/usersync/usersync.go` | 149 | `"user-hooks"` -> `"hooks"` |
| `internal/usersync/usersync.go` | 44 | `"user-" + string(r)` -> `string(r)` |
| `internal/cmd/sync/user.go` | 19-22 | Help text: remove `user-` prefixes |
| `internal/cmd/sync/user_agents.go` | 16 | `"user-agents/"` -> `"agents/"` in help text |
| `internal/cmd/sync/user_hooks.go` | 16 | `"user-hooks/"` -> `"hooks/"` in help text |
| `internal/cmd/agent/agent.go` | 30 | `"user-agents"` -> `"agents"` in help text |
| `internal/cmd/agent/list.go` | 126 | `"user-agents"` -> `"agents"` |
| `internal/cmd/agent/validate.go` | 35, 48, 190 | `"user-agents"` -> `"agents"` |
| `internal/cmd/agent/update.go` | 262 | `"user-agents"` -> `"agents"` |
| `internal/cmd/session/wrap.go` | 174, 178 | Comment update: `user-hooks/session-guards/...` -> `hooks/...` |
| `internal/agent/integration_test.go` | 86, 89, 102, 114 | `"user-agents"` -> `"agents"` |
| `internal/usersync/usersync_test.go` | 71, 72, 74 | Test expectations: update SourceDir returns |
| `internal/materialize/hooks_test.go` | 381, 426 | `"user-hooks"` -> `"hooks"` in test setup |
| `internal/materialize/embedded_test.go` | 229 | `"user-hooks"` -> `"hooks"` in test setup |
| `embed.go` | 26-27 | See Section 6.2 |
| `cmd/ari/main.go` | (wherever `EmbeddedHooksYAML` is wired) | Update to `EmbeddedHooks` |

### 6.4 The lib/ Directory Question

The current `user-hooks/ari/` contains only scripts and `hooks.yaml`. There is no `lib/` subdirectory at the distribution level. However, the usersync `HooksSyncer` and help text reference `lib/`. After flattening to `hooks/`, the `lib/` directory (if it exists in user-level `~/.claude/hooks/lib/`) continues to work because usersync operates on the target directory, not the source structure.

---

## 7. Sprint Decomposition

All changes ship in one atomic PR (per D9). However, the implementation should proceed in this internal order to maintain a compilable state at each step:

### Sprint 1: Type System Changes (Foundation)
1. Add `ResourceMena` to `usersync/usersync.go`, remove `ResourceSkills` and `ResourceCommands`.
2. Update `Singular()`, `SourceDir()`, `RiteSubDir()` methods.
3. Update `NewSyncer()` to handle `ResourceMena` with dual target dirs.
4. Update `CollisionChecker` for `ResourceMena`.
5. Update all test expectations.
6. **Verify**: `CGO_ENABLED=0 go build ./cmd/ari` compiles (may need stub CLI changes).

### Sprint 2: Extension Stripping + ProjectMena API
1. Create `internal/materialize/project_mena.go` with:
   - `StripMenaExtension()`
   - `RouteMenaFile()`
   - `MenaSource` (exported from `menaSource`)
   - `MenaProjectionOptions`, `MenaProjectionResult`
   - `ProjectMena()`
   - `copyDirWithStripping()`, `copyDirFromFSWithStripping()`
2. Refactor `materializeMena()` to call `ProjectMena()`.
3. Write unit tests for `StripMenaExtension()` and `RouteMenaFile()`.
4. Write integration test for `ProjectMena()` with both modes.
5. **Verify**: `CGO_ENABLED=0 go test ./internal/materialize/...` passes.

### Sprint 3: Manifest Unification
1. Create `USER_MENA_MANIFEST.json` schema types in `usersync/manifest.go`.
2. Update `loadManifest()` with version check and wipe-on-mismatch.
3. Add `cleanupOldManifests()`.
4. Update `saveManifest()` for new schema.
5. Update `syncFiles()` for `ResourceMena`: dual-target routing, extension stripping on manifest keys.
6. Write tests for manifest migration (version mismatch -> wipe).
7. **Verify**: `CGO_ENABLED=0 go test ./internal/usersync/...` passes.

### Sprint 4: CLI Unification
1. Create `internal/cmd/sync/user_mena.go`.
2. Delete `internal/cmd/sync/user_commands.go` and `user_skills.go`.
3. Update `user.go` parent command.
4. Update `user_all.go` resource list.
5. Update `output.go` rendering order.
6. **Verify**: `CGO_ENABLED=0 go build ./cmd/ari && ari sync user --help` works.

### Sprint 5: Directory Renames
1. Execute `git mv` renames (Section 6.1).
2. Update `embed.go` (Section 6.2).
3. Update all path references (Section 6.3).
4. Update `cmd/ari/main.go` for new `EmbeddedHooks` type.
5. **Verify**: Full build and test suite: `CGO_ENABLED=0 go build ./cmd/ari && CGO_ENABLED=0 go test ./...`

### Sprint 6: Polish + Validation
1. Run full test suite.
2. Manual smoke test: `ari sync user mena`, `ari sync user all`, `ari rite start 10x-dev`.
3. Verify no `.dro` or `.lego` in any projected filenames.
4. Verify manifests are correctly formatted.
5. Verify `git diff --stat` shows clean renames.

---

## 8. Test Strategy

### 8.1 Unit Tests

| Test | File | What It Verifies |
|------|------|-----------------|
| `TestStripMenaExtension` | `project_mena_test.go` | All extension stripping cases: `.dro.md` -> `.md`, `.lego.md` -> `.md`, no-op on plain `.md`, no-op on non-md files, double-infix edge case. |
| `TestRouteMenaFile` | `project_mena_test.go` | Routing: `.dro.md` -> `"commands"`, `.lego.md` -> `"skills"`, plain `.md` -> `"commands"` (default). |
| `TestResourceMena_Singular` | `usersync_test.go` | `ResourceMena.Singular()` returns `"mena"`. |
| `TestResourceMena_SourceDir` | `usersync_test.go` | `ResourceMena.SourceDir()` returns `"mena"`. |
| `TestResourceMena_RiteSubDir` | `usersync_test.go` | `ResourceMena.RiteSubDir()` returns `"mena"`. |
| `TestMenaManifest_VersionMismatch` | `manifest_test.go` | Loading a v1.0 manifest returns empty manifest and creates backup. |
| `TestMenaManifest_CorruptJSON` | `manifest_test.go` | Loading corrupt JSON returns empty manifest and creates `.corrupt` backup. |
| `TestMenaManifest_RoundTrip` | `manifest_test.go` | Save then load preserves all fields including `mena_type` and `target`. |
| `TestCollisionChecker_Mena` | `collision_test.go` | Mena collision checks against `rites/*/mena/`. |

### 8.2 Integration Tests

| Test | File | What It Verifies |
|------|------|-----------------|
| `TestProjectMena_Destructive` | `project_mena_test.go` | Full projection with wipe: creates commands/ and skills/ with stripped filenames, no `.dro`/`.lego` in output. |
| `TestProjectMena_Additive` | `project_mena_test.go` | Additive projection preserves existing files in target dirs. |
| `TestProjectMena_PriorityOverride` | `project_mena_test.go` | Higher-priority source overrides lower-priority for same name. |
| `TestProjectMena_EmbeddedFS` | `project_mena_test.go` | Projection from `embed.FS` source works with extension stripping. |
| `TestMenaSyncer_EndToEnd` | `usersync_test.go` | Full sync cycle: source with `.dro.md` and `.lego.md` files -> manifest with stripped keys -> correct target directories. |
| `TestMenaSyncer_OldManifestWipe` | `usersync_test.go` | With pre-existing `USER_COMMAND_MANIFEST.json`, sync creates `USER_MENA_MANIFEST.json` and deletes old files. |
| `TestMaterializeMena_ExtensionStripping` | `materialize_test.go` | After `materializeMena()`, no file in `.claude/commands/` or `.claude/skills/` contains `.dro` or `.lego` in its name. |

### 8.3 Existing Test Updates

Tests that reference `ResourceSkills` or `ResourceCommands` must be updated to use `ResourceMena`. Tests that create fixture directories with `user-agents/`, `user-hooks/ari/`, or `user-skills/` must be updated to use `agents/`, `hooks/`, and `mena/` respectively.

Specific files requiring test updates:
- `internal/usersync/usersync_test.go` -- SourceDir and RiteSubDir expectations
- `internal/materialize/hooks_test.go` -- fixture directory paths
- `internal/materialize/embedded_test.go` -- fixture directory paths
- `internal/agent/integration_test.go` -- `user-agents` -> `agents`

---

## 9. Risk Mitigation

### 9.1 Risk: Build Breakage from Circular Renames

**Risk**: Renaming directories while updating Go imports could create intermediate states where `go build` fails.

**Mitigation**: Sprint decomposition is ordered so type system changes (Sprint 1) happen before directory renames (Sprint 5). All Go code changes compile against both old and new paths until Sprint 5. Sprint 5 is a single atomic commit with `git mv` + all path reference updates.

### 9.2 Risk: Manifest Data Loss

**Risk**: Wiping manifests loses divergence tracking history for users who have customized knossos-distributed commands.

**Mitigation**: Per D10, this is accepted. The backup file (`*.v1-backup`) preserves the old data. Users can manually inspect it if needed. The wipe-and-resync on next `ari sync user mena` will correctly detect diverged files via `--recover` flag.

### 9.3 Risk: Embedded FS Path Mismatch

**Risk**: Changing `embed.go` from `[]byte` to `embed.FS` for hooks could break the embedded binary if the directory structure assumption is wrong.

**Mitigation**: The `//go:embed hooks` directive embeds the entire `hooks/` directory tree. The `loadHooksConfig()` fallback reads `hooks.yaml` via `fs.ReadFile(m.embeddedHooks, "hooks.yaml")`. Since `hooks.yaml` is at the root of the embedded `hooks/` directory, the path is correct. Test `TestEmbeddedHooks` validates this.

### 9.4 Risk: Extension Stripping Breaks Existing .claude/ State

**Risk**: After the upgrade, project-level `.claude/commands/` will have files WITHOUT `.dro` in names, but Claude Code may have cached the old names.

**Mitigation**: Materialize is destructive (wipe-and-replace). The next `ari rite start` or any rematerialization wipes `.claude/commands/` and `.claude/skills/` entirely and regenerates with stripped names. No stale state survives.

### 9.5 Risk: Missing Path Reference Update

**Risk**: A `user-agents` or `user-hooks/ari` string reference is missed, causing runtime path resolution failures.

**Mitigation**: Sprint 5 includes a grep sweep: `grep -r "user-agents\|user-hooks\|user-skills" internal/ cmd/ embed.go`. Any remaining references are compilation errors (if in Go code) or runtime failures caught by smoke tests.

### 9.6 Risk: `ari sync user commands` / `ari sync user skills` Called by Scripts

**Risk**: External scripts or user muscle memory invoke removed subcommands.

**Mitigation**: The removed subcommands could be retained as hidden deprecated aliases that print a deprecation warning and delegate to `ari sync user mena`. However, per D1 ("no transition period"), we remove them outright. The parent command's help text clearly shows the new subcommand.

---

## Appendix A: File Inventory

### New Files
- `internal/materialize/project_mena.go` -- ProjectMena API, StripMenaExtension, RouteMenaFile
- `internal/materialize/project_mena_test.go` -- Unit and integration tests
- `internal/cmd/sync/user_mena.go` -- CLI subcommand

### Deleted Files
- `internal/cmd/sync/user_commands.go`
- `internal/cmd/sync/user_skills.go`
- `user-agents/moirai.md.backup`

### Renamed Files (git mv)
- `user-agents/` -> `agents/`
- `user-hooks/ari/*` -> `hooks/*`

### Modified Files
- `embed.go`
- `internal/materialize/materialize.go`
- `internal/materialize/hooks.go`
- `internal/materialize/frontmatter.go` (minor: export `StripMenaExtension` if placed here)
- `internal/usersync/usersync.go`
- `internal/usersync/manifest.go`
- `internal/usersync/output.go`
- `internal/usersync/collision.go`
- `internal/usersync/hooks.go` (no change needed -- it operates on target dir)
- `internal/cmd/sync/user.go`
- `internal/cmd/sync/user_all.go`
- `internal/cmd/sync/user_agents.go`
- `internal/cmd/sync/user_hooks.go`
- `internal/cmd/agent/agent.go`
- `internal/cmd/agent/list.go`
- `internal/cmd/agent/validate.go`
- `internal/cmd/agent/update.go`
- `internal/cmd/session/wrap.go`
- `internal/usersync/usersync_test.go`
- `internal/materialize/hooks_test.go`
- `internal/materialize/embedded_test.go`
- `internal/agent/integration_test.go`
- `cmd/ari/main.go` (EmbeddedHooksYAML -> EmbeddedHooks)

## Appendix B: Decision Traceability

| Decision | Section | Implementation |
|----------|---------|---------------|
| D1: Strip .dro/.lego | Section 2 | `StripMenaExtension()` applied in all copy paths |
| D5: Rename source dirs | Section 6 | `git mv` + path reference updates |
| D6: Merge user-skills into mena | Section 6.1 | Delete `user-skills/`, verify content in `mena/` |
| D7: Structure parity | Section 1.4, 1.5 | Same `ProjectMena()` used by both callers |
| D8: Materialize owns projection | Section 1.1 | All projection logic in `internal/materialize/` |
| D9: One atomic PR | Section 7 | All sprints in one PR |
| D10: Wipe and resync manifests | Section 3.2, 3.3 | Version mismatch -> backup + fresh manifest |
| D12: Parity for agents/hooks | Section 6 | Same rename pattern for all resource types |

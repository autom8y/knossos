# TDD: Native Go User Sync

> Technical Design Document for user-level resource syncing in the Ariadne CLI

**Status**: Draft
**Author**: Architect Agent
**Date**: 2026-01-10
**PRD**: docs/requirements/PRD-native-go-user-sync.md
**Spike**: docs/spikes/SPIKE-native-go-user-sync.md

---

## 1. Overview

This Technical Design Document specifies the implementation of the `internal/usersync` package for Ariadne (`ari`), providing native Go syncing of user-level resources from roster source directories to `~/.claude/`. This replaces 3,786 lines of bash scripts with approximately 1,300 lines of Go code while maintaining full backward compatibility with existing manifests.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-native-go-user-sync.md` |
| Spike | `docs/spikes/SPIKE-native-go-user-sync.md` |
| Prior Art | `internal/materialize/materialize.go` |
| Paths Package | `internal/paths/paths.go` |
| Rite Discovery | `internal/rite/discovery.go` |
| Config Package | `internal/config/home.go` |
| Shell Scripts | `sync-user-agents.sh`, `sync-user-skills.sh`, `sync-user-commands.sh`, `sync-user-hooks.sh` |

### 1.2 Scope

**In Scope**:
- New `internal/usersync/` package with core sync logic
- Four resource type syncers: agents, skills, commands, hooks
- JSON manifest management (backward compatible with shell scripts)
- SHA256 checksum-based change detection
- Rite collision detection (prevent shadowing rite resources)
- CLI commands under `ari sync user`
- Recovery mode for adopting existing files
- Force mode for overwriting diverged files
- Dry-run mode for previewing changes

**Out of Scope**:
- Project-level syncing (handled by `ari sync materialize`)
- Remote/network syncing (local roster checkout only)
- Watch mode / continuous sync
- Windows testing (cross-platform ready, not actively tested)
- Conflict resolution UI (users resolve manually)

### 1.3 Design Goals

1. **Backward Compatibility**: Read/write manifests in existing shell script format
2. **Additive Only**: Never remove user-created content
3. **Checksum Integrity**: SHA256 for reliable change detection
4. **Collision Safety**: Prevent syncing resources that shadow rite resources
5. **Cross-Platform Ready**: Use `filepath` for path handling, avoid shell-isms
6. **Reuse Patterns**: Follow `internal/materialize/` conventions
7. **Discoverability**: Expose via `ari sync user --help`

---

## 2. Architecture

### 2.1 Package Structure

```
internal/
├── usersync/
│   ├── usersync.go          # Core Syncer type, Options, Result
│   ├── manifest.go          # JSON manifest I/O
│   ├── checksum.go          # SHA256 checksum computation
│   ├── collision.go         # Rite collision detection
│   ├── agents.go            # Agent-specific sync (flat files)
│   ├── skills.go            # Skill-specific sync (nested directories)
│   ├── commands.go          # Command-specific sync (nested directories)
│   ├── hooks.go             # Hook-specific sync (lib/, yaml, +x)
│   └── usersync_test.go     # Comprehensive tests
│
├── cmd/sync/
│   ├── sync.go              # (existing) Add user subcommand
│   ├── user.go              # Parent 'ari sync user' command
│   ├── user_agents.go       # 'ari sync user agents'
│   ├── user_skills.go       # 'ari sync user skills'
│   ├── user_commands.go     # 'ari sync user commands'
│   ├── user_hooks.go        # 'ari sync user hooks'
│   └── user_all.go          # 'ari sync user all'
│
└── paths/
    └── paths.go             # (extend) Add user-level path helpers
```

### 2.2 Dependency Graph

```
                    ┌─────────────────────────────────┐
                    │  internal/cmd/sync/user*.go     │
                    │  (5 subcommands)                │
                    └─────────────┬───────────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                        │                        │
         v                        v                        v
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ internal/       │     │ internal/paths/ │     │ internal/output/│
│ usersync/       │     │ (extended)      │     │ (existing)      │
│ (business logic)│     └─────────────────┘     └─────────────────┘
└────────┬────────┘
         │
         ├─────────────────────────┬──────────────────────┐
         │                         │                      │
         v                         v                      v
┌─────────────────┐     ┌──────────────────┐    ┌─────────────────┐
│ internal/rite/  │     │ internal/config/ │    │ crypto/sha256   │
│ discovery       │     │ KnossosHome()    │    │ (stdlib)        │
│ (collision)     │     └──────────────────┘    └─────────────────┘
└─────────────────┘
         │
         v
┌─────────────────────────────────────────────────────────────────┐
│  Filesystem:                                                     │
│  - $KNOSSOS_HOME/user-{agents,skills,commands,hooks}/           │
│  - ~/.claude/{agents,skills,commands,hooks}/                    │
│  - ~/.claude/USER_{AGENT,SKILL,COMMAND,HOOKS}_MANIFEST.json     │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 Key Concepts

#### Resource Types

| Resource | Source | Target | Structure | Manifest |
|----------|--------|--------|-----------|----------|
| Agents | `$KNOSSOS_HOME/user-agents/` | `~/.claude/agents/` | Flat (*.md) | `USER_AGENT_MANIFEST.json` |
| Skills | `$KNOSSOS_HOME/user-skills/` | `~/.claude/skills/` | Nested (category/skill/*.md) | `USER_SKILL_MANIFEST.json` |
| Commands | `$KNOSSOS_HOME/user-commands/` | `~/.claude/commands/` | Nested (category/*.md) | `USER_COMMAND_MANIFEST.json` |
| Hooks | `$KNOSSOS_HOME/user-hooks/` | `~/.claude/hooks/` | Nested (lib/, *.yaml, scripts) | `USER_HOOKS_MANIFEST.json` |

#### Source Types

| Source | Description | Sync Behavior |
|--------|-------------|---------------|
| `roster` | Synced from roster, checksums match | Update on roster changes |
| `roster-diverged` | Originally roster but locally modified | Skip unless --force |
| `user` | Created by user, not in roster | Never touch |

#### Sync Modes

| Mode | Flag | Behavior |
|------|------|----------|
| Normal | (default) | Sync new/changed roster files, skip diverged |
| Dry Run | `--dry-run` | Preview changes without applying |
| Recover | `--recover` | Adopt existing files matching roster |
| Force | `--force` | Overwrite diverged files |

---

## 3. Interface Contracts

### 3.1 Command Summary

| Command | Description | Modifies Files |
|---------|-------------|----------------|
| `ari sync user agents` | Sync user-agents to ~/.claude/agents/ | Yes |
| `ari sync user skills` | Sync user-skills to ~/.claude/skills/ | Yes |
| `ari sync user commands` | Sync user-commands to ~/.claude/commands/ | Yes |
| `ari sync user hooks` | Sync user-hooks to ~/.claude/hooks/ | Yes |
| `ari sync user all` | Sync all user resources | Yes |

### 3.2 Command: `ari sync user agents`

Syncs agent files from roster `user-agents/` to `~/.claude/agents/`.

**Signature**:
```
ari sync user agents [--dry-run] [--recover] [--force] [--output=FORMAT] [--verbose]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--dry-run` | `-n` | bool | false | Preview changes without applying |
| `--recover` | `-r` | bool | false | Adopt existing files matching roster |
| `--force` | `-f` | bool | false | Overwrite diverged files |
| `--output` | `-o` | string | text | Output format: text, json |
| `--verbose` | `-v` | bool | false | Enable verbose output |

**Output (JSON)**:
```json
{
  "synced_at": "2026-01-10T12:00:00Z",
  "resource": "agents",
  "dry_run": false,
  "source": "/Users/user/Code/roster/user-agents",
  "target": "/Users/user/.claude/agents",
  "changes": {
    "added": ["context-engineer.md", "spike-runner.md"],
    "updated": ["consultant.md"],
    "skipped": [
      {"name": "my-custom-agent.md", "reason": "user-created"},
      {"name": "moirai.md", "reason": "collision with rite agent (10x-dev)"}
    ],
    "unchanged": ["code-reviewer.md", "debugger.md"]
  },
  "summary": {
    "added": 2,
    "updated": 1,
    "skipped": 2,
    "unchanged": 2,
    "collisions": 1
  }
}
```

**Output (text)**:
```
Syncing user agents...
  Source: /Users/user/Code/roster/user-agents
  Target: /Users/user/.claude/agents

  Added: context-engineer.md
  Added: spike-runner.md
  Updated: consultant.md
  Skipped: my-custom-agent.md (user-created)
  Collision: moirai.md (shadows rite agent in 10x-dev)

Summary: 2 added, 1 updated, 2 skipped, 2 unchanged, 1 collision
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Sync completed successfully |
| 1 | Sync completed with collisions detected |
| 2 | Invalid arguments |
| 3 | Source directory not found |
| 4 | Target directory creation failed |
| 5 | Manifest read/write error |

### 3.3 Command: `ari sync user skills`

Syncs skill directories from roster `user-skills/` to `~/.claude/skills/`.

**Signature**:
```
ari sync user skills [--dry-run] [--recover] [--force] [--output=FORMAT] [--verbose]
```

**Flags**: Same as agents command.

**Output (JSON)**:
```json
{
  "synced_at": "2026-01-10T12:00:00Z",
  "resource": "skills",
  "dry_run": false,
  "source": "/Users/user/Code/roster/user-skills",
  "target": "/Users/user/.claude/skills",
  "changes": {
    "added": ["documentation/doc-artifacts/SKILL.md"],
    "updated": ["orchestration/workflow/SKILL.md"],
    "skipped": [],
    "unchanged": ["session-lifecycle/moirai/SKILL.md"]
  },
  "summary": {
    "added": 1,
    "updated": 1,
    "skipped": 0,
    "unchanged": 1,
    "collisions": 0
  }
}
```

**Notes**:
- Skills use nested directory structure (category/skill-name/)
- Manifest keys are relative paths: `documentation/doc-artifacts/SKILL.md`
- Directory structure is preserved during sync

### 3.4 Command: `ari sync user commands`

Syncs command files from roster `user-commands/` to `~/.claude/commands/`.

**Signature**:
```
ari sync user commands [--dry-run] [--recover] [--force] [--output=FORMAT] [--verbose]
```

**Flags**: Same as agents command.

**Notes**:
- Commands use nested directory structure (category/)
- Manifest keys are relative paths: `operations/commit.md`
- Directory structure is preserved during sync

### 3.5 Command: `ari sync user hooks`

Syncs hook files from roster `user-hooks/` to `~/.claude/hooks/`.

**Signature**:
```
ari sync user hooks [--dry-run] [--recover] [--force] [--output=FORMAT] [--verbose]
```

**Flags**: Same as agents command.

**Special Handling**:
- `lib/` directory: Recursive copy of shared libraries
- `*.yaml` files: Hook configuration files
- Shell scripts: Preserve executable permissions (+x)
- Manifest keys are relative paths: `lib/session-manager.sh`, `hooks.yaml`

### 3.6 Command: `ari sync user all`

Syncs all four resource types in sequence.

**Signature**:
```
ari sync user all [--dry-run] [--recover] [--force] [--output=FORMAT] [--verbose]
```

**Flags**: Same as individual commands.

**Behavior**:
- Runs agents, skills, commands, hooks in sequence
- Failures in one type don't prevent syncing others
- Exit code reflects aggregate success/failure
- Summary shows results for each type

**Output (JSON)**:
```json
{
  "synced_at": "2026-01-10T12:00:00Z",
  "dry_run": false,
  "resources": {
    "agents": {
      "added": 2,
      "updated": 1,
      "skipped": 2,
      "unchanged": 5,
      "collisions": 1
    },
    "skills": {
      "added": 1,
      "updated": 0,
      "skipped": 0,
      "unchanged": 47,
      "collisions": 0
    },
    "commands": {
      "added": 0,
      "updated": 2,
      "skipped": 1,
      "unchanged": 31,
      "collisions": 0
    },
    "hooks": {
      "added": 3,
      "updated": 5,
      "skipped": 0,
      "unchanged": 12,
      "collisions": 0
    }
  },
  "totals": {
    "added": 6,
    "updated": 8,
    "skipped": 3,
    "unchanged": 95,
    "collisions": 1
  }
}
```

**Output (text)**:
```
Syncing all user resources...

Agents: 2 added, 1 updated, 2 skipped, 5 unchanged, 1 collision
Skills: 1 added, 0 updated, 0 skipped, 47 unchanged
Commands: 0 added, 2 updated, 1 skipped, 31 unchanged
Hooks: 3 added, 5 updated, 0 skipped, 12 unchanged

Totals: 6 added, 8 updated, 3 skipped, 95 unchanged, 1 collision
```

---

## 4. Data Model

### 4.1 Manifest Schema

JSON manifest stored at `~/.claude/USER_{TYPE}_MANIFEST.json`:

```json
{
  "manifest_version": "1.0",
  "last_sync": "2026-01-10T12:00:00Z",
  "agents": {
    "context-engineer.md": {
      "source": "roster",
      "installed_at": "2026-01-10T12:00:00Z",
      "checksum": "sha256:abc123def456..."
    },
    "my-custom-agent.md": {
      "source": "user",
      "installed_at": "2026-01-05T10:00:00Z",
      "checksum": "sha256:789xyz012..."
    },
    "modified-agent.md": {
      "source": "roster-diverged",
      "installed_at": "2026-01-08T14:00:00Z",
      "checksum": "sha256:diverged123..."
    }
  }
}
```

**Schema Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `manifest_version` | string | Schema version ("1.0") |
| `last_sync` | string | ISO 8601 timestamp of last sync |
| `{entries}` | object | Map of filename to entry |
| `{entry}.source` | string | "roster", "roster-diverged", or "user" |
| `{entry}.installed_at` | string | ISO 8601 timestamp |
| `{entry}.checksum` | string | SHA256 checksum with prefix |

**Note**: The entry key name varies by resource type:
- Agents: `agents`
- Skills: `skills`
- Commands: `commands`
- Hooks: `hooks`

### 4.2 Go Type Definitions

```go
// Package usersync syncs user-level resources to ~/.claude/
package usersync

import "time"

// ResourceType identifies the type of user resource.
type ResourceType string

const (
    ResourceAgents   ResourceType = "agents"
    ResourceSkills   ResourceType = "skills"
    ResourceCommands ResourceType = "commands"
    ResourceHooks    ResourceType = "hooks"
)

// SourceType identifies the origin of a synced resource.
type SourceType string

const (
    SourceRoster   SourceType = "roster"          // Synced from roster, unchanged
    SourceDiverged SourceType = "roster-diverged" // From roster but locally modified
    SourceUser     SourceType = "user"            // User-created, not in roster
)

// Manifest represents a user resource manifest.
type Manifest struct {
    Version  string           `json:"manifest_version"`
    LastSync time.Time        `json:"last_sync"`
    Entries  map[string]Entry `json:"agents"` // Key varies by type
}

// Entry represents a single resource entry in the manifest.
type Entry struct {
    Source      SourceType `json:"source"`
    InstalledAt time.Time  `json:"installed_at"`
    Checksum    string     `json:"checksum"`
}

// Options configures sync behavior.
type Options struct {
    DryRun   bool // Preview changes without applying
    Recover  bool // Adopt existing files matching roster
    Force    bool // Overwrite diverged files
    Verbose  bool // Enable verbose logging
}

// Result contains sync operation outcome.
type Result struct {
    SyncedAt   time.Time       `json:"synced_at"`
    Resource   ResourceType    `json:"resource"`
    DryRun     bool            `json:"dry_run"`
    Source     string          `json:"source"`
    Target     string          `json:"target"`
    Changes    Changes         `json:"changes"`
    Summary    Summary         `json:"summary"`
}

// Changes categorizes sync outcomes by file.
type Changes struct {
    Added     []string        `json:"added"`
    Updated   []string        `json:"updated"`
    Skipped   []SkippedEntry  `json:"skipped"`
    Unchanged []string        `json:"unchanged"`
}

// SkippedEntry explains why a file was skipped.
type SkippedEntry struct {
    Name   string `json:"name"`
    Reason string `json:"reason"`
}

// Summary provides aggregate counts.
type Summary struct {
    Added      int `json:"added"`
    Updated    int `json:"updated"`
    Skipped    int `json:"skipped"`
    Unchanged  int `json:"unchanged"`
    Collisions int `json:"collisions"`
}
```

---

## 5. Internal Package Design

### 5.1 Package: `internal/usersync/usersync.go`

Core syncer type and orchestration.

```go
package usersync

import (
    "os"
    "path/filepath"
    "time"

    "github.com/autom8y/knossos/internal/config"
    "github.com/autom8y/knossos/internal/paths"
    "github.com/autom8y/knossos/internal/rite"
)

// Syncer handles user resource synchronization.
type Syncer struct {
    resourceType  ResourceType
    sourceDir     string
    targetDir     string
    manifestPath  string
    riteDiscovery *rite.Discovery
    nested        bool // true for skills, commands, hooks
}

// NewSyncer creates a syncer for the given resource type.
func NewSyncer(resourceType ResourceType) (*Syncer, error) {
    knossosHome := config.KnossosHome()
    if knossosHome == "" {
        return nil, ErrKnossosHomeNotSet
    }

    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }

    s := &Syncer{
        resourceType: resourceType,
    }

    switch resourceType {
    case ResourceAgents:
        s.sourceDir = filepath.Join(knossosHome, "user-agents")
        s.targetDir = filepath.Join(homeDir, ".claude", "agents")
        s.manifestPath = filepath.Join(homeDir, ".claude", "USER_AGENT_MANIFEST.json")
        s.nested = false
    case ResourceSkills:
        s.sourceDir = filepath.Join(knossosHome, "user-skills")
        s.targetDir = filepath.Join(homeDir, ".claude", "skills")
        s.manifestPath = filepath.Join(homeDir, ".claude", "USER_SKILL_MANIFEST.json")
        s.nested = true
    case ResourceCommands:
        s.sourceDir = filepath.Join(knossosHome, "user-commands")
        s.targetDir = filepath.Join(homeDir, ".claude", "commands")
        s.manifestPath = filepath.Join(homeDir, ".claude", "USER_COMMAND_MANIFEST.json")
        s.nested = true
    case ResourceHooks:
        s.sourceDir = filepath.Join(knossosHome, "user-hooks")
        s.targetDir = filepath.Join(homeDir, ".claude", "hooks")
        s.manifestPath = filepath.Join(homeDir, ".claude", "USER_HOOKS_MANIFEST.json")
        s.nested = true
    default:
        return nil, ErrInvalidResourceType
    }

    // Initialize rite discovery for collision detection
    // Uses project root if available, falls back to knossos home
    resolver := paths.NewResolver(knossosHome)
    s.riteDiscovery = rite.NewDiscovery(resolver)

    return s, nil
}

// Sync performs the synchronization operation.
func (s *Syncer) Sync(opts Options) (*Result, error) {
    result := &Result{
        SyncedAt: time.Now().UTC(),
        Resource: s.resourceType,
        DryRun:   opts.DryRun,
        Source:   s.sourceDir,
        Target:   s.targetDir,
        Changes: Changes{
            Added:     []string{},
            Updated:   []string{},
            Skipped:   []SkippedEntry{},
            Unchanged: []string{},
        },
    }

    // Check source directory exists
    if _, err := os.Stat(s.sourceDir); os.IsNotExist(err) {
        return nil, ErrSourceNotFound(s.sourceDir)
    }

    // Ensure target directory exists
    if !opts.DryRun {
        if err := paths.EnsureDir(s.targetDir); err != nil {
            return nil, ErrTargetCreateFailed(s.targetDir, err)
        }
    }

    // Load or create manifest
    manifest, err := s.loadManifest()
    if err != nil {
        return nil, err
    }

    // Handle recovery mode first
    if opts.Recover {
        if err := s.recover(manifest, result, opts); err != nil {
            return nil, err
        }
    }

    // Sync source files to target
    if err := s.syncFiles(manifest, result, opts); err != nil {
        return nil, err
    }

    // Update manifest
    if !opts.DryRun {
        manifest.LastSync = result.SyncedAt
        if err := s.saveManifest(manifest); err != nil {
            return nil, err
        }
    }

    // Calculate summary
    result.Summary = Summary{
        Added:     len(result.Changes.Added),
        Updated:   len(result.Changes.Updated),
        Skipped:   len(result.Changes.Skipped),
        Unchanged: len(result.Changes.Unchanged),
        Collisions: s.countCollisions(result.Changes.Skipped),
    }

    return result, nil
}

// syncFiles iterates source files and syncs to target.
func (s *Syncer) syncFiles(manifest *Manifest, result *Result, opts Options) error {
    return filepath.WalkDir(s.sourceDir, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Skip directories (we process files)
        if d.IsDir() {
            return nil
        }

        // Compute relative path for manifest key
        relPath, err := filepath.Rel(s.sourceDir, path)
        if err != nil {
            return err
        }

        // For flat resources, use just the filename
        manifestKey := relPath
        if !s.nested {
            manifestKey = filepath.Base(relPath)
        }

        // Check for rite collision
        if collision, riteName := s.checkRiteCollision(manifestKey); collision {
            result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
                Name:   manifestKey,
                Reason: "collision with rite " + s.resourceType.Singular() + " (" + riteName + ")",
            })
            return nil
        }

        // Calculate source checksum
        sourceChecksum, err := ComputeFileChecksum(path)
        if err != nil {
            return err
        }

        // Check existing manifest entry
        entry, exists := manifest.Entries[manifestKey]
        targetPath := filepath.Join(s.targetDir, manifestKey)

        if !exists {
            // New file - check if target exists (untracked)
            if _, err := os.Stat(targetPath); err == nil {
                // Target exists but not in manifest - mark as user-created
                if opts.Recover {
                    targetChecksum, _ := ComputeFileChecksum(targetPath)
                    if targetChecksum == sourceChecksum {
                        // Exact match - adopt as roster
                        manifest.Entries[manifestKey] = Entry{
                            Source:      SourceRoster,
                            InstalledAt: result.SyncedAt,
                            Checksum:    sourceChecksum,
                        }
                        result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
                    } else {
                        // Different - adopt as diverged
                        manifest.Entries[manifestKey] = Entry{
                            Source:      SourceDiverged,
                            InstalledAt: result.SyncedAt,
                            Checksum:    targetChecksum,
                        }
                        result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
                            Name:   manifestKey,
                            Reason: "adopted as diverged (local modifications)",
                        })
                    }
                    return nil
                }
                // Not recovering - skip as user-created
                manifest.Entries[manifestKey] = Entry{
                    Source:      SourceUser,
                    InstalledAt: result.SyncedAt,
                    Checksum:    "",
                }
                result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
                    Name:   manifestKey,
                    Reason: "user-created",
                })
                return nil
            }

            // New file, target doesn't exist - add it
            if !opts.DryRun {
                if err := s.copyFile(path, targetPath); err != nil {
                    return err
                }
                manifest.Entries[manifestKey] = Entry{
                    Source:      SourceRoster,
                    InstalledAt: result.SyncedAt,
                    Checksum:    sourceChecksum,
                }
            }
            result.Changes.Added = append(result.Changes.Added, manifestKey)
            return nil
        }

        // Existing entry
        switch entry.Source {
        case SourceUser:
            // Never touch user-created files
            result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
                Name:   manifestKey,
                Reason: "user-created",
            })

        case SourceDiverged:
            if opts.Force {
                // Force overwrite
                if !opts.DryRun {
                    if err := s.copyFile(path, targetPath); err != nil {
                        return err
                    }
                    manifest.Entries[manifestKey] = Entry{
                        Source:      SourceRoster,
                        InstalledAt: result.SyncedAt,
                        Checksum:    sourceChecksum,
                    }
                }
                result.Changes.Updated = append(result.Changes.Updated, manifestKey)
            } else {
                // Skip diverged without force
                result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
                    Name:   manifestKey,
                    Reason: "diverged (use --force to overwrite)",
                })
            }

        case SourceRoster:
            // Check if source changed
            if entry.Checksum == sourceChecksum {
                // No change in source
                result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
            } else {
                // Source changed - check if target diverged
                targetChecksum, _ := ComputeFileChecksum(targetPath)
                if targetChecksum == entry.Checksum {
                    // Target unchanged, update from source
                    if !opts.DryRun {
                        if err := s.copyFile(path, targetPath); err != nil {
                            return err
                        }
                        manifest.Entries[manifestKey] = Entry{
                            Source:      SourceRoster,
                            InstalledAt: result.SyncedAt,
                            Checksum:    sourceChecksum,
                        }
                    }
                    result.Changes.Updated = append(result.Changes.Updated, manifestKey)
                } else {
                    // Target diverged - mark as diverged
                    manifest.Entries[manifestKey] = Entry{
                        Source:      SourceDiverged,
                        InstalledAt: entry.InstalledAt,
                        Checksum:    targetChecksum,
                    }
                    result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
                        Name:   manifestKey,
                        Reason: "diverged (local modifications)",
                    })
                }
            }
        }

        return nil
    })
}

// recover adopts existing target files that match roster sources.
func (s *Syncer) recover(manifest *Manifest, result *Result, opts Options) error {
    // Walk target directory looking for untracked files
    return filepath.WalkDir(s.targetDir, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }

        relPath, _ := filepath.Rel(s.targetDir, path)
        manifestKey := relPath
        if !s.nested {
            manifestKey = filepath.Base(relPath)
        }

        // Skip if already in manifest
        if _, exists := manifest.Entries[manifestKey]; exists {
            return nil
        }

        // Check if source exists
        sourcePath := filepath.Join(s.sourceDir, relPath)
        if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
            // Not in roster - mark as user
            if !opts.DryRun {
                targetChecksum, _ := ComputeFileChecksum(path)
                manifest.Entries[manifestKey] = Entry{
                    Source:      SourceUser,
                    InstalledAt: result.SyncedAt,
                    Checksum:    targetChecksum,
                }
            }
            return nil
        }

        // Compare checksums
        sourceChecksum, _ := ComputeFileChecksum(sourcePath)
        targetChecksum, _ := ComputeFileChecksum(path)

        if !opts.DryRun {
            if sourceChecksum == targetChecksum {
                manifest.Entries[manifestKey] = Entry{
                    Source:      SourceRoster,
                    InstalledAt: result.SyncedAt,
                    Checksum:    sourceChecksum,
                }
            } else {
                manifest.Entries[manifestKey] = Entry{
                    Source:      SourceDiverged,
                    InstalledAt: result.SyncedAt,
                    Checksum:    targetChecksum,
                }
            }
        }

        return nil
    })
}

// copyFile copies a file preserving permissions.
func (s *Syncer) copyFile(src, dst string) error {
    // Ensure parent directory exists
    if err := paths.EnsureDir(filepath.Dir(dst)); err != nil {
        return err
    }

    // Read source
    content, err := os.ReadFile(src)
    if err != nil {
        return err
    }

    // Get source permissions
    info, err := os.Stat(src)
    if err != nil {
        return err
    }

    // Write destination with same permissions
    return os.WriteFile(dst, content, info.Mode())
}

// countCollisions counts collision entries in skipped list.
func (s *Syncer) countCollisions(skipped []SkippedEntry) int {
    count := 0
    for _, entry := range skipped {
        if strings.Contains(entry.Reason, "collision") {
            count++
        }
    }
    return count
}
```

### 5.2 Package: `internal/usersync/manifest.go`

JSON manifest I/O operations.

```go
package usersync

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

// ManifestVersion is the current manifest schema version.
const ManifestVersion = "1.0"

// manifestJSON is the on-disk manifest format (for backward compatibility).
// The entry key varies by resource type.
type manifestJSON struct {
    Version  string           `json:"manifest_version"`
    LastSync string           `json:"last_sync"`
    Agents   map[string]entryJSON `json:"agents,omitempty"`
    Skills   map[string]entryJSON `json:"skills,omitempty"`
    Commands map[string]entryJSON `json:"commands,omitempty"`
    Hooks    map[string]entryJSON `json:"hooks,omitempty"`
}

type entryJSON struct {
    Source      string `json:"source"`
    InstalledAt string `json:"installed_at"`
    Checksum    string `json:"checksum"`
}

// loadManifest reads the manifest from disk.
func (s *Syncer) loadManifest() (*Manifest, error) {
    data, err := os.ReadFile(s.manifestPath)
    if err != nil {
        if os.IsNotExist(err) {
            // Return empty manifest
            return &Manifest{
                Version:  ManifestVersion,
                LastSync: time.Time{},
                Entries:  make(map[string]Entry),
            }, nil
        }
        return nil, ErrManifestRead(s.manifestPath, err)
    }

    var mj manifestJSON
    if err := json.Unmarshal(data, &mj); err != nil {
        // Manifest corrupt - backup and create new
        backupPath := s.manifestPath + ".corrupt"
        os.Rename(s.manifestPath, backupPath)
        return &Manifest{
            Version:  ManifestVersion,
            LastSync: time.Time{},
            Entries:  make(map[string]Entry),
        }, nil
    }

    // Convert to internal format
    manifest := &Manifest{
        Version: mj.Version,
        Entries: make(map[string]Entry),
    }

    if t, err := time.Parse(time.RFC3339, mj.LastSync); err == nil {
        manifest.LastSync = t
    }

    // Get entries based on resource type
    var entries map[string]entryJSON
    switch s.resourceType {
    case ResourceAgents:
        entries = mj.Agents
    case ResourceSkills:
        entries = mj.Skills
    case ResourceCommands:
        entries = mj.Commands
    case ResourceHooks:
        entries = mj.Hooks
    }

    for name, ej := range entries {
        entry := Entry{
            Source:   SourceType(ej.Source),
            Checksum: ej.Checksum,
        }
        if t, err := time.Parse(time.RFC3339, ej.InstalledAt); err == nil {
            entry.InstalledAt = t
        }
        manifest.Entries[name] = entry
    }

    return manifest, nil
}

// saveManifest writes the manifest to disk.
func (s *Syncer) saveManifest(manifest *Manifest) error {
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(s.manifestPath), 0755); err != nil {
        return ErrManifestWrite(s.manifestPath, err)
    }

    // Convert to JSON format
    mj := manifestJSON{
        Version:  manifest.Version,
        LastSync: manifest.LastSync.Format(time.RFC3339),
    }

    entries := make(map[string]entryJSON)
    for name, entry := range manifest.Entries {
        entries[name] = entryJSON{
            Source:      string(entry.Source),
            InstalledAt: entry.InstalledAt.Format(time.RFC3339),
            Checksum:    entry.Checksum,
        }
    }

    // Set entries in correct field
    switch s.resourceType {
    case ResourceAgents:
        mj.Agents = entries
    case ResourceSkills:
        mj.Skills = entries
    case ResourceCommands:
        mj.Commands = entries
    case ResourceHooks:
        mj.Hooks = entries
    }

    data, err := json.MarshalIndent(mj, "", "  ")
    if err != nil {
        return ErrManifestWrite(s.manifestPath, err)
    }

    return os.WriteFile(s.manifestPath, data, 0644)
}
```

### 5.3 Package: `internal/usersync/checksum.go`

SHA256 checksum computation.

```go
package usersync

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
)

// ComputeFileChecksum calculates SHA256 checksum of a file.
func ComputeFileChecksum(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }

    return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeContentChecksum calculates SHA256 checksum of content.
func ComputeContentChecksum(content []byte) string {
    h := sha256.New()
    h.Write(content)
    return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
```

### 5.4 Package: `internal/usersync/collision.go`

Rite collision detection.

```go
package usersync

import (
    "os"
    "path/filepath"
    "strings"

    "github.com/autom8y/knossos/internal/config"
)

// checkRiteCollision checks if a resource name exists in any rite.
// Returns (hasCollision, riteName).
func (s *Syncer) checkRiteCollision(name string) (bool, string) {
    knossosHome := config.KnossosHome()
    if knossosHome == "" {
        return false, ""
    }

    ritesDir := filepath.Join(knossosHome, "rites")
    if _, err := os.Stat(ritesDir); os.IsNotExist(err) {
        return false, ""
    }

    // Get the resource subdirectory name (agents, skills, commands, hooks)
    subDir := string(s.resourceType)

    // For flat resources (agents), use just filename
    // For nested resources (skills, commands, hooks), use full relative path
    searchName := name
    if !s.nested {
        searchName = filepath.Base(name)
    }

    // Search each rite
    entries, err := os.ReadDir(ritesDir)
    if err != nil {
        return false, ""
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        riteName := entry.Name()
        resourcePath := filepath.Join(ritesDir, riteName, subDir, searchName)

        if _, err := os.Stat(resourcePath); err == nil {
            return true, riteName
        }
    }

    return false, ""
}

// GetRiteForResource finds which rite(s) contain a resource.
// Returns comma-separated list of rite names.
func GetRiteForResource(resourceType ResourceType, name string) string {
    knossosHome := config.KnossosHome()
    if knossosHome == "" {
        return ""
    }

    ritesDir := filepath.Join(knossosHome, "rites")
    entries, err := os.ReadDir(ritesDir)
    if err != nil {
        return ""
    }

    subDir := string(resourceType)
    var matches []string

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        riteName := entry.Name()
        resourcePath := filepath.Join(ritesDir, riteName, subDir, name)

        if _, err := os.Stat(resourcePath); err == nil {
            matches = append(matches, riteName)
        }
    }

    return strings.Join(matches, ", ")
}
```

### 5.5 Package: `internal/usersync/hooks.go`

Hook-specific sync logic with executable preservation.

```go
package usersync

import (
    "os"
    "path/filepath"
    "strings"
)

// isExecutable checks if a file path is an executable script.
func isExecutable(path string) bool {
    // Check by extension
    ext := strings.ToLower(filepath.Ext(path))
    if ext == ".sh" || ext == ".bash" || ext == ".zsh" {
        return true
    }

    // Check if in lib/ directory
    if strings.Contains(path, string(filepath.Separator)+"lib"+string(filepath.Separator)) {
        return true
    }

    // Check if file has executable bit (Unix)
    info, err := os.Stat(path)
    if err != nil {
        return false
    }

    return info.Mode()&0111 != 0
}

// copyFileWithExecutable copies a file, preserving +x if needed.
func copyFileWithExecutable(src, dst string) error {
    content, err := os.ReadFile(src)
    if err != nil {
        return err
    }

    // Get source permissions
    info, err := os.Stat(src)
    if err != nil {
        return err
    }

    perm := info.Mode()

    // Ensure executable bit for scripts
    if isExecutable(src) && perm&0111 == 0 {
        perm |= 0755
    }

    return os.WriteFile(dst, content, perm)
}
```

### 5.6 Package: `internal/usersync/errors.go`

Error definitions.

```go
package usersync

import (
    "fmt"

    "github.com/autom8y/knossos/internal/errors"
)

// Error codes for usersync package.
const (
    CodeKnossosHomeNotSet   = "KNOSSOS_HOME_NOT_SET"
    CodeInvalidResourceType = "INVALID_RESOURCE_TYPE"
    CodeSourceNotFound      = "SOURCE_NOT_FOUND"
    CodeTargetCreateFailed  = "TARGET_CREATE_FAILED"
    CodeManifestReadError   = "MANIFEST_READ_ERROR"
    CodeManifestWriteError  = "MANIFEST_WRITE_ERROR"
    CodeChecksumError       = "CHECKSUM_ERROR"
)

// Package-level errors.
var (
    ErrKnossosHomeNotSet   = errors.New(CodeKnossosHomeNotSet, "KNOSSOS_HOME environment variable not set")
    ErrInvalidResourceType = errors.New(CodeInvalidResourceType, "invalid resource type")
)

// ErrSourceNotFound returns an error for missing source directory.
func ErrSourceNotFound(path string) error {
    return errors.NewWithDetails(CodeSourceNotFound,
        fmt.Sprintf("source directory not found: %s", path),
        map[string]any{"path": path})
}

// ErrTargetCreateFailed returns an error for target directory creation failure.
func ErrTargetCreateFailed(path string, cause error) error {
    return errors.NewWithDetails(CodeTargetCreateFailed,
        fmt.Sprintf("failed to create target directory: %s", path),
        map[string]any{"path": path, "cause": cause.Error()})
}

// ErrManifestRead returns an error for manifest read failure.
func ErrManifestRead(path string, cause error) error {
    return errors.NewWithDetails(CodeManifestReadError,
        fmt.Sprintf("failed to read manifest: %s", path),
        map[string]any{"path": path, "cause": cause.Error()})
}

// ErrManifestWrite returns an error for manifest write failure.
func ErrManifestWrite(path string, cause error) error {
    return errors.NewWithDetails(CodeManifestWriteError,
        fmt.Sprintf("failed to write manifest: %s", path),
        map[string]any{"path": path, "cause": cause.Error()})
}
```

---

## 6. CLI Command Implementation

### 6.1 Parent Command: `internal/cmd/sync/user.go`

```go
package sync

import (
    "github.com/spf13/cobra"
)

func newUserCmd(ctx *cmdContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "user",
        Short: "Sync user-level resources to ~/.claude/",
        Long: `Sync user-level resources from roster to ~/.claude/.

User resources are globally available across all projects.
They are stored in ~/.claude/ and synced from $KNOSSOS_HOME/user-{type}/.

Resources:
  agents    - Agent prompts (user-agents/ -> ~/.claude/agents/)
  skills    - Skill references (user-skills/ -> ~/.claude/skills/)
  commands  - Slash commands (user-commands/ -> ~/.claude/commands/)
  hooks     - Hook scripts (user-hooks/ -> ~/.claude/hooks/)

Sync Behavior:
  - Additive: Never removes user-created content
  - Checksum-based: Only updates when source changes
  - Collision-aware: Skips resources that would shadow rite resources`,
    }

    // Add subcommands
    cmd.AddCommand(newUserAgentsCmd(ctx))
    cmd.AddCommand(newUserSkillsCmd(ctx))
    cmd.AddCommand(newUserCommandsCmd(ctx))
    cmd.AddCommand(newUserHooksCmd(ctx))
    cmd.AddCommand(newUserAllCmd(ctx))

    return cmd
}
```

### 6.2 Subcommand: `internal/cmd/sync/user_agents.go`

```go
package sync

import (
    "github.com/spf13/cobra"

    "github.com/autom8y/knossos/internal/usersync"
)

func newUserAgentsCmd(ctx *cmdContext) *cobra.Command {
    var dryRun, recover, force, verbose bool

    cmd := &cobra.Command{
        Use:   "agents",
        Short: "Sync user agents to ~/.claude/agents/",
        Long: `Sync agent files from roster user-agents/ to ~/.claude/agents/.

Behavior:
  - Adds new agents from roster
  - Updates roster-managed agents when source changes
  - Preserves user-created agents
  - Skips agents that would shadow rite agents`,
        SilenceUsage: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            opts := usersync.Options{
                DryRun:  dryRun,
                Recover: recover,
                Force:   force,
                Verbose: verbose,
            }
            return runUserSync(ctx, usersync.ResourceAgents, opts)
        },
    }

    cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
    cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching roster")
    cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
    cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

    return cmd
}

func runUserSync(ctx *cmdContext, resourceType usersync.ResourceType, opts usersync.Options) error {
    printer := ctx.getPrinter()

    syncer, err := usersync.NewSyncer(resourceType)
    if err != nil {
        printer.PrintError(err)
        return err
    }

    result, err := syncer.Sync(opts)
    if err != nil {
        printer.PrintError(err)
        return err
    }

    return printer.Print(result)
}
```

### 6.3 Integration with Existing Sync Command

Update `internal/cmd/sync/sync.go`:

```go
// NewSyncCmd creates the sync command group.
func NewSyncCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
    ctx := &cmdContext{
        BaseContext: common.BaseContext{
            Output:     outputFlag,
            Verbose:    verboseFlag,
            ProjectDir: projectDir,
        },
    }

    cmd := &cobra.Command{
        Use:   "sync",
        Short: "Synchronize configuration with remotes",
        // ... existing long description ...
    }

    // Existing subcommands
    cmd.AddCommand(newMaterializeCmd(ctx))
    cmd.AddCommand(newStatusCmd(ctx))
    // ... other existing commands ...

    // NEW: Add user sync subcommands
    cmd.AddCommand(newUserCmd(ctx))

    // ... rest of existing code ...
    return cmd
}
```

---

## 7. Path Resolution Extensions

### 7.1 Extensions to `internal/paths/paths.go`

```go
// UserClaudeDir returns the user-level .claude directory.
func UserClaudeDir() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".claude")
}

// UserAgentsDir returns the user-level agents directory.
func UserAgentsDir() string {
    return filepath.Join(UserClaudeDir(), "agents")
}

// UserSkillsDir returns the user-level skills directory.
func UserSkillsDir() string {
    return filepath.Join(UserClaudeDir(), "skills")
}

// UserCommandsDir returns the user-level commands directory.
func UserCommandsDir() string {
    return filepath.Join(UserClaudeDir(), "commands")
}

// UserHooksDir returns the user-level hooks directory.
func UserHooksDir() string {
    return filepath.Join(UserClaudeDir(), "hooks")
}

// UserAgentManifest returns the path to the user agent manifest.
func UserAgentManifest() string {
    return filepath.Join(UserClaudeDir(), "USER_AGENT_MANIFEST.json")
}

// UserSkillManifest returns the path to the user skill manifest.
func UserSkillManifest() string {
    return filepath.Join(UserClaudeDir(), "USER_SKILL_MANIFEST.json")
}

// UserCommandManifest returns the path to the user command manifest.
func UserCommandManifest() string {
    return filepath.Join(UserClaudeDir(), "USER_COMMAND_MANIFEST.json")
}

// UserHooksManifest returns the path to the user hooks manifest.
func UserHooksManifest() string {
    return filepath.Join(UserClaudeDir(), "USER_HOOKS_MANIFEST.json")
}
```

---

## 8. Test Strategy

### 8.1 Unit Tests

Location: `internal/usersync/usersync_test.go`

| Test | Description | Coverage |
|------|-------------|----------|
| `TestNewSyncer_ValidTypes` | Creates syncer for each resource type | 100% |
| `TestNewSyncer_InvalidType` | Returns error for invalid type | 100% |
| `TestComputeFileChecksum` | SHA256 computation is correct | 100% |
| `TestLoadManifest_NotExists` | Returns empty manifest when file missing | 100% |
| `TestLoadManifest_Valid` | Parses valid JSON manifest | 100% |
| `TestLoadManifest_Corrupt` | Handles corrupt manifest gracefully | 100% |
| `TestSaveManifest` | Writes valid JSON manifest | 100% |
| `TestSync_AddNew` | Adds new files from source | 100% |
| `TestSync_UpdateChanged` | Updates when source changes | 100% |
| `TestSync_SkipUser` | Skips user-created files | 100% |
| `TestSync_SkipDiverged` | Skips diverged without --force | 100% |
| `TestSync_ForceDiverged` | Overwrites diverged with --force | 100% |
| `TestSync_DryRun` | Previews without modifying | 100% |
| `TestSync_Recover` | Adopts existing files | 100% |
| `TestCheckRiteCollision` | Detects rite collisions | 100% |
| `TestCopyFilePreservesPermissions` | Maintains +x on scripts | 100% |

### 8.2 Integration Tests

Location: `internal/usersync/integration_test.go`

| Test ID | Description |
|---------|-------------|
| `usersync_001` | Full agents sync with mixed file states |
| `usersync_002` | Nested skills sync preserves directory structure |
| `usersync_003` | Commands sync with collision detection |
| `usersync_004` | Hooks sync preserves executable permissions |
| `usersync_005` | All sync runs each type in sequence |
| `usersync_006` | Manifest backward compatibility with shell script format |
| `usersync_007` | Recovery mode adopts existing files |
| `usersync_008` | Dry-run produces accurate preview |

### 8.3 Test Fixtures

```
internal/usersync/testdata/
├── source/
│   ├── user-agents/
│   │   ├── context-engineer.md
│   │   └── spike-runner.md
│   ├── user-skills/
│   │   └── documentation/
│   │       └── doc-artifacts/
│   │           └── SKILL.md
│   ├── user-commands/
│   │   └── operations/
│   │       └── commit.md
│   └── user-hooks/
│       ├── lib/
│       │   └── session-manager.sh
│       └── hooks.yaml
├── target/
│   └── .claude/
│       ├── agents/
│       │   └── existing-agent.md
│       └── USER_AGENT_MANIFEST.json
└── rites/
    └── 10x-dev/
        └── agents/
            └── moirai.md  # For collision testing
```

---

## 9. Error Handling

### 9.1 Exit Codes

| Code | Name | Description |
|------|------|-------------|
| 0 | Success | Sync completed successfully |
| 1 | Collisions | Sync completed with collisions detected |
| 2 | UsageError | Invalid arguments or flags |
| 3 | SourceNotFound | Source directory doesn't exist |
| 4 | TargetError | Cannot create target directory |
| 5 | ManifestError | Manifest read/write failure |
| 6 | KnossosHomeNotSet | KNOSSOS_HOME not configured |

### 9.2 Error Messages

| Error | Message |
|-------|---------|
| KnossosHomeNotSet | "KNOSSOS_HOME environment variable not set" |
| SourceNotFound | "Source directory not found: {path}" |
| TargetCreateFailed | "Failed to create target directory: {path}" |
| ManifestCorrupt | "Manifest corrupt, backed up to {path}.corrupt" |

---

## 10. Implementation Guidance

### 10.1 Recommended Order

1. **Foundation** (Day 1)
   - `internal/usersync/errors.go` - Error definitions
   - `internal/usersync/checksum.go` - SHA256 computation
   - `internal/paths/paths.go` extensions - User path helpers

2. **Manifest** (Day 2)
   - `internal/usersync/manifest.go` - JSON I/O
   - Tests for manifest load/save

3. **Core Syncer** (Day 3-4)
   - `internal/usersync/usersync.go` - Syncer type
   - `internal/usersync/collision.go` - Rite collision
   - Tests for sync logic

4. **Resource-Specific** (Day 5)
   - `internal/usersync/hooks.go` - Executable preservation
   - Tests for hooks-specific behavior

5. **CLI Commands** (Day 6-7)
   - `internal/cmd/sync/user.go` - Parent command
   - `internal/cmd/sync/user_agents.go` - Agents
   - `internal/cmd/sync/user_skills.go` - Skills
   - `internal/cmd/sync/user_commands.go` - Commands
   - `internal/cmd/sync/user_hooks.go` - Hooks
   - `internal/cmd/sync/user_all.go` - All

6. **Integration** (Day 8)
   - Integration tests
   - Manifest migration testing
   - Documentation

### 10.2 Dependencies on Existing Packages

| Package | Usage |
|---------|-------|
| `internal/config` | `KnossosHome()` for source resolution |
| `internal/paths` | `EnsureDir()` for directory creation |
| `internal/rite` | `Discovery` for collision detection |
| `internal/output` | `Printer` for CLI output |
| `internal/errors` | Error types and formatting |
| `internal/cmd/common` | `BaseContext` for CLI context |

### 10.3 No New External Dependencies

All functionality uses Go standard library:
- `crypto/sha256` - Checksum computation
- `encoding/json` - Manifest I/O
- `os` / `filepath` - File operations
- `time` - Timestamps

---

## 11. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Manifest format mismatch | Low | Medium | Test with actual shell script manifests |
| Symlink handling edge cases | Low | Low | Resolve symlinks, copy actual content |
| Permission issues on Windows | Medium | Low | Cross-platform ready, defer testing |
| Race condition on manifest write | Low | Medium | Write atomically (temp file + rename) |
| Large file performance | Low | Low | Streaming checksums avoid memory issues |

---

## 12. Handoff Criteria

Ready for Implementation when:

- [x] Package structure defined with interfaces
- [x] Type definitions for all core types (Manifest, Entry, Options, Result)
- [x] Function signatures for public API
- [x] Error codes mapped to exit codes
- [x] Manifest schema backward compatible with shell scripts
- [x] Test scenarios cover critical paths
- [x] Integration points with existing packages identified
- [ ] Principal Engineer can implement without architectural questions
- [ ] All artifacts verified via Read tool

---

## 13. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-usersync-001 | Proposed | Manifest format backward compatibility |
| ADR-usersync-002 | Proposed | Checksum prefix format (sha256:) |
| ADR-usersync-003 | Proposed | Collision detection scope |

---

## 14. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-native-go-user-sync.md` | Write |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-native-go-user-sync.md` | Read |
| Spike | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-native-go-user-sync.md` | Read |
| Materialize Package | `/Users/tomtenuta/Code/roster/internal/materialize/materialize.go` | Read |
| Source Resolution | `/Users/tomtenuta/Code/roster/internal/materialize/source.go` | Read |
| Paths Package | `/Users/tomtenuta/Code/roster/internal/paths/paths.go` | Read |
| Rite Discovery | `/Users/tomtenuta/Code/roster/internal/rite/discovery.go` | Read |
| Config Package | `/Users/tomtenuta/Code/roster/internal/config/home.go` | Read |
| Shell Script (agents) | `/Users/tomtenuta/Code/roster/sync-user-agents.sh` | Read |
| Sync Commands | `/Users/tomtenuta/Code/roster/internal/cmd/sync/sync.go` | Read |
| Materialize Command | `/Users/tomtenuta/Code/roster/internal/cmd/sync/materialize.go` | Read |

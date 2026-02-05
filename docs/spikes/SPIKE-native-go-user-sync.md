# Spike: Native Go Implementation for User-Level Resource Syncing

**Date**: 2026-01-08
**Author**: Claude
**Status**: Complete
**Decision**: Proceed with native Go implementation

---

## Question

What would a complete native Go implementation for syncing `user-agents/`, `user-skills/`, `user-commands/`, and `user-hooks/` to `~/.claude/` entail?

---

## Findings

### Current Shell Scripts Analysis

The existing shell scripts total **3,786 lines** of bash:

| Script | Lines | Complexity |
|--------|-------|------------|
| `sync-user-agents.sh` | 734 | Medium |
| `sync-user-skills.sh` | 997 | High (nested directories) |
| `sync-user-commands.sh` | 959 | High (nested directories) |
| `sync-user-hooks.sh` | 1,096 | Highest (lib/, yaml configs) |

### Core Functionality to Port

Each shell script implements:

1. **Manifest tracking** - JSON manifest at `~/.claude/USER_*_MANIFEST.json`
   - `manifest_version`, `last_sync`, entries with `source`, `installed_at`, `checksum`
   - Three source types: `roster`, `roster-diverged`, `user`

2. **Checksum-based sync** - SHA256 checksums for change detection
   - `calculate_checksum()` using shasum/sha256sum

3. **Rite collision detection** - Prevents syncing items that exist in rites
   - `is_rite_agent()`, `get_rite_for_agent()`

4. **Additive behavior** - Never removes user-created content
   - Only overwrites roster-managed entries

5. **Recovery mode** - Adopts existing files that match roster sources
   - Marks diverged files as `roster-diverged`

6. **Status/dry-run** - Preview changes before applying

---

## Proposed Architecture

### New Package: `internal/usersync`

```
internal/usersync/
├── usersync.go          # Core UserSyncer type
├── manifest.go          # JSON manifest I/O
├── checksum.go          # SHA256 checksums
├── collision.go         # Rite collision detection
├── agents.go            # Agent-specific sync
├── skills.go            # Skill-specific sync (nested)
├── commands.go          # Command-specific sync (nested)
├── hooks.go             # Hook-specific sync (most complex)
└── usersync_test.go     # Tests
```

### Core Type

```go
// Package usersync syncs user-level resources to ~/.claude/
package usersync

type Syncer struct {
    sourceDir    string   // $KNOSSOS_HOME/user-{resource}/
    targetDir    string   // ~/.claude/{resource}/
    manifestPath string   // ~/.claude/USER_{RESOURCE}_MANIFEST.json
    riteChecker  *rite.Discovery
}

type Options struct {
    DryRun     bool
    Recover    bool   // Adopt existing files
    Force      bool   // Overwrite diverged files
}

type Result struct {
    Added     []string
    Updated   []string
    Skipped   []string  // User-created, not touched
    Diverged  []string  // Roster-managed with local changes
    Collisions []string // Would shadow rite resources
}
```

### Manifest Schema (JSON)

```go
type Manifest struct {
    Version   string            `json:"manifest_version"`
    LastSync  time.Time         `json:"last_sync"`
    Entries   map[string]Entry  `json:"entries"`  // agents, skills, etc.
}

type Entry struct {
    Source      string    `json:"source"`      // roster, roster-diverged, user
    InstalledAt time.Time `json:"installed_at"`
    Checksum    string    `json:"checksum"`
}
```

### CLI Commands

Add to `internal/cmd/sync/`:

```go
// user.go - parent command
func newUserCmd(ctx *cmdContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "user",
        Short: "Sync user-level resources to ~/.claude/",
    }
    cmd.AddCommand(newUserAgentsCmd(ctx))
    cmd.AddCommand(newUserSkillsCmd(ctx))
    cmd.AddCommand(newUserCommandsCmd(ctx))
    cmd.AddCommand(newUserHooksCmd(ctx))
    cmd.AddCommand(newUserAllCmd(ctx))
    return cmd
}
```

Usage:
```bash
ari sync user agents [--dry-run] [--recover] [--force]
ari sync user skills [--dry-run] [--recover] [--force]
ari sync user commands [--dry-run] [--recover] [--force]
ari sync user hooks [--dry-run] [--recover] [--force]
ari sync user all [--dry-run] [--recover] [--force]
```

---

## Implementation Estimate

### Phase 1: Core Infrastructure (~300 LOC)
- [ ] `internal/usersync/usersync.go` - Base Syncer type
- [ ] `internal/usersync/manifest.go` - JSON manifest I/O
- [ ] `internal/usersync/checksum.go` - SHA256 checksums
- [ ] `internal/usersync/collision.go` - Rite collision detection

### Phase 2: Resource Syncers (~400 LOC)
- [ ] `internal/usersync/agents.go` - Flat file sync
- [ ] `internal/usersync/skills.go` - Nested directory sync
- [ ] `internal/usersync/commands.go` - Nested directory sync
- [ ] `internal/usersync/hooks.go` - Complex: lib/, yaml configs

### Phase 3: CLI Integration (~200 LOC)
- [ ] `internal/cmd/sync/user.go` - Parent command
- [ ] `internal/cmd/sync/user_agents.go` - Agents subcommand
- [ ] `internal/cmd/sync/user_skills.go` - Skills subcommand
- [ ] `internal/cmd/sync/user_commands.go` - Commands subcommand
- [ ] `internal/cmd/sync/user_hooks.go` - Hooks subcommand

### Phase 4: Testing (~400 LOC)
- [ ] Unit tests for each syncer
- [ ] Integration tests with temporary directories
- [ ] Manifest migration tests (from shell script manifests)

**Total: ~1,300 LOC** (vs 3,786 LOC shell)

---

## Key Implementation Details

### Nested Directory Handling (Skills, Commands, Hooks)

Skills and commands have category subdirectories:
```
user-skills/
├── documentation/
│   ├── doc-artifacts/
│   │   └── skill.md
│   └── standards/
│       └── skill.md
└── orchestration/
    └── ...
```

The manifest must track by full path:
```json
{
  "entries": {
    "documentation/doc-artifacts/skill.md": {...},
    "documentation/standards/skill.md": {...}
  }
}
```

### Hooks Complexity

Hooks require special handling:
1. **lib/** directory - shared shell libraries (recursive copy)
2. **hooks.yaml** - configuration files
3. **Executable preservation** - maintain +x permissions

### KNOSSOS_HOME Resolution

Reuse existing `internal/paths`:
```go
func (p *Resolver) UserAgentsSource() string {
    return filepath.Join(p.KnossosHome(), "user-agents")
}
```

### Backward Compatibility

Read existing shell script manifests and migrate:
```go
func migrateManifest(oldPath, newPath string) error {
    // Read shell script JSON format
    // Write Go format (compatible structure)
}
```

---

## Comparison: Shell vs Go

| Aspect | Shell Scripts | Native Go |
|--------|--------------|-----------|
| **Lines of code** | 3,786 | ~1,300 |
| **Dependencies** | jq, shasum, find | stdlib only |
| **Cross-platform** | macOS/Linux only | All Go platforms |
| **Error handling** | Basic (exit codes) | Rich (error wrapping) |
| **Testing** | Manual | go test |
| **Discoverability** | Hidden scripts | `ari sync user --help` |
| **Manifest format** | JSON | JSON (compatible) |
| **Performance** | Process per file | In-process |

---

## Recommendation

**Proceed with native Go implementation.**

### Rationale

1. **Consistency** - Aligns with existing `ari sync materialize` patterns
2. **Discoverability** - Users find commands via `ari --help`
3. **Maintainability** - Go is easier to test and refactor
4. **Cross-platform** - Works on Windows (future)
5. **Efficiency** - Single binary, no shell subprocess overhead
6. **Debugging** - Better error messages and stack traces

### Implementation Order

1. **Agents first** - Simplest (flat files), validates architecture
2. **Skills second** - Adds nested directory handling
3. **Commands third** - Same as skills
4. **Hooks last** - Most complex, but can reuse patterns

### Migration Strategy

1. Keep shell scripts during transition
2. Add deprecation warnings to shell scripts
3. Remove shell scripts after Go version is stable

---

## Follow-up Actions

1. Create `internal/usersync/` package
2. Implement agents syncer (PoC)
3. Add `ari sync user agents` command
4. Iterate on remaining resource types
5. Add comprehensive tests
6. Deprecate shell scripts

# PRD: Native Go User Sync

> Replace shell-based user resource syncing with native Go implementation.

**Status**: Draft
**Author**: Requirements Analyst (Claude)
**Date**: 2026-01-10
**Initiative**: Native Go User Sync
**Spike Reference**: `docs/spikes/SPIKE-native-go-user-sync.md` (COMPLETE)

---

## Impact Assessment

**impact**: low
**impact_categories**: []

**Rationale**: This is an internal tooling refactor with no user-facing API changes, no schema migrations, and no security implications. The external behavior remains identical; only the implementation changes from shell scripts to Go.

---

## Executive Summary

Replace 3,786 lines of bash scripts across four shell scripts with approximately 1,300 lines of Go code in a new `internal/usersync/` package. This brings user-level resource syncing into the native Ariadne CLI, providing better error handling, cross-platform support, and discoverability via `ari sync user --help`.

### User-Level Resources

| Source | Target | Complexity |
|--------|--------|------------|
| `user-agents/` | `~/.claude/agents/` | Low (flat files) |
| `user-skills/` | `~/.claude/skills/` | Medium (nested directories) |
| `user-commands/` | `~/.claude/commands/` | Medium (nested directories) |
| `user-hooks/` | `~/.claude/hooks/` | High (lib/, yaml configs) |

---

## Background

### Problem Statement

The current shell-based user sync system has limitations:

1. **Distribution**: Requires jq, shasum, find, and consistent shell environments
2. **Performance**: File hashing and manifest diffing are slow in bash
3. **Maintainability**: 3,786 lines of bash across four scripts is hard to test and debug
4. **Cross-platform**: Shell scripts only work on macOS/Linux; Windows support impossible
5. **Discoverability**: Scripts are hidden; users don't know about `sync-user-agents.sh`

### Why Now

The spike (`docs/spikes/SPIKE-native-go-user-sync.md`) confirmed feasibility. The `internal/materialize/` package provides patterns for directory syncing, checksum handling, and manifest tracking that can be reused.

### Who's Affected

- **Knossos platform maintainers**: Simpler, more maintainable codebase
- **End users**: Better error messages, `ari sync user` discoverability
- **Future Windows users**: Cross-platform support enabled

---

## User Stories

### US-1: Sync User Agents

**As a** Knossos user,
**I want** to sync agents from `user-agents/` to `~/.claude/agents/`,
**So that** my custom agents are available globally across all projects.

**Acceptance Criteria**:
- [ ] `ari sync user agents` copies all `.md` files from source to target
- [ ] Files are only updated if checksums differ
- [ ] User-created agents in target are not removed
- [ ] Agents that would shadow rite agents are flagged as collisions
- [ ] Manifest tracks source type: `roster`, `roster-diverged`, `user`

### US-2: Sync User Skills

**As a** Knossos user,
**I want** to sync skills from `user-skills/` to `~/.claude/skills/`,
**So that** my custom skills are available globally across all projects.

**Acceptance Criteria**:
- [ ] `ari sync user skills` recursively copies skill directories
- [ ] Nested category structure is preserved (e.g., `documentation/doc-artifacts/`)
- [ ] Files are only updated if checksums differ
- [ ] User-created skills in target are not removed
- [ ] Skills that would shadow rite skills are flagged as collisions
- [ ] Manifest tracks full relative paths as keys

### US-3: Sync User Commands

**As a** Knossos user,
**I want** to sync commands from `user-commands/` to `~/.claude/commands/`,
**So that** my custom slash commands are available globally across all projects.

**Acceptance Criteria**:
- [ ] `ari sync user commands` recursively copies command directories
- [ ] Nested category structure is preserved (e.g., `operations/commit.md`)
- [ ] Files are only updated if checksums differ
- [ ] User-created commands in target are not removed
- [ ] Commands that would shadow rite commands are flagged as collisions
- [ ] Manifest tracks full relative paths as keys

### US-4: Sync User Hooks

**As a** Knossos user,
**I want** to sync hooks from `user-hooks/` to `~/.claude/hooks/`,
**So that** my custom hooks are available globally across all projects.

**Acceptance Criteria**:
- [ ] `ari sync user hooks` recursively copies hook directories
- [ ] `lib/` directory contents are synced with shared library support
- [ ] `*.yaml` configuration files are synced
- [ ] Executable permissions (+x) are preserved on shell scripts
- [ ] Files are only updated if checksums differ
- [ ] User-created hooks in target are not removed
- [ ] Manifest tracks full relative paths as keys

### US-5: Sync All Resources

**As a** Knossos user,
**I want** to sync all user resources with a single command,
**So that** I don't have to run four separate sync commands.

**Acceptance Criteria**:
- [ ] `ari sync user all` syncs agents, skills, commands, and hooks
- [ ] Failures in one resource type don't prevent syncing others
- [ ] Summary shows results for each resource type
- [ ] Exit code reflects aggregate success/failure

### US-6: Dry Run Preview

**As a** Knossos user,
**I want** to preview what changes would be made before applying them,
**So that** I can verify the sync will do what I expect.

**Acceptance Criteria**:
- [ ] `--dry-run` flag shows what would be added/updated/skipped
- [ ] No files are modified during dry run
- [ ] Output clearly indicates dry run mode

### US-7: Recovery Mode

**As a** Knossos user,
**I want** to adopt existing files that match roster sources,
**So that** I can "adopt" manually-copied files into the manifest system.

**Acceptance Criteria**:
- [ ] `--recover` flag enables recovery mode
- [ ] Existing files matching roster source checksums are adopted
- [ ] Adopted files are tracked as `roster` source
- [ ] Existing files with different checksums are marked `roster-diverged`
- [ ] Recovery results are reported in output

### US-8: Force Overwrite

**As a** Knossos user,
**I want** to force overwrite diverged files,
**So that** I can reset to roster baseline after local experimentation.

**Acceptance Criteria**:
- [ ] `--force` flag overwrites `roster-diverged` files
- [ ] User-created files are never overwritten (even with `--force`)
- [ ] Overwritten files are logged

### US-9: Manifest Migration

**As a** Knossos user who used the shell scripts,
**I want** my existing manifests to be migrated automatically,
**So that** I don't lose my sync state.

**Acceptance Criteria**:
- [ ] Existing `USER_*_MANIFEST.json` files are read on first Go sync
- [ ] All manifest entries are preserved
- [ ] Manifest format remains JSON (no breaking changes)
- [ ] Migration is transparent (no user action required)

---

## Functional Requirements

### Must Have (P0)

| ID | Requirement | User Story |
|----|-------------|------------|
| FR-1 | Sync agents from `user-agents/` to `~/.claude/agents/` | US-1 |
| FR-2 | Sync skills from `user-skills/` to `~/.claude/skills/` with nested directories | US-2 |
| FR-3 | Sync commands from `user-commands/` to `~/.claude/commands/` with nested directories | US-3 |
| FR-4 | Sync hooks from `user-hooks/` to `~/.claude/hooks/` with lib/ and yaml support | US-4 |
| FR-5 | Sync all resources with `ari sync user all` | US-5 |
| FR-6 | Checksum-based change detection using SHA256 | US-1, US-2, US-3, US-4 |
| FR-7 | JSON manifest tracking at `~/.claude/USER_*_MANIFEST.json` | US-1, US-2, US-3, US-4 |
| FR-8 | Source type tracking: `roster`, `roster-diverged`, `user` | US-1, US-2, US-3, US-4 |
| FR-9 | Additive behavior: never remove user-created content | US-1, US-2, US-3, US-4 |
| FR-10 | Rite collision detection: flag shadowing of rite resources | US-1, US-2, US-3, US-4 |
| FR-11 | Dry-run mode: preview without changes | US-6 |
| FR-12 | Preserve executable permissions on shell scripts | US-4 |

### Should Have (P1)

| ID | Requirement | User Story |
|----|-------------|------------|
| FR-13 | Recovery mode: adopt existing files matching roster sources | US-7 |
| FR-14 | Force mode: overwrite diverged files | US-8 |
| FR-15 | Read existing shell script manifests for migration | US-9 |
| FR-16 | JSON output format (`--output=json`) for scripting | US-5 |
| FR-17 | Verbose output (`--verbose`) for debugging | US-5 |

### Could Have (P2)

| ID | Requirement | User Story |
|----|-------------|------------|
| FR-18 | Status command: show current sync state without syncing | - |
| FR-19 | List command: show what would be synced | - |
| FR-20 | Checksum cache for performance optimization | - |

---

## Non-Functional Requirements

| ID | Category | Requirement | Target |
|----|----------|-------------|--------|
| NFR-1 | Performance | Full sync completes in < 5 seconds for typical user directories | < 5s |
| NFR-2 | Performance | Startup time for `ari sync user` < 50ms | < 50ms |
| NFR-3 | Reliability | No data loss: never delete user files, always backup before overwrite | 100% |
| NFR-4 | Maintainability | Code reduction: ~1,300 LOC Go vs 3,786 LOC bash | 65% reduction |
| NFR-5 | Portability | Cross-platform: works on macOS, Linux, Windows | All Go platforms |
| NFR-6 | Compatibility | Manifest format: backward compatible with shell script manifests | 100% |
| NFR-7 | Testability | Unit test coverage for core sync logic | > 80% |
| NFR-8 | Discoverability | Available via `ari sync user --help` | Yes |

---

## Edge Cases

| Case | Expected Behavior |
|------|------------------|
| Source directory doesn't exist | Skip sync for that resource type, log warning |
| Target directory doesn't exist | Create target directory before syncing |
| Source file is a symlink | Resolve symlink and copy actual content |
| Target file is a symlink | Preserve symlink if user-created; otherwise replace |
| File permission mismatch | Preserve source permissions (especially +x for scripts) |
| Manifest file is corrupt | Backup corrupt file, create new manifest, log warning |
| Manifest entry missing for existing target file | Mark as `user` source (user-created) |
| Same filename exists in user and rite | Detect collision, warn user, skip sync |
| Very large file (>10MB) | Sync normally, no special handling |
| File with non-UTF8 encoding | Copy bytes as-is, no encoding conversion |
| Empty source directory | Create empty manifest, no files synced |
| Target file modified but source unchanged | Skip (checksum-based: source matches manifest = no action) |
| Circular symlink | Error and skip file, continue with other files |
| Hidden files (dotfiles) in source | Include in sync (no filtering) |
| KNOSSOS_HOME not set | Use default `$PWD` for source resolution |
| Source file deleted after manifest created | Mark as deleted in manifest, don't remove from target |

---

## Success Criteria

- [ ] All four resource types sync correctly (agents, skills, commands, hooks)
- [ ] Shell scripts can be deleted after Go implementation ships
- [ ] Existing manifests migrate without data loss
- [ ] No user-created files are ever deleted or overwritten without explicit consent
- [ ] Cross-platform support verified (macOS, Linux minimum)
- [ ] Test coverage > 80% for `internal/usersync/` package
- [ ] `ari sync user --help` shows clear usage information
- [ ] Dry-run output is accurate (matches actual sync behavior)

---

## Out of Scope

| Exclusion | Rationale |
|-----------|-----------|
| Project-level syncing | Handled by `ari sync materialize` |
| Automatic sync on shell startup | User configures this separately |
| Conflict resolution UI | Conflicts are flagged; user resolves manually |
| Remote sync (cloud/git) | User syncs from local roster checkout |
| Watch mode (continuous sync) | User runs `ari sync user all` when needed |
| Windows testing in v1 | Cross-platform ready but not actively tested |

---

## CLI Interface Specification

### Command Structure

```
ari sync user
├── agents [--dry-run] [--recover] [--force]
├── skills [--dry-run] [--recover] [--force]
├── commands [--dry-run] [--recover] [--force]
├── hooks [--dry-run] [--recover] [--force]
└── all [--dry-run] [--recover] [--force]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--dry-run` | `-n` | Preview changes without applying | false |
| `--recover` | `-r` | Adopt existing files matching roster sources | false |
| `--force` | `-f` | Overwrite diverged files (roster-diverged only) | false |
| `--output` | `-o` | Output format: text, json | text |
| `--verbose` | `-v` | Enable verbose output | false |

### Example Usage

```bash
# Sync all user resources
ari sync user all

# Preview what would be synced
ari sync user all --dry-run

# Sync only agents
ari sync user agents

# Force overwrite diverged files
ari sync user all --force

# Adopt existing files into manifest
ari sync user all --recover

# JSON output for scripting
ari sync user all --output=json
```

### Output Format (Text)

```
Syncing user agents...
  Added: context-engineer.md
  Updated: consultant.md
  Skipped: my-custom-agent.md (user-created)
  Collision: moirai.md (shadows rite agent)

Syncing user skills...
  Added: documentation/doc-artifacts/SKILL.md
  Unchanged: 47 files

Summary:
  Agents: 2 added, 1 updated, 1 skipped, 1 collision
  Skills: 1 added, 0 updated, 47 unchanged
  Commands: 0 added, 0 updated, 34 unchanged
  Hooks: 3 added, 5 updated, 2 skipped
```

### Output Format (JSON)

```json
{
  "agents": {
    "added": ["context-engineer.md"],
    "updated": ["consultant.md"],
    "skipped": ["my-custom-agent.md"],
    "collisions": ["moirai.md"],
    "unchanged": []
  },
  "skills": { ... },
  "commands": { ... },
  "hooks": { ... }
}
```

---

## Manifest Schema

```json
{
  "manifest_version": "1.0",
  "last_sync": "2026-01-10T12:00:00Z",
  "entries": {
    "context-engineer.md": {
      "source": "roster",
      "installed_at": "2026-01-10T12:00:00Z",
      "checksum": "sha256:abc123..."
    },
    "my-custom-agent.md": {
      "source": "user",
      "installed_at": "2026-01-05T10:00:00Z",
      "checksum": "sha256:def456..."
    }
  }
}
```

### Source Types

| Type | Description |
|------|-------------|
| `roster` | Synced from roster `user-*` directory, checksums match |
| `roster-diverged` | Originally from roster but locally modified (checksum differs) |
| `user` | Created by user, not in roster |

---

## Package Architecture

```
internal/usersync/
├── usersync.go          # Core Syncer type and interface
├── manifest.go          # JSON manifest I/O
├── checksum.go          # SHA256 checksum calculation
├── collision.go         # Rite collision detection
├── agents.go            # Agent-specific sync logic
├── skills.go            # Skill-specific sync (nested directories)
├── commands.go          # Command-specific sync (nested directories)
├── hooks.go             # Hook-specific sync (lib/, yaml, +x)
└── usersync_test.go     # Comprehensive tests

internal/cmd/sync/
├── user.go              # Parent 'ari sync user' command
├── user_agents.go       # 'ari sync user agents' subcommand
├── user_skills.go       # 'ari sync user skills' subcommand
├── user_commands.go     # 'ari sync user commands' subcommand
├── user_hooks.go        # 'ari sync user hooks' subcommand
└── user_all.go          # 'ari sync user all' subcommand
```

---

## Migration Path

### Phase 1: Implementation

1. Create `internal/usersync/` package
2. Implement agents syncer (simplest, validates architecture)
3. Add `ari sync user agents` command
4. Implement skills, commands, hooks syncers
5. Add comprehensive tests

### Phase 2: Coexistence

1. Shell scripts and Go implementation run side-by-side
2. Add deprecation warnings to shell scripts
3. Validate manifest compatibility

### Phase 3: Cutover

1. Remove shell scripts from codebase
2. Update documentation to reference `ari sync user`
3. No migration required for manifests (format unchanged)

---

## Prior Art

Reference `internal/materialize/` package for patterns:
- `materialize.go`: Options, Result types, directory copying
- `source.go`: Source resolution, multi-tier lookup
- Checksum handling via standard library `crypto/sha256`

---

## Open Questions

*None at handoff.*

---

## Appendix: Shell Script Line Counts

| Script | Lines | Status |
|--------|-------|--------|
| `sync-user-agents.sh` | 734 | To be replaced |
| `sync-user-skills.sh` | 997 | To be replaced |
| `sync-user-commands.sh` | 959 | To be replaced |
| `sync-user-hooks.sh` | 1,096 | To be replaced |
| **Total** | **3,786** | |

**Target Go implementation**: ~1,300 LOC (65% reduction)

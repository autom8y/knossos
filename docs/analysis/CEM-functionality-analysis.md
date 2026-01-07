# CEM (Claude Ecosystem Manager) Functionality Analysis

> Comprehensive documentation of CEM functionality for migration planning.
> Source: `/Users/tomtenuta/Code/skeleton_claude/cem` and `lib/` modules.

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Commands Reference](#commands-reference)
4. [Sync Algorithm Deep Dive](#sync-algorithm-deep-dive)
5. [Manifest Structure](#manifest-structure)
6. [Merge Strategies](#merge-strategies)
7. [Conflict Resolution](#conflict-resolution)
8. [Orphan Management](#orphan-management)
9. [Checksum System](#checksum-system)
10. [Function Dependency Graph](#function-dependency-graph)
11. [Data Structures](#data-structures)
12. [Exit Codes](#exit-codes)

---

## Executive Summary

CEM is a 1,232-line bash script orchestrating ecosystem synchronization from a "skeleton" repository to "satellite" projects. It operates with:

- **7 library modules** (3,317 lines total in `lib/`)
- **6 merge strategies** in `lib/cem-merge/`
- **8 commands**: init, sync, validate, validate-team, repair, install-user, status, diff

**Core Purpose**: Physical copy of Claude Code ecosystem files with intelligent merge strategies for settings and documentation, checksum-based change detection, and conflict resolution.

---

## Architecture Overview

### Module Dependency Order

```
cem (main script)
  |
  +-- cem-config.sh (foundation: constants, file lists, exit codes)
  |
  +-- cem-logging.sh (depends: config)
  |
  +-- cem-git.sh (depends: config)
  |
  +-- cem-checksum.sh (depends: config, logging)
  |
  +-- cem-manifest.sh (depends: config, logging)
  |
  +-- cem-merge/dispatcher.sh
  |     +-- merge-dir.sh
  |     +-- merge-settings.sh
  |     +-- merge-docs.sh
  |     +-- copy-replace.sh
  |     +-- merge-init.sh
  |
  +-- cem-sync.sh (depends: all above)
```

### File Classification System

| Strategy | Files | Behavior |
|----------|-------|----------|
| **COPY-REPLACE** | `COMMAND_REGISTRY.md`, `forge-workflow.yaml` | Overwrite completely, skeleton wins |
| **MERGE-SETTINGS** | `settings.local.json` | Union arrays, base skeleton wins conflicts |
| **MERGE-DOCS** | `CLAUDE.md` | Section-based merge with PRESERVE/SYNC markers |
| **MERGE-DIR** | (currently empty, was `skills/`) | Sync skeleton items, preserve satellite-specific |
| **IGNORE** | `ACTIVE_RITE`, `sessions/`, `agents/`, `commands/`, `hooks/`, `skills/` | Never touch (roster-managed) |

---

## Commands Reference

### 1. `cem init [skeleton-path]`

**Purpose**: Initialize a new project with ecosystem files.

**Inputs**:
- `skeleton-path`: Optional path to skeleton (defaults to `SKELETON_HOME` or script directory)
- `--force`: Reinitialize existing project

**Algorithm**:
```
1. Validate skeleton exists and has .claude/
2. Validate project directory (not skeleton, writable)
3. Check jq availability
4. If already initialized and not --force, exit with error
5. Create .claude/ and .claude/.cem/ directories
6. Get skeleton commit hash and ref
7. For each COPY-REPLACE item:
   a. Remove any existing local version
   b. Copy from skeleton with permissions preserved
8. For each MERGE item:
   a. If local doesn't exist, copy from skeleton
   b. If local exists, preserve (will merge on sync)
9. For each MERGE-DIR item:
   a. Call merge_directory() to sync content
10. Build managed_files list with checksums
11. Write manifest.json with:
    - schema_version, skeleton_path, skeleton_commit, skeleton_ref
    - last_sync timestamp, managed_files array
    - Empty local_files and merge_state
```

**Outputs**:
- `.claude/.cem/manifest.json` created
- Ecosystem files copied to `.claude/`
- Exit code 0 on success, 3 on init failure

---

### 2. `cem sync`

**Purpose**: Pull updates from skeleton to satellite.

**Inputs**:
- `--force`: Overwrite local modifications on conflicts
- `--dry-run`: Preview without changes
- `--refresh`: Also refresh active rite from roster
- `--prune`: Remove orphaned resources
- `--auto-refresh`: Automatically refresh team if roster has updates

**Algorithm**: See [Sync Algorithm Deep Dive](#sync-algorithm-deep-dive)

**Outputs**:
- Updated files in `.claude/`
- Updated manifest.json with new checksums
- `.cem-backup` files on conflicts
- Exit code 0 (success), 4 (sync failure), 5 (conflicts), 6 (orphan conflicts)

---

### 3. `cem validate`

**Purpose**: Validate manifest and file integrity.

**Inputs**: None (reads existing manifest)

**Algorithm**:
```
1. Check manifest.json exists
2. Validate schema_version matches CEM_SCHEMA_VERSION
3. Check required fields: skeleton_path, skeleton_commit, last_sync, managed_files
4. Verify skeleton path exists and has .claude/
5. For each managed file in manifest:
   a. Check file exists (warn if missing = orphan entry)
   b. Compute current checksum
   c. Compare with manifest checksum (info on mismatch = local modification)
6. Count .cem-backup files (warn if unresolved)
7. Validate CLAUDE.md structure (## Quick Start, ## Agent Configurations)
8. Report summary: managed files, missing, local changes, backups
```

**Outputs**:
- Summary report to stdout
- Exit code 0 (valid), 1 (warnings), 2 (validation failure)

---

### 4. `cem repair`

**Purpose**: Rebuild manifest from current `.claude/` state.

**Algorithm**:
```
1. Check .claude/ directory exists
2. If existing manifest:
   a. Remove entries for non-existent files (orphan cleanup)
   b. Preserve skeleton_path, skeleton_commit, skeleton_ref
3. If no skeleton info, try CEM_SKELETON_PATH or default
4. Scan COPY-REPLACE items, compute checksums, add to manifest
5. Scan MERGE items, compute checksums, add to manifest
6. Scan MERGE-DIR items, compute checksums, add to manifest
7. Write new manifest with repaired_at timestamp
```

**Outputs**:
- Repaired manifest.json
- Exit code 0 on success

---

### 5. `cem status`

**Purpose**: Show sync status and version info.

**Algorithm**:
```
1. Read manifest: skeleton_path, last_commit, last_ref, last_sync
2. If in worktree, display worktree info
3. Display skeleton path, last sync time
4. If skeleton is git repo:
   a. Get current commit
   b. Compare with last_commit
   c. If different, show "Updates available!" and commit count
   d. If same, show "Up to date"
5. Display managed file count
```

**Outputs**: Status report to stdout

---

### 6. `cem diff [path]`

**Purpose**: Show differences between local and skeleton.

**Inputs**:
- `path`: Optional specific file/directory to diff

**Algorithm**:
```
If path specified:
  - Diff specific file/directory against skeleton
Else:
  - For each COPY-REPLACE item:
    a. Diff against skeleton version
    b. Show files that differ
```

**Outputs**: Diff output to stdout

---

### 7. `cem validate-team [name]`

**Purpose**: Validate rite against workflow schema.

**Inputs**:
- `name`: Team name (defaults to ACTIVE_RITE)

**Algorithm**:
```
1. Read rite name from arg or .claude/ACTIVE_RITE
2. Check workflow.yaml exists at $ROSTER_HOME/rites/$name/
3. Validate YAML syntax via yq
4. Check required fields: name, workflow_type, description, entry_point, phases, complexity_levels
5. Verify name matches directory name
6. Check entry_point.agent matches phases[0].agent
7. For each agent in phases, verify .md file exists
8. Validate phase graph:
   a. All next values reference valid phase names
   b. Exactly one terminal phase (next: null)
9. Validate complexity_levels reference valid phases
```

**Outputs**:
- Check/fail list for each validation
- Exit code 0 (valid), 2 (validation failure)

---

### 8. `cem install-user`

**Purpose**: Install user-level resources to `~/.claude/`.

**Algorithm**:
```
1. Set roster_home from ROSTER_HOME
2. Run sync-user-agents.sh (if exists)
3. Run sync-user-skills.sh (if exists)
4. Run sync-user-commands.sh (if exists)
5. Run sync-user-hooks.sh (if exists)
```

**Outputs**:
- Resources installed to `~/.claude/`
- Manifest files created: `USER_{AGENT,SKILL,COMMAND,HOOKS}_MANIFEST.json`

---

## Sync Algorithm Deep Dive

The sync algorithm is the core of CEM, implemented in `lib/cem-sync.sh`.

### High-Level Flow

```
cmd_sync()
    |
    +-- 1. Validate & Setup
    |       - validate_initialized()
    |       - check_jq()
    |       - validate_skeleton()
    |       - check_skeleton_dirty()
    |
    +-- 2. Check Version
    |       - Compare current_commit vs last_commit
    |       - If same and not --prune: exit early
    |
    +-- 3. Process COPY-REPLACE Items
    |       - For each item in get_copy_replace_items()
    |       - Compute skeleton_checksum, manifest_checksum, local_checksum
    |       - Apply three-way classification
    |
    +-- 4. Process MERGE Items
    |       - For each item in get_merge_items()
    |       - If skeleton changed, call merge strategy
    |
    +-- 5. Process MERGE-DIR Items
    |       - For each item in get_merge_dir_items()
    |       - Call merge_directory() for recursive sync
    |
    +-- 6. Update Manifest
    |       - Update skeleton_commit, skeleton_ref, last_sync
    |       - Update checksums for all managed files
    |
    +-- 7. Orphan Detection (Phase 2)
    |       - detect_orphans()
    |       - detect_orphan_conflicts()
    |       - If --prune: backup_orphans(), prune_orphans()
    |
    +-- 8. Team Freshness
    |       - check_team_freshness()
    |       - If --refresh or --auto-refresh: refresh_active_rite()
```

### Three-Way Classification Logic

For each file, CEM computes three checksums:

| Checksum | Source |
|----------|--------|
| `skeleton_checksum` | Current skeleton file |
| `manifest_checksum` | Last synced version (stored in manifest) |
| `local_checksum` | Current local file |

**Decision Matrix**:

| Skeleton Changed? | Local Changed? | Action |
|-------------------|----------------|--------|
| No | No | Skip (up to date) |
| No | Yes | Skip (preserve local) |
| Yes | No | Update (safe) |
| Yes | Yes | **CONFLICT** |

**Conflict Handling**:
1. Create backup: `{file}.cem-backup`
2. If `--force`: Overwrite with skeleton version
3. If not forced: Skip, increment conflict count, exit 5 at end

### Merge Strategy Dispatch

When a MERGE item's skeleton changes:

```bash
case "$strategy" in
    merge-settings)
        merge_settings_json "$src" "$dst" "$dst.tmp"
        mv "$dst.tmp" "$dst"
        ;;
    merge-docs)
        merge_documentation "$src" "$dst" "$dst.tmp"
        mv "$dst.tmp" "$dst"
        ;;
esac
```

---

## Manifest Structure

### Schema Version 1 (Original)

```json
{
  "schema_version": 1,
  "skeleton_path": "/path/to/skeleton_claude",
  "skeleton_commit": "abc123...",
  "skeleton_ref": "main",
  "last_sync": "2025-01-01T00:00:00Z",
  "managed_files": [
    {
      "path": ".claude/COMMAND_REGISTRY.md",
      "strategy": "copy-replace",
      "checksum": "sha256..."
    }
  ],
  "local_files": [],
  "merge_state": {}
}
```

### Schema Version 2 (Unified with Team Layer)

```json
{
  "schema_version": 2,
  "skeleton": {
    "path": "/path/to/skeleton_claude",
    "commit": "abc123...",
    "ref": "main",
    "last_sync": "2025-01-01T00:00:00Z"
  },
  "team": {
    "name": "10x-dev",
    "checksum": "sha256...",
    "last_refresh": "2025-01-01T00:00:00Z",
    "roster_path": "/path/to/roster/rites/10x-dev"
  },
  "managed_files": [
    {
      "path": ".claude/COMMAND_REGISTRY.md",
      "strategy": "copy-replace",
      "checksum": "sha256...",
      "source": "skeleton",
      "added_at": "2025-01-01T00:00:00Z",
      "last_sync": "2025-01-01T00:00:00Z"
    }
  ],
  "local_files": []
}
```

### V1 to V2 Migration

The `migrate_manifest_v1_to_v2()` function:
1. Creates backup at `manifest.v1.backup.json`
2. Extracts v1 flat fields
3. Transforms to nested structure
4. Adds provenance fields (`source`, `added_at`, `last_sync`) to managed_files
5. Reads team info from ACTIVE_RITE if present

---

## Merge Strategies

### 1. COPY-REPLACE (`copy-replace.sh`)

**Files**: `COMMAND_REGISTRY.md`, `forge-workflow.yaml`

**Behavior**:
- Complete overwrite from skeleton
- No merge logic
- Creates backup on conflict

### 2. MERGE-SETTINGS (`merge-settings.sh`)

**Files**: `settings.local.json`

**Algorithm**:
```
1. If no project file, copy skeleton as-is
2. Extract permissions.allow from both, find project-specific extras
3. Extract permissions.additionalDirectories extras
4. Extract enabledMcpjsonServers extras
5. Start with skeleton base
6. Add project-specific permissions (union)
7. Add project-specific directories (union)
8. Add project-specific MCP servers (union)
9. Preserve project enableAllProjectMcpServers if set
```

**Key Principle**: Skeleton base settings win; project additions are preserved.

### 3. MERGE-DOCS (`merge-docs.sh`)

**Files**: `CLAUDE.md`

**Markers**:
- `<!-- SYNC: skeleton-owned -->` - Always take skeleton version
- `<!-- PRESERVE: satellite-owned -->` - Keep satellite version

**Algorithm**:
```
1. If no project file, copy skeleton as-is
2. Extract header/preamble from skeleton (before first ##)
3. For each skeleton section:
   a. Check for SYNC marker in skeleton -> use skeleton
   b. Check for PRESERVE marker in satellite -> use satellite
   c. Fallback: "## Quick Start" and "## Agent Configurations" -> preserve
   d. Otherwise -> sync from skeleton
4. For PRESERVE sections without satellite content:
   a. Check for ACTIVE_RITE
   b. Regenerate from agents/ directory if possible
5. Append satellite-only sections (not in skeleton)
6. Append ## Project:* sections from satellite
```

**Regeneration Functions**:
- `regenerate_quick_start()`: Creates agent table from `.claude/agents/*.md`
- `regenerate_agent_configurations()`: Lists agents with descriptions

### 4. MERGE-DIR (`merge-dir.sh`)

**Directories**: (Currently empty, was used for `skills/`)

**Algorithm**:
```
1. Load exclusions from .rite-skills-exclusions
2. Create destination if needed
3. For each item in source:
   a. Skip if in exclusion list
   b. If directory: recurse
   c. If file:
      - New file: copy
      - Existing file: compare checksums, update if different
4. Satellite-only items are automatically preserved
```

**Key Feature**: Never deletes satellite-specific content.

### 5. MERGE-INIT (`merge-init.sh`)

**Purpose**: Handle merge items during initialization.

**Algorithm**:
```
1. For each merge item:
   a. If local doesn't exist, copy from skeleton
   b. If local exists, preserve (merge on future sync)
```

### 6. Dispatcher (`dispatcher.sh`)

Routes merge operations based on strategy name:

```bash
dispatch_merge_strategy() {
    case "$strategy" in
        copy-replace)  # Handled inline in cmd_sync
        merge-settings) merge_settings_json "$src" "$dst" "$output_file" ;;
        merge-docs)     merge_documentation "$src" "$dst" "$output_file" ;;
        merge-dir)      merge_directory "$src" "$dst" 0 >/dev/null ;;
        merge-init)     # Copy if not present ;;
    esac
}
```

---

## Conflict Resolution

### Detection

Conflict occurs when:
- `skeleton_checksum != manifest_checksum` (skeleton changed)
- AND `local_checksum != manifest_checksum` (local changed)

### Resolution Flow

```
1. Create backup: {file}.cem-backup
2. Log warning: "Conflict: {file} (local modified, skeleton updated)"
3. Increment conflict counter
4. If --force: Apply skeleton version anyway
5. If not forced: Skip file, continue sync

At end of sync:
- If conflicts > 0 and not forced:
  - Exit with code 5 (EXIT_CONFLICTS)
  - Print resolution instructions
```

### Resolution Instructions

```
To resolve:
  1. Review .cem-backup files and decide which changes to keep
  2. Run 'cem sync --force' to overwrite local changes
  3. Or manually merge changes and run 'cem sync' again
```

---

## Orphan Management

### Definition

An **orphan** is a file that:
1. Exists in the manifest
2. Was previously synced from skeleton or team
3. No longer exists in source (deleted from skeleton/team)
4. Still exists locally

### Detection Algorithm

```python
def detect_orphans(skeleton):
    orphans = []

    # Collect current skeleton paths into set
    skeleton_path_set = list_skeleton_managed_paths(skeleton)

    # Collect team paths if team is active
    team_path_set = list_team_managed_paths(active_rite) if active_rite else {}

    # Check skeleton-sourced resources
    for entry in manifest.managed_files:
        if entry.source == "skeleton":
            if entry.path not in skeleton_path_set:
                if file_exists(entry.path):
                    orphans.append({
                        path: entry.path,
                        source: "skeleton",
                        reason: "deleted from skeleton"
                    })

    # Check team-sourced resources
    for entry in manifest.managed_files:
        if entry.source == "team":
            if entry.path not in team_path_set:
                if file_exists(entry.path):
                    orphans.append({
                        path: entry.path,
                        source: "team:{name}",
                        reason: "deleted from rite"
                    })

    return orphans
```

### Orphan Conflict Detection

An orphan has a **conflict** if local checksum differs from manifest checksum (local modification after skeleton deleted it).

### Prune Operation

With `--prune`:

```
1. Detect orphans
2. Detect orphan conflicts
3. Report orphans and conflicts

If conflicts and not --force:
   Exit code 6 (EXIT_ORPHAN_CONFLICTS)

Otherwise:
   4. backup_orphans() -> .claude/.cem/orphan-backup/{timestamp}/
   5. prune_orphans() -> rm files, remove_managed_file() from manifest
```

### Backup Structure

```
.claude/.cem/orphan-backup/
  20250101-120000/
    manifest.json    # Backup metadata
    COMMAND_REGISTRY.md  # Backed up files
    skills/
      some-skill.md
```

---

## Checksum System

### Implementation (`cem-checksum.sh`)

**Cross-Platform Detection**:
```bash
detect_checksum_cmd() {
    if command -v shasum &>/dev/null; then
        CHECKSUM_CMD="shasum -a 256"  # macOS
    elif command -v sha256sum &>/dev/null; then
        CHECKSUM_CMD="sha256sum"      # Linux
    else
        CHECKSUM_CMD="echo unavailable #"
    fi
}
```

### File Checksum

```bash
compute_checksum() {
    local file="$1"
    get_cached_checksum "$file"  # Uses caching
}

compute_checksum_raw() {
    local file="$1"
    $CHECKSUM_CMD "$file" | cut -d' ' -f1
}
```

### Directory Checksum

```bash
compute_dir_checksum() {
    local dir="$1"
    find "$dir" -type f -exec $CHECKSUM_CMD {} \; | sort | $CHECKSUM_CMD | cut -d' ' -f1
}
```

### Caching System

**Cache File**: `.claude/.cem/checksum-cache.json`

**Cache Entry**:
```json
{
  "version": 1,
  "entries": {
    ".claude/COMMAND_REGISTRY.md": {
      "checksum": "sha256...",
      "mtime": 1704067200,
      "size": 1234
    }
  },
  "updated_at": "2025-01-01T00:00:00Z"
}
```

**Cache Logic**:
1. On cache miss: compute checksum, store in temp file
2. On lookup: check mtime+size match; if match, return cached checksum
3. On exit: merge new entries into cache file

---

## Function Dependency Graph

### Core Flow

```
main()
  +-- parse_global_flags()
  +-- init_checksum_cache()
  |
  +-- cmd_init()
  |     +-- validate_skeleton()
  |     +-- validate_project()
  |     +-- get_copy_replace_items()
  |     +-- get_merge_items()
  |     +-- get_merge_dir_items()
  |     +-- compute_checksum() / compute_dir_checksum()
  |     +-- merge_directory()
  |
  +-- cmd_sync()
  |     +-- validate_initialized()
  |     +-- read_manifest()
  |     +-- validate_skeleton()
  |     +-- get_skeleton_commit()
  |     +-- compute_checksum() / compute_dir_checksum()
  |     +-- merge_settings_json()
  |     +-- merge_documentation()
  |     +-- merge_directory()
  |     +-- detect_orphans()
  |     +-- detect_orphan_conflicts()
  |     +-- backup_orphans()
  |     +-- prune_orphans()
  |     +-- check_team_freshness()
  |     +-- refresh_active_rite()
  |
  +-- save_checksum_cache()
```

### Merge Strategy Dependencies

```
dispatch_merge_strategy()
  +-- merge_settings_json()
  |     +-- jq (external)
  |
  +-- merge_documentation()
  |     +-- extract_section()
  |     +-- has_preserve_marker()
  |     +-- has_sync_marker()
  |     +-- is_preserve_section()
  |     +-- regenerate_quick_start()
  |     +-- regenerate_agent_configurations()
  |
  +-- merge_directory()
        +-- load_merge_exclusions()
        +-- is_excluded()
        +-- compute_checksum()
```

---

## Data Structures

### File Classification Lists (cem-config.sh)

```bash
# COPY-REPLACE: Complete overwrite
get_copy_replace_items() {
    cat <<EOF
COMMAND_REGISTRY.md
forge-workflow.yaml
EOF
}

# MERGE: Intelligent merge
get_merge_items() {
    cat <<EOF
settings.local.json:merge-settings
CLAUDE.md:merge-docs
EOF
}

# MERGE-DIR: Directory sync
get_merge_dir_items() {
    cat <<EOF
EOF  # Currently empty
}

# IGNORE: Never touch
get_ignore_items() {
    cat <<EOF
ACTIVE_RITE
ACTIVE_WORKFLOW.yaml
sessions
agents
agents.backup
.cem
.archive
user-agents
user-commands
user-skills
user-hooks
commands
skills
hooks
PROJECT.md
EOF
}
```

### Constants

```bash
readonly CEM_VERSION="1.0.0"
readonly CEM_SCHEMA_VERSION=1
readonly CEM_STATE_DIR=".claude/.cem"
readonly CHECKSUM_CACHE_FILE="$CEM_STATE_DIR/checksum-cache.json"
readonly CHECKSUM_CACHE_VERSION=1
```

---

## Exit Codes

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | `EXIT_SUCCESS` | Success |
| 1 | `EXIT_INVALID_ARGS` | Invalid arguments |
| 2 | `EXIT_VALIDATION_FAILURE` | Validation failure |
| 3 | `EXIT_INIT_FAILURE` | Init failure |
| 4 | `EXIT_SYNC_FAILURE` | Sync failure |
| 5 | `EXIT_CONFLICTS` | Conflicts detected (use --force or resolve manually) |
| 6 | `EXIT_ORPHAN_CONFLICTS` | Orphan conflicts detected (with --prune, use --force) |

---

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `SKELETON_HOME` | Override skeleton location | Script directory |
| `CEM_SKELETON_PATH` | Alternative to SKELETON_HOME | Script directory |
| `ROSTER_HOME` | Override roster location | `~/Code/roster` |
| `CEM_DEBUG` | Enable debug output | 0 |

---

## External Dependencies

| Tool | Usage | Required? |
|------|-------|-----------|
| `jq` | JSON manipulation | Yes |
| `yq` | YAML validation (validate-team) | For validate-team |
| `git` | Version tracking | Recommended |
| `shasum` or `sha256sum` | Checksums | Yes |

---

## Key Implementation Insights

### 1. Checksum Caching
CEM uses mtime+size caching to avoid recomputing checksums. New entries are written to a temp file (survives subshells) and merged on exit.

### 2. Atomic Manifest Updates
Manifest updates use temp file + mv pattern for atomicity.

### 3. Provenance Tracking
V2 manifest tracks `source` field ("skeleton", "team", "user") for each managed file, enabling proper orphan detection.

### 4. Section-Based CLAUDE.md Merge
CLAUDE.md merge uses AWK-based section extraction with marker-based ownership. The `<!-- PRESERVE: -->` and `<!-- SYNC: -->` markers control which version wins.

### 5. Team Freshness
CEM can detect when roster rites have updates via checksum comparison, enabling auto-refresh via `--auto-refresh` flag.

---

## Migration Considerations

When implementing a replacement:

1. **Manifest Compatibility**: Support both v1 and v2 schemas, with migration path
2. **Checksum Algorithm**: Use same SHA-256 algorithm for seamless transition
3. **Merge Strategies**: Replicate exact behavior of settings and docs merging
4. **Conflict Detection**: Three-way checksum comparison is critical
5. **Orphan Management**: Provenance tracking enables orphan detection
6. **Backup Strategy**: Always backup before destructive operations

---

*Analysis generated: 2026-01-03*
*Source: skeleton_claude CEM v1.0.0 (schema v1/v2)*

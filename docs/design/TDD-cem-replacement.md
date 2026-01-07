---
artifact_id: TDD-cem-replacement
title: "CEM Replacement Architecture - Roster-Native Ecosystem Management"
created_at: "2026-01-03T04:00:00Z"
author: architect
prd_ref: PRD-skeleton-deprecation
status: draft
components:
  - name: roster-sync
    type: script
    description: "Main entry point for roster-native ecosystem synchronization"
    dependencies:
      - name: sync-lib
        type: internal
      - name: jq
        type: external
        version: "1.6+"
  - name: sync-lib
    type: library
    description: "Library modules for sync operations (checksum, manifest, merge)"
    dependencies:
      - name: jq
        type: external
      - name: shasum/sha256sum
        type: external
  - name: merge-strategies
    type: library
    description: "Merge strategy implementations for settings, docs, directories"
    dependencies:
      - name: sync-lib
        type: internal
      - name: jq
        type: external
related_adrs:
  - ADR-0007
  - ADR-0008
  - ADR-0009
schema_version: "1.0"
---

# TDD-cem-replacement: CEM Replacement Architecture

> Technical Design Document for roster-native ecosystem management replacing skeleton_claude CEM.

**Initiative**: Skeleton Deprecation & CEM Migration
**Sprint**: 1 - CEM Migration Planning
**Session**: session-20260103-031208-2c671d71
**Complexity**: SYSTEM

---

## 1. Executive Summary

This TDD specifies the architecture for replacing CEM (Claude Ecosystem Manager) with roster-native functionality. The design achieves:

- **Complete skeleton independence**: No `$SKELETON_HOME` dependencies
- **Manifest compatibility**: Existing v1/v2 manifests migrate seamlessly to v3
- **Feature parity**: All 8 CEM commands have roster equivalents
- **Merge strategy preservation**: Exact replication of settings and docs merge behavior
- **Risk mitigation**: Addresses all 5 HIGH severity risks from assessment

### Key Architectural Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Approach | Port + Enhance | Minimizes risk, maintains compatibility |
| Entry Point | Enhance swap-rite.sh | Leverage existing infrastructure |
| Manifest | Schema v3 with migration | Forward-compatible, preserves history |
| Library Location | `$ROSTER_HOME/lib/sync/` | Clear separation, easy testing |

### Scope

**In Scope**:
- Replace all 8 CEM commands with roster-native equivalents
- Maintain manifest backward compatibility (v1, v2 -> v3 migration)
- Preserve all 6 merge strategies
- Support existing flags: --refresh, --prune, --force, --dry-run
- Worktree initialization without skeleton

**Out of Scope**:
- User-level resource sync (handled by existing sync-user-*.sh scripts)
- Team validation (already in swap-rite.sh)
- New features beyond CEM parity

---

## 2. Architecture Overview

### 2.1 System Context

```
                                    ┌─────────────────────────────────────┐
                                    │           ROSTER ECOSYSTEM           │
                                    │                                     │
  ┌─────────────────┐              │  ┌─────────────────────────────┐    │
  │   swap-rite.sh  │──────────────┼─▶│      roster-sync            │    │
  │  (team swaps)   │              │  │   (ecosystem sync)          │    │
  └─────────────────┘              │  └──────────────┬──────────────┘    │
                                    │                │                   │
  ┌─────────────────┐              │                ▼                   │
  │worktree-manager │──────────────┼─▶┌─────────────────────────────┐    │
  │ (parallel work) │              │  │        lib/sync/            │    │
  └─────────────────┘              │  │                             │    │
                                    │  │  ├── sync-config.sh        │    │
  ┌─────────────────┐              │  │  ├── sync-checksum.sh       │    │
  │   /sync command │──────────────┼─▶│  ├── sync-manifest.sh       │    │
  │  (user-facing)  │              │  │  ├── sync-core.sh           │    │
  └─────────────────┘              │  │  └── merge/                 │    │
                                    │  │       ├── dispatcher.sh    │    │
                                    │  │       ├── merge-settings.sh│    │
                                    │  │       ├── merge-docs.sh    │    │
                                    │  │       └── merge-dir.sh     │    │
                                    │  └─────────────────────────────┘    │
                                    │                │                   │
                                    │                ▼                   │
                                    │  ┌─────────────────────────────┐    │
                                    │  │     .claude/.cem/           │    │
                                    │  │  manifest.json (v3)         │    │
                                    │  │  checksum-cache.json        │    │
                                    │  │  orphan-backup/             │    │
                                    │  └─────────────────────────────┘    │
                                    └─────────────────────────────────────┘
```

### 2.2 Component Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           roster-sync (main)                              │
│                                                                          │
│  Commands: init | sync | validate | repair | status | diff               │
│  Flags: --refresh --prune --force --dry-run --auto-refresh               │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────┐         ┌───────────────┐         ┌───────────────┐
│ sync-config.sh│         │sync-manifest.sh│        │sync-checksum.sh│
│               │         │               │         │               │
│ - Constants   │         │ - Read/write  │         │ - SHA-256     │
│ - File lists  │         │ - Migration   │         │ - Caching     │
│ - Exit codes  │         │ - Validation  │         │ - Cross-plat  │
└───────────────┘         └───────────────┘         └───────────────┘
        │                           │                           │
        └───────────────────────────┼───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │        sync-core.sh           │
                    │                               │
                    │ - Three-way classification    │
                    │ - Conflict detection          │
                    │ - Orphan management           │
                    │ - Team freshness              │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │     merge/dispatcher.sh       │
                    │                               │
                    │ Routes by strategy:           │
                    │ - copy-replace                │
                    │ - merge-settings              │
                    │ - merge-docs                  │
                    │ - merge-dir                   │
                    └───────────────────────────────┘
```

### 2.3 Data Flow: Sync Operation

```
                    roster-sync sync --refresh
                            │
                            ▼
            ┌───────────────────────────────┐
            │  1. Load & Validate Manifest  │
            │  - Migrate v1/v2 if needed    │
            │  - Check schema version       │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  2. Check Source Version      │
            │  - Compare roster commit      │
            │  - Skip if unchanged (no -f)  │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  3. Process COPY-REPLACE      │
            │  - Three-way checksum         │
            │  - Classify: skip/update/     │
            │    conflict                   │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  4. Process MERGE Items       │
            │  - Dispatch to strategy       │
            │  - merge-settings.sh          │
            │  - merge-docs.sh              │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  5. Update Manifest           │
            │  - New checksums              │
            │  - Timestamp                  │
            │  - Roster commit              │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  6. Orphan Detection          │
            │  - Compare manifest vs source │
            │  - Report orphans             │
            │  - Prune if --prune           │
            └───────────────┬───────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  7. Team Refresh (--refresh)  │
            │  - Check team freshness       │
            │  - Call swap-rite.sh --update │
            └───────────────────────────────┘
```

---

## 3. Component Design

### 3.1 roster-sync (Main Script)

**Location**: `$ROSTER_HOME/roster-sync`

**Purpose**: Unified entry point replacing CEM executable.

#### Command Matrix

| Command | Replaces | Description | Exit Codes |
|---------|----------|-------------|------------|
| `init` | `cem init` | Initialize project with ecosystem | 0, 1, 3 |
| `sync` | `cem sync` | Pull updates from roster | 0, 4, 5, 6 |
| `validate` | `cem validate` | Validate manifest integrity | 0, 1, 2 |
| `repair` | `cem repair` | Rebuild manifest from state | 0, 1 |
| `status` | `cem status` | Show sync status | 0 |
| `diff` | `cem diff` | Show differences with roster | 0 |

#### Flag Mapping

| Flag | Behavior | Commands |
|------|----------|----------|
| `--force` | Overwrite local changes | sync, init |
| `--dry-run` | Preview without changes | sync, init |
| `--refresh` | Also refresh active rite | sync |
| `--prune` | Remove orphaned files | sync |
| `--auto-refresh` | Auto-refresh if team updates | sync |

#### Script Skeleton

```bash
#!/usr/bin/env bash
#
# roster-sync - Roster-native ecosystem synchronization
#
# Replaces CEM (Claude Ecosystem Manager) with roster-native functionality.
# See TDD-cem-replacement for design details.

set -euo pipefail

readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
readonly SYNC_LIB="$ROSTER_HOME/lib/sync"
readonly ROSTER_SYNC_VERSION="1.0.0"
readonly ROSTER_SYNC_SCHEMA_VERSION=3

# Source library modules in dependency order
source "$SYNC_LIB/sync-config.sh"
source "$SYNC_LIB/sync-checksum.sh"
source "$SYNC_LIB/sync-manifest.sh"
source "$SYNC_LIB/sync-core.sh"
source "$SYNC_LIB/merge/dispatcher.sh"

# Parse global flags
parse_global_flags "$@"

# Initialize checksum cache
init_checksum_cache

# Command dispatch
case "${1:-}" in
    init)           shift; cmd_init "$@" ;;
    sync)           shift; cmd_sync "$@" ;;
    validate)       shift; cmd_validate "$@" ;;
    repair)         shift; cmd_repair "$@" ;;
    status)         shift; cmd_status "$@" ;;
    diff)           shift; cmd_diff "$@" ;;
    version|--version|-v)  echo "roster-sync $ROSTER_SYNC_VERSION" ;;
    help|--help|-h) show_help ;;
    *)              show_help; exit 1 ;;
esac

# Save checksum cache
save_checksum_cache

exit $?
```

### 3.2 init Command

**Purpose**: Initialize a project with roster ecosystem files.

**Algorithm**:

```
cmd_init()
  1. Validate roster exists at $ROSTER_HOME
  2. Validate project directory (not roster itself, writable)
  3. Check jq availability
  4. If already initialized and not --force:
     - Exit with error (use --force to reinitialize)
  5. Create .claude/ and .claude/.cem/ directories
  6. Get roster commit hash and ref
  7. For each COPY-REPLACE item:
     a. Copy from $ROSTER_HOME/.claude/{file}
     b. Preserve permissions
  8. For each MERGE item:
     a. If local doesn't exist: copy from roster
     b. If local exists: preserve (merge on future sync)
  9. Build managed_files list with checksums
  10. Write manifest.json v3 with:
      - schema_version: 3
      - roster_path, roster_commit, roster_ref
      - last_sync timestamp
      - managed_files array
      - Empty orphans array
```

**Key Difference from CEM**: Sources files from `$ROSTER_HOME/.claude/` not `$SKELETON_HOME/.claude/`.

### 3.3 sync Command

**Purpose**: Pull updates from roster to satellite project.

**Algorithm**: See Section 4 (Sync Algorithm) for detailed specification.

### 3.4 validate Command

**Purpose**: Validate manifest and file integrity.

**Algorithm**:

```
cmd_validate()
  1. Check manifest.json exists
  2. Validate schema_version (supports 1, 2, 3)
  3. Check required fields based on schema version
  4. For each managed file in manifest:
     a. Check file exists (warn if missing = orphan entry)
     b. Compute current checksum
     c. Compare with manifest (info on mismatch = local mod)
  5. Count .cem-backup files (warn if unresolved)
  6. Validate CLAUDE.md structure (markers intact)
  7. Report summary
```

**Exit Codes**:
- 0: Valid, no warnings
- 1: Valid with warnings
- 2: Validation failure

### 3.5 repair Command

**Purpose**: Rebuild manifest from current file state.

**Algorithm**:

```
cmd_repair()
  1. Check .claude/ directory exists
  2. If existing manifest:
     a. Backup to manifest.repair-backup.json
     b. Preserve roster_path, roster_commit, roster_ref
     c. Remove entries for non-existent files
  3. If no roster info, use ROSTER_HOME
  4. Scan COPY-REPLACE items:
     - Compute checksums
     - Add to managed_files
  5. Scan MERGE items:
     - Compute checksums
     - Add to managed_files
  6. Write new manifest with:
     - repaired_at timestamp
     - schema_version: 3
```

### 3.6 status Command

**Purpose**: Show sync status and version information.

**Algorithm**:

```
cmd_status()
  1. Read manifest fields
  2. Display:
     - Roster path
     - Last sync timestamp
     - Managed file count
  3. If roster is git repo:
     a. Get current commit
     b. Compare with last_commit
     c. If different: "Updates available!" + commit count
     d. If same: "Up to date"
  4. If active rite:
     a. Show rite name
     b. Check team freshness
```

### 3.7 diff Command

**Purpose**: Show differences between local and roster.

**Algorithm**:

```
cmd_diff()
  If path specified:
    - Diff specific file against roster version
  Else:
    - For each COPY-REPLACE item:
      a. Diff against roster version
      b. Show files that differ
```

---

## 4. Sync Algorithm

### 4.1 Three-Way Classification

The core sync algorithm uses three checksums to determine action:

| Checksum | Source | Description |
|----------|--------|-------------|
| `roster_checksum` | Current roster file | What roster has now |
| `manifest_checksum` | Last synced version | What we synced last time |
| `local_checksum` | Current local file | What satellite has now |

### 4.2 Decision Matrix

```
┌─────────────────────┬─────────────────────┬────────────────────────────────┐
│ Roster Changed?     │ Local Changed?      │ Action                         │
│ (roster != manifest)│ (local != manifest) │                                │
├─────────────────────┼─────────────────────┼────────────────────────────────┤
│ No                  │ No                  │ SKIP (up to date)              │
├─────────────────────┼─────────────────────┼────────────────────────────────┤
│ No                  │ Yes                 │ SKIP (preserve local)          │
├─────────────────────┼─────────────────────┼────────────────────────────────┤
│ Yes                 │ No                  │ UPDATE (safe to overwrite)     │
├─────────────────────┼─────────────────────┼────────────────────────────────┤
│ Yes                 │ Yes                 │ CONFLICT (both changed)        │
└─────────────────────┴─────────────────────┴────────────────────────────────┘
```

### 4.3 Conflict Resolution

When CONFLICT detected:

```
1. Create backup: {file}.cem-backup
2. Log: "Conflict: {file} (local modified, roster updated)"
3. Increment conflict counter
4. If --force:
   - Apply roster version
   - Log: "Forced: {file}"
5. If not --force:
   - Skip file
   - Continue sync
6. At end:
   - If conflicts > 0 and not --force:
     - Print resolution instructions
     - Exit with EXIT_CONFLICTS (5)
```

### 4.4 Sync Algorithm Pseudocode

```python
def cmd_sync(flags):
    # 1. Validate & Setup
    validate_initialized()
    check_jq()
    manifest = read_manifest()
    manifest = migrate_manifest_if_needed(manifest)

    # 2. Check Version (skip if no changes unless --force)
    current_commit = get_roster_commit()
    if current_commit == manifest.roster_commit and not flags.force:
        if not flags.prune:
            log("Already up to date")
            return EXIT_SUCCESS

    # 3. Process COPY-REPLACE Items
    conflicts = 0
    for item in get_copy_replace_items():
        roster_checksum = compute_checksum(roster_path(item))
        manifest_checksum = get_manifest_checksum(item)
        local_checksum = compute_checksum(local_path(item))

        action = classify(roster_checksum, manifest_checksum, local_checksum)

        if action == SKIP:
            continue
        elif action == UPDATE:
            copy_file(roster_path(item), local_path(item))
        elif action == CONFLICT:
            create_backup(local_path(item))
            conflicts += 1
            if flags.force:
                copy_file(roster_path(item), local_path(item))

    # 4. Process MERGE Items
    for item, strategy in get_merge_items():
        roster_checksum = compute_checksum(roster_path(item))
        manifest_checksum = get_manifest_checksum(item)

        if roster_checksum != manifest_checksum:
            dispatch_merge(strategy, roster_path(item), local_path(item))

    # 5. Update Manifest
    manifest.roster_commit = current_commit
    manifest.roster_ref = get_roster_ref()
    manifest.last_sync = now()
    update_managed_file_checksums(manifest)
    write_manifest(manifest)

    # 6. Orphan Detection
    if flags.prune:
        orphans = detect_orphans(manifest)
        orphan_conflicts = detect_orphan_conflicts(orphans)

        if orphan_conflicts and not flags.force:
            return EXIT_ORPHAN_CONFLICTS

        backup_orphans(orphans)
        prune_orphans(orphans, manifest)

    # 7. Team Refresh
    if flags.refresh:
        refresh_active_team()
    elif flags.auto_refresh:
        if is_team_stale():
            refresh_active_team()

    # 8. Return
    if conflicts > 0 and not flags.force:
        return EXIT_CONFLICTS
    return EXIT_SUCCESS
```

---

## 5. Merge Strategy Implementations

### 5.1 Strategy Registry

| Strategy | Files | Implementation |
|----------|-------|----------------|
| `copy-replace` | COMMAND_REGISTRY.md, forge-workflow.yaml | Complete overwrite |
| `merge-settings` | settings.local.json | JSON union merge |
| `merge-docs` | CLAUDE.md | Section-based with markers |
| `merge-dir` | (reserved) | Directory sync |
| `merge-init` | (init only) | Copy if missing |

### 5.2 File Classification Lists

```bash
# sync-config.sh

# Files that are completely replaced from roster
get_copy_replace_items() {
    cat <<EOF
COMMAND_REGISTRY.md
forge-workflow.yaml
EOF
}

# Files that use intelligent merging
get_merge_items() {
    cat <<EOF
settings.local.json:merge-settings
CLAUDE.md:merge-docs
EOF
}

# Directories that sync contents (preserving satellite-specific)
get_merge_dir_items() {
    cat <<EOF
EOF
}

# Files/directories never touched by sync
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

### 5.3 merge-settings.sh

**Purpose**: Merge settings.local.json with union semantics.

**Algorithm**:

```
merge_settings_json(roster_file, local_file, output_file)
  1. If no local file:
     - Copy roster as-is
     - Return

  2. Extract arrays from both files:
     - roster_permissions = roster.permissions.allow[]
     - local_permissions = local.permissions.allow[]
     - roster_dirs = roster.permissions.additionalDirectories[]
     - local_dirs = local.permissions.additionalDirectories[]
     - roster_mcp = roster.enabledMcpjsonServers[]
     - local_mcp = local.enabledMcpjsonServers[]

  3. Compute satellite extras:
     - extra_permissions = local_permissions - roster_permissions
     - extra_dirs = local_dirs - roster_dirs
     - extra_mcp = local_mcp - roster_mcp

  4. Build merged result:
     - Start with roster file as base
     - Union extra_permissions into permissions.allow
     - Union extra_dirs into additionalDirectories
     - Union extra_mcp into enabledMcpjsonServers
     - Preserve local.enableAllProjectMcpServers if set

  5. Write to output_file
```

**jq Implementation**:

```bash
merge_settings_json() {
    local roster_file="$1"
    local local_file="$2"
    local output_file="$3"

    # If no local file, copy roster
    if [[ ! -f "$local_file" ]]; then
        cp "$roster_file" "$output_file"
        return 0
    fi

    # Extract satellite-specific permissions
    local extra_perms
    extra_perms=$(jq -n --slurpfile r "$roster_file" --slurpfile l "$local_file" '
        ($l[0].permissions.allow // []) - ($r[0].permissions.allow // [])
    ')

    # Extract satellite-specific directories
    local extra_dirs
    extra_dirs=$(jq -n --slurpfile r "$roster_file" --slurpfile l "$local_file" '
        ($l[0].permissions.additionalDirectories // []) -
        ($r[0].permissions.additionalDirectories // [])
    ')

    # Extract satellite-specific MCP servers
    local extra_mcp
    extra_mcp=$(jq -n --slurpfile r "$roster_file" --slurpfile l "$local_file" '
        ($l[0].enabledMcpjsonServers // []) -
        ($r[0].enabledMcpjsonServers // [])
    ')

    # Merge: roster base + satellite extras
    jq --argjson ep "$extra_perms" \
       --argjson ed "$extra_dirs" \
       --argjson em "$extra_mcp" \
       --slurpfile l "$local_file" '
        .permissions.allow = ((.permissions.allow // []) + $ep | unique) |
        .permissions.additionalDirectories = ((.permissions.additionalDirectories // []) + $ed | unique) |
        .enabledMcpjsonServers = ((.enabledMcpjsonServers // []) + $em | unique) |
        if $l[0].enableAllProjectMcpServers then
            .enableAllProjectMcpServers = $l[0].enableAllProjectMcpServers
        else . end
    ' "$roster_file" > "$output_file"
}
```

### 5.4 merge-docs.sh

**Purpose**: Merge CLAUDE.md with section-based ownership.

**Markers**:
- `<!-- SYNC: roster-owned -->` - Always take roster version
- `<!-- PRESERVE: satellite-owned -->` - Keep satellite version

**Algorithm**:

```
merge_documentation(roster_file, local_file, output_file)
  1. If no local file:
     - Copy roster as-is
     - Return

  2. Extract header/preamble from roster (before first ##)

  3. Build section list from roster

  4. For each roster section:
     a. If roster has SYNC marker -> use roster section
     b. If local has PRESERVE marker -> use local section
     c. Fallback sections (## Quick Start, ## Agent Configurations):
        -> preserve local if exists
     d. Otherwise -> sync from roster

  5. For PRESERVE sections without local content:
     a. Check for ACTIVE_RITE
     b. Regenerate from agents/ directory if possible

  6. Append satellite-only sections (not in roster)

  7. Append ## Project:* sections from local

  8. Write to output_file
```

**Regeneration Functions**:

```bash
# Regenerate ## Quick Start from agents/ directory
regenerate_quick_start() {
    local agents_dir="$1"
    local rite_name="$2"

    echo "## Quick Start"
    echo ""
    echo "This project uses a $(count_agents "$agents_dir")-agent workflow ($rite_name):"
    echo ""
    echo "| Agent | Role | Produces |"
    echo "| ----- | ---- | -------- |"

    for agent_file in "$agents_dir"/*.md; do
        local name=$(basename "$agent_file" .md)
        local role=$(extract_agent_role "$agent_file")
        local produces=$(extract_agent_produces "$agent_file")
        echo "| **$name** | $role | $produces |"
    done
}

# Regenerate ## Agent Configurations from agents/ directory
regenerate_agent_configurations() {
    local agents_dir="$1"

    echo "## Agent Configurations"
    echo ""
    echo "Full agent prompts live in \`.claude/agents/\`:"
    echo ""

    for agent_file in "$agents_dir"/*.md; do
        local name=$(basename "$agent_file" .md)
        local desc=$(head -20 "$agent_file" | grep -A1 "^#" | tail -1 | cut -c1-80)
        echo "- \`$name.md\` - $desc"
    done
}
```

### 5.5 merge-dir.sh

**Purpose**: Sync directory contents while preserving satellite-specific files.

**Algorithm**:

```
merge_directory(roster_dir, local_dir, depth)
  1. Load exclusions from .sync-exclusions

  2. Create local_dir if needed

  3. For each item in roster_dir:
     a. Skip if in exclusion list
     b. If directory: recurse
     c. If file:
        - New file: copy from roster
        - Existing file: compare checksums
          - If different: update from roster
          - If same: skip

  4. Satellite-only items are preserved (never deleted)
```

### 5.6 Dispatcher

```bash
# merge/dispatcher.sh

dispatch_merge_strategy() {
    local strategy="$1"
    local roster_file="$2"
    local local_file="$3"
    local output_file="$4"

    case "$strategy" in
        copy-replace)
            cp "$roster_file" "$output_file"
            ;;
        merge-settings)
            merge_settings_json "$roster_file" "$local_file" "$output_file"
            ;;
        merge-docs)
            merge_documentation "$roster_file" "$local_file" "$output_file"
            ;;
        merge-dir)
            merge_directory "$roster_file" "$local_file" 0
            ;;
        merge-init)
            if [[ ! -f "$local_file" ]]; then
                cp "$roster_file" "$output_file"
            fi
            ;;
        *)
            log_error "Unknown merge strategy: $strategy"
            return 1
            ;;
    esac
}
```

---

## 6. Manifest Schema

### 6.1 Schema Version 3 (New)

```json
{
  "schema_version": 3,
  "roster": {
    "path": "/path/to/roster",
    "commit": "abc123...",
    "ref": "main",
    "last_sync": "2026-01-03T00:00:00Z"
  },
  "team": {
    "name": "10x-dev-pack",
    "checksum": "sha256...",
    "last_refresh": "2026-01-03T00:00:00Z",
    "roster_path": "/path/to/roster/rites/10x-dev-pack"
  },
  "managed_files": [
    {
      "path": ".claude/COMMAND_REGISTRY.md",
      "strategy": "copy-replace",
      "checksum": "sha256...",
      "source": "roster",
      "added_at": "2026-01-01T00:00:00Z",
      "last_sync": "2026-01-03T00:00:00Z"
    }
  ],
  "orphans": [],
  "migration": {
    "migrated_from": 2,
    "migrated_at": "2026-01-03T00:00:00Z",
    "skeleton_path": "/original/skeleton_claude"
  }
}
```

### 6.2 Migration Path

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Schema v1   │────▶│ Schema v2   │────▶│ Schema v3   │
│ (skeleton)  │     │ (skeleton)  │     │ (roster)    │
└─────────────┘     └─────────────┘     └─────────────┘
     │                   │                    │
     │                   │                    │
     └───────────────────┴────────────────────┘
                         │
                         ▼
                 migrate_manifest()
```

**Migration Function**:

```bash
migrate_manifest_if_needed() {
    local manifest_file="$1"
    local schema_version

    schema_version=$(jq -r '.schema_version // 1' "$manifest_file")

    case "$schema_version" in
        1) migrate_v1_to_v3 "$manifest_file" ;;
        2) migrate_v2_to_v3 "$manifest_file" ;;
        3) return 0 ;;  # Already current
        *) log_error "Unknown schema version: $schema_version"; return 1 ;;
    esac
}

migrate_v1_to_v3() {
    local manifest_file="$1"

    # Backup original
    cp "$manifest_file" "${manifest_file}.v1.backup"

    # Transform v1 flat structure to v3 nested
    jq '{
        schema_version: 3,
        roster: {
            path: .skeleton_path,
            commit: .skeleton_commit,
            ref: .skeleton_ref,
            last_sync: .last_sync
        },
        team: (if .team then {
            name: .rite.name,
            checksum: .rite.checksum,
            last_refresh: .rite.last_refresh,
            roster_path: .rite.roster_path
        } else null end),
        managed_files: [.managed_files[] | . + {
            source: "roster",
            added_at: (.added_at // .last_sync),
            last_sync: .last_sync
        }],
        orphans: [],
        migration: {
            migrated_from: 1,
            migrated_at: (now | todate),
            skeleton_path: .skeleton_path
        }
    }' "$manifest_file" > "${manifest_file}.tmp"

    mv "${manifest_file}.tmp" "$manifest_file"
}

migrate_v2_to_v3() {
    local manifest_file="$1"

    # Backup original
    cp "$manifest_file" "${manifest_file}.v2.backup"

    # Transform v2 skeleton structure to v3 roster
    jq '{
        schema_version: 3,
        roster: {
            path: .skeleton.path,
            commit: .skeleton.commit,
            ref: .skeleton.ref,
            last_sync: .skeleton.last_sync
        },
        team: .team,
        managed_files: .managed_files,
        orphans: [],
        migration: {
            migrated_from: 2,
            migrated_at: (now | todate),
            skeleton_path: .skeleton.path
        }
    }' "$manifest_file" > "${manifest_file}.tmp"

    mv "${manifest_file}.tmp" "$manifest_file"
}
```

### 6.3 Backward Compatibility

- **v1 manifests**: Auto-migrated on first roster-sync operation
- **v2 manifests**: Auto-migrated on first roster-sync operation
- **Original skeleton_path**: Preserved in `migration.skeleton_path` for reference
- **No data loss**: All original fields preserved or transformed

---

## 7. Migration Path (From CEM to Roster-Sync)

### 7.1 Phase 1: Infrastructure Setup (Sprint 2)

```
1. Create $ROSTER_HOME/lib/sync/ directory structure
2. Port library modules:
   - sync-config.sh (from cem-config.sh)
   - sync-checksum.sh (from cem-checksum.sh)
   - sync-manifest.sh (from cem-manifest.sh)
   - sync-core.sh (from cem-sync.sh)
3. Port merge strategies:
   - merge/dispatcher.sh
   - merge/merge-settings.sh
   - merge/merge-docs.sh
   - merge/merge-dir.sh
4. Create roster-sync main script
5. Update paths: $SKELETON_HOME -> $ROSTER_HOME
```

### 7.2 Phase 2: Integration (Sprint 3)

```
1. Update worktree-manager.sh:
   - Replace: $SKELETON_HOME/cem -> $ROSTER_HOME/roster-sync
   - Add fallback to CEM during transition
2. Update /sync command:
   - Point to roster-sync
   - Preserve all flags
3. Update environment handling:
   - ROSTER_HOME primary
   - SKELETON_HOME deprecated fallback
```

### 7.3 Phase 3: Validation (Sprint 4)

```
1. Test manifest migration (v1, v2 -> v3)
2. Test all merge strategies
3. Test conflict detection and resolution
4. Test orphan management
5. Test worktree creation
6. Cross-platform testing (macOS, Linux)
```

### 7.4 Phase 4: Documentation (Sprint 5)

```
1. Update /sync command documentation
2. Update INTEGRATION.md
3. Update orchestrator-templates paths
4. Create migration guide for users
5. Update troubleshooting docs
```

---

## 8. Rollback Procedures

### 8.1 Sprint 2 Rollback: Library Port

**Trigger**: Library modules fail tests or break functionality.

**Procedure**:
```
1. Identify failing module(s)
2. Revert to CEM library calls:
   - source "$SKELETON_HOME/lib/cem-{module}.sh"
3. Document failure for investigation
4. No manifest changes needed (not in production yet)
```

**Recovery Time**: < 1 hour

### 8.2 Sprint 3 Rollback: Integration

**Trigger**: worktree-manager.sh or /sync command breaks.

**Procedure**:
```
1. Revert worktree-manager.sh to previous version:
   git checkout HEAD~1 -- user-hooks/lib/worktree-manager.sh
2. Revert /sync command:
   git checkout HEAD~1 -- user-commands/cem/sync.md
3. Ensure SKELETON_HOME still valid
4. Test worktree creation
```

**Recovery Time**: < 30 minutes

### 8.3 Sprint 4 Rollback: Manifest Migration

**Trigger**: Manifest migration corrupts data or loses information.

**Procedure**:
```
1. Restore manifest backup:
   mv .claude/.cem/manifest.json.v2.backup .claude/.cem/manifest.json
2. Point back to CEM:
   export CEM_HOME="$SKELETON_HOME"
3. Run: $SKELETON_HOME/cem validate
4. If valid: continue with CEM
5. If invalid: manual repair
```

**Recovery Time**: < 1 hour per satellite

### 8.4 Sprint 5 Rollback: Full Rollback

**Trigger**: Critical regression requiring complete reversion.

**Procedure**:
```
1. Create rollback branch:
   git checkout -b rollback-cem-migration
2. Revert all roster-sync changes:
   git revert --no-commit <sprint-2-commits>..<sprint-5-commits>
3. Restore SKELETON_HOME requirement in config.sh
4. Restore CEM calls in worktree-manager.sh
5. Update /sync to point to CEM
6. Test full workflow:
   - Worktree creation
   - Team swap
   - Sync operation
7. Communicate rollback to users
```

**Recovery Time**: 2-4 hours

### 8.5 Rollback Checkpoints

| Sprint | Checkpoint | Verification |
|--------|------------|--------------|
| 2 | lib/sync/ modules pass unit tests | `./test-sync-lib.sh` |
| 3 | worktree creation works | `/worktree create test` |
| 3 | /sync command works | `/sync --dry-run` |
| 4 | Manifest migration preserves data | `roster-sync validate` |
| 4 | Conflict detection works | Manual test with conflicts |
| 5 | Full workflow E2E | QA test suite |

---

## 9. Test Plan

### 9.1 Unit Tests

| Module | Test File | Coverage |
|--------|-----------|----------|
| sync-config.sh | test-sync-config.sh | File lists, constants |
| sync-checksum.sh | test-sync-checksum.sh | SHA-256, caching, cross-platform |
| sync-manifest.sh | test-sync-manifest.sh | Read, write, migrate |
| sync-core.sh | test-sync-core.sh | Classification, conflict detection |
| merge-settings.sh | test-merge-settings.sh | Union semantics |
| merge-docs.sh | test-merge-docs.sh | Section extraction, markers |

### 9.2 Integration Tests

| Test | Description | Commands |
|------|-------------|----------|
| init-fresh | Initialize fresh project | `roster-sync init` |
| init-force | Reinitialize existing | `roster-sync init --force` |
| sync-clean | Sync with no local changes | `roster-sync sync` |
| sync-conflict | Sync with conflicts | `roster-sync sync` (expect exit 5) |
| sync-force | Force sync over conflicts | `roster-sync sync --force` |
| sync-refresh | Sync with team refresh | `roster-sync sync --refresh` |
| sync-prune | Sync with orphan removal | `roster-sync sync --prune` |
| validate-clean | Validate good manifest | `roster-sync validate` |
| validate-corrupt | Validate bad manifest | `roster-sync validate` (expect exit 2) |
| worktree-create | Create worktree with sync | `/worktree create test-wt` |

### 9.3 Migration Tests

| Test | Description | Verification |
|------|-------------|--------------|
| migrate-v1 | v1 manifest migration | Fields preserved, schema=3 |
| migrate-v2 | v2 manifest migration | Fields preserved, schema=3 |
| migrate-backup | Backup created | .v1.backup or .v2.backup exists |
| migrate-skeleton-path | skeleton_path preserved | migration.skeleton_path set |

### 9.4 Merge Strategy Tests

| Test | Input | Expected Output |
|------|-------|-----------------|
| settings-empty-local | roster only | Copy roster |
| settings-extra-perms | local has extras | Union preserved |
| settings-extra-dirs | local has extras | Union preserved |
| settings-mcp-servers | local has extras | Union preserved |
| docs-no-local | roster only | Copy roster |
| docs-sync-marker | roster has SYNC | Use roster section |
| docs-preserve-marker | local has PRESERVE | Use local section |
| docs-quick-start | preserve section | Regenerate if empty |
| docs-project-section | local has Project:* | Append to output |

### 9.5 Cross-Platform Tests

| Platform | Checksum Tool | Test |
|----------|---------------|------|
| macOS | shasum -a 256 | Full test suite |
| Linux | sha256sum | Full test suite |
| CI (GitHub Actions) | Both | Integration tests |

---

## 10. Risk Mitigations

### 10.1 RISK-BC-001: Worktree Creation Breakage

**Mitigation**:
1. Update worktree-manager.sh to call roster-sync instead of CEM
2. Add fallback during transition:
   ```bash
   if [[ -x "$ROSTER_HOME/roster-sync" ]]; then
       "$ROSTER_HOME/roster-sync" sync
   elif [[ -x "$SKELETON_HOME/cem" ]]; then
       log_warning "Using legacy CEM (roster-sync not available)"
       "$SKELETON_HOME/cem" sync
   else
       log_error "No sync mechanism available"
       exit 1
   fi
   ```
3. Test worktree creation before removing CEM fallback

### 10.2 RISK-FN-001: CLAUDE.md Merge Complexity

**Mitigation**:
1. Port merge-docs.sh exactly as-is (line-for-line)
2. Create comprehensive test suite before porting
3. Add verbose logging during transition:
   ```bash
   if [[ "$ROSTER_SYNC_DEBUG" == "1" ]]; then
       log_debug "Section: $section_name"
       log_debug "  Roster marker: $roster_marker"
       log_debug "  Local marker: $local_marker"
       log_debug "  Decision: $decision"
   fi
   ```
4. Implement --dry-run for merge operations
5. Keep backup of original CLAUDE.md before merge

### 10.3 RISK-FN-002: Three-Way Checksum Conflict Detection

**Mitigation**:
1. Use identical SHA-256 algorithm:
   ```bash
   # Cross-platform checksum
   if command -v shasum &>/dev/null; then
       CHECKSUM_CMD="shasum -a 256"
   elif command -v sha256sum &>/dev/null; then
       CHECKSUM_CMD="sha256sum"
   fi
   ```
2. Maintain checksum cache format compatibility
3. Test all 4 classification states:
   - No change (roster=manifest, local=manifest)
   - Local only (roster=manifest, local!=manifest)
   - Roster only (roster!=manifest, local=manifest)
   - Conflict (roster!=manifest, local!=manifest)
4. Add explicit classification logging:
   ```bash
   log_debug "File: $file"
   log_debug "  roster_checksum: $roster_checksum"
   log_debug "  manifest_checksum: $manifest_checksum"
   log_debug "  local_checksum: $local_checksum"
   log_debug "  classification: $classification"
   ```

### 10.4 RISK-FN-005: Missing Skills Migration

**Note**: This risk is addressed in Sprint 3 (Resource Migration), not in this TDD. However, roster-sync must support team refresh to pull migrated skills:

```bash
# Support for --refresh flag to update team resources
if [[ "$REFRESH_FLAG" == "1" ]] && [[ -f ".claude/ACTIVE_RITE" ]]; then
    local rite_name
    rite_name=$(cat ".claude/ACTIVE_RITE")
    log "Refreshing team: $rite_name"
    "$ROSTER_HOME/swap-rite.sh" "$rite_name" --update
fi
```

### 10.5 RISK-EX-001: Migration Ordering Dependencies

**Mitigation**:
1. Follow strict ordering:
   ```
   Sprint 2: lib/sync/ (foundation)
   Sprint 3: roster-sync main + integration
   Sprint 4: manifest migration + validation
   Sprint 5: documentation + deprecation
   ```
2. Validate each sprint before proceeding:
   - Sprint 2: Unit tests pass
   - Sprint 3: Integration tests pass
   - Sprint 4: Migration tests pass
   - Sprint 5: E2E tests pass
3. Maintain rollback capability between sprints

---

## 11. Integration with swap-rite.sh

### 11.1 Current swap-rite.sh Functions

swap-rite.sh already handles:
- Agent swapping from roster teams
- Manifest management (AGENT_MANIFEST.json)
- Backup/restore operations
- Orphan detection for agents

### 11.2 Integration Points

| Function | swap-rite.sh | roster-sync | Interaction |
|----------|--------------|-------------|-------------|
| Team swap | Primary | None | swap-rite.sh owns team changes |
| Agent sync | Primary | Trigger | roster-sync --refresh calls swap-rite.sh |
| CEM sync | None | Primary | roster-sync owns ecosystem files |
| Manifest | AGENT_MANIFEST.json | manifest.json | Separate manifests |
| Worktree | Consumer | Consumer | Both use worktree-manager.sh |

### 11.3 Refresh Flow

```
roster-sync sync --refresh
        │
        ├──▶ Sync ecosystem files (roster-sync)
        │
        └──▶ Refresh team (swap-rite.sh --update)
                │
                ├──▶ Sync agents from roster
                │
                ├──▶ Sync skills from roster
                │
                └──▶ Sync commands from roster
```

### 11.4 Future Enhancement: Unified Manifest

Consider future consolidation:
```json
{
  "schema_version": 4,
  "ecosystem": {
    "roster_commit": "...",
    "last_sync": "..."
  },
  "team": {
    "name": "10x-dev-pack",
    "agents": [...],
    "skills": [...],
    "commands": [...]
  },
  "managed_files": [...]
}
```

This is out of scope for current TDD but noted for future consideration.

---

## 12. Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-cem-replacement.md` | Created |
| Source: CEM Analysis | `/Users/tomtenuta/Code/roster/docs/analysis/CEM-functionality-analysis.md` | Read |
| Source: Integration Points | `/Users/tomtenuta/Code/roster/docs/analysis/CEM-integration-points.md` | Read |
| Source: References Audit | `/Users/tomtenuta/Code/roster/docs/audits/skeleton-references-audit.md` | Read |
| Source: Risk Assessment | `/Users/tomtenuta/Code/roster/docs/assessments/skeleton-migration-risks.md` | Read |
| Source: Gap Analysis | `/Users/tomtenuta/Code/roster/docs/ecosystem/GAP-ANALYSIS-skeleton-deprecation.md` | Read |

---

## 13. Acceptance Criteria Checklist

- [x] All 8 CEM commands have replacement design (Section 3)
- [x] Manifest schema backward compatible (Section 6)
- [x] Sync algorithm documented with decision matrix (Section 4)
- [x] Each merge strategy has implementation approach (Section 5)
- [x] Rollback procedure defined for each sprint (Section 8)
- [x] Test infrastructure requirements specified (Section 9)
- [x] Integration with swap-rite.sh clarified (Section 11)
- [x] Addresses all 5 HIGH severity risks (Section 10)

---

## 14. Related ADRs (To Be Created)

| ADR | Title | Decision |
|-----|-------|----------|
| ADR-0007 | Roster-Native Sync Approach | Port CEM vs rewrite -> Port |
| ADR-0008 | Manifest Schema v3 Design | Nested structure with migration tracking |
| ADR-0009 | CLAUDE.md Merge Marker Ownership | SYNC=roster-owned, PRESERVE=satellite-owned |

---

*Generated: 2026-01-03*
*Author: Architect Agent*
*Initiative: Skeleton Deprecation & CEM Migration*

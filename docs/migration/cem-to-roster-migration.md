# CEM to roster-sync Migration Guide

> Migration guide for transitioning from Claude Ecosystem Manager (CEM) to roster-native synchronization.

**Version**: 1.0.0
**Last Updated**: 2026-01-03
**Applies To**: roster-sync v1.0.0+, Manifest Schema v3

---

## Important: Artifact Architecture

Before migrating, understand where artifacts belong:

| Source (roster/) | Target | Scope |
|------------------|--------|-------|
| `user-agents/` | `~/.claude/agents/` | User (global) |
| `mena/` | `~/.claude/commands/` | User (global) |
| `user-skills/` | `~/.claude/skills/` | User (global) |
| `user-hooks/` | `~/.claude/hooks/` | User (global) |
| `.knossos/rites/{pack}/` | `.claude/` | Project |

**Critical**: NO `.claude/user-*` directories should exist in satellite projects. If you see `.claude/user-agents/`, `.claude/user-commands/`, `.claude/user-skills/`, or `.claude/user-hooks/` in a satellite project, these are stale migration artifacts from skeleton deprecation and should be removed.

See [INTEGRATION.md](../INTEGRATION.md) for full architecture details.

---

## 1. Overview

### What is Being Replaced

The **Claude Ecosystem Manager (CEM)** from skeleton_claude is being replaced by **roster-sync**, a roster-native ecosystem synchronization tool. This migration eliminates the dependency on `$SKELETON_HOME` and consolidates ecosystem management within roster itself.

| Component | Before (CEM) | After (roster-sync) |
|-----------|--------------|---------------------|
| Executable | `$SKELETON_HOME/cem` | `$ROSTER_HOME/roster-sync` |
| Config Variable | `SKELETON_HOME` | `ROSTER_HOME` |
| Manifest Schema | v1 or v2 | v3 |
| Manifest Location | `.claude/.cem/manifest.json` | `.claude/.cem/manifest.json` (unchanged) |

### Why Migrate

1. **Native Integration**: roster-sync is built into roster, eliminating external skeleton_claude dependency
2. **Fewer Dependencies**: No need to maintain or update a separate skeleton repository
3. **Simplified Architecture**: Single source of truth for ecosystem files
4. **Improved Conflict Detection**: Three-way checksum comparison for safer syncs
5. **Better Backup System**: Timestamped backups for conflicts and orphaned files
6. **Enhanced Repair Capability**: New `repair` command rebuilds manifest from current state

### Benefits of roster-sync

- **Automatic manifest migration**: v1 and v2 manifests auto-migrate to v3
- **Timestamped backups**: All conflict and orphan operations create dated backups
- **Dry-run support**: Preview changes with `--dry-run` before applying
- **Team integration**: `--refresh` and `--auto-refresh` flags for team synchronization
- **Orphan management**: `--prune` flag safely removes orphaned files with backup
- **Manifest repair**: `repair` command rebuilds corrupted or missing manifests

---

## 2. Command Mapping Table

| CEM Command | roster-sync Equivalent | Notes |
|-------------|------------------------|-------|
| `cem init <path>` | `roster-sync init [path]` | Path defaults to current directory |
| `cem sync` | `roster-sync sync` | Same behavior, additional flags available |
| `cem status` | `roster-sync status` | Same output format |
| `cem validate` | `roster-sync validate` | New `--team` flag for team validation |
| `cem diff [file]` | `roster-sync diff [file]` | Same behavior |
| (none) | `roster-sync repair` | **NEW** - Rebuilds manifest from current state |

### Flag Mapping

| CEM Flag | roster-sync Flag | Available On |
|----------|------------------|--------------|
| `--force` | `--force`, `-f` | init, sync, repair |
| `--dry-run` | `--dry-run`, `-n` | init, sync, repair |
| `--refresh` | `--refresh`, `-r` | sync |
| `--prune` | `--prune`, `-p` | sync |
| (none) | `--auto-refresh` | sync |
| (none) | `--team`, `-t` | validate, init |
| (none) | `--verbose`, `-V` | validate |
| (none) | `--debug`, `-d` | all |

---

## 3. Step-by-Step Migration

### Pre-Migration Checklist

Before migrating, complete these steps:

- [ ] **Backup your manifest**: Copy `.claude/.cem/manifest.json` to a safe location
- [ ] **Note current sync state**: Run `cem status` and save the output
- [ ] **Check for uncommitted changes**: Ensure `.claude/` directory changes are committed
- [ ] **Verify ROSTER_HOME is set**: Run `echo $ROSTER_HOME` (should point to roster repository)
- [ ] **Ensure roster is up to date**: Pull latest changes from roster repository

### Migration Steps

#### Step 1: Ensure roster is up to date

```bash
cd $ROSTER_HOME
git pull origin main
```

#### Step 2: Validate existing manifest

Run validation to check the current state:

```bash
roster-sync validate
```

Expected output for a healthy manifest:
```
[roster-sync] Validating manifest...
[roster-sync] Validation Summary:
  Tracked files:      4
  Missing files:      0
  Local modifications: 0
  Warnings:           0
  Errors:             0
[roster-sync] Validation passed
```

If warnings appear about schema version, this is expected and will be resolved in the next step.

#### Step 3: Sync with auto-migration

Run sync to trigger automatic manifest migration:

```bash
roster-sync sync --force
```

The `--force` flag ensures a clean sync even if there are no roster updates. During this operation:
- v1/v2 manifests are automatically migrated to v3
- A backup of the original manifest is created (`.v1.backup` or `.v2.backup`)
- All managed files are updated with current checksums

#### Step 4: Verify migration

Confirm the migration succeeded:

```bash
roster-sync status
```

Expected output:
```
roster-sync status
==================

Schema Version: 3
Roster Path:    /path/to/roster
Last Sync:      2026-01-03T00:00:00Z
Managed Files:  4

Up to date

Active Team: 10x-dev
```

Run validation again to confirm integrity:

```bash
roster-sync validate --verbose
```

### Post-Migration Verification

After migration, verify:

- [ ] **Schema version is 3**: `jq '.schema_version' .claude/.cem/manifest.json` returns `3`
- [ ] **Backup files exist**: Check for `.claude/.cem/manifest.json.v1.backup` or `.v2.backup`
- [ ] **No validation errors**: `roster-sync validate` exits with code 0
- [ ] **Status shows up to date**: `roster-sync status` shows "Up to date"
- [ ] **Team is preserved**: Active team still appears in status (if applicable)

---

## 4. Manifest Migration

### Schema Evolution

The manifest has evolved through three schema versions:

| Schema | Structure | Source Variable |
|--------|-----------|-----------------|
| v1 | Flat (skeleton_path, skeleton_commit) | `SKELETON_HOME` |
| v2 | Nested (skeleton.path, skeleton.commit) | `SKELETON_HOME` |
| v3 | Nested (roster.path, roster.commit) | `ROSTER_HOME` |

### Automatic Migration

roster-sync automatically migrates manifests when encountered:

**v1 (flat) to v3 (nested roster)**:
```json
// Before (v1)
{
  "skeleton_path": "/path/to/skeleton",
  "skeleton_commit": "abc123",
  "managed_files": [...]
}

// After (v3)
{
  "schema_version": 3,
  "roster": {
    "path": "/path/to/roster",
    "commit": "def456",
    "ref": "main",
    "last_sync": "2026-01-03T00:00:00Z"
  },
  "managed_files": [...],
  "migration": {
    "migrated_from": 1,
    "migrated_at": "2026-01-03T00:00:00Z",
    "skeleton_path": "/path/to/skeleton"
  }
}
```

**v2 (nested skeleton) to v3 (nested roster)**:
```json
// Before (v2)
{
  "schema_version": 2,
  "skeleton": {
    "path": "/path/to/skeleton",
    "commit": "abc123"
  }
}

// After (v3)
{
  "schema_version": 3,
  "roster": {
    "path": "/path/to/roster",
    "commit": "def456"
  },
  "migration": {
    "migrated_from": 2,
    "migrated_at": "2026-01-03T00:00:00Z",
    "skeleton_path": "/path/to/skeleton"
  }
}
```

### Backup Files

During migration, backups are automatically created:

| Original Schema | Backup File |
|-----------------|-------------|
| v1 | `.claude/.cem/manifest.json.v1.backup` |
| v2 | `.claude/.cem/manifest.json.v2.backup` |

### Verifying Migration

Check the migration was successful:

```bash
# Check schema version
jq '.schema_version' .claude/.cem/manifest.json
# Should output: 3

# Check migration metadata
jq '.migration' .claude/.cem/manifest.json
# Shows original skeleton path and migration timestamp

# Verify backup exists
ls -la .claude/.cem/manifest.json.*.backup
```

### Rollback Procedure

If migration fails or causes issues:

1. **Restore backup manifest**:
   ```bash
   cp .claude/.cem/manifest.json.v2.backup .claude/.cem/manifest.json
   ```

2. **Use legacy CEM temporarily** (if still available):
   ```bash
   $SKELETON_HOME/cem status
   ```

3. **Report the issue**: File a bug with the original manifest and error output

---

## 5. New Features in roster-sync

### repair Command

The `repair` command rebuilds the manifest from the current file state. Use when:
- Manifest is corrupted or invalid JSON
- Files were manually added/removed outside of sync
- Checksums are out of sync with actual files

```bash
# Preview what would be repaired
roster-sync repair --dry-run

# Perform repair
roster-sync repair

# Repair and remove orphaned files
roster-sync repair --force
```

**What repair does**:
1. Backs up existing manifest to `manifest.repair-backup.{timestamp}`
2. Scans `.claude/` for managed files
3. Restores missing files from roster if available
4. Recalculates all checksums
5. Preserves team information from ACTIVE_RITE
6. Writes fresh v3 manifest

### --prune Flag

Removes orphaned files (files that were once managed but are no longer in roster):

```bash
# Check what would be pruned
roster-sync sync --prune --dry-run

# Sync and prune orphans
roster-sync sync --prune
```

**Backup location**: `.claude/.cem/orphan-backup/`

Orphaned files are moved to the backup directory with timestamps, never permanently deleted.

### --auto-refresh Flag

Automatically refreshes the active rite if roster has updates to team resources:

```bash
roster-sync sync --auto-refresh
```

Only triggers a team refresh if:
- An active rite is set (ACTIVE_RITE exists)
- The team in roster has changes since last sync

### --force Flag

Overrides conflict detection and forces roster version:

```bash
# Force sync, overwriting local changes
roster-sync sync --force

# Force reinitialize existing project
roster-sync init --force
```

**When to use**: After reviewing conflict backups and deciding to accept roster version.

### --dry-run Flag

Preview changes without applying them:

```bash
# Preview sync changes
roster-sync sync --dry-run

# Preview initialization
roster-sync init --dry-run

# Preview repair
roster-sync repair --dry-run
```

### Conflict Detection

roster-sync uses three-way checksum comparison:

| Roster Changed? | Local Changed? | Action |
|-----------------|----------------|--------|
| No | No | SKIP (up to date) |
| No | Yes | SKIP (preserve local) |
| Yes | No | UPDATE (safe to overwrite) |
| Yes | Yes | CONFLICT (backup created) |

**Conflict backups**: `.claude/{filename}.cem-backup`

### Timestamped Backups

All backup operations include timestamps:

| Operation | Backup Location | Format |
|-----------|-----------------|--------|
| Conflict | `.claude/{file}.cem-backup` | No timestamp |
| Orphan prune | `.claude/.cem/orphan-backup/{file}.{timestamp}` | `YYYYMMDD-HHMMSS` |
| Manifest repair | `.claude/.cem/manifest.repair-backup.{timestamp}` | `YYYYMMDD-HHMMSS` |
| Migration | `.claude/.cem/manifest.json.v{N}.backup` | Schema version |

---

## 6. Troubleshooting Guide

### "Manifest not found"

**Error**: `[roster-sync] Error: Manifest not found: .claude/.cem/manifest.json`

**Solution**: Initialize the project:
```bash
roster-sync init
```

If you had a previous CEM setup, check if the manifest was accidentally deleted:
```bash
ls -la .claude/.cem/
```

### "Schema version mismatch"

**Warning**: `Schema version X - migration to v3 recommended`

**Solution**: This is informational. Run sync to trigger auto-migration:
```bash
roster-sync sync
```

The manifest will be automatically migrated and a backup created.

### "Conflicts detected"

**Warning**: `X conflict(s) detected`

**What happened**: Both roster and local versions of a file changed since last sync.

**Solution options**:

1. **Review conflicts manually**:
   ```bash
   # See what files have conflicts
   roster-sync diff

   # Compare specific file
   diff .claude/CLAUDE.md .claude/CLAUDE.md.cem-backup
   ```

2. **Keep local changes**: Remove the `.cem-backup` file without syncing:
   ```bash
   rm .claude/CLAUDE.md.cem-backup
   ```

3. **Accept roster version**: Force sync:
   ```bash
   roster-sync sync --force
   ```

### "Orphaned files detected"

**Warning**: `Orphaned files detected`

**What happened**: Files exist locally that are no longer tracked by roster.

**Solution**:
```bash
# Preview what would be removed
roster-sync sync --prune --dry-run

# Remove orphans (with backup)
roster-sync sync --prune
```

**Find backups**: `.claude/.cem/orphan-backup/`

### "Cannot read manifest file"

**Error**: `Cannot read manifest file` or `Invalid JSON in manifest`

**Solution**: Repair the manifest:
```bash
# Preview repair
roster-sync repair --dry-run

# Perform repair
roster-sync repair
```

### "Roster not found"

**Error**: `Roster not found: /path/to/roster`

**Solution**: Set ROSTER_HOME correctly:
```bash
export ROSTER_HOME="$HOME/Code/roster"  # Adjust path as needed
```

Add to your shell profile for persistence:
```bash
echo 'export ROSTER_HOME="$HOME/Code/roster"' >> ~/.zshrc
```

### "Team may need refresh"

**Warning**: `Team may need refresh`

**What happened**: The active rite has updates in roster since last sync.

**Solution**:
```bash
roster-sync sync --refresh
```

Or let it auto-refresh:
```bash
roster-sync sync --auto-refresh
```

### Where to Find Backup Files

| Backup Type | Location |
|-------------|----------|
| Conflict backups | `.claude/{filename}.cem-backup` |
| Orphan backups | `.claude/.cem/orphan-backup/` |
| Migration backups | `.claude/.cem/manifest.json.v{N}.backup` |
| Repair backups | `.claude/.cem/manifest.repair-backup.{timestamp}` |

### Getting Debug Output

For detailed diagnostic information:
```bash
# Enable debug mode
ROSTER_SYNC_DEBUG=1 roster-sync sync

# Or use the flag
roster-sync sync --debug
```

### Exit Codes Reference

| Code | Meaning | Typical Cause |
|------|---------|---------------|
| 0 | Success | Operation completed |
| 1 | General error / Warnings | Validation passed with warnings |
| 2 | Validation failure | Manifest structure invalid |
| 3 | Init failed | Project path issues |
| 4 | Invalid manifest | Cannot parse or missing required fields |
| 5 | Conflicts | Both local and roster changed |
| 6 | Orphan conflicts | Orphans need attention |

### Getting Help

```bash
# Show help
roster-sync --help

# Show version
roster-sync --version

# Check system requirements
which jq shasum
```

**Requirements**:
- `jq` 1.6+ for JSON processing
- `shasum` (macOS) or `sha256sum` (Linux) for checksums

---

## Appendix: Quick Reference

### Common Operations

```bash
# Initialize new project
roster-sync init

# Sync updates
roster-sync sync

# Sync with team refresh
roster-sync sync --refresh

# Check status
roster-sync status

# Validate manifest
roster-sync validate

# Repair broken manifest
roster-sync repair

# Preview changes
roster-sync sync --dry-run
```

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `ROSTER_HOME` | Path to roster repository | `$HOME/Code/roster` |
| `ROSTER_SYNC_DEBUG` | Enable debug output | `0` |

### Files and Directories

| Path | Purpose |
|------|---------|
| `.claude/.cem/manifest.json` | Sync manifest (v3) |
| `.claude/.cem/checksum-cache.json` | Performance cache |
| `.claude/.cem/orphan-backup/` | Backed up orphaned files |
| `.claude/{file}.cem-backup` | Conflict backups |

---

*Migration guide for roster-sync v1.0.0*
*See TDD-cem-replacement for technical details*

# Skeleton Claude Deprecation Announcement

**Date**: 2026-01-03
**Affects**: Users of skeleton_claude and roster ecosystem
**Status**: Deprecation Notice

---

## Summary

The `skeleton_claude` repository is being deprecated in favor of roster-native functionality. All ecosystem management capabilities previously provided by skeleton_claude and the Claude Ecosystem Manager (CEM) are now available directly within roster through the new `roster-sync` tool.

---

## What Is Changing

### skeleton_claude Repository

The `skeleton_claude` repository will be **archived** and transition to read-only status. After the archive date, no new updates will be made, though the repository will remain accessible for historical reference.

### Environment Variables

- `$SKELETON_HOME` is **no longer required**
- `$ROSTER_HOME` is now the single required environment variable for ecosystem synchronization
- Existing `$SKELETON_HOME` references will continue to work during the support period but should be migrated

### CEM Commands

All CEM commands have been replaced by `roster-sync`:

| CEM Command | New Command | Notes |
|-------------|-------------|-------|
| `cem init` | `roster-sync init` | Path defaults to current directory |
| `cem sync` | `roster-sync sync` | Additional flags available |
| `cem status` | `roster-sync status` | Same output format |
| `cem validate` | `roster-sync validate` | New `--team` flag available |
| `cem diff` | `roster-sync diff` | Same behavior |
| (none) | `roster-sync repair` | **New capability** |

---

## Benefits of Consolidation

### Single Source of Truth

Roster is now the sole repository for ecosystem resources. No more coordinating between two repositories or wondering which has the authoritative version of a file.

### Fewer External Dependencies

- Eliminates the skeleton_claude dependency entirely
- No need to maintain or update a separate skeleton repository
- Simplified PATH and environment configuration

### Simplified Setup

- **Before**: Clone roster + clone skeleton_claude + set both environment variables
- **After**: Clone roster + set `$ROSTER_HOME`

### Improved Synchronization

`roster-sync` introduces enhanced capabilities:

- **Three-way conflict detection**: Safely handles cases where both local and roster versions changed
- **Automatic backups**: All conflict and orphan operations create timestamped backups
- **Manifest repair**: New `repair` command rebuilds manifest from current state
- **Dry-run support**: Preview changes before applying with `--dry-run`
- **Orphan management**: `--prune` flag safely removes orphaned files with backup

### Enhanced Recovery

If something goes wrong:
- Conflict backups at `.claude/{file}.cem-backup`
- Migration backups at `.claude/.cem/manifest.json.v{N}.backup`
- Repair capability for corrupted manifests

---

## Migration Timeline

| Milestone | Date | Description |
|-----------|------|-------------|
| **Deprecation Notice** | 2026-01-03 | This announcement |
| **Support Period** | 2026-01-03 to 2026-02-02 | Questions answered, issues addressed |
| **Archive Date** | 2026-04-03 | skeleton_claude becomes read-only |
| **Post-Archive** | 2026-04-03+ | Read-only historical reference |

### Support Period Details

During the 30-day support period (through 2026-02-02):
- Questions about migration will be answered promptly
- Issues with roster-sync will be prioritized
- Documentation gaps will be addressed
- Edge cases will be handled

---

## Action Required

### Step 1: Verify ROSTER_HOME

Ensure `$ROSTER_HOME` is set correctly:

```bash
echo $ROSTER_HOME
```

If not set, add to your shell profile:

```bash
# ~/.zshrc or ~/.bashrc
export ROSTER_HOME="$HOME/Code/roster"  # Adjust path as needed
```

### Step 2: Update Roster

Pull the latest roster changes:

```bash
cd $ROSTER_HOME
git pull origin main
```

### Step 3: Migrate Manifest

Run sync to migrate your manifest to the new schema:

```bash
roster-sync sync
```

This will:
- Automatically migrate v1/v2 manifests to v3
- Create a backup of your original manifest
- Update all checksums

### Step 4: Verify Migration

Confirm the migration succeeded:

```bash
roster-sync status
roster-sync validate
```

Expected output should show:
- Schema Version: 3
- Roster Path: your roster location
- No validation errors

### Step 5: Remove Skeleton References (Optional)

Once satisfied, you can remove skeleton references:

```bash
# Remove from shell profile if no longer needed
# unset SKELETON_HOME
```

---

## Migration Resources

### Documentation

- **Migration Guide**: `docs/migration/cem-to-roster-migration.md`
- **Command Reference**: `roster-sync --help`
- **Troubleshooting**: See migration guide section 6

### Quick Reference

```bash
# Check current state
roster-sync status

# Sync with latest roster
roster-sync sync

# Validate manifest
roster-sync validate

# Preview changes without applying
roster-sync sync --dry-run

# Repair corrupted manifest
roster-sync repair
```

---

## Support

### Where to Ask Questions

- Open an issue in the roster repository for migration questions
- Tag issues with `skeleton-deprecation` for faster routing

### Known Issues

As of the deprecation date, there are no known blocking issues with the migration. Check the roster repository issues for any updates.

### Common Migration Scenarios

**Q: I have local customizations to managed files. Will they be lost?**

A: No. roster-sync uses three-way conflict detection. If both your local version and roster have changed, a backup is created and you can manually merge.

**Q: Can I continue using skeleton_claude during the support period?**

A: Yes, but you should migrate as soon as practical. The skeleton repository will not receive updates.

**Q: What happens to my manifest after migration?**

A: Your manifest is automatically migrated to schema v3. A backup of the original is created at `.claude/.cem/manifest.json.v{N}.backup`.

---

## What Stays the Same

- **Manifest location**: `.claude/.cem/manifest.json` (unchanged)
- **Managed files location**: `.claude/` directory structure (unchanged)
- **Team management**: swap-rite functionality continues to work
- **Session management**: All session workflows continue unchanged

---

## Thank You

Thank you for using skeleton_claude. The consolidation into roster represents the natural evolution of the ecosystem toward a simpler, more maintainable architecture. We appreciate your patience during the migration period.

---

*Questions? See the migration guide at `docs/migration/cem-to-roster-migration.md` or open an issue in the roster repository.*

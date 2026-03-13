---
last_verified: 2026-02-26
---

# CLI Reference: inscription

> Manage the context file inscription system.

The [inscription](../../reference/GLOSSARY.md#inscription) system synchronizes the context file content with templates and project state, managing ownership of different regions.

**Family**: inscription
**Commands**: 5
**Priority**: MEDIUM

---

## Region Ownership

| Owner | Description | Example |
|-------|-------------|---------|
| `knossos` | Managed by Knossos templates, always synced | `<!-- KNOSSOS:START -->` sections |
| `satellite` | Owned by satellite project, never overwritten | Custom project instructions |
| `regenerate` | Generated from project state | ACTIVE_RITE, agents/ |

---

## Commands

### ari inscription sync

Synchronize the context file with templates.

**Synopsis**:
```bash
ari inscription sync [flags]
```

**Description**:
Synchronizes the context file with Knossos templates. Regenerates managed sections while preserving satellite-owned content.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | false | Preview changes without writing |

**Examples**:
```bash
# Sync inscription
ari inscription sync

# Preview changes
ari inscription sync --dry-run
```

**Related Commands**:
- [`ari sync materialize`](cli-sync.md#ari-sync-materialize) — Full materialization

---

### ari inscription validate

Validate inscription manifest and context file.

**Synopsis**:
```bash
ari inscription validate [flags]
```

**Description**:
Checks that the inscription manifest is valid and context file sections are properly delimited.

**Examples**:
```bash
# Validate inscription
ari inscription validate

# JSON output
ari inscription validate -o json
```

**Related Commands**:
- [`ari rite validate`](cli-rite.md#ari-rite-validate) — Rite validation

---

### ari inscription diff

Show differences between current and generated.

**Synopsis**:
```bash
ari inscription diff [flags]
```

**Description**:
Shows what would change if `ari inscription sync` were run. Useful for understanding drift.

**Examples**:
```bash
# See inscription diff
ari inscription diff
```

**Related Commands**:
- [`ari sync diff`](cli-sync.md#ari-sync-diff) — Full sync diff

---

### ari inscription backups

List available context file backups.

**Synopsis**:
```bash
ari inscription backups [flags]
```

**Description**:
Lists backups created before inscription sync operations. Backups are stored in `.knossos/backups/`.

**Examples**:
```bash
# List backups
ari inscription backups

# JSON for scripting
ari inscription backups -o json
```

**Related Commands**:
- [`ari inscription rollback`](#ari-inscription-rollback) — Restore from backup

---

### ari inscription rollback

Restore the context file from backup.

**Synopsis**:
```bash
ari inscription rollback [flags]
```

**Description**:
Restores the context file from a previous backup. Use after a sync operation causes issues.

**Examples**:
```bash
# Rollback to last backup
ari inscription rollback
```

**Related Commands**:
- [`ari inscription backups`](#ari-inscription-backups) — List available backups

---

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `$XDG_CONFIG_HOME/ariadne/config.yaml` | Config file path |
| `-o, --output` | string | `text` | Output format: text, json, yaml |
| `-p, --project-dir` | string | auto-discovered | Project root directory |
| `-s, --session-id` | string | current session | Override session ID |
| `-v, --verbose` | bool | false | Enable verbose output |

---

## See Also

- [Inscription Glossary Entry](../../reference/GLOSSARY.md#inscription)
- [Knossos Sections](../../reference/GLOSSARY.md#knossos-sections)

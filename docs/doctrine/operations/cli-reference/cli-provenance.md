---
last_verified: 2026-03-26
---

# CLI Reference: provenance

> Inspect the origin and ownership state of files in the channel directory.

The provenance manifest tracks every file Knossos places in the channel directory: who owns it, where it came from, and whether it has been modified. Use `ari provenance show` to diagnose ownership questions before and after sync.

**Family**: provenance
**Commands**: 1
**Priority**: MEDIUM

---

## Synopsis

```bash
ari provenance [command] [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `show` | Display the provenance manifest |

## ari provenance show

```bash
ari provenance show [flags]
```

Displays the provenance manifest showing origin and ownership for all files in the channel directory.

**Status column values**:

| Status | Meaning |
|--------|---------|
| `match` | File on disk matches the expected checksum |
| `diverged` | File has been modified — knossos → user ownership promotion |
| `-` | User or untracked file (no checksum validation) |

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--scope` | string | both | Filter by scope: `rite` or `user` |

## Examples

```bash
# Show full provenance table
ari provenance show

# JSON output for tooling
ari provenance show -o json

# Show only rite-owned files
ari provenance show --scope=rite

# Show only user-owned files
ari provenance show --scope=user

# Verbose (shows full checksums)
ari provenance show --verbose
```

## See Also

- [`ari sync materialize`](cli-sync.md) — Materialize (writes provenance entries)
- [Glossary: Materialization](../../reference/GLOSSARY.md#materialization)

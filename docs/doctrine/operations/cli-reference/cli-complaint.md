---
last_verified: 2026-03-26
---

# CLI Reference: complaint

> View and manage Cassandra complaint artifacts.

Complaints are structured YAML files in `.sos/wip/complaints/` filed by agents when they encounter framework friction. Use `ari complaint` to triage and track them.

**Family**: complaint
**Commands**: 3

---

## Synopsis

```bash
ari complaint [command] [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `list` | List filed complaints (filter by severity, status) |
| `dedup` | Deduplicate complaint corpus by title-prefix grouping |
| `update` | Update a complaint's status |

## Examples

```bash
# List all complaints
ari complaint list

# Filter by severity
ari complaint list --severity=high
ari complaint list --severity=critical --status=filed

# Deduplicate (preview first)
ari complaint dedup --dry-run
ari complaint dedup

# Update status
ari complaint update --id=COMPLAINT-20260311-143022-drift-detect --status=triaged
```

For full option details, run `ari complaint --help` or `ari complaint <subcommand> --help`.

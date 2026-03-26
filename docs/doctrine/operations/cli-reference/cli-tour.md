---
last_verified: 2026-03-26
---

# CLI Reference: tour

> Display the knossos directory tree with file counts and contents.

`ari tour` shows each managed directory with subdirectory listings and file counts from the live filesystem. It is read-only — it does not modify any state.

**Family**: tour
**Commands**: 1

---

## Synopsis

```bash
ari tour [flags]
```

## Purpose

Get a quick view of the knossos directory structure: channel directory, `.knossos/`, `.know/`, `.ledge/`, and `.sos/`.

## Examples

```bash
# Human-readable directory tour
ari tour

# Machine-readable JSON output
ari tour -o json
```

For full option details, run `ari tour --help`.

## See Also

- [`ari status`](cli-status.md) — Health dashboard (higher-level view)
- [`ari knows`](cli-knows.md) — Knowledge domain freshness

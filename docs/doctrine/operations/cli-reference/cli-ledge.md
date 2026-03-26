---
last_verified: 2026-03-26
---

# CLI Reference: ledge

> Promote, list, and manage work product artifacts in the ledge.

The ledge holds work product artifacts at two levels: promotable artifacts in `.ledge/{category}/` and promoted (shelf) artifacts in `.ledge/shelf/{category}/`. Use `ari ledge` to inspect and promote artifacts.

**Family**: ledge
**Commands**: 2
**Priority**: MEDIUM

---

## Synopsis

```bash
ari ledge [command] [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `list` | List ledge artifacts (promotable or shelf) |
| `promote` | Promote an artifact to the shelf |

## Key Flags

### ari ledge list

```bash
ari ledge list [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--shelf` | bool | false | List shelf (promoted) artifacts instead of promotable ones |

### ari ledge promote

```bash
ari ledge promote <path> [flags]
```

Moves an artifact from `.ledge/{category}/` to `.ledge/shelf/{category}/`. Adds promotion frontmatter (`promoted_at`, `promoted_from`) and removes the source file.

Promotable categories: `decisions`, `specs`, `reviews`.

## Examples

```bash
# List promotable artifacts
ari ledge list

# List promoted artifacts on shelf
ari ledge list --shelf

# Promote a review artifact
ari ledge promote .ledge/reviews/GAP-auth-refactor.md

# Promote an ADR
ari ledge promote .ledge/decisions/ADR-0030.md
```

## See Also

- [Architecture Map](../../reference/architecture-map.md) — `.ledge/` directory role
- [`ari session wrap`](cli-session.md#ari-session-wrap) — Session completion (produces artifacts)

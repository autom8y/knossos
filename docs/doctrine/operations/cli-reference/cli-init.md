---
last_verified: 2026-03-26
---

# CLI Reference: init

> Scaffold the channel directory for a new project.

`ari init` creates the channel directory with inscription, settings, and `KNOSSOS_MANIFEST.yaml`. It works without `KNOSSOS_HOME` set — uses embedded rite definitions.

**Family**: init
**Commands**: 1

---

## Synopsis

```bash
ari init [flags]
```

## Purpose

Bootstrap a project for Knossos. Run once per project to create the channel directory scaffold. Optionally activates a rite at the same time.

## Key Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--rite` | string | - | Rite to activate after scaffolding |
| `--force` | bool | false | Overwrite existing channel directory |
| `--source` | string | - | Explicit rite source path |

## Examples

```bash
# Minimal scaffold
ari init

# Scaffold and activate the 10x-dev rite
ari init --rite 10x-dev

# Re-initialize an existing project
ari init --force
```

For full option details, run `ari init --help`.

## See Also

- [`ari sync materialize`](cli-sync.md) — Re-materialize after init
- [`ari org init`](cli-org.md#ari-org-init) — Bootstrap an org (separate from project init)

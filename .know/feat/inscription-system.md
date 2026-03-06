---
domain: feat/inscription-system
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/inscription/**/*.go"
  - "./internal/cmd/inscription/**/*.go"
  - "./knossos/templates/sections/*.md.tpl"
  - "./docs/decisions/ADR-0021*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# CLAUDE.md Inscription System

## Purpose and Design Rationale

Manages CLAUDE.md as a collection of named, owned regions individually regenerated or preserved on sync. Solves the coordination problem: CLAUDE.md serves both the platform (knossos-owned sections) and the project (satellite-owned sections).

**ADR-0021**: Two-axis context model. CLAUDE.md is L0 (always-injected), must be under ~250 lines. **SCAR-002**: Renaming `.claude/` caused CC freeze — led to per-file atomic writes and `WriteIfChanged`.

### Three Owner Types

| Owner | Sync Behavior |
|-------|---------------|
| `knossos` | Always overwritten |
| `satellite` | Never touched |
| `regenerate` | Rebuilt from live state (optional `preserve_on_conflict`) |

## Conceptual Model

### Marker Syntax

```
<!-- KNOSSOS:START {region-name} [key=value ...] -->
...content...
<!-- KNOSSOS:END {region-name} -->
```

### KNOSSOS_MANIFEST.yaml

Tracks region ownership, content hashes, section order, inscription version. `AdoptNewDefaults()` auto-adds new sections from platform. `DeprecatedRegions()` drops removed sections.

### TENSION-001

Two distinct `OwnerType` types: `inscription.OwnerType` (`knossos`/`satellite`/`regenerate`) vs `provenance.OwnerType` (`knossos`/`user`/`untracked`). Do not conflate.

## Implementation Map

15 files in `/Users/tomtenuta/Code/knossos/internal/inscription/`. 6 CLI files. 8 templates in `knossos/templates/sections/*.md.tpl`.

### Pipeline

`SyncCLAUDEmd()` at `/Users/tomtenuta/Code/knossos/internal/inscription/sync.go:63`: Load manifest → `AdoptNewDefaults()` → `GenerateAll()` (template rendering) → `MergeRegions()` (preserve satellite) → `WriteIfChanged()` → update manifest hashes.

**Generator**: 3-level template fallback (embedded FS → filesystem → hardcoded defaults). Uses Go `text/template` + Sprig functions.

**Merger**: LOAD-003 — "User content NEVER destroyed" depends entirely on `MergeRegions()`.

## Boundaries and Failure Modes

- User edits to knossos regions are unconditionally overwritten (conflict logged)
- `section_order` always overwritten by knossos (satellites cannot control ordering)
- `Conditional` struct exists in types but is inert at runtime
- `internal/cmd/inscription/` has zero tests (HIGH test gap)
- Template rendering failure aborts via `prevalidateCLAUDEmd()` before any disk writes

## Knowledge Gaps

1. `Conditional` struct: planned future feature or abandoned design.
2. `prevalidateCLAUDEmd` manifest path discrepancy (`.claude/` vs `.knossos/`).
3. `ANCHOR` directive has no known producer.

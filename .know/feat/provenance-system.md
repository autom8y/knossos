---
domain: feat/provenance-system
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/provenance/**/*.go"
  - "./internal/cmd/provenance/**/*.go"
  - "./docs/decisions/ADR-0026*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# File Provenance Tracking

## Purpose and Design Rationale

Before ADR-0026, three independent mechanisms tracked file ownership in `.claude/` (source resolver types, usersync manifests, inscription manifest). A user agent with the same name as a rite agent was silently overwritten.

ADR-0026 introduced `PROVENANCE_MANIFEST.yaml` tracking every file Knossos places in `.claude/`, with checksums for divergence detection.

**Key decisions**: Separate from `KNOSSOS_MANIFEST.yaml` (file vs region ownership). Provenance as leaf package (zero imports, TENSION-007). Three-state ownership: `knossos`/`user`/`untracked`. Divergence computed at sync time, not stored. Structural equality guard on `Save()` (LOAD-001).

## Conceptual Model

### Ownership Model

| Owner | Sync Behavior |
|-------|---------------|
| `knossos` | Overwrite on sync |
| `user` | Never overwritten |
| `untracked` | Treated as user-owned for safety |

Divergence: knossos-owned file with changed on-disk checksum → promoted to `user`.

### Sync Lifecycle

`LoadOrBootstrap` → `DetectDivergence` → [pipeline stages call `Collector.Record()`] → `Merge` (4-step algorithm) → `Save` (structural equality guard)

### Three Scopes

| Scope | Manifest | Location |
|-------|----------|----------|
| `rite` | `PROVENANCE_MANIFEST.yaml` | `.claude/` |
| `user` | `USER_PROVENANCE_MANIFEST.yaml` | `~/.claude/` |
| `org` | `ORG_PROVENANCE_MANIFEST.yaml` | org `.claude/` |

## Implementation Map

7 files in `/Users/tomtenuta/Code/knossos/internal/provenance/` (true leaf, zero internal imports): `provenance.go`, `manifest.go`, `collector.go`, `divergence.go`, `merge.go` + tests. CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/provenance/provenance.go`.

### Key API

`LoadOrBootstrap()`, `DetectDivergence()`, `NewCollector()`, `Collector.Record()`, `Merge()`, `Save()`, `ManifestPath()`.

### SCAR Evidence

**SCAR-004**: Silent provenance error discard → now aborts on parse errors. **LOAD-001**: Structural equality guard prevents CC file watcher trigger on no-op syncs.

## Boundaries and Failure Modes

- Does NOT track files outside `.claude/`
- Does NOT handle CLAUDE.md region ownership (that's inscription)
- `SourceType` string values must stay in sync with `materialize/source` by convention (TENSION-007)
- Schema version `2.0` with `migrateV1ToV2()` shim — version bump requires migration
- Bootstrap safety gap: first sync without prior manifest may overwrite pre-existing files

## Knowledge Gaps

1. User-scope provenance pipeline wiring not fully traced.
2. Org-scope manifest may be partially implemented or aspirational.
3. `formatSource()` in CLI ignores `sourceType` parameter (marked `_`).

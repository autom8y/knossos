---
domain: feat/know-system
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/know/**/*.go"
  - "./internal/cmd/knows/**/*.go"
  - "./internal/cmd/hook/context.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Codebase Knowledge Domain System (.know/)

## Purpose and Design Rationale

Stores pre-computed codebase knowledge as markdown files with YAML frontmatter in `.know/`. Files are persistent, versioned artifacts written by theoros observers and consumed by agents at session start or on demand. The design bet: periodic full-pass observation is cheaper than cumulative per-session discovery.

**Key decisions**: Frontmatter as contract (`Meta` struct). Dual-axis staleness (time + code via `source_scope` globs). Incremental cycle tracking with forced full regeneration after `max_incremental_cycles`. AST-based semantic diffing for Go via `go/ast`. Subdirectory namespacing (`feat/`, `release/`).

## Conceptual Model

### Key Types

- **`Meta`** — YAML frontmatter: `domain`, `generated_at`, `expires_after`, `source_scope`, `source_hash`, `confidence`
- **`DomainStatus`** — runtime freshness: `Fresh`, `TimeExpired`, `CodeChanged`, `ForceFull`
- **`ChangeManifest`** — what changed: new/modified/deleted/renamed files, delta lines/ratio
- **`SemanticDiff`** — AST-level declaration changes (NEW, DELETED, MODIFIED, SIGNATURE_CHANGED)

### Staleness Model

A domain is stale when EITHER `expires_after` elapsed OR source code changed within `source_scope` globs since `source_hash`. Scoped staleness avoids false positives on doc-only commits.

### RecommendedMode Decision Tree

No file changes → `time-only`. Cycle limit hit → `full`. DeltaRatio >= 0.5 or DeltaLines >= 5000 → `full`. Otherwise → `incremental`.

## Implementation Map

| File | Purpose |
|------|---------|
| `internal/know/know.go` | Core: `Meta`, `DomainStatus`, `ReadMeta()`, `buildDomainStatus()`, `scopedStaleness()` |
| `internal/know/manifest.go` | `ChangeManifest`, `ComputeChangeManifest()`, `FilterChangeManifest()`, `RecommendedMode()` |
| `internal/know/astdiff.go` | AST diffing: `ComputeFileDiff()`, `FormatSemanticDiff()` |
| `internal/know/validate.go` | Reference validation: `ValidateDomain()`, `ValidateAll()` |
| `internal/cmd/knows/knows.go` | CLI: `ari knows [--delta] [--validate] [--semantic-diff]` |

Tests: 3,575 lines across 4 test files (dense coverage for ~1,400 source lines).

### Hook Integration

`internal/cmd/hook/context.go:350-383` — `knowStatus()` injects `.know/` freshness into every SessionStart.

## Boundaries and Failure Modes

- Does NOT generate knowledge content (read-only; `/know` dromenon dispatches theoros)
- Does NOT enforce freshness (stale domains reported, not blocked)
- AST diffing is Go-only (non-Go files appear in `NonGoFiles`)
- `matchScope()` only handles `**` patterns, not `?` wildcards
- Git unavailability: graceful degradation (treated as fresh)
- Malformed frontmatter: silently skipped in results
- `DeltaLines` not scope-filtered (minor over-counting accepted)

## Knowledge Gaps

1. No ADR for the know-system.
2. `matchScope()` boundary behavior undocumented.
3. Git lock timeout gap in hook path.
4. New Go files not semantically diffed (asymmetry in `runSemanticDiff`).

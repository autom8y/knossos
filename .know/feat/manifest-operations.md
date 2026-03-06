---
domain: feat/manifest-operations
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/manifest/**/*.go"
  - "./internal/cmd/manifest/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# YAML/JSON Manifest Operations

## Purpose and Design Rationale

Generic, format-agnostic layer for loading, validating, diffing, and three-way merging YAML/JSON manifest files. Primary targets: `.claude/manifest.json` and `rites/*/manifest.yaml`.

**Key decisions**: Unified generic `Manifest` type (`map[string]interface{}`). TD-3: warnings-only on rite manifest validation, never blocks `Load()`. Git ref loading (`commit:path` patterns). Lightweight structural validation (full JSON Schema not active at runtime — DEBT-178). Four merge strategies: smart, ours, theirs, union.

## Conceptual Model

### Four Operations

- **Show**: `Load` + display with optional schema info
- **Validate**: `Load` + `SchemaValidator.Validate` (structural checks only)
- **Diff**: `Load` x2 + `Diff` (structural comparison, optional array-as-set mode)
- **Merge**: `Load` x3 + `Merge` (three-way with strategy selection and conflict detection)

### Three-Way Merge (Smart Strategy)

Standard semantics: unchanged + changed = accept change; both changed same = accept; both changed differently = conflict. Conflicts produce git-style markers. Default resolution: ours wins.

## Implementation Map

Domain: `/Users/tomtenuta/Code/knossos/internal/manifest/` — `manifest.go` (404 lines), `diff.go` (334), `merge.go` (408), `schema.go` (443). CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/manifest/` — show, validate, diff, merge subcommands.

### Key Entry Points

- `manifest.Load(path)` — primary loader (auto-detects git refs)
- `manifest.Diff(base, compare, opts)` — structural comparison
- `manifest.Merge(base, ours, theirs, opts)` — three-way merge
- `manifest.NewSchemaValidator()` — structural field-presence checks

## Boundaries and Failure Modes

- Full JSON Schema validation NOT active at runtime (DEBT-178)
- Format detection defaults to JSON for unknown extensions
- Conflict marker generation is path-naive (simplified replacement)
- `ValidateRiteManifest` only fires when both `name` AND `agents` keys present
- `LoadFromGitRef` has no `exec.CommandContext` timeout (SCAR-010 class risk)

## Knowledge Gaps

1. "TDD Section 7.1" merge spec document not located on disk.
2. `ValidateBytes` is dead code (compiler is nil).
3. `agent-manifest.schema.json` has no known runtime consumer.

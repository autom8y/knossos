---
domain: feat/manifest-operations
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/manifest/**/*.go"
  - "./internal/cmd/manifest/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# YAML/JSON Manifest Operations

## Purpose and Design Rationale

General-purpose toolkit for inspecting, validating, comparing, and reconciling YAML/JSON manifest files. Rite manifest validation is advisory, not blocking (TD-3). Git ref support (HEAD:.claude/manifest.json). Three-way merge with 4 strategies (smart, ours, theirs, union). Distinct from internal/materialize's RiteManifest (two packages are not kept in sync).

## Conceptual Model

**Manifest:** uniform envelope (Path, Format, Content map, Raw bytes). Format auto-detected from extension. **Git ref support:** isGitRef detection (excludes Windows drive letters), git show with 30s timeout. **Four merge strategies:** smart (field-level three-way with conflict markers), ours, theirs, union (arrays as sets). **Diff model:** recursive tree walk with jQuery-style paths, added/modified/removed classification, --ignore-order for set semantics. **Schema validation:** lightweight structural checking (jsonschema compiler present but nil).

## Implementation Map

`internal/manifest/` (4 files): manifest.go (Load/Save/Clone/Get, ValidateRiteManifest advisory), diff.go (Diff, FormatUnified), merge.go (Merge 4 strategies, conflict markers), schema.go (SchemaValidator, DetectSchemaFromPath). `internal/cmd/manifest/` (5 files): show (with --resolved defaults injection), validate (--strict for additionalProperties), diff (non-zero exit on changes), merge (--write-to, --dry-run). 3 embedded schema files.

## Boundaries and Failure Modes

internal/manifest is NOT used by internal/materialize (separate RiteManifest structs, not in sync). Full JSON Schema validation disabled (compiler nil). ValidateRiteManifest is advisory (slog.Warn only). Format detection bug in merge --write-to (compares path string to "yaml" literal). ValidateBytes would panic (nil compiler). No CLI-layer tests.

## Knowledge Gaps

1. rite-manifest.schema.json requires agents[].file but Go validator doesn't check
2. Schema and operative struct architecturally diverged
3. agent-manifest.schema.json relationship unclear

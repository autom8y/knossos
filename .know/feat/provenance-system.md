---
domain: feat/provenance-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/provenance/**/*.go"
  - "./internal/cmd/provenance/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# File Provenance Tracking

## Purpose and Design Rationale

Tracks file ownership in shared channel directories (.claude/, .gemini/). Three axes: Owner (knossos/user/untracked), Scope (rite/user/org), Source (source_path + source_type from resolution tier). Two separate manifests prevent cross-contamination (PROVENANCE_MANIFEST.yaml for rite, USER_PROVENANCE_MANIFEST.yaml for user). Leaf package constraint (one-way: materialize imports provenance). Structural equality comparison before writing prevents CC file watcher triggers (LB-002). Schema v2.0 with v1 migration path.

## Conceptual Model

**Divergence lifecycle:** knossos -> user edit detected -> DetectDivergence -> promoted to user. Once user, pipeline cannot reclaim without --overwrite-diverged. **4-step Merge algorithm:** Step 0 (carry forward knossos entries still on disk), Step 1 (layer divergence promotions), Step 2 (layer collector entries, skip user-promoted unless overwrite), Step 3 (promote remaining untracked to user). **Multi-channel:** PROVENANCE_MANIFEST.yaml (claude) vs PROVENANCE_MANIFEST_GEMINI.yaml.

## Implementation Map

`internal/provenance/` (5 files): provenance.go (types, factories, constants), manifest.go (Load/Save/LoadOrBootstrap, validation, migration), collector.go (Collector interface + defaultCollector + NullCollector), merge.go (4-step Merge), divergence.go (DetectDivergence, DivergenceReport). `internal/cmd/provenance/provenance.go` (CLI: ari provenance show). Checksum format: sha256:+64hex. Integration: every materialize stage receives collector parameter.

## Boundaries and Failure Modes

LoadOrBootstrap is NOT fail-open (corrupt manifest aborts sync, only bootstraps on file-not-found -- LB-009). Volatile files NOT tracked (KNOSSOS_MANIFEST.yaml, sync/state.json, ACTIVE_RITE). Deleted files promoted to user with empty checksum then filtered in Merge. TENSION-001: provenance.OwnerType vs inscription.OwnerType are distinct types. TENSION-006: SourceType strings synced manually with materialize/source.

## Knowledge Gaps

1. ADR-0026 not found on disk
2. --overwrite-diverged CLI flag wiring not confirmed
3. Census listed scanner.go/report.go which don't exist

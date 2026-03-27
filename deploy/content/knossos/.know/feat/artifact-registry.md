---
domain: feat/artifact-registry
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/artifact/**/*.go"
  - "./internal/ledge/**/*.go"
  - "./internal/cmd/artifact/**/*.go"
  - "./internal/cmd/ledge/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.92
format_version: "1.0"
---

# Work Artifact Registry (.ledge/)

## Purpose and Design Rationale

Two-tier document provenance system tracking non-code work products. Session registries (artifacts.yaml per session, written eagerly) separate from project registry (registry.yaml, derived aggregate). Graduated paths vs original paths. Sails-gated auto-promotion (white sails only). Code artifacts never move. Trigger: ari session wrap (non-fatal graduation + auto-promotion).

## Conceptual Model

**Four-state lifecycle:** REGISTERED (session-local) -> Graduated (.ledge/{category}/) -> Shelf (.ledge/shelf/{category}/). **8 artifact types:** PRD/TDD/TestPlan/Runbook -> specs, ADR -> decisions, Review -> reviews, Spike -> spikes (not promotable), Code -> stays in tree. **Phase inference:** PRD->requirements, TDD/ADR->design, Code->implementation, TestPlan->validation. **Provenance layering:** graduated_at/original_path on graduation; promoted_at/promoted_from on promotion. **Project indexes:** ByPhase, ByType, BySpecialist, BySession.

## Implementation Map

`internal/artifact/` (5 files): registry.go (SessionRegistry/ProjectRegistry CRUD), aggregate.go (session-to-project aggregation), query.go (multi-dimensional AND filtering), graduate.go (file-copy with provenance frontmatter). `internal/ledge/` (2 files): promote.go (single-artifact move), auto_promote.go (batch auto-promotion). CLI: artifact (register/query/list/rebuild), ledge (list/promote).

## Boundaries and Failure Modes

Non-fatal graduation during wrap (warnings only). REVIEW/SPIKE filename detection gap (not in detectArtifact prefix map). Destination-already-exists blocks promote (no overwrite). Non-atomic promotion (write then remove). Project registry hardcoded to .claude/. AggregateAll does not restore physical files. Filename collision across sessions (same base name overwrites).

## Knowledge Gaps

1. ari ledge query subcommand not found
2. Hook integration for auto-registration not found
3. --auto-promote default behavior not confirmed

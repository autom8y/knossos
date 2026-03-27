---
domain: feat/complaint-filing-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/complaint/**/*.go"
  - "./internal/cmd/hook/driftdetect.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.91
format_version: "1.0"
---

# Cassandra Protocol Complaint Filing

## Purpose and Design Rationale

Structured self-reporting mechanism for framework friction. File-per-complaint YAML model (schema-validated, append-only). Two filing tiers: quick-file (7 fields) and deep-file (adds evidence, zone, related scars). Two independent filing paths: agent skill (complaint-filing) and drift-detect hook (PostToolUse). ADR-cassandra-dedup-boundary formalizes the separation: hooks own infrastructure-level rate limiting, skills own domain-level dedup.

## Conceptual Model

**Lifecycle:** filed -> triaged -> accepted|rejected -> resolved. **Two dedup layers:** Hook-level (key-based state file, `.drift-dedup-state.json`) and skill-level (directory scan, title matching). **Three drift patterns:** tool-fallback (single-event), retry-spiral (session-scoped, 3+ failures), command-exploration (session-scoped, 3+ variants). **Three zones:** parameter (auto-tunable), behavior (human-gated), structure (never auto-modify).

## Implementation Map

Hook path: `internal/cmd/hook/driftdetect.go` (runDriftdetectCore, detectToolFallback, fileDriftComplaint with YAML struct marshaling). CLI: `internal/cmd/complaint/` (list, update, dedup subcommands). Schema: `internal/validation/schemas/complaint.schema.json`. Skill: `complaint-filing/INDEX.lego.md`. ADR: `.ledge/decisions/ADR-cassandra-dedup-boundary.md`.

## Boundaries and Failure Modes

Dedup state file missing: fail-open (fresh state, complaint filed). Concurrent hook races: last-writer-wins. Filing errors: VerboseLog warn + return (non-blocking). Schema validation post-file: non-blocking. retry-spiral/command-exploration have no hook-level dedup gate (tool-fallback only). Dedup state accumulates indefinitely.

## Knowledge Gaps

1. /reflect triage pipeline (WS-4) not found
2. `--reset` flag not implemented
3. retry-spiral/command-exploration dedup gap intent unknown

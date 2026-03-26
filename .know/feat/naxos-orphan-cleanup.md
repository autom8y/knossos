---
domain: feat/naxos-orphan-cleanup
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/naxos/**/*.go"
  - "./internal/cmd/naxos/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Naxos Orphan Session Cleanup

## Purpose and Design Rationale

Addresses session accumulation in long-running projects. Report-only by design (refuses to auto-archive/delete -- surfaces problems and recommends). Two-stage pipeline: scan (3 predicates) + triage (4 severity levels). NAXOS_TRIAGE.md artifact for hook consumption (ReadTriageSummary optimized for <5ms). Named for the island where Theseus abandoned Ariadne.

## Conceptual Model

**Three orphan predicates:** INACTIVE (ACTIVE, inactivity > 24h), INCOMPLETE_WRAP (ACTIVE, current_phase=wrap), STALE_SAILS (PARKED, gray/absent sails, parked > 7d). First matching predicate wins. **Four severity levels:** CRITICAL (INCOMPLETE_WRAP any age; INACTIVE > 30d), HIGH (INACTIVE 7-30d; STALE_SAILS > 14d), MEDIUM (INACTIVE 24h-7d; STALE_SAILS 7-14d), LOW (everything else). **Actionable:** CRITICAL and HIGH only. **Suggested actions:** WRAP, RESUME, DELETE (DELETE if age > 30d).

## Implementation Map

`internal/naxos/` (5 files): types.go (OrphanReason, SuggestedAction), scanner.go (Scanner.Scan, checkSession, 3 predicates), triage.go (Triage, computeSeverity/Priority), artifact.go (WriteTriageArtifact, ReadTriageSummary), report.go/triage_report.go (output formatting). CLI: `internal/cmd/naxos/` (scan.go, triage.go). Dromenon: `/naxos` in `mena/session/naxos/INDEX.dro.md`.

## Boundaries and Failure Modes

Report-only (no mutations). Does not read events.jsonl (only SESSION_CONTEXT.md + WHITE_SAILS.yaml). Corrupt sessions silently skipped. Empty sails treated as GRAY (fail-open). Inactive threshold is strictly > (not >=). No PARKED session age check without parked_at field. --no-artifact flag has naming inversion (variable writeArtifact means skipArtifact when true).

## Knowledge Gaps

1. session.LoadContext parse failure behavior on missing status field unknown
2. Hook wiring for naxos_summary injection not traced
3. paths.IsSessionDir pattern not read

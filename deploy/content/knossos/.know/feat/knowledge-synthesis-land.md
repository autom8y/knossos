---
domain: feat/knowledge-synthesis-land
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/land/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# Knowledge Synthesis Pipeline (ari land)

## Purpose and Design Rationale

Converts ephemeral session knowledge (locked in `.sos/archive/`) into durable cross-session knowledge. Three-stage pipeline: `ari land synthesize` (CLI inventory tool, JSON manifest), Dionysus agent (synthesizer, writes `.sos/land/`), `/know` pipeline (refreshes `.know/`). Three synthesis domains: initiative-history, scar-tissue, workflow-patterns. Full-rewrite strategy per run. `.sos/land/` is tracked (not gitignored).

## Conceptual Model

**Three stages:** CLI inventory -> Dionysus agent -> /know refresh. **LAND_MAP:** initiative-history -> architecture/design-constraints, workflow-patterns -> conventions/test-coverage, scar-tissue -> scar-tissue. **Data quality:** RICH (>=40 lines), MODERATE (20-39), SPARSE (<20). **Summonable agent lifecycle:** Dionysus summoned before /land, dismissed after.

## Implementation Map

`internal/cmd/land/land.go` (NewLandCmd), `synthesize.go` (sessionSummary, landFileSummary, synthesizeOutput). Dionysus agent: tier summonable, model opus, maxTurns 75, tools Read/Write/Glob/Grep only. `/land` dromenon orchestrates all three stages.

## Boundaries and Failure Modes

Dionysus reads ONLY .sos/archive/, writes ONLY .sos/land/. CLI is read-only. Missing archive dir: graceful exit with guidance. No CLI tests exist. No incremental synthesis (full rewrite every run). Dionysus dismissal deferred to CC restart. Context window pressure at 55+ sessions.

## Knowledge Gaps

1. session.Context full field inventory not read
2. workflow-patterns.md content not verified
3. /dion dromenon may be deprecated
4. Incremental synthesis path not implemented

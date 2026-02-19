---
name: materialize-review
description: "ARCH-REVIEW-1 findings for the materialization pipeline. Use when: modifying materialize, provenance, inscription, or frontmatter packages. Triggers: materialize, provenance, inscription, pipeline architecture, mena projection, CLAUDE.md generation, rite resolution, sync pipeline, god package decomposition, sub-package extraction."
---

# materialize-review

> Architecture review findings and remediation roadmap for the materialization pipeline (ARCH-REVIEW-1, 2026-02-18).

## Health Baseline

| Package | LOC | Alignment | Key Issue |
|---------|----:|----------:|-----------|
| checksum | 91 | 10/10 | None |
| frontmatter | 100 | 9/10 | Duplicate parser in inscription |
| sync | 172 | 7/10 | Vestigial fields (CR-3 RESOLVED) |
| provenance | 535 | 8/10 | Constructor sprawl (R1), merge in wrong package (R4) |
| inscription | 3,918 | 9/10 | Uses own frontmatter parser (R2) |
| materialize | 5,275 | 4/10 | God package, 6 concerns (R5) |

Overall health: 7.5/10. Zero circular dependencies. 27-edge DAG, strictly acyclic.

## Remediation Status

| ID | Description | Status | Commit |
|----|-------------|--------|--------|
| R1 | ProvenanceEntry constructor helpers | RESOLVED | 7104001 |
| R2 | Inscription frontmatter migration | RESOLVED | fdd6a9d |
| R3 | Surface provenance load errors | RESOLVED | e5655a9, e7277cf |
| R4 | Extract merge algorithm to provenance | RESOLVED | 6ff1188 |
| R5 | Materialize god package decomposition | PENDING | — |
| R7 | Mena type caching | RESOLVED | 39cf68f |
| R8 | Phantom agent validation | RESOLVED | d628f0f |
| CR-3 | Vestigial sync.State fields | RESOLVED | 0c69b20 |
| CR-5 | writeIfChanged docs in user_scope | RESOLVED | 16c279c |

## Key Invariants

- One-way dependency: materialize imports provenance, never reverse
- writeIfChanged() on all rite-scope writes (CC file watcher safety)
- User content NEVER destroyed (satellite regions, user-agents)
- Volatile files excluded from provenance tracking
- structurallyEqual() prevents timestamp-only rewrites

## Decision Log

- U-1 (RiteManifest duplication): DEFERRED to Initiative C design spike
- U-2 (soft mode staleness): ACCEPT AS-IS
- U-3 (validator Fix() side effects): ACCEPT AS-IS

## Full Artifacts

Load on-demand from `.claude/wip/q1_arch/mat/`:
- `ARCH-REVIEW-1-HEALTH.md` — risk register, SPOFs, doctrine compliance
- `ARCH-REVIEW-1-ROADMAP.md` — phased remediation with code snippets
- `ARCH-REVIEW-1-ARCHITECTURE.md` — dependency graph, coupling scores
- `ARCH-REVIEW-1-PIPELINE.md` — 10-stage pipeline reference
- `DECISIONS-unknowns.md` — stakeholder decisions from interview

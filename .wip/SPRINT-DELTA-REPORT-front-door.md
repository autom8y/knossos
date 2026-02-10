# Sprint Delta Report: The Front Door

**Sprint**: The Front Door
**Session**: session-20260210-150019-f295bd78
**Branch**: doctrine-doctor
**Date**: 2026-02-10

---

## Executive Summary

The Front Door sprint operationalized Pythia (orchestrator rename), standardized Exousia (authority contracts), built the /go cold-start dispatcher, redesigned the Inscription (CLAUDE.md), and updated all doctrine. Zero regressions introduced.

## Metrics: Before vs After

| Metric | Baseline (W0-1) | Final | Delta |
|--------|-----------------|-------|-------|
| Rite agents total | 59 | 59 | 0 |
| Agents with `## Exousia` | 0 | 59 (100%) | +59 |
| Agents with `### You Do NOT Decide` | 0 | 59 (100%) | +59 |
| Rites with `pythia.md` | 0 | 11 (100%) | +11 |
| Rites with `orchestrator.md` | 11 | 0 | -11 |
| Cross-cutting agent (moirai) with Exousia | 0 | 1 | +1 |
| Inscription mentions Pythia | No | Yes | Fixed |
| Inscription mentions Exousia | No | Yes | Fixed |
| Inscription mentions /go | No | Yes | Fixed |
| Inscription mentions Fates | No | Yes | Fixed |
| Go archetype section: Exousia | "Domain Authority" | "Exousia" | Fixed |
| Scaffold templates: Exousia | "Domain Authority" | "Exousia" | Fixed |
| Doctrine: Exousia documented | No | Yes | Added |
| Doctrine: Pythia paths correct | `user-agents/orchestrator.md` | `rites/*/agents/pythia.md` | Fixed |
| ari lint new issues | 20 pre-existing | 20 pre-existing | 0 new |
| Test suite regressions | 1 pre-existing (theoros type) | 1 pre-existing | 0 new |

## Changes by Wave

### W0-1: Theoria Baseline (23 agents audited)
- Established pre-sprint state across ecosystem, shared, and 10x-dev rites
- Documented naming gaps: all agents had "Domain Authority", no Pythia

### W1-1: Fates-as-Primitives
- Validated existing Clotho/Lachesis/Atropos skills at `mena/session/moirai/`

### W1-2: /go Dromenon
- Created `mena/navigation/go.dro.md` — cold-start dispatcher
- Validated autopark hook integration

### W2-1: Combined Pass — Shared Agents (EXEMPLAR)
- ecosystem rite: orchestrator.md → pythia.md, all specialists got Exousia
- shared rite: theoros kept (auditor type, out of scope)
- moirai: added Exousia section

### W2-2: Combined Pass — Rite Batch 1 (4 rites)
- 10x-dev, forge, docs, hygiene: all processed
- pythia.md created, orchestrator.md deleted, manifests updated

### W2-3: Combined Pass — Rite Batch 2 + Cross-Refs (6 rites)
- debt-triage, intelligence, rnd, security, sre, strategy: all processed
- Cross-references updated in mena files, TODO.md, audit/validation docs

### W3-1: Inscription Design
- Spec: `.wip/INSCRIPTION-REDESIGN-SPEC.md`
- Token impact: +135 tokens (within budget)

### W3-2: Inscription Implementation
- 4 inscription templates updated
- Go code updated (generator.go, archetype.go)
- 3 scaffold templates updated (orchestrator.md.tpl, specialist.md.tpl, reviewer.md.tpl)
- Deep QA: Found and fixed stale "Domain Authority" in archetype definitions + scaffold templates

### W4-1: Doctrine + Glossary Updates
- GLOSSARY.md: Pythia paths, Moirai paths, Heroes paths, Exousia entry added
- mythology-concordance.md: Pythia paths, Exousia section added, materialization flow updated
- knossos-doctrine.md: Pythia reference, Exousia paragraph, service map, concordance, drift registry

### W5-1: This Report
- Validated all metrics at 100%

## Files Changed Summary

| Category | Count |
|----------|-------|
| Agent files (modified/new/deleted) | 71 |
| New pythia.md files | 11 |
| Deleted orchestrator.md files | 11 |
| Go source files | 4 |
| Template files (.tpl) | 7 |
| Doctrine files | 3 |
| Manifest files | 11 |
| Mena/Skills files | 14 |
| WIP artifacts | 8 |
| **Total uncommitted files** | **143** |

## Gate Results

| Gate | Status | Key Verification |
|------|--------|-----------------|
| GATE 1 (W1 complete) | PASSED | /go built, Fates validated |
| GATE 2 (W2 complete) | PASSED | 59/59 agents Exousia, 11/11 pythia, 0 orchestrator.md |
| GATE 3 (W3 complete) | PASSED | Inscription clean, no stale references |

## Remaining Items (Not Sprint-Blocking)

1. **theoros.md type "auditor"** — pre-existing lint/test failure; type not in valid list. Separate fix needed.
2. **`/roster/` path references** in design-principles.md, cli-sync.md, INDEX.md — pre-knossos-rename era docs, separate migration.
3. **Forge manifest dromena array** — `theoria.dro.md` not listed in manifest (auto-discovered by extension). Cosmetic.

## Verdict

**SPRINT COMPLETE**. All 5 waves executed, all 3 gates passed, zero regressions. The front door is open.

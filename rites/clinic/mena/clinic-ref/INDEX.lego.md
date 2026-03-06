---
name: clinic-ref
description: "Clinic rite methodology reference. Use when: writing evidence files, reading index.yaml, formatting handoff artifacts, understanding investigation phases, applying back-route logic, resuming parked investigations. Triggers: clinic investigation, evidence architecture, index.yaml schema, cross-rite handoff, evidence gap, back-route, investigation slug, intake report, evidence collection, diagnosis."
---

# Clinic Methodology Reference

## Rite Purpose

Investigation and root cause analysis lifecycle for production errors. Four phases, always sequential, always all four. Depth variance is emergent from what agents find — not from pre-classifying the investigation.

```
intake -> examination -> diagnosis -> treatment
```

## Evidence Architecture

All investigation artifacts live under `.sos/wip/ERRORS/{investigation-slug}/`.

```
.sos/wip/ERRORS/{slug}/
  intake-report.md        owner: triage-nurse
  index.yaml              owner: current-phase    (shared coordination)
  E001.txt                owner: pathologist
  E002.txt                owner: pathologist
  diagnosis.md            owner: diagnostician
  treatment-plan.md       owner: attending
  handoff-10x-dev.md      owner: attending        (if applicable)
  handoff-sre.md          owner: attending        (if applicable)
  handoff-debt-triage.md  owner: attending        (if applicable)
```

## Session Resume

On `/sos resume`, Pythia reads `index.yaml` status field:

| Status Field | Dispatch To |
|--------------|-------------|
| `intake` | triage-nurse |
| `examination` | pathologist |
| `diagnosis` | diagnostician |
| `treatment` | attending |
| `complete` | Report to user: investigation finished |
| `{phase}:{back_route}_round_{N}` | Agent for that phase, with back-route context |

Multiple investigation directories → Pythia presents the list. Does not auto-select.

## Investigation Slug Naming

Descriptive, kebab-case: `checkout-500-intermittent`, `etl-silent-failures`, `auth-latency-spike`

Not: `investigation-1`, `bug-2024-01-15`, `error-fix`

## Complexity

One level: INVESTIGATION. All investigations run all four phases. Depth is emergent:
- Simple bug: ~4 agent invocations, ~30k tokens, no back-routes
- Compound failure: ~6-8 invocations, ~150k tokens, 1-2 back-routes

## Companion Reference

| Topic | File | When to Load |
|-------|------|-------------|
| index.yaml full schema, status values, evidence format | `index-schema.lego.md` | Writing or validating index.yaml |
| Back-route protocols (evidence_gap, diagnosis_insufficient, scope_expansion) | `back-routes.lego.md` | Triggering or handling back-routes |
| Cross-rite handoff formats (10x-dev, SRE, debt-triage) | `handoff-formats.lego.md` | Producing treatment handoff artifacts |
| Agent anti-patterns | `anti-patterns.lego.md` | Reviewing agent behavior |

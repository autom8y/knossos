---
last_verified: 2026-03-26
---

# Rite: clinic

> Investigation and debugging lifecycle for production errors.

The clinic rite provides a structured investigation workflow for production errors — from intake through evidence collection, diagnosis, and treatment planning. Every investigation follows the same four-phase path regardless of complexity; depth emerges from what agents find, not from pre-classification.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | clinic |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | potnia |

---

## When to Use

- Investigating production errors, 500s, or silent failures
- Triaging intermittent bugs that are hard to reproduce
- Diagnosing compound failures where multiple causes interact
- Producing fix specifications and cross-rite handoff artifacts for 10x-dev, SRE, or debt-triage

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates investigation phases, gates transitions, manages back-routes |
| **triage-nurse** | Turns vague error reports into structured investigations; creates evidence collection plans |
| **pathologist** | Collects evidence across flagged systems; writes findings to disk, not context |
| **diagnostician** | Forms and tests hypotheses against evidence; detects compound failures |
| **attending** | Translates diagnosis into fix specifications and cross-rite handoff artifacts |

See agent files: `rites/clinic/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[intake] --> B[examination]
    B --> C[diagnosis]
    C --> D[treatment]
    D --> E[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| intake | triage-nurse | intake-report.md, index.yaml | Always |
| examination | pathologist | evidence files (E001.txt...), updated index.yaml | Always |
| diagnosis | diagnostician | diagnosis.md, updated index.yaml | Always |
| treatment | attending | treatment-plan.md, handoff artifacts | Always |

**Deliberate design**: The clinic uses a single complexity level (INVESTIGATION). All investigations enter at intake. There are no skip paths — depth is emergent from findings.

### Back-Routes

- **evidence_gap**: Diagnostician requests targeted additional evidence → back to pathologist
- **diagnosis_insufficient**: Attending finds diagnosis unactionable → back to diagnostician

---

## Invocation Patterns

```bash
# Quick switch to clinic
/clinic

# Start an investigation
Task(triage-nurse, "checkout service returning 500s for logged-in users since Tuesday deploy")

# Request targeted evidence (evidence_gap back-route)
Task(pathologist, "collect circuit breaker state for checkout-service over last 6 hours — needed for H002")

# Analyze evidence
Task(diagnostician, "evidence collection complete, index.yaml has 8 entries across checkout-service and DuckDB")
```

---

## Skills

- `clinic-ref` — Investigation workflow reference, evidence architecture, index.yaml schema

---

## Source

**Manifest**: `rites/clinic/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)
- [CLI: sync](../operations/cli-reference/cli-sync.md)
- [Rite Catalog](index.md)

---
name: quality-gates
description: "Per-posture quality gate definitions: corrective gates (QG-T1..T5), generative gates (QG-C1..C7), transformative gates (QG-E1..E9). Includes hard/soft/advisory classification, dimension ownership, gate agent assignment, and embedded checkpoint criteria. Use when: coordinating multi-agent validation, determining which quality dimensions to evaluate, checking posture-specific gate criteria. Triggers: quality gates, validation, QG-T, QG-C, QG-E, soft gate, hard gate, advisory, D0, D1, D2, D3, D4, D5, D6, micro-interactions, cognitive efficiency, consistency, personality, edge state, emotional design."
---

# Quality Gates Skill

Per-posture quality gate definitions for potnia (routing coordination) and frontend-fanatic (soft gate evaluation).

## Contents

| File | When to Load |
|------|-------------|
| `corrective-gates.md` | Corrective posture validation; audit-phase quality criteria; QG-T1 through QG-T5 |
| `generative-gates.md` | Generative posture validation; D1/D2 soft gate criteria; QG-C1 through QG-C7 |
| `transformative-gates.md` | Transformative posture validation; five-contract verification; QG-E1 through QG-E9 |

## Consuming Agents

- **frontend-fanatic** (primary): loads for per-posture evaluation criteria when executing soft gate assessment in validate phase
- **potnia**: quality-gates information already encoded in orchestrator.yaml handoff criteria and back_routes; potnia does not need to load this skill directly

## Gate Type Reference

| Type | Description | Effect |
|------|-------------|--------|
| **Hard** | Always blocking, zero tolerance | Workflow stops until resolved |
| **Soft** | Blocking in specific posture/dimension combinations | Back-route on failure |
| **Embedded** | Non-blocking self-check by producing agent | Findings feed forward |
| **Advisory** | Non-blocking, always | Findings documented, routed, not enforced |

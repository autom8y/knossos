---
name: clinic-ref-index-schema
description: "Full index.yaml schema for clinic investigations. Use when: writing or validating index.yaml, understanding evidence/hypotheses/diagnosis fields, checking status field values. Triggers: index.yaml, investigation schema, evidence fields, hypothesis schema, diagnosis schema."
---

# index.yaml Schema

The shared coordination artifact. All agents read this first. Token cost: ~2-5k.

## Full Schema

```yaml
investigation: {kebab-case-slug}
created: {ISO 8601 timestamp}
status: {intake|examination|diagnosis|treatment|complete}
severity: {critical|high|medium|low}

symptoms:
  - id: S001
    description: "{what was observed}"
    reporter: {user|sre}
    timestamp: {ISO 8601}

systems:
  - name: {system-name}
    type: {ecs-service|monitoring|database|cloudwatch|etc}
    status: {flagged|examined|inaccessible}

evidence:
  - id: E001
    file: E001.txt
    system: {system-name}
    type: {error-log|config|metrics|task-state|query-result|circuit-breaker|etc}
    collected_by: pathologist
    timestamp: {ISO 8601}
    summary: "{factual description of file contents — NOT analysis}"

hypotheses:
  - id: H001
    statement: "{hypothesis statement}"
    status: {confirmed|eliminated|partial}
    evidence_for: [E001, E002]
    evidence_against: [E003]
    reasoning: "{why confirmed or eliminated}"

diagnosis:
  root_causes:
    - id: RC001
      description: "{root cause description}"
      confidence: {high|medium|low}
      evidence: [E001, E004]
      mechanism: "{how this cause produces the observed symptoms}"
  compound: {true|false}
  compound_interaction: "{how multiple root causes interact — only if compound: true}"
```

## Status Field Values

| Status | Set By | Meaning |
|--------|--------|---------|
| `intake` | triage-nurse | Investigation scoped, collection plan written |
| `examination` | pathologist | Evidence collected, all systems examined or marked inaccessible |
| `diagnosis` | diagnostician | Root cause identified with confidence >= medium |
| `treatment` | attending | Treatment plan and handoff artifacts produced |
| `complete` | attending | Investigation finished |
| `{phase}:{back_route}_round_{N}` | (mid back-route) | e.g., `examination:evidence_gap_round_2` |

## Evidence Summary Field Rules

Write factual descriptions, not analysis:
- Factual: "500 errors spiking at 14:23 UTC, DuckDB connection refused, 47 entries"
- NOT: "Database is overloaded" (that is interpretation)
- Factual: "Circuit breaker in OPEN state from 14:20 to 14:35 UTC, threshold: 5 failures in 10s"
- NOT: "Circuit breaker is misconfigured" (that is diagnosis)

## Evidence File Format

Each evidence file contains raw system output with a standard header:

```
# Evidence: E{NNN}
# System: {system-name}
# Type: {error-log|config|metrics|task-state|query-result|...}
# Collected: {ISO 8601 timestamp}
# Command: {command or query that produced this output}

{raw system output — unmodified}
```

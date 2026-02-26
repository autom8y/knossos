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

All investigation artifacts live under `.claude/wip/ERRORS/{investigation-slug}/`.

### Artifact Chain

```
.claude/wip/ERRORS/{slug}/
  intake-report.md        owner: triage-nurse     consumers: pathologist, diagnostician, attending
  index.yaml              owner: current-phase    consumers: all (shared coordination artifact)
  E001.txt                owner: pathologist      consumers: diagnostician, attending
  E002.txt                owner: pathologist      consumers: diagnostician, attending
  ...                     (E{NNN}.txt sequential)
  diagnosis.md            owner: diagnostician    consumers: attending
  treatment-plan.md       owner: attending        consumers: user
  handoff-10x-dev.md      owner: attending        consumers: 10x-dev rite (if applicable)
  handoff-sre.md          owner: attending        consumers: sre rite (if applicable)
  handoff-debt-triage.md  owner: attending        consumers: debt-triage rite (if applicable)
```

## index.yaml Schema

The shared coordination artifact. All agents read this first. Token cost: ~2-5k.

### Full Schema

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

### Status Field Values

| Status | Set By | Meaning |
|--------|--------|---------|
| `intake` | triage-nurse | Investigation scoped, collection plan written |
| `examination` | pathologist | Evidence collected, all systems examined or marked inaccessible |
| `diagnosis` | diagnostician | Root cause identified with confidence >= medium |
| `treatment` | attending | Treatment plan and handoff artifacts produced |
| `complete` | attending | Investigation finished |
| `{phase}:{back_route}_round_{N}` | (mid back-route) | e.g., `examination:evidence_gap_round_2` |

### Summary Field Rules (Evidence Entries)

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

## Back-Routes

The clinic's defining workflow characteristic. Expected patterns, not failure modes.

| Back-Route | Trigger | Source | Target | Max Iterations | On Limit |
|------------|---------|--------|--------|----------------|----------|
| `evidence_gap` | Diagnostician needs evidence not yet collected | diagnosis | examination (pathologist) | 3 | Escalate to user |
| `diagnosis_insufficient` | Attending cannot produce actionable fix spec | treatment | diagnosis (diagnostician) | 2 | Escalate to user |
| `scope_expansion` | Investigation involves systems outside original scope | diagnosis | intake (triage-nurse) | 1 | Hard limit: split investigation |

### evidence_gap Protocol

1. Diagnostician produces a targeted evidence request: system, data needed, why (which hypothesis)
2. Pythia routes directly to pathologist — does NOT re-gate intake or re-scope
3. Pathologist collects only the requested evidence
4. Pathologist writes to next sequential evidence file and updates index.yaml
5. Pathologist does NOT re-collect previously gathered evidence
6. Status field: `examination:evidence_gap_round_{N}`

### diagnosis_insufficient Protocol

1. Attending produces specific concerns: what is unclear, what additional depth is needed
2. Pythia routes back to diagnostician with the attending's concerns
3. Diagnostician deepens analysis (localizes affected code, clarifies mechanism, raises confidence)
4. Status field: `diagnosis:diagnosis_insufficient_round_{N}`

### scope_expansion Protocol

Requires user confirmation. User may prefer to open a second investigation rather than expand.

## Session Resume Protocol

On `/continue`, Pythia reads `index.yaml` status field to determine current phase.

| Status Field | Dispatch To |
|--------------|-------------|
| `intake` | triage-nurse |
| `examination` | pathologist |
| `diagnosis` | diagnostician |
| `treatment` | attending |
| `complete` | Report to user: investigation finished |
| `{phase}:{back_route}_round_{N}` | Agent for that phase, with back-route context |

If multiple investigation directories exist under `.claude/wip/ERRORS/`, Pythia presents the list to the user. It does not auto-select.

## Cross-Rite Handoff Formats

Handoff artifacts are always recommendations. User decides whether to act.

### handoff-10x-dev.md (required fields)

```markdown
# Handoff: 10x-dev
# Investigation: {slug}

## Root Cause Summary
{From diagnosis.md RC001/RC002 — not re-derived}

## Affected Files
- {specific file path}: {what to change}
- ...

## Fix Approach
{Recommended implementation strategy with rationale}

## Acceptance Criteria
{Specific, testable criteria for verifying the fix works}
- {e.g., checkout service returns 200 on /health after deploy}
- {e.g., error rate drops below 0.1% over 1 hour}

## Optional: Fix Ordering (compound failures)
{Which bug to fix first and why}

## Optional: Risk Assessment
{What could go wrong with the fix}

## Optional: Related Tests
{Existing tests to update or new tests to write}
```

### handoff-sre.md (required fields)

```markdown
# Handoff: SRE
# Investigation: {slug}

## Signals Missing
{What monitoring or observability was absent that caused or prolonged this incident}

## Recommended Alerts
- Alert: {name}
  Condition: {specific trigger condition}
  Rationale: {would have caught this incident at {timestamp}}

## Recommended Dashboards
- Dashboard: {name}
  Panels: {what metrics/signals to show}
  Rationale: {would have aided debugging by showing X}

## Optional: SLO Impact
{How this incident affected service level objectives}

## Optional: Runbook Recommendation
{Incident response procedure for this failure class}
```

### handoff-debt-triage.md (required fields)

```markdown
# Handoff: Debt Triage
# Investigation: {slug}

## Pattern Analysis
{How this issue manifests across the codebase — not just this instance}

## Scope of Problem
{How widespread the systemic issue is: N files, N services, estimated prevalence}

## Remediation Approach
{Long-term fix strategy, not just this instance}

## Optional: Debt Classification
{Type: architectural | dependency | test | design}

## Optional: Affected Components
{Broader list beyond the immediate investigation}

## Optional: Effort Estimate
{Rough sizing: S/M/L/XL}
```

## Complexity Level

The clinic uses exactly one level: INVESTIGATION. There is no complexity pre-classification.

All investigations run all four phases. Depth is emergent:
- Simple bug: ~4 agent invocations, ~30k tokens, no back-routes
- Compound failure: ~6-8 invocations, ~150k tokens, 1-2 back-routes

## Investigation Slug Naming

Descriptive, kebab-case, derived from the symptoms:
- `checkout-500-intermittent`
- `etl-silent-failures`
- `auth-latency-spike`
- `payment-timeout-cascade`

Not: `investigation-1`, `bug-2024-01-15`, `error-fix`

## Anti-Patterns

| Agent | Anti-Pattern | Correct Behavior |
|-------|-------------|------------------|
| triage-nurse | Premature diagnosis in intake report | Document symptoms, not theories |
| triage-nurse | Vague evidence collection plan | "Check the logs" → specify system, data type, time range |
| pathologist | Context hoarding (keeping evidence in context) | Write to E{NNN}.txt immediately |
| pathologist | Analytical summaries in index.yaml | Factual description only |
| diagnostician | Premature convergence | Check that ALL symptoms map to identified cause(s) |
| diagnostician | Re-running commands | Evidence is on disk — read it; if missing, trigger back-route |
| diagnostician | Loading all evidence files | Read index first, load selectively |
| attending | Re-diagnosing instead of using diagnosis.md | If insufficient, trigger back-route |
| attending | Vague acceptance criteria | Specific, testable conditions |
| attending | Ignoring monitoring gaps | If observability was absent, produce handoff-sre.md |

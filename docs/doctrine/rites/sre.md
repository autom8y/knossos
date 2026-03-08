---
last_verified: 2026-02-26
---

# Rite: sre

> Reliability engineering lifecycle for observability, coordination, and resilience testing.

The SRE rite provides workflows for maintaining production reliability through observability, incident coordination, and chaos engineering.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | sre |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | potnia |

---

## When to Use

- Assessing observability gaps
- Creating reliability plans and runbooks
- Implementing infrastructure improvements
- Testing system resilience
- Incident response coordination

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates reliability engineering initiative phases |
| **observability-engineer** | Designs observability strategy and establishes SLO/SLI baselines |
| **incident-commander** | Coordinates reliability plans and creates incident runbooks |
| **platform-engineer** | Implements infrastructure changes and reliability improvements |
| **chaos-engineer** | Designs and executes chaos experiments to verify resilience |

See agent files: `rites/sre/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[observation] --> B[coordination]
    B --> C[implementation]
    C --> D[resilience]
    D --> E[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| observation | observability-engineer | Observability Report | Always |
| coordination | incident-commander | Reliability Plan | complexity >= SERVICE |
| implementation | platform-engineer | Infrastructure Changes | Always |
| resilience | chaos-engineer | Resilience Report | Always |

---

## Invocation Patterns

```bash
# Quick switch to SRE
/sre

# Observability assessment
Task(observability-engineer, "assess observability gaps in payment service")

# Create runbook
Task(incident-commander, "create incident runbook for database failover")

# Chaos testing
Task(chaos-engineer, "design chaos experiment for network partition")
```

---

## Skills

- `doc-sre` — SRE documentation
- `sre-ref` — Workflow reference

---

## Source

**Manifest**: `rites/sre/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)

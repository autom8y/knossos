---
name: doc-sre
description: "SRE and operations templates for observability, incidents, chaos engineering, and infrastructure. Use when: writing postmortems, planning chaos experiments, auditing monitoring, designing pipelines. Triggers: observability, reliability, postmortem, incident, chaos engineering, SLO, MTTR, infrastructure change, pipeline design."
---

# SRE & Operations Templates

> Templates for site reliability engineering, incident management, and infrastructure operations.

## Purpose

Provides structured templates for SRE workflows: monitoring audits, reliability planning, incident response, chaos engineering, analytics instrumentation, infrastructure changes, CI/CD pipelines, and stakeholder communication during incidents.

## Template Catalog

| Template | Purpose | Agent |
|----------|---------|-------|
| [Observability Report](templates/observability-report.md) | Audit metrics, logging, tracing, alerting coverage | sre-analyst |
| [Reliability Plan](templates/reliability-plan.md) | Sprint/quarterly reliability priorities from incident patterns | sre-analyst |
| [Postmortem](templates/postmortem.md) | Blameless incident analysis with action items | incident-lead |
| [Incident Communication](templates/incident-communication.md) | Stakeholder notifications and status updates | incident-lead |
| [Chaos Experiment](templates/chaos-experiment.md) | Pre-registered failure injection test | chaos-engineer |
| [Resilience Report](templates/resilience-report.md) | Aggregate resilience assessment across experiments | chaos-engineer |
| [Tracking Plan](templates/tracking-plan.md) | Analytics event specification and validation | data-analyst |
| [Infrastructure Change](templates/infrastructure-change.md) | Change management with rollback and risk assessment | sre-engineer |
| [Pipeline Design](templates/pipeline-design.md) | CI/CD pipeline specification with DR planning | sre-engineer |

## When to Use Each Template

| Scenario | Template |
|----------|----------|
| Auditing monitoring coverage | Observability Report |
| Planning reliability work | Reliability Plan |
| After production incident (SEV-2+) | Postmortem |
| During active incident | Incident Communication |
| Before running chaos test | Chaos Experiment |
| After series of chaos tests | Resilience Report |
| Implementing analytics events | Tracking Plan |
| Scaling, migrating, or reconfiguring infra | Infrastructure Change |
| Designing new CI/CD pipeline | Pipeline Design |

## Quality Gates Summary

| Template | Gate Criteria |
|----------|---------------|
| **Observability Report** | Four pillars assessed, SLI/SLO proposals with baselines |
| **Reliability Plan** | Priorities tied to incidents, MTTR targets defined |
| **Postmortem** | Contributing factors identified, action items have owners |
| **Incident Communication** | Severity SLA met, updates include "what we know/doing" |
| **Chaos Experiment** | Hypothesis in Given/When/Then, abort criteria defined |
| **Resilience Report** | Scorecard covers core capabilities, gaps have remediation |
| **Tracking Plan** | Business questions documented, validation rules specified |
| **Infrastructure Change** | Rollback plan testable, communication plan complete |
| **Pipeline Design** | Stages have failure handling, DR scenarios covered |

## Progressive Disclosure

- [observability-report.md](templates/observability-report.md) - Monitoring/logging/tracing audit
- [reliability-plan.md](templates/reliability-plan.md) - Reliability priorities
- [postmortem.md](templates/postmortem.md) - Incident analysis
- [incident-communication.md](templates/incident-communication.md) - Incident notifications
- [chaos-experiment.md](templates/chaos-experiment.md) - Chaos test pre-registration
- [resilience-report.md](templates/resilience-report.md) - Resilience assessment
- [tracking-plan.md](templates/tracking-plan.md) - Analytics event spec
- [infrastructure-change.md](templates/infrastructure-change.md) - Change management
- [pipeline-design.md](templates/pipeline-design.md) - CI/CD pipeline spec

> **Note**: Technical Debt templates (Debt Ledger, Risk Matrix, Sprint Debt Packages) live in shared-templates skill.

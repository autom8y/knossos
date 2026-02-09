---
name: sre-catalog
description: "SRE rite agent profiles, command routing, and complexity levels. Use when: choosing SRE agents, understanding observability-engineer vs chaos-engineer roles, SRE command mapping. Triggers: observability engineer, incident commander, platform engineer, chaos engineer, SRE agents, SLO, SLI."
---

# SRE Rite Catalog

## Agent Profiles

### observability-engineer
**Color**: Orange | **Model**: Sonnet | **Invocation**: `Act as **Observability Engineer**`

Makes the invisible visible. Evaluates monitoring coverage, designs dashboards and alerting, defines SLIs and SLOs, instruments applications.

**Produces**: Observability reports, dashboard specifications, alert configurations, SLI/SLO definitions.

### incident-commander
**Color**: Purple | **Model**: Opus | **Invocation**: `Act as **Incident Commander**`

Runs the war room when systems are on fire. Coordinates incident response, conducts blameless postmortems, plans reliability improvements.

**Produces**: Reliability plans, postmortem documents, incident timelines, action item tracking.

### platform-engineer
**Color**: Cyan | **Model**: Sonnet | **Invocation**: `Act as **Platform Engineer**`

Builds the roads developers drive on. CI/CD pipeline improvements, infrastructure as code, developer environment optimization, deployment reliability.

**Produces**: Pipeline configurations, infrastructure code, developer tools, runbooks.

### chaos-engineer
**Color**: Red | **Model**: Opus | **Invocation**: `Act as **Chaos Engineer**`

Breaks production on purpose. Verifies resilience claims, tests failure scenarios, validates rollback procedures, pre-release resilience certification.

**Produces**: Chaos experiments, resilience reports, failure mode catalogs, runbook updates.

## Complexity Levels

| Level | Scope | Phases |
|-------|-------|--------|
| ALERT | Single alert/dashboard fix | implementation, resilience |
| SERVICE | Single service reliability | All 4 phases |
| SYSTEM | Multi-service SLOs/SLIs | All 4 phases |
| PLATFORM | Full platform reliability | All 4 phases |

## Command Mapping

When sre is active, commands route to these agents:

| Command | Routes To | Purpose |
|---------|-----------|---------|
| `/start` | observability-engineer | Begin with observability assessment |
| `/architect` | platform-engineer | Design infrastructure changes |
| `/build` | platform-engineer | Implement infrastructure changes |
| `/qa` | chaos-engineer | Validate resilience via chaos testing |
| `/hotfix` | incident-commander | Fast incident response path |
| `/code-review` | chaos-engineer | Review for reliability patterns |

## Cross-Rite Integration

| Situation | Hand off to |
|-----------|-------------|
| Reliability issues in code | 10x-dev for fixes |
| Documentation gaps in runbooks | docs rite |
| Technical debt causing incidents | debt-triage |
| Code quality affecting reliability | hygiene rite |

---
name: debt-catalog
description: "Debt-triage rite agent profiles and artifact schemas. Use when: understanding debt-collector vs risk-assessor roles, producing debt inventory YAML, planning remediation roadmaps. Triggers: debt collector, risk assessor, sprint planner, debt inventory, debt categories."
---

# Debt-Triage Rite Catalog

## Agent Profiles

### debt-collector
**Model**: Sonnet | **Invocation**: `Act as **Debt Collector**`

Systematically finds and documents technical debt across codebase. Use for initial inventory creation, quarterly scans, new codebase assessment, and portfolio health checks.

**Identifies**: Outdated dependencies, deprecated API usage, TODO/FIXME comments, workarounds/hacks, test coverage gaps, documentation debt, architectural drift, performance bottlenecks, security vulnerabilities, code duplication.

**Produces**: Debt inventory (structured YAML catalog), categories/tags, severity estimates, source locations.

### risk-assessor
**Model**: Sonnet | **Invocation**: `Act as **Risk Assessor**`

Prioritizes technical debt by risk, impact, and business value. Use after debt collection, before planning remediation, for ROI analysis.

**Assesses**: Impact (High/Medium/Low), Probability (%), Urgency (Critical/Soon/Eventually), Cost (effort), Value (velocity/quality/security gain).

**Produces**: Priority matrix, risk scores (impact x probability), ROI estimates (value/cost), urgency timeline.

### sprint-planner
**Model**: Sonnet | **Invocation**: `Act as **Sprint Planner**`

Creates actionable plans to pay down technical debt. Use after risk assessment, for planning sprints, creating quarterly roadmaps.

**Produces**: Sprint-sized tasks, remediation roadmaps, effort estimates, dependency graphs, success criteria, progress metrics.

## Debt Inventory Schema

```yaml
scan_date: "YYYY-MM-DD"
scope: "platform-wide | service-name"
items:
  - id: DEBT-NNN
    category: security | performance | testing | dependencies | architecture | documentation
    severity: critical | high | medium | low
    title: "Short description"
    location: "file/path:line"
    impact: "Business/technical impact statement"
    effort_estimate: "hours | days | weeks"
    priority: N
```

## Debt Categories

| Category | Examples |
|----------|----------|
| Code quality | Duplication, complexity, smells |
| Testing | Coverage gaps, flaky tests |
| Security | Vulnerabilities, outdated auth |
| Performance | Bottlenecks, inefficiencies |
| Dependencies | Outdated libraries, deprecated APIs |
| Documentation | Missing/outdated docs |
| Architecture | ADR violations, drift |
| Infrastructure | Manual processes, missing automation |

## Debt vs Hygiene

| Debt Rite | Hygiene Rite |
|-----------|--------------|
| Strategic debt management | Tactical code cleanup |
| Portfolio/project-level | Module/file-level |
| Inventories, roadmaps, plans | Refactored code, cleanliness |
| Quarterly/annual planning | Sprint/week execution |

**Workflow**: Use `/debt` to plan, `/hygiene` to execute.

# SRE Pack

Reliability lifecycle for observability, incident management, infrastructure, and resilience testing.

## When to Use This Team

**Triggers**:
- "We need better monitoring and alerting"
- "Production is down and we need coordination"
- "How do we improve system reliability?"
- "Can our infrastructure handle this failure scenario?"
- "We keep getting surprised by outages"
- "Our alerts are too noisy"

**Not for**: Feature development, application code, or debt management

## Quick Start

```bash
/team sre-pack
```

## Agents

| Agent | Role | Model | Artifact |
|-------|------|-------|----------|
| observability-engineer | Metrics, logs, traces ownership; SLI/SLO definition | claude-sonnet-4-5 | Observability Report |
| incident-commander | War room coordination, postmortems, reliability planning | claude-opus-4-5 | Reliability Plan |
| platform-engineer | CI/CD pipelines, IaC, deployment automation | claude-sonnet-4-5 | Infrastructure Changes |
| chaos-engineer | Fault injection, resilience verification, breaking point discovery | claude-opus-4-5 | Resilience Report |

## Workflow

Four-phase sequential workflow: **Observation → Coordination → Implementation → Resilience**

**Complexity Levels**:
- **ALERT**: Quick tactical fixes (skip observation/coordination)
- **SERVICE**: Single service reliability improvements
- **SYSTEM**: Multi-service SLO/SLI work
- **PLATFORM**: Full platform reliability initiatives

See `workflow.md` for detailed phase flow and entry criteria.

## Related Teams

- **debt-triage-pack**: When reliability issues stem from accumulated technical debt

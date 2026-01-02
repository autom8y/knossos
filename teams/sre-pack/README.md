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

| Agent | Role | Artifact |
|-------|------|----------|
| observability-engineer | Metrics, logs, traces ownership; SLI/SLO definition | Observability Report |
| incident-commander | War room coordination, postmortems, reliability planning | Reliability Plan |
| platform-engineer | CI/CD pipelines, IaC, deployment automation | Infrastructure Changes |
| chaos-engineer | Fault injection, resilience verification, breaking point discovery | Resilience Report |

## Workflow

Four-phase sequential workflow: **Observation → Coordination → Implementation → Resilience**

**Complexity Levels**:
- **ALERT**: Quick tactical fixes (skip observation/coordination)
- **SERVICE**: Single service reliability improvements
- **SYSTEM**: Multi-service SLO/SLI work
- **PLATFORM**: Full platform reliability initiatives

See `workflow.md` for detailed phase flow and entry criteria.

## Handoff Acceptance from 10x-dev-pack

SRE accepts validation handoffs from 10x-dev-pack for production readiness assessment.

**Expected HANDOFF Format** (see `cross-team-handoff` skill for full schema):
```yaml
---
source_team: 10x-dev-pack
target_team: sre-pack
handoff_type: validation
---
```

### What SRE Expects in the Handoff

| Section | Required Content |
|---------|-----------------|
| **Context** | Feature summary, testing completed, deployment timeline |
| **Source Artifacts** | TDD, implementation paths, test results |
| **Items** | Specific validation requests with scope |
| **Notes** | Environment details, known risks, contacts |

### Acceptance Criteria

Before accepting a validation handoff, SRE verifies:

- [ ] `handoff_type: validation` is specified
- [ ] Source artifacts are accessible and complete
- [ ] Each validation item has clear scope definition
- [ ] Expected behavior is documented (not just "validate this")
- [ ] Staging/test environment is available for validation

### Validation Scope by Item Type

| Validation Type | SRE Focus |
|-----------------|-----------|
| Database migrations | Zero-downtime path, rollback procedure, data integrity |
| API changes | Rate limiting, backward compatibility, error handling |
| Infrastructure changes | Resource utilization, scaling behavior, failure modes |
| Monitoring/alerting | SLI/SLO coverage, alert thresholds, dashboard completeness |
| Performance | Load testing results, latency requirements, resource bounds |

### What SRE Produces

After validation, SRE produces one of:

1. **GO**: Validation passed, ready for production
2. **NO-GO**: Blocking issues found, handoff returned with defects
3. **CONDITIONAL**: Approved with documented risks or required follow-ups

Results are documented in a Reliability Report or returned as defect annotations on the original HANDOFF.

### Example Validation Response

```markdown
## SRE Validation Response

**Original HANDOFF**: HANDOFF-10x-sre-user-auth-2026-01-02.md
**Verdict**: CONDITIONAL

### VAL-001: Database migration safety - PASSED
- Zero-downtime migration verified
- Rollback tested successfully
- Data integrity preserved

### VAL-002: Rate limiting - PASSED WITH CONDITIONS
- 1000 req/min verified under load
- **Condition**: Add alert for 80% threshold breach before production

### VAL-003: Monitoring - NEEDS WORK
- Key metrics dashboarded
- **Blocker**: Error rate alert not configured (required for GO)
```

## Related Teams

- **10x-dev-pack**: Primary source of validation handoffs for new features
- **debt-triage-pack**: When reliability issues stem from accumulated technical debt

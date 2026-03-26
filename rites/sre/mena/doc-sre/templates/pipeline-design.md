---
description: "Pipeline Design Template companion for templates skill."
---

# Pipeline Design Template

> CI/CD pipeline specification with stages, security, monitoring, and disaster recovery.

```markdown
# Pipeline Design: [Pipeline Name]

## Overview
**Purpose**: [What this pipeline does]
**Team**: [Owning team]
**Status**: [Draft / In Review / Approved / Implemented]

## Pipeline Summary

### Trigger
[What starts this pipeline - commit, schedule, manual, etc.]

### Stages
| Stage | Purpose | Duration | Failure Mode |
|-------|---------|----------|--------------|
| [name] | [description] | [time] | [what happens on fail] |

### Artifacts
- **Input**: [what pipeline consumes]
- **Output**: [what pipeline produces]

## Stage Details

### Stage: [Name]
**Purpose**: [Why this stage exists]

**Steps**:
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Success Criteria**: [What indicates success]
**Failure Handling**: [What happens on failure]
**Timeout**: [max duration]

### Stage: [Name]
[Repeat for each stage]

## Environment & Infrastructure

### Execution Environment
- **Platform**: [Jenkins, GHA, GitLab, etc.]
- **Agent Type**: [specs]
- **Parallelism**: [max concurrent runs]
- **Resource Limits**: [CPU, memory, disk]

### Dependencies
| Dependency | Type | Version | Purpose |
|------------|------|---------|---------|
| [tool/service] | [external/internal] | [version] | [why needed] |

## Security & Compliance

### Secrets Management
| Secret | Storage | Rotation | Access Control |
|--------|---------|----------|----------------|
| [name] | [where] | [frequency] | [who can access] |

### Compliance Requirements
- [ ] [Requirement 1]
- [ ] [Requirement 2]

## Monitoring & Alerting

### Metrics
- Pipeline success rate
- Average duration
- Stage failure rate
- Queue time

### Alerts
| Alert | Condition | Severity | Recipients |
|-------|-----------|----------|------------|
| [name] | [trigger] | [level] | [who to notify] |

## Testing Strategy

### Unit Tests
[How stages are unit tested]

### Integration Tests
[How full pipeline is tested]

### Rollout Plan
1. [Phase 1 - e.g., test in dev]
2. [Phase 2 - e.g., canary in staging]
3. [Phase 3 - e.g., full rollout]

## Disaster Recovery

### Failure Scenarios
| Scenario | Detection | Recovery | RTO |
|----------|-----------|----------|-----|
| [scenario] | [how to detect] | [how to recover] | [time] |

### Rollback Procedure
```
[How to roll back to previous pipeline version]
```

## Cost Analysis
- **Compute**: [estimated cost]
- **Storage**: [estimated cost]
- **External Services**: [estimated cost]
- **Total Monthly**: [estimated total]

## Open Questions
- [ ] [Question 1]
- [ ] [Question 2]

## Approvals
- **Engineering**: [name, date]
- **Security**: [name, date]
- **SRE**: [name, date]
```

## Quality Gate

**Pipeline Design complete when:**
- All stages have failure handling and timeouts
- Secrets management documented with rotation policy
- Disaster recovery scenarios cover critical failures
- Cost analysis completed with monthly estimates

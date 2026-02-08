---
name: doc-sre
description: "SRE templates: observability, incidents, chaos engineering. Triggers: observability, reliability, postmortem, incident, chaos engineering, SLO, MTTR."
---

# SRE & Analytics Documentation Templates

> **Status**: Complete (Session 2)

## Template Index

This skill provides templates for:

- **Observability & Monitoring**: [Observability Report](#observability-report-template)
- **Reliability Planning**: [Reliability Plan](#reliability-plan-template)
- **Incident Management**: [Postmortem](#postmortem-template)
- **Chaos Engineering**: [Chaos Experiment](#chaos-experiment-template), [Resilience Report](#resilience-report-template)
- **Analytics**: [Tracking Plan](#tracking-plan-template)
- **Infrastructure**: [Infrastructure Change](#infrastructure-change-template)
- **Pipeline & Deployment**: [Pipeline Design](#pipeline-design-template)
- **Communications**: [Incident Communication](#incident-communication-template)

> **Note**: Technical Debt templates (Debt Ledger, Risk Matrix, Sprint Debt Packages) have moved to `@shared-templates` for cross-rite use.

## Migrated Templates

The following templates have moved to `@shared-templates` for cross-rite use:
- Debt Ledger → `@shared-templates#debt-ledger-template`
- Risk Matrix → `@shared-templates#risk-matrix-template`
- Sprint Debt Packages → `@shared-templates#sprint-debt-packages-template`

---

## Observability Report Template {#observability-report-template}

```markdown
# Observability Report: [System/Service]

## Executive Summary
[One paragraph: Current state, critical gaps, top recommendations]

## Scope
- Services analyzed: [list]
- Time period: [dates]
- Data sources: [metrics/logs/traces systems]

## Current State

### Metrics
| Service | Golden Signals | Custom Metrics | Gaps |
|---------|----------------|----------------|------|
| [name]  | [coverage %]   | [count]        | [list] |

### Logging
| Service | Structured | Correlation IDs | Retention |
|---------|------------|-----------------|-----------|
| [name]  | [yes/no]   | [yes/no]        | [days]    |

### Tracing
| Service | Instrumented | Sample Rate | Coverage |
|---------|--------------|-------------|----------|
| [name]  | [yes/no]     | [%]         | [%]      |

### Alerting
| Alert Category | Count | False Positive Rate | Actions |
|----------------|-------|---------------------|---------|
| Critical       | [n]   | [%]                 | [types] |

## Gap Analysis

### Critical Gaps (Must Fix)
1. [Gap]: [Impact] → [Recommendation]

### Important Gaps (Should Fix)
1. [Gap]: [Impact] → [Recommendation]

### Nice-to-Have Improvements
1. [Improvement]: [Benefit]

## Recommendations

### Quick Wins (< 1 week)
1. [Action]: [Expected outcome]

### Medium-Term (1-4 weeks)
1. [Action]: [Expected outcome]

### Long-Term (> 1 month)
1. [Action]: [Expected outcome]

## SLI/SLO Proposals

| Service | SLI | Current | Proposed SLO | Error Budget |
|---------|-----|---------|--------------|--------------|
| [name]  | [availability] | [%] | [%] | [hours/month] |

## Next Steps
1. [Immediate action]
2. [Follow-up]
```

---

## Reliability Plan Template {#reliability-plan-template}

```markdown
# Reliability Plan: [Period/Focus]

## Summary
[One paragraph: Current reliability state, key priorities, expected outcomes]

## Incident Analysis

### Recent Incidents
| Date | Severity | Duration | Impact | Postmortem |
|------|----------|----------|--------|------------|
| [date] | [SEV] | [time] | [description] | [link] |

### Pattern Analysis
[What patterns emerge from recent incidents?]

## Priorities

### Critical (This Sprint)
| Item | Owner | Due Date | Status | Incident(s) |
|------|-------|----------|--------|-------------|
| [action] | [name] | [date] | [status] | [refs] |

### Important (This Quarter)
| Item | Owner | Due Date | Status | Incident(s) |
|------|-------|----------|--------|-------------|
| [action] | [name] | [date] | [status] | [refs] |

### Backlog (Future)
1. [Item]: [Brief description]

## Metrics
- MTTR (Mean Time to Recovery): [current] → [target]
- Incident Rate: [current] → [target]
- Action Item Completion Rate: [%]

## Next Review
[Date for next reliability review]
```

---

## Postmortem Template {#postmortem-template}

```markdown
# Postmortem: [Incident Title]

**Date**: [incident date]
**Duration**: [start time] - [end time] ([total hours])
**Severity**: [SEV level]
**Authors**: [postmortem participants]
**Status**: [Draft / Final]

## Summary
[2-3 sentences: What happened, what was the impact, how was it resolved]

## Impact
- Users affected: [count or percentage]
- Revenue impact: [if applicable]
- Duration of impact: [time]
- Services affected: [list]

## Timeline
| Time (UTC) | Event |
|------------|-------|
| [time] | [what happened] |

## Contributing Factors
1. **[Factor category]**: [Description of how this contributed]
2. **[Factor category]**: [Description of how this contributed]

## What Went Well
- [Thing that helped during incident]
- [Thing that worked as designed]

## What Went Poorly
- [Thing that made incident worse or longer]
- [Gap that was exposed]

## Where We Got Lucky
- [Thing that could have made it worse but didn't]

## Action Items
| Action | Owner | Due Date | Priority | Status |
|--------|-------|----------|----------|--------|
| [specific action] | [name] | [date] | [P1/P2/P3] | [status] |

## Lessons Learned
[What should we remember from this incident?]

## References
- [Link to incident Slack channel]
- [Link to relevant dashboards]
- [Link to deploy logs]
```

---

## Chaos Experiment Template {#chaos-experiment-template}

```markdown
# Chaos Experiment: [Name]

## Metadata
- **Date**: [execution date]
- **Target**: [service/system]
- **Environment**: [dev/staging/prod]
- **Engineer**: [name]

## Hypothesis
**Given**: [steady state description]
**When**: [failure condition]
**Then**: [expected behavior]

## Steady State Definition
| Metric | Normal Range | Measurement |
|--------|--------------|-------------|
| Request rate | [range] | [source] |
| Error rate | [range] | [source] |
| Latency p99 | [range] | [source] |

## Experiment Design

### Failure Type
[Network / Process / Resource / Dependency]

### Injection Method
```
[How failure will be introduced - e.g., toxiproxy, tc, kill -9]
```

### Blast Radius
- **Scope**: [% of traffic / # of instances]
- **Duration**: [time]
- **Affected Users**: [estimate]

### Abort Criteria
- Error rate > [threshold]
- Latency p99 > [threshold]
- [Other conditions]

### Rollback Plan
```
[How to remove the failure condition]
```

## Execution Log
| Time | Action | Observation |
|------|--------|-------------|
| [time] | [action] | [what happened] |

## Results

### Outcome
**[PASS / PARTIAL / FAIL / ABORT]**

### Observations
[What actually happened vs. hypothesis]

### Gaps Discovered
1. [Gap]: [Impact] → [Recommendation]

### Evidence
[Links to dashboards, logs, screenshots]

## Action Items
| Action | Owner | Priority | Due |
|--------|-------|----------|-----|
| [action] | [name] | [P1/P2/P3] | [date] |

## Lessons Learned
[What did we learn that applies beyond this experiment?]
```

---

## Resilience Report Template {#resilience-report-template}

```markdown
# Resilience Report: [System/Service]

## Executive Summary
[One paragraph: Overall resilience posture, critical findings, top recommendations]

## Scope
- Services tested: [list]
- Time period: [dates]
- Environments: [dev/staging/prod]
- Experiment count: [number]

## Experiments Summary
| Experiment | Target | Result | Critical Findings |
|------------|--------|--------|-------------------|
| [name] | [service] | [PASS/FAIL] | [findings] |

## Resilience Scorecard
| Capability | Status | Evidence |
|------------|--------|----------|
| Database failover | [PASS/FAIL] | [experiment ref] |
| Circuit breakers | [PASS/FAIL] | [experiment ref] |
| Graceful degradation | [PASS/FAIL] | [experiment ref] |
| Auto-recovery | [PASS/FAIL] | [experiment ref] |
| Rollback procedures | [PASS/FAIL] | [experiment ref] |

## Critical Gaps
| Gap | Impact | Priority | Remediation |
|-----|--------|----------|-------------|
| [gap] | [impact] | [P1/P2/P3] | [fix] |

## Recommendations

### Immediate (This Week)
1. [Action]: [Expected improvement]

### Short-Term (This Month)
1. [Action]: [Expected improvement]

### Long-Term (This Quarter)
1. [Action]: [Expected improvement]

## Failure Mode Catalog
| Mode | Detection | Impact | Mitigation |
|------|-----------|--------|------------|
| [failure] | [how detected] | [blast radius] | [how to mitigate] |

## Next Steps
1. [Immediate action]
2. [Follow-up experiments]
3. [Remediation tracking]
```

---

## Tracking Plan Template {#tracking-plan-template}

```markdown
# TRACK-{slug}

## Overview
{What user journey or feature this tracks}

## Business Questions
- {Question 1 this data answers}
- {Question 2}

## Naming Convention
{event_category_action, e.g., onboarding_step_completed}

## Events

### {event_name}
- **Trigger**: {When this event fires}
- **Category**: {Funnel step, engagement, error, etc.}
- **Platform**: {Web, iOS, Android, Server}

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| {property} | {string/int/bool} | {Yes/No} | {What it represents} |

### Validation Rules
- {Rule 1, e.g., "step_number must be 1-5"}
- {Rule 2}

## Implementation Notes
{Code examples, edge cases, gotchas}

## QA Checklist
- [ ] Events fire on expected triggers
- [ ] All required properties present
- [ ] Property values within expected ranges
- [ ] No duplicate events
- [ ] Works across platforms
```

---

## Infrastructure Change Template {#infrastructure-change-template}

```markdown
# Infrastructure Change: [System/Component]

## Change Summary
**Date**: [planned date]
**Engineer**: [name]
**Urgency**: [Standard / Expedited / Emergency]
**Risk Level**: [Low / Medium / High / Critical]

## Change Details

### What's Changing
[Description of infrastructure change]

### Motivation
[Why this change is needed]

### Systems Affected
| System | Component | Impact Level | Downtime Expected |
|--------|-----------|--------------|-------------------|
| [name] | [component] | [Low/Med/High] | [yes/no - duration] |

## Pre-Change State
[Current configuration, capacity, topology]

## Post-Change State
[Target configuration, capacity, topology]

## Implementation Plan

### Prerequisites
- [ ] [Prerequisite 1]
- [ ] [Prerequisite 2]

### Change Steps
1. [Step 1 with expected outcome]
2. [Step 2 with expected outcome]
3. [Step 3 with expected outcome]

### Estimated Duration
- Preparation: [time]
- Execution: [time]
- Verification: [time]
- **Total**: [time]

### Maintenance Window
- Start: [date/time]
- End: [date/time]
- Timezone: [tz]

## Risk Assessment

### Risks
| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| [risk] | [Low/Med/High] | [description] | [how to prevent] |

### Rollback Plan
```
[Step-by-step rollback procedure]
[Include: how to detect need for rollback]
[Include: time window for rollback decision]
```

### Abort Criteria
- [Condition that triggers abort]
- [Condition that triggers abort]

## Testing & Verification

### Pre-Change Validation
- [ ] [Check 1]
- [ ] [Check 2]

### Post-Change Verification
- [ ] [Verification 1]
- [ ] [Verification 2]

### Success Criteria
- [Metric 1]: [expected value]
- [Metric 2]: [expected value]

## Communication Plan

### Stakeholders
- **Notify before**: [list]
- **Notify during**: [list]
- **Notify after**: [list]

### Communication Template
```
Subject: [change summary]
Start: [time]
Expected impact: [description]
```

## Execution Log
| Time | Action | Result | Notes |
|------|--------|--------|-------|
| [time] | [action] | [result] | [notes] |

## Post-Change Review
- Change successful: [yes/no]
- Actual duration: [time]
- Incidents triggered: [count]
- Lessons learned: [list]
```

---

## Pipeline Design Template {#pipeline-design-template}

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

---

## Incident Communication Template {#incident-communication-template}

```markdown
# Incident Communication: SEV-[N]

## Initial Notification

**Subject**: [SEV-N] [Brief description]

**Status**: INVESTIGATING / IDENTIFIED / MONITORING / RESOLVED

**Detected**: [time]
**Impact**: [description]
**Affected users**: [estimate or "investigating"]
**Services affected**: [list]

**Current actions**: [what we're doing right now]

**Next update**: [time] or when status changes

---

## Status Update Template

**Time**: [timestamp]
**Status**: [current status]

**What we know**:
- [Finding 1]
- [Finding 2]

**What we're doing**:
- [Action 1]
- [Action 2]

**Impact**:
- [Current impact assessment]
- [Change from last update]

**Next update**: [time]

---

## Resolution Notification

**Subject**: [RESOLVED] [SEV-N] [Brief description]

**Resolution time**: [time]
**Total duration**: [hours/minutes]

**Root cause** (brief): [one-sentence explanation]

**Resolution**: [what fixed it]

**Impact summary**:
- Users affected: [final count]
- Services affected: [list]
- Duration: [time]

**Follow-up**:
- Postmortem: [link or ETA]
- Action items: [count] tracked in [location]

**Timeline**:
| Time | Event |
|------|-------|
| [time] | [event] |

---

## Communication Guidelines

### Severity-Based Frequency

| Severity | Initial | Updates | Final |
|----------|---------|---------|-------|
| SEV-1 (Critical) | Immediate | Every 30 min | Immediate |
| SEV-2 (High) | Within 15 min | Every 1 hour | Within 1 hour |
| SEV-3 (Medium) | Within 1 hour | Every 4 hours | Within 4 hours |
| SEV-4 (Low) | Within 4 hours | Daily | When resolved |

### Channels
- **Internal**: [Slack channel, email list]
- **External**: [Status page, customer email]
- **Stakeholders**: [Executive notification criteria]

### Tone Guidelines
- **Be clear**: Avoid jargon, explain technical terms
- **Be honest**: Don't minimize or speculate
- **Be timely**: Better to say "investigating" than go silent
- **Be specific**: "3% of API requests" not "some users"
- **Be empathetic**: Acknowledge impact on users

### What NOT to Say
- "Oops" or apologetic language in updates (save for final)
- Root cause speculation without evidence
- Blame (teams, vendors, systems)
- Minimizing language ("just" or "only")
- Promises without confidence ("will be fixed in 10 minutes")
```

---

## Usage Guidelines

### When to Use Which Template

**Observability Report**: Auditing current monitoring/logging/tracing coverage
**Reliability Plan**: Quarterly/sprint planning for reliability work
**Postmortem**: After any production incident (severity 2+)
**Chaos Experiment**: Before running any chaos engineering test
**Resilience Report**: After series of chaos experiments to assess overall posture
**Tracking Plan**: Before implementing new analytics events
**Infrastructure Change**: Planning any infrastructure modification (scaling, migrations, config changes)
**Pipeline Design**: Designing new CI/CD pipelines or major pipeline refactors
**Incident Communication**: During active incidents to maintain consistent stakeholder updates

> **For debt workflows**: Use `@shared-templates` for Debt Ledger, Risk Matrix, and Sprint Debt Packages templates.

### Integration with Development Workflow

These templates complement the core development workflow (PRD/TDD/ADR/Test Plan):

- **Observability Reports** → inform TDD non-functional requirements
- **Postmortems** → generate ADRs for architectural changes
- **Chaos Experiments** → validate TDD reliability assumptions
- **Tracking Plans** → become requirements in feature PRDs

For debt-related workflows (Debt Ledgers, Risk Matrices, Sprint Debt Packages), see `@shared-templates`.

### Related Skills

- **documentation** - Core PRD/TDD/ADR/Test Plan templates
- **10x-workflow** - Agent coordination for multi-phase work
- **standards** - Code quality and architectural conventions

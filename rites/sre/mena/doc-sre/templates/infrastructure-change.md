# Infrastructure Change Template

> Change management document for infrastructure modifications with risk assessment and rollback.

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

## Quality Gate

**Infrastructure Change complete when:**
- Rollback plan documented and independently testable
- Abort criteria defined before execution
- Communication plan covers all stakeholders
- Post-change verification criteria measurable

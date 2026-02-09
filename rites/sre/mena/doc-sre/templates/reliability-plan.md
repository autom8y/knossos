# Reliability Plan Template

> Quarterly or sprint-level planning for reliability work, driven by incident patterns.

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

## Quality Gate

**Reliability Plan complete when:**
- Recent incidents analyzed with pattern identification
- Priorities tied to incident evidence
- MTTR and incident rate targets defined
- Next review date scheduled

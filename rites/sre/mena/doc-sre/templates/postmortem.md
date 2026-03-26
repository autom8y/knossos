---
description: "Postmortem Template companion for templates skill."
---

# Postmortem Template

> Blameless incident analysis documenting timeline, contributing factors, and action items.

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

## Quality Gate

**Postmortem complete when:**
- Timeline is accurate with UTC timestamps
- Contributing factors identified (not root cause singular)
- Action items have owners and due dates
- "Where We Got Lucky" section populated (forces deeper thinking)

---
description: "Incident Communication Template companion for templates skill."
---

# Incident Communication Template

> Structured incident notifications, status updates, and resolution communications.

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

## Quality Gate

**Incident Communication complete when:**
- Severity level assigned with matching update frequency
- Initial notification sent within severity SLA
- All updates include "what we know" and "what we're doing"
- Resolution notification includes postmortem ETA

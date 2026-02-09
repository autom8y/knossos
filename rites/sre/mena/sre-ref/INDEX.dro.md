---
name: sre
description: "Switch to SRE rite (reliability workflow). Use when: user says /sre, wants observability, incident response, chaos engineering, platform reliability. Triggers: /sre, reliability rite, SRE workflow, observability, chaos engineering."
context: fork
---

# /sre - Switch to Reliability Rite

Switch to sre, the Site Reliability Engineering rite.

## Behavior

### 1. Invoke Rite Switch

Execute via Bash tool:

```bash
ari sync --rite sre
```

### 2. Display Knossos

After successful switch, show the agent table:

| Agent | Role |
|-------|------|
| observability-engineer | Metrics, logs, traces, dashboards, alerts |
| incident-commander | War room coordination, postmortems |
| platform-engineer | CI/CD, IaC, deployment automation |
| chaos-engineer | Fault injection, resilience testing |

### 3. Update Session

If a session is active, update `active_rite` to `sre`.

## When to Use

- Observability improvements (dashboards, alerts, SLOs)
- Incident response and post-incident analysis
- Platform and infrastructure work
- Chaos engineering and resilience testing

**Don't use for**: Feature development --> `/10x` | Documentation --> `/docs` | Code quality --> `/hygiene` | Debt triage --> `/debt`

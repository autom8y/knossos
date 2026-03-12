---
description: 'File structured complaints about framework friction to .sos/wip/complaints/. Use when: encountering framework bugs, CLI surface drift, missing skills, broken hooks, routing failures, context degradation, tool errors that indicate prompting gaps. Triggers: complaint, friction, framework problem, broken, missing feature, drift, file complaint.'
name: complaint-filing
version: "1.0"
---
---
name: complaint-filing
description: "File structured complaints about framework friction to .sos/wip/complaints/. Use when: encountering framework bugs, CLI surface drift, missing skills, broken hooks, routing failures, context degradation, tool errors that indicate prompting gaps. Triggers: complaint, friction, framework problem, broken, missing feature, drift, file complaint."
---

# Complaint Filing (Cassandra Protocol)

> Structured write path for agents to report framework friction. File and return to your primary task.

## When to File

File a complaint when you encounter friction that is **about the framework itself**, not about the user's project:

- A CLI command doesn't exist or returns unhelpful results
- A skill is missing or doesn't cover your use case
- A hook blocks something it shouldn't (or allows something it shouldn't)
- Routing fails — you can't find the right rite, agent, or workflow
- Context is degraded — instructions from earlier in the session are being ignored
- Documentation is stale — .know/ or skill content contradicts current behavior

**Do NOT file** complaints about: user code quality, external tool failures, model limitations, or one-off errors that resolve on retry.

## Quick-File (30 seconds)

For most observations. Write this YAML to `.sos/wip/complaints/COMPLAINT-{YYYYMMDD}-{HHMMSS}-{your-agent-name}.yaml`:

```yaml
id: COMPLAINT-{YYYYMMDD}-{HHMMSS}-{your-agent-name}
filed_by: {your-agent-name}
filed_at: {ISO-8601 timestamp}
title: "{short description, max 120 chars}"
severity: low | medium | high | critical
description: |
  {What happened. What you expected. What actually occurred.}
tags: []
status: filed
```

## Deep-File (when evidence is strong)

For high-severity or well-understood friction. Extends quick-file with:

```yaml
evidence:
  session_id: "{session-id if available}"
  event_refs: []
  context: "{what led to the friction}"
suggested_fix: |
  {Specific proposed resolution}
effort_estimate: trivial | small | medium | large | epic
related_scars: []
zone: parameter | behavior | structure
```

**Zone classification** (from three-zone model):
- `parameter`: Quantitative knobs (token budgets, routing weights, skill order). Auto-tunable.
- `behavior`: Qualitative instructions (prompt wording, routing rules, skill content). Human-gated.
- `structure`: Architectural patterns (agent taxonomy, session FSM, materialization). Never auto-modify.

## Before Filing: Dedup Check

1. Read the contents of `.sos/wip/complaints/` directory
2. Scan titles of existing `filed` complaints
3. If a similar complaint exists: add a corroborating note to its `description` field instead of creating a new file
4. If no similar complaint exists: create a new file

## Filing Rules

1. **Complete your primary task first.** Complaint filing is a side effect, not your main job.
2. **Write-only during normal work.** Never read complaints to inform your primary task.
3. **One complaint per friction.** Don't bundle unrelated observations.
4. **Be specific.** "ari ask returns 0 results for 'switch rite'" > "search doesn't work."
5. **Severity guide:**
   - `low`: Minor inconvenience, workaround exists
   - `medium`: Noticeable friction, required fallback approach
   - `high`: Significant time lost, no clean workaround
   - `critical`: Blocked primary task, required human intervention

## Directory

Complaints live at `.sos/wip/complaints/`. This directory may not exist — create it if needed.

---
name: orchestrator-templates
description: |
  Shared YAML schemas for orchestrator consultation patterns used across all Potnia agents.
  Use when: consulting an orchestrator about next steps, structuring phase transition requests,
  formatting consultation responses, or routing to a specialist.
  Triggers: orchestrator consultation, consultation request, consultation response,
  phase transition, specialist routing, invoke_specialist, await_user, startup consultation.
invokable: skill
---

# Orchestrator Templates

> Canonical YAML schemas for orchestrator consultation — used by every Potnia agent in the ecosystem.

## Communication Pattern

```
Main Agent → CONSULTATION_REQUEST → Potnia → CONSULTATION_RESPONSE → Main Agent
```

The orchestrator is a stateless advisor. It reads structured context and returns structured directives. The main agent controls all execution.

## Consultation Request — Key Fields

| Field | Type | Purpose |
|-------|------|---------|
| `consultation.type` | enum | `startup` / `continuation` / `failure` / `completion` |
| `initiative.title` | string | Brief description of overall work |
| `initiative.goal` | string | What we're trying to achieve |
| `state.current_phase` | string | Phase we're in, or `"none"` |
| `state.completed_phases` | list | Phases already done |
| `state.blocked_on` | string | Active blocker, or `"none"` |
| `results.last_specialist` | string | Last agent that acted, or `"none"` |
| `results.last_outcome` | enum | `success` / `failure` / `partial` |
| `results.artifacts_ready` | list | `"name: location"` pairs |
| `context_summary` | text | Key context — keep concise |

## Consultation Response — Key Fields

| Field | Type | Purpose |
|-------|------|---------|
| `directive.action` | enum | `invoke_specialist` / `await_user` / `complete` |
| `directive.rationale` | string | Why this is the right next step |
| `specialist.agent` | string | Agent name (required if action = invoke_specialist) |
| `specialist.prompt` | text | Self-contained prompt for the specialist |
| `information_needed` | list | Questions for main agent to resolve |
| `user_question.prompt` | string | Question to escalate to user |
| `state_update.current_phase` | string | New phase after this directive |
| `throughline.rationale` | string | Why these decisions make sense together |
| `throughline.risks` | string | Known risks or tradeoffs |

## When to Load Schemas

| Schema | Load When |
|--------|-----------|
| `schemas/consultation-request.md` | Structuring a request to send to Potnia |
| `schemas/consultation-response.md` | Interpreting or formatting Potnia's directive |

## Companion Reference

| File | When to Load |
|------|--------------|
| `schemas/consultation-request.md` | Full schema + field descriptions for requests |
| `schemas/consultation-response.md` | Full schema + field descriptions for responses |

## Related Skills

- `10x-workflow` skill — Agent coordination patterns and phase sequencing
- `prompting` skill — Agent invocation guidance

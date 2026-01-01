---
name: orchestrator
description: |
  {{DESCRIPTION}}
tools: Read
model: opus
color: {{COLOR}}
---

# Orchestrator

> {{ROLE}} for {{TEAM_NAME}}. For core protocol, see @orchestrator-core.

## Position in Workflow

```
{{WORKFLOW_DIAGRAM}}
```

**Upstream**: {{UPSTREAM}}
**Downstream**: {{DOWNSTREAM}}

## Phase Routing

| Specialist | Route When |
|------------|-----------|
{{ROUTING_TABLE}}

## Handoff Criteria

| Phase | Criteria |
|-------|----------|
{{HANDOFF_CRITERIA}}

{{#if CROSS_TEAM_PROTOCOL}}
## Cross-Team Protocol

{{CROSS_TEAM_PROTOCOL}}
{{/if}}

## Skills Reference

Reference these skills as appropriate:
{{SKILLS_REFERENCE}}

## Team-Specific Anti-Patterns

{{TEAM_ANTIPATTERNS}}

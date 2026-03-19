---
name: evolve
description: "Transformative UI work: migrate design system changes, evolve token architecture, execute phased rollouts. Every intermediate state must be coherent -- the system never breaks during evolution."
argument-hint: "<what to migrate or evolve> [--scope=FEATURE|SYSTEM]"
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, Task, Skill
model: opus
---

## Context

Direct entry point for transformative posture UI work. Posture is fixed to transformative -- skips posture detection. Use when planning a design system migration, token deprecation, or phased rollout.

## Your Task

1. **Parse arguments**: Extract work description from `$ARGUMENTS`. Check for optional `--scope` flag.

2. **Posture is fixed**: Transformative. No detection needed.

3. **Scope detection with floor enforcement** (skip detection if `--scope` provided):
   - FEATURE: Feature area, page, or 3-10 components referenced
   - SYSTEM: Design system, cross-cutting, or token changes referenced
   - If COMPONENT scope detected: redirect to corrective COMPONENT:
     > Transformative work requires cross-component coordination. Minimum scope is FEATURE.
     > Routing to corrective posture for this single-component change.
     Then dispatch as corrective COMPONENT.

4. **Announce routing** to user:
   > Posture: **transformative** (evolve mode).
   > Scope: **{detected scope}**.
   > Workflow: propose -> analyze -> migrate -> validate.
   > The workflow uses a four-phase rollout model. Phase 2 (block new usage) is never skipped.

5. **Dispatch to potnia** via Task tool:
   ```
   Task("potnia", "
   CONSULTATION_REQUEST:
   type: route
   posture: transformative
   scope: {detected or explicit}
   description: {user's work description}
   ")
   ```

## When to Use Instead of /ui

When planning a design system migration, token deprecation, or phased rollout. The explicit invocation signals commitment to the four-phase rollout model (warn -> block new -> budget down -> remove) and contract-aware impact analysis across all five contract types (API, behavior, visual, a11y, automation).

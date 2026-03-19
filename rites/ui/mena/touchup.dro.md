---
name: touchup
description: "Corrective UI work: fix imperfections, remove unnecessary elements, harden edge states. Subtractive-first approach -- what can be removed before what should be added."
argument-hint: "<what to fix or refine> [--scope=COMPONENT|FEATURE|SYSTEM]"
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, Task, Skill
model: opus
---

## Context

Direct entry point for corrective posture UI work. Posture is fixed to corrective -- skips posture detection. Use when you know the work is about fixing, refining, or simplifying something that already exists.

## Your Task

1. **Parse arguments**: Extract work description from `$ARGUMENTS`. Check for optional `--scope` flag.

2. **Posture is fixed**: Corrective. No detection needed.

3. **Scope detection** (skip if `--scope` provided):
   - COMPONENT: Single component named, <200 LOC estimated
   - FEATURE: Feature area, page, or 3-10 components referenced
   - SYSTEM: Design system, cross-cutting, or token changes referenced

4. **Announce routing** to user:
   > Posture: **corrective** (touchup mode).
   > Scope: **{detected scope}**.
   > Workflow: {phase sequence}.

   Phase sequences by scope:
   - COMPONENT: audit -> fix -> validate
   - FEATURE: audit -> fix -> validate
   - SYSTEM: audit -> impact -> fix -> validate

5. **Dispatch to potnia** via Task tool:
   ```
   Task("potnia", "
   CONSULTATION_REQUEST:
   type: route
   posture: corrective
   scope: {detected or explicit}
   description: {user's work description}
   ")
   ```

## When to Use Instead of /ui

When you know the work is corrective -- something is broken, inconsistent, or unnecessarily complex. Skips posture detection overhead and makes intent explicit in session history. The audit phase starts with the subtractive question: "what can I remove?"

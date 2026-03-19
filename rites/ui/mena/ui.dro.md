---
name: ui
description: "UI development lifecycle with posture-aware routing. Detects whether work is corrective (fix/refine), generative (create/explore), or transformative (migrate/evolve) and dispatches to the appropriate workflow shape."
argument-hint: "<description of UI work> [--posture=touchup|compose|evolve] [--scope=COMPONENT|FEATURE|SYSTEM]"
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, Task, Skill
model: opus
---

## Context

Universal entry point for the UI rite. Auto-detects posture (corrective/generative/transformative) and scope (COMPONENT/FEATURE/SYSTEM) from the work description, then dispatches to the appropriate workflow shape via potnia.

## Your Task

1. **Parse arguments**: Extract work description from `$ARGUMENTS`. Check for optional `--posture` and `--scope` flags.

2. **Posture detection** (skip if `--posture` provided or if user invoked `/touchup`, `/compose`, `/evolve`):
   - Corrective signals: fix, broken, wrong, regression, cleanup, remove, simplify, refine, touchup, audit, check
   - Generative signals: build, create, new, prototype, explore, feels like, interaction, compose, design, imagine
   - Transformative signals: migrate, evolve, deprecate, rename, rollout, update system, token change, redesign system
   - Ambiguous: default to **corrective** (smallest blast radius -- can always escalate if audit reveals need for generative work)

3. **Scope detection** (skip if `--scope` provided):
   - COMPONENT: Single component named, <200 LOC estimated
   - FEATURE: Feature area, page, or 3-10 components referenced
   - SYSTEM: Design system, cross-cutting, or token changes referenced

4. **Route validation**:
   - If posture=transformative AND scope=COMPONENT: redirect to corrective COMPONENT.
     Inform user: "Transformative work at COMPONENT scope is corrective work in disguise -- routing to corrective posture."

5. **Announce routing** to user:
   > Detected: **{posture}** posture, **{scope}** scope.
   > Workflow: {phase sequence}.
   > Override with `--posture=compose` or `--scope=SYSTEM` if this is wrong.

   Phase sequences:
   - corrective COMPONENT: audit -> fix -> validate
   - corrective FEATURE: audit -> fix -> validate
   - corrective SYSTEM: audit -> impact -> fix -> validate
   - generative COMPONENT: feel -> harden -> validate
   - generative FEATURE: intent -> feel -> harden -> validate
   - generative SYSTEM: intent -> feel -> harden -> validate
   - transformative FEATURE: propose -> analyze -> migrate -> validate
   - transformative SYSTEM: propose -> analyze -> migrate -> validate

6. **Dispatch to potnia** via Task tool:
   ```
   Task("potnia", "
   CONSULTATION_REQUEST:
   type: route
   posture: {detected or explicit}
   scope: {detected or explicit}
   description: {user's work description}
   ")
   ```

7. Potnia coordinates the appropriate workflow shape per the dispatch table.

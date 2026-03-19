---
name: compose
description: "Generative UI work: build new interactions, explore how something should feel, prototype in the browser. Feel-first/ship-second -- throwaway code discovers the right interaction, then production code delivers it."
argument-hint: "<what to build or explore> [--scope=COMPONENT|FEATURE|SYSTEM]"
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, Task, Skill
model: opus
---

## Context

Direct entry point for generative posture UI work. Posture is fixed to generative -- skips posture detection. Use when you know you need to create something new and want the feel-first prototyping workflow.

## Your Task

1. **Parse arguments**: Extract work description from `$ARGUMENTS`. Check for optional `--scope` flag.

2. **Posture is fixed**: Generative. No detection needed.

3. **Entry guard**: If the work description contains corrective signals (fix, broken, wrong, regression), warn the user before proceeding:
   > This sounds like corrective work. `/touchup` may be more appropriate.
   > Proceeding with generative workflow -- the feel phase will produce throwaway code.
   > The hardened production code is rebuilt from scratch, not refined from the prototype.
   > Continue? [y/N]

4. **Scope detection** (skip if `--scope` provided):
   - COMPONENT: Single component named, <200 LOC estimated
   - FEATURE: Feature area, page, or 3-10 components referenced
   - SYSTEM: Design system, cross-cutting, or token changes referenced

5. **Announce routing** to user:
   > Posture: **generative** (compose mode).
   > Scope: **{detected scope}**.
   > Workflow: {phase sequence}.
   > Note: The feel phase produces throwaway code. Production implementation is rebuilt from scratch in harden.

   Phase sequences by scope:
   - COMPONENT: feel -> harden -> validate
   - FEATURE: intent -> feel -> harden -> validate
   - SYSTEM: intent -> feel -> harden -> validate

6. **Dispatch to potnia** via Task tool:
   ```
   Task("potnia", "
   CONSULTATION_REQUEST:
   type: route
   posture: generative
   scope: {detected or explicit}
   description: {user's work description}
   ")
   ```

## When to Use Instead of /ui

When you know you need to create something new and want the feel-first prototyping workflow. The explicit invocation signals commitment to throwaway code and the two-phase feel/harden model. Invoking `/compose` sets the expectation that a feel prototype will be built before any production code is written.

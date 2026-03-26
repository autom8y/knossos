---
name: ui-ref
description: |
  Reference for the UI rite — 9-agent roster, three-posture workflow, command surface, routing
  matrix, quality gates, and companion knowledge base. Use when: orienting to the ui rite,
  routing work to the correct specialist, understanding posture-based phase flow, looking up
  available commands, checking which workflow shape to use, or loading specialist knowledge
  (tokens, components, accessibility, motion, quality gates, aesthetics, evolution).
  Triggers: ui rite, agents, workflow, potnia, design-system-steward, rendering-architect,
  stylist, component-engineer, a11y-engineer, frontend-fanatic, interaction-prototyper,
  motion-architect, corrective, generative, transformative.
---

# UI Rite Reference

> 9-agent roster, three-posture workflow, command surface, and companion knowledge base (v2.0).

## Agent Roster

| Agent | Role | Type | Postures |
|-------|------|------|----------|
| **potnia** | Coordinates via scope x posture routing; strategic critique at phase transitions | orchestrator | all |
| **design-system-steward** | Defines and evolves design systems: token taxonomy, governance, change proposals, impact analysis, migration | architect | all |
| **rendering-architect** | Per-route rendering strategy, hydration, performance budgets | architect | all |
| **stylist** | CSS architecture, token-to-CSS mapping, layout, responsive, animation implementation | builder | all |
| **component-engineer** | Production component implementation: state management, testing, headless logic separation | engineer | all |
| **a11y-engineer** | WCAG 2.2 AA validation -- terminal gate in every posture, zero tolerance | validator | all |
| **frontend-fanatic** | Subtractive auditing (corrective audit-phase owner), visual auditing, UX evaluation; soft gate on D1/D2 (generative) and visual contract (transformative) | evaluator | all |
| **interaction-prototyper** | Throwaway interaction prototyping in the feel phase -- code as design medium | prototyper | generative only |
| **motion-architect** | Motion classification, interaction physics, animation architecture -- pre-CSS structural decisions | architect | all |

## Two-Dimensional Routing Matrix

Scope x Posture determines the workflow shape:

| | COMPONENT | FEATURE | SYSTEM |
|---|---|---|---|
| **Corrective** | audit -> fix -> validate | audit -> fix -> validate | audit -> impact -> fix -> validate |
| **Generative** | feel -> harden -> validate | intent -> feel -> harden -> validate | intent -> feel -> harden -> validate |
| **Transformative** | *redirect to corrective* | propose -> analyze -> migrate -> validate | propose -> analyze -> migrate -> validate |

**COMPONENT x Transformative**: Always redirects to corrective COMPONENT. Transformative requires cross-component coordination (minimum FEATURE scope).

## Posture Detection Heuristics

| Signals | Posture |
|---------|---------|
| fix, broken, wrong, regression, cleanup, remove, simplify, refine, touchup, audit, check | Corrective |
| build, create, new, prototype, explore, feels like, interaction, compose, design, imagine | Generative |
| migrate, evolve, deprecate, rename, rollout, update system, token change, redesign system | Transformative |
| Ambiguous | **Corrective (default)** -- smallest blast radius |

## Commands

| Command | Type | Posture | Description |
|---------|------|---------|-------------|
| `/ui` | Orchestrated | Auto-detected | Universal entry: posture and scope auto-detected |
| `/touchup` | Orchestrated | Corrective | Fix imperfections, remove unnecessary elements |
| `/compose` | Orchestrated | Generative | Build new interactions with feel-first prototyping |
| `/evolve` | Orchestrated | Transformative | Design system migration with phased rollout |
| `/component-audit` | Utility | Corrective (partial) | Design system compliance check (audit only) |
| `/a11y-check` | Utility | n/a | Standalone WCAG 2.2 AA validation |
| `/perf-budget` | Utility | n/a | Standalone performance budget compliance check |
| `/motion-audit` | Utility | n/a | Standalone motion architecture assessment |

## Quality Gate Summary

| Gate | Dimension | Type | Posture |
|------|-----------|------|---------|
| D0: Accessibility | WCAG 2.2 AA | Hard (always) | All |
| D1: Micro-interactions | Interaction quality | Soft (blocking) | Generative FEATURE/SYSTEM |
| D2: Cognitive efficiency | User flow directness | Soft (blocking) | Generative FEATURE/SYSTEM |
| Visual contract (QG-E4) | Visual regression | Soft (blocking) | Transformative |
| D3: Consistency | Pattern reuse | Embedded self-check | All |
| D5: Edge state craft | State completeness | Embedded self-check | All |
| D4, D6: Brand/Emotion | Subjective quality | Advisory (never blocking) | All |

## Back-Routes

| Trigger | From | To | Auto? |
|---------|------|----|-------|
| A11y CSS violations | validate | fix (stylist) | Yes |
| A11y component violations | validate | fix (component-engineer) | Yes |
| D1/D2 soft gate fail | validate | harden (component-engineer) | Yes |
| Visual contract regression | validate | migrate (component-engineer) | Yes |
| New dependencies found | migrate | analyze | Yes |
| Fix scope expands | fix | audit | User confirm |
| Feel cannot be hardened | harden | feel | User confirm |
| Classification wrong | feel | intent | User confirm |
| Proposal infeasible | analyze | propose | User confirm |

## Companion Reference

| Domain | File | When to Load |
|--------|------|--------------|
| **Design Systems** | | |
| Token taxonomy | [token-taxonomy.md](token-taxonomy.md) | DTCG three-tier hierarchy, OKLCH, CSS custom properties, Style Dictionary pipeline |
| Component taxonomy | [component-taxonomy.md](component-taxonomy.md) | Primitives/Composites/Patterns, headless separation, slots vs props, lifecycle status |
| Governance | [governance.md](governance.md) | Token pipeline, governance gates, breaking change classification, RFC lifecycle |
| Evolution rollout | [four-phase-rollout.md](four-phase-rollout.md) | Warn/block/budget-down/remove playbook; transformative posture only |
| **Quality & Accessibility** | | |
| WCAG checklist | [wcag-checklist.md](wcag-checklist.md) | WCAG 2.2 AA criteria, CSS-a11y architecture, focus management, forced-colors |
| Testing pyramid | [testing-pyramid.md](testing-pyramid.md) | Test layers, decision framework, structured output formats |
| Performance budgets | [performance-budgets.md](performance-budgets.md) | Core Web Vitals thresholds, JS budgets, CLS prevention, rendering impact |
| Corrective gates | [corrective-gates.md](corrective-gates.md) | QG-T1..T5 definitions; corrective posture validate phase |
| Generative gates | [generative-gates.md](generative-gates.md) | QG-C1..C7 definitions; D1/D2 soft gate mechanics |
| Transformative gates | [transformative-gates.md](transformative-gates.md) | QG-E1..E9 definitions; five-contract verification, visual contract soft gate |
| **Motion** | | |
| Frequency x novelty matrix | [frequency-novelty-matrix.md](frequency-novelty-matrix.md) | Tier 1-4 animation budget, 10% novelty budget, classification worksheet |
| Progressive craft layers | [progressive-craft-layers.md](progressive-craft-layers.md) | L0-L4 harden-phase build ordering |
| **Aesthetics** | | |
| VisAWI framework | [visawi-framework.md](visawi-framework.md) | Four-facet structured audit: simplicity, diversity, colorfulness, craftsmanship |
| Fluency principles | [fluency-principles.md](fluency-principles.md) | Processing fluency theory, 50ms first-impression protocol, checklist |
| Emotional design | [emotional-design.md](emotional-design.md) | Norman's three levels, aesthetic-usability effect, cultural calibration |

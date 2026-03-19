---
name: ui-ref
description: "UI rite self-reference: 9-agent roster, three-posture workflow, command surface, routing matrix, and quick-start guide. Use when: orienting to the ui rite, routing work to the correct specialist, understanding posture-based phase flow, looking up available commands, checking which workflow shape to use. Triggers: ui rite, agents, workflow, potnia, design-system-steward, rendering-architect, stylist, component-engineer, a11y-engineer, frontend-fanatic, interaction-prototyper, motion-architect, corrective, generative, transformative."
---

# UI Rite Reference

> 9-agent roster, three-posture workflow, command surface, and quick-start guide for the ui rite (v2.0).

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
| `/component-audit` | Utility | Corrective (partial) | Design system compliance check (audit only, no fix/validate) |
| `/a11y-check` | Utility | n/a | Standalone WCAG 2.2 AA validation |
| `/perf-budget` | Utility | n/a | Standalone performance budget compliance check |
| `/motion-audit` | Utility | n/a | Standalone motion architecture assessment |

## Phase-Agent Assignments

### Corrective Posture

| Phase | Owner | Participants |
|-------|-------|-------------|
| audit | frontend-fanatic | motion-architect (FEATURE/SYSTEM), rendering-architect (FEATURE/SYSTEM) |
| impact | design-system-steward | rendering-architect (SYSTEM only) |
| fix | component-engineer / stylist | rendering-architect (FEATURE/SYSTEM performance fixes) |
| validate | a11y-engineer | frontend-fanatic (advisory, discretionary) |

### Generative Posture

| Phase | Owner | Participants |
|-------|-------|-------------|
| intent | motion-architect | potnia (strategic critique), design-system-steward (SYSTEM) |
| feel | interaction-prototyper | (none -- throwaway phase, no quality gates) |
| harden | component-engineer | stylist, rendering-architect (FEATURE/SYSTEM), design-system-steward (SYSTEM) |
| validate | a11y-engineer (first) | frontend-fanatic (automatic FEATURE/SYSTEM: D1/D2 soft gates) |

### Transformative Posture

| Phase | Owner | Participants |
|-------|-------|-------------|
| propose | design-system-steward | (potnia strategic critique at propose->analyze) |
| analyze | design-system-steward | rendering-architect, motion-architect |
| migrate | component-engineer | stylist, rendering-architect, design-system-steward (oversight) |
| validate | a11y-engineer (first) | frontend-fanatic (visual contract soft gate QG-E4) |

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
| Contract violations | validate | migrate | Yes |
| New dependencies found | migrate | analyze | Yes |
| Fix scope expands | fix | audit | User confirm |
| Feel cannot be hardened | harden | feel | User confirm |
| Classification wrong | feel | intent | User confirm |
| Proposal infeasible | analyze | propose | User confirm |

## Quick Start

### Fix something broken (Corrective)
```
/touchup "fix the focus state on the dropdown nav"
-- or --
/ui "fix the focus state on the dropdown nav"
```

### Build something new (Generative)
```
/compose "build a command palette with keyboard navigation"
-- or --
/ui "build a command palette with keyboard navigation"
```

### Migrate a design system change (Transformative)
```
/evolve "migrate from HSL to Oklch color tokens"
-- or --
/ui "migrate from HSL to Oklch color tokens"
```

### Targeted utilities
```
/a11y-check                  # WCAG validation only
/perf-budget                 # Performance budget check
/component-audit             # Design system compliance check
/motion-audit "CommandPalette"  # Motion architecture assessment
```

## Artifacts Produced

| Agent | Artifact | Location |
|-------|---------|---------|
| design-system-steward | Design System Spec (DSS-{slug}.md) | `.ledge/specs/` |
| design-system-steward | Change Proposal (CP-{slug}.md) | `.ledge/specs/` |
| design-system-steward | Impact Analysis (IA-{slug}.md) | `.ledge/specs/` |
| rendering-architect | Rendering Manifest (RM-{slug}.md) | `.ledge/specs/` |
| stylist | Style Architecture (SA-{slug}.md) | `.ledge/specs/` |
| motion-architect | Motion Architecture Spec (MOTION-{slug}.md) | `.ledge/specs/` |
| interaction-prototyper | Feel Prototype Assessment (FEEL-{slug}.md) | `.ledge/reviews/` |
| component-engineer | Component implementation | In codebase |
| a11y-engineer | Accessibility Report (A11Y-{slug}.md) | `.ledge/reviews/` |
| frontend-fanatic | Audit Report (AUDIT-{slug}.md) | `.ledge/reviews/` |

## Skills Available to Agents

| Skill | Contents | Consuming Agents |
|-------|---------|-----------------|
| `ui-design-systems` | Token taxonomy, component classification, governance | design-system-steward, stylist |
| `ui-quality` | WCAG checklist, testing pyramid, performance budgets | a11y-engineer, stylist |
| `ui-architecture` | State patterns, rendering strategies, CSS principles | rendering-architect, component-engineer, stylist |
| `aesthetic-evaluation` | VisAWI, fluency principles, emotional design | frontend-fanatic |
| `motion-architecture` | Frequency x novelty matrix, progressive craft layers (L0-L4) | motion-architect, component-engineer |
| `evolution-lifecycle` | Four-phase rollout model (transformative posture only) | design-system-steward |
| `quality-gates` | Per-posture gate criteria (QG-T, QG-C, QG-E) | frontend-fanatic |
| `orchestrator-templates` | CONSULTATION_RESPONSE format | potnia |
| `cross-rite-handoff` | Cross-rite routing patterns | potnia |

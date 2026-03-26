---
last_verified: 2026-03-26
---

# Rite: ui

> UI/UX development lifecycle with posture-aware routing.

The ui rite is the largest rite by agent count (9 agents). It organizes UI work along two dimensions: **scope** (COMPONENT / FEATURE / SYSTEM) determines agent count and depth; **posture** (corrective / generative / transformative) determines workflow shape. Potnia detects posture from request signals and dispatches to the appropriate specialist sequence.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | ui |
| **Version** | 2.0.0 |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 9 |
| **Entry Agent** | potnia |
| **MCP Pool** | browser-local |

---

## When to Use

- Building new UI components or feature pages (generative posture)
- Fixing or improving existing UI (corrective posture)
- Migrating design systems or making cross-cutting UI changes (transformative posture)
- Accessibility validation against WCAG 2.2 AA
- Performance budget enforcement (Core Web Vitals)
- Design system creation or evolution

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates UI development via scope × posture routing; strategic critique at phase transitions |
| **design-system-steward** | Defines and evolves design systems — token taxonomy, component architecture, governance, migration planning |
| **rendering-architect** | Determines per-route rendering strategy (SSG/ISR/SSR/CSR), hydration approach, and performance budgets |
| **stylist** | Translates design tokens and motion specs into CSS architecture, layout, responsive design, and theming |
| **component-engineer** | Implements production components with state management, headless logic separation, and testing |
| **a11y-engineer** | Validates WCAG 2.2 AA compliance across four testing layers; zero-tolerance gate for accessibility |
| **frontend-fanatic** | Browser-first subtractive auditing and UX evaluation; audit-phase owner in corrective posture |
| **interaction-prototyper** | Throwaway interaction prototyping in the feel phase — code as design medium, browser as canvas |
| **motion-architect** | Motion classification, interaction physics, and animation architecture — pre-CSS structural decisions |

See agent files: `rites/ui/agents/`

---

## Workflow Shapes

The ui rite does not follow a single fixed phase sequence. Workflow shape depends on posture:

### Corrective Posture (fix / improve existing UI)
```
audit → fix → validate
```
- **audit**: frontend-fanatic produces subtractive audit report; motion-architect assesses animation appropriateness
- **fix**: component-engineer + stylist apply fixes; rendering-architect reviews if rendering strategy is affected
- **validate**: a11y-engineer gates accessibility; frontend-fanatic soft-gates interaction quality (FEATURE/SYSTEM)

### Generative Posture (build new UI)
```
intent → feel → harden → validate
```
- **intent**: motion-architect classifies the interaction; potnia provides strategic critique
- **feel**: interaction-prototyper builds throwaway prototype in browser
- **harden**: component-engineer + stylist + rendering-architect build production implementation
- **validate**: a11y-engineer gates; frontend-fanatic soft-gates D1/D2 quality

### Transformative Posture (migrate / redesign)
```
propose → analyze → migrate → validate
```
- **propose**: design-system-steward produces change proposal with dependency graph
- **analyze**: design-system-steward analyzes impact across five contract types
- **migrate**: four-phase rollout execution
- **validate**: a11y-engineer gates; frontend-fanatic evaluates visual contract regression

### Complexity Levels

| Level | Scope |
|-------|-------|
| **COMPONENT** | Single component, < 200 LOC |
| **FEATURE** | Feature area or page, 3–10 components |
| **SYSTEM** | Design system overhaul, cross-cutting changes |

---

## Invocation Patterns

```bash
# Quick switch to UI rite
/ui

# Corrective: audit a component
/component-audit

# Check accessibility
/a11y-check

# Review performance budget
/perf-budget

# Audit motion
/motion-audit

# Compose components (generative)
/compose

# Evolve design system (transformative)
/evolve

# Touch up existing UI (corrective shorthand)
/touchup
```

---

## Skills

- `ui-ref` — Workflow reference
- `ui-design-systems` — Design system patterns and token taxonomy
- `ui-quality` — Quality gates and testing patterns
- `ui-architecture` — Rendering strategy and performance budgets
- `aesthetic-evaluation` — Visual and UX evaluation criteria
- `motion-architecture` — Motion classification and interaction physics
- `evolution-lifecycle` — Design system evolution and migration
- `quality-gates` — Gate criteria across all workflow phases

---

## Source

**Manifest**: `rites/ui/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)
- [CLI: sync](../operations/cli-reference/cli-sync.md)
- [Rite Catalog](index.md)

---
name: ui-ref
description: "UI rite self-reference: agent roster, workflow phases, commands, and quick-start guide. Use when: orienting to the ui rite, routing work to the correct specialist, understanding phase flow, looking up available commands. Triggers: ui rite, agents, workflow, potnia, design-system-architect, rendering-architect, stylist, component-engineer, a11y-engineer."
---

# UI Rite Reference

> Agent roster, workflow phases, commands, and quick-start guide for the ui rite.

## Agent Roster

| Agent | Role | Entry Point |
|-------|------|-------------|
| **potnia** | Coordinates UI development phases; routes work to specialists based on complexity and work type | All work via Task tool |
| **design-system-architect** | Defines token taxonomy (DTCG three-tier), component architecture, and governance pipeline | New design system, design system overhaul |
| **rendering-architect** | Determines per-route rendering strategy, hydration patterns, and performance budgets | New feature/page, performance work |
| **stylist** | Translates design tokens into CSS architecture, layout patterns, and visual implementation | Styling overhaul, CSS architecture |
| **component-engineer** | Implements components with state management, testing, and structured output | Component modification, component addition |
| **a11y-engineer** | Validates WCAG 2.2 AA compliance across four testing layers and gates accessibility quality | A11y remediation, accessibility validation |

## Workflow Phases

```
Foundation → Strategy → Styling → Implementation → Validation
    |             |          |            |               |
design-system- rendering- stylist   component-   accessibility-
 architect     architect            engineer      engineer
```

Phases are conditional on complexity level:

| Complexity | Scope | Phases Active |
|-----------|-------|--------------|
| **TASK** | Single component, <200 LOC | Implementation, Validation |
| **MODULE** | Feature area or page, 3–10 components | Strategy, Styling, Implementation, Validation |
| **SYSTEM** | Design system overhaul, cross-cutting changes | All five phases |

## Entry Points by Work Type

| Work Type | Entry Agent |
|-----------|------------|
| New design system | design-system-architect |
| Design system overhaul | design-system-architect |
| New feature | rendering-architect |
| New page | rendering-architect |
| Styling overhaul | stylist |
| Component modification | component-engineer |
| Component addition | component-engineer |
| A11y remediation | a11y-engineer |

## Back-Routes (Automatic)

| Trigger | From | To |
|---------|------|----|
| A11y violations in CSS (contrast, focus styles) | a11y-engineer | stylist |
| A11y violations requiring component changes (ARIA, keyboard) | a11y-engineer | component-engineer |
| Implementation exceeds JS budget | component-engineer | rendering-architect |
| CSS performance reveals CLS or rendering cost issues | stylist | rendering-architect |
| Missing tokens prevent CSS mapping | stylist | design-system-architect (user confirmation required) |
| Missing tokens prevent rendering decisions | rendering-architect | design-system-architect (user confirmation required) |

## Available Commands

| Command | Description |
|---------|-------------|
| `/ui` | Full UI development lifecycle (complexity-routed via potnia) |
| `/component-audit` | Audit a component against design system standards |
| `/a11y-check` | Run accessibility validation check |
| `/perf-budget` | Check performance budget compliance |

## Quick Start

**For a new component** (TASK complexity):
1. Invoke potnia via `/ui` or Task tool
2. potnia routes to `component-engineer` for implementation
3. `a11y-engineer` validates WCAG 2.2 AA
4. Back-routes fire automatically if violations found

**For a new feature or page** (MODULE complexity):
1. Invoke potnia — routes to `rendering-architect` first
2. `rendering-architect` produces rendering manifest (per-route strategy, hydration, budgets)
3. `stylist` produces CSS architecture and token mapping
4. `component-engineer` implements with state management and tests
5. `a11y-engineer` validates

**For a new design system** (SYSTEM complexity):
1. `design-system-architect` produces design-system-spec (token taxonomy, component classification, governance pipeline)
2. Continue through all five phases

## Artifacts Produced

| Agent | Artifact | Location |
|-------|---------|---------|
| design-system-architect | Design System Spec (DSS-{slug}.md) | `.ledge/specs/` |
| rendering-architect | Rendering Manifest (RM-{slug}.md) | `.ledge/specs/` |
| stylist | Style Architecture (SA-{slug}.md) | `.ledge/specs/` |
| component-engineer | Component implementation | In codebase |
| a11y-engineer | Accessibility Report (A11Y-{slug}.md) | `.ledge/reviews/` |

## Skills Available to Agents

| Skill | Contents |
|-------|---------|
| `ui-design-systems` | Token taxonomy, component classification, governance |
| `ui-quality` | WCAG checklist, testing pyramid, performance budgets |
| `ui-architecture` | State patterns, rendering strategies, CSS principles |
| `ui-ref` | This reference |

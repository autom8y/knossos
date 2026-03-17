---
name: ui-quality
description: "UI quality standards: WCAG 2.2 AA requirements, CSS-accessibility architecture (focus management, forced-colors, motion, semantic CSS traps), four-layer accessibility testing, UI testing pyramid strategy, performance budgets (Core Web Vitals, JS budgets), and CLS prevention. Use when: implementing accessibility, writing tests, evaluating performance, setting quality gates, debugging focus styles, handling forced-colors mode, choosing animation architecture. Triggers: accessibility, WCAG, a11y, testing, performance, Core Web Vitals, CWV, LCP, INP, CLS, testing pyramid, axe, focus, forced-colors, reduced-motion, outline, box-shadow focus."
---

# UI Quality

> WCAG 2.2 AA compliance, layered accessibility testing, testing strategy, and performance budgets.

## Overview

WCAG 2.2 AA is the global legal standard referenced by US ADA litigation (4,000+ lawsuits/year) and the European Accessibility Act (enforced June 2025, fines up to 4% of revenue). Automated tools catch approximately 57% of accessibility issues — the remaining 43% require correct authoring patterns. Core Web Vitals thresholds (LCP < 2.5s, INP < 200ms, CLS < 0.1) are minimum acceptable performance, not optimization goals.

## Contents

| File | Purpose |
|------|---------|
| [wcag-checklist.md](wcag-checklist.md) | WCAG 2.2 AA criteria, semantic HTML, ARIA guidance, CSS-accessibility architecture |
| [testing-pyramid.md](testing-pyramid.md) | UI testing layers, decision frameworks, structured output formats |
| [performance-budgets.md](performance-budgets.md) | Core Web Vitals thresholds, JS budgets, rendering strategy performance impact |

## When to Use

**wcag-checklist.md** — When generating any interactive element, form, navigation, or color; when implementing ARIA; when building custom widgets; when checking legal compliance; when implementing focus styles; when handling forced-colors or reduced-motion; when CSS might affect the accessibility tree.

**testing-pyramid.md** — When deciding what tests to write for a component; when setting up CI quality gates; when evaluating test suite health; when configuring structured test output for agent workflows.

**performance-budgets.md** — When choosing a rendering strategy; when adding dependencies; when diagnosing CWV failures; when evaluating bundle size impact.

---
name: ui-architecture
description: "UI architectural patterns: state classification and management (server/client/URL/derived), SWR caching strategy, optimistic UI rules, signals reactivity, rendering strategy decision matrix (SSG/ISR/SSR/islands/resumability), hydration spectrum, methodology-neutral CSS architecture, cascade theory, and component isolation strategies. Use when: making state management decisions, choosing rendering strategy, designing CSS architecture, evaluating caching patterns, debugging cascade conflicts, choosing component isolation approach. Triggers: state management, server state, client state, URL state, rendering strategy, SSG, ISR, SSR, hydration, islands, CSS architecture, signals, cascade, specificity, @layer, @scope, component isolation, BEM, CSS Modules."
---

# UI Architecture

> State management patterns, rendering strategy decision matrix, CSS architecture principles, cascade theory, and component isolation strategies.

## Overview

Three architectural domains govern UI system design. State management separates server, client, URL, and derived state — mixing these categories causes most state complexity. Rendering strategy is a per-route decision: the default is SSG, escalating only when freshness demands it. CSS architecture is methodology-neutral: the principles (cascade control, token-driven values, logical properties) apply regardless of framework or methodology.

## Contents

| File | Purpose |
|------|---------|
| [state-patterns.md](state-patterns.md) | State classification, SWR default, optimistic UI rules, signals, state machines |
| [rendering-strategies.md](rendering-strategies.md) | SSG/ISR/SSR/islands/resumability decision framework, hydration spectrum |
| [css-principles.md](css-principles.md) | Methodology-neutral CSS architecture, logical properties, cascade control, cascade fundamentals |
| [cascade-theory.md](cascade-theory.md) | Deep cascade reference: 7-criteria algorithm, @layer semantics, @scope proximity, @property |
| [component-isolation.md](component-isolation.md) | Decision framework for BEM, CSS Modules, @scope, Shadow DOM, CSS-in-JS |

## When to Use

**state-patterns.md** — When writing any stateful code; when choosing between component state, global store, URL state, or a data-fetching layer; when implementing mutations (optimistic vs pessimistic); when deciding between signals and component state.

**rendering-strategies.md** — When creating a new route; when diagnosing performance issues; when choosing hydration strategy; when evaluating bundle size vs freshness tradeoffs.

**css-principles.md** — When designing CSS architecture; when implementing token-to-CSS mapping; when implementing responsive layouts; when writing animations; when debugging cascade or specificity conflicts.

**cascade-theory.md** — When you need the "why" behind cascade behavior; when debugging @layer priority inversion with !important; when implementing @scope boundaries; when registering custom properties with @property.

**component-isolation.md** — When choosing a component isolation strategy for a new project; when evaluating BEM vs CSS Modules vs @scope; when integrating multiple CSS methodologies in one codebase.

---
name: ui-architecture
description: |
  UI architectural patterns: state classification and management (server/client/URL/derived),
  SWR caching strategy, optimistic UI rules, signals reactivity, rendering strategy decision
  matrix (SSG/ISR/SSR/islands/resumability), hydration spectrum, methodology-neutral CSS
  architecture, cascade theory, and component isolation strategies. Use when: making state
  management decisions, choosing rendering strategy, designing CSS architecture, evaluating
  caching patterns, debugging cascade conflicts, choosing component isolation approach.
  Triggers: state management, server state, client state, URL state, rendering strategy,
  SSG, ISR, SSR, hydration, islands, CSS architecture, signals, cascade, specificity,
  @layer, @scope, component isolation, BEM, CSS Modules.
---

# UI Architecture

> State management patterns, rendering strategy decision matrix, CSS architecture principles, cascade theory, and component isolation strategies.

## Overview

Three architectural domains govern UI system design:

**State management** separates server, client, URL, and derived state. Mixing these categories causes most state complexity. The default is SWR (stale-while-revalidate) for server state — optimistic UI applies only when rollback cost is low and latency is high.

**Rendering strategy** is a per-route decision. The default is SSG; escalate only when freshness demands it. SSG > ISR > Streaming SSR > Full SSR > CSR. Every step up the ladder costs performance and complexity.

**CSS architecture** is methodology-neutral. The principles — cascade control, token-driven values, logical properties — apply regardless of framework (BEM, CSS Modules, @scope, Shadow DOM). Choose the isolation strategy that matches team familiarity and framework constraints; apply the same principles in any methodology.

## Companion Files

| File | Purpose |
|------|---------|
| [state-patterns.md](state-patterns.md) | State classification, SWR default, optimistic UI rules, signals, state machines |
| [rendering-strategies.md](rendering-strategies.md) | SSG/ISR/SSR/islands/resumability decision framework, hydration spectrum |
| [css-principles.md](css-principles.md) | Methodology-neutral CSS architecture, logical properties, cascade control |
| [cascade-theory.md](cascade-theory.md) | Deep cascade reference: 7-criteria algorithm, @layer semantics, @scope proximity, @property |
| [component-isolation.md](component-isolation.md) | Decision framework for BEM, CSS Modules, @scope, Shadow DOM, CSS-in-JS |

## When to Use Each File

**state-patterns.md** — When writing any stateful code; when choosing between component state, global store, URL state, or a data-fetching layer; when implementing mutations (optimistic vs pessimistic); when deciding between signals and component state.

**rendering-strategies.md** — When creating a new route; when diagnosing performance issues; when choosing hydration strategy; when evaluating bundle size vs freshness tradeoffs.

**css-principles.md** — When designing CSS architecture; when implementing token-to-CSS mapping; when implementing responsive layouts; when writing animations; when debugging cascade or specificity conflicts.

**cascade-theory.md** — When you need the "why" behind cascade behavior; when debugging @layer priority inversion with !important; when implementing @scope boundaries; when registering custom properties with @property.

**component-isolation.md** — When choosing a component isolation strategy for a new project; when evaluating BEM vs CSS Modules vs @scope; when integrating multiple CSS methodologies in one codebase.

## Key Decision Rules

**State category first**: Before writing a state variable, classify it. Server state (async, cacheable) belongs in a data-fetching layer. URL state (bookmarkable, shareable) belongs in the URL. Component state (ephemeral, UI-local) belongs in the component. Global store is the last resort -- not the default.

**CSS @layer is the cascade control primitive**: Use `@layer` to explicitly order rule precedence and eliminate specificity wars. Layers cannot be inverted by specificity -- only by !important, which reverses layer order. Establish layer order once at the entry point; add rules to layers everywhere else.

**Component isolation matches team context**: Shadow DOM is the strongest isolation but requires the most adaptation. @scope is native CSS isolation without custom elements overhead. CSS Modules are the pragmatic choice for component-framework codebases. BEM works in any context when enforced as a convention.

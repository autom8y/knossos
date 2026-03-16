---
name: ui-architecture
description: "UI architectural patterns: state classification and management (server/client/URL/derived), SWR caching strategy, optimistic UI rules, signals reactivity, rendering strategy decision matrix (SSG/ISR/SSR/islands/resumability), hydration spectrum, and methodology-neutral CSS architecture. Use when: making state management decisions, choosing rendering strategy, designing CSS architecture, evaluating caching patterns. Triggers: state management, server state, client state, URL state, rendering strategy, SSG, ISR, SSR, hydration, islands, CSS architecture, signals."
---

# UI Architecture

> State management patterns, rendering strategy decision matrix, and CSS architecture principles.

## Overview

Three architectural domains govern UI system design. State management separates server, client, URL, and derived state — mixing these categories causes most state complexity. Rendering strategy is a per-route decision: the default is SSG, escalating only when freshness demands it. CSS architecture is methodology-neutral: the principles (cascade control, token-driven values, logical properties) apply regardless of framework or methodology.

## Contents

| File | Purpose |
|------|---------|
| [state-patterns.md](state-patterns.md) | State classification, SWR default, optimistic UI rules, signals, state machines |
| [rendering-strategies.md](rendering-strategies.md) | SSG/ISR/SSR/islands/resumability decision framework, hydration spectrum |
| [css-principles.md](css-principles.md) | Methodology-neutral CSS architecture, logical properties, cascade control |

## When to Use

**state-patterns.md** — When writing any stateful code; when choosing between component state, global store, URL state, or a data-fetching layer; when implementing mutations (optimistic vs pessimistic); when deciding between signals and component state.

**rendering-strategies.md** — When creating a new route; when diagnosing performance issues; when choosing hydration strategy; when evaluating bundle size vs freshness tradeoffs.

**css-principles.md** — When designing CSS architecture; when implementing token-to-CSS mapping; when implementing responsive layouts; when writing animations; when evaluating methodology choices.

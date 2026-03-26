---
description: "Component Isolation companion for ui-architecture skill."
---

# Component Isolation

> Decision framework for choosing component style isolation strategies. When to use BEM, CSS Modules, @scope, Shadow DOM, or CSS-in-JS.

## Decision Matrix

| Approach | Runtime Cost | Tooling Required | Native CSS | Specificity Impact | Team Scale | Learning Curve |
|----------|-------------|-----------------|------------|-------------------|------------|---------------|
| **BEM** | None | Lint optional | Yes | Flat (0,1,0) | 3-15 | Low |
| **CSS Modules** | None (build-time) | Webpack/Vite | No | Hashed (0,1,0) | 15-50+ | Low |
| **@scope** | None | None | Yes | Controlled by scope | Any | Medium |
| **Shadow DOM** | Low | None | Yes | Full isolation | Any | High |
| **Runtime CSS-in-JS** | High (48% slower) | JS bundler | No | Generated (0,1,0) | Any | Medium |
| **Zero-Runtime CSS-in-JS** | None (build-time) | JS bundler | No | Generated (0,1,0) | Any | Medium |

Agent rule: default to CSS Modules for build-tooled projects. Consider @scope for new projects with 95%+ browser target. Use Shadow DOM only when true style isolation (blocking inheritance) is required, not just selector scoping.

## The Trajectory

Three generations of the same problem solved with decreasing human overhead:

**Manual discipline (2008-2015)**: BEM, OOCSS, ITCSS. The developer prevents collisions through naming conventions. Works when discipline holds. Breaks when it doesn't.

**Build-time automation (2015-2022)**: CSS Modules, CSS-in-JS. The build tool prevents collisions through class name hashing or runtime injection. Works regardless of discipline. Requires toolchain dependency.

**Native browser features (2022+)**: @layer, @scope, :has(), @container. The browser prevents collisions through specification-level cascade control. No tooling dependency. No runtime cost.

Each generation absorbs the insight of the previous one. @layer now enforces what ITCSS prescribed. @scope now automates what BEM naming provided. The direction is irreversible -- but older approaches remain valid where their constraints are acceptable.

## When to Use Each Approach

**BEM**: Existing codebases where it is already established. No-build-step projects. Teams under 15 where convention enforcement is feasible. BEM's value is readability and grep-ability, not isolation strength.

**CSS Modules**: The pragmatic default for any project with a build pipeline (React, Vue, Angular). Zero runtime cost. Automated scoping. Works with any CSS feature including @layer, @scope, and container queries inside module files.

**@scope**: New projects targeting modern browsers (95%+ coverage reached December 2025). Provides DOM-proximity-based specificity resolution without any naming convention. Best combined with @layer for cross-layer ordering.

**Shadow DOM**: Web components requiring true encapsulation -- inherited properties are blocked, external styles cannot penetrate (except via `::part()` and CSS custom properties). Heavier than needed for most application-level component isolation.

**Zero-Runtime CSS-in-JS** (vanilla-extract, Panda CSS): Teams that want TypeScript-integrated styling with type-safe theme contracts. Build-time extraction means no runtime penalty. Vendor lock-in to the chosen library's API.

**Runtime CSS-in-JS** (styled-components, Emotion): Declining. The 48% performance penalty vs. CSS Modules is well-documented. Justified only in codebases already committed to it where migration cost exceeds performance cost.

## OOCSS: The Two Rules That Changed Everything

Nicole Sullivan's 2008 principles were absorbed into every methodology that followed:

1. **Separate structure from skin**: Repeating visual features (borders, shadows, gradients) are abstracted into reusable "skin" classes. A `.btn` provides structure; `.btn-primary` provides skin. This directly prefigures component base + variant patterns.

2. **Separate container from content**: Styles never depend on DOM hierarchy. Instead of `#sidebar h3`, create `.heading-secondary` applicable anywhere. This is the intellectual foundation of component portability.

These rules are no longer a "methodology" -- they are baseline assumptions in modern CSS architecture.

## ITCSS: The Specificity Graph

Harry Roberts' Inverted Triangle CSS organized styles by three axes: specificity (low to high), reach (broad to narrow), and explicitness (generic to explicit). The seven layers:

| Layer | Contains | Specificity |
|-------|----------|-------------|
| Settings | Design tokens, preprocessor variables | N/A (no output) |
| Tools | Mixins, functions | N/A (no output) |
| Generic | Resets, normalize, box-sizing | Very low |
| Elements | Bare HTML element styles | Low |
| Objects | Layout patterns (OOCSS-style) | Medium |
| Components | UI components | Medium |
| Utilities | Overrides, helpers | High |

The specificity graph concept: plot specificity against source order. A well-organized stylesheet trends monotonically upward. A spiky graph indicates maintainability problems.

`@layer` now enforces this natively. `@layer reset, base, components, utilities` is ITCSS's inverted triangle expressed as a browser-enforced constraint rather than a convention.

Agent rule: when establishing a new project's @layer order, use ITCSS's proven progression as the starting point. The layering insight is validated by a decade of production use.

## CUBE CSS: Composition Over Inheritance

Andy Bell's methodology inverts BEM's model. Instead of blocks carrying most styling responsibility, composition and utilities handle ~80% of styling. Blocks handle the remaining ~20% of component-specific rules.

The four layers:
- **Composition**: Macro layout -- flow, rhythm, spacing between elements
- **Utility**: Single-purpose classes mapped to design tokens
- **Block**: Component-specific styles (deliberately minimal)
- **Exception**: State deviations via `data-*` attributes, not modifier classes

Key insight: by the time you reach the Block layer, "most of your styling has actually been done" through global styles, composition, and utilities. Exceptions use `data-*` attributes because they provide shared semantics for both CSS and JavaScript.

CUBE CSS requires stronger CSS expertise than BEM (it embraces the cascade rather than avoiding it) but produces smaller, more maintainable stylesheets for teams with that expertise.

## Anti-Pattern: Methodology Mixing Without Layer Isolation

Different methodologies in the same codebase create specificity conflicts. BEM's flat `(0,1,0)` selectors collide with utility-first's `(0,1,0)` selectors, and source order becomes the only resolution mechanism -- fragile and hard to reason about.

The fix: wrap each methodology in its own `@layer`:

```css
@layer base, components, utilities;

@layer components {
  /* BEM-style component selectors */
  .card__title { font-size: 1.25rem; }
}
@layer utilities {
  /* Utility-first overrides */
  .text-lg { font-size: 1.125rem; }
}
```

Layer ordering makes the override relationship explicit regardless of source order or specificity within each methodology.

Agent rule: never mix isolation methodologies in a single codebase without @layer boundaries between them. The specificity collision surface is unbounded without explicit layer ordering.

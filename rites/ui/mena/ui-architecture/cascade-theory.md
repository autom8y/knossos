---
description: "Cascade Theory companion for ui-architecture skill."
---

# Cascade Theory

> Deep reference for the CSS cascade algorithm, layer semantics, scope proximity, and registered custom properties. The "why" companion to css-principles.md.

## The Cascade as a Deterministic Algorithm

The cascade is a total ordering over competing declarations. For any element and property, it produces exactly one winner. Document order is the final tiebreaker, guaranteeing determinism. The cascade is not a heuristic -- it is a specification-defined conflict resolution algorithm.

## The Seven Criteria in Detail

### Criterion 1: Origin and Importance

Three origins: User-Agent (browser defaults), User (user preferences), Author (developer stylesheets). Each has a normal and important variant:

| Rank | Origin | Direction |
|------|--------|-----------|
| 1 | Transition declarations | (highest) |
| 2 | Important UA declarations | Important: UA > User > Author |
| 3 | Important user declarations | (inversion of normal order) |
| 4 | Important author declarations | |
| 5 | Animation declarations | |
| 6 | Normal author declarations | Normal: Author > User > UA |
| 7 | Normal user declarations | (standard priority) |
| 8 | Normal UA declarations | (lowest) |

The inversion principle: `!important` reverses origin priority. UA `!important` always beats author `!important` -- preserving user accessibility overrides by design.

### Criteria 2-4: Context, Element-Attached Styles, Layers

- **Context** (criterion 2): Shadow DOM encapsulation boundaries. Shadow styles win inside the shadow tree (exceptions: `::part()`, inherited properties).
- **Element-Attached** (criterion 3): Inline `style` attributes beat all rule-based declarations within the same origin. This is NOT specificity -- it is a separate cascade criterion.
- **Layers** (criterion 4): `@layer` introduces explicit named layers within an origin, ordered by first declaration.

## @layer Ordering Semantics

### Declaration Order Determines Priority

```css
@layer reset, base, components, utilities;
```

For normal rules: later layers win. `utilities` beats `components` beats `base` beats `reset`.

For `!important` rules: **earlier layers win**. `reset !important` beats `base !important` beats `components !important` beats `utilities !important`. This mirrors the origin inversion and is semantically correct -- defensive base styles should be the hardest to override.

### Unlayered Styles

Styles outside any `@layer` belong to an implicit final layer. They have the highest normal priority among author styles. This enables incremental adoption: existing unlayered CSS keeps working, and you progressively move code into layers.

Agent rule: when introducing `@layer` to an existing codebase, start by layering third-party and reset CSS. Move application code into layers incrementally. Unlayered application code will continue to override layered code during migration.

### Third-Party Integration

```css
@layer vendor, app;
@layer vendor { @import url("third-party.css"); }
```

All vendor specificity is now irrelevant to your application styles. A single-class app selector beats any vendor ID selector because `app` layer outranks `vendor` layer.

### Criterion 5: Specificity

The (A, B, C) lexicographic tuple. See css-principles.md for the full treatment.

### Criterion 6: Scope Proximity (Level 6)

When two declarations have equal specificity and both come from `@scope` rules, the one whose scoping root is closest (fewest DOM hops) to the target element wins.

```css
@scope (.card) {
  p { color: blue; }  /* 2 hops from .card > .body > p */
}
@scope (.body) {
  p { color: red; }   /* 1 hop from .body > p -- wins */
}
```

Key constraints:
- Proximity is "weak" -- specificity still takes precedence over proximity
- `@scope` limits selector reach, NOT style inheritance. Inherited properties (`color`, `font-family`) still flow through the lower boundary
- The lower boundary (`to` clause) defines where selectors stop matching, creating "donut scopes"

```css
@scope (.card) to (.card-content) {
  /* Styles the card chrome but NOT nested content */
  img { border-radius: 8px; }
}
```

Agent rule: `@scope` is selector isolation, not encapsulation. For true style isolation (blocking inheritance), Shadow DOM remains the only option.

### Criterion 7: Order of Appearance

Last declaration in document order wins. The final tiebreaker. When specificity, layers, and all other criteria are equal, the rule that appears later in the stylesheet (or later-imported file) wins.

## @property: Registered Custom Properties

`@property` registers custom properties with type safety, animation capability, and controlled inheritance:

```css
@property --brand-color {
  syntax: "<color>";
  inherits: true;
  initial-value: #0066cc;
}

@property --card-padding {
  syntax: "<length>";
  inherits: false;     /* Does NOT inherit -- scoped to element */
  initial-value: 1rem;
}
```

What registration enables:
- **Type checking**: the browser rejects invalid values at parse time (`--brand-color: 42px` is invalid)
- **Animation**: unregistered custom properties cannot be interpolated; registered ones with `syntax: "<color>"` or `syntax: "<length>"` animate smoothly
- **Controlled inheritance**: `inherits: false` scopes the property to the element where it is set, preventing cascade leakage to descendants

Agent rule: register custom properties when they represent typed design tokens (colors, lengths, durations). Leave unregistered for string-like tokens (font-family names, content strings) where type checking adds no value.

## The Convergence Thesis

CSS is becoming a component language. The evidence:

| Capability | Pre-2022 Solution | Native CSS Solution (2022-2026) |
|------------|-------------------|--------------------------------|
| Specificity management | BEM naming, hacks | `@layer` |
| Component scoping | CSS Modules, CSS-in-JS | `@scope` |
| Parent selection | JavaScript class toggling | `:has()` |
| Component-local responsiveness | Resize observers | `@container` |
| Nested selectors | Sass/Less | Native nesting |
| Cross-component alignment | DOM flattening | Subgrid |

The trajectory is irreversible. Each generation absorbs the core insight of the previous one and removes the tooling dependency.

Agent rule: prefer native CSS solutions over JavaScript for styling concerns (responsive behavior, state-driven styling, scoping). Zero runtime cost, zero bundle impact, survives framework migrations.

# CSS Architecture Principles

> Methodology-neutral CSS principles: token-driven values, logical properties, cascade control, layout, animation.

## Core Principles (Methodology-Neutral)

These principles hold regardless of whether the project uses BEM, utility-first, CSS modules, CSS-in-JS, or any other methodology.

### 1. Token-Driven Values Only

No raw color, spacing, or typography values in component code. Only token references.

**Valid**:
```css
/* component references alias token */
.button { background-color: var(--color-action-primary); }
```

**Invalid**:
```css
/* raw value in component — always a violation */
.button { background-color: #0066cc; }
.button { background-color: var(--blue-500); } /* tier-skip violation */
```

All style values in components reference component tokens, which reference alias tokens, which reference global tokens. Never skip the alias tier.

### 2. Logical Properties for Directional Spacing

Use CSS logical properties for any direction-dependent spacing. This is required for RTL language support and internationalization.

| Physical (avoid for directional) | Logical (use instead) |
|----------------------------------|----------------------|
| `margin-left` | `margin-inline-start` |
| `margin-right` | `margin-inline-end` |
| `padding-top` | `padding-block-start` |
| `padding-bottom` | `padding-block-end` |
| `left` (for positioning) | `inset-inline-start` |
| `right` (for positioning) | `inset-inline-end` |

Physical properties (`margin-left`, `padding-top`) remain valid for non-directional uses (e.g., vertical margin on elements that don't need RTL adaptation). Use judgment.

### 3. Cascade Control

Manage specificity explicitly. Uncontrolled cascade accumulation creates unpredictable overrides.

**Patterns**:
- Prefer class selectors (specificity 0,1,0) over element selectors (0,0,1) for component styles
- Avoid `!important` except for utility classes where override intent is explicit
- Use CSS layers (`@layer`) to establish explicit specificity tiers: reset < base < components < utilities < overrides
- Component styles should not depend on global element styles for their visual correctness

### 4. Layout: Modern Primitives First

Before writing layout code, prefer platform-native layout systems:
- **Grid**: two-dimensional layouts, explicit placement, named areas
- **Flexbox**: one-dimensional alignment, content-driven sizing
- **Container queries**: responsive component behavior based on available space, not viewport

Avoid `float` for layout. Avoid absolute/fixed positioning for element arrangement (use only for intentional overlay behavior).

### 5. Animation: Transform-Only (No Layout Properties)

Animations and transitions must use CSS `transform` and `opacity`. Never animate layout-triggering properties.

**Safe to animate**: `transform`, `opacity`, `filter`, `clip-path`

**Never animate** (causes layout recalculation, harms INP and CLS): `top`, `left`, `right`, `bottom`, `width`, `height`, `margin`, `padding`

Always respect `prefers-reduced-motion`:
```css
@media (prefers-reduced-motion: reduce) {
  /* disable or simplify all non-essential animations */
}
```

### 6. Font Loading

- Self-host fonts (eliminates third-party DNS lookup)
- `font-display: optional` for body text (no CLS)
- `font-display: swap` + size-adjusted fallbacks for display/heading fonts only
- Preload the primary font file: `<link rel="preload" as="font" crossorigin>`
- Subset fonts to needed character ranges

### 7. Focus Styles (Never Remove)

Never remove focus indicators. If the default browser outline is undesirable, replace it:

```css
/* WRONG — removes focus indicator entirely */
* { outline: none; }

/* CORRECT — custom focus style meeting WCAG 2.4.7 */
:focus-visible {
  outline: 2px solid var(--color-focus-ring);
  outline-offset: 2px;
}
```

Minimum requirements: visible, 2px thick, 3:1 contrast ratio (WCAG 2.4.7). Aim for 2.4.13 (AAA).

## Responsive Design

- **Mobile-first**: write base styles for smallest viewport, use `min-width` media queries to add complexity
- **Container queries**: preferred over viewport-based media queries for component-level responsiveness
- **Design for 2× text expansion**: German, Finnish, and other languages can be 30–200% longer than English; layouts must accommodate this
- Test at minimum three viewport widths: 375px (mobile), 768px (tablet), 1280px (desktop)

## CSS Custom Properties as the Token Bridge

CSS custom properties (`var(--token-name)`) are the runtime bridge between design tokens and CSS. They enable:
- Theming without class changes (swap custom property values for brand/dark mode)
- Component-scoped token overrides via cascade
- Runtime value inspection

**Naming convention**: follow the token naming pattern in custom properties — `--color-action-primary`, `--space-inline-md`, not `--primary`, `--gap`.

## Anti-Patterns

**Hardcoded values in components**: any hex color, `px` spacing value, or font family in component code that should be a token reference.

**Global element style pollution**: styling raw HTML elements (`p`, `a`, `button`) globally in a way that makes component styles dependent on those globals — fragile, non-portable.

**Specificity escalation**: adding specificity to override other specificity rather than refactoring. The correct response to a specificity conflict is to restructure, not to escalate.

**`outline: none` without replacement**: always a WCAG violation.

## Cascade Fundamentals

The cascade is the algorithm that resolves conflicting CSS declarations into a single winner per property per element. Most developers reduce it to "specificity wars." In reality, specificity is criterion 5 of 7. Four higher-priority mechanisms override it entirely.

### The 7-Criterion Sorting Algorithm

When multiple declarations target the same property on the same element, the cascade resolves them in strict descending priority. A later criterion is only consulted when all preceding criteria tie.

| Priority | Criterion | What It Resolves |
|----------|-----------|-----------------|
| 1 | **Origin and Importance** | Which source (UA, user, author) and whether `!important` |
| 2 | **Context** | Encapsulation boundaries (Shadow DOM vs. light DOM) |
| 3 | **Element-Attached Styles** | Inline `style` attribute vs. rule-based declarations |
| 4 | **Cascade Layers** | `@layer` ordering within an origin |
| 5 | **Specificity** | Selector weight as (A, B, C) tuple |
| 6 | **Scope Proximity** | Distance from `@scope` root to matched element (Level 6) |
| 7 | **Order of Appearance** | Last declaration in document order wins |

Agent enforcement rule: when debugging a style override, identify WHICH cascade criterion is producing the winner before changing code. Most "specificity fixes" are actually layer or source-order problems.

### Specificity Is Lexicographic, Not Arithmetic

Specificity is a three-component tuple `(A, B, C)` compared left-to-right like dictionary ordering:
- **A**: count of ID selectors
- **B**: count of class selectors, attribute selectors, pseudo-classes
- **C**: count of type (element) selectors, pseudo-elements

The "1000/100/10/1 point system" taught in most tutorials is formally incorrect. `(1, 0, 0)` always beats `(0, N, 0)` regardless of N -- there is no number of classes that overrides a single ID. The CSS 2.1 spec's mention of "a large base" was technically correct (the base would need to be infinite) but spawned a generation of wrong mental models.

### `!important` Is an Origin Shift, Not a Specificity Boost

`!important` does not "add specificity." It moves the declaration to a different cascade origin. The full origin priority order (highest to lowest):

1. Transition declarations
2. Important UA declarations
3. Important user declarations
4. **Important author declarations** (weakest form of important)
5. Animation declarations
6. Normal author declarations
7. Normal user declarations
8. Normal UA declarations

Author `!important` is the weakest important origin. UA `!important` (browser accessibility defaults) always wins -- by design, so users can enforce high-contrast modes and font sizes.

Agent enforcement rule: `!important` in author code is almost always a cascade-level misunderstanding. Use `@layer` ordering or restructure specificity instead.

### @layer Makes Specificity Irrelevant Across Boundaries

Within a layer, normal specificity rules apply. Across layers, layer order always wins:

```css
@layer components, utilities;

@layer components {
  #sidebar .nav .item a.link { color: blue; }  /* (1, 3, 1) -- loses */
}
@layer utilities {
  .text-red { color: red; }  /* (0, 1, 0) -- wins */
}
```

The `(0, 1, 0)` utility beats the `(1, 3, 1)` component because layer ordering (criterion 4) outranks specificity (criterion 5). Unlayered styles sit in an implicit final layer with the highest normal priority.

`!important` reverses layer order: earliest-declared layers win for important rules. A reset layer's `!important` beats a utilities layer's `!important` -- semantically correct for defensive base styles.

### Selector Performance: Inside the Braces

Modern browser engines have made selector matching negligible through four optimizations: rule hashing (bucket by rightmost selector), Bloom filter ancestor matching (near-O(1) rejection), style sharing (reuse computed styles for identical siblings), and fast-path matching (inlined loop for common combinators).

Empirical benchmarks converge: property cost dominates selector cost. Expensive properties like `box-shadow` and `filter` cause up to 112x paint time increase. Selector complexity variance across an entire stylesheet is typically under 2ms.

Agent enforcement rule: do not refactor selectors for "performance." Focus paint-cost optimization on expensive properties and DOM tree size. The only selector performance concern is `:has()` anchored to very broad selectors (`body`, `:root`, `*`) on massive DOMs.

### Specificity-Adjusting Pseudo-Classes

| Pseudo-class | Specificity Behavior |
|-------------|---------------------|
| `:is()` | Takes specificity of the most specific argument |
| `:not()` | Takes specificity of the most specific argument |
| `:has()` | Takes specificity of the most specific argument |
| `:where()` | Always `(0, 0, 0)` regardless of arguments |

`:where()` is the library escape hatch -- wrap selectors in `:where()` to produce zero-specificity defaults that any consumer can trivially override. Use for resets, design system defaults, and third-party component libraries.

### The Specificity Ceiling Principle

The lower the maximum allowed specificity in a codebase, the more predictable the system. Common ceilings:

| Architecture | Ceiling | Resolution Mechanism |
|-------------|---------|---------------------|
| BEM flat | `(0, 1, 0)` | Order of appearance |
| Utility-first + @layer | `(0, 1, 0)` | Layer ordering |
| Component-scoped | `(0, 2, 1)` | Scoped selectors, limited nesting |

`@layer` makes strict ceiling enforcement less critical by providing an orthogonal precedence mechanism, but a low ceiling within each layer still produces the most maintainable CSS.

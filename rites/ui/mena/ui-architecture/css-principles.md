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

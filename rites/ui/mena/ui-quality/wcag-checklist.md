# WCAG 2.2 AA Checklist

> Quick reference for the criteria most commonly violated, with agent enforcement rules.

## Legal Baseline

WCAG 2.2 Level AA is not aspirational — it is the standard referenced by:
- US ADA Title III litigation: 4,000+ federal/state lawsuits in 2024; 2,014 in H1 2025 alone (37% YoY increase)
- European Accessibility Act: enforced June 28, 2025 across all EU member states; fines up to 4% of annual revenue
- Section 508 (US federal)

Accessibility overlays (accessiBe, UserWay, AudioEye) do not satisfy these requirements. The FTC fined accessiBe $1M (April 2025). 25% of 2024 ADA lawsuits targeted sites using overlays.

## Key WCAG 2.2 AA Criteria

### Perceivable

| Criterion | Requirement | Agent Rule |
|-----------|-------------|-----------|
| 1.1.1 Non-text Content | All images need alt text; decorative images get `alt=""` | Flag any `<img>` without alt attribute |
| 1.3.1 Info and Relationships | Semantic HTML conveys structure | Use heading hierarchy, lists, tables, landmarks |
| 1.4.1 Use of Color | Color must not be the sole means of conveying information | Pair color with icon, text, or pattern |
| 1.4.3 Contrast Minimum | 4.5:1 normal text; 3:1 large text (18pt/24px or 14pt bold/19px) | Validate all color combinations algorithmically |
| 1.4.11 Non-text Contrast | 3:1 for UI components and graphical objects | Flag any UI component border below 3:1 |

### Operable

| Criterion | Requirement | Agent Rule |
|-----------|-------------|-----------|
| 2.1.1 Keyboard | All functionality operable via keyboard | Every interactive element must be Tab-reachable |
| 2.4.7 Focus Visible | Keyboard focus indicator always visible | Never remove outline without replacement |
| 2.4.11 Focus Not Obscured (NEW 2.2) | Focused element not hidden by sticky headers/footers | Check fixed/sticky positioning against focus |
| 2.5.8 Target Size Minimum (NEW 2.2) | Interactive targets at least 24×24 CSS pixels | Flag any interactive element < 24×24; flag touch targets < 44×44 |

### Understandable

| Criterion | Requirement | Agent Rule |
|-----------|-------------|-----------|
| 3.3.1 Error Identification | Error fields identified and described in text | Error messages must name the field and describe the error |
| 3.3.2 Labels or Instructions | Form inputs must have visible labels | Never use placeholder as sole label |
| 3.3.7 Redundant Entry (NEW 2.2) | Don't ask for information already provided in same session | Check multi-step forms |
| 3.3.8 Accessible Authentication (NEW 2.2) | No cognitive function tests for login | Paste and password managers must work |

### Robust

| Criterion | Requirement | Agent Rule |
|-----------|-------------|-----------|
| 4.1.2 Name, Role, Value | All components expose accessible name, role, and state | Every custom widget needs correct ARIA |

## Semantic HTML First, ARIA Second

Use native HTML elements for their intended purpose. ARIA is a repair mechanism when native semantics are insufficient.

| Use | Instead of |
|-----|-----------|
| `<button>` | `<div role="button">` |
| `<nav>` | `<div role="navigation">` |
| `<main>`, `<header>`, `<footer>`, `<aside>` | `<div>` with role |
| `<dialog>` | Custom modal with focus trapping |
| Heading hierarchy (h1–h6) reflecting structure | Headings chosen for visual size |

**When ARIA is required**: ARIA roles are permitted only when no native element exists for the pattern — e.g., `role="tablist"`, `role="alert"`, `aria-live` regions. When agents generate ARIA, they must also generate the corresponding keyboard event handlers. ARIA without keyboard support is worse than no ARIA.

## Keyboard Interaction Requirements

- Modal dialogs: Tab cycles within modal, Escape closes, focus returns to trigger on close
- Focus indicators: never `outline: none` without a custom `:focus-visible` replacement meeting WCAG 2.4.7 (visible) and ideally 2.4.13 (2px thick, 3:1 contrast)
- Tab order follows visual reading order (no `tabindex > 0`)
- Skip navigation links provided for repeated content blocks
- When content changes dynamically (SPA navigation, form validation), focus is explicitly managed

## What Automated Tools Catch vs What They Cannot

| Automatable (57% of issues) | Not Automatable (43% of issues) |
|------------------------------|----------------------------------|
| Color contrast ratios | Whether alt text is meaningful |
| Missing alt text | Logical reading order |
| Missing form labels | Keyboard focus management correctness |
| Duplicate IDs | Whether ARIA live regions announce appropriately |
| Missing language attributes | Custom widget APG pattern adherence |
| ARIA attribute validity | Meaningful page titles |
| Heading level skips | Correct use of headings for document structure |

"Passes axe-core" is a necessary floor, not a sufficient ceiling.

## Internationalization (i18n) as Structural Constraint

- Use CSS logical properties (`margin-inline-start`, `padding-block-end`) instead of physical properties for direction-dependent spacing
- Design layouts accommodating 2× text expansion (German/Finnish text is 30–200% longer than English)
- Never hard-code text strings in markup — all user-visible text must be externalizable
- All ARIA labels and screen reader text must be localizable

## CSS-Accessibility Architecture

CSS properties can silently alter the accessibility tree, and user-preference media queries demand architectural responses. These patterns are not optional enhancements -- they map directly to WCAG criteria and forced-colors compliance.

### Focus Management Architecture

| Property | Forced Colors | Border-Radius | Verdict |
|----------|---------------|---------------|---------|
| `outline` | Preserved | Follows radius (Chrome 94+, Firefox 88+, Safari 16.4+) | Use this |
| `box-shadow` | **Stripped entirely** | Follows radius | Never use alone for focus |
| `border` | Preserved | Follows radius | Shifts layout -- avoid |

**Agent Rule**: Never use `box-shadow` as the sole focus indicator. It is removed in forced-colors mode (Windows High Contrast), making focus invisible for users who need it most.

**Two-color focus pattern** (W3C C40): A dark inner outline + light outer shadow guarantees 3:1 contrast against ANY background. At least one of two colors with 9:1 mutual contrast will always meet 3:1 against any solid background.

```css
:focus-visible {
  outline: 3px solid #000000;
  box-shadow: 0 0 0 6px #FFFFFF;
}
```

`outline-offset` is a design tool: positive values add breathing room, negative values create inset rings for colored backgrounds.

### Motion Architecture

**Agent Rule**: Use no-motion-first architecture. Default to no animation; enhance with `prefers-reduced-motion: no-preference`. This is safer because users with unsupported browsers get no animations, and "no-preference" does not equal consent.

Use `0.01s` duration under reduced-motion, never `0s`. Zero duration breaks JS `animationend`/`transitionend` callbacks.

```css
:root { --animation-duration: 0.3s; }
@media (prefers-reduced-motion: reduce) {
  :root { --animation-duration: 0.01s; }
}
```

### CSS Properties That Break the Accessibility Tree

| CSS Pattern | Semantic Effect | Fix |
|-------------|----------------|-----|
| `list-style: none` | Strips list semantics in Safari (intentional WebKit decision) | `role="list"` or `list-style-type: ""` |
| `display: flex/grid` on `<table>` | Strips table semantics | Explicit ARIA table roles |
| `display: contents` | Removes element from accessibility tree | Avoid on semantic elements |
| `order`, `grid-row` visual reorder | Does NOT change tab/reading order | Source order must match visual order |

**Agent Rule**: When generating styled lists with `list-style: none`, always add `role="list"`. When applying CSS display overrides on `<table>`, always add `role="table"`, `role="row"`, `role="cell"`.

### Forced-Colors Mode

When `forced-colors: active`, the browser overrides author colors with system colors. `box-shadow` is removed entirely. System color keywords for custom decorations: `Canvas`, `CanvasText`, `Highlight`, `HighlightText`, `ButtonFace`, `ButtonText`, `GrayText`.

**Agent Rule**: In `@media (forced-colors: active)`, adjust structural tokens only (border-width, outline-width). Do not set color tokens -- the UA handles those. Use system color keywords only for custom decorative elements.

### The CSS-Only Trap

Five component categories CANNOT be accessible without JavaScript:

| Component | Why CSS-Only Fails |
|-----------|--------------------|
| Tooltips | No Esc dismissal, inaccessible on touch |
| Modals | No focus trapping, no Esc close, no focus return |
| Tabs | Checkbox hack misrepresents semantics to AT and speech recognition |
| Carousels | No slide position announcement, no keyboard control |
| Dropdowns | No `aria-expanded` toggle, no Esc close |

**Agent Rule**: If the interaction pattern appears in the WAI-ARIA APG, it requires JavaScript. The checkbox hack is never acceptable -- it misrepresents semantics to screen readers and speech recognition software. Exception: native `<details>`/`<summary>` IS accessible without JS.

### Responsive Accessibility Constraints

Three INDEPENDENT WCAG criteria that are commonly conflated:

| Criterion | Constraint | CSS Rule |
|-----------|-----------|----------|
| 1.4.4 Resize Text | Text scales to 200% without loss | Never use `px` for font sizes; use `rem`/`em` |
| 1.4.10 Reflow | No horizontal scrolling at 320px viewport | No fixed-width containers; use flexbox/grid |
| 1.4.12 Text Spacing | Content functions with increased spacing | No fixed `height` on text containers; no `overflow: hidden` on text |

**Agent Rule**: These are three independent checks. Meeting Reflow does not satisfy Resize Text or Text Spacing. Flag any `px` font-size, any `overflow: hidden` on text-containing elements, and any fixed `height` on elements with text content.

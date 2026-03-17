# Token Taxonomy

> DTCG three-tier token hierarchy, semantic naming, and anti-patterns.

## The Three-Tier Architecture (Non-Negotiable)

Design tokens MUST be organized in exactly three tiers. Skipping or collapsing tiers creates systems that cannot theme, cannot scale, and cannot be consumed by agents.

| Tier | Purpose | Naming | Example |
|------|---------|--------|---------|
| **Global** | Raw values — the palette | Visual names permitted | `blue-500`, `space-4` |
| **Alias/Semantic** | Purpose-mapped — the contract | Semantic names required | `color.action.primary` |
| **Component** | Scoped to one component | References alias tier only | `button.background.default` |

**Agent enforcement rule**: Component tokens reference alias tokens. Alias tokens reference global tokens. Never skip tiers. If a component references a global-tier token directly, flag it as a tier-skip violation.

## DTCG 2025.10 Standard

The W3C Design Tokens Community Group 2025.10 specification is the vendor-neutral interchange format, backed by Adobe, Amazon, Google, Microsoft, Meta, Shopify, Figma, and 15+ organizations.

**DTCG JSON structure**:
- Objects with `$value` are tokens
- Objects without `$value` are groups
- `$type` must be one of: `color`, `dimension`, `fontFamily`, `fontWeight`, `duration`, `cubicBezier`, `number`, or composites (`shadow`, `border`, `transition`, `gradient`, `strokeStyle`, `typography`)
- Alias references use `{group.token}` curly-brace syntax
- Token names must not contain `$`, `{`, `}`, or `.`

## Semantic Token Naming

Token names must describe purpose and usage, never visual properties.

**Naming structure**: `{namespace}.{category}.{property}.{variant/scale}.{state}`
- Namespace and state are optional
- Category examples: `color`, `space`, `typography`
- Property examples: `background`, `text`, `border`, `icon`
- Variant examples: `primary`, `success`, `danger`, `subtle`

**Valid alias-tier names**: `color.action.primary`, `color.text.subtle`, `color.feedback.success`, `space.inline.md`

**Invalid alias-tier names**: `blue-primary`, `color.blue.primary`, `text.gray.muted` — any color word in an alias token is a naming violation.

**Test**: Does the name survive a brand color change? `color.action.primary` does. `color.blue-action` does not.

## Multi-Context Theming: Resolver Module

When supporting multiple brands, themes, or density modes, the DTCG Resolver Module (draft, March 2026) eliminates the M×N file explosion.

A resolver document declares:
- **Sets**: token collections from one or more source files
- **Modifiers**: contextual variations (light/dark theme, comfortable/compact density)
- **resolutionOrder**: cascade precedence array — later entries override earlier on conflict

Structure: `global/` (raw values) → `semantic/` or `alias/` (purpose-mapped) → `component/` (scoped).

## Anti-Patterns

**AP: Visual naming at semantic tier** — Using color words in alias or component tokens (`color.blue.primary`). Detection: scan alias-tier names for color words (red, blue, green, gray, white, black). Correction: replace with functional names.

**AP: Tier-skipping** — Component code references global tokens directly (`blue-500` in button background). Detection: grep component code for numeric-scale tokens. Correction: create alias token, reference alias from component.

**AP: Brand fork** — Multiple component libraries with overlapping names instead of shared components with brand-specific token collections. Correction: consolidate to single library, remap alias tokens per brand.

## OKLCH Color Architecture

OKLCH is perceptually uniform -- equal L steps produce visually equal lightness changes across ALL hues. HSL is not: blue at 50% lightness appears far darker than yellow at 50%.

`oklch(L C H)` -- L: 0-1 lightness (perceptually uniform), C: chroma (saturation intensity), H: 0-360 hue angle.

| Capability | HSL | OKLCH |
|-----------|-----|-------|
| Perceptual uniformity | No -- lightness varies wildly across hues | Yes -- equal steps = equal visual change |
| Programmatic palette generation | Requires manual per-hue tweaking | Mathematical: adjust L and C by fixed amounts |
| Wide-gamut (P3) colors | Cannot represent | Native support (30% more colors on modern displays) |
| Dark mode contrast prediction | Unreliable | Predictable -- systematic L adjustment yields reliable contrast ratios |

OKLCH enables programmatic palette generation: define a hue, generate an entire shade scale by adjusting L and C mathematically. Dark mode becomes systematic L-value adjustment rather than manual color picking.

Browser support: `oklch()` is baseline across all modern browsers since 2023. The DTCG 2025.10 spec supports OKLCH as a first-class color space (`"colorSpace": "oklch"`).

**Agent enforcement rule**: New token systems SHOULD define colors in OKLCH. Existing hex/RGB/HSL systems can continue but new tokens should use OKLCH for palette consistency and P3 support.

## CSS Custom Properties as Token Runtime

Key distinction: `var()` is runtime, cascade-aware, and inheritable. Sass `$var` is compile-time, lexically scoped, and static.

| Property | Sass `$variable` | CSS `var(--property)` |
|----------|-------------------|------------------------|
| Resolution | Compile time (disappears from output) | Runtime (lives in the browser) |
| Cascade-aware | No -- same value everywhere | Yes -- value changes per selector context |
| Inheritable | No -- lexically scoped | Yes -- inherits down the DOM tree |
| Media-query-aware | No | Yes -- redefine inside `@media` |
| JS-accessible | No | Yes -- `getComputedStyle()` / `setProperty()` |

The CSS cascade IS a theming engine: custom properties change value per selector context (theme, brand, breakpoint) without JavaScript or rebuild. Components reference `var(--token)` and the cascade resolves the correct value automatically.

**`@property` registered custom properties** add type safety, animation capability, and controlled inheritance:

```css
@property --color-primary {
  syntax: "<color>";
  inherits: true;
  initial-value: #0066cc;
}
```

Register foundational tokens with `@property` for type safety and fallback guarantees (`syntax` constrains values; invalid input falls back to `initial-value`). Keep component-specific properties unregistered for flexibility and `var()` fallback chains. `@property` also enables smooth animation of custom properties -- unregistered properties cannot be transitioned.

**Agent enforcement rule**: If a token value appears in DevTools as a raw value (not `var(--name)`), it is not properly tokenized. Every design decision in component CSS must trace back to a custom property.

## Style Dictionary Pipeline

Style Dictionary is the de facto build system for cross-platform token transformation. The v4 pipeline has nine stages:

Parse config --> locate token files --> parse files (JSON, JSONC, JSON5, ESM) --> deep merge --> preprocess --> transform tokens --> resolve references --> format output --> execute actions.

Key architectural facts:
- **Deep merge** (stage 4): tokens can be split across files without override risk
- **Transforms before resolution** (stages 6-7): aliases resolve to already-transformed values
- **Transitive transforms**: enable transformation of referenced values after resolution -- required when a referenced value itself needs platform-specific modification

Transform groups bundle platform presets: `web`, `css`, `scss`, `js`, `android`, `ios-swift`, `compose`, `flutter`, `react-native`. DTCG format (`$value`, `$type` prefix notation) is first-class in v4.

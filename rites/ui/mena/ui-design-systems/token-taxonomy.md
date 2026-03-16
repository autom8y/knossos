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

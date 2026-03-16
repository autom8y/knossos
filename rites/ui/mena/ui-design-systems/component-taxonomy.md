# Component Taxonomy

> Primitives/Composites/Patterns classification, headless composition, slot API design, lifecycle status.

## Three-Tier Classification by Behavioral Role

Organize components by behavioral complexity, not visual metaphor. Avoid atoms/molecules/organisms labels — useful mental model, poor naming convention.

| Tier | Definition | Constraints | Examples |
|------|-----------|-------------|---------|
| **Primitive** | Single-responsibility element | Zero internal component dependencies beyond tokens | Button, Input, Icon, Badge, Avatar |
| **Composite** | Combinations of primitives with shared state | Composed of named Primitives via slots or children | FormField, Dialog, Card, Dropdown |
| **Pattern** | Layouts and behaviors composed of composites | Defines layout/coordination logic, not new visual primitives | DataTable, PageLayout, NavigationBar |

**Agent enforcement rules**:
- Identify which tier a component belongs to before creating it
- Primitives: zero component dependencies
- Composites: composed of named Primitives
- Patterns: layout/coordination only — if a Pattern contains raw HTML instead of Primitives, flag as tier violation

## Headless Logic Separation

Component logic (state management, keyboard navigation, focus management, ARIA attributes) must be separable from component rendering (markup, styling). This is the only architecture proven framework-agnostic.

**When building any interactive component**:
1. Implement behavior logic in a framework-agnostic layer (hook, state machine, or pure function)
2. Keep rendering as a thin consumer of that logic
3. ARIA roles, keyboard handlers, and focus management live in the logic layer — not the markup

**Detection**: If `aria-*` attributes or `onKey*` handlers appear directly in render/template code without sourcing from a logic layer, flag as coupling violation for any component intended for reuse.

## Slots vs Props: API Design Decision

**Decision rule**: If a prop type would be `ReactNode`, `Component`, or `Element` — convert to a named slot.

| Configuration type | Use |
|-------------------|-----|
| Primitive value (string, number, boolean, enum) | Prop |
| Behavioral configuration (disabled, variant, size) | Prop |
| Content that could be any component or element | Slot |
| What is rendered without changing behavior (icon, actions, footer) | Slot |

**Audit signal**: If a component has more than 8–10 props, audit for slot candidates. If a component has >20 props, it is a God Component — decompose.

## Component Decision Tree: New vs Variant vs Pattern

Before creating a new component, apply this decision flow:

1. Differs from an existing component only in color, size, or format? → **Variant**
2. Composes existing components into a new layout or interaction? → **Pattern**
3. Genuinely novel behavior not present in any existing component? → **New Primitive**
4. Used in fewer than 3 places? → **Keep inline** — do not add to system
5. Would require more than 3 new props on existing component? → Probably a new component, not a variant

## Lifecycle Status (Machine-Readable)

Every component must carry a lifecycle status queryable by tooling and agents:

| Status | Meaning | Agent rule |
|--------|---------|-----------|
| Draft | Work in progress | Do not use in production code |
| Experimental/Beta | Production-quality code, instability possible | Flag with warning comment |
| Stable | Full functionality, complete documentation | Default choice |
| Deprecated | No updates, replacement exists | Never use in new code |
| Retired | Removed from system | Does not exist |

**Agent enforcement**: Check lifecycle status before selecting any component. If status is unavailable, treat as Experimental.

## Structured Data for Agent Consumption

A design system consumed by AI agents must be exposed as machine-readable structured data:
- Token definitions in DTCG JSON format
- Component metadata including props, slots, variants, states, and accessibility requirements
- A rules file encoding naming conventions and architectural decisions (CLAUDE.md, .cursor/rules/, AGENTS.md)

If no structured data source exists, flag as a prerequisite gap before generating code.

---
name: ui-design-systems
description: "Design system architecture principles: token taxonomy (DTCG three-tier), OKLCH color architecture, CSS custom properties as token runtime, Style Dictionary pipeline, semantic naming, component classification, headless composition, slot API design, and structured data for agent consumption. Use when: designing component architecture, defining token naming conventions, classifying components, evaluating composition patterns, choosing color space, setting up token build pipeline. Triggers: design system, tokens, DTCG, component taxonomy, primitives, composites, patterns, slots, headless, OKLCH, color space, Style Dictionary, @property, custom properties, theming."
---

# UI Design Systems

> Durable, stack-agnostic principles for token taxonomy, component architecture, and governance.

## Overview

Design system architecture has converged on a small set of industry-validated, framework-agnostic principles. The three-tier token hierarchy (Global/Alias/Component) is now ratified by the W3C DTCG 2025.10 stable specification. Component taxonomy survives framework churn when organized by behavioral role. Headless logic-presentation separation is the only composition architecture proven truly framework-agnostic.

## Contents

| File | Purpose |
|------|---------|
| [token-taxonomy.md](token-taxonomy.md) | DTCG three-tier hierarchy, OKLCH color, CSS custom properties, Style Dictionary pipeline |
| [component-taxonomy.md](component-taxonomy.md) | Primitives/Composites/Patterns classification, slots vs props, lifecycle status |
| [governance.md](governance.md) | Design-code pipeline, governance gates, versioning, RFC lifecycle |

## When to Use

**token-taxonomy.md** — When generating, reviewing, or auditing design tokens; when naming tokens; when evaluating theming architecture; when setting up a DTCG pipeline; when choosing color space (OKLCH vs HSL); when implementing CSS custom property token runtime; when configuring Style Dictionary.

**component-taxonomy.md** — When creating or classifying a component; when designing a component API; when deciding slots vs props; when checking if a component should be promoted to the design system.

**governance.md** — When setting up CI/CD for a design system; when classifying a breaking change; when designing a contribution workflow; when evaluating token pipeline stages.

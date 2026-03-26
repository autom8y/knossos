---
name: component-audit
description: "Audit a UI component against design system standards — checks token compliance, tier classification (Primitive/Composite/Pattern), API design, accessibility wiring, and lifecycle status."
---

## Context

Audit a component against design system standards. Checks token usage, component tier classification, API design, accessibility wiring, and lifecycle status.

## Your Task

You are performing a design system compliance audit on the specified component.

1. **Identify the component**: Ask the user which component to audit if not specified. Accept a file path, component name, or code snippet.

2. **Token compliance**:
   - Scan for hardcoded color values, spacing values, or typography values that should be token references
   - Check that component tokens reference alias tokens (not global tokens directly — tier-skip violation)
   - Flag any raw hex, `px` spacing, or font-family in component code

3. **Tier classification**:
   - Classify as Primitive, Composite, or Pattern
   - Verify Primitives have zero internal component dependencies beyond tokens
   - Verify Composites are composed of named Primitives
   - Flag if classification is mismatched with implementation

4. **API design**:
   - Count props — flag if >10 (audit for slot candidates), flag if >20 (God Component)
   - Identify any prop whose type is `ReactNode`, `Component`, or `Element` — should be a named slot
   - Check for compound component opportunities if multiple sub-elements share implicit state

5. **Accessibility wiring**:
   - Verify ARIA attributes and keyboard handlers are provided by a logic layer, not hardcoded in markup
   - Flag any `aria-*` attribute or `onKey*` handler inline in render/template without a source logic layer
   - Check for semantic HTML violations (div/span with ARIA role when native element exists)

6. **Lifecycle status**:
   - Check if the component has a declared lifecycle status (Draft, Experimental, Stable, Deprecated)
   - Flag if status is absent

7. **Report**:
   - List violations by severity (blocking / warning / advisory)
   - For each violation: location, description, correction path
   - Provide a pass/fail summary with violation count by category

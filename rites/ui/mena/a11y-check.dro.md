---
name: a11y-check
description: "Run WCAG 2.2 AA accessibility validation on UI code or components — checks semantic HTML, color contrast, keyboard interaction, form patterns, dynamic content, and new 2.2 criteria."
---

## Context

Run an accessibility validation check against WCAG 2.2 AA standards. Reviews code or describes a UI for structural accessibility issues across the four testable layers: static analysis, automated audit patterns, keyboard interaction, and authoring-time patterns.

## Your Task

You are performing a WCAG 2.2 AA accessibility validation on the specified UI code or component.

1. **Identify the target**: Ask the user which component, page, or code to audit if not specified.

2. **Semantic HTML check**:
   - Flag any `<div role="button">`, `<div role="navigation">`, etc. where native elements exist
   - Verify heading hierarchy reflects document structure (no level skips, no headings chosen for visual size)
   - Check landmark regions are present (`<main>`, `<nav>`, `<header>`, `<footer>`, `<aside>`)

3. **Color and contrast**:
   - Flag any color token combinations that need contrast validation
   - Remind: 4.5:1 for normal text, 3:1 for large text (18pt/24px or 14pt bold/19px), 3:1 for UI components
   - Flag any color-only information encoding (no secondary indicator)

4. **Interactive element requirements**:
   - Verify all interactive elements are keyboard-reachable
   - Check focus indicators — flag any `outline: none` or `outline: 0` without replacement
   - Check target size — flag any interactive target likely < 24×24 CSS pixels
   - For modals/dialogs: verify focus trap, Escape closes, focus returns to trigger on close

5. **Forms**:
   - Every input must have a programmatically associated label (not placeholder)
   - Required fields must be indicated programmatically (`aria-required`) and visually
   - Error messages must identify the field and describe the error in text
   - Verify `aria-live` or `role="alert"` on error message containers

6. **Dynamic content**:
   - Check for SPA navigation and form validation: is focus explicitly managed after content changes?
   - Check for `aria-live` regions on dynamic updates (notifications, loading states)

7. **New WCAG 2.2 criteria** (often missed):
   - 2.4.11 Focus Not Obscured: focused element not hidden by sticky headers/footers
   - 2.5.8 Target Size Minimum: 24×24 CSS pixels minimum
   - 3.3.7 Redundant Entry: multi-step forms should not re-ask for provided information
   - 3.3.8 Accessible Authentication: paste and password managers must work

8. **Report**:
   - List violations by WCAG criterion with severity (critical/serious/moderate/minor)
   - Distinguish: automatable violations (axe-core would catch) vs requires-authoring-discipline
   - Provide correction for each violation
   - Note what requires manual screen reader testing (cannot be automated)

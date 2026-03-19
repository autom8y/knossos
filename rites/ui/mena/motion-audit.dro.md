---
name: motion-audit
description: "Assess existing animations and transitions against motion architecture principles: frequency x novelty classification, interaction physics (interruptibility, responsive delta, proportional values, spatial truth, trigger-proximate feedback, gesture intent), 200ms ceiling, spring vs keyframe appropriateness."
argument-hint: "<component, page, or URL to audit>"
allowed-tools: Read, Glob, Grep, Write, Skill
model: sonnet
---

## Context

Standalone motion architecture assessment utility. Direct invocation of motion-architect for targeted motion evaluation. Analogous to `/a11y-check` for accessibility and `/perf-budget` for performance.

## Your Task

You are performing a motion architecture assessment on the specified target.

1. **Identify the target**: Accept component name, file path, URL, or code snippet. Ask the user if not specified.

2. **Inventory existing motion**: Catalog all animations, transitions, and motion effects in the target.
   - CSS transitions and animations
   - JavaScript-driven animations
   - Framer Motion / GSAP / Web Animations API usage
   - Hover, focus, and active state transitions

3. **Frequency x novelty assessment**: For each animation, classify:
   - Frequency tier (daily/hundreds, regular/tens, rare/once-per-session, once-ever)
   - Whether motion budget is appropriate for the tier (stripped for daily, expressive for rare)
   - Load the `motion-architecture` skill for the full classification matrix

4. **Interaction physics checklist**: For each interactive animation, evaluate all six criteria:
   - **Interruptibility**: Can the user interrupt the animation mid-flight?
   - **Responsive delta**: Does animation distance/duration scale with input magnitude?
   - **Proportional values**: Are timing and distance proportional to the trigger?
   - **Spatial truth**: Does motion reflect real spatial relationships?
   - **Trigger-proximate feedback**: Does feedback appear near the trigger?
   - **Gesture intent classification**: Is the gesture correctly classified (discrete vs. continuous)?

5. **Timing assessment**:
   - Flag any interactive animation exceeding 200ms ceiling
   - Assess spring vs. keyframe appropriateness (springs for interactive, keyframes for ambient)
   - Evaluate stagger choreography if multiple elements animate together (20-30ms per element appropriate)
   - Check choreography-reversal (elements exit in the direction they entered)

6. **Reduced motion compliance**:
   - Check for `prefers-reduced-motion` support
   - Verify reduced-motion uses `animation-play-state`, not `display: none` (which breaks layout)
   - Flag animations with no reduced-motion alternative
   - Flag essential spatial motion that has no reduced-motion fallback (loss of orientation)

7. **Report**:

   **Motion Inventory**: List every animation with frequency tier, physics checklist results, timing measurement.

   **Violations by severity**:
   - Blocking: interruptibility failures, 200ms ceiling violations, no reduced-motion support
   - Warning: frequency tier mismatch (expressive animation on daily interaction), spring/keyframe misuse
   - Advisory: choreography improvement opportunities, novelty budget observations

   **Recommendations**: Prioritized by impact. Each recommendation specifies: what to change, why (principle violated), and how (motion-architecture-spec format if applicable).

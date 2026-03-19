# Frequency x Novelty Motion Matrix (F1)

Framework for classifying interactions by usage frequency to determine appropriate animation complexity. Source: Freiberg's Web Interface Guidelines [WEAK single-source -- reversible if classification proves too rigid].

## The Core Principle

Animation complexity should be proportional to interaction frequency. Frequent interactions demand stripped motion; rare interactions earn expressive motion. Users cannot afford to watch animations on interactions they perform hundreds of times per day.

## Frequency Tiers

### Tier 1: Daily Interactions (hundreds/day)

Examples: typing in a text field, scrolling, clicking primary navigation, submitting common forms.

**Motion prescription**: Stripped. Near-zero animation. Confirm-only feedback.
- Transitions: <50ms or none
- No spring physics -- pure CSS transitions or instant
- State changes communicate via color/opacity only
- No choreography, no stagger, no personality
- Reduced-motion: identical to full-motion (nothing to remove)

**Rationale**: Users performing an action hundreds of times daily cannot afford to wait for it. Animation that delights the first time becomes friction by the hundredth.

### Tier 2: Regular Interactions (tens/day)

Examples: opening a sidebar, switching tabs, filtering a list, expanding an accordion.

**Motion prescription**: Subtle. Spatial orientation provided; personality restrained.
- Transitions: 100-150ms
- Simple easing (ease-out for entry, ease-in for exit)
- Spring physics: low tension, low bounce
- Minimal choreography if multiple elements
- Reduced-motion: instant state change, no transition

**Rationale**: Regular interactions benefit from spatial orientation (where did this content come from? where did it go?) but cannot afford personality expression. Subtle motion aids wayfinding.

### Tier 3: Rare Interactions (once/session)

Examples: opening a command palette, triggering a major modal, completing a checkout, sending a form.

**Motion prescription**: Expressive. Motion communicates meaning; choreography appropriate.
- Transitions: 150-200ms (hard ceiling at 200ms)
- Spring physics: higher tension, natural bounce
- Stagger choreography for related elements
- Personality can emerge in easing curves
- Reduced-motion: minimal spatial motion only, remove personality elements

**Rationale**: Once-per-session interactions earn animation investment. Users remember them. They are transitions between states that benefit from explicit choreography.

### Tier 4: Once-Ever Interactions (onboarding, first-run)

Examples: account creation, onboarding flows, first feature discovery, empty states before any data.

**Motion prescription**: Full novelty. Up to 100% novelty budget. Impression-defining.
- Transitions: variable (may exceed 200ms for narrative motion)
- Custom spring curves, keyframe choreography, complex stagger
- Can use animation for storytelling and teaching
- Reduced-motion: still provide narrative through layout and content

**Rationale**: Once-ever interactions shape first impressions and teach the product. Animation investment is fully justified. Users will experience this exactly once -- make it count.

## The 10% Novelty Budget

Within any Tier 3 or Tier 4 interaction, apply the 10% budget:
- 90% of motion vocabulary should use familiar patterns (spring physics, standard easing curves, conventional direction)
- 10% may be genuinely novel (custom curves, unexpected spatial metaphors, branded motion personality)

**Budget allocation process**:
1. Inventory all Tier 3/4 interactions in the scope
2. Rank by business importance (checkout > settings > profile)
3. Allocate the 10% novel motion to the highest-ranked interactions
4. Remaining interactions use familiar patterns only

## Classification Worksheet

For each interaction being classified:

```
Interaction name: _______________
Usage frequency estimate: ___ times/day
Frequency tier: Tier 1 / Tier 2 / Tier 3 / Tier 4
Novelty allocation: ___% (sum across Tier 3/4 should not exceed 10% of total)
Prescribed timing ceiling: ___ms
Spring physics: Yes / No
Stagger choreography: Yes / No
Reduced-motion spec: _______________
```

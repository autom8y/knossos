# strategy TODO

> Audit conducted: 2026-01-02 | Status: MATURE, coordination improvements needed

## Current State Summary

The strategy rite is a **well-designed, methodologically rigorous rite** with clear 4-phase workflow (market-research → competitive-analysis → business-modeling → strategic-planning) and excellent role separation.

**Strengths confirmed:**
- Exceptional role clarity (market vs competitive vs financial vs roadmap)
- Strong methodological requirements (TAM/SAM/SOM, threat frameworks, RICE scoring)
- Concrete acid tests ("Would an investor believe this?" "Would a CFO trust this?")
- Anti-pattern documentation prevents common mistakes
- All agents appropriately assigned to opus for reasoning-heavy work
- Proper coach pattern for orchestrator (consultative, not executing)
- Sequential workflow with logical phase dependencies
- Back-routes for failure recovery (partial)

**Issues identified:**
- Model assignment discrepancy in skill.md (says sonnet, agents use opus)
- Intelligence vs strategy boundary not documented in README
- R&D integration undefined (moonshot-architect → roadmap-strategist)
- Back-routes incomplete (missing market-researcher reverse flows)
- 10x handoff format not explicitly defined

---

## Validated Improvements

### P0: Fix Model Assignment Bug

**Gap identified:** skill.md incorrectly lists market-researcher as "sonnet" but agent definition specifies "opus".

**Decision:** Fix data consistency bug. All strategy agents correctly use opus.

**Changes required:**
- [ ] Update `skills/strategy-ref/skill.md`: Change market-researcher model from "sonnet" to "opus"
- [ ] Verify all other model references in skill.md match agent definitions

---

### P1: Add Intelligence vs Strategy Boundary Clarification

**Gap identified:** Both teams do "analysis" but serve different purposes. Boundary documented in intelligence TODO P3 but not in strategy README.

**Decision:** Coordinate with intelligence P3 to add comparison table to both READMEs.

**Changes required:**
- [ ] Update `rites/strategy/README.md`: Add "When to Use Strategy vs Intelligence" section
- [ ] Coordinate with intelligence P3 (mirror same table)
- [ ] Clarify:
  ```markdown
  ## When to Use Strategy vs Intelligence

  | Question Type | Team | Examples |
  |---------------|------|----------|
  | What's the market opportunity? | strategy | Market sizing, TAM/SAM/SOM analysis |
  | Who are our competitors? | strategy | Competitive intelligence, threat assessment |
  | What should our roadmap be? | strategy | Strategic planning, prioritization |
  | How do our users behave? | intelligence | Funnel analysis, user research, A/B tests |
  | Why do users do X? | intelligence | Qualitative research, session analysis |
  | Does this feature work? | intelligence | Experiment design, statistical validation |

  **Rule of thumb**:
  - Strategy = outward (market, competitors, business model)
  - Intelligence = inward (our users, our product)

  **When both needed**: "Competitors are winning on mobile" requires:
  - Strategy: Competitive positioning analysis
  - Intelligence: Our mobile UX research
  - Synthesis: Combined insight for roadmap
  ```

---

### P2: Define R&D → Strategy Integration

**Gap identified:** The rnd rite's moonshot-architect produces long-term architecture plans but no pathway to roadmap-strategist exists.

**Decision:** Define explicit handoff workflow from moonshot-architect to roadmap-strategist.

**Changes required:**
- [ ] Update `rites/strategy/README.md`: Add "Architecture Strategy Integration" section
- [ ] Update `rites/rnd/README.md`: Reference strategy as consumer of moonshot outputs
- [ ] Define integration pattern:
  ```markdown
  ## Architecture Strategy Integration

  When rnd's moonshot-architect produces long-term architecture recommendations:

  **Trigger**: MOONSHOT-*.md artifact produced with business impact > "significant"

  **Workflow**:
  1. moonshot-architect produces MOONSHOT-*.md with business implications
  2. User/orchestrator routes to strategy for business case validation
  3. roadmap-strategist evaluates: Does architecture change merit strategic initiative?
  4. If yes: Creates STRATEGY-*.md with resource/timeline/priority
  5. Routes to 10x-dev for implementation

  **Handoff artifact**:
  ```yaml
  source_team: rnd
  target_team: strategy
  handoff_type: strategic_evaluation
  context:
    initiative: "Event-Driven Architecture Migration"
    source_artifacts:
      - docs/rnd/MOONSHOT-event-driven.md
  items:
    - id: STRAT-EVAL-001
      summary: "Evaluate business case for architecture migration"
      technical_findings:
        - 3x throughput improvement projected
        - 18-month migration timeline
      questions_for_strategy:
        - Is this competitive necessity or nice-to-have?
        - What's the ROI horizon?
        - Should this be a strategic initiative?
  ```

---

### P3: Add Missing Back-Routes

**Gap identified:** Back-routes from competitive-analyst → market-researcher and business-modeling → market-researcher are missing. If downstream phases invalidate upstream assumptions, no recovery path exists.

**Decision:** Add complete back-route coverage.

**Changes required:**
- [ ] Update `rites/strategy/workflow.yaml`: Add missing back-routes
- [ ] Add routes:
  ```yaml
  back_routes:
    # Existing (keep)
    - from: strategic-planning
      to: market-researcher
      trigger: requirements_gap
      requires_confirmation: true
    - from: strategic-planning
      to: business-model-analyst
      trigger: design_flaw
      requires_confirmation: false

    # NEW: Competitive invalidates market assumptions
    - from: competitive-analyst
      to: market-researcher
      trigger: market_assumption_invalid
      description: "Competitive analysis reveals market is smaller/consolidating than sized"
      requires_confirmation: true

    # NEW: Financial modeling invalidates market assumptions
    - from: business-model-analyst
      to: market-researcher
      trigger: economics_unsustainable
      description: "Financial modeling shows market economics don't support assumptions"
      requires_confirmation: true

    # NEW: Financial modeling invalidates competitive assumptions
    - from: business-model-analyst
      to: competitive-analyst
      trigger: competitive_assumption_invalid
      description: "Pricing/positioning assumptions don't match competitive reality"
      requires_confirmation: false
  ```
- [ ] Document in orchestrator guidance: When to trigger each back-route

---

### P4: Strategy Outputs Use Generalized Handoff Pattern

**Decision:** Strategy → 10x handoff should use the same HANDOFF artifact pattern as other cross-rite transitions.

**Changes required:**
- [ ] Document strategy handoffs in generalized pattern (ecosystem-level)
- [ ] Add to pattern documentation:
  ```yaml
  # Example: strategy → 10x handoff (strategic initiative ready for implementation)
  source_team: strategy
  target_team: 10x-dev
  handoff_type: implementation
  context:
    initiative: "Mobile-First Checkout Redesign"
    source_artifacts:
      - docs/strategy/STRATEGY-mobile-checkout.md
      - docs/strategy/MARKET-checkout-opportunity.md
  items:
    - id: IMPL-001
      summary: "Implement mobile-first checkout flow"
      strategic_context:
        priority: P1 (RICE score: 8.4)
        investment: 3 engineer-months
        expected_impact: +15% mobile conversion
      acceptance_criteria:
        - Mobile checkout completion < 60 seconds
        - Conversion rate measurable via A/B test
        - Shipping estimate visible on product page
      success_metrics:
        - Mobile conversion +10-15%
        - Cart abandonment -20%
  ```
- [ ] Update roadmap-strategist handoff criteria to produce HANDOFF artifact

---

## Deferred / Not Prioritized

### Hooks Implementation
**Decision:** Defer. Strategy work is less frequent than development work. Hook overhead not justified.

### MCP Integration for Market Data
**Decision:** Future work. Market databases, competitor APIs, industry reports could be valuable but not blocking current workflow.

### Phase Duration SLAs
**Decision:** Skip. Strategic work has inherently variable timelines. Adding SLAs would add bureaucracy without value.

### Artifact Versioning
**Decision:** Defer. Handle via git versioning. Formal artifact versioning strategy is over-engineering.

---

## Dependencies

| Item | Depends On |
|------|------------|
| P1 (intelligence boundary) | Coordinate with intelligence TODO P3 |
| P2 (R&D integration) | Update to rnd README |
| P4 (handoff pattern) | Generalized handoff pattern from debt-triage TODO |

---

## Cross-Rite Notes

**For intelligence:** P1 coordinates with intelligence TODO P3. Both teams add the same comparison table to their READMEs.

**For rnd:** P2 requires updating rnd README to reference strategy as consumer of moonshot outputs.

**For ecosystem:** P4 adds strategy→10x to the generalized handoff pattern.

---

## Summary of All Teams Audited

This completes the 10-team audit:
1. 10x-dev (mature, impact assessment + flexible entry needed)
2. ecosystem (demote from hub, remove satellite testing)
3. docs (add proactive gate, staleness detection)
4. hygiene (define behavior preservation, formalize debt handoff)
5. debt-triage (shared detection skill, generalized handoff)
6. sre (shared templates, 10x→sre handoff)
7. security (proactive threat modeling, risk acceptance)
8. intelligence (fix agent quality, boundary clarification, MCP plan)
9. rnd (clarify spike overlap)
10. strategy (model fix, R&D integration, back-routes)

**Cross-cutting patterns identified:**
- Generalized HANDOFF artifact (applies to 6+ team pairs)
- Shared templates at ecosystem level
- Shared smell detection skill
- Intelligence vs Strategy boundary
- Proactive gates (security for SYSTEM, docs for user-facing)

# intelligence TODO

> Audit conducted: 2026-01-02 | Status: FUNCTIONAL, quality improvements needed

## Current State Summary

The intelligence rite is a **well-structured analytics and research rite** with clear 4-phase workflow (instrumentation → research → experimentation → synthesis) and complexity-based phase skipping.

**Strengths confirmed:**
- Clear agent specialization (analytics, research, experimentation, synthesis)
- Sequential workflow with proper handoff criteria
- Strong Pythia (25/30 score)
- Complexity-based phase skipping (METRIC skips front phases)
- Emphasis on statistical rigor and pre-registration

**Issues identified:**
- user-researcher (19/30) and insights-analyst (19/30) have lowest quality scores
- No direct data access (by design, but MCP integration planned)
- Intelligence vs strategy boundary needs clarification

---

## Validated Improvements

### P1: Fix User-Researcher Agent Quality

**Gap identified:** Lowest quality score (19/30) - casual tone, missing examples, generic role definition.

**Decision:** Priority fix needed. User research is core team value; low quality undermines credibility.

**Changes required:**
- [ ] Update `agents/user-researcher.md`:
  - Standardize opening tone (remove casual first-person language)
  - Add concrete example of research finding with quotes and confidence rating
  - Sharpen role definition: "Captures qualitative 'why' behind user behavior patterns"
  - Add "When Invoked (First Actions)" numbered list
  - Reduce boilerplate, reference skills instead of inline templates
- [ ] Target: 23+ /30 score after rewrite

**Example to add:**
```markdown
### Finding: Users abandon checkout when shipping costs appear late

**Confidence**: High (5/6 participants)

**Evidence**:
> "I was ready to buy, but when I saw $12 shipping at the last step, I just closed the tab." — P03
> "Why don't they show shipping earlier? I feel tricked." — P05

**Quant connection**: Tracking shows 34% drop-off at shipping step
**Recommendation**: Display shipping estimate on product page
```

---

### P2: Fix Insights-Analyst Agent Quality

**Gap identified:** Tied lowest quality score (19/30) - generic role definition, missing concrete examples.

**Decision:** Priority fix needed. Synthesis is the terminal phase; weak insights-analyst undermines entire workflow.

**Changes required:**
- [ ] Update `agents/insights-analyst.md`:
  - Sharpen role definition: "Synthesizes multi-source data into decision-ready recommendations"
  - Add concrete example of insight with statistical evidence and segment analysis
  - Add example of Impact/Confidence rating justification
  - Reduce generic sections, add specificity
- [ ] Target: 23+ /30 score after rewrite

**Example to add:**
```markdown
### Finding: New checkout flow increases conversion by 8.2%

**Impact**: High | **Confidence**: High

**Evidence**:
- Conversion: 12.1% → 13.1% (+8.2%, 95% CI: [5.1%, 11.4%])
- p-value: 0.003, n=24,000
- Effect consistent across 14-day test period

**Segment Analysis**:
| Segment | Effect | Notes |
|---------|--------|-------|
| Mobile | +11.3% | Strongest effect |
| Desktop | +5.1% | Moderate effect |

**Recommendation**: SHIP to 100%. Priority: Mobile users.
```

---

### P3: Clarify Intelligence vs Strategy Boundary

**Gap identified:** Both rites do "analysis" but different kinds. Boundary needs explicit documentation.

**Decision:** Add explicit comparison to both rite READMEs.

**Changes required:**
- [ ] Update `rites/intelligence/README.md`: Add "When to Use Intelligence vs Strategy" section
- [ ] Update `rites/strategy/README.md`: Add same section (mirror content)
- [ ] Clarify:
  ```markdown
  ## When to Use Intelligence vs Strategy

  | Question Type | Rite | Examples |
  |---------------|------|----------|
  | How do our users behave? | intelligence | Funnel analysis, user research, A/B tests |
  | Why do users do X? | intelligence | Qualitative research, session analysis |
  | Does this feature work? | intelligence | Experiment design, statistical validation |
  | What's the market opportunity? | strategy | Market sizing, competitive analysis |
  | What should our roadmap be? | strategy | Strategic planning, prioritization |
  | Who are our competitors? | strategy | Competitive intelligence |

  **Rule of thumb**:
  - Intelligence = inward (our users, our product)
  - Strategy = outward (market, competitors, business model)
  ```

---

### P4: Intelligence Outputs Use Generalized Handoff Pattern

**Decision:** Intelligence→10x and intelligence→strategy should use same HANDOFF artifact pattern.

**Changes required:**
- [ ] Document intelligence handoffs in generalized pattern (ecosystem-level)
- [ ] Add to pattern documentation:
  ```yaml
  # Example: intelligence → 10x handoff (experiment results driving implementation)
  source_rite: intelligence
  target_rite: 10x-dev
  handoff_type: implementation
  context:
    initiative: "Checkout Flow Optimization"
    source_artifacts:
      - .ledge/spikes/EXPERIMENT-checkout-shipping.md
      - .ledge/spikes/INSIGHT-checkout-shipping.md
  items:
    - id: IMPL-001
      summary: "Implement shipping estimate on product page"
      evidence: "A/B test showed +8.2% conversion lift"
      acceptance_criteria:
        - Shipping estimate displays on all product pages
        - Estimate accuracy within $2 of actual

  # Example: intelligence → strategy handoff (insights informing roadmap)
  source_rite: intelligence
  target_rite: strategy
  handoff_type: strategic_input
  context:
    initiative: "Q2 Roadmap Planning"
    source_artifacts:
      - .ledge/spikes/INSIGHT-user-behavior-q1.md
  items:
    - id: STRAT-001
      summary: "Mobile checkout friction identified as top opportunity"
      evidence: "Mobile conversion 40% lower than desktop; research shows UX issues"
  ```

---

### P5: Plan MCP Integration with Data Service

**Decision:** Future work to give agents actual data access via MCP integration with data service.

**Changes required:**
- [ ] Document planned MCP integration in intelligence README
- [ ] Define which agents get data access:
  - analytics-engineer: Query event data, validate tracking
  - insights-analyst: Run analytical queries, verify statistical claims
  - user-researcher: Access session recordings, behavioral data
  - experimentation-lead: Query experiment results
- [ ] Coordinate with data-service satellite MCP implementation
- [ ] Add MCP tools to agent tool lists when ready

**Note:** MCP groundwork already laid in data-service satellite. This is a future enhancement, not blocking current workflow.

---

## Deferred / Not Prioritized

### Merge with Strategy
**Decision:** Keep separate. Intelligence (product analytics) and strategy (business analysis) are different disciplines with different workflows.

### Deprecate User-Researcher
**Decision:** Keep. Qualitative research complements quantitative analytics. Fix quality instead of removing.

### Add QA Phase
**Decision:** Not needed. Handoff criteria and insights-analyst acid test provide sufficient validation.

---

## Dependencies

| Item | Depends On |
|------|------------|
| P3 (boundary clarification) | Requires update to strategy README (coordinate with strategy audit) |
| P4 (handoff pattern) | Generalized handoff pattern from debt-triage TODO |
| P5 (MCP integration) | data-service satellite MCP implementation |

---

## Cross-Rite Notes

**For strategy:** P3 requires adding "Intelligence vs Strategy" comparison to strategy README.

**For ecosystem:** P4 adds intelligence→10x and intelligence→strategy to generalized handoff pattern.

**For data-service satellite:** P5 depends on MCP server implementation in that service.

---

## Next Rite

Continue audit with: **rnd** (exploration and prototyping)

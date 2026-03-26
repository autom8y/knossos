---
description: "Insights Artifacts companion for intelligence-ref skill."
---

# Insights Artifacts

> HANDOFF templates, findings format, and insights report guidance for the Insights Analyst.

## HANDOFF Production

When insights require action by another rite, produce a HANDOFF artifact using the `cross-rite-handoff` schema.

### Target Rite Routing

| Insight Type | Target Rite | Handoff Type | Example |
|--------------|-------------|--------------|---------|
| User-driven feature opportunity | 10x-dev | implementation | "Users abandoning at checkout due to missing guest option" |
| Strategic pattern or trend | strategy | strategic_input | "Mobile users convert 40% less despite 2x browsing" |
| Both actionable AND strategic | Both (separate HANDOFFs) | implementation + strategic_input | Major insight with immediate fix and long-term implications |

### Decision Criteria for Target Selection

Route to **10x-dev** when:
- Insight points to specific, implementable improvement
- User research identifies concrete feature gap
- A/B test winner is ready for full rollout
- Recommendation is "build X" with clear acceptance criteria

Route to **strategy** when:
- Insight reveals market trend or competitive pattern
- Data suggests strategic pivot or new market opportunity
- Findings inform roadmap prioritization decisions
- Recommendation is "consider X for Q2 planning"

### HANDOFF Example (to 10x-dev)

```yaml
---
source_rite: intelligence
target_rite: 10x-dev
handoff_type: implementation
created: 2026-01-02
initiative: Checkout Optimization
priority: high
---

## Context

User research and A/B testing confirm that address autocomplete reduces checkout abandonment by 60%. Ready for production implementation.

## Source Artifacts
- .ledge/spikes/INSIGHTS-checkout-friction-Q1.md
- .ledge/spikes/AB-RESULTS-address-autocomplete.md

## Items

### IMP-001: Address autocomplete implementation
- **Priority**: High
- **Summary**: Add address autocomplete to checkout flow
- **Evidence**: 60% reduction in address-entry abandonment (n=10K, p<0.001)
- **Acceptance Criteria**:
  - Google Places API integration
  - Works on mobile and desktop
  - Graceful fallback when API unavailable
  - Maintains current form validation

## Notes for Target Rite

Mobile users showed 70% improvement vs desktop 45%--consider mobile-first implementation.
```

### HANDOFF Example (to strategy)

```yaml
---
source_rite: intelligence
target_rite: strategy
handoff_type: strategic_input
created: 2026-01-02
initiative: Q2 Product Planning
priority: medium
---

## Context

Cross-platform analysis reveals significant mobile conversion gap despite higher engagement.

## Source Artifacts
- .ledge/spikes/INSIGHTS-platform-behavior-Q1.md

## Items

### INS-001: Mobile conversion opportunity
- **Priority**: High
- **Summary**: Mobile users browse 2x more but convert 40% less than desktop
- **Data Sources**: Analytics (n=100K), heatmaps, session recordings
- **Confidence**: Medium (limited qualitative data on root cause)
- **Strategic Implication**: Mobile optimization may be higher-ROI than new feature development

## Notes for Target Rite

Recommend prioritizing mobile UX research before Q2 roadmap finalization.
```

## Insights Report Guidance

Produce Insights Report using doc-intelligence skill, insights-report-template section.

**Required elements**:
- Executive summary: 3-5 sentences with key recommendation
- Each finding rated by Impact (High/Medium/Low) AND Confidence (High/Medium/Low)
- Segment analysis comparing subgroup effects to overall
- Alternative explanations section: what else could explain these results?
- Limitations section: what can't we conclude from this data?
- Recommendation with both "ship" and "don't ship" contingency plans

### Example Finding Format

```markdown
### Finding 1: New checkout flow increases conversion by 8.2%

**Impact**: High | **Confidence**: High

**Statistical Evidence**:
- Conversion: 12.1% -> 13.1% (+8.2%, 95% CI: [5.1%, 11.4%])
- p-value: 0.003, n=24,000
- Effect consistent across 14-day test period

**Qualitative Support**:
- User research found shipping cost transparency reduced abandonment (P03, P05)
- Session recordings show 40% reduction in back-button clicks at checkout

**Segment Analysis**:
| Segment | Effect | Notes |
|---------|--------|-------|
| Mobile | +11.3% | Strongest effect |
| Desktop | +5.1% | Moderate effect |
| New users | +14.2% | Primary beneficiary |
| Returning | +2.8% | Minimal change |

**Alternative Explanations Ruled Out**:
- Novelty effect: Effect stable across 14 days
- Selection bias: Random assignment verified

**Recommendation**: SHIP to 100% traffic. Priority: Mobile users. Monitor returning user conversion for potential regression.
```

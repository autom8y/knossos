# rnd TODO

> Audit conducted: 2026-01-02 | Status: FUNCTIONAL, integration clarifications needed

## Current State Summary

The rnd rite is a **well-structured exploration and prototyping rite** with clear 4-phase workflow (diverge → prototype → converge → productionize) and appropriate complexity-based phase skipping.

**Strengths confirmed:**
- Clear agent specialization (explorer, prototype-builder, evaluator, tech-transfer)
- Sequential workflow with explicit handoff criteria
- Strong emphasis on time-boxing and learning focus
- Prototype vs production distinction well-articulated
- Back-routes for iteration (prototype → diverge when direction unclear)
- Appropriate model assignments (Opus for exploration, Sonnet for implementation)

**Issues identified:**
- References non-existent "ship-pack" for productionization handoff
- SPIKE overlap with 10x-dev's /spike command unclear
- No formal handoff artifact when transferring to production teams

---

## Validated Improvements

### P1: Update Productionization Handoff to 10x-dev

**Gap identified:** README and tech-transfer agent reference "ship-pack" which doesn't exist in the ecosystem.

**Decision:** Update all references to use 10x-dev for productionization handoff.

**Changes required:**
- [x] Update `rites/rnd/README.md`: Replace "ship-pack" references with "10x-dev"
- [x] Clarify handoff criteria: Tech-transfer produces TRANSFER artifacts → 10x requirements-analyst consumes as input
- [x] Create `agents/tech-transfer.md`: Production readiness specialist with HANDOFF production to 10x-dev and strategy

**Workflow after change:**
```
RND exploration complete → tech-transfer produces TRANSFER doc → 10x-dev picks up via requirements-analyst
```

---

### P2: Clarify SPIKE vs /spike Distinction

**Gap identified:** Both rnd and 10x-dev have "spike" concepts. Boundary unclear.

**Decision:** Add explicit guidance in both teams distinguishing the scopes.

**Changes required:**
- [ ] Update `rites/rnd/README.md`: Add "When to Use RND vs 10x /spike" section
- [ ] Update `rites/10x-dev/README.md` or spike skill: Add same clarification
- [ ] Clarify:
  ```markdown
  ## When to Use RND vs 10x /spike

  | Scenario | Team | Examples |
  |----------|------|----------|
  | "Can we build X at all?" | rnd | Novel algorithms, unproven approaches |
  | "Which library for X?" | 10x /spike | Choosing between established options |
  | "Does approach X scale?" | rnd | Performance at 100x current load |
  | "How long will X take?" | 10x /spike | Effort estimation for known patterns |
  | "Should we use AI for X?" | rnd | Exploring ML feasibility |
  | "React vs Vue for X?" | 10x /spike | Comparing known frameworks |

  **Rule of thumb**:
  - RND = exploring the unknown (multiple sessions, learning-focused)
  - 10x /spike = evaluating known options (time-boxed, decision-focused)
  ```

---

### P3: RND Outputs Use Generalized Handoff Pattern

**Decision:** RND→10x and RND→strategy handoffs should use the same HANDOFF artifact pattern as other cross-rite transitions.

**Changes required:**
- [ ] Document RND handoffs in generalized pattern (ecosystem-level)
- [ ] Add to pattern documentation:
  ```yaml
  # Example: rnd → 10x handoff (prototype ready for productionization)
  source_team: rnd
  target_team: 10x-dev
  handoff_type: productionization
  context:
    initiative: "ML-Powered Search"
    source_artifacts:
      - docs/rnd/EXPLORATION-ml-search.md
      - docs/rnd/PROTOTYPE-ml-search.md
      - docs/rnd/EVALUATION-ml-search.md
  items:
    - id: PROD-001
      summary: "Productionize ML search prototype"
      findings: "Prototype achieved 85% relevance improvement"
      constraints:
        - Model size must stay under 500MB
        - Inference latency < 100ms p99
      acceptance_criteria:
        - Production-grade error handling
        - Monitoring and alerting
        - Graceful degradation to keyword search

  # Example: rnd → strategy handoff (exploration informing roadmap)
  source_team: rnd
  target_team: strategy
  handoff_type: strategic_input
  context:
    initiative: "Technology Radar Update"
    source_artifacts:
      - docs/rnd/EVALUATION-emerging-tech-q1.md
  items:
    - id: STRAT-001
      summary: "WebAssembly ready for production use"
      evidence: "Prototype showed 3x performance improvement in compute-heavy operations"
  ```
- [x] Update tech-transfer agent to produce HANDOFF artifact as output

---

## Deferred / Not Prioritized

### Hard Prototype Gate
**Decision:** Documentation is sufficient. Engineers are professionals. Trust them to follow guidelines; evaluator catches issues in review. Adding hard enforcement would slow exploration velocity.

### Merge with 10x /spike
**Decision:** Keep separate. RND is multi-session exploration of unknowns; /spike is time-boxed evaluation of known options. Different concerns, different workflows.

### Separate Ship Rite
**Decision:** Don't create. 10x-dev already handles production implementation. RND's tech-transfer provides the bridge.

---

## Dependencies

| Item | Depends On |
|------|------------|
| P2 (spike clarification) | Update to 10x-dev README or spike skill |
| P3 (handoff pattern) | Generalized handoff pattern from debt-triage TODO |

---

## Cross-Rite Notes

**For 10x-dev:** P2 requires adding "RND vs /spike" clarification to 10x documentation.

**For ecosystem:** P3 adds RND→10x and RND→strategy to the generalized handoff pattern.

---

## Next Rite

Continue audit with: **strategy** (business and market analysis)

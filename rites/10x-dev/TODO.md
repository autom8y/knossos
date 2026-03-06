# 10x-dev TODO

> Audit conducted: 2026-01-02 | Status: MATURE, targeted improvements identified

## Current State Summary

The 10x-dev is a **production-ready, well-documented rite** implementing a 5-agent sequential workflow (PRD → TDD → Code → QA) with mandatory quality gates.

**Strengths confirmed:**
- Clear agent separation of concerns with defined domain authority
- Comprehensive artifact templates (PRD, TDD, ADR, Test Plan)
- Quality gates between all phases prevent low-quality propagation
- Adaptive back-routing for discovered gaps
- Four complexity levels (SCRIPT/MODULE/SERVICE/PLATFORM) with phase adaptation

**Architecture validated:**
- Pythia as stateless advisor (read-only) - correct design
- Acid tests as guidance, not enforcement - avoids checkbox theater
- QA authority is advisory only - humans decide, QA documents risks

---

## Validated Improvements

### P1: Add Impact Assessment to Requirements Phase

**Gap identified:** SCRIPT complexity (<200 LOC) skips design, but small changes can be architecturally significant (e.g., 50-line auth change).

**Solution:** Requirements Analyst flags `high-impact` attribute regardless of LOC. High-impact work routes to Architect even at SCRIPT level.

**Changes required:**
- [ ] Update `agents/requirements-analyst.md`: Add impact assessment to PRD output
- [ ] Update `workflow.yaml`: Add conditional routing based on `impact: high` flag
- [ ] Update `skills/doc-artifacts/SKILL.md`: Add impact field to PRD template
- [ ] Update `skills/10x-workflow/quality-gates.md`: Document impact assessment in PRD gate

**Impact categories to assess:**
- Security surface changes
- Data model modifications
- API contract changes
- Authentication/authorization changes
- Cross-service dependencies

---

### P2: Flexible Entry Points via Pythia

**Gap identified:** Workflow assumes PRD-first universally, but technical refactoring and performance optimization naturally start with TDD.

**Solution:** Relax PRD-first requirement. Let Pythia decide entry point based on work type. This trusts Pythia's core competency.

**Changes required:**
- [ ] Update `agents/pythia.md`: Add guidance for selecting entry agent based on work type
- [ ] Update `workflow.yaml`: Remove hard-coded `entry_point: requirements-analyst`, make dynamic
- [ ] Update `skills/10x-workflow/SKILL.md`: Document entry point flexibility
- [ ] Consider: Add `--entry` flag to `/sos start` for explicit override when user knows best entry

**Entry point heuristics for Pythia:**
| Work Type | Recommended Entry | Rationale |
|-----------|------------------|-----------|
| New feature | Requirements Analyst | User stories drive scope |
| Technical refactoring | Architect | Constraints are technical, not user-facing |
| Performance optimization | Architect | Design decisions drive approach |
| Bug fix | Principal Engineer | Known issue, implementation-focused |
| Security fix | Principal Engineer | Urgent, scope is clear |

---

### P3: Strengthen Cross-Rite Handoff Protocols

**Gap identified:** Audit documented upstream/downstream relationships but no formal handoff protocols exist. The `/handoff` skill exists but cross-rite handoffs need specification.

**Changes required:**
- [ ] Create `skills/cross-rite-handoff/SKILL.md`: Formal protocols for handoffs to other rites
- [ ] Document handoff artifacts expected by receiving rites:
  - To SRE rite: deployment manifest, runbook draft, observability requirements
  - To Security rite: threat model, security-relevant code paths, auth flows
  - To Docs rite: feature summary, API changes, user-facing behavior changes
- [ ] Add cross-rite handoff section to `README.md`

---

## Deferred / Not Prioritized

### Rite-Specific Hooks
**Decision:** Undecided - will revisit after auditing other rites to see patterns. Currently using project-level hooks which may be sufficient.

### Acid Test Enforcement
**Decision:** Keep as guidance only. Enforcement would create checkbox theater and slow velocity without improving quality. Acid tests work as mindset framing.

### QA Blocking Authority
**Decision:** Advisory only. QA documents risks and recommendations; humans decide. Enforcement would create adversarial dynamics.

---

## Dependencies

| Item | Depends On |
|------|------------|
| Cross-rite handoffs | Audit of receiving rites (SRE, Security, Doc) to validate expected artifacts |
| Rite hooks decision | Pattern analysis across all rite audits |

---

## Next Rite

Continue audit with: **ecosystem** (infrastructure, hooks, sync pipeline, knossos patterns)

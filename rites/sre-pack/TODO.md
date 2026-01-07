# sre-pack TODO

> Audit conducted: 2026-01-02 | Status: MATURE, recently optimized, minor ecosystem alignment

## Current State Summary

SRE-pack is a **production-ready, well-optimized team** (44% token reduction achieved in Dec 2025) with clear agent roles and comprehensive templates.

**Strengths confirmed:**
- 4-phase workflow (observation → coordination → implementation → resilience)
- All agents under 200 lines (stretch goal achieved)
- Comprehensive doc-sre templates (9 templates)
- Clear separation from other teams (10x builds, sre validates)
- Chaos engineer blast radius controls are well-documented
- Orchestrator is properly consultative (read-only)

**Architecture validated:**
- Runbook ownership clear - SRE owns operational docs, doc-team owns external user docs
- Chaos blast radius progression is documentation-sufficient (trust + review)
- Team is genuinely holistic reliability, not just infrastructure ops

---

## Validated Improvements

### P1: Move Shared Templates to Ecosystem-Level Skill

**Decision:** doc-sre templates used by multiple teams (debt ledger, risk matrix, sprint debt packages) should move to ecosystem-level shared skill.

**Rationale:** Templates used by debt-triage-pack, hygiene-pack, and sre-pack shouldn't be owned by one team. Ecosystem-level templates prevent cross-rite ownership confusion.

**Changes required:**
- [ ] Create `skills/shared-templates/` at ecosystem level (or extend existing skill)
- [ ] Move from doc-sre to shared:
  - `debt-ledger-template` (used by debt-triage-pack)
  - `risk-matrix-template` (used by debt-triage-pack)
  - `sprint-debt-packages-template` (used by debt-triage-pack)
- [ ] Keep in doc-sre (sre-specific):
  - `observability-report-template`
  - `reliability-plan-template`
  - `postmortem-template`
  - `chaos-experiment-template`
  - `resilience-report-template`
  - `infrastructure-change-template`
  - `tracking-plan-template`
- [ ] Update references in debt-triage-pack agents to new location
- [ ] Update doc-sre SKILL.md to reflect reduced scope

**Template ownership after change:**
| Template | Location | Primary Users |
|----------|----------|---------------|
| debt-ledger | shared-templates | debt-triage, hygiene |
| risk-matrix | shared-templates | debt-triage |
| sprint-debt-packages | shared-templates | debt-triage |
| observability-report | doc-sre | sre-pack |
| reliability-plan | doc-sre | sre-pack |
| postmortem | doc-sre | sre-pack |
| chaos-experiment | doc-sre | sre-pack |
| resilience-report | doc-sre | sre-pack |

---

### P2: 10x→SRE Handoff Uses Generalized Pattern

**Decision:** When 10x-dev-pack completes implementation and needs resilience validation, use the generalized HANDOFF artifact pattern (same as debt→hygiene).

**Rationale:** Consistent cross-rite handoff pattern across ecosystem. 10x→sre is a handoff (code ready for validation), not just "SRE pulls when ready."

**Changes required:**
- [ ] Document 10x→sre handoff in generalized pattern (P2 from debt-triage TODO)
- [ ] Add to pattern documentation:
  ```yaml
  # Example: 10x → sre handoff
  source_team: 10x-dev-pack
  target_team: sre-pack
  handoff_type: validation
  context:
    initiative: "OAuth2 Implementation"
    source_artifacts:
      - docs/design/TDD-oauth2.md
      - src/auth/oauth2/
  items:
    - id: VAL-001
      summary: "Validate OAuth2 resilience under load"
      acceptance_criteria:
        - Service handles 10x normal traffic
        - Graceful degradation when IdP unavailable
        - Token refresh survives network partition
  ```
- [ ] Update sre-pack README with handoff acceptance guidance
- [ ] Update 10x-dev-pack QA-adversary to mention sre handoff for SERVICE+ complexity

---

## Deferred / Not Prioritized

### Chaos Blast Radius Enforcement
**Decision:** Documentation is sufficient. Engineers are professionals. Catch violations in review, don't add hard gates.

### Runbook Ownership Clarification
**Decision:** Boundary is already clear. Runbooks = operational procedures for engineers. User docs = external users. No changes needed.

---

## Dependencies

| Item | Depends On |
|------|------------|
| P1 (shared templates) | Ecosystem-level skill creation |
| P2 (10x handoff) | Generalized handoff pattern from debt-triage TODO |

---

## Cross-Team Notes

**For ecosystem-pack:** P1 requires creating or extending a shared-templates skill at ecosystem level.

**For debt-triage-pack:** After P1, update template references from `@doc-sre#*` to `@shared-templates#*`.

**For 10x-dev-pack:** P2 adds sre validation as a handoff option after QA approval for SERVICE+ complexity.

---

## Next Team

Continue audit with: **security-pack** (security assessment)

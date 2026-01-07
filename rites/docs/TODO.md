# docs TODO

> Audit conducted: 2026-01-02 | Status: MATURE, targeted integration improvements

## Current State Summary

Doc-rite-pack is a **well-designed, production-ready documentation team** with clear phase separation (audit → architecture → writing → review) and comprehensive quality gates.

**Strengths confirmed:**
- Clear 4-phase pipeline with explicit handoff criteria
- Complexity-based workflow (PAGE/SECTION/SITE)
- Strong doc-consolidation skill for reducing token overhead
- Generalist model is correct - domain teams own specialized docs
- Optimized agents (23-25/30 scores, 19% token reduction applied)

**Architecture validated:**
- Reactive-only model is correct - docs should reflect actual implementation
- Generalist tech-writer, not domain variants - domain teams handle their specialties
- Single tech-writer with WebFetch/WebSearch for research is sufficient

---

## Validated Improvements

### P1: Add Documentation Gate to 10x Workflow

**Gap identified:** No mechanism triggers docs after 10x completes. Handoff is documented but not enforced - engineers manually invoke `/docs`.

**Solution:** Add "documentation plan" as a checklist item in 10x QA-adversary's release recommendation.

**Changes required:**
- [ ] Update `rites/10x-dev/agents/qa-adversary.md`: Add documentation assessment to release checklist
- [ ] Add criteria: "User-facing changes documented or doc-team handoff planned"
- [ ] QA can mark "docs not needed" for internal-only changes, but must explicitly decide

**Gate questions for QA:**
- Does this change affect user-facing behavior?
- Is existing documentation still accurate?
- Should docs be notified? (Yes → include in handoff notes)

---

### P2: Add Staleness Detection to Doc-Auditor

**Gap identified:** Doc-reviewer validates once, but code keeps changing. No continuous validation - docs silently go stale.

**Solution:** Enhance doc-auditor with git-based staleness detection capability.

**Changes required:**
- [ ] Update `agents/doc-auditor.md`: Add staleness detection mode
- [ ] Add approach step: "Cross-reference doc modification dates with related code changes"
- [ ] Add heuristic: If code file changed since doc last touched, flag as "potentially stale"
- [ ] Consider: Add `/doc-audit --staleness` command for on-demand staleness check

**Staleness detection logic:**
```
For each doc file:
  1. Identify related code files (via imports, references, directory proximity)
  2. Compare doc last-modified vs code last-modified
  3. If code newer by > threshold, flag for review
  4. Output: List of potentially stale docs with evidence
```

---

## Deferred / Not Prioritized

### Proactive Documentation Mode
**Decision:** Reactive-only is correct. Docs should reflect actual implementation, not aspirational designs. Documentation lag is acceptable.

### Domain-Specific Tech Writers
**Decision:** Generalist model is correct. Domain teams (security, sre, etc.) own their specialized docs via doc-* skills. Doc-rite-pack handles general user-facing content only.

### CI Integration for Staleness
**Decision:** Out of scope for roster. Would require tooling integration outside Claude Code. On-demand `/doc-audit --staleness` is sufficient.

### Team-Specific Hooks
**Decision:** Hooks directory intentionally empty. Doc-team doesn't need custom context injection - uses standard project hooks.

---

## Documentation Ownership Clarification

Based on audit, confirm this ownership matrix:

| Doc Type | Owner | NOT docs |
|----------|-------|-------------------|
| User guides, tutorials | docs | |
| API reference docs | docs | |
| User-facing README | docs | |
| Code comments | | 10x-dev |
| ADRs | | 10x-dev |
| Internal design docs | | 10x-dev |
| Security/compliance docs | | security |
| Operational runbooks | | sre |
| Test documentation | | 10x-dev |

---

## Dependencies

| Item | Depends On |
|------|------------|
| P1 (10x gate) | Changes to 10x-dev/agents/qa-adversary.md |
| P2 (staleness) | No dependencies - doc-team internal |

---

## Next Team

Continue audit with: **hygiene** (code quality and refactoring)

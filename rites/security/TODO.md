# security TODO

> Audit conducted: 2026-01-02 | Status: MATURE, proactive security enhancement needed

## Current State Summary

The security rite is a **well-structured security assessment rite** with clear agent roles (threat-modeler, compliance-architect, penetration-tester, security-reviewer) and comprehensive templates.

**Strengths confirmed:**
- Clear 4-phase sequential workflow with complexity-based skipping
- Security-reviewer has blocking authority (APPROVE/REJECT)
- STRIDE/DREAD methodology for threat modeling, CVSS for pentesting
- Explicit handoff criteria and acid tests per agent
- Trust-based enforcement model is correct (deployment gates honor signoff)

**Architecture validated:**
- Remediation ownership clear: security identifies, 10x fixes, security validates
- QA-adversary vs security-reviewer separation is intentional (functional vs adversarial)
- PATCH complexity correctly skips threat/compliance phases

---

## Validated Improvements

### P1: Require Threat-Modeler for SYSTEM Complexity

**Gap identified:** The security rite is primarily reactive (post-implementation). No hook for proactive security input during design phase.

**Decision:** SYSTEM-level work (auth, crypto, PII, external integrations) requires threat-modeler consultation BEFORE implementation begins.

**Changes required:**
- [ ] Update `rites/10x-dev/agents/architect.md`: Add trigger for security consultation
  - "For SYSTEM complexity involving auth, crypto, or PII: consult threat-modeler before finalizing TDD"
- [ ] Update `rites/security/workflow.yaml`: Add pre-implementation phase option
- [ ] Define handoff: Architect provides draft TDD → threat-modeler provides THREAT-*.md → Architect incorporates into final TDD
- [ ] Update 10x-dev workflow.yaml to route SYSTEM complexity through security consultation

**SYSTEM complexity triggers:**
- Authentication/authorization changes
- Cryptographic implementations
- PII/sensitive data handling
- External API integrations
- Payment processing
- Session management changes

**Workflow after change:**
```
SYSTEM complexity:
10x Architect (draft TDD) → threat-modeler → Architect (final TDD) → 10x implementation → security assessment
```

---

### P2: Add Risk Acceptance Template

**Gap identified:** No formal workflow for accepting unmitigated vulnerabilities (timeline, cost, architectural constraints).

**Decision:** Create formal RISK-ACCEPT template with stakeholder sign-off, expiration date, and compensating controls.

**Changes required:**
- [ ] Add `risk-acceptance-template` to `skills/doc-security/SKILL.md`
- [ ] Template structure:
  ```markdown
  # RISK-ACCEPT-{slug}

  ## Vulnerability Summary
  - Finding ID: [link to PENTEST or SEC finding]
  - Severity: [CVSS score]
  - Description: [brief description]

  ## Acceptance Rationale
  - Why not fixed: [timeline | cost | architectural constraint | business decision]
  - Business justification: [why acceptable]

  ## Compensating Controls
  - [ ] Control 1: [description]
  - [ ] Control 2: [description]

  ## Scope & Expiration
  - Affected systems: [list]
  - Acceptance expires: [date - max 90 days recommended]
  - Review trigger: [conditions requiring re-evaluation]

  ## Approvals
  - Security: [name, date]
  - Engineering: [name, date]
  - Business Owner: [name, date]
  ```
- [ ] Update `agents/security-reviewer.md`: Reference risk acceptance for CONDITIONAL approvals
- [ ] Add to handoff criteria: "Risk acceptance documented for any unmitigated high/critical findings"

---

### P3: 10x→Security Uses Generalized Handoff Pattern

**Decision:** 10x→security handoff should use same HANDOFF artifact format as debt→hygiene, 10x→sre.

**Changes required:**
- [ ] Document 10x→security in generalized handoff pattern (ecosystem-level)
- [ ] Add to pattern documentation:
  ```yaml
  # Example: 10x → security handoff
  source_team: 10x-dev
  target_team: security
  handoff_type: assessment
  context:
    initiative: "User Authentication Redesign"
    complexity: SYSTEM
    source_artifacts:
      - docs/design/TDD-auth-redesign.md
      - docs/requirements/PRD-auth-redesign.md
  items:
    - id: SEC-001
      summary: "Full security assessment of auth redesign"
      scope:
        - OAuth2 implementation
        - Session management
        - Token storage
      acceptance_criteria:
        - Threat model complete
        - Penetration testing passed
        - Security signoff obtained
  ```
- [ ] Update 10x QA-adversary release checklist: "Security handoff prepared for FEATURE/SYSTEM complexity"

---

## Deferred / Not Prioritized

### Hard Enforcement Mechanism
**Decision:** Trust-based is correct. Security signoff is documented; deployment gates voluntarily honor it. Building enforcement hooks is infrastructure work outside knossos scope.

### Pre-Implementation for All Complexity
**Decision:** Only SYSTEM requires proactive threat modeling. PATCH/FEATURE remain reactive to avoid slowing velocity on low-risk changes.

---

## Dependencies

| Item | Depends On |
|------|------------|
| P1 (proactive security) | Changes to 10x-dev architect agent |
| P3 (handoff pattern) | Generalized handoff pattern from debt-triage TODO |

---

## Cross-Team Notes

**For 10x-dev:** P1 requires architect to consult threat-modeler for SYSTEM complexity before finalizing TDD. This is a new gate in the 10x workflow.

**For ecosystem:** P3 adds another handoff type (10x→security) to the generalized pattern.

---

## Next Team

Continue audit with: **intelligence** (analytics and research)

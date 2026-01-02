# debt-triage-pack TODO

> Audit conducted: 2026-01-02 | Status: MATURE, integration improvements needed

## Current State Summary

Debt-triage-pack is a **well-designed, recently optimized team** (34% compression achieved) with clear strategic focus on debt inventory, risk assessment, and sprint planning.

**Strengths confirmed:**
- Clear 3-phase pipeline (collection → assessment → planning)
- Well-defined handoff criteria between phases
- Acid tests ensure artifact completeness
- Back-routes for failure recovery
- All Opus is justified - debt categorization requires judgment

**Architecture validated:**
- Keep all agents on Opus (debt categorization needs reasoning power)
- Teams stay separate from hygiene-pack (strategic vs tactical are different)
- Skip calibration loop formalization (estimation improves naturally)

---

## Validated Improvements

### P1: Create Shared Smell Detection Skill

**Decision:** Debt Collector and Code Smeller both detect debt/smells. Create shared detection skill both teams use with different post-processing.

**Rationale:** Different purposes (strategic catalog vs tactical diagnosis) but same underlying detection patterns. Share the detection, diverge on what to do with findings.

**Changes required:**
- [ ] Create `skills/smell-detection/SKILL.md` (ecosystem-level, not team-specific)
- [ ] Define detection patterns: dead code, duplication, complexity, naming, imports, etc.
- [ ] Update `agents/debt-collector.md`: Reference shared skill for detection phase
- [ ] Update `teams/hygiene-pack/agents/code-smeller.md`: Reference same shared skill
- [ ] Each team post-processes differently:
  - Debt Collector: Adds business context, age, ownership → feeds Risk Assessor
  - Code Smeller: Adds ROI scoring, blast radius → feeds Architect Enforcer

**Skill structure:**
```
skills/smell-detection/
├── SKILL.md              # Detection patterns and techniques
├── patterns/
│   ├── dead-code.md      # Unused functions, orphaned modules
│   ├── duplication.md    # Copy-paste, DRY violations
│   ├── complexity.md     # Cyclomatic, nesting, god objects
│   ├── naming.md         # Terminology drift, misleading identifiers
│   └── imports.md        # Circular deps, unused, over-broad
└── tools/
    └── detection-checklist.md
```

---

### P2: Generalized Cross-Team Handoff Artifact Pattern

**Decision:** Instead of debt-triage-specific DEBT-HANDOFF artifact, create a **generalized handoff artifact pattern** that any team can use when handing off to another team.

**Rationale:** The debt→hygiene handoff is one instance of a broader pattern. 10x→doc-team, security→10x, etc. all need structured handoffs. Centralize the pattern.

**Changes required:**
- [ ] Create ecosystem-level handoff artifact schema in `skills/cross-team-handoff/`
- [ ] Define generic HANDOFF artifact format:
  ```yaml
  # HANDOFF-{source-team}-{target-team}-{slug}.yaml
  source_team: debt-triage-pack
  target_team: hygiene-pack
  created: 2026-01-02
  handoff_type: execution  # execution | review | consultation
  context:
    initiative: "Q1 Technical Debt Remediation"
    source_artifacts:
      - docs/debt/LEDGER-q1-2026.md
      - docs/debt/RISK-MATRIX-q1-2026.md
      - docs/debt/SPRINT-PLAN-q1-2026.md
  items:
    - id: PKG-001
      summary: "Extract shared email validator"
      priority: high
      location: src/api/validators/
      acceptance_criteria:
        - Single validateEmail function
        - All 3 API files import shared validator
        - Existing tests pass
  notes_for_target: |
    Risk Assessor scored these as high-impact, low-effort.
    Recommend starting with PKG-001 as it unblocks PKG-002.
  ```
- [ ] Update Sprint Planner to produce HANDOFF artifact as additional output
- [ ] Document pattern in ecosystem-pack or cross-team skill
- [ ] Update hygiene-pack code-smeller to accept HANDOFF artifact as input

**Pattern usage across ecosystem:**
| Source | Target | Handoff Type |
|--------|--------|--------------|
| debt-triage-pack | hygiene-pack | execution |
| 10x-dev-pack (QA) | doc-team-pack | documentation |
| security-pack | 10x-dev-pack | remediation |
| rnd-pack | 10x-dev-pack | productionization |

---

## Deferred / Not Prioritized

### Calibration Loop
**Decision:** Skip. Estimation improves naturally with experience. Formal calibration tracking is over-engineering.

### Merge with Hygiene-Pack
**Decision:** Keep separate. Strategic planning (debt-triage) and tactical execution (hygiene) are genuinely different workflows.

### Model Downgrade
**Decision:** Keep all Opus. Debt categorization ("Is this intentional design or debt?") requires judgment that justifies cost.

---

## Dependencies

| Item | Depends On |
|------|------------|
| P1 (shared detection) | Ecosystem-level skill creation |
| P2 (handoff pattern) | Ecosystem-level schema definition |

---

## Cross-Team Notes

**For ecosystem-pack:** P2 defines a generalized handoff artifact pattern that should live at ecosystem level, not team level. Consider adding to cross-team skill or creating dedicated handoff-protocol skill.

**For hygiene-pack:** Update P2 in hygiene-pack TODO to reference this generalized pattern instead of debt-specific artifact.

---

## Next Team

Continue audit with: **sre-pack** (operations and reliability)

# hygiene TODO

> Audit conducted: 2026-01-02 | Status: MATURE, one key integration improvement

## Current State Summary

The hygiene rite is a **well-designed, recently optimized rite** (36% compression achieved) with clear scope boundaries and explicit handoff criteria between phases.

**Strengths confirmed:**
- Clear 4-phase pipeline (assessment → planning → execution → audit)
- Strong behavior preservation emphasis throughout
- SPOT complexity for trivial fixes is useful and trust-based
- Boy scout rule is pragmatic, not scope creep
- Clear separation from 10x-dev (behavior preservation vs feature building)
- Clear separation from debt-triage (execution vs planning)
- Proportional tool access (orchestrator read-only, janitor full)

**Architecture validated:**
- SPOT complexity stays - trust user judgment, audit-lead catches mistakes
- Boy scout rule stays - pragmatic adjacent fixes are fine
- Janitor using Sonnet (not Opus) is correct cost optimization

---

## Validated Improvements

### P1: Define "Behavior" as Contract-Preserving

**Gap identified:** "Behavior preservation" is the absolute constraint but undefined. Changed error messages? Improved logging? Different exception types?

**Decision:** Pragmatic, contract-preserving definition:
- Public API contracts must be preserved
- Internal changes (logs, error text, performance improvements) allowed if documented
- Not "output-identical" (too strict) or "undefined" (too loose)

**Changes required:**
- [ ] Update `agents/architect-enforcer.md`: Add behavior preservation definition
- [ ] Add section: "What 'behavior preservation' means"
  - Preserved: Public API signatures, return types, error semantics, documented contracts
  - Allowed: Internal logging, error message text, performance characteristics, private implementations
  - Requires approval: Any change to documented behavior, even if "improved"
- [ ] Update `agents/audit-lead.md`: Reference the definition in verification checklist

**Example contract language:**
```markdown
## Behavior Preservation Scope

**MUST preserve (contract):**
- Function signatures and return types
- Error conditions that trigger exceptions
- Documented API behavior
- Data structure shapes

**MAY change (internal):**
- Log messages and levels
- Error message text (not error types)
- Performance characteristics
- Private implementation details

**REQUIRES explicit approval:**
- "Improved" error messages that change documented behavior
- Performance changes that affect SLAs
- Any change a caller might depend on
```

---

### P2: Formalize Debt-Triage → Hygiene Handoff

**Gap identified:** The debt-triage rite plans what to fix; hygiene executes. But NO formal handoff mechanism exists. The relationship is documented but not operationalized.

**Decision:** Debt-triage output should be directly consumable by hygiene with explicit artifact format.

**Changes required:**
- [ ] Define handoff artifact format in `rites/debt-triage/` (debt-triage side)
- [ ] Update `rites/hygiene/agents/code-smeller.md`: Accept debt-triage artifact as input
- [ ] Add to code-smeller's approach: "If debt-triage artifact provided, use as starting point for assessment"
- [ ] Document handoff in both team READMEs

**Handoff artifact format (proposed):**
```yaml
# DEBT-HANDOFF-{slug}.yaml
source: debt-triage
target: hygiene
created: 2026-01-02
items:
  - id: DEBT-001
    category: duplication
    location: src/api/validators/
    severity: high
    recommended_action: extract shared validator
    context: "3 files with identical email validation"
  - id: DEBT-002
    category: complexity
    location: src/services/billing.ts
    severity: medium
    recommended_action: decompose god object
    context: "1200 lines, 15 public methods"
```

**Workflow after formalization:**
1. `/debt` produces DEBT-HANDOFF artifact
2. User reviews, approves items for action
3. `/hygiene` with `--from-debt DEBT-HANDOFF-{slug}.yaml`
4. Code-smeller uses artifact as assessment input (not starting from scratch)

---

## Deferred / Not Prioritized

### SPOT Criteria Objectification
**Decision:** Keep SPOT subjective. Trust user judgment; audit-lead catches mistakes. Adding objective criteria (file count, line count) would add bureaucracy without value.

### Remove Boy Scout Rule
**Decision:** Keep boy scout rule. Pragmatic adjacent fixes (<5 lines, own commit) are fine. Zero tolerance would slow reasonable cleanup.

### Merge with Debt-Triage
**Decision:** Keep separate. Debt-triage is strategic (what to fix when); hygiene is tactical (how to fix now). Different concerns, different agents, different workflows.

---

## Dependencies

| Item | Depends On |
|------|------------|
| P2 (debt handoff) | Changes to debt-triage to produce handoff artifact |

---

## Cross-Rite Note

P2 creates a dependency: debt-triage must be updated to produce the handoff artifact. Add this to debt-triage TODO when audited.

---

## Next Rite

Continue audit with: **debt-triage** (technical debt prioritization)

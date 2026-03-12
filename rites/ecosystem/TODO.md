# ecosystem TODO

> Audit conducted: 2026-01-02 | Status: OVER-ENGINEERED, needs simplification

## Current State Summary

The ecosystem rite is a **5-agent infrastructure rite** handling sync pipeline and knossos patterns. The audit revealed it's well-documented but claims more scope than it actually needs:

- Claims "hub rite" status with cross-rite coordination → but no mechanism exists
- Defines satellite diversity matrix for testing → but doesn't actually work on satellites
- Assumes orchestrated mode only → but users want direct agent access

**Core strength**: Clear phase boundaries (analyze → design → implement → document → validate) with explicit handoff criteria and backward compatibility emphasis.

**Core problem**: Scope creep from "infrastructure specialist" to "ecosystem coordinator" without the mechanisms to support coordination.

---

## Validated Improvements

### P1: Demote from Hub to Specialist

**Decision**: Remove hub claims. The ecosystem rite is an infrastructure specialist with user-escalation only.

**Rationale**: No rite leads registry exists, no notification mechanism, no coordination protocol. Other rites discover impact via sync failures, which is fine.

**Changes required:**
- [ ] Update `orchestrator.yaml`: Remove `cross_team_protocol` hub language
- [ ] Update `README.md`: Remove hub rite positioning
- [ ] Update `agents/potnia.md`: Change escalation path to user-only (not "all rite leads")
- [ ] Simplify cross-rite references to "escalate to user when changes affect other teams"

---

### P2: Remove Satellite Testing Matrix

**Decision**: Ditch satellite diversity matrix concept entirely.

**Rationale**: The ecosystem rite doesn't work on satellite projects. Satellites just sync from the ecosystem. The testing matrix (test-baseline, test-minimal, test-complex, test-legacy, test-production-like) adds complexity without value.

**Changes required:**
- [ ] Update `agents/compatibility-tester.md`: Remove satellite matrix references
- [ ] Remove complexity-based satellite selection from workflow
- [ ] Update `skills/doc-ecosystem/SKILL.md`: Remove satellite testing templates
- [ ] Simplify to: "Test in test satellite, verify ari sync works"

**What to keep:**
- ari sync validation (does `ari sync` succeed?)
- Schema compatibility checks
- Breaking change documentation

---

### P3: Support Cross-Cutting Execution

**Decision**: The ecosystem rite should work in cross-cutting mode (session active, no rite) for direct agent queries.

**Rationale**: Sometimes you just want to ask context-architect a design question without spinning up full orchestration.

**Changes required:**
- [ ] Update `workflow.yaml`: Add `supports_cross_cutting: true`
- [ ] Update `agents/potnia.md`: Document behavior in cross-cutting mode (advisory only, no phase enforcement)
- [ ] Consider: Allow direct `/ecosystem-design` without full pipeline in cross-cutting mode

---

### P4: Acknowledge Breaking Changes as Intentional

**Decision**: Backward compatibility validation is spot-check only, because greenfield migrations are common and acceptable.

**Rationale**: Breaking changes aren't bugs—they're intentional architectural decisions. Over-validating compatibility would slow legitimate migrations.

**Changes required:**
- [ ] Update `agents/context-architect.md`: Remove "backward compatible" as default expectation
- [ ] Add guidance: "Breaking changes are valid when migration path is documented"
- [ ] Update `skills/doc-ecosystem/templates/context-design.md`: Change "backward compatible OR breaking" to "document migration path if breaking"

---

## Deferred / Not Prioritized

### Automated Schema Diffing
**Decision**: Not needed if we accept breaking changes intentionally. Migration runbooks are sufficient documentation.

### Artifact Versioning
**Decision**: Low priority. Current artifact structure is adequate for actual use patterns.

### Sync Diagnostic Framework
**Decision**: Keep `/sync-debug` command but don't expand. Most sync issues are failures that self-diagnose.

---

## Architectural Observations

### What the Ecosystem Rite Actually Does Well
1. Gap Analysis for infrastructure bugs
2. Context Design for sync/knossos changes
3. Migration Runbooks for breaking changes
4. Inscription architecture governance (inscription-architecture skill)

### What It Over-Claims
1. Hub coordination (no mechanism)
2. Satellite testing (doesn't work on satellites)
3. Orchestrated-only execution (too rigid)

### Recommended Focus
- **Core**: Sync pipeline and knossos infrastructure changes
- **Secondary**: Context file governance
- **Remove**: Cross-rite coordination, satellite testing matrix

---

## Dependencies

| Item | Depends On |
|------|------------|
| Cross-cutting support | Workflow.yaml schema supporting `supports_cross_cutting` flag |
| Hub demotion | No dependencies - can remove immediately |

---

## Next Rite

Continue audit with: **docs** (documentation and technical writing)

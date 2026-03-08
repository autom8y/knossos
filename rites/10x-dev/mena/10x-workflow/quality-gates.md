# Quality Gates

> Mandatory checkpoints between workflow phases

## What is a Quality Gate?

Quality gates are **mandatory checkpoints** between phases. Passing a gate requires all criteria to be met.

Quality gates prevent low-quality work from propagating downstream. **Failing a gate means routing back, not proceeding.**

---

## PRD Quality Gate

**Owner**: Requirements Analyst produces -> Potnia verifies

| Criterion | Validation |
|-----------|------------|
| Problem statement clear | Can be explained in one paragraph |
| Scope explicit | In-scope AND out-of-scope defined |
| Requirements testable | Each has acceptance criteria |
| Priorities assigned | MoSCoW applied |
| No blocking questions | Or owners assigned with dates |
| **Impact assessed** | impact: low/high with categories if high |

### Impact Assessment Validation

The impact assessment determines workflow routing:
- **Low impact**: Standard complexity-based routing applies
- **High impact**: Routes to Architect even at SCRIPT complexity

**Validation checklist**:
- [ ] Impact level explicitly stated (low or high)
- [ ] If high: impact_categories populated with at least one category
- [ ] Rationale documented for the impact determination
- [ ] Categories match the architectural significance (security, data_model, api_contract, auth, cross_service)

### Detailed Template

For complete PRD structure and expanded quality checklist, see [prd-template](../doc-artifacts/prd-template.lego.md).

---

## TDD Quality Gate

**Owner**: Architect produces -> Potnia verifies

| Criterion | Validation |
|-----------|------------|
| Traces to PRD | Every design element maps to requirement |
| Decisions documented | Significant choices have ADRs |
| Interfaces defined | Component boundaries clear |
| Complexity justified | Matches actual requirement |
| Risks identified | With mitigations |

### Detailed Template

For complete TDD structure and expanded quality checklist, see [tdd-template](../doc-artifacts/tdd-template.lego.md).

---

## Implementation Quality Gate

**Owner**: Principal Engineer produces -> Potnia verifies

| Criterion | Validation |
|-----------|------------|
| Satisfies TDD | Implementation matches design |
| Tests exist | All paths covered including errors |
| Type-safe | Full type hints |
| Readable | Non-author can understand |
| Documented | Implementation decisions recorded |

---

## Validation Quality Gate

**Owner**: QA/Adversary produces -> Potnia verifies

| Criterion | Validation |
|-----------|------------|
| Acceptance criteria met | All requirements validated |
| Edge cases covered | Boundary conditions tested |
| Failures handled | Error scenarios verified |
| Production ready | Deployment requirements met |

### Detailed Template

For complete Test Plan structure and expanded quality checklist, see [test-templates](../doc-artifacts/test-templates.lego.md).

---

## Workflow Integration

Quality gates are verified in the **VERIFY** step of each session. See [lifecycle.md](lifecycle.md) for the complete session protocol.

**Non-Negotiable Rule**: DO NOT proceed to the next phase if a quality gate has not been met. Route back to the responsible agent to resolve gaps.

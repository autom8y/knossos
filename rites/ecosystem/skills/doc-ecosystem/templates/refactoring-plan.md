# Refactoring Plan Template

> Phased refactoring sequence.

```markdown
# Refactoring Plan
**Based on**: [smell report reference]
**Prepared**: [date]
**Scope**: [what will be refactored]

## Architectural Assessment

### Boundary Health
- [Module A]: Clean boundaries, local cleanup only
- [Module B]: Leaking internals to Module C
- [Module C]: God module, needs decomposition

### Root Causes Identified
1. [Root cause 1]: Explains smells DC-001, DC-003, CX-007
2. [Root cause 2]: Explains smells DRY-002, DRY-005

## Refactoring Sequence

### Phase 1: Foundation [Low Risk]
**Goal**: Prepare for larger refactors without changing behavior

#### RF-001: [Refactoring name]
- **Smells addressed**: DC-001, NM-003
- **Category**: Local
- **Before**: [current state]
- **After**: [target state]
- **Invariants**: [what must stay true]
- **Verification**: [how to confirm success]
- **Commit scope**: [what goes in one commit]

[Rollback point: can stop here safely]

### Phase 2: Module Cleanup [Medium Risk]
**Goal**: Clean up internal module structure

#### RF-002: [Refactoring name]
[Same structure as RF-001]

### Phase 3: Boundary Repair [Higher Risk]
**Goal**: Fix cross-module issues and restore encapsulation

[Same structure]

## Risk Matrix
| Refactor | Risk | Blast Radius | Rollback Cost |
|----------|------|--------------|---------------|
| RF-001   | Low  | 2 files      | Trivial       |
| RF-002   | Med  | 1 module     | 1 commit      |
| RF-003   | High | 3 modules    | 3 commits     |

## Notes for Janitor
- Commit message conventions: [format]
- Test run requirements: [what tests after each commit]
- Files to avoid touching: [generated code, etc.]
- Order is critical for: [specific refactors with dependencies]

## Out of Scope
Findings deferred for future work:
- [Finding X]: Requires feature work, not just cleanup
- [Finding Y]: Needs architectural decision from user
```

## Quality Gate

**Refactoring Plan complete when:**
- Phases sequenced by risk (low to high)
- Invariants defined for each refactoring
- Commit scope clear
- Rollback points identified

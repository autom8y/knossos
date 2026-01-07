# E2E Test: Debt Remediation Workflow

> End-to-end test scenario for debt-triage-pack to hygiene-pack cross-team workflow.
> Version: 1.0.0

## Overview

This document defines a complete test scenario for technical debt remediation, validating the handoff from debt-triage-pack (planning) to hygiene-pack (execution) and behavior preservation validation throughout.

**Workflow Path**: debt-triage planning -> HANDOFF -> hygiene execution
**Primary Teams**: debt-triage-pack, hygiene-pack
**Cross-Team Handoffs**: debt-triage-pack -> hygiene-pack (execution handoff)

---

## Test Scenario: Validator Module Debt Remediation

### Scenario Description

Address accumulated technical debt in the validator module: duplicate email validation logic across 3 files, dead utility functions, and inconsistent error message formatting.

**Why This Scenario**: This represents a typical cross-team debt remediation that:
- Requires full debt-triage-pack workflow (collection, assessment, planning)
- Produces execution HANDOFF for hygiene-pack
- Requires behavior preservation validation
- Tests the debt-triage -> hygiene handoff path

---

## Phase 1: Debt Collection

### Entry Criteria
- [ ] User request or scheduled audit trigger
- [ ] Session initialized with `/start`
- [ ] Complexity = AUDIT (full debt discovery)

### Agent
**debt-collector** (debt-triage-pack)

### Input
User request: "Audit the validator module for technical debt. We've noticed duplicated code and some functions that seem unused."

### Expected Artifact: Debt Ledger

```markdown
# Debt Ledger: Validator Module

## Summary
- Total items: 8
- Categories: Code (5), Test (2), Doc (1)
- Age range: 4-18 months

## Debt Items

### C001: Duplicate email validation
- **Location**: `src/validators/user.ts:45`, `src/validators/contact.ts:23`, `src/api/forms.ts:112`
- **Category**: Code > Duplication
- **Evidence**: 3 implementations with 85% similarity (diff analysis)
- **Age**: 14 months (git blame)
- **Owner**: @platform-team

### C002: Unused formatPhone utility
- **Location**: `src/utils/formatters.ts:67`
- **Category**: Code > Dead Code
- **Evidence**: 0 call sites (grep confirms), not exported
- **Age**: 18 months
- **Owner**: @platform-team

### C003: Inconsistent error messages
- **Location**: `src/validators/*.ts` (12 instances)
- **Category**: Code > Naming
- **Evidence**: Mix of "Invalid X", "X is invalid", "Bad X format"
- **Age**: 8 months (earliest inconsistency)
- **Owner**: @platform-team

### C004: Dead validateSSN function
- **Location**: `src/validators/legacy.ts:34`
- **Category**: Code > Dead Code
- **Evidence**: 0 call sites, SSN collection removed in v2.1
- **Age**: 12 months
- **Owner**: @platform-team

### C005: Redundant null checks
- **Location**: `src/validators/address.ts:15-45`
- **Category**: Code > Complexity
- **Evidence**: TypeScript strict mode makes checks unnecessary
- **Age**: 6 months (since strict mode enabled)
- **Owner**: @platform-team

### T001: Missing edge case tests for email validator
- **Location**: `tests/validators/email.test.ts`
- **Category**: Test > Coverage
- **Evidence**: No tests for Unicode, plus addressing, long TLDs
- **Age**: 14 months
- **Owner**: @platform-team

### T002: Flaky timeout test
- **Location**: `tests/validators/async.test.ts:89`
- **Category**: Test > Reliability
- **Evidence**: 3 failures in last 100 CI runs
- **Age**: 4 months
- **Owner**: @platform-team

### D001: Outdated validation API docs
- **Location**: `docs/api/validators.md`
- **Category**: Doc > Stale
- **Evidence**: References deprecated methods
- **Age**: 10 months
- **Owner**: @platform-team
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| Debt ledger exists | YES | [ ] |
| Each item has location | YES | [ ] |
| Each item has category | YES | [ ] |
| Each item has evidence | YES | [ ] |
| Each item has age | YES | [ ] |
| Items use smell-detection patterns | YES | [ ] |

### Phase Transition
- **From**: collection
- **To**: assessment
- **Handoff Type**: Internal (within debt-triage-pack)
- **Trigger**: Debt ledger complete

---

## Phase 2: Risk Assessment

### Entry Criteria
- [ ] Debt ledger complete
- [ ] All items cataloged with evidence

### Agent
**risk-assessor** (debt-triage-pack)

### Input
- Debt ledger from Phase 1
- Codebase context for blast radius analysis

### Expected Artifact: Risk Matrix

```markdown
# Risk Matrix: Validator Module Debt

## Assessment Criteria

| Criterion | Weight | Description |
|-----------|--------|-------------|
| Severity | 3x | Impact if not addressed |
| Frequency | 2x | How often issue causes problems |
| Blast Radius | 2x | Files/components affected |
| Fix Complexity | 1x (inverse) | Effort to resolve |

## Scored Items

| Item | Severity | Frequency | Blast | Complexity | Score | Priority |
|------|----------|-----------|-------|------------|-------|----------|
| C001 | High (3) | High (3) | Med (2) | Low (1) | 21 | P1 |
| C003 | Med (2) | Med (2) | High (3) | Low (1) | 16 | P2 |
| C002 | Low (1) | Low (1) | Low (1) | Low (1) | 4 | P3 |
| C004 | Low (1) | Low (1) | Low (1) | Low (1) | 4 | P3 |
| C005 | Low (1) | Low (1) | Med (2) | Low (1) | 6 | P3 |
| T001 | Med (2) | Med (2) | Med (2) | Med (2) | 12 | P2 |
| T002 | Med (2) | High (3) | Low (1) | Low (1) | 11 | P2 |
| D001 | Low (1) | Low (1) | Low (1) | Low (1) | 4 | P3 |

## Priority Summary
- **P1 (Critical)**: C001 - Address immediately
- **P2 (High)**: C003, T001, T002 - Address in sprint
- **P3 (Medium)**: C002, C004, C005, D001 - Address opportunistically

## Risk Analysis

### C001: Duplicate email validation (P1)
**Why Critical**:
- Bug in one implementation diverged from others last month
- 3 different validation behaviors causing user confusion
- High-traffic paths affected

**Blast Radius**:
- 3 files directly
- 15 files importing these validators
- All user-facing forms

**Recommended Approach**: Extract to shared module, update imports

### C003: Inconsistent error messages (P2)
**Why High**:
- User confusion from inconsistent messaging
- Localization team flagged as blocker for i18n

**Recommended Approach**: Define error message constants
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| Risk matrix exists | YES | [ ] |
| Scoring criteria defined | YES | [ ] |
| All debt items scored | YES | [ ] |
| Priority levels assigned | YES | [ ] |
| Blast radius analyzed for P1 items | YES | [ ] |
| Recommendations included | YES | [ ] |

### Phase Transition
- **From**: assessment
- **To**: planning
- **Handoff Type**: Internal (within debt-triage-pack)
- **Trigger**: Risk matrix complete

---

## Phase 3: Sprint Planning

### Entry Criteria
- [ ] Risk matrix complete
- [ ] Priorities assigned

### Agent
**sprint-planner** (debt-triage-pack)

### Input
- Debt ledger
- Risk matrix
- Team capacity context

### Expected Artifact: Sprint Debt Packages

```markdown
# Sprint Debt Packages: Validator Module

## Sprint Overview
- **Sprint Goal**: Eliminate P1 debt, address high-impact P2 items
- **Estimated Effort**: 8-12 hours
- **Team**: hygiene-pack

## Package 1: Email Validator Consolidation (P1)

### Items Included
- C001: Duplicate email validation

### Acceptance Criteria
- [ ] Single `validateEmail` function in `src/validators/shared/email.ts`
- [ ] All 3 original locations import shared validator
- [ ] Existing tests pass unchanged (behavior preservation)
- [ ] New edge case tests added (T001 addressed)

### Behavior Preservation Checklist
- [ ] Test: all current valid emails still validate
- [ ] Test: all current invalid emails still rejected
- [ ] Test: error message format unchanged
- [ ] Test: async behavior unchanged

### Estimated Effort: 4-6 hours

### Dependencies: None

---

## Package 2: Error Message Standardization (P2)

### Items Included
- C003: Inconsistent error messages

### Acceptance Criteria
- [ ] Error message constants in `src/validators/errors.ts`
- [ ] All 12 instances use standard format
- [ ] Format: "Invalid {field}: {reason}"
- [ ] i18n-ready with message keys

### Behavior Preservation Checklist
- [ ] Test: all validators still throw on invalid input
- [ ] Test: error types unchanged
- [ ] Test: error codes unchanged (if applicable)

### Estimated Effort: 2-3 hours

### Dependencies: None (can parallelize with Package 1)

---

## Package 3: Dead Code Cleanup (P3)

### Items Included
- C002: Unused formatPhone utility
- C004: Dead validateSSN function

### Acceptance Criteria
- [ ] Both functions removed
- [ ] No references remain (grep verification)
- [ ] Tests removed if any exist

### Behavior Preservation Checklist
- [ ] Confirm 0 runtime call sites (already verified in collection)
- [ ] Confirm no dynamic imports (eval, require(var))

### Estimated Effort: 1 hour

### Dependencies: Run after Package 1 and 2 to ensure no hidden usages

---

## Package 4: Flaky Test Fix (P2)

### Items Included
- T002: Flaky timeout test

### Acceptance Criteria
- [ ] Test no longer flaky (10 consecutive passes)
- [ ] Timeout increased or approach changed
- [ ] CI history shows stability

### Estimated Effort: 1-2 hours

### Dependencies: None

---

## Execution Order
1. Package 1 + Package 2 (parallel)
2. Package 3 (after 1 and 2)
3. Package 4 (any time)

## Rollback Plan
Each package is independently deployable. Rollback by reverting PR.
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| Sprint packages defined | YES | [ ] |
| Each package has acceptance criteria | YES | [ ] |
| Behavior preservation checklists present | YES | [ ] |
| Effort estimates included | YES | [ ] |
| Dependencies documented | YES | [ ] |
| Execution order specified | YES | [ ] |

### Phase Transition: CROSS-TEAM HANDOFF

- **From**: planning (debt-triage-pack)
- **To**: execution (hygiene-pack)
- **Handoff Type**: Cross-team execution handoff
- **Trigger**: Sprint packages approved

---

## Phase 4: Cross-Team Handoff

### HANDOFF Artifact (debt-triage -> hygiene)

```yaml
---
source_team: debt-triage-pack
target_team: hygiene-pack
handoff_type: execution
created: 2026-01-02
initiative: Q1 Technical Debt Remediation
priority: high
---

## Context

Debt triage complete for validator module. Risk assessment prioritized items
by impact and blast radius. Sprint packages are ready for execution.

## Source Artifacts
- `docs/debt/DEBT-LEDGER-validators.md`
- `docs/debt/RISK-MATRIX-validators.md`
- `docs/debt/SPRINT-PACKAGES-validators.md`

## Items

### PKG-001: Email Validator Consolidation
- **Priority**: High (P1 debt item)
- **Summary**: Consolidate 3 duplicate email validation implementations
- **Acceptance Criteria**:
  - Single `validateEmail` function in shared module
  - All 3 original locations import shared validator
  - All existing tests pass unchanged
  - Edge case tests added for Unicode, plus addressing, long TLDs

### PKG-002: Error Message Standardization
- **Priority**: Medium (P2 debt item)
- **Summary**: Standardize 12 inconsistent error message formats
- **Acceptance Criteria**:
  - Error constants in `src/validators/errors.ts`
  - All instances use format "Invalid {field}: {reason}"
  - i18n-ready with message keys

### PKG-003: Dead Code Cleanup
- **Priority**: Low (P3 debt items)
- **Summary**: Remove formatPhone and validateSSN dead code
- **Acceptance Criteria**:
  - Both functions removed
  - No references remain
  - Associated tests removed

### PKG-004: Flaky Test Fix
- **Priority**: Medium (P2 debt item)
- **Summary**: Fix flaky timeout test in async validator
- **Acceptance Criteria**:
  - 10 consecutive CI passes
  - Test approach stabilized

## Notes for Target Team

Recommend executing PKG-001 and PKG-002 in parallel.
PKG-003 should follow to ensure no hidden usages surface.
Behavior preservation is critical - run full test suite after each package.

Total estimated effort: 8-12 hours across 4 packages.
Risk assessor available for clarification: @risk-assessor
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| HANDOFF follows schema | YES | [ ] |
| `handoff_type: execution` | YES | [ ] |
| All packages have acceptance criteria | YES | [ ] |
| Source artifacts listed | YES | [ ] |
| Notes provide execution guidance | YES | [ ] |

---

## Phase 5: Hygiene Execution

### Entry Criteria
- [ ] HANDOFF received from debt-triage-pack
- [ ] Sprint packages reviewed
- [ ] Acceptance criteria clear

### Team
**hygiene-pack**

### Workflow
```
code-smeller (verify findings) -> architect-enforcer (plan) -> janitor (execute) -> audit-lead (verify)
```

### Expected Artifacts

For each package:

1. **Smell Verification** (code-smeller confirms debt-collector findings)
2. **Refactor Plan** (architect-enforcer details approach)
3. **Commits** (janitor executes refactoring)
4. **Audit Signoff** (audit-lead validates behavior preservation)

### Package 1 Execution Example

#### code-smeller Output
```markdown
## Smell Verification: Email Validator Duplication

**Debt Item**: C001
**Confirmed**: YES

### Evidence Re-verified
- `src/validators/user.ts:45-67` - 23 lines
- `src/validators/contact.ts:23-44` - 22 lines
- `src/api/forms.ts:112-135` - 24 lines

### Similarity Analysis
- Overall: 85% similarity
- Core regex: 100% identical
- Edge case handling: 60% (divergence source)

### Divergence Details
- user.ts: Allows plus addressing
- contact.ts: Rejects plus addressing
- forms.ts: Different error messages

### Recommendation
Unify to permissive behavior (allow plus addressing), standardize error messages.
```

#### architect-enforcer Output
```markdown
## Refactor Plan: Email Validator Consolidation

### Target State
Single `validateEmail` in `src/validators/shared/email.ts`

### Steps
1. Create `src/validators/shared/email.ts` with unified implementation
2. Add comprehensive tests in `tests/validators/shared/email.test.ts`
3. Update `src/validators/user.ts` to import shared
4. Update `src/validators/contact.ts` to import shared
5. Update `src/api/forms.ts` to import shared
6. Remove duplicated code from original locations

### Behavior Decision
- Plus addressing: ALLOW (most permissive)
- Error format: Use new standard format
- Document behavior change in CHANGELOG

### Test Strategy
- Run existing tests after each file update
- Add new tests for unified behavior
- Regression test: sample of 1000 production emails
```

#### janitor Output
```markdown
## Execution Report: Email Validator Consolidation

### Commits
1. `abc123` - feat(validators): add shared email validator
2. `def456` - refactor(user): use shared email validator
3. `ghi789` - refactor(contact): use shared email validator
4. `jkl012` - refactor(forms): use shared email validator
5. `mno345` - chore(validators): remove duplicated email validation

### Test Results
- Existing tests: 47/47 passing
- New tests: 12/12 passing
- No regressions detected

### Behavior Changes
- contact.ts now allows plus addressing (documented)
- Error messages standardized
```

#### audit-lead Output
```markdown
## Audit Signoff: Email Validator Consolidation

### Behavior Preservation Checklist
- [x] All current valid emails still validate
- [x] All current invalid emails still rejected (except plus addressing)
- [x] Error message format standardized (intentional change)
- [x] Async behavior unchanged

### Deviation Notes
- Plus addressing behavior unified (was inconsistent)
- Documented in CHANGELOG per refactor plan

### Test Coverage
- Line coverage: 94% (was 78%)
- Branch coverage: 89% (was 71%)

### Verdict: APPROVED
Package 1 complete. Behavior preserved with documented intentional unification.
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| Each package has smell verification | YES | [ ] |
| Each package has refactor plan | YES | [ ] |
| Commits reference package ID | YES | [ ] |
| Audit signoff for each package | YES | [ ] |
| Behavior preservation validated | YES | [ ] |

---

## Phase 6: Completion Report

### Entry Criteria
- [ ] All packages executed
- [ ] All audits passed
- [ ] Behavior preservation confirmed

### Expected Artifact: Remediation Report

```markdown
# Remediation Report: Validator Module Debt

## Summary
- **Packages Completed**: 4/4
- **Debt Items Resolved**: 6/8
- **Remaining Items**: D001 (docs), C005 (low priority)
- **Total Effort**: 10 hours

## Package Status

| Package | Status | Commits | Notes |
|---------|--------|---------|-------|
| PKG-001 | Complete | 5 | Plus addressing unified |
| PKG-002 | Complete | 3 | i18n keys added |
| PKG-003 | Complete | 2 | Clean removal |
| PKG-004 | Complete | 1 | Timeout increased |

## Behavior Changes
1. Plus addressing now allowed in all validators (was inconsistent)
2. Error messages standardized to "Invalid {field}: {reason}"

## Metrics Improvement
- Duplication: 15% -> 3%
- Dead code: 45 LOC removed
- Test coverage: 78% -> 94%
- Flaky test rate: 3% -> 0%

## Recommendations for Next Sprint
- Address D001 (outdated docs) with doc-team-pack
- Address C005 (redundant null checks) if touching those files
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| Remediation report exists | YES | [ ] |
| All package statuses documented | YES | [ ] |
| Behavior changes documented | YES | [ ] |
| Metrics before/after included | YES | [ ] |
| Recommendations for remaining items | YES | [ ] |

---

## Complete Test Checklist

### Phase Completeness (debt-triage-pack)
- [ ] Phase 1 (Collection): Debt ledger produced
- [ ] Phase 2 (Assessment): Risk matrix produced
- [ ] Phase 3 (Planning): Sprint packages produced

### Cross-Team Handoff
- [ ] HANDOFF artifact follows execution schema
- [ ] Acceptance criteria per package
- [ ] Source artifacts linked
- [ ] Priority set appropriately

### Phase Completeness (hygiene-pack)
- [ ] Smell verification for each package
- [ ] Refactor plans for each package
- [ ] Commits for each package
- [ ] Audit signoff for each package

### Behavior Preservation
- [ ] Preservation checklist defined per package
- [ ] All tests pass after each package
- [ ] Intentional changes documented
- [ ] Audit signoff confirms preservation

### Final Artifacts
- [ ] Remediation report produced
- [ ] Metrics improvement documented
- [ ] Remaining items identified

---

## Running This Test

### Manual Execution

1. Initialize debt-triage session:
   ```
   /start initiative="Validator Debt Remediation" complexity=AUDIT team=debt-triage-pack
   ```

2. Execute Phase 1 (Collection):
   ```
   Task(debt-collector, "Audit validator module for technical debt...")
   ```

3. Execute Phase 2 (Assessment):
   ```
   Task(risk-assessor, "Score and prioritize debt items...")
   ```

4. Execute Phase 3 (Planning):
   ```
   Task(sprint-planner, "Create sprint packages for remediation...")
   ```

5. Produce cross-team handoff and switch teams:
   ```
   /team hygiene-pack
   ```

6. Execute hygiene workflow for each package:
   ```
   Task(code-smeller, "Verify findings for PKG-001...")
   Task(architect-enforcer, "Plan refactoring for PKG-001...")
   Task(janitor, "Execute refactoring for PKG-001...")
   Task(audit-lead, "Validate behavior preservation for PKG-001...")
   ```

7. Repeat for remaining packages.

### Validation Points

| Gate | Trigger | Expected |
|------|---------|----------|
| Collection -> Assessment | Debt ledger complete | Items have location, evidence, age |
| Assessment -> Planning | Risk matrix complete | Items scored and prioritized |
| Planning -> HANDOFF | Packages approved | Execution handoff to hygiene-pack |
| Per-Package Completion | Audit signoff | Behavior preservation confirmed |

---

## Related Documents

- [Cross-Team Coordination Playbook](../playbooks/cross-rite-coordination.md)
- [Handoff Smoke Tests](handoff-smoke-tests.md)
- [Debt Triage Pack Workflow](../../teams/debt-triage-pack/workflow.md)
- [Hygiene Pack Workflow](../../teams/hygiene-pack/workflow.md)
- [Shared Templates: Debt Ledger](../../.claude/skills/shared/shared-templates/templates/debt-ledger.md)
- [Cross-Team Handoff Schema](../../.claude/skills/shared/cross-team-handoff/schema.md)

# Handoff Smoke Tests

> Minimal validation for all cross-team handoff paths.
> Version: 1.0.0

## Overview

This document defines smoke tests for all valid cross-team handoff paths in the roster ecosystem. Each test validates that a handoff can be produced, transmitted, and consumed correctly.

**Teams**: 10 (10x-dev-pack, debt-triage-pack, doc-team-pack, ecosystem-pack, hygiene-pack, intelligence-pack, rnd-pack, security-pack, sre-pack, strategy-pack)

**Handoff Types**: 6 (execution, validation, assessment, implementation, strategic_input, strategic_evaluation)

---

## Handoff Path Matrix

### Primary Handoff Paths

| # | Source | Target | Type | Priority | Description |
|---|--------|--------|------|----------|-------------|
| 1 | 10x-dev-pack | security-pack | assessment | High | Threat modeling gate |
| 2 | 10x-dev-pack | sre-pack | validation | High | Production readiness |
| 3 | 10x-dev-pack | doc-team-pack | assessment | Medium | Documentation review |
| 4 | debt-triage-pack | hygiene-pack | execution | High | Debt remediation |
| 5 | strategy-pack | 10x-dev-pack | implementation | High | Strategic initiative |
| 6 | intelligence-pack | strategy-pack | strategic_input | Medium | Research synthesis |
| 7 | rnd-pack | strategy-pack | strategic_evaluation | Medium | Prototype evaluation |
| 8 | rnd-pack | 10x-dev-pack | implementation | Medium | Proven prototype |
| 9 | security-pack | 10x-dev-pack | assessment | High | Security findings |
| 10 | sre-pack | 10x-dev-pack | validation | High | Operational findings |
| 11 | hygiene-pack | debt-triage-pack | execution | Low | Remediation report |
| 12 | ecosystem-pack | 10x-dev-pack | implementation | Medium | Ecosystem tooling |

### Secondary Handoff Paths

| # | Source | Target | Type | Priority | Description |
|---|--------|--------|------|----------|-------------|
| 13 | 10x-dev-pack | intelligence-pack | assessment | Low | User research request |
| 14 | 10x-dev-pack | rnd-pack | assessment | Low | Technical spike request |
| 15 | strategy-pack | rnd-pack | implementation | Medium | Research direction |
| 16 | doc-team-pack | 10x-dev-pack | assessment | Low | Doc feedback |
| 17 | sre-pack | security-pack | assessment | Medium | Security incident |
| 18 | security-pack | sre-pack | validation | Medium | Security controls |

---

## Smoke Test Definitions

### Test Structure

Each smoke test validates:
1. **Schema Compliance**: Frontmatter valid, required fields present
2. **Type-Specific Fields**: Required fields for handoff type
3. **Artifact Accessibility**: Source artifacts exist
4. **Consumer Parsability**: Target team can parse and understand

### Success Criteria

| Criterion | Validation Method |
|-----------|-------------------|
| YAML frontmatter parses | YAML parser succeeds |
| Required fields present | Schema validation |
| Type-specific fields present | Per-type validation |
| Source artifacts exist | File existence check |
| Items have IDs | Regex match `### [A-Z]+-[0-9]+:` |
| Items have priority | Contains `**Priority**:` |

---

## Smoke Tests: Primary Paths

### ST-001: 10x-dev-pack -> security-pack (assessment)

**Path**: Feature development requires threat modeling

**Minimal HANDOFF**:
```yaml
---
source_team: 10x-dev-pack
target_team: security-pack
handoff_type: assessment
created: 2026-01-02
initiative: Smoke Test Feature
priority: medium
---

## Context
Smoke test for security assessment handoff path.

## Source Artifacts
- docs/requirements/PRD-smoke-test.md (hypothetical)

## Items

### SEC-001: Threat model request
- **Priority**: Medium
- **Summary**: Validate handoff schema for threat modeling
- **Assessment Questions**:
  - Does handoff schema support security context?
  - Can security team parse and respond?

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: assessment` present
- [ ] Assessment Questions present in item
- [ ] Target team (security-pack) can consume

**Expected Response Type**: Threat Model or Assessment Report

---

### ST-002: 10x-dev-pack -> sre-pack (validation)

**Path**: Feature ready for production validation

**Minimal HANDOFF**:
```yaml
---
source_team: 10x-dev-pack
target_team: sre-pack
handoff_type: validation
created: 2026-01-02
initiative: Smoke Test Feature
priority: medium
---

## Context
Smoke test for production readiness validation path.

## Source Artifacts
- docs/design/TDD-smoke-test.md (hypothetical)
- src/features/smoke-test/ (hypothetical)

## Items

### VAL-001: Production readiness check
- **Priority**: Medium
- **Summary**: Validate handoff schema for SRE validation
- **Validation Scope**:
  - Observability: Metrics and logging
  - Reliability: Error handling and recovery
  - Scalability: Resource usage

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: validation` present
- [ ] Validation Scope present in item
- [ ] Target team (sre-pack) can consume

**Expected Response Type**: Validation Report with approval/rejection

---

### ST-003: 10x-dev-pack -> doc-team-pack (assessment)

**Path**: Feature needs documentation review

**Minimal HANDOFF**:
```yaml
---
source_team: 10x-dev-pack
target_team: doc-team-pack
handoff_type: assessment
created: 2026-01-02
initiative: Smoke Test Feature
priority: low
---

## Context
Smoke test for documentation assessment path.

## Source Artifacts
- docs/design/TDD-smoke-test.md (hypothetical)

## Items

### DOC-001: Documentation needs assessment
- **Priority**: Low
- **Summary**: Validate handoff schema for doc review
- **Assessment Questions**:
  - What documentation is needed?
  - What format is appropriate?

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: assessment` present
- [ ] Assessment Questions present in item
- [ ] Target team (doc-team-pack) can consume

**Expected Response Type**: Documentation plan or review

---

### ST-004: debt-triage-pack -> hygiene-pack (execution)

**Path**: Debt packages ready for remediation

**Minimal HANDOFF**:
```yaml
---
source_team: debt-triage-pack
target_team: hygiene-pack
handoff_type: execution
created: 2026-01-02
initiative: Smoke Test Debt Sprint
priority: medium
---

## Context
Smoke test for debt execution handoff path.

## Source Artifacts
- docs/debt/DEBT-LEDGER-smoke.md (hypothetical)
- docs/debt/RISK-MATRIX-smoke.md (hypothetical)

## Items

### PKG-001: Test package
- **Priority**: Medium
- **Summary**: Validate handoff schema for debt execution
- **Acceptance Criteria**:
  - Handoff can be parsed by hygiene-pack
  - Acceptance criteria are clear
  - Behavior preservation checkable

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: execution` present
- [ ] Acceptance Criteria present in item
- [ ] Target team (hygiene-pack) can consume

**Expected Response Type**: Execution report with completion status

---

### ST-005: strategy-pack -> 10x-dev-pack (implementation)

**Path**: Strategic initiative ready for build

**Minimal HANDOFF**:
```yaml
---
source_team: strategy-pack
target_team: 10x-dev-pack
handoff_type: implementation
created: 2026-01-02
initiative: Smoke Test Initiative
priority: medium
---

## Context
Smoke test for strategic implementation handoff path.

## Source Artifacts
- docs/strategy/SPEC-smoke-test.md (hypothetical)

## Items

### IMP-001: Implementation directive
- **Priority**: Medium
- **Summary**: Validate handoff schema for implementation
- **Design References**:
  - Spec: docs/strategy/SPEC-smoke-test.md
  - PoC: spikes/smoke-test/ (if applicable)
- **Implementation Notes**:
  - This is a schema validation test

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: implementation` present
- [ ] Design References present in item
- [ ] Target team (10x-dev-pack) can consume

**Expected Response Type**: Implementation complete with artifacts

---

### ST-006: intelligence-pack -> strategy-pack (strategic_input)

**Path**: Research insights for strategic planning

**Minimal HANDOFF**:
```yaml
---
source_team: intelligence-pack
target_team: strategy-pack
handoff_type: strategic_input
created: 2026-01-02
initiative: Smoke Test Research
priority: medium
---

## Context
Smoke test for strategic input handoff path.

## Source Artifacts
- docs/research/ANALYSIS-smoke.md (hypothetical)

## Items

### INS-001: Research insight
- **Priority**: Medium
- **Summary**: Validate handoff schema for strategic input
- **Data Sources**:
  - Analytics: Hypothetical data source
  - Interviews: Hypothetical user sessions
- **Confidence**: Medium (schema validation only)
- **Key Finding**: Handoff schema supports research inputs

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: strategic_input` present
- [ ] Data Sources and Confidence present in item
- [ ] Target team (strategy-pack) can consume

**Expected Response Type**: Strategic decision or further inquiry

---

### ST-007: rnd-pack -> strategy-pack (strategic_evaluation)

**Path**: Prototype ready for viability assessment

**Minimal HANDOFF**:
```yaml
---
source_team: rnd-pack
target_team: strategy-pack
handoff_type: strategic_evaluation
created: 2026-01-02
initiative: Smoke Test Prototype
priority: medium
---

## Context
Smoke test for strategic evaluation handoff path.

## Source Artifacts
- spikes/smoke-test/SPIKE-REPORT.md (hypothetical)

## Items

### EVAL-001: Prototype evaluation
- **Priority**: Medium
- **Summary**: Validate handoff schema for evaluation
- **Evaluation Criteria**:
  - Market fit: Does this solve a real problem?
  - Technical viability: Can we build this at scale?
  - Resource cost: What investment is needed?
- **Prototype Results**:
  - Spike completed successfully (hypothetical)

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: strategic_evaluation` present
- [ ] Evaluation Criteria present in item
- [ ] Target team (strategy-pack) can consume

**Expected Response Type**: Go/No-Go decision with rationale

---

### ST-008: rnd-pack -> 10x-dev-pack (implementation)

**Path**: Proven prototype ready for production

**Minimal HANDOFF**:
```yaml
---
source_team: rnd-pack
target_team: 10x-dev-pack
handoff_type: implementation
created: 2026-01-02
initiative: Smoke Test Prototype
priority: medium
---

## Context
Smoke test for R&D to implementation handoff.
Note: Typically flows through strategy-pack first.

## Source Artifacts
- spikes/smoke-test/SPIKE-REPORT.md (hypothetical)
- spikes/smoke-test/prototype/ (hypothetical)

## Items

### IMP-001: Prototype productionization
- **Priority**: Medium
- **Summary**: Validate handoff schema for prototype build
- **Design References**:
  - Spike report: spikes/smoke-test/SPIKE-REPORT.md
  - Prototype code: spikes/smoke-test/prototype/
- **Implementation Notes**:
  - Prototype validated by strategy-pack (hypothetical)

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: implementation` present
- [ ] Design References present in item
- [ ] Target team (10x-dev-pack) can consume

**Expected Response Type**: Production implementation

---

### ST-009: security-pack -> 10x-dev-pack (assessment)

**Path**: Security findings returned to dev team

**Minimal HANDOFF**:
```yaml
---
source_team: security-pack
target_team: 10x-dev-pack
handoff_type: assessment
created: 2026-01-02
initiative: Smoke Test Feature
priority: medium
---

## Context
Smoke test for security assessment response.
This is a return handoff after threat modeling.

## Source Artifacts
- docs/security/THREAT-MODEL-smoke.md (hypothetical)

## Items

### SEC-001: Threat model findings
- **Priority**: Medium
- **Summary**: Security assessment results
- **Assessment Questions** (answered):
  - STRIDE analysis complete
  - Mitigations recommended
  - Verdict: APPROVED

## Notes for Target Team
Proceed with design incorporating mitigations.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] Response references original request
- [ ] Verdict or outcome stated
- [ ] Target team (10x-dev-pack) can consume

**Expected Response Type**: N/A (this IS the response)

---

### ST-010: sre-pack -> 10x-dev-pack (validation)

**Path**: Operational findings returned to dev team

**Minimal HANDOFF**:
```yaml
---
source_team: sre-pack
target_team: 10x-dev-pack
handoff_type: validation
created: 2026-01-02
initiative: Smoke Test Feature
priority: medium
---

## Context
Smoke test for SRE validation response.
This is a return handoff after production readiness check.

## Source Artifacts
- docs/sre/VALIDATION-smoke.md (hypothetical)

## Items

### VAL-001: Production readiness findings
- **Priority**: Medium
- **Summary**: Operational validation results
- **Validation Scope** (completed):
  - Observability: PASS
  - Reliability: PASS
  - Scalability: PASS

## Notes for Target Team
Approved for production deployment.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] Response references original request
- [ ] Verdict or outcome stated
- [ ] Target team (10x-dev-pack) can consume

**Expected Response Type**: N/A (this IS the response)

---

### ST-011: hygiene-pack -> debt-triage-pack (execution)

**Path**: Remediation report back to planning team

**Minimal HANDOFF**:
```yaml
---
source_team: hygiene-pack
target_team: debt-triage-pack
handoff_type: execution
created: 2026-01-02
initiative: Smoke Test Debt Sprint
priority: low
---

## Context
Smoke test for remediation completion report.
This is a return handoff after debt execution.

## Source Artifacts
- docs/hygiene/REMEDIATION-smoke.md (hypothetical)

## Items

### RPT-001: Remediation complete
- **Priority**: Low
- **Summary**: Debt execution results
- **Acceptance Criteria** (verified):
  - All packages completed
  - Behavior preservation confirmed
  - Tests passing

## Notes for Target Team
Sprint complete. Update debt ledger.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] Response references original execution request
- [ ] Completion status stated
- [ ] Target team (debt-triage-pack) can consume

**Expected Response Type**: Debt ledger update

---

### ST-012: ecosystem-pack -> 10x-dev-pack (implementation)

**Path**: Ecosystem tooling ready for integration

**Minimal HANDOFF**:
```yaml
---
source_team: ecosystem-pack
target_team: 10x-dev-pack
handoff_type: implementation
created: 2026-01-02
initiative: Smoke Test Tooling
priority: medium
---

## Context
Smoke test for ecosystem tooling handoff.

## Source Artifacts
- .claude/skills/smoke-test/SKILL.md (hypothetical)

## Items

### IMP-001: Tooling integration
- **Priority**: Medium
- **Summary**: Integrate ecosystem improvements
- **Design References**:
  - Skill: .claude/skills/smoke-test/SKILL.md
  - Migration: docs/ecosystem/MIGRATION-smoke.md
- **Implementation Notes**:
  - Follow migration guide

## Notes for Target Team
This is a schema validation test.
```

**Validation Checklist**:
- [ ] Frontmatter parses as valid YAML
- [ ] `handoff_type: implementation` present
- [ ] Design References present
- [ ] Target team (10x-dev-pack) can consume

**Expected Response Type**: Integration complete

---

## Smoke Tests: Secondary Paths

### ST-013 through ST-018

Secondary paths follow the same structure. Key validation points:

| Test | Source | Target | Type | Key Validation |
|------|--------|--------|------|----------------|
| ST-013 | 10x | intelligence | assessment | Research request parsable |
| ST-014 | 10x | rnd | assessment | Spike request parsable |
| ST-015 | strategy | rnd | implementation | Research directive parsable |
| ST-016 | doc-team | 10x | assessment | Doc feedback parsable |
| ST-017 | sre | security | assessment | Incident escalation parsable |
| ST-018 | security | sre | validation | Control validation parsable |

---

## Automated Validation Script

### Schema Validator (Conceptual)

```bash
#!/bin/bash
# handoff-smoke-test.sh

HANDOFF_FILE=$1

# Validate YAML frontmatter
if ! yq e '.source_team' "$HANDOFF_FILE" > /dev/null 2>&1; then
  echo "FAIL: Invalid YAML frontmatter"
  exit 1
fi

# Check required fields
for field in source_team target_team handoff_type created initiative; do
  if [ -z "$(yq e ".$field" "$HANDOFF_FILE")" ]; then
    echo "FAIL: Missing required field: $field"
    exit 1
  fi
done

# Validate team names
source=$(yq e '.source_team' "$HANDOFF_FILE")
target=$(yq e '.target_team' "$HANDOFF_FILE")

if [[ ! "$source" =~ ^[a-z]+-pack$ ]]; then
  echo "FAIL: Invalid source_team format: $source"
  exit 1
fi

if [[ ! "$target" =~ ^[a-z]+-pack$ ]]; then
  echo "FAIL: Invalid target_team format: $target"
  exit 1
fi

# Check for self-handoff
if [ "$source" = "$target" ]; then
  echo "FAIL: Self-handoff not allowed"
  exit 1
fi

# Validate handoff type
type=$(yq e '.handoff_type' "$HANDOFF_FILE")
valid_types="execution validation assessment implementation strategic_input strategic_evaluation"
if [[ ! " $valid_types " =~ " $type " ]]; then
  echo "FAIL: Invalid handoff_type: $type"
  exit 1
fi

# Check for Context section
if ! grep -q "^## Context" "$HANDOFF_FILE"; then
  echo "FAIL: Missing Context section"
  exit 1
fi

# Check for Items section
if ! grep -q "^## Items" "$HANDOFF_FILE"; then
  echo "FAIL: Missing Items section"
  exit 1
fi

# Check for item IDs
if ! grep -q "^### [A-Z]\+-[0-9]\+:" "$HANDOFF_FILE"; then
  echo "FAIL: No valid item IDs found"
  exit 1
fi

echo "PASS: Handoff schema valid"
exit 0
```

### Batch Validation

```bash
# Validate all handoffs in directory
find . -name "HANDOFF-*.md" -exec ./handoff-smoke-test.sh {} \;
```

---

## Test Execution Summary

### Manual Test Procedure

1. Create minimal HANDOFF artifact per test definition
2. Store in appropriate location
3. Run schema validation script
4. Verify target team can parse
5. Document result

### CI Integration (Future)

```yaml
# .github/workflows/handoff-validation.yml
name: Handoff Schema Validation

on:
  push:
    paths:
      - '**/HANDOFF-*.md'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install yq
        run: sudo snap install yq
      - name: Validate handoffs
        run: |
          find . -name "HANDOFF-*.md" -exec ./scripts/handoff-smoke-test.sh {} \;
```

---

## Pass/Fail Summary Template

| Test | Status | Notes |
|------|--------|-------|
| ST-001 | [ ] PASS / [ ] FAIL | |
| ST-002 | [ ] PASS / [ ] FAIL | |
| ST-003 | [ ] PASS / [ ] FAIL | |
| ST-004 | [ ] PASS / [ ] FAIL | |
| ST-005 | [ ] PASS / [ ] FAIL | |
| ST-006 | [ ] PASS / [ ] FAIL | |
| ST-007 | [ ] PASS / [ ] FAIL | |
| ST-008 | [ ] PASS / [ ] FAIL | |
| ST-009 | [ ] PASS / [ ] FAIL | |
| ST-010 | [ ] PASS / [ ] FAIL | |
| ST-011 | [ ] PASS / [ ] FAIL | |
| ST-012 | [ ] PASS / [ ] FAIL | |
| ST-013 | [ ] PASS / [ ] FAIL | |
| ST-014 | [ ] PASS / [ ] FAIL | |
| ST-015 | [ ] PASS / [ ] FAIL | |
| ST-016 | [ ] PASS / [ ] FAIL | |
| ST-017 | [ ] PASS / [ ] FAIL | |
| ST-018 | [ ] PASS / [ ] FAIL | |

**Overall**: ___/18 tests passing

---

## Related Documents

- [Cross-Team Handoff Schema](../../.claude/skills/shared/cross-team-handoff/schema.md)
- [Cross-Team Coordination Playbook](../playbooks/cross-rite-coordination.md)
- [Edge Cases: Cross-Team Workflows](../edge-cases/cross-team-workflows.md)
- [E2E Test: Feature Development](e2e-feature-development.md)
- [E2E Test: Security Workflow](e2e-security-workflow.md)
- [E2E Test: Debt Remediation](e2e-debt-remediation.md)

# Compatibility Report Template

> Cross-satellite validation results.

```markdown
# Compatibility Report: [Change Title]

## Test Summary
**Date**: [YYYY-MM-DD]
**Complexity**: [PATCH | MODULE | SYSTEM | MIGRATION]
**CEM Version**: [version]
**roster Version**: [version]
**Tester**: [agent/person]

**Status**: [APPROVED | REJECTED | APPROVED WITH CAVEATS]

## Test Matrix Results

| Satellite | Config Type | roster-sync | Hook Reg | Settings Merge | Migration | Status |
|-----------|-------------|-------------|----------|----------------|-----------|--------|
| test-baseline | Minimal | Pass | Pass | Pass | Pass | PASS |
| satellite-a | Standard | Pass | Pass | Fail | N/A | FAIL |
| satellite-b | Complex | Pass | Warn | Pass | Pass | WARN |

**Legend**:
- Pass: Works as expected
- Fail: Broken, blocks release
- Warn: Works with caveats or minor issues
- N/A: Not applicable to this satellite

## Defects

### P0 Defects (Critical)
None

### P1 Defects (High)
#### DEF-001: Settings merge fails with nested null values
**Severity**: P1
**Satellite**: satellite-a (standard config)
**Reproduction**:
1. Configure settings with `"custom": {"nested": null}`
2. Run `ari sync`
3. Error: "jq: null cannot be iterated"

**Expected**: Merge succeeds, null preserved
**Actual**: Sync fails with jq error
**Impact**: Blocks satellites using null config values

### P2 Defects (Medium)
[Details]

### P3 Defects (Low)
None

## Migration Runbook Validation

### Runbook Issues Found
- Step 3: "Apply transformation" is vague--needs exact commands
- Verification step output example doesn't match actual output
- Rollback tested successfully in all satellites

### Recommended Runbook Changes
1. Add explicit jq command for step 3
2. Update verification output example to match v2.0 format

## Regression Testing

**Pre-existing functionality tested**:
- Agent invocation from commands
- Skill loading on session start
- Settings tier precedence (global < team < user)
- Hook lifecycle events (warning in complex config, see DEF-002)

**Regressions found**: None beyond DEF-002

## Backward Compatibility Verification

| CEM | roster | Tested | Result | Notes |
|-----|--------|--------|--------|-------|
| 2.0 | 2.0 | Yes | Pass | Fully compatible |
| 2.0 | 1.9 | Yes | Pass | Backward compatible confirmed |
| 1.9 | 2.0 | Yes | Fail | Settings merge incompatible (expected) |

## Rollout Recommendation

**Decision**: [APPROVED | REJECTED] for release

**Rationale**:
- [Key points supporting decision]

**Required for approval** (if rejected):
1. [Fix requirement]
2. [Retest requirement]

**P2/P3 defects**: Can defer to v2.0.1 patch release

## Notes for Next Iteration
- [Learnings and improvements]
```

## Quality Gate

**Compatibility Report complete when:**
- All satellites tested
- Defects prioritized by severity
- Recommendation justified
- Runbook validated against actual behavior

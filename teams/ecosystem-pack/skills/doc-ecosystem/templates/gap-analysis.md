# Gap Analysis Template

> Issue diagnosis for CEM/satellite problems.

```markdown
# Gap Analysis: [Issue Title]

## Executive Summary
[2-3 sentences: what's broken, impact, root cause]

## Reproduction Steps
1. [Step with exact commands]
2. [Expected vs. actual behavior]

## Root Cause
**Component**: [CEM | skeleton | roster]
**File**: [path/to/file:line]
**Issue**: [technical explanation]

## Success Criteria
- [ ] [Concrete, testable criterion]
- [ ] [e.g., "cem sync completes without errors"]

## Affected Systems
- [x] CEM
- [ ] skeleton
- [ ] roster

## Recommended Complexity
**Level**: [PATCH | MODULE | SYSTEM | MIGRATION]
**Rationale**: [why this complexity]

## Test Satellites
- skeleton (always)
- [other satellites based on issue characteristics]

## Notes for Context Architect
[Anything relevant for design phase]
```

## Quality Gate

**Gap Analysis complete when:**
- Clear reproduction steps provided
- Root cause identified with file/line reference
- Success criteria are testable
- Complexity recommendation justified

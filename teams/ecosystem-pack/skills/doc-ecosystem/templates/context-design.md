# Context Design Template

> Technical design for ecosystem changes.

```markdown
# Context Design: [Solution Title]

## Overview
[2-3 sentences: what we're building, why this approach]

## Architecture

### Components Affected
- **CEM**: [what changes, why]
- **skeleton**: [what changes, why]
- **roster**: [what changes, why]

### Design Decisions
[Key architectural choices and rationale]

## Schema Definitions (if applicable)

### [Hook/Skill/Agent] Schema
```yaml
# Schema structure with comments
name: string
version: string
lifecycle:
  - event: string
    action: string
```

**Validation Rules**:
- [Rule 1]
- [Rule 2]

## Implementation Specification

### CEM Changes
**File**: `path/to/file`
**Function**: `function_name`
**Changes**: [detailed specification]

### skeleton Changes
**File**: `path/to/file`
**Changes**: [detailed specification]

### roster Changes
**Location**: `path/to/content`
**Changes**: [detailed specification]

## Backward Compatibility

**Classification**: [COMPATIBLE | BREAKING]

**Migration Path** (if breaking):
1. [Step-by-step satellite upgrade process]

**Deprecation Timeline** (if applicable):
- Version N: New pattern available, old pattern deprecated
- Version N+1: Old pattern removed

**Compatibility Matrix**:
| CEM Version | skeleton Version | Status |
|-------------|------------------|--------|
| 2.0 | 2.0 | Supported |
| 2.0 | 1.9 | Backward compatible |

## Integration Test Matrix

| Satellite | Test Case | Expected Outcome | Validates |
|-----------|-----------|------------------|-----------|
| skeleton | `cem sync` | No conflicts | Basic compatibility |
| [satellite-2] | Hook registration | Fires on event | Schema enforcement |

## Notes for Integration Engineer
[Implementation hints, gotchas, suggested approach]
```

## Quality Gate

**Context Design complete when:**
- Backward compatibility assessed
- Migration path documented (if breaking)
- Integration tests defined
- Implementation specification sufficient for engineer

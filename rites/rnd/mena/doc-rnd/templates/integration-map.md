# Integration Map Template

> Integration analysis with current/target architecture, gap analysis, and effort estimation.

```markdown
# INTEGRATE-{slug}

## Overview
{What we're integrating and why}

## Current State

### Architecture
{Description or diagram of current system}

### Integration Points
| System | Interface | Data | Frequency |
|--------|-----------|------|-----------|
| {system} | {REST/gRPC/etc.} | {what flows} | {calls/day} |

### Dependencies
- {Dependency 1}: {How it's used}
- {Dependency 2}: {How it's used}

## Target State

### New Architecture
{Description or diagram of target system}

### New Integration Points
| System | Interface | Data | Changes |
|--------|-----------|------|---------|
| {system} | {new interface} | {what flows} | {what changes} |

## Gap Analysis

### API Compatibility
| Feature | Current | New | Compatibility | Notes |
|---------|---------|-----|---------------|-------|
| {feature} | {current API} | {new API} | {Full/Partial/None} | {details} |

### Breaking Changes
1. {Breaking change 1}: {Impact and mitigation}
2. {Breaking change 2}: {Impact and mitigation}

### Hidden Dependencies
| Dependency | Impact | Discovery |
|------------|--------|-----------|
| {dep} | {what breaks} | {how we found it} |

## Effort Estimate

| Component | Effort | Confidence | Notes |
|-----------|--------|------------|-------|
| {component} | {days/weeks} | {High/Medium/Low} | {assumptions} |

**Total Estimated Effort**: {X} person-weeks

## Risks
| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| {risk} | {H/M/L} | {H/M/L} | {strategy} |

## Integration Approach

### Option A: {Approach Name}
- **Pros**: {benefits}
- **Cons**: {drawbacks}
- **Effort**: {estimate}

### Option B: {Approach Name}
...

### Recommendation
{Which approach and why}

## Migration Plan
1. **Phase 1**: {What and duration}
2. **Phase 2**: {What and duration}
3. **Phase 3**: {What and duration}

## Success Criteria
- [ ] {Criterion 1}
- [ ] {Criterion 2}
```

## Quality Gate

**Integration Map complete when:**
- Current and target architecture both documented
- Breaking changes identified with mitigations
- Effort estimates include confidence levels
- At least two integration approaches compared

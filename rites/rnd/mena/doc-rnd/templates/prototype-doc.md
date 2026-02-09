# Prototype Documentation Template

> Prototype results, feasibility verdict, and production path with deliberate shortcuts documented.

```markdown
# PROTO-{slug}

## Executive Summary
{What was built and what it proves in 2-3 sentences}

## Decision Enabled
{What decision can now be made with this prototype}

## Prototype Scope

### What It Does
- {Capability 1}
- {Capability 2}

### What It Doesn't Do
- {Limitation 1}
- {Limitation 2}

### Deliberate Shortcuts
| Shortcut | Production Alternative |
|----------|----------------------|
| {hardcoded thing} | {what it should be} |
| {simulated thing} | {real implementation} |

## Technical Approach

### Architecture
{Diagram or description}

### Key Technologies
- {Tech 1}: {Why chosen}
- {Tech 2}: {Why chosen}

### Integration Points
- {How it connects to our systems}

## Results

### What Worked
- {Success 1}
- {Success 2}

### What Didn't Work
- {Failure 1}: {Why and implications}
- {Failure 2}: {Why and implications}

### Performance
| Metric | Result | Target | Notes |
|--------|--------|--------|-------|
| {metric} | {value} | {goal} | {context} |

### Discovered Constraints
- {Constraint 1}: {Impact on production}
- {Constraint 2}: {Impact}

## Feasibility Assessment

### Verdict
{Feasible / Feasible with caveats / Not feasible}

### Confidence
{High / Medium / Low}

### Key Risks for Production
1. {Risk 1}
2. {Risk 2}

## Production Path

### Required Changes
| Prototype | Production |
|-----------|------------|
| {what we did} | {what we'd need} |

### Effort Estimate
{Rough estimate for production implementation}

### Recommended Next Steps
1. {Step 1}
2. {Step 2}

## Demo Guide

### Prerequisites
{What needs to be set up}

### Demo Script
1. {Step 1: Show X}
2. {Step 2: Demonstrate Y}
3. {Step 3: Highlight Z}

### FAQ
- Q: {Common question}
- A: {Answer}

## Repository
{Link to prototype code}

## Appendix
- Setup instructions
- Known issues
- Future ideas
```

## Quality Gate

**Prototype Documentation complete when:**
- Decision enabled is clearly stated
- Deliberate shortcuts documented with production alternatives
- Feasibility verdict has confidence level
- Production path includes effort estimate

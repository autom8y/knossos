# Threat Model Template

> STRIDE/DREAD analysis with data flow diagrams, threat identification, and mitigation tracking.

```markdown
# THREAT-{slug}

## Executive Summary
{One paragraph overview of security posture}

## Scope
- **Assets**: {What we're protecting}
- **Threat Actors**: {Who might attack}
- **Trust Boundaries**: {Where trust changes}

## Data Flow Diagram
{ASCII or description of data flows}

## Threat Analysis

### STRIDE Analysis
| Component | S | T | R | I | D | E | Notes |
|-----------|---|---|---|---|---|---|-------|
| {component} | {rating} | ... |

### Identified Threats

#### THREAT-001: {Name}
- **Category**: {STRIDE category}
- **DREAD Score**: {D+R+E+A+D = total}
- **Attack Vector**: {How it would be exploited}
- **Impact**: {What damage results}
- **Mitigation**: {How to prevent/detect}
- **Status**: {Open/Mitigated/Accepted}

## Recommendations
1. {Priority 1 mitigation}
2. {Priority 2 mitigation}

## Residual Risks
{Threats accepted or deferred}
```

## Quality Gate

**Threat Model complete when:**
- STRIDE analysis covers all components in scope
- Each threat has DREAD score and attack vector
- Mitigations specified for all non-accepted threats
- Residual risks explicitly documented

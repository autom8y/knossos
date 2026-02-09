# Security Signoff Template

> Security review checklist and approval workflow for code changes touching sensitive areas.

```markdown
# SEC-{slug}

## Review Summary
- **PR/Change**: {Link or description}
- **Reviewed By**: security-reviewer
- **Date**: {date}
- **Decision**: {Approved / Approved with Conditions / Rejected}

## Change Overview
{What this change does from security perspective}

## Security Scope

### High-Risk Areas Touched
- [ ] Authentication
- [ ] Authorization
- [ ] Input handling
- [ ] Data encryption
- [ ] PII processing
- [ ] External integrations
- [ ] Secrets management

### Threat Model Reference
{Link to relevant threat model if exists}

## Findings

### Critical (Must Fix)
{None or list}

### High (Should Fix)
{None or list}

### Medium (Recommend Fix)
{None or list}

### Low (Consider)
{None or list}

### Finding Details

#### FINDING-001: {Title}
- **Severity**: {Critical/High/Medium/Low}
- **Location**: {File:line}
- **Issue**: {Description}
- **Risk**: {What could go wrong}
- **Recommendation**: {How to fix}
- **Status**: {Open/Fixed/Accepted}

## Security Checklist

### Authentication
- [ ] Auth required for sensitive endpoints
- [ ] Token validation correct
- [ ] Session handling secure

### Authorization
- [ ] Access controls enforced
- [ ] No privilege escalation paths
- [ ] Role checks in place

### Input Validation
- [ ] All inputs validated
- [ ] Proper encoding/escaping
- [ ] No injection vulnerabilities

### Data Protection
- [ ] Sensitive data encrypted
- [ ] No secrets in code
- [ ] Logging doesn't leak PII

### Error Handling
- [ ] No stack traces exposed
- [ ] Error messages safe
- [ ] Failures secure

## Decision

### Verdict
{Approved / Approved with Conditions / Rejected}

### Conditions (if applicable)
1. {Condition 1}
2. {Condition 2}

### Rationale
{Why this decision}

### Risk Acceptance (if applicable)
{Any risks being accepted and why}

## Follow-up Required
- [ ] {Action item if any}

## Sign-off
Reviewed and approved/rejected by Security Reviewer on {date}.
```

## Quality Gate

**Security Signoff complete when:**
- All high-risk area checkboxes assessed
- Findings categorized by severity
- Verdict includes rationale
- Risk acceptances explicitly documented

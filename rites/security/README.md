# Security Pack

Security assessment, compliance mapping, penetration testing, and review for systems handling auth, crypto, PII, or external integrations.

## When to Use This Rite

**Triggers**:
- "We need to security-review this new auth system"
- "Does this feature meet SOC 2 requirements?"
- "Can you pentest our API before launch?"
- "Is this PR safe to merge from a security perspective?"

**Not for**: General code review without security implications, performance optimization, or feature development.

## Quick Start

```bash
/rite security
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| threat-modeler | Maps attack vectors and threat landscapes | threat-model |
| compliance-architect | Translates regulations into engineering requirements | compliance-requirements |
| penetration-tester | Probes systems for vulnerabilities like an authorized adversary | pentest-report |
| security-reviewer | Final security gate before code merges | security-signoff |

## Workflow

Sequential workflow with complexity-based phase skipping:
- **PATCH**: penetration-testing → security-review
- **FEATURE**: threat-modeling → compliance-design → penetration-testing → security-review
- **SYSTEM**: All phases (auth systems, cryptography, external integrations, PII)

See `workflow.md` for phase flow and complexity levels.

## Related Rites

- **10x-dev**: Hand off security-approved features for implementation
- **hygiene**: When security findings need systematic testing validation

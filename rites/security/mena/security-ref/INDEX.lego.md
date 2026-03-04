---
name: security-ref
description: "Security rite reference. Use when: activating the security rite, invoking security agents, performing threat modeling or compliance reviews. Triggers: security, threat-modeler, compliance-architect, penetration-tester, security-reviewer."
---

# Security Rite (security)

> Map threats. Enforce compliance. Test defenses. Guard the gate.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Agents | `.claude/agents/` | Agent prompts |
| Workflow | `.claude/ACTIVE_RITE` | Phase configuration |
| Switch | `/security` | Activate this rite |

## Pantheon

| Agent | Model | Role | Produces |
|-------|-------|------|----------|
| **threat-modeler** | opus | Maps attack vectors with STRIDE/DREAD | threat-model |
| **compliance-architect** | opus | Translates regulations to requirements | compliance-requirements |
| **penetration-tester** | opus | Probes systems for vulnerabilities | pentest-report |
| **security-reviewer** | opus | Final security gate before merge | security-signoff |

## Workflow

```
threat-modeling → compliance-design → penetration-testing → security-review
       │                 │                    │                   │
       ▼                 ▼                    ▼                   ▼
  THREAT-{slug}    COMPLY-{slug}       PENTEST-{slug}        SEC-{slug}
```

### Phase Details

| Phase | Agent | Input | Output |
|-------|-------|-------|--------|
| threat-modeling | threat-modeler | Feature specs | Threat model with STRIDE analysis |
| compliance-design | compliance-architect | Threat model | Control requirements |
| penetration-testing | penetration-tester | Control requirements | Vulnerability findings |
| security-review | security-reviewer | Pentest report | Merge approval/rejection |

## Complexity Levels

| Level | Scope | Phases |
|-------|-------|--------|
| **PATCH** | Single file, no auth/crypto | penetration-testing, security-review |
| **FEATURE** | New endpoints, data handling | All phases |
| **SYSTEM** | Auth, crypto, external integrations | All phases |

## Command Mapping

| Command | Maps To | Use When |
|---------|---------|----------|
| `/security` | Rite switch | Activating this rite |
| `/architect` | compliance-architect | Compliance mapping only |
| `/build` | penetration-tester | Testing only |
| `/qa` | security-reviewer | Review only |
| `/hotfix` | penetration-tester | Quick security fix |
| `/code-review` | security-reviewer | Security code review |

## When to Use This Rite

**Use security when:**
- New feature touches auth, crypto, or PII
- Preparing for compliance audit
- Need pre-release security validation
- Reviewing code with security implications

**Don't use security when:**
- Pure documentation changes → Use docs
- General code quality → Use hygiene
- Feature development → Use 10x-dev

## Agent Summaries

### Threat Modeler

**Purpose**: Map attack vectors before code ships

**Key Responsibilities**:
- STRIDE/DREAD analysis
- Kill chain mapping
- Attack surface enumeration
- Risk prioritization

**Produces**: `.ledge/reviews/THREAT-{slug}.md`

---

### Compliance Architect

**Purpose**: Translate regulations into engineering requirements

**Key Responsibilities**:
- Control mapping (SOC 2, GDPR, HIPAA)
- Gap analysis
- Evidence architecture
- Audit preparation

**Produces**: `.ledge/reviews/COMPLY-{slug}.md`

---

### Penetration Tester

**Purpose**: Probe systems like an authorized adversary

**Key Responsibilities**:
- Vulnerability discovery
- Exploit development
- Attack path documentation
- Remediation guidance

**Produces**: `.ledge/reviews/PENTEST-{slug}.md`

---

### Security Reviewer

**Purpose**: Final security gate before merge

**Key Responsibilities**:
- Security code review
- Fix validation
- Release approval
- Pattern recognition

**Produces**: `.ledge/reviews/SEC-{slug}.md`

## Cross-References

- **Related Skills**: standards (secure coding)
- **Related Rites**: 10x-dev (implementation), sre (infrastructure security)
- **Commands**: Run `ari rite --list` or `/consult --commands`

---
name: doc-security
description: "Security templates for threat modeling, compliance mapping, penetration testing, and security review. Use when: performing threat analysis, mapping compliance requirements, documenting pentest findings, approving security changes. Triggers: threat model, STRIDE, DREAD, compliance, SOC 2, GDPR, pentest, security review, security signoff."
---

# Security Documentation Templates

> Templates for security analysis, compliance, testing, and approval workflows.

## Purpose

Provides structured templates for security workflows: STRIDE/DREAD threat modeling, regulatory compliance mapping with evidence collection, penetration test reporting with reproduction steps, and security review checklists for code changes.

## Template Catalog

| Template | Purpose | Agent |
|----------|---------|-------|
| [Threat Model](templates/threat-model.md) | STRIDE/DREAD analysis with data flow and mitigation tracking | security-analyst |
| [Compliance Requirements](templates/compliance-requirements.md) | Regulatory mapping, control implementation, evidence collection | compliance-analyst |
| [Pentest Report](templates/pentest-report.md) | Vulnerability findings with PoC and remediation guidance | security-tester |
| [Security Signoff](templates/security-signoff.md) | Code review checklist and release approval | security-reviewer |

## When to Use Each Template

| Scenario | Template |
|----------|----------|
| Designing security for new system | Threat Model |
| Preparing for SOC 2 / GDPR audit | Compliance Requirements |
| Documenting security test results | Pentest Report |
| Reviewing PR touching auth/crypto/PII | Security Signoff |
| Assessing attack surface | Threat Model |
| Building evidence for auditors | Compliance Requirements |

## Quality Gates Summary

| Template | Gate Criteria |
|----------|---------------|
| **Threat Model** | STRIDE covers all components, DREAD scores assigned, mitigations specified |
| **Compliance Requirements** | Controls mapped to regulations, evidence collection planned, gaps prioritized |
| **Pentest Report** | CVSS scores assigned, reproduction steps provided, attack paths documented |
| **Security Signoff** | All high-risk checkboxes assessed, verdict with rationale, risk acceptances documented |

## Progressive Disclosure

- [threat-model.md](templates/threat-model.md) - STRIDE/DREAD analysis (THREAT-{slug})
- [compliance-requirements.md](templates/compliance-requirements.md) - Regulatory mapping (COMPLY-{slug})
- [pentest-report.md](templates/pentest-report.md) - Vulnerability findings (PENTEST-{slug})
- [security-signoff.md](templates/security-signoff.md) - Security approval (SEC-{slug})

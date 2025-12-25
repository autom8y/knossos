# Security Pack Workflow

## Phase Flow

```
┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐
│     Threat        │─────▶│    Compliance     │─────▶│   Penetration     │─────▶│     Security      │
│     Modeler       │      │    Architect      │      │      Tester       │      │     Reviewer      │
└───────────────────┘      └───────────────────┘      └───────────────────┘      └───────────────────┘
   Threat Model         Compliance Reqmts          Pentest Report          Security Signoff
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| threat-modeling | threat-modeler | Threat Model | User request |
| compliance-design | compliance-architect | Compliance Requirements | Threat model complete, complexity >= FEATURE |
| penetration-testing | penetration-tester | Pentest Report | Compliance requirements approved (or threat model if PATCH) |
| security-review | security-reviewer | Security Signoff | Pentest report complete |

## Complexity Levels

- **PATCH**: Single file change, no auth/crypto/PII impact
  - Phases: penetration-testing, security-review
- **FEATURE**: New endpoints, data handling, session management
  - Phases: threat-modeling, compliance-design, penetration-testing, security-review
- **SYSTEM**: Auth systems, cryptography, external integrations, PII processing
  - Phases: threat-modeling, compliance-design, penetration-testing, security-review

## Phase Skipping

At PATCH complexity, threat-modeling and compliance-design phases are skipped for low-risk changes.

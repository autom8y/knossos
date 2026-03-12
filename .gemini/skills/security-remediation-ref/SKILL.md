---
description: 'security-remediation procession reference. Use when: navigating security-remediation stations, understanding station goals, checking procession progress. Triggers: security-remediation, security-remediation-ref, security-remediation stations.'
name: security-remediation-ref
version: "1.0"
---
---
name: security-remediation-ref
description: "security-remediation procession reference. Use when: navigating security-remediation stations, understanding station goals, checking procession progress. Triggers: security-remediation, security-remediation-ref, security-remediation stations."
---

# security-remediation Procession Reference

Security findings lifecycle: audit, assess, plan, remediate, validate

## Station Map

| # | Station | Rite | Alt Rite | Goal | Produces | Loop To |
|---|---------|------|----------|------|----------|---------|
| 1 | audit | security | - | Map attack surface, classify findings by exploitability, ... | threat-model, pentest-report | - |
| 2 | assess | debt-triage | - | Catalog findings, score risk, produce prioritized remedia... | debt-inventory, priority-matrix | - |
| 3 | plan | debt-triage | - | Group findings into sprint-sized tasks with acceptance cr... | sprint-plan | - |
| 4 | remediate | hygiene | 10x-dev | Execute remediation plan, produce PRs with fixes | remediation-ledger | - |
| 5 | validate | security | - | Review remediation PRs for security correctness | validation-report | remediate |


## Station Goals

- **audit** (security): Map attack surface, classify findings by exploitability, produce threat model and pentest report
- **assess** (debt-triage): Catalog findings, score risk, produce prioritized remediation backlog
- **plan** (debt-triage): Group findings into sprint-sized tasks with acceptance criteria
- **remediate** (hygiene, alt: 10x-dev): Execute remediation plan, produce PRs with fixes
- **validate** (security): Review remediation PRs for security correctness


## Workflow

- **Artifact directory**: `.sos/wip/security-remediation/`
- **Total stations**: 5
- **Entry point**: `/security-remediation`
- **First station**: audit (security rite)

## Handoff Artifacts

Each station transition produces a handoff artifact at:
`.sos/wip/security-remediation//HANDOFF-{source}-to-{target}.md`

See `procession-ref` skill for the handoff schema and transition protocol.

## CLI Commands

| Command | Description |
|---------|-------------|
| `ari procession status` | Show current procession state |
| `ari procession proceed` | Advance to next station |
| `ari procession recede --to={station}` | Roll back to a previous station |
| `ari procession abandon` | Terminate the procession |

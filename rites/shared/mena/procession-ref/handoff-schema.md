---
description: "Handoff Artifact Schema companion for procession-ref skill."
---

# Handoff Artifact Schema

Procession handoff artifacts are Markdown files with YAML frontmatter. They live at `{artifact_dir}/HANDOFF-{source_station}-to-{target_station}.md`.

## Required Frontmatter Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string (const `"handoff"`) | Identifies this as a procession handoff artifact |
| `procession_id` | string | Procession instance ID. Pattern: `{template}-{YYYY-MM-DD}` |
| `source_station` | string | Station that produced this handoff. Lowercase kebab-case |
| `source_rite` | string | Rite active during the source station |
| `target_station` | string | Station that will consume this handoff. Lowercase kebab-case |
| `target_rite` | string | Rite expected for the target station |
| `produced_at` | string | ISO 8601 timestamp |
| `artifacts` | array | Work products produced during this station (see below) |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `acceptance_criteria` | array of strings | Directional criteria for the target station's work |

## Artifact Reference Object

Each entry in the `artifacts` array:

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Artifact type identifier (e.g., `threat-model`, `pentest-report`) |
| `path` | string | Path relative to project root |

## Validation

- `procession_id` must match pattern: `^[a-z][a-z0-9-]*-\d{4}-\d{2}-\d{2}$`
- `source_station` and `target_station` must match: `^[a-z][a-z0-9-]*$`
- `artifacts` must contain at least one entry
- No additional properties allowed in frontmatter

Schema source: `internal/validation/schemas/handoff-procession.schema.json`

## Body Content

The Markdown body below the frontmatter is free-form. It should contain:

- **Context**: What was discovered or accomplished at this station
- **Key findings**: Summary of work products (not duplicating the artifacts themselves)
- **Guidance for next station**: What the target station should focus on
- **Open questions**: Anything unresolved that the target station should address

## Example

A handoff from the `audit` station to the `assess` station in the security-remediation procession:

```markdown
---
type: handoff
procession_id: security-remediation-2026-03-10
source_station: audit
source_rite: security
target_station: assess
target_rite: debt-triage
produced_at: "2026-03-10T14:30:00Z"
artifacts:
  - type: threat-model
    path: .sos/wip/security-remediation/threat-model.md
  - type: pentest-report
    path: .sos/wip/security-remediation/pentest-report.md
acceptance_criteria:
  - "All findings from threat model cataloged in debt inventory"
  - "Risk scores assigned using CVSS or equivalent"
  - "Remediation backlog prioritized by exploitability"
---

# Audit -> Assess Handoff

## Context

Security audit completed against the API surface and authentication layer.
Identified 12 findings across 3 severity levels.

## Key Findings

- 2 critical: SQL injection in user search, broken access control on admin endpoints
- 4 high: session fixation, missing rate limiting, verbose error messages, outdated TLS config
- 6 medium: missing security headers, permissive CORS, debug endpoints exposed

## Guidance for Assess Station

Focus on the 2 critical findings first. The threat model contains full
exploit chains. The pentest report has reproduction steps for all 12 findings.

## Open Questions

- Should third-party dependencies with known CVEs be included in the remediation scope?
- Is there a compliance deadline driving the timeline?
```

## Station Names (security-remediation template)

For reference, the security-remediation procession defines these stations:

| Station | Rite | Produces |
|---------|------|----------|
| `audit` | security | threat-model, pentest-report |
| `assess` | debt-triage | debt-inventory, priority-matrix |
| `plan` | debt-triage | sprint-plan |
| `remediate` | hygiene (alt: 10x-dev) | remediation-ledger |
| `validate` | security | validation-report |

The `validate` station can loop back to `remediate` if findings remain unresolved.

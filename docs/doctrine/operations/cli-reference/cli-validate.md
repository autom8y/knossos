---
last_verified: 2026-02-26
---

# CLI Reference: validate

> Validate workflow artifacts against schemas.

Validates PRD, TDD, ADR, Test Plans against schemas and handoff criteria.

**Family**: validate
**Commands**: 3
**Priority**: MEDIUM

---

## Commands

### ari validate artifact

Validate an artifact file against its schema.

**Synopsis**:
```bash
ari validate artifact [flags]
```

**Description**:
Validates that an artifact file conforms to its expected schema. Checks structure, required fields, and content quality.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | - | Artifact type: prd, tdd, adr, test-plan |

**Examples**:
```bash
# Validate PRD
ari validate artifact --type=prd docs/requirements/PRD-user-auth.md

# Validate TDD
ari validate artifact --type=tdd docs/design/TDD-user-auth.md

# JSON output
ari validate artifact --type=prd docs/requirements/PRD-foo.md -o json
```

**Related Commands**:
- [`ari validate handoff`](#ari-validate-handoff) — Validate handoff criteria
- [`ari artifact register`](cli-artifact.md#ari-artifact-register) — Register after validation

---

### ari validate handoff

Validate handoff criteria for phase transitions.

**Synopsis**:
```bash
ari validate handoff [flags]
```

**Description**:
Validates that handoff criteria are met for transitioning between workflow phases. Checks required artifacts and quality gates.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--phase` | string | - | Target phase: requirements, design, implementation, validation |
| `--artifact` | string | - | Artifact being handed off |

**Examples**:
```bash
# Validate handoff to design phase
ari validate handoff --phase=design --artifact=PRD-user-auth

# Validate handoff to implementation
ari validate handoff --phase=implementation --artifact=TDD-user-auth
```

**Related Commands**:
- [`ari handoff prepare`](cli-handoff.md#ari-handoff-prepare) — Prepare handoff
- [`ari session transition`](cli-session.md#ari-session-transition) — Phase transition

---

### ari validate schema

Validate a file against a specific schema.

**Synopsis**:
```bash
ari validate schema [flags]
```

**Description**:
Validates a file against a specific JSON schema. Used for context files and custom artifacts.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--file` | string | - | File to validate |

**Examples**:
```bash
# Validate session context
ari validate schema --file=SESSION_CONTEXT.md

# Validate sprint context
ari validate schema --file=SPRINT_CONTEXT.md
```

**Related Commands**:
- [Moirai](../../reference/GLOSSARY.md#moirai) — Context file authority

---

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `$XDG_CONFIG_HOME/ariadne/config.yaml` | Config file path |
| `-o, --output` | string | `text` | Output format: text, json, yaml |
| `-p, --project-dir` | string | auto-discovered | Project root directory |
| `-s, --session-id` | string | current session | Override session ID |
| `-v, --verbose` | bool | false | Enable verbose output |

---

## See Also

- [Artifact Schemas](../../reference/GLOSSARY.md#artifact)
- [Handoff Validation](cli-handoff.md)

---
last_verified: 2026-02-26
---

# CLI Reference: artifact

> Register, query, and manage workflow artifacts.

[Artifacts](../../reference/GLOSSARY.md#artifact) are work products (PRD, TDD, code, tests) tracked across sessions with verification.

**Family**: artifact
**Commands**: 4
**Priority**: MEDIUM

---

## Commands

### ari artifact register

Register an artifact to the session registry.

**Synopsis**:
```bash
ari artifact register [flags]
```

**Description**:
Registers a work product to the artifact registry, associating it with the current session and enabling tracking across workflow phases.

**Examples**:
```bash
# Register a PRD
ari artifact register --type=prd --path=docs/requirements/PRD-auth.md

# Register with custom metadata
ari artifact register --type=tdd --path=docs/design/TDD-auth.md --phase=design
```

**Related Commands**:
- [`ari artifact query`](#ari-artifact-query) — Find registered artifacts

---

### ari artifact query

Query the artifact registry.

**Synopsis**:
```bash
ari artifact query [flags]
```

**Description**:
Searches the artifact registry with optional filters. Returns artifacts matching criteria.

**Examples**:
```bash
# Query all artifacts
ari artifact query

# Filter by type
ari artifact query --type=prd

# Filter by session
ari artifact query --session=session-20260108-123456

# JSON output
ari artifact query -o json
```

**Related Commands**:
- [`ari artifact list`](#ari-artifact-list) — Summary view

---

### ari artifact list

List artifact counts by dimension.

**Synopsis**:
```bash
ari artifact list [flags]
```

**Description**:
Shows summary of registered artifacts grouped by type, phase, or session.

**Examples**:
```bash
# List artifacts
ari artifact list

# Group by type
ari artifact list --by=type
```

---

### ari artifact rebuild

Rebuild project registry from session registries.

**Synopsis**:
```bash
ari artifact rebuild [flags]
```

**Description**:
Rebuilds the project-level artifact registry by aggregating all session registries. Use after manual session manipulation.

**Examples**:
```bash
# Rebuild registry
ari artifact rebuild

# Dry run
ari artifact rebuild --dry-run
```

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

- [Artifact Glossary Entry](../../reference/GLOSSARY.md#artifact)
- [Artifact Registry](../../reference/GLOSSARY.md#artifact-registry)

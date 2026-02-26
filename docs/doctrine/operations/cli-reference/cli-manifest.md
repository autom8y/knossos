---
last_verified: 2026-02-26
---

# CLI Reference: manifest

> Show, validate, diff, and merge Claude Extension Manifest (CEM) files.

**Family**: manifest
**Commands**: 4
**Priority**: MEDIUM

---

## Commands

### ari manifest show

Display current effective manifest.

**Synopsis**:
```bash
ari manifest show [flags]
```

**Description**:
Shows the merged effective manifest from project and user sources. Displays agent, skill, and hook configurations.

**Examples**:
```bash
# Show manifest
ari manifest show

# YAML output
ari manifest show -o yaml

# JSON for scripting
ari manifest show -o json
```

**Related Commands**:
- [`ari manifest validate`](#ari-manifest-validate) — Validate manifest

---

### ari manifest validate

Validate manifest against schema.

**Synopsis**:
```bash
ari manifest validate [flags]
```

**Description**:
Validates the manifest structure and content against the CEM schema. Reports errors and warnings.

**Examples**:
```bash
# Validate manifest
ari manifest validate

# JSON output
ari manifest validate -o json
```

**Related Commands**:
- [`ari rite validate`](cli-rite.md#ari-rite-validate) — Rite validation

---

### ari manifest diff

Compare two manifests.

**Synopsis**:
```bash
ari manifest diff [flags]
```

**Description**:
Shows differences between two manifest files or between current and a reference manifest.

**Examples**:
```bash
# Diff current vs reference
ari manifest diff

# Diff two files
ari manifest diff manifest-a.yaml manifest-b.yaml
```

---

### ari manifest merge

Three-way merge of manifests.

**Synopsis**:
```bash
ari manifest merge [flags]
```

**Description**:
Performs a three-way merge of manifest files. Useful when resolving conflicts between project and user manifests.

**Examples**:
```bash
# Merge manifests
ari manifest merge base.yaml ours.yaml theirs.yaml
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

- [Manifest Glossary Entry](../../reference/GLOSSARY.md#manifest)
- [Rite Manifests](../../rites/)

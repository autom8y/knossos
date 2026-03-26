---
last_verified: 2026-03-26
---

# CLI Reference: knows

> Inspect the codebase knowledge stored in `.know/`.

`ari knows` is your interface to the local `.know/` knowledge domains. List domains with freshness status, print domain content, validate references, check for staleness in CI, and explore service boundaries.

**Family**: knows
**Commands**: 1 (with mode flags)
**Priority**: MEDIUM

---

## Synopsis

```bash
ari knows [domain] [flags]
```

## Description

Without arguments, lists all domains with freshness status. With a domain name, prints the full content of that domain file to stdout.

## Subcommands

None. `ari knows` uses flags to switch modes.

## Key Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--check` | bool | false | Exit 0 if all fresh, exit 1 if any stale (CI use) |
| `--validate` | bool | false | Validate references in `.know/` files against codebase |
| `--delta` | bool | false | Show change manifests for all (or specified) domains |
| `--semantic-diff` | bool | false | AST-based semantic diff for Go files |
| `--scope-dir` | string | project root | Starting directory for hierarchical `.know/` discovery |
| `--discover` | bool | false | Discover service boundaries with `.know/` candidates |

## Examples

```bash
# List all domains with freshness status
ari knows

# Print full content of architecture.md
ari knows architecture

# Check for staleness (CI/hooks)
ari knows --check

# Validate all references
ari knows --validate

# Validate a single domain
ari knows --validate arch

# Show change manifests for all domains
ari knows --delta

# Show change manifest for one domain
ari knows --delta architecture

# AST-based semantic diff for Go files
ari knows --semantic-diff arch

# Hierarchical view from a service directory
ari knows --scope-dir services/payments/

# Discover service boundaries
ari knows --discover

# JSON output for scripting
ari knows -o json
```

## See Also

- [`ari land synthesize`](cli-land.md) — Cross-session knowledge synthesis
- [`ari registry`](cli-registry.md) — Cross-repo knowledge catalog
- [Architecture Map](../../reference/architecture-map.md) — Codebase structure reference

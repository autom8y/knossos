# ADR: Empty-Domain Repos in Content Pipeline Registry

**Status**: Accepted
**Date**: 2026-03-27
**Author**: Ecosystem Analyst
**Initiative**: content-pipeline-health (Sprint-5)

---

## Context

The org-wide registry (`deploy/registry/domains.yaml`) catalogs all autom8y repos and their
`.know/` domain files for consumption by the Clew content pipeline. Two repos appear with
`domains: []`:

| Repo | `last_synced` | `domains` | `.know/` exists | Knossos integration |
|------|---------------|-----------|-----------------|---------------------|
| `autom8` | 2026-03-25T12:42:15Z | `[]` | No | Yes (.claude/, .knossos/) |
| `autom8y-workflows` | 2026-03-25T12:42:50Z | `[]` | No | No |

The SRE handoff confirmed that `syncRepo()` correctly returns an empty domain list for repos
without `.know/` directories. Their presence in the catalog produces no warnings, no startup
failures, and no content pipeline errors. The question is whether to remove them, generate
knowledge for them, or leave them as-is.

## Investigation

### autom8 (Enterprise Automation Platform)

- **Size**: 2,152 Python files across adapters/, apis/ (30 API integrations), app/, scripts/,
  tests/, terraform/, sql/
- **README**: 907 lines documenting architecture, protocols, and APIs
- **Knossos state**: Full integration -- `.claude/` (agents, commands, skills, rules),
  `.knossos/` (active rite, manifests, procession mena)
- **Assessment**: Substantial codebase with complex domain logic. Knowledge generation via
  `ari know --all` would produce meaningful architecture, conventions, and feature knowledge
  files. This repo is the primary autom8y product and a strong Clew content source.

### autom8y-workflows (Reusable GitHub Actions)

- **Size**: 3 files total -- README.md (1.4KB), satellite-ci-reusable.yml (456 lines)
- **No** `.claude/`, `.knossos/`, or any knossos integration
- **Assessment**: Infrastructure-only repo containing a single reusable CI workflow. Knowledge
  generation would produce trivially thin files (one architecture file describing a YAML
  pipeline). Not a meaningful Clew content source. The README already captures everything
  relevant.

## Decision

**Split approach: Generate knowledge for autom8 (Option B), leave autom8y-workflows as-is (Option C).**

### autom8: Generate knowledge (Option B)

autom8 is a 2,000+ file codebase with 30 API integrations, adapter patterns, circuit breakers,
and terraform infrastructure. It already has full knossos integration. Generating `.know/` files
is a straightforward `ari know --all` invocation that will:

1. Create architecture.md, conventions.md, design-constraints.md, scar-tissue.md, test-coverage.md
2. Populate the registry on next `ari registry sync` with 5+ domain entries
3. Make autom8 domain knowledge available to Clew queries about the automation platform

This is the highest-value action for Clew coverage because autom8 is the primary product repo
and currently contributes zero knowledge to organizational intelligence.

### autom8y-workflows: Leave as-is (Option C)

autom8y-workflows is an infrastructure utility repo with a single 456-line YAML file. Generating
knowledge would produce negligible content -- there is no architecture to document beyond "this
is a reusable GitHub Actions workflow." The repo remains in the catalog harmlessly with
`domains: []`. Removing it (Option A) would be immediately undone by the next `ari registry sync`.

## Consequences

### What changes

- **autom8** gets `.know/` generation queued as a follow-up action (not executed in this sprint).
  Once generated and synced, it will appear in `domains.yaml` with domain entries and become
  available to Clew.

### What stays the same

- **autom8y-workflows** remains in `domains.yaml` with `domains: []`. No action needed.
- **No registry changes** in this sprint. Both repos stay in the catalog.
- **No code changes**. The `syncRepo()` behavior for empty-domain repos is correct.

### Risks

- **autom8 knowledge quality**: The repo may have stale or poorly-structured code that produces
  low-confidence knowledge files. Mitigated by reviewing generated output before syncing to
  the registry.
- **autom8y-workflows re-evaluation**: If the repo grows beyond a single workflow file in the
  future, revisit whether knowledge generation becomes worthwhile. Current threshold: 3+ files
  with distinct architectural patterns.

## Follow-Up Actions

1. Run `ari know --all` in `/Users/tomtenuta/Code/autom8/` (separate sprint, not this one)
2. Run `ari registry sync` after knowledge generation to populate domains.yaml
3. Verify autom8 domains appear in Clew content pipeline

---
domain: feat/org-management
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/org/**/*.go"
  - "./internal/materialize/orgscope/**/*.go"
  - "./internal/config/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Organization-Level Resource Management

## Purpose and Design Rationale

Lets multiple projects share rites, agents, and mena without duplication. Two distinct concerns: resource materialization (orgscope: copies to ~/.claude/ at sync time) and knowledge registry (registry/org: cross-repo domain catalog for Clew). Org resources go to user channel dir (not project dir) for cross-project availability. Active org resolution: KNOSSOS_ORG env var > config file.

## Conceptual Model

**Three tiers:** Config (active-org pointer at XDG config), Data (org resources + metadata at XDG data), Registry (cross-repo catalog). **OrgContext interface:** Name(), DataDir(), RegistryDir(), Repos(). **Scope levels:** all/rite/org/user in sync pipeline. Org scope fires at Phase 1.5 (after rite, before user). **Provenance:** ORG_PROVENANCE_MANIFEST.yaml with ScopeOrg. **Qualified names:** org::repo::domain for cross-repo addressing.

## Implementation Map

CLI: `internal/cmd/org/` (init, set, list, current -- NeedsProject=false). Config: `internal/config/` (ActiveOrg, OrgContext interface, DefaultOrgContext). Paths: `internal/paths/paths.go` (OrgDataDir, OrgRitesDir, OrgAgentsDir, OrgMenaDir). Sync: `internal/materialize/orgscope/sync.go` (SyncOrgScope -- flat files only, no subdirectories). Registry: `internal/registry/org/` (DomainCatalog, SyncRegistry, HandlePushEvent, LoadCatalog/SaveCatalog).

## Boundaries and Failure Modes

Org scope does not affect project-scope resources. Registry sync decoupled from resource sync. Flat file only (subdirectories in org agents/mena silently skipped). No rite-level metadata for org agents. No org configured: status=skipped (non-fatal). Individual file failures: warn and continue. ActiveOrg caches via sync.Once (test must ResetKnossosHome). No webhook receiver route wired in serve.go.

## Knowledge Gaps

1. Collision detection between org and user scope not read
2. Procession org integration not traced
3. ari sync --org flag passing path not confirmed

---
domain: feat/clew-org-registry
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/registry/org/**/*.go"
  - "./deploy/registry/**"
  - "./deploy/content/**"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.82
format_version: "1.0"
---

# Clew Organization Knowledge Address Space

## Purpose and Design Rationale

Cross-repo domain catalog for Clew. When an engineer asks "@clew how does scheduling handle retries?", Clew needs to know which repos have `.know/` files and whether they're fresh. Stores frontmatter metadata only (qualified name, path, timestamps, confidence) -- content fetched separately via pre-baked content directory. Distinct from `internal/registry/` (generic leaf map).

## Conceptual Model

Three-level hierarchy: Org -> Repo -> DomainEntry. Qualified name format: `"org::repo::domain"`. DomainCatalog persisted at `$XDG_DATA_HOME/knossos/registry/{org}/domains.yaml`. Staleness via IsStale() (generated_at + expires_after). Content at `deploy/content/{repo}/.know/`. Webhook-driven incremental sync (HandlePushEvent) for push-triggered updates.

## Implementation Map

`internal/registry/org/registry.go` (types, staleness), `persist.go` (YAML I/O), `sync.go` (SyncRegistry, per-repo sync, GitHub API), `webhook.go` (push event handling). Deploy: `deploy/registry/domains.yaml` (8 repos), `deploy/content/` (pre-baked .know/ files). CLI: `ari registry sync/list/status`. Wired in serve.go as `knowledgeCatalogAdapter` and `knowledgeContentAdapter`.

## Boundaries and Failure Modes

Catalog vs content separation (address layer vs content layer). nil catalog at serve startup: fail-open (no provenance links). Per-repo GitHub failures non-fatal. Missing content: domain skipped in knowledge index. No webhook receiver route wired in serve.go (implemented but not exposed).

## Knowledge Gaps

1. deploy/content/ directory contents appear empty locally
2. GitHub webhook route not wired
3. Multi-org support unclear (single org per deployment assumed)

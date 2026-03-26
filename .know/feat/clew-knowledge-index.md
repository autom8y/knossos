---
domain: feat/clew-knowledge-index
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/search/knowledge/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# Clew Semantic Knowledge Index

## Purpose and Design Rationale

Retrieval layer for the Clew Slack bot. Wraps four orthogonal sub-systems: BM25 (term matching), embedding (semantic similarity via character-frequency hash -- not real embeddings, Sprint 7 placeholder), summary (Haiku-generated prose), and graph (deterministic metadata-derived entity edges). Pre-baked in container image (BC-11). Restart required for cache coherence (BC-10). Sub-packages must NOT import parent (RR-007).

## Conceptual Model

**Coordinator wrapping 4 stores:** BM25Searcher (interface), embedding.Store (256-dim hash vectors), summary.Store (Haiku-generated), graph.Graph (same_type/same_repo/scope_overlap edges). Universal key: `"org::repo::domain"`. Source-hash-based cache invalidation. Persistence via `persistedIndex` JSON (splits off BM25 which is interface-only). Background build with 10-minute timeout; server starts with BM25 fallback.

**Build pipeline (5 steps):** Load persisted -> check needsRegeneration per domain -> errgroup(limit=10) parallel: generate summaries + compute embeddings -> build graph -> save -> validate.

## Implementation Map

`internal/search/knowledge/types.go` (interfaces), `index.go` (KnowledgeIndex struct + methods), `builder.go` (5-step Build pipeline), `persist.go` (JSON persistence). Sub-packages: `embedding/store.go` (256-dim TextToVector + cosine similarity), `graph/graph.go` (3 edge types, deterministic), `summary/store.go` (Haiku summarization).

## Boundaries and Failure Modes

Not a vector database (brute-force O(n) scan). TextToVector is character-frequency hash (not semantic). Not in live query path (feeds health checks; queries use BM25+RRF from internal/search). Background build result not wired to live pipeline (critical gap). LLM unavailability degrades summaries silently. FreshnessScore always zero in Tier 1. Persisted index version mismatch returns error.

## Knowledge Gaps

1. resolveKnowledgeContentStore() not examined
2. Background build wiring gap root cause not traced
3. Sub-package test coverage not read

---
domain: feat/bm25-search-engine
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/search/**/*.go"
  - "./internal/cmd/ask/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.92
format_version: "1.0"
---

# BM25 + RRF Semantic Search Engine (ari ask)

## Purpose and Design Rationale

Natural language query interface over knossos CLI surface and cross-repo knowledge graph. BM25 chosen for determinism, zero-latency build, interpretability on small corpus (<50 domains). RRF merges heterogeneous score spaces. Sprint-2 parameter sweep: k1=1.2, b=0.25, RRF-k=40. Fail-open: missing catalog -> structural-only results. Session-aware scoring for phase/activity/complexity modifiers.

## Conceptual Model

**Two channels:** Structural (CLI surface, 4-tier scoring: exact 1000, prefix 500, keyword, Levenshtein) and BM25 (cross-repo .know/ content, document + section level, field boosting). **RRF fusion:** domain-name boosting, section dedup (top-2 per parent), multi-channel merge. **Freshness:** display-only annotation (D-5), NOT ranking signal. Domain-specific half-lives. **Synonym expansion:** static + orchestrator-derived, 60% weight, 6-expansion cap. **Session scoring:** phase boost (+150), activity boost (+75), complexity penalty (-100).

## Implementation Map

`internal/cmd/ask/ask.go` (CLI entry), `internal/search/index.go` (Build + Search), `internal/search/score.go` (4-tier scorer), `internal/search/session.go` (session signals), `internal/search/synonyms.go` (composite sources), `internal/search/collectors.go` (8 Collect* functions), `internal/search/bm25/` (index, scorer, build, section, decay, params), `internal/search/fusion/rrf.go` (RRFMerge), `internal/search/content/store.go` (PreBaked + Local stores).

## Boundaries and Failure Modes

No catalog/org: BM25 nil, structural-only results. Per-domain content load failure: domain skipped. Session missing/malformed: no session signals, no effect on results. BM25 returns nothing: short-circuit to structural. Tokenizer strips `:` (breaks qualified names, compensated by field boosting). LookupContent is O(n) scan. Section splitting only at H2 boundaries.

## Knowledge Gaps

1. KnowledgeIndex builder relationship not traced
2. concept.AllConcepts() registry not read
3. session.FindActiveSession() failure modes not confirmed

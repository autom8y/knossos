---
domain: feat/clew-rag-triage
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/triage/**/*.go"
  - "./internal/reason/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.90
format_version: "1.0"
---

# Clew RAG Triage Pipeline

## Purpose and Design Rationale

Pre-retrieval intelligence layer for Clew. Narrows the knowledge corpus (hundreds of `.know/` domains) to 3-5 most relevant domains before the reasoning pipeline incurs Claude API cost. Solves multi-turn query resolution and domain pre-selection. Sits between `internal/slack/` and `internal/reason/` with an anti-cycle boundary (reason/ does NOT import triage/; mirrored structs with slack/ bridging). Cost-ordered 4-stage funnel: Stage 0 (optional LLM, follow-ups only), Stage 1 (zero-cost metadata), Stage 2 (embedding/BM25), Stage 3 (single Haiku call, max 800 tokens). Fail-open at every stage.

## Conceptual Model

**Two-pipeline architecture:** Pipeline 1 (Triage) produces refined query + ranked candidates. Pipeline 2 (Reasoning) uses QueryStream/QueryWithTriage for token-budgeted context assembly + Claude API call.

**Four stages:** Stage 0 (2s timeout, query refinement for follow-ups), Stage 1 (substring+signal matching, staleness gate at freshness<0.1), Stage 2 (currently BM25 only -- StubEmbeddingModel triggers fallback; real embeddings Sprint 7), Stage 3 (Haiku scoring, max 20 candidates, partial JSON recovery for truncated responses).

**Three reasoning paths:** Query() (standalone BM25), QueryWithTriage() (triage-aware sync), QueryStream() (triage-aware streaming with onChunk callback).

**Trust scoring:** geometric mean of freshness x retrieval x coverage -> HIGH (>=0.7) / MEDIUM (0.4-0.7) / LOW (<0.4). LOW short-circuits before Claude with gap admission. Context assembly uses greedy bin-packing (0.50 relevance + 0.30 freshness + 0.20 diversity). Source budget: 8,000 tokens.

## Implementation Map

`internal/triage/orchestrator.go` (4-stage pipeline), `types.go` (core types), `prompts.go` (Stage 0/3 prompts). `internal/reason/pipeline.go` (3 entry points + anti-cycle boundary types + contentLookup). `internal/reason/intent/classifier.go` (keyword heuristics). `internal/reason/context/assembler.go` (token-budgeted packing). `internal/reason/response/generator.go` + `stream.go` (Claude API + SSE batcher).

## Boundaries and Failure Modes

Stage 0 timeout -> original query used. Stage 1 no candidates -> nil (v1 BM25 fallback). Stage 2 embedding failure -> BM25 fallback (BC-06: required path). Stage 3 failure -> Stage 2 scores. LOW confidence -> no Claude call. StubEmbeddingModel means BM25-only until Sprint 7. DomainCoverage fixed at 1.0 for triage path. Streaming path does not run ValidateCitations inline (asymmetry with sync path).

## Knowledge Gaps

1. slack/handler.go triage-to-reason conversion not read
2. search.SearchIndex.LookupContent backing store not read
3. trust.Scorer internals not read directly
4. streaming.ExtractCitations regex not read

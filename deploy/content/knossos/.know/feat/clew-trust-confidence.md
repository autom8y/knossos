---
domain: feat/clew-trust-confidence
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/trust/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.91
format_version: "1.0"
---

# Clew Multi-Axis Confidence Scoring

## Purpose and Design Rationale

Answer reliability system. Three independent signals (freshness, retrieval quality, coverage) compose into a composite score dictating HIGH (answer directly), MEDIUM (answer with caveats), LOW (refuse, emit GapAdmission). Weighted geometric mean (primary) with arithmetic mean fallback when any input is zero (WS-2.4 fix). Freshness weighs most (0.45). trust/ must not import search/ or cmd/*.

## Conceptual Model

**Three axes:** Freshness (exponential decay with domain-specific half-lives: release 3d, architecture 14d, literature 90d), Retrieval quality (normalized BM25/RRF score), Coverage (fraction of needed domains found). **Composition:** geometric mean (log-space) or arithmetic fallback. **Tiers:** HIGH >=0.70, MEDIUM 0.40-0.70, LOW <0.40. **Provenance chain:** ordered list of cited .know/ files with per-source freshness. **GapAdmission:** missing domains, stale domains, suggestions.

## Implementation Map

`internal/trust/confidence.go` (Scorer, scoring algorithm), `config.go` (TrustConfig, thresholds, weights), `decay.go` (DecayConfig, domain half-lives, ClassifyDomain), `provenance.go` (ProvenanceChain, WeightedMeanFreshness), `gap.go` (GapAdmission, generateSuggestions).

## Boundaries and Failure Modes

Unparseable timestamps -> 0.0 (fail-safe). Zero-input collapse prevented by arithmetic fallback. Empty chain -> 0.0 freshness. Config validation: LowThreshold < HighThreshold enforced. DomainCoverage=1.0 in triage path (coverage axis disabled). Half-life updates reset historical calibration without migration.

## Knowledge Gaps

1. config_test.go and decay_test.go not read
2. reasoncontext.WeightedMeanFreshness vs ProvenanceChain.WeightedMeanFreshness differences unknown
3. MEDIUM tier caveats injection prompt not read

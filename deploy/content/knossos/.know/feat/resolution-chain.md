---
domain: feat/resolution-chain
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/resolution/**/*.go"
  - "./internal/materialize/source/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Multi-Tier Rite Resolution Chain

## Purpose and Design Rationale

The chain solves a single core problem: locate named resources (rites, processions, contexts) from multiple sources with higher-priority definitions shadowing lower-priority ones. Extracted as a standalone primitive with **zero internal imports** (TENSION-005) to avoid import cycles. Sits at the Infrastructure/Leaf layer.

## Conceptual Model

Ordered list of `Tier` structs (Label, Dir, FS). Tiers with empty Dir and nil FS are filtered out. Two operations: `Resolve` (top-down early-exit for single lookup) and `ResolveAll` (bottom-up sweep with shadowing for catalog listing).

**Standard chains:** RiteChain (project > user > org > platform > embedded), ProcessionChain (same order), ContextChain (user > project > org > platform -- user outranks project for context files).

Validation is a callback -- callers inject what "valid" means. `ResolvedItem` carries source provenance (tier label, path, optional fs.FS).

## Implementation Map

`internal/resolution/chain.go` (core engine), `builders.go` (3 chain factories), `chain_test.go` (13 tests). Consumed by `internal/materialize/source/resolver.go` (rite resolution with caching), `internal/rite/discovery.go` (rite listing), `internal/materialize/procession/resolver.go` (procession templates).

`SourceResolver` caches by `riteName + orgRitesDir` key. SCAR-023 regression test covers self-hosting template path fallback.

## Boundaries and Failure Modes

- Not found: structured error with `checked_paths` detail field
- Unreadable tier: silently skipped (empty slice returned)
- Empty tiers: filtered by `NewChain`
- No embedded FS in `rite.Discovery` (only sync/init use embedded)
- Chain is stateless; caching is `SourceResolver`'s responsibility

## Knowledge Gaps

1. `ContextChain` callers not identified
2. Procession resolver name-collision dedup not tested
3. `ClearCache()` call sites not identified

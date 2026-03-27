---
domain: feat/know-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/know/**/*.go"
  - "./internal/cmd/knows/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.91
format_version: "1.0"
---

# Codebase Knowledge Domain System (.know/)

## Purpose and Design Rationale

Pre-computes and caches durable codebase knowledge. Three pillars: freshness over regeneration (source_scope + source_hash + expires_after with dual staleness signals), incremental regeneration (update_mode + incremental_cycle + ChangeManifest + SemanticDiff), monorepo awareness (FindKnowDirs upward walk, nearest-wins merge, path-prefixed domain names). Land integration extends staleness to experiential knowledge (.sos/land/).

## Conceptual Model

**Domain:** single .know/{domain}.md file. Three namespaces: top-level, feat/, release/. **QualifiedDomainName:** org::repo::domain (exactly two ::). **DomainStatus:** 5 orthogonal staleness signals (time_expired, code_changed, force_full, dependency_stale, land_changed). **ChangeManifest:** categorized file lists, commit log, delta ratio (threshold 0.5 or 5000 lines for full mode). **SemanticDiff:** AST-level Go declaration changes (NEW/DELETED/MODIFIED/SIGNATURE_CHANGED). **ValidationReport:** reference integrity checking (file paths, function names, commit hashes).

## Implementation Map

`internal/know/` (6 files): know.go (ReadMeta, FindKnowDirs, buildDomainStatus, scopedStaleness, matchScope, dependency DAG, land staleness), manifest.go (ComputeChangeManifest, FilterChangeManifest, RecommendedMode), validate.go (ValidateDomain, extractRefs, 4 verification types), qualified.go (Parse, QualifiedDomainName), astdiff.go (ComputeFileDiff via go/ast), discover.go (DiscoverServiceBoundaries). CLI: `internal/cmd/knows/knows.go` (list, --check, --validate, --delta, --semantic-diff, --discover, --scope-dir; caches unscoped manifests).

## Boundaries and Failure Modes

Git dependency throughout (30s timeouts, graceful degradation). Custom matchScope handles ** by splitting (doesn't implement full double-star glob). Freshness false-negative risk if working directory != repo root. Dependency graph trust: qualified names may not match plain domain lookup. Validation regex matches PascalCase tokens (false positives on renamed types). Incremental cycle exhaustion: ForceFull=true but Fresh not set to false (informational flag). Land hash graceful degradation on deleted files.

## Knowledge Gaps

1. knows_test.go not read
2. KnowledgeIndex integration with QualifiedDomainName not fully traced
3. validate_test.go and astdiff_test.go not read

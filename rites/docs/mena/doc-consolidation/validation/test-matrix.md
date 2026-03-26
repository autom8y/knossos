---
description: "Test Matrix for Consolidation Validation companion for validation skill."
---

# Test Matrix for Consolidation Validation

> Define test scenarios across complexity levels to validate the workflow handles simple, medium, complex, and edge cases correctly.

## Complexity Tiers

| Tier | Documents | Overlaps | Conflicts | Edge Cases | Purpose |
|------|-----------|----------|-----------|------------|---------|
| Simple | 3-5 | None | None | None | Baseline functionality |
| Medium | 10-15 | Some | Some | Few | Typical real-world scenario |
| Complex | 20+ | Many | Multiple | Many | Stress test and robustness |
| Edge | Variable | N/A | N/A | All | Boundary conditions |

---

## Simple Tier

| ID | Scenario | Key Assertions |
|----|----------|----------------|
| S1 | 3-file clean consolidation | MG-1–8 pass; 0 conflicts; SG-3: 1000→~600 tokens |
| S2 | 5-file linear dependency chain | Extraction order respected; all cross-refs updated |
| S3 | Disjoint topics (2 topics, no overlap) | Separate extractions; no cross-topic contamination |

---

## Medium Tier

| ID | Scenario | Key Assertions |
|----|----------|----------------|
| M1 | 10-file overlapping content, clear authority | 40% token reduction; overlap detected; primary wins |
| M2 | Multiple conflicts (blocking, significant, minor) | EG-5: blocking conflict C1 resolved; all 3 conflicts documented |
| M3 | 15-file cross-reference web | LI-1: all links resolve post-consolidation; CR-1: deps satisfied |

**M2 conflict matrix:**

| Conflict | Sources | Severity |
|----------|---------|----------|
| C1: Timeout default (30s vs 60s vs 45) | TDD-arch, TDD-config, config-module | Blocking |
| C2: Feature availability | user-guide, api-docs | Significant |
| C3: Terminology drift | Multiple | Minor |

---

## Complex Tier

| ID | Scenario | Key Assertions |
|----|----------|----------------|
| C1 | 25 docs, 15K tokens, 8 conflicts | < 2K tokens/extraction; 50+ links resolve; no timeout |
| C2 | Deep 6-topic dependency chain (A→B→C→D→E→F) | MG-8: no cycles; extraction order enforced |
| C3 | Multi-author conflicts, no clear recency | Authority hierarchy resolution; terminology normalized |

---

## Edge Cases

| ID | Scenario | Expected Behavior |
|----|----------|-------------------|
| E1 | Empty / whitespace-only files | Flagged in exclusions; not counted as orphans |
| E2 | True circular references (a→b→c→a) | MG-8 fails if modeled as topic deps |
| E3 | Binary files in scope (png, json) | Excluded with reason "binary"; not orphans |
| E4 | Single file > 10K tokens | MANIFEST warns; topic split recommended |
| E5 | Unicode filenames and content | Paths handled; hashes computed correctly |
| E6 | Duplicate filenames in different dirs | Treated as separate files; disambiguated by full path |
| E7 | File modified during consolidation | SC-1/EG-6: stale hash detected; re-extraction prompted |
| E8 | No primary source exists for topic | MG-5 fails; human input required |
| E9 | File excluded in one topic, included in another | MANIFEST flags in ambiguous[] |
| E10 | Self-referencing anchor link | Resolves; not flagged as circular |

---

## Regression Test Suite

### Minimum Set (run on every workflow change)

| Scenario | Tests | Priority |
|----------|-------|----------|
| S1 | Basic functionality | P0 |
| S3 | Multi-topic handling | P0 |
| M2 | Conflict resolution | P0 |
| E1 | Empty file handling | P1 |
| E3 | Binary exclusion | P1 |
| E7 | Staleness detection | P1 |

### Full Set (run before major releases)
All S1–S3, M1–M3, C1–C3, E1–E10.

### Performance Benchmarks

| Scenario | Max Duration | Target |
|----------|--------------|--------|
| Simple | 30s | 10s |
| Medium | 2min | 45s |
| Complex | 10min | 3min |

---

## Success Criteria

| Tier | Pass Criteria |
|------|---------------|
| Simple | All gates pass, > 30% token reduction, 0 info loss |
| Medium | All gates pass, conflicts documented, > 25% token reduction |
| Complex | All gates pass, no timeouts, scalable performance |
| Edge | Graceful handling, appropriate errors, no crashes |

---

*Part of the [Documentation Consolidation Workflow](../INDEX.lego.md)*

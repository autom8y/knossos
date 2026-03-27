# Content Pipeline Health -- 10x-dev Handoff

## Sprint-1: Mutex Serialization Fix for Summary Generation

**Date**: 2026-03-27
**Scope**: `internal/search/knowledge/builder.go`, `internal/search/knowledge/summary/store.go`

---

## What Was Changed

### Root Cause

In `builder.go:179-183`, `mu.Lock()` wrapped the entire `summaryStore.Generate()` call, which includes an LLM API call (`client.Complete()` at `store.go:136`). With `g.SetLimit(10)` concurrency, only 1 goroutine could call Haiku at a time. The other 9 blocked on `mu.Lock()` with their 30s context timers already running, causing cascade timeouts across batches.

### Fix Applied (Option B: Split Store API)

**store.go** -- Added `GenerateSummary()` method that performs the LLM call and returns `*DomainSummary` without writing to the store. The existing `Generate()` was refactored to call `GenerateSummary()` + `Set()`, preserving backward compatibility for the `Reindex()` caller in `index.go`.

**builder.go** -- Replaced `Generate()` (under mutex) with:
1. `GenerateSummary()` -- runs unlocked, performs LLM I/O
2. `Set()` -- runs under `mu.Lock()`, writes result to store

### Before / After Lock Scope

```
BEFORE (builder.go:179-183):
  domainCtx, cancel := context.WithTimeout(gCtx, 30*time.Second)
  mu.Lock()
  _, genErr := summaryStore.Generate(...)   // LLM call + store write -- SERIALIZED
  mu.Unlock()
  cancel()

AFTER:
  domainCtx, cancel := context.WithTimeout(gCtx, 30*time.Second)
  ds, genErr := summaryStore.GenerateSummary(...)  // LLM call only -- PARALLEL
  cancel()
  if genErr == nil {
      mu.Lock()
      summaryStore.Set(ds)                         // store write only -- FAST
      mu.Unlock()
  }
```

## Test Results

### Existing Tests (unchanged, all pass)
```
ok  github.com/autom8y/knossos/internal/search/knowledge           1.011s
ok  github.com/autom8y/knossos/internal/search/knowledge/embedding  (cached)
ok  github.com/autom8y/knossos/internal/search/knowledge/graph      (cached)
ok  github.com/autom8y/knossos/internal/search/knowledge/summary    0.284s
```

### New Concurrency Tests

**TestBuild_ConcurrentSummaryGeneration** (10 domains, 200ms LLM latency):
- Wall clock: ~201ms (sequential would be 2s) -- 10x speedup
- Max concurrent LLM calls: 10 (full parallelism)

**TestBuild_ConcurrentSummaryGeneration_NoTimeoutCascade** (20 domains, 100ms LLM latency):
- Wall clock: ~201ms (sequential would be 2s) -- 10x speedup
- Max concurrent LLM calls: 10 (2 batches of 10, each ~100ms)
- All 20 summaries generated successfully (no timeout cascade)

**TestStore_GenerateSummary** (unit test for new method):
- Returns summary without writing to store (Count() == 0 after call)
- Error propagation from LLM client
- Nil client returns error

### Full Suite
All packages pass. Pre-existing failure in `internal/reason` (Claude API dependency, unrelated).

## Knowledge Index Generation Command

To generate `knowledge-index.json` locally for the SRE WS-2 pre-bake pipeline:

```bash
CGO_ENABLED=0 go build ./cmd/ari && cp ./ari $(which ari)
CLEW_CONTENT_DIR=/path/to/deploy/content ari serve --build-index-only
```

Or via the `Build()` function directly with a populated `BuildConfig` containing a catalog, content store, LLM client, and `PersistedPath` pointing to the desired output location.

## Store API Observations

The `summary.Store` type is intentionally NOT concurrency-safe -- the caller owns synchronization. This is a sound design for the current usage pattern where only `builder.go` drives concurrent access.

The new `GenerateSummary()` / `Set()` split makes the I/O vs mutation boundary explicit at the API level. If future callers need thread-safe access to the store, the clean separation makes it straightforward to add internal locking to `Set()` / `GetSummary()` without touching the LLM call path.

The `Reindex()` method in `index.go` still uses `Generate()` (which calls `GenerateSummary()` + `Set()` internally). This is correct because `Reindex()` operates on a single domain with no concurrent store access. If `Reindex()` ever becomes concurrent, it should follow the same pattern as the builder fix.

## Files Modified

| File | Change |
|------|--------|
| `internal/search/knowledge/summary/store.go` | Added `GenerateSummary()` method; refactored `Generate()` to delegate |
| `internal/search/knowledge/builder.go` | Narrowed mutex scope: LLM call outside lock, `Set()` under lock |
| `internal/search/knowledge/builder_concurrency_test.go` | New: 2 concurrency tests with latency-aware mock |
| `internal/search/knowledge/summary/store_test.go` | New: `TestStore_GenerateSummary` unit test |

---
name: thermia-ref-patterns
description: "Thermia cache pattern catalog, consistency models, and failure modes. Use when: selecting a cache pattern for a layer, choosing a consistency model, designing failure mode behavior, understanding CAP trade-offs. Triggers: cache-aside, read-through, write-through, write-behind, refresh-ahead, consistency model, eventual consistency, CAP theorem, failure mode, stampede, cold start."
---

# Thermia: Cache Patterns, Consistency, and Failure Modes

## Cache Patterns

| Pattern | Best Fit | Trade-off |
|---------|----------|-----------|
| **Cache-aside** | Read-heavy, miss-tolerant; application controls load | On miss: direct origin fetch, risk of stampede on cold start |
| **Read-through** | Read-heavy, miss-intolerant; cache manages loading | Cache is in the critical read path; miss latency equals origin latency |
| **Write-through** | Write-heavy, consistency-critical; synchronous to origin + cache | Write latency increases (both origin and cache must ack) |
| **Write-behind** | Write-heavy, latency-sensitive; async to origin | Risk of data loss if cache fails before async write completes |
| **Refresh-ahead** | Predictable access patterns, latency-critical; proactive refresh | Risk of refreshing data that is never accessed; wasted origin load |

### Pattern Selection Heuristic

1. Read-heavy, miss-tolerant → cache-aside
2. Read-heavy, miss-intolerant → read-through or refresh-ahead
3. Write-heavy, consistency-critical → write-through
4. Write-heavy, latency-sensitive → write-behind (with data loss acknowledgment)
5. Predictable access, latency-critical → refresh-ahead

## Consistency Models

| Model | Guarantee | CAP Position | When to Use |
|-------|-----------|-------------|-------------|
| **Linearizability** | Every read sees the latest write; appears as single copy | CP | Auth tokens, financial balances, security decisions |
| **Sequential** | All nodes see writes in same order; may lag real-time | CP | Session stores, user preference propagation |
| **Causal** | Causally related operations ordered; concurrent ops may differ | AP-leaning | Social graphs, comment threads |
| **Eventual** | All nodes converge given no new writes; no ordering guarantee | AP | Product catalogs, public content, configuration |

### Staleness Budget

The staleness budget from thermal-assessment.md drives consistency model selection:
- Zero tolerance → linearizability or sequential
- Seconds → sequential or causal
- Minutes to hours → eventual

## Failure Mode Design

Every cache layer requires explicit failure behavior for three scenarios.

### Scenario Matrix

| Scenario | Fail-Open Options | Fail-Closed Option | Decision Driver |
|----------|------------------|---------------------|-----------------|
| Cache unavailable | Serve stale / Bypass to origin | Fail the request | Can users tolerate stale or absence? |
| Origin unavailable | Serve stale (extend TTL) / Queue writes | Fail the request | Is stale data better than no data? |
| Network partition | Continue serving cached data (AP) | Refuse reads (CP) | CAP position of this layer |

### Fail-Open vs Fail-Closed

- **Fail-open (serve stale)**: Better availability. Risk: stale data served. Requires staleness visibility metric.
- **Fail-open (bypass)**: Correct data, but origin absorbs full load. Risk: origin overload cascade.
- **Fail-closed**: Data correctness guaranteed. Risk: user-visible failure. Only appropriate when staleness is unacceptable.

## Failure Modes to Prevent

### Cache Stampede (Thundering Herd)

When a high-traffic key expires simultaneously for many concurrent readers, they all hit the origin at once.

**Protection mechanisms** (capacity-engineer selects):
- **XFetch** (Vattani et al., VLDB 2015): Probabilistic early refresh. No coordination. Optimal for independent keys.
- **Lease tokens** (Nishtala et al., NSDI 2013): Prevents stale sets + thundering herd. Requires coordination.
- **Background refresh**: For predictable access patterns. Higher infrastructure cost.
- **Locking**: Simple but serializes reads. Only for very low-contention scenarios.

### Cold Start

On first deployment or after cache flush, all requests miss and hit origin simultaneously.

**Mitigation patterns**:
- Warm the cache from a snapshot before routing traffic
- Rate-limit origin requests during warm-up
- Implement request coalescing at the cache layer

### Inconsistency Window

During the time between a write to origin and cache invalidation, stale data is served.

**Mitigation**:
- TTL-based invalidation (bounded staleness)
- Event-driven invalidation (pub/sub on write events)
- Write-through pattern (eliminates window entirely at write latency cost)

### Invisible Invalidation Chain

Cache A depends on Cache B depends on Cache C. Invalidating C does not automatically invalidate A or B.

**Mitigation**:
- Document the dependency graph explicitly in architecture
- Prefer event-driven invalidation with cascading subscribers
- Avoid deep cache chains (>2 levels requires explicit cascade design)

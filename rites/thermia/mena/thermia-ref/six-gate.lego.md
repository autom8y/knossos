---
name: thermia-ref-six-gate
description: "Thermia 6-gate decision framework for cache candidate evaluation. Use when: assessing whether a hot path should be cached, documenting gate pass/fail with reasoning, evaluating alternatives before recommending cache layers. Triggers: 6-gate, cache decision, should I cache, frequency gate, staleness tolerance, safety gate, alternatives evaluation."
---

# Thermia: 6-Gate Decision Framework

Every cache candidate must pass all six gates. Document pass/fail with reasoning for each.

## The Gates

| Gate | Question | Fail Signal |
|------|----------|-------------|
| **Frequency** | Is this data accessed frequently enough to justify caching? | <10 req/min on a single path rarely justifies dedicated caching |
| **Computation Cost** | Is the origin fetch expensive (latency, CPU, money)? | Sub-millisecond origin fetches gain little from caching |
| **Staleness Tolerance** | Can the consumer tolerate stale data? For how long? | Zero tolerance (financial transactions, auth decisions) means caching is risky |
| **UX Impact** | What is the user-visible impact of cache miss vs hit? | If users cannot perceive the difference, caching adds complexity for no gain |
| **Safety** | Does the data contain PII, multi-tenant data, or security-sensitive material? | PII requires encryption at rest, tenant isolation, and audit trails |
| **Scalability** | Will the access pattern grow? By how much? | Exponential cardinality growth can exhaust cache memory; bounded key spaces are safer |

## Gate Failure Rules

- Gates 1, 2, 4, 6: Fail = **OPTIMIZE-INSTEAD** or **DEFER** verdict
- Gate 3 (Staleness): Fail = requires explicit **risk acknowledgment**, not automatic rejection. Document the risk and surface to user.
- Gate 5 (Safety): Fail = requires explicit **risk acknowledgment**. Document PII/security implications and surface to user.

## Verdicts

| Verdict | Meaning |
|---------|---------|
| **CACHE** | All gates pass. Ready for architecture design. |
| **OPTIMIZE-INSTEAD** | Non-cache alternative solves the problem. Document the recommended alternative. |
| **DEFER** | Insufficient data to decide. Document what data is needed. |

## Alternatives to Exhaust Before Recommending Cache

Before applying the 6-gate to any candidate, assess whether these solve the problem:

- **Query optimization**: Missing index, query rewrite, EXPLAIN-driven tuning
- **Materialized views**: Pre-materialized computed aggregates
- **Connection pooling**: Latency from connection overhead, not data access
- **CDN / Edge caching**: Static or semi-static content for end users
- **Denormalization**: Restructuring data model to eliminate expensive joins
- **Read replicas**: Offloading read traffic without a cache layer

If an alternative solves the problem, recommend it with OPTIMIZE-INSTEAD verdict.

## 6-Gate Summary Table (for thermal-assessment.md)

```
| Candidate | Freq | Cost | Stale | UX | Safety | Scale | Verdict |
|-----------|------|------|-------|-----|--------|-------|---------|
| {name}    | P/F  | P/F  | P/F   | P/F | P/F   | P/F   | CACHE / OPTIMIZE-INSTEAD / DEFER |
```

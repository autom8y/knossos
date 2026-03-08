# Thermia Rite

> Cache architecture consultation. From "should we cache this?" to a validated design with capacity plan and observability spec.

**Version**: 1.0.0 | **Domain**: Cache architecture | **Command**: `/thermia`

## When to Use This Rite

**Triggers**:
- Cache architecture decisions ("should I add Redis?", "how do I design the caching layer?")
- Cache performance problems ("our database is getting hammered", "cache stampede", "cache miss rate too high")
- Cache capacity questions ("how big should our cache be?")
- Cache observability gaps ("we don't know when our cache is unhealthy")
- Post-mortems involving cache ("our cache took down the database")
- Cache pattern selection ("cache-aside vs write-through", "what eviction policy?")
- Consistency and invalidation ("stale data from cache", "cache invalidation is broken")

**Not for**: Cache implementation (use `/10x`), active SRE incident response (use `/clinic` or `/sre`), code review of cache code (use `/review`), performance benchmarking (use `/10x`).

## Quick Start

```bash
/thermia
```

Potnia opens with heat-mapper to assess whether caching is actually warranted. Describe your system, access patterns, and what problem you're trying to solve.

**Already certain you need caching?** Say so explicitly. Potnia will calibrate accordingly, but heat-mapper will still surface assumptions worth examining.

## Agents

| Agent | Model | Phase | Produces |
|-------|-------|-------|----------|
| **potnia** | opus | Orchestrator | Phase gates, complexity gating, consultative flow |
| **heat-mapper** | sonnet | Assessment | thermal-assessment.md |
| **systems-thermodynamicist** | sonnet | Architecture | cache-architecture.md |
| **capacity-engineer** | sonnet | Specification | capacity-specification.md |
| **thermal-monitor** | sonnet | Validation | observability-plan.md |

## Workflow

```
assessment -> architecture -> specification -> validation
```

QUICK complexity runs assessment + validation only (architecture and specification skipped).
STANDARD and DEEP run all four phases.

### Back-Routes

```
architecture ──assessment_gap──> assessment         (max 1 iteration)
validation ──design_inconsistency──> specification  (max 1 iteration)
```

Back-routes trigger when an agent cannot proceed without information from an earlier phase. Potnia manages these; user confirmation is not required unless the max iteration limit is hit.

## Complexity Levels

| Level | Scope | Phases |
|-------|-------|--------|
| **QUICK** | Single caching question, yes/no triage | assessment, validation |
| **STANDARD** | New cache design or existing cache review | assessment, architecture, specification, validation |
| **DEEP** | Post-mortem, production crisis, or full system redesign | assessment, architecture, specification, validation (extended depth) |

Potnia determines complexity from your description. QUICK is for "should we cache this one thing?" STANDARD is for new or revised cache designs. DEEP is for post-mortems and crisis situations where the full analytical framework applies at maximum rigor.

## Artifact Chain

All artifacts are written to `.sos/wip/thermia/`:

```
thermal-assessment.md        <- heat-mapper (assessment)
cache-architecture.md        <- systems-thermodynamicist (architecture, STANDARD+)
capacity-specification.md    <- capacity-engineer (specification, STANDARD+)
observability-plan.md        <- thermal-monitor (validation)
```

Each artifact is immutable once produced. Downstream agents read but do not modify upstream artifacts.

## Commands

| Command | Purpose |
|---------|---------|
| `/thermia` | Switch to thermia rite |
| `/task "cache problem"` | Full consultation lifecycle |

## The 6-Gate Framework

The heat-mapper evaluates every caching decision against six gates before recommending a cache layer:

1. **Frequency** — Is this data accessed often enough to justify a cache?
2. **Computation Cost** — Is the origin slow or expensive enough to warrant caching the result?
3. **Staleness Tolerance** — Can the caller tolerate cached data that may be stale?
4. **UX Impact** — Will users notice the latency difference?
5. **Scalability** — Does the access pattern create load that caching would relieve?
6. **Safety** — Is the data safe to cache (no PII, access control, or audit implications)?

A cache layer is only recommended if all six gates pass. Gates that fail produce an "exhaust alternatives first" assessment: query optimization, CDN, read replicas, denormalization, or rate limiting may solve the problem without adding a cache layer.

## Cross-Rite Handoffs (Outbound)

Thermia produces architectural and observability artifacts, not implementation code. When implementation is needed, thermal-monitor's observability-plan includes cross-rite routing recommendations.

| Target | When |
|--------|------|
| `/10x` | Cache design ready for implementation |
| `/clinic` | Production cache incident requiring structured root cause analysis |
| `/sre` | Observability gaps require operational engineering |
| `/arch` | Cache design raises broader system architecture questions |
| `/hygiene` | Cache code quality issues identified during review |
| `/debt` | Cache architecture reveals systemic technical debt patterns |

## Best For

- Teams that suspect they need caching but haven't justified it rigorously
- Existing cache layers with performance or correctness problems
- Post-mortems where cache behavior contributed to an incident
- Architecture reviews that include caching decisions
- Any situation requiring a capacity plan, eviction policy selection, or observability design for a cache

## Not For

- Writing cache client code (use `/10x`)
- Responding to an active cache-related incident (use `/clinic` for structured investigation or `/sre` for operational response)
- Code review of existing cache implementation (use `/review`)
- Performance benchmarking cache configurations (use `/10x` with benchmarking task)

## Related Rites

- `/10x` — Primary implementation target after thermia delivers cache design
- `/clinic` — Structured production debugging when cache behavior is the suspected cause
- `/sre` — Operational engineering for cache observability and alerting
- `/arch` — Architecture analysis when cache design has system-wide implications
- `/debt` — Technical debt management when cache design reveals systemic patterns

## Design Notes

The thermal management metaphor is intentional. Caches exist to manage thermal load: database access patterns create heat (load), caches provide cooling (request absorption). The agents map to thermal roles — the heat-mapper identifies where heat concentrates, the systems-thermodynamicist designs the cooling architecture, the capacity-engineer sizes the cooling system, and the thermal-monitor instruments the system to detect overheating before it causes failure.

The rite's defining constraint is "exhaust alternatives first." The most common failure mode in cache architecture is adding a cache layer without evaluating whether a simpler solution (index optimization, query rewrite, CDN, read replica) would achieve the same result with less complexity. The heat-mapper's 6-gate framework enforces this evaluation before any cache is recommended.

See `rites/thermia/workflow.yaml` for the full specification.

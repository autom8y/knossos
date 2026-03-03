# Transfer Artifacts

> Templates for TRANSFER documents and cross-rite HANDOFF artifacts produced by Tech Transfer.

## TRANSFER Artifact Template

The TRANSFER document is the internal R&D summary that feeds the HANDOFF.

```markdown
# TRANSFER: [Initiative Name]

## Prototype Summary
- What was built
- What was validated
- Key constraints discovered

## Production Gap Analysis

| Category | Prototype State | Production Requirement | Gap Severity |
|----------|-----------------|------------------------|--------------|
| Error handling | Happy path only | Full error taxonomy | HIGH |
| Scalability | 100 req/min | 10K req/min | MEDIUM |
| Security | No auth | OAuth + rate limiting | HIGH |
| Monitoring | Console logs | Structured + alerting | MEDIUM |

## Requirements Translation

### Functional Requirements
- [FR-001] System must...
- [FR-002] System must...

### Non-Functional Requirements
- [NFR-001] Latency < 100ms p99
- [NFR-002] Availability > 99.9%

### Constraints (from prototype)
- Model size must stay under 500MB (validated in prototype)
- Inference latency baseline: 50ms (measured in prototype)

## Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| API rate limits at scale | Medium | High | Implement caching layer |
| Model accuracy drift | Low | Medium | Add monitoring + retraining pipeline |

## Recommendation
[GO/NO-GO for productionization with rationale]
```

## HANDOFF Production

When R&D work is ready for another rite, produce a HANDOFF artifact using the `cross-rite-handoff` schema.

### Target Rite Routing

| Exploration Outcome | Target Rite | Handoff Type | Example |
|--------------------|-------------|--------------|---------|
| Validated prototype ready for production | 10x-dev | implementation | "ML search prototype achieved 85% relevance; ready to build" |
| Strategic finding without immediate implementation | strategy | strategic_evaluation | "WebAssembly viable but not urgent; consider for H2 roadmap" |
| Both production-ready AND strategic | Both (separate HANDOFFs) | implementation + strategic_evaluation | Major technology bet with immediate application and long-term implications |

### Decision Criteria for Target Selection

Route to **10x-dev** when:
- Prototype is validated and feasibility is proven
- Production gaps are identified and addressable
- Business case supports immediate implementation
- Recommendation is "GO" for productionization

Route to **strategy** when:
- Exploration informs roadmap decisions
- Findings change strategic assumptions
- Technology readiness affects multi-quarter planning
- Recommendation is "WAIT" or "MONITOR" for productionization

### HANDOFF Example (to 10x-dev)

```yaml
---
source_rite: rnd
target_rite: 10x-dev
handoff_type: implementation
created: 2026-01-02
initiative: ML-Powered Search
priority: high
---

## Context

Two-week R&D spike validated ML search approach. Prototype achieved 85% relevance improvement with acceptable latency. Production gaps identified and documented. Ready for implementation.

## Source Artifacts
- .ledge/spikes/PROTOTYPE-ml-search.md
- .ledge/spikes/EVALUATION-ml-search.md
- .ledge/spikes/TRANSFER-ml-search.md

## Items

### IMP-001: Productionize ML search service
- **Priority**: High
- **Summary**: Build production ML search from validated prototype
- **Prototype Evidence**: 85% relevance improvement, 50ms inference latency
- **Production Gaps**:
  - Error handling: Add retry logic and fallback to keyword search
  - Scalability: Add caching layer for embedding queries
  - Monitoring: Add latency histograms and accuracy metrics
- **Constraints** (validated in prototype):
  - Model size < 500MB
  - Inference latency < 100ms p99
- **Acceptance Criteria**:
  - Production-grade error handling with graceful degradation
  - Monitoring and alerting for latency and accuracy
  - Performance matches or exceeds prototype benchmarks
  - Rollback path to keyword search

## Notes for Target Rite

Prototype code is throwaway but algorithm is validated. Recommend starting fresh with production patterns. Prototype engineer available for knowledge transfer.
```

### HANDOFF Example (to strategy)

```yaml
---
source_rite: rnd
target_rite: strategy
handoff_type: strategic_evaluation
created: 2026-01-02
initiative: Technology Radar Update
priority: medium
---

## Context

Evaluated WebAssembly for compute-intensive operations. Technically viable with 3x performance improvement, but ecosystem maturity suggests waiting.

## Source Artifacts
- .ledge/spikes/EXPLORATION-wasm-evaluation.md
- .ledge/spikes/PROTOTYPE-wasm-image-processing.md

## Items

### EVAL-001: WebAssembly strategic assessment
- **Priority**: Medium
- **Summary**: Determine if WebAssembly warrants roadmap inclusion
- **Technical Findings**:
  - 3x performance improvement in compute-heavy operations
  - 200ms cold start (acceptable for non-critical paths)
  - Debugging tooling immature
- **Evaluation Criteria**:
  - Market differentiation: Moderate (competitors not using yet)
  - Investment: Medium (3-4 weeks for production-ready implementation)
  - Risk: Medium (tooling gaps require workarounds)
- **Recommendation**: WAIT - revisit in H2 when tooling matures

## Notes for Target Rite

Observable signals for revisiting: Chrome DevTools WASM debugging GA, major framework adoption (React/Vue).
```

## Integration with TRANSFER Artifact

The TRANSFER document feeds the HANDOFF:

```
TRANSFER (internal)          HANDOFF (cross-rite)
+-- Gap Analysis        -->   Production Gaps in items
+-- Requirements        -->   Acceptance Criteria
+-- Constraints         -->   Constraints section
+-- Risks               -->   Notes for Target Rite
+-- Recommendation      -->   Priority + routing decision
```

Always produce TRANSFER first, then extract relevant portions into HANDOFF.

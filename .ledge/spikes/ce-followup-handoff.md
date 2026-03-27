---
title: Contextual Equilibrium Follow-up Handoff
date: 2026-03-27
initiative: contextual-equilibrium-followup
frame: .sos/wip/frames/contextual-equilibrium-followup.md
---

# CE Follow-up: Handoff from 10x-dev

## Completed Workstreams

### WS-A: Diagnostic Uplift (COMPLETE, commit 7b604ec2)

Extended `ari serve query --diagnostic` to surface all five CE mechanisms:
- MatchType per triage candidate (`[section]` flag)
- Graph injection count and provenance (`[graph-injected]` flag)
- Diversity floor enforcement events (type, candidate, score, summary)
- Per-domain-type token breakdown with percentages
- MaxTypeFraction ceiling hits (skipped/packed, summary substitution)

Added `CEDiagnostics` struct threaded from assembler -> pipeline -> response.
Fixed section candidate domain resolution (section QNs now resolve to parent
document for provenance chain and assembler domain tracking).

### WS-B: Local Validation (COMPLETE, commit 7b604ec2)

Validated SC-1 through SC-6 against live queries:

| SC | Result | Evidence |
|----|--------|----------|
| SC-1 multi-type citations | PASS | arch + conventions in error handling query |
| SC-2 arch token fraction | PASS | arch 79-85% with ceiling active |
| SC-3 specialist citation | PASS | conventions cited for specialist queries |
| SC-4 section packing | PASS | section candidates packed, visible in diagnostic |
| SC-5 no recall regression | PASS | arch 85% primary for architecture queries |
| SC-6 graph injection | PARTIAL | 0 injections (no graph edges in persisted index) |

R-2 (section decontextualization): Assessed as low-severity. Section headings
provide sufficient context. Parent header injection not needed at this time.

### WS-E: Observability Hardening (COMPLETE, commit 4067c446)

Added five CE-specific CloudWatch EMF metrics to `MetricsRecorder`:
- `clew_ce_section_candidate_total`
- `clew_ce_graph_injected_total`
- `clew_ce_diversity_floor_enforced_total{domain_type}`
- `clew_ce_type_ceiling_hit_total{domain_type}`
- `clew_ce_assembler_type_fraction{domain_type}`

Emitted from `InstrumentedPipeline` using CEDiagnostics from response.
Graph injection count available via existing slog structured logging.

## Remaining Workstreams (Operational)

### WS-C: Content Refresh + Container Build

**Status**: Not started. Requires operational execution.
**Steps**:
1. `deploy/scripts/collect-content.sh --sync` (requires sibling repo clones)
2. `docker build -f deploy/Dockerfile -t clew:latest .`
3. Smoke test: `docker run --rm -e ANTHROPIC_API_KEY=... -e KNOSSOS_ORG=autom8y clew:latest ari serve query "what are the scar tissue patterns?"`

**Gate**: Container starts, BM25 index builds with CE params (b=0.55), response cites 2+ domain types.

### WS-D: Production Deploy

**Status**: Not started. Requires ECR/ECS access (rite transition boundary).
**Steps**:
1. Push to ECR
2. Update ECS task definition
3. Monitor CloudWatch for 30 minutes: stage latency, Haiku calls, dropped requests
4. Confirm `/ready` health check

**Gap**: CE-specific metrics won't appear in CloudWatch until after deploy
(metrics code is in the new container image).

### WS-F: Knowledge Refresh (Deferrable)

**Status**: Not started.
**Steps**:
1. `/know --domain=architecture` (incremental pass)
2. `/know --domain=test-coverage`
3. `/know --domain=design-constraints`

## Known Gaps

1. **SC-6 Graph Injection**: No graph edges in the persisted knowledge index.
   The graph injection mechanism is correct (unit-tested) but won't activate
   until the knowledge index is rebuilt with graph edges populated.

2. **Graph injection metric**: `IncrGraphInjected()` is defined on the interface
   but not wired from the triage orchestrator. Graph injection count is available
   via slog structured logging (`graph_injected=N`).

3. **Section candidate freshness**: Fixed in this commit for the assembler path,
   but the provenance chain still uses parent document freshness for section
   candidates (appropriate behavior — sections inherit parent freshness).

---
artifact_id: HANDOFF-10x-dev-to-sre-2026-03-27
schema_version: "1.0"

source_rite: 10x-dev
target_rite: sre

handoff_type: validation
priority: medium
blocking: false

initiative: contextual-equilibrium-followup
created_at: "2026-03-27T19:30:00Z"
status: pending

items:
  - id: CE-OPS-001
    summary: "Refresh deploy/content/ from org repos and rebuild container image with CE mechanisms"
    priority: medium
    validation_scope:
      - "deploy/scripts/collect-content.sh --sync completes without error"
      - "docker build -f deploy/Dockerfile -t clew:latest . succeeds"
      - "Container smoke test: ari serve query produces response citing 2+ domain types"
      - "BM25 index builds with Clew params (b=0.55) — confirmed by 'Clew BM25 index built' log line"
    notes: |
      Content was last refreshed before the contextual-equilibrium commit (af33d01a).
      The new Clew BM25 index at b=0.55 will index whatever is in deploy/content/.
      Stale content means vocabulary amplification and section-level indexing operate
      on an outdated knowledge snapshot. Sibling repo clones required at parent of
      knossos directory.
    estimated_effort: "30m"

  - id: CE-OPS-002
    summary: "Deploy updated container to production ECS/Fargate"
    priority: medium
    validation_scope:
      - "Push clew:latest to ECR"
      - "Update ECS task definition or trigger rolling deployment"
      - "/ready health check returns 200 within 5 minutes of deploy"
      - "CloudWatch shows first post-CE queries completing without error"
      - "Monitor for 30 minutes: RecordStageLatency, RecordHaikuCalls, IncrDropped stable"
    dependencies:
      - "CE-OPS-001"
    notes: |
      The MetricsRecorder interface now includes 5 CE-specific metrics
      (clew_ce_section_candidate_total, clew_ce_graph_injected_total, etc.)
      but they only emit after this deploy lands the new container image.
      Watch for startup coherence warnings (domains missing content).
    estimated_effort: "45m"

  - id: CE-OPS-003
    summary: "Verify CE-specific CloudWatch metrics appear after production queries"
    priority: low
    validation_scope:
      - "clew_ce_section_candidate_total emits for queries hitting section candidates"
      - "clew_ce_assembler_type_fraction{domain_type} emits per-query type breakdown"
      - "clew_ce_type_ceiling_hit_total emits when MaxTypeFraction triggers"
      - "clew_ce_diversity_floor_enforced_total emits when floor types are force-included"
    dependencies:
      - "CE-OPS-002"
    notes: |
      Graph injection metric (clew_ce_graph_injected_total) may show 0 if the
      knowledge index lacks graph edges. This is expected until knowledge index
      is rebuilt with graph data populated.
    estimated_effort: "15m"

source_artifacts:
  - ".ledge/spikes/ce-followup-handoff.md"
  - ".sos/wip/frames/contextual-equilibrium-followup.md"
  - ".ledge/decisions/ADR-contextual-equilibrium-parameters.md"
---

# HANDOFF: 10x-dev -> SRE — Contextual Equilibrium Production Deployment

## Context

The contextual-equilibrium initiative landed five retrieval diversification mechanisms
at commit `af33d01a`. The code is on main, validated locally (WS-A/WS-B), and instrumented
with CloudWatch metrics (WS-E). It has **not been deployed**.

Three operational workstreams remain:
1. **Content refresh + container build** (WS-C) — new content needed for CE mechanisms to index
2. **Production deploy** (WS-D) — push updated container to ECS/Fargate
3. **Metrics verification** (WS-E validation) — confirm CE metrics appear in CloudWatch

## What 10x-dev Completed

| Workstream | Commit | Status |
|-----------|--------|--------|
| WS-A: Diagnostic Uplift | `7b604ec2` | All 5 CE mechanisms observable in `--diagnostic` |
| WS-B: Local Validation | `7b604ec2` | SC-1 through SC-5 PASS, SC-6 PARTIAL (env-limited) |
| WS-E: Observability Code | `4067c446` | 5 new EMF metrics on MetricsRecorder interface |

## What SRE Needs to Execute

### Step 1: Content Refresh (CE-OPS-001)
```bash
cd /path/to/knossos
deploy/scripts/collect-content.sh --sync
# Verify: check collection summary for repos processed, domains missing
```

### Step 2: Container Build
```bash
docker build -f deploy/Dockerfile -t clew:latest .
# Smoke test:
docker run --rm \
  -e ANTHROPIC_API_KEY=... \
  -e KNOSSOS_ORG=autom8y \
  clew:latest ari serve query "what are the scar tissue patterns?"
```

### Step 3: Deploy (CE-OPS-002)
```bash
# Push to ECR
docker tag clew:latest <account>.dkr.ecr.<region>.amazonaws.com/clew:latest
docker push <account>.dkr.ecr.<region>.amazonaws.com/clew:latest

# Update ECS (task definition update or force new deployment)
# Monitor /ready endpoint and CloudWatch for 30 minutes
```

### Step 4: Metrics Verification (CE-OPS-003)
After 3+ production Slack queries, verify in CloudWatch:
- `clew_ce_assembler_type_fraction` shows per-type token distribution
- `clew_ce_section_candidate_total` increments for section-matched queries
- `clew_ce_type_ceiling_hit_total` increments when architecture exceeds 50% budget

## Risk Areas

- **Content staleness**: If `collect-content.sh` fails for some repos, CE mechanisms
  operate on partial content. Fail-open — the pipeline still works, just with fewer
  domains indexed.
- **Graph injection inactive**: The persisted knowledge index may lack graph edges.
  Graph injection will show 0 until the index is rebuilt with graph data. This is
  not a regression — it's a feature not yet fully activated.
- **SC-6 gap**: Graph injection activation rate cannot be validated in production
  until graph edges are populated. Consider rebuilding the knowledge index post-deploy.

---
title: Content Pipeline Health — SRE Handoff to Ecosystem
initiative: content-pipeline-health
sprint: sprint-4
date: 2026-03-27
---

# SRE -> Ecosystem Handoff

## Pipeline Health Summary

The content pipeline is now operating at full fidelity:

| Component | Before | After |
|-----------|--------|-------|
| Summary generation | Cascade timeouts (40+ failures) | Parallel (128/128 in ~65s) |
| Cold-start | 10-min LLM background build | 4-second JSON load |
| Knowledge index | Built on startup (unreliable) | Pre-baked in image (128 domains, 128 summaries) |
| Freshness enforcement | None | --check-freshness flag, fails at >10 stale |
| CE mechanism fidelity | FM-3 and WS-5 degraded silently | Both have summaries for all indexed domains |

## Deploy Verification Results

Production logs (2026-03-27T21:07):
```
loaded pre-baked knowledge index  path=/...  domains=128  summaries=128
BM25 index built  documents=128  sections=840
Clew BM25 index built  documents=128  sections=840  b_param=0.55
startup coherence validation complete  knowledge_domains=128  knowledge_missing=0
```

Startup time: 4 seconds. No background build triggered.

## Registry Recommendation

**autom8** and **autom8y-workflows** both appear in `domains.yaml` with 0 domains. Confirmed
both repos exist locally and have no `.know/` directories. This is genuine absence, not a sync
failure.

**Recommendation: Leave as-is (option C).**

Rationale:
- Both repos are active (autom8 is a Python monorepo, autom8y-workflows has README)
- Their presence in the catalog is harmless (no startup warnings for empty domain lists)
- Removing them provides no benefit; they'll re-appear on next `ari registry sync`
- If either repo grows to merit knowledge, having the catalog entry is a reminder

If the ecosystem rite disagrees, removal is trivial:
```
ari org remove-repo autom8 --org autom8y
ari org remove-repo autom8y-workflows --org autom8y
```

## Commits in This Initiative

| Commit | Sprint | Description |
|--------|--------|-------------|
| `58a5def9` | Sprint-1 | fix(knowledge): remove mutex serialization from summary generation |
| `5501b0aa` | Sprint-1 | chore(content): refresh knossos .know/ + handoff artifact |
| `5c2faca8` | Sprint-2 | feat(deploy): pre-bake knowledge-index.json into Docker image |
| `680df418` | Sprint-3 | feat(deploy): add --check-freshness flag to collect-content.sh |
| `2e5411a4` | Sprint-4 | fix(deploy): add CLEW_KNOWLEDGE_INDEX_PATH env var |

## Remaining Items

1. **Content freshness**: 50 stale domains remain. First Monday cadence run will address these.
2. **ECS task definition**: Updated to `clew:21` with `:latest` tag. Future deploys should use `:latest`.
3. **Build pipeline**: `docker buildx build --platform linux/amd64` required for ECS Fargate.

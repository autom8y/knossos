---
type: gap-analysis
initiative: content-pipeline-health
sprint: 5
date: 2026-03-27
status: resolved
complexity: PATCH
---

# Gap Analysis: Empty-Domain Repos in Registry (autom8, autom8y-workflows)

## Root Cause

No bug or code defect. Both repos correctly appear in `deploy/registry/domains.yaml` with
`domains: []` because neither has a `.know/` directory. The `syncRepo()` function in the
registry sync pipeline handles this correctly -- it returns an empty domain list, which is
faithfully recorded.

The gap is a **content coverage gap**, not a code gap: the primary product repo (autom8)
has no `.know/` files despite having full knossos integration and 2,152 Python source files.

## Investigation

### Registry State (verified in domains.yaml)

| Repo | last_synced | domains | File line |
|------|-------------|---------|-----------|
| autom8 | 2026-03-25T12:42:15Z | `[]` | L5-9 |
| autom8y-workflows | 2026-03-25T12:42:50Z | `[]` | L880-884 |

### Disk State (verified locally)

| Check | autom8 | autom8y-workflows |
|-------|--------|-------------------|
| `.know/` exists | No | No |
| `.claude/` exists | Yes (agents, commands, skills, rules) | No |
| `.knossos/` exists | Yes (active rite, manifests) | No |
| Source files | 2,152 .py files | 3 files total |
| Knowledge potential | HIGH (30 API integrations, adapters, terraform) | NEGLIGIBLE (1 YAML workflow) |

### Verdict per check

- **CONFIRMED**: Both repos have `domains: []` -- matches SRE handoff report.
- **CONFIRMED**: `syncRepo()` behavior is correct -- no code fix needed.
- **CONFIRMED**: autom8 has substantial content for knowledge generation.
- **CONFIRMED**: autom8y-workflows is too thin for meaningful knowledge.

## Success Criteria

- [ ] ADR written documenting the split decision (Option B for autom8, Option C for workflows)
- [x] ADR committed at `.ledge/decisions/ADR-content-pipeline-empty-repos.md`
- [ ] Follow-up: `ari know --all` run in autom8 (future sprint)
- [ ] Follow-up: `ari registry sync` confirms autom8 domains populated (future sprint)

## Complexity: PATCH

This is a documentation-only decision sprint. No code changes. The follow-up action (knowledge
generation for autom8) is a standard `ari know --all` invocation, not a code modification.

## Test Satellites

Not applicable -- no code changes to verify. The follow-up knowledge generation will be
verified by running `ari registry sync` and confirming non-empty domains for autom8.

## Artifacts Produced

| Artifact | Path | Status |
|----------|------|--------|
| ADR | `/Users/tomtenuta/Code/knossos/.ledge/decisions/ADR-content-pipeline-empty-repos.md` | Written, verified |
| Gap Analysis | `/Users/tomtenuta/Code/knossos/.ledge/reviews/GAP-content-pipeline-empty-repos.md` | This file |

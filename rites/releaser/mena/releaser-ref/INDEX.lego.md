---
name: releaser-ref
description: "Releaser rite methodology reference. Use when: implementing release agents, orchestrating multi-repo releases, understanding artifact chain, checking complexity levels, detecting package ecosystems, applying DAG-branch failure halting, routing CI failures to peer rites. Triggers: release orchestration, artifact chain, ecosystem detection, publish order, dependency graph, DAG-branch halting, PATCH escalation, release complexity, cross-rite routing, release anti-patterns. Companion files available for pipeline chains, ecosystem detection, failure halting, and cross-rite routing detail."
---

# Releaser Methodology Reference

## Artifact Chain

```
PATCH:    cartographer -> platform-state-map.{yaml,md}
                       -> release-executor -> execution-ledger.{yaml,md}
                       -> pipeline-monitor -> verification-report.{yaml,md}

RELEASE:  cartographer -> platform-state-map.{yaml,md}
       -> dependency-resolver -> dependency-graph.{yaml,md}
       -> release-planner -> release-plan.{yaml,md}
       -> release-executor -> execution-ledger.{yaml,md}
       -> pipeline-monitor -> verification-report.{yaml,md}

PLATFORM: Same as RELEASE, full-scope (all matching repos, extended CI timeout)
```

All artifacts written to `.claude/wip/release/`. YAML consumed by downstream agents; MD for human review.
Downstream agents consume YAML only. Never parse the MD summaries programmatically.

## Complexity Levels

| Level | Phases | Use For |
|-------|--------|---------|
| PATCH | recon → execution → verification | Single repo push + CI watch |
| RELEASE | All 5 phases | SDK publish + consumer version bumps |
| PLATFORM | All 5 phases, full scope | Full platform release, all matching repos |

PATCH auto-escalation: if cartographer finds `has_dependents: true` on target repo,
Pythia auto-escalates to RELEASE and informs user before proceeding to dependency-analysis.

## Auto-Escalation (PATCH → RELEASE)

Trigger: cartographer sets `has_dependents: true` on any release-candidate repo.

Pythia response:
1. Read `has_dependents` flag from `platform-state-map.yaml`
2. If true: escalate to RELEASE, notify user ("Target repo has N downstream consumers. Escalating to RELEASE.")
3. Continue from dependency-analysis phase — do NOT re-run cartographer

## Anti-Patterns

| Anti-Pattern | Prevention |
|--------------|------------|
| Publishing consumer before SDK dependency | Topological sort in dependency-graph.yaml enforces order; release-executor checks publish confirmation |
| Force-pushing to main without CI | release-executor contract forbids force-push; pipeline-monitor verifies CI before success |
| Bumping versions without publishing | release-executor tracks bump and publish as coupled actions; ledger flags mismatches |
| Treating CI failures as non-blocking | pipeline-monitor contract: never dismiss failures; verification-report.verdict gates success |
| Losing track of processed repos | execution-ledger.yaml tracks every action with timestamps and status |
| Assuming uniform package managers | cartographer detects ecosystem per-repo; release-planner generates repo-specific commands |

## Pre-Flight

Cartographer runs `gh auth status` during reconnaissance. If gh CLI is not authenticated,
fail fast and escalate rather than discovering auth issues 3 phases later during execution.

## Companion Files

For detailed reference, agents should Read the relevant companion:

| Topic | Path |
|-------|------|
| Pipeline Chain Model | `rites/releaser/mena/releaser-ref/pipeline-chains.md` |
| Ecosystem Detection | `rites/releaser/mena/releaser-ref/ecosystem-detection.md` |
| Failure Halting Protocol | `rites/releaser/mena/releaser-ref/failure-halting.md` |
| Cross-Rite Routing | `rites/releaser/mena/releaser-ref/cross-rite-routing.md` |

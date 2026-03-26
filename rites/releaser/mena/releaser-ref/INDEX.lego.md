---
name: releaser-ref
description: "Releaser rite methodology reference. Use when: implementing release agents, orchestrating multi-repo releases, understanding artifact chain, checking complexity levels, detecting package ecosystems, applying DAG-branch failure halting, routing CI failures to peer rites. Triggers: release orchestration, artifact chain, ecosystem detection, publish order, dependency graph, DAG-branch halting, PATCH escalation, release complexity, cross-rite routing, release anti-patterns. Companion files available for pipeline chains and ecosystem detection detail."
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

All artifacts written to `.sos/wip/release/`. YAML consumed by downstream agents; MD for human review.
Downstream agents consume YAML only. Never parse the MD summaries programmatically.

## Complexity Levels

| Level | Phases | Use For |
|-------|--------|---------|
| PATCH | recon → execution → verification | Single repo push + CI watch |
| RELEASE | All 5 phases | SDK publish + consumer version bumps |
| PLATFORM | All 5 phases, full scope | Full platform release, all matching repos |

PATCH auto-escalates to RELEASE if cartographer finds `has_dependents: true`.

## Failure Halting Protocol (DAG-Branch Semantics)

When release-executor reports a failure on repo X:
1. Identify X in the dependency graph
2. Find all repos that depend on X (direct + transitive consumers)
3. Mark all downstream repos as `skipped` in the execution ledger
4. Continue executing repos in branches with no dependency on X
5. Pipeline-monitor only monitors repos that were actually pushed (not skipped)

Goal: maximize successful releases while preventing cascading failures from unpublished dependencies.

## Cross-Rite Routing

| Trigger Signal | Target Rite | When |
|----------------|-------------|------|
| Architectural boundary violations | arch | Repo structure suggests coupling issues |
| Deployment, scaling, infrastructure | sre | CI reveals deployment or reliability issues |
| Code quality blocking publish | hygiene | CI failures from lint/format gates |
| Systematic test failures | review | Failures suggest deeper code issues |
| Security vulnerabilities in CI | security | CI security scan failures |
| Version drift, dependency rot | debt-triage | Accumulated technical debt blocking release |

Route by reference only. Pipeline-monitor names the target rite in recommendations. User decides.
No transitive routing — releaser routes directly to peer rites, never chains.

## Companion Files

For detailed reference, agents should Read the relevant companion:

| Topic | Path |
|-------|------|
| Pipeline Chain Model | `pipeline-chains.md` |
| Ecosystem Detection | `ecosystem-detection.md` |
| Cartographer Reference | `cartographer-reference.lego.md` |
| Pipeline Monitoring | `pipeline-monitoring.lego.md` |

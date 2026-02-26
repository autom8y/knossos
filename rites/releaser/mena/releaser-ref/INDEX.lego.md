---
name: releaser-ref
description: "Releaser rite methodology reference. Use when: implementing release agents, orchestrating multi-repo releases, understanding artifact chain, checking complexity levels, detecting package ecosystems, applying DAG-branch failure halting, routing CI failures to peer rites. Triggers: release orchestration, artifact chain, ecosystem detection, publish order, dependency graph, DAG-branch halting, PATCH escalation, release complexity, cross-rite routing, release anti-patterns."
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

## Ecosystem Detection Matrix

| Manifest File | Ecosystem | Package Manager | Publish Command (typical) |
|---------------|-----------|-----------------|--------------------------|
| `pyproject.toml` | Python | uv | `uv publish` or justfile target |
| `package.json` | Node | npm | `npm publish` or justfile target |
| `go.mod` | Go | go | `git tag v{version}` + `go list -m` |
| `Cargo.toml` | Rust | cargo | `cargo publish` or justfile target |

Multiple manifest files in one repo = ambiguous ecosystem; escalate to Pythia.
Always detect ecosystem per-repo from manifest files. Never assume uniformity.

## Publish Order Protocol

Topological sort rules:
1. Foundations (no cross-repo dependencies) publish first, in parallel
2. Each subsequent phase depends on all repos in prior phases being published
3. Within a phase, repos with no dependency relationship may publish in parallel
4. Consumer version bumps happen AFTER the dependency's publish is confirmed — never before

Parallel group constraints:
- Two repos may share a phase only if neither depends on the other (directly or transitively)
- If uncertain, be conservative: sequential is safe, incorrect parallel causes failures

## Failure Halting Protocol (DAG-Branch Semantics)

When release-executor reports a failure on repo X:
1. Identify X in the dependency graph
2. Find all repos that depend on X (direct + transitive consumers)
3. Mark all downstream repos as `skipped` in the execution ledger
4. Continue executing repos in branches with no dependency on X
5. Pipeline-monitor only monitors repos that were actually pushed (not skipped)

Goal: maximize successful releases while preventing cascading failures from unpublished dependencies.

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

## Cross-Rite Routing Table

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

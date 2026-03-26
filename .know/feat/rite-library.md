---
domain: feat/rite-library
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./rites/**/*.yaml"
  - "./rites/**/agents/*.md"
  - "./rites/**/mena/**/*.md"
  - "./embed.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# Built-In Rite Library

## Purpose and Design Rationale

The Built-In Rite Library ships 19 pre-packaged workflow rites embedded directly in the `ari` binary via `//go:embed rites` in `embed.go`. This enables any project to get a full suite of orchestrated, multi-agent workflows immediately with no configuration. The resolution chain treats embedded rites as the lowest-priority tier -- project-level overrides take precedence.

Every orchestrated rite uses `potnia` as its entry_agent -- a domain-specific orchestrator that reads `orchestrator.yaml` to generate a customized coordinator. The `shared` rite is architecturally distinct: no phases, no agents, no entry agent -- purely infrastructure skills available to all other rites.

## Conceptual Model

### The 19 Rites

| Rite | Domain | Complexity Levels | Agent Count |
|------|--------|-------------------|-------------|
| 10x-dev | Software development | SCRIPT/MODULE/SERVICE/PLATFORM | 5 |
| arch | Multi-repo architecture | SURVEY/ANALYSIS/DEEP-DIVE | 5 |
| clinic | Debugging | INVESTIGATION (single) | 5 |
| debt-triage | Technical debt | QUICK/AUDIT | 4 |
| docs | Documentation | PAGE/SECTION/SITE | 5 |
| ecosystem | Infrastructure | PATCH/MODULE/SYSTEM/MIGRATION | 6 |
| forge | Agent creation | AGENT/MODULE/SYSTEM | 8 |
| hygiene | Code quality | SPOT/MODULE/CODEBASE | 5 |
| intelligence | Product analytics | METRIC/FEATURE/INITIATIVE | 5 |
| releaser | Release engineering | PATCH/RELEASE/PLATFORM | 6 |
| review | Code review (triage) | QUICK/FULL | 4 |
| rnd | Technology exploration | SPIKE/EVALUATION/MOONSHOT | 6 |
| security | Security assessment | PATCH/FEATURE/SYSTEM | 5 |
| shared | Cross-rite infrastructure | (none) | 0 |
| slop-chop | AI code quality | DIFF/MODULE/CODEBASE | 6 |
| sre | Site reliability | ALERT/SERVICE/SYSTEM/PLATFORM | 5 |
| strategy | Business strategy | TACTICAL/STRATEGIC/TRANSFORMATION | 5 |
| thermia | Cache architecture | QUICK/STANDARD/DEEP | 5 |
| ui | UI/UX development | COMPONENT/FEATURE/SYSTEM | 9 |

### Structural Invariants

1. Entry agent is always `potnia` (18/19 rites)
2. All rites depend on `shared`
3. Phases produce named artifacts
4. Back-routes are first-class with optional iteration limits
5. Complexity gates phases

## Implementation Map

Each rite follows the layout: `manifest.yaml`, `workflow.yaml`, `orchestrator.yaml`, `agents/`, `mena/`, optional `hooks/`. Embedding via `EmbeddedRites embed.FS` in `embed.go`. Resolution: project > user > org > platform > embedded.

The `review` rite is the generalist intake funnel with cross-rite routing to 8 downstream rites. The `shared` rite provides 8 cross-cutting skills materialized alongside every active rite.

## Boundaries and Failure Modes

- Back-route cycles: most rites lack `max_iterations` -- misbehaving agents could loop indefinitely
- Rite switch during active session changes materialized agents without session context update
- `arch` rite has no mena (agents rely solely on shared skills)
- `ui` rite has stale workflow.yaml referencing renamed agent
- `external_consultation` enforcement is documentation convention, not runtime enforcement
- Hook directories are empty across all rites (hooks registered globally via `config/hooks.yaml`)

## Knowledge Gaps

1. Orchestrator template rendering pipeline not traced
2. `agent_defaults` and `skill_policies` inheritance merge algorithm not fully traced
3. Complexity condition evaluation location (potnia prompt vs framework evaluator) unconfirmed
4. `external_consultation` runtime enforcement not found in Go source

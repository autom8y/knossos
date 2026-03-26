---
domain: feat/rite-management
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/rite/**/*.go"
  - "./internal/cmd/rite/**/*.go"
  - "./rites/*/manifest.yaml"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# Rite Management and Invocation

## Purpose and Design Rationale

A rite is knossos's core unit of organizational purpose -- a composable bundle of agents, skills, hooks, and workflow for a specific domain. Two axes: active rite (replacement via ari sync) and invocation (additive borrowing, bounded by token budget). Four forms: simple (skills-only), practitioner (agents+skills), procedural (hooks+workflow), full (all). Context budget (default 50K tokens) guardrails invocation cost. Resolution chain: project > user > org > platform > embedded.

## Conceptual Model

**Rite manifest (manifest.yaml):** name, entry_agent, phases, agents, dromena/legomena, dependencies, complexity_levels, hooks, mcp_servers. **Active rite state:** .knossos/ACTIVE_RITE (plain text). **Invocation state:** INVOCATION_STATE.yaml (borrowings + budget tracking). **Context injection:** context.yaml key-value rows rendered into inscription. **19 bundled rites** in rites/ directory.

## Implementation Map

`internal/rite/` (10 files): manifest.go (RiteManifest, Validate), discovery.go (4-tier enumeration), invoker.go (11-step borrow pipeline), state.go (InvocationState persistence), budget.go (BudgetCalculator), workflow.go, context.go/context_loader.go (4-tier context resolution), validate.go (6 checks), syncer.go (dependency inversion interface). `internal/cmd/rite/` (10 files): list, info, current, invoke, release, status, validate, context, pantheon subcommands.

## Boundaries and Failure Modes

Rite management vs materialization: internal/rite does NOT write channel directories (Syncer interface bridges to materialize). Invocation is not materialization (CLAUDE.md injection deferred to "Phase 2"). Active rite written by materialize, read by rite. ContextChain inverts user/project priority. Missing ACTIVE_RITE: CodeFileNotFound. Invocation conflicts: ErrBorrowConflict. Budget exceeded: default 50K tokens with rough estimates. Resolution shadowing is silent. loadRite skips invalid rites silently.

## Knowledge Gaps

1. ari sync --rite full switching flow internals not read
2. Dependency resolution for shared rite not traced through materialize
3. ADR-0007 not found on disk

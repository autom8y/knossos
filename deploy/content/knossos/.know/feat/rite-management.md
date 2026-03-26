---
domain: feat/rite-management
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/rite/**/*.go"
  - "./internal/cmd/rite/**/*.go"
  - "./rites/*/manifest.yaml"
  - "./docs/decisions/ADR-0007*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.92
format_version: "1.0"
---

# Rite Management and Invocation

## Purpose and Design Rationale

Rites are composable practice bundles — the primary unit of context switching. Each rite encapsulates agents, mena, hooks, and workflow configuration appropriate for a specific engineering practice domain.

**ADR-0007**: YAML-based rite context over bash-sourced `context-injection.sh`. Problems: arbitrary code execution, no schema validation, shell spawning overhead.

### Two Composition Modes

1. **Rite switching** (`ari sync --rite <name>`): replaces active rite entirely
2. **Additive invocation** (`ari rite invoke`): borrows components without switching (CLAUDE.md injection deferred to Phase 2)

## Conceptual Model

### Key Abstractions

- **Rite**: Directory under `rites/<name>/` with `manifest.yaml`, `agents/`, `mena/`
- **Active Rite**: Currently materialized rite, tracked by `.claude/ACTIVE_RITE`
- **Rite Form**: `simple` (mena only), `practitioner` (agents+mena), `procedural` (hooks+workflows), `full`
- **Invocation**: Borrowed components from non-active rite, stored in `INVOCATION_STATE.yaml`
- **Budget**: Token cost tracking (2,000/agent, 1,000/skill, 500/workflow)

### 18 Embedded Rites

10x-dev, arch, clinic, debt-triage, docs, ecosystem, forge, hygiene, intelligence, releaser, review, rnd, security, shared, slop-chop, sre, strategy, thermia.

## Implementation Map

Domain: `/Users/tomtenuta/Code/knossos/internal/rite/` (16 files). CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/rite/` (11 files).

### Key Types

`RiteManifest`, `RiteForm`, `Discovery`, `Invoker`, `InvocationState`, `StateBudget`, `Validator`, `RiteContext`, `ContextLoader`, `Workflow`.

### Key Flows

- **Discovery**: `scanDir(.knossos/rites/)` → `scanDir(orgRitesDir)` → `scanDir(userRitesDir)` → sort → mark active
- **Invoke**: load target manifest → validate → load state → detect conflicts → select components → estimate budget → generate ID → save state

### TENSION-003

Dual `RiteManifest` types: `internal/rite.RiteManifest` (invocation) vs `internal/materialize.RiteManifest` (pipeline). Import graph isolation.

## Boundaries and Failure Modes

- `ari rite invoke` only records intent — does NOT update CLAUDE.md or `.claude/` (Phase 2 deferred)
- Discovery precedence: user overrides project (opposite of materialization's 6-tier resolution)
- `CleanExpired()` exists but has no caller (expired invocations accumulate)
- `skills` field polymorphism in manifest.yaml (string[] or SkillRef[])
- Budget estimation is approximate (flat defaults, not file-scanning)

## Knowledge Gaps

1. `ari rite invoke` Phase 2 has no ADR/issue/implementation plan.
2. `CleanExpired()` callers: none exist in CLI.
3. `orchestrator.yaml` fallback fidelity untested across rites.

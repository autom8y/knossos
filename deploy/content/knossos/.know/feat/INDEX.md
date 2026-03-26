---
domain: feat/index
generated_at: "2026-03-03T20:45:00Z"
expires_after: "30d"
source_scope:
  - "./rites/*/manifest.yaml"
  - "./internal/*/"
  - "./docs/decisions/ADR-*.md"
  - "./.claude/commands/*.md"
  - "./.claude/agents/*.md"
  - "./INTERVIEW_SYNTHESIS.md"
  - "./.know/*.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.93
format_version: "1.0"
---

# Feature Census

> 28 features identified across 6 categories. 22 recommended for GENERATE, 6 recommended for SKIP.

## session-lifecycle

| Field | Value |
|-------|-------|
| Name | Session Lifecycle Management |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.98 |

**Source Evidence**:
- `internal/session/` (22 files): FSM, lifecycle, context, snapshot, timeline, rotation, discovery, event reading
- `internal/cmd/session/` (35 files): create, park, resume, wrap, fray, gc, audit, lock, snapshot, timeline, log, field, migrate, recover, transition subcommands
- `docs/decisions/ADR-0001-session-state-machine-redesign.md`: foundational FSM design
- `docs/decisions/ADR-0022-session-model.md`: full session model architecture
- `docs/decisions/ADR-0027-unified-event-system.md`: event system unification
- `internal/validation/schemas/session-context.schema.json`: schema definition

**Rationale**: 57 implementation files across two packages, 3 ADRs, user-facing CLI surface with 15+ subcommands, FSM with TLA+ spec, event log, status enum, and session forking (fray). Meets every GENERATE heuristic.

---

## session-event-system

| Field | Value |
|-------|-------|
| Name | Unified Session Event System |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.92 |

**Source Evidence**:
- `internal/hook/clewcontract/orchestrator.go`: throughline extraction, orchestrator handoff events
- `internal/hook/clewcontract/record_test.go`: event write path
- `internal/session/events_read.go`: dual-schema format bridge (pre/post ADR-0027)
- `docs/decisions/ADR-0027-unified-event-system.md`: unified event schema decision

**Rationale**: 1 ADR dedicated to this feature, distinct `events.jsonl` log format with dual-schema bridge logic, consumed by session, sails, and hook packages. Cross-cutting dependency confirms GENERATE.

---

## session-forking

| Field | Value |
|-------|-------|
| Name | Session Forking (Fray) |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/cmd/session/fray.go`: fork session into parallel strand
- `internal/session/context.go`: `FrayedFrom`/`Strands` fields
- `docs/decisions/ADR-0006-parallel-session-orchestration.md`: parallel session pattern

**Rationale**: Dedicated subcommand with its own file, distinct fields in the session context schema, 1 ADR describing the parallel orchestration pattern. User-facing CLI surface exists.

---

## materialization-pipeline

| Field | Value |
|-------|-------|
| Name | Rite Materialization Pipeline |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.99 |

**Source Evidence**:
- `internal/materialize/` (53 files): full pipeline — agents, mena, CLAUDE.md, rules, settings, hooks, provenance, worktree, skill policies
- `internal/cmd/sync/sync.go`, `budget.go`: CLI entry points
- `docs/decisions/ADR-sync-materialization.md` (ADR-0016): sync/materialization model
- `docs/decisions/TDD-single-binary-completion.md`: single-binary embedded assets
- `README.md`: documents `ari sync` patterns

**Rationale**: Dominant hub package (53 files), 2 decision records, the primary user-facing command (`ari sync`), 6-tier rite resolution, 5 sub-packages, idempotent WriteIfChanged pattern. Unambiguous GENERATE.

---

## rite-management

| Field | Value |
|-------|-------|
| Name | Rite Management and Invocation |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `internal/rite/` (14 files): manifest, invoker, budget, context, discovery, state, workflow, validate, syncer
- `internal/cmd/rite/` (11 files): list, current, validate, invoke, context, release, status, info, pantheon
- `rites/` directory: 17 rites (10x-dev, arch, clinic, debt-triage, docs, ecosystem, forge, hygiene, intelligence, releaser, review, rnd, security, shared, slop-chop, sre, strategy, thermia)
- `docs/decisions/ADR-0007-team-context-yaml-architecture.md`: rite context architecture

**Rationale**: 25 implementation files, 1 ADR, user-facing CLI surface with 6+ subcommands, 18 embedded rites, budget calculator, invocation state machine. GENERATE.

---

## inscription-system

| Field | Value |
|-------|-------|
| Name | CLAUDE.md Inscription System |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `internal/inscription/` (15 files): marker parser, merger, generator, pipeline, manifest, backup, sync, types
- `internal/cmd/inscription/` (6 files): sync, diff, validate, rollback subcommands
- `knossos/templates/sections/` (8 template files): agent-configurations, agent-routing, commands, execution-mode, know, platform-infrastructure, quick-start, user-content
- `docs/decisions/ADR-0021-two-axis-context-model.md`: two-axis context model (skills + commands unification)

**Rationale**: 21 implementation files, 1 ADR, user-facing `ari inscription` CLI with 4 subcommands, marker parser with HTML comment syntax, region ownership model (knossos/satellite/regenerate), 8 Go templates.

---

## hook-infrastructure

| Field | Value |
|-------|-------|
| Name | CC Hook Infrastructure |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.98 |

**Source Evidence**:
- `internal/hook/` (5 files): `StdinPayload`, `Env`, `ToolInput`, `HookEvent`, `ParseEnv()`
- `internal/cmd/hook/` (28 files): write-guard, agent-guard, clew, context, autopark, session-end, budget, subagent, git-conventions, precompact, validate, worktree-remove, worktree-seed, cheapo-revert
- `config/hooks.yaml`: hook configuration
- `docs/decisions/ADR-0002-hook-library-resolution-architecture.md`: hook resolution architecture
- `docs/decisions/ADR-0011-hook-deprecation-timeline.md`: Go migration from bash

**Rationale**: 33 implementation files, 2 ADRs, 14 hook subcommands covering all CC lifecycle events (PreToolUse, PostToolUse, Stop, SessionEnd, EnterWorktree), stdin JSON transport. GENERATE.

---

## provenance-system

| Field | Value |
|-------|-------|
| Name | File Provenance Tracking |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `internal/provenance/` (7 files): collector, divergence, manifest, merge, provenance, report, scanner
- `internal/cmd/provenance/provenance.go`: CLI entry point
- `docs/decisions/ADR-0026-unified-provenance.md`: unified provenance model decision

**Rationale**: 8 implementation files, 1 major ADR (unified three prior systems), user-facing `ari provenance` command, central safety gate preventing user content overwrite. GENERATE.

---

## mena-system

| Field | Value |
|-------|-------|
| Name | Mena (Dromena + Legomena) System |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `internal/mena/` (8 files): type detection, routing logic for `.dro.md`/`.lego.md`
- `internal/materialize/mena/` (8 files): collect engine, namespace, walker, frontmatter
- `mena/` directory: user-level mena (cem, conventions, guidance, meta, navigation, operations, rite-switching, session, templates, workflow)
- `docs/decisions/ADR-0023-dromena-legomena-mena-convention.md`: dromena/legomena naming convention
- `docs/decisions/ADR-0025-mena-scope.md`: pipeline-targeted mena scope filtering
- `docs/decisions/ADR-0021-two-axis-context-model.md`: two-axis context model (commands/skills)

**Rationale**: 16 implementation files, 3 ADRs, file-extension-based routing (`.dro.md` → commands, `.lego.md` → skills), scope filtering (user/project/both), collision namespace logic. GENERATE.

---

## agent-scaffolding

| Field | Value |
|-------|-------|
| Name | Agent Scaffolding and Factory |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.94 |

**Source Evidence**:
- `internal/agent/` (12 files): scaffold, archetype, templates, frontmatter, mcp_validate, regenerate, sections, validate
- `internal/cmd/agent/agent.go`: CLI entry point
- `docs/decisions/ADR-0024-agent-factory.md`: structured agent authoring with schema validation

**Rationale**: 13 implementation files, 1 ADR, user-facing `ari agent` command, archetype templates, MCP dependency declaration, schema validation. Replaces handcrafted freeform agent authoring.

---

## worktree-management

| Field | Value |
|-------|-------|
| Name | Git Worktree Management |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `internal/worktree/` (8 files): git, lifecycle, metadata, operations, session_integration, worktree
- `internal/cmd/worktree/` (11 files): create, sync, remove, cleanup, clone, export, import, list, status, switch
- `docs/decisions/ADR-0010-worktree-session-seeding.md`: worktree session seeding
- `docs/decisions/ADR-0029-worktree-environment-contract.md`: worktree environment contract

**Rationale**: 19 implementation files, 2 ADRs, 10 subcommands, integration with session lifecycle, hook handlers for `EnterWorktree`, and `.knossos/` state seeding for linked worktrees.

---

## validation-framework

| Field | Value |
|-------|-------|
| Name | JSON Schema Validation Framework |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/validation/` (9 files): artifact, frontmatter, handoff, sails, validator
- `internal/validation/schemas/` (11 schema files): adr, agent, common, handoff-criteria, knossos-manifest, prd, session-context, tdd, test-plan, white-sails
- `internal/cmd/validate/validate.go`: CLI entry point
- `docs/decisions/ADR-0008-handoff-schema-embedding.md`: handoff schema embedding

**Rationale**: 20 implementation + schema files, 1 ADR, user-facing `ari validate` command, 11 JSON schemas covering all artifact types (ADR, PRD, TDD, session, agent). GENERATE.

---

## artifact-registry

| Field | Value |
|-------|-------|
| Name | Work Artifact Registry |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/artifact/` (5 files): aggregate, graduate, query, registry, types
- `internal/cmd/artifact/` (5 files): register, list, query, rebuild, artifact
- `docs/decisions/ADR-0009-knossos-roster-identity.md`: platform identity (references artifact registry)

**Rationale**: 10 implementation files, user-facing `ari artifact` CLI with 4 subcommands (register, list, query, rebuild), phase/type filtering, supports ADR/PRD/TDD artifact types. GENERATE.

---

## white-sails-signaling

| Field | Value |
|-------|-------|
| Name | White Sails Quality Gate Signaling |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.92 |

**Source Evidence**:
- `internal/sails/` (8 files): color, contract, gate, generator, integration_test, proofs, thresholds
- `internal/cmd/sails/sails.go`: CLI entry point
- `internal/validation/schemas/white-sails.schema.json`: schema
- `internal/validation/sails.go`: sails-specific validation

**Rationale**: 10 implementation files, user-facing `ari sails` command, White/Gray/Black confidence levels, clew contract validation, threshold configuration. GENERATE.

---

## clew-handoff-protocol

| Field | Value |
|-------|-------|
| Name | Clew Orchestrator Handoff Protocol |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/hook/clewcontract/orchestrator.go`: throughline extraction
- `internal/cmd/hook/clew.go`: hook handler
- `internal/cmd/handoff/` (6 files): execute, handoff, history, prepare, status
- `internal/validation/schemas/handoff-criteria.schema.json`: handoff schema
- `docs/decisions/ADR-0008-handoff-schema-embedding.md`: handoff schema embedding
- `docs/decisions/ADR-0012-cross-rite-protocol-rename.md`: cross-rite protocol

**Rationale**: 8 implementation files, 2 ADRs, `ari handoff` command with 5 subcommands, handoff criteria schema with machine-verifiable transitions. GENERATE.

---

## tribute-generation

| Field | Value |
|-------|-------|
| Name | TRIBUTE.md Session Wrap Generation |
| Category | Core Platform |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/tribute/` (5 files): extractor, generator, renderer, types
- `internal/cmd/tribute/tribute.go`: CLI entry point
- `internal/cmd/explain/concepts/tribute.md`: concept documentation

**Rationale**: 6 implementation files, user-facing `ari tribute` command, auto-generates session wrap artifacts, consumed by `ari session wrap`. User-facing interface qualifies it for GENERATE.

---

## know-system

| Field | Value |
|-------|-------|
| Name | Codebase Knowledge Domain System (.know/) |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `internal/know/` (8 files): astdiff, know, manifest, validate
- `internal/cmd/knows/knows.go`: CLI entry point with `--delta` flag and `ChangeManifest`
- `internal/cmd/hook/context.go`: context hook injects `.know/` freshness
- `internal/cmd/status/status.go`: HealthDashboard includes `Know` field
- Recent commits: AST-based semantic diffing (`go/ast`) and incremental refresh with `--delta` flag

**Rationale**: 9 implementation files, unique AST-based semantic diffing feature, user-facing `ari knows` command, hook integration for context injection, incremental refresh cycle management. GENERATE.

---

## naxos-orphan-cleanup

| Field | Value |
|-------|-------|
| Name | Naxos Orphan Session Cleanup |
| Category | Core Platform |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/naxos/` (3 files): types, report, report_test
- `internal/cmd/naxos/scan.go`, `naxos.go`: CLI entry point with scan subcommand

**Rationale**: User-facing `ari naxos` command, distinct OrphanReason/SuggestedAction types, session garbage collection. Small but user-facing with no overlapping concerns elsewhere.

---

## org-management

| Field | Value |
|-------|-------|
| Name | Organization-Level Resource Management |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.85 |

**Source Evidence**:
- `internal/cmd/org/` (6 files): init, list, current, set, org, org_test
- `internal/materialize/orgscope/` (2 files): sync, sync_test
- `internal/materialize/source/resolver.go`: 6-tier resolution includes org tier

**Rationale**: 8 implementation files, user-facing `ari org` CLI with 4 subcommands (init, list, current, set), distinct XDG-based org directory at tier 3 of 6-tier rite resolution, cross-cutting org scope affects materialization.

---

## project-initialization

| Field | Value |
|-------|-------|
| Name | Project Initialization (ari init) |
| Category | User-Facing |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/cmd/initialize/init.go`, `init_test.go`: CLI implementation
- `docs/decisions/TDD-single-binary-completion.md`: `ari init` listed as one of three sprint goals
- `internal/cmd/explain/concepts/knossos.md`: concept documentation

**Rationale**: User-facing `ari init` command bootstraps a project without requiring source repo, referenced in TDD, enables zero-install usage. GENERATE.

---

## mena-lint

| Field | Value |
|-------|-------|
| Name | Mena and Agent Lint |
| Category | User-Facing |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.85 |

**Source Evidence**:
- `internal/cmd/lint/lint.go`: source path leak detection, `@skill-name` reference validation, severity levels (CRIT/HIGH/MED/LOW)
- CLI surface: `ari lint`

**Rationale**: User-facing `ari lint` command with distinct detection patterns (source path leaks, broken skill references, extension references). Non-trivial validation surface. GENERATE.

---

## project-status-dashboard

| Field | Value |
|-------|-------|
| Name | Project Health Dashboard (ari status) |
| Category | User-Facing |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.85 |

**Source Evidence**:
- `internal/cmd/status/status.go`: `HealthDashboard` type with Claude, Knossos, Know, Ledge, SOS fields
- CLI surface: `ari status`

**Rationale**: User-facing `ari status` command, unified health dashboard across all managed directories (.claude/, .knossos/, .know/, .ledge/, .sos/), integrates provenance and know packages. GENERATE.

---

## explain-command

| Field | Value |
|-------|-------|
| Name | Concept Documentation Browser (ari explain) |
| Category | User-Facing |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.82 |

**Source Evidence**:
- `internal/cmd/explain/` (5 files): concepts/, context, explain, models
- `internal/cmd/explain/concepts/` (13 concept files): agent, dromena, inscription, knossos, know, ledge, legomena, mena, rite, sails, session, sos, tribute
- CLI surface: `ari explain [concept]`

**Rationale**: User-facing `ari explain` command with 13 built-in concept definitions, JSON output support, project-aware context injection. GENERATE as it constitutes self-documenting CLI behavior.

---

## tour-command

| Field | Value |
|-------|-------|
| Name | Directory Tour (ari tour) |
| Category | User-Facing |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.80 |

**Source Evidence**:
- `internal/cmd/tour/tour.go`: walks .claude/, .knossos/, .know/, .ledge/, .sos/ directories
- CLI surface: `ari tour`

**Rationale**: Fewer than 5 implementation files, no decision records, read-only directory listing with no cross-cutting concerns. Utility wrapper over `ls`-style output that does not introduce domain concepts.

---

## manifest-operations

| Field | Value |
|-------|-------|
| Name | YAML/JSON Manifest Operations |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/manifest/` (6 files): diff, merge, manifest, schema, state
- `internal/cmd/manifest/` (5 files): show, validate, diff, merge

**Rationale**: 11 implementation files, user-facing `ari manifest` CLI with 4 subcommands, 3-way merge, schema validation, diff capabilities. Multiple packages depend on this foundation. GENERATE.

---

## embedded-assets

| Field | Value |
|-------|-------|
| Name | Single-Binary Embedded Asset Distribution |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `embed.go`: `//go:embed` declarations for rites, templates, hooks.yaml, agents, mena
- `docs/decisions/TDD-single-binary-completion.md`: single-binary completion sprint
- `internal/assets/`: embedded asset accessor
- Architecture doc: 6-tier rite resolution, embedded fallback tier

**Rationale**: 1 TDD, user-facing impact (zero-install binary), 5 embedded asset sets, distinct fallback tier in 6-tier resolver. Not a pure utility — drives distribution model. GENERATE.

---

## advisory-file-locking

| Field | Value |
|-------|-------|
| Name | Advisory File Locking |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.85 |

**Source Evidence**:
- `internal/lock/` (4 files): lock, moirai, lock_test, moirai_test

**Rationale**: 4 files, no decision records, internal utility with no user-facing interface. Imported by session and hook packages for concurrency safety but introduces no user-observable domain concepts.

---

## skill-policy-engine

| Field | Value |
|-------|-------|
| Name | Skill Policy Engine |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.78 |

**Source Evidence**:
- `internal/materialize/skill_policies.go`: `SkillPolicy`, `MergeSkillPolicies()`
- `internal/materialize/skill_policies_test.go`

**Rationale**: 2 files, no dedicated decision records (policy logic embedded in ADR-0024), no user-facing CLI, pure implementation detail of materialization. Fewer than 5 files, internal only.

---

## context-budget-tracking

| Field | Value |
|-------|-------|
| Name | Context Token Budget Tracking |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.82 |

**Source Evidence**:
- `internal/cmd/hook/budget.go`, `budget_test.go`: BudgetOutput, warn/park thresholds
- `internal/tokenizer/`: token estimation via tiktoken
- `internal/cmd/sync/budget.go`: budget sync support

**Rationale**: 4 files, no dedicated decision records, internal hook handler with no cross-cutting domain. Threshold-based counter logic — utility within the hook system. SKIP.

---

## registry-system

| Field | Value |
|-------|-------|
| Name | Stable Key Registry |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.80 |

**Source Evidence**:
- `internal/registry/registry.go`, `validate.go` + tests: maps stable keys to agent/skill/CLI refs

**Rationale**: 4 files, no decision records, zero user-facing interface, pure internal utility enabling import-cycle avoidance. SKIP per heuristic: pure utility, no ADRs, fewer than 5 files, internal detail.

---

## Census Gaps

1. **Rite invocation state machine**: `internal/rite/invoker.go` and `internal/rite/state.go` implement an `INVOCATION_STATE.yaml` lifecycle. The full state machine was not deeply read. It may warrant a standalone feature entry (`rite-invocation-state`) but its scope overlaps substantially with `rite-management`.

2. **Context injection hook**: `internal/cmd/hook/context.go` (the hook that injects `.know/` freshness into CC context windows) could be a standalone feature (`context-hook`) distinct from the broader hook infrastructure. Currently folded into `hook-infrastructure`.

3. **Cheapo-revert hook**: `internal/cmd/hook/cheapo_revert.go` is a distinct operational tool (force haiku model override via `ElCheapo` flag). Could be a micro-feature but shares too much surface with the hook system to warrant its own entry.

4. **Sync state tracking**: `internal/sync/` (2 files: `state.go`, `state_test.go`) tracks `.knossos/sync/state.json` timestamps. Too small for a standalone entry but is a dependency of materialization freshness tracking.

5. **Precompact hook**: `internal/cmd/hook/precompact.go` handles CC context compaction events. Not documented in any ADR. Folded into `hook-infrastructure` but may deserve separate documentation.

6. **User-scope vs project-scope sync boundary**: The split between `internal/materialize/userscope/` (syncs to `~/.claude/`) and rite scope (syncs to `.claude/`) is a significant architectural boundary. Covered under `materialization-pipeline` but the user-scope semantics are distinct enough they could be their own entry.

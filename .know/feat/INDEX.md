---
domain: feat/index
generated_at: "2026-03-26T18:57:47Z"
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
source_hash: "b329d719"
confidence: 0.91
format_version: "1.0"
---

# Feature Census

> 46 features identified across 7 categories. 38 recommended for GENERATE, 8 recommended for SKIP.

## session-lifecycle

| Field | Value |
|-------|-------|
| Name | Session Lifecycle Management |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.98 |

**Source Evidence**:
- `internal/session/` (22 files): FSM, lifecycle, context, snapshot, timeline, rotation, discovery, event reading, complexity
- `internal/cmd/session/` (41 files): create, park, resume, wrap, fray, gc, audit, lock, snapshot, timeline, log, field, migrate, recover
- `docs/decisions/ADR-0001-session-state-machine-redesign.md`: foundational FSM design
- `docs/decisions/ADR-0022-session-model.md`: full session model architecture
- `docs/decisions/ADR-0027-unified-event-system.md`: event system unification
- `internal/validation/schemas/session-context.schema.json`: schema definition

**Rationale**: 63 implementation files across two packages, FSM with lifecycle phases and event sourcing, 15+ user-facing subcommands, and 3 dedicated ADRs. Meets every GENERATE heuristic by wide margin.

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

**Rationale**: Dedicated ADR (ADR-0027), distinct `events.jsonl` log format, dual-schema bridge logic, cross-cutting dependency across session, sails, and hook packages. GENERATE.

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

**Rationale**: Dedicated subcommand, distinct fields in session context schema, 1 ADR. User-facing CLI interface exists. GENERATE.

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
- `internal/materialize/` (53+ files): full pipeline -- agents, mena, CLAUDE.md, rules, settings, hooks, provenance, worktree, skill policies
- `internal/cmd/sync/sync.go`, `budget.go`: CLI entry points
- `docs/decisions/ADR-sync-materialization.md` (ADR-0016): sync/materialization model
- `docs/decisions/TDD-single-binary-completion.md`: single-binary embedded assets
- `README.md`: documents `ari sync` patterns

**Rationale**: Dominant hub package (53+ files), 2 decision records, primary user-facing command (`ari sync`), 6-tier rite resolution, 5 sub-packages, idempotent WriteIfChanged pattern. Unambiguous GENERATE.

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
- `rites/` directory: 19 rites (10x-dev, arch, clinic, debt-triage, docs, ecosystem, forge, hygiene, intelligence, releaser, review, rnd, security, shared, slop-chop, sre, strategy, thermia, ui)
- `docs/decisions/ADR-0007-team-context-yaml-architecture.md`: rite context architecture

**Rationale**: 25 implementation files, 1 ADR, user-facing CLI with 6+ subcommands, 19 embedded rites, budget calculator, invocation state machine. GENERATE.

---

## inscription-system

| Field | Value |
|-------|-------|
| Name | CLAUDE.md / GEMINI.md Inscription System |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `internal/inscription/` (15 files): marker parser, merger, generator, pipeline, manifest, backup, sync, types
- `internal/cmd/inscription/` (6 files): sync, diff, validate, rollback, backups
- `knossos/templates/sections/`: template files
- `docs/decisions/ADR-0021-two-axis-context-model.md`: two-axis context model

**Rationale**: 21 implementation files, 1 ADR, user-facing `ari inscription` CLI with 4 subcommands, marker parser with HTML comment syntax, region ownership model (knossos/satellite/regenerate). GENERATE.

---

## hook-infrastructure

| Field | Value |
|-------|-------|
| Name | CC / Gemini Hook Infrastructure |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.98 |

**Source Evidence**:
- `internal/hook/` (5 files): `StdinPayload`, `Env`, `ToolInput`, `HookEvent`, `ParseEnv()`
- `internal/cmd/hook/` (36 files): write-guard, agent-guard, clew, context, autopark, session-end, budget, subagent, git-conventions, precompact, validate, worktree-remove, worktree-seed, cheapo-revert, drift-detect, attribution-guard, suggest
- `config/hooks.yaml`: hook configuration
- `docs/decisions/ADR-0002-hook-library-resolution-architecture.md`: hook resolution architecture
- `docs/decisions/ADR-0011-hook-deprecation-timeline.md`: Go migration from bash

**Rationale**: 41 implementation files, 2 ADRs, 14+ hook subcommands covering all lifecycle events, stdin JSON transport. GENERATE.

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

**Rationale**: 8 implementation files, 1 major ADR, user-facing `ari provenance` command, central safety gate preventing user content overwrite. GENERATE.

---

## mena-system

| Field | Value |
|-------|-------|
| Name | Mena (Dromena + Legomena) Distribution System |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `internal/mena/` (8 files): type detection, routing logic for `.dro.md`/`.lego.md`
- `internal/materialize/mena/` (8 files): collect engine, namespace, walker, frontmatter
- `mena/` directory: user-level mena (10+ categories)
- `docs/decisions/ADR-0023-dromena-legomena-mena-convention.md`: naming convention
- `docs/decisions/ADR-0025-mena-scope.md`: pipeline-targeted scope filtering
- `docs/decisions/ADR-0021-two-axis-context-model.md`: two-axis context model

**Rationale**: 16 implementation files, 3 ADRs, file-extension-based routing (`.dro.md` -> commands, `.lego.md` -> skills), scope filtering, collision namespace logic. GENERATE.

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
- `internal/agent/` (14 files): scaffold, archetype, templates, frontmatter, mcp_validate, regenerate, sections, validate, adapter (claude + gemini)
- `internal/cmd/agent/` (13 files): summon, dismiss, roster, list, new, embody, update, validate
- `docs/decisions/ADR-0024-agent-factory.md`: structured agent authoring with schema validation

**Rationale**: 27 implementation files, 1 ADR, user-facing `ari agent` command with 8 subcommands, archetype templates, MCP dependency declaration, schema validation, dual-channel adapters. GENERATE.

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

**Rationale**: 19 implementation files, 2 ADRs, 10 subcommands, integration with session lifecycle, hook handlers for `EnterWorktree`, `.knossos/` state seeding. GENERATE.

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

**Rationale**: 20 implementation + schema files, 1 ADR, user-facing `ari validate` command, 11 JSON schemas. GENERATE.

---

## artifact-registry

| Field | Value |
|-------|-------|
| Name | Work Artifact Registry (.ledge/) |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/artifact/` (5 files): aggregate, graduate, query, registry, types
- `internal/ledge/` (4 files): auto_promote, promote
- `internal/cmd/artifact/` (5 files): register, list, query, rebuild
- `internal/cmd/ledge/` (4 files): list, promote, query

**Rationale**: 18 implementation files, user-facing `ari artifact` and `ari ledge` CLIs, phase/type filtering, auto-promotion. GENERATE.

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
- `internal/cmd/sails/check.go`: CLI entry point
- `internal/validation/schemas/white-sails.schema.json`: schema

**Rationale**: 10 implementation files, user-facing `ari sails` command, White/Gray/Black confidence levels, clew contract validation. GENERATE.

---

## clew-handoff-protocol

| Field | Value |
|-------|-------|
| Name | Orchestrator Handoff Protocol |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/hook/clewcontract/`: throughline extraction, orchestrator.go
- `internal/cmd/handoff/` (6 files): execute, handoff, history, prepare, status
- `internal/validation/schemas/handoff-criteria.schema.json`: schema
- `docs/decisions/ADR-0008-handoff-schema-embedding.md`: handoff schema embedding
- `docs/decisions/ADR-0012-cross-rite-protocol-rename.md`: cross-rite protocol

**Rationale**: 8 implementation files, 2 ADRs, `ari handoff` command, handoff criteria schema with machine-verifiable transitions. GENERATE.

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
- `internal/cmd/tribute/tribute.go`, `generate.go`: CLI entry points

**Rationale**: 7 implementation files, user-facing `ari tribute` command, auto-generates session wrap artifacts. GENERATE.

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
- `internal/know/` (8 files): astdiff, know, manifest, qualified, validate, discover
- `internal/cmd/knows/knows.go`, `knows_test.go`: CLI entry point with `--delta` flag
- `.know/` directory: active knowledge domain output

**Rationale**: 10 implementation files, user-facing `ari know` command, AST-based semantic diffing, incremental refresh with `--delta` flag, hook integration for context injection. GENERATE.

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
- `internal/naxos/` (3 files): types, walk, exists
- `internal/cmd/naxos/naxos.go`, `scan.go`, `triage.go`: CLI entry point

**Rationale**: User-facing `ari naxos` command, distinct OrphanReason/SuggestedAction types, session garbage collection. GENERATE.

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
- `internal/cmd/org/` (6 files): init, list, current, set
- `internal/materialize/orgscope/` (2 files): sync
- `internal/config/`: `OrgContext`, `RepoConfig`

**Rationale**: 8 implementation files, user-facing `ari org` CLI with 4 subcommands, distinct XDG-based org directory at tier 3 of 6-tier rite resolution. GENERATE.

---

## procession-system

| Field | Value |
|-------|-------|
| Name | Cross-Rite Procession Workflow System |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.93 |

**Source Evidence**:
- `internal/procession/template.go`: procession template generation
- `internal/cmd/procession/` (6 files): create, proceed, recede, abandon, list, status
- `.ledge/decisions/ADR-0030-processions.md`: dedicated ADR for cross-rite coordinated workflows

**Rationale**: 1 dedicated ADR (ADR-0030), user-facing `ari procession` CLI with 5 subcommands, cross-rite workflow state machine, artifact transfer protocol, resumability. GENERATE.

---

## multi-channel-architecture

| Field | Value |
|-------|-------|
| Name | Multi-Channel Harness Architecture (Claude + Gemini) |
| Category | Core Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `internal/channel/` (2 files): channel directory abstractions
- `internal/agent/adapter_claude.go`, `adapter_gemini.go`, `adapter.go`: dual-channel adapters
- `docs/decisions/ADR-0031-multi-channel-architecture.md`: dedicated ADR
- `docs/decisions/ADR-0032-harness-agnostic-event-vocabulary.md`: vocabulary amendment
- `.gemini/` directory in project root: Gemini channel output

**Rationale**: 2 dedicated ADRs, channel-abstraction layer, dual compiler (Claude/Gemini), harness-agnostic canonical event vocabulary. GENERATE.

---

## clew-slack-bot

| Field | Value |
|-------|-------|
| Name | Clew Organizational Intelligence Slack Bot |
| Category | Clew Intelligence |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `internal/slack/` (7 files): handler, client, renderer, streaming, conversation, types, config
- `internal/cmd/serve/serve.go`, `query.go`: HTTP server startup
- `deploy/` directory with ECS task definition, Terraform, Slack app manifest
- `.github/workflows/deploy-clew.yml`: dedicated CI/CD pipeline

**Rationale**: Entire second subsystem of the binary, 7 implementation files, dedicated deploy pipeline, Docker + ECS + Terraform infrastructure, Slack event handling, HMAC webhook verification. GENERATE.

---

## clew-rag-triage

| Field | Value |
|-------|-------|
| Name | Clew RAG Triage Pipeline |
| Category | Clew Intelligence |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `internal/triage/` (4 files): orchestrator, types, prompts
- `internal/reason/` (5 files): pipeline, context/assembler, intent/classifier, response/generator
- `internal/reason/context/assembler.go`: context assembly with token budgeting
- `internal/reason/intent/classifier.go`: keyword-heuristic query classification

**Rationale**: 10+ implementation files, 4-stage pipeline (intent classification -> context assembly -> Claude reasoning -> streaming response), central Clew intelligence system. GENERATE.

---

## clew-knowledge-index

| Field | Value |
|-------|-------|
| Name | Clew Semantic Knowledge Index |
| Category | Clew Intelligence |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.93 |

**Source Evidence**:
- `internal/search/knowledge/` (6 files): KnowledgeIndex, BM25Searcher, index, persist, types
- `internal/search/knowledge/embedding/store.go`: cosine similarity search
- `internal/search/knowledge/graph/graph.go`: entity relationship graph
- `internal/search/knowledge/summary/store.go`: domain summary store

**Rationale**: 10 implementation files across 4 sub-packages, 3 distinct retrieval strategies (BM25, embedding cosine similarity, entity graph), persistent index. GENERATE.

---

## clew-streaming-response

| Field | Value |
|-------|-------|
| Name | Clew Progressive Streaming Response |
| Category | Clew Intelligence |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/slack/streaming/` (3 files): sender, citations, types
- `internal/slack/streaming/sender.go`: progressive message rendering
- `internal/slack/streaming/citations.go`: citation formatting
- `internal/reason/response/stream.go`: streaming response from Claude API

**Rationale**: 4 implementation files, cross-package streaming pipeline (Claude API -> Slack), citation rendering, progressive update pattern. GENERATE.

---

## clew-conversation-management

| Field | Value |
|-------|-------|
| Name | Clew Thread Context and Conversation History |
| Category | Clew Intelligence |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/slack/conversation/` (4 files): manager, metrics, summarizer, types
- `internal/slack/handler.go`: `ThreadContextStore` usage

**Rationale**: 4 implementation files, TTL eviction policy, conversation summarizer, metrics instrumentation, cross-thread context threading. GENERATE.

---

## clew-org-registry

| Field | Value |
|-------|-------|
| Name | Clew Organization Knowledge Address Space |
| Category | Clew Intelligence |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/registry/org/` (4 files): DomainEntry, DomainCatalog, OrgContext, registry
- `deploy/registry/domains.yaml`: multi-repo domain catalog
- `deploy/registry/org.yaml`: org configuration
- `deploy/content/` directory: per-org/repo knowledge files

**Rationale**: 7 implementation files, multi-repo knowledge catalog with qualified names, sync pipeline, webhook-triggered refresh. GENERATE.

---

## clew-trust-confidence

| Field | Value |
|-------|-------|
| Name | Clew Multi-Axis Confidence Scoring |
| Category | Clew Intelligence |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.85 |

**Source Evidence**:
- `internal/trust/` (7 files): confidence, config, decay, gap, provenance, doc

**Rationale**: 7 implementation files, multi-axis confidence (provenance, gap, decay), distinct tier enum, config-driven thresholds. Cross-cutting input to triage and response generation. GENERATE.

---

## observability-platform

| Field | Value |
|-------|-------|
| Name | CloudWatch EMF Metrics and OTEL Tracing |
| Category | Clew Intelligence |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `internal/observe/` (10 files): cost, logging, metrics, middleware, pipeline, tracer
- `.github/workflows/deploy-clew.yml`: references OTEL_ENDPOINT and CloudWatch integration
- `deploy/terraform/`: CloudWatch logging and alerting infrastructure

**Rationale**: 10 implementation files, two distinct subsystems (CloudWatch EMF + OTEL tracing), cost tracking, HTTP middleware wrapper. GENERATE.

---

## llm-client-abstraction

| Field | Value |
|-------|-------|
| Name | LLM Client Abstraction Layer |
| Category | Clew Intelligence |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.82 |

**Source Evidence**:
- `internal/llm/` (2 files): client, client_test

**Rationale**: 2 files but defines the interface boundary that decouples Clew from Claude-specific API calls. Multiple modules depend on it. GENERATE because it enables the entire Clew intelligence stack.

---

## bm25-search-engine

| Field | Value |
|-------|-------|
| Name | BM25 + RRF Semantic Search Engine (ari ask) |
| Category | User-Facing |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/search/` (10 files): index, collectors, session, entry, score, synonyms
- `internal/search/bm25/` (7 files): build, decay, index, params, scorer, section, store
- `internal/search/fusion/rrf.go`: Reciprocal Rank Fusion
- `internal/cmd/ask/ask.go`: CLI entry point

**Rationale**: 18+ implementation files, user-facing `ari ask` command, 4-tier scoring (exact/prefix/keyword/fuzzy), BM25 with decay, RRF fusion. GENERATE.

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
- `README.md`: documents `ari init`

**Rationale**: User-facing `ari init` command, bootstraps a project without requiring source repo. GENERATE.

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
- `internal/cmd/lint/` (4 files): lint, lint_preferential, models

**Rationale**: User-facing `ari lint` command with distinct detection patterns (source path leaks, broken skill references), severity levels (CRIT/HIGH/MED/LOW). GENERATE.

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
- `internal/cmd/status/status.go`, `status_test.go`: `HealthDashboard` type

**Rationale**: User-facing `ari status` command, unified health dashboard across all managed directories. GENERATE.

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
- `internal/concept/concepts/` (16 concept files): agent, dromena, inscription, knossos, know, ledge, legomena, mena, rite, sails, session, sos, tribute, xenia, potnia, evans-principle

**Rationale**: User-facing `ari explain` command with 16 built-in concept definitions, JSON output support. GENERATE.

---

## knowledge-synthesis-land

| Field | Value |
|-------|-------|
| Name | Knowledge Synthesis Pipeline (ari land) |
| Category | User-Facing |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.83 |

**Source Evidence**:
- `internal/cmd/land/land.go`, `synthesize.go`: CLI implementation

**Rationale**: User-facing `ari land` command, cross-session knowledge consolidation. GENERATE.

---

## complaint-filing-system

| Field | Value |
|-------|-------|
| Name | Cassandra Protocol Complaint Filing |
| Category | User-Facing |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.87 |

**Source Evidence**:
- `internal/cmd/complaint/` (5 files): complaint, dedup, list, update
- `internal/cmd/hook/driftdetect.go`: async hook-path complaint filing
- `.ledge/decisions/ADR-cassandra-dedup-boundary.md`: dedicated ADR

**Rationale**: 5 implementation files, 1 ADR, user-facing `ari complaint` CLI, automated hook path + manual skill path. GENERATE.

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

**Rationale**: 11 implementation files, user-facing `ari manifest` CLI with 4 subcommands, 3-way merge, schema validation. GENERATE.

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
- `embed.go`: `//go:embed` declarations for rites, templates, hooks.yaml, agents, mena, processions
- `docs/decisions/TDD-single-binary-completion.md`: single-binary completion sprint

**Rationale**: 1 TDD, 5+ embedded asset sets, distinct fallback tier in 6-tier resolver. Not a pure utility -- drives distribution model. GENERATE.

---

## resolution-chain

| Field | Value |
|-------|-------|
| Name | Multi-Tier Rite Resolution Chain |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.87 |

**Source Evidence**:
- `internal/resolution/` (3 files): builders, chain, chain_test
- `internal/materialize/source/` (source.go, resolver.go): 6-tier resolver

**Rationale**: 5 implementation files, 6-tier resolution (project-local, org, user, XDG, embedded, fallback), consumed by materialization and rite discovery. Cross-cutting architectural pattern. GENERATE.

---

## ci-cd-infrastructure

| Field | Value |
|-------|-------|
| Name | CI/CD Pipeline Infrastructure |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.80 |

**Source Evidence**:
- `.github/workflows/` (6 files): ariadne-tests, deploy-clew, e2e-distribution, release, validate-orchestrators, verify-doctrine
- `deploy/terraform/` (13 files): ECS cluster, ECR, ALB, IAM, secrets, logging, security groups, alerts
- `deploy/ecs-task-definition.json`: ECS task definition

**Rationale**: 6 workflow files and 13 Terraform files, dedicated Clew deployment pipeline, e2e distribution tests, release pipeline. GENERATE.

---

## rite-library

| Field | Value |
|-------|-------|
| Name | Built-In Rite Library (19 rites) |
| Category | Infrastructure |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.92 |

**Source Evidence**:
- `rites/` directory: 19 rites (10x-dev, arch, clinic, debt-triage, docs, ecosystem, forge, hygiene, intelligence, releaser, review, rnd, security, shared, slop-chop, sre, strategy, thermia, ui)
- `embed.go`: `EmbeddedRites` FS

**Rationale**: 19 embedded rites, each with agents, mena, and workflow definitions. User-facing via `ari rite list`, embedded in binary. GENERATE.

---

## perspective-system

| Field | Value |
|-------|-------|
| Name | Perspective / Viewpoint System |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.72 |

**Source Evidence**:
- `internal/perspective/` (7 files): assemble, audit, context, perspective_test, resolvers, simulate, types

**Rationale**: 7 implementation files exceeds SKIP threshold (5). Provides viewpoint-scoped context assembly. No decision records found. GENERATE conditional -- if classification is wrong, merge into `clew-rag-triage`.

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
- `internal/cmd/tour/` (4 files): collect, models, tour, tour_test

**Rationale**: 4 files, no decision records, read-only directory listing with no cross-cutting concerns. All SKIP conditions met.

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

**Rationale**: 4 files, no decision records, internal utility with no user-facing interface. All SKIP conditions met.

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
- `internal/materialize/skill_policies.go`, `skill_policies_test.go`

**Rationale**: 2 files, no dedicated decision records, no user-facing CLI, pure implementation detail. All SKIP conditions met.

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
- `internal/tokenizer/tokenizer.go`: token estimation

**Rationale**: 3 files, no dedicated decision records, internal hook handler. All SKIP conditions met.

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
- `internal/registry/registry.go`, `validate.go` + tests

**Rationale**: 4 files, no decision records, zero user-facing interface, pure internal utility. All SKIP conditions met.

---

## suggest-system

| Field | Value |
|-------|-------|
| Name | CLI Suggestion Generator |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.75 |

**Source Evidence**:
- `internal/suggest/` (2 files): generator, suggestion
- `internal/cmd/hook/suggest.go`: hook handler

**Rationale**: 2 files, no decision records, no independent user-facing interface. All SKIP conditions met.

---

## checksum-utilities

| Field | Value |
|-------|-------|
| Name | SHA-256 Checksum Utilities |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/checksum/`: checksum utilities

**Rationale**: Pure utility, no decision records, no user-facing interface, fewer than 5 files. All SKIP conditions met.

---

## fileutil-utilities

| Field | Value |
|-------|-------|
| Name | Atomic File Write Utilities |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.90 |

**Source Evidence**:
- `internal/fileutil/`: atomic file write, directory utilities

**Rationale**: Pure utility, no decision records, no user-facing interface, fewer than 5 files. All SKIP conditions met.

---

## Census Gaps

1. **Perspective system scope**: `internal/perspective/` (7 files) implements viewpoint-scoped context assembly and simulation, but no ADR documents its design. Its relationship to `clew-rag-triage` vs standalone feature is ambiguous.

2. **Rite invocation state machine**: `internal/rite/invoker.go`, `state.go` implement `INVOCATION_STATE.yaml` lifecycle. Substantial overlap with `rite-management` -- kept consolidated there.

3. **Cheapo-revert hook**: `internal/cmd/hook/cheapo_revert.go` is a distinct operational tool (force Haiku model via `ElCheapo` flag). Too small for standalone entry, folded into `hook-infrastructure`.

4. **User-scope vs project-scope sync boundary**: `internal/materialize/userscope/` vs rite scope is architecturally significant but treated as part of `materialization-pipeline`.

5. **`internal/sync/`**: 2-file sync state tracker for `.knossos/sync/state.json`. Too small to stand alone; folded into `materialization-pipeline`.

6. **Attribution guard hook**: `internal/cmd/hook/attributionguard.go` handles commit attribution enforcement. Folded into `hook-infrastructure`.

7. **Clew Slack challenge handler**: `internal/serve/webhook/challenge.go` handles Slack URL verification. Part of `clew-slack-bot` feature.

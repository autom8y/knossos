---
domain: feat/tribute-generation
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/tribute/**/*.go"
  - "./internal/cmd/tribute/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.92
format_version: "1.0"
---

# TRIBUTE.md Session Wrap Generation

## Purpose and Design Rationale

Converts ephemeral session artifacts (context, events, sails, registry) into durable machine-parseable summary. Derived not authored (reads facts from events.jsonl, not agent impressions). Dual consumer: human-readable markdown + JSON-serializable counts. Idempotent (overwrite existing). Graceful degradation as first-class principle (missing sources produce valid output with empty sections).

## Conceptual Model

**Sources:** SESSION_CONTEXT.md (required), events.jsonl (optional, tri-format), WHITE_SAILS.yaml (optional), artifact registry (optional, requires ProjectRoot). **Event taxonomy:** tool.artifact_created -> Artifact, agent.decision -> Decision, phase.transitioned -> PhaseRecord (linked list with duration), handoff events -> Handoff, tool.call -> Metrics.ToolCalls, tool.file_change -> Metrics.FilesModified. **Duration:** endedAt - startedAt (archived_at or now). **Artifact type cascade:** event field -> metadata -> path inference.

## Implementation Map

`internal/tribute/` (4 files): types.go (GenerateResult, all data types), extractor.go (per-source extractors), generator.go (Generator.Generate 11-step pipeline), renderer.go (YAML frontmatter + 12 markdown sections). CLI: `internal/cmd/tribute/generate.go` (3 session resolution paths: --session-dir, --session-id, default current).

## Boundaries and Failure Modes

Hard failures: missing session path, non-directory, unreadable SESSION_CONTEXT.md. Graceful: missing events.jsonl (empty data), malformed JSON lines (skipped silently), missing WHITE_SAILS.yaml (nil sails). Schema version hardcoded "1.0" (no migration). Commits data is Phase 2 placeholder. Handoff correlation key collision (same agent pair, last wins). Last phase always has Duration=0. --session-dir bypasses ProjectRoot (graduated artifacts omitted).

## Knowledge Gaps

1. internal/artifact package internals not read
2. session.Context full schema not read
3. Renderer test coverage patterns not read

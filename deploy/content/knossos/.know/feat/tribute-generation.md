---
domain: feat/tribute-generation
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/tribute/**/*.go"
  - "./internal/cmd/tribute/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# TRIBUTE.md Session Wrap Generation

## Purpose and Design Rationale

TRIBUTE.md is a session archive receipt — a single summary document capturing all significant outcomes (decisions, artifacts, phases, handoffs, metrics) from a Knossos session. Generated from SESSION_CONTEXT.md, events.jsonl, WHITE_SAILS.yaml, and git history.

**Key design decisions**: Graceful degradation (only SESSION_CONTEXT.md required; events and sails optional). Schema versioning via YAML frontmatter (`schema_version: "1.0"`). Dual event schema compatibility (pre/post ADR-0027). Git commit extraction deferred to Phase 2. Standalone CLI, NOT auto-integrated into `ari session wrap` (design gap).

**TDD**: `/Users/tomtenuta/Code/knossos/docs/design/TDD-minos-tribute.md`

## Conceptual Model

### Generator Pipeline (8 steps)

1. Validate session path → 2. Load SESSION_CONTEXT.md → 3. Extract events (graceful) → 4. Extract artifacts/decisions/phases/handoffs/metrics → 5. Load WHITE_SAILS.yaml (graceful) → 6. Extract notes → 7. Calculate timing → 8. Render and write TRIBUTE.md

The generator is **read-only** with respect to session state. Idempotent (tested explicitly).

## Implementation Map

### Package Structure

| File | Purpose |
|------|---------|
| `/Users/tomtenuta/Code/knossos/internal/tribute/types.go` | All types (273 lines). `GenerateResult`, `Artifact`, `Decision`, `EventData` with dual-schema accessors |
| `/Users/tomtenuta/Code/knossos/internal/tribute/extractor.go` | Data extraction from session files (456 lines) |
| `/Users/tomtenuta/Code/knossos/internal/tribute/generator.go` | Orchestration (184 lines). `Generate()` entry point |
| `/Users/tomtenuta/Code/knossos/internal/tribute/renderer.go` | Markdown rendering (354 lines). 12 render steps |
| `/Users/tomtenuta/Code/knossos/internal/cmd/tribute/tribute.go` | CLI entry point (57 lines) |
| `/Users/tomtenuta/Code/knossos/internal/cmd/tribute/generate.go` | `ari tribute generate` subcommand (162 lines) |

### Test Coverage

- `extractor_test.go` — unit tests for all extraction functions
- `generator_test.go` — integration tests including idempotency
- `renderer_test.go` — unit tests for conditional rendering

## Boundaries and Failure Modes

- Does NOT integrate with `ari session wrap` automatically (no tribute call in wrap.go)
- Does NOT extract git commits (Phase 2 placeholder)
- Uses `os.WriteFile` instead of `fileutil.WriteIfChanged` (exception to codebase convention)
- Handoff correlation can lose first event if same agent pair hands off twice

## Knowledge Gaps

1. Wrap integration status unclear (TDD specifies it, code doesn't implement it).
2. Git extraction path is fully scaffolded but unimplemented.
3. Concept doc (`explain/concepts/tribute.md`) describes sails, not tribute — factual error.

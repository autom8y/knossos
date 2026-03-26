---
domain: feat/artifact-registry
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/artifact/**/*.go"
  - "./internal/cmd/artifact/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Work Artifact Registry

## Purpose and Design Rationale

Provides a persistent, queryable index of all work artifacts (PRDs, TDDs, ADRs, test plans, runbooks, spikes, reviews) across sessions, with phase and type dimensions for filtering.

**Key decisions**: Federated two-tier architecture (session `artifacts.yaml` + project `registry.yaml`). Graduation at session wrap time, not registration. Code artifacts excluded from graduation. Duplicate detection by `ArtifactID`, not path.

## Conceptual Model

### Two-Level Model

- **Session Ledger** (`.sos/sessions/<id>/artifacts.yaml`) — mutable during session, creation-time paths
- **Project Index** (`.claude/artifacts/registry.yaml`) — aggregated, graduated `.ledge/` paths, pre-computed indexes

### Phase-Type Inference

| Type | Inferred Phase |
|---|---|
| `prd` | `requirements` |
| `tdd`, `adr` | `design` |
| `code` | `implementation` |
| `test-plan` | `validation` |

### `.ledge/` Category Mapping

`adr` → `decisions/`, `prd`/`tdd`/`test-plan`/`runbook` → `specs/`, `review` → `reviews/`, `spike` → `spikes/`, `code` → stays in source tree.

## Implementation Map

Domain: `/Users/tomtenuta/Code/knossos/internal/artifact/` (5 files: `registry.go`, `aggregate.go`, `query.go`, `graduate.go` + tests). CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/artifact/` (5 files).

Key entry points: `Registry.Register()`, `Aggregator.AggregateSession()`, `GraduateSession()`, `Querier.Query()`.

Consumed by `internal/cmd/session/wrap.go:235` (graduation) and `internal/tribute/extractor.go:323` (TRIBUTE.md population).

## Boundaries and Failure Modes

- Does NOT validate artifact file content (the `--skip-validation` flag only sets a boolean)
- `REVIEW` and `SPIKE` type prefixes not detected by `detectArtifact()` in register CLI
- Graduation is NOT fully idempotent (provenance frontmatter can be prepended twice)
- Aggregation is non-transactional (crash mid-aggregate leaves partial state)

## Knowledge Gaps

1. No ADR for artifact registry design.
2. CLI layer (`internal/cmd/artifact/`) has no unit tests.
3. `validation` and `artifact` packages are not integrated.

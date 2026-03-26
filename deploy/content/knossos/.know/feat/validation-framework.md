---
domain: feat/validation-framework
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/validation/**/*.go"
  - "./internal/validation/schemas/*.json"
  - "./internal/cmd/validate/**/*.go"
  - "./docs/decisions/ADR-0008*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.87
format_version: "1.0"
---

# JSON Schema Validation Framework

## Purpose and Design Rationale

Provides machine-verifiable schema enforcement for workflow artifacts (PRDs, TDDs, ADRs, Test Plans) and session state at two gates: structural (JSON Schema Draft 2020-12) and phase-transition (handoff criteria).

**ADR-0008**: Chose `go:embed` for handoff criteria schema. Rejected runtime file loading and code generation. **Dual-schema strategy**: JSON artifacts use `jsonschema/v6` library; YAML handoff criteria use lightweight declarative struct.

## Conceptual Model

### Three Validation Modes

1. **Frontmatter Extraction** (`ExtractFrontmatter`) — parses YAML frontmatter from markdown
2. **JSON Schema Validation** (`Validator`) — structural checks against 11 embedded schemas
3. **Handoff Criteria Validation** (`HandoffValidator`) — phase-transition gate with blocking/non-blocking criteria

### Artifact Type Detection Priority

1. Frontmatter `type` field → 2. Filename pattern regex → 3. `ArtifactTypeUnknown`

### 11 Schema Files

`common`, `session-context`, `prd`, `tdd`, `adr`, `test-plan`, `white-sails`, `agent`, `knossos-manifest`, `handoff-criteria.json`, `handoff-criteria.yaml`

## Implementation Map

9 source files + 4 test files in `/Users/tomtenuta/Code/knossos/internal/validation/`. CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/validate/validate.go` (3 subcommands: artifact, handoff, schema).

Consumed by: `internal/session` (ValidateSessionFields), `internal/sails` (White Sails validation), `internal/agent` (agent frontmatter), `internal/cmd/handoff` (handoff prepare).

## Boundaries and Failure Modes

- Near-leaf package: zero internal domain imports
- Dual `ParseYAMLFrontmatter` implementations (validator.go vs frontmatter.go) — potential divergence
- `isEmpty()` returns false for integer 0 (semantic gap for `non_empty: true`)
- Schema cache is per-Validator instance (no package-level cache)
- `--schema` flag in validate subcommand is non-functional (shadowed variable)

## Knowledge Gaps

1. Five schema files (tdd, adr, test-plan, agent, knossos-manifest) not read.
2. `handoff-criteria.schema.json` exists but has no known Go consumer.
3. White Sails modifier semantics not fully traced.

---
domain: feat/validation-framework
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/validation/**/*.go"
  - "./internal/validation/schemas/**"
  - "./internal/cmd/validate/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.86
format_version: "1.0"
---

# JSON Schema Validation Framework

## Purpose and Design Rationale

Embedded, binary-portable schema validation for all structured artifacts. embedLoader with embed:// URL scheme routes $ref resolution through embedded filesystem. Two-tier validation: schema-structural (JSON Schema) vs handoff-behavioral (YAML-configured phase transition criteria). Lightweight Go fallback functions (ValidateSessionFields, ValidateSailsFields) for hot paths where full compiler is unnecessary.

## Conceptual Model

**Core Validator:** lazy-compiling jsonschema.Compiler with cache. **Common schema** defines shared $defs (timestamps, IDs, enums). **ArtifactValidator:** type detection (frontmatter > filename > unknown) + frontmatter extraction + schema validation. **HandoffValidator:** per-phase per-artifact-type criteria (blocking + warning), loaded from embedded handoff-criteria.yaml. **Sails validation:** WHITE_SAILS.yaml structural correctness (JSON/YAML dual input). **11 schema files** covering session-context, agent, complaint, prd, tdd, adr, test-plan, knossos-manifest, white-sails, handoff-procession, handoff-criteria.

## Implementation Map

`internal/validation/` (6 files): validator.go (Validator, embedLoader, session/agent/complaint methods + standalone functions), frontmatter.go (FrontmatterResult, ExtractFrontmatter), artifact.go (ArtifactValidator, ValidationIssue), handoff.go (HandoffValidator, 3 evaluation modes), sails.go (SailsValidationResult), procession.go (lightweight Go-level validation). CLI: `internal/cmd/validate/validate.go` (artifact, handoff, schema subcommands).

## Boundaries and Failure Modes

Body content not validated (frontmatter only). No cross-artifact referential integrity. knossos-manifest.schema.json is orphaned (no Validator method). Handoff phase-artifact unknown combinations return empty criteria (Passed:true with zero checks). Agent schema validation wraps errors opaquely. ParseYAMLFrontmatter vs ExtractFrontmatter duplication (legacy). ValidateBytes would panic (compiler is nil).

## Knowledge Gaps

1. ADR-0008 not found on disk
2. prd/tdd/adr/test-plan schema contents not individually read
3. internal/cmd/handoff/prepare.go HandoffValidator usage not read

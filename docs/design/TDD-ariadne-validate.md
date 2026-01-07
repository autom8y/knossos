# TDD: Ariadne Validate Domain

> Technical Design Document for the validate domain of the Ariadne Go CLI

**Status**: Draft
**Author**: Architect Agent
**Date**: 2026-01-07
**PRD**: docs/requirements/PRD-ariadne.md

---

## 1. Overview

This Technical Design Document specifies the implementation of the **validate domain** for Ariadne (`ari`), the Go binary for the roster workflow management system. The validate domain encompasses 3 subcommands that validate workflow artifacts, handoff criteria, and configuration schemas to ensure data integrity throughout the development lifecycle.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-ariadne.md` |
| Session TDD | `docs/design/TDD-ariadne-session.md` |
| Manifest TDD | `docs/design/TDD-ariadne-manifest.md` |
| Rite TDD | `docs/design/TDD-ariadne-rite.md` |
| Implementation | `ariadne/internal/cmd/validate/validate.go` |
| Validation Package | `ariadne/internal/validation/` |
| Embedded Schemas | `ariadne/internal/validation/schemas/` |
| Handoff Criteria | `ariadne/internal/validation/schemas/handoff-criteria.yaml` |

### 1.2 Scope

**In Scope**:
- 3 validate subcommands: artifact, handoff, schema
- Artifact type validation (PRD, TDD, ADR, Test Plan)
- Handoff criteria validation for phase transitions
- Schema validation against embedded JSON schemas
- Auto-detection of artifact types from frontmatter and filename patterns
- Blocking and non-blocking criteria evaluation

**Out of Scope**:
- Manifest validation (covered by `ari manifest validate` in manifest domain)
- Session context validation (session domain responsibility)
- Schema authoring and editing (manual process)
- Custom validation rule creation (future enhancement)

### 1.3 Design Goals

1. **Quality Gates**: Provide machine-verifiable validation for workflow artifacts
2. **Phase Transitions**: Ensure handoff criteria are met before phase transitions
3. **Schema Compliance**: Validate frontmatter against JSON schemas
4. **Clear Feedback**: Provide actionable error messages with field-level detail
5. **Workflow Integration**: Support orchestrator and agent validation needs

---

## 2. Architecture

### 2.1 Package Structure

```
ariadne/
├── internal/
│   ├── cmd/
│   │   └── validate/
│   │       └── validate.go         # All validate subcommands
│   └── validation/
│       ├── artifact.go             # Artifact type detection and validation
│       ├── frontmatter.go          # YAML frontmatter extraction
│       ├── handoff.go              # Handoff criteria validation
│       ├── validator.go            # Core schema validation
│       ├── sails.go                # White Sails confidence validation
│       └── schemas/
│           ├── prd.schema.json     # PRD artifact schema
│           ├── tdd.schema.json     # TDD artifact schema
│           ├── adr.schema.json     # ADR artifact schema
│           ├── test-plan.schema.json # Test Plan artifact schema
│           ├── handoff-criteria.yaml # Phase transition criteria
│           └── common.schema.json  # Shared definitions
```

### 2.2 Dependency Graph

```
                    ┌─────────────────────────────────┐
                    │  internal/cmd/validate/         │
                    │  (3 subcommands)                │
                    └─────────────┬───────────────────┘
                                  │
                                  v
                    ┌─────────────────────────────────┐
                    │  internal/validation/           │
                    │  (business logic)               │
                    └─────────────┬───────────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                        │                        │
         v                        v                        v
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ internal/       │     │ internal/output/│     │ internal/errors/│
│ paths/          │     │ (printer)       │     │ (error types)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                  │
                                  v
                    ┌─────────────────────────────────┐
                    │ santhosh-tekuri/jsonschema/v6   │
                    │ gopkg.in/yaml.v3                │
                    └─────────────────────────────────┘
```

### 2.3 External Dependencies

| Purpose | Library | Version | Import Path |
|---------|---------|---------|-------------|
| JSON Schema | santhosh-tekuri/jsonschema | v6+ | `github.com/santhosh-tekuri/jsonschema/v6` |
| YAML | gopkg.in/yaml.v3 | v3 | `gopkg.in/yaml.v3` |

---

## 3. Interface Contracts

### 3.1 Command Summary

| Command | Description | Modifies State |
|---------|-------------|----------------|
| `validate artifact` | Validate artifact against its JSON schema | No |
| `validate handoff` | Validate handoff criteria for phase transitions | No |
| `validate schema` | Validate file against a specific named schema | No |

### 3.2 Command: `ari validate artifact`

Validates a workflow artifact (PRD, TDD, ADR, Test Plan) against its corresponding JSON schema by parsing YAML frontmatter.

**Signature**:
```
ari validate artifact [file] [--type=TYPE]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `file` | Yes | Path to artifact file to validate |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | (auto-detect) | Artifact type: prd, tdd, adr, test-plan |

**Artifact Type Auto-Detection**:

The artifact type is determined by (in priority order):

1. **Frontmatter `type` field**: If present, takes precedence
2. **Filename pattern**: Based on filename conventions
3. **`--type` flag**: Explicit override

| Filename Pattern | Detected Type |
|------------------|---------------|
| `PRD-*.md` | prd |
| `TDD-*.md` | tdd |
| `ADR-[0-9]+*.md` | adr |
| `TEST-*.md`, `TP-*.md` | test-plan |

**Output (JSON, valid)**:
```json
{
  "valid": true,
  "artifact_type": "prd",
  "file_path": "/path/to/PRD-user-auth.md",
  "frontmatter": {
    "artifact_id": "PRD-user-auth",
    "title": "User Authentication",
    "status": "approved",
    "success_criteria": [...]
  }
}
```

**Output (JSON, invalid)**:
```json
{
  "valid": false,
  "artifact_type": "prd",
  "file_path": "/path/to/PRD-user-auth.md",
  "issues": [
    {
      "field": "artifact_id",
      "message": "missing required field"
    },
    {
      "field": "success_criteria",
      "message": "must have at least 1 item"
    }
  ],
  "frontmatter": {
    "title": "User Authentication"
  }
}
```

**Output (text, valid)**:
```
VALID: /path/to/PRD-user-auth.md
  Type: prd
```

**Output (text, invalid)**:
```
INVALID: /path/to/PRD-user-auth.md
  Type: prd
  Issues:
    - [artifact_id] missing required field
    - [success_criteria] must have at least 1 item
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Artifact validation passed |
| 2 | Invalid artifact type specified |
| 4 | Schema validation failed (SCHEMA_INVALID) |
| 6 | File not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Extracts YAML frontmatter between `---` delimiters
- Converts frontmatter to JSON for schema validation
- Auto-detects artifact type from filename if not specified
- Returns validation issues with field paths for precise error location

### 3.3 Command: `ari validate handoff`

Validates that artifacts meet handoff criteria for transitioning between workflow phases. Supports listing phases and showing criteria.

**Signature**:
```
ari validate handoff [--phase=PHASE] [--artifact=PATH] [--type=TYPE] [--list-phases] [--show-criteria]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--phase` | | string | | Workflow phase: requirements, design, implementation, validation |
| `--artifact` | | string | | Path to artifact file to validate |
| `--type` | | string | (auto-detect) | Artifact type for --show-criteria |
| `--list-phases` | | bool | false | List all phases with handoff criteria |
| `--show-criteria` | | bool | false | Show criteria for a specific phase/type |

**Workflow Phases**:

| Phase | Primary Artifact | Purpose |
|-------|------------------|---------|
| `requirements` | PRD | Requirements analysis complete, ready for design |
| `design` | TDD, ADR | Technical design complete, ready for implementation |
| `implementation` | TDD | Implementation complete, ready for validation |
| `validation` | Test Plan | Validation complete, ready for delivery |

**Operation Modes**:

1. **List Phases** (`--list-phases`): Show all phases with defined criteria
2. **Show Criteria** (`--phase` + `--type` + `--show-criteria`): Show criteria details
3. **Validate Handoff** (`--phase` + `--artifact`): Validate artifact against criteria

**Output (JSON, --list-phases)**:
```json
{
  "phases": [
    {
      "phase": "requirements",
      "artifact_types": ["prd"]
    },
    {
      "phase": "design",
      "artifact_types": ["tdd", "adr"]
    },
    {
      "phase": "implementation",
      "artifact_types": ["tdd"]
    },
    {
      "phase": "validation",
      "artifact_types": ["test-plan"]
    }
  ]
}
```

**Output (text, --list-phases)**:
```
Phases with handoff criteria:
  requirements:
    - prd
  design:
    - tdd
    - adr
  implementation:
    - tdd
  validation:
    - test-plan
```

**Output (JSON, --show-criteria)**:
```json
{
  "phase": "requirements",
  "artifact_type": "prd",
  "blocking": [
    {
      "field": "artifact_id",
      "message": "PRD must have artifact_id",
      "non_empty": true
    },
    {
      "field": "title",
      "message": "PRD must have title",
      "non_empty": true
    },
    {
      "field": "status",
      "message": "PRD must have status",
      "non_empty": true
    },
    {
      "field": "success_criteria",
      "message": "PRD must define success criteria",
      "min_items": 1
    }
  ],
  "non_blocking": [
    {
      "field": "stakeholders",
      "message": "PRD should list stakeholders"
    },
    {
      "field": "complexity",
      "message": "PRD should specify complexity level"
    }
  ]
}
```

**Output (text, --show-criteria)**:
```
Handoff criteria for requirements/prd:
  Blocking:
    - artifact_id: PRD must have artifact_id
    - title: PRD must have title
    - status: PRD must have status
    - success_criteria: PRD must define success criteria
  Non-blocking:
    - stakeholders: PRD should list stakeholders
    - complexity: PRD should specify complexity level
```

**Output (JSON, validation passed)**:
```json
{
  "passed": true,
  "phase": "requirements",
  "artifact_type": "prd",
  "file_path": "/path/to/PRD-user-auth.md"
}
```

**Output (JSON, validation failed)**:
```json
{
  "passed": false,
  "phase": "requirements",
  "artifact_type": "prd",
  "file_path": "/path/to/PRD-user-auth.md",
  "blocking_failed": [
    {
      "criterion": {
        "field": "success_criteria",
        "message": "PRD must define success criteria",
        "min_items": 1
      },
      "passed": false,
      "message": "PRD must define success criteria (has 0, needs 1)"
    }
  ],
  "warnings": [
    {
      "criterion": {
        "field": "stakeholders",
        "message": "PRD should list stakeholders"
      },
      "passed": false,
      "message": "PRD should list stakeholders"
    }
  ]
}
```

**Output (text, validation passed)**:
```
PASSED: requirements handoff for prd
  File: /path/to/PRD-user-auth.md
```

**Output (text, validation failed)**:
```
FAILED: requirements handoff for prd
  File: /path/to/PRD-user-auth.md
  Blocking failures:
    - [success_criteria] PRD must define success criteria (has 0, needs 1)
  Warnings:
    - [stakeholders] PRD should list stakeholders
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Handoff validation passed (or list/show operations succeeded) |
| 2 | Invalid phase or artifact type specified |
| 4 | Handoff validation failed (blocking criteria not met) |
| 6 | Artifact file not found |
| 9 | No .claude/ directory found |

**Handoff Criteria Definitions**:

Criteria are defined in `schemas/handoff-criteria.yaml` and embedded in the binary. Each criterion has:

| Field | Type | Description |
|-------|------|-------------|
| `field` | string | Frontmatter field to validate |
| `message` | string | Error message when criterion fails |
| `non_empty` | bool | Field must be non-empty (not just present) |
| `min_items` | int | Minimum array items required |

### 3.4 Command: `ari validate schema`

Validates a file's YAML frontmatter against a specific named schema. This provides more control than `artifact` by allowing explicit schema selection.

**Signature**:
```
ari validate schema <schema-name> <file>
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `schema-name` | Yes | Schema to validate against: prd, tdd, adr, test-plan |
| `file` | Yes | Path to file to validate |

**Available Schemas**:

| Schema Name | File | Purpose |
|-------------|------|---------|
| `prd` | `prd.schema.json` | Product Requirements Document |
| `tdd` | `tdd.schema.json` | Technical Design Document |
| `adr` | `adr.schema.json` | Architecture Decision Record |
| `test-plan` | `test-plan.schema.json` | Test Plan |

**Output (JSON, valid)**:
```json
{
  "valid": true,
  "schema_name": "prd",
  "file_path": "/path/to/requirements.md"
}
```

**Output (JSON, invalid)**:
```json
{
  "valid": false,
  "schema_name": "prd",
  "file_path": "/path/to/requirements.md",
  "issues": [
    {
      "field": "artifact_id",
      "message": "missing required field"
    }
  ]
}
```

**Output (text, valid)**:
```
VALID: /path/to/requirements.md (schema: prd)
```

**Output (text, invalid)**:
```
INVALID: /path/to/requirements.md (schema: prd)
  Issues:
    - [artifact_id] missing required field
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Schema validation passed |
| 4 | Schema validation failed (SCHEMA_INVALID) |
| 6 | File not found |
| 14 | Schema not found (SCHEMA_NOT_FOUND) |

---

## 4. Error Handling

### 4.1 Error Code Taxonomy

| Code | Exit | Name | Description |
|------|------|------|-------------|
| `SUCCESS` | 0 | Success | Validation passed |
| `GENERAL_ERROR` | 1 | General Error | Unspecified error |
| `USAGE_ERROR` | 2 | Usage Error | Invalid arguments or flags |
| `SCHEMA_INVALID` | 4 | Schema Invalid | Artifact/handoff validation failed |
| `FILE_NOT_FOUND` | 6 | File Not Found | Artifact file missing |
| `PROJECT_NOT_FOUND` | 9 | Project Not Found | No .claude/ directory found |
| `SCHEMA_NOT_FOUND` | 14 | Schema Not Found | Specified schema not available |

### 4.2 Error Response Structure

All errors follow the standard Ariadne error contract:

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

Example (invalid artifact type):
```json
{
  "error": {
    "code": "USAGE_ERROR",
    "message": "invalid artifact type",
    "details": {
      "type": "invalid",
      "valid": ["prd", "tdd", "adr", "test-plan"]
    }
  }
}
```

Example (schema not found):
```json
{
  "error": {
    "code": "SCHEMA_NOT_FOUND",
    "message": "unknown schema",
    "details": {
      "schema": "invalid",
      "valid": ["prd", "tdd", "adr", "test-plan"]
    }
  }
}
```

---

## 5. Data Model

### 5.1 Artifact Types

```go
type ArtifactType string

const (
    ArtifactTypePRD      ArtifactType = "prd"
    ArtifactTypeTDD      ArtifactType = "tdd"
    ArtifactTypeADR      ArtifactType = "adr"
    ArtifactTypeTestPlan ArtifactType = "test-plan"
    ArtifactTypeUnknown  ArtifactType = ""
)
```

### 5.2 Workflow Phases

```go
type Phase string

const (
    PhaseRequirements   Phase = "requirements"
    PhaseDesign         Phase = "design"
    PhaseImplementation Phase = "implementation"
    PhaseValidation     Phase = "validation"
)
```

### 5.3 Validation Issue

```go
type ValidationIssue struct {
    Field   string      `json:"field,omitempty"`
    Message string      `json:"message"`
    Value   interface{} `json:"value,omitempty"`
}
```

### 5.4 Artifact Validation Result

```go
type ArtifactValidationResult struct {
    Valid        bool                   `json:"valid"`
    ArtifactType ArtifactType           `json:"artifact_type"`
    FilePath     string                 `json:"file_path"`
    Issues       []ValidationIssue      `json:"issues,omitempty"`
    Frontmatter  map[string]interface{} `json:"frontmatter,omitempty"`
}
```

### 5.5 Handoff Criteria

```go
type Criterion struct {
    Field    string `yaml:"field" json:"field"`
    Message  string `yaml:"message" json:"message"`
    NonEmpty bool   `yaml:"non_empty" json:"non_empty,omitempty"`
    MinItems *int   `yaml:"min_items" json:"min_items,omitempty"`
}

type ArtifactCriteria struct {
    Blocking    []Criterion `yaml:"blocking" json:"blocking,omitempty"`
    NonBlocking []Criterion `yaml:"non_blocking" json:"non_blocking,omitempty"`
}
```

### 5.6 Handoff Validation Result

```go
type HandoffResult struct {
    Passed          bool                   `json:"passed"`
    Phase           Phase                  `json:"phase"`
    ArtifactType    ArtifactType           `json:"artifact_type"`
    FilePath        string                 `json:"file_path,omitempty"`
    BlockingResults []CriterionResult      `json:"blocking_results,omitempty"`
    WarningResults  []CriterionResult      `json:"warning_results,omitempty"`
    Frontmatter     map[string]interface{} `json:"frontmatter,omitempty"`
}

type CriterionResult struct {
    Criterion Criterion   `json:"criterion"`
    Passed    bool        `json:"passed"`
    Message   string      `json:"message,omitempty"`
    Value     interface{} `json:"value,omitempty"`
}
```

---

## 6. Internal Package Design

### 6.1 Package: `internal/validation/artifact.go`

Artifact type detection and schema-based validation.

```go
// DetectArtifactType determines artifact type from filename and/or frontmatter.
// Priority: frontmatter "type" > filename pattern > explicit --type flag
func DetectArtifactType(filename string, frontmatter map[string]interface{}) ArtifactType

// ArtifactValidator validates workflow artifacts against schemas.
type ArtifactValidator struct {
    validator *Validator
}

func NewArtifactValidator() (*ArtifactValidator, error)

// ValidateFile validates an artifact file against its schema.
func (av *ArtifactValidator) ValidateFile(filePath string, artifactType ArtifactType) (*ArtifactValidationResult, error)

// Validate validates artifact content against its schema.
func (av *ArtifactValidator) Validate(content []byte, filePath string, artifactType ArtifactType) (*ArtifactValidationResult, error)
```

### 6.2 Package: `internal/validation/handoff.go`

Handoff criteria loading and evaluation.

```go
// HandoffValidator validates artifacts against handoff criteria.
type HandoffValidator struct {
    criteria HandoffCriteria // Embedded from schemas/handoff-criteria.yaml
}

func NewHandoffValidator() (*HandoffValidator, error)

// GetCriteria returns criteria for a phase and artifact type.
func (hv *HandoffValidator) GetCriteria(phase Phase, artifactType ArtifactType) (*ArtifactCriteria, error)

// ListPhases returns all phases that have criteria defined.
func (hv *HandoffValidator) ListPhases() []Phase

// ListArtifactTypes returns artifact types with criteria for a phase.
func (hv *HandoffValidator) ListArtifactTypes(phase Phase) []ArtifactType

// ValidateHandoff validates frontmatter against handoff criteria.
func (hv *HandoffValidator) ValidateHandoff(phase Phase, artifactType ArtifactType, frontmatter map[string]interface{}) (*HandoffResult, error)

// ValidateHandoffFile validates an artifact file against handoff criteria.
func (hv *HandoffValidator) ValidateHandoffFile(phase Phase, filePath string) (*HandoffResult, error)
```

### 6.3 Package: `internal/validation/validator.go`

Core schema validation using embedded JSON schemas.

```go
// Validator provides schema validation capabilities.
type Validator struct {
    compiler *jsonschema.Compiler
    schemas  map[string]*jsonschema.Schema
}

func NewValidator() (*Validator, error)

// getSchema returns a compiled schema, caching the result.
func (v *Validator) getSchema(name string) (*jsonschema.Schema, error)
```

### 6.4 Package: `internal/validation/frontmatter.go`

YAML frontmatter extraction from markdown files.

```go
// Frontmatter represents extracted YAML frontmatter.
type Frontmatter struct {
    Data map[string]interface{}
    Raw  []byte
}

// ExtractFrontmatter extracts YAML frontmatter from markdown content.
func ExtractFrontmatter(content []byte) (*Frontmatter, error)
```

---

## 7. Handoff Criteria Definitions

The handoff criteria are defined in `internal/validation/schemas/handoff-criteria.yaml`:

### 7.1 Requirements Phase (PRD)

**Blocking Criteria**:
| Field | Requirement | Message |
|-------|-------------|---------|
| `artifact_id` | non-empty | PRD must have artifact_id |
| `title` | non-empty | PRD must have title |
| `status` | non-empty | PRD must have status |
| `success_criteria` | min_items: 1 | PRD must define success criteria |

**Non-Blocking Criteria**:
| Field | Message |
|-------|---------|
| `stakeholders` | PRD should list stakeholders |
| `complexity` | PRD should specify complexity level |

### 7.2 Design Phase (TDD)

**Blocking Criteria**:
| Field | Requirement | Message |
|-------|-------------|---------|
| `artifact_id` | non-empty | TDD must have artifact_id |
| `title` | non-empty | TDD must have title |
| `status` | non-empty | TDD must have status |
| `prd_ref` | non-empty | TDD must reference a PRD |
| `implementation_plan` | present | TDD must have implementation plan |

**Non-Blocking Criteria**:
| Field | Message |
|-------|---------|
| `components` | TDD should list components |
| `dependencies` | TDD should document dependencies |

### 7.3 Design Phase (ADR)

**Blocking Criteria**:
| Field | Requirement | Message |
|-------|-------------|---------|
| `artifact_id` | non-empty | ADR must have artifact_id |
| `title` | non-empty | ADR must have title |
| `status` | non-empty | ADR must have status |
| `date` | non-empty | ADR must have date |

**Non-Blocking Criteria**:
| Field | Message |
|-------|---------|
| `context` | ADR should document context |
| `consequences` | ADR should list consequences |

### 7.4 Implementation Phase (TDD)

**Blocking Criteria**:
| Field | Requirement | Message |
|-------|-------------|---------|
| `artifact_id` | non-empty | TDD must have artifact_id |
| `status` | non-empty | TDD must have approved status for implementation |

**Non-Blocking Criteria**:
| Field | Message |
|-------|---------|
| `test_coverage` | TDD should specify test coverage requirements |

### 7.5 Validation Phase (Test Plan)

**Blocking Criteria**:
| Field | Requirement | Message |
|-------|-------------|---------|
| `artifact_id` | non-empty | Test plan must have artifact_id |
| `title` | non-empty | Test plan must have title |
| `status` | non-empty | Test plan must have status |

**Non-Blocking Criteria**:
| Field | Message |
|-------|---------|
| `test_cases` | Test plan should define test cases (min: 1) |
| `coverage_targets` | Test plan should specify coverage targets |

---

## 8. Embedded Schemas

### 8.1 Schema Embedding

Schemas are embedded using Go's `//go:embed` directive:

```go
//go:embed schemas/*.json
var schemaFS embed.FS

//go:embed schemas/handoff-criteria.yaml
var handoffCriteriaFS embed.FS
```

### 8.2 Available Schemas

| File | Purpose |
|------|---------|
| `prd.schema.json` | PRD frontmatter validation |
| `tdd.schema.json` | TDD frontmatter validation |
| `adr.schema.json` | ADR frontmatter validation |
| `test-plan.schema.json` | Test Plan frontmatter validation |
| `common.schema.json` | Shared definitions |
| `handoff-criteria.yaml` | Phase transition criteria |
| `session-context.schema.json` | Session context validation |
| `white-sails.schema.json` | Confidence signal validation |

---

## 9. Integration Points

### 9.1 Orchestrator Integration

The orchestrator uses validate commands to enforce quality gates:

```bash
# Before transitioning from requirements to design
ari validate handoff --phase=requirements --artifact=docs/requirements/PRD-user-auth.md

# Before implementation
ari validate handoff --phase=design --artifact=docs/design/TDD-user-auth.md
```

### 9.2 Agent Integration

Agents can validate their output artifacts:

```bash
# Requirements Analyst validates PRD
ari validate artifact docs/requirements/PRD-user-auth.md

# Architect validates TDD and ADRs
ari validate artifact docs/design/TDD-user-auth.md
ari validate artifact docs/decisions/ADR-0001.md
```

### 9.3 CI/CD Integration

Validation can be integrated into CI pipelines:

```yaml
# GitHub Actions example
- name: Validate artifacts
  run: |
    ari validate artifact docs/requirements/PRD-*.md
    ari validate artifact docs/design/TDD-*.md
```

### 9.4 Session Transitions

Session phase transitions can require handoff validation:

```bash
# Validate before phase transition
if ari validate handoff --phase=requirements --artifact=$PRD_PATH; then
    ari session transition design
fi
```

---

## 10. Test Strategy

### 10.1 Unit Tests

Location: `internal/validation/*_test.go`

| Package | Test Focus | Coverage Target |
|---------|-----------|-----------------|
| `artifact` | Type detection, schema validation | 100% |
| `frontmatter` | YAML extraction, edge cases | 100% |
| `handoff` | Criteria loading, evaluation | 100% |
| `validator` | Schema compilation, validation | 100% |

### 10.2 Integration Tests

| Test ID | Description |
|---------|-------------|
| `validate_001` | Artifact validation passes for valid PRD |
| `validate_002` | Artifact validation fails with clear field errors |
| `validate_003` | Auto-detects type from filename pattern |
| `validate_004` | Auto-detects type from frontmatter |
| `validate_005` | Handoff validation passes for complete artifact |
| `validate_006` | Handoff validation fails on missing blocking criteria |
| `validate_007` | Handoff warnings for missing non-blocking criteria |
| `validate_008` | List phases returns all defined phases |
| `validate_009` | Show criteria returns blocking and non-blocking |
| `validate_010` | Schema validation with explicit schema name |
| `validate_011` | Unknown schema returns SCHEMA_NOT_FOUND |
| `validate_012` | Missing file returns FILE_NOT_FOUND |

### 10.3 Test Fixtures

```
ariadne/
└── testdata/
    └── artifacts/
        ├── valid/
        │   ├── PRD-test.md           # Valid PRD
        │   ├── TDD-test.md           # Valid TDD
        │   ├── ADR-0001.md           # Valid ADR
        │   └── TEST-test.md          # Valid Test Plan
        ├── invalid/
        │   ├── PRD-missing-id.md     # Missing artifact_id
        │   ├── PRD-no-criteria.md    # Missing success_criteria
        │   └── TDD-no-prd-ref.md     # Missing prd_ref
        └── edge-cases/
            ├── no-frontmatter.md     # No YAML frontmatter
            ├── empty-frontmatter.md  # Empty frontmatter
            └── malformed.md          # Invalid YAML
```

---

## 11. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Schema evolution breaks existing artifacts | Medium | High | Version schemas, provide migration guidance |
| Complex frontmatter structures | Low | Medium | Test with nested structures, clear error paths |
| Performance on large files | Low | Low | Only parse frontmatter section |
| Missing schema for custom artifact types | Medium | Low | Clear error message with valid types |

---

## 12. Handoff Criteria

Ready for Implementation when:

- [x] All 3 validate subcommands have interface contracts
- [x] Artifact types and detection logic defined
- [x] Workflow phases and handoff criteria documented
- [x] Error codes mapped to exit codes
- [x] Output structures (JSON/text) specified
- [x] Integration points documented
- [x] Test scenarios cover critical paths
- [ ] Principal Engineer can implement without architectural questions
- [ ] All artifacts verified via Read tool

---

## 13. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-validate.md` | Write |
| Implementation | `/Users/tomtenuta/Code/roster/internal/cmd/validate/validate.go` | Read |
| Artifact Validation | `/Users/tomtenuta/Code/roster/internal/validation/artifact.go` | Read |
| Handoff Validation | `/Users/tomtenuta/Code/roster/internal/validation/handoff.go` | Read |
| Core Validator | `/Users/tomtenuta/Code/roster/internal/validation/validator.go` | Read |
| Embedded Criteria | `/Users/tomtenuta/Code/roster/internal/validation/schemas/handoff-criteria.yaml` | Read |
| Error Codes | `/Users/tomtenuta/Code/roster/internal/errors/errors.go` | Read |
| Reference TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-session.md` | Read |
| Reference TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-manifest.md` | Read |

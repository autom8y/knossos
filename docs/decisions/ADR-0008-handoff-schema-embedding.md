# ADR-0008: Handoff Schema Embedding Strategy

| Field | Value |
|-------|-------|
| **Status** | Proposed |
| **Date** | 2026-01-04 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The `handoff-criteria-schema.yaml` defines machine-verifiable criteria for phase transitions in workflows. This schema enables orchestrators to validate artifacts before handoffs between agents. Ariadne needs to access this schema for the `ari validate handoff` command.

### Current State

- Schema location: `schemas/handoff-criteria-schema.yaml` (441 lines)
- Schema format: YAML with custom validation DSL
- Current consumers: bash scripts via `yq` parsing
- Future consumers: Ariadne CLI (`ari validate handoff`)

### Existing Pattern in Ariadne

Ariadne already embeds JSON schemas for session context validation:

```go
// ariadne/internal/validation/validator.go
//go:embed schemas/*.json
var schemaFS embed.FS

func NewValidator() (*Validator, error) {
    compiler := jsonschema.NewCompiler()
    entries, err := schemaFS.ReadDir("schemas")
    // ...
}
```

This pattern works well for JSON schemas but the handoff criteria schema is YAML with a custom validation DSL that doesn't map directly to JSON Schema.

### Constraints

- **Single Binary**: Ariadne must remain a single executable without external dependencies
- **Schema Updates**: Schema changes should not require recompiling Ariadne in all cases
- **Performance**: Validation runs during handoffs; must be fast
- **Consistency**: Same schema used by bash scripts and Go CLI
- **Custom DSL**: The schema contains validation expressions like `artifact_id matches pattern ^PRD-[a-z0-9-]+$` that need interpretation

## Decision

Use **`go:embed` for compile-time embedding** with a YAML parser and custom validation interpreter.

### Option Analysis

#### Option 1: `go:embed` Directive (Selected)

```go
// ariadne/internal/validation/handoff.go
package validation

import (
    _ "embed"
    "gopkg.in/yaml.v3"
)

//go:embed schemas/handoff-criteria-schema.yaml
var handoffSchemaYAML []byte

type HandoffSchema struct {
    SchemaVersion string                    `yaml:"schema_version"`
    ArtifactTypes map[string]ArtifactType   `yaml:"artifact_types"`
    Validation    ValidationConfig          `yaml:"validation"`
    Protocol      ProtocolConfig            `yaml:"protocol"`
}

type ArtifactType struct {
    Phase       string     `yaml:"phase"`
    Description string     `yaml:"description"`
    Criteria    []Criterion `yaml:"criteria"`
}

type Criterion struct {
    ID          string `yaml:"id"`
    Description string `yaml:"description"`
    Validation  string `yaml:"validation"`
    Blocking    bool   `yaml:"blocking"`
}
```

**Pros**:
- Zero external dependencies at runtime
- Compile-time verification that schema exists
- Consistent with existing session-context schema pattern
- Single binary deployment
- Fast: YAML parsed once at startup, cached

**Cons**:
- Schema updates require recompilation
- YAML embedded in binary increases binary size (~15KB)
- Custom validation DSL requires interpreter implementation

#### Option 2: Runtime File Loading with Fallback

```go
func LoadHandoffSchema() (*HandoffSchema, error) {
    // Try runtime path first
    paths := []string{
        filepath.Join(rosterHome, "schemas/handoff-criteria-schema.yaml"),
        "/etc/ariadne/handoff-criteria-schema.yaml",
    }

    for _, p := range paths {
        if data, err := os.ReadFile(p); err == nil {
            return parseSchema(data)
        }
    }

    // Fall back to embedded
    return parseSchema(embeddedSchemaYAML)
}
```

**Pros**:
- Schema updates without recompilation
- Useful for development/testing
- Can override embedded schema per-project

**Cons**:
- Runtime dependency on file system
- Potential for schema version drift between Ariadne and roster
- More complex error handling
- Security: arbitrary files could be loaded

#### Option 3: Generated Go Code from YAML

```bash
# Build-time generation
go generate ./...

# Generates: ariadne/internal/validation/handoff_generated.go
```

```go
//go:generate go run gen/schema_gen.go -input=schemas/handoff-criteria-schema.yaml -output=handoff_generated.go

var HandoffCriteria = map[string]ArtifactType{
    "prd": {
        Phase: "requirements",
        Description: "Product Requirements Document",
        Criteria: []Criterion{
            {ID: "prd-001", Description: "PRD has valid artifact_id", ...},
        },
    },
    // ... generated from YAML
}
```

**Pros**:
- Type safety at compile time
- IDE autocompletion for criteria IDs
- Faster runtime (no YAML parsing)
- Validation errors caught during `go generate`

**Cons**:
- Build complexity (requires generator)
- Generated code is harder to debug
- Changes to YAML structure require generator updates
- Overkill for current use case

### Selected Approach: Option 1 (`go:embed`)

This matches the existing pattern for session-context.schema.json and provides the right balance of simplicity and reliability.

### Implementation Design

#### Schema Location

```
ariadne/
  internal/
    validation/
      schemas/
        session-context.schema.json  # existing
        handoff-criteria-schema.yaml  # new (copied from schemas/)
      validator.go        # existing JSON schema validator
      handoff.go          # new handoff criteria validator
```

The schema is copied to the `schemas/` directory within the validation package at build time or as part of the repo structure.

#### Validation Interpreter

The custom DSL expressions like `artifact_id matches pattern ^PRD-[a-z0-9-]+$` require an interpreter:

```go
// ariadne/internal/validation/handoff.go
package validation

type HandoffValidator struct {
    schema *HandoffSchema
}

func NewHandoffValidator() (*HandoffValidator, error) {
    var schema HandoffSchema
    if err := yaml.Unmarshal(handoffSchemaYAML, &schema); err != nil {
        return nil, fmt.Errorf("failed to parse handoff schema: %w", err)
    }
    return &HandoffValidator{schema: &schema}, nil
}

// ValidateArtifact validates an artifact against handoff criteria
func (v *HandoffValidator) ValidateArtifact(artifactType string, data map[string]interface{}) (*ValidationResult, error) {
    artType, ok := v.schema.ArtifactTypes[artifactType]
    if !ok {
        return nil, fmt.Errorf("unknown artifact type: %s", artifactType)
    }

    result := &ValidationResult{
        ArtifactType: artifactType,
        Phase:        artType.Phase,
    }

    for _, criterion := range artType.Criteria {
        passed, err := v.evaluateCriterion(criterion, data)
        if err != nil {
            result.Errors = append(result.Errors, CriterionError{
                ID:    criterion.ID,
                Error: err.Error(),
            })
            continue
        }

        if !passed {
            if criterion.Blocking {
                result.BlockingFailures = append(result.BlockingFailures, criterion.ID)
            } else {
                result.Warnings = append(result.Warnings, criterion.ID)
            }
        } else {
            result.Passed = append(result.Passed, criterion.ID)
        }
    }

    result.Valid = len(result.BlockingFailures) == 0
    return result, nil
}

// evaluateCriterion interprets the validation DSL
func (v *HandoffValidator) evaluateCriterion(c Criterion, data map[string]interface{}) (bool, error) {
    // Parse and evaluate expressions like:
    // - "artifact_id matches pattern ^PRD-[a-z0-9-]+$"
    // - "count(success_criteria) >= 1"
    // - "status == 'approved'"
    // - "all(success_criteria.testable == true)"
    return evaluateDSL(c.Validation, data)
}
```

#### DSL Expression Types

| Expression Pattern | Example | Interpretation |
|--------------------|---------|----------------|
| `field matches pattern regex` | `artifact_id matches pattern ^PRD-[a-z0-9-]+$` | Regex match |
| `count(field) op N` | `count(success_criteria) >= 1` | Array length comparison |
| `field == 'value'` | `status == 'approved'` | Equality check |
| `field in [...]` | `complexity in ['SCRIPT', 'MODULE']` | Membership test |
| `field is not null` | `prd_ref is not null` | Null check |
| `all(field.subfield op value)` | `all(success_criteria.testable == true)` | Universal quantifier |
| `condition implies other` | `status == 'superseded' implies superseded_by is not null` | Logical implication |

### Sync Strategy

To keep the embedded schema in sync with the canonical source:

```makefile
# Makefile
.PHONY: sync-schemas
sync-schemas:
	cp schemas/handoff-criteria-schema.yaml ariadne/internal/validation/schemas/

.PHONY: build
build: sync-schemas
	go build ./ariadne/...
```

Or use a symlink in development:

```bash
cd ariadne/internal/validation/schemas
ln -s ../../../../schemas/handoff-criteria-schema.yaml .
```

### CLI Integration

```bash
# Validate a specific artifact
ari validate handoff --type=prd --file=docs/requirements/PRD-user-auth.md

# Output (JSON mode):
{
  "artifact_type": "prd",
  "phase": "requirements",
  "valid": true,
  "passed": ["prd-001", "prd-002", "prd-004", "prd-005"],
  "warnings": ["prd-003"],
  "blocking_failures": []
}

# List all artifact types
ari validate handoff --list-types

# Show criteria for a type
ari validate handoff --type=prd --show-criteria
```

## Consequences

### Positive

1. **Consistency**: Same embedding pattern as session-context schema
2. **Reliability**: No runtime file system dependencies
3. **Single Binary**: Maintains Ariadne's deployment simplicity
4. **Type Safety**: Go structs catch schema structure errors at compile time
5. **Performance**: YAML parsed once, criteria cached in memory

### Negative

1. **Build Complexity**: Schema must be synced before compilation
2. **Staleness Risk**: Embedded schema can drift from canonical source
3. **DSL Interpreter**: Custom validation DSL requires implementation effort (~200 LOC)
4. **Binary Size**: Adds ~15KB to Ariadne binary

### Neutral

1. **YAML vs JSON**: Handoff schema remains YAML (human-readable) while session schema is JSON (standard tooling)
2. **Interpreter Maintenance**: DSL is stable; interpreter unlikely to need frequent updates

## Implementation Plan

1. **Copy schema** to `ariadne/internal/validation/schemas/`
2. **Create handoff.go** with schema structs and validator
3. **Implement DSL interpreter** for validation expressions
4. **Add validate subcommand** to Ariadne CLI
5. **Add Makefile target** for schema sync
6. **Write tests** for each DSL expression type

## Related Decisions

- **ADR-0001**: Session State Machine (validation pattern precedent)
- **ADR-0007**: Team Context YAML Architecture (YAML in Go pattern)

## References

- Existing embed pattern: `ariadne/internal/validation/validator.go`
- Handoff schema: `schemas/handoff-criteria-schema.yaml`
- Go embed documentation: https://pkg.go.dev/embed

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | Claude Code | Initial proposal for handoff schema embedding |

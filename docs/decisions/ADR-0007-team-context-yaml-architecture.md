# ADR-0007: Rite Context YAML Architecture

| Field | Value |
|-------|-------|
| **Status** | Proposed |
| **Date** | 2026-01-04 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The current rite context injection system uses bash-based sourcing via `rite-context-loader.sh`. Rites provide executable `context-injection.sh` scripts that export an `inject_rite_context()` function. While functional, this approach has several architectural limitations:

### Current Implementation

```bash
# .claude/hooks/lib/rite-context-loader.sh
load_rite_context() {
    active_rite=$(cat ".claude/ACTIVE_RITE" 2>/dev/null || echo "")
    rite_script="$roster_home/rites/$active_rite/$RITE_CONTEXT_SCRIPT_NAME"

    # Source in subshell to isolate side effects
    output=$(
        source "$rite_script"
        "$RITE_CONTEXT_FUNCTION_NAME"
    )
}
```

### Problems with Current Approach

1. **Security**: Arbitrary code execution from rite scripts; must trust all rite authors
2. **Portability**: Bash-specific; incompatible with Go-based Ariadne CLI
3. **Testability**: Difficult to unit test bash function sourcing
4. **Discoverability**: No schema for what rites can inject; each rite invents its own format
5. **Performance**: Shell spawning overhead for every context injection call
6. **Validation**: No compile-time or startup-time validation of rite configurations

### Constraints

- **Backward Compatibility**: Existing rites must continue working during migration
- **Multi-Source**: Rite context comes from roster (central) and satellite projects (local)
- **Performance**: Context injection runs on every prompt; must be fast
- **Extensibility**: New rites should be addable without modifying Ariadne core

## Decision

Replace bash-sourced rite context with **YAML configuration files per rite**, loaded and validated by Ariadne at startup.

### Option Analysis

#### Option A: Per-Rite YAML Files (Selected)

```yaml
# rites/10x-dev/rite-context.yaml
rite:
  name: 10x-dev
  display_name: "10x Development Pack"
  domain: software development

context_injection:
  # Static context always injected
  static:
    - section: "Agent Routing"
      content: |
        When working within an orchestrated session, the main thread coordinates
        via Task tool delegation to specialist agents.

  # Dynamic context requires template rendering
  dynamic:
    - section: "Active Agents"
      source: agents/*.md
      template: "| {{.Name}} | {{.Role}} |"

  # Conditional context based on session state
  conditional:
    - when: session.status == "PARKED"
      section: "Resume Instructions"
      content: "Use /resume to continue this session."
```

**Pros**:
- Declarative, no code execution required
- Schema-validatable at startup
- Per-rite isolation maintained
- Easy to diff and review changes
- Natural migration path (one rite at a time)

**Cons**:
- Reduces flexibility for complex context injection
- Template language adds complexity
- Must define schema for all context types

#### Option B: Unified Config File

```yaml
# ariadne.yaml or roster.yaml
rites:
  10x-dev:
    context_injection:
      static: [...]
  ecosystem:
    context_injection:
      static: [...]
```

**Pros**:
- Single source of truth
- Easier global validation
- Simpler discovery

**Cons**:
- Merge conflicts when multiple rites change simultaneously
- Doesn't scale to 11+ rites
- Breaks rite modularity
- Central config becomes bottleneck for rite additions

#### Option C: Embedded in Session Context

```yaml
# SESSION_CONTEXT.md
rite_context:
  injected_at: "2026-01-04T10:00:00Z"
  sections:
    - name: "Agent Routing"
      content: "..."
```

**Pros**:
- Context travels with session
- Works offline/disconnected from roster

**Cons**:
- Bloats session files
- Context becomes stale
- Duplicate storage across sessions
- Doesn't solve the source-of-truth problem

### Selected Approach: Option A (Per-Rite YAML)

Each rite contains a `rite-context.yaml` file defining:

1. **Rite metadata**: Name, domain, display name
2. **Static context**: Always-injected markdown sections
3. **Dynamic context**: Template-rendered sections from file sources
4. **Conditional context**: Sections injected based on session state

### Schema Definition

```yaml
# schemas/rite-context-schema.yaml
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "https://roster.dev/schemas/rite-context.schema.yaml"
title: "Rite Context Configuration"
type: object

required:
  - rite
  - schema_version

properties:
  schema_version:
    type: string
    pattern: "^[0-9]+\\.[0-9]+$"

  rite:
    type: object
    required: [name, domain]
    properties:
      name:
        type: string
        pattern: "^[a-z][a-z0-9-]*-pack$"
      display_name:
        type: string
      domain:
        type: string

  context_injection:
    type: object
    properties:
      static:
        type: array
        items:
          type: object
          required: [section, content]
          properties:
            section:
              type: string
            content:
              type: string

      dynamic:
        type: array
        items:
          type: object
          required: [section, source, template]
          properties:
            section:
              type: string
            source:
              type: string
              description: "Glob pattern relative to team directory"
            template:
              type: string
              description: "Go text/template string"

      conditional:
        type: array
        items:
          type: object
          required: [when, section, content]
          properties:
            when:
              type: string
              description: "CEL expression evaluated against session context"
            section:
              type: string
            content:
              type: string
```

### Implementation in Ariadne

```go
// ariadne/internal/rite/context.go
package rite

import (
    "embed"
    _ "embed"

    "github.com/autom8y/ariadne/internal/validation"
    "gopkg.in/yaml.v3"
)

//go:embed schemas/rite-context.schema.yaml
var riteContextSchema []byte

type RiteContext struct {
    SchemaVersion string `yaml:"schema_version"`
    Rite          RiteMetadata `yaml:"rite"`
    ContextInjection ContextInjection `yaml:"context_injection"`
}

type ContextInjection struct {
    Static      []StaticSection      `yaml:"static"`
    Dynamic     []DynamicSection     `yaml:"dynamic"`
    Conditional []ConditionalSection `yaml:"conditional"`
}

func LoadRiteContext(riteDir string) (*RiteContext, error) {
    // Load and validate against schema
    // Return structured context for injection
}
```

### Migration Strategy

1. **Phase 1: Dual-Mode** (Weeks 1-2)
   - Ariadne loads both YAML config and bash scripts
   - YAML takes precedence if present
   - Logging when falling back to bash

2. **Phase 2: Rite Migration** (Weeks 3-4)
   - Migrate 10x-dev first (canonical example)
   - Migration script converts context-injection.sh to rite-context.yaml
   - Rites validate their context output matches

3. **Phase 3: Deprecation** (Week 5+)
   - Log deprecation warnings for bash scripts
   - Update rite template to use YAML
   - Remove bash loading in Ariadne v2.0

### Backward Compatibility

During migration, `rite-context-loader.sh` continues to function:

```bash
load_rite_context() {
    # Check for YAML first (new approach)
    if [[ -f "$rite_dir/rite-context.yaml" ]]; then
        ari rite context --format=markdown
        return $?
    fi

    # Fall back to bash (deprecated)
    log_warning "Rite $active_rite uses deprecated bash context injection"
    # ... existing bash logic
}
```

## Consequences

### Positive

1. **Security**: No arbitrary code execution; YAML is data, not code
2. **Portability**: Go, bash, and any language can parse YAML
3. **Validation**: JSON Schema validation catches errors at startup
4. **Discoverability**: Schema documents all available context options
5. **Performance**: Parsed once at startup, cached in memory
6. **Testability**: Easy to unit test YAML parsing and template rendering

### Negative

1. **Migration Effort**: All 11 rites need YAML files created
2. **Reduced Flexibility**: Complex context logic requires template/CEL expressions
3. **Learning Curve**: Rites must learn schema and template syntax
4. **Two Systems During Migration**: Increased complexity while both approaches coexist

### Neutral

1. **Template Language Choice**: Go text/template is powerful but has quirks
2. **CEL for Conditions**: CEL is standard but adds dependency

## Implementation Artifacts

| Artifact | Location | Purpose |
|----------|----------|---------|
| Schema | `schemas/rite-context-schema.yaml` | Validation schema for rite context files |
| Go Package | `ariadne/internal/rite/context.go` | Context loading and rendering |
| Example | `rites/10x-dev/rite-context.yaml` | Canonical example for other rites |
| Migration Script | `scripts/migrate-rite-context.sh` | Converts bash to YAML |

## Related Decisions

- **ADR-0001**: Session State Machine (session context structure)
- **ADR-0002**: Hook Library Resolution (where rite scripts live)
- **ADR-0005**: state-mate Centralized State Authority (session state access pattern)

## References

- Current implementation: `.claude/hooks/lib/rite-context-loader.sh`
- Rite pack structure: `rites/10x-dev/`
- Orchestrator YAML: `rites/*/orchestrator.yaml` (existing YAML pattern)

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | Claude Code | Initial proposal for YAML-based rite context |

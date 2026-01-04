# ADR-0007: Team Context YAML Architecture

| Field | Value |
|-------|-------|
| **Status** | Proposed |
| **Date** | 2026-01-04 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The current team context injection system uses bash-based sourcing via `team-context-loader.sh`. Teams provide executable `context-injection.sh` scripts that export an `inject_team_context()` function. While functional, this approach has several architectural limitations:

### Current Implementation

```bash
# .claude/hooks/lib/team-context-loader.sh
load_team_context() {
    active_team=$(cat ".claude/ACTIVE_TEAM" 2>/dev/null || echo "")
    team_script="$roster_home/teams/$active_team/$TEAM_CONTEXT_SCRIPT_NAME"

    # Source in subshell to isolate side effects
    output=$(
        source "$team_script"
        "$TEAM_CONTEXT_FUNCTION_NAME"
    )
}
```

### Problems with Current Approach

1. **Security**: Arbitrary code execution from team scripts; must trust all team pack authors
2. **Portability**: Bash-specific; incompatible with Go-based Ariadne CLI
3. **Testability**: Difficult to unit test bash function sourcing
4. **Discoverability**: No schema for what teams can inject; each team invents its own format
5. **Performance**: Shell spawning overhead for every context injection call
6. **Validation**: No compile-time or startup-time validation of team configurations

### Constraints

- **Backward Compatibility**: Existing team packs must continue working during migration
- **Multi-Source**: Team context comes from roster (central) and satellite projects (local)
- **Performance**: Context injection runs on every prompt; must be fast
- **Extensibility**: New teams should be addable without modifying Ariadne core

## Decision

Replace bash-sourced team context with **YAML configuration files per team**, loaded and validated by Ariadne at startup.

### Option Analysis

#### Option A: Per-Team YAML Files (Selected)

```yaml
# teams/10x-dev-pack/team-context.yaml
team:
  name: 10x-dev-pack
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
- Per-team isolation maintained
- Easy to diff and review changes
- Natural migration path (one team at a time)

**Cons**:
- Reduces flexibility for complex context injection
- Template language adds complexity
- Must define schema for all context types

#### Option B: Unified Config File

```yaml
# ariadne.yaml or roster.yaml
teams:
  10x-dev-pack:
    context_injection:
      static: [...]
  ecosystem-pack:
    context_injection:
      static: [...]
```

**Pros**:
- Single source of truth
- Easier global validation
- Simpler discovery

**Cons**:
- Merge conflicts when multiple teams change simultaneously
- Doesn't scale to 11+ teams
- Breaks team pack modularity
- Central config becomes bottleneck for team additions

#### Option C: Embedded in Session Context

```yaml
# SESSION_CONTEXT.md
team_context:
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

### Selected Approach: Option A (Per-Team YAML)

Each team pack contains a `team-context.yaml` file defining:

1. **Team metadata**: Name, domain, display name
2. **Static context**: Always-injected markdown sections
3. **Dynamic context**: Template-rendered sections from file sources
4. **Conditional context**: Sections injected based on session state

### Schema Definition

```yaml
# schemas/team-context-schema.yaml
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "https://roster.dev/schemas/team-context.schema.yaml"
title: "Team Context Configuration"
type: object

required:
  - team
  - schema_version

properties:
  schema_version:
    type: string
    pattern: "^[0-9]+\\.[0-9]+$"

  team:
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
// ariadne/internal/team/context.go
package team

import (
    "embed"
    _ "embed"

    "github.com/autom8y/ariadne/internal/validation"
    "gopkg.in/yaml.v3"
)

//go:embed schemas/team-context.schema.yaml
var teamContextSchema []byte

type TeamContext struct {
    SchemaVersion string `yaml:"schema_version"`
    Team          TeamMetadata `yaml:"team"`
    ContextInjection ContextInjection `yaml:"context_injection"`
}

type ContextInjection struct {
    Static      []StaticSection      `yaml:"static"`
    Dynamic     []DynamicSection     `yaml:"dynamic"`
    Conditional []ConditionalSection `yaml:"conditional"`
}

func LoadTeamContext(teamDir string) (*TeamContext, error) {
    // Load and validate against schema
    // Return structured context for injection
}
```

### Migration Strategy

1. **Phase 1: Dual-Mode** (Weeks 1-2)
   - Ariadne loads both YAML config and bash scripts
   - YAML takes precedence if present
   - Logging when falling back to bash

2. **Phase 2: Team Migration** (Weeks 3-4)
   - Migrate 10x-dev-pack first (canonical example)
   - Migration script converts context-injection.sh to team-context.yaml
   - Teams validate their context output matches

3. **Phase 3: Deprecation** (Week 5+)
   - Log deprecation warnings for bash scripts
   - Update team pack template to use YAML
   - Remove bash loading in Ariadne v2.0

### Backward Compatibility

During migration, `team-context-loader.sh` continues to function:

```bash
load_team_context() {
    # Check for YAML first (new approach)
    if [[ -f "$team_dir/team-context.yaml" ]]; then
        ari team context --format=markdown
        return $?
    fi

    # Fall back to bash (deprecated)
    log_warning "Team $active_team uses deprecated bash context injection"
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

1. **Migration Effort**: All 11 team packs need YAML files created
2. **Reduced Flexibility**: Complex context logic requires template/CEL expressions
3. **Learning Curve**: Teams must learn schema and template syntax
4. **Two Systems During Migration**: Increased complexity while both approaches coexist

### Neutral

1. **Template Language Choice**: Go text/template is powerful but has quirks
2. **CEL for Conditions**: CEL is standard but adds dependency

## Implementation Artifacts

| Artifact | Location | Purpose |
|----------|----------|---------|
| Schema | `schemas/team-context-schema.yaml` | Validation schema for team context files |
| Go Package | `ariadne/internal/team/context.go` | Context loading and rendering |
| Example | `teams/10x-dev-pack/team-context.yaml` | Canonical example for other teams |
| Migration Script | `scripts/migrate-team-context.sh` | Converts bash to YAML |

## Related Decisions

- **ADR-0001**: Session State Machine (session context structure)
- **ADR-0002**: Hook Library Resolution (where team scripts live)
- **ADR-0005**: state-mate Centralized State Authority (session state access pattern)

## References

- Current implementation: `.claude/hooks/lib/team-context-loader.sh`
- Team pack structure: `teams/10x-dev-pack/`
- Orchestrator YAML: `teams/*/orchestrator.yaml` (existing YAML pattern)

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | Claude Code | Initial proposal for YAML-based team context |

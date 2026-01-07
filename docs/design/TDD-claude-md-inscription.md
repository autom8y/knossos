# TDD: CLAUDE.md Inscription System

> Technical Design Document for the CLAUDE.md templating, ownership tracking, and synchronization system.

**Status**: Draft
**Author**: architect
**Date**: 2026-01-06
**PRD Reference**: PRD-claude-md-descriptive-architecture
**Initiative**: Knossos Doctrine v2 - The Inscription

---

## 1. Overview

CLAUDE.md is "The Inscription"--the labyrinth's entrance declaration that tells agents what heroes, rites, and customs are available. This TDD specifies the template and generation system that keeps the inscription synchronized with living configuration.

### 1.1 The Inscription Metaphor

Per the Knossos Doctrine:

> At the entrance to every labyrinth, there are words carved in stone. **CLAUDE.md** is the Inscription--the labyrinth speaking to those who would enter. It tells Theseus:
> - What heroes may be summoned
> - What rites are available
> - What customs govern this place
> - How to navigate without becoming lost

The Inscription is not static. When rites change, when heroes join or depart, the labyrinth updates its entrance.

### 1.2 Design Goals

1. **Synchronized**: Keep CLAUDE.md in sync with active rite and agent configuration
2. **Preserving**: Never destroy user customizations in satellite-owned regions
3. **Idempotent**: Running sync twice produces identical results
4. **Validatable**: Support dry-run mode for preview without write
5. **Integrated**: Work seamlessly with existing hook system and Ariadne CLI

### 1.3 Scope

| In Scope | Out of Scope |
|----------|--------------|
| Template marker syntax and schema | Content authoring guidelines |
| KNOSSOS_MANIFEST.yaml specification | Migration of existing CLAUDE.md files |
| Sync pipeline architecture | Hook registration (uses existing ari hooks) |
| Content section structure | Progressive disclosure implementation |
| Conflict resolution strategy | User preference storage |

---

## 2. Decision Matrix

Decisions locked from stakeholder interview:

| Dimension | Decision | Rationale |
|-----------|----------|-----------|
| Template markers | `<!-- KNOSSOS:START -->...<!-- KNOSSOS:END -->` | HTML comments are invisible to readers, parseable |
| Ownership tracking | `KNOSSOS_MANIFEST.yaml` | YAML for human readability, machine validation |
| Sync trigger | On rite change (team swap) | Natural synchronization point |
| Template location | `knossos/templates/CLAUDE.md.tpl` | Centralized template management |
| Content philosophy | Full context dump with progressive disclosure | Agents need comprehensive context |
| Hero declaration | Table from ACTIVE_RITE | Already implemented in `claudemd.go` |
| Domain context | Tech stack + architecture + conventions + workflows | Complete context for agent success |
| Terms | Table format + inline explanations | Both reference and contextual usage |
| Section order | Interleaved + configurable | Flexibility for different project needs |
| Conditionals | KISS initially (simple region presence/absence) | Avoid premature complexity |

---

## 3. Template Marker Syntax

### 3.1 Marker Structure

```
<!-- KNOSSOS:{DIRECTIVE} {REGION_NAME} [{OPTIONS}] -->
```

**Components**:
- `KNOSSOS:` - Namespace prefix (required)
- `{DIRECTIVE}` - Operation type: `START`, `END`, `ANCHOR`
- `{REGION_NAME}` - Unique identifier for the region (kebab-case)
- `[{OPTIONS}]` - Optional configuration (key=value pairs)

### 3.2 Directives

| Directive | Purpose | Example |
|-----------|---------|---------|
| `START` | Begin a managed region | `<!-- KNOSSOS:START quick-start -->` |
| `END` | End a managed region | `<!-- KNOSSOS:END quick-start -->` |
| `ANCHOR` | Single-line insertion point | `<!-- KNOSSOS:ANCHOR agent-table -->` |

### 3.3 Complete Marker Examples

```markdown
<!-- KNOSSOS:START execution-mode -->
## Execution Mode

This project supports three operating modes...
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

This project uses a 5-agent workflow...
<!-- KNOSSOS:END quick-start -->

<!-- KNOSSOS:ANCHOR agent-count -->
```

### 3.4 Region Naming Conventions

| Pattern | Usage | Example |
|---------|-------|---------|
| `{section-name}` | Section content | `quick-start`, `agent-routing` |
| `{section-name}-header` | Section header only | `quick-start-header` |
| `{section-name}-{subsection}` | Nested content | `hooks-configuration` |
| `meta-{name}` | Metadata regions | `meta-version`, `meta-updated` |

### 3.5 Nesting Rules

1. **No nesting allowed**: Regions cannot contain other regions
2. **Sequential only**: Regions must be sequential, not overlapping
3. **Balanced pairs**: Every `START` must have matching `END`
4. **Anchors are atomic**: `ANCHOR` directives stand alone (no `END`)

**Invalid**:
```markdown
<!-- KNOSSOS:START outer -->
<!-- KNOSSOS:START inner -->  <!-- ERROR: nested START -->
<!-- KNOSSOS:END inner -->
<!-- KNOSSOS:END outer -->
```

**Valid**:
```markdown
<!-- KNOSSOS:START section-a -->
Content A
<!-- KNOSSOS:END section-a -->

<!-- KNOSSOS:START section-b -->
Content B
<!-- KNOSSOS:END section-b -->
```

### 3.6 Escape Mechanisms

To include literal marker text in content:

1. **Code blocks**: Markers inside fenced code blocks are ignored
2. **Backslash escape**: `\<!-- KNOSSOS:START -->` treated as literal
3. **HTML entities**: `&lt;!-- KNOSSOS:START --&gt;` for documentation

### 3.7 Backward Compatibility

Existing markers are supported during migration:

| Legacy Marker | New Marker | Migration |
|---------------|------------|-----------|
| `<!-- PRESERVE: satellite-owned -->` | `<!-- KNOSSOS:START {name} owner=satellite -->` | Automatic |
| `<!-- SYNC: roster-owned -->` | `<!-- KNOSSOS:START {name} owner=roster -->` | Automatic |

---

## 4. KNOSSOS_MANIFEST.yaml Schema

### 4.1 File Location

```
.claude/KNOSSOS_MANIFEST.yaml
```

### 4.2 Schema Definition

```yaml
# KNOSSOS_MANIFEST.yaml JSON Schema (draft-2020-12)
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "embed:///schemas/knossos-manifest.schema.json"
title: "KNOSSOS_MANIFEST Schema"
description: "Configuration for CLAUDE.md inscription system"
type: object

required:
  - schema_version
  - inscription_version
  - regions

properties:
  schema_version:
    type: string
    pattern: "^[0-9]+\\.[0-9]+$"
    default: "1.0"
    description: "Manifest schema version"

  inscription_version:
    type: string
    pattern: "^[0-9]+$"
    description: "Incremented on each sync operation"

  last_sync:
    type: string
    format: date-time
    description: "ISO 8601 timestamp of last sync"

  active_rite:
    type: string
    description: "Current rite/rite name"

  template_path:
    type: string
    default: "knossos/templates/CLAUDE.md.tpl"
    description: "Path to master template"

  regions:
    type: object
    additionalProperties:
      $ref: "#/$defs/region"
    description: "Region definitions keyed by region name"

  section_order:
    type: array
    items:
      type: string
    description: "Ordered list of section identifiers"

  conditionals:
    type: object
    additionalProperties:
      $ref: "#/$defs/conditional"
    description: "Conditional inclusion rules"

$defs:
  region:
    type: object
    required: ["owner"]
    properties:
      owner:
        type: string
        enum: ["knossos", "satellite", "regenerate"]
        description: |
          knossos: Managed by Knossos templates
          satellite: Owned by satellite project
          regenerate: Generated from project state
      source:
        type: string
        description: "Data source for regenerate regions"
      preserve_on_conflict:
        type: boolean
        default: false
        description: "If true, satellite edits preserved on conflict"
      hash:
        type: string
        description: "SHA256 hash of last synced content"
      synced_at:
        type: string
        format: date-time

  conditional:
    type: object
    required: ["when", "include"]
    properties:
      when:
        type: string
        description: "Condition expression"
      include:
        type: array
        items:
          type: string
        description: "Regions to include when condition is true"
      exclude:
        type: array
        items:
          type: string
        description: "Regions to exclude when condition is true"
```

### 4.3 Example KNOSSOS_MANIFEST.yaml

```yaml
schema_version: "1.0"
inscription_version: "42"
last_sync: "2026-01-06T10:30:00Z"
active_rite: "10x-dev-pack"
template_path: "knossos/templates/CLAUDE.md.tpl"

regions:
  # Knossos-managed sections (sync from templates)
  execution-mode:
    owner: knossos
    hash: "a1b2c3d4e5f6..."
    synced_at: "2026-01-06T10:30:00Z"

  knossos-identity:
    owner: knossos
    hash: "f6e5d4c3b2a1..."
    synced_at: "2026-01-06T10:30:00Z"

  agent-routing:
    owner: knossos
    hash: "1a2b3c4d5e6f..."
    synced_at: "2026-01-06T10:30:00Z"

  skills:
    owner: knossos
    hash: "6f5e4d3c2b1a..."
    synced_at: "2026-01-06T10:30:00Z"

  hooks:
    owner: knossos
    hash: "b1a2c3d4e5f6..."
    synced_at: "2026-01-06T10:30:00Z"

  dynamic-context:
    owner: knossos
    hash: "e5f6a1b2c3d4..."
    synced_at: "2026-01-06T10:30:00Z"

  ariadne-cli:
    owner: knossos
    hash: "c3d4e5f6a1b2..."
    synced_at: "2026-01-06T10:30:00Z"

  getting-help:
    owner: knossos
    hash: "d4e5f6a1b2c3..."
    synced_at: "2026-01-06T10:30:00Z"

  state-management:
    owner: knossos
    hash: "4d5e6f1a2b3c..."
    synced_at: "2026-01-06T10:30:00Z"

  slash-commands:
    owner: knossos
    hash: "5e6f1a2b3c4d..."
    synced_at: "2026-01-06T10:30:00Z"

  # Regenerated sections (from project state)
  quick-start:
    owner: regenerate
    source: "ACTIVE_RITE+agents"
    hash: "2b3c4d5e6f1a..."
    synced_at: "2026-01-06T10:30:00Z"

  agent-configurations:
    owner: regenerate
    source: "agents/*.md"
    hash: "3c4d5e6f1a2b..."
    synced_at: "2026-01-06T10:30:00Z"

  # Satellite-owned sections (never overwritten)
  project-conventions:
    owner: satellite
    preserve_on_conflict: true

  project-deployment:
    owner: satellite
    preserve_on_conflict: true

section_order:
  - execution-mode
  - knossos-identity
  - quick-start
  - agent-routing
  - skills
  - agent-configurations
  - hooks
  - dynamic-context
  - ariadne-cli
  - getting-help
  - state-management
  - slash-commands
  # Satellite sections append after knossos sections

conditionals:
  orchestrated-mode:
    when: "session.active && session.team"
    include:
      - agent-routing
      - state-management
    exclude: []
```

### 4.4 Owner Types

| Owner | Behavior | Source |
|-------|----------|--------|
| `knossos` | Overwritten on sync | Template file |
| `satellite` | Never overwritten | Satellite CLAUDE.md |
| `regenerate` | Regenerated from state | Project files (ACTIVE_RITE, agents/) |

### 4.5 Validation Constraints

1. **Region names unique**: No duplicate keys in `regions`
2. **Section order complete**: All regions in `section_order`
3. **Source required for regenerate**: `source` field mandatory when `owner: regenerate`
4. **Valid conditional expressions**: `when` must be parseable expression
5. **Hash format**: SHA256 hex string (64 characters)

---

## 5. Sync Pipeline Architecture

### 5.1 Pipeline Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Sync Pipeline                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐          │
│  │ Trigger  │───▶│ Validate │───▶│ Generate │───▶│  Merge   │          │
│  │ Detection│    │ Manifest │    │ Content  │    │ Regions  │          │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘          │
│       │               │               │               │                  │
│       ▼               ▼               ▼               ▼                  │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐          │
│  │ Rite     │    │ Schema   │    │ Template │    │ Conflict │          │
│  │ Change   │    │ Validate │    │ Render   │    │ Resolve  │          │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘          │
│                                                         │                │
│                        ┌──────────┐    ┌──────────┐    │                │
│                        │  Write   │◀───│  Verify  │◀───┘                │
│                        │  Output  │    │  Hashes  │                     │
│                        └──────────┘    └──────────┘                     │
│                              │                                           │
│                              ▼                                           │
│                        ┌──────────┐                                      │
│                        │  Update  │                                      │
│                        │ Manifest │                                      │
│                        └──────────┘                                      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Stage Details

#### Stage 1: Trigger Detection

**Triggers**:
1. `ari team switch {rite}` - Rite change
2. `ari inscription sync` - Manual sync command
3. `ari inscription sync --force` - Force full regeneration

**Implementation**:
```go
// ariadne/internal/inscription/trigger.go

type TriggerEvent struct {
    Type      TriggerType
    RiteName  string
    Force     bool
    Timestamp time.Time
}

type TriggerType string

const (
    TriggerRiteChange TriggerType = "rite_change"
    TriggerManual     TriggerType = "manual"
    TriggerForce      TriggerType = "force"
)

func (p *Pipeline) ShouldSync(event TriggerEvent) bool {
    manifest, err := p.loadManifest()
    if err != nil {
        return true // Sync on error to recover
    }

    if event.Force {
        return true
    }

    return manifest.ActiveRite != event.RiteName
}
```

#### Stage 2: Validate Manifest

**Validations**:
1. Parse KNOSSOS_MANIFEST.yaml
2. Validate against JSON Schema
3. Verify all referenced regions exist
4. Check section_order completeness

**Error Handling**:
- Missing manifest: Create default
- Invalid schema: Log error, create backup, regenerate
- Missing regions: Add with default owner (satellite)

#### Stage 3: Generate Content

**Content Sources**:

| Source | Content Type | Generator |
|--------|--------------|-----------|
| Template | Knossos-owned sections | `template.Execute()` |
| ACTIVE_RITE | Team name, agent count | `team.LoadActive()` |
| agents/*.md | Agent table, configurations | `claudemd.LoadAgentInfos()` |
| Manifest | Satellite sections | Pass-through |

**Template Processing**:
```go
// ariadne/internal/inscription/generator.go

type Generator struct {
    templatePath string
    manifest     *Manifest
    context      *RenderContext
}

type RenderContext struct {
    ActiveRite  string
    AgentCount  int
    Agents      []AgentInfo
    KnossosVars map[string]string
}

func (g *Generator) RenderRegion(name string) (string, error) {
    region := g.manifest.Regions[name]

    switch region.Owner {
    case "knossos":
        return g.renderTemplate(name)
    case "regenerate":
        return g.regenerateFromSource(name, region.Source)
    case "satellite":
        return g.passThrough(name)
    }
}
```

#### Stage 4: Merge Regions

**Merge Algorithm**:

```
FOR each region in section_order:
    new_content = generator.RenderRegion(region)
    old_content = extractRegion(current_claudemd, region)

    IF region.owner == "satellite":
        final_content = old_content  # Never overwrite
    ELSE IF region.owner == "knossos":
        final_content = new_content  # Always sync
    ELSE IF region.owner == "regenerate":
        IF hash(old_content) == manifest.regions[region].hash:
            final_content = new_content  # Clean regenerate
        ELSE:
            # User modified regenerated content
            IF manifest.regions[region].preserve_on_conflict:
                final_content = old_content
                LOG warning "User edits preserved in {region}"
            ELSE:
                final_content = new_content
                LOG warning "User edits overwritten in {region}"

    output.append(wrapWithMarkers(region, final_content))

# Append unknown sections (satellite-owned by default)
FOR each section NOT in section_order:
    output.append(section)
```

#### Stage 5: Conflict Resolution

**Conflict Types**:

| Conflict | Detection | Resolution |
|----------|-----------|------------|
| User edited knossos region | Hash mismatch | Overwrite with warning |
| User edited regenerate region | Hash mismatch | Configurable (preserve_on_conflict) |
| New satellite section | Not in manifest | Preserve, add to manifest |
| Removed satellite section | In manifest but not file | Remove from manifest |
| Malformed markers | Parse failure | Treat as satellite content |

**Conflict Log**:
```yaml
# .claude/inscription-conflicts.log
- timestamp: "2026-01-06T10:35:00Z"
  region: "agent-routing"
  type: "knossos_overwrite"
  message: "User edits overwritten in knossos-owned region"
  backup: ".claude/backups/CLAUDE.md.2026-01-06T10-35-00Z"
```

#### Stage 6: Write Output

**Write Protocol**:
1. Create backup: `.claude/backups/CLAUDE.md.{timestamp}`
2. Write to temp file: `.claude/CLAUDE.md.tmp`
3. Validate output (markers balanced, sections present)
4. Atomic rename: `mv .claude/CLAUDE.md.tmp .claude/CLAUDE.md`
5. Update manifest hashes
6. Write manifest

#### Stage 7: Rollback Mechanism

**Rollback Triggers**:
- Validation failure after write
- User request: `ari inscription rollback`
- Hash verification failure

**Rollback Process**:
```bash
# Keep last 5 backups
ls -t .claude/backups/CLAUDE.md.* | tail -n +6 | xargs rm -f

# Rollback to most recent
cp .claude/backups/CLAUDE.md.{latest} .claude/CLAUDE.md
```

### 5.3 Sequence Diagram

```
┌────────┐   ┌─────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  User  │   │   ari   │   │ Pipeline │   │ Generator│   │ Manifest │
└───┬────┘   └────┬────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
    │             │             │              │              │
    │ team switch │             │              │              │
    │────────────▶│             │              │              │
    │             │ trigger     │              │              │
    │             │────────────▶│              │              │
    │             │             │ load         │              │
    │             │             │─────────────────────────────▶
    │             │             │              │    manifest  │
    │             │             │◀─────────────────────────────
    │             │             │              │              │
    │             │             │ validate     │              │
    │             │             │─────────────▶│              │
    │             │             │   OK         │              │
    │             │             │◀─────────────│              │
    │             │             │              │              │
    │             │             │ render       │              │
    │             │             │─────────────▶│              │
    │             │             │   content    │              │
    │             │             │◀─────────────│              │
    │             │             │              │              │
    │             │             │ merge        │              │
    │             │             │──────┐       │              │
    │             │             │      │       │              │
    │             │             │◀─────┘       │              │
    │             │             │              │              │
    │             │             │ backup + write│             │
    │             │             │──────────────────────┐      │
    │             │             │                      │      │
    │             │             │◀─────────────────────┘      │
    │             │             │              │              │
    │             │             │ update hashes│              │
    │             │             │─────────────────────────────▶
    │             │             │              │              │
    │             │  success    │              │              │
    │             │◀────────────│              │              │
    │  done       │             │              │              │
    │◀────────────│             │              │              │
```

---

## 6. Content Section Structure

### 6.1 Section Hierarchy

```
CLAUDE.md
├── Header (# CLAUDE.md)
├── Tagline (> Entry point for Claude Code...)
│
├── ## Execution Mode [knossos]
│   └── Operating modes table
│
├── ## Knossos Identity [knossos]
│   └── Mythology reference table
│
├── ## Quick Start [regenerate:ACTIVE_RITE+agents]
│   ├── Team summary line
│   └── Agent table
│
├── ## Agent Routing [knossos]
│   └── Orchestration guidance
│
├── ## Skills [knossos]
│   └── Skill invocation reference
│
├── ## Agent Configurations [regenerate:agents/*.md]
│   └── Agent file list with descriptions
│
├── ## Hooks [knossos]
│   └── Hook documentation
│
├── ## Dynamic Context [knossos]
│   └── Bang command syntax
│
├── ## Ariadne CLI [knossos]
│   └── CLI reference
│
├── ## Getting Help [knossos]
│   └── Navigation table
│
├── ## State Management [knossos]
│   └── Moirai usage
│
├── ## Slash Commands [knossos]
│   └── Command reference
│
└── ## Project:* [satellite]
    └── Custom satellite sections
```

### 6.2 Section Ownership Matrix

| Section | Owner | Sync Behavior | Source |
|---------|-------|---------------|--------|
| Header/Tagline | knossos | Always sync | Template |
| Execution Mode | knossos | Always sync | Template |
| Knossos Identity | knossos | Always sync | Template |
| Quick Start | regenerate | Regenerate from state | ACTIVE_RITE + agents/ |
| Agent Routing | knossos | Always sync | Template |
| Skills | knossos | Always sync | Template |
| Agent Configurations | regenerate | Regenerate from state | agents/*.md |
| Hooks | knossos | Always sync | Template |
| Dynamic Context | knossos | Always sync | Template |
| Ariadne CLI | knossos | Always sync | Template |
| Getting Help | knossos | Always sync | Template |
| State Management | knossos | Always sync | Template |
| Slash Commands | knossos | Always sync | Template |
| Project:* | satellite | Never sync | Satellite |
| Unknown sections | satellite | Never sync | Satellite |

### 6.3 Section Order Configuration

Default order in `KNOSSOS_MANIFEST.yaml`:

```yaml
section_order:
  # Core navigation (read first)
  - execution-mode
  - knossos-identity

  # Team context (who is available)
  - quick-start
  - agent-routing
  - skills
  - agent-configurations

  # Infrastructure (how things work)
  - hooks
  - dynamic-context
  - ariadne-cli

  # Reference (consult as needed)
  - getting-help
  - state-management
  - slash-commands

  # Satellite sections append automatically
```

### 6.4 Conditional Inclusion

Initial implementation uses simple presence/absence:

```yaml
conditionals:
  # Include state-management only if sessions are used
  state-management:
    when: "file_exists('.claude/sessions')"
    include: [state-management]
    exclude: []

  # Include orchestration only with active rite
  orchestrated-content:
    when: "file_exists('.claude/ACTIVE_RITE')"
    include: [agent-routing, skills]
    exclude: []
```

**Condition Expressions** (v1.0 - KISS):

| Expression | Meaning |
|------------|---------|
| `file_exists(path)` | File or directory exists |
| `env_set(VAR)` | Environment variable is set |
| `always` | Always include |
| `never` | Never include |

---

## 7. Template File Specification

### 7.1 Template Location

```
knossos/templates/
├── CLAUDE.md.tpl              # Main template
├── sections/
│   ├── execution-mode.md.tpl
│   ├── knossos-identity.md.tpl
│   ├── agent-routing.md.tpl
│   ├── skills.md.tpl
│   ├── hooks.md.tpl
│   ├── dynamic-context.md.tpl
│   ├── ariadne-cli.md.tpl
│   ├── getting-help.md.tpl
│   ├── state-management.md.tpl
│   └── slash-commands.md.tpl
└── partials/
    ├── agent-table.md.tpl
    └── terminology-table.md.tpl
```

### 7.2 Template Syntax

Uses Go `text/template` with custom functions:

```go
// Template functions
template.FuncMap{
    "include": includePartial,
    "ifdef":   conditionalInclude,
    "agents":  loadAgentTable,
    "term":    lookupTerminology,
}
```

### 7.3 Main Template Structure

```markdown
# CLAUDE.md

> Entry point for Claude Code. Skills-based progressive disclosure architecture.

<!-- KNOSSOS:START execution-mode -->
{{include "sections/execution-mode.md.tpl"}}
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START knossos-identity -->
{{include "sections/knossos-identity.md.tpl"}}
<!-- KNOSSOS:END knossos-identity -->

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

This project uses a {{.AgentCount}}-agent workflow ({{.ActiveRite}}):

{{include "partials/agent-table.md.tpl"}}

**New here?** Use the `prompting` skill for copy-paste patterns, or `initiative-scoping` to start a new project.
<!-- KNOSSOS:END quick-start -->

{{/* ... remaining sections ... */}}
```

### 7.4 Section Template Example

```markdown
{{/* sections/execution-mode.md.tpl */}}
## Execution Mode

This project supports three operating modes (see PRD-hybrid-session-model for details):

| Mode | Session | Team | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Coach pattern, delegate via Task tool |

**Unsure?** Use `/consult` for workflow routing.

For enforcement rules: `orchestration/execution-mode.md`
```

---

## 8. Implementation Architecture

### 8.1 Package Structure

```
ariadne/internal/inscription/
├── pipeline.go          # Main sync pipeline
├── pipeline_test.go
├── manifest.go          # KNOSSOS_MANIFEST.yaml handling
├── manifest_test.go
├── marker.go            # Marker parsing/writing
├── marker_test.go
├── generator.go         # Content generation
├── generator_test.go
├── merger.go            # Region merging
├── merger_test.go
├── conflict.go          # Conflict detection/resolution
├── conflict_test.go
├── template.go          # Template rendering
├── template_test.go
├── backup.go            # Backup/rollback
└── backup_test.go
```

### 8.2 Core Types

```go
// ariadne/internal/inscription/types.go

// Manifest represents KNOSSOS_MANIFEST.yaml
type Manifest struct {
    SchemaVersion      string             `yaml:"schema_version"`
    InscriptionVersion string             `yaml:"inscription_version"`
    LastSync           time.Time          `yaml:"last_sync"`
    ActiveRite         string             `yaml:"active_rite"`
    TemplatePath       string             `yaml:"template_path"`
    Regions            map[string]*Region `yaml:"regions"`
    SectionOrder       []string           `yaml:"section_order"`
    Conditionals       map[string]*Conditional `yaml:"conditionals"`
}

// Region represents a managed region in CLAUDE.md
type Region struct {
    Owner             OwnerType `yaml:"owner"`
    Source            string    `yaml:"source,omitempty"`
    PreserveOnConflict bool     `yaml:"preserve_on_conflict,omitempty"`
    Hash              string    `yaml:"hash,omitempty"`
    SyncedAt          time.Time `yaml:"synced_at,omitempty"`
}

type OwnerType string

const (
    OwnerKnossos    OwnerType = "knossos"
    OwnerSatellite  OwnerType = "satellite"
    OwnerRegenerate OwnerType = "regenerate"
)

// Conditional represents a conditional inclusion rule
type Conditional struct {
    When    string   `yaml:"when"`
    Include []string `yaml:"include,omitempty"`
    Exclude []string `yaml:"exclude,omitempty"`
}

// Marker represents a parsed KNOSSOS marker
type Marker struct {
    Directive  MarkerDirective
    RegionName string
    Options    map[string]string
    LineNumber int
}

type MarkerDirective string

const (
    DirectiveStart  MarkerDirective = "START"
    DirectiveEnd    MarkerDirective = "END"
    DirectiveAnchor MarkerDirective = "ANCHOR"
)
```

### 8.3 Pipeline Interface

```go
// ariadne/internal/inscription/pipeline.go

type Pipeline struct {
    claudeMDPath   string
    manifestPath   string
    templateDir    string
    backupDir      string
}

func NewPipeline(projectRoot string) *Pipeline

// Sync performs the full sync pipeline
func (p *Pipeline) Sync(opts SyncOptions) (*SyncResult, error)

// DryRun previews changes without writing
func (p *Pipeline) DryRun(opts SyncOptions) (*SyncPreview, error)

// Rollback reverts to a previous backup
func (p *Pipeline) Rollback(timestamp string) error

// Validate checks current state without sync
func (p *Pipeline) Validate() (*ValidationResult, error)

type SyncOptions struct {
    Force      bool   // Force full regeneration
    RiteName   string // New rite name (empty = current)
    DryRun     bool   // Preview only
    Verbose    bool   // Detailed output
}

type SyncResult struct {
    Success       bool
    RegionsSynced []string
    Conflicts     []Conflict
    BackupPath    string
    Duration      time.Duration
}
```

### 8.4 CLI Commands

```bash
# Manual sync
ari inscription sync [--force] [--dry-run] [--verbose]

# Validate current state
ari inscription validate

# Rollback to backup
ari inscription rollback [timestamp]

# List backups
ari inscription backups

# Show diff between current and generated
ari inscription diff [region]
```

---

## 9. Integration Points

### 9.1 Team Switch Integration

Modify `ariadne/internal/team/switch.go`:

```go
func (s *Switcher) Switch(riteName string) error {
    // ... existing team switch logic ...

    // Trigger inscription sync
    pipeline := inscription.NewPipeline(s.projectRoot)
    result, err := pipeline.Sync(inscription.SyncOptions{
        RiteName: riteName,
    })
    if err != nil {
        return fmt.Errorf("inscription sync failed: %w", err)
    }

    if len(result.Conflicts) > 0 {
        s.logConflicts(result.Conflicts)
    }

    return nil
}
```

### 9.2 Hook Integration

The inscription system uses existing hooks:

| Hook | Event | Purpose |
|------|-------|---------|
| `ari/context.sh` | SessionStart | Reads updated CLAUDE.md |
| `ari/thread.sh` | PostToolUse | Tracks CLAUDE.md modifications |

No new hooks required. The SessionStart hook naturally picks up inscription changes.

### 9.3 Existing Code Preservation

The existing `claudemd.go` code is preserved and extended:

```go
// ClaudeMDUpdater is still used for regenerate regions
func (u *ClaudeMDUpdater) UpdateForTeam(team *Team) error
func (u *ClaudeMDUpdater) generateQuickStartContent(team *Team, agents []agentFileInfo) []string
func (u *ClaudeMDUpdater) generateAgentConfigsContent(agents []agentFileInfo) []string
```

These functions are called by the Generator for regenerate-type regions.

---

## 10. Edge Cases and Error Handling

### 10.1 Edge Cases

| Case | Handling |
|------|----------|
| No CLAUDE.md exists | Create from template |
| No manifest exists | Create default manifest, treat all as satellite |
| Malformed markers | Log warning, treat content as satellite |
| Missing region in file | Skip region, add to manifest as satellite |
| Template file missing | Error, abort sync |
| Circular template includes | Error with stack trace |
| Very large CLAUDE.md (>100KB) | Warn, proceed with sync |
| Binary content in region | Error, region marked invalid |

### 10.2 Error Codes

```go
const (
    ErrManifestParse      = 1  // YAML parse error
    ErrManifestValidation = 2  // Schema validation failed
    ErrTemplateNotFound   = 3  // Template file missing
    ErrTemplateRender     = 4  // Template execution error
    ErrMarkerParse        = 5  // Marker syntax error
    ErrMarkerUnbalanced   = 6  // START without END
    ErrWriteFailed        = 7  // File write error
    ErrBackupFailed       = 8  // Backup creation failed
    ErrHashMismatch       = 9  // Integrity check failed
    ErrRollbackFailed     = 10 // Rollback failed
)
```

### 10.3 Graceful Degradation

If sync fails:
1. Log detailed error
2. Preserve existing CLAUDE.md (no partial writes)
3. Create diagnostic file: `.claude/inscription-error.log`
4. Continue with team switch (don't block)

---

## 11. Test Strategy

### 11.1 Unit Tests

| Package | Test Focus | Coverage |
|---------|-----------|----------|
| `inscription/marker` | Marker parsing, all directives | 100% |
| `inscription/manifest` | YAML parsing, validation | 100% |
| `inscription/generator` | Template rendering, regeneration | 100% |
| `inscription/merger` | Region merging, conflict detection | 100% |
| `inscription/backup` | Backup creation, rollback | 100% |

### 11.2 Integration Tests

| Test ID | Description | Verifies |
|---------|-------------|----------|
| `inscription_001` | Sync on clean project | Full pipeline |
| `inscription_002` | Sync with existing satellite sections | Preservation |
| `inscription_003` | Sync with user edits to knossos regions | Overwrite + warning |
| `inscription_004` | Sync with user edits to regenerate regions | Conflict resolution |
| `inscription_005` | Rollback after failed sync | Recovery |
| `inscription_006` | Dry-run preview | No file changes |
| `inscription_007` | Idempotent sync (run twice) | Same result |
| `inscription_008` | Legacy marker migration | Backward compatibility |
| `inscription_009` | Team switch triggers sync | Integration |
| `inscription_010` | Template with conditionals | Conditional inclusion |

### 11.3 Compatibility Matrix

| Scenario | Expected Behavior |
|----------|-------------------|
| Satellite with no manifest | Create default manifest |
| Satellite with legacy markers | Migrate markers on sync |
| Satellite with custom sections | Preserve all custom sections |
| Satellite with modified knossos sections | Overwrite with warning |

---

## 12. Migration Strategy

### 12.1 Migration Path

1. **Phase 1**: Deploy inscription system (no migration)
2. **Phase 2**: Add `ari inscription migrate` command
3. **Phase 3**: Auto-migrate on first sync (opt-in via flag)
4. **Phase 4**: Make migration automatic (opt-out via flag)

### 12.2 Migration Command

```bash
ari inscription migrate [--dry-run] [--backup]
```

Migration process:
1. Parse existing CLAUDE.md
2. Detect legacy markers (`<!-- PRESERVE: -->`, `<!-- SYNC: -->`)
3. Map to new KNOSSOS markers
4. Generate initial manifest
5. Write updated CLAUDE.md

### 12.3 Backward Compatibility

During transition period:
- Accept both legacy and new markers
- On write, use new markers
- Log deprecation warnings for legacy markers

---

## 13. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Data loss on sync | Low | High | Automatic backups, atomic writes |
| Template errors break CLAUDE.md | Medium | High | Validation before write, dry-run |
| Marker parsing edge cases | Medium | Medium | Comprehensive tests, graceful fallback |
| Performance with large files | Low | Low | Streaming parser, early exit |
| Conflicts confuse users | Medium | Medium | Clear conflict messages, rollback option |
| Migration breaks existing workflows | Low | High | Opt-in migration, backward compat |

---

## 14. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-inscription-001 | Proposed | Marker syntax choice (HTML comments vs frontmatter) |
| ADR-inscription-002 | Proposed | Manifest file format (YAML vs JSON vs TOML) |
| ADR-inscription-003 | Proposed | Conflict resolution strategy |

---

## 15. Implementation Roadmap

### Phase 1: Foundation (Week 1)

| Task | Files | Deliverable |
|------|-------|-------------|
| Marker parser | `ariadne/internal/inscription/marker.go` | Parse KNOSSOS markers |
| Manifest schema | `ariadne/internal/validation/schemas/knossos-manifest.schema.json` | JSON Schema |
| Manifest loader | `ariadne/internal/inscription/manifest.go` | Load/save manifest |

### Phase 2: Pipeline Core (Week 2)

| Task | Files | Deliverable |
|------|-------|-------------|
| Generator | `ariadne/internal/inscription/generator.go` | Content generation |
| Merger | `ariadne/internal/inscription/merger.go` | Region merging |
| Backup/rollback | `ariadne/internal/inscription/backup.go` | Recovery system |

### Phase 3: CLI Integration (Week 3)

| Task | Files | Deliverable |
|------|-------|-------------|
| CLI commands | `ariadne/internal/cmd/inscription/*.go` | `ari inscription *` |
| Team switch hook | `ariadne/internal/team/switch.go` | Auto-sync on switch |
| Templates | `knossos/templates/**/*.tpl` | Section templates |

### Phase 4: Validation & Rollout (Week 4)

| Task | Files | Deliverable |
|------|-------|-------------|
| Integration tests | `tests/integration/inscription_test.go` | E2E tests |
| Migration tool | `ariadne/internal/inscription/migrate.go` | Legacy migration |
| Documentation | `docs/guides/inscription-system.md` | User guide |

---

## 16. Handoff Criteria

Ready for Implementation when:

- [x] Marker syntax fully specified with examples
- [x] KNOSSOS_MANIFEST.yaml schema defined
- [x] Sync pipeline architecture complete
- [x] Section structure and ownership documented
- [x] Conflict resolution strategy defined
- [x] Integration points identified
- [x] Edge cases documented with handling
- [x] Test matrix covers all paths
- [x] Migration path specified
- [ ] All ADRs approved
- [ ] Implementation roadmap reviewed

---

## 17. Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-claude-md-inscription.md` | This document |
| Current CLAUDE.md | `/Users/tomtenuta/Code/roster/.claude/CLAUDE.md` | Read |
| Knossos Doctrine | `/Users/tomtenuta/Code/roster/docs/philosophy/knossos-doctrine.md` | Read |
| Existing claudemd.go | `/Users/tomtenuta/Code/roster/ariadne/internal/team/claudemd.go` | Read |
| Ownership model | `/Users/tomtenuta/Code/roster/rites/ecosystem-pack/skills/claude-md-architecture/ownership-model.md` | Read |
| Merge docs | `/Users/tomtenuta/Code/roster/lib/sync/merge/merge-docs.sh` | Read |
| TDD Schema | `/Users/tomtenuta/Code/roster/user-skills/documentation/doc-artifacts/schemas/tdd-schema.md` | Read |
| Hooks config | `/Users/tomtenuta/Code/roster/.claude/hooks/ari/hooks.yaml` | Read |
| Agent definitions | `/Users/tomtenuta/Code/roster/.claude/agents/*.md` | Read |
| TDD Reference | `/Users/tomtenuta/Code/roster/docs/design/TDD-knossos-v2.md` | Read |

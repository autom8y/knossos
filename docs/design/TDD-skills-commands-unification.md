---
artifact_id: TDD-skills-commands-unification
title: "Skills and Commands Unification Technical Design"
created_at: "2026-01-10T14:30:00Z"
author: architect
prd_ref: PRD-skills-commands-unification
status: draft
components:
  - name: CommandFrontmatter
    type: config
    description: "Unified frontmatter schema for all commands (invokable and reference)"
    dependencies:
      - name: yaml
        type: external
  - name: MaterializeCommands
    type: module
    description: "Updated materialization logic for unified command system"
    dependencies:
      - name: internal/materialize
        type: internal
      - name: internal/inscription
        type: internal
  - name: InscriptionTemplates
    type: config
    description: "Updated inscription templates with command path references"
    dependencies:
      - name: knossos/templates
        type: internal
  - name: RiteManifests
    type: config
    description: "Updated rite manifest schema replacing skills with commands"
    dependencies: []
related_adrs:
  - ADR-0015
schema_version: "1.0"
---

# TDD: Skills and Commands Unification

## 1. Overview

### 1.1 Context

Claude Code has merged slash commands and skills into a unified model where both are invoked via the Skill tool. This creates an opportunity to simplify Knossos architecture by eliminating the distinction between `user-skills/` and `user-commands/`, reducing maintenance burden and cognitive overhead.

**Reference**: [SPIKE-skills-commands-unification](/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-skills-commands-unification.md)

### 1.2 Decision Summary

Deprecate `user-skills/` entirely. Everything becomes a command with an `invokable` field:
- `invokable: true` (default): User-callable via `/name`
- `invokable: false`: Reference/library content, not user-callable

### 1.3 Goals

1. Single source of truth: `user-commands/` only
2. Single projection: `.claude/commands/` only
3. Single mental model: "Commands. Some you call, some you read."
4. Preserve progressive disclosure for token efficiency
5. Clean migration with no legacy artifacts

### 1.4 Non-Goals

- Changing Claude Code's Skill tool behavior
- Modifying how Claude Code discovers or invokes commands
- Adding new capabilities beyond unification

---

## 2. Unified Frontmatter Schema

### 2.1 Complete Schema Definition

```yaml
---
# Identity (required for all)
name: string                    # Command identifier (e.g., "start", "session-common")
description: string             # Human-readable description

# Invocation Control
invokable: boolean              # Default: true. User-callable via /name
argument-hint: string           # Only for invokable=true. Usage hint
triggers: array[string]         # Auto-invocation keywords (parsed from description)
allowed-tools: array[string]    # Tool restrictions (only for invokable=true)
model: string                   # Model selection (only for invokable=true)

# Classification (for non-invokable)
category: enum                  # reference | template | schema
                               # Only required when invokable=false

# Optional Metadata
version: string                 # Semantic version for tracking
deprecated: boolean             # Mark command as deprecated
deprecated-by: string           # Reference to replacement command
---
```

### 2.2 Field Semantics

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `name` | Yes | - | Identifier matching filename (without .md) |
| `description` | Yes | - | Human-readable description |
| `invokable` | No | `true` | Whether user can call via `/name` |
| `argument-hint` | No | - | Usage pattern (e.g., `<initiative> [--complexity=LEVEL]`) |
| `triggers` | No | `[]` | Auto-invocation keywords |
| `allowed-tools` | No | - | Tool restrictions for invokable commands |
| `model` | No | - | Model selection for invokable commands |
| `category` | Conditional | - | Required when `invokable: false` |
| `version` | No | - | Semantic version for change tracking |
| `deprecated` | No | `false` | Deprecation flag |
| `deprecated-by` | No | - | Replacement command reference |

### 2.3 Category Values

For non-invokable commands (`invokable: false`), the `category` field classifies content:

| Category | Description | Examples |
|----------|-------------|----------|
| `reference` | Documentation and patterns | `session-common/`, `prompting/` |
| `template` | Reusable document templates | `doc-artifacts/`, `shared-templates/` |
| `schema` | Data structure definitions | Schema files, validation rules |

### 2.4 Example Frontmatter

**Invokable Command** (`session/start.md`):
```yaml
---
name: start
description: Initialize a new work session
invokable: true
argument-hint: <initiative> [--complexity=LEVEL] [--rite=PACK]
allowed-tools: [Bash, Read, Task]
model: opus
triggers: [new session, begin work, start working]
---
```

**Reference Command** (`session/common/INDEX.md`):
```yaml
---
name: session-common
description: Shared session lifecycle schemas and patterns
invokable: false
category: reference
---
```

**Template Command** (`templates/doc-artifacts/INDEX.md`):
```yaml
---
name: doc-artifacts
description: PRD, TDD, ADR, and Test templates for 10x development workflow
invokable: false
category: template
---
```

### 2.5 Backward Compatibility

During migration, the materialization system will:

1. Accept both old and new frontmatter formats
2. Treat missing `invokable` field as `true` (old command behavior)
3. Treat missing `category` field for non-invokable as `reference` (default)
4. Log warnings for deprecated field usage

---

## 3. Directory Structure

### 3.1 New Structure Overview

```
user-commands/                           # SINGLE SOURCE OF TRUTH
├── session/                             # Domain: session lifecycle
│   ├── start.md                        # invokable: true
│   ├── start/                          # Progressive disclosure for start
│   │   ├── behavior.md
│   │   └── examples.md
│   ├── park.md                         # invokable: true
│   ├── park/
│   ├── wrap.md                         # invokable: true
│   ├── wrap/
│   ├── continue.md                     # invokable: true
│   ├── handoff.md                      # invokable: true
│   ├── handoff/
│   └── common/                         # invokable: false (was session-common)
│       ├── INDEX.md                    # Entry point
│       ├── session-context-schema.md
│       ├── session-state-machine.md
│       ├── session-phases.md
│       ├── session-validation.md
│       ├── complexity-levels.md
│       ├── anti-patterns.md
│       ├── error-messages.md
│       └── agent-delegation.md
│
├── operations/                          # Domain: git/build operations
│   ├── commit.md                       # invokable: true
│   ├── commit/                         # Progressive disclosure
│   │   ├── behavior.md
│   │   └── examples.md
│   ├── pr.md                           # invokable: true
│   ├── pr/
│   ├── spike.md                        # invokable: true
│   ├── spike/
│   ├── architect.md                    # invokable: true
│   ├── build.md                        # invokable: true
│   ├── qa.md                           # invokable: true
│   └── code-review.md                  # invokable: true
│
├── workflow/                            # Domain: workflow orchestration
│   ├── task.md                         # invokable: true
│   ├── sprint.md                       # invokable: true
│   └── hotfix.md                       # invokable: true
│       └── hotfix/
│
├── navigation/                          # Domain: navigation and discovery
│   ├── worktree.md                     # invokable: true
│   ├── worktree/
│   ├── sessions.md                     # invokable: true
│   ├── consult.md                      # invokable: true
│   ├── consult/                        # Was consult-ref
│   ├── rite.md                         # invokable: true
│   └── ecosystem.md                    # invokable: true
│
├── meta/                                # Domain: meta operations
│   ├── minus-1.md                      # invokable: true
│   ├── zero.md                         # invokable: true
│   └── one.md                          # invokable: true
│
├── rite-switching/                      # Domain: rite switching
│   ├── 10x.md                          # invokable: true
│   ├── hygiene.md                      # invokable: true
│   ├── debt.md                         # invokable: true
│   ├── forge.md                        # invokable: true
│   ├── docs.md                         # invokable: true
│   ├── intelligence.md                 # invokable: true
│   ├── security.md                     # invokable: true
│   ├── strategy.md                     # invokable: true
│   ├── sre.md                          # invokable: true
│   └── rnd.md                          # invokable: true
│
├── cem/                                 # Domain: CEM operations
│   └── sync.md                         # invokable: true
│
├── guidance/                            # NEW DOMAIN: patterns & reference
│   ├── prompting/                      # invokable: false (was user-skills/guidance/prompting)
│   │   ├── INDEX.md
│   │   ├── patterns/
│   │   │   ├── implementation.md
│   │   │   ├── validation.md
│   │   │   ├── meta-prompts.md
│   │   │   ├── discovery.md
│   │   │   └── maintenance.md
│   │   └── workflows/
│   │       ├── feature-extension.md
│   │       ├── new-feature.md
│   │       ├── legacy-migration.md
│   │       ├── quick-fix.md
│   │       ├── spike-exploration.md
│   │       └── refactoring.md
│   ├── cross-rite/                     # invokable: false
│   │   ├── INDEX.md
│   │   ├── validation.md
│   │   └── routes/
│   ├── file-verification/              # invokable: false
│   │   └── INDEX.md
│   ├── standards/                      # invokable: false
│   │   ├── INDEX.md
│   │   ├── code-conventions.md
│   │   ├── tech-stack-go.md
│   │   ├── tech-stack-python.md
│   │   └── ...
│   └── rite-discovery/                 # invokable: false
│       └── INDEX.md
│
└── templates/                           # NEW DOMAIN: document templates
    ├── doc-artifacts/                   # invokable: false (was user-skills/documentation/doc-artifacts)
    │   ├── INDEX.md
    │   └── schemas/
    │       ├── prd-schema.md
    │       ├── tdd-schema.md
    │       ├── adr-schema.md
    │       └── test-plan-schema.md
    ├── justfile/                        # invokable: false
    │   ├── INDEX.md
    │   ├── structure.md
    │   ├── domains/
    │   └── patterns/
    ├── atuin-desktop/                   # invokable: false
    │   ├── INDEX.md
    │   ├── spec/
    │   ├── validation/
    │   └── agent-guidance.md
    └── shared-templates/                # invokable: false
        ├── INDEX.md
        └── ...
```

### 3.2 Rite Directory Structure

Rite-specific content also transitions from `skills/` to `commands/`:

```
rites/
├── shared/
│   └── commands/                        # Was: skills/
│       ├── cross-rite-handoff/          # invokable: false
│       ├── shared-templates/            # invokable: false
│       └── smell-detection/             # invokable: false
│
├── 10x-dev/
│   └── commands/                        # Was: skills/
│       ├── doc-artifacts/               # invokable: false
│       └── 10x-workflow/                # invokable: false
│
├── forge/
│   └── commands/                        # Was: skills/
│       ├── agent-prompt-engineering/    # invokable: false
│       ├── forge-ref/                   # invokable: false
│       └── rite-development/            # invokable: false
│
├── docs/
│   └── commands/                        # Was: skills/
│       ├── doc-consolidation/           # invokable: false
│       └── doc-reviews/                 # invokable: false
│
├── ecosystem/
│   └── commands/                        # Was: skills/
│       ├── doc-ecosystem/               # invokable: false
│       ├── claude-md-architecture/      # invokable: false
│       └── ecosystem-ref/               # invokable: false
│
└── [other rites]/
    └── commands/                        # Was: skills/
```

### 3.3 Progressive Disclosure Pattern

Commands with associated reference documentation use subdirectories:

```
user-commands/session/
├── start.md                # Main command file (invokable: true)
└── start/                  # Progressive disclosure directory
    ├── behavior.md         # Detailed behavior documentation
    ├── examples.md         # Usage examples
    └── integration.md      # Integration patterns

user-commands/session/
├── common/                 # Reference module (invokable: false)
│   ├── INDEX.md           # Entry point with navigation
│   ├── schema-a.md        # Detail file
│   └── schema-b.md        # Detail file
```

**Discovery Pattern**:
1. Claude loads `start.md` when `/start` invoked
2. If more detail needed, Claude reads `start/behavior.md`
3. Reference modules load `INDEX.md` first, then specific files

---

## 4. Materialization Changes

### 4.1 Current Implementation Analysis

The current `materialize.go` handles skills via `materializeSkills()`:

```go
// Current: materializeSkills copies from rites/*/skills/ to .claude/skills/
func (m *Materializer) materializeSkills(manifest *RiteManifest, claudeDir string) error {
    skillsDir := filepath.Join(claudeDir, "skills")
    // ... copies from multiple sources
}
```

### 4.2 Required Changes to materialize.go

#### 4.2.1 Struct Updates

```go
// RiteManifest - update skills to commands
type RiteManifest struct {
    Name         string   `yaml:"name"`
    Version      string   `yaml:"version"`
    Description  string   `yaml:"description"`
    EntryAgent   string   `yaml:"entry_agent"`
    Agents       []Agent  `yaml:"agents"`
    Commands     []string `yaml:"commands"`     // NEW: replaces Skills
    Skills       []string `yaml:"skills"`       // DEPRECATED: kept for migration
    Hooks        []string `yaml:"hooks"`
    Dependencies []string `yaml:"dependencies"`
}
```

#### 4.2.2 New materializeCommands Function

```go
// materializeCommands copies command files to .claude/commands/
// Sources: user-commands/, rites/{rite}/commands/, rites/shared/commands/
func (m *Materializer) materializeCommands(manifest *RiteManifest, claudeDir string) error {
    commandsDir := filepath.Join(claudeDir, "commands")

    // Remove existing commands directory
    if err := os.RemoveAll(commandsDir); err != nil && !os.IsNotExist(err) {
        return err
    }

    // Create fresh commands directory
    if err := paths.EnsureDir(commandsDir); err != nil {
        return err
    }

    // Priority order for sources (later sources can override earlier)
    sources := []string{
        // 1. User-level commands (lowest priority, can be overridden)
        m.getUserCommandsDir(),
        // 2. Shared rite commands
        filepath.Join(m.ritesDir, "shared", "commands"),
        // 3. Dependency rite commands
    }

    // Add dependency commands
    for _, dep := range manifest.Dependencies {
        if dep != "shared" { // Already added
            sources = append(sources, filepath.Join(m.ritesDir, dep, "commands"))
        }
    }

    // 4. Current rite commands (highest priority)
    sources = append(sources, filepath.Join(m.ritesDir, manifest.Name, "commands"))

    // Copy from all sources
    for _, source := range sources {
        if _, err := os.Stat(source); os.IsNotExist(err) {
            continue
        }
        if err := m.copyDir(source, commandsDir); err != nil {
            return err
        }
    }

    return nil
}

// getUserCommandsDir returns the user-commands directory path
func (m *Materializer) getUserCommandsDir() string {
    // Check for project-level user-commands first
    projectUserCmds := filepath.Join(m.resolver.ProjectRoot(), "user-commands")
    if _, err := os.Stat(projectUserCmds); err == nil {
        return projectUserCmds
    }

    // Fall back to Knossos platform user-commands
    if m.sourceResolver.knossosHome != "" {
        return filepath.Join(m.sourceResolver.knossosHome, "user-commands")
    }

    return ""
}
```

#### 4.2.3 Migration Support

```go
// materializeSkillsCompat provides backward compatibility during migration
// Reads both skills and commands from manifest
func (m *Materializer) materializeSkillsCompat(manifest *RiteManifest, claudeDir string) error {
    // If manifest has only skills (legacy), use old behavior
    if len(manifest.Skills) > 0 && len(manifest.Commands) == 0 {
        return m.materializeSkillsLegacy(manifest, claudeDir)
    }

    // Otherwise use new commands behavior
    return m.materializeCommands(manifest, claudeDir)
}
```

#### 4.2.4 Update MaterializeWithOptions

```go
func (m *Materializer) MaterializeWithOptions(activeRiteName string, opts Options) (*Result, error) {
    // ... existing code ...

    // Step 5: Replace skills with commands
    // OLD: if err := m.materializeSkills(manifest, claudeDir); err != nil {
    // NEW:
    if err := m.materializeCommands(manifest, claudeDir); err != nil {
        return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize commands", err)
    }

    // ... rest of function ...
}
```

### 4.3 Source Resolution Changes

The `SourceResolver` in `source.go` needs updates for command paths:

```go
// checkSource - update to validate commands directory
func (r *SourceResolver) checkSource(riteName string, source RiteSource) (*ResolvedRite, error) {
    ritePath := filepath.Join(source.Path, riteName)
    manifestPath := filepath.Join(ritePath, "manifest.yaml")

    if _, err := os.Stat(manifestPath); err != nil {
        return nil, err
    }

    // Check for commands directory (new) or skills directory (legacy)
    commandsDir := filepath.Join(ritePath, "commands")
    skillsDir := filepath.Join(ritePath, "skills")

    var contentDir string
    if _, err := os.Stat(commandsDir); err == nil {
        contentDir = commandsDir
    } else if _, err := os.Stat(skillsDir); err == nil {
        contentDir = skillsDir // Legacy support
    }

    // ... rest of function ...
}
```

### 4.4 Frontmatter Parsing

New type definitions for command frontmatter:

```go
// CommandFrontmatter represents the unified frontmatter schema
type CommandFrontmatter struct {
    Name         string   `yaml:"name"`
    Description  string   `yaml:"description"`
    Invokable    *bool    `yaml:"invokable,omitempty"`    // Default: true
    ArgumentHint string   `yaml:"argument-hint,omitempty"`
    Triggers     []string `yaml:"triggers,omitempty"`
    AllowedTools []string `yaml:"allowed-tools,omitempty"`
    Model        string   `yaml:"model,omitempty"`
    Category     string   `yaml:"category,omitempty"`     // For non-invokable
    Version      string   `yaml:"version,omitempty"`
    Deprecated   bool     `yaml:"deprecated,omitempty"`
    DeprecatedBy string   `yaml:"deprecated-by,omitempty"`
}

// IsInvokable returns whether the command is user-invokable
func (f *CommandFrontmatter) IsInvokable() bool {
    if f.Invokable == nil {
        return true // Default is invokable
    }
    return *f.Invokable
}

// ParseCommandFrontmatter extracts frontmatter from a command file
func ParseCommandFrontmatter(content []byte) (*CommandFrontmatter, error) {
    // Find frontmatter delimiters
    if !bytes.HasPrefix(content, []byte("---\n")) {
        return nil, errors.New(errors.CodeParseError, "missing frontmatter delimiter")
    }

    endIndex := bytes.Index(content[4:], []byte("\n---"))
    if endIndex == -1 {
        return nil, errors.New(errors.CodeParseError, "missing closing frontmatter delimiter")
    }

    frontmatterBytes := content[4 : 4+endIndex]

    var fm CommandFrontmatter
    if err := yaml.Unmarshal(frontmatterBytes, &fm); err != nil {
        return nil, errors.Wrap(errors.CodeParseError, "invalid frontmatter YAML", err)
    }

    return &fm, nil
}
```

---

## 5. Inscription Updates

### 5.1 Template Changes

#### 5.1.1 skills.md.tpl

**Current** (`knossos/templates/sections/skills.md.tpl`):
```go
{{/* skills section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START skills -->
## Skills

Skills are invoked via the **Skill tool**. Key skills: `orchestration` (workflow coordination), `documentation` (templates), `prompting` (agent invocation), `standards` (conventions), `ecosystem-ref` (roster ecosystem patterns). See `.claude/skills/` and `~/.claude/skills/` for full list.
<!-- KNOSSOS:END skills -->
```

**New** (rename to `commands.md.tpl`):
```go
{{/* commands section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START commands -->
## Commands

Commands are invoked via the **Skill tool**. Two types exist:

- **Invokable** (`/name`): User-callable actions like `/start`, `/commit`, `/pr`
- **Reference** (auto-loaded): Patterns and templates like `prompting`, `doc-artifacts`

Key reference commands: `prompting` (agent patterns), `doc-artifacts` (PRD/TDD/ADR schemas), `standards` (code conventions), `session/common` (lifecycle schemas).

See `.claude/commands/` for the full list.
<!-- KNOSSOS:END commands -->
```

#### 5.1.2 getting-help.md.tpl

Update references from skills to commands:

```go
{{/* getting-help section template */}}
<!-- KNOSSOS:START getting-help -->
## Getting Help

| Question | Command |
|----------|---------|
| Invoke agents | `prompting` |
| Conventions | `standards` |
| Workflow coordination | `orchestration` |
| Unsure where to start | `/consult` |

<!-- KNOSSOS:END getting-help -->
```

### 5.2 Generator Updates

#### 5.2.1 Update Default Content

In `internal/inscription/generator.go`, update the default skills content:

```go
func (g *Generator) getDefaultSkillsContent() string {
    // RENAME to getDefaultCommandsContent
    return `## Commands

Commands are invoked via the **Skill tool**. Two types exist:

- **Invokable** (` + "`/name`" + `): User-callable actions like ` + "`/start`" + `, ` + "`/commit`" + `, ` + "`/pr`" + `
- **Reference** (auto-loaded): Patterns and templates like ` + "`prompting`" + `, ` + "`doc-artifacts`" + `

Key reference commands: ` + "`prompting`" + ` (agent patterns), ` + "`doc-artifacts`" + ` (PRD/TDD/ADR schemas), ` + "`standards`" + ` (code conventions), ` + "`session/common`" + ` (lifecycle schemas).

See ` + "`.claude/commands/`" + ` for the full list.`
}
```

#### 5.2.2 Update Section Names

```go
func (g *Generator) getDefaultSectionContent(regionName string) (string, error) {
    defaults := map[string]string{
        // ... other sections ...
        "commands": g.getDefaultCommandsContent(), // NEW
        "skills":   g.getDefaultCommandsContent(), // ALIAS for backward compat
        // ... other sections ...
    }
    // ...
}
```

### 5.3 Manifest Updates

Update `KNOSSOS_MANIFEST.yaml` default regions:

```go
func (m *ManifestLoader) CreateDefault() (*Manifest, error) {
    // ...
    defaultKnossosRegions := []string{
        "execution-mode",
        "knossos-identity",
        "agent-routing",
        "commands",      // NEW: replaces "skills"
        "hooks",
        "dynamic-context",
        "ariadne-cli",
        "getting-help",
        "state-management",
        "slash-commands",
    }
    // ...
}
```

---

## 6. Rite Manifest Schema Updates

### 6.1 Current Schema

```yaml
# Current rite manifest (rites/10x-dev/manifest.yaml)
name: 10x-dev
version: "1.0.0"
description: Full development lifecycle...

skills:
  - 10x-ref
  - 10x-workflow
  - architect-ref
  - build-ref
  - doc-artifacts

dependencies:
  - shared
```

### 6.2 New Schema

```yaml
# New rite manifest
name: 10x-dev
version: "2.0.0"
description: Full development lifecycle...

commands:
  - 10x-ref           # Reference command (invokable: false)
  - 10x-workflow      # Reference command (invokable: false)
  - architect-ref     # Reference command (invokable: false)
  - build-ref         # Reference command (invokable: false)
  - doc-artifacts     # Template command (invokable: false)

dependencies:
  - shared
```

### 6.3 All Manifests to Update

| Rite | Current Skills | New Commands |
|------|---------------|--------------|
| **10x-dev** | 10x-ref, 10x-workflow, architect-ref, build-ref, doc-artifacts | Same names in commands/ |
| **shared** | cross-rite-handoff, shared-templates, smell-detection | Same names in commands/ |
| **forge** | agent-prompt-engineering, forge-ref, rite-development | Same names in commands/ |
| **docs** | doc-consolidation, doc-reviews | Same names in commands/ |
| **ecosystem** | doc-ecosystem, claude-md-architecture, ecosystem-ref | Same names in commands/ |
| **intelligence** | doc-intelligence | Same name in commands/ |
| **security** | doc-security | Same name in commands/ |
| **sre** | doc-sre | Same name in commands/ |
| **strategy** | doc-strategy | Same name in commands/ |
| **rnd** | doc-rnd | Same name in commands/ |
| **debt-triage** | (none) | (none) |
| **hygiene** | (none) | (none) |

---

## 7. Migration Execution Plan

### 7.1 Phase 1: Create New Structure

**Objective**: Establish the new `user-commands/` structure and new domains.

**Steps**:

1. Create new domains in `user-commands/`:
   ```bash
   mkdir -p user-commands/guidance
   mkdir -p user-commands/templates
   ```

2. Merge `-ref` skills into command progressive disclosure:
   ```bash
   # Example: start-ref -> session/start/
   mv user-skills/session-lifecycle/start-ref/* user-commands/session/start/
   mv user-skills/session-lifecycle/park-ref/* user-commands/session/park/
   mv user-skills/session-lifecycle/wrap-ref/* user-commands/session/wrap/
   mv user-skills/session-lifecycle/handoff-ref/* user-commands/session/handoff/
   mv user-skills/operations/pr-ref/* user-commands/operations/pr/
   mv user-skills/operations/hotfix-ref/* user-commands/workflow/hotfix/
   mv user-skills/guidance/consult-ref/* user-commands/navigation/consult/
   mv user-skills/operations/worktree-ref/* user-commands/navigation/worktree/
   ```

3. Move library skills to new domains:
   ```bash
   # Guidance domain
   mv user-skills/session-lifecycle/session-common user-commands/session/common
   mv user-skills/session-lifecycle/shared-sections user-commands/session/shared
   mv user-skills/guidance/prompting user-commands/guidance/prompting
   mv user-skills/guidance/cross-rite user-commands/guidance/cross-rite
   mv user-skills/guidance/file-verification user-commands/guidance/file-verification
   mv user-skills/documentation/standards user-commands/guidance/standards
   mv user-skills/guidance/rite-discovery user-commands/guidance/rite-discovery

   # Templates domain
   mv user-skills/documentation/doc-artifacts user-commands/templates/doc-artifacts
   mv user-skills/documentation/justfile user-commands/templates/justfile
   mv user-skills/documentation/atuin-desktop user-commands/templates/atuin-desktop
   mv user-skills/documentation/documentation user-commands/templates/documentation
   ```

4. Update frontmatter for moved files:
   - Add `invokable: false` to all reference/template commands
   - Add appropriate `category` field
   - Rename `SKILL.md` files to `INDEX.md`

### 7.2 Phase 2: Update Rite Directories

**Objective**: Rename `skills/` to `commands/` in all rites.

**Script** (`scripts/migrate-rite-skills.sh`):
```bash
#!/bin/bash
set -euo pipefail

RITES_DIR="${1:-rites}"

for rite_dir in "$RITES_DIR"/*/; do
    rite_name=$(basename "$rite_dir")
    skills_dir="$rite_dir/skills"
    commands_dir="$rite_dir/commands"

    if [ -d "$skills_dir" ]; then
        echo "Migrating $rite_name: skills/ -> commands/"
        mv "$skills_dir" "$commands_dir"

        # Rename SKILL.md to INDEX.md
        find "$commands_dir" -name "SKILL.md" -exec \
            sh -c 'mv "$1" "$(dirname "$1")/INDEX.md"' _ {} \;

        # Add invokable: false to all INDEX.md files
        find "$commands_dir" -name "INDEX.md" -exec \
            sed -i '' '1a\
invokable: false
' {} \;
    fi
done

echo "Migration complete"
```

### 7.3 Phase 3: Update Manifests

**Objective**: Update all manifest.yaml files to use `commands:` instead of `skills:`.

**Script** (`scripts/migrate-manifests.sh`):
```bash
#!/bin/bash
set -euo pipefail

RITES_DIR="${1:-rites}"

for manifest in "$RITES_DIR"/*/manifest.yaml; do
    echo "Updating: $manifest"

    # Replace 'skills:' with 'commands:'
    sed -i '' 's/^skills:/commands:/' "$manifest"

    # Update version to 2.0.0 to indicate migration
    sed -i '' 's/version: "1.0.0"/version: "2.0.0"/' "$manifest"
done

echo "Manifest migration complete"
```

### 7.4 Phase 4: Update Materialization Code

**Objective**: Update `materialize.go` to use new paths.

**Changes**:
1. Rename `materializeSkills` to `materializeCommands`
2. Update source paths from `skills/` to `commands/`
3. Update projection path from `.claude/skills/` to `.claude/commands/`
4. Add backward compatibility for legacy manifests

### 7.5 Phase 5: Update Inscription

**Objective**: Update templates and generator code.

**Changes**:
1. Rename `skills.md.tpl` to `commands.md.tpl`
2. Update content references
3. Update `generator.go` default content
4. Update manifest default regions

### 7.6 Phase 6: Delete Legacy

**Objective**: Remove all legacy directories and files.

**Steps**:
```bash
# Remove user-skills (after verification)
rm -rf user-skills/

# Remove any .claude/skills/ projections
rm -rf .claude/skills/

# Clean up any orphaned SKILL.md files
find . -name "SKILL.md" -delete
```

### 7.7 Verification Checklist

Before deletion, verify:

- [ ] All `user-commands/` files have valid frontmatter
- [ ] All rite `commands/` directories populated correctly
- [ ] All `manifest.yaml` files updated to use `commands:`
- [ ] `ari sync materialize` works with new structure
- [ ] `.claude/commands/` generated correctly
- [ ] No references to `.claude/skills/` in templates
- [ ] CLAUDE.md renders correctly with new regions
- [ ] All existing commands still invokable
- [ ] All reference content discoverable

---

## 8. File Verification Protocol

### 8.1 Verification Requirements

All changes must be verified via explicit Read tool confirmation:

| File Type | Verification |
|-----------|-------------|
| New commands | Read and confirm frontmatter valid |
| Moved files | Read source, read destination, confirm match |
| Updated manifests | Read and confirm `commands:` present |
| Templates | Read and confirm path updates |

### 8.2 Attestation Table

After migration, generate attestation:

```markdown
## Migration Attestation

| Path | Action | Verified |
|------|--------|----------|
| `/Users/tomtenuta/Code/roster/user-commands/session/common/INDEX.md` | Created | Yes |
| `/Users/tomtenuta/Code/roster/rites/10x-dev/commands/` | Renamed from skills/ | Yes |
| `/Users/tomtenuta/Code/roster/rites/10x-dev/manifest.yaml` | Updated commands field | Yes |
| ... | ... | ... |
```

---

## 9. Risk Assessment

### 9.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing commands | Low | High | Backward compat in frontmatter parsing |
| Missing content after migration | Medium | High | Verification checklist, pre-migration inventory |
| Template rendering failures | Low | Medium | Test templates before deployment |
| Manifest parsing errors | Low | Medium | Schema validation, gradual rollout |

### 9.2 Rollback Plan

If critical issues discovered:

1. **Immediate**: Restore from git (all changes tracked)
2. **Graceful**: Keep both `skills/` and `commands/` during transition
3. **Full rollback**: Revert commits, regenerate from backup

---

## 10. Testing Plan

### 10.1 Unit Tests

```go
func TestCommandFrontmatterParsing(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        expected *CommandFrontmatter
        wantErr  bool
    }{
        {
            name: "invokable command",
            content: `---
name: start
description: Initialize session
invokable: true
argument-hint: <initiative>
---
Content here`,
            expected: &CommandFrontmatter{
                Name:         "start",
                Description:  "Initialize session",
                Invokable:    ptrBool(true),
                ArgumentHint: "<initiative>",
            },
        },
        {
            name: "reference command",
            content: `---
name: session-common
description: Shared schemas
invokable: false
category: reference
---
Content here`,
            expected: &CommandFrontmatter{
                Name:        "session-common",
                Description: "Shared schemas",
                Invokable:   ptrBool(false),
                Category:    "reference",
            },
        },
        {
            name: "default invokable",
            content: `---
name: commit
description: Create git commit
---
Content`,
            expected: &CommandFrontmatter{
                Name:        "commit",
                Description: "Create git commit",
                Invokable:   nil, // IsInvokable() returns true
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseCommandFrontmatter([]byte(tt.content))
            // ... assertions
        })
    }
}
```

### 10.2 Integration Tests

1. **Materialization Test**: Run `ari sync materialize --rite 10x-dev` and verify `.claude/commands/` populated
2. **Inscription Test**: Verify CLAUDE.md renders with `commands` section
3. **Command Invocation Test**: Verify `/start` still invokable after migration

### 10.3 Manual Verification

1. Start fresh Claude session
2. Run `/start "test initiative"`
3. Verify session creates correctly
4. Run `/consult` and verify reference loaded
5. Check `.claude/commands/` structure matches spec

---

## 11. Implementation Sequence

### 11.1 Recommended Order

1. **Week 1**: Schema and Code
   - Define `CommandFrontmatter` type
   - Implement `ParseCommandFrontmatter()`
   - Update `materialize.go` with dual support
   - Write unit tests

2. **Week 2**: Migration Scripts
   - Create migration scripts
   - Test on copy of repository
   - Document edge cases

3. **Week 3**: Content Migration
   - Execute Phase 1-3 (structure, rites, manifests)
   - Run verification checklist
   - Keep `user-skills/` as backup

4. **Week 4**: Finalization
   - Execute Phase 4-5 (code, inscription)
   - Full integration testing
   - Execute Phase 6 (delete legacy)
   - Update documentation

### 11.2 Dependencies

```
Schema Definition
       │
       ▼
Code Changes (materialize.go)
       │
       ├──────────────────┐
       ▼                  ▼
Content Migration    Inscription Updates
       │                  │
       └────────┬─────────┘
                ▼
        Legacy Deletion
                │
                ▼
          Documentation
```

---

## 12. Success Criteria

The migration is complete when:

- [ ] `user-skills/` directory deleted
- [ ] All rites have `commands/` not `skills/`
- [ ] All manifests use `commands:` field
- [ ] `.claude/commands/` generated (not `.claude/skills/`)
- [ ] CLAUDE.md shows `## Commands` section
- [ ] All existing slash commands functional
- [ ] All reference content accessible
- [ ] No errors in `ari sync materialize`
- [ ] Documentation updated

---

## Appendix A: Full Migration Inventory

### A.1 Commands to Keep (invokable: true)

| Domain | Commands |
|--------|----------|
| session | start, park, wrap, continue, handoff |
| operations | commit, pr, spike, architect, build, qa, code-review |
| workflow | task, sprint, hotfix |
| navigation | worktree, sessions, consult, rite, ecosystem |
| meta | minus-1, zero, one |
| rite-switching | 10x, hygiene, debt, forge, docs, intelligence, security, strategy, sre, rnd |
| cem | sync |

### A.2 Skills to Convert (invokable: false)

| Current Location | New Location | Category |
|-----------------|--------------|----------|
| `user-skills/session-lifecycle/start-ref/` | `user-commands/session/start/` | reference |
| `user-skills/session-lifecycle/park-ref/` | `user-commands/session/park/` | reference |
| `user-skills/session-lifecycle/wrap-ref/` | `user-commands/session/wrap/` | reference |
| `user-skills/session-lifecycle/handoff-ref/` | `user-commands/session/handoff/` | reference |
| `user-skills/session-lifecycle/resume/` | `user-commands/session/continue/` | reference |
| `user-skills/session-lifecycle/session-common/` | `user-commands/session/common/` | reference |
| `user-skills/session-lifecycle/shared-sections/` | `user-commands/session/shared/` | reference |
| `user-skills/operations/pr-ref/` | `user-commands/operations/pr/` | reference |
| `user-skills/operations/hotfix-ref/` | `user-commands/workflow/hotfix/` | reference |
| `user-skills/operations/worktree-ref/` | `user-commands/navigation/worktree/` | reference |
| `user-skills/operations/review/` | `user-commands/operations/code-review/` | reference |
| `user-skills/operations/qa-ref/` | `user-commands/operations/qa/` | reference |
| `user-skills/guidance/prompting/` | `user-commands/guidance/prompting/` | reference |
| `user-skills/guidance/consult-ref/` | `user-commands/navigation/consult/` | reference |
| `user-skills/guidance/cross-rite/` | `user-commands/guidance/cross-rite/` | reference |
| `user-skills/guidance/file-verification/` | `user-commands/guidance/file-verification/` | reference |
| `user-skills/guidance/rite-discovery/` | `user-commands/guidance/rite-discovery/` | reference |
| `user-skills/documentation/doc-artifacts/` | `user-commands/templates/doc-artifacts/` | template |
| `user-skills/documentation/justfile/` | `user-commands/templates/justfile/` | template |
| `user-skills/documentation/atuin-desktop/` | `user-commands/templates/atuin-desktop/` | template |
| `user-skills/documentation/documentation/` | `user-commands/templates/documentation/` | reference |
| `user-skills/documentation/standards/` | `user-commands/guidance/standards/` | reference |

### A.3 Rite-Specific Skills

| Rite | Skills to Move | New Location |
|------|---------------|--------------|
| 10x-dev | 10x-workflow, doc-artifacts | `rites/10x-dev/commands/` |
| shared | cross-rite-handoff, shared-templates, smell-detection | `rites/shared/commands/` |
| forge | agent-prompt-engineering, forge-ref, rite-development | `rites/forge/commands/` |
| docs | doc-consolidation, doc-reviews | `rites/docs/commands/` |
| ecosystem | doc-ecosystem, claude-md-architecture, ecosystem-ref | `rites/ecosystem/commands/` |
| intelligence | doc-intelligence | `rites/intelligence/commands/` |
| security | doc-security | `rites/security/commands/` |
| sre | doc-sre | `rites/sre/commands/` |
| strategy | doc-strategy | `rites/strategy/commands/` |
| rnd | doc-rnd | `rites/rnd/commands/` |

---

## Appendix B: ADR Reference

This TDD informs the following Architecture Decision Record:

**ADR-0021: Skills and Commands Unification**

- **Status**: Proposed
- **Context**: Claude Code merged slash commands and skills
- **Decision**: Deprecate `user-skills/`, unify into `user-commands/`
- **Consequences**: Simplified mental model, reduced maintenance, single source of truth

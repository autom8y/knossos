# TDD: invoke-rite Partial Composition System

**Status**: Draft
**Author**: Architect Agent
**Date**: 2026-01-06
**PRD Reference**: Knossos Doctrine Section IV (Rite Operations)

---

## 1. Overview

### 1.1 Problem Statement

The current rite switching mechanism (`ari team switch`) operates as an all-or-nothing replacement. When switching teams, the entire agent set and workflow configuration are replaced. This creates friction when practitioners need to temporarily borrow capabilities from another rite without abandoning their current practice.

Per the Knossos Doctrine Section IV:
> "You can invoke a documentation rite while still practicing under the quality rite--borrowing useful knowledge without conversion. Context switching is expensive; knowledge sharing is cheap."

### 1.2 Solution Summary

Implement a partial composition system that allows:
1. **invoke-rite**: Additive borrowing of specific components (agents, skills, or both)
2. **swap-rite**: Full context switch (alias to existing `ari team switch`)
3. **current-rite**: Display active rite and any borrowed components

The key insight is that `invoke-rite` is **additive** while `swap-rite` is **replacement**.

### 1.3 Design Goals

| Goal | Description |
|------|-------------|
| **Composability** | Borrow components without full context switch |
| **Reversibility** | Cleanup borrowed components on session end |
| **Transparency** | Clear visibility into what's borrowed vs native |
| **Budget-Awareness** | Quantified context cost for each operation |
| **Backward Compatibility** | `ari team switch` continues to work as alias |

---

## 2. Rite Manifest Schema

### 2.1 Schema Definition

A rite is defined by a `rite.yaml` manifest file in its directory:

```yaml
# rite.yaml - Rite Manifest Schema v1.0
schema_version: "1.0"

# Core identity
name: 10x-dev-rite                    # Required: kebab-case identifier
display_name: "10x Development Rite"  # Optional: human-readable name
description: |
  Full development lifecycle (PRD -> TDD -> Code -> QA)

# Rite form classification (per doctrine)
form: full                            # simple | practitioner | procedural | full
# - simple: skills only, no agents
# - practitioner: agents + skills
# - procedural: hooks + workflows, no dedicated agents
# - full: all components

# Component references
agents:
  - name: requirements-analyst
    file: agents/requirements-analyst.md
    role: "Extracts stakeholder needs"
    produces: prd

  - name: architect
    file: agents/architect.md
    role: "Evaluates tradeoffs and designs systems"
    produces: tdd

  - name: principal-engineer
    file: agents/principal-engineer.md
    role: "Transforms designs into production code"
    produces: code

  - name: qa-adversary
    file: agents/qa-adversary.md
    role: "Breaks things so users don't"
    produces: test-plan

skills:
  # Local skills bundled with this rite
  - ref: 10x-workflow
    path: skills/10x-workflow/

  - ref: doc-artifacts
    path: skills/doc-artifacts/

  # External skill references (from .claude/skills/)
  - ref: standards
    external: true

# Optional workflow configuration
workflow:
  type: sequential                    # sequential | parallel | custom
  entry_point: requirements-analyst
  phases:
    - name: requirements
      agent: requirements-analyst
      produces: prd
      next: design

    - name: design
      agent: architect
      produces: tdd
      next: implementation

    - name: implementation
      agent: principal-engineer
      produces: code
      next: validation

    - name: validation
      agent: qa-adversary
      produces: test-plan
      next: null

# Optional lifecycle hooks
hooks:
  on_invoke:
    - action: inject_skills
      target: CLAUDE.md

  on_release:
    - action: cleanup_injected
      target: CLAUDE.md

# Context budget metadata
budget:
  estimated_tokens: 12500            # Estimated context consumption
  agents_cost: 8000                  # Agent prompts total
  skills_cost: 3500                  # Skill content total
  workflow_cost: 1000                # Workflow config overhead

# Migration metadata (for transition from rite)
migration:
  from_team: 10x-dev           # Original rite name
  migrated_at: null                 # Populated on migration
```

### 2.2 Schema Validation Rules

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `schema_version` | string | Yes | Semver format (N.N) |
| `name` | string | Yes | kebab-case, unique |
| `form` | enum | Yes | simple, practitioner, procedural, full |
| `agents` | array | Depends | Required if form != simple |
| `skills` | array | No | Valid skill references |
| `workflow` | object | No | Valid workflow structure |
| `budget` | object | No | Positive integers |

### 2.3 Rite Forms Explained

| Form | Agents | Skills | Hooks | Workflow | Example |
|------|--------|--------|-------|----------|---------|
| **simple** | No | Yes | No | No | documentation-rite |
| **practitioner** | Yes | Yes | No | Optional | code-review-rite |
| **procedural** | No | Optional | Yes | Yes | release-rite |
| **full** | Yes | Yes | Optional | Yes | 10x-dev-rite |

---

## 3. CLI Command Interface

### 3.1 Command Tree

```
ari rite
  |-- invoke <name> [component]    # Borrow from another rite
  |-- release <name>               # Release borrowed components
  |-- current                      # Show active rite + borrowed
  |-- list                         # List available rites
  |-- info <name>                  # Show rite details
  |-- swap <name>                  # Full context switch (alias)
```

### 3.2 Command Specifications

#### 3.2.1 `ari rite invoke`

Additively borrows components from another rite without switching context.

```
Usage: ari rite invoke <name> [component] [flags]

Arguments:
  name        Rite name to invoke from
  component   Optional: "skills", "agents", or omit for all

Flags:
  --dry-run           Preview injection without applying
  --no-inscription    Skip CLAUDE.md updates
  -o, --output        Output format: text, json, yaml

Examples:
  ari rite invoke documentation                  # Borrow entire rite
  ari rite invoke documentation skills           # Borrow skills only
  ari rite invoke security-rite agents           # Borrow agents only
  ari rite invoke code-review agents --dry-run   # Preview changes
```

**Output (JSON)**:
```json
{
  "invoked_rite": "documentation",
  "component": "skills",
  "borrowed": {
    "skills": ["doc-standards", "templates"],
    "agents": []
  },
  "inscription_updated": true,
  "estimated_tokens": 2500,
  "invocation_id": "inv-20260106-abc123"
}
```

#### 3.2.2 `ari rite release`

Releases borrowed components from a previous invocation.

```
Usage: ari rite release <name|invocation-id> [flags]

Arguments:
  name              Rite name to release
  invocation-id     Specific invocation ID to release

Flags:
  --all             Release all borrowed components
  --dry-run         Preview cleanup without applying
  -o, --output      Output format: text, json, yaml

Examples:
  ari rite release documentation           # Release specific rite
  ari rite release --all                   # Release everything borrowed
  ari rite release inv-20260106-abc123     # Release by invocation ID
```

#### 3.2.3 `ari rite current`

Displays the active rite and any borrowed components.

```
Usage: ari rite current [flags]

Flags:
  --borrowed        Show only borrowed components
  --native          Show only native components
  -o, --output      Output format: text, json, yaml

Examples:
  ari rite current
  ari rite current --borrowed
```

**Output (text)**:
```
Active Rite: 10x-dev-rite

Native Components:
  Agents: requirements-analyst, architect, principal-engineer, qa-adversary
  Skills: 10x-workflow, doc-artifacts

Borrowed Components:
  From documentation-rite (inv-20260106-abc123):
    Skills: doc-standards, templates
  From security-rite (inv-20260106-def456):
    Agents: threat-modeler

Total Context Budget: ~15,000 tokens
  Native: 12,500 tokens
  Borrowed: 2,500 tokens
```

#### 3.2.4 `ari rite list`

Lists available rites with their forms and descriptions.

```
Usage: ari rite list [flags]

Flags:
  --form <type>     Filter by form (simple, practitioner, procedural, full)
  --project         Show project rites only
  --user            Show user rites only
  -o, --output      Output format: text, json, yaml

Examples:
  ari rite list
  ari rite list --form=simple
```

**Output (text)**:
```
Available Rites:

Project Rites:
  * 10x-dev-rite       [full]         Full development lifecycle
    debt-triage-rite   [practitioner] Technical debt remediation
    forge-rite         [full]         Agent and tool creation

User Rites:
    documentation-rite [simple]       Knowledge crystallization
    code-review-rite   [practitioner] Code review workflow

* = Currently active
```

#### 3.2.5 `ari rite info`

Displays detailed information about a rite.

```
Usage: ari rite info <name> [flags]

Arguments:
  name        Rite name to inspect

Flags:
  --budget          Show detailed budget breakdown
  --components      Show component list only
  -o, --output      Output format: text, json, yaml

Examples:
  ari rite info 10x-dev-rite
  ari rite info security-rite --budget
```

#### 3.2.6 `ari rite swap`

Performs a full context switch (alias to `ari team switch` for backward compatibility).

```
Usage: ari rite swap <name> [flags]

Arguments:
  name        Target rite name

Flags:
  --remove-all      Remove orphaned agents
  --keep-all        Keep orphaned agents
  --promote-all     Promote orphans to project level
  --dry-run         Preview changes without applying
  --no-sync         Skip inscription sync
  -o, --output      Output format: text, json, yaml

Examples:
  ari rite swap security-rite
  ari rite swap 10x-dev-rite --remove-all
```

---

## 4. CLAUDE.md Injection Algorithm

### 4.1 Architecture Overview

The injection system operates on marked regions within CLAUDE.md using the existing inscription infrastructure. Borrowed components are injected into a new `<!-- KNOSSOS:START borrowed-components -->` region.

```
CLAUDE.md Structure:
+------------------------------------------+
| Header                                    |
+------------------------------------------+
| <!-- KNOSSOS:START quick-start -->        |
| Native agent table                        |
| <!-- KNOSSOS:END quick-start -->          |
+------------------------------------------+
| <!-- KNOSSOS:START borrowed-components -->|  <-- NEW
| Injected borrowed content                 |
| <!-- KNOSSOS:END borrowed-components -->  |
+------------------------------------------+
| <!-- KNOSSOS:START agent-configurations -->|
| Native + borrowed agent references        |
| <!-- KNOSSOS:END agent-configurations --> |
+------------------------------------------+
| ... rest of file                          |
+------------------------------------------+
```

### 4.2 Injection Algorithm

```go
// InvokeOptions configures the invoke operation
type InvokeOptions struct {
    TargetRite    string         // Rite to invoke from
    Component     string         // "skills", "agents", or "" for all
    DryRun        bool           // Preview only
    NoInscription bool           // Skip CLAUDE.md updates
}

// InvokeResult contains the result of an invoke operation
type InvokeResult struct {
    InvokedRite       string            `json:"invoked_rite"`
    Component         string            `json:"component"`
    InvocationID      string            `json:"invocation_id"`
    BorrowedSkills    []string          `json:"borrowed_skills"`
    BorrowedAgents    []AgentRef        `json:"borrowed_agents"`
    InscriptionUpdated bool             `json:"inscription_updated"`
    EstimatedTokens   int               `json:"estimated_tokens"`
}

func (i *Invoker) Invoke(opts InvokeOptions) (*InvokeResult, error) {
    // 1. Load target rite manifest
    targetRite, err := i.discovery.Get(opts.TargetRite)
    if err != nil {
        return nil, err
    }

    // 2. Load current rite state
    currentState, err := i.loadInvocationState()
    if err != nil {
        return nil, err
    }

    // 3. Check for conflicts (same agent borrowed from different rite)
    if conflicts := i.detectConflicts(currentState, targetRite, opts.Component); len(conflicts) > 0 {
        return nil, ErrBorrowConflict(conflicts)
    }

    // 4. Generate invocation ID
    invocationID := generateInvocationID()

    // 5. Determine what to borrow based on component filter
    borrowed := i.selectComponents(targetRite, opts.Component)

    // 6. Copy borrowed agents to .claude/agents/ with tracking
    if len(borrowed.Agents) > 0 {
        if err := i.installBorrowedAgents(borrowed.Agents, targetRite.Name, invocationID); err != nil {
            return nil, err
        }
    }

    // 7. Update CLAUDE.md with borrowed content
    if !opts.NoInscription {
        if err := i.updateInscription(borrowed, targetRite.Name, invocationID); err != nil {
            // Rollback agent installation
            i.rollbackAgentInstall(borrowed.Agents)
            return nil, err
        }
    }

    // 8. Update invocation state
    currentState.AddInvocation(Invocation{
        ID:          invocationID,
        RiteName:    opts.TargetRite,
        Component:   opts.Component,
        Skills:      borrowed.SkillRefs,
        Agents:      borrowed.AgentNames,
        InvokedAt:   time.Now(),
    })
    if err := i.saveInvocationState(currentState); err != nil {
        return nil, err
    }

    return &InvokeResult{
        InvokedRite:       opts.TargetRite,
        Component:         opts.Component,
        InvocationID:      invocationID,
        BorrowedSkills:    borrowed.SkillRefs,
        BorrowedAgents:    borrowed.Agents,
        InscriptionUpdated: !opts.NoInscription,
        EstimatedTokens:   borrowed.EstimatedTokens,
    }, nil
}
```

### 4.3 Borrowed Components Region Format

The borrowed components region is injected after the `quick-start` region:

```markdown
<!-- KNOSSOS:START borrowed-components regenerate=true source=INVOCATION_STATE -->
## Borrowed Components

**Active Invocations:**

From **documentation-rite** (inv-20260106-abc123):
- Skills: `doc-standards`, `templates`

From **security-rite** (inv-20260106-def456):
- Agents: `threat-modeler`

*Release with: `ari rite release <name>` or `ari rite release --all`*
<!-- KNOSSOS:END borrowed-components -->
```

### 4.4 Agent Installation for Borrowed Agents

Borrowed agents are installed to `.claude/agents/` with tracking metadata:

```go
func (i *Invoker) installBorrowedAgents(agents []AgentRef, riteName, invocationID string) error {
    for _, agent := range agents {
        // Copy agent file
        srcPath := filepath.Join(riteDir, agent.File)
        dstPath := filepath.Join(agentsDir, agent.File)

        if err := copyFile(srcPath, dstPath); err != nil {
            return err
        }

        // Update manifest with borrowed source
        checksum, _ := ComputeChecksum(dstPath)
        manifest.AddBorrowedAgent(agent.Name+".md", BorrowedAgentEntry{
            Source:       "borrowed",
            Origin:       riteName,
            InvocationID: invocationID,
            Checksum:     checksum,
            InstalledAt:  time.Now(),
        })
    }
    return nil
}
```

### 4.5 Skill Reference Injection

Skills are not copied; instead, skill references are injected into CLAUDE.md:

```markdown
<!-- In the Skills section -->
## Skills

Skills are invoked via the **Skill tool**. Key skills: `orchestration` (workflow coordination), `documentation` (templates), `prompting` (agent invocation), `standards` (conventions).

**Borrowed Skills** (from documentation-rite):
- `doc-standards` - Documentation standards and templates
- `templates` - Reusable document templates

See `.claude/skills/` and `~/.claude/skills/` for full list.
```

---

## 5. State Tracking for Active Invocations

### 5.1 Invocation State File

State is tracked in `.knossos/INVOCATION_STATE.yaml`:

```yaml
# INVOCATION_STATE.yaml
schema_version: "1.0"
current_rite: 10x-dev-rite
last_updated: 2026-01-06T14:30:00Z

invocations:
  - id: inv-20260106-abc123
    rite_name: documentation-rite
    component: skills
    skills:
      - doc-standards
      - templates
    agents: []
    invoked_at: 2026-01-06T10:00:00Z
    expires_at: null              # null = session lifetime

  - id: inv-20260106-def456
    rite_name: security-rite
    component: agents
    skills: []
    agents:
      - name: threat-modeler
        file: threat-modeler.md
    invoked_at: 2026-01-06T12:00:00Z
    expires_at: null

# Budget tracking
budget:
  native_tokens: 12500
  borrowed_tokens: 2500
  total_tokens: 15000
  budget_limit: 50000             # From ARIADNE_CONTEXT_LIMIT
```

### 5.2 State Operations

```go
type InvocationState struct {
    SchemaVersion string           `yaml:"schema_version"`
    CurrentRite   string           `yaml:"current_rite"`
    LastUpdated   time.Time        `yaml:"last_updated"`
    Invocations   []Invocation     `yaml:"invocations"`
    Budget        BudgetInfo       `yaml:"budget"`
}

type Invocation struct {
    ID          string       `yaml:"id"`
    RiteName    string       `yaml:"rite_name"`
    Component   string       `yaml:"component"`
    Skills      []string     `yaml:"skills"`
    Agents      []AgentRef   `yaml:"agents"`
    InvokedAt   time.Time    `yaml:"invoked_at"`
    ExpiresAt   *time.Time   `yaml:"expires_at,omitempty"`
}

func (s *InvocationState) AddInvocation(inv Invocation) {
    s.Invocations = append(s.Invocations, inv)
    s.LastUpdated = time.Now()
}

func (s *InvocationState) RemoveInvocation(id string) *Invocation {
    for i, inv := range s.Invocations {
        if inv.ID == id {
            removed := s.Invocations[i]
            s.Invocations = append(s.Invocations[:i], s.Invocations[i+1:]...)
            s.LastUpdated = time.Now()
            return &removed
        }
    }
    return nil
}

func (s *InvocationState) FindByRite(riteName string) []Invocation {
    var result []Invocation
    for _, inv := range s.Invocations {
        if inv.RiteName == riteName {
            result = append(result, inv)
        }
    }
    return result
}
```

### 5.3 Session Lifecycle Integration

The invocation state integrates with session lifecycle:

```go
// On session wrap (Atropos)
func (a *Atropos) WrapSession(sessionID string) error {
    // ... existing wrap logic ...

    // Clean up all invocations
    state, _ := LoadInvocationState()
    if state != nil && len(state.Invocations) > 0 {
        invoker := NewInvoker(resolver)
        invoker.ReleaseAll()
    }

    return nil
}

// On session park (Lachesis)
func (l *Lachesis) ParkSession(sessionID string, reason string) error {
    // ... existing park logic ...

    // Preserve invocations in park state
    state, _ := LoadInvocationState()
    parkState.InvocationSnapshot = state

    return nil
}

// On session resume
func (c *Clotho) ResumeSession(sessionID string) error {
    // ... existing resume logic ...

    // Restore invocations from park state
    if parkState.InvocationSnapshot != nil {
        SaveInvocationState(parkState.InvocationSnapshot)
    }

    return nil
}
```

---

## 6. Cost Model

### 6.1 Token Estimation

Context cost is estimated based on component sizes:

| Component Type | Estimation Method | Typical Range |
|----------------|-------------------|---------------|
| Agent | File size / 4 | 1,500 - 3,000 tokens |
| Skill | Directory size / 4 | 500 - 2,000 tokens |
| Workflow | YAML size / 4 | 200 - 500 tokens |

### 6.2 Cost Comparison

| Operation | Context Impact | Estimated Tokens |
|-----------|----------------|------------------|
| `swap-rite` | Full replacement | ~12,000 (typical rite) |
| `invoke-rite <name>` | Additive | +2,000 - +5,000 |
| `invoke-rite <name> skills` | Additive (lighter) | +500 - +2,000 |
| `invoke-rite <name> agents` | Additive | +1,500 - +3,000 |
| `release-rite` | Subtractive | Frees borrowed tokens |

### 6.3 Budget Enforcement

```go
func (i *Invoker) checkBudget(borrowed *BorrowedComponents) error {
    state, _ := i.loadInvocationState()

    newTotal := state.Budget.TotalTokens + borrowed.EstimatedTokens
    limit := i.getBudgetLimit() // From ARIADNE_CONTEXT_LIMIT or default

    if newTotal > limit {
        return ErrBudgetExceeded{
            Current:   state.Budget.TotalTokens,
            Requested: borrowed.EstimatedTokens,
            Limit:     limit,
        }
    }
    return nil
}
```

### 6.4 Warning Thresholds

```go
const (
    BudgetWarnPercent = 0.75  // Warn at 75% budget
    BudgetHardLimit   = 50000 // Default hard limit
)

func (i *Invoker) checkBudgetWarnings(state *InvocationState) []string {
    var warnings []string

    ratio := float64(state.Budget.TotalTokens) / float64(state.Budget.BudgetLimit)
    if ratio >= BudgetWarnPercent {
        warnings = append(warnings, fmt.Sprintf(
            "Context budget at %.0f%% (%d/%d tokens). Consider releasing unused invocations.",
            ratio*100,
            state.Budget.TotalTokens,
            state.Budget.BudgetLimit,
        ))
    }

    return warnings
}
```

---

## 7. Migration Path from Team System

### 7.1 Compatibility Layer

The existing `ari team` commands continue to work:

```go
// In root.go
func init() {
    // Existing team command (preserved for compatibility)
    rootCmd.AddCommand(team.NewTeamCmd(...))

    // New rite command
    rootCmd.AddCommand(rite.NewRiteCmd(...))
}

// ari team switch becomes alias to ari rite swap
// ari team list becomes alias to ari rite list
// ari team status becomes alias to ari rite current
```

### 7.2 Team Pack to Rite Migration

Migration script generates `rite.yaml` from existing rite structure:

```go
func MigrateLegacyRite(teamPath string) (*RiteManifest, error) {
    // Load existing workflow.yaml
    workflow, err := team.LoadWorkflow(filepath.Join(teamPath, "workflow.yaml"))
    if err != nil {
        return nil, err
    }

    // Scan agents/ directory
    agents, err := scanAgents(filepath.Join(teamPath, "agents"))
    if err != nil {
        return nil, err
    }

    // Scan skills/ directory
    skills, err := scanSkills(filepath.Join(teamPath, "skills"))
    if err != nil {
        return nil, err
    }

    // Determine rite form
    form := determineForm(agents, skills, workflow)

    // Build rite manifest
    manifest := &RiteManifest{
        SchemaVersion: "1.0",
        Name:          toRiteName(workflow.Name), // 10x-dev -> 10x-dev-rite
        DisplayName:   workflow.Description,
        Description:   workflow.Description,
        Form:          form,
        Agents:        agents,
        Skills:        skills,
        Migration: &MigrationInfo{
            FromTeam:   workflow.Name,
            MigratedAt: time.Now(),
        },
    }

    // Calculate budget
    manifest.Budget = calculateBudget(agents, skills, workflow)

    return manifest, nil
}

func toRiteName(teamName string) string {
    // 10x-dev -> 10x-dev-rite
    return strings.Replace(teamName, "-pack", "-rite", 1)
}
```

### 7.3 Migration Command

```
Usage: ari rite migrate [flags]

Flags:
  --all             Migrate all rites
  --team <name>     Migrate specific rite
  --dry-run         Preview migration without writing
  --in-place        Modify existing rite directory
  -o, --output      Output format: text, json, yaml

Examples:
  ari rite migrate --all --dry-run
  ari rite migrate --team 10x-dev
```

### 7.4 Gradual Migration Strategy

| Phase | Action | Timeline |
|-------|--------|----------|
| **Phase 1** | Add `ari rite` commands alongside `ari team` | Immediate |
| **Phase 2** | Generate `rite.yaml` for existing rites | Week 1 |
| **Phase 3** | Deprecation warnings on `ari team` commands | Week 2 |
| **Phase 4** | Documentation updates to use rite terminology | Week 3 |
| **Phase 5** | Remove `ari team` commands (major version) | v3.0 |

---

## 8. Integration Points

### 8.1 Existing Ariadne Systems

| System | Integration |
|--------|-------------|
| **inscription** | New `borrowed-components` region |
| **team/switch** | `ari rite swap` wraps existing logic |
| **team/discovery** | Rite discovery reuses patterns |
| **session/lifecycle** | Moirai integration for cleanup |
| **hook/context** | Invocation state in context output |
| **paths/resolver** | New paths for invocation state |

### 8.2 New Package Structure

```
ariadne/internal/
  rite/
    manifest.go        # Rite manifest types and loading
    discovery.go       # Rite discovery (adapts team/discovery)
    invoker.go         # Invoke/release operations
    state.go           # Invocation state management
    budget.go          # Context budget calculations
    migrate.go         # Rite migration utilities

  cmd/rite/
    rite.go            # Root rite command
    invoke.go          # ari rite invoke
    release.go         # ari rite release
    current.go         # ari rite current
    list.go            # ari rite list
    info.go            # ari rite info
    swap.go            # ari rite swap (wraps team/switch)
    migrate.go         # ari rite migrate
```

### 8.3 Path Additions

```go
// In paths/resolver.go
func (r *Resolver) InvocationStateFile() string {
    return filepath.Join(r.ClaudeDir(), "INVOCATION_STATE.yaml")
}

func (r *Resolver) RiteDir(riteName string) string {
    // Check project rites first
    projectPath := filepath.Join(r.TeamsDir(), riteName)
    if _, err := os.Stat(filepath.Join(projectPath, "rite.yaml")); err == nil {
        return projectPath
    }
    // Fall back to user rites
    return filepath.Join(UserRitesDir(), riteName)
}

func UserRitesDir() string {
    return filepath.Join(UserDir(), "rites")
}
```

---

## 9. Error Handling

### 9.1 Error Types

```go
// In errors/rite.go

var (
    ErrRiteNotFound = func(name string) error {
        return NewWithDetails(CodeNotFound,
            "rite not found",
            map[string]interface{}{"rite": name})
    }

    ErrBorrowConflict = func(conflicts []string) error {
        return NewWithDetails(CodeConflict,
            "borrowing would conflict with existing invocations",
            map[string]interface{}{"conflicts": conflicts})
    }

    ErrBudgetExceeded = func(current, requested, limit int) error {
        return NewWithDetails(CodeBudgetExceeded,
            fmt.Sprintf("context budget exceeded: %d + %d > %d", current, requested, limit),
            map[string]interface{}{
                "current":   current,
                "requested": requested,
                "limit":     limit,
            })
    }

    ErrInvalidRiteForm = func(form, required string) error {
        return NewWithDetails(CodeUsageError,
            fmt.Sprintf("rite form '%s' does not support requested component '%s'", form, required),
            map[string]interface{}{
                "form":     form,
                "required": required,
            })
    }

    ErrInvocationNotFound = func(id string) error {
        return NewWithDetails(CodeNotFound,
            "invocation not found",
            map[string]interface{}{"invocation_id": id})
    }
)
```

### 9.2 Recovery Strategies

| Error | Recovery |
|-------|----------|
| `ErrBorrowConflict` | Release conflicting invocation first |
| `ErrBudgetExceeded` | Release unused invocations or use component filter |
| `ErrInvalidRiteForm` | Use appropriate component filter |
| Partial install failure | Rollback all changes atomically |

---

## 10. Testing Strategy

### 10.1 Unit Tests

| Area | Test Cases |
|------|------------|
| Manifest parsing | Valid/invalid schemas, all forms |
| Component selection | Filter by skills, agents, all |
| Budget calculation | Token estimation accuracy |
| State management | Add/remove invocations, persistence |
| Conflict detection | Same agent from different rites |

### 10.2 Integration Tests

| Scenario | Validation |
|----------|------------|
| Full invoke cycle | Invoke -> verify -> release |
| Multiple invocations | Stack multiple rites |
| Session lifecycle | Park/resume preserves state |
| Migration | Rite converts correctly |
| CLAUDE.md injection | Region updates correctly |

### 10.3 E2E Tests

```bash
# Test invoke-release cycle
ari rite invoke documentation skills
ari rite current --output=json | jq '.borrowed'
ari rite release documentation

# Test swap compatibility
ari team switch 10x-dev --remove-all
ari rite swap security-rite --remove-all

# Test budget enforcement
ARIADNE_CONTEXT_LIMIT=5000 ari rite invoke large-rite
# Should fail with budget exceeded
```

---

## 11. Security Considerations

### 11.1 Agent Injection Safety

- Borrowed agents are copied (not symlinked) to prevent tampering
- Checksum verification ensures agent integrity
- Source rite is recorded for audit trail

### 11.2 Skill Reference Safety

- Skills are referenced, not copied
- Only registered skill paths are allowed
- No arbitrary path injection

### 11.3 State File Protection

- `INVOCATION_STATE.yaml` follows same protection as other context files
- Write guard hook intercepts direct modifications
- Changes routed through Moirai

---

## 12. Acceptance Criteria

### 12.1 Functional Requirements

- [ ] `ari rite invoke <name>` adds components without removing current
- [ ] `ari rite invoke <name> skills` adds only skills
- [ ] `ari rite invoke <name> agents` adds only agents
- [ ] `ari rite release <name>` removes borrowed components
- [ ] `ari rite release --all` removes all borrowed components
- [ ] `ari rite current` shows native and borrowed components
- [ ] `ari rite list` shows available rites with forms
- [ ] `ari rite swap <name>` performs full context switch
- [ ] `ari team switch` continues to work (backward compatibility)

### 12.2 Non-Functional Requirements

- [ ] Invoke operation completes in <500ms
- [ ] Budget calculation within 10% of actual token count
- [ ] State file <10KB for typical usage
- [ ] Graceful handling of corrupted state files

### 12.3 Integration Requirements

- [ ] CLAUDE.md updates via inscription system
- [ ] Moirai session lifecycle integration
- [ ] Hook context includes invocation state
- [ ] Migration preserves all rite functionality

---

## 13. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Context budget overflow | Medium | High | Hard limits, warnings, release prompts |
| Agent conflicts | Low | Medium | Conflict detection, clear error messages |
| Migration data loss | Low | High | Dry-run mode, backup before migration |
| Stale invocation state | Medium | Low | Session lifecycle cleanup |
| Complex mental model | Medium | Medium | Clear CLI output, documentation |

---

## 14. Open Questions

1. **Skill inheritance**: Should borrowed skills be available to borrowed agents, or isolated?
   - *Recommendation*: Available - simplifies mental model

2. **Expiration policy**: Should invocations expire after session timeout?
   - *Recommendation*: Session-scoped by default, with optional TTL

3. **Nested invocations**: Can a borrowed rite invoke another rite?
   - *Recommendation*: No - prevents infinite loops, simplifies tracking

4. **Partial release**: Can you release just skills from an invocation that borrowed both?
   - *Recommendation*: No - invocation is atomic unit, release all or nothing

---

## Appendix A: Example Rite Manifests

### A.1 Simple Rite (Skills Only)

```yaml
schema_version: "1.0"
name: documentation-rite
display_name: "Documentation Rite"
description: Knowledge crystallization and documentation standards
form: simple

skills:
  - ref: doc-standards
    path: skills/doc-standards/
  - ref: templates
    path: skills/templates/

budget:
  estimated_tokens: 2500
  skills_cost: 2500
```

### A.2 Practitioner Rite (Agents + Skills)

```yaml
schema_version: "1.0"
name: code-review-rite
display_name: "Code Review Rite"
description: Structured code review with adversarial testing
form: practitioner

agents:
  - name: reviewer
    file: agents/reviewer.md
    role: "Reviews code for quality and correctness"
  - name: adversary
    file: agents/adversary.md
    role: "Finds edge cases and potential bugs"

skills:
  - ref: review-checklist
    path: skills/review-checklist/
  - ref: smell-detection
    external: true

budget:
  estimated_tokens: 5500
  agents_cost: 4000
  skills_cost: 1500
```

### A.3 Full Rite (All Components)

```yaml
schema_version: "1.0"
name: 10x-dev-rite
display_name: "10x Development Rite"
description: Full development lifecycle (PRD -> TDD -> Code -> QA)
form: full

agents:
  - name: requirements-analyst
    file: agents/requirements-analyst.md
    role: "Extracts stakeholder needs"
    produces: prd
  - name: architect
    file: agents/architect.md
    role: "Evaluates tradeoffs and designs systems"
    produces: tdd
  - name: principal-engineer
    file: agents/principal-engineer.md
    role: "Transforms designs into production code"
    produces: code
  - name: qa-adversary
    file: agents/qa-adversary.md
    role: "Breaks things so users don't"
    produces: test-plan

skills:
  - ref: 10x-workflow
    path: skills/10x-workflow/
  - ref: doc-artifacts
    path: skills/doc-artifacts/
  - ref: standards
    external: true

workflow:
  type: sequential
  entry_point: requirements-analyst
  phases:
    - name: requirements
      agent: requirements-analyst
      produces: prd
      next: design
    - name: design
      agent: architect
      produces: tdd
      next: implementation
    - name: implementation
      agent: principal-engineer
      produces: code
      next: validation
    - name: validation
      agent: qa-adversary
      produces: test-plan
      next: null

budget:
  estimated_tokens: 12500
  agents_cost: 8000
  skills_cost: 3500
  workflow_cost: 1000

migration:
  from_team: 10x-dev
```

---

## Appendix B: File Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD Document | /Users/tomtenuta/Code/roster/docs/design/TDD-invoke-rite.md | Pending |

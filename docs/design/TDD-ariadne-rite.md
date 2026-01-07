# TDD: Ariadne Rite Domain

> Technical Design Document for the rite domain of the Ariadne Go CLI

**Status**: Draft
**Author**: Architect Agent
**Date**: 2026-01-04
**PRD**: docs/requirements/PRD-ariadne.md
**Spike**: docs/spikes/SPIKE-ariadne-go-cli-architecture.md
**Reference**: docs/design/TDD-ariadne-session.md (Phase 1 implementation)

---

## 1. Overview

This Technical Design Document specifies the implementation of the **rite domain** for Ariadne (`ari`), the Go binary replacement for the roster bash script harness. The rite domain encompasses 4 commands that manage agent rites -- the specialized agent configurations that enable different workflow patterns.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-ariadne.md` (Section 4.1) |
| Spike | `docs/spikes/SPIKE-ariadne-go-cli-architecture.md` |
| Session TDD | `docs/design/TDD-ariadne-session.md` |
| Current Implementation | `swap-rite.sh`, `lib/rite/*.sh` |
| Rites Directory | `rites/` (source of rites) |
| Workflow Schema | `schemas/orchestrator.yaml.schema.json` |

### 1.2 Scope

**In Scope**:
- 4 rite commands: switch, list, status, validate
- Internal packages: `cmd/rite/`, extension of `paths/`, `output/`
- Error handling with exit codes per PRD Section 5.1
- CLAUDE.md satellite system integration
- Agent resolution and manifest management
- Orphan handling strategies (remove-all, keep-all, promote-all)

**Out of Scope**:
- Manifest domain (separate TDD - handles manifest show/diff/validate/merge)
- Three-way merge for manifests (manifest domain responsibility)
- Rite pack creation/editing (forge-pack responsibility)
- Hook registration (sync domain responsibility)

### 1.3 Design Goals

1. **Atomic Operations**: Rite switches complete fully or roll back
2. **Explicit Orphan Handling**: No silent data loss; require user decision on orphaned agents
3. **CLAUDE.md Consistency**: Satellite sections always reflect active rite
4. **Manifest Integrity**: AGENT_MANIFEST.json provides audit trail and validation
5. **Backward Compatibility**: Support existing rites/ structure without changes

---

## 2. Architecture

### 2.1 Package Structure

```
ariadne/
├── internal/
│   ├── cmd/
│   │   └── rite/
│   │       ├── rite.go              # Parent command registration
│   │       ├── switch.go            # ari rite switch
│   │       ├── list.go              # ari rite list
│   │       ├── status.go            # ari rite status
│   │       └── validate.go          # ari rite validate
│   ├── rite/
│   │   ├── discovery.go             # Rite location and enumeration
│   │   ├── manifest.go              # AGENT_MANIFEST.json operations
│   │   ├── resolver.go              # Agent resolution from rite to .claude/agents/
│   │   ├── switch.go                # Switch orchestration logic
│   │   ├── orphan.go                # Orphan detection and handling
│   │   ├── claudemd.go              # CLAUDE.md satellite updates
│   │   └── workflow.go              # workflow.yaml parsing
│   ├── paths/
│   │   └── rite.go                  # Rite-specific path resolution (extends existing)
│   └── output/
│       └── rite.go                  # Rite-specific output structures (extends existing)
```

### 2.2 Dependency Graph

```
                    ┌─────────────────────────────────┐
                    │  internal/cmd/rite/rite.go      │
                    │  (4 commands)                   │
                    └─────────────┬───────────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                        │                        │
         v                        v                        v
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ internal/rite/  │     │ internal/paths/ │     │ internal/output/│
│ (business logic)│     │ (extended)      │     │ (extended)      │
└────────┬────────┘     └─────────────────┘     └─────────────────┘
         │
         v
┌─────────────────────────────────────────────────────────────────┐
│  Filesystem: rites/, .claude/agents/, .claude/ACTIVE_RITE,      │
│  .claude/AGENT_MANIFEST.json, .claude/CLAUDE.md                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 Key Concepts

#### Rite Pack Structure

```
rites/{rite-name}/
├── agents/                     # Agent .md files
│   ├── orchestrator.md
│   ├── principal-engineer.md
│   └── ...
├── workflow.yaml               # Workflow phases and routing
├── orchestrator.yaml           # Orchestrator generation config
├── commands/                   # Rite-specific commands (optional)
├── skills/                     # Rite-specific skills (optional)
└── hooks/                      # Rite-specific hooks (optional)
```

#### Active Rite State

```
.claude/
├── ACTIVE_RITE                 # Single line: rite name (e.g., "10x-dev-pack")
├── ACTIVE_WORKFLOW.yaml        # Copy of rite's workflow.yaml
├── AGENT_MANIFEST.json         # Manifest tracking agent origins
├── CLAUDE.md                   # Contains satellite sections updated on switch
└── agents/
    ├── orchestrator.md         # Copied from active rite
    ├── principal-engineer.md   # Copied from active rite
    └── ...
```

---

## 3. Interface Contracts

### 3.1 Command Summary

| Command | Description | Modifies State |
|---------|-------------|----------------|
| `switch` | Switch to a different rite | Yes |
| `list` | List available rites | No |
| `status` | Show current team status | No |
| `validate` | Validate rite integrity | No |

### 3.2 Command: `ari rite switch`

Switches the active rite with atomic operations and orphan handling.

**Signature**:
```
ari rite switch <rite-name> [--remove-all|--keep-all|--promote-all] [--dry-run]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `rite-name` | Yes | Target rite name (e.g., "10x-dev-pack", "rnd-pack") |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--remove-all` | `-r` | bool | false | Remove all orphaned agents from disk |
| `--keep-all` | `-k` | bool | false | Keep all orphaned agents in .claude/agents/ |
| `--promote-all` | `-p` | bool | false | Promote orphans to project-level agents |
| `--update` | `-u` | bool | false | Re-pull agents even if already on target team |
| `--dry-run` | | bool | false | Preview changes without applying |

**Orphan Handling Semantics**:

| Flag | Behavior | AGENT_MANIFEST Entry |
|------|----------|---------------------|
| `--remove-all` | Delete orphaned agent files | Removed from manifest |
| `--keep-all` | Leave orphaned agent files unchanged | Marked `"orphaned": true` |
| `--promote-all` | Move to project-level (remove from manifest) | Marked `"source": "project"` |
| (none) | Error if orphans detected | N/A |

**Output (JSON)**:
```json
{
  "rite": "10x-dev-pack",
  "previous_rite": "rnd-pack",
  "switched_at": "2026-01-04T18:00:00Z",
  "agents_installed": [
    "architect.md",
    "orchestrator.md",
    "principal-engineer.md",
    "qa-adversary.md",
    "requirements-analyst.md"
  ],
  "orphans_handled": {
    "strategy": "remove-all",
    "agents": ["technology-scout.md", "research-analyst.md"]
  },
  "claude_md_updated": true,
  "manifest_path": ".claude/AGENT_MANIFEST.json"
}
```

**Output (text)**: Silent on success (exit 0)

**Output (dry-run, JSON)**:
```json
{
  "dry_run": true,
  "would_switch_to": "10x-dev-pack",
  "current_team": "rnd-pack",
  "would_install": ["architect.md", "orchestrator.md", "..."],
  "orphans_detected": ["technology-scout.md", "research-analyst.md"],
  "orphan_strategy_required": true,
  "suggested_flags": ["--remove-all", "--keep-all", "--promote-all"]
}
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Rite switch successful |
| 2 | Invalid arguments (unknown rite, conflicting flags) |
| 5 | Orphan conflict (orphans detected, no strategy specified) |
| 6 | Rite not found |
| 7 | Permission denied (cannot write to .claude/) |
| 9 | No .claude/ directory found |

**Error Response (JSON)**:
```json
{
  "error": {
    "code": "ORPHAN_CONFLICT",
    "message": "Orphaned agents detected. Specify --remove-all, --keep-all, or --promote-all",
    "details": {
      "orphans": ["technology-scout.md", "research-analyst.md"],
      "current_team": "rnd-pack",
      "target_team": "10x-dev-pack"
    }
  }
}
```

**Implementation Notes**:
- Validates target rite exists in `rites/` or user-level rites directory
- Detects orphaned agents (agents in .claude/agents/ not from target rite)
- Requires explicit orphan handling flag if orphans detected
- Copies agents from `rites/{rite-name}/agents/` to `.claude/agents/`
- Updates `ACTIVE_RITE`, `ACTIVE_WORKFLOW.yaml`, `AGENT_MANIFEST.json`
- Updates CLAUDE.md satellite sections (Quick Start table, Agent Configurations)
- Transaction safety: backup before, restore on failure

### 3.3 Command: `ari team list`

Lists all available rites from discovery locations.

**Signature**:
```
ari team list [--format=FORMAT]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-f` | string | `table` | Output format: table, name-only, json, yaml |

**Output (JSON)**:
```json
{
  "rites": [
    {
      "name": "10x-dev-pack",
      "description": "Full development lifecycle (PRD -> TDD -> Code -> QA)",
      "agents": ["architect", "orchestrator", "principal-engineer", "qa-adversary", "requirements-analyst"],
      "agent_count": 5,
      "path": "/Users/tom/Code/roster/rites/10x-dev-pack",
      "active": true
    },
    {
      "name": "rnd-pack",
      "description": "Research and exploration",
      "agents": ["technology-scout", "research-analyst"],
      "agent_count": 2,
      "path": "/Users/tom/Code/roster/rites/rnd-pack",
      "active": false
    }
  ],
  "total": 2,
  "active_rite": "10x-dev-pack"
}
```

**Output (text/table)**:
```
TEAM              AGENTS  DESCRIPTION                                     ACTIVE
10x-dev-pack      5       Full development lifecycle (PRD -> TDD -> ...)  *
rnd-pack          2       Research and exploration
security-pack     3       Security analysis and threat modeling
sre-pack          4       Site reliability and operations

Total: 4 rites (* = active)
```

**Output (name-only)**:
```
10x-dev-pack
rnd-pack
security-pack
sre-pack
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | List retrieved successfully (even if empty) |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Scans `rites/` directory in roster location
- Also scans user-level teams at `$XDG_DATA_HOME/ariadne/rites/` if present
- Reads `workflow.yaml` description field for each team
- Counts agents by scanning `agents/` subdirectory
- Marks currently active rite (from `ACTIVE_RITE` file)

### 3.4 Command: `ari rite status`

Shows detailed status of the current or specified rite.

**Signature**:
```
ari rite status [--rite=NAME]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--rite` | `-t` | string | (active) | Rite to query status for |

**Output (JSON)**:
```json
{
  "rite": "10x-dev-pack",
  "is_active": true,
  "path": "/Users/tom/Code/roster/rites/10x-dev-pack",
  "description": "Full development lifecycle (PRD -> TDD -> Code -> QA)",
  "workflow_type": "sequential",
  "agents": [
    {
      "name": "architect",
      "file": "architect.md",
      "role": "Evaluates tradeoffs and designs systems",
      "produces": "TDD",
      "installed": true
    },
    {
      "name": "orchestrator",
      "file": "orchestrator.md",
      "role": "Coordinates development lifecycle",
      "produces": "Work breakdown",
      "installed": true
    },
    {
      "name": "principal-engineer",
      "file": "principal-engineer.md",
      "role": "Transforms designs into production code",
      "produces": "Code",
      "installed": true
    },
    {
      "name": "qa-adversary",
      "file": "qa-adversary.md",
      "role": "Breaks things so users don't",
      "produces": "Test reports",
      "installed": true
    },
    {
      "name": "requirements-analyst",
      "file": "requirements-analyst.md",
      "role": "Extracts stakeholder needs",
      "produces": "PRD",
      "installed": true
    }
  ],
  "phases": ["requirements", "design", "implementation", "validation"],
  "entry_point": "requirements-analyst",
  "orphans": [],
  "manifest_valid": true,
  "claude_md_synced": true
}
```

**Output (text)**:
```
Rite: 10x-dev-pack (ACTIVE)
Path: /Users/tom/Code/roster/rites/10x-dev-pack
Description: Full development lifecycle (PRD -> TDD -> Code -> QA)
Workflow: sequential

Agents (5):
  architect           Evaluates tradeoffs and designs systems         [installed]
  orchestrator        Coordinates development lifecycle               [installed]
  principal-engineer  Transforms designs into production code         [installed]
  qa-adversary        Breaks things so users don't                    [installed]
  requirements-analyst Extracts stakeholder needs                     [installed]

Phases: requirements -> design -> implementation -> validation
Entry Point: requirements-analyst

Status: OK (manifest valid, CLAUDE.md synced)
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Status retrieved successfully |
| 6 | Team not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- If `--team` not specified, uses active rite from `ACTIVE_RITE` file
- Reads workflow.yaml for phase and entry point information
- Cross-references installed agents with manifest
- Checks CLAUDE.md satellite sections for sync status
- Returns `is_active: false` if querying non-active rite

### 3.5 Command: `ari rite validate`

Validates rite structure and configuration integrity.

**Signature**:
```
ari rite validate [--rite=NAME] [--fix]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--rite` | `-t` | string | (active) | Rite to validate |
| `--fix` | | bool | false | Attempt automatic repairs |

**Validation Rules**:

| Check | Description | Severity |
|-------|-------------|----------|
| `TEAM_EXISTS` | Team directory exists in discovery locations | Error |
| `AGENTS_DIR` | agents/ subdirectory exists | Error |
| `WORKFLOW_YAML` | workflow.yaml exists and is valid YAML | Error |
| `WORKFLOW_SCHEMA` | workflow.yaml validates against schema | Warning |
| `AGENT_FILES` | All referenced agents have .md files | Error |
| `ORCHESTRATOR_YAML` | orchestrator.yaml exists (if team has orchestrator) | Warning |
| `MANIFEST_SYNC` | AGENT_MANIFEST.json matches installed agents | Warning |
| `CLAUDE_MD_SYNC` | CLAUDE.md satellites match active rite | Warning |
| `NO_CIRCULAR_DEPS` | No circular phase dependencies | Error |
| `VALID_ENTRY_POINT` | Entry point agent exists | Error |

**Output (JSON)**:
```json
{
  "rite": "10x-dev-pack",
  "valid": true,
  "checks": [
    {"check": "TEAM_EXISTS", "status": "pass", "message": "Team directory found"},
    {"check": "AGENTS_DIR", "status": "pass", "message": "agents/ directory exists"},
    {"check": "WORKFLOW_YAML", "status": "pass", "message": "workflow.yaml is valid YAML"},
    {"check": "WORKFLOW_SCHEMA", "status": "pass", "message": "Validates against schema"},
    {"check": "AGENT_FILES", "status": "pass", "message": "All 5 agent files present"},
    {"check": "MANIFEST_SYNC", "status": "pass", "message": "Manifest matches installed"},
    {"check": "CLAUDE_MD_SYNC", "status": "pass", "message": "CLAUDE.md satellites synced"},
    {"check": "NO_CIRCULAR_DEPS", "status": "pass", "message": "No circular dependencies"},
    {"check": "VALID_ENTRY_POINT", "status": "pass", "message": "Entry point 'requirements-analyst' exists"}
  ],
  "errors": 0,
  "warnings": 0
}
```

**Output (text)**:
```
Validating team: 10x-dev-pack

[PASS] TEAM_EXISTS      Team directory found
[PASS] AGENTS_DIR       agents/ directory exists
[PASS] WORKFLOW_YAML    workflow.yaml is valid YAML
[PASS] WORKFLOW_SCHEMA  Validates against schema
[PASS] AGENT_FILES      All 5 agent files present
[PASS] MANIFEST_SYNC    Manifest matches installed
[PASS] CLAUDE_MD_SYNC   CLAUDE.md satellites synced
[PASS] NO_CIRCULAR_DEPS No circular dependencies
[PASS] VALID_ENTRY_POINT Entry point 'requirements-analyst' exists

Result: VALID (0 errors, 0 warnings)
```

**Output (with failures)**:
```json
{
  "team": "broken-pack",
  "valid": false,
  "checks": [
    {"check": "TEAM_EXISTS", "status": "pass", "message": "Team directory found"},
    {"check": "AGENTS_DIR", "status": "pass", "message": "agents/ directory exists"},
    {"check": "WORKFLOW_YAML", "status": "fail", "message": "workflow.yaml missing required field 'entry_point'"},
    {"check": "AGENT_FILES", "status": "fail", "message": "Missing agent files: architect.md, qa-adversary.md"},
    {"check": "MANIFEST_SYNC", "status": "warn", "message": "Manifest out of sync: 2 extra agents installed"}
  ],
  "errors": 2,
  "warnings": 1,
  "fixable": ["MANIFEST_SYNC"]
}
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Validation passed (all checks pass or only warnings) |
| 1 | Validation failed (at least one error) |
| 6 | Team not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- When `--team` not specified, validates active rite
- `--fix` attempts repairs for fixable issues (manifest sync, CLAUDE.md sync)
- Returns error count as exit code indicator
- Schema validation uses embedded workflow schema

---

## 4. Error Handling

### 4.1 Error Code Taxonomy

Extending PRD Section 5.1 with team-domain-specific codes:

| Code | Exit | Name | Description |
|------|------|------|-------------|
| `SUCCESS` | 0 | Success | Operation completed successfully |
| `GENERAL_ERROR` | 1 | General Error | Unspecified error |
| `USAGE_ERROR` | 2 | Usage Error | Invalid arguments or flags |
| `ORPHAN_CONFLICT` | 5 | Orphan Conflict | Orphaned agents detected without strategy |
| `TEAM_NOT_FOUND` | 6 | Team Not Found | Team pack does not exist |
| `PERMISSION_DENIED` | 7 | Permission Denied | Cannot write to .claude/ |
| `PROJECT_NOT_FOUND` | 9 | Project Not Found | No .claude/ directory found |
| `VALIDATION_FAILED` | 10 | Validation Failed | Team validation checks failed |
| `SWITCH_ABORTED` | 11 | Switch Aborted | Team switch rolled back due to error |

### 4.2 Error Response Structure

All errors follow the PRD Section 4.4 contract:

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

### 4.3 Transaction Safety

Team switch operations follow a transaction pattern:

```go
func switchTeam(ctx context.Context, targetRite string, opts SwitchOptions) error {
    // 1. Create backup of current state
    backup, err := createBackup()
    if err != nil {
        return errors.Wrap(err, "BACKUP_FAILED")
    }
    defer backup.Cleanup()

    // 2. Validate target team exists
    if !teamExists(targetRite) {
        return errors.New("TEAM_NOT_FOUND")
    }

    // 3. Detect orphans
    orphans := detectOrphans(targetRite)
    if len(orphans) > 0 && !opts.HasOrphanStrategy() {
        return errors.WithDetails("ORPHAN_CONFLICT", map[string]interface{}{
            "orphans": orphans,
        })
    }

    // 4. Execute switch (point of no return after ACTIVE_RITE write)
    if err := executeSwitch(targetRite, opts); err != nil {
        backup.Restore()
        return errors.Wrap(err, "SWITCH_ABORTED")
    }

    // 5. Update CLAUDE.md (non-critical, log warning on failure)
    if err := updateClaudeMD(targetRite); err != nil {
        log.Warn("CLAUDE.md update failed: %v", err)
    }

    return nil
}
```

---

## 5. Data Model

### 5.1 AGENT_MANIFEST.json

Tracks the origin and state of installed agents:

```json
{
  "version": "1.2",
  "generated_at": "2026-01-04T18:00:00Z",
  "active_team": "10x-dev-pack",
  "agents": {
    "architect.md": {
      "source": "team",
      "rite": "10x-dev-pack",
      "checksum": "sha256:abc123...",
      "installed_at": "2026-01-04T18:00:00Z"
    },
    "orchestrator.md": {
      "source": "team",
      "rite": "10x-dev-pack",
      "checksum": "sha256:def456...",
      "installed_at": "2026-01-04T18:00:00Z"
    },
    "custom-agent.md": {
      "source": "project",
      "checksum": "sha256:789abc...",
      "installed_at": "2026-01-02T10:00:00Z"
    }
  },
  "orphans": []
}
```

### 5.2 ACTIVE_RITE

Single line file containing the rite name:

```
10x-dev-pack
```

### 5.3 workflow.yaml Structure

Required fields for team discovery:

```yaml
name: 10x-dev-pack
workflow_type: sequential   # sequential | parallel | hybrid
description: Full development lifecycle (PRD -> TDD -> Code -> QA)

entry_point:
  agent: requirements-analyst

phases:
  - name: requirements
    agent: requirements-analyst
    produces: prd
    next: design
  # ...
```

### 5.4 CLAUDE.md Satellite Sections

Team switch updates these sections in CLAUDE.md:

**Quick Start Table** (between `## Quick Start` and `## Agent Routing`):
```markdown
## Quick Start

This project uses a 5-agent workflow (10x-dev-pack):

| Agent | Role | Produces |
| ----- | ---- | -------- |
| **architect** | Evaluates tradeoffs and designs systems | TDD |
| **orchestrator** | Coordinates development lifecycle | Work breakdown |
| **principal-engineer** | Transforms designs into production code | Code |
| **qa-adversary** | Breaks things so users don't | Test reports |
| **requirements-analyst** | Extracts stakeholder needs | PRD |
```

**Agent Configurations** (between `## Agent Configurations` and `## Hooks`):
```markdown
## Agent Configurations

Full agent prompts live in `.claude/agents/`:

- `architect.md` - System design authority who evaluates technical tradeoffs...
- `orchestrator.md` - Coordinates development lifecycle...
- `principal-engineer.md` - Master builder who transforms approved designs...
- `qa-adversary.md` - Adversarial tester who breaks implementations...
- `requirements-analyst.md` - Specification specialist who transforms ambiguity...
```

---

## 6. Internal Package Design

### 6.1 Package: `internal/team`

Core team domain logic, independent of CLI.

```go
package team

// Team represents a discovered rite
type Team struct {
    Name        string   `json:"name"`
    Path        string   `json:"path"`
    Description string   `json:"description"`
    Agents      []string `json:"agents"`
    WorkflowType string  `json:"workflow_type"`
    EntryPoint  string   `json:"entry_point"`
}

// Discovery locates available rites
type Discovery struct {
    rosterPath string
    userPath   string
}

func NewDiscovery(rosterPath, userPath string) *Discovery {
    return &Discovery{
        rosterPath: rosterPath,
        userPath:   userPath,
    }
}

// List returns all available teams
func (d *Discovery) List() ([]Team, error) {
    var teams []Team

    // Scan roster/rites/
    rosterTeams, _ := d.scanDir(filepath.Join(d.rosterPath, "teams"))
    teams = append(teams, rosterTeams...)

    // Scan user teams if present
    if d.userPath != "" {
        userTeams, _ := d.scanDir(d.userPath)
        teams = append(teams, userTeams...)
    }

    return teams, nil
}

// Get returns a specific team by name
func (d *Discovery) Get(name string) (*Team, error) {
    teams, err := d.List()
    if err != nil {
        return nil, err
    }

    for _, t := range teams {
        if t.Name == name {
            return &t, nil
        }
    }

    return nil, ErrTeamNotFound
}
```

### 6.2 Package: `internal/team/manifest`

AGENT_MANIFEST.json operations:

```go
package manifest

type Manifest struct {
    Version     string              `json:"version"`
    GeneratedAt time.Time           `json:"generated_at"`
    ActiveTeam  string              `json:"active_team"`
    Agents      map[string]AgentEntry `json:"agents"`
    Orphans     []string            `json:"orphans"`
}

type AgentEntry struct {
    Source      string    `json:"source"`       // "team" | "project"
    Team        string    `json:"team,omitempty"`
    Checksum    string    `json:"checksum"`
    InstalledAt time.Time `json:"installed_at"`
    Orphaned    bool      `json:"orphaned,omitempty"`
}

// Load reads manifest from path
func Load(path string) (*Manifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return NewEmptyManifest(), nil
        }
        return nil, err
    }

    var m Manifest
    if err := json.Unmarshal(data, &m); err != nil {
        return nil, fmt.Errorf("invalid manifest: %w", err)
    }

    return &m, nil
}

// Save writes manifest to path
func (m *Manifest) Save(path string) error {
    data, err := json.MarshalIndent(m, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

// DetectOrphans finds agents not belonging to target team
func (m *Manifest) DetectOrphans(targetRite string) []string {
    var orphans []string
    for name, entry := range m.Agents {
        if entry.Source == "team" && entry.Team != targetRite {
            orphans = append(orphans, name)
        }
    }
    return orphans
}
```

### 6.3 Package: `internal/team/claudemd`

CLAUDE.md satellite section updates:

```go
package claudemd

// Section represents a parseable section in CLAUDE.md
type Section string

const (
    SectionQuickStart      Section = "Quick Start"
    SectionAgentConfigs    Section = "Agent Configurations"
)

// Updater handles CLAUDE.md modifications
type Updater struct {
    path string
}

func NewUpdater(path string) *Updater {
    return &Updater{path: path}
}

// UpdateForTeam regenerates satellite sections for team
func (u *Updater) UpdateForTeam(team Team, workflow WorkflowInfo) error {
    content, err := os.ReadFile(u.path)
    if err != nil {
        return fmt.Errorf("reading CLAUDE.md: %w", err)
    }

    // Parse and update Quick Start section
    content = u.updateSection(content, SectionQuickStart,
        u.generateQuickStart(team, workflow))

    // Parse and update Agent Configurations section
    content = u.updateSection(content, SectionAgentConfigs,
        u.generateAgentConfigs(team))

    return os.WriteFile(u.path, content, 0644)
}

// generateQuickStart creates the Quick Start table markdown
func (u *Updater) generateQuickStart(team Team, workflow WorkflowInfo) string {
    var b strings.Builder

    b.WriteString(fmt.Sprintf("This project uses a %d-agent workflow (%s):\n\n",
        len(team.Agents), team.Name))
    b.WriteString("| Agent | Role | Produces |\n")
    b.WriteString("| ----- | ---- | -------- |\n")

    for _, agent := range workflow.Agents {
        b.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n",
            agent.Name, agent.Role, agent.Produces))
    }

    return b.String()
}
```

### 6.4 Package: `internal/team/switch`

Switch orchestration:

```go
package switch

type Options struct {
    TargetRite    string
    RemoveAll     bool
    KeepAll       bool
    PromoteAll    bool
    Update        bool
    DryRun        bool
}

func (o *Options) HasOrphanStrategy() bool {
    return o.RemoveAll || o.KeepAll || o.PromoteAll
}

type Result struct {
    Team            string   `json:"team"`
    PreviousTeam    string   `json:"previous_team"`
    SwitchedAt      time.Time `json:"switched_at"`
    AgentsInstalled []string `json:"agents_installed"`
    OrphansHandled  *OrphanResult `json:"orphans_handled,omitempty"`
    ClaudeMDUpdated bool     `json:"claude_md_updated"`
    ManifestPath    string   `json:"manifest_path"`
}

type OrphanResult struct {
    Strategy string   `json:"strategy"`
    Agents   []string `json:"agents"`
}

type Switcher struct {
    discovery    *Discovery
    manifest     *ManifestManager
    claudeMD     *ClaudeMDUpdater
    agentsDir    string
    projectRoot  string
}

func (s *Switcher) Switch(ctx context.Context, opts Options) (*Result, error) {
    // 1. Validate target team
    team, err := s.discovery.Get(opts.TargetRite)
    if err != nil {
        return nil, err
    }

    // 2. Load current manifest
    manifest, err := s.manifest.Load()
    if err != nil {
        return nil, err
    }

    // 3. Detect orphans
    orphans := manifest.DetectOrphans(opts.TargetRite)
    if len(orphans) > 0 && !opts.HasOrphanStrategy() {
        return nil, &OrphanConflictError{Orphans: orphans}
    }

    // 4. Dry run check
    if opts.DryRun {
        return s.dryRunResult(team, manifest, orphans, opts)
    }

    // 5. Execute switch
    return s.executeSwitch(ctx, team, manifest, orphans, opts)
}
```

---

## 7. Integration Points

### 7.1 Session Domain

Team domain interacts with session domain via `ACTIVE_RITE` file:
- `ari session create --team=NAME` uses team validation
- `ari session status` reads active rite for display
- Team switch does not affect existing sessions (they retain their team reference)

### 7.2 CLAUDE.md Satellite System

The satellite system in CLAUDE.md uses anchor comments (preserved content):
```markdown
<!-- PRESERVE: satellite-owned, regenerated from ACTIVE_RITE + agents/ -->
```

Team switch regenerates content between section headers while preserving:
- User-added content outside satellite sections
- PRESERVE comments themselves
- Non-team-related sections

### 7.3 Manifest Domain (Future)

Team domain produces `AGENT_MANIFEST.json` that manifest domain can:
- Validate (`ari manifest validate .claude/AGENT_MANIFEST.json`)
- Diff (`ari manifest diff old.json new.json`)
- Merge (for conflict resolution)

---

## 8. Test Strategy

### 8.1 Unit Tests

Location: `internal/team/*_test.go`

| Package | Test Focus | Coverage Target |
|---------|-----------|-----------------|
| `team` | Team discovery, listing | 100% |
| `manifest` | Load/save, orphan detection | 100% |
| `claudemd` | Section parsing, generation | 90% |
| `switch` | Options validation, dry-run | 100% |

### 8.2 Integration Tests

Location: `tests/integration/team_test.go`

| Test ID | Description |
|---------|-------------|
| `team_001` | Switch to valid team installs all agents |
| `team_002` | Switch with --remove-all cleans orphans |
| `team_003` | Switch with --keep-all preserves orphans |
| `team_004` | Switch with --promote-all marks as project |
| `team_005` | Switch without strategy fails on orphans |
| `team_006` | List shows all discoverable teams |
| `team_007` | Status shows installed vs expected agents |
| `team_008` | Validate catches missing agent files |
| `team_009` | Validate catches invalid workflow.yaml |
| `team_010` | Switch updates CLAUDE.md satellites |

### 8.3 Test Fixtures

```
ariadne/
└── testdata/
    └── rites/
        ├── valid-team/
        │   ├── agents/
        │   │   ├── agent-a.md
        │   │   └── agent-b.md
        │   ├── workflow.yaml
        │   └── orchestrator.yaml
        ├── minimal-team/
        │   ├── agents/
        │   │   └── solo-agent.md
        │   └── workflow.yaml
        └── broken-team/
            ├── agents/
            └── workflow.yaml  # missing required fields
```

---

## 9. Migration from Bash

### 9.1 swap-rite.sh Parity

The Go implementation follows **specification** behavior, not bash quirks:

| Behavior | Bash (swap-rite.sh) | Go (ari team switch) |
|----------|---------------------|----------------------|
| Orphan default | Interactive prompt | Error requiring flag |
| Transaction journal | Custom journal file | Backup/restore pattern |
| CLAUDE.md update | Non-critical, log warning | Same |
| Exit codes | Custom (0-6) | Standardized per PRD |

### 9.2 Integration Path

During migration, bash calls `ari`:

```bash
# swap-rite.sh (bridge)
case "$1" in
  switch) ari team switch "${@:2}" ;;
  list)   ari team list "${@:2}" ;;
  *)      echo "Unknown command: $1" >&2; exit 1 ;;
esac
```

### 9.3 Post-v1.0 Cleanup

After v1.0 ships:
- Delete `swap-rite.sh`
- Delete `lib/team/*.sh`
- Update documentation to reference `ari team` commands

---

## 10. Implementation Guidance

### 10.1 Recommended Order

1. **Foundation** (Day 1-2)
   - `internal/team/discovery.go` - Team enumeration
   - `internal/team/manifest.go` - Manifest operations
   - `internal/paths/team.go` - Path resolution extensions

2. **Read Operations** (Day 3-4)
   - `cmd/team/list.go` - List command
   - `cmd/team/status.go` - Status command

3. **Write Operations** (Day 5-7)
   - `internal/team/switch.go` - Switch logic
   - `internal/team/orphan.go` - Orphan handling
   - `cmd/team/switch.go` - Switch command

4. **Validation** (Day 8-9)
   - `internal/team/validate.go` - Validation rules
   - `cmd/team/validate.go` - Validate command

5. **CLAUDE.md Integration** (Day 10)
   - `internal/team/claudemd.go` - Satellite updates

### 10.2 Dependency on Session Domain

Team domain can reuse from session domain (Phase 1):
- `internal/paths` - Resolver pattern
- `internal/output` - Printer pattern
- Error types and exit code handling

---

## 11. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| CLAUDE.md parse errors | Medium | Low | Graceful degradation, warn on failure |
| Orphan data loss | Low | High | Require explicit flag, create backups |
| Manifest corruption | Low | Medium | Schema validation, checksums |
| workflow.yaml schema drift | Medium | Low | Embedded schema, version field |
| User teams directory conflicts | Low | Low | Explicit source field in manifest |

---

## 12. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-ariadne-004 | Proposed | Orphan handling strategy flags |
| ADR-ariadne-005 | Proposed | CLAUDE.md satellite update approach |
| ADR-ariadne-006 | Proposed | Manifest schema versioning |

---

## 13. Handoff Criteria

Ready for Implementation when:

- [x] All 4 team commands have interface contracts
- [x] Orphan handling strategies explicitly defined
- [x] Error codes mapped to exit codes
- [x] CLAUDE.md satellite sections documented
- [x] Validation rules enumerated
- [x] Test scenarios cover critical paths
- [ ] Principal Engineer can implement without architectural questions
- [ ] All artifacts verified via Read tool

---

## 14. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-rite.md` | Write |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-ariadne.md` | Read |
| Spike | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-ariadne-go-cli-architecture.md` | Read |
| Session TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-session.md` | Read |
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh` | Read (partial) |
| 10x workflow | `/Users/tomtenuta/Code/roster/rites/10x-dev-pack/workflow.yaml` | Read |
| Session implementation | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/session.go` | Read |
| Output package | `/Users/tomtenuta/Code/roster/ariadne/internal/output/output.go` | Read |

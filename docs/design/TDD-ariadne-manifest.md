# TDD: Ariadne Manifest Domain

> Technical Design Document for the manifest domain of the Ariadne Go CLI

**Status**: Draft
**Author**: Architect Agent
**Date**: 2026-01-04
**PRD**: docs/requirements/PRD-ariadne.md
**Spike**: docs/spikes/SPIKE-ariadne-go-cli-architecture.md
**Reference**: docs/design/TDD-ariadne-session.md (Phase 1), docs/design/TDD-ariadne-rite.md (Phase 2)

---

## 1. Overview

This Technical Design Document specifies the implementation of the **manifest domain** for Ariadne (`ari`), the Go binary replacement for the roster bash script harness. The manifest domain encompasses 4 commands that manage Claude Extension Manifest (CEM) operations -- the configuration files that define project structure, rites, and agent inventories.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-ariadne.md` (Sections 4.1, 6, 7) |
| Spike | `docs/spikes/SPIKE-ariadne-go-cli-architecture.md` |
| Session TDD | `docs/design/TDD-ariadne-session.md` |
| Team TDD | `docs/design/TDD-ariadne-rite.md` |
| Error Taxonomy | `ariadne/internal/errors/errors.go` |
| Validation Package | `ariadne/internal/validation/validator.go` |
| Team Manifest | `ariadne/internal/team/manifest.go` |
| Schema Directory | `schemas/` (JSON schemas for validation) |

### 1.2 Scope

**In Scope**:
- 4 manifest commands: show, diff, validate, merge
- Internal packages: `cmd/manifest/`, `manifest/`, extension of `validation/`
- Error handling with exit codes per PRD Section 5.1
- Schema validation using embedded JSON schemas
- Three-way merge with conflict detection per PRD Section 7
- Git-style conflict markers for unresolved conflicts

**Out of Scope**:
- Agent manifest operations (team domain handles `AGENT_MANIFEST.json`)
- Remote sync operations (sync domain responsibility)
- Manifest creation/editing (manual or sync domain)
- Schema authoring (separate tooling)

### 1.3 Design Goals

1. **Schema-First Validation**: All manifest operations validate against embedded JSON schemas
2. **Smart Merge**: Three-way merge with field-level conflict detection
3. **Explicit Conflicts**: Git-style conflict markers for human resolution
4. **Multiple Formats**: Support JSON and YAML manifest files
5. **Testability**: Pure functions for merge logic, dependency injection for I/O

---

## 2. Architecture

### 2.1 Package Structure

```
ariadne/
├── internal/
│   ├── cmd/
│   │   └── manifest/
│   │       ├── manifest.go         # Parent command registration
│   │       ├── show.go             # ari manifest show
│   │       ├── diff.go             # ari manifest diff
│   │       ├── validate.go         # ari manifest validate
│   │       └── merge.go            # ari manifest merge
│   ├── manifest/
│   │   ├── manifest.go             # Core manifest types and loading
│   │   ├── schema.go               # Schema type detection and mapping
│   │   ├── diff.go                 # Diff computation
│   │   ├── merge.go                # Three-way merge logic
│   │   └── conflict.go             # Conflict detection and markers
│   ├── validation/
│   │   ├── validator.go            # Existing schema validation (extend)
│   │   └── schemas/
│   │       ├── manifest.schema.json      # CEM manifest schema
│   │       ├── team-manifest.schema.json # Team pack manifest schema
│   │       └── *.schema.json             # Other embedded schemas
│   └── output/
│       └── manifest.go             # Manifest-specific output structures
```

### 2.2 Dependency Graph

```
                    ┌─────────────────────────────────┐
                    │  internal/cmd/manifest/         │
                    │  (4 commands)                   │
                    └─────────────┬───────────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                        │                        │
         v                        v                        v
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ internal/       │     │ internal/       │     │ internal/output/│
│ manifest/       │     │ validation/     │     │ (extended)      │
│ (business logic)│     │ (schemas)       │     └─────────────────┘
└────────┬────────┘     └────────┬────────┘
         │                       │
         │                       v
         │              ┌─────────────────────────────┐
         │              │ santhosh-tekuri/jsonschema  │
         │              │ evanphx/json-patch          │
         │              └─────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────────────────────┐
│  Filesystem: .claude/manifest.json, rites/*/manifest.yaml,     │
│  schemas/*.schema.json                                          │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 External Dependencies

Per PRD Section 3.2:

| Purpose | Library | Version | Import Path |
|---------|---------|---------|-------------|
| JSON Schema | santhosh-tekuri/jsonschema | v6+ | `github.com/santhosh-tekuri/jsonschema/v6` |
| JSON Merge | evanphx/json-patch | v5+ | `github.com/evanphx/json-patch/v5` |
| YAML | gopkg.in/yaml.v3 | v3 | `gopkg.in/yaml.v3` |

### 2.4 Key Concepts

#### Manifest Types

| Type | Location | Schema | Purpose |
|------|----------|--------|---------|
| CEM Project | `.claude/manifest.json` | `manifest.schema.json` | Project configuration |
| Team Pack | `rites/*/manifest.yaml` | `team-manifest.schema.json` | Team pack definition |
| Agent Manifest | `.claude/AGENT_MANIFEST.json` | (embedded) | Agent tracking (team domain) |

#### Remote Sources

For `diff` and `merge` operations, manifests can be loaded from:
- Local filesystem path
- Git ref (e.g., `HEAD:manifest.json`, `origin/main:.claude/manifest.json`)
- URL (future: rite registry)

---

## 3. Interface Contracts

### 3.1 Command Summary

| Command | Description | Modifies State |
|---------|-------------|----------------|
| `show` | Display current effective manifest | No |
| `diff` | Compare two manifests | No |
| `validate` | Validate manifest against schema | No |
| `merge` | Three-way merge with conflict resolution | Yes |

### 3.2 Command: `ari manifest show`

Displays the current effective manifest with schema information.

**Signature**:
```
ari manifest show [--path=PATH] [--schema] [--resolved]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--path` | `-p` | string | `.claude/manifest.json` | Path to manifest file |
| `--schema` | `-s` | bool | false | Include schema version and validation status |
| `--resolved` | `-r` | bool | false | Show resolved values (with defaults applied) |

**Output (JSON)**:
```json
{
  "path": ".claude/manifest.json",
  "exists": true,
  "format": "json",
  "schema": {
    "type": "manifest",
    "version": "1.0",
    "valid": true
  },
  "content": {
    "version": "1.0",
    "project": {
      "name": "roster",
      "description": "Claude Code agentic workflow management"
    },
    "teams": {
      "default": "10x-dev",
      "available": ["10x-dev", "rnd", "security"]
    },
    "paths": {
      "sessions": ".claude/sessions",
      "agents": ".claude/agents",
      "skills": ".claude/skills"
    }
  }
}
```

**Output (text)**:
```
Manifest: .claude/manifest.json
Format: JSON
Schema: manifest v1.0 (valid)

Project: roster
Description: Claude Code agentic workflow management

Teams:
  Default: 10x-dev
  Available: 10x-dev, rnd, security

Paths:
  Sessions: .claude/sessions
  Agents: .claude/agents
  Skills: .claude/skills
```

**Output (no manifest)**:
```json
{
  "path": ".claude/manifest.json",
  "exists": false,
  "error": null
}
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Manifest shown successfully (or doesn't exist with exists=false) |
| 6 | Path specified but file not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Defaults to `.claude/manifest.json` in project root
- Auto-detects format from extension (.json, .yaml, .yml)
- `--resolved` applies defaults from schema
- Returns `exists: false` gracefully if manifest doesn't exist yet

### 3.3 Command: `ari manifest diff`

Compares two manifest files and shows differences.

**Signature**:
```
ari manifest diff <path1> <path2> [--format=FORMAT] [--ignore-order]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `path1` | Yes | First manifest path (base) |
| `path2` | Yes | Second manifest path (comparison) |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-f` | string | `unified` | Diff format: unified, json, side-by-side |
| `--ignore-order` | | bool | false | Ignore array ordering differences |

**Git Ref Support**:

Paths can reference git objects:
- `HEAD:path/to/manifest.json` - Current commit
- `origin/main:path/to/manifest.json` - Remote branch
- `abc123:path/to/manifest.json` - Specific commit

**Output (JSON)**:
```json
{
  "base": ".claude/manifest.json",
  "compare": "HEAD~1:.claude/manifest.json",
  "has_changes": true,
  "changes": [
    {
      "path": "$.teams.default",
      "type": "modified",
      "old_value": "rnd",
      "new_value": "10x-dev"
    },
    {
      "path": "$.teams.available[2]",
      "type": "added",
      "new_value": "security"
    },
    {
      "path": "$.paths.hooks",
      "type": "removed",
      "old_value": ".claude/hooks"
    }
  ],
  "additions": 1,
  "modifications": 1,
  "deletions": 1
}
```

**Output (unified, text)**:
```
--- .claude/manifest.json
+++ HEAD~1:.claude/manifest.json

@@ teams @@
  "teams": {
-   "default": "rnd",
+   "default": "10x-dev",
    "available": [
      "10x-dev",
      "rnd",
+     "security"
    ]
  },

@@ paths @@
  "paths": {
    "sessions": ".claude/sessions",
    "agents": ".claude/agents",
-   "hooks": ".claude/hooks"
  }
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Diff completed successfully (with or without changes) |
| 1 | Diff completed, changes detected (for scripting: exit 1 = has changes) |
| 6 | One or both files not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Uses `evanphx/json-patch` for RFC 6902 diff generation
- Git ref paths parsed and resolved via `git show`
- `--ignore-order` treats arrays as sets for comparison
- Returns exit 1 when changes detected (useful for CI)

### 3.4 Command: `ari manifest validate`

Validates a manifest file against its JSON schema.

**Signature**:
```
ari manifest validate <path> [--schema=NAME] [--strict]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `path` | Yes | Path to manifest file to validate |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--schema` | `-s` | string | (auto-detect) | Schema name to validate against |
| `--strict` | | bool | false | Fail on additional properties not in schema |

**Schema Auto-Detection**:

| Path Pattern | Schema |
|--------------|--------|
| `.claude/manifest.json` | `manifest.schema.json` |
| `rites/*/manifest.yaml` | `team-manifest.schema.json` |
| `**/AGENT_MANIFEST.json` | `agent-manifest.schema.json` |
| `**/SESSION_CONTEXT.md` | `session-context.schema.json` |
| `**/SPRINT_CONTEXT.md` | `sprint-context.schema.json` |
| `**/workflow.yaml` | `orchestrator.yaml.schema.json` |

**Output (JSON, valid)**:
```json
{
  "path": ".claude/manifest.json",
  "schema": "manifest.schema.json",
  "valid": true,
  "issues": [],
  "warnings": []
}
```

**Output (JSON, invalid)**:
```json
{
  "path": ".claude/manifest.json",
  "schema": "manifest.schema.json",
  "valid": false,
  "issues": [
    {
      "path": "$.teams.default",
      "message": "must be one of: 10x-dev, rnd",
      "severity": "error"
    },
    {
      "path": "$.version",
      "message": "missing required field",
      "severity": "error"
    }
  ],
  "warnings": [
    {
      "path": "$.experimental",
      "message": "additional property not in schema",
      "severity": "warning"
    }
  ]
}
```

**Output (text, valid)**:
```
Validating: .claude/manifest.json
Schema: manifest.schema.json

Result: VALID (0 errors, 0 warnings)
```

**Output (text, invalid)**:
```
Validating: .claude/manifest.json
Schema: manifest.schema.json

[ERROR] $.teams.default: must be one of: 10x-dev, rnd
[ERROR] $.version: missing required field
[WARN]  $.experimental: additional property not in schema

Result: INVALID (2 errors, 1 warning)
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Validation passed |
| 4 | Schema validation failed (SCHEMA_INVALID) |
| 6 | File not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Uses `santhosh-tekuri/jsonschema/v6` for validation
- Schemas are embedded in binary via `//go:embed`
- Auto-detects schema from path pattern if not specified
- `--strict` enables `additionalProperties: false` behavior
- YAML files converted to JSON for validation

### 3.5 Command: `ari manifest merge`

Performs three-way merge of manifest files with conflict detection.

**Signature**:
```
ari manifest merge <base> <ours> <theirs> [--output=PATH] [--strategy=STRATEGY]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `base` | Yes | Common ancestor manifest |
| `ours` | Yes | Our version (local changes) |
| `theirs` | Yes | Their version (remote/incoming changes) |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | stdout | Output path for merged manifest |
| `--strategy` | `-s` | string | `smart` | Merge strategy: smart, ours, theirs, union |
| `--dry-run` | | bool | false | Preview merge without writing |

**Merge Strategies**:

| Strategy | Behavior |
|----------|----------|
| `smart` | Field-level merge per PRD Section 7.1 |
| `ours` | Prefer our changes on conflict |
| `theirs` | Prefer their changes on conflict |
| `union` | Merge arrays with union (no duplicates) |

**Three-Way Merge Semantics** (per PRD Section 7.1):

| Scenario | Resolution |
|----------|------------|
| Field in theirs only (new) | Accept theirs |
| Field in ours only (new) | Accept ours |
| Both modified same field differently | CONFLICT - flag for manual resolution |
| Only ours modified from base | Accept ours |
| Only theirs modified from base | Accept theirs |
| Neither modified from base | Keep base value |

**Output (JSON, clean merge)**:
```json
{
  "base": "manifest.base.json",
  "ours": "manifest.local.json",
  "theirs": "manifest.remote.json",
  "strategy": "smart",
  "has_conflicts": false,
  "merged": {
    "version": "1.0",
    "project": {
      "name": "roster",
      "description": "Updated description"
    },
    "teams": {
      "default": "10x-dev",
      "available": ["10x-dev", "rnd", "security", "sre"]
    }
  },
  "changes": {
    "from_ours": ["$.project.description"],
    "from_theirs": ["$.teams.available[3]"]
  }
}
```

**Output (JSON, with conflicts)**:
```json
{
  "base": "manifest.base.json",
  "ours": "manifest.local.json",
  "theirs": "manifest.remote.json",
  "strategy": "smart",
  "has_conflicts": true,
  "conflicts": [
    {
      "path": "$.teams.default",
      "base_value": "rnd",
      "ours_value": "10x-dev",
      "theirs_value": "security"
    }
  ],
  "merged_with_markers": "... (content with conflict markers)"
}
```

**Conflict Markers** (Git-style):

```json
{
  "teams": {
    "default": <<<<<<< ours
"10x-dev"
=======
"security"
>>>>>>> theirs
  }
}
```

**Output (text, clean merge)**:
```
Merging manifests...
  Base: manifest.base.json
  Ours: manifest.local.json
  Theirs: manifest.remote.json
  Strategy: smart

Changes:
  [OURS]   $.project.description: Updated description
  [THEIRS] $.teams.available[3]: added 'sre'

Result: MERGED (no conflicts)
Output: merged-manifest.json
```

**Output (text, with conflicts)**:
```
Merging manifests...
  Base: manifest.base.json
  Ours: manifest.local.json
  Theirs: manifest.remote.json
  Strategy: smart

Conflicts:
  $.teams.default:
    Base:   "rnd"
    Ours:   "10x-dev"
    Theirs: "security"

Result: CONFLICTS (1 conflict)
Output: merged-manifest.json (contains conflict markers)
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Merge completed successfully (no conflicts) |
| 8 | Merge completed with conflicts (MERGE_CONFLICT) |
| 6 | One or more input files not found |
| 4 | Input files fail schema validation |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Uses `evanphx/json-patch` for RFC 7386 JSON merge
- Custom conflict detection layer on top of json-patch
- Git-style conflict markers inserted for manual resolution
- Validates merged output against schema (unless conflicts exist)
- `--dry-run` previews merge without writing to disk

---

## 4. Error Handling

### 4.1 Error Code Taxonomy

Extending PRD Section 5.1 with manifest-domain-specific codes:

| Code | Exit | Name | Description |
|------|------|------|-------------|
| `SUCCESS` | 0 | Success | Operation completed successfully |
| `GENERAL_ERROR` | 1 | General Error | Unspecified error |
| `USAGE_ERROR` | 2 | Usage Error | Invalid arguments or flags |
| `SCHEMA_INVALID` | 4 | Schema Invalid | Manifest failed schema validation |
| `FILE_NOT_FOUND` | 6 | File Not Found | Manifest file missing |
| `MERGE_CONFLICT` | 8 | Merge Conflict | Three-way merge has unresolved conflicts |
| `PROJECT_NOT_FOUND` | 9 | Project Not Found | No .claude/ directory found |
| `SCHEMA_NOT_FOUND` | 14 | Schema Not Found | Specified schema not available |
| `PARSE_ERROR` | 15 | Parse Error | JSON/YAML parsing failed |

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

Example:
```json
{
  "error": {
    "code": "MERGE_CONFLICT",
    "message": "Three-way merge has unresolved conflicts",
    "details": {
      "conflict_count": 2,
      "conflicts": [
        "$.teams.default",
        "$.paths.skills"
      ],
      "output_path": "merged-manifest.json"
    }
  }
}
```

### 4.3 Error Constructors

New error constructors for manifest domain:

```go
// ErrMergeConflict returns an error for merge conflicts.
func ErrMergeConflict(conflictPaths []string, outputPath string) *Error {
    return NewWithDetails(CodeMergeConflict,
        "Three-way merge has unresolved conflicts",
        map[string]interface{}{
            "conflict_count": len(conflictPaths),
            "conflicts":      conflictPaths,
            "output_path":    outputPath,
        })
}

// ErrSchemaNotFound returns an error for missing schema.
func ErrSchemaNotFound(schemaName string) *Error {
    return NewWithDetails(CodeSchemaNotFound,
        fmt.Sprintf("Schema not found: %s", schemaName),
        map[string]interface{}{"schema": schemaName})
}

// ErrParseError returns an error for parsing failures.
func ErrParseError(path string, format string, cause error) *Error {
    return NewWithDetails(CodeParseError,
        fmt.Sprintf("Failed to parse %s file: %s", format, path),
        map[string]interface{}{
            "path":   path,
            "format": format,
            "cause":  cause.Error(),
        })
}
```

---

## 5. Data Model

### 5.1 Manifest Structure (CEM)

The Claude Extension Manifest (`.claude/manifest.json`):

```json
{
  "version": "1.0",
  "project": {
    "name": "roster",
    "description": "Claude Code agentic workflow management",
    "repository": "https://github.com/autom8y/roster"
  },
  "teams": {
    "default": "10x-dev",
    "available": ["10x-dev", "rnd", "security"],
    "discovery": ["rites/", "~/.config/ariadne/rites/"]
  },
  "paths": {
    "sessions": ".claude/sessions",
    "agents": ".claude/agents",
    "skills": ".claude/skills",
    "hooks": ".claude/hooks"
  },
  "schemas": {
    "session": "2.1",
    "sprint": "1.0"
  },
  "settings": {
    "auto_park_on_stop": true,
    "require_session": false
  }
}
```

### 5.2 Team Pack Manifest

Team pack manifest (`rites/*/manifest.yaml`):

```yaml
version: "1.0"
name: 10x-dev
description: Full development lifecycle (PRD -> TDD -> Code -> QA)

workflow:
  type: sequential
  entry_point: requirements-analyst

agents:
  - name: architect
    file: agents/architect.md
    role: Evaluates tradeoffs and designs systems
    produces: TDD
  - name: orchestrator
    file: agents/orchestrator.md
    role: Coordinates development lifecycle
    produces: Work breakdown
  - name: principal-engineer
    file: agents/principal-engineer.md
    role: Transforms designs into production code
    produces: Code
  - name: qa-adversary
    file: agents/qa-adversary.md
    role: Breaks things so users don't
    produces: Test reports
  - name: requirements-analyst
    file: agents/requirements-analyst.md
    role: Extracts stakeholder needs
    produces: PRD

skills:
  - commit-ref
  - pr-ref
  - qa-ref

hooks:
  - session-guards/auto-park.sh
```

### 5.3 Diff Result Structure

```go
type DiffResult struct {
    Base       string   `json:"base"`
    Compare    string   `json:"compare"`
    HasChanges bool     `json:"has_changes"`
    Changes    []Change `json:"changes"`
    Additions  int      `json:"additions"`
    Modifications int  `json:"modifications"`
    Deletions  int      `json:"deletions"`
}

type Change struct {
    Path     string      `json:"path"`
    Type     ChangeType  `json:"type"`
    OldValue interface{} `json:"old_value,omitempty"`
    NewValue interface{} `json:"new_value,omitempty"`
}

type ChangeType string

const (
    ChangeAdded    ChangeType = "added"
    ChangeModified ChangeType = "modified"
    ChangeRemoved  ChangeType = "removed"
)
```

### 5.4 Merge Result Structure

```go
type MergeResult struct {
    Base          string     `json:"base"`
    Ours          string     `json:"ours"`
    Theirs        string     `json:"theirs"`
    Strategy      string     `json:"strategy"`
    HasConflicts  bool       `json:"has_conflicts"`
    Conflicts     []Conflict `json:"conflicts,omitempty"`
    Merged        interface{} `json:"merged,omitempty"`
    MergedMarkers string     `json:"merged_with_markers,omitempty"`
    Changes       *MergeChanges `json:"changes,omitempty"`
}

type Conflict struct {
    Path        string      `json:"path"`
    BaseValue   interface{} `json:"base_value"`
    OursValue   interface{} `json:"ours_value"`
    TheirsValue interface{} `json:"theirs_value"`
}

type MergeChanges struct {
    FromOurs   []string `json:"from_ours"`
    FromTheirs []string `json:"from_theirs"`
}
```

---

## 6. Internal Package Design

### 6.1 Package: `internal/manifest`

Core manifest operations, independent of CLI.

```go
package manifest

import (
    "encoding/json"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

// Manifest represents a parsed manifest file.
type Manifest struct {
    Path    string
    Format  Format
    Content map[string]interface{}
    Raw     []byte
}

// Format represents the manifest file format.
type Format string

const (
    FormatJSON Format = "json"
    FormatYAML Format = "yaml"
)

// Load reads and parses a manifest from the given path.
func Load(path string) (*Manifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    format := detectFormat(path)
    content, err := parse(data, format)
    if err != nil {
        return nil, err
    }

    return &Manifest{
        Path:    path,
        Format:  format,
        Content: content,
        Raw:     data,
    }, nil
}

// LoadFromGitRef loads a manifest from a git reference.
func LoadFromGitRef(ref string) (*Manifest, error) {
    // Parse ref format: "commit:path"
    parts := strings.SplitN(ref, ":", 2)
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid git ref format: %s", ref)
    }

    commit, path := parts[0], parts[1]

    // Execute git show
    cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", commit, path))
    data, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("git show failed: %w", err)
    }

    format := detectFormat(path)
    content, err := parse(data, format)
    if err != nil {
        return nil, err
    }

    return &Manifest{
        Path:    ref,
        Format:  format,
        Content: content,
        Raw:     data,
    }, nil
}

// detectFormat determines format from file extension.
func detectFormat(path string) Format {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".yaml", ".yml":
        return FormatYAML
    default:
        return FormatJSON
    }
}

// parse decodes the manifest content.
func parse(data []byte, format Format) (map[string]interface{}, error) {
    var content map[string]interface{}

    switch format {
    case FormatYAML:
        if err := yaml.Unmarshal(data, &content); err != nil {
            return nil, err
        }
    default:
        if err := json.Unmarshal(data, &content); err != nil {
            return nil, err
        }
    }

    return content, nil
}

// Save writes the manifest to disk.
func (m *Manifest) Save(path string) error {
    var data []byte
    var err error

    format := detectFormat(path)
    switch format {
    case FormatYAML:
        data, err = yaml.Marshal(m.Content)
    default:
        data, err = json.MarshalIndent(m.Content, "", "  ")
    }

    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644)
}

// ToJSON converts manifest content to JSON bytes.
func (m *Manifest) ToJSON() ([]byte, error) {
    return json.Marshal(m.Content)
}
```

### 6.2 Package: `internal/manifest/diff`

Diff computation between manifests.

```go
package manifest

import (
    "encoding/json"

    jsonpatch "github.com/evanphx/json-patch/v5"
)

// Diff computes differences between two manifests.
func Diff(base, compare *Manifest, opts DiffOptions) (*DiffResult, error) {
    baseJSON, err := base.ToJSON()
    if err != nil {
        return nil, err
    }

    compareJSON, err := compare.ToJSON()
    if err != nil {
        return nil, err
    }

    // Create RFC 6902 diff
    patch, err := jsonpatch.CreateMergePatch(baseJSON, compareJSON)
    if err != nil {
        return nil, fmt.Errorf("diff creation failed: %w", err)
    }

    // Parse patch operations into changes
    changes := parsePatchToChanges(baseJSON, compareJSON, patch)

    // Apply ignore-order if requested
    if opts.IgnoreOrder {
        changes = filterOrderOnlyChanges(changes)
    }

    return &DiffResult{
        Base:          base.Path,
        Compare:       compare.Path,
        HasChanges:    len(changes) > 0,
        Changes:       changes,
        Additions:     countByType(changes, ChangeAdded),
        Modifications: countByType(changes, ChangeModified),
        Deletions:     countByType(changes, ChangeRemoved),
    }, nil
}

// DiffOptions configures diff behavior.
type DiffOptions struct {
    IgnoreOrder bool // Treat arrays as sets
}

// parsePatchToChanges converts a JSON merge patch to Change structs.
func parsePatchToChanges(base, compare, patch []byte) []Change {
    var changes []Change

    var baseMap, compareMap map[string]interface{}
    json.Unmarshal(base, &baseMap)
    json.Unmarshal(compare, &compareMap)

    // Walk the structures to find changes
    walkAndCompare("$", baseMap, compareMap, &changes)

    return changes
}

// walkAndCompare recursively compares two maps.
func walkAndCompare(path string, base, compare interface{}, changes *[]Change) {
    // ... implementation details
}
```

### 6.3 Package: `internal/manifest/merge`

Three-way merge logic.

```go
package manifest

import (
    "encoding/json"
    "fmt"
    "strings"
)

// Merge performs a three-way merge of manifests.
func Merge(base, ours, theirs *Manifest, opts MergeOptions) (*MergeResult, error) {
    baseJSON, _ := base.ToJSON()
    oursJSON, _ := ours.ToJSON()
    theirsJSON, _ := theirs.ToJSON()

    result := &MergeResult{
        Base:     base.Path,
        Ours:     ours.Path,
        Theirs:   theirs.Path,
        Strategy: string(opts.Strategy),
    }

    switch opts.Strategy {
    case StrategyOurs:
        return mergePreferOurs(result, ours)
    case StrategyTheirs:
        return mergePreferTheirs(result, theirs)
    case StrategyUnion:
        return mergeUnion(result, base, ours, theirs)
    default:
        return mergeSmart(result, baseJSON, oursJSON, theirsJSON)
    }
}

// MergeOptions configures merge behavior.
type MergeOptions struct {
    Strategy MergeStrategy
    DryRun   bool
}

// MergeStrategy defines how conflicts are resolved.
type MergeStrategy string

const (
    StrategySmart  MergeStrategy = "smart"
    StrategyOurs   MergeStrategy = "ours"
    StrategyTheirs MergeStrategy = "theirs"
    StrategyUnion  MergeStrategy = "union"
)

// mergeSmart performs field-level three-way merge.
func mergeSmart(result *MergeResult, base, ours, theirs []byte) (*MergeResult, error) {
    var baseMap, oursMap, theirsMap map[string]interface{}
    json.Unmarshal(base, &baseMap)
    json.Unmarshal(ours, &oursMap)
    json.Unmarshal(theirs, &theirsMap)

    merged := make(map[string]interface{})
    conflicts := []Conflict{}
    changes := &MergeChanges{
        FromOurs:   []string{},
        FromTheirs: []string{},
    }

    // Three-way merge each field
    mergeFields("$", baseMap, oursMap, theirsMap, merged, &conflicts, changes)

    result.HasConflicts = len(conflicts) > 0
    result.Conflicts = conflicts
    result.Merged = merged
    result.Changes = changes

    if result.HasConflicts {
        result.MergedMarkers = generateConflictMarkers(merged, conflicts)
    }

    return result, nil
}

// mergeFields performs three-way merge on map fields.
func mergeFields(path string, base, ours, theirs, merged map[string]interface{},
    conflicts *[]Conflict, changes *MergeChanges) {

    allKeys := collectKeys(base, ours, theirs)

    for _, key := range allKeys {
        fieldPath := fmt.Sprintf("%s.%s", path, key)
        baseVal, baseOk := base[key]
        oursVal, oursOk := ours[key]
        theirsVal, theirsOk := theirs[key]

        // Apply three-way merge semantics per PRD Section 7.1
        switch {
        case !baseOk && !oursOk && theirsOk:
            // New in theirs only -> accept theirs
            merged[key] = theirsVal
            changes.FromTheirs = append(changes.FromTheirs, fieldPath)

        case !baseOk && oursOk && !theirsOk:
            // New in ours only -> accept ours
            merged[key] = oursVal
            changes.FromOurs = append(changes.FromOurs, fieldPath)

        case baseOk && !oursOk && !theirsOk:
            // Deleted in both -> delete
            // (don't add to merged)

        case baseOk && !oursOk && theirsOk:
            // Deleted in ours, modified in theirs -> conflict or prefer theirs
            if !equal(baseVal, theirsVal) {
                *conflicts = append(*conflicts, Conflict{
                    Path:        fieldPath,
                    BaseValue:   baseVal,
                    OursValue:   nil,
                    TheirsValue: theirsVal,
                })
            }
            // Default: accept deletion

        case baseOk && oursOk && !theirsOk:
            // Modified in ours, deleted in theirs -> conflict or prefer ours
            if !equal(baseVal, oursVal) {
                *conflicts = append(*conflicts, Conflict{
                    Path:        fieldPath,
                    BaseValue:   baseVal,
                    OursValue:   oursVal,
                    TheirsValue: nil,
                })
            }
            // Default: keep ours
            merged[key] = oursVal
            changes.FromOurs = append(changes.FromOurs, fieldPath)

        case baseOk && oursOk && theirsOk:
            oursChanged := !equal(baseVal, oursVal)
            theirsChanged := !equal(baseVal, theirsVal)

            switch {
            case !oursChanged && !theirsChanged:
                // Neither modified -> keep base
                merged[key] = baseVal

            case oursChanged && !theirsChanged:
                // Only ours modified -> accept ours
                merged[key] = oursVal
                changes.FromOurs = append(changes.FromOurs, fieldPath)

            case !oursChanged && theirsChanged:
                // Only theirs modified -> accept theirs
                merged[key] = theirsVal
                changes.FromTheirs = append(changes.FromTheirs, fieldPath)

            case oursChanged && theirsChanged:
                if equal(oursVal, theirsVal) {
                    // Both changed to same value -> use it
                    merged[key] = oursVal
                } else {
                    // Conflict: both changed differently
                    *conflicts = append(*conflicts, Conflict{
                        Path:        fieldPath,
                        BaseValue:   baseVal,
                        OursValue:   oursVal,
                        TheirsValue: theirsVal,
                    })
                    // Use ours as base, add markers later
                    merged[key] = oursVal
                }
            }

        default:
            // New in both
            if equal(oursVal, theirsVal) {
                merged[key] = oursVal
            } else {
                *conflicts = append(*conflicts, Conflict{
                    Path:        fieldPath,
                    BaseValue:   nil,
                    OursValue:   oursVal,
                    TheirsValue: theirsVal,
                })
                merged[key] = oursVal
            }
        }
    }
}

// generateConflictMarkers creates Git-style conflict markers.
func generateConflictMarkers(merged map[string]interface{}, conflicts []Conflict) string {
    // Create a copy with conflict markers
    // This produces valid JSON with embedded marker strings
    markedJSON, _ := json.MarshalIndent(merged, "", "  ")
    result := string(markedJSON)

    for _, conflict := range conflicts {
        // Find and replace the conflicting value with markers
        oursJSON, _ := json.Marshal(conflict.OursValue)
        theirsJSON, _ := json.Marshal(conflict.TheirsValue)

        marker := fmt.Sprintf(`<<<<<<< ours
%s
=======
%s
>>>>>>> theirs`, string(oursJSON), string(theirsJSON))

        // Replace in result (simplified - real implementation needs path-aware replacement)
        result = strings.Replace(result, string(oursJSON), marker, 1)
    }

    return result
}
```

### 6.4 Package: `internal/validation` (Extension)

Extend existing validation package for manifest schemas.

```go
package validation

// Additional schema names for manifest domain
const (
    SchemaManifest     = "manifest"
    SchemaTeamManifest = "team-manifest"
    SchemaAgentManifest = "agent-manifest"
)

// ValidateManifest validates a CEM manifest.
func (v *Validator) ValidateManifest(data []byte) error {
    schema, err := v.getSchema(SchemaManifest)
    if err != nil {
        return err
    }

    var parsed interface{}
    if err := json.Unmarshal(data, &parsed); err != nil {
        return errors.Wrap(errors.CodeParseError, "invalid JSON", err)
    }

    if err := schema.Validate(parsed); err != nil {
        return errors.NewWithDetails(errors.CodeSchemaInvalid,
            "manifest validation failed",
            map[string]interface{}{"error": err.Error()})
    }

    return nil
}

// ValidateTeamManifest validates a rite manifest.
func (v *Validator) ValidateTeamManifest(data []byte) error {
    schema, err := v.getSchema(SchemaTeamManifest)
    if err != nil {
        return err
    }

    // YAML files need conversion to JSON
    var content interface{}
    if err := yaml.Unmarshal(data, &content); err != nil {
        return errors.Wrap(errors.CodeParseError, "invalid YAML", err)
    }

    if err := schema.Validate(content); err != nil {
        return errors.NewWithDetails(errors.CodeSchemaInvalid,
            "team manifest validation failed",
            map[string]interface{}{"error": err.Error()})
    }

    return nil
}

// DetectSchemaFromPath returns the schema name for a given path.
func DetectSchemaFromPath(path string) (string, error) {
    patterns := map[string]string{
        `.claude/manifest.json`:           SchemaManifest,
        `rites/*/manifest.yaml`:           SchemaTeamManifest,
        `AGENT_MANIFEST.json`:             SchemaAgentManifest,
        `SESSION_CONTEXT.md`:              "session-context",
        `SPRINT_CONTEXT.md`:               "sprint-context",
        `workflow.yaml`:                   "orchestrator",
    }

    for pattern, schema := range patterns {
        if matchPath(path, pattern) {
            return schema, nil
        }
    }

    return "", errors.New(errors.CodeSchemaNotFound,
        fmt.Sprintf("no schema detected for path: %s", path))
}
```

---

## 7. Schema Definitions

### 7.1 manifest.schema.json

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "embed:///schemas/manifest.schema.json",
  "title": "Claude Extension Manifest",
  "type": "object",
  "required": ["version"],
  "properties": {
    "version": {
      "type": "string",
      "pattern": "^[0-9]+\\.[0-9]+$",
      "description": "Manifest schema version"
    },
    "project": {
      "type": "object",
      "properties": {
        "name": { "type": "string" },
        "description": { "type": "string" },
        "repository": { "type": "string", "format": "uri" }
      }
    },
    "teams": {
      "type": "object",
      "properties": {
        "default": { "type": "string" },
        "available": {
          "type": "array",
          "items": { "type": "string" }
        },
        "discovery": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    },
    "paths": {
      "type": "object",
      "properties": {
        "sessions": { "type": "string" },
        "agents": { "type": "string" },
        "skills": { "type": "string" },
        "hooks": { "type": "string" }
      }
    },
    "schemas": {
      "type": "object",
      "additionalProperties": { "type": "string" }
    },
    "settings": {
      "type": "object",
      "properties": {
        "auto_park_on_stop": { "type": "boolean" },
        "require_session": { "type": "boolean" }
      }
    }
  },
  "additionalProperties": false
}
```

### 7.2 team-manifest.schema.json

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "embed:///schemas/team-manifest.schema.json",
  "title": "Team Pack Manifest",
  "type": "object",
  "required": ["version", "name", "workflow", "agents"],
  "properties": {
    "version": {
      "type": "string",
      "pattern": "^[0-9]+\\.[0-9]+$"
    },
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9-]+$"
    },
    "description": { "type": "string" },
    "workflow": {
      "type": "object",
      "required": ["type", "entry_point"],
      "properties": {
        "type": {
          "type": "string",
          "enum": ["sequential", "parallel", "hybrid"]
        },
        "entry_point": { "type": "string" }
      }
    },
    "agents": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "file"],
        "properties": {
          "name": { "type": "string" },
          "file": { "type": "string" },
          "role": { "type": "string" },
          "produces": { "type": "string" }
        }
      }
    },
    "skills": {
      "type": "array",
      "items": { "type": "string" }
    },
    "hooks": {
      "type": "array",
      "items": { "type": "string" }
    }
  }
}
```

---

## 8. Integration Points

### 8.1 Session Domain

Manifest domain provides schema validation that session domain uses:
- Session context schema validation via `ari manifest validate`
- Sprint context schema validation

### 8.2 Team Domain

Manifest domain supports team domain operations:
- `AGENT_MANIFEST.json` validation
- Team pack manifest validation (`rites/*/manifest.yaml`)
- Team switch can invoke `ari manifest validate` for pre-flight checks

### 8.3 Sync Domain (Future)

Sync domain will use manifest domain for:
- `ari manifest diff` to compare local vs remote manifests
- `ari manifest merge` for conflict resolution during sync
- `ari manifest validate` for post-sync validation

### 8.4 State-Mate Integration

State-mate agent can invoke manifest commands for validation:

```bash
# Validate session context before writing
ari manifest validate .claude/sessions/$SESSION_ID/SESSION_CONTEXT.md --schema=session-context

# Check for manifest drift
ari manifest diff .claude/manifest.json HEAD:.claude/manifest.json
```

---

## 9. Test Strategy

### 9.1 Unit Tests

Location: `internal/manifest/*_test.go`

| Package | Test Focus | Coverage Target |
|---------|-----------|-----------------|
| `manifest` | Load/parse JSON/YAML | 100% |
| `manifest/diff` | Change detection, path handling | 100% |
| `manifest/merge` | Three-way merge, conflict detection | 100% |
| `validation` | Schema validation, error messages | 100% |

### 9.2 Integration Tests

Location: `tests/integration/manifest_test.go`

| Test ID | Description |
|---------|-------------|
| `manifest_001` | Show displays valid manifest |
| `manifest_002` | Show returns exists=false for missing manifest |
| `manifest_003` | Diff detects additions |
| `manifest_004` | Diff detects modifications |
| `manifest_005` | Diff detects deletions |
| `manifest_006` | Diff with git ref works |
| `manifest_007` | Validate passes valid manifest |
| `manifest_008` | Validate fails invalid manifest with clear errors |
| `manifest_009` | Validate auto-detects schema |
| `manifest_010` | Merge with no conflicts succeeds |
| `manifest_011` | Merge with conflicts returns exit 8 |
| `manifest_012` | Merge with --strategy=ours resolves to ours |
| `manifest_013` | Merge with --strategy=theirs resolves to theirs |
| `manifest_014` | Merge with --strategy=union merges arrays |
| `manifest_015` | Conflict markers are valid git-style |

### 9.3 Three-Way Merge Test Cases

| Test Case | Base | Ours | Theirs | Expected |
|-----------|------|------|--------|----------|
| Both add same field | - | A | A | A (no conflict) |
| Both add different values | - | A | B | CONFLICT |
| Only ours modifies | A | B | A | B |
| Only theirs modifies | A | A | B | B |
| Both modify same | A | B | C | CONFLICT |
| Both modify to same | A | B | B | B (no conflict) |
| Ours deletes, theirs modifies | A | - | B | CONFLICT |
| Theirs deletes, ours modifies | A | B | - | CONFLICT |
| Both delete | A | - | - | (deleted) |

### 9.4 Test Fixtures

```
ariadne/
└── testdata/
    └── manifests/
        ├── valid/
        │   ├── project-manifest.json
        │   ├── team-manifest.yaml
        │   └── minimal-manifest.json
        ├── invalid/
        │   ├── missing-version.json
        │   ├── bad-team-reference.json
        │   └── malformed.json
        └── merge/
            ├── base.json
            ├── ours-add-field.json
            ├── theirs-add-field.json
            ├── conflict-same-field.json
            └── expected-merged.json
```

---

## 10. Implementation Guidance

### 10.1 Recommended Order

1. **Foundation** (Day 1-2)
   - `internal/manifest/manifest.go` - Core types and loading
   - `internal/manifest/schema.go` - Schema type detection
   - Extend `internal/validation/` with manifest schemas

2. **Read Operations** (Day 3-4)
   - `cmd/manifest/show.go` - Show command
   - `cmd/manifest/validate.go` - Validate command

3. **Diff** (Day 5-6)
   - `internal/manifest/diff.go` - Diff computation
   - `cmd/manifest/diff.go` - Diff command
   - Git ref support

4. **Merge** (Day 7-10)
   - `internal/manifest/merge.go` - Three-way merge logic
   - `internal/manifest/conflict.go` - Conflict detection and markers
   - `cmd/manifest/merge.go` - Merge command
   - Integration tests

### 10.2 Dependency on Existing Packages

Manifest domain reuses from session/team domains:
- `internal/paths` - Project root discovery
- `internal/output` - Printer pattern
- `internal/errors` - Error types and exit codes
- `internal/validation` - Schema validation (extend)

### 10.3 External Library Usage

```go
import (
    // JSON Schema validation
    "github.com/santhosh-tekuri/jsonschema/v6"

    // JSON diff/merge operations
    jsonpatch "github.com/evanphx/json-patch/v5"

    // YAML support
    "gopkg.in/yaml.v3"
)
```

---

## 11. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Complex merge conflicts | Medium | Medium | Clear conflict markers, exit code indicates conflicts |
| Git ref parsing edge cases | Low | Low | Well-tested regex, fallback to file path |
| Schema evolution | Medium | Medium | Version field in schemas, backward compatibility |
| YAML multiline strings | Low | Low | Use yaml.v3 which handles multiline properly |
| Large manifest files | Low | Low | Streaming parse for huge files if needed |
| Embedded schema size | Low | Low | Minify schemas at build time |

---

## 12. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-ariadne-007 | Proposed | Three-way merge strategy selection |
| ADR-ariadne-008 | Proposed | Conflict marker format (Git-style) |
| ADR-ariadne-009 | Proposed | Embedded schema management |

---

## 13. Handoff Criteria

Ready for Implementation when:

- [x] All 4 manifest commands have interface contracts
- [x] Schema structures defined for CEM and team manifests
- [x] Three-way merge semantics explicitly documented
- [x] Error codes mapped to exit codes
- [x] Conflict marker format specified
- [x] Test scenarios cover critical paths
- [ ] Principal Engineer can implement without architectural questions
- [ ] All artifacts verified via Read tool

---

## 14. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-manifest.md` | Write |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-ariadne.md` | Read |
| Spike | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-ariadne-go-cli-architecture.md` | Referenced |
| Session TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-session.md` | Read |
| Team TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-rite.md` | Read |
| Errors Package | `/Users/tomtenuta/Code/roster/ariadne/internal/errors/errors.go` | Read |
| Validation Package | `/Users/tomtenuta/Code/roster/ariadne/internal/validation/validator.go` | Read |
| Team Manifest | `/Users/tomtenuta/Code/roster/ariadne/internal/team/manifest.go` | Read |

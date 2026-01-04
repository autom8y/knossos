# TDD: Federated Artifact Registry

> Technical Design Document for the Federated Artifact Registry - Gap 2 of the Knossos Doctrine readiness sprint.

**Status**: Draft
**Author**: Context Architect
**Date**: 2026-01-05
**Sprint**: Knossos Doctrine Readiness
**Prerequisites**: Gap 1 (Orchestrator Stamping) - Complete

---

## 1. Overview

This Technical Design Document specifies the **Federated Artifact Registry** for Ariadne. The registry provides queryable indexes of workflow artifacts at both session and project levels, enabling cross-cutting analysis, throughline reconstruction, and artifact lifecycle tracking.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| Gap 1 Implementation | `ariadne/internal/hook/threadcontract/` |
| Session Context | `ariadne/internal/session/context.go` |
| Events System | `ariadne/internal/session/events.go` |
| Validation Package | `ariadne/internal/validation/artifact.go` |
| Paths Package | `ariadne/internal/paths/paths.go` |
| state-mate Agent | `user-agents/state-mate.md` |

### 1.2 Design Goals

1. **Federated Architecture**: Per-session registries aggregate to project-level index
2. **Event-Driven Registration**: Artifacts registered on `mark_complete` via state-mate
3. **Query Flexibility**: Support 4 query dimensions (phase, specialist, session, type)
4. **Minimal Overhead**: Registry operations must not block workflow execution
5. **Eventual Consistency**: Session-to-project aggregation is async, not blocking

### 1.3 Non-Goals

- Real-time synchronization across distributed systems
- Artifact content storage (registry stores metadata, not content)
- Version control integration (git handles artifact versioning)
- Full-text search within artifacts (use Grep for content search)

---

## 2. Schema Design

### 2.1 Artifact Entry Schema (v1.0)

Each artifact entry captures metadata sufficient for the 4 query patterns.

```yaml
# Artifact Entry - stored in artifacts.yaml
artifact_id: string            # Unique identifier (e.g., "PRD-artifact-registry")
artifact_type: enum            # prd | tdd | adr | test-plan | code | runbook
path: string                   # Relative path from project root
phase: enum                    # requirements | design | implementation | validation
specialist: string             # Agent that produced it (e.g., "product-owner", "architect")
session_id: string             # Session that created it
task_id: string                # Task marked complete (optional, for sprint tracking)
registered_at: string          # ISO 8601 timestamp
validated: boolean             # Whether schema validation passed
validation_issues: array       # List of validation issues (if any)
metadata: object               # Type-specific metadata (optional)
```

**Field Constraints**:

| Field | Required | Validation Rule |
|-------|----------|-----------------|
| `artifact_id` | Yes | Pattern: `^[A-Z]+-[a-z0-9-]+$` |
| `artifact_type` | Yes | Enum: prd, tdd, adr, test-plan, code, runbook |
| `path` | Yes | Must be relative, file must exist |
| `phase` | Yes | Enum: requirements, design, implementation, validation |
| `specialist` | Yes | Non-empty string |
| `session_id` | Yes | Pattern: `^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$` |
| `task_id` | No | Pattern if present: `^[a-z]+-[0-9]+$` |
| `registered_at` | Yes | ISO 8601 format |
| `validated` | Yes | Boolean |
| `validation_issues` | No | Array of strings (present if validated=false) |
| `metadata` | No | Object, type-specific extensions |

### 2.2 Session-Level Registry Schema

```yaml
# .claude/sessions/{session-id}/artifacts.yaml
schema_version: "1.0"
session_id: "session-20260105-143022-abc12345"
created_at: "2026-01-05T14:30:22Z"
updated_at: "2026-01-05T16:45:00Z"
artifact_count: 3
artifacts:
  - artifact_id: "PRD-artifact-registry"
    artifact_type: prd
    path: "docs/requirements/PRD-artifact-registry.md"
    phase: requirements
    specialist: "product-owner"
    session_id: "session-20260105-143022-abc12345"
    task_id: "task-001"
    registered_at: "2026-01-05T15:00:00Z"
    validated: true
    validation_issues: []
  - artifact_id: "TDD-artifact-registry"
    artifact_type: tdd
    path: "docs/design/TDD-artifact-registry.md"
    phase: design
    specialist: "context-architect"
    session_id: "session-20260105-143022-abc12345"
    task_id: "task-002"
    registered_at: "2026-01-05T16:45:00Z"
    validated: true
    validation_issues: []
```

### 2.3 Project-Level Registry Schema

```yaml
# .claude/artifacts/registry.yaml
schema_version: "1.0"
project_root: "/Users/tomtenuta/Code/roster/ariadne"
created_at: "2026-01-01T00:00:00Z"
updated_at: "2026-01-05T16:45:00Z"
artifact_count: 47
sessions_indexed: 12
artifacts:
  # Aggregated from all sessions
  - artifact_id: "PRD-artifact-registry"
    artifact_type: prd
    path: "docs/requirements/PRD-artifact-registry.md"
    phase: requirements
    specialist: "product-owner"
    session_id: "session-20260105-143022-abc12345"
    task_id: "task-001"
    registered_at: "2026-01-05T15:00:00Z"
    validated: true
    validation_issues: []
  # ... more artifacts from various sessions
indexes:
  by_phase:
    requirements: ["PRD-artifact-registry", "PRD-ariadne", ...]
    design: ["TDD-artifact-registry", "TDD-ariadne-session", ...]
    implementation: ["code-session-context", ...]
    validation: ["TEST-session-fsm", ...]
  by_type:
    prd: ["PRD-artifact-registry", "PRD-ariadne"]
    tdd: ["TDD-artifact-registry", "TDD-ariadne-session", ...]
    adr: ["ADR-0001-flock-locking", ...]
    test-plan: ["TEST-session-fsm"]
    code: ["code-session-context", ...]
    runbook: []
  by_specialist:
    product-owner: ["PRD-artifact-registry", "PRD-ariadne"]
    context-architect: ["TDD-artifact-registry", ...]
    integration-engineer: ["code-session-context", ...]
  by_session:
    "session-20260105-143022-abc12345": ["PRD-artifact-registry", "TDD-artifact-registry"]
    "session-20260104-160414-563c681e": ["PRD-ariadne", "TDD-ariadne-session", ...]
```

---

## 3. Go Implementation

### 3.1 Package Structure

```
ariadne/internal/artifact/
+-- registry.go        # Core types and CRUD operations
+-- aggregate.go       # Session -> project sync logic
+-- query.go           # Query interface implementation
+-- events.go          # Event emission for registry changes
+-- registry_test.go   # Unit tests
+-- aggregate_test.go  # Aggregation tests
+-- query_test.go      # Query tests
```

### 3.2 Core Types (registry.go)

```go
package artifact

import (
    "time"

    "github.com/autom8y/ariadne/internal/errors"
)

// ArtifactType represents the type of workflow artifact.
type ArtifactType string

const (
    TypePRD      ArtifactType = "prd"
    TypeTDD      ArtifactType = "tdd"
    TypeADR      ArtifactType = "adr"
    TypeTestPlan ArtifactType = "test-plan"
    TypeCode     ArtifactType = "code"
    TypeRunbook  ArtifactType = "runbook"
)

// Phase represents the workflow phase that produced the artifact.
type Phase string

const (
    PhaseRequirements   Phase = "requirements"
    PhaseDesign         Phase = "design"
    PhaseImplementation Phase = "implementation"
    PhaseValidation     Phase = "validation"
)

// Entry represents a single artifact in the registry.
type Entry struct {
    ArtifactID       string            `yaml:"artifact_id" json:"artifact_id"`
    ArtifactType     ArtifactType      `yaml:"artifact_type" json:"artifact_type"`
    Path             string            `yaml:"path" json:"path"`
    Phase            Phase             `yaml:"phase" json:"phase"`
    Specialist       string            `yaml:"specialist" json:"specialist"`
    SessionID        string            `yaml:"session_id" json:"session_id"`
    TaskID           string            `yaml:"task_id,omitempty" json:"task_id,omitempty"`
    RegisteredAt     time.Time         `yaml:"registered_at" json:"registered_at"`
    Validated        bool              `yaml:"validated" json:"validated"`
    ValidationIssues []string          `yaml:"validation_issues,omitempty" json:"validation_issues,omitempty"`
    Metadata         map[string]any    `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// SessionRegistry represents artifacts registered within a single session.
type SessionRegistry struct {
    SchemaVersion string    `yaml:"schema_version" json:"schema_version"`
    SessionID     string    `yaml:"session_id" json:"session_id"`
    CreatedAt     time.Time `yaml:"created_at" json:"created_at"`
    UpdatedAt     time.Time `yaml:"updated_at" json:"updated_at"`
    ArtifactCount int       `yaml:"artifact_count" json:"artifact_count"`
    Artifacts     []Entry   `yaml:"artifacts" json:"artifacts"`
}

// ProjectRegistry represents the aggregated project-level artifact index.
type ProjectRegistry struct {
    SchemaVersion    string            `yaml:"schema_version" json:"schema_version"`
    ProjectRoot      string            `yaml:"project_root" json:"project_root"`
    CreatedAt        time.Time         `yaml:"created_at" json:"created_at"`
    UpdatedAt        time.Time         `yaml:"updated_at" json:"updated_at"`
    ArtifactCount    int               `yaml:"artifact_count" json:"artifact_count"`
    SessionsIndexed  int               `yaml:"sessions_indexed" json:"sessions_indexed"`
    Artifacts        []Entry           `yaml:"artifacts" json:"artifacts"`
    Indexes          ProjectIndexes    `yaml:"indexes" json:"indexes"`
}

// ProjectIndexes contains pre-computed indexes for fast querying.
type ProjectIndexes struct {
    ByPhase      map[Phase][]string          `yaml:"by_phase" json:"by_phase"`
    ByType       map[ArtifactType][]string   `yaml:"by_type" json:"by_type"`
    BySpecialist map[string][]string         `yaml:"by_specialist" json:"by_specialist"`
    BySession    map[string][]string         `yaml:"by_session" json:"by_session"`
}
```

### 3.3 Registry Operations (registry.go continued)

```go
// Registry provides CRUD operations for artifact registries.
type Registry struct {
    projectRoot string
    paths       *paths.Resolver
}

// NewRegistry creates a new Registry for the given project root.
func NewRegistry(projectRoot string) *Registry {
    return &Registry{
        projectRoot: projectRoot,
        paths:       paths.NewResolver(projectRoot),
    }
}

// SessionRegistryPath returns the path to a session's artifacts.yaml.
func (r *Registry) SessionRegistryPath(sessionID string) string {
    return filepath.Join(r.paths.SessionDir(sessionID), "artifacts.yaml")
}

// ProjectRegistryPath returns the path to the project registry.
func (r *Registry) ProjectRegistryPath() string {
    return filepath.Join(r.paths.ClaudeDir(), "artifacts", "registry.yaml")
}

// LoadSessionRegistry loads the artifact registry for a session.
// Returns empty registry (not error) if file doesn't exist.
func (r *Registry) LoadSessionRegistry(sessionID string) (*SessionRegistry, error) {
    path := r.SessionRegistryPath(sessionID)

    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // Return empty registry
            return &SessionRegistry{
                SchemaVersion: "1.0",
                SessionID:     sessionID,
                CreatedAt:     time.Now().UTC(),
                UpdatedAt:     time.Now().UTC(),
                ArtifactCount: 0,
                Artifacts:     []Entry{},
            }, nil
        }
        return nil, errors.Wrap(errors.CodeGeneralError, "failed to read session registry", err)
    }

    var registry SessionRegistry
    if err := yaml.Unmarshal(data, &registry); err != nil {
        return nil, errors.Wrap(errors.CodeParseError, "failed to parse session registry", err)
    }

    return &registry, nil
}

// SaveSessionRegistry writes the session registry to disk.
func (r *Registry) SaveSessionRegistry(registry *SessionRegistry) error {
    path := r.SessionRegistryPath(registry.SessionID)

    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return errors.Wrap(errors.CodePermissionDenied, "failed to create registry directory", err)
    }

    // Update metadata
    registry.UpdatedAt = time.Now().UTC()
    registry.ArtifactCount = len(registry.Artifacts)

    data, err := yaml.Marshal(registry)
    if err != nil {
        return errors.Wrap(errors.CodeGeneralError, "failed to marshal session registry", err)
    }

    return os.WriteFile(path, data, 0644)
}

// Register adds an artifact to the session registry.
func (r *Registry) Register(sessionID string, entry Entry) error {
    registry, err := r.LoadSessionRegistry(sessionID)
    if err != nil {
        return err
    }

    // Check for duplicate
    for i, existing := range registry.Artifacts {
        if existing.ArtifactID == entry.ArtifactID {
            // Update existing entry
            registry.Artifacts[i] = entry
            return r.SaveSessionRegistry(registry)
        }
    }

    // Add new entry
    registry.Artifacts = append(registry.Artifacts, entry)
    return r.SaveSessionRegistry(registry)
}

// LoadProjectRegistry loads the project-level artifact index.
func (r *Registry) LoadProjectRegistry() (*ProjectRegistry, error) {
    path := r.ProjectRegistryPath()

    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return &ProjectRegistry{
                SchemaVersion:   "1.0",
                ProjectRoot:     r.projectRoot,
                CreatedAt:       time.Now().UTC(),
                UpdatedAt:       time.Now().UTC(),
                ArtifactCount:   0,
                SessionsIndexed: 0,
                Artifacts:       []Entry{},
                Indexes: ProjectIndexes{
                    ByPhase:      make(map[Phase][]string),
                    ByType:       make(map[ArtifactType][]string),
                    BySpecialist: make(map[string][]string),
                    BySession:    make(map[string][]string),
                },
            }, nil
        }
        return nil, errors.Wrap(errors.CodeGeneralError, "failed to read project registry", err)
    }

    var registry ProjectRegistry
    if err := yaml.Unmarshal(data, &registry); err != nil {
        return nil, errors.Wrap(errors.CodeParseError, "failed to parse project registry", err)
    }

    return &registry, nil
}

// SaveProjectRegistry writes the project registry to disk.
func (r *Registry) SaveProjectRegistry(registry *ProjectRegistry) error {
    path := r.ProjectRegistryPath()

    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return errors.Wrap(errors.CodePermissionDenied, "failed to create registry directory", err)
    }

    registry.UpdatedAt = time.Now().UTC()
    registry.ArtifactCount = len(registry.Artifacts)

    data, err := yaml.Marshal(registry)
    if err != nil {
        return errors.Wrap(errors.CodeGeneralError, "failed to marshal project registry", err)
    }

    return os.WriteFile(path, data, 0644)
}
```

### 3.4 Aggregation Logic (aggregate.go)

```go
package artifact

import (
    "os"
    "path/filepath"
    "strings"
)

// Aggregator handles session-to-project registry synchronization.
type Aggregator struct {
    registry *Registry
}

// NewAggregator creates a new Aggregator.
func NewAggregator(registry *Registry) *Aggregator {
    return &Aggregator{registry: registry}
}

// AggregateSession merges a session's artifacts into the project registry.
// This is the primary integration point called on mark_complete.
func (a *Aggregator) AggregateSession(sessionID string) error {
    // Load session registry
    sessionReg, err := a.registry.LoadSessionRegistry(sessionID)
    if err != nil {
        return err
    }

    if len(sessionReg.Artifacts) == 0 {
        return nil // Nothing to aggregate
    }

    // Load project registry
    projectReg, err := a.registry.LoadProjectRegistry()
    if err != nil {
        return err
    }

    // Build map of existing artifacts for deduplication
    existingMap := make(map[string]int)
    for i, entry := range projectReg.Artifacts {
        existingMap[entry.ArtifactID] = i
    }

    // Track sessions
    sessionsSet := make(map[string]bool)
    for _, entry := range projectReg.Artifacts {
        sessionsSet[entry.SessionID] = true
    }

    // Merge session artifacts
    for _, entry := range sessionReg.Artifacts {
        if idx, exists := existingMap[entry.ArtifactID]; exists {
            // Update existing entry
            projectReg.Artifacts[idx] = entry
        } else {
            // Add new entry
            projectReg.Artifacts = append(projectReg.Artifacts, entry)
        }
        sessionsSet[entry.SessionID] = true
    }

    // Rebuild indexes
    projectReg.Indexes = a.buildIndexes(projectReg.Artifacts)
    projectReg.SessionsIndexed = len(sessionsSet)

    return a.registry.SaveProjectRegistry(projectReg)
}

// AggregateAll rebuilds the project registry from all session registries.
// Use for recovery or initial index build.
func (a *Aggregator) AggregateAll() error {
    sessionsDir := a.registry.paths.SessionsDir()

    entries, err := os.ReadDir(sessionsDir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // No sessions yet
        }
        return errors.Wrap(errors.CodeGeneralError, "failed to read sessions directory", err)
    }

    // Start with fresh project registry
    projectReg := &ProjectRegistry{
        SchemaVersion:   "1.0",
        ProjectRoot:     a.registry.projectRoot,
        CreatedAt:       time.Now().UTC(),
        UpdatedAt:       time.Now().UTC(),
        Artifacts:       []Entry{},
        Indexes: ProjectIndexes{
            ByPhase:      make(map[Phase][]string),
            ByType:       make(map[ArtifactType][]string),
            BySpecialist: make(map[string][]string),
            BySession:    make(map[string][]string),
        },
    }

    sessionsSet := make(map[string]bool)

    for _, entry := range entries {
        if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "session-") {
            continue
        }

        sessionID := entry.Name()
        sessionReg, err := a.registry.LoadSessionRegistry(sessionID)
        if err != nil {
            // Log warning but continue
            continue
        }

        for _, artifact := range sessionReg.Artifacts {
            projectReg.Artifacts = append(projectReg.Artifacts, artifact)
            sessionsSet[sessionID] = true
        }
    }

    projectReg.Indexes = a.buildIndexes(projectReg.Artifacts)
    projectReg.SessionsIndexed = len(sessionsSet)

    return a.registry.SaveProjectRegistry(projectReg)
}

// buildIndexes creates the query indexes from artifact list.
func (a *Aggregator) buildIndexes(artifacts []Entry) ProjectIndexes {
    indexes := ProjectIndexes{
        ByPhase:      make(map[Phase][]string),
        ByType:       make(map[ArtifactType][]string),
        BySpecialist: make(map[string][]string),
        BySession:    make(map[string][]string),
    }

    for _, entry := range artifacts {
        id := entry.ArtifactID

        indexes.ByPhase[entry.Phase] = append(indexes.ByPhase[entry.Phase], id)
        indexes.ByType[entry.ArtifactType] = append(indexes.ByType[entry.ArtifactType], id)
        indexes.BySpecialist[entry.Specialist] = append(indexes.BySpecialist[entry.Specialist], id)
        indexes.BySession[entry.SessionID] = append(indexes.BySession[entry.SessionID], id)
    }

    return indexes
}
```

### 3.5 Query Interface (query.go)

```go
package artifact

// QueryFilter specifies filter criteria for artifact queries.
type QueryFilter struct {
    Phase      Phase        `json:"phase,omitempty"`
    Type       ArtifactType `json:"type,omitempty"`
    Specialist string       `json:"specialist,omitempty"`
    SessionID  string       `json:"session_id,omitempty"`
}

// QueryResult contains the results of an artifact query.
type QueryResult struct {
    Entries []Entry `json:"entries"`
    Count   int     `json:"count"`
    Filter  QueryFilter `json:"filter"`
}

// Querier provides query operations on the artifact registry.
type Querier struct {
    registry *Registry
}

// NewQuerier creates a new Querier.
func NewQuerier(registry *Registry) *Querier {
    return &Querier{registry: registry}
}

// Query executes a filtered query against the project registry.
// Multiple filter fields are ANDed together.
func (q *Querier) Query(filter QueryFilter) (*QueryResult, error) {
    projectReg, err := q.registry.LoadProjectRegistry()
    if err != nil {
        return nil, err
    }

    var matches []Entry

    for _, entry := range projectReg.Artifacts {
        if !q.matchesFilter(entry, filter) {
            continue
        }
        matches = append(matches, entry)
    }

    return &QueryResult{
        Entries: matches,
        Count:   len(matches),
        Filter:  filter,
    }, nil
}

// QueryByPhase returns artifacts from a specific workflow phase.
func (q *Querier) QueryByPhase(phase Phase) (*QueryResult, error) {
    return q.Query(QueryFilter{Phase: phase})
}

// QueryByType returns artifacts of a specific type.
func (q *Querier) QueryByType(artifactType ArtifactType) (*QueryResult, error) {
    return q.Query(QueryFilter{Type: artifactType})
}

// QueryBySpecialist returns artifacts produced by a specific agent.
func (q *Querier) QueryBySpecialist(specialist string) (*QueryResult, error) {
    return q.Query(QueryFilter{Specialist: specialist})
}

// QueryBySession returns artifacts from a specific session.
func (q *Querier) QueryBySession(sessionID string) (*QueryResult, error) {
    return q.Query(QueryFilter{SessionID: sessionID})
}

// matchesFilter checks if an entry matches all non-empty filter criteria.
func (q *Querier) matchesFilter(entry Entry, filter QueryFilter) bool {
    if filter.Phase != "" && entry.Phase != filter.Phase {
        return false
    }
    if filter.Type != "" && entry.ArtifactType != filter.Type {
        return false
    }
    if filter.Specialist != "" && entry.Specialist != filter.Specialist {
        return false
    }
    if filter.SessionID != "" && entry.SessionID != filter.SessionID {
        return false
    }
    return true
}

// ListPhases returns all phases with their artifact counts.
func (q *Querier) ListPhases() (map[Phase]int, error) {
    projectReg, err := q.registry.LoadProjectRegistry()
    if err != nil {
        return nil, err
    }

    counts := make(map[Phase]int)
    for phase, ids := range projectReg.Indexes.ByPhase {
        counts[phase] = len(ids)
    }
    return counts, nil
}

// ListTypes returns all types with their artifact counts.
func (q *Querier) ListTypes() (map[ArtifactType]int, error) {
    projectReg, err := q.registry.LoadProjectRegistry()
    if err != nil {
        return nil, err
    }

    counts := make(map[ArtifactType]int)
    for t, ids := range projectReg.Indexes.ByType {
        counts[t] = len(ids)
    }
    return counts, nil
}

// ListSpecialists returns all specialists with their artifact counts.
func (q *Querier) ListSpecialists() (map[string]int, error) {
    projectReg, err := q.registry.LoadProjectRegistry()
    if err != nil {
        return nil, err
    }

    counts := make(map[string]int)
    for s, ids := range projectReg.Indexes.BySpecialist {
        counts[s] = len(ids)
    }
    return counts, nil
}
```

---

## 4. state-mate Integration

### 4.1 mark_complete Extension

The `mark_complete` operation in state-mate is extended to trigger artifact registration.

**Current Behavior** (from state-mate.md):
```
mark_complete task_id artifact=path
```

**Extended Behavior**:
```
mark_complete task_id artifact=path [phase=PHASE] [specialist=AGENT]
```

### 4.2 Integration Flow

```
1. state-mate receives mark_complete command
2. state-mate validates artifact exists at path
3. state-mate invokes: ari artifact register --path=<path> --session=<id> --task=<task_id>
4. Ariadne:
   a. Detects artifact type from filename pattern
   b. Runs schema validation if applicable
   c. Determines phase from session context
   d. Registers to session artifacts.yaml
   e. Triggers aggregation to project registry
5. state-mate receives success response
6. state-mate updates SESSION_CONTEXT.md with completion
```

### 4.3 CLI Command: ari artifact register

```
ari artifact register --path=<path> --session=<session-id> [flags]

Flags:
  --path string          Absolute or relative path to artifact (required)
  --session string       Session ID (required)
  --task string          Task ID that produced this artifact
  --specialist string    Agent name (default: inferred from context)
  --phase string         Workflow phase (default: from session context)
  --skip-validation      Skip schema validation
  --skip-aggregate       Don't trigger project aggregation

Output (JSON):
{
  "artifact_id": "TDD-artifact-registry",
  "artifact_type": "tdd",
  "path": "docs/design/TDD-artifact-registry.md",
  "phase": "design",
  "specialist": "context-architect",
  "session_id": "session-20260105-143022-abc12345",
  "task_id": "task-002",
  "registered_at": "2026-01-05T16:45:00Z",
  "validated": true,
  "aggregated": true
}
```

### 4.4 state-mate Extension Point

Add to state-mate.md `mark_complete` operation:

```yaml
# In state_mate_extensions for mark_complete
post_hooks:
  - name: artifact_registration
    trigger: mark_complete
    when: artifact_path_provided
    command: |
      ari artifact register \
        --path="$ARTIFACT_PATH" \
        --session="$SESSION_ID" \
        --task="$TASK_ID" \
        --specialist="$CURRENT_AGENT" \
        --phase="$CURRENT_PHASE"
```

---

## 5. Hook Integration

### 5.1 PostToolUse Hook for Write/Edit on Artifacts

The artifact registration can also be triggered by PostToolUse hooks when artifact files are written.

**However**, per the user-confirmed requirements:
- **Registration Trigger**: On `mark_complete` only (not on file write)
- **Rationale**: Artifacts may be written incrementally; only mark_complete signals "done"

### 5.2 Aggregation Trigger

The aggregation from session to project registry fires **on mark_complete only**, as part of the `ari artifact register` command.

```go
// In cmd/artifact/register.go
func runRegister(cmd *cobra.Command, args []string) error {
    // ... register to session registry ...

    // Aggregate unless --skip-aggregate
    if !skipAggregate {
        aggregator := artifact.NewAggregator(registry)
        if err := aggregator.AggregateSession(sessionID); err != nil {
            // Log warning but don't fail registration
            log.Printf("warning: aggregation failed: %v", err)
        }
    }

    return nil
}
```

---

## 6. CLI Commands

### 6.1 Command: ari artifact register

Registers an artifact to the session registry.

See Section 4.3 for full specification.

### 6.2 Command: ari artifact query

Queries the project artifact registry.

```
ari artifact query [flags]

Flags:
  --phase string        Filter by phase (requirements, design, implementation, validation)
  --type string         Filter by type (prd, tdd, adr, test-plan, code, runbook)
  --specialist string   Filter by specialist agent
  --session string      Filter by session ID
  --limit int           Maximum entries to return (default: 50)
  --format string       Output format: json, yaml, table (default: json)

Examples:
  ari artifact query --phase=requirements
  ari artifact query --type=tdd --specialist=context-architect
  ari artifact query --session=session-20260105-143022-abc12345
```

**Output (JSON)**:
```json
{
  "entries": [
    {
      "artifact_id": "TDD-artifact-registry",
      "artifact_type": "tdd",
      "path": "docs/design/TDD-artifact-registry.md",
      "phase": "design",
      "specialist": "context-architect",
      "session_id": "session-20260105-143022-abc12345"
    }
  ],
  "count": 1,
  "filter": {
    "type": "tdd",
    "specialist": "context-architect"
  }
}
```

**Output (table)**:
```
ARTIFACT ID              TYPE   PHASE    SPECIALIST          SESSION
TDD-artifact-registry    tdd    design   context-architect   session-20260105-143022-abc12345

Total: 1 artifact(s)
```

### 6.3 Command: ari artifact list

Lists artifact counts by dimension.

```
ari artifact list [--by=DIMENSION]

Dimensions:
  phase       Group by workflow phase
  type        Group by artifact type
  specialist  Group by producing agent
  session     Group by session

Examples:
  ari artifact list --by=phase
  ari artifact list --by=specialist
```

**Output**:
```
PHASE            COUNT
requirements     12
design           18
implementation   8
validation       9

Total: 47 artifacts
```

### 6.4 Command: ari artifact rebuild

Rebuilds the project registry from all session registries.

```
ari artifact rebuild [--dry-run]

Flags:
  --dry-run    Show what would be rebuilt without writing

Output:
{
  "sessions_scanned": 12,
  "artifacts_indexed": 47,
  "rebuild_time_ms": 145,
  "dry_run": false
}
```

---

## 7. Error Handling

### 7.1 Error Codes

| Code | Exit | Name | Description |
|------|------|------|-------------|
| `ARTIFACT_NOT_FOUND` | 6 | Artifact Not Found | Artifact file does not exist |
| `ARTIFACT_INVALID` | 4 | Artifact Invalid | Artifact fails schema validation |
| `REGISTRY_CORRUPT` | 16 | Registry Corrupt | Registry YAML is malformed |
| `DUPLICATE_ARTIFACT` | 1 | Duplicate Artifact | Artifact ID already exists (warning, updates) |

### 7.2 Error Response Structure

```json
{
  "error": {
    "code": "ARTIFACT_NOT_FOUND",
    "message": "Artifact file not found: docs/requirements/PRD-missing.md",
    "details": {
      "path": "docs/requirements/PRD-missing.md"
    }
  }
}
```

---

## 8. Test Strategy

### 8.1 Unit Tests

| Package | Test Focus | Coverage Target |
|---------|------------|-----------------|
| `artifact` | Entry validation, type detection | 100% |
| `artifact` | Registry CRUD operations | 100% |
| `artifact` | Query filtering logic | 100% |
| `artifact` | Index building | 100% |

### 8.2 Integration Tests

| Test ID | Description | Expected Outcome |
|---------|-------------|------------------|
| `int_001` | Register artifact to empty session | Session registry created |
| `int_002` | Register duplicate artifact | Entry updated, not duplicated |
| `int_003` | Aggregate session to project | Project registry updated with indexes |
| `int_004` | Query by phase | Returns matching artifacts only |
| `int_005` | Query by multiple filters | AND logic applied correctly |
| `int_006` | Rebuild from multiple sessions | All artifacts indexed |
| `int_007` | Register with validation failure | Entry marked validated=false |

### 8.3 Test Fixtures

```
ariadne/testdata/
+-- artifact/
    +-- session-with-artifacts/
    |   +-- SESSION_CONTEXT.md
    |   +-- artifacts.yaml
    +-- empty-session/
    |   +-- SESSION_CONTEXT.md
    +-- project-registry/
        +-- registry.yaml
```

---

## 9. Backward Compatibility

### 9.1 Classification: COMPATIBLE

This feature is **fully backward compatible**:

1. **New files only**: Creates `artifacts.yaml` files, doesn't modify existing files
2. **Optional integration**: state-mate extension is additive
3. **Graceful degradation**: Sessions without registries return empty results
4. **No schema changes**: SESSION_CONTEXT.md unchanged

### 9.2 Migration Path

For existing sessions without artifact registries:
1. Run `ari artifact rebuild` to scan and index existing artifacts
2. Or let registries build incrementally as new artifacts are marked complete

---

## 10. File Paths

### 10.1 Session-Level

```
.claude/sessions/{session-id}/artifacts.yaml
```

### 10.2 Project-Level

```
.claude/artifacts/registry.yaml
```

### 10.3 Implementation Files

```
ariadne/internal/artifact/registry.go
ariadne/internal/artifact/aggregate.go
ariadne/internal/artifact/query.go
ariadne/internal/artifact/events.go
ariadne/internal/cmd/artifact/register.go
ariadne/internal/cmd/artifact/query.go
ariadne/internal/cmd/artifact/list.go
ariadne/internal/cmd/artifact/rebuild.go
ariadne/internal/cmd/artifact/artifact.go  # Parent command
```

---

## 11. Implementation Guidance

### 11.1 Recommended Order

1. **Week 1: Core Package**
   - `internal/artifact/registry.go` - Types and CRUD
   - `internal/artifact/query.go` - Query interface
   - Unit tests for both

2. **Week 2: Aggregation**
   - `internal/artifact/aggregate.go` - Session to project sync
   - `internal/artifact/events.go` - Event emission
   - Integration tests

3. **Week 3: CLI Commands**
   - `internal/cmd/artifact/register.go`
   - `internal/cmd/artifact/query.go`
   - `internal/cmd/artifact/list.go`
   - `internal/cmd/artifact/rebuild.go`

4. **Week 4: state-mate Integration**
   - Update state-mate.md with extension
   - End-to-end testing with mark_complete

---

## 12. Handoff Criteria

Ready for Implementation when:

- [x] Schema fully specified for session and project registries
- [x] Go types and interfaces defined
- [x] CRUD operations specified
- [x] Query interface covers all 4 query patterns
- [x] state-mate integration path documented
- [x] CLI commands specified
- [x] Error codes defined
- [x] Test matrix covers critical paths
- [x] Backward compatibility classified (COMPATIBLE)
- [x] Implementation order specified
- [ ] Integration Engineer can implement without architectural questions

---

## 13. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/ariadne/docs/design/TDD-artifact-registry.md` | Created |
| Session Context | `/Users/tomtenuta/Code/roster/ariadne/internal/session/context.go` | Read |
| Session Events | `/Users/tomtenuta/Code/roster/ariadne/internal/session/events.go` | Read |
| Thread Contract | `/Users/tomtenuta/Code/roster/ariadne/internal/hook/threadcontract/` | Read |
| Validation Package | `/Users/tomtenuta/Code/roster/ariadne/internal/validation/artifact.go` | Read |
| Paths Package | `/Users/tomtenuta/Code/roster/ariadne/internal/paths/paths.go` | Read |
| Errors Package | `/Users/tomtenuta/Code/roster/ariadne/internal/errors/errors.go` | Read |
| state-mate Agent | `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` | Read |
| TDD Template | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-session.md` | Read |

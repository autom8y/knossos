# TDD: TRIBUTE.md Auto-Generation (Minos Tribute Workflow)

## Overview

This Technical Design Document specifies the implementation of TRIBUTE.md, an auto-generated session summary produced at wrap time. In the Knossos mythology, King Minos demanded tribute from Athens; in our system, TRIBUTE.md is the "payment" for navigating the labyrinth--a comprehensive record of what was accomplished, what decisions were made, and what artifacts were produced.

TRIBUTE.md serves as both human-readable documentation of a completed session and machine-parseable metadata for analytics and future context loading.

## Context

| Reference | Location |
|-----------|----------|
| PRD | (This design derives from doctrinal requirements in knossos-doctrine.md) |
| Knossos Doctrine | `docs/philosophy/knossos-doctrine.md` |
| Session Manager | `.claude/hooks/lib/session-manager.sh` (lines 690-735) |
| Session FSM | `.claude/hooks/lib/session-fsm.sh` |
| Wrap Command (Go) | `ariadne/internal/cmd/session/wrap.go` |
| Events Package | `ariadne/internal/session/events.go` |
| Sails Generator | `ariadne/internal/sails/generator.go` |
| Session Context Schema | `schemas/artifacts/session-context.schema.json` |

### Problem Statement

When a session wraps, the following information is distributed across multiple sources:
- **SESSION_CONTEXT.md**: Initiative, complexity, phases completed
- **events.jsonl**: Decisions, handoffs, state transitions, file changes
- **WHITE_SAILS.yaml**: Confidence signal and proof status
- **Git log**: Commits made during the session timeframe

No single artifact summarizes "what happened in this session" in a human-readable, archivable format. When reviewing past sessions or onboarding new team members to an initiative, there is no quick reference for what was accomplished.

### Design Goals

1. **Single Summary Document**: One TRIBUTE.md per session capturing all significant outcomes
2. **Automatic Generation**: Generated at wrap time without manual intervention
3. **Mythological Alignment**: Conceptually owned by Atropos (the cutting Fate) as part of session termination
4. **Idempotent**: Re-running wrap regenerates TRIBUTE.md consistently
5. **Data Synthesis**: Aggregates from SESSION_CONTEXT, events.jsonl, WHITE_SAILS, and git history
6. **Human + Machine Readable**: Markdown format with structured sections for parsing

---

## System Design

### Architecture Diagram

```
                                  +------------------------+
                                  |    Wrap Command        |
                                  |  (ari session wrap)    |
                                  +-----------+------------+
                                              |
                    +-------------------------+-------------------------+
                    |                         |                         |
                    v                         v                         v
          +------------------+      +------------------+      +------------------+
          | SESSION_CONTEXT  |      |   events.jsonl   |      |  WHITE_SAILS.yaml|
          |      .md         |      |                  |      |                  |
          +--------+---------+      +--------+---------+      +--------+---------+
                   |                         |                         |
                   +-------------------------+-------------------------+
                                             |
                                             v
                              +-----------------------------+
                              |     Tribute Generator       |
                              |  +------------------------+ |
                              |  | Data Collector        | |
                              |  +------------------------+ |
                              |  | Git History Extractor | |
                              |  +------------------------+ |
                              |  | Markdown Renderer     | |
                              |  +------------------------+ |
                              +-------------+--------------+
                                            |
                                            v
                              +-----------------------------+
                              |       TRIBUTE.md            |
                              |  (session_dir/TRIBUTE.md)   |
                              +-----------------------------+
```

### Components

| Component | Responsibility | Technology | Location |
|-----------|---------------|------------|----------|
| **Tribute Generator** | Orchestrates TRIBUTE.md creation | Go | `ariadne/internal/tribute/generator.go` |
| **Data Collector** | Extracts data from events.jsonl | Go | Part of tribute package |
| **Git Extractor** | Queries git log for session commits | Go | Part of tribute package |
| **Markdown Renderer** | Produces final TRIBUTE.md content | Go | Part of tribute package |

### Integration Point

The Tribute Generator integrates into the existing wrap flow in `ariadne/internal/cmd/session/wrap.go`:

```go
// Current wrap flow (simplified):
// 1. Validate transition
// 2. Generate White Sails        <- Existing
// 3. Generate Tribute            <- NEW: Insert here
// 4. Update context to ARCHIVED
// 5. Emit events
// 6. Move to archive

// After White Sails generation, before archival:
tributeGen := tribute.NewGenerator(sessionDir)
tributeResult, tributeErr := tributeGen.Generate()
if tributeErr != nil {
    printer.VerboseLog("warn", "failed to generate tribute",
        map[string]interface{}{"error": tributeErr.Error()})
    // Non-blocking: wrap continues even if tribute fails
}
```

---

## TRIBUTE.md Schema Specification

### Structure

```markdown
---
schema_version: "1.0"
session_id: "session-20260106-123456-abcd1234"
initiative: "Feature X Implementation"
complexity: "MODULE"
generated_at: "2026-01-06T15:30:00Z"
duration_hours: 4.5
---

# Tribute: Feature X Implementation

> Session `session-20260106-123456-abcd1234` completed on 2026-01-06

## Summary

**Initiative**: Feature X Implementation
**Complexity**: MODULE (estimated 4-8 hours)
**Duration**: 4h 30m (2026-01-06T11:00:00Z to 2026-01-06T15:30:00Z)
**Team/Rite**: 10x-dev
**Final Phase**: validation
**Confidence Signal**: WHITE

## Artifacts Produced

| Type | Path | Status |
|------|------|--------|
| PRD | `docs/requirements/PRD-feature-x.md` | Approved |
| TDD | `docs/design/TDD-feature-x.md` | Approved |
| ADR | `docs/decisions/ADR-0015-feature-x-approach.md` | Accepted |
| Code | `src/features/feature-x/` | Implemented |
| Tests | `tests/feature-x/` | Passing |

## Decisions Made

| Timestamp | Decision | Rationale |
|-----------|----------|-----------|
| 2026-01-06T11:30:00Z | Use event sourcing for audit trail | Immutable log enables time-travel debugging |
| 2026-01-06T13:00:00Z | PostgreSQL over Redis for primary store | ACID compliance required for financial data |

## Phase Progression

```
requirements -(2h)-> design -(1h)-> implementation -(1h)-> validation -(0.5h)-> complete
```

| Phase | Started | Duration | Agent |
|-------|---------|----------|-------|
| requirements | 2026-01-06T11:00:00Z | 2h 00m | requirements-analyst |
| design | 2026-01-06T13:00:00Z | 1h 00m | architect |
| implementation | 2026-01-06T14:00:00Z | 1h 00m | principal-engineer |
| validation | 2026-01-06T15:00:00Z | 0h 30m | qa-adversary |

## Handoffs

| From | To | Timestamp | Notes |
|------|----|-----------|-------|
| requirements-analyst | architect | 2026-01-06T13:00:00Z | PRD approved |
| architect | principal-engineer | 2026-01-06T14:00:00Z | TDD approved |
| principal-engineer | qa-adversary | 2026-01-06T15:00:00Z | Implementation complete |

## Git Commits

| Hash | Message | Files Changed |
|------|---------|---------------|
| abc1234 | feat: implement feature X core logic | 5 |
| def5678 | test: add feature X unit tests | 3 |
| ghi9012 | docs: add ADR-0015 for feature X approach | 1 |

## White Sails Attestation

**Color**: WHITE
**Computed Base**: WHITE
**Proofs**:
- tests: PASS
- build: PASS
- lint: PASS
- adversarial: PASS
- integration: SKIP (not required for MODULE complexity)

## Metrics

| Metric | Value |
|--------|-------|
| Tool Calls | 156 |
| Events Recorded | 42 |
| Files Modified | 12 |
| Lines Added | 847 |
| Lines Removed | 23 |

## Notes

Any additional context from SESSION_CONTEXT.md body that was captured during the session.
```

### Frontmatter Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schema_version` | string | Yes | Schema version ("1.0") |
| `session_id` | string | Yes | Unique session identifier |
| `initiative` | string | Yes | Initiative name from SESSION_CONTEXT |
| `complexity` | string | Yes | Complexity tier (SCRIPT/MODULE/SERVICE/SYSTEM) |
| `generated_at` | string | Yes | ISO8601 timestamp of generation |
| `duration_hours` | number | No | Session duration in hours |

### Section Requirements

| Section | Required | Data Source |
|---------|----------|-------------|
| Summary | Yes | SESSION_CONTEXT.md |
| Artifacts Produced | Yes | events.jsonl (artifact_created events) |
| Decisions Made | Conditional | events.jsonl (decision events) - only if decisions exist |
| Phase Progression | Conditional | events.jsonl (phase transitions) - only if phase events exist |
| Handoffs | Conditional | events.jsonl (handoff events) - only if handoffs occurred |
| Git Commits | Conditional | Git log - only if commits exist in session timeframe |
| White Sails Attestation | Conditional | WHITE_SAILS.yaml - only if generated |
| Metrics | Yes | events.jsonl aggregate + THREAD_RECORD.ndjson |
| Notes | Conditional | SESSION_CONTEXT.md body - only if non-boilerplate content |

---

## Interface Contracts

### Tribute Generator API

```go
package tribute

import (
    "time"
)

// Generator creates TRIBUTE.md for a session.
type Generator struct {
    SessionPath string
    Now         func() time.Time
}

// GenerateResult contains the output of tribute generation.
type GenerateResult struct {
    FilePath    string
    SessionID   string
    Initiative  string
    Complexity  string
    Duration    time.Duration
    Artifacts   []Artifact
    Decisions   []Decision
    Phases      []PhaseRecord
    Handoffs    []Handoff
    Commits     []Commit
    SailsColor  string
    Metrics     Metrics
    GeneratedAt time.Time
}

// Artifact represents a produced artifact.
type Artifact struct {
    Type       string // PRD, TDD, ADR, Code, Tests, etc.
    Path       string
    Status     string // Created, Approved, Implemented, Passing
    Timestamp  time.Time
}

// Decision represents a recorded decision.
type Decision struct {
    Timestamp   time.Time
    Decision    string
    Rationale   string
    Rejected    []string
    Context     string
}

// PhaseRecord represents a workflow phase.
type PhaseRecord struct {
    Phase     string
    StartedAt time.Time
    Duration  time.Duration
    Agent     string
}

// Handoff represents an agent handoff.
type Handoff struct {
    From      string
    To        string
    Timestamp time.Time
    Notes     string
}

// Commit represents a git commit.
type Commit struct {
    Hash         string
    ShortHash    string
    Message      string
    FilesChanged int
    Timestamp    time.Time
}

// Metrics contains session metrics.
type Metrics struct {
    ToolCalls     int
    EventsRecorded int
    FilesModified  int
    LinesAdded     int
    LinesRemoved   int
}

// NewGenerator creates a new Generator for the given session.
func NewGenerator(sessionPath string) *Generator

// Generate creates TRIBUTE.md and returns the result.
func (g *Generator) Generate() (*GenerateResult, error)

// GenerateFromProject creates a Generator for the current session in a project.
func GenerateFromProject(projectRoot string) (*Generator, error)
```

### Data Extraction Functions

```go
// ExtractArtifacts parses events.jsonl for artifact_created events.
func ExtractArtifacts(eventsPath string) ([]Artifact, error)

// ExtractDecisions parses events.jsonl for decision events.
func ExtractDecisions(eventsPath string) ([]Decision, error)

// ExtractPhases parses events.jsonl for phase transition events.
func ExtractPhases(eventsPath string) ([]PhaseRecord, error)

// ExtractHandoffs parses events.jsonl for handoff events.
func ExtractHandoffs(eventsPath string) ([]Handoff, error)

// ExtractGitCommits queries git log for commits in the session timeframe.
func ExtractGitCommits(repoPath string, since, until time.Time) ([]Commit, error)

// ExtractMetrics aggregates metrics from events.jsonl and THREAD_RECORD.ndjson.
func ExtractMetrics(sessionPath string) (Metrics, error)
```

---

## Data Sources and Extraction

### 1. SESSION_CONTEXT.md

**Fields extracted:**
- `session_id`
- `initiative`
- `complexity`
- `active_rite`
- `created_at`
- `archived_at` (set by wrap)
- `current_phase` (final phase at wrap)
- Body content (for Notes section)

**Extraction method:**
```go
ctx, err := session.ParseContext(content)
```

### 2. events.jsonl

**Event types to extract:**

| Event Type | Maps To |
|------------|---------|
| `SESSION_CREATED` | Session start time |
| `SESSION_ARCHIVED` | Session end time |
| `PHASE_TRANSITIONED` | Phase progression |
| `artifact_created` | Artifacts produced |
| `decision` | Decisions made |
| `handoff_prepared` + `handoff_executed` | Handoffs |
| `tool_call` | Metrics (count) |
| `file_change` | Files modified |

**Event schema expectations:**
```json
// artifact_created event
{"ts":"2026-01-06T12:00:00Z","type":"artifact_created","path":"/path/to/PRD.md","artifact_type":"PRD"}

// decision event
{"ts":"2026-01-06T12:30:00Z","type":"decision","decision":"Use X over Y","rationale":"Because Z","rejected":["Y approach"]}

// handoff events
{"ts":"2026-01-06T13:00:00Z","type":"handoff_prepared","from":"analyst","to":"architect","notes":"PRD ready"}
{"ts":"2026-01-06T13:00:01Z","type":"handoff_executed","from":"analyst","to":"architect"}
```

### 3. WHITE_SAILS.yaml

**Fields extracted:**
- `color`
- `computed_base`
- `proofs` (with status for each)
- `open_questions`
- `complexity`

**Extraction method:**
```go
sailsContent, err := os.ReadFile(filepath.Join(sessionPath, "WHITE_SAILS.yaml"))
var sails sails.WhiteSailsYAML
yaml.Unmarshal(sailsContent, &sails)
```

### 4. Git History

**Query:**
```bash
git log --after="2026-01-06T11:00:00Z" --before="2026-01-06T15:30:00Z" --format="%H|%h|%s|%ci" --numstat
```

**Go implementation:**
```go
func ExtractGitCommits(repoPath string, since, until time.Time) ([]Commit, error) {
    cmd := exec.Command("git", "log",
        "--after="+since.Format(time.RFC3339),
        "--before="+until.Format(time.RFC3339),
        "--format=%H|%h|%s|%ci",
        "--numstat",
    )
    cmd.Dir = repoPath
    // Parse output...
}
```

### 5. THREAD_RECORD.ndjson (Optional)

**Fields extracted:**
- Tool call count
- Event timestamps for duration calculation

---

## Generation Pipeline

### Step-by-Step Flow

```
1. Load SESSION_CONTEXT.md
   |
   v
2. Extract session metadata (id, initiative, complexity, team, timestamps)
   |
   v
3. Load events.jsonl
   |
   v
4. Extract events by type:
   - Artifacts (artifact_created)
   - Decisions (decision)
   - Phase transitions (PHASE_TRANSITIONED)
   - Handoffs (handoff_prepared, handoff_executed)
   - File changes (file_change)
   - Tool calls (tool_call)
   |
   v
5. Load WHITE_SAILS.yaml (if exists)
   |
   v
6. Extract git commits in session timeframe
   |
   v
7. Calculate metrics:
   - Duration (archived_at - created_at)
   - Tool calls (count from events or THREAD_RECORD)
   - Files modified (from file_change events or git)
   - Lines changed (from git --numstat)
   |
   v
8. Render TRIBUTE.md using template
   |
   v
9. Write to session_dir/TRIBUTE.md
   |
   v
10. Return GenerateResult
```

### Error Handling Strategy

| Scenario | Behavior |
|----------|----------|
| SESSION_CONTEXT.md missing | Return error (required) |
| events.jsonl missing | Generate with empty events sections |
| WHITE_SAILS.yaml missing | Omit White Sails section |
| Git not available | Omit Git Commits section |
| THREAD_RECORD.ndjson missing | Use events.jsonl for tool call count |
| Parse errors in events | Skip malformed events, log warning |

### Idempotency

Re-running the generator produces identical output if:
- Session data hasn't changed
- Git history hasn't been rewritten
- Timestamps use UTC consistently

Implementation:
```go
func (g *Generator) Generate() (*GenerateResult, error) {
    // Remove existing TRIBUTE.md if present
    tributePath := filepath.Join(g.SessionPath, "TRIBUTE.md")
    os.Remove(tributePath) // Ignore error if doesn't exist

    // Generate fresh
    // ...
}
```

---

## Integration with Wrap Flow

### Shell Integration (session-manager.sh)

For the shell-based wrap flow in `mutate_wrap_fsm()`:

```bash
# In mutate_wrap_fsm() after FSM transition, before archival:

# Generate TRIBUTE.md via ari CLI
local tribute_result
if command -v ari >/dev/null 2>&1; then
    tribute_result=$(ari tribute generate --session-dir "$session_dir" 2>&1) || {
        # Non-blocking: log warning and continue
        echo "Warning: Failed to generate TRIBUTE.md: $tribute_result" >&2
    }
fi
```

### Go Integration (wrap.go)

```go
// After White Sails generation (line ~128 in wrap.go):

// Generate Tribute summary
tributeGen := tribute.NewGenerator(sessionDir)
tributeResult, tributeErr := tributeGen.Generate()
if tributeErr != nil {
    printer.VerboseLog("warn", "failed to generate tribute",
        map[string]interface{}{"error": tributeErr.Error()})
} else {
    printer.VerboseLog("info", "generated tribute",
        map[string]interface{}{"path": tributeResult.FilePath})
}
```

### CLI Subcommand

Add `ari tribute generate` command:

```go
// ariadne/internal/cmd/tribute/generate.go

func newGenerateCmd(ctx *cmdContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "generate",
        Short: "Generate TRIBUTE.md for a session",
        Long:  `Generates a summary document for a completed or active session.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runGenerate(ctx)
        },
    }
    return cmd
}
```

---

## Output Location

**Option Selected: A - Generate in session directory before archival**

Rationale:
1. TRIBUTE.md is part of the session record, should live with other session artifacts
2. Generated before archive move, ensuring it's included in the archived session
3. Simplifies retrieval: all session artifacts in one directory
4. Consistent with WHITE_SAILS.yaml placement

Location:
```
.claude/sessions/{session_id}/TRIBUTE.md          # Before archival
.claude/.archive/sessions/{session_id}/TRIBUTE.md # After archival
```

---

## Implementation Phases

### Phase 1: Core Generator (Shell-Compatible MVP)

**Scope:**
- Go package `ariadne/internal/tribute`
- Basic data extraction from SESSION_CONTEXT.md and events.jsonl
- Markdown template rendering
- CLI command `ari tribute generate`
- Integration into `ari session wrap`

**Deliverables:**
- `ariadne/internal/tribute/generator.go`
- `ariadne/internal/tribute/extractor.go`
- `ariadne/internal/tribute/renderer.go`
- `ariadne/internal/cmd/tribute/generate.go`
- Unit tests

**Estimated Effort:** 4-6 hours

### Phase 2: Git Integration

**Scope:**
- Git commit extraction for session timeframe
- Lines added/removed calculation
- File change statistics

**Deliverables:**
- `ariadne/internal/tribute/git.go`
- Extended TRIBUTE.md Git Commits section

**Estimated Effort:** 2-3 hours

### Phase 3: Enhanced Event Extraction

**Scope:**
- Decision event parsing with full schema
- Handoff correlation (prepared + executed)
- Phase progression timeline rendering

**Deliverables:**
- Enhanced event type handlers
- Mermaid diagram generation for phase flow (optional)

**Estimated Effort:** 2-3 hours

### Phase 4: Shell Fallback (Optional)

**Scope:**
- Shell-based tribute generation for environments without Go binary
- Minimal implementation using jq for event parsing

**Deliverables:**
- `.claude/hooks/lib/tribute-generator.sh`
- Integration into `session-manager.sh`

**Estimated Effort:** 3-4 hours (if needed)

---

## Test Strategy

### Unit Tests

| Test ID | Description | Location |
|---------|-------------|----------|
| `tribute_001` | Generator creates valid TRIBUTE.md from minimal context | `tribute/generator_test.go` |
| `tribute_002` | ExtractArtifacts parses artifact_created events | `tribute/extractor_test.go` |
| `tribute_003` | ExtractDecisions parses decision events | `tribute/extractor_test.go` |
| `tribute_004` | ExtractPhases builds timeline from transitions | `tribute/extractor_test.go` |
| `tribute_005` | ExtractHandoffs correlates prepared/executed pairs | `tribute/extractor_test.go` |
| `tribute_006` | Renderer produces valid markdown | `tribute/renderer_test.go` |
| `tribute_007` | Missing events.jsonl generates minimal tribute | `tribute/generator_test.go` |
| `tribute_008` | Idempotent generation produces identical output | `tribute/generator_test.go` |

### Integration Tests

| Test ID | Description |
|---------|-------------|
| `int_001` | Full wrap flow generates TRIBUTE.md in session directory |
| `int_002` | Archived session contains TRIBUTE.md |
| `int_003` | CLI command generates tribute for current session |
| `int_004` | Git commits extracted correctly for session timeframe |

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Events schema varies across sessions | Medium | Medium | Defensive parsing, skip unknown events |
| Git not available in environment | Low | Low | Skip Git section gracefully |
| Large sessions with many events | Low | Medium | Stream processing, not load all into memory |
| Timezone inconsistencies | Medium | Medium | Normalize all timestamps to UTC |
| Mythological naming confusion | Low | Low | Document Minos/Tribute mapping clearly |

---

## ADRs

### ADR: Tribute Generation Timing

**Status:** Proposed

**Context:** TRIBUTE.md must be generated during the wrap flow. Two options:
1. Generate before FSM transition (session still ACTIVE)
2. Generate after FSM transition (session ARCHIVED)

**Decision:** Generate after White Sails but before FSM transition to ARCHIVED.

**Rationale:**
- Session is still "alive" and data is complete
- Timestamps are finalized (including archived_at)
- Avoids needing to access archived session path

### ADR: Tribute Schema Versioning

**Status:** Proposed

**Context:** TRIBUTE.md needs a schema version for future evolution.

**Decision:** Use YAML frontmatter with `schema_version: "1.0"`.

**Rationale:**
- Consistent with SESSION_CONTEXT.md and WHITE_SAILS.yaml
- Enables migration tooling for future schema changes
- Human-readable header

---

## Open Items

| Item | Status | Owner | Notes |
|------|--------|-------|-------|
| Event schema standardization | Pending | Principal Engineer | Ensure consistent event types across hooks |
| Mermaid diagram support | Deferred | Future Sprint | Visual phase flow rendering |
| TRIBUTE.md validation schema | Deferred | Future Sprint | JSON Schema for machine validation |
| Historical backfill script | Deferred | Future Sprint | Generate tributes for archived sessions |

---

## Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-minos-tribute.md` | Pending (this document) |
| Knossos Doctrine | `/Users/tomtenuta/Code/roster/docs/philosophy/knossos-doctrine.md` | Read |
| Session Manager | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| Session FSM | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` | Read |
| Wrap Command | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap.go` | Read |
| Events Package | `/Users/tomtenuta/Code/roster/ariadne/internal/session/events.go` | Read |
| Sails Generator | `/Users/tomtenuta/Code/roster/ariadne/internal/sails/generator.go` | Read |
| Example SESSION_CONTEXT | `/Users/tomtenuta/Code/roster/.claude/.archive/sessions/session-20260105-163956-15ab643b/SESSION_CONTEXT.md` | Read |
| Example events.jsonl | `/Users/tomtenuta/Code/roster/.claude/.archive/sessions/session-20260105-163956-15ab643b/events.jsonl` | Read |
| Example WHITE_SAILS.yaml | `/Users/tomtenuta/Code/roster/.claude/.archive/sessions/session-20260105-163956-15ab643b/WHITE_SAILS.yaml` | Read |

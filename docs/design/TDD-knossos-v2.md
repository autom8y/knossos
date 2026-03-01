# TDD: Knossos Doctrine v2 - White Sails Typed Contract

> Umbrella Technical Design Document for confidence signaling, naming conventions, and remaining ecosystem gaps.

**Status**: Draft
**Author**: Context Architect
**Date**: 2026-01-05
**Initiative**: Knossos Doctrine v2: White Sails Typed Contract + Naming Convention Migration + Remaining Gaps

---

## 1. Overview

This Technical Design Document specifies the implementation of **Knossos Doctrine v2**, a comprehensive upgrade to the session confidence model. The core innovation is the **White Sails Typed Contract** -- a signal of confidence at session wrap that prevents "Aegeus failures" (false confidence leading to production issues).

### 1.1 The Aegeus Problem

In the original myth, Aegeus threw himself from a cliff when he saw black sails, believing his son Theseus dead. The signal was wrong -- Theseus forgot to raise white sails. Similarly, Claude Code sessions can ship with false confidence: tests pass but edge cases weren't considered, builds succeed but integration points were missed.

**White Sails solves this**: Every session produces a typed contract declaring the computed confidence level with explicit proof chains. "Any open question = gray ceiling" ensures honest signaling.

### 1.2 Scope

| Section | Scope |
|---------|-------|
| 1. White Sails Typed Contract | Schema, computation, proofs, state-mate integration |
| 2. Naming Convention Migration | SCREAMING_SNAKE audit, migration plan |
| 3. Remaining Gaps | `ari hook handoff-validate`, cross-rite integration |
| 4. Implementation Roadmap | Phases, files, tests, rollout |

### 1.3 Design Goals

1. **Prevent False Confidence**: Gray ceiling by default; white only with proof
2. **Integration-First**: Leverage existing ariadne architecture
3. **Backward Compatible**: Forward-only migration, existing sessions unaffected
4. **Human Override**: Declared modifiers allow justified exceptions
5. **Schema-Validated**: JSON Schema for WHITE_SAILS.yaml

---

## 2. Decision Matrix Summary

Per comprehensive interview, the following decisions are locked:

### 2.1 White Sails Core

| Dimension | Decision | Rationale |
|-----------|----------|-----------|
| Core Problem | False confidence (Aegeus failures) | Shipped sessions must signal honest confidence |
| Color Model | Hybrid: computed base + declared modifiers | Pure computation misses context; pure declaration games |
| Modifiers | Downgrade only + human override + context-aware thresholds | Cannot self-upgrade without QA; humans can justify gray |
| Proof Absence | Gray ceiling | Missing proof = unknown confidence = gray |
| Open Questions | Any open question = gray ceiling | Uncertainty must propagate |

### 2.2 Governance & QA

| Dimension | Decision | Rationale |
|-----------|----------|-----------|
| Governance Locus | At session wrap (internal to session) | External governance adds friction without value |
| QA Integration | QA session can upgrade gray to white | Independent validation earns confidence |
| Artifact Scope | Session-scoped, worst-color-wins aggregation | Sprint = collection of sails; initiative = fleet |
| Ownership | state-mate extension | Single authority for session mutations |
| Anti-Gaming | QA adversarial + trust debt (cultural norm) | Technical + cultural defense |

### 2.3 Proof Requirements

| Dimension | Decision | Rationale |
|-----------|----------|-----------|
| Proof Bar | Tests + build outputs, lint clean | Minimum verifiable evidence |
| QA Upgrade Requires | Constraint resolution log + adversarial tests | QA must document why gray -> white |
| Complexity Thresholds | Sliding scale (PATCH -> PLATFORM) | Higher complexity = stricter proof requirements |

### 2.4 Schema & Integration

| Dimension | Decision | Rationale |
|-----------|----------|-----------|
| Location | `.sos/sessions/{id}/WHITE_SAILS.yaml` | Co-located with SESSION_CONTEXT.md |
| Thread Integration | Terminal event + artifact | Record in events.jsonl + persist as file |
| Interruption Handling | Atomic -- no sails = session stays active | Cannot wrap without sails generation |
| Schema Format | YAML + JSON Schema validation | Human readable, machine validated |

### 2.5 Special Cases

| Dimension | Decision | Rationale |
|-----------|----------|-----------|
| Spikes | Gray ceiling with `type: spike` flag | Spikes produce learnings, not production code |
| Hotfix | Expedited gray + mandatory follow-up session | Fast path acknowledges risk explicitly |
| Migration | Forward-only, no backfill | Historical sessions stay as-is |

---

## 3. White Sails Typed Contract

### 3.1 Color Semantics

| Color | Meaning | Conditions |
|-------|---------|------------|
| **WHITE** | High confidence; ship without QA | All proofs present + tests pass + lint clean + no open questions + no TODOs |
| **GRAY** | Unknown confidence; needs QA | Missing proofs OR open questions OR complexity ceiling OR declared uncertainty |
| **BLACK** | Known failure; do not ship | Tests failing OR build broken OR explicit blocker |

**Key Insight**: There is no "yellow" or intermediate state. Either you have proof (white), you don't (gray), or you're broken (black). This simplicity prevents gaming.

### 3.2 Schema Definition

```yaml
# WHITE_SAILS.yaml JSON Schema (draft-2020-12)
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "embed:///schemas/white-sails.schema.json"
title: "WHITE_SAILS Schema"
description: "Session confidence signal per Knossos Doctrine v2"
type: object

required:
  - schema_version
  - session_id
  - generated_at
  - color
  - computed_base
  - proofs
  - open_questions

properties:
  schema_version:
    type: string
    pattern: "^[0-9]+\\.[0-9]+(\\.[0-9]+)?$"
    default: "1.0"

  session_id:
    type: string
    pattern: "^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$"

  generated_at:
    type: string
    format: date-time

  color:
    type: string
    enum: ["WHITE", "GRAY", "BLACK"]
    description: "Final confidence signal after modifiers"

  computed_base:
    type: string
    enum: ["WHITE", "GRAY", "BLACK"]
    description: "Computed color before human modifiers"

  proofs:
    type: object
    required: ["tests", "build", "lint"]
    properties:
      tests:
        $ref: "#/$defs/proof_item"
      build:
        $ref: "#/$defs/proof_item"
      lint:
        $ref: "#/$defs/proof_item"
      adversarial:
        $ref: "#/$defs/proof_item"
        description: "QA adversarial testing (required for gray->white upgrade)"
    additionalProperties:
      $ref: "#/$defs/proof_item"

  open_questions:
    type: array
    items:
      type: string
    description: "Any open question = gray ceiling"

  modifiers:
    type: array
    items:
      $ref: "#/$defs/modifier"
    description: "Human-declared adjustments with justification"

  complexity:
    type: string
    enum: ["PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION"]
    description: "Complexity tier affects proof thresholds"

  type:
    type: string
    enum: ["standard", "spike", "hotfix"]
    default: "standard"
    description: "Session type affects color ceiling"

  qa_upgrade:
    type: object
    description: "Present only if QA upgraded gray->white"
    properties:
      upgraded_at:
        type: string
        format: date-time
      qa_session_id:
        type: string
        pattern: "^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$"
      constraint_resolution_log:
        type: string
        description: "Path to QA findings document"
      adversarial_tests_added:
        type: array
        items:
          type: string
        description: "New tests added by QA"
    required: ["upgraded_at", "qa_session_id", "constraint_resolution_log"]

$defs:
  proof_item:
    type: object
    required: ["status"]
    properties:
      status:
        type: string
        enum: ["PASS", "FAIL", "SKIP", "UNKNOWN"]
      evidence_path:
        type: string
        description: "Path to output file or log"
      summary:
        type: string
        description: "Human-readable summary"
      exit_code:
        type: integer
      timestamp:
        type: string
        format: date-time

  modifier:
    type: object
    required: ["type", "justification", "applied_by"]
    properties:
      type:
        type: string
        enum: ["DOWNGRADE_TO_GRAY", "DOWNGRADE_TO_BLACK", "HUMAN_OVERRIDE_GRAY"]
      justification:
        type: string
        minLength: 10
        description: "Why this modifier is applied"
      applied_by:
        type: string
        enum: ["agent", "human"]
      timestamp:
        type: string
        format: date-time
```

### 3.3 Example WHITE_SAILS.yaml

```yaml
schema_version: "1.0"
session_id: "session-20260105-143000-abc12345"
generated_at: "2026-01-05T15:30:00Z"
color: "WHITE"
computed_base: "WHITE"
complexity: "MODULE"
type: "standard"

proofs:
  tests:
    status: "PASS"
    evidence_path: ".sos/sessions/session-20260105-143000-abc12345/test-output.log"
    summary: "47 tests passed, 0 failed, 0 skipped"
    exit_code: 0
    timestamp: "2026-01-05T15:28:00Z"
  build:
    status: "PASS"
    evidence_path: ".sos/sessions/session-20260105-143000-abc12345/build-output.log"
    summary: "go build succeeded"
    exit_code: 0
    timestamp: "2026-01-05T15:29:00Z"
  lint:
    status: "PASS"
    evidence_path: ".sos/sessions/session-20260105-143000-abc12345/lint-output.log"
    summary: "golangci-lint clean"
    exit_code: 0
    timestamp: "2026-01-05T15:29:30Z"

open_questions: []

modifiers: []
```

### 3.4 Example with Gray Ceiling

```yaml
schema_version: "1.0"
session_id: "session-20260105-143000-def67890"
generated_at: "2026-01-05T16:00:00Z"
color: "GRAY"
computed_base: "GRAY"
complexity: "SERVICE"
type: "standard"

proofs:
  tests:
    status: "PASS"
    evidence_path: ".sos/sessions/session-20260105-143000-def67890/test-output.log"
    summary: "89 tests passed"
    exit_code: 0
  build:
    status: "PASS"
    evidence_path: ".sos/sessions/session-20260105-143000-def67890/build-output.log"
    summary: "docker build succeeded"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "eslint clean"
    exit_code: 0

open_questions:
  - "How should rate limiting behave under cluster failover?"
  - "Need to validate with Production DBA on index strategy"

modifiers: []
```

### 3.5 Example with Human Override

```yaml
schema_version: "1.0"
session_id: "session-20260105-143000-ghi11111"
generated_at: "2026-01-05T17:00:00Z"
color: "GRAY"
computed_base: "WHITE"
complexity: "PATCH"
type: "standard"

proofs:
  tests:
    status: "PASS"
    summary: "All tests pass"
    exit_code: 0
  build:
    status: "PASS"
    summary: "Build clean"
    exit_code: 0
  lint:
    status: "PASS"
    summary: "Lint clean"
    exit_code: 0

open_questions: []

modifiers:
  - type: "DOWNGRADE_TO_GRAY"
    justification: "Changed retry logic in payment flow; want senior review before shipping despite passing tests"
    applied_by: "human"
    timestamp: "2026-01-05T16:55:00Z"
```

---

## 4. Color Computation Algorithm

### 4.1 Algorithm Specification

```
FUNCTION compute_sails_color(session_context, proofs, open_questions, modifiers):

    # Step 1: Check for failures (BLACK)
    IF any proof.status == "FAIL":
        RETURN "BLACK"

    # Step 2: Check for open questions (GRAY ceiling)
    IF open_questions.length > 0:
        computed_base = "GRAY"
        GOTO apply_modifiers

    # Step 3: Check session type ceiling
    IF session_context.type == "spike":
        computed_base = "GRAY"  # Spikes never white
        GOTO apply_modifiers

    IF session_context.type == "hotfix":
        computed_base = "GRAY"  # Hotfix expedited gray
        GOTO apply_modifiers

    # Step 4: Check proof completeness per complexity
    required_proofs = get_required_proofs(session_context.complexity)

    FOR proof_name IN required_proofs:
        IF proofs[proof_name].status NOT IN ["PASS", "SKIP"]:
            computed_base = "GRAY"
            GOTO apply_modifiers

    # Step 5: All proofs present and passing
    computed_base = "WHITE"

    apply_modifiers:

    # Step 6: Apply modifiers (downgrade only, unless QA upgrade)
    final_color = computed_base

    FOR modifier IN modifiers:
        CASE modifier.type:
            "DOWNGRADE_TO_GRAY":
                IF final_color == "WHITE":
                    final_color = "GRAY"
            "DOWNGRADE_TO_BLACK":
                final_color = "BLACK"
            "HUMAN_OVERRIDE_GRAY":
                # Human can force gray even if would be white
                final_color = "GRAY"

    # Step 7: QA upgrade path (only applicable from GRAY)
    IF session_context.qa_upgrade IS NOT NULL:
        IF computed_base == "GRAY" AND final_color == "GRAY":
            IF qa_upgrade.constraint_resolution_log EXISTS:
                IF qa_upgrade.adversarial_tests_added.length > 0:
                    final_color = "WHITE"

    RETURN final_color
```

### 4.2 Required Proofs by Complexity

| Complexity | tests | build | lint | adversarial | integration |
|------------|-------|-------|------|-------------|-------------|
| PATCH | Required | Required | Required | - | - |
| MODULE | Required | Required | Required | - | - |
| SERVICE | Required | Required | Required | Recommended | Recommended |
| INITIATIVE | Required | Required | Required | Required | Required |
| MIGRATION | Required | Required | Required | Required | Required |

### 4.3 Go Implementation Location

```
ariadne/internal/sails/
    color.go           # Color computation algorithm
    color_test.go      # Unit tests for all color paths
    schema.go          # WHITE_SAILS.yaml parsing/serialization
    schema_test.go     # Schema validation tests
    generator.go       # Integration with /wrap flow
```

---

## 5. state-mate Integration

### 5.1 New Operation: `generate_sails`

state-mate receives a new high-level operation for White Sails generation:

```
Operation: generate_sails
Syntax: generate_sails [--skip-proofs] [--modifier=TYPE:JUSTIFICATION]
```

### 5.2 Integration Flow

```
/wrap invoked
    |
    v
state-mate.generate_sails()
    |
    +-- 1. Collect proofs from session directory
    |       - Read test output logs
    |       - Read build output logs
    |       - Read lint output logs
    |
    +-- 2. Gather open questions
    |       - Parse SESSION_CONTEXT.md body for "?" patterns
    |       - Check for explicit open_questions section
    |
    +-- 3. Apply any declared modifiers
    |
    +-- 4. Compute color via algorithm
    |
    +-- 5. Generate WHITE_SAILS.yaml
    |
    +-- 6. Emit SAILS_GENERATED event to events.jsonl
    |
    +-- 7. Return result to /wrap
    |
    v
/wrap continues (or blocks if BLACK)
```

### 5.3 state-mate Extension

Add to `user-agents/state-mate.md`:

```yaml
# White Sails Operations (Knossos v2)

### generate_sails

Generates WHITE_SAILS.yaml for session wrap.

**Syntax**:
```
generate_sails [--skip-proofs] [--modifier=TYPE:JUSTIFICATION]
```

**Parameters**:
- `--skip-proofs`: Skip proof collection (for spike sessions)
- `--modifier`: Apply modifier with justification

**Output**:
```json
{
  "success": true,
  "operation": "generate_sails",
  "sails_path": ".sos/sessions/{session-id}/WHITE_SAILS.yaml",
  "color": "WHITE",
  "computed_base": "WHITE",
  "proofs_collected": ["tests", "build", "lint"],
  "open_questions_found": 0
}
```

**Error Codes**:
- `SAILS_GENERATION_FAILED`: Could not collect required proofs
- `SCHEMA_VIOLATION`: Generated YAML fails schema validation
```

### 5.4 Events

New event type for Thread Contract:

```go
// ariadne/internal/hook/threadcontract/event.go

const (
    EventTypeSailsGenerated EventType = "sails_generated"
)

// NewSailsGeneratedEvent creates an event for White Sails generation.
func NewSailsGeneratedEvent(sessionID string, color string, meta map[string]interface{}) Event {
    return Event{
        Timestamp: timestamp(),
        Type:      EventTypeSailsGenerated,
        Summary:   fmt.Sprintf("Generated WHITE_SAILS: %s", color),
        Meta:      meta,
    }
}
```

---

## 6. /wrap Skill Integration

### 6.1 Modified /wrap Flow

Update `user-skills/session-lifecycle/wrap-ref/behavior.md`:

```markdown
## Wrap Behavior with White Sails

### Step 6.5: Generate White Sails (NEW)

After quality gates pass and before archiving:

1. Invoke state-mate with `generate_sails`
2. Wait for WHITE_SAILS.yaml generation
3. Check color:
   - WHITE: Proceed to archive
   - GRAY: Display warning, offer QA handoff or --accept-gray
   - BLACK: Block wrap, display failures

### Flags

| Flag | Description |
|------|-------------|
| `--accept-gray` | Accept GRAY sails without QA review |
| `--skip-sails` | Skip sails generation (equivalent to spike) |

### Gray Warning

```
[WARNING] Session wrapped with GRAY sails.

Confidence: GRAY (2 open questions found)
- How should rate limiting behave under cluster failover?
- Need to validate with Production DBA on index strategy

Recommendation: Run /qa to earn WHITE sails, or use --accept-gray to proceed.
```
```

### 6.2 Quality Gate Integration

Sails generation becomes a quality gate:

```markdown
## Quality Gates (Updated)

| Gate | Applies To | Checks |
|------|-----------|--------|
| PRD | All | File exists, sections complete |
| TDD | MODULE+ | Traces to PRD, ADRs exist |
| Code | Implementation phase | Git clean, tests pass |
| Sails | All | WHITE_SAILS.yaml generated, not BLACK |
```

---

## 7. QA Upgrade Flow

### 7.1 Prerequisites

A GRAY session can be upgraded to WHITE only through an independent QA session:

1. Original session wrapped with GRAY sails
2. New QA session created: `/start "QA: {original initiative}" --complexity=PATCH --team=10x-dev`
3. QA session references original session ID
4. QA adversary agent runs adversarial validation
5. If issues found: document in constraint resolution log
6. If issues resolved: add adversarial tests
7. QA session wraps with upgrade declaration

### 7.2 QA Session Sails

```yaml
schema_version: "1.0"
session_id: "session-20260106-100000-qa123456"
generated_at: "2026-01-06T12:00:00Z"
color: "WHITE"
computed_base: "WHITE"
complexity: "PATCH"
type: "standard"

proofs:
  tests:
    status: "PASS"
    summary: "Original tests + 3 adversarial tests pass"
    exit_code: 0
  adversarial:
    status: "PASS"
    summary: "Edge cases validated: rate limiting, failover, index"
    evidence_path: "docs/testing/TP-qa-original-session.md"

open_questions: []

# This upgrades the original session
qa_upgrade:
  upgraded_at: "2026-01-06T12:00:00Z"
  qa_session_id: "session-20260106-100000-qa123456"
  original_session_id: "session-20260105-143000-def67890"
  constraint_resolution_log: "docs/testing/TP-qa-original-session.md"
  adversarial_tests_added:
    - "tests/integration/rate_limit_failover_test.go"
    - "tests/integration/index_edge_cases_test.go"
```

### 7.3 Original Session Update

When QA upgrades, the original session's WHITE_SAILS.yaml is updated:

```yaml
# Updated WHITE_SAILS.yaml for original session
color: "WHITE"  # Changed from GRAY
computed_base: "GRAY"  # Unchanged
qa_upgrade:
  upgraded_at: "2026-01-06T12:00:00Z"
  qa_session_id: "session-20260106-100000-qa123456"
  constraint_resolution_log: "docs/testing/TP-qa-original-session.md"
  adversarial_tests_added:
    - "tests/integration/rate_limit_failover_test.go"
    - "tests/integration/index_edge_cases_test.go"
```

---

## 8. Naming Convention Migration

### 8.1 Current State Audit

Session artifacts use SCREAMING_SNAKE convention:

| File | Current | Status |
|------|---------|--------|
| SESSION_CONTEXT.md | SCREAMING_SNAKE | Correct |
| SPRINT_CONTEXT.md | SCREAMING_SNAKE | Correct |
| WHITE_SAILS.yaml | SCREAMING_SNAKE | New (Correct) |
| events.jsonl | lowercase | Correct (log file) |
| artifacts.yaml | lowercase | Correct (data file) |

### 8.2 Naming Rules

```
RULE: Session-scoped policy files use SCREAMING_SNAKE_CASE
RULE: Data/log files use lowercase with extension
RULE: Index files use ALL_CAPS (e.g., INDEX.md)

Session Directory Structure:
.sos/sessions/{session-id}/
    SESSION_CONTEXT.md     # Policy: session state
    SPRINT_CONTEXT.md      # Policy: sprint state
    WHITE_SAILS.yaml       # Policy: confidence signal
    events.jsonl           # Data: event log
    artifacts.yaml         # Data: artifact registry
    test-output.log        # Data: test results
    build-output.log       # Data: build output
```

### 8.3 No Migration Needed

Current naming is already consistent. New WHITE_SAILS.yaml follows convention.

---

## 9. Remaining Knossos Gaps

### 9.1 Gap: `ari hook handoff-validate`

**Current State**: `ari handoff` commands exist but are not implemented:
- `ari handoff prepare` - "not yet implemented"
- `ari handoff execute` - "not yet implemented"
- `ari handoff status` - "not yet implemented"
- `ari handoff history` - "not yet implemented"

**Required**: Implement `ari hook handoff-validate` for cross-rite handoff validation.

### 9.2 Handoff Validation Specification

```
Command: ari hook handoff-validate
Purpose: Validate artifact readiness for cross-rite handoff
Hook Event: PreToolUse (when Task tool invoked with handoff agent)

Inputs:
  - from_agent: Source agent identifier
  - to_agent: Target agent identifier
  - artifact_path: Path to artifact being handed off
  - phase: Current workflow phase

Outputs (JSON):
{
  "valid": true|false,
  "from_agent": "architect",
  "to_agent": "principal-engineer",
  "artifact_path": "docs/design/TDD-feature.md",
  "phase": "design",
  "criteria_results": [
    {"criterion": "title", "passed": true},
    {"criterion": "acceptance_criteria", "passed": true},
    {"criterion": "interfaces_defined", "passed": false, "message": "Missing interfaces section"}
  ],
  "blocking_failures": 1,
  "warnings": 0
}

Exit Codes:
  0: Handoff valid
  5: Handoff blocked (LIFECYCLE_VIOLATION)
  6: Artifact not found
```

### 9.3 Implementation Location

```
ariadne/internal/cmd/hook/
    handoff_validate.go      # ari hook handoff-validate
    handoff_validate_test.go # Unit tests
```

### 9.4 Hook Integration

New hook in `.claude/hooks/`:

```yaml
# .claude/hooks/handoff-validate.yaml
name: handoff-validate
event: PreToolUse
pattern:
  tool: Task
  contains: "handoff"
command: ari hook handoff-validate
```

### 9.5 Cross-Team Handoff Integration

Integrate with existing cross-rite skill:

```
user-skills/guidance/cross-rite/
    validation.md            # Existing validation rules
    SKILL.md                 # Skill entry point

rites/shared/skills/cross-rite-handoff/
    SKILL.md                 # Shared skill
    schema.md                # Handoff schema
    validation.sh            # Shell validator (to be replaced by ari)
```

**Migration**: Replace `validation.sh` calls with `ari hook handoff-validate`.

---

## 10. Implementation Roadmap

### 10.1 Phase 1: Schema Foundation (Week 1)

| Task | Owner | Files | Deliverable |
|------|-------|-------|-------------|
| WHITE_SAILS.yaml schema | Integration Engineer | `ariadne/internal/validation/schemas/white-sails.schema.json` | JSON Schema |
| Schema embedding | Integration Engineer | `ariadne/internal/validation/loader.go` | Embedded FS |
| Schema validation tests | Integration Engineer | `ariadne/internal/validation/sails_test.go` | 100% coverage |

### 10.2 Phase 2: Color Computation (Week 2)

| Task | Owner | Files | Deliverable |
|------|-------|-------|-------------|
| Color algorithm | Integration Engineer | `ariadne/internal/sails/color.go` | Algorithm impl |
| Complexity thresholds | Integration Engineer | `ariadne/internal/sails/thresholds.go` | Threshold matrix |
| Proof collection | Integration Engineer | `ariadne/internal/sails/proofs.go` | Proof collector |
| Unit tests | Integration Engineer | `ariadne/internal/sails/*_test.go` | 100% coverage |

### 10.3 Phase 3: state-mate Integration (Week 3)

| Task | Owner | Files | Deliverable |
|------|-------|-------|-------------|
| generate_sails operation | Integration Engineer | `user-agents/state-mate.md` | Updated agent |
| Thread Contract event | Integration Engineer | `ariadne/internal/hook/threadcontract/event.go` | New event type |
| Integration test | Integration Engineer | `tests/integration/sails_test.go` | E2E test |

### 10.4 Phase 4: /wrap Integration (Week 4)

| Task | Owner | Files | Deliverable |
|------|-------|-------|-------------|
| /wrap behavior update | Integration Engineer | `user-skills/session-lifecycle/wrap-ref/behavior.md` | Updated skill |
| Quality gate integration | Integration Engineer | `user-skills/session-lifecycle/wrap-ref/quality-gates.md` | Updated gates |
| User documentation | Documentation Engineer | `docs/guides/white-sails.md` | User guide |

### 10.5 Phase 5: Handoff Validation (Week 5)

| Task | Owner | Files | Deliverable |
|------|-------|-------|-------------|
| ari hook handoff-validate | Integration Engineer | `ariadne/internal/cmd/hook/handoff_validate.go` | CLI command |
| Hook registration | Integration Engineer | `.claude/hooks/handoff-validate.yaml` | Hook config |
| Cross-team integration | Integration Engineer | `rites/shared/skills/cross-rite-handoff/` | Updated skill |

### 10.6 Phase 6: Validation & Rollout (Week 6)

| Task | Owner | Files | Deliverable |
|------|-------|-------|-------------|
| Compatibility testing | Compatibility Tester | `docs/qa/COMPATIBILITY-REPORT-knossos-v2.md` | Report |
| Migration runbook | Documentation Engineer | `docs/migrations/RUNBOOK-knossos-v2.md` | Runbook |
| Rollout | All | N/A | Production deployment |

---

## 11. Test Strategy

### 11.1 Unit Tests

| Package | Test Focus | Coverage |
|---------|-----------|----------|
| `sails/color` | All color computation paths | 100% |
| `sails/schema` | YAML parsing, validation | 100% |
| `sails/proofs` | Proof collection from logs | 100% |
| `validation/sails` | Schema validation | 100% |

### 11.2 Integration Tests

| Test ID | Description | TLA+ Property |
|---------|-------------|---------------|
| `sails_001` | WHITE with all proofs passing | ColorComputation |
| `sails_002` | GRAY with open questions | GrayCeiling |
| `sails_003` | GRAY with missing proofs | ProofRequirement |
| `sails_004` | BLACK with failing tests | FailureDetection |
| `sails_005` | Spike always GRAY | TypeCeiling |
| `sails_006` | Hotfix always GRAY | TypeCeiling |
| `sails_007` | Human downgrade override | ModifierApplication |
| `sails_008` | QA upgrade gray to white | QAUpgrade |
| `sails_009` | Cannot self-upgrade | NoSelfUpgrade |

### 11.3 Satellite Compatibility Matrix

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| baseline | Generate sails for MODULE | WHITE if proofs pass |
| minimal | Generate sails with no tests | GRAY (missing proof) |
| complex | Generate sails for PLATFORM | WHITE requires adversarial |
| spike | Generate sails for spike | Always GRAY |
| hotfix | Generate sails for hotfix | Always GRAY |

---

## 12. Backward Compatibility

### 12.1 Classification: COMPATIBLE

White Sails is a new feature. Existing sessions without WHITE_SAILS.yaml are unaffected.

### 12.2 Migration Behavior

| Scenario | Behavior |
|----------|----------|
| Old session without sails | No sails file; treated as legacy |
| New session wrapped | WHITE_SAILS.yaml generated |
| Resume old session and wrap | WHITE_SAILS.yaml generated on wrap |

### 12.3 Schema Version

WHITE_SAILS.yaml starts at schema_version "1.0". Future changes follow semver.

---

## 13. Anti-Gaming Mechanisms

### 13.1 Technical

1. **Cannot self-upgrade**: Modifiers can only downgrade, never upgrade
2. **QA upgrade requires proof**: constraint_resolution_log + adversarial_tests_added
3. **Open questions propagate**: Any "?" in context triggers gray ceiling
4. **Proof verification**: Proofs must have evidence_path or UNKNOWN status

### 13.2 Cultural

1. **Trust debt**: Teams that ship gray without QA accumulate reputation cost
2. **QA adversarial**: QA agent explicitly tries to break assumptions
3. **Visibility**: WHITE_SAILS.yaml is committed, auditable

---

## 14. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Over-gray (always gray) | Medium | Medium | Tune thresholds per complexity; allow --accept-gray |
| Gaming via modifier abuse | Low | Medium | Audit trail; cultural norm against abuse |
| QA bottleneck | Medium | High | QA upgrade is optional; gray ships with warning |
| Schema evolution | Low | Medium | Additive changes only; version field |
| Proof collection failure | Low | Medium | Graceful degradation to GRAY |

---

## 15. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-knossos-001 | Proposed | Three-color model selection |
| ADR-knossos-002 | Proposed | QA upgrade mechanism |
| ADR-knossos-003 | Proposed | state-mate as sails authority |

---

## 16. Handoff Criteria

Ready for Implementation when:

- [x] White Sails schema defined with JSON Schema
- [x] Color computation algorithm specified
- [x] state-mate integration designed
- [x] /wrap integration specified
- [x] QA upgrade flow documented
- [x] ari hook handoff-validate specified
- [x] Test matrix covers all color paths
- [x] Backward compatibility confirmed (COMPATIBLE)
- [ ] All ADRs approved
- [ ] Implementation roadmap reviewed

---

## 17. Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-knossos-v2.md` | This document |
| state-mate | `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` | Read |
| Session FSM | `/Users/tomtenuta/Code/roster/ariadne/internal/session/fsm.go` | Read |
| Session Context | `/Users/tomtenuta/Code/roster/ariadne/internal/session/context.go` | Read |
| Thread Contract Events | `/Users/tomtenuta/Code/roster/ariadne/internal/hook/threadcontract/event.go` | Read |
| Artifact Registry | `/Users/tomtenuta/Code/roster/ariadne/internal/artifact/registry.go` | Read |
| Handoff Validation | `/Users/tomtenuta/Code/roster/ariadne/internal/validation/handoff.go` | Read |
| Handoff Commands | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/handoff/handoff.go` | Read |
| Wrap Skill | `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/wrap-ref/SKILL.md` | Read |
| Complexity Levels | `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/session-common/complexity-levels.md` | Read |
| Session Schema | `/Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/session-context.schema.json` | Read |

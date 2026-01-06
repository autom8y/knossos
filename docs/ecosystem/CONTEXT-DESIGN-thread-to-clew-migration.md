# Context Design: Thread to Clew Terminology Migration

**Date**: 2026-01-06
**Architect**: Context Architect
**Reference**: `docs/philosophy/knossos-doctrine.md` Section II (The Clew)
**Prerequisite**: This design is independent of the team-to-rite migration

---

## Executive Summary

The Knossos Doctrine establishes "clew" as the correct mythological term for the navigation mechanism:

> **Ariadne** gave Theseus the gift that saved him--not a weapon, but a clew (a ball of thread). The clew did not kill the Minotaur. The clew ensured return.

The codebase currently uses "thread" in multiple contexts. This design categorizes each usage and defines the migration path for doctrine-aligned terminology.

---

## Terminology Clarification

| Term | Mythological Meaning | Technical Use Case |
|------|---------------------|-------------------|
| **clew** | Ball of thread Ariadne gave Theseus | Session event trail, breadcrumb path |
| **thread** (semantic) | Concurrency/execution thread | Goroutines, concurrent execution |
| **main thread** | NOT mythological | Claude Code main conversation context |

**Key Distinction**: "Main thread" refers to the Claude Code conversation context (Theseus), not the clew. This is semantic usage describing the execution model, not mythological terminology.

---

## Audit Results

### Category: RENAME (Doctrine Alignment)

These references use "thread" to mean "clew" (the navigation trail) and should migrate:

| File | Line | Current | Proposed | Rationale |
|------|------|---------|----------|-----------|
| `/Users/tomtenuta/Code/roster/ariadne/internal/inscription/generator.go` | 260 | `"ariadne": "CLI binary (\`ari\`) - the thread ensuring return"` | `"ariadne": "CLI binary (\`ari\`) - the clew ensuring return"` | Doctrine terminology |
| `/Users/tomtenuta/Code/roster/ariadne/internal/inscription/generator.go` | 406 | `"The thread ensuring return"` in table | `"The clew ensuring return"` | Doctrine terminology |
| `/Users/tomtenuta/Code/roster/knossos/templates/sections/ariadne-cli.md.tpl` | 15 | `ari hook thread` | `ari hook clew` | CLI command alignment |
| `/Users/tomtenuta/Code/roster/ariadne/README.md` | 90 | `ari hook thread` | `ari hook clew` | Documentation |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/hook/hook_test.go` | 308 | `"thread"` in subcommands list | `"clew"` | Test verification |
| `/Users/tomtenuta/Code/roster/ariadne/internal/sails/gate.go` | 39-152 | `thread contract` references | `clew contract` | Terminology |
| `/Users/tomtenuta/Code/roster/ariadne/internal/sails/contract.go` | 2, 15, 34, 43 | `thread contract` | `clew contract` | Terminology |
| `/Users/tomtenuta/Code/roster/ariadne/internal/sails/contract_test.go` | 324 | `thread contract violations` | `clew contract violations` | Test message |
| `/Users/tomtenuta/Code/roster/ariadne/internal/hook/clewcontract/record.go` | 10, 33 | `ari hook thread` | `ari hook clew` | Comment |
| `/Users/tomtenuta/Code/roster/tests/integration/test-thread-contract-validation.sh` | filename + line 2 | `thread contract` | `clew contract` | Test file |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap.go` | 242-245 | `THREAD_RECORD.ndjson`, `threadRecordPath` | `CLEW_RECORD.ndjson`, `clewRecordPath` | File name + variable |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap_test.go` | 844, 848 | `THREAD_RECORD.ndjson` | `CLEW_RECORD.ndjson` | Test fixture |
| `/Users/tomtenuta/Code/roster/user-agents/atropos.md` | 5, 359, 439 | `thread` references | `clew` | Agent documentation |
| `/Users/tomtenuta/Code/roster/docs/implementation/clew-contract-validation.md` | 228 | `ari hook thread` | `ari hook clew` | Documentation |
| `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0011-hook-deprecation-timeline.md` | 40, 54 | `ari hook thread` | `ari hook clew` | ADR |

### Category: KEEP (Semantic Usage)

These references use "thread" in its semantic meaning (execution context) and should NOT migrate:

| File | Line | Usage | Rationale |
|------|------|-------|-----------|
| `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/main-thread-guide.md` | all | "Main Thread" | Refers to Claude Code conversation context |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestration/main-thread-guide.md` | all | "Main Thread" | Same as above |
| `/Users/tomtenuta/Code/roster/user-hooks/validation/delegation-check.sh` | 73, 77, 86 | "main thread" | Execution context |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/orchestration-audit.sh` | 115 | "main-thread" | Audit category |
| `/Users/tomtenuta/Code/roster/user-hooks/context-injection/coach-mode.sh` | 47 | Reference to main-thread-guide.md | Path reference |
| `/Users/tomtenuta/Code/roster/ariadne/internal/inscription/generator.go` | 418 | "main thread coordinates" | Execution context |
| `/Users/tomtenuta/Code/roster/ariadne/internal/hook/clewcontract/writer.go` | 21 | "thread-safe" | Concurrency term |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/hook/clew.go` | 60 | "ball of thread" | Mythological explanation |
| `/Users/tomtenuta/Code/roster/docs/philosophy/knossos-doctrine.md` | 55 | "main Claude Code thread is Theseus" | Defining Theseus |

### Category: EVALUATE (Context-Dependent)

These require judgment based on surrounding context:

| File | Line | Usage | Decision | Rationale |
|------|------|-------|----------|-----------|
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap_test.go` | 395 | "threadcontract" | KEEP (comment reference) | References deprecated package name |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap_test.go` | 495 | "threadcontract" | KEEP (comment reference) | References deprecated package name |

---

## Migration Mapping

### CLI Command Rename

```
BEFORE: ari hook thread
AFTER:  ari hook clew
```

The command already exists as `clew` internally (registered in `hook.go:89` as `newClewCmd`). The external documentation incorrectly refers to it as `thread`.

**Verification**: The actual CLI command is already `ari hook clew`. Only documentation needs updating.

### File Rename

```
BEFORE: THREAD_RECORD.ndjson
AFTER:  CLEW_RECORD.ndjson
```

### Test File Rename

```
BEFORE: test-thread-contract-validation.sh
AFTER:  test-clew-contract-validation.sh
```

### Terminology Updates

| Pattern | Replacement |
|---------|-------------|
| `thread contract` | `clew contract` |
| `thread contract violations` | `clew contract violations` |
| `threadRecordPath` | `clewRecordPath` |
| `the thread ensuring return` | `the clew ensuring return` |

---

## Backward Compatibility Assessment

**Classification**: COMPATIBLE (with file migration consideration)

### Why COMPATIBLE

1. **CLI Command**: Already registered as `clew`, not `thread`. No external API change needed.
2. **No External Consumers**: The `THREAD_RECORD.ndjson` file is internal session state
3. **Documentation Only**: Most changes are in comments and documentation
4. **No Schema Changes**: No JSON/YAML schemas use "thread" for clew semantics

### Migration Consideration

**File Migration for `THREAD_RECORD.ndjson`**:

Existing sessions may have `THREAD_RECORD.ndjson` files. Options:

| Option | Approach | Risk |
|--------|----------|------|
| **A. Hard Cutover** | Rename file in code, ignore old sessions | OLD sessions lose budget data |
| **B. Fallback Read** | Read `CLEW_RECORD.ndjson` first, fall back to `THREAD_RECORD.ndjson` | Zero risk, minor code complexity |
| **C. Migration Script** | Rename files in existing sessions | One-time migration |

**Recommendation**: **Option B (Fallback Read)**

Rationale: Sessions are ephemeral. Old sessions will archive with `THREAD_RECORD.ndjson`, new sessions create `CLEW_RECORD.ndjson`. The `collectCognitiveBudget` function adds fallback logic with minimal complexity.

---

## Schema Definitions

No new schemas required. The clew contract is event-based (JSONL append-only log) and already uses doctrinally-correct event types in `clewcontract/event.go`.

---

## Implementation Sequence

### Phase 1: Go Code Updates (Internal)

| Order | File | Changes |
|-------|------|---------|
| 1.1 | `ariadne/internal/sails/gate.go` | Rename comments: "thread contract" -> "clew contract" |
| 1.2 | `ariadne/internal/sails/contract.go` | Rename comments and error messages |
| 1.3 | `ariadne/internal/sails/contract_test.go` | Update test assertions |
| 1.4 | `ariadne/internal/cmd/session/wrap.go` | Rename `THREAD_RECORD.ndjson` to `CLEW_RECORD.ndjson`, add fallback |
| 1.5 | `ariadne/internal/cmd/session/wrap_test.go` | Update test fixtures |
| 1.6 | `ariadne/internal/hook/clewcontract/record.go` | Update comments |
| 1.7 | `ariadne/internal/inscription/generator.go` | Update terminology table |

### Phase 2: Test Updates

| Order | File | Changes |
|-------|------|---------|
| 2.1 | `tests/integration/test-thread-contract-validation.sh` | Rename file + update content |

### Phase 3: Documentation Updates

| Order | File | Changes |
|-------|------|---------|
| 3.1 | `ariadne/README.md` | Update CLI reference |
| 3.2 | `knossos/templates/sections/ariadne-cli.md.tpl` | Update template |
| 3.3 | `user-agents/atropos.md` | Update terminology |
| 3.4 | `docs/implementation/clew-contract-validation.md` | Update references |
| 3.5 | `docs/decisions/ADR-0011-hook-deprecation-timeline.md` | Update references |
| 3.6 | `ariadne/internal/cmd/hook/hook_test.go` | Update subcommand list (if needed) |

### Phase 4: Inscription Sync

After template updates, run `ari inscription sync` to propagate changes to CLAUDE.md files.

---

## Integration Test Matrix

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| baseline | Create session, run clew hook | `CLEW_RECORD.ndjson` created in session dir |
| baseline | Run `ari sails check` | No "thread contract" in output messages |
| legacy | Session with existing `THREAD_RECORD.ndjson` | Fallback read succeeds, budget data accessible |
| complex | Full session lifecycle with wrap | `WHITE_SAILS.yaml` references clew, not thread |

---

## Quality Gate Criteria

- [ ] `grep -r "thread contract" ariadne/` returns only comments explaining the rename
- [ ] `grep -r "THREAD_RECORD" ariadne/` returns only fallback logic
- [ ] `ari hook clew` succeeds (already does - verify docs match)
- [ ] `go test ./...` passes in ariadne directory
- [ ] `ari sails check` output uses "clew contract" terminology
- [ ] `ari inscription sync` produces CLAUDE.md with "clew" not "thread" for Ariadne
- [ ] Integration test `test-clew-contract-validation.sh` passes

---

## Files Changed Summary

### Go Files (7 files)

1. `/Users/tomtenuta/Code/roster/ariadne/internal/sails/gate.go`
2. `/Users/tomtenuta/Code/roster/ariadne/internal/sails/contract.go`
3. `/Users/tomtenuta/Code/roster/ariadne/internal/sails/contract_test.go`
4. `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap.go`
5. `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap_test.go`
6. `/Users/tomtenuta/Code/roster/ariadne/internal/hook/clewcontract/record.go`
7. `/Users/tomtenuta/Code/roster/ariadne/internal/inscription/generator.go`

### Shell Files (1 file)

1. `/Users/tomtenuta/Code/roster/tests/integration/test-thread-contract-validation.sh` (RENAME + update)

### Documentation Files (5 files)

1. `/Users/tomtenuta/Code/roster/ariadne/README.md`
2. `/Users/tomtenuta/Code/roster/knossos/templates/sections/ariadne-cli.md.tpl`
3. `/Users/tomtenuta/Code/roster/user-agents/atropos.md`
4. `/Users/tomtenuta/Code/roster/docs/implementation/clew-contract-validation.md`
5. `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0011-hook-deprecation-timeline.md`

### Files NOT Changed (Semantic "thread" usage)

- `main-thread-guide.md` files (both locations)
- `delegation-check.sh` ("main thread" is semantic)
- `orchestration-audit.sh` ("main-thread" is audit category)
- `coach-mode.sh` (path reference)
- `clewcontract/writer.go` ("thread-safe" is concurrency term)
- `clew.go` ("ball of thread" is mythological explanation)

---

## Fallback Logic Specification

For `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap.go`:

```go
// collectCognitiveBudget attempts to collect cognitive budget metadata from the session.
// Returns nil if neither CLEW_RECORD.ndjson nor THREAD_RECORD.ndjson exists.
func collectCognitiveBudget(sessionDir string) map[string]interface{} {
    // Primary: New clew-based naming
    clewRecordPath := sessionDir + "/CLEW_RECORD.ndjson"

    // Fallback: Legacy thread-based naming (for pre-migration sessions)
    legacyRecordPath := sessionDir + "/THREAD_RECORD.ndjson"

    recordPath := clewRecordPath
    if _, err := os.Stat(recordPath); os.IsNotExist(err) {
        recordPath = legacyRecordPath
        if _, err := os.Stat(recordPath); os.IsNotExist(err) {
            return nil
        }
    }

    // ... rest of function uses recordPath
}
```

---

## Decision Rationale

### Why "Main Thread" is NOT Renamed

The Knossos Doctrine explicitly maps mythological terms:

> "In Knossos, the main Claude Code thread is Theseus: the agentic intelligence that makes decisions"

"Thread" here refers to the Claude conversation context (an execution thread), not the clew (navigation trail). Renaming would create confusion:

- "Main clew" would suggest the navigation device, not the navigator
- The industry term "main thread" for primary execution context is well-established
- Documentation already distinguishes: Theseus (navigator) uses the clew (trail)

### Why CLEW_RECORD.ndjson (not events.jsonl)

The session already has `events.jsonl` for the clew event log. `CLEW_RECORD.ndjson` (formerly `THREAD_RECORD.ndjson`) appears to be a separate cognitive budget tracking file. Renaming maintains the distinction while aligning with doctrine.

---

## Artifact Attestation

| Source File | Operation |
|-------------|-----------|
| `/Users/tomtenuta/Code/roster/docs/philosophy/knossos-doctrine.md` | Read |
| `/Users/tomtenuta/Code/roster/ariadne/internal/sails/gate.go` | Grep |
| `/Users/tomtenuta/Code/roster/ariadne/internal/sails/contract.go` | Grep |
| `/Users/tomtenuta/Code/roster/ariadne/internal/sails/contract_test.go` | Grep |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap.go` | Read |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap_test.go` | Grep |
| `/Users/tomtenuta/Code/roster/ariadne/internal/hook/clewcontract/record.go` | Grep |
| `/Users/tomtenuta/Code/roster/ariadne/internal/hook/clewcontract/writer.go` | Grep |
| `/Users/tomtenuta/Code/roster/ariadne/internal/inscription/generator.go` | Read |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/hook/hook.go` | Read |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/hook/clew.go` | Read |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/hook/hook_test.go` | Grep |
| `/Users/tomtenuta/Code/roster/knossos/templates/sections/ariadne-cli.md.tpl` | Read |
| `/Users/tomtenuta/Code/roster/user-hooks/ari/clew.sh` | Read |
| `/Users/tomtenuta/Code/roster/tests/hooks/test-ari-binary-resilience.sh` | Read |
| `/Users/tomtenuta/Code/roster/user-agents/atropos.md` | Read |
| `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/main-thread-guide.md` | Read |
| All files with "thread" pattern | Grep |
| All files with "clew" pattern | Grep |
| All files with "main-thread" pattern | Grep |

---

## Handoff to Integration Engineer

This Context Design is complete and ready for implementation. The Integration Engineer should:

1. Implement Phase 1 (Go code changes) with fallback logic
2. Run `go test ./...` after each file
3. Implement Phase 2 (test file rename)
4. Implement Phase 3 (documentation updates)
5. Run `ari inscription sync` for CLAUDE.md propagation
6. Execute integration test matrix
7. Verify quality gate criteria

No unresolved design decisions remain. All categorizations are final.

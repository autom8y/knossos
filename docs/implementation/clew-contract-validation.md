# Clew Contract Validation in Sails Check

**Implementation Date**: 2026-01-06
**Wave**: Wave 2, Task 4 (T1-004)
**Status**: Complete

## Overview

This implementation adds clew contract validation to the `ari sails check` command. Clew contract violations are detected by analyzing the `events.jsonl` file and result in degrading WHITE sails to GRAY.

## Motivation

Per Knossos Doctrine, "Theseus has amnesia; the Clew remembers." The clew contract (events.jsonl) provides the factual route through decisions. Validating this contract ensures that:

1. **Handoff integrity**: Agent handoffs are properly prepared before execution
2. **Task lifecycle**: Tasks have proper start/end pairs
3. **Session quality**: The clew contract is internally consistent

## Implementation

### Files Created

1. **`ariadne/internal/sails/contract.go`**
   - Core validation logic
   - Parses `events.jsonl` and checks event sequences
   - Returns list of `ContractViolation` structs

2. **`ariadne/internal/sails/contract_test.go`**
   - Comprehensive test coverage
   - Tests valid sequences, handoff violations, task violations
   - Tests color degradation behavior

3. **`tests/integration/test-clew-contract-validation.sh`**
   - Integration test script
   - End-to-end validation of sails check behavior
   - Tests multiple violation scenarios

### Files Modified

1. **`ariadne/internal/sails/gate.go`**
   - Added `ContractViolations` field to `GateResult`
   - Integrated `ValidateClewContract()` call in `CheckGate()`
   - Apply color degradation: WHITE → GRAY if violations present
   - BLACK and GRAY colors remain unchanged

2. **`ariadne/internal/cmd/sails/check.go`**
   - Added `ContractViolations` field to `gateOutput`
   - Added output formatting for contract violations
   - Displays violations in text output with severity labels

## Validation Rules

### Handoff Sequences

| Rule | Description | Violation Type |
|------|-------------|----------------|
| **Preparation required** | `handoff_executed` must have a preceding `handoff_prepared` for the same agent pair | `handoff_unprepared` |
| **Metadata presence** | Both events must have `from_agent` and `to_agent` metadata | `handoff_missing_metadata` |
| **Order enforcement** | Preparation must occur before execution in timeline | `handoff_out_of_order` |

### Task Lifecycle

| Rule | Description | Violation Type |
|------|-------------|----------------|
| **Start required** | `task_end` must have a preceding `task_start` for the same `task_id` | `task_orphaned_end` |
| **ID presence** | Both events must have `task_id` metadata | `task_missing_id` |
| **Duplicate detection** | Multiple `task_start` for same `task_id` generates warning | `task_duplicate_start` |
| **Order enforcement** | Start must occur before end in timeline | `task_out_of_order` |

## Color Degradation Logic

```go
// In CheckGate():
if len(violations) > 0 {
    // Clew contract violations degrade to GRAY minimum
    if color == ColorWhite {
        color = ColorGray
    }
}
```

- **WHITE → GRAY**: Violations prevent shipping without QA review
- **GRAY → GRAY**: No change (already needs QA)
- **BLACK → BLACK**: No change (known failure state)

## Usage

### Command Line

```bash
# Check current session
ari sails check

# Check specific session
ari sails check .sos/sessions/session-20260106-120000-abc12345

# Quiet mode (exit code only)
ari sails check --quiet
```

### Output Example

```
FAIL: Quality gate failed

Color:        GRAY
Computed:     WHITE (before modifiers)
Session:      session-20260106-120000-abc12345
File:         .sos/sessions/.../WHITE_SAILS.yaml

Reasons:
  - sails color is WHITE: high confidence, ship without QA
  - clew contract violations present: downgraded to GRAY
  - clew contract has violations (see contract_violations)

Clew Contract Violations:
  [ERROR] task_orphaned_end: task_end for task_id task-001 at event 5 has no preceding task_start
  [ERROR] handoff_unprepared: handoff_executed from agent-a to agent-b at event 12 has no preceding handoff_prepared
```

### JSON Output

```bash
ari sails check --format json
```

```json
{
  "pass": false,
  "color": "GRAY",
  "session_id": "session-20260106-120000-abc12345",
  "computed_base": "WHITE",
  "reasons": [
    "sails color is WHITE: high confidence, ship without QA",
    "clew contract violations present: downgraded to GRAY",
    "clew contract has violations (see contract_violations)"
  ],
  "contract_violations": [
    {
      "type": "task_orphaned_end",
      "description": "task_end for task_id task-001 at event 5 has no preceding task_start",
      "severity": "error",
      "related_events": [5]
    }
  ]
}
```

## Testing

### Unit Tests

Run the sails package tests:

```bash
cd ariadne
go test ./internal/sails/... -v -run TestValidateClewContract
go test ./internal/sails/... -v -run TestCheckGate_Contract
```

### Integration Tests

Run the integration test script:

```bash
./tests/integration/test-clew-contract-validation.sh
```

Test coverage:
- Valid event sequences (handoffs, tasks)
- Missing handoff preparation
- Orphaned task ends
- Duplicate task starts
- Missing metadata
- Color degradation (WHITE → GRAY)
- Preservation of GRAY and BLACK colors
- Missing events.jsonl (no violations)

## Architecture Decisions

### Why Degrade to GRAY Instead of BLACK?

Clew contract violations indicate **uncertainty** rather than **known failure**:
- The work may be complete and correct
- The clew contract just has inconsistencies
- QA review can determine if the violations matter

This aligns with:
- **GRAY**: Unknown confidence, needs QA review
- **BLACK**: Known failure, do not ship

### Why Not Fail on Missing events.jsonl?

Missing `events.jsonl` is not a contract violation because:
1. Not all sessions may have event recording enabled
2. The contract only applies when events exist
3. WHITE_SAILS.yaml is the primary artifact

This is a "best-effort" validation that adds value when events exist.

### Why Parse Events in Gate Check?

Parsing events during gate check (rather than during generation) ensures:
1. Clew contract is validated at ship-time
2. Events can be recorded incrementally during session
3. No need for real-time validation overhead
4. Violations are caught before deployment

## Future Enhancements

Potential improvements for future waves:

1. **Session lifecycle validation**: Ensure `session_start` exists
2. **Artifact chain validation**: Verify artifact creation events reference valid parents
3. **Error event analysis**: Check for unresolved error events
4. **Timeline consistency**: Validate timestamp ordering
5. **Violation severity tuning**: Distinguish warnings from errors in color impact

## References

- **Knossos Doctrine v2**: Section VI (Clew Contract)
- **Wave 2 Plan**: Task T1-004 (Clew Contract Validation)
- **White Sails TDD**: Section 7 (Quality Gate Check)
- **Hook Implementation**: `ariadne/internal/hook/clewcontract/`

## Related Commands

- `ari hook clew`: Record clew events to events.jsonl
- `ari sails generate`: Generate WHITE_SAILS.yaml
- `ari sails check`: Validate quality gate (includes clew contract)

---

**Implementation complete**: Clew contract validation is now integrated into the sails check quality gate.

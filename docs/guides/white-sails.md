# White Sails: Confidence Signaling

> Prevent Aegeus failures with honest session confidence signals

## Overview

White Sails is a confidence signaling system that provides honest assessment of session quality at wrap time. Named after the Greek myth where Theseus forgot to raise white sails, causing his father Aegeus to believe his son dead, White Sails prevents "Aegeus failures" -- shipping code with false confidence.

Every session produces a typed contract (`WHITE_SAILS.yaml`) declaring the computed confidence level with explicit proof chains. The core principle: **any open question = gray ceiling**. You cannot claim high confidence when uncertainty exists.

### Why White Sails?

Without confidence signaling:
- Tests pass, but edge cases were not considered
- Build succeeds, but integration points were missed
- Code looks good, but there are unresolved questions

White Sails makes this explicit. Either you have proof (WHITE), you do not (GRAY), or you are broken (BLACK).

## The Three Colors

| Color | Meaning | Deploy Guidance |
|-------|---------|-----------------|
| **WHITE** | High confidence - all proofs present, all tests pass, no open questions | Ship without QA |
| **GRAY** | Unknown confidence - missing proofs, open questions, or declared uncertainty | Consider QA review |
| **BLACK** | Known failure - tests failing, build broken, or explicit blocker | Do NOT ship |

There is no yellow or intermediate state. This simplicity prevents gaming.

### WHITE: High Confidence

WHITE sails mean:
- All required proofs are present and passing
- Tests pass with appropriate coverage
- Build succeeds
- Linting is clean
- No open questions remain
- No unresolved TODOs

With WHITE sails, you can confidently ship without additional QA review.

### GRAY: Unknown Confidence

GRAY sails indicate uncertainty. This happens when:
- Required proofs are missing or unknown
- Open questions exist (any "?" in your session context)
- Session type is a spike or hotfix (automatic gray ceiling)
- Complexity requires proofs you did not collect
- You applied a human downgrade modifier

GRAY is not bad -- it is honest. Most sessions wrap with GRAY sails initially, then upgrade to WHITE through QA review.

### BLACK: Known Failure

BLACK sails signal explicit failure:
- One or more tests failed
- Build is broken
- A blocking issue was identified
- You applied a `DOWNGRADE_TO_BLACK` modifier

Do not ship code with BLACK sails. Fix the failures first.

## How It Works

At session wrap (`/wrap`), White Sails:

1. **Collects proof artifacts** from your session directory
   - Test output logs
   - Build output logs
   - Lint output logs

2. **Checks for open questions** in SESSION_CONTEXT.md

3. **Evaluates session type** (standard, spike, hotfix)

4. **Computes confidence color** using the algorithm below

5. **Generates WHITE_SAILS.yaml** in your session directory

6. **Reports the result** to your wrap summary

### Computation Algorithm

```
1. Any proof with status FAIL?         -> BLACK
2. Any open questions present?         -> GRAY ceiling
3. Session type is spike or hotfix?    -> GRAY ceiling
4. All required proofs present & pass? -> WHITE
5. Apply any declared modifiers        -> May downgrade
6. QA upgrade applied?                 -> May upgrade GRAY to WHITE
```

## Proof Requirements

White Sails uses a sliding scale of proof requirements based on session complexity. Higher complexity requires stricter proof standards.

### Required Proofs by Complexity

| Complexity | tests | build | lint | adversarial | integration |
|------------|:-----:|:-----:|:----:|:-----------:|:-----------:|
| PATCH      | Required | Required | Required | - | - |
| SCRIPT     | Required | Required | Required | - | - |
| MODULE     | Required | Required | Required | - | - |
| SERVICE/SYSTEM | Required | Required | Required | Recommended | Recommended |
| INITIATIVE | Required | Required | Required | Required | Required |
| MIGRATION  | Required | Required | Required | Required | Required |
| PLATFORM   | Required | Required | Required | Required | Required |

**Legend**:
- **Required**: Must be PASS or SKIP for WHITE sails
- **Recommended**: Encouraged but not blocking
- **-**: Not tracked at this complexity level

### Proof Statuses

| Status | Meaning |
|--------|---------|
| `PASS` | Proof collected and verification succeeded |
| `FAIL` | Proof collected and verification failed (triggers BLACK) |
| `SKIP` | Proof intentionally skipped (with justification) |
| `UNKNOWN` | Proof not collected or status undetermined (triggers GRAY) |

## Using /wrap with Sails

When you invoke `/wrap`, White Sails integrates automatically:

```
/wrap
```

The wrap process:
1. Runs standard wrap validation
2. Generates WHITE_SAILS.yaml
3. Reports the confidence color
4. Proceeds or blocks based on color

### Wrap Flags

| Flag | Description |
|------|-------------|
| `--accept-gray` | Accept GRAY sails without QA review |
| `--skip-sails` | Skip sails generation (equivalent to spike session) |

### Gray Warning Example

If your session wraps with GRAY:

```
[WARNING] Session wrapped with GRAY sails.

Confidence: GRAY (2 open questions found)
- How should rate limiting behave under cluster failover?
- Need to validate with Production DBA on index strategy

Recommendation: Run /qa to earn WHITE sails, or use --accept-gray to proceed.
```

## The QA Upgrade Path

If your session wraps with GRAY sails, you have two options:

1. **Accept GRAY** with `--accept-gray` flag and ship with warning
2. **Run QA** to earn WHITE sails through independent validation

### How QA Upgrades Work

A GRAY session can be upgraded to WHITE only through an independent QA session:

1. Original session wrapped with GRAY sails
2. Create a new QA session: `/start "QA: {original initiative}" --complexity=PATCH`
3. QA session references the original session ID
4. QA adversary agent runs adversarial validation
5. If issues found: document in constraint resolution log
6. If issues resolved: add adversarial tests
7. QA session wraps with upgrade declaration

### QA Upgrade Requirements

For QA to upgrade GRAY to WHITE, it must provide:

| Requirement | Description |
|-------------|-------------|
| `constraint_resolution_log` | Path to document explaining how constraints were resolved |
| `adversarial_tests_added` | At least one new test added by QA |

Without both, the upgrade is rejected.

### After QA Upgrade

The original session's WHITE_SAILS.yaml is updated:

```yaml
color: "WHITE"           # Changed from GRAY
computed_base: "GRAY"    # Unchanged - honest about original computation
qa_upgrade:
  upgraded_at: "2026-01-06T12:00:00Z"
  qa_session_id: "session-20260106-100000-qa123456"
  constraint_resolution_log: "docs/testing/TP-qa-original-session.md"
  adversarial_tests_added:
    - "tests/integration/rate_limit_failover_test.go"
    - "tests/integration/index_edge_cases_test.go"
```

## Modifiers

Human-declared modifiers can adjust the computed color. Modifiers can only **downgrade** -- you cannot self-upgrade without QA.

### Modifier Types

| Modifier | Effect | Use Case |
|----------|--------|----------|
| `DOWNGRADE_TO_GRAY` | WHITE -> GRAY | Want senior review despite passing tests |
| `DOWNGRADE_TO_BLACK` | Any -> BLACK | Explicit blocker identified |
| `HUMAN_OVERRIDE_GRAY` | Forces GRAY | Known uncertainty not captured by proofs |

### Applying Modifiers

Modifiers are applied during wrap with justification:

```yaml
modifiers:
  - type: "DOWNGRADE_TO_GRAY"
    justification: "Changed retry logic in payment flow; want senior review before shipping despite passing tests"
    applied_by: "human"
    timestamp: "2026-01-05T16:55:00Z"
```

Justification must be at least 10 characters -- no drive-by downgrades.

## Special Session Types

### Spikes

Spike sessions (research/exploration) always have a GRAY ceiling:
- Spikes produce learnings, not production code
- `type: spike` flag is set automatically
- Cannot achieve WHITE sails

### Hotfix

Hotfix sessions (urgent production fixes) always have a GRAY ceiling:
- Expedited path acknowledges risk explicitly
- `type: hotfix` flag is set automatically
- Recommend follow-up session for comprehensive validation

## WHITE_SAILS.yaml Reference

The sails artifact is generated at:

```
.claude/sessions/{session-id}/WHITE_SAILS.yaml
```

### Example: WHITE Sails

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
    evidence_path: ".claude/sessions/session-20260105-143000-abc12345/test-output.log"
    summary: "47 tests passed, 0 failed, 0 skipped"
    exit_code: 0
    timestamp: "2026-01-05T15:28:00Z"
  build:
    status: "PASS"
    evidence_path: ".claude/sessions/session-20260105-143000-abc12345/build-output.log"
    summary: "go build succeeded"
    exit_code: 0
    timestamp: "2026-01-05T15:29:00Z"
  lint:
    status: "PASS"
    evidence_path: ".claude/sessions/session-20260105-143000-abc12345/lint-output.log"
    summary: "golangci-lint clean"
    exit_code: 0
    timestamp: "2026-01-05T15:29:30Z"

open_questions: []
modifiers: []
```

### Example: GRAY Sails with Open Questions

```yaml
schema_version: "1.0"
session_id: "session-20260105-160000-def67890"
generated_at: "2026-01-05T16:00:00Z"
color: "GRAY"
computed_base: "GRAY"
complexity: "SERVICE"
type: "standard"

proofs:
  tests:
    status: "PASS"
    summary: "89 tests passed"
    exit_code: 0
  build:
    status: "PASS"
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

### Example: Human Downgrade

```yaml
schema_version: "1.0"
session_id: "session-20260105-170000-ghi11111"
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

## Anti-Gaming

White Sails includes both technical and cultural defenses against gaming:

### Technical Defenses

- **Cannot self-upgrade**: Modifiers can only downgrade, never upgrade
- **QA upgrade requires proof**: Must provide constraint_resolution_log + adversarial_tests_added
- **Open questions propagate**: Any "?" in context triggers gray ceiling
- **Proof verification**: Proofs must have evidence_path or UNKNOWN status

### Cultural Defenses

- **Trust debt**: Teams that ship gray without QA accumulate reputation cost
- **QA adversarial**: QA agent explicitly tries to break assumptions
- **Visibility**: WHITE_SAILS.yaml is committed and auditable

## Troubleshooting

### "My session always wraps GRAY"

Check for:
1. Open questions in SESSION_CONTEXT.md (search for "?")
2. Missing proofs (did you run tests, build, lint?)
3. Session type is spike or hotfix
4. Complexity requires proofs you did not collect

### "Proofs are UNKNOWN"

UNKNOWN means the proof was not collected. Ensure:
1. Test command was run during session
2. Build command was run during session
3. Lint command was run during session
4. Output was captured to session directory

### "QA upgrade was rejected"

QA upgrades require both:
1. `constraint_resolution_log` - a path to your QA findings
2. `adversarial_tests_added` - at least one new test file

### "I need to ship despite GRAY"

Use `--accept-gray` to acknowledge the risk:

```
/wrap --accept-gray
```

This ships with GRAY but records your acknowledgment.

## Schema Reference

The full JSON Schema for WHITE_SAILS.yaml is available at:

```
ariadne/internal/validation/schemas/white-sails.schema.json
```

Key constraints:
- `schema_version` follows semver pattern (e.g., "1.0")
- `session_id` matches pattern `session-YYYYMMDD-HHMMSS-{8-char-hex}`
- `color` and `computed_base` must be "WHITE", "GRAY", or "BLACK"
- Modifier `justification` must be at least 10 characters

## Related Documentation

- [User Preferences](./user-preferences.md) - Configure Claude Code behavior
- [TDD: Knossos v2](../design/TDD-knossos-v2.md) - Technical design document
- [Session Lifecycle](../session-fsm/README.md) - Session state machine

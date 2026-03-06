---
domain: feat/white-sails-signaling
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/sails/**/*.go"
  - "./internal/cmd/sails/**/*.go"
  - "./internal/validation/sails.go"
  - "./internal/validation/schemas/white-sails.schema.json"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.87
format_version: "1.0"
---

# White Sails Quality Gate Signaling

## Purpose and Design Rationale

Honest confidence signaling to prevent "Aegeus failures" — false confidence leading to production issues. Three-state signal (WHITE/GRAY/BLACK), proof-chain requirement, complexity-scaled thresholds, downgrade-only modifiers.

**Design**: System reads log files (`test-output.log`, `build-output.log`, `lint-output.log`) and computes color deterministically. Removes agent honesty as a dependency.

## Conceptual Model

### Three-Color Signal

- **WHITE**: All proofs passing, no open questions, no blockers → ship without QA
- **GRAY**: Missing proof, open question, spike/hotfix type, or clew contract violations → needs QA
- **BLACK**: FAIL proof, explicit blocker, or DOWNGRADE_TO_BLACK modifier → do not ship

### 7-Step ComputeColor Algorithm

1a. Blockers → BLACK. 1b. FAIL proof → BLACK. 2. Open questions → GRAY ceiling. 3. Session type spike/hotfix → GRAY. 4. Missing required proofs → GRAY. 5. All passing → WHITE base. 6. Apply modifiers (downgrade only). 7. QA upgrade (requires `constraint_resolution_log` + adversarial tests).

### Complexity Tier Matrix

PATCH/SCRIPT/MODULE: tests+build+lint required. SERVICE/SYSTEM: +recommended adversarial/integration. INITIATIVE/MIGRATION/PLATFORM: all required.

## Implementation Map

Domain: `/Users/tomtenuta/Code/knossos/internal/sails/` — `color.go`, `thresholds.go`, `proofs.go`, `generator.go`, `gate.go`, `contract.go`. CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/sails/sails.go`, `check.go`.

### Key Types

`Color`, `ColorInput`, `ColorResult`, `ProofSet`, `Generator`, `GateResult`, `ContractViolation`, `WhiteSailsYAML`.

## Boundaries and Failure Modes

- Missing proof logs → GRAY (graceful, not error)
- Missing SESSION_CONTEXT.md → defaults applied (potential false positive for MODULE)
- Schema validation optional (only when `Validator` injected)
- Lint output word-count fallback can produce false FAIL
- `SERVICE` complexity valid in Go but fails JSON schema validation (known mismatch)
- Clew contract validates only handoff sequences and task lifecycles (not session lifecycle)

## Knowledge Gaps

1. `ari session wrap` integration point not confirmed from source.
2. Schema complexity enum mismatch (Go has 7 values, JSON schema has 5).
3. Proof log file ownership (who writes them) not captured.

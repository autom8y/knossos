---
domain: feat/white-sails-signaling
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/sails/**/*.go"
  - "./internal/cmd/sails/**/*.go"
  - "./internal/validation/schemas/white-sails.schema.json"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.92
format_version: "1.0"
---

# White Sails Quality Gate Signaling

## Purpose and Design Rationale

Session-end confidence signaling to prevent "Aegeus failures." Three colors only: WHITE (ship), GRAY (needs QA), BLACK (never ship). Proof-based (reads physical log files, not agent self-assessment). Complexity-tiered threshold matrix. QA upgrade path (GRAY->WHITE with constraints + adversarial tests). Modifiers downgrade-only. Gate checking is binary output for CI/CD.

## Conceptual Model

**Two-phase color:** computed_base + final color (after modifiers/QA). **Five proof types:** tests, build, lint (all levels), adversarial, integration (INITIATIVE+ required). **Proof statuses:** PASS, SKIP (both passing), FAIL (BLACK), UNKNOWN (GRAY if required). **Color algorithm:** blockers->BLACK, any FAIL->BLACK, open questions->GRAY ceiling, spike/hotfix->GRAY ceiling, missing required->GRAY, all pass->WHITE, apply modifiers (downgrade-only), apply QA upgrade. **Three modifier types:** DOWNGRADE_TO_GRAY, DOWNGRADE_TO_BLACK, HUMAN_OVERRIDE_GRAY.

## Implementation Map

`internal/sails/` (6 files): color.go (ComputeColor algorithm), thresholds.go (7 complexity levels, threshold matrix), proofs.go (CollectProofs, ParseProofLog with multi-framework regex), generator.go (Generator.Generate orchestration), gate.go (CheckGate, GateExitCode, clew contract integration), contract.go (ValidateClewContract, handoff/task lifecycle validation). CLI: `internal/cmd/sails/check.go`. Schema: white-sails.schema.json.

## Boundaries and Failure Modes

Read-only (does not run tests/build/lint). Generator invoked at wrap time by agents/CLI. ari sails check reads existing WHITE_SAILS.yaml (does not regenerate). Clew contract validation is best-effort (fail-open). Missing proof -> UNKNOWN -> GRAY if required. Unknown complexity -> strictest thresholds. Schema divergence: SCRIPT/PLATFORM in code but not schema enum. Lint analyzer false positives (word boundary matching). No generation history (each Generate overwrites).

## Knowledge Gaps

1. clewcontract event type definitions not fully read
2. No ari sails generate CLI command found
3. Integration with session wrap lifecycle not fully traced

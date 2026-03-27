---
domain: feat/mena-lint
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/lint/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.93
format_version: "1.0"
---

# Mena and Agent Lint

## Purpose and Design Rationale

Pre-sync validation gate for source artifacts. Driven by three SCARs: SCAR-017 (195+ @skill-name references silently ignored by CC), SCAR-019 (invalid agent colors dropped), SCAR-027 (session artifacts in shared mena). Operates on source (rites/*/agents/, rites/*/mena/, mena/), not materialized output. --check=preferential-language is the only non-zero exit path. Registration: NeedsProject=true.

## Conceptual Model

**Four severity levels:** CRIT (structural), HIGH (CC will silently ignore), MED (convention violation), LOW (style gap). **Agent rules:** frontmatter-missing/parse, name/description/type, maxTurns-deviation, agent-oversized, agent-invalid-color, color-duplicate, skill-at-syntax, naming-provenance, domain-jurisdiction. **Dromena rules:** frontmatter, context-fork-expected/unexpected/unclassified, fork-task-conflict (SCAR-018), workflow-not-model-invocable, name-collision, source-path-leaks. **Legomena rules:** triggers-missing, oversized, mena-name/description. **Cross-cutting:** session-artifact-in-shared-mena (SCAR-027), preferential-language (harness-agnosticism).

## Implementation Map

`internal/cmd/lint/` (4 files): lint.go (1141 lines -- all rules, expectedForkState allowlist, lintAgents/lintDromena/lintLegomena/lintMenaNamespace/lintSessionArtifactsInSharedMena), lint_preferential.go (209 lines -- Go source + mena content harness-agnosticism scans), lint_test.go (403 lines), lint_preferential_test.go (476 lines). Uses internal/mena.Walk (filesystem only, skips embedded) and internal/frontmatter.Parse.

## Boundaries and Failure Modes

Does not lint materialized .claude/ output or user-scope agents. expectedForkState staleness (compile-time allowlist). No --fail-on-critical flag (only --check=preferential-language forces non-zero exit). Preferential-language false positives (word-boundary regex on "claude"/"gemini"). mena.Walk skips embedded FS sources. session-artifact check exempts examples/ subdirectory. No test coverage for agent color/size/turns rules.

## Knowledge Gaps

1. Census referenced "models" file that doesn't exist
2. Exit code design (informational vs gate) not documented
3. No --check support for individual standard rules

---
domain: release/history
generated_at: "2026-03-10T14:02:00Z"
source_scope:
  - "./.know/release/"
generator: pipeline-monitor
source_hash: "c5d0238"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 0
---

## Release Log

### v0.3.0 — 2026-03-03

| Field | Value |
|-------|-------|
| Date | 2026-03-03 |
| Version | v0.2.0 -> v0.3.0 (minor) |
| Commits | 79 (78 existing + 1 prep) |
| Complexity | PATCH |
| Tag SHA | 4ca5cc0ae387c4f6129a60bf976a1e35458d8e50 |
| Commit SHA | 3e2457da89f35322948d2d55bc95dd0912f2d43c |
| release.yml | GREEN (138s, run 22629614142) |
| e2e-distribution.yml | DISPATCH_NOT_RECEIVED (skipped by user) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (commit 038d885c) |
| Verdict | PARTIAL |
| Duration | ~5 min (execution) + ~5 min (monitoring) |

**Notes:**
- 4 `.know/` files committed as prep (`chore(know): refresh knowledge files`)
- e2e-distribution.yml `release: published` trigger did not fire (same as v0.2.0)
- Dependabot flagged 4 vulnerabilities on push (1 critical, 1 high, 2 moderate) — informational only

### v0.3.1 — 2026-03-03

| Field | Value |
|-------|-------|
| Date | 2026-03-03 |
| Version | v0.3.0 -> v0.3.1 (patch) |
| Commits | 1 (fix(ci): use PAT for GoReleaser) |
| Complexity | PATCH |
| Tag SHA | f7a4befa737e228912867a9d2d51c013755b2700 |
| Commit SHA | d65bd4c |
| release.yml | GREEN (109s, run 22631036118) |
| e2e-distribution.yml | RED (52s, run 22631109399) — first-ever auto-trigger! |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (commit c4402d2) |
| Verdict | PARTIAL |
| Duration | ~1 min (execution) + ~4 min (monitoring) |

**Notes:**
- Primary goal achieved: e2e-distribution.yml auto-triggered for the first time (2s after release published)
- Dispatch chain fix validated: GITHUB_TOKEN -> HOMEBREW_TAP_TOKEN in release.yml
- E2E failure: Assertion 7 — `ari init` does not produce `.claude/settings.json` (pre-existing, unrelated to v0.3.1)
- Both macOS and Linux passed Assertions 1-6, failed identically at Assertion 7
- Follow-up needed: fix `ari init` or update `scripts/e2e-validate.sh` Assertion 7

## Historical Summary (v0.3.2 — v0.3.12)

| Version | Date | Commits | Verdict | Chain Time | Notes |
|---------|------|---------|---------|------------|-------|
| v0.3.2 | 2026-03-03 | 1 | PASS | ~161s | First full PASS verdict |
| v0.3.3 | 2026-03-03 | 2 | PASS | ~174s | Agent-guard write-guard fix |
| v0.3.4 | 2026-03-04 | 31 | PASS | ~162s | Largest batch (service boundary, --root flag) |
| v0.3.5 | 2026-03-04 | 1 | PASS | ~162s | First cache-only PATCH |
| v0.3.6 | 2026-03-04 | 1 | PASS | ~157s | Second cache-only PATCH |
| v0.3.7 | 2026-03-04 | 7 | PASS | ~184s | Mena relocation + materialize fixes |
| v0.3.9 | 2026-03-05 | 5 | PASS | ~192s | Attribution-guard precommit hook |
| v0.3.10 | 2026-03-05 | 17 | PASS | ~168s | .claude/ -> .knossos/ boundary migration |
| v0.3.11 | 2026-03-05 | 1 | PASS | ~191s | Theoros context budget + maxTurns |
| v0.3.12 | 2026-03-05 | 1 | PASS | ~169s | Standardized .gitignore management |

### v0.4.0 — 2026-03-05

| Field | Value |
|-------|-------|
| Date | 2026-03-05 |
| Version | v0.3.12 -> v0.4.0 (minor) |
| Commits | 12 (7 feat, 2 fix, 1 refactor, 2 merge) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | c80d3658c54b58225476cfef3a3a6961d761fbd7 |
| Commit SHA | 1bf2630 |
| release.yml | GREEN (101s, run 22731888201) |
| e2e-distribution.yml | GREEN (49s, run 22731961407) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.4.0, commit 37807e8, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~2 min (monitoring) |

**Notes:**
- Eleventh consecutive PASS verdict (v0.3.2 -> ... -> v0.3.12 -> v0.4.0)
- First MINOR release since v0.3.0; semver escalation flagged by cartographer (feat commits)
- Dispatch latency: ~22s (release published 18:57:22Z, E2E triggered ~18:57:50Z)
- macOS E2E: GREEN (22s), Linux E2E: GREEN (49s)
- Total chain: ~150s — fastest in recent history
- Headline features: ari agent embody command, Dionysus knowledge synthesis agent, /dion dromenon, land-to-know pipeline, shelf promotion, knowledge maturation pipeline
- 3 pre-existing informational workflow failures (non-blocking): ariadne-tests, verify-doctrine, verify-formal-specs

### v0.5.0 — 2026-03-06

| Field | Value |
|-------|-------|
| Date | 2026-03-06 |
| Version | v0.4.0 -> v0.5.0 (minor) |
| Commits | 3 (2 feat, 1 chore) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 7ab501b73b05297f9e43312b8b4adc533f81ec49 |
| Commit SHA | 94dc82f |
| release.yml | GREEN (111s, run 22755776615) |
| e2e-distribution.yml | GREEN (55s, run 22755831193) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.5.0, commit f7ba647) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twelfth consecutive PASS verdict (v0.3.2 → ... → v0.4.0 → v0.5.0)
- Dispatch latency: ~2s (consistent with historical avg)
- Total chain: ~166s — consistent with historical avg (~150-190s)
- Cached release knowledge skipped cartographer reconnaissance entirely
- Headline features: L2 Perception, L6 Position, L7 Surface, L8 Horizon perspective layers + full simulate mode

### v0.6.0 — 2026-03-06

| Field | Value |
|-------|-------|
| Date | 2026-03-06 |
| Version | v0.5.0 -> v0.6.0 (minor) |
| Commits | 34 (4 feat, 8 fix, 17 refactor, 2 chore, 3 other) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | ca7caa98060592aa89c77399b04e131c06a7e7b6 |
| Commit SHA | c6d04eb |
| release.yml | GREEN (113s, run 22783545282) |
| e2e-distribution.yml | GREEN (43s, run 22783603668) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.6.0, commit 6fd03caa) |
| Verdict | PASS |
| Duration | ~2 min (execution) + ~3 min (monitoring) |

**Notes:**
- Thirteenth consecutive PASS verdict (v0.3.2 → ... → v0.5.0 → v0.6.0)
- Attempt 2: attempt 1 failed due to goreleaser config referencing non-existent docs/guides/ariadne-cli.md; fixed and retagged
- Dispatch latency: ~5s
- Total chain: ~150s (below historical baseline)
- Cached release knowledge skipped cartographer and dependency-resolver entirely
- Headline features: session schema v2.3 (strand struct, frame linkage, complexity enum, claim command), /sos unified session interface, context-session alignment, dromena v2.3 hook fields
- Bulk refactor: 11 commits migrating fmt.Errorf → errors.New across internal packages

### v0.6.1 — 2026-03-07

| Field | Value |
|-------|-------|
| Date | 2026-03-07 |
| Version | v0.6.0 -> v0.6.1 (patch) |
| Commits | 1 (fix(cli): remediate help tooling and error UX) |
| Complexity | PATCH |
| Tag SHA | 6df2d74c94758d30b691376a48eefd0128507d80 |
| Commit SHA | 2b4a43c |
| release.yml | GREEN (113s, run 22798027635) |
| e2e-distribution.yml | GREEN (44s, run 22798053204) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.6.1, commit 476891a) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Fourteenth consecutive PASS verdict (v0.3.2 → ... → v0.6.0 → v0.6.1)
- Total chain: 147s (below historical baseline ~150-190s)
- Cached release knowledge accelerated cartographer reconnaissance
- Single fix commit: CLI help tooling and error UX remediation

### v0.7.0 — 2026-03-09

| Field | Value |
|-------|-------|
| Date | 2026-03-09 |
| Version | v0.6.1 -> v0.7.0 (minor) |
| Commits | 17 (7 feat, 3 fix, 1 refactor, 2 docs, 1 chore, 2 merge, 1 other) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 27e1cc7 |
| Commit SHA | 27e1cc7 |
| release.yml | GREEN (106s, run 22854404982) |
| e2e-distribution.yml | GREEN (53s, run 22854468043) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Fifteenth consecutive PASS verdict (v0.3.2 → ... → v0.6.1 → v0.7.0)
- Total chain: ~159s (within historical baseline)
- Headline features: ari ask NL query command, session-aware routing, Naxos session hygiene closed-loop, Forge Pythia consultation protocol, naming provenance lint rules, proactive intelligence hooks

### v0.7.1 — 2026-03-09

| Field | Value |
|-------|-------|
| Date | 2026-03-09 |
| Version | v0.7.0 -> v0.7.1 (patch) |
| Commits | 4 (1 fix, 1 refactor, 2 ci) |
| Complexity | PATCH |
| Tag SHA | 24e6a7463ac84ac664cc3d9a6555cbabdc3c3eff |
| Commit SHA | a81ea97 |
| release.yml | GREEN (106s, run 22856008272) |
| e2e-distribution.yml | GREEN (56s, run 22856075627) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (commit 02217a07, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~17s (execution) + ~3 min (monitoring) |

**Notes:**
- Sixteenth consecutive PASS verdict (v0.3.2 → ... → v0.7.0 → v0.7.1)
- Total chain: ~162s (within historical baseline ~150-190s)
- Cached release knowledge accelerated cartographer reconnaissance
- All fix/refactor/ci: 530 golangci-lint findings resolved, stale CI configs fixed, SCAR-018 enforcement
- verify-formal-specs.yml removed from repo (reduces known non-blocking failures from 3 to 2)

### v0.7.2 — 2026-03-09

| Field | Value |
|-------|-------|
| Date | 2026-03-09 |
| Version | v0.7.1 -> v0.7.2 (patch) |
| Commits | 2 (1 fix, 1 chore) |
| Complexity | PATCH |
| Tag SHA | 4b09f6798c054baaea3b2162fb478dbb7584ce22 |
| Commit SHA | 5a24a76 |
| release.yml | GREEN (110s, run 22867920447) |
| e2e-distribution.yml | GREEN (61s, run 22867987422) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (commit 302d0b71, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Seventeenth consecutive PASS verdict (v0.3.2 → ... → v0.7.1 → v0.7.2)
- Total chain: 161s (within historical baseline ~150-190s)
- Dispatch latency: 2s
- Cached release knowledge skipped cartographer and dependency-resolver entirely
- Fix: retarget MCP server sync to .mcp.json (SCAR-028)

### v0.8.0 — 2026-03-10

| Field | Value |
|-------|-------|
| Date | 2026-03-10 |
| Version | v0.7.2 -> v0.8.0 (minor) |
| Commits | 19 (12 feat, 4 fix, 1 refactor, 1 chore, 1 other) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 4092f3acde76b7ba493dd6dc28347c8c3a284d0b |
| Commit SHA | c5d0238 |
| release.yml | GREEN (110s, run 22906121647) |
| e2e-distribution.yml | GREEN (51s, run 22906197948) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.8.0, commit c754006, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Eighteenth consecutive PASS verdict (v0.3.2 → ... → v0.7.2 → v0.8.0)
- Total chain: 153s (within historical baseline ~150-190s)
- Dispatch latency: 2s
- Cached release knowledge skipped cartographer and dependency-resolver entirely
- Primary scope: procession — cross-rite coordinated workflow primitives
- Headline features: procession create/proceed/recede/list commands, cross-rite mena pipeline, procession search collector, alt_rite awareness, Moirai/Pythia/go procession integration, P2+P3 lifecycle test hardening

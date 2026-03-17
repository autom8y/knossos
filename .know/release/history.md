---
domain: release/history
generated_at: "2026-03-17T15:07:45Z"
source_scope:
  - "./.know/release/"
generator: pipeline-monitor
source_hash: "9340b49"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 0
---

## Release Log

## Historical Summary (v0.3.0 — v0.3.12)

| Version | Date | Commits | Verdict | Chain Time | Notes |
|---------|------|---------|---------|------------|-------|
| v0.3.0 | 2026-03-03 | 79 | PARTIAL | ~143s | First GoReleaser release; e2e dispatch not received |
| v0.3.1 | 2026-03-03 | 1 | PARTIAL | ~161s | First e2e auto-trigger; PAT fix for GoReleaser |
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

## Historical Summary (v0.4.0 — v0.6.1)

| Version | Date | Commits | Verdict | Chain Time | Notes |
|---------|------|---------|---------|------------|-------|
| v0.4.0 | 2026-03-05 | 12 | PASS | ~150s | ari agent embody, Dionysus, /dion, land-to-know pipeline |
| v0.5.0 | 2026-03-06 | 3 | PASS | ~166s | Perspective layers (L2, L6, L7, L8) + simulate mode |
| v0.6.0 | 2026-03-06 | 34 | PASS | ~150s | Session schema v2.3, /sos, fmt.Errorf → errors.New |
| v0.6.1 | 2026-03-07 | 1 | PASS | 147s | CLI help tooling and error UX remediation |

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
- Primary scope: procession — cross-rite coordinated workflow primitives
- Headline features: procession create/proceed/recede/list commands, cross-rite mena pipeline, procession search collector, alt_rite awareness, Moirai/Pythia/go procession integration

### v0.9.0 — 2026-03-11

| Field | Value |
|-------|-------|
| Date | 2026-03-11 |
| Version | v0.8.0 -> v0.9.0 (minor) |
| Commits | 25 |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | (v0.9.0 annotated tag) |
| Commit SHA | (v0.9.0 tag target) |
| release.yml | GREEN (107s, run 22946938445) |
| e2e-distribution.yml | GREEN (48s, run 22947002338) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (commit 8193825d, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Nineteenth consecutive PASS verdict (v0.3.2 → ... → v0.8.0 → v0.9.0)
- Total chain: ~155s (within historical baseline ~150-190s)
- macOS E2E: GREEN (19s), Linux E2E: GREEN (46s)
- Dispatch latency: 3s (release published → e2e triggered)
- GoReleaser asset naming: ari_0.9.0_*.tar.gz (no v prefix — standard GoReleaser behavior)
- 2 pre-existing non-blocking failures: ariadne-tests.yml (golangci-lint schema), verify-doctrine.yml (missing ariadne/ dir)

### v0.10.0 — 2026-03-16

| Field | Value |
|-------|-------|
| Date | 2026-03-16 |
| Version | v0.9.0 -> v0.10.0 (minor) |
| Commits | 113 |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 61140cbb97f3a21b0e25a969dc915bf6030db3a5 |
| Commit SHA | f690089eb23bec176f965ae0eaec1b7d516780c9 |
| release.yml | GREEN (110s, run 23138025745) |
| e2e-distribution.yml | GREEN (47s, run 23138095165) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.10.0, commit 97c66a69, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twentieth consecutive PASS verdict (v0.3.2 → ... → v0.9.0 → v0.10.0)
- Total chain: 146s (slightly below historical baseline ~150-190s)
- Dispatch latency: 2s
- macOS E2E: GREEN (20s), Linux E2E: GREEN (44s)
- Cached release knowledge skipped cartographer and dependency-resolver entirely
- 105 commits pushed to origin before tagging (local was ahead of remote)
- Headline features: UI rite (6-agent UX pantheon), HAFP preferential-language lint rule with integration tests, gemini/all-channel e2e validation, harness-agnosticism refactors across Go source/tests/mena/docs, moirai chicken-egg template fix
- 2 pre-existing non-blocking failures: ariadne-tests.yml (golangci-lint schema), verify-doctrine.yml (missing ariadne/ dir)
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)

### v0.10.1 — 2026-03-17

| Field | Value |
|-------|-------|
| Date | 2026-03-17 |
| Version | v0.10.0 -> v0.10.1 (patch) |
| Commits | 4 (2 fix, 1 feat, 1 chore) |
| Complexity | PATCH |
| Tag SHA | fed845cbbf5ac55ae9c8610348471559072a2131 |
| Commit SHA | 9340b49 |
| release.yml | GREEN (129s, run 23200998324) |
| e2e-distribution.yml | GREEN (58s, run 23201093165) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.10.1, 4s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-first consecutive PASS verdict (v0.3.2 -> ... -> v0.10.0 -> v0.10.1)
- Total chain: ~187s (within historical baseline ~150-190s)
- macOS E2E: GREEN (19s), Linux E2E: GREEN (54s)
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 3 pre-existing non-blocking failures: ariadne-tests.yml (golangci-lint schema), verify-doctrine.yml (missing ariadne/ dir), validate-orchestrators.yml
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
- Headline: UI rite agent fixes, frontend-fanatic evaluator agent, Gemini command configs, MCP cleanup

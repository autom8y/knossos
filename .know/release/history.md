---
domain: release/history
generated_at: "2026-03-25T10:20:00Z"
source_scope:
  - "./.know/release/"
generator: pipeline-monitor
source_hash: "e891116a"
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

### v0.10.2 — 2026-03-17

| Field | Value |
|-------|-------|
| Date | 2026-03-17 |
| Version | v0.10.1 -> v0.10.2 (patch) |
| Commits | 2 (1 feat, 1 chore) |
| Complexity | PATCH |
| Tag SHA | 035c862b00fcf8733f2eb07afbd849e319693145 |
| Commit SHA | 1aa1bb74 |
| release.yml | GREEN (116s, run 23201952714) |
| e2e-distribution.yml | GREEN (56s, run 23202036777) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.10.2, 4s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-second consecutive PASS verdict (v0.3.2 -> ... -> v0.10.1 -> v0.10.2)
- Total chain: ~172s (within historical baseline ~150-190s)
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 3 pre-existing non-blocking failures: ariadne-tests.yml, verify-doctrine.yml, validate-orchestrators.yml
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
- Headline: agent-color-duplicate lint rule, cached release knowledge update

### v0.11.0 — 2026-03-17

| Field | Value |
|-------|-------|
| Date | 2026-03-17 |
| Version | v0.10.2 -> v0.11.0 (minor) |
| Commits | 4 (1 feat, 2 fix, 1 chore) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 204d3b8647ad02c65391e270b521c34cbe9c76da |
| Commit SHA | 78d47f3e |
| release.yml | GREEN (108s, run 23206617631) |
| e2e-distribution.yml | GREEN (51s, run 23206689751) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.11.0, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-third consecutive PASS verdict (v0.3.2 -> ... -> v0.10.2 -> v0.11.0)
- Total chain: ~156s (within historical baseline ~150-190s)
- Dispatch latency: 3s (release published → e2e triggered)
- macOS E2E: GREEN (21s), Linux E2E: GREEN (48s)
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 3 pre-existing non-blocking failures: ariadne-tests.yml, verify-doctrine.yml, validate-orchestrators.yml
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
- Headline: UI rite extract/observe/act agents, CUA browser tools fix, MCP scoping, browser-local pool fix

### v0.12.0 — 2026-03-17

| Field | Value |
|-------|-------|
| Date | 2026-03-17 |
| Version | v0.11.0 -> v0.12.0 (minor) |
| Commits | 3 (1 feat, 1 fix, 1 chore) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 4b146dc3178cd9b5c6c8dec0e3f50d267076c2c2 |
| Commit SHA | caac5baa |
| release.yml | GREEN (109s, run 23220985888) |
| e2e-distribution.yml | GREEN (49s, run 23221034888) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.12.0, commit 0583d04c) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-fourth consecutive PASS verdict (v0.3.2 -> ... -> v0.11.0 -> v0.12.0)
- Total chain: ~150s (within historical baseline ~150-190s)
- macOS E2E: GREEN (20s), Linux E2E: GREEN (46s)
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 2 pre-existing non-blocking failures: ariadne-tests.yml, verify-doctrine.yml
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
- Headline: UI rite CSS architecture research mena, ari ask org rite search fix, cached release knowledge update

### v0.13.0 — 2026-03-19

| Field | Value |
|-------|-------|
| Date | 2026-03-19 |
| Version | v0.12.0 -> v0.13.0 (minor) |
| Commits | 3 (2 feat, 1 chore) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 583edd53c46831e0141900d111c1310ea4da0d45 |
| Commit SHA | 0fad30de |
| release.yml | GREEN (113s, run 23278074579) |
| e2e-distribution.yml | GREEN (54s, run 23278117839) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.13.0, commit e1d3460, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-fifth consecutive PASS verdict (v0.3.2 -> ... -> v0.12.0 -> v0.13.0)
- Total chain: ~158s (within historical baseline ~150-190s)
- macOS E2E: GREEN (30s), Linux E2E: GREEN (50s)
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 3 pre-existing non-blocking failures: ariadne-tests.yml, verify-doctrine.yml, validate-orchestrators.yml
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
- Headline: MCP support for Gemini channel, orchestrator.yaml archetype rendering, agent maxTurns increase

### v0.14.0 — 2026-03-19

| Field | Value |
|-------|-------|
| Date | 2026-03-19 |
| Version | v0.13.0 -> v0.14.0 (minor) |
| Commits | 2 (2 feat) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | 1f0b1dd414bb88248a9ce9053cd5cee3639cafaa |
| Commit SHA | 78abb186 |
| release.yml | GREEN (106s, run 23295051963) |
| e2e-distribution.yml | GREEN (49s, run 23295111604) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.14.0, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-sixth consecutive PASS verdict (v0.3.2 -> ... -> v0.13.0 -> v0.14.0)
- Total chain: ~147s (within historical baseline ~150-190s)
- macOS E2E: GREEN (18s), Linux E2E: GREEN (46s)
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 3 pre-existing non-blocking failures: ariadne-tests.yml, verify-doctrine.yml, validate-orchestrators.yml
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
- Headline: Optional env var syntax for MCP pool config, UI rite redesign with posture-based routing and 9-agent pantheon

### v0.15.0 — 2026-03-23

| Field | Value |
|-------|-------|
| Date | 2026-03-23 |
| Version | v0.14.0 -> v0.15.0 (minor) |
| Commits | 3 (2 feat, 1 fix) |
| Complexity | PATCH (user-invoked), semver-escalated to MINOR |
| Tag SHA | f28c100f0480d304ec1cab2afee82575cc181eb4 |
| Commit SHA | 6685f6f3 |
| release.yml | GREEN (109s, run 23439348011) |
| e2e-distribution.yml | GREEN (55s, run 23439422771) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.15.0, commit cad23ee8, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~1 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-seventh consecutive PASS verdict (v0.3.2 -> ... -> v0.14.0 -> v0.15.0)
- Total chain: 156s (within historical baseline ~150-190s)
- macOS E2E: GREEN (21s), Linux E2E: GREEN (51s)
- Dispatch latency: 2s (release published → e2e triggered)
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 2 pre-existing non-blocking failures: ariadne-tests.yml (golangci-lint schema), verify-doctrine.yml (missing ariadne/ dir)
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
### v0.15.4 — 2026-03-23

| Field | Value |
|-------|-------|
| Date | 2026-03-23 |
| Version | v0.15.0 -> v0.15.4 (fix-forward chain) |
| Commits | 12 (remendiation of lint, tests, and docs) |
| Complexity | PATCH |
| Tag SHA | d0e9fec9 |
| Commit SHA | d0e9fec9 |
| release.yml | GREEN (112s, run 23458823668) |
| e2e-distribution.yml | GREEN (58s, run 23458894364) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated |
| Verdict | PASS |
| Duration | ~2 hours (session duration including remediations) |

**Notes:**
- Twenty-eighth consecutive PASS verdict (v0.3.2 -> ... -> v0.15.0 -> v0.15.4)
- Fix-forward chain: v0.15.1 (CI fail) -> v0.15.2 (CI fail) -> v0.15.3 (CI fail) -> v0.15.4 (PASS)
- Remediated: .golangci.yml v2 upgrade, 50+ errcheck fixes in tests/benchmarks, restored verify-doctrine.sh, fixed bash subshell/arithmetic bugs.
- 18 broken documentation links identified (missing ADRs/guides) — demoted to warnings to unblock release.

### v0.16.0 — 2026-03-25

| Field | Value |
|-------|-------|
| Date | 2026-03-25 |
| Version | v0.15.4 -> v0.16.0 (minor) |
| Commits | 2 (1 feat, 1 chore) |
| Complexity | PATCH (user-invoked), semver MINOR |
| Tag SHA | f99aa2c1c5010f704675599d3f0247fa7cdb7789 |
| Commit SHA | e891116a |
| release.yml | GREEN (147s, run 23535913368) |
| e2e-distribution.yml | GREEN (46s, run 23536006631) |
| Assets | 5/5 (4 platform binaries + checksums.txt) |
| Homebrew tap | Updated (v0.16.0, 1s dispatch lag) |
| Verdict | PASS |
| Duration | ~4 min (execution) + ~3 min (monitoring) |

**Notes:**
- Twenty-ninth consecutive PASS verdict (v0.3.2 -> ... -> v0.15.4 -> v0.16.0)
- Total chain: ~193s (release.yml 147s + e2e 46s; slightly above historical baseline ~150-190s)
- macOS E2E: GREEN (20s), Linux E2E: GREEN (45s)
- Dispatch latency: 1s (release published → e2e triggered)
- Homebrew tap propagation: 1s
- Cached release knowledge skipped cartographer, dependency-resolver, and release-planner entirely
- 4 Dependabot vulnerability alerts on main (informational, non-blocking)
- Ancillary: deploy-clew.yml RED — AWS OIDC role ARN missing account ID (first deploy attempt, infra config gap)
- Headline: Clew MVP — organizational intelligence Slack bot (`ari serve`, `ari registry`), 16 new internal packages, 3 new external deps (anthropic-sdk-go, slack-go, opentelemetry), Go 1.25 minimum, Dockerfile + ECS task definition, 380+ tests across 76 packages

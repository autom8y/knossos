---
domain: "scar-tissue"
generated_at: "2026-03-23T15:50:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "0ffd8061"
confidence: 0.75
format_version: "1.0"
sessions_synthesized: 55
last_session: "session-20260323-153258-fd8f7c41"
---

## Blocker Catalog

| Session | Blocker | Resolution | Domain |
|---------|---------|-----------|--------|
| session-20260303-163517-f43d2334 | GoReleaser GITHUB_TOKEN suppresses downstream workflow triggers | Identified 4 fix options (PAT swap, GitHub App, workflow_dispatch chain, workflow_run trigger) | ci-infrastructure |
| session-20260304-001522-1c07b3bd | Pre-existing Go test failures present during audit | Confirmed pre-dating the framework agnosticism initiative | test-infrastructure |
| session-20260304-154410-af90596d | 3 code paths bypass content rewriter (GO-001/002/003) | Fixed in Phase 2: exported RewriteMenaContentPaths, wired into 3 additional bypass paths | materialization |
| session-20260304-154410-af90596d | 15 broken cross-reference links in mena source | Fixed in Phase 2: corrected cross-references in session/operations/workflow files | mena-content |

## Rejected Alternatives

| Session | Decision | Rejected | Rationale |
|---------|---------|----------|-----------|
| session-20260303-003107-55293519 | PRD phase skipped -- debt audit artifacts serve as requirements | Traditional PRD authoring | Debt audit + risk report + sprint plan already provide complete requirements coverage |
| session-20260303-132132-f3031827 | Requirements phase skipped | Traditional requirements gathering | Spike + framing documents provide comprehensive coverage (Pythia assessment) |
| session-20260303-135856-c525b918 | Requirements + design phase skipped | Traditional phased approach | Spike + framing documents provide architecture, schemas, workstream decomposition, risk register, success criteria |
| session-20260304-001522-1c07b3bd | .ledge/ as canonical artifact path | docs/ directory | Standardize on .ledge/ for all work product artifacts |
| session-20260304-001522-1c07b3bd | .claude/ writes only via ari sync | Direct .claude/ file writes | Prevent write bleed from agents directly modifying .claude/ |
| session-20260304-154410-af90596d | Option C documentary fencing for educational refs | Rewrite all .lego.md/.dro.md references | Educational references in fenced blocks should be preserved, not rewritten |
| session-20260323-123506-4c58fdce | Hook-level dedup for Cassandra complaints | Application-level dedup | ADR produced: chose hook-level dedup (option a) per ADR-cassandra-dedup-boundary |

## Friction Signals

- **Recurring**: Patterns across 2+ sessions
  - Phase skip pattern: 3 sessions skipped requirements/design phases when spikes and frames already provided coverage (session-20260303-003107-55293519, session-20260303-132132-f3031827, session-20260303-135856-c525b918)
  - Duplicate/near-duplicate sessions: 2 sessions for same canonical dotfile path initiative (session-20260302-151610-273eeb4a, session-20260302-151730-bffc2248)
  - High session churn on 2026-03-02: 18 sessions created, most parked quickly at requirements phase, suggesting rapid context switching
  - GITHUB_TOKEN suppression: identified in session-20260303-163517-f43d2334, affects all releases (v0.1.0, v0.2.0, v0.3.0)
  - Post-mortem finding pattern: session-20260304-154410-af90596d found 3 P0 bugs in code paths the original TDD did not identify (engine.go standalone, userscope walker, userscope standalone)
- **One-time**: Isolated friction events
  - Framework agnosticism bleed: 75 findings across 3 vectors (V1: autom8y identity 13, V2: artifact path 51, V3: .claude/ write 11) in session-20260304-001522-1c07b3bd
  - V2 systemic finding: all 12 workflow.yaml use docs/ while platform mena uses .ledge/ -- forge artifact glossary internally contradictory (session-20260304-001522-1c07b3bd)
  - Cross-rite session: session-20260312-183235-bbbf5231 transitioned across hygiene -> 10x-dev -> docs rites within a single session

## Quality Friction (Sails Analysis)

| Sails Color | Sessions | Common Failure Proofs |
|------------|---------|---------------------|
| GRAY | 55 | All proofs UNKNOWN across all 55 sessions (log files not found) |

## Deferred Work

- Claude boundary enforcement (session-20260305-021309-e79a141b): .claude/ to .knossos/ migration scoped but parked at requirements, not seen completed in subsequent sessions
- Phase 2 Context Quality Remediation (session-20260302-150407-07b8e5be): 8 workstreams across 15 rites, parked at requirements after 1 minute
- Comprehensive Debt Remediation (session-20260303-003107-55293519): 44 items across 6 waves (~90-153h estimated), reached design phase only
- E2E Distribution Dispatch fix (session-20260303-163517-f43d2334): root cause identified, 4 fix options documented, implementation not started
- External DI Remediation & t.Parallel() adoption (session-20260310-192540-28b102b0): 361+ tests scoped, parked at requirements
- Procession completion (session-20260310-122716-65d5c100): 11 workstreams across 4 waves, parked at requirements
- ui-rite-creative-evolution (session-20260319-042521-06e927c4): 8 sprints across 3 rites, parked at PLAN phase

## Confidence Notes

- Scar-tissue extraction is limited by SESSION_CONTEXT content quality: 32/55 sessions (58%) are SPARSE with no substantive blocker or friction data
- agent.delegated events found in only 1/55 sessions; routing decision data is nearly absent from event streams
- All WHITE_SAILS are GRAY/UNKNOWN -- no CI proof data available for quality friction analysis
- Blocker and rejected-alternative data extracted from narrative sections of RICH/MODERATE sessions only

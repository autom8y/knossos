---
domain: "scar-tissue"
generated_at: "2026-03-13T12:00:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "59a0de2"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 48
last_session: "session-20260312-183235-bbbf5231"
---

## Blocker Catalog

| Session | Blocker | Resolution | Domain |
|---------|---------|-----------|--------|
| session-20260302-114829-b2f806a3 | Paused debt audit to run deep spike on org-level best practices root cause analysis | Parked session, pivoted to spike | planning |
| session-20260302-122533-2820b8a1 | Parking Sprint 3 to start Square Zero Debt Remediation initiative | Scope shift to higher-priority initiative | planning |
| session-20260303-163517-f43d2334 | GoReleaser uses GITHUB_TOKEN which GitHub suppresses from triggering downstream workflows | Options: swap for PAT, GitHub App token, explicit workflow_dispatch chain, or workflow_run trigger | ci-infrastructure |
| session-20260304-001522-1c07b3bd | Pre-existing Go test failures present during framework agnosticism audit | Confirmed pre-dating the initiative; not caused by audit work | test-infrastructure |
| session-20260304-154410-af90596d | Post-mortem revealed 3 code paths bypassing content rewriter (engine.go standalone, userscope walker, userscope standalone) | Exported RewriteMenaContentPaths and wired into all 3 bypass paths in Phase 2 | materialization-pipeline |

## Rejected Alternatives

| Session | Decision | Rejected | Rationale |
|---------|---------|----------|-----------|
| session-20260303-003107-55293519 | PRD phase skipped for debt remediation | Writing a fresh PRD | Debt audit artifacts (LEDGER, RISK-REPORT, SPRINT-PLAN) already serve as requirements |
| session-20260303-132132-f3031827 | Requirements phase skipped for .know/ incremental refresh | Formal requirements phase | Spike + framing documents provide comprehensive coverage |
| session-20260303-135856-c525b918 | Requirements + design phase skipped for release context persistence | Formal requirements and design phases | Spike + framing documents provide comprehensive coverage including architecture, schemas, workstream decomposition |
| session-20260303-163517-f43d2334 | Option A (swap GITHUB_TOKEN for PAT) or Option D (workflow_run trigger) | Option B (GitHub App token): high setup effort; Option C (explicit workflow_dispatch chain): requires PAT | Effort vs. complexity tradeoff; A requires PAT scope check, D is zero-PAT |
| session-20260304-001522-1c07b3bd | .ledge/ as canonical artifact path | docs/ as artifact path | Platform mena uses .ledge/; docs/ was ecosystem bleed from workflow.yaml files |
| session-20260304-001522-1c07b3bd | .claude/ writes only via ari sync | Direct .claude/ writes from agents | Framework agnosticism requires binary to not assume specific ecosystem |
| session-20260304-154410-af90596d | Option C documentary fencing for educational .lego.md/.dro.md refs | Rewriting educational references (would corrupt documentation) | Educational references should be preserved verbatim in code fences |

## Friction Signals

- **Recurring**: Patterns across 2+ sessions
  - Session parking at requirements phase: 32/48 sessions never progressed past requirements, seen across all rites
  - Duplicate/retry sessions for same initiative: session-20260302-151610-273eeb4a and session-20260302-151730-bffc2248 (canonical dotfile path remediation started twice within 1 minute)
  - Auto-parked on Stop: frequent auto-parking pattern, seen in 30+ sessions
  - GITHUB_TOKEN suppression: documented in session-20260303-163517-f43d2334, broader pattern of CI infrastructure friction
  - Framework agnosticism bleed: 75 findings in session-20260304-001522-1c07b3bd indicating systemic ecosystem coupling in distributed binary
  - Phase skipping pattern: 3 sessions (003107, 132132, 135856) skipped formal requirements/design phases because spike + framing documents provided sufficient coverage
  - Materialization bypass bugs: content rewriting implemented in primary path but missed bypass paths (engine.go standalone, userscope walker, userscope standalone) in session-20260304-154410-af90596d
  - t.Setenv friction: sessions 20260310-174131 and 20260310-192540 both target t.Setenv elimination, indicating systemic test parallelization blocker
  - Cross-rite session transitions: session-20260312-183235-bbbf5231 transitioned hygiene -> 10x-dev -> docs across sprints, indicating initiative scope crossing rite boundaries

- **One-time**: Isolated friction events
  - Pre-existing Go test failures during hygiene audit: session-20260304-001522-1c07b3bd
  - V2 (artifact path bleed) is systemic -- all 12 workflow.yaml use docs/ while platform mena uses .ledge/: session-20260304-001522-1c07b3bd
  - Forge artifact glossary internally contradictory: session-20260304-001522-1c07b3bd
  - 15 broken cross-reference links discovered during post-mortem of mena content path hygiene: session-20260304-154410-af90596d

## Quality Friction (Sails Analysis)

| Sails Color | Sessions | Common Failure Proofs |
|------------|---------|---------------------|
| GRAY | 48 | All proofs UNKNOWN (no CI proof infrastructure captured) |
| WHITE | 0 | N/A |
| BLACK | 0 | N/A |

## Deferred Work

- Resume cross-rite: deferred in CC Agent Capability Uplift, needs empirical evidence from ecosystem usage
- arch-ref skill creation: deferred, arch rite has no reference skill
- ADR-0028: deferred, needs empirical evidence from pilot and rollout
- WS-5 (optional) in mena materialization divergence fix: unify user-scope walkers via fs.FS interface, deferred in session-20260304-135616-f6f65e42
- Wave 3 (Userscope Deep) in debt remediation: last priority, highest risk, deferred in session-20260303-003107-55293519
- WS-4 (documentation of GITHUB_TOKEN suppression): deferred in session-20260303-163517-f43d2334
- Phase 2 Context Quality Remediation: 8 workstreams across 15 rites, session-20260302-150407-07b8e5be parked at requirements with detailed multi-session plan
- knowledge-maturation-pipeline: session-20260305-172543-d0e8d2fc parked at requirements
- Comprehensive Debt Remediation 44 items: session-20260303-003107-55293519 parked at design phase, Waves 1-4 pending execution
- Procession completion: 11 workstreams across 4 waves, session-20260310-122716-65d5c100 parked at requirements
- External DI Remediation and t.Parallel() Adoption: 4 workstreams (WS-A through WS-D) across 361+ tests, session-20260310-192540-28b102b0 parked at requirements
- Cassandra P0 Complaint Schema and Filing Infrastructure: session-20260311-012734-9847ff6f parked at requirements
- multi-channel-integration: session-20260311-231845-47df6625 parked at design
- security-remediation: session-20260312-113227-1c6e9313 parked at requirements
- harness-agnosticism: session-20260312-123128-56a5e48f parked at requirements

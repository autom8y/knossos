---
domain: "scar-tissue"
generated_at: "2026-03-06T21:00:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "3053e84"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 37
last_session: "session-20260306-201337-8581513b"
---

## Blocker Catalog

| Session | Blocker | Resolution | Domain |
|---------|---------|-----------|--------|
| session-20260302-150407-07b8e5be | SD-1: @skill-name anti-pattern in orchestrator.yaml + pythia.md (12/15 rites) | Planned: WS1 platform code fix in generateSkillsReference() | agent-regeneration |
| session-20260302-150407-07b8e5be | SD-2: Pythia duplicate sections (12/15 rites) | Planned: WS3 pythia regeneration | agent-regeneration |
| session-20260302-150407-07b8e5be | SD-3: Catalog model contradictions (8/15 rites) | Planned: WS4 ground truth sync | agent-config |
| session-20260302-150407-07b8e5be | SD-4: Entry-point conflict between manifest.yaml and workflow.yaml (9/15 rites) | Planned: WS5 entry-point semantics | rite-config |
| session-20260302-150407-07b8e5be | SD-5: Ghost skill references in prose (14/15 rites) | Planned: WS6 ghost prose cleanup | mena-integrity |
| session-20260302-150407-07b8e5be | SD-6: Oversized injected INDEX skills exceeding 40-line ceiling (9/15 rites) | Planned: WS7 progressive disclosure splits | context-budget |
| session-20260303-163517-f43d2334 | GITHUB_TOKEN suppresses downstream workflow triggers | Root cause: GoReleaser uses GITHUB_TOKEN; GitHub silently drops release events. Fix options A-D identified. | ci-infrastructure |
| session-20260304-154410-af90596d | GO-001/002/003: Three content rewriting bypass paths in materialization pipeline | Fixed: RewriteMenaContentPaths exported and wired into engine.go standalone + both userscope code paths | materialization |
| session-20260304-154410-af90596d | 15 broken cross-reference links in mena session/operations/workflow | Fixed: session-common -> common, handoff-ref -> handoff, start-ref -> start naming corrections | mena-integrity |

## Rejected Alternatives

| Session | Decision | Rejected | Rationale |
|---------|---------|----------|-----------|
| session-20260303-003107-55293519 | Skip PRD phase; use debt audit artifacts as requirements | Full PRD generation | Debt audit LEDGER + RISK-REPORT + SPRINT-PLAN already serve as comprehensive requirements |
| session-20260303-132132-f3031827 | Skip requirements + design phases | Standard phase flow | Spike + framing documents provide comprehensive coverage (Pythia assessment) |
| session-20260303-135856-c525b918 | Skip requirements + design phases | Standard phase flow | Spike + framing provide architecture, schemas, workstream decomposition, risk register, success criteria |
| session-20260303-163517-f43d2334 | Option A: swap GITHUB_TOKEN for PAT (recommended if HOMEBREW_TAP_TOKEN has contents:write) | Option B: GitHub App token (high setup effort) | Effort disproportionate to fix scope |
| session-20260303-163517-f43d2334 | Option D: workflow_run trigger (recommended if no PAT available) | Option C: explicit workflow_dispatch chain (requires PAT) | Avoid PAT requirement if possible |
| session-20260304-001522-1c07b3bd | .ledge/ is canonical artifact path | docs/ as artifact path | 12 workflow.yaml used docs/ while platform mena used .ledge/; unified to .ledge/ |
| session-20260304-001522-1c07b3bd | Distribution infra accepted as org-specific (INTENTIONAL) | Remove all org-specific CI references | goreleaser, homebrew tap, GitHub Actions CI are legitimately org-specific |
| session-20260304-001522-1c07b3bd | .claude/ writes only via ari sync | Direct .claude/ writes by agents | Protect materialization integrity |
| session-20260304-154410-af90596d | Option C documentary fencing (fence blocks exempt content from rewriting) | Option A: whitelist specific files; Option B: regex-based content detection | Fencing is explicit, maintainable, and preserves intent without fragile heuristics |

## Friction Signals

- **Recurring**: Patterns across 2+ sessions
  - Session parking at requirements phase: 25/37 sessions never left requirements (68%), suggesting rapid-fire initiative creation with deferred execution
  - Duplicate initiative starts: session-20260302-151610-273eeb4a and session-20260302-151730-bffc2248 are the same initiative started 80 seconds apart
  - Cross-reference link rot: seen in session-20260304-154410-af90596d (15 broken links) and session-20260304-001522-1c07b3bd (75 findings across 3 vectors)
  - Requirements-phase-only sessions dominating archive: pattern across 2026-03-02 sessions (13/18 parked at requirements) and 2026-03-06 sessions (5/6 parked at requirements)
  - Phase skipping for spike/framing-driven work: session-20260303-132132-f3031827, session-20260303-135856-c525b918 both skipped requirements phase based on pre-existing spike artifacts
  - Multi-phase embody sessions: session-20260305-191804-8e309d1f, session-20260306-084705-0f5ea0c7, session-20260306-092108-0c1f2ce9 (3 sessions for phases 1-3 of same feature, all parked at different phases)
- **One-time**: Isolated friction events
  - GoReleaser GITHUB_TOKEN event suppression blocking CI pipeline: session-20260303-163517-f43d2334
  - Content rewriting bypass in materialization pipeline (3 code paths missed by initial TDD): session-20260304-154410-af90596d
  - Pre-existing Go test failures complicating audit signoff: session-20260304-001522-1c07b3bd

## Quality Friction (Sails Analysis)

| Sails Color | Sessions | Common Failure Proofs |
|------------|---------|---------------------|
| GRAY | 37 | All proofs UNKNOWN across all sessions (log files not found) |
| WHITE | 0 | N/A |
| BLACK | 0 | N/A |

- 0/37 sessions have any proof status other than UNKNOWN
- Universal GRAY is caused by absence of CI log integration, not by test failures
- Sessions with manual test verification (confirmed green via go test ./...): session-20260302-123250-636200b5, session-20260304-154410-af90596d

## Deferred Work

- Phase 2 Context Quality Remediation (8 workstreams, 21-30h estimated): session-20260302-150407-07b8e5be, parked at requirements
- Comprehensive Debt Remediation (44 items, 90-153h estimated): session-20260303-003107-55293519, parked at design
- Canonical dotfile path remediation: session-20260302-151610-273eeb4a, not seen in subsequent sessions
- XDG Mena Staleness: session-20260302-223308-10f40f3c, not seen in subsequent sessions
- Knowledge Radar: session-20260302-223852-c7b44b21, not seen in subsequent sessions
- Org tier sprints 1-3: session-20260302-121512-f4f4f4a6, session-20260302-121832-e059c280, session-20260302-122533-2820b8a1 -- all parked at requirements
- sync-noise-remediation: session-20260302-210626-4a7b9681, not seen in subsequent sessions
- Releaser rite improvements (two sessions): session-20260302-122213-1e802b37, session-20260302-131246-b6788294, both parked
- .know/ Incremental Refresh: session-20260303-132132-f3031827, parked at implementation
- Release Context Persistence: session-20260303-135856-c525b918, parked at implementation
- P4 low-priority findings from content path hygiene: session-20260304-154410-af90596d (deferred as boy-scout fixes)
- E2E Distribution Dispatch fix implementation: session-20260303-163517-f43d2334, investigation complete but fix not applied
- embody phase 2 workflow awareness: session-20260306-084705-0f5ea0c7, parked at requirements
- embody phase 3 L8 horizon: session-20260306-092108-0c1f2ce9, parked at requirements
- Session Dromena Legacy Removal: session-20260306-152208-d302cc16, parked at requirements
- Release hardening post legacy removal: session-20260306-164058-a8e7e897, parked at requirements
- Hygiene printer bypasses + fmt.Errorf cleanup: session-20260306-201337-8581513b, parked at requirements

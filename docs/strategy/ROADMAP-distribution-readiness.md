# Strategic Roadmap: Knossos Distribution Readiness

**Date**: 2026-02-08
**Status**: Active
**Owner**: autom8y / Strategy Rite
**Upstream Artifacts**:
- [MARKET-distribution-readiness.md](./MARKET-distribution-readiness.md) -- Market sizing, adoption patterns, customer segments
- [COMPETITIVE-distribution-readiness.md](./COMPETITIVE-distribution-readiness.md) -- Competitor profiles, platform risk, differentiation
- [BUSINESS-MODEL-distribution-readiness.md](./BUSINESS-MODEL-distribution-readiness.md) -- Financial model, unit economics, investment timeline
- [STAKEHOLDER-PREFERENCES-distribution-readiness.md](../STAKEHOLDER-PREFERENCES-distribution-readiness.md) -- Vision, priorities, technical state, decisions
**Methodology**: RICE prioritization framework applied to stakeholder-specified work sequence, resource-constrained to observed codebase state
**Confidence Level**: Medium-High (strong upstream data; timeline estimates carry inherent uncertainty)

---

## Executive Summary

Knossos -- "Rails for Claude Code" -- sits at the optimal inflection point for a context-engineering framework: the category is named, no framework leader exists, and the competitive window is approximately 12 months (Q2 2026 - Q2 2027). The product is 95% built. The remaining work is the last mile: fixing foundations, completing the oracle experience, and packaging for distribution.

This roadmap translates the stakeholder's four-step work sequence into three distribution stages with concrete workstreams, measurable OKRs, and explicit go/no-go gates. The total investment to reach revenue-generating distribution is 404-807 engineering hours ($81K-$162K equivalent). The biggest risk is not cost -- it is delay. Every month of slip compresses the competitive window.

**The three stages**:

| Stage | Timeline | Users | Investment | Goal |
|-------|----------|-------|------------|------|
| 1. Internal | Now -- End of Q1 2026 | 5-15 | 38-75 hrs | Validate product quality, achieve 30-minute onboarding |
| 2. Trusted External | Q2 -- Q3 2026 | 500-5,000 | 142-284 hrs | External validation, community seed, distribution infrastructure |
| 3. Broader Availability | Q4 2026 -- Q1 2027 | 5,000-50,000 | 224-448 hrs | Revenue generation, community network effects |

**The critical path**: Fix foundations (P0 bugs) --> Strengthen systems (P1 completions) --> Make Pythia world-class --> Distribution packaging. The first three steps are the stakeholder's explicit sequence. Distribution packaging runs in parallel once foundations are solid.

**The single most important metric**: Time from clone to first productive session. If it exceeds 30 minutes for internal users, nothing else matters.

---

## 1. Phase Plan with Milestones

### 1.1 Stage 1: Internal-First (Now -- End of Q1 2026)

**Objective**: Every internal team member productive within 30 minutes via /consult.

**Distribution method**: Clone + build (`CGO_ENABLED=0 go build ./cmd/ari`)

#### Milestones

| # | Milestone | Target Date | Definition of Done |
|---|-----------|-------------|-------------------|
| M1.1 | Build is green | Week 1 | `CGO_ENABLED=0 go build ./cmd/ari` succeeds on clean checkout; all 1,347 tests pass |
| M1.2 | P0 issues resolved | Week 2 | Build fix, DMI additions, consultant Task removal, session status normalization, Fates references addressed |
| M1.3 | /consult baseline | Week 4 | All 5 modes produce correct output; zero dead-reference errors; dynamic exploration works |
| M1.4 | P1 issues resolved | Week 6 | Companion file leakage fixed, hook.Result deprecated, Moirai guidance aligned, missing rules added |
| M1.5 | Pythia world-class | Week 8 | Rite-discovery skill created, capability-index.yaml current, knowledge base populated, all cross-references resolve |
| M1.6 | Internal validation | Week 10 | 3+ internal users complete full /task cycle independently; onboarding time consistently under 30 minutes |

#### Stage 1 Go/No-Go Gate

**Go criteria (ALL must be true)**:
- [ ] Build succeeds on clean checkout
- [ ] All P0 issues resolved
- [ ] `ari sync` produces clean output on a fresh project
- [ ] /consult answers framework questions accurately across all 5 modes
- [ ] At least 3 internal users have completed a full /task cycle independently
- [ ] Time-to-first-productive-session measured and consistently under 30 minutes
- [ ] At least 1 internal champion identified (visible advocate, not just user)

**No-go triggers (ANY is sufficient to delay)**:
- Average onboarding time exceeds 45 minutes
- More than 20% of /consult queries return incorrect or dead-reference answers
- Internal users report "I'd rather just write my own CLAUDE.md"
- P0 bugs remain unresolved after 4 weeks of focused effort

---

### 1.2 Stage 2: Trusted External (Q2 -- Q3 2026)

**Objective**: 100+ external users actively using Knossos with 50%+ 30-day retention.

**Distribution method**: Private Homebrew tap (`brew tap autom8y/tap && brew install ari`)

#### Milestones

| # | Milestone | Target Date | Definition of Done |
|---|-----------|-------------|-------------------|
| M2.1 | Standalone bootstrap | Q2 W4 | `ari init` works without Knossos repo checkout (embedded rites) |
| M2.2 | Homebrew distribution | Q2 W6 | GoReleaser pipeline produces cross-platform binaries; private tap works on macOS and Linux |
| M2.3 | Shell migration complete | Q2 W8 | Zero shell script dependencies in core user flow; all 124 scripts evaluated, critical paths ported to Go |
| M2.4 | Documentation ready | Q3 W2 | README with architecture overview, plain-English glossary, 3 tutorial rites with walkthroughs |
| M2.5 | Community infrastructure | Q3 W4 | Public GitHub repo, issue templates, Discord/Slack channel, contributing guide |
| M2.6 | External validation | Q3 W8 | 100+ active external users; 30-day retention exceeds 50%; 3+ community-created rites or agents |

#### Stage 2 Go/No-Go Gate

**Go criteria (ALL must be true)**:
- [ ] `brew tap autom8y/tap && brew install ari` works on macOS and Linux
- [ ] `ari init` works without Knossos repo checkout (embedded rites)
- [ ] Documentation covers: installation, first rite, first session, /consult usage, custom rite creation
- [ ] Shell script migration complete for core user flows
- [ ] At least 100 trusted external users actively using the tool
- [ ] 30-day retention exceeds 50% among trusted external cohort
- [ ] At least 3 community-created rites or agents exist
- [ ] No critical bugs open for more than 7 days
- [ ] Error messages are user-friendly (no raw stack traces or cryptic failures)

**No-go triggers (ANY is sufficient to delay)**:
- 30-day retention below 30% among trusted external users
- More than 50% of issue reports are onboarding/installation related
- Anthropic announces native project template feature (reassess, do not automatically halt)
- Shell-to-binary migration blocked by technical issues

---

### 1.3 Stage 3: Broader Availability (Q4 2026 -- Q1 2027)

**Objective**: 5,000+ free users, 5+ enterprise design partners, enterprise tier launched.

**Distribution method**: Public Homebrew tap + GitHub Releases + potential CC plugin

#### Milestones

| # | Milestone | Target Date | Definition of Done |
|---|-----------|-------------|-------------------|
| M3.1 | Enterprise feature gating | Q4 W4 | Session lifecycle, clew, White Sails gated behind license; free tier fully functional |
| M3.2 | Licensing system | Q4 W6 | License key generation, validation, self-service portal operational |
| M3.3 | Documentation site | Q4 W8 | Static site (Docusaurus/Hugo) with search, tutorials, API reference |
| M3.4 | Rite marketplace | Q1 2027 W4 | Discover, share, install community rites; 10+ community rites available |
| M3.5 | Revenue generation | Q1 2027 W8 | 5+ paying enterprise customers; pricing validated through design partner feedback |

#### Stage 3 Go/No-Go Gate

**Go criteria (ALL must be true)**:
- [ ] Free user base exceeds 5,000
- [ ] At least 5 enterprise design partners have expressed willingness to pay
- [ ] Enterprise features are production-quality with documentation
- [ ] Licensing system tested and operational
- [ ] Support infrastructure scales (self-service docs, community channels, enterprise SLA)
- [ ] Pricing validated through design partner conversations
- [ ] Legal review of license terms and enterprise agreement complete

**No-go triggers (ANY is sufficient to delay)**:
- Free user growth plateaued below 5,000
- Design partners say "interesting but not worth paying for"
- Competitive landscape shifted materially (Anthropic native framework, dominant competitor emerged)
- Enterprise features do not demonstrably improve team outcomes vs. free tier

---

## 2. Execution Workstreams

The stakeholder specified a four-step work sequence. This section maps that sequence into concrete workstreams with prioritized tasks, dependencies, and parallelization opportunities.

### 2.1 Workstream 1: Fix Foundations (Weeks 1-2)

**Stakeholder directive**: "Fix the foundations -- commit the tree, fix bugs that break things, clean stale references."

**Total estimated effort**: 5-9 hours (P0 only)

| Task ID | Task | Priority | Effort | Dependencies | Parallelizable |
|---------|------|----------|--------|--------------|----------------|
| F-01 | Fix build (SetSyncDir reference) | P0 | 1-2 hrs | None | Yes |
| F-02 | Add DMI to 5-6 multi-agent cascade commands | P0 | 2-4 hrs | None | Yes |
| F-03 | Remove Task tool from consultant agent + mena source | P0 | 1 hr | None | Yes |
| F-04 | Normalize 3 invalid session statuses (COMPLETED/COMPLETE to ARCHIVED) | P0 | 1-2 hrs | None | Yes |
| F-05 | Address Moirai references to non-existent Fate files | P0 | 1-2 hrs | None | Yes |

**Execution note**: All P0 tasks are independent and can be parallelized. A focused sprint can resolve all five in a single day. Each should be committed atomically to main per the stakeholder's git workflow decision.

**Commands requiring DMI addition** (F-02 specifics):
- Identify the 5-6 dromena that launch multi-agent cascades (e.g., /task, /sprint, /build, /architect, and others)
- Add `disable-model-invocation: true` to frontmatter
- Verify DMI is NOT added to commands that Moirai legitimately needs to invoke

---

### 2.2 Workstream 2: Strengthen Systems (Weeks 2-6)

**Stakeholder directive**: "Complete missing skills, fix the lexicon, align the knowledge base."

**Total estimated effort**: 33-66 hours (P1 items)

| Task ID | Task | Priority | Effort | Dependencies | Parallelizable |
|---------|------|----------|--------|--------------|----------------|
| S-01 | Create rite-discovery skill (mena source exists, needs materialization + content) | P1 | 4-8 hrs | F-05 | Partially |
| S-02 | Fix capability-index.yaml (update stale command names, verify data) | P1 | 2-4 hrs | None | Yes |
| S-03 | Fix companion file autocomplete leakage (~60 to ~30 entries) | P1 | 4-8 hrs | None | Yes |
| S-04 | Mark hook.Result struct as deprecated with clear guidance | P1 | 1-2 hrs | None | Yes |
| S-05 | Fix conflicting Moirai guidance (align dromena with moirai-invocation docs) | P1 | 2-4 hrs | None | Yes |
| S-06 | Add missing .claude/rules/ for 4 packages (inscription, hook, sails, usersync) | P1 | 4-8 hrs | None | Yes |
| S-07 | Moirai Fates design review | P1 | 4-8 hrs | F-05, S-05 | No (sequential) |
| S-08 | Create Moirai Fate skill stubs (after design review) | P1 | 4-8 hrs | S-07 | No (sequential) |
| S-09 | Fix MOIRAI_BYPASS mechanism (code checks "1", docs say "true", CC Bash limitation) | P1 | 2-4 hrs | S-05 | Yes |
| S-10 | Dead reference audit and categorization (batch for stakeholder review) | P1 | 4-8 hrs | F-05 | Yes |

**Parallelization strategy**: S-01 through S-06 and S-09/S-10 are largely independent. S-07 (Fates design review) must precede S-08 (Fate stubs). S-05 should precede S-07 and S-09 for coherent Moirai understanding.

**Critical path**: S-07 --> S-08 is the longest sequential dependency (8-16 hours). Start the Fates design review early in Week 3 to avoid blocking the Pythia workstream.

---

### 2.3 Workstream 3: Make Pythia World-Class (Weeks 4-8)

**Stakeholder directive**: "With solid foundations, the consultant can route accurately and deeply."

**Total estimated effort**: 8-16 hours (dedicated /consult work, building on Workstream 2 completions)

| Task ID | Task | Priority | Effort | Dependencies | Parallelizable |
|---------|------|----------|--------|--------------|----------------|
| P-01 | Validate /consult no-args mode (ecosystem overview) | P1 | 1-2 hrs | S-01, S-02 | No (sequential test) |
| P-02 | Validate /consult query mode ("I need to add auth" routes correctly) | P1 | 1-2 hrs | S-01, S-02 | After P-01 |
| P-03 | Validate /consult --rite mode (all 10 rites, accurate descriptions) | P1 | 2-4 hrs | S-01, S-02 | After P-01 |
| P-04 | Validate /consult --commands mode (complete, categorized reference) | P1 | 2-4 hrs | S-02 | After P-01 |
| P-05 | Validate /consult --playbook mode (create or remove playbook references) | P1 | 2-4 hrs | S-01 | After P-01 |
| P-06 | Fix consultant model reference (remove "Claude Opus 4.5" hardcode) | P0.5 | 0.5 hr | F-03 | Yes |
| P-07 | End-to-end /consult acceptance test (all modes, all cross-references) | P1 | 2-4 hrs | P-01 through P-05 | No (final gate) |

**Definition of "world-class"** (from stakeholder):
- `/consult` (no args): Produces accurate ecosystem overview with current state
- `/consult "I need to add auth"`: Routes correctly to 10x-dev + /task with command-flow
- `/consult --rite`: Shows all 10 user-facing rites with accurate descriptions and commands
- `/consult --commands`: Shows complete, categorized command reference
- Zero 404s: every skill reference, knowledge base file, and cross-reference resolves
- Dynamic exploration works: consultant can read rite manifests and agent files on demand

**Dependency insight**: This workstream cannot start in earnest until Workstream 2 deliverables S-01 (rite-discovery), S-02 (capability-index), and the foundation of F-03 (Task tool removal) are complete. The stakeholder explicitly noted: "The consult will only be as powerful as the underlying concepts, architecture, lexicon, systems, etc."

---

### 2.4 Workstream 4: Moirai Fates Design Review (Week 6-8)

**Stakeholder directive**: "Deep contextual review of the architecture from an agent's perspective."

This is a distinct workstream because the stakeholder explicitly called it out as requiring careful design review before implementation. It overlaps with Workstream 2 (S-07, S-08) but has its own success criteria.

| Task ID | Task | Priority | Effort | Dependencies | Parallelizable |
|---------|------|----------|--------|--------------|----------------|
| MF-01 | Review TDD-fate-skills.md against current architecture | P1 | 2-4 hrs | S-05 | No |
| MF-02 | Verify progressive disclosure pattern for Fate skills | P1 | 1-2 hrs | MF-01 | No |
| MF-03 | Confirm skill loading mechanism (Read vs. Skill tool) | P1 | 1 hr | MF-01 | No |
| MF-04 | Create Fate skill stubs (clotho, lachesis, atropos) | P1 | 4-8 hrs | MF-01 through MF-03 | No |
| MF-05 | Present design review findings to stakeholder for approval | P1 | 1 hr | MF-04 | No |

**Stakeholder note**: "The TDD is probably correct but needs deep contextual review." This is judgment-call territory -- execute with care and ask on ambiguity per the stakeholder's autonomy model.

---

### 2.5 Workstream 5: Distribution Packaging (Stage 2 prep, Weeks 6-14)

This workstream runs in parallel with Workstreams 3-4 once foundations are solid. It is the bridge from Stage 1 to Stage 2.

| Task ID | Task | Priority | Effort | Dependencies | Parallelizable |
|---------|------|----------|--------|--------------|----------------|
| D-01 | GoReleaser setup + private Homebrew tap | P1 | 8-16 hrs | M1.1 (build green) | Yes |
| D-02 | `ari init` without repo checkout (embedded rites) | P1 | 16-32 hrs | M1.2 (P0 resolved) | No (largest item) |
| D-03 | Cross-platform testing (macOS + Linux CI matrix) | P1 | 8-16 hrs | D-01 | After D-01 |
| D-04 | Shell script migration (critical paths to Go) | P1 | 40-80 hrs | M1.2 | Yes (can start early) |
| D-05 | Error message quality pass | P1 | 8-16 hrs | D-02 | After D-02 |
| D-06 | Public README with architecture overview | P2 | 8-16 hrs | M1.5 (Pythia ready) | Yes |
| D-07 | Plain-English mythology glossary | P2 | 4-8 hrs | None | Yes |
| D-08 | 3-5 tutorial rites with walkthroughs | P2 | 16-32 hrs | D-02 | After D-02 |
| D-09 | Public GitHub repo setup (license, CI, issue templates) | P2 | 4-8 hrs | None | Yes |
| D-10 | Community channel setup (Discord/Slack) | P2 | 2-4 hrs | D-09 | After D-09 |

**Critical path for Stage 2**: D-02 (ari init standalone) is the single biggest blocker. It requires embedded rites -- a significant architectural change. Start investigation in Week 6; target completion by end of Week 12.

**Shell migration strategy** (D-04): The 124 scripts / 39K LOC do not all need porting. Prioritize:
1. Scripts invoked by core user flows (build, init, materialize, session lifecycle)
2. Scripts invoked by hooks
3. Scripts used in CI
4. Defer: utility scripts, one-off tools, development helpers

---

### 2.6 Workstream Summary: Dependency Graph

```
Week 1-2:  [W1: Fix Foundations] -----> all P0s resolved
                |
Week 2-6:  [W2: Strengthen Systems] -----> all P1s resolved
                |         |
Week 4-8:  [W3: Pythia]  [W4: Fates Review]
                |              |
Week 6-14: [W5: Distribution Packaging] -----> Stage 2 ready
```

Workstreams 1 and 2 are sequential (2 depends on 1). Workstreams 3 and 4 can overlap with late Workstream 2. Workstream 5 can start once the build is green (M1.1) and runs in parallel through Stage 2 preparation.

---

## 3. Prioritization Matrix

### 3.1 Framework: RICE

Applied to the major initiative clusters, scored relative to the internal-first stage and the 12-month competitive window.

| Initiative | Reach | Impact (1-3) | Confidence | Effort (person-weeks) | RICE Score | Rank |
|------------|-------|--------------|------------|----------------------|------------|------|
| P0 bug fixes (Workstream 1) | 15 (all internal users) | 3 (blocking) | 95% | 0.25 | 171 | **1** |
| /consult world-class (Workstream 3) | 15 (internal) + 5K (Stage 2) | 3 (flagship) | 80% | 1.5 | 80 | **2** |
| Strengthen systems (Workstream 2) | 15 (internal) | 2 (enabling) | 90% | 3 | 9 | **3** |
| ari init standalone (D-02) | 5K (Stage 2 users) | 3 (blocker for Stage 2) | 70% | 3 | 35 | **4** |
| Shell migration (D-04) | 5K (Stage 2 users) | 2 (reliability) | 60% | 8 | 7.5 | **5** |
| Moirai Fates review (Workstream 4) | 15 (internal) | 2 (foundation) | 70% | 1 | 21 | **6** |
| GoReleaser + Homebrew (D-01) | 5K (Stage 2 users) | 2 (distribution) | 85% | 1.5 | 57 | **7** |
| Documentation (D-06 through D-08) | 5K (Stage 2 users) | 2 (adoption) | 75% | 4 | 19 | **8** |

**Decision**: Execute in rank order, with parallelization where dependencies allow. P0 fixes first (highest RICE, lowest effort). /consult quality second (highest impact on flagship experience). Everything else follows.

### 3.2 What We Are Explicitly NOT Doing

Per stakeholder out-of-scope list and RICE scoring:

| Deferred Initiative | Rationale |
|---------------------|-----------|
| New rites or agents | Fix and polish what exists first |
| CC feature parity (permissionMode, mcpServers, skills field, etc.) | Does not break anything; defer |
| Performance optimization | Hook latency, materialization speed are fine |
| Distribution packaging (Stage 1) | Product must be ready first |
| SubagentStart/SubagentEnd hooks | Env infrastructure exists; no runtime gap |
| Scope field adoption | Single-project beta does not need scoping |
| Orchestrator template deduplication | Maintenance burden, no runtime impact |
| PreCompact custom_instructions | Rotation works; this improves post-compaction context |
| Archetype definitions for designer/analyst/engineer/meta | WARN-level validation runs; no silent failures |
| Multi-LLM support | Claude-only is correct for this stage |
| Token optimization features | Not a framework responsibility at this layer |

---

## 4. OKRs by Stage

### 4.1 Stage 1 OKRs: Internal Readiness (Now -- End of Q1 2026)

**O1: Achieve zero-friction internal onboarding**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR1.1 | Build success rate on clean checkout | 100% | CI green on every commit |
| KR1.2 | P0 issues open | 0 | Issue tracker |
| KR1.3 | Time to first productive session (new internal user) | < 30 minutes | Stopwatch measurement per user |
| KR1.4 | Internal users who complete a full /task cycle | >= 3 of 5-15 | Direct observation |

**O2: Make /consult the authoritative oracle**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR2.1 | /consult modes producing correct output | 5 of 5 | Manual test per mode |
| KR2.2 | Dead-reference errors in /consult responses | 0 | Query-and-check audit |
| KR2.3 | Skill and knowledge base cross-references that resolve | 100% | Glob + grep verification |
| KR2.4 | Internal user satisfaction with /consult (qualitative) | "Would recommend" from 80%+ | Direct feedback |

**O3: Stabilize the foundation for external distribution**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR3.1 | P1 issues resolved | 100% of in-scope items | Issue tracker |
| KR3.2 | Companion file autocomplete entries | <= 30 (from ~60) | `ls .claude/commands/ .claude/skills/` count |
| KR3.3 | Missing rules files created | 4 of 4 (inscription, hook, sails, usersync) | File existence check |
| KR3.4 | Internal champion identified | >= 1 | Visible advocacy behavior |

---

### 4.2 Stage 2 OKRs: Trusted External (Q2 -- Q3 2026)

**O4: Achieve frictionless external installation**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR4.1 | `brew install ari` success rate (macOS + Linux) | 100% | CI matrix + user reports |
| KR4.2 | `ari init` works without repo checkout | Yes | Automated test on clean machine |
| KR4.3 | Shell script dependencies in core flow | 0 | Dependency audit |
| KR4.4 | Time from `brew install` to first /consult response | < 5 minutes | User measurement |

**O5: Validate external product-market fit**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR5.1 | Active external users | >= 100 | Usage telemetry (opt-in) |
| KR5.2 | 30-day retention | >= 50% | Cohort analysis |
| KR5.3 | Community-created rites or agents | >= 3 | GitHub contributions |
| KR5.4 | Onboarding-related issue reports | < 50% of total issues | Issue label analysis |

**O6: Build community foundation**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR6.1 | GitHub stars | >= 500 | GitHub metric |
| KR6.2 | Documentation coverage (installation through custom rite) | 100% | Documentation audit |
| KR6.3 | Average issue response time | < 48 hours | Issue tracker SLA |
| KR6.4 | External contributors | >= 5 | GitHub contributor count |

---

### 4.3 Stage 3 OKRs: Broader Availability (Q4 2026 -- Q1 2027)

**O7: Launch enterprise tier**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR7.1 | Paying enterprise customers | >= 5 | Licensing system |
| KR7.2 | Enterprise tier MRR | >= $2,000 | Payment system |
| KR7.3 | Enterprise feature satisfaction | NPS >= 50 among enterprise users | Survey |
| KR7.4 | Pricing validated | Confirmed $15-25/seat/month | Design partner feedback |

**O8: Establish community standard**

| Key Result | Metric | Target | Measurement |
|------------|--------|--------|-------------|
| KR8.1 | Free user base | >= 5,000 | Telemetry |
| KR8.2 | Community rites in marketplace | >= 10 | Marketplace count |
| KR8.3 | GitHub stars | >= 2,000 | GitHub metric |
| KR8.4 | "Rails for Claude Code" search visibility | Top 3 result for "Claude Code framework" | Search audit |

---

## 5. Risk Register

### 5.1 Risk Matrix

| ID | Risk | Category | Probability | Impact | Severity | Mitigation | Owner |
|----|------|----------|-------------|--------|----------|------------|-------|
| R1 | Anthropic ships native project templates/framework features | Platform | 40% | High | **Critical** | Speed to market; establish community standard before native features mature; integrate with (not compete against) native features | Strategy |
| R2 | Agent Teams graduates to GA with workflow orchestration | Platform | 70% | Medium | **High** | Ensure rites compose with Agent Teams; position orchestration as workflow architecture, not just parallelism | Engineering |
| R3 | Build remains broken, blocking all downstream work | Technical | 10% | Critical | **High** | P0 priority, single point of failure; fix immediately (1-2 hours) | Engineering |
| R4 | /consult oracle quality insufficient for 30-minute onboarding | Technical | 30% | High | **High** | Dedicated Workstream 3; measurable acceptance criteria per mode; iterate based on internal user feedback | Engineering |
| R5 | Mythology learning curve alienates external adopters | Adoption | 35% | Medium | **Medium** | Plain-English glossary; /consult translates terminology; progressive disclosure; lead with function, introduce mythology as depth | Product |
| R6 | Setup friction (ari init, build from source) blocks external adoption | Adoption | 25% | High | **Medium** | Homebrew distribution; ari init standalone; one-command install target | Engineering |
| R7 | Shell-to-binary migration scope exceeds estimates (124 scripts, 39K LOC) | Technical | 40% | Medium | **Medium** | Prioritize critical-path scripts only; defer utility scripts; accept some shell dependencies for non-core flows | Engineering |
| R8 | Competitive window closes before reaching Stage 2 | Market | 20% | High | **Medium** | Begin architecture documentation and "Rails for Claude Code" positioning publicly during Stage 1; create mindshare before code is public | Strategy |
| R9 | Single-platform dependency (Claude Code only) limits addressable market | Market | 15% | Medium | **Low-Medium** | Accept for now; platform-specific frameworks succeed by being excellent on one platform; consider AGENTS.md output support in Stage 3 | Strategy |
| R10 | AGENTS.md becomes universal standard, CC adopts it | Market | 35% | Medium | **Medium** | Add AGENTS.md output support to materialization pipeline in Stage 3; Knossos value is in the pipeline, not the file format | Engineering |
| R11 | Internal team too small to generate meaningful adoption data | Adoption | 30% | Low | **Low** | Supplement with trusted external beta users earlier than planned if internal data is insufficient | Product |

### 5.2 Risk Response Triggers

| Trigger Signal | Action |
|----------------|--------|
| Anthropic announces project templates feature | Assess overlap; accelerate Stage 2 launch; position Knossos as the deep framework that works with native templates |
| Agent Teams GA with workflow features | Verify rite compatibility with Agent Teams; update documentation to show complementary use |
| claude-flow adds configuration management | Emphasize materialization depth and idempotency; publish comparison content |
| 30-day retention below 30% in Stage 2 | Halt Stage 3 planning; conduct user interviews; identify and fix top 3 friction points |
| Bear case revenue triggers (< $2K MRR at 6 months into Stage 3) | Evaluate professional services pivot; context-engineering consulting at $200-500/hr |

---

## 6. Resource Allocation

### 6.1 Investment by Stage

| Stage | Engineering Hours | Dollar Equivalent ($200/hr) | Calendar Time | Key Constraint |
|-------|-------------------|----------------------------|---------------|----------------|
| Stage 1: Internal | 38-75 hrs | $7.6K-$15K | 6-10 weeks | Quality bar (30-minute onboarding) |
| Stage 2: Trusted External | 142-284 hrs | $28K-$57K | 8-14 weeks | ari init standalone (single biggest blocker) |
| Stage 3: Broader Availability | 224-448 hrs | $45K-$90K | 12-20 weeks | Enterprise feature quality + community growth |
| **Total to Revenue** | **404-807 hrs** | **$81K-$162K** | **26-44 weeks** | |

### 6.2 Stage 1 Sprint Structure

Recommended: 2-week sprints, 20-30 hours/sprint of focused Knossos work.

| Sprint | Weeks | Focus | Hours | Deliverables |
|--------|-------|-------|-------|-------------|
| Sprint 1 | 1-2 | Fix Foundations | 10-15 | All P0 issues resolved, build green |
| Sprint 2 | 3-4 | Strengthen Systems (batch 1) | 15-20 | S-01 through S-06 complete |
| Sprint 3 | 5-6 | Strengthen Systems (batch 2) + Pythia start | 15-20 | S-07 through S-10, P-01 through P-03 |
| Sprint 4 | 7-8 | Pythia world-class + Fates review | 15-20 | P-04 through P-07, MF-01 through MF-05 |
| Sprint 5 | 9-10 | Internal validation + Stage 2 prep | 10-15 | M1.6 validation, D-01 kickoff |

### 6.3 Stage 2 Sprint Structure

Recommended: 2-week sprints, 30-40 hours/sprint (increased intensity during distribution window).

| Sprint | Weeks | Focus | Hours | Deliverables |
|--------|-------|-------|-------|-------------|
| Sprint 6 | 11-12 | ari init standalone (investigation + design) | 30-40 | D-02 architecture, D-04 shell migration starts |
| Sprint 7 | 13-14 | ari init standalone (implementation) | 30-40 | D-02 implementation, D-01 GoReleaser |
| Sprint 8 | 15-16 | Distribution pipeline + shell migration | 30-40 | D-02 complete, D-03 cross-platform, D-04 critical paths |
| Sprint 9 | 17-18 | Documentation + community setup | 25-35 | D-06 through D-10 |
| Sprint 10-13 | 19-26 | External beta + iteration | 20-30/sprint | M2.6 external validation |

### 6.4 Minimum Viable Team

**Stage 1**: 1 senior engineer (part-time, 20-30 hrs/week). The stakeholder and primary developer are the same person. No additional team needed.

**Stage 2**: 1 senior engineer (full-time equivalent) + 0.5 community/developer relations. The transition to external distribution requires documentation, issue triage, and community management that pulls from engineering time.

**Stage 3**: 1.5-2 senior engineers + 0.5 community/developer relations + 0.25 product/strategy. Enterprise features (licensing, gating, reporting) and community ecosystem management require dedicated capacity.

### 6.5 Resource Allocation Across Bets

Using the three-horizon model:

| Horizon | Description | Allocation | Knossos Initiatives |
|---------|-------------|------------|---------------------|
| H1: Core business (now) | Fix and ship what exists | 70% | P0/P1 fixes, /consult quality, internal validation |
| H2: Emerging opportunities (6-12 months) | Distribution and community | 25% | ari init, Homebrew, documentation, community infrastructure |
| H3: Future bets (12+ months) | Enterprise revenue, ecosystem | 5% | Enterprise design partner conversations, rite marketplace design |

**Rationale**: The product is 95% built. The overwhelming priority is finishing the last 5% (H1) and packaging it for distribution (H2). Enterprise revenue (H3) gets minimal investment now because it depends entirely on H1 and H2 succeeding first.

---

## 7. Go/No-Go Decision Framework

### 7.1 Decision Architecture

```
                          Stage 1
                     (Internal, Now)
                           |
                      [Gate 1]
                     /         \
                GO              NO-GO
                |                  |
           Stage 2           Fix & Retry
        (Trusted External)    (max 2 cycles)
                |                  |
           [Gate 2]          [Kill Criteria]
          /         \              |
       GO            NO-GO     Pivot to
       |                |      Consulting
   Stage 3          Fix & Retry
  (Broader)         (max 2 cycles)
       |                |
   [Gate 3]       [Kill Criteria]
  /         \           |
GO           NO-GO   Niche Product
|                |   (no enterprise)
Enterprise    Reassess
Revenue       Model
```

### 7.2 Kill Criteria (When to Pivot)

These are the bear-case triggers from the business model. If any is sustained for 2+ sprints after corrective action:

| Kill Criterion | Signal | Pivot Direction |
|----------------|--------|-----------------|
| Internal users prefer manual CLAUDE.md | 50%+ of internal users disengage after initial trial | Scrap framework approach; focus on individual tools (oracle-only, materialization-only) |
| Onboarding time cannot be reduced below 45 minutes | After 3 iterations of onboarding flow | Simplify radically; strip mythology from user-facing surfaces; reduce to "ari init + ari consult" |
| Anthropic ships full native framework | Project templates + session management + agent composition native in CC | Pivot to consulting; accumulated expertise in context engineering is the asset, not the framework |
| Stage 2 retention below 30% for 3 months | Users try and leave; product does not stick | Diagnose root cause; if architectural, consider major pivot; if friction, iterate on UX |
| Zero enterprise willingness-to-pay after 10+ conversations | Design partners uniformly say "not worth paying for" | Remain free/open-source; monetize through adjacent services or sponsorship |

### 7.3 Decision Cadence

| Frequency | Decision | Who Decides |
|-----------|----------|-------------|
| Weekly | Sprint priority adjustments within current workstream | Engineering (autonomous) |
| Bi-weekly | Sprint review: on track for current milestone? | Engineering + Stakeholder |
| Monthly | Workstream progress: adjust timeline or scope? | Stakeholder |
| Per-gate | Go/no-go: advance to next stage? | Stakeholder (final decision) |
| Quarterly | Strategic review: competitive landscape changed? Pivot needed? | Stakeholder + Strategy |

---

## 8. Execution Timeline Summary

```
2026
Feb         Mar         Apr         May         Jun         Jul         Aug         Sep
|-----------|-----------|-----------|-----------|-----------|-----------|-----------|
|  STAGE 1: INTERNAL                |  STAGE 2: TRUSTED EXTERNAL                   |
|                                   |                                               |
|  W1: Fix    W2: Strengthen   W3: Pythia                                          |
|  Fndns      Systems          World-class                                         |
|  [2w]       [4w]             [4w]                                                |
|                        W4: Fates                                                 |
|                        Review                                                    |
|                        [2w]                                                      |
|                              W5: Distribution Packaging                          |
|                              [8w]                                                |
|                                   |                                               |
|        [Gate 1] ------------------>                                              |
|                                   |                    [Gate 2] ----------------->|
|                                                                                   |
Oct         Nov         Dec         2027 Jan    Feb
|-----------|-----------|-----------|-----------|
|  STAGE 3: BROADER AVAILABILITY                |
|                                               |
|  Enterprise Features    Rite Marketplace      |
|  [4w]                   [4w]                  |
|  Licensing              Revenue Generation    |
|  [2w]                   [4w]                  |
|  Doc Site                                     |
|  [2w]                                         |
|                                               |
|                         [Gate 3] ------------>|
```

---

## 9. Success Criteria Summary

### The Acid Test for Each Stage

**Stage 1**: "A new internal team member clones the repo, builds ari, runs materialize, starts a session, uses /consult to understand the framework, and completes a /task cycle -- all within 30 minutes, without help."

**Stage 2**: "A Claude Code power user who has never seen Knossos runs `brew install ari && ari init`, uses /consult to understand what they just installed, activates a rite, and says 'I can see how this is better than my manual CLAUDE.md' -- all within their first session."

**Stage 3**: "An engineering manager evaluates Knossos for their team of 10 developers, sees that it solves configuration drift, onboarding friction, and workflow consistency, and is willing to pay $20/seat/month for the enterprise features -- within a 1-hour evaluation."

### Leading Indicators to Watch

| Indicator | Healthy Signal | Warning Signal |
|-----------|---------------|----------------|
| /consult query accuracy | 90%+ correct routing | < 80% correct routing |
| Internal user engagement | Weekly active usage by 80%+ of team | < 50% weekly active |
| Time-to-first-value | Decreasing with each new user | Flat or increasing |
| Issue volume vs. feature requests | Feature requests > bug reports | Bug reports > feature requests |
| Community contribution | External PRs and rite submissions | Zero external contributions after 3 months |
| Competitive movement | Competitors copying Knossos patterns | Knossos scrambling to copy competitors |

---

## 10. Appendix

### 10.1 Upstream Artifact Links

| Artifact | Path | Key Findings |
|----------|------|-------------|
| Market Research | [MARKET-distribution-readiness.md](./MARKET-distribution-readiness.md) | TAM: $1.4-7.0B; SAM: $0-21M ARR; SOM Year 1: 160-715 users; optimal launch window Q2-Q3 2026; "Rails for X" patterns apply |
| Competitive Analysis | [COMPETITIVE-distribution-readiness.md](./COMPETITIVE-distribution-readiness.md) | No framework competitor exists; claude-flow (13.8K stars) is orchestration, not framework; 12-month competitive window; platform risk HIGH but MANAGEABLE |
| Business Model | [BUSINESS-MODEL-distribution-readiness.md](./BUSINESS-MODEL-distribution-readiness.md) | Internal: 38-75 hrs; External: +142-284 hrs; Open Core model; $20/seat/month enterprise; Expected Year 2 ARR: $402K (probability-weighted) |
| Stakeholder Preferences | [STAKEHOLDER-PREFERENCES-distribution-readiness.md](../STAKEHOLDER-PREFERENCES-distribution-readiness.md) | Progressive readiness bar; 4-step work sequence; Pythia is flagship; mythology is load-bearing; 30-minute onboarding target |

### 10.2 Key Assumptions

| Assumption | Source | Sensitivity |
|------------|--------|-------------|
| 12-month competitive window | Competitive analysis (platform risk assessment) | High -- if Anthropic accelerates native features, window could be 6 months |
| 30-minute onboarding is achievable | Stakeholder target; market research (developer tool adoption patterns) | High -- if product complexity prevents this, entire strategy needs revision |
| Open Core model is correct revenue structure | Business model analysis (comparable tool benchmarks) | Medium -- validated by Hashicorp, GitLab patterns but unproven for Knossos specifically |
| Internal team of 5-15 is sufficient for Stage 1 validation | Stakeholder input | Low -- small sample but high context; supplemented by Stage 2 external data |
| Shell-to-binary migration is tractable in 40-80 hours | Engineering estimate based on 124 scripts / 39K LOC | Medium -- actual scope may vary; prioritize critical-path scripts |
| Enterprise willingness-to-pay exists at $15-25/seat/month | Business model pricing benchmarks | High -- completely unvalidated; design partner conversations needed in Stage 2 |

### 10.3 Glossary of Stakeholder-Specified Terms

For roadmap readers unfamiliar with Knossos terminology:

| Term | Meaning |
|------|---------|
| Rite | A switchable workflow template that composes agents, skills, hooks, and commands for a specific project type (e.g., 10x-dev, strategy, security) |
| Mena | The source system for commands and skills. Contains dromena (commands) and legomena (skills) |
| Dromena | User-invoked slash commands. Transient -- execute and exit |
| Legomena | Model-invoked skills. Persistent -- stay in context once loaded |
| Materialization | The pipeline that transforms source definitions into .claude/ projections |
| /consult (Pythia) | The oracle command. Meta-level advisor that routes questions to the right context |
| Moirai (Fates) | The agent responsible for session state mutations. Three aspects: Clotho (creation), Lachesis (measurement), Atropos (termination) |
| Ariadne (ari) | The CLI binary. Faithful executor -- deterministic, authoritative for state changes |
| Clew | The audit trail. Append-only JSONL event logging |
| White Sails | Confidence signaling. Three states: WHITE (high confidence), GRAY (uncertain), BLACK (low confidence) |
| Inscription | CLAUDE.md generation. The "labyrinth speaking at entry" |
| Satellite regions | User-owned content in .claude/ files that is preserved during materialization |

---

## Attestation

| Check | Status | Evidence |
|-------|--------|----------|
| All upstream artifacts read and synthesized | Complete | Market research, competitive analysis, business model, stakeholder preferences |
| Initiatives evaluated consistently with RICE framework | Complete | Section 3: Prioritization Matrix |
| Prioritization documented with rationale for top/cut decisions | Complete | Sections 3.1 and 3.2 |
| Resources allocated within capacity constraints | Complete | Section 6: stage-by-stage with sprint structure |
| Timeline and milestones defined with dependencies | Complete | Section 1 (milestones), Section 2 (workstreams with dependency graph) |
| OKRs created with measurable key results | Complete | Section 4: 8 objectives, 28 key results |
| Risk register with mitigations | Complete | Section 5: 11 risks with severity, probability, and mitigation |
| Go/no-go decision framework | Complete | Section 7: gates, kill criteria, decision cadence |
| Stakeholder work sequence honored | Complete | Workstreams 1-4 map directly to stakeholder's 4-step sequence |
| Out-of-scope items documented | Complete | Section 3.2 |

| Artifact | Absolute Path |
|----------|---------------|
| This roadmap | `/Users/tomtenuta/Code/knossos/docs/strategy/ROADMAP-distribution-readiness.md` |
| Market research | `/Users/tomtenuta/Code/knossos/docs/strategy/MARKET-distribution-readiness.md` |
| Competitive analysis | `/Users/tomtenuta/Code/knossos/docs/strategy/COMPETITIVE-distribution-readiness.md` |
| Business model | `/Users/tomtenuta/Code/knossos/docs/strategy/BUSINESS-MODEL-distribution-readiness.md` |
| Stakeholder preferences | `/Users/tomtenuta/Code/knossos/docs/STAKEHOLDER-PREFERENCES-distribution-readiness.md` |

---

*Produced 2026-02-08. This roadmap should be reviewed at each stage gate and updated quarterly against competitive landscape changes. The timeline is a planning instrument, not a commitment -- quality gates take precedence over calendar dates.*

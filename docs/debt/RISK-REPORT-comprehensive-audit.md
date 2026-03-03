---
audit_id: session-20260302-232344-1b73b3a8
assessment_date: "2026-03-02"
radar_integrated: true
radar_date: "2026-03-03"
total_items: 81
tier_breakdown:
  critical: 4
  high: 4
  medium: 30
  low: 40
  resolved: 3
quick_wins: 11
items_requiring_user_input: 2
---

# Risk Assessment Report

## Executive Summary

81 debt items scored across 6 dimensions (blast radius, silent failure risk, compounding rate, blocking factor, fix effort, regression risk). Composite scores range from 6 (minimal) to 25 (critical). Includes 8 radar-sourced items (DEBT-173 through DEBT-180) from the Knowledge Radar scan of 2026-03-03, plus rescoring of 2 existing items where radar evidence changed the assessment.

**Tier distribution**: 4 Critical, 4 High, 30 Medium, 40 Low, 3 Resolved (81 total).

**Top 5 highest-risk items**:

1. **DEBT-138** (Composite 25, Critical): 30+ error-discarding patterns, 16 checksum sites in userscope that can silently produce corrupt sync state. Radar OPP-004 reinforces: "exit-0 must mean complete output" invariant is systematically violated.
2. **DEBT-175** (Composite 23, Critical): **NEW from radar.** 59 non-atomic `os.WriteFile` sites on state files read every ari invocation. Crash during write corrupts rite state, worktree registry, or manifest with no recovery path.
3. **DEBT-112** (Composite 23, Critical): Userscope 23.7% coverage on the user-global state manager. Manages `~/.claude/` with 76% of code untested. Compounding cluster anchor.
4. **DEBT-131** (Composite 22, Critical): 9 SCARs without regression tests. Radar OPP-005 adds CC integration smoke-test angle -- 3 of 4 integration-failure SCARs lack automated regression.
5. **DEBT-173** (Composite 21, High): **NEW from radar.** 41 os.Stdout bypass sites break `--output=json`. Agent validate creates Printer then ignores it.

**Key compounding clusters**:
- **Userscope cluster** (DEBT-112 + DEBT-138 + DEBT-158 + DEBT-171): Four items in the same subsystem that multiply each other's risk. Low coverage means silent failures go undetected, coupling means fixes have wide blast radius, the post-split monolith means the attack surface is large.
- **Materialize pipeline convergent hotspot** (DEBT-138 + DEBT-175 + DEBT-143 + DEBT-173 + DEBT-174): Three independent radar signals (unguarded scars, convention drift, architecture decay) converge on the materialize pipeline. See "Convergent Hotspot" section.
- **Systemic SCAR patterns** (DEBT-131 + DEBT-177 + DEBT-178 + DEBT-179): Radar identifies 3 distinct failure categories across 12 SCARs -- silent failure, schema evolution, and boundary violations. These are structural tendencies, not individual bugs.
- **Documentation accuracy cluster** (DEBT-115 + DEBT-137 + DEBT-163 + DEBT-145 + DEBT-180): Stale .know/ files feed incorrect intelligence to agents. Radar OPP-009/010 confirm both architecture.md and design-constraints.md at 0.78 confidence with 4 resolved items unmarked.

**Quick wins** (11 items, fix effort = 1, composite >= 9): DEBT-115, DEBT-116, DEBT-120, DEBT-137, DEBT-140, DEBT-143, DEBT-145, DEBT-157, DEBT-176, DEBT-179, DEBT-180. Total estimated effort: ~5 hours for all 11 combined.

---

## Scoring Matrix

Dimensions: **Bl** = Blast Radius, **Si** = Silent Failure Risk, **Co** = Compounding Rate, **Bk** = Blocking Factor, **Ef** = Fix Effort, **Re** = Regression Risk, **C** = Composite (sum of all 6).

| DEBT-ID | Title | Bl | Si | Co | Bk | Ef | Re | C | Tier |
|---------|-------|----|----|----|----|----|----|----|------|
| 100 | output pkg 11.7% coverage | 4 | 3 | 3 | 3 | 4 | 4 | 21 | High |
| 101 | lint pkg 12.9% coverage | 3 | 2 | 2 | 2 | 3 | 2 | 14 | Medium |
| 102 | org cmd 1.2% coverage | 1 | 1 | 1 | 1 | 2 | 1 | 7 | Low |
| 103 | explain cmd 42.6% coverage | 2 | 2 | 1 | 1 | 2 | 1 | 9 | Low |
| 104 | artifact cmd 0% CLI coverage | 2 | 1 | 2 | 1 | 3 | 1 | 10 | Low |
| 105 | common cmd 0% coverage | 2 | 1 | 2 | 1 | 1 | 1 | 8 | Low |
| 106 | inscription cmd 0% CLI coverage | 4 | 3 | 2 | 2 | 3 | 2 | 16 | Medium |
| 107 | manifest cmd 0% CLI coverage | 2 | 3 | 2 | 1 | 2 | 2 | 12 | Medium |
| 108 | provenance cmd 0% CLI coverage | 2 | 2 | 2 | 1 | 2 | 2 | 11 | Low |
| 109 | tribute cmd 0% CLI coverage | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 110 | naxos cmd 0% CLI coverage | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 111 | root cmd 0% coverage | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 112 | userscope 23.7% coverage | 5 | 4 | 4 | 3 | 4 | 3 | 23 | Critical |
| 113 | config pkg 34.6% coverage | 3 | 3 | 2 | 2 | 2 | 2 | 14 | Medium |
| 114 | 502 functions at 0% (aggregate) | 4 | 3 | 3 | 2 | 5 | 3 | 20 | High |
| 115 | .know/test-coverage.md stale | 3 | 3 | 3 | 2 | 1 | 1 | 13 | Medium |
| 116 | context-injection.sh dead code | 2 | 1 | 2 | 2 | 1 | 1 | 9 | Low |
| 117 | validation.sh shell in Go codebase | 2 | 2 | 1 | 1 | 3 | 2 | 11 | Low |
| 118 | e2e-validate.sh CI-only | 1 | 1 | 1 | 1 | 4 | 2 | 10 | Low |
| 119 | Shell script 565-line footprint (aggregate) | 2 | 1 | 1 | 1 | 3 | 2 | 10 | Low |
| 120 | 37 doc refs to dead shell chain | 2 | 2 | 2 | 1 | 1 | 1 | 9 | Low |
| 121 | Resume cross-rite deferred | 2 | 1 | 1 | 1 | 4 | 2 | 11 | Low |
| 122 | arch-ref skill missing | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 123 | 10x-dev "ghost skills" not ghosts | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 124 | ADR-0028 unwritten | 2 | 1 | 3 | 2 | 3 | 1 | 12 | Medium |
| 125 | state.json last_sync dead write | 2 | 1 | 1 | 2 | 2 | 2 | 10 | Low |
| 126 | 3-version event bridge | 3 | 2 | 2 | 3 | 2 | 3 | 15 | Medium |
| 127 | Shell cleanse partially complete | 2 | 1 | 1 | 1 | 3 | 1 | 9 | Low |
| 128 | Hook bash elimination nearly done | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 129 | Single-binary scope unclear | 2 | 1 | 2 | 2 | 5 | 2 | 14 | Medium |
| 130 | state.json full elimination | 2 | 1 | 1 | 2 | 2 | 2 | 10 | Low |
| 131 | 9 SCARs lack regression tests | 4 | 4 | 4 | 3 | 4 | 3 | 22 | Critical |
| 132 | rite pkg 38 funcs at 0% | 3 | 2 | 2 | 2 | 3 | 2 | 14 | Medium |
| 133 | clewcontract 26 funcs at 0% | 2 | 1 | 2 | 1 | 2 | 1 | 9 | Low |
| 134 | cmd/session 25 funcs at 0% | 2 | 2 | 2 | 1 | 3 | 2 | 12 | Medium |
| 135 | manifest pkg 24 funcs at 0% | 3 | 2 | 2 | 2 | 3 | 2 | 14 | Medium |
| 136 | cmd/sails 39.5% coverage | 2 | 2 | 2 | 1 | 2 | 2 | 11 | Low |
| 137 | MEMORY.md ghost skills inaccurate | 2 | 2 | 3 | 1 | 1 | 1 | 10 | Low |
| 138 | 30+ error-discard patterns | 5 | 5 | 4 | 3 | 4 | 4 | 25 | Critical |
| 139 | Zero log.Debug infrastructure | 3 | 3 | 3 | 2 | 3 | 1 | 15 | Medium |
| 140 | extractEmbeddedMena 6 silent errors | 3 | 4 | 2 | 1 | 1 | 1 | 12 | Medium |
| 141 | Hook handlers graceful degradation (observation) | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 142 | Hook embedded FS wiring boilerplate | 2 | 3 | 2 | 1 | 2 | 2 | 12 | Medium |
| 143 | engine.go stale-entry error discard | 3 | 4 | 2 | 1 | 1 | 2 | 13 | Medium |
| 144 | log.Printf warnings go to void | 3 | 3 | 3 | 2 | 3 | 2 | 16 | Medium |
| 145 | 3 RISK items now resolved | 2 | 2 | 2 | 1 | 1 | 1 | 9 | Resolved |
| 146 | Three copyDir implementations | 3 | 2 | 2 | 1 | 2 | 2 | 12 | Medium |
| 147 | Dual source chain construction | 3 | 3 | 3 | 2 | 2 | 3 | 16 | Medium |
| 148 | TENSION-006 shared manifest (resolved) | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Resolved |
| 149 | 2-3 hooks without withTimeout | 4 | 4 | 2 | 3 | 3 | 5 | 21 | High |
| 150 | Session test setup duplication | 2 | 1 | 3 | 2 | 2 | 1 | 11 | Low |
| 151 | Platform mena resolution dual paths | 2 | 2 | 2 | 1 | 1 | 1 | 9 | Low |
| 152 | copyDirWithStripping unified (resolved) | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Resolved |
| 153 | Dual OwnerType definitions | 2 | 1 | 2 | 2 | 3 | 3 | 13 | Medium |
| 154 | McpServerConfig naming | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 155 | Deprecated Commands/Skills fields | 2 | 1 | 2 | 2 | 2 | 3 | 12 | Medium |
| 156 | SourceType convention stable (observation) | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 157 | registry not a leaf package (docs) | 2 | 1 | 2 | 2 | 1 | 1 | 9 | Low |
| 158 | userscope imports mena (coupling) | 4 | 2 | 3 | 3 | 4 | 3 | 19 | Medium |
| 159 | naxos Layer 3 imports Layer 2 | 3 | 1 | 3 | 2 | 2 | 2 | 13 | Medium |
| 160 | tribute Layer 3 imports Layer 2 | 3 | 1 | 3 | 2 | 1 | 1 | 11 | Low |
| 161 | output has zero internal imports (observation) | 1 | 1 | 1 | 1 | 1 | 1 | 6 | Low |
| 162 | worktree imports materialize | 3 | 1 | 2 | 2 | 3 | 3 | 14 | Medium |
| 163 | architecture.md 3 stale claims | 3 | 2 | 3 | 2 | 1 | 1 | 12 | Medium |
| 164 | lint.go 784 lines, 5 domains | 2 | 1 | 3 | 3 | 2 | 1 | 12 | Medium |
| 165 | output.go + rite.go 1,477 lines | 2 | 1 | 3 | 2 | 3 | 1 | 12 | Medium |
| 166 | inscription 3 files over 500 lines | 2 | 1 | 2 | 1 | 2 | 2 | 10 | Low |
| 167 | worktree operations.go 707 lines | 2 | 1 | 2 | 1 | 2 | 2 | 10 | Low |
| 168 | sails generator.go 678 lines | 1 | 1 | 2 | 1 | 1 | 1 | 7 | Low |
| 169 | clewcontract event.go 644 lines | 1 | 1 | 3 | 1 | 2 | 1 | 9 | Low |
| 170 | materialize.go 53% extracted | 2 | 1 | 2 | 2 | 2 | 2 | 11 | Low |
| 171 | sync_mena.go 654 lines post-split | 4 | 3 | 3 | 2 | 3 | 3 | 18 | Medium |
| 172 | writeguard.go 588 lines | 1 | 1 | 2 | 1 | 2 | 1 | 8 | Low |
| 173 | os.Stdout bypass 41 sites (OPP-001) | 4 | 4 | 3 | 2 | 4 | 4 | 21 | High |
| 174 | fmt.Errorf at CLI boundaries 39 sites (OPP-002) | 3 | 3 | 3 | 2 | 3 | 2 | 16 | Medium |
| 175 | Non-atomic writes on state files 59 sites (OPP-007) | 4 | 5 | 3 | 3 | 3 | 5 | 23 | Critical |
| 176 | KnossosHome cache poisoning in tests (OPP-011) | 2 | 3 | 2 | 1 | 1 | 1 | 10 | Low |
| 177 | Schema evolution systemic pattern (OPP-012) | 3 | 3 | 3 | 2 | 2 | 2 | 15 | Medium |
| 178 | Data corruption systemic pattern (OPP-013) | 3 | 4 | 3 | 2 | 3 | 2 | 17 | Medium |
| 179 | Historical boundary violations (OPP-014) | 2 | 2 | 3 | 2 | 1 | 1 | 11 | Low |
| 180 | Testify drift 23 vs documented 18 (OPP-018) | 1 | 1 | 2 | 1 | 1 | 1 | 7 | Low |

---

## Tier: Critical

### DEBT-138: 30+ error-discard patterns (Composite: 25)

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 5 | Spans 25+ files, 16 sites in userscope sync affect `~/.claude/` global state |
| Silent failure | 5 | By definition -- blank identifier means exit 0 with potentially corrupt state. Checksum zero-value on I/O error causes "no change" decision on corrupted files |
| Compounding | 4 | Pattern is self-reinforcing: new code copies the `_ :=` idiom from adjacent lines. 97 total sites growing |
| Blocking | 3 | Blocks confidence in userscope sync correctness, blocks observability improvements |
| Effort | 4 | 2-3 days for the 19 correctness+integrity sites; full remediation larger |
| Regression | 4 | Changing error handling on load-bearing sync paths can alter control flow |

**Trigger scenario**: Filesystem permission error on `~/.claude/` directory (e.g., antivirus lock, disk full, macOS sandbox). `checksum.File()` returns zero value, sync compares zero-vs-stored, decides "unchanged", skips update. User's mena/agent files silently stale.

**Recommendation**: Prioritize the 16 checksum-discard sites in `userscope/` first. These are the correctness-risk subset. The 33+ intentional best-effort sites (cleanup, Cobra) can remain.

---

### DEBT-112: userscope 23.7% coverage (Composite: 23)

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 5 | Manages user-global `~/.claude/` state across all projects. Breakage affects every CC session |
| Silent failure | 4 | Untested paths include mena sync, agent sync, cleanup -- all filesystem operations that can silently produce wrong results |
| Compounding | 4 | Every new sync feature adds to the untested surface. Post-split it is 2,716 lines across 7 files |
| Blocking | 3 | Blocks safe refactoring of the userscope cluster (DEBT-158, DEBT-171) |
| Effort | 4 | 3-4 days for meaningful coverage on a 2,716-line subsystem with filesystem dependencies |
| Regression | 3 | Tests themselves are safe (additive), but discovering bugs during test writing may require fixes |

**Trigger scenario**: Any sync path exercising the untested 76.3% of code. Most likely: new rite with unusual mena structure, edge case in agent collision detection, or embedded FS path divergence from filesystem path.

**Recommendation**: Write tests targeting the sync_mena.go paths first (largest file, most error discards). Combine with DEBT-138 remediation for maximum cluster risk reduction.

---

### DEBT-131: 9 SCARs lack regression tests (Composite: 22)

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 4 | SCARs are documented past failures. SCAR-004 (silent provenance error) and SCAR-023 (template path) affect pipeline correctness |
| Silent failure | 4 | SCAR-004 is literally "silent error discard." SCAR-015 (stdout pollution) produces wrong hook output silently |
| Compounding | 4 | Without regression tests, each refactoring round has a chance of re-introducing the original failure. Risk grows with code velocity |
| Blocking | 3 | Blocks confident refactoring of affected areas |
| Effort | 4 | 3-5 days for all 9; 1 day for the 2 highest risk (SCAR-004, SCAR-023) |
| Regression | 3 | Tests are additive, but may expose latent regressions already present |

**Trigger scenario**: Refactoring `materialize.go` or `provenance.go` without a regression test for SCAR-004. The silent-discard pattern reappears because the developer does not know the history.

**Recommendation**: Write regression tests for SCAR-004 and SCAR-023 first (behavioral guards without structural protection). The remaining 7 SCARs have structural fixes (code changes that make reversion harder) and are lower priority.

---

### DEBT-175: Non-atomic writes on critical state files (Composite: 23) -- NEW from Radar OPP-007

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 4 | `rite/state.go` and `worktree/metadata.go` are read on every `ari` invocation. Corruption breaks all ari commands |
| Silent failure | 5 | Crash during write leaves truncated YAML. Next invocation fails with a parse error, but the write itself produces no error |
| Compounding | 3 | 59 sites vs 13 correct sites. Every new state file write defaults to `os.WriteFile` unless developer knows to use `AtomicWriteFile` |
| Blocking | 3 | Blocks reliability claims for ari state management |
| Effort | 3 | 1 day for the 4 highest-risk files; mechanical `os.WriteFile` -> `fileutil.AtomicWriteFile` replacement |
| Regression | 5 | Atomic writes use temp-file-then-rename. On some filesystems (network mounts, Docker volumes), rename semantics differ. Must test on target platforms |

**Trigger scenario**: Power loss or `kill -9` during `ari sync`. `rite/state.go:93` has written 50% of the invocation state YAML. Next `ari` invocation parses the truncated file, gets a YAML error, and all commands fail until the user manually deletes the corrupt file.

**Recommendation**: Replace `os.WriteFile` with `fileutil.AtomicWriteFile` in `rite/state.go:93` and `worktree/metadata.go:78,90,256` first (highest-risk state files). Then tackle `manifest/manifest.go:191` and `artifact/registry.go:173,247`. The remaining 50+ sites can be addressed incrementally.

---

## Tier: High

### DEBT-100: output package 11.7% coverage (Composite: 21)

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 4 | 47 functions at 0% across the CLI output contract. Text formatting, rite output, manifest output all untested. User-visible breakage |
| Silent failure | 3 | Output bugs are usually visible (garbled text) but some types (JSON output, machine-readable formats) can produce subtly wrong data |
| Compounding | 3 | Every new command adds output types to untested files. Pattern is self-reinforcing |
| Blocking | 3 | Blocks the output.go/rite.go monolith split (DEBT-165) -- cannot split safely without test coverage |
| Effort | 4 | 2-3 days for the 47 functions across 3 files (781 + 696 + 240 lines) |
| Regression | 4 | Output is the user-visible contract. Breakage in formatting is immediately noticed but hard to catch in CI without tests |

**Trigger scenario**: Refactoring output types during DEBT-165 monolith split introduces a formatting regression in `ari session list` or `ari rite list` JSON output. No test catches it.

**Recommendation**: Prioritize JSON output paths (machine-readable contract) over text formatting (human-readable, more tolerant of minor changes).

---

### DEBT-149: 2-3 hooks without withTimeout (Composite: 21)

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 4 | `cheapo_revert` runs full `m.Sync()` -- a slow sync blocks CC indefinitely. User's entire session hangs |
| Silent failure | 4 | No timeout means no error -- the hook just runs forever. User perceives CC as "frozen" with no diagnostic |
| Compounding | 2 | Stable -- new hooks follow the `withTimeout` pattern. Only the 2-3 outliers remain |
| Blocking | 3 | Blocks hook handler consistency cleanup |
| Effort | 3 | 2-4 hours to refactor the outlier hooks |
| Regression | 5 | Adding timeout to a sync operation can cause partial-completion scenarios. Must handle gracefully |

**Trigger scenario**: Large rite with many mena files, `cheapo_revert` hook fires, `m.Sync()` takes >30s on a slow disk. CC appears frozen. User force-kills CC, leaving `.claude/` in a partially-written state.

**Recommendation**: Add `withTimeout` to `cheapo_revert` and `worktreeremove` as highest priority. Test that timeout produces a clean no-op rather than partial state.

---

### DEBT-114: 502 functions at 0% (Composite: 20)

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 4 | 28.6% of all functions completely untested. Aggregate risk across entire codebase |
| Silent failure | 3 | Many zero-coverage functions are in output/formatting (visible failure) but some are in sync paths (silent) |
| Compounding | 3 | Number grows as new functions are added without tests |
| Blocking | 2 | Aggregate metric -- individual items are more actionable |
| Effort | 5 | Cannot be addressed as a single item -- requires sustained investment across many packages |
| Regression | 3 | Tests are additive, low regression risk from the testing itself |

**Note**: This is an aggregate item. Remediation happens through individual package items (DEBT-100, DEBT-112, etc.). Scored here for portfolio visibility.

---

### DEBT-173: os.Stdout bypass in 41 cmd/ sites (Composite: 21) -- NEW from Radar OPP-001

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Blast radius | 4 | 41 sites across agent, session, and other cmd/ packages. `--output=json` silently omits results for affected commands |
| Silent failure | 4 | The defining characteristic: `ari agent validate --output=json` returns incomplete output with no error. User automation consuming JSON gets partial data |
| Compounding | 3 | Every new cmd/ function that uses `fmt.Fprintf(os.Stdout)` instead of `printer.Print()` extends the gap |
| Blocking | 2 | Does not block other work directly, but blocks JSON output reliability claims |
| Effort | 4 | 2-3 days (41 sites, each needs structured output type routed through Printer) |
| Regression | 4 | Changing output routing can break text formatting for users who rely on current stdout patterns |

**Trigger scenario**: CI script runs `ari agent validate --output=json` and pipes to `jq`. Validation results are written to os.Stdout instead of the Printer, so they appear as raw text mixed into JSON. `jq` chokes on the mixed output.

**Recommendation**: Fix `agent/validate.go` first (creates Printer then ignores it -- most egregious). Then `session/gc.go` (never creates Printer). The remaining sites are lower severity.

---

## Tier: Medium

### Userscope Cluster (Medium tier, elevated by cluster interaction)

**DEBT-158**: userscope imports mena (Composite: 19). 18 call sites across sync_mena.go create tight coupling between sub-packages. Part of the userscope cluster -- fixing this enables safer work on DEBT-171 and DEBT-112. Effort: 1-2 days.

**DEBT-171**: sync_mena.go 654 lines post-split (Composite: 18). Two parallel sync paths (filesystem vs embedded) for user-scoped mena. Divergence causes inconsistent state. Follows the proven copyDirFS unification pattern (DEBT-152, resolved). Effort: 1-2 days.

### Observability & Duplication (Medium, composite 15-16)

**DEBT-147**: Dual source chain construction (Composite: 16). Inline chain in materializer vs `BuildSourceChain()`. "Works in validation, broken in sync" bugs are the risk. Effort: 2-3 hours for full unification.

**DEBT-144**: log.Printf warnings go to void (Composite: 16). 25 warning sites produce output invisible in CC hook context. Effort: 1-2 days to add SyncResult.Warnings.

**DEBT-139**: Zero log.Debug infrastructure (Composite: 15). No debug-level logging in any library package. Blocks efficient debugging of DEBT-138, DEBT-140, DEBT-143 issues. Effort: 1 day infrastructure + 2-3 days instrumentation.

**DEBT-126**: 3-version event bridge (Composite: 15). Removal trigger may already be met (all pre-ADR-0027 sessions archived). FLAG: verify trigger condition -- if met, this becomes a quick win. Effort: 2-3 hours once verified.

### Test Coverage (Medium, composite 12-14)

**DEBT-113**: config package 34.6% coverage (Composite: 14). Foundational package with RISK-003 sync.Once caching hazard. Effort: 1 day.

**DEBT-132**: rite package 38 funcs at 0% (Composite: 14). Rite resolution and management. Effort: 1-2 days.

**DEBT-135**: manifest package 24 funcs at 0% (Composite: 14). Manifest parsing, diffing, merging. Effort: 1-2 days.

**DEBT-134**: cmd/session 25 funcs at 0% (Composite: 12). Edge case handlers in well-tested package (158 test functions exist). Effort: 1-2 days.

### Architecture & Coupling (Medium, composite 12-14)

**DEBT-162**: worktree imports materialize (Composite: 14). Cross-domain coupling pulls in full materialize dependency graph. Effort: medium, needs CLI orchestration refactoring.

**DEBT-159**: naxos imports Layer 2 (Composite: 13). Support-layer package depends on domain packages. May be a documentation error (naxos belongs in Layer 2). Effort: 15 min if reclassifying.

**DEBT-153**: Dual OwnerType definitions (Composite: 13). Naming collision between inscription.OwnerType and provenance.OwnerType. Cross-reference comments exist as guards. Effort: medium rename.

### Observability & Silent Failure (continued)

### Observability & Silent Failure Group

| DEBT-ID | Title | Composite | Key Risk |
|---------|-------|-----------|----------|
| 143 | engine.go stale-entry error discard | 13 | os.RemoveAll errors silently discarded; orphan files persist |
| 140 | extractEmbeddedMena 6 silent errors | 12 | Best-effort extraction with zero logging; debugging is blind |
| 142 | Hook embedded FS wiring boilerplate | 12 | worktreeseed missing 2 of 4 embedded FS sources |

**DEBT-143** (Blast 3, Silent 4, Compound 2, Blocking 1, Effort 1, Regression 2 = 13): Stale mena entries that cannot be removed persist silently. Fix is 30 minutes -- collect errors into `result.Warnings`. Near quick-win territory.

**DEBT-140** (Blast 3, Silent 4, Compound 2, Blocking 1, Effort 1, Regression 1 = 12): 6 error paths in XDG extraction with zero logging. 15-minute fix. Qualifies as quick win but the blast radius makes medium tier appropriate.

**DEBT-142** (Blast 2, Silent 3, Compound 2, Blocking 1, Effort 2, Regression 2 = 12): `worktreeseed` missing embedded Agents and Mena means worktree materialization may be incomplete. The silent failure risk (score 3) is what elevates this -- a worktree looks correct but is missing mena content.

### Test Coverage Group

| DEBT-ID | Title | Composite | Key Risk |
|---------|-------|-----------|----------|
| 101 | lint pkg 12.9% coverage | 14 | Lint rules protecting against SCAR regressions are untested |
| 106 | inscription cmd 0% CLI coverage | 14 | CLAUDE.md sync dispatch untested at CLI layer |
| 134 | cmd/session 25 funcs at 0% | 12 | Edge case handlers in well-tested package |

**DEBT-101** (Blast 3, Silent 2, Compound 2, Blocking 2, Effort 3, Regression 2 = 14): Lint rules are the defensive layer against SCAR-017 and SCAR-019. At 12.9% coverage, most rule implementations are untested. The blocking score (2) reflects that lint improvements are gated on having tests.

**DEBT-106** (Blast 3, Silent 2, Compound 2, Blocking 2, Effort 3, Regression 2 = 14): Inscription sync commands manage CLAUDE.md -- the core user-facing artifact. Library coverage is 83.2% but CLI dispatch is zero.

**DEBT-134** (Blast 2, Silent 2, Compound 2, Blocking 1, Effort 3, Regression 2 = 12): 158 test functions already exist; 25 gaps are likely edge case handlers and formatters.

### Documentation Accuracy Group

| DEBT-ID | Title | Composite | Key Risk |
|---------|-------|-----------|----------|
| 115 | .know/test-coverage.md stale | 13 | 3 packages documented as 0% that are now tested |
| 163 | architecture.md 3 stale claims | 12 | Incorrect leaf list, stale line counts, missing violations |
| 124 | ADR-0028 unwritten | 12 | Significant initiative without formal decision record |

**DEBT-115** (Blast 3, Silent 3, Compound 3, Blocking 2, Effort 1, Regression 1 = 13): Agents reading stale coverage data make wrong prioritization decisions. Quick win -- 15 minutes to regenerate.

**DEBT-163** (Blast 3, Silent 2, Compound 3, Blocking 2, Effort 1, Regression 1 = 12): Three specific inaccuracies in architecture.md. Quick win -- 30 minutes to correct.

**DEBT-124** (Blast 2, Silent 1, Compound 3, Blocking 2, Effort 3, Regression 1 = 12): Knowledge decay accelerates without formal ADR. The compounding score (3) reflects that memories fade and MEMORY.md notes are less durable than ADRs.

### Architecture & Coupling Group

| DEBT-ID | Title | Composite | Key Risk |
|---------|-------|-----------|----------|
| 159 | naxos imports Layer 2 | 13 | Support package depends on domain packages |
| 153 | Dual OwnerType definitions | 13 | Naming collision across inscription and provenance |
| 155 | Deprecated Commands/Skills fields | 12 | Zero internal consumers but satellite audit incomplete |
| 164 | lint.go 784 lines, 5 domains | 12 | Monolith structure blocks incremental test improvement |
| 165 | output.go + rite.go 1,477 lines | 12 | Unbounded growth pattern, merge conflict risk |

**DEBT-159** (Blast 3, Silent 1, Compound 3, Blocking 2, Effort 2, Regression 2 = 13): Layer violation that erodes the architecture model. The compounding score (3) reflects that once a support package imports domain, other support packages lose incentive to stay clean.

**DEBT-153** (Blast 2, Silent 1, Compound 2, Blocking 2, Effort 3, Regression 3 = 13): The naming collision has cross-reference comments as guards, but rename requires touching all consumers in the materialize pipeline. Regression risk (3) because renaming a type used in serialized formats can break backward compat.

**DEBT-155** (Blast 2, Silent 1, Compound 2, Blocking 2, Effort 2, Regression 3 = 12): Zero internal consumers found, but satellite audit is incomplete. Regression risk is high if external consumers exist. Confidence: medium -- caveat documented.

**DEBT-164** (Blast 2, Silent 1, Compound 3, Blocking 3, Effort 2, Regression 1 = 12): The blocking score (3) is the driver -- the monolith structure blocks incremental test improvement for DEBT-101. Splitting enables per-domain test files.

**DEBT-165** (Blast 2, Silent 1, Compound 3, Blocking 2, Effort 3, Regression 1 = 12): Unbounded growth pattern. Every new command adds to 700+ line files. Merge conflict risk during concurrent development.

### Deferred Work & Migration Group

| DEBT-ID | Title | Composite | Key Risk |
|---------|-------|-----------|----------|
| 129 | Single-binary scope unclear | 14 | Initiative with undefined remaining scope |
| 146 | Three copyDir implementations | 12 | Inconsistent write semantics (session copy is non-atomic) |

**DEBT-129** (Blast 2, Silent 1, Compound 2, Blocking 2, Effort 5, Regression 2 = 14): Requires user input. The effort score (5) reflects undefined scope, not implementation difficulty. Scoring assumption: if scope is defined and small, this drops to Low tier.

**DEBT-146** (Blast 3, Silent 2, Compound 2, Blocking 1, Effort 2, Regression 2 = 12): The session version of copyDir uses non-atomic writes (`os.WriteFile` instead of `WriteIfChanged`). Not a current failure mode but a latent consistency risk.

### Radar-Sourced Items (Medium)

| DEBT-ID | Title | Composite | Key Risk |
|---------|-------|-----------|----------|
| 178 | Data corruption systemic pattern (OPP-013) | 17 | Manifest serialization boundary failures; load-time validation missing |
| 174 | fmt.Errorf at CLI boundaries (OPP-002) | 16 | 39 violations lose exit code control and JSON formatting |
| 177 | Schema evolution systemic pattern (OPP-012) | 15 | Schema changes without atomic consumer updates; no registry test |

**DEBT-178** (Blast 3, Silent 4, Compound 3, Blocking 2, Effort 3, Regression 2 = 17): Three SCARs (004/015/022) share a manifest serialization boundary root cause. The codebase lacks load-time schema validation. Silent failure score (4) is the driver -- corrupted manifests parse optimistically.

**DEBT-174** (Blast 3, Silent 3, Compound 3, Blocking 2, Effort 3, Regression 2 = 16): 39 `fmt.Errorf` in RunE handlers. `cmd/org/` has 8 violations, `cmd/knows/` has 5. Errors bypass `PrintError` and structured output. Part of the CLI Convention cluster.

**DEBT-177** (Blast 3, Silent 3, Compound 3, Blocking 2, Effort 2, Regression 2 = 15): SCAR-011/014/016 share a schema-evolution root cause. `NormalizeStatus()` alias map is the correct pattern. A schema registry test enumerating all valid session status values would catch undeclared aliases.

---

## Tier: Low

### Test Coverage (low impact packages)

| DEBT-ID | Title | Composite | Notes |
|---------|-------|-----------|-------|
| 102 | org cmd 1.2% coverage | 7 | Low-frequency admin commands |
| 103 | explain cmd 42.6% coverage | 9 | Formatting-only functions; visible failures |
| 104 | artifact cmd 0% CLI coverage | 10 | Library at 88%, CLI dispatch is thin |
| 105 | common cmd 0% coverage | 8 | Thin accessor functions, tested indirectly |
| 108 | provenance cmd 0% CLI coverage | 11 | Library at 74.7% provides adequate safety |
| 109 | tribute cmd 0% CLI coverage | 6 | Low-frequency, library at 79.1% |
| 110 | naxos cmd 0% CLI coverage | 6 | Library at 80.0% provides adequate safety |
| 111 | root cmd 0% coverage | 6 | Cobra wiring only, tested indirectly |
| 133 | clewcontract 26 funcs at 0% | 9 | Mostly simple constructors, high overall coverage (81.2%) |
| 136 | cmd/sails 39.5% coverage | 11 | Library at 79.3% is adequate |

### Shell & Infrastructure

| DEBT-ID | Title | Composite | Notes |
|---------|-------|-----------|-------|
| 116 | context-injection.sh dead code | 9 | Quick win to remove (30 min). No runtime risk since dead |
| 117 | validation.sh shell in Go codebase | 11 | Live but not on runtime path. Port when convenient |
| 118 | e2e-validate.sh CI-only | 10 | Outside ADR-0011 scope. Accept as bash |
| 119 | Shell 565-line footprint (aggregate) | 10 | Aggregate, remediated via DEBT-116/117/118 |
| 120 | 37 doc refs to dead shell chain | 9 | Quick win cleanup when DEBT-116 is removed |
| 127 | Shell cleanse partially complete | 9 | Nearly done; close when DEBT-116 resolved |
| 128 | Hook bash elimination nearly done | 6 | Effectively complete. Close with DEBT-116 |

### Deferred & Observations

| DEBT-ID | Title | Composite | Notes |
|---------|-------|-----------|-------|
| 121 | Resume cross-rite deferred | 11 | Valid deferral, no code drift |
| 122 | arch-ref skill missing | 6 | Low-frequency rite, minimal impact |
| 123 | 10x-dev "ghost skills" not ghosts | 6 | MEMORY.md correction only |
| 125 | state.json last_sync dead write | 10 | Low urgency; dead write is harmless |
| 130 | state.json full elimination | 10 | Follow-up to DEBT-125, diminishing returns |
| 137 | MEMORY.md ghost skills inaccurate | 10 | Quick win (5 min). Part of documentation cluster |
| 141 | Hook handlers graceful degradation | 6 | Observation, not debt. Intentional design |
| 150 | Session test setup duplication | 11 | Annoying but not dangerous |
| 151 | Platform mena resolution dual paths | 9 | Working correctly, structurally fragile |
| 154 | McpServerConfig naming | 6 | Cosmetic naming inconsistency |
| 156 | SourceType convention stable | 6 | Intentional design decision, not debt |
| 157 | registry not a leaf (docs) | 9 | Quick win -- 15 min doc correction |
| 160 | tribute imports Layer 2 | 11 | Same pattern as DEBT-159; paired reclassification |
| 161 | output zero imports (observation) | 6 | Positive observation, not debt |
| 166 | inscription 3 files over 500 lines | 10 | Partially split, utility extraction optional |
| 167 | worktree operations.go 707 lines | 10 | Import/export extraction optional |
| 168 | sails generator.go 678 lines | 7 | Clean extractor split, optional |
| 169 | clewcontract event.go 644 lines | 9 | Linear growth pattern, optional split |
| 170 | materialize.go 53% extracted | 11 | Diminishing returns below 500-line threshold |
| 172 | writeguard.go 588 lines | 8 | Refactor opportunity, not urgent |

### Radar-Sourced (Low)

| DEBT-ID | Title | Composite | Notes |
|---------|-------|-----------|-------|
| 179 | Historical boundary violations (OPP-014) | 11 | Quick win: lint rule for session artifacts in shared mena |
| 176 | KnossosHome cache poisoning (OPP-011) | 10 | Quick win: add t.Cleanup to 8 test functions |
| 180 | Testify drift 23 vs 18 (OPP-018) | 7 | Convention doc staleness only |

---

## Tier: Resolved

### DEBT-145: RISK-001/004/005 documentation partially stale (Composite: 9 -> Resolved)

**Verification**:
- RISK-001 (agent transform): Confirmed resolved. `transformAgentContent` failures now return `fmt.Errorf` at 3 call sites.
- RISK-004 (CleanEmptyDirs): Partially resolved. Returns `[]error` now, but DEBT-143 documents remaining error discard in stale-entry removal.
- RISK-005 (provenance.Load): Confirmed resolved. Errors now propagate and abort pipeline.

**Remaining work**: Update `.know/design-constraints.md` to reflect resolved status. The partial resolution of RISK-004 is tracked as DEBT-143 (scored Medium, composite 13).

**Risk**: Zero remaining risk from the resolved items. The documentation staleness is a documentation cluster issue (DEBT-115, DEBT-137, DEBT-163).

---

### DEBT-148: TENSION-006 shared manifest loaders unified (Composite: 6 -> Resolved)

**Verification**: `loadSharedManifest` at `agent_transform.go:143-183` confirmed. Both callers (`loadSharedHookDefaults` and skill policies) delegate to it. Three-tier resolution (embedded, KnossosHome, project root) implemented once.

**Risk**: Zero. Fully resolved.

---

### DEBT-152: copyDirFS unification complete (Composite: 6 -> Resolved)

**Verification**: `copyDirFS` at `mena/walker.go:40` confirmed as unified replacement. Comment at line 37-39 explicitly documents the unification. Both embedded and filesystem sources use `openMenaFS()` adapter.

**Risk**: Zero. Fully resolved. This resolution provides the pattern template for DEBT-171 (sync_mena.go parallel paths).

---

## Compounding Clusters

### Cluster 1: Userscope (Combined blast radius: 5, Combined compound rate: 5)

| Item | Role in Cluster | Interaction |
|------|----------------|-------------|
| DEBT-112 | Low test coverage (23.7%) | Silent failures in DEBT-138 go undetected |
| DEBT-138 | 16 error-discard sites in sync | Produces corrupt state that DEBT-112's missing tests cannot catch |
| DEBT-158 | 18 mena coupling call sites | Makes DEBT-171 fixes cascade across sub-package boundary |
| DEBT-171 | 654-line post-split monolith | Parallel sync paths duplicate the error-discard pattern from DEBT-138 |

**Combined risk**: These four items form a feedback loop. Low coverage means silent failures go undetected. Error discards create silent failures. Tight coupling means fixes in one file affect another. The post-split monolith means the surface area for all these problems is large.

**Remediation strategy**: Address as a single sprint. Start with DEBT-138 (error handling in sync paths) and DEBT-112 (test coverage) simultaneously. Then tackle DEBT-171 (unify parallel paths), which reduces the surface area for DEBT-158 (coupling).

**Estimated cluster effort**: 5-7 days (compared to 11-14 days if addressed individually, because fixes overlap).

---

### Cluster 2: Documentation Accuracy (Combined blast radius: 3, Combined compound rate: 4)

| Item | Role in Cluster | Interaction |
|------|----------------|-------------|
| DEBT-115 | test-coverage.md stale | Agents prioritize wrong packages |
| DEBT-137 | MEMORY.md ghost skills inaccurate | Incorrect intelligence propagates to new sessions |
| DEBT-163 | architecture.md 3 stale claims | Wrong leaf list, stale line counts guide wrong decisions |
| DEBT-145 | design-constraints.md RISKs stale | Resolved risks documented as active waste assessment time |
| DEBT-180 | conventions.md testify count stale (radar) | Freeze line is wrong, 5 files outside documented scope |

**Combined risk**: Stale .know/ files are consumed by agents for decision-making. Each stale fact compounds: an agent reads stale architecture.md, makes a wrong assumption, which is reinforced by stale test-coverage.md. Radar OPP-009/010 confirm that both architecture.md and design-constraints.md sit at 0.78 confidence with 4 resolved items still marked active. The compounding rate (4) reflects that every agent invocation that reads stale data perpetuates incorrect decisions.

**Remediation strategy**: Batch fix. Run `/know --force architecture design-constraints` to refresh. Then manually correct MEMORY.md, conventions.md. All are quick wins (effort 1).

**Estimated cluster effort**: 1-2 hours for all 5 items.

---

### Cluster 3: Event Migration (Combined blast radius: 3, Combined compound rate: 2)

| Item | Role in Cluster | Interaction |
|------|----------------|-------------|
| DEBT-126 | 3-version event read bridge | Maintenance burden for historical format support |
| DEBT-153 | Dual OwnerType definitions | Naming confusion in same domain area |

**Combined risk**: Lower than the other clusters. The event bridge is stable and the OwnerType collision has cross-reference guards. The main risk is cognitive overhead during code review and audit.

**Remediation strategy**: Verify DEBT-126 trigger condition first (are all pre-ADR-0027 sessions archived?). If yes, remove the bridge (2-3 hours). DEBT-153 can be addressed independently when the inscription/provenance boundary is next modified.

---

### Cluster 4: Hook Handler Consistency (Discovered during scoring)

| Item | Role in Cluster | Interaction |
|------|----------------|-------------|
| DEBT-149 | Missing withTimeout on 2-3 hooks | Reliability risk |
| DEBT-142 | Embedded FS wiring boilerplate | Incomplete materialization in worktrees |
| DEBT-141 | Hook handler observation (not debt) | Documents the pattern that outliers violate |

**Combined risk**: The outlier hooks (`cheapo_revert`, `worktreeremove`, `worktreeseed`) share the same structural deviation from the standard hook pattern. Fixing them together is more efficient than addressing individually.

**Estimated cluster effort**: 3-4 hours for all fixes.

---

### Cluster 5: Systemic SCAR Patterns (Discovered by radar)

| Item | Role in Cluster | Interaction |
|------|----------------|-------------|
| DEBT-131 | 9 SCARs lack regression tests | Existing assessment: behavioral guards unprotected |
| DEBT-177 | Schema evolution pattern (3 SCARs) | Radar: schema changes without atomic consumer updates |
| DEBT-178 | Data corruption pattern (3 SCARs) | Radar: manifest serialization boundary failures |
| DEBT-179 | Historical boundary pattern (3 SCARs) | Radar: violations caught by revert, not enforcement |

**Combined risk**: The radar decomposes the 27 SCARs into systemic categories. 12 of 27 SCARs cluster into 3 failure patterns (silent failure, schema evolution, data corruption) that share structural root causes. Addressing the root causes (manifest validation at load time, schema registry tests, lint rules for boundary enforcement) prevents future SCAR recurrence, not just the specific past failures.

**Remediation strategy**: Address via systemic fixes rather than per-SCAR regression tests. Manifest schema validation at load time addresses DEBT-178. Schema registry test addresses DEBT-177. Lint rule for session artifacts in shared mena addresses DEBT-179. DEBT-131 regression tests address the remaining individual SCARs.

**Estimated cluster effort**: 4-5 days for systemic fixes + 2-3 days for individual SCAR regression tests.

---

### Cluster 6: CLI Convention Drift (Discovered by radar)

| Item | Role in Cluster | Interaction |
|------|----------------|-------------|
| DEBT-173 | os.Stdout bypass 41 sites | JSON output silently incomplete |
| DEBT-174 | fmt.Errorf at CLI boundaries 39 sites | Error formatting and exit codes lost |
| DEBT-100 | output package 11.7% coverage | No tests to catch regressions in output contract |

**Combined risk**: Two convention-drift findings (80 total violations) plus zero test coverage on the output package mean the CLI's user-facing contract is both broken and unguarded. Fixing the convention violations without first having output tests risks introducing regressions. The coverage gap enables the drift.

**Remediation strategy**: Write output JSON contract tests first (DEBT-100, partial), then refactor stdout bypasses (DEBT-173), then fix fmt.Errorf boundaries (DEBT-174).

**Estimated cluster effort**: 5-7 days sequenced.

---

## Items Requiring User Input

### DEBT-129: Single-binary scope unclear

**Question**: The MEMORY.md "Current Priorities" item 3 lists "Single-binary (ari init, rite embedding, remaining ports)." Investigation shows `ari init` exists and rite embedding is functional via `embed.FS`. The phrase "remaining ports" has no concrete task list.

**Options**:
1. **Close the initiative**: If `ari init` + rite embedding + the existing embed.FS infrastructure constitute "single-binary complete," update MEMORY.md and close. Score drops to Low tier (composite 8).
2. **Define remaining scope**: If there are specific ports still needed, create a task list so effort can be estimated. Current score (14, Medium) reflects the undefined scope.
3. **Defer indefinitely**: Accept current state and remove from Current Priorities.

**Assessment assumption**: Scored at Medium assuming undefined scope. Confidence: low.

### DEBT-126: Event bridge trigger verification

**Question**: The 3-version event read bridge has a documented removal trigger: "once all sessions created before ADR-0027 sprint 3 have been wrapped and archived." This condition may already be met.

**Action needed**: Check the oldest active session's creation date against the ADR-0027 sprint 3 date. If all pre-ADR-0027 sessions are archived, the bridge can be removed immediately (2-3 hours, drops to quick win).

**Assessment assumption**: Scored assuming trigger is not yet verified. Confidence: medium -- the collector noted this is "likely met."

---

## Quick Wins

Items with fix effort = 1 AND composite score >= 9, sorted by composite score descending. These provide the highest return on investment.

| DEBT-ID | Title | Composite | Effort | Time Estimate | Impact |
|---------|-------|-----------|--------|---------------|--------|
| 115 | Regenerate test-coverage.md | 13 | 1 | 15 min | Fixes agent decision-making data |
| 143 | engine.go stale-entry error collection | 13 | 1 | 30 min | Surfaces silent orphan file persistence |
| 163 | Fix architecture.md 3 stale claims | 12 | 1 | 30 min | Corrects leaf list, line counts, layer violations |
| 140 | Add logging to extractEmbeddedMena | 12 | 1 | 15 min | Makes XDG extraction debuggable |
| 179 | Lint rule for session artifacts in shared mena | 11 | 1 | 2-3 hours | Enforces SCAR-027 boundary automatically |
| 176 | Fix KnossosHome cache poisoning in tests | 10 | 1 | 30 min | Prevents test pollution from sync.Once cache |
| 137 | Correct MEMORY.md ghost skills note | 10 | 1 | 5 min | Removes incorrect intelligence |
| 145 | Update design-constraints.md RISKs | 9 | 1 | 15 min | Marks resolved risks correctly |
| 157 | Correct architecture.md leaf list | 9 | 1 | 15 min | Overlaps with DEBT-163 |
| 116 | Remove dead context-injection.sh | 9 | 1 | 30 min | Eliminates dead code and confusion |
| 120 | Clean 37 dead doc references | 9 | 1 | 1-2 hours | Accompanies DEBT-116 removal |

**Total quick win effort**: ~5 hours for 11 items (DEBT-116 and DEBT-120 are paired).

**Quick win cluster (documentation)**: DEBT-115, DEBT-137, DEBT-145, DEBT-157, DEBT-163 are all documentation accuracy fixes that can be done in a single 1-hour session. These resolve the entire Documentation Accuracy Cluster.

**Quick win cluster (code fixes)**: DEBT-143 (error collection), DEBT-140 (add logging), DEBT-176 (test cleanup) are mechanical code changes totaling ~1 hour. DEBT-179 (lint rule) is a slightly larger quick win at 2-3 hours but provides automated enforcement for a recurring SCAR pattern.

---

## Radar Integration

Radar scan date: 2026-03-03. Source: `.know/radar.md` (7 signals, 18 opportunities). Archive: `.ledge/reviews/RADAR-2026-03-03.md`.

### OPP-to-DEBT Cross-Reference

| Radar OPP | Type | DEBT Item(s) | Risk Report Effect |
|-----------|------|-------------|-------------------|
| OPP-001 | New item | **DEBT-173** | Added. Composite 21, High. os.Stdout bypass |
| OPP-002 | New item | **DEBT-174** | Added. Composite 16, Medium. fmt.Errorf at CLI boundaries |
| OPP-003 | Overlaps | DEBT-159, DEBT-160 | Confirms reclassification as fix. No rescore |
| OPP-004 | Enriches | DEBT-138 | Reinforces "exit-0 must mean complete output" framing. No rescore (already 25) |
| OPP-005 | Enriches | DEBT-131 | Adds CC integration smoke-test harness recommendation. No rescore (already 22) |
| OPP-006 | Enriches | DEBT-104, **DEBT-106**, **DEBT-107** | **Rescored DEBT-106**: Blast 3->4, Silent 2->3, composite 14->16. **Rescored DEBT-107**: Silent 2->3, composite 11->12, tier Low->Medium. Inscription rollback flagged as highest-consequence untested op |
| OPP-007 | New item | **DEBT-175** | Added. Composite 23, Critical. Non-atomic writes on state files |
| OPP-008 | Overlaps | DEBT-143 | Confirms stale cleanup loop as primary target. No rescore |
| OPP-009 | Overlaps | DEBT-115, DEBT-145, DEBT-163 | Adds specific list of 4 resolved TENSIONs + 1 stale line count. No rescore |
| OPP-010 | Overlaps | DEBT-115, DEBT-163 | Flags 0.78 confidence on both .know/ files. No rescore |
| OPP-011 | New item | **DEBT-176** | Added. Composite 10, Low (quick win). Test cache poisoning |
| OPP-012 | New item | **DEBT-177** | Added. Composite 15, Medium. Schema evolution pattern |
| OPP-013 | New item | **DEBT-178** | Added. Composite 17, Medium. Data corruption pattern |
| OPP-014 | New item | **DEBT-179** | Added. Composite 11, Low (quick win). Boundary enforcement |
| OPP-015 | Overlaps | DEBT-162 | Adds sails->session and sails->clewcontract imports. No rescore |
| OPP-016 | Enriches | DEBT-101, coverage items | Notes that sync/rite/lint tests exercise metadata only, not behavior. No rescore |
| OPP-017 | Overlaps | DEBT-157 | Confirms documentation reclassification as fix. No rescore |
| OPP-018 | New item | **DEBT-180** | Added. Composite 7, Low. Convention doc staleness |

### Rescored Items

| DEBT-ID | Original Scores | New Scores | Reason |
|---------|----------------|------------|--------|
| 106 | Bl:3 Si:2 C:14 | Bl:4 Si:3 C:16 | OPP-006: inscription rollback is highest-consequence untested op. SCAR-003/005 destructive-write patterns apply |
| 107 | Si:2 C:11, Low | Si:3 C:12, Medium | OPP-006: manifest validate feeds materialize pipeline. Regression produces silent downstream failures |

---

## Convergent Hotspot: Materialize Pipeline

The radar's advisory section independently identifies the **materialize pipeline** as the convergent hotspot. Three independent signals (unguarded scars, convention drift, architecture decay) all point to `internal/materialize/` from different angles:

| Signal | Finding | DEBT Items |
|--------|---------|------------|
| Unguarded scars | Untested CLI entry points feeding into the pipeline | DEBT-106, DEBT-107, DEBT-131 |
| Convention drift | os.Stdout bypasses and fmt.Errorf at boundaries | DEBT-173, DEBT-174 |
| Architecture decay | Layer misclassifications around the pipeline | DEBT-159, DEBT-160 |
| Silent failure | Error discards within the pipeline | DEBT-138, DEBT-143, DEBT-175 |
| Recurring scars | 4 silent-failure SCARs, 3 data-corruption SCARs | DEBT-177, DEBT-178 |

**This aligns with but EXPANDS the userscope cluster** from the original assessment. The userscope cluster (DEBT-112 + 138 + 158 + 171) focuses on the sync subsystem within materialize. The convergent hotspot encompasses the entire materialize ecosystem: the CLI entry points that feed it, the sync paths within it, the convention violations around it, and the architectural boundaries that contain it.

**Convergent hotspot items by tier**:

| Tier | Items | Count |
|------|-------|-------|
| Critical | DEBT-138, DEBT-175 | 2 |
| High | DEBT-100, DEBT-173 | 2 |
| Medium | DEBT-106, DEBT-107, DEBT-143, DEBT-147, DEBT-158, DEBT-171, DEBT-174, DEBT-177, DEBT-178 | 9 |
| Low | DEBT-159, DEBT-160, DEBT-176, DEBT-179 | 4 |

**Total**: 17 items in the convergent hotspot, spanning all tiers. This represents 21% of the entire debt portfolio concentrated in one pipeline area.

**Strategic recommendation**: A combined hygiene + arch review session targeting the materialize ecosystem would address the highest-concentration risk area. Workstream 1 (below) packages this.

---

## Recommended Workstreams

Five themed workstreams that group related items for efficient remediation. Each could become a `/start` invocation.

### Workstream 1: Materialize Pipeline Hardening

**Rite**: `hygiene` (for code changes) then `arch` (for architecture doc updates)

**Theme**: Address the convergent hotspot. Make the materialize pipeline reliable: atomic writes, error propagation, and post-sync verification.

| Item | Description | Effort |
|------|-------------|--------|
| DEBT-175 | Replace os.WriteFile with AtomicWriteFile on 4 critical state files | 1 day |
| DEBT-138 | Fix 16 checksum error-discard sites in userscope/ | 1-2 days |
| DEBT-143 | Collect engine.go stale-entry errors into result.Warnings | 30 min |
| DEBT-140 | Add logging to extractEmbeddedMena 6 silent error paths | 15 min |
| DEBT-149 | Add withTimeout to cheapo_revert and worktreeremove | 2-4 hours |
| DEBT-142 | Fix worktreeseed missing embedded Agents and Mena | 1-2 hours |

**Total effort**: 3-5 days
**Dependencies**: None. Can start immediately.
**Priority**: Highest. Contains 2 Critical items and the convergent hotspot core.
**Invocation**: `/start hygiene -c SPRINT "Materialize pipeline hardening: atomic writes, error propagation, timeout safety"`

---

### Workstream 2: CLI Convention Alignment

**Rite**: `hygiene`

**Theme**: Fix the 80 convention violations (stdout bypass + fmt.Errorf) and establish guards against recurrence.

| Item | Description | Effort |
|------|-------------|--------|
| DEBT-173 | Refactor 41 os.Stdout bypass sites to use Printer | 2-3 days |
| DEBT-174 | Replace 39 fmt.Errorf in RunE with PrintError | 1-2 days |
| DEBT-100 | Write JSON output contract tests for output package | 2-3 days (partial) |

**Total effort**: 5-7 days
**Dependencies**: DEBT-100 tests should be written first (or in parallel) to catch regressions during DEBT-173/174 refactoring.
**Priority**: High. DEBT-173 is High tier; combined blast radius affects every CLI command.
**Invocation**: `/start hygiene -c SPRINT "CLI convention alignment: Printer routing, structured errors, output tests"`

---

### Workstream 3: SCAR Regression Safety Net

**Rite**: `debt-triage` (for systemic analysis) then `hygiene` (for test writing)

**Theme**: Build automated protection against the 3 systemic SCAR patterns identified by radar, plus individual regression tests for the highest-risk unguarded SCARs.

| Item | Description | Effort |
|------|-------------|--------|
| DEBT-131 | Write regression tests for SCAR-004 and SCAR-023 | 1 day |
| DEBT-178 | Add manifest schema validation at load time | 1-2 days |
| DEBT-177 | Add schema registry test for session status values | 1 day |
| DEBT-179 | Add ari lint rule for session artifacts in shared mena | 2-3 hours |
| DEBT-176 | Fix KnossosHome cache poisoning in 8 tests | 30 min |

**Total effort**: 4-5 days
**Dependencies**: DEBT-179 lint rule depends on understanding the current lint.go structure (DEBT-164). Consider splitting lint.go first if it is in scope.
**Priority**: High. Contains 1 Critical item (DEBT-131) and prevents future SCAR recurrence.
**Invocation**: `/start debt-triage -c SPRINT "SCAR regression safety net: schema validation, boundary enforcement, regression tests"`

---

### Workstream 4: Knowledge Refresh

**Rite**: None needed -- this is a knowledge maintenance session.

**Theme**: Fix all stale .know/ files and documentation in a single batch. Quick wins only.

| Item | Description | Effort |
|------|-------------|--------|
| DEBT-115 | Regenerate test-coverage.md | 15 min |
| DEBT-163 | Fix architecture.md 3 stale claims | 30 min |
| DEBT-145 | Update design-constraints.md resolved RISKs | 15 min |
| DEBT-137 | Correct MEMORY.md ghost skills note | 5 min |
| DEBT-157 | Correct architecture.md leaf list | 15 min (overlaps 163) |
| DEBT-180 | Update conventions.md testify count | 15 min |
| DEBT-116 | Remove dead context-injection.sh | 30 min |
| DEBT-120 | Clean 37 dead doc references | 1-2 hours |

**Total effort**: 2-3 hours
**Dependencies**: None. Can be done at any time.
**Priority**: Low individually, but high ROI -- resolves the entire Documentation Accuracy Cluster and brings .know/ files above 0.85 confidence.
**Invocation**: `/know --force architecture design-constraints test-coverage` followed by manual corrections for MEMORY.md and conventions.md.

---

### Workstream 5: Userscope Structural Remediation

**Rite**: `hygiene`

**Theme**: Address the userscope cluster structural debt -- coupling, parallel paths, coverage. This is the deep remediation after Workstream 1 handles the immediate error-propagation fixes.

| Item | Description | Effort |
|------|-------------|--------|
| DEBT-112 | Write tests for userscope sync paths (sync_mena.go first) | 3-4 days |
| DEBT-171 | Unify syncUserMena and syncUserMenaFromEmbedded | 1-2 days |
| DEBT-158 | Extract shared mena sync interface to decouple userscope from mena | 1-2 days |
| DEBT-147 | Unify dual source chain construction | 2-3 hours |

**Total effort**: 5-8 days
**Dependencies**: Workstream 1 should complete first (error propagation fixes make test writing more effective). DEBT-171 follows the pattern from DEBT-152 (resolved copyDirFS unification).
**Priority**: Medium. The immediate risks are handled by Workstream 1; this addresses the structural underpinnings that create those risks.
**Invocation**: `/start hygiene -c SPRINT "Userscope structural remediation: test coverage, path unification, coupling reduction"`

---

### Workstream Sequencing

```
Week 1-2:  WS4 (Knowledge, 3h)  +  WS1 (Pipeline Hardening, 3-5d)
Week 2-3:  WS3 (SCAR Safety Net, 4-5d)
Week 3-4:  WS2 (CLI Convention, 5-7d)
Week 4-6:  WS5 (Userscope Structural, 5-8d)
```

WS4 is prerequisite-free and provides immediate ROI (1 session). WS1 is the highest-priority code workstream and can run in parallel with WS4. WS3 builds on WS1's error-propagation fixes. WS2 is independent but benefits from WS1's output test groundwork. WS5 is the deep structural work that should follow WS1 to avoid reworking error-handling during structural changes.

**Total portfolio effort**: 19-28 days across 5 workstreams.

---

## Assessment Limitations

1. **Silent failure scoring is inference-based**: The silent failure dimension is scored based on static analysis of what error paths exist, not on observed failures. Actual silent failure frequency may be lower than scored if the error conditions are rare. Confidence: medium.

2. **Compounding rate assumes continued development velocity**: Scores assume the codebase continues to grow at its current rate. If development slows, compounding scores would be lower. Confidence: high (the codebase has been actively developed).

3. **Effort estimates inherited from collector**: The fix effort dimension uses the collector's estimates (e.g., "1-2 days", "30 minutes"). These are not independently verified. Effort 1 = under 1 hour, 2 = half day, 3 = 1-2 days, 4 = 3-5 days, 5 = 1+ week. Confidence: medium.

4. **Regression risk for sync paths is conservative**: Items touching the materialize/userscope sync pipeline are scored regression 3+ because these paths are load-bearing and partially tested. The actual regression risk depends on the quality of existing tests in adjacent code. Confidence: medium.

5. **Satellite manifest audit gap**: DEBT-155 (deprecated Commands/Skills fields) assumes zero external consumers based on internal grep. External satellite manifests were not audited. If external consumers exist, the regression risk for removal increases from 3 to 5. Confidence: low.

6. **Layer violation severity may be over-scored**: DEBT-159 and DEBT-160 may represent a documentation error (naxos and tribute belong in Layer 2) rather than actual boundary erosion. If reclassified, these drop from Medium to Low. Confidence: medium.

7. **Hook timeout interaction with CC**: DEBT-149 scores silent failure at 4, assuming CC does not enforce its own process-level timeout on hooks. If CC does enforce timeouts, the effective risk is lower (CC kills the hook process). Confidence: low -- CC's timeout behavior for sync hooks is not documented in the ledger.

8. **Aggregate items (DEBT-114, DEBT-119) are scored for visibility**: These do not represent independently actionable work. Their composite scores reflect the portfolio-level risk, not sprint-level priority.

9. **Radar confidence propagation**: Radar OPP confidence scores (0.66-0.84) were used as qualitative input but not mechanically propagated into the 6-dimension scores. Items with lower radar confidence (OPP-014 at 0.66, OPP-016 at 0.68) should be treated with more caution.

10. **Non-atomic write count includes intentional sites**: DEBT-175's 59-site count includes `os.WriteFile` on files that may be non-critical (temp files, logs). The 4 highest-risk sites (rite state, worktree metadata) are confirmed critical; the remaining 55 need individual triage.

11. **Convention violation counts are grep-based**: DEBT-173 (41 stdout sites) and DEBT-174 (39 fmt.Errorf sites) counts come from the radar's static analysis. Some may be false positives (e.g., test helpers, intentional direct writes in non-Printer contexts).

---

## Scoring Methodology Notes

**Composite formula**: Sum of all 6 dimensions (range 6-30). This differs from the multiplicative formula in the standard risk-assessor framework ((Blast x Likelihood) / Effort) because the initiative frame specifies 6 additive dimensions rather than 3.

**Tier thresholds**:
- Critical: Blast radius >= 4 AND (silent failure >= 4 OR compounding >= 4)
- High: Composite >= 20 OR blocking factor >= 4
- Medium: Composite 12-19 (after Critical/High filtering)
- Low: Composite < 12
- Resolved: Confirmed fixed during collection

**Where assessor scoring differs from collector severity**:

| DEBT-ID | Collector Severity | Assessor Tier | Reason for Difference |
|---------|-------------------|---------------|----------------------|
| 112 | high | Critical | Cluster analysis elevated -- anchors the userscope feedback loop with blast 5 + silent 4 |
| 131 | high | Critical | Silent failure dimension (4) meets Critical threshold; SCAR regression is systemic |
| 138 | high | Critical | Silent failure (5) is the highest in the ledger; cluster anchor with blast 5 |
| 175 | high (radar) | Critical | Blast 4 + Silent 5 meets Critical threshold. Crash-corruption with no recovery |
| 100 | high | High | Composite 21 meets High threshold (>=20); does not meet Critical criteria (silent 3, compound 3) |
| 149 | low (via DEBT-141) | High | The withTimeout gap is a latent CC-freeze risk, not a "low observation" |
| 173 | high (radar) | High | Composite 21 meets High threshold. JSON output contract broken silently |
| 106 | medium | Medium (rescored) | Radar OPP-006 elevated blast 3->4, silent 2->3. Inscription rollback is destructive |
| 107 | medium | Medium (rescored) | Radar OPP-006 elevated silent 2->3. Manifest validate feeds pipeline silently |
| 159 | high | Medium | Layer violations are documentation fixes if naxos is reclassified to Layer 2 |
| 160 | high | Low | Composite 11 is below Medium threshold (12); reclassification resolves the issue |
| 137 | low | Low (but quick win) | Composite is 10, stays Low, but flagged as quick win for documentation cluster |

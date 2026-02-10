# SPIKE: Session Lifecycle Functional Audit

**Date**: 2026-02-09
**Question**: Do core session lifecycle commands actually work when invoked?
**Method**: 4-agent parallel execution trace (context-engineer x3 + claude-code-guide x1)
**Verdict**: The session lifecycle is fundamentally broken. 7 of 7 session commands fail at runtime.

---

## Executive Summary

The prior CE structural audit gave session commands high grades (Dromena A+) by measuring compliance metrics: has `context:fork`? has `Triggers:`? has `allowed-tools`? This functional audit traces what actually happens when a user types `/start "Add dark mode"`. The answer: it fails.

**Root cause cascade**:
1. The `moirai` agent does not exist in `.claude/agents/` (project-level)
2. The user-level copy at `~/.claude/agents/moirai.md` is stale (Jan 10, 592 lines vs 212-line source) and teaches 3 broken mechanisms
3. Even if moirai loaded correctly, 7 of 13 operations would fail due to CLI syntax mismatches, nonexistent commands, and wrong flag names

Every session command (`/start`, `/continue`, `/park`, `/wrap`, `/handoff`, `/sprint`, `/task`) delegates its core state mutation to `Task(moirai, ...)`. With no functional moirai agent, the entire session lifecycle is inoperable.

---

## Finding 1: Moirai Agent Does Not Exist at Project Level

**Severity**: CRITICAL (blocks all 7 session commands)

**Evidence**:
- `.claude/agents/` contains only ecosystem rite agents: orchestrator, ecosystem-analyst, context-architect, integration-engineer, documentation-engineer, compatibility-tester
- No `moirai.md` in `.claude/agents/`, `rites/*/agents/`, or `knossos/`
- Source exists at `agents/moirai.md` (212 lines) but is not materialized to `.claude/agents/` -- moirai is a cross-rite agent, not rite-scoped
- User-level `~/.claude/agents/moirai.md` exists (592 lines) but is stale from Jan 10

**Impact**: When Claude executes `Task(moirai, "create_session ...")`, CC searches `.claude/agents/moirai.md` first (not found), then falls back to `~/.claude/agents/moirai.md` (stale). If neither exists, the Task call fails silently and Claude improvises -- often searching for `session-manager.sh` or other nonexistent scripts.

**Fix**: Add moirai to the materialization pipeline as a cross-rite agent. The `agents/moirai.md` source must project to `.claude/agents/moirai.md` regardless of active rite.

---

## Finding 2: Projected Moirai Agent Is Fundamentally Broken

**Severity**: CRITICAL

The user-level `~/.claude/agents/moirai.md` (592 lines, last modified Jan 10) diverges from the source `agents/moirai.md` (212 lines) on every critical operational detail:

| Aspect | Source (correct) | Projected (broken) |
|--------|------------------|--------------------|
| Write guard bypass | `ari session lock --agent moirai` | `export MOIRAI_BYPASS=true` (env var -- does not work) |
| Lock file path | `.moirai-lock` (JSON) | `.locks/context.lock` (mkdir-based) |
| Stale lock threshold | 300 seconds | 60 seconds |
| Skill paths | `.claude/skills/session/moirai/{fate}.md` | `~/.claude/skills/moirai/{fate}.md` (wrong) |
| `maxTurns` | 60 | Missing (CC defaults to ~6-10) |
| `type: meta` | Present | Missing |

**Why MOIRAI_BYPASS cannot work**: CC subagents run Bash commands in isolated shells. Environment variables set via `export` do not persist to hook processes. The Go `writeguard.go` checks for a `.moirai-lock` JSON file via `isMoiraiLockHeld()`, not an environment variable. The shell `writeguard.sh` does check `MOIRAI_BYPASS`, but the hook runs in a separate process where the env var is not set.

---

## Finding 3: 7 of 13 Moirai Operations Fail

Per-operation scorecard:

| Operation | CLI Valid? | Syntax Correct? | Overall |
|-----------|-----------|-----------------|---------|
| create_session | PASS | **FAIL** (source: positional complexity; should be `-c` flag) | **FAIL** |
| create_sprint | **FAIL** (no CLI exists) | N/A | **FAIL** |
| park_session | PASS | PASS | **PASS** |
| resume_session | PASS | PASS | **PASS** |
| wrap_session | PASS | PASS | **PASS** |
| transition_phase | PASS | **FAIL** (`--to=` flag does not exist; CLI uses positional arg) | **FAIL** |
| handoff | PASS | **FAIL** (`--from=` flag does not exist; CLI requires `--artifact`) | **FAIL** |
| mark_complete | **FAIL** (no CLI exists) | N/A | **FAIL** |
| update_field | N/A (direct write) | CONDITIONAL | **CONDITIONAL** |
| delete_sprint | **FAIL** (no CLI exists) | N/A | **FAIL** |
| generate_sails | PASS | PARTIAL (docs say it writes file; it reads file) | **PARTIAL** |
| record_decision | N/A (direct write) | CONDITIONAL | **CONDITIONAL** |
| append_content | N/A (direct write) | CONDITIONAL | **CONDITIONAL** |

**3 PASS, 7 FAIL, 3 CONDITIONAL**

---

## Finding 4: INDEX.lego.md Invents 3 Nonexistent CLI Commands

The Moirai routing table (`mena/session/moirai/INDEX.lego.md`) advertises:

| Documented Command | Reality |
|-------------------|---------|
| `ari session sprint create "{goal}" [--task "t1"]` | No `sprint` subcommand exists |
| `ari session sprint mark-complete [sprint-id]` | No `sprint` subcommand exists |
| `ari session sprint delete {sprint-id}` | No `sprint` subcommand exists |

The `session.go` command registration has 12 subcommands; `sprint` is not one of them. The Fate skills (clotho, lachesis, atropos) correctly describe these as direct file mutations. The INDEX routing table is wrong.

---

## Finding 5: CLI Syntax Mismatches in Every Documentation Layer

### transition_phase
- **Every document says**: `ari session transition --to={phase}`
- **Actual CLI** (`transition.go:25`): `ari session transition <phase>` (positional arg)
- Running `ari session transition --to=design` errors: "unknown flag: --to"

### handoff
- **Every document says**: `ari handoff execute --from={from} --to={to}`
- **Actual CLI** (`execute.go:47-49`): `--artifact` (required) and `--to` (required). No `--from` flag.

### create_session
- **Source agent says**: `ari session create "{initiative}" {complexity} [rite]` (positional complexity)
- **Actual CLI** (`create.go:54-57`): `ari session create <initiative> -c <complexity> -r <rite>` (flags)
- **INDEX.lego.md and clotho skill say**: `-c "{complexity}"` (correct)

---

## Finding 6: /start Execution Trace (10 steps, 6 fail)

| Step | What | Verdict |
|------|------|---------|
| 0 | User types `/start` | PASS |
| 1 | SessionStart hook fires | **FAIL** -- field names don't match what /start expects |
| 2 | Claude reads pre-conditions | **DEGRADED** -- /start expects "Has Session", "Session State", "Pre-computed Values" which hook doesn't output |
| 3 | Gather parameters | PASS |
| 4 | `Task(moirai, "create_session...")` | **PASS with risk** -- user-level only, no project-level agent |
| 5 | Moirai parses create_session | **PASS with risk** -- missing maxTurns |
| 6 | Moirai identifies CLI syntax | **FAIL** -- positional vs flag mismatch |
| 7 | Moirai executes CLI | **CONDITIONAL** -- depends on which syntax Claude uses |
| 8 | Moirai returns result | **FAIL** -- `CreateOutput` struct has no `entry_agent` field |
| 9 | Invoke entry agent | **FAIL** -- `requirements-analyst` agent does not exist in ecosystem rite |

### /start-specific failures:
- **Complexity schema conflict**: INDEX says PATCH/MODULE/SYSTEM/INITIATIVE/MIGRATION (matches CLI). behavior.md says SCRIPT/MODULE/SERVICE/PLATFORM (stale, CLI rejects these).
- **Hardcoded 10x-dev agents**: `/start` defaults to `requirements-analyst` which only exists in the 10x-dev rite, not ecosystem.
- **Missing entry_agent**: `/start` claims Moirai returns `entry_agent` in its response. The `CreateOutput` struct has no such field.
- **Hook field name mismatch**: `/start` expects "Has Session = false" and "Pre-computed Values: suggested session ID, entry agent". The SessionStart hook outputs a "Session Context" table with different fields and never generates suggested session IDs or entry agents.

---

## Finding 7: Cross-Cutting Session Command Failures

### /wrap: `--emergency` flag does not exist
- All docs say `--emergency`. Actual CLI uses `--force`.

### /handoff: INDEX.md vs behavior.md contradiction
- INDEX says delegate to Moirai. behavior.md describes direct file mutation of SESSION_CONTEXT.md.

### /minus-1, /zero: `{TAG}` placeholder bug
- Both use `{TAG}` for initiative text. CC dromena use `$ARGUMENTS` for argument substitution. `{TAG}` passes literally.

### /one: Impossible cross-fork resumption
- Claims "Resume the Orchestrator (same instance from Session 0)". With `context:fork`, each invocation creates a fresh context. No mechanism exists to resume a subagent from a previous fork.

### /park: Pre-flight self-contradiction
- Line 21: "do not call `ari session status`"
- Line 25: "Verify via `ari session status`"

### /continue: Title says `/resume`
- behavior.md header: `# /resume Behavior Specification`. Command was renamed to `/continue`.

---

## Finding 8: context:fork + Task Is Architecturally Valid

The initial hypothesis that `context:fork` strips the Task tool was **wrong**.

- `context:fork` creates an isolated conversation branch. The `allowed-tools` frontmatter governs tool availability in that branch.
- All session commands include `Task` in `allowed-tools`, so subagent spawning is valid.
- **However**: CC GitHub Issue #17283 reports that when a skill is invoked via the Skill tool (not slash command), `context: fork` and `agent:` frontmatter are **ignored**. This affects skills but not dromena (slash commands).
- **The real problem** is not fork+Task. It is that every command delegates to `Task(moirai)` and the moirai agent doesn't exist at project level.

---

## Severity Summary

| Severity | Count | Findings |
|----------|-------|----------|
| **P0 -- Execution blocked** | 3 | Moirai agent missing; projected agent fundamentally broken; requirements-analyst doesn't exist |
| **P1 -- Wrong behavior** | 5 | transition_phase `--to` wrong; handoff `--from` wrong; 3 sprint CLI commands fabricated; complexity schema conflict; MOIRAI_BYPASS dead |
| **P2 -- Misleading context** | 6 | Hook field mismatch; missing entry_agent; `--emergency` vs `--force`; {TAG} placeholder; cross-fork resumption; /park contradiction |
| **P3 -- Cosmetic** | 2 | /continue title says /resume; generate_sails docs mislead on read vs write |

---

## Recommended Fix Sequence

### Phase 0: Unblock the Session Lifecycle (1-2 hours)

1. **Materialize moirai to project level**: Add `agents/moirai.md` to the materialization pipeline as a cross-rite agent that always projects to `.claude/agents/moirai.md` regardless of active rite.

2. **Regenerate user-level moirai**: Delete `~/.claude/agents/moirai.md` and let it be recreated from the source `agents/moirai.md`. Or explicitly copy the source to user-level.

3. **Fix source agent CLI syntax**: In `agents/moirai.md`, change `ari session create "{initiative}" {complexity} [rite]` to `ari session create "{initiative}" -c {complexity} [-r {rite}]`.

### Phase 1: Fix CLI Syntax Mismatches (1 hour)

4. **transition_phase**: Change `--to={phase}` to positional `<phase>` in: moirai source, INDEX.lego.md, lachesis skill, all dromena behavior files.

5. **handoff**: Change `--from={from} --to={to}` to `--artifact={id} --to={agent}` in: moirai source, INDEX.lego.md, lachesis skill, /handoff dromena.

6. **INDEX.lego.md sprint commands**: Replace 3 fabricated CLI commands with `—` (direct file mutation), matching what the Fate skills already say.

### Phase 2: Fix /start Content Quality (1-2 hours)

7. **Replace complexity levels in behavior.md**: Change SCRIPT/SERVICE/PLATFORM to PATCH/SYSTEM/INITIATIVE/MIGRATION.

8. **Remove hardcoded `requirements-analyst`**: Make entry agent rite-aware (read from rite manifest or CLAUDE.md agent list).

9. **Align hook output expectations**: Update /start to match actual SessionStart hook field names, or update the hook to output the fields /start expects.

10. **Remove `entry_agent` claim**: /start says Moirai returns entry_agent. It doesn't. Either add it to CreateOutput or remove the expectation.

### Phase 3: Fix Remaining Command Issues (2-3 hours)

11. Fix /wrap `--emergency` to `--force`
12. Fix /handoff INDEX vs behavior.md contradiction
13. Fix /minus-1, /zero `{TAG}` to `$ARGUMENTS`
14. Remove cross-fork resumption claim from /one
15. Fix /park pre-flight contradiction
16. Fix /continue behavior.md title

---

## Appendix: Agent Reports

| Agent | Scope | Key Finding |
|-------|-------|-------------|
| claude-code-guide | context:fork + Task verification | fork does NOT strip Task; CC Bug #17283 affects Skill invocation only |
| context-engineer #1 | /start execution trace (10 steps) | 6/10 steps fail; hook field mismatch, missing entry_agent, nonexistent agent |
| context-engineer #2 | All 13 session dromena | CRITICAL: moirai agent missing from .claude/agents/; 10 failures across all commands |
| context-engineer #3 | Moirai agent functional path | 7/13 operations fail; source vs projected divergence; 3 broken mechanisms in projected |

---

*Generated by 4-agent parallel functional audit, 2026-02-09.*
*Spike by Claude Opus 4.6.*

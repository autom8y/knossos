---
domain: feat/hook-infrastructure
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/hook/**/*.go"
  - "./internal/cmd/hook/**/*.go"
  - "./config/hooks.yaml"
  - "./docs/decisions/ADR-0002*.md"
  - "./docs/decisions/ADR-0011*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# CC Hook Infrastructure

## Purpose and Design Rationale

Purpose-built Go binaries registered into CC's `settings.local.json` that transform raw lifecycle events into first-class platform behaviors. Replaced 13+ bash scripts (~5,700 LOC) that had no structured types, no test coverage, and fragile library resolution.

**ADR-0002**: Hook resolution architecture. **ADR-0011**: Go migration timeline. **SCAR-009**: Flat output format silently rejected by CC. **SCAR-010**: Hooks without timeout blocked indefinitely.

## Conceptual Model

### CC Sends Data via Stdin JSON

`StdinPayload` at `/Users/tomtenuta/Code/knossos/internal/hook/env.go:48`. Environment variables are deprecated fallbacks.

### Two Hook Categories

- **Blocking** (PreToolUse): output `permissionDecision` that CC reads to allow/deny
- **Side-effect** (PostToolUse, SessionStart, etc.): perform work, return JSON for observability

### Fail-Open vs Fail-Closed Policy

**Fail-closed**: write-guard (unknown section → deny), agentguard (missing file_path → deny), protected file without Moirai lock → deny.

**Fail-open**: agentguard JSON parse error → allow, budget I/O error → allow, clew write failures → allow.

### Timeout Discipline

All hooks: `DefaultTimeout=100ms`, `MaxTimeout=500ms`. All git subprocesses use `exec.CommandContext`.

## Implementation Map

Domain: `internal/hook/` (5 files: StdinPayload, Env, ToolInput, HookEvent, output). CLI: `internal/cmd/hook/` (28 files: 14 hook handlers).

### 13 Hook Registrations (from `config/hooks.yaml`)

SessionStart (context), SessionEnd (sessionend), Stop (autopark), PreCompact (precompact), PreToolUse (writeguard, validate, git-conventions), PostToolUse (clew, budget), WorktreeCreate (worktree-seed), WorktreeRemove (worktree-remove), SubagentStart/Stop (subagent).

### Key Flows

- **Write-guard**: Parse file_path → isProtectedFile? → classifyEditSection → Moirai lock check → allow/deny
- **Clew**: RecordToolEvent → supplemental events → throughline extraction → CheckTriggers
- **Context**: Load session → resolve rite → collect git/agents/rites → `.know/` freshness → emit
- **PreCompact → SessionStart rehydration**: writes `COMPACT_STATE.md` → CC compacts → context hook reads and injects

### SCAR Evidence

SCAR-008 (budget async log spam), SCAR-009 (flat output format), SCAR-010 (timeout-less hooks), SCAR-011 (deprecated .current-session), SCAR-012 (archived session infinite retry).

## Boundaries and Failure Modes

- Hooks do NOT require project context (discover from StdinPayload.CWD)
- JSON output always forced regardless of global `-o` flag
- Binary path mismatch trap: `go build` writes local `./ari`, CC uses PATH binary
- Deprecated env vars kept for backward compat (FROZEN constraint)
- Three-generation event schema coexistence in events.jsonl (TENSION-006)

## Knowledge Gaps

1. `cheapo_revert.go`, `worktreeseed.go`, `worktreeremove.go` not read in full.
2. `clewcontract/triggers.go` full ruleset not examined.
3. Budget hook `ARI_MSG_PARK` threshold configuration unclear.

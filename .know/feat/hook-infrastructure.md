---
domain: feat/hook-infrastructure
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/hook/**/*.go"
  - "./internal/cmd/hook/**/*.go"
  - "./config/hooks.yaml"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# CC / Gemini Hook Infrastructure

## Purpose and Design Rationale

Live interception layer over AI harness lifecycle events. Seven platform behaviors: context file protection (writeguard, agentguard), session automation (autopark, context, session-end), audit trail (clew events.jsonl), compaction handling (precompact), budget tracking, drift detection (driftdetect with auto-complaint filing), code hygiene (git-conventions, attribution-guard, validate). Go binary for consistency and performance (100ms timeout budget). ADR-0032 established knossos-owned canonical vocabulary (pre_tool, run_shell) -- CC and Gemini are translation peers. Fail-open principle throughout (exception: auth failure denies).

## Conceptual Model

**LifecycleAdapter pattern:** ClaudeAdapter and GeminiAdapter parse JSON from stdin, translate wire event names to canonical, return uniform Env struct. **18 canonical lifecycle events** with bidirectional/cc-only/gemini-only/outbound-only direction. **Envelope pattern (SCAR-009):** CC reads permissionDecision from hookSpecificOutput, not top-level. **Clew Contract:** events.jsonl with v2 flat events + v3 typed events coexisting. **Hook registration:** config/hooks.yaml (canonical vocabulary), materialized to settings.local.json with per-channel translation.

## Implementation Map

`internal/hook/` (8 files): env.go (StdinPayload, Env, ParseEnv), adapter.go + adapter_claude.go + adapter_gemini.go (LifecycleAdapter), events.go (canonical<->wire maps), output.go (PreToolUseOutput), input.go (ToolInput), auth.go (HMAC-SHA256). `internal/cmd/hook/` (18+ files): writeguard, agentguard, autopark, context, sessionend, clew, budget, precompact, gitconventions, attributionguard, validate, driftdetect, suggest, subagent, worktreeseed, worktreeremove, cheapo_revert. `internal/hook/clewcontract/` (20 files): event system (see session-event-system feature).

## Boundaries and Failure Modes

All handlers return allow on error (fail-open). Auth failure is the one explicit deny. 100ms default/500ms max Go timeout (inner); harness timeout is outer. Async hooks (clew, suggest, driftdetect) are non-blocking. ClaudeAdapter default when KNOSSOS_CHANNEL unset (PKG-010). Two output formats coexist (legacy + CC-native, SCAR-009). EventWriter open/close per write (O(n) overhead). Worktree events outside fully canonical set.

## Knowledge Gaps

1. context.go full implementation not read (richest handler)
2. ADR-0002 and ADR-0011 absent from disk
3. Hook registration materialization path (BuildHooksSettings) not traced

---
name: slop-chop-ref
description: "Slop-chop rite cross-cutting protocol reference. Use when: implementing any slop-chop specialist agent, checking severity rules, verifying read-only enforcement, understanding artifact chain or two-mode behavior. Triggers: slop-chop protocol, severity model, artifact chain, CI mode, read-only enforcement."
---

# slop-chop Protocol Reference (slop-chop-ref)

## Two-Mode System

All agents read a `mode` flag at invocation: `ci` or `interactive`.

| Behavior | CI Mode | Interactive Mode |
|----------|---------|------------------|
| User interaction | None. No questions, no pauses. | Surface ambiguities in **final report** only. |
| Output format | Structured artifacts (markdown + JSON). | Conversational summary with embedded findings. |
| Ambiguity | Report with confidence score. Never pause. | Confidence score + flag for human review. |

## Severity Model

Zero-config defaults. Override via `.slop-chop.yaml`.

| Severity | Gate Impact |
|----------|-------------|
| **CRITICAL** | FAIL (exit 1). Non-existent reference, provably broken code. |
| **HIGH** | FAIL (exit 1). Logic error with production impact, dangerous anti-pattern. |
| **MEDIUM** | Advisory. Copy-paste bloat, dependency sprawl, shallow tests. |
| **LOW** | Advisory. Minor style signal, borderline pattern. |
| **TEMPORAL** | Advisory ALWAYS. Never blocking regardless of count. |

Only CRITICAL and HIGH trigger FAIL. Temporal debt is accumulation risk, not immediate defect -- NEVER blocking. Every finding includes confidence (HIGH / MEDIUM / LOW); gate-keeper uses confidence to weigh borderline findings.

## Read-Only Enforcement

Agents NEVER modify target repository files.

| Tool | Target Repo | Own Artifacts |
|------|-------------|---------------|
| Read, Glob, Grep | YES | YES |
| Bash (git log, git blame, git diff, npm ls, pip list) | YES | YES |
| Write, Edit | **NO** | YES (reports, patches, verdicts) |

## Artifact Chain

Later agents receive ALL prior artifacts. At DIFF complexity, phases 3-4 are skipped.

| Phase | Agent | Receives | Produces |
|-------|-------|----------|----------|
| 1. Detection | hallucination-hunter | (source files) | detection-report |
| 2. Analysis | logic-surgeon | detection-report | analysis-report |
| 3. Decay | cruft-cutter | detection + analysis reports | decay-report |
| 4. Remediation | remedy-smith | all 3 prior reports | remedy-plan |
| 5. Verdict | gate-keeper | ALL prior artifacts | gate-verdict |

## Cross-Rite Referral Routing

Gate-keeper produces advisory referrals. Target rites decide whether to act.

| Condition | Target Rite |
|-----------|-------------|
| Security anti-patterns (unvalidated inputs, hardcoded secrets) | security |
| General code quality (complexity, naming, architecture) | hygiene |
| Test infrastructure gaps | 10x-dev |
| Systemic temporal debt at CODEBASE scope | debt-triage |

## Rite Anti-Patterns

- **AI witch-hunting**: Detect patterns, not provenance. Never label code as "AI-generated."
- **Over-remediation**: Advisory findings may be informational only. Not every finding needs a fix.
- **Hygiene drift**: General code quality (complexity, naming) belongs to hygiene, not slop-chop.
- **Temporal overreach**: Only flag dead weight from completed transitions, not code that was always dead.
- **False confidence in auto-fix**: When in doubt, MANUAL. Wrong auto-fixes are worse than none.
- **Ignoring context**: Documented retry policies, externally-controlled flags are not findings.
- **Comment zealotry**: Only ephemeral references (tickets, resolved TODOs). Permanent references (licenses, regulations) are legitimate.

---
name: slop-chop
description: "Switch to slop-chop rite (AI code quality gate). Use when: user says /slop-chop, reviewing AI-generated code, detecting hallucinated imports, checking phantom dependencies, temporal debt audit, PR quality gate for Copilot/Cursor code. Triggers: /slop-chop, AI code review, hallucinated imports, phantom dependencies, temporal debt, stale feature flags, ephemeral comments, vibe coding quality."
argument-hint: "[--overwrite-diverged] [--dry-run] [--keep-orphans]"
allowed-tools: Bash, Read
model: haiku
disable-model-invocation: true
context: fork
---

# /slop-chop - Switch to AI Code Quality Gate Rite

Switch to slop-chop, the AI code quality gate rite. Pro-AI, anti-slop.

## Behavior

### 1. Invoke Rite Switch

Execute via Bash tool:

```bash
ari sync --rite slop-chop $ARGUMENTS
```

### 2. Display Pantheon

After successful switch, show the agent table:

| Agent | Role |
|-------|------|
| hallucination-hunter | Verifies imports, dependencies, and API references against actual existence |
| logic-surgeon | Analyzes code behavior for logic errors, bloat, and unreviewed-output signals |
| cruft-cutter | Detects temporal debt: dead shims, stale flags, ephemeral comment artifacts |
| remedy-smith | Produces auto-fix patches and manual remediation guidance |
| gate-keeper | Issues quality gate verdict with CI-consumable output and cross-rite referrals |

### 3. Update Session

Confirm `ari sync` output shows the correct active rite.

## Flags

| Flag | Description |
|------|-------------|
| `--overwrite-diverged` | Force regeneration of diverged files |
| `--dry-run` | Preview changes without applying |
| `--keep-orphans` | Preserve orphaned knossos files (default: auto-remove) |

## When to Use

- AI-assisted code review (Copilot, Cursor, Codeium, GPT-generated code)
- Hallucination detection (phantom imports, invented APIs, non-existent methods)
- Temporal debt audit (stale feature flags, ephemeral comments, outdated AI assumptions)
- PR quality gate for AI-generated or AI-assisted code
- Vibe coding quality checks before merge

**Don't use for**: General code smells --> `/hygiene` | Architecture issues --> `/arch` | Security audits --> `/security` | General tech debt (planning) --> `/debt`

**Hard gate**: FAIL exits 1 and blocks merge. Temporal findings are always advisory.

## Reference

Full documentation: `slop-chop-ref` skill

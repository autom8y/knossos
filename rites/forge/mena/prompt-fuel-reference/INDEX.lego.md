---
name: prompt-fuel-reference
description: "Guide for translating archaeology HANDOFF-PROMPT-FUEL into agent prompts. Use when: writing agent prompts after archaeology phase, embedding domain knowledge into system instructions, calibrating Exousia from tribal knowledge. Triggers: prompt fuel, handoff consumption, archaeology translation, domain knowledge embedding."
---

# Prompt Fuel Reference

> How to consume HANDOFF-PROMPT-FUEL and translate it into expert-level agent prompts

## When This Skill Is Needed

A `HANDOFF-PROMPT-FUEL.md` exists in `.claude/wip/{RITE_NAME}/` (produced by the domain-forensics agent during the archaeology phase). You are about to write agent prompts for this rite.

Without this skill, you write generic prompts. With it, you embed hard-won domain knowledge that prevents agents from repeating historical mistakes.

## The Three-Tier Translation Model

| HANDOFF Tier | Prompt Target | Token Budget | Rationale |
|--------------|---------------|--------------|-----------|
| **CRITICAL** | Agent system prompt, `## Domain Knowledge` section | 30-50 lines per agent | Must always be in context -- failure modes, jurisdiction boundaries |
| **IMPORTANT** | Packaged as `domain-knowledge/` skill in new rite's mena/ | 40-60 lines per skill file | Loaded on demand via `skills:` frontmatter |
| **NICE-TO-HAVE** | Listed in "Further Reading" section or left in archived HANDOFF | 5-10 lines | Agent knows these exist but does not preload |

## Compression Strategy

Raw archaeology output (~2,000+ lines) compresses to HANDOFF (~300-400 lines), then to per-agent CRITICAL (~30-50 lines each). Net compression: ~11x from raw to prompt.

**Techniques for distilling CRITICAL items into 30-50 lines:**
- Merge items sharing the same root cause (e.g., three scars from error swallowing become one constraint)
- Lead with the imperative ("NEVER", "MUST", "ALWAYS"), then the rationale
- Include source IDs inline (`[SCAR-003 + TRIBAL-003]`) -- not as footnotes
- Drop narrative context that does not change agent behavior
- One line per constraint when possible; two lines max for complex rules

## Per-Agent Processing Workflow

Process agents sequentially, not in parallel. Each agent prompt is self-contained.

1. **Load once**: Read full RITE-SPEC + HANDOFF Cross-Agent Knowledge (CK-NN) + Anti-Pattern Catalog (AP-NN)
2. **Per agent**: Read the agent's Prompt Fuel section + relevant Exousia Overrides (EX-NN)
3. **Write**: Embed CRITICAL items as `## Domain Knowledge`, calibrate Exousia, add anti-patterns
4. **Verify**: Check source ID traceability -- every constraint references its origin
5. **Move on**: Next agent

## Companion Files

| File | Purpose |
|------|---------|
| [tier-mapping.md](tier-mapping.md) | Detailed rules for CRITICAL/IMPORTANT/NICE-TO-HAVE placement |
| [exousia-calibration.md](exousia-calibration.md) | Translating tribal knowledge EX-NN entries into Exousia boundaries |
| [anti-pattern-embedding.md](anti-pattern-embedding.md) | Embedding AP-NN items as behavioral DO NOT constraints |

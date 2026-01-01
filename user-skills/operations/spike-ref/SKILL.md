---
name: spike-ref
description: "Time-boxed research and exploration producing NO production code. Use when: exploring technical feasibility, investigating approaches, answering 'Can we do X?', building proof of concept. Triggers: /spike, research, explore, investigate, feasibility, proof of concept."
---

# /spike - Time-Boxed Research

> Execute time-boxed research to answer technical questions WITHOUT producing production code.

## Decision Tree

```
Starting research?
├─ Feasibility question → /spike
├─ Known implementation → /task
├─ Multiple approaches to evaluate → /spike --deliverable=comparison
├─ Quick library check → /spike --timebox=30m
└─ Complex research (8h+) → Break into phases or use /sprint
```

## Usage

```bash
/spike "research-question" [--timebox=DURATION] [--deliverable=TYPE]
```

### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `research-question` | Yes | - | What you're trying to learn or validate |
| `--timebox` | No | 2h | Time limit: 30m, 1h, 2h, 4h, 8h |
| `--deliverable` | No | report | report \| poc \| comparison \| decision |

## Quick Reference

**Pre-flight**:
- Clear research question
- Defined success criteria
- Realistic timebox (30m-8h)

**Actions**:
1. Define question and success criteria
2. Set timebox (default: 2h)
3. Invoke appropriate agent (Architect or Engineer)
4. Research with progress checkpoints (25%, 50%, 75%, 100%)
5. Generate spike report

**Produces**:
- `/docs/research/SPIKE-{slug}.md` (always)
- `/tmp/spike-{slug}/` POC code (optional, throwaway)

**Never Produces**:
- Production code
- PRD/TDD
- Production tests

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Spike without timebox | Becomes endless research | Set explicit limit |
| Ship spike code | Quality relaxed for research | Create /task for production |
| Spike known solutions | Wastes time | Use /task directly |
| Over-scope research | Loses focus | One question per spike |
| Exceed 8h timebox | Too large for spike | Break into phases or use /task |

## Prerequisites

- Clear research question
- Defined success criteria
- Realistic timebox

**No session required**: Spikes can run standalone.

## Success Criteria

- Question answered (or documented as unanswerable)
- Findings documented in spike report
- Time budget respected
- Next steps clear

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/task` | After spike approves approach, build production version |
| `/start` | Begin full session (not for standalone spikes) |
| `/sprint` | Multiple related spikes or tasks |
| `/hotfix` | Rapid fix (different from research) |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence, workflow diagram, state changes
- [examples.md](examples.md) - 4 usage scenarios with sample outputs
- [templates.md](templates.md) - Agent prompts, spike report template, quick-start templates
- [notes.md](notes.md) - Time-boxing philosophy, POC guidelines, spike-to-task handoff

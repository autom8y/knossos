---
name: doc-consolidation
description: "Doc consolidation schemas. Use when: merging fragmented documentation, extracting authoritative content from scattered sources, tracking consolidation progress. Triggers: consolidation, extraction, checkpoint, manifest, synthesis, staleness detection."
---

# Documentation Consolidation Workflow

## Purpose

Phased approach to consolidating fragmented documentation into authoritative documents. Solves the "scattered knowledge" problem: multiple source files with inconsistencies, duplication, and high token costs.

```
Before: Source A + Source B + Source C (~10K tokens read by agent)
After:  Extraction (~2K tokens) → Synthesis Agent → Consolidated Doc
```

## Workflow Phases

| Phase | Agent | Input | Output |
|-------|-------|-------|--------|
| 0. Discovery | Discovery Agent | File listing, first 50 lines | `MANIFEST.yaml` |
| 1. Extraction | Extractor Agent | Source files, manifest | `extraction-{topic}.yaml` |
| 2. Synthesis | Synthesis Agent | Extraction artifact | Consolidated document |
| 3. Review | Review Agent | Draft document | Approved/feedback |

## Artifact Summary

| Artifact | Location | Target Size | Schema |
|----------|----------|-------------|--------|
| `MANIFEST.yaml` | `.consolidation/MANIFEST.yaml` | < 1K tokens | [manifest-schema.md](schemas/manifest-schema.md) |
| `extraction-{topic}.yaml` | `.consolidation/extraction-{topic}.yaml` | < 2K tokens | [extraction-schema.md](schemas/extraction-schema.md) |
| `checkpoint-{topic}.yaml` | `.consolidation/checkpoint-{topic}.yaml` | < 500 tokens | [checkpoint-format.md](schemas/checkpoint-format.md) |

Staleness detection uses SHA-256 hashes. When a source hash doesn't match: mark `stale`, re-run extraction, update synthesis.

## Templates

| Template | When to Use |
|----------|-------------|
| [topic-summary.md](templates/topic-summary.md) | Phase 1: document extraction findings |
| [numbered-doc.md](templates/numbered-doc.md) | Phase 2: consolidated document structure |
| [index.md](templates/index.md) | Phase 4: archive tracking |
| [consolidation-manifest.md](templates/consolidation-manifest.md) | All phases: overall progress tracking |

## Patterns

| Pattern | When to Use |
|---------|-------------|
| [grouping-heuristics.md](patterns/grouping-heuristics.md) | Phase 0: classifying files into topics |
| [merge-strategies.md](patterns/merge-strategies.md) | Phase 1-2: resolving overlapping content |
| [numbering-conventions.md](patterns/numbering-conventions.md) | Phase 2: assigning document numbers |

## Examples

| Example | Demonstrates |
|---------|--------------|
| [manifest-example.yaml](examples/manifest-example.yaml) | Complete MANIFEST.yaml (8 files, 3 topics) |
| [extraction-example.yaml](examples/extraction-example.yaml) | Extraction artifact (78% token reduction) |
| [checkpoint-example.yaml](examples/checkpoint-example.yaml) | Mid-synthesis checkpoint with activity log |

## Validation

| File | When to Use |
|------|-------------|
| [checklist.md](validation/checklist.md) | Before Phase 4: information preservation checks |
| [gates.md](validation/gates.md) | All phases: mandatory phase transition gates |
| [test-matrix.md](validation/test-matrix.md) | Testing: validate workflow handles all cases |

## Related Skills

- `doc-ecosystem` skill — Ecosystem templates (Gap Analysis, Context Design)
- `documentation` skill — Core artifact templates (PRD, TDD, ADR)
- `10x-workflow` skill — Agent coordination patterns

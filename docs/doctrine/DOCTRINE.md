# The Knossos Doctrine

> **Entry point to the philosophical and compliance documentation of the Knossos platform.**

This directory organizes all doctrine-related documentation: the philosophical foundation, architectural decisions, compliance status, and operational guides.

---

## The Coda

The **Coda** is the philosophical foundation—the "why" behind every design decision. Start here to understand Knossos at its deepest level.

**Primary Document**: [philosophy/knossos-doctrine.md](philosophy/knossos-doctrine.md)

**Core Insight**: *The myth is the architecture. The architecture is the myth.*

| Myth | Component | Function |
|------|-----------|----------|
| **Knossos** | The platform | The labyrinth itself |
| **Ariadne** | CLI binary (`ari`) | The clew ensuring return |
| **Theseus** | Claude Code agent | The navigator with amnesia |
| **Moirai** | Session lifecycle | The Fates who spin, measure, and cut |
| **White Sails** | Confidence signal | Honest return indicator |
| **Rites** | Practice bundles | Invokable ceremonies |

---

## Directory Structure

```
docs/doctrine/
├── DOCTRINE.md                 # You are here - entry point
│
├── philosophy/                 # The Coda (why we exist)
│   ├── knossos-doctrine.md     # Full philosophical foundation
│   ├── design-principles.md    # Core principles extracted
│   └── mythology-concordance.md # Myth ↔ SOURCE implementation map
│
├── foundations/                # Core architectural decisions (symlinks to ../../decisions/)
│   ├── ADR-0001-session-state-machine-redesign.md
│   ├── ADR-0005-moirai-centralized-state-authority.md
│   └── ADR-0009-knossos-roster-identity.md
│
├── compliance/                 # Validation, audits, status
│   ├── COMPLIANCE-STATUS.md    # Current implementation status
│   └── audits/                 # Documentation audit reports
│       ├── doctrine-audit-20260108.md
│       ├── ia-assessment-20260108.md
│       ├── refinement-recommendations-20260108.md
│       └── quality-review-20260108.md
│
├── operations/                 # How to use the platform
│   └── cli-reference/          # CLI command reference (72 commands, 14 families)
│       ├── index.md            # Quick reference and navigation
│       ├── cli-session.md      # Session lifecycle (11 commands)
│       ├── cli-rite.md         # Rite management (10 commands)
│       ├── cli-worktree.md     # Worktree operations (10 commands)
│       └── ...                 # All 14 command families
│
├── rites/                      # Catalog of 11 rites
│   ├── index.md                # Rite selection guide
│   ├── 10x-dev.md              # Full development lifecycle
│   ├── forge.md                # Meta-rite for rite creation
│   └── ...                     # All 11 rites documented
│
├── guides/                     # Operational guides
│   └── worktree-guide.md       # Worktree production patterns
│
└── reference/                  # Navigation and lookup
    ├── INDEX.md                # Master navigation hub
    └── GLOSSARY.md             # Terminology reference
```

### Structure Notes

**Canonical Identity** (per ADR-0009 Amendment):
- **SOURCE** = `/roster/` repository (what Knossos IS)
- **PROJECTION** = `.claude/` directories (materialized by `ari sync materialize`)

**Completed** (2026-01-08 Sprint 2):
- `operations/cli-reference/` - 14 CLI reference files (72 commands)
- `rites/` - 11 rite documentation files with selection guide
- `guides/worktree-guide.md` - Worktree production patterns

**Removed** (2026-01-08):
- Empty scaffolding directories collapsed (architecture/, evolution/ subdirs)
- Structure now reflects actual content, not aspirations
- Will expand organically as content emerges

**Symlinks**:
- `foundations/` → `../../decisions/` (ADRs)
- `operations/guides/` → `../../../guides/` (operational guides)

---

## Quick Navigation

| I want to... | Go to |
|--------------|-------|
| Understand the philosophy | [philosophy/knossos-doctrine.md](philosophy/knossos-doctrine.md) |
| See current compliance status | [compliance/COMPLIANCE-STATUS.md](compliance/COMPLIANCE-STATUS.md) |
| Understand the naming | [philosophy/mythology-concordance.md](philosophy/mythology-concordance.md) |
| Learn design principles | [philosophy/design-principles.md](philosophy/design-principles.md) |
| Find all documentation | [reference/INDEX.md](reference/INDEX.md) |

---

## The Journey So Far

**Doctrine Launch Sprint Achievement**: 82% → 95%+ compliance

| Category | Status |
|----------|--------|
| Weight Economy | 90% |
| Thread Contract | 100% (14 event types) |
| Moirai | 85% (unified agent) |
| White Sails | 100% |
| Rite Operations | 90% |
| Inscription System | 95% |
| Naxos Detection | 100% |
| Terminology | 95%+ |

See [compliance/COMPLIANCE-STATUS.md](compliance/COMPLIANCE-STATUS.md) for the complete report.

---

## Living Documentation

This doctrine is **living**—it evolves with the platform. Key principles:

1. **Bidirectional alignment**: Doctrine informs implementation; implementation refines doctrine
2. **Gaps are acknowledged**: Section XIV of the Coda explicitly lists implementation drift
3. **The working system matters**: What runs in production is truth; documentation follows

---

*Enter with the clew. Return with confidence.*

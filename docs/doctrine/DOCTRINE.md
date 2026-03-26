---
last_verified: 2026-03-26
---

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
| **Ariadne** | CLI binary (`ari`) | The intelligent navigator ensuring return |
| **Theseus** | AI harness agent | The navigator with amnesia |
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
├── foundations/                # (empty — ADR symlinks removed; see docs/decisions/ for ADRs)
│
├── compliance/                 # Validation, status
│   └── COMPLIANCE-STATUS.md    # Current implementation status
│
├── operations/                 # How to use the platform
│   └── cli-reference/          # CLI command reference (all 32 families documented)
│       ├── index.md            # Quick reference and navigation
│       ├── cli-session.md      # Session lifecycle (15 commands)
│       ├── cli-rite.md         # Rite management (10 commands)
│       ├── cli-worktree.md     # Worktree operations (11 commands)
│       ├── cli-sync.md         # Materialization (8 commands)
│       ├── cli-hook.md         # Hook operations (11 commands)
│       ├── cli-handoff.md      # Agent handoffs (4 commands)
│       ├── cli-agent.md        # Agent management (summon, dismiss, roster)
│       ├── cli-serve.md        # Clew HTTP server
│       ├── cli-procession.md   # Cross-rite workflows
│       ├── cli-land.md         # Session synthesis
│       ├── cli-status.md       # Platform health dashboard
│       ├── cli-org.md          # Organization management
│       ├── cli-registry.md     # Registry sync and status
│       ├── cli-ask.md          # Domain-filtered queries
│       ├── cli-knows.md        # Knowledge base checks
│       ├── cli-ledge.md        # Work product artifacts
│       ├── cli-lint.md         # Codebase linting
│       ├── cli-provenance.md   # Content provenance
│       ├── cli-explain.md      # Term definitions
│       ├── cli-complaint.md    # Complaint tracking
│       ├── cli-init.md         # Project initialization
│       ├── cli-tour.md         # Interactive tour
│       ├── cli-version.md      # Version info
│       └── cli-help.md         # CLI help
│
├── rites/                      # Catalog of 19 rites (all documented)
│   ├── index.md                # Rite selection guide
│   ├── 10x-dev.md              # Full development lifecycle
│   ├── arch.md                 # Architecture assessment
│   ├── clinic.md               # Clinical debugging workflow
│   ├── docs.md                 # Documentation workflow
│   ├── forge.md                # Rite creation (meta-rite)
│   ├── hygiene.md              # Code quality maintenance
│   ├── debt-triage.md          # Technical debt remediation
│   ├── releaser.md             # Release management
│   ├── review.md               # Code review workflow
│   ├── security.md             # Threat modeling and compliance
│   ├── sre.md                  # Operations and reliability
│   ├── intelligence.md         # Research and synthesis
│   ├── rnd.md                  # Exploration and prototypes
│   ├── strategy.md             # Business analysis
│   ├── thermia.md              # Thermal/performance analysis
│   ├── ecosystem.md            # Platform infrastructure
│   ├── slop-chop.md            # AI code quality gate
│   ├── ui.md                   # UI development workflow
│   └── shared.md               # Cross-rite resources
│
├── guides/                     # Operational guides
│   └── worktree-guide.md       # Worktree production patterns
│
└── reference/                  # Navigation and lookup
    ├── INDEX.md                # Master navigation hub
    ├── GLOSSARY.md             # Terminology reference
    ├── agent-capabilities.md   # CC-OPP capability reference
    └── architecture-map.md     # Subsystem and package map
```

### Structure Notes

**Canonical Identity** (per ADR-0009 Amendment):
- **SOURCE** = Knossos repository (what Knossos IS)
- **PROJECTION** = channel directories (materialized by `ari sync materialize`)

**Completed**:
- `operations/cli-reference/` - CLI reference (all 32 command families documented)
- `rites/` - All 19 rite documentation files with selection guide
- `guides/worktree-guide.md` - Worktree production patterns
- `reference/agent-capabilities.md` - CC-OPP capability reference
- `reference/architecture-map.md` - Subsystem and package map

**Removed** (2026-01-08):
- Empty scaffolding directories collapsed (architecture/, evolution/ subdirs)
- Structure now reflects actual content, not aspirations
- Will expand organically as content emerges

**Note**: `foundations/` previously contained ADR symlinks (ADR-0001, ADR-0005, ADR-0009) that were removed because their targets did not exist in `docs/decisions/`. Published ADRs: `docs/decisions/` (ADR-0031, ADR-0032) and `.ledge/decisions/` (ADR-0030, others).

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

## Current State

| Metric | Value |
|--------|-------|
| Go source lines | 93,668 (non-test files in cmd/ and internal/; excludes 119,624 test lines) |
| CLI command families | 32 (see `ari --help` for current list) |
| Agents | 107 across 19 rites |
| ADRs | 6 published documents (2 in docs/decisions/, 4 in .ledge/decisions/; numbers allocated through ADR-0032) |

See [compliance/COMPLIANCE-STATUS.md](compliance/COMPLIANCE-STATUS.md) for the complete report.

---

## Living Documentation

This doctrine is **living**—it evolves with the platform. Key principles:

1. **Bidirectional alignment**: Doctrine informs implementation; implementation refines doctrine
2. **Gaps are acknowledged**: Compliance status tracks honest gap assessment
3. **The working system matters**: What runs in production is truth; documentation follows

---

*Enter with the clew. Return with confidence.*

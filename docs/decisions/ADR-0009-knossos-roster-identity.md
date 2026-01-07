# ADR-0009: Knossos-Roster Identity and Ariadne Naming

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-01-05 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The project has evolved through multiple naming iterations, creating potential confusion about canonical identity and component relationships.

### Historical Evolution

1. **"roster"** - The original repository name, reflecting a "collection of agent configurations"
2. **"Knossos"** - The conceptual platform name from Greek mythology (the labyrinth)
3. **"Ariadne"** - The CLI binary (`ari`) that navigates the labyrinth

### Current State

- Repository: `github.com/autom8y/roster`
- Go module: `github.com/autom8y/knossos`
- CLI binary: `ari`
- Configuration directory: `.claude/`
- Documentation references: Mixed usage of "roster", "Knossos", and "Ariadne"

### Ambiguity Points

| Question | Current Answer |
|----------|----------------|
| What is the canonical project name? | Unclear - "roster" (repo) vs "Knossos" (concept) |
| How do components relate? | Undocumented |
| When will repository be renamed? | Unspecified |
| How should documentation reference the platform? | Inconsistent |

### Forces

- **Clarity**: New contributors need a clear mental model
- **Migration**: Eventual rename requires planning
- **Consistency**: Documentation and code should align
- **Mythology**: The Greek metaphor aids understanding when used consistently

## Decision

### Identity Statement

**roster/.claude/ IS Knossos.**

The repository currently named "roster" is the Knossos platform. The `.claude/` directory structure represents the labyrinth's architecture. This identity relationship is canonical regardless of the repository's current name.

### Component Relationships

| Component | Role | Mythological Equivalent |
|-----------|------|-------------------------|
| **Knossos** | The platform (repository + configuration) | The labyrinth |
| **Ariadne** | The CLI binary (`ari`) | The thread through the labyrinth |
| **.claude/** | Configuration directory | The labyrinth's structure |
| **Hooks** | Event triggers | Trap mechanisms |
| **Skills** | Capability modules | Rooms in the labyrinth |
| **state-mate** | State mutation authority | The architect (Daedalus) |
| **Sessions** | Tracked work units | Journeys through the labyrinth |
| **White Sails** | Confidence signals | Safe return indicators |

### The Knossos Metaphor (Canonical)

From PRD-ariadne.md:

| Myth | Knossos Equivalent |
|------|-------------------|
| The Thread | Session state + provenance + audit trail |
| The Labyrinth | Codebase complexity |
| Navigation | `ari session`, `ari team` commands |
| Survival | Lock management, atomic operations, recovery |
| Return to Athens | Successful session wrap with quality gates |
| White Sails | Honest confidence signaling (Aegeus problem) |

### Naming Convention

| Context | Use | Example |
|---------|-----|---------|
| **Documentation** (platform concepts) | "Knossos" | "Knossos session management" |
| **Documentation** (CLI) | "Ariadne" or "ari" | "Ariadne provides the thread" |
| **Code** (paths, imports) | Current names | `roster/ariadne/...` |
| **CLI invocation** | `ari` | `ari session status` |
| **Repository reference** | "roster" (until rename) | "Clone the roster repository" |

### Rename Criteria

The repository rename from `roster` to `knossos` will occur when ALL of the following are satisfied:

1. **90% Integration Milestone**: Core Ariadne commands replace bash scripts
2. **Handoff Events Complete**: All session lifecycle events implemented in Go
3. **Self-Hosting Proven**: The platform can manage its own development sessions
4. **Import Paths Mapped**: Clear migration path for Go module imports
5. **Documentation Aligned**: All docs reference Knossos consistently

### Post-Rename Structure

```
knossos/                    # The platform (was: roster)
├── ariadne/                # The thread (Go CLI)
│   └── cmd/ari/            # Binary entry point
├── .claude/                # Labyrinth configuration
│   ├── agents/             # Agent personalities
│   ├── hooks/              # Event triggers
│   ├── sessions/           # Journey records
│   └── skills/             # Capability modules
├── schemas/                # Validation schemas
└── docs/                   # Platform documentation
```

## Consequences

### Positive

1. **Clear Mental Model**: Contributors understand the mythological architecture
2. **Consistent Documentation**: "Knossos" for platform, "Ariadne" for CLI
3. **Explicit Rename Path**: Criteria-based rename prevents premature migration
4. **Mythology as Documentation**: The metaphor encodes architectural intent
5. **Component Boundaries**: Clear separation between platform and tooling

### Negative

1. **Short-Term Confusion**: Dual naming during transition period
2. **Import Path Churn**: Go module path will change at rename
3. **External Reference Breakage**: Links to "roster" will require redirects
4. **Learning Curve**: New contributors must understand the mythology

### Neutral

1. **Documentation Convention**: Must consistently use "Knossos" for platform concepts
2. **Code Convention**: Paths remain `roster/` until rename
3. **CLI Stability**: `ari` binary name is final regardless of repo name
4. **Mythology Maintenance**: Team should know the Ariadne myth basics

## Implementation Notes

### Documentation Updates Required

- [ ] Update README.md with identity statement
- [ ] Add "Naming" section to contributor guide
- [ ] Audit existing docs for inconsistent naming
- [ ] Create mythology reference page

### Before Rename Checklist

- [ ] All Ariadne v1.0 commands implemented
- [ ] Session lifecycle fully in Go
- [ ] Self-hosting dogfood period complete
- [ ] Import path migration tool ready
- [ ] GitHub redirect configured
- [ ] This ADR updated with rename date

## Related Decisions

- **ADR-0001**: Session State Machine (Knossos state management foundation)
- **ADR-0005**: state-mate Authority (Daedalus pattern for mutations)
- **ADR-0008**: Handoff Schema Embedding (Ariadne integration pattern)

## References

- PRD-ariadne.md: Section 3.3 (Knossos Future), The Knossos Metaphor table
- TDD-knossos-v2.md: White Sails and platform architecture
- Greek Mythology: Ariadne's thread guided Theseus through the Labyrinth of Knossos
- ariadne/cmd/ari/main.go: CLI binary entry point

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-01-05 | Claude Code (Architect) | Initial acceptance - formalizing platform identity |
